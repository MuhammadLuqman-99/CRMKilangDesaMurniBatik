// Package messaging provides messaging infrastructure for the Sales Pipeline service.
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// Exchange names
	SalesEventsExchange = "sales.events"
	SalesTopicExchange  = "sales.topic"

	// Queue names - Leads
	LeadCreatedQueue    = "sales.lead.created"
	LeadUpdatedQueue    = "sales.lead.updated"
	LeadConvertedQueue  = "sales.lead.converted"
	LeadQualifiedQueue  = "sales.lead.qualified"

	// Queue names - Opportunities
	OpportunityCreatedQueue  = "sales.opportunity.created"
	OpportunityUpdatedQueue  = "sales.opportunity.updated"
	OpportunityWonQueue      = "sales.opportunity.won"
	OpportunityLostQueue     = "sales.opportunity.lost"
	OpportunityStagedQueue   = "sales.opportunity.stage_changed"

	// Queue names - Deals
	DealCreatedQueue   = "sales.deal.created"
	DealUpdatedQueue   = "sales.deal.updated"
	DealCompletedQueue = "sales.deal.completed"
	DealCancelledQueue = "sales.deal.cancelled"

	// Queue names - Pipelines
	PipelineCreatedQueue = "sales.pipeline.created"
	PipelineUpdatedQueue = "sales.pipeline.updated"
)

// ============================================================================
// Configuration
// ============================================================================

// RabbitMQConfig holds RabbitMQ configuration.
type RabbitMQConfig struct {
	URL               string
	Exchange          string
	ExchangeType      string
	Durable           bool
	AutoDelete        bool
	DeliveryMode      uint8
	ContentType       string
	ReconnectDelay    time.Duration
	MaxReconnectTries int
	PrefetchCount     int
}

// DefaultRabbitMQConfig returns default RabbitMQ configuration.
func DefaultRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Exchange:          SalesEventsExchange,
		ExchangeType:      "topic",
		Durable:           true,
		AutoDelete:        false,
		DeliveryMode:      amqp.Persistent,
		ContentType:       "application/json",
		ReconnectDelay:    5 * time.Second,
		MaxReconnectTries: 10,
		PrefetchCount:     10,
	}
}

// ============================================================================
// Event Publisher Interface
// ============================================================================

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	// Publish publishes a single domain event
	Publish(ctx context.Context, event domain.DomainEvent) error

	// PublishBatch publishes multiple domain events
	PublishBatch(ctx context.Context, events []domain.DomainEvent) error

	// Close closes the publisher connection
	Close() error

	// IsConnected checks if the publisher is connected
	IsConnected() bool
}

// ============================================================================
// RabbitMQ Publisher Implementation
// ============================================================================

// RabbitMQPublisher implements EventPublisher using RabbitMQ.
type RabbitMQPublisher struct {
	config      RabbitMQConfig
	conn        *amqp.Connection
	channel     *amqp.Channel
	mu          sync.RWMutex
	closed      bool
	notifyClose chan *amqp.Error
}

// NewRabbitMQPublisher creates a new RabbitMQ event publisher.
func NewRabbitMQPublisher(config RabbitMQConfig) (*RabbitMQPublisher, error) {
	publisher := &RabbitMQPublisher{
		config: config,
	}

	if err := publisher.connect(); err != nil {
		return nil, err
	}

	return publisher, nil
}

// connect establishes connection to RabbitMQ.
func (p *RabbitMQPublisher) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, err := amqp.Dial(p.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	if err := ch.ExchangeDeclare(
		p.config.Exchange,
		p.config.ExchangeType,
		p.config.Durable,
		p.config.AutoDelete,
		false, // internal
		false, // no-wait
		nil,   // arguments
	); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Enable publisher confirms
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	p.conn = conn
	p.channel = ch
	p.notifyClose = make(chan *amqp.Error, 1)
	p.channel.NotifyClose(p.notifyClose)

	// Start reconnection handler
	go p.handleReconnect()

	return nil
}

// handleReconnect handles automatic reconnection.
func (p *RabbitMQPublisher) handleReconnect() {
	for {
		select {
		case err := <-p.notifyClose:
			if err == nil {
				return // Normal close
			}

			p.mu.Lock()
			if p.closed {
				p.mu.Unlock()
				return
			}
			p.mu.Unlock()

			// Attempt to reconnect
			for i := 0; i < p.config.MaxReconnectTries; i++ {
				time.Sleep(p.config.ReconnectDelay)
				if err := p.connect(); err == nil {
					break
				}
			}
		}
	}
}

