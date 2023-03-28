package internal

import (
	"net"
	"net/url"
	"time"
)

type (
	// EnvParsable is a constraint for supported environment variable types parsers.
	EnvParsable interface {
		String | Int | Uint | Float | Time | bool | url.URL | []url.URL | net.IP | []net.IP
	}

	// String is a constraint for strings and slice of strings.
	String interface {
		string | []string
	}

	// Int is a constraint for integer and slice of integers.
	Int interface {
		int | []int | int8 | []int8 | int16 | []int16 | int32 | []int32 | int64 | []int64
	}

	// Uint is a constraint for unsigned integer and slice of unsigned integers.
	Uint interface {
		uint | []uint | uint8 | []uint8 | uint16 | []uint16 | uint32 | []uint32 | uint64 | []uint64
	}

	// Float is a constraint for floats and slice of floats.
	Float interface {
		float32 | []float32 | float64 | []float64
	}

	// Time is a constraint for time.Time and time.Duration.
	Time interface {
		time.Time | []time.Time | time.Duration | []time.Duration
	}
)
