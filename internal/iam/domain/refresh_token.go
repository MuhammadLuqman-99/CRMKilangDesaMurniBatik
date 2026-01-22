// Package domain contains the domain layer for the IAM service.
package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DeviceInfo represents information about the device that requested the token.
type DeviceInfo struct {
	DeviceID     string `json:"device_id,omitempty"`
	DeviceType   string `json:"device_type,omitempty"`
	DeviceName   string `json:"device_name,omitempty"`
	OS           string `json:"os,omitempty"`
	OSVersion    string `json:"os_version,omitempty"`
	Browser      string `json:"browser,omitempty"`
	BrowserVersion string `json:"browser_version,omitempty"`
}

// RefreshToken represents a refresh token entity.
type RefreshToken struct {
	BaseEntity
	userID     uuid.UUID
	tokenHash  string
	deviceInfo DeviceInfo
	ipAddress  string
	userAgent  string
	expiresAt  time.Time
	revokedAt  *time.Time
}

// NewRefreshToken creates a new RefreshToken entity.
// Returns the entity and the plain token that should be sent to the client.
func NewRefreshToken(userID uuid.UUID, expiresAt time.Time, ipAddress, userAgent string, deviceInfo DeviceInfo) (*RefreshToken, string, error) {
	if userID == uuid.Nil {
		return nil, "", ErrRefreshTokenUserRequired
	}

	if expiresAt.Before(time.Now().UTC()) {
		return nil, "", ErrRefreshTokenExpired
	}

	// Generate a random token
	plainToken, err := generateSecureToken(32)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the token for storage
	tokenHash := hashToken(plainToken)

	token := &RefreshToken{
		BaseEntity: NewBaseEntity(),
		userID:     userID,
		tokenHash:  tokenHash,
		deviceInfo: deviceInfo,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
		expiresAt:  expiresAt,
	}

	return token, plainToken, nil
}

// ReconstructRefreshToken reconstructs a RefreshToken from persistence.
func ReconstructRefreshToken(
	id uuid.UUID,
	userID uuid.UUID,
	tokenHash string,
	deviceInfo DeviceInfo,
	ipAddress, userAgent string,
	expiresAt time.Time,
	revokedAt *time.Time,
	createdAt time.Time,
) *RefreshToken {
	return &RefreshToken{
		BaseEntity: BaseEntity{
			ID:        id,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		userID:     userID,
		tokenHash:  tokenHash,
		deviceInfo: deviceInfo,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
		expiresAt:  expiresAt,
		revokedAt:  revokedAt,
	}
}

// Getters

// UserID returns the user ID associated with the token.
func (t *RefreshToken) UserID() uuid.UUID {
	return t.userID
}

// TokenHash returns the token hash.
func (t *RefreshToken) TokenHash() string {
	return t.tokenHash
}

// DeviceInfo returns the device information.
func (t *RefreshToken) DeviceInfo() DeviceInfo {
	return t.deviceInfo
}

// IPAddress returns the IP address.
func (t *RefreshToken) IPAddress() string {
	return t.ipAddress
}

// UserAgent returns the user agent.
func (t *RefreshToken) UserAgent() string {
	return t.userAgent
}

// ExpiresAt returns the expiration time.
func (t *RefreshToken) ExpiresAt() time.Time {
	return t.expiresAt
}

// RevokedAt returns the revocation time.
func (t *RefreshToken) RevokedAt() *time.Time {
	return t.revokedAt
}

// Status checks

// IsExpired returns true if the token has expired.
func (t *RefreshToken) IsExpired() bool {
	return time.Now().UTC().After(t.expiresAt)
}

// IsRevoked returns true if the token has been revoked.
func (t *RefreshToken) IsRevoked() bool {
	return t.revokedAt != nil
}

// IsValid returns true if the token is valid (not expired and not revoked).
func (t *RefreshToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}

// TimeUntilExpiration returns the duration until the token expires.
func (t *RefreshToken) TimeUntilExpiration() time.Duration {
	return time.Until(t.expiresAt)
}

// Behaviors

// Verify verifies that the provided plain token matches this refresh token.
func (t *RefreshToken) Verify(plainToken string) bool {
	return hashToken(plainToken) == t.tokenHash
}

// Revoke revokes the refresh token.
func (t *RefreshToken) Revoke() {
	if t.revokedAt != nil {
		return
	}

	now := time.Now().UTC()
	t.revokedAt = &now
	t.MarkUpdated()
}

// Rotate creates a new refresh token and revokes the current one.
// Returns the new token entity and the plain token for the client.
func (t *RefreshToken) Rotate(expiresAt time.Time) (*RefreshToken, string, error) {
	// Revoke current token
	t.Revoke()

	// Create new token with same device info
	return NewRefreshToken(t.userID, expiresAt, t.ipAddress, t.userAgent, t.deviceInfo)
}

// Helper functions

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken creates a SHA-256 hash of the token.
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// HashPlainToken is a helper to hash a plain token for comparison.
func HashPlainToken(plainToken string) string {
	return hashToken(plainToken)
}

// RefreshTokenFamily manages a family of refresh tokens for rotation detection.
type RefreshTokenFamily struct {
	familyID uuid.UUID
	tokens   []*RefreshToken
}

// NewRefreshTokenFamily creates a new token family.
func NewRefreshTokenFamily() *RefreshTokenFamily {
	return &RefreshTokenFamily{
		familyID: uuid.New(),
		tokens:   make([]*RefreshToken, 0),
	}
}

// FamilyID returns the family ID.
func (f *RefreshTokenFamily) FamilyID() uuid.UUID {
	return f.familyID
}

// AddToken adds a token to the family.
func (f *RefreshTokenFamily) AddToken(token *RefreshToken) {
	f.tokens = append(f.tokens, token)
}

// RevokeAll revokes all tokens in the family.
func (f *RefreshTokenFamily) RevokeAll() {
	for _, token := range f.tokens {
		token.Revoke()
	}
}

// ActiveTokens returns all active (non-revoked, non-expired) tokens.
func (f *RefreshTokenFamily) ActiveTokens() []*RefreshToken {
	var active []*RefreshToken
	for _, token := range f.tokens {
		if token.IsValid() {
			active = append(active, token)
		}
	}
	return active
}

// RefreshToken errors
var (
	ErrRefreshTokenNotFound     = fmt.Errorf("refresh token not found")
	ErrRefreshTokenUserRequired = fmt.Errorf("user ID is required")
	ErrRefreshTokenExpired      = fmt.Errorf("refresh token has expired")
	ErrRefreshTokenRevoked      = fmt.Errorf("refresh token has been revoked")
	ErrRefreshTokenInvalid      = fmt.Errorf("invalid refresh token")
	ErrRefreshTokenReused       = fmt.Errorf("refresh token reuse detected")
)
