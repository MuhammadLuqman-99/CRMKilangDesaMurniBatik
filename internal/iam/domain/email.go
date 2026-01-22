// Package domain contains the domain layer for the IAM service.
package domain

import (
	"fmt"
	"net/mail"
	"strings"
)

// Email represents an email address value object.
// It ensures the email is always valid and normalized (lowercase).
type Email struct {
	value string
}

// NewEmail creates a new Email value object.
func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return Email{}, ErrEmailEmpty
	}

	// Validate email format using net/mail
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return Email{}, ErrEmailInvalidFormat
	}

	// Normalize to lowercase
	normalized := strings.ToLower(addr.Address)

	// Additional validation for length
	if len(normalized) > 255 {
		return Email{}, ErrEmailTooLong
	}

	// Check for minimum length (a@b.c = 5 chars minimum)
	if len(normalized) < 5 {
		return Email{}, ErrEmailTooShort
	}

	return Email{value: normalized}, nil
}

// MustNewEmail creates a new Email or panics if invalid.
func MustNewEmail(value string) Email {
	e, err := NewEmail(value)
	if err != nil {
		panic(err)
	}
	return e
}

// String returns the string representation of the email.
func (e Email) String() string {
	return e.value
}

// Value returns the email value.
func (e Email) Value() string {
	return e.value
}

// Equals checks if two emails are equal.
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// IsEmpty returns true if the email is empty.
func (e Email) IsEmpty() bool {
	return e.value == ""
}

// Domain returns the domain part of the email.
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// LocalPart returns the local part (before @) of the email.
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// MaskEmail returns a masked version of the email for display.
// Example: john.doe@example.com -> j******e@example.com
func (e Email) MaskEmail() string {
	if e.IsEmpty() {
		return ""
	}

	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return e.value
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local + "@" + domain
	}

	masked := string(local[0]) + strings.Repeat("*", len(local)-2) + string(local[len(local)-1])
	return masked + "@" + domain
}

// Email errors
var (
	ErrEmailEmpty         = fmt.Errorf("email cannot be empty")
	ErrEmailInvalidFormat = fmt.Errorf("email has invalid format")
	ErrEmailTooLong       = fmt.Errorf("email exceeds maximum length of 255 characters")
	ErrEmailTooShort      = fmt.Errorf("email is too short")
)
