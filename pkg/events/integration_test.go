// Package events contains event bus integration tests.
package events

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/logger"
	"github.com/kilang-desa-murni/crm/pkg/testing/containers"
	"github.com/kilang-desa-murni/crm/pkg/testing/fixtures"
	"github.com/kilang-desa-murni/crm/pkg/testing/helpers"
)

var (
	testRabbitMQ  *containers.RabbitMQContainer
	testCtx       context.Context
	testCtxCancel context.CancelFunc
)

// TestMain sets up and tears down the RabbitMQ container.
func TestMain(m *testing.M) {
	// Skip if in short mode
	if testing.Short() {
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup RabbitMQ container
	var err error
	testRabbitMQ, err = containers.NewRabbitMQContainer(ctx, containers.DefaultRabbitMQConfig())
	if err != nil {
		panic("failed to create RabbitMQ container: " + err.Error())
	}

	// Setup test queues
	if err := testRabbitMQ.SetupTestQueues(); err != nil {
		panic("failed to setup test queues: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if testRabbitMQ != nil {
		testRabbitMQ.Cleanup()
		testRabbitMQ.Close()
	}

	os.Exit(code)
}

func setupTest(t *testing.T) {
	t.Helper()
	helpers.SkipIfShort(t)
	testCtx, testCtxCancel = helpers.DefaultTestContext()
}

func cleanupTest(t *testing.T) {
	t.Helper()
	if testCtxCancel != nil {
		testCtxCancel()
	}
}

// ============================================================================
// RabbitMQ Event Bus Integration Tests
// ============================================================================

func TestRabbitMQEventBus_Connect(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully connects to RabbitMQ", func(t *testing.T) {
		cfg := &config.RabbitMQConfig{
			URL:               testRabbitMQ.ConnectionURL(),
			Exchange:          "test.events",
			ExchangeType:      "topic",
			PrefetchCount:     10,
			ReconnectDelay:    5 * time.Second,
			MaxReconnectDelay: 60 * time.Second,
		}

		log := logger.New(logger.Config{Level: "debug"})
		bus, err := NewRabbitMQEventBus(cfg, log)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, bus)

		defer bus.Close()
	})

	t.Run("fails with invalid connection URL", func(t *testing.T) {
		cfg := &config.RabbitMQConfig{
			URL:               "amqp://invalid:invalid@localhost:99999/",
			Exchange:          "test.events",
			ExchangeType:      "topic",
			PrefetchCount:     10,
			ReconnectDelay:    1 * time.Second,
			MaxReconnectDelay: 5 * time.Second,
		}

		log := logger.New(logger.Config{Level: "debug"})
		_, err := NewRabbitMQEventBus(cfg, log)
		helpers.AssertError(t, err)
	})
}

func TestRabbitMQEventBus_Publish(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully publishes an event", func(t *testing.T) {
		// Create event
		event := createTestEvent("user.created", fixtures.TestIDs.UserID1.String())

		// Publish using the container
		eventBytes, err := json.Marshal(event)
		helpers.RequireNoError(t, err)

		err = testRabbitMQ.Publish(testCtx, "user.created", eventBytes)
		helpers.AssertNoError(t, err)

		// Consume and verify
		msg, err := testRabbitMQ.ConsumeOne(testCtx, "test.user.created", 5*time.Second)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, msg.Body)

		var receivedEvent map[string]interface{}
		err = json.Unmarshal(msg.Body, &receivedEvent)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "user.created", receivedEvent["type"])

		// Acknowledge the message
		msg.Ack(false)
	})

	t.Run("publishes event with headers", func(t *testing.T) {
		event := createTestEvent("customer.created", fixtures.TestIDs.CustomerID1.String())
		eventBytes, err := json.Marshal(event)
		helpers.RequireNoError(t, err)

		headers := amqp.Table{
			"tenant_id":    fixtures.TestIDs.TenantID1.String(),
			"aggregate_id": fixtures.TestIDs.CustomerID1.String(),
			"version":      int32(1),
		}

		err = testRabbitMQ.PublishWithHeaders(testCtx, "customer.created", eventBytes, headers)
		helpers.AssertNoError(t, err)

		msg, err := testRabbitMQ.ConsumeOne(testCtx, "test.customer.created", 5*time.Second)
		helpers.AssertNoError(t, err)

		// Verify headers
		helpers.AssertEqual(t, fixtures.TestIDs.TenantID1.String(), msg.Headers["tenant_id"])
		msg.Ack(false)
	})
}

