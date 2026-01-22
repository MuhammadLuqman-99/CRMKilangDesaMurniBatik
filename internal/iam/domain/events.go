// Package domain contains the domain layer for the IAM service.
package domain

import (
	"github.com/google/uuid"
)

// Event type constants
const (
	// User events
	EventTypeUserCreated         = "user.created"
	EventTypeUserUpdated         = "user.updated"
	EventTypeUserDeleted         = "user.deleted"
	EventTypeUserActivated       = "user.activated"
	EventTypeUserDeactivated     = "user.deactivated"
	EventTypeUserSuspended       = "user.suspended"
	EventTypeUserEmailChanged    = "user.email_changed"
	EventTypeUserEmailVerified   = "user.email_verified"
	EventTypeUserPasswordChanged = "user.password_changed"
	EventTypeUserLoggedIn        = "user.logged_in"
	EventTypeUserRoleAssigned    = "user.role_assigned"
	EventTypeUserRoleRemoved     = "user.role_removed"

	// Role events
	EventTypeRoleCreated           = "role.created"
	EventTypeRoleUpdated           = "role.updated"
	EventTypeRoleDeleted           = "role.deleted"
	EventTypeRolePermissionAdded   = "role.permission_added"
	EventTypeRolePermissionRemoved = "role.permission_removed"

	// Tenant events
	EventTypeTenantCreated         = "tenant.created"
	EventTypeTenantUpdated         = "tenant.updated"
	EventTypeTenantDeleted         = "tenant.deleted"
	EventTypeTenantActivated       = "tenant.activated"
	EventTypeTenantDeactivated     = "tenant.deactivated"
	EventTypeTenantSuspended       = "tenant.suspended"
	EventTypeTenantTrialStarted    = "tenant.trial_started"
	EventTypeTenantPlanChanged     = "tenant.plan_changed"
	EventTypeTenantSettingsUpdated = "tenant.settings_updated"
)

// Aggregate type constants
const (
	AggregateTypeUser   = "user"
	AggregateTypeRole   = "role"
	AggregateTypeTenant = "tenant"
)

// ============================================================================
// User Events
// ============================================================================

// UserCreatedEvent is raised when a new user is created.
type UserCreatedEvent struct {
	BaseDomainEvent
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

// NewUserCreatedEvent creates a new UserCreatedEvent.
func NewUserCreatedEvent(user *User) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserCreated, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
		Email:           user.Email().String(),
		FirstName:       user.FirstName(),
		LastName:        user.LastName(),
	}
}

// UserUpdatedEvent is raised when a user is updated.
type UserUpdatedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
}

// NewUserUpdatedEvent creates a new UserUpdatedEvent.
func NewUserUpdatedEvent(user *User) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserUpdated, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
	}
}

// UserDeletedEvent is raised when a user is deleted.
type UserDeletedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
}

// NewUserDeletedEvent creates a new UserDeletedEvent.
func NewUserDeletedEvent(user *User) *UserDeletedEvent {
	return &UserDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserDeleted, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
		Email:           user.Email().String(),
	}
}

// UserActivatedEvent is raised when a user is activated.
type UserActivatedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
}

// NewUserActivatedEvent creates a new UserActivatedEvent.
func NewUserActivatedEvent(user *User) *UserActivatedEvent {
	return &UserActivatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserActivated, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
	}
}

// UserDeactivatedEvent is raised when a user is deactivated.
type UserDeactivatedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
}

// NewUserDeactivatedEvent creates a new UserDeactivatedEvent.
func NewUserDeactivatedEvent(user *User) *UserDeactivatedEvent {
	return &UserDeactivatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserDeactivated, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
	}
}

// UserSuspendedEvent is raised when a user is suspended.
type UserSuspendedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
	Reason   string    `json:"reason"`
}

// NewUserSuspendedEvent creates a new UserSuspendedEvent.
func NewUserSuspendedEvent(user *User, reason string) *UserSuspendedEvent {
	return &UserSuspendedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserSuspended, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
		Reason:          reason,
	}
}

// UserEmailChangedEvent is raised when a user's email is changed.
type UserEmailChangedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
	OldEmail string    `json:"old_email"`
	NewEmail string    `json:"new_email"`
}

// NewUserEmailChangedEvent creates a new UserEmailChangedEvent.
func NewUserEmailChangedEvent(user *User, oldEmail, newEmail Email) *UserEmailChangedEvent {
	return &UserEmailChangedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserEmailChanged, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
		OldEmail:        oldEmail.String(),
		NewEmail:        newEmail.String(),
	}
}

// UserEmailVerifiedEvent is raised when a user's email is verified.
type UserEmailVerifiedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
}

// NewUserEmailVerifiedEvent creates a new UserEmailVerifiedEvent.
func NewUserEmailVerifiedEvent(user *User) *UserEmailVerifiedEvent {
	return &UserEmailVerifiedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserEmailVerified, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
		Email:           user.Email().String(),
	}
}

