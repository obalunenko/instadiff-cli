// Package config provide configuration.
package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"

	log "github.com/obalunenko/logger"
)

// Config represents config for InstaDiff service.
type Config struct {
	storage   storage
	instagram instagram
	debug     bool
}

type instagram struct {
	save      bool
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
	url        string
	db         string
	collection string
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

// Debug returns current dubug status.
func (c *Config) Debug() bool {
	return c.debug
}

// SetDebug updates debug status.
func (c *Config) SetDebug(debug bool) {
	c.debug = debug
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

// MongoDBCollection returns configured MongoDB collection.
func (c Config) MongoDBCollection() string {
	return c.storage.mongo.collection
}

// StoreSession returns flag if session should be stored locally.
func (c Config) StoreSession() bool {
	return c.instagram.save
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
				url:        viper.GetString("storage.mongo.url"),
				db:         viper.GetString("storage.mongo.db"),
				collection: viper.GetString("storage.mongo.collection"),
			},
		},
		instagram: instagram{
			save:      viper.GetBool("instagram.save"),
			whitelist: viper.GetStringSlice("instagram.whitelist"),
			limits: limits{
				unfollow: viper.GetInt("instagram.limits.unfollow"),
			},
			sleep: viper.GetInt64("instagram.sleep"),
		},
		debug: false,
	}

	return cfg, nil
}
