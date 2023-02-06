// Package internal provides internal implementation logic for environment variables parsing.
package internal

import (
	"fmt"
	"time"
)

// NewEnvParser is a constructor for EnvParser.
func NewEnvParser(v any) EnvParser {
	var p EnvParser

	switch t := v.(type) {
	case string, []string:
		p = newStringParser(t)
	case int, []int, int64, []int64, uint64, []uint64, uint, []uint, uint32, []uint32:
		p = newIntParser(t)
	case bool:
		p = boolParser(t)
	case float64:
		p = float64Parser(t)
	case []float64:
		p = float64SliceParser(t)
	case time.Time:
		p = timeParser(t)
	case time.Duration:
		p = durationParser(t)
	default:
		p = nil
	}

	if p == nil {
		panic(fmt.Sprintf("unsupported type :%T", v))
	}

	return p
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

func newIntParser(v any) EnvParser {
	switch t := v.(type) {
	case int:
		return intParser(t)
	case []int:
		return intSliceParser(t)
	case int64:
		return int64Parser(t)
	case []int64:
		return int64SliceParser(t)
	case uint64:
		return uint64Parser(t)
	case []uint64:
		return uint64SliceParser(t)
	case uint:
		return uintParser(t)
	case []uint:
		return uintSliceParser(t)
	case []uint32:
		return uint32SliceParser(t)
	case uint32:
		return uint32Parser(t)
	default:
		return nil
	}
}

// EnvParser interface for parsing environment variables.
type EnvParser interface {
	ParseEnv(key string, defaltVal any, options Parameters) any
}

type stringParser string

func (s stringParser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := stringOrDefault(key, defaltVal.(string))

	return val
}

type stringSliceParser []string

func (s stringSliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := stringSliceOrDefault(key, defaltVal.([]string), sep)

	return val
}

type intParser int

func (i intParser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := intOrDefault(key, defaltVal.(int))

	return val
}

type intSliceParser []int

func (i intSliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := intSliceOrDefault(key, defaltVal.([]int), sep)

	return val
}

type float64SliceParser []float64

func (i float64SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := float64SliceOrDefault(key, defaltVal.([]float64), sep)

	return val
}

type int64Parser int64

func (i int64Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := int64OrDefault(key, defaltVal.(int64))

	return val
}

type int64SliceParser []int64

func (i int64SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := int64SliceOrDefault(key, defaltVal.([]int64), sep)

	return val
}

type float64Parser float64

func (f float64Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := float64OrDefault(key, defaltVal.(float64))

	return val
}

type boolParser bool

func (b boolParser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := boolOrDefault(key, defaltVal.(bool))

	return val
}

type timeParser time.Time

func (t timeParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	layout := options.Layout

	val := timeOrDefault(key, defaltVal.(time.Time), layout)

	return val
}

type durationParser time.Duration

func (d durationParser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := durationOrDefault(key, defaltVal.(time.Duration))

	return val
}

type uint64Parser uint64

func (d uint64Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := uint64OrDefault(key, defaltVal.(uint64))

	return val
}

type uint64SliceParser []uint64

func (i uint64SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := uint64SliceOrDefault(key, defaltVal.([]uint64), sep)

	return val
}

type uintParser uint

func (d uintParser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := uintOrDefault(key, defaltVal.(uint))

	return val
}

type uintSliceParser []uint

func (i uintSliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := uintSliceOrDefault(key, defaltVal.([]uint), sep)

	return val
}

type uint32SliceParser []uint32

func (i uint32SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := uint32SliceOrDefault(key, defaltVal.([]uint32), sep)

	return val
}

type uint32Parser uint

func (d uint32Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := uint32OrDefault(key, defaltVal.(uint32))

	return val
}
