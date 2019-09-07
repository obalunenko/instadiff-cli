package service

import (
	"log"

	"github.com/pkg/errors"
	"github.com/tducasse/goinsta"

	"github.com/oleg-balunenko/insta-follow-diff/internal/config"
	"github.com/oleg-balunenko/insta-follow-diff/internal/models"
)

// ErrLimitExceed returned when limit for action exceeded.
var ErrLimitExceed = errors.New("limit exceeded")

// Service represents service for operating instagram account.
type Service struct {
	instagramClient *goinsta.Instagram
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
	err := cl.Login()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to login")
	}
	log.Printf("logged in as %s \n", cl.Informations.Username)

	lmts := limits{
		follow:   cfg.FollowLimits(),
		unFollow: cfg.UnFollowLimits(),
	}

	inst := &Service{
		instagramClient: cl,
		limits:          lmts,
		whitelist:       cfg.Whitelist(),
		debug:           cfg.IsDebug(),
	}

	return inst, inst.stop, nil
}

// GetFollowers returns list of followers for logged in user.
func (svc *Service) GetFollowers() ([]models.UserInfo, error) {
	resp, err := svc.instagramClient.SelfTotalUserFollowers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get followers")
	}
	followers := make([]models.UserInfo, 0, len(resp.Users))
	for _, us := range resp.Users {
		followers = append(followers, models.MakeUserInfo(us.ID, us.Username, us.FullName))
	}

	return followers, nil
}

// GetFollowings returns list of followings for logged in user.
func (svc *Service) GetFollowings() ([]models.UserInfo, error) {
	resp, err := svc.instagramClient.SelfTotalUserFollowing()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get followings")
	}
	followings := make([]models.UserInfo, 0, len(resp.Users))
	for _, us := range resp.Users {
		followings = append(followings, models.MakeUserInfo(us.ID, us.Username, us.FullName))
	}

	return followings, nil
}

// GetNotMutualFollowers returns list of users that not following back.
func (svc *Service) GetNotMutualFollowers() ([]models.UserInfo, error) {
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

	var notmutual []models.UserInfo
	for _, fu := range followings {
		if _, mutual := followersMap[fu.ID]; !mutual {
			notmutual = append(notmutual, fu)
		}
	}

	return notmutual, nil
}

// UnFollow removes user from followings.
func (svc *Service) UnFollow(user models.UserInfo) error {
	if svc.debug {
		log.Printf("unFollow user %+v\n", user)
		return nil
	}
	response, err := svc.instagramClient.UnFollow(user.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to unFollow user %+v", user)
	}
	log.Printf("unFollow user %v: status %v \n", user, response)
	return nil
}

// Follow adds user to followings.
func (svc *Service) Follow(user models.UserInfo) error {
	if svc.debug {
		log.Printf("follow user %+v\n", user)
		return nil
	}
	response, err := svc.instagramClient.Follow(user.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to follow user %+v", user)
	}
	log.Printf("follow user %v: status %v \n", user, response)
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
		err = svc.UnFollow(nu)
		if err != nil {
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

	var count int
	for _, nu := range notMutual {
		if _, ok := svc.whitelist[nu.UserName]; !ok {
			err = svc.UnFollow(nu)
			if err != nil {
				log.Printf("failed to unFollow user %v:%v", nu, err)
				continue
			}
			count++
		} else {
			log.Printf("skip whitelisted user %v", nu)
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
	_ = svc.instagramClient.Logout()
}
