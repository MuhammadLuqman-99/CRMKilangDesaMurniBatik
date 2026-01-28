// Package domain contains the domain layer for the Notification service.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NotificationRepository defines the interface for notification persistence.
type NotificationRepository interface {
	// Create creates a new notification.
	Create(ctx context.Context, notification *Notification) error

	// Update updates an existing notification.
	Update(ctx context.Context, notification *Notification) error

	// Delete soft deletes a notification.
	Delete(ctx context.Context, id uuid.UUID) error

	// HardDelete permanently deletes a notification.
	HardDelete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a notification by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Notification, error)

	// FindByCode finds a notification by code.
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Notification, error)

	// List lists notifications with filtering and pagination.
	List(ctx context.Context, filter NotificationFilter) (*NotificationList, error)

	// FindByRecipient finds notifications for a recipient.
	FindByRecipient(ctx context.Context, tenantID uuid.UUID, recipientID uuid.UUID, filter NotificationFilter) (*NotificationList, error)

	// FindByEmail finds notifications sent to an email address.
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string, filter NotificationFilter) (*NotificationList, error)

	// FindByStatus finds notifications by status.
	FindByStatus(ctx context.Context, tenantID uuid.UUID, status NotificationStatus, filter NotificationFilter) (*NotificationList, error)

	// FindPending finds pending notifications ready to send.
	FindPending(ctx context.Context, limit int) ([]*Notification, error)

	// FindScheduled finds scheduled notifications that are due.
	FindScheduled(ctx context.Context, before time.Time, limit int) ([]*Notification, error)

	// FindRetryable finds failed notifications that can be retried.
	FindRetryable(ctx context.Context, before time.Time, limit int) ([]*Notification, error)

	// FindByBatch finds notifications in a batch.
	FindByBatch(ctx context.Context, batchID uuid.UUID) ([]*Notification, error)

	// FindBySourceEvent finds notifications triggered by a source event.
	FindBySourceEvent(ctx context.Context, tenantID uuid.UUID, sourceEvent string, sourceEntityID uuid.UUID) ([]*Notification, error)

	// CountByTenant counts notifications for a tenant.
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)

	// CountByStatus counts notifications by status for a tenant.
	CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[NotificationStatus]int64, error)

	// CountByChannel counts notifications by channel for a tenant.
	CountByChannel(ctx context.Context, tenantID uuid.UUID) (map[NotificationChannel]int64, error)

	// GetStats gets notification statistics.
	GetStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*NotificationStats, error)

	// BulkCreate creates multiple notifications.
	BulkCreate(ctx context.Context, notifications []*Notification) error

	// BulkUpdateStatus updates status for multiple notifications.
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status NotificationStatus) error

	// Exists checks if a notification exists.
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// GetVersion gets the current version for optimistic locking.
	GetVersion(ctx context.Context, id uuid.UUID) (int, error)

	// FindUnread finds unread in-app notifications for a user.
	FindUnread(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*Notification, error)

	// CountUnread counts unread in-app notifications for a user.
	CountUnread(ctx context.Context, tenantID, userID uuid.UUID) (int64, error)

	// MarkAllRead marks all in-app notifications as read for a user.
	MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error

	// DeleteOld deletes notifications older than a specified date.
	DeleteOld(ctx context.Context, before time.Time) (int64, error)
}

// NotificationFilter defines filtering options for notification queries.
type NotificationFilter struct {
	TenantID       *uuid.UUID              `json:"tenant_id,omitempty"`
	IDs            []uuid.UUID             `json:"ids,omitempty"`
	Codes          []string                `json:"codes,omitempty"`
	Channels       []NotificationChannel   `json:"channels,omitempty"`
	Types          []NotificationType      `json:"types,omitempty"`
	Statuses       []NotificationStatus    `json:"statuses,omitempty"`
	Priorities     []NotificationPriority  `json:"priorities,omitempty"`
	RecipientID    *uuid.UUID              `json:"recipient_id,omitempty"`
	RecipientEmail string                  `json:"recipient_email,omitempty"`
	RecipientPhone string                  `json:"recipient_phone,omitempty"`
	TemplateID     *uuid.UUID              `json:"template_id,omitempty"`
	BatchID        *uuid.UUID              `json:"batch_id,omitempty"`
	SourceEvent    string                  `json:"source_event,omitempty"`
	SourceEntityID *uuid.UUID              `json:"source_entity_id,omitempty"`
	CreatedAfter   *time.Time              `json:"created_after,omitempty"`
	CreatedBefore  *time.Time              `json:"created_before,omitempty"`
	SentAfter      *time.Time              `json:"sent_after,omitempty"`
	SentBefore     *time.Time              `json:"sent_before,omitempty"`
	IncludeDeleted bool                    `json:"include_deleted,omitempty"`
	Offset         int                     `json:"offset"`
	Limit          int                     `json:"limit"`
	SortBy         string                  `json:"sort_by,omitempty"`
	SortOrder      string                  `json:"sort_order,omitempty"` // "asc" or "desc"
}

