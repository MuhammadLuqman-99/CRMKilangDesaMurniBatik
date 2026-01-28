// Package http contains the HTTP interface layer for the Sales Pipeline service.
package http

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Idempotency Middleware
// ============================================================================

// IdempotencyConfig holds configuration for idempotency middleware.
type IdempotencyConfig struct {
	// KeyHeader is the header name for the idempotency key.
	// Default: "Idempotency-Key"
	KeyHeader string

	// KeyTTL is how long to store idempotency keys.
	// Default: 24 hours
	KeyTTL time.Duration

	// RequireKey when true, requests without idempotency key are rejected.
	// Default: false
	RequireKey bool

	// Methods are the HTTP methods that require idempotency.
	// Default: POST, PUT, PATCH
	Methods []string
}

// DefaultIdempotencyConfig returns the default idempotency configuration.
func DefaultIdempotencyConfig() IdempotencyConfig {
	return IdempotencyConfig{
		KeyHeader:  "Idempotency-Key",
		KeyTTL:     24 * time.Hour,
		RequireKey: false,
		Methods:    []string{"POST", "PUT", "PATCH"},
	}
}

// IdempotencyMiddleware provides idempotency key handling for requests.
type IdempotencyMiddleware struct {
	idempotencyRepo domain.IdempotencyRepository
	config          IdempotencyConfig
}

// NewIdempotencyMiddleware creates a new idempotency middleware.
func NewIdempotencyMiddleware(
	idempotencyRepo domain.IdempotencyRepository,
	config IdempotencyConfig,
) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		idempotencyRepo: idempotencyRepo,
		config:          config,
	}
}

// Middleware returns the HTTP middleware handler.
func (m *IdempotencyMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this method requires idempotency
		if !m.requiresIdempotency(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		// Get idempotency key from header
		idempotencyKey := r.Header.Get(m.config.KeyHeader)

		// If no key provided
		if idempotencyKey == "" {
			if m.config.RequireKey {
				m.respondError(w, http.StatusBadRequest, "missing_idempotency_key",
					"Idempotency-Key header is required for this request")
				return
			}
			// Generate a key from request path and body hash
			idempotencyKey = m.generateKey(r)
		}

		// Get tenant ID from context
		tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
		if !ok || tenantID == uuid.Nil {
			// No tenant, proceed without idempotency
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		// Check for existing idempotency key
		existing, err := m.idempotencyRepo.Get(ctx, tenantID, idempotencyKey)
		if err == nil && existing != nil {
			// Key exists - check status
			if existing.Status == domain.IdempotencyStatusProcessing {
				// Request is still being processed
				m.respondError(w, http.StatusConflict, "request_in_progress",
					"A request with this idempotency key is currently being processed")
				return
			}

			if existing.Status == domain.IdempotencyStatusCompleted {
				// Return cached response
				if existing.ResponseCode != nil && existing.ResponseBody != nil {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Idempotency-Replayed", "true")
					w.WriteHeader(*existing.ResponseCode)
					w.Write(existing.ResponseBody)
					return
				}
			}
		}

		// Create idempotency key as processing
		idemKey := domain.NewIdempotencyKey(tenantID, idempotencyKey, m.config.KeyTTL)
		idemKey.MarkProcessing()

		// Store the key (best-effort)
		_ = m.idempotencyRepo.Store(ctx, idemKey)

		// Create response wrapper to capture the response
		wrapper := NewIdempotencyResponseWrapper(w)

		// Process the request
		next.ServeHTTP(wrapper, r)

		// Store the response for replay
		if wrapper.statusCode >= 200 && wrapper.statusCode < 300 {
			idemKey.MarkCompleted(&wrapper.statusCode, wrapper.body.Bytes())
		} else {
			idemKey.MarkFailed(&wrapper.statusCode, wrapper.body.Bytes())
		}

		// Update the idempotency key with the response (best-effort)
		_ = m.idempotencyRepo.Store(ctx, idemKey)
	})
}

// respondError writes an error JSON response.
func (m *IdempotencyMiddleware) respondError(w http.ResponseWriter, status int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   errorCode,
		"message": message,
	})
}

// requiresIdempotency checks if the HTTP method requires idempotency handling.
func (m *IdempotencyMiddleware) requiresIdempotency(method string) bool {
	for _, mtd := range m.config.Methods {
		if mtd == method {
			return true
		}
	}
	return false
}

// generateKey generates an idempotency key from the request.
func (m *IdempotencyMiddleware) generateKey(r *http.Request) string {
	// Create a hash from path + method + timestamp (rounded to minute)
	// This isn't true idempotency but provides some protection
	timestamp := time.Now().Truncate(time.Minute).Unix()
	data := r.Method + r.URL.Path + string(rune(timestamp))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// ============================================================================
// Response Wrapper for Idempotency
// ============================================================================

// IdempotencyResponseWrapper wraps http.ResponseWriter to capture the response.
type IdempotencyResponseWrapper struct {
	http.ResponseWriter
	statusCode int
	body       *idempotencyBuffer
	written    bool
}

// idempotencyBuffer is a simple buffer for response body.
type idempotencyBuffer struct {
	data []byte
}

func (b *idempotencyBuffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *idempotencyBuffer) Bytes() []byte {
	return b.data
}

// NewIdempotencyResponseWrapper creates a new response wrapper.
func NewIdempotencyResponseWrapper(w http.ResponseWriter) *IdempotencyResponseWrapper {
	return &IdempotencyResponseWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &idempotencyBuffer{},
	}
}

// WriteHeader captures the status code.
func (rw *IdempotencyResponseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body.
func (rw *IdempotencyResponseWrapper) Write(b []byte) (int, error) {
	rw.body.Write(b)
	rw.written = true
	return rw.ResponseWriter.Write(b)
}
