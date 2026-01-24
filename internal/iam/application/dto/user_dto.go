// Package dto contains Data Transfer Objects for the application layer.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// User DTOs
// ============================================================================

// RegisterUserRequest represents a user registration request.
type RegisterUserRequest struct {
	TenantID  uuid.UUID `json:"tenant_id" validate:"required"`
	Email     string    `json:"email" validate:"required,email,max=255"`
	Password  string    `json:"password" validate:"required,min=8,max=128"`
	FirstName string    `json:"first_name" validate:"max=100"`
	LastName  string    `json:"last_name" validate:"max=100"`
	Phone     string    `json:"phone" validate:"max=50"`
}

// RegisterUserResponse represents a user registration response.
type RegisterUserResponse struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"access_token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	ExpiresAt    int64    `json:"expires_at,omitempty"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	TenantSlug string `json:"tenant_slug" validate:"required,max=100"`
	Email      string `json:"email" validate:"required,email,max=255"`
	Password   string `json:"password" validate:"required"`
	DeviceInfo *DeviceInfoDTO `json:"device_info,omitempty"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresAt    int64    `json:"expires_at"`
	TokenType    string   `json:"token_type"`
}

// RefreshTokenRequest represents a token refresh request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents a token refresh response.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TokenType    string `json:"token_type"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	AllDevices   bool   `json:"all_devices"`
}

// UpdateUserRequest represents an update user request.
type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"max=100"`
	LastName  string `json:"last_name" validate:"max=100"`
	Phone     string `json:"phone" validate:"max=50"`
	AvatarURL string `json:"avatar_url" validate:"max=500,omitempty,url"`
}

// UpdateUserEmailRequest represents an email change request.
type UpdateUserEmailRequest struct {
	NewEmail string `json:"new_email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=128"`
}

// ResetPasswordRequest represents a password reset request.
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordConfirmRequest represents a password reset confirmation.
type ResetPasswordConfirmRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// VerifyEmailRequest represents an email verification request.
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// AssignRoleRequest represents a role assignment request.
type AssignRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// RemoveRoleRequest represents a role removal request.
type RemoveRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// ListUsersRequest represents a list users request.
type ListUsersRequest struct {
	TenantID      uuid.UUID `json:"tenant_id" validate:"required"`
	Page          int       `json:"page" validate:"min=1"`
	PageSize      int       `json:"page_size" validate:"min=1,max=100"`
	SortBy        string    `json:"sort_by" validate:"omitempty,oneof=created_at updated_at email first_name last_name"`
	SortDirection string    `json:"sort_direction" validate:"omitempty,oneof=asc desc"`
	Status        string    `json:"status" validate:"omitempty,oneof=active inactive suspended pending"`
	Search        string    `json:"search" validate:"max=100"`
	IncludeRoles  bool      `json:"include_roles"`
}

// ListUsersResponse represents a list users response.
type ListUsersResponse struct {
	Users      []*UserDTO      `json:"users"`
	Pagination *PaginationDTO `json:"pagination"`
}

// UserDTO represents a user data transfer object.
type UserDTO struct {
	ID              uuid.UUID   `json:"id"`
	TenantID        uuid.UUID   `json:"tenant_id"`
	Email           string      `json:"email"`
	FirstName       string      `json:"first_name"`
	LastName        string      `json:"last_name"`
	FullName        string      `json:"full_name"`
	AvatarURL       string      `json:"avatar_url,omitempty"`
	Phone           string      `json:"phone,omitempty"`
	Status          string      `json:"status"`
	EmailVerifiedAt *time.Time  `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time  `json:"last_login_at,omitempty"`
	Roles           []*RoleDTO  `json:"roles,omitempty"`
	Permissions     []string    `json:"permissions,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// DeviceInfoDTO represents device information.
type DeviceInfoDTO struct {
	DeviceID       string `json:"device_id,omitempty"`
	DeviceType     string `json:"device_type,omitempty"`
	DeviceName     string `json:"device_name,omitempty"`
	OS             string `json:"os,omitempty"`
	OSVersion      string `json:"os_version,omitempty"`
	Browser        string `json:"browser,omitempty"`
	BrowserVersion string `json:"browser_version,omitempty"`
}

// PaginationDTO represents pagination information.
type PaginationDTO struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewPaginationDTO creates a new pagination DTO.
func NewPaginationDTO(page, pageSize int, totalItems int64) *PaginationDTO {
	totalPages := totalItems / int64(pageSize)
	if totalItems%int64(pageSize) > 0 {
		totalPages++
	}

	return &PaginationDTO{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    int64(page) < totalPages,
		HasPrev:    page > 1,
	}
}
