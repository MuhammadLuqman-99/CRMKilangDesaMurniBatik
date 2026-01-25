package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
)

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

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	var req dto.CreateDealRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.Create(ctx, tenantID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, deal)
}

// GetDeal handles GET /deals/{dealID}
func (h *Handler) GetDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	deal, err := h.dealUseCase.GetByID(ctx, tenantID, dealID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// GetDealByCode handles GET /deals/code/{code}
func (h *Handler) GetDealByCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	code := chi.URLParam(r, "code")
	if code == "" {
		h.respondError(w, ErrMissingParameter("code"))
		return
	}

	deal, err := h.dealUseCase.GetByCode(ctx, tenantID, code)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// UpdateDeal handles PUT /deals/{dealID}
func (h *Handler) UpdateDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	var req dto.UpdateDealRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.Update(ctx, tenantID, dealID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// DeleteDeal handles DELETE /deals/{dealID}
func (h *Handler) DeleteDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	if err := h.dealUseCase.Delete(ctx, tenantID, dealID, userID); err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusNoContent, nil)
}

// ListDeals handles GET /deals
func (h *Handler) ListDeals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	filter := h.parseDealFilter(r)

	deals, err := h.dealUseCase.List(ctx, tenantID, filter)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deals)
}

// ============================================================================
// Line Item Operations
// ============================================================================

// AddLineItem handles POST /deals/{dealID}/line-items
func (h *Handler) AddLineItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	var req dto.AddLineItemRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.AddLineItem(ctx, tenantID, dealID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// UpdateLineItem handles PUT /deals/{dealID}/line-items/{lineItemID}
func (h *Handler) UpdateLineItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	lineItemIDStr := chi.URLParam(r, "lineItemID")
	lineItemID, err := uuid.Parse(lineItemIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("lineItemID", "invalid UUID format"))
		return
	}

	var req dto.UpdateLineItemRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.UpdateLineItem(ctx, tenantID, dealID, lineItemID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// RemoveLineItem handles DELETE /deals/{dealID}/line-items/{lineItemID}
func (h *Handler) RemoveLineItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	lineItemIDStr := chi.URLParam(r, "lineItemID")
	lineItemID, err := uuid.Parse(lineItemIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("lineItemID", "invalid UUID format"))
		return
	}

	deal, err := h.dealUseCase.RemoveLineItem(ctx, tenantID, dealID, lineItemID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// FulfillLineItem handles POST /deals/{dealID}/line-items/{lineItemID}/fulfill
func (h *Handler) FulfillLineItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	lineItemIDStr := chi.URLParam(r, "lineItemID")
	lineItemID, err := uuid.Parse(lineItemIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("lineItemID", "invalid UUID format"))
		return
	}

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.FulfillLineItem(ctx, tenantID, dealID, lineItemID, userID, req.Quantity)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// ============================================================================
// Invoice Operations
// ============================================================================

// CreateInvoice handles POST /deals/{dealID}/invoices
func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	var req dto.CreateInvoiceRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.CreateInvoice(ctx, tenantID, dealID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, deal)
}

// SendInvoice handles POST /deals/{dealID}/invoices/{invoiceID}/send
func (h *Handler) SendInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	invoiceIDStr := chi.URLParam(r, "invoiceID")
	invoiceID, err := uuid.Parse(invoiceIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("invoiceID", "invalid UUID format"))
		return
	}

	deal, err := h.dealUseCase.SendInvoice(ctx, tenantID, dealID, invoiceID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
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

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	var req dto.RecordPaymentRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.RecordPayment(ctx, tenantID, dealID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, deal)
}

// ============================================================================
// Status Operations
// ============================================================================

// ActivateDeal handles POST /deals/{dealID}/activate
func (h *Handler) ActivateDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	deal, err := h.dealUseCase.Activate(ctx, tenantID, dealID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// FulfillDeal handles POST /deals/{dealID}/fulfill
func (h *Handler) FulfillDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	deal, err := h.dealUseCase.Fulfill(ctx, tenantID, dealID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// CancelDeal handles POST /deals/{dealID}/cancel
func (h *Handler) CancelDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.Cancel(ctx, tenantID, dealID, userID, req.Reason)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// PutDealOnHold handles POST /deals/{dealID}/hold
func (h *Handler) PutDealOnHold(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	deal, err := h.dealUseCase.PutOnHold(ctx, tenantID, dealID, userID, req.Reason)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// ResumeDeal handles POST /deals/{dealID}/resume
func (h *Handler) ResumeDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	userIDPtr, _ := h.getUserID(ctx)
	if userIDPtr == nil {
		h.respondError(w, ErrUnauthorized("user identification required"))
		return
	}
	userID := *userIDPtr

	dealIDStr := chi.URLParam(r, "dealID")
	dealID, err := uuid.Parse(dealIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("dealID", "invalid UUID format"))
		return
	}

	deal, err := h.dealUseCase.Resume(ctx, tenantID, dealID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, deal)
}

// ============================================================================
// Statistics
// ============================================================================

// GetDealStatistics handles GET /deals/statistics
func (h *Handler) GetDealStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	stats, err := h.dealUseCase.GetStatistics(ctx, tenantID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// GetDealsByOwner handles GET /deals/by-owner/{ownerID}
func (h *Handler) GetDealsByOwner(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("deals by owner endpoint not yet implemented"))
}

// GetDealsByCustomer handles GET /deals/by-customer/{customerID}
func (h *Handler) GetDealsByCustomer(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("deals by customer endpoint not yet implemented"))
}

