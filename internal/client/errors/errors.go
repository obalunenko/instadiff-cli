// Package errors defines common client errors.
package errors

import "errors"

var (
	// ErrEmptyInput returned in case when user input is empty.
	ErrEmptyInput = errors.New("should not be empty")
	// ErrUserNotFound returned in case when user not found.
	ErrUserNotFound = errors.New("user not found")
)
