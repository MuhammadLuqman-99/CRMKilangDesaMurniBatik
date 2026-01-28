// Package ports defines the port interfaces for the Notification service application layer.
package ports

import (
	"context"
	"time"

	"github.com/kilang-desa-murni/crm/internal/notification/domain"
)

// EmailProvider defines the interface for email delivery providers.
type EmailProvider interface {
	// SendEmail sends an email notification.
	SendEmail(ctx context.Context, request EmailRequest) (*EmailResponse, error)
	// ValidateEmail validates an email address.
	ValidateEmail(ctx context.Context, email string) (bool, error)
	// GetProviderName returns the provider name.
	GetProviderName() string
	// IsAvailable checks if the provider is available.
	IsAvailable(ctx context.Context) bool
}

// EmailRequest represents a request to send an email.
type EmailRequest struct {
	MessageID    string
	From         string
	ReplyTo      string
	To           []string
	CC           []string
	BCC          []string
	Subject      string
	HTMLBody     string
	TextBody     string
	Attachments  []EmailAttachment
	Headers      map[string]string
	Tags         map[string]string
	TrackOpens   bool
	TrackClicks  bool
	ScheduledAt  *time.Time
}

// EmailAttachment represents an email attachment.
type EmailAttachment struct {
	Filename    string
	ContentType string
	Content     []byte
	ContentID   string // For inline attachments
}

// EmailResponse represents the response from sending an email.
type EmailResponse struct {
	MessageID    string
	ProviderID   string
	Provider     string
	Status       string
	StatusCode   int
	ErrorMessage string
	SentAt       time.Time
}

// SMSProvider defines the interface for SMS delivery providers.
type SMSProvider interface {
	// SendSMS sends an SMS notification.
	SendSMS(ctx context.Context, request SMSRequest) (*SMSResponse, error)
	// ValidatePhoneNumber validates a phone number.
	ValidatePhoneNumber(ctx context.Context, phone string) (bool, error)
	// GetProviderName returns the provider name.
	GetProviderName() string
	// IsAvailable checks if the provider is available.
	IsAvailable(ctx context.Context) bool
	// GetDeliveryStatus retrieves the delivery status for a message.
	GetDeliveryStatus(ctx context.Context, messageID string) (*SMSDeliveryStatus, error)
}

// SMSRequest represents a request to send an SMS.
type SMSRequest struct {
	MessageID   string
	From        string
	To          string
	Body        string
	Unicode     bool
	Flash       bool
	ScheduledAt *time.Time
	Tags        map[string]string
}

// SMSResponse represents the response from sending an SMS.
type SMSResponse struct {
	MessageID    string
	ProviderID   string
	Provider     string
	Status       string
	StatusCode   int
	ErrorMessage string
	SegmentCount int
	SentAt       time.Time
}

// SMSDeliveryStatus represents the delivery status of an SMS.
type SMSDeliveryStatus struct {
	MessageID     string
	ProviderID    string
	Status        string
	DeliveredAt   *time.Time
	ErrorCode     string
	ErrorMessage  string
}

// PushProvider defines the interface for push notification providers.
type PushProvider interface {
	// SendPush sends a push notification.
	SendPush(ctx context.Context, request PushRequest) (*PushResponse, error)
	// ValidateDeviceToken validates a device token.
	ValidateDeviceToken(ctx context.Context, token string, platform string) (bool, error)
	// GetProviderName returns the provider name.
	GetProviderName() string
	// IsAvailable checks if the provider is available.
	IsAvailable(ctx context.Context) bool
}

// PushRequest represents a request to send a push notification.
type PushRequest struct {
	MessageID    string
	DeviceToken  string
	Platform     string // ios, android, web
	Title        string
	Body         string
	ImageURL     string
	Data         map[string]string
	Badge        *int
	Sound        string
	ClickAction  string
	Category     string
	ThreadID     string
	Priority     string // high, normal
	TTL          int    // time-to-live in seconds
	CollapseKey  string
	ScheduledAt  *time.Time
}

// PushResponse represents the response from sending a push notification.
type PushResponse struct {
	MessageID       string
	ProviderID      string
	Provider        string
	Status          string
	StatusCode      int
	ErrorMessage    string
	FailureReason   string
	SentAt          time.Time
}

// WebhookProvider defines the interface for webhook delivery.
type WebhookProvider interface {
	// SendWebhook sends a webhook notification.
	SendWebhook(ctx context.Context, request WebhookRequest) (*WebhookResponse, error)
	// GetProviderName returns the provider name.
	GetProviderName() string
}

// WebhookRequest represents a request to send a webhook.
type WebhookRequest struct {
	MessageID    string
	URL          string
	Method       string
	Headers      map[string]string
	Body         []byte
	ContentType  string
	Timeout      time.Duration
	RetryCount   int
	SignatureKey string
}

