// Package db implements database interactions.
package db

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/obalunenko/instadiff-cli/internal/models"
)

// MongoParams represents mongo db configuration parameters.
type MongoParams struct {
	URL        string
	Database   string
	Collection string
}

type mongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func newMongoDB(params MongoParams) (*mongoDB, error) {
	ctx := context.TODO()

	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(params.URL))
	if err != nil {
		return nil, err
	}

	database := cl.Database(params.Database)
	collection := database.Collection(params.Collection)

	return &mongoDB{
		client:     cl,
		database:   database,
		collection: collection,
	}, nil
}

func (m mongoDB) InsertUsersBatch(ctx context.Context, users models.UsersBatch) error {
	if _, err := m.collection.InsertOne(ctx, users); err != nil {
		return fmt.Errorf("insert batch: %w", err)
	}

	return nil
}

func (m mongoDB) GetLastUsersBatchByType(ctx context.Context,
	batchType models.UsersBatchType) (models.UsersBatch, error) {
	filter := bson.M{"batch_type": batchType}
	resp := m.collection.FindOne(ctx, filter, &options.FindOneOptions{
		AllowPartialResults: nil,
		BatchSize:           nil,
		Collation:           nil,
		Comment:             nil,
		CursorType:          nil,
		Hint:                nil,
		Max:                 nil,
		MaxAwaitTime:        nil,
		MaxTime:             nil,
		Min:                 nil,
		NoCursorTimeout:     nil,
		OplogReplay:         nil,
		Projection:          nil,
		ReturnKey:           nil,
		ShowRecordID:        nil,
		Skip:                nil,
		Snapshot:            nil,
		Sort:                bson.M{"$natural": -1},
	})

	if err := resp.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.EmptyUsersBatch, ErrNoData
		}

		return models.EmptyUsersBatch, fmt.Errorf("find batch [%s]: %w", batchType.String(), err)
	}

	var ub models.UsersBatch

	if err := resp.Decode(&ub); err != nil {
		return models.EmptyUsersBatch, fmt.Errorf("decode response: %w", err)
	}

	return ub, nil
}
