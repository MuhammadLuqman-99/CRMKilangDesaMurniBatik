// Package http provides HTTP handlers for the Sales Pipeline service.
package http

import (
	"errors"
	"net/http"

	"github.com/kilang-desa-murni/crm/internal/sales/application"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Error Response Structure
// ============================================================================

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	StatusCode int               `json:"-"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	RequestID  string            `json:"request_id,omitempty"`
}

// Error implements the error interface.
func (e *ErrorResponse) Error() string {
	return e.Message
}

// ============================================================================
// HTTP Error Constructors
// ============================================================================

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

// ErrUnprocessableEntity creates an unprocessable entity error.
func ErrUnprocessableEntity(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusUnprocessableEntity,
		Code:       "UNPROCESSABLE_ENTITY",
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

// ErrPreconditionFailed creates a precondition failed error.
func ErrPreconditionFailed(message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: http.StatusPreconditionFailed,
		Code:       "PRECONDITION_FAILED",
		Message:    message,
	}
}

// ============================================================================
// Error Mapping
// ============================================================================

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
	var appErr *application.Error
	if errors.As(err, &appErr) {
		return mapAppError(appErr)
	}

	// Check for domain errors - Lead
	if errors.Is(err, domain.ErrLeadNotFound) {
		return ErrNotFound("lead")
	}
	if errors.Is(err, domain.ErrLeadAlreadyConverted) {
		return ErrConflict("lead has already been converted")
	}
	if errors.Is(err, domain.ErrLeadNotQualified) {
		return ErrBadRequest("lead must be qualified before conversion")
	}
	if errors.Is(err, domain.ErrInvalidLeadStatus) {
		return ErrBadRequest("invalid lead status")
	}
	if errors.Is(err, domain.ErrInvalidLeadSource) {
		return ErrBadRequest("invalid lead source")
	}
	if errors.Is(err, domain.ErrLeadAlreadyQualified) {
		return ErrConflict("lead is already qualified")
	}
	if errors.Is(err, domain.ErrLeadAlreadyDisqualified) {
		return ErrConflict("lead is already disqualified")
	}

	// Check for domain errors - Opportunity
	if errors.Is(err, domain.ErrOpportunityNotFound) {
		return ErrNotFound("opportunity")
	}
	if errors.Is(err, domain.ErrOpportunityClosed) {
		return ErrConflict("opportunity is already closed")
	}
	if errors.Is(err, domain.ErrInvalidOpportunityStatus) {
		return ErrBadRequest("invalid opportunity status")
	}
	if errors.Is(err, domain.ErrOpportunityAlreadyWon) {
		return ErrConflict("opportunity is already won")
	}
	if errors.Is(err, domain.ErrOpportunityAlreadyLost) {
		return ErrConflict("opportunity is already lost")
	}
	if errors.Is(err, domain.ErrInvalidStageTransition) {
		return ErrBadRequest("invalid stage transition")
	}
	if errors.Is(err, domain.ErrProductNotFound) {
		return ErrNotFound("product")
	}
	if errors.Is(err, domain.ErrContactNotInOpportunity) {
		return ErrNotFound("contact not found in opportunity")
	}

	// Check for domain errors - Deal
	if errors.Is(err, domain.ErrDealNotFound) {
		return ErrNotFound("deal")
	}
	if errors.Is(err, domain.ErrDealAlreadyExists) {
		return ErrConflict("deal already exists")
	}
	if errors.Is(err, domain.ErrInvalidDealStatus) {
		return ErrBadRequest("invalid deal status")
	}
	if errors.Is(err, domain.ErrDealNotActive) {
		return ErrConflict("deal is not active")
	}
	if errors.Is(err, domain.ErrDealAlreadyCompleted) {
		return ErrConflict("deal is already completed")
	}
	if errors.Is(err, domain.ErrDealAlreadyCancelled) {
		return ErrConflict("deal is already cancelled")
	}
	if errors.Is(err, domain.ErrInvoiceNotFound) {
		return ErrNotFound("invoice")
	}
	if errors.Is(err, domain.ErrPaymentNotFound) {
		return ErrNotFound("payment")
	}
	if errors.Is(err, domain.ErrPaymentExceedsAmount) {
		return ErrBadRequest("payment amount exceeds remaining balance")
	}
	if errors.Is(err, domain.ErrLineItemNotFound) {
		return ErrNotFound("line item")
	}

	// Check for domain errors - Pipeline
	if errors.Is(err, domain.ErrPipelineNotFound) {
		return ErrNotFound("pipeline")
	}
	if errors.Is(err, domain.ErrPipelineAlreadyExists) {
		return ErrConflict("pipeline already exists")
	}
	if errors.Is(err, domain.ErrPipelineInactive) {
		return ErrConflict("pipeline is inactive")
	}
	if errors.Is(err, domain.ErrPipelineHasOpportunities) {
		return ErrConflict("pipeline has active opportunities")
	}
	if errors.Is(err, domain.ErrStageNotFound) {
		return ErrNotFound("stage")
	}
	if errors.Is(err, domain.ErrStageAlreadyExists) {
		return ErrConflict("stage already exists")
	}
	if errors.Is(err, domain.ErrStageInUse) {
		return ErrConflict("stage is in use")
	}
	if errors.Is(err, domain.ErrInvalidStageOrder) {
		return ErrBadRequest("invalid stage order")
	}
	if errors.Is(err, domain.ErrMinimumStagesRequired) {
		return ErrBadRequest("minimum 2 stages required")
	}
	if errors.Is(err, domain.ErrDefaultPipelineRequired) {
		return ErrConflict("at least one default pipeline is required")
	}
	if errors.Is(err, domain.ErrCannotDeleteDefaultPipeline) {
		return ErrConflict("cannot delete default pipeline")
	}

	// Check for domain errors - Money/Currency
	if errors.Is(err, domain.ErrInvalidCurrency) {
		return ErrBadRequest("invalid currency")
	}
	if errors.Is(err, domain.ErrCurrencyMismatch) {
		return ErrBadRequest("currency mismatch")
	}
	if errors.Is(err, domain.ErrNegativeAmount) {
		return ErrBadRequest("amount cannot be negative")
	}

	// Default to internal server error
	return ErrInternalServer("an unexpected error occurred")
}

// mapAppError maps application errors to HTTP errors.
func mapAppError(err *application.Error) *ErrorResponse {
	switch err.Code {
	// Validation errors
	case application.ErrCodeValidationFailed:
		details := make(map[string]string)
		if d, ok := err.Details.(map[string]string); ok {
			details = d
		}
		return ErrValidation(err.Message, details)
	case application.ErrCodeInvalidInput:
		return ErrBadRequest(err.Message)

	// Not found errors
	case application.ErrCodeLeadNotFound,
		application.ErrCodeOpportunityNotFound,
		application.ErrCodeDealNotFound,
		application.ErrCodePipelineNotFound,
		application.ErrCodeStageNotFound,
		application.ErrCodeProductNotFound,
		application.ErrCodeContactNotFound,
		application.ErrCodeCustomerNotFound,
		application.ErrCodeInvoiceNotFound,
		application.ErrCodePaymentNotFound:
		return ErrNotFound(err.Message)

	// Conflict errors
	case application.ErrCodeLeadAlreadyConverted,
		application.ErrCodeLeadAlreadyQualified,
		application.ErrCodeOpportunityClosed,
		application.ErrCodeOpportunityAlreadyWon,
		application.ErrCodeOpportunityAlreadyLost,
		application.ErrCodeDealAlreadyExists,
		application.ErrCodeDealAlreadyCompleted,
		application.ErrCodeDealAlreadyCancelled,
		application.ErrCodePipelineAlreadyExists,
		application.ErrCodeStageAlreadyExists,
		application.ErrCodeVersionConflict:
		return ErrConflict(err.Message)

	// Business rule violations
	case application.ErrCodeLeadNotQualified,
		application.ErrCodeInvalidStageTransition,
		application.ErrCodeDealNotActive,
		application.ErrCodePaymentExceedsBalance,
		application.ErrCodePipelineInactive,
		application.ErrCodePipelineHasOpportunities,
		application.ErrCodeMinimumStagesRequired,
		application.ErrCodeCannotDeleteDefaultPipeline:
		return ErrUnprocessableEntity(err.Message)

	// Authorization errors
	case application.ErrCodeUnauthorized:
		return ErrUnauthorized(err.Message)
	case application.ErrCodeForbidden,
		application.ErrCodeTenantMismatch:
		return ErrForbidden(err.Message)

	// External service errors
	case application.ErrCodeExternalServiceError:
		return ErrServiceUnavailable(err.Message)

	// Internal errors
	case application.ErrCodeInternalError,
		application.ErrCodeDatabaseError:
		return ErrInternalServer("an unexpected error occurred")

	default:
		return ErrInternalServer("an unexpected error occurred")
	}
}
