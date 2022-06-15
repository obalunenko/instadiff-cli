package db

import (
	"errors"
)

// ErrNoData returned when no data found in collection.
var ErrNoData = errors.New("no data in collection")
