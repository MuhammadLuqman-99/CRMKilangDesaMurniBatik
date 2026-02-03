// Package service provides the status page business logic.
package service

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kilang-desa-murni/crm/internal/shared/status/domain"
)

// HealthChecker performs health checks on registered services.
type HealthChecker struct {
	mu            sync.RWMutex
	services      map[uuid.UUID]*domain.Service
	healthHistory map[uuid.UUID][]domain.HealthCheck
	httpClient    *http.Client
	checkInterval time.Duration
	maxHistory    int
}

// NewHealthChecker creates a new HealthChecker.
func NewHealthChecker(checkInterval time.Duration) *HealthChecker {
	return &HealthChecker{
		services:      make(map[uuid.UUID]*domain.Service),
		healthHistory: make(map[uuid.UUID][]domain.HealthCheck),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		checkInterval: checkInterval,
		maxHistory:    1000, // Keep last 1000 checks per service
	}
}

// RegisterService registers a service for health monitoring.
func (h *HealthChecker) RegisterService(service *domain.Service) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.services[service.ID] = service
	h.healthHistory[service.ID] = make([]domain.HealthCheck, 0)
}

// UnregisterService removes a service from monitoring.
func (h *HealthChecker) UnregisterService(serviceID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.services, serviceID)
	delete(h.healthHistory, serviceID)
}

// GetService returns a service by ID.
func (h *HealthChecker) GetService(serviceID uuid.UUID) (*domain.Service, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	service, ok := h.services[serviceID]
	return service, ok
}

// GetAllServices returns all registered services.
func (h *HealthChecker) GetAllServices() []*domain.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()

	services := make([]*domain.Service, 0, len(h.services))
	for _, service := range h.services {
		services = append(services, service)
	}
	return services
}

// CheckService performs a health check on a single service.
func (h *HealthChecker) CheckService(ctx context.Context, serviceID uuid.UUID) (*domain.HealthCheck, error) {
	h.mu.RLock()
	service, ok := h.services[serviceID]
	h.mu.RUnlock()

	if !ok {
		return nil, ErrServiceNotFound
	}

	check := h.performCheck(ctx, service)

	h.mu.Lock()
	defer h.mu.Unlock()

	// Update service status
	service.Status = check.Status
	service.LastCheck = check.CheckedAt
	service.ResponseMs = check.ResponseMs
	service.UpdatedAt = time.Now()

	// Add to history
	history := h.healthHistory[serviceID]
	history = append(history, *check)
	if len(history) > h.maxHistory {
		history = history[len(history)-h.maxHistory:]
	}
	h.healthHistory[serviceID] = history

	// Recalculate uptime
	service.Uptime = h.calculateUptime(serviceID)

	return check, nil
}

// CheckAllServices performs health checks on all services.
func (h *HealthChecker) CheckAllServices(ctx context.Context) []*domain.HealthCheck {
	h.mu.RLock()
	serviceIDs := make([]uuid.UUID, 0, len(h.services))
	for id := range h.services {
		serviceIDs = append(serviceIDs, id)
	}
	h.mu.RUnlock()

	checks := make([]*domain.HealthCheck, 0, len(serviceIDs))
	var wg sync.WaitGroup

	results := make(chan *domain.HealthCheck, len(serviceIDs))

	for _, id := range serviceIDs {
		wg.Add(1)
		go func(serviceID uuid.UUID) {
			defer wg.Done()
			check, err := h.CheckService(ctx, serviceID)
			if err == nil {
				results <- check
			}
		}(id)
	}

	wg.Wait()
	close(results)

	for check := range results {
		checks = append(checks, check)
	}

	return checks
}

// performCheck executes the actual health check.
func (h *HealthChecker) performCheck(ctx context.Context, service *domain.Service) *domain.HealthCheck {
	check := &domain.HealthCheck{
		ID:        uuid.New(),
		ServiceID: service.ID,
		CheckedAt: time.Now(),
	}

	if service.Endpoint == "" {
		// No endpoint configured, assume operational
		check.Status = domain.StatusOperational
		check.ResponseMs = 0
		check.Message = "No health endpoint configured"
		return check
	}

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, service.Endpoint, nil)
	if err != nil {
		check.Status = domain.StatusMajorOutage
		check.Message = "Failed to create request: " + err.Error()
		return check
	}

	resp, err := h.httpClient.Do(req)
	elapsed := time.Since(start)
	check.ResponseMs = elapsed.Milliseconds()

	if err != nil {
		check.Status = domain.StatusMajorOutage
		check.Message = "Connection failed: " + err.Error()
		return check
	}
	defer resp.Body.Close()

	// Determine status based on response code and latency
	switch {
	case resp.StatusCode >= 500:
		check.Status = domain.StatusMajorOutage
		check.Message = "Server error: " + resp.Status
	case resp.StatusCode >= 400:
		check.Status = domain.StatusPartialOutage
		check.Message = "Client error: " + resp.Status
	case elapsed > 5*time.Second:
		check.Status = domain.StatusDegraded
		check.Message = "High latency detected"
	case elapsed > 2*time.Second:
		check.Status = domain.StatusDegraded
		check.Message = "Elevated latency"
	default:
		check.Status = domain.StatusOperational
		check.Message = "OK"
	}

	return check
}

