// Package dto contains the Data Transfer Objects for the Notification service.
package dto

import (
	"time"
)

// NotificationDTO represents a notification data transfer object.
type NotificationDTO struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	Type            string                 `json:"type"`
	Channel         string                 `json:"channel"`
	Priority        string                 `json:"priority"`
	Status          string                 `json:"status"`
	Subject         string                 `json:"subject,omitempty"`
	Body            string                 `json:"body"`
	HTMLBody        string                 `json:"html_body,omitempty"`
	Recipient       RecipientDTO           `json:"recipient"`
	Sender          *SenderDTO             `json:"sender,omitempty"`
	TemplateID      string                 `json:"template_id,omitempty"`
	TemplateVersion int                    `json:"template_version,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	SourceEvent     *SourceEventDTO        `json:"source_event,omitempty"`
	DeliveryInfo    *DeliveryInfoDTO       `json:"delivery_info,omitempty"`
	RetryInfo       *RetryInfoDTO          `json:"retry_info,omitempty"`
	TrackingInfo    *TrackingInfoDTO       `json:"tracking_info,omitempty"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	SentAt          *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt     *time.Time             `json:"delivered_at,omitempty"`
	ReadAt          *time.Time             `json:"read_at,omitempty"`
	FailedAt        *time.Time             `json:"failed_at,omitempty"`
	CancelledAt     *time.Time             `json:"cancelled_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Version         int                    `json:"version"`
}

// RecipientDTO represents a notification recipient.
type RecipientDTO struct {
	ID          string `json:"id,omitempty"`
	Type        string `json:"type"` // user, customer, external
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	DeviceToken string `json:"device_token,omitempty"`
	WebhookURL  string `json:"webhook_url,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Locale      string `json:"locale,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}

// SenderDTO represents a notification sender.
type SenderDTO struct {
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	ReplyTo  string `json:"reply_to,omitempty"`
	SenderID string `json:"sender_id,omitempty"`
}

// SourceEventDTO represents the source event that triggered the notification.
type SourceEventDTO struct {
	EventType     string `json:"event_type"`
	AggregateType string `json:"aggregate_type"`
	AggregateID   string `json:"aggregate_id"`
	EventID       string `json:"event_id,omitempty"`
}

// DeliveryInfoDTO represents delivery information.
type DeliveryInfoDTO struct {
	Provider     string `json:"provider"`
	ProviderID   string `json:"provider_id,omitempty"`
	StatusCode   int    `json:"status_code,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// RetryInfoDTO represents retry information.
type RetryInfoDTO struct {
	RetryCount    int        `json:"retry_count"`
	MaxRetries    int        `json:"max_retries"`
	LastRetryAt   *time.Time `json:"last_retry_at,omitempty"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty"`
	RetryStrategy string     `json:"retry_strategy,omitempty"`
}

// TrackingInfoDTO represents email tracking information.
type TrackingInfoDTO struct {
	TrackOpens  bool       `json:"track_opens"`
	TrackClicks bool       `json:"track_clicks"`
	OpenedAt    *time.Time `json:"opened_at,omitempty"`
	OpenCount   int        `json:"open_count"`
	ClickedAt   *time.Time `json:"clicked_at,omitempty"`
	ClickCount  int        `json:"click_count"`
	ClickedURLs []string   `json:"clicked_urls,omitempty"`
}

