// Package domain contains the domain layer for the Customer service.
// It implements Domain-Driven Design (DDD) patterns including entities,
// value objects, aggregates, and domain events for customer relationship management.
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
	ID        uuid.UUID  `json:"id" bson:"_id"`
	CreatedAt time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
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
	Version      int           `json:"version" bson:"version"`
	domainEvents []DomainEvent `json:"-" bson:"-"`
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
	EventType() string
	OccurredAt() time.Time
	AggregateID() uuid.UUID
	AggregateType() string
	TenantID() uuid.UUID
	Version() int
}

// BaseDomainEvent provides common fields for domain events.
type BaseDomainEvent struct {
	eventType     string
	occurredAt    time.Time
	aggregateID   uuid.UUID
	aggregateType string
	tenantID      uuid.UUID
	version       int
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

// TenantID returns the tenant ID.
func (e *BaseDomainEvent) TenantID() uuid.UUID {
	return e.tenantID
}

// Version returns the event version.
func (e *BaseDomainEvent) Version() int {
	return e.version
}

// NewBaseDomainEvent creates a new base domain event.
func NewBaseDomainEvent(eventType string, aggregateID uuid.UUID, aggregateType string, tenantID uuid.UUID, version int) BaseDomainEvent {
	return BaseDomainEvent{
		eventType:     eventType,
		occurredAt:    time.Now().UTC(),
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		tenantID:      tenantID,
		version:       version,
	}
}

// Metadata holds common metadata for entities.
type Metadata struct {
	Source      string            `json:"source,omitempty" bson:"source,omitempty"`
	Tags        []string          `json:"tags,omitempty" bson:"tags,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" bson:"labels,omitempty"`
	ImportedAt  *time.Time        `json:"imported_at,omitempty" bson:"imported_at,omitempty"`
	ImportBatch string            `json:"import_batch,omitempty" bson:"import_batch,omitempty"`
}

// HasTag checks if the metadata contains a specific tag.
func (m *Metadata) HasTag(tag string) bool {
	for _, t := range m.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag if not already present.
func (m *Metadata) AddTag(tag string) {
	if !m.HasTag(tag) {
		m.Tags = append(m.Tags, tag)
	}
}

// RemoveTag removes a tag.
func (m *Metadata) RemoveTag(tag string) {
	for i, t := range m.Tags {
		if t == tag {
			m.Tags = append(m.Tags[:i], m.Tags[i+1:]...)
			return
		}
	}
}

// GetLabel returns a label value.
func (m *Metadata) GetLabel(key string) string {
	if m.Labels == nil {
		return ""
	}
	return m.Labels[key]
}

// SetLabel sets a label.
func (m *Metadata) SetLabel(key, value string) {
	if m.Labels == nil {
		m.Labels = make(map[string]string)
	}
	m.Labels[key] = value
}

// AuditInfo holds audit trail information.
type AuditInfo struct {
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" bson:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
	DeletedBy     *uuid.UUID `json:"deleted_by,omitempty" bson:"deleted_by,omitempty"`
	LastViewedAt  *time.Time `json:"last_viewed_at,omitempty" bson:"last_viewed_at,omitempty"`
	LastViewedBy  *uuid.UUID `json:"last_viewed_by,omitempty" bson:"last_viewed_by,omitempty"`
	LastContactAt *time.Time `json:"last_contact_at,omitempty" bson:"last_contact_at,omitempty"`
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

// RecordView records a view.
func (a *AuditInfo) RecordView(userID uuid.UUID) {
	now := time.Now().UTC()
	a.LastViewedAt = &now
	a.LastViewedBy = &userID
}

// RecordContact records a contact interaction.
func (a *AuditInfo) RecordContact() {
	now := time.Now().UTC()
	a.LastContactAt = &now
}
