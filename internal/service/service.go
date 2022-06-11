// Package service implements instagram account operations and business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Davincible/goinsta"

	log "github.com/obalunenko/logger"

	"github.com/obalunenko/instadiff-cli/internal/config"
	"github.com/obalunenko/instadiff-cli/internal/db"
	"github.com/obalunenko/instadiff-cli/internal/models"
	"github.com/obalunenko/instadiff-cli/pkg/bar"
	"github.com/obalunenko/instadiff-cli/pkg/spinner"
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

	dbc, err := db.Connect(ctx, db.Params{
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

	stopFunc := func() error {
		return svc.stop()
	}

	return &svc, stopFunc, nil
}

// GetFollowers returns list of followers for logged-in user.
func (svc *Service) GetFollowers(ctx context.Context) ([]models.User, error) {
	stop := spinner.Set("Fetching followers", "", "yellow")
	defer stop()

	var noPreviousData bool

	bt := models.UsersBatchTypeFollowers

	oldBatch, err := svc.storage.GetLastUsersBatchByType(ctx, bt)
	if err != nil {
		if errors.Is(err, db.ErrNoData) {
			noPreviousData = true
		} else {
			return nil, fmt.Errorf("get last batch [%s]: %w", bt.String(), err)
		}
	}

	followers, err := makeUsersList(ctx, svc.instagram.client.Account.Followers())
	if err != nil {
		return nil, fmt.Errorf("make users list: %w", err)
	}

	if !noPreviousData {
		lostFlw := getLost(oldBatch.Users, followers)
		lostBatch := models.UsersBatch{
			Users:     lostFlw,
			Type:      models.UsersBatchTypeLostFollowers,
			CreatedAt: time.Now(),
		}

		_, err = svc.storeUsers(ctx, lostBatch)
		if err != nil && !errors.Is(err, ErrNoUsers) {
			log.WithError(ctx, err).WithField("batch_type", lostBatch.Type.String()).Error("Failed store users")
		}

		newFlw := getNew(oldBatch.Users, followers)
		newBatch := models.UsersBatch{
			Users:     newFlw,
			Type:      models.UsersBatchTypeNewFollowers,
			CreatedAt: time.Now(),
		}

		_, err = svc.storeUsers(ctx, newBatch)
		if err != nil && !errors.Is(err, ErrNoUsers) {
			log.WithError(ctx, err).WithField("batch_type", newBatch.Type.String()).Error("Failed store users")
		}
	}

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

	var noPreviousData bool

	bt := models.UsersBatchTypeFollowings

	oldBatch, err := svc.storage.GetLastUsersBatchByType(ctx, bt)
	if err != nil {
		if errors.Is(err, db.ErrNoData) {
			noPreviousData = true
		} else {
			return nil, fmt.Errorf("get last batch [%s]: %w", bt.String(), err)
		}
	}

	followings, err := makeUsersList(ctx, svc.instagram.client.Account.Followers())
	if err != nil {
		return nil, fmt.Errorf("make users list: %w", err)
	}

	if !noPreviousData {
		lostFlw := getLost(oldBatch.Users, followings)
		lostBatch := models.UsersBatch{
			Users:     lostFlw,
			Type:      models.UsersBatchTypeLostFollowings,
			CreatedAt: time.Now(),
		}

		_, err = svc.storeUsers(ctx, lostBatch)
		if err != nil && !errors.Is(err, ErrNoUsers) {
			log.WithError(ctx, err).WithField("batch_type", lostBatch.Type.String()).Error("Failed store users")
		}

		newFlw := getNew(oldBatch.Users, followings)
		newBatch := models.UsersBatch{
			Users:     newFlw,
			Type:      models.UsersBatchTypeNewFollowings,
			CreatedAt: time.Now(),
		}

		_, err = svc.storeUsers(ctx, newBatch)
		if err != nil && !errors.Is(err, ErrNoUsers) {
			log.WithError(ctx, err).WithField("batch_type", newBatch.Type.String()).Error("Failed store users")
		}
	}

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

func makeUsersList(ctx context.Context, users *goinsta.Users) ([]models.User, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	usersList := make([]models.User, 0, len(users.Users))

	for users.Next() {
		for i := range users.Users {
			usersList = append(usersList,
				models.MakeUser(users.Users[i].ID, users.Users[i].Username, users.Users[i].FullName))
		}
	}

	if err := users.Error(); err != nil {
		if !errors.Is(err, goinsta.ErrNoMore) {
			return nil, fmt.Errorf("users iterate: %w", err)
		}
	}

	return usersList, nil
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
	if ctx.Err() != nil {
		return ctx.Err()
	}

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
	if ctx.Err() != nil {
		return ctx.Err()
	}

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
		return 0, makeNoUsersError(models.UsersBatchTypeNotMutual)
	}

	log.WithFields(ctx, log.Fields{
		"count":       len(notMutual),
		"whitelisted": len(svc.instagram.Whitelist()),
	}).Info("Not mutual followers")

	diff := svc.whitelistNotMutual(notMutual)

	if len(diff) == 0 {
		return 0, makeNoUsersError(models.UsersBatchTypeNotMutual)
	}

	return svc.unfollowUsers(ctx, notMutual)
}

func (svc *Service) whitelistNotMutual(notMutual []models.User) []models.User {
	result := notMutual[:0]

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
	return svc.removeFollowers(ctx, usernames)
}

func makeProgressBar(ctx context.Context, capacity int) bar.Bar {
	pBar := bar.New(capacity, getBarType(ctx))

	go pBar.Run(ctx)

	return pBar
}

func (svc *Service) removeFollowers(ctx context.Context, users []string) (int, error) {
	if len(users) == 0 {
		return 0, nil
	}

	const double = 2

	pBar := makeProgressBar(ctx, len(users)*double)
	defer pBar.Finish()

	var count int

	ticker := time.NewTicker(svc.instagram.sleep)
	defer ticker.Stop()

	const errsLimit = 3

	var errsNum int

LOOP:
	for i, un := range users {
		if errsNum >= errsLimit {
			return count, ErrCorrupted
		}

		if i == 0 {
			pBar.Progress() <- struct{}{}

			err := svc.removeUser(un)

			pBar.Progress() <- struct{}{}

			if err != nil {
				log.WithError(ctx, err).WithField("username", un).Error("Failed to remove follower")

				errsNum++

				continue
			}
		}

		select {
		case <-ctx.Done():
			break LOOP
		case <-ticker.C:
			pBar.Progress() <- struct{}{}

			err := svc.removeUser(un)

			pBar.Progress() <- struct{}{}

			if err != nil {
				log.WithError(ctx, err).WithField("username", un).Error("Failed to remove follower")

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

func (svc *Service) removeUser(username string) error {
	u, err := svc.instagram.client.Profiles.ByName(username)
	if err != nil {
		return fmt.Errorf("user lookup: %w", err)
	}

	u.SetInstagram(svc.instagram.client)

	if err = u.Block(false); err != nil {
		return fmt.Errorf("block user: %w", err)
	}

	if err = u.Unblock(); err != nil {
		return fmt.Errorf("unblock user: %w", err)
	}

	return nil
}

func (svc *Service) unfollowUsers(ctx context.Context, users []models.User) (int, error) {
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	if len(users) == 0 {
		return 0, makeNoUsersError(models.UsersBatchTypeNotMutual)
	}

	const double = 2

	pBar := makeProgressBar(ctx, len(users)*double)
	defer pBar.Finish()

	var count int

	ticker := time.NewTicker(svc.instagram.sleep)
	defer ticker.Stop()

	const errsLimit = 3

	var errsNum int

LOOP:
	for i, u := range users {
		if errsNum >= errsLimit {
			return count, ErrCorrupted
		}

		if i == 0 {
			pBar.Progress() <- struct{}{}

			err := svc.unfollowWhitelist(ctx, u)

			pBar.Progress() <- struct{}{}
			if err != nil {
				log.WithError(ctx, err).WithField("username", u.UserName).Error("Failed to unfollow")
				errsNum++

				continue
			}

			count++
		}

		select {
		case <-ctx.Done():
			break LOOP
		case <-ticker.C:
			pBar.Progress() <- struct{}{}

			err := svc.unfollowWhitelist(ctx, u)

			pBar.Progress() <- struct{}{}
			if err != nil {
				log.WithError(ctx, err).WithField("username", u.UserName).Error("Failed to unfollow")
				errsNum++

				continue
			}

			count++
		}

		if count >= svc.instagram.limits.unFollow {
			return count, ErrLimitExceed
		}
	}

	return count, nil
}

func (svc *Service) unfollowWhitelist(ctx context.Context, u models.User) error {
	whitelist := svc.instagram.Whitelist()

	if _, exist := whitelist[u.UserName]; exist {
		return nil
	}

	if _, exist := whitelist[strconv.FormatInt(u.ID, 10)]; exist {
		return nil
	}

	if err := svc.UnFollow(ctx, u); err != nil {
		return err
	}

	return nil
}

// stop logs out from instagram and clean sessions.
// Should be called in defer after creating new instance from New().
func (svc *Service) stop() error {
	if err := svc.instagram.client.Logout(); err != nil {
		// wierd error - just ignore it.
		if strings.Contains(err.Error(), "405 Method Not Allowed") {
			return nil
		}

		return fmt.Errorf("logout: %w", err)
	}

	return nil
}

type isBotResult struct {
	user  models.User
	err   error
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
				pBar.Progress() <- struct{}{}

				if result.err != nil {
					log.WithError(ctx, err).Error("Failed to check if user")

					continue
				}

				m.Lock()
				if result.isBot {
					businessAccs = append(businessAccs, result.user)
				}
				m.Unlock()
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

	if err := svc.instagram.client.Account.Followers().Error(); err != nil {
		if !errors.Is(err, goinsta.ErrNoMore) {
			return nil, fmt.Errorf("users iterate: %w", err)
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
		resultChan <- isBotResult{
			user:  models.User{},
			err:   ctx.Err(),
			isBot: false,
		}

		return
	}

	isBot, err := svc.isBotOrBusiness(ctx, u)
	if err != nil {
		resultChan <- isBotResult{
			user:  models.User{},
			err:   fmt.Errorf("check user[%s]: %w", u.Username, err),
			isBot: isBot,
		}
	}

	resultChan <- isBotResult{
		user:  models.MakeUser(u.ID, u.Username, u.FullName),
		isBot: isBot,
		err:   nil,
	}
}

func (svc *Service) isBotOrBusiness(ctx context.Context, user *goinsta.User) (bool, error) {
	user.SetInstagram(svc.instagram.client)

	const businessMarkNumFollowers = 500

	flws := user.Following()
	flwsNum := len(flws.Users)

	for flws.Next() {
		flwsNum += len(flws.Users)
	}

	if err := flws.Error(); err != nil {
		if !errors.Is(err, goinsta.ErrNoMore) {
			return false, fmt.Errorf("users iterate: %w", err)
		}
	}

	log.WithFields(ctx, log.Fields{
		"username":   user.Username,
		"followings": flwsNum,
	}).Debug("Processing user")

	return flwsNum >= businessMarkNumFollowers ||
		user.CanBeReportedAsFraud ||
		user.HasAnonymousProfilePicture ||
		user.HasAffiliateShop ||
		user.HasPlacedOrders, nil
}

// GetDiffFollowers returns batches with lost and new followers.
func (svc *Service) GetDiffFollowers(ctx context.Context) ([]models.UsersBatch, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	lostBatch, err := svc.storage.GetLastUsersBatchByType(ctx, models.UsersBatchTypeLostFollowers)
	if err != nil {
		if errors.Is(err, ErrNoUsers) {
			lostBatch = models.UsersBatch{
				Users:     nil,
				Type:      models.UsersBatchTypeLostFollowers,
				CreatedAt: time.Now(),
			}
		} else {
			return nil, fmt.Errorf("get lost folowers: %w", err)
		}
	}

	newBatch, err := svc.storage.GetLastUsersBatchByType(ctx, models.UsersBatchTypeNewFollowers)
	if err != nil {
		if errors.Is(err, ErrNoUsers) {
			lostBatch = models.UsersBatch{
				Users:     nil,
				Type:      models.UsersBatchTypeNewFollowers,
				CreatedAt: time.Now(),
			}
		} else {
			return nil, fmt.Errorf("get new folowers: %w", err)
		}
	}

	return []models.UsersBatch{lostBatch, newBatch}, nil
}

func getLost(oldlist, newlist []models.User) []models.User {
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

func getNew(oldlist, newlist []models.User) []models.User {
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
