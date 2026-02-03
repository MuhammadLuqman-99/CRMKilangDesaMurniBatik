// Package service provides the status page business logic.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kilang-desa-murni/crm/internal/shared/status/domain"
)

// StatusService manages the status page functionality.
type StatusService struct {
	mu              sync.RWMutex
	healthChecker   *HealthChecker
	incidents       map[uuid.UUID]*domain.Incident
	maintenances    map[uuid.UUID]*domain.ScheduledMaintenance
	subscribers     map[uuid.UUID]*domain.Subscriber
	notifier        Notifier
}

// Notifier interface for sending notifications.
type Notifier interface {
	SendIncidentNotification(ctx context.Context, incident *domain.Incident, subscribers []*domain.Subscriber) error
	SendMaintenanceNotification(ctx context.Context, maintenance *domain.ScheduledMaintenance, subscribers []*domain.Subscriber) error
	SendResolvedNotification(ctx context.Context, incident *domain.Incident, subscribers []*domain.Subscriber) error
}

// NewStatusService creates a new StatusService.
func NewStatusService(healthChecker *HealthChecker, notifier Notifier) *StatusService {
	return &StatusService{
		healthChecker: healthChecker,
		incidents:     make(map[uuid.UUID]*domain.Incident),
		maintenances:  make(map[uuid.UUID]*domain.ScheduledMaintenance),
		subscribers:   make(map[uuid.UUID]*domain.Subscriber),
		notifier:      notifier,
	}
}

// InitializeDefaultServices sets up the default CRM services for monitoring.
func (s *StatusService) InitializeDefaultServices() {
	services := []*domain.Service{
		domain.NewService("API Gateway", "Main API gateway handling all requests", "Core Infrastructure", "/health", 1),
		domain.NewService("IAM Service", "Identity and Access Management", "Core Services", "/api/v1/iam/health", 2),
		domain.NewService("Customer Service", "Customer data management", "Core Services", "/api/v1/customers/health", 3),
		domain.NewService("Sales Service", "Sales pipeline management", "Core Services", "/api/v1/sales/health", 4),
		domain.NewService("Notification Service", "Email and push notifications", "Core Services", "/api/v1/notifications/health", 5),
		domain.NewService("PostgreSQL Primary", "Primary PostgreSQL database", "Databases", "", 6),
		domain.NewService("PostgreSQL Replica", "PostgreSQL read replica", "Databases", "", 7),
		domain.NewService("MongoDB", "Customer data store", "Databases", "", 8),
		domain.NewService("Redis Cache", "Caching and sessions", "Databases", "", 9),
		domain.NewService("RabbitMQ", "Message queue", "Infrastructure", "", 10),
	}

	for _, service := range services {
		s.healthChecker.RegisterService(service)
	}
}

// GetStatusSummary returns the overall status summary.
func (s *StatusService) GetStatusSummary(ctx context.Context) *domain.StatusSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := s.healthChecker.GetAllServices()

	// Convert to non-pointer slice
	serviceList := make([]domain.Service, len(services))
	for i, svc := range services {
		serviceList[i] = *svc
	}

	// Get active incidents
	activeIncidents := make([]domain.Incident, 0)
	for _, incident := range s.incidents {
		if incident.IsActive() {
			activeIncidents = append(activeIncidents, *incident)
		}
	}

	// Get upcoming maintenances
	now := time.Now()
	upcomingMaintenance := make([]domain.ScheduledMaintenance, 0)
	for _, m := range s.maintenances {
		if m.ScheduledEnd.After(now) {
			upcomingMaintenance = append(upcomingMaintenance, *m)
		}
	}

	// Calculate uptime percentages
	uptimeLast24h := s.calculateAverageUptime(services, 24*time.Hour)
	uptimeLast7d := s.calculateAverageUptime(services, 7*24*time.Hour)
	uptimeLast30d := s.calculateAverageUptime(services, 30*24*time.Hour)
	uptimeLast90d := s.calculateAverageUptime(services, 90*24*time.Hour)

	return &domain.StatusSummary{
		OverallStatus:       s.healthChecker.GetOverallStatus(),
		Services:            serviceList,
		ActiveIncidents:     activeIncidents,
		UpcomingMaintenance: upcomingMaintenance,
		UptimeLast24h:       uptimeLast24h,
		UptimeLast7d:        uptimeLast7d,
		UptimeLast30d:       uptimeLast30d,
		UptimeLast90d:       uptimeLast90d,
		LastUpdated:         time.Now(),
	}
}

