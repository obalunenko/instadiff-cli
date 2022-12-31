// Package client provides client for social networks.
package client

import (
	"context"
	"io"
	"time"

	"github.com/obalunenko/instadiff-cli/internal/client/instagram"
	"github.com/obalunenko/instadiff-cli/internal/media"
	"github.com/obalunenko/instadiff-cli/internal/models"
)

// Client is a common social client interface.
type Client interface {
	Username(ctx context.Context) string
	GetUserByName(ctx context.Context, username string) (models.User, error)
	Followers(ctx context.Context) ([]models.User, error)
	UserFollowers(ctx context.Context, user models.User) ([]models.User, error)
	Followings(ctx context.Context) ([]models.User, error)
	UserFollowings(ctx context.Context, user models.User) ([]models.User, error)
	Follow(ctx context.Context, user models.User) error
	Unfollow(ctx context.Context, user models.User) error
	Block(ctx context.Context, user models.User) error
	Unblock(ctx context.Context, user models.User) error
	IsUseless(ctx context.Context, user models.User, threshold int) (bool, error)
	UploadMedia(ctx context.Context, file io.Reader, mt media.Type) error
	Logout(ctx context.Context) error
}

// Params holds Client constructor parameters.
type Params struct {
	SessionPath string
	Sleep       time.Duration
	Username    string
}

// New creates Client. Also returns logout func.
func New(ctx context.Context, p Params) (Client, error) {
	cl, err := makeInstagramClient(ctx, instagram.Params{
		Sleep:       p.Sleep,
		SessionPath: p.SessionPath,
		Username:    p.Username,
	})
	if err != nil {
		return nil, err
	}

	return cl, nil
}

func makeInstagramClient(ctx context.Context, params instagram.Params) (Client, error) {
	return instagram.New(ctx, params)
}
