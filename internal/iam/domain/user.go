// Package domain contains the domain layer for the IAM service.
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

// IsValid returns true if the status is valid.
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended, UserStatusPending:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status.
func (s UserStatus) String() string {
	return string(s)
}

// User represents a user entity in the IAM domain.
type User struct {
	BaseAggregateRoot
	tenantID        uuid.UUID
	email           Email
	passwordHash    Password
	firstName       string
	lastName        string
	avatarURL       string
	phone           string
	status          UserStatus
	emailVerifiedAt *time.Time
	lastLoginAt     *time.Time
	metadata        map[string]interface{}
	roles           []*Role
}

// NewUser creates a new User entity.
func NewUser(tenantID uuid.UUID, email Email, password Password, firstName, lastName string) (*User, error) {
	if tenantID == uuid.Nil {
		return nil, ErrUserTenantRequired
	}

	if email.IsEmpty() {
		return nil, ErrUserEmailRequired
	}

	if password.IsEmpty() {
		return nil, ErrUserPasswordRequired
	}

	user := &User{
		BaseAggregateRoot: NewBaseAggregateRoot(),
		tenantID:          tenantID,
		email:             email,
		passwordHash:      password,
		firstName:         firstName,
		lastName:          lastName,
		status:            UserStatusPending,
		metadata:          make(map[string]interface{}),
		roles:             make([]*Role, 0),
	}

	// Add domain event
	user.AddDomainEvent(NewUserCreatedEvent(user))

	return user, nil
}

// ReconstructUser reconstructs a User entity from persistence.
// This should only be used by repositories when loading from database.
func ReconstructUser(
	id uuid.UUID,
	tenantID uuid.UUID,
	email Email,
	passwordHash Password,
	firstName, lastName, avatarURL, phone string,
	status UserStatus,
	emailVerifiedAt, lastLoginAt *time.Time,
	metadata map[string]interface{},
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *User {
	user := &User{
		BaseAggregateRoot: BaseAggregateRoot{
			BaseEntity: BaseEntity{
				ID:        id,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				DeletedAt: deletedAt,
			},
			domainEvents: make([]DomainEvent, 0),
		},
		tenantID:        tenantID,
		email:           email,
		passwordHash:    passwordHash,
		firstName:       firstName,
		lastName:        lastName,
		avatarURL:       avatarURL,
		phone:           phone,
		status:          status,
		emailVerifiedAt: emailVerifiedAt,
		lastLoginAt:     lastLoginAt,
		metadata:        metadata,
		roles:           make([]*Role, 0),
	}

	if user.metadata == nil {
		user.metadata = make(map[string]interface{})
	}

	return user
}

// Getters

// TenantID returns the tenant ID.
func (u *User) TenantID() uuid.UUID {
	return u.tenantID
}

// Email returns the user's email.
func (u *User) Email() Email {
	return u.email
}

// PasswordHash returns the password hash.
func (u *User) PasswordHash() Password {
	return u.passwordHash
}

// FirstName returns the user's first name.
func (u *User) FirstName() string {
	return u.firstName
}

// LastName returns the user's last name.
func (u *User) LastName() string {
	return u.lastName
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	if u.firstName == "" && u.lastName == "" {
		return ""
	}
	if u.firstName == "" {
		return u.lastName
	}
	if u.lastName == "" {
		return u.firstName
	}
	return u.firstName + " " + u.lastName
}

// AvatarURL returns the user's avatar URL.
func (u *User) AvatarURL() string {
	return u.avatarURL
}

// Phone returns the user's phone number.
func (u *User) Phone() string {
	return u.phone
}

// Status returns the user's status.
func (u *User) Status() UserStatus {
	return u.status
}

// EmailVerifiedAt returns when the email was verified.
func (u *User) EmailVerifiedAt() *time.Time {
	return u.emailVerifiedAt
}

// LastLoginAt returns the last login timestamp.
func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}

