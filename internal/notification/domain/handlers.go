// Package domain contains the domain layer for the Notification service.
package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventHandler defines the interface for handling domain events.
type EventHandler interface {
	// HandleEvent handles an external event and returns notifications to send.
	HandleEvent(ctx context.Context, event *ExternalEvent) ([]*Notification, error)

	// Supports returns true if this handler supports the given event type.
	Supports(eventType ExternalEventType) bool

	// Priority returns the handler priority (higher = runs first).
	Priority() int
}

// BaseEventHandler provides common functionality for event handlers.
type BaseEventHandler struct {
	templateRepo TemplateRepository
	prefRepo     PreferenceRepository
	triggerRepo  TriggerRepository
}

// NewBaseEventHandler creates a new base event handler.
func NewBaseEventHandler(
	templateRepo TemplateRepository,
	prefRepo PreferenceRepository,
	triggerRepo TriggerRepository,
) *BaseEventHandler {
	return &BaseEventHandler{
		templateRepo: templateRepo,
		prefRepo:     prefRepo,
		triggerRepo:  triggerRepo,
	}
}

// GetTemplate gets a template by code.
func (h *BaseEventHandler) GetTemplate(ctx context.Context, tenantID uuid.UUID, code string) (*NotificationTemplate, error) {
	return h.templateRepo.FindByCode(ctx, tenantID, code)
}

// CheckPreference checks if a user has opted in to receive notifications.
func (h *BaseEventHandler) CheckPreference(ctx context.Context, tenantID, userID uuid.UUID, channel NotificationChannel, notifType NotificationType) (bool, error) {
	optedOut, err := h.prefRepo.IsOptedOut(ctx, tenantID, userID, channel, notifType)
	if err != nil {
		return false, err
	}
	return !optedOut, nil
}

// GetTriggers gets active triggers for an event type.
func (h *BaseEventHandler) GetTriggers(ctx context.Context, tenantID uuid.UUID, eventType ExternalEventType) ([]*NotificationTrigger, error) {
	return h.triggerRepo.FindByEventType(ctx, tenantID, eventType)
}

// CreateNotificationFromTrigger creates a notification from a trigger and event.
func (h *BaseEventHandler) CreateNotificationFromTrigger(ctx context.Context, trigger *NotificationTrigger, event *ExternalEvent, recipient *Recipient) (*Notification, error) {
	// Get template
	template, err := h.templateRepo.FindByCode(ctx, event.TenantID, trigger.TemplateCode)
	if err != nil {
		return nil, fmt.Errorf("template not found: %s", trigger.TemplateCode)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("template is not active: %s", trigger.TemplateCode)
	}

	// Determine notification type from template
	notifType := template.Type

	// Create notification
	notification, err := NewNotification(event.TenantID, notifType, trigger.Channel, "")
	if err != nil {
		return nil, err
	}

	// Set template reference
	notification.SetTemplate(template.ID, template.Code)

	// Map event data to template variables
	data := trigger.MapData(event)

	// Render content based on channel
	switch trigger.Channel {
	case ChannelEmail:
		rendered, err := template.RenderEmail(data, recipient.Locale)
		if err != nil {
			return nil, err
		}
		notification.SetSubject(rendered.Subject)
		notification.Body = rendered.Body
		notification.HTMLBody = rendered.HTMLBody
		if err := notification.SetRecipientEmail(recipient.Email, recipient.Name); err != nil {
			return nil, err
		}
		if rendered.FromAddress != "" {
			notification.SetFromAddress(rendered.FromAddress, rendered.FromName)
		}
		notification.TrackOpens = rendered.TrackOpens
		notification.TrackClicks = rendered.TrackClicks

	case ChannelSMS:
		rendered, err := template.RenderSMS(data, recipient.Locale)
		if err != nil {
			return nil, err
		}
		notification.Body = rendered.Body
		if err := notification.SetRecipientPhone(recipient.Phone, recipient.Name); err != nil {
			return nil, err
		}

	case ChannelPush:
		rendered, err := template.RenderPush(data, recipient.Locale)
		if err != nil {
			return nil, err
		}
		notification.SetSubject(rendered.Title)
		notification.Body = rendered.Body
		if recipient.UserID != "" {
			userID, _ := uuid.Parse(recipient.UserID)
			notification.SetRecipientDevice(userID, recipient.DeviceToken, recipient.Name)
		}

	case ChannelInApp:
		rendered, err := template.RenderInApp(data, recipient.Locale)
		if err != nil {
			return nil, err
		}
		notification.SetSubject(rendered.Title)
		notification.Body = rendered.Body
		if recipient.UserID != "" {
			userID, _ := uuid.Parse(recipient.UserID)
			notification.SetRecipientUser(userID, recipient.Name)
		}
	}

	// Set source event reference
	notification.SetSourceEvent(string(event.EventType), &event.AggregateID, event.AggregateType)

	// Set data for additional context
	notification.SetData(data)

	// Increment template usage
	template.RecordUsage()

	return notification, nil
}

