// Package getenv provides functionality for loading environment variables.
package getenv

import (
	"time"

	"github.com/obalunenko/getenv/internal"
	"github.com/obalunenko/getenv/option"
)

// EnvOrDefault retrieves the value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func EnvOrDefault[T internal.EnvParsable](key string, defaultVal T, options ...option.Option) T {
	w := internal.NewEnvParser(defaultVal)

	params := newParseParams(options)

	val := w.ParseEnv(key, defaultVal, params)

	return val.(T)
}

func newParseParams(opts []option.Option) internal.Parameters {
	var p internal.Parameters

	for _, opt := range opts {
		opt.Apply(&p)
	}

	return p
}

// IntOrDefault retrieves the int value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func IntOrDefault(key string, defaultVal int) int {
	return EnvOrDefault(key, defaultVal)
}

// StringOrDefault retrieves the string value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func StringOrDefault(key, defaultVal string) string {
	return EnvOrDefault(key, defaultVal)
}

// BoolOrDefault retrieves the bool value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func BoolOrDefault(key string, defaultVal bool) bool {
	return EnvOrDefault(key, defaultVal)
}

// StringSliceOrDefault retrieves the string slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func StringSliceOrDefault(key string, defaultVal []string, sep string) []string {
	return EnvOrDefault(key, defaultVal, option.WithSeparator(sep))
}

// IntSliceOrDefault retrieves the int slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func IntSliceOrDefault(key string, defaultVal []int, sep string) []int {
	return EnvOrDefault(key, defaultVal, option.WithSeparator(sep))
}

// Float64SliceOrDefault retrieves the float64 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func Float64SliceOrDefault(key string, defaultVal []float64, sep string) []float64 {
	return EnvOrDefault(key, defaultVal, option.WithSeparator(sep))
}

// DurationOrDefault retrieves the time.Duration value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func DurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	return EnvOrDefault(key, defaultVal)
}

// TimeOrDefault retrieves the time.Time value of the environment variable named
// by the key represented by layout.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func TimeOrDefault(key string, defaultVal time.Time, layout string) time.Time {
	return EnvOrDefault(key, defaultVal, option.WithTimeLayout(layout))
}

// Int64OrDefault retrieves the int64 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func Int64OrDefault(key string, defaultVal int64) int64 {
	return EnvOrDefault(key, defaultVal)
}

// Float64OrDefault retrieves the float64 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func Float64OrDefault(key string, defaultVal float64) float64 {
	return EnvOrDefault(key, defaultVal)
}

// Int64SliceOrDefault retrieves the int6464 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
//
// Deprecated: use EnvOrDefault.
func Int64SliceOrDefault(key string, defaultVal []int64, sep string) []int64 {
	return EnvOrDefault(key, defaultVal, option.WithSeparator(sep))
}
