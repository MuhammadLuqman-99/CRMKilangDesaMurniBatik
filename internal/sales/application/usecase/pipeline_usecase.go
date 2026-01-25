package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application"
	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
	"github.com/kilang-desa-murni/crm/internal/sales/application/ports"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Pipeline Use Case Interface
// ============================================================================

// PipelineUseCase defines the interface for pipeline operations.
type PipelineUseCase interface {
	// CRUD operations
	Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreatePipelineRequest) (*dto.PipelineResponse, error)
	GetByID(ctx context.Context, tenantID, pipelineID uuid.UUID) (*dto.PipelineResponse, error)
	Update(ctx context.Context, tenantID, pipelineID, userID uuid.UUID, req *dto.UpdatePipelineRequest) (*dto.PipelineResponse, error)
	Delete(ctx context.Context, tenantID, pipelineID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *dto.PipelineFilterRequest) (*dto.PipelineListResponse, error)

	// Pipeline operations
	GetDefault(ctx context.Context, tenantID uuid.UUID) (*dto.PipelineResponse, error)
	SetDefault(ctx context.Context, tenantID, pipelineID, userID uuid.UUID) (*dto.PipelineResponse, error)
	Activate(ctx context.Context, tenantID, pipelineID, userID uuid.UUID) (*dto.PipelineResponse, error)
	Deactivate(ctx context.Context, tenantID, pipelineID, userID uuid.UUID) (*dto.PipelineResponse, error)
	Clone(ctx context.Context, tenantID, userID uuid.UUID, req *dto.ClonePipelineRequest) (*dto.PipelineResponse, error)

	// Stage operations
	AddStage(ctx context.Context, tenantID, pipelineID, userID uuid.UUID, req *dto.AddStageRequest) (*dto.PipelineResponse, error)
	UpdateStage(ctx context.Context, tenantID, pipelineID, stageID, userID uuid.UUID, req *dto.UpdateStageRequest) (*dto.PipelineResponse, error)
	RemoveStage(ctx context.Context, tenantID, pipelineID, stageID, userID uuid.UUID) (*dto.PipelineResponse, error)
	ReorderStages(ctx context.Context, tenantID, pipelineID, userID uuid.UUID, req *dto.ReorderStagesRequest) (*dto.PipelineResponse, error)

	// Analytics
	GetStatistics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*dto.PipelineStatisticsDTO, error)
	ComparePipelines(ctx context.Context, tenantID uuid.UUID, req *dto.PipelineComparisonRequest) (*dto.PipelineComparisonResponse, error)
	GetForecast(ctx context.Context, tenantID uuid.UUID, req *dto.ForecastRequest) (*dto.ForecastResponse, error)

	// Templates
	GetTemplates(ctx context.Context) ([]*dto.PipelineTemplateDTO, error)
	CreateFromTemplate(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateFromTemplateRequest) (*dto.PipelineResponse, error)
}

// ============================================================================
// Pipeline Use Case Implementation
// ============================================================================

// pipelineUseCase implements PipelineUseCase.
type pipelineUseCase struct {
	pipelineRepo    domain.PipelineRepository
	opportunityRepo domain.OpportunityRepository
	eventPublisher  ports.EventPublisher
	cacheService    ports.CacheService
	idGenerator     ports.IDGenerator
}

// NewPipelineUseCase creates a new pipeline use case.
func NewPipelineUseCase(
	pipelineRepo domain.PipelineRepository,
	opportunityRepo domain.OpportunityRepository,
	eventPublisher ports.EventPublisher,
	cacheService ports.CacheService,
	idGenerator ports.IDGenerator,
) PipelineUseCase {
	return &pipelineUseCase{
		pipelineRepo:    pipelineRepo,
		opportunityRepo: opportunityRepo,
		eventPublisher:  eventPublisher,
		cacheService:    cacheService,
		idGenerator:     idGenerator,
	}
}

// ============================================================================
// CRUD Operations
// ============================================================================

