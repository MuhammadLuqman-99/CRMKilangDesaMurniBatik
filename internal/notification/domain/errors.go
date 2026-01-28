// Package domain contains the domain layer for the Notification service.
package domain

import (
	"errors"
	"fmt"
)

// Domain errors for the Notification service.
var (
	// General errors
	ErrNotificationNotFound      = errors.New("notification not found")
	ErrTemplateNotFound          = errors.New("notification template not found")
	ErrPreferenceNotFound        = errors.New("notification preference not found")
	ErrInvalidNotification       = errors.New("invalid notification")
	ErrInvalidTemplate           = errors.New("invalid notification template")
	ErrInvalidChannel            = errors.New("invalid notification channel")
	ErrInvalidPriority           = errors.New("invalid notification priority")
	ErrInvalidStatus             = errors.New("invalid notification status")
	ErrInvalidRecipient          = errors.New("invalid recipient")
	ErrInvalidContent            = errors.New("invalid notification content")

	// Delivery errors
	ErrDeliveryFailed            = errors.New("notification delivery failed")
	ErrChannelNotConfigured      = errors.New("notification channel not configured")
	ErrChannelDisabled           = errors.New("notification channel disabled")
	ErrRecipientOptedOut         = errors.New("recipient has opted out of notifications")
	ErrRateLimitExceeded         = errors.New("notification rate limit exceeded")
	ErrQuotaExceeded             = errors.New("notification quota exceeded")
	ErrProviderError             = errors.New("notification provider error")
	ErrProviderUnavailable       = errors.New("notification provider unavailable")
	ErrInvalidProviderConfig     = errors.New("invalid provider configuration")

	// Template errors
	ErrTemplateNameRequired      = errors.New("template name is required")
	ErrTemplateNameTooLong       = errors.New("template name too long")
	ErrTemplateAlreadyExists     = errors.New("template with this name already exists")
	ErrTemplateInUse             = errors.New("template is in use and cannot be deleted")
	ErrInvalidTemplateContent    = errors.New("invalid template content")
	ErrTemplateRenderFailed      = errors.New("template rendering failed")
	ErrMissingTemplateVariable   = errors.New("missing required template variable")
	ErrInvalidTemplateVariable   = errors.New("invalid template variable")

	// Email specific errors
	ErrInvalidEmailAddress       = errors.New("invalid email address")
	ErrEmailAddressRequired      = errors.New("email address is required")
	ErrInvalidEmailSubject       = errors.New("invalid email subject")
	ErrEmailSubjectTooLong       = errors.New("email subject too long")
	ErrInvalidEmailBody          = errors.New("invalid email body")
	ErrEmailBodyTooLarge         = errors.New("email body too large")
	ErrInvalidEmailAttachment    = errors.New("invalid email attachment")
	ErrAttachmentTooLarge        = errors.New("email attachment too large")
	ErrTooManyAttachments        = errors.New("too many email attachments")
	ErrInvalidFromAddress        = errors.New("invalid from email address")
	ErrInvalidReplyToAddress     = errors.New("invalid reply-to email address")

	// SMS specific errors
	ErrInvalidPhoneNumber        = errors.New("invalid phone number")
	ErrPhoneNumberRequired       = errors.New("phone number is required")
	ErrSMSBodyTooLong            = errors.New("SMS body too long")
	ErrSMSBodyRequired           = errors.New("SMS body is required")
	ErrInvalidSenderID           = errors.New("invalid SMS sender ID")
	ErrSMSNotSupported           = errors.New("SMS not supported for this country")

	// Push notification errors
	ErrInvalidDeviceToken        = errors.New("invalid device token")
	ErrDeviceTokenRequired       = errors.New("device token is required")
	ErrDeviceNotRegistered       = errors.New("device not registered for push notifications")
	ErrPushNotificationTooLarge  = errors.New("push notification payload too large")
	ErrInvalidPushPayload        = errors.New("invalid push notification payload")

	// In-app notification errors
	ErrInvalidUserID             = errors.New("invalid user ID")
	ErrUserIDRequired            = errors.New("user ID is required")
	ErrInAppNotificationTooLong  = errors.New("in-app notification too long")

	// Webhook errors
	ErrInvalidWebhookURL         = errors.New("invalid webhook URL")
	ErrWebhookURLRequired        = errors.New("webhook URL is required")
	ErrWebhookRequestFailed      = errors.New("webhook request failed")
	ErrWebhookTimeout            = errors.New("webhook request timeout")

	// Scheduling errors
	ErrInvalidScheduledTime      = errors.New("invalid scheduled time")
	ErrScheduledTimeInPast       = errors.New("scheduled time is in the past")
	ErrAlreadySent               = errors.New("notification has already been sent")
	ErrAlreadyCancelled          = errors.New("notification has already been cancelled")
	ErrCannotCancel              = errors.New("notification cannot be cancelled")
	ErrCannotRetry               = errors.New("notification cannot be retried")
	ErrMaxRetriesExceeded        = errors.New("maximum retry attempts exceeded")

	// Batch errors
	ErrBatchNotFound             = errors.New("notification batch not found")
	ErrBatchTooLarge             = errors.New("notification batch too large")
	ErrBatchEmpty                = errors.New("notification batch is empty")
	ErrBatchInProgress           = errors.New("notification batch is already in progress")

	// Tenant errors
	ErrTenantIDRequired          = errors.New("tenant ID is required")
	ErrTenantNotConfigured       = errors.New("tenant notification settings not configured")
	ErrTenantChannelDisabled     = errors.New("notification channel disabled for tenant")

	// Concurrency errors
	ErrConcurrentModification    = errors.New("concurrent modification detected")
	ErrVersionMismatch           = errors.New("version mismatch")
)

