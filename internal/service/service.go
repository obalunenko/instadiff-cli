// Package service implements instagram account operations and business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Davincible/goinsta"

	"github.com/obalunenko/instadiff-cli/internal/config"
	"github.com/obalunenko/instadiff-cli/internal/db"
	"github.com/obalunenko/instadiff-cli/internal/models"
	"github.com/obalunenko/instadiff-cli/pkg/bar"
	"github.com/obalunenko/instadiff-cli/pkg/spinner"
	log "github.com/obalunenko/logger"
)

var (
	// ErrLimitExceed returned when limit for action exceeded.
	ErrLimitExceed = errors.New("limit exceeded")
	// ErrCorrupted returned when instagram returned error response more than one time during loop processing.
	ErrCorrupted = errors.New("unable to continue - instagram responses with errors")
)

// Service represents service for operating instagram account.
type Service struct {
	instagram instagram
	storage   db.DB
	debug     bool
}

type instagram struct {
	client    *goinsta.Instagram
	whitelist map[string]struct{}
	limits    limits
	sleep     time.Duration
}

func (i instagram) Whitelist() map[string]struct{} {
	return i.whitelist
}

func (i instagram) Limits() limits {
	return i.limits
}

func (i instagram) Client() *goinsta.Instagram {
	return i.client
}

func (i instagram) Sleep() time.Duration {
	return i.sleep
}

type limits struct {
	unFollow int
}

// StopFunc closure func that will stop service.
type StopFunc func() error

