// Package events provides event bus abstractions for the CRM application.
package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/logger"
)

// RabbitMQEventBus implements EventBus using RabbitMQ.
type RabbitMQEventBus struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	config       *config.RabbitMQConfig
	log          *logger.Logger
	mu           sync.RWMutex
	closed       bool
	reconnecting bool
	consumers    map[string]*consumer
}

type consumer struct {
	queue    string
	handler  Handler
	delivery <-chan amqp.Delivery
	done     chan struct{}
}

// NewRabbitMQEventBus creates a new RabbitMQ event bus.
func NewRabbitMQEventBus(cfg *config.RabbitMQConfig, log *logger.Logger) (*RabbitMQEventBus, error) {
	bus := &RabbitMQEventBus{
		config:    cfg,
		log:       log,
		consumers: make(map[string]*consumer),
	}

	if err := bus.connect(); err != nil {
		return nil, err
	}

	// Start connection monitor
	go bus.monitorConnection()

	return bus, nil
}

// connect establishes connection to RabbitMQ.
func (b *RabbitMQEventBus) connect() error {
	conn, err := amqp.Dial(b.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare the exchange
	if err := channel.ExchangeDeclare(
		b.config.Exchange,     // name
		b.config.ExchangeType, // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Set QoS
	if err := channel.Qos(b.config.PrefetchCount, 0, false); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	b.mu.Lock()
	b.conn = conn
	b.channel = channel
	b.mu.Unlock()

	b.log.Info().
		Str("url", b.config.URL).
		Str("exchange", b.config.Exchange).
		Msg("Connected to RabbitMQ")

	return nil
}

// monitorConnection monitors the connection and reconnects if necessary.
func (b *RabbitMQEventBus) monitorConnection() {
	for {
		b.mu.RLock()
		if b.closed {
			b.mu.RUnlock()
			return
		}
		conn := b.conn
		b.mu.RUnlock()

		if conn == nil {
			time.Sleep(b.config.ReconnectDelay)
			continue
		}

		// Wait for connection close notification
		connClose := conn.NotifyClose(make(chan *amqp.Error))
		err := <-connClose

		if err != nil {
			b.log.Error().Err(err).Msg("RabbitMQ connection closed")
		}

		b.mu.Lock()
		if b.closed {
			b.mu.Unlock()
			return
		}
		b.reconnecting = true
		b.mu.Unlock()

		// Attempt to reconnect with exponential backoff
		delay := b.config.ReconnectDelay
		for {
			b.mu.RLock()
			if b.closed {
				b.mu.RUnlock()
				return
			}
			b.mu.RUnlock()

			b.log.Info().Dur("delay", delay).Msg("Attempting to reconnect to RabbitMQ")

			if err := b.connect(); err != nil {
				b.log.Error().Err(err).Msg("Failed to reconnect to RabbitMQ")
				time.Sleep(delay)
				delay = delay * 2
				if delay > b.config.MaxReconnectDelay {
					delay = b.config.MaxReconnectDelay
				}
				continue
			}

			// Reconnection successful, restore consumers
			b.restoreConsumers()

			b.mu.Lock()
			b.reconnecting = false
			b.mu.Unlock()

			break
		}
	}
}

// restoreConsumers restores consumers after reconnection.
func (b *RabbitMQEventBus) restoreConsumers() {
	b.mu.RLock()
	consumers := make(map[string]*consumer, len(b.consumers))
	for k, v := range b.consumers {
		consumers[k] = v
	}
	b.mu.RUnlock()

	for queue, c := range consumers {
		if err := b.subscribeToQueue(queue, c.handler); err != nil {
			b.log.Error().Err(err).Str("queue", queue).Msg("Failed to restore consumer")
		}
	}
}

// Publish publishes an event to the event bus.
func (b *RabbitMQEventBus) Publish(ctx context.Context, event *Event) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return fmt.Errorf("event bus is closed")
	}
	channel := b.channel
	b.mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("channel is not available")
	}

	body, err := event.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    event.Timestamp,
		MessageId:    event.ID,
		Type:         string(event.Type),
		Headers: amqp.Table{
			"tenant_id":    event.TenantID,
			"aggregate_id": event.AggregateID,
			"version":      event.Version,
		},
		Body: body,
	}

	routingKey := string(event.Type)

	if err := channel.PublishWithContext(
		ctx,
		b.config.Exchange, // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		msg,
	); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	b.log.Debug().
		Str("event_id", event.ID).
		Str("event_type", string(event.Type)).
		Str("tenant_id", event.TenantID).
		Msg("Event published")

	return nil
}

