// Package http provides HTTP handlers for the Customer service.
package http

import (
	"errors"
	"net/http"

	"github.com/kilang-desa-murni/crm/internal/customer/application"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	StatusCode int               `json:"-"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *ErrorResponse) Error() string {
	return e.Message
}

// HTTP Error constructors

// ErrBadRequest creates a bad request error.
func ErrBadRequest(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
	}
}

// ErrUnauthorized creates an unauthorized error.
func ErrUnauthorized(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

// ErrForbidden creates a forbidden error.
func ErrForbidden(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
	}
}

// ErrNotFound creates a not found error.
func ErrNotFound(resource string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    resource + " not found",
	}
}

// ErrConflict creates a conflict error.
func ErrConflict(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    message,
	}
}

// ErrInternalServer creates an internal server error.
func ErrInternalServer(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
		Message:    message,
	}
}

// ErrValidation creates a validation error.
func ErrValidation(message string, details map[string]string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusUnprocessableEntity,
		Code:       "VALIDATION_ERROR",
		Message:    message,
		Details:    details,
	}
}

// ErrMissingParameter creates a missing parameter error.
func ErrMissingParameter(param string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "MISSING_PARAMETER",
		Message:    "missing required parameter: " + param,
	}
}

// ErrInvalidParameter creates an invalid parameter error.
func ErrInvalidParameter(param, reason string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "INVALID_PARAMETER",
		Message:    "invalid parameter " + param + ": " + reason,
	}
}

// ErrInvalidJSON creates an invalid JSON error.
func ErrInvalidJSON(detail string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "INVALID_JSON",
		Message:    "invalid JSON: " + detail,
	}
}

// ErrInvalidRequest creates an invalid request error.
func ErrInvalidRequest(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "INVALID_REQUEST",
		Message:    message,
	}
}

// ErrTooManyRequests creates a rate limit error.
func ErrTooManyRequests(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusTooManyRequests,
		Code:       "TOO_MANY_REQUESTS",
		Message:    message,
	}
}

// ErrServiceUnavailable creates a service unavailable error.
func ErrServiceUnavailable(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusServiceUnavailable,
		Code:       "SERVICE_UNAVAILABLE",
		Message:    message,
	}
}

// toHTTPError converts application/domain errors to HTTP errors.
func toHTTPError(err error) *ErrorResponse {
	if err == nil {
		return nil
	}

	// Check if it's already an HTTP error
	var httpErr *ErrorResponse
	if errors.As(err, &httpErr) {
		return httpErr
	}

	// Check for application errors
	var appErr *application.AppError
	if errors.As(err, &appErr) {
		return mapAppError(appErr)
	}

	// Check for domain errors
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		return mapDomainError(domainErr)
	}

	// Check for validation errors
	var validationErrs domain.ValidationErrors
	if errors.As(err, &validationErrs) {
		details := make(map[string]string)
		for _, verr := range validationErrs {
			details[verr.Field] = verr.Message
		}
		return ErrValidation("validation failed", details)
	}

	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		return ErrValidation(validationErr.Message, map[string]string{
			validationErr.Field: validationErr.Message,
		})
	}

	// Check for specific domain errors
	if errors.Is(err, domain.ErrCustomerNotFound) {
		return ErrNotFound("customer")
	}
	if errors.Is(err, domain.ErrContactNotFound) {
		return ErrNotFound("contact")
	}
	if errors.Is(err, domain.ErrSegmentNotFound) {
		return ErrNotFound("segment")
	}
	if errors.Is(err, domain.ErrNoteNotFound) {
		return ErrNotFound("note")
	}
	if errors.Is(err, domain.ErrActivityNotFound) {
		return ErrNotFound("activity")
	}
	if errors.Is(err, domain.ErrCustomerAlreadyExists) {
		return ErrConflict("customer already exists")
	}
	if errors.Is(err, domain.ErrContactAlreadyExists) {
		return ErrConflict("contact already exists")
	}
	if errors.Is(err, domain.ErrDuplicateCustomerCode) {
		return ErrConflict("customer code already exists")
	}
	if errors.Is(err, domain.ErrDuplicateCustomerEmail) {
		return ErrConflict("customer email already exists")
	}
	if errors.Is(err, domain.ErrVersionConflict) {
		return ErrConflict("concurrent modification detected, please retry")
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		return ErrUnauthorized("unauthorized")
	}
	if errors.Is(err, domain.ErrForbidden) {
		return ErrForbidden("forbidden")
	}

	// Default to internal server error
	return ErrInternalServer("an unexpected error occurred")
}

// mapAppError maps application errors to HTTP errors.
func mapAppError(err *application.AppError) *ErrorResponse {
	switch err.Code {
	case application.ErrCodeInvalidInput:
		return ErrBadRequest(err.Message)
	case application.ErrCodeNotFound:
		return ErrNotFound(err.Message)
	case application.ErrCodeAlreadyExists:
		return ErrConflict(err.Message)
	case application.ErrCodeValidationFailed:
		details := make(map[string]string)
		if err.Details != nil {
			if d, ok := err.Details.(map[string]string); ok {
				details = d
			}
		}
		return ErrValidation(err.Message, details)
	case application.ErrCodeTenantMismatch:
		return ErrForbidden(err.Message)
	case application.ErrCodeBusinessRule:
		return ErrBadRequest(err.Message)
	case application.ErrCodeInternalError:
		return ErrInternalServer(err.Message)
	default:
		return ErrInternalServer("an unexpected error occurred")
	}
}

// mapDomainError maps domain errors to HTTP errors.
func mapDomainError(err *domain.DomainError) *ErrorResponse {
	switch err.Code {
	case domain.ErrCodeCustomerNotFound, domain.ErrCodeContactNotFound:
		return ErrNotFound(err.Message)
	case domain.ErrCodeCustomerAlreadyExists, domain.ErrCodeContactAlreadyExists:
		return ErrConflict(err.Message)
	case domain.ErrCodeDuplicateCode, domain.ErrCodeDuplicateEmail:
		return ErrConflict(err.Message)
	case domain.ErrCodeVersionMismatch:
		return ErrConflict(err.Message)
	case domain.ErrCodeInvalidCustomerData, domain.ErrCodeInvalidContactData:
		return ErrBadRequest(err.Message)
	case domain.ErrCodeValidationFailed:
		return ErrBadRequest(err.Message)
	case domain.ErrCodeUnauthorized:
		return ErrUnauthorized(err.Message)
	case domain.ErrCodeForbidden:
		return ErrForbidden(err.Message)
	default:
		return ErrInternalServer("an unexpected error occurred")
	}
}
