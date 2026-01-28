// Package domain contains the domain layer for the Notification service.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Aggregate type constants
const (
	AggregateTypeNotification = "notification"
	AggregateTypeTemplate     = "template"
	AggregateTypePreference   = "preference"
)

// ============================================================================
// Notification Events
// ============================================================================

// NotificationCreatedEvent is raised when a notification is created.
type NotificationCreatedEvent struct {
	BaseDomainEvent
	NotificationCode string              `json:"notification_code"`
	Channel          NotificationChannel `json:"channel"`
	Type             NotificationType    `json:"type"`
	Priority         NotificationPriority `json:"priority"`
	RecipientID      *uuid.UUID          `json:"recipient_id,omitempty"`
	RecipientEmail   string              `json:"recipient_email,omitempty"`
	RecipientPhone   string              `json:"recipient_phone,omitempty"`
	TemplateID       *uuid.UUID          `json:"template_id,omitempty"`
	SourceEvent      string              `json:"source_event,omitempty"`
	CreatedBy        *uuid.UUID          `json:"created_by,omitempty"`
}

// NewNotificationCreatedEvent creates a new notification created event.
func NewNotificationCreatedEvent(n *Notification) *NotificationCreatedEvent {
	return &NotificationCreatedEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.created", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		Channel:          n.Channel,
		Type:             n.Type,
		Priority:         n.Priority,
		RecipientID:      n.RecipientID,
		RecipientEmail:   n.RecipientEmail,
		RecipientPhone:   n.RecipientPhone,
		TemplateID:       n.TemplateID,
		SourceEvent:      n.SourceEvent,
		CreatedBy:        n.CreatedBy,
	}
}

// NotificationQueuedEvent is raised when a notification is queued for delivery.
type NotificationQueuedEvent struct {
	BaseDomainEvent
	NotificationCode string              `json:"notification_code"`
	Channel          NotificationChannel `json:"channel"`
	Priority         NotificationPriority `json:"priority"`
}

// NewNotificationQueuedEvent creates a new notification queued event.
func NewNotificationQueuedEvent(n *Notification) *NotificationQueuedEvent {
	return &NotificationQueuedEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.queued", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		Channel:          n.Channel,
		Priority:         n.Priority,
	}
}

// NotificationScheduledEvent is raised when a notification is scheduled.
type NotificationScheduledEvent struct {
	BaseDomainEvent
	NotificationCode string    `json:"notification_code"`
	ScheduledAt      time.Time `json:"scheduled_at"`
}

// NewNotificationScheduledEvent creates a new notification scheduled event.
func NewNotificationScheduledEvent(n *Notification) *NotificationScheduledEvent {
	return &NotificationScheduledEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.scheduled", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		ScheduledAt:      *n.ScheduledAt,
	}
}

// NotificationSentEvent is raised when a notification is sent.
type NotificationSentEvent struct {
	BaseDomainEvent
	NotificationCode  string              `json:"notification_code"`
	Channel           NotificationChannel `json:"channel"`
	RecipientID       *uuid.UUID          `json:"recipient_id,omitempty"`
	RecipientEmail    string              `json:"recipient_email,omitempty"`
	RecipientPhone    string              `json:"recipient_phone,omitempty"`
	Provider          string              `json:"provider,omitempty"`
	ProviderMessageID string              `json:"provider_message_id,omitempty"`
	SentAt            time.Time           `json:"sent_at"`
	AttemptCount      int                 `json:"attempt_count"`
}

// NewNotificationSentEvent creates a new notification sent event.
func NewNotificationSentEvent(n *Notification) *NotificationSentEvent {
	return &NotificationSentEvent{
		BaseDomainEvent:   NewBaseDomainEvent("notification.sent", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode:  n.Code,
		Channel:           n.Channel,
		RecipientID:       n.RecipientID,
		RecipientEmail:    n.RecipientEmail,
		RecipientPhone:    n.RecipientPhone,
		Provider:          n.Provider,
		ProviderMessageID: n.ProviderMessageID,
		SentAt:            *n.SentAt,
		AttemptCount:      n.AttemptCount,
	}
}

// NotificationDeliveredEvent is raised when a notification is delivered.
type NotificationDeliveredEvent struct {
	BaseDomainEvent
	NotificationCode string              `json:"notification_code"`
	Channel          NotificationChannel `json:"channel"`
	RecipientID      *uuid.UUID          `json:"recipient_id,omitempty"`
	DeliveredAt      time.Time           `json:"delivered_at"`
}

// NewNotificationDeliveredEvent creates a new notification delivered event.
func NewNotificationDeliveredEvent(n *Notification) *NotificationDeliveredEvent {
	return &NotificationDeliveredEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.delivered", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		Channel:          n.Channel,
		RecipientID:      n.RecipientID,
		DeliveredAt:      *n.DeliveredAt,
	}
}

