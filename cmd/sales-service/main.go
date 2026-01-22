// Sales Service - Sales Pipeline Management
// ==========================================
// This service handles leads, opportunities, deals, and sales pipelines.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kilang-desa-murni/crm/pkg/auth"
	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/database"
	"github.com/kilang-desa-murni/crm/pkg/events"
	"github.com/kilang-desa-murni/crm/pkg/logger"
	"github.com/kilang-desa-murni/crm/pkg/middleware"
	"github.com/kilang-desa-murni/crm/pkg/response"
	"github.com/kilang-desa-murni/crm/pkg/tracer"
)

// Version information (set during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override service-specific settings
	cfg.App.Name = "sales-service"
	cfg.Server.Port = 8083

	// Initialize logger
	log := logger.New(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
		Caller: cfg.Logger.Caller,
	})
	log = log.With().Service(cfg.App.Name).Logger()
	logger.SetGlobal(log)

	log.Info().
		Str("version", Version).
		Str("build_time", BuildTime).
		Str("git_commit", GitCommit).
		Msg("Starting Sales service")

	// Initialize tracer
	tr, err := tracer.New(&cfg.Tracer, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tracer")
	}
	defer tr.Close(context.Background())

	// Initialize PostgreSQL
	db, err := database.NewPostgres(&cfg.Database, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer db.Close()

	// Initialize Redis
	redis, err := database.NewRedis(&cfg.Redis, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()

	// Initialize RabbitMQ Event Bus
	eventBus, err := events.NewRabbitMQEventBus(&cfg.RabbitMQ, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer eventBus.Close()

	// Initialize JWT Manager for token validation
	jwtManager := auth.NewJWTManager(&cfg.JWT)

	// Create HTTP router
	mux := http.NewServeMux()

	// Health check endpoint
	startTime := time.Now()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]response.HealthCheck)

		// Check PostgreSQL
		if err := db.Health(r.Context()); err != nil {
			checks["postgresql"] = response.HealthCheck{Status: "unhealthy", Message: err.Error()}
		} else {
			checks["postgresql"] = response.HealthCheck{Status: "healthy"}
		}

		// Check Redis
		if err := redis.Health(r.Context()); err != nil {
			checks["redis"] = response.HealthCheck{Status: "unhealthy", Message: err.Error()}
		} else {
			checks["redis"] = response.HealthCheck{Status: "healthy"}
		}

		status := "healthy"
		for _, check := range checks {
			if check.Status != "healthy" {
				status = "unhealthy"
				break
			}
		}

		response.Health(w, status, Version, time.Since(startTime), checks)
	})

	// Metrics endpoint (placeholder)
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("# Metrics placeholder\n"))
	})

	// Lead API routes
	mux.HandleFunc("GET /api/v1/leads", func(w http.ResponseWriter, r *http.Request) {
		response.Paginated(w, []interface{}{}, 1, 10, 0)
	})

	mux.HandleFunc("POST /api/v1/leads", func(w http.ResponseWriter, r *http.Request) {
		response.Created(w, map[string]string{"message": "Create lead - TODO"})
	})

	mux.HandleFunc("GET /api/v1/leads/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Get lead", "id": id})
	})

	mux.HandleFunc("PUT /api/v1/leads/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Update lead", "id": id})
	})

	mux.HandleFunc("POST /api/v1/leads/{id}/convert", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Convert lead to opportunity", "id": id})
	})

	// Opportunity API routes
	mux.HandleFunc("GET /api/v1/opportunities", func(w http.ResponseWriter, r *http.Request) {
		response.Paginated(w, []interface{}{}, 1, 10, 0)
	})

	mux.HandleFunc("POST /api/v1/opportunities", func(w http.ResponseWriter, r *http.Request) {
		response.Created(w, map[string]string{"message": "Create opportunity - TODO"})
	})

	mux.HandleFunc("GET /api/v1/opportunities/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Get opportunity", "id": id})
	})

	mux.HandleFunc("PUT /api/v1/opportunities/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Update opportunity", "id": id})
	})

	mux.HandleFunc("POST /api/v1/opportunities/{id}/move-stage", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Move opportunity stage", "id": id})
	})

	mux.HandleFunc("POST /api/v1/opportunities/{id}/win", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Mark opportunity as won", "id": id})
	})

	mux.HandleFunc("POST /api/v1/opportunities/{id}/lose", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Mark opportunity as lost", "id": id})
	})

	// Pipeline API routes
	mux.HandleFunc("GET /api/v1/pipelines", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, []interface{}{})
	})

	mux.HandleFunc("POST /api/v1/pipelines", func(w http.ResponseWriter, r *http.Request) {
		response.Created(w, map[string]string{"message": "Create pipeline - TODO"})
	})

	mux.HandleFunc("GET /api/v1/pipelines/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Get pipeline", "id": id})
	})

	mux.HandleFunc("GET /api/v1/pipelines/{id}/analytics", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]interface{}{
			"pipeline_id":        id,
			"total_value":        0,
			"weighted_value":     0,
			"opportunities_count": 0,
			"win_rate":           0,
		})
	})

	// Deal API routes
	mux.HandleFunc("GET /api/v1/deals", func(w http.ResponseWriter, r *http.Request) {
		response.Paginated(w, []interface{}{}, 1, 10, 0)
	})

	mux.HandleFunc("GET /api/v1/deals/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Get deal", "id": id})
	})

	// Apply middleware
	handler := middleware.Chain(
		middleware.RequestID,
		middleware.Logger(log),
		middleware.Recover(log),
		middleware.CORS([]string{"*"}, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, []string{"*"}),
		middleware.ContentType("application/json"),
		middleware.Auth(jwtManager),
	)(mux)

	// Create public handler (without auth)
	publicMux := http.NewServeMux()
	publicMux.Handle("/health", mux)
	publicMux.Handle("/metrics", mux)
	publicMux.Handle("/", handler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      publicMux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info().
			Str("addr", server.Addr).
			Msg("HTTP server started")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
}
