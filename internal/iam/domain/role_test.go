// Package domain contains the domain layer for the IAM service.
package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewRole(t *testing.T) {
	tenantID := uuid.New()
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersRead)

	tests := []struct {
		name        string
		tenantID    *uuid.UUID
		roleName    string
		description string
		permissions *PermissionSet
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "valid role with tenant",
			tenantID:    &tenantID,
			roleName:    "test_role",
			description: "Test role description",
			permissions: permissions,
			wantErr:     false,
		},
		{
			name:        "valid role without tenant (system role)",
			tenantID:    nil,
			roleName:    "system_role",
			description: "System role",
			permissions: permissions,
			wantErr:     false,
		},
		{
			name:        "valid role with nil permissions",
			tenantID:    &tenantID,
			roleName:    "role_no_perms",
			description: "Role without permissions",
			permissions: nil,
			wantErr:     false,
		},
		{
			name:        "empty name returns error",
			tenantID:    &tenantID,
			roleName:    "",
			description: "Description",
			permissions: permissions,
			wantErr:     true,
			expectedErr: ErrRoleNameRequired,
		},
		{
			name:        "whitespace only name returns error",
			tenantID:    &tenantID,
			roleName:    "   ",
			description: "Description",
			permissions: permissions,
			wantErr:     true,
			expectedErr: ErrRoleNameRequired,
		},
		{
			name:        "name too long returns error",
			tenantID:    &tenantID,
			roleName:    strings.Repeat("a", 101),
			description: "Description",
			permissions: permissions,
			wantErr:     true,
			expectedErr: ErrRoleNameTooLong,
		},
		{
			name:        "name exactly 100 chars is valid",
			tenantID:    &tenantID,
			roleName:    strings.Repeat("a", 100),
			description: "Description",
			permissions: permissions,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := NewRole(tt.tenantID, tt.roleName, tt.description, tt.permissions)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewRole() expected error, got nil")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("NewRole() error = %v, expectedErr = %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("NewRole() unexpected error = %v", err)
				}
				if role == nil {
					t.Fatal("NewRole() returned nil role")
				}
				if role.ID == uuid.Nil {
					t.Error("NewRole() should generate ID")
				}
				if role.IsSystem() {
					t.Error("NewRole() should not be system role by default")
				}
				// Check domain event was added
				events := role.GetDomainEvents()
				if len(events) != 1 {
					t.Errorf("NewRole() should add RoleCreatedEvent, got %d events", len(events))
				}
			}
		})
	}
}

func TestNewSystemRole(t *testing.T) {
	permissions := NewPermissionSet()
	permissions.Add(PermissionFullAccess)

	role, err := NewSystemRole("super_admin", "Super Admin Role", permissions)
	if err != nil {
		t.Fatalf("NewSystemRole() unexpected error = %v", err)
	}

	if !role.IsSystem() {
		t.Error("NewSystemRole() should create system role")
	}
	if role.TenantID() != nil {
		t.Error("NewSystemRole() tenant ID should be nil")
	}
	if role.IsTenantRole() {
		t.Error("NewSystemRole() should not be tenant role")
	}
}

func TestReconstructRole(t *testing.T) {
	id := uuid.New()
	tenantID := uuid.New()
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersRead)
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	role := ReconstructRole(
		id,
		&tenantID,
		"reconstructed_role",
		"Reconstructed description",
		permissions,
		false,
		createdAt,
		updatedAt,
	)

	if role.ID != id {
		t.Errorf("ReconstructRole() ID = %v, want %v", role.ID, id)
	}
	if *role.TenantID() != tenantID {
		t.Errorf("ReconstructRole() TenantID = %v, want %v", role.TenantID(), tenantID)
	}
	if role.Name() != "reconstructed_role" {
		t.Errorf("ReconstructRole() Name = %v, want %v", role.Name(), "reconstructed_role")
	}
	if role.Description() != "Reconstructed description" {
		t.Errorf("ReconstructRole() Description = %v", role.Description())
	}
	if role.IsSystem() {
		t.Error("ReconstructRole() should not be system role")
	}
	// Reconstructed role should have no events
	if len(role.GetDomainEvents()) != 0 {
		t.Error("ReconstructRole() should not have domain events")
	}
}