// Metadata returns the user's metadata.
func (u *User) Metadata() map[string]interface{} {
	return u.metadata
}

// Roles returns the user's roles.
func (u *User) Roles() []*Role {
	return u.roles
}

// IsEmailVerified returns true if the user's email is verified.
func (u *User) IsEmailVerified() bool {
	return u.emailVerifiedAt != nil
}

// IsActive returns true if the user is active.
func (u *User) IsActive() bool {
	return u.status == UserStatusActive
}

// CanLogin returns true if the user can login.
func (u *User) CanLogin() bool {
	return u.status == UserStatusActive && !u.IsDeleted()
}

// Behaviors

// UpdateProfile updates the user's profile information.
func (u *User) UpdateProfile(firstName, lastName, phone, avatarURL string) {
	u.firstName = firstName
	u.lastName = lastName
	u.phone = phone
	u.avatarURL = avatarURL
	u.MarkUpdated()

	u.AddDomainEvent(NewUserUpdatedEvent(u))
}

// UpdateEmail updates the user's email address.
func (u *User) UpdateEmail(email Email) error {
	if email.IsEmpty() {
		return ErrUserEmailRequired
	}

	oldEmail := u.email
	u.email = email
	u.emailVerifiedAt = nil // Reset email verification
	u.MarkUpdated()

	u.AddDomainEvent(NewUserEmailChangedEvent(u, oldEmail, email))

	return nil
}

// ChangePassword changes the user's password.
func (u *User) ChangePassword(newPassword Password) error {
	if newPassword.IsEmpty() {
		return ErrUserPasswordRequired
	}

	u.passwordHash = newPassword
	u.MarkUpdated()

	u.AddDomainEvent(NewUserPasswordChangedEvent(u))

	return nil
}

// VerifyEmail marks the user's email as verified.
func (u *User) VerifyEmail() {
	if u.emailVerifiedAt != nil {
		return // Already verified
	}

	now := time.Now().UTC()
	u.emailVerifiedAt = &now
	u.MarkUpdated()

	// If user was pending, activate them
	if u.status == UserStatusPending {
		u.status = UserStatusActive
	}

	u.AddDomainEvent(NewUserEmailVerifiedEvent(u))
}

// RecordLogin records a successful login.
func (u *User) RecordLogin() {
	now := time.Now().UTC()
	u.lastLoginAt = &now
	u.MarkUpdated()

	u.AddDomainEvent(NewUserLoggedInEvent(u))
}

// Activate activates the user.
func (u *User) Activate() error {
	if u.status == UserStatusActive {
		return nil
	}

	if u.IsDeleted() {
		return ErrUserDeleted
	}

	u.status = UserStatusActive
	u.MarkUpdated()

	u.AddDomainEvent(NewUserActivatedEvent(u))

	return nil
}

// Deactivate deactivates the user.
func (u *User) Deactivate() error {
	if u.status == UserStatusInactive {
		return nil
	}

	u.status = UserStatusInactive
	u.MarkUpdated()

	u.AddDomainEvent(NewUserDeactivatedEvent(u))

	return nil
}

// Suspend suspends the user.
func (u *User) Suspend(reason string) error {
	if u.status == UserStatusSuspended {
		return nil
	}

	u.status = UserStatusSuspended
	u.metadata["suspend_reason"] = reason
	u.metadata["suspended_at"] = time.Now().UTC()
	u.MarkUpdated()

	u.AddDomainEvent(NewUserSuspendedEvent(u, reason))

	return nil
}

// Unsuspend removes the suspension from the user.
func (u *User) Unsuspend() error {
	if u.status != UserStatusSuspended {
		return ErrUserNotSuspended
	}

	u.status = UserStatusActive
	delete(u.metadata, "suspend_reason")
	delete(u.metadata, "suspended_at")
	u.MarkUpdated()

	return nil
}

// Delete soft deletes the user.
func (u *User) Delete() {
	if u.IsDeleted() {
		return
	}

	u.MarkDeleted()

	u.AddDomainEvent(NewUserDeletedEvent(u))
}

