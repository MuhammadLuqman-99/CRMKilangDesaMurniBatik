// Package database provides database connection utilities for the CRM application.
package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/logger"
)

// MongoDB wraps the mongo.Client and provides database operations.
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	config   *config.MongoDBConfig
	log      *logger.Logger
}

// NewMongoDB creates a new MongoDB connection.
func NewMongoDB(cfg *config.MongoDBConfig, log *logger.Logger) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	// Configure client options
	clientOpts := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetServerSelectionTimeout(cfg.ServerTimeout)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Info().
		Str("database", cfg.Database).
		Msg("Connected to MongoDB")

	return &MongoDB{
		client:   client,
		database: client.Database(cfg.Database),
		config:   cfg,
		log:      log,
	}, nil
}

// Close closes the MongoDB connection.
func (m *MongoDB) Close(ctx context.Context) error {
	m.log.Info().Msg("Closing MongoDB connection")
	return m.client.Disconnect(ctx)
}

// Health checks the MongoDB connection health.
func (m *MongoDB) Health(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}

// Database returns the MongoDB database.
func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

// Client returns the MongoDB client.
func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

// Collection returns a collection from the database.
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

// Transaction executes a function within a MongoDB transaction.
func (m *MongoDB) Transaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// WithSession executes a function within a MongoDB session.
func (m *MongoDB) WithSession(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	return m.client.UseSession(ctx, fn)
}

// CreateIndexes creates indexes for a collection.
func (m *MongoDB) CreateIndexes(ctx context.Context, collection string, indexes []mongo.IndexModel) error {
	coll := m.Collection(collection)

	opts := options.CreateIndexes().SetMaxTime(30 * time.Second)
	_, err := coll.Indexes().CreateMany(ctx, indexes, opts)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	m.log.Info().
		Str("collection", collection).
		Int("count", len(indexes)).
		Msg("Created indexes")

	return nil
}

// DropCollection drops a collection from the database.
func (m *MongoDB) DropCollection(ctx context.Context, collection string) error {
	return m.Collection(collection).Drop(ctx)
}

// ListCollections lists all collections in the database.
func (m *MongoDB) ListCollections(ctx context.Context) ([]string, error) {
	return m.database.ListCollectionNames(ctx, map[string]interface{}{})
}
