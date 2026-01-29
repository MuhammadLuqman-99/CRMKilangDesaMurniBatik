// Package database contains Redis integration tests.
package database

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/kilang-desa-murni/crm/pkg/testing/containers"
	"github.com/kilang-desa-murni/crm/pkg/testing/fixtures"
	"github.com/kilang-desa-murni/crm/pkg/testing/helpers"
)

var (
	testRedis     *containers.RedisContainer
	testCtx       context.Context
	testCtxCancel context.CancelFunc
)

// TestMain sets up and tears down the Redis container.
func TestMain(m *testing.M) {
	// Skip if in short mode
	if testing.Short() {
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup Redis container
	var err error
	testRedis, err = containers.NewRedisContainer(ctx, containers.DefaultRedisConfig())
	if err != nil {
		panic("failed to create Redis container: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if testRedis != nil {
		testRedis.Close()
	}

	os.Exit(code)
}

func setupTest(t *testing.T) {
	t.Helper()
	helpers.SkipIfShort(t)
	testCtx, testCtxCancel = helpers.DefaultTestContext()
	// Clean database before each test
	testRedis.FlushDB(testCtx)
}

func cleanupTest(t *testing.T) {
	t.Helper()
	if testCtxCancel != nil {
		testCtxCancel()
	}
}

// ============================================================================
// String Operations Tests
// ============================================================================

func TestRedis_StringOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("set and get string value", func(t *testing.T) {
		key := "test:string:1"
		value := "test-value"

		err := testRedis.Set(testCtx, key, value, 0)
		helpers.AssertNoError(t, err)

		result, err := testRedis.Get(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, value, result)
	})

	t.Run("set with expiration", func(t *testing.T) {
		key := "test:expiring:1"
		value := "expiring-value"

		err := testRedis.Set(testCtx, key, value, 1*time.Second)
		helpers.AssertNoError(t, err)

		// Value should exist
		result, err := testRedis.Get(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, value, result)

		// Wait for expiration
		time.Sleep(1100 * time.Millisecond)

		// Value should not exist
		_, err = testRedis.Get(testCtx, key)
		helpers.AssertError(t, err)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		_, err := testRedis.Get(testCtx, "non-existent-key")
		helpers.AssertError(t, err)
	})

	t.Run("delete key", func(t *testing.T) {
		key := "test:delete:1"
		err := testRedis.Set(testCtx, key, "value", 0)
		helpers.AssertNoError(t, err)

		err = testRedis.Del(testCtx, key)
		helpers.AssertNoError(t, err)

		_, err = testRedis.Get(testCtx, key)
		helpers.AssertError(t, err)
	})

	t.Run("delete multiple keys", func(t *testing.T) {
		keys := []string{"test:del:1", "test:del:2", "test:del:3"}
		for _, key := range keys {
			err := testRedis.Set(testCtx, key, "value", 0)
			helpers.AssertNoError(t, err)
		}

		err := testRedis.Del(testCtx, keys...)
		helpers.AssertNoError(t, err)

		for _, key := range keys {
			exists, _ := testRedis.Exists(testCtx, key)
			helpers.AssertEqual(t, int64(0), exists)
		}
	})
}

func TestRedis_JSONOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("store and retrieve JSON object", func(t *testing.T) {
		type Session struct {
			UserID   string `json:"user_id"`
			TenantID string `json:"tenant_id"`
			Email    string `json:"email"`
		}

		session := Session{
			UserID:   fixtures.TestIDs.UserID1.String(),
			TenantID: fixtures.TestIDs.TenantID1.String(),
			Email:    "test@example.com",
		}

		key := "session:" + fixtures.TestIDs.UserID1.String()
		jsonData, _ := json.Marshal(session)

		err := testRedis.Set(testCtx, key, jsonData, 0)
		helpers.AssertNoError(t, err)

		result, err := testRedis.Get(testCtx, key)
		helpers.AssertNoError(t, err)

		var retrieved Session
		err = json.Unmarshal([]byte(result), &retrieved)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, session.UserID, retrieved.UserID)
		helpers.AssertEqual(t, session.TenantID, retrieved.TenantID)
	})
}

