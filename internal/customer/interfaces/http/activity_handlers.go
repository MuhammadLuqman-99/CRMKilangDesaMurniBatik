// Package http provides HTTP handlers for the Customer service.
package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/usecase"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Activity Handlers
// ============================================================================

// LogActivity handles POST /api/v1/customers/{customerId}/activities
func (h *Handler) LogActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.CreateActivityRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.LogActivityInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Request:    &req,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	activity, err := h.logActivity.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    activity,
	})
}

// GetActivity handles GET /api/v1/customers/{customerId}/activities/{activityId}
func (h *Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	activityID, err := getUUIDParam(r, "activityId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.GetActivityInput{
		TenantID:   tenantID,
		CustomerID: customerID,
		ActivityID: activityID,
	}

	activity, err := h.getActivity.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    activity,
	})
}

// ListActivities handles GET /api/v1/customers/{customerId}/activities
func (h *Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	// Parse activity types from query
	var activityTypes []domain.ActivityType
	if types := getQueryStringSlice(r, "types"); len(types) > 0 {
		for _, t := range types {
			activityTypes = append(activityTypes, domain.ActivityType(t))
		}
	}

	input := usecase.ListActivitiesInput{
		TenantID:      tenantID,
		CustomerID:    customerID,
		ActivityTypes: activityTypes,
		PerformedBy:   getQueryUUID(r, "performed_by"),
		StartDate:     getQueryTime(r, "start_date"),
		EndDate:       getQueryTime(r, "end_date"),
		Offset:        getQueryInt(r, "offset", 0),
		Limit:         getQueryInt(r, "limit", 20),
		SortBy:        getQueryString(r, "sort_by"),
		SortOrder:     getQueryString(r, "sort_order"),
	}

	result, err := h.listActivities.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Activities,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}