// NotificationList represents a paginated list of notifications.
type NotificationList struct {
	Notifications []*Notification `json:"notifications"`
	Total         int64           `json:"total"`
	Offset        int             `json:"offset"`
	Limit         int             `json:"limit"`
	HasMore       bool            `json:"has_more"`
}

// NotificationStats holds notification statistics.
type NotificationStats struct {
	TotalCount       int64                          `json:"total_count"`
	SentCount        int64                          `json:"sent_count"`
	DeliveredCount   int64                          `json:"delivered_count"`
	FailedCount      int64                          `json:"failed_count"`
	PendingCount     int64                          `json:"pending_count"`
	ByChannel        map[NotificationChannel]int64  `json:"by_channel"`
	ByStatus         map[NotificationStatus]int64   `json:"by_status"`
	ByType           map[NotificationType]int64     `json:"by_type"`
	OpenRate         float64                        `json:"open_rate"`
	ClickRate        float64                        `json:"click_rate"`
	BounceRate       float64                        `json:"bounce_rate"`
	ComplaintRate    float64                        `json:"complaint_rate"`
	AverageDeliveryTime float64                     `json:"average_delivery_time_seconds"`
	Period           string                         `json:"period"`
	StartDate        time.Time                      `json:"start_date"`
	EndDate          time.Time                      `json:"end_date"`
}

// TemplateRepository defines the interface for template persistence.
type TemplateRepository interface {
	// Create creates a new template.
	Create(ctx context.Context, template *NotificationTemplate) error

	// Update updates an existing template.
	Update(ctx context.Context, template *NotificationTemplate) error

	// Delete soft deletes a template.
	Delete(ctx context.Context, id uuid.UUID) error

	// HardDelete permanently deletes a template.
	HardDelete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a template by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*NotificationTemplate, error)

	// FindByCode finds a template by code.
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*NotificationTemplate, error)

	// FindByName finds a template by name.
	FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*NotificationTemplate, error)

	// List lists templates with filtering and pagination.
	List(ctx context.Context, filter TemplateFilter) (*TemplateList, error)

	// FindByType finds templates by notification type.
	FindByType(ctx context.Context, tenantID uuid.UUID, notifType NotificationType) ([]*NotificationTemplate, error)

	// FindByChannel finds templates that support a channel.
	FindByChannel(ctx context.Context, tenantID uuid.UUID, channel NotificationChannel) ([]*NotificationTemplate, error)

	// FindDefault finds the default template for a type.
	FindDefault(ctx context.Context, tenantID uuid.UUID, notifType NotificationType) (*NotificationTemplate, error)

	// FindActive finds all active templates.
	FindActive(ctx context.Context, tenantID uuid.UUID) ([]*NotificationTemplate, error)

	// CountByTenant counts templates for a tenant.
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)

	// Exists checks if a template exists.
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// ExistsByCode checks if a template code exists.
	ExistsByCode(ctx context.Context, tenantID uuid.UUID, code string) (bool, error)

	// GetVersion gets the current version for optimistic locking.
	GetVersion(ctx context.Context, id uuid.UUID) (int, error)

	// IncrementUsageCount increments the usage count for a template.
	IncrementUsageCount(ctx context.Context, id uuid.UUID) error

	// FindByTag finds templates with a specific tag.
	FindByTag(ctx context.Context, tenantID uuid.UUID, tag string) ([]*NotificationTemplate, error)

	// Search searches templates by name or description.
	Search(ctx context.Context, tenantID uuid.UUID, query string, filter TemplateFilter) (*TemplateList, error)
}

