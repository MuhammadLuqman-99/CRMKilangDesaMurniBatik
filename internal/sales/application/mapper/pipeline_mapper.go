// Package mapper provides mapping functions between domain entities and DTOs.
package mapper

import (
	"github.com/google/uuid"

	"github.com/DesmondSanctworker/CRMKilangDesaMurniBatik/internal/sales/application/dto"
	"github.com/DesmondSanctworker/CRMKilangDesaMurniBatik/internal/sales/domain"
)

// ============================================================================
// Pipeline Mappers
// ============================================================================

// PipelineMapper handles mapping between Pipeline entities and DTOs.
type PipelineMapper struct{}

// NewPipelineMapper creates a new PipelineMapper instance.
func NewPipelineMapper() *PipelineMapper {
	return &PipelineMapper{}
}

// ToResponse maps a Pipeline entity to PipelineResponse DTO.
func (m *PipelineMapper) ToResponse(pipeline *domain.Pipeline, includeStats bool) *dto.PipelineResponse {
	if pipeline == nil {
		return nil
	}

	response := &dto.PipelineResponse{
		ID:              pipeline.ID.String(),
		TenantID:        pipeline.TenantID.String(),
		Name:            pipeline.Name,
		IsActive:        pipeline.IsActive,
		IsDefault:       pipeline.IsDefault,
		WinReasons:      pipeline.WinReasons,
		LossReasons:     pipeline.LossReasons,
		DefaultCurrency: pipeline.Currency,
		CreatedAt:       pipeline.CreatedAt,
		UpdatedAt:       pipeline.UpdatedAt,
		CreatedBy:       pipeline.CreatedBy.String(),
		Version:         pipeline.Version,
	}

	// Description
	if pipeline.Description != "" {
		response.Description = dto.StringPtr(pipeline.Description)
	}

	// Stages
	response.Stages = m.mapStages(pipeline.Stages)
	response.StageCount = len(pipeline.GetActiveStages())

	// Custom fields schema
	response.CustomFieldsSchema = m.mapCustomFieldDefs(pipeline.CustomFields)

	// Statistics (optional)
	if includeStats {
		response.Statistics = m.buildStatistics(pipeline)
	}

	return response
}

// ToBriefResponse maps a Pipeline entity to PipelineBriefResponse DTO.
func (m *PipelineMapper) ToBriefResponse(pipeline *domain.Pipeline) *dto.PipelineBriefResponse {
	if pipeline == nil {
		return nil
	}

	return &dto.PipelineBriefResponse{
		ID:               pipeline.ID.String(),
		Name:             pipeline.Name,
		IsActive:         pipeline.IsActive,
		IsDefault:        pipeline.IsDefault,
		StageCount:       len(pipeline.GetActiveStages()),
		OpportunityCount: pipeline.OpportunityCount,
		CreatedAt:        pipeline.CreatedAt,
	}
}

// ToListResponse maps a slice of Pipeline entities to PipelineListResponse DTO.
func (m *PipelineMapper) ToListResponse(
	pipelines []*domain.Pipeline,
	page, pageSize int,
	totalItems int64,
) *dto.PipelineListResponse {
	briefResponses := make([]*dto.PipelineBriefResponse, 0, len(pipelines))
	for _, pipeline := range pipelines {
		briefResponses = append(briefResponses, m.ToBriefResponse(pipeline))
	}

	return &dto.PipelineListResponse{
		Pipelines:  briefResponses,
		Pagination: dto.NewPaginationResponse(page, pageSize, totalItems),
	}
}

// ToStageResponse maps a Stage entity to StageResponse DTO.
func (m *PipelineMapper) ToStageResponse(stage *domain.Stage) *dto.StageResponse {
	if stage == nil {
		return nil
	}

	response := &dto.StageResponse{
		ID:          stage.ID.String(),
		PipelineID:  stage.PipelineID.String(),
		Name:        stage.Name,
		Type:        string(stage.Type),
		Order:       stage.Order,
		Probability: stage.Probability,
		Color:       stage.Color,
		IsActive:    stage.IsActive,
		CreatedAt:   stage.CreatedAt,
		UpdatedAt:   stage.UpdatedAt,
	}

	if stage.Description != "" {
		response.Description = dto.StringPtr(stage.Description)
	}

	if stage.RottenDays > 0 {
		response.RottenDays = dto.IntPtr(stage.RottenDays)
	}

	// Auto actions
	response.AutoActions = m.mapAutoActions(stage.AutoActions)

	return response
}

// ToStageBriefResponse maps a Stage entity to StageBriefResponse DTO.
func (m *PipelineMapper) ToStageBriefResponse(stage *domain.Stage) *dto.StageBriefResponse {
	if stage == nil {
		return nil
	}

	return &dto.StageBriefResponse{
		ID:          stage.ID.String(),
		Name:        stage.Name,
		Type:        string(stage.Type),
		Order:       stage.Order,
		Probability: stage.Probability,
		Color:       stage.Color,
	}
}

