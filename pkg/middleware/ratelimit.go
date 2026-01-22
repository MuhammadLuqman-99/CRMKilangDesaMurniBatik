// Package middleware provides HTTP middleware utilities for the CRM application.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kilang-desa-murni/crm/pkg/database"
	"github.com/kilang-desa-murni/crm/pkg/errors"
	"github.com/kilang-desa-murni/crm/pkg/response"
)

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, int, int, time.Time, error)
}

// RateLimitConfig holds rate limiter configuration.
type RateLimitConfig struct {
	Requests int           // Number of requests allowed
	Window   time.Duration // Time window
	KeyFunc  func(*http.Request) string
}

// DefaultKeyFunc returns the default key function (uses IP address).
func DefaultKeyFunc(r *http.Request) string {
	return r.RemoteAddr
}

// TenantKeyFunc returns a key function that uses tenant ID.
func TenantKeyFunc(r *http.Request) string {
	tenantID := TenantIDFromContext(r.Context())
	if tenantID == "" {
		return r.RemoteAddr
	}
	return fmt.Sprintf("tenant:%s", tenantID)
}

// UserKeyFunc returns a key function that uses user ID.
func UserKeyFunc(r *http.Request) string {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		return r.RemoteAddr
	}
	return fmt.Sprintf("user:%s", userID)
}

// InMemoryRateLimiter implements rate limiting using in-memory storage.
type InMemoryRateLimiter struct {
	config RateLimitConfig
	mu     sync.RWMutex
	store  map[string]*bucket
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
func NewInMemoryRateLimiter(config RateLimitConfig) *InMemoryRateLimiter {
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}

	limiter := &InMemoryRateLimiter{
		config: config,
		store:  make(map[string]*bucket),
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request is allowed.
func (l *InMemoryRateLimiter) Allow(ctx context.Context, key string) (bool, int, int, time.Time, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	resetAt := now.Add(l.config.Window)

	b, exists := l.store[key]
	if !exists || now.Sub(b.lastReset) >= l.config.Window {
		// Create new bucket or reset existing one
		l.store[key] = &bucket{
			tokens:    l.config.Requests - 1,
			lastReset: now,
		}
		return true, l.config.Requests - 1, l.config.Requests, resetAt, nil
	}

	if b.tokens <= 0 {
		return false, 0, l.config.Requests, b.lastReset.Add(l.config.Window), nil
	}

	b.tokens--
	return true, b.tokens, l.config.Requests, b.lastReset.Add(l.config.Window), nil
}

// cleanup removes expired buckets periodically.
func (l *InMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(l.config.Window * 2)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, b := range l.store {
			if now.Sub(b.lastReset) >= l.config.Window*2 {
				delete(l.store, key)
			}
		}
		l.mu.Unlock()
	}
}

// RedisRateLimiter implements rate limiting using Redis.
type RedisRateLimiter struct {
	redis  *database.RedisClient
	config RateLimitConfig
}

// NewRedisRateLimiter creates a new Redis-backed rate limiter.
func NewRedisRateLimiter(redis *database.RedisClient, config RateLimitConfig) *RedisRateLimiter {
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}

	return &RedisRateLimiter{
		redis:  redis,
		config: config,
	}
}

// Allow checks if a request is allowed using Redis.
func (l *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, int, int, time.Time, error) {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	now := time.Now()
	resetAt := now.Add(l.config.Window)

	// Use Redis pipeline for atomic operations
	pipe := l.redis.Pipeline()

	// Increment counter
	incr := pipe.Incr(ctx, redisKey)

	// Set expiration if key is new
	pipe.Expire(ctx, redisKey, l.config.Window)

	// Get TTL
	ttl := pipe.TTL(ctx, redisKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, l.config.Requests, resetAt, fmt.Errorf("failed to execute rate limit check: %w", err)
	}

	count := int(incr.Val())
	remaining := l.config.Requests - count
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time from TTL
	if ttlDuration := ttl.Val(); ttlDuration > 0 {
		resetAt = now.Add(ttlDuration)
	}

	return count <= l.config.Requests, remaining, l.config.Requests, resetAt, nil
}

// RateLimit creates rate limiting middleware.
func RateLimit(limiter RateLimiter, config RateLimitConfig) func(http.Handler) http.Handler {
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := config.KeyFunc(r)

			allowed, remaining, limit, resetAt, err := limiter.Allow(r.Context(), key)
			if err != nil {
				// Log error but allow request to proceed
				response.Error(w, errors.ErrInternal("Rate limit check failed"))
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(resetAt).Seconds()), 10))
				response.Error(w, errors.ErrTooManyRequests("Rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
