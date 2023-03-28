package internal

import (
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// intOrDefault retrieves the int value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func intOrDefault(key string, defaultVal int) int {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(env)
	if err != nil {
		return defaultVal
	}

	return val
}

// stringOrDefault retrieves the string value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func stringOrDefault(key, defaultVal string) string {
	env, ok := os.LookupEnv(key)
	if !ok || env == "" {
		return defaultVal
	}

	return env
}

// boolOrDefault retrieves the bool value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func boolOrDefault(key string, defaultVal bool) bool {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := strconv.ParseBool(env)
	if err != nil {
		return defaultVal
	}

	return val
}

// stringSliceOrDefault retrieves the string slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func stringSliceOrDefault(key string, defaultVal []string, sep string) []string {
	if sep == "" {
		return defaultVal
	}

	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val := strings.Split(env, sep)

	return val
}

// intSliceOrDefault retrieves the int slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func intSliceOrDefault(key string, defaultVal []int, sep string) []int {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]int, 0, len(valraw))

	for _, s := range valraw {
		v, err := strconv.Atoi(s)
		if err != nil {
			return defaultVal
		}

		val = append(val, v)
	}

	return val
}

// float32SliceOrDefault retrieves the float32 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func float32SliceOrDefault(key string, defaultVal []float32, sep string) []float32 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]float32, 0, len(valraw))

	const (
		bitsize = 32
	)

	for _, s := range valraw {
		v, err := strconv.ParseFloat(s, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, float32(v))
	}

	return val
}

// float64SliceOrDefault retrieves the float64 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func float64SliceOrDefault(key string, defaultVal []float64, sep string) []float64 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]float64, 0, len(valraw))

	const (
		bitsize = 64
	)

	for _, s := range valraw {
		v, err := strconv.ParseFloat(s, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, v)
	}

	return val
}

// int64SliceOrDefault retrieves the int64 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func int64SliceOrDefault(key string, defaultVal []int64, sep string) []int64 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]int64, 0, len(valraw))

	const (
		base    = 10
		bitsize = 64
	)

	for _, s := range valraw {
		v, err := strconv.ParseInt(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, v)
	}

	return val
}

// int8SliceOrDefault retrieves the int8 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func int8SliceOrDefault(key string, defaultVal []int8, sep string) []int8 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]int8, 0, len(valraw))

	const (
		base    = 10
		bitsize = 8
	)

	for _, s := range valraw {
		v, err := strconv.ParseInt(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, int8(v))
	}

	return val
}

// int32SliceOrDefault retrieves the int32 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func int32SliceOrDefault(key string, defaultVal []int32, sep string) []int32 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]int32, 0, len(valraw))

	const (
		base    = 10
		bitsize = 32
	)

	for _, s := range valraw {
		v, err := strconv.ParseInt(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, int32(v))
	}

	return val
}

// int16SliceOrDefault retrieves the int16 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func int16SliceOrDefault(key string, defaultVal []int16, sep string) []int16 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]int16, 0, len(valraw))

	const (
		base    = 10
		bitsize = 16
	)

	for _, s := range valraw {
		v, err := strconv.ParseInt(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, int16(v))
	}

	return val
}

// durationOrDefault retrieves the time.Duration value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func durationOrDefault(key string, defaultVal time.Duration) time.Duration {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := time.ParseDuration(env)
	if err != nil {
		return defaultVal
	}

	return val
}

// timeOrDefault retrieves the time.Time value of the environment variable named
// by the key represented by layout.
// If variable not set or value is empty - defaultVal will be returned.
func timeOrDefault(key string, defaultVal time.Time, layout string) time.Time {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := time.Parse(layout, env)
	if err != nil {
		return defaultVal
	}

	return val
}

// timeSliceOrDefault retrieves the []time.Time value of the environment variable named
// by the key represented by layout.
// If variable not set or value is empty - defaultVal will be returned.
func timeSliceOrDefault(key string, defaultVal []time.Time, layout, separator string) []time.Time {
	valraw := stringSliceOrDefault(key, nil, separator)
	if valraw == nil {
		return defaultVal
	}

	val := make([]time.Time, 0, len(valraw))

	for _, s := range valraw {
		v, err := time.Parse(layout, s)
		if err != nil {
			return defaultVal
		}

		val = append(val, v)
	}

	return val
}

// durationSliceOrDefault retrieves the []time.Duration value of the environment variable named
// by the key represented by layout.
// If variable not set or value is empty - defaultVal will be returned.
func durationSliceOrDefault(key string, defaultVal []time.Duration, separator string) []time.Duration {
	valraw := stringSliceOrDefault(key, nil, separator)
	if valraw == nil {
		return defaultVal
	}

	val := make([]time.Duration, 0, len(valraw))

	for _, s := range valraw {
		v, err := time.ParseDuration(s)
		if err != nil {
			return defaultVal
		}

		val = append(val, v)
	}

	return val
}

// int64OrDefault retrieves the int64 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func int64OrDefault(key string, defaultVal int64) int64 {
	env := stringOrDefault(key, "")
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

// int8OrDefault retrieves the int8 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func int8OrDefault(key string, defaultVal int8) int8 {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 8
	)

	val, err := strconv.ParseInt(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return int8(val)
}

// int16OrDefault retrieves the int16 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func int16OrDefault(key string, defaultVal int16) int16 {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 16
	)

	val, err := strconv.ParseInt(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return int16(val)
}

