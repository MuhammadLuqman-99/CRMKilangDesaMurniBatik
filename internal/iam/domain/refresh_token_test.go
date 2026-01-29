// Package domain contains the domain layer for the IAM service.
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewRefreshToken(t *testing.T) {
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	deviceInfo := DeviceInfo{
		DeviceID:   "device-123",
		DeviceType: "mobile",
		DeviceName: "iPhone 14",
		OS:         "iOS",
		OSVersion:  "17.0",
		Browser:    "Safari",
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		expiresAt  time.Time
		ipAddress  string
		userAgent  string
		deviceInfo DeviceInfo
		wantErr    bool
	}{
		{
			name:       "valid refresh token",
			userID:     userID,
			expiresAt:  expiresAt,
			ipAddress:  "192.168.1.1",
			userAgent:  "Mozilla/5.0",
			deviceInfo: deviceInfo,
			wantErr:    false,
		},
		{
			name:       "valid with empty device info",
			userID:     userID,
			expiresAt:  expiresAt,
			ipAddress:  "192.168.1.1",
			userAgent:  "Mozilla/5.0",
			deviceInfo: DeviceInfo{},
			wantErr:    false,
		},
		{
			name:       "nil user ID returns error",
			userID:     uuid.Nil,
			expiresAt:  expiresAt,
			ipAddress:  "192.168.1.1",
			userAgent:  "Mozilla/5.0",
			deviceInfo: deviceInfo,
			wantErr:    true,
		},
		{
			name:       "expired time returns error",
			userID:     userID,
			expiresAt:  time.Now().Add(-1 * time.Hour), // Already expired
			ipAddress:  "192.168.1.1",
			userAgent:  "Mozilla/5.0",
			deviceInfo: deviceInfo,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, plainToken, err := NewRefreshToken(tt.userID, tt.expiresAt, tt.ipAddress, tt.userAgent, tt.deviceInfo)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRefreshToken() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewRefreshToken() unexpected error = %v", err)
				}
				if token == nil {
					t.Fatal("NewRefreshToken() returned nil token")
				}
				if plainToken == "" {
					t.Error("NewRefreshToken() should return plain token")
				}
				if token.ID == uuid.Nil {
					t.Error("NewRefreshToken() should generate ID")
				}
				if token.UserID() != tt.userID {
					t.Errorf("NewRefreshToken() UserID = %v, want %v", token.UserID(), tt.userID)
				}
				if token.TokenHash() == "" {
					t.Error("NewRefreshToken() should set token hash")
				}
				if token.TokenHash() == plainToken {
					t.Error("NewRefreshToken() token hash should be different from plain token")
				}
				if token.IPAddress() != tt.ipAddress {
					t.Errorf("NewRefreshToken() IPAddress = %v, want %v", token.IPAddress(), tt.ipAddress)
				}
				if token.UserAgent() != tt.userAgent {
					t.Errorf("NewRefreshToken() UserAgent = %v, want %v", token.UserAgent(), tt.userAgent)
				}
			}
		})
	}
}

func TestReconstructRefreshToken(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	tokenHash := "hashed_token_value"
	deviceInfo := DeviceInfo{
		DeviceID:   "device-123",
		DeviceType: "desktop",
	}
	createdAt := time.Now().Add(-time.Hour)
	expiresAt := time.Now().Add(24 * time.Hour)
	revokedAt := time.Now().Add(-30 * time.Minute)

	token := ReconstructRefreshToken(
		id,
		userID,
		tokenHash,
		deviceInfo,
		"192.168.1.1",
		"Mozilla/5.0",
		expiresAt,
		&revokedAt,
		createdAt,
	)

	if token.ID != id {
		t.Errorf("ReconstructRefreshToken() ID = %v, want %v", token.ID, id)
	}
	if token.UserID() != userID {
		t.Errorf("ReconstructRefreshToken() UserID = %v, want %v", token.UserID(), userID)
	}
	if token.TokenHash() != tokenHash {
		t.Errorf("ReconstructRefreshToken() TokenHash = %v, want %v", token.TokenHash(), tokenHash)
	}
	if token.DeviceInfo().DeviceID != "device-123" {
		t.Errorf("ReconstructRefreshToken() DeviceInfo.DeviceID = %v", token.DeviceInfo().DeviceID)
	}
	if token.IPAddress() != "192.168.1.1" {
		t.Errorf("ReconstructRefreshToken() IPAddress = %v", token.IPAddress())
	}
	if token.UserAgent() != "Mozilla/5.0" {
		t.Errorf("ReconstructRefreshToken() UserAgent = %v", token.UserAgent())
	}
	if token.RevokedAt() == nil || !token.RevokedAt().Equal(revokedAt) {
		t.Errorf("ReconstructRefreshToken() RevokedAt = %v, want %v", token.RevokedAt(), revokedAt)
	}
}

