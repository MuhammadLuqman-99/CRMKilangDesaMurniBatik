// Package domain contains the domain layer for the Customer service.
package domain

import (
	"errors"
	"fmt"
)

// Domain errors for Customer service.
var (
	// Customer errors
	ErrCustomerNotFound          = errors.New("customer not found")
	ErrCustomerAlreadyExists     = errors.New("customer already exists")
	ErrCustomerDeleted           = errors.New("customer is deleted")
	ErrCustomerInactive          = errors.New("customer is inactive")
	ErrCustomerBlocked           = errors.New("customer is blocked")
	ErrInvalidCustomerStatus     = errors.New("invalid customer status")
	ErrInvalidCustomerType       = errors.New("invalid customer type")
	ErrInvalidCustomerData       = errors.New("invalid customer data")
	ErrCustomerVersionMismatch   = errors.New("customer version mismatch (concurrent modification)")
	ErrDuplicateCustomerCode     = errors.New("customer code already exists")
	ErrDuplicateCustomerEmail    = errors.New("customer email already exists")

	// Contact errors
	ErrContactNotFound           = errors.New("contact not found")
	ErrContactAlreadyExists      = errors.New("contact already exists")
	ErrContactDeleted            = errors.New("contact is deleted")
	ErrInvalidContactData        = errors.New("invalid contact data")
	ErrPrimaryContactRequired    = errors.New("at least one primary contact is required")
	ErrCannotDeletePrimaryContact = errors.New("cannot delete primary contact when other contacts exist")
	ErrDuplicateContactEmail     = errors.New("contact email already exists for this customer")
	ErrMaxContactsExceeded       = errors.New("maximum number of contacts exceeded")

	// Value object errors
	ErrInvalidEmail              = errors.New("invalid email address")
	ErrInvalidPhoneNumber        = errors.New("invalid phone number")
	ErrInvalidAddress            = errors.New("invalid address")
	ErrInvalidSocialProfile      = errors.New("invalid social profile")
	ErrInvalidURL                = errors.New("invalid URL")
	ErrInvalidCountryCode        = errors.New("invalid country code")
	ErrInvalidPostalCode         = errors.New("invalid postal code")
	ErrInvalidCurrency           = errors.New("invalid currency code")

	// Segment errors
	ErrSegmentNotFound           = errors.New("segment not found")
	ErrSegmentAlreadyExists      = errors.New("segment already exists")
	ErrInvalidSegmentCriteria    = errors.New("invalid segment criteria")

	// Note errors
	ErrNoteNotFound              = errors.New("note not found")
	ErrInvalidNoteContent        = errors.New("invalid note content")

	// Activity errors
	ErrActivityNotFound          = errors.New("activity not found")
	ErrInvalidActivityType       = errors.New("invalid activity type")

	// Import/Export errors
	ErrImportFailed              = errors.New("import failed")
	ErrImportNotFound            = errors.New("import not found")
	ErrExportFailed              = errors.New("export failed")
	ErrInvalidImportFormat       = errors.New("invalid import format")
	ErrImportValidationFailed    = errors.New("import validation failed")

	// General errors
	ErrInvalidTenantID           = errors.New("invalid tenant ID")
	ErrUnauthorized              = errors.New("unauthorized access")
	ErrForbidden                 = errors.New("forbidden action")
	ErrVersionConflict           = errors.New("version conflict - concurrent modification detected")
)

// ValidationError represents a validation error with field details.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Value   any    `json:"value,omitempty"`
}

// Error returns the error message.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message, code string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	}
}

// NewValidationErrorWithValue creates a new validation error with the invalid value.
func NewValidationErrorWithValue(field, message, code string, value any) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
		Value:   value,
	}
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []*ValidationError

// Error returns a combined error message.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%d validation errors: %s (and %d more)", len(e), e[0].Error(), len(e)-1)
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Add adds a validation error.
func (e *ValidationErrors) Add(err *ValidationError) {
	*e = append(*e, err)
}

// AddField adds a validation error for a field.
func (e *ValidationErrors) AddField(field, message, code string) {
	*e = append(*e, NewValidationError(field, message, code))
}

