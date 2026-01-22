// Package validator provides request validation utilities for the CRM application.
package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/kilang-desa-murni/crm/pkg/errors"
)

// Validator wraps the go-playground validator.
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator instance.
func New() *Validator {
	v := validator.New()

	// Register custom tag name function to use JSON tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validations
	registerCustomValidations(v)

	return &Validator{validate: v}
}

// Validate validates a struct and returns an error with field-level details.
func (v *Validator) Validate(s interface{}) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.Wrap(err, errors.ErrCodeValidation, "validation failed")
	}

	appErr := errors.New(errors.ErrCodeValidation, "Validation failed")

	for _, e := range validationErrors {
		field := e.Field()
		message := formatValidationError(e)
		appErr.WithField(field, message)
	}

	return appErr
}

// ValidateVar validates a single variable.
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.Wrap(err, errors.ErrCodeValidation, "validation failed")
	}

	if len(validationErrors) > 0 {
		return errors.New(errors.ErrCodeValidation, formatValidationError(validationErrors[0]))
	}

	return nil
}

// DecodeAndValidate decodes JSON from request body and validates the struct.
func (v *Validator) DecodeAndValidate(r *http.Request, dst interface{}) error {
	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return errors.Wrap(err, errors.ErrCodeBadRequest, "Invalid JSON body")
	}

	// Validate
	return v.Validate(dst)
}

// registerCustomValidations registers custom validation functions.
func registerCustomValidations(v *validator.Validate) {
	// Phone number validation
	v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Basic phone validation - adjust regex as needed
		match, _ := regexp.MatchString(`^[+]?[\d\s-]{10,20}$`, phone)
		return match
	})

	// UUID validation
	v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		uuid := fl.Field().String()
		match, _ := regexp.MatchString(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, uuid)
		return match
	})

	// Slug validation
	v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		slug := fl.Field().String()
		match, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
		return match
	})

	// Safe string validation (no HTML/scripts)
	v.RegisterValidation("safestring", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		// Check for common XSS patterns
		dangerousPatterns := []string{"<script", "javascript:", "onclick", "onerror", "onload"}
		lowerStr := strings.ToLower(str)
		for _, pattern := range dangerousPatterns {
			if strings.Contains(lowerStr, pattern) {
				return false
			}
		}
		return true
	})

	// Password strength validation
	v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}

		var hasUpper, hasLower, hasDigit, hasSpecial bool
		for _, char := range password {
			switch {
			case 'A' <= char && char <= 'Z':
				hasUpper = true
			case 'a' <= char && char <= 'z':
				hasLower = true
			case '0' <= char && char <= '9':
				hasDigit = true
			case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
				hasSpecial = true
			}
		}

		return hasUpper && hasLower && hasDigit && hasSpecial
	})

	// Money amount validation (positive with max 2 decimal places)
	v.RegisterValidation("money", func(fl validator.FieldLevel) bool {
		value := fl.Field().Float()
		if value < 0 {
			return false
		}
		// Check for max 2 decimal places
		str := fmt.Sprintf("%.2f", value)
		parsed := 0.0
		fmt.Sscanf(str, "%f", &parsed)
		return parsed == value
	})

	// Percentage validation (0-100)
	v.RegisterValidation("percentage", func(fl validator.FieldLevel) bool {
		value := fl.Field().Float()
		return value >= 0 && value <= 100
	})
}

// formatValidationError formats a validation error into a human-readable message.
func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address"
	case "min":
		if e.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must be at least %s characters", e.Param())
		}
		return fmt.Sprintf("Must be at least %s", e.Param())
	case "max":
		if e.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must be at most %s characters", e.Param())
		}
		return fmt.Sprintf("Must be at most %s", e.Param())
	case "len":
		return fmt.Sprintf("Must be exactly %s characters", e.Param())
	case "eq":
		return fmt.Sprintf("Must be equal to %s", e.Param())
	case "ne":
		return fmt.Sprintf("Must not be equal to %s", e.Param())
	case "gt":
		return fmt.Sprintf("Must be greater than %s", e.Param())
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", e.Param())
	case "lt":
		return fmt.Sprintf("Must be less than %s", e.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", e.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", e.Param())
	case "uuid":
		return "Invalid UUID format"
	case "url":
		return "Invalid URL format"
	case "phone":
		return "Invalid phone number format"
	case "slug":
		return "Invalid slug format (lowercase letters, numbers, and hyphens only)"
	case "safestring":
		return "Contains potentially unsafe content"
	case "strongpassword":
		return "Password must be at least 8 characters with uppercase, lowercase, digit, and special character"
	case "money":
		return "Invalid money amount (must be positive with max 2 decimal places)"
	case "percentage":
		return "Must be a percentage between 0 and 100"
	case "alpha":
		return "Must contain only letters"
	case "alphanum":
		return "Must contain only letters and numbers"
	case "numeric":
		return "Must be a number"
	case "boolean":
		return "Must be true or false"
	case "datetime":
		return "Invalid datetime format"
	default:
		return fmt.Sprintf("Failed validation: %s", e.Tag())
	}
}

// Global validator instance
var globalValidator = New()

// Validate validates a struct using the global validator.
func Validate(s interface{}) error {
	return globalValidator.Validate(s)
}

// ValidateVar validates a variable using the global validator.
func ValidateVar(field interface{}, tag string) error {
	return globalValidator.ValidateVar(field, tag)
}

// DecodeAndValidate decodes and validates using the global validator.
func DecodeAndValidate(r *http.Request, dst interface{}) error {
	return globalValidator.DecodeAndValidate(r, dst)
}
