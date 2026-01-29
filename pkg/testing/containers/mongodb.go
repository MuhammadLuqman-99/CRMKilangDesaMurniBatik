// Package containers provides test container implementations for integration testing.
package containers

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDBContainer represents a MongoDB test container configuration.
type MongoDBContainer struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	Client   *mongo.Client
	DB       *mongo.Database
}

// MongoDBContainerConfig holds configuration for MongoDB container.
type MongoDBContainerConfig struct {
	Database string
	User     string
	Password string
}

// DefaultMongoDBConfig returns default MongoDB configuration.
func DefaultMongoDBConfig() MongoDBContainerConfig {
	return MongoDBContainerConfig{
		Database: "crm_customers_test",
		User:     "crm_test",
		Password: "crm_test_password",
	}
}

// NewMongoDBContainer creates a new MongoDB container for testing.
// For integration tests, this connects to the docker-compose MongoDB instance.
func NewMongoDBContainer(ctx context.Context, cfg MongoDBContainerConfig) (*MongoDBContainer, error) {
	container := &MongoDBContainer{
		Host:     getEnvOrDefault("TEST_MONGODB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_MONGODB_PORT", "27017"),
		Database: getEnvOrDefault("TEST_MONGODB_DB", cfg.Database),
		User:     getEnvOrDefault("TEST_MONGODB_USER", cfg.User),
		Password: getEnvOrDefault("TEST_MONGODB_PASSWORD", cfg.Password),
	}

	// Build connection URI
	uri := container.ConnectionURI()

	// Create client options
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(10).
		SetMinPoolSize(2).
		SetMaxConnIdleTime(10 * time.Minute).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(10 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	container.Client = client
	container.DB = client.Database(container.Database)

	return container, nil
}

// ConnectionURI returns the MongoDB connection URI.
func (c *MongoDBContainer) ConnectionURI() string {
	if c.User != "" && c.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin",
			c.User, c.Password, c.Host, c.Port, c.Database)
	}
	return fmt.Sprintf("mongodb://%s:%s/%s", c.Host, c.Port, c.Database)
}

// GetClient returns the MongoDB client.
func (c *MongoDBContainer) GetClient() *mongo.Client {
	return c.Client
}

// GetDB returns the MongoDB database.
func (c *MongoDBContainer) GetDB() *mongo.Database {
	return c.DB
}

// Collection returns a collection from the database.
func (c *MongoDBContainer) Collection(name string) *mongo.Collection {
	return c.DB.Collection(name)
}

// CreateIndexes creates indexes on a collection.
func (c *MongoDBContainer) CreateIndexes(ctx context.Context, collection string, indexes []mongo.IndexModel) error {
	col := c.DB.Collection(collection)
	_, err := col.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	return nil
}

// CreateTextIndex creates a text index on specified fields.
func (c *MongoDBContainer) CreateTextIndex(ctx context.Context, collection string, fields []string) error {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: "text"})
	}

	index := mongo.IndexModel{
		Keys: keys,
		Options: options.Index().
			SetBackground(true).
			SetName("text_search_index"),
	}

	col := c.DB.Collection(collection)
	_, err := col.Indexes().CreateOne(ctx, index)
	if err != nil {
		return fmt.Errorf("failed to create text index: %w", err)
	}
	return nil
}

// DropCollection drops a collection.
func (c *MongoDBContainer) DropCollection(ctx context.Context, collection string) error {
	return c.DB.Collection(collection).Drop(ctx)
}

// DropAllCollections drops all collections in the database.
func (c *MongoDBContainer) DropAllCollections(ctx context.Context) error {
	collections, err := c.DB.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	for _, col := range collections {
		if err := c.DropCollection(ctx, col); err != nil {
			return fmt.Errorf("failed to drop collection %s: %w", col, err)
		}
	}

	return nil
}

// CleanCollection removes all documents from a collection.
func (c *MongoDBContainer) CleanCollection(ctx context.Context, collection string) error {
	_, err := c.DB.Collection(collection).DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to clean collection %s: %w", collection, err)
	}
	return nil
}

// CleanAllCollections removes all documents from all collections.
func (c *MongoDBContainer) CleanAllCollections(ctx context.Context) error {
	collections, err := c.DB.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	for _, col := range collections {
		if err := c.CleanCollection(ctx, col); err != nil {
			return err
		}
	}

	return nil
}

