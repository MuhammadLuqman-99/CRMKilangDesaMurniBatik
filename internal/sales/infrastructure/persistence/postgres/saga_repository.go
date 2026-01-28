// Package postgres contains PostgreSQL repository implementations for the sales service.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Saga Repository
// ============================================================================

// sagaRow represents a lead conversion saga database row.
type sagaRow struct {
	ID               uuid.UUID      `db:"id"`
	TenantID         uuid.UUID      `db:"tenant_id"`
	LeadID           uuid.UUID      `db:"lead_id"`
	IdempotencyKey   string         `db:"idempotency_key"`
	State            string         `db:"state"`
	CurrentStepIndex int            `db:"current_step_index"`
	Steps            []byte         `db:"steps"`
	Request          []byte         `db:"request"`
	Result           []byte         `db:"result"`
	OpportunityID    uuid.NullUUID  `db:"opportunity_id"`
	CustomerID       uuid.NullUUID  `db:"customer_id"`
	ContactID        uuid.NullUUID  `db:"contact_id"`
	CustomerCreated  bool           `db:"customer_created"`
	Error            sql.NullString `db:"error"`
	ErrorCode        sql.NullString `db:"error_code"`
	FailedStepType   sql.NullString `db:"failed_step_type"`
	InitiatedBy      uuid.UUID      `db:"initiated_by"`
	StartedAt        time.Time      `db:"started_at"`
	CompletedAt      sql.NullTime   `db:"completed_at"`
	Version          int            `db:"version"`
	Metadata         []byte         `db:"metadata"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

// SagaRepository implements domain.SagaRepository for PostgreSQL.
type SagaRepository struct {
	db *sqlx.DB
}

// NewSagaRepository creates a new SagaRepository.
func NewSagaRepository(db *sqlx.DB) *SagaRepository {
	return &SagaRepository{db: db}
}

// Create inserts a new saga into the database.
func (r *SagaRepository) Create(ctx context.Context, saga *domain.LeadConversionSaga) error {
	exec := getExecutor(ctx, r.db)

	stepsJSON, err := saga.StepsJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	requestJSON, err := saga.RequestJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	metadataJSON, err := saga.MetadataJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO sales.lead_conversion_sagas (
			id, tenant_id, lead_id, idempotency_key, state, current_step_index,
			steps, request, opportunity_id, customer_id, contact_id, customer_created,
			error, error_code, failed_step_type, initiated_by,
			started_at, completed_at, version, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)`

	_, err = exec.ExecContext(ctx, query,
		saga.ID,
		saga.TenantID,
		saga.LeadID,
		saga.IdempotencyKey,
		string(saga.State),
		saga.CurrentStepIndex,
		stepsJSON,
		requestJSON,
		nullUUID(saga.OpportunityID),
		nullUUID(saga.CustomerID),
		nullUUID(saga.ContactID),
		saga.CustomerCreated,
		nullStringPtr(saga.Error),
		nullStringPtr(saga.ErrorCode),
		r.nullStepType(saga.FailedStepType),
		saga.InitiatedBy,
		saga.StartedAt,
		NewNullTime(saga.CompletedAt).NullTime,
		saga.Version,
		metadataJSON,
		saga.CreatedAt,
		saga.UpdatedAt,
	)

	if err != nil {
		if IsUniqueViolation(err) {
			return domain.ErrIdempotencyKeyExists
		}
		return fmt.Errorf("failed to create saga: %w", err)
	}

	return nil
}

// GetByID retrieves a saga by its ID.
func (r *SagaRepository) GetByID(ctx context.Context, tenantID, sagaID uuid.UUID) (*domain.LeadConversionSaga, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, lead_id, idempotency_key, state, current_step_index,
			steps, request, result, opportunity_id, customer_id, contact_id, customer_created,
			error, error_code, failed_step_type, initiated_by,
			started_at, completed_at, version, metadata, created_at, updated_at
		FROM sales.lead_conversion_sagas
		WHERE tenant_id = $1 AND id = $2`

	var row sagaRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, sagaID); err != nil {
		if IsNotFoundError(err) {
			return nil, domain.ErrSagaNotFound
		}
		return nil, fmt.Errorf("failed to get saga: %w", err)
	}

	return r.toDomain(&row)
}

// GetByLeadID retrieves a saga by lead ID (returns the most recent one).
func (r *SagaRepository) GetByLeadID(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.LeadConversionSaga, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, lead_id, idempotency_key, state, current_step_index,
			steps, request, result, opportunity_id, customer_id, contact_id, customer_created,
			error, error_code, failed_step_type, initiated_by,
			started_at, completed_at, version, metadata, created_at, updated_at
		FROM sales.lead_conversion_sagas
		WHERE tenant_id = $1 AND lead_id = $2
		ORDER BY created_at DESC
		LIMIT 1`

	var row sagaRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, leadID); err != nil {
		if IsNotFoundError(err) {
			return nil, domain.ErrSagaNotFound
		}
		return nil, fmt.Errorf("failed to get saga by lead ID: %w", err)
	}

	return r.toDomain(&row)
}