// Create creates a new pipeline.
func (uc *pipelineUseCase) Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreatePipelineRequest) (*dto.PipelineResponse, error) {
	// Determine currency
	currency := "USD"
	if req.DefaultCurrency != "" {
		currency = req.DefaultCurrency
	}

	// Create pipeline using domain factory
	// NewPipeline(tenantID uuid.UUID, name, currency string, createdBy uuid.UUID)
	pipeline, err := domain.NewPipeline(tenantID, req.Name, currency, userID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeValidation, "failed to create pipeline", err)
	}

	// Set optional fields - Description is string, not *string
	if req.Description != nil {
		pipeline.Description = *req.Description
	}
	pipeline.IsDefault = req.IsDefault
	// Note: AllowSkipStages, RequireWonReason, RequireLostReason don't exist in domain

	if req.WinReasons != nil {
		pipeline.WinReasons = req.WinReasons
	}
	if req.LossReasons != nil {
		pipeline.LossReasons = req.LossReasons
	}

	// Add custom fields schema
	if len(req.CustomFieldsSchema) > 0 {
		pipeline.CustomFields = make([]domain.CustomFieldDef, len(req.CustomFieldsSchema))
		for i, schema := range req.CustomFieldsSchema {
			pipeline.CustomFields[i] = uc.mapCustomFieldSchemaDTOToDomain(schema)
		}
	}

	// Add stages if provided (replacing default stages)
	if len(req.Stages) > 0 {
		// Clear default stages and add requested ones
		pipeline.Stages = nil
		for _, stageReq := range req.Stages {
			stageType := domain.StageType(stageReq.Type)
			// Probability is int (not *int) in CreateStageRequest
			probability := stageReq.Probability
			// AddStage(name string, stageType StageType, probability int) (*Stage, error)
			stage, err := pipeline.AddStage(stageReq.Name, stageType, probability)
			if err != nil {
				return nil, application.WrapError(application.ErrCodeValidation, "failed to add stage: "+stageReq.Name, err)
			}
			// Set additional stage properties
			if stageReq.Description != nil {
				stage.Description = *stageReq.Description
			}
			// Color is string (not *string) in CreateStageRequest
			if stageReq.Color != "" {
				stage.Color = stageReq.Color
			}
			if stageReq.RottenDays != nil {
				stage.RottenDays = *stageReq.RottenDays
			}
		}
	}
	// Default stages are already added by NewPipeline

	// Validate pipeline has required stage types
	if err := uc.validatePipelineStages(pipeline); err != nil {
		return nil, err
	}

	// If this is the default pipeline, unset other defaults
	if req.IsDefault {
		if err := uc.unsetOtherDefaults(ctx, tenantID, pipeline.ID); err != nil {
			return nil, application.WrapError(application.ErrCodeInternal, "failed to update default pipeline", err)
		}
	}

	// Save pipeline
	if err := uc.pipelineRepo.Create(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to save pipeline", err)
	}

	// Publish pipeline created event (Pipeline doesn't have events storage)
	event := domain.NewPipelineCreatedEvent(pipeline)
	uc.publishEvent(ctx, event)

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// GetByID retrieves a pipeline by ID.
func (uc *pipelineUseCase) GetByID(ctx context.Context, tenantID, pipelineID uuid.UUID) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// Update updates a pipeline.
func (uc *pipelineUseCase) Update(ctx context.Context, tenantID, pipelineID, userID uuid.UUID, req *dto.UpdatePipelineRequest) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Check version
	if pipeline.Version != req.Version {
		return nil, application.ErrVersionMismatch(req.Version, pipeline.Version)
	}

	// Update fields using domain method
	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	description := pipeline.Description
	if req.Description != nil {
		description = *req.Description
	}
	pipeline.Update(name, description)

	// Note: AllowSkipStages, RequireWonReason, RequireLostReason don't exist in domain

	if req.WinReasons != nil {
		pipeline.WinReasons = req.WinReasons
	}
	if req.LossReasons != nil {
		pipeline.LossReasons = req.LossReasons
	}
	if req.DefaultCurrency != nil {
		pipeline.Currency = *req.DefaultCurrency
	}

	// Update custom fields schema
	if req.CustomFieldsSchema != nil {
		pipeline.CustomFields = make([]domain.CustomFieldDef, len(req.CustomFieldsSchema))
		for i, schema := range req.CustomFieldsSchema {
			pipeline.CustomFields[i] = uc.mapCustomFieldSchemaDTOToDomain(schema)
		}
	}

	// Handle default flag
	if req.IsDefault != nil && *req.IsDefault && !pipeline.IsDefault {
		if err := uc.unsetOtherDefaults(ctx, tenantID, pipelineID); err != nil {
			return nil, application.WrapError(application.ErrCodeInternal, "failed to update default pipeline", err)
		}
		pipeline.SetAsDefault()
	}

	// Update version
	pipeline.Version++

	// Save changes
	if err := uc.pipelineRepo.Update(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update pipeline", err)
	}

	// Publish update event (Pipeline doesn't have events storage)
	event := domain.NewPipelineUpdatedEvent(pipeline)
	uc.publishEvent(ctx, event)

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// Delete deletes a pipeline.
func (uc *pipelineUseCase) Delete(ctx context.Context, tenantID, pipelineID, userID uuid.UUID) error {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return application.ErrPipelineNotFound(pipelineID)
	}

	// Check if pipeline is default
	if pipeline.IsDefault {
		return application.ErrPipelineDefaultRequired()
	}

	// Check if pipeline has opportunities
	opportunities, _, err := uc.opportunityRepo.GetByPipeline(ctx, tenantID, pipelineID, domain.ListOptions{PageSize: 1})
	if err == nil && len(opportunities) > 0 {
		return application.ErrPipelineHasOpportunities(pipelineID)
	}

	// Delete pipeline
	if err := uc.pipelineRepo.Delete(ctx, tenantID, pipelineID); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to delete pipeline", err)
	}

	// Note: PipelineDeletedEvent doesn't exist in domain, skip publishing

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return nil
}

// List lists pipelines with filtering.
func (uc *pipelineUseCase) List(ctx context.Context, tenantID uuid.UUID, filter *dto.PipelineFilterRequest) (*dto.PipelineListResponse, error) {
	// Set pagination defaults
	opts := domain.DefaultListOptions()
	if filter.Page > 0 {
		opts.Page = filter.Page
	}
	if filter.PageSize > 0 {
		opts.PageSize = filter.PageSize
	}
	if filter.SortBy != "" {
		opts.SortBy = filter.SortBy
	}
	if filter.SortOrder != "" {
		opts.SortOrder = filter.SortOrder
	}

	// Get pipelines
	pipelines, total, err := uc.pipelineRepo.List(ctx, tenantID, opts)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to list pipelines", err)
	}

	// Map to response
	pipelineResponses := make([]*dto.PipelineBriefResponse, len(pipelines))
	for i, pipeline := range pipelines {
		var opportunityCount int64
		if filter.IncludeOpportunityCounts {
			_, count, _ := uc.opportunityRepo.GetByPipeline(ctx, tenantID, pipeline.ID, domain.ListOptions{PageSize: 1})
			opportunityCount = count
		}
		pipelineResponses[i] = &dto.PipelineBriefResponse{
			ID:               pipeline.ID.String(),
			Name:             pipeline.Name,
			IsActive:         pipeline.IsActive,
			IsDefault:        pipeline.IsDefault,
			StageCount:       len(pipeline.Stages),
			OpportunityCount: opportunityCount,
			CreatedAt:        pipeline.CreatedAt,
		}
	}

	return &dto.PipelineListResponse{
		Pipelines:  pipelineResponses,
		Pagination: dto.NewPaginationResponse(opts.Page, opts.PageSize, total),
	}, nil
}

