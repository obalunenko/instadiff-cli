// Package client provides client for social networks.
package client

import (
	"context"

	"github.com/obalunenko/instadiff-cli/internal/client/instagram"
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
	Logout(ctx context.Context) error
}

// New creates Client. Also returns logout func.
func New(ctx context.Context, cfgPath string) (Client, error) {
	cl, err := makeInstagramClient(ctx, cfgPath)
	if err != nil {
		return nil, err
	}

	return cl, nil
}

func makeInstagramClient(ctx context.Context, cfgPath string) (Client, error) {
	return instagram.New(ctx, cfgPath)
}