// GetByIdempotencyKey retrieves a saga by its idempotency key.
func (r *SagaRepository) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*domain.LeadConversionSaga, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, lead_id, idempotency_key, state, current_step_index,
			steps, request, result, opportunity_id, customer_id, contact_id, customer_created,
			error, error_code, failed_step_type, initiated_by,
			started_at, completed_at, version, metadata, created_at, updated_at
		FROM sales.lead_conversion_sagas
		WHERE tenant_id = $1 AND idempotency_key = $2`

	var row sagaRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, key); err != nil {
		if IsNotFoundError(err) {
			return nil, domain.ErrSagaNotFound
		}
		return nil, fmt.Errorf("failed to get saga by idempotency key: %w", err)
	}

	return r.toDomain(&row)
}

// Update updates an existing saga with optimistic locking.
func (r *SagaRepository) Update(ctx context.Context, saga *domain.LeadConversionSaga) error {
	exec := getExecutor(ctx, r.db)

	stepsJSON, err := saga.StepsJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	requestJSON, err := saga.RequestJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resultJSON, err := saga.ResultJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	metadataJSON, err := saga.MetadataJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE sales.lead_conversion_sagas SET
			state = $3,
			current_step_index = $4,
			steps = $5,
			request = $6,
			result = $7,
			opportunity_id = $8,
			customer_id = $9,
			contact_id = $10,
			customer_created = $11,
			error = $12,
			error_code = $13,
			failed_step_type = $14,
			completed_at = $15,
			version = version + 1,
			metadata = $16,
			updated_at = $17
		WHERE tenant_id = $1 AND id = $2 AND version = $18`

	result, err := exec.ExecContext(ctx, query,
		saga.TenantID,
		saga.ID,
		string(saga.State),
		saga.CurrentStepIndex,
		stepsJSON,
		requestJSON,
		resultJSON,
		nullUUID(saga.OpportunityID),
		nullUUID(saga.CustomerID),
		nullUUID(saga.ContactID),
		saga.CustomerCreated,
		nullStringPtr(saga.Error),
		nullStringPtr(saga.ErrorCode),
		r.nullStepType(saga.FailedStepType),
		NewNullTime(saga.CompletedAt).NullTime,
		metadataJSON,
		time.Now().UTC(),
		saga.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update saga: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("saga not found or version mismatch")
	}

	saga.Version++
	return nil
}

// GetPendingSagas retrieves sagas that are stuck in non-terminal states.
func (r *SagaRepository) GetPendingSagas(ctx context.Context, olderThan time.Duration, limit int) ([]*domain.LeadConversionSaga, error) {
	exec := getExecutor(ctx, r.db)

	cutoffTime := time.Now().Add(-olderThan)

	query := `
		SELECT id, tenant_id, lead_id, idempotency_key, state, current_step_index,
			steps, request, result, opportunity_id, customer_id, contact_id, customer_created,
			error, error_code, failed_step_type, initiated_by,
			started_at, completed_at, version, metadata, created_at, updated_at
		FROM sales.lead_conversion_sagas
		WHERE state IN ('started', 'running', 'compensating')
			AND started_at < $1
		ORDER BY started_at ASC
		LIMIT $2`

	var rows []sagaRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, cutoffTime, limit); err != nil {
		return nil, fmt.Errorf("failed to get pending sagas: %w", err)
	}

	return r.toDomainSlice(rows)
}

// GetByState retrieves sagas by their current state.
func (r *SagaRepository) GetByState(ctx context.Context, tenantID uuid.UUID, state domain.SagaState, opts domain.ListOptions) ([]*domain.LeadConversionSaga, int64, error) {
	exec := getExecutor(ctx, r.db)

	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM sales.lead_conversion_sagas
		WHERE tenant_id = $1 AND state = $2`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, tenantID, string(state)); err != nil {
		return nil, 0, fmt.Errorf("failed to count sagas: %w", err)
	}

	// Data query
	query := `
		SELECT id, tenant_id, lead_id, idempotency_key, state, current_step_index,
			steps, request, result, opportunity_id, customer_id, contact_id, customer_created,
			error, error_code, failed_step_type, initiated_by,
			started_at, completed_at, version, metadata, created_at, updated_at
		FROM sales.lead_conversion_sagas
		WHERE tenant_id = $1 AND state = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	var rows []sagaRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, tenantID, string(state), opts.Limit(), opts.Offset()); err != nil {
		return nil, 0, fmt.Errorf("failed to get sagas by state: %w", err)
	}

	sagas, err := r.toDomainSlice(rows)
	if err != nil {
		return nil, 0, err
	}

	return sagas, total, nil
}

