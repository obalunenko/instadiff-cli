package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	reset := SetUpMongoContainer(ctx, m, "5.0.9", ContainerParams{
		User:          "user",
		UserPassword:  "pwd",
		ExpireSeconds: 30,
	})

	code := m.Run()

	reset()

	os.Exit(code)
}

func TestMongoDB(t *testing.T) {
	ctx := context.Background()

	dbc := ConnectForTesting(t, "", "users")

	now := time.Now()

	tm := now.AddDate(0, 0, -2)

	b := models.UsersBatch{
		Users:     followersFixture1,
		Type:      models.UsersBatchTypeFollowers,
		CreatedAt: tm,
	}

	err := dbc.InsertUsersBatch(ctx, b)
	require.NoError(t, err)

	gotbatches, err := dbc.GetAllUsersBatchByType(ctx, models.UsersBatchTypeFollowings)
	require.NoError(t, err)

	assert.Empty(t, gotbatches)

	gotbatches, err = dbc.GetAllUsersBatchByType(ctx, models.UsersBatchTypeFollowers)
	require.NoError(t, err)

	assert.Equal(t, resetBatchesTime([]models.UsersBatch{b}), resetBatchesTime(gotbatches))

	gotbatch, err := dbc.GetLastUsersBatchByType(ctx, models.UsersBatchTypeFollowings)
	require.ErrorIs(t, err, ErrNoData)

	assert.Equal(t, resetBatchTime(models.MakeUsersBatch(models.UsersBatchTypeFollowings, nil, time.Now())), resetBatchTime(gotbatch))

	gotbatch, err = dbc.GetLastUsersBatchByType(ctx, models.UsersBatchTypeFollowers)
	require.NoError(t, err)

	assert.Equal(t, resetBatchTime(b), resetBatchTime(gotbatch))

	b2 := models.UsersBatch{
		Users:     followersFixture2,
		Type:      models.UsersBatchTypeFollowers,
		CreatedAt: now,
	}

	err = dbc.InsertUsersBatch(ctx, b2)
	require.NoError(t, err)

	gotbatches, err = dbc.GetAllUsersBatchByType(ctx, models.UsersBatchTypeFollowers)
	require.NoError(t, err)

	assert.Equal(t, resetBatchesTime([]models.UsersBatch{b2, b}), resetBatchesTime(gotbatches))

	gotbatch, err = dbc.GetLastUsersBatchByType(ctx, models.UsersBatchTypeFollowers)
	require.NoError(t, err)

	assert.Equal(t, resetBatchTime(b2), resetBatchTime(gotbatch))
}
