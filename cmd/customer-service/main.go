// Customer Service - Customer and Contact Management
// ===================================================
// This service handles customer and contact lifecycle management.
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
	cfg.App.Name = "customer-service"
	cfg.Server.Port = 8082

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
		Msg("Starting Customer service")

	// Initialize tracer
	tr, err := tracer.New(&cfg.Tracer, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tracer")
	}
	defer tr.Close(context.Background())

	// Initialize MongoDB
	mongodb, err := database.NewMongoDB(&cfg.MongoDB, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongodb.Close(context.Background())

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

		// Check MongoDB
		if err := mongodb.Health(r.Context()); err != nil {
			checks["mongodb"] = response.HealthCheck{Status: "unhealthy", Message: err.Error()}
		} else {
			checks["mongodb"] = response.HealthCheck{Status: "healthy"}
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

	// Customer API routes
	mux.HandleFunc("GET /api/v1/customers", func(w http.ResponseWriter, r *http.Request) {
		response.Paginated(w, []interface{}{}, 1, 10, 0)
	})

	mux.HandleFunc("POST /api/v1/customers", func(w http.ResponseWriter, r *http.Request) {
		response.Created(w, map[string]string{"message": "Create customer - TODO"})
	})

	mux.HandleFunc("GET /api/v1/customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Get customer", "id": id})
	})

	mux.HandleFunc("PUT /api/v1/customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Update customer", "id": id})
	})

	mux.HandleFunc("DELETE /api/v1/customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.NoContent(w)
	})

	// Contact API routes
	mux.HandleFunc("GET /api/v1/customers/{customerId}/contacts", func(w http.ResponseWriter, r *http.Request) {
		customerId := r.PathValue("customerId")
		response.OK(w, map[string]interface{}{"customer_id": customerId, "contacts": []interface{}{}})
	})

	mux.HandleFunc("POST /api/v1/customers/{customerId}/contacts", func(w http.ResponseWriter, r *http.Request) {
		customerId := r.PathValue("customerId")
		response.Created(w, map[string]string{"message": "Create contact", "customer_id": customerId})
	})

	mux.HandleFunc("GET /api/v1/customers/{customerId}/contacts/{contactId}", func(w http.ResponseWriter, r *http.Request) {
		customerId := r.PathValue("customerId")
		contactId := r.PathValue("contactId")
		response.OK(w, map[string]string{"customer_id": customerId, "contact_id": contactId})
	})

	mux.HandleFunc("PUT /api/v1/customers/{customerId}/contacts/{contactId}", func(w http.ResponseWriter, r *http.Request) {
		customerId := r.PathValue("customerId")
		contactId := r.PathValue("contactId")
		response.OK(w, map[string]string{"message": "Update contact", "customer_id": customerId, "contact_id": contactId})
	})

	mux.HandleFunc("DELETE /api/v1/customers/{customerId}/contacts/{contactId}", func(w http.ResponseWriter, r *http.Request) {
		response.NoContent(w)
	})

	// Search endpoint
	mux.HandleFunc("GET /api/v1/customers/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		response.OK(w, map[string]interface{}{"query": query, "results": []interface{}{}})
	})

	// Import/Export endpoints
	mux.HandleFunc("POST /api/v1/customers/import", func(w http.ResponseWriter, r *http.Request) {
		response.Accepted(w, map[string]string{"message": "Import started - TODO"})
	})

	mux.HandleFunc("GET /api/v1/customers/export", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"message": "Export - TODO"})
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