// GetCompensatingSagas retrieves sagas that are currently compensating.
func (r *SagaRepository) GetCompensatingSagas(ctx context.Context, limit int) ([]*domain.LeadConversionSaga, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, lead_id, idempotency_key, state, current_step_index,
			steps, request, result, opportunity_id, customer_id, contact_id, customer_created,
			error, error_code, failed_step_type, initiated_by,
			started_at, completed_at, version, metadata, created_at, updated_at
		FROM sales.lead_conversion_sagas
		WHERE state = 'compensating'
		ORDER BY started_at ASC
		LIMIT $1`

	var rows []sagaRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("failed to get compensating sagas: %w", err)
	}

	return r.toDomainSlice(rows)
}

// GetFailedSagas retrieves sagas that failed and need attention.
func (r *SagaRepository) GetFailedSagas(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.LeadConversionSaga, int64, error) {
	exec := getExecutor(ctx, r.db)

	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM sales.lead_conversion_sagas
		WHERE tenant_id = $1 AND state = 'failed'`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, tenantID); err != nil {
		return nil, 0, fmt.Errorf("failed to count failed sagas: %w", err)
	}

	// Data query
	query := `
		SELECT id, tenant_id, lead_id, idempotency_key, state, current_step_index,
			steps, request, result, opportunity_id, customer_id, contact_id, customer_created,
			error, error_code, failed_step_type, initiated_by,
			started_at, completed_at, version, metadata, created_at, updated_at
		FROM sales.lead_conversion_sagas
		WHERE tenant_id = $1 AND state = 'failed'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var rows []sagaRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, tenantID, opts.Limit(), opts.Offset()); err != nil {
		return nil, 0, fmt.Errorf("failed to get failed sagas: %w", err)
	}

	sagas, err := r.toDomainSlice(rows)
	if err != nil {
		return nil, 0, err
	}

	return sagas, total, nil
}

// DeleteOldCompletedSagas deletes sagas that completed successfully and are older than specified time.
func (r *SagaRepository) DeleteOldCompletedSagas(ctx context.Context, olderThan time.Duration) (int64, error) {
	exec := getExecutor(ctx, r.db)

	cutoffTime := time.Now().Add(-olderThan)

	query := `
		DELETE FROM sales.lead_conversion_sagas
		WHERE state IN ('completed', 'compensated')
			AND completed_at < $1`

	result, err := exec.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old completed sagas: %w", err)
	}

	return result.RowsAffected()
}

// CountByState returns counts of sagas grouped by state.
func (r *SagaRepository) CountByState(ctx context.Context, tenantID uuid.UUID) (map[domain.SagaState]int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT state, COUNT(*) as count
		FROM sales.lead_conversion_sagas
		WHERE tenant_id = $1
		GROUP BY state`

	rows, err := exec.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by state: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.SagaState]int64)
	for rows.Next() {
		var state string
		var count int64
		if err := rows.Scan(&state, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[domain.SagaState(state)] = count
	}

	return result, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func (r *SagaRepository) toDomain(row *sagaRow) (*domain.LeadConversionSaga, error) {
	saga := &domain.LeadConversionSaga{
		ID:               row.ID,
		TenantID:         row.TenantID,
		LeadID:           row.LeadID,
		IdempotencyKey:   row.IdempotencyKey,
		State:            domain.SagaState(row.State),
		CurrentStepIndex: row.CurrentStepIndex,
		CustomerCreated:  row.CustomerCreated,
		InitiatedBy:      row.InitiatedBy,
		StartedAt:        row.StartedAt,
		Version:          row.Version,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}

	// Steps
	if err := saga.SetStepsFromJSON(row.Steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
	}

	// Request
	if err := saga.SetRequestFromJSON(row.Request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Result
	if row.Result != nil {
		if err := saga.SetResultFromJSON(row.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	// Metadata
	if err := saga.SetMetadataFromJSON(row.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Optional fields
	if row.OpportunityID.Valid {
		saga.OpportunityID = &row.OpportunityID.UUID
	}
	if row.CustomerID.Valid {
		saga.CustomerID = &row.CustomerID.UUID
	}
	if row.ContactID.Valid {
		saga.ContactID = &row.ContactID.UUID
	}
	if row.Error.Valid {
		saga.Error = &row.Error.String
	}
	if row.ErrorCode.Valid {
		saga.ErrorCode = &row.ErrorCode.String
	}
	if row.FailedStepType.Valid {
		stepType := domain.SagaStepType(row.FailedStepType.String)
		saga.FailedStepType = &stepType
	}
	if row.CompletedAt.Valid {
		saga.CompletedAt = &row.CompletedAt.Time
	}

	return saga, nil
}

func (r *SagaRepository) toDomainSlice(rows []sagaRow) ([]*domain.LeadConversionSaga, error) {
	sagas := make([]*domain.LeadConversionSaga, 0, len(rows))
	for _, row := range rows {
		saga, err := r.toDomain(&row)
		if err != nil {
			return nil, err
		}
		sagas = append(sagas, saga)
	}
	return sagas, nil
}

func (r *SagaRepository) nullStepType(stepType *domain.SagaStepType) sql.NullString {
	if stepType == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(*stepType), Valid: true}
}

// ============================================================================
// Saga Step History Repository (for audit trail)
// ============================================================================

// sagaStepHistoryRow represents a saga step history database row.
type sagaStepHistoryRow struct {
	ID            uuid.UUID      `db:"id"`
	SagaID        uuid.UUID      `db:"saga_id"`
	TenantID      uuid.UUID      `db:"tenant_id"`
	StepID        uuid.UUID      `db:"step_id"`
	StepType      string         `db:"step_type"`
	StepOrder     int            `db:"step_order"`
	Status        string         `db:"status"`
	Input         []byte         `db:"input"`
	Output        []byte         `db:"output"`
	Error         sql.NullString `db:"error"`
	StartedAt     sql.NullTime   `db:"started_at"`
	CompletedAt   sql.NullTime   `db:"completed_at"`
	CompensatedAt sql.NullTime   `db:"compensated_at"`
	RetryCount    int            `db:"retry_count"`
	CreatedAt     time.Time      `db:"created_at"`
}

// SaveStepHistory saves a saga step to the history table.
func (r *SagaRepository) SaveStepHistory(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	exec := getExecutor(ctx, r.db)

	var inputJSON, outputJSON string
	var err error

	inputJSON, err = ToJSON(step.Input)
	if err != nil {
		inputJSON = "{}"
	}

	outputJSON, err = ToJSON(step.Output)
	if err != nil {
		outputJSON = "{}"
	}

	query := `
		INSERT INTO sales.saga_step_history (
			id, saga_id, tenant_id, step_id, step_type, step_order,
			status, input, output, error,
			started_at, completed_at, compensated_at, retry_count, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	_, err = exec.ExecContext(ctx, query,
		uuid.New(),
		saga.ID,
		saga.TenantID,
		step.ID,
		string(step.Type),
		step.Order,
		string(step.Status),
		inputJSON,
		outputJSON,
		nullStringPtr(step.Error),
		NewNullTime(step.StartedAt).NullTime,
		NewNullTime(step.CompletedAt).NullTime,
		NewNullTime(step.CompensatedAt).NullTime,
		step.RetryCount,
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to save step history: %w", err)
	}

	return nil
}

// GetStepHistory retrieves the step history for a saga.
func (r *SagaRepository) GetStepHistory(ctx context.Context, sagaID uuid.UUID) ([]*domain.SagaStep, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT step_id, step_type, step_order, status, input, output, error,
			started_at, completed_at, compensated_at, retry_count
		FROM sales.saga_step_history
		WHERE saga_id = $1
		ORDER BY step_order ASC, created_at ASC`

	var rows []sagaStepHistoryRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, sagaID); err != nil {
		return nil, fmt.Errorf("failed to get step history: %w", err)
	}

	steps := make([]*domain.SagaStep, 0, len(rows))
	for _, row := range rows {
		step := &domain.SagaStep{
			ID:         row.StepID,
			Type:       domain.SagaStepType(row.StepType),
			Order:      row.StepOrder,
			Status:     domain.SagaStepStatus(row.Status),
			RetryCount: row.RetryCount,
		}

		if row.Error.Valid {
			step.Error = &row.Error.String
		}
		if row.StartedAt.Valid {
			step.StartedAt = &row.StartedAt.Time
		}
		if row.CompletedAt.Valid {
			step.CompletedAt = &row.CompletedAt.Time
		}
		if row.CompensatedAt.Valid {
			step.CompensatedAt = &row.CompensatedAt.Time
		}

		// Unmarshal input/output
		if row.Input != nil {
			step.Input = make(map[string]interface{})
			_ = json.Unmarshal(row.Input, &step.Input)
		}
		if row.Output != nil {
			step.Output = make(map[string]interface{})
			_ = json.Unmarshal(row.Output, &step.Output)
		}

		steps = append(steps, step)
	}

	return steps, nil
}
