// Package db implements database interactions.
package db

import (
	"context"
	"time"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

type localDB struct {
	users map[models.UsersBatchType][]models.UsersBatch
}

func (l *localDB) Close(_ context.Context) error {
	return nil
}

func newLocalDB() *localDB {
	return &localDB{
		users: make(map[models.UsersBatchType][]models.UsersBatch),
	}
}

func (l *localDB) InsertUsersBatch(ctx context.Context, users models.UsersBatch) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		bt := users.Type

		if !bt.Valid() {
			return models.MakeInvalidBatchTypeError(bt)
		}

		s := l.users[bt]

		s = append(s, users)

		l.users[bt] = s

		return nil
	}
}

func (l *localDB) GetLastUsersBatchByType(ctx context.Context, bt models.UsersBatchType) (models.UsersBatch, error) {
	select {
	case <-ctx.Done():
		return models.MakeUsersBatch(bt, nil, time.Now()), ctx.Err()
	default:
		if !bt.Valid() {
			return models.MakeUsersBatch(bt, nil, time.Now()), models.MakeInvalidBatchTypeError(bt)
		}

		batches, exist := l.users[bt]
		if !exist || len(batches) == 0 {
			return models.MakeUsersBatch(bt, nil, time.Now()), ErrNoData
		}

		last := batches[len(batches)-1]

		return last, nil
	}
}

func (l *localDB) GetAllUsersBatchByType(ctx context.Context, batchType models.UsersBatchType) ([]models.UsersBatch, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if !batchType.Valid() {
			return nil, models.MakeInvalidBatchTypeError(batchType)
		}

		batches, exist := l.users[batchType]
		if !exist || len(batches) == 0 {
			return nil, ErrNoData
		}

		return batches, nil
	}
}
