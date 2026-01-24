// Package redis contains Redis-based cache implementations.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService implements the application ports.CacheService interface using Redis.
type CacheService struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
}

// CacheConfig holds configuration for the cache service.
type CacheConfig struct {
	Prefix     string
	DefaultTTL time.Duration
}

// NewCacheService creates a new Redis cache service.
func NewCacheService(client *redis.Client, config CacheConfig) *CacheService {
	if config.Prefix == "" {
		config.Prefix = "iam:"
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 15 * time.Minute
	}

	return &CacheService{
		client:     client,
		prefix:     config.Prefix,
		defaultTTL: config.DefaultTTL,
	}
}

// Get retrieves a value from the cache.
func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := s.prefix + key

	data, err := s.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Set stores a value in the cache with the default TTL.
func (s *CacheService) Set(ctx context.Context, key string, value interface{}) error {
	return s.SetWithTTL(ctx, key, value, s.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (s *CacheService) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := s.prefix + key

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := s.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

// Delete removes a value from the cache.
func (s *CacheService) Delete(ctx context.Context, key string) error {
	fullKey := s.prefix + key

	if err := s.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// DeleteByPattern removes all values matching a pattern from the cache.
func (s *CacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	fullPattern := s.prefix + pattern

	iter := s.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := s.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	return nil
}

// Exists checks if a key exists in the cache.
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := s.prefix + key

	result, err := s.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence: %w", err)
	}

	return result > 0, nil
}

// TTL returns the remaining TTL of a key.
func (s *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := s.prefix + key

	ttl, err := s.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get cache TTL: %w", err)
	}

	return ttl, nil
}

// Expire sets a new TTL for an existing key.
func (s *CacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := s.prefix + key

	if err := s.client.Expire(ctx, fullKey, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache expiry: %w", err)
	}

	return nil
}

// Increment atomically increments a counter.
func (s *CacheService) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := s.prefix + key

	result, err := s.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment cache value: %w", err)
	}

	return result, nil
}

// IncrementWithTTL atomically increments a counter and sets TTL if new.
func (s *CacheService) IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	fullKey := s.prefix + key

	// Use a transaction to increment and set TTL atomically
	pipe := s.client.TxPipeline()
	incr := pipe.Incr(ctx, fullKey)
	pipe.Expire(ctx, fullKey, ttl)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment cache value with TTL: %w", err)
	}

	return incr.Val(), nil
}

// Flush clears all keys with the service prefix.
func (s *CacheService) Flush(ctx context.Context) error {
	return s.DeleteByPattern(ctx, "*")
}

// Ping checks if Redis is available.
func (s *CacheService) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (s *CacheService) Close() error {
	return s.client.Close()
}

// SetNX sets a value only if it doesn't exist (used for locks).
func (s *CacheService) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	fullKey := s.prefix + key

	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal cache value: %w", err)
	}

	result, err := s.client.SetNX(ctx, fullKey, data, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set cache value: %w", err)
	}

	return result, nil
}

// GetSet atomically sets a new value and returns the old value.
func (s *CacheService) GetSet(ctx context.Context, key string, value interface{}, dest interface{}) error {
	fullKey := s.prefix + key

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal new cache value: %w", err)
	}

	oldData, err := s.client.GetSet(ctx, fullKey, data).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get-set cache value: %w", err)
	}

	if err := json.Unmarshal(oldData, dest); err != nil {
		return fmt.Errorf("failed to unmarshal old cache value: %w", err)
	}

	return nil
}

// Cache errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)
