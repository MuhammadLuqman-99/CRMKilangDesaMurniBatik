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
	Name        string  `json:"name" validate:"required,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=5000"`

	// Source
	OpportunityID *string `json:"opportunity_id,omitempty" validate:"omitempty,uuid"`

	// Customer
	CustomerID string `json:"customer_id" validate:"required,uuid"`

	// Value
	TotalAmount int64  `json:"total_amount" validate:"required,min=0"`
	Currency    string `json:"currency" validate:"required,len=3"`

	// Line Items
	LineItems []*DealLineItemRequestDTO `json:"line_items,omitempty" validate:"omitempty,max=100,dive"`

	// Dates
	ClosedDate *string `json:"closed_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	SignedDate *string `json:"signed_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	StartDate  *string `json:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	EndDate    *string `json:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02"`

	// Assignment
	OwnerID *string `json:"owner_id,omitempty" validate:"omitempty,uuid"`

	// Contract Details
	ContractNumber *string `json:"contract_number,omitempty" validate:"omitempty,max=100"`
	ContractTerms  *string `json:"contract_terms,omitempty" validate:"omitempty,max=5000"`
	PaymentTerms   *string `json:"payment_terms,omitempty" validate:"omitempty,oneof=net_15 net_30 net_45 net_60 net_90 due_on_receipt prepaid custom"`
	PaymentMethod  *string `json:"payment_method,omitempty" validate:"omitempty,oneof=credit_card bank_transfer check cash wire_transfer other"`

	// Billing
	BillingContactID *string     `json:"billing_contact_id,omitempty" validate:"omitempty,uuid"`
	BillingAddress   *AddressDTO `json:"billing_address,omitempty"`

	// Shipping
	ShippingContactID *string     `json:"shipping_contact_id,omitempty" validate:"omitempty,uuid"`
	ShippingAddress   *AddressDTO `json:"shipping_address,omitempty"`
	ShippingMethod    *string     `json:"shipping_method,omitempty" validate:"omitempty,max=100"`
	ShippingCost      *int64      `json:"shipping_cost,omitempty" validate:"omitempty,min=0"`

	// Tax
	TaxRate   *int   `json:"tax_rate,omitempty" validate:"omitempty,min=0,max=10000"` // basis points
	TaxAmount *int64 `json:"tax_amount,omitempty" validate:"omitempty,min=0"`

	// Discounts
	DiscountPercent *int   `json:"discount_percent,omitempty" validate:"omitempty,min=0,max=100"`
	DiscountAmount  *int64 `json:"discount_amount,omitempty" validate:"omitempty,min=0"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Notes        *string                `json:"notes,omitempty" validate:"omitempty,max=5000"`
}

