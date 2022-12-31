package media

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=Type -trimprefix=Type,type -linecomment

const (
	TypeUndefined Type = iota // undefined

	TypeStoryPhoto // story_photo

	typeSentinel // sentinel
)

type Type uint

func (t Type) Valid() bool {
	return t > TypeUndefined && t < typeSentinel
}

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
