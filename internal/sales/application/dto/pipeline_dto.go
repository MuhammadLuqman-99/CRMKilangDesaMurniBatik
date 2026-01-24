package dto

import (
	"time"
)

// ============================================================================
// Pipeline Request DTOs
// ============================================================================

// CreatePipelineRequest represents a request to create a new pipeline.
type CreatePipelineRequest struct {
	// Basic Information
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`

	// Settings
	IsDefault          bool `json:"is_default"`
	AllowSkipStages    bool `json:"allow_skip_stages"`
	RequireWonReason   bool `json:"require_won_reason"`
	RequireLostReason  bool `json:"require_lost_reason"`

	// Stages
	Stages []*CreateStageRequest `json:"stages,omitempty" validate:"omitempty,max=20,dive"`

	// Win/Loss Reasons
	WinReasons  []string `json:"win_reasons,omitempty" validate:"omitempty,max=20,dive,max=100"`
	LossReasons []string `json:"loss_reasons,omitempty" validate:"omitempty,max=20,dive,max=100"`

	// Custom Fields Schema
	CustomFieldsSchema []*CustomFieldSchemaDTO `json:"custom_fields_schema,omitempty" validate:"omitempty,max=50,dive"`

	// Currency settings
	DefaultCurrency string `json:"default_currency,omitempty" validate:"omitempty,len=3"`
}

// UpdatePipelineRequest represents a request to update a pipeline.
type UpdatePipelineRequest struct {
	// Basic Information
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`

	// Settings
	IsDefault          *bool `json:"is_default,omitempty"`
	AllowSkipStages    *bool `json:"allow_skip_stages,omitempty"`
	RequireWonReason   *bool `json:"require_won_reason,omitempty"`
	RequireLostReason  *bool `json:"require_lost_reason,omitempty"`

	// Win/Loss Reasons
	WinReasons  []string `json:"win_reasons,omitempty" validate:"omitempty,max=20,dive,max=100"`
	LossReasons []string `json:"loss_reasons,omitempty" validate:"omitempty,max=20,dive,max=100"`

	// Custom Fields Schema
	CustomFieldsSchema []*CustomFieldSchemaDTO `json:"custom_fields_schema,omitempty" validate:"omitempty,max=50,dive"`

	// Currency settings
	DefaultCurrency *string `json:"default_currency,omitempty" validate:"omitempty,len=3"`

	// Version for optimistic locking
	Version int `json:"version" validate:"required,min=1"`
}

// CreateStageRequest represents a request to create a new stage.
type CreateStageRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Type        string  `json:"type" validate:"required,oneof=open won lost qualifying negotiating"`
	Order       int     `json:"order" validate:"required,min=0"`
	Probability int     `json:"probability" validate:"required,min=0,max=100"`
	Color       string  `json:"color" validate:"required,hexcolor"`

	// Settings
	RottenDays       *int  `json:"rotten_days,omitempty" validate:"omitempty,min=1"`
	RequiredFields   []string `json:"required_fields,omitempty" validate:"omitempty,max=20,dive,max=50"`
	AutoActions      []*StageAutoActionDTO `json:"auto_actions,omitempty" validate:"omitempty,max=10,dive"`
}

// UpdateStageRequest represents a request to update a stage.
type UpdateStageRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Probability *int    `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor"`

	// Settings
	RottenDays       *int     `json:"rotten_days,omitempty" validate:"omitempty,min=1"`
	RequiredFields   []string `json:"required_fields,omitempty" validate:"omitempty,max=20,dive,max=50"`
	AutoActions      []*StageAutoActionDTO `json:"auto_actions,omitempty" validate:"omitempty,max=10,dive"`

	// Activation
	IsActive *bool `json:"is_active,omitempty"`
}

// ReorderStagesRequest represents a request to reorder stages.
type ReorderStagesRequest struct {
	StageOrders []StageOrderDTO `json:"stage_orders" validate:"required,min=1,max=20,dive"`
}

