package internal

import (
	"errors"
	"fmt"
)

var (
	// ErrNotSet is an error that is returned when the environment variable is not set.
	ErrNotSet = errors.New("not set")
	// ErrInvalidValue is an error that is returned when the environment variable is not valid.
	ErrInvalidValue = errors.New("invalid value")
)

func newErrInvalidValue(msg string) error {
	return newWrapErr(msg, ErrInvalidValue)
}

func newErrNotSet(msg string) error {
	return newWrapErr(msg, ErrNotSet)
}

func newWrapErr(msg string, wrapErr error) error {
	return fmt.Errorf("%s: %w", msg, wrapErr)
}
