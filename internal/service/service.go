// Package service implements instagram account operations and business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	log "github.com/sirupsen/logrus"
	"github.com/tcnksm/go-input"

	"github.com/oleg-balunenko/instadiff-cli/internal/config"
	"github.com/oleg-balunenko/instadiff-cli/internal/db"
	"github.com/oleg-balunenko/instadiff-cli/internal/models"
	"github.com/oleg-balunenko/instadiff-cli/pkg/bar"
)

var (
	// ErrLimitExceed returned when limit for action exceeded.
	ErrLimitExceed = errors.New("limit exceeded")
	// ErrCorrupted returned when instagram returned error response more than one time during loop processing.
	ErrCorrupted = errors.New("unable to continue - instagram responses with errors")
)

// Service represents service for operating instagram account.
type Service struct {
	ctx       context.Context
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
type StopFunc func()

// New creates new instance of Service instance and returns closure func that will stop service.
//
// Usage:
// svc, stop, err := New(config.Config{})
// if err != nil{
// // handle error
// }
// defer stop()
//
func New(ctx context.Context, cfg config.Config, cfgPath string) (*Service, StopFunc, error) {
	cl, err := makeClient(cfg, cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("make client: %w", err)
	}

	log.Printf("logged in as %s \n", cl.Account.Username)

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
		ctx: ctx,
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

	stopFunc := func() {
		svc.stop()
	}

	return &svc, stopFunc, nil
}

func makeClient(cfg config.Config, cfgPath string) (*goinsta.Instagram, error) {
	var (
		cl *goinsta.Instagram
	)

	sessFile := filepath.Join(cfgPath, fmt.Sprintf("%s.sess", cfg.Username()))

	if i, err := goinsta.Import(sessFile); err == nil {
		log.Infof("session imported from file: %s", sessFile)

		cl = i

		return i, nil
	}

	cl = goinsta.New(cfg.Username(), cfg.Password())

	if err := cl.Login(); err != nil {
		switch v := err.(type) {
		case goinsta.ChallengeError:
			cl, err = challenge(cl, v.Challenge.APIPath)
			if err != nil {
				return nil, fmt.Errorf("challenge: %w", err)
			}

		default:
			return nil, fmt.Errorf("failed to login: %w", err)
		}
	}

	if cfg.StoreSession() {
		if err := cl.Export(sessFile); err != nil {
			log.Errorf("save session: %v", err)
		}
	}

	return cl, nil
}

func challenge(cl *goinsta.Instagram, chURL string) (*goinsta.Instagram, error) {
	if err := cl.Challenge.Process(chURL); err != nil {
		return nil, fmt.Errorf("process challenge: %w", err)
	}

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	code, err := ui.Ask("What is SMS code for instagram?",
		&input.Options{
			Default:  "000000",
			Required: true,
			Loop:     true,
		})
	if err != nil {
		return nil, fmt.Errorf("process input: %w", err)
	}

	if err = cl.Challenge.SendSecurityCode(code); err != nil {
		return nil, fmt.Errorf("send security code: %w", err)
	}

	cl.Account = cl.Challenge.LoggedInUser

	return cl, nil
}

// GetFollowers returns list of followers for logged in user.
func (svc *Service) GetFollowers() ([]models.User, error) {
	followers := makeUsersList(svc.instagram.client.Account.Followers())

	if len(followers) == 0 {
		return nil, makeNoUsersError(models.UsersBatchTypeFollowers)
	}

	err := svc.storage.InsertUsersBatch(svc.ctx, models.UsersBatch{
		Users:     followers,
		Type:      models.UsersBatchTypeFollowers,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Errorf("failed to insert %s to db: %v", models.UsersBatchTypeFollowers, err)
	}

	return followers, nil
}

// GetFollowings returns list of followings for logged in user.
func (svc *Service) GetFollowings() ([]models.User, error) {
	followings := makeUsersList(svc.instagram.client.Account.Following())

	if len(followings) == 0 {
		return nil, makeNoUsersError(models.UsersBatchTypeFollowings)
	}

	err := svc.storage.InsertUsersBatch(svc.ctx, models.UsersBatch{
		Users:     followings,
		Type:      models.UsersBatchTypeFollowings,
		CreatedAt: time.Now(),
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
		return nil, fmt.Errorf("get followers: %w", err)
	}

	log.Infof("Total followers: %d", len(followers))

	followings, err := svc.GetFollowings()
	if err != nil {
		return nil, fmt.Errorf("get followings: %w", err)
	}

	log.Infof("Total followings: %d", len(followings))

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

	err = svc.storage.InsertUsersBatch(svc.ctx, models.UsersBatch{
		Users:     notmutual,
		Type:      models.UsersBatchTypeNotMutual,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Errorf("Failed to insert %s in storage: %v", models.UsersBatchTypeNotMutual, err)
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
	us.SetInstagram(svc.instagram.client)

	if err := us.Unfollow(); err != nil {
		return fmt.Errorf("[%s] unfollow: %w", user.UserName, err)
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
	us.SetInstagram(svc.instagram.client)

	if err := us.Follow(); err != nil {
		return fmt.Errorf("[%s] follow: %w", user.UserName, err)
	}

	return nil
}

func getBarType() bar.BType {
	bType := bar.BTypeRendered
	if log.GetLevel() != log.InfoLevel {
		bType = bar.BTypeVoid
	}

	return bType
}

// UnFollowAllNotMutualExceptWhitelisted clean followings from users that not following back
// except of whitelisted users.
func (svc *Service) UnFollowAllNotMutualExceptWhitelisted() (int, error) {
	notMutual, err := svc.GetNotMutualFollowers()
	if err != nil {
		return 0, fmt.Errorf("get not mutual followers: %w", err)
	}

	if len(notMutual) == 0 {
		return 0, nil
	}

	log.Infof("Not mutual followers: %d", len(notMutual))
	log.Infof("Whitelisted: %d", len(svc.instagram.Whitelist()))

	pBar := bar.New(len(notMutual)-len(svc.instagram.Whitelist()), getBarType())

	go pBar.Run(svc.ctx)

	defer func() {
		pBar.Finish()
	}()

	return svc.processNotMutual(pBar, notMutual)
}

func (svc *Service) processNotMutual(pBar bar.Bar, notMutual []models.User) (int, error) {
	var count int

	ticker := time.NewTicker(svc.instagram.sleep)
	defer ticker.Stop()

	const errsLimit = 3

	var errsNum int

	whitelist := svc.instagram.Whitelist()

LOOP:
	for i, nu := range notMutual {
		if errsNum >= errsLimit {
			return count, ErrCorrupted
		}

		if i != 0 {
			pBar.Progress() <- struct{}{}
		}

		select {
		case <-svc.ctx.Done():
			break LOOP
		case <-ticker.C:
			if _, ok := whitelist[nu.UserName]; !ok {
				if err := svc.UnFollow(nu); err != nil {
					log.Errorf("failed to unfollow [%s]: %v", nu.UserName, err)
					errsNum++
					continue
				}
				count++
			}

			if count >= svc.instagram.limits.unFollow {
				return count, ErrLimitExceed
			}
		}
	}

	return count, nil
}

// stop logs out from instagram and clean sessions.
// Should be called in defer after creating new instance from New().
func (svc *Service) stop() {
	if err := svc.instagram.client.Logout(); err != nil {
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

	ctx, cancel := context.WithCancel(svc.ctx)
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
			svc.processUser(ctx, &processWG, &svc.instagram.client.Account.Followers().Users[i], processResultChan)
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
		log.Errorf("canceled context: %v", ctx.Err())
		return
	}

	resultChan <- isBotResult{
		user:  models.MakeUser(u.ID, u.Username, u.FullName),
		isBot: svc.isBotOrBusiness(u),
	}
}

func (svc *Service) isBotOrBusiness(user *goinsta.User) bool {
	user.SetInstagram(svc.instagram.client)

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

// GetDiffFollowers returns batches with lost and new followers.
func (svc *Service) GetDiffFollowers() ([]models.UsersBatch, error) {
	ctx, cancel := context.WithCancel(svc.ctx)

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

	newList, err := svc.GetFollowers()
	if err != nil {
		return nil, fmt.Errorf("get followers: %w", err)
	}

	now := time.Now()

	if noPreviousData {
		oldBatch.Users = newList
	}

	lostFlw := getLostFollowers(oldBatch.Users, newList)
	lostBatch := models.UsersBatch{
		Users:     lostFlw,
		Type:      models.UsersBatchTypeLostFollowers,
		CreatedAt: now,
	}

	if err := svc.storage.InsertUsersBatch(ctx, lostBatch); err != nil {
		log.Errorf("Failed to insert [%s]: %v", lostBatch.Type.String(), err)
	}

	newFlw := getNewFollowers(oldBatch.Users, newList)
	newBatch := models.UsersBatch{
		Users:     newFlw,
		Type:      models.UsersBatchTypeNewFollowers,
		CreatedAt: now,
	}

	if err := svc.storage.InsertUsersBatch(ctx, newBatch); err != nil {
		log.Errorf("Failed to insert [%s]: %v", newBatch.Type.String(), err)
	}

	return []models.UsersBatch{lostBatch, newBatch}, nil
}

func getLostFollowers(old []models.User, new []models.User) []models.User {
	var diff []models.User

	for _, oU := range old {
		var found bool

		for _, nU := range new {
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

func getNewFollowers(old []models.User, new []models.User) []models.User {
	var diff []models.User

	for _, nU := range new {
		var found bool

		for _, oU := range old {
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