// int32OrDefault retrieves the int32 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func int32OrDefault(key string, defaultVal int32) int32 {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 32
	)

	val, err := strconv.ParseInt(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return int32(val)
}

// float32OrDefault retrieves the float32 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func float32OrDefault(key string, defaultVal float32) float32 {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		bitsize = 32
	)

	val, err := strconv.ParseFloat(env, bitsize)
	if err != nil {
		return defaultVal
	}

	return float32(val)
}

// float64OrDefault retrieves the float64 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func float64OrDefault(key string, defaultVal float64) float64 {
	env := stringOrDefault(key, "")
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

// uint64OrDefault retrieves the unt64 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func uint64OrDefault(key string, defaultVal uint64) uint64 {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 64
	)

	val, err := strconv.ParseUint(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return val
}

// uint64SliceOrDefault retrieves the uint64 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func uint64SliceOrDefault(key string, defaultVal []uint64, sep string) []uint64 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]uint64, 0, len(valraw))

	const (
		base    = 10
		bitsize = 64
	)

	for _, s := range valraw {
		v, err := strconv.ParseUint(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, v)
	}

	return val
}

// uint8OrDefault retrieves the unt8 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func uint8OrDefault(key string, defaultVal uint8) uint8 {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 8
	)

	val, err := strconv.ParseUint(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return uint8(val)
}

// uintOrDefault retrieves the unt value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func uintOrDefault(key string, defaultVal uint) uint {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 32
	)

	val, err := strconv.ParseUint(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return uint(val)
}

// uintSliceOrDefault retrieves the uint slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func uintSliceOrDefault(key string, defaultVal []uint, sep string) []uint {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]uint, 0, len(valraw))

	const (
		base    = 10
		bitsize = 32
	)

	for _, s := range valraw {
		v, err := strconv.ParseUint(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, uint(v))
	}

	return val
}

// uint8SliceOrDefault retrieves the uint8 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func uint8SliceOrDefault(key string, defaultVal []uint8, sep string) []uint8 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]uint8, 0, len(valraw))

	const (
		base    = 10
		bitsize = 8
	)

	for _, s := range valraw {
		v, err := strconv.ParseUint(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, uint8(v))
	}

	return val
}

// uint16SliceOrDefault retrieves the uint16 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func uint16SliceOrDefault(key string, defaultVal []uint16, sep string) []uint16 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]uint16, 0, len(valraw))

	const (
		base    = 10
		bitsize = 16
	)

	for _, s := range valraw {
		v, err := strconv.ParseUint(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, uint16(v))
	}

	return val
}

// uint32SliceOrDefault retrieves the uint32 slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func uint32SliceOrDefault(key string, defaultVal []uint32, sep string) []uint32 {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]uint32, 0, len(valraw))

	const (
		base    = 10
		bitsize = 32
	)

	for _, s := range valraw {
		v, err := strconv.ParseUint(s, base, bitsize)
		if err != nil {
			return defaultVal
		}

		val = append(val, uint32(v))
	}

	return val
}

// uint16OrDefault retrieves the unt16 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func uint16OrDefault(key string, defaultVal uint16) uint16 {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 16
	)

	val, err := strconv.ParseUint(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return uint16(val)
}

// uint32OrDefault retrieves the unt32 value of the environment variable named
// by the key.
// If variable not set or value is empty - defaultVal will be returned.
func uint32OrDefault(key string, defaultVal uint32) uint32 {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	const (
		base    = 10
		bitsize = 32
	)

	val, err := strconv.ParseUint(env, base, bitsize)
	if err != nil {
		return defaultVal
	}

	return uint32(val)
}

// urlOrDefault retrieves the url.URL value of the environment variable named
// by the key represented by layout.
// If variable not set or value is empty - defaultVal will be returned.
func urlOrDefault(key string, defaultVal url.URL) url.URL {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val, err := url.Parse(env)
	if err != nil {
		return defaultVal
	}

	return *val
}

// urlSliceOrDefault retrieves the url.URL slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func urlSliceOrDefault(key string, defaultVal []url.URL, sep string) []url.URL {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]url.URL, 0, len(valraw))

	for _, s := range valraw {
		v, err := url.Parse(s)
		if err != nil {
			return defaultVal
		}

		val = append(val, *v)
	}

	return val
}

// ipOrDefault retrieves the net.IP value of the environment variable named
// by the key represented by layout.
// If variable not set or value is empty - defaultVal will be returned.
func ipOrDefault(key string, defaultVal net.IP) net.IP {
	env := stringOrDefault(key, "")
	if env == "" {
		return defaultVal
	}

	val := net.ParseIP(env)
	if val == nil {
		return defaultVal
	}

	return val
}

// ipSliceOrDefault retrieves the net.IP slice value of the environment variable named
// by the key and separated by sep.
// If variable not set or value is empty - defaultVal will be returned.
func ipSliceOrDefault(key string, defaultVal []net.IP, sep string) []net.IP {
	valraw := stringSliceOrDefault(key, nil, sep)
	if valraw == nil {
		return defaultVal
	}

	val := make([]net.IP, 0, len(valraw))

	for _, s := range valraw {
		v := net.ParseIP(s)
		if v == nil {
			return defaultVal
		}

		val = append(val, v)
	}

	return val
}
