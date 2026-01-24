// Package application contains the application layer for the Customer service.
package application

import (
	"fmt"

	"github.com/google/uuid"
)

// Error codes for the Customer application layer.
const (
	// Customer errors
	ErrCodeCustomerNotFound          = "CUSTOMER_NOT_FOUND"
	ErrCodeCustomerAlreadyExists     = "CUSTOMER_ALREADY_EXISTS"
	ErrCodeCustomerInvalidStatus     = "CUSTOMER_INVALID_STATUS"
	ErrCodeCustomerVersionConflict   = "CUSTOMER_VERSION_CONFLICT"
	ErrCodeCustomerValidation        = "CUSTOMER_VALIDATION_ERROR"
	ErrCodeCustomerDuplicate         = "CUSTOMER_DUPLICATE"
	ErrCodeCustomerMaxContactsReached = "CUSTOMER_MAX_CONTACTS_REACHED"
	ErrCodeCustomerCannotDelete      = "CUSTOMER_CANNOT_DELETE"
	ErrCodeCustomerCannotMerge       = "CUSTOMER_CANNOT_MERGE"

	// Contact errors
	ErrCodeContactNotFound        = "CONTACT_NOT_FOUND"
	ErrCodeContactAlreadyExists   = "CONTACT_ALREADY_EXISTS"
	ErrCodeContactInvalidStatus   = "CONTACT_INVALID_STATUS"
	ErrCodeContactVersionConflict = "CONTACT_VERSION_CONFLICT"
	ErrCodeContactValidation      = "CONTACT_VALIDATION_ERROR"
	ErrCodeContactDuplicate       = "CONTACT_DUPLICATE"
	ErrCodeContactIsBlocked       = "CONTACT_IS_BLOCKED"
	ErrCodeContactCannotDelete    = "CONTACT_CANNOT_DELETE"

	// Authorization errors
	ErrCodeUnauthorized     = "UNAUTHORIZED"
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeInvalidTenant    = "INVALID_TENANT"
	ErrCodeTenantMismatch   = "TENANT_MISMATCH"

	// Import/Export errors
	ErrCodeImportFailed      = "IMPORT_FAILED"
	ErrCodeExportFailed      = "EXPORT_FAILED"
	ErrCodeInvalidFormat     = "INVALID_FORMAT"
	ErrCodeInvalidData       = "INVALID_DATA"
	ErrCodeFileTooLarge      = "FILE_TOO_LARGE"

	// General errors
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeInvalidInput      = "INVALID_INPUT"
	ErrCodeOperationFailed   = "OPERATION_FAILED"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// ApplicationError represents an application-level error.
type ApplicationError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	StatusCode int                    `json:"-"`
}

// Error implements the error interface.
func (e *ApplicationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *ApplicationError) Unwrap() error {
	return e.Cause
}

// WithDetail adds a detail to the error.
func (e *ApplicationError) WithDetail(key string, value interface{}) *ApplicationError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithCause sets the underlying cause.
func (e *ApplicationError) WithCause(cause error) *ApplicationError {
	e.Cause = cause
	return e
}

// ============================================================================
// Error Constructors
// ============================================================================

// NewApplicationError creates a new ApplicationError.
func NewApplicationError(code, message string, statusCode int) *ApplicationError {
	return &ApplicationError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Customer Errors

// ErrCustomerNotFound creates a customer not found error.
func ErrCustomerNotFound(id uuid.UUID) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerNotFound,
		Message:    fmt.Sprintf("customer not found: %s", id),
		Details:    map[string]interface{}{"customer_id": id},
		StatusCode: 404,
	}
}

// ErrCustomerAlreadyExists creates a customer already exists error.
func ErrCustomerAlreadyExists(field, value string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerAlreadyExists,
		Message:    fmt.Sprintf("customer with %s '%s' already exists", field, value),
		Details:    map[string]interface{}{"field": field, "value": value},
		StatusCode: 409,
	}
}

// ErrCustomerInvalidStatus creates an invalid status transition error.
func ErrCustomerInvalidStatus(from, to string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerInvalidStatus,
		Message:    fmt.Sprintf("invalid status transition from '%s' to '%s'", from, to),
		Details:    map[string]interface{}{"from": from, "to": to},
		StatusCode: 400,
	}
}

