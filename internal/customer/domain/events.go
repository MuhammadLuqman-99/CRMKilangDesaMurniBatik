// Package domain contains the domain layer for the Customer service.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Event types for Customer domain.
const (
	// Customer events
	EventTypeCustomerCreated       = "customer.created"
	EventTypeCustomerUpdated       = "customer.updated"
	EventTypeCustomerDeleted       = "customer.deleted"
	EventTypeCustomerStatusChanged = "customer.status_changed"
	EventTypeCustomerConverted     = "customer.converted"
	EventTypeCustomerChurned       = "customer.churned"
	EventTypeCustomerTierChanged   = "customer.tier_changed"
	EventTypeCustomerOwnerAssigned = "customer.owner_assigned"
	EventTypeCustomerTagAdded      = "customer.tag_added"
	EventTypeCustomerTagRemoved    = "customer.tag_removed"
	EventTypeCustomerMerged        = "customer.merged"
	EventTypeCustomerImported      = "customer.imported"

	// Contact events
	EventTypeContactAdded   = "customer.contact.added"
	EventTypeContactUpdated = "customer.contact.updated"
	EventTypeContactRemoved = "customer.contact.removed"

	// Note events
	EventTypeNoteAdded   = "customer.note.added"
	EventTypeNoteUpdated = "customer.note.updated"
	EventTypeNoteRemoved = "customer.note.removed"

	// Activity events
	EventTypeActivityLogged = "customer.activity.logged"

	// Segment events
	EventTypeCustomerAddedToSegment     = "customer.segment.added"
	EventTypeCustomerRemovedFromSegment = "customer.segment.removed"
)

// AggregateType for Customer domain.
const AggregateTypeCustomer = "customer"

// ============================================================================
// Customer Events
// ============================================================================

// CustomerCreatedEvent is raised when a customer is created.
type CustomerCreatedEvent struct {
	BaseDomainEvent
	CustomerID   uuid.UUID      `json:"customer_id"`
	Code         string         `json:"code"`
	Name         string         `json:"name"`
	Type         CustomerType   `json:"type"`
	Status       CustomerStatus `json:"status"`
	Source       CustomerSource `json:"source"`
	Email        string         `json:"email,omitempty"`
	OwnerID      *uuid.UUID     `json:"owner_id,omitempty"`
	CreatedBy    *uuid.UUID     `json:"created_by,omitempty"`
}

// NewCustomerCreatedEvent creates a new CustomerCreatedEvent.
func NewCustomerCreatedEvent(customer *Customer) *CustomerCreatedEvent {
	return &CustomerCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerCreated,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID: customer.ID,
		Code:       customer.Code,
		Name:       customer.Name,
		Type:       customer.Type,
		Status:     customer.Status,
		Source:     customer.Source,
		Email:      customer.Email.String(),
		OwnerID:    customer.OwnerID,
		CreatedBy:  customer.AuditInfo.CreatedBy,
	}
}

// CustomerUpdatedEvent is raised when a customer is updated.
type CustomerUpdatedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID              `json:"customer_id"`
	Changes    map[string]interface{} `json:"changes"`
	UpdatedBy  *uuid.UUID             `json:"updated_by,omitempty"`
}

// NewCustomerUpdatedEvent creates a new CustomerUpdatedEvent.
func NewCustomerUpdatedEvent(customer *Customer, changes map[string]interface{}) *CustomerUpdatedEvent {
	return &CustomerUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerUpdated,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID: customer.ID,
		Changes:    changes,
		UpdatedBy:  customer.AuditInfo.UpdatedBy,
	}
}

