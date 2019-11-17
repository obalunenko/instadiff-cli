// Package service implements instagram account operations and business logic.
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oleg-balunenko/instadiff-cli/internal/bar"
	"github.com/oleg-balunenko/instadiff-cli/internal/config"
	"github.com/oleg-balunenko/instadiff-cli/internal/db"
	"github.com/oleg-balunenko/instadiff-cli/internal/models"
)

// ErrLimitExceed returned when limit for action exceeded.
var ErrLimitExceed = errors.New("limit exceeded")

// Service represents service for operating instagram account.
type Service struct {
	instagramClient *goinsta.Instagram
	database        db.DB
	limits          limits
	whitelist       map[string]struct{}
	debug           bool
}

type limits struct {
	follow   int
	unFollow int
}

// New creates new instance of Service instance and returns closure func that will stop service.
//
// Usage:
// svc, stop, err := New(config.Config{})
// if err != nil{
// // handle error
// }
// defer stop()
func New(cfg config.Config) (*Service, func(), error) {
	cl := goinsta.New(cfg.Username(), cfg.Password())

	if err := cl.Login(); err != nil {
		return nil, nil, errors.Wrap(err, "failed to login")
	}

	log.Printf("logged in as %s \n", cl.Account.Username)

	lmts := limits{
		follow:   cfg.FollowLimits(),
		unFollow: cfg.UnFollowLimits(),
	}

	var dbc db.DB

	dbc, err := db.Connect(db.Params{
		LocalDB: cfg.IsLocalDBEnabled(),
		MongoParams: db.MongoParams{
			URL:        cfg.MongoConfigURL(),
			Database:   cfg.MongoDBName(),
			Collection: cfg.MongoDBCollection(),
		},
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to connect to db")
	}

	inst := &Service{
		instagramClient: cl,
		database:        dbc,
		limits:          lmts,
		whitelist:       cfg.Whitelist(),
		debug:           cfg.Debug(),
	}

	return inst, inst.stop, nil
}

// GetFollowers returns list of followers for logged in user.
func (svc *Service) GetFollowers() ([]models.User, error) {
	users := svc.instagramClient.Account.Followers()
	followers := makeUsersList(users)

	if len(followers) == 0 {
		return nil, errors.New("no followers")
	}

	err := svc.database.InsertUsersBatch(context.TODO(), models.UsersBatch{
		Users: followers,
		Type:  models.UsersBatchTypeFollowers,
	})
	if err != nil {
		log.Errorf("failed to insert %s to db: %v", models.UsersBatchTypeFollowers, err)
	}

	return followers, nil
}

// GetFollowings returns list of followings for logged in user.
func (svc *Service) GetFollowings() ([]models.User, error) {
	users := svc.instagramClient.Account.Following()

	followings := makeUsersList(users)

	if len(followings) == 0 {
		return nil, errors.New("no followings")
	}

	err := svc.database.InsertUsersBatch(context.TODO(), models.UsersBatch{
		Users: followings,
		Type:  models.UsersBatchTypeFollowings,
	})
	if err != nil {
		log.Errorf("failed to insert %s to db: %v", models.UsersBatchTypeFollowings, err)
	}

	return followings, nil
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
func (svc *Service) GetNotMutualFollowers() ([]models.User, error) {
	followers, err := svc.GetFollowers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get followers")
	}

	followings, err := svc.GetFollowings()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get followings")
	}

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

	err = svc.database.InsertUsersBatch(context.TODO(), models.UsersBatch{
		Users: notmutual,
		Type:  models.UsersBatchTypeNotMutual,
	})
	if err != nil {
		log.Errorf("Failed to insert %s in database: %v", models.UsersBatchTypeNotMutual, err)
	}

	return notmutual, nil
}

// UnFollow removes user from followings.
func (svc *Service) UnFollow(user models.User) error {
	log.Debugf("Unfollow user %s", user.UserName)

	if svc.debug {
		return nil
	}

	us := goinsta.User{ID: user.ID, Username: user.UserName}
	us.SetInstagram(svc.instagramClient)

	if err := us.Unfollow(); err != nil {
		return errors.Wrapf(err, "failed to unfollow user %v", user)
	}

	return nil
}

