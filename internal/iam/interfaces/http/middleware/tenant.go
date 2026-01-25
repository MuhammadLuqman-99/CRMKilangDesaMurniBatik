// Package middleware contains HTTP middleware implementations.
package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	iamhttp "github.com/kilang-desa-murni/crm/internal/iam/interfaces/http"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// TenantResolver interface for resolving tenant from request.
type TenantResolver interface {
	ResolveBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	ResolveByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
}

// TenantContext middleware for multi-tenant applications.
type TenantContext struct {
	resolver TenantResolver
}

// NewTenantContext creates a new tenant context middleware.
func NewTenantContext(resolver TenantResolver) *TenantContext {
	return &TenantContext{resolver: resolver}
}

// RequireTenant returns middleware that requires a valid tenant.
func (m *TenantContext) RequireTenant() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get tenant from various sources

			// 1. From X-Tenant-ID header
			tenantIDStr := r.Header.Get("X-Tenant-ID")
			if tenantIDStr != "" {
				tenantID, err := uuid.Parse(tenantIDStr)
				if err != nil {
					iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid tenant ID format", nil)
					return
				}

				tenant, err := m.resolver.ResolveByID(r.Context(), tenantID)
				if err != nil {
					iamhttp.WriteError(w, http.StatusNotFound, iamhttp.ErrCodeNotFound, "tenant not found", nil)
					return
				}

				if !tenant.IsActive() {
					iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, "tenant is not active", nil)
					return
				}

				ctx := context.WithValue(r.Context(), TenantIDKey, tenant.GetID())
				ctx = context.WithValue(ctx, "tenant", tenant)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 2. From X-Tenant-Slug header
			tenantSlug := r.Header.Get("X-Tenant-Slug")
			if tenantSlug != "" {
				tenant, err := m.resolver.ResolveBySlug(r.Context(), tenantSlug)
				if err != nil {
					iamhttp.WriteError(w, http.StatusNotFound, iamhttp.ErrCodeNotFound, "tenant not found", nil)
					return
				}

				if !tenant.IsActive() {
					iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, "tenant is not active", nil)
					return
				}

				ctx := context.WithValue(r.Context(), TenantIDKey, tenant.GetID())
				ctx = context.WithValue(ctx, "tenant", tenant)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// 3. From authenticated user's token claims
			claims := GetTokenClaims(r.Context())
			if claims != nil && claims.TenantID != uuid.Nil {
				tenant, err := m.resolver.ResolveByID(r.Context(), claims.TenantID)
				if err != nil {
					iamhttp.WriteError(w, http.StatusNotFound, iamhttp.ErrCodeNotFound, "tenant not found", nil)
					return
				}

				if !tenant.IsActive() {
					iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, "tenant is not active", nil)
					return
				}

				ctx := context.WithValue(r.Context(), "tenant", tenant)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No tenant found
			iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "tenant context required", nil)
		})
	}
}

// OptionalTenant returns middleware that optionally resolves tenant.
func (m *TenantContext) OptionalTenant() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Try X-Tenant-ID header
			tenantIDStr := r.Header.Get("X-Tenant-ID")
			if tenantIDStr != "" {
				tenantID, err := uuid.Parse(tenantIDStr)
				if err == nil {
					tenant, err := m.resolver.ResolveByID(ctx, tenantID)
					if err == nil && tenant.IsActive() {
						ctx = context.WithValue(ctx, TenantIDKey, tenant.GetID())
						ctx = context.WithValue(ctx, "tenant", tenant)
					}
				}
			}

			// Try X-Tenant-Slug header
			if tenantIDStr == "" {
				tenantSlug := r.Header.Get("X-Tenant-Slug")
				if tenantSlug != "" {
					tenant, err := m.resolver.ResolveBySlug(ctx, tenantSlug)
					if err == nil && tenant.IsActive() {
						ctx = context.WithValue(ctx, TenantIDKey, tenant.GetID())
						ctx = context.WithValue(ctx, "tenant", tenant)
					}
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ValidateTenantAccess ensures authenticated user belongs to the requested tenant.
func (m *TenantContext) ValidateTenantAccess() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get tenant from context (set by RequireTenant)
			tenant := GetTenant(r.Context())
			if tenant == nil {
				iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "tenant context required", nil)
				return
			}

			// Get authenticated user's tenant
			claims := GetTokenClaims(r.Context())
			if claims == nil {
				iamhttp.WriteError(w, http.StatusUnauthorized, iamhttp.ErrCodeUnauthorized, "authentication required", nil)
				return
			}

			// Validate user belongs to requested tenant
			if claims.TenantID != tenant.GetID() {
				iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, "access denied to this tenant", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetTenant extracts the tenant from context.
func GetTenant(ctx context.Context) *domain.Tenant {
	if tenant, ok := ctx.Value("tenant").(*domain.Tenant); ok {
		return tenant
	}
	return nil
}

// TenantFromPathResolver creates middleware that resolves tenant from path parameter.
func TenantFromPathResolver(resolver TenantResolver, paramName string, getParam func(*http.Request, string) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantSlug := getParam(r, paramName)
			if tenantSlug == "" {
				iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "tenant identifier required", nil)
				return
			}

			// Try to parse as UUID first
			tenantID, err := uuid.Parse(tenantSlug)
			var tenant *domain.Tenant

			if err == nil {
				tenant, err = resolver.ResolveByID(r.Context(), tenantID)
			} else {
				tenant, err = resolver.ResolveBySlug(r.Context(), tenantSlug)
			}

			if err != nil {
				iamhttp.WriteError(w, http.StatusNotFound, iamhttp.ErrCodeNotFound, "tenant not found", nil)
				return
			}

			if !tenant.IsActive() {
				iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, "tenant is not active", nil)
				return
			}

			ctx := context.WithValue(r.Context(), TenantIDKey, tenant.GetID())
			ctx = context.WithValue(ctx, "tenant", tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