// StageOrderDTO represents a stage order.
type StageOrderDTO struct {
	StageID string `json:"stage_id" validate:"required,uuid"`
	Order   int    `json:"order" validate:"required,min=0"`
}

// AddStageRequest represents a request to add a stage to a pipeline.
type AddStageRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=100"`
	Description    *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Type           string  `json:"type" validate:"required,oneof=open won lost qualifying negotiating"`
	InsertAfterID  *string `json:"insert_after_id,omitempty" validate:"omitempty,uuid"`
	InsertBeforeID *string `json:"insert_before_id,omitempty" validate:"omitempty,uuid"`
	Probability    int     `json:"probability" validate:"required,min=0,max=100"`
	Color          string  `json:"color" validate:"required,hexcolor"`

	// Settings
	RottenDays     *int     `json:"rotten_days,omitempty" validate:"omitempty,min=1"`
	RequiredFields []string `json:"required_fields,omitempty" validate:"omitempty,max=20,dive,max=50"`
}

// PipelineFilterRequest represents filter options for listing pipelines.
type PipelineFilterRequest struct {
	// Status
	IsActive *bool `json:"is_active,omitempty"`

	// Search
	SearchQuery string `json:"search_query,omitempty" validate:"omitempty,max=200"`

	// Include counts
	IncludeOpportunityCounts bool `json:"include_opportunity_counts,omitempty"`

	// Pagination
	Page     int `json:"page,omitempty" validate:"omitempty,min=1"`
	PageSize int `json:"page_size,omitempty" validate:"omitempty,min=1,max=100"`

	// Sorting
	SortBy    string `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at updated_at name"`
	SortOrder string `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// ============================================================================
// Pipeline Response DTOs
// ============================================================================

// PipelineResponse represents a pipeline in API responses.
type PipelineResponse struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`

	// Basic Information
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`

	// Status
	IsActive  bool `json:"is_active"`
	IsDefault bool `json:"is_default"`

	// Settings
	AllowSkipStages   bool `json:"allow_skip_stages"`
	RequireWonReason  bool `json:"require_won_reason"`
	RequireLostReason bool `json:"require_lost_reason"`

	// Stages
	Stages     []*StageResponse `json:"stages"`
	StageCount int              `json:"stage_count"`

	// Win/Loss Reasons
	WinReasons  []string `json:"win_reasons,omitempty"`
	LossReasons []string `json:"loss_reasons,omitempty"`

	// Custom Fields Schema
	CustomFieldsSchema []*CustomFieldSchemaDTO `json:"custom_fields_schema,omitempty"`

	// Currency settings
	DefaultCurrency string `json:"default_currency"`

	// Statistics (optional, based on request)
	Statistics *PipelineStatisticsDTO `json:"statistics,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	Version   int       `json:"version"`
}

