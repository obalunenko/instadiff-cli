// Package service implements instagram account operations and business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	log "github.com/obalunenko/logger"

	"github.com/obalunenko/instadiff-cli/internal/actions"

	"github.com/obalunenko/instadiff-cli/internal/client"
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
	incognito bool
}

type instagram struct {
	client    client.Client
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

func (i instagram) Client() client.Client {
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
// svc, err := New(config.Config{})
// if err != nil{
// // handle error
// }
// defer svc.Stop().
//
func New(ctx context.Context, cfg config.Config, cfgPath string, isIncognito bool) (*Service, error) {
	cl, err := client.New(ctx, cfgPath)
	if err != nil {
		return nil, fmt.Errorf("make client: %w", err)
	}

	uname := cl.Username(ctx)

	log.WithField(ctx, "username", uname).Info("Logged-in")

	stop := spinner.Set("Connecting to DB", "", "yellow")
	defer stop()

	dbc, err := db.Connect(ctx, db.Params{
		LocalDB: cfg.IsLocalDBEnabled(),
		MongoParams: db.MongoParams{
			URL:        cfg.MongoConfigURL(),
			Database:   cfg.MongoDBName(),
			Collection: db.BuildCollectionName(uname),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
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
		incognito: isIncognito,
	}

	return &svc, nil
}

// Stop stops the service and closes clients connections.
func (svc *Service) Stop(ctx context.Context) error {
	var errs error

	if err := svc.storage.Close(ctx); err != nil {
		errs = multierror.Append(errs, err)
	}

	if !svc.incognito {
		return errs
	}

	if err := svc.instagram.Client().Logout(ctx); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs
}

// GetFollowers returns list of followers for logged-in user.
func (svc *Service) GetFollowers(ctx context.Context) ([]models.User, error) {
	stop := spinner.Set("Fetching followers", "", "yellow")
	defer stop()

	bt := models.UsersBatchTypeFollowers

	return svc.getUsers(ctx, bt)
}

// GetFollowings returns list of followings for logged-in user.
func (svc *Service) GetFollowings(ctx context.Context) ([]models.User, error) {
	stop := spinner.Set("Fetching followings", "", "yellow")
	defer stop()

	bt := models.UsersBatchTypeFollowings

	return svc.getUsers(ctx, bt)
}

func (svc *Service) getUsers(ctx context.Context, bt models.UsersBatchType) ([]models.User, error) {
	var (
		users []models.User
		err   error
	)

	switch bt {
	case models.UsersBatchTypeFollowers:
		users, err = svc.instagram.client.Followers(ctx)
	case models.UsersBatchTypeFollowings:
		users, err = svc.instagram.client.Followings(ctx)
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

	return svc.actUser(ctx, user, actions.UserActionUnfollow, false)
}

// Follow adds user to followings.
func (svc *Service) Follow(ctx context.Context, user models.User) error {
	log.WithField(ctx, "username", user.UserName).Debug("Follow user")

	return svc.actUser(ctx, user, actions.UserActionFollow, false)
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

	return svc.unfollowUsers(ctx, notMutual, true)
}

// UnfollowUsers unfollows users by the name passed.
func (svc *Service) UnfollowUsers(ctx context.Context, usernames []string) (int, error) {
	var f userListProcessFunc = func(ctx context.Context, uslist []models.User) (int, error) {
		return svc.unfollowUsers(ctx, uslist, false)
	}

	return svc.processByUsernames(ctx, usernames, f)
}

// FollowUsers follows users by the name passed.
func (svc *Service) FollowUsers(ctx context.Context, usernames []string) (int, error) {
	var f userListProcessFunc = func(ctx context.Context, uslist []models.User) (int, error) {
		return svc.followUsers(ctx, uslist)
	}

	return svc.processByUsernames(ctx, usernames, f)
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

func (svc *Service) getUsersByUsername(ctx context.Context, usernames []string) ([]models.User, error) {
	stop := spinner.Set("Fetching users by names", "", "yellow")
	defer stop()

	if len(usernames) == 0 {
		return nil, ErrNoUsernamesPassed
	}

	users := make([]models.User, 0, len(usernames))

	for _, un := range usernames {
		u, err := svc.instagram.client.GetUserByName(ctx, un)
		if err != nil {
			return nil, fmt.Errorf("get user profile by name[%s]: %w", un, err)
		}

		users = append(users, u)
	}

	return users, nil
}

// RemoveFollowersByUsername removes all provided users by blocking and unblocking them to bypass Instagram limits.
func (svc *Service) RemoveFollowersByUsername(ctx context.Context, usernames []string) (int, error) {
	f := func(ctx context.Context, uslist []models.User) (int, error) {
		return svc.removeFollowers(ctx, uslist)
	}

	return svc.processByUsernames(ctx, usernames, f)
}

type userListProcessFunc func(ctx context.Context, uslist []models.User) (int, error)

func (svc *Service) processByUsernames(ctx context.Context, usernames []string, f userListProcessFunc) (int, error) {
	uslist, err := svc.getUsersByUsername(ctx, usernames)
	if err != nil {
		return 0, fmt.Errorf("get users by usernames: %w", err)
	}

	return f(ctx, uslist)
}

func makeProgressBar(ctx context.Context, capacity int) bar.Bar {
	pBar := bar.New(capacity, getBarType(ctx))

	go pBar.Run(ctx)

	return pBar
}

func (svc *Service) removeFollowers(ctx context.Context, users []models.User) (int, error) {
	return svc.actUsers(ctx, users, actions.UserActionRemove, false)
}

func (svc *Service) unfollowUsers(ctx context.Context, users []models.User, useWhitelist bool) (int, error) {
	return svc.actUsers(ctx, users, actions.UserActionUnfollow, useWhitelist)
}

func (svc *Service) followUsers(ctx context.Context, users []models.User) (int, error) {
	return svc.actUsers(ctx, users, actions.UserActionFollow, false)
}

func (svc *Service) actUsers(ctx context.Context, users []models.User, act actions.UserAction, useWhitelist bool) (int, error) {
	const (
		double    = 2
		errsLimit = 3
	)

	pBar := makeProgressBar(ctx, len(users)*double)
	defer pBar.Finish()

	var (
		skipped bool
		count   int
		errsNum int
	)

	for i, u := range users {
		if i != 0 && !skipped {
			time.Sleep(svc.instagram.Sleep())
		}

		skipped = false

		if ctx.Err() != nil {
			break
		}

		pBar.Progress() <- struct{}{}

		err := svc.actUser(ctx, u, act, useWhitelist)

		pBar.Progress() <- struct{}{}

		if err != nil {
			if errors.Is(err, ErrUserInWhitelist) {
				skipped = true

				continue
			}

			log.WithError(ctx, err).
				WithField("username", u.UserName).
				WithField("action", act.String()).
				Error("Failed to make action")

			errsNum++

			if errsNum >= errsLimit {
				return count, ErrCorrupted
			}

			continue
		}

		count++

		if count >= svc.instagram.limits.unFollow {
			return count, ErrLimitExceed
		}
	}

	return count, nil
}

func (svc *Service) actUser(ctx context.Context, u models.User, act actions.UserAction, useWhitelist bool) error {
	log.WithField(ctx, "action", act.String()).Debug("Action in progress")

	defer log.WithField(ctx, "action", act.String()).Debug("Action finished")

	whitelist := svc.instagram.Whitelist()

	canUseWhitelist := act == actions.UserActionUnfollow ||
		act == actions.UserActionBlock ||
		act == actions.UserActionRemove

	if canUseWhitelist && useWhitelist {
		if _, exist := whitelist[u.UserName]; exist {
			return ErrUserInWhitelist
		}

		const base = 10

		if _, exist := whitelist[strconv.FormatInt(u.ID, base)]; exist {
			return ErrUserInWhitelist
		}
	}

	var err error

	cli := svc.instagram.Client()

	switch act {
	case actions.UserActionFollow:
		err = cli.Follow(ctx, u)
	case actions.UserActionUnfollow:
		err = cli.Unfollow(ctx, u)
	case actions.UserActionBlock:
		err = cli.Block(ctx, u)
	case actions.UserActionUnblock:
		err = cli.Unblock(ctx, u)
	case actions.UserActionRemove:
		if err = cli.Block(ctx, u); err != nil {
			return fmt.Errorf("block user: %w", err)
		}

		if err = cli.Unblock(ctx, u); err != nil {
			return fmt.Errorf("unblock user: %w", err)
		}
	default:
		err = fmt.Errorf("unsupported action: %s", act.String())
	}

	return err
}

type isBotResult struct {
	user  models.User
	err   error
	isBot bool
}

// GetUselessFollowers ranges all followers and tried to detect bots or business accounts.
// These accounts could be blocked as they are not useful for statistic.
func (svc *Service) GetUselessFollowers(ctx context.Context) ([]models.User, error) {
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

	businessAccs := make([]models.User, 0, len(users))

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

	processWG.Add(len(users))

	for i := range users {
		u := users[i]

		go svc.processUser(ctx, &processWG, u, processResultChan)
	}

	processWG.Wait()

	if len(businessAccs) == 0 {
		return nil, makeNoUsersError(models.UsersBatchTypeUselessFollowers)
	}

	return businessAccs, nil
}

func (svc *Service) processUser(ctx context.Context, wg *sync.WaitGroup, u models.User, resultChan chan isBotResult) {
	defer wg.Done()

	if ctx.Err() != nil {
		resultChan <- isBotResult{
			user:  models.User{},
			err:   ctx.Err(),
			isBot: false,
		}

		return
	}

	isBot, err := svc.isUseless(ctx, u)
	if err != nil {
		resultChan <- isBotResult{
			user:  models.User{},
			err:   fmt.Errorf("check user[%s]: %w", u.UserName, err),
			isBot: isBot,
		}
	}

	resultChan <- isBotResult{
		user:  u,
		isBot: isBot,
		err:   nil,
	}
}

func (svc *Service) isUseless(ctx context.Context, user models.User) (bool, error) {
	const businessMarkNumFollowers = 500

	return svc.instagram.client.IsUseless(ctx, user, businessMarkNumFollowers)
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
