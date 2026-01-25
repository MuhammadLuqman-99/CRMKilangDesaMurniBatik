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
	var appErr *application.AppError
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
	if errors.Is(err, domain.ErrOpportunityAlreadyClosed) {
		return ErrConflict("opportunity is already closed")
	}
	if errors.Is(err, domain.ErrOpportunityNotClosed) {
		return ErrBadRequest("opportunity is not closed")
	}
	if errors.Is(err, domain.ErrInvalidOpportunityStatus) {
		return ErrBadRequest("invalid opportunity status")
	}
	if errors.Is(err, domain.ErrInvalidStageTransition) {
		return ErrBadRequest("invalid stage transition")
	}
	if errors.Is(err, domain.ErrOpportunityVersionMismatch) {
		return ErrConflict("opportunity version mismatch")
	}
	if errors.Is(err, domain.ErrCannotReopenAfterDays) {
		return ErrConflict("cannot reopen opportunity after 30 days")
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
	if errors.Is(err, domain.ErrDealAlreadyClosed) {
		return ErrConflict("deal is already closed")
	}
	if errors.Is(err, domain.ErrDealAlreadyFulfilled) {
		return ErrConflict("deal is already fulfilled")
	}
	if errors.Is(err, domain.ErrDealCannotBeCancelled) {
		return ErrConflict("deal cannot be cancelled")
	}
	if errors.Is(err, domain.ErrDealVersionMismatch) {
		return ErrConflict("deal version mismatch")
	}
	if errors.Is(err, domain.ErrInvalidPaymentTerm) {
		return ErrBadRequest("invalid payment term")
	}
	if errors.Is(err, domain.ErrInvoiceAlreadyExists) {
		return ErrConflict("invoice already exists")
	}
	if errors.Is(err, domain.ErrPaymentExceedsBalance) {
		return ErrBadRequest("payment amount exceeds remaining balance")
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
func mapAppError(err *application.AppError) *ErrorResponse {
	switch err.Code {
	// Validation errors
	case application.ErrCodeValidation:
		details := make(map[string]string)
		for k, v := range err.Details {
			if s, ok := v.(string); ok {
				details[k] = s
			}
		}
		return ErrValidation(err.Message, details)

	// Not found errors
	case application.ErrCodeNotFound,
		application.ErrCodeLeadNotFound,
		application.ErrCodeOpportunityNotFound,
		application.ErrCodeDealNotFound,
		application.ErrCodePipelineNotFound,
		application.ErrCodePipelineStageNotFound,
		application.ErrCodeProductNotFound,
		application.ErrCodeContactNotFound,
		application.ErrCodeCustomerNotFound,
		application.ErrCodeDealInvoiceNotFound,
		application.ErrCodeDealPaymentNotFound,
		application.ErrCodeDealLineItemNotFound,
		application.ErrCodeOpportunityProductNotFound,
		application.ErrCodeOpportunityContactNotFound,
		application.ErrCodeUserNotFound:
		return ErrNotFound(err.Message)

	// Conflict/Already exists errors
	case application.ErrCodeConflict,
		application.ErrCodeAlreadyExists,
		application.ErrCodeLeadAlreadyExists,
		application.ErrCodeLeadAlreadyConverted,
		application.ErrCodeLeadAlreadyQualified,
		application.ErrCodeLeadAlreadyDisqualified,
		application.ErrCodeOpportunityAlreadyExists,
		application.ErrCodeOpportunityClosed,
		application.ErrCodeOpportunityAlreadyWon,
		application.ErrCodeOpportunityAlreadyLost,
		application.ErrCodeDealAlreadyExists,
		application.ErrCodeDealCompleted,
		application.ErrCodeDealCancelled,
		application.ErrCodePipelineAlreadyExists,
		application.ErrCodePipelineStageDuplicate,
		application.ErrCodeOpportunityContactDuplicate,
		application.ErrCodeOpportunityProductDuplicate,
		application.ErrCodeVersionMismatch,
		application.ErrCodeConcurrentModification:
		return ErrConflict(err.Message)

	// Business rule violations
	case application.ErrCodeLeadNotQualified,
		application.ErrCodeLeadInvalidStatus,
		application.ErrCodeLeadInvalidTransition,
		application.ErrCodeOpportunityInvalidStatus,
		application.ErrCodeOpportunityInvalidTransition,
		application.ErrCodeDealInvalidStatus,
		application.ErrCodeDealInvalidTransition,
		application.ErrCodeDealPaymentExceedsBalance,
		application.ErrCodeDealFulfillmentExceeds,
		application.ErrCodeDealCannotCancel,
		application.ErrCodePipelineInactive,
		application.ErrCodePipelineStageInactive,
		application.ErrCodePipelineHasOpportunities,
		application.ErrCodePipelineStageHasOpportunities,
		application.ErrCodePipelineMinStagesRequired,
		application.ErrCodePipelineWonStageRequired,
		application.ErrCodePipelineLostStageRequired,
		application.ErrCodePipelineDefaultRequired,
		application.ErrCodePipelineCannotDelete,
		application.ErrCodePipelineStageCannotDelete,
		application.ErrCodePipelineInvalidStageOrder,
		application.ErrCodeCurrencyMismatch,
		application.ErrCodeCurrencyInvalid:
		return ErrUnprocessableEntity(err.Message)

	// Authorization errors
	case application.ErrCodeUnauthorized:
		return ErrUnauthorized(err.Message)
	case application.ErrCodeForbidden:
		return ErrForbidden(err.Message)

	// Rate limiting
	case application.ErrCodeRateLimited:
		return &ErrorResponse{
			StatusCode: http.StatusTooManyRequests,
			Code:       "RATE_LIMITED",
			Message:    err.Message,
		}

	// External service errors
	case application.ErrCodeServiceUnavailable,
		application.ErrCodeCustomerServiceError,
		application.ErrCodeProductServiceError,
		application.ErrCodeUserServiceError:
		return ErrServiceUnavailable(err.Message)

	// Internal errors
	case application.ErrCodeInternal,
		application.ErrCodeEventPublishFailed,
		application.ErrCodeCacheError,
		application.ErrCodeSearchIndexError,
		application.ErrCodeNotificationError,
		application.ErrCodeLeadAssignmentFailed,
		application.ErrCodeLeadConversionFailed,
		application.ErrCodeLeadScoringFailed,
		application.ErrCodeOpportunityAssignmentFailed,
		application.ErrCodeOpportunityWinFailed,
		application.ErrCodeOpportunityLoseFailed,
		application.ErrCodeDealNumberGeneration:
		return ErrInternalServer("an unexpected error occurred")

	default:
		return ErrInternalServer("an unexpected error occurred")
	}
}
