package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Lead Repository
// ============================================================================

// LeadRepository defines the interface for lead persistence operations.
type LeadRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, lead *Lead) error
	GetByID(ctx context.Context, tenantID, leadID uuid.UUID) (*Lead, error)
	Update(ctx context.Context, lead *Lead) error
	Delete(ctx context.Context, tenantID, leadID uuid.UUID) error

	// Query operations
	List(ctx context.Context, tenantID uuid.UUID, filter LeadFilter, opts ListOptions) ([]*Lead, int64, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*Lead, error)
	GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*Lead, error)

	// Status-based queries
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status LeadStatus, opts ListOptions) ([]*Lead, int64, error)
	GetQualifiedLeads(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Lead, int64, error)
	GetUnassignedLeads(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Lead, int64, error)

	// Assignment queries
	GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts ListOptions) ([]*Lead, int64, error)
	GetBySource(ctx context.Context, tenantID uuid.UUID, source LeadSource, opts ListOptions) ([]*Lead, int64, error)

	// Score-based queries
	GetHighScoreLeads(ctx context.Context, tenantID uuid.UUID, minScore int, opts ListOptions) ([]*Lead, int64, error)
	GetLeadsForNurturing(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Lead, int64, error)

	// Time-based queries
	GetCreatedBetween(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts ListOptions) ([]*Lead, int64, error)
	GetUpdatedSince(ctx context.Context, tenantID uuid.UUID, since time.Time, opts ListOptions) ([]*Lead, int64, error)
	GetStaleLeads(ctx context.Context, tenantID uuid.UUID, staleDays int, opts ListOptions) ([]*Lead, int64, error)

	// Bulk operations
	BulkCreate(ctx context.Context, leads []*Lead) error
	BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, leadIDs []uuid.UUID, newOwnerID uuid.UUID) error
	BulkUpdateStatus(ctx context.Context, tenantID uuid.UUID, leadIDs []uuid.UUID, status LeadStatus) error

	// Statistics
	CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[LeadStatus]int64, error)
	CountBySource(ctx context.Context, tenantID uuid.UUID) (map[LeadSource]int64, error)
	GetConversionRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error)
}

// LeadFilter defines filtering options for lead queries.
type LeadFilter struct {
	// Status filters
	Statuses []LeadStatus `json:"statuses,omitempty"`

	// Source filters
	Sources []LeadSource `json:"sources,omitempty"`

	// Assignment filters
	OwnerIDs   []uuid.UUID `json:"owner_ids,omitempty"`
	Unassigned *bool       `json:"unassigned,omitempty"`

	// Score filters
	MinScore *int `json:"min_score,omitempty"`
	MaxScore *int `json:"max_score,omitempty"`

	// Time filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`

	// Search
	SearchQuery string `json:"search_query,omitempty"`

	// Tag filter
	Tags []string `json:"tags,omitempty"`

	// Campaign filter
	CampaignID *uuid.UUID `json:"campaign_id,omitempty"`
}

// ============================================================================
// Opportunity Repository
// ============================================================================

// OpportunityRepository defines the interface for opportunity persistence operations.
type OpportunityRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, opportunity *Opportunity) error
	GetByID(ctx context.Context, tenantID, opportunityID uuid.UUID) (*Opportunity, error)
	Update(ctx context.Context, opportunity *Opportunity) error
	Delete(ctx context.Context, tenantID, opportunityID uuid.UUID) error

	// Query operations
	List(ctx context.Context, tenantID uuid.UUID, filter OpportunityFilter, opts ListOptions) ([]*Opportunity, int64, error)

	// Status-based queries
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status OpportunityStatus, opts ListOptions) ([]*Opportunity, int64, error)
	GetOpenOpportunities(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)
	GetWonOpportunities(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)
	GetLostOpportunities(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)

	// Pipeline and stage queries
	GetByPipeline(ctx context.Context, tenantID, pipelineID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)
	GetByStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)

	// Relationship queries
	GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)
	GetByContact(ctx context.Context, tenantID, contactID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)
	GetByLead(ctx context.Context, tenantID, leadID uuid.UUID) (*Opportunity, error)
	GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)

	// Date-based queries
	GetClosingThisMonth(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)
	GetClosingThisQuarter(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)
	GetOverdueOpportunities(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Opportunity, int64, error)
	GetByExpectedCloseDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts ListOptions) ([]*Opportunity, int64, error)

	// Value-based queries
	GetHighValueOpportunities(ctx context.Context, tenantID uuid.UUID, minAmount int64, currency string, opts ListOptions) ([]*Opportunity, int64, error)
	GetTotalPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error)
	GetWeightedPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error)

	// Bulk operations
	BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, newOwnerID uuid.UUID) error
	BulkUpdateStage(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, stageID uuid.UUID) error

	// Statistics
	CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[OpportunityStatus]int64, error)
	CountByStage(ctx context.Context, tenantID, pipelineID uuid.UUID) (map[uuid.UUID]int64, error)
	GetWinRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error)
	GetAverageDealSize(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error)
	GetAverageSalesCycle(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (int, error) // days
}

