// Package domain contains the domain layer for the IAM service.
package domain

import (
	"testing"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid email",
			input:   "test@example.com",
			want:    "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with uppercase",
			input:   "Test@Example.COM",
			want:    "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with spaces",
			input:   "  test@example.com  ",
			want:    "test@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			input:   "test@mail.example.com",
			want:    "test@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus sign",
			input:   "test+alias@example.com",
			want:    "test+alias@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			input:   "test.user@example.com",
			want:    "test.user@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only spaces",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "missing @",
			input:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "missing domain",
			input:   "test@",
			wantErr: true,
		},
		{
			name:    "missing local part",
			input:   "@example.com",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "a@b",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEmail(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("NewEmail(%q) unexpected error: %v", tt.input, err)
				return
			}

			if email.Value() != tt.want {
				t.Errorf("NewEmail(%q) = %q, want %q", tt.input, email.Value(), tt.want)
			}
		})
	}
}

func TestMustNewEmail(t *testing.T) {
	// Test valid email
	t.Run("valid email", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewEmail panicked unexpectedly: %v", r)
			}
		}()
		email := MustNewEmail("test@example.com")
		if email.Value() != "test@example.com" {
			t.Errorf("MustNewEmail returned wrong value: %s", email.Value())
		}
	})

	// Test invalid email panics
	t.Run("invalid email panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNewEmail should have panicked on invalid email")
			}
		}()
		MustNewEmail("")
	})
}

func TestEmail_String(t *testing.T) {
	email := MustNewEmail("test@example.com")
	if email.String() != "test@example.com" {
		t.Errorf("String() = %s, want test@example.com", email.String())
	}
}

func TestEmail_Equals(t *testing.T) {
	email1 := MustNewEmail("test@example.com")
	email2 := MustNewEmail("TEST@EXAMPLE.COM")
	email3 := MustNewEmail("other@example.com")

	if !email1.Equals(email2) {
		t.Error("Emails with same normalized value should be equal")
	}

	if email1.Equals(email3) {
		t.Error("Different emails should not be equal")
	}
}

func TestEmail_IsEmpty(t *testing.T) {
	email := MustNewEmail("test@example.com")
	if email.IsEmpty() {
		t.Error("Non-empty email should not be empty")
	}

	emptyEmail := Email{}
	if !emptyEmail.IsEmpty() {
		t.Error("Empty email should be empty")
	}
}

func TestEmail_Domain(t *testing.T) {
	email := MustNewEmail("test@example.com")
	if email.Domain() != "example.com" {
		t.Errorf("Domain() = %s, want example.com", email.Domain())
	}
}

func TestEmail_LocalPart(t *testing.T) {
	email := MustNewEmail("test@example.com")
	if email.LocalPart() != "test" {
		t.Errorf("LocalPart() = %s, want test", email.LocalPart())
	}
}

func TestEmail_MaskEmail(t *testing.T) {
	tests := []struct {
		email string
		want  string
	}{
		{"test@example.com", "t**t@example.com"},
		{"ab@example.com", "ab@example.com"},
		{"abc@example.com", "a*c@example.com"},
		{"john.doe@example.com", "j******e@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			email := MustNewEmail(tt.email)
			if email.MaskEmail() != tt.want {
				t.Errorf("MaskEmail() = %s, want %s", email.MaskEmail(), tt.want)
			}
		})
	}

	// Test empty email
	emptyEmail := Email{}
	if emptyEmail.MaskEmail() != "" {
		t.Error("Empty email should return empty mask")
	}
}