// TemplateFilter defines filtering options for template queries.
type TemplateFilter struct {
	TenantID       *uuid.UUID              `json:"tenant_id,omitempty"`
	IDs            []uuid.UUID             `json:"ids,omitempty"`
	Codes          []string                `json:"codes,omitempty"`
	Types          []NotificationType      `json:"types,omitempty"`
	Channels       []NotificationChannel   `json:"channels,omitempty"`
	Categories     []string                `json:"categories,omitempty"`
	Tags           []string                `json:"tags,omitempty"`
	IsActive       *bool                   `json:"is_active,omitempty"`
	IsDefault      *bool                   `json:"is_default,omitempty"`
	Query          string                  `json:"query,omitempty"` // Search query
	CreatedAfter   *time.Time              `json:"created_after,omitempty"`
	CreatedBefore  *time.Time              `json:"created_before,omitempty"`
	IncludeDeleted bool                    `json:"include_deleted,omitempty"`
	Offset         int                     `json:"offset"`
	Limit          int                     `json:"limit"`
	SortBy         string                  `json:"sort_by,omitempty"`
	SortOrder      string                  `json:"sort_order,omitempty"`
}

// TemplateList represents a paginated list of templates.
type TemplateList struct {
	Templates []*NotificationTemplate `json:"templates"`
	Total     int64                   `json:"total"`
	Offset    int                     `json:"offset"`
	Limit     int                     `json:"limit"`
	HasMore   bool                    `json:"has_more"`
}

// PreferenceRepository defines the interface for notification preference persistence.
type PreferenceRepository interface {
	// Save saves user notification preferences.
	Save(ctx context.Context, preference *NotificationPreference) error

	// FindByUser finds preferences for a user.
	FindByUser(ctx context.Context, tenantID, userID uuid.UUID) (*NotificationPreference, error)

	// FindByUserAndChannel finds preferences for a user and channel.
	FindByUserAndChannel(ctx context.Context, tenantID, userID uuid.UUID, channel NotificationChannel) (*ChannelPreference, error)

	// UpdateChannelPreference updates a single channel preference.
	UpdateChannelPreference(ctx context.Context, tenantID, userID uuid.UUID, channel NotificationChannel, enabled bool) error

	// UpdateTypePreference updates a single type preference.
	UpdateTypePreference(ctx context.Context, tenantID, userID uuid.UUID, notifType NotificationType, enabled bool) error

	// Delete deletes user preferences.
	Delete(ctx context.Context, tenantID, userID uuid.UUID) error

	// FindOptedOut finds users who have opted out of a channel or type.
	FindOptedOut(ctx context.Context, tenantID uuid.UUID, channel NotificationChannel, notifType NotificationType) ([]uuid.UUID, error)

	// IsOptedOut checks if a user has opted out.
	IsOptedOut(ctx context.Context, tenantID, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (bool, error)
}

// NotificationPreference represents user notification preferences.
type NotificationPreference struct {
	BaseEntity
	TenantID       uuid.UUID                        `json:"tenant_id" db:"tenant_id"`
	UserID         uuid.UUID                        `json:"user_id" db:"user_id"`
	GlobalOptOut   bool                             `json:"global_opt_out" db:"global_opt_out"`
	ChannelPrefs   map[NotificationChannel]*ChannelPreference `json:"channel_preferences" db:"-"`
	TypePrefs      map[NotificationType]*TypePreference       `json:"type_preferences" db:"-"`
	QuietHours     *QuietHours                      `json:"quiet_hours,omitempty" db:"-"`
	Timezone       string                           `json:"timezone" db:"timezone"`
	Locale         string                           `json:"locale" db:"locale"`
}

// ChannelPreference represents preferences for a specific channel.
type ChannelPreference struct {
	Channel   NotificationChannel `json:"channel"`
	Enabled   bool                `json:"enabled"`
	Address   string              `json:"address,omitempty"`   // Email, phone, etc.
	Verified  bool                `json:"verified"`
}

// TypePreference represents preferences for a specific notification type.
type TypePreference struct {
	Type    NotificationType `json:"type"`
	Enabled bool             `json:"enabled"`
}

// QuietHours represents quiet hours when notifications should not be sent.
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // Format: "HH:MM"
	EndTime   string `json:"end_time"`   // Format: "HH:MM"
	Days      []int  `json:"days"`       // 0=Sunday, 6=Saturday
}

