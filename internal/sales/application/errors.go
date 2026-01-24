package application

import (
	"fmt"
)

// ============================================================================
// Error Codes
// ============================================================================

// ErrorCode represents an application error code.
type ErrorCode string

const (
	// General errors
	ErrCodeInternal           ErrorCode = "INTERNAL_ERROR"
	ErrCodeValidation         ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists      ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict           ErrorCode = "CONFLICT"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeRateLimited        ErrorCode = "RATE_LIMITED"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// Lead errors
	ErrCodeLeadNotFound              ErrorCode = "LEAD_NOT_FOUND"
	ErrCodeLeadAlreadyExists         ErrorCode = "LEAD_ALREADY_EXISTS"
	ErrCodeLeadAlreadyConverted      ErrorCode = "LEAD_ALREADY_CONVERTED"
	ErrCodeLeadNotQualified          ErrorCode = "LEAD_NOT_QUALIFIED"
	ErrCodeLeadAlreadyQualified      ErrorCode = "LEAD_ALREADY_QUALIFIED"
	ErrCodeLeadAlreadyDisqualified   ErrorCode = "LEAD_ALREADY_DISQUALIFIED"
	ErrCodeLeadInvalidStatus         ErrorCode = "LEAD_INVALID_STATUS"
	ErrCodeLeadInvalidTransition     ErrorCode = "LEAD_INVALID_STATUS_TRANSITION"
	ErrCodeLeadDuplicateEmail        ErrorCode = "LEAD_DUPLICATE_EMAIL"
	ErrCodeLeadAssignmentFailed      ErrorCode = "LEAD_ASSIGNMENT_FAILED"
	ErrCodeLeadConversionFailed      ErrorCode = "LEAD_CONVERSION_FAILED"
	ErrCodeLeadScoringFailed         ErrorCode = "LEAD_SCORING_FAILED"

	// Opportunity errors
	ErrCodeOpportunityNotFound            ErrorCode = "OPPORTUNITY_NOT_FOUND"
	ErrCodeOpportunityAlreadyExists       ErrorCode = "OPPORTUNITY_ALREADY_EXISTS"
	ErrCodeOpportunityClosed              ErrorCode = "OPPORTUNITY_CLOSED"
	ErrCodeOpportunityAlreadyWon          ErrorCode = "OPPORTUNITY_ALREADY_WON"
	ErrCodeOpportunityAlreadyLost         ErrorCode = "OPPORTUNITY_ALREADY_LOST"
	ErrCodeOpportunityInvalidStatus       ErrorCode = "OPPORTUNITY_INVALID_STATUS"
	ErrCodeOpportunityInvalidTransition   ErrorCode = "OPPORTUNITY_INVALID_STAGE_TRANSITION"
	ErrCodeOpportunityStageNotFound       ErrorCode = "OPPORTUNITY_STAGE_NOT_FOUND"
	ErrCodeOpportunityProductNotFound     ErrorCode = "OPPORTUNITY_PRODUCT_NOT_FOUND"
	ErrCodeOpportunityContactNotFound     ErrorCode = "OPPORTUNITY_CONTACT_NOT_FOUND"
	ErrCodeOpportunityContactDuplicate    ErrorCode = "OPPORTUNITY_CONTACT_DUPLICATE"
	ErrCodeOpportunityProductDuplicate    ErrorCode = "OPPORTUNITY_PRODUCT_DUPLICATE"
	ErrCodeOpportunityAssignmentFailed    ErrorCode = "OPPORTUNITY_ASSIGNMENT_FAILED"
	ErrCodeOpportunityWinFailed           ErrorCode = "OPPORTUNITY_WIN_FAILED"
	ErrCodeOpportunityLoseFailed          ErrorCode = "OPPORTUNITY_LOSE_FAILED"

	// Deal errors
	ErrCodeDealNotFound              ErrorCode = "DEAL_NOT_FOUND"
	ErrCodeDealAlreadyExists         ErrorCode = "DEAL_ALREADY_EXISTS"
	ErrCodeDealCancelled             ErrorCode = "DEAL_CANCELLED"
	ErrCodeDealCompleted             ErrorCode = "DEAL_COMPLETED"
	ErrCodeDealInvalidStatus         ErrorCode = "DEAL_INVALID_STATUS"
	ErrCodeDealInvalidTransition     ErrorCode = "DEAL_INVALID_STATUS_TRANSITION"
	ErrCodeDealLineItemNotFound      ErrorCode = "DEAL_LINE_ITEM_NOT_FOUND"
	ErrCodeDealInvoiceNotFound       ErrorCode = "DEAL_INVOICE_NOT_FOUND"
	ErrCodeDealPaymentNotFound       ErrorCode = "DEAL_PAYMENT_NOT_FOUND"
	ErrCodeDealPaymentExceedsBalance ErrorCode = "DEAL_PAYMENT_EXCEEDS_BALANCE"
	ErrCodeDealFulfillmentExceeds    ErrorCode = "DEAL_FULFILLMENT_EXCEEDS_QUANTITY"
	ErrCodeDealInvoiceAlreadyPaid    ErrorCode = "DEAL_INVOICE_ALREADY_PAID"
	ErrCodeDealCannotCancel          ErrorCode = "DEAL_CANNOT_CANCEL"
	ErrCodeDealNumberGeneration      ErrorCode = "DEAL_NUMBER_GENERATION_FAILED"

	// Pipeline errors
	ErrCodePipelineNotFound          ErrorCode = "PIPELINE_NOT_FOUND"
	ErrCodePipelineAlreadyExists     ErrorCode = "PIPELINE_ALREADY_EXISTS"
	ErrCodePipelineInactive          ErrorCode = "PIPELINE_INACTIVE"
	ErrCodePipelineHasOpportunities  ErrorCode = "PIPELINE_HAS_OPPORTUNITIES"
	ErrCodePipelineCannotDelete      ErrorCode = "PIPELINE_CANNOT_DELETE"
	ErrCodePipelineDefaultRequired   ErrorCode = "PIPELINE_DEFAULT_REQUIRED"
	ErrCodePipelineStageNotFound     ErrorCode = "PIPELINE_STAGE_NOT_FOUND"
	ErrCodePipelineStageInactive     ErrorCode = "PIPELINE_STAGE_INACTIVE"
	ErrCodePipelineStageDuplicate    ErrorCode = "PIPELINE_STAGE_DUPLICATE"
	ErrCodePipelineStageHasOpportunities ErrorCode = "PIPELINE_STAGE_HAS_OPPORTUNITIES"
	ErrCodePipelineStageCannotDelete ErrorCode = "PIPELINE_STAGE_CANNOT_DELETE"
	ErrCodePipelineInvalidStageOrder ErrorCode = "PIPELINE_INVALID_STAGE_ORDER"
	ErrCodePipelineMinStagesRequired ErrorCode = "PIPELINE_MIN_STAGES_REQUIRED"
	ErrCodePipelineWonStageRequired  ErrorCode = "PIPELINE_WON_STAGE_REQUIRED"
	ErrCodePipelineLostStageRequired ErrorCode = "PIPELINE_LOST_STAGE_REQUIRED"

	// Customer/Contact errors
	ErrCodeCustomerNotFound          ErrorCode = "CUSTOMER_NOT_FOUND"
	ErrCodeContactNotFound           ErrorCode = "CONTACT_NOT_FOUND"
	ErrCodeCustomerServiceError      ErrorCode = "CUSTOMER_SERVICE_ERROR"

	// Product errors
	ErrCodeProductNotFound           ErrorCode = "PRODUCT_NOT_FOUND"
	ErrCodeProductServiceError       ErrorCode = "PRODUCT_SERVICE_ERROR"

	// User errors
	ErrCodeUserNotFound              ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserServiceError          ErrorCode = "USER_SERVICE_ERROR"

	// Currency errors
	ErrCodeCurrencyMismatch          ErrorCode = "CURRENCY_MISMATCH"
	ErrCodeCurrencyInvalid           ErrorCode = "CURRENCY_INVALID"

	// Version/Concurrency errors
	ErrCodeVersionMismatch           ErrorCode = "VERSION_MISMATCH"
	ErrCodeConcurrentModification    ErrorCode = "CONCURRENT_MODIFICATION"

	// External service errors
	ErrCodeEventPublishFailed        ErrorCode = "EVENT_PUBLISH_FAILED"
	ErrCodeCacheError                ErrorCode = "CACHE_ERROR"
	ErrCodeSearchIndexError          ErrorCode = "SEARCH_INDEX_ERROR"
	ErrCodeNotificationError         ErrorCode = "NOTIFICATION_ERROR"
)

