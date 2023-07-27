// Package internal provides internal implementation logic for environment variables parsing.
package internal

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

// NewEnvParser is a constructor for EnvParser.
func NewEnvParser(v any) EnvParser {
	var p EnvParser

	switch t := v.(type) {
	case string, []string:
		p = newStringParser(t)
	case int, []int, int8, []int8, int16, []int16, int32, []int32, int64, []int64:
		p = newIntParser(t)
	case uint, []uint, uint8, []uint8, uint16, []uint16, uint32, []uint32, uint64, []uint64, uintptr, []uintptr:
		p = newUintParser(t)
	case bool, []bool:
		p = newBoolParser(t)
	case float32, []float32, float64, []float64:
		p = newFloatParser(t)
	case time.Time, []time.Time, time.Duration, []time.Duration:
		p = newTimeParser(t)
	case url.URL, []url.URL:
		p = newURLParser(t)
	case net.IP, []net.IP:
		p = newIPParser(t)
	case complex64, []complex64, complex128, []complex128:
		p = newComplexParser(t)
	default:
		p = nil
	}

	if p == nil {
		panic(fmt.Sprintf("unsupported type :%T", v))
	}

	return p
}

// newComplexParser is a constructor for complex parsers.
func newComplexParser(v any) EnvParser {
	switch v.(type) {
	case complex64:
		return complexParser[complex64]{}
	case []complex64:
		return complexSliceParser[[]complex64, complex64]{}
	case complex128:
		return complexParser[complex128]{}
	case []complex128:
		return complexSliceParser[[]complex128, complex128]{}
	default:
		return nil
	}
}

// newURLParser is a constructor for url.URL parsers.
func newURLParser(v any) EnvParser {
	switch t := v.(type) {
	case url.URL:
		return urlParser(t)
	case []url.URL:
		return urlSliceParser(t)
	default:
		return nil
	}
}

// newIPParser is a constructor for net.IP parsers.
func newIPParser(v any) EnvParser {
	switch t := v.(type) {
	case net.IP:
		return ipParser(t)
	case []net.IP:
		return ipSliceParser(t)
	default:
		return nil
	}
}

func newStringParser(v any) EnvParser {
	switch t := v.(type) {
	case string:
		return stringParser(t)
	case []string:
		return stringSliceParser(t)
	default:
		return nil
	}
}

// newIntParser is a constructor for integer parsers.
func newIntParser(v any) EnvParser {
	switch v.(type) {
	case int:
		return numberParser[int]{}
	case []int:
		return numberSliceParser[[]int, int]{}
	case int8:
		return numberParser[int8]{}
	case []int8:
		return numberSliceParser[[]int8, int8]{}
	case int16:
		return numberParser[int16]{}
	case []int16:
		return numberSliceParser[[]int16, int16]{}
	case int32:
		return numberParser[int32]{}
	case []int32:
		return numberSliceParser[[]int32, int32]{}
	case int64:
		return numberParser[int64]{}
	case []int64:
		return numberSliceParser[[]int64, int64]{}
	default:
		return nil
	}
}

// newUintParser is a constructor for unsigned integer parsers.
func newUintParser(v any) EnvParser {
	switch v.(type) {
	case uint8:
		return numberParser[uint8]{}
	case []uint8:
		return numberSliceParser[[]uint8, uint8]{}
	case uint:
		return numberParser[uint]{}
	case []uint:
		return numberSliceParser[[]uint, uint]{}
	case uint16:
		return numberParser[uint16]{}
	case []uint16:
		return numberSliceParser[[]uint16, uint16]{}
	case uint32:
		return numberParser[uint32]{}
	case []uint32:
		return numberSliceParser[[]uint32, uint32]{}
	case uint64:
		return numberParser[uint64]{}
	case []uint64:
		return numberSliceParser[[]uint64, uint64]{}
	case uintptr:
		return numberParser[uintptr]{}
	case []uintptr:
		return numberSliceParser[[]uintptr, uintptr]{}
	default:
		return nil
	}
}