// IsChannelEnabled checks if a channel is enabled for the user.
func (p *NotificationPreference) IsChannelEnabled(channel NotificationChannel) bool {
	if p.GlobalOptOut {
		return false
	}
	if pref, ok := p.ChannelPrefs[channel]; ok {
		return pref.Enabled
	}
	return true // Default to enabled
}

// IsTypeEnabled checks if a notification type is enabled for the user.
func (p *NotificationPreference) IsTypeEnabled(notifType NotificationType) bool {
	if p.GlobalOptOut {
		return false
	}
	if pref, ok := p.TypePrefs[notifType]; ok {
		return pref.Enabled
	}
	return true // Default to enabled
}

// TriggerRepository defines the interface for notification trigger persistence.
type TriggerRepository interface {
	// Create creates a new trigger.
	Create(ctx context.Context, trigger *NotificationTrigger) error

	// Update updates an existing trigger.
	Update(ctx context.Context, trigger *NotificationTrigger) error

	// Delete deletes a trigger.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a trigger by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*NotificationTrigger, error)

	// FindByEventType finds triggers for an event type.
	FindByEventType(ctx context.Context, tenantID uuid.UUID, eventType ExternalEventType) ([]*NotificationTrigger, error)

	// FindActive finds all active triggers for a tenant.
	FindActive(ctx context.Context, tenantID uuid.UUID) ([]*NotificationTrigger, error)

	// FindByTemplate finds triggers using a template.
	FindByTemplate(ctx context.Context, tenantID uuid.UUID, templateCode string) ([]*NotificationTrigger, error)
}

// OutboxRepository defines the interface for the transactional outbox pattern.
type OutboxRepository interface {
	// Create creates an outbox entry.
	Create(ctx context.Context, entry *OutboxEntry) error

	// MarkAsProcessed marks an entry as processed.
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error

	// MarkAsFailed marks an entry as failed.
	MarkAsFailed(ctx context.Context, id uuid.UUID, err string) error

	// FindPending finds pending outbox entries.
	FindPending(ctx context.Context, limit int) ([]*OutboxEntry, error)

	// FindFailed finds failed outbox entries.
	FindFailed(ctx context.Context, limit int) ([]*OutboxEntry, error)

	// DeleteOld deletes old processed entries.
	DeleteOld(ctx context.Context, before time.Time) error
}