// ============================================================================
// Pipeline Operations
// ============================================================================

// GetDefault retrieves the default pipeline.
func (uc *pipelineUseCase) GetDefault(ctx context.Context, tenantID uuid.UUID) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetDefaultPipeline(ctx, tenantID)
	if err != nil {
		return nil, application.NewAppError(application.ErrCodePipelineNotFound, "no default pipeline found")
	}

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// SetDefault sets a pipeline as the default.
func (uc *pipelineUseCase) SetDefault(ctx context.Context, tenantID, pipelineID, userID uuid.UUID) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	if !pipeline.IsActive {
		return nil, application.ErrPipelineInactive(pipelineID)
	}

	// Unset other defaults
	if err := uc.unsetOtherDefaults(ctx, tenantID, pipelineID); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update default pipeline", err)
	}

	// Set as default using domain method
	pipeline.SetAsDefault()
	pipeline.Version++

	// Save changes
	if err := uc.pipelineRepo.Update(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update pipeline", err)
	}

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// Activate activates a pipeline.
func (uc *pipelineUseCase) Activate(ctx context.Context, tenantID, pipelineID, userID uuid.UUID) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Activate - domain method doesn't take arguments or return error
	pipeline.Activate()
	pipeline.Version++

	// Save changes
	if err := uc.pipelineRepo.Update(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update pipeline", err)
	}

	// Note: Pipeline doesn't have events storage

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// Deactivate deactivates a pipeline.
func (uc *pipelineUseCase) Deactivate(ctx context.Context, tenantID, pipelineID, userID uuid.UUID) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Check if pipeline is default
	if pipeline.IsDefault {
		return nil, application.NewAppError(application.ErrCodeConflict, "cannot deactivate default pipeline")
	}

	// Check if pipeline has open opportunities
	openOpps, count, _ := uc.opportunityRepo.GetOpenOpportunities(ctx, tenantID, domain.ListOptions{PageSize: 1})
	for _, opp := range openOpps {
		if opp.PipelineID == pipelineID {
			return nil, application.NewAppError(application.ErrCodeConflict, "pipeline has open opportunities")
		}
	}
	_ = count

	// Deactivate - domain method takes no args but returns error
	if err := pipeline.Deactivate(); err != nil {
		return nil, application.WrapError(application.ErrCodeConflict, err.Error(), err)
	}
	pipeline.Version++

	// Save changes
	if err := uc.pipelineRepo.Update(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update pipeline", err)
	}

	// Note: Pipeline doesn't have events storage

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// Clone clones a pipeline.
func (uc *pipelineUseCase) Clone(ctx context.Context, tenantID, userID uuid.UUID, req *dto.ClonePipelineRequest) (*dto.PipelineResponse, error) {
	// Parse source pipeline ID
	sourcePipelineID, err := uuid.Parse(req.SourcePipelineID)
	if err != nil {
		return nil, application.ErrValidation("invalid source_pipeline_id format")
	}

	// Get source pipeline
	source, err := uc.pipelineRepo.GetByID(ctx, tenantID, sourcePipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(sourcePipelineID)
	}

	// Clone pipeline manually (Pipeline doesn't have Clone method)
	now := time.Now().UTC()
	cloneID := uuid.New()

	// Clone stages
	clonedStages := make([]*domain.Stage, len(source.Stages))
	for i, s := range source.Stages {
		clonedStages[i] = &domain.Stage{
			ID:          uuid.New(),
			PipelineID:  cloneID,
			Name:        s.Name,
			Description: s.Description,
			Type:        s.Type,
			Order:       s.Order,
			Probability: s.Probability,
			Color:       s.Color,
			IsActive:    s.IsActive,
			RottenDays:  s.RottenDays,
			AutoActions: s.AutoActions,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}

	clone := &domain.Pipeline{
		ID:           cloneID,
		TenantID:     tenantID,
		Name:         req.Name,
		Description:  source.Description,
		IsDefault:    false,
		IsActive:     true,
		Currency:     source.Currency,
		Stages:       clonedStages,
		WinReasons:   source.WinReasons,
		LossReasons:  source.LossReasons,
		CreatedBy:    userID,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}

	// Set description if provided
	if req.Description != nil {
		clone.Description = *req.Description
	}

	// Include or exclude custom fields
	if req.IncludeCustomFields {
		clone.CustomFields = source.CustomFields
	}

	// Save clone
	if err := uc.pipelineRepo.Create(ctx, clone); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to save cloned pipeline", err)
	}

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, clone), nil
}

// ============================================================================
// Stage Operations
// ============================================================================