func TestRefreshToken_Getters(t *testing.T) {
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	deviceInfo := DeviceInfo{
		DeviceID:       "device-123",
		DeviceType:     "mobile",
		DeviceName:     "Test Device",
		OS:             "iOS",
		OSVersion:      "17.0",
		Browser:        "Safari",
		BrowserVersion: "17.0",
	}

	token, _, _ := NewRefreshToken(userID, expiresAt, "192.168.1.1", "Mozilla/5.0", deviceInfo)

	if token.UserID() != userID {
		t.Errorf("RefreshToken.UserID() = %v, want %v", token.UserID(), userID)
	}
	if token.TokenHash() == "" {
		t.Error("RefreshToken.TokenHash() should not be empty")
	}
	if token.DeviceInfo().DeviceID != "device-123" {
		t.Errorf("RefreshToken.DeviceInfo().DeviceID = %v", token.DeviceInfo().DeviceID)
	}
	if token.DeviceInfo().DeviceName != "Test Device" {
		t.Errorf("RefreshToken.DeviceInfo().DeviceName = %v", token.DeviceInfo().DeviceName)
	}
	if token.IPAddress() != "192.168.1.1" {
		t.Errorf("RefreshToken.IPAddress() = %v", token.IPAddress())
	}
	if token.UserAgent() != "Mozilla/5.0" {
		t.Errorf("RefreshToken.UserAgent() = %v", token.UserAgent())
	}
	if token.RevokedAt() != nil {
		t.Error("RefreshToken.RevokedAt() should be nil for new token")
	}
}

func TestRefreshToken_IsExpired(t *testing.T) {
	userID := uuid.New()

	// Not expired
	token1, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	if token1.IsExpired() {
		t.Error("RefreshToken.IsExpired() should return false for non-expired token")
	}

	// Expired (using reconstruction)
	token2 := ReconstructRefreshToken(
		uuid.New(),
		userID,
		"hash",
		DeviceInfo{},
		"192.168.1.1",
		"Mozilla/5.0",
		time.Now().Add(-1*time.Hour), // Already expired
		nil,
		time.Now().Add(-24*time.Hour),
	)
	if !token2.IsExpired() {
		t.Error("RefreshToken.IsExpired() should return true for expired token")
	}
}

func TestRefreshToken_IsRevoked(t *testing.T) {
	userID := uuid.New()

	// Not revoked
	token1, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	if token1.IsRevoked() {
		t.Error("RefreshToken.IsRevoked() should return false for non-revoked token")
	}

	// Revoked
	revokedAt := time.Now()
	token2 := ReconstructRefreshToken(
		uuid.New(),
		userID,
		"hash",
		DeviceInfo{},
		"192.168.1.1",
		"Mozilla/5.0",
		time.Now().Add(24*time.Hour),
		&revokedAt,
		time.Now(),
	)
	if !token2.IsRevoked() {
		t.Error("RefreshToken.IsRevoked() should return true for revoked token")
	}
}

func TestRefreshToken_IsValid(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		expiresAt time.Time
		revokedAt *time.Time
		expected  bool
	}{
		{
			name:      "valid token",
			expiresAt: time.Now().Add(24 * time.Hour),
			revokedAt: nil,
			expected:  true,
		},
		{
			name:      "expired token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			revokedAt: nil,
			expected:  false,
		},
		{
			name:      "revoked token",
			expiresAt: time.Now().Add(24 * time.Hour),
			revokedAt: func() *time.Time { t := time.Now(); return &t }(),
			expected:  false,
		},
		{
			name:      "expired and revoked token",
			expiresAt: time.Now().Add(-1 * time.Hour),
			revokedAt: func() *time.Time { t := time.Now(); return &t }(),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := ReconstructRefreshToken(
				uuid.New(),
				userID,
				"hash",
				DeviceInfo{},
				"192.168.1.1",
				"Mozilla/5.0",
				tt.expiresAt,
				tt.revokedAt,
				time.Now(),
			)
			if token.IsValid() != tt.expected {
				t.Errorf("RefreshToken.IsValid() = %v, want %v", token.IsValid(), tt.expected)
			}
		})
	}
}

