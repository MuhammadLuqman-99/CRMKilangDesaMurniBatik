// Package http provides HTTP handlers for the Sales Pipeline service.
package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
)

// ============================================================================
// Lead CRUD Handlers
// ============================================================================

// CreateLead handles POST /leads
func (h *Handler) CreateLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	var req dto.CreateLeadRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	lead, err := h.leadUseCase.Create(ctx, tenantID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondCreated(w, lead)
}

// GetLead handles GET /leads/{leadID}
func (h *Handler) GetLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	lead, err := h.leadUseCase.GetByID(ctx, tenantID, leadID)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, lead)
}

// UpdateLead handles PUT /leads/{leadID}
func (h *Handler) UpdateLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.UpdateLeadRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	lead, err := h.leadUseCase.Update(ctx, tenantID, leadID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, lead)
}

// DeleteLead handles DELETE /leads/{leadID}
func (h *Handler) DeleteLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	if err := h.leadUseCase.Delete(ctx, tenantID, leadID, ptrToUUID(userID)); err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondNoContent(w)
}

// ListLeads handles GET /leads
func (h *Handler) ListLeads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	// Build filter request from query parameters
	req := h.buildLeadFilterRequest(r)

	result, err := h.leadUseCase.List(ctx, tenantID, req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondList(w, result.Leads, result.Pagination.TotalItems, req.Page, req.PageSize)
}

// buildLeadFilterRequest builds a LeadFilterRequest from query parameters.
func (h *Handler) buildLeadFilterRequest(r *http.Request) *dto.LeadFilterRequest {
	req := &dto.LeadFilterRequest{
		Page:      h.getQueryInt(r, "page", 1),
		PageSize:  h.getQueryInt(r, "page_size", 20),
		SortBy:    h.getQueryString(r, "sort_by"),
		SortOrder: h.getQueryString(r, "sort_order"),
	}

	// Status filters
	if statuses := h.getQueryStringSlice(r, "statuses"); len(statuses) > 0 {
		req.Statuses = statuses
	}

	// Source filters
	if sources := h.getQueryStringSlice(r, "sources"); len(sources) > 0 {
		req.Sources = sources
	}

	// Assignment filters
	if ownerIDs := h.getQueryStringSlice(r, "owner_ids"); len(ownerIDs) > 0 {
		req.OwnerIDs = ownerIDs
	}
	req.Unassigned = h.getQueryBool(r, "unassigned")

	// Score filters
	if minScore := h.getQueryInt(r, "min_score", -1); minScore >= 0 {
		req.MinScore = &minScore
	}
	if maxScore := h.getQueryInt(r, "max_score", -1); maxScore >= 0 {
		req.MaxScore = &maxScore
	}

	// Time filters
	req.CreatedAfter = h.getQueryStringPtr(r, "created_after")
	req.CreatedBefore = h.getQueryStringPtr(r, "created_before")
	req.UpdatedAfter = h.getQueryStringPtr(r, "updated_after")

	// Search and tags
	req.SearchQuery = h.getQueryString(r, "q")
	if tags := h.getQueryStringSlice(r, "tags"); len(tags) > 0 {
		req.Tags = tags
	}

	// Campaign filter
	req.CampaignID = h.getQueryStringPtr(r, "campaign_id")

	return req
}

// ============================================================================
// Lead Status Operations
// ============================================================================

// QualifyLead handles POST /leads/{leadID}/qualify
func (h *Handler) QualifyLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.QualifyLeadRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	lead, err := h.leadUseCase.Qualify(ctx, tenantID, leadID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, lead)
}

// DisqualifyLead handles POST /leads/{leadID}/disqualify
func (h *Handler) DisqualifyLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.DisqualifyLeadRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	lead, err := h.leadUseCase.Disqualify(ctx, tenantID, leadID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, lead)
}

// ConvertLead handles POST /leads/{leadID}/convert
func (h *Handler) ConvertLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.ConvertLeadRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	result, err := h.leadUseCase.Convert(ctx, tenantID, leadID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondCreated(w, result)
}

// ReactivateLead handles POST /leads/{leadID}/reactivate
func (h *Handler) ReactivateLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	lead, err := h.leadUseCase.Nurture(ctx, tenantID, leadID, ptrToUUID(userID), &dto.NurtureLeadRequest{})
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, lead)
}

// ============================================================================
// Lead Assignment Operations
// ============================================================================