// WebhookResponse represents the response from sending a webhook.
type WebhookResponse struct {
	MessageID    string
	URL          string
	Status       string
	StatusCode   int
	ResponseBody []byte
	ErrorMessage string
	SentAt       time.Time
	Duration     time.Duration
}

// InAppProvider defines the interface for in-app notification delivery.
type InAppProvider interface {
	// SendInApp sends an in-app notification.
	SendInApp(ctx context.Context, request InAppRequest) (*InAppResponse, error)
	// MarkAsRead marks a notification as read.
	MarkAsRead(ctx context.Context, userID, notificationID string) error
	// MarkAllAsRead marks all notifications as read for a user.
	MarkAllAsRead(ctx context.Context, userID string) error
	// GetUnreadCount gets the unread notification count for a user.
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	// GetProviderName returns the provider name.
	GetProviderName() string
}

// InAppRequest represents a request to send an in-app notification.
type InAppRequest struct {
	MessageID     string
	UserID        string
	Title         string
	Body          string
	Category      string
	Priority      string
	ActionURL     string
	ActionText    string
	ImageURL      string
	Data          map[string]interface{}
	ExpiresAt     *time.Time
	ScheduledAt   *time.Time
}

// InAppResponse represents the response from sending an in-app notification.
type InAppResponse struct {
	MessageID    string
	UserID       string
	Status       string
	ErrorMessage string
	SentAt       time.Time
}

// SlackProvider defines the interface for Slack notification delivery.
type SlackProvider interface {
	// SendSlackMessage sends a Slack notification.
	SendSlackMessage(ctx context.Context, request SlackRequest) (*SlackResponse, error)
	// GetProviderName returns the provider name.
	GetProviderName() string
	// IsAvailable checks if the provider is available.
	IsAvailable(ctx context.Context) bool
}

// SlackRequest represents a request to send a Slack message.
type SlackRequest struct {
	MessageID   string
	WebhookURL  string
	Channel     string
	Username    string
	IconEmoji   string
	IconURL     string
	Text        string
	Blocks      []map[string]interface{}
	Attachments []map[string]interface{}
	ThreadTS    string
	UnfurlLinks bool
	UnfurlMedia bool
}

// SlackResponse represents the response from sending a Slack message.
type SlackResponse struct {
	MessageID    string
	Status       string
	StatusCode   int
	ErrorMessage string
	SentAt       time.Time
}

// WhatsAppProvider defines the interface for WhatsApp notification delivery.
type WhatsAppProvider interface {
	// SendWhatsAppMessage sends a WhatsApp notification.
	SendWhatsAppMessage(ctx context.Context, request WhatsAppRequest) (*WhatsAppResponse, error)
	// GetProviderName returns the provider name.
	GetProviderName() string
	// IsAvailable checks if the provider is available.
	IsAvailable(ctx context.Context) bool
}

// WhatsAppRequest represents a request to send a WhatsApp message.
type WhatsAppRequest struct {
	MessageID    string
	From         string
	To           string
	TemplateID   string
	TemplateLang string
	Components   []WhatsAppComponent
	Text         string
	MediaURL     string
	MediaType    string
	Caption      string
}

// WhatsAppComponent represents a WhatsApp template component.
type WhatsAppComponent struct {
	Type       string
	SubType    string
	Index      int
	Parameters []WhatsAppParameter
}

// WhatsAppParameter represents a WhatsApp template parameter.
type WhatsAppParameter struct {
	Type     string
	Text     string
	Currency *WhatsAppCurrency
	DateTime *WhatsAppDateTime
	Image    *WhatsAppMedia
	Document *WhatsAppMedia
	Video    *WhatsAppMedia
}

// WhatsAppCurrency represents a WhatsApp currency parameter.
type WhatsAppCurrency struct {
	FallbackValue string
	Code          string
	Amount1000    int64
}

// WhatsAppDateTime represents a WhatsApp datetime parameter.
type WhatsAppDateTime struct {
	FallbackValue string
}

// WhatsAppMedia represents a WhatsApp media parameter.
type WhatsAppMedia struct {
	Link     string
	Caption  string
	Filename string
}

// WhatsAppResponse represents the response from sending a WhatsApp message.
type WhatsAppResponse struct {
	MessageID    string
	ProviderID   string
	Status       string
	StatusCode   int
	ErrorMessage string
	SentAt       time.Time
}

// TelegramProvider defines the interface for Telegram notification delivery.
type TelegramProvider interface {
	// SendTelegramMessage sends a Telegram notification.
	SendTelegramMessage(ctx context.Context, request TelegramRequest) (*TelegramResponse, error)
	// GetProviderName returns the provider name.
	GetProviderName() string
	// IsAvailable checks if the provider is available.
	IsAvailable(ctx context.Context) bool
}

