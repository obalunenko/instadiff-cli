package internal

import (
	"time"
)

// EnvParsable is a constraint for supported environment variable types parsers.
type EnvParsable interface {
	String | Int | Float | Time | bool
}

// String is a constraint for strings and slice of strings.
type String interface {
	string | []string
}

// Int is a constraint for integer and slice of integers.
type Int interface {
	int | []int | int64 | []int64 | uint64 | []uint64
}

// Float is a constraint for floats and slice of floats.
type Float interface {
	float64 | []float64
}

// Time is a constraint for time.Time and time.Duration.
type Time interface {
	time.Time | time.Duration
}
