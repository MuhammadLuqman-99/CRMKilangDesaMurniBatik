// Package domain contains the domain layer for the Notification service.
package domain

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ============================================================================
// NotificationChannel Value Object
// ============================================================================

// NotificationChannel represents a communication channel for notifications.
type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelSMS      NotificationChannel = "sms"
	ChannelPush     NotificationChannel = "push"
	ChannelInApp    NotificationChannel = "in_app"
	ChannelWebhook  NotificationChannel = "webhook"
	ChannelSlack    NotificationChannel = "slack"
	ChannelWhatsApp NotificationChannel = "whatsapp"
	ChannelTelegram NotificationChannel = "telegram"
)

// ValidChannels is a map of valid notification channels.
var ValidChannels = map[NotificationChannel]bool{
	ChannelEmail:    true,
	ChannelSMS:      true,
	ChannelPush:     true,
	ChannelInApp:    true,
	ChannelWebhook:  true,
	ChannelSlack:    true,
	ChannelWhatsApp: true,
	ChannelTelegram: true,
}

// IsValid checks if the channel is valid.
func (c NotificationChannel) IsValid() bool {
	return ValidChannels[c]
}

// String returns the string representation.
func (c NotificationChannel) String() string {
	return string(c)
}

// DisplayName returns a human-readable name for the channel.
func (c NotificationChannel) DisplayName() string {
	names := map[NotificationChannel]string{
		ChannelEmail:    "Email",
		ChannelSMS:      "SMS",
		ChannelPush:     "Push Notification",
		ChannelInApp:    "In-App Notification",
		ChannelWebhook:  "Webhook",
		ChannelSlack:    "Slack",
		ChannelWhatsApp: "WhatsApp",
		ChannelTelegram: "Telegram",
	}
	if name, ok := names[c]; ok {
		return name
	}
	return string(c)
}

// RequiresDeviceToken returns true if channel requires device token.
func (c NotificationChannel) RequiresDeviceToken() bool {
	return c == ChannelPush
}

// RequiresPhoneNumber returns true if channel requires phone number.
func (c NotificationChannel) RequiresPhoneNumber() bool {
	return c == ChannelSMS || c == ChannelWhatsApp
}

// RequiresEmail returns true if channel requires email address.
func (c NotificationChannel) RequiresEmail() bool {
	return c == ChannelEmail
}

// RequiresUserID returns true if channel requires user ID.
func (c NotificationChannel) RequiresUserID() bool {
	return c == ChannelInApp
}

// AllChannels returns all valid notification channels.
func AllChannels() []NotificationChannel {
	return []NotificationChannel{
		ChannelEmail,
		ChannelSMS,
		ChannelPush,
		ChannelInApp,
		ChannelWebhook,
		ChannelSlack,
		ChannelWhatsApp,
		ChannelTelegram,
	}
}

// ParseChannel parses a string into a NotificationChannel.
func ParseChannel(s string) (NotificationChannel, error) {
	channel := NotificationChannel(strings.ToLower(strings.TrimSpace(s)))
	if !channel.IsValid() {
		return "", ErrInvalidChannel
	}
	return channel, nil
}

// ============================================================================
// NotificationPriority Value Object
// ============================================================================

// NotificationPriority represents the priority level of a notification.
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityNormal   NotificationPriority = "normal"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// ValidPriorities is a map of valid priorities.
var ValidPriorities = map[NotificationPriority]bool{
	PriorityLow:      true,
	PriorityNormal:   true,
	PriorityHigh:     true,
	PriorityCritical: true,
}

// IsValid checks if the priority is valid.
func (p NotificationPriority) IsValid() bool {
	return ValidPriorities[p]
}

// String returns the string representation.
func (p NotificationPriority) String() string {
	return string(p)
}

// Weight returns a numeric weight for sorting/comparison.
func (p NotificationPriority) Weight() int {
	weights := map[NotificationPriority]int{
		PriorityLow:      1,
		PriorityNormal:   2,
		PriorityHigh:     3,
		PriorityCritical: 4,
	}
	return weights[p]
}