// Publish publishes a domain event to RabbitMQ.
func (p *RabbitMQPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("publisher is closed")
	}

	if p.channel == nil {
		return fmt.Errorf("channel is not available")
	}

	// Serialize the entire event (events embed BaseEvent which has JSON tags)
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Build routing key from event type (e.g., "lead.created", "opportunity.won")
	routingKey := fmt.Sprintf("sales.%s", event.EventType())

	// Create message
	msg := amqp.Publishing{
		DeliveryMode: p.config.DeliveryMode,
		ContentType:  p.config.ContentType,
		Body:         body,
		Timestamp:    time.Now().UTC(),
		MessageId:    event.EventID().String(),
		Headers: amqp.Table{
			"event_type":     event.EventType(),
			"aggregate_type": event.AggregateType(),
			"aggregate_id":   event.AggregateID().String(),
			"tenant_id":      event.TenantID().String(),
			"version":        int32(event.Version()),
			"occurred_at":    event.OccurredAt().Format(time.RFC3339Nano),
		},
	}

	// Publish with context timeout
	confirm, err := p.channel.PublishWithDeferredConfirmWithContext(
		ctx,
		p.config.Exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		msg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Wait for confirmation
	if !confirm.Wait() {
		return fmt.Errorf("failed to confirm event publication")
	}

	return nil
}

// PublishBatch publishes multiple events.
func (p *RabbitMQPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.EventID(), err)
		}
	}
	return nil
}

// Close closes the RabbitMQ connection.
func (p *RabbitMQPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true

	if p.channel != nil {
		p.channel.Close()
	}

	if p.conn != nil {
		return p.conn.Close()
	}

	return nil
}

// IsConnected checks if the publisher is connected.
func (p *RabbitMQPublisher) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return !p.closed && p.conn != nil && !p.conn.IsClosed()
}

// DeclareQueues declares the necessary queues and bindings for sales events.
func (p *RabbitMQPublisher) DeclareQueues() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel == nil {
		return fmt.Errorf("channel is not available")
	}

	queues := []struct {
		name       string
		routingKey string
	}{
		// Lead queues
		{LeadCreatedQueue, "sales.lead.created"},
		{LeadUpdatedQueue, "sales.lead.updated"},
		{LeadConvertedQueue, "sales.lead.converted"},
		{LeadQualifiedQueue, "sales.lead.qualified"},

		// Opportunity queues
		{OpportunityCreatedQueue, "sales.opportunity.created"},
		{OpportunityUpdatedQueue, "sales.opportunity.updated"},
		{OpportunityWonQueue, "sales.opportunity.won"},
		{OpportunityLostQueue, "sales.opportunity.lost"},
		{OpportunityStagedQueue, "sales.opportunity.stage_changed"},

		// Deal queues
		{DealCreatedQueue, "sales.deal.created"},
		{DealUpdatedQueue, "sales.deal.updated"},
		{DealCompletedQueue, "sales.deal.completed"},
		{DealCancelledQueue, "sales.deal.cancelled"},

		// Pipeline queues
		{PipelineCreatedQueue, "sales.pipeline.created"},
		{PipelineUpdatedQueue, "sales.pipeline.updated"},
	}

	for _, q := range queues {
		// Declare queue
		_, err := p.channel.QueueDeclare(
			q.name,
			p.config.Durable,
			p.config.AutoDelete,
			false, // exclusive
			false, // no-wait
			amqp.Table{
				"x-dead-letter-exchange": fmt.Sprintf("%s.dlx", p.config.Exchange),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.name, err)
		}

		// Bind queue to exchange
		if err := p.channel.QueueBind(
			q.name,
			q.routingKey,
			p.config.Exchange,
			false, // no-wait
			nil,   // arguments
		); err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", q.name, err)
		}
	}

	// Declare dead letter exchange and queues
	if err := p.declareDLX(); err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	return nil
}

// declareDLX declares the dead letter exchange and queue.
func (p *RabbitMQPublisher) declareDLX() error {
	dlxName := fmt.Sprintf("%s.dlx", p.config.Exchange)
	dlqName := fmt.Sprintf("%s.dlq", p.config.Exchange)

	// Declare DLX exchange
	if err := p.channel.ExchangeDeclare(
		dlxName,
		"fanout",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	// Declare DLQ
	_, err := p.channel.QueueDeclare(
		dlqName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Bind DLQ to DLX
	if err := p.channel.QueueBind(
		dlqName,
		"#", // catch all
		dlxName,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind DLQ: %w", err)
	}

	return nil
}

// ============================================================================
// Outbox Processor
// ============================================================================

// OutboxProcessor processes outbox entries and publishes events.
type OutboxProcessor struct {
	publisher   EventPublisher
	outboxRepo  OutboxRepository
	batchSize   int
	interval    time.Duration
	maxRetries  int
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// OutboxRepository defines the interface for outbox operations.
type OutboxRepository interface {
	FindUnpublished(ctx context.Context, limit int) ([]*OutboxEntry, error)
	FindRetryable(ctx context.Context, maxRetries int, limit int) ([]*OutboxEntry, error)
	MarkAsPublished(ctx context.Context, id uuid.UUID) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, errorMsg string) error
	DeletePublished(ctx context.Context, olderThan time.Time) (int64, error)
}

// OutboxEntry represents an outbox entry.
type OutboxEntry struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	EventType     string
	AggregateID   uuid.UUID
	AggregateType string
	Payload       []byte
	RetryCount    int
	CreatedAt     time.Time
}

// OutboxProcessorConfig holds configuration for the outbox processor.
type OutboxProcessorConfig struct {
	BatchSize     int
	PollInterval  time.Duration
	MaxRetries    int
	CleanupAge    time.Duration
	CleanupPeriod time.Duration
}

// DefaultOutboxProcessorConfig returns default configuration.
func DefaultOutboxProcessorConfig() OutboxProcessorConfig {
	return OutboxProcessorConfig{
		BatchSize:     100,
		PollInterval:  1 * time.Second,
		MaxRetries:    5,
		CleanupAge:    24 * time.Hour,
		CleanupPeriod: 1 * time.Hour,
	}
}

// NewOutboxProcessor creates a new outbox processor.
func NewOutboxProcessor(publisher EventPublisher, outboxRepo OutboxRepository, config OutboxProcessorConfig) *OutboxProcessor {
	return &OutboxProcessor{
		publisher:  publisher,
		outboxRepo: outboxRepo,
		batchSize:  config.BatchSize,
		interval:   config.PollInterval,
		maxRetries: config.MaxRetries,
		stopCh:     make(chan struct{}),
	}
}

// Start starts the outbox processor.
func (p *OutboxProcessor) Start(ctx context.Context) {
	p.wg.Add(1)
	go p.run(ctx)
}

// Stop stops the outbox processor gracefully.
func (p *OutboxProcessor) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}

