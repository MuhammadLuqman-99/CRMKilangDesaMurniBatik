// Package redis contains Redis-based cache implementations.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklistService manages blacklisted tokens in Redis.
type TokenBlacklistService struct {
	client *redis.Client
	prefix string
}

// NewTokenBlacklistService creates a new token blacklist service.
func NewTokenBlacklistService(client *redis.Client, prefix string) *TokenBlacklistService {
	if prefix == "" {
		prefix = "token:blacklist:"
	}
	return &TokenBlacklistService{
		client: client,
		prefix: prefix,
	}
}

// Blacklist adds a token to the blacklist with the given TTL.
// The TTL should be set to the remaining validity time of the token.
func (s *TokenBlacklistService) Blacklist(ctx context.Context, tokenID string, ttl time.Duration) error {
	key := s.prefix + tokenID

	// Store the blacklist entry with the token's remaining TTL
	if err := s.client.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted.
func (s *TokenBlacklistService) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := s.prefix + tokenID

	result, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return result > 0, nil
}

// Remove removes a token from the blacklist (rarely needed but available).
func (s *TokenBlacklistService) Remove(ctx context.Context, tokenID string) error {
	key := s.prefix + tokenID

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}

	return nil
}

// BlacklistBatch adds multiple tokens to the blacklist.
func (s *TokenBlacklistService) BlacklistBatch(ctx context.Context, tokenIDs []string, ttl time.Duration) error {
	if len(tokenIDs) == 0 {
		return nil
	}

	pipe := s.client.Pipeline()
	for _, tokenID := range tokenIDs {
		key := s.prefix + tokenID
		pipe.Set(ctx, key, "1", ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to blacklist tokens: %w", err)
	}

	return nil
}

// Count returns the number of blacklisted tokens.
func (s *TokenBlacklistService) Count(ctx context.Context) (int64, error) {
	var count int64
	iter := s.client.Scan(ctx, 0, s.prefix+"*", 0).Iterator()

	for iter.Next(ctx) {
		count++
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to count blacklisted tokens: %w", err)
	}

	return count, nil
}

// SessionBlacklistService manages blacklisted user sessions.
type SessionBlacklistService struct {
	client *redis.Client
	prefix string
}

// NewSessionBlacklistService creates a new session blacklist service.
func NewSessionBlacklistService(client *redis.Client, prefix string) *SessionBlacklistService {
	if prefix == "" {
		prefix = "session:blacklist:"
	}
	return &SessionBlacklistService{
		client: client,
		prefix: prefix,
	}
}

// BlacklistUserSessions invalidates all sessions for a user after a specific time.
func (s *SessionBlacklistService) BlacklistUserSessions(ctx context.Context, userID string, after time.Time, ttl time.Duration) error {
	key := s.prefix + userID

	// Store the timestamp after which all tokens are invalid
	if err := s.client.Set(ctx, key, after.Unix(), ttl).Err(); err != nil {
		return fmt.Errorf("failed to blacklist user sessions: %w", err)
	}

	return nil
}

// IsSessionValid checks if a user's session issued at a specific time is still valid.
func (s *SessionBlacklistService) IsSessionValid(ctx context.Context, userID string, issuedAt time.Time) (bool, error) {
	key := s.prefix + userID

	blacklistTime, err := s.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			// No blacklist entry, session is valid
			return true, nil
		}
		return false, fmt.Errorf("failed to check session validity: %w", err)
	}

	// Session is valid if it was issued after the blacklist time
	return issuedAt.Unix() > blacklistTime, nil
}

// RemoveUserBlacklist removes the user's session blacklist entry.
func (s *SessionBlacklistService) RemoveUserBlacklist(ctx context.Context, userID string) error {
	key := s.prefix + userID

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove user session blacklist: %w", err)
	}

	return nil
}

// RateLimitService provides token-based rate limiting.
type RateLimitService struct {
	client *redis.Client
	prefix string
}

// NewRateLimitService creates a new rate limit service.
func NewRateLimitService(client *redis.Client, prefix string) *RateLimitService {
	if prefix == "" {
		prefix = "ratelimit:"
	}
	return &RateLimitService{
		client: client,
		prefix: prefix,
	}
}

// Allow checks if a request is allowed within rate limits.
// Returns the remaining requests and whether the request is allowed.
func (s *RateLimitService) Allow(ctx context.Context, key string, limit int64, window time.Duration) (remaining int64, allowed bool, err error) {
	fullKey := s.prefix + key

	pipe := s.client.TxPipeline()
	incr := pipe.Incr(ctx, fullKey)
	pipe.Expire(ctx, fullKey, window)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	count := incr.Val()
	if count > limit {
		return 0, false, nil
	}

	return limit - count, true, nil
}

// Reset resets the rate limit counter for a key.
func (s *RateLimitService) Reset(ctx context.Context, key string) error {
	fullKey := s.prefix + key

	if err := s.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	return nil
}

// GetRemaining gets the remaining requests for a key.
func (s *RateLimitService) GetRemaining(ctx context.Context, key string, limit int64) (int64, error) {
	fullKey := s.prefix + key

	count, err := s.client.Get(ctx, fullKey).Int64()
	if err != nil {
		if err == redis.Nil {
			return limit, nil
		}
		return 0, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// SlidingWindowRateLimit implements sliding window rate limiting for more accurate limits.
func (s *RateLimitService) SlidingWindowRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (allowed bool, err error) {
	fullKey := s.prefix + "sw:" + key
	now := time.Now()
	windowStart := now.Add(-window)

	pipe := s.client.TxPipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, fullKey, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count current entries in window
	count := pipe.ZCard(ctx, fullKey)

	// Add new entry if under limit
	pipe.ZAdd(ctx, fullKey, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: now.UnixNano(),
	})

	// Set expiry on the set
	pipe.Expire(ctx, fullKey, window)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check sliding window rate limit: %w", err)
	}

	return count.Val() < limit, nil
}
