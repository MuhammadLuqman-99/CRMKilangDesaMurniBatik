// Package domain contains the domain layer for the IAM service.
// It implements Domain-Driven Design (DDD) patterns including entities,
// value objects, aggregates, and domain events.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Entity is the base interface for all domain entities.
// Entities have a unique identity that persists through time.
type Entity interface {
	GetID() uuid.UUID
}

// AggregateRoot is the base interface for aggregate roots.
// Aggregate roots are the entry points to aggregates and ensure consistency.
type AggregateRoot interface {
	Entity
	GetDomainEvents() []DomainEvent
	ClearDomainEvents()
	AddDomainEvent(event DomainEvent)
}

// BaseEntity provides common fields for all entities.
type BaseEntity struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
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

// BaseAggregateRoot provides common fields for aggregate roots.
type BaseAggregateRoot struct {
	BaseEntity
	domainEvents []DomainEvent
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

// NewBaseEntity creates a new base entity with generated ID and timestamps.
func NewBaseEntity() BaseEntity {
	now := time.Now().UTC()
	return BaseEntity{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewBaseAggregateRoot creates a new base aggregate root.
func NewBaseAggregateRoot() BaseAggregateRoot {
	return BaseAggregateRoot{
		BaseEntity:   NewBaseEntity(),
		domainEvents: make([]DomainEvent, 0),
	}
}

// DomainEvent is the base interface for all domain events.
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
	AggregateType() string
}

// BaseDomainEvent provides common fields for domain events.
type BaseDomainEvent struct {
	eventType     string
	occurredAt    time.Time
	aggregateID   uuid.UUID
	aggregateType string
}

// EventType returns the event type.
func (e *BaseDomainEvent) EventType() string {
	return e.eventType
}

// OccurredAt returns when the event occurred.
func (e *BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// AggregateID returns the aggregate ID.
func (e *BaseDomainEvent) AggregateID() uuid.UUID {
	return e.aggregateID
}

// AggregateType returns the aggregate type.
func (e *BaseDomainEvent) AggregateType() string {
	return e.aggregateType
}

// NewBaseDomainEvent creates a new base domain event.
func NewBaseDomainEvent(eventType string, aggregateID uuid.UUID, aggregateType string) BaseDomainEvent {
	return BaseDomainEvent{
		eventType:     eventType,
		occurredAt:    time.Now().UTC(),
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
	}
}
