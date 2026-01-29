// Package containers provides test container implementations for integration testing.
package containers

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQContainer represents a RabbitMQ test container configuration.
type RabbitMQContainer struct {
	Host          string
	Port          string
	User          string
	Password      string
	VHost         string
	Exchange      string
	ExchangeType  string
	Conn          *amqp.Connection
	Channel       *amqp.Channel
	mu            sync.RWMutex
	queues        []string
	consumers     map[string]<-chan amqp.Delivery
}

// RabbitMQContainerConfig holds configuration for RabbitMQ container.
type RabbitMQContainerConfig struct {
	User         string
	Password     string
	VHost        string
	Exchange     string
	ExchangeType string
}

// DefaultRabbitMQConfig returns default RabbitMQ configuration.
func DefaultRabbitMQConfig() RabbitMQContainerConfig {
	return RabbitMQContainerConfig{
		User:         "guest",
		Password:     "guest",
		VHost:        "/",
		Exchange:     "crm.events.test",
		ExchangeType: "topic",
	}
}

// NewRabbitMQContainer creates a new RabbitMQ container for testing.
// For integration tests, this connects to the docker-compose RabbitMQ instance.
func NewRabbitMQContainer(ctx context.Context, cfg RabbitMQContainerConfig) (*RabbitMQContainer, error) {
	container := &RabbitMQContainer{
		Host:         getEnvOrDefault("TEST_RABBITMQ_HOST", "localhost"),
		Port:         getEnvOrDefault("TEST_RABBITMQ_PORT", "5672"),
		User:         getEnvOrDefault("TEST_RABBITMQ_USER", cfg.User),
		Password:     getEnvOrDefault("TEST_RABBITMQ_PASSWORD", cfg.Password),
		VHost:        getEnvOrDefault("TEST_RABBITMQ_VHOST", cfg.VHost),
		Exchange:     cfg.Exchange,
		ExchangeType: cfg.ExchangeType,
		queues:       make([]string, 0),
		consumers:    make(map[string]<-chan amqp.Delivery),
	}

	// Build connection URL
	url := container.ConnectionURL()

	// Connect to RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Open a channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare the exchange
	err = channel.ExchangeDeclare(
		container.Exchange,     // name
		container.ExchangeType, // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare dead letter exchange
	err = channel.ExchangeDeclare(
		container.Exchange+".dlx", // name
		"topic",                   // type
		true,                      // durable
		false,                     // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLX exchange: %w", err)
	}

	container.Conn = conn
	container.Channel = channel

	return container, nil
}

// ConnectionURL returns the RabbitMQ connection URL.
func (c *RabbitMQContainer) ConnectionURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		c.User, c.Password, c.Host, c.Port, c.VHost)
}

// GetConnection returns the RabbitMQ connection.
func (c *RabbitMQContainer) GetConnection() *amqp.Connection {
	return c.Conn
}

// GetChannel returns the RabbitMQ channel.
func (c *RabbitMQContainer) GetChannel() *amqp.Channel {
	return c.Channel
}

// DeclareQueue declares a queue and binds it to the exchange.
func (c *RabbitMQContainer) DeclareQueue(queueName string, routingKeys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Declare queue with dead letter exchange
	queue, err := c.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    c.Exchange + ".dlx",
			"x-dead-letter-routing-key": queueName + ".dlq",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange for each routing key
	for _, routingKey := range routingKeys {
		err = c.Channel.QueueBind(
			queue.Name,  // queue name
			routingKey,  // routing key
			c.Exchange,  // exchange
			false,       // no-wait
			nil,         // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	c.queues = append(c.queues, queueName)

	return nil
}

// DeclareDeadLetterQueue declares a dead letter queue.
func (c *RabbitMQContainer) DeclareDeadLetterQueue(queueName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	dlqName := queueName + ".dlq"

	// Declare DLQ
	_, err := c.Channel.QueueDeclare(
		dlqName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Bind DLQ to DLX
	err = c.Channel.QueueBind(
		dlqName,               // queue name
		dlqName,               // routing key
		c.Exchange+".dlx",     // exchange
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ: %w", err)
	}

	return nil
}

// Publish publishes a message to the exchange.
func (c *RabbitMQContainer) Publish(ctx context.Context, routingKey string, message []byte) error {
	return c.Channel.PublishWithContext(
		ctx,
		c.Exchange,  // exchange
		routingKey,  // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         message,
		},
	)
}

// PublishWithHeaders publishes a message with headers.
func (c *RabbitMQContainer) PublishWithHeaders(ctx context.Context, routingKey string, message []byte, headers amqp.Table) error {
	return c.Channel.PublishWithContext(
		ctx,
		c.Exchange,  // exchange
		routingKey,  // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers:      headers,
			Body:         message,
		},
	)
}

// Consume starts consuming messages from a queue.
func (c *RabbitMQContainer) Consume(queueName string) (<-chan amqp.Delivery, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	msgs, err := c.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume from queue: %w", err)
	}

	c.consumers[queueName] = msgs

	return msgs, nil
}