// PipelineBriefResponse represents a brief pipeline summary.
type PipelineBriefResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	IsActive        bool   `json:"is_active"`
	IsDefault       bool   `json:"is_default"`
	StageCount      int    `json:"stage_count"`
	OpportunityCount int64  `json:"opportunity_count,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// PipelineListResponse represents a paginated list of pipelines.
type PipelineListResponse struct {
	Pipelines  []*PipelineBriefResponse `json:"pipelines"`
	Pagination PaginationResponse       `json:"pagination"`
}

// StageResponse represents a stage in API responses.
type StageResponse struct {
	ID          string  `json:"id"`
	PipelineID  string  `json:"pipeline_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Type        string  `json:"type"`
	Order       int     `json:"order"`
	Probability int     `json:"probability"`
	Color       string  `json:"color"`
	IsActive    bool    `json:"is_active"`

	// Settings
	RottenDays     *int                  `json:"rotten_days,omitempty"`
	RequiredFields []string              `json:"required_fields,omitempty"`
	AutoActions    []*StageAutoActionDTO `json:"auto_actions,omitempty"`

	// Statistics (optional)
	OpportunityCount *int64   `json:"opportunity_count,omitempty"`
	TotalValue       *MoneyDTO `json:"total_value,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StageBriefResponse represents a brief stage summary.
type StageBriefResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Order       int    `json:"order"`
	Probability int    `json:"probability"`
	Color       string `json:"color"`
}

// ============================================================================
// Supporting DTOs
// ============================================================================

// StageAutoActionDTO represents an automatic action for a stage.
type StageAutoActionDTO struct {
	Type       string                 `json:"type" validate:"required,oneof=send_email create_task update_field assign_owner notify_webhook"`
	Trigger    string                 `json:"trigger" validate:"required,oneof=on_enter on_exit after_duration"`
	Config     map[string]interface{} `json:"config" validate:"required"`
	DelayHours *int                   `json:"delay_hours,omitempty" validate:"omitempty,min=0"`
	IsActive   bool                   `json:"is_active"`
}

// CustomFieldSchemaDTO represents a custom field schema definition.
type CustomFieldSchemaDTO struct {
	Name         string                 `json:"name" validate:"required,min=1,max=50"`
	Label        string                 `json:"label" validate:"required,min=1,max=100"`
	Type         string                 `json:"type" validate:"required,oneof=text number date datetime boolean select multiselect url email phone currency percentage"`
	Description  *string                `json:"description,omitempty" validate:"omitempty,max=500"`
	Required     bool                   `json:"required"`
	DefaultValue interface{}            `json:"default_value,omitempty"`
	Options      []CustomFieldOptionDTO `json:"options,omitempty" validate:"omitempty,max=50,dive"`
	Validation   *CustomFieldValidation `json:"validation,omitempty"`
	DisplayOrder int                    `json:"display_order"`
	IsActive     bool                   `json:"is_active"`
}

// CustomFieldOptionDTO represents an option for select/multiselect fields.
type CustomFieldOptionDTO struct {
	Value string `json:"value" validate:"required,max=100"`
	Label string `json:"label" validate:"required,max=100"`
	Color *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// CustomFieldValidation represents validation rules for a custom field.
type CustomFieldValidation struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Pattern   *string  `json:"pattern,omitempty"`
}

// PipelineStatisticsDTO represents pipeline statistics.
type PipelineStatisticsDTO struct {
	TotalOpportunities   int64              `json:"total_opportunities"`
	OpenOpportunities    int64              `json:"open_opportunities"`
	WonOpportunities     int64              `json:"won_opportunities"`
	LostOpportunities    int64              `json:"lost_opportunities"`
	TotalValue           MoneyDTO           `json:"total_value"`
	WeightedValue        MoneyDTO           `json:"weighted_value"`
	WinRate              float64            `json:"win_rate"`
	AverageSalesCycle    int                `json:"average_sales_cycle_days"`
	AverageDealSize      MoneyDTO           `json:"average_deal_size"`
	StageDistribution    map[string]int64   `json:"stage_distribution"`
	ConversionRates      map[string]float64 `json:"conversion_rates"`
}

// PipelineComparisonRequest represents a request to compare pipelines.
type PipelineComparisonRequest struct {
	PipelineIDs []string `json:"pipeline_ids" validate:"required,min=2,max=5,dive,uuid"`
	StartDate   *string  `json:"start_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	EndDate     *string  `json:"end_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// PipelineComparisonResponse represents a pipeline comparison result.
type PipelineComparisonResponse struct {
	Pipelines []PipelineComparisonItemDTO `json:"pipelines"`
	Period    *DateRangeDTO               `json:"period,omitempty"`
}

// PipelineComparisonItemDTO represents a single pipeline in comparison.
type PipelineComparisonItemDTO struct {
	PipelineID         string   `json:"pipeline_id"`
	PipelineName       string   `json:"pipeline_name"`
	TotalOpportunities int64    `json:"total_opportunities"`
	WonOpportunities   int64    `json:"won_opportunities"`
	LostOpportunities  int64    `json:"lost_opportunities"`
	WinRate            float64  `json:"win_rate"`
	TotalValue         MoneyDTO `json:"total_value"`
	WeightedValue      MoneyDTO `json:"weighted_value"`
	AverageSalesCycle  int      `json:"average_sales_cycle_days"`
	AverageDealSize    MoneyDTO `json:"average_deal_size"`
}

