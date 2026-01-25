package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
)

// ============================================================================
// Pipeline Handler Methods
// ============================================================================

// CreatePipeline handles POST /pipelines
func (h *Handler) CreatePipeline(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreatePipelineRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	pipeline, err := h.pipelineUseCase.Create(ctx, tenantID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, pipeline)
}

// GetPipeline handles GET /pipelines/{pipelineID}
func (h *Handler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	pipeline, err := h.pipelineUseCase.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// UpdatePipeline handles PUT /pipelines/{pipelineID}
func (h *Handler) UpdatePipeline(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	var req dto.UpdatePipelineRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	pipeline, err := h.pipelineUseCase.Update(ctx, tenantID, pipelineID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// DeletePipeline handles DELETE /pipelines/{pipelineID}
func (h *Handler) DeletePipeline(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	if err := h.pipelineUseCase.Delete(ctx, tenantID, pipelineID, userID); err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusNoContent, nil)
}

// ListPipelines handles GET /pipelines
func (h *Handler) ListPipelines(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	filter := h.parsePipelineFilter(r)

	pipelines, err := h.pipelineUseCase.List(ctx, tenantID, filter)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipelines)
}

// GetDefaultPipeline handles GET /pipelines/default
func (h *Handler) GetDefaultPipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipeline, err := h.pipelineUseCase.GetDefault(ctx, tenantID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// SetDefaultPipeline handles POST /pipelines/{pipelineID}/set-default
func (h *Handler) SetDefaultPipeline(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	pipeline, err := h.pipelineUseCase.SetDefault(ctx, tenantID, pipelineID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// ActivatePipeline handles POST /pipelines/{pipelineID}/activate
func (h *Handler) ActivatePipeline(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	pipeline, err := h.pipelineUseCase.Activate(ctx, tenantID, pipelineID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// DeactivatePipeline handles POST /pipelines/{pipelineID}/deactivate
func (h *Handler) DeactivatePipeline(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	pipeline, err := h.pipelineUseCase.Deactivate(ctx, tenantID, pipelineID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// ClonePipeline handles POST /pipelines/{pipelineID}/clone
func (h *Handler) ClonePipeline(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	sourcePipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	var req dto.ClonePipelineRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	req.SourcePipelineID = sourcePipelineID.String()

	pipeline, err := h.pipelineUseCase.Clone(ctx, tenantID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, pipeline)
}

// ============================================================================
// Stage Operations
// ============================================================================

// AddStage handles POST /pipelines/{pipelineID}/stages
func (h *Handler) AddStage(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	var req dto.AddStageRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	pipeline, err := h.pipelineUseCase.AddStage(ctx, tenantID, pipelineID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// UpdateStage handles PUT /pipelines/{pipelineID}/stages/{stageID}
func (h *Handler) UpdateStage(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	stageIDStr := chi.URLParam(r, "stageID")
	stageID, err := uuid.Parse(stageIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("stageID", "invalid UUID format"))
		return
	}

	var req dto.UpdateStageRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	pipeline, err := h.pipelineUseCase.UpdateStage(ctx, tenantID, pipelineID, stageID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// RemoveStage handles DELETE /pipelines/{pipelineID}/stages/{stageID}
func (h *Handler) RemoveStage(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	stageIDStr := chi.URLParam(r, "stageID")
	stageID, err := uuid.Parse(stageIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("stageID", "invalid UUID format"))
		return
	}

	pipeline, err := h.pipelineUseCase.RemoveStage(ctx, tenantID, pipelineID, stageID, userID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// ReorderStages handles POST /pipelines/{pipelineID}/stages/reorder
func (h *Handler) ReorderStages(w http.ResponseWriter, r *http.Request) {
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

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	var req dto.ReorderStagesRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	pipeline, err := h.pipelineUseCase.ReorderStages(ctx, tenantID, pipelineID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// ============================================================================
// Analytics & Statistics
// ============================================================================

// GetPipelineStatistics handles GET /pipelines/{pipelineID}/statistics
func (h *Handler) GetPipelineStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineIDStr := chi.URLParam(r, "pipelineID")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, ErrInvalidParameter("pipelineID", "invalid UUID format"))
		return
	}

	stats, err := h.pipelineUseCase.GetStatistics(ctx, tenantID, pipelineID)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// ComparePipelines handles POST /pipelines/compare
func (h *Handler) ComparePipelines(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	var req dto.PipelineComparisonRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	comparison, err := h.pipelineUseCase.ComparePipelines(ctx, tenantID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, comparison)
}

// GetForecast handles POST /pipelines/forecast
func (h *Handler) GetForecast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	var req dto.ForecastRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	forecast, err := h.pipelineUseCase.GetForecast(ctx, tenantID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, forecast)
}

// ============================================================================
// Templates
// ============================================================================

// GetPipelineTemplates handles GET /pipelines/templates
func (h *Handler) GetPipelineTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	templates, err := h.pipelineUseCase.GetTemplates(ctx)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, templates)
}

// CreatePipelineFromTemplate handles POST /pipelines/from-template
func (h *Handler) CreatePipelineFromTemplate(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateFromTemplateRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrInvalidJSON(err.Error()))
		return
	}

	pipeline, err := h.pipelineUseCase.CreateFromTemplate(ctx, tenantID, userID, &req)
	if err != nil {
		h.respondError(w, toHTTPError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, pipeline)
}

// GetPipelinesByType handles GET /pipelines/by-type/{type}
func (h *Handler) GetPipelinesByType(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("pipelines by type endpoint not yet implemented"))
}

// GetPipelineVelocity handles GET /pipelines/{pipelineID}/velocity
func (h *Handler) GetPipelineVelocity(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("pipeline velocity endpoint not yet implemented"))
}

// GetStageConversionRates handles GET /pipelines/{pipelineID}/conversion-rates
func (h *Handler) GetStageConversionRates(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("stage conversion rates endpoint not yet implemented"))
}

// AddPipelineStage handles POST /pipelines/{pipelineID}/stages
func (h *Handler) AddPipelineStage(w http.ResponseWriter, r *http.Request) {
	h.AddStage(w, r)
}

// ReorderPipelineStages handles PUT /pipelines/{pipelineID}/stages/reorder
func (h *Handler) ReorderPipelineStages(w http.ResponseWriter, r *http.Request) {
	h.ReorderStages(w, r)
}

// UpdatePipelineStage handles PUT /pipelines/{pipelineID}/stages/{stageID}
func (h *Handler) UpdatePipelineStage(w http.ResponseWriter, r *http.Request) {
	h.UpdateStage(w, r)
}

// RemovePipelineStage handles DELETE /pipelines/{pipelineID}/stages/{stageID}
func (h *Handler) RemovePipelineStage(w http.ResponseWriter, r *http.Request) {
	h.RemoveStage(w, r)
}

// ActivatePipelineStage handles POST /pipelines/{pipelineID}/stages/{stageID}/activate
func (h *Handler) ActivatePipelineStage(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("activate stage endpoint not yet implemented"))
}

// DeactivatePipelineStage handles POST /pipelines/{pipelineID}/stages/{stageID}/deactivate
func (h *Handler) DeactivatePipelineStage(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, ErrUnprocessableEntity("deactivate stage endpoint not yet implemented"))
}

// ============================================================================
// Helper Methods
// ============================================================================

// parsePipelineFilter parses pipeline filter from query parameters
func (h *Handler) parsePipelineFilter(r *http.Request) *dto.PipelineFilterRequest {
	q := r.URL.Query()

	filter := &dto.PipelineFilterRequest{
		SearchQuery: q.Get("search"),
	}

	// Parse is_active
	if isActive := q.Get("is_active"); isActive != "" {
		if isActive == "true" {
			b := true
			filter.IsActive = &b
		} else if isActive == "false" {
			b := false
			filter.IsActive = &b
		}
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