// ValidationError represents a validation error with field information.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message, code string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	}
}

// ValidationErrors represents a collection of validation errors.
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface.
func (e ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", e.Errors[0].Message)
}

// HasErrors returns true if there are any errors.
func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// AddField adds a field validation error.
func (e *ValidationErrors) AddField(field, message, code string) {
	e.Errors = append(e.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// Add adds a validation error.
func (e *ValidationErrors) Add(err ValidationError) {
	e.Errors = append(e.Errors, err)
}

// DomainError represents a domain-specific error with additional context.
type DomainError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Inner      error                  `json:"-"`
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Inner)
	}
	return e.Message
}

// Unwrap returns the inner error.
func (e *DomainError) Unwrap() error {
	return e.Inner
}

// NewDomainError creates a new domain error.
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the error.
func (e *DomainError) WithDetails(details map[string]interface{}) *DomainError {
	e.Details = details
	return e
}

// WithInner wraps an inner error.
func (e *DomainError) WithInner(inner error) *DomainError {
	e.Inner = inner
	return e
}

// DeliveryError represents a notification delivery error with retry information.
type DeliveryError struct {
	NotificationID string `json:"notification_id"`
	Channel        string `json:"channel"`
	Provider       string `json:"provider"`
	Message        string `json:"message"`
	Code           string `json:"code"`
	Retryable      bool   `json:"retryable"`
	RetryAfter     int    `json:"retry_after_seconds,omitempty"`
	Inner          error  `json:"-"`
}

// Error implements the error interface.
func (e *DeliveryError) Error() string {
	return fmt.Sprintf("delivery failed for notification %s via %s: %s", e.NotificationID, e.Channel, e.Message)
}

// Unwrap returns the inner error.
func (e *DeliveryError) Unwrap() error {
	return e.Inner
}

// NewDeliveryError creates a new delivery error.
func NewDeliveryError(notificationID, channel, provider, message, code string, retryable bool) *DeliveryError {
	return &DeliveryError{
		NotificationID: notificationID,
		Channel:        channel,
		Provider:       provider,
		Message:        message,
		Code:           code,
		Retryable:      retryable,
	}
}

// WithRetryAfter sets the retry delay in seconds.
func (e *DeliveryError) WithRetryAfter(seconds int) *DeliveryError {
	e.RetryAfter = seconds
	return e
}

// WithInner wraps an inner error.
func (e *DeliveryError) WithInner(inner error) *DeliveryError {
	e.Inner = inner
	return e
}

// TemplateError represents a template-related error.
type TemplateError struct {
	TemplateName string   `json:"template_name"`
	Message      string   `json:"message"`
	Code         string   `json:"code"`
	MissingVars  []string `json:"missing_variables,omitempty"`
	InvalidVars  []string `json:"invalid_variables,omitempty"`
	Inner        error    `json:"-"`
}

// Error implements the error interface.
func (e *TemplateError) Error() string {
	return fmt.Sprintf("template error '%s': %s", e.TemplateName, e.Message)
}

// Unwrap returns the inner error.
func (e *TemplateError) Unwrap() error {
	return e.Inner
}

// NewTemplateError creates a new template error.
func NewTemplateError(templateName, message, code string) *TemplateError {
	return &TemplateError{
		TemplateName: templateName,
		Message:      message,
		Code:         code,
	}
}

// WithMissingVars adds missing variable names.
func (e *TemplateError) WithMissingVars(vars []string) *TemplateError {
	e.MissingVars = vars
	return e
}

// WithInvalidVars adds invalid variable names.
func (e *TemplateError) WithInvalidVars(vars []string) *TemplateError {
	e.InvalidVars = vars
	return e
}
