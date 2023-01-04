// Package getenv provides functionality for loading environment variables.
package getenv

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// IntOrDefault retrieves the int value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func IntOrDefault(key string, defaultVal int) int {
	env := StringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(env)
	if err != nil {
		return defaultVal
	}

	return val
}

// StringOrDefault retrieves the string value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func StringOrDefault(key, defaultVal string) string {
	env, ok := os.LookupEnv(key)
	if !ok || env == "" {
		return defaultVal
	}

	return env
}

// BoolOrDefault retrieves the bool value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func BoolOrDefault(key string, defaultVal bool) bool {
	env := StringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := strconv.ParseBool(env)
	if err != nil {
		return defaultVal
	}

	return val
}

// StringSliceOrDefault retrieves the string slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func StringSliceOrDefault(key string, defaultVal []string, sep string) []string {
	env := StringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val := strings.Split(env, sep)

	return val
}

// DurationOrDefault retrieves the time.Duration value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func DurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	env := StringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := time.ParseDuration(env)
	if err != nil {
		return defaultVal
	}

	return val
}

// TimeOrDefault retrieves the time.Time value of the environment variable named
// by the key represented by layout.
// If variable not set or value is empty - defaultVal will be returned.
func TimeOrDefault(key string, defaultVal time.Time, layout string) time.Time {
	env := StringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := time.Parse(layout, env)
	if err != nil {
		return defaultVal
	}

	return val
}

// Int64OrDefault retrieves the int64 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func Int64OrDefault(key string, defaultVal int64) int64 {
	env := StringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 64
	)

	val, err := strconv.ParseInt(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return val
}

// Float64OrDefault retrieves the float64 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func Float64OrDefault(key string, defaultVal float64) float64 {
	env := StringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		bitsize = 64
	)

	val, err := strconv.ParseFloat(env, bitsize)
	if err != nil {
		return defaultVal
	}

	return val
}
