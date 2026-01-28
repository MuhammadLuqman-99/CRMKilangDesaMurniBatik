// Package domain contains the domain layer for the Notification service.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a notification that can be sent through various channels.
type Notification struct {
	BaseAggregateRoot

	// Core identification
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Code     string    `json:"code" db:"code"` // Human-readable identifier (e.g., NOTIF-20240128-001)

	// Notification details
	Type        NotificationType     `json:"type" db:"type"`
	Channel     NotificationChannel  `json:"channel" db:"channel"`
	Priority    NotificationPriority `json:"priority" db:"priority"`
	Status      NotificationStatus   `json:"status" db:"status"`

	// Template reference (optional)
	TemplateID   *uuid.UUID `json:"template_id,omitempty" db:"template_id"`
	TemplateName string     `json:"template_name,omitempty" db:"template_name"`

	// Recipient information
	RecipientID    *uuid.UUID `json:"recipient_id,omitempty" db:"recipient_id"`       // User ID if applicable
	RecipientEmail string     `json:"recipient_email,omitempty" db:"recipient_email"` // Email address
	RecipientPhone string     `json:"recipient_phone,omitempty" db:"recipient_phone"` // Phone number
	RecipientName  string     `json:"recipient_name,omitempty" db:"recipient_name"`   // Display name
	DeviceToken    string     `json:"device_token,omitempty" db:"device_token"`       // Push notification token

	// Content
	Subject     string                 `json:"subject,omitempty" db:"subject"`
	Body        string                 `json:"body" db:"body"`
	HTMLBody    string                 `json:"html_body,omitempty" db:"html_body"`
	Data        map[string]interface{} `json:"data,omitempty" db:"data"`            // Template variables / payload
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`    // Additional metadata
	Attachments []Attachment           `json:"attachments,omitempty" db:"-"`

	// Email specific fields
	FromAddress   string   `json:"from_address,omitempty" db:"from_address"`
	FromName      string   `json:"from_name,omitempty" db:"from_name"`
	ReplyTo       string   `json:"reply_to,omitempty" db:"reply_to"`
	CC            []string `json:"cc,omitempty" db:"-"`
	BCC           []string `json:"bcc,omitempty" db:"-"`
	Headers       map[string]string `json:"headers,omitempty" db:"-"`

	// Scheduling
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`

	// Delivery tracking
	SentAt       *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	ReadAt       *time.Time `json:"read_at,omitempty" db:"read_at"`
	FailedAt     *time.Time `json:"failed_at,omitempty" db:"failed_at"`
	CancelledAt  *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`

	// Retry tracking
	AttemptCount int        `json:"attempt_count" db:"attempt_count"`
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty" db:"last_attempt_at"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty" db:"next_retry_at"`
	RetryPolicy   *RetryPolicy `json:"retry_policy,omitempty" db:"-"`

	// Error information
	ErrorCode    string `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage string `json:"error_message,omitempty" db:"error_message"`
	ProviderError string `json:"provider_error,omitempty" db:"provider_error"`

	// Provider information
	Provider          string `json:"provider,omitempty" db:"provider"`
	ProviderMessageID string `json:"provider_message_id,omitempty" db:"provider_message_id"`

	// Tracking
	TrackOpens  bool `json:"track_opens" db:"track_opens"`
	TrackClicks bool `json:"track_clicks" db:"track_clicks"`
	OpenCount   int  `json:"open_count" db:"open_count"`
	ClickCount  int  `json:"click_count" db:"click_count"`

	// Reference to source event/entity
	SourceEvent     string     `json:"source_event,omitempty" db:"source_event"`
	SourceEntityID  *uuid.UUID `json:"source_entity_id,omitempty" db:"source_entity_id"`
	SourceEntityType string    `json:"source_entity_type,omitempty" db:"source_entity_type"`
	CorrelationID   string     `json:"correlation_id,omitempty" db:"correlation_id"`

	// Batch information
	BatchID    *uuid.UUID `json:"batch_id,omitempty" db:"batch_id"`
	BatchIndex int        `json:"batch_index,omitempty" db:"batch_index"`

	// Audit
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
}

// Attachment represents a file attachment for email notifications.
type Attachment struct {
	ID          uuid.UUID `json:"id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	URL         string    `json:"url,omitempty"`
	Content     []byte    `json:"-"` // Binary content (not serialized)
	Inline      bool      `json:"inline"`
	ContentID   string    `json:"content_id,omitempty"` // For inline images
}

