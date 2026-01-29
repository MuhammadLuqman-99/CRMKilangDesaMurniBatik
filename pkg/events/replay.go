// Package events provides event bus abstractions for the CRM application.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Event Store Interface
// ============================================================================

// EventStore defines the interface for event persistence.
type EventStore interface {
	// Save stores an event.
	Save(ctx context.Context, event *Event) error

	// SaveBatch stores multiple events.
	SaveBatch(ctx context.Context, events []*Event) error

	// GetByID retrieves an event by ID.
	GetByID(ctx context.Context, id string) (*Event, error)

	// GetByAggregateID retrieves events for an aggregate.
	GetByAggregateID(ctx context.Context, aggregateID string) ([]*Event, error)

	// GetByType retrieves events by type.
	GetByType(ctx context.Context, eventType EventType, limit int) ([]*Event, error)

	// GetByTimeRange retrieves events within a time range.
	GetByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*Event, error)

	// GetByTenant retrieves events for a tenant.
	GetByTenant(ctx context.Context, tenantID string, limit int) ([]*Event, error)

	// GetAll retrieves all events with pagination.
	GetAll(ctx context.Context, offset, limit int) ([]*Event, error)

	// Count returns the total number of events.
	Count(ctx context.Context) (int64, error)

	// CountByType returns the count of events by type.
	CountByType(ctx context.Context, eventType EventType) (int64, error)
}

// ============================================================================
// Event Replay Configuration
// ============================================================================

// ReplayConfig holds configuration for event replay.
type ReplayConfig struct {
	// BatchSize is the number of events to process in each batch
	BatchSize int

	// Concurrency is the number of concurrent handlers
	Concurrency int

	// StartTime filters events from this time (inclusive)
	StartTime *time.Time

	// EndTime filters events until this time (inclusive)
	EndTime *time.Time

	// EventTypes filters by specific event types (empty = all)
	EventTypes []EventType

	// TenantID filters by tenant (empty = all)
	TenantID string

	// AggregateID filters by aggregate (empty = all)
	AggregateID string

	// SkipErrors continues on handler errors
	SkipErrors bool

	// DryRun simulates replay without executing handlers
	DryRun bool

	// ProgressCallback is called with progress updates
	ProgressCallback func(processed, total int64, currentEvent *Event)

	// ErrorCallback is called when an error occurs
	ErrorCallback func(event *Event, err error)
}

// DefaultReplayConfig returns default replay configuration.
func DefaultReplayConfig() ReplayConfig {
	return ReplayConfig{
		BatchSize:   100,
		Concurrency: 1,
		SkipErrors:  false,
		DryRun:      false,
	}
}

// ============================================================================
// Event Replayer
// ============================================================================

// EventReplayer replays events from the event store.
type EventReplayer struct {
	store    EventStore
	handlers map[EventType][]Handler
	mu       sync.RWMutex
}

// NewEventReplayer creates a new event replayer.
func NewEventReplayer(store EventStore) *EventReplayer {
	return &EventReplayer{
		store:    store,
		handlers: make(map[EventType][]Handler),
	}
}

// RegisterHandler registers a handler for an event type.
func (r *EventReplayer) RegisterHandler(eventType EventType, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[eventType] = append(r.handlers[eventType], handler)
}

// UnregisterHandlers removes all handlers for an event type.
func (r *EventReplayer) UnregisterHandlers(eventType EventType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.handlers, eventType)
}

// ClearHandlers removes all handlers.
func (r *EventReplayer) ClearHandlers() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = make(map[EventType][]Handler)
}

// ReplayResult holds the result of a replay operation.
type ReplayResult struct {
	TotalEvents    int64         `json:"total_events"`
	ProcessedCount int64         `json:"processed_count"`
	SuccessCount   int64         `json:"success_count"`
	ErrorCount     int64         `json:"error_count"`
	SkippedCount   int64         `json:"skipped_count"`
	Duration       time.Duration `json:"duration"`
	Errors         []ReplayError `json:"errors,omitempty"`
}