// DateRangeDTO represents a date range.
type DateRangeDTO struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// PipelineTemplateDTO represents a pipeline template.
type PipelineTemplateDTO struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Industry    string                `json:"industry"`
	Stages      []*StageTemplateDTO   `json:"stages"`
	WinReasons  []string              `json:"win_reasons"`
	LossReasons []string              `json:"loss_reasons"`
}

// StageTemplateDTO represents a stage in a pipeline template.
type StageTemplateDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Probability int    `json:"probability"`
	Color       string `json:"color"`
}

// CreateFromTemplateRequest represents a request to create a pipeline from template.
type CreateFromTemplateRequest struct {
	TemplateID  string  `json:"template_id" validate:"required"`
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	IsDefault   bool    `json:"is_default"`
}

// ClonePipelineRequest represents a request to clone a pipeline.
type ClonePipelineRequest struct {
	SourcePipelineID string  `json:"source_pipeline_id" validate:"required,uuid"`
	Name             string  `json:"name" validate:"required,min=1,max=100"`
	Description      *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	IncludeCustomFields bool `json:"include_custom_fields"`
}

// ForecastRequest represents a request for pipeline forecast.
type ForecastRequest struct {
	PipelineID  *string `json:"pipeline_id,omitempty" validate:"omitempty,uuid"`
	StartDate   string  `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate     string  `json:"end_date" validate:"required,datetime=2006-01-02"`
	Currency    string  `json:"currency" validate:"required,len=3"`
	GroupBy     string  `json:"group_by,omitempty" validate:"omitempty,oneof=day week month quarter"`
}

// ForecastResponse represents a pipeline forecast.
type ForecastResponse struct {
	PipelineID    *string           `json:"pipeline_id,omitempty"`
	PipelineName  *string           `json:"pipeline_name,omitempty"`
	Period        DateRangeDTO      `json:"period"`
	Currency      string            `json:"currency"`
	TotalForecast MoneyDTO          `json:"total_forecast"`
	BestCase      MoneyDTO          `json:"best_case"`
	WorstCase     MoneyDTO          `json:"worst_case"`
	Committed     MoneyDTO          `json:"committed"` // High probability deals
	Pipeline      MoneyDTO          `json:"pipeline"`  // Medium probability
	Upside        MoneyDTO          `json:"upside"`    // Low probability
	ByPeriod      []ForecastPeriodDTO `json:"by_period"`
	ByOwner       []OwnerForecastDTO  `json:"by_owner,omitempty"`
}

// ForecastPeriodDTO represents forecast for a specific period.
type ForecastPeriodDTO struct {
	Period           string   `json:"period"` // date string based on GroupBy
	ExpectedRevenue  MoneyDTO `json:"expected_revenue"`
	WeightedRevenue  MoneyDTO `json:"weighted_revenue"`
	OpportunityCount int64    `json:"opportunity_count"`
	Committed        MoneyDTO `json:"committed"`
	Pipeline         MoneyDTO `json:"pipeline"`
	Upside           MoneyDTO `json:"upside"`
}

// OwnerForecastDTO represents forecast by owner.
type OwnerForecastDTO struct {
	OwnerID          string   `json:"owner_id"`
	OwnerName        string   `json:"owner_name"`
	ExpectedRevenue  MoneyDTO `json:"expected_revenue"`
	WeightedRevenue  MoneyDTO `json:"weighted_revenue"`
	OpportunityCount int64    `json:"opportunity_count"`
	Committed        MoneyDTO `json:"committed"`
	Pipeline         MoneyDTO `json:"pipeline"`
}