// calculateUptime calculates the uptime percentage for a service.
func (h *HealthChecker) calculateUptime(serviceID uuid.UUID) float64 {
	history := h.healthHistory[serviceID]
	if len(history) == 0 {
		return 100.0
	}

	operational := 0
	for _, check := range history {
		if check.Status == domain.StatusOperational || check.Status == domain.StatusDegraded {
			operational++
		}
	}

	return float64(operational) / float64(len(history)) * 100.0
}

// GetServiceMetrics returns detailed metrics for a service.
func (h *HealthChecker) GetServiceMetrics(serviceID uuid.UUID) (*domain.ServiceMetrics, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	service, ok := h.services[serviceID]
	if !ok {
		return nil, ErrServiceNotFound
	}

	history := h.healthHistory[serviceID]

	metrics := &domain.ServiceMetrics{
		ServiceID:     serviceID,
		ServiceName:   service.Name,
		CurrentStatus: service.Status,
		TotalChecks:   int64(len(history)),
	}

	if len(history) == 0 {
		metrics.UptimePercent = 100.0
		return metrics, nil
	}

	// Calculate response time stats
	var totalMs int64
	responseTimes := make([]int64, 0, len(history))

	for _, check := range history {
		totalMs += check.ResponseMs
		responseTimes = append(responseTimes, check.ResponseMs)

		if check.Status == domain.StatusOperational || check.Status == domain.StatusDegraded {
			metrics.SuccessfulChecks++
		} else {
			metrics.FailedChecks++
		}
	}

	metrics.AvgResponseMs = totalMs / int64(len(history))
	metrics.UptimePercent = float64(metrics.SuccessfulChecks) / float64(metrics.TotalChecks) * 100.0

	// Sort for percentiles
	sortInt64s(responseTimes)
	metrics.P95ResponseMs = percentile(responseTimes, 95)
	metrics.P99ResponseMs = percentile(responseTimes, 99)

	// Calculate daily uptime
	metrics.DailyUptime = h.calculateDailyUptime(history)

	return metrics, nil
}

// calculateDailyUptime calculates daily uptime from health check history.
func (h *HealthChecker) calculateDailyUptime(history []domain.HealthCheck) []domain.DailyUptime {
	dailyStats := make(map[string]*domain.DailyUptime)

	for _, check := range history {
		dateStr := check.CheckedAt.Format("2006-01-02")
		if _, ok := dailyStats[dateStr]; !ok {
			date, _ := time.Parse("2006-01-02", dateStr)
			dailyStats[dateStr] = &domain.DailyUptime{
				Date: date,
			}
		}

		stats := dailyStats[dateStr]
		if check.Status == domain.StatusOperational || check.Status == domain.StatusDegraded {
			// Assuming checks are every minute, each successful check = 1 minute uptime
		} else {
			stats.DowntimeMinutes++
			stats.IncidentCount++
		}
	}

	// Convert to slice and calculate percentages
	result := make([]domain.DailyUptime, 0, len(dailyStats))
	for _, stats := range dailyStats {
		totalMinutes := 1440 // Minutes in a day
		uptimeMinutes := totalMinutes - stats.DowntimeMinutes
		stats.UptimePercent = float64(uptimeMinutes) / float64(totalMinutes) * 100.0
		result = append(result, *stats)
	}

	return result
}

// GetOverallStatus returns the overall system status.
func (h *HealthChecker) GetOverallStatus() domain.ServiceStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.services) == 0 {
		return domain.StatusOperational
	}

	majorOutages := 0
	partialOutages := 0
	degraded := 0

	for _, service := range h.services {
		switch service.Status {
		case domain.StatusMajorOutage:
			majorOutages++
		case domain.StatusPartialOutage:
			partialOutages++
		case domain.StatusDegraded:
			degraded++
		}
	}

	switch {
	case majorOutages > 0:
		return domain.StatusMajorOutage
	case partialOutages > 0:
		return domain.StatusPartialOutage
	case degraded > 0:
		return domain.StatusDegraded
	default:
		return domain.StatusOperational
	}
}

// StartBackgroundChecks starts periodic health checks.
func (h *HealthChecker) StartBackgroundChecks(ctx context.Context) {
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.CheckAllServices(ctx)
		}
	}
}

// Helper functions

func sortInt64s(nums []int64) {
	for i := 0; i < len(nums)-1; i++ {
		for j := 0; j < len(nums)-i-1; j++ {
			if nums[j] > nums[j+1] {
				nums[j], nums[j+1] = nums[j+1], nums[j]
			}
		}
	}
}

func percentile(sorted []int64, p int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := (len(sorted) - 1) * p / 100
	return sorted[idx]
}

// Errors

var (
	ErrServiceNotFound = &StatusError{Code: "SERVICE_NOT_FOUND", Message: "Service not found"}
)

// StatusError represents a status service error.
type StatusError struct {
	Code    string
	Message string
}

func (e *StatusError) Error() string {
	return e.Message
}