// IsHigherThan checks if this priority is higher than another.
func (p NotificationPriority) IsHigherThan(other NotificationPriority) bool {
	return p.Weight() > other.Weight()
}

// ParsePriority parses a string into a NotificationPriority.
func ParsePriority(s string) (NotificationPriority, error) {
	priority := NotificationPriority(strings.ToLower(strings.TrimSpace(s)))
	if !priority.IsValid() {
		return "", ErrInvalidPriority
	}
	return priority, nil
}

// ============================================================================
// NotificationStatus Value Object
// ============================================================================

// NotificationStatus represents the delivery status of a notification.
type NotificationStatus string

const (
	StatusPending    NotificationStatus = "pending"
	StatusQueued     NotificationStatus = "queued"
	StatusSending    NotificationStatus = "sending"
	StatusSent       NotificationStatus = "sent"
	StatusDelivered  NotificationStatus = "delivered"
	StatusRead       NotificationStatus = "read"
	StatusFailed     NotificationStatus = "failed"
	StatusCancelled  NotificationStatus = "cancelled"
	StatusScheduled  NotificationStatus = "scheduled"
	StatusRetrying   NotificationStatus = "retrying"
	StatusBounced    NotificationStatus = "bounced"
	StatusComplained NotificationStatus = "complained"
)

// ValidStatuses is a map of valid statuses.
var ValidStatuses = map[NotificationStatus]bool{
	StatusPending:    true,
	StatusQueued:     true,
	StatusSending:    true,
	StatusSent:       true,
	StatusDelivered:  true,
	StatusRead:       true,
	StatusFailed:     true,
	StatusCancelled:  true,
	StatusScheduled:  true,
	StatusRetrying:   true,
	StatusBounced:    true,
	StatusComplained: true,
}

// IsValid checks if the status is valid.
func (s NotificationStatus) IsValid() bool {
	return ValidStatuses[s]
}

// String returns the string representation.
func (s NotificationStatus) String() string {
	return string(s)
}

// IsFinal returns true if the status is a final state.
func (s NotificationStatus) IsFinal() bool {
	finalStatuses := map[NotificationStatus]bool{
		StatusDelivered:  true,
		StatusRead:       true,
		StatusFailed:     true,
		StatusCancelled:  true,
		StatusBounced:    true,
		StatusComplained: true,
	}
	return finalStatuses[s]
}

// IsSuccess returns true if the status indicates successful delivery.
func (s NotificationStatus) IsSuccess() bool {
	successStatuses := map[NotificationStatus]bool{
		StatusSent:      true,
		StatusDelivered: true,
		StatusRead:      true,
	}
	return successStatuses[s]
}

// CanTransitionTo checks if this status can transition to another.
func (s NotificationStatus) CanTransitionTo(target NotificationStatus) bool {
	transitions := map[NotificationStatus][]NotificationStatus{
		StatusPending:   {StatusQueued, StatusScheduled, StatusCancelled},
		StatusScheduled: {StatusQueued, StatusCancelled},
		StatusQueued:    {StatusSending, StatusCancelled},
		StatusSending:   {StatusSent, StatusFailed, StatusRetrying},
		StatusRetrying:  {StatusSending, StatusFailed},
		StatusSent:      {StatusDelivered, StatusBounced, StatusComplained},
		StatusDelivered: {StatusRead},
	}

	allowed, ok := transitions[s]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == target {
			return true
		}
	}
	return false
}

// ParseStatus parses a string into a NotificationStatus.
func ParseStatus(s string) (NotificationStatus, error) {
	status := NotificationStatus(strings.ToLower(strings.TrimSpace(s)))
	if !status.IsValid() {
		return "", ErrInvalidStatus
	}
	return status, nil
}

// ============================================================================
// NotificationType Value Object
// ============================================================================

// NotificationType represents the type/category of notification.
type NotificationType string

