// Package service implements instagram account operations and business logic.
package service

import (
	"errors"
	"fmt"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

var (
	// ErrLimitExceed returned when limit for action exceeded.
	ErrLimitExceed = errors.New("limit exceeded")
	// ErrCorrupted returned when instagram returned error response more than one time during loop processing.
	ErrCorrupted = errors.New("unable to continue - instagram responses with errors")
	// ErrNoUsers means that no users found.
	ErrNoUsers = errors.New("no users")
	// ErrNoUsernamesPassed returns when usernames list is empty.
	ErrNoUsernamesPassed = errors.New("no usernames passed")
	// ErrUserInWhitelist means that user skipped.
	ErrUserInWhitelist = errors.New("user in whitelist")
	// ErrUserNotFound returned when user not found.
	ErrUserNotFound = errors.New("user not found")
)

func makeNoUsersError(t models.UsersBatchType) error {
	return fmt.Errorf("%s: %w", t.String(), ErrNoUsers)
}
