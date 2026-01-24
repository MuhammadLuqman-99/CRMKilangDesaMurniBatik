// Package dto contains Data Transfer Objects for the application layer.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Role DTOs
// ============================================================================

// CreateRoleRequest represents a role creation request.
type CreateRoleRequest struct {
	TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=2,max=100"`
	Description string    `json:"description" validate:"max=500"`
	Permissions []string  `json:"permissions" validate:"dive,required"`
}

// CreateRoleResponse represents a role creation response.
type CreateRoleResponse struct {
	Role *RoleDTO `json:"role"`
}

// UpdateRoleRequest represents a role update request.
type UpdateRoleRequest struct {
	Name        string   `json:"name" validate:"min=2,max=100"`
	Description string   `json:"description" validate:"max=500"`
	Permissions []string `json:"permissions" validate:"dive,required"`
}

// UpdateRoleResponse represents a role update response.
type UpdateRoleResponse struct {
	Role *RoleDTO `json:"role"`
}

// ListRolesRequest represents a list roles request.
type ListRolesRequest struct {
	TenantID      uuid.UUID `json:"tenant_id" validate:"required"`
	Page          int       `json:"page" validate:"min=1"`
	PageSize      int       `json:"page_size" validate:"min=1,max=100"`
	SortBy        string    `json:"sort_by" validate:"omitempty,oneof=created_at name"`
	SortDirection string    `json:"sort_direction" validate:"omitempty,oneof=asc desc"`
	IncludeSystem bool      `json:"include_system"`
	Search        string    `json:"search" validate:"max=100"`
}

// ListRolesResponse represents a list roles response.
type ListRolesResponse struct {
	Roles      []*RoleDTO     `json:"roles"`
	Pagination *PaginationDTO `json:"pagination"`
}

// GetRoleResponse represents a get role response.
type GetRoleResponse struct {
	Role *RoleDTO `json:"role"`
}

// AddPermissionRequest represents a request to add permission to role.
type AddPermissionRequest struct {
	Permission string `json:"permission" validate:"required"`
}

// RemovePermissionRequest represents a request to remove permission from role.
type RemovePermissionRequest struct {
	Permission string `json:"permission" validate:"required"`
}

// RoleDTO represents a role data transfer object.
type RoleDTO struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    *uuid.UUID `json:"tenant_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Permissions []string   `json:"permissions"`
	IsSystem    bool       `json:"is_system"`
	UserCount   int64      `json:"user_count,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PermissionDTO represents a permission data transfer object.
type PermissionDTO struct {
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
}

// PermissionGroupDTO represents a group of permissions.
type PermissionGroupDTO struct {
	Resource    string           `json:"resource"`
	Description string           `json:"description"`
	Permissions []*PermissionDTO `json:"permissions"`
}

// ListAvailablePermissionsResponse represents available permissions.
type ListAvailablePermissionsResponse struct {
	Groups []*PermissionGroupDTO `json:"groups"`
}