func TestReconstructRole_NilPermissions(t *testing.T) {
	id := uuid.New()
	role := ReconstructRole(
		id,
		nil,
		"role",
		"description",
		nil,
		false,
		time.Now(),
		time.Now(),
	)

	if role.Permissions() == nil {
		t.Error("ReconstructRole() should initialize nil permissions")
	}
	if role.Permissions().Len() != 0 {
		t.Error("ReconstructRole() permissions should be empty")
	}
}

func TestRole_Getters(t *testing.T) {
	tenantID := uuid.New()
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersRead)

	role, _ := NewRole(&tenantID, "test_role", "Test description", permissions)

	if role.Name() != "test_role" {
		t.Errorf("Role.Name() = %v, want %v", role.Name(), "test_role")
	}
	if role.Description() != "Test description" {
		t.Errorf("Role.Description() = %v", role.Description())
	}
	if role.TenantID() == nil || *role.TenantID() != tenantID {
		t.Errorf("Role.TenantID() = %v, want %v", role.TenantID(), tenantID)
	}
	if role.Permissions() == nil {
		t.Error("Role.Permissions() should not be nil")
	}
	if role.IsSystem() {
		t.Error("Role.IsSystem() should be false")
	}
	if !role.IsTenantRole() {
		t.Error("Role.IsTenantRole() should be true")
	}
}

func TestRole_UpdateDetails(t *testing.T) {
	tenantID := uuid.New()
	role, _ := NewRole(&tenantID, "original_name", "Original description", nil)
	role.ClearDomainEvents()

	err := role.UpdateDetails("new_name", "New description")
	if err != nil {
		t.Fatalf("Role.UpdateDetails() unexpected error = %v", err)
	}

	if role.Name() != "new_name" {
		t.Errorf("Role.UpdateDetails() name = %v, want %v", role.Name(), "new_name")
	}
	if role.Description() != "New description" {
		t.Errorf("Role.UpdateDetails() description = %v", role.Description())
	}

	events := role.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Role.UpdateDetails() should add event, got %d events", len(events))
	}
}

func TestRole_UpdateDetails_SystemRole(t *testing.T) {
	role, _ := NewSystemRole("system_role", "System", nil)

	err := role.UpdateDetails("new_name", "New description")
	if err != ErrCannotModifySystemRole {
		t.Errorf("Role.UpdateDetails() on system role should return ErrCannotModifySystemRole, got %v", err)
	}
}

func TestRole_UpdateDetails_Validation(t *testing.T) {
	tenantID := uuid.New()
	role, _ := NewRole(&tenantID, "original", "Description", nil)

	// Empty name
	err := role.UpdateDetails("", "Description")
	if err != ErrRoleNameRequired {
		t.Errorf("Role.UpdateDetails() with empty name should return ErrRoleNameRequired, got %v", err)
	}

	// Name too long
	err = role.UpdateDetails(strings.Repeat("a", 101), "Description")
	if err != ErrRoleNameTooLong {
		t.Errorf("Role.UpdateDetails() with long name should return ErrRoleNameTooLong, got %v", err)
	}
}

func TestRole_AddPermission(t *testing.T) {
	tenantID := uuid.New()
	role, _ := NewRole(&tenantID, "role", "Description", nil)
	role.ClearDomainEvents()

	perm := PermissionUsersRead
	err := role.AddPermission(perm)
	if err != nil {
		t.Fatalf("Role.AddPermission() unexpected error = %v", err)
	}

	if !role.HasPermission(perm) {
		t.Error("Role.AddPermission() should add permission")
	}

	events := role.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Role.AddPermission() should add event, got %d events", len(events))
	}
}

func TestRole_AddPermission_SystemRole(t *testing.T) {
	role, _ := NewSystemRole("system_role", "System", nil)

	err := role.AddPermission(PermissionUsersRead)
	if err != ErrCannotModifySystemRole {
		t.Errorf("Role.AddPermission() on system role should return ErrCannotModifySystemRole, got %v", err)
	}
}

