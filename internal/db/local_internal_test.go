package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

func Test_localDB_GetLastUsersBatchByType(t *testing.T) {
	const notExistBatch models.UsersBatchType = 999

	type args struct {
		batchType models.UsersBatchType
	}

	tests := []struct {
		name    string
		args    args
		want    models.UsersBatch
		wantErr bool
	}{
		{
			name: "get followers",
			args: args{
				batchType: models.UsersBatchTypeFollowers,
			},
			want: models.UsersBatch{
				Users:     followersFixture2,
				Type:      models.UsersBatchTypeFollowers,
				CreatedAt: time.Time{},
			},
			wantErr: false,
		},
		{
			name: "get unknown type",
			args: args{
				batchType: models.UsersBatchTypeUnknown,
			},
			want:    resetBatchTime(models.MakeUsersBatch(models.UsersBatchTypeUnknown, nil, time.Now())),
			wantErr: true,
		},
		{
			name: "get invalid type",
			args: args{
				batchType: notExistBatch,
			},
			want:    resetBatchTime(models.MakeUsersBatch(models.UsersBatchTypeUnknown, nil, time.Now())),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			l := setUpLocalDBWithFixtures(t)

			got, err := l.GetLastUsersBatchByType(context.TODO(), tt.args.batchType)
			if tt.wantErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, resetBatchTime(tt.want), resetBatchTime(got))
		})
	}
}

func Test_localDB_InsertUsersBatch(t *testing.T) {
	ldb := newLocalDB()
	bType := models.UsersBatchTypeFollowers

	gotBatch, err := ldb.GetLastUsersBatchByType(context.TODO(), bType)
	require.ErrorIs(t, err, ErrNoData)
	assert.Equal(t, resetBatchTime(models.MakeUsersBatch(bType, nil, time.Now())), resetBatchTime(gotBatch))

	goldenBatch := models.UsersBatch{
		Users:     followersFixture2,
		Type:      models.UsersBatchTypeFollowers,
		CreatedAt: time.Time{},
	}

	err = ldb.InsertUsersBatch(context.TODO(), goldenBatch)
	require.NoError(t, err)

	gotBatch, err = ldb.GetLastUsersBatchByType(context.TODO(), bType)
	require.NoError(t, err)

	assert.Equal(t, resetBatchTime(goldenBatch), resetBatchTime(gotBatch))
}