func TestRefreshToken_TimeUntilExpiration(t *testing.T) {
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)

	token, _, _ := NewRefreshToken(userID, expiresAt, "192.168.1.1", "Mozilla/5.0", DeviceInfo{})

	duration := token.TimeUntilExpiration()
	// Should be approximately 24 hours (allow some tolerance)
	if duration < 23*time.Hour || duration > 25*time.Hour {
		t.Errorf("RefreshToken.TimeUntilExpiration() = %v, expected ~24h", duration)
	}
}

func TestRefreshToken_Verify(t *testing.T) {
	userID := uuid.New()
	token, plainToken, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})

	// Valid verification
	if !token.Verify(plainToken) {
		t.Error("RefreshToken.Verify() should return true for correct token")
	}

	// Invalid verification
	if token.Verify("wrong_token") {
		t.Error("RefreshToken.Verify() should return false for incorrect token")
	}

	// Empty token
	if token.Verify("") {
		t.Error("RefreshToken.Verify() should return false for empty token")
	}
}

func TestRefreshToken_Revoke(t *testing.T) {
	userID := uuid.New()
	token, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})

	token.Revoke()

	if !token.IsRevoked() {
		t.Error("RefreshToken.Revoke() should mark token as revoked")
	}
	if token.RevokedAt() == nil {
		t.Error("RefreshToken.Revoke() should set RevokedAt")
	}
	if !token.IsValid() == true {
		t.Error("RefreshToken.IsValid() should return false after revoke")
	}

	// Revoking again should be idempotent
	originalRevokedAt := token.RevokedAt()
	token.Revoke()
	if !token.RevokedAt().Equal(*originalRevokedAt) {
		t.Error("RefreshToken.Revoke() should be idempotent")
	}
}

func TestRefreshToken_Rotate(t *testing.T) {
	userID := uuid.New()
	originalToken, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{
		DeviceID:   "device-123",
		DeviceType: "mobile",
	})

	newExpiresAt := time.Now().Add(48 * time.Hour)
	newToken, newPlainToken, err := originalToken.Rotate(newExpiresAt)

	if err != nil {
		t.Fatalf("RefreshToken.Rotate() unexpected error = %v", err)
	}

	// Original token should be revoked
	if !originalToken.IsRevoked() {
		t.Error("RefreshToken.Rotate() should revoke original token")
	}

	// New token should be created
	if newToken == nil {
		t.Fatal("RefreshToken.Rotate() should return new token")
	}
	if newPlainToken == "" {
		t.Error("RefreshToken.Rotate() should return new plain token")
	}
	if newToken.ID == originalToken.ID {
		t.Error("RefreshToken.Rotate() should create new token with different ID")
	}
	if newToken.UserID() != originalToken.UserID() {
		t.Error("RefreshToken.Rotate() should preserve user ID")
	}
	if newToken.IPAddress() != originalToken.IPAddress() {
		t.Error("RefreshToken.Rotate() should preserve IP address")
	}
	if newToken.UserAgent() != originalToken.UserAgent() {
		t.Error("RefreshToken.Rotate() should preserve user agent")
	}
	if newToken.DeviceInfo().DeviceID != "device-123" {
		t.Error("RefreshToken.Rotate() should preserve device info")
	}
	if !newToken.IsValid() {
		t.Error("RefreshToken.Rotate() new token should be valid")
	}
}

func TestHashPlainToken(t *testing.T) {
	plainToken := "my_secret_token"
	hash1 := HashPlainToken(plainToken)
	hash2 := HashPlainToken(plainToken)

	// Same input should produce same hash
	if hash1 != hash2 {
		t.Error("HashPlainToken() should produce same hash for same input")
	}

	// Different input should produce different hash
	hash3 := HashPlainToken("different_token")
	if hash1 == hash3 {
		t.Error("HashPlainToken() should produce different hash for different input")
	}

	// Hash should be non-empty
	if hash1 == "" {
		t.Error("HashPlainToken() should return non-empty hash")
	}

	// Hash should not equal plain token
	if hash1 == plainToken {
		t.Error("HashPlainToken() should return different value from input")
	}
}

