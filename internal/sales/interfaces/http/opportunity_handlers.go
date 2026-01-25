// Package http provides HTTP handlers for the Sales Pipeline service.
package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
)

// ============================================================================
// Opportunity CRUD Handlers
// ============================================================================

// CreateOpportunity handles POST /opportunities
func (h *Handler) CreateOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	var req dto.CreateOpportunityRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	if req.OwnerID == uuid.Nil && userID != uuid.Nil {
		req.OwnerID = userID
	}

	opportunity, err := h.opportunityUseCase.Create(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondCreated(w, opportunity)
}

// GetOpportunity handles GET /opportunities/{opportunityId}
func (h *Handler) GetOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// UpdateOpportunity handles PUT /opportunities/{opportunityId}
func (h *Handler) UpdateOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.UpdateOpportunityRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.UpdatedBy = userID

	opportunity, err := h.opportunityUseCase.Update(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// DeleteOpportunity handles DELETE /opportunities/{opportunityId}
func (h *Handler) DeleteOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	if err := h.opportunityUseCase.Delete(ctx, tenantID, opportunityID); err != nil {
		respondError(w, err)
		return
	}

	respondNoContent(w)
}

// ListOpportunities handles GET /opportunities
func (h *Handler) ListOpportunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	filter := buildOpportunityFilter(r)
	opts := buildListOptions(r)

	req := &dto.ListOpportunitiesRequest{
		TenantID: tenantID,
		Filter:   filter,
		Options:  opts,
	}

	opportunities, total, err := h.opportunityUseCase.List(ctx, req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondList(w, opportunities, total, opts.Page, opts.PageSize)
}

// ============================================================================
// Stage Operations
// ============================================================================

// MoveToStage handles POST /opportunities/{opportunityId}/stage
func (h *Handler) MoveToStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.MoveToStageRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.ChangedBy = userID

	opportunity, err := h.opportunityUseCase.MoveToStage(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// WinOpportunity handles POST /opportunities/{opportunityId}/win
func (h *Handler) WinOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.WinOpportunityRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.ClosedBy = userID

	result, err := h.opportunityUseCase.Win(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, result)
}

