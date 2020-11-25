// Package db implements database interactions.
package db

import (
	"context"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

type localDB struct {
	users map[models.UsersBatchType]models.UsersBatch
}

func newLocalDB() *localDB {
	return &localDB{
		users: make(map[models.UsersBatchType]models.UsersBatch),
	}
}

func (l *localDB) InsertUsersBatch(ctx context.Context, users models.UsersBatch) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if !users.Type.Valid() {
			return models.MakeInvalidBatchTypeError(users.Type)
		}

		l.users[users.Type] = users

		return nil
	}
}

func (l *localDB) GetLastUsersBatchByType(ctx context.Context,
	batchType models.UsersBatchType) (models.UsersBatch, error) {
	select {
	case <-ctx.Done():
		return models.EmptyUsersBatch, ctx.Err()
	default:
		if !batchType.Valid() {
			return models.EmptyUsersBatch, models.MakeInvalidBatchTypeError(batchType)
		}

		batch, exist := l.users[batchType]
		if !exist {
			return models.EmptyUsersBatch, ErrNoData
		}

		return batch, nil
	}
}
