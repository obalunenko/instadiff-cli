// Package service implements instagram account operations and business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/TheForgotten69/goinsta/v2"
	log "github.com/sirupsen/logrus"

	"github.com/obalunenko/instadiff-cli/internal/config"
	"github.com/obalunenko/instadiff-cli/internal/db"
	"github.com/obalunenko/instadiff-cli/internal/models"
	"github.com/obalunenko/instadiff-cli/pkg/bar"
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
// defer stop().
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

	us := goinsta.User{
		ID:                         user.ID,
		Username:                   user.UserName,
		FullName:                   "",
		Biography:                  "",
		ProfilePicURL:              "",
		Email:                      "",
		PhoneNumber:                "",
		IsBusiness:                 false,
		Gender:                     0,
		ProfilePicID:               "",
		HasAnonymousProfilePicture: false,
		IsPrivate:                  false,
		IsUnpublished:              false,
		AllowedCommenterType:       "",
		IsVerified:                 false,
		MediaCount:                 0,
		FollowerCount:              0,
		FollowingCount:             0,
		FollowingTagCount:          0,
		MutualFollowersID:          nil,
		ProfileContext:             "",
		GeoMediaCount:              0,
		ExternalURL:                "",
		HasBiographyTranslation:    false,
		ExternalLynxURL:            "",
		BiographyWithEntities: struct {
			RawText  string        `json:"raw_text"`
			Entities []interface{} `json:"entities"`
		}{},
		UsertagsCount:                0,
		HasChaining:                  false,
		IsFavorite:                   false,
		IsFavoriteForStories:         false,
		IsFavoriteForHighlights:      false,
		CanBeReportedAsFraud:         false,
		ShowShoppableFeed:            false,
		ShoppablePostsCount:          0,
		ReelAutoArchive:              "",
		HasHighlightReels:            false,
		PublicEmail:                  "",
		PublicPhoneNumber:            "",
		PublicPhoneCountryCode:       "",
		ContactPhoneNumber:           "",
		CityID:                       0,
		CityName:                     "",
		AddressStreet:                "",
		DirectMessaging:              "",
		Latitude:                     0,
		Longitude:                    0,
		Category:                     "",
		BusinessContactMethod:        "",
		IncludeDirectBlacklistStatus: false,
		HdProfilePicURLInfo:          goinsta.PicURLInfo{},
		HdProfilePicVersions:         nil,
		School:                       goinsta.School{},
		Byline:                       "",
		SocialContext:                "",
		SearchSocialContext:          "",
		MutualFollowersCount:         0,
		LatestReelMedia:              0,
		IsCallToActionEnabled:        false,
		FbPageCallToActionID:         "",
		Zip:                          "",
		Friendship:                   goinsta.Friendship{},
	}
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

	us := goinsta.User{
		ID:                         user.ID,
		Username:                   user.UserName,
		FullName:                   "",
		Biography:                  "",
		ProfilePicURL:              "",
		Email:                      "",
		PhoneNumber:                "",
		IsBusiness:                 false,
		Gender:                     0,
		ProfilePicID:               "",
		HasAnonymousProfilePicture: false,
		IsPrivate:                  false,
		IsUnpublished:              false,
		AllowedCommenterType:       "",
		IsVerified:                 false,
		MediaCount:                 0,
		FollowerCount:              0,
		FollowingCount:             0,
		FollowingTagCount:          0,
		MutualFollowersID:          nil,
		ProfileContext:             "",
		GeoMediaCount:              0,
		ExternalURL:                "",
		HasBiographyTranslation:    false,
		ExternalLynxURL:            "",
		BiographyWithEntities: struct {
			RawText  string        `json:"raw_text"`
			Entities []interface{} `json:"entities"`
		}{},
		UsertagsCount:                0,
		HasChaining:                  false,
		IsFavorite:                   false,
		IsFavoriteForStories:         false,
		IsFavoriteForHighlights:      false,
		CanBeReportedAsFraud:         false,
		ShowShoppableFeed:            false,
		ShoppablePostsCount:          0,
		ReelAutoArchive:              "",
		HasHighlightReels:            false,
		PublicEmail:                  "",
		PublicPhoneNumber:            "",
		PublicPhoneCountryCode:       "",
		ContactPhoneNumber:           "",
		CityID:                       0,
		CityName:                     "",
		AddressStreet:                "",
		DirectMessaging:              "",
		Latitude:                     0,
		Longitude:                    0,
		Category:                     "",
		BusinessContactMethod:        "",
		IncludeDirectBlacklistStatus: false,
		HdProfilePicURLInfo:          goinsta.PicURLInfo{},
		HdProfilePicVersions:         nil,
		School:                       goinsta.School{},
		Byline:                       "",
		SocialContext:                "",
		SearchSocialContext:          "",
		MutualFollowersCount:         0,
		LatestReelMedia:              0,
		IsCallToActionEnabled:        false,
		FbPageCallToActionID:         "",
		Zip:                          "",
		Friendship:                   goinsta.Friendship{},
	}
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

	diff := svc.whitelistNotMutual(notMutual)

	if len(diff) == 0 {
		return 0, nil
	}

	pBar := bar.New(len(diff), getBarType())

	go pBar.Run(svc.ctx)

	defer func() {
		pBar.Finish()
	}()

	return svc.unfollowUsers(pBar, notMutual)
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
	users := make([]*goinsta.User, 0)
	for _, un := range usernames {
		u, err := svc.instagram.client.Profiles.ByName(un)
		if err != nil {
			return nil, err
		} else {

		}
		users = append(users, u)
		time.Sleep(svc.instagram.sleep)
	}

	return users, nil
}

func (svc *Service) RemoveFollowersByUsername(usernames []string) (int, error) {
	pBar := bar.New(len(usernames), getBarType())

	go pBar.Run(svc.ctx)

	defer func() {
		pBar.Finish()
	}()

	return svc.removeFollowers(pBar, usernames)
}

func (svc *Service) removeFollowers(pBar bar.Bar, users []string) (int, error) {
	var count int

	ticker := time.NewTicker(svc.instagram.sleep)
	defer ticker.Stop()

	const errsLimit = 3

	var errsNum int

	// whitelist := svc.instagram.Whitelist()

LOOP:
	for _, un := range users {
		if errsNum >= errsLimit {
			return count, ErrCorrupted
		}

		select {
		case <-svc.ctx.Done():
			break LOOP
		case <-ticker.C:
			pBar.Progress() <- struct{}{}

			// if _, exist := whitelist[u.Username]; exist {
			// 	continue
			// }

			u, err := svc.instagram.client.Profiles.ByName(un)
			if err != nil {
				log.Errorf("user lookup failed [%s]: %v", un, err)
				errsNum++
				continue
			}

			if err := u.Block(); err != nil {
				log.Errorf("failed to block follower [%s]: %v", u.Username, err)
				errsNum++

				continue
			}
			if err := u.Unblock(); err != nil {
				log.Errorf("failed to unblock follower [%s]: %v", u.Username, err)
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

func (svc *Service) unfollowUsers(pBar bar.Bar, users []models.User) (int, error) {
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
		case <-svc.ctx.Done():
			break LOOP
		case <-ticker.C:
			if _, exist := whitelist[u.UserName]; exist {
				continue
			}

			pBar.Progress() <- struct{}{}

			if err := svc.UnFollow(u); err != nil {
				log.Errorf("failed to unfollow [%s]: %v", u.UserName, err)
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

	log.Debugf("processig[%s]: following[%d] \n", user.Username, flwsNum)

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