// AddStage adds a stage to a pipeline.
func (uc *pipelineUseCase) AddStage(ctx context.Context, tenantID, pipelineID, userID uuid.UUID, req *dto.AddStageRequest) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Use domain's AddStage method
	// AddStage(name string, stageType StageType, probability int) (*Stage, error)
	stageType := domain.StageType(req.Type)
	stage, err := pipeline.AddStage(req.Name, stageType, req.Probability)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeValidation, err.Error(), err)
	}

	// Set additional stage properties
	if req.Description != nil {
		stage.Description = *req.Description
	}
	stage.Color = req.Color
	if req.RottenDays != nil {
		stage.RottenDays = *req.RottenDays
	}
	// Note: RequiredFields doesn't exist on Stage

	// Update metadata
	pipeline.UpdatedAt = time.Now().UTC()
	pipeline.Version++

	// Save changes
	if err := uc.pipelineRepo.Update(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update pipeline", err)
	}

	// Note: Pipeline doesn't have events storage

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// UpdateStage updates a stage in a pipeline.
func (uc *pipelineUseCase) UpdateStage(ctx context.Context, tenantID, pipelineID, stageID, userID uuid.UUID, req *dto.UpdateStageRequest) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Find stage
	var stage *domain.Stage
	for _, s := range pipeline.Stages {
		if s.ID == stageID {
			stage = s
			break
		}
	}
	if stage == nil {
		return nil, application.ErrPipelineStageNotFound(pipelineID, stageID)
	}

	// Update fields
	if req.Name != nil {
		stage.Name = *req.Name
	}
	if req.Description != nil {
		stage.Description = *req.Description
	}
	if req.Probability != nil {
		stage.Probability = *req.Probability
	}
	if req.Color != nil {
		stage.Color = *req.Color
	}
	if req.RottenDays != nil {
		stage.RottenDays = *req.RottenDays
	}
	// Note: RequiredFields doesn't exist on Stage

	// Map auto actions (AutoAction has Type string, Config map, DelayHours int)
	if req.AutoActions != nil {
		stage.AutoActions = make([]domain.AutoAction, len(req.AutoActions))
		for i, action := range req.AutoActions {
			autoAction := domain.AutoAction{
				Type:   action.Type,
				Config: action.Config,
			}
			if action.DelayHours != nil {
				autoAction.DelayHours = *action.DelayHours
			}
			stage.AutoActions[i] = autoAction
		}
	}
	if req.IsActive != nil {
		stage.IsActive = *req.IsActive
	}

	stage.UpdatedAt = time.Now().UTC()

	// Update metadata
	pipeline.UpdatedAt = time.Now().UTC()
	pipeline.Version++

	// Save changes
	if err := uc.pipelineRepo.Update(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update pipeline", err)
	}

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// RemoveStage removes a stage from a pipeline.
func (uc *pipelineUseCase) RemoveStage(ctx context.Context, tenantID, pipelineID, stageID, userID uuid.UUID) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Check if stage has opportunities
	opportunities, _, _ := uc.opportunityRepo.GetByStage(ctx, tenantID, pipelineID, stageID, domain.ListOptions{PageSize: 1})
	if len(opportunities) > 0 {
		return nil, application.ErrPipelineStageHasOpportunities(stageID)
	}

	// Remove stage
	if err := pipeline.RemoveStage(stageID); err != nil {
		return nil, application.WrapError(application.ErrCodeValidation, err.Error(), err)
	}

	// Validate remaining stages
	if err := uc.validatePipelineStages(pipeline); err != nil {
		return nil, err
	}

	// Update metadata
	pipeline.UpdatedAt = time.Now().UTC()
	pipeline.Version++

	// Save changes
	if err := uc.pipelineRepo.Update(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update pipeline", err)
	}

	// Note: Pipeline doesn't have events storage

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// ReorderStages reorders stages in a pipeline.
func (uc *pipelineUseCase) ReorderStages(ctx context.Context, tenantID, pipelineID, userID uuid.UUID, req *dto.ReorderStagesRequest) (*dto.PipelineResponse, error) {
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Build stage ID list in order (domain expects []uuid.UUID in desired order)
	// Sort by order from request
	type stageOrder struct {
		id    uuid.UUID
		order int
	}
	orders := make([]stageOrder, 0, len(req.StageOrders))
	for _, order := range req.StageOrders {
		stageID, _ := uuid.Parse(order.StageID)
		orders = append(orders, stageOrder{id: stageID, order: order.Order})
	}
	// Sort by order
	for i := 0; i < len(orders)-1; i++ {
		for j := i + 1; j < len(orders); j++ {
			if orders[j].order < orders[i].order {
				orders[i], orders[j] = orders[j], orders[i]
			}
		}
	}
	// Build stage ID list
	stageIDs := make([]uuid.UUID, len(orders))
	for i, o := range orders {
		stageIDs[i] = o.id
	}

	// Reorder stages - domain expects []uuid.UUID
	if err := pipeline.ReorderStages(stageIDs); err != nil {
		return nil, application.WrapError(application.ErrCodeValidation, err.Error(), err)
	}

	// Update metadata
	pipeline.UpdatedAt = time.Now().UTC()
	pipeline.Version++

	// Save changes
	if err := uc.pipelineRepo.Update(ctx, pipeline); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update pipeline", err)
	}

	// Invalidate cache
	uc.invalidatePipelineCache(ctx, tenantID)

	return uc.mapPipelineToResponse(ctx, pipeline), nil
}

// ============================================================================
// Analytics
// ============================================================================

