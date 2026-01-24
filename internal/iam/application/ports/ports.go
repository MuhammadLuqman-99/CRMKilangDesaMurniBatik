// Package ports defines the interfaces (ports) for the application layer.
// These interfaces define how the application layer interacts with external systems.
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Authentication Ports
// ============================================================================

// PasswordHasher defines the interface for password hashing operations.
type PasswordHasher interface {
	// Hash hashes a plain text password.
	Hash(password string) (string, error)

	// Verify verifies a password against a hash.
	Verify(password, hash string) (bool, error)
}

// TokenService defines the interface for JWT token operations.
type TokenService interface {
	// GenerateAccessToken generates an access token.
	GenerateAccessToken(claims *TokenClaims) (string, error)

	// GenerateRefreshToken generates a refresh token.
	GenerateRefreshToken() (string, error)

	// ValidateAccessToken validates an access token and returns claims.
	ValidateAccessToken(token string) (*TokenClaims, error)

	// GetAccessTokenExpiry returns the access token expiry duration.
	GetAccessTokenExpiry() time.Duration

	// GetRefreshTokenExpiry returns the refresh token expiry duration.
	GetRefreshTokenExpiry() time.Duration
}

// TokenClaims represents the claims in a JWT token.
type TokenClaims struct {
	UserID      uuid.UUID `json:"user_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	IssuedAt    int64     `json:"iat"`
	ExpiresAt   int64     `json:"exp"`
}

// ============================================================================
// Event Publishing Ports
// ============================================================================

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	// Publish publishes a domain event.
	Publish(ctx context.Context, event DomainEventMessage) error

	// PublishMany publishes multiple domain events.
	PublishMany(ctx context.Context, events []DomainEventMessage) error
}

// DomainEventMessage represents a domain event message for publishing.
type DomainEventMessage struct {
	EventType     string                 `json:"event_type"`
	AggregateID   uuid.UUID              `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Payload       map[string]interface{} `json:"payload"`
	OccurredAt    time.Time              `json:"occurred_at"`
	TenantID      *uuid.UUID             `json:"tenant_id,omitempty"`
}

// ============================================================================
// Email Ports
// ============================================================================

// EmailService defines the interface for sending emails.
type EmailService interface {
	// SendWelcomeEmail sends a welcome email to a new user.
	SendWelcomeEmail(ctx context.Context, email, firstName string, verificationToken string) error

	// SendPasswordResetEmail sends a password reset email.
	SendPasswordResetEmail(ctx context.Context, email, firstName, resetToken string) error

	// SendEmailVerificationEmail sends an email verification email.
	SendEmailVerificationEmail(ctx context.Context, email, firstName, verificationToken string) error

	// SendPasswordChangedEmail notifies user of password change.
	SendPasswordChangedEmail(ctx context.Context, email, firstName string) error
}

// ============================================================================
// Cache Ports
// ============================================================================

// CacheService defines the interface for caching operations.
type CacheService interface {
	// Get retrieves a value from cache.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with expiration.
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete removes a value from cache.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache.
	Exists(ctx context.Context, key string) (bool, error)

	// Increment increments a counter in cache.
	Increment(ctx context.Context, key string) (int64, error)

	// SetNX sets a value only if it doesn't exist.
	SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error)
}

// ============================================================================
// Audit Logging Ports
// ============================================================================

// AuditLogger defines the interface for audit logging.
type AuditLogger interface {
	// Log logs an audit event.
	Log(ctx context.Context, entry AuditEntry) error
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	TenantID   uuid.UUID              `json:"tenant_id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
	OldValues  map[string]interface{} `json:"old_values,omitempty"`
	NewValues  map[string]interface{} `json:"new_values,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
}

// Audit action constants
const (
	AuditActionCreate     = "create"
	AuditActionUpdate     = "update"
	AuditActionDelete     = "delete"
	AuditActionLogin      = "login"
	AuditActionLogout     = "logout"
	AuditActionLoginFailed = "login_failed"
	AuditActionPasswordChanged = "password_changed"
	AuditActionRoleAssigned    = "role_assigned"
	AuditActionRoleRemoved     = "role_removed"
)

// ============================================================================
// Transaction Ports
// ============================================================================

// TransactionManager defines the interface for managing database transactions.
type TransactionManager interface {
	// WithTransaction executes a function within a transaction.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error

	// BeginTx starts a new transaction.
	BeginTx(ctx context.Context) (context.Context, error)

	// CommitTx commits the current transaction.
	CommitTx(ctx context.Context) error

	// RollbackTx rolls back the current transaction.
	RollbackTx(ctx context.Context) error
}

// ============================================================================
// Rate Limiting Ports
// ============================================================================

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	// Allow checks if the request is allowed.
	Allow(ctx context.Context, key string) (bool, error)

	// AllowN checks if N requests are allowed.
	AllowN(ctx context.Context, key string, n int) (bool, error)

	// Reset resets the rate limit for a key.
	Reset(ctx context.Context, key string) error
}

// ============================================================================
// Token Blacklist Ports
// ============================================================================

// TokenBlacklist defines the interface for token blacklisting.
type TokenBlacklist interface {
	// Add adds a token to the blacklist.
	Add(ctx context.Context, token string, expiration time.Duration) error

	// IsBlacklisted checks if a token is blacklisted.
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

// ============================================================================
// Verification Token Ports
// ============================================================================

// VerificationTokenService defines the interface for verification tokens.
type VerificationTokenService interface {
	// GenerateEmailVerificationToken generates an email verification token.
	GenerateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error)

	// ValidateEmailVerificationToken validates an email verification token.
	ValidateEmailVerificationToken(ctx context.Context, token string) (uuid.UUID, error)

	// GeneratePasswordResetToken generates a password reset token.
	GeneratePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error)

	// ValidatePasswordResetToken validates a password reset token.
	ValidatePasswordResetToken(ctx context.Context, token string) (uuid.UUID, error)

	// InvalidateToken invalidates a verification token.
	InvalidateToken(ctx context.Context, token string) error
}