func TestNewRefreshTokenFamily(t *testing.T) {
	family := NewRefreshTokenFamily()

	if family == nil {
		t.Fatal("NewRefreshTokenFamily() returned nil")
	}
	if family.FamilyID() == uuid.Nil {
		t.Error("NewRefreshTokenFamily() should generate family ID")
	}
	if len(family.tokens) != 0 {
		t.Error("NewRefreshTokenFamily() should have empty tokens")
	}
}

func TestRefreshTokenFamily_AddToken(t *testing.T) {
	family := NewRefreshTokenFamily()
	userID := uuid.New()

	token, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	family.AddToken(token)

	if len(family.tokens) != 1 {
		t.Errorf("RefreshTokenFamily.AddToken() len = %v, want 1", len(family.tokens))
	}

	// Add another token
	token2, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	family.AddToken(token2)

	if len(family.tokens) != 2 {
		t.Errorf("RefreshTokenFamily.AddToken() len = %v, want 2", len(family.tokens))
	}
}

func TestRefreshTokenFamily_RevokeAll(t *testing.T) {
	family := NewRefreshTokenFamily()
	userID := uuid.New()

	token1, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	token2, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	token3, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})

	family.AddToken(token1)
	family.AddToken(token2)
	family.AddToken(token3)

	family.RevokeAll()

	for i, token := range family.tokens {
		if !token.IsRevoked() {
			t.Errorf("RefreshTokenFamily.RevokeAll() token %d should be revoked", i)
		}
	}
}

func TestRefreshTokenFamily_ActiveTokens(t *testing.T) {
	family := NewRefreshTokenFamily()
	userID := uuid.New()

	// Add valid token
	token1, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	family.AddToken(token1)

	// Add another valid token
	token2, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	family.AddToken(token2)

	// Add revoked token
	token3, _, _ := NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", DeviceInfo{})
	token3.Revoke()
	family.AddToken(token3)

	// Add expired token
	token4 := ReconstructRefreshToken(
		uuid.New(),
		userID,
		"hash",
		DeviceInfo{},
		"192.168.1.1",
		"Mozilla/5.0",
		time.Now().Add(-1*time.Hour), // Expired
		nil,
		time.Now().Add(-24*time.Hour),
	)
	family.AddToken(token4)

	activeTokens := family.ActiveTokens()

	if len(activeTokens) != 2 {
		t.Errorf("RefreshTokenFamily.ActiveTokens() len = %v, want 2", len(activeTokens))
	}

	// All active tokens should be valid
	for _, token := range activeTokens {
		if !token.IsValid() {
			t.Error("RefreshTokenFamily.ActiveTokens() should only return valid tokens")
		}
	}
}

func TestRefreshTokenFamily_FamilyID(t *testing.T) {
	family := NewRefreshTokenFamily()

	familyID := family.FamilyID()
	if familyID == uuid.Nil {
		t.Error("RefreshTokenFamily.FamilyID() should return non-nil UUID")
	}

	// Should return same ID on multiple calls
	if family.FamilyID() != familyID {
		t.Error("RefreshTokenFamily.FamilyID() should be consistent")
	}
}

func TestDeviceInfo(t *testing.T) {
	deviceInfo := DeviceInfo{
		DeviceID:       "device-123",
		DeviceType:     "mobile",
		DeviceName:     "iPhone 14 Pro",
		OS:             "iOS",
		OSVersion:      "17.0",
		Browser:        "Safari",
		BrowserVersion: "17.0",
	}

	if deviceInfo.DeviceID != "device-123" {
		t.Errorf("DeviceInfo.DeviceID = %v", deviceInfo.DeviceID)
	}
	if deviceInfo.DeviceType != "mobile" {
		t.Errorf("DeviceInfo.DeviceType = %v", deviceInfo.DeviceType)
	}
	if deviceInfo.DeviceName != "iPhone 14 Pro" {
		t.Errorf("DeviceInfo.DeviceName = %v", deviceInfo.DeviceName)
	}
	if deviceInfo.OS != "iOS" {
		t.Errorf("DeviceInfo.OS = %v", deviceInfo.OS)
	}
	if deviceInfo.OSVersion != "17.0" {
		t.Errorf("DeviceInfo.OSVersion = %v", deviceInfo.OSVersion)
	}
	if deviceInfo.Browser != "Safari" {
		t.Errorf("DeviceInfo.Browser = %v", deviceInfo.Browser)
	}
	if deviceInfo.BrowserVersion != "17.0" {
		t.Errorf("DeviceInfo.BrowserVersion = %v", deviceInfo.BrowserVersion)
	}
}
