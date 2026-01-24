// Package redis contains Redis-based cache implementations.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ClientConfig holds Redis client configuration.
type ClientConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultClientConfig returns default Redis client configuration.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// NewClient creates a new Redis client.
func NewClient(config ClientConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// NewClusterClient creates a new Redis cluster client.
func NewClusterClient(addrs []string, password string) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	return client, nil
}

// SentinelConfig holds Redis Sentinel configuration.
type SentinelConfig struct {
	MasterName       string
	SentinelAddrs    []string
	Password         string
	SentinelPassword string
	DB               int
	PoolSize         int
}

// NewSentinelClient creates a new Redis client with Sentinel support.
func NewSentinelClient(config SentinelConfig) (*redis.Client, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       config.MasterName,
		SentinelAddrs:    config.SentinelAddrs,
		Password:         config.Password,
		SentinelPassword: config.SentinelPassword,
		DB:               config.DB,
		PoolSize:         config.PoolSize,
		MinIdleConns:     2,
		MaxRetries:       3,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      3 * time.Second,
		WriteTimeout:     3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis via Sentinel: %w", err)
	}

	return client, nil
}

// HealthChecker provides Redis health checking functionality.
type HealthChecker struct {
	client *redis.Client
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(client *redis.Client) *HealthChecker {
	return &HealthChecker{client: client}
}

// Check performs a health check on the Redis connection.
func (h *HealthChecker) Check(ctx context.Context) error {
	return h.client.Ping(ctx).Err()
}

// Info returns Redis server information.
func (h *HealthChecker) Info(ctx context.Context) (map[string]string, error) {
	info, err := h.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse info string into map (simplified)
	result := map[string]string{
		"raw_info": info,
	}

	return result, nil
}

// PoolStats returns connection pool statistics.
func (h *HealthChecker) PoolStats() *redis.PoolStats {
	return h.client.PoolStats()
}
