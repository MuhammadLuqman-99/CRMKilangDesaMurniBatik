// Package middleware contains HTTP middleware implementations.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	iamhttp "github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/interfaces/http"
)

// RateLimiter interface for rate limiting implementations.
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (remaining int64, allowed bool, err error)
}

// RateLimitConfig holds rate limit configuration.
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	BurstSize         int
	KeyFunc           func(*http.Request) string
}

// DefaultRateLimitConfig returns default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60,
		RequestsPerHour:   1000,
		BurstSize:         10,
		KeyFunc:           IPBasedKey,
	}
}

// IPBasedKey returns the client IP as rate limit key.
func IPBasedKey(r *http.Request) string {
	// Try X-Forwarded-For first (for proxied requests)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return "ip:" + forwarded
	}

	// Try X-Real-IP
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return "ip:" + realIP
	}

	// Fall back to RemoteAddr
	return "ip:" + r.RemoteAddr
}

// UserBasedKey returns the user ID as rate limit key.
func UserBasedKey(r *http.Request) string {
	userID := GetUserID(r.Context())
	if userID.String() != "00000000-0000-0000-0000-000000000000" {
		return "user:" + userID.String()
	}
	return IPBasedKey(r)
}

// TenantBasedKey returns the tenant ID as rate limit key.
func TenantBasedKey(r *http.Request) string {
	tenantID := GetTenantID(r.Context())
	if tenantID.String() != "00000000-0000-0000-0000-000000000000" {
		return "tenant:" + tenantID.String()
	}
	return IPBasedKey(r)
}

// EndpointBasedKey combines IP and endpoint as rate limit key.
func EndpointBasedKey(r *http.Request) string {
	return fmt.Sprintf("endpoint:%s:%s:%s", r.Method, r.URL.Path, IPBasedKey(r))
}

// RateLimit middleware with external rate limiter.
func RateLimit(limiter RateLimiter, config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := config.KeyFunc(r)

			remaining, allowed, err := limiter.Allow(r.Context(), key, int64(config.RequestsPerMinute), time.Minute)
			if err != nil {
				// On error, allow request but log
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

			if !allowed {
				w.Header().Set("Retry-After", "60")
				iamhttp.WriteError(w, http.StatusTooManyRequests, iamhttp.ErrCodeTooManyRequests, "rate limit exceeded", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// InMemoryRateLimiter provides a simple in-memory rate limiter.
type InMemoryRateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*rateLimitEntry
	cleanup time.Duration
}

type rateLimitEntry struct {
	count     int64
	resetTime time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
func NewInMemoryRateLimiter(cleanupInterval time.Duration) *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		entries: make(map[string]*rateLimitEntry),
		cleanup: cleanupInterval,
	}

	// Start cleanup goroutine
	go limiter.cleanupLoop()

	return limiter
}

// Allow checks if a request is allowed.
func (l *InMemoryRateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (remaining int64, allowed bool, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	entry, exists := l.entries[key]

	if !exists || now.After(entry.resetTime) {
		// Create new entry or reset expired entry
		l.entries[key] = &rateLimitEntry{
			count:     1,
			resetTime: now.Add(window),
		}
		return limit - 1, true, nil
	}

	// Check if limit exceeded
	if entry.count >= limit {
		return 0, false, nil
	}

	// Increment count
	entry.count++
	return limit - entry.count, true, nil
}

func (l *InMemoryRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for key, entry := range l.entries {
			if now.After(entry.resetTime) {
				delete(l.entries, key)
			}
		}
		l.mu.Unlock()
	}
}

// SimpleRateLimit provides a simple rate limiting middleware without external dependencies.
func SimpleRateLimit(requestsPerMinute int, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	limiter := NewInMemoryRateLimiter(5 * time.Minute)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)

			remaining, allowed, _ := limiter.Allow(r.Context(), key, int64(requestsPerMinute), time.Minute)

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

			if !allowed {
				w.Header().Set("Retry-After", "60")
				iamhttp.WriteError(w, http.StatusTooManyRequests, iamhttp.ErrCodeTooManyRequests, "rate limit exceeded", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ThrottleConfig holds throttle configuration for specific endpoints.
type ThrottleConfig struct {
	Limit  int
	Window time.Duration
}

// EndpointThrottle creates endpoint-specific throttling middleware.
func EndpointThrottle(limiter RateLimiter, config ThrottleConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := EndpointBasedKey(r)

			remaining, allowed, err := limiter.Allow(r.Context(), key, int64(config.Limit), config.Window)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

			if !allowed {
				retryAfter := int(config.Window.Seconds())
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				iamhttp.WriteError(w, http.StatusTooManyRequests, iamhttp.ErrCodeTooManyRequests, "rate limit exceeded", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoginThrottle creates throttling specifically for login attempts.
func LoginThrottle(limiter RateLimiter, maxAttempts int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP-based key for login throttling
			key := "login:" + IPBasedKey(r)

			remaining, allowed, err := limiter.Allow(r.Context(), key, int64(maxAttempts), window)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				retryAfter := int(window.Seconds())
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				iamhttp.WriteError(w, http.StatusTooManyRequests, iamhttp.ErrCodeTooManyRequests,
					fmt.Sprintf("too many login attempts, please try again in %d seconds", retryAfter), nil)
				return
			}

			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

			next.ServeHTTP(w, r)
		})
	}
}
