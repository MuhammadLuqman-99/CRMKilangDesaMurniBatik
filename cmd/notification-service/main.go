// Notification Service - Email and SMS Notifications
// ===================================================
// This service handles email, SMS, and in-app notifications.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	cfg.App.Name = "notification-service"
	cfg.Server.Port = 8084

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
		Msg("Starting Notification service")

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

	// Subscribe to events
	go func() {
		eventTypes := []events.EventType{
			events.EventTypeUserCreated,
			events.EventTypeLeadCreated,
			events.EventTypeOpportunityWon,
			events.EventTypeOpportunityLost,
			events.EventTypeEmailSend,
			events.EventTypeSMSSend,
		}

		err := eventBus.Subscribe(context.Background(), eventTypes, func(ctx context.Context, event *events.Event) error {
			log.Info().
				Str("event_id", event.ID).
				Str("event_type", string(event.Type)).
				Str("tenant_id", event.TenantID).
				Msg("Received event")

			// Handle different event types
			switch event.Type {
			case events.EventTypeUserCreated:
				// Send welcome email
				log.Info().Str("user_id", event.AggregateID).Msg("Sending welcome email")
			case events.EventTypeLeadCreated:
				// Notify sales team
				log.Info().Str("lead_id", event.AggregateID).Msg("Notifying sales team of new lead")
			case events.EventTypeOpportunityWon:
				// Send confirmation to customer
				log.Info().Str("opportunity_id", event.AggregateID).Msg("Sending deal confirmation")
			case events.EventTypeOpportunityLost:
				// Send follow-up survey
				log.Info().Str("opportunity_id", event.AggregateID).Msg("Sending follow-up survey")
			case events.EventTypeEmailSend:
				// Send email
				log.Info().Interface("data", event.Data).Msg("Sending email")
			case events.EventTypeSMSSend:
				// Send SMS
				log.Info().Interface("data", event.Data).Msg("Sending SMS")
			}

			return nil
		})

		if err != nil {
			log.Error().Err(err).Msg("Failed to subscribe to events")
		}
	}()

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

	// Notification API routes
	mux.HandleFunc("GET /api/v1/notifications", func(w http.ResponseWriter, r *http.Request) {
		response.Paginated(w, []interface{}{}, 1, 10, 0)
	})

	mux.HandleFunc("GET /api/v1/notifications/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Get notification", "id": id})
	})

	// Template API routes
	mux.HandleFunc("GET /api/v1/notifications/templates", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, []interface{}{})
	})

	mux.HandleFunc("POST /api/v1/notifications/templates", func(w http.ResponseWriter, r *http.Request) {
		response.Created(w, map[string]string{"message": "Create template - TODO"})
	})

	mux.HandleFunc("GET /api/v1/notifications/templates/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Get template", "id": id})
	})

	mux.HandleFunc("PUT /api/v1/notifications/templates/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		response.OK(w, map[string]string{"message": "Update template", "id": id})
	})

	mux.HandleFunc("DELETE /api/v1/notifications/templates/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.NoContent(w)
	})

	// Direct send endpoints (for internal use)
	mux.HandleFunc("POST /api/v1/notifications/send/email", func(w http.ResponseWriter, r *http.Request) {
		response.Accepted(w, map[string]string{"message": "Email queued for sending"})
	})

	mux.HandleFunc("POST /api/v1/notifications/send/sms", func(w http.ResponseWriter, r *http.Request) {
		response.Accepted(w, map[string]string{"message": "SMS queued for sending"})
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
