// Package application contains the application layer for the Notification service.
package application

import (
	"errors"
	"fmt"
)

// ErrorCode represents an application-level error code.
type ErrorCode string

// Application error codes for the Notification service.
const (
	// General errors
	ErrCodeInternal           ErrorCode = "NOTIFICATION_INTERNAL_ERROR"
	ErrCodeValidation         ErrorCode = "NOTIFICATION_VALIDATION_ERROR"
	ErrCodeNotFound           ErrorCode = "NOTIFICATION_NOT_FOUND"
	ErrCodeAlreadyExists      ErrorCode = "NOTIFICATION_ALREADY_EXISTS"
	ErrCodeUnauthorized       ErrorCode = "NOTIFICATION_UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "NOTIFICATION_FORBIDDEN"
	ErrCodeConflict           ErrorCode = "NOTIFICATION_CONFLICT"
	ErrCodeInvalidInput       ErrorCode = "NOTIFICATION_INVALID_INPUT"
	ErrCodeInvalidState       ErrorCode = "NOTIFICATION_INVALID_STATE"

	// Notification-specific errors
	ErrCodeNotificationNotFound    ErrorCode = "NOTIFICATION_NOT_FOUND"
	ErrCodeNotificationFailed      ErrorCode = "NOTIFICATION_FAILED"
	ErrCodeNotificationCancelled   ErrorCode = "NOTIFICATION_CANCELLED"
	ErrCodeNotificationAlreadySent ErrorCode = "NOTIFICATION_ALREADY_SENT"
	ErrCodeDeliveryFailed          ErrorCode = "NOTIFICATION_DELIVERY_FAILED"
	ErrCodeRetryFailed             ErrorCode = "NOTIFICATION_RETRY_FAILED"
	ErrCodeMaxRetriesExceeded      ErrorCode = "NOTIFICATION_MAX_RETRIES_EXCEEDED"
	ErrCodeRateLimitExceeded       ErrorCode = "NOTIFICATION_RATE_LIMIT_EXCEEDED"
	ErrCodeQuotaExceeded           ErrorCode = "NOTIFICATION_QUOTA_EXCEEDED"

	// Template errors
	ErrCodeTemplateNotFound      ErrorCode = "TEMPLATE_NOT_FOUND"
	ErrCodeTemplateAlreadyExists ErrorCode = "TEMPLATE_ALREADY_EXISTS"
	ErrCodeTemplateInUse         ErrorCode = "TEMPLATE_IN_USE"
	ErrCodeTemplateRenderFailed  ErrorCode = "TEMPLATE_RENDER_FAILED"
	ErrCodeTemplateInvalid       ErrorCode = "TEMPLATE_INVALID"
	ErrCodeTemplateVersionError  ErrorCode = "TEMPLATE_VERSION_ERROR"

	// Channel errors
	ErrCodeChannelNotConfigured ErrorCode = "CHANNEL_NOT_CONFIGURED"
	ErrCodeChannelDisabled      ErrorCode = "CHANNEL_DISABLED"
	ErrCodeChannelUnavailable   ErrorCode = "CHANNEL_UNAVAILABLE"
	ErrCodeProviderError        ErrorCode = "PROVIDER_ERROR"
	ErrCodeProviderUnavailable  ErrorCode = "PROVIDER_UNAVAILABLE"

	// Email-specific errors
	ErrCodeInvalidEmailAddress ErrorCode = "INVALID_EMAIL_ADDRESS"
	ErrCodeEmailDeliveryFailed ErrorCode = "EMAIL_DELIVERY_FAILED"
	ErrCodeEmailBounced        ErrorCode = "EMAIL_BOUNCED"
	ErrCodeEmailComplaint      ErrorCode = "EMAIL_COMPLAINT"
	ErrCodeEmailUnsubscribed   ErrorCode = "EMAIL_UNSUBSCRIBED"

	// SMS-specific errors
	ErrCodeInvalidPhoneNumber ErrorCode = "INVALID_PHONE_NUMBER"
	ErrCodeSMSDeliveryFailed  ErrorCode = "SMS_DELIVERY_FAILED"
	ErrCodeSMSOptedOut        ErrorCode = "SMS_OPTED_OUT"
	ErrCodeSMSNotSupported    ErrorCode = "SMS_NOT_SUPPORTED"

	// Push notification errors
	ErrCodeInvalidDeviceToken   ErrorCode = "INVALID_DEVICE_TOKEN"
	ErrCodePushDeliveryFailed   ErrorCode = "PUSH_DELIVERY_FAILED"
	ErrCodeDeviceNotRegistered  ErrorCode = "DEVICE_NOT_REGISTERED"
	ErrCodePushPayloadTooLarge  ErrorCode = "PUSH_PAYLOAD_TOO_LARGE"

	// In-app notification errors
	ErrCodeInAppDeliveryFailed ErrorCode = "IN_APP_DELIVERY_FAILED"
	ErrCodeUserNotFound        ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserInactive        ErrorCode = "USER_INACTIVE"

	// Webhook errors
	ErrCodeInvalidWebhookURL    ErrorCode = "INVALID_WEBHOOK_URL"
	ErrCodeWebhookDeliveryFailed ErrorCode = "WEBHOOK_DELIVERY_FAILED"
	ErrCodeWebhookTimeout       ErrorCode = "WEBHOOK_TIMEOUT"

	// Preference errors
	ErrCodePreferenceNotFound   ErrorCode = "PREFERENCE_NOT_FOUND"
	ErrCodePreferenceConflict   ErrorCode = "PREFERENCE_CONFLICT"
	ErrCodeRecipientOptedOut    ErrorCode = "RECIPIENT_OPTED_OUT"

	// Batch errors
	ErrCodeBatchNotFound    ErrorCode = "BATCH_NOT_FOUND"
	ErrCodeBatchTooLarge    ErrorCode = "BATCH_TOO_LARGE"
	ErrCodeBatchEmpty       ErrorCode = "BATCH_EMPTY"
	ErrCodeBatchInProgress  ErrorCode = "BATCH_IN_PROGRESS"

	// Scheduling errors
	ErrCodeInvalidScheduledTime ErrorCode = "INVALID_SCHEDULED_TIME"
	ErrCodeScheduledTimeInPast  ErrorCode = "SCHEDULED_TIME_IN_PAST"
	ErrCodeAlreadyCancelled     ErrorCode = "ALREADY_CANCELLED"
	ErrCodeCannotCancel         ErrorCode = "CANNOT_CANCEL"
	ErrCodeCannotRetry          ErrorCode = "CANNOT_RETRY"

	// Tenant errors
	ErrCodeTenantNotConfigured   ErrorCode = "TENANT_NOT_CONFIGURED"
	ErrCodeTenantChannelDisabled ErrorCode = "TENANT_CHANNEL_DISABLED"
)