// UserPasswordChangedEvent is raised when a user's password is changed.
type UserPasswordChangedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
}

// NewUserPasswordChangedEvent creates a new UserPasswordChangedEvent.
func NewUserPasswordChangedEvent(user *User) *UserPasswordChangedEvent {
	return &UserPasswordChangedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserPasswordChanged, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
	}
}

// UserLoggedInEvent is raised when a user logs in.
type UserLoggedInEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
}

// NewUserLoggedInEvent creates a new UserLoggedInEvent.
func NewUserLoggedInEvent(user *User) *UserLoggedInEvent {
	return &UserLoggedInEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserLoggedIn, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
		Email:           user.Email().String(),
	}
}

// UserRoleAssignedEvent is raised when a role is assigned to a user.
type UserRoleAssignedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleName string    `json:"role_name"`
}

// NewUserRoleAssignedEvent creates a new UserRoleAssignedEvent.
func NewUserRoleAssignedEvent(user *User, role *Role) *UserRoleAssignedEvent {
	return &UserRoleAssignedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserRoleAssigned, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
		RoleID:          role.GetID(),
		RoleName:        role.Name(),
	}
}

// UserRoleRemovedEvent is raised when a role is removed from a user.
type UserRoleRemovedEvent struct {
	BaseDomainEvent
	TenantID uuid.UUID `json:"tenant_id"`
	RoleID   uuid.UUID `json:"role_id"`
	RoleName string    `json:"role_name"`
}

// NewUserRoleRemovedEvent creates a new UserRoleRemovedEvent.
func NewUserRoleRemovedEvent(user *User, role *Role) *UserRoleRemovedEvent {
	return &UserRoleRemovedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeUserRoleRemoved, user.GetID(), AggregateTypeUser),
		TenantID:        user.TenantID(),
		RoleID:          role.GetID(),
		RoleName:        role.Name(),
	}
}

// ============================================================================
// Role Events
// ============================================================================

