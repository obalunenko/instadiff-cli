// Package instagram provides interactions with Instagram social account.
package instagram

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	clientErrors "github.com/obalunenko/instadiff-cli/internal/client/errors"

	"github.com/Davincible/goinsta/v3"
	log "github.com/obalunenko/logger"
	"github.com/tcnksm/go-input"

	"github.com/obalunenko/instadiff-cli/internal/actions"

	"github.com/obalunenko/instadiff-cli/pkg/spinner"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

// Client represents instagram client.
type Client struct {
	client   *goinsta.Instagram
	sessFile string
	sleep    time.Duration
}

// Params holds Client constructor parameters.
type Params struct {
	Sleep       time.Duration
	SessionPath string
	Username    string
}

// New Client constructor.
func New(ctx context.Context, p Params) (*Client, error) {
	var (
		err   error
		uname string
	)

	uname = p.Username

	if uname == "" {
		uname, err = usernameInput()
		if err != nil {
			return nil, fmt.Errorf("username: %w", err)
		}
	}

	sessFile := filepath.Join(p.SessionPath, fmt.Sprintf("%s.sess", uname))

	cl, err := importFromFile(ctx, sessFile)
	if err == nil {
		return cl, nil
	}

	pwd, err := passwordInput()
	if err != nil {
		return nil, fmt.Errorf("password: %w", err)
	}

	cl, err = login(ctx, uname, pwd, sessFile)
	if err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return cl, nil
}

func importFromFile(ctx context.Context, sessFile string) (*Client, error) {
	stop := spinner.Set("Trying to import previous session..", "", "yellow")

	i, err := goinsta.Import(sessFile)

	stop()

	if err != nil {
		return nil, err
	}

	log.WithField(ctx, "session_file", sessFile).Info("Session imported")

	return syncInstagram(ctx, i, sessFile)
}

func syncInstagram(ctx context.Context, cli *goinsta.Instagram, sessFile string) (*Client, error) {
	stop := spinner.Set("Refreshing account info", "", "yellow")

	if err := cli.OpenApp(); err != nil {
		log.WithError(ctx, err).Error("Failed to refresh app info")
	}

	stop()

	stop = spinner.Set("Exporting session", "", "yellow")

	if err := cli.Export(sessFile); err != nil {
		log.WithError(ctx, err).Error("Failed to save session")
	}

	stop()

	return &Client{
		client:   cli,
		sessFile: sessFile,
	}, nil
}

func login(ctx context.Context, uname, pwd, sessFile string) (*Client, error) {
	insta := goinsta.New(uname, pwd)

	stop := spinner.Set("Sending log in request..", "", "yellow")

	err := insta.Login()

	stop()

	insta, err = maybeChallengeRequired(insta, err)
	if err != nil {
		return nil, err
	}

	return syncInstagram(ctx, insta, sessFile)
}