// AppError represents an application-level error with code and message.
type AppError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Inner   error                  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Inner)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the inner error.
func (e *AppError) Unwrap() error {
	return e.Inner
}

// WithDetails adds details to the error.
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithDetail adds a single detail to the error.
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithInner wraps an inner error.
func (e *AppError) WithInner(inner error) *AppError {
	e.Inner = inner
	return e
}

// Is checks if the error matches a target error.
func (e *AppError) Is(target error) bool {
	var appErr *AppError
	if errors.As(target, &appErr) {
		return e.Code == appErr.Code
	}
	return false
}

// NewAppError creates a new application error.
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, inner error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: message,
		Inner:   inner,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
	}
}

// NewValidationErrorWithDetails creates a new validation error with field details.
func NewValidationErrorWithDetails(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resourceType, identifier string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found: %s", resourceType, identifier),
		Details: map[string]interface{}{
			"resource_type": resourceType,
			"identifier":    identifier,
		},
	}
}

// NewAlreadyExistsError creates a new already exists error.
func NewAlreadyExistsError(resourceType, identifier string) *AppError {
	return &AppError{
		Code:    ErrCodeAlreadyExists,
		Message: fmt.Sprintf("%s already exists: %s", resourceType, identifier),
		Details: map[string]interface{}{
			"resource_type": resourceType,
			"identifier":    identifier,
		},
	}
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

// NewForbiddenError creates a new forbidden error.
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
	}
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: message,
	}
}

// NewInvalidInputError creates a new invalid input error.
func NewInvalidInputError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidInput,
		Message: message,
	}
}

// NewInvalidStateError creates a new invalid state error.
func NewInvalidStateError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidState,
		Message: message,
	}
}

// Notification-specific error constructors

// NewNotificationNotFoundError creates a notification not found error.
func NewNotificationNotFoundError(notificationID string) *AppError {
	return &AppError{
		Code:    ErrCodeNotificationNotFound,
		Message: fmt.Sprintf("notification not found: %s", notificationID),
		Details: map[string]interface{}{
			"notification_id": notificationID,
		},
	}
}

// NewNotificationFailedError creates a notification failed error.
func NewNotificationFailedError(notificationID, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeNotificationFailed,
		Message: fmt.Sprintf("notification failed: %s", reason),
		Details: map[string]interface{}{
			"notification_id": notificationID,
			"reason":          reason,
		},
	}
}