func TestRabbitMQEventBus_Subscribe(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully subscribes and receives events", func(t *testing.T) {
		// Start consuming
		msgs, err := testRabbitMQ.Consume("test.lead.created")
		helpers.RequireNoError(t, err)

		// Publish events
		for i := 0; i < 5; i++ {
			event := createTestEvent("lead.created", fixtures.TestIDs.LeadID1.String())
			eventBytes, _ := json.Marshal(event)
			err := testRabbitMQ.Publish(testCtx, "lead.created", eventBytes)
			helpers.RequireNoError(t, err)
		}

		// Receive events
		receivedCount := 0
		timeout := time.After(10 * time.Second)

		for receivedCount < 5 {
			select {
			case msg := <-msgs:
				receivedCount++
				msg.Ack(false)
			case <-timeout:
				t.Fatalf("Timeout waiting for messages. Received: %d/5", receivedCount)
			}
		}

		helpers.AssertEqual(t, 5, receivedCount)
	})

	t.Run("handles multiple concurrent consumers", func(t *testing.T) {
		const numConsumers = 3
		const numMessages = 30
		var received int32

		// Start multiple consumers
		var wg sync.WaitGroup
		for i := 0; i < numConsumers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				msgs, err := testRabbitMQ.Consume("test.all.events")
				if err != nil {
					return
				}

				timeout := time.After(15 * time.Second)
				for {
					select {
					case msg := <-msgs:
						atomic.AddInt32(&received, 1)
						msg.Ack(false)
					case <-timeout:
						return
					}
				}
			}()
		}

		// Small delay to ensure consumers are ready
		time.Sleep(500 * time.Millisecond)

		// Publish messages
		for i := 0; i < numMessages; i++ {
			event := createTestEvent("deal.won", fixtures.TestIDs.DealID1.String())
			eventBytes, _ := json.Marshal(event)
			err := testRabbitMQ.Publish(testCtx, "deal.won", eventBytes)
			helpers.RequireNoError(t, err)
		}

		// Wait for consumers with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(20 * time.Second):
		}

		helpers.AssertGreater(t, int(atomic.LoadInt32(&received)), 0)
	})
}

func TestRabbitMQEventBus_MessageAcknowledgment(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("ack removes message from queue", func(t *testing.T) {
		// Purge queue first
		testRabbitMQ.PurgeQueue("test.notification.email")

		// Publish event
		event := createTestEvent("notification.email.send", fixtures.TestIDs.NotificationID1.String())
		eventBytes, _ := json.Marshal(event)
		err := testRabbitMQ.Publish(testCtx, "notification.email.send", eventBytes)
		helpers.RequireNoError(t, err)

		// Consume and ack
		msg, err := testRabbitMQ.ConsumeOne(testCtx, "test.notification.email", 5*time.Second)
		helpers.AssertNoError(t, err)
		msg.Ack(false)

		// Verify queue is empty
		queueInfo, err := testRabbitMQ.GetQueueInfo("test.notification.email")
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 0, queueInfo.Messages)
	})

	t.Run("nack with requeue puts message back", func(t *testing.T) {
		// Purge queue first
		testRabbitMQ.PurgeQueue("test.notification.sms")

		// Publish event
		event := createTestEvent("notification.sms.send", fixtures.TestIDs.NotificationID1.String())
		eventBytes, _ := json.Marshal(event)
		err := testRabbitMQ.Publish(testCtx, "notification.sms.send", eventBytes)
		helpers.RequireNoError(t, err)

		// Consume and nack with requeue
		msg, err := testRabbitMQ.ConsumeOne(testCtx, "test.notification.sms", 5*time.Second)
		helpers.AssertNoError(t, err)
		msg.Nack(false, true)

		// Small delay for requeue
		time.Sleep(100 * time.Millisecond)

		// Verify message is back in queue
		queueInfo, err := testRabbitMQ.GetQueueInfo("test.notification.sms")
		helpers.AssertNoError(t, err)
		helpers.AssertGreaterOrEqual(t, queueInfo.Messages, 1)

		// Clean up
		testRabbitMQ.PurgeQueue("test.notification.sms")
	})

	t.Run("reject without requeue removes message", func(t *testing.T) {
		// Purge queue first
		testRabbitMQ.PurgeQueue("test.user.updated")

		// Publish event
		event := createTestEvent("user.updated", fixtures.TestIDs.UserID1.String())
		eventBytes, _ := json.Marshal(event)
		err := testRabbitMQ.Publish(testCtx, "user.updated", eventBytes)
		helpers.RequireNoError(t, err)

		// Consume and reject
		msg, err := testRabbitMQ.ConsumeOne(testCtx, "test.user.updated", 5*time.Second)
		helpers.AssertNoError(t, err)
		msg.Reject(false)

		// Small delay
		time.Sleep(100 * time.Millisecond)

		// Verify queue is empty
		queueInfo, err := testRabbitMQ.GetQueueInfo("test.user.updated")
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 0, queueInfo.Messages)
	})
}

