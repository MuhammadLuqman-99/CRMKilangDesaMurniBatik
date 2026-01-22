// IAM Service - Identity and Access Management
// =============================================
// This service handles user authentication, authorization, and tenant management.
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
	cfg.App.Name = "iam-service"
	cfg.Server.Port = 8081

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
		Msg("Starting IAM service")

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

	// Initialize JWT Manager
	jwtManager := auth.NewJWTManager(&cfg.JWT)

	// Initialize Password Hasher
	_ = auth.NewPasswordHasher(nil)

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

	// API routes
	mux.HandleFunc("POST /api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"message": "Register endpoint - TODO"})
	})

	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		// Demo login endpoint
		tokenPair, err := jwtManager.GenerateTokenPair(
			"user-123",
			"tenant-456",
			"demo@example.com",
			[]string{"admin"},
		)
		if err != nil {
			response.Error(w, err)
			return
		}
		response.OK(w, tokenPair)
	})

	mux.HandleFunc("POST /api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"message": "Refresh endpoint - TODO"})
	})

	mux.HandleFunc("POST /api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		response.NoContent(w)
	})

	// Protected routes (require authentication)
	mux.HandleFunc("GET /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"message": "List users - TODO"})
	})

	mux.HandleFunc("GET /api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Get user", "id": id})
	})

	mux.HandleFunc("GET /api/v1/roles", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"message": "List roles - TODO"})
	})

	// Apply middleware
	handler := middleware.Chain(
		middleware.RequestID,
		middleware.Logger(log),
		middleware.Recover(log),
		middleware.CORS([]string{"*"}, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, []string{"*"}),
		middleware.ContentType("application/json"),
	)(mux)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
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