// OpportunityFilter defines filtering options for opportunity queries.
type OpportunityFilter struct {
	// Status filters
	Statuses []OpportunityStatus `json:"statuses,omitempty"`

	// Pipeline filters
	PipelineIDs []uuid.UUID `json:"pipeline_ids,omitempty"`
	StageIDs    []uuid.UUID `json:"stage_ids,omitempty"`

	// Relationship filters
	CustomerIDs []uuid.UUID `json:"customer_ids,omitempty"`
	ContactIDs  []uuid.UUID `json:"contact_ids,omitempty"`
	OwnerIDs    []uuid.UUID `json:"owner_ids,omitempty"`
	LeadID      *uuid.UUID  `json:"lead_id,omitempty"`

	// Value filters
	MinAmount *int64  `json:"min_amount,omitempty"`
	MaxAmount *int64  `json:"max_amount,omitempty"`
	Currency  *string `json:"currency,omitempty"`

	// Probability filter
	MinProbability *int `json:"min_probability,omitempty"`
	MaxProbability *int `json:"max_probability,omitempty"`

	// Date filters
	ExpectedCloseDateAfter  *time.Time `json:"expected_close_date_after,omitempty"`
	ExpectedCloseDateBefore *time.Time `json:"expected_close_date_before,omitempty"`
	CreatedAfter            *time.Time `json:"created_after,omitempty"`
	CreatedBefore           *time.Time `json:"created_before,omitempty"`

	// Search
	SearchQuery string `json:"search_query,omitempty"`

	// Product filter
	ProductIDs []uuid.UUID `json:"product_ids,omitempty"`

	// Source filter
	Sources []string `json:"sources,omitempty"`
}

// ============================================================================
// Deal Repository
// ============================================================================

// DealRepository defines the interface for deal persistence operations.
type DealRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, deal *Deal) error
	GetByID(ctx context.Context, tenantID, dealID uuid.UUID) (*Deal, error)
	Update(ctx context.Context, deal *Deal) error
	Delete(ctx context.Context, tenantID, dealID uuid.UUID) error

	// Query operations
	List(ctx context.Context, tenantID uuid.UUID, filter DealFilter, opts ListOptions) ([]*Deal, int64, error)
	GetByNumber(ctx context.Context, tenantID uuid.UUID, dealNumber string) (*Deal, error)

	// Status-based queries
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status DealStatus, opts ListOptions) ([]*Deal, int64, error)
	GetActiveDeals(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)
	GetCompletedDeals(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)

	// Relationship queries
	GetByOpportunity(ctx context.Context, tenantID, opportunityID uuid.UUID) (*Deal, error)
	GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)
	GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)

	// Payment queries
	GetDealsWithPendingPayments(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)
	GetOverduePayments(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)
	GetFullyPaidDeals(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)

	// Fulfillment queries
	GetDealsForFulfillment(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)
	GetPartiallyFulfilledDeals(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Deal, int64, error)

	// Date-based queries
	GetByClosedDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts ListOptions) ([]*Deal, int64, error)
	GetBySignedDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts ListOptions) ([]*Deal, int64, error)

	// Value-based queries
	GetTotalRevenue(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error)
	GetTotalReceivedPayments(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error)
	GetOutstandingAmount(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error)

	// Statistics
	CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[DealStatus]int64, error)
	GetAverageDealValue(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error)
	GetMonthlyRevenue(ctx context.Context, tenantID uuid.UUID, currency string, year int) (map[int]int64, error)
}

