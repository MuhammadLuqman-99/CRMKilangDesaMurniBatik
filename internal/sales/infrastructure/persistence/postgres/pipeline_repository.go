// Package postgres provides PostgreSQL implementations for sales domain repositories.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Pipeline Repository Implementation
// ============================================================================

// PipelineRepository implements domain.PipelineRepository using PostgreSQL.
type PipelineRepository struct {
	db *sqlx.DB
}

// NewPipelineRepository creates a new PipelineRepository instance.
func NewPipelineRepository(db *sqlx.DB) *PipelineRepository {
	return &PipelineRepository{db: db}
}

// pipelineRow represents the database row structure for pipelines.
type pipelineRow struct {
	ID               uuid.UUID       `db:"id"`
	TenantID         uuid.UUID       `db:"tenant_id"`
	Name             string          `db:"name"`
	Description      sql.NullString  `db:"description"`
	IsDefault        bool            `db:"is_default"`
	IsActive         bool            `db:"is_active"`
	Currency         string          `db:"currency"`
	WinReasons       pq.StringArray  `db:"win_reasons"`
	LossReasons      pq.StringArray  `db:"loss_reasons"`
	RequiredFields   pq.StringArray  `db:"required_fields"`
	CustomFields     json.RawMessage `db:"custom_fields"`
	OpportunityCount int64           `db:"opportunity_count"`
	TotalValueAmount int64           `db:"total_value_amount"`
	TotalValueCurr   string          `db:"total_value_currency"`
	WonValueAmount   int64           `db:"won_value_amount"`
	WonValueCurr     string          `db:"won_value_currency"`
	CreatedBy        uuid.UUID       `db:"created_by"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
	DeletedAt        sql.NullTime    `db:"deleted_at"`
	Version          int             `db:"version"`
}

// stageRow represents the database row structure for stages.
type stageRow struct {
	ID          uuid.UUID       `db:"id"`
	TenantID    uuid.UUID       `db:"tenant_id"`
	PipelineID  uuid.UUID       `db:"pipeline_id"`
	Name        string          `db:"name"`
	Description sql.NullString  `db:"description"`
	Type        string          `db:"type"`
	Order       int             `db:"stage_order"`
	Probability int             `db:"probability"`
	Color       sql.NullString  `db:"color"`
	IsActive    bool            `db:"is_active"`
	RottenDays  int             `db:"rotten_days"`
	AutoActions json.RawMessage `db:"auto_actions"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

// Create creates a new pipeline in the database.
func (r *PipelineRepository) Create(ctx context.Context, pipeline *domain.Pipeline) error {
	executor := getExecutor(ctx, r.db)

	// Marshal custom fields
	customFieldsJSON, err := json.Marshal(pipeline.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	// Insert pipeline
	query := `
		INSERT INTO pipelines (
			id, tenant_id, name, description, is_default, is_active, currency,
			win_reasons, loss_reasons, required_fields, custom_fields,
			opportunity_count, total_value_amount, total_value_currency,
			won_value_amount, won_value_currency, created_by, created_at,
			updated_at, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20
		)`

	_, err = executor.ExecContext(ctx, query,
		pipeline.ID,
		pipeline.TenantID,
		pipeline.Name,
		nullString(pipeline.Description),
		pipeline.IsDefault,
		pipeline.IsActive,
		pipeline.Currency,
		pq.Array(pipeline.WinReasons),
		pq.Array(pipeline.LossReasons),
		pq.Array(pipeline.RequiredFields),
		customFieldsJSON,
		pipeline.OpportunityCount,
		pipeline.TotalValue.Amount,
		pipeline.TotalValue.Currency,
		pipeline.WonValue.Amount,
		pipeline.WonValue.Currency,
		pipeline.CreatedBy,
		pipeline.CreatedAt,
		pipeline.UpdatedAt,
		pipeline.Version,
	)
	if err != nil {
		if IsUniqueViolation(err) {
			return domain.ErrPipelineAlreadyExists
		}
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Insert stages
	if len(pipeline.Stages) > 0 {
		if err := r.insertStages(ctx, pipeline.TenantID, pipeline.ID, pipeline.Stages); err != nil {
			return fmt.Errorf("failed to insert stages: %w", err)
		}
	}

	return nil
}

// insertStages inserts multiple stages for a pipeline.
func (r *PipelineRepository) insertStages(ctx context.Context, tenantID, pipelineID uuid.UUID, stages []*domain.Stage) error {
	executor := getExecutor(ctx, r.db)

	query := `
		INSERT INTO pipeline_stages (
			id, tenant_id, pipeline_id, name, description, type, stage_order,
			probability, color, is_active, rotten_days, auto_actions,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	for _, stage := range stages {
		autoActionsJSON, err := json.Marshal(stage.AutoActions)
		if err != nil {
			return fmt.Errorf("failed to marshal auto actions: %w", err)
		}

		_, err = executor.ExecContext(ctx, query,
			stage.ID,
			tenantID,
			pipelineID,
			stage.Name,
			nullString(stage.Description),
			string(stage.Type),
			stage.Order,
			stage.Probability,
			nullString(stage.Color),
			stage.IsActive,
			stage.RottenDays,
			autoActionsJSON,
			stage.CreatedAt,
			stage.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert stage %s: %w", stage.Name, err)
		}
	}

	return nil
}

// GetByID retrieves a pipeline by ID.
func (r *PipelineRepository) GetByID(ctx context.Context, tenantID, pipelineID uuid.UUID) (*domain.Pipeline, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, name, description, is_default, is_active, currency,
			win_reasons, loss_reasons, required_fields, custom_fields,
			opportunity_count, total_value_amount, total_value_currency,
			won_value_amount, won_value_currency, created_by, created_at,
			updated_at, version
		FROM pipelines
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`

	var row pipelineRow
	err := sqlx.GetContext(ctx, executor, &row, query, pipelineID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPipelineNotFound
		}
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Get stages
	stages, err := r.getStages(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stages: %w", err)
	}

	return r.toDomainPipeline(&row, stages)
}

// getStages retrieves all stages for a pipeline.
func (r *PipelineRepository) getStages(ctx context.Context, tenantID, pipelineID uuid.UUID) ([]*domain.Stage, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, pipeline_id, name, description, type, stage_order,
			probability, color, is_active, rotten_days, auto_actions,
			created_at, updated_at
		FROM pipeline_stages
		WHERE pipeline_id = $1 AND tenant_id = $2
		ORDER BY stage_order ASC`

	var rows []stageRow
	err := sqlx.SelectContext(ctx, executor, &rows, query, pipelineID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stages: %w", err)
	}

	stages := make([]*domain.Stage, len(rows))
	for i, row := range rows {
		stage, err := r.toDomainStage(&row)
		if err != nil {
			return nil, fmt.Errorf("failed to convert stage: %w", err)
		}
		stages[i] = stage
	}

	return stages, nil
}

// Update updates an existing pipeline.
func (r *PipelineRepository) Update(ctx context.Context, pipeline *domain.Pipeline) error {
	executor := getExecutor(ctx, r.db)

	// Marshal custom fields
	customFieldsJSON, err := json.Marshal(pipeline.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		UPDATE pipelines SET
			name = $3,
			description = $4,
			is_default = $5,
			is_active = $6,
			currency = $7,
			win_reasons = $8,
			loss_reasons = $9,
			required_fields = $10,
			custom_fields = $11,
			opportunity_count = $12,
			total_value_amount = $13,
			total_value_currency = $14,
			won_value_amount = $15,
			won_value_currency = $16,
			updated_at = $17,
			version = version + 1
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL AND version = $18`

	result, err := executor.ExecContext(ctx, query,
		pipeline.ID,
		pipeline.TenantID,
		pipeline.Name,
		nullString(pipeline.Description),
		pipeline.IsDefault,
		pipeline.IsActive,
		pipeline.Currency,
		pq.Array(pipeline.WinReasons),
		pq.Array(pipeline.LossReasons),
		pq.Array(pipeline.RequiredFields),
		customFieldsJSON,
		pipeline.OpportunityCount,
		pipeline.TotalValue.Amount,
		pipeline.TotalValue.Currency,
		pipeline.WonValue.Amount,
		pipeline.WonValue.Currency,
		time.Now().UTC(),
		pipeline.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to update pipeline: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrPipelineNotFound
	}

	// Sync stages - delete and re-insert
	if err := r.syncStages(ctx, pipeline.TenantID, pipeline.ID, pipeline.Stages); err != nil {
		return fmt.Errorf("failed to sync stages: %w", err)
	}

	pipeline.Version++
	return nil
}

// syncStages syncs stages by deleting existing and inserting new ones.
func (r *PipelineRepository) syncStages(ctx context.Context, tenantID, pipelineID uuid.UUID, stages []*domain.Stage) error {
	executor := getExecutor(ctx, r.db)

	// Delete existing stages
	_, err := executor.ExecContext(ctx,
		`DELETE FROM pipeline_stages WHERE pipeline_id = $1 AND tenant_id = $2`,
		pipelineID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete stages: %w", err)
	}

	// Insert new stages
	if len(stages) > 0 {
		return r.insertStages(ctx, tenantID, pipelineID, stages)
	}

	return nil
}

// Delete soft-deletes a pipeline.
func (r *PipelineRepository) Delete(ctx context.Context, tenantID, pipelineID uuid.UUID) error {
	executor := getExecutor(ctx, r.db)

	query := `
		UPDATE pipelines SET
			deleted_at = $3,
			updated_at = $3
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`

	result, err := executor.ExecContext(ctx, query, pipelineID, tenantID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete pipeline: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrPipelineNotFound
	}

	return nil
}

// List retrieves pipelines with pagination.
func (r *PipelineRepository) List(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Pipeline, int64, error) {
	executor := getExecutor(ctx, r.db)

	// Count total
	countQuery := `SELECT COUNT(*) FROM pipelines WHERE tenant_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := sqlx.GetContext(ctx, executor, &total, countQuery, tenantID); err != nil {
		return nil, 0, fmt.Errorf("failed to count pipelines: %w", err)
	}

	// Build query
	qb := NewQueryBuilder(`
		SELECT id, tenant_id, name, description, is_default, is_active, currency,
			win_reasons, loss_reasons, required_fields, custom_fields,
			opportunity_count, total_value_amount, total_value_currency,
			won_value_amount, won_value_currency, created_by, created_at,
			updated_at, version
		FROM pipelines`)

	qb.Where("tenant_id = ?", tenantID)
	qb.Where("deleted_at IS NULL")

	// Apply sorting
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := opts.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	qb.OrderBy(sortBy, sortOrder)

	// Apply pagination
	qb.Limit(opts.Limit())
	qb.Offset(opts.Offset())

	query, args := qb.Build()

	var rows []pipelineRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list pipelines: %w", err)
	}

	// Convert to domain objects with stages
	pipelines := make([]*domain.Pipeline, len(rows))
	for i, row := range rows {
		stages, err := r.getStages(ctx, tenantID, row.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get stages for pipeline %s: %w", row.ID, err)
		}

		pipeline, err := r.toDomainPipeline(&row, stages)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert pipeline: %w", err)
		}
		pipelines[i] = pipeline
	}

	return pipelines, total, nil
}

// GetActivePipelines retrieves all active pipelines for a tenant.
func (r *PipelineRepository) GetActivePipelines(ctx context.Context, tenantID uuid.UUID) ([]*domain.Pipeline, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, name, description, is_default, is_active, currency,
			win_reasons, loss_reasons, required_fields, custom_fields,
			opportunity_count, total_value_amount, total_value_currency,
			won_value_amount, won_value_currency, created_by, created_at,
			updated_at, version
		FROM pipelines
		WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY is_default DESC, name ASC`

	var rows []pipelineRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, tenantID); err != nil {
		return nil, fmt.Errorf("failed to get active pipelines: %w", err)
	}

	pipelines := make([]*domain.Pipeline, len(rows))
	for i, row := range rows {
		stages, err := r.getStages(ctx, tenantID, row.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get stages: %w", err)
		}

		pipeline, err := r.toDomainPipeline(&row, stages)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pipeline: %w", err)
		}
		pipelines[i] = pipeline
	}

	return pipelines, nil
}

// GetDefaultPipeline retrieves the default pipeline for a tenant.
func (r *PipelineRepository) GetDefaultPipeline(ctx context.Context, tenantID uuid.UUID) (*domain.Pipeline, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, name, description, is_default, is_active, currency,
			win_reasons, loss_reasons, required_fields, custom_fields,
			opportunity_count, total_value_amount, total_value_currency,
			won_value_amount, won_value_currency, created_by, created_at,
			updated_at, version
		FROM pipelines
		WHERE tenant_id = $1 AND is_default = true AND deleted_at IS NULL
		LIMIT 1`

	var row pipelineRow
	err := sqlx.GetContext(ctx, executor, &row, query, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPipelineNotFound
		}
		return nil, fmt.Errorf("failed to get default pipeline: %w", err)
	}

	stages, err := r.getStages(ctx, tenantID, row.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stages: %w", err)
	}

	return r.toDomainPipeline(&row, stages)
}

// ============================================================================
// Stage Operations
// ============================================================================

// GetStageByID retrieves a specific stage by ID.
func (r *PipelineRepository) GetStageByID(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*domain.Stage, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, pipeline_id, name, description, type, stage_order,
			probability, color, is_active, rotten_days, auto_actions,
			created_at, updated_at
		FROM pipeline_stages
		WHERE id = $1 AND pipeline_id = $2 AND tenant_id = $3`

	var row stageRow
	err := sqlx.GetContext(ctx, executor, &row, query, stageID, pipelineID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrStageNotFound
		}
		return nil, fmt.Errorf("failed to get stage: %w", err)
	}

	return r.toDomainStage(&row)
}