// run is the main processing loop.
func (p *OutboxProcessor) run(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.processEntries(ctx)
		}
	}
}

// processEntries processes pending outbox entries.
func (p *OutboxProcessor) processEntries(ctx context.Context) {
	// Process new entries
	entries, err := p.outboxRepo.FindUnpublished(ctx, p.batchSize)
	if err != nil {
		return
	}

	p.publishEntries(ctx, entries)

	// Process retryable entries
	retryable, err := p.outboxRepo.FindRetryable(ctx, p.maxRetries, p.batchSize)
	if err != nil {
		return
	}

	p.publishEntries(ctx, retryable)
}

// publishEntries publishes a batch of outbox entries.
func (p *OutboxProcessor) publishEntries(ctx context.Context, entries []*OutboxEntry) {
	for _, entry := range entries {
		// Create event from outbox entry
		event := &outboxEvent{
			id:            entry.ID,
			tenantID:      entry.TenantID,
			aggregateID:   entry.AggregateID,
			aggregateType: entry.AggregateType,
			eventType:     entry.EventType,
			payload:       entry.Payload,
			occurredAt:    entry.CreatedAt,
			version:       entry.RetryCount + 1,
		}

		// Publish the event
		if err := p.publisher.Publish(ctx, event); err != nil {
			_ = p.outboxRepo.MarkAsFailed(ctx, entry.ID, err.Error())
			continue
		}

		// Mark as published
		_ = p.outboxRepo.MarkAsPublished(ctx, entry.ID)
	}
}

// outboxEvent implements domain.DomainEvent for outbox entries.
type outboxEvent struct {
	id            uuid.UUID
	tenantID      uuid.UUID
	aggregateID   uuid.UUID
	aggregateType string
	eventType     string
	payload       []byte
	occurredAt    time.Time
	version       int
}

func (e *outboxEvent) EventID() uuid.UUID      { return e.id }
func (e *outboxEvent) TenantID() uuid.UUID     { return e.tenantID }
func (e *outboxEvent) AggregateID() uuid.UUID  { return e.aggregateID }
func (e *outboxEvent) AggregateType() string   { return e.aggregateType }
func (e *outboxEvent) EventType() string       { return e.eventType }
func (e *outboxEvent) Version() int            { return e.version }
func (e *outboxEvent) OccurredAt() time.Time   { return e.occurredAt }

// MarshalJSON returns the payload as JSON
func (e *outboxEvent) MarshalJSON() ([]byte, error) {
	return e.payload, nil
}

// ============================================================================
// Outbox Cleanup Worker
// ============================================================================

// OutboxCleanupWorker periodically cleans up old published entries.
type OutboxCleanupWorker struct {
	outboxRepo    OutboxRepository
	cleanupAge    time.Duration
	cleanupPeriod time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// NewOutboxCleanupWorker creates a new cleanup worker.
func NewOutboxCleanupWorker(outboxRepo OutboxRepository, config OutboxProcessorConfig) *OutboxCleanupWorker {
	return &OutboxCleanupWorker{
		outboxRepo:    outboxRepo,
		cleanupAge:    config.CleanupAge,
		cleanupPeriod: config.CleanupPeriod,
		stopCh:        make(chan struct{}),
	}
}

// Start starts the cleanup worker.
func (w *OutboxCleanupWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
}

// Stop stops the cleanup worker gracefully.
func (w *OutboxCleanupWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
}

// run is the main cleanup loop.
func (w *OutboxCleanupWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.cleanupPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			olderThan := time.Now().UTC().Add(-w.cleanupAge)
			_, _ = w.outboxRepo.DeletePublished(ctx, olderThan)
		}
	}
}

// Ensure RabbitMQPublisher implements EventPublisher
var _ EventPublisher = (*RabbitMQPublisher)(nil)