// NotificationListDTO represents a paginated list of notifications.
type NotificationListDTO struct {
	Items      []NotificationDTO `json:"items"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// NotificationSummaryDTO represents a summary of a notification.
type NotificationSummaryDTO struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Type        string    `json:"type"`
	Channel     string    `json:"channel"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	Subject     string    `json:"subject,omitempty"`
	RecipientID string    `json:"recipient_id"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// NotificationStatsDTO represents notification statistics.
type NotificationStatsDTO struct {
	TenantID     string                       `json:"tenant_id"`
	Period       string                       `json:"period"`
	StartDate    time.Time                    `json:"start_date"`
	EndDate      time.Time                    `json:"end_date"`
	TotalSent    int64                        `json:"total_sent"`
	TotalDelivered int64                      `json:"total_delivered"`
	TotalFailed  int64                        `json:"total_failed"`
	TotalOpened  int64                        `json:"total_opened"`
	TotalClicked int64                        `json:"total_clicked"`
	ByChannel    map[string]ChannelStatsDTO   `json:"by_channel"`
	ByType       map[string]TypeStatsDTO      `json:"by_type"`
	ByStatus     map[string]int64             `json:"by_status"`
}

// ChannelStatsDTO represents statistics for a notification channel.
type ChannelStatsDTO struct {
	Channel      string  `json:"channel"`
	Sent         int64   `json:"sent"`
	Delivered    int64   `json:"delivered"`
	Failed       int64   `json:"failed"`
	Opened       int64   `json:"opened"`
	Clicked      int64   `json:"clicked"`
	DeliveryRate float64 `json:"delivery_rate"`
	OpenRate     float64 `json:"open_rate"`
	ClickRate    float64 `json:"click_rate"`
}

// TypeStatsDTO represents statistics for a notification type.
type TypeStatsDTO struct {
	Type         string  `json:"type"`
	Sent         int64   `json:"sent"`
	Delivered    int64   `json:"delivered"`
	Failed       int64   `json:"failed"`
	DeliveryRate float64 `json:"delivery_rate"`
}

// === Request DTOs ===

// SendEmailRequest represents a request to send an email notification.
type SendEmailRequest struct {
	TenantID        string                 `json:"tenant_id" validate:"required,uuid"`
	Type            string                 `json:"type" validate:"required"`
	Priority        string                 `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	To              []string               `json:"to" validate:"required,min=1,dive,email"`
	CC              []string               `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC             []string               `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	From            string                 `json:"from,omitempty" validate:"omitempty,email"`
	FromName        string                 `json:"from_name,omitempty"`
	ReplyTo         string                 `json:"reply_to,omitempty" validate:"omitempty,email"`
	Subject         string                 `json:"subject" validate:"required_without=TemplateID,max=500"`
	Body            string                 `json:"body,omitempty"`
	HTMLBody        string                 `json:"html_body,omitempty"`
	TemplateID      string                 `json:"template_id,omitempty" validate:"omitempty,uuid"`
	TemplateVersion *int                   `json:"template_version,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	Attachments     []AttachmentDTO        `json:"attachments,omitempty" validate:"omitempty,max=10,dive"`
	Tags            []string               `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	TrackOpens      *bool                  `json:"track_opens,omitempty"`
	TrackClicks     *bool                  `json:"track_clicks,omitempty"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	SourceEvent     *SourceEventDTO        `json:"source_event,omitempty"`
	IdempotencyKey  string                 `json:"idempotency_key,omitempty" validate:"omitempty,max=100"`
}

// AttachmentDTO represents an email attachment.
type AttachmentDTO struct {
	Filename    string `json:"filename" validate:"required,max=255"`
	ContentType string `json:"content_type" validate:"required"`
	Content     string `json:"content" validate:"required"` // Base64 encoded
	ContentID   string `json:"content_id,omitempty"`        // For inline attachments
}

// SendSMSRequest represents a request to send an SMS notification.
type SendSMSRequest struct {
	TenantID        string                 `json:"tenant_id" validate:"required,uuid"`
	Type            string                 `json:"type" validate:"required"`
	Priority        string                 `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	To              string                 `json:"to" validate:"required"`
	From            string                 `json:"from,omitempty"`
	Body            string                 `json:"body" validate:"required_without=TemplateID,max=1600"`
	TemplateID      string                 `json:"template_id,omitempty" validate:"omitempty,uuid"`
	TemplateVersion *int                   `json:"template_version,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	Tags            []string               `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Unicode         *bool                  `json:"unicode,omitempty"`
	Flash           *bool                  `json:"flash,omitempty"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	SourceEvent     *SourceEventDTO        `json:"source_event,omitempty"`
	IdempotencyKey  string                 `json:"idempotency_key,omitempty" validate:"omitempty,max=100"`
}

// SendInAppRequest represents a request to send an in-app notification.
type SendInAppRequest struct {
	TenantID        string                 `json:"tenant_id" validate:"required,uuid"`
	Type            string                 `json:"type" validate:"required"`
	Priority        string                 `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	UserID          string                 `json:"user_id" validate:"required,uuid"`
	Title           string                 `json:"title" validate:"required_without=TemplateID,max=255"`
	Body            string                 `json:"body" validate:"required_without=TemplateID,max=2000"`
	Category        string                 `json:"category,omitempty" validate:"omitempty,max=50"`
	ActionURL       string                 `json:"action_url,omitempty" validate:"omitempty,url,max=2000"`
	ActionText      string                 `json:"action_text,omitempty" validate:"omitempty,max=50"`
	ImageURL        string                 `json:"image_url,omitempty" validate:"omitempty,url,max=2000"`
	TemplateID      string                 `json:"template_id,omitempty" validate:"omitempty,uuid"`
	TemplateVersion *int                   `json:"template_version,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	Data            map[string]interface{} `json:"data,omitempty"`
	Tags            []string               `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	SourceEvent     *SourceEventDTO        `json:"source_event,omitempty"`
	IdempotencyKey  string                 `json:"idempotency_key,omitempty" validate:"omitempty,max=100"`
}

