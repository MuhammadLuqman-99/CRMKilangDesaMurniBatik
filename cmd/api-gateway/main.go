// API Gateway - Central Entry Point
// ==================================
// This service acts as the API gateway, routing requests to appropriate services.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kilang-desa-murni/crm/pkg/auth"
	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/database"
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

// ServiceURLs holds the URLs for backend services
type ServiceURLs struct {
	IAM          string
	Customer     string
	Sales        string
	Notification string
}

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override service-specific settings
	cfg.App.Name = "api-gateway"
	cfg.Server.Port = 8080

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
		Msg("Starting API Gateway")

	// Initialize tracer
	tr, err := tracer.New(&cfg.Tracer, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tracer")
	}
	defer tr.Close(context.Background())

	// Initialize Redis for rate limiting
	redis, err := database.NewRedis(&cfg.Redis, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()

	// Initialize JWT Manager for token validation
	jwtManager := auth.NewJWTManager(&cfg.JWT)

	// Get service URLs from environment
	serviceURLs := ServiceURLs{
		IAM:          getEnv("IAM_SERVICE_URL", "http://localhost:8081"),
		Customer:     getEnv("CUSTOMER_SERVICE_URL", "http://localhost:8082"),
		Sales:        getEnv("SALES_SERVICE_URL", "http://localhost:8083"),
		Notification: getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8084"),
	}

	// Create reverse proxies for each service
	iamProxy := createReverseProxy(serviceURLs.IAM, log)
	customerProxy := createReverseProxy(serviceURLs.Customer, log)
	salesProxy := createReverseProxy(serviceURLs.Sales, log)
	notificationProxy := createReverseProxy(serviceURLs.Notification, log)

	// Create rate limiter
	rateLimitConfig := middleware.RateLimitConfig{
		Requests: 100,
		Window:   time.Minute,
		KeyFunc:  middleware.TenantKeyFunc,
	}
	rateLimiter := middleware.NewRedisRateLimiter(redis, rateLimitConfig)

	// Create HTTP router
	mux := http.NewServeMux()

	// Health check endpoint
	startTime := time.Now()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]response.HealthCheck)

		// Check Redis
		if err := redis.Health(r.Context()); err != nil {
			checks["redis"] = response.HealthCheck{Status: "unhealthy", Message: err.Error()}
		} else {
			checks["redis"] = response.HealthCheck{Status: "healthy"}
		}

		// Check backend services
		services := map[string]string{
			"iam":          serviceURLs.IAM + "/health",
			"customer":     serviceURLs.Customer + "/health",
			"sales":        serviceURLs.Sales + "/health",
			"notification": serviceURLs.Notification + "/health",
		}

		client := &http.Client{Timeout: 5 * time.Second}
		for name, url := range services {
			resp, err := client.Get(url)
			if err != nil {
				checks[name] = response.HealthCheck{Status: "unhealthy", Message: err.Error()}
			} else {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					checks[name] = response.HealthCheck{Status: "healthy"}
				} else {
					checks[name] = response.HealthCheck{Status: "unhealthy", Message: fmt.Sprintf("status: %d", resp.StatusCode)}
				}
			}
		}

		status := "healthy"
		for _, check := range checks {
			if check.Status != "healthy" {
				status = "degraded"
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

	// API Documentation endpoint
	mux.HandleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]interface{}{
			"name":    "CRM Kilang Desa Murni Batik API",
			"version": Version,
			"endpoints": map[string]string{
				"auth":         "/api/v1/auth/*",
				"users":        "/api/v1/users/*",
				"customers":    "/api/v1/customers/*",
				"leads":        "/api/v1/leads/*",
				"opportunities": "/api/v1/opportunities/*",
				"pipelines":    "/api/v1/pipelines/*",
				"deals":        "/api/v1/deals/*",
				"notifications": "/api/v1/notifications/*",
			},
		})
	})

	// Route to IAM service (authentication endpoints - no auth required)
	mux.HandleFunc("/api/v1/auth/", func(w http.ResponseWriter, r *http.Request) {
		iamProxy.ServeHTTP(w, r)
	})

	// Route to IAM service (user management - auth required)
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		iamProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/roles/", func(w http.ResponseWriter, r *http.Request) {
		iamProxy.ServeHTTP(w, r)
	})

	// Route to Customer service
	mux.HandleFunc("/api/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		customerProxy.ServeHTTP(w, r)
	})

	// Route to Sales service
	mux.HandleFunc("/api/v1/leads/", func(w http.ResponseWriter, r *http.Request) {
		salesProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/opportunities/", func(w http.ResponseWriter, r *http.Request) {
		salesProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/pipelines/", func(w http.ResponseWriter, r *http.Request) {
		salesProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/api/v1/deals/", func(w http.ResponseWriter, r *http.Request) {
		salesProxy.ServeHTTP(w, r)
	})

	// Route to Notification service
	mux.HandleFunc("/api/v1/notifications/", func(w http.ResponseWriter, r *http.Request) {
		notificationProxy.ServeHTTP(w, r)
	})

	// Apply middleware for public endpoints (health, metrics, api docs)
	publicHandler := middleware.Chain(
		middleware.RequestID,
		middleware.Logger(log),
		middleware.Recover(log),
		middleware.CORS([]string{"*"}, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, []string{"*"}),
	)(mux)

	// Apply middleware for protected endpoints
	protectedHandler := middleware.Chain(
		middleware.RequestID,
		middleware.Logger(log),
		middleware.Recover(log),
		middleware.CORS([]string{"*"}, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, []string{"*"}),
		middleware.RateLimit(rateLimiter, rateLimitConfig),
		middleware.Auth(jwtManager),
	)(mux)

	// Create main handler that selects appropriate handler based on path
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" || r.URL.Path == "/api" ||
			r.URL.Path == "/api/v1/auth/login" || r.URL.Path == "/api/v1/auth/register" ||
			r.URL.Path == "/api/v1/auth/refresh" {
			publicHandler.ServeHTTP(w, r)
			return
		}

		// Protected endpoints
		protectedHandler.ServeHTTP(w, r)
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mainHandler,
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

// createReverseProxy creates a reverse proxy for a target URL.
func createReverseProxy(target string, log *logger.Logger) *httputil.ReverseProxy {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatal().Err(err).Str("target", target).Msg("Invalid target URL")
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().
			Err(err).
			Str("target", target).
			Str("path", r.URL.Path).
			Msg("Proxy error")

		response.Error(w, fmt.Errorf("service unavailable: %w", err))
	}

	// Modify request before sending to backend
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)

		// Forward original host and protocol
		r.Header.Set("X-Forwarded-Host", r.Host)
		r.Header.Set("X-Forwarded-Proto", "http")

		// Forward request ID if present
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			r.Header.Set("X-Request-ID", requestID)
		}
	}

	// Modify response from backend
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Add gateway headers
		resp.Header.Set("X-Gateway", "crm-api-gateway")
		return nil
	}

	return proxy
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// aggregateResponses aggregates responses from multiple services (for future use).
func aggregateResponses(ctx context.Context, urls []string) ([]interface{}, error) {
	results := make([]interface{}, len(urls))
	client := &http.Client{Timeout: 10 * time.Second}

	for i, url := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		results[i] = string(body)
	}

	return results, nil
}