// NotificationReadEvent is raised when an in-app notification is read.
type NotificationReadEvent struct {
	BaseDomainEvent
	NotificationCode string     `json:"notification_code"`
	RecipientID      *uuid.UUID `json:"recipient_id,omitempty"`
	ReadAt           time.Time  `json:"read_at"`
}

// NewNotificationReadEvent creates a new notification read event.
func NewNotificationReadEvent(n *Notification) *NotificationReadEvent {
	return &NotificationReadEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.read", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		RecipientID:      n.RecipientID,
		ReadAt:           *n.ReadAt,
	}
}

// NotificationFailedEvent is raised when a notification delivery fails.
type NotificationFailedEvent struct {
	BaseDomainEvent
	NotificationCode string              `json:"notification_code"`
	Channel          NotificationChannel `json:"channel"`
	ErrorCode        string              `json:"error_code"`
	ErrorMessage     string              `json:"error_message"`
	ProviderError    string              `json:"provider_error,omitempty"`
	AttemptCount     int                 `json:"attempt_count"`
	FailedAt         time.Time           `json:"failed_at"`
	Retryable        bool                `json:"retryable"`
}

// NewNotificationFailedEvent creates a new notification failed event.
func NewNotificationFailedEvent(n *Notification) *NotificationFailedEvent {
	return &NotificationFailedEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.failed", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		Channel:          n.Channel,
		ErrorCode:        n.ErrorCode,
		ErrorMessage:     n.ErrorMessage,
		ProviderError:    n.ProviderError,
		AttemptCount:     n.AttemptCount,
		FailedAt:         *n.FailedAt,
		Retryable:        n.CanRetry(),
	}
}

// NotificationRetryScheduledEvent is raised when a retry is scheduled.
type NotificationRetryScheduledEvent struct {
	BaseDomainEvent
	NotificationCode string    `json:"notification_code"`
	AttemptCount     int       `json:"attempt_count"`
	NextRetryAt      time.Time `json:"next_retry_at"`
}

// NewNotificationRetryScheduledEvent creates a new retry scheduled event.
func NewNotificationRetryScheduledEvent(n *Notification) *NotificationRetryScheduledEvent {
	return &NotificationRetryScheduledEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.retry_scheduled", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		AttemptCount:     n.AttemptCount,
		NextRetryAt:      *n.NextRetryAt,
	}
}

// NotificationCancelledEvent is raised when a notification is cancelled.
type NotificationCancelledEvent struct {
	BaseDomainEvent
	NotificationCode string    `json:"notification_code"`
	CancelledAt      time.Time `json:"cancelled_at"`
}

// NewNotificationCancelledEvent creates a new notification cancelled event.
func NewNotificationCancelledEvent(n *Notification) *NotificationCancelledEvent {
	return &NotificationCancelledEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.cancelled", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		CancelledAt:      *n.CancelledAt,
	}
}

// NotificationBouncedEvent is raised when an email bounces.
type NotificationBouncedEvent struct {
	BaseDomainEvent
	NotificationCode string `json:"notification_code"`
	RecipientEmail   string `json:"recipient_email"`
	BounceType       string `json:"bounce_type"`
}

// NewNotificationBouncedEvent creates a new notification bounced event.
func NewNotificationBouncedEvent(n *Notification, bounceType string) *NotificationBouncedEvent {
	return &NotificationBouncedEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.bounced", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		RecipientEmail:   n.RecipientEmail,
		BounceType:       bounceType,
	}
}

// NotificationComplainedEvent is raised when a spam complaint is received.
type NotificationComplainedEvent struct {
	BaseDomainEvent
	NotificationCode string `json:"notification_code"`
	RecipientEmail   string `json:"recipient_email"`
}

// NewNotificationComplainedEvent creates a new notification complained event.
func NewNotificationComplainedEvent(n *Notification) *NotificationComplainedEvent {
	return &NotificationComplainedEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.complained", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		RecipientEmail:   n.RecipientEmail,
	}
}

