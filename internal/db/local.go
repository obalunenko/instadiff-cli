// Package db implements database interactions.
package db

import (
	"context"

	"github.com/pkg/errors"

	"github.com/oleg-balunenko/instadiff-cli/internal/models"
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
		if users.Type.Valid() {
			l.users[users.Type] = users
			return nil
		}

		return errors.Errorf("invalid users type: %s", users.Type)
	}
}

func (l *localDB) GetLastUsersBatchByType(ctx context.Context,
	batchType models.UsersBatchType) (models.UsersBatch, error) {
	select {
	case <-ctx.Done():
		return models.EmptyUsersBatch, ctx.Err()
	default:
		if batchType.Valid() {
			batch, exist := l.users[batchType]
			if !exist {
				return models.EmptyUsersBatch, nil
			}

			return batch, nil
		}

		return models.EmptyUsersBatch, errors.Errorf("invalid users type: %s", batchType)
	}
}
