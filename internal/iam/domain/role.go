// Package domain contains the domain layer for the IAM service.
package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Predefined system role names
const (
	RoleNameSuperAdmin = "super_admin"
	RoleNameAdmin      = "admin"
	RoleNameManager    = "manager"
	RoleNameSalesRep   = "sales_rep"
	RoleNameViewer     = "viewer"
)

// Role represents a role entity in the IAM domain.
type Role struct {
	BaseAggregateRoot
	tenantID    *uuid.UUID // nil for system roles
	name        string
	description string
	permissions *PermissionSet
	isSystem    bool
}

// NewRole creates a new Role entity.
func NewRole(tenantID *uuid.UUID, name, description string, permissions *PermissionSet) (*Role, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrRoleNameRequired
	}

	if len(name) > 100 {
		return nil, ErrRoleNameTooLong
	}

	if permissions == nil {
		permissions = NewPermissionSet()
	}

	role := &Role{
		BaseAggregateRoot: NewBaseAggregateRoot(),
		tenantID:          tenantID,
		name:              name,
		description:       description,
		permissions:       permissions,
		isSystem:          false,
	}

	role.AddDomainEvent(NewRoleCreatedEvent(role))

	return role, nil
}

// NewSystemRole creates a new system role (not tenant-specific).
func NewSystemRole(name, description string, permissions *PermissionSet) (*Role, error) {
	role, err := NewRole(nil, name, description, permissions)
	if err != nil {
		return nil, err
	}
	role.isSystem = true
	return role, nil
}

// ReconstructRole reconstructs a Role entity from persistence.
func ReconstructRole(
	id uuid.UUID,
	tenantID *uuid.UUID,
	name, description string,
	permissions *PermissionSet,
	isSystem bool,
	createdAt, updatedAt time.Time,
) *Role {
	if permissions == nil {
		permissions = NewPermissionSet()
	}

	return &Role{
		BaseAggregateRoot: BaseAggregateRoot{
			BaseEntity: BaseEntity{
				ID:        id,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			domainEvents: make([]DomainEvent, 0),
		},
		tenantID:    tenantID,
		name:        name,
		description: description,
		permissions: permissions,
		isSystem:    isSystem,
	}
}

// Getters

// TenantID returns the tenant ID (nil for system roles).
func (r *Role) TenantID() *uuid.UUID {
	return r.tenantID
}

// Name returns the role name.
func (r *Role) Name() string {
	return r.name
}

// Description returns the role description.
func (r *Role) Description() string {
	return r.description
}

// Permissions returns the role's permission set.
func (r *Role) Permissions() *PermissionSet {
	return r.permissions
}

// IsSystem returns true if this is a system role.
func (r *Role) IsSystem() bool {
	return r.isSystem
}

// IsTenantRole returns true if this is a tenant-specific role.
func (r *Role) IsTenantRole() bool {
	return r.tenantID != nil
}

// Behaviors

// UpdateDetails updates the role's details.
func (r *Role) UpdateDetails(name, description string) error {
	if r.isSystem {
		return ErrCannotModifySystemRole
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return ErrRoleNameRequired
	}

	if len(name) > 100 {
		return ErrRoleNameTooLong
	}

	r.name = name
	r.description = description
	r.MarkUpdated()

	r.AddDomainEvent(NewRoleUpdatedEvent(r))

	return nil
}

// AddPermission adds a permission to the role.
func (r *Role) AddPermission(permission Permission) error {
	if r.isSystem {
		return ErrCannotModifySystemRole
	}

	r.permissions.Add(permission)
	r.MarkUpdated()

	r.AddDomainEvent(NewRolePermissionAddedEvent(r, permission))

	return nil
}

// AddPermissions adds multiple permissions to the role.
func (r *Role) AddPermissions(permissions ...Permission) error {
	if r.isSystem {
		return ErrCannotModifySystemRole
	}

	for _, p := range permissions {
		r.permissions.Add(p)
	}
	r.MarkUpdated()

	return nil
}

// RemovePermission removes a permission from the role.
func (r *Role) RemovePermission(permission Permission) error {
	if r.isSystem {
		return ErrCannotModifySystemRole
	}

	r.permissions.Remove(permission)
	r.MarkUpdated()

	r.AddDomainEvent(NewRolePermissionRemovedEvent(r, permission))

	return nil
}

// SetPermissions replaces all permissions.
func (r *Role) SetPermissions(permissions *PermissionSet) error {
	if r.isSystem {
		return ErrCannotModifySystemRole
	}

	r.permissions = permissions
	r.MarkUpdated()

	return nil
}

// HasPermission checks if the role has a specific permission.
func (r *Role) HasPermission(permission Permission) bool {
	return r.permissions.HasPermission(permission)
}

// Delete soft deletes the role.
func (r *Role) Delete() error {
	if r.isSystem {
		return ErrCannotDeleteSystemRole
	}

	if r.IsDeleted() {
		return nil
	}

	r.MarkDeleted()

	r.AddDomainEvent(NewRoleDeletedEvent(r))

	return nil
}

// Predefined system roles creation helpers

// CreateSuperAdminRole creates the super admin system role.
func CreateSuperAdminRole() *Role {
	permissions := NewPermissionSet()
	permissions.Add(PermissionFullAccess)

	role, _ := NewSystemRole(
		RoleNameSuperAdmin,
		"Super Administrator with full access",
		permissions,
	)
	return role
}

// CreateAdminRole creates the admin system role.
func CreateAdminRole() *Role {
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersAll)
	permissions.Add(PermissionRolesAll)
	permissions.Add(PermissionSettingsAll)

	role, _ := NewSystemRole(
		RoleNameAdmin,
		"Administrator",
		permissions,
	)
	return role
}

// CreateManagerRole creates the manager system role.
func CreateManagerRole() *Role {
	permissions := NewPermissionSet()
	permissions.Add(PermissionUsersRead)
	permissions.Add(PermissionCustomersAll)
	permissions.Add(MustNewPermission(ResourceDeals, ActionAll))

	role, _ := NewSystemRole(
		RoleNameManager,
		"Manager",
		permissions,
	)
	return role
}

// CreateSalesRepRole creates the sales rep system role.
func CreateSalesRepRole() *Role {
	permissions := NewPermissionSet()
	permissions.Add(PermissionCustomersRead)
	permissions.Add(PermissionCustomersCreate)
	permissions.Add(MustNewPermission(ResourceCustomers, ActionUpdate))
	permissions.Add(PermissionLeadsAll)
	permissions.Add(PermissionOpportunitiesAll)

	role, _ := NewSystemRole(
		RoleNameSalesRep,
		"Sales Representative",
		permissions,
	)
	return role
}

// CreateViewerRole creates the viewer system role.
func CreateViewerRole() *Role {
	permissions := NewPermissionSet()
	permissions.Add(MustNewPermission(ResourceAll, ActionRead))

	role, _ := NewSystemRole(
		RoleNameViewer,
		"Read-only access",
		permissions,
	)
	return role
}

// Role errors
var (
	ErrRoleNameRequired       = fmt.Errorf("role name is required")
	ErrRoleNameTooLong        = fmt.Errorf("role name exceeds maximum length of 100 characters")
	ErrCannotModifySystemRole = fmt.Errorf("cannot modify system role")
	ErrCannotDeleteSystemRole = fmt.Errorf("cannot delete system role")
	ErrRoleDuplicate          = fmt.Errorf("role with this name already exists")
)