// NotificationOpenedEvent is raised when an email is opened.
type NotificationOpenedEvent struct {
	BaseDomainEvent
	NotificationCode string    `json:"notification_code"`
	RecipientEmail   string    `json:"recipient_email"`
	OpenedAt         time.Time `json:"opened_at"`
	OpenCount        int       `json:"open_count"`
}

// NewNotificationOpenedEvent creates a new notification opened event.
func NewNotificationOpenedEvent(n *Notification) *NotificationOpenedEvent {
	return &NotificationOpenedEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.opened", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		RecipientEmail:   n.RecipientEmail,
		OpenedAt:         time.Now().UTC(),
		OpenCount:        n.OpenCount,
	}
}

// NotificationClickedEvent is raised when a link in an email is clicked.
type NotificationClickedEvent struct {
	BaseDomainEvent
	NotificationCode string    `json:"notification_code"`
	RecipientEmail   string    `json:"recipient_email"`
	ClickedAt        time.Time `json:"clicked_at"`
	ClickCount       int       `json:"click_count"`
	URL              string    `json:"url,omitempty"`
}

// NewNotificationClickedEvent creates a new notification clicked event.
func NewNotificationClickedEvent(n *Notification, url string) *NotificationClickedEvent {
	return &NotificationClickedEvent{
		BaseDomainEvent:  NewBaseDomainEvent("notification.clicked", AggregateTypeNotification, n.ID, n.TenantID, n.Version),
		NotificationCode: n.Code,
		RecipientEmail:   n.RecipientEmail,
		ClickedAt:        time.Now().UTC(),
		ClickCount:       n.ClickCount,
		URL:              url,
	}
}

// ============================================================================
// Template Events
// ============================================================================

// TemplateCreatedEvent is raised when a template is created.
type TemplateCreatedEvent struct {
	BaseDomainEvent
	TemplateCode string                `json:"template_code"`
	Name         string                `json:"name"`
	Type         NotificationType      `json:"type"`
	Channels     []NotificationChannel `json:"channels"`
	CreatedBy    *uuid.UUID            `json:"created_by,omitempty"`
}

// NewTemplateCreatedEvent creates a new template created event.
func NewTemplateCreatedEvent(t *NotificationTemplate) *TemplateCreatedEvent {
	return &TemplateCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("template.created", AggregateTypeTemplate, t.ID, t.TenantID, t.Version),
		TemplateCode:    t.Code,
		Name:            t.Name,
		Type:            t.Type,
		Channels:        t.Channels,
		CreatedBy:       t.CreatedBy,
	}
}

// TemplateUpdatedEvent is raised when a template is updated.
type TemplateUpdatedEvent struct {
	BaseDomainEvent
	TemplateCode    string     `json:"template_code"`
	Name            string     `json:"name"`
	TemplateVersion int        `json:"template_version"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty"`
}

// NewTemplateUpdatedEvent creates a new template updated event.
func NewTemplateUpdatedEvent(t *NotificationTemplate) *TemplateUpdatedEvent {
	return &TemplateUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("template.updated", AggregateTypeTemplate, t.ID, t.TenantID, t.Version),
		TemplateCode:    t.Code,
		Name:            t.Name,
		TemplateVersion: t.TemplateVersion,
		UpdatedBy:       t.UpdatedBy,
	}
}

// TemplatePublishedEvent is raised when a template is published.
type TemplatePublishedEvent struct {
	BaseDomainEvent
	TemplateCode    string    `json:"template_code"`
	Name            string    `json:"name"`
	TemplateVersion int       `json:"template_version"`
	PublishedAt     time.Time `json:"published_at"`
}

// NewTemplatePublishedEvent creates a new template published event.
func NewTemplatePublishedEvent(t *NotificationTemplate) *TemplatePublishedEvent {
	return &TemplatePublishedEvent{
		BaseDomainEvent: NewBaseDomainEvent("template.published", AggregateTypeTemplate, t.ID, t.TenantID, t.Version),
		TemplateCode:    t.Code,
		Name:            t.Name,
		TemplateVersion: t.TemplateVersion,
		PublishedAt:     *t.PublishedAt,
	}
}

// TemplateActivatedEvent is raised when a template is activated.
type TemplateActivatedEvent struct {
	BaseDomainEvent
	TemplateCode string `json:"template_code"`
	Name         string `json:"name"`
}

// NewTemplateActivatedEvent creates a new template activated event.
func NewTemplateActivatedEvent(t *NotificationTemplate) *TemplateActivatedEvent {
	return &TemplateActivatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("template.activated", AggregateTypeTemplate, t.ID, t.TenantID, t.Version),
		TemplateCode:    t.Code,
		Name:            t.Name,
	}
}