// SendPushRequest represents a request to send a push notification.
type SendPushRequest struct {
	TenantID        string                 `json:"tenant_id" validate:"required,uuid"`
	Type            string                 `json:"type" validate:"required"`
	Priority        string                 `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	UserID          string                 `json:"user_id" validate:"required,uuid"`
	DeviceToken     string                 `json:"device_token,omitempty"`
	Platform        string                 `json:"platform,omitempty" validate:"omitempty,oneof=ios android web"`
	Title           string                 `json:"title" validate:"required_without=TemplateID,max=255"`
	Body            string                 `json:"body" validate:"required_without=TemplateID,max=2000"`
	ImageURL        string                 `json:"image_url,omitempty" validate:"omitempty,url,max=2000"`
	ClickAction     string                 `json:"click_action,omitempty" validate:"omitempty,max=255"`
	Category        string                 `json:"category,omitempty" validate:"omitempty,max=50"`
	Badge           *int                   `json:"badge,omitempty"`
	Sound           string                 `json:"sound,omitempty" validate:"omitempty,max=100"`
	TemplateID      string                 `json:"template_id,omitempty" validate:"omitempty,uuid"`
	TemplateVersion *int                   `json:"template_version,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	Data            map[string]string      `json:"data,omitempty"`
	Tags            []string               `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	TTL             *int                   `json:"ttl,omitempty"` // Time-to-live in seconds
	CollapseKey     string                 `json:"collapse_key,omitempty" validate:"omitempty,max=100"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	SourceEvent     *SourceEventDTO        `json:"source_event,omitempty"`
	IdempotencyKey  string                 `json:"idempotency_key,omitempty" validate:"omitempty,max=100"`
}

// SendWebhookRequest represents a request to send a webhook notification.
type SendWebhookRequest struct {
	TenantID        string                 `json:"tenant_id" validate:"required,uuid"`
	Type            string                 `json:"type" validate:"required"`
	Priority        string                 `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	URL             string                 `json:"url" validate:"required,url"`
	Method          string                 `json:"method" validate:"omitempty,oneof=GET POST PUT PATCH DELETE"`
	Headers         map[string]string      `json:"headers,omitempty"`
	Body            map[string]interface{} `json:"body,omitempty"`
	ContentType     string                 `json:"content_type" validate:"omitempty"`
	Timeout         *int                   `json:"timeout,omitempty"` // Timeout in seconds
	SignatureKey    string                 `json:"signature_key,omitempty"`
	TemplateID      string                 `json:"template_id,omitempty" validate:"omitempty,uuid"`
	TemplateVersion *int                   `json:"template_version,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	Tags            []string               `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	SourceEvent     *SourceEventDTO        `json:"source_event,omitempty"`
	IdempotencyKey  string                 `json:"idempotency_key,omitempty" validate:"omitempty,max=100"`
}