// ============================================================================
// Hash Operations Tests
// ============================================================================

func TestRedis_HashOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("set and get hash field", func(t *testing.T) {
		key := "user:" + fixtures.TestIDs.UserID1.String()

		err := testRedis.HSet(testCtx, key,
			"email", "user@example.com",
			"name", "John Doe",
			"status", "active",
		)
		helpers.AssertNoError(t, err)

		email, err := testRedis.HGet(testCtx, key, "email")
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "user@example.com", email)

		name, err := testRedis.HGet(testCtx, key, "name")
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "John Doe", name)
	})

	t.Run("get all hash fields", func(t *testing.T) {
		key := "config:" + fixtures.TestIDs.TenantID1.String()

		err := testRedis.HSet(testCtx, key,
			"timezone", "Asia/Jakarta",
			"language", "id",
			"currency", "IDR",
		)
		helpers.AssertNoError(t, err)

		all, err := testRedis.HGetAll(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 3, len(all))
		helpers.AssertEqual(t, "Asia/Jakarta", all["timezone"])
		helpers.AssertEqual(t, "id", all["language"])
		helpers.AssertEqual(t, "IDR", all["currency"])
	})

	t.Run("get non-existent hash field", func(t *testing.T) {
		key := "hash:nonexistent"
		_, err := testRedis.HGet(testCtx, key, "field")
		helpers.AssertError(t, err)
	})
}

// ============================================================================
// Set Operations Tests
// ============================================================================

func TestRedis_SetOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("add and check set members", func(t *testing.T) {
		key := "user:roles:" + fixtures.TestIDs.UserID1.String()

		err := testRedis.SAdd(testCtx, key, "admin", "user", "manager")
		helpers.AssertNoError(t, err)

		// Check membership
		isMember, err := testRedis.SIsMember(testCtx, key, "admin")
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, isMember)

		isMember, err = testRedis.SIsMember(testCtx, key, "superadmin")
		helpers.AssertNoError(t, err)
		helpers.AssertFalse(t, isMember)
	})

	t.Run("get all set members", func(t *testing.T) {
		key := "tenant:features:" + fixtures.TestIDs.TenantID1.String()

		err := testRedis.SAdd(testCtx, key, "crm", "sales", "notifications")
		helpers.AssertNoError(t, err)

		members, err := testRedis.SMembers(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, members, 3)
	})
}

// ============================================================================
// List Operations Tests
// ============================================================================

func TestRedis_ListOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("push and pop from list", func(t *testing.T) {
		key := "queue:events"

		// Push to list
		err := testRedis.RPush(testCtx, key, "event1", "event2", "event3")
		helpers.AssertNoError(t, err)

		// Get length
		length, err := testRedis.LLen(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(3), length)

		// Get range
		items, err := testRedis.LRange(testCtx, key, 0, -1)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, items, 3)
		helpers.AssertEqual(t, "event1", items[0])
		helpers.AssertEqual(t, "event3", items[2])
	})

	t.Run("lpush adds to head", func(t *testing.T) {
		key := "recent:activities"

		err := testRedis.LPush(testCtx, key, "activity1")
		helpers.AssertNoError(t, err)

		err = testRedis.LPush(testCtx, key, "activity2")
		helpers.AssertNoError(t, err)

		items, err := testRedis.LRange(testCtx, key, 0, -1)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "activity2", items[0]) // Most recent first
	})
}

// ============================================================================
// Counter Operations Tests
// ============================================================================

func TestRedis_CounterOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("increment counter", func(t *testing.T) {
		key := "counter:pageviews"

		// Increment
		val, err := testRedis.Incr(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(1), val)

		val, err = testRedis.Incr(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(2), val)
	})

	t.Run("increment by value", func(t *testing.T) {
		key := "counter:score"

		val, err := testRedis.IncrBy(testCtx, key, 10)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(10), val)

		val, err = testRedis.IncrBy(testCtx, key, 5)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(15), val)

		// Decrement by using negative value
		val, err = testRedis.IncrBy(testCtx, key, -3)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(12), val)
	})
}

