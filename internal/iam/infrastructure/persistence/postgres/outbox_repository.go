// Package postgres contains PostgreSQL repository implementations.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// OutboxRow represents an outbox database row.
type OutboxRow struct {
	ID            uuid.UUID  `db:"id"`
	EventType     string     `db:"event_type"`
	AggregateID   uuid.UUID  `db:"aggregate_id"`
	AggregateType string     `db:"aggregate_type"`
	Payload       []byte     `db:"payload"`
	Published     bool       `db:"published"`
	PublishedAt   *time.Time `db:"published_at"`
	CreatedAt     time.Time  `db:"created_at"`
}

// ToEntry converts an OutboxRow to an OutboxEntry.
func (r *OutboxRow) ToEntry() *domain.OutboxEntry {
	return &domain.OutboxEntry{
		ID:            r.ID,
		EventType:     r.EventType,
		AggregateID:   r.AggregateID,
		AggregateType: r.AggregateType,
		Payload:       r.Payload,
		Published:     r.Published,
		PublishedAt:   r.PublishedAt,
		CreatedAt:     r.CreatedAt,
	}
}

// OutboxRepository implements domain.OutboxRepository using PostgreSQL.
type OutboxRepository struct {
	db *sqlx.DB
}

// NewOutboxRepository creates a new OutboxRepository.
func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// Create creates a new outbox entry.
func (r *OutboxRepository) Create(ctx context.Context, entry *domain.OutboxEntry) error {
	// Ensure ID is set
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	// Ensure CreatedAt is set
	createdAt, ok := entry.CreatedAt.(time.Time)
	if !ok || createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	query := `
		INSERT INTO outbox (id, event_type, aggregate_id, aggregate_type, payload, published, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.getDB(ctx).ExecContext(ctx, query,
		entry.ID,
		entry.EventType,
		entry.AggregateID,
		entry.AggregateType,
		entry.Payload,
		false,
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create outbox entry: %w", err)
	}

	return nil
}

// MarkAsPublished marks an outbox entry as published.
func (r *OutboxRepository) MarkAsPublished(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE outbox SET
			published = true,
			published_at = $1
		WHERE id = $2 AND published = false`

	result, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox entry as published: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// FindUnpublished finds all unpublished outbox entries.
func (r *OutboxRepository) FindUnpublished(ctx context.Context, limit int) ([]*domain.OutboxEntry, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, payload, published, published_at, created_at
		FROM outbox
		WHERE published = false
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	var rows []OutboxRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find unpublished outbox entries: %w", err)
	}

	entries := make([]*domain.OutboxEntry, len(rows))
	for i, row := range rows {
		entries[i] = row.ToEntry()
	}

	return entries, nil
}

// DeletePublished deletes published outbox entries older than the specified time.
func (r *OutboxRepository) DeletePublished(ctx context.Context, olderThan interface{}) (int64, error) {
	var cutoff time.Time
	switch v := olderThan.(type) {
	case time.Time:
		cutoff = v
	case *time.Time:
		if v != nil {
			cutoff = *v
		}
	default:
		return 0, fmt.Errorf("invalid olderThan type: expected time.Time")
	}

	if cutoff.IsZero() {
		cutoff = time.Now().UTC().Add(-24 * time.Hour) // Default: 24 hours ago
	}

	query := `DELETE FROM outbox WHERE published = true AND published_at < $1`

	result, err := r.getDB(ctx).ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete published outbox entries: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// FindByID finds an outbox entry by ID.
func (r *OutboxRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.OutboxEntry, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, payload, published, published_at, created_at
		FROM outbox
		WHERE id = $1`

	var row OutboxRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to find outbox entry: %w", err)
	}

	return row.ToEntry(), nil
}

// FindByAggregateID finds outbox entries by aggregate ID.
func (r *OutboxRepository) FindByAggregateID(ctx context.Context, aggregateID uuid.UUID) ([]*domain.OutboxEntry, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, payload, published, published_at, created_at
		FROM outbox
		WHERE aggregate_id = $1
		ORDER BY created_at ASC`

	var rows []OutboxRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find outbox entries by aggregate: %w", err)
	}

	entries := make([]*domain.OutboxEntry, len(rows))
	for i, row := range rows {
		entries[i] = row.ToEntry()
	}

	return entries, nil
}

// CountUnpublished counts unpublished outbox entries.
func (r *OutboxRepository) CountUnpublished(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM outbox WHERE published = false`

	var count int64
	err := r.getDB(ctx).GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count unpublished outbox entries: %w", err)
	}

	return count, nil
}

// MarkMultipleAsPublished marks multiple outbox entries as published.
func (r *OutboxRepository) MarkMultipleAsPublished(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE outbox SET
			published = true,
			published_at = $1
		WHERE id = ANY($2) AND published = false`

	_, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC(), ids)
	if err != nil {
		return fmt.Errorf("failed to mark outbox entries as published: %w", err)
	}

	return nil
}

// RequeueStuckEntries requeues entries that have been stuck for too long.
func (r *OutboxRepository) RequeueStuckEntries(ctx context.Context, stuckSince time.Time) (int64, error) {
	// This is useful for retry scenarios where entries might get stuck due to failures
	// Entries are considered stuck if they're unpublished and older than stuckSince
	query := `
		UPDATE outbox SET
			created_at = $1
		WHERE published = false AND created_at < $2`

	result, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC(), stuckSince)
	if err != nil {
		return 0, fmt.Errorf("failed to requeue stuck outbox entries: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// getDB returns the database connection, checking for transaction in context.
func (r *OutboxRepository) getDB(ctx context.Context) sqlx.ExtContext {
	if tx := getTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}
