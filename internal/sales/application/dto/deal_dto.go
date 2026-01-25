package dto

import (
	"time"
)

// ============================================================================
// Deal Request DTOs
// ============================================================================

// CreateDealRequest represents a request to create a new deal.
type CreateDealRequest struct {
	// Basic Information
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description,omitempty" validate:"omitempty,max=5000"`

	// Source - OpportunityID is required for creating a deal
	OpportunityID string `json:"opportunity_id" validate:"required,uuid"`

	// Payment Terms
	PaymentTerm     string `json:"payment_term,omitempty" validate:"omitempty,oneof=net_15 net_30 net_45 net_60 net_90 due_on_receipt prepaid custom"`
	PaymentTermDays int    `json:"payment_term_days,omitempty" validate:"omitempty,min=0,max=365"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Notes        string                 `json:"notes,omitempty" validate:"omitempty,max=5000"`
}

// UpdateDealRequest represents a request to update a deal.
type UpdateDealRequest struct {
	// Basic Information
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=5000"`

	// Payment Terms
	PaymentTerm     *string `json:"payment_term,omitempty" validate:"omitempty,oneof=net_15 net_30 net_45 net_60 net_90 due_on_receipt prepaid custom"`
	PaymentTermDays *int    `json:"payment_term_days,omitempty" validate:"omitempty,min=0,max=365"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Notes        *string                `json:"notes,omitempty" validate:"omitempty,max=5000"`

	// Version for optimistic locking
	Version int `json:"version" validate:"required,min=1"`
}

// AddLineItemRequest represents a request to add a line item to a deal.
type AddLineItemRequest struct {
	ProductID    string  `json:"product_id" validate:"required,uuid"`
	ProductName  string  `json:"product_name" validate:"required,max=200"`
	ProductSKU   string  `json:"product_sku,omitempty" validate:"omitempty,max=100"`
	Description  string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	Quantity     int     `json:"quantity" validate:"required,min=1"`
	UnitPrice    int64   `json:"unit_price" validate:"required,min=0"`
	Currency     string  `json:"currency" validate:"required,len=3"`
	Discount     float64 `json:"discount,omitempty" validate:"omitempty,min=0"`
	DiscountType string  `json:"discount_type,omitempty" validate:"omitempty,oneof=percentage fixed"`
	Tax          float64 `json:"tax,omitempty" validate:"omitempty,min=0"`
	TaxType      string  `json:"tax_type,omitempty" validate:"omitempty,oneof=percentage fixed"`
	Notes        string  `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// UpdateLineItemRequest represents a request to update a line item.
type UpdateLineItemRequest struct {
	ProductName *string  `json:"product_name,omitempty" validate:"omitempty,max=200"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	Quantity    *int     `json:"quantity,omitempty" validate:"omitempty,min=1"`
	UnitPrice   *int64   `json:"unit_price,omitempty" validate:"omitempty,min=0"`
	Discount    *float64 `json:"discount,omitempty" validate:"omitempty,min=0"`
	Tax         *float64 `json:"tax,omitempty" validate:"omitempty,min=0"`
	Notes       *string  `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// GenerateInvoiceRequest represents a request to generate an invoice.
type GenerateInvoiceRequest struct {
	InvoiceNumber   string  `json:"invoice_number" validate:"required,max=100"`
	DueDate         string  `json:"due_date" validate:"required,datetime=2006-01-02"`
	Amount          *int64  `json:"amount,omitempty" validate:"omitempty,min=0"` // Partial invoice amount
	Currency        *string `json:"currency,omitempty" validate:"omitempty,len=3"`
	Notes           *string `json:"notes,omitempty" validate:"omitempty,max=2000"`
	LineItemIDs     []string `json:"line_item_ids,omitempty" validate:"omitempty,dive,uuid"` // Specific line items
	BillingAddress  *AddressDTO `json:"billing_address,omitempty"`
	SendToCustomer  bool    `json:"send_to_customer"`
}

// RecordPaymentRequest represents a request to record a payment.
type RecordPaymentRequest struct {
	InvoiceID       *string `json:"invoice_id,omitempty" validate:"omitempty,uuid"`
	Amount          int64   `json:"amount" validate:"required,min=1"`
	Currency        string  `json:"currency" validate:"required,len=3"`
	PaymentDate     string  `json:"payment_date" validate:"required,datetime=2006-01-02"`
	PaymentMethod   string  `json:"payment_method" validate:"required,oneof=credit_card bank_transfer check cash wire_transfer other"`
	ReferenceNumber *string `json:"reference_number,omitempty" validate:"omitempty,max=100"`
	Notes           *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// UpdateFulfillmentRequest represents a request to update fulfillment status.
type UpdateFulfillmentRequest struct {
	LineItemID  string  `json:"line_item_id" validate:"required,uuid"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	Notes       *string `json:"notes,omitempty" validate:"omitempty,max=500"`
	TrackingNumber *string `json:"tracking_number,omitempty" validate:"omitempty,max=100"`
	CarrierName    *string `json:"carrier_name,omitempty" validate:"omitempty,max=100"`
	FulfilledDate  *string `json:"fulfilled_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// BulkUpdateFulfillmentRequest represents a request to bulk update fulfillment.
type BulkUpdateFulfillmentRequest struct {
	Items []UpdateFulfillmentRequest `json:"items" validate:"required,min=1,max=100,dive"`
}

// CancelDealRequest represents a request to cancel a deal.
type CancelDealRequest struct {
	Reason string  `json:"reason" validate:"required,oneof=customer_request budget_constraints product_issues contract_violation duplicate other"`
	Notes  *string `json:"notes,omitempty" validate:"omitempty,max=2000"`
}

// AssignDealRequest represents a request to assign a deal.
type AssignDealRequest struct {
	OwnerID string  `json:"owner_id" validate:"required,uuid"`
	Notes   *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// DealFilterRequest represents filter options for listing deals.
type DealFilterRequest struct {
	// Status filters
	Statuses []string `json:"statuses,omitempty" validate:"omitempty,dive,oneof=draft pending active completed cancelled refunded"`

	// Relationship filters
	CustomerIDs   []string `json:"customer_ids,omitempty" validate:"omitempty,dive,uuid"`
	OpportunityID *string  `json:"opportunity_id,omitempty" validate:"omitempty,uuid"`
	OwnerIDs      []string `json:"owner_ids,omitempty" validate:"omitempty,dive,uuid"`

	// Value filters
	MinAmount *int64  `json:"min_amount,omitempty" validate:"omitempty,min=0"`
	MaxAmount *int64  `json:"max_amount,omitempty" validate:"omitempty,min=0"`
	Currency  *string `json:"currency,omitempty" validate:"omitempty,len=3"`

	// Payment status
	HasPendingPayments *bool `json:"has_pending_payments,omitempty"`
	FullyPaid          *bool `json:"fully_paid,omitempty"`
	HasOverduePayments *bool `json:"has_overdue_payments,omitempty"`

	// Fulfillment status
	FulfillmentProgress *int  `json:"fulfillment_progress,omitempty" validate:"omitempty,min=0,max=100"`
	FullyFulfilled      *bool `json:"fully_fulfilled,omitempty"`

	// Date filters
	ClosedDateAfter  *string `json:"closed_date_after,omitempty" validate:"omitempty,datetime=2006-01-02"`
	ClosedDateBefore *string `json:"closed_date_before,omitempty" validate:"omitempty,datetime=2006-01-02"`
	SignedDateAfter  *string `json:"signed_date_after,omitempty" validate:"omitempty,datetime=2006-01-02"`
	SignedDateBefore *string `json:"signed_date_before,omitempty" validate:"omitempty,datetime=2006-01-02"`
	CreatedAfter     *string `json:"created_after,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	CreatedBefore    *string `json:"created_before,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`

	// Search
	SearchQuery string  `json:"search_query,omitempty" validate:"omitempty,max=200"`
	DealNumber  *string `json:"deal_number,omitempty" validate:"omitempty,max=100"`

	// Tags
	Tags []string `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`

	// Pagination
	Page     int `json:"page,omitempty" validate:"omitempty,min=1"`
	PageSize int `json:"page_size,omitempty" validate:"omitempty,min=1,max=100"`

	// Sorting
	SortBy    string `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at updated_at closed_date signed_date total_amount deal_number"`
	SortOrder string `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// ============================================================================
// Deal Response DTOs
// ============================================================================

// DealResponse represents a deal in API responses.
type DealResponse struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Code     string `json:"code"`

	// Basic Information
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	// Status
	Status string `json:"status"`

	// Source
	OpportunityID string                    `json:"opportunity_id"`
	Opportunity   *OpportunityBriefResponse `json:"opportunity,omitempty"`
	PipelineID    string                    `json:"pipeline_id"`
	WonReason     string                    `json:"won_reason,omitempty"`

	// Customer
	CustomerID   string            `json:"customer_id"`
	CustomerName string            `json:"customer_name"`
	Customer     *CustomerBriefDTO `json:"customer,omitempty"`

	// Primary Contact
	PrimaryContactID   *string          `json:"primary_contact_id,omitempty"`
	PrimaryContactName string           `json:"primary_contact_name,omitempty"`
	PrimaryContact     *ContactBriefDTO `json:"primary_contact,omitempty"`

	// Assignment
	OwnerID   string        `json:"owner_id"`
	OwnerName string        `json:"owner_name"`
	Owner     *UserBriefDTO `json:"owner,omitempty"`
	TeamID    *string       `json:"team_id,omitempty"`

	// Currency
	Currency string `json:"currency"`

	// Value Summary
	Subtotal          MoneyDTO `json:"subtotal"`
	TotalDiscount     MoneyDTO `json:"total_discount"`
	TotalTax          MoneyDTO `json:"total_tax"`
	TotalAmount       MoneyDTO `json:"total_amount"`
	PaidAmount        MoneyDTO `json:"paid_amount"`
	OutstandingAmount MoneyDTO `json:"outstanding_amount"`

	// Payment Terms
	PaymentTerm     string `json:"payment_term"`
	PaymentTermDays int    `json:"payment_term_days"`

	// Payment Progress
	PaymentProgress float64 `json:"payment_progress"` // percentage
	IsFullyPaid     bool    `json:"is_fully_paid"`

	// Line Items
	LineItems     []*DealLineItemDTO `json:"line_items,omitempty"`
	LineItemCount int                `json:"line_item_count"`

	// Timeline
	Timeline *DealTimelineDTO `json:"timeline,omitempty"`

	// Contract
	ContractURL    string  `json:"contract_url,omitempty"`
	ContractNumber *string `json:"contract_number,omitempty"`
	ContractTerms  *string `json:"contract_terms,omitempty"`

	// Invoicing
	Invoices            []*InvoiceDTO `json:"invoices,omitempty"`
	InvoiceCount        int           `json:"invoice_count"`
	OverdueInvoiceCount int           `json:"overdue_invoice_count"`

	// Payments
	Payments     []*PaymentDTO `json:"payments,omitempty"`
	PaymentCount int           `json:"payment_count"`

	// Fulfillment
	FulfillmentProgress float64 `json:"fulfillment_progress"` // percentage

	// Status Timestamps
	WonAt       *time.Time `json:"won_at,omitempty"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	FulfilledAt *time.Time `json:"fulfilled_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Notes        string                 `json:"notes,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	Version   int       `json:"version"`
}

// DealBriefResponse represents a brief deal summary.
type DealBriefResponse struct {
	ID                  string     `json:"id"`
	Code                string     `json:"code"`
	Name                string     `json:"name"`
	Status              string     `json:"status"`
	CustomerID          string     `json:"customer_id"`
	CustomerName        string     `json:"customer_name"`
	OwnerID             string     `json:"owner_id"`
	OwnerName           string     `json:"owner_name"`
	TotalAmount         MoneyDTO   `json:"total_amount"`
	OutstandingAmount   MoneyDTO   `json:"outstanding_amount"`
	FulfillmentProgress float64    `json:"fulfillment_progress"`
	PaymentProgress     float64    `json:"payment_progress"`
	IsFullyPaid         bool       `json:"is_fully_paid"`
	WonAt               *time.Time `json:"won_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

// DealListResponse represents a paginated list of deals.
type DealListResponse struct {
	Deals      []*DealBriefResponse `json:"deals"`
	Pagination PaginationResponse   `json:"pagination"`
	Summary    *DealSummaryDTO      `json:"summary,omitempty"`
}

// DealSummaryDTO represents a summary of deals in a list.
type DealSummaryDTO struct {
	TotalCount          int64    `json:"total_count"`
	TotalValue          MoneyDTO `json:"total_value"`
	TotalPaid           MoneyDTO `json:"total_paid"`
	TotalPending        MoneyDTO `json:"total_pending"`
	FullyPaidCount      int64    `json:"fully_paid_count"`
	PartiallyPaidCount  int64    `json:"partially_paid_count"`
	UnpaidCount         int64    `json:"unpaid_count"`
	FullyFulfilledCount int64    `json:"fully_fulfilled_count"`
}

// ============================================================================
// Supporting DTOs
// ============================================================================

// DealLineItemRequestDTO represents a line item to add to a deal.
type DealLineItemRequestDTO struct {
	ProductID       string  `json:"product_id" validate:"required,uuid"`
	ProductName     string  `json:"product_name" validate:"required,max=200"`
	ProductSKU      *string `json:"product_sku,omitempty" validate:"omitempty,max=100"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Quantity        int     `json:"quantity" validate:"required,min=1"`
	UnitPrice       int64   `json:"unit_price" validate:"required,min=0"`
	Currency        string  `json:"currency" validate:"required,len=3"`
	DiscountPercent *int    `json:"discount_percent,omitempty" validate:"omitempty,min=0,max=100"`
	DiscountAmount  *int64  `json:"discount_amount,omitempty" validate:"omitempty,min=0"`
	TaxRate         *int    `json:"tax_rate,omitempty" validate:"omitempty,min=0,max=10000"`
	Notes           *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// DealLineItemResponseDTO represents a line item in a deal.
type DealLineItemResponseDTO struct {
	ID              string   `json:"id"`
	ProductID       string   `json:"product_id"`
	ProductName     string   `json:"product_name"`
	ProductSKU      *string  `json:"product_sku,omitempty"`
	Description     *string  `json:"description,omitempty"`
	Quantity        int      `json:"quantity"`
	UnitPrice       MoneyDTO `json:"unit_price"`
	DiscountPercent int      `json:"discount_percent"`
	DiscountAmount  MoneyDTO `json:"discount_amount"`
	TaxRate         int      `json:"tax_rate"` // basis points
	TaxAmount       MoneyDTO `json:"tax_amount"`
	TotalPrice      MoneyDTO `json:"total_price"`
	FulfilledQty    int      `json:"fulfilled_qty"`
	PendingQty      int      `json:"pending_qty"`
	Notes           *string  `json:"notes,omitempty"`
}

// InvoiceResponseDTO represents an invoice.
type InvoiceResponseDTO struct {
	ID            string     `json:"id"`
	InvoiceNumber string     `json:"invoice_number"`
	Amount        MoneyDTO   `json:"amount"`
	Status        string     `json:"status"` // draft, sent, paid, overdue, cancelled
	IssuedDate    time.Time  `json:"issued_date"`
	DueDate       time.Time  `json:"due_date"`
	PaidDate      *time.Time `json:"paid_date,omitempty"`
	PaidAmount    MoneyDTO   `json:"paid_amount"`
	Notes         *string    `json:"notes,omitempty"`
}

// PaymentResponseDTO represents a payment.
type PaymentResponseDTO struct {
	ID              string    `json:"id"`
	InvoiceID       *string   `json:"invoice_id,omitempty"`
	Amount          MoneyDTO  `json:"amount"`
	PaymentDate     time.Time `json:"payment_date"`
	PaymentMethod   string    `json:"payment_method"`
	ReferenceNumber *string   `json:"reference_number,omitempty"`
	Status          string    `json:"status"` // pending, completed, failed, refunded
	Notes           *string   `json:"notes,omitempty"`
	RecordedAt      time.Time `json:"recorded_at"`
	RecordedBy      string    `json:"recorded_by"`
}

// FulfillmentResponseDTO represents a fulfillment record.
type FulfillmentResponseDTO struct {
	ID             string    `json:"id"`
	LineItemID     string    `json:"line_item_id"`
	ProductName    string    `json:"product_name"`
	Quantity       int       `json:"quantity"`
	FulfilledDate  time.Time `json:"fulfilled_date"`
	TrackingNumber *string   `json:"tracking_number,omitempty"`
	CarrierName    *string   `json:"carrier_name,omitempty"`
	Notes          *string   `json:"notes,omitempty"`
	FulfilledBy    string    `json:"fulfilled_by"`
}

// ============================================================================
// Domain-Aligned DTOs (for use case responses)
// ============================================================================

// DealTimelineDTO represents the timeline of a deal.
type DealTimelineDTO struct {
	QuoteDate       *time.Time `json:"quote_date,omitempty"`
	ContractDate    *time.Time `json:"contract_date,omitempty"`
	StartDate       *time.Time `json:"start_date,omitempty"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	RenewalDate     *time.Time `json:"renewal_date,omitempty"`
	FirstPaymentDue *time.Time `json:"first_payment_due,omitempty"`
}

// DealLineItemDTO represents a line item in a deal (domain-aligned).
type DealLineItemDTO struct {
	ID                string     `json:"id"`
	ProductID         string     `json:"product_id"`
	ProductName       string     `json:"product_name"`
	ProductSKU        string     `json:"product_sku,omitempty"`
	Description       string     `json:"description,omitempty"`
	Quantity          int        `json:"quantity"`
	UnitPrice         MoneyDTO   `json:"unit_price"`
	Discount          float64    `json:"discount"`
	DiscountType      string     `json:"discount_type,omitempty"` // percentage, fixed
	Tax               float64    `json:"tax"`
	TaxType           string     `json:"tax_type,omitempty"` // percentage, fixed
	Subtotal          MoneyDTO   `json:"subtotal"`
	TaxAmount         MoneyDTO   `json:"tax_amount"`
	Total             MoneyDTO   `json:"total"`
	FulfilledQty      int        `json:"fulfilled_qty"`
	RemainingQuantity int        `json:"remaining_quantity"`
	IsFulfilled       bool       `json:"is_fulfilled"`
	DeliveryDate      *time.Time `json:"delivery_date,omitempty"`
	Notes             string     `json:"notes,omitempty"`
}

// InvoiceDTO represents an invoice in a deal (domain-aligned).
type InvoiceDTO struct {
	ID                string     `json:"id"`
	InvoiceNumber     string     `json:"invoice_number"`
	Amount            MoneyDTO   `json:"amount"`
	DueDate           time.Time  `json:"due_date"`
	Status            string     `json:"status"` // draft, sent, paid, overdue, cancelled
	SentAt            *time.Time `json:"sent_at,omitempty"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	PaidAmount        MoneyDTO   `json:"paid_amount"`
	OutstandingAmount MoneyDTO   `json:"outstanding_amount"`
	IsOverdue         bool       `json:"is_overdue"`
	Notes             string     `json:"notes,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// PaymentDTO represents a payment in a deal (domain-aligned).
type PaymentDTO struct {
	ID            string    `json:"id"`
	InvoiceID     *string   `json:"invoice_id,omitempty"`
	Amount        MoneyDTO  `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Reference     string    `json:"reference,omitempty"`
	ReceivedAt    time.Time `json:"received_at"`
	ReceivedBy    string    `json:"received_by"`
	Notes         string    `json:"notes,omitempty"`
}

// CreateInvoiceRequest represents a request to create an invoice.
type CreateInvoiceRequest struct {
	InvoiceNumber string `json:"invoice_number" validate:"required,max=100"`
	Amount        int64  `json:"amount" validate:"required,min=1"`
	DueDate       string `json:"due_date" validate:"required,datetime=2006-01-02"`
	Notes         string `json:"notes,omitempty" validate:"omitempty,max=2000"`
}

// DealStatisticsResponse represents deal statistics.
type DealStatisticsResponse struct {
	TotalDeals           int64            `json:"total_deals"`
	ByStatus             map[string]int64 `json:"by_status"`
	TotalRevenue         MoneyDTO         `json:"total_revenue"`
	TotalCollected       MoneyDTO         `json:"total_collected"`
	TotalOutstanding     MoneyDTO         `json:"total_outstanding"`
	AverageDealSize      MoneyDTO         `json:"average_deal_size"`
	FullyPaidDeals       int64            `json:"fully_paid_deals"`
	PartiallyPaidDeals   int64            `json:"partially_paid_deals"`
	UnpaidDeals          int64            `json:"unpaid_deals"`
	FullyFulfilledDeals  int64            `json:"fully_fulfilled_deals"`
	DealsThisMonth       int64            `json:"deals_this_month"`
	RevenueThisMonth     MoneyDTO         `json:"revenue_this_month"`
	CollectedThisMonth   MoneyDTO         `json:"collected_this_month"`
}

// RevenueReportResponse represents a revenue report.
type RevenueReportResponse struct {
	Period           string                 `json:"period"` // daily, weekly, monthly, quarterly, yearly
	StartDate        time.Time              `json:"start_date"`
	EndDate          time.Time              `json:"end_date"`
	TotalRevenue     MoneyDTO               `json:"total_revenue"`
	TotalCollected   MoneyDTO               `json:"total_collected"`
	TotalOutstanding MoneyDTO               `json:"total_outstanding"`
	DealCount        int64                  `json:"deal_count"`
	AverageDealSize  MoneyDTO               `json:"average_deal_size"`
	ByCustomer       []*CustomerRevenueDTO  `json:"by_customer,omitempty"`
	ByProduct        []*ProductRevenueDTO   `json:"by_product,omitempty"`
	ByOwner          []*OwnerRevenueDTO     `json:"by_owner,omitempty"`
	Trend            []*RevenueTrendDTO     `json:"trend,omitempty"`
}

// CustomerRevenueDTO represents revenue by customer.
type CustomerRevenueDTO struct {
	CustomerID   string   `json:"customer_id"`
	CustomerName string   `json:"customer_name"`
	Revenue      MoneyDTO `json:"revenue"`
	DealCount    int64    `json:"deal_count"`
	Percentage   float64  `json:"percentage"`
}

// ProductRevenueDTO represents revenue by product.
type ProductRevenueDTO struct {
	ProductID   string   `json:"product_id"`
	ProductName string   `json:"product_name"`
	Revenue     MoneyDTO `json:"revenue"`
	Quantity    int64    `json:"quantity"`
	Percentage  float64  `json:"percentage"`
}

// OwnerRevenueDTO represents revenue by owner.
type OwnerRevenueDTO struct {
	OwnerID    string   `json:"owner_id"`
	OwnerName  string   `json:"owner_name"`
	Revenue    MoneyDTO `json:"revenue"`
	DealCount  int64    `json:"deal_count"`
	Percentage float64  `json:"percentage"`
}

// RevenueTrendDTO represents a revenue trend data point.
type RevenueTrendDTO struct {
	Date       string   `json:"date"`
	Revenue    MoneyDTO `json:"revenue"`
	Collected  MoneyDTO `json:"collected"`
	DealCount  int64    `json:"deal_count"`
}
