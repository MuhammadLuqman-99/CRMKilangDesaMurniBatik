// Package errors provides custom error types and utilities for the CRM application.
// It implements a structured error handling approach with error codes, HTTP status mapping,
// and support for error wrapping and stack traces.
package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// ErrorCode represents a unique error code for categorizing errors.
type ErrorCode string

// Error codes for the application
const (
	// General errors
	ErrCodeUnknown          ErrorCode = "UNKNOWN"
	ErrCodeInternal         ErrorCode = "INTERNAL_ERROR"
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeTooManyRequests  ErrorCode = "TOO_MANY_REQUESTS"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout          ErrorCode = "TIMEOUT"

	// Authentication errors
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrCodeRefreshTokenExpired ErrorCode = "REFRESH_TOKEN_EXPIRED"

	// Tenant errors
	ErrCodeTenantNotFound    ErrorCode = "TENANT_NOT_FOUND"
	ErrCodeTenantSuspended   ErrorCode = "TENANT_SUSPENDED"
	ErrCodeTenantLimitExceeded ErrorCode = "TENANT_LIMIT_EXCEEDED"

	// User errors
	ErrCodeUserNotFound     ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserDisabled     ErrorCode = "USER_DISABLED"
	ErrCodeEmailExists      ErrorCode = "EMAIL_EXISTS"
	ErrCodeWeakPassword     ErrorCode = "WEAK_PASSWORD"

	// Customer errors
	ErrCodeCustomerNotFound ErrorCode = "CUSTOMER_NOT_FOUND"
	ErrCodeContactNotFound  ErrorCode = "CONTACT_NOT_FOUND"

	// Sales errors
	ErrCodeLeadNotFound        ErrorCode = "LEAD_NOT_FOUND"
	ErrCodeOpportunityNotFound ErrorCode = "OPPORTUNITY_NOT_FOUND"
	ErrCodeDealNotFound        ErrorCode = "DEAL_NOT_FOUND"
	ErrCodeInvalidStageTransition ErrorCode = "INVALID_STAGE_TRANSITION"

	// Database errors
	ErrCodeDBConnection ErrorCode = "DB_CONNECTION_ERROR"
	ErrCodeDBQuery      ErrorCode = "DB_QUERY_ERROR"
	ErrCodeDBTransaction ErrorCode = "DB_TRANSACTION_ERROR"

	// External service errors
	ErrCodeExternalService ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrCodeEmailDelivery   ErrorCode = "EMAIL_DELIVERY_ERROR"
	ErrCodeSMSDelivery     ErrorCode = "SMS_DELIVERY_ERROR"
)

// httpStatusMap maps error codes to HTTP status codes
var httpStatusMap = map[ErrorCode]int{
	ErrCodeUnknown:            http.StatusInternalServerError,
	ErrCodeInternal:           http.StatusInternalServerError,
	ErrCodeValidation:         http.StatusBadRequest,
	ErrCodeNotFound:           http.StatusNotFound,
	ErrCodeAlreadyExists:      http.StatusConflict,
	ErrCodeUnauthorized:       http.StatusUnauthorized,
	ErrCodeForbidden:          http.StatusForbidden,
	ErrCodeBadRequest:         http.StatusBadRequest,
	ErrCodeConflict:           http.StatusConflict,
	ErrCodeTooManyRequests:    http.StatusTooManyRequests,
	ErrCodeServiceUnavailable: http.StatusServiceUnavailable,
	ErrCodeTimeout:            http.StatusGatewayTimeout,
	ErrCodeInvalidCredentials: http.StatusUnauthorized,
	ErrCodeTokenExpired:       http.StatusUnauthorized,
	ErrCodeTokenInvalid:       http.StatusUnauthorized,
	ErrCodeRefreshTokenExpired: http.StatusUnauthorized,
	ErrCodeTenantNotFound:     http.StatusNotFound,
	ErrCodeTenantSuspended:    http.StatusForbidden,
	ErrCodeTenantLimitExceeded: http.StatusForbidden,
	ErrCodeUserNotFound:       http.StatusNotFound,
	ErrCodeUserDisabled:       http.StatusForbidden,
	ErrCodeEmailExists:        http.StatusConflict,
	ErrCodeWeakPassword:       http.StatusBadRequest,
	ErrCodeCustomerNotFound:   http.StatusNotFound,
	ErrCodeContactNotFound:    http.StatusNotFound,
	ErrCodeLeadNotFound:       http.StatusNotFound,
	ErrCodeOpportunityNotFound: http.StatusNotFound,
	ErrCodeDealNotFound:       http.StatusNotFound,
	ErrCodeInvalidStageTransition: http.StatusBadRequest,
	ErrCodeDBConnection:       http.StatusServiceUnavailable,
	ErrCodeDBQuery:            http.StatusInternalServerError,
	ErrCodeDBTransaction:      http.StatusInternalServerError,
	ErrCodeExternalService:    http.StatusBadGateway,
	ErrCodeEmailDelivery:      http.StatusBadGateway,
	ErrCodeSMSDelivery:        http.StatusBadGateway,
}

