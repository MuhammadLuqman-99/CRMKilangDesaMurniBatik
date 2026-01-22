// Package domain contains the domain layer for the IAM service.
package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TenantStatus represents the status of a tenant.
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusPending   TenantStatus = "pending"
	TenantStatusTrial     TenantStatus = "trial"
)

// IsValid returns true if the status is valid.
func (s TenantStatus) IsValid() bool {
	switch s {
	case TenantStatusActive, TenantStatusInactive, TenantStatusSuspended, TenantStatusPending, TenantStatusTrial:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status.
func (s TenantStatus) String() string {
	return string(s)
}

// TenantPlan represents the subscription plan.
type TenantPlan string

const (
	TenantPlanFree       TenantPlan = "free"
	TenantPlanStarter    TenantPlan = "starter"
	TenantPlanPro        TenantPlan = "pro"
	TenantPlanEnterprise TenantPlan = "enterprise"
)

// IsValid returns true if the plan is valid.
func (p TenantPlan) IsValid() bool {
	switch p {
	case TenantPlanFree, TenantPlanStarter, TenantPlanPro, TenantPlanEnterprise:
		return true
	default:
		return false
	}
}

// String returns the string representation of the plan.
func (p TenantPlan) String() string {
	return string(p)
}

// MaxUsers returns the maximum number of users for the plan.
func (p TenantPlan) MaxUsers() int {
	switch p {
	case TenantPlanFree:
		return 3
	case TenantPlanStarter:
		return 10
	case TenantPlanPro:
		return 50
	case TenantPlanEnterprise:
		return -1 // Unlimited
	default:
		return 0
	}
}

// MaxContacts returns the maximum number of contacts for the plan.
func (p TenantPlan) MaxContacts() int {
	switch p {
	case TenantPlanFree:
		return 100
	case TenantPlanStarter:
		return 1000
	case TenantPlanPro:
		return 10000
	case TenantPlanEnterprise:
		return -1 // Unlimited
	default:
		return 0
	}
}

// Slug validation regex
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// TenantSettings represents tenant-specific settings.
type TenantSettings struct {
	Timezone           string                 `json:"timezone"`
	DateFormat         string                 `json:"date_format"`
	Currency           string                 `json:"currency"`
	Language           string                 `json:"language"`
	NotificationsEmail bool                   `json:"notifications_email"`
	Custom             map[string]interface{} `json:"custom"`
}

// DefaultTenantSettings returns default settings for a tenant.
func DefaultTenantSettings() TenantSettings {
	return TenantSettings{
		Timezone:           "UTC",
		DateFormat:         "YYYY-MM-DD",
		Currency:           "USD",
		Language:           "en",
		NotificationsEmail: true,
		Custom:             make(map[string]interface{}),
	}
}

// Tenant represents a tenant aggregate root in the IAM domain.
type Tenant struct {
	BaseAggregateRoot
	name     string
	slug     string
	status   TenantStatus
	plan     TenantPlan
	settings TenantSettings
	metadata map[string]interface{}
}

// NewTenant creates a new Tenant aggregate root.
func NewTenant(name, slug string) (*Tenant, error) {
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))

	if name == "" {
		return nil, ErrTenantNameRequired
	}

	if len(name) > 255 {
		return nil, ErrTenantNameTooLong
	}

	if slug == "" {
		return nil, ErrTenantSlugRequired
	}

	if len(slug) > 100 {
		return nil, ErrTenantSlugTooLong
	}

	if !slugRegex.MatchString(slug) {
		return nil, ErrTenantSlugInvalid
	}

	tenant := &Tenant{
		BaseAggregateRoot: NewBaseAggregateRoot(),
		name:              name,
		slug:              slug,
		status:            TenantStatusPending,
		plan:              TenantPlanFree,
		settings:          DefaultTenantSettings(),
		metadata:          make(map[string]interface{}),
	}

	tenant.AddDomainEvent(NewTenantCreatedEvent(tenant))

	return tenant, nil
}

// NewTenantWithPlan creates a new Tenant with a specific plan.
func NewTenantWithPlan(name, slug string, plan TenantPlan) (*Tenant, error) {
	tenant, err := NewTenant(name, slug)
	if err != nil {
		return nil, err
	}

	if !plan.IsValid() {
		return nil, ErrTenantPlanInvalid
	}

	tenant.plan = plan
	return tenant, nil
}

