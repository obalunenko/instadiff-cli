// Package config provide configuration.
package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/obalunenko/logger"
	"github.com/spf13/viper"
)

// Config represents config for InstaDiff service.
type Config struct {
	storage   storage
	instagram instagram
}

type instagram struct {
	whitelist []string
	limits    limits
	sleep     int64
}

type limits struct {
	unfollow int
}

type storage struct {
	local bool
	mongo mongo
}

type mongo struct {
	url string
	db  string
}

// UnFollowLimits returns unFollow action daily limits.
func (c Config) UnFollowLimits() int {
	return c.instagram.limits.unfollow
}

// Sleep returns wait duration from for instagram operations to avoid blocks.
func (c Config) Sleep() time.Duration {
	return time.Second * time.Duration(c.instagram.sleep)
}

// Whitelist returns map of whitelisted users.
func (c Config) Whitelist() map[string]struct{} {
	if len(c.instagram.whitelist) == 0 {
		return nil
	}

	wl := make(map[string]struct{}, len(c.instagram.whitelist))
	for _, l := range c.instagram.whitelist {
		wl[l] = struct{}{}
	}

	return wl
}

// IsLocalDBEnabled returns local DB enabled status.
func (c Config) IsLocalDBEnabled() bool {
	return c.storage.local
}

// MongoConfigURL returns configured MongoDB URL.
func (c Config) MongoConfigURL() string {
	return c.storage.mongo.url
}

// MongoDBName returns configured MongoDB name.
func (c Config) MongoDBName() string {
	return c.storage.mongo.db
}

// Load loads config from passed filepath.
func Load(ctx context.Context, path string) (Config, error) {
	var cfg Config

	if path == "" {
		return Config{}, ErrEmptyPath
	}

	viper.SetConfigFile(path)

	// Reads the config file.
	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	// Reset viper to free memory.
	defer viper.Reset()

	viper.SetEnvPrefix("instadiff")

	replacer := strings.NewReplacer(".", "_")

	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	// Confirms which config file is used.
	log.WithField(ctx, "config_path", viper.ConfigFileUsed()).Info("Using config file")

	cfg = Config{
		storage: storage{
			local: viper.GetBool("storage.local"),
			mongo: mongo{
				url: viper.GetString("storage.mongo.url"),
				db:  viper.GetString("storage.mongo.db"),
			},
		},
		instagram: instagram{
			whitelist: viper.GetStringSlice("instagram.whitelist"),
			limits: limits{
				unfollow: viper.GetInt("instagram.limits.unfollow"),
			},
			sleep: viper.GetInt64("instagram.sleep"),
		},
	}

	return cfg, nil
}
