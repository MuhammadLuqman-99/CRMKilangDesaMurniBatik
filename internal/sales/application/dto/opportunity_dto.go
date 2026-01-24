package dto

import (
	"time"
)

// ============================================================================
// Opportunity Request DTOs
// ============================================================================

// CreateOpportunityRequest represents a request to create a new opportunity.
type CreateOpportunityRequest struct {
	// Basic Information
	Name        string  `json:"name" validate:"required,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=5000"`

	// Pipeline and Stage
	PipelineID string  `json:"pipeline_id" validate:"required,uuid"`
	StageID    *string `json:"stage_id,omitempty" validate:"omitempty,uuid"`

	// Value
	Amount   int64  `json:"amount" validate:"required,min=0"`
	Currency string `json:"currency" validate:"required,len=3"`

	// Probability
	Probability *int `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`

	// Dates
	ExpectedCloseDate string  `json:"expected_close_date" validate:"required,datetime=2006-01-02"`
	ActualCloseDate   *string `json:"actual_close_date,omitempty" validate:"omitempty,datetime=2006-01-02"`

	// Relationships
	CustomerID *string `json:"customer_id,omitempty" validate:"omitempty,uuid"`
	LeadID     *string `json:"lead_id,omitempty" validate:"omitempty,uuid"`

	// Contacts
	PrimaryContactID *string                         `json:"primary_contact_id,omitempty" validate:"omitempty,uuid"`
	Contacts         []*OpportunityContactRequestDTO `json:"contacts,omitempty" validate:"omitempty,max=20,dive"`

	// Products
	Products []*OpportunityProductRequestDTO `json:"products,omitempty" validate:"omitempty,max=100,dive"`

	// Assignment
	OwnerID *string `json:"owner_id,omitempty" validate:"omitempty,uuid"`

	// Source and Attribution
	Source        string  `json:"source,omitempty" validate:"omitempty,max=100"`
	SourceDetails *string `json:"source_details,omitempty" validate:"omitempty,max=500"`
	CampaignID    *string `json:"campaign_id,omitempty" validate:"omitempty,uuid"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`

	// Competitors
	Competitors []*CompetitorDTO `json:"competitors,omitempty" validate:"omitempty,max=10,dive"`

	// Notes
	Notes *string `json:"notes,omitempty" validate:"omitempty,max=5000"`
}

// UpdateOpportunityRequest represents a request to update an opportunity.
type UpdateOpportunityRequest struct {
	// Basic Information
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=5000"`

	// Value
	Amount   *int64  `json:"amount,omitempty" validate:"omitempty,min=0"`
	Currency *string `json:"currency,omitempty" validate:"omitempty,len=3"`

	// Probability
	Probability *int `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`

	// Dates
	ExpectedCloseDate *string `json:"expected_close_date,omitempty" validate:"omitempty,datetime=2006-01-02"`

	// Relationships
	CustomerID       *string `json:"customer_id,omitempty" validate:"omitempty,uuid"`
	PrimaryContactID *string `json:"primary_contact_id,omitempty" validate:"omitempty,uuid"`

	// Source
	Source        *string `json:"source,omitempty" validate:"omitempty,max=100"`
	SourceDetails *string `json:"source_details,omitempty" validate:"omitempty,max=500"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`

	// Notes
	Notes *string `json:"notes,omitempty" validate:"omitempty,max=5000"`

	// Version for optimistic locking
	Version int `json:"version" validate:"required,min=1"`
}