// TemplateDeactivatedEvent is raised when a template is deactivated.
type TemplateDeactivatedEvent struct {
	BaseDomainEvent
	TemplateCode string `json:"template_code"`
	Name         string `json:"name"`
}

// NewTemplateDeactivatedEvent creates a new template deactivated event.
func NewTemplateDeactivatedEvent(t *NotificationTemplate) *TemplateDeactivatedEvent {
	return &TemplateDeactivatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("template.deactivated", AggregateTypeTemplate, t.ID, t.TenantID, t.Version),
		TemplateCode:    t.Code,
		Name:            t.Name,
	}
}

// TemplateDeletedEvent is raised when a template is deleted.
type TemplateDeletedEvent struct {
	BaseDomainEvent
	TemplateCode string `json:"template_code"`
	Name         string `json:"name"`
}

// NewTemplateDeletedEvent creates a new template deleted event.
func NewTemplateDeletedEvent(t *NotificationTemplate) *TemplateDeletedEvent {
	return &TemplateDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent("template.deleted", AggregateTypeTemplate, t.ID, t.TenantID, t.Version),
		TemplateCode:    t.Code,
		Name:            t.Name,
	}
}

// ============================================================================
// Preference Events
// ============================================================================

// PreferenceUpdatedEvent is raised when user preferences are updated.
type PreferenceUpdatedEvent struct {
	BaseDomainEvent
	UserID   uuid.UUID           `json:"user_id"`
	Channel  NotificationChannel `json:"channel"`
	Enabled  bool                `json:"enabled"`
}

// NewPreferenceUpdatedEvent creates a new preference updated event.
func NewPreferenceUpdatedEvent(tenantID, userID uuid.UUID, channel NotificationChannel, enabled bool) *PreferenceUpdatedEvent {
	return &PreferenceUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("preference.updated", AggregateTypePreference, userID, tenantID, 1),
		UserID:          userID,
		Channel:         channel,
		Enabled:         enabled,
	}
}

// PreferenceOptOutEvent is raised when a user opts out of notifications.
type PreferenceOptOutEvent struct {
	BaseDomainEvent
	UserID    uuid.UUID            `json:"user_id"`
	Channel   NotificationChannel  `json:"channel,omitempty"`
	Type      NotificationType     `json:"type,omitempty"`
	OptOutAll bool                 `json:"opt_out_all"`
}

// NewPreferenceOptOutEvent creates a new preference opt-out event.
func NewPreferenceOptOutEvent(tenantID, userID uuid.UUID, channel NotificationChannel, notifType NotificationType, optOutAll bool) *PreferenceOptOutEvent {
	return &PreferenceOptOutEvent{
		BaseDomainEvent: NewBaseDomainEvent("preference.opt_out", AggregateTypePreference, userID, tenantID, 1),
		UserID:          userID,
		Channel:         channel,
		Type:            notifType,
		OptOutAll:       optOutAll,
	}
}

// ============================================================================
// Batch Events
// ============================================================================

// BatchCreatedEvent is raised when a notification batch is created.
type BatchCreatedEvent struct {
	BaseDomainEvent
	BatchID          uuid.UUID           `json:"batch_id"`
	Channel          NotificationChannel `json:"channel"`
	TotalCount       int                 `json:"total_count"`
	TemplateID       *uuid.UUID          `json:"template_id,omitempty"`
}

// NewBatchCreatedEvent creates a new batch created event.
func NewBatchCreatedEvent(tenantID, batchID uuid.UUID, channel NotificationChannel, totalCount int, templateID *uuid.UUID) *BatchCreatedEvent {
	return &BatchCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("batch.created", "batch", batchID, tenantID, 1),
		BatchID:         batchID,
		Channel:         channel,
		TotalCount:      totalCount,
		TemplateID:      templateID,
	}
}

// BatchCompletedEvent is raised when a notification batch is completed.
type BatchCompletedEvent struct {
	BaseDomainEvent
	BatchID      uuid.UUID `json:"batch_id"`
	TotalCount   int       `json:"total_count"`
	SuccessCount int       `json:"success_count"`
	FailureCount int       `json:"failure_count"`
	CompletedAt  time.Time `json:"completed_at"`
}

