// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event.
type DomainEvent interface {
	EventID() uuid.UUID
	EventType() string
	AggregateID() uuid.UUID
	AggregateType() string
	TenantID() uuid.UUID
	OccurredAt() time.Time
	Version() int
}

// BaseEvent provides common event functionality.
type BaseEvent struct {
	ID            uuid.UUID `json:"id"`
	Type          string    `json:"type"`
	AggrID        uuid.UUID `json:"aggregate_id"`
	AggrType      string    `json:"aggregate_type"`
	TenantIDValue uuid.UUID `json:"tenant_id"`
	Occurred      time.Time `json:"occurred_at"`
	Ver           int       `json:"version"`
}

// EventID returns the event ID.
func (e BaseEvent) EventID() uuid.UUID { return e.ID }

// EventType returns the event type.
func (e BaseEvent) EventType() string { return e.Type }

// AggregateID returns the aggregate ID.
func (e BaseEvent) AggregateID() uuid.UUID { return e.AggrID }

// AggregateType returns the aggregate type.
func (e BaseEvent) AggregateType() string { return e.AggrType }

// TenantID returns the tenant ID.
func (e BaseEvent) TenantID() uuid.UUID { return e.TenantIDValue }

// OccurredAt returns when the event occurred.
func (e BaseEvent) OccurredAt() time.Time { return e.Occurred }

// Version returns the event version.
func (e BaseEvent) Version() int { return e.Ver }

// newBaseEvent creates a new base event.
func newBaseEvent(eventType, aggregateType string, aggregateID, tenantID uuid.UUID, version int) BaseEvent {
	return BaseEvent{
		ID:            uuid.New(),
		Type:          eventType,
		AggrID:        aggregateID,
		AggrType:      aggregateType,
		TenantIDValue: tenantID,
		Occurred:      time.Now().UTC(),
		Ver:           version,
	}
}

// ============================================================================
// Lead Events
// ============================================================================

// LeadCreatedEvent is raised when a lead is created.
type LeadCreatedEvent struct {
	BaseEvent
	LeadCode    string     `json:"lead_code"`
	ContactName string     `json:"contact_name"`
	CompanyName string     `json:"company_name"`
	Source      LeadSource `json:"source"`
	OwnerID     *uuid.UUID `json:"owner_id,omitempty"`
}

