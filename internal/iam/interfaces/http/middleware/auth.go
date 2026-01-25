// Package middleware contains HTTP middleware implementations.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	iamhttp "github.com/kilang-desa-murni/crm/internal/iam/interfaces/http"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
)

// Context keys for storing auth information
type contextKey string

const (
	UserIDKey       contextKey = "user_id"
	TenantIDKey     contextKey = "tenant_id"
	EmailKey        contextKey = "email"
	RolesKey        contextKey = "roles"
	PermissionsKey  contextKey = "permissions"
	TokenClaimsKey  contextKey = "token_claims"
	IsAuthenticatedKey contextKey = "is_authenticated"
)

// AuthMiddleware provides authentication middleware.
type AuthMiddleware struct {
	tokenService   ports.TokenService
	tokenBlacklist ports.TokenBlacklist
}

// NewAuthMiddleware creates a new auth middleware.
func NewAuthMiddleware(tokenService ports.TokenService, tokenBlacklist ports.TokenBlacklist) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService:   tokenService,
		tokenBlacklist: tokenBlacklist,
	}
}

// RequireAuth returns middleware that requires authentication.
func (m *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				iamhttp.WriteError(w, http.StatusUnauthorized, iamhttp.ErrCodeUnauthorized, "missing or invalid authorization header", nil)
				return
			}

			// Check if token is blacklisted
			if m.tokenBlacklist != nil {
				blacklisted, err := m.tokenBlacklist.IsBlacklisted(r.Context(), token)
				if err != nil {
					iamhttp.WriteError(w, http.StatusInternalServerError, iamhttp.ErrCodeInternalServer, "failed to verify token", nil)
					return
				}
				if blacklisted {
					iamhttp.WriteError(w, http.StatusUnauthorized, iamhttp.ErrCodeUnauthorized, "token has been revoked", nil)
					return
				}
			}

			// Validate token
			claims, err := m.tokenService.ValidateAccessToken(token)
			if err != nil {
				iamhttp.WriteError(w, http.StatusUnauthorized, iamhttp.ErrCodeUnauthorized, "invalid or expired token", nil)
				return
			}

			// Add claims to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, RolesKey, claims.Roles)
			ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)
			ctx = context.WithValue(ctx, TokenClaimsKey, claims)
			ctx = context.WithValue(ctx, IsAuthenticatedKey, true)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth returns middleware that optionally authenticates.
func (m *AuthMiddleware) OptionalAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				// No token, continue without auth
				ctx := context.WithValue(r.Context(), IsAuthenticatedKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Try to validate token
			claims, err := m.tokenService.ValidateAccessToken(token)
			if err != nil {
				// Invalid token, continue without auth
				ctx := context.WithValue(r.Context(), IsAuthenticatedKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Add claims to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, RolesKey, claims.Roles)
			ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)
			ctx = context.WithValue(ctx, TokenClaimsKey, claims)
			ctx = context.WithValue(ctx, IsAuthenticatedKey, true)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns middleware that requires specific permissions.
func (m *AuthMiddleware) RequirePermission(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userPermissions := GetPermissions(r.Context())
			if !hasAnyPermission(userPermissions, permissions) {
				iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, "insufficient permissions", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllPermissions returns middleware that requires all specified permissions.
func (m *AuthMiddleware) RequireAllPermissions(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userPermissions := GetPermissions(r.Context())
			if !hasAllPermissions(userPermissions, permissions) {
				iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, "insufficient permissions", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns middleware that requires specific roles.
func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles := GetRoles(r.Context())
			if !hasAnyRole(userRoles, roles) {
				iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, "insufficient role", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions to extract auth info from context

// GetUserID extracts user ID from context.
func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetTenantID extracts tenant ID from context.
func GetTenantID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// GetEmail extracts email from context.
func GetEmail(ctx context.Context) string {
	if email, ok := ctx.Value(EmailKey).(string); ok {
		return email
	}
	return ""
}

// GetRoles extracts roles from context.
func GetRoles(ctx context.Context) []string {
	if roles, ok := ctx.Value(RolesKey).([]string); ok {
		return roles
	}
	return nil
}

// GetPermissions extracts permissions from context.
func GetPermissions(ctx context.Context) []string {
	if perms, ok := ctx.Value(PermissionsKey).([]string); ok {
		return perms
	}
	return nil
}

// GetTokenClaims extracts token claims from context.
func GetTokenClaims(ctx context.Context) *ports.TokenClaims {
	if claims, ok := ctx.Value(TokenClaimsKey).(*ports.TokenClaims); ok {
		return claims
	}
	return nil
}

// IsAuthenticated checks if the request is authenticated.
func IsAuthenticated(ctx context.Context) bool {
	if authenticated, ok := ctx.Value(IsAuthenticatedKey).(bool); ok {
		return authenticated
	}
	return false
}

// Helper functions

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return parts[1]
}

func hasAnyPermission(userPerms, requiredPerms []string) bool {
	permSet := make(map[string]bool)
	for _, p := range userPerms {
		permSet[p] = true
		// Handle wildcards
		if strings.HasSuffix(p, ":*") {
			resource := strings.TrimSuffix(p, ":*")
			permSet[resource+":*"] = true
		}
		if p == "*:*" {
			return true // Super admin
		}
	}

	for _, required := range requiredPerms {
		if permSet[required] {
			return true
		}
		// Check wildcard match
		parts := strings.SplitN(required, ":", 2)
		if len(parts) == 2 {
			if permSet[parts[0]+":*"] {
				return true
			}
		}
	}

	return false
}

func hasAllPermissions(userPerms, requiredPerms []string) bool {
	for _, required := range requiredPerms {
		if !hasAnyPermission(userPerms, []string{required}) {
			return false
		}
	}
	return true
}

func hasAnyRole(userRoles, requiredRoles []string) bool {
	roleSet := make(map[string]bool)
	for _, r := range userRoles {
		roleSet[strings.ToLower(r)] = true
	}

	for _, required := range requiredRoles {
		if roleSet[strings.ToLower(required)] {
			return true
		}
	}

	return false
}
