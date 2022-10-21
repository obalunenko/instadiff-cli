package config

import (
	"errors"
)

// ErrEmptyPath returned when empty path is passed.
var (
	ErrEmptyPath = errors.New("config path is empty")
)