// NewBatchCompletedEvent creates a new batch completed event.
func NewBatchCompletedEvent(tenantID, batchID uuid.UUID, total, success, failure int) *BatchCompletedEvent {
	return &BatchCompletedEvent{
		BaseDomainEvent: NewBaseDomainEvent("batch.completed", "batch", batchID, tenantID, 1),
		BatchID:         batchID,
		TotalCount:      total,
		SuccessCount:    success,
		FailureCount:    failure,
		CompletedAt:     time.Now().UTC(),
	}
}

// ============================================================================
// External Event Handlers (Events from other services)
// ============================================================================

// ExternalEventType represents types of events from other services.
type ExternalEventType string

const (
	// IAM Service Events
	ExternalEventUserCreated       ExternalEventType = "user.created"
	ExternalEventUserActivated     ExternalEventType = "user.activated"
	ExternalEventUserDeactivated   ExternalEventType = "user.deactivated"
	ExternalEventPasswordChanged   ExternalEventType = "user.password_changed"
	ExternalEventEmailVerified     ExternalEventType = "user.email_verified"
	ExternalEventUserRoleAssigned  ExternalEventType = "user.role_assigned"

	// Customer Service Events
	ExternalEventCustomerCreated   ExternalEventType = "customer.created"
	ExternalEventCustomerUpdated   ExternalEventType = "customer.updated"
	ExternalEventCustomerConverted ExternalEventType = "customer.converted"
	ExternalEventCustomerChurned   ExternalEventType = "customer.churned"

	// Sales Service Events
	ExternalEventLeadCreated       ExternalEventType = "lead.created"
	ExternalEventLeadConverted     ExternalEventType = "lead.converted"
	ExternalEventLeadQualified     ExternalEventType = "lead.qualified"
	ExternalEventOpportunityCreated ExternalEventType = "opportunity.created"
	ExternalEventOpportunityWon    ExternalEventType = "opportunity.won"
	ExternalEventOpportunityLost   ExternalEventType = "opportunity.lost"
	ExternalEventDealCreated       ExternalEventType = "deal.created"
	ExternalEventDealFulfilled     ExternalEventType = "deal.fulfilled"
	ExternalEventDealCancelled     ExternalEventType = "deal.cancelled"
	ExternalEventInvoiceCreated    ExternalEventType = "deal.invoice_created"
	ExternalEventPaymentReceived   ExternalEventType = "deal.payment_received"
)

// ExternalEvent represents an event received from another service.
type ExternalEvent struct {
	EventID       uuid.UUID              `json:"event_id"`
	EventType     ExternalEventType      `json:"event_type"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	AggregateID   uuid.UUID              `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	OccurredAt    time.Time              `json:"occurred_at"`
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// GetString gets a string value from the payload.
func (e *ExternalEvent) GetString(key string) string {
	if val, ok := e.Payload[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetUUID gets a UUID value from the payload.
func (e *ExternalEvent) GetUUID(key string) uuid.UUID {
	if val, ok := e.Payload[key]; ok {
		if str, ok := val.(string); ok {
			if id, err := uuid.Parse(str); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

// GetInt gets an int value from the payload.
func (e *ExternalEvent) GetInt(key string) int {
	if val, ok := e.Payload[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// GetBool gets a bool value from the payload.
func (e *ExternalEvent) GetBool(key string) bool {
	if val, ok := e.Payload[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// NotificationTrigger represents a rule for triggering notifications from events.
type NotificationTrigger struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	EventType    ExternalEventType      `json:"event_type"`
	TemplateCode string                 `json:"template_code"`
	Channel      NotificationChannel    `json:"channel"`
	IsActive     bool                   `json:"is_active"`
	Conditions   map[string]interface{} `json:"conditions,omitempty"`
	DataMapping  map[string]string      `json:"data_mapping,omitempty"`
	Delay        int                    `json:"delay_seconds,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ShouldTrigger checks if the trigger should fire for the given event.
func (t *NotificationTrigger) ShouldTrigger(event *ExternalEvent) bool {
	if !t.IsActive {
		return false
	}
	if string(event.EventType) != string(t.EventType) {
		return false
	}
	if event.TenantID != t.TenantID {
		return false
	}
	// TODO: Evaluate conditions
	return true
}

// MapData maps event payload to template data using the data mapping.
func (t *NotificationTrigger) MapData(event *ExternalEvent) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy all payload data
	for k, v := range event.Payload {
		result[k] = v
	}

	// Apply custom mapping
	for targetKey, sourceKey := range t.DataMapping {
		if val, ok := event.Payload[sourceKey]; ok {
			result[targetKey] = val
		}
	}

	return result
}
