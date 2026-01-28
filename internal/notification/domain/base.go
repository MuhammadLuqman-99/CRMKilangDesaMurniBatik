// Package domain contains the domain layer for the Notification service.
// It implements Domain-Driven Design (DDD) patterns including entities,
// value objects, aggregates, and domain events for notification management.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Entity is the base interface for all domain entities.
type Entity interface {
	GetID() uuid.UUID
}

// AggregateRoot is the base interface for aggregate roots.
type AggregateRoot interface {
	Entity
	GetDomainEvents() []DomainEvent
	ClearDomainEvents()
	AddDomainEvent(event DomainEvent)
	GetVersion() int
	IncrementVersion()
}

// BaseEntity provides common fields for all entities.
type BaseEntity struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at,omitempty"`
}

// GetID returns the entity ID.
func (e *BaseEntity) GetID() uuid.UUID {
	return e.ID
}

// IsDeleted returns true if the entity is soft deleted.
func (e *BaseEntity) IsDeleted() bool {
	return e.DeletedAt != nil
}

// MarkUpdated updates the UpdatedAt timestamp.
func (e *BaseEntity) MarkUpdated() {
	e.UpdatedAt = time.Now().UTC()
}

// MarkDeleted sets the DeletedAt timestamp.
func (e *BaseEntity) MarkDeleted() {
	now := time.Now().UTC()
	e.DeletedAt = &now
	e.UpdatedAt = now
}

// Restore clears the DeletedAt timestamp.
func (e *BaseEntity) Restore() {
	e.DeletedAt = nil
	e.UpdatedAt = time.Now().UTC()
}

// BaseAggregateRoot provides common fields for aggregate roots.
type BaseAggregateRoot struct {
	BaseEntity
	Version      int           `json:"version" db:"version"`
	domainEvents []DomainEvent `json:"-" db:"-"`
}

// GetDomainEvents returns all pending domain events.
func (a *BaseAggregateRoot) GetDomainEvents() []DomainEvent {
	return a.domainEvents
}

// ClearDomainEvents clears all pending domain events.
func (a *BaseAggregateRoot) ClearDomainEvents() {
	a.domainEvents = nil
}

// AddDomainEvent adds a domain event to the aggregate.
func (a *BaseAggregateRoot) AddDomainEvent(event DomainEvent) {
	a.domainEvents = append(a.domainEvents, event)
}

// GetVersion returns the aggregate version for optimistic locking.
func (a *BaseAggregateRoot) GetVersion() int {
	return a.Version
}

// IncrementVersion increments the aggregate version.
func (a *BaseAggregateRoot) IncrementVersion() {
	a.Version++
}

// NewBaseEntity creates a new base entity with generated ID and timestamps.
func NewBaseEntity() BaseEntity {
	now := time.Now().UTC()
	return BaseEntity{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewBaseEntityWithID creates a new base entity with a specific ID.
func NewBaseEntityWithID(id uuid.UUID) BaseEntity {
	now := time.Now().UTC()
	return BaseEntity{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewBaseAggregateRoot creates a new base aggregate root.
func NewBaseAggregateRoot() BaseAggregateRoot {
	return BaseAggregateRoot{
		BaseEntity:   NewBaseEntity(),
		Version:      1,
		domainEvents: make([]DomainEvent, 0),
	}
}

// DomainEvent is the base interface for all domain events.
type DomainEvent interface {
	EventID() uuid.UUID
	EventType() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
	AggregateType() string
	TenantID() uuid.UUID
	Version() int
}

// BaseDomainEvent provides common fields for domain events.
type BaseDomainEvent struct {
	ID            uuid.UUID `json:"id"`
	Type          string    `json:"type"`
	AggrID        uuid.UUID `json:"aggregate_id"`
	AggrType      string    `json:"aggregate_type"`
	TenantIDValue uuid.UUID `json:"tenant_id"`
	Occurred      time.Time `json:"occurred_at"`
	Ver           int       `json:"version"`
}

// EventID returns the event ID.
func (e *BaseDomainEvent) EventID() uuid.UUID {
	return e.ID
}

// EventType returns the event type.
func (e *BaseDomainEvent) EventType() string {
	return e.Type
}

// OccurredAt returns when the event occurred.
func (e *BaseDomainEvent) OccurredAt() time.Time {
	return e.Occurred
}

// AggregateID returns the aggregate ID.
func (e *BaseDomainEvent) AggregateID() uuid.UUID {
	return e.AggrID
}

// AggregateType returns the aggregate type.
func (e *BaseDomainEvent) AggregateType() string {
	return e.AggrType
}

// TenantID returns the tenant ID.
func (e *BaseDomainEvent) TenantID() uuid.UUID {
	return e.TenantIDValue
}

// Version returns the event version.
func (e *BaseDomainEvent) Version() int {
	return e.Ver
}

// NewBaseDomainEvent creates a new base domain event.
func NewBaseDomainEvent(eventType, aggregateType string, aggregateID, tenantID uuid.UUID, version int) BaseDomainEvent {
	return BaseDomainEvent{
		ID:            uuid.New(),
		Type:          eventType,
		AggrID:        aggregateID,
		AggrType:      aggregateType,
		TenantIDValue: tenantID,
		Occurred:      time.Now().UTC(),
		Ver:           version,
	}
}

// AuditInfo holds audit trail information.
type AuditInfo struct {
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`
	DeletedBy *uuid.UUID `json:"deleted_by,omitempty" db:"deleted_by"`
}

// SetCreatedBy sets the creator.
func (a *AuditInfo) SetCreatedBy(userID uuid.UUID) {
	a.CreatedBy = &userID
}

// SetUpdatedBy sets the last updater.
func (a *AuditInfo) SetUpdatedBy(userID uuid.UUID) {
	a.UpdatedBy = &userID
}

// SetDeletedBy sets who deleted the entity.
func (a *AuditInfo) SetDeletedBy(userID uuid.UUID) {
	a.DeletedBy = &userID
}