// CustomerDeletedEvent is raised when a customer is deleted.
type CustomerDeletedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID  `json:"customer_id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	DeletedBy  *uuid.UUID `json:"deleted_by,omitempty"`
}

// NewCustomerDeletedEvent creates a new CustomerDeletedEvent.
func NewCustomerDeletedEvent(customer *Customer) *CustomerDeletedEvent {
	return &CustomerDeletedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerDeleted,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID: customer.ID,
		Code:       customer.Code,
		Name:       customer.Name,
		DeletedBy:  customer.AuditInfo.DeletedBy,
	}
}

// CustomerStatusChangedEvent is raised when customer status changes.
type CustomerStatusChangedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID      `json:"customer_id"`
	OldStatus  CustomerStatus `json:"old_status"`
	NewStatus  CustomerStatus `json:"new_status"`
}

// NewCustomerStatusChangedEvent creates a new CustomerStatusChangedEvent.
func NewCustomerStatusChangedEvent(customer *Customer, oldStatus, newStatus CustomerStatus) *CustomerStatusChangedEvent {
	return &CustomerStatusChangedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerStatusChanged,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID: customer.ID,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
	}
}

// CustomerConvertedEvent is raised when a lead/prospect is converted.
type CustomerConvertedEvent struct {
	BaseDomainEvent
	CustomerID   uuid.UUID      `json:"customer_id"`
	FromStatus   CustomerStatus `json:"from_status"`
	ConvertedAt  time.Time      `json:"converted_at"`
	ConvertedBy  *uuid.UUID     `json:"converted_by,omitempty"`
}

// NewCustomerConvertedEvent creates a new CustomerConvertedEvent.
func NewCustomerConvertedEvent(customer *Customer, fromStatus CustomerStatus) *CustomerConvertedEvent {
	return &CustomerConvertedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerConverted,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID:  customer.ID,
		FromStatus:  fromStatus,
		ConvertedAt: time.Now().UTC(),
		ConvertedBy: customer.AuditInfo.UpdatedBy,
	}
}

// CustomerChurnedEvent is raised when a customer churns.
type CustomerChurnedEvent struct {
	BaseDomainEvent
	CustomerID  uuid.UUID  `json:"customer_id"`
	ChurnReason string     `json:"churn_reason,omitempty"`
	ChurnedAt   time.Time  `json:"churned_at"`
}

// NewCustomerChurnedEvent creates a new CustomerChurnedEvent.
func NewCustomerChurnedEvent(customer *Customer, reason string) *CustomerChurnedEvent {
	return &CustomerChurnedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerChurned,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID:  customer.ID,
		ChurnReason: reason,
		ChurnedAt:   time.Now().UTC(),
	}
}

// CustomerTierChangedEvent is raised when customer tier changes.
type CustomerTierChangedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID    `json:"customer_id"`
	OldTier    CustomerTier `json:"old_tier"`
	NewTier    CustomerTier `json:"new_tier"`
}

// NewCustomerTierChangedEvent creates a new CustomerTierChangedEvent.
func NewCustomerTierChangedEvent(customer *Customer, oldTier, newTier CustomerTier) *CustomerTierChangedEvent {
	return &CustomerTierChangedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerTierChanged,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID: customer.ID,
		OldTier:    oldTier,
		NewTier:    newTier,
	}
}

// CustomerOwnerAssignedEvent is raised when a customer owner is assigned.
type CustomerOwnerAssignedEvent struct {
	BaseDomainEvent
	CustomerID   uuid.UUID  `json:"customer_id"`
	OldOwnerID   *uuid.UUID `json:"old_owner_id,omitempty"`
	NewOwnerID   uuid.UUID  `json:"new_owner_id"`
}

// NewCustomerOwnerAssignedEvent creates a new CustomerOwnerAssignedEvent.
func NewCustomerOwnerAssignedEvent(customer *Customer, oldOwner *uuid.UUID, newOwner uuid.UUID) *CustomerOwnerAssignedEvent {
	return &CustomerOwnerAssignedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerOwnerAssigned,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID: customer.ID,
		OldOwnerID: oldOwner,
		NewOwnerID: newOwner,
	}
}

// CustomerTagAddedEvent is raised when a tag is added.
type CustomerTagAddedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID `json:"customer_id"`
	Tag        string    `json:"tag"`
}

// NewCustomerTagAddedEvent creates a new CustomerTagAddedEvent.
func NewCustomerTagAddedEvent(customer *Customer, tag string) *CustomerTagAddedEvent {
	return &CustomerTagAddedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerTagAdded,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID: customer.ID,
		Tag:        tag,
	}
}

// CustomerTagRemovedEvent is raised when a tag is removed.
type CustomerTagRemovedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID `json:"customer_id"`
	Tag        string    `json:"tag"`
}

// NewCustomerTagRemovedEvent creates a new CustomerTagRemovedEvent.
func NewCustomerTagRemovedEvent(customer *Customer, tag string) *CustomerTagRemovedEvent {
	return &CustomerTagRemovedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerTagRemoved,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID: customer.ID,
		Tag:        tag,
	}
}

// CustomerMergedEvent is raised when customers are merged.
type CustomerMergedEvent struct {
	BaseDomainEvent
	TargetCustomerID uuid.UUID   `json:"target_customer_id"`
	SourceCustomerIDs []uuid.UUID `json:"source_customer_ids"`
	MergedBy         *uuid.UUID  `json:"merged_by,omitempty"`
}

// NewCustomerMergedEvent creates a new CustomerMergedEvent.
func NewCustomerMergedEvent(targetID, tenantID uuid.UUID, sourceIDs []uuid.UUID, mergedBy *uuid.UUID) *CustomerMergedEvent {
	return &CustomerMergedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerMerged,
			targetID,
			AggregateTypeCustomer,
			tenantID,
			1,
		),
		TargetCustomerID:  targetID,
		SourceCustomerIDs: sourceIDs,
		MergedBy:          mergedBy,
	}
}

// CustomerImportedEvent is raised when a customer is imported.
type CustomerImportedEvent struct {
	BaseDomainEvent
	CustomerID  uuid.UUID `json:"customer_id"`
	ImportBatch string    `json:"import_batch"`
	Source      string    `json:"source"`
}

// NewCustomerImportedEvent creates a new CustomerImportedEvent.
func NewCustomerImportedEvent(customer *Customer, batch, source string) *CustomerImportedEvent {
	return &CustomerImportedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerImported,
			customer.ID,
			AggregateTypeCustomer,
			customer.TenantID,
			customer.Version,
		),
		CustomerID:  customer.ID,
		ImportBatch: batch,
		Source:      source,
	}
}

// ============================================================================
// Contact Events
// ============================================================================

// ContactAddedEvent is raised when a contact is added to a customer.
type ContactAddedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID `json:"customer_id"`
	ContactID  uuid.UUID `json:"contact_id"`
}

// NewContactAddedEvent creates a new ContactAddedEvent.
func NewContactAddedEvent(customerID, tenantID, contactID uuid.UUID) *ContactAddedEvent {
	return &ContactAddedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeContactAdded,
			customerID,
			AggregateTypeCustomer,
			tenantID,
			1,
		),
		CustomerID: customerID,
		ContactID:  contactID,
	}
}

// ContactUpdatedEvent is raised when a contact is updated.
type ContactUpdatedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID              `json:"customer_id"`
	ContactID  uuid.UUID              `json:"contact_id"`
	Changes    map[string]interface{} `json:"changes"`
}

// NewContactUpdatedEvent creates a new ContactUpdatedEvent.
func NewContactUpdatedEvent(customerID, tenantID, contactID uuid.UUID, changes map[string]interface{}) *ContactUpdatedEvent {
	return &ContactUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeContactUpdated,
			customerID,
			AggregateTypeCustomer,
			tenantID,
			1,
		),
		CustomerID: customerID,
		ContactID:  contactID,
		Changes:    changes,
	}
}

// ContactRemovedEvent is raised when a contact is removed.
type ContactRemovedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID `json:"customer_id"`
	ContactID  uuid.UUID `json:"contact_id"`
}

// NewContactRemovedEvent creates a new ContactRemovedEvent.
func NewContactRemovedEvent(customerID, tenantID, contactID uuid.UUID) *ContactRemovedEvent {
	return &ContactRemovedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeContactRemoved,
			customerID,
			AggregateTypeCustomer,
			tenantID,
			1,
		),
		CustomerID: customerID,
		ContactID:  contactID,
	}
}

// ============================================================================
// Note Events
// ============================================================================

// NoteAddedEvent is raised when a note is added.
type NoteAddedEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID  `json:"customer_id"`
	NoteID     uuid.UUID  `json:"note_id"`
	Content    string     `json:"content"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty"`
}

