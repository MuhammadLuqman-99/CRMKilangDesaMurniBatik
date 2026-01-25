// Package http provides HTTP handlers for the Sales Pipeline service.
package http

import (
	"net/http"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
)

// ============================================================================
// Opportunity CRUD Handlers
// ============================================================================

// CreateOpportunity handles POST /opportunities
func (h *Handler) CreateOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	var req dto.CreateOpportunityRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.Create(ctx, tenantID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondCreated(w, opportunity)
}

// GetOpportunity handles GET /opportunities/{opportunityId}
func (h *Handler) GetOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// UpdateOpportunity handles PUT /opportunities/{opportunityId}
func (h *Handler) UpdateOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.UpdateOpportunityRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.Update(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// DeleteOpportunity handles DELETE /opportunities/{opportunityId}
func (h *Handler) DeleteOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	if err := h.opportunityUseCase.Delete(ctx, tenantID, opportunityID, ptrToUUID(userID)); err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondNoContent(w)
}

// ListOpportunities handles GET /opportunities
func (h *Handler) ListOpportunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	req := h.buildOpportunityFilterRequest(r)

	result, err := h.opportunityUseCase.List(ctx, tenantID, req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondList(w, result.Opportunities, result.Pagination.TotalItems, req.Page, req.PageSize)
}

// buildOpportunityFilterRequest builds an OpportunityFilterRequest from query parameters.
func (h *Handler) buildOpportunityFilterRequest(r *http.Request) *dto.OpportunityFilterRequest {
	req := &dto.OpportunityFilterRequest{
		Page:      h.getQueryInt(r, "page", 1),
		PageSize:  h.getQueryInt(r, "page_size", 20),
		SortBy:    h.getQueryString(r, "sort_by"),
		SortOrder: h.getQueryString(r, "sort_order"),
	}

	// Status filters
	if statuses := h.getQueryStringSlice(r, "statuses"); len(statuses) > 0 {
		req.Statuses = statuses
	}

	// Pipeline filters
	req.PipelineIDs = h.getQueryStringSlice(r, "pipeline_ids")
	req.StageIDs = h.getQueryStringSlice(r, "stage_ids")

	// Relationship filters
	req.CustomerIDs = h.getQueryStringSlice(r, "customer_ids")
	req.OwnerIDs = h.getQueryStringSlice(r, "owner_ids")

	// Value filters
	if minAmount := h.getQueryInt64(r, "min_amount"); minAmount != nil {
		req.MinAmount = minAmount
	}
	if maxAmount := h.getQueryInt64(r, "max_amount"); maxAmount != nil {
		req.MaxAmount = maxAmount
	}

	// Search
	req.SearchQuery = h.getQueryString(r, "q")

	return req
}

// ============================================================================
// Stage Operations
// ============================================================================

// MoveOpportunityStage handles POST /opportunities/{opportunityId}/stage
func (h *Handler) MoveOpportunityStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.MoveStageRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.MoveStage(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// WinOpportunity handles POST /opportunities/{opportunityId}/win
func (h *Handler) WinOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.WinOpportunityRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	result, err := h.opportunityUseCase.Win(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, result)
}

// LoseOpportunity handles POST /opportunities/{opportunityId}/lose
func (h *Handler) LoseOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.LoseOpportunityRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	result, err := h.opportunityUseCase.Lose(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, result)
}

