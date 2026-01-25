package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
	"github.com/kilang-desa-murni/crm/internal/sales/application/usecase"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Deal Request/Response DTOs
// ============================================================================

// CreateDealRequest represents a request to create a new deal
type CreateDealRequest struct {
	OpportunityID   string           `json:"opportunity_id"`
	CustomerID      string           `json:"customer_id"`
	ContactID       string           `json:"contact_id,omitempty"`
	OwnerID         string           `json:"owner_id,omitempty"`
	Title           string           `json:"title"`
	Description     string           `json:"description,omitempty"`
	TotalAmount     string           `json:"total_amount"`
	Currency        string           `json:"currency"`
	ExpectedCloseAt string           `json:"expected_close_at,omitempty"`
	LineItems       []LineItemInput  `json:"line_items,omitempty"`
	Metadata        map[string]any   `json:"metadata,omitempty"`
}

// UpdateDealRequest represents a request to update an existing deal
type UpdateDealRequest struct {
	Title           *string         `json:"title,omitempty"`
	Description     *string         `json:"description,omitempty"`
	TotalAmount     *string         `json:"total_amount,omitempty"`
	DiscountAmount  *string         `json:"discount_amount,omitempty"`
	DiscountPercent *string         `json:"discount_percent,omitempty"`
	TaxAmount       *string         `json:"tax_amount,omitempty"`
	Currency        *string         `json:"currency,omitempty"`
	OwnerID         *string         `json:"owner_id,omitempty"`
	ExpectedCloseAt *string         `json:"expected_close_at,omitempty"`
	Metadata        map[string]any  `json:"metadata,omitempty"`
}

// LineItemInput represents input for a deal line item
type LineItemInput struct {
	ProductID   string         `json:"product_id"`
	ProductName string         `json:"product_name"`
	SKU         string         `json:"sku,omitempty"`
	Quantity    int            `json:"quantity"`
	UnitPrice   string         `json:"unit_price"`
	Discount    string         `json:"discount,omitempty"`
	Tax         string         `json:"tax,omitempty"`
	Description string         `json:"description,omitempty"`
}

// AddLineItemRequest represents a request to add a line item
type AddLineItemRequest struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	SKU         string `json:"sku,omitempty"`
	Quantity    int    `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	Discount    string `json:"discount,omitempty"`
	Tax         string `json:"tax,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateLineItemRequest represents a request to update a line item