// AddStage adds a new stage to a pipeline.
func (r *PipelineRepository) AddStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *domain.Stage) error {
	executor := getExecutor(ctx, r.db)

	autoActionsJSON, err := json.Marshal(stage.AutoActions)
	if err != nil {
		return fmt.Errorf("failed to marshal auto actions: %w", err)
	}

	query := `
		INSERT INTO pipeline_stages (
			id, tenant_id, pipeline_id, name, description, type, stage_order,
			probability, color, is_active, rotten_days, auto_actions,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	_, err = executor.ExecContext(ctx, query,
		stage.ID,
		tenantID,
		pipelineID,
		stage.Name,
		nullString(stage.Description),
		string(stage.Type),
		stage.Order,
		stage.Probability,
		nullString(stage.Color),
		stage.IsActive,
		stage.RottenDays,
		autoActionsJSON,
		stage.CreatedAt,
		stage.UpdatedAt,
	)
	if err != nil {
		if IsUniqueViolation(err) {
			return domain.ErrStageAlreadyExists
		}
		return fmt.Errorf("failed to add stage: %w", err)
	}

	// Update pipeline updated_at
	_, err = executor.ExecContext(ctx,
		`UPDATE pipelines SET updated_at = $1 WHERE id = $2 AND tenant_id = $3`,
		time.Now().UTC(), pipelineID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update pipeline timestamp: %w", err)
	}

	return nil
}

// UpdateStage updates an existing stage.
func (r *PipelineRepository) UpdateStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *domain.Stage) error {
	executor := getExecutor(ctx, r.db)

	autoActionsJSON, err := json.Marshal(stage.AutoActions)
	if err != nil {
		return fmt.Errorf("failed to marshal auto actions: %w", err)
	}

	query := `
		UPDATE pipeline_stages SET
			name = $4,
			description = $5,
			type = $6,
			stage_order = $7,
			probability = $8,
			color = $9,
			is_active = $10,
			rotten_days = $11,
			auto_actions = $12,
			updated_at = $13
		WHERE id = $1 AND pipeline_id = $2 AND tenant_id = $3`

	result, err := executor.ExecContext(ctx, query,
		stage.ID,
		pipelineID,
		tenantID,
		stage.Name,
		nullString(stage.Description),
		string(stage.Type),
		stage.Order,
		stage.Probability,
		nullString(stage.Color),
		stage.IsActive,
		stage.RottenDays,
		autoActionsJSON,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to update stage: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrStageNotFound
	}

	// Update pipeline updated_at
	_, err = executor.ExecContext(ctx,
		`UPDATE pipelines SET updated_at = $1 WHERE id = $2 AND tenant_id = $3`,
		time.Now().UTC(), pipelineID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update pipeline timestamp: %w", err)
	}

	return nil
}

// RemoveStage removes (deactivates) a stage from a pipeline.
func (r *PipelineRepository) RemoveStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) error {
	executor := getExecutor(ctx, r.db)

	// Check minimum stages
	var activeCount int
	err := sqlx.GetContext(ctx, executor, &activeCount,
		`SELECT COUNT(*) FROM pipeline_stages WHERE pipeline_id = $1 AND tenant_id = $2 AND is_active = true`,
		pipelineID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to count active stages: %w", err)
	}

	if activeCount <= 2 {
		return domain.ErrMinimumStagesRequired
	}

	// Deactivate stage
	query := `
		UPDATE pipeline_stages SET
			is_active = false,
			updated_at = $4
		WHERE id = $1 AND pipeline_id = $2 AND tenant_id = $3`

	result, err := executor.ExecContext(ctx, query, stageID, pipelineID, tenantID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to remove stage: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrStageNotFound
	}

	// Update pipeline updated_at
	_, err = executor.ExecContext(ctx,
		`UPDATE pipelines SET updated_at = $1 WHERE id = $2 AND tenant_id = $3`,
		time.Now().UTC(), pipelineID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update pipeline timestamp: %w", err)
	}

	return nil
}

// ReorderStages reorders stages in a pipeline.
func (r *PipelineRepository) ReorderStages(ctx context.Context, tenantID, pipelineID uuid.UUID, stageIDs []uuid.UUID) error {
	executor := getExecutor(ctx, r.db)

	// Verify all stages exist
	for _, stageID := range stageIDs {
		var exists bool
		err := sqlx.GetContext(ctx, executor, &exists,
			`SELECT EXISTS(SELECT 1 FROM pipeline_stages WHERE id = $1 AND pipeline_id = $2 AND tenant_id = $3)`,
			stageID, pipelineID, tenantID)
		if err != nil {
			return fmt.Errorf("failed to verify stage: %w", err)
		}
		if !exists {
			return domain.ErrStageNotFound
		}
	}

	// Update order for each stage
	now := time.Now().UTC()
	for order, stageID := range stageIDs {
		_, err := executor.ExecContext(ctx,
			`UPDATE pipeline_stages SET stage_order = $1, updated_at = $2 WHERE id = $3 AND pipeline_id = $4 AND tenant_id = $5`,
			order+1, now, stageID, pipelineID, tenantID)
		if err != nil {
			return fmt.Errorf("failed to update stage order: %w", err)
		}
	}

	// Update pipeline updated_at
	_, err := executor.ExecContext(ctx,
		`UPDATE pipelines SET updated_at = $1 WHERE id = $2 AND tenant_id = $3`,
		now, pipelineID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to update pipeline timestamp: %w", err)
	}

	return nil
}

// ============================================================================
// Statistics Operations
// ============================================================================

// GetPipelineStatistics retrieves aggregated statistics for a pipeline.
func (r *PipelineRepository) GetPipelineStatistics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*domain.PipelineStatistics, error) {
	executor := getExecutor(ctx, r.db)

	stats := &domain.PipelineStatistics{
		PipelineID:        pipelineID,
		StageDistribution: make(map[uuid.UUID]int64),
		ConversionRates:   make(map[uuid.UUID]float64),
	}

	// Get pipeline details first
	var currency string
	err := sqlx.GetContext(ctx, executor, &currency,
		`SELECT currency FROM pipelines WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`,
		pipelineID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPipelineNotFound
		}
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Get opportunity counts by status
	type statusCount struct {
		Status string `db:"status"`
		Count  int64  `db:"count"`
	}
	var statusCounts []statusCount
	err = sqlx.SelectContext(ctx, executor, &statusCounts, `
		SELECT status, COUNT(*) as count
		FROM opportunities
		WHERE pipeline_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		GROUP BY status`,
		pipelineID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	for _, sc := range statusCounts {
		stats.TotalOpportunities += sc.Count
		switch domain.OpportunityStatus(sc.Status) {
		case domain.OpportunityStatusOpen:
			stats.OpenOpportunities += sc.Count
		case domain.OpportunityStatusWon:
			stats.WonOpportunities += sc.Count
		case domain.OpportunityStatusLost:
			stats.LostOpportunities += sc.Count
		}
	}

	// Calculate win rate
	closedOpps := stats.WonOpportunities + stats.LostOpportunities
	if closedOpps > 0 {
		stats.WinRate = float64(stats.WonOpportunities) / float64(closedOpps) * 100
	}

	// Get total pipeline value
	var totalValue sql.NullInt64
	err = sqlx.GetContext(ctx, executor, &totalValue, `
		SELECT COALESCE(SUM(amount), 0)
		FROM opportunities
		WHERE pipeline_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
			AND status = $3 AND currency = $4`,
		pipelineID, tenantID, domain.OpportunityStatusOpen, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get total value: %w", err)
	}
	stats.TotalValue = domain.Money{Amount: totalValue.Int64, Currency: currency}

	// Get weighted pipeline value
	var weightedValue sql.NullInt64
	err = sqlx.GetContext(ctx, executor, &weightedValue, `
		SELECT COALESCE(SUM(amount * probability / 100), 0)
		FROM opportunities
		WHERE pipeline_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
			AND status = $3 AND currency = $4`,
		pipelineID, tenantID, domain.OpportunityStatusOpen, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get weighted value: %w", err)
	}
	stats.WeightedValue = domain.Money{Amount: weightedValue.Int64, Currency: currency}

	// Get average sales cycle (days from created to won)
	var avgCycle sql.NullFloat64
	err = sqlx.GetContext(ctx, executor, &avgCycle, `
		SELECT AVG(EXTRACT(EPOCH FROM (actual_close_date - created_at)) / 86400)
		FROM opportunities
		WHERE pipeline_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
			AND status = $3 AND actual_close_date IS NOT NULL`,
		pipelineID, tenantID, domain.OpportunityStatusWon)
	if err != nil {
		return nil, fmt.Errorf("failed to get average sales cycle: %w", err)
	}
	if avgCycle.Valid {
		stats.AverageSalesCycle = int(avgCycle.Float64)
	}

	// Get stage distribution
	type stageDistribution struct {
		StageID uuid.UUID `db:"stage_id"`
		Count   int64     `db:"count"`
	}
	var stageDist []stageDistribution
	err = sqlx.SelectContext(ctx, executor, &stageDist, `
		SELECT stage_id, COUNT(*) as count
		FROM opportunities
		WHERE pipeline_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
			AND status = $3
		GROUP BY stage_id`,
		pipelineID, tenantID, domain.OpportunityStatusOpen)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage distribution: %w", err)
	}
	for _, sd := range stageDist {
		stats.StageDistribution[sd.StageID] = sd.Count
	}

	// Calculate conversion rates between stages
	stages, err := r.getStages(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stages: %w", err)
	}

	for i := 0; i < len(stages)-1; i++ {
		currentStage := stages[i]
		nextStage := stages[i+1]

		// Skip closed stages
		if currentStage.Type.IsClosedType() {
			continue
		}

		// Count opportunities that moved from current to next stage
		var moved, stayed int64
		err = sqlx.GetContext(ctx, executor, &moved, `
			SELECT COUNT(DISTINCT o.id)
			FROM opportunities o
			JOIN opportunity_stage_history osh ON o.id = osh.opportunity_id AND o.tenant_id = osh.tenant_id
			WHERE o.pipeline_id = $1 AND o.tenant_id = $2
				AND osh.from_stage_id = $3 AND osh.to_stage_id = $4`,
			pipelineID, tenantID, currentStage.ID, nextStage.ID)
		if err != nil {
			continue // Skip on error
		}

		err = sqlx.GetContext(ctx, executor, &stayed, `
			SELECT COUNT(DISTINCT o.id)
			FROM opportunities o
			JOIN opportunity_stage_history osh ON o.id = osh.opportunity_id AND o.tenant_id = osh.tenant_id
			WHERE o.pipeline_id = $1 AND o.tenant_id = $2
				AND osh.from_stage_id = $3`,
			pipelineID, tenantID, currentStage.ID)
		if err != nil {
			continue
		}

		if stayed > 0 {
			stats.ConversionRates[currentStage.ID] = float64(moved) / float64(stayed) * 100
		}
	}

	return stats, nil
}

// GetStageStatistics retrieves statistics for a specific stage.
func (r *PipelineRepository) GetStageStatistics(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*domain.StageStatistics, error) {
	executor := getExecutor(ctx, r.db)

	stats := &domain.StageStatistics{
		StageID: stageID,
	}

	// Get pipeline currency
	var currency string
	err := sqlx.GetContext(ctx, executor, &currency,
		`SELECT currency FROM pipelines WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`,
		pipelineID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPipelineNotFound
		}
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	// Verify stage exists
	var stageExists bool
	err = sqlx.GetContext(ctx, executor, &stageExists,
		`SELECT EXISTS(SELECT 1 FROM pipeline_stages WHERE id = $1 AND pipeline_id = $2 AND tenant_id = $3)`,
		stageID, pipelineID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify stage: %w", err)
	}
	if !stageExists {
		return nil, domain.ErrStageNotFound
	}

	// Get opportunities in this stage
	type stageStats struct {
		Count      int64           `db:"count"`
		TotalValue sql.NullInt64   `db:"total_value"`
		AvgAge     sql.NullFloat64 `db:"avg_age"`
	}
	var ss stageStats
	err = sqlx.GetContext(ctx, executor, &ss, `
		SELECT
			COUNT(*) as count,
			COALESCE(SUM(amount), 0) as total_value,
			AVG(EXTRACT(EPOCH FROM (NOW() - created_at)) / 86400) as avg_age
		FROM opportunities
		WHERE stage_id = $1 AND pipeline_id = $2 AND tenant_id = $3
			AND deleted_at IS NULL AND status = $4 AND currency = $5`,
		stageID, pipelineID, tenantID, domain.OpportunityStatusOpen, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage stats: %w", err)
	}

	stats.TotalOpportunities = ss.Count
	stats.TotalValue = domain.Money{Amount: ss.TotalValue.Int64, Currency: currency}
	if ss.AvgAge.Valid {
		stats.AverageOpportunityAge = int(ss.AvgAge.Float64)
	}

	// Get average time in stage (from stage history)
	var avgTimeInStage sql.NullFloat64
	err = sqlx.GetContext(ctx, executor, &avgTimeInStage, `
		SELECT AVG(EXTRACT(EPOCH FROM (exited_at - entered_at)) / 86400)
		FROM opportunity_stage_history
		WHERE to_stage_id = $1 AND tenant_id = $2 AND exited_at IS NOT NULL`,
		stageID, tenantID)
	if err == nil && avgTimeInStage.Valid {
		stats.AverageTimeInStage = int(avgTimeInStage.Float64)
	}

	// Get conversion rate (opportunities that moved to next stage / total that entered)
	var entered, moved int64
	err = sqlx.GetContext(ctx, executor, &entered, `
		SELECT COUNT(DISTINCT opportunity_id)
		FROM opportunity_stage_history
		WHERE to_stage_id = $1 AND tenant_id = $2`,
		stageID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entered count: %w", err)
	}

	err = sqlx.GetContext(ctx, executor, &moved, `
		SELECT COUNT(DISTINCT opportunity_id)
		FROM opportunity_stage_history
		WHERE from_stage_id = $1 AND tenant_id = $2`,
		stageID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get moved count: %w", err)
	}

	if entered > 0 {
		stats.ConversionRate = float64(moved) / float64(entered) * 100
	}

	return stats, nil
}

// ============================================================================
// Conversion Helpers
// ============================================================================

// toDomainPipeline converts a database row to a domain Pipeline.
func (r *PipelineRepository) toDomainPipeline(row *pipelineRow, stages []*domain.Stage) (*domain.Pipeline, error) {
	// Parse custom fields
	var customFields []domain.CustomFieldDef
	if len(row.CustomFields) > 0 && string(row.CustomFields) != "null" {
		if err := json.Unmarshal(row.CustomFields, &customFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
		}
	}

	return &domain.Pipeline{
		ID:               row.ID,
		TenantID:         row.TenantID,
		Name:             row.Name,
		Description:      nullStringValue(row.Description),
		IsDefault:        row.IsDefault,
		IsActive:         row.IsActive,
		Currency:         row.Currency,
		Stages:           stages,
		WinReasons:       row.WinReasons,
		LossReasons:      row.LossReasons,
		RequiredFields:   row.RequiredFields,
		CustomFields:     customFields,
		OpportunityCount: row.OpportunityCount,
		TotalValue: domain.Money{
			Amount:   row.TotalValueAmount,
			Currency: row.TotalValueCurr,
		},
		WonValue: domain.Money{
			Amount:   row.WonValueAmount,
			Currency: row.WonValueCurr,
		},
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		Version:   row.Version,
	}, nil
}

// toDomainStage converts a database row to a domain Stage.
func (r *PipelineRepository) toDomainStage(row *stageRow) (*domain.Stage, error) {
	// Parse auto actions
	var autoActions []domain.AutoAction
	if len(row.AutoActions) > 0 && string(row.AutoActions) != "null" {
		if err := json.Unmarshal(row.AutoActions, &autoActions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal auto actions: %w", err)
		}
	}

	return &domain.Stage{
		ID:          row.ID,
		PipelineID:  row.PipelineID,
		Name:        row.Name,
		Description: nullStringValue(row.Description),
		Type:        domain.StageType(row.Type),
		Order:       row.Order,
		Probability: row.Probability,
		Color:       nullStringValue(row.Color),
		IsActive:    row.IsActive,
		RottenDays:  row.RottenDays,
		AutoActions: autoActions,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}