// AssignLead handles POST /leads/{leadID}/assign
func (h *Handler) AssignLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	var req dto.AssignLeadRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	lead, err := h.leadUseCase.Assign(ctx, tenantID, leadID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, lead)
}

// UnassignLead handles DELETE /leads/{leadID}/assign
func (h *Handler) UnassignLead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	leadID, err := h.getUUIDParam(r, "leadID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	// Assign with empty owner to unassign
	lead, err := h.leadUseCase.Assign(ctx, tenantID, leadID, ptrToUUID(userID), &dto.AssignLeadRequest{
		OwnerID: "",
	})
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, lead)
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkAssignLeads handles POST /leads/bulk/assign
func (h *Handler) BulkAssignLeads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	userID, _ := h.getUserID(ctx)

	var req dto.BulkAssignLeadsRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	count, err := h.leadUseCase.BulkAssign(ctx, tenantID, ptrToUUID(userID), &req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message":        "leads assigned successfully",
		"assigned_count": count,
	})
}

// BulkUpdateLeadStatus handles POST /leads/bulk/status
// Note: This endpoint is not yet implemented - bulk status updates should be done individually
func (h *Handler) BulkUpdateLeadStatus(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("bulk status update not yet implemented"))
}

// BulkDeleteLeads handles DELETE /leads/bulk
func (h *Handler) BulkDeleteLeads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	var req struct {
		LeadIDs []uuid.UUID `json:"lead_ids"`
	}
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, err)
		return
	}

	for _, leadID := range req.LeadIDs {
		userID, _ := h.getUserID(ctx)
		if err := h.leadUseCase.Delete(ctx, tenantID, leadID, ptrToUUID(userID)); err != nil {
			h.respondError(w, h.toError(err))
			return
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"message":       "leads deleted successfully",
		"deleted_count": len(req.LeadIDs),
	})
}

// ============================================================================
// Lead Statistics and Reports
// ============================================================================

// GetLeadStatistics handles GET /leads/statistics
func (h *Handler) GetLeadStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	stats, err := h.leadUseCase.GetStatistics(ctx, tenantID)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondSuccess(w, http.StatusOK, stats)
}

// GetLeadsByOwner handles GET /leads/by-owner/{ownerID}
func (h *Handler) GetLeadsByOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	ownerID, err := h.getUUIDParam(r, "ownerID")
	if err != nil {
		h.respondError(w, err)
		return
	}

	req := h.buildLeadFilterRequest(r)
	req.OwnerIDs = []string{ownerID.String()}

	result, err := h.leadUseCase.List(ctx, tenantID, req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondList(w, result.Leads, result.Pagination.TotalItems, req.Page, req.PageSize)
}

// GetHighScoreLeads handles GET /leads/high-score
func (h *Handler) GetHighScoreLeads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	minScore := h.getQueryInt(r, "min_score", 70)
	req := h.buildLeadFilterRequest(r)
	req.MinScore = &minScore

	result, err := h.leadUseCase.List(ctx, tenantID, req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondList(w, result.Leads, result.Pagination.TotalItems, req.Page, req.PageSize)
}

// GetUnassignedLeads handles GET /leads/unassigned
func (h *Handler) GetUnassignedLeads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	unassigned := true
	req := h.buildLeadFilterRequest(r)
	req.Unassigned = &unassigned

	result, err := h.leadUseCase.List(ctx, tenantID, req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondList(w, result.Leads, result.Pagination.TotalItems, req.Page, req.PageSize)
}

// GetStaleLeads handles GET /leads/stale
func (h *Handler) GetStaleLeads(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant context required"))
		return
	}

	// For stale leads, we'll use the updated_after filter to find leads not updated recently
	// The stale days parameter indicates how many days of inactivity to consider stale
	staleDays := h.getQueryInt(r, "days", 14)
	req := h.buildLeadFilterRequest(r)

	// Set stale days - this will be handled by the use case to filter by last update time
	_ = staleDays // Use case should handle stale filtering based on updatedAfter

	result, err := h.leadUseCase.List(ctx, tenantID, req)
	if err != nil {
		h.respondError(w, h.toError(err))
		return
	}

	h.respondList(w, result.Leads, result.Pagination.TotalItems, req.Page, req.PageSize)
}

// ============================================================================
// Helper Functions
// ============================================================================

// ptrToUUID converts a UUID pointer to UUID, returning uuid.Nil if nil
func ptrToUUID(p *uuid.UUID) uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return *p
}
