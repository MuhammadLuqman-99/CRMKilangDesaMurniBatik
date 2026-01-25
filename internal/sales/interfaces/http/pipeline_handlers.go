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
// Pipeline Request/Response DTOs
// ============================================================================

// CreatePipelineRequest represents a request to create a new pipeline
type CreatePipelineRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Type        string              `json:"type,omitempty"`
	Currency    string              `json:"currency,omitempty"`
	IsDefault   bool                `json:"is_default,omitempty"`
	Stages      []CreateStageInput  `json:"stages,omitempty"`
	Metadata    map[string]any      `json:"metadata,omitempty"`
}

// UpdatePipelineRequest represents a request to update a pipeline
type UpdatePipelineRequest struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Type        *string        `json:"type,omitempty"`
	Currency    *string        `json:"currency,omitempty"`
	IsDefault   *bool          `json:"is_default,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CreateStageInput represents input for creating a pipeline stage
type CreateStageInput struct {
	Name             string         `json:"name"`
	Description      string         `json:"description,omitempty"`
	Color            string         `json:"color,omitempty"`
	Probability      int            `json:"probability"`
	RottenDays       int            `json:"rotten_days,omitempty"`
	IsWinStage       bool           `json:"is_win_stage,omitempty"`
	IsLostStage      bool           `json:"is_lost_stage,omitempty"`
	RequiredFields   []string       `json:"required_fields,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// AddStageRequest represents a request to add a stage to a pipeline
