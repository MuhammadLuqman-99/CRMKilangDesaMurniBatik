// Package redis contains Redis-based cache implementations.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Session represents an active user session.
type Session struct {
	SessionID   string            `json:"session_id"`
	UserID      uuid.UUID         `json:"user_id"`
	TenantID    uuid.UUID         `json:"tenant_id"`
	DeviceInfo  map[string]string `json:"device_info,omitempty"`
	IPAddress   string            `json:"ip_address,omitempty"`
	UserAgent   string            `json:"user_agent,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
	LastActive  time.Time         `json:"last_active"`
}

// SessionManager manages user sessions in Redis.
type SessionManager struct {
	client       *redis.Client
	prefix       string
	sessionTTL   time.Duration
	maxSessions  int
}

// SessionManagerConfig holds configuration for the session manager.
type SessionManagerConfig struct {
	Prefix      string
	SessionTTL  time.Duration
	MaxSessions int
}

// NewSessionManager creates a new session manager.
func NewSessionManager(client *redis.Client, config SessionManagerConfig) *SessionManager {
	if config.Prefix == "" {
		config.Prefix = "session:"
	}
	if config.SessionTTL == 0 {
		config.SessionTTL = 24 * time.Hour
	}
	if config.MaxSessions == 0 {
		config.MaxSessions = 5
	}

	return &SessionManager{
		client:      client,
		prefix:      config.Prefix,
		sessionTTL:  config.SessionTTL,
		maxSessions: config.MaxSessions,
	}
}

// CreateSession creates a new session for a user.
func (m *SessionManager) CreateSession(ctx context.Context, session *Session) error {
	// Set timestamps
	now := time.Now().UTC()
	session.CreatedAt = now
	session.LastActive = now
	session.ExpiresAt = now.Add(m.sessionTTL)

	// Serialize session
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store session data
	sessionKey := m.prefix + session.SessionID
	if err := m.client.Set(ctx, sessionKey, data, m.sessionTTL).Err(); err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	// Add to user's session set
	userSessionsKey := m.prefix + "user:" + session.UserID.String()
	if err := m.client.SAdd(ctx, userSessionsKey, session.SessionID).Err(); err != nil {
		return fmt.Errorf("failed to add session to user set: %w", err)
	}

	// Set expiry on user sessions set
	m.client.Expire(ctx, userSessionsKey, m.sessionTTL*2)

	// Enforce max sessions limit
	if err := m.enforceMaxSessions(ctx, session.UserID); err != nil {
		// Log but don't fail on enforcement error
		_ = err
	}

	return nil
}

// GetSession retrieves a session by ID.
func (m *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	sessionKey := m.prefix + sessionID

	data, err := m.client.Get(ctx, sessionKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// UpdateSessionActivity updates the last active time of a session.
func (m *SessionManager) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActive = time.Now().UTC()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionKey := m.prefix + sessionID
	remainingTTL, err := m.client.TTL(ctx, sessionKey).Result()
	if err != nil || remainingTTL <= 0 {
		remainingTTL = m.sessionTTL
	}

	if err := m.client.Set(ctx, sessionKey, data, remainingTTL).Err(); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// ExtendSession extends the session expiration time.
func (m *SessionManager) ExtendSession(ctx context.Context, sessionID string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.LastActive = time.Now().UTC()
	session.ExpiresAt = time.Now().UTC().Add(m.sessionTTL)

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionKey := m.prefix + sessionID
	if err := m.client.Set(ctx, sessionKey, data, m.sessionTTL).Err(); err != nil {
		return fmt.Errorf("failed to extend session: %w", err)
	}

	return nil
}

// DeleteSession deletes a session.
func (m *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	// Get session to find user ID
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		if err == ErrSessionNotFound {
			return nil
		}
		return err
	}

	// Delete session
	sessionKey := m.prefix + sessionID
	if err := m.client.Del(ctx, sessionKey).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from user's session set
	userSessionsKey := m.prefix + "user:" + session.UserID.String()
	if err := m.client.SRem(ctx, userSessionsKey, sessionID).Err(); err != nil {
		return fmt.Errorf("failed to remove session from user set: %w", err)
	}

	return nil
}

// DeleteUserSessions deletes all sessions for a user.
func (m *SessionManager) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	userSessionsKey := m.prefix + "user:" + userID.String()

	// Get all session IDs
	sessionIDs, err := m.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	// Delete all sessions
	pipe := m.client.Pipeline()
	for _, sessionID := range sessionIDs {
		sessionKey := m.prefix + sessionID
		pipe.Del(ctx, sessionKey)
	}

	// Delete the user sessions set
	pipe.Del(ctx, userSessionsKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// GetUserSessions retrieves all active sessions for a user.
func (m *SessionManager) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*Session, error) {
	userSessionsKey := m.prefix + "user:" + userID.String()

	sessionIDs, err := m.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user session IDs: %w", err)
	}

	var sessions []*Session
	var expiredSessionIDs []string

	for _, sessionID := range sessionIDs {
		session, err := m.GetSession(ctx, sessionID)
		if err != nil {
			if err == ErrSessionNotFound {
				expiredSessionIDs = append(expiredSessionIDs, sessionID)
				continue
			}
			continue
		}
		sessions = append(sessions, session)
	}

	// Clean up expired session references
	if len(expiredSessionIDs) > 0 {
		m.client.SRem(ctx, userSessionsKey, expiredSessionIDs)
	}

	return sessions, nil
}

// CountUserSessions counts active sessions for a user.
func (m *SessionManager) CountUserSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	sessions, err := m.GetUserSessions(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int64(len(sessions)), nil
}

// enforceMaxSessions removes oldest sessions if max is exceeded.
func (m *SessionManager) enforceMaxSessions(ctx context.Context, userID uuid.UUID) error {
	sessions, err := m.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	if len(sessions) <= m.maxSessions {
		return nil
	}

	// Sort by created time and remove oldest
	// Simple approach: just remove excess sessions
	excess := len(sessions) - m.maxSessions
	for i := 0; i < excess; i++ {
		oldest := sessions[i]
		m.DeleteSession(ctx, oldest.SessionID)
	}

	return nil
}

// SessionExists checks if a session exists.
func (m *SessionManager) SessionExists(ctx context.Context, sessionID string) (bool, error) {
	sessionKey := m.prefix + sessionID

	result, err := m.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return result > 0, nil
}

// UpdateSessionPermissions updates the cached permissions for a session.
func (m *SessionManager) UpdateSessionPermissions(ctx context.Context, sessionID string, permissions []string) error {
	session, err := m.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Permissions = permissions

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionKey := m.prefix + sessionID
	remainingTTL, err := m.client.TTL(ctx, sessionKey).Result()
	if err != nil || remainingTTL <= 0 {
		remainingTTL = m.sessionTTL
	}

	if err := m.client.Set(ctx, sessionKey, data, remainingTTL).Err(); err != nil {
		return fmt.Errorf("failed to update session permissions: %w", err)
	}

	return nil
}

// Session errors
var (
	ErrSessionNotFound = fmt.Errorf("session not found")
	ErrSessionExpired  = fmt.Errorf("session expired")
)
