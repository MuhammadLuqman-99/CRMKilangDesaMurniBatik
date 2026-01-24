// Package dto contains Data Transfer Objects for the application layer.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Tenant DTOs
// ============================================================================

// CreateTenantRequest represents a tenant creation request.
type CreateTenantRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=255"`
	Slug  string `json:"slug" validate:"required,min=2,max=100,alphanumdash"`
	Plan  string `json:"plan" validate:"omitempty,oneof=free starter pro enterprise"`
}

// CreateTenantResponse represents a tenant creation response.
type CreateTenantResponse struct {
	Tenant *TenantDTO `json:"tenant"`
}

// CreateTenantWithAdminRequest represents creating a tenant with admin user.
type CreateTenantWithAdminRequest struct {
	// Tenant info
	TenantName string `json:"tenant_name" validate:"required,min=2,max=255"`
	TenantSlug string `json:"tenant_slug" validate:"required,min=2,max=100,alphanumdash"`
	Plan       string `json:"plan" validate:"omitempty,oneof=free starter pro enterprise"`

	// Admin user info
	AdminEmail     string `json:"admin_email" validate:"required,email,max=255"`
	AdminPassword  string `json:"admin_password" validate:"required,min=8,max=128"`
	AdminFirstName string `json:"admin_first_name" validate:"max=100"`
	AdminLastName  string `json:"admin_last_name" validate:"max=100"`
}

// CreateTenantWithAdminResponse represents the response for creating tenant with admin.
type CreateTenantWithAdminResponse struct {
	Tenant       *TenantDTO `json:"tenant"`
	AdminUser    *UserDTO   `json:"admin_user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	ExpiresAt    int64      `json:"expires_at"`
}

// UpdateTenantRequest represents a tenant update request.
type UpdateTenantRequest struct {
	Name string `json:"name" validate:"min=2,max=255"`
}

// UpdateTenantResponse represents a tenant update response.
type UpdateTenantResponse struct {
	Tenant *TenantDTO `json:"tenant"`
}

// UpdateTenantSettingsRequest represents a tenant settings update request.
type UpdateTenantSettingsRequest struct {
	Timezone           *string `json:"timezone" validate:"omitempty,max=50"`
	DateFormat         *string `json:"date_format" validate:"omitempty,max=20"`
	Currency           *string `json:"currency" validate:"omitempty,len=3"`
	Language           *string `json:"language" validate:"omitempty,len=2"`
	NotificationsEmail *bool   `json:"notifications_email"`
}

// UpdateTenantPlanRequest represents a plan change request.
type UpdateTenantPlanRequest struct {
	Plan string `json:"plan" validate:"required,oneof=free starter pro enterprise"`
}

// ActivateTenantRequest represents a tenant activation request.
type ActivateTenantRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

// SuspendTenantRequest represents a tenant suspension request.
type SuspendTenantRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Reason   string    `json:"reason" validate:"required,max=500"`
}

// StartTrialRequest represents a trial start request.
type StartTrialRequest struct {
	TenantID  uuid.UUID `json:"tenant_id" validate:"required"`
	TrialDays int       `json:"trial_days" validate:"min=1,max=90"`
}

// ListTenantsRequest represents a list tenants request.
type ListTenantsRequest struct {
	Page          int    `json:"page" validate:"min=1"`
	PageSize      int    `json:"page_size" validate:"min=1,max=100"`
	SortBy        string `json:"sort_by" validate:"omitempty,oneof=created_at name slug"`
	SortDirection string `json:"sort_direction" validate:"omitempty,oneof=asc desc"`
	Status        string `json:"status" validate:"omitempty,oneof=active inactive suspended pending trial"`
	Plan          string `json:"plan" validate:"omitempty,oneof=free starter pro enterprise"`
	Search        string `json:"search" validate:"max=100"`
}

// ListTenantsResponse represents a list tenants response.
type ListTenantsResponse struct {
	Tenants    []*TenantDTO   `json:"tenants"`
	Pagination *PaginationDTO `json:"pagination"`
}

// GetTenantResponse represents a get tenant response.
type GetTenantResponse struct {
	Tenant *TenantDTO `json:"tenant"`
}

// TenantDTO represents a tenant data transfer object.
type TenantDTO struct {
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name"`
	Slug      string             `json:"slug"`
	Status    string             `json:"status"`
	Plan      string             `json:"plan"`
	Settings  *TenantSettingsDTO `json:"settings"`
	Limits    *TenantLimitsDTO   `json:"limits,omitempty"`
	Usage     *TenantUsageDTO    `json:"usage,omitempty"`
	TrialInfo *TrialInfoDTO      `json:"trial_info,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// TenantSettingsDTO represents tenant settings.
type TenantSettingsDTO struct {
	Timezone           string `json:"timezone"`
	DateFormat         string `json:"date_format"`
	Currency           string `json:"currency"`
	Language           string `json:"language"`
	NotificationsEmail bool   `json:"notifications_email"`
}

// TenantLimitsDTO represents tenant plan limits.
type TenantLimitsDTO struct {
	MaxUsers    int `json:"max_users"`
	MaxContacts int `json:"max_contacts"`
}

// TenantUsageDTO represents tenant usage statistics.
type TenantUsageDTO struct {
	UserCount    int64 `json:"user_count"`
	ContactCount int64 `json:"contact_count"`
}

// TrialInfoDTO represents trial information.
type TrialInfoDTO struct {
	IsTrialing   bool       `json:"is_trialing"`
	TrialStarted *time.Time `json:"trial_started,omitempty"`
	TrialEnds    *time.Time `json:"trial_ends,omitempty"`
	DaysLeft     int        `json:"days_left,omitempty"`
}

// TenantStatsDTO represents tenant statistics.
type TenantStatsDTO struct {
	TotalTenants   int64 `json:"total_tenants"`
	ActiveTenants  int64 `json:"active_tenants"`
	TrialTenants   int64 `json:"trial_tenants"`
	PendingTenants int64 `json:"pending_tenants"`
	PlanBreakdown  map[string]int64 `json:"plan_breakdown"`
}