// mapStages maps domain Stages to StageResponse DTOs.
func (m *PipelineMapper) mapStages(stages []*domain.Stage) []*dto.StageResponse {
	if len(stages) == 0 {
		return nil
	}

	result := make([]*dto.StageResponse, 0, len(stages))
	for _, stage := range stages {
		if stage.IsActive {
			result = append(result, m.ToStageResponse(stage))
		}
	}

	return result
}

// mapAutoActions maps domain AutoAction to StageAutoActionDTO.
func (m *PipelineMapper) mapAutoActions(actions []domain.AutoAction) []*dto.StageAutoActionDTO {
	if len(actions) == 0 {
		return nil
	}

	result := make([]*dto.StageAutoActionDTO, 0, len(actions))
	for _, action := range actions {
		actionDTO := &dto.StageAutoActionDTO{
			Type:     action.Type,
			Trigger:  "on_enter", // Default trigger
			Config:   action.Config,
			IsActive: true,
		}

		if action.DelayHours > 0 {
			actionDTO.DelayHours = dto.IntPtr(action.DelayHours)
		}

		result = append(result, actionDTO)
	}

	return result
}

// mapCustomFieldDefs maps domain CustomFieldDef to CustomFieldSchemaDTO.
func (m *PipelineMapper) mapCustomFieldDefs(fields []domain.CustomFieldDef) []*dto.CustomFieldSchemaDTO {
	if len(fields) == 0 {
		return nil
	}

	result := make([]*dto.CustomFieldSchemaDTO, 0, len(fields))
	for i, field := range fields {
		fieldDTO := &dto.CustomFieldSchemaDTO{
			Name:         field.Name,
			Label:        field.Label,
			Type:         field.Type,
			Required:     field.Required,
			DisplayOrder: i + 1,
			IsActive:     true,
		}

		if field.Placeholder != "" {
			fieldDTO.Description = dto.StringPtr(field.Placeholder)
		}

		if field.Default != "" {
			fieldDTO.DefaultValue = field.Default
		}

		// Map options for select/multiselect fields
		if len(field.Options) > 0 {
			fieldDTO.Options = make([]dto.CustomFieldOptionDTO, 0, len(field.Options))
			for _, opt := range field.Options {
				fieldDTO.Options = append(fieldDTO.Options, dto.CustomFieldOptionDTO{
					Value: opt,
					Label: opt,
				})
			}
		}

		result = append(result, fieldDTO)
	}

	return result
}

// buildStatistics builds pipeline statistics from the pipeline entity.
func (m *PipelineMapper) buildStatistics(pipeline *domain.Pipeline) *dto.PipelineStatisticsDTO {
	return &dto.PipelineStatisticsDTO{
		TotalOpportunities: pipeline.OpportunityCount,
		TotalValue: dto.MoneyDTO{
			Amount:   pipeline.TotalValue.Amount,
			Currency: pipeline.TotalValue.Currency,
			Display:  formatMoney(pipeline.TotalValue),
		},
		WeightedValue: dto.MoneyDTO{
			Amount:   pipeline.TotalValue.Amount / 2, // Simplified weighted calculation
			Currency: pipeline.TotalValue.Currency,
			Display:  formatMoney(pipeline.TotalValue),
		},
	}
}

// ToStageType maps stage type string to StageType domain type.
func (m *PipelineMapper) ToStageType(stageType string) domain.StageType {
	switch stageType {
	case "open":
		return domain.StageTypeOpen
	case "won":
		return domain.StageTypeWon
	case "lost":
		return domain.StageTypeLost
	case "qualifying":
		return domain.StageTypeQualifying
	case "negotiating":
		return domain.StageTypeNegotiating
	default:
		return domain.StageTypeOpen
	}
}

// ToAutoAction maps StageAutoActionDTO to AutoAction domain type.
func (m *PipelineMapper) ToAutoAction(actionDTO *dto.StageAutoActionDTO) domain.AutoAction {
	action := domain.AutoAction{
		Type:   actionDTO.Type,
		Config: actionDTO.Config,
	}

	if actionDTO.DelayHours != nil {
		action.DelayHours = *actionDTO.DelayHours
	}

	return action
}

