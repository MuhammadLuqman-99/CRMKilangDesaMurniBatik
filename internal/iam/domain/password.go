// Package domain contains the domain layer for the IAM service.
package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"unicode"
)

// PasswordPolicy defines the password requirements.
type PasswordPolicy struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
	RequireSpecial   bool
}

// DefaultPasswordPolicy returns the default password policy.
func DefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:        8,
		MaxLength:        128,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   false,
	}
}

// StrictPasswordPolicy returns a strict password policy.
func StrictPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:        12,
		MaxLength:        128,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   true,
	}
}

// Password represents a password value object.
// It stores the hashed password, never the plain text.
type Password struct {
	hash string
}

// PasswordHasher is the interface for password hashing.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}

// NewPassword creates a new Password from a plain text password.
// It validates and hashes the password using the provided hasher.
func NewPassword(plainText string, hasher PasswordHasher, policy PasswordPolicy) (Password, error) {
	// Validate password against policy
	if err := validatePassword(plainText, policy); err != nil {
		return Password{}, err
	}

	// Hash the password
	hash, err := hasher.Hash(plainText)
	if err != nil {
		return Password{}, fmt.Errorf("failed to hash password: %w", err)
	}

	return Password{hash: hash}, nil
}

// NewPasswordFromHash creates a Password from an existing hash.
// This is used when loading from database.
func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

// Hash returns the password hash.
func (p Password) Hash() string {
	return p.hash
}

// Verify checks if the provided plain text matches the password hash.
func (p Password) Verify(plainText string, hasher PasswordHasher) (bool, error) {
	return hasher.Verify(plainText, p.hash)
}

// IsEmpty returns true if the password hash is empty.
func (p Password) IsEmpty() bool {
	return p.hash == ""
}

// validatePassword validates a password against the policy.
func validatePassword(password string, policy PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return fmt.Errorf("%w: minimum %d characters required", ErrPasswordTooShort, policy.MinLength)
	}

	if len(password) > policy.MaxLength {
		return fmt.Errorf("%w: maximum %d characters allowed", ErrPasswordTooLong, policy.MaxLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		return ErrPasswordNoUppercase
	}

	if policy.RequireLowercase && !hasLower {
		return ErrPasswordNoLowercase
	}

	if policy.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}

	if policy.RequireSpecial && !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// ValidatePasswordStrength validates password strength and returns errors.
func ValidatePasswordStrength(password string, policy PasswordPolicy) []error {
	var errors []error

	if len(password) < policy.MinLength {
		errors = append(errors, fmt.Errorf("password must be at least %d characters", policy.MinLength))
	}

	if len(password) > policy.MaxLength {
		errors = append(errors, fmt.Errorf("password must be at most %d characters", policy.MaxLength))
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if policy.RequireUppercase && !hasUpper {
		errors = append(errors, fmt.Errorf("password must contain at least one uppercase letter"))
	}

	if policy.RequireLowercase && !hasLower {
		errors = append(errors, fmt.Errorf("password must contain at least one lowercase letter"))
	}

	if policy.RequireDigit && !hasDigit {
		errors = append(errors, fmt.Errorf("password must contain at least one digit"))
	}

	if policy.RequireSpecial && !hasSpecial {
		errors = append(errors, fmt.Errorf("password must contain at least one special character"))
	}

	return errors
}

// GenerateRandomPassword generates a random password of the specified length.
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}

	// Encode to base64 and trim to desired length
	password := base64.RawURLEncoding.EncodeToString(bytes)
	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}

// GenerateSecurePassword generates a password that meets the given policy.
func GenerateSecurePassword(policy PasswordPolicy) (string, error) {
	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		digits    = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	length := policy.MinLength
	if length < 12 {
		length = 12
	}

	var charset string
	var required []byte

	if policy.RequireUppercase {
		charset += uppercase
		idx, _ := randInt(len(uppercase))
		required = append(required, uppercase[idx])
	}

	if policy.RequireLowercase {
		charset += lowercase
		idx, _ := randInt(len(lowercase))
		required = append(required, lowercase[idx])
	}

	if policy.RequireDigit {
		charset += digits
		idx, _ := randInt(len(digits))
		required = append(required, digits[idx])
	}

	if policy.RequireSpecial {
		charset += special
		idx, _ := randInt(len(special))
		required = append(required, special[idx])
	}

	// If no requirements, use all character sets
	if charset == "" {
		charset = uppercase + lowercase + digits
	}

	// Generate remaining characters
	remaining := length - len(required)
	if remaining < 0 {
		remaining = 0
	}

	result := make([]byte, remaining)
	for i := range result {
		idx, err := randInt(len(charset))
		if err != nil {
			return "", err
		}
		result[i] = charset[idx]
	}

	// Combine required and random characters
	password := append(required, result...)

	// Shuffle the password
	for i := len(password) - 1; i > 0; i-- {
		j, err := randInt(i + 1)
		if err != nil {
			return "", err
		}
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}

// randInt returns a random int in [0, max)
func randInt(max int) (int, error) {
	if max <= 0 {
		return 0, nil
	}
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		return 0, err
	}
	return int(b[0]) % max, nil
}

// Password errors
var (
	ErrPasswordTooShort    = fmt.Errorf("password is too short")
	ErrPasswordTooLong     = fmt.Errorf("password is too long")
	ErrPasswordNoUppercase = fmt.Errorf("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase = fmt.Errorf("password must contain at least one lowercase letter")
	ErrPasswordNoDigit     = fmt.Errorf("password must contain at least one digit")
	ErrPasswordNoSpecial   = fmt.Errorf("password must contain at least one special character")
	ErrPasswordInvalid     = fmt.Errorf("invalid password")
	ErrPasswordMismatch    = fmt.Errorf("passwords do not match")
)