func (s *StatusService) calculateAverageUptime(services []*domain.Service, duration time.Duration) float64 {
	if len(services) == 0 {
		return 100.0
	}

	var totalUptime float64
	for _, svc := range services {
		totalUptime += svc.Uptime
	}
	return totalUptime / float64(len(services))
}

// CreateIncident creates a new incident.
func (s *StatusService) CreateIncident(ctx context.Context, title string, severity domain.IncidentSeverity, affectedServiceIDs []uuid.UUID, message string) (*domain.Incident, error) {
	incident := domain.NewIncident(title, severity, affectedServiceIDs, message)

	s.mu.Lock()
	s.incidents[incident.ID] = incident

	// Update affected services status
	for _, serviceID := range affectedServiceIDs {
		if service, ok := s.healthChecker.GetService(serviceID); ok {
			switch severity {
			case domain.SeverityCritical:
				service.Status = domain.StatusMajorOutage
			case domain.SeverityMajor:
				service.Status = domain.StatusPartialOutage
			case domain.SeverityMinor:
				service.Status = domain.StatusDegraded
			}
		}
	}
	s.mu.Unlock()

	// Notify subscribers
	if s.notifier != nil {
		subscribers := s.getNotifiableSubscribers(true, false, false)
		go s.notifier.SendIncidentNotification(ctx, incident, subscribers)
	}

	return incident, nil
}

// UpdateIncident updates an existing incident.
func (s *StatusService) UpdateIncident(ctx context.Context, incidentID uuid.UUID, status domain.IncidentStatus, message, updatedBy string) (*domain.Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	incident, ok := s.incidents[incidentID]
	if !ok {
		return nil, ErrIncidentNotFound
	}

	wasActive := incident.IsActive()
	incident.AddUpdate(status, message, updatedBy)

	// If resolved, restore service status
	if status == domain.IncidentResolved {
		for _, serviceID := range incident.AffectedServices {
			if service, ok := s.healthChecker.GetService(serviceID); ok {
				service.Status = domain.StatusOperational
			}
		}

		// Notify resolved
		if s.notifier != nil && wasActive {
			subscribers := s.getNotifiableSubscribers(false, false, true)
			go s.notifier.SendResolvedNotification(ctx, incident, subscribers)
		}
	}

	return incident, nil
}

// GetIncident returns an incident by ID.
func (s *StatusService) GetIncident(incidentID uuid.UUID) (*domain.Incident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incident, ok := s.incidents[incidentID]
	if !ok {
		return nil, ErrIncidentNotFound
	}
	return incident, nil
}

// GetAllIncidents returns all incidents.
func (s *StatusService) GetAllIncidents(activeOnly bool, limit int) []*domain.Incident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	incidents := make([]*domain.Incident, 0)
	for _, incident := range s.incidents {
		if activeOnly && !incident.IsActive() {
			continue
		}
		incidents = append(incidents, incident)
	}

	// Sort by created date descending
	for i := 0; i < len(incidents)-1; i++ {
		for j := 0; j < len(incidents)-i-1; j++ {
			if incidents[j].CreatedAt.Before(incidents[j+1].CreatedAt) {
				incidents[j], incidents[j+1] = incidents[j+1], incidents[j]
			}
		}
	}

	if limit > 0 && len(incidents) > limit {
		incidents = incidents[:limit]
	}

	return incidents
}

