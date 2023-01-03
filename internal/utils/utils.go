// Package utils provide useful common functionality.
package utils

import (
	"context"

	log "github.com/obalunenko/logger"
)

// LogError helper for closure funcs error handling.
func LogError(ctx context.Context, err error, msg string) {
	if err != nil {
		log.WithError(ctx, err).Error(msg)
	}
}