// DealFilter defines filtering options for deal queries.
type DealFilter struct {
	// Status filters
	Statuses []DealStatus `json:"statuses,omitempty"`

	// Relationship filters
	CustomerIDs   []uuid.UUID `json:"customer_ids,omitempty"`
	OpportunityID *uuid.UUID  `json:"opportunity_id,omitempty"`
	OwnerIDs      []uuid.UUID `json:"owner_ids,omitempty"`

	// Value filters
	MinAmount *int64  `json:"min_amount,omitempty"`
	MaxAmount *int64  `json:"max_amount,omitempty"`
	Currency  *string `json:"currency,omitempty"`

	// Payment status
	HasPendingPayments *bool `json:"has_pending_payments,omitempty"`
	FullyPaid          *bool `json:"fully_paid,omitempty"`

	// Fulfillment status
	FulfillmentProgress *int `json:"fulfillment_progress,omitempty"` // minimum percentage

	// Date filters
	ClosedDateAfter  *time.Time `json:"closed_date_after,omitempty"`
	ClosedDateBefore *time.Time `json:"closed_date_before,omitempty"`
	SignedDateAfter  *time.Time `json:"signed_date_after,omitempty"`
	SignedDateBefore *time.Time `json:"signed_date_before,omitempty"`

	// Search
	SearchQuery string `json:"search_query,omitempty"`

	// Deal number search
	DealNumber *string `json:"deal_number,omitempty"`
}

// ============================================================================
// Pipeline Repository
// ============================================================================

// PipelineRepository defines the interface for pipeline persistence operations.
type PipelineRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, pipeline *Pipeline) error
	GetByID(ctx context.Context, tenantID, pipelineID uuid.UUID) (*Pipeline, error)
	Update(ctx context.Context, pipeline *Pipeline) error
	Delete(ctx context.Context, tenantID, pipelineID uuid.UUID) error

	// Query operations
	List(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*Pipeline, int64, error)
	GetActivePipelines(ctx context.Context, tenantID uuid.UUID) ([]*Pipeline, error)
	GetDefaultPipeline(ctx context.Context, tenantID uuid.UUID) (*Pipeline, error)

	// Stage operations
	GetStageByID(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*Stage, error)
	AddStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *Stage) error
	UpdateStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *Stage) error
	RemoveStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) error
	ReorderStages(ctx context.Context, tenantID, pipelineID uuid.UUID, stageIDs []uuid.UUID) error

	// Statistics
	GetPipelineStatistics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*PipelineStatistics, error)
	GetStageStatistics(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*StageStatistics, error)
}

// PipelineStatistics contains aggregated statistics for a pipeline.
type PipelineStatistics struct {
	PipelineID         uuid.UUID             `json:"pipeline_id"`
	TotalOpportunities int64                 `json:"total_opportunities"`
	OpenOpportunities  int64                 `json:"open_opportunities"`
	WonOpportunities   int64                 `json:"won_opportunities"`
	LostOpportunities  int64                 `json:"lost_opportunities"`
	TotalValue         Money                 `json:"total_value"`
	WeightedValue      Money                 `json:"weighted_value"`
	WinRate            float64               `json:"win_rate"`
	AverageSalesCycle  int                   `json:"average_sales_cycle_days"`
	StageDistribution  map[uuid.UUID]int64   `json:"stage_distribution"`
	ConversionRates    map[uuid.UUID]float64 `json:"conversion_rates"` // stage -> next stage conversion
}

// StageStatistics contains aggregated statistics for a pipeline stage.
type StageStatistics struct {
	StageID               uuid.UUID `json:"stage_id"`
	TotalOpportunities    int64     `json:"total_opportunities"`
	TotalValue            Money     `json:"total_value"`
	AverageTimeInStage    int       `json:"average_time_in_stage_days"`
	ConversionRate        float64   `json:"conversion_rate"`
	AverageOpportunityAge int       `json:"average_opportunity_age_days"`
}

// ============================================================================
// Common Types
// ============================================================================

// ListOptions defines common options for list queries.
type ListOptions struct {
	// Pagination
	Page     int `json:"page"`
	PageSize int `json:"page_size"`

	// Sorting
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"` // "asc" or "desc"

	// Include soft-deleted records
	IncludeDeleted bool `json:"include_deleted"`
}

// DefaultListOptions returns default list options.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Page:           1,
		PageSize:       20,
		SortBy:         "created_at",
		SortOrder:      "desc",
		IncludeDeleted: false,
	}
}

// Offset calculates the offset for pagination.
func (o ListOptions) Offset() int {
	if o.Page < 1 {
		o.Page = 1
	}
	return (o.Page - 1) * o.PageSize
}

// Limit returns the page size with a maximum limit.
func (o ListOptions) Limit() int {
	if o.PageSize <= 0 {
		return 20
	}
	if o.PageSize > 100 {
		return 100
	}
	return o.PageSize
}

// ============================================================================
// Event Store Interface
// ============================================================================