// New creates new instance of Service instance and returns closure func that will stop service.
//
// Usage:
// svc, stop, err := New(config.Config{})
// if err != nil{
// // handle error
// }
// defer stop().
//
func New(ctx context.Context, cfg config.Config, cfgPath string) (*Service, StopFunc, error) {
	cl, err := makeClient(ctx, cfg, cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("make client: %w", err)
	}

	log.WithField(ctx, "username", cl.Account.Username).Info("Logged-in")

	if err = cl.OpenApp(); err != nil {
		log.WithError(ctx, err).Error("Failed to refresh app info")
	}

	dbc, err := db.Connect(db.Params{
		LocalDB: cfg.IsLocalDBEnabled(),
		MongoParams: db.MongoParams{
			URL:        cfg.MongoConfigURL(),
			Database:   cfg.MongoDBName(),
			Collection: cfg.MongoDBCollection(),
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("db connect: %w", err)
	}

	svc := Service{
		instagram: instagram{
			client:    cl,
			whitelist: cfg.Whitelist(),
			limits: limits{
				unFollow: cfg.UnFollowLimits(),
			},
			sleep: cfg.Sleep(),
		},
		storage: dbc,
		debug:   cfg.Debug(),
	}

	stopFunc := StopFunc(func() error {
		return svc.stop()
	})

	return &svc, stopFunc, nil
}

// GetFollowers returns list of followers for logged-in user.
func (svc *Service) GetFollowers(ctx context.Context) ([]models.User, error) {
	stop := spinner.Set("Fetching followers", "", "yellow")
	defer stop()

	followers := makeUsersList(svc.instagram.client.Account.Followers())

	bt := models.UsersBatchTypeFollowers

	return svc.storeUsers(ctx, models.UsersBatch{
		Users:     followers,
		Type:      bt,
		CreatedAt: time.Now(),
	})
}

// GetFollowings returns list of followings for logged-in user.
func (svc *Service) GetFollowings(ctx context.Context) ([]models.User, error) {
	stop := spinner.Set("Fetching followings", "", "yellow")
	defer stop()

	followings := makeUsersList(svc.instagram.client.Account.Following())

	bt := models.UsersBatchTypeFollowings

	return svc.storeUsers(ctx, models.UsersBatch{
		Users:     followings,
		Type:      bt,
		CreatedAt: time.Now(),
	})
}

func (svc *Service) storeUsers(ctx context.Context, batch models.UsersBatch) ([]models.User, error) {
	if len(batch.Users) == 0 {
		return nil, makeNoUsersError(batch.Type)
	}

	err := svc.storage.InsertUsersBatch(ctx, batch)
	if err != nil {
		log.WithFields(ctx, log.Fields{
			"error":      err,
			"batch_type": batch.Type.String(),
		}).Error("Failed to insert batch to db")
	}

	return batch.Users, nil
}

func makeUsersList(users *goinsta.Users) []models.User {
	usersList := make([]models.User, 0, len(users.Users))

	for users.Next() {
		for i := range users.Users {
			usersList = append(usersList,
				models.MakeUser(users.Users[i].ID, users.Users[i].Username, users.Users[i].FullName))
		}
	}

	return usersList
}

// GetNotMutualFollowers returns list of users that not following back.
func (svc *Service) GetNotMutualFollowers(ctx context.Context) ([]models.User, error) {
	followers, err := svc.GetFollowers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get followers: %w", err)
	}

	log.WithField(ctx, "count", len(followers)).Info("Total followers")

	followings, err := svc.GetFollowings(ctx)
	if err != nil {
		return nil, fmt.Errorf("get followings: %w", err)
	}

	log.WithField(ctx, "count", len(followings)).Info("Total followings")

	stop := spinner.Set("Detecting not mutual followers", "", "yellow")
	defer stop()

	followersMap := make(map[int64]struct{}, len(followers))

	for _, fu := range followers {
		followersMap[fu.ID] = struct{}{}
	}

	var notmutual []models.User

	for _, fu := range followings {
		if _, mutual := followersMap[fu.ID]; !mutual {
			notmutual = append(notmutual, fu)
		}
	}

	return svc.storeUsers(ctx, models.UsersBatch{
		Users:     notmutual,
		Type:      models.UsersBatchTypeNotMutual,
		CreatedAt: time.Now(),
	})
}

// UnFollow removes user from followings.
func (svc *Service) UnFollow(ctx context.Context, user models.User) error {
	log.WithField(ctx, "username", user.UserName).Debug("Unfollow user")

	if svc.debug {
		return nil
	}

	us := goinsta.User{
		ID:       user.ID,
		Username: user.UserName,
	}

	us.SetInstagram(svc.instagram.client)

	if err := us.Unfollow(); err != nil {
		return fmt.Errorf("[%s] unfollow: %w", user.UserName, err)
	}

	return nil
}

// Follow adds user to followings.
func (svc *Service) Follow(ctx context.Context, user models.User) error {
	log.WithField(ctx, "username", user.UserName).Debug("Follow user")

	if svc.debug {
		return nil
	}

	us := goinsta.User{
		ID:       user.ID,
		Username: user.UserName,
	}

	us.SetInstagram(svc.instagram.client)

	if err := us.Follow(); err != nil {
		return fmt.Errorf("[%s] follow: %w", user.UserName, err)
	}

	return nil
}

func getBarType(ctx context.Context) bar.BType {
	bType := bar.BTypeRendered

	if log.FromContext(ctx).LogLevel().IsDebug() {
		bType = bar.BTypeVoid
	}

	return bType
}

// UnFollowAllNotMutualExceptWhitelisted clean followings from users that not following back
// except of whitelisted users.
func (svc *Service) UnFollowAllNotMutualExceptWhitelisted(ctx context.Context) (int, error) {
	notMutual, err := svc.GetNotMutualFollowers(ctx)
	if err != nil {
		return 0, fmt.Errorf("get not mutual followers: %w", err)
	}

	if len(notMutual) == 0 {
		return 0, nil
	}

	log.WithFields(ctx, log.Fields{
		"count":       len(notMutual),
		"whitelisted": len(svc.instagram.Whitelist()),
	}).Info("Not mutual followers")

	diff := svc.whitelistNotMutual(notMutual)

	if len(diff) == 0 {
		return 0, nil
	}

	pBar := bar.New(len(diff), getBarType(ctx))

	go pBar.Run(ctx)

	defer func() {
		pBar.Finish()
	}()

	return svc.unfollowUsers(ctx, pBar, notMutual)
}

func (svc *Service) whitelistNotMutual(notMutual []models.User) []models.User {
	result := make([]models.User, 0, len(notMutual))

	whitelist := svc.instagram.Whitelist()

	for i := range notMutual {
		u := notMutual[i]

		if _, exist := whitelist[u.UserName]; !exist {
			result = append(result, u)
		}
	}

	return result
}

func (svc *Service) getUsersByUsername(usernames []string) ([]*goinsta.User, error) {
	stop := spinner.Set("Fetching users by names", "", "yellow")
	defer stop()

	users := make([]*goinsta.User, 0, len(usernames))

	for _, un := range usernames {
		u, err := svc.instagram.client.Profiles.ByName(un)
		if err != nil {
			return nil, fmt.Errorf("get user profile by name[%s]: %w", un, err)
		}

		users = append(users, u)
	}

	return users, nil
}

// RemoveFollowersByUsername removes all provided users by blocking and unblocking them to bypass Instagram limits.
func (svc *Service) RemoveFollowersByUsername(ctx context.Context, usernames []string) (int, error) {
	pBar := bar.New(len(usernames), getBarType(ctx))

	go pBar.Run(ctx)

	defer func() {
		pBar.Finish()
	}()

	return svc.removeFollowers(ctx, pBar, usernames)
}

func (svc *Service) removeFollowers(ctx context.Context, pBar bar.Bar, users []string) (int, error) {
	var count int

	ticker := time.NewTicker(svc.instagram.sleep)
	defer ticker.Stop()

	const errsLimit = 3

	var errsNum int

LOOP:
	for _, un := range users {
		if errsNum >= errsLimit {
			return count, ErrCorrupted
		}

		select {
		case <-ctx.Done():
			break LOOP
		case <-ticker.C:
			pBar.Progress() <- struct{}{}

			u, err := svc.instagram.client.Profiles.ByName(un)
			if err != nil {
				log.WithError(ctx, err).WithField("username", un).Error("User lookup failed")

				errsNum++
				continue
			}

			u.SetInstagram(svc.instagram.client)

			if err = u.Block(false); err != nil {
				log.WithError(ctx, err).WithField("username", u.Username).Error("Failed to block follower")

				errsNum++

				continue
			}

			if err = u.Unblock(); err != nil {
				log.WithError(ctx, err).WithField("username", u.Username).Error("Failed to unblock follower")

				errsNum++

				continue
			}

			count++

			if count >= svc.instagram.limits.unFollow {
				return count, ErrLimitExceed
			}

			time.Sleep(svc.instagram.sleep)
		}
	}

	return count, nil
}

func (svc *Service) unfollowUsers(ctx context.Context, pBar bar.Bar, users []models.User) (int, error) {
	var count int

	ticker := time.NewTicker(svc.instagram.sleep)
	defer ticker.Stop()

	const errsLimit = 3

	var errsNum int

	whitelist := svc.instagram.Whitelist()

LOOP:
	for _, u := range users {
		if errsNum >= errsLimit {
			return count, ErrCorrupted
		}

		select {
		case <-ctx.Done():
			break LOOP
		case <-ticker.C:
			if _, exist := whitelist[u.UserName]; exist {
				continue
			}

			if _, exist := whitelist[strconv.FormatInt(u.ID, 10)]; exist {
				continue
			}

			pBar.Progress() <- struct{}{}

			if err := svc.UnFollow(ctx, u); err != nil {
				log.WithError(ctx, err).WithField("username", u.UserName).Error("Failed to unfollow")
				errsNum++

				continue
			}

			count++

			if count >= svc.instagram.limits.unFollow {
				return count, ErrLimitExceed
			}
		}
	}

	return count, nil
}

// stop logs out from instagram and clean sessions.
// Should be called in defer after creating new instance from New().
func (svc *Service) stop() error {
	if err := svc.instagram.client.Logout(); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	return nil
}

type isBotResult struct {
	user  models.User
	isBot bool
}

// GetBusinessAccountsOrBotsFromFollowers ranges all followers and tried to detect bots or business accounts.
// These accounts could be blocked as they are not useful for statistic.
func (svc *Service) GetBusinessAccountsOrBotsFromFollowers(ctx context.Context) ([]models.User, error) {
	users, err := svc.GetFollowers(ctx)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	bType := getBarType(ctx)

	pBar := bar.New(len(users), bType)

	go pBar.Run(ctx)

	defer func() {
		pBar.Finish()
	}()

	followers := svc.instagram.client.Account.Followers()
	businessAccs := make([]models.User, 0, len(followers.Users))

	processResultChan := make(chan isBotResult)

	var (
		mu        sync.Mutex
		processWG sync.WaitGroup
	)

	go func(ctx context.Context, m *sync.Mutex) {
		for {
			select {
			case result := <-processResultChan:
				m.Lock()
				if result.isBot {
					businessAccs = append(businessAccs, result.user)
				}
				m.Unlock()
				pBar.Progress() <- struct{}{}
			case <-ctx.Done():
				return
			}
		}
	}(ctx, &mu)

	for svc.instagram.client.Account.Followers().Next() {
		for i := range svc.instagram.client.Account.Followers().Users {
			svc.processUser(ctx, &processWG, svc.instagram.client.Account.Followers().Users[i], processResultChan)
		}
	}

	if len(businessAccs) == 0 {
		return nil, makeNoUsersError(models.UsersBatchTypeBusinessAccounts)
	}

	return businessAccs, nil
}

func (svc *Service) processUser(ctx context.Context, group *sync.WaitGroup, u *goinsta.User,
	resultChan chan isBotResult) {
	group.Add(1)

	defer func() {
		group.Done()
	}()

	if ctx.Err() != nil {
		log.WithError(ctx, ctx.Err()).Error("Context canceled")

		return
	}

	resultChan <- isBotResult{
		user:  models.MakeUser(u.ID, u.Username, u.FullName),
		isBot: svc.isBotOrBusiness(ctx, u),
	}
}

func (svc *Service) isBotOrBusiness(ctx context.Context, user *goinsta.User) bool {
	user.SetInstagram(svc.instagram.client)

	const businessMarkNumFollowers = 500

	flws := user.Following()
	flwsNum := len(flws.Users)

	for flws.Next() {
		flwsNum += len(flws.Users)
	}

	log.WithFields(ctx, log.Fields{
		"username":   user.Username,
		"followings": flwsNum,
	}).Debug("Processing user")

	return flwsNum >= businessMarkNumFollowers ||
		user.CanBeReportedAsFraud ||
		user.HasAnonymousProfilePicture ||
		user.HasAffiliateShop ||
		user.HasPlacedOrders
}

// GetDiffFollowers returns batches with lost and new followers.
func (svc *Service) GetDiffFollowers(ctx context.Context) ([]models.UsersBatch, error) {
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		cancel()
	}()

	var noPreviousData bool

	bType := models.UsersBatchTypeFollowers

	oldBatch, err := svc.storage.GetLastUsersBatchByType(ctx, bType)
	if err != nil {
		if errors.Is(err, db.ErrNoData) {
			noPreviousData = true
		} else {
			return nil, fmt.Errorf("get last batch [%s]: %w", bType.String(), err)
		}
	}

	newList, err := svc.GetFollowers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get followers: %w", err)
	}

	if noPreviousData {
		oldBatch.Users = newList
	}

	lostFlw := getLostFollowers(oldBatch.Users, newList)
	lostBatch := models.UsersBatch{
		Users:     lostFlw,
		Type:      models.UsersBatchTypeLostFollowers,
		CreatedAt: time.Now(),
	}

	_, err = svc.storeUsers(ctx, lostBatch)
	if err != nil && !errors.Is(err, ErrNoUsers) {
		log.WithError(ctx, err).WithField("batch_type", lostBatch.Type.String()).Error("Failed store users")
	}

	newFlw := getNewFollowers(oldBatch.Users, newList)
	newBatch := models.UsersBatch{
		Users:     newFlw,
		Type:      models.UsersBatchTypeNewFollowers,
		CreatedAt: time.Now(),
	}

	_, err = svc.storeUsers(ctx, newBatch)
	if err != nil && !errors.Is(err, ErrNoUsers) {
		log.WithError(ctx, err).WithField("batch_type", newBatch.Type.String()).Error("Failed store users")
	}

	return []models.UsersBatch{lostBatch, newBatch}, nil
}

func getLostFollowers(oldlist []models.User, newlist []models.User) []models.User {
	var diff []models.User

	for _, oU := range oldlist {
		var found bool

		for _, nU := range newlist {
			if oU.ID == nU.ID {
				found = true

				break
			}
		}

		if !found {
			diff = append(diff, oU)
		}
	}

	return diff
}

func getNewFollowers(oldlist []models.User, newlist []models.User) []models.User {
	var diff []models.User

	for _, nU := range newlist {
		var found bool

		for _, oU := range oldlist {
			if oU.ID == nU.ID {
				found = true

				break
			}
		}

		if !found {
			diff = append(diff, nU)
		}
	}

	return diff
}
