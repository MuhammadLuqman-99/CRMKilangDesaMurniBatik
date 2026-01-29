// Package domain contains the domain layer for the IAM service.
package domain

import (
	"testing"
)

func TestNewPermission(t *testing.T) {
	tests := []struct {
		name        string
		resource    string
		action      string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid permission users:read",
			resource: ResourceUsers,
			action:   ActionRead,
			wantErr:  false,
		},
		{
			name:     "valid permission with wildcard action",
			resource: ResourceUsers,
			action:   ActionAll,
			wantErr:  false,
		},
		{
			name:     "valid permission with wildcard resource",
			resource: ResourceAll,
			action:   ActionRead,
			wantErr:  false,
		},
		{
			name:     "valid full wildcard permission",
			resource: ResourceAll,
			action:   ActionAll,
			wantErr:  false,
		},
		{
			name:        "empty resource returns error",
			resource:    "",
			action:      ActionRead,
			wantErr:     true,
			expectedErr: ErrInvalidPermissionResource,
		},
		{
			name:        "empty action returns error",
			resource:    ResourceUsers,
			action:      "",
			wantErr:     true,
			expectedErr: ErrInvalidPermissionAction,
		},
		{
			name:     "valid permission with underscores",
			resource: "custom_resource",
			action:   "custom_action",
			wantErr:  false,
		},
		{
			name:     "spaces are trimmed",
			resource: "  users  ",
			action:   "  read  ",
			wantErr:  false,
		},
		{
			name:     "uppercase converted to lowercase",
			resource: "USERS",
			action:   "READ",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, err := NewPermission(tt.resource, tt.action)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPermission() expected error, got nil")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("NewPermission() error = %v, expectedErr = %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Errorf("NewPermission() unexpected error = %v", err)
				}
				if perm.Resource() == "" {
					t.Errorf("NewPermission() resource should not be empty")
				}
				if perm.Action() == "" {
					t.Errorf("NewPermission() action should not be empty")
				}
			}
		})
	}
}

func TestMustNewPermission(t *testing.T) {
	// Test valid permission
	t.Run("valid permission does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewPermission() panicked unexpectedly: %v", r)
			}
		}()
		perm := MustNewPermission(ResourceUsers, ActionRead)
		if perm.Resource() != ResourceUsers {
			t.Errorf("MustNewPermission() resource = %v, want %v", perm.Resource(), ResourceUsers)
		}
	})

	// Test invalid permission causes panic
	t.Run("invalid permission panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustNewPermission() expected panic for invalid permission")
			}
		}()
		MustNewPermission("", ActionRead)
	})
}

func TestParsePermission(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantErr          bool
		expectedResource string
		expectedAction   string
	}{
		{
			name:             "valid permission string",
			input:            "users:read",
			wantErr:          false,
			expectedResource: "users",
			expectedAction:   "read",
		},
		{
			name:             "wildcard permission",
			input:            "*",
			wantErr:          false,
			expectedResource: "*",
			expectedAction:   "*",
		},
		{
			name:             "resource wildcard with action",
			input:            "*:read",
			wantErr:          false,
			expectedResource: "*",
			expectedAction:   "read",
		},
		{
			name:             "resource with action wildcard",
			input:            "users:*",
			wantErr:          false,
			expectedResource: "users",
			expectedAction:   "*",
		},
		{
			name:    "invalid format no colon",
			input:   "usersread",
			wantErr: true,
		},
		{
			name:    "invalid format multiple colons",
			input:   "users:read:extra",
			wantErr: true,
		},
		{
			name:             "trims whitespace",
			input:            "  users:read  ",
			wantErr:          false,
			expectedResource: "users",
			expectedAction:   "read",
		},
		{
			name:             "converts to lowercase",
			input:            "USERS:READ",
			wantErr:          false,
			expectedResource: "users",
			expectedAction:   "read",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, err := ParsePermission(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePermission() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ParsePermission() unexpected error = %v", err)
				}
				if perm.Resource() != tt.expectedResource {
					t.Errorf("ParsePermission() resource = %v, want %v", perm.Resource(), tt.expectedResource)
				}
				if perm.Action() != tt.expectedAction {
					t.Errorf("ParsePermission() action = %v, want %v", perm.Action(), tt.expectedAction)
				}
			}
		})
	}
}

func TestMustParsePermission(t *testing.T) {
	t.Run("valid string does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParsePermission() panicked unexpectedly: %v", r)
			}
		}()
		perm := MustParsePermission("users:read")
		if perm.String() != "users:read" {
			t.Errorf("MustParsePermission() got %v, want %v", perm.String(), "users:read")
		}
	})

	t.Run("invalid string panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustParsePermission() expected panic")
			}
		}()
		MustParsePermission("invalid")
	})
}