// ============================================================================
// Application Error
// ============================================================================

// AppError represents an application-level error.
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	StackTrace string                 `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithDetail adds a detail to the error.
func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithCause sets the underlying cause.
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// ============================================================================
// Error Constructors
// ============================================================================

// NewAppError creates a new application error.
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewAppErrorf creates a new application error with formatted message.
func NewAppErrorf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// WrapError wraps an error with an application error.
func WrapError(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// ============================================================================
// Common Error Constructors
// ============================================================================

// Internal errors
func ErrInternal(message string, cause error) *AppError {
	return WrapError(ErrCodeInternal, message, cause)
}

func ErrValidation(message string) *AppError {
	return NewAppError(ErrCodeValidation, message)
}

func ErrValidationWithDetails(message string, details map[string]interface{}) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

func ErrNotFound(entityType string, id interface{}) *AppError {
	return NewAppErrorf(ErrCodeNotFound, "%s not found: %v", entityType, id)
}

func ErrAlreadyExists(entityType string, identifier interface{}) *AppError {
	return NewAppErrorf(ErrCodeAlreadyExists, "%s already exists: %v", entityType, identifier)
}

func ErrConflict(message string) *AppError {
	return NewAppError(ErrCodeConflict, message)
}

func ErrUnauthorized(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, message)
}

func ErrForbidden(message string) *AppError {
	return NewAppError(ErrCodeForbidden, message)
}

func ErrRateLimited(message string) *AppError {
	return NewAppError(ErrCodeRateLimited, message)
}

func ErrServiceUnavailable(service string) *AppError {
	return NewAppErrorf(ErrCodeServiceUnavailable, "%s service is unavailable", service)
}

// Lead errors
func ErrLeadNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeLeadNotFound, "lead not found: %v", id)
}

func ErrLeadAlreadyConverted(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeLeadAlreadyConverted, "lead already converted: %v", id)
}

func ErrLeadNotQualified(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeLeadNotQualified, "lead is not qualified: %v", id)
}

func ErrLeadAlreadyQualified(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeLeadAlreadyQualified, "lead is already qualified: %v", id)
}

func ErrLeadInvalidStatusTransition(from, to string) *AppError {
	return NewAppErrorf(ErrCodeLeadInvalidTransition, "invalid lead status transition from %s to %s", from, to)
}

func ErrLeadDuplicateEmail(email string) *AppError {
	return NewAppErrorf(ErrCodeLeadDuplicateEmail, "lead with email already exists: %s", email)
}

func ErrLeadConversionFailed(id interface{}, reason string) *AppError {
	return NewAppErrorf(ErrCodeLeadConversionFailed, "failed to convert lead %v: %s", id, reason)
}

// Opportunity errors
func ErrOpportunityNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeOpportunityNotFound, "opportunity not found: %v", id)
}

func ErrOpportunityClosed(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeOpportunityClosed, "opportunity is closed: %v", id)
}

func ErrOpportunityAlreadyWon(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeOpportunityAlreadyWon, "opportunity is already won: %v", id)
}

func ErrOpportunityAlreadyLost(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeOpportunityAlreadyLost, "opportunity is already lost: %v", id)
}

func ErrOpportunityInvalidStageTransition(from, to string) *AppError {
	return NewAppErrorf(ErrCodeOpportunityInvalidTransition, "invalid stage transition from %s to %s", from, to)
}

func ErrOpportunityStageNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeOpportunityStageNotFound, "stage not found: %v", id)
}

func ErrOpportunityProductNotFound(opportunityID, productID interface{}) *AppError {
	return NewAppErrorf(ErrCodeOpportunityProductNotFound, "product %v not found in opportunity %v", productID, opportunityID)
}

func ErrOpportunityContactNotFound(opportunityID, contactID interface{}) *AppError {
	return NewAppErrorf(ErrCodeOpportunityContactNotFound, "contact %v not found in opportunity %v", contactID, opportunityID)
}

func ErrOpportunityContactDuplicate(opportunityID, contactID interface{}) *AppError {
	return NewAppErrorf(ErrCodeOpportunityContactDuplicate, "contact %v already exists in opportunity %v", contactID, opportunityID)
}

// Deal errors
func ErrDealNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeDealNotFound, "deal not found: %v", id)
}

func ErrDealCancelled(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeDealCancelled, "deal is cancelled: %v", id)
}

func ErrDealCompleted(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeDealCompleted, "deal is already completed: %v", id)
}

func ErrDealInvalidStatusTransition(from, to string) *AppError {
	return NewAppErrorf(ErrCodeDealInvalidTransition, "invalid deal status transition from %s to %s", from, to)
}

func ErrDealLineItemNotFound(dealID, lineItemID interface{}) *AppError {
	return NewAppErrorf(ErrCodeDealLineItemNotFound, "line item %v not found in deal %v", lineItemID, dealID)
}

func ErrDealPaymentExceedsBalance(amount, balance int64) *AppError {
	return NewAppErrorf(ErrCodeDealPaymentExceedsBalance, "payment amount %d exceeds remaining balance %d", amount, balance)
}

func ErrDealFulfillmentExceedsQuantity(fulfilled, remaining int) *AppError {
	return NewAppErrorf(ErrCodeDealFulfillmentExceeds, "fulfillment quantity %d exceeds remaining quantity %d", fulfilled, remaining)
}

func ErrDealCannotCancel(id interface{}, reason string) *AppError {
	return NewAppErrorf(ErrCodeDealCannotCancel, "cannot cancel deal %v: %s", id, reason)
}

// Pipeline errors
func ErrPipelineNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodePipelineNotFound, "pipeline not found: %v", id)
}

func ErrPipelineInactive(id interface{}) *AppError {
	return NewAppErrorf(ErrCodePipelineInactive, "pipeline is inactive: %v", id)
}

func ErrPipelineHasOpportunities(id interface{}) *AppError {
	return NewAppErrorf(ErrCodePipelineHasOpportunities, "cannot delete pipeline %v: has associated opportunities", id)
}

func ErrPipelineDefaultRequired() *AppError {
	return NewAppError(ErrCodePipelineDefaultRequired, "at least one default pipeline is required")
}

func ErrPipelineStageNotFound(pipelineID, stageID interface{}) *AppError {
	return NewAppErrorf(ErrCodePipelineStageNotFound, "stage %v not found in pipeline %v", stageID, pipelineID)
}

func ErrPipelineStageInactive(stageID interface{}) *AppError {
	return NewAppErrorf(ErrCodePipelineStageInactive, "stage is inactive: %v", stageID)
}

func ErrPipelineStageHasOpportunities(stageID interface{}) *AppError {
	return NewAppErrorf(ErrCodePipelineStageHasOpportunities, "cannot delete stage %v: has associated opportunities", stageID)
}

func ErrPipelineMinStagesRequired(min int) *AppError {
	return NewAppErrorf(ErrCodePipelineMinStagesRequired, "pipeline requires at least %d stages", min)
}

func ErrPipelineWonStageRequired() *AppError {
	return NewAppError(ErrCodePipelineWonStageRequired, "pipeline requires at least one 'won' stage")
}

func ErrPipelineLostStageRequired() *AppError {
	return NewAppError(ErrCodePipelineLostStageRequired, "pipeline requires at least one 'lost' stage")
}

// Currency errors
func ErrCurrencyMismatch(expected, actual string) *AppError {
	return NewAppErrorf(ErrCodeCurrencyMismatch, "currency mismatch: expected %s, got %s", expected, actual)
}

func ErrCurrencyInvalid(currency string) *AppError {
	return NewAppErrorf(ErrCodeCurrencyInvalid, "invalid currency: %s", currency)
}

// Version errors
func ErrVersionMismatch(expected, actual int) *AppError {
	return NewAppErrorf(ErrCodeVersionMismatch, "version mismatch: expected %d, got %d", expected, actual)
}

func ErrConcurrentModification(entityType string, id interface{}) *AppError {
	return NewAppErrorf(ErrCodeConcurrentModification, "concurrent modification detected for %s: %v", entityType, id)
}

// External service errors
func ErrCustomerNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeCustomerNotFound, "customer not found: %v", id)
}

func ErrContactNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeContactNotFound, "contact not found: %v", id)
}

func ErrProductNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeProductNotFound, "product not found: %v", id)
}

func ErrUserNotFound(id interface{}) *AppError {
	return NewAppErrorf(ErrCodeUserNotFound, "user not found: %v", id)
}

func ErrEventPublishFailed(eventType string, cause error) *AppError {
	return WrapError(ErrCodeEventPublishFailed, fmt.Sprintf("failed to publish event: %s", eventType), cause)
}

func ErrCacheError(operation string, cause error) *AppError {
	return WrapError(ErrCodeCacheError, fmt.Sprintf("cache error during %s", operation), cause)
}

// ============================================================================
// Error Type Checking
// ============================================================================

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts an AppError from an error.
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// IsNotFoundError checks if an error is a not found error.
func IsNotFoundError(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		switch appErr.Code {
		case ErrCodeNotFound,
			ErrCodeLeadNotFound,
			ErrCodeOpportunityNotFound,
			ErrCodeDealNotFound,
			ErrCodePipelineNotFound,
			ErrCodePipelineStageNotFound,
			ErrCodeCustomerNotFound,
			ErrCodeContactNotFound,
			ErrCodeProductNotFound,
			ErrCodeUserNotFound,
			ErrCodeDealLineItemNotFound,
			ErrCodeDealInvoiceNotFound,
			ErrCodeDealPaymentNotFound,
			ErrCodeOpportunityProductNotFound,
			ErrCodeOpportunityContactNotFound:
			return true
		}
	}
	return false
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == ErrCodeValidation
	}
	return false
}

// IsConflictError checks if an error is a conflict error.
func IsConflictError(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		switch appErr.Code {
		case ErrCodeConflict,
			ErrCodeAlreadyExists,
			ErrCodeVersionMismatch,
			ErrCodeConcurrentModification,
			ErrCodeLeadAlreadyExists,
			ErrCodeLeadAlreadyConverted,
			ErrCodeLeadDuplicateEmail,
			ErrCodeOpportunityAlreadyExists,
			ErrCodeOpportunityContactDuplicate,
			ErrCodeOpportunityProductDuplicate,
			ErrCodeDealAlreadyExists,
			ErrCodePipelineAlreadyExists,
			ErrCodePipelineStageDuplicate:
			return true
		}
	}
	return false
}

// IsUnauthorizedError checks if an error is an unauthorized error.
func IsUnauthorizedError(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == ErrCodeUnauthorized
	}
	return false
}

// IsForbiddenError checks if an error is a forbidden error.
func IsForbiddenError(err error) bool {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code == ErrCodeForbidden
	}
	return false
}