func TestRabbitMQEventBus_QueueOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("purge queue removes all messages", func(t *testing.T) {
		queueName := "test.purge.queue"

		// Create queue
		err := testRabbitMQ.DeclareQueue(queueName, "test.purge.*")
		helpers.RequireNoError(t, err)

		// Publish multiple messages
		for i := 0; i < 10; i++ {
			event := createTestEvent("test.purge.event", fixtures.TestIDs.UserID1.String())
			eventBytes, _ := json.Marshal(event)
			err := testRabbitMQ.Publish(testCtx, "test.purge.event", eventBytes)
			helpers.RequireNoError(t, err)
		}

		// Small delay
		time.Sleep(100 * time.Millisecond)

		// Purge queue
		count, err := testRabbitMQ.PurgeQueue(queueName)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 10, count)

		// Verify queue is empty
		queueInfo, err := testRabbitMQ.GetQueueInfo(queueName)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 0, queueInfo.Messages)

		// Cleanup
		testRabbitMQ.DeleteQueue(queueName)
	})

	t.Run("get queue info returns correct counts", func(t *testing.T) {
		queueName := "test.info.queue"

		// Create queue
		err := testRabbitMQ.DeclareQueue(queueName, "test.info.*")
		helpers.RequireNoError(t, err)

		// Publish messages
		for i := 0; i < 5; i++ {
			event := createTestEvent("test.info.event", fixtures.TestIDs.UserID1.String())
			eventBytes, _ := json.Marshal(event)
			err := testRabbitMQ.Publish(testCtx, "test.info.event", eventBytes)
			helpers.RequireNoError(t, err)
		}

		// Small delay
		time.Sleep(100 * time.Millisecond)

		// Get queue info
		queueInfo, err := testRabbitMQ.GetQueueInfo(queueName)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 5, queueInfo.Messages)
		helpers.AssertEqual(t, queueName, queueInfo.Name)

		// Cleanup
		testRabbitMQ.PurgeQueue(queueName)
		testRabbitMQ.DeleteQueue(queueName)
	})
}

func TestRabbitMQEventBus_RoutingPatterns(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("topic exchange with wildcard routing", func(t *testing.T) {
		// Create queue with wildcard binding
		queueName := "test.wildcard.queue"
		err := testRabbitMQ.DeclareQueue(queueName, "user.*")
		helpers.RequireNoError(t, err)

		// Start consuming
		msgs, err := testRabbitMQ.Consume(queueName)
		helpers.RequireNoError(t, err)

		// Publish events with different routing keys
		routingKeys := []string{"user.created", "user.updated", "user.deleted"}
		for _, key := range routingKeys {
			event := createTestEvent(key, fixtures.TestIDs.UserID1.String())
			eventBytes, _ := json.Marshal(event)
			err := testRabbitMQ.Publish(testCtx, key, eventBytes)
			helpers.RequireNoError(t, err)
		}

		// Receive all events
		received := make([]string, 0)
		timeout := time.After(5 * time.Second)

		for len(received) < 3 {
			select {
			case msg := <-msgs:
				var event map[string]interface{}
				json.Unmarshal(msg.Body, &event)
				received = append(received, event["type"].(string))
				msg.Ack(false)
			case <-timeout:
				break
			}
		}

		helpers.AssertLen(t, received, 3)

		// Cleanup
		testRabbitMQ.DeleteQueue(queueName)
	})

	t.Run("hash wildcard matches all", func(t *testing.T) {
		// Purge the catch-all queue
		testRabbitMQ.PurgeQueue("test.all.events")

		// Publish events with different routing keys
		events := []string{"a.b.c", "x.y.z", "foo.bar.baz"}
		for _, key := range events {
			event := createTestEvent(key, fixtures.TestIDs.UserID1.String())
			eventBytes, _ := json.Marshal(event)
			err := testRabbitMQ.Publish(testCtx, key, eventBytes)
			helpers.RequireNoError(t, err)
		}

		// Small delay
		time.Sleep(200 * time.Millisecond)

		// All events should be in the catch-all queue
		queueInfo, err := testRabbitMQ.GetQueueInfo("test.all.events")
		helpers.AssertNoError(t, err)
		helpers.AssertGreaterOrEqual(t, queueInfo.Messages, 3)

		// Cleanup
		testRabbitMQ.PurgeQueue("test.all.events")
	})
}

