// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Opportunity errors
var (
	ErrOpportunityNotFound        = errors.New("opportunity not found")
	ErrOpportunityAlreadyExists   = errors.New("opportunity already exists")
	ErrOpportunityAlreadyClosed   = errors.New("opportunity is already closed")
	ErrOpportunityNotClosed       = errors.New("opportunity is not closed")
	ErrInvalidOpportunityStatus   = errors.New("invalid opportunity status")
	ErrInvalidStageTransition     = errors.New("invalid stage transition")
	ErrOpportunityVersionMismatch = errors.New("opportunity version mismatch")
	ErrCannotReopenAfterDays      = errors.New("cannot reopen opportunity after 30 days")
)

// OpportunityStatus represents the status of an opportunity.
type OpportunityStatus string

const (
	OpportunityStatusOpen   OpportunityStatus = "open"
	OpportunityStatusWon    OpportunityStatus = "won"
	OpportunityStatusLost   OpportunityStatus = "lost"
)

// ValidOpportunityStatuses returns all valid opportunity statuses.
func ValidOpportunityStatuses() []OpportunityStatus {
	return []OpportunityStatus{
		OpportunityStatusOpen,
		OpportunityStatusWon,
		OpportunityStatusLost,
	}
}

// IsValid checks if the opportunity status is valid.
func (s OpportunityStatus) IsValid() bool {
	for _, valid := range ValidOpportunityStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// IsClosed returns true if the status is closed (won or lost).
func (s OpportunityStatus) IsClosed() bool {
	return s == OpportunityStatusWon || s == OpportunityStatusLost
}

// OpportunityPriority represents the priority of an opportunity.
type OpportunityPriority string

const (
	OpportunityPriorityLow    OpportunityPriority = "low"
	OpportunityPriorityMedium OpportunityPriority = "medium"
	OpportunityPriorityHigh   OpportunityPriority = "high"
	OpportunityPriorityCritical OpportunityPriority = "critical"
)

// OpportunityProduct represents a product/service in an opportunity.
type OpportunityProduct struct {
	ID          uuid.UUID `json:"id" bson:"id"`
	ProductID   uuid.UUID `json:"product_id" bson:"product_id"`
	ProductName string    `json:"product_name" bson:"product_name"`
	SKU         string    `json:"sku,omitempty" bson:"sku,omitempty"`
	Quantity    int       `json:"quantity" bson:"quantity"`
	UnitPrice   Money     `json:"unit_price" bson:"unit_price"`
	Discount    float64   `json:"discount" bson:"discount"` // Percentage
	Tax         float64   `json:"tax" bson:"tax"`           // Percentage
	TotalPrice  Money     `json:"total_price" bson:"total_price"`
	Notes       string    `json:"notes,omitempty" bson:"notes,omitempty"`
}

// CalculateTotalPrice calculates the total price for the product line.
func (p *OpportunityProduct) CalculateTotalPrice() {
	basePrice := p.UnitPrice.Multiply(float64(p.Quantity))
	discountAmount := basePrice.Multiply(p.Discount / 100)
	afterDiscount, _ := basePrice.Subtract(discountAmount)
	taxAmount := afterDiscount.Multiply(p.Tax / 100)
	total, _ := afterDiscount.Add(taxAmount)
	p.TotalPrice = total
}

// OpportunityContact represents a contact associated with an opportunity.
type OpportunityContact struct {
	ContactID   uuid.UUID `json:"contact_id" bson:"contact_id"`
	CustomerID  uuid.UUID `json:"customer_id" bson:"customer_id"`
	Name        string    `json:"name" bson:"name"`
	Email       string    `json:"email" bson:"email"`
	Phone       string    `json:"phone,omitempty" bson:"phone,omitempty"`
	Role        string    `json:"role" bson:"role"` // decision_maker, influencer, gatekeeper, etc.
	IsPrimary   bool      `json:"is_primary" bson:"is_primary"`
}

// StageHistory represents a stage transition in the opportunity's history.
type StageHistory struct {
	StageID     uuid.UUID  `json:"stage_id" bson:"stage_id"`
	StageName   string     `json:"stage_name" bson:"stage_name"`
	EnteredAt   time.Time  `json:"entered_at" bson:"entered_at"`
	ExitedAt    *time.Time `json:"exited_at,omitempty" bson:"exited_at,omitempty"`
	Duration    int        `json:"duration" bson:"duration"` // Duration in hours
	MovedBy     uuid.UUID  `json:"moved_by" bson:"moved_by"`
	Notes       string     `json:"notes,omitempty" bson:"notes,omitempty"`
}

// CloseInfo contains information about closing an opportunity.
type CloseInfo struct {
	ClosedAt     time.Time  `json:"closed_at" bson:"closed_at"`
	ClosedBy     uuid.UUID  `json:"closed_by" bson:"closed_by"`
	Reason       string     `json:"reason" bson:"reason"`
	Notes        string     `json:"notes,omitempty" bson:"notes,omitempty"`
	CompetitorID *uuid.UUID `json:"competitor_id,omitempty" bson:"competitor_id,omitempty"`
	CompetitorName string   `json:"competitor_name,omitempty" bson:"competitor_name,omitempty"`
}

// Opportunity represents a sales opportunity in the pipeline.
type Opportunity struct {
	ID               uuid.UUID              `json:"id" bson:"_id"`
	TenantID         uuid.UUID              `json:"tenant_id" bson:"tenant_id"`
	Code             string                 `json:"code" bson:"code"` // e.g., "OP-2024-001"
	Name             string                 `json:"name" bson:"name"`
	Description      string                 `json:"description,omitempty" bson:"description,omitempty"`
	Status           OpportunityStatus      `json:"status" bson:"status"`
	Priority         OpportunityPriority    `json:"priority" bson:"priority"`

	// Pipeline and Stage
	PipelineID       uuid.UUID              `json:"pipeline_id" bson:"pipeline_id"`
	PipelineName     string                 `json:"pipeline_name" bson:"pipeline_name"`
	StageID          uuid.UUID              `json:"stage_id" bson:"stage_id"`
	StageName        string                 `json:"stage_name" bson:"stage_name"`
	StageEnteredAt   time.Time              `json:"stage_entered_at" bson:"stage_entered_at"`
	StageHistory     []StageHistory         `json:"stage_history" bson:"stage_history"`

	// Customer and Contacts
	CustomerID       uuid.UUID              `json:"customer_id" bson:"customer_id"`
	CustomerName     string                 `json:"customer_name" bson:"customer_name"`
	Contacts         []OpportunityContact   `json:"contacts" bson:"contacts"`

	// Value
	Amount           Money                  `json:"amount" bson:"amount"`
	WeightedAmount   Money                  `json:"weighted_amount" bson:"weighted_amount"`
	Probability      int                    `json:"probability" bson:"probability"` // 0-100
	Products         []OpportunityProduct   `json:"products,omitempty" bson:"products,omitempty"`

	// Timeline
	ExpectedCloseDate *time.Time            `json:"expected_close_date,omitempty" bson:"expected_close_date,omitempty"`
	ActualCloseDate   *time.Time            `json:"actual_close_date,omitempty" bson:"actual_close_date,omitempty"`

	// Source
	LeadID           *uuid.UUID             `json:"lead_id,omitempty" bson:"lead_id,omitempty"`
	Source           string                 `json:"source,omitempty" bson:"source,omitempty"`
	Campaign         string                 `json:"campaign,omitempty" bson:"campaign,omitempty"`
	CampaignID       *uuid.UUID             `json:"campaign_id,omitempty" bson:"campaign_id,omitempty"`

	// Ownership
	OwnerID          uuid.UUID              `json:"owner_id" bson:"owner_id"`
	OwnerName        string                 `json:"owner_name" bson:"owner_name"`
	TeamID           *uuid.UUID             `json:"team_id,omitempty" bson:"team_id,omitempty"`

	// Close Information
	CloseInfo        *CloseInfo             `json:"close_info,omitempty" bson:"close_info,omitempty"`

	// Activity Tracking
	LastActivityAt   *time.Time             `json:"last_activity_at,omitempty" bson:"last_activity_at,omitempty"`
	NextActivityAt   *time.Time             `json:"next_activity_at,omitempty" bson:"next_activity_at,omitempty"`
	ActivityCount    int                    `json:"activity_count" bson:"activity_count"`

	// Metadata
	Tags             []string               `json:"tags,omitempty" bson:"tags,omitempty"`
	CustomFields     map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
	Notes            string                 `json:"notes,omitempty" bson:"notes,omitempty"`

	// Timestamps
	CreatedBy        uuid.UUID              `json:"created_by" bson:"created_by"`
	CreatedAt        time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" bson:"updated_at"`
	DeletedAt        *time.Time             `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Version          int                    `json:"version" bson:"version"`

	// Domain events
	events []DomainEvent `json:"-" bson:"-"`
}

// NewOpportunity creates a new opportunity.
func NewOpportunity(
	tenantID uuid.UUID,
	name string,
	pipeline *Pipeline,
	customerID uuid.UUID,
	customerName string,
	amount Money,
	ownerID uuid.UUID,
	ownerName string,
	createdBy uuid.UUID,
) (*Opportunity, error) {
	if name == "" {
		return nil, errors.New("opportunity name is required")
	}
	if pipeline == nil {
		return nil, errors.New("pipeline is required")
	}

	firstStage := pipeline.GetFirstStage()
	if firstStage == nil {
		return nil, errors.New("pipeline has no stages")
	}

	now := time.Now().UTC()

	// Calculate weighted amount based on first stage probability
	weightedAmount := amount.Multiply(float64(firstStage.Probability) / 100)

	opp := &Opportunity{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Name:           name,
		Status:         OpportunityStatusOpen,
		Priority:       OpportunityPriorityMedium,
		PipelineID:     pipeline.ID,
		PipelineName:   pipeline.Name,
		StageID:        firstStage.ID,
		StageName:      firstStage.Name,
		StageEnteredAt: now,
		StageHistory: []StageHistory{
			{
				StageID:   firstStage.ID,
				StageName: firstStage.Name,
				EnteredAt: now,
				MovedBy:   createdBy,
			},
		},
		CustomerID:     customerID,
		CustomerName:   customerName,
		Contacts:       make([]OpportunityContact, 0),
		Amount:         amount,
		WeightedAmount: weightedAmount,
		Probability:    firstStage.Probability,
		Products:       make([]OpportunityProduct, 0),
		OwnerID:        ownerID,
		OwnerName:      ownerName,
		Tags:           make([]string, 0),
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
		Version:        1,
		events:         make([]DomainEvent, 0),
	}

	opp.AddEvent(NewOpportunityCreatedEvent(opp))
	return opp, nil
}

// NewOpportunityFromLead creates an opportunity from a converted lead.
func NewOpportunityFromLead(
	lead *Lead,
	pipeline *Pipeline,
	customerID uuid.UUID,
	customerName string,
	createdBy uuid.UUID,
) (*Opportunity, error) {
	if !lead.CanConvert() {
		return nil, ErrLeadNotQualified
	}

	opp, err := NewOpportunity(
		lead.TenantID,
		lead.Company.Name+" - "+lead.Contact.FullName(),
		pipeline,
		customerID,
		customerName,
		lead.EstimatedValue,
		*lead.OwnerID,
		lead.OwnerName,
		createdBy,
	)
	if err != nil {
		return nil, err
	}

	opp.LeadID = &lead.ID
	opp.Source = string(lead.Source)
	opp.Campaign = lead.Campaign
	opp.CampaignID = lead.CampaignID

	// Add lead contact as opportunity contact
	opp.AddContact(OpportunityContact{
		ContactID:  uuid.New(), // Will be updated after contact creation
		CustomerID: customerID,
		Name:       lead.Contact.FullName(),
		Email:      lead.Contact.Email,
		Phone:      lead.Contact.Phone,
		Role:       "primary",
		IsPrimary:  true,
	})

	return opp, nil
}

// Update updates opportunity details.
func (o *Opportunity) Update(name, description string, priority OpportunityPriority) {
	if name != "" {
		o.Name = name
	}
	o.Description = description
	o.Priority = priority
	o.UpdatedAt = time.Now().UTC()

	o.AddEvent(NewOpportunityUpdatedEvent(o))
}

// SetAmount updates the opportunity amount.
func (o *Opportunity) SetAmount(amount Money) {
	o.Amount = amount
	o.recalculateWeightedAmount()
	o.UpdatedAt = time.Now().UTC()

	o.AddEvent(NewOpportunityAmountChangedEvent(o))
}

// recalculateWeightedAmount recalculates the weighted amount.
func (o *Opportunity) recalculateWeightedAmount() {
	o.WeightedAmount = o.Amount.Multiply(float64(o.Probability) / 100)
}

// MoveToStage moves the opportunity to a new stage.
func (o *Opportunity) MoveToStage(newStage *Stage, movedBy uuid.UUID, notes string) error {
	if o.Status.IsClosed() {
		return ErrOpportunityAlreadyClosed
	}

	now := time.Now().UTC()

	// Close the current stage history entry
	if len(o.StageHistory) > 0 {
		lastEntry := &o.StageHistory[len(o.StageHistory)-1]
		lastEntry.ExitedAt = &now
		lastEntry.Duration = int(now.Sub(lastEntry.EnteredAt).Hours())
	}

	oldStageID := o.StageID
	oldStageName := o.StageName

	// Update to new stage
	o.StageID = newStage.ID
	o.StageName = newStage.Name
	o.StageEnteredAt = now
	o.Probability = newStage.Probability
	o.recalculateWeightedAmount()

	// Add new stage history entry
	o.StageHistory = append(o.StageHistory, StageHistory{
		StageID:   newStage.ID,
		StageName: newStage.Name,
		EnteredAt: now,
		MovedBy:   movedBy,
		Notes:     notes,
	})

	o.UpdatedAt = now

	o.AddEvent(NewOpportunityStageChangedEvent(o, oldStageID, oldStageName))
	return nil
}

// Win marks the opportunity as won.
func (o *Opportunity) Win(wonStage *Stage, reason, notes string, wonBy uuid.UUID) error {
	if o.Status.IsClosed() {
		return ErrOpportunityAlreadyClosed
	}

	now := time.Now().UTC()

	// Move to won stage
	if err := o.MoveToStage(wonStage, wonBy, "Won: "+reason); err != nil {
		return err
	}

	o.Status = OpportunityStatusWon
	o.Probability = 100
	o.recalculateWeightedAmount()
	o.ActualCloseDate = &now
	o.CloseInfo = &CloseInfo{
		ClosedAt: now,
		ClosedBy: wonBy,
		Reason:   reason,
		Notes:    notes,
	}
	o.UpdatedAt = now

	o.AddEvent(NewOpportunityWonEvent(o, reason))
	return nil
}

// Lose marks the opportunity as lost.
func (o *Opportunity) Lose(lostStage *Stage, reason, notes string, competitorID *uuid.UUID, competitorName string, lostBy uuid.UUID) error {
	if o.Status.IsClosed() {
		return ErrOpportunityAlreadyClosed
	}

	now := time.Now().UTC()

	// Move to lost stage
	if err := o.MoveToStage(lostStage, lostBy, "Lost: "+reason); err != nil {
		return err
	}

	o.Status = OpportunityStatusLost
	o.Probability = 0
	o.recalculateWeightedAmount()
	o.ActualCloseDate = &now
	o.CloseInfo = &CloseInfo{
		ClosedAt:       now,
		ClosedBy:       lostBy,
		Reason:         reason,
		Notes:          notes,
		CompetitorID:   competitorID,
		CompetitorName: competitorName,
	}
	o.UpdatedAt = now

	o.AddEvent(NewOpportunityLostEvent(o, reason, competitorName))
	return nil
}

// Reopen reopens a closed opportunity.
func (o *Opportunity) Reopen(pipeline *Pipeline, reopenedBy uuid.UUID, notes string) error {
	if !o.Status.IsClosed() {
		return ErrOpportunityNotClosed
	}

	// Check if opportunity can be reopened (within 30 days)
	if o.ActualCloseDate != nil {
		daysSinceClosed := int(time.Since(*o.ActualCloseDate).Hours() / 24)
		if daysSinceClosed > 30 {
			return ErrCannotReopenAfterDays
		}
	}

	firstStage := pipeline.GetFirstStage()
	if firstStage == nil {
		return errors.New("pipeline has no stages")
	}

	now := time.Now().UTC()

	o.Status = OpportunityStatusOpen
	o.StageID = firstStage.ID
	o.StageName = firstStage.Name
	o.StageEnteredAt = now
	o.Probability = firstStage.Probability
	o.recalculateWeightedAmount()
	o.ActualCloseDate = nil
	o.CloseInfo = nil

	// Add reopened stage to history
	o.StageHistory = append(o.StageHistory, StageHistory{
		StageID:   firstStage.ID,
		StageName: firstStage.Name + " (Reopened)",
		EnteredAt: now,
		MovedBy:   reopenedBy,
		Notes:     notes,
	})

	o.UpdatedAt = now

	o.AddEvent(NewOpportunityReopenedEvent(o))
	return nil
}

// AddContact adds a contact to the opportunity.
func (o *Opportunity) AddContact(contact OpportunityContact) {
	// If this is the first contact or marked as primary, ensure only one primary
	if contact.IsPrimary {
		for i := range o.Contacts {
			o.Contacts[i].IsPrimary = false
		}
	}

	o.Contacts = append(o.Contacts, contact)
	o.UpdatedAt = time.Now().UTC()
}

// RemoveContact removes a contact from the opportunity.
func (o *Opportunity) RemoveContact(contactID uuid.UUID) {
	for i, c := range o.Contacts {
		if c.ContactID == contactID {
			o.Contacts = append(o.Contacts[:i], o.Contacts[i+1:]...)
			o.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// SetPrimaryContact sets a contact as the primary contact.
func (o *Opportunity) SetPrimaryContact(contactID uuid.UUID) {
	for i := range o.Contacts {
		o.Contacts[i].IsPrimary = o.Contacts[i].ContactID == contactID
	}
	o.UpdatedAt = time.Now().UTC()
}

// GetPrimaryContact returns the primary contact.
func (o *Opportunity) GetPrimaryContact() *OpportunityContact {
	for _, c := range o.Contacts {
		if c.IsPrimary {
			return &c
		}
	}
	if len(o.Contacts) > 0 {
		return &o.Contacts[0]
	}
	return nil
}

// AddProduct adds a product to the opportunity.
func (o *Opportunity) AddProduct(product OpportunityProduct) {
	product.ID = uuid.New()
	product.CalculateTotalPrice()
	o.Products = append(o.Products, product)
	o.recalculateAmountFromProducts()
	o.UpdatedAt = time.Now().UTC()
}

// UpdateProduct updates a product in the opportunity.
func (o *Opportunity) UpdateProduct(productID uuid.UUID, quantity int, unitPrice Money, discount, tax float64) error {
	for i := range o.Products {
		if o.Products[i].ID == productID {
			o.Products[i].Quantity = quantity
			o.Products[i].UnitPrice = unitPrice
			o.Products[i].Discount = discount
			o.Products[i].Tax = tax
			o.Products[i].CalculateTotalPrice()
			o.recalculateAmountFromProducts()
			o.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return errors.New("product not found")
}

// RemoveProduct removes a product from the opportunity.
func (o *Opportunity) RemoveProduct(productID uuid.UUID) {
	for i, p := range o.Products {
		if p.ID == productID {
			o.Products = append(o.Products[:i], o.Products[i+1:]...)
			o.recalculateAmountFromProducts()
			o.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// recalculateAmountFromProducts recalculates the total amount from products.
func (o *Opportunity) recalculateAmountFromProducts() {
	if len(o.Products) == 0 {
		return
	}

	total := o.Products[0].TotalPrice
	for i := 1; i < len(o.Products); i++ {
		result, _ := total.Add(o.Products[i].TotalPrice)
		total = result
	}

	o.Amount = total
	o.recalculateWeightedAmount()
}

// AssignOwner assigns a new owner.
func (o *Opportunity) AssignOwner(ownerID uuid.UUID, ownerName string) {
	oldOwnerID := o.OwnerID
	o.OwnerID = ownerID
	o.OwnerName = ownerName
	o.UpdatedAt = time.Now().UTC()

	o.AddEvent(NewOpportunityOwnerChangedEvent(o, oldOwnerID))
}

// SetExpectedCloseDate sets the expected close date.
func (o *Opportunity) SetExpectedCloseDate(date time.Time) {
	o.ExpectedCloseDate = &date
	o.UpdatedAt = time.Now().UTC()
}

// RecordActivity records an activity on the opportunity.
func (o *Opportunity) RecordActivity(activityAt time.Time) {
	o.LastActivityAt = &activityAt
	o.ActivityCount++
	o.UpdatedAt = time.Now().UTC()
}

// SetNextActivity sets the next scheduled activity.
func (o *Opportunity) SetNextActivity(activityAt time.Time) {
	o.NextActivityAt = &activityAt
	o.UpdatedAt = time.Now().UTC()
}

// AddTag adds a tag.
func (o *Opportunity) AddTag(tag string) {
	for _, t := range o.Tags {
		if t == tag {
			return
		}
	}
	o.Tags = append(o.Tags, tag)
	o.UpdatedAt = time.Now().UTC()
}

// RemoveTag removes a tag.
func (o *Opportunity) RemoveTag(tag string) {
	for i, t := range o.Tags {
		if t == tag {
			o.Tags = append(o.Tags[:i], o.Tags[i+1:]...)
			o.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// SetCustomField sets a custom field value.
func (o *Opportunity) SetCustomField(key string, value interface{}) {
	if o.CustomFields == nil {
		o.CustomFields = make(map[string]interface{})
	}
	o.CustomFields[key] = value
	o.UpdatedAt = time.Now().UTC()
}

// Delete soft deletes the opportunity.
func (o *Opportunity) Delete() {
	now := time.Now().UTC()
	o.DeletedAt = &now
	o.UpdatedAt = now

	o.AddEvent(NewOpportunityDeletedEvent(o))
}

// Restore restores a soft-deleted opportunity.
func (o *Opportunity) Restore() {
	o.DeletedAt = nil
	o.UpdatedAt = time.Now().UTC()
}

// IsDeleted returns true if the opportunity is deleted.
func (o *Opportunity) IsDeleted() bool {
	return o.DeletedAt != nil
}

// IsOpen returns true if the opportunity is open.
func (o *Opportunity) IsOpen() bool {
	return o.Status == OpportunityStatusOpen
}

// IsWon returns true if the opportunity is won.
func (o *Opportunity) IsWon() bool {
	return o.Status == OpportunityStatusWon
}

// IsLost returns true if the opportunity is lost.
func (o *Opportunity) IsLost() bool {
	return o.Status == OpportunityStatusLost
}

// DaysInPipeline returns the number of days in the pipeline.
func (o *Opportunity) DaysInPipeline() int {
	end := time.Now()
	if o.ActualCloseDate != nil {
		end = *o.ActualCloseDate
	}
	return int(end.Sub(o.CreatedAt).Hours() / 24)
}

// DaysInCurrentStage returns the number of days in current stage.
func (o *Opportunity) DaysInCurrentStage() int {
	return int(time.Since(o.StageEnteredAt).Hours() / 24)
}

// DaysUntilExpectedClose returns the number of days until expected close.
func (o *Opportunity) DaysUntilExpectedClose() int {
	if o.ExpectedCloseDate == nil {
		return 0
	}
	return int(time.Until(*o.ExpectedCloseDate).Hours() / 24)
}

// IsOverdue returns true if the opportunity is past its expected close date.
func (o *Opportunity) IsOverdue() bool {
	if o.ExpectedCloseDate == nil || o.Status.IsClosed() {
		return false
	}
	return time.Now().After(*o.ExpectedCloseDate)
}

// IsStale returns true if there's been no activity for too long.
func (o *Opportunity) IsStale(staleDays int) bool {
	if o.Status.IsClosed() {
		return false
	}
	if o.LastActivityAt == nil {
		return o.DaysInPipeline() > staleDays
	}
	return int(time.Since(*o.LastActivityAt).Hours()/24) > staleDays
}

// AddEvent adds a domain event.
func (o *Opportunity) AddEvent(event DomainEvent) {
	o.events = append(o.events, event)
}

// GetEvents returns all domain events.
func (o *Opportunity) GetEvents() []DomainEvent {
	return o.events
}

// ClearEvents clears all domain events.
func (o *Opportunity) ClearEvents() {
	o.events = make([]DomainEvent, 0)
}
