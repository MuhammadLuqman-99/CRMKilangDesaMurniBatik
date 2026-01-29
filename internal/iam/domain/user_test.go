// Package domain contains the domain layer for the IAM service.
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserStatus_IsValid(t *testing.T) {
	tests := []struct {
		status UserStatus
		want   bool
	}{
		{UserStatusActive, true},
		{UserStatusInactive, true},
		{UserStatusSuspended, true},
		{UserStatusPending, true},
		{UserStatus("invalid"), false},
		{UserStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("UserStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestUserStatus_String(t *testing.T) {
	status := UserStatusActive
	if status.String() != "active" {
		t.Errorf("String() = %s, want active", status.String())
	}
}

func TestNewUser(t *testing.T) {
	tenantID := uuid.New()
	email := MustNewEmail("test@example.com")
	password := NewPasswordFromHash("hashed_password")

	tests := []struct {
		name      string
		tenantID  uuid.UUID
		email     Email
		password  Password
		firstName string
		lastName  string
		wantErr   bool
	}{
		{
			name:      "valid user",
			tenantID:  tenantID,
			email:     email,
			password:  password,
			firstName: "John",
			lastName:  "Doe",
			wantErr:   false,
		},
		{
			name:      "valid user without names",
			tenantID:  tenantID,
			email:     email,
			password:  password,
			firstName: "",
			lastName:  "",
			wantErr:   false,
		},
		{
			name:      "missing tenant ID",
			tenantID:  uuid.Nil,
			email:     email,
			password:  password,
			firstName: "John",
			lastName:  "Doe",
			wantErr:   true,
		},
		{
			name:      "missing email",
			tenantID:  tenantID,
			email:     Email{},
			password:  password,
			firstName: "John",
			lastName:  "Doe",
			wantErr:   true,
		},
		{
			name:      "missing password",
			tenantID:  tenantID,
			email:     email,
			password:  Password{},
			firstName: "John",
			lastName:  "Doe",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.tenantID, tt.email, tt.password, tt.firstName, tt.lastName)

			if tt.wantErr {
				if err == nil {
					t.Error("NewUser() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewUser() unexpected error: %v", err)
				return
			}

			// Verify user properties
			if user.TenantID() != tt.tenantID {
				t.Errorf("TenantID() = %s, want %s", user.TenantID(), tt.tenantID)
			}
			if !user.Email().Equals(tt.email) {
				t.Errorf("Email() = %s, want %s", user.Email(), tt.email)
			}
			if user.FirstName() != tt.firstName {
				t.Errorf("FirstName() = %s, want %s", user.FirstName(), tt.firstName)
			}
			if user.LastName() != tt.lastName {
				t.Errorf("LastName() = %s, want %s", user.LastName(), tt.lastName)
			}
			if user.Status() != UserStatusPending {
				t.Errorf("Status() = %s, want pending", user.Status())
			}

			// Verify domain event was created
			events := user.GetDomainEvents()
			if len(events) == 0 {
				t.Error("Expected UserCreatedEvent to be added")
			}
		})
	}
}

func TestReconstructUser(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	email := MustNewEmail("test@example.com")
	password := NewPasswordFromHash("hashed_password")
	now := time.Now().UTC()

	user := ReconstructUser(
		id, tenantID, email, password,
		"John", "Doe", "https://avatar.com/john.png", "+1234567890",
		UserStatusActive, &now, &now,
		map[string]interface{}{"key": "value"},
		now, now, nil,
	)

	if user.GetID() != id {
		t.Errorf("ID() = %s, want %s", user.GetID(), id)
	}
	if user.TenantID() != tenantID {
		t.Errorf("TenantID() = %s, want %s", user.TenantID(), tenantID)
	}
	if user.FirstName() != "John" {
		t.Errorf("FirstName() = %s, want John", user.FirstName())
	}
	if user.AvatarURL() != "https://avatar.com/john.png" {
		t.Errorf("AvatarURL() = %s, want https://avatar.com/john.png", user.AvatarURL())
	}
	if user.Phone() != "+1234567890" {
		t.Errorf("Phone() = %s, want +1234567890", user.Phone())
	}
	if user.Status() != UserStatusActive {
		t.Errorf("Status() = %s, want active", user.Status())
	}
	if !user.IsEmailVerified() {
		t.Error("IsEmailVerified() should be true")
	}
}

func TestUser_FullName(t *testing.T) {
	tests := []struct {
		firstName string
		lastName  string
		want      string
	}{
		{"John", "Doe", "John Doe"},
		{"John", "", "John"},
		{"", "Doe", "Doe"},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			tenantID := uuid.New()
			email := MustNewEmail("test@example.com")
			password := NewPasswordFromHash("hashed")
			user, _ := NewUser(tenantID, email, password, tt.firstName, tt.lastName)

			if user.FullName() != tt.want {
				t.Errorf("FullName() = %s, want %s", user.FullName(), tt.want)
			}
		})
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	user := createTestUser(t)

	user.UpdateProfile("Jane", "Smith", "+9876543210", "https://new-avatar.com/jane.png")

	if user.FirstName() != "Jane" {
		t.Errorf("FirstName() = %s, want Jane", user.FirstName())
	}
	if user.LastName() != "Smith" {
		t.Errorf("LastName() = %s, want Smith", user.LastName())
	}
	if user.Phone() != "+9876543210" {
		t.Errorf("Phone() = %s, want +9876543210", user.Phone())
	}
	if user.AvatarURL() != "https://new-avatar.com/jane.png" {
		t.Errorf("AvatarURL() = %s, want https://new-avatar.com/jane.png", user.AvatarURL())
	}

	// Check domain event
	events := user.GetDomainEvents()
	hasUpdateEvent := false
	for _, e := range events {
		if _, ok := e.(*UserUpdatedEvent); ok {
			hasUpdateEvent = true
			break
		}
	}
	if !hasUpdateEvent {
		t.Error("Expected UserUpdatedEvent to be added")
	}
}

func TestUser_UpdateEmail(t *testing.T) {
	user := createTestUser(t)
	user.VerifyEmail() // Verify first

	if !user.IsEmailVerified() {
		t.Error("Email should be verified initially")
	}

	newEmail := MustNewEmail("new@example.com")
	err := user.UpdateEmail(newEmail)
	if err != nil {
		t.Errorf("UpdateEmail() error: %v", err)
	}

	if !user.Email().Equals(newEmail) {
		t.Errorf("Email() = %s, want %s", user.Email(), newEmail)
	}

	// Email verification should be reset
	if user.IsEmailVerified() {
		t.Error("Email verification should be reset after email change")
	}

	// Test with empty email
	err = user.UpdateEmail(Email{})
	if err == nil {
		t.Error("UpdateEmail() should error with empty email")
	}
}

func TestUser_ChangePassword(t *testing.T) {
	user := createTestUser(t)

	newPassword := NewPasswordFromHash("new_hashed_password")
	err := user.ChangePassword(newPassword)
	if err != nil {
		t.Errorf("ChangePassword() error: %v", err)
	}

	if user.PasswordHash().Hash() != "new_hashed_password" {
		t.Error("Password should be updated")
	}

	// Test with empty password
	err = user.ChangePassword(Password{})
	if err == nil {
		t.Error("ChangePassword() should error with empty password")
	}
}

func TestUser_VerifyEmail(t *testing.T) {
	user := createTestUser(t)

	if user.IsEmailVerified() {
		t.Error("Email should not be verified initially")
	}

	user.VerifyEmail()

	if !user.IsEmailVerified() {
		t.Error("Email should be verified after VerifyEmail()")
	}
	if user.Status() != UserStatusActive {
		t.Error("User should be activated after email verification")
	}

	// Calling again should have no effect
	originalTime := user.EmailVerifiedAt()
	user.VerifyEmail()
	if *user.EmailVerifiedAt() != *originalTime {
		t.Error("Verifying again should not change timestamp")
	}
}

func TestUser_RecordLogin(t *testing.T) {
	user := createTestUser(t)

	if user.LastLoginAt() != nil {
		t.Error("LastLoginAt should be nil initially")
	}

	user.RecordLogin()

	if user.LastLoginAt() == nil {
		t.Error("LastLoginAt should be set after RecordLogin()")
	}
}

func TestUser_Activate(t *testing.T) {
	user := createTestUser(t)

	err := user.Activate()
	if err != nil {
		t.Errorf("Activate() error: %v", err)
	}

	if user.Status() != UserStatusActive {
		t.Errorf("Status() = %s, want active", user.Status())
	}

	// Activating again should have no effect
	err = user.Activate()
	if err != nil {
		t.Errorf("Activate() again error: %v", err)
	}
}

func TestUser_Deactivate(t *testing.T) {
	user := createTestUser(t)
	user.Activate()

	err := user.Deactivate()
	if err != nil {
		t.Errorf("Deactivate() error: %v", err)
	}

	if user.Status() != UserStatusInactive {
		t.Errorf("Status() = %s, want inactive", user.Status())
	}
}

func TestUser_Suspend(t *testing.T) {
	user := createTestUser(t)
	user.Activate()

	err := user.Suspend("Violation of terms")
	if err != nil {
		t.Errorf("Suspend() error: %v", err)
	}

	if user.Status() != UserStatusSuspended {
		t.Errorf("Status() = %s, want suspended", user.Status())
	}

	reason, ok := user.GetMetadata("suspend_reason")
	if !ok || reason != "Violation of terms" {
		t.Error("Suspend reason should be stored in metadata")
	}
}

func TestUser_Unsuspend(t *testing.T) {
	user := createTestUser(t)
	user.Suspend("Test reason")

	err := user.Unsuspend()
	if err != nil {
		t.Errorf("Unsuspend() error: %v", err)
	}

	if user.Status() != UserStatusActive {
		t.Errorf("Status() = %s, want active", user.Status())
	}

	_, ok := user.GetMetadata("suspend_reason")
	if ok {
		t.Error("Suspend reason should be removed from metadata")
	}

	// Unsuspending non-suspended user should error
	err = user.Unsuspend()
	if err == nil {
		t.Error("Unsuspend() should error for non-suspended user")
	}
}

func TestUser_Delete(t *testing.T) {
	user := createTestUser(t)

	user.Delete()

	if !user.IsDeleted() {
		t.Error("User should be deleted")
	}

	// Deleting again should have no effect
	user.Delete()
	if !user.IsDeleted() {
		t.Error("User should still be deleted")
	}
}

func TestUser_CanLogin(t *testing.T) {
	user := createTestUser(t)

	// Pending user cannot login
	if user.CanLogin() {
		t.Error("Pending user should not be able to login")
	}

	// Active user can login
	user.Activate()
	if !user.CanLogin() {
		t.Error("Active user should be able to login")
	}

	// Deleted user cannot login
	user.Delete()
	if user.CanLogin() {
		t.Error("Deleted user should not be able to login")
	}
}

func TestUser_RoleManagement(t *testing.T) {
	user := createTestUser(t)
	role := createTestRole(t)

	// Assign role
	err := user.AssignRole(role)
	if err != nil {
		t.Errorf("AssignRole() error: %v", err)
	}

	if !user.HasRole(role.GetID()) {
		t.Error("User should have the assigned role")
	}
	if !user.HasRoleByName(role.Name()) {
		t.Error("User should have the assigned role by name")
	}

	// Cannot assign same role twice
	err = user.AssignRole(role)
	if err == nil {
		t.Error("Should not be able to assign same role twice")
	}

	// Cannot assign nil role
	err = user.AssignRole(nil)
	if err == nil {
		t.Error("Should not be able to assign nil role")
	}

	// Remove role
	err = user.RemoveRole(role.GetID())
	if err != nil {
		t.Errorf("RemoveRole() error: %v", err)
	}

	if user.HasRole(role.GetID()) {
		t.Error("User should not have the removed role")
	}

	// Cannot remove non-existent role
	err = user.RemoveRole(uuid.New())
	if err == nil {
		t.Error("Should not be able to remove non-existent role")
	}
}

func TestUser_Metadata(t *testing.T) {
	user := createTestUser(t)

	// Set metadata
	user.SetMetadata("key1", "value1")
	user.SetMetadata("key2", 123)

	// Get metadata
	val, ok := user.GetMetadata("key1")
	if !ok || val != "value1" {
		t.Error("Should get metadata value")
	}

	// Non-existent key
	_, ok = user.GetMetadata("nonexistent")
	if ok {
		t.Error("Should not find non-existent key")
	}

	// Delete metadata
	user.DeleteMetadata("key1")
	_, ok = user.GetMetadata("key1")
	if ok {
		t.Error("Key should be deleted")
	}
}

// Helper functions
func createTestUser(t *testing.T) *User {
	t.Helper()
	tenantID := uuid.New()
	email := MustNewEmail("test@example.com")
	password := NewPasswordFromHash("hashed_password")
	user, err := NewUser(tenantID, email, password, "John", "Doe")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func createTestRole(t *testing.T) *Role {
	t.Helper()
	tenantID := uuid.New()
	role, err := NewRole(tenantID, "Test Role", "Test role description")
	if err != nil {
		t.Fatalf("Failed to create test role: %v", err)
	}
	return role
}