// MoveStageRequest represents a request to move an opportunity to a different stage.
type MoveStageRequest struct {
	StageID     string  `json:"stage_id" validate:"required,uuid"`
	Probability *int    `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`
	Notes       *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// WinOpportunityRequest represents a request to mark an opportunity as won.
type WinOpportunityRequest struct {
	WonStageID    *string `json:"won_stage_id,omitempty" validate:"omitempty,uuid"`
	WonReason     string  `json:"won_reason" validate:"required,max=200"`
	WonNotes      *string `json:"won_notes,omitempty" validate:"omitempty,max=2000"`
	ActualAmount  *int64  `json:"actual_amount,omitempty" validate:"omitempty,min=0"`
	ActualCloseDate *string `json:"actual_close_date,omitempty" validate:"omitempty,datetime=2006-01-02"`

	// Deal creation
	CreateDeal     bool    `json:"create_deal"`
	DealOwnerID    *string `json:"deal_owner_id,omitempty" validate:"omitempty,uuid"`
	PaymentTerms   *string `json:"payment_terms,omitempty" validate:"omitempty,max=100"`
	ContractTerms  *string `json:"contract_terms,omitempty" validate:"omitempty,max=2000"`
}

// LoseOpportunityRequest represents a request to mark an opportunity as lost.
type LoseOpportunityRequest struct {
	LostStageID    *string `json:"lost_stage_id,omitempty" validate:"omitempty,uuid"`
	LostReason     string  `json:"lost_reason" validate:"required,max=200"`
	LostNotes      *string `json:"lost_notes,omitempty" validate:"omitempty,max=2000"`
	CompetitorID   *string `json:"competitor_id,omitempty" validate:"omitempty,uuid"`
	CompetitorName *string `json:"competitor_name,omitempty" validate:"omitempty,max=200"`
}

// ReopenOpportunityRequest represents a request to reopen a closed opportunity.
type ReopenOpportunityRequest struct {
	StageID           string  `json:"stage_id" validate:"required,uuid"`
	ExpectedCloseDate string  `json:"expected_close_date" validate:"required,datetime=2006-01-02"`
	Probability       *int    `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`
	Notes             *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// AddProductRequest represents a request to add a product to an opportunity.
type AddProductRequest struct {
	ProductID       string  `json:"product_id" validate:"required,uuid"`
	ProductName     string  `json:"product_name" validate:"required,max=200"`
	Quantity        int     `json:"quantity" validate:"required,min=1"`
	UnitPrice       int64   `json:"unit_price" validate:"required,min=0"`
	Currency        string  `json:"currency" validate:"required,len=3"`
	DiscountPercent *int    `json:"discount_percent,omitempty" validate:"omitempty,min=0,max=100"`
	DiscountAmount  *int64  `json:"discount_amount,omitempty" validate:"omitempty,min=0"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// UpdateProductRequest represents a request to update a product in an opportunity.
type UpdateProductRequest struct {
	Quantity        *int    `json:"quantity,omitempty" validate:"omitempty,min=1"`
	UnitPrice       *int64  `json:"unit_price,omitempty" validate:"omitempty,min=0"`
	DiscountPercent *int    `json:"discount_percent,omitempty" validate:"omitempty,min=0,max=100"`
	DiscountAmount  *int64  `json:"discount_amount,omitempty" validate:"omitempty,min=0"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// AddContactRequest represents a request to add a contact to an opportunity.
type AddContactRequest struct {
	ContactID  string  `json:"contact_id" validate:"required,uuid"`
	Role       string  `json:"role" validate:"required,oneof=decision_maker influencer evaluator champion blocker user technical_contact billing_contact other"`
	IsPrimary  bool    `json:"is_primary"`
	Notes      *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// UpdateContactRequest represents a request to update a contact in an opportunity.
type UpdateContactRequest struct {
	Role      *string `json:"role,omitempty" validate:"omitempty,oneof=decision_maker influencer evaluator champion blocker user technical_contact billing_contact other"`
	IsPrimary *bool   `json:"is_primary,omitempty"`
	Notes     *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// AddCompetitorRequest represents a request to add a competitor to an opportunity.
type AddCompetitorRequest struct {
	Name      string  `json:"name" validate:"required,max=200"`
	Website   *string `json:"website,omitempty" validate:"omitempty,url,max=500"`
	Strengths *string `json:"strengths,omitempty" validate:"omitempty,max=1000"`
	Weaknesses *string `json:"weaknesses,omitempty" validate:"omitempty,max=1000"`
	ThreatLevel string `json:"threat_level" validate:"required,oneof=low medium high"`
	Notes      *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// AssignOpportunityRequest represents a request to assign an opportunity.
type AssignOpportunityRequest struct {
	OwnerID string  `json:"owner_id" validate:"required,uuid"`
	Notes   *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// BulkMoveStageRequest represents a request to bulk move opportunities to a stage.
type BulkMoveStageRequest struct {
	OpportunityIDs []string `json:"opportunity_ids" validate:"required,min=1,max=100,dive,uuid"`
	StageID        string   `json:"stage_id" validate:"required,uuid"`
	Notes          *string  `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// BulkAssignOpportunitiesRequest represents a request to bulk assign opportunities.
type BulkAssignOpportunitiesRequest struct {
	OpportunityIDs []string `json:"opportunity_ids" validate:"required,min=1,max=100,dive,uuid"`
	OwnerID        string   `json:"owner_id" validate:"required,uuid"`
	Notes          *string  `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// OpportunityFilterRequest represents filter options for listing opportunities.
type OpportunityFilterRequest struct {
	// Status filters
	Statuses []string `json:"statuses,omitempty" validate:"omitempty,dive,oneof=open won lost"`

	// Pipeline and Stage filters
	PipelineIDs []string `json:"pipeline_ids,omitempty" validate:"omitempty,dive,uuid"`
	StageIDs    []string `json:"stage_ids,omitempty" validate:"omitempty,dive,uuid"`

	// Relationship filters
	CustomerIDs []string `json:"customer_ids,omitempty" validate:"omitempty,dive,uuid"`
	ContactIDs  []string `json:"contact_ids,omitempty" validate:"omitempty,dive,uuid"`
	OwnerIDs    []string `json:"owner_ids,omitempty" validate:"omitempty,dive,uuid"`
	LeadID      *string  `json:"lead_id,omitempty" validate:"omitempty,uuid"`

	// Value filters
	MinAmount *int64  `json:"min_amount,omitempty" validate:"omitempty,min=0"`
	MaxAmount *int64  `json:"max_amount,omitempty" validate:"omitempty,min=0"`
	Currency  *string `json:"currency,omitempty" validate:"omitempty,len=3"`

	// Probability filter
	MinProbability *int `json:"min_probability,omitempty" validate:"omitempty,min=0,max=100"`
	MaxProbability *int `json:"max_probability,omitempty" validate:"omitempty,min=0,max=100"`

	// Date filters
	ExpectedCloseDateAfter  *string `json:"expected_close_date_after,omitempty" validate:"omitempty,datetime=2006-01-02"`
	ExpectedCloseDateBefore *string `json:"expected_close_date_before,omitempty" validate:"omitempty,datetime=2006-01-02"`
	CreatedAfter            *string `json:"created_after,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	CreatedBefore           *string `json:"created_before,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`

	// Search
	SearchQuery string `json:"search_query,omitempty" validate:"omitempty,max=200"`

	// Tags
	Tags []string `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`

	// Product filter
	ProductIDs []string `json:"product_ids,omitempty" validate:"omitempty,dive,uuid"`

	// Source filter
	Sources []string `json:"sources,omitempty" validate:"omitempty,dive,max=100"`

	// Closing filters
	ClosingThisWeek   bool `json:"closing_this_week,omitempty"`
	ClosingThisMonth  bool `json:"closing_this_month,omitempty"`
	ClosingThisQuarter bool `json:"closing_this_quarter,omitempty"`
	Overdue           bool `json:"overdue,omitempty"`

	// Pagination
	Page     int `json:"page,omitempty" validate:"omitempty,min=1"`
	PageSize int `json:"page_size,omitempty" validate:"omitempty,min=1,max=100"`

	// Sorting
	SortBy    string `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at updated_at expected_close_date amount probability name"`
	SortOrder string `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// ============================================================================
// Opportunity Response DTOs
// ============================================================================

// OpportunityResponse represents an opportunity in API responses.
type OpportunityResponse struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`

	// Basic Information
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`

	// Status
	Status string `json:"status"` // open, won, lost

	// Pipeline and Stage
	PipelineID   string            `json:"pipeline_id"`
	Pipeline     *PipelineBriefDTO `json:"pipeline,omitempty"`
	StageID      string            `json:"stage_id"`
	Stage        *StageBriefDTO    `json:"stage,omitempty"`
	StageHistory []*StageHistoryDTO `json:"stage_history,omitempty"`

	// Value
	Amount         MoneyDTO `json:"amount"`
	WeightedAmount MoneyDTO `json:"weighted_amount"`

	// Probability
	Probability int `json:"probability"`

	// Dates
	ExpectedCloseDate time.Time  `json:"expected_close_date"`
	ActualCloseDate   *time.Time `json:"actual_close_date,omitempty"`

	// Relationships
	CustomerID *string          `json:"customer_id,omitempty"`
	Customer   *CustomerBriefDTO `json:"customer,omitempty"`
	LeadID     *string          `json:"lead_id,omitempty"`
	Lead       *LeadBriefResponse `json:"lead,omitempty"`

	// Contacts
	PrimaryContactID *string                           `json:"primary_contact_id,omitempty"`
	PrimaryContact   *ContactBriefDTO                  `json:"primary_contact,omitempty"`
	Contacts         []*OpportunityContactResponseDTO `json:"contacts,omitempty"`

	// Products
	Products     []*OpportunityProductResponseDTO `json:"products,omitempty"`
	ProductCount int                              `json:"product_count"`

	// Assignment
	OwnerID string        `json:"owner_id"`
	Owner   *UserBriefDTO `json:"owner,omitempty"`

	// Source and Attribution
	Source        string  `json:"source,omitempty"`
	SourceDetails *string `json:"source_details,omitempty"`
	CampaignID    *string `json:"campaign_id,omitempty"`

	// Win/Loss Information
	WonAt          *time.Time `json:"won_at,omitempty"`
	WonBy          *string    `json:"won_by,omitempty"`
	WonReason      *string    `json:"won_reason,omitempty"`
	WonNotes       *string    `json:"won_notes,omitempty"`
	LostAt         *time.Time `json:"lost_at,omitempty"`
	LostBy         *string    `json:"lost_by,omitempty"`
	LostReason     *string    `json:"lost_reason,omitempty"`
	LostNotes      *string    `json:"lost_notes,omitempty"`
	CompetitorID   *string    `json:"competitor_id,omitempty"`
	CompetitorName *string    `json:"competitor_name,omitempty"`

	// Competitors
	Competitors []*CompetitorDTO `json:"competitors,omitempty"`

	// Deal Information (if converted)
	DealID *string `json:"deal_id,omitempty"`

	// Additional Information
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	Notes        *string                `json:"notes,omitempty"`

	// Activity Summary
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
	ActivityCount  int        `json:"activity_count"`
	DaysInStage    int        `json:"days_in_stage"`
	DaysOpen       int        `json:"days_open"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	Version   int       `json:"version"`
}

// OpportunityBriefResponse represents a brief opportunity summary.
type OpportunityBriefResponse struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Status            string     `json:"status"`
	Amount            MoneyDTO   `json:"amount"`
	WeightedAmount    MoneyDTO   `json:"weighted_amount"`
	Probability       int        `json:"probability"`
	StageID           string     `json:"stage_id"`
	StageName         string     `json:"stage_name"`
	ExpectedCloseDate time.Time  `json:"expected_close_date"`
	CustomerID        *string    `json:"customer_id,omitempty"`
	CustomerName      *string    `json:"customer_name,omitempty"`
	OwnerID           string     `json:"owner_id"`
	OwnerName         string     `json:"owner_name"`
	DaysOpen          int        `json:"days_open"`
	CreatedAt         time.Time  `json:"created_at"`
}

// OpportunityListResponse represents a paginated list of opportunities.
type OpportunityListResponse struct {
	Opportunities []*OpportunityBriefResponse `json:"opportunities"`
	Pagination    PaginationResponse          `json:"pagination"`
	Summary       *OpportunitySummaryDTO      `json:"summary,omitempty"`
}

// OpportunitySummaryDTO represents a summary of opportunities in a list.
type OpportunitySummaryDTO struct {
	TotalCount       int64    `json:"total_count"`
	TotalValue       MoneyDTO `json:"total_value"`
	WeightedValue    MoneyDTO `json:"weighted_value"`
	AverageProbability float64 `json:"average_probability"`
	AverageDaysOpen  float64  `json:"average_days_open"`
}

// OpportunityWinResponse represents the result of winning an opportunity.
type OpportunityWinResponse struct {
	OpportunityID string  `json:"opportunity_id"`
	Status        string  `json:"status"`
	DealID        *string `json:"deal_id,omitempty"`
	Message       string  `json:"message"`
}

// OpportunityLoseResponse represents the result of losing an opportunity.
type OpportunityLoseResponse struct {
	OpportunityID string `json:"opportunity_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

// ============================================================================
// Supporting DTOs
// ============================================================================

// OpportunityContactRequestDTO represents a contact to add to an opportunity.
type OpportunityContactRequestDTO struct {
	ContactID string  `json:"contact_id" validate:"required,uuid"`
	Role      string  `json:"role" validate:"required,oneof=decision_maker influencer evaluator champion blocker user technical_contact billing_contact other"`
	IsPrimary bool    `json:"is_primary"`
	Notes     *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// OpportunityContactResponseDTO represents a contact in an opportunity.
type OpportunityContactResponseDTO struct {
	ContactID   string           `json:"contact_id"`
	Contact     *ContactBriefDTO `json:"contact,omitempty"`
	Role        string           `json:"role"`
	IsPrimary   bool             `json:"is_primary"`
	Notes       *string          `json:"notes,omitempty"`
	AddedAt     time.Time        `json:"added_at"`
}

// OpportunityProductRequestDTO represents a product to add to an opportunity.
type OpportunityProductRequestDTO struct {
	ProductID       string  `json:"product_id" validate:"required,uuid"`
	ProductName     string  `json:"product_name" validate:"required,max=200"`
	Quantity        int     `json:"quantity" validate:"required,min=1"`
	UnitPrice       int64   `json:"unit_price" validate:"required,min=0"`
	Currency        string  `json:"currency" validate:"required,len=3"`
	DiscountPercent *int    `json:"discount_percent,omitempty" validate:"omitempty,min=0,max=100"`
	DiscountAmount  *int64  `json:"discount_amount,omitempty" validate:"omitempty,min=0"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// OpportunityProductResponseDTO represents a product in an opportunity.
type OpportunityProductResponseDTO struct {
	ID              string   `json:"id"`
	ProductID       string   `json:"product_id"`
	ProductName     string   `json:"product_name"`
	Quantity        int      `json:"quantity"`
	UnitPrice       MoneyDTO `json:"unit_price"`
	DiscountPercent int      `json:"discount_percent"`
	DiscountAmount  MoneyDTO `json:"discount_amount"`
	TotalPrice      MoneyDTO `json:"total_price"`
	Description     *string  `json:"description,omitempty"`
}

// StageHistoryDTO represents a stage transition in opportunity history.
type StageHistoryDTO struct {
	StageID     string    `json:"stage_id"`
	StageName   string    `json:"stage_name"`
	EnteredAt   time.Time `json:"entered_at"`
	ExitedAt    *time.Time `json:"exited_at,omitempty"`
	Duration    *int      `json:"duration_seconds,omitempty"`
	ChangedBy   string    `json:"changed_by"`
	Notes       *string   `json:"notes,omitempty"`
}

// CompetitorDTO represents a competitor.
type CompetitorDTO struct {
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name" validate:"required,max=200"`
	Website     *string `json:"website,omitempty" validate:"omitempty,url,max=500"`
	Strengths   *string `json:"strengths,omitempty" validate:"omitempty,max=1000"`
	Weaknesses  *string `json:"weaknesses,omitempty" validate:"omitempty,max=1000"`
	ThreatLevel string  `json:"threat_level" validate:"required,oneof=low medium high"`
	Notes       *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// PipelineBriefDTO represents a brief pipeline summary.
type PipelineBriefDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// StageBriefDTO represents a brief stage summary.
type StageBriefDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Order       int    `json:"order"`
	Probability int    `json:"probability"`
	Color       string `json:"color"`
}

// CustomerBriefDTO represents a brief customer summary.
type CustomerBriefDTO struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Code   string  `json:"code"`
	Type   string  `json:"type"`
	Status string  `json:"status"`
	Email  *string `json:"email,omitempty"`
}

// ContactBriefDTO represents a brief contact summary.
type ContactBriefDTO struct {
	ID        string  `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	FullName  string  `json:"full_name"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone,omitempty"`
	JobTitle  *string `json:"job_title,omitempty"`
}

// OpportunityStatisticsResponse represents opportunity statistics.
type OpportunityStatisticsResponse struct {
	TotalOpportunities  int64              `json:"total_opportunities"`
	OpenOpportunities   int64              `json:"open_opportunities"`
	WonOpportunities    int64              `json:"won_opportunities"`
	LostOpportunities   int64              `json:"lost_opportunities"`
	TotalPipelineValue  MoneyDTO           `json:"total_pipeline_value"`
	WeightedPipelineValue MoneyDTO         `json:"weighted_pipeline_value"`
	WinRate             float64            `json:"win_rate"`
	AverageDealSize     MoneyDTO           `json:"average_deal_size"`
	AverageSalesCycle   int                `json:"average_sales_cycle_days"`
	ByStage             map[string]int64   `json:"by_stage"`
	ByStatus            map[string]int64   `json:"by_status"`
	ClosingThisMonth    int64              `json:"closing_this_month"`
	ClosingThisQuarter  int64              `json:"closing_this_quarter"`
}

// PipelineAnalyticsResponse represents pipeline analytics.
type PipelineAnalyticsResponse struct {
	PipelineID        string                    `json:"pipeline_id"`
	PipelineName      string                    `json:"pipeline_name"`
	TotalOpportunities int64                    `json:"total_opportunities"`
	TotalValue        MoneyDTO                  `json:"total_value"`
	WeightedValue     MoneyDTO                  `json:"weighted_value"`
	WinRate           float64                   `json:"win_rate"`
	AverageSalesCycle int                       `json:"average_sales_cycle_days"`
	Stages            []*StageAnalyticsDTO      `json:"stages"`
	ConversionFunnel  []*FunnelStepDTO          `json:"conversion_funnel"`
	Trends            *PipelineTrendsDTO        `json:"trends,omitempty"`
}

// StageAnalyticsDTO represents analytics for a pipeline stage.
type StageAnalyticsDTO struct {
	StageID           string   `json:"stage_id"`
	StageName         string   `json:"stage_name"`
	StageOrder        int      `json:"stage_order"`
	OpportunityCount  int64    `json:"opportunity_count"`
	TotalValue        MoneyDTO `json:"total_value"`
	WeightedValue     MoneyDTO `json:"weighted_value"`
	AverageTimeInStage int     `json:"average_time_in_stage_days"`
	ConversionRate    float64  `json:"conversion_rate"`
}

// FunnelStepDTO represents a step in the conversion funnel.
type FunnelStepDTO struct {
	StageID         string  `json:"stage_id"`
	StageName       string  `json:"stage_name"`
	Count           int64   `json:"count"`
	ConversionRate  float64 `json:"conversion_rate"`
	DropoffRate     float64 `json:"dropoff_rate"`
}

// PipelineTrendsDTO represents pipeline trends over time.
type PipelineTrendsDTO struct {
	Period          string            `json:"period"` // daily, weekly, monthly
	NewOpportunities []TrendDataPoint `json:"new_opportunities"`
	WonDeals        []TrendDataPoint  `json:"won_deals"`
	LostDeals       []TrendDataPoint  `json:"lost_deals"`
	PipelineValue   []TrendDataPoint  `json:"pipeline_value"`
}

// TrendDataPoint represents a single data point in a trend.
type TrendDataPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}
