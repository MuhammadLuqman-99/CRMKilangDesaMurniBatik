// Package mongodb provides MongoDB implementations for customer repositories.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

const (
	outboxCollection = "customer_outbox"
	maxRetries       = 5
)

// OutboxRepository implements domain.OutboxRepository using MongoDB.
type OutboxRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewOutboxRepository creates a new OutboxRepository.
func NewOutboxRepository(db *mongo.Database) *OutboxRepository {
	return &OutboxRepository{
		db:         db,
		collection: db.Collection(outboxCollection),
	}
}

// Create creates an outbox entry.
func (r *OutboxRepository) Create(ctx context.Context, entry *domain.OutboxEntry) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	_, err := r.collection.InsertOne(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to create outbox entry: %w", err)
	}

	return nil
}

// MarkAsProcessed marks an entry as processed.
func (r *OutboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"processed_at": now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to mark entry as processed: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("outbox entry not found")
	}

	return nil
}

// MarkAsFailed marks an entry as failed.
func (r *OutboxRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"failed_at": now,
			"error":     errMsg,
		},
		"$inc": bson.M{"retry_count": 1},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to mark entry as failed: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("outbox entry not found")
	}

	return nil
}

// FindPending finds pending outbox entries.
func (r *OutboxRepository) FindPending(ctx context.Context, limit int) ([]*domain.OutboxEntry, error) {
	filter := bson.M{
		"processed_at": nil,
		"$or": []bson.M{
			{"failed_at": nil},
			{"retry_count": bson.M{"$lt": maxRetries}},
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*domain.OutboxEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// FindFailed finds failed outbox entries.
func (r *OutboxRepository) FindFailed(ctx context.Context, limit int) ([]*domain.OutboxEntry, error) {
	filter := bson.M{
		"processed_at": nil,
		"failed_at":    bson.M{"$ne": nil},
		"retry_count":  bson.M{"$lt": maxRetries},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "failed_at", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find failed entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*domain.OutboxEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// DeleteOld deletes old processed entries.
func (r *OutboxRepository) DeleteOld(ctx context.Context, before time.Time) error {
	filter := bson.M{
		"processed_at": bson.M{"$lt": before},
	}

	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete old entries: %w", err)
	}

	return nil
}

// BulkCreate creates multiple outbox entries.
func (r *OutboxRepository) BulkCreate(ctx context.Context, entries []*domain.OutboxEntry) error {
	if len(entries) == 0 {
		return nil
	}

	now := time.Now().UTC()
	docs := make([]interface{}, len(entries))
	for i, entry := range entries {
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = now
		}
		docs[i] = entry
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to bulk create outbox entries: %w", err)
	}

	return nil
}

// GetPendingCount returns the count of pending entries.
func (r *OutboxRepository) GetPendingCount(ctx context.Context) (int64, error) {
	filter := bson.M{
		"processed_at": nil,
		"$or": []bson.M{
			{"failed_at": nil},
			{"retry_count": bson.M{"$lt": maxRetries}},
		},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending entries: %w", err)
	}

	return count, nil
}

// GetFailedCount returns the count of failed entries.
func (r *OutboxRepository) GetFailedCount(ctx context.Context) (int64, error) {
	filter := bson.M{
		"processed_at": nil,
		"failed_at":    bson.M{"$ne": nil},
		"retry_count":  bson.M{"$gte": maxRetries},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count failed entries: %w", err)
	}

	return count, nil
}

// ResetFailed resets failed entries for retry.
func (r *OutboxRepository) ResetFailed(ctx context.Context, id uuid.UUID) error {
	update := bson.M{
		"$set": bson.M{
			"failed_at": nil,
			"error":     "",
		},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to reset failed entry: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("outbox entry not found")
	}

	return nil
}

// FindByEventType finds entries by event type.
func (r *OutboxRepository) FindByEventType(ctx context.Context, eventType string, limit int) ([]*domain.OutboxEntry, error) {
	filter := bson.M{
		"event_type": eventType,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries by event type: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*domain.OutboxEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// FindByAggregateID finds entries by aggregate ID.
func (r *OutboxRepository) FindByAggregateID(ctx context.Context, aggregateID uuid.UUID, limit int) ([]*domain.OutboxEntry, error) {
	filter := bson.M{
		"aggregate_id": aggregateID,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find entries by aggregate ID: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*domain.OutboxEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// LockEntry attempts to lock an entry for processing.
func (r *OutboxRepository) LockEntry(ctx context.Context, id uuid.UUID, workerID string, lockDuration time.Duration) (bool, error) {
	now := time.Now().UTC()
	lockUntil := now.Add(lockDuration)

	filter := bson.M{
		"_id":          id,
		"processed_at": nil,
		"$or": []bson.M{
			{"locked_until": nil},
			{"locked_until": bson.M{"$lt": now}},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"locked_by":    workerID,
			"locked_until": lockUntil,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, fmt.Errorf("failed to lock entry: %w", err)
	}

	return result.ModifiedCount > 0, nil
}

// UnlockEntry unlocks an entry.
func (r *OutboxRepository) UnlockEntry(ctx context.Context, id uuid.UUID) error {
	update := bson.M{
		"$set": bson.M{
			"locked_by":    nil,
			"locked_until": nil,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to unlock entry: %w", err)
	}

	return nil
}

// FindStaleLockedEntries finds entries with stale locks.
func (r *OutboxRepository) FindStaleLockedEntries(ctx context.Context) ([]*domain.OutboxEntry, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"processed_at": nil,
		"locked_until": bson.M{"$lt": now, "$ne": nil},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find stale locked entries: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*domain.OutboxEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode entries: %w", err)
	}

	return entries, nil
}

// ResetStaleLocks resets stale locks on entries.
func (r *OutboxRepository) ResetStaleLocks(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	filter := bson.M{
		"processed_at": nil,
		"locked_until": bson.M{"$lt": now, "$ne": nil},
	}

	update := bson.M{
		"$set": bson.M{
			"locked_by":    nil,
			"locked_until": nil,
		},
	}

	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to reset stale locks: %w", err)
	}

	return result.ModifiedCount, nil
}
