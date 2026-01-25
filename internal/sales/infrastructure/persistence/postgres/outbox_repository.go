// Package postgres provides PostgreSQL implementations for sales domain repositories.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ============================================================================
// Outbox Entry Types
// ============================================================================

// OutboxEntry represents a transactional outbox entry for reliable event publishing.
type OutboxEntry struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	EventType     string     `json:"event_type" db:"event_type"`
	AggregateID   uuid.UUID  `json:"aggregate_id" db:"aggregate_id"`
	AggregateType string     `json:"aggregate_type" db:"aggregate_type"`
	Payload       []byte     `json:"payload" db:"payload"`
	Metadata      []byte     `json:"metadata" db:"metadata"`
	Published     bool       `json:"published" db:"published"`
	PublishedAt   *time.Time `json:"published_at,omitempty" db:"published_at"`
	RetryCount    int        `json:"retry_count" db:"retry_count"`
	LastError     *string    `json:"last_error,omitempty" db:"last_error"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// NewOutboxEntry creates a new outbox entry.
func NewOutboxEntry(tenantID uuid.UUID, eventType string, aggregateID uuid.UUID, aggregateType string, payload interface{}, metadata map[string]string) (*OutboxEntry, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	now := time.Now().UTC()
	return &OutboxEntry{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EventType:     eventType,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		Payload:       payloadBytes,
		Metadata:      metadataBytes,
		Published:     false,
		RetryCount:    0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// ============================================================================
// Outbox Repository Implementation
// ============================================================================

// OutboxRepository implements transactional outbox pattern for reliable event publishing.
type OutboxRepository struct {
	db *sqlx.DB
}

// NewOutboxRepository creates a new OutboxRepository instance.
func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// outboxRow represents the database row structure for outbox entries.
type outboxRow struct {
	ID            uuid.UUID      `db:"id"`
	TenantID      uuid.UUID      `db:"tenant_id"`
	EventType     string         `db:"event_type"`
	AggregateID   uuid.UUID      `db:"aggregate_id"`
	AggregateType string         `db:"aggregate_type"`
	Payload       []byte         `db:"payload"`
	Metadata      []byte         `db:"metadata"`
	Published     bool           `db:"published"`
	PublishedAt   sql.NullTime   `db:"published_at"`
	RetryCount    int            `db:"retry_count"`
	LastError     sql.NullString `db:"last_error"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

// Create creates a new outbox entry.
func (r *OutboxRepository) Create(ctx context.Context, entry *OutboxEntry) error {
	executor := getExecutor(ctx, r.db)

	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	now := time.Now().UTC()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now

	query := `
		INSERT INTO sales_outbox (
			id, tenant_id, event_type, aggregate_id, aggregate_type,
			payload, metadata, published, retry_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	_, err := executor.ExecContext(ctx, query,
		entry.ID,
		entry.TenantID,
		entry.EventType,
		entry.AggregateID,
		entry.AggregateType,
		entry.Payload,
		entry.Metadata,
		entry.Published,
		entry.RetryCount,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create outbox entry: %w", err)
	}

	return nil
}

// CreateBatch creates multiple outbox entries in a batch.
func (r *OutboxRepository) CreateBatch(ctx context.Context, entries []*OutboxEntry) error {
	if len(entries) == 0 {
		return nil
	}

	executor := getExecutor(ctx, r.db)
	now := time.Now().UTC()

	query := `
		INSERT INTO sales_outbox (
			id, tenant_id, event_type, aggregate_id, aggregate_type,
			payload, metadata, published, retry_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	for _, entry := range entries {
		if entry.ID == uuid.Nil {
			entry.ID = uuid.New()
		}
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = now
		}
		entry.UpdatedAt = now

		_, err := executor.ExecContext(ctx, query,
			entry.ID,
			entry.TenantID,
			entry.EventType,
			entry.AggregateID,
			entry.AggregateType,
			entry.Payload,
			entry.Metadata,
			entry.Published,
			entry.RetryCount,
			entry.CreatedAt,
			entry.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create outbox entry %s: %w", entry.ID, err)
		}
	}

	return nil
}