// GetByField returns all errors for a specific field.
func (e ValidationErrors) GetByField(field string) ValidationErrors {
	var result ValidationErrors
	for _, err := range e {
		if err.Field == field {
			result = append(result, err)
		}
	}
	return result
}

// ToError returns an error if there are validation errors, nil otherwise.
func (e ValidationErrors) ToError() error {
	if len(e) == 0 {
		return nil
	}
	return e
}

// DomainError represents a domain-specific error.
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
	Details any    `json:"details,omitempty"`
}

// Error returns the error message.
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error.
func NewDomainError(code, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewDomainErrorWithDetails creates a new domain error with details.
func NewDomainErrorWithDetails(code, message string, err error, details any) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: details,
	}
}

// Error codes for Customer service.
const (
	ErrCodeCustomerNotFound        = "CUSTOMER_NOT_FOUND"
	ErrCodeCustomerAlreadyExists   = "CUSTOMER_ALREADY_EXISTS"
	ErrCodeCustomerDeleted         = "CUSTOMER_DELETED"
	ErrCodeCustomerInactive        = "CUSTOMER_INACTIVE"
	ErrCodeCustomerBlocked         = "CUSTOMER_BLOCKED"
	ErrCodeInvalidCustomerStatus   = "INVALID_CUSTOMER_STATUS"
	ErrCodeInvalidCustomerType     = "INVALID_CUSTOMER_TYPE"
	ErrCodeInvalidCustomerData     = "INVALID_CUSTOMER_DATA"
	ErrCodeVersionMismatch         = "VERSION_MISMATCH"
	ErrCodeDuplicateCode           = "DUPLICATE_CODE"
	ErrCodeDuplicateEmail          = "DUPLICATE_EMAIL"
	ErrCodeContactNotFound         = "CONTACT_NOT_FOUND"
	ErrCodeContactAlreadyExists    = "CONTACT_ALREADY_EXISTS"
	ErrCodeInvalidContactData      = "INVALID_CONTACT_DATA"
	ErrCodePrimaryContactRequired  = "PRIMARY_CONTACT_REQUIRED"
	ErrCodeMaxContactsExceeded     = "MAX_CONTACTS_EXCEEDED"
	ErrCodeInvalidEmail            = "INVALID_EMAIL"
	ErrCodeInvalidPhoneNumber      = "INVALID_PHONE_NUMBER"
	ErrCodeInvalidAddress          = "INVALID_ADDRESS"
	ErrCodeInvalidURL              = "INVALID_URL"
	ErrCodeImportFailed            = "IMPORT_FAILED"
	ErrCodeExportFailed            = "EXPORT_FAILED"
	ErrCodeUnauthorized            = "UNAUTHORIZED"
	ErrCodeForbidden               = "FORBIDDEN"
	ErrCodeValidationFailed        = "VALIDATION_FAILED"
)

// IsNotFoundError checks if the error is a not found error.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrCustomerNotFound) ||
		errors.Is(err, ErrContactNotFound) ||
		errors.Is(err, ErrSegmentNotFound) ||
		errors.Is(err, ErrNoteNotFound) ||
		errors.Is(err, ErrActivityNotFound)
}

// IsValidationError checks if the error is a validation error.
func IsValidationError(err error) bool {
	_, ok := err.(ValidationErrors)
	if ok {
		return true
	}
	_, ok = err.(*ValidationError)
	return ok
}

// IsDuplicateError checks if the error is a duplicate error.
func IsDuplicateError(err error) bool {
	return errors.Is(err, ErrCustomerAlreadyExists) ||
		errors.Is(err, ErrContactAlreadyExists) ||
		errors.Is(err, ErrDuplicateCustomerCode) ||
		errors.Is(err, ErrDuplicateCustomerEmail) ||
		errors.Is(err, ErrDuplicateContactEmail)
}

// IsConflictError checks if the error is a conflict error.
func IsConflictError(err error) bool {
	return errors.Is(err, ErrCustomerVersionMismatch) ||
		errors.Is(err, ErrPrimaryContactRequired) ||
		errors.Is(err, ErrCannotDeletePrimaryContact)
}