// NewDeliveryFailedError creates a delivery failed error.
func NewDeliveryFailedError(notificationID, channel, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeDeliveryFailed,
		Message: fmt.Sprintf("delivery failed via %s: %s", channel, reason),
		Details: map[string]interface{}{
			"notification_id": notificationID,
			"channel":         channel,
			"reason":          reason,
		},
	}
}

// NewRetryFailedError creates a retry failed error.
func NewRetryFailedError(notificationID string, attempt int, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeRetryFailed,
		Message: fmt.Sprintf("retry attempt %d failed: %s", attempt, reason),
		Details: map[string]interface{}{
			"notification_id": notificationID,
			"attempt":         attempt,
			"reason":          reason,
		},
	}
}

// NewMaxRetriesExceededError creates a max retries exceeded error.
func NewMaxRetriesExceededError(notificationID string, maxRetries int) *AppError {
	return &AppError{
		Code:    ErrCodeMaxRetriesExceeded,
		Message: fmt.Sprintf("maximum retry attempts (%d) exceeded", maxRetries),
		Details: map[string]interface{}{
			"notification_id": notificationID,
			"max_retries":     maxRetries,
		},
	}
}

// NewRateLimitExceededError creates a rate limit exceeded error.
func NewRateLimitExceededError(channel string, limit int, window string) *AppError {
	return &AppError{
		Code:    ErrCodeRateLimitExceeded,
		Message: fmt.Sprintf("rate limit exceeded for %s: %d per %s", channel, limit, window),
		Details: map[string]interface{}{
			"channel": channel,
			"limit":   limit,
			"window":  window,
		},
	}
}

// NewQuotaExceededError creates a quota exceeded error.
func NewQuotaExceededError(channel string, quota int, period string) *AppError {
	return &AppError{
		Code:    ErrCodeQuotaExceeded,
		Message: fmt.Sprintf("quota exceeded for %s: %d per %s", channel, quota, period),
		Details: map[string]interface{}{
			"channel": channel,
			"quota":   quota,
			"period":  period,
		},
	}
}

// Template error constructors

// NewTemplateNotFoundError creates a template not found error.
func NewTemplateNotFoundError(templateID string) *AppError {
	return &AppError{
		Code:    ErrCodeTemplateNotFound,
		Message: fmt.Sprintf("template not found: %s", templateID),
		Details: map[string]interface{}{
			"template_id": templateID,
		},
	}
}

// NewTemplateAlreadyExistsError creates a template already exists error.
func NewTemplateAlreadyExistsError(templateName string) *AppError {
	return &AppError{
		Code:    ErrCodeTemplateAlreadyExists,
		Message: fmt.Sprintf("template already exists: %s", templateName),
		Details: map[string]interface{}{
			"template_name": templateName,
		},
	}
}

// NewTemplateInUseError creates a template in use error.
func NewTemplateInUseError(templateID string, usageCount int) *AppError {
	return &AppError{
		Code:    ErrCodeTemplateInUse,
		Message: fmt.Sprintf("template is in use and cannot be deleted: %s", templateID),
		Details: map[string]interface{}{
			"template_id": templateID,
			"usage_count": usageCount,
		},
	}
}

// NewTemplateRenderFailedError creates a template render failed error.
func NewTemplateRenderFailedError(templateID, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeTemplateRenderFailed,
		Message: fmt.Sprintf("template rendering failed: %s", reason),
		Details: map[string]interface{}{
			"template_id": templateID,
			"reason":      reason,
		},
	}
}

// Channel error constructors

// NewChannelNotConfiguredError creates a channel not configured error.
func NewChannelNotConfiguredError(channel string) *AppError {
	return &AppError{
		Code:    ErrCodeChannelNotConfigured,
		Message: fmt.Sprintf("notification channel not configured: %s", channel),
		Details: map[string]interface{}{
			"channel": channel,
		},
	}
}

// NewChannelDisabledError creates a channel disabled error.
func NewChannelDisabledError(channel string) *AppError {
	return &AppError{
		Code:    ErrCodeChannelDisabled,
		Message: fmt.Sprintf("notification channel is disabled: %s", channel),
		Details: map[string]interface{}{
			"channel": channel,
		},
	}
}

// NewProviderErrorError creates a provider error.
func NewProviderErrorError(provider, message string, inner error) *AppError {
	return &AppError{
		Code:    ErrCodeProviderError,
		Message: fmt.Sprintf("provider error from %s: %s", provider, message),
		Details: map[string]interface{}{
			"provider": provider,
		},
		Inner: inner,
	}
}

