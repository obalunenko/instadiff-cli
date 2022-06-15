package models

import (
	"errors"
	"fmt"
)

// ErrInvalidUsersBatchType means that batch type not supported.
var ErrInvalidUsersBatchType = errors.New("invalid users batch type")

// MakeInvalidBatchTypeError returns ErrInvalidUsersBatchType with added bathtype info.
func MakeInvalidBatchTypeError(t UsersBatchType) error {
	return fmt.Errorf("%s: %w", t.String(), ErrInvalidUsersBatchType)
}