func maybeChallengeRequired(insta *goinsta.Instagram, err error) (*goinsta.Instagram, error) {
	switch {
	case errors.Is(err, nil):
		return insta, nil
	case errors.Is(err, goinsta.ErrChallengeRequired):
		var chErr *goinsta.ChallengeError

		if !errors.As(err, &chErr) {
			return nil, fmt.Errorf("failed to get challenge details: %w", err)
		}

		insta, err = challenge(insta, chErr.Challenge.APIPath)
		if err != nil {
			return nil, fmt.Errorf("challenge: %w", err)
		}
	case errors.Is(err, goinsta.Err2FARequired) || errors.Is(err, goinsta.Err2FANoCode):
		var code string

		code, err = twoFactorCode()
		if err != nil {
			return nil, fmt.Errorf("2fa ocde: %w", err)
		}

		stop := spinner.Set("Sending 2fa code..", "", "yellow")
		defer stop()

		if err = insta.TwoFactorInfo.Login2FA(code); err != nil {
			return nil, fmt.Errorf("login 2fa: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected: %w", err)
	}

	return insta, nil
}

func usernameInput() (string, error) {
	ask := "What is your username?"
	key := "username"

	return getPrompt(ask, key)
}

func passwordInput() (string, error) {
	ask := "What is your password?"
	key := "password"

	return getPrompt(ask, key)
}

func twoFactorCode() (string, error) {
	ask := "What is your two factor code?"
	key := "2fa code"

	return getPrompt(ask, key)
}

func getPrompt(ask, key string) (string, error) {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	in, err := ui.Ask(ask,
		&input.Options{
			Default:     "",
			Loop:        true,
			Required:    true,
			HideDefault: false,
			HideOrder:   false,
			Hide:        false,
			Mask:        false,
			MaskDefault: false,
			MaskVal:     "",
			ValidateFunc: func(s string) error {
				s = strings.TrimSpace(s)
				if s == "" {
					return clientErrors.ErrEmptyInput
				}

				return nil
			},
		})
	if err != nil {
		return "", fmt.Errorf("%s input: %w", key, err)
	}

	return in, nil
}

func challenge(cl *goinsta.Instagram, chURL string) (*goinsta.Instagram, error) {
	if err := cl.Challenge.ProcessOld(chURL); err != nil {
		return nil, fmt.Errorf("process challenge: %w", err)
	}

	ask := "What is SMS code for instagram?"
	key := "SMS code"

	code, err := getPrompt(ask, key)
	if err != nil {
		return nil, fmt.Errorf("get prompt: %w", err)
	}

	stop := spinner.Set("Sending security code..", "", "yellow")
	defer stop()

	if err = cl.Challenge.SendSecurityCode(code); err != nil {
		return nil, fmt.Errorf("send security code: %w", err)
	}

	return cl, nil
}

// IsUseless reports where user is useless for statistics.
func (c *Client) IsUseless(ctx context.Context, user models.User, threshold int) (bool, error) {
	u, err := c.client.Profiles.ByName(user.UserName)
	if err != nil {
		return false, err
	}

	followings, err := c.UserFollowings(ctx, user)
	if err != nil {
		return false, err
	}

	u.SetInstagram(c.client)

	if err := u.Info(); err != nil {
		return false, fmt.Errorf("update info: %w", err)
	}

	log.WithFields(ctx, log.Fields{
		"username":    user.UserName,
		"followings":  len(followings),
		"posts_count": u.MediaCount,
	}).Debug("Processing user for useless")

	return len(followings) >= threshold || u.CanBeReportedAsFraud || u.IsBusiness || u.MediaCount == 0, nil
}

// UserFollowers returns user followers.
func (c *Client) UserFollowers(ctx context.Context, user models.User) ([]models.User, error) {
	u, err := c.client.Profiles.ByName(user.UserName)
	if err != nil {
		return nil, err
	}

	u.SetInstagram(c.client)

	return c.makeUsersList(ctx, u.Followers())
}

// UserFollowings returns user followings.
func (c *Client) UserFollowings(ctx context.Context, user models.User) ([]models.User, error) {
	u, err := c.client.Profiles.ByName(user.UserName)
	if err != nil {
		return nil, err
	}

	u.SetInstagram(c.client)

	return c.makeUsersList(ctx, u.Following())
}

// GetUserByName finds user by username.
func (c *Client) GetUserByName(_ context.Context, username string) (models.User, error) {
	u, err := c.client.Profiles.ByName(username)
	if err != nil {
		if isErrUserNotFound(err) {
			return models.User{}, clientErrors.ErrUserNotFound
		}

		return models.User{}, err
	}

	return models.MakeUser(u.ID, u.Username, u.FullName), nil
}

func isErrUserNotFound(err error) bool {
	return strings.Contains(err.Error(), "user_not_found")
}

// Block user.
func (c *Client) Block(ctx context.Context, user models.User) error {
	return c.actUser(ctx, user, actions.UserActionBlock)
}

// Unblock user.
func (c *Client) Unblock(ctx context.Context, user models.User) error {
	return c.actUser(ctx, user, actions.UserActionUnblock)
}

// Follow user.
func (c *Client) Follow(ctx context.Context, user models.User) error {
	return c.actUser(ctx, user, actions.UserActionFollow)
}

// Unfollow user.
func (c *Client) Unfollow(ctx context.Context, user models.User) error {
	return c.actUser(ctx, user, actions.UserActionUnfollow)
}

// Followers returns list of followers.
func (c *Client) Followers(ctx context.Context) ([]models.User, error) {
	return c.makeUsersList(ctx, c.client.Account.Followers())
}

// Followings returns list of followings.
func (c *Client) Followings(ctx context.Context) ([]models.User, error) {
	return c.makeUsersList(ctx, c.client.Account.Following())
}

// Username returns current account username.
func (c *Client) Username(_ context.Context) string {
	return c.client.Account.Username
}

func (c *Client) actUser(ctx context.Context, user models.User, act actions.UserAction) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	us := goinsta.User{
		ID:       user.ID,
		Username: user.UserName,
	}

	us.SetInstagram(c.client)

	var f func() error

	switch act {
	case actions.UserActionFollow:
		f = us.Follow
	case actions.UserActionUnfollow:
		f = us.Unfollow
	case actions.UserActionBlock:
		f = func() error {
			return us.Block(false)
		}
	case actions.UserActionUnblock:
		f = us.Unblock
	default:
		return fmt.Errorf("unsupported user action type: %s", act.String())
	}

	if err := f(); err != nil {
		return fmt.Errorf("action[%s]: %w", act.String(), err)
	}

	return nil
}

// Logout clean session and send logout request.
func (c *Client) Logout(ctx context.Context) error {
	cl := c.client

	if err := os.Remove(c.sessFile); err != nil {
		return fmt.Errorf("remove ssession file: %w", err)
	}

	log.WithField(ctx, "file_path", c.sessFile).Info("Session file removed")

	if err := cl.Logout(); err != nil {
		// weird error - just ignore it.
		if strings.Contains(err.Error(), "405 Method Not Allowed") {
			return nil
		}

		return fmt.Errorf("logout: %w", err)
	}

	log.WithField(ctx, "username", cl.Account.Username).Info("Logged out")

	return nil
}

func (c *Client) makeUsersList(ctx context.Context, users *goinsta.Users) ([]models.User, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	seen := make(map[int64]bool, len(users.Users))

	usersList := make([]models.User, 0, len(users.Users))

	var itr int

	for users.Next() {
		if itr != 0 {
			time.Sleep(c.sleep)
		}

		for i := range users.Users {
			u := users.Users[i]

			if seen[u.ID] {
				continue
			}

			usersList = append(usersList,
				models.MakeUser(u.ID, u.Username, u.FullName))

			seen[u.ID] = true
		}

		itr++
	}

	if err := users.Error(); err != nil {
		if !errors.Is(err, goinsta.ErrNoMore) {
			return nil, fmt.Errorf("users iterate: %w", err)
		}
	}

	return usersList, nil
}