const (
	TypeSystem       NotificationType = "system"
	TypeMarketing    NotificationType = "marketing"
	TypeTransactional NotificationType = "transactional"
	TypeAlert        NotificationType = "alert"
	TypeReminder     NotificationType = "reminder"
	TypeWelcome      NotificationType = "welcome"
	TypeVerification NotificationType = "verification"
	TypePasswordReset NotificationType = "password_reset"
	TypeInvoice      NotificationType = "invoice"
	TypePromotion    NotificationType = "promotion"
	TypeUpdate       NotificationType = "update"
	TypeComment      NotificationType = "comment"
	TypeMention      NotificationType = "mention"
	TypeAssignment   NotificationType = "assignment"
	TypeDeadline     NotificationType = "deadline"
)

// ValidTypes is a map of valid notification types.
var ValidTypes = map[NotificationType]bool{
	TypeSystem:        true,
	TypeMarketing:     true,
	TypeTransactional: true,
	TypeAlert:         true,
	TypeReminder:      true,
	TypeWelcome:       true,
	TypeVerification:  true,
	TypePasswordReset: true,
	TypeInvoice:       true,
	TypePromotion:     true,
	TypeUpdate:        true,
	TypeComment:       true,
	TypeMention:       true,
	TypeAssignment:    true,
	TypeDeadline:      true,
}

// IsValid checks if the type is valid.
func (t NotificationType) IsValid() bool {
	return ValidTypes[t]
}

// String returns the string representation.
func (t NotificationType) String() string {
	return string(t)
}

// RequiresOptIn returns true if this type requires explicit user opt-in.
func (t NotificationType) RequiresOptIn() bool {
	optInTypes := map[NotificationType]bool{
		TypeMarketing:  true,
		TypePromotion:  true,
	}
	return optInTypes[t]
}

// CanUnsubscribe returns true if users can unsubscribe from this type.
func (t NotificationType) CanUnsubscribe() bool {
	// Users cannot unsubscribe from system/security notifications
	cannotUnsubscribe := map[NotificationType]bool{
		TypeSystem:        true,
		TypeVerification:  true,
		TypePasswordReset: true,
	}
	return !cannotUnsubscribe[t]
}

// ParseType parses a string into a NotificationType.
func ParseType(s string) (NotificationType, error) {
	t := NotificationType(strings.ToLower(strings.TrimSpace(s)))
	if !t.IsValid() {
		return "", NewValidationError("type", "invalid notification type", "INVALID_TYPE")
	}
	return t, nil
}

// ============================================================================
// EmailAddress Value Object
// ============================================================================

// EmailAddress represents a validated email address.
type EmailAddress struct {
	address    string
	local      string
	domain     string
	normalized string
}

// Email regex pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// NewEmailAddress creates a new EmailAddress value object.
func NewEmailAddress(address string) (EmailAddress, error) {
	address = strings.TrimSpace(address)

	if address == "" {
		return EmailAddress{}, ErrEmailAddressRequired
	}

	if len(address) > 254 {
		return EmailAddress{}, NewValidationError("email", "email address too long (max 254 characters)", "EMAIL_TOO_LONG")
	}

	if !emailRegex.MatchString(address) {
		return EmailAddress{}, ErrInvalidEmailAddress
	}

	parts := strings.SplitN(address, "@", 2)
	if len(parts) != 2 {
		return EmailAddress{}, ErrInvalidEmailAddress
	}

	local := parts[0]
	domain := strings.ToLower(parts[1])

	if len(local) > 64 {
		return EmailAddress{}, NewValidationError("email", "local part too long (max 64 characters)", "EMAIL_LOCAL_TOO_LONG")
	}

	normalized := strings.ToLower(address)

	return EmailAddress{
		address:    address,
		local:      local,
		domain:     domain,
		normalized: normalized,
	}, nil
}

// String returns the original email address.
func (e EmailAddress) String() string {
	return e.address
}

// Normalized returns the normalized (lowercase) email.
func (e EmailAddress) Normalized() string {
	return e.normalized
}

// Local returns the local part (before @).
func (e EmailAddress) Local() string {
	return e.local
}

// Domain returns the domain part (after @).
func (e EmailAddress) Domain() string {
	return e.domain
}

