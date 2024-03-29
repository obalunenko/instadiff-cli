// Package db implements database interactions.
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/obalunenko/instadiff-cli/internal/models"
	"github.com/obalunenko/instadiff-cli/internal/utils"
)

// BuildCollectionName constructs collection name.
func BuildCollectionName(s string) string {
	const (
		sep = "_"
		pfx = "statistics"
	)

	return s + sep + pfx
}

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

// Close closes connections.
func (m *mongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func newMongoDB(ctx context.Context, params MongoParams) (*mongoDB, error) {
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(params.URL))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err = cl.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	database := cl.Database(params.Database)
	collection := database.Collection(params.Collection)

	return &mongoDB{
		client:     cl,
		database:   database,
		collection: collection,
	}, nil
}

func (m *mongoDB) InsertUsersBatch(ctx context.Context, users models.UsersBatch) error {
	if _, err := m.collection.InsertOne(ctx, users); err != nil {
		return fmt.Errorf("insert batch: %w", err)
	}

	return nil
}

func (m *mongoDB) GetLastUsersBatchByType(ctx context.Context, bt models.UsersBatchType) (models.UsersBatch, error) {
	filter := bson.M{"batch_type": bt}

	resp := m.collection.FindOne(ctx, filter, &options.FindOneOptions{
		Sort: bson.M{"$natural": -1},
	})

	if err := resp.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.MakeUsersBatch(bt, nil, time.Now()), ErrNoData
		}

		return models.MakeUsersBatch(bt, nil, time.Now()), fmt.Errorf("find batch [%s]: %w", bt.String(), err)
	}

	var ub models.UsersBatch

	if err := resp.Decode(&ub); err != nil {
		return models.MakeUsersBatch(bt, nil, time.Now()), fmt.Errorf("decode response: %w", err)
	}

	return ub, nil
}

func (m *mongoDB) GetAllUsersBatchByType(ctx context.Context, bt models.UsersBatchType) ([]models.UsersBatch, error) {
	filter := bson.M{"batch_type": bt}

	resp, err := m.collection.Find(ctx, filter, &options.FindOptions{
		Sort: bson.M{"$natural": -1},
	})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNoData
		}

		return nil, fmt.Errorf("find batches [%s]: %w", bt.String(), err)
	}

	defer func() {
		utils.LogError(ctx, resp.Close(ctx), "mongo: Failed to close cursor")
	}()

	var batches []models.UsersBatch

	for resp.Next(ctx) {
		var ub models.UsersBatch

		if err := resp.Decode(&ub); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		batches = append(batches, ub)
	}

	if err := resp.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNoData
		}
	}

	return batches, nil
}
