// Package http contains HTTP interface implementations.
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/kilang-desa-murni/crm/internal/iam/interfaces/http/handler"
	"github.com/kilang-desa-murni/crm/internal/iam/interfaces/http/middleware"
)

// RouterConfig holds configuration for the router.
type RouterConfig struct {
	Logger             *slog.Logger
	RateLimitRequests  int
	RequestTimeout     time.Duration
	EnableCORS         bool
	CORSConfig         middleware.CORSConfig
}

// DefaultRouterConfig returns default router configuration.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		Logger:             slog.Default(),
		RateLimitRequests:  60,
		RequestTimeout:     30 * time.Second,
		EnableCORS:         true,
		CORSConfig:         middleware.DefaultCORSConfig(),
	}
}

// Handlers holds all HTTP handlers.
type Handlers struct {
	Auth   *handler.AuthHandler
	User   *handler.UserHandler
	Role   *handler.RoleHandler
	Tenant *handler.TenantHandler
}

// Middlewares holds all middleware instances.
type Middlewares struct {
	Auth   *middleware.AuthMiddleware
	Tenant *middleware.TenantContext
}

// NewRouter creates a new HTTP router with all routes configured.
func NewRouter(config RouterConfig, handlers Handlers, middlewares Middlewares) *chi.Mux {
	r := chi.NewRouter()

	// Apply global middleware
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(config.Logger))
	r.Use(middleware.Recovery(config.Logger))
	r.Use(middleware.SecurityHeaders())

	if config.EnableCORS {
		r.Use(middleware.CORS(config.CORSConfig))
	}

	r.Use(middleware.Timeout(config.RequestTimeout))

	// Rate limiting
	rateLimiter := middleware.NewInMemoryRateLimiter(5 * time.Minute)
	r.Use(middleware.RateLimit(rateLimiter, middleware.RateLimitConfig{
		RequestsPerMinute: config.RateLimitRequests,
		KeyFunc:           middleware.IPBasedKey,
	}))

	// Health check endpoint
	r.Get("/health", healthHandler)
	r.Get("/ready", readyHandler)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.Auth.Register)
			r.Post("/login", handlers.Auth.Login)
			r.Post("/refresh", handlers.Auth.RefreshToken)

			// Authenticated auth routes
			r.Group(func(r chi.Router) {
				r.Use(middlewares.Auth.RequireAuth())
				r.Post("/logout", handlers.Auth.Logout)
				r.Get("/me", handlers.Auth.Me)
				r.Put("/password", handlers.Auth.ChangePassword)
			})
		})

		// Tenant routes
		r.Route("/tenants", func(r chi.Router) {
			// Public tenant endpoints
			r.Get("/check-slug", handlers.Tenant.CheckSlugAvailability)
			r.Get("/by-slug/{slug}", handlers.Tenant.GetBySlug)

			// Authenticated tenant management
			r.Group(func(r chi.Router) {
				r.Use(middlewares.Auth.RequireAuth())

				r.Get("/", handlers.Tenant.List)
				r.Post("/", handlers.Tenant.Create)
				r.Get("/{id}", handlers.Tenant.Get)
				r.Put("/{id}", handlers.Tenant.Update)
				r.Delete("/{id}", handlers.Tenant.Delete)
				r.Put("/{id}/status", handlers.Tenant.UpdateStatus)
				r.Put("/{id}/plan", handlers.Tenant.UpdatePlan)
				r.Get("/{id}/stats", handlers.Tenant.GetStats)
			})
		})

		// User routes (require tenant context)
		r.Route("/users", func(r chi.Router) {
			r.Use(middlewares.Auth.RequireAuth())

			r.Get("/", handlers.User.List)
			r.Get("/{id}", handlers.User.Get)
			r.Put("/{id}", handlers.User.Update)
			r.Delete("/{id}", handlers.User.Delete)
			r.Put("/{id}/status", handlers.User.UpdateStatus)
			r.Get("/{id}/roles", handlers.User.GetRoles)
			r.Post("/{id}/roles", handlers.User.AssignRole)
			r.Delete("/{id}/roles", handlers.User.RemoveRole)
			r.Get("/{id}/permissions", handlers.User.GetPermissions)
		})

		// Role routes (require tenant context)
		r.Route("/roles", func(r chi.Router) {
			r.Use(middlewares.Auth.RequireAuth())

			r.Get("/", handlers.Role.List)
			r.Post("/", handlers.Role.Create)
			r.Get("/system", handlers.Role.GetSystemRoles)
			r.Get("/{id}", handlers.Role.Get)
			r.Put("/{id}", handlers.Role.Update)
			r.Delete("/{id}", handlers.Role.Delete)
			r.Get("/{id}/users", handlers.Role.GetRoleUsers)
			r.Post("/{id}/permissions", handlers.Role.AddPermissions)
			r.Delete("/{id}/permissions", handlers.Role.RemovePermissions)
		})
	})

	return r
}

// healthHandler handles health check requests.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteSuccess(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// readyHandler handles readiness check requests.
func readyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Check database and Redis connectivity
	WriteSuccess(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// GetPathParam is a helper to extract path parameters from chi router.
func GetPathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

// PathParamFunc returns a function that extracts path parameters.
func PathParamFunc() func(*http.Request, string) string {
	return GetPathParam
}
