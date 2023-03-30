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
	"github.com/obalunenko/getenv/internal"
	"github.com/obalunenko/getenv/option"
)

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