// NewNotification creates a new notification.
func NewNotification(
	tenantID uuid.UUID,
	notificationType NotificationType,
	channel NotificationChannel,
	body string,
) (*Notification, error) {
	// Validate inputs
	if !notificationType.IsValid() {
		return nil, NewValidationError("type", "invalid notification type", "INVALID_TYPE")
	}
	if !channel.IsValid() {
		return nil, ErrInvalidChannel
	}
	if body == "" {
		return nil, NewValidationError("body", "notification body is required", "REQUIRED")
	}

	n := &Notification{
		BaseAggregateRoot: NewBaseAggregateRoot(),
		TenantID:          tenantID,
		Type:              notificationType,
		Channel:           channel,
		Priority:          PriorityNormal,
		Status:            StatusPending,
		Body:              body,
		Data:              make(map[string]interface{}),
		Metadata:          make(map[string]interface{}),
		RetryPolicy:       &RetryPolicy{MaxAttempts: 3, InitialInterval: 60, MaxInterval: 3600, Multiplier: 2.0},
	}

	// Generate notification code
	n.Code = generateNotificationCode()

	return n, nil
}

// generateNotificationCode generates a unique notification code.
func generateNotificationCode() string {
	now := time.Now().UTC()
	return "NOTIF-" + now.Format("20060102") + "-" + uuid.New().String()[:8]
}

// ============================================================================
// Recipient Methods
// ============================================================================

// SetRecipientEmail sets the email recipient.
func (n *Notification) SetRecipientEmail(email, name string) error {
	if n.Channel != ChannelEmail {
		return ErrInvalidChannel
	}
	if _, err := NewEmailAddress(email); err != nil {
		return err
	}
	n.RecipientEmail = email
	n.RecipientName = name
	n.MarkUpdated()
	return nil
}

// SetRecipientPhone sets the phone recipient.
func (n *Notification) SetRecipientPhone(phone, name string) error {
	if n.Channel != ChannelSMS && n.Channel != ChannelWhatsApp {
		return ErrInvalidChannel
	}
	if _, err := NewPhoneNumber(phone); err != nil {
		return err
	}
	n.RecipientPhone = phone
	n.RecipientName = name
	n.MarkUpdated()
	return nil
}

// SetRecipientUser sets the user recipient for in-app notifications.
func (n *Notification) SetRecipientUser(userID uuid.UUID, name string) error {
	if n.Channel != ChannelInApp {
		return ErrInvalidChannel
	}
	n.RecipientID = &userID
	n.RecipientName = name
	n.MarkUpdated()
	return nil
}

// SetRecipientDevice sets the device token for push notifications.
func (n *Notification) SetRecipientDevice(userID uuid.UUID, deviceToken, name string) error {
	if n.Channel != ChannelPush {
		return ErrInvalidChannel
	}
	if deviceToken == "" {
		return ErrDeviceTokenRequired
	}
	n.RecipientID = &userID
	n.DeviceToken = deviceToken
	n.RecipientName = name
	n.MarkUpdated()
	return nil
}

// ============================================================================
// Content Methods
// ============================================================================

// SetSubject sets the notification subject.
func (n *Notification) SetSubject(subject string) error {
	if len(subject) > 255 {
		return ErrEmailSubjectTooLong
	}
	n.Subject = subject
	n.MarkUpdated()
	return nil
}

// SetHTMLBody sets the HTML body for email notifications.
func (n *Notification) SetHTMLBody(html string) {
	n.HTMLBody = html
	n.MarkUpdated()
}

// SetData sets the template data/payload.
func (n *Notification) SetData(data map[string]interface{}) {
	n.Data = data
	n.MarkUpdated()
}

// AddData adds a single data field.
func (n *Notification) AddData(key string, value interface{}) {
	if n.Data == nil {
		n.Data = make(map[string]interface{})
	}
	n.Data[key] = value
	n.MarkUpdated()
}