// ToCustomFieldDef maps CustomFieldSchemaDTO to CustomFieldDef domain type.
func (m *PipelineMapper) ToCustomFieldDef(fieldDTO *dto.CustomFieldSchemaDTO) domain.CustomFieldDef {
	field := domain.CustomFieldDef{
		Name:     fieldDTO.Name,
		Type:     fieldDTO.Type,
		Label:    fieldDTO.Label,
		Required: fieldDTO.Required,
	}

	if fieldDTO.Description != nil {
		field.Placeholder = *fieldDTO.Description
	}

	if fieldDTO.DefaultValue != nil {
		if defaultStr, ok := fieldDTO.DefaultValue.(string); ok {
			field.Default = defaultStr
		}
	}

	// Map options
	if len(fieldDTO.Options) > 0 {
		field.Options = make([]string, 0, len(fieldDTO.Options))
		for _, opt := range fieldDTO.Options {
			field.Options = append(field.Options, opt.Value)
		}
	}

	return field
}

// ToStageOrderList maps ReorderStagesRequest to a list of stage IDs.
func (m *PipelineMapper) ToStageOrderList(req *dto.ReorderStagesRequest) ([]uuid.UUID, error) {
	stageIDs := make([]uuid.UUID, 0, len(req.StageOrders))
	for _, order := range req.StageOrders {
		id, err := dto.ParseUUIDRequired(order.StageID)
		if err != nil {
			return nil, err
		}
		stageIDs = append(stageIDs, id)
	}
	return stageIDs, nil
}

// ToPipelineBriefDTO creates a PipelineBriefDTO for embedding in other responses.
func (m *PipelineMapper) ToPipelineBriefDTO(pipeline *domain.Pipeline) *dto.PipelineBriefDTO {
	if pipeline == nil {
		return nil
	}

	return &dto.PipelineBriefDTO{
		ID:        pipeline.ID.String(),
		Name:      pipeline.Name,
		IsDefault: pipeline.IsDefault,
	}
}

// ToStageBriefDTO creates a StageBriefDTO for embedding in other responses.
func (m *PipelineMapper) ToStageBriefDTO(stage *domain.Stage) *dto.StageBriefDTO {
	if stage == nil {
		return nil
	}

	return &dto.StageBriefDTO{
		ID:          stage.ID.String(),
		Name:        stage.Name,
		Type:        string(stage.Type),
		Order:       stage.Order,
		Probability: stage.Probability,
		Color:       stage.Color,
	}
}

// ToForecastResponse creates a ForecastResponse DTO.
func (m *PipelineMapper) ToForecastResponse(
	pipelineID *uuid.UUID,
	pipelineName *string,
	startDate, endDate string,
	currency string,
	totalForecast, bestCase, worstCase, committed, pipeline, upside domain.Money,
	byPeriod []dto.ForecastPeriodDTO,
	byOwner []dto.OwnerForecastDTO,
) *dto.ForecastResponse {
	response := &dto.ForecastResponse{
		Currency: currency,
		TotalForecast: dto.MoneyDTO{
			Amount:   totalForecast.Amount,
			Currency: totalForecast.Currency,
			Display:  formatMoney(totalForecast),
		},
		BestCase: dto.MoneyDTO{
			Amount:   bestCase.Amount,
			Currency: bestCase.Currency,
			Display:  formatMoney(bestCase),
		},
		WorstCase: dto.MoneyDTO{
			Amount:   worstCase.Amount,
			Currency: worstCase.Currency,
			Display:  formatMoney(worstCase),
		},
		Committed: dto.MoneyDTO{
			Amount:   committed.Amount,
			Currency: committed.Currency,
			Display:  formatMoney(committed),
		},
		Pipeline: dto.MoneyDTO{
			Amount:   pipeline.Amount,
			Currency: pipeline.Currency,
			Display:  formatMoney(pipeline),
		},
		Upside: dto.MoneyDTO{
			Amount:   upside.Amount,
			Currency: upside.Currency,
			Display:  formatMoney(upside),
		},
		ByPeriod: byPeriod,
		ByOwner:  byOwner,
	}

	if pipelineID != nil {
		pipelineIDStr := pipelineID.String()
		response.PipelineID = &pipelineIDStr
	}

	if pipelineName != nil {
		response.PipelineName = pipelineName
	}

	return response
}

// ToComparisonResponse creates a PipelineComparisonResponse DTO.
func (m *PipelineMapper) ToComparisonResponse(
	pipelines []dto.PipelineComparisonItemDTO,
	period *dto.DateRangeDTO,
) *dto.PipelineComparisonResponse {
	return &dto.PipelineComparisonResponse{
		Pipelines: pipelines,
		Period:    period,
	}
}

// ToTemplateDTO creates a PipelineTemplateDTO.
func (m *PipelineMapper) ToTemplateDTO(
	id, name, description, industry string,
	stages []*dto.StageTemplateDTO,
	winReasons, lossReasons []string,
) *dto.PipelineTemplateDTO {
	return &dto.PipelineTemplateDTO{
		ID:          id,
		Name:        name,
		Description: description,
		Industry:    industry,
		Stages:      stages,
		WinReasons:  winReasons,
		LossReasons: lossReasons,
	}
}

