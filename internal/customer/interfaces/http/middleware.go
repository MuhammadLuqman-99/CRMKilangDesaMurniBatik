// Package http provides HTTP handlers for the Customer service.
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// ContextKey is a type for context keys used in middleware.
type ContextKey string

const (
	// ContextKeyTenantID is the context key for tenant ID.
	ContextKeyTenantID ContextKey = "tenant_id"
	// ContextKeyUserID is the context key for user ID.
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyRequestID is the context key for request ID.
	ContextKeyRequestID ContextKey = "request_id"
	// ContextKeyLogger is the context key for request logger.
	ContextKeyLogger ContextKey = "logger"
	// ContextKeyStartTime is the context key for request start time.
	ContextKeyStartTime ContextKey = "start_time"
)

// TenantFromContext extracts the tenant ID from the context.
func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(ContextKeyTenantID).(uuid.UUID)
	return tenantID, ok
}

// UserFromContext extracts the user ID from the context.
func UserFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return userID, ok
}

// RequestIDFromContext extracts the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	reqID, ok := ctx.Value(ContextKeyRequestID).(string)
	if !ok {
		return ""
	}
	return reqID
}

// LoggerFromContext extracts the logger from the context.
func LoggerFromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(ContextKeyLogger).(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}
	return logger
}

// ============================================================================
// Middleware Configuration
// ============================================================================

// MiddlewareConfig contains configuration for middleware.
type MiddlewareConfig struct {
	// RateLimitRequests is the number of requests per second.
	RateLimitRequests float64
	// RateLimitBurst is the maximum burst size.
	RateLimitBurst int
	// TenantRateLimitRequests is per-tenant rate limit.
	TenantRateLimitRequests float64
	// TenantRateLimitBurst is per-tenant burst size.
	TenantRateLimitBurst int
	// RequestTimeout is the maximum request duration.
	RequestTimeout time.Duration
	// MaxRequestBodySize is the maximum request body size in bytes.
	MaxRequestBodySize int64
	// AllowedOrigins for CORS.
	AllowedOrigins []string
	// AllowedMethods for CORS.
	AllowedMethods []string
	// AllowedHeaders for CORS.
	AllowedHeaders []string
	// ExposedHeaders for CORS.
	ExposedHeaders []string
	// AllowCredentials for CORS.
	AllowCredentials bool
	// MaxAge for CORS preflight cache.
	MaxAge int
}

// DefaultMiddlewareConfig returns the default middleware configuration.
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		RateLimitRequests:       100,
		RateLimitBurst:          200,
		TenantRateLimitRequests: 50,
		TenantRateLimitBurst:    100,
		RequestTimeout:          30 * time.Second,
		MaxRequestBodySize:      10 * 1024 * 1024, // 10MB
		AllowedOrigins:          []string{"*"},
		AllowedMethods:          []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:          []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID", "X-Request-ID"},
		ExposedHeaders:          []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials:        true,
		MaxAge:                  86400,
	}
}

// ============================================================================
// Request ID Middleware
// ============================================================================

// RequestID middleware adds a unique request ID to each request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)
		ctx = context.WithValue(ctx, ContextKeyStartTime, time.Now())

		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ============================================================================
// Logging Middleware
// ============================================================================

// RequestLogger creates a logging middleware with the given logger.
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := RequestIDFromContext(r.Context())

			// Create request-scoped logger
			reqLogger := logger.With(
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)

			ctx := context.WithValue(r.Context(), ContextKeyLogger, reqLogger)

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			defer func() {
				duration := time.Since(start)

				// Log at different levels based on status code
				status := ww.Status()
				fields := []zap.Field{
					zap.Int("status", status),
					zap.Int("bytes", ww.BytesWritten()),
					zap.Duration("duration", duration),
				}

				if tenantID, ok := TenantFromContext(ctx); ok {
					fields = append(fields, zap.String("tenant_id", tenantID.String()))
				}
				if userID, ok := UserFromContext(ctx); ok {
					fields = append(fields, zap.String("user_id", userID.String()))
				}

				switch {
				case status >= 500:
					reqLogger.Error("request completed with server error", fields...)
				case status >= 400:
					reqLogger.Warn("request completed with client error", fields...)
				default:
					reqLogger.Info("request completed", fields...)
				}
			}()

			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}

// ============================================================================
// Recovery Middleware
// ============================================================================