func TestPermission_String(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		action   string
		expected string
	}{
		{
			name:     "regular permission",
			resource: "users",
			action:   "read",
			expected: "users:read",
		},
		{
			name:     "full wildcard",
			resource: "*",
			action:   "*",
			expected: "*",
		},
		{
			name:     "resource wildcard",
			resource: "*",
			action:   "read",
			expected: "*:read",
		},
		{
			name:     "action wildcard",
			resource: "users",
			action:   "*",
			expected: "users:*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, _ := NewPermission(tt.resource, tt.action)
			if perm.String() != tt.expected {
				t.Errorf("Permission.String() = %v, want %v", perm.String(), tt.expected)
			}
		})
	}
}

func TestPermission_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		action   string
		expected bool
	}{
		{
			name:     "valid permission",
			resource: "users",
			action:   "read",
			expected: true,
		},
		{
			name:     "full wildcard is valid",
			resource: "*",
			action:   "*",
			expected: true,
		},
		{
			name:     "partial wildcard is valid",
			resource: "users",
			action:   "*",
			expected: true,
		},
		{
			name:     "underscores allowed",
			resource: "custom_resource",
			action:   "custom_action",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, err := NewPermission(tt.resource, tt.action)
			if err != nil {
				t.Fatalf("NewPermission() failed: %v", err)
			}
			if perm.IsValid() != tt.expected {
				t.Errorf("Permission.IsValid() = %v, want %v", perm.IsValid(), tt.expected)
			}
		})
	}
}

func TestPermission_IsWildcard(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		action   string
		expected bool
	}{
		{
			name:     "full wildcard",
			resource: "*",
			action:   "*",
			expected: true,
		},
		{
			name:     "resource wildcard only",
			resource: "*",
			action:   "read",
			expected: false,
		},
		{
			name:     "action wildcard only",
			resource: "users",
			action:   "*",
			expected: false,
		},
		{
			name:     "no wildcard",
			resource: "users",
			action:   "read",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, _ := NewPermission(tt.resource, tt.action)
			if perm.IsWildcard() != tt.expected {
				t.Errorf("Permission.IsWildcard() = %v, want %v", perm.IsWildcard(), tt.expected)
			}
		})
	}
}

func TestPermission_HasWildcardResource(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		action   string
		expected bool
	}{
		{
			name:     "has wildcard resource",
			resource: "*",
			action:   "read",
			expected: true,
		},
		{
			name:     "no wildcard resource",
			resource: "users",
			action:   "*",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, _ := NewPermission(tt.resource, tt.action)
			if perm.HasWildcardResource() != tt.expected {
				t.Errorf("Permission.HasWildcardResource() = %v, want %v", perm.HasWildcardResource(), tt.expected)
			}
		})
	}
}

func TestPermission_HasWildcardAction(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		action   string
		expected bool
	}{
		{
			name:     "has wildcard action",
			resource: "users",
			action:   "*",
			expected: true,
		},
		{
			name:     "no wildcard action",
			resource: "*",
			action:   "read",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, _ := NewPermission(tt.resource, tt.action)
			if perm.HasWildcardAction() != tt.expected {
				t.Errorf("Permission.HasWildcardAction() = %v, want %v", perm.HasWildcardAction(), tt.expected)
			}
		})
	}
}

func TestPermission_Implies(t *testing.T) {
	tests := []struct {
		name     string
		perm     Permission
		other    Permission
		expected bool
	}{
		{
			name:     "full wildcard implies everything",
			perm:     MustNewPermission("*", "*"),
			other:    MustNewPermission("users", "read"),
			expected: true,
		},
		{
			name:     "action wildcard implies specific action",
			perm:     MustNewPermission("users", "*"),
			other:    MustNewPermission("users", "read"),
			expected: true,
		},
		{
			name:     "resource wildcard implies specific resource",
			perm:     MustNewPermission("*", "read"),
			other:    MustNewPermission("users", "read"),
			expected: true,
		},
		{
			name:     "exact match implies",
			perm:     MustNewPermission("users", "read"),
			other:    MustNewPermission("users", "read"),
			expected: true,
		},
		{
			name:     "different resource does not imply",
			perm:     MustNewPermission("users", "read"),
			other:    MustNewPermission("customers", "read"),
			expected: false,
		},
		{
			name:     "different action does not imply",
			perm:     MustNewPermission("users", "read"),
			other:    MustNewPermission("users", "write"),
			expected: false,
		},
		{
			name:     "specific does not imply wildcard",
			perm:     MustNewPermission("users", "read"),
			other:    MustNewPermission("users", "*"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.perm.Implies(tt.other) != tt.expected {
				t.Errorf("Permission.Implies() = %v, want %v", tt.perm.Implies(tt.other), tt.expected)
			}
		})
	}
}