// GetOverdueInvoices handles GET /deals/overdue-invoices
func (h *Handler) GetOverdueInvoices(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("overdue invoices endpoint not yet implemented"))
}

// GetPendingPayments handles GET /deals/pending-payments
func (h *Handler) GetPendingPayments(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("pending payments endpoint not yet implemented"))
}

// GetRevenueByPeriod handles GET /deals/revenue
func (h *Handler) GetRevenueByPeriod(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("revenue by period endpoint not yet implemented"))
}

// BulkAssignDeals handles POST /deals/bulk/assign
func (h *Handler) BulkAssignDeals(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("bulk assign deals endpoint not yet implemented"))
}

// BulkUpdateDealStatus handles POST /deals/bulk/status
func (h *Handler) BulkUpdateDealStatus(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("bulk update deal status endpoint not yet implemented"))
}

// SubmitDealForApproval handles POST /deals/{dealID}/submit
func (h *Handler) SubmitDealForApproval(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("submit for approval endpoint not yet implemented"))
}

// ApproveDeal handles POST /deals/{dealID}/approve
func (h *Handler) ApproveDeal(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("approve deal endpoint not yet implemented"))
}

// RejectDeal handles POST /deals/{dealID}/reject
func (h *Handler) RejectDeal(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("reject deal endpoint not yet implemented"))
}

// StartDealFulfillment handles POST /deals/{dealID}/start-fulfillment
func (h *Handler) StartDealFulfillment(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("start fulfillment endpoint not yet implemented"))
}

// WinDeal handles POST /deals/{dealID}/win
func (h *Handler) WinDeal(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("win deal endpoint not yet implemented"))
}

// LoseDeal handles POST /deals/{dealID}/lose
func (h *Handler) LoseDeal(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("lose deal endpoint not yet implemented"))
}

// ReopenDeal handles POST /deals/{dealID}/reopen
func (h *Handler) ReopenDeal(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("reopen deal endpoint not yet implemented"))
}

// AddDealLineItem handles POST /deals/{dealID}/line-items
func (h *Handler) AddDealLineItem(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("add line item endpoint not yet implemented"))
}

// UpdateDealLineItem handles PUT /deals/{dealID}/line-items/{lineItemID}
func (h *Handler) UpdateDealLineItem(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("update line item endpoint not yet implemented"))
}

// RemoveDealLineItem handles DELETE /deals/{dealID}/line-items/{lineItemID}
func (h *Handler) RemoveDealLineItem(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("remove line item endpoint not yet implemented"))
}

// CreateDealInvoice handles POST /deals/{dealID}/invoices
func (h *Handler) CreateDealInvoice(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("create invoice endpoint not yet implemented"))
}

// UpdateDealInvoice handles PUT /deals/{dealID}/invoices/{invoiceID}
func (h *Handler) UpdateDealInvoice(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("update invoice endpoint not yet implemented"))
}

// IssueInvoice handles POST /deals/{dealID}/invoices/{invoiceID}/issue
func (h *Handler) IssueInvoice(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("issue invoice endpoint not yet implemented"))
}

// CancelInvoice handles POST /deals/{dealID}/invoices/{invoiceID}/cancel
func (h *Handler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("cancel invoice endpoint not yet implemented"))
}

// UpdatePayment handles PUT /deals/{dealID}/payments/{paymentID}
func (h *Handler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("update payment endpoint not yet implemented"))
}

// RefundPayment handles POST /deals/{dealID}/payments/{paymentID}/refund
func (h *Handler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("refund payment endpoint not yet implemented"))
}

// ============================================================================
// Helper Methods
// ============================================================================

// parseDealFilter parses deal filter from query parameters
func (h *Handler) parseDealFilter(r *http.Request) *dto.DealFilterRequest {
	q := r.URL.Query()

	filter := &dto.DealFilterRequest{
		SearchQuery: q.Get("search"),
	}

	// Parse statuses
	if statuses := q["status"]; len(statuses) > 0 {
		filter.Statuses = statuses
	}

	// Parse customer IDs
	if customerIDs := q["customer_id"]; len(customerIDs) > 0 {
		filter.CustomerIDs = customerIDs
	}

	// Parse owner IDs
	if ownerIDs := q["owner_id"]; len(ownerIDs) > 0 {
		filter.OwnerIDs = ownerIDs
	}

	// Parse pagination
	if page := q.Get("page"); page != "" {
		if p, err := parseInt(page); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if pageSize := q.Get("page_size"); pageSize != "" {
		if ps, err := parseInt(pageSize); err == nil && ps > 0 {
			filter.PageSize = ps
		}
	}

	// Parse sorting
	filter.SortBy = q.Get("sort_by")
	filter.SortOrder = q.Get("sort_order")

	return filter
}

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	var i int
	_, err := parseIntValue(s, &i)
	return i, err
}

// parseIntValue parses a string to an int pointer
func parseIntValue(s string, v *int) (bool, error) {
	if s == "" {
		return false, nil
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + int(c-'0')
	}
	*v = n
	return true, nil
}
