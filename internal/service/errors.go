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
)

func makeNoUsersError(t models.UsersBatchType) error {
	return fmt.Errorf("%s: %w", t.String(), ErrNoUsers)
}
