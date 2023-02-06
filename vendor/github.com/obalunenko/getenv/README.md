![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/obalunenko/getenv)
[![Go Reference](https://pkg.go.dev/badge/github.com/obalunenko/getenv.svg)](https://pkg.go.dev/github.com/obalunenko/getenv)
[![Go Report Card](https://goreportcard.com/badge/github.com/obalunenko/getenv)](https://goreportcard.com/report/github.com/obalunenko/getenv)
[![codecov](https://codecov.io/gh/obalunenko/getenv/branch/master/graph/badge.svg)](https://codecov.io/gh/obalunenko/getenv)
![coverbadger-tag-do-not-edit](https://img.shields.io/badge/coverage-96.48%25-brightgreen?longCache=true&style=flat)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=obalunenko_getenv&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=obalunenko_getenv)

# getenv

Package getenv provides functionality for loading environment variables and parse them into go builtin types.

Types supported:

```text
- string
- []string
- int
- []int
- int64
- []int64
- uint64
- []uint64
- uint
- []uint
- uint32
- []uint32
- float64
- []float64
- time.Time
- time.Duration
- bool
```

## Examples

### EnvOrDefault

EnvOrDefault retrieves the value of the environment variable named
by the key.
If variable not set or value is empty - defaultVal will be returned.

```golang
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/obalunenko/getenv"
	"github.com/obalunenko/getenv/option"
)

func main() {
	key := "GH_GETENV_TEST"

	defer func() {
		if err := os.Unsetenv("GH_GETENV_TEST"); err != nil {
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

}

```

 Output:

```
[string]: golly
[int]: 123
[time.Time]: 2022-01-20 00:00:00 +0000 UTC
[[]float64]: [26.89 0.67]
[time.Duration]: 2h35m0s
```
