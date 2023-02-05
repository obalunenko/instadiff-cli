// Package getenv provides functionality for loading environment variables.
package getenv

import (
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