// SendBatchRequest represents a request to send batch notifications.
type SendBatchRequest struct {
	TenantID       string                 `json:"tenant_id" validate:"required,uuid"`
	Type           string                 `json:"type" validate:"required"`
	Channel        string                 `json:"channel" validate:"required,oneof=email sms push in_app webhook"`
	Priority       string                 `json:"priority" validate:"omitempty,oneof=low normal high critical"`
	Recipients     []RecipientDTO         `json:"recipients" validate:"required,min=1,max=1000,dive"`
	Subject        string                 `json:"subject,omitempty" validate:"omitempty,max=500"`
	Body           string                 `json:"body,omitempty" validate:"omitempty,max=10000"`
	HTMLBody       string                 `json:"html_body,omitempty"`
	TemplateID     string                 `json:"template_id,omitempty" validate:"omitempty,uuid"`
	TemplateVersion *int                  `json:"template_version,omitempty"`
	Variables      map[string]interface{} `json:"variables,omitempty"`
	Tags           []string               `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ScheduledAt    *time.Time             `json:"scheduled_at,omitempty"`
	SourceEvent    *SourceEventDTO        `json:"source_event,omitempty"`
	IdempotencyKey string                 `json:"idempotency_key,omitempty" validate:"omitempty,max=100"`
}

// RetryNotificationRequest represents a request to retry a failed notification.
type RetryNotificationRequest struct {
	NotificationID string `json:"notification_id" validate:"required,uuid"`
	Force          bool   `json:"force,omitempty"`
}

// CancelNotificationRequest represents a request to cancel a notification.
type CancelNotificationRequest struct {
	NotificationID string `json:"notification_id" validate:"required,uuid"`
	Reason         string `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// GetNotificationRequest represents a request to get a notification.
type GetNotificationRequest struct {
	TenantID       string `json:"tenant_id" validate:"required,uuid"`
	NotificationID string `json:"notification_id" validate:"required,uuid"`
}

// ListNotificationsRequest represents a request to list notifications.
type ListNotificationsRequest struct {
	TenantID    string     `json:"tenant_id" validate:"required,uuid"`
	Channel     string     `json:"channel,omitempty" validate:"omitempty,oneof=email sms push in_app webhook slack whatsapp telegram"`
	Type        string     `json:"type,omitempty"`
	Status      string     `json:"status,omitempty"`
	Priority    string     `json:"priority,omitempty" validate:"omitempty,oneof=low normal high critical"`
	RecipientID string     `json:"recipient_id,omitempty" validate:"omitempty,uuid"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Page        int        `json:"page" validate:"min=1"`
	PageSize    int        `json:"page_size" validate:"min=1,max=100"`
	SortBy      string     `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at sent_at status priority"`
	SortOrder   string     `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// GetNotificationStatsRequest represents a request to get notification statistics.
type GetNotificationStatsRequest struct {
	TenantID  string     `json:"tenant_id" validate:"required,uuid"`
	StartDate time.Time  `json:"start_date" validate:"required"`
	EndDate   time.Time  `json:"end_date" validate:"required"`
	Channel   string     `json:"channel,omitempty"`
	Type      string     `json:"type,omitempty"`
	GroupBy   string     `json:"group_by,omitempty" validate:"omitempty,oneof=day week month channel type"`
}

// === Response DTOs ===

// SendNotificationResponse represents a response after sending a notification.
type SendNotificationResponse struct {
	NotificationID string    `json:"notification_id"`
	Status         string    `json:"status"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	Message        string    `json:"message,omitempty"`
}

// SendBatchResponse represents a response after sending batch notifications.
type SendBatchResponse struct {
	BatchID        string                        `json:"batch_id"`
	TotalCount     int                           `json:"total_count"`
	AcceptedCount  int                           `json:"accepted_count"`
	RejectedCount  int                           `json:"rejected_count"`
	Notifications  []SendNotificationResponse    `json:"notifications,omitempty"`
	RejectedItems  []RejectedNotificationDTO     `json:"rejected_items,omitempty"`
	Message        string                        `json:"message,omitempty"`
}

// RejectedNotificationDTO represents a rejected notification in a batch.
type RejectedNotificationDTO struct {
	Index        int    `json:"index"`
	RecipientID  string `json:"recipient_id,omitempty"`
	Recipient    string `json:"recipient,omitempty"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

// RetryNotificationResponse represents a response after retrying a notification.
type RetryNotificationResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
	RetryCount     int    `json:"retry_count"`
	Message        string `json:"message,omitempty"`
}

// CancelNotificationResponse represents a response after cancelling a notification.
type CancelNotificationResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
	Message        string `json:"message,omitempty"`
}

// DeliveryStatusResponse represents a delivery status response.
type DeliveryStatusResponse struct {
	NotificationID string     `json:"notification_id"`
	Channel        string     `json:"channel"`
	Status         string     `json:"status"`
	Provider       string     `json:"provider,omitempty"`
	ProviderID     string     `json:"provider_id,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	OpenedAt       *time.Time `json:"opened_at,omitempty"`
	ClickedAt      *time.Time `json:"clicked_at,omitempty"`
	FailedAt       *time.Time `json:"failed_at,omitempty"`
	ErrorCode      string     `json:"error_code,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	RetryCount     int        `json:"retry_count"`
}

// === Preference DTOs ===

// NotificationPreferenceDTO represents a notification preference.
type NotificationPreferenceDTO struct {
	ID               string                      `json:"id"`
	TenantID         string                      `json:"tenant_id"`
	UserID           string                      `json:"user_id"`
	ChannelSettings  map[string]ChannelSettingDTO `json:"channel_settings"`
	TypeSettings     map[string]TypeSettingDTO    `json:"type_settings"`
	QuietHours       *QuietHoursDTO              `json:"quiet_hours,omitempty"`
	Timezone         string                      `json:"timezone"`
	GlobalOptOut     bool                        `json:"global_opt_out"`
	CreatedAt        time.Time                   `json:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at"`
}

// ChannelSettingDTO represents settings for a notification channel.
type ChannelSettingDTO struct {
	Channel   string `json:"channel"`
	Enabled   bool   `json:"enabled"`
	OptedOut  bool   `json:"opted_out"`
}

// TypeSettingDTO represents settings for a notification type.
type TypeSettingDTO struct {
	Type     string   `json:"type"`
	Enabled  bool     `json:"enabled"`
	Channels []string `json:"channels,omitempty"`
}

// QuietHoursDTO represents quiet hours settings.
type QuietHoursDTO struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
	Days      []int  `json:"days"`       // 0-6, where 0 is Sunday
	Timezone  string `json:"timezone"`
}

// UpdatePreferenceRequest represents a request to update notification preferences.
type UpdatePreferenceRequest struct {
	TenantID        string                       `json:"tenant_id" validate:"required,uuid"`
	UserID          string                       `json:"user_id" validate:"required,uuid"`
	ChannelSettings map[string]ChannelSettingDTO `json:"channel_settings,omitempty"`
	TypeSettings    map[string]TypeSettingDTO    `json:"type_settings,omitempty"`
	QuietHours      *QuietHoursDTO               `json:"quiet_hours,omitempty"`
	Timezone        string                       `json:"timezone,omitempty"`
	GlobalOptOut    *bool                        `json:"global_opt_out,omitempty"`
}

// GetPreferenceRequest represents a request to get notification preferences.
type GetPreferenceRequest struct {
	TenantID string `json:"tenant_id" validate:"required,uuid"`
	UserID   string `json:"user_id" validate:"required,uuid"`
}
