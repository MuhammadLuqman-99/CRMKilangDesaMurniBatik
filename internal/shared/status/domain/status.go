// Package domain provides domain models for the status page service.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ServiceStatus represents the operational status of a service.
type ServiceStatus string

const (
	StatusOperational      ServiceStatus = "operational"
	StatusDegraded         ServiceStatus = "degraded"
	StatusPartialOutage    ServiceStatus = "partial_outage"
	StatusMajorOutage      ServiceStatus = "major_outage"
	StatusUnderMaintenance ServiceStatus = "maintenance"
)

// IncidentSeverity represents the severity level of an incident.
type IncidentSeverity string

const (
	SeverityCritical IncidentSeverity = "critical"
	SeverityMajor    IncidentSeverity = "major"
	SeverityMinor    IncidentSeverity = "minor"
)

// IncidentStatus represents the status of an incident.
type IncidentStatus string

const (
	IncidentInvestigating IncidentStatus = "investigating"
	IncidentIdentified    IncidentStatus = "identified"
	IncidentMonitoring    IncidentStatus = "monitoring"
	IncidentResolved      IncidentStatus = "resolved"
)

// Service represents a monitored service.
type Service struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      ServiceStatus `json:"status"`
	Group       string        `json:"group"`
	Order       int           `json:"order"`
	Endpoint    string        `json:"-"` // Internal endpoint for health checks
	LastCheck   time.Time     `json:"last_check"`
	Uptime      float64       `json:"uptime"` // Percentage
	ResponseMs  int64         `json:"response_ms"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// HealthCheck represents a single health check result.
type HealthCheck struct {
	ID         uuid.UUID     `json:"id"`
	ServiceID  uuid.UUID     `json:"service_id"`
	Status     ServiceStatus `json:"status"`
	ResponseMs int64         `json:"response_ms"`
	Message    string        `json:"message,omitempty"`
	CheckedAt  time.Time     `json:"checked_at"`
}

// Incident represents a service incident.
type Incident struct {
	ID               uuid.UUID        `json:"id"`
	Title            string           `json:"title"`
	Status           IncidentStatus   `json:"status"`
	Severity         IncidentSeverity `json:"severity"`
	AffectedServices []uuid.UUID      `json:"affected_services"`
	Updates          []IncidentUpdate `json:"updates"`
	StartedAt        time.Time        `json:"started_at"`
	ResolvedAt       *time.Time       `json:"resolved_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// IncidentUpdate represents an update to an incident.
type IncidentUpdate struct {
	ID        uuid.UUID      `json:"id"`
	Status    IncidentStatus `json:"status"`
	Message   string         `json:"message"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy string         `json:"created_by"`
}

// ScheduledMaintenance represents a planned maintenance window.
type ScheduledMaintenance struct {
	ID               uuid.UUID   `json:"id"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	AffectedServices []uuid.UUID `json:"affected_services"`
	ScheduledStart   time.Time   `json:"scheduled_start"`
	ScheduledEnd     time.Time   `json:"scheduled_end"`
	Status           string      `json:"status"` // scheduled, in_progress, completed
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// Subscriber represents an email/SMS subscriber for status updates.
type Subscriber struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	Verified     bool      `json:"verified"`
	VerifyToken  string    `json:"-"`
	Preferences  SubscriberPreferences `json:"preferences"`
	CreatedAt    time.Time `json:"created_at"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at,omitempty"`
}

// SubscriberPreferences represents notification preferences.
type SubscriberPreferences struct {
	NotifyIncidents    bool `json:"notify_incidents"`
	NotifyMaintenance  bool `json:"notify_maintenance"`
	NotifyResolved     bool `json:"notify_resolved"`
	EmailNotifications bool `json:"email_notifications"`
	SMSNotifications   bool `json:"sms_notifications"`
}

// StatusSummary represents the overall system status.
type StatusSummary struct {
	OverallStatus    ServiceStatus           `json:"overall_status"`
	Services         []Service               `json:"services"`
	ActiveIncidents  []Incident              `json:"active_incidents"`
	UpcomingMaintenance []ScheduledMaintenance `json:"upcoming_maintenance"`
	UptimeLast24h    float64                 `json:"uptime_last_24h"`
	UptimeLast7d     float64                 `json:"uptime_last_7d"`
	UptimeLast30d    float64                 `json:"uptime_last_30d"`
	UptimeLast90d    float64                 `json:"uptime_last_90d"`
	LastUpdated      time.Time               `json:"last_updated"`
}

// DailyUptime represents daily uptime statistics.
type DailyUptime struct {
	Date           time.Time `json:"date"`
	UptimePercent  float64   `json:"uptime_percent"`
	IncidentCount  int       `json:"incident_count"`
	DowntimeMinutes int      `json:"downtime_minutes"`
}

// ServiceMetrics represents detailed metrics for a service.
type ServiceMetrics struct {
	ServiceID       uuid.UUID     `json:"service_id"`
	ServiceName     string        `json:"service_name"`
	CurrentStatus   ServiceStatus `json:"current_status"`
	AvgResponseMs   int64         `json:"avg_response_ms"`
	P95ResponseMs   int64         `json:"p95_response_ms"`
	P99ResponseMs   int64         `json:"p99_response_ms"`
	TotalChecks     int64         `json:"total_checks"`
	SuccessfulChecks int64        `json:"successful_checks"`
	FailedChecks    int64         `json:"failed_checks"`
	UptimePercent   float64       `json:"uptime_percent"`
	DailyUptime     []DailyUptime `json:"daily_uptime"`
}

// NewService creates a new Service.
func NewService(name, description, group, endpoint string, order int) *Service {
	now := time.Now()
	return &Service{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Status:      StatusOperational,
		Group:       group,
		Order:       order,
		Endpoint:    endpoint,
		LastCheck:   now,
		Uptime:      100.0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewIncident creates a new Incident.
func NewIncident(title string, severity IncidentSeverity, affectedServices []uuid.UUID, initialMessage string) *Incident {
	now := time.Now()
	return &Incident{
		ID:               uuid.New(),
		Title:            title,
		Status:           IncidentInvestigating,
		Severity:         severity,
		AffectedServices: affectedServices,
		Updates: []IncidentUpdate{
			{
				ID:        uuid.New(),
				Status:    IncidentInvestigating,
				Message:   initialMessage,
				CreatedAt: now,
				CreatedBy: "System",
			},
		},
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddUpdate adds an update to the incident.
func (i *Incident) AddUpdate(status IncidentStatus, message, createdBy string) {
	now := time.Now()
	i.Updates = append(i.Updates, IncidentUpdate{
		ID:        uuid.New(),
		Status:    status,
		Message:   message,
		CreatedAt: now,
		CreatedBy: createdBy,
	})
	i.Status = status
	i.UpdatedAt = now

	if status == IncidentResolved {
		i.ResolvedAt = &now
	}
}

// IsActive returns true if the incident is not resolved.
func (i *Incident) IsActive() bool {
	return i.Status != IncidentResolved
}

// Duration returns the duration of the incident.
func (i *Incident) Duration() time.Duration {
	if i.ResolvedAt != nil {
		return i.ResolvedAt.Sub(i.StartedAt)
	}
	return time.Since(i.StartedAt)
}
