package internal

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func getString(key string) (string, error) {
	env, ok := os.LookupEnv(key)
	if !ok || env == "" {
		return "", newErrNotSet(fmt.Sprintf("%q", key))
	}

	return env, nil
}

func getBool(key string) (bool, error) {
	env, err := getString(key)
	if err != nil {
		return false, err
	}

	val, err := strconv.ParseBool(env)
	if err != nil {
		return false, newErrInvalidValue(err.Error())
	}

	return val, nil
}

func getBoolSlice(key, sep string) ([]bool, error) {
	if sep == "" {
		return nil, ErrInvalidValue
	}

	env, err := getString(key)
	if err != nil {
		return nil, err
	}

	val := strings.Split(env, sep)

	b := make([]bool, 0, len(val))

	for _, s := range val {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return nil, newErrInvalidValue(err.Error())
		}

		b = append(b, v)
	}

	return b, nil
}

func getStringSlice(key, sep string) ([]string, error) {
	if sep == "" {
		return nil, ErrInvalidValue
	}

	env, err := getString(key)
	if err != nil {
		return nil, err
	}

	val := strings.Split(env, sep)

	return val, nil
}

func parseNumberGen[T Number](raw string) (T, error) {
	var tt T

	const (
		base = 10
	)

	rt := reflect.TypeOf(tt)

	switch rt.Kind() { //nolint:exhaustive // All supported types are covered.
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(raw, base, rt.Bits())
		if err != nil {
			return tt, ErrInvalidValue
		}

		return any(T(val)).(T), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		val, err := strconv.ParseUint(raw, base, rt.Bits())
		if err != nil {
			return tt, ErrInvalidValue
		}

		return any(T(val)).(T), nil
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(raw, rt.Bits())
		if err != nil {
			return tt, ErrInvalidValue
		}

		return any(T(val)).(T), nil
	default:
		return tt, ErrInvalidValue
	}
}

func parseNumberSliceGen[S []T, T Number](raw []string) (S, error) {
	var tt S

	val := make(S, 0, len(raw))

	for _, s := range raw {
		v, err := parseNumberGen[T](s)
		if err != nil {
			return tt, err
		}

		val = append(val, v)
	}

	return val, nil
}

func getNumberSliceGen[S []T, T Number](key, sep string) (S, error) {
	env, err := getStringSlice(key, sep)
	if err != nil {
		return nil, err
	}

	return parseNumberSliceGen[S, T](env)
}

func getNumberGen[T Number](key string) (T, error) {
	env, err := getString(key)
	if err != nil {
		return 0, err
	}

	return parseNumberGen[T](env)
}

func getDuration(key string) (time.Duration, error) {
	env, err := getString(key)
	if err != nil {
		return 0, err
	}

	val, err := time.ParseDuration(env)
	if err != nil {
		return 0, newErrInvalidValue(err.Error())
	}

	return val, nil
}

func getTime(key, layout string) (time.Time, error) {
	env, err := getString(key)
	if err != nil {
		return time.Time{}, err
	}

	val, err := time.Parse(layout, env)
	if err != nil {
		return time.Time{}, newErrInvalidValue(err.Error())
	}

	return val, nil
}

func getTimeSlice(key, layout, sep string) ([]time.Time, error) {
	env, err := getStringSlice(key, sep)
	if err != nil {
		return nil, err
	}

	val := make([]time.Time, 0, len(env))

	for _, s := range env {
		v, err := time.Parse(layout, s)
		if err != nil {
			return nil, newErrInvalidValue(err.Error())
		}

		val = append(val, v)
	}

	return val, nil
}

func getDurationSlice(key, sep string) ([]time.Duration, error) {
	env, err := getStringSlice(key, sep)
	if err != nil {
		return nil, err
	}

	val := make([]time.Duration, 0, len(env))

	for _, s := range env {
		v, err := time.ParseDuration(s)
		if err != nil {
			return nil, newErrInvalidValue(err.Error())
		}

		val = append(val, v)
	}

	return val, nil
}

func getURL(key string) (url.URL, error) {
	env, err := getString(key)
	if err != nil {
		return url.URL{}, err
	}

	val, err := url.Parse(env)
	if err != nil {
		return url.URL{}, newErrInvalidValue(err.Error())
	}

	return *val, nil
}

func getURLSlice(key, sep string) ([]url.URL, error) {
	env, err := getStringSlice(key, sep)
	if err != nil {
		return nil, err
	}

	val := make([]url.URL, 0, len(env))

	for _, s := range env {
		v, err := url.Parse(s)
		if err != nil {
			return nil, newErrInvalidValue(err.Error())
		}

		val = append(val, *v)
	}

	return val, nil
}

func getIP(key string) (net.IP, error) {
	env, err := getString(key)
	if err != nil {
		return nil, err
	}

	val := net.ParseIP(env)
	if val == nil {
		return nil, ErrInvalidValue
	}

	return val, nil
}

func getIPSlice(key, sep string) ([]net.IP, error) {
	env, err := getStringSlice(key, sep)
	if err != nil {
		return nil, err
	}

	val := make([]net.IP, 0, len(env))

	for _, s := range env {
		v := net.ParseIP(s)
		if v == nil {
			return nil, ErrInvalidValue
		}

		val = append(val, v)
	}

	return val, nil
}

func parseComplexGen[T Complex](raw string) (T, error) {
	var tt T

	var bitsize int

	switch any(tt).(type) {
	case complex64:
		bitsize = 64
	case complex128:
		bitsize = 128
	}

	val, err := strconv.ParseComplex(raw, bitsize)
	if err != nil {
		return tt, newErrInvalidValue(err.Error())
	}

	return any(T(val)).(T), nil
}

func parseComplexSliceGen[S []T, T Complex](raw []string) (S, error) {
	var tt S

	val := make(S, 0, len(raw))

	for _, s := range raw {
		v, err := parseComplexGen[T](s)
		if err != nil {
			return tt, err
		}

		val = append(val, v)
	}

	return val, nil
}

func getComplexSliceGen[S []T, T Complex](key, sep string) (S, error) {
	env, err := getStringSlice(key, sep)
	if err != nil {
		return nil, err
	}

	return parseComplexSliceGen[S, T](env)
}

func getComplexGen[T Complex](key string) (T, error) {
	env, err := getString(key)
	if err != nil {
		return 0, err
	}

	return parseComplexGen[T](env)
}
