// Package application contains the application layer for the IAM service.
// It implements use cases that orchestrate domain logic and infrastructure.
package application

import (
	"errors"
	"fmt"
)

// Application layer error codes
const (
	ErrCodeValidation        = "VALIDATION_ERROR"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeConflict          = "CONFLICT"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeForbidden         = "FORBIDDEN"
	ErrCodeInternal          = "INTERNAL_ERROR"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired      = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid      = "TOKEN_INVALID"
	ErrCodeRateLimited       = "RATE_LIMITED"
	ErrCodeTenantInactive    = "TENANT_INACTIVE"
	ErrCodeUserInactive      = "USER_INACTIVE"
	ErrCodeEmailNotVerified  = "EMAIL_NOT_VERIFIED"
	ErrCodePermissionDenied  = "PERMISSION_DENIED"
)

// AppError represents an application layer error with code and context.
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Err        error                  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithDetail adds a detail to the error.
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails adds multiple details to the error.
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// NewAppError creates a new application error.
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Predefined application errors

// ErrValidation creates a validation error.
func ErrValidation(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

// ErrNotFound creates a not found error.
func ErrNotFound(entity string, id interface{}) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", entity),
		Details: map[string]interface{}{"id": id},
	}
}

// ErrConflict creates a conflict error.
func ErrConflict(message string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: message,
	}
}

// ErrUnauthorized creates an unauthorized error.
func ErrUnauthorized(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

// ErrForbidden creates a forbidden error.
func ErrForbidden(message string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
	}
}

// ErrInternal creates an internal error.
func ErrInternal(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     err,
	}
}

// ErrInvalidCredentials creates an invalid credentials error.
func ErrInvalidCredentials() *AppError {
	return &AppError{
		Code:    ErrCodeInvalidCredentials,
		Message: "Invalid email or password",
	}
}

// ErrTokenExpired creates a token expired error.
func ErrTokenExpired() *AppError {
	return &AppError{
		Code:    ErrCodeTokenExpired,
		Message: "Token has expired",
	}
}

// ErrTokenInvalid creates a token invalid error.
func ErrTokenInvalid() *AppError {
	return &AppError{
		Code:    ErrCodeTokenInvalid,
		Message: "Token is invalid",
	}
}

// ErrRateLimited creates a rate limited error.
func ErrRateLimited() *AppError {
	return &AppError{
		Code:    ErrCodeRateLimited,
		Message: "Too many requests, please try again later",
	}
}

// ErrTenantInactive creates a tenant inactive error.
func ErrTenantInactive() *AppError {
	return &AppError{
		Code:    ErrCodeTenantInactive,
		Message: "Tenant is not active",
	}
}

// ErrUserInactive creates a user inactive error.
func ErrUserInactive() *AppError {
	return &AppError{
		Code:    ErrCodeUserInactive,
		Message: "User account is not active",
	}
}

// ErrEmailNotVerified creates an email not verified error.
func ErrEmailNotVerified() *AppError {
	return &AppError{
		Code:    ErrCodeEmailNotVerified,
		Message: "Email address has not been verified",
	}
}

// ErrPermissionDenied creates a permission denied error.
func ErrPermissionDenied(permission string) *AppError {
	return &AppError{
		Code:    ErrCodePermissionDenied,
		Message: "Permission denied",
		Details: map[string]interface{}{"required_permission": permission},
	}
}

// IsAppError checks if the error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts the AppError from an error.
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// IsNotFoundError checks if the error is a not found error.
func IsNotFoundError(err error) bool {
	appErr := GetAppError(err)
	return appErr != nil && appErr.Code == ErrCodeNotFound
}

// IsValidationError checks if the error is a validation error.
func IsValidationError(err error) bool {
	appErr := GetAppError(err)
	return appErr != nil && appErr.Code == ErrCodeValidation
}

// IsUnauthorizedError checks if the error is an unauthorized error.
func IsUnauthorizedError(err error) bool {
	appErr := GetAppError(err)
	return appErr != nil && appErr.Code == ErrCodeUnauthorized
}

// IsForbiddenError checks if the error is a forbidden error.
func IsForbiddenError(err error) bool {
	appErr := GetAppError(err)
	return appErr != nil && appErr.Code == ErrCodeForbidden
}
