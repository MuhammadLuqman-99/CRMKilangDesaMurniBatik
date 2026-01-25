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

	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Unit of Work Implementation
// ============================================================================

// UnitOfWork implements domain.UnitOfWork for PostgreSQL transactions.
type UnitOfWork struct {
	db              *sqlx.DB
	tx              *sqlx.Tx
	ctx             context.Context
	leadRepo        *LeadRepository
	opportunityRepo *OpportunityRepository
	dealRepo        *DealRepository
	pipelineRepo    *PipelineRepository
	eventStore      *EventStore
	isTransaction   bool
}

// NewUnitOfWork creates a new UnitOfWork instance.
func NewUnitOfWork(db *sqlx.DB) *UnitOfWork {
	return &UnitOfWork{
		db:              db,
		leadRepo:        NewLeadRepository(db),
		opportunityRepo: NewOpportunityRepository(db),
		dealRepo:        NewDealRepository(db),
		pipelineRepo:    NewPipelineRepository(db),
		eventStore:      NewEventStore(db),
		isTransaction:   false,
	}
}

// Begin starts a new transaction and returns a new UnitOfWork with the transaction.
func (uow *UnitOfWork) Begin(ctx context.Context) (domain.UnitOfWork, error) {
	// If already in transaction, return the same UnitOfWork
	if uow.isTransaction {
		return uow, nil
	}

	tx, err := uow.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a new UnitOfWork with the transaction context
	txCtx := setTxToContext(ctx, tx)

	return &UnitOfWork{
		db:              uow.db,
		tx:              tx,
		ctx:             txCtx,
		leadRepo:        uow.leadRepo,
		opportunityRepo: uow.opportunityRepo,
		dealRepo:        uow.dealRepo,
		pipelineRepo:    uow.pipelineRepo,
		eventStore:      uow.eventStore,
		isTransaction:   true,
	}, nil
}

// Commit commits the current transaction.
func (uow *UnitOfWork) Commit() error {
	if !uow.isTransaction || uow.tx == nil {
		return fmt.Errorf("no active transaction to commit")
	}

	if err := uow.tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	uow.tx = nil
	uow.isTransaction = false
	return nil
}

// Rollback rolls back the current transaction.
func (uow *UnitOfWork) Rollback() error {
	if !uow.isTransaction || uow.tx == nil {
		return fmt.Errorf("no active transaction to rollback")
	}

	if err := uow.tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	uow.tx = nil
	uow.isTransaction = false
	return nil
}

// Leads returns the lead repository.
func (uow *UnitOfWork) Leads() domain.LeadRepository {
	return uow.leadRepo
}

// Opportunities returns the opportunity repository.
func (uow *UnitOfWork) Opportunities() domain.OpportunityRepository {
	return uow.opportunityRepo
}

// Deals returns the deal repository.
func (uow *UnitOfWork) Deals() domain.DealRepository {
	return uow.dealRepo
}

// Pipelines returns the pipeline repository.
func (uow *UnitOfWork) Pipelines() domain.PipelineRepository {
	return uow.pipelineRepo
}

// Events returns the event store.
func (uow *UnitOfWork) Events() domain.EventStore {
	return uow.eventStore
}

// Context returns the transaction context (for repositories to use).
func (uow *UnitOfWork) Context() context.Context {
	if uow.ctx != nil {
		return uow.ctx
	}
	return context.Background()
}

// ============================================================================
// Transaction Helper Methods
// ============================================================================

// WithTransaction executes a function within a transaction.
// This is a convenience method that handles Begin, Commit, and Rollback.
func (uow *UnitOfWork) WithTransaction(ctx context.Context, fn func(ctx context.Context, uow domain.UnitOfWork) error) error {
	// Start transaction
	txUow, err := uow.Begin(ctx)
	if err != nil {
		return err
	}

	// Get the transaction context
	txCtx := txUow.(*UnitOfWork).Context()

	// Handle panic
	defer func() {
		if p := recover(); p != nil {
			_ = txUow.Rollback()
			panic(p)
		}
	}()

	// Execute the function
	if err := fn(txCtx, txUow); err != nil {
		if rbErr := txUow.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return err
	}

	// Commit the transaction
	if err := txUow.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithSerializableTransaction executes a function within a serializable transaction.
// This provides the highest isolation level for critical operations.
func (uow *UnitOfWork) WithSerializableTransaction(ctx context.Context, fn func(ctx context.Context, uow domain.UnitOfWork) error) error {
	tx, err := uow.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin serializable transaction: %w", err)
	}

	txCtx := setTxToContext(ctx, tx)

	txUow := &UnitOfWork{
		db:              uow.db,
		tx:              tx,
		ctx:             txCtx,
		leadRepo:        uow.leadRepo,
		opportunityRepo: uow.opportunityRepo,
		dealRepo:        uow.dealRepo,
		pipelineRepo:    uow.pipelineRepo,
		eventStore:      uow.eventStore,
		isTransaction:   true,
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txCtx, txUow); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit serializable transaction: %w", err)
	}

	return nil
}