// SetTemplate associates a template with this notification.
func (n *Notification) SetTemplate(templateID uuid.UUID, templateName string) {
	n.TemplateID = &templateID
	n.TemplateName = templateName
	n.MarkUpdated()
}

// AddAttachment adds an attachment to the notification.
func (n *Notification) AddAttachment(attachment Attachment) error {
	if n.Channel != ChannelEmail {
		return NewValidationError("attachment", "attachments only supported for email channel", "UNSUPPORTED")
	}
	if len(n.Attachments) >= 10 {
		return ErrTooManyAttachments
	}
	if attachment.Size > 25*1024*1024 { // 25 MB limit
		return ErrAttachmentTooLarge
	}
	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}
	n.Attachments = append(n.Attachments, attachment)
	n.MarkUpdated()
	return nil
}

// ============================================================================
// Email Configuration Methods
// ============================================================================

// SetFromAddress sets the from address for email notifications.
func (n *Notification) SetFromAddress(email, name string) error {
	if n.Channel != ChannelEmail {
		return ErrInvalidChannel
	}
	if email != "" {
		if _, err := NewEmailAddress(email); err != nil {
			return ErrInvalidFromAddress
		}
	}
	n.FromAddress = email
	n.FromName = name
	n.MarkUpdated()
	return nil
}

// SetReplyTo sets the reply-to address.
func (n *Notification) SetReplyTo(email string) error {
	if n.Channel != ChannelEmail {
		return ErrInvalidChannel
	}
	if email != "" {
		if _, err := NewEmailAddress(email); err != nil {
			return ErrInvalidReplyToAddress
		}
	}
	n.ReplyTo = email
	n.MarkUpdated()
	return nil
}

// AddCC adds a CC recipient.
func (n *Notification) AddCC(email string) error {
	if n.Channel != ChannelEmail {
		return ErrInvalidChannel
	}
	if _, err := NewEmailAddress(email); err != nil {
		return err
	}
	n.CC = append(n.CC, email)
	n.MarkUpdated()
	return nil
}

// AddBCC adds a BCC recipient.
func (n *Notification) AddBCC(email string) error {
	if n.Channel != ChannelEmail {
		return ErrInvalidChannel
	}
	if _, err := NewEmailAddress(email); err != nil {
		return err
	}
	n.BCC = append(n.BCC, email)
	n.MarkUpdated()
	return nil
}

// SetHeader sets a custom email header.
func (n *Notification) SetHeader(key, value string) {
	if n.Headers == nil {
		n.Headers = make(map[string]string)
	}
	n.Headers[key] = value
	n.MarkUpdated()
}

// ============================================================================
// Status Transition Methods
// ============================================================================

// Queue marks the notification as queued for delivery.
func (n *Notification) Queue() error {
	if !n.Status.CanTransitionTo(StatusQueued) {
		return NewValidationError("status", "cannot queue notification in current status", "INVALID_TRANSITION")
	}
	n.Status = StatusQueued
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationQueuedEvent(n))
	return nil
}

// Schedule schedules the notification for future delivery.
func (n *Notification) Schedule(scheduledAt time.Time) error {
	if scheduledAt.Before(time.Now().UTC()) {
		return ErrScheduledTimeInPast
	}
	if !n.Status.CanTransitionTo(StatusScheduled) {
		return NewValidationError("status", "cannot schedule notification in current status", "INVALID_TRANSITION")
	}
	n.Status = StatusScheduled
	n.ScheduledAt = &scheduledAt
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationScheduledEvent(n))
	return nil
}

// MarkSending marks the notification as being sent.
func (n *Notification) MarkSending() error {
	if !n.Status.CanTransitionTo(StatusSending) {
		return NewValidationError("status", "cannot mark as sending in current status", "INVALID_TRANSITION")
	}
	n.Status = StatusSending
	now := time.Now().UTC()
	n.LastAttemptAt = &now
	n.AttemptCount++
	n.MarkUpdated()
	return nil
}