// UpdateDealRequest represents a request to update a deal.
type UpdateDealRequest struct {
	// Basic Information
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=5000"`

	// Dates
	SignedDate *string `json:"signed_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	StartDate  *string `json:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	EndDate    *string `json:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02"`

	// Contract Details
	ContractNumber *string `json:"contract_number,omitempty" validate:"omitempty,max=100"`
	ContractTerms  *string `json:"contract_terms,omitempty" validate:"omitempty,max=5000"`
	PaymentTerms   *string `json:"payment_terms,omitempty" validate:"omitempty,oneof=net_15 net_30 net_45 net_60 net_90 due_on_receipt prepaid custom"`
	PaymentMethod  *string `json:"payment_method,omitempty" validate:"omitempty,oneof=credit_card bank_transfer check cash wire_transfer other"`

	// Billing
	BillingContactID *string     `json:"billing_contact_id,omitempty" validate:"omitempty,uuid"`
	BillingAddress   *AddressDTO `json:"billing_address,omitempty"`

	// Shipping
	ShippingContactID *string     `json:"shipping_contact_id,omitempty" validate:"omitempty,uuid"`
	ShippingAddress   *AddressDTO `json:"shipping_address,omitempty"`
	ShippingMethod    *string     `json:"shipping_method,omitempty" validate:"omitempty,max=100"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Notes        *string                `json:"notes,omitempty" validate:"omitempty,max=5000"`

	// Version for optimistic locking
	Version int `json:"version" validate:"required,min=1"`
}

// AddLineItemRequest represents a request to add a line item to a deal.
type AddLineItemRequest struct {
	ProductID       string  `json:"product_id" validate:"required,uuid"`
	ProductName     string  `json:"product_name" validate:"required,max=200"`
	ProductSKU      *string `json:"product_sku,omitempty" validate:"omitempty,max=100"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Quantity        int     `json:"quantity" validate:"required,min=1"`
	UnitPrice       int64   `json:"unit_price" validate:"required,min=0"`
	Currency        string  `json:"currency" validate:"required,len=3"`
	DiscountPercent *int    `json:"discount_percent,omitempty" validate:"omitempty,min=0,max=100"`
	DiscountAmount  *int64  `json:"discount_amount,omitempty" validate:"omitempty,min=0"`
	TaxRate         *int    `json:"tax_rate,omitempty" validate:"omitempty,min=0,max=10000"` // basis points
	Notes           *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// UpdateLineItemRequest represents a request to update a line item.
type UpdateLineItemRequest struct {
	ProductName     *string `json:"product_name,omitempty" validate:"omitempty,max=200"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Quantity        *int    `json:"quantity,omitempty" validate:"omitempty,min=1"`
	UnitPrice       *int64  `json:"unit_price,omitempty" validate:"omitempty,min=0"`
	DiscountPercent *int    `json:"discount_percent,omitempty" validate:"omitempty,min=0,max=100"`
	DiscountAmount  *int64  `json:"discount_amount,omitempty" validate:"omitempty,min=0"`
	TaxRate         *int    `json:"tax_rate,omitempty" validate:"omitempty,min=0,max=10000"`
	Notes           *string `json:"notes,omitempty" validate:"omitempty,max=500"`
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
	ID         string `json:"id"`
	TenantID   string `json:"tenant_id"`
	DealNumber string `json:"deal_number"`

	// Basic Information
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`

	// Status
	Status string `json:"status"`

	// Source
	OpportunityID *string                   `json:"opportunity_id,omitempty"`
	Opportunity   *OpportunityBriefResponse `json:"opportunity,omitempty"`

	// Customer
	CustomerID string            `json:"customer_id"`
	Customer   *CustomerBriefDTO `json:"customer,omitempty"`

	// Value Summary
	Subtotal       MoneyDTO `json:"subtotal"`
	DiscountAmount MoneyDTO `json:"discount_amount"`
	TaxAmount      MoneyDTO `json:"tax_amount"`
	ShippingCost   MoneyDTO `json:"shipping_cost"`
	TotalAmount    MoneyDTO `json:"total_amount"`

	// Line Items
	LineItems     []*DealLineItemResponseDTO `json:"line_items,omitempty"`
	LineItemCount int                        `json:"line_item_count"`

	// Dates
	ClosedDate *time.Time `json:"closed_date,omitempty"`
	SignedDate *time.Time `json:"signed_date,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`

	// Assignment
	OwnerID string        `json:"owner_id"`
	Owner   *UserBriefDTO `json:"owner,omitempty"`

	// Contract Details
	ContractNumber *string `json:"contract_number,omitempty"`
	ContractTerms  *string `json:"contract_terms,omitempty"`
	PaymentTerms   string  `json:"payment_terms"`
	PaymentMethod  *string `json:"payment_method,omitempty"`

	// Billing
	BillingContactID *string          `json:"billing_contact_id,omitempty"`
	BillingContact   *ContactBriefDTO `json:"billing_contact,omitempty"`
	BillingAddress   *AddressDTO      `json:"billing_address,omitempty"`

	// Shipping
	ShippingContactID *string          `json:"shipping_contact_id,omitempty"`
	ShippingContact   *ContactBriefDTO `json:"shipping_contact,omitempty"`
	ShippingAddress   *AddressDTO      `json:"shipping_address,omitempty"`
	ShippingMethod    *string          `json:"shipping_method,omitempty"`

	// Invoicing
	Invoices       []*InvoiceResponseDTO `json:"invoices,omitempty"`
	TotalInvoiced  MoneyDTO              `json:"total_invoiced"`
	InvoiceCount   int                   `json:"invoice_count"`

	// Payments
	Payments      []*PaymentResponseDTO `json:"payments,omitempty"`
	TotalPaid     MoneyDTO              `json:"total_paid"`
	TotalPending  MoneyDTO              `json:"total_pending"`
	PaymentCount  int                   `json:"payment_count"`
	PaymentStatus string                `json:"payment_status"` // unpaid, partial, paid, overpaid

	// Fulfillment
	FulfillmentProgress int                       `json:"fulfillment_progress"` // percentage
	FulfillmentStatus   string                    `json:"fulfillment_status"`   // unfulfilled, partial, fulfilled
	Fulfillments        []*FulfillmentResponseDTO `json:"fulfillments,omitempty"`

	// Cancellation
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
	CancelledBy     *string    `json:"cancelled_by,omitempty"`
	CancelReason    *string    `json:"cancel_reason,omitempty"`
	CancelNotes     *string    `json:"cancel_notes,omitempty"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Notes        *string                `json:"notes,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	Version   int       `json:"version"`
}

// DealBriefResponse represents a brief deal summary.
type DealBriefResponse struct {
	ID                  string     `json:"id"`
	DealNumber          string     `json:"deal_number"`
	Name                string     `json:"name"`
	Status              string     `json:"status"`
	TotalAmount         MoneyDTO   `json:"total_amount"`
	CustomerID          string     `json:"customer_id"`
	CustomerName        string     `json:"customer_name"`
	OwnerID             string     `json:"owner_id"`
	OwnerName           string     `json:"owner_name"`
	PaymentStatus       string     `json:"payment_status"`
	FulfillmentProgress int        `json:"fulfillment_progress"`
	ClosedDate          *time.Time `json:"closed_date,omitempty"`
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
