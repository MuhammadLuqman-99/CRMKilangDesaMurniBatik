// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Deal errors
var (
	ErrDealNotFound           = errors.New("deal not found")
	ErrDealAlreadyExists      = errors.New("deal already exists")
	ErrDealAlreadyClosed      = errors.New("deal is already closed")
	ErrDealAlreadyFulfilled   = errors.New("deal is already fulfilled")
	ErrDealCannotBeCancelled  = errors.New("deal cannot be cancelled")
	ErrInvalidDealStatus      = errors.New("invalid deal status")
	ErrDealVersionMismatch    = errors.New("deal version mismatch")
	ErrInvalidPaymentTerm     = errors.New("invalid payment term")
	ErrInvoiceAlreadyExists   = errors.New("invoice already exists")
	ErrPaymentExceedsBalance  = errors.New("payment exceeds outstanding balance")
)

// DealStatus represents the status of a deal.
type DealStatus string

const (
	DealStatusDraft       DealStatus = "draft"
	DealStatusPending     DealStatus = "pending"
	DealStatusActive      DealStatus = "active"
	DealStatusOnHold      DealStatus = "on_hold"
	DealStatusFulfilled   DealStatus = "fulfilled"
	DealStatusCancelled   DealStatus = "cancelled"
)

// ValidDealStatuses returns all valid deal statuses.
func ValidDealStatuses() []DealStatus {
	return []DealStatus{
		DealStatusDraft,
		DealStatusPending,
		DealStatusActive,
		DealStatusOnHold,
		DealStatusFulfilled,
		DealStatusCancelled,
	}
}