// ReplayError holds error information for a failed event.
type ReplayError struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// Replay replays events based on the configuration.
func (r *EventReplayer) Replay(ctx context.Context, config ReplayConfig) (*ReplayResult, error) {
	startTime := time.Now()
	result := &ReplayResult{
		Errors: make([]ReplayError, 0),
	}

	// Get total count
	total, err := r.store.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}
	result.TotalEvents = total

	// Process events in batches
	offset := 0
	for {
		select {
		case <-ctx.Done():
			result.Duration = time.Since(startTime)
			return result, ctx.Err()
		default:
		}

		// Fetch batch
		events, err := r.fetchBatch(ctx, config, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch batch: %w", err)
		}

		if len(events) == 0 {
			break
		}

		// Process batch
		for _, event := range events {
			select {
			case <-ctx.Done():
				result.Duration = time.Since(startTime)
				return result, ctx.Err()
			default:
			}

			// Check event type filter
			if len(config.EventTypes) > 0 && !containsEventType(config.EventTypes, event.Type) {
				result.SkippedCount++
				continue
			}

			// Check tenant filter
			if config.TenantID != "" && event.TenantID != config.TenantID {
				result.SkippedCount++
				continue
			}

			// Check aggregate filter
			if config.AggregateID != "" && event.AggregateID != config.AggregateID {
				result.SkippedCount++
				continue
			}

			result.ProcessedCount++

			// Report progress
			if config.ProgressCallback != nil {
				config.ProgressCallback(result.ProcessedCount, total, event)
			}

			// Skip actual execution in dry run mode
			if config.DryRun {
				result.SuccessCount++
				continue
			}

			// Execute handlers
			if err := r.executeHandlers(ctx, event); err != nil {
				result.ErrorCount++
				replayErr := ReplayError{
					EventID:   event.ID,
					EventType: string(event.Type),
					Error:     err.Error(),
					Timestamp: time.Now(),
				}
				result.Errors = append(result.Errors, replayErr)

				if config.ErrorCallback != nil {
					config.ErrorCallback(event, err)
				}

				if !config.SkipErrors {
					result.Duration = time.Since(startTime)
					return result, fmt.Errorf("handler error for event %s: %w", event.ID, err)
				}
			} else {
				result.SuccessCount++
			}
		}

		offset += config.BatchSize
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// fetchBatch fetches a batch of events.
func (r *EventReplayer) fetchBatch(ctx context.Context, config ReplayConfig, offset int) ([]*Event, error) {
	// Use time range if specified
	if config.StartTime != nil || config.EndTime != nil {
		start := time.Time{}
		end := time.Now()

		if config.StartTime != nil {
			start = *config.StartTime
		}
		if config.EndTime != nil {
			end = *config.EndTime
		}

		return r.store.GetByTimeRange(ctx, start, end, config.BatchSize)
	}

	// Use aggregate filter if specified
	if config.AggregateID != "" {
		return r.store.GetByAggregateID(ctx, config.AggregateID)
	}

	// Default: get all with pagination
	return r.store.GetAll(ctx, offset, config.BatchSize)
}

// executeHandlers executes all registered handlers for an event.
func (r *EventReplayer) executeHandlers(ctx context.Context, event *Event) error {
	r.mu.RLock()
	handlers := r.handlers[event.Type]
	r.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

// containsEventType checks if a slice contains an event type.
func containsEventType(slice []EventType, item EventType) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ============================================================================
// Replay Job
// ============================================================================

// ReplayJob represents a replay job that can be run asynchronously.
type ReplayJob struct {
	ID        string       `json:"id"`
	Config    ReplayConfig `json:"config"`
	Status    JobStatus    `json:"status"`
	Result    *ReplayResult `json:"result,omitempty"`
	Error     string       `json:"error,omitempty"`
	StartedAt *time.Time   `json:"started_at,omitempty"`
	EndedAt   *time.Time   `json:"ended_at,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// JobStatus represents the status of a replay job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// ReplayJobManager manages replay jobs.
type ReplayJobManager struct {
	replayer *EventReplayer
	jobs     map[string]*ReplayJob
	cancels  map[string]context.CancelFunc
	mu       sync.RWMutex
}

// NewReplayJobManager creates a new replay job manager.
func NewReplayJobManager(replayer *EventReplayer) *ReplayJobManager {
	return &ReplayJobManager{
		replayer: replayer,
		jobs:     make(map[string]*ReplayJob),
		cancels:  make(map[string]context.CancelFunc),
	}
}

// CreateJob creates a new replay job.
func (m *ReplayJobManager) CreateJob(config ReplayConfig) *ReplayJob {
	job := &ReplayJob{
		ID:        uuid.New().String(),
		Config:    config,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
	}

	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()

	return job
}

// StartJob starts a replay job asynchronously.
func (m *ReplayJobManager) StartJob(jobID string) error {
	m.mu.Lock()
	job, ok := m.jobs[jobID]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status != JobStatusPending {
		m.mu.Unlock()
		return fmt.Errorf("job already started or completed")
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancels[jobID] = cancel

	now := time.Now()
	job.Status = JobStatusRunning
	job.StartedAt = &now
	m.mu.Unlock()

	// Run replay asynchronously
	go func() {
		result, err := m.replayer.Replay(ctx, job.Config)

		m.mu.Lock()
		defer m.mu.Unlock()

		endTime := time.Now()
		job.EndedAt = &endTime
		job.Result = result

		if err != nil {
			if ctx.Err() == context.Canceled {
				job.Status = JobStatusCancelled
			} else {
				job.Status = JobStatusFailed
				job.Error = err.Error()
			}
		} else {
			job.Status = JobStatusCompleted
		}

		delete(m.cancels, jobID)
	}()

	return nil
}

// CancelJob cancels a running replay job.
func (m *ReplayJobManager) CancelJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, ok := m.cancels[jobID]
	if !ok {
		return fmt.Errorf("job not running: %s", jobID)
	}

	cancel()
	return nil
}

// GetJob retrieves a job by ID.
func (m *ReplayJobManager) GetJob(jobID string) (*ReplayJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs lists all jobs.
func (m *ReplayJobManager) ListJobs() []*ReplayJob {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobs := make([]*ReplayJob, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// DeleteJob deletes a job.
func (m *ReplayJobManager) DeleteJob(jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	job, ok := m.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status == JobStatusRunning {
		return fmt.Errorf("cannot delete running job")
	}

	delete(m.jobs, jobID)
	return nil
}

// ============================================================================
// Snapshot Support
// ============================================================================

// Snapshot represents a state snapshot at a point in time.
type Snapshot struct {
	ID          string          `json:"id"`
	AggregateID string          `json:"aggregate_id"`
	Version     int             `json:"version"`
	State       json.RawMessage `json:"state"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SnapshotStore defines the interface for snapshot persistence.
type SnapshotStore interface {
	Save(ctx context.Context, snapshot *Snapshot) error
	GetLatest(ctx context.Context, aggregateID string) (*Snapshot, error)
	GetByVersion(ctx context.Context, aggregateID string, version int) (*Snapshot, error)
	Delete(ctx context.Context, aggregateID string, beforeVersion int) error
}

// SnapshotReplayer replays events with snapshot support.
type SnapshotReplayer struct {
	*EventReplayer
	snapshotStore    SnapshotStore
	snapshotInterval int
}

// NewSnapshotReplayer creates a snapshot-aware replayer.
func NewSnapshotReplayer(eventStore EventStore, snapshotStore SnapshotStore, interval int) *SnapshotReplayer {
	return &SnapshotReplayer{
		EventReplayer:    NewEventReplayer(eventStore),
		snapshotStore:    snapshotStore,
		snapshotInterval: interval,
	}
}

// ReplayFromSnapshot replays events starting from the latest snapshot.
func (r *SnapshotReplayer) ReplayFromSnapshot(ctx context.Context, aggregateID string, applySnapshot func([]byte) error, applyEvent func(*Event) error) error {
	// Get latest snapshot
	snapshot, err := r.snapshotStore.GetLatest(ctx, aggregateID)
	if err != nil {
		// No snapshot, replay from beginning
		snapshot = nil
	}

	// Apply snapshot if exists
	var startVersion int
	if snapshot != nil {
		if err := applySnapshot(snapshot.State); err != nil {
			return fmt.Errorf("failed to apply snapshot: %w", err)
		}
		startVersion = snapshot.Version + 1
	}

	// Get events after snapshot
	events, err := r.store.GetByAggregateID(ctx, aggregateID)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// Apply events
	for _, event := range events {
		if event.Version >= startVersion {
			if err := applyEvent(event); err != nil {
				return fmt.Errorf("failed to apply event %s: %w", event.ID, err)
			}
		}
	}

	return nil
}
