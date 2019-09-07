package config

import (
	"log"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config represents config for InstaDiff service.
type Config struct {
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

// IsDebug returns debug mode status.
func (c Config) IsDebug() bool {
	return c.debug
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
	defer viper.Reset()

	viper.SetEnvPrefix("instadiff")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	// Confirms which config file is used.
	log.Printf("Using config: %s\n\n", viper.ConfigFileUsed())

	cfg = Config{
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
		debug: viper.GetBool("debug"),
	}

	return cfg, nil
}