// ============================================================================
// IAM Service Event Handlers
// ============================================================================

// UserCreatedHandler handles user.created events (Welcome email).
type UserCreatedHandler struct {
	*BaseEventHandler
}

// NewUserCreatedHandler creates a new user created handler.
func NewUserCreatedHandler(base *BaseEventHandler) *UserCreatedHandler {
	return &UserCreatedHandler{BaseEventHandler: base}
}

// Supports returns true if this handler supports the event type.
func (h *UserCreatedHandler) Supports(eventType ExternalEventType) bool {
	return eventType == ExternalEventUserCreated
}

// Priority returns the handler priority.
func (h *UserCreatedHandler) Priority() int {
	return 100
}

// HandleEvent handles the user.created event.
func (h *UserCreatedHandler) HandleEvent(ctx context.Context, event *ExternalEvent) ([]*Notification, error) {
	var notifications []*Notification

	// Extract user data from event
	email := event.GetString("email")
	firstName := event.GetString("first_name")
	lastName := event.GetString("last_name")
	userID := event.GetUUID("user_id")

	if email == "" {
		return nil, fmt.Errorf("email is required for welcome notification")
	}

	// Get triggers for this event
	triggers, err := h.GetTriggers(ctx, event.TenantID, event.EventType)
	if err != nil {
		// If no triggers configured, use default welcome template
		triggers = []*NotificationTrigger{
			{
				EventType:    ExternalEventUserCreated,
				TemplateCode: "welcome_email",
				Channel:      ChannelEmail,
				IsActive:     true,
			},
		}
	}

	recipient := NewRecipient().
		WithUserID(userID.String()).
		WithEmail(email).
		WithName(firstName + " " + lastName)

	for _, trigger := range triggers {
		if !trigger.ShouldTrigger(event) {
			continue
		}

		// Check user preferences (if user ID available)
		if userID != uuid.Nil {
			canSend, err := h.CheckPreference(ctx, event.TenantID, userID, trigger.Channel, TypeWelcome)
			if err == nil && !canSend {
				continue
			}
		}

		notification, err := h.CreateNotificationFromTrigger(ctx, trigger, event, recipient)
		if err != nil {
			// Log error but continue with other triggers
			continue
		}

		// Set high priority for welcome emails
		notification.SetPriority(PriorityHigh)

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// ============================================================================
// Sales Service Event Handlers
// ============================================================================

// LeadCreatedHandler handles lead.created events (Sales team notification).
type LeadCreatedHandler struct {
	*BaseEventHandler
}

// NewLeadCreatedHandler creates a new lead created handler.
func NewLeadCreatedHandler(base *BaseEventHandler) *LeadCreatedHandler {
	return &LeadCreatedHandler{BaseEventHandler: base}
}

// Supports returns true if this handler supports the event type.
func (h *LeadCreatedHandler) Supports(eventType ExternalEventType) bool {
	return eventType == ExternalEventLeadCreated
}

// Priority returns the handler priority.
func (h *LeadCreatedHandler) Priority() int {
	return 90
}

// HandleEvent handles the lead.created event.
func (h *LeadCreatedHandler) HandleEvent(ctx context.Context, event *ExternalEvent) ([]*Notification, error) {
	var notifications []*Notification

	// Extract lead data
	leadCode := event.GetString("lead_code")
	contactName := event.GetString("contact_name")
	companyName := event.GetString("company_name")
	ownerID := event.GetUUID("owner_id")

	if ownerID == uuid.Nil {
		// No owner assigned, skip notification
		return notifications, nil
	}

	// Get triggers for this event
	triggers, err := h.GetTriggers(ctx, event.TenantID, event.EventType)
	if err != nil {
		triggers = []*NotificationTrigger{
			{
				EventType:    ExternalEventLeadCreated,
				TemplateCode: "lead_assigned",
				Channel:      ChannelEmail,
				IsActive:     true,
			},
			{
				EventType:    ExternalEventLeadCreated,
				TemplateCode: "lead_assigned",
				Channel:      ChannelInApp,
				IsActive:     true,
			},
		}
	}

	// Add custom data
	event.Payload["lead_code"] = leadCode
	event.Payload["contact_name"] = contactName
	event.Payload["company_name"] = companyName

	// Create recipient (owner)
	recipient := NewRecipient().
		WithUserID(ownerID.String())

	for _, trigger := range triggers {
		if !trigger.ShouldTrigger(event) {
			continue
		}

		canSend, err := h.CheckPreference(ctx, event.TenantID, ownerID, trigger.Channel, TypeAssignment)
		if err == nil && !canSend {
			continue
		}

		notification, err := h.CreateNotificationFromTrigger(ctx, trigger, event, recipient)
		if err != nil {
			continue
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// DealWonHandler handles opportunity.won events (Customer confirmation).
type DealWonHandler struct {
	*BaseEventHandler
}

// NewDealWonHandler creates a new deal won handler.
func NewDealWonHandler(base *BaseEventHandler) *DealWonHandler {
	return &DealWonHandler{BaseEventHandler: base}
}

// Supports returns true if this handler supports the event type.
func (h *DealWonHandler) Supports(eventType ExternalEventType) bool {
	return eventType == ExternalEventOpportunityWon
}

// Priority returns the handler priority.
func (h *DealWonHandler) Priority() int {
	return 95
}

// HandleEvent handles the opportunity.won event.
func (h *DealWonHandler) HandleEvent(ctx context.Context, event *ExternalEvent) ([]*Notification, error) {
	var notifications []*Notification

	// Extract data
	opportunityCode := event.GetString("opportunity_code")
	customerID := event.GetUUID("customer_id")
	customerEmail := event.GetString("customer_email")
	customerName := event.GetString("customer_name")
	amount := event.GetInt("amount")
	currency := event.GetString("currency")

	if customerEmail == "" {
		return notifications, nil
	}

	triggers, err := h.GetTriggers(ctx, event.TenantID, event.EventType)
	if err != nil {
		triggers = []*NotificationTrigger{
			{
				EventType:    ExternalEventOpportunityWon,
				TemplateCode: "deal_won_confirmation",
				Channel:      ChannelEmail,
				IsActive:     true,
			},
		}
	}

	event.Payload["opportunity_code"] = opportunityCode
	event.Payload["amount"] = amount
	event.Payload["currency"] = currency

	recipient := NewRecipient().
		WithUserID(customerID.String()).
		WithEmail(customerEmail).
		WithName(customerName)

	for _, trigger := range triggers {
		if !trigger.ShouldTrigger(event) {
			continue
		}

		notification, err := h.CreateNotificationFromTrigger(ctx, trigger, event, recipient)
		if err != nil {
			continue
		}

		// Deal confirmations are high priority
		notification.SetPriority(PriorityHigh)

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// DealLostHandler handles opportunity.lost events (Follow-up survey).
type DealLostHandler struct {
	*BaseEventHandler
}

// NewDealLostHandler creates a new deal lost handler.
func NewDealLostHandler(base *BaseEventHandler) *DealLostHandler {
	return &DealLostHandler{BaseEventHandler: base}
}

// Supports returns true if this handler supports the event type.
func (h *DealLostHandler) Supports(eventType ExternalEventType) bool {
	return eventType == ExternalEventOpportunityLost
}

// Priority returns the handler priority.
func (h *DealLostHandler) Priority() int {
	return 80
}

// HandleEvent handles the opportunity.lost event.
func (h *DealLostHandler) HandleEvent(ctx context.Context, event *ExternalEvent) ([]*Notification, error) {
	var notifications []*Notification

	customerEmail := event.GetString("customer_email")
	customerName := event.GetString("customer_name")
	lostReason := event.GetString("lost_reason")

	if customerEmail == "" {
		return notifications, nil
	}

	triggers, err := h.GetTriggers(ctx, event.TenantID, event.EventType)
	if err != nil {
		triggers = []*NotificationTrigger{
			{
				EventType:    ExternalEventOpportunityLost,
				TemplateCode: "deal_lost_survey",
				Channel:      ChannelEmail,
				IsActive:     true,
				Delay:        86400, // 24 hours delay
			},
		}
	}

	event.Payload["lost_reason"] = lostReason

	recipient := NewRecipient().
		WithEmail(customerEmail).
		WithName(customerName)

	for _, trigger := range triggers {
		if !trigger.ShouldTrigger(event) {
			continue
		}

		notification, err := h.CreateNotificationFromTrigger(ctx, trigger, event, recipient)
		if err != nil {
			continue
		}

		// Schedule if delay is configured
		if trigger.Delay > 0 {
			scheduledAt := time.Now().UTC().Add(time.Duration(trigger.Delay) * time.Second)
			notification.Schedule(scheduledAt)
		}

		// Low priority for survey emails
		notification.SetPriority(PriorityLow)

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// ============================================================================
// Event Handler Registry
// ============================================================================

// EventHandlerRegistry manages event handlers.
type EventHandlerRegistry struct {
	handlers map[ExternalEventType][]EventHandler
}

// NewEventHandlerRegistry creates a new event handler registry.
func NewEventHandlerRegistry() *EventHandlerRegistry {
	return &EventHandlerRegistry{
		handlers: make(map[ExternalEventType][]EventHandler),
	}
}

// Register registers an event handler.
func (r *EventHandlerRegistry) Register(handler EventHandler) {
	for _, eventType := range r.getSupportedEventTypes(handler) {
		r.handlers[eventType] = append(r.handlers[eventType], handler)
	}
}

// getSupportedEventTypes returns all event types supported by a handler.
func (r *EventHandlerRegistry) getSupportedEventTypes(handler EventHandler) []ExternalEventType {
	var supported []ExternalEventType
	allEventTypes := []ExternalEventType{
		ExternalEventUserCreated,
		ExternalEventUserActivated,
		ExternalEventUserDeactivated,
		ExternalEventPasswordChanged,
		ExternalEventEmailVerified,
		ExternalEventUserRoleAssigned,
		ExternalEventCustomerCreated,
		ExternalEventCustomerUpdated,
		ExternalEventCustomerConverted,
		ExternalEventCustomerChurned,
		ExternalEventLeadCreated,
		ExternalEventLeadConverted,
		ExternalEventLeadQualified,
		ExternalEventOpportunityCreated,
		ExternalEventOpportunityWon,
		ExternalEventOpportunityLost,
		ExternalEventDealCreated,
		ExternalEventDealFulfilled,
		ExternalEventDealCancelled,
		ExternalEventInvoiceCreated,
		ExternalEventPaymentReceived,
	}

	for _, eventType := range allEventTypes {
		if handler.Supports(eventType) {
			supported = append(supported, eventType)
		}
	}
	return supported
}

// GetHandlers returns handlers for an event type, sorted by priority.
func (r *EventHandlerRegistry) GetHandlers(eventType ExternalEventType) []EventHandler {
	handlers := r.handlers[eventType]

	// Sort by priority (higher first)
	for i := 0; i < len(handlers)-1; i++ {
		for j := i + 1; j < len(handlers); j++ {
			if handlers[j].Priority() > handlers[i].Priority() {
				handlers[i], handlers[j] = handlers[j], handlers[i]
			}
		}
	}

	return handlers
}

// HandleEvent dispatches an event to all registered handlers.
func (r *EventHandlerRegistry) HandleEvent(ctx context.Context, event *ExternalEvent) ([]*Notification, error) {
	handlers := r.GetHandlers(event.EventType)
	if len(handlers) == 0 {
		return nil, nil
	}

	var allNotifications []*Notification
	for _, handler := range handlers {
		notifications, err := handler.HandleEvent(ctx, event)
		if err != nil {
			// Log error but continue with other handlers
			continue
		}
		allNotifications = append(allNotifications, notifications...)
	}

	return allNotifications, nil
}

// ============================================================================
// Default Handler Registration
// ============================================================================

// RegisterDefaultHandlers registers all default event handlers.
func RegisterDefaultHandlers(
	registry *EventHandlerRegistry,
	templateRepo TemplateRepository,
	prefRepo PreferenceRepository,
	triggerRepo TriggerRepository,
) {
	base := NewBaseEventHandler(templateRepo, prefRepo, triggerRepo)

	// IAM Service handlers
	registry.Register(NewUserCreatedHandler(base))

	// Sales Service handlers
	registry.Register(NewLeadCreatedHandler(base))
	registry.Register(NewDealWonHandler(base))
	registry.Register(NewDealLostHandler(base))
}
