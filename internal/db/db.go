// Package db implements database interactions.
package db

import (
	"context"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

// DB represents database interaction contract.
type DB interface {
	// InsertUsersBatch creates record in database with passed users batch.
	InsertUsersBatch(ctx context.Context, users models.UsersBatch) error
	// GetLastUsersBatchByType returns last created users batch by passed batch type.
	GetLastUsersBatchByType(ctx context.Context, batchType models.UsersBatchType) (models.UsersBatch, error)
	// GetAllUsersBatchByType returns all users batches by passed batch type.
	GetAllUsersBatchByType(ctx context.Context, batchType models.UsersBatchType) ([]models.UsersBatch, error)
	// Close closes connections.
	Close(ctx context.Context) error
}

// Params used for DB constructor.
type Params struct {
	LocalDB     bool
	MongoParams MongoParams
}

// Connect returns specific database connection.
// MongoDB if mongo is enabled.
// LocalMemory in other case.
func Connect(ctx context.Context, params Params) (DB, error) {
	if params.LocalDB {
		return newLocalDB(), nil
	}

	return newMongoDB(ctx, params.MongoParams)
}
