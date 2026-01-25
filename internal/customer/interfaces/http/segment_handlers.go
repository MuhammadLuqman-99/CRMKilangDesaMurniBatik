// Package http provides HTTP handlers for the Customer service.
package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/usecase"
)

// ============================================================================
// Segment CRUD Handlers
// ============================================================================

// CreateSegment handles POST /api/v1/segments
func (h *Handler) CreateSegment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	var req dto.CreateSegmentRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.CreateSegmentInput{
		TenantID:  tenantID,
		UserID:    userID,
		Request:   &req,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	segment, err := h.createSegment.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    segment,
	})
}

// GetSegment handles GET /api/v1/segments/{segmentId}
func (h *Handler) GetSegment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	segmentID, err := getUUIDParam(r, "segmentId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.GetSegmentInput{
		TenantID:  tenantID,
		SegmentID: segmentID,
	}

	segment, err := h.getSegment.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    segment,
	})
}

// UpdateSegment handles PUT /api/v1/segments/{segmentId}
func (h *Handler) UpdateSegment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	segmentID, err := getUUIDParam(r, "segmentId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.UpdateSegmentRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.UpdateSegmentInput{
		TenantID:  tenantID,
		UserID:    userID,
		SegmentID: segmentID,
		Request:   &req,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	segment, err := h.updateSegment.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    segment,
	})
}

// DeleteSegment handles DELETE /api/v1/segments/{segmentId}
func (h *Handler) DeleteSegment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	segmentID, err := getUUIDParam(r, "segmentId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.DeleteSegmentInput{
		TenantID:  tenantID,
		UserID:    userID,
		SegmentID: segmentID,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	err = h.deleteSegment.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "segment deleted"},
	})
}

// ============================================================================
// Segment List Handlers
// ============================================================================

// ListSegments handles GET /api/v1/segments
func (h *Handler) ListSegments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	input := usecase.ListSegmentsInput{
		TenantID:   tenantID,
		ActiveOnly: getQueryBool(r, "active_only"),
		Type:       getQueryString(r, "type"),
		Offset:     getQueryInt(r, "offset", 0),
		Limit:      getQueryInt(r, "limit", 20),
		SortBy:     getQueryString(r, "sort_by"),
		SortOrder:  getQueryString(r, "sort_order"),
	}

	result, err := h.listSegments.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Segments,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// ============================================================================
// Segment Action Handlers
// ============================================================================

// RefreshSegment handles POST /api/v1/segments/{segmentId}/refresh
func (h *Handler) RefreshSegment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	segmentID, err := getUUIDParam(r, "segmentId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.RefreshSegmentInput{
		TenantID:  tenantID,
		UserID:    userID,
		SegmentID: segmentID,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	segment, err := h.refreshSegment.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    segment,
	})
}

// GetSegmentCustomers handles GET /api/v1/segments/{segmentId}/customers
func (h *Handler) GetSegmentCustomers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	segmentID, err := getUUIDParam(r, "segmentId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.GetSegmentCustomersInput{
		TenantID:  tenantID,
		SegmentID: segmentID,
		Offset:    getQueryInt(r, "offset", 0),
		Limit:     getQueryInt(r, "limit", 20),
		SortBy:    getQueryString(r, "sort_by"),
		SortOrder: getQueryString(r, "sort_order"),
	}

	result, err := h.getSegmentCustomers.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Customers,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}
