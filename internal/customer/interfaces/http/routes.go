// Package http provides HTTP handlers for the Customer service.
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/usecase"
)

// RouterConfig contains configuration for the router.
type RouterConfig struct {
	// Logger for request logging.
	Logger *zap.Logger
	// Middleware configuration.
	MiddlewareConfig *MiddlewareConfig
	// Auth configuration.
	AuthConfig AuthConfig
	// EnableProfiling enables pprof endpoints.
	EnableProfiling bool
	// APIPrefix is the prefix for all API routes.
	APIPrefix string
	// Version is the API version.
	Version string
}

// DefaultRouterConfig returns default router configuration.
func DefaultRouterConfig(logger *zap.Logger) *RouterConfig {
	return &RouterConfig{
		Logger:           logger,
		MiddlewareConfig: DefaultMiddlewareConfig(),
		AuthConfig: AuthConfig{
			SkipPaths: []string{"/health", "/ready", "/metrics"},
		},
		EnableProfiling: false,
		APIPrefix:       "/api",
		Version:         "v1",
	}
}

// Router creates and configures the HTTP router.
type Router struct {
	mux     *chi.Mux
	handler *Handler
	config  *RouterConfig
}

// NewRouter creates a new router with the given configuration.
func NewRouter(handler *Handler, config *RouterConfig) *Router {
	if config == nil {
		config = DefaultRouterConfig(zap.NewNop())
	}

	r := &Router{
		mux:     chi.NewRouter(),
		handler: handler,
		config:  config,
	}

	r.setupMiddleware()
	r.setupRoutes()

	return r
}

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Mux returns the underlying chi.Mux.
func (r *Router) Mux() *chi.Mux {
	return r.mux
}

// setupMiddleware configures global middleware.
func (r *Router) setupMiddleware() {
	// Request ID (first, so all other middleware can use it)
	r.mux.Use(RequestID)

	// Recovery from panics
	r.mux.Use(Recoverer(r.config.Logger))

	// Request logging
	r.mux.Use(RequestLogger(r.config.Logger))

	// Security headers
	r.mux.Use(SecurityHeaders)

	// CORS
	r.mux.Use(CORS(r.config.MiddlewareConfig))

	// Request timeout
	r.mux.Use(Timeout(r.config.MiddlewareConfig.RequestTimeout))

	// Request body size limit
	r.mux.Use(MaxBodySize(r.config.MiddlewareConfig.MaxRequestBodySize))

	// Global rate limiting
	r.mux.Use(RateLimiter(
		r.config.MiddlewareConfig.RateLimitRequests,
		r.config.MiddlewareConfig.RateLimitBurst,
	))

	// Real IP (for logging behind proxies)
	r.mux.Use(middleware.RealIP)

	// Compress responses
	r.mux.Use(middleware.Compress(5))

	// Clean path
	r.mux.Use(middleware.CleanPath)

	// Profiling endpoints (development only)
	if r.config.EnableProfiling {
		r.mux.Mount("/debug", middleware.Profiler())
	}
}

// setupRoutes configures all routes.
func (r *Router) setupRoutes() {
	// Health check endpoints (no auth required)
	r.mux.Get("/health", r.healthCheck)
	r.mux.Get("/ready", r.readinessCheck)
	r.mux.Get("/metrics", r.metricsHandler)

	// API routes
	r.mux.Route(r.config.APIPrefix+"/"+r.config.Version, func(router chi.Router) {
		// Tenant extraction for all API routes
		router.Use(TenantExtractor)

		// Per-tenant rate limiting
		tenantLimiter := NewTenantRateLimiter(
			r.config.MiddlewareConfig.TenantRateLimitRequests,
			r.config.MiddlewareConfig.TenantRateLimitBurst,
		)
		router.Use(tenantLimiter.Middleware)

		// Authentication
		router.Use(Authenticator(r.config.AuthConfig))
		router.Use(RequireAuth)

		// Require JSON content type for POST/PUT/PATCH
		router.Use(RequireContentType("application/json"))

		// Customer routes
		router.Route("/customers", r.customerRoutes)

		// Segment routes
		router.Route("/segments", r.segmentRoutes)

		// Import routes
		router.Route("/imports", r.importRoutes)
	})

	// 404 handler
	r.mux.NotFound(r.notFoundHandler)

	// 405 handler
	r.mux.MethodNotAllowed(r.methodNotAllowedHandler)
}

