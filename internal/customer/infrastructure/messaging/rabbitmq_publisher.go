// Package messaging provides messaging infrastructure for the Customer service.
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/ports"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
)

const (
	// Exchange names
	CustomerEventsExchange = "customer.events"
	CustomerTopicExchange  = "customer.topic"

	// Queue names
	CustomerCreatedQueue = "customer.created"
	CustomerUpdatedQueue = "customer.updated"
	CustomerDeletedQueue = "customer.deleted"
	ContactCreatedQueue  = "contact.created"
	ContactUpdatedQueue  = "contact.updated"
	ContactDeletedQueue  = "contact.deleted"
)

// RabbitMQConfig holds RabbitMQ configuration.
type RabbitMQConfig struct {
	URL              string
	Exchange         string
	ExchangeType     string
	Durable          bool
	AutoDelete       bool
	DeliveryMode     uint8
	ContentType      string
	ReconnectDelay   time.Duration
	MaxReconnectTries int
}

// DefaultRabbitMQConfig returns default RabbitMQ configuration.
func DefaultRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Exchange:         CustomerEventsExchange,
		ExchangeType:     "topic",
		Durable:          true,
		AutoDelete:       false,
		DeliveryMode:     amqp.Persistent,
		ContentType:      "application/json",
		ReconnectDelay:   5 * time.Second,
		MaxReconnectTries: 10,
	}
}

// RabbitMQPublisher implements ports.EventPublisher using RabbitMQ.
type RabbitMQPublisher struct {
	config     RabbitMQConfig
	conn       *amqp.Connection
	channel    *amqp.Channel
	mu         sync.RWMutex
	closed     bool
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

// Publish publishes a domain event.
func (p *RabbitMQPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("publisher is closed")
	}

	if p.channel == nil {
		return fmt.Errorf("channel is not available")
	}

	// Serialize event
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Build routing key from event type
	routingKey := event.EventType()

	// Create message
	msg := amqp.Publishing{
		DeliveryMode: p.config.DeliveryMode,
		ContentType:  p.config.ContentType,
		Body:         body,
		Timestamp:    time.Now().UTC(),
		MessageId:    event.EventID().String(),
		Headers: amqp.Table{
			"event_type":   event.EventType(),
			"aggregate_id": event.AggregateID().String(),
			"occurred_at":  event.OccurredAt().Format(time.RFC3339),
		},
	}

	// Publish with context timeout
	return p.channel.PublishWithContext(
		ctx,
		p.config.Exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		msg,
	)
}

// PublishBatch publishes multiple events.
func (p *RabbitMQPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
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

// DeclareQueues declares the necessary queues and bindings.
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
		{CustomerCreatedQueue, "customer.created"},
		{CustomerUpdatedQueue, "customer.updated"},
		{CustomerDeletedQueue, "customer.deleted"},
		{ContactCreatedQueue, "contact.created"},
		{ContactUpdatedQueue, "contact.updated"},
		{ContactDeletedQueue, "contact.deleted"},
	}

	for _, q := range queues {
		// Declare queue
		_, err := p.channel.QueueDeclare(
			q.name,
			p.config.Durable,
			p.config.AutoDelete,
			false, // exclusive
			false, // no-wait
			nil,   // arguments
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

	return nil
}

// Ensure RabbitMQPublisher implements ports.EventPublisher
var _ ports.EventPublisher = (*RabbitMQPublisher)(nil)

// ============================================================================
// Outbox Processor
// ============================================================================

// OutboxProcessor processes outbox entries and publishes events.
type OutboxProcessor struct {
	publisher ports.EventPublisher
	outbox    domain.OutboxRepository
	batchSize int
	interval  time.Duration
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// NewOutboxProcessor creates a new outbox processor.
func NewOutboxProcessor(publisher ports.EventPublisher, outbox domain.OutboxRepository, batchSize int, interval time.Duration) *OutboxProcessor {
	return &OutboxProcessor{
		publisher: publisher,
		outbox:    outbox,
		batchSize: batchSize,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// Start starts the outbox processor.
func (p *OutboxProcessor) Start(ctx context.Context) {
	p.wg.Add(1)
	go p.run(ctx)
}

// Stop stops the outbox processor.
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
	entries, err := p.outbox.FindPending(ctx, p.batchSize)
	if err != nil {
		return
	}

	for _, entry := range entries {
		// Deserialize the event
		event := &outboxEvent{
			id:          entry.ID,
			eventType:   entry.EventType,
			aggregateID: entry.AggregateID,
			payload:     entry.Payload,
			occurredAt:  entry.CreatedAt,
		}

		// Publish the event
		if err := p.publisher.Publish(ctx, event); err != nil {
			_ = p.outbox.MarkAsFailed(ctx, entry.ID, err.Error())
			continue
		}

		// Mark as processed
		_ = p.outbox.MarkAsProcessed(ctx, entry.ID)
	}
}

// outboxEvent implements domain.DomainEvent for outbox entries.
type outboxEvent struct {
	id          interface{}
	eventType   string
	aggregateID interface{}
	payload     []byte
	occurredAt  time.Time
}

func (e *outboxEvent) EventID() interface{}      { return e.id }
func (e *outboxEvent) EventType() string         { return e.eventType }
func (e *outboxEvent) AggregateID() interface{}  { return e.aggregateID }
func (e *outboxEvent) OccurredAt() time.Time     { return e.occurredAt }
func (e *outboxEvent) MarshalJSON() ([]byte, error) { return e.payload, nil }