// ============================================================================
// Health Check Methods
// ============================================================================

// Ping checks if the database connection is alive.
func (uow *UnitOfWork) Ping(ctx context.Context) error {
	return uow.db.PingContext(ctx)
}

// Close closes the database connection.
func (uow *UnitOfWork) Close() error {
	return uow.db.Close()
}

// DB returns the underlying database connection for advanced use cases.
func (uow *UnitOfWork) DB() *sqlx.DB {
	return uow.db
}

// ============================================================================
// Event Store Implementation
// ============================================================================

// EventStore implements domain.EventStore for persisting domain events.
type EventStore struct {
	db *sqlx.DB
}

// NewEventStore creates a new EventStore instance.
func NewEventStore(db *sqlx.DB) *EventStore {
	return &EventStore{db: db}
}

// eventRow represents the database row structure for domain events.
type eventRow struct {
	ID            string    `db:"id"`
	TenantID      string    `db:"tenant_id"`
	AggregateID   string    `db:"aggregate_id"`
	AggregateType string    `db:"aggregate_type"`
	EventType     string    `db:"event_type"`
	EventData     []byte    `db:"event_data"`
	Metadata      []byte    `db:"metadata"`
	Version       int       `db:"version"`
	OccurredAt    time.Time `db:"occurred_at"`
	CreatedAt     time.Time `db:"created_at"`
}