// NewProviderUnavailableError creates a provider unavailable error.
func NewProviderUnavailableError(provider string) *AppError {
	return &AppError{
		Code:    ErrCodeProviderUnavailable,
		Message: fmt.Sprintf("notification provider unavailable: %s", provider),
		Details: map[string]interface{}{
			"provider": provider,
		},
	}
}

// Email error constructors

// NewInvalidEmailAddressError creates an invalid email address error.
func NewInvalidEmailAddressError(email string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidEmailAddress,
		Message: fmt.Sprintf("invalid email address: %s", email),
		Details: map[string]interface{}{
			"email": email,
		},
	}
}

// NewEmailDeliveryFailedError creates an email delivery failed error.
func NewEmailDeliveryFailedError(email, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeEmailDeliveryFailed,
		Message: fmt.Sprintf("email delivery failed to %s: %s", email, reason),
		Details: map[string]interface{}{
			"email":  email,
			"reason": reason,
		},
	}
}

// NewEmailBouncedError creates an email bounced error.
func NewEmailBouncedError(email, bounceType string) *AppError {
	return &AppError{
		Code:    ErrCodeEmailBounced,
		Message: fmt.Sprintf("email bounced for %s: %s", email, bounceType),
		Details: map[string]interface{}{
			"email":       email,
			"bounce_type": bounceType,
		},
	}
}

// SMS error constructors

// NewInvalidPhoneNumberError creates an invalid phone number error.
func NewInvalidPhoneNumberError(phone string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidPhoneNumber,
		Message: fmt.Sprintf("invalid phone number: %s", phone),
		Details: map[string]interface{}{
			"phone": phone,
		},
	}
}

// NewSMSDeliveryFailedError creates an SMS delivery failed error.
func NewSMSDeliveryFailedError(phone, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeSMSDeliveryFailed,
		Message: fmt.Sprintf("SMS delivery failed to %s: %s", phone, reason),
		Details: map[string]interface{}{
			"phone":  phone,
			"reason": reason,
		},
	}
}

// NewSMSOptedOutError creates an SMS opted out error.
func NewSMSOptedOutError(phone string) *AppError {
	return &AppError{
		Code:    ErrCodeSMSOptedOut,
		Message: fmt.Sprintf("recipient has opted out of SMS notifications: %s", phone),
		Details: map[string]interface{}{
			"phone": phone,
		},
	}
}

// Push notification error constructors

// NewInvalidDeviceTokenError creates an invalid device token error.
func NewInvalidDeviceTokenError(token string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidDeviceToken,
		Message: "invalid device token",
		Details: map[string]interface{}{
			"token_prefix": token[:min(8, len(token))] + "...",
		},
	}
}

// NewPushDeliveryFailedError creates a push delivery failed error.
func NewPushDeliveryFailedError(deviceID, reason string) *AppError {
	return &AppError{
		Code:    ErrCodePushDeliveryFailed,
		Message: fmt.Sprintf("push notification delivery failed: %s", reason),
		Details: map[string]interface{}{
			"device_id": deviceID,
			"reason":    reason,
		},
	}
}

// NewDeviceNotRegisteredError creates a device not registered error.
func NewDeviceNotRegisteredError(deviceID string) *AppError {
	return &AppError{
		Code:    ErrCodeDeviceNotRegistered,
		Message: "device not registered for push notifications",
		Details: map[string]interface{}{
			"device_id": deviceID,
		},
	}
}

// In-app notification error constructors

// NewInAppDeliveryFailedError creates an in-app delivery failed error.
func NewInAppDeliveryFailedError(userID, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeInAppDeliveryFailed,
		Message: fmt.Sprintf("in-app notification delivery failed: %s", reason),
		Details: map[string]interface{}{
			"user_id": userID,
			"reason":  reason,
		},
	}
}

// NewUserNotFoundError creates a user not found error.
func NewUserNotFoundError(userID string) *AppError {
	return &AppError{
		Code:    ErrCodeUserNotFound,
		Message: fmt.Sprintf("user not found: %s", userID),
		Details: map[string]interface{}{
			"user_id": userID,
		},
	}
}

// Webhook error constructors

// NewInvalidWebhookURLError creates an invalid webhook URL error.
func NewInvalidWebhookURLError(url string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidWebhookURL,
		Message: fmt.Sprintf("invalid webhook URL: %s", url),
		Details: map[string]interface{}{
			"url": url,
		},
	}
}

