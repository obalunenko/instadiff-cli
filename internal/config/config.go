// Package config provide configuration.
package config

import (
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config represents config for InstaDiff service.
type Config struct {
	db        db
	user      user
	whitelist []string
	limits    limits
	debug     bool
}

type instagram struct {
	username string
	password string
}

type user struct {
	instagram instagram
}

type limits struct {
	unfollow int
}

type db struct {
	local               bool
	mongoURL            string
	mongoDBName         string
	mongoCollectionName string
}

// Username returns username.
func (c Config) Username() string {
	return c.user.instagram.username
}

// Password returns password.
func (c Config) Password() string {
	return c.user.instagram.password
}

// UnFollowLimits returns unFollow action daily limits.
func (c Config) UnFollowLimits() int {
	return c.limits.unfollow
}

// FollowLimits returns follow action daily limits.
func (c Config) FollowLimits() int {
	return c.limits.unfollow
}

// Whitelist returns map of whitelisted users.
func (c Config) Whitelist() map[string]struct{} {
	if len(c.whitelist) == 0 {
		return nil
	}

	wl := make(map[string]struct{}, len(c.whitelist))
	for _, l := range c.whitelist {
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
	if debug {
		log.Println("debug mode set")
	}

	c.debug = debug
}

// IsLocalDBEnabled returns local DB enabled status
func (c Config) IsLocalDBEnabled() bool {
	return c.db.local
}

// MongoConfigURL returns configured MongoDB URL
func (c Config) MongoConfigURL() string {
	return c.db.mongoURL
}

// MongoDBName returns configured MongoDB name
func (c Config) MongoDBName() string {
	return c.db.mongoDBName
}

// MongoDBCollection returns configured MongoDB collection
func (c Config) MongoDBCollection() string {
	return c.db.mongoCollectionName
}

// Load loads config from passed filepath
func Load(path string) (Config, error) {
	var cfg Config

	if path == "" {
		return Config{}, errors.New("config path is empty")
	}

	viper.SetConfigFile(path)

	// Reads the config file
	if err := viper.ReadInConfig(); err != nil {
		return cfg, errors.Wrapf(err, "failed to read config form path: %s", path)
	}

	// Reset viper to free memory.
	defer func() {
		viper.Reset()
	}()

	viper.SetEnvPrefix("instadiff")

	replacer := strings.NewReplacer(".", "_")

	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	// Confirms which config file is used.
	log.Infof("Using config: %s\n\n", viper.ConfigFileUsed())

	cfg = Config{
		db: db{
			local:               viper.GetBool("db.local"),
			mongoURL:            viper.GetString("db.mongoURL"),
			mongoDBName:         viper.GetString("db.mongoDBName"),
			mongoCollectionName: viper.GetString("db.mongoCollectionName"),
		},
		user: user{
			instagram: instagram{
				username: viper.GetString("user.instagram.username"),
				password: viper.GetString("user.instagram.password"),
			},
		},
		whitelist: viper.GetStringSlice("whitelist"),
		limits: limits{
			unfollow: viper.GetInt("limits.unfollow"),
		},
	}

	return cfg, nil
}
