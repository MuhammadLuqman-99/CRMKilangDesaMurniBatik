// Package events provides event bus abstractions for the CRM application.
// It supports publishing and subscribing to domain events using RabbitMQ.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event.
type EventType string

// Common event types
const (
	// IAM events
	EventTypeUserCreated         EventType = "iam.user.created"
	EventTypeUserUpdated         EventType = "iam.user.updated"
	EventTypeUserDeleted         EventType = "iam.user.deleted"
	EventTypeUserRoleAssigned    EventType = "iam.user.role_assigned"
	EventTypeUserRoleRevoked     EventType = "iam.user.role_revoked"
	EventTypeUserPasswordChanged EventType = "iam.user.password_changed"
	EventTypeTenantCreated       EventType = "iam.tenant.created"
	EventTypeTenantUpdated       EventType = "iam.tenant.updated"
	EventTypeTenantSuspended     EventType = "iam.tenant.suspended"

	// Customer events
	EventTypeCustomerCreated EventType = "customer.created"
	EventTypeCustomerUpdated EventType = "customer.updated"
	EventTypeCustomerDeleted EventType = "customer.deleted"
	EventTypeContactCreated  EventType = "customer.contact.created"
	EventTypeContactUpdated  EventType = "customer.contact.updated"
	EventTypeContactDeleted  EventType = "customer.contact.deleted"

	// Sales events
	EventTypeLeadCreated           EventType = "sales.lead.created"
	EventTypeLeadUpdated           EventType = "sales.lead.updated"
	EventTypeLeadQualified         EventType = "sales.lead.qualified"
	EventTypeLeadConverted         EventType = "sales.lead.converted"
	EventTypeLeadLost              EventType = "sales.lead.lost"
	EventTypeOpportunityCreated    EventType = "sales.opportunity.created"
	EventTypeOpportunityUpdated    EventType = "sales.opportunity.updated"
	EventTypeOpportunityStageMoved EventType = "sales.opportunity.stage_moved"
	EventTypeOpportunityWon        EventType = "sales.opportunity.won"
	EventTypeOpportunityLost       EventType = "sales.opportunity.lost"
	EventTypeDealCreated           EventType = "sales.deal.created"
	EventTypeDealUpdated           EventType = "sales.deal.updated"

	// Notification events
	EventTypeEmailSend EventType = "notification.email.send"
	EventTypeSMSSend   EventType = "notification.sms.send"
)

// Event represents a domain event.
type Event struct {
	ID          string                 `json:"id"`
	Type        EventType              `json:"type"`
	TenantID    string                 `json:"tenant_id"`
	AggregateID string                 `json:"aggregate_id"`
	Version     int                    `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// NewEvent creates a new event.
func NewEvent(eventType EventType, tenantID, aggregateID string, data map[string]interface{}) *Event {
	return &Event{
		ID:          uuid.New().String(),
		Type:        eventType,
		TenantID:    tenantID,
		AggregateID: aggregateID,
		Version:     1,
		Timestamp:   time.Now().UTC(),
		Data:        data,
		Metadata:    make(map[string]string),
	}
}

// WithMetadata adds metadata to the event.
func (e *Event) WithMetadata(key, value string) *Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// WithVersion sets the event version.
func (e *Event) WithVersion(version int) *Event {
	e.Version = version
	return e
}

// Marshal serializes the event to JSON.
func (e *Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// Unmarshal deserializes an event from JSON.
func Unmarshal(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &event, nil
}

// Publisher defines the interface for publishing events.
type Publisher interface {
	Publish(ctx context.Context, event *Event) error
	PublishBatch(ctx context.Context, events []*Event) error
	Close() error
}

// Subscriber defines the interface for subscribing to events.
type Subscriber interface {
	Subscribe(ctx context.Context, eventTypes []EventType, handler Handler) error
	Unsubscribe() error
	Close() error
}

// Handler is a function that handles an event.
type Handler func(ctx context.Context, event *Event) error

// EventBus combines Publisher and Subscriber interfaces.
type EventBus interface {
	Publisher
	Subscriber
}

// OutboxEntry represents an entry in the transactional outbox.
type OutboxEntry struct {
	ID        string    `json:"id" db:"id"`
	EventType EventType `json:"event_type" db:"event_type"`
	Payload   []byte    `json:"payload" db:"payload"`
	Published bool      `json:"published" db:"published"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewOutboxEntry creates a new outbox entry from an event.
func NewOutboxEntry(event *Event) (*OutboxEntry, error) {
	payload, err := event.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event for outbox: %w", err)
	}

	return &OutboxEntry{
		ID:        event.ID,
		EventType: event.Type,
		Payload:   payload,
		Published: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

// OutboxRepository defines the interface for outbox persistence.
type OutboxRepository interface {
	Save(ctx context.Context, entry *OutboxEntry) error
	GetUnpublished(ctx context.Context, limit int) ([]*OutboxEntry, error)
	MarkPublished(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

// Middleware defines event middleware for cross-cutting concerns.
type Middleware func(Handler) Handler

// WithRetry creates a middleware that retries failed event handling.
func WithRetry(maxRetries int, delay time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, event *Event) error {
			var lastErr error
			for i := 0; i <= maxRetries; i++ {
				if err := next(ctx, event); err != nil {
					lastErr = err
					if i < maxRetries {
						time.Sleep(delay * time.Duration(i+1))
					}
					continue
				}
				return nil
			}
			return fmt.Errorf("max retries exceeded: %w", lastErr)
		}
	}
}

// ChainMiddleware chains multiple middleware together.
func ChainMiddleware(handler Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