// ConsumeOne consumes a single message from a queue with timeout.
func (c *RabbitMQContainer) ConsumeOne(ctx context.Context, queueName string, timeout time.Duration) (amqp.Delivery, error) {
	msgs, err := c.Consume(queueName)
	if err != nil {
		return amqp.Delivery{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case msg := <-msgs:
		return msg, nil
	case <-ctx.Done():
		return amqp.Delivery{}, fmt.Errorf("timeout waiting for message")
	}
}

// ConsumeN consumes N messages from a queue with timeout.
func (c *RabbitMQContainer) ConsumeN(ctx context.Context, queueName string, n int, timeout time.Duration) ([]amqp.Delivery, error) {
	msgs, err := c.Consume(queueName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	received := make([]amqp.Delivery, 0, n)
	for i := 0; i < n; i++ {
		select {
		case msg := <-msgs:
			received = append(received, msg)
		case <-ctx.Done():
			return received, fmt.Errorf("timeout waiting for message %d/%d", len(received)+1, n)
		}
	}

	return received, nil
}

// PurgeQueue purges all messages from a queue.
func (c *RabbitMQContainer) PurgeQueue(queueName string) (int, error) {
	return c.Channel.QueuePurge(queueName, false)
}

// DeleteQueue deletes a queue.
func (c *RabbitMQContainer) DeleteQueue(queueName string) error {
	_, err := c.Channel.QueueDelete(queueName, false, false, false)
	return err
}

// GetQueueInfo returns information about a queue.
func (c *RabbitMQContainer) GetQueueInfo(queueName string) (amqp.Queue, error) {
	return c.Channel.QueueInspect(queueName)
}

// Cleanup cleans up all queues created during testing.
func (c *RabbitMQContainer) Cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	// Purge and delete all queues
	for _, queueName := range c.queues {
		if _, err := c.Channel.QueuePurge(queueName, false); err != nil {
			errs = append(errs, fmt.Errorf("failed to purge queue %s: %w", queueName, err))
		}
		if _, err := c.Channel.QueueDelete(queueName, false, false, false); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete queue %s: %w", queueName, err))
		}
	}

	c.queues = make([]string, 0)
	c.consumers = make(map[string]<-chan amqp.Delivery)

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// Close closes the RabbitMQ connection.
func (c *RabbitMQContainer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	if c.Channel != nil {
		if err := c.Channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	return nil
}

// WaitForReady waits for RabbitMQ to be ready.
func (c *RabbitMQContainer) WaitForReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for RabbitMQ to be ready")
		case <-ticker.C:
			if c.Conn != nil && !c.Conn.IsClosed() {
				return nil
			}
		}
	}
}

// SetupTestQueues sets up test queues for integration testing.
func (c *RabbitMQContainer) SetupTestQueues() error {
	testQueues := []struct {
		name        string
		routingKeys []string
	}{
		{"test.user.created", []string{"user.created"}},
		{"test.user.updated", []string{"user.updated"}},
		{"test.customer.created", []string{"customer.created"}},
		{"test.lead.created", []string{"lead.created"}},
		{"test.lead.converted", []string{"lead.converted"}},
		{"test.deal.won", []string{"deal.won"}},
		{"test.deal.lost", []string{"deal.lost"}},
		{"test.notification.email", []string{"notification.email.*"}},
		{"test.notification.sms", []string{"notification.sms.*"}},
		{"test.all.events", []string{"#"}}, // Catch all
	}

	for _, q := range testQueues {
		if err := c.DeclareQueue(q.name, q.routingKeys...); err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.name, err)
		}
		if err := c.DeclareDeadLetterQueue(q.name); err != nil {
			return fmt.Errorf("failed to declare DLQ for %s: %w", q.name, err)
		}
	}

	return nil
}

// Ack acknowledges a message.
func (c *RabbitMQContainer) Ack(delivery amqp.Delivery) error {
	return delivery.Ack(false)
}

// Nack negatively acknowledges a message.
func (c *RabbitMQContainer) Nack(delivery amqp.Delivery, requeue bool) error {
	return delivery.Nack(false, requeue)
}

// Reject rejects a message.
func (c *RabbitMQContainer) Reject(delivery amqp.Delivery, requeue bool) error {
	return delivery.Reject(requeue)
}