// PublishBatch publishes multiple events to the event bus.
func (b *RabbitMQEventBus) PublishBatch(ctx context.Context, events []*Event) error {
	for _, event := range events {
		if err := b.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.ID, err)
		}
	}
	return nil
}

// Subscribe subscribes to events of the specified types.
func (b *RabbitMQEventBus) Subscribe(ctx context.Context, eventTypes []EventType, handler Handler) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return fmt.Errorf("event bus is closed")
	}
	channel := b.channel
	b.mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("channel is not available")
	}

	// Create a unique queue name for this subscriber
	queueName := fmt.Sprintf("subscriber.%d", time.Now().UnixNano())

	// Declare queue
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    b.config.Exchange + ".dlx",
			"x-dead-letter-routing-key": queueName + ".dlq",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange for each event type
	for _, eventType := range eventTypes {
		if err := channel.QueueBind(
			queue.Name,        // queue name
			string(eventType), // routing key
			b.config.Exchange, // exchange
			false,             // no-wait
			nil,               // arguments
		); err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	return b.subscribeToQueue(queue.Name, handler)
}

// subscribeToQueue starts consuming from a queue.
func (b *RabbitMQEventBus) subscribeToQueue(queueName string, handler Handler) error {
	b.mu.RLock()
	channel := b.channel
	b.mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("channel is not available")
	}

	delivery, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	done := make(chan struct{})

	c := &consumer{
		queue:    queueName,
		handler:  handler,
		delivery: delivery,
		done:     done,
	}

	b.mu.Lock()
	b.consumers[queueName] = c
	b.mu.Unlock()

	go b.consume(c)

	b.log.Info().Str("queue", queueName).Msg("Started consuming events")

	return nil
}

// consume processes messages from the delivery channel.
func (b *RabbitMQEventBus) consume(c *consumer) {
	for {
		select {
		case <-c.done:
			return
		case d, ok := <-c.delivery:
			if !ok {
				return
			}

			event, err := Unmarshal(d.Body)
			if err != nil {
				b.log.Error().Err(err).Msg("Failed to unmarshal event")
				d.Nack(false, false) // Don't requeue malformed messages
				continue
			}

			ctx := context.Background()
			if err := c.handler(ctx, event); err != nil {
				b.log.Error().
					Err(err).
					Str("event_id", event.ID).
					Str("event_type", string(event.Type)).
					Msg("Failed to handle event")
				d.Nack(false, true) // Requeue for retry
				continue
			}

			d.Ack(false)

			b.log.Debug().
				Str("event_id", event.ID).
				Str("event_type", string(event.Type)).
				Msg("Event handled successfully")
		}
	}
}

// Unsubscribe unsubscribes from all events.
func (b *RabbitMQEventBus) Unsubscribe() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, c := range b.consumers {
		close(c.done)
	}
	b.consumers = make(map[string]*consumer)

	return nil
}

// Close closes the event bus connection.
func (b *RabbitMQEventBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true

	// Close all consumers
	for _, c := range b.consumers {
		close(c.done)
	}

	var errs []error

	if b.channel != nil {
		if err := b.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if b.conn != nil {
		if err := b.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	b.log.Info().Msg("RabbitMQ event bus closed")

	if len(errs) > 0 {
		return fmt.Errorf("errors closing event bus: %v", errs)
	}

	return nil
}
