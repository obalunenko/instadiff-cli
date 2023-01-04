// Package media provides list of media types that could be uploaded.
package media

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=Type -trimprefix=Type,type -linecomment

const (
	// TypeUndefined represents undefined media.
	TypeUndefined Type = iota // undefined

	// TypeStoryPhoto represents story photo media.
	TypeStoryPhoto // story_photo

	// typeSentinel should be always last, marks boundary of valid values.
	typeSentinel // sentinel
)

// Type represents instagram  media upload type.
type Type uint

// Valid checks if Type value is in valid boundaries.
func (t Type) Valid() bool {
	return t > TypeUndefined && t < typeSentinel
}

// Parse parses Type from string.
func Parse(v string) (Type, error) {
	var (
		err error
		mt  Type
	)

	switch {
	case strings.EqualFold(v, TypeStoryPhoto.String()):
		mt = TypeStoryPhoto
	default:
		mt = TypeUndefined

		err = fmt.Errorf("unknown Type (%s)", v)
	}

	return mt, err
}
