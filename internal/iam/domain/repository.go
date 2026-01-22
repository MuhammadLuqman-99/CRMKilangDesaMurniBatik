// Package domain contains the domain layer for the IAM service.
package domain

import (
	"context"

	"github.com/google/uuid"
)

// ============================================================================
// User Repository
// ============================================================================

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// Create creates a new user.
	Create(ctx context.Context, user *User) error

	// Update updates an existing user.
	Update(ctx context.Context, user *User) error

	// Delete soft deletes a user.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a user by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)

	// FindByEmail finds a user by email within a tenant.
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email Email) (*User, error)

	// FindByTenant finds all users for a tenant with pagination.
	FindByTenant(ctx context.Context, tenantID uuid.UUID, opts UserQueryOptions) ([]*User, int64, error)

	// FindByRoleID finds all users with a specific role.
	FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]*User, error)

	// ExistsByEmail checks if a user with the email exists in the tenant.
	ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email Email) (bool, error)

	// CountByTenant returns the number of users for a tenant.
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// UserQueryOptions defines options for querying users.
type UserQueryOptions struct {
	Page          int
	PageSize      int
	SortBy        string
	SortDirection string
	Status        *UserStatus
	Search        string
	IncludeRoles  bool
}

// DefaultUserQueryOptions returns default query options.
func DefaultUserQueryOptions() UserQueryOptions {
	return UserQueryOptions{
		Page:          1,
		PageSize:      20,
		SortBy:        "created_at",
		SortDirection: "desc",
		IncludeRoles:  true,
	}
}

// ============================================================================
// Role Repository
// ============================================================================

// RoleRepository defines the interface for role persistence operations.
type RoleRepository interface {
	// Create creates a new role.
	Create(ctx context.Context, role *Role) error

	// Update updates an existing role.
	Update(ctx context.Context, role *Role) error

	// Delete soft deletes a role.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a role by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Role, error)

	// FindByName finds a role by name within a tenant.
	FindByName(ctx context.Context, tenantID *uuid.UUID, name string) (*Role, error)

	// FindByTenant finds all roles for a tenant (including system roles).
	FindByTenant(ctx context.Context, tenantID uuid.UUID, opts RoleQueryOptions) ([]*Role, int64, error)

	// FindSystemRoles finds all system roles.
	FindSystemRoles(ctx context.Context) ([]*Role, error)

	// FindByUserID finds all roles assigned to a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Role, error)

	// ExistsByName checks if a role with the name exists in the tenant.
	ExistsByName(ctx context.Context, tenantID *uuid.UUID, name string) (bool, error)

	// AssignRoleToUser assigns a role to a user.
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error

	// RemoveRoleFromUser removes a role from a user.
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
}

// RoleQueryOptions defines options for querying roles.
type RoleQueryOptions struct {
	Page           int
	PageSize       int
	SortBy         string
	SortDirection  string
	IncludeSystem  bool
	Search         string
}

// DefaultRoleQueryOptions returns default query options.
func DefaultRoleQueryOptions() RoleQueryOptions {
	return RoleQueryOptions{
		Page:          1,
		PageSize:      20,
		SortBy:        "name",
		SortDirection: "asc",
		IncludeSystem: true,
	}
}

// ============================================================================
// Tenant Repository
// ============================================================================

// TenantRepository defines the interface for tenant persistence operations.
type TenantRepository interface {
	// Create creates a new tenant.
	Create(ctx context.Context, tenant *Tenant) error

	// Update updates an existing tenant.
	Update(ctx context.Context, tenant *Tenant) error

	// Delete soft deletes a tenant.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a tenant by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Tenant, error)

	// FindBySlug finds a tenant by slug.
	FindBySlug(ctx context.Context, slug string) (*Tenant, error)

	// FindAll finds all tenants with pagination.
	FindAll(ctx context.Context, opts TenantQueryOptions) ([]*Tenant, int64, error)

	// ExistsBySlug checks if a tenant with the slug exists.
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// Count returns the total number of tenants.
	Count(ctx context.Context) (int64, error)
}

// TenantQueryOptions defines options for querying tenants.
type TenantQueryOptions struct {
	Page          int
	PageSize      int
	SortBy        string
	SortDirection string
	Status        *TenantStatus
	Plan          *TenantPlan
	Search        string
}

// DefaultTenantQueryOptions returns default query options.
func DefaultTenantQueryOptions() TenantQueryOptions {
	return TenantQueryOptions{
		Page:          1,
		PageSize:      20,
		SortBy:        "created_at",
		SortDirection: "desc",
	}
}