// GetStatistics retrieves pipeline statistics.
func (uc *pipelineUseCase) GetStatistics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*dto.PipelineStatisticsDTO, error) {
	// Get pipeline statistics
	stats, err := uc.pipelineRepo.GetPipelineStatistics(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to get pipeline statistics", err)
	}

	// Map stage distribution to string keys
	stageDistribution := make(map[string]int64)
	for stageID, count := range stats.StageDistribution {
		stageDistribution[stageID.String()] = count
	}

	// Map conversion rates to string keys
	conversionRates := make(map[string]float64)
	for stageID, rate := range stats.ConversionRates {
		conversionRates[stageID.String()] = rate
	}

	return &dto.PipelineStatisticsDTO{
		TotalOpportunities:  stats.TotalOpportunities,
		OpenOpportunities:   stats.OpenOpportunities,
		WonOpportunities:    stats.WonOpportunities,
		LostOpportunities:   stats.LostOpportunities,
		TotalValue: dto.MoneyDTO{
			Amount:   stats.TotalValue.Amount,
			Currency: stats.TotalValue.Currency,
		},
		WeightedValue: dto.MoneyDTO{
			Amount:   stats.WeightedValue.Amount,
			Currency: stats.WeightedValue.Currency,
		},
		WinRate:           stats.WinRate,
		AverageSalesCycle: stats.AverageSalesCycle,
		StageDistribution: stageDistribution,
		ConversionRates:   conversionRates,
	}, nil
}

// ComparePipelines compares multiple pipelines.
func (uc *pipelineUseCase) ComparePipelines(ctx context.Context, tenantID uuid.UUID, req *dto.PipelineComparisonRequest) (*dto.PipelineComparisonResponse, error) {
	comparisons := make([]dto.PipelineComparisonItemDTO, len(req.PipelineIDs))

	for i, pipelineIDStr := range req.PipelineIDs {
		pipelineID, _ := uuid.Parse(pipelineIDStr)

		// Get pipeline
		pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
		if err != nil {
			continue
		}

		// Get statistics
		stats, err := uc.pipelineRepo.GetPipelineStatistics(ctx, tenantID, pipelineID)
		if err != nil {
			continue
		}

		comparisons[i] = dto.PipelineComparisonItemDTO{
			PipelineID:         pipelineID.String(),
			PipelineName:       pipeline.Name,
			TotalOpportunities: stats.TotalOpportunities,
			WonOpportunities:   stats.WonOpportunities,
			LostOpportunities:  stats.LostOpportunities,
			WinRate:            stats.WinRate,
			TotalValue: dto.MoneyDTO{
				Amount:   stats.TotalValue.Amount,
				Currency: stats.TotalValue.Currency,
			},
			WeightedValue: dto.MoneyDTO{
				Amount:   stats.WeightedValue.Amount,
				Currency: stats.WeightedValue.Currency,
			},
			AverageSalesCycle: stats.AverageSalesCycle,
		}
	}

	resp := &dto.PipelineComparisonResponse{
		Pipelines: comparisons,
	}

	// Set period if dates provided
	if req.StartDate != nil && req.EndDate != nil {
		startDate, _ := time.Parse("2006-01-02", *req.StartDate)
		endDate, _ := time.Parse("2006-01-02", *req.EndDate)
		resp.Period = &dto.DateRangeDTO{
			StartDate: startDate,
			EndDate:   endDate,
		}
	}

	return resp, nil
}

// GetForecast retrieves pipeline forecast.
func (uc *pipelineUseCase) GetForecast(ctx context.Context, tenantID uuid.UUID, req *dto.ForecastRequest) (*dto.ForecastResponse, error) {
	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, application.ErrValidation("invalid start_date format")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, application.ErrValidation("invalid end_date format")
	}

	// Build filter
	filter := domain.OpportunityFilter{
		Statuses:               []domain.OpportunityStatus{domain.OpportunityStatusOpen},
		ExpectedCloseDateAfter: &startDate,
		ExpectedCloseDateBefore: &endDate,
		Currency:               &req.Currency,
	}

	if req.PipelineID != nil {
		pipelineID, _ := uuid.Parse(*req.PipelineID)
		filter.PipelineIDs = []uuid.UUID{pipelineID}
	}

	// Get opportunities
	opportunities, _, err := uc.opportunityRepo.List(ctx, tenantID, filter, domain.ListOptions{PageSize: 1000})
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to get opportunities", err)
	}

	// Calculate forecast
	var totalForecast, committed, pipeline, upside int64
	byPeriod := make(map[string]*dto.ForecastPeriodDTO)

	for _, opp := range opportunities {
		amount := opp.WeightedAmount.Amount

		// Categorize by probability
		if opp.Probability >= 75 {
			committed += amount
		} else if opp.Probability >= 25 {
			pipeline += amount
		} else {
			upside += amount
		}
		totalForecast += amount

		// Group by period - ExpectedCloseDate is *time.Time
		var periodKey string
		if opp.ExpectedCloseDate != nil {
			periodKey = uc.getPeriodKey(*opp.ExpectedCloseDate, req.GroupBy)
		} else {
			periodKey = "Unknown"
		}
		if _, ok := byPeriod[periodKey]; !ok {
			byPeriod[periodKey] = &dto.ForecastPeriodDTO{
				Period: periodKey,
				ExpectedRevenue: dto.MoneyDTO{
					Amount:   0,
					Currency: req.Currency,
				},
				WeightedRevenue: dto.MoneyDTO{
					Amount:   0,
					Currency: req.Currency,
				},
				Committed: dto.MoneyDTO{
					Amount:   0,
					Currency: req.Currency,
				},
				Pipeline: dto.MoneyDTO{
					Amount:   0,
					Currency: req.Currency,
				},
				Upside: dto.MoneyDTO{
					Amount:   0,
					Currency: req.Currency,
				},
			}
		}
		byPeriod[periodKey].ExpectedRevenue.Amount += opp.Amount.Amount
		byPeriod[periodKey].WeightedRevenue.Amount += amount
		byPeriod[periodKey].OpportunityCount++

		if opp.Probability >= 75 {
			byPeriod[periodKey].Committed.Amount += amount
		} else if opp.Probability >= 25 {
			byPeriod[periodKey].Pipeline.Amount += amount
		} else {
			byPeriod[periodKey].Upside.Amount += amount
		}
	}

	// Convert map to slice
	periods := make([]dto.ForecastPeriodDTO, 0, len(byPeriod))
	for _, period := range byPeriod {
		periods = append(periods, *period)
	}

	// Calculate best/worst case
	bestCase := committed + pipeline + upside
	worstCase := committed

	return &dto.ForecastResponse{
		Period: dto.DateRangeDTO{
			StartDate: startDate,
			EndDate:   endDate,
		},
		Currency: req.Currency,
		TotalForecast: dto.MoneyDTO{
			Amount:   totalForecast,
			Currency: req.Currency,
		},
		BestCase: dto.MoneyDTO{
			Amount:   bestCase,
			Currency: req.Currency,
		},
		WorstCase: dto.MoneyDTO{
			Amount:   worstCase,
			Currency: req.Currency,
		},
		Committed: dto.MoneyDTO{
			Amount:   committed,
			Currency: req.Currency,
		},
		Pipeline: dto.MoneyDTO{
			Amount:   pipeline,
			Currency: req.Currency,
		},
		Upside: dto.MoneyDTO{
			Amount:   upside,
			Currency: req.Currency,
		},
		ByPeriod: periods,
	}, nil
}

