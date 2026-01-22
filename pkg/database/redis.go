// Package database provides database connection utilities for the CRM application.
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/logger"
)

// RedisClient wraps the redis.Client and provides cache operations.
type RedisClient struct {
	client *redis.Client
	config *config.RedisConfig
	log    *logger.Logger
}

// NewRedis creates a new Redis client connection.
func NewRedis(cfg *config.RedisConfig, log *logger.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Info().
		Str("addr", cfg.Addr()).
		Int("db", cfg.DB).
		Msg("Connected to Redis")

	return &RedisClient{
		client: client,
		config: cfg,
		log:    log,
	}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	r.log.Info().Msg("Closing Redis connection")
	return r.client.Close()
}

// Health checks the Redis connection health.
func (r *RedisClient) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Client returns the underlying redis.Client.
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Set sets a key-value pair with an expiration time.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value by key and unmarshals it into the target.
func (r *RedisClient) Get(ctx context.Context, key string, target interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to get value: %w", err)
	}
	return json.Unmarshal(data, target)
}

// GetString retrieves a string value by key.
func (r *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("failed to get value: %w", err)
	}
	return val, nil
}

// Delete deletes one or more keys.
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists.
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return result > 0, nil
}

// Expire sets an expiration time on a key.
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL returns the remaining time to live of a key.
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Incr increments the integer value of a key by 1.
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy increments the integer value of a key by the given amount.
func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// Decr decrements the integer value of a key by 1.
func (r *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// SetNX sets a key-value pair only if the key does not exist.
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.SetNX(ctx, key, data, expiration).Result()
}

// HSet sets a hash field.
func (r *RedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.HSet(ctx, key, field, data).Err()
}

// HGet retrieves a hash field.
func (r *RedisClient) HGet(ctx context.Context, key, field string, target interface{}) error {
	data, err := r.client.HGet(ctx, key, field).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to get hash field: %w", err)
	}
	return json.Unmarshal(data, target)
}

// HGetAll retrieves all fields and values of a hash.
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel deletes one or more hash fields.
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// SAdd adds one or more members to a set.
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers returns all members of a set.
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SIsMember checks if a member is part of a set.
func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// SRem removes one or more members from a set.
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// Publish publishes a message to a channel.
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return r.client.Publish(ctx, channel, data).Err()
}

// Subscribe subscribes to one or more channels.
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// FlushDB deletes all keys in the current database.
func (r *RedisClient) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Keys returns all keys matching the pattern.
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Scan iterates over keys matching the pattern.
func (r *RedisClient) Scan(ctx context.Context, cursor uint64, pattern string, count int64) ([]string, uint64, error) {
	return r.client.Scan(ctx, cursor, pattern, count).Result()
}

// Pipeline creates a new pipeline.
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline creates a new transactional pipeline.
func (r *RedisClient) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// ErrKeyNotFound is returned when a key is not found in Redis.
var ErrKeyNotFound = fmt.Errorf("key not found")
