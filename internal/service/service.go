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
	incognito bool
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
func New(ctx context.Context, cfg config.Config, cfgPath string, isDebug, isIncognito bool) (*Service, StopFunc, error) {
	cl, lf, err := makeClient(ctx, cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("make client: %w", err)
	}

	uname := cl.Account.Username

	log.WithField(ctx, "username", uname).Info("Logged-in")

	if err = cl.OpenApp(); err != nil {
		log.WithError(ctx, err).Error("Failed to refresh app info")
	}

	dbc, err := db.Connect(ctx, db.Params{
		LocalDB: cfg.IsLocalDBEnabled(),
		MongoParams: db.MongoParams{
			URL:        cfg.MongoConfigURL(),
			Database:   cfg.MongoDBName(),
			Collection: db.BuildCollectionName(uname),
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
		storage:   dbc,
		debug:     isDebug,
		incognito: isIncognito,
	}

	return &svc, svc.stop(lf), nil
}

func (svc *Service) stop(f logoutFunc) StopFunc {
	return func() error {
		if !svc.incognito {
			return nil
		}

		return f()
	}
}

// GetFollowers returns list of followers for logged-in user.
func (svc *Service) GetFollowers(ctx context.Context) ([]models.User, error) {
	stop := spinner.Set("Fetching followers", "", "yellow")
	defer stop()

	bt := models.UsersBatchTypeFollowers

	return svc.getFollowersFollowings(ctx, bt)
}

// GetFollowings returns list of followings for logged-in user.
func (svc *Service) GetFollowings(ctx context.Context) ([]models.User, error) {
	stop := spinner.Set("Fetching followings", "", "yellow")
	defer stop()

	bt := models.UsersBatchTypeFollowings

	return svc.getFollowersFollowings(ctx, bt)
}

func (svc *Service) getFollowersFollowings(ctx context.Context, bt models.UsersBatchType) ([]models.User, error) {
	var (
		users []models.User
		err   error
	)

	switch bt {
	case models.UsersBatchTypeFollowers:
		users, err = makeUsersList(ctx, svc.instagram.client.Account.Followers())
	case models.UsersBatchTypeFollowings:
		users, err = makeUsersList(ctx, svc.instagram.client.Account.Following())
	default:
		return nil, fmt.Errorf("not supported batch type for this func: %s", bt.String())
	}

	if err != nil {
		return nil, fmt.Errorf("make users list: %w", err)
	}

	if err = svc.findDiffUsers(ctx, users, bt); err != nil {
		return nil, fmt.Errorf("find diff users: %w", err)
	}

	err = svc.storeUsers(ctx, models.UsersBatch{
		Users:     users,
		Type:      bt,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("store users [%s]: %w", bt.String(), err)
	}

	return users, nil
}

func (svc *Service) findDiffUsers(ctx context.Context, users []models.User, bt models.UsersBatchType) error {
	var lbt, nbt models.UsersBatchType

	now := time.Now()

	switch bt {
	case models.UsersBatchTypeFollowers:
		lbt, nbt = models.UsersBatchTypeLostFollowers, models.UsersBatchTypeNewFollowers
	case models.UsersBatchTypeFollowings:
		lbt, nbt = models.UsersBatchTypeLostFollowings, models.UsersBatchTypeNewFollowings
	default:
		return fmt.Errorf("not supported batch type for this func: %s", bt.String())
	}

	oldBatch, err := svc.storage.GetLastUsersBatchByType(ctx, bt)
	if err != nil {
		if errors.Is(err, db.ErrNoData) {
			return nil
		}

		return fmt.Errorf("get last batch [%s]: %w", bt.String(), err)
	}

	lostFlw := getLost(oldBatch.Users, users)

	lostBatch := models.UsersBatch{
		Users:     lostFlw,
		Type:      lbt,
		CreatedAt: now,
	}

	if err = svc.storeUsers(ctx, lostBatch); err != nil && !errors.Is(err, ErrNoUsers) {
		return fmt.Errorf("store users [%s]: %w", lostBatch.Type, err)
	}

	newFlw := getNew(oldBatch.Users, users)

	newBatch := models.UsersBatch{
		Users:     newFlw,
		Type:      nbt,
		CreatedAt: now,
	}

	if err = svc.storeUsers(ctx, newBatch); err != nil && !errors.Is(err, ErrNoUsers) {
		return fmt.Errorf("store users [%s]: %w", newBatch.Type, err)
	}

	return nil
}

func (svc *Service) storeUsers(ctx context.Context, batch models.UsersBatch) error {
	if len(batch.Users) == 0 {
		return makeNoUsersError(batch.Type)
	}

	err := svc.storage.InsertUsersBatch(ctx, batch)
	if err != nil {
		return fmt.Errorf("insert users batch [%s]: %w", batch.Type.String(), err)
	}

	return nil
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

	bt := models.UsersBatchTypeNotMutual

	err = svc.storeUsers(ctx, models.UsersBatch{
		Users:     notmutual,
		Type:      bt,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("store users [%s]: %w", bt, err)
	}

	return notmutual, nil
}

// UnFollow removes user from followings.
func (svc *Service) UnFollow(ctx context.Context, user models.User) error {
	log.WithField(ctx, "username", user.UserName).Debug("Unfollow user")

	return svc.actUser(ctx, user, userActionUnfollow)
}

// Follow adds user to followings.
func (svc *Service) Follow(ctx context.Context, user models.User) error {
	log.WithField(ctx, "username", user.UserName).Debug("Follow user")

	return svc.actUser(ctx, user, userActionFollow)
}

//go:generate stringer -type=userAction -trimprefix=userAction

type userAction uint

const (
	userActionUnknown userAction = iota

	userActionFollow
	userActionUnfollow
	userActionBlock
	userActionUnblock

	userActionSentinel
)

func (svc *Service) actUser(ctx context.Context, user models.User, act userAction) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if svc.debug {
		return nil
	}

	us := goinsta.User{
		ID:       user.ID,
		Username: user.UserName,
	}

	us.SetInstagram(svc.instagram.client)

	var f func() error

	switch act {
	case userActionFollow:
		f = us.Follow
	case userActionUnfollow:
		f = us.Unfollow
	case userActionBlock:
		f = func() error {
			return us.Block(false)
		}
	case userActionUnblock:
		f = us.Unblock
	default:
		return fmt.Errorf("unsupported user action type: %s", act.String())
	}

	if err := f(); err != nil {
		return fmt.Errorf("action[%s]: %w", act.String(), err)
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

func (svc *Service) UnfollowUsers(ctx context.Context, usernames []string) (int, error) {
	if len(usernames) == 0 {
		return 0, errors.New("no usernames passed")
	}

	uslist, err := svc.getUsersByUsername(usernames)
	if err != nil {
		return 0, fmt.Errorf("get users by usernames: %w", err)
	}

	for i := range uslist {
		u := uslist[i]

		// TODO: add progress bar and sleep for ddos ban prevent.
		//  Add count

		if err = svc.actUser(ctx, u, userActionUnfollow); err != nil {
			return 0, fmt.Errorf("act user[%s]: %w", u.UserName, err)
		}
	}

	return 0, nil
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

func (svc *Service) getUsersByUsername(usernames []string) ([]models.User, error) {
	stop := spinner.Set("Fetching users by names", "", "yellow")
	defer stop()

	users := make([]models.User, 0, len(usernames))

	for _, un := range usernames {
		u, err := svc.instagram.client.Profiles.ByName(un)
		if err != nil {
			return nil, fmt.Errorf("get user profile by name[%s]: %w", un, err)
		}

		users = append(users, models.MakeUser(u.ID, u.Username, u.FullName))
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

			err := svc.removeUser(ctx, un)

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

			err := svc.removeUser(ctx, un)

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

func (svc *Service) removeUser(ctx context.Context, username string) error {
	u, err := svc.instagram.client.Profiles.ByName(username)
	if err != nil {
		return fmt.Errorf("user lookup: %w", err)
	}

	um := models.User{
		ID:       u.ID,
		UserName: u.Username,
		FullName: u.FullName,
	}

	if err = svc.actUser(ctx, um, userActionBlock); err != nil {
		return fmt.Errorf("action user: %w", err)
	}

	if err = svc.actUser(ctx, um, userActionUnblock); err != nil {
		return fmt.Errorf("action user: %w", err)
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

	if err := svc.actUser(ctx, u, userActionUnfollow); err != nil {
		return err
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

	dt := models.DiffTypeFollowers

	return svc.getDiff(ctx, dt)
}

// GetDiffFollowings returns batches with lost and new followings.
func (svc *Service) GetDiffFollowings(ctx context.Context) ([]models.UsersBatch, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dt := models.DiffTypeFollowings

	return svc.getDiff(ctx, dt)
}

func (svc *Service) getDiff(ctx context.Context, dt models.DiffType) ([]models.UsersBatch, error) {
	var batchTypes []models.UsersBatchType

	switch dt {
	case models.DiffTypeFollowers:
		batchTypes = []models.UsersBatchType{models.UsersBatchTypeNewFollowers, models.UsersBatchTypeLostFollowers}
	case models.DiffTypeFollowings:
		batchTypes = []models.UsersBatchType{models.UsersBatchTypeNewFollowings, models.UsersBatchTypeLostFollowings}
	default:
		return nil, fmt.Errorf("unsupported diff type [%s]", dt.String())
	}

	const respnum = 2

	resp := make([]models.UsersBatch, 0, respnum)

	for i := range batchTypes {
		bt := batchTypes[i]

		users, err := svc.storage.GetLastUsersBatchByType(ctx, bt)
		if err != nil {
			if errors.Is(err, db.ErrNoData) {
				users = models.UsersBatch{
					Users:     nil,
					Type:      bt,
					CreatedAt: time.Now(),
				}
			} else {
				return nil, fmt.Errorf("get users [%s]: %w", bt.String(), err)
			}
		}

		resp = append(resp, users)
	}

	return resp, nil
}

// GetHistoryDiffFollowings returns diff history of followings for an account.
func (svc *Service) GetHistoryDiffFollowings(ctx context.Context) (models.DiffHistory, error) {
	dt := models.DiffTypeFollowings

	if ctx.Err() != nil {
		return models.DiffHistory{}, ctx.Err()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return svc.getHistoryDiff(ctx, dt)
}

// GetHistoryDiffFollowers returns diff history of followers for an account.
func (svc *Service) GetHistoryDiffFollowers(ctx context.Context) (models.DiffHistory, error) {
	dt := models.DiffTypeFollowers

	if ctx.Err() != nil {
		return models.DiffHistory{}, ctx.Err()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return svc.getHistoryDiff(ctx, dt)
}

func (svc *Service) getHistoryDiff(ctx context.Context, dt models.DiffType) (models.DiffHistory, error) {
	var batchTypes []models.UsersBatchType

	switch dt {
	case models.DiffTypeFollowers:
		batchTypes = []models.UsersBatchType{models.UsersBatchTypeNewFollowers, models.UsersBatchTypeLostFollowers}
	case models.DiffTypeFollowings:
		batchTypes = []models.UsersBatchType{models.UsersBatchTypeNewFollowings, models.UsersBatchTypeLostFollowings}
	default:
		return models.DiffHistory{}, fmt.Errorf("unsupported diff type [%s]", dt.String())
	}

	resp := models.MakeDiffHistory(dt)

	for i := range batchTypes {
		bt := batchTypes[i]

		users, err := svc.storage.GetAllUsersBatchByType(ctx, bt)
		if err != nil {
			if errors.Is(err, db.ErrNoData) {
				users = []models.UsersBatch{{
					Users:     nil,
					Type:      bt,
					CreatedAt: time.Now(),
				}}
			} else {
				return models.DiffHistory{}, fmt.Errorf("get users [%s]: %w", bt.String(), err)
			}
		}

		resp.Add(users...)
	}

	return resp, nil
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
