// Package containers provides test container implementations for integration testing.
package containers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisContainer represents a Redis test container configuration.
type RedisContainer struct {
	Host     string
	Port     string
	Password string
	DB       int
	Client   *redis.Client
}

// RedisContainerConfig holds configuration for Redis container.
type RedisContainerConfig struct {
	Password string
	DB       int
}

// DefaultRedisConfig returns default Redis configuration.
func DefaultRedisConfig() RedisContainerConfig {
	return RedisContainerConfig{
		Password: "crm_redis_password",
		DB:       0,
	}
}

// NewRedisContainer creates a new Redis container for testing.
// For integration tests, this connects to the docker-compose Redis instance.
func NewRedisContainer(ctx context.Context, cfg RedisContainerConfig) (*RedisContainer, error) {
	container := &RedisContainer{
		Host:     getEnvOrDefault("TEST_REDIS_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_REDIS_PORT", "6379"),
		Password: getEnvOrDefault("TEST_REDIS_PASSWORD", cfg.Password),
		DB:       cfg.DB,
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", container.Host, container.Port),
		Password:     container.Password,
		DB:           container.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
		PoolTimeout:  4 * time.Second,
	})

	// Ping to verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	container.Client = client

	return container, nil
}

// GetClient returns the Redis client.
func (c *RedisContainer) GetClient() *redis.Client {
	return c.Client
}

// FlushDB flushes the current database.
func (c *RedisContainer) FlushDB(ctx context.Context) error {
	return c.Client.FlushDB(ctx).Err()
}

// FlushAll flushes all databases.
func (c *RedisContainer) FlushAll(ctx context.Context) error {
	return c.Client.FlushAll(ctx).Err()
}

// Set sets a key-value pair.
func (c *RedisContainer) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key.
func (c *RedisContainer) Get(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// Del deletes keys.
func (c *RedisContainer) Del(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist.
func (c *RedisContainer) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.Client.Exists(ctx, keys...).Result()
}

// Keys returns keys matching a pattern.
func (c *RedisContainer) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.Client.Keys(ctx, pattern).Result()
}

// HSet sets hash field values.
func (c *RedisContainer) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.Client.HSet(ctx, key, values...).Err()
}

// HGet gets a hash field value.
func (c *RedisContainer) HGet(ctx context.Context, key, field string) (string, error) {
	return c.Client.HGet(ctx, key, field).Result()
}

// HGetAll gets all hash field values.
func (c *RedisContainer) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.Client.HGetAll(ctx, key).Result()
}

// SAdd adds members to a set.
func (c *RedisContainer) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.Client.SAdd(ctx, key, members...).Err()
}

// SMembers returns all members of a set.
func (c *RedisContainer) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.Client.SMembers(ctx, key).Result()
}

// SIsMember checks if a member is in a set.
func (c *RedisContainer) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.Client.SIsMember(ctx, key, member).Result()
}

// LPush pushes values to the head of a list.
func (c *RedisContainer) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.Client.LPush(ctx, key, values...).Err()
}

// RPush pushes values to the tail of a list.
func (c *RedisContainer) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.Client.RPush(ctx, key, values...).Err()
}

// LRange returns a range of elements from a list.
func (c *RedisContainer) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.Client.LRange(ctx, key, start, stop).Result()
}

// LLen returns the length of a list.
func (c *RedisContainer) LLen(ctx context.Context, key string) (int64, error) {
	return c.Client.LLen(ctx, key).Result()
}

// Incr increments a key.
func (c *RedisContainer) Incr(ctx context.Context, key string) (int64, error) {
	return c.Client.Incr(ctx, key).Result()
}

// IncrBy increments a key by a value.
func (c *RedisContainer) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.Client.IncrBy(ctx, key, value).Result()
}

// Expire sets a timeout on a key.
func (c *RedisContainer) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.Client.Expire(ctx, key, expiration).Result()
}

// TTL returns the time-to-live of a key.
func (c *RedisContainer) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}

// Publish publishes a message to a channel.
func (c *RedisContainer) Publish(ctx context.Context, channel string, message interface{}) error {
	return c.Client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels.
func (c *RedisContainer) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.Client.Subscribe(ctx, channels...)
}

// Pipeline returns a pipeline for batch operations.
func (c *RedisContainer) Pipeline() redis.Pipeliner {
	return c.Client.Pipeline()
}

// TxPipeline returns a transactional pipeline.
func (c *RedisContainer) TxPipeline() redis.Pipeliner {
	return c.Client.TxPipeline()
}

// Close closes the Redis client connection.
func (c *RedisContainer) Close() error {
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}

// WaitForReady waits for Redis to be ready.
func (c *RedisContainer) WaitForReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for Redis to be ready")
		case <-ticker.C:
			if err := c.Client.Ping(ctx).Err(); err == nil {
				return nil
			}
		}
	}
}

// SetupTestData sets up test data for cache testing.
func (c *RedisContainer) SetupTestData(ctx context.Context) error {
	// Set some test data
	testData := map[string]interface{}{
		"test:key1":     "value1",
		"test:key2":     "value2",
		"test:counter":  0,
		"session:test1": `{"user_id":"123","tenant_id":"456"}`,
	}

	for key, value := range testData {
		if err := c.Client.Set(ctx, key, value, 0).Err(); err != nil {
			return fmt.Errorf("failed to set test data for key %s: %w", key, err)
		}
	}

	return nil
}

// GetInfo returns Redis server info.
func (c *RedisContainer) GetInfo(ctx context.Context) (string, error) {
	return c.Client.Info(ctx).Result()
}

// DBSize returns the number of keys in the current database.
func (c *RedisContainer) DBSize(ctx context.Context) (int64, error) {
	return c.Client.DBSize(ctx).Result()
}
