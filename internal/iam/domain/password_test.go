// Package domain contains the domain layer for the IAM service.
package domain

import (
	"strings"
	"testing"
)

// mockHasher implements PasswordHasher for testing
type mockHasher struct {
	hashPrefix string
	shouldFail bool
}

func (m *mockHasher) Hash(password string) (string, error) {
	if m.shouldFail {
		return "", ErrPasswordInvalid
	}
	return m.hashPrefix + password, nil
}

func (m *mockHasher) Verify(password, hash string) (bool, error) {
	if m.shouldFail {
		return false, ErrPasswordInvalid
	}
	return hash == m.hashPrefix+password, nil
}

func TestDefaultPasswordPolicy(t *testing.T) {
	policy := DefaultPasswordPolicy()

	if policy.MinLength != 8 {
		t.Errorf("MinLength = %d, want 8", policy.MinLength)
	}
	if policy.MaxLength != 128 {
		t.Errorf("MaxLength = %d, want 128", policy.MaxLength)
	}
	if !policy.RequireUppercase {
		t.Error("RequireUppercase should be true")
	}
	if !policy.RequireLowercase {
		t.Error("RequireLowercase should be true")
	}
	if !policy.RequireDigit {
		t.Error("RequireDigit should be true")
	}
	if policy.RequireSpecial {
		t.Error("RequireSpecial should be false")
	}
}

func TestStrictPasswordPolicy(t *testing.T) {
	policy := StrictPasswordPolicy()

	if policy.MinLength != 12 {
		t.Errorf("MinLength = %d, want 12", policy.MinLength)
	}
	if !policy.RequireSpecial {
		t.Error("RequireSpecial should be true for strict policy")
	}
}

func TestNewPassword(t *testing.T) {
	hasher := &mockHasher{hashPrefix: "hashed_"}
	policy := DefaultPasswordPolicy()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errType  error
	}{
		{
			name:     "valid password",
			password: "Password123",
			wantErr:  false,
		},
		{
			name:     "valid complex password",
			password: "MyStr0ngP@ssw0rd!",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1",
			wantErr:  true,
			errType:  ErrPasswordTooShort,
		},
		{
			name:     "no uppercase",
			password: "password123",
			wantErr:  true,
			errType:  ErrPasswordNoUppercase,
		},
		{
			name:     "no lowercase",
			password: "PASSWORD123",
			wantErr:  true,
			errType:  ErrPasswordNoLowercase,
		},
		{
			name:     "no digit",
			password: "PasswordABC",
			wantErr:  true,
			errType:  ErrPasswordNoDigit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwd, err := NewPassword(tt.password, hasher, policy)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPassword(%q) expected error, got nil", tt.password)
				}
				return
			}

			if err != nil {
				t.Errorf("NewPassword(%q) unexpected error: %v", tt.password, err)
				return
			}

			expectedHash := "hashed_" + tt.password
			if pwd.Hash() != expectedHash {
				t.Errorf("Hash() = %s, want %s", pwd.Hash(), expectedHash)
			}
		})
	}
}

func TestNewPassword_WithStrictPolicy(t *testing.T) {
	hasher := &mockHasher{hashPrefix: "hashed_"}
	policy := StrictPasswordPolicy()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid strict password",
			password: "MyStr0ngP@ss!",
			wantErr:  false,
		},
		{
			name:     "missing special character",
			password: "MyStr0ngPassword",
			wantErr:  true,
		},
		{
			name:     "too short for strict policy",
			password: "P@ssw0rd!",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPassword(tt.password, hasher, policy)

			if tt.wantErr && err == nil {
				t.Errorf("NewPassword(%q) expected error, got nil", tt.password)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("NewPassword(%q) unexpected error: %v", tt.password, err)
			}
		})
	}
}

func TestNewPassword_HasherFailure(t *testing.T) {
	hasher := &mockHasher{shouldFail: true}
	policy := DefaultPasswordPolicy()

	_, err := NewPassword("Password123", hasher, policy)
	if err == nil {
		t.Error("Expected error when hasher fails")
	}
}

func TestNewPasswordFromHash(t *testing.T) {
	hash := "existing_hash_value"
	pwd := NewPasswordFromHash(hash)

	if pwd.Hash() != hash {
		t.Errorf("Hash() = %s, want %s", pwd.Hash(), hash)
	}
}