func TestPermission_Equals(t *testing.T) {
	tests := []struct {
		name     string
		perm1    Permission
		perm2    Permission
		expected bool
	}{
		{
			name:     "equal permissions",
			perm1:    MustNewPermission("users", "read"),
			perm2:    MustNewPermission("users", "read"),
			expected: true,
		},
		{
			name:     "different resource",
			perm1:    MustNewPermission("users", "read"),
			perm2:    MustNewPermission("customers", "read"),
			expected: false,
		},
		{
			name:     "different action",
			perm1:    MustNewPermission("users", "read"),
			perm2:    MustNewPermission("users", "write"),
			expected: false,
		},
		{
			name:     "wildcards equal",
			perm1:    MustNewPermission("*", "*"),
			perm2:    MustNewPermission("*", "*"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.perm1.Equals(tt.perm2) != tt.expected {
				t.Errorf("Permission.Equals() = %v, want %v", tt.perm1.Equals(tt.perm2), tt.expected)
			}
		})
	}
}

func TestNewPermissionSet(t *testing.T) {
	ps := NewPermissionSet()
	if ps == nil {
		t.Fatal("NewPermissionSet() returned nil")
	}
	if ps.Len() != 0 {
		t.Errorf("NewPermissionSet() should be empty, got len = %d", ps.Len())
	}
	if !ps.IsEmpty() {
		t.Errorf("NewPermissionSet() should be empty")
	}
}

func TestNewPermissionSetFromStrings(t *testing.T) {
	tests := []struct {
		name        string
		perms       []string
		wantErr     bool
		expectedLen int
	}{
		{
			name:        "valid permissions",
			perms:       []string{"users:read", "users:write", "customers:*"},
			wantErr:     false,
			expectedLen: 3,
		},
		{
			name:        "empty slice",
			perms:       []string{},
			wantErr:     false,
			expectedLen: 0,
		},
		{
			name:    "invalid permission in slice",
			perms:   []string{"users:read", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps, err := NewPermissionSetFromStrings(tt.perms)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPermissionSetFromStrings() expected error")
				}
			} else {
				if err != nil {
					t.Errorf("NewPermissionSetFromStrings() unexpected error = %v", err)
				}
				if ps.Len() != tt.expectedLen {
					t.Errorf("NewPermissionSetFromStrings() len = %v, want %v", ps.Len(), tt.expectedLen)
				}
			}
		})
	}
}

func TestPermissionSet_Add(t *testing.T) {
	ps := NewPermissionSet()
	perm := MustNewPermission("users", "read")

	ps.Add(perm)
	if ps.Len() != 1 {
		t.Errorf("PermissionSet.Add() len = %v, want 1", ps.Len())
	}
	if !ps.Contains(perm) {
		t.Errorf("PermissionSet.Add() should contain added permission")
	}

	// Adding same permission should not increase count (it's a set)
	ps.Add(perm)
	if ps.Len() != 1 {
		t.Errorf("PermissionSet.Add() duplicate should not increase len, got %v", ps.Len())
	}
}

func TestPermissionSet_Remove(t *testing.T) {
	ps := NewPermissionSet()
	perm := MustNewPermission("users", "read")

	ps.Add(perm)
	ps.Remove(perm)

	if ps.Len() != 0 {
		t.Errorf("PermissionSet.Remove() len = %v, want 0", ps.Len())
	}
	if ps.Contains(perm) {
		t.Errorf("PermissionSet.Remove() should not contain removed permission")
	}
}

func TestPermissionSet_Contains(t *testing.T) {
	ps := NewPermissionSet()
	perm1 := MustNewPermission("users", "read")
	perm2 := MustNewPermission("customers", "write")

	ps.Add(perm1)

	if !ps.Contains(perm1) {
		t.Errorf("PermissionSet.Contains() should return true for existing permission")
	}
	if ps.Contains(perm2) {
		t.Errorf("PermissionSet.Contains() should return false for non-existing permission")
	}
}

func TestPermissionSet_HasPermission(t *testing.T) {
	ps := NewPermissionSet()
	ps.Add(MustNewPermission("users", "*"))

	tests := []struct {
		name     string
		required Permission
		expected bool
	}{
		{
			name:     "has exact permission with wildcard",
			required: MustNewPermission("users", "read"),
			expected: true,
		},
		{
			name:     "has another action on same resource",
			required: MustNewPermission("users", "write"),
			expected: true,
		},
		{
			name:     "does not have different resource",
			required: MustNewPermission("customers", "read"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ps.HasPermission(tt.required) != tt.expected {
				t.Errorf("PermissionSet.HasPermission() = %v, want %v", ps.HasPermission(tt.required), tt.expected)
			}
		})
	}
}