// ============================================================================
// TTL Operations Tests
// ============================================================================

func TestRedis_TTLOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("set and check TTL", func(t *testing.T) {
		key := "ttl:test"
		err := testRedis.Set(testCtx, key, "value", 60*time.Second)
		helpers.AssertNoError(t, err)

		ttl, err := testRedis.TTL(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, ttl > 50*time.Second && ttl <= 60*time.Second)
	})

	t.Run("expire existing key", func(t *testing.T) {
		key := "expire:test"
		err := testRedis.Set(testCtx, key, "value", 0)
		helpers.AssertNoError(t, err)

		// Set expiration
		ok, err := testRedis.Expire(testCtx, key, 30*time.Second)
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, ok)

		ttl, err := testRedis.TTL(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, ttl > 25*time.Second && ttl <= 30*time.Second)
	})

	t.Run("key without TTL", func(t *testing.T) {
		key := "no:ttl"
		err := testRedis.Set(testCtx, key, "value", 0)
		helpers.AssertNoError(t, err)

		ttl, err := testRedis.TTL(testCtx, key)
		helpers.AssertNoError(t, err)
		// TTL returns -1 for keys without expiry
		helpers.AssertEqual(t, time.Duration(-1), ttl)
	})
}

// ============================================================================
// Key Pattern Matching Tests
// ============================================================================

func TestRedis_KeyPatternMatching(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("find keys by pattern", func(t *testing.T) {
		// Create test keys
		tenantID := fixtures.TestIDs.TenantID1.String()
		prefix := "cache:" + tenantID + ":"

		testRedis.Set(testCtx, prefix+"customer:1", "data1", 0)
		testRedis.Set(testCtx, prefix+"customer:2", "data2", 0)
		testRedis.Set(testCtx, prefix+"lead:1", "data3", 0)

		// Find all keys for tenant
		keys, err := testRedis.Keys(testCtx, prefix+"*")
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, keys, 3)

		// Find only customer keys
		keys, err = testRedis.Keys(testCtx, prefix+"customer:*")
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, keys, 2)
	})
}

// ============================================================================
// Pub/Sub Tests
// ============================================================================

func TestRedis_PubSub(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("publish and subscribe", func(t *testing.T) {
		channel := "notifications"
		message := "Hello, World!"

		// Subscribe
		pubsub := testRedis.Subscribe(testCtx, channel)
		defer pubsub.Close()

		// Wait for subscription to be ready
		_, err := pubsub.Receive(testCtx)
		helpers.AssertNoError(t, err)

		// Publish in goroutine
		go func() {
			time.Sleep(100 * time.Millisecond)
			testRedis.Publish(testCtx, channel, message)
		}()

		// Receive message
		ch := pubsub.Channel()
		select {
		case msg := <-ch:
			helpers.AssertEqual(t, message, msg.Payload)
			helpers.AssertEqual(t, channel, msg.Channel)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for pub/sub message")
		}
	})
}

// ============================================================================
// Pipeline Tests
// ============================================================================

func TestRedis_Pipeline(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("execute pipeline commands", func(t *testing.T) {
		pipe := testRedis.Pipeline()

		// Queue commands
		pipe.Set(testCtx, "pipe:key1", "value1", 0)
		pipe.Set(testCtx, "pipe:key2", "value2", 0)
		pipe.Set(testCtx, "pipe:key3", "value3", 0)
		getCmd := pipe.Get(testCtx, "pipe:key1")

		// Execute pipeline
		_, err := pipe.Exec(testCtx)
		helpers.AssertNoError(t, err)

		// Verify
		val, err := getCmd.Result()
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "value1", val)
	})

	t.Run("transactional pipeline", func(t *testing.T) {
		pipe := testRedis.TxPipeline()

		// Queue commands in transaction
		pipe.Set(testCtx, "tx:key1", "value1", 0)
		pipe.Incr(testCtx, "tx:counter")
		pipe.Incr(testCtx, "tx:counter")

		// Execute
		_, err := pipe.Exec(testCtx)
		helpers.AssertNoError(t, err)

		// Verify counter
		val, err := testRedis.Get(testCtx, "tx:counter")
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "2", val)
	})
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestRedis_ConcurrentAccess(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("concurrent counter increments", func(t *testing.T) {
		key := "concurrent:counter"
		const numGoroutines = 100
		const incrementsPerGoroutine = 10

		var wg sync.WaitGroup
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < incrementsPerGoroutine; j++ {
					testRedis.Incr(testCtx, key)
				}
			}()
		}

		wg.Wait()

		// Verify final count
		val, err := testRedis.Get(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "1000", val) // 100 * 10 = 1000
	})

	t.Run("concurrent set operations", func(t *testing.T) {
		const numGoroutines = 50
		var wg sync.WaitGroup
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				key := helpers.GenerateRandomString(20)
				err := testRedis.Set(testCtx, key, "value", 0)
				errChan <- err
			}(i)
		}

		wg.Wait()
		close(errChan)

		// All operations should succeed
		for err := range errChan {
			helpers.AssertNoError(t, err)
		}
	})
}

