// Package service implements instagram account operations and business logic.
package service

import (
	"errors"
	"fmt"

	"github.com/oleg-balunenko/instadiff-cli/internal/models"
)

// ErrNoUsers means that no users found.
var ErrNoUsers = errors.New("no users")

func makeNoUsersError(t models.UsersBatchType) error {
	return fmt.Errorf("%s: %w", t.String(), ErrNoUsers)
}