// AppError represents a structured application error.
type AppError struct {
	Code       ErrorCode         `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	Fields     map[string]string `json:"fields,omitempty"`
	cause      error
	stackTrace string
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.cause
}

// HTTPStatus returns the HTTP status code for this error.
func (e *AppError) HTTPStatus() int {
	if status, ok := httpStatusMap[e.Code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// WithDetails adds additional details to the error.
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithField adds a field-specific error.
func (e *AppError) WithField(field, message string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	e.Fields[field] = message
	return e
}

// WithFields adds multiple field-specific errors.
func (e *AppError) WithFields(fields map[string]string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	for k, v := range fields {
		e.Fields[k] = v
	}
	return e
}

// StackTrace returns the stack trace of where the error was created.
func (e *AppError) StackTrace() string {
	return e.stackTrace
}

// captureStackTrace captures the current stack trace.
func captureStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		}
		if !more {
			break
		}
	}
	return sb.String()
}

// New creates a new AppError with the given code and message.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		stackTrace: captureStackTrace(),
	}
}

// Newf creates a new AppError with a formatted message.
func Newf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		stackTrace: captureStackTrace(),
	}
}

// Wrap wraps an existing error with an AppError.
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:       code,
		Message:    message,
		cause:      err,
		stackTrace: captureStackTrace(),
	}
}

// Wrapf wraps an existing error with a formatted message.
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		cause:      err,
		stackTrace: captureStackTrace(),
	}
}

// Convenience constructors for common errors

// ErrInternal creates an internal server error.
func ErrInternal(message string) *AppError {
	return New(ErrCodeInternal, message)
}

// ErrInternalWrap wraps an error as an internal server error.
func ErrInternalWrap(err error, message string) *AppError {
	return Wrap(err, ErrCodeInternal, message)
}

// ErrNotFound creates a not found error.
func ErrNotFound(resource string) *AppError {
	return Newf(ErrCodeNotFound, "%s not found", resource)
}

// ErrValidation creates a validation error.
func ErrValidation(message string) *AppError {
	return New(ErrCodeValidation, message)
}

// ErrUnauthorized creates an unauthorized error.
func ErrUnauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message)
}

// ErrForbidden creates a forbidden error.
func ErrForbidden(message string) *AppError {
	return New(ErrCodeForbidden, message)
}

// ErrBadRequest creates a bad request error.
func ErrBadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message)
}

// ErrConflict creates a conflict error.
func ErrConflict(message string) *AppError {
	return New(ErrCodeConflict, message)
}

// ErrAlreadyExists creates an already exists error.
func ErrAlreadyExists(resource string) *AppError {
	return Newf(ErrCodeAlreadyExists, "%s already exists", resource)
}

// ErrTooManyRequests creates a rate limit error.
func ErrTooManyRequests(message string) *AppError {
	return New(ErrCodeTooManyRequests, message)
}

// ErrServiceUnavailable creates a service unavailable error.
func ErrServiceUnavailable(service string) *AppError {
	return Newf(ErrCodeServiceUnavailable, "%s is currently unavailable", service)
}

// ErrTimeout creates a timeout error.
func ErrTimeout(operation string) *AppError {
	return Newf(ErrCodeTimeout, "%s timed out", operation)
}

// IsAppError checks if the error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError attempts to convert an error to an AppError.
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetCode returns the error code from an error, or ErrCodeUnknown if not an AppError.
func GetCode(err error) ErrorCode {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code
	}
	return ErrCodeUnknown
}

// GetHTTPStatus returns the HTTP status code from an error.
func GetHTTPStatus(err error) int {
	if appErr, ok := AsAppError(err); ok {
		return appErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// Is checks if an error has a specific error code.
func Is(err error, code ErrorCode) bool {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code == code
	}
	return false
}