// Save persists domain events to the database.
func (es *EventStore) Save(ctx context.Context, events ...domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	executor := getExecutor(ctx, es.db)

	query := `
		INSERT INTO sales_domain_events (
			id, tenant_id, aggregate_id, aggregate_type, event_type,
			event_data, metadata, version, occurred_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	for _, event := range events {
		// Serialize the entire event as event data
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}

		// Empty metadata - could be extended if needed
		metadata := []byte("{}")

		_, err = executor.ExecContext(ctx, query,
			event.EventID().String(),
			event.TenantID().String(),
			event.AggregateID().String(),
			event.AggregateType(),
			event.EventType(),
			eventData,
			metadata,
			event.Version(),
			event.OccurredAt(),
			time.Now().UTC(),
		)
		if err != nil {
			return fmt.Errorf("failed to save event: %w", err)
		}
	}

	return nil
}

// GetByAggregateID retrieves all events for an aggregate.
func (es *EventStore) GetByAggregateID(ctx context.Context, tenantID, aggregateID uuid.UUID) ([]domain.DomainEvent, error) {
	executor := getExecutor(ctx, es.db)

	query := `
		SELECT id, tenant_id, aggregate_id, aggregate_type, event_type,
			event_data, metadata, version, occurred_at, created_at
		FROM sales_domain_events
		WHERE tenant_id = $1 AND aggregate_id = $2
		ORDER BY version ASC, occurred_at ASC`

	var rows []eventRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, tenantID.String(), aggregateID.String()); err != nil {
		return nil, fmt.Errorf("failed to get events by aggregate ID: %w", err)
	}

	return es.toDomainEvents(rows)
}

// GetByAggregateType retrieves events by aggregate type.
func (es *EventStore) GetByAggregateType(ctx context.Context, tenantID uuid.UUID, aggregateType string, opts domain.ListOptions) ([]domain.DomainEvent, error) {
	executor := getExecutor(ctx, es.db)

	query := `
		SELECT id, tenant_id, aggregate_id, aggregate_type, event_type,
			event_data, metadata, version, occurred_at, created_at
		FROM sales_domain_events
		WHERE tenant_id = $1 AND aggregate_type = $2
		ORDER BY occurred_at DESC
		LIMIT $3 OFFSET $4`

	var rows []eventRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query,
		tenantID.String(), aggregateType, opts.Limit(), opts.Offset()); err != nil {
		return nil, fmt.Errorf("failed to get events by aggregate type: %w", err)
	}

	return es.toDomainEvents(rows)
}

// GetByEventType retrieves events by event type.
func (es *EventStore) GetByEventType(ctx context.Context, tenantID uuid.UUID, eventType string, opts domain.ListOptions) ([]domain.DomainEvent, error) {
	executor := getExecutor(ctx, es.db)

	query := `
		SELECT id, tenant_id, aggregate_id, aggregate_type, event_type,
			event_data, metadata, version, occurred_at, created_at
		FROM sales_domain_events
		WHERE tenant_id = $1 AND event_type = $2
		ORDER BY occurred_at DESC
		LIMIT $3 OFFSET $4`

	var rows []eventRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query,
		tenantID.String(), eventType, opts.Limit(), opts.Offset()); err != nil {
		return nil, fmt.Errorf("failed to get events by event type: %w", err)
	}

	return es.toDomainEvents(rows)
}

// GetEventsSince retrieves events that occurred after a specific time.
func (es *EventStore) GetEventsSince(ctx context.Context, tenantID uuid.UUID, since time.Time, opts domain.ListOptions) ([]domain.DomainEvent, error) {
	executor := getExecutor(ctx, es.db)

	query := `
		SELECT id, tenant_id, aggregate_id, aggregate_type, event_type,
			event_data, metadata, version, occurred_at, created_at
		FROM sales_domain_events
		WHERE tenant_id = $1 AND occurred_at > $2
		ORDER BY occurred_at ASC
		LIMIT $3 OFFSET $4`

	var rows []eventRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query,
		tenantID.String(), since, opts.Limit(), opts.Offset()); err != nil {
		return nil, fmt.Errorf("failed to get events since: %w", err)
	}

	return es.toDomainEvents(rows)
}

// GetEventsInRange retrieves events within a time range.
func (es *EventStore) GetEventsInRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]domain.DomainEvent, error) {
	executor := getExecutor(ctx, es.db)

	query := `
		SELECT id, tenant_id, aggregate_id, aggregate_type, event_type,
			event_data, metadata, version, occurred_at, created_at
		FROM sales_domain_events
		WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at <= $3
		ORDER BY occurred_at ASC
		LIMIT $4 OFFSET $5`

	var rows []eventRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query,
		tenantID.String(), start, end, opts.Limit(), opts.Offset()); err != nil {
		return nil, fmt.Errorf("failed to get events in range: %w", err)
	}

	return es.toDomainEvents(rows)
}

// toDomainEvents converts database rows to domain events.
func (es *EventStore) toDomainEvents(rows []eventRow) ([]domain.DomainEvent, error) {
	events := make([]domain.DomainEvent, len(rows))
	for i, row := range rows {
		id, _ := uuid.Parse(row.ID)
		tenantID, _ := uuid.Parse(row.TenantID)
		aggregateID, _ := uuid.Parse(row.AggregateID)

		var data map[string]interface{}
		if err := json.Unmarshal(row.EventData, &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}

		var metadata map[string]string
		if len(row.Metadata) > 0 && string(row.Metadata) != "null" {
			if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event metadata: %w", err)
			}
		}

		events[i] = &storedEvent{
			id:            id,
			tenantID:      tenantID,
			aggregateID:   aggregateID,
			aggregateType: row.AggregateType,
			eventType:     row.EventType,
			data:          data,
			metadata:      metadata,
			version:       row.Version,
			occurredAt:    row.OccurredAt,
		}
	}
	return events, nil
}

// storedEvent represents a persisted domain event.
type storedEvent struct {
	id            uuid.UUID
	tenantID      uuid.UUID
	aggregateID   uuid.UUID
	aggregateType string
	eventType     string
	data          map[string]interface{}
	metadata      map[string]string
	version       int
	occurredAt    time.Time
}

func (e *storedEvent) EventID() uuid.UUID      { return e.id }
func (e *storedEvent) TenantID() uuid.UUID     { return e.tenantID }
func (e *storedEvent) AggregateID() uuid.UUID  { return e.aggregateID }
func (e *storedEvent) AggregateType() string   { return e.aggregateType }
func (e *storedEvent) EventType() string       { return e.eventType }
func (e *storedEvent) Version() int            { return e.version }
func (e *storedEvent) OccurredAt() time.Time   { return e.occurredAt }

// Data returns the event data (for internal use, not part of DomainEvent interface).
func (e *storedEvent) Data() map[string]interface{} { return e.data }

// Metadata returns the event metadata (for internal use, not part of DomainEvent interface).
func (e *storedEvent) Metadata() map[string]string { return e.metadata }