// customerRoutes sets up customer-related routes.
func (r *Router) customerRoutes(router chi.Router) {
	// Customer CRUD
	router.Post("/", r.handler.CreateCustomer)
	router.Get("/", r.handler.SearchCustomers)
	router.Get("/export", r.handler.ExportCustomers)

	// Single customer operations
	router.Route("/{customerId}", func(router chi.Router) {
		router.Get("/", r.handler.GetCustomer)
		router.Put("/", r.handler.UpdateCustomer)
		router.Delete("/", r.handler.DeleteCustomer)

		// Customer actions
		router.Post("/restore", r.handler.RestoreCustomer)
		router.Post("/activate", r.handler.ActivateCustomer)
		router.Post("/deactivate", r.handler.DeactivateCustomer)
		router.Post("/block", r.handler.BlockCustomer)
		router.Post("/unblock", r.handler.UnblockCustomer)

		// Contact routes (nested under customer)
		router.Route("/contacts", r.contactRoutes)

		// Note routes
		router.Route("/notes", r.noteRoutes)

		// Activity routes
		router.Route("/activities", r.activityRoutes)

		// Segment membership
		router.Post("/segments/{segmentId}", r.handler.AddToSegment)
		router.Delete("/segments/{segmentId}", r.handler.RemoveFromSegment)
	})
}

// contactRoutes sets up contact-related routes.
func (r *Router) contactRoutes(router chi.Router) {
	router.Post("/", r.handler.AddContact)
	router.Get("/", r.handler.ListContacts)

	router.Route("/{contactId}", func(router chi.Router) {
		router.Get("/", r.handler.GetContact)
		router.Put("/", r.handler.UpdateContact)
		router.Delete("/", r.handler.DeleteContact)
		router.Post("/primary", r.handler.SetPrimaryContact)
	})
}

// noteRoutes sets up note-related routes.
func (r *Router) noteRoutes(router chi.Router) {
	router.Post("/", r.handler.AddNote)
	router.Get("/", r.handler.ListNotes)

	router.Route("/{noteId}", func(router chi.Router) {
		router.Get("/", r.handler.GetNote)
		router.Put("/", r.handler.UpdateNote)
		router.Delete("/", r.handler.DeleteNote)
		router.Post("/pin", r.handler.PinNote)
		router.Delete("/pin", r.handler.UnpinNote)
	})
}

// activityRoutes sets up activity-related routes.
func (r *Router) activityRoutes(router chi.Router) {
	router.Post("/", r.handler.LogActivity)
	router.Get("/", r.handler.ListActivities)

	router.Route("/{activityId}", func(router chi.Router) {
		router.Get("/", r.handler.GetActivity)
	})
}

// segmentRoutes sets up segment-related routes.
func (r *Router) segmentRoutes(router chi.Router) {
	router.Post("/", r.handler.CreateSegment)
	router.Get("/", r.handler.ListSegments)

	router.Route("/{segmentId}", func(router chi.Router) {
		router.Get("/", r.handler.GetSegment)
		router.Put("/", r.handler.UpdateSegment)
		router.Delete("/", r.handler.DeleteSegment)
		router.Post("/refresh", r.handler.RefreshSegment)
		router.Get("/customers", r.handler.GetSegmentCustomers)
	})
}

// importRoutes sets up import-related routes.
func (r *Router) importRoutes(router chi.Router) {
	router.Post("/", r.handler.ImportCustomers)
	router.Get("/", r.handler.ListImports)

	router.Route("/{importId}", func(router chi.Router) {
		router.Get("/", r.handler.GetImportStatus)
		router.Get("/errors", r.handler.GetImportErrors)
		router.Delete("/", r.handler.CancelImport)
	})
}

// ============================================================================
// Health and Status Handlers
// ============================================================================

// HealthResponse represents health check response.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// ReadinessResponse represents readiness check response.
type ReadinessResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Checks    map[string]ComponentCheck `json:"checks,omitempty"`
}

// ComponentCheck represents a component health check.
type ComponentCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// healthCheck handles health check requests.
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   r.config.Version,
	})
}

// readinessCheck handles readiness check requests.
func (r *Router) readinessCheck(w http.ResponseWriter, req *http.Request) {
	// In production, this would check database, cache, message queue connections
	checks := map[string]ComponentCheck{
		"database": {
			Status: "healthy",
		},
		"cache": {
			Status: "healthy",
		},
		"message_queue": {
			Status: "healthy",
		},
	}

	allHealthy := true
	for _, check := range checks {
		if check.Status != "healthy" {
			allHealthy = false
			break
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(ReadinessResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	})
}

// metricsHandler handles metrics requests.
func (r *Router) metricsHandler(w http.ResponseWriter, req *http.Request) {
	// In production, this would integrate with Prometheus or other metrics system
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("# Metrics endpoint\n"))
}

// notFoundHandler handles 404 errors.
func (r *Router) notFoundHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Code:    "NOT_FOUND",
		Message: "the requested resource was not found",
	})
}

// methodNotAllowedHandler handles 405 errors.
func (r *Router) methodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Code:    "METHOD_NOT_ALLOWED",
		Message: "the requested method is not allowed for this resource",
	})
}

// ============================================================================
// Handler Factory
// ============================================================================