type UpdateLineItemRequest struct {
	Quantity    *int    `json:"quantity,omitempty"`
	UnitPrice   *string `json:"unit_price,omitempty"`
	Discount    *string `json:"discount,omitempty"`
	Tax         *string `json:"tax,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateInvoiceRequest represents a request to create an invoice
type CreateInvoiceRequest struct {
	Amount      string `json:"amount"`
	DueDate     string `json:"due_date"`
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// UpdateInvoiceRequest represents a request to update an invoice
type UpdateInvoiceRequest struct {
	Amount      *string `json:"amount,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// RecordPaymentRequest represents a request to record a payment
type RecordPaymentRequest struct {
	InvoiceID     string `json:"invoice_id"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Reference     string `json:"reference,omitempty"`
	Notes         string `json:"notes,omitempty"`
	PaidAt        string `json:"paid_at,omitempty"`
}

// UpdatePaymentRequest represents a request to update a payment
type UpdatePaymentRequest struct {
	Amount        *string `json:"amount,omitempty"`
	PaymentMethod *string `json:"payment_method,omitempty"`
	Reference     *string `json:"reference,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

// ApproveRequest represents a request for approval actions
type ApproveRequest struct {
	ApproverID string `json:"approver_id"`
	Notes      string `json:"notes,omitempty"`
}

// RejectRequest represents a request for rejection actions
type RejectRequest struct {
	RejectorID string `json:"rejector_id"`
	Reason     string `json:"reason"`
}

// CancelDealRequest represents a request to cancel a deal
type CancelDealRequest struct {
	Reason string `json:"reason"`
}

// FulfillmentUpdateRequest represents a request to update fulfillment
type FulfillmentUpdateRequest struct {
	Status      string `json:"status"`
	TrackingRef string `json:"tracking_ref,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

// DealResponse represents a deal in API responses
type DealResponse struct {
	ID              string              `json:"id"`
	TenantID        string              `json:"tenant_id"`
	DealNumber      string              `json:"deal_number"`
	OpportunityID   string              `json:"opportunity_id"`
	CustomerID      string              `json:"customer_id"`
	ContactID       *string             `json:"contact_id,omitempty"`
	OwnerID         *string             `json:"owner_id,omitempty"`
	Title           string              `json:"title"`
	Description     string              `json:"description,omitempty"`
	Status          string              `json:"status"`
	TotalAmount     string              `json:"total_amount"`
	DiscountAmount  string              `json:"discount_amount"`
	DiscountPercent string              `json:"discount_percent"`
	TaxAmount       string              `json:"tax_amount"`
	NetAmount       string              `json:"net_amount"`
	PaidAmount      string              `json:"paid_amount"`
	Currency        string              `json:"currency"`
	ExpectedCloseAt *string             `json:"expected_close_at,omitempty"`
	ClosedAt        *string             `json:"closed_at,omitempty"`
	WonAt           *string             `json:"won_at,omitempty"`
	LostAt          *string             `json:"lost_at,omitempty"`
	LostReason      string              `json:"lost_reason,omitempty"`
	LineItems       []LineItemResponse  `json:"line_items,omitempty"`
	Invoices        []InvoiceResponse   `json:"invoices,omitempty"`
	Payments        []PaymentResponse   `json:"payments,omitempty"`
	Metadata        map[string]any      `json:"metadata,omitempty"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
}

// LineItemResponse represents a line item in API responses
type LineItemResponse struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	SKU         string `json:"sku,omitempty"`
	Quantity    int    `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	Discount    string `json:"discount"`
	Tax         string `json:"tax"`
	TotalPrice  string `json:"total_price"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// InvoiceResponse represents an invoice in API responses
type InvoiceResponse struct {
	ID            string `json:"id"`
	InvoiceNumber string `json:"invoice_number"`
	Amount        string `json:"amount"`
	PaidAmount    string `json:"paid_amount"`
	Status        string `json:"status"`
	DueDate       string `json:"due_date"`
	IssuedAt      string `json:"issued_at"`
	PaidAt        *string `json:"paid_at,omitempty"`
	CancelledAt   *string `json:"cancelled_at,omitempty"`
	Description   string `json:"description,omitempty"`
	Notes         string `json:"notes,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// PaymentResponse represents a payment in API responses
type PaymentResponse struct {
	ID            string  `json:"id"`
	InvoiceID     string  `json:"invoice_id"`
	Amount        string  `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	Reference     string  `json:"reference,omitempty"`
	Status        string  `json:"status"`
	PaidAt        string  `json:"paid_at"`
	RefundedAt    *string `json:"refunded_at,omitempty"`
	Notes         string  `json:"notes,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// DealStatisticsResponse represents deal statistics
type DealStatisticsResponse struct {
	TotalDeals          int64  `json:"total_deals"`
	DraftDeals          int64  `json:"draft_deals"`
	PendingApproval     int64  `json:"pending_approval"`
	ApprovedDeals       int64  `json:"approved_deals"`
	InProgressDeals     int64  `json:"in_progress_deals"`
	WonDeals            int64  `json:"won_deals"`
	LostDeals           int64  `json:"lost_deals"`
	CancelledDeals      int64  `json:"cancelled_deals"`
	TotalValue          string `json:"total_value"`
	WonValue            string `json:"won_value"`
	PendingValue        string `json:"pending_value"`
	AverageValue        string `json:"average_value"`
	WinRate             string `json:"win_rate"`
	AverageDaysToClose  int    `json:"average_days_to_close"`
}

// ============================================================================
// Deal Handler Methods
// ============================================================================

// CreateDeal handles POST /deals
func (h *Handler) CreateDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	var req CreateDealRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	// Parse required UUIDs
	opportunityID, err := uuid.Parse(req.OpportunityID)
	if err != nil {
		h.respondError(w, ErrValidation("opportunity_id", "invalid UUID format"))
		return
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		h.respondError(w, ErrValidation("customer_id", "invalid UUID format"))
		return
	}

	// Parse total amount
	totalAmount, err := decimal.NewFromString(req.TotalAmount)
	if err != nil {
		h.respondError(w, ErrValidation("total_amount", "invalid decimal format"))
		return
	}

	// Parse optional UUIDs
	var contactID, ownerID *uuid.UUID
	if req.ContactID != "" {
		id, err := uuid.Parse(req.ContactID)
		if err != nil {
			h.respondError(w, ErrValidation("contact_id", "invalid UUID format"))
			return
		}
		contactID = &id
	}
	if req.OwnerID != "" {
		id, err := uuid.Parse(req.OwnerID)
		if err != nil {
			h.respondError(w, ErrValidation("owner_id", "invalid UUID format"))
			return
		}
		ownerID = &id
	} else if userID != nil {
		ownerID = userID
	}

	// Parse expected close date
	var expectedCloseAt *time.Time
	if req.ExpectedCloseAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpectedCloseAt)
		if err != nil {
			h.respondError(w, ErrValidation("expected_close_at", "invalid date format, use RFC3339"))
			return
		}
		expectedCloseAt = &t
	}

	// Build line items
	var lineItems []command.CreateDealLineItemInput
	for i, item := range req.LineItems {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			h.respondError(w, ErrValidation("line_items["+strconv.Itoa(i)+"].product_id", "invalid UUID format"))
			return
		}

		unitPrice, err := decimal.NewFromString(item.UnitPrice)
		if err != nil {
			h.respondError(w, ErrValidation("line_items["+strconv.Itoa(i)+"].unit_price", "invalid decimal format"))
			return
		}

		var discount, tax decimal.Decimal
		if item.Discount != "" {
			discount, err = decimal.NewFromString(item.Discount)
			if err != nil {
				h.respondError(w, ErrValidation("line_items["+strconv.Itoa(i)+"].discount", "invalid decimal format"))
				return
			}
		}
		if item.Tax != "" {
			tax, err = decimal.NewFromString(item.Tax)
			if err != nil {
				h.respondError(w, ErrValidation("line_items["+strconv.Itoa(i)+"].tax", "invalid decimal format"))
				return
			}
		}

		lineItems = append(lineItems, command.CreateDealLineItemInput{
			ProductID:   productID,
			ProductName: item.ProductName,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   unitPrice,
			Discount:    discount,
			Tax:         tax,
			Description: item.Description,
		})
	}

	// Build command
	cmd := command.CreateDealCommand{
		TenantID:        tenantID,
		OpportunityID:   opportunityID,
		CustomerID:      customerID,
		ContactID:       contactID,
		OwnerID:         ownerID,
		Title:           req.Title,
		Description:     req.Description,
		TotalAmount:     totalAmount,
		Currency:        req.Currency,
		ExpectedCloseAt: expectedCloseAt,
		LineItems:       lineItems,
		Metadata:        req.Metadata,
	}

	deal, err := h.dealUseCase.CreateDeal(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toDealResponse(deal))
}

// GetDeal handles GET /deals/{dealID}
func (h *Handler) GetDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	qry := query.GetDealQuery{
		TenantID: tenantID,
		DealID:   dealID,
	}

	deal, err := h.dealUseCase.GetDeal(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// UpdateDeal handles PUT /deals/{dealID}
func (h *Handler) UpdateDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	var req UpdateDealRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	// Parse optional amounts
	var totalAmount, discountAmount, discountPercent, taxAmount *decimal.Decimal
	if req.TotalAmount != nil {
		amt, err := decimal.NewFromString(*req.TotalAmount)
		if err != nil {
			h.respondError(w, ErrValidation("total_amount", "invalid decimal format"))
			return
		}
		totalAmount = &amt
	}
	if req.DiscountAmount != nil {
		amt, err := decimal.NewFromString(*req.DiscountAmount)
		if err != nil {
			h.respondError(w, ErrValidation("discount_amount", "invalid decimal format"))
			return
		}
		discountAmount = &amt
	}
	if req.DiscountPercent != nil {
		pct, err := decimal.NewFromString(*req.DiscountPercent)
		if err != nil {
			h.respondError(w, ErrValidation("discount_percent", "invalid decimal format"))
			return
		}
		discountPercent = &pct
	}
	if req.TaxAmount != nil {
		amt, err := decimal.NewFromString(*req.TaxAmount)
		if err != nil {
			h.respondError(w, ErrValidation("tax_amount", "invalid decimal format"))
			return
		}
		taxAmount = &amt
	}

	// Parse optional owner ID
	var ownerID *uuid.UUID
	if req.OwnerID != nil {
		id, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			h.respondError(w, ErrValidation("owner_id", "invalid UUID format"))
			return
		}
		ownerID = &id
	}

	// Parse expected close date
	var expectedCloseAt *time.Time
	if req.ExpectedCloseAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpectedCloseAt)
		if err != nil {
			h.respondError(w, ErrValidation("expected_close_at", "invalid date format, use RFC3339"))
			return
		}
		expectedCloseAt = &t
	}

	cmd := command.UpdateDealCommand{
		TenantID:        tenantID,
		DealID:          dealID,
		Title:           req.Title,
		Description:     req.Description,
		TotalAmount:     totalAmount,
		DiscountAmount:  discountAmount,
		DiscountPercent: discountPercent,
		TaxAmount:       taxAmount,
		Currency:        req.Currency,
		OwnerID:         ownerID,
		ExpectedCloseAt: expectedCloseAt,
		Metadata:        req.Metadata,
	}

	deal, err := h.dealUseCase.UpdateDeal(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// DeleteDeal handles DELETE /deals/{dealID}
func (h *Handler) DeleteDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	cmd := command.DeleteDealCommand{
		TenantID: tenantID,
		DealID:   dealID,
	}

	if err := h.dealUseCase.DeleteDeal(ctx, cmd); err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondSuccess(w, http.StatusNoContent, nil)
}

// ListDeals handles GET /deals
func (h *Handler) ListDeals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	filter := h.buildDealFilter(r)
	opts := h.buildListOptions(r)

	qry := query.ListDealsQuery{
		TenantID:    tenantID,
		Filter:      filter,
		ListOptions: opts,
	}

	deals, total, err := h.dealUseCase.ListDeals(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	responses := make([]DealResponse, len(deals))
	for i, deal := range deals {
		responses[i] = h.toDealResponse(deal)
	}

	h.respondList(w, responses, total, opts.Page, opts.PageSize)
}

// ============================================================================
// Deal Status Operations
// ============================================================================

// SubmitDealForApproval handles POST /deals/{dealID}/submit
func (h *Handler) SubmitDealForApproval(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	cmd := command.SubmitDealForApprovalCommand{
		TenantID: tenantID,
		DealID:   dealID,
	}

	deal, err := h.dealUseCase.SubmitForApproval(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// ApproveDeal handles POST /deals/{dealID}/approve
func (h *Handler) ApproveDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	var req ApproveRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		h.respondError(w, ErrValidation("approver_id", "invalid UUID format"))
		return
	}

	cmd := command.ApproveDealCommand{
		TenantID:   tenantID,
		DealID:     dealID,
		ApproverID: approverID,
		Notes:      req.Notes,
	}

	deal, err := h.dealUseCase.ApproveDeal(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// RejectDeal handles POST /deals/{dealID}/reject
func (h *Handler) RejectDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	var req RejectRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	rejectorID, err := uuid.Parse(req.RejectorID)
	if err != nil {
		h.respondError(w, ErrValidation("rejector_id", "invalid UUID format"))
		return
	}

	cmd := command.RejectDealCommand{
		TenantID:   tenantID,
		DealID:     dealID,
		RejectorID: rejectorID,
		Reason:     req.Reason,
	}

	deal, err := h.dealUseCase.RejectDeal(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// StartDealFulfillment handles POST /deals/{dealID}/start-fulfillment
func (h *Handler) StartDealFulfillment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	cmd := command.StartDealFulfillmentCommand{
		TenantID: tenantID,
		DealID:   dealID,
	}

	deal, err := h.dealUseCase.StartFulfillment(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// WinDeal handles POST /deals/{dealID}/win
func (h *Handler) WinDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	cmd := command.WinDealCommand{
		TenantID: tenantID,
		DealID:   dealID,
	}

	deal, err := h.dealUseCase.WinDeal(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// LoseDeal handles POST /deals/{dealID}/lose
func (h *Handler) LoseDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	cmd := command.LoseDealCommand{
		TenantID: tenantID,
		DealID:   dealID,
		Reason:   req.Reason,
	}

	deal, err := h.dealUseCase.LoseDeal(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// CancelDeal handles POST /deals/{dealID}/cancel
func (h *Handler) CancelDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	var req CancelDealRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	cmd := command.CancelDealCommand{
		TenantID: tenantID,
		DealID:   dealID,
		Reason:   req.Reason,
	}

	deal, err := h.dealUseCase.CancelDeal(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// ReopenDeal handles POST /deals/{dealID}/reopen
func (h *Handler) ReopenDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	cmd := command.ReopenDealCommand{
		TenantID: tenantID,
		DealID:   dealID,
	}

	deal, err := h.dealUseCase.ReopenDeal(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// ============================================================================
// Line Item Operations
// ============================================================================

// AddDealLineItem handles POST /deals/{dealID}/line-items
func (h *Handler) AddDealLineItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	var req AddLineItemRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		h.respondError(w, ErrValidation("product_id", "invalid UUID format"))
		return
	}

	unitPrice, err := decimal.NewFromString(req.UnitPrice)
	if err != nil {
		h.respondError(w, ErrValidation("unit_price", "invalid decimal format"))
		return
	}

	var discount, tax decimal.Decimal
	if req.Discount != "" {
		discount, err = decimal.NewFromString(req.Discount)
		if err != nil {
			h.respondError(w, ErrValidation("discount", "invalid decimal format"))
			return
		}
	}
	if req.Tax != "" {
		tax, err = decimal.NewFromString(req.Tax)
		if err != nil {
			h.respondError(w, ErrValidation("tax", "invalid decimal format"))
			return
		}
	}

	cmd := command.AddDealLineItemCommand{
		TenantID:    tenantID,
		DealID:      dealID,
		ProductID:   productID,
		ProductName: req.ProductName,
		SKU:         req.SKU,
		Quantity:    req.Quantity,
		UnitPrice:   unitPrice,
		Discount:    discount,
		Tax:         tax,
		Description: req.Description,
	}

	deal, err := h.dealUseCase.AddLineItem(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toDealResponse(deal))
}

// UpdateDealLineItem handles PUT /deals/{dealID}/line-items/{lineItemID}
func (h *Handler) UpdateDealLineItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	lineItemID, err := uuid.Parse(chi.URLParam(r, "lineItemID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid line item ID"))
		return
	}

	var req UpdateLineItemRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	// Parse optional amounts
	var unitPrice, discount, tax *decimal.Decimal
	if req.UnitPrice != nil {
		up, err := decimal.NewFromString(*req.UnitPrice)
		if err != nil {
			h.respondError(w, ErrValidation("unit_price", "invalid decimal format"))
			return
		}
		unitPrice = &up
	}
	if req.Discount != nil {
		d, err := decimal.NewFromString(*req.Discount)
		if err != nil {
			h.respondError(w, ErrValidation("discount", "invalid decimal format"))
			return
		}
		discount = &d
	}
	if req.Tax != nil {
		t, err := decimal.NewFromString(*req.Tax)
		if err != nil {
			h.respondError(w, ErrValidation("tax", "invalid decimal format"))
			return
		}
		tax = &t
	}

	cmd := command.UpdateDealLineItemCommand{
		TenantID:    tenantID,
		DealID:      dealID,
		LineItemID:  lineItemID,
		Quantity:    req.Quantity,
		UnitPrice:   unitPrice,
		Discount:    discount,
		Tax:         tax,
		Description: req.Description,
	}

	deal, err := h.dealUseCase.UpdateLineItem(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// RemoveDealLineItem handles DELETE /deals/{dealID}/line-items/{lineItemID}
func (h *Handler) RemoveDealLineItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	lineItemID, err := uuid.Parse(chi.URLParam(r, "lineItemID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid line item ID"))
		return
	}

	cmd := command.RemoveDealLineItemCommand{
		TenantID:   tenantID,
		DealID:     dealID,
		LineItemID: lineItemID,
	}

	deal, err := h.dealUseCase.RemoveLineItem(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// ============================================================================
// Invoice Operations
// ============================================================================

// CreateDealInvoice handles POST /deals/{dealID}/invoices
func (h *Handler) CreateDealInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	var req CreateInvoiceRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		h.respondError(w, ErrValidation("amount", "invalid decimal format"))
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		h.respondError(w, ErrValidation("due_date", "invalid date format, use YYYY-MM-DD"))
		return
	}

	cmd := command.CreateDealInvoiceCommand{
		TenantID:    tenantID,
		DealID:      dealID,
		Amount:      amount,
		DueDate:     dueDate,
		Description: req.Description,
		Notes:       req.Notes,
	}

	deal, err := h.dealUseCase.CreateInvoice(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toDealResponse(deal))
}

// UpdateDealInvoice handles PUT /deals/{dealID}/invoices/{invoiceID}
func (h *Handler) UpdateDealInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid invoice ID"))
		return
	}

	var req UpdateInvoiceRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	// Parse optional amount
	var amount *decimal.Decimal
	if req.Amount != nil {
		amt, err := decimal.NewFromString(*req.Amount)
		if err != nil {
			h.respondError(w, ErrValidation("amount", "invalid decimal format"))
			return
		}
		amount = &amt
	}

	// Parse optional due date
	var dueDate *time.Time
	if req.DueDate != nil {
		dd, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			h.respondError(w, ErrValidation("due_date", "invalid date format, use YYYY-MM-DD"))
			return
		}
		dueDate = &dd
	}

	cmd := command.UpdateDealInvoiceCommand{
		TenantID:    tenantID,
		DealID:      dealID,
		InvoiceID:   invoiceID,
		Amount:      amount,
		DueDate:     dueDate,
		Description: req.Description,
		Notes:       req.Notes,
	}

	deal, err := h.dealUseCase.UpdateInvoice(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// IssueInvoice handles POST /deals/{dealID}/invoices/{invoiceID}/issue
func (h *Handler) IssueInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid invoice ID"))
		return
	}

	cmd := command.IssueDealInvoiceCommand{
		TenantID:  tenantID,
		DealID:    dealID,
		InvoiceID: invoiceID,
	}

	deal, err := h.dealUseCase.IssueInvoice(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// CancelInvoice handles POST /deals/{dealID}/invoices/{invoiceID}/cancel
func (h *Handler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid invoice ID"))
		return
	}

	cmd := command.CancelDealInvoiceCommand{
		TenantID:  tenantID,
		DealID:    dealID,
		InvoiceID: invoiceID,
	}

	deal, err := h.dealUseCase.CancelInvoice(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// ============================================================================
// Payment Operations
// ============================================================================

// RecordPayment handles POST /deals/{dealID}/payments
func (h *Handler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	var req RecordPaymentRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	invoiceID, err := uuid.Parse(req.InvoiceID)
	if err != nil {
		h.respondError(w, ErrValidation("invoice_id", "invalid UUID format"))
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		h.respondError(w, ErrValidation("amount", "invalid decimal format"))
		return
	}

	// Parse optional paid at
	paidAt := time.Now()
	if req.PaidAt != "" {
		paidAt, err = time.Parse(time.RFC3339, req.PaidAt)
		if err != nil {
			h.respondError(w, ErrValidation("paid_at", "invalid date format, use RFC3339"))
			return
		}
	}

	cmd := command.RecordDealPaymentCommand{
		TenantID:      tenantID,
		DealID:        dealID,
		InvoiceID:     invoiceID,
		Amount:        amount,
		PaymentMethod: req.PaymentMethod,
		Reference:     req.Reference,
		Notes:         req.Notes,
		PaidAt:        paidAt,
	}

	deal, err := h.dealUseCase.RecordPayment(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toDealResponse(deal))
}

// UpdatePayment handles PUT /deals/{dealID}/payments/{paymentID}
func (h *Handler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	paymentID, err := uuid.Parse(chi.URLParam(r, "paymentID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid payment ID"))
		return
	}

	var req UpdatePaymentRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	// Parse optional amount
	var amount *decimal.Decimal
	if req.Amount != nil {
		amt, err := decimal.NewFromString(*req.Amount)
		if err != nil {
			h.respondError(w, ErrValidation("amount", "invalid decimal format"))
			return
		}
		amount = &amt
	}

	cmd := command.UpdateDealPaymentCommand{
		TenantID:      tenantID,
		DealID:        dealID,
		PaymentID:     paymentID,
		Amount:        amount,
		PaymentMethod: req.PaymentMethod,
		Reference:     req.Reference,
		Notes:         req.Notes,
	}

	deal, err := h.dealUseCase.UpdatePayment(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// RefundPayment handles POST /deals/{dealID}/payments/{paymentID}/refund
func (h *Handler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealID, err := h.getUUIDParam(r, "dealID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid deal ID"))
		return
	}

	paymentID, err := uuid.Parse(chi.URLParam(r, "paymentID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid payment ID"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	cmd := command.RefundDealPaymentCommand{
		TenantID:  tenantID,
		DealID:    dealID,
		PaymentID: paymentID,
		Reason:    req.Reason,
	}

	deal, err := h.dealUseCase.RefundPayment(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toDealResponse(deal))
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkAssignDeals handles POST /deals/bulk/assign
func (h *Handler) BulkAssignDeals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	var req struct {
		DealIDs []string `json:"deal_ids"`
		OwnerID string   `json:"owner_id"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		h.respondError(w, ErrValidation("owner_id", "invalid UUID format"))
		return
	}

	dealIDs := make([]uuid.UUID, 0, len(req.DealIDs))
	for i, idStr := range req.DealIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.respondError(w, ErrValidation("deal_ids["+strconv.Itoa(i)+"]", "invalid UUID format"))
			return
		}
		dealIDs = append(dealIDs, id)
	}

	cmd := command.BulkAssignDealsCommand{
		TenantID: tenantID,
		DealIDs:  dealIDs,
		OwnerID:  ownerID,
	}

	count, err := h.dealUseCase.BulkAssign(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message":       "deals assigned successfully",
		"updated_count": count,
	})
}