// TelegramRequest represents a request to send a Telegram message.
type TelegramRequest struct {
	MessageID          string
	ChatID             string
	Text               string
	ParseMode          string // HTML, Markdown, MarkdownV2
	DisablePreview     bool
	DisableNotification bool
	ProtectContent     bool
	ReplyToMessageID   int64
	ReplyMarkup        interface{}
}

// TelegramResponse represents the response from sending a Telegram message.
type TelegramResponse struct {
	MessageID    string
	ProviderID   int64
	Status       string
	StatusCode   int
	ErrorMessage string
	SentAt       time.Time
}

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	// Publish publishes a domain event.
	Publish(ctx context.Context, event domain.DomainEvent) error
	// PublishBatch publishes multiple domain events.
	PublishBatch(ctx context.Context, events []domain.DomainEvent) error
}

// CacheService defines the interface for caching.
type CacheService interface {
	// Get retrieves a value from the cache.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores a value in the cache with an optional TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes a value from the cache.
	Delete(ctx context.Context, key string) error
	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)
	// GetOrSet gets a value from cache or sets it using the loader function.
	GetOrSet(ctx context.Context, key string, loader func() ([]byte, error), ttl time.Duration) ([]byte, error)
	// Invalidate invalidates cache entries by pattern.
	Invalidate(ctx context.Context, pattern string) error
}

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	// Allow checks if the action is allowed under rate limits.
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	// GetRemaining returns the remaining quota.
	GetRemaining(ctx context.Context, key string) (int, error)
	// Reset resets the rate limit for a key.
	Reset(ctx context.Context, key string) error
}

// QuotaManager defines the interface for quota management.
type QuotaManager interface {
	// CheckQuota checks if the action is within quota.
	CheckQuota(ctx context.Context, tenantID, channel string) (bool, error)
	// ConsumeQuota consumes quota for an action.
	ConsumeQuota(ctx context.Context, tenantID, channel string, amount int) error
	// GetUsage returns the current quota usage.
	GetUsage(ctx context.Context, tenantID, channel string) (*QuotaUsage, error)
	// ResetQuota resets the quota for a tenant/channel.
	ResetQuota(ctx context.Context, tenantID, channel string) error
}

// QuotaUsage represents quota usage information.
type QuotaUsage struct {
	TenantID     string
	Channel      string
	Used         int
	Limit        int
	ResetAt      time.Time
}

// UserService defines the interface for user-related operations.
type UserService interface {
	// GetUser retrieves a user by ID.
	GetUser(ctx context.Context, userID string) (*UserInfo, error)
	// GetUserByEmail retrieves a user by email.
	GetUserByEmail(ctx context.Context, email string) (*UserInfo, error)
	// GetUsersByIDs retrieves multiple users by IDs.
	GetUsersByIDs(ctx context.Context, userIDs []string) ([]*UserInfo, error)
	// IsUserActive checks if a user is active.
	IsUserActive(ctx context.Context, userID string) (bool, error)
}

// UserInfo represents user information.
type UserInfo struct {
	ID           string
	TenantID     string
	Email        string
	Phone        string
	FirstName    string
	LastName     string
	DisplayName  string
	Timezone     string
	Locale       string
	IsActive     bool
	DeviceTokens []string
	Metadata     map[string]interface{}
}

// CustomerService defines the interface for customer-related operations.
type CustomerService interface {
	// GetCustomer retrieves a customer by ID.
	GetCustomer(ctx context.Context, customerID string) (*CustomerInfo, error)
	// GetCustomerByEmail retrieves a customer by email.
	GetCustomerByEmail(ctx context.Context, tenantID, email string) (*CustomerInfo, error)
}

// CustomerInfo represents customer information.
type CustomerInfo struct {
	ID          string
	TenantID    string
	Email       string
	Phone       string
	FirstName   string
	LastName    string
	CompanyName string
	Timezone    string
	Locale      string
	IsActive    bool
	Tags        []string
	Metadata    map[string]interface{}
}

// TemplateRenderer defines the interface for template rendering.
type TemplateRenderer interface {
	// Render renders a template with the given variables.
	Render(ctx context.Context, template string, variables map[string]interface{}) (string, error)
	// RenderTemplate renders a template entity with the given variables.
	RenderTemplate(ctx context.Context, template *domain.NotificationTemplate, variables map[string]interface{}, locale string) (*RenderedTemplate, error)
	// Validate validates a template syntax.
	Validate(ctx context.Context, template string) error
}

// RenderedTemplate represents a rendered template result.
type RenderedTemplate struct {
	Subject  string
	Body     string
	HTMLBody string
	TextBody string
}