// ============================================================================
// Refresh Token Repository
// ============================================================================

// RefreshTokenRepository defines the interface for refresh token persistence operations.
type RefreshTokenRepository interface {
	// Create creates a new refresh token.
	Create(ctx context.Context, token *RefreshToken) error

	// Update updates an existing refresh token.
	Update(ctx context.Context, token *RefreshToken) error

	// Delete deletes a refresh token.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a refresh token by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*RefreshToken, error)

	// FindByTokenHash finds a refresh token by its hash.
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// FindByUserID finds all refresh tokens for a user.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error)

	// FindActiveByUserID finds all active (non-revoked, non-expired) refresh tokens for a user.
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error)

	// RevokeByUserID revokes all refresh tokens for a user.
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error

	// RevokeByID revokes a specific refresh token.
	RevokeByID(ctx context.Context, id uuid.UUID) error

	// DeleteExpired deletes all expired refresh tokens.
	DeleteExpired(ctx context.Context) (int64, error)

	// CountActiveByUserID counts active refresh tokens for a user.
	CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

// ============================================================================
// Audit Log Repository
// ============================================================================

// AuditLogEntry represents an audit log entry.
type AuditLogEntry struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	OldValues  map[string]interface{}
	NewValues  map[string]interface{}
	IPAddress  string
	UserAgent  string
	CreatedAt  interface{} // time.Time
}

// AuditLogRepository defines the interface for audit log persistence operations.
type AuditLogRepository interface {
	// Create creates a new audit log entry.
	Create(ctx context.Context, entry *AuditLogEntry) error

	// FindByTenant finds audit logs for a tenant with pagination.
	FindByTenant(ctx context.Context, tenantID uuid.UUID, opts AuditLogQueryOptions) ([]*AuditLogEntry, int64, error)

	// FindByEntity finds audit logs for a specific entity.
	FindByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*AuditLogEntry, error)

	// FindByUser finds audit logs for a specific user.
	FindByUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, opts AuditLogQueryOptions) ([]*AuditLogEntry, int64, error)
}

// AuditLogQueryOptions defines options for querying audit logs.
type AuditLogQueryOptions struct {
	Page          int
	PageSize      int
	SortDirection string
	Action        string
	EntityType    string
	StartDate     interface{} // time.Time
	EndDate       interface{} // time.Time
}

// DefaultAuditLogQueryOptions returns default query options.
func DefaultAuditLogQueryOptions() AuditLogQueryOptions {
	return AuditLogQueryOptions{
		Page:          1,
		PageSize:      50,
		SortDirection: "desc",
	}
}

// ============================================================================
// Outbox Repository (for Transactional Outbox Pattern)
// ============================================================================

// OutboxEntry represents an outbox entry for domain events.
type OutboxEntry struct {
	ID            uuid.UUID
	EventType     string
	AggregateID   uuid.UUID
	AggregateType string
	Payload       []byte
	Published     bool
	PublishedAt   interface{} // *time.Time
	CreatedAt     interface{} // time.Time
}

// OutboxRepository defines the interface for outbox persistence operations.
type OutboxRepository interface {
	// Create creates a new outbox entry.
	Create(ctx context.Context, entry *OutboxEntry) error

	// MarkAsPublished marks an outbox entry as published.
	MarkAsPublished(ctx context.Context, id uuid.UUID) error

	// FindUnpublished finds all unpublished outbox entries.
	FindUnpublished(ctx context.Context, limit int) ([]*OutboxEntry, error)

	// DeletePublished deletes published outbox entries older than the specified time.
	DeletePublished(ctx context.Context, olderThan interface{}) (int64, error)
}

// ============================================================================
// Unit of Work (for transaction management)
// ============================================================================

// UnitOfWork defines the interface for managing transactions.
type UnitOfWork interface {
	// Begin starts a new transaction.
	Begin(ctx context.Context) (context.Context, error)

	// Commit commits the current transaction.
	Commit(ctx context.Context) error

	// Rollback rolls back the current transaction.
	Rollback(ctx context.Context) error

	// UserRepository returns the user repository.
	UserRepository() UserRepository

	// RoleRepository returns the role repository.
	RoleRepository() RoleRepository

	// TenantRepository returns the tenant repository.
	TenantRepository() TenantRepository

	// RefreshTokenRepository returns the refresh token repository.
	RefreshTokenRepository() RefreshTokenRepository

	// AuditLogRepository returns the audit log repository.
	AuditLogRepository() AuditLogRepository

	// OutboxRepository returns the outbox repository.
	OutboxRepository() OutboxRepository
}