// BulkUpdateDealStatus handles POST /deals/bulk/status
func (h *Handler) BulkUpdateDealStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	var req struct {
		DealIDs []string `json:"deal_ids"`
		Status  string   `json:"status"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	dealIDs := make([]uuid.UUID, 0, len(req.DealIDs))
	for i, idStr := range req.DealIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.respondError(w, ErrValidation("deal_ids["+strconv.Itoa(i)+"]", "invalid UUID format"))
			return
		}
		dealIDs = append(dealIDs, id)
	}

	cmd := command.BulkUpdateDealStatusCommand{
		TenantID: tenantID,
		DealIDs:  dealIDs,
		Status:   entity.DealStatus(req.Status),
	}

	count, err := h.dealUseCase.BulkUpdateStatus(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message":       "deal statuses updated successfully",
		"updated_count": count,
	})
}

// ============================================================================
// Statistics and Reports
// ============================================================================

// GetDealStatistics handles GET /deals/statistics
func (h *Handler) GetDealStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	qry := query.GetDealStatisticsQuery{
		TenantID: tenantID,
	}

	stats, err := h.dealUseCase.GetStatistics(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, DealStatisticsResponse{
		TotalDeals:         stats.TotalDeals,
		DraftDeals:         stats.DraftDeals,
		PendingApproval:    stats.PendingApproval,
		ApprovedDeals:      stats.ApprovedDeals,
		InProgressDeals:    stats.InProgressDeals,
		WonDeals:           stats.WonDeals,
		LostDeals:          stats.LostDeals,
		CancelledDeals:     stats.CancelledDeals,
		TotalValue:         stats.TotalValue.String(),
		WonValue:           stats.WonValue.String(),
		PendingValue:       stats.PendingValue.String(),
		AverageValue:       stats.AverageValue.String(),
		WinRate:            stats.WinRate.String(),
		AverageDaysToClose: stats.AverageDaysToClose,
	})
}

// GetDealsByOwner handles GET /deals/by-owner/{ownerID}
func (h *Handler) GetDealsByOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	ownerID, err := uuid.Parse(chi.URLParam(r, "ownerID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid owner ID"))
		return
	}

	opts := h.buildListOptions(r)

	qry := query.GetDealsByOwnerQuery{
		TenantID:    tenantID,
		OwnerID:     ownerID,
		ListOptions: opts,
	}

	deals, total, err := h.dealUseCase.GetByOwner(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	responses := make([]DealResponse, len(deals))
	for i, deal := range deals {
		responses[i] = h.toDealResponse(deal)
	}

	h.respondList(w, responses, total, opts.Page, opts.PageSize)
}

// GetDealsByCustomer handles GET /deals/by-customer/{customerID}
func (h *Handler) GetDealsByCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	customerID, err := uuid.Parse(chi.URLParam(r, "customerID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid customer ID"))
		return
	}

	opts := h.buildListOptions(r)

	qry := query.GetDealsByCustomerQuery{
		TenantID:    tenantID,
		CustomerID:  customerID,
		ListOptions: opts,
	}

	deals, total, err := h.dealUseCase.GetByCustomer(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	responses := make([]DealResponse, len(deals))
	for i, deal := range deals {
		responses[i] = h.toDealResponse(deal)
	}

	h.respondList(w, responses, total, opts.Page, opts.PageSize)
}

// GetOverdueInvoices handles GET /deals/overdue-invoices
func (h *Handler) GetOverdueInvoices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	opts := h.buildListOptions(r)

	qry := query.GetDealsWithOverdueInvoicesQuery{
		TenantID:    tenantID,
		ListOptions: opts,
	}

	deals, total, err := h.dealUseCase.GetWithOverdueInvoices(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	responses := make([]DealResponse, len(deals))
	for i, deal := range deals {
		responses[i] = h.toDealResponse(deal)
	}

	h.respondList(w, responses, total, opts.Page, opts.PageSize)
}

// GetPendingPayments handles GET /deals/pending-payments
func (h *Handler) GetPendingPayments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	opts := h.buildListOptions(r)

	qry := query.GetDealsWithPendingPaymentsQuery{
		TenantID:    tenantID,
		ListOptions: opts,
	}

	deals, total, err := h.dealUseCase.GetWithPendingPayments(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	responses := make([]DealResponse, len(deals))
	for i, deal := range deals {
		responses[i] = h.toDealResponse(deal)
	}

	h.respondList(w, responses, total, opts.Page, opts.PageSize)
}

// GetRevenueByPeriod handles GET /deals/revenue
func (h *Handler) GetRevenueByPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	// Parse date range from query params
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var start, end time.Time
	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			h.respondError(w, ErrValidation("start_date", "invalid date format, use YYYY-MM-DD"))
			return
		}
	} else {
		// Default to start of current month
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			h.respondError(w, ErrValidation("end_date", "invalid date format, use YYYY-MM-DD"))
			return
		}
	} else {
		// Default to end of current month
		now := time.Now()
		end = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, time.UTC)
	}

	qry := query.GetRevenueByPeriodQuery{
		TenantID:  tenantID,
		StartDate: start,
		EndDate:   end,
	}

	revenue, err := h.dealUseCase.GetRevenueByPeriod(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"start_date":    start.Format("2006-01-02"),
		"end_date":      end.Format("2006-01-02"),
		"total_revenue": revenue.TotalRevenue.String(),
		"won_deals":     revenue.WonDeals,
		"paid_amount":   revenue.PaidAmount.String(),
		"pending_amount": revenue.PendingAmount.String(),
	})
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

// toDealResponse converts a domain deal to an API response
func (h *Handler) toDealResponse(deal *entity.Deal) DealResponse {
	resp := DealResponse{
		ID:              deal.ID.String(),
		TenantID:        deal.TenantID.String(),
		DealNumber:      deal.DealNumber,
		OpportunityID:   deal.OpportunityID.String(),
		CustomerID:      deal.CustomerID.String(),
		Title:           deal.Title,
		Description:     deal.Description,
		Status:          string(deal.Status),
		TotalAmount:     deal.TotalAmount.String(),
		DiscountAmount:  deal.DiscountAmount.String(),
		DiscountPercent: deal.DiscountPercent.String(),
		TaxAmount:       deal.TaxAmount.String(),
		NetAmount:       deal.NetAmount.String(),
		PaidAmount:      deal.PaidAmount.String(),
		Currency:        deal.Currency,
		LostReason:      deal.LostReason,
		Metadata:        deal.Metadata,
		CreatedAt:       deal.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       deal.UpdatedAt.Format(time.RFC3339),
	}

	if deal.ContactID != nil {
		s := deal.ContactID.String()
		resp.ContactID = &s
	}
	if deal.OwnerID != nil {
		s := deal.OwnerID.String()
		resp.OwnerID = &s
	}
	if deal.ExpectedCloseAt != nil {
		s := deal.ExpectedCloseAt.Format(time.RFC3339)
		resp.ExpectedCloseAt = &s
	}
	if deal.ClosedAt != nil {
		s := deal.ClosedAt.Format(time.RFC3339)
		resp.ClosedAt = &s
	}
	if deal.WonAt != nil {
		s := deal.WonAt.Format(time.RFC3339)
		resp.WonAt = &s
	}
	if deal.LostAt != nil {
		s := deal.LostAt.Format(time.RFC3339)
		resp.LostAt = &s
	}

	// Map line items
	resp.LineItems = make([]LineItemResponse, len(deal.LineItems))
	for i, item := range deal.LineItems {
		resp.LineItems[i] = LineItemResponse{
			ID:          item.ID.String(),
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductName,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.String(),
			Discount:    item.Discount.String(),
			Tax:         item.Tax.String(),
			TotalPrice:  item.TotalPrice.String(),
			Description: item.Description,
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
		}
	}

	// Map invoices
	resp.Invoices = make([]InvoiceResponse, len(deal.Invoices))
	for i, inv := range deal.Invoices {
		invResp := InvoiceResponse{
			ID:            inv.ID.String(),
			InvoiceNumber: inv.InvoiceNumber,
			Amount:        inv.Amount.String(),
			PaidAmount:    inv.PaidAmount.String(),
			Status:        string(inv.Status),
			DueDate:       inv.DueDate.Format("2006-01-02"),
			IssuedAt:      inv.IssuedAt.Format(time.RFC3339),
			Description:   inv.Description,
			Notes:         inv.Notes,
			CreatedAt:     inv.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     inv.UpdatedAt.Format(time.RFC3339),
		}
		if inv.PaidAt != nil {
			s := inv.PaidAt.Format(time.RFC3339)
			invResp.PaidAt = &s
		}
		if inv.CancelledAt != nil {
			s := inv.CancelledAt.Format(time.RFC3339)
			invResp.CancelledAt = &s
		}
		resp.Invoices[i] = invResp
	}

	// Map payments
	resp.Payments = make([]PaymentResponse, len(deal.Payments))
	for i, pmt := range deal.Payments {
		pmtResp := PaymentResponse{
			ID:            pmt.ID.String(),
			InvoiceID:     pmt.InvoiceID.String(),
			Amount:        pmt.Amount.String(),
			PaymentMethod: pmt.PaymentMethod,
			Reference:     pmt.Reference,
			Status:        string(pmt.Status),
			PaidAt:        pmt.PaidAt.Format(time.RFC3339),
			Notes:         pmt.Notes,
			CreatedAt:     pmt.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     pmt.UpdatedAt.Format(time.RFC3339),
		}
		if pmt.RefundedAt != nil {
			s := pmt.RefundedAt.Format(time.RFC3339)
			pmtResp.RefundedAt = &s
		}
		resp.Payments[i] = pmtResp
	}

	return resp
}
