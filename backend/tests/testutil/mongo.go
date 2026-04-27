package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	mongoopts "go.mongodb.org/mongo-driver/mongo/options"
)

// TestMongo holds a MongoDB testcontainer and its client.
type TestMongo struct {
	Client    *mongo.Client
	URI       string
	container *mongodb.MongoDBContainer
	ctx       context.Context
}

// NewTestMongo starts a MongoDB testcontainer and returns a connected TestMongo.
func NewTestMongo(ctx context.Context) (*TestMongo, error) {
	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		return nil, fmt.Errorf("start mongodb container: %w", err)
	}

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		_ = container.Terminate(ctx) // #nosec G104
		return nil, fmt.Errorf("get mongodb connection string: %w", err)
	}

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, mongoopts.Client().ApplyURI(uri))
	if err != nil {
		_ = container.Terminate(ctx) // #nosec G104
		return nil, fmt.Errorf("connect to mongodb: %w", err)
	}

	return &TestMongo{
		Client:    client,
		URI:       uri,
		container: container,
		ctx:       ctx,
	}, nil
}

// DropCollections drops all collections in the given database.
func (m *TestMongo) DropCollections(dbName string) error {
	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	db := m.Client.Database(dbName)
	names, err := db.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("list collections: %w", err)
	}
	for _, name := range names {
		if err := db.Collection(name).Drop(ctx); err != nil {
			return fmt.Errorf("drop collection %s: %w", name, err)
		}
	}
	return nil
}

// Close disconnects the client and terminates the container.
func (m *TestMongo) Close() error {
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()
	if err := m.Client.Disconnect(ctx); err != nil {
		return err
	}
	return m.container.Terminate(m.ctx)
}
