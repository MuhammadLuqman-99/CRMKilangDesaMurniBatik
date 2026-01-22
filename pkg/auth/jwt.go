// Package auth provides authentication and authorization utilities for the CRM application.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/errors"
)

// TokenType represents the type of JWT token.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims represents the JWT claims for the application.
type Claims struct {
	jwt.RegisteredClaims
	UserID      string            `json:"user_id"`
	TenantID    string            `json:"tenant_id"`
	Email       string            `json:"email"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions,omitempty"`
	TokenType   TokenType         `json:"token_type"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTManager handles JWT token operations.
type JWTManager struct {
	config *config.JWTConfig
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		config: cfg,
	}
}

// GenerateTokenPair generates an access and refresh token pair.
func (m *JWTManager) GenerateTokenPair(userID, tenantID, email string, roles []string) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessToken, accessExp, err := m.generateToken(userID, tenantID, email, roles, TokenTypeAccess, now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, _, err := m.generateToken(userID, tenantID, email, roles, TokenTypeRefresh, now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(m.config.AccessExpiry.Seconds()),
		ExpiresAt:    accessExp,
	}, nil
}

// GenerateAccessToken generates only an access token.
func (m *JWTManager) GenerateAccessToken(userID, tenantID, email string, roles []string) (string, error) {
	token, _, err := m.generateToken(userID, tenantID, email, roles, TokenTypeAccess, time.Now())
	return token, err
}

// GenerateRefreshToken generates only a refresh token.
func (m *JWTManager) GenerateRefreshToken(userID, tenantID, email string, roles []string) (string, error) {
	token, _, err := m.generateToken(userID, tenantID, email, roles, TokenTypeRefresh, time.Now())
	return token, err
}

// generateToken generates a JWT token with the specified parameters.
func (m *JWTManager) generateToken(userID, tenantID, email string, roles []string, tokenType TokenType, now time.Time) (string, time.Time, error) {
	var expiry time.Duration
	if tokenType == TokenTypeAccess {
		expiry = m.config.AccessExpiry
	} else {
		expiry = m.config.RefreshExpiry
	}

	expiresAt := now.Add(expiry)

	// Generate a unique token ID
	jti, err := generateTokenID()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate token ID: %w", err)
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    m.config.Issuer,
			Audience:  jwt.ClaimStrings{m.config.Audience},
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID:    userID,
		TenantID:  tenantID,
		Email:     email,
		Roles:     roles,
		TokenType: tokenType,
	}

	var signingMethod jwt.SigningMethod
	switch m.config.SigningAlgorithm {
	case "HS256":
		signingMethod = jwt.SigningMethodHS256
	case "HS384":
		signingMethod = jwt.SigningMethodHS384
	case "HS512":
		signingMethod = jwt.SigningMethodHS512
	default:
		signingMethod = jwt.SigningMethodHS256
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	signedToken, err := token.SignedString([]byte(m.config.Secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims.
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		switch m.config.SigningAlgorithm {
		case "HS256":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		case "HS384":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		case "HS512":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		}
		return []byte(m.config.Secret), nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, errors.New(errors.ErrCodeTokenExpired, "token has expired")
		}
		return nil, errors.Wrap(err, errors.ErrCodeTokenInvalid, "invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New(errors.ErrCodeTokenInvalid, "invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != m.config.Issuer {
		return nil, errors.New(errors.ErrCodeTokenInvalid, "invalid token issuer")
	}

	// Validate audience
	if !claims.VerifyAudience(m.config.Audience, true) {
		return nil, errors.New(errors.ErrCodeTokenInvalid, "invalid token audience")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token.
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, errors.New(errors.ErrCodeTokenInvalid, "not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token.
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		if errors.Is(err, errors.ErrCodeTokenExpired) {
			return nil, errors.New(errors.ErrCodeRefreshTokenExpired, "refresh token has expired")
		}
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New(errors.ErrCodeTokenInvalid, "not a refresh token")
	}

	return claims, nil
}

// RefreshTokenPair refreshes a token pair using a valid refresh token.
func (m *JWTManager) RefreshTokenPair(refreshTokenString string) (*TokenPair, error) {
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	return m.GenerateTokenPair(claims.UserID, claims.TenantID, claims.Email, claims.Roles)
}

// generateTokenID generates a unique token ID.
func generateTokenID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Context keys for storing claims
type contextKey string

const (
	claimsKey contextKey = "claims"
)

// ClaimsFromContext extracts claims from context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

// ContextWithClaims returns a context with claims.
func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// UserIDFromContext extracts user ID from context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return "", false
	}
	return claims.UserID, true
}

// TenantIDFromContext extracts tenant ID from context.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return "", false
	}
	return claims.TenantID, true
}

// RolesFromContext extracts roles from context.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return nil, false
	}
	return claims.Roles, true
}

// HasRole checks if the user has a specific role.
func HasRole(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles.
func HasAnyRole(ctx context.Context, roles ...string) bool {
	for _, role := range roles {
		if HasRole(ctx, role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the user has all of the specified roles.
func HasAllRoles(ctx context.Context, roles ...string) bool {
	for _, role := range roles {
		if !HasRole(ctx, role) {
			return false
		}
	}
	return true
}