// ReconstructTenant reconstructs a Tenant from persistence.
func ReconstructTenant(
	id uuid.UUID,
	name, slug string,
	status TenantStatus,
	plan TenantPlan,
	settings TenantSettings,
	metadata map[string]interface{},
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Tenant {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &Tenant{
		BaseAggregateRoot: BaseAggregateRoot{
			BaseEntity: BaseEntity{
				ID:        id,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				DeletedAt: deletedAt,
			},
			domainEvents: make([]DomainEvent, 0),
		},
		name:     name,
		slug:     slug,
		status:   status,
		plan:     plan,
		settings: settings,
		metadata: metadata,
	}
}

// Getters

// Name returns the tenant name.
func (t *Tenant) Name() string {
	return t.name
}

// Slug returns the tenant slug.
func (t *Tenant) Slug() string {
	return t.slug
}

// Status returns the tenant status.
func (t *Tenant) Status() TenantStatus {
	return t.status
}

// Plan returns the tenant plan.
func (t *Tenant) Plan() TenantPlan {
	return t.plan
}

// Settings returns the tenant settings.
func (t *Tenant) Settings() TenantSettings {
	return t.settings
}

// Metadata returns the tenant metadata.
func (t *Tenant) Metadata() map[string]interface{} {
	return t.metadata
}

// IsActive returns true if the tenant is active.
func (t *Tenant) IsActive() bool {
	return t.status == TenantStatusActive || t.status == TenantStatusTrial
}

// IsSuspended returns true if the tenant is suspended.
func (t *Tenant) IsSuspended() bool {
	return t.status == TenantStatusSuspended
}

// Behaviors

// UpdateName updates the tenant name.
func (t *Tenant) UpdateName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrTenantNameRequired
	}

	if len(name) > 255 {
		return ErrTenantNameTooLong
	}

	t.name = name
	t.MarkUpdated()

	t.AddDomainEvent(NewTenantUpdatedEvent(t))

	return nil
}

// UpdateSettings updates the tenant settings.
func (t *Tenant) UpdateSettings(settings TenantSettings) {
	t.settings = settings
	t.MarkUpdated()

	t.AddDomainEvent(NewTenantSettingsUpdatedEvent(t))
}

// UpdateSettingsPartial updates specific settings fields.
func (t *Tenant) UpdateSettingsPartial(updates map[string]interface{}) {
	for key, value := range updates {
		switch key {
		case "timezone":
			if v, ok := value.(string); ok {
				t.settings.Timezone = v
			}
		case "date_format":
			if v, ok := value.(string); ok {
				t.settings.DateFormat = v
			}
		case "currency":
			if v, ok := value.(string); ok {
				t.settings.Currency = v
			}
		case "language":
			if v, ok := value.(string); ok {
				t.settings.Language = v
			}
		case "notifications_email":
			if v, ok := value.(bool); ok {
				t.settings.NotificationsEmail = v
			}
		default:
			// Store in custom settings
			if t.settings.Custom == nil {
				t.settings.Custom = make(map[string]interface{})
			}
			t.settings.Custom[key] = value
		}
	}
	t.MarkUpdated()
}

// Activate activates the tenant.
func (t *Tenant) Activate() error {
	if t.status == TenantStatusActive {
		return nil
	}

	if t.IsDeleted() {
		return ErrTenantDeleted
	}

	t.status = TenantStatusActive
	t.MarkUpdated()

	t.AddDomainEvent(NewTenantActivatedEvent(t))

	return nil
}

// Deactivate deactivates the tenant.
func (t *Tenant) Deactivate() error {
	if t.status == TenantStatusInactive {
		return nil
	}

	t.status = TenantStatusInactive
	t.MarkUpdated()

	t.AddDomainEvent(NewTenantDeactivatedEvent(t))

	return nil
}

// Suspend suspends the tenant.
func (t *Tenant) Suspend(reason string) error {
	if t.status == TenantStatusSuspended {
		return nil
	}

	t.status = TenantStatusSuspended
	t.metadata["suspend_reason"] = reason
	t.metadata["suspended_at"] = time.Now().UTC()
	t.MarkUpdated()

	t.AddDomainEvent(NewTenantSuspendedEvent(t, reason))

	return nil
}

// Unsuspend removes the suspension from the tenant.
func (t *Tenant) Unsuspend() error {
	if t.status != TenantStatusSuspended {
		return ErrTenantNotSuspended
	}

	t.status = TenantStatusActive
	delete(t.metadata, "suspend_reason")
	delete(t.metadata, "suspended_at")
	t.MarkUpdated()

	return nil
}