// InsertOne inserts a single document into a collection.
func (c *MongoDBContainer) InsertOne(ctx context.Context, collection string, document interface{}) error {
	_, err := c.DB.Collection(collection).InsertOne(ctx, document)
	return err
}

// InsertMany inserts multiple documents into a collection.
func (c *MongoDBContainer) InsertMany(ctx context.Context, collection string, documents []interface{}) error {
	_, err := c.DB.Collection(collection).InsertMany(ctx, documents)
	return err
}

// FindOne finds a single document from a collection.
func (c *MongoDBContainer) FindOne(ctx context.Context, collection string, filter interface{}, result interface{}) error {
	return c.DB.Collection(collection).FindOne(ctx, filter).Decode(result)
}

// Find finds documents from a collection.
func (c *MongoDBContainer) Find(ctx context.Context, collection string, filter interface{}, results interface{}) error {
	cursor, err := c.DB.Collection(collection).Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, results)
}

// CountDocuments counts documents in a collection.
func (c *MongoDBContainer) CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error) {
	return c.DB.Collection(collection).CountDocuments(ctx, filter)
}

// ExecuteInSession executes a function within a MongoDB session/transaction.
func (c *MongoDBContainer) ExecuteInSession(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := c.Client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	return err
}

// Close closes the MongoDB client connection.
func (c *MongoDBContainer) Close(ctx context.Context) error {
	if c.Client != nil {
		return c.Client.Disconnect(ctx)
	}
	return nil
}

// WaitForReady waits for MongoDB to be ready.
func (c *MongoDBContainer) WaitForReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for MongoDB to be ready")
		case <-ticker.C:
			if err := c.Client.Ping(ctx, readpref.Primary()); err == nil {
				return nil
			}
		}
	}
}

// SetupCustomerCollections sets up all necessary collections and indexes for the Customer service.
func (c *MongoDBContainer) SetupCustomerCollections(ctx context.Context) error {
	// Create customers collection with indexes
	customersIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "email.address", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "owner_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "tags", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "segments", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "deleted_at", Value: 1}},
		},
	}

	if err := c.CreateIndexes(ctx, "customers", customersIndexes); err != nil {
		return fmt.Errorf("failed to create customers indexes: %w", err)
	}

	// Create text index for search
	if err := c.CreateTextIndex(ctx, "customers", []string{"name", "email.address", "code", "notes"}); err != nil {
		// Ignore if index already exists
		if !isIndexExistsError(err) {
			return fmt.Errorf("failed to create text index: %w", err)
		}
	}

	// Create contacts collection with indexes
	contactsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "customer_id", Value: 1}},
		},
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "email.address", Value: 1}},
			Options: options.Index().SetSparse(true),
		},
		{
			Keys: bson.D{{Key: "is_primary", Value: 1}},
		},
	}

	if err := c.CreateIndexes(ctx, "contacts", contactsIndexes); err != nil {
		return fmt.Errorf("failed to create contacts indexes: %w", err)
	}

	// Create segments collection with indexes
	segmentsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}, {Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
	}

	if err := c.CreateIndexes(ctx, "segments", segmentsIndexes); err != nil {
		return fmt.Errorf("failed to create segments indexes: %w", err)
	}

	// Create activities collection with indexes
	activitiesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "customer_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "type", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
	}

	if err := c.CreateIndexes(ctx, "activities", activitiesIndexes); err != nil {
		return fmt.Errorf("failed to create activities indexes: %w", err)
	}

	// Create notes collection with indexes
	notesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "customer_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_pinned", Value: -1}, {Key: "created_at", Value: -1}},
		},
	}

	if err := c.CreateIndexes(ctx, "notes", notesIndexes); err != nil {
		return fmt.Errorf("failed to create notes indexes: %w", err)
	}

	// Create outbox collection with indexes
	outboxIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "published", Value: 1}, {Key: "created_at", Value: 1}},
		},
	}

	if err := c.CreateIndexes(ctx, "outbox", outboxIndexes); err != nil {
		return fmt.Errorf("failed to create outbox indexes: %w", err)
	}

	return nil
}

func isIndexExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "already exists") ||
		contains(errStr, "IndexOptionsConflict") ||
		contains(errStr, "index with name")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
