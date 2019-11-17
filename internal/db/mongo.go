// Package db implements database interactions.
package db

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/oleg-balunenko/instadiff-cli/internal/models"
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
		return errors.Wrap(err, "failed to insert batch")
	}

	return nil
}

func (m mongoDB) GetLastUsersBatchByType(ctx context.Context,
	batchType models.UsersBatchType) (models.UsersBatch, error) {
	filter := bson.M{"batch_type": batchType}
	resp := m.collection.FindOne(ctx, filter)

	if err := resp.Err(); err != nil {
		return models.EmptyUsersBatch, errors.Wrapf(err, "failed to find batch for %s", batchType)
	}

	var ub models.UsersBatch

	if err := resp.Decode(&ub); err != nil {
		return models.EmptyUsersBatch, errors.Wrap(err, "failed tp decode response")
	}

	return ub, nil
}