// OutboxEntry represents an outbox entry.
type OutboxEntry struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	EventType    string     `json:"event_type" db:"event_type"`
	AggregateID  uuid.UUID  `json:"aggregate_id" db:"aggregate_id"`
	Payload      []byte     `json:"payload" db:"payload"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	FailedAt     *time.Time `json:"failed_at,omitempty" db:"failed_at"`
	Error        string     `json:"error,omitempty" db:"error"`
	RetryCount   int        `json:"retry_count" db:"retry_count"`
}

// EventStoreRepository defines the interface for event store persistence.
type EventStoreRepository interface {
	// Append appends events to the event store.
	Append(ctx context.Context, events []DomainEvent) error

	// GetEvents gets events for an aggregate.
	GetEvents(ctx context.Context, aggregateID uuid.UUID, afterVersion int) ([]DomainEvent, error)

	// GetEventsByType gets events by type.
	GetEventsByType(ctx context.Context, tenantID uuid.UUID, eventType string, limit int) ([]DomainEvent, error)

	// GetEventsSince gets events since a timestamp.
	GetEventsSince(ctx context.Context, tenantID uuid.UUID, since time.Time, limit int) ([]DomainEvent, error)
}

// UnitOfWork defines the interface for transaction management.
type UnitOfWork interface {
	// Begin begins a new transaction.
	Begin(ctx context.Context) (context.Context, error)

	// Commit commits the transaction.
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error

	// Notifications returns the notification repository.
	Notifications() NotificationRepository

	// Templates returns the template repository.
	Templates() TemplateRepository

	// Preferences returns the preference repository.
	Preferences() PreferenceRepository

	// Triggers returns the trigger repository.
	Triggers() TriggerRepository

	// Outbox returns the outbox repository.
	Outbox() OutboxRepository

	// EventStore returns the event store repository.
	EventStore() EventStoreRepository
}

// DeliveryLog represents a delivery attempt log entry.
type DeliveryLog struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	NotificationID uuid.UUID   `json:"notification_id" db:"notification_id"`
	TenantID       uuid.UUID   `json:"tenant_id" db:"tenant_id"`
	Channel        NotificationChannel `json:"channel" db:"channel"`
	Provider       string      `json:"provider" db:"provider"`
	Status         string      `json:"status" db:"status"`
	AttemptNumber  int         `json:"attempt_number" db:"attempt_number"`
	Request        string      `json:"request,omitempty" db:"request"`
	Response       string      `json:"response,omitempty" db:"response"`
	ErrorCode      string      `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage   string      `json:"error_message,omitempty" db:"error_message"`
	LatencyMs      int64       `json:"latency_ms" db:"latency_ms"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
}

// DeliveryLogRepository defines the interface for delivery log persistence.
type DeliveryLogRepository interface {
	// Create creates a delivery log entry.
	Create(ctx context.Context, log *DeliveryLog) error

	// FindByNotification finds logs for a notification.
	FindByNotification(ctx context.Context, notificationID uuid.UUID) ([]*DeliveryLog, error)

	// FindByTenant finds logs for a tenant within a time range.
	FindByTenant(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, limit int) ([]*DeliveryLog, error)

	// DeleteOld deletes logs older than a specified date.
	DeleteOld(ctx context.Context, before time.Time) (int64, error)
}

// WebhookLog represents a webhook delivery log entry.
type WebhookLog struct {
	ID             uuid.UUID `json:"id" db:"id"`
	NotificationID uuid.UUID `json:"notification_id" db:"notification_id"`
	TenantID       uuid.UUID `json:"tenant_id" db:"tenant_id"`
	WebhookURL     string    `json:"webhook_url" db:"webhook_url"`
	Method         string    `json:"method" db:"method"`
	Headers        string    `json:"headers,omitempty" db:"headers"`
	RequestBody    string    `json:"request_body,omitempty" db:"request_body"`
	ResponseStatus int       `json:"response_status" db:"response_status"`
	ResponseBody   string    `json:"response_body,omitempty" db:"response_body"`
	LatencyMs      int64     `json:"latency_ms" db:"latency_ms"`
	Success        bool      `json:"success" db:"success"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Suppression represents an email suppression (bounce, complaint, unsubscribe).
type Suppression struct {
	ID          uuid.UUID           `json:"id" db:"id"`
	TenantID    uuid.UUID           `json:"tenant_id" db:"tenant_id"`
	Email       string              `json:"email" db:"email"`
	Type        SuppressionType     `json:"type" db:"type"`
	Reason      string              `json:"reason,omitempty" db:"reason"`
	Source      string              `json:"source,omitempty" db:"source"` // bounce, complaint, manual
	SourceEvent *uuid.UUID          `json:"source_event,omitempty" db:"source_event"`
	CreatedAt   time.Time           `json:"created_at" db:"created_at"`
	ExpiresAt   *time.Time          `json:"expires_at,omitempty" db:"expires_at"`
}

// SuppressionType represents the type of suppression.
type SuppressionType string

const (
	SuppressionTypeBounce      SuppressionType = "bounce"
	SuppressionTypeComplaint   SuppressionType = "complaint"
	SuppressionTypeUnsubscribe SuppressionType = "unsubscribe"
	SuppressionTypeManual      SuppressionType = "manual"
)

// SuppressionRepository defines the interface for suppression list persistence.
type SuppressionRepository interface {
	// Add adds an email to the suppression list.
	Add(ctx context.Context, suppression *Suppression) error

	// Remove removes an email from the suppression list.
	Remove(ctx context.Context, tenantID uuid.UUID, email string) error

	// IsSuppressed checks if an email is suppressed.
	IsSuppressed(ctx context.Context, tenantID uuid.UUID, email string) (bool, error)

	// FindByEmail finds suppression records for an email.
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]*Suppression, error)

	// FindByType finds suppressions by type.
	FindByType(ctx context.Context, tenantID uuid.UUID, supType SuppressionType, limit int) ([]*Suppression, error)

	// Count counts suppressions by type.
	Count(ctx context.Context, tenantID uuid.UUID) (map[SuppressionType]int64, error)

	// DeleteExpired deletes expired suppressions.
	DeleteExpired(ctx context.Context) (int64, error)
}