// RoleCreatedEvent is raised when a new role is created.
type RoleCreatedEvent struct {
	BaseDomainEvent
	TenantID    *uuid.UUID `json:"tenant_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IsSystem    bool       `json:"is_system"`
}

// NewRoleCreatedEvent creates a new RoleCreatedEvent.
func NewRoleCreatedEvent(role *Role) *RoleCreatedEvent {
	return &RoleCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeRoleCreated, role.GetID(), AggregateTypeRole),
		TenantID:        role.TenantID(),
		Name:            role.Name(),
		Description:     role.Description(),
		IsSystem:        role.IsSystem(),
	}
}

// RoleUpdatedEvent is raised when a role is updated.
type RoleUpdatedEvent struct {
	BaseDomainEvent
	TenantID *uuid.UUID `json:"tenant_id,omitempty"`
	Name     string     `json:"name"`
}

// NewRoleUpdatedEvent creates a new RoleUpdatedEvent.
func NewRoleUpdatedEvent(role *Role) *RoleUpdatedEvent {
	return &RoleUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeRoleUpdated, role.GetID(), AggregateTypeRole),
		TenantID:        role.TenantID(),
		Name:            role.Name(),
	}
}

// RoleDeletedEvent is raised when a role is deleted.
type RoleDeletedEvent struct {
	BaseDomainEvent
	TenantID *uuid.UUID `json:"tenant_id,omitempty"`
	Name     string     `json:"name"`
}

// NewRoleDeletedEvent creates a new RoleDeletedEvent.
func NewRoleDeletedEvent(role *Role) *RoleDeletedEvent {
	return &RoleDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeRoleDeleted, role.GetID(), AggregateTypeRole),
		TenantID:        role.TenantID(),
		Name:            role.Name(),
	}
}

// RolePermissionAddedEvent is raised when a permission is added to a role.
type RolePermissionAddedEvent struct {
	BaseDomainEvent
	TenantID   *uuid.UUID `json:"tenant_id,omitempty"`
	RoleName   string     `json:"role_name"`
	Permission string     `json:"permission"`
}

// NewRolePermissionAddedEvent creates a new RolePermissionAddedEvent.
func NewRolePermissionAddedEvent(role *Role, permission Permission) *RolePermissionAddedEvent {
	return &RolePermissionAddedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeRolePermissionAdded, role.GetID(), AggregateTypeRole),
		TenantID:        role.TenantID(),
		RoleName:        role.Name(),
		Permission:      permission.String(),
	}
}

// RolePermissionRemovedEvent is raised when a permission is removed from a role.
type RolePermissionRemovedEvent struct {
	BaseDomainEvent
	TenantID   *uuid.UUID `json:"tenant_id,omitempty"`
	RoleName   string     `json:"role_name"`
	Permission string     `json:"permission"`
}

// NewRolePermissionRemovedEvent creates a new RolePermissionRemovedEvent.
func NewRolePermissionRemovedEvent(role *Role, permission Permission) *RolePermissionRemovedEvent {
	return &RolePermissionRemovedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeRolePermissionRemoved, role.GetID(), AggregateTypeRole),
		TenantID:        role.TenantID(),
		RoleName:        role.Name(),
		Permission:      permission.String(),
	}
}

// ============================================================================
// Tenant Events
// ============================================================================

// TenantCreatedEvent is raised when a new tenant is created.
type TenantCreatedEvent struct {
	BaseDomainEvent
	Name string     `json:"name"`
	Slug string     `json:"slug"`
	Plan TenantPlan `json:"plan"`
}

// NewTenantCreatedEvent creates a new TenantCreatedEvent.
func NewTenantCreatedEvent(tenant *Tenant) *TenantCreatedEvent {
	return &TenantCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantCreated, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
		Slug:            tenant.Slug(),
		Plan:            tenant.Plan(),
	}
}

// TenantUpdatedEvent is raised when a tenant is updated.
type TenantUpdatedEvent struct {
	BaseDomainEvent
	Name string `json:"name"`
}

// NewTenantUpdatedEvent creates a new TenantUpdatedEvent.
func NewTenantUpdatedEvent(tenant *Tenant) *TenantUpdatedEvent {
	return &TenantUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantUpdated, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
	}
}

// TenantDeletedEvent is raised when a tenant is deleted.
type TenantDeletedEvent struct {
	BaseDomainEvent
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// NewTenantDeletedEvent creates a new TenantDeletedEvent.
func NewTenantDeletedEvent(tenant *Tenant) *TenantDeletedEvent {
	return &TenantDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantDeleted, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
		Slug:            tenant.Slug(),
	}
}

// TenantActivatedEvent is raised when a tenant is activated.
type TenantActivatedEvent struct {
	BaseDomainEvent
	Name string `json:"name"`
}

// NewTenantActivatedEvent creates a new TenantActivatedEvent.
func NewTenantActivatedEvent(tenant *Tenant) *TenantActivatedEvent {
	return &TenantActivatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantActivated, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
	}
}

// TenantDeactivatedEvent is raised when a tenant is deactivated.
type TenantDeactivatedEvent struct {
	BaseDomainEvent
	Name string `json:"name"`
}

// NewTenantDeactivatedEvent creates a new TenantDeactivatedEvent.
func NewTenantDeactivatedEvent(tenant *Tenant) *TenantDeactivatedEvent {
	return &TenantDeactivatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantDeactivated, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
	}
}

// TenantSuspendedEvent is raised when a tenant is suspended.
type TenantSuspendedEvent struct {
	BaseDomainEvent
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// NewTenantSuspendedEvent creates a new TenantSuspendedEvent.
func NewTenantSuspendedEvent(tenant *Tenant, reason string) *TenantSuspendedEvent {
	return &TenantSuspendedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantSuspended, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
		Reason:          reason,
	}
}

// TenantTrialStartedEvent is raised when a tenant's trial period starts.
type TenantTrialStartedEvent struct {
	BaseDomainEvent
	Name         string `json:"name"`
	TrialDays    int    `json:"trial_days"`
}

// NewTenantTrialStartedEvent creates a new TenantTrialStartedEvent.
func NewTenantTrialStartedEvent(tenant *Tenant, trialDays int) *TenantTrialStartedEvent {
	return &TenantTrialStartedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantTrialStarted, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
		TrialDays:       trialDays,
	}
}

// TenantPlanChangedEvent is raised when a tenant's plan changes.
type TenantPlanChangedEvent struct {
	BaseDomainEvent
	Name    string     `json:"name"`
	OldPlan TenantPlan `json:"old_plan"`
	NewPlan TenantPlan `json:"new_plan"`
}

// NewTenantPlanChangedEvent creates a new TenantPlanChangedEvent.
func NewTenantPlanChangedEvent(tenant *Tenant, oldPlan, newPlan TenantPlan) *TenantPlanChangedEvent {
	return &TenantPlanChangedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantPlanChanged, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
		OldPlan:         oldPlan,
		NewPlan:         newPlan,
	}
}

// TenantSettingsUpdatedEvent is raised when a tenant's settings are updated.
type TenantSettingsUpdatedEvent struct {
	BaseDomainEvent
	Name string `json:"name"`
}

// NewTenantSettingsUpdatedEvent creates a new TenantSettingsUpdatedEvent.
func NewTenantSettingsUpdatedEvent(tenant *Tenant) *TenantSettingsUpdatedEvent {
	return &TenantSettingsUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(EventTypeTenantSettingsUpdated, tenant.GetID(), AggregateTypeTenant),
		Name:            tenant.Name(),
	}
}