// NewNoteAddedEvent creates a new NoteAddedEvent.
func NewNoteAddedEvent(customerID, tenantID, noteID uuid.UUID, content string, createdBy *uuid.UUID) *NoteAddedEvent {
	return &NoteAddedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeNoteAdded,
			customerID,
			AggregateTypeCustomer,
			tenantID,
			1,
		),
		CustomerID: customerID,
		NoteID:     noteID,
		Content:    content,
		CreatedBy:  createdBy,
	}
}

// ============================================================================
// Activity Events
// ============================================================================

// ActivityType represents types of customer activities.
type ActivityType string

const (
	ActivityTypeCall     ActivityType = "call"
	ActivityTypeEmail    ActivityType = "email"
	ActivityTypeMeeting  ActivityType = "meeting"
	ActivityTypeNote     ActivityType = "note"
	ActivityTypeTask     ActivityType = "task"
	ActivityTypeDeal     ActivityType = "deal"
	ActivityTypeVisit    ActivityType = "visit"
	ActivityTypeDemo     ActivityType = "demo"
	ActivityTypeProposal ActivityType = "proposal"
	ActivityTypeOther    ActivityType = "other"
)

// ActivityLoggedEvent is raised when an activity is logged.
type ActivityLoggedEvent struct {
	BaseDomainEvent
	CustomerID   uuid.UUID    `json:"customer_id"`
	ActivityID   uuid.UUID    `json:"activity_id"`
	ActivityType ActivityType `json:"activity_type"`
	Subject      string       `json:"subject"`
	Description  string       `json:"description,omitempty"`
	OccurredAt   time.Time    `json:"occurred_at"`
	PerformedBy  *uuid.UUID   `json:"performed_by,omitempty"`
}