func TestRole_AddPermissions(t *testing.T) {
	tenantID := uuid.New()
	role, _ := NewRole(&tenantID, "role", "Description", nil)

	err := role.AddPermissions(PermissionUsersRead, PermissionUsersWrite, PermissionCustomersRead)
	if err != nil {
		t.Fatalf("Role.AddPermissions() unexpected error = %v", err)
	}

	if !role.HasPermission(PermissionUsersRead) {
		t.Error("Role.AddPermissions() should add PermissionUsersRead")
	}
	if !role.HasPermission(PermissionCustomersRead) {
		t.Error("Role.AddPermissions() should add PermissionCustomersRead")
	}
}

func TestRole_AddPermissions_SystemRole(t *testing.T) {
	role, _ := NewSystemRole("system_role", "System", nil)

	err := role.AddPermissions(PermissionUsersRead, PermissionCustomersRead)
	if err != ErrCannotModifySystemRole {
		t.Errorf("Role.AddPermissions() on system role should return ErrCannotModifySystemRole, got %v", err)
	}
}

func TestRole_RemovePermission(t *testing.T) {
	tenantID := uuid.New()
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersRead)
	permissions.Add(PermissionUsersWrite)
	role, _ := NewRole(&tenantID, "role", "Description", permissions)
	role.ClearDomainEvents()

	err := role.RemovePermission(PermissionUsersRead)
	if err != nil {
		t.Fatalf("Role.RemovePermission() unexpected error = %v", err)
	}

	if role.HasPermission(PermissionUsersRead) {
		t.Error("Role.RemovePermission() should remove permission")
	}
	if !role.HasPermission(PermissionUsersWrite) {
		t.Error("Role.RemovePermission() should not affect other permissions")
	}

	events := role.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Role.RemovePermission() should add event, got %d events", len(events))
	}
}

func TestRole_RemovePermission_SystemRole(t *testing.T) {
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersRead)
	role, _ := NewSystemRole("system_role", "System", permissions)

	err := role.RemovePermission(PermissionUsersRead)
	if err != ErrCannotModifySystemRole {
		t.Errorf("Role.RemovePermission() on system role should return ErrCannotModifySystemRole, got %v", err)
	}
}

func TestRole_SetPermissions(t *testing.T) {
	tenantID := uuid.New()
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersRead)
	role, _ := NewRole(&tenantID, "role", "Description", permissions)

	newPermissions := NewPermissionSet()
	newPermissions.Add(PermissionCustomersAll)

	err := role.SetPermissions(newPermissions)
	if err != nil {
		t.Fatalf("Role.SetPermissions() unexpected error = %v", err)
	}

	if role.HasPermission(PermissionUsersRead) {
		t.Error("Role.SetPermissions() should replace old permissions")
	}
	if !role.HasPermission(PermissionCustomersRead) {
		t.Error("Role.SetPermissions() should set new permissions")
	}
}

func TestRole_SetPermissions_SystemRole(t *testing.T) {
	role, _ := NewSystemRole("system_role", "System", nil)

	newPermissions := NewPermissionSet()
	newPermissions.Add(PermissionCustomersAll)

	err := role.SetPermissions(newPermissions)
	if err != ErrCannotModifySystemRole {
		t.Errorf("Role.SetPermissions() on system role should return ErrCannotModifySystemRole, got %v", err)
	}
}

func TestRole_HasPermission(t *testing.T) {
	tenantID := uuid.New()
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersAll)
	role, _ := NewRole(&tenantID, "role", "Description", permissions)

	tests := []struct {
		name       string
		permission Permission
		expected   bool
	}{
		{
			name:       "has exact permission with wildcard",
			permission: PermissionUsersRead,
			expected:   true,
		},
		{
			name:       "has another action from wildcard",
			permission: PermissionUsersCreate,
			expected:   true,
		},
		{
			name:       "does not have different resource",
			permission: PermissionCustomersRead,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if role.HasPermission(tt.permission) != tt.expected {
				t.Errorf("Role.HasPermission() = %v, want %v", role.HasPermission(tt.permission), tt.expected)
			}
		})
	}
}