// StartTrial starts a trial period for the tenant.
func (t *Tenant) StartTrial(durationDays int) error {
	if t.status != TenantStatusPending {
		return ErrTenantAlreadyActivated
	}

	t.status = TenantStatusTrial
	t.metadata["trial_started_at"] = time.Now().UTC()
	t.metadata["trial_ends_at"] = time.Now().UTC().AddDate(0, 0, durationDays)
	t.MarkUpdated()

	t.AddDomainEvent(NewTenantTrialStartedEvent(t, durationDays))

	return nil
}

// IsTrialExpired returns true if the trial period has expired.
func (t *Tenant) IsTrialExpired() bool {
	if t.status != TenantStatusTrial {
		return false
	}

	if endDate, ok := t.metadata["trial_ends_at"].(time.Time); ok {
		return time.Now().UTC().After(endDate)
	}

	return false
}

// UpgradePlan upgrades the tenant's plan.
func (t *Tenant) UpgradePlan(plan TenantPlan) error {
	if !plan.IsValid() {
		return ErrTenantPlanInvalid
	}

	oldPlan := t.plan
	t.plan = plan
	t.MarkUpdated()

	t.AddDomainEvent(NewTenantPlanChangedEvent(t, oldPlan, plan))

	return nil
}

// DowngradePlan downgrades the tenant's plan.
func (t *Tenant) DowngradePlan(plan TenantPlan) error {
	if !plan.IsValid() {
		return ErrTenantPlanInvalid
	}

	oldPlan := t.plan
	t.plan = plan
	t.MarkUpdated()

	t.AddDomainEvent(NewTenantPlanChangedEvent(t, oldPlan, plan))

	return nil
}

// Delete soft deletes the tenant.
func (t *Tenant) Delete() {
	if t.IsDeleted() {
		return
	}

	t.MarkDeleted()

	t.AddDomainEvent(NewTenantDeletedEvent(t))
}

// CanAddUser checks if the tenant can add more users.
func (t *Tenant) CanAddUser(currentUserCount int) bool {
	maxUsers := t.plan.MaxUsers()
	if maxUsers == -1 {
		return true // Unlimited
	}
	return currentUserCount < maxUsers
}

// CanAddContact checks if the tenant can add more contacts.
func (t *Tenant) CanAddContact(currentContactCount int) bool {
	maxContacts := t.plan.MaxContacts()
	if maxContacts == -1 {
		return true // Unlimited
	}
	return currentContactCount < maxContacts
}

// Metadata Management

// SetMetadata sets a metadata value.
func (t *Tenant) SetMetadata(key string, value interface{}) {
	if t.metadata == nil {
		t.metadata = make(map[string]interface{})
	}
	t.metadata[key] = value
	t.MarkUpdated()
}

// GetMetadata gets a metadata value.
func (t *Tenant) GetMetadata(key string) (interface{}, bool) {
	if t.metadata == nil {
		return nil, false
	}
	val, ok := t.metadata[key]
	return val, ok
}

// DeleteMetadata deletes a metadata value.
func (t *Tenant) DeleteMetadata(key string) {
	if t.metadata == nil {
		return
	}
	delete(t.metadata, key)
	t.MarkUpdated()
}

// Tenant errors
var (
	ErrTenantNotFound         = fmt.Errorf("tenant not found")
	ErrTenantNameRequired     = fmt.Errorf("tenant name is required")
	ErrTenantNameTooLong      = fmt.Errorf("tenant name exceeds maximum length of 255 characters")
	ErrTenantSlugRequired     = fmt.Errorf("tenant slug is required")
	ErrTenantSlugTooLong      = fmt.Errorf("tenant slug exceeds maximum length of 100 characters")
	ErrTenantSlugInvalid      = fmt.Errorf("tenant slug must contain only lowercase letters, numbers, and hyphens")
	ErrTenantSlugExists       = fmt.Errorf("tenant slug already exists")
	ErrTenantPlanInvalid      = fmt.Errorf("invalid tenant plan")
	ErrTenantDeleted          = fmt.Errorf("tenant is deleted")
	ErrTenantNotSuspended     = fmt.Errorf("tenant is not suspended")
	ErrTenantAlreadyActivated = fmt.Errorf("tenant is already activated")
	ErrTenantNotActive        = fmt.Errorf("tenant is not active")
	ErrTenantLimitReached     = fmt.Errorf("tenant limit reached for current plan")
)