// Scheduler defines the interface for notification scheduling.
type Scheduler interface {
	// Schedule schedules a notification for future delivery.
	Schedule(ctx context.Context, notificationID string, scheduledAt time.Time) error
	// Cancel cancels a scheduled notification.
	Cancel(ctx context.Context, notificationID string) error
	// Reschedule reschedules a notification.
	Reschedule(ctx context.Context, notificationID string, scheduledAt time.Time) error
	// GetScheduled returns all scheduled notifications due for delivery.
	GetScheduled(ctx context.Context, before time.Time, limit int) ([]string, error)
}

// MetricsCollector defines the interface for metrics collection.
type MetricsCollector interface {
	// IncrementCounter increments a counter metric.
	IncrementCounter(ctx context.Context, name string, tags map[string]string)
	// RecordHistogram records a histogram metric.
	RecordHistogram(ctx context.Context, name string, value float64, tags map[string]string)
	// RecordGauge records a gauge metric.
	RecordGauge(ctx context.Context, name string, value float64, tags map[string]string)
	// RecordDuration records a duration metric.
	RecordDuration(ctx context.Context, name string, duration time.Duration, tags map[string]string)
}

// Logger defines the interface for logging.
type Logger interface {
	// Debug logs a debug message.
	Debug(msg string, fields map[string]interface{})
	// Info logs an info message.
	Info(msg string, fields map[string]interface{})
	// Warn logs a warning message.
	Warn(msg string, fields map[string]interface{})
	// Error logs an error message.
	Error(msg string, err error, fields map[string]interface{})
	// WithContext returns a logger with context fields.
	WithContext(ctx context.Context) Logger
	// WithFields returns a logger with additional fields.
	WithFields(fields map[string]interface{}) Logger
}

// SuppressionService defines the interface for suppression list management.
type SuppressionService interface {
	// IsSuppressed checks if a recipient is suppressed.
	IsSuppressed(ctx context.Context, tenantID, channel, recipient string) (bool, error)
	// AddSuppression adds a recipient to the suppression list.
	AddSuppression(ctx context.Context, tenantID, channel, recipient, reason string) error
	// RemoveSuppression removes a recipient from the suppression list.
	RemoveSuppression(ctx context.Context, tenantID, channel, recipient string) error
	// GetSuppressionReason gets the suppression reason for a recipient.
	GetSuppressionReason(ctx context.Context, tenantID, channel, recipient string) (string, error)
}

// DeliveryTracker defines the interface for tracking notification delivery.
type DeliveryTracker interface {
	// TrackSent tracks when a notification is sent.
	TrackSent(ctx context.Context, notificationID, channel, provider string) error
	// TrackDelivered tracks when a notification is delivered.
	TrackDelivered(ctx context.Context, notificationID string, deliveredAt time.Time) error
	// TrackFailed tracks when a notification fails.
	TrackFailed(ctx context.Context, notificationID, errorCode, errorMessage string) error
	// TrackOpened tracks when a notification is opened.
	TrackOpened(ctx context.Context, notificationID string, openedAt time.Time) error
	// TrackClicked tracks when a notification link is clicked.
	TrackClicked(ctx context.Context, notificationID, linkID string, clickedAt time.Time) error
	// GetDeliveryStatus gets the delivery status for a notification.
	GetDeliveryStatus(ctx context.Context, notificationID string) (*DeliveryStatus, error)
}

// DeliveryStatus represents the delivery status of a notification.
type DeliveryStatus struct {
	NotificationID string
	Channel        string
	Provider       string
	Status         string
	SentAt         *time.Time
	DeliveredAt    *time.Time
	OpenedAt       *time.Time
	ClickedAt      *time.Time
	FailedAt       *time.Time
	ErrorCode      string
	ErrorMessage   string
	RetryCount     int
}

// TransactionManager defines the interface for transaction management.
type TransactionManager interface {
	// Begin starts a new transaction.
	Begin(ctx context.Context) (Transaction, error)
	// RunInTransaction runs a function within a transaction.
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Transaction represents a database transaction.
type Transaction interface {
	// Commit commits the transaction.
	Commit() error
	// Rollback rolls back the transaction.
	Rollback() error
	// Context returns the transaction context.
	Context() context.Context
}

// IdGenerator defines the interface for ID generation.
type IdGenerator interface {
	// Generate generates a new unique ID.
	Generate() string
	// GenerateWithPrefix generates a new unique ID with a prefix.
	GenerateWithPrefix(prefix string) string
}

// TimeProvider defines the interface for time operations.
type TimeProvider interface {
	// Now returns the current time.
	Now() time.Time
	// NowUTC returns the current UTC time.
	NowUTC() time.Time
}

// EncryptionService defines the interface for encryption operations.
type EncryptionService interface {
	// Encrypt encrypts data.
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	// Decrypt decrypts data.
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
	// Hash hashes data.
	Hash(data []byte) string
}