// NewActivityLoggedEvent creates a new ActivityLoggedEvent.
func NewActivityLoggedEvent(
	customerID, tenantID, activityID uuid.UUID,
	activityType ActivityType,
	subject, description string,
	occurredAt time.Time,
	performedBy *uuid.UUID,
) *ActivityLoggedEvent {
	return &ActivityLoggedEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeActivityLogged,
			customerID,
			AggregateTypeCustomer,
			tenantID,
			1,
		),
		CustomerID:   customerID,
		ActivityID:   activityID,
		ActivityType: activityType,
		Subject:      subject,
		Description:  description,
		OccurredAt:   occurredAt,
		PerformedBy:  performedBy,
	}
}

// ============================================================================
// Segment Events
// ============================================================================

// CustomerAddedToSegmentEvent is raised when customer is added to a segment.
type CustomerAddedToSegmentEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID `json:"customer_id"`
	SegmentID  uuid.UUID `json:"segment_id"`
}

// NewCustomerAddedToSegmentEvent creates a new event.
func NewCustomerAddedToSegmentEvent(customerID, tenantID, segmentID uuid.UUID) *CustomerAddedToSegmentEvent {
	return &CustomerAddedToSegmentEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerAddedToSegment,
			customerID,
			AggregateTypeCustomer,
			tenantID,
			1,
		),
		CustomerID: customerID,
		SegmentID:  segmentID,
	}
}

// CustomerRemovedFromSegmentEvent is raised when customer is removed from a segment.
type CustomerRemovedFromSegmentEvent struct {
	BaseDomainEvent
	CustomerID uuid.UUID `json:"customer_id"`
	SegmentID  uuid.UUID `json:"segment_id"`
}

// NewCustomerRemovedFromSegmentEvent creates a new event.
func NewCustomerRemovedFromSegmentEvent(customerID, tenantID, segmentID uuid.UUID) *CustomerRemovedFromSegmentEvent {
	return &CustomerRemovedFromSegmentEvent{
		BaseDomainEvent: NewBaseDomainEvent(
			EventTypeCustomerRemovedFromSegment,
			customerID,
			AggregateTypeCustomer,
			tenantID,
			1,
		),
		CustomerID: customerID,
		SegmentID:  segmentID,
	}
}

// ============================================================================
// Event Helpers
// ============================================================================

// EventMetadata holds common event metadata.
type EventMetadata struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	CausationID   string `json:"causation_id,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	IPAddress     string `json:"ip_address,omitempty"`
	UserAgent     string `json:"user_agent,omitempty"`
}

// EventEnvelope wraps an event with metadata for publishing.
type EventEnvelope struct {
	ID          uuid.UUID      `json:"id"`
	Type        string         `json:"type"`
	AggregateID uuid.UUID      `json:"aggregate_id"`
	TenantID    uuid.UUID      `json:"tenant_id"`
	Version     int            `json:"version"`
	OccurredAt  time.Time      `json:"occurred_at"`
	Payload     DomainEvent    `json:"payload"`
	Metadata    *EventMetadata `json:"metadata,omitempty"`
}

// NewEventEnvelope creates a new event envelope.
func NewEventEnvelope(event DomainEvent, metadata *EventMetadata) EventEnvelope {
	return EventEnvelope{
		ID:          uuid.New(),
		Type:        event.EventType(),
		AggregateID: event.AggregateID(),
		TenantID:    event.TenantID(),
		Version:     event.Version(),
		OccurredAt:  event.OccurredAt(),
		Payload:     event,
		Metadata:    metadata,
	}
}
