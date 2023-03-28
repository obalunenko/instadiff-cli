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
	case uint, []uint, uint8, []uint8, uint16, []uint16, uint32, []uint32, uint64, []uint64:
		p = newUintParser(t)
	case bool:
		p = boolParser(t)
	case float32, []float32, float64, []float64:
		p = newFloatParser(t)
	case time.Time, []time.Time, time.Duration, []time.Duration:
		p = newTimeParser(t)
	case url.URL:
		p = urlParser(t)
	case []url.URL:
		p = urlSliceParser(t)
	case net.IP:
		p = ipParser(t)
	case []net.IP:
		p = ipSliceParser(t)
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
	case int8:
		return int8Parser(t)
	case []int8:
		return int8SliceParser(t)
	case int16:
		return int16Parser(t)
	case []int16:
		return int16SliceParser(t)
	case int32:
		return int32Parser(t)
	case []int32:
		return int32SliceParser(t)
	case int64:
		return int64Parser(t)
	case []int64:
		return int64SliceParser(t)
	default:
		return nil
	}
}

func newUintParser(v any) EnvParser {
	switch t := v.(type) {
	case uint8:
		return uint8Parser(t)
	case []uint8:
		return uint8SliceParser(t)
	case uint:
		return uintParser(t)
	case []uint:
		return uintSliceParser(t)
	case uint16:
		return uint16Parser(t)
	case []uint16:
		return uint16SliceParser(t)
	case uint32:
		return uint32Parser(t)
	case []uint32:
		return uint32SliceParser(t)
	case uint64:
		return uint64Parser(t)
	case []uint64:
		return uint64SliceParser(t)
	default:
		return nil
	}
}

func newFloatParser(v any) EnvParser {
	switch t := v.(type) {
	case float32:
		return float32Parser(t)
	case []float32:
		return float32SliceParser(t)
	case float64:
		return float64Parser(t)
	case []float64:
		return float64SliceParser(t)
	default:
		return nil
	}
}

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

type float32SliceParser []float32

func (i float32SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := float32SliceOrDefault(key, defaltVal.([]float32), sep)

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

type int8Parser int8

func (i int8Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := int8OrDefault(key, defaltVal.(int8))

	return val
}

type int16Parser int16

func (i int16Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := int16OrDefault(key, defaltVal.(int16))

	return val
}

type int32Parser int32

func (i int32Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := int32OrDefault(key, defaltVal.(int32))

	return val
}

type int8SliceParser []int8

func (i int8SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := int8SliceOrDefault(key, defaltVal.([]int8), sep)

	return val
}

type int16SliceParser []int16

func (i int16SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := int16SliceOrDefault(key, defaltVal.([]int16), sep)

	return val
}

type int32SliceParser []int32

func (i int32SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := int32SliceOrDefault(key, defaltVal.([]int32), sep)

	return val
}

type int64SliceParser []int64

func (i int64SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := int64SliceOrDefault(key, defaltVal.([]int64), sep)

	return val
}

type float32Parser float32

func (f float32Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := float32OrDefault(key, defaltVal.(float32))

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

type timeSliceParser []time.Time

func (t timeSliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	layout := options.Layout
	sep := options.Separator

	val := timeSliceOrDefault(key, defaltVal.([]time.Time), layout, sep)

	return val
}

type durationSliceParser []time.Duration

func (t durationSliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := durationSliceOrDefault(key, defaltVal.([]time.Duration), sep)

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

type uint8Parser uint

func (d uint8Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := uint8OrDefault(key, defaltVal.(uint8))

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

type uint8SliceParser []uint8

func (i uint8SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := uint8SliceOrDefault(key, defaltVal.([]uint8), sep)

	return val
}

type uint32SliceParser []uint32

func (i uint32SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := uint32SliceOrDefault(key, defaltVal.([]uint32), sep)

	return val
}

type uint16SliceParser []uint16

func (i uint16SliceParser) ParseEnv(key string, defaltVal any, options Parameters) any {
	sep := options.Separator

	val := uint16SliceOrDefault(key, defaltVal.([]uint16), sep)

	return val
}

type uint16Parser uint

func (d uint16Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := uint16OrDefault(key, defaltVal.(uint16))

	return val
}

type uint32Parser uint

func (d uint32Parser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := uint32OrDefault(key, defaltVal.(uint32))

	return val
}

type urlParser url.URL

func (t urlParser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := urlOrDefault(key, defaltVal.(url.URL))

	return val
}

type urlSliceParser []url.URL

func (t urlSliceParser) ParseEnv(key string, defaltVal any, opts Parameters) any {
	separator := opts.Separator

	val := urlSliceOrDefault(key, defaltVal.([]url.URL), separator)

	return val
}

type ipParser net.IP

func (t ipParser) ParseEnv(key string, defaltVal any, _ Parameters) any {
	val := ipOrDefault(key, defaltVal.(net.IP))

	return val
}

type ipSliceParser []net.IP

func (t ipSliceParser) ParseEnv(key string, defaltVal any, opts Parameters) any {
	separator := opts.Separator

	val := ipSliceOrDefault(key, defaltVal.([]net.IP), separator)

	return val
}