func TestPassword_Verify(t *testing.T) {
	hasher := &mockHasher{hashPrefix: "hashed_"}
	policy := DefaultPasswordPolicy()

	pwd, _ := NewPassword("Password123", hasher, policy)

	// Correct password
	valid, err := pwd.Verify("Password123", hasher)
	if err != nil {
		t.Errorf("Verify() unexpected error: %v", err)
	}
	if !valid {
		t.Error("Verify() should return true for correct password")
	}

	// Wrong password
	valid, err = pwd.Verify("WrongPassword", hasher)
	if err != nil {
		t.Errorf("Verify() unexpected error: %v", err)
	}
	if valid {
		t.Error("Verify() should return false for wrong password")
	}
}

func TestPassword_IsEmpty(t *testing.T) {
	hasher := &mockHasher{hashPrefix: "hashed_"}
	policy := DefaultPasswordPolicy()

	pwd, _ := NewPassword("Password123", hasher, policy)
	if pwd.IsEmpty() {
		t.Error("Password with hash should not be empty")
	}

	emptyPwd := Password{}
	if !emptyPwd.IsEmpty() {
		t.Error("Password without hash should be empty")
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	policy := DefaultPasswordPolicy()

	tests := []struct {
		name          string
		password      string
		expectedErrs  int
	}{
		{
			name:         "valid password",
			password:     "Password123",
			expectedErrs: 0,
		},
		{
			name:         "all requirements missing",
			password:     "pass",
			expectedErrs: 3, // too short, no uppercase, no digit
		},
		{
			name:         "only lowercase",
			password:     "verylongpassword",
			expectedErrs: 2, // no uppercase, no digit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidatePasswordStrength(tt.password, policy)
			if len(errs) != tt.expectedErrs {
				t.Errorf("ValidatePasswordStrength() returned %d errors, want %d", len(errs), tt.expectedErrs)
			}
		})
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	// Test default length
	pwd, err := GenerateRandomPassword(0)
	if err != nil {
		t.Errorf("GenerateRandomPassword() error: %v", err)
	}
	if len(pwd) < 8 {
		t.Errorf("Generated password length = %d, want >= 8", len(pwd))
	}

	// Test custom length
	pwd, err = GenerateRandomPassword(16)
	if err != nil {
		t.Errorf("GenerateRandomPassword() error: %v", err)
	}
	if len(pwd) != 16 {
		t.Errorf("Generated password length = %d, want 16", len(pwd))
	}

	// Test uniqueness
	pwd1, _ := GenerateRandomPassword(12)
	pwd2, _ := GenerateRandomPassword(12)
	if pwd1 == pwd2 {
		t.Error("GenerateRandomPassword should generate unique passwords")
	}
}

func TestGenerateSecurePassword(t *testing.T) {
	policy := DefaultPasswordPolicy()
	pwd, err := GenerateSecurePassword(policy)

	if err != nil {
		t.Errorf("GenerateSecurePassword() error: %v", err)
	}

	// Validate the generated password meets policy
	errs := ValidatePasswordStrength(pwd, policy)
	if len(errs) > 0 {
		t.Errorf("Generated password doesn't meet policy: %v", errs)
	}

	// Check it has required characters
	hasUpper := strings.ContainsAny(pwd, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(pwd, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(pwd, "0123456789")

	if !hasUpper {
		t.Error("Generated password should have uppercase letter")
	}
	if !hasLower {
		t.Error("Generated password should have lowercase letter")
	}
	if !hasDigit {
		t.Error("Generated password should have digit")
	}
}

func TestGenerateSecurePassword_StrictPolicy(t *testing.T) {
	policy := StrictPasswordPolicy()
	pwd, err := GenerateSecurePassword(policy)

	if err != nil {
		t.Errorf("GenerateSecurePassword() error: %v", err)
	}

	// Validate the generated password meets strict policy
	errs := ValidatePasswordStrength(pwd, policy)
	if len(errs) > 0 {
		t.Errorf("Generated password doesn't meet strict policy: %v", errs)
	}

	// Check it has special characters
	hasSpecial := strings.ContainsAny(pwd, "!@#$%^&*()_+-=[]{}|;:,.<>?")
	if !hasSpecial {
		t.Error("Generated password should have special character for strict policy")
	}
}

func TestPasswordTooLong(t *testing.T) {
	hasher := &mockHasher{hashPrefix: "hashed_"}
	policy := PasswordPolicy{
		MinLength:        8,
		MaxLength:        20,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}

	longPassword := "Password123" + strings.Repeat("a", 20)
	_, err := NewPassword(longPassword, hasher, policy)

	if err == nil {
		t.Error("Expected error for password exceeding max length")
	}
}