// ReopenOpportunity handles POST /opportunities/{opportunityId}/reopen
func (h *Handler) ReopenOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.ReopenOpportunityRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.Reopen(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// ============================================================================
// Product Operations
// ============================================================================

// AddOpportunityProduct handles POST /opportunities/{opportunityId}/products
func (h *Handler) AddOpportunityProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.AddProductRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.AddProduct(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// UpdateOpportunityProduct handles PUT /opportunities/{opportunityId}/products/{productId}
func (h *Handler) UpdateOpportunityProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	productID, err := h.getUUIDParam(r, "productId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.UpdateProductRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.UpdateProduct(ctx, tenantID, opportunityID, productID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// RemoveOpportunityProduct handles DELETE /opportunities/{opportunityId}/products/{productId}
func (h *Handler) RemoveOpportunityProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	productID, err := h.getUUIDParam(r, "productId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.RemoveProduct(ctx, tenantID, opportunityID, productID, ptrToUUID(userID))
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// ============================================================================
// Contact Operations
// ============================================================================

// AddOpportunityContact handles POST /opportunities/{opportunityId}/contacts
func (h *Handler) AddOpportunityContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.AddContactRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.AddContact(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// UpdateOpportunityContact handles PUT /opportunities/{opportunityId}/contacts/{contactId}
func (h *Handler) UpdateOpportunityContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	contactID, err := h.getUUIDParam(r, "contactId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.UpdateContactRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.UpdateContact(ctx, tenantID, opportunityID, contactID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// RemoveOpportunityContact handles DELETE /opportunities/{opportunityId}/contacts/{contactId}
func (h *Handler) RemoveOpportunityContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	contactID, err := h.getUUIDParam(r, "contactId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.RemoveContact(ctx, tenantID, opportunityID, contactID, ptrToUUID(userID))
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// ============================================================================
// Competitor Operations
// ============================================================================

// AddOpportunityCompetitor handles POST /opportunities/{opportunityId}/competitors
func (h *Handler) AddOpportunityCompetitor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.AddCompetitorRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.AddCompetitor(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// ============================================================================
// Assignment Operations
// ============================================================================

// AssignOpportunity handles POST /opportunities/{opportunityId}/assign
func (h *Handler) AssignOpportunity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	opportunityID, err := h.getUUIDParam(r, "opportunityId")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.AssignOpportunityRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	opportunity, err := h.opportunityUseCase.Assign(ctx, tenantID, opportunityID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, opportunity)
}

// BulkAssignOpportunities handles POST /opportunities/bulk/assign
func (h *Handler) BulkAssignOpportunities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	var req dto.BulkAssignOpportunitiesRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	if err := h.opportunityUseCase.BulkAssign(ctx, tenantID, ptrToUUID(userID), &req); err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message": "opportunities assigned successfully",
	})
}

// ============================================================================
// Statistics
// ============================================================================

// GetOpportunityStatistics handles GET /opportunities/statistics
func (h *Handler) GetOpportunityStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	stats, err := h.opportunityUseCase.GetStatistics(ctx, tenantID)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, stats)
}

// GetPipelineValue handles GET /opportunities/pipeline-value
func (h *Handler) GetPipelineValue(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("pipeline value endpoint not yet implemented"))
}

// GetClosingThisMonth handles GET /opportunities/closing-this-month
func (h *Handler) GetClosingThisMonth(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("closing this month endpoint not yet implemented"))
}

// GetOverdueOpportunities handles GET /opportunities/overdue
func (h *Handler) GetOverdueOpportunities(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("overdue opportunities endpoint not yet implemented"))
}

// BulkMoveStage handles POST /opportunities/bulk/move-stage
func (h *Handler) BulkMoveStage(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("bulk move stage endpoint not yet implemented"))
}

// MoveOpportunityToStage handles POST /opportunities/{opportunityId}/move-stage
func (h *Handler) MoveOpportunityToStage(w http.ResponseWriter, r *http.Request) {
	// Delegate to MoveOpportunityStage
	h.MoveOpportunityStage(w, r)
}

// GetOpportunityStageHistory handles GET /opportunities/{opportunityId}/stage-history
func (h *Handler) GetOpportunityStageHistory(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("stage history endpoint not yet implemented"))
}

// SetOpportunityPrimaryContact handles POST /opportunities/{opportunityId}/contacts/{contactId}/set-primary
func (h *Handler) SetOpportunityPrimaryContact(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("set primary contact endpoint not yet implemented"))
}