// NewWebhookDeliveryFailedError creates a webhook delivery failed error.
func NewWebhookDeliveryFailedError(url string, statusCode int, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeWebhookDeliveryFailed,
		Message: fmt.Sprintf("webhook delivery failed to %s: %s", url, reason),
		Details: map[string]interface{}{
			"url":         url,
			"status_code": statusCode,
			"reason":      reason,
		},
	}
}

// NewWebhookTimeoutError creates a webhook timeout error.
func NewWebhookTimeoutError(url string, timeout int) *AppError {
	return &AppError{
		Code:    ErrCodeWebhookTimeout,
		Message: fmt.Sprintf("webhook request timed out after %d seconds", timeout),
		Details: map[string]interface{}{
			"url":            url,
			"timeout_seconds": timeout,
		},
	}
}

// Preference error constructors

// NewPreferenceNotFoundError creates a preference not found error.
func NewPreferenceNotFoundError(userID string) *AppError {
	return &AppError{
		Code:    ErrCodePreferenceNotFound,
		Message: fmt.Sprintf("notification preference not found for user: %s", userID),
		Details: map[string]interface{}{
			"user_id": userID,
		},
	}
}

// NewRecipientOptedOutError creates a recipient opted out error.
func NewRecipientOptedOutError(recipientID, channel string) *AppError {
	return &AppError{
		Code:    ErrCodeRecipientOptedOut,
		Message: fmt.Sprintf("recipient has opted out of %s notifications", channel),
		Details: map[string]interface{}{
			"recipient_id": recipientID,
			"channel":      channel,
		},
	}
}

// Batch error constructors

// NewBatchNotFoundError creates a batch not found error.
func NewBatchNotFoundError(batchID string) *AppError {
	return &AppError{
		Code:    ErrCodeBatchNotFound,
		Message: fmt.Sprintf("notification batch not found: %s", batchID),
		Details: map[string]interface{}{
			"batch_id": batchID,
		},
	}
}

// NewBatchTooLargeError creates a batch too large error.
func NewBatchTooLargeError(size, maxSize int) *AppError {
	return &AppError{
		Code:    ErrCodeBatchTooLarge,
		Message: fmt.Sprintf("batch size %d exceeds maximum allowed size %d", size, maxSize),
		Details: map[string]interface{}{
			"size":     size,
			"max_size": maxSize,
		},
	}
}

// NewBatchEmptyError creates a batch empty error.
func NewBatchEmptyError(batchID string) *AppError {
	return &AppError{
		Code:    ErrCodeBatchEmpty,
		Message: "notification batch is empty",
		Details: map[string]interface{}{
			"batch_id": batchID,
		},
	}
}

// Scheduling error constructors

// NewInvalidScheduledTimeError creates an invalid scheduled time error.
func NewInvalidScheduledTimeError(reason string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidScheduledTime,
		Message: fmt.Sprintf("invalid scheduled time: %s", reason),
	}
}

// NewScheduledTimeInPastError creates a scheduled time in past error.
func NewScheduledTimeInPastError() *AppError {
	return &AppError{
		Code:    ErrCodeScheduledTimeInPast,
		Message: "scheduled time cannot be in the past",
	}
}

// NewCannotCancelError creates a cannot cancel error.
func NewCannotCancelError(notificationID, status string) *AppError {
	return &AppError{
		Code:    ErrCodeCannotCancel,
		Message: fmt.Sprintf("notification cannot be cancelled in status: %s", status),
		Details: map[string]interface{}{
			"notification_id": notificationID,
			"current_status":  status,
		},
	}
}

// NewCannotRetryError creates a cannot retry error.
func NewCannotRetryError(notificationID, status string) *AppError {
	return &AppError{
		Code:    ErrCodeCannotRetry,
		Message: fmt.Sprintf("notification cannot be retried in status: %s", status),
		Details: map[string]interface{}{
			"notification_id": notificationID,
			"current_status":  status,
		},
	}
}

// Tenant error constructors

// NewTenantNotConfiguredError creates a tenant not configured error.
func NewTenantNotConfiguredError(tenantID string) *AppError {
	return &AppError{
		Code:    ErrCodeTenantNotConfigured,
		Message: fmt.Sprintf("notification settings not configured for tenant: %s", tenantID),
		Details: map[string]interface{}{
			"tenant_id": tenantID,
		},
	}
}

// NewTenantChannelDisabledError creates a tenant channel disabled error.
func NewTenantChannelDisabledError(tenantID, channel string) *AppError {
	return &AppError{
		Code:    ErrCodeTenantChannelDisabled,
		Message: fmt.Sprintf("notification channel %s is disabled for tenant: %s", channel, tenantID),
		Details: map[string]interface{}{
			"tenant_id": tenantID,
			"channel":   channel,
		},
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
