// Package middleware provides HTTP middleware utilities for the CRM application.
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/pkg/auth"
	"github.com/kilang-desa-murni/crm/pkg/errors"
	"github.com/kilang-desa-murni/crm/pkg/logger"
	"github.com/kilang-desa-murni/crm/pkg/response"
)

// Context keys
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TenantIDKey  contextKey = "tenant_id"
	UserIDKey    contextKey = "user_id"
	StartTimeKey contextKey = "start_time"
)

// RequestID adds a unique request ID to each request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists in header
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext extracts the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// Logger logs each request with relevant information.
func Logger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Add start time to context
			ctx := context.WithValue(r.Context(), StartTimeKey, start)

			// Get logger with context fields
			requestID := RequestIDFromContext(r.Context())
			reqLog := log.With().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				RequestID(requestID).
				Logger()

			// Add logger to context
			ctx = reqLog.WithContext(ctx)

			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Log request completion
			duration := time.Since(start)
			event := reqLog.Info()
			if wrapped.statusCode >= 400 {
				event = reqLog.Warn()
			}
			if wrapped.statusCode >= 500 {
				event = reqLog.Error()
			}

			event.
				Int("status", wrapped.statusCode).
				Dur("duration", duration).
				Int("bytes", wrapped.bytesWritten).
				Str("user_agent", r.UserAgent()).
				Msg("Request completed")
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

// Recover recovers from panics and returns a 500 error.
func Recover(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error().
						Interface("panic", err).
						Str("path", r.URL.Path).
						Str("method", r.Method).
						Stack().
						Msg("Panic recovered")

					appErr := errors.ErrInternal("An internal error occurred")
					response.Error(w, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS handles Cross-Origin Resource Sharing.
func CORS(allowedOrigins []string, allowedMethods []string, allowedHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Auth authenticates requests using JWT tokens.
func Auth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, errors.ErrUnauthorized("Missing authorization header"))
				return
			}

			// Check for Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Error(w, errors.ErrUnauthorized("Invalid authorization header format"))
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := jwtManager.ValidateAccessToken(tokenString)
			if err != nil {
				response.Error(w, err)
				return
			}

			// Add claims to context
			ctx := auth.ContextWithClaims(r.Context(), claims)

			// Add tenant ID to context
			ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)

			// Add user ID to context
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantIDFromContext extracts the tenant ID from context.
func TenantIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(TenantIDKey).(string); ok {
		return id
	}
	return ""
}

// UserIDFromContext extracts the user ID from context.
func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// RequireRoles ensures the user has at least one of the specified roles.
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !auth.HasAnyRole(r.Context(), roles...) {
				response.Error(w, errors.ErrForbidden("Insufficient permissions"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllRoles ensures the user has all of the specified roles.
func RequireAllRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !auth.HasAllRoles(r.Context(), roles...) {
				response.Error(w, errors.ErrForbidden("Insufficient permissions"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout adds a timeout to the request context.
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
					response.Error(w, errors.ErrTimeout("Request timed out"))
				}
			}
		})
	}
}

// ContentType ensures the request has the expected content type.
func ContentType(contentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				ct := r.Header.Get("Content-Type")
				if ct == "" {
					response.Error(w, errors.ErrBadRequest("Content-Type header is required"))
					return
				}

				// Check content type (ignore charset and other parameters)
				if !strings.HasPrefix(ct, contentType) {
					response.Error(w, errors.ErrBadRequest("Invalid Content-Type"))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Chain chains multiple middleware together.
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