// IsEmpty returns true if the email is empty.
func (e EmailAddress) IsEmpty() bool {
	return e.address == ""
}

// Equals checks if two emails are equal (case-insensitive).
func (e EmailAddress) Equals(other EmailAddress) bool {
	return e.normalized == other.normalized
}

// MarshalJSON implements json.Marshaler.
func (e EmailAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.address)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *EmailAddress) UnmarshalJSON(data []byte) error {
	var address string
	if err := json.Unmarshal(data, &address); err != nil {
		return err
	}
	email, err := NewEmailAddress(address)
	if err != nil {
		return err
	}
	*e = email
	return nil
}

// ============================================================================
// PhoneNumber Value Object
// ============================================================================

// PhoneNumber represents a validated phone number.
type PhoneNumber struct {
	raw         string
	countryCode string
	number      string
	e164        string
}

// Phone number regex patterns
var (
	digitsOnlyRegex = regexp.MustCompile(`[^\d+]`)
	phoneRegex      = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

// NewPhoneNumber creates a new PhoneNumber value object.
func NewPhoneNumber(raw string) (PhoneNumber, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return PhoneNumber{}, ErrPhoneNumberRequired
	}

	// Clean the number
	hasPlus := strings.HasPrefix(raw, "+")
	cleaned := digitsOnlyRegex.ReplaceAllString(raw, "")

	if hasPlus {
		cleaned = "+" + cleaned
	}

	// Validate E.164 format
	if !phoneRegex.MatchString(cleaned) {
		return PhoneNumber{}, ErrInvalidPhoneNumber
	}

	e164 := cleaned
	if !hasPlus {
		e164 = "+" + cleaned
	}

	return PhoneNumber{
		raw:    raw,
		number: strings.TrimPrefix(e164, "+"),
		e164:   e164,
	}, nil
}

// String returns the E.164 formatted number.
func (p PhoneNumber) String() string {
	return p.e164
}

// E164 returns the E.164 formatted number.
func (p PhoneNumber) E164() string {
	return p.e164
}

// Raw returns the original input.
func (p PhoneNumber) Raw() string {
	return p.raw
}

// IsEmpty returns true if the phone number is empty.
func (p PhoneNumber) IsEmpty() bool {
	return p.e164 == ""
}

// Equals checks if two phone numbers are equal.
func (p PhoneNumber) Equals(other PhoneNumber) bool {
	return p.e164 == other.e164
}

// MarshalJSON implements json.Marshaler.
func (p PhoneNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.e164)
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PhoneNumber) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	phone, err := NewPhoneNumber(raw)
	if err != nil {
		return err
	}
	*p = phone
	return nil
}

// ============================================================================
// DeviceToken Value Object
// ============================================================================

// DeviceToken represents a push notification device token.
type DeviceToken struct {
	token    string
	platform DevicePlatform
}

// DevicePlatform represents a mobile platform.
type DevicePlatform string

const (
	PlatformIOS     DevicePlatform = "ios"
	PlatformAndroid DevicePlatform = "android"
	PlatformWeb     DevicePlatform = "web"
)

// ValidPlatforms is a map of valid platforms.
var ValidPlatforms = map[DevicePlatform]bool{
	PlatformIOS:     true,
	PlatformAndroid: true,
	PlatformWeb:     true,
}

// NewDeviceToken creates a new DeviceToken value object.
func NewDeviceToken(token string, platform DevicePlatform) (DeviceToken, error) {
	token = strings.TrimSpace(token)

	if token == "" {
		return DeviceToken{}, ErrDeviceTokenRequired
	}

	if !ValidPlatforms[platform] {
		return DeviceToken{}, NewValidationError("platform", "invalid device platform", "INVALID_PLATFORM")
	}

	// Basic validation based on platform
	if platform == PlatformIOS && len(token) != 64 {
		return DeviceToken{}, NewValidationError("token", "invalid iOS device token length", "INVALID_IOS_TOKEN")
	}

	return DeviceToken{
		token:    token,
		platform: platform,
	}, nil
}