func TestRole_Delete(t *testing.T) {
	tenantID := uuid.New()
	role, _ := NewRole(&tenantID, "role", "Description", nil)
	role.ClearDomainEvents()

	err := role.Delete()
	if err != nil {
		t.Fatalf("Role.Delete() unexpected error = %v", err)
	}

	if !role.IsDeleted() {
		t.Error("Role.Delete() should mark role as deleted")
	}

	events := role.GetDomainEvents()
	if len(events) != 1 {
		t.Errorf("Role.Delete() should add event, got %d events", len(events))
	}

	// Deleting again should be idempotent
	err = role.Delete()
	if err != nil {
		t.Errorf("Role.Delete() second call should be idempotent, got error = %v", err)
	}
}

func TestRole_Delete_SystemRole(t *testing.T) {
	role, _ := NewSystemRole("system_role", "System", nil)

	err := role.Delete()
	if err != ErrCannotDeleteSystemRole {
		t.Errorf("Role.Delete() on system role should return ErrCannotDeleteSystemRole, got %v", err)
	}
}

func TestCreateSuperAdminRole(t *testing.T) {
	role := CreateSuperAdminRole()

	if role == nil {
		t.Fatal("CreateSuperAdminRole() returned nil")
	}
	if role.Name() != RoleNameSuperAdmin {
		t.Errorf("CreateSuperAdminRole() name = %v, want %v", role.Name(), RoleNameSuperAdmin)
	}
	if !role.IsSystem() {
		t.Error("CreateSuperAdminRole() should be system role")
	}
	if !role.HasPermission(PermissionFullAccess) {
		t.Error("CreateSuperAdminRole() should have full access")
	}
}

func TestCreateAdminRole(t *testing.T) {
	role := CreateAdminRole()

	if role == nil {
		t.Fatal("CreateAdminRole() returned nil")
	}
	if role.Name() != RoleNameAdmin {
		t.Errorf("CreateAdminRole() name = %v, want %v", role.Name(), RoleNameAdmin)
	}
	if !role.IsSystem() {
		t.Error("CreateAdminRole() should be system role")
	}
	if !role.HasPermission(PermissionUsersRead) {
		t.Error("CreateAdminRole() should have users permission")
	}
}

func TestCreateManagerRole(t *testing.T) {
	role := CreateManagerRole()

	if role == nil {
		t.Fatal("CreateManagerRole() returned nil")
	}
	if role.Name() != RoleNameManager {
		t.Errorf("CreateManagerRole() name = %v, want %v", role.Name(), RoleNameManager)
	}
	if !role.IsSystem() {
		t.Error("CreateManagerRole() should be system role")
	}
	if !role.HasPermission(PermissionCustomersRead) {
		t.Error("CreateManagerRole() should have customers permission")
	}
}

func TestCreateSalesRepRole(t *testing.T) {
	role := CreateSalesRepRole()

	if role == nil {
		t.Fatal("CreateSalesRepRole() returned nil")
	}
	if role.Name() != RoleNameSalesRep {
		t.Errorf("CreateSalesRepRole() name = %v, want %v", role.Name(), RoleNameSalesRep)
	}
	if !role.IsSystem() {
		t.Error("CreateSalesRepRole() should be system role")
	}
	if !role.HasPermission(PermissionLeadsRead) {
		t.Error("CreateSalesRepRole() should have leads permission")
	}
	if !role.HasPermission(PermissionOpportunitiesRead) {
		t.Error("CreateSalesRepRole() should have opportunities permission")
	}
}

func TestCreateViewerRole(t *testing.T) {
	role := CreateViewerRole()

	if role == nil {
		t.Fatal("CreateViewerRole() returned nil")
	}
	if role.Name() != RoleNameViewer {
		t.Errorf("CreateViewerRole() name = %v, want %v", role.Name(), RoleNameViewer)
	}
	if !role.IsSystem() {
		t.Error("CreateViewerRole() should be system role")
	}
	// Viewer should have read-only access
	if !role.HasPermission(PermissionUsersRead) {
		t.Error("CreateViewerRole() should have users read permission")
	}
}

func TestRole_NameTrimming(t *testing.T) {
	tenantID := uuid.New()
	role, err := NewRole(&tenantID, "  trimmed_name  ", "Description", nil)
	if err != nil {
		t.Fatalf("NewRole() unexpected error = %v", err)
	}

	if role.Name() != "trimmed_name" {
		t.Errorf("NewRole() should trim name, got %v", role.Name())
	}
}