// LoseOpportunity handles POST /opportunities/{opportunityId}/lose
func (h *Handler) LoseOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.LoseOpportunityRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.ClosedBy = userID

	opportunity, err := h.opportunityUseCase.Lose(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// ReopenOpportunity handles POST /opportunities/{opportunityId}/reopen
func (h *Handler) ReopenOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.ReopenOpportunityRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.ReopenedBy = userID

	opportunity, err := h.opportunityUseCase.Reopen(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// GetStageHistory handles GET /opportunities/{opportunityId}/stage-history
func (h *Handler) GetStageHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	history, err := h.opportunityUseCase.GetStageHistory(ctx, tenantID, opportunityID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, history)
}

// ============================================================================
// Product Operations
// ============================================================================

// AddProduct handles POST /opportunities/{opportunityId}/products
func (h *Handler) AddOpportunityProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.AddOpportunityProductRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.AddedBy = userID

	opportunity, err := h.opportunityUseCase.AddProduct(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// UpdateProduct handles PUT /opportunities/{opportunityId}/products/{productId}
func (h *Handler) UpdateOpportunityProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	productID, err := getUUIDParam(r, "productId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.UpdateOpportunityProductRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.ProductID = productID
	req.UpdatedBy = userID

	opportunity, err := h.opportunityUseCase.UpdateProduct(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// RemoveProduct handles DELETE /opportunities/{opportunityId}/products/{productId}
func (h *Handler) RemoveOpportunityProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	productID, err := getUUIDParam(r, "productId")
	if err != nil {
		respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.RemoveProduct(ctx, tenantID, opportunityID, productID, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// ============================================================================
// Contact Operations
// ============================================================================

// AddContact handles POST /opportunities/{opportunityId}/contacts
func (h *Handler) AddOpportunityContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.AddOpportunityContactRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.AddedBy = userID

	opportunity, err := h.opportunityUseCase.AddContact(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// UpdateContact handles PUT /opportunities/{opportunityId}/contacts/{contactId}
func (h *Handler) UpdateOpportunityContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	contactID, err := getUUIDParam(r, "contactId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.UpdateOpportunityContactRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.OpportunityID = opportunityID
	req.ContactID = contactID
	req.UpdatedBy = userID

	opportunity, err := h.opportunityUseCase.UpdateContact(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// RemoveContact handles DELETE /opportunities/{opportunityId}/contacts/{contactId}
func (h *Handler) RemoveOpportunityContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	contactID, err := getUUIDParam(r, "contactId")
	if err != nil {
		respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.RemoveContact(ctx, tenantID, opportunityID, contactID, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// SetPrimaryContact handles POST /opportunities/{opportunityId}/contacts/{contactId}/primary
func (h *Handler) SetOpportunityPrimaryContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	contactID, err := getUUIDParam(r, "contactId")
	if err != nil {
		respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.SetPrimaryContact(ctx, tenantID, opportunityID, contactID, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// ============================================================================
// Competitor Operations
// ============================================================================

// AddCompetitor handles POST /opportunities/{opportunityId}/competitors
func (h *Handler) AddCompetitor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req struct {
		Competitor string `json:"competitor"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.SetCompetitor(ctx, tenantID, opportunityID, req.Competitor, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// ============================================================================
// Assignment Operations
// ============================================================================

// AssignOpportunity handles POST /opportunities/{opportunityId}/assign
func (h *Handler) AssignOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := getUUIDParam(r, "opportunityId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req struct {
		OwnerID uuid.UUID `json:"owner_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.Assign(ctx, tenantID, opportunityID, req.OwnerID, userID)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, opportunity)
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkAssignOpportunities handles POST /opportunities/bulk/assign
func (h *Handler) BulkAssignOpportunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	var req dto.BulkAssignOpportunitiesRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.AssignedBy = userID

	result, err := h.opportunityUseCase.BulkAssign(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, result)
}

// BulkMoveStage handles POST /opportunities/bulk/stage
func (h *Handler) BulkMoveStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	var req dto.BulkMoveStageRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	req.TenantID = tenantID
	req.ChangedBy = userID

	result, err := h.opportunityUseCase.BulkMoveStage(ctx, &req)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, result)
}

// ============================================================================
// Statistics and Analytics
// ============================================================================

// GetOpportunityStatistics handles GET /opportunities/statistics
func (h *Handler) GetOpportunityStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	pipelineID := getQueryUUID(r, "pipeline_id")
	startDate := getQueryTime(r, "start_date")
	endDate := getQueryTime(r, "end_date")

	stats, err := h.opportunityUseCase.GetStatistics(ctx, tenantID, pipelineID, startDate, endDate)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, stats)
}

// GetPipelineValue handles GET /opportunities/pipeline-value
func (h *Handler) GetPipelineValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	pipelineID := getQueryUUID(r, "pipeline_id")
	currency := getQueryString(r, "currency")
	if currency == "" {
		currency = "IDR"
	}

	value, err := h.opportunityUseCase.GetPipelineValue(ctx, tenantID, pipelineID, currency)
	if err != nil {
		respondError(w, err)
		return
	}

	respondSuccess(w, value)
}

// GetOpportunitiesByPipeline handles GET /opportunities/by-pipeline/{pipelineId}
func (h *Handler) GetOpportunitiesByPipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	pipelineID, err := getUUIDParam(r, "pipelineId")
	if err != nil {
		respondError(w, err)
		return
	}

	opts := buildListOptions(r)

	opportunities, total, err := h.opportunityUseCase.GetByPipeline(ctx, tenantID, pipelineID, &opts)
	if err != nil {
		respondError(w, err)
		return
	}

	respondList(w, opportunities, total, opts.Page, opts.PageSize)
}

// GetOpportunitiesByStage handles GET /opportunities/by-stage/{stageId}
func (h *Handler) GetOpportunitiesByStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	stageID, err := getUUIDParam(r, "stageId")
	if err != nil {
		respondError(w, err)
		return
	}

	pipelineID := getQueryUUID(r, "pipeline_id")
	if pipelineID == nil {
		respondError(w, ErrMissingParameter("pipeline_id"))
		return
	}

	opts := buildListOptions(r)

	opportunities, total, err := h.opportunityUseCase.GetByStage(ctx, tenantID, *pipelineID, stageID, &opts)
	if err != nil {
		respondError(w, err)
		return
	}

	respondList(w, opportunities, total, opts.Page, opts.PageSize)
}

// GetClosingThisMonth handles GET /opportunities/closing-this-month
func (h *Handler) GetClosingThisMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opts := buildListOptions(r)

	opportunities, total, err := h.opportunityUseCase.GetClosingThisMonth(ctx, tenantID, &opts)
	if err != nil {
		respondError(w, err)
		return
	}

	respondList(w, opportunities, total, opts.Page, opts.PageSize)
}

// GetOverdueOpportunities handles GET /opportunities/overdue
func (h *Handler) GetOverdueOpportunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opts := buildListOptions(r)

	opportunities, total, err := h.opportunityUseCase.GetOverdue(ctx, tenantID, &opts)
	if err != nil {
		respondError(w, err)
		return
	}

	respondList(w, opportunities, total, opts.Page, opts.PageSize)
}
