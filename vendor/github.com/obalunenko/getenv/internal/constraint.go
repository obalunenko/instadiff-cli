package internal

import (
	"net"
	"net/url"
	"time"
)

type (
	// EnvParsable is a constraint for types that can be parsed from environment variable.
	EnvParsable interface {
		String | Int | Uint | Float | Time | Bool | URL | IP | Complex
	}

	// String is a constraint for string and slice of strings.
	String interface {
		string | []string
	}

	// Int is a constraint for integer and slice of integers.
	Int interface {
		int | []int | int8 | []int8 | int16 | []int16 | int32 | []int32 | int64 | []int64
	}

	// Uint is a constraint for unsigned integer and slice of unsigned integers.
	Uint interface {
		uint | []uint | uint8 | []uint8 | uint16 | []uint16 | uint32 | []uint32 | uint64 | []uint64 | uintptr | []uintptr
	}

	// Float is a constraint for float and slice of floats.
	Float interface {
		float32 | []float32 | float64 | []float64
	}

	// Time is a constraint for time.Time and slice of time.Time.
	Time interface {
		time.Time | []time.Time | time.Duration | []time.Duration
	}

	// Bool is a constraint for bool and slice of bool.
	Bool interface {
		bool | []bool
	}

	// URL is a constraint for url.URL and slice of url.URL.
	URL interface {
		url.URL | []url.URL
	}

	// IP is a constraint for net.IP and slice of net.IP.
	IP interface {
		net.IP | []net.IP
	}

	// Complex is a constraint for complex and slice of complex.
	Complex interface {
		complex64 | []complex64 | complex128 | []complex128
	}
)