// MarkSent marks the notification as sent.
func (n *Notification) MarkSent(providerMessageID string) error {
	if !n.Status.CanTransitionTo(StatusSent) {
		return NewValidationError("status", "cannot mark as sent in current status", "INVALID_TRANSITION")
	}
	n.Status = StatusSent
	now := time.Now().UTC()
	n.SentAt = &now
	n.ProviderMessageID = providerMessageID
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationSentEvent(n))
	return nil
}

// MarkDelivered marks the notification as delivered.
func (n *Notification) MarkDelivered() error {
	if !n.Status.CanTransitionTo(StatusDelivered) {
		return NewValidationError("status", "cannot mark as delivered in current status", "INVALID_TRANSITION")
	}
	n.Status = StatusDelivered
	now := time.Now().UTC()
	n.DeliveredAt = &now
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationDeliveredEvent(n))
	return nil
}

// MarkRead marks the notification as read (for in-app notifications).
func (n *Notification) MarkRead() error {
	if !n.Status.CanTransitionTo(StatusRead) {
		return NewValidationError("status", "cannot mark as read in current status", "INVALID_TRANSITION")
	}
	n.Status = StatusRead
	now := time.Now().UTC()
	n.ReadAt = &now
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationReadEvent(n))
	return nil
}

// MarkFailed marks the notification as failed.
func (n *Notification) MarkFailed(errorCode, errorMessage, providerError string) error {
	n.Status = StatusFailed
	now := time.Now().UTC()
	n.FailedAt = &now
	n.ErrorCode = errorCode
	n.ErrorMessage = errorMessage
	n.ProviderError = providerError
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationFailedEvent(n))
	return nil
}

// MarkRetrying marks the notification for retry.
func (n *Notification) MarkRetrying(nextRetryAt time.Time) error {
	if !n.Status.CanTransitionTo(StatusRetrying) {
		return NewValidationError("status", "cannot retry notification in current status", "INVALID_TRANSITION")
	}
	if n.RetryPolicy != nil && !n.RetryPolicy.ShouldRetry(n.AttemptCount) {
		return ErrMaxRetriesExceeded
	}
	n.Status = StatusRetrying
	n.NextRetryAt = &nextRetryAt
	n.MarkUpdated()
	n.AddDomainEvent(NewNotificationRetryScheduledEvent(n))
	return nil
}

// Cancel cancels the notification.
func (n *Notification) Cancel() error {
	if n.Status.IsFinal() {
		return ErrCannotCancel
	}
	if n.Status == StatusSending {
		return ErrCannotCancel
	}
	n.Status = StatusCancelled
	now := time.Now().UTC()
	n.CancelledAt = &now
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationCancelledEvent(n))
	return nil
}

// MarkBounced marks an email as bounced.
func (n *Notification) MarkBounced(bounceType, bounceMessage string) error {
	if n.Channel != ChannelEmail {
		return ErrInvalidChannel
	}
	if !n.Status.CanTransitionTo(StatusBounced) {
		return NewValidationError("status", "cannot mark as bounced in current status", "INVALID_TRANSITION")
	}
	n.Status = StatusBounced
	n.ErrorCode = "BOUNCED"
	n.ErrorMessage = bounceType + ": " + bounceMessage
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationBouncedEvent(n, bounceType))
	return nil
}

// MarkComplained marks an email as complained (spam report).
func (n *Notification) MarkComplained() error {
	if n.Channel != ChannelEmail {
		return ErrInvalidChannel
	}
	if !n.Status.CanTransitionTo(StatusComplained) {
		return NewValidationError("status", "cannot mark as complained in current status", "INVALID_TRANSITION")
	}
	n.Status = StatusComplained
	n.ErrorCode = "COMPLAINED"
	n.ErrorMessage = "Recipient marked email as spam"
	n.MarkUpdated()
	n.IncrementVersion()
	n.AddDomainEvent(NewNotificationComplainedEvent(n))
	return nil
}

// ============================================================================
// Tracking Methods
// ============================================================================

// RecordOpen records an email open event.
func (n *Notification) RecordOpen() {
	n.OpenCount++
	n.MarkUpdated()
}

// RecordClick records a link click event.
func (n *Notification) RecordClick() {
	n.ClickCount++
	n.MarkUpdated()
}

// EnableTracking enables open and click tracking.
func (n *Notification) EnableTracking(opens, clicks bool) {
	n.TrackOpens = opens
	n.TrackClicks = clicks
	n.MarkUpdated()
}