// ScheduleMaintenance schedules a maintenance window.
func (s *StatusService) ScheduleMaintenance(ctx context.Context, title, description string, affectedServiceIDs []uuid.UUID, start, end time.Time) (*domain.ScheduledMaintenance, error) {
	maintenance := &domain.ScheduledMaintenance{
		ID:               uuid.New(),
		Title:            title,
		Description:      description,
		AffectedServices: affectedServiceIDs,
		ScheduledStart:   start,
		ScheduledEnd:     end,
		Status:           "scheduled",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	s.mu.Lock()
	s.maintenances[maintenance.ID] = maintenance
	s.mu.Unlock()

	// Notify subscribers
	if s.notifier != nil {
		subscribers := s.getNotifiableSubscribers(false, true, false)
		go s.notifier.SendMaintenanceNotification(ctx, maintenance, subscribers)
	}

	return maintenance, nil
}

// GetScheduledMaintenances returns all scheduled maintenances.
func (s *StatusService) GetScheduledMaintenances(upcomingOnly bool) []*domain.ScheduledMaintenance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	maintenances := make([]*domain.ScheduledMaintenance, 0)

	for _, m := range s.maintenances {
		if upcomingOnly && m.ScheduledEnd.Before(now) {
			continue
		}
		maintenances = append(maintenances, m)
	}

	return maintenances
}

// Subscribe adds a new subscriber.
func (s *StatusService) Subscribe(ctx context.Context, email, phone string, preferences domain.SubscriberPreferences) (*domain.Subscriber, error) {
	subscriber := &domain.Subscriber{
		ID:          uuid.New(),
		Email:       email,
		Phone:       phone,
		Verified:    false,
		VerifyToken: generateToken(),
		Preferences: preferences,
		CreatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.subscribers[subscriber.ID] = subscriber
	s.mu.Unlock()

	// TODO: Send verification email

	return subscriber, nil
}

// VerifySubscriber verifies a subscriber's email/phone.
func (s *StatusService) VerifySubscriber(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sub := range s.subscribers {
		if sub.VerifyToken == token && !sub.Verified {
			sub.Verified = true
			sub.VerifyToken = ""
			return nil
		}
	}

	return ErrInvalidToken
}

// Unsubscribe removes a subscriber.
func (s *StatusService) Unsubscribe(subscriberID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscribers[subscriberID]
	if !ok {
		return ErrSubscriberNotFound
	}

	now := time.Now()
	sub.UnsubscribedAt = &now
	return nil
}

// getNotifiableSubscribers returns subscribers who should be notified.
func (s *StatusService) getNotifiableSubscribers(incidents, maintenance, resolved bool) []*domain.Subscriber {
	subscribers := make([]*domain.Subscriber, 0)

	for _, sub := range s.subscribers {
		if !sub.Verified || sub.UnsubscribedAt != nil {
			continue
		}

		if incidents && sub.Preferences.NotifyIncidents {
			subscribers = append(subscribers, sub)
		} else if maintenance && sub.Preferences.NotifyMaintenance {
			subscribers = append(subscribers, sub)
		} else if resolved && sub.Preferences.NotifyResolved {
			subscribers = append(subscribers, sub)
		}
	}

	return subscribers
}

// GetServiceMetrics returns metrics for a specific service.
func (s *StatusService) GetServiceMetrics(serviceID uuid.UUID) (*domain.ServiceMetrics, error) {
	return s.healthChecker.GetServiceMetrics(serviceID)
}

// Helper function to generate a random token.
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Errors
var (
	ErrIncidentNotFound   = &StatusError{Code: "INCIDENT_NOT_FOUND", Message: "Incident not found"}
	ErrSubscriberNotFound = &StatusError{Code: "SUBSCRIBER_NOT_FOUND", Message: "Subscriber not found"}
	ErrInvalidToken       = &StatusError{Code: "INVALID_TOKEN", Message: "Invalid verification token"}
)