// EventStore defines the interface for persisting and retrieving domain events.
type EventStore interface {
	// Save persists domain events.
	Save(ctx context.Context, events ...DomainEvent) error

	// GetByAggregateID retrieves all events for an aggregate.
	GetByAggregateID(ctx context.Context, tenantID, aggregateID uuid.UUID) ([]DomainEvent, error)

	// GetByAggregateType retrieves events by aggregate type.
	GetByAggregateType(ctx context.Context, tenantID uuid.UUID, aggregateType string, opts ListOptions) ([]DomainEvent, error)

	// GetByEventType retrieves events by event type.
	GetByEventType(ctx context.Context, tenantID uuid.UUID, eventType string, opts ListOptions) ([]DomainEvent, error)

	// GetEventsSince retrieves events that occurred after a specific time.
	GetEventsSince(ctx context.Context, tenantID uuid.UUID, since time.Time, opts ListOptions) ([]DomainEvent, error)

	// GetEventsInRange retrieves events within a time range.
	GetEventsInRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts ListOptions) ([]DomainEvent, error)
}

// ============================================================================
// Saga Repository Interface
// ============================================================================

// SagaRepository defines the interface for lead conversion saga persistence operations.
type SagaRepository interface {
	// Create creates a new saga instance.
	Create(ctx context.Context, saga *LeadConversionSaga) error

	// GetByID retrieves a saga by its ID.
	GetByID(ctx context.Context, tenantID, sagaID uuid.UUID) (*LeadConversionSaga, error)

	// GetByLeadID retrieves a saga by lead ID.
	GetByLeadID(ctx context.Context, tenantID, leadID uuid.UUID) (*LeadConversionSaga, error)

	// GetByIdempotencyKey retrieves a saga by its idempotency key.
	GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*LeadConversionSaga, error)

	// Update updates an existing saga.
	Update(ctx context.Context, saga *LeadConversionSaga) error

	// GetPendingSagas retrieves sagas that are stuck in non-terminal states.
	GetPendingSagas(ctx context.Context, olderThan time.Duration, limit int) ([]*LeadConversionSaga, error)

	// GetByState retrieves sagas by their current state.
	GetByState(ctx context.Context, tenantID uuid.UUID, state SagaState, opts ListOptions) ([]*LeadConversionSaga, int64, error)

	// GetCompensatingSagas retrieves sagas that are currently compensating.
	GetCompensatingSagas(ctx context.Context, limit int) ([]*LeadConversionSaga, error)

	// GetFailedSagas retrieves sagas that failed and need attention.
	GetFailedSagas(ctx context.Context, tenantID uuid.UUID, opts ListOptions) ([]*LeadConversionSaga, int64, error)

	// DeleteOldCompletedSagas deletes sagas that completed successfully and are older than specified time.
	DeleteOldCompletedSagas(ctx context.Context, olderThan time.Duration) (int64, error)

	// CountByState returns counts of sagas grouped by state.
	CountByState(ctx context.Context, tenantID uuid.UUID) (map[SagaState]int64, error)
}

// ============================================================================
// Idempotency Repository Interface
// ============================================================================

// IdempotencyRepository defines the interface for idempotency key storage and retrieval.
type IdempotencyRepository interface {
	// Store saves an idempotency key with its associated resource ID.
	Store(ctx context.Context, key *IdempotencyKey) error

	// Get retrieves the resource ID associated with an idempotency key.
	Get(ctx context.Context, tenantID uuid.UUID, key string) (*IdempotencyKey, error)

	// Exists checks if an idempotency key exists.
	Exists(ctx context.Context, tenantID uuid.UUID, key string) (bool, error)

	// Delete removes an idempotency key.
	Delete(ctx context.Context, tenantID uuid.UUID, key string) error

	// DeleteExpired removes all expired idempotency keys.
	DeleteExpired(ctx context.Context) (int64, error)

	// Extend extends the expiration time of an idempotency key.
	Extend(ctx context.Context, tenantID uuid.UUID, key string, newExpiry time.Time) error
}

// ============================================================================
// Unit of Work Interface
// ============================================================================

// UnitOfWork defines the interface for transactional operations.
type UnitOfWork interface {
	// Begin starts a new transaction.
	Begin(ctx context.Context) (UnitOfWork, error)

	// Commit commits the transaction.
	Commit() error

	// Rollback rolls back the transaction.
	Rollback() error

	// Repositories
	Leads() LeadRepository
	Opportunities() OpportunityRepository
	Deals() DealRepository
	Pipelines() PipelineRepository
	Events() EventStore
	Sagas() SagaRepository
	IdempotencyKeys() IdempotencyRepository
}