// IsValid checks if the deal status is valid.
func (s DealStatus) IsValid() bool {
	for _, valid := range ValidDealStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// IsClosed returns true if the status is closed.
func (s DealStatus) IsClosed() bool {
	return s == DealStatusFulfilled || s == DealStatusCancelled
}

// PaymentTerm represents payment terms for a deal.
type PaymentTerm string

const (
	PaymentTermImmediate PaymentTerm = "immediate"
	PaymentTermNet7      PaymentTerm = "net_7"
	PaymentTermNet15     PaymentTerm = "net_15"
	PaymentTermNet30     PaymentTerm = "net_30"
	PaymentTermNet45     PaymentTerm = "net_45"
	PaymentTermNet60     PaymentTerm = "net_60"
	PaymentTermNet90     PaymentTerm = "net_90"
	PaymentTermCustom    PaymentTerm = "custom"
)

// DaysUntilDue returns the number of days until payment is due.
func (t PaymentTerm) DaysUntilDue() int {
	switch t {
	case PaymentTermImmediate:
		return 0
	case PaymentTermNet7:
		return 7
	case PaymentTermNet15:
		return 15
	case PaymentTermNet30:
		return 30
	case PaymentTermNet45:
		return 45
	case PaymentTermNet60:
		return 60
	case PaymentTermNet90:
		return 90
	default:
		return 30
	}
}

// DealLineItem represents a line item in a deal.
type DealLineItem struct {
	ID            uuid.UUID `json:"id" bson:"id"`
	ProductID     uuid.UUID `json:"product_id" bson:"product_id"`
	ProductName   string    `json:"product_name" bson:"product_name"`
	ProductSKU    string    `json:"product_sku,omitempty" bson:"product_sku,omitempty"`
	Description   string    `json:"description,omitempty" bson:"description,omitempty"`
	Quantity      int       `json:"quantity" bson:"quantity"`
	UnitPrice     Money     `json:"unit_price" bson:"unit_price"`
	Discount      float64   `json:"discount" bson:"discount"`
	DiscountType  string    `json:"discount_type" bson:"discount_type"` // percentage, fixed
	Tax           float64   `json:"tax" bson:"tax"`
	TaxType       string    `json:"tax_type" bson:"tax_type"` // percentage, fixed
	Subtotal      Money     `json:"subtotal" bson:"subtotal"`
	TaxAmount     Money     `json:"tax_amount" bson:"tax_amount"`
	Total         Money     `json:"total" bson:"total"`
	FulfilledQty  int       `json:"fulfilled_qty" bson:"fulfilled_qty"`
	DeliveryDate  *time.Time `json:"delivery_date,omitempty" bson:"delivery_date,omitempty"`
	Notes         string    `json:"notes,omitempty" bson:"notes,omitempty"`
}

// Calculate calculates the line item totals.
func (li *DealLineItem) Calculate() {
	basePrice := li.UnitPrice.Multiply(float64(li.Quantity))

	// Calculate discount
	var discountAmount Money
	if li.DiscountType == "percentage" {
		discountAmount = basePrice.Multiply(li.Discount / 100)
	} else {
		discountAmount, _ = NewMoneyFromFloat(li.Discount, li.UnitPrice.Currency)
	}

	li.Subtotal, _ = basePrice.Subtract(discountAmount)

	// Calculate tax
	if li.TaxType == "percentage" {
		li.TaxAmount = li.Subtotal.Multiply(li.Tax / 100)
	} else {
		li.TaxAmount, _ = NewMoneyFromFloat(li.Tax, li.UnitPrice.Currency)
	}

	li.Total, _ = li.Subtotal.Add(li.TaxAmount)
}

// IsFulfilled returns true if the line item is fully fulfilled.
func (li *DealLineItem) IsFulfilled() bool {
	return li.FulfilledQty >= li.Quantity
}

// RemainingQuantity returns the quantity remaining to be fulfilled.
func (li *DealLineItem) RemainingQuantity() int {
	return li.Quantity - li.FulfilledQty
}

// Invoice represents an invoice for a deal.
type Invoice struct {
	ID            uuid.UUID   `json:"id" bson:"id"`
	InvoiceNumber string      `json:"invoice_number" bson:"invoice_number"`
	Amount        Money       `json:"amount" bson:"amount"`
	DueDate       time.Time   `json:"due_date" bson:"due_date"`
	Status        string      `json:"status" bson:"status"` // draft, sent, paid, overdue, cancelled
	SentAt        *time.Time  `json:"sent_at,omitempty" bson:"sent_at,omitempty"`
	PaidAt        *time.Time  `json:"paid_at,omitempty" bson:"paid_at,omitempty"`
	PaidAmount    Money       `json:"paid_amount" bson:"paid_amount"`
	Notes         string      `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt     time.Time   `json:"created_at" bson:"created_at"`
}

// IsPaid returns true if the invoice is paid.
func (i *Invoice) IsPaid() bool {
	return i.Status == "paid"
}

// IsOverdue returns true if the invoice is overdue.
func (i *Invoice) IsOverdue() bool {
	return i.Status != "paid" && time.Now().After(i.DueDate)
}

// OutstandingAmount returns the amount still owed.
func (i *Invoice) OutstandingAmount() Money {
	result, _ := i.Amount.Subtract(i.PaidAmount)
	return result
}

// Payment represents a payment received for a deal.
type Payment struct {
	ID            uuid.UUID  `json:"id" bson:"id"`
	InvoiceID     *uuid.UUID `json:"invoice_id,omitempty" bson:"invoice_id,omitempty"`
	Amount        Money      `json:"amount" bson:"amount"`
	PaymentMethod string     `json:"payment_method" bson:"payment_method"`
	Reference     string     `json:"reference,omitempty" bson:"reference,omitempty"`
	ReceivedAt    time.Time  `json:"received_at" bson:"received_at"`
	ReceivedBy    uuid.UUID  `json:"received_by" bson:"received_by"`
	Notes         string     `json:"notes,omitempty" bson:"notes,omitempty"`
}

// DealTimeline represents important dates in the deal lifecycle.
type DealTimeline struct {
	QuoteDate       *time.Time `json:"quote_date,omitempty" bson:"quote_date,omitempty"`
	ContractDate    *time.Time `json:"contract_date,omitempty" bson:"contract_date,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty" bson:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty" bson:"end_date,omitempty"`
	RenewalDate     *time.Time `json:"renewal_date,omitempty" bson:"renewal_date,omitempty"`
	FirstPaymentDue *time.Time `json:"first_payment_due,omitempty" bson:"first_payment_due,omitempty"`
}

// Deal represents a closed deal (won opportunity).
type Deal struct {
	ID                uuid.UUID              `json:"id" bson:"_id"`
	TenantID          uuid.UUID              `json:"tenant_id" bson:"tenant_id"`
	Code              string                 `json:"code" bson:"code"` // e.g., "DL-2024-001"
	Name              string                 `json:"name" bson:"name"`
	Description       string                 `json:"description,omitempty" bson:"description,omitempty"`
	Status            DealStatus             `json:"status" bson:"status"`

	// Source
	OpportunityID     uuid.UUID              `json:"opportunity_id" bson:"opportunity_id"`
	PipelineID        uuid.UUID              `json:"pipeline_id" bson:"pipeline_id"`
	WonReason         string                 `json:"won_reason" bson:"won_reason"`

	// Customer
	CustomerID        uuid.UUID              `json:"customer_id" bson:"customer_id"`
	CustomerName      string                 `json:"customer_name" bson:"customer_name"`
	PrimaryContactID  *uuid.UUID             `json:"primary_contact_id,omitempty" bson:"primary_contact_id,omitempty"`
	PrimaryContactName string                `json:"primary_contact_name,omitempty" bson:"primary_contact_name,omitempty"`

	// Value
	Currency          string                 `json:"currency" bson:"currency"`
	Subtotal          Money                  `json:"subtotal" bson:"subtotal"`
	TotalDiscount     Money                  `json:"total_discount" bson:"total_discount"`
	TotalTax          Money                  `json:"total_tax" bson:"total_tax"`
	TotalAmount       Money                  `json:"total_amount" bson:"total_amount"`
	PaidAmount        Money                  `json:"paid_amount" bson:"paid_amount"`
	OutstandingAmount Money                  `json:"outstanding_amount" bson:"outstanding_amount"`

	// Line Items
	LineItems         []DealLineItem         `json:"line_items" bson:"line_items"`

	// Payment
	PaymentTerm       PaymentTerm            `json:"payment_term" bson:"payment_term"`
	PaymentTermDays   int                    `json:"payment_term_days" bson:"payment_term_days"`
	Invoices          []Invoice              `json:"invoices" bson:"invoices"`
	Payments          []Payment              `json:"payments" bson:"payments"`

	// Timeline
	Timeline          DealTimeline           `json:"timeline" bson:"timeline"`
	WonAt             time.Time              `json:"won_at" bson:"won_at"`
	ActivatedAt       *time.Time             `json:"activated_at,omitempty" bson:"activated_at,omitempty"`
	FulfilledAt       *time.Time             `json:"fulfilled_at,omitempty" bson:"fulfilled_at,omitempty"`
	CancelledAt       *time.Time             `json:"cancelled_at,omitempty" bson:"cancelled_at,omitempty"`

	// Ownership
	OwnerID           uuid.UUID              `json:"owner_id" bson:"owner_id"`
	OwnerName         string                 `json:"owner_name" bson:"owner_name"`
	TeamID            *uuid.UUID             `json:"team_id,omitempty" bson:"team_id,omitempty"`

	// Metadata
	Tags              []string               `json:"tags,omitempty" bson:"tags,omitempty"`
	CustomFields      map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
	Notes             string                 `json:"notes,omitempty" bson:"notes,omitempty"`
	ContractURL       string                 `json:"contract_url,omitempty" bson:"contract_url,omitempty"`

	// Timestamps
	CreatedBy         uuid.UUID              `json:"created_by" bson:"created_by"`
	CreatedAt         time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" bson:"updated_at"`
	DeletedAt         *time.Time             `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Version           int                    `json:"version" bson:"version"`

	// Domain events
	events []DomainEvent `json:"-" bson:"-"`
}

// NewDealFromOpportunity creates a new deal from a won opportunity.
func NewDealFromOpportunity(opportunity *Opportunity, createdBy uuid.UUID) (*Deal, error) {
	if !opportunity.IsWon() {
		return nil, errors.New("opportunity must be won to create a deal")
	}

	now := time.Now().UTC()

	deal := &Deal{
		ID:              uuid.New(),
		TenantID:        opportunity.TenantID,
		Name:            opportunity.Name,
		Description:     opportunity.Description,
		Status:          DealStatusDraft,
		OpportunityID:   opportunity.ID,
		PipelineID:      opportunity.PipelineID,
		WonReason:       opportunity.CloseInfo.Reason,
		CustomerID:      opportunity.CustomerID,
		CustomerName:    opportunity.CustomerName,
		Currency:        opportunity.Amount.Currency,
		Subtotal:        opportunity.Amount,
		TotalAmount:     opportunity.Amount,
		OutstandingAmount: opportunity.Amount,
		LineItems:       make([]DealLineItem, 0),
		PaymentTerm:     PaymentTermNet30,
		PaymentTermDays: 30,
		Invoices:        make([]Invoice, 0),
		Payments:        make([]Payment, 0),
		Timeline: DealTimeline{
			ContractDate: &now,
		},
		WonAt:      *opportunity.ActualCloseDate,
		OwnerID:    opportunity.OwnerID,
		OwnerName:  opportunity.OwnerName,
		TeamID:     opportunity.TeamID,
		Tags:       opportunity.Tags,
		CreatedBy:  createdBy,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
		events:     make([]DomainEvent, 0),
	}

	// Set primary contact
	primaryContact := opportunity.GetPrimaryContact()
	if primaryContact != nil {
		deal.PrimaryContactID = &primaryContact.ContactID
		deal.PrimaryContactName = primaryContact.Name
	}

	// Copy products to line items
	for _, p := range opportunity.Products {
		lineItem := DealLineItem{
			ID:          uuid.New(),
			ProductID:   p.ProductID,
			ProductName: p.ProductName,
			ProductSKU:  p.SKU,
			Quantity:    p.Quantity,
			UnitPrice:   p.UnitPrice,
			Discount:    p.Discount,
			DiscountType: "percentage",
			Tax:         p.Tax,
			TaxType:     "percentage",
		}
		lineItem.Calculate()
		deal.LineItems = append(deal.LineItems, lineItem)
	}

	// Initialize zero money values
	deal.TotalDiscount, _ = Zero(deal.Currency)
	deal.TotalTax, _ = Zero(deal.Currency)
	deal.PaidAmount, _ = Zero(deal.Currency)

	deal.recalculateTotals()

	deal.AddEvent(NewDealCreatedEvent(deal))
	return deal, nil
}

// Update updates deal details.
func (d *Deal) Update(name, description string, paymentTerm PaymentTerm, paymentTermDays int) {
	if name != "" {
		d.Name = name
	}
	d.Description = description
	d.PaymentTerm = paymentTerm
	if paymentTerm == PaymentTermCustom {
		d.PaymentTermDays = paymentTermDays
	} else {
		d.PaymentTermDays = paymentTerm.DaysUntilDue()
	}
	d.UpdatedAt = time.Now().UTC()

	d.AddEvent(NewDealUpdatedEvent(d))
}

// Activate activates the deal.
func (d *Deal) Activate() error {
	if d.Status == DealStatusActive {
		return nil
	}
	if d.Status.IsClosed() {
		return ErrDealAlreadyClosed
	}

	now := time.Now().UTC()
	d.Status = DealStatusActive
	d.ActivatedAt = &now
	d.UpdatedAt = now

	d.AddEvent(NewDealActivatedEvent(d))
	return nil
}

// PutOnHold puts the deal on hold.
func (d *Deal) PutOnHold(reason string) error {
	if d.Status.IsClosed() {
		return ErrDealAlreadyClosed
	}

	d.Status = DealStatusOnHold
	d.Notes = reason
	d.UpdatedAt = time.Now().UTC()

	return nil
}

// Resume resumes a deal from hold.
func (d *Deal) Resume() error {
	if d.Status != DealStatusOnHold {
		return errors.New("deal is not on hold")
	}

	d.Status = DealStatusActive
	d.UpdatedAt = time.Now().UTC()

	return nil
}

// Fulfill marks the deal as fulfilled.
func (d *Deal) Fulfill() error {
	if d.Status == DealStatusFulfilled {
		return ErrDealAlreadyFulfilled
	}
	if d.Status == DealStatusCancelled {
		return ErrDealAlreadyClosed
	}

	// Check if all line items are fulfilled
	for _, li := range d.LineItems {
		if !li.IsFulfilled() {
			return errors.New("not all line items are fulfilled")
		}
	}

	now := time.Now().UTC()
	d.Status = DealStatusFulfilled
	d.FulfilledAt = &now
	d.UpdatedAt = now

	d.AddEvent(NewDealFulfilledEvent(d))
	return nil
}

// Cancel cancels the deal.
func (d *Deal) Cancel(reason string) error {
	if d.Status.IsClosed() {
		return ErrDealAlreadyClosed
	}

	// Can't cancel if there are paid invoices
	for _, inv := range d.Invoices {
		if inv.IsPaid() {
			return ErrDealCannotBeCancelled
		}
	}

	now := time.Now().UTC()
	d.Status = DealStatusCancelled
	d.CancelledAt = &now
	d.Notes = reason
	d.UpdatedAt = now

	d.AddEvent(NewDealCancelledEvent(d, reason))
	return nil
}

// AddLineItem adds a line item to the deal.
func (d *Deal) AddLineItem(item DealLineItem) error {
	if d.Status.IsClosed() {
		return ErrDealAlreadyClosed
	}

	item.ID = uuid.New()
	item.Calculate()
	d.LineItems = append(d.LineItems, item)
	d.recalculateTotals()
	d.UpdatedAt = time.Now().UTC()

	return nil
}

// UpdateLineItem updates a line item in the deal.
func (d *Deal) UpdateLineItem(itemID uuid.UUID, quantity int, unitPrice Money, discount, tax float64) error {
	if d.Status.IsClosed() {
		return ErrDealAlreadyClosed
	}

	for i := range d.LineItems {
		if d.LineItems[i].ID == itemID {
			d.LineItems[i].Quantity = quantity
			d.LineItems[i].UnitPrice = unitPrice
			d.LineItems[i].Discount = discount
			d.LineItems[i].Tax = tax
			d.LineItems[i].Calculate()
			d.recalculateTotals()
			d.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return errors.New("line item not found")
}

// RemoveLineItem removes a line item from the deal.
func (d *Deal) RemoveLineItem(itemID uuid.UUID) error {
	if d.Status.IsClosed() {
		return ErrDealAlreadyClosed
	}

	for i, li := range d.LineItems {
		if li.ID == itemID {
			d.LineItems = append(d.LineItems[:i], d.LineItems[i+1:]...)
			d.recalculateTotals()
			d.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return errors.New("line item not found")
}

// FulfillLineItem marks quantity as fulfilled for a line item.
func (d *Deal) FulfillLineItem(itemID uuid.UUID, quantity int) error {
	for i := range d.LineItems {
		if d.LineItems[i].ID == itemID {
			newFulfilled := d.LineItems[i].FulfilledQty + quantity
			if newFulfilled > d.LineItems[i].Quantity {
				return errors.New("fulfilled quantity exceeds ordered quantity")
			}
			d.LineItems[i].FulfilledQty = newFulfilled
			d.UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return errors.New("line item not found")
}

// recalculateTotals recalculates all totals.
func (d *Deal) recalculateTotals() {
	subtotal, _ := Zero(d.Currency)
	totalTax, _ := Zero(d.Currency)
	totalDiscount, _ := Zero(d.Currency)
	total, _ := Zero(d.Currency)

	for _, li := range d.LineItems {
		subtotal, _ = subtotal.Add(li.Subtotal)
		totalTax, _ = totalTax.Add(li.TaxAmount)
		total, _ = total.Add(li.Total)

		// Calculate discount amount
		basePrice := li.UnitPrice.Multiply(float64(li.Quantity))
		discountAmount, _ := basePrice.Subtract(li.Subtotal)
		totalDiscount, _ = totalDiscount.Add(discountAmount)
	}

	d.Subtotal = subtotal
	d.TotalTax = totalTax
	d.TotalDiscount = totalDiscount
	d.TotalAmount = total
	d.OutstandingAmount, _ = d.TotalAmount.Subtract(d.PaidAmount)
}

// CreateInvoice creates a new invoice for the deal.
func (d *Deal) CreateInvoice(invoiceNumber string, amount Money, dueDate time.Time) (*Invoice, error) {
	// Check if amount exceeds outstanding
	gt, err := amount.GreaterThan(d.OutstandingAmount)
	if err != nil {
		return nil, err
	}
	if gt {
		return nil, errors.New("invoice amount exceeds outstanding balance")
	}

	// Check for duplicate invoice number
	for _, inv := range d.Invoices {
		if inv.InvoiceNumber == invoiceNumber {
			return nil, ErrInvoiceAlreadyExists
		}
	}

	paidAmount, _ := Zero(d.Currency)
	invoice := Invoice{
		ID:            uuid.New(),
		InvoiceNumber: invoiceNumber,
		Amount:        amount,
		DueDate:       dueDate,
		Status:        "draft",
		PaidAmount:    paidAmount,
		CreatedAt:     time.Now().UTC(),
	}

	d.Invoices = append(d.Invoices, invoice)
	d.UpdatedAt = time.Now().UTC()

	d.AddEvent(NewDealInvoiceCreatedEvent(d, &invoice))
	return &invoice, nil
}

// SendInvoice marks an invoice as sent.
func (d *Deal) SendInvoice(invoiceID uuid.UUID) error {
	for i := range d.Invoices {
		if d.Invoices[i].ID == invoiceID {
			now := time.Now().UTC()
			d.Invoices[i].Status = "sent"
			d.Invoices[i].SentAt = &now
			d.UpdatedAt = now
			return nil
		}
	}
	return errors.New("invoice not found")
}

// RecordPayment records a payment for the deal.
func (d *Deal) RecordPayment(payment Payment) error {
	// Check if payment exceeds outstanding
	gt, err := payment.Amount.GreaterThan(d.OutstandingAmount)
	if err != nil {
		return err
	}
	if gt {
		return ErrPaymentExceedsBalance
	}

	payment.ID = uuid.New()
	d.Payments = append(d.Payments, payment)

	// Update paid amount
	d.PaidAmount, _ = d.PaidAmount.Add(payment.Amount)
	d.OutstandingAmount, _ = d.TotalAmount.Subtract(d.PaidAmount)

	// Update invoice if specified
	if payment.InvoiceID != nil {
		for i := range d.Invoices {
			if d.Invoices[i].ID == *payment.InvoiceID {
				d.Invoices[i].PaidAmount, _ = d.Invoices[i].PaidAmount.Add(payment.Amount)
				if d.Invoices[i].PaidAmount.Amount >= d.Invoices[i].Amount.Amount {
					now := time.Now().UTC()
					d.Invoices[i].Status = "paid"
					d.Invoices[i].PaidAt = &now
				}
				break
			}
		}
	}

	d.UpdatedAt = time.Now().UTC()

	d.AddEvent(NewDealPaymentReceivedEvent(d, &payment))
	return nil
}

// IsFullyPaid returns true if the deal is fully paid.
func (d *Deal) IsFullyPaid() bool {
	return d.OutstandingAmount.Amount <= 0
}

// GetOverdueInvoices returns all overdue invoices.
func (d *Deal) GetOverdueInvoices() []Invoice {
	var overdue []Invoice
	for _, inv := range d.Invoices {
		if inv.IsOverdue() {
			overdue = append(overdue, inv)
		}
	}
	return overdue
}

// SetTimeline sets timeline dates.
func (d *Deal) SetTimeline(timeline DealTimeline) {
	d.Timeline = timeline
	d.UpdatedAt = time.Now().UTC()
}

// SetContractURL sets the contract document URL.
func (d *Deal) SetContractURL(url string) {
	d.ContractURL = url
	d.UpdatedAt = time.Now().UTC()
}

// AddTag adds a tag.
func (d *Deal) AddTag(tag string) {
	for _, t := range d.Tags {
		if t == tag {
			return
		}
	}
	d.Tags = append(d.Tags, tag)
	d.UpdatedAt = time.Now().UTC()
}

// RemoveTag removes a tag.
func (d *Deal) RemoveTag(tag string) {
	for i, t := range d.Tags {
		if t == tag {
			d.Tags = append(d.Tags[:i], d.Tags[i+1:]...)
			d.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// SetCustomField sets a custom field value.
func (d *Deal) SetCustomField(key string, value interface{}) {
	if d.CustomFields == nil {
		d.CustomFields = make(map[string]interface{})
	}
	d.CustomFields[key] = value
	d.UpdatedAt = time.Now().UTC()
}

// Delete soft deletes the deal.
func (d *Deal) Delete() error {
	if d.Status == DealStatusActive || d.Status == DealStatusFulfilled {
		return errors.New("cannot delete active or fulfilled deal")
	}

	now := time.Now().UTC()
	d.DeletedAt = &now
	d.UpdatedAt = now

	d.AddEvent(NewDealDeletedEvent(d))
	return nil
}

// Restore restores a soft-deleted deal.
func (d *Deal) Restore() {
	d.DeletedAt = nil
	d.UpdatedAt = time.Now().UTC()
}

// IsDeleted returns true if the deal is deleted.
func (d *Deal) IsDeleted() bool {
	return d.DeletedAt != nil
}

// FulfillmentProgress returns the fulfillment progress as a percentage.
func (d *Deal) FulfillmentProgress() float64 {
	if len(d.LineItems) == 0 {
		return 0
	}

	totalQty := 0
	fulfilledQty := 0
	for _, li := range d.LineItems {
		totalQty += li.Quantity
		fulfilledQty += li.FulfilledQty
	}

	if totalQty == 0 {
		return 0
	}

	return float64(fulfilledQty) / float64(totalQty) * 100
}

// PaymentProgress returns the payment progress as a percentage.
func (d *Deal) PaymentProgress() float64 {
	if d.TotalAmount.Amount == 0 {
		return 0
	}

	return float64(d.PaidAmount.Amount) / float64(d.TotalAmount.Amount) * 100
}

// AddEvent adds a domain event.
func (d *Deal) AddEvent(event DomainEvent) {
	d.events = append(d.events, event)
}

// GetEvents returns all domain events.
func (d *Deal) GetEvents() []DomainEvent {
	return d.events
}

// ClearEvents clears all domain events.
func (d *Deal) ClearEvents() {
	d.events = make([]DomainEvent, 0)
}