// ============================================================================
// Cache Pattern Tests
// ============================================================================

func TestRedis_CachePatterns(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("cache aside pattern", func(t *testing.T) {
		key := "cache:customer:" + fixtures.TestIDs.CustomerID1.String()

		// Simulate cache miss
		_, err := testRedis.Get(testCtx, key)
		helpers.AssertError(t, err) // Cache miss

		// Load from "database" and cache
		customerData := `{"id":"` + fixtures.TestIDs.CustomerID1.String() + `","name":"Test Customer"}`
		err = testRedis.Set(testCtx, key, customerData, 5*time.Minute)
		helpers.AssertNoError(t, err)

		// Cache hit
		result, err := testRedis.Get(testCtx, key)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, customerData, result)
	})

	t.Run("cache invalidation pattern", func(t *testing.T) {
		tenantID := fixtures.TestIDs.TenantID1.String()
		keys := []string{
			"cache:" + tenantID + ":customer:1",
			"cache:" + tenantID + ":customer:2",
			"cache:" + tenantID + ":lead:1",
		}

		// Populate cache
		for _, key := range keys {
			testRedis.Set(testCtx, key, "data", 0)
		}

		// Invalidate all tenant cache
		pattern := "cache:" + tenantID + ":*"
		keysToDelete, _ := testRedis.Keys(testCtx, pattern)
		if len(keysToDelete) > 0 {
			testRedis.Del(testCtx, keysToDelete...)
		}

		// Verify all deleted
		for _, key := range keys {
			exists, _ := testRedis.Exists(testCtx, key)
			helpers.AssertEqual(t, int64(0), exists)
		}
	})
}

// ============================================================================
// Rate Limiting Tests
// ============================================================================

func TestRedis_RateLimiting(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("simple rate limiter", func(t *testing.T) {
		key := "ratelimit:" + fixtures.TestIDs.UserID1.String()
		limit := int64(5)
		window := 10 * time.Second

		// Simulate requests
		for i := 0; i < 5; i++ {
			count, _ := testRedis.Incr(testCtx, key)
			if i == 0 {
				testRedis.Expire(testCtx, key, window)
			}
			helpers.AssertTrue(t, count <= limit, "Request should be allowed")
		}

		// 6th request should exceed limit
		count, _ := testRedis.Incr(testCtx, key)
		helpers.AssertTrue(t, count > limit, "Request should exceed limit")
	})
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkRedis_Set(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := helpers.GenerateRandomString(20)
		_ = testRedis.Set(ctx, key, "value", 0)
	}
}

func BenchmarkRedis_Get(b *testing.B) {
	ctx := context.Background()
	key := "benchmark:get"
	testRedis.Set(ctx, key, "value", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = testRedis.Get(ctx, key)
	}
}

func BenchmarkRedis_Pipeline(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipe := testRedis.Pipeline()
		for j := 0; j < 10; j++ {
			pipe.Set(ctx, helpers.GenerateRandomString(20), "value", 0)
		}
		_, _ = pipe.Exec(ctx)
	}
}