// ============================================================================
// Templates
// ============================================================================

// GetTemplates retrieves available pipeline templates.
func (uc *pipelineUseCase) GetTemplates(ctx context.Context) ([]*dto.PipelineTemplateDTO, error) {
	// Return predefined templates
	templates := []*dto.PipelineTemplateDTO{
		{
			ID:          "b2b-sales",
			Name:        "B2B Sales Pipeline",
			Description: "Standard B2B sales pipeline with qualification, proposal, and negotiation stages",
			Industry:    "General",
			Stages: []*dto.StageTemplateDTO{
				{Name: "Qualification", Type: "qualifying", Probability: 10, Color: "#6366F1"},
				{Name: "Needs Analysis", Type: "open", Probability: 20, Color: "#8B5CF6"},
				{Name: "Proposal", Type: "open", Probability: 40, Color: "#A855F7"},
				{Name: "Negotiation", Type: "negotiating", Probability: 60, Color: "#D946EF"},
				{Name: "Closed Won", Type: "won", Probability: 100, Color: "#22C55E"},
				{Name: "Closed Lost", Type: "lost", Probability: 0, Color: "#EF4444"},
			},
			WinReasons:  []string{"Best fit", "Price", "Relationship", "Product features", "Service quality"},
			LossReasons: []string{"Price", "Lost to competitor", "No decision", "Budget constraints", "Bad timing"},
		},
		{
			ID:          "saas-sales",
			Name:        "SaaS Sales Pipeline",
			Description: "Pipeline optimized for SaaS product sales with demo and trial stages",
			Industry:    "Technology",
			Stages: []*dto.StageTemplateDTO{
				{Name: "Lead", Type: "qualifying", Probability: 5, Color: "#6366F1"},
				{Name: "Demo Scheduled", Type: "open", Probability: 20, Color: "#8B5CF6"},
				{Name: "Demo Completed", Type: "open", Probability: 35, Color: "#A855F7"},
				{Name: "Trial", Type: "open", Probability: 50, Color: "#D946EF"},
				{Name: "Contract", Type: "negotiating", Probability: 75, Color: "#EC4899"},
				{Name: "Closed Won", Type: "won", Probability: 100, Color: "#22C55E"},
				{Name: "Closed Lost", Type: "lost", Probability: 0, Color: "#EF4444"},
			},
			WinReasons:  []string{"Product fit", "Pricing", "Integration capabilities", "Support quality", "Time to value"},
			LossReasons: []string{"Price too high", "Missing features", "Lost to competitor", "No budget", "Went with existing solution"},
		},
		{
			ID:          "real-estate",
			Name:        "Real Estate Pipeline",
			Description: "Pipeline for real estate sales with showing and offer stages",
			Industry:    "Real Estate",
			Stages: []*dto.StageTemplateDTO{
				{Name: "New Inquiry", Type: "qualifying", Probability: 5, Color: "#6366F1"},
				{Name: "Showing Scheduled", Type: "open", Probability: 15, Color: "#8B5CF6"},
				{Name: "Showing Done", Type: "open", Probability: 30, Color: "#A855F7"},
				{Name: "Offer Made", Type: "negotiating", Probability: 50, Color: "#D946EF"},
				{Name: "Offer Accepted", Type: "negotiating", Probability: 75, Color: "#EC4899"},
				{Name: "Under Contract", Type: "negotiating", Probability: 90, Color: "#F97316"},
				{Name: "Closed", Type: "won", Probability: 100, Color: "#22C55E"},
				{Name: "Lost", Type: "lost", Probability: 0, Color: "#EF4444"},
			},
			WinReasons:  []string{"Location", "Price", "Property condition", "Timing", "Market conditions"},
			LossReasons: []string{"Price too high", "Property issues", "Financing fell through", "Changed mind", "Went with another property"},
		},
	}

	return templates, nil
}

