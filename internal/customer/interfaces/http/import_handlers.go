// Package http provides HTTP handlers for the Customer service.
package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/usecase"
)

// ============================================================================
// Import Handlers
// ============================================================================

// GetImportStatus handles GET /api/v1/imports/{importId}
func (h *Handler) GetImportStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	importID, err := getUUIDParam(r, "importId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.GetImportStatusInput{
		TenantID: tenantID,
		ImportID: importID,
	}

	status, err := h.getImportStatus.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    status,
	})
}

// GetImportErrors handles GET /api/v1/imports/{importId}/errors
func (h *Handler) GetImportErrors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	importID, err := getUUIDParam(r, "importId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.GetImportErrorsInput{
		TenantID: tenantID,
		ImportID: importID,
		Offset:   getQueryInt(r, "offset", 0),
		Limit:    getQueryInt(r, "limit", 50),
	}

	result, err := h.getImportErrors.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Errors,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// ListImports handles GET /api/v1/imports
func (h *Handler) ListImports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	input := usecase.ListImportsInput{
		TenantID:  tenantID,
		Status:    getQueryString(r, "status"),
		StartDate: getQueryTime(r, "start_date"),
		EndDate:   getQueryTime(r, "end_date"),
		Offset:    getQueryInt(r, "offset", 0),
		Limit:     getQueryInt(r, "limit", 20),
		SortBy:    getQueryString(r, "sort_by"),
		SortOrder: getQueryString(r, "sort_order"),
	}

	result, err := h.listImports.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Imports,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// CancelImport handles DELETE /api/v1/imports/{importId}
func (h *Handler) CancelImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	importID, err := getUUIDParam(r, "importId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.CancelImportInput{
		TenantID:  tenantID,
		UserID:    userID,
		ImportID:  importID,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	err = h.cancelImport.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "import cancelled"},
	})
}
