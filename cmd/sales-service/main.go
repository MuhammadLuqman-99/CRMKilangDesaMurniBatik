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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"

	"github.com/kilang-desa-murni/crm/internal/sales/application/usecase"
	"github.com/kilang-desa-murni/crm/internal/sales/infrastructure/messaging"
	"github.com/kilang-desa-murni/crm/internal/sales/infrastructure/persistence/postgres"
	saleshttp "github.com/kilang-desa-murni/crm/internal/sales/interfaces/http"
	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/database"
	"github.com/kilang-desa-murni/crm/pkg/logger"
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

	// Wrap sql.DB with sqlx for repositories
	sqlxDB := sqlx.NewDb(db.DB, "postgres")

	// Initialize Redis
	redisClient, err := database.NewRedis(&cfg.Redis, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize RabbitMQ Publisher
	rabbitConfig := messaging.RabbitMQConfig{
		URL:               cfg.RabbitMQ.URL,
		Exchange:          messaging.SalesEventsExchange,
		ExchangeType:      cfg.RabbitMQ.ExchangeType,
		Durable:           true,
		AutoDelete:        false,
		DeliveryMode:      2, // Persistent
		ContentType:       "application/json",
		ReconnectDelay:    cfg.RabbitMQ.ReconnectDelay,
		MaxReconnectTries: 10,
		PrefetchCount:     cfg.RabbitMQ.PrefetchCount,
	}

	eventPublisher, err := messaging.NewRabbitMQPublisher(rabbitConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize event publisher")
	}
	defer eventPublisher.Close()

	// Declare queues
	if err := eventPublisher.DeclareQueues(); err != nil {
		log.Warn().Err(err).Msg("Failed to declare queues (non-fatal)")
	}

	// Initialize repositories
	leadRepo := postgres.NewLeadRepository(sqlxDB)
	opportunityRepo := postgres.NewOpportunityRepository(sqlxDB)
	dealRepo := postgres.NewDealRepository(sqlxDB)
	pipelineRepo := postgres.NewPipelineRepository(sqlxDB)

	// Initialize use cases
	leadUseCase := usecase.NewLeadUseCase(
		leadRepo,
		opportunityRepo,
		pipelineRepo,
		eventPublisher,
		nil, // customerService - inject if available
		nil, // userService - inject if available
		nil, // cacheService - inject if available
		nil, // searchService - inject if available
		nil, // idGenerator - inject if available
	)

	opportunityUseCase := usecase.NewOpportunityUseCase(
		opportunityRepo,
		pipelineRepo,
		dealRepo,
		eventPublisher,
		nil, // customerService
		nil, // userService
		nil, // productService
		nil, // cacheService
		nil, // searchService
		nil, // idGenerator
	)

	dealUseCase := usecase.NewDealUseCase(
		dealRepo,
		opportunityRepo,
		eventPublisher,
		nil, // customerService
		nil, // userService
		nil, // productService
		nil, // cacheService
		nil, // searchService
		nil, // idGenerator
		nil, // notificationService
	)

	pipelineUseCase := usecase.NewPipelineUseCase(
		pipelineRepo,
		opportunityRepo,
		eventPublisher,
		nil, // cacheService
		nil, // idGenerator
	)

	// Initialize HTTP handlers
	handler := saleshttp.NewHandler(saleshttp.HandlerDependencies{
		LeadUseCase:        leadUseCase,
		OpportunityUseCase: opportunityUseCase,
		DealUseCase:        dealUseCase,
		PipelineUseCase:    pipelineUseCase,
		MiddlewareConfig: saleshttp.MiddlewareConfig{
			JWTSecret:         cfg.JWT.Secret,
			JWTIssuer:         cfg.JWT.Issuer,
			JWTAudience:       cfg.JWT.Audience,
			SkipAuth:          cfg.App.Environment == "development",
			AllowedOrigins:    []string{"*"},
			RateLimitRequests: 100,
			RateLimitWindow:   time.Minute,
		},
	})

	// Create Chi router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Health check endpoint (no auth)
	startTime := time.Now()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]response.HealthCheck)

		// Check PostgreSQL
		if err := db.Health(r.Context()); err != nil {
			checks["postgresql"] = response.HealthCheck{Status: "unhealthy", Message: err.Error()}
		} else {
			checks["postgresql"] = response.HealthCheck{Status: "healthy"}
		}

		// Check Redis
		if err := redisClient.Health(r.Context()); err != nil {
			checks["redis"] = response.HealthCheck{Status: "unhealthy", Message: err.Error()}
		} else {
			checks["redis"] = response.HealthCheck{Status: "healthy"}
		}

		// Check RabbitMQ
		if !eventPublisher.IsConnected() {
			checks["rabbitmq"] = response.HealthCheck{Status: "unhealthy", Message: "connection closed"}
		} else {
			checks["rabbitmq"] = response.HealthCheck{Status: "healthy"}
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
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("# Metrics placeholder\n"))
	})

	// Register Sales API routes
	handler.RegisterRoutes(r)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	// Close event publisher
	if err := eventPublisher.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close event publisher")
	}

	// Graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
}