// newFloatParser is a constructor for float parsers.
func newFloatParser(v any) EnvParser {
	switch v.(type) {
	case float32:
		return numberParser[float32]{}
	case []float32:
		return numberSliceParser[[]float32, float32]{}
	case float64:
		return numberParser[float64]{}
	case []float64:
		return numberSliceParser[[]float64, float64]{}
	default:
		return nil
	}
}

// newTimeParser is a constructor for time parsers.
func newTimeParser(v any) EnvParser {
	switch t := v.(type) {
	case time.Time:
		return timeParser(t)
	case []time.Time:
		return timeSliceParser(t)
	case time.Duration:
		return durationParser(t)
	case []time.Duration:
		return durationSliceParser(t)
	default:
		return nil
	}
}

// newBoolParser is a constructor for boolParser.
func newBoolParser(v any) EnvParser {
	switch t := v.(type) {
	case bool:
		return boolParser(t)
	case []bool:
		return boolSliceParser(t)
	default:
		return nil
	}
}

// EnvParser interface for parsing environment variables.
type EnvParser interface {
	// ParseEnv parses environment variable by key and returns value.
	ParseEnv(key string, options Parameters) (any, error)
}

// stringParser is a parser for string type.
type stringParser string

func (s stringParser) ParseEnv(key string, _ Parameters) (any, error) {
	return getString(key)
}

type stringSliceParser []string

func (s stringSliceParser) ParseEnv(key string, options Parameters) (any, error) {
	sep := options.Separator

	return getStringSlice(key, sep)
}

type numberParser[T Number] struct{}

func (n numberParser[T]) ParseEnv(key string, _ Parameters) (any, error) {
	return getNumberGen[T](key)
}

type numberSliceParser[S []T, T Number] struct{}

func (i numberSliceParser[S, T]) ParseEnv(key string, options Parameters) (any, error) {
	sep := options.Separator

	return getNumberSliceGen[S, T](key, sep)
}

type boolParser bool

func (b boolParser) ParseEnv(key string, _ Parameters) (any, error) {
	return getBool(key)
}

type timeParser time.Time

func (t timeParser) ParseEnv(key string, options Parameters) (any, error) {
	layout := options.Layout

	return getTime(key, layout)
}

type timeSliceParser []time.Time

func (t timeSliceParser) ParseEnv(key string, options Parameters) (any, error) {
	layout := options.Layout
	sep := options.Separator

	return getTimeSlice(key, layout, sep)
}

type durationSliceParser []time.Duration

func (t durationSliceParser) ParseEnv(key string, options Parameters) (any, error) {
	sep := options.Separator

	return getDurationSlice(key, sep)
}

type durationParser time.Duration

func (d durationParser) ParseEnv(key string, _ Parameters) (any, error) {
	return getDuration(key)
}

// stringSliceParser is a parser for []string
type urlParser url.URL

func (t urlParser) ParseEnv(key string, _ Parameters) (any, error) {
	return getURL(key)
}

// urlSliceParser is a parser for []url.URL
type urlSliceParser []url.URL

func (t urlSliceParser) ParseEnv(key string, opts Parameters) (any, error) {
	separator := opts.Separator

	return getURLSlice(key, separator)
}

// ipParser is a parser for net.IP
type ipParser net.IP

func (t ipParser) ParseEnv(key string, _ Parameters) (any, error) {
	return getIP(key)
}

// ipSliceParser is a parser for []net.IP
type ipSliceParser []net.IP

func (t ipSliceParser) ParseEnv(key string, opts Parameters) (any, error) {
	separator := opts.Separator

	return getIPSlice(key, separator)
}

// boolSliceParser is a parser for []bool
type boolSliceParser []bool

func (b boolSliceParser) ParseEnv(key string, options Parameters) (any, error) {
	sep := options.Separator

	return getBoolSlice(key, sep)
}

type complexParser[T Complex] struct{}

func (n complexParser[T]) ParseEnv(key string, _ Parameters) (any, error) {
	return getComplexGen[T](key)
}

type complexSliceParser[S []T, T Complex] struct{}

func (i complexSliceParser[S, T]) ParseEnv(key string, options Parameters) (any, error) {
	sep := options.Separator

	return getComplexSliceGen[S, T](key, sep)
}