// ============================================================================
// Source Reference Methods
// ============================================================================

// SetSourceEvent sets the source event that triggered this notification.
func (n *Notification) SetSourceEvent(eventType string, entityID *uuid.UUID, entityType string) {
	n.SourceEvent = eventType
	n.SourceEntityID = entityID
	n.SourceEntityType = entityType
	n.MarkUpdated()
}

// SetCorrelationID sets the correlation ID for tracing.
func (n *Notification) SetCorrelationID(correlationID string) {
	n.CorrelationID = correlationID
	n.MarkUpdated()
}

// SetBatch sets the batch information.
func (n *Notification) SetBatch(batchID uuid.UUID, index int) {
	n.BatchID = &batchID
	n.BatchIndex = index
	n.MarkUpdated()
}

// ============================================================================
// Priority Methods
// ============================================================================

// SetPriority sets the notification priority.
func (n *Notification) SetPriority(priority NotificationPriority) error {
	if !priority.IsValid() {
		return ErrInvalidPriority
	}
	n.Priority = priority
	n.MarkUpdated()
	return nil
}

// ============================================================================
// Validation Methods
// ============================================================================

// Validate validates the notification before sending.
func (n *Notification) Validate() error {
	var errs ValidationErrors

	// Basic validation
	if n.TenantID == uuid.Nil {
		errs.AddField("tenant_id", "tenant ID is required", "REQUIRED")
	}
	if n.Body == "" {
		errs.AddField("body", "notification body is required", "REQUIRED")
	}

	// Channel-specific validation
	switch n.Channel {
	case ChannelEmail:
		if n.RecipientEmail == "" {
			errs.AddField("recipient_email", "email address is required for email channel", "REQUIRED")
		}
		if n.Subject == "" {
			errs.AddField("subject", "subject is required for email notifications", "REQUIRED")
		}
	case ChannelSMS, ChannelWhatsApp:
		if n.RecipientPhone == "" {
			errs.AddField("recipient_phone", "phone number is required for SMS channel", "REQUIRED")
		}
	case ChannelPush:
		if n.DeviceToken == "" {
			errs.AddField("device_token", "device token is required for push notifications", "REQUIRED")
		}
	case ChannelInApp:
		if n.RecipientID == nil {
			errs.AddField("recipient_id", "user ID is required for in-app notifications", "REQUIRED")
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// CanRetry returns true if the notification can be retried.
func (n *Notification) CanRetry() bool {
	if n.RetryPolicy == nil {
		return false
	}
	return n.RetryPolicy.ShouldRetry(n.AttemptCount) &&
		(n.Status == StatusFailed || n.Status == StatusRetrying)
}

// GetNextRetryInterval returns the next retry interval in seconds.
func (n *Notification) GetNextRetryInterval() int {
	if n.RetryPolicy == nil {
		return 0
	}
	return n.RetryPolicy.GetInterval(n.AttemptCount)
}

// ============================================================================
// Query Methods
// ============================================================================

// IsPending returns true if the notification is pending.
func (n *Notification) IsPending() bool {
	return n.Status == StatusPending || n.Status == StatusScheduled
}

// IsDelivered returns true if the notification was delivered.
func (n *Notification) IsDelivered() bool {
	return n.Status == StatusDelivered || n.Status == StatusRead
}

// IsFailed returns true if the notification failed.
func (n *Notification) IsFailed() bool {
	return n.Status == StatusFailed || n.Status == StatusBounced || n.Status == StatusComplained
}

// IsCancelled returns true if the notification was cancelled.
func (n *Notification) IsCancelled() bool {
	return n.Status == StatusCancelled
}

// IsScheduled returns true if the notification is scheduled.
func (n *Notification) IsScheduled() bool {
	return n.Status == StatusScheduled && n.ScheduledAt != nil
}

// IsDue returns true if a scheduled notification is due for sending.
func (n *Notification) IsDue() bool {
	if !n.IsScheduled() {
		return false
	}
	return n.ScheduledAt.Before(time.Now().UTC()) || n.ScheduledAt.Equal(time.Now().UTC())
}