// String returns the token.
func (d DeviceToken) String() string {
	return d.token
}

// Token returns the token value.
func (d DeviceToken) Token() string {
	return d.token
}

// Platform returns the device platform.
func (d DeviceToken) Platform() DevicePlatform {
	return d.platform
}

// IsEmpty returns true if the token is empty.
func (d DeviceToken) IsEmpty() bool {
	return d.token == ""
}

// Equals checks if two device tokens are equal.
func (d DeviceToken) Equals(other DeviceToken) bool {
	return d.token == other.token && d.platform == other.platform
}

// ============================================================================
// Recipient Value Object
// ============================================================================

// Recipient represents a notification recipient with channel-specific addresses.
type Recipient struct {
	UserID      string       `json:"user_id,omitempty"`
	Email       string       `json:"email,omitempty"`
	Phone       string       `json:"phone,omitempty"`
	DeviceToken string       `json:"device_token,omitempty"`
	Platform    string       `json:"platform,omitempty"`
	WebhookURL  string       `json:"webhook_url,omitempty"`
	SlackUserID string       `json:"slack_user_id,omitempty"`
	TelegramID  string       `json:"telegram_id,omitempty"`
	Name        string       `json:"name,omitempty"`
	Locale      string       `json:"locale,omitempty"`
	Timezone    string       `json:"timezone,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewRecipient creates a new Recipient.
func NewRecipient() *Recipient {
	return &Recipient{
		Metadata: make(map[string]interface{}),
	}
}

// WithUserID sets the user ID.
func (r *Recipient) WithUserID(userID string) *Recipient {
	r.UserID = userID
	return r
}

// WithEmail sets the email address.
func (r *Recipient) WithEmail(email string) *Recipient {
	r.Email = email
	return r
}

// WithPhone sets the phone number.
func (r *Recipient) WithPhone(phone string) *Recipient {
	r.Phone = phone
	return r
}

// WithDeviceToken sets the device token for push notifications.
func (r *Recipient) WithDeviceToken(token, platform string) *Recipient {
	r.DeviceToken = token
	r.Platform = platform
	return r
}

// WithWebhookURL sets the webhook URL.
func (r *Recipient) WithWebhookURL(url string) *Recipient {
	r.WebhookURL = url
	return r
}

// WithName sets the recipient name.
func (r *Recipient) WithName(name string) *Recipient {
	r.Name = name
	return r
}

// WithLocale sets the recipient locale.
func (r *Recipient) WithLocale(locale string) *Recipient {
	r.Locale = locale
	return r
}

// WithTimezone sets the recipient timezone.
func (r *Recipient) WithTimezone(timezone string) *Recipient {
	r.Timezone = timezone
	return r
}

// WithMetadata sets a metadata value.
func (r *Recipient) WithMetadata(key string, value interface{}) *Recipient {
	if r.Metadata == nil {
		r.Metadata = make(map[string]interface{})
	}
	r.Metadata[key] = value
	return r
}

// ValidateForChannel validates the recipient has required fields for a channel.
func (r *Recipient) ValidateForChannel(channel NotificationChannel) error {
	switch channel {
	case ChannelEmail:
		if r.Email == "" {
			return ErrEmailAddressRequired
		}
		if _, err := NewEmailAddress(r.Email); err != nil {
			return err
		}
	case ChannelSMS, ChannelWhatsApp:
		if r.Phone == "" {
			return ErrPhoneNumberRequired
		}
		if _, err := NewPhoneNumber(r.Phone); err != nil {
			return err
		}
	case ChannelPush:
		if r.DeviceToken == "" {
			return ErrDeviceTokenRequired
		}
	case ChannelInApp:
		if r.UserID == "" {
			return ErrUserIDRequired
		}
	case ChannelWebhook:
		if r.WebhookURL == "" {
			return ErrWebhookURLRequired
		}
	}
	return nil
}

// IsEmpty returns true if recipient has no contact information.
func (r *Recipient) IsEmpty() bool {
	return r.UserID == "" && r.Email == "" && r.Phone == "" &&
		r.DeviceToken == "" && r.WebhookURL == "" && r.SlackUserID == "" && r.TelegramID == ""
}

// ============================================================================
// TemplateVariable Value Object
// ============================================================================

// TemplateVariable represents a variable that can be used in templates.
type TemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // string, number, boolean, date, array, object
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Description  string      `json:"description,omitempty"`
	Example      interface{} `json:"example,omitempty"`
}

// ValidVariableTypes contains valid variable types.
var ValidVariableTypes = map[string]bool{
	"string":  true,
	"number":  true,
	"boolean": true,
	"date":    true,
	"array":   true,
	"object":  true,
}

// NewTemplateVariable creates a new template variable.
func NewTemplateVariable(name, varType string, required bool) (*TemplateVariable, error) {
	name = strings.TrimSpace(name)
	varType = strings.ToLower(strings.TrimSpace(varType))

	if name == "" {
		return nil, NewValidationError("name", "variable name is required", "REQUIRED")
	}

	if !ValidVariableTypes[varType] {
		return nil, NewValidationError("type", "invalid variable type", "INVALID_TYPE")
	}

	return &TemplateVariable{
		Name:     name,
		Type:     varType,
		Required: required,
	}, nil
}

// WithDefaultValue sets the default value.
func (v *TemplateVariable) WithDefaultValue(value interface{}) *TemplateVariable {
	v.DefaultValue = value
	return v
}

// WithDescription sets the description.
func (v *TemplateVariable) WithDescription(desc string) *TemplateVariable {
	v.Description = desc
	return v
}

// WithExample sets an example value.
func (v *TemplateVariable) WithExample(example interface{}) *TemplateVariable {
	v.Example = example
	return v
}

// ============================================================================
// RetryPolicy Value Object
// ============================================================================

// RetryPolicy defines how notification delivery should be retried.
type RetryPolicy struct {
	MaxAttempts     int   `json:"max_attempts"`
	InitialInterval int   `json:"initial_interval_seconds"` // Initial retry interval in seconds
	MaxInterval     int   `json:"max_interval_seconds"`     // Maximum retry interval in seconds
	Multiplier      float64 `json:"multiplier"`             // Backoff multiplier
	RetryableErrors []string `json:"retryable_errors,omitempty"`
}

// DefaultRetryPolicy returns the default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:     3,
		InitialInterval: 60,   // 1 minute
		MaxInterval:     3600, // 1 hour
		Multiplier:      2.0,
	}
}

// NewRetryPolicy creates a new retry policy.
func NewRetryPolicy(maxAttempts, initialInterval, maxInterval int, multiplier float64) (*RetryPolicy, error) {
	if maxAttempts < 0 {
		return nil, NewValidationError("max_attempts", "max attempts cannot be negative", "INVALID_VALUE")
	}
	if initialInterval < 0 {
		return nil, NewValidationError("initial_interval", "initial interval cannot be negative", "INVALID_VALUE")
	}
	if maxInterval < initialInterval {
		return nil, NewValidationError("max_interval", "max interval must be >= initial interval", "INVALID_VALUE")
	}
	if multiplier < 1.0 {
		return nil, NewValidationError("multiplier", "multiplier must be >= 1.0", "INVALID_VALUE")
	}

	return &RetryPolicy{
		MaxAttempts:     maxAttempts,
		InitialInterval: initialInterval,
		MaxInterval:     maxInterval,
		Multiplier:      multiplier,
	}, nil
}

// GetInterval calculates the retry interval for a given attempt.
func (r RetryPolicy) GetInterval(attempt int) int {
	if attempt <= 0 {
		return r.InitialInterval
	}

	interval := float64(r.InitialInterval)
	for i := 1; i < attempt && i < r.MaxAttempts; i++ {
		interval *= r.Multiplier
	}

	if int(interval) > r.MaxInterval {
		return r.MaxInterval
	}
	return int(interval)
}

// ShouldRetry returns true if another retry attempt should be made.
func (r RetryPolicy) ShouldRetry(attempt int) bool {
	return attempt < r.MaxAttempts
}