// HandlerDependencies contains all dependencies needed to create handlers.
type HandlerDependencies struct {
	// Customer use cases
	CreateCustomer     *usecase.CreateCustomerUseCase
	GetCustomer        *usecase.GetCustomerUseCase
	UpdateCustomer     *usecase.UpdateCustomerUseCase
	DeleteCustomer     *usecase.DeleteCustomerUseCase
	SearchCustomers    *usecase.SearchCustomersUseCase
	RestoreCustomer    *usecase.RestoreCustomerUseCase
	ActivateCustomer   *usecase.ActivateCustomerUseCase
	DeactivateCustomer *usecase.DeactivateCustomerUseCase
	BlockCustomer      *usecase.BlockCustomerUseCase
	UnblockCustomer    *usecase.UnblockCustomerUseCase
	ExportCustomers    *usecase.ExportCustomersUseCase
	ImportCustomers    *usecase.ImportCustomersUseCase

	// Contact use cases
	AddContact        *usecase.AddContactUseCase
	GetContact        *usecase.GetContactUseCase
	UpdateContact     *usecase.UpdateContactUseCase
	DeleteContact     *usecase.DeleteContactUseCase
	ListContacts      *usecase.ListContactsUseCase
	SetPrimaryContact *usecase.SetPrimaryContactUseCase

	// Note use cases
	AddNote    *usecase.AddNoteUseCase
	GetNote    *usecase.GetNoteUseCase
	UpdateNote *usecase.UpdateNoteUseCase
	DeleteNote *usecase.DeleteNoteUseCase
	ListNotes  *usecase.ListNotesUseCase
	PinNote    *usecase.PinNoteUseCase
	UnpinNote  *usecase.UnpinNoteUseCase

	// Activity use cases
	LogActivity    *usecase.LogActivityUseCase
	GetActivity    *usecase.GetActivityUseCase
	ListActivities *usecase.ListActivitiesUseCase

	// Segment use cases
	CreateSegment       *usecase.CreateSegmentUseCase
	GetSegment          *usecase.GetSegmentUseCase
	UpdateSegment       *usecase.UpdateSegmentUseCase
	DeleteSegment       *usecase.DeleteSegmentUseCase
	ListSegments        *usecase.ListSegmentsUseCase
	RefreshSegment      *usecase.RefreshSegmentUseCase
	GetSegmentCustomers *usecase.GetSegmentCustomersUseCase
	AddToSegment        *usecase.AddToSegmentUseCase
	RemoveFromSegment   *usecase.RemoveFromSegmentUseCase

	// Import use cases
	GetImportStatus *usecase.GetImportStatusUseCase
	GetImportErrors *usecase.GetImportErrorsUseCase
	ListImports     *usecase.ListImportsUseCase
	CancelImport    *usecase.CancelImportUseCase
}

// NewHandler creates a new handler with all dependencies.
func NewHandler(deps HandlerDependencies) *Handler {
	return &Handler{
		createCustomer:      deps.CreateCustomer,
		getCustomer:         deps.GetCustomer,
		updateCustomer:      deps.UpdateCustomer,
		deleteCustomer:      deps.DeleteCustomer,
		searchCustomers:     deps.SearchCustomers,
		restoreCustomer:     deps.RestoreCustomer,
		activateCustomer:    deps.ActivateCustomer,
		deactivateCustomer:  deps.DeactivateCustomer,
		blockCustomer:       deps.BlockCustomer,
		unblockCustomer:     deps.UnblockCustomer,
		exportCustomers:     deps.ExportCustomers,
		importCustomers:     deps.ImportCustomers,
		addContact:          deps.AddContact,
		getContact:          deps.GetContact,
		updateContact:       deps.UpdateContact,
		deleteContact:       deps.DeleteContact,
		listContacts:        deps.ListContacts,
		setPrimaryContact:   deps.SetPrimaryContact,
		addNote:             deps.AddNote,
		getNote:             deps.GetNote,
		updateNote:          deps.UpdateNote,
		deleteNote:          deps.DeleteNote,
		listNotes:           deps.ListNotes,
		pinNote:             deps.PinNote,
		unpinNote:           deps.UnpinNote,
		logActivity:         deps.LogActivity,
		getActivity:         deps.GetActivity,
		listActivities:      deps.ListActivities,
		createSegment:       deps.CreateSegment,
		getSegment:          deps.GetSegment,
		updateSegment:       deps.UpdateSegment,
		deleteSegment:       deps.DeleteSegment,
		listSegments:        deps.ListSegments,
		refreshSegment:      deps.RefreshSegment,
		getSegmentCustomers: deps.GetSegmentCustomers,
		addToSegment:        deps.AddToSegment,
		removeFromSegment:   deps.RemoveFromSegment,
		getImportStatus:     deps.GetImportStatus,
		getImportErrors:     deps.GetImportErrors,
		listImports:         deps.ListImports,
		cancelImport:        deps.CancelImport,
	}
}