// GetByID retrieves an outbox entry by ID.
func (r *OutboxRepository) GetByID(ctx context.Context, id uuid.UUID) (*OutboxEntry, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, event_type, aggregate_id, aggregate_type,
			payload, metadata, published, published_at, retry_count,
			last_error, created_at, updated_at
		FROM sales_outbox
		WHERE id = $1`

	var row outboxRow
	err := sqlx.GetContext(ctx, executor, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get outbox entry: %w", err)
	}

	return r.toOutboxEntry(&row), nil
}

// FindUnpublished finds unpublished outbox entries ready for processing.
func (r *OutboxRepository) FindUnpublished(ctx context.Context, limit int) ([]*OutboxEntry, error) {
	executor := getExecutor(ctx, r.db)

	if limit <= 0 {
		limit = 100
	}

	// Use FOR UPDATE SKIP LOCKED to prevent concurrent processing
	query := `
		SELECT id, tenant_id, event_type, aggregate_id, aggregate_type,
			payload, metadata, published, published_at, retry_count,
			last_error, created_at, updated_at
		FROM sales_outbox
		WHERE published = false
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	var rows []outboxRow
	err := sqlx.SelectContext(ctx, executor, &rows, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find unpublished outbox entries: %w", err)
	}

	entries := make([]*OutboxEntry, len(rows))
	for i, row := range rows {
		entries[i] = r.toOutboxEntry(&row)
	}

	return entries, nil
}

// FindUnpublishedByTenant finds unpublished entries for a specific tenant.
func (r *OutboxRepository) FindUnpublishedByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*OutboxEntry, error) {
	executor := getExecutor(ctx, r.db)

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, tenant_id, event_type, aggregate_id, aggregate_type,
			payload, metadata, published, published_at, retry_count,
			last_error, created_at, updated_at
		FROM sales_outbox
		WHERE tenant_id = $1 AND published = false
		ORDER BY created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED`

	var rows []outboxRow
	err := sqlx.SelectContext(ctx, executor, &rows, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find unpublished outbox entries: %w", err)
	}

	entries := make([]*OutboxEntry, len(rows))
	for i, row := range rows {
		entries[i] = r.toOutboxEntry(&row)
	}

	return entries, nil
}

// FindByAggregateID finds outbox entries by aggregate ID.
func (r *OutboxRepository) FindByAggregateID(ctx context.Context, tenantID, aggregateID uuid.UUID) ([]*OutboxEntry, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, event_type, aggregate_id, aggregate_type,
			payload, metadata, published, published_at, retry_count,
			last_error, created_at, updated_at
		FROM sales_outbox
		WHERE tenant_id = $1 AND aggregate_id = $2
		ORDER BY created_at ASC`

	var rows []outboxRow
	err := sqlx.SelectContext(ctx, executor, &rows, query, tenantID, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find outbox entries by aggregate: %w", err)
	}

	entries := make([]*OutboxEntry, len(rows))
	for i, row := range rows {
		entries[i] = r.toOutboxEntry(&row)
	}

	return entries, nil
}