func TestRabbitMQEventBus_ErrorHandling(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("handles malformed JSON gracefully", func(t *testing.T) {
		// Publish malformed JSON
		err := testRabbitMQ.Publish(testCtx, "test.malformed", []byte("not json"))
		helpers.AssertNoError(t, err) // Publishing should succeed

		// Consumer should be able to receive (but may fail to parse)
		// This tests the infrastructure handles it
	})

	t.Run("handles large messages", func(t *testing.T) {
		largeData := make([]byte, 1024*1024) // 1MB
		for i := range largeData {
			largeData[i] = 'x'
		}

		event := map[string]interface{}{
			"type": "large.event",
			"data": string(largeData),
		}
		eventBytes, err := json.Marshal(event)
		helpers.RequireNoError(t, err)

		err = testRabbitMQ.Publish(testCtx, "large.event", eventBytes)
		helpers.AssertNoError(t, err)
	})
}

func TestRabbitMQEventBus_BatchConsume(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("consumes batch of messages", func(t *testing.T) {
		queueName := "test.batch.queue"

		// Create queue
		err := testRabbitMQ.DeclareQueue(queueName, "batch.*")
		helpers.RequireNoError(t, err)

		// Publish multiple messages
		for i := 0; i < 10; i++ {
			event := createTestEvent("batch.event", fixtures.TestIDs.UserID1.String())
			eventBytes, _ := json.Marshal(event)
			err := testRabbitMQ.Publish(testCtx, "batch.event", eventBytes)
			helpers.RequireNoError(t, err)
		}

		// Consume batch
		messages, err := testRabbitMQ.ConsumeN(testCtx, queueName, 10, 10*time.Second)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, messages, 10)

		// Ack all
		for _, msg := range messages {
			msg.Ack(false)
		}

		// Cleanup
		testRabbitMQ.DeleteQueue(queueName)
	})
}

// ============================================================================
// Event Type Tests
// ============================================================================

func TestEventTypes(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	testCases := []struct {
		name       string
		eventType  string
		routingKey string
		queue      string
	}{
		{"user created event", "user.created", "user.created", "test.user.created"},
		{"customer created event", "customer.created", "customer.created", "test.customer.created"},
		{"lead created event", "lead.created", "lead.created", "test.lead.created"},
		{"lead converted event", "lead.converted", "lead.converted", "test.lead.converted"},
		{"deal won event", "deal.won", "deal.won", "test.deal.won"},
		{"deal lost event", "deal.lost", "deal.lost", "test.deal.lost"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Purge queue
			testRabbitMQ.PurgeQueue(tc.queue)

			// Publish
			event := createTestEvent(tc.eventType, fixtures.TestIDs.UserID1.String())
			eventBytes, _ := json.Marshal(event)
			err := testRabbitMQ.Publish(testCtx, tc.routingKey, eventBytes)
			helpers.AssertNoError(t, err)

			// Consume
			msg, err := testRabbitMQ.ConsumeOne(testCtx, tc.queue, 5*time.Second)
			helpers.AssertNoError(t, err)

			var received map[string]interface{}
			err = json.Unmarshal(msg.Body, &received)
			helpers.AssertNoError(t, err)
			helpers.AssertEqual(t, tc.eventType, received["type"])

			msg.Ack(false)
		})
	}
}

// ============================================================================
// Performance Tests
// ============================================================================

func BenchmarkRabbitMQ_Publish(b *testing.B) {
	ctx := context.Background()
	event := createTestEvent("benchmark.event", fixtures.TestIDs.UserID1.String())
	eventBytes, _ := json.Marshal(event)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = testRabbitMQ.Publish(ctx, "benchmark.event", eventBytes)
	}
}

func BenchmarkRabbitMQ_PublishAndConsume(b *testing.B) {
	ctx := context.Background()
	queueName := "benchmark.queue"

	// Create queue
	_ = testRabbitMQ.DeclareQueue(queueName, "benchmark.*")
	msgs, _ := testRabbitMQ.Consume(queueName)

	event := createTestEvent("benchmark.event", fixtures.TestIDs.UserID1.String())
	eventBytes, _ := json.Marshal(event)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = testRabbitMQ.Publish(ctx, "benchmark.event", eventBytes)
		msg := <-msgs
		msg.Ack(false)
	}

	// Cleanup
	testRabbitMQ.DeleteQueue(queueName)
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTestEvent(eventType, aggregateID string) map[string]interface{} {
	return map[string]interface{}{
		"id":             fixtures.NewUUID().String(),
		"type":           eventType,
		"tenant_id":      fixtures.TestIDs.TenantID1.String(),
		"aggregate_id":   aggregateID,
		"aggregate_type": "Test",
		"version":        1,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"data": map[string]interface{}{
			"test": "data",
		},
	}
}