// NewLeadCreatedEvent creates a new lead created event.
func NewLeadCreatedEvent(lead *Lead) *LeadCreatedEvent {
	return &LeadCreatedEvent{
		BaseEvent:   newBaseEvent("lead.created", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:    lead.Code,
		ContactName: lead.Contact.FullName(),
		CompanyName: lead.Company.Name,
		Source:      lead.Source,
		OwnerID:     lead.OwnerID,
	}
}

// LeadUpdatedEvent is raised when a lead is updated.
type LeadUpdatedEvent struct {
	BaseEvent
	LeadCode string `json:"lead_code"`
}

// NewLeadUpdatedEvent creates a new lead updated event.
func NewLeadUpdatedEvent(lead *Lead) *LeadUpdatedEvent {
	return &LeadUpdatedEvent{
		BaseEvent: newBaseEvent("lead.updated", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:  lead.Code,
	}
}

// LeadContactedEvent is raised when a lead is contacted.
type LeadContactedEvent struct {
	BaseEvent
	LeadCode    string    `json:"lead_code"`
	ContactedAt time.Time `json:"contacted_at"`
}

// NewLeadContactedEvent creates a new lead contacted event.
func NewLeadContactedEvent(lead *Lead) *LeadContactedEvent {
	return &LeadContactedEvent{
		BaseEvent:   newBaseEvent("lead.contacted", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:    lead.Code,
		ContactedAt: *lead.LastContactedAt,
	}
}

// LeadQualifiedEvent is raised when a lead is qualified.
type LeadQualifiedEvent struct {
	BaseEvent
	LeadCode string `json:"lead_code"`
	Score    int    `json:"score"`
}

// NewLeadQualifiedEvent creates a new lead qualified event.
func NewLeadQualifiedEvent(lead *Lead) *LeadQualifiedEvent {
	return &LeadQualifiedEvent{
		BaseEvent: newBaseEvent("lead.qualified", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:  lead.Code,
		Score:     lead.Score.Score,
	}
}

// LeadDisqualifiedEvent is raised when a lead is disqualified.
type LeadDisqualifiedEvent struct {
	BaseEvent
	LeadCode string `json:"lead_code"`
	Reason   string `json:"reason"`
}

// NewLeadDisqualifiedEvent creates a new lead disqualified event.
func NewLeadDisqualifiedEvent(lead *Lead, reason string) *LeadDisqualifiedEvent {
	return &LeadDisqualifiedEvent{
		BaseEvent: newBaseEvent("lead.disqualified", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:  lead.Code,
		Reason:    reason,
	}
}

// LeadConvertedEvent is raised when a lead is converted to an opportunity.
type LeadConvertedEvent struct {
	BaseEvent
	LeadCode      string    `json:"lead_code"`
	OpportunityID uuid.UUID `json:"opportunity_id"`
	CustomerID    *uuid.UUID `json:"customer_id,omitempty"`
}

// NewLeadConvertedEvent creates a new lead converted event.
func NewLeadConvertedEvent(lead *Lead, opportunityID uuid.UUID) *LeadConvertedEvent {
	return &LeadConvertedEvent{
		BaseEvent:     newBaseEvent("lead.converted", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:      lead.Code,
		OpportunityID: opportunityID,
		CustomerID:    lead.ConversionInfo.CustomerID,
	}
}

// LeadOwnerAssignedEvent is raised when a lead owner is assigned.
type LeadOwnerAssignedEvent struct {
	BaseEvent
	LeadCode       string     `json:"lead_code"`
	NewOwnerID     uuid.UUID  `json:"new_owner_id"`
	PreviousOwnerID *uuid.UUID `json:"previous_owner_id,omitempty"`
}

// NewLeadOwnerAssignedEvent creates a new lead owner assigned event.
func NewLeadOwnerAssignedEvent(lead *Lead, previousOwnerID *uuid.UUID) *LeadOwnerAssignedEvent {
	return &LeadOwnerAssignedEvent{
		BaseEvent:       newBaseEvent("lead.owner_assigned", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:        lead.Code,
		NewOwnerID:      *lead.OwnerID,
		PreviousOwnerID: previousOwnerID,
	}
}

// LeadScoredEvent is raised when a lead score changes.
type LeadScoredEvent struct {
	BaseEvent
	LeadCode string     `json:"lead_code"`
	Score    int        `json:"score"`
	Rating   LeadRating `json:"rating"`
}

// NewLeadScoredEvent creates a new lead scored event.
func NewLeadScoredEvent(lead *Lead) *LeadScoredEvent {
	return &LeadScoredEvent{
		BaseEvent: newBaseEvent("lead.scored", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:  lead.Code,
		Score:     lead.Score.Score,
		Rating:    lead.Rating,
	}
}

// LeadDeletedEvent is raised when a lead is deleted.
type LeadDeletedEvent struct {
	BaseEvent
	LeadCode string `json:"lead_code"`
}

// NewLeadDeletedEvent creates a new lead deleted event.
func NewLeadDeletedEvent(lead *Lead) *LeadDeletedEvent {
	return &LeadDeletedEvent{
		BaseEvent: newBaseEvent("lead.deleted", "lead", lead.ID, lead.TenantID, lead.Version),
		LeadCode:  lead.Code,
	}
}

// ============================================================================
// Opportunity Events
// ============================================================================

// OpportunityCreatedEvent is raised when an opportunity is created.
type OpportunityCreatedEvent struct {
	BaseEvent
	OpportunityCode string    `json:"opportunity_code"`
	Name            string    `json:"name"`
	CustomerID      uuid.UUID `json:"customer_id"`
	PipelineID      uuid.UUID `json:"pipeline_id"`
	StageID         uuid.UUID `json:"stage_id"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	OwnerID         uuid.UUID `json:"owner_id"`
	LeadID          *uuid.UUID `json:"lead_id,omitempty"`
}

// NewOpportunityCreatedEvent creates a new opportunity created event.
func NewOpportunityCreatedEvent(opp *Opportunity) *OpportunityCreatedEvent {
	return &OpportunityCreatedEvent{
		BaseEvent:       newBaseEvent("opportunity.created", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
		Name:            opp.Name,
		CustomerID:      opp.CustomerID,
		PipelineID:      opp.PipelineID,
		StageID:         opp.StageID,
		Amount:          opp.Amount.Amount,
		Currency:        opp.Amount.Currency,
		OwnerID:         opp.OwnerID,
		LeadID:          opp.LeadID,
	}
}

// OpportunityUpdatedEvent is raised when an opportunity is updated.
type OpportunityUpdatedEvent struct {
	BaseEvent
	OpportunityCode string `json:"opportunity_code"`
}

// NewOpportunityUpdatedEvent creates a new opportunity updated event.
func NewOpportunityUpdatedEvent(opp *Opportunity) *OpportunityUpdatedEvent {
	return &OpportunityUpdatedEvent{
		BaseEvent:       newBaseEvent("opportunity.updated", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
	}
}

// OpportunityStageChangedEvent is raised when an opportunity moves to a new stage.
type OpportunityStageChangedEvent struct {
	BaseEvent
	OpportunityCode string    `json:"opportunity_code"`
	OldStageID      uuid.UUID `json:"old_stage_id"`
	OldStageName    string    `json:"old_stage_name"`
	NewStageID      uuid.UUID `json:"new_stage_id"`
	NewStageName    string    `json:"new_stage_name"`
	Probability     int       `json:"probability"`
}

// NewOpportunityStageChangedEvent creates a new opportunity stage changed event.
func NewOpportunityStageChangedEvent(opp *Opportunity, oldStageID uuid.UUID, oldStageName string) *OpportunityStageChangedEvent {
	return &OpportunityStageChangedEvent{
		BaseEvent:       newBaseEvent("opportunity.stage_changed", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
		OldStageID:      oldStageID,
		OldStageName:    oldStageName,
		NewStageID:      opp.StageID,
		NewStageName:    opp.StageName,
		Probability:     opp.Probability,
	}
}

// OpportunityAmountChangedEvent is raised when an opportunity amount changes.
type OpportunityAmountChangedEvent struct {
	BaseEvent
	OpportunityCode string `json:"opportunity_code"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	WeightedAmount  int64  `json:"weighted_amount"`
}

// NewOpportunityAmountChangedEvent creates a new opportunity amount changed event.
func NewOpportunityAmountChangedEvent(opp *Opportunity) *OpportunityAmountChangedEvent {
	return &OpportunityAmountChangedEvent{
		BaseEvent:       newBaseEvent("opportunity.amount_changed", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
		Amount:          opp.Amount.Amount,
		Currency:        opp.Amount.Currency,
		WeightedAmount:  opp.WeightedAmount.Amount,
	}
}

// OpportunityWonEvent is raised when an opportunity is won.
type OpportunityWonEvent struct {
	BaseEvent
	OpportunityCode string    `json:"opportunity_code"`
	CustomerID      uuid.UUID `json:"customer_id"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	WonReason       string    `json:"won_reason"`
	WonBy           uuid.UUID `json:"won_by"`
}

// NewOpportunityWonEvent creates a new opportunity won event.
func NewOpportunityWonEvent(opp *Opportunity, reason string) *OpportunityWonEvent {
	return &OpportunityWonEvent{
		BaseEvent:       newBaseEvent("opportunity.won", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
		CustomerID:      opp.CustomerID,
		Amount:          opp.Amount.Amount,
		Currency:        opp.Amount.Currency,
		WonReason:       reason,
		WonBy:           opp.CloseInfo.ClosedBy,
	}
}

// OpportunityLostEvent is raised when an opportunity is lost.
type OpportunityLostEvent struct {
	BaseEvent
	OpportunityCode string  `json:"opportunity_code"`
	LostReason      string  `json:"lost_reason"`
	CompetitorName  string  `json:"competitor_name,omitempty"`
	LostBy          uuid.UUID `json:"lost_by"`
}

// NewOpportunityLostEvent creates a new opportunity lost event.
func NewOpportunityLostEvent(opp *Opportunity, reason, competitorName string) *OpportunityLostEvent {
	return &OpportunityLostEvent{
		BaseEvent:       newBaseEvent("opportunity.lost", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
		LostReason:      reason,
		CompetitorName:  competitorName,
		LostBy:          opp.CloseInfo.ClosedBy,
	}
}

// OpportunityReopenedEvent is raised when an opportunity is reopened.
type OpportunityReopenedEvent struct {
	BaseEvent
	OpportunityCode string `json:"opportunity_code"`
}

// NewOpportunityReopenedEvent creates a new opportunity reopened event.
func NewOpportunityReopenedEvent(opp *Opportunity) *OpportunityReopenedEvent {
	return &OpportunityReopenedEvent{
		BaseEvent:       newBaseEvent("opportunity.reopened", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
	}
}

// OpportunityOwnerChangedEvent is raised when an opportunity owner changes.
type OpportunityOwnerChangedEvent struct {
	BaseEvent
	OpportunityCode string    `json:"opportunity_code"`
	NewOwnerID      uuid.UUID `json:"new_owner_id"`
	PreviousOwnerID uuid.UUID `json:"previous_owner_id"`
}

// NewOpportunityOwnerChangedEvent creates a new opportunity owner changed event.
func NewOpportunityOwnerChangedEvent(opp *Opportunity, previousOwnerID uuid.UUID) *OpportunityOwnerChangedEvent {
	return &OpportunityOwnerChangedEvent{
		BaseEvent:       newBaseEvent("opportunity.owner_changed", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
		NewOwnerID:      opp.OwnerID,
		PreviousOwnerID: previousOwnerID,
	}
}

// OpportunityDeletedEvent is raised when an opportunity is deleted.
type OpportunityDeletedEvent struct {
	BaseEvent
	OpportunityCode string `json:"opportunity_code"`
}

// NewOpportunityDeletedEvent creates a new opportunity deleted event.
func NewOpportunityDeletedEvent(opp *Opportunity) *OpportunityDeletedEvent {
	return &OpportunityDeletedEvent{
		BaseEvent:       newBaseEvent("opportunity.deleted", "opportunity", opp.ID, opp.TenantID, opp.Version),
		OpportunityCode: opp.Code,
	}
}

// ============================================================================
// Deal Events
// ============================================================================

// DealCreatedEvent is raised when a deal is created.
type DealCreatedEvent struct {
	BaseEvent
	DealCode      string    `json:"deal_code"`
	Name          string    `json:"name"`
	OpportunityID uuid.UUID `json:"opportunity_id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	TotalAmount   int64     `json:"total_amount"`
	Currency      string    `json:"currency"`
	OwnerID       uuid.UUID `json:"owner_id"`
}

// NewDealCreatedEvent creates a new deal created event.
func NewDealCreatedEvent(deal *Deal) *DealCreatedEvent {
	return &DealCreatedEvent{
		BaseEvent:     newBaseEvent("deal.created", "deal", deal.ID, deal.TenantID, deal.Version),
		DealCode:      deal.Code,
		Name:          deal.Name,
		OpportunityID: deal.OpportunityID,
		CustomerID:    deal.CustomerID,
		TotalAmount:   deal.TotalAmount.Amount,
		Currency:      deal.Currency,
		OwnerID:       deal.OwnerID,
	}
}

// DealUpdatedEvent is raised when a deal is updated.
type DealUpdatedEvent struct {
	BaseEvent
	DealCode string `json:"deal_code"`
}

// NewDealUpdatedEvent creates a new deal updated event.
func NewDealUpdatedEvent(deal *Deal) *DealUpdatedEvent {
	return &DealUpdatedEvent{
		BaseEvent: newBaseEvent("deal.updated", "deal", deal.ID, deal.TenantID, deal.Version),
		DealCode:  deal.Code,
	}
}

// DealActivatedEvent is raised when a deal is activated.
type DealActivatedEvent struct {
	BaseEvent
	DealCode    string    `json:"deal_code"`
	ActivatedAt time.Time `json:"activated_at"`
}

// NewDealActivatedEvent creates a new deal activated event.
func NewDealActivatedEvent(deal *Deal) *DealActivatedEvent {
	return &DealActivatedEvent{
		BaseEvent:   newBaseEvent("deal.activated", "deal", deal.ID, deal.TenantID, deal.Version),
		DealCode:    deal.Code,
		ActivatedAt: *deal.ActivatedAt,
	}
}

// DealFulfilledEvent is raised when a deal is fulfilled.
type DealFulfilledEvent struct {
	BaseEvent
	DealCode    string    `json:"deal_code"`
	CustomerID  uuid.UUID `json:"customer_id"`
	TotalAmount int64     `json:"total_amount"`
	Currency    string    `json:"currency"`
}

// NewDealFulfilledEvent creates a new deal fulfilled event.
func NewDealFulfilledEvent(deal *Deal) *DealFulfilledEvent {
	return &DealFulfilledEvent{
		BaseEvent:   newBaseEvent("deal.fulfilled", "deal", deal.ID, deal.TenantID, deal.Version),
		DealCode:    deal.Code,
		CustomerID:  deal.CustomerID,
		TotalAmount: deal.TotalAmount.Amount,
		Currency:    deal.Currency,
	}
}

// DealCancelledEvent is raised when a deal is cancelled.
type DealCancelledEvent struct {
	BaseEvent
	DealCode string `json:"deal_code"`
	Reason   string `json:"reason"`
}

// NewDealCancelledEvent creates a new deal cancelled event.
func NewDealCancelledEvent(deal *Deal, reason string) *DealCancelledEvent {
	return &DealCancelledEvent{
		BaseEvent: newBaseEvent("deal.cancelled", "deal", deal.ID, deal.TenantID, deal.Version),
		DealCode:  deal.Code,
		Reason:    reason,
	}
}

// DealInvoiceCreatedEvent is raised when an invoice is created for a deal.
type DealInvoiceCreatedEvent struct {
	BaseEvent
	DealCode      string    `json:"deal_code"`
	InvoiceID     uuid.UUID `json:"invoice_id"`
	InvoiceNumber string    `json:"invoice_number"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	DueDate       time.Time `json:"due_date"`
}

// NewDealInvoiceCreatedEvent creates a new deal invoice created event.
func NewDealInvoiceCreatedEvent(deal *Deal, invoice *Invoice) *DealInvoiceCreatedEvent {
	return &DealInvoiceCreatedEvent{
		BaseEvent:     newBaseEvent("deal.invoice_created", "deal", deal.ID, deal.TenantID, deal.Version),
		DealCode:      deal.Code,
		InvoiceID:     invoice.ID,
		InvoiceNumber: invoice.InvoiceNumber,
		Amount:        invoice.Amount.Amount,
		Currency:      deal.Currency,
		DueDate:       invoice.DueDate,
	}
}

// DealPaymentReceivedEvent is raised when a payment is received for a deal.
type DealPaymentReceivedEvent struct {
	BaseEvent
	DealCode          string     `json:"deal_code"`
	PaymentID         uuid.UUID  `json:"payment_id"`
	InvoiceID         *uuid.UUID `json:"invoice_id,omitempty"`
	Amount            int64      `json:"amount"`
	Currency          string     `json:"currency"`
	PaymentMethod     string     `json:"payment_method"`
	OutstandingAmount int64      `json:"outstanding_amount"`
}

// NewDealPaymentReceivedEvent creates a new deal payment received event.
func NewDealPaymentReceivedEvent(deal *Deal, payment *Payment) *DealPaymentReceivedEvent {
	return &DealPaymentReceivedEvent{
		BaseEvent:         newBaseEvent("deal.payment_received", "deal", deal.ID, deal.TenantID, deal.Version),
		DealCode:          deal.Code,
		PaymentID:         payment.ID,
		InvoiceID:         payment.InvoiceID,
		Amount:            payment.Amount.Amount,
		Currency:          deal.Currency,
		PaymentMethod:     payment.PaymentMethod,
		OutstandingAmount: deal.OutstandingAmount.Amount,
	}
}

// DealDeletedEvent is raised when a deal is deleted.
type DealDeletedEvent struct {
	BaseEvent
	DealCode string `json:"deal_code"`
}

// NewDealDeletedEvent creates a new deal deleted event.
func NewDealDeletedEvent(deal *Deal) *DealDeletedEvent {
	return &DealDeletedEvent{
		BaseEvent: newBaseEvent("deal.deleted", "deal", deal.ID, deal.TenantID, deal.Version),
		DealCode:  deal.Code,
	}
}

// ============================================================================
// Pipeline Events
// ============================================================================

// PipelineCreatedEvent is raised when a pipeline is created.
type PipelineCreatedEvent struct {
	BaseEvent
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// NewPipelineCreatedEvent creates a new pipeline created event.
func NewPipelineCreatedEvent(pipeline *Pipeline) *PipelineCreatedEvent {
	return &PipelineCreatedEvent{
		BaseEvent: newBaseEvent("pipeline.created", "pipeline", pipeline.ID, pipeline.TenantID, pipeline.Version),
		Name:      pipeline.Name,
		Currency:  pipeline.Currency,
	}
}

// PipelineUpdatedEvent is raised when a pipeline is updated.
type PipelineUpdatedEvent struct {
	BaseEvent
	Name string `json:"name"`
}

// NewPipelineUpdatedEvent creates a new pipeline updated event.
func NewPipelineUpdatedEvent(pipeline *Pipeline) *PipelineUpdatedEvent {
	return &PipelineUpdatedEvent{
		BaseEvent: newBaseEvent("pipeline.updated", "pipeline", pipeline.ID, pipeline.TenantID, pipeline.Version),
		Name:      pipeline.Name,
	}
}