// ErrCustomerVersionConflict creates a version conflict error.
func ErrCustomerVersionConflict(id uuid.UUID, expected, actual int) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerVersionConflict,
		Message:    fmt.Sprintf("customer %s has been modified (expected version %d, actual %d)", id, expected, actual),
		Details:    map[string]interface{}{"customer_id": id, "expected_version": expected, "actual_version": actual},
		StatusCode: 409,
	}
}

// ErrCustomerValidation creates a validation error.
func ErrCustomerValidation(message string, details map[string]interface{}) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerValidation,
		Message:    message,
		Details:    details,
		StatusCode: 400,
	}
}

// ErrCustomerDuplicate creates a duplicate customer error.
func ErrCustomerDuplicate(matchType string, matchID uuid.UUID) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerDuplicate,
		Message:    fmt.Sprintf("potential duplicate customer found (matched by %s)", matchType),
		Details:    map[string]interface{}{"match_type": matchType, "match_id": matchID},
		StatusCode: 409,
	}
}

// ErrCustomerMaxContactsReached creates a max contacts error.
func ErrCustomerMaxContactsReached(id uuid.UUID, max int) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerMaxContactsReached,
		Message:    fmt.Sprintf("customer %s has reached maximum contacts limit (%d)", id, max),
		Details:    map[string]interface{}{"customer_id": id, "max_contacts": max},
		StatusCode: 400,
	}
}

// ErrCustomerCannotDelete creates a cannot delete error.
func ErrCustomerCannotDelete(id uuid.UUID, reason string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerCannotDelete,
		Message:    fmt.Sprintf("customer %s cannot be deleted: %s", id, reason),
		Details:    map[string]interface{}{"customer_id": id, "reason": reason},
		StatusCode: 400,
	}
}

// ErrCustomerCannotMerge creates a cannot merge error.
func ErrCustomerCannotMerge(reason string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeCustomerCannotMerge,
		Message:    fmt.Sprintf("customers cannot be merged: %s", reason),
		Details:    map[string]interface{}{"reason": reason},
		StatusCode: 400,
	}
}

// Contact Errors

// ErrContactNotFound creates a contact not found error.
func ErrContactNotFound(id uuid.UUID) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeContactNotFound,
		Message:    fmt.Sprintf("contact not found: %s", id),
		Details:    map[string]interface{}{"contact_id": id},
		StatusCode: 404,
	}
}

// ErrContactAlreadyExists creates a contact already exists error.
func ErrContactAlreadyExists(field, value string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeContactAlreadyExists,
		Message:    fmt.Sprintf("contact with %s '%s' already exists", field, value),
		Details:    map[string]interface{}{"field": field, "value": value},
		StatusCode: 409,
	}
}

// ErrContactVersionConflict creates a version conflict error.
func ErrContactVersionConflict(id uuid.UUID, expected, actual int) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeContactVersionConflict,
		Message:    fmt.Sprintf("contact %s has been modified (expected version %d, actual %d)", id, expected, actual),
		Details:    map[string]interface{}{"contact_id": id, "expected_version": expected, "actual_version": actual},
		StatusCode: 409,
	}
}

// ErrContactValidation creates a validation error.
func ErrContactValidation(message string, details map[string]interface{}) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeContactValidation,
		Message:    message,
		Details:    details,
		StatusCode: 400,
	}
}

// ErrContactIsBlocked creates a blocked contact error.
func ErrContactIsBlocked(id uuid.UUID) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeContactIsBlocked,
		Message:    fmt.Sprintf("contact %s is blocked", id),
		Details:    map[string]interface{}{"contact_id": id},
		StatusCode: 400,
	}
}

// Authorization Errors

// ErrUnauthorized creates an unauthorized error.
func ErrUnauthorized(message string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		StatusCode: 401,
	}
}

// ErrForbidden creates a forbidden error.
func ErrForbidden(resource, action string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeForbidden,
		Message:    fmt.Sprintf("access denied: cannot %s %s", action, resource),
		Details:    map[string]interface{}{"resource": resource, "action": action},
		StatusCode: 403,
	}
}

// ErrInvalidTenant creates an invalid tenant error.
func ErrInvalidTenant(tenantID uuid.UUID) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeInvalidTenant,
		Message:    fmt.Sprintf("invalid or inactive tenant: %s", tenantID),
		Details:    map[string]interface{}{"tenant_id": tenantID},
		StatusCode: 400,
	}
}