// Role Management

// AssignRole assigns a role to the user.
func (u *User) AssignRole(role *Role) error {
	if role == nil {
		return ErrRoleNil
	}

	// Check if role is already assigned
	for _, r := range u.roles {
		if r.GetID() == role.GetID() {
			return ErrRoleAlreadyAssigned
		}
	}

	u.roles = append(u.roles, role)
	u.MarkUpdated()

	u.AddDomainEvent(NewUserRoleAssignedEvent(u, role))

	return nil
}

// RemoveRole removes a role from the user.
func (u *User) RemoveRole(roleID uuid.UUID) error {
	for i, r := range u.roles {
		if r.GetID() == roleID {
			// Remove role from slice
			u.roles = append(u.roles[:i], u.roles[i+1:]...)
			u.MarkUpdated()

			u.AddDomainEvent(NewUserRoleRemovedEvent(u, r))

			return nil
		}
	}

	return ErrRoleNotFound
}

// HasRole checks if the user has a specific role.
func (u *User) HasRole(roleID uuid.UUID) bool {
	for _, r := range u.roles {
		if r.GetID() == roleID {
			return true
		}
	}
	return false
}

// HasRoleByName checks if the user has a role with the given name.
func (u *User) HasRoleByName(roleName string) bool {
	for _, r := range u.roles {
		if r.Name() == roleName {
			return true
		}
	}
	return false
}

// SetRoles sets the user's roles (used by repository when loading).
func (u *User) SetRoles(roles []*Role) {
	u.roles = roles
}

// GetPermissions returns all permissions from all assigned roles.
func (u *User) GetPermissions() *PermissionSet {
	ps := NewPermissionSet()
	for _, role := range u.roles {
		ps.Merge(role.Permissions())
	}
	return ps
}

// HasPermission checks if the user has a specific permission.
func (u *User) HasPermission(permission Permission) bool {
	return u.GetPermissions().HasPermission(permission)
}

// HasAnyPermission checks if the user has any of the given permissions.
func (u *User) HasAnyPermission(permissions ...Permission) bool {
	return u.GetPermissions().HasAnyPermission(permissions...)
}

// HasAllPermissions checks if the user has all of the given permissions.
func (u *User) HasAllPermissions(permissions ...Permission) bool {
	return u.GetPermissions().HasAllPermissions(permissions...)
}

// Metadata Management

// SetMetadata sets a metadata value.
func (u *User) SetMetadata(key string, value interface{}) {
	if u.metadata == nil {
		u.metadata = make(map[string]interface{})
	}
	u.metadata[key] = value
	u.MarkUpdated()
}

// GetMetadata gets a metadata value.
func (u *User) GetMetadata(key string) (interface{}, bool) {
	if u.metadata == nil {
		return nil, false
	}
	val, ok := u.metadata[key]
	return val, ok
}

// DeleteMetadata deletes a metadata value.
func (u *User) DeleteMetadata(key string) {
	if u.metadata == nil {
		return
	}
	delete(u.metadata, key)
	u.MarkUpdated()
}

// User errors
var (
	ErrUserNotFound        = fmt.Errorf("user not found")
	ErrUserTenantRequired  = fmt.Errorf("tenant ID is required")
	ErrUserEmailRequired   = fmt.Errorf("email is required")
	ErrUserPasswordRequired = fmt.Errorf("password is required")
	ErrUserAlreadyExists   = fmt.Errorf("user already exists")
	ErrUserNotActive       = fmt.Errorf("user is not active")
	ErrUserNotSuspended    = fmt.Errorf("user is not suspended")
	ErrUserDeleted         = fmt.Errorf("user is deleted")
	ErrUserInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrRoleNil             = fmt.Errorf("role cannot be nil")
	ErrRoleAlreadyAssigned = fmt.Errorf("role is already assigned to user")
	ErrRoleNotFound        = fmt.Errorf("role not found")
)