// CreateFromTemplate creates a pipeline from a template.
func (uc *pipelineUseCase) CreateFromTemplate(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateFromTemplateRequest) (*dto.PipelineResponse, error) {
	// Get templates
	templates, _ := uc.GetTemplates(ctx)

	// Find template
	var template *dto.PipelineTemplateDTO
	for _, t := range templates {
		if t.ID == req.TemplateID {
			template = t
			break
		}
	}
	if template == nil {
		return nil, application.NewAppError(application.ErrCodeNotFound, "template not found")
	}

	// Create pipeline from template
	createReq := &dto.CreatePipelineRequest{
		Name:        req.Name,
		Description: req.Description,
		IsDefault:   req.IsDefault,
		Stages:      make([]*dto.CreateStageRequest, len(template.Stages)),
		WinReasons:  template.WinReasons,
		LossReasons: template.LossReasons,
	}

	for i, stage := range template.Stages {
		createReq.Stages[i] = &dto.CreateStageRequest{
			Name:        stage.Name,
			Type:        stage.Type,
			Order:       i + 1,
			Probability: stage.Probability,
			Color:       stage.Color,
		}
	}

	return uc.Create(ctx, tenantID, userID, createReq)
}

// ============================================================================
// Helper Methods
// ============================================================================

func (uc *pipelineUseCase) mapPipelineToResponse(ctx context.Context, pipeline *domain.Pipeline) *dto.PipelineResponse {
	resp := &dto.PipelineResponse{
		ID:              pipeline.ID.String(),
		TenantID:        pipeline.TenantID.String(),
		Name:            pipeline.Name,
		IsActive:        pipeline.IsActive,
		IsDefault:       pipeline.IsDefault,
		StageCount:      len(pipeline.Stages),
		WinReasons:      pipeline.WinReasons,
		LossReasons:     pipeline.LossReasons,
		DefaultCurrency: pipeline.Currency, // Domain uses Currency, not DefaultCurrency
		CreatedAt:       pipeline.CreatedAt,
		UpdatedAt:       pipeline.UpdatedAt,
		CreatedBy:       pipeline.CreatedBy.String(),
		Version:         pipeline.Version,
	}

	// Description is string in domain, *string in DTO
	if pipeline.Description != "" {
		resp.Description = &pipeline.Description
	}

	// Note: AllowSkipStages, RequireWonReason, RequireLostReason, UpdatedBy don't exist in domain

	// Map stages
	resp.Stages = make([]*dto.StageResponse, len(pipeline.Stages))
	for i, stage := range pipeline.Stages {
		stageResp := &dto.StageResponse{
			ID:          stage.ID.String(),
			PipelineID:  pipeline.ID.String(),
			Name:        stage.Name,
			Type:        string(stage.Type),
			Order:       stage.Order,
			Probability: stage.Probability,
			Color:       stage.Color,
			IsActive:    stage.IsActive,
			CreatedAt:   stage.CreatedAt,
			UpdatedAt:   stage.UpdatedAt,
		}

		// Description is string in domain, *string in DTO
		if stage.Description != "" {
			stageResp.Description = &stage.Description
		}

		// RottenDays is int in domain, *int in DTO
		if stage.RottenDays > 0 {
			stageResp.RottenDays = &stage.RottenDays
		}

		// Note: RequiredFields doesn't exist on Stage

		// Map auto actions (AutoAction has Type string, Config map, DelayHours int)
		if len(stage.AutoActions) > 0 {
			stageResp.AutoActions = make([]*dto.StageAutoActionDTO, len(stage.AutoActions))
			for j, action := range stage.AutoActions {
				delayHours := action.DelayHours
				stageResp.AutoActions[j] = &dto.StageAutoActionDTO{
					Type:       action.Type,
					Config:     action.Config,
					DelayHours: &delayHours,
				}
			}
		}

		resp.Stages[i] = stageResp
	}

	// Map custom fields schema
	if len(pipeline.CustomFields) > 0 {
		resp.CustomFieldsSchema = make([]*dto.CustomFieldSchemaDTO, len(pipeline.CustomFields))
		for i, schema := range pipeline.CustomFields {
			resp.CustomFieldsSchema[i] = uc.mapCustomFieldSchemaDomainToDTO(schema)
		}
	}

	return resp
}