func TestPermissionSet_HasAnyPermission(t *testing.T) {
	ps := NewPermissionSet()
	ps.Add(MustNewPermission("users", "read"))

	// Should have at least one
	if !ps.HasAnyPermission(
		MustNewPermission("users", "read"),
		MustNewPermission("customers", "read"),
	) {
		t.Errorf("PermissionSet.HasAnyPermission() should return true")
	}

	// Should not have any
	if ps.HasAnyPermission(
		MustNewPermission("customers", "read"),
		MustNewPermission("orders", "read"),
	) {
		t.Errorf("PermissionSet.HasAnyPermission() should return false")
	}
}

func TestPermissionSet_HasAllPermissions(t *testing.T) {
	ps := NewPermissionSet()
	ps.Add(MustNewPermission("users", "*"))
	ps.Add(MustNewPermission("customers", "read"))

	// Should have all
	if !ps.HasAllPermissions(
		MustNewPermission("users", "read"),
		MustNewPermission("users", "write"),
		MustNewPermission("customers", "read"),
	) {
		t.Errorf("PermissionSet.HasAllPermissions() should return true")
	}

	// Should not have all
	if ps.HasAllPermissions(
		MustNewPermission("users", "read"),
		MustNewPermission("orders", "read"),
	) {
		t.Errorf("PermissionSet.HasAllPermissions() should return false")
	}
}

func TestPermissionSet_List(t *testing.T) {
	ps := NewPermissionSet()
	ps.Add(MustNewPermission("users", "read"))
	ps.Add(MustNewPermission("customers", "write"))

	list := ps.List()
	if len(list) != 2 {
		t.Errorf("PermissionSet.List() len = %v, want 2", len(list))
	}
}

func TestPermissionSet_Strings(t *testing.T) {
	ps := NewPermissionSet()
	ps.Add(MustNewPermission("users", "read"))
	ps.Add(MustNewPermission("customers", "write"))

	strings := ps.Strings()
	if len(strings) != 2 {
		t.Errorf("PermissionSet.Strings() len = %v, want 2", len(strings))
	}
}

func TestPermissionSet_Merge(t *testing.T) {
	ps1 := NewPermissionSet()
	ps1.Add(MustNewPermission("users", "read"))

	ps2 := NewPermissionSet()
	ps2.Add(MustNewPermission("customers", "write"))

	ps1.Merge(ps2)

	if ps1.Len() != 2 {
		t.Errorf("PermissionSet.Merge() len = %v, want 2", ps1.Len())
	}
	if !ps1.Contains(MustNewPermission("users", "read")) {
		t.Errorf("PermissionSet.Merge() should contain original permission")
	}
	if !ps1.Contains(MustNewPermission("customers", "write")) {
		t.Errorf("PermissionSet.Merge() should contain merged permission")
	}
}

func TestPermissionSet_Clone(t *testing.T) {
	ps := NewPermissionSet()
	ps.Add(MustNewPermission("users", "read"))
	ps.Add(MustNewPermission("customers", "write"))

	clone := ps.Clone()

	if clone.Len() != ps.Len() {
		t.Errorf("PermissionSet.Clone() len = %v, want %v", clone.Len(), ps.Len())
	}

	// Modify original, clone should not change
	ps.Add(MustNewPermission("orders", "read"))
	if clone.Len() != 2 {
		t.Errorf("PermissionSet.Clone() should be independent of original")
	}
}

func TestPredefinedPermissions(t *testing.T) {
	// Test that predefined permissions are valid
	predefined := []Permission{
		PermissionUsersCreate,
		PermissionUsersRead,
		PermissionUsersUpdate,
		PermissionUsersDelete,
		PermissionUsersList,
		PermissionUsersAll,
		PermissionRolesCreate,
		PermissionRolesRead,
		PermissionRolesUpdate,
		PermissionRolesDelete,
		PermissionRolesList,
		PermissionRolesAll,
		PermissionCustomersCreate,
		PermissionCustomersRead,
		PermissionCustomersUpdate,
		PermissionCustomersDelete,
		PermissionCustomersList,
		PermissionCustomersAll,
		PermissionLeadsCreate,
		PermissionLeadsRead,
		PermissionLeadsUpdate,
		PermissionLeadsDelete,
		PermissionLeadsList,
		PermissionLeadsAll,
		PermissionOpportunitiesCreate,
		PermissionOpportunitiesRead,
		PermissionOpportunitiesUpdate,
		PermissionOpportunitiesDelete,
		PermissionOpportunitiesList,
		PermissionOpportunitiesAll,
		PermissionSettingsRead,
		PermissionSettingsUpdate,
		PermissionSettingsAll,
		PermissionFullAccess,
	}

	for _, p := range predefined {
		if !p.IsValid() {
			t.Errorf("Predefined permission %s is not valid", p.String())
		}
	}
}
