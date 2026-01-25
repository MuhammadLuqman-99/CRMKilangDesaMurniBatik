package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ============================================================================
// Context Keys
// ============================================================================

type contextKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey contextKey = "tenant_id"
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
	// UserRolesKey is the context key for user roles
	UserRolesKey contextKey = "user_roles"
	// UserPermissionsKey is the context key for user permissions
	UserPermissionsKey contextKey = "user_permissions"
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
)

// ============================================================================
// JWT Claims
// ============================================================================

// JWTClaims represents the custom JWT claims structure
type JWTClaims struct {
	jwt.RegisteredClaims
	TenantID    uuid.UUID `json:"tenant_id"`
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
}

// ============================================================================
// Middleware Configuration
// ============================================================================

// MiddlewareConfig holds middleware configuration
type MiddlewareConfig struct {
	JWTSecret          string
	JWTIssuer          string
	JWTAudience        string
	SkipAuth           bool
	AllowedOrigins     []string
	RateLimitRequests  int
	RateLimitWindow    time.Duration
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		JWTSecret:         "your-secret-key-change-in-production",
		JWTIssuer:         "crm-kilang",
		JWTAudience:       "crm-kilang-api",
		SkipAuth:          false,
		AllowedOrigins:    []string{"*"},
		RateLimitRequests: 100,
		RateLimitWindow:   time.Minute,
	}
}

// SetMiddlewareConfig sets the middleware configuration
func (h *Handler) SetMiddlewareConfig(config MiddlewareConfig) {
	h.middlewareConfig = config
}

// ============================================================================
// Authentication Middleware
// ============================================================================

// AuthMiddleware validates JWT tokens and extracts user information
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if configured
		if h.middlewareConfig.SkipAuth {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			h.respondError(w, ErrUnauthorized("missing authorization header"))
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			h.respondError(w, ErrUnauthorized("invalid authorization header format"))
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(h.middlewareConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			h.respondError(w, ErrUnauthorized("invalid or expired token"))
			return
		}

		// Validate claims
		if claims.TenantID == uuid.Nil {
			h.respondError(w, ErrUnauthorized("invalid tenant in token"))
			return
		}

		if claims.UserID == uuid.Nil {
			h.respondError(w, ErrUnauthorized("invalid user in token"))
			return
		}

		// Add claims to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
		ctx = context.WithValue(ctx, UserPermissionsKey, claims.Permissions)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ============================================================================
// Tenant Middleware
// ============================================================================

// TenantMiddleware ensures tenant context is set
func (h *Handler) TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Check if tenant ID is already set (from auth middleware)
		if tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok && tenantID != uuid.Nil {
			next.ServeHTTP(w, r)
			return
		}

		// Try to get tenant ID from header (for internal service calls)
		tenantHeader := r.Header.Get("X-Tenant-ID")
		if tenantHeader != "" {
			tenantID, err := uuid.Parse(tenantHeader)
			if err != nil {
				h.respondError(w, ErrBadRequest("invalid tenant ID in header"))
				return
			}
			ctx = context.WithValue(ctx, TenantIDKey, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		h.respondError(w, ErrUnauthorized("tenant identification required"))
	})
}

// ============================================================================
// Permission Middleware
// ============================================================================

// RequirePermission creates middleware that checks for a specific permission
func (h *Handler) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			permissions, ok := ctx.Value(UserPermissionsKey).([]string)
			if !ok {
				h.respondError(w, ErrForbidden("no permissions found"))
				return
			}

			// Check for wildcard permission or specific permission
			hasPermission := false
			for _, p := range permissions {
				if p == "*" || p == permission {
					hasPermission = true
					break
				}
				// Check for wildcard resource permissions (e.g., "leads:*" matches "leads:read")
				if strings.HasSuffix(p, ":*") {
					resource := strings.TrimSuffix(p, ":*")
					if strings.HasPrefix(permission, resource+":") {
						hasPermission = true
						break
					}
				}
				// Check for wildcard action permissions (e.g., "*:read" matches "leads:read")
				if strings.HasPrefix(p, "*:") {
					action := strings.TrimPrefix(p, "*:")
					if strings.HasSuffix(permission, ":"+action) {
						hasPermission = true
						break
					}
				}
			}

			if !hasPermission {
				h.respondError(w, ErrForbidden("insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates middleware that checks for a specific role
func (h *Handler) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			roles, ok := ctx.Value(UserRolesKey).([]string)
			if !ok {
				h.respondError(w, ErrForbidden("no roles found"))
				return
			}

			hasRole := false
			for _, r := range roles {
				if r == role || r == "super_admin" {
					hasRole = true
					break
				}
			}

			if !hasRole {
				h.respondError(w, ErrForbidden("insufficient role"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole creates middleware that checks for any of the specified roles
func (h *Handler) RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userRoles, ok := ctx.Value(UserRolesKey).([]string)
			if !ok {
				h.respondError(w, ErrForbidden("no roles found"))
				return
			}

			hasRole := false
			for _, userRole := range userRoles {
				if userRole == "super_admin" {
					hasRole = true
					break
				}
				for _, requiredRole := range roles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				h.respondError(w, ErrForbidden("insufficient role"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// CORS Middleware
// ============================================================================

// CORSMiddleware handles CORS headers
func (h *Handler) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		for _, o := range h.middlewareConfig.AllowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Tenant-ID, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-Total-Count")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// Request ID Middleware
// ============================================================================

// RequestIDMiddleware adds a unique request ID to each request
func (h *Handler) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ============================================================================
// Rate Limiting Middleware (Simple In-Memory Implementation)
// ============================================================================

// RateLimiter provides simple rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean old requests
	var validRequests []time.Time
	for _, t := range rl.requests[key] {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	rl.requests[key] = append(validRequests, now)
	return true
}

// RateLimitMiddleware applies rate limiting per tenant
func (h *Handler) RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get tenant ID for rate limiting key
			key := "anonymous"
			if tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
				key = tenantID.String()
			}

			if !limiter.Allow(key) {
				h.respondError(w, &ErrorResponse{
					StatusCode: http.StatusTooManyRequests,
					Code:       "RATE_LIMIT_EXCEEDED",
					Message:    "too many requests, please try again later",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Security Headers Middleware
// ============================================================================

// SecurityHeadersMiddleware adds security headers to responses
func (h *Handler) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// Logging Middleware
// ============================================================================

// RequestLogger is a custom response writer that captures status code
type RequestLogger struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(w http.ResponseWriter) *RequestLogger {
	return &RequestLogger{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code
func (rl *RequestLogger) WriteHeader(code int) {
	rl.statusCode = code
	rl.ResponseWriter.WriteHeader(code)
}

// Write captures the response size
func (rl *RequestLogger) Write(b []byte) (int, error) {
	n, err := rl.ResponseWriter.Write(b)
	rl.bytes += n
	return n, err
}

// ============================================================================
// Context Helpers
// ============================================================================

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	return tenantID, ok
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetUserRolesFromContext extracts user roles from context
func GetUserRolesFromContext(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(UserRolesKey).([]string)
	return roles, ok
}

// GetUserPermissionsFromContext extracts user permissions from context
func GetUserPermissionsFromContext(ctx context.Context) ([]string, bool) {
	permissions, ok := ctx.Value(UserPermissionsKey).([]string)
	return permissions, ok
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}
