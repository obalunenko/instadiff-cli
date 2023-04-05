![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/obalunenko/getenv)
[![Go Reference](https://pkg.go.dev/badge/github.com/obalunenko/getenv.svg)](https://pkg.go.dev/github.com/obalunenko/getenv)
[![Go Report Card](https://goreportcard.com/badge/github.com/obalunenko/getenv)](https://goreportcard.com/report/github.com/obalunenko/getenv)
[![codecov](https://codecov.io/gh/obalunenko/getenv/branch/master/graph/badge.svg)](https://codecov.io/gh/obalunenko/getenv)
![coverbadger-tag-do-not-edit](https://img.shields.io/badge/coverage-100%25-brightgreen?longCache=true&style=flat)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=obalunenko_getenv&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=obalunenko_getenv)

# getenv

Package getenv provides a simple way to get environment variables.
It's type-safe and supports built-in types and slices of them.

Types supported:
- string
- []string
- int
- []int
- int8
- []int8
- int16
- []int16
- int32
- []int32
- int64
- []int64
- uint8
- []uint8
- uint16
- []uint16
- uint64
- []uint64
- uint
- []uint
- uintptr
- []uintptr
- uint32
- []uint32
- float32
- []float32
- float64
- []float64
- time.Time
- []time.Time
- time.Duration
- []time.Duration
- bool
- []bool
- url.URL
- []url.URL
- net.IP
- []net.IP
- complex64
- []complex64
- complex128
- []complex128

## Examples

### Env

Env retrieves the value of the environment variable named by the key.
If the variable is present in the environment the value will be parsed and returned.
Otherwise, an error will be returned.

```golang
package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/obalunenko/getenv"
	"github.com/obalunenko/getenv/option"
)

func main() {
	const (
		key = "GH_GETENV_TEST"
	)

	var (
		val any
		err error
	)

	defer func() {
		if err = os.Unsetenv(key); err != nil {
			panic(err)
		}
	}()

	// string
	if err = os.Setenv(key, "golly"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[string](key)
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// int
	if err = os.Setenv(key, "123"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[int](key)
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// int conversion error
	if err = os.Setenv(key, "123s4"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[int](key)
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// time.Time
	if err = os.Setenv(key, "2022-01-20"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[time.Time](key, option.WithTimeLayout("2006-01-02"))
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// []float64
	if err = os.Setenv(key, "26.89,0.67"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[[]float64](key, option.WithSeparator(","))
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// time.Duration
	if err = os.Setenv(key, "2h35m"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[time.Duration](key)
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// url.URL
	if err = os.Setenv(key, "https://test:abcd123@golangbyexample.com:8000/tutorials/intro?type=advance&compact=false#history"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[url.URL](key)
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// net.IP
	if err = os.Setenv(key, "2001:cb8::17"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[net.IP](key)
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// []string
	if err = os.Setenv(key, "a,b,c,d"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[[]string](key, option.WithSeparator(","))
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// complex128
	if err = os.Setenv(key, "1+2i"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[complex128](key)
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

	// []complex64
	if err = os.Setenv(key, "1+2i,3+4i"); err != nil {
		panic(err)
	}

	val, err = getenv.Env[[]complex64](key, option.WithSeparator(","))
	fmt.Printf("[%T]: %v; err: %v\n", val, val, err)

}

```

Output:

```
[string]: golly; err: <nil>
[int]: 123; err: <nil>
[int]: 0; err: could not parse variable[GH_GETENV_TEST] value[123s4] to type[int]: invalid value
[time.Time]: 2022-01-20 00:00:00 +0000 UTC; err: <nil>
[[]float64]: [26.89 0.67]; err: <nil>
[time.Duration]: 2h35m0s; err: <nil>
[url.URL]: {https  test:abcd123 golangbyexample.com:8000 /tutorials/intro  false false type=advance&compact=false history }; err: <nil>
[net.IP]: 2001:cb8::17; err: <nil>
[[]string]: [a b c d]; err: <nil>
[complex128]: (1+2i); err: <nil>
[[]complex64]: [(1+2i) (3+4i)]; err: <nil>
```

### EnvOrDefault

EnvOrDefault retrieves the value of the environment variable named by the key.
If the variable is present in the environment the value will be parsed and returned.
Otherwise, the default value will be returned.
The value returned will be of the same type as the default value.

```golang
package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/obalunenko/getenv"
	"github.com/obalunenko/getenv/option"
)

func main() {
	const (
		key = "GH_GETENV_TEST"
	)

	defer func() {
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}()

	var val any

	// string
	if err := os.Setenv(key, "golly"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, "golly")
	fmt.Printf("[%T]: %v\n", val, val)

	// int
	if err := os.Setenv(key, "123"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, -99)
	fmt.Printf("[%T]: %v\n", val, val)

	// time.Time
	if err := os.Setenv(key, "2022-01-20"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key,
		time.Date(1992, 12, 1, 0, 0, 0, 0, time.UTC),
		option.WithTimeLayout("2006-01-02"),
	)
	fmt.Printf("[%T]: %v\n", val, val)

	// []float64
	if err := os.Setenv(key, "26.89,0.67"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, []float64{-99},
		option.WithSeparator(","),
	)
	fmt.Printf("[%T]: %v\n", val, val)

	// time.Duration
	if err := os.Setenv(key, "2h35m"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, time.Second)
	fmt.Printf("[%T]: %v\n", val, val)

	// url.URL
	if err := os.Setenv(key, "https://test:abcd123@golangbyexample.com:8000/tutorials/intro?type=advance&compact=false#history"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, url.URL{})
	fmt.Printf("[%T]: %v\n", val, val)

	// net.IP
	if err := os.Setenv(key, "2001:cb8::17"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, net.IP{})
	fmt.Printf("[%T]: %v\n", val, val)

	// []string
	if err := os.Setenv(key, "a,b,c,d"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, []string{}, option.WithSeparator(","))
	fmt.Printf("[%T]: %v\n", val, val)

	// complex128
	if err := os.Setenv(key, "1+2i"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, complex128(0))
	fmt.Printf("[%T]: %v\n", val, val)

	// []complex64
	if err := os.Setenv(key, "1+2i,3+4i"); err != nil {
		panic(err)
	}

	val = getenv.EnvOrDefault(key, []complex64{}, option.WithSeparator(","))
	fmt.Printf("[%T]: %v\n", val, val)

}

```

Output:

```
[string]: golly
[int]: 123
[time.Time]: 2022-01-20 00:00:00 +0000 UTC
[[]float64]: [26.89 0.67]
[time.Duration]: 2h35m0s
[url.URL]: {https  test:abcd123 golangbyexample.com:8000 /tutorials/intro  false false type=advance&compact=false history }
[net.IP]: 2001:cb8::17
[[]string]: [a b c d]
[complex128]: (1+2i)
[[]complex64]: [(1+2i) (3+4i)]
```