type AddStageRequest struct {
	Name             string         `json:"name"`
	Description      string         `json:"description,omitempty"`
	Color            string         `json:"color,omitempty"`
	Probability      int            `json:"probability"`
	Position         int            `json:"position,omitempty"`
	RottenDays       int            `json:"rotten_days,omitempty"`
	IsWinStage       bool           `json:"is_win_stage,omitempty"`
	IsLostStage      bool           `json:"is_lost_stage,omitempty"`
	RequiredFields   []string       `json:"required_fields,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// UpdateStageRequest represents a request to update a pipeline stage
type UpdateStageRequest struct {
	Name             *string        `json:"name,omitempty"`
	Description      *string        `json:"description,omitempty"`
	Color            *string        `json:"color,omitempty"`
	Probability      *int           `json:"probability,omitempty"`
	RottenDays       *int           `json:"rotten_days,omitempty"`
	IsWinStage       *bool          `json:"is_win_stage,omitempty"`
	IsLostStage      *bool          `json:"is_lost_stage,omitempty"`
	RequiredFields   []string       `json:"required_fields,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// ReorderStagesRequest represents a request to reorder pipeline stages
type ReorderStagesRequest struct {
	StageIDs []string `json:"stage_ids"`
}

// ClonePipelineRequest represents a request to clone a pipeline
type ClonePipelineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// PipelineResponse represents a pipeline in API responses
type PipelineResponse struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type"`
	Currency    string          `json:"currency"`
	IsDefault   bool            `json:"is_default"`
	IsActive    bool            `json:"is_active"`
	Stages      []StageResponse `json:"stages,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// StageResponse represents a pipeline stage in API responses
type StageResponse struct {
	ID             string         `json:"id"`
	PipelineID     string         `json:"pipeline_id"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	Color          string         `json:"color,omitempty"`
	Position       int            `json:"position"`
	Probability    int            `json:"probability"`
	RottenDays     int            `json:"rotten_days"`
	IsWinStage     bool           `json:"is_win_stage"`
	IsLostStage    bool           `json:"is_lost_stage"`
	IsActive       bool           `json:"is_active"`
	RequiredFields []string       `json:"required_fields,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
}

// PipelineStatisticsResponse represents pipeline statistics
type PipelineStatisticsResponse struct {
	PipelineID           string                   `json:"pipeline_id"`
	PipelineName         string                   `json:"pipeline_name"`
	TotalOpportunities   int64                    `json:"total_opportunities"`
	TotalValue           string                   `json:"total_value"`
	WonOpportunities     int64                    `json:"won_opportunities"`
	LostOpportunities    int64                    `json:"lost_opportunities"`
	WonValue             string                   `json:"won_value"`
	WinRate              string                   `json:"win_rate"`
	AverageDealSize      string                   `json:"average_deal_size"`
	AverageDaysToClose   int                      `json:"average_days_to_close"`
	ConversionRate       string                   `json:"conversion_rate"`
	StageStatistics      []StageStatisticsItem    `json:"stage_statistics,omitempty"`
}

// StageStatisticsItem represents statistics for a single stage
type StageStatisticsItem struct {
	StageID          string `json:"stage_id"`
	StageName        string `json:"stage_name"`
	OpportunityCount int64  `json:"opportunity_count"`
	TotalValue       string `json:"total_value"`
	AverageValue     string `json:"average_value"`
	AverageDaysInStage int  `json:"average_days_in_stage"`
	ConversionRate   string `json:"conversion_rate"`
}

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

	var req CreatePipelineRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	// Build stages input
	stages := make([]command.CreatePipelineStageInput, len(req.Stages))
	for i, s := range req.Stages {
		stages[i] = command.CreatePipelineStageInput{
			Name:           s.Name,
			Description:    s.Description,
			Color:          s.Color,
			Position:       i + 1,
			Probability:    s.Probability,
			RottenDays:     s.RottenDays,
			IsWinStage:     s.IsWinStage,
			IsLostStage:    s.IsLostStage,
			RequiredFields: s.RequiredFields,
			Metadata:       s.Metadata,
		}
	}

	pipelineType := entity.PipelineTypeSales
	if req.Type != "" {
		pipelineType = entity.PipelineType(req.Type)
	}

	currency := "IDR"
	if req.Currency != "" {
		currency = req.Currency
	}

	cmd := command.CreatePipelineCommand{
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Type:        pipelineType,
		Currency:    currency,
		IsDefault:   req.IsDefault,
		Stages:      stages,
		Metadata:    req.Metadata,
	}

	pipeline, err := h.pipelineUseCase.CreatePipeline(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toPipelineResponse(pipeline))
}

// GetPipeline handles GET /pipelines/{pipelineID}
func (h *Handler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	qry := query.GetPipelineQuery{
		TenantID:   tenantID,
		PipelineID: pipelineID,
	}

	pipeline, err := h.pipelineUseCase.GetPipeline(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// UpdatePipeline handles PUT /pipelines/{pipelineID}
func (h *Handler) UpdatePipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	var req UpdatePipelineRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	cmd := command.UpdatePipelineCommand{
		TenantID:    tenantID,
		PipelineID:  pipelineID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Currency:    req.Currency,
		IsDefault:   req.IsDefault,
		Metadata:    req.Metadata,
	}

	pipeline, err := h.pipelineUseCase.UpdatePipeline(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// DeletePipeline handles DELETE /pipelines/{pipelineID}
func (h *Handler) DeletePipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	cmd := command.DeletePipelineCommand{
		TenantID:   tenantID,
		PipelineID: pipelineID,
	}

	if err := h.pipelineUseCase.DeletePipeline(ctx, cmd); err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondSuccess(w, http.StatusNoContent, nil)
}

// ListPipelines handles GET /pipelines
func (h *Handler) ListPipelines(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	// Parse filter parameters
	pipelineType := r.URL.Query().Get("type")
	isActive := r.URL.Query().Get("is_active")
	opts := h.buildListOptions(r)

	filter := query.PipelineFilter{}
	if pipelineType != "" {
		t := entity.PipelineType(pipelineType)
		filter.Type = &t
	}
	if isActive != "" {
		active := isActive == "true"
		filter.IsActive = &active
	}

	qry := query.ListPipelinesQuery{
		TenantID:    tenantID,
		Filter:      filter,
		ListOptions: opts,
	}

	pipelines, total, err := h.pipelineUseCase.ListPipelines(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	responses := make([]PipelineResponse, len(pipelines))
	for i, pipeline := range pipelines {
		responses[i] = h.toPipelineResponse(pipeline)
	}

	h.respondList(w, responses, total, opts.Page, opts.PageSize)
}

// ============================================================================
// Pipeline Status Operations
// ============================================================================

// ActivatePipeline handles POST /pipelines/{pipelineID}/activate
func (h *Handler) ActivatePipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	cmd := command.ActivatePipelineCommand{
		TenantID:   tenantID,
		PipelineID: pipelineID,
	}

	pipeline, err := h.pipelineUseCase.ActivatePipeline(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// DeactivatePipeline handles POST /pipelines/{pipelineID}/deactivate
func (h *Handler) DeactivatePipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	cmd := command.DeactivatePipelineCommand{
		TenantID:   tenantID,
		PipelineID: pipelineID,
	}

	pipeline, err := h.pipelineUseCase.DeactivatePipeline(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// SetDefaultPipeline handles POST /pipelines/{pipelineID}/set-default
func (h *Handler) SetDefaultPipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	cmd := command.SetDefaultPipelineCommand{
		TenantID:   tenantID,
		PipelineID: pipelineID,
	}

	pipeline, err := h.pipelineUseCase.SetDefault(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// ClonePipeline handles POST /pipelines/{pipelineID}/clone
func (h *Handler) ClonePipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	var req ClonePipelineRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	cmd := command.ClonePipelineCommand{
		TenantID:         tenantID,
		SourcePipelineID: pipelineID,
		Name:             req.Name,
		Description:      req.Description,
	}

	pipeline, err := h.pipelineUseCase.ClonePipeline(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toPipelineResponse(pipeline))
}

// ============================================================================
// Stage Operations
// ============================================================================

// AddPipelineStage handles POST /pipelines/{pipelineID}/stages
func (h *Handler) AddPipelineStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	var req AddStageRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	cmd := command.AddPipelineStageCommand{
		TenantID:       tenantID,
		PipelineID:     pipelineID,
		Name:           req.Name,
		Description:    req.Description,
		Color:          req.Color,
		Position:       req.Position,
		Probability:    req.Probability,
		RottenDays:     req.RottenDays,
		IsWinStage:     req.IsWinStage,
		IsLostStage:    req.IsLostStage,
		RequiredFields: req.RequiredFields,
		Metadata:       req.Metadata,
	}

	pipeline, err := h.pipelineUseCase.AddStage(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusCreated, h.toPipelineResponse(pipeline))
}

// UpdatePipelineStage handles PUT /pipelines/{pipelineID}/stages/{stageID}
func (h *Handler) UpdatePipelineStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	stageID, err := uuid.Parse(chi.URLParam(r, "stageID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid stage ID"))
		return
	}

	var req UpdateStageRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	cmd := command.UpdatePipelineStageCommand{
		TenantID:       tenantID,
		PipelineID:     pipelineID,
		StageID:        stageID,
		Name:           req.Name,
		Description:    req.Description,
		Color:          req.Color,
		Probability:    req.Probability,
		RottenDays:     req.RottenDays,
		IsWinStage:     req.IsWinStage,
		IsLostStage:    req.IsLostStage,
		RequiredFields: req.RequiredFields,
		Metadata:       req.Metadata,
	}

	pipeline, err := h.pipelineUseCase.UpdateStage(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// RemovePipelineStage handles DELETE /pipelines/{pipelineID}/stages/{stageID}
func (h *Handler) RemovePipelineStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	stageID, err := uuid.Parse(chi.URLParam(r, "stageID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid stage ID"))
		return
	}

	// Check if migration stage is specified
	migrateToStageID := r.URL.Query().Get("migrate_to_stage")
	var migrateToStage *uuid.UUID
	if migrateToStageID != "" {
		id, err := uuid.Parse(migrateToStageID)
		if err != nil {
			h.respondError(w, ErrValidation("migrate_to_stage", "invalid UUID format"))
			return
		}
		migrateToStage = &id
	}

	cmd := command.RemovePipelineStageCommand{
		TenantID:              tenantID,
		PipelineID:            pipelineID,
		StageID:               stageID,
		MigrateOpportunitiesTo: migrateToStage,
	}

	pipeline, err := h.pipelineUseCase.RemoveStage(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// ReorderPipelineStages handles PUT /pipelines/{pipelineID}/stages/reorder
func (h *Handler) ReorderPipelineStages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	var req ReorderStagesRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.respondError(w, ErrBadRequest(err.Error()))
		return
	}

	stageIDs := make([]uuid.UUID, 0, len(req.StageIDs))
	for i, idStr := range req.StageIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.respondError(w, ErrValidation("stage_ids["+strconv.Itoa(i)+"]", "invalid UUID format"))
			return
		}
		stageIDs = append(stageIDs, id)
	}

	cmd := command.ReorderPipelineStagesCommand{
		TenantID:   tenantID,
		PipelineID: pipelineID,
		StageIDs:   stageIDs,
	}

	pipeline, err := h.pipelineUseCase.ReorderStages(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// ActivatePipelineStage handles POST /pipelines/{pipelineID}/stages/{stageID}/activate
func (h *Handler) ActivatePipelineStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	stageID, err := uuid.Parse(chi.URLParam(r, "stageID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid stage ID"))
		return
	}

	cmd := command.ActivatePipelineStageCommand{
		TenantID:   tenantID,
		PipelineID: pipelineID,
		StageID:    stageID,
	}

	pipeline, err := h.pipelineUseCase.ActivateStage(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// DeactivatePipelineStage handles POST /pipelines/{pipelineID}/stages/{stageID}/deactivate
func (h *Handler) DeactivatePipelineStage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	stageID, err := uuid.Parse(chi.URLParam(r, "stageID"))
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid stage ID"))
		return
	}

	cmd := command.DeactivatePipelineStageCommand{
		TenantID:   tenantID,
		PipelineID: pipelineID,
		StageID:    stageID,
	}

	pipeline, err := h.pipelineUseCase.DeactivateStage(ctx, cmd)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// ============================================================================
// Statistics and Reports
// ============================================================================

// GetPipelineStatistics handles GET /pipelines/{pipelineID}/statistics
func (h *Handler) GetPipelineStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	// Parse optional date range
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var start, end *time.Time
	if startDate != "" {
		t, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			h.respondError(w, ErrValidation("start_date", "invalid date format, use YYYY-MM-DD"))
			return
		}
		start = &t
	}
	if endDate != "" {
		t, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			h.respondError(w, ErrValidation("end_date", "invalid date format, use YYYY-MM-DD"))
			return
		}
		end = &t
	}

	qry := query.GetPipelineStatisticsQuery{
		TenantID:   tenantID,
		PipelineID: pipelineID,
		StartDate:  start,
		EndDate:    end,
	}

	stats, err := h.pipelineUseCase.GetStatistics(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	// Build response
	resp := PipelineStatisticsResponse{
		PipelineID:         stats.PipelineID.String(),
		PipelineName:       stats.PipelineName,
		TotalOpportunities: stats.TotalOpportunities,
		TotalValue:         stats.TotalValue.String(),
		WonOpportunities:   stats.WonOpportunities,
		LostOpportunities:  stats.LostOpportunities,
		WonValue:           stats.WonValue.String(),
		WinRate:            stats.WinRate.String(),
		AverageDealSize:    stats.AverageDealSize.String(),
		AverageDaysToClose: stats.AverageDaysToClose,
		ConversionRate:     stats.ConversionRate.String(),
	}

	resp.StageStatistics = make([]StageStatisticsItem, len(stats.StageStatistics))
	for i, stageStat := range stats.StageStatistics {
		resp.StageStatistics[i] = StageStatisticsItem{
			StageID:            stageStat.StageID.String(),
			StageName:          stageStat.StageName,
			OpportunityCount:   stageStat.OpportunityCount,
			TotalValue:         stageStat.TotalValue.String(),
			AverageValue:       stageStat.AverageValue.String(),
			AverageDaysInStage: stageStat.AverageDaysInStage,
			ConversionRate:     stageStat.ConversionRate.String(),
		}
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// GetPipelineVelocity handles GET /pipelines/{pipelineID}/velocity
func (h *Handler) GetPipelineVelocity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	// Parse period (days)
	period := h.getQueryInt(r, "period", 30)

	qry := query.GetPipelineVelocityQuery{
		TenantID:   tenantID,
		PipelineID: pipelineID,
		PeriodDays: period,
	}

	velocity, err := h.pipelineUseCase.GetVelocity(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"pipeline_id":          velocity.PipelineID.String(),
		"period_days":          velocity.PeriodDays,
		"opportunities_entered": velocity.OpportunitiesEntered,
		"opportunities_won":     velocity.OpportunitiesWon,
		"opportunities_lost":    velocity.OpportunitiesLost,
		"total_value_entered":   velocity.TotalValueEntered.String(),
		"total_value_won":       velocity.TotalValueWon.String(),
		"average_days_to_close": velocity.AverageDaysToClose,
		"velocity_per_day":      velocity.VelocityPerDay.String(),
	})
}

// GetStageConversionRates handles GET /pipelines/{pipelineID}/conversion-rates
func (h *Handler) GetStageConversionRates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	qry := query.GetStageConversionRatesQuery{
		TenantID:   tenantID,
		PipelineID: pipelineID,
	}

	rates, err := h.pipelineUseCase.GetStageConversionRates(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	// Build response
	stageRates := make([]map[string]any, len(rates.StageRates))
	for i, rate := range rates.StageRates {
		stageRates[i] = map[string]any{
			"from_stage_id":    rate.FromStageID.String(),
			"from_stage_name":  rate.FromStageName,
			"to_stage_id":      rate.ToStageID.String(),
			"to_stage_name":    rate.ToStageName,
			"conversion_rate":  rate.ConversionRate.String(),
			"average_days":     rate.AverageDays,
			"total_converted":  rate.TotalConverted,
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"pipeline_id":         pipelineID.String(),
		"overall_conversion":  rates.OverallConversion.String(),
		"stage_conversion_rates": stageRates,
	})
}

// GetForecast handles GET /pipelines/{pipelineID}/forecast
func (h *Handler) GetForecast(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineID, err := h.getUUIDParam(r, "pipelineID")
	if err != nil {
		h.respondError(w, ErrBadRequest("invalid pipeline ID"))
		return
	}

	// Parse forecast period
	period := h.getQueryInt(r, "period", 90) // Default 90 days

	qry := query.GetPipelineForecastQuery{
		TenantID:    tenantID,
		PipelineID:  pipelineID,
		ForecastDays: period,
	}

	forecast, err := h.pipelineUseCase.GetForecast(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	// Build stage forecasts
	stageForecast := make([]map[string]any, len(forecast.StageForecast))
	for i, sf := range forecast.StageForecast {
		stageForecast[i] = map[string]any{
			"stage_id":           sf.StageID.String(),
			"stage_name":         sf.StageName,
			"opportunity_count":  sf.OpportunityCount,
			"total_value":        sf.TotalValue.String(),
			"weighted_value":     sf.WeightedValue.String(),
			"expected_close":     sf.ExpectedClose.String(),
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]any{
		"pipeline_id":         forecast.PipelineID.String(),
		"forecast_period_days": forecast.ForecastPeriodDays,
		"total_pipeline_value": forecast.TotalPipelineValue.String(),
		"weighted_pipeline_value": forecast.WeightedPipelineValue.String(),
		"expected_revenue":    forecast.ExpectedRevenue.String(),
		"best_case_revenue":   forecast.BestCaseRevenue.String(),
		"worst_case_revenue":  forecast.WorstCaseRevenue.String(),
		"stage_forecast":      stageForecast,
	})
}

// GetDefaultPipeline handles GET /pipelines/default
func (h *Handler) GetDefaultPipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	qry := query.GetDefaultPipelineQuery{
		TenantID: tenantID,
	}

	pipeline, err := h.pipelineUseCase.GetDefault(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	h.respondJSON(w, http.StatusOK, h.toPipelineResponse(pipeline))
}

// GetPipelinesByType handles GET /pipelines/by-type/{type}
func (h *Handler) GetPipelinesByType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, err := h.getTenantID(ctx)
	if err != nil {
		h.respondError(w, ErrUnauthorized("tenant identification required"))
		return
	}

	pipelineType := chi.URLParam(r, "type")
	opts := h.buildListOptions(r)

	qry := query.GetPipelinesByTypeQuery{
		TenantID:    tenantID,
		Type:        entity.PipelineType(pipelineType),
		ListOptions: opts,
	}

	pipelines, total, err := h.pipelineUseCase.GetByType(ctx, qry)
	if err != nil {
		h.respondError(w, h.mapAppError(err))
		return
	}

	responses := make([]PipelineResponse, len(pipelines))
	for i, pipeline := range pipelines {
		responses[i] = h.toPipelineResponse(pipeline)
	}

	h.respondList(w, responses, total, opts.Page, opts.PageSize)
}

// ============================================================================
// Response Mapping Helpers
// ============================================================================

// toPipelineResponse converts a domain pipeline to an API response
func (h *Handler) toPipelineResponse(pipeline *entity.Pipeline) PipelineResponse {
	resp := PipelineResponse{
		ID:          pipeline.ID.String(),
		TenantID:    pipeline.TenantID.String(),
		Name:        pipeline.Name,
		Description: pipeline.Description,
		Type:        string(pipeline.Type),
		Currency:    pipeline.Currency,
		IsDefault:   pipeline.IsDefault,
		IsActive:    pipeline.IsActive,
		Metadata:    pipeline.Metadata,
		CreatedAt:   pipeline.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   pipeline.UpdatedAt.Format(time.RFC3339),
	}

	// Map stages
	resp.Stages = make([]StageResponse, len(pipeline.Stages))
	for i, stage := range pipeline.Stages {
		resp.Stages[i] = StageResponse{
			ID:             stage.ID.String(),
			PipelineID:     stage.PipelineID.String(),
			Name:           stage.Name,
			Description:    stage.Description,
			Color:          stage.Color,
			Position:       stage.Position,
			Probability:    stage.Probability,
			RottenDays:     stage.RottenDays,
			IsWinStage:     stage.IsWinStage,
			IsLostStage:    stage.IsLostStage,
			IsActive:       stage.IsActive,
			RequiredFields: stage.RequiredFields,
			Metadata:       stage.Metadata,
			CreatedAt:      stage.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      stage.UpdatedAt.Format(time.RFC3339),
		}
	}

	return resp
}

// Helper function to convert decimal to string (for unused import)
var _ = decimal.Zero