// Recoverer middleware recovers from panics and logs the error.
func Recoverer(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					requestID := RequestIDFromContext(r.Context())

					logger.Error("panic recovered",
						zap.String("request_id", requestID),
						zap.Any("panic", rvr),
						zap.String("stack", string(debug.Stack())),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(ErrorResponse{
						Code:    "INTERNAL_ERROR",
						Message: "an unexpected error occurred",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Tenant Extraction Middleware
// ============================================================================

// TenantExtractor extracts tenant ID from request headers or JWT claims.
func TenantExtractor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			// Try to get from query parameter as fallback
			tenantIDStr = r.URL.Query().Get("tenant_id")
		}

		if tenantIDStr == "" {
			writeError(w, ErrBadRequest("tenant ID is required"))
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			writeError(w, ErrInvalidParameter("tenant_id", "invalid UUID format"))
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyTenantID, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireTenant middleware ensures a valid tenant ID is present.
func RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := TenantFromContext(r.Context()); !ok {
			writeError(w, ErrUnauthorized("tenant context required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// Authentication Middleware
// ============================================================================

// AuthConfig contains authentication configuration.
type AuthConfig struct {
	// JWTSecret is the secret key for JWT validation.
	JWTSecret string
	// JWTIssuer is the expected issuer.
	JWTIssuer string
	// JWTAudience is the expected audience.
	JWTAudience string
	// SkipPaths are paths that don't require authentication.
	SkipPaths []string
}

// Authenticator creates an authentication middleware.
// In production, this would validate JWT tokens from the IAM service.
func Authenticator(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for certain paths
			for _, path := range config.SkipPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, ErrUnauthorized("authorization header required"))
				return
			}

			// Parse Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeError(w, ErrUnauthorized("invalid authorization header format"))
				return
			}

			token := parts[1]

			// In production, validate the JWT token here
			// For now, we'll extract user ID from a header (for development)
			userIDStr := r.Header.Get("X-User-ID")
			if userIDStr == "" {
				// In production, this would be extracted from the validated JWT
				_ = token // Use token for validation
				writeError(w, ErrUnauthorized("invalid token"))
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				writeError(w, ErrUnauthorized("invalid user ID"))
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth middleware ensures the request is authenticated.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserFromContext(r.Context()); !ok {
			writeError(w, ErrUnauthorized("authentication required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// Rate Limiting Middleware
// ============================================================================

// RateLimiter creates a global rate limiting middleware.
func RateLimiter(requestsPerSecond float64, burst int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				w.Header().Set("Retry-After", "1")
				writeError(w, ErrTooManyRequests("rate limit exceeded"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TenantRateLimiter creates a per-tenant rate limiting middleware.
type TenantRateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
	mu       sync.Mutex
}

// NewTenantRateLimiter creates a new tenant rate limiter.
func NewTenantRateLimiter(requestsPerSecond float64, burst int) *TenantRateLimiter {
	return &TenantRateLimiter{
		rate:  rate.Limit(requestsPerSecond),
		burst: burst,
	}
}

// getLimiter gets or creates a rate limiter for a tenant.
func (trl *TenantRateLimiter) getLimiter(tenantID uuid.UUID) *rate.Limiter {
	if limiter, ok := trl.limiters.Load(tenantID); ok {
		return limiter.(*rate.Limiter)
	}

	trl.mu.Lock()
	defer trl.mu.Unlock()

	// Double-check after acquiring lock
	if limiter, ok := trl.limiters.Load(tenantID); ok {
		return limiter.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(trl.rate, trl.burst)
	trl.limiters.Store(tenantID, limiter)
	return limiter
}

// Middleware returns the rate limiting middleware.
func (trl *TenantRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := TenantFromContext(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		limiter := trl.getLimiter(tenantID)
		if !limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			writeError(w, ErrTooManyRequests("tenant rate limit exceeded"))
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(trl.rate)))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%.0f", float64(limiter.Tokens())))

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// Request Timeout Middleware
// ============================================================================

// Timeout creates a request timeout middleware.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					writeError(w, ErrServiceUnavailable("request timeout"))
				}
				return
			}
		})
	}
}

// ============================================================================
// Request Body Size Limiter
// ============================================================================

// MaxBodySize limits the size of request bodies.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// CORS Middleware
// ============================================================================

// CORS creates a CORS middleware.
func CORS(config *MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, o := range config.AllowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Content Type Middleware
// ============================================================================

// ContentTypeJSON ensures responses have JSON content type.
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// RequireContentType ensures requests have the specified content type.
func RequireContentType(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for GET and DELETE requests
			if r.Method == http.MethodGet || r.Method == http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}

			ct := r.Header.Get("Content-Type")
			if ct == "" {
				writeError(w, ErrBadRequest("Content-Type header is required"))
				return
			}

			// Extract media type (ignore charset and other parameters)
			mediaType := strings.Split(ct, ";")[0]
			mediaType = strings.TrimSpace(mediaType)

			if !strings.EqualFold(mediaType, contentType) {
				writeError(w, ErrBadRequest(fmt.Sprintf("Content-Type must be %s", contentType)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Security Headers Middleware
// ============================================================================

// SecurityHeaders adds security headers to responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Clickjacking protection
		w.Header().Set("X-Frame-Options", "DENY")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// HSTS (only for HTTPS)
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// Health Check Bypass
// ============================================================================

// HealthCheckBypass skips middleware for health check endpoints.
func HealthCheckBypass(healthPaths []string, middleware func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range healthPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}
			middleware(next).ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// writeError writes an error response.
func writeError(w http.ResponseWriter, err *ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	_ = json.NewEncoder(w).Encode(err)
}
