// Package cache provides caching infrastructure for the Customer service.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/ports"
)

// RedisCacheConfig holds Redis cache configuration.
type RedisCacheConfig struct {
	Address      string
	Password     string
	DB           int
	DefaultTTL   time.Duration
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
	KeyPrefix    string
}

// DefaultRedisCacheConfig returns default Redis cache configuration.
func DefaultRedisCacheConfig() RedisCacheConfig {
	return RedisCacheConfig{
		Address:      "localhost:6379",
		Password:     "",
		DB:           0,
		DefaultTTL:   15 * time.Minute,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 5,
		KeyPrefix:    "customer:",
	}
}

// RedisCache implements ports.CacheService using Redis.
type RedisCache struct {
	client     *redis.Client
	config     RedisCacheConfig
}

// NewRedisCache creates a new Redis cache service.
func NewRedisCache(config RedisCacheConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Password:     config.Password,
		DB:           config.DB,
		MaxRetries:   config.MaxRetries,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		config: config,
	}, nil
}

// buildKey builds a cache key with the prefix.
func (c *RedisCache) buildKey(key string) string {
	return c.config.KeyPrefix + key
}

// Get retrieves a value from cache.
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := c.buildKey(key)

	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ports.ErrCacheMiss
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	return data, nil
}

// Set stores a value in cache with optional TTL.
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	fullKey := c.buildKey(key)

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	if err := c.client.Set(ctx, fullKey, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a value from cache.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.buildKey(key)

	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// DeletePattern removes all keys matching a pattern.
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.buildKey(pattern)

	iter := c.client.Scan(ctx, 0, fullPattern, 100).Iterator()
	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// Exists checks if a key exists in cache.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.buildKey(key)

	result, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return result > 0, nil
}

// GetOrSet gets a value or sets it if not present.
func (c *RedisCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, getter func() ([]byte, error)) ([]byte, error) {
	// Try to get from cache
	data, err := c.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	if err != ports.ErrCacheMiss {
		return nil, err
	}

	// Fetch from source
	data, err = getter()
	if err != nil {
		return nil, err
	}

	// Store in cache (ignore errors - cache is optional)
	_ = c.Set(ctx, key, data, ttl)

	return data, nil
}

// Invalidate invalidates cache for a specific entity.
func (c *RedisCache) Invalidate(ctx context.Context, entityType string, entityID uuid.UUID) error {
	key := fmt.Sprintf("%s:%s", entityType, entityID.String())
	return c.Delete(ctx, key)
}

// InvalidateByTenant invalidates all cache for a tenant.
func (c *RedisCache) InvalidateByTenant(ctx context.Context, tenantID uuid.UUID) error {
	pattern := fmt.Sprintf("tenant:%s:*", tenantID.String())
	return c.DeletePattern(ctx, pattern)
}

// TTL returns the remaining time-to-live for a cached entry.
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := c.buildKey(key)

	ttl, err := c.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// Refresh extends the TTL of a cached entry.
func (c *RedisCache) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := c.buildKey(key)

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	success, err := c.client.Expire(ctx, fullKey, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to refresh TTL: %w", err)
	}

	if !success {
		return ports.ErrCacheMiss
	}

	return nil
}

// Stats returns cache statistics.
func (c *RedisCache) Stats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.client.Info(ctx, "stats", "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	return map[string]interface{}{
		"info": info,
	}, nil
}

// Ping checks if the cache is available.
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// FlushDB flushes the current database (use with caution).
func (c *RedisCache) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Ensure RedisCache implements ports.CacheService
var _ ports.CacheService = (*RedisCache)(nil)

// ============================================================================
// In-Memory Cache (for testing and development)
// ============================================================================

// InMemoryCache implements ports.CacheService using in-memory storage.
type InMemoryCache struct {
	data       map[string]cacheEntry
	defaultTTL time.Duration
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewInMemoryCache creates a new in-memory cache.
func NewInMemoryCache(defaultTTL time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		data:       make(map[string]cacheEntry),
		defaultTTL: defaultTTL,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.expiresAt) {
				delete(c.data, key)
			}
		}
	}
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	entry, ok := c.data[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, ports.ErrCacheMiss
	}

	return entry.value, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	c.data[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}

func (c *InMemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	// In-memory cache doesn't support pattern-based invalidation efficiently
	// For simplicity, clear all cache
	c.data = make(map[string]cacheEntry)
	return nil
}

func (c *InMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	entry, ok := c.data[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return false, nil
	}
	return true, nil
}

func (c *InMemoryCache) GetOrSet(ctx context.Context, key string, ttl time.Duration, getter func() ([]byte, error)) ([]byte, error) {
	data, err := c.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	data, err = getter()
	if err != nil {
		return nil, err
	}

	_ = c.Set(ctx, key, data, ttl)
	return data, nil
}

func (c *InMemoryCache) Invalidate(ctx context.Context, entityType string, entityID uuid.UUID) error {
	key := fmt.Sprintf("%s:%s", entityType, entityID.String())
	return c.Delete(ctx, key)
}

func (c *InMemoryCache) InvalidateByTenant(ctx context.Context, tenantID uuid.UUID) error {
	// In-memory cache doesn't support pattern-based invalidation efficiently
	// For simplicity, clear all cache
	c.data = make(map[string]cacheEntry)
	return nil
}

func (c *InMemoryCache) Close() error {
	c.data = nil
	return nil
}

// Ensure InMemoryCache implements ports.CacheService
var _ ports.CacheService = (*InMemoryCache)(nil)