// GetPredefinedTemplates returns predefined pipeline templates.
func (m *PipelineMapper) GetPredefinedTemplates() []*dto.PipelineTemplateDTO {
	return []*dto.PipelineTemplateDTO{
		{
			ID:          "b2b-sales",
			Name:        "B2B Sales Pipeline",
			Description: "Standard B2B sales pipeline for enterprise deals",
			Industry:    "general",
			Stages: []*dto.StageTemplateDTO{
				{Name: "Lead", Description: "Initial lead qualification", Type: "qualifying", Probability: 10, Color: "#3498db"},
				{Name: "Discovery", Description: "Understanding customer needs", Type: "open", Probability: 20, Color: "#9b59b6"},
				{Name: "Demo", Description: "Product demonstration", Type: "open", Probability: 40, Color: "#1abc9c"},
				{Name: "Proposal", Description: "Proposal sent", Type: "negotiating", Probability: 60, Color: "#f39c12"},
				{Name: "Negotiation", Description: "Contract negotiation", Type: "negotiating", Probability: 80, Color: "#e67e22"},
				{Name: "Closed Won", Description: "Deal won", Type: "won", Probability: 100, Color: "#27ae60"},
				{Name: "Closed Lost", Description: "Deal lost", Type: "lost", Probability: 0, Color: "#e74c3c"},
			},
			WinReasons:  []string{"Price", "Product Fit", "Relationship", "Timing", "Features"},
			LossReasons: []string{"Price Too High", "Competitor Won", "No Budget", "No Decision", "Product Limitations"},
		},
		{
			ID:          "saas-sales",
			Name:        "SaaS Sales Pipeline",
			Description: "Optimized for SaaS subscription sales",
			Industry:    "technology",
			Stages: []*dto.StageTemplateDTO{
				{Name: "MQL", Description: "Marketing Qualified Lead", Type: "qualifying", Probability: 5, Color: "#3498db"},
				{Name: "SQL", Description: "Sales Qualified Lead", Type: "qualifying", Probability: 15, Color: "#9b59b6"},
				{Name: "Discovery Call", Description: "Initial discovery call", Type: "open", Probability: 25, Color: "#1abc9c"},
				{Name: "Product Demo", Description: "Product demonstration", Type: "open", Probability: 40, Color: "#f39c12"},
				{Name: "Trial", Description: "Free trial period", Type: "open", Probability: 60, Color: "#e67e22"},
				{Name: "Proposal", Description: "Pricing proposal sent", Type: "negotiating", Probability: 75, Color: "#d35400"},
				{Name: "Contract", Description: "Contract review", Type: "negotiating", Probability: 90, Color: "#c0392b"},
				{Name: "Closed Won", Description: "Deal won", Type: "won", Probability: 100, Color: "#27ae60"},
				{Name: "Closed Lost", Description: "Deal lost", Type: "lost", Probability: 0, Color: "#e74c3c"},
			},
			WinReasons:  []string{"ROI", "Ease of Use", "Integration", "Support", "Price"},
			LossReasons: []string{"Price", "Competitor", "Not Ready", "No Champion", "Integration Issues"},
		},
		{
			ID:          "real-estate",
			Name:        "Real Estate Sales Pipeline",
			Description: "Pipeline for real estate transactions",
			Industry:    "real_estate",
			Stages: []*dto.StageTemplateDTO{
				{Name: "New Inquiry", Description: "Initial property inquiry", Type: "qualifying", Probability: 10, Color: "#3498db"},
				{Name: "Property Viewing", Description: "Scheduled viewing", Type: "open", Probability: 25, Color: "#9b59b6"},
				{Name: "Interested", Description: "Buyer shows interest", Type: "open", Probability: 40, Color: "#1abc9c"},
				{Name: "Offer Made", Description: "Offer submitted", Type: "negotiating", Probability: 60, Color: "#f39c12"},
				{Name: "Negotiation", Description: "Price negotiation", Type: "negotiating", Probability: 75, Color: "#e67e22"},
				{Name: "Under Contract", Description: "Contract signed", Type: "negotiating", Probability: 90, Color: "#d35400"},
				{Name: "Closed", Description: "Sale completed", Type: "won", Probability: 100, Color: "#27ae60"},
				{Name: "Lost", Description: "Sale lost", Type: "lost", Probability: 0, Color: "#e74c3c"},
			},
			WinReasons:  []string{"Price", "Location", "Condition", "Agent Relationship", "Financing"},
			LossReasons: []string{"Price Too High", "Lost to Competitor", "Financing Issues", "Inspection Issues", "Cold Feet"},
		},
	}
}