func (uc *pipelineUseCase) createStageFromRequest(req *dto.CreateStageRequest) *domain.Stage {
	now := time.Now().UTC()
	stage := &domain.Stage{
		ID:          uc.idGenerator.GenerateID(),
		Name:        req.Name,
		Type:        domain.StageType(req.Type),
		Order:       req.Order,
		Probability: req.Probability,
		Color:       req.Color,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Description is *string in DTO, string in domain
	if req.Description != nil {
		stage.Description = *req.Description
	}

	// RottenDays is *int in DTO, int in domain
	if req.RottenDays != nil {
		stage.RottenDays = *req.RottenDays
	}

	// Note: RequiredFields doesn't exist on Stage

	// Map auto actions (AutoAction has Type string, Config map, DelayHours int)
	if len(req.AutoActions) > 0 {
		stage.AutoActions = make([]domain.AutoAction, len(req.AutoActions))
		for i, action := range req.AutoActions {
			autoAction := domain.AutoAction{
				Type:   action.Type,
				Config: action.Config,
			}
			if action.DelayHours != nil {
				autoAction.DelayHours = *action.DelayHours
			}
			stage.AutoActions[i] = autoAction
		}
	}

	return stage
}

func (uc *pipelineUseCase) addDefaultStages(pipeline *domain.Pipeline) {
	defaultStages := []*domain.Stage{
		{
			ID:          uc.idGenerator.GenerateID(),
			Name:        "Qualification",
			Type:        domain.StageTypeQualifying,
			Order:       1,
			Probability: 10,
			Color:       "#6366F1",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uc.idGenerator.GenerateID(),
			Name:        "Proposal",
			Type:        domain.StageTypeOpen,
			Order:       2,
			Probability: 30,
			Color:       "#8B5CF6",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uc.idGenerator.GenerateID(),
			Name:        "Negotiation",
			Type:        domain.StageTypeNegotiating,
			Order:       3,
			Probability: 60,
			Color:       "#D946EF",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uc.idGenerator.GenerateID(),
			Name:        "Closed Won",
			Type:        domain.StageTypeWon,
			Order:       4,
			Probability: 100,
			Color:       "#22C55E",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uc.idGenerator.GenerateID(),
			Name:        "Closed Lost",
			Type:        domain.StageTypeLost,
			Order:       5,
			Probability: 0,
			Color:       "#EF4444",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, stage := range defaultStages {
		stage.PipelineID = pipeline.ID
		pipeline.Stages = append(pipeline.Stages, stage)
	}
}

func (uc *pipelineUseCase) validatePipelineStages(pipeline *domain.Pipeline) error {
	// Check minimum stages
	if len(pipeline.Stages) < 3 {
		return application.ErrPipelineMinStagesRequired(3)
	}

	// Check for won stage
	hasWonStage := false
	hasLostStage := false
	for _, stage := range pipeline.Stages {
		if stage.Type == domain.StageTypeWon && stage.IsActive {
			hasWonStage = true
		}
		if stage.Type == domain.StageTypeLost && stage.IsActive {
			hasLostStage = true
		}
	}

	if !hasWonStage {
		return application.ErrPipelineWonStageRequired()
	}
	if !hasLostStage {
		return application.ErrPipelineLostStageRequired()
	}

	return nil
}

func (uc *pipelineUseCase) unsetOtherDefaults(ctx context.Context, tenantID, exceptPipelineID uuid.UUID) error {
	pipelines, _, err := uc.pipelineRepo.List(ctx, tenantID, domain.ListOptions{PageSize: 100})
	if err != nil {
		return err
	}

	for _, p := range pipelines {
		if p.ID != exceptPipelineID && p.IsDefault {
			p.IsDefault = false
			if err := uc.pipelineRepo.Update(ctx, p); err != nil {
				return err
			}
		}
	}

	return nil
}

func (uc *pipelineUseCase) mapCustomFieldSchemaDTOToDomain(dtoField *dto.CustomFieldSchemaDTO) domain.CustomFieldDef {
	field := domain.CustomFieldDef{
		Name:     dtoField.Name,
		Label:    dtoField.Label,
		Type:     dtoField.Type,
		Required: dtoField.Required,
	}
	if dtoField.DefaultValue != nil {
		if defaultStr, ok := dtoField.DefaultValue.(string); ok {
			field.Default = defaultStr
		}
	}

	// Map options - convert from structured DTO to simple strings
	if len(dtoField.Options) > 0 {
		field.Options = make([]string, len(dtoField.Options))
		for i, opt := range dtoField.Options {
			field.Options[i] = opt.Value
		}
	}

	return field
}

func (uc *pipelineUseCase) mapCustomFieldSchemaDomainToDTO(field domain.CustomFieldDef) *dto.CustomFieldSchemaDTO {
	result := &dto.CustomFieldSchemaDTO{
		Name:         field.Name,
		Label:        field.Label,
		Type:         field.Type,
		Required:     field.Required,
		DefaultValue: field.Default,
		IsActive:     true, // Domain doesn't have IsActive, default to true
	}

	// Map options - convert from simple strings to structured DTO
	if len(field.Options) > 0 {
		result.Options = make([]dto.CustomFieldOptionDTO, len(field.Options))
		for i, opt := range field.Options {
			result.Options[i] = dto.CustomFieldOptionDTO{
				Value: opt,
				Label: opt, // Use value as label since domain only stores value
			}
		}
	}

	return result
}

func (uc *pipelineUseCase) getPeriodKey(date time.Time, groupBy string) string {
	switch groupBy {
	case "day":
		return date.Format("2006-01-02")
	case "week":
		year, week := date.ISOWeek()
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, (week-1)*7).Format("2006-01-02")
	case "quarter":
		quarter := (date.Month()-1)/3 + 1
		return date.Format("2006") + "-Q" + string('0'+rune(quarter))
	default:
		return date.Format("2006-01")
	}
}

func (uc *pipelineUseCase) publishEvent(ctx context.Context, event domain.DomainEvent) error {
	if uc.eventPublisher == nil {
		return nil
	}

	return uc.eventPublisher.Publish(ctx, ports.Event{
		ID:            event.EventID().String(),
		Type:          event.EventType(),
		AggregateID:   event.AggregateID().String(),
		AggregateType: event.AggregateType(),
		TenantID:      event.TenantID().String(),
		OccurredAt:    event.OccurredAt(),
		Version:       event.Version(),
	})
}

func (uc *pipelineUseCase) invalidatePipelineCache(ctx context.Context, tenantID uuid.UUID) {
	if uc.cacheService == nil {
		return
	}

	pattern := "pipeline:" + tenantID.String() + ":*"
	uc.cacheService.DeletePattern(ctx, pattern)
}