// MarkAsPublished marks an outbox entry as published.
func (r *OutboxRepository) MarkAsPublished(ctx context.Context, id uuid.UUID) error {
	executor := getExecutor(ctx, r.db)

	now := time.Now().UTC()
	query := `
		UPDATE sales_outbox SET
			published = true,
			published_at = $2,
			updated_at = $2
		WHERE id = $1 AND published = false`

	result, err := executor.ExecContext(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("failed to mark outbox entry as published: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// MarkMultipleAsPublished marks multiple outbox entries as published.
func (r *OutboxRepository) MarkMultipleAsPublished(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	executor := getExecutor(ctx, r.db)
	now := time.Now().UTC()

	query := `
		UPDATE sales_outbox SET
			published = true,
			published_at = $1,
			updated_at = $1
		WHERE id = ANY($2) AND published = false`

	_, err := executor.ExecContext(ctx, query, now, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to mark outbox entries as published: %w", err)
	}

	return nil
}

// MarkAsFailed marks an outbox entry as failed and increments retry count.
func (r *OutboxRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	executor := getExecutor(ctx, r.db)

	now := time.Now().UTC()
	query := `
		UPDATE sales_outbox SET
			retry_count = retry_count + 1,
			last_error = $2,
			updated_at = $3
		WHERE id = $1`

	result, err := executor.ExecContext(ctx, query, id, errorMsg, now)
	if err != nil {
		return fmt.Errorf("failed to mark outbox entry as failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// FindRetryable finds entries that can be retried (failed but under max retries).
func (r *OutboxRepository) FindRetryable(ctx context.Context, maxRetries int, limit int) ([]*OutboxEntry, error) {
	executor := getExecutor(ctx, r.db)

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, tenant_id, event_type, aggregate_id, aggregate_type,
			payload, metadata, published, published_at, retry_count,
			last_error, created_at, updated_at
		FROM sales_outbox
		WHERE published = false AND retry_count > 0 AND retry_count < $1
		ORDER BY updated_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED`

	var rows []outboxRow
	err := sqlx.SelectContext(ctx, executor, &rows, query, maxRetries, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find retryable outbox entries: %w", err)
	}

	entries := make([]*OutboxEntry, len(rows))
	for i, row := range rows {
		entries[i] = r.toOutboxEntry(&row)
	}

	return entries, nil
}

// DeletePublished deletes published outbox entries older than the specified time.
func (r *OutboxRepository) DeletePublished(ctx context.Context, olderThan time.Time) (int64, error) {
	executor := getExecutor(ctx, r.db)

	if olderThan.IsZero() {
		olderThan = time.Now().UTC().Add(-24 * time.Hour)
	}

	query := `DELETE FROM sales_outbox WHERE published = true AND published_at < $1`

	result, err := executor.ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete published outbox entries: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// DeleteFailed deletes failed outbox entries that exceeded max retries.
func (r *OutboxRepository) DeleteFailed(ctx context.Context, maxRetries int, olderThan time.Time) (int64, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		DELETE FROM sales_outbox
		WHERE published = false AND retry_count >= $1 AND created_at < $2`

	result, err := executor.ExecContext(ctx, query, maxRetries, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete failed outbox entries: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// CountUnpublished counts unpublished outbox entries.
func (r *OutboxRepository) CountUnpublished(ctx context.Context) (int64, error) {
	executor := getExecutor(ctx, r.db)

	query := `SELECT COUNT(*) FROM sales_outbox WHERE published = false`

	var count int64
	err := sqlx.GetContext(ctx, executor, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count unpublished outbox entries: %w", err)
	}

	return count, nil
}

// CountUnpublishedByTenant counts unpublished entries for a specific tenant.
func (r *OutboxRepository) CountUnpublishedByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	executor := getExecutor(ctx, r.db)

	query := `SELECT COUNT(*) FROM sales_outbox WHERE tenant_id = $1 AND published = false`

	var count int64
	err := sqlx.GetContext(ctx, executor, &count, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to count unpublished outbox entries: %w", err)
	}

	return count, nil
}

// CountByStatus counts outbox entries by published status.
func (r *OutboxRepository) CountByStatus(ctx context.Context) (published, unpublished, failed int64, err error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT
			COALESCE(SUM(CASE WHEN published = true THEN 1 ELSE 0 END), 0) as published,
			COALESCE(SUM(CASE WHEN published = false AND retry_count = 0 THEN 1 ELSE 0 END), 0) as unpublished,
			COALESCE(SUM(CASE WHEN published = false AND retry_count > 0 THEN 1 ELSE 0 END), 0) as failed
		FROM sales_outbox`

	type counts struct {
		Published   int64 `db:"published"`
		Unpublished int64 `db:"unpublished"`
		Failed      int64 `db:"failed"`
	}

	var c counts
	err = sqlx.GetContext(ctx, executor, &c, query)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to count outbox entries: %w", err)
	}

	return c.Published, c.Unpublished, c.Failed, nil
}

// RequeueStuckEntries requeues entries that have been stuck for too long.
func (r *OutboxRepository) RequeueStuckEntries(ctx context.Context, stuckSince time.Time) (int64, error) {
	executor := getExecutor(ctx, r.db)

	now := time.Now().UTC()
	query := `
		UPDATE sales_outbox SET
			updated_at = $1,
			retry_count = 0,
			last_error = NULL
		WHERE published = false AND updated_at < $2`

	result, err := executor.ExecContext(ctx, query, now, stuckSince)
	if err != nil {
		return 0, fmt.Errorf("failed to requeue stuck outbox entries: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// ============================================================================
// Conversion Helpers
// ============================================================================

// toOutboxEntry converts a database row to an OutboxEntry.
func (r *OutboxRepository) toOutboxEntry(row *outboxRow) *OutboxEntry {
	entry := &OutboxEntry{
		ID:            row.ID,
		TenantID:      row.TenantID,
		EventType:     row.EventType,
		AggregateID:   row.AggregateID,
		AggregateType: row.AggregateType,
		Payload:       row.Payload,
		Metadata:      row.Metadata,
		Published:     row.Published,
		RetryCount:    row.RetryCount,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}

	if row.PublishedAt.Valid {
		entry.PublishedAt = &row.PublishedAt.Time
	}

	if row.LastError.Valid {
		entry.LastError = &row.LastError.String
	}

	return entry
}