// Follow adds user to followings.
func (svc *Service) Follow(user models.User) error {
	log.Debugf("Follow user %s", user.UserName)

	if svc.debug {
		return nil
	}

	us := goinsta.User{ID: user.ID, Username: user.UserName}
	us.SetInstagram(svc.instagramClient)

	if err := us.Follow(); err != nil {
		return errors.Wrapf(err, "failed to follow user %v", user)
	}

	return nil
}

// UnFollowAllNotMutual clean followings from users that not following back.
func (svc *Service) UnFollowAllNotMutual() (int, error) {
	notMutual, err := svc.GetNotMutualFollowers()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get not mutual followers list")
	}

	if len(notMutual) == 0 {
		return 0, nil
	}

	var count int

	for _, nu := range notMutual {
		if err := svc.UnFollow(nu); err != nil {
			log.Printf("failed to unFollow user %v:%v", nu, err)
			continue
		}

		count++

		if count >= svc.limits.unFollow {
			return count, ErrLimitExceed
		}
	}

	return count, nil
}

// UnFollowAllNotMutualExceptWhitelisted clean followings from users that not following back
// except of whitelisted users.
func (svc *Service) UnFollowAllNotMutualExceptWhitelisted() (int, error) {
	notMutual, err := svc.GetNotMutualFollowers()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get not mutual followers list")
	}

	if len(notMutual) == 0 {
		return 0, nil
	}

	log.Debugf("Not mutual followers: %d", len(notMutual))

	ctx, cancel := context.WithCancel(context.TODO())

	defer func() {
		cancel()
	}()

	bType := bar.BTypeRendered
	if log.GetLevel() != log.InfoLevel {
		bType = bar.BTypeVoid
	}

	pBar := bar.New(len(notMutual), bType)

	go pBar.Run(ctx)

	defer func() {
		pBar.Finish()
	}()

	var count int

	for _, nu := range notMutual {
		pBar.Progress() <- struct{}{}

		if _, ok := svc.whitelist[nu.UserName]; !ok {
			if err := svc.UnFollow(nu); err != nil {
				log.Errorf("failed to unfollow [%s]: %v", nu.UserName, err)
				continue
			}
			count++
		}

		if count >= svc.limits.unFollow {
			return count, ErrLimitExceed
		}
	}

	return count, nil
}

// stop logs out from instagram and clean sessions.
// Should be called in defer after creating new instance from New().
func (svc *Service) stop() {
	if err := svc.instagramClient.Logout(); err != nil {
		log.Errorf("logout: %v", err)
	}
}

type isBotResult struct {
	user  models.User
	isBot bool
}

// GetBusinessAccountsOrBotsFromFollowers ranges all followers and tried to detect bots or business accounts.
// These accounts could be blocked as they are not useful for statistic.
func (svc *Service) GetBusinessAccountsOrBotsFromFollowers() ([]models.User, error) {
	users, err := svc.GetFollowers()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bType := bar.BTypeRendered
	if log.GetLevel() != log.InfoLevel {
		bType = bar.BTypeVoid
	}

	pBar := bar.New(len(users), bType)

	go pBar.Run(ctx)

	defer func() {
		pBar.Finish()
	}()

	followers := svc.instagramClient.Account.Followers()
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

	for followers.Next() {
		for i := range followers.Users {
			svc.processUser(ctx, &processWG, &followers.Users[i], processResultChan)
		}
	}

	if len(businessAccs) == 0 {
		return nil, errors.New("no business accounts")
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
		log.Errorf("canceled context: %v", ctx.Err())
		return
	}

	isBot := svc.isBotOrBusiness(u)
	resultChan <- isBotResult{
		user:  models.MakeUser(u.ID, u.Username, u.FullName),
		isBot: isBot,
	}
}

func (svc *Service) isBotOrBusiness(user *goinsta.User) bool {
	user.SetInstagram(svc.instagramClient)

	const businessMarkNumFollowers = 500

	flws := user.Following()
	flwsNum := len(flws.Users)

	for flws.Next() {
		flwsNum += len(flws.Users)
	}

	if flwsNum >= businessMarkNumFollowers {
		return true
	}

	fmt.Printf("processig[%s]: following[%d] \n", user.Username, flwsNum)

	if user.CanBeReportedAsFraud {
		return true
	}

	if user.HasAnonymousProfilePicture {
		return true
	}

	return false
}