// ErrTenantMismatch creates a tenant mismatch error.
func ErrTenantMismatch(expected, actual uuid.UUID) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeTenantMismatch,
		Message:    "resource belongs to a different tenant",
		Details:    map[string]interface{}{"expected": expected, "actual": actual},
		StatusCode: 403,
	}
}

// Import/Export Errors

// ErrImportFailed creates an import failed error.
func ErrImportFailed(message string, failedRows int) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeImportFailed,
		Message:    message,
		Details:    map[string]interface{}{"failed_rows": failedRows},
		StatusCode: 400,
	}
}

// ErrExportFailed creates an export failed error.
func ErrExportFailed(message string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeExportFailed,
		Message:    message,
		StatusCode: 500,
	}
}

// ErrInvalidFormat creates an invalid format error.
func ErrInvalidFormat(format string, supported []string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeInvalidFormat,
		Message:    fmt.Sprintf("unsupported format: %s", format),
		Details:    map[string]interface{}{"format": format, "supported": supported},
		StatusCode: 400,
	}
}

// ErrFileTooLarge creates a file too large error.
func ErrFileTooLarge(size, maxSize int64) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeFileTooLarge,
		Message:    fmt.Sprintf("file size %d exceeds maximum allowed size %d", size, maxSize),
		Details:    map[string]interface{}{"size": size, "max_size": maxSize},
		StatusCode: 413,
	}
}

// General Errors

// ErrInternalError creates an internal error.
func ErrInternalError(message string, cause error) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeInternalError,
		Message:    message,
		Cause:      cause,
		StatusCode: 500,
	}
}

// ErrInvalidInput creates an invalid input error.
func ErrInvalidInput(message string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeInvalidInput,
		Message:    message,
		StatusCode: 400,
	}
}

// ErrOperationFailed creates an operation failed error.
func ErrOperationFailed(operation string, cause error) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeOperationFailed,
		Message:    fmt.Sprintf("operation '%s' failed", operation),
		Cause:      cause,
		StatusCode: 500,
	}
}

// ErrRateLimitExceeded creates a rate limit error.
func ErrRateLimitExceeded(limit int, window string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeRateLimitExceeded,
		Message:    fmt.Sprintf("rate limit exceeded: %d requests per %s", limit, window),
		Details:    map[string]interface{}{"limit": limit, "window": window},
		StatusCode: 429,
	}
}

// ErrServiceUnavailable creates a service unavailable error.
func ErrServiceUnavailable(service string) *ApplicationError {
	return &ApplicationError{
		Code:       ErrCodeServiceUnavailable,
		Message:    fmt.Sprintf("service '%s' is temporarily unavailable", service),
		Details:    map[string]interface{}{"service": service},
		StatusCode: 503,
	}
}

// ============================================================================
// Error Type Checking
// ============================================================================

// IsNotFoundError checks if the error is a not found error.
func IsNotFoundError(err error) bool {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.Code == ErrCodeCustomerNotFound || appErr.Code == ErrCodeContactNotFound
	}
	return false
}

// IsConflictError checks if the error is a conflict error.
func IsConflictError(err error) bool {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.Code == ErrCodeCustomerAlreadyExists ||
			appErr.Code == ErrCodeContactAlreadyExists ||
			appErr.Code == ErrCodeCustomerVersionConflict ||
			appErr.Code == ErrCodeContactVersionConflict ||
			appErr.Code == ErrCodeCustomerDuplicate ||
			appErr.Code == ErrCodeContactDuplicate
	}
	return false
}

// IsValidationError checks if the error is a validation error.
func IsValidationError(err error) bool {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.Code == ErrCodeCustomerValidation ||
			appErr.Code == ErrCodeContactValidation ||
			appErr.Code == ErrCodeInvalidInput
	}
	return false
}

// IsAuthorizationError checks if the error is an authorization error.
func IsAuthorizationError(err error) bool {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.Code == ErrCodeUnauthorized ||
			appErr.Code == ErrCodeForbidden ||
			appErr.Code == ErrCodeInvalidTenant ||
			appErr.Code == ErrCodeTenantMismatch
	}
	return false
}

// GetStatusCode returns the HTTP status code for an error.
func GetStatusCode(err error) int {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.StatusCode
	}
	return 500
}
