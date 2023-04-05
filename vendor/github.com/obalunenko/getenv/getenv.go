// Package getenv provides a simple way to get environment variables.
// It's type-safe and supports built-in types and slices of them.
//
// Types supported:
// - string
// - []string
// - int
// - []int
// - int8
// - []int8
// - int16
// - []int16
// - int32
// - []int32
// - int64
// - []int64
// - uint8
// - []uint8
// - uint16
// - []uint16
// - uint64
// - []uint64
// - uint
// - []uint
// - uintptr
// - []uintptr
// - uint32
// - []uint32
// - float32
// - []float32
// - float64
// - []float64
// - time.Time
// - []time.Time
// - time.Duration
// - []time.Duration
// - bool
// - []bool
// - url.URL
// - []url.URL
// - net.IP
// - []net.IP
// - complex64
// - []complex64
// - complex128
// - []complex128
package getenv

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/obalunenko/getenv/internal"
	"github.com/obalunenko/getenv/option"
)

var (
	// ErrNotSet is an error that is returned when the environment variable is not set.
	ErrNotSet = errors.New("not set")
	// ErrInvalidValue is an error that is returned when the environment variable is not valid.
	ErrInvalidValue = errors.New("invalid value")
)

// Env retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value will be parsed and returned.
// Otherwise, an error will be returned.
func Env[T internal.EnvParsable](key string, options ...option.Option) (T, error) {
	// Create a default value of the same type as the value that we want to get.
	var defVal T

	val := EnvOrDefault(key, defVal, options...)

	// If the value is equal to the default value, it means that the value was not parsed.
	// This means that the environment variable was not set, or it was set to an invalid value.
	if reflect.DeepEqual(val, defVal) {
		v, ok := os.LookupEnv(key)
		if !ok {
			return val, fmt.Errorf("could not get variable[%s]: %w", key, ErrNotSet)
		}

		return val, fmt.Errorf("could not parse variable[%s] value[%v] to type[%T]: %w", key, v, defVal, ErrInvalidValue)
	}

	return val, nil
}

// EnvOrDefault retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value will be parsed and returned.
// Otherwise, the default value will be returned.
// The value returned will be of the same type as the default value.
func EnvOrDefault[T internal.EnvParsable](key string, defaultVal T, options ...option.Option) T {
	w := internal.NewEnvParser(defaultVal)

	params := newParseParams(options)

	val := w.ParseEnv(key, defaultVal, params)

	return val.(T)
}

// newParseParams creates new parameters from options.
func newParseParams(opts []option.Option) internal.Parameters {
	var p internal.Parameters

	for _, opt := range opts {
		opt.Apply(&p)
	}

	return p
}
