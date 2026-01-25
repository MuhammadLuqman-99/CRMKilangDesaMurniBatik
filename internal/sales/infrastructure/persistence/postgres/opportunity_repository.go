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
// Opportunity Repository
// ============================================================================

// opportunityRow represents an opportunity database row.
type opportunityRow struct {
	ID                 uuid.UUID      `db:"id"`
	TenantID           uuid.UUID      `db:"tenant_id"`
	Name               string         `db:"name"`
	Description        sql.NullString `db:"description"`
	Status             string         `db:"status"`
	Priority           string         `db:"priority"`
	PipelineID         uuid.UUID      `db:"pipeline_id"`
	PipelineName       sql.NullString `db:"pipeline_name"`
	StageID            uuid.UUID      `db:"stage_id"`
	StageName          sql.NullString `db:"stage_name"`
	StageEnteredAt     time.Time      `db:"stage_entered_at"`
	Amount             int64          `db:"amount"`
	Currency           string         `db:"currency"`
	WeightedAmount     int64          `db:"weighted_amount"`
	Probability        int            `db:"probability"`
	ExpectedCloseDate  sql.NullTime   `db:"expected_close_date"`
	ActualCloseDate    sql.NullTime   `db:"actual_close_date"`
	CustomerID         uuid.UUID      `db:"customer_id"`
	CustomerName       sql.NullString `db:"customer_name"`
	LeadID             uuid.NullUUID  `db:"lead_id"`
	OwnerID            uuid.UUID      `db:"owner_id"`
	OwnerName          sql.NullString `db:"owner_name"`
	Source             sql.NullString `db:"source"`
	Campaign           sql.NullString `db:"campaign"`
	CampaignID         uuid.NullUUID  `db:"campaign_id"`
	Notes              sql.NullString `db:"notes"`
	Tags               StringArray    `db:"tags"`
	CustomFields       NullableJSON   `db:"custom_fields"`
	CloseReason        sql.NullString `db:"close_reason"`
	CloseNotes         sql.NullString `db:"close_notes"`
	ClosedAt           sql.NullTime   `db:"closed_at"`
	ClosedBy           uuid.NullUUID  `db:"closed_by"`
	CompetitorID       uuid.NullUUID  `db:"competitor_id"`
	CompetitorName     sql.NullString `db:"competitor_name"`
	DealID             uuid.NullUUID  `db:"deal_id"`
	ActivityCount      int            `db:"activity_count"`
	LastActivityAt     sql.NullTime   `db:"last_activity_at"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	CreatedBy          uuid.UUID      `db:"created_by"`
	UpdatedBy          uuid.UUID      `db:"updated_by"`
	DeletedAt          sql.NullTime   `db:"deleted_at"`
	Version            int            `db:"version"`
}

// OpportunityRepository implements domain.OpportunityRepository for PostgreSQL.
type OpportunityRepository struct {
	db *sqlx.DB
}

// NewOpportunityRepository creates a new OpportunityRepository.
func NewOpportunityRepository(db *sqlx.DB) *OpportunityRepository {
	return &OpportunityRepository{db: db}
}

// allowedOpportunitySortColumns maps API sort fields to database columns.
var allowedOpportunitySortColumns = map[string]string{
	"created_at":          "o.created_at",
	"updated_at":          "o.updated_at",
	"expected_close_date": "o.expected_close_date",
	"amount":              "o.amount",
	"probability":         "o.probability",
	"name":                "o.name",
	"status":              "o.status",
}

// Create inserts a new opportunity into the database.
func (r *OpportunityRepository) Create(ctx context.Context, opp *domain.Opportunity) error {
	exec := getExecutor(ctx, r.db)

	customFieldsJSON, err := ToJSON(opp.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		INSERT INTO sales.opportunities (
			id, tenant_id, name, description, status, priority,
			pipeline_id, pipeline_name, stage_id, stage_name, stage_entered_at,
			amount, currency, weighted_amount, probability,
			expected_close_date, customer_id, customer_name, lead_id,
			owner_id, owner_name, source, campaign, campaign_id,
			notes, tags, custom_fields, activity_count,
			created_at, updated_at, created_by, updated_by, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28,
			$29, $30, $31, $32, $33
		)`

	_, err = exec.ExecContext(ctx, query,
		opp.ID,
		opp.TenantID,
		opp.Name,
		nullString(opp.Description),
		string(opp.Status),
		string(opp.Priority),
		opp.PipelineID,
		nullString(opp.PipelineName),
		opp.StageID,
		nullString(opp.StageName),
		opp.StageEnteredAt,
		opp.Amount.Amount,
		opp.Amount.Currency,
		opp.WeightedAmount.Amount,
		opp.Probability,
		NewNullTime(opp.ExpectedCloseDate).NullTime,
		opp.CustomerID,
		nullString(opp.CustomerName),
		nullUUID(opp.LeadID),
		opp.OwnerID,
		nullString(opp.OwnerName),
		nullString(opp.Source),
		nullString(opp.Campaign),
		nullUUID(opp.CampaignID),
		nullString(opp.Notes),
		opp.Tags,
		customFieldsJSON,
		opp.ActivityCount,
		opp.CreatedAt,
		opp.UpdatedAt,
		opp.CreatedBy,
		opp.CreatedBy,
		opp.Version,
	)

	if err != nil {
		if IsUniqueViolation(err) {
			return fmt.Errorf("opportunity already exists: %w", err)
		}
		return fmt.Errorf("failed to create opportunity: %w", err)
	}

	// Insert products
	if err := r.insertProducts(ctx, opp.ID, opp.TenantID, opp.Products); err != nil {
		return err
	}

	// Insert contacts
	if err := r.insertContacts(ctx, opp.ID, opp.TenantID, opp.Contacts); err != nil {
		return err
	}

	// Insert stage history
	if err := r.insertStageHistory(ctx, opp.ID, opp.TenantID, opp.StageHistory); err != nil {
		return err
	}

	return nil
}

// GetByID retrieves an opportunity by ID.
func (r *OpportunityRepository) GetByID(ctx context.Context, tenantID, opportunityID uuid.UUID) (*domain.Opportunity, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT o.id, o.tenant_id, o.name, o.description, o.status, o.priority,
			o.pipeline_id, o.pipeline_name, o.stage_id, o.stage_name, o.stage_entered_at,
			o.amount, o.currency, o.weighted_amount, o.probability,
			o.expected_close_date, o.actual_close_date,
			o.customer_id, o.customer_name, o.lead_id, o.owner_id, o.owner_name,
			o.source, o.campaign, o.campaign_id, o.notes, o.tags, o.custom_fields,
			o.close_reason, o.close_notes, o.closed_at, o.closed_by,
			o.competitor_id, o.competitor_name, o.deal_id,
			o.activity_count, o.last_activity_at,
			o.created_at, o.updated_at, o.created_by, o.updated_by, o.deleted_at, o.version
		FROM sales.opportunities o
		WHERE o.tenant_id = $1 AND o.id = $2 AND o.deleted_at IS NULL`

	var row opportunityRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, opportunityID); err != nil {
		if IsNotFoundError(err) {
			return nil, fmt.Errorf("opportunity not found")
		}
		return nil, fmt.Errorf("failed to get opportunity: %w", err)
	}

	opp, err := r.toDomain(&row)
	if err != nil {
		return nil, err
	}

	// Load products
	products, err := r.getProducts(ctx, opportunityID, tenantID)
	if err != nil {
		return nil, err
	}
	opp.Products = products

	// Load contacts
	contacts, err := r.getContacts(ctx, opportunityID, tenantID)
	if err != nil {
		return nil, err
	}
	opp.Contacts = contacts

	// Load stage history
	history, err := r.getStageHistory(ctx, opportunityID, tenantID)
	if err != nil {
		return nil, err
	}
	opp.StageHistory = history

	return opp, nil
}

// Update updates an existing opportunity.
func (r *OpportunityRepository) Update(ctx context.Context, opp *domain.Opportunity) error {
	exec := getExecutor(ctx, r.db)

	customFieldsJSON, err := ToJSON(opp.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		UPDATE sales.opportunities SET
			name = $3, description = $4, status = $5, priority = $6,
			pipeline_id = $7, pipeline_name = $8, stage_id = $9, stage_name = $10,
			stage_entered_at = $11, amount = $12, currency = $13, weighted_amount = $14,
			probability = $15, expected_close_date = $16, actual_close_date = $17,
			customer_id = $18, customer_name = $19, lead_id = $20,
			owner_id = $21, owner_name = $22, source = $23, campaign = $24, campaign_id = $25,
			notes = $26, tags = $27, custom_fields = $28,
			close_reason = $29, close_notes = $30, closed_at = $31, closed_by = $32,
			competitor_id = $33, competitor_name = $34,
			activity_count = $35, last_activity_at = $36,
			updated_at = $37, updated_by = $38, version = version + 1
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL AND version = $39`

	var closeReason, closeNotes interface{}
	var closedAt, closedBy, competitorID, competitorName interface{}

	if opp.CloseInfo != nil {
		closeReason = opp.CloseInfo.Reason
		closeNotes = opp.CloseInfo.Notes
		closedAt = opp.CloseInfo.ClosedAt
		closedBy = opp.CloseInfo.ClosedBy
		if opp.CloseInfo.CompetitorID != nil {
			competitorID = *opp.CloseInfo.CompetitorID
		}
		competitorName = opp.CloseInfo.CompetitorName
	}

	result, err := exec.ExecContext(ctx, query,
		opp.TenantID,
		opp.ID,
		opp.Name,
		nullString(opp.Description),
		string(opp.Status),
		string(opp.Priority),
		opp.PipelineID,
		nullString(opp.PipelineName),
		opp.StageID,
		nullString(opp.StageName),
		opp.StageEnteredAt,
		opp.Amount.Amount,
		opp.Amount.Currency,
		opp.WeightedAmount.Amount,
		opp.Probability,
		NewNullTime(opp.ExpectedCloseDate).NullTime,
		NewNullTime(opp.ActualCloseDate).NullTime,
		opp.CustomerID,
		nullString(opp.CustomerName),
		nullUUID(opp.LeadID),
		opp.OwnerID,
		nullString(opp.OwnerName),
		nullString(opp.Source),
		nullString(opp.Campaign),
		nullUUID(opp.CampaignID),
		nullString(opp.Notes),
		opp.Tags,
		customFieldsJSON,
		closeReason,
		closeNotes,
		closedAt,
		closedBy,
		competitorID,
		competitorName,
		opp.ActivityCount,
		NewNullTime(opp.LastActivityAt).NullTime,
		time.Now().UTC(),
		opp.CreatedBy,
		opp.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update opportunity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("opportunity not found or version mismatch")
	}

	// Update products - delete and re-insert
	if err := r.deleteProducts(ctx, opp.ID, opp.TenantID); err != nil {
		return err
	}
	if err := r.insertProducts(ctx, opp.ID, opp.TenantID, opp.Products); err != nil {
		return err
	}

	// Update contacts - delete and re-insert
	if err := r.deleteContacts(ctx, opp.ID, opp.TenantID); err != nil {
		return err
	}
	if err := r.insertContacts(ctx, opp.ID, opp.TenantID, opp.Contacts); err != nil {
		return err
	}

	// Insert new stage history entries
	if err := r.insertStageHistory(ctx, opp.ID, opp.TenantID, opp.StageHistory); err != nil {
		return err
	}

	opp.Version++
	return nil
}

// Delete soft-deletes an opportunity.
func (r *OpportunityRepository) Delete(ctx context.Context, tenantID, opportunityID uuid.UUID) error {
	exec := getExecutor(ctx, r.db)

	query := `
		UPDATE sales.opportunities
		SET deleted_at = $3, updated_at = $3
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`

	result, err := exec.ExecContext(ctx, query, tenantID, opportunityID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete opportunity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("opportunity not found")
	}

	return nil
}

// List retrieves opportunities with filtering and pagination.
func (r *OpportunityRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.OpportunityFilter, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	exec := getExecutor(ctx, r.db)

	baseQuery := `
		SELECT o.id, o.tenant_id, o.name, o.description, o.status, o.priority,
			o.pipeline_id, o.pipeline_name, o.stage_id, o.stage_name, o.stage_entered_at,
			o.amount, o.currency, o.weighted_amount, o.probability,
			o.expected_close_date, o.actual_close_date,
			o.customer_id, o.customer_name, o.lead_id, o.owner_id, o.owner_name,
			o.source, o.campaign, o.campaign_id, o.notes, o.tags, o.custom_fields,
			o.close_reason, o.close_notes, o.closed_at, o.closed_by,
			o.competitor_id, o.competitor_name, o.deal_id,
			o.activity_count, o.last_activity_at,
			o.created_at, o.updated_at, o.created_by, o.updated_by, o.deleted_at, o.version
		FROM sales.opportunities o
		WHERE o.tenant_id = $1 AND o.deleted_at IS NULL`

	qb := NewQueryBuilder(baseQuery)
	qb.args = append(qb.args, tenantID)

	// Apply filters
	r.applyFilters(qb, filter)

	// Get total count
	countQuery, countArgs := qb.BuildCount()
	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, countArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to count opportunities: %w", err)
	}

	// Apply sorting and pagination
	sortColumn := ValidateSortColumn(opts.SortBy, allowedOpportunitySortColumns)
	sortOrder := ValidateSortOrder(opts.SortOrder)
	qb.OrderBy(sortColumn, sortOrder)
	qb.Limit(opts.Limit())
	qb.Offset(opts.Offset())

	query, args := qb.Build()

	var rows []opportunityRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list opportunities: %w", err)
	}

	opportunities := make([]*domain.Opportunity, 0, len(rows))
	for _, row := range rows {
		opp, err := r.toDomain(&row)
		if err != nil {
			return nil, 0, err
		}
		opportunities = append(opportunities, opp)
	}

	return opportunities, total, nil
}

// GetByStatus retrieves opportunities by status.
func (r *OpportunityRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.OpportunityStatus, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	filter := domain.OpportunityFilter{
		Statuses: []domain.OpportunityStatus{status},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetOpenOpportunities retrieves open opportunities.
func (r *OpportunityRepository) GetOpenOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return r.GetByStatus(ctx, tenantID, domain.OpportunityStatusOpen, opts)
}

// GetWonOpportunities retrieves won opportunities.
func (r *OpportunityRepository) GetWonOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return r.GetByStatus(ctx, tenantID, domain.OpportunityStatusWon, opts)
}

// GetLostOpportunities retrieves lost opportunities.
func (r *OpportunityRepository) GetLostOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return r.GetByStatus(ctx, tenantID, domain.OpportunityStatusLost, opts)
}

// GetByPipeline retrieves opportunities in a pipeline.
func (r *OpportunityRepository) GetByPipeline(ctx context.Context, tenantID, pipelineID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	filter := domain.OpportunityFilter{
		PipelineIDs: []uuid.UUID{pipelineID},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetByStage retrieves opportunities in a stage.
func (r *OpportunityRepository) GetByStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	filter := domain.OpportunityFilter{
		PipelineIDs: []uuid.UUID{pipelineID},
		StageIDs:    []uuid.UUID{stageID},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetByCustomer retrieves opportunities for a customer.
func (r *OpportunityRepository) GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	filter := domain.OpportunityFilter{
		CustomerIDs: []uuid.UUID{customerID},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetByContact retrieves opportunities involving a contact.
func (r *OpportunityRepository) GetByContact(ctx context.Context, tenantID, contactID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	filter := domain.OpportunityFilter{
		ContactIDs: []uuid.UUID{contactID},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetByLead retrieves the opportunity created from a lead.
func (r *OpportunityRepository) GetByLead(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.Opportunity, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT o.id, o.tenant_id, o.name, o.description, o.status, o.priority,
			o.pipeline_id, o.pipeline_name, o.stage_id, o.stage_name, o.stage_entered_at,
			o.amount, o.currency, o.weighted_amount, o.probability,
			o.expected_close_date, o.actual_close_date,
			o.customer_id, o.customer_name, o.lead_id, o.owner_id, o.owner_name,
			o.source, o.campaign, o.campaign_id, o.notes, o.tags, o.custom_fields,
			o.close_reason, o.close_notes, o.closed_at, o.closed_by,
			o.competitor_id, o.competitor_name, o.deal_id,
			o.activity_count, o.last_activity_at,
			o.created_at, o.updated_at, o.created_by, o.updated_by, o.deleted_at, o.version
		FROM sales.opportunities o
		WHERE o.tenant_id = $1 AND o.lead_id = $2 AND o.deleted_at IS NULL`

	var row opportunityRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, leadID); err != nil {
		if IsNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get opportunity by lead: %w", err)
	}

	return r.toDomain(&row)
}

// GetByOwner retrieves opportunities for an owner.
func (r *OpportunityRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	filter := domain.OpportunityFilter{
		OwnerIDs: []uuid.UUID{ownerID},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetClosingThisMonth retrieves opportunities expected to close this month.
func (r *OpportunityRepository) GetClosingThisMonth(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	filter := domain.OpportunityFilter{
		Statuses:                []domain.OpportunityStatus{domain.OpportunityStatusOpen},
		ExpectedCloseDateAfter:  &startOfMonth,
		ExpectedCloseDateBefore: &endOfMonth,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetClosingThisQuarter retrieves opportunities expected to close this quarter.
func (r *OpportunityRepository) GetClosingThisQuarter(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	now := time.Now()
	quarter := (int(now.Month()) - 1) / 3
	startOfQuarter := time.Date(now.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, time.UTC)
	endOfQuarter := startOfQuarter.AddDate(0, 3, 0)

	filter := domain.OpportunityFilter{
		Statuses:                []domain.OpportunityStatus{domain.OpportunityStatusOpen},
		ExpectedCloseDateAfter:  &startOfQuarter,
		ExpectedCloseDateBefore: &endOfQuarter,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetOverdueOpportunities retrieves opportunities past their expected close date.
func (r *OpportunityRepository) GetOverdueOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	now := time.Now()
	filter := domain.OpportunityFilter{
		Statuses:                []domain.OpportunityStatus{domain.OpportunityStatusOpen},
		ExpectedCloseDateBefore: &now,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetByExpectedCloseDate retrieves opportunities by expected close date range.
func (r *OpportunityRepository) GetByExpectedCloseDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	filter := domain.OpportunityFilter{
		ExpectedCloseDateAfter:  &start,
		ExpectedCloseDateBefore: &end,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetHighValueOpportunities retrieves opportunities above a certain value.
func (r *OpportunityRepository) GetHighValueOpportunities(ctx context.Context, tenantID uuid.UUID, minAmount int64, currency string, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	filter := domain.OpportunityFilter{
		MinAmount: &minAmount,
		Currency:  &currency,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetTotalPipelineValue returns the total value of open opportunities.
func (r *OpportunityRepository) GetTotalPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM sales.opportunities
		WHERE tenant_id = $1 AND currency = $2 AND status = 'open' AND deleted_at IS NULL`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, query, tenantID, currency); err != nil {
		return 0, fmt.Errorf("failed to get total pipeline value: %w", err)
	}

	return total, nil
}

// GetWeightedPipelineValue returns the weighted value of open opportunities.
func (r *OpportunityRepository) GetWeightedPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(SUM(weighted_amount), 0)
		FROM sales.opportunities
		WHERE tenant_id = $1 AND currency = $2 AND status = 'open' AND deleted_at IS NULL`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, query, tenantID, currency); err != nil {
		return 0, fmt.Errorf("failed to get weighted pipeline value: %w", err)
	}

	return total, nil
}

// BulkUpdateOwner updates owner for multiple opportunities.
func (r *OpportunityRepository) BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, newOwnerID uuid.UUID) error {
	if len(opportunityIDs) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	query := `
		UPDATE sales.opportunities
		SET owner_id = $2, updated_at = $3
		WHERE tenant_id = $1 AND id = ANY($4) AND deleted_at IS NULL`

	_, err := exec.ExecContext(ctx, query, tenantID, newOwnerID, time.Now().UTC(), opportunityIDs)
	if err != nil {
		return fmt.Errorf("failed to bulk update owner: %w", err)
	}

	return nil
}

// BulkUpdateStage updates stage for multiple opportunities.
func (r *OpportunityRepository) BulkUpdateStage(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, stageID uuid.UUID) error {
	if len(opportunityIDs) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	query := `
		UPDATE sales.opportunities
		SET stage_id = $2, stage_entered_at = $3, updated_at = $3
		WHERE tenant_id = $1 AND id = ANY($4) AND deleted_at IS NULL`

	now := time.Now().UTC()
	_, err := exec.ExecContext(ctx, query, tenantID, stageID, now, opportunityIDs)
	if err != nil {
		return fmt.Errorf("failed to bulk update stage: %w", err)
	}

	return nil
}

// CountByStatus counts opportunities by status.
func (r *OpportunityRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.OpportunityStatus]int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT status, COUNT(*) as count
		FROM sales.opportunities
		WHERE tenant_id = $1 AND deleted_at IS NULL
		GROUP BY status`

	rows, err := exec.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.OpportunityStatus]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[domain.OpportunityStatus(status)] = count
	}

	return result, nil
}

// CountByStage counts opportunities by stage in a pipeline.
func (r *OpportunityRepository) CountByStage(ctx context.Context, tenantID, pipelineID uuid.UUID) (map[uuid.UUID]int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT stage_id, COUNT(*) as count
		FROM sales.opportunities
		WHERE tenant_id = $1 AND pipeline_id = $2 AND status = 'open' AND deleted_at IS NULL
		GROUP BY stage_id`

	rows, err := exec.QueryContext(ctx, query, tenantID, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by stage: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]int64)
	for rows.Next() {
		var stageID uuid.UUID
		var count int64
		if err := rows.Scan(&stageID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[stageID] = count
	}

	return result, nil
}

// GetWinRate calculates the win rate.
func (r *OpportunityRepository) GetWinRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT
			COALESCE(
				CAST(SUM(CASE WHEN status = 'won' THEN 1 ELSE 0 END) AS FLOAT) /
				NULLIF(SUM(CASE WHEN status IN ('won', 'lost') THEN 1 ELSE 0 END), 0),
				0
			) as win_rate
		FROM sales.opportunities
		WHERE tenant_id = $1
			AND deleted_at IS NULL
			AND closed_at >= $2
			AND closed_at < $3`

	var rate float64
	if err := sqlx.GetContext(ctx, exec, &rate, query, tenantID, start, end); err != nil {
		return 0, fmt.Errorf("failed to get win rate: %w", err)
	}

	return rate, nil
}

// GetAverageDealSize calculates the average deal size.
func (r *OpportunityRepository) GetAverageDealSize(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(AVG(amount), 0)::bigint
		FROM sales.opportunities
		WHERE tenant_id = $1
			AND currency = $2
			AND status = 'won'
			AND deleted_at IS NULL
			AND closed_at >= $3
			AND closed_at < $4`

	var avg int64
	if err := sqlx.GetContext(ctx, exec, &avg, query, tenantID, currency, start, end); err != nil {
		return 0, fmt.Errorf("failed to get average deal size: %w", err)
	}

	return avg, nil
}

// GetAverageSalesCycle calculates the average sales cycle in days.
func (r *OpportunityRepository) GetAverageSalesCycle(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (int, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(AVG(EXTRACT(DAY FROM (closed_at - created_at))), 0)::int
		FROM sales.opportunities
		WHERE tenant_id = $1
			AND status = 'won'
			AND deleted_at IS NULL
			AND closed_at >= $2
			AND closed_at < $3`

	var avgDays int
	if err := sqlx.GetContext(ctx, exec, &avgDays, query, tenantID, start, end); err != nil {
		return 0, fmt.Errorf("failed to get average sales cycle: %w", err)
	}

	return avgDays, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func (r *OpportunityRepository) applyFilters(qb *QueryBuilder, filter domain.OpportunityFilter) {
	// Status filter
	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, s := range filter.Statuses {
			statuses[i] = string(s)
		}
		qb.WhereInStrings("o.status", statuses)
	}

	// Pipeline filter
	if len(filter.PipelineIDs) > 0 {
		qb.WhereIn("o.pipeline_id", filter.PipelineIDs)
	}

	// Stage filter
	if len(filter.StageIDs) > 0 {
		qb.WhereIn("o.stage_id", filter.StageIDs)
	}

	// Customer filter
	if len(filter.CustomerIDs) > 0 {
		qb.WhereIn("o.customer_id", filter.CustomerIDs)
	}

	// Owner filter
	if len(filter.OwnerIDs) > 0 {
		qb.WhereIn("o.owner_id", filter.OwnerIDs)
	}

	// Lead filter
	if filter.LeadID != nil {
		qb.Where(fmt.Sprintf("o.lead_id = $%d", qb.NextParam()), *filter.LeadID)
	}

	// Amount filters
	if filter.MinAmount != nil {
		qb.Where(fmt.Sprintf("o.amount >= $%d", qb.NextParam()), *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		qb.Where(fmt.Sprintf("o.amount <= $%d", qb.NextParam()), *filter.MaxAmount)
	}
	if filter.Currency != nil {
		qb.Where(fmt.Sprintf("o.currency = $%d", qb.NextParam()), *filter.Currency)
	}

	// Probability filters
	if filter.MinProbability != nil {
		qb.Where(fmt.Sprintf("o.probability >= $%d", qb.NextParam()), *filter.MinProbability)
	}
	if filter.MaxProbability != nil {
		qb.Where(fmt.Sprintf("o.probability <= $%d", qb.NextParam()), *filter.MaxProbability)
	}

	// Date filters
	if filter.ExpectedCloseDateAfter != nil {
		qb.Where(fmt.Sprintf("o.expected_close_date >= $%d", qb.NextParam()), *filter.ExpectedCloseDateAfter)
	}
	if filter.ExpectedCloseDateBefore != nil {
		qb.Where(fmt.Sprintf("o.expected_close_date < $%d", qb.NextParam()), *filter.ExpectedCloseDateBefore)
	}
	if filter.CreatedAfter != nil {
		qb.Where(fmt.Sprintf("o.created_at >= $%d", qb.NextParam()), *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		qb.Where(fmt.Sprintf("o.created_at < $%d", qb.NextParam()), *filter.CreatedBefore)
	}

	// Search filter
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		qb.Where(fmt.Sprintf(`(
			o.name ILIKE $%d OR
			o.customer_name ILIKE $%d
		)`, qb.NextParam(), qb.NextParam()), searchPattern, searchPattern)
	}
}

func (r *OpportunityRepository) toDomain(row *opportunityRow) (*domain.Opportunity, error) {
	opp := &domain.Opportunity{
		ID:           row.ID,
		TenantID:     row.TenantID,
		Name:         row.Name,
		Description:  row.Description.String,
		Status:       domain.OpportunityStatus(row.Status),
		Priority:     domain.OpportunityPriority(row.Priority),
		PipelineID:   row.PipelineID,
		PipelineName: row.PipelineName.String,
		StageID:      row.StageID,
		StageName:    row.StageName.String,
		StageEnteredAt: row.StageEnteredAt,
		Amount: domain.Money{
			Amount:   row.Amount,
			Currency: row.Currency,
		},
		WeightedAmount: domain.Money{
			Amount:   row.WeightedAmount,
			Currency: row.Currency,
		},
		Probability:   row.Probability,
		CustomerID:    row.CustomerID,
		CustomerName:  row.CustomerName.String,
		OwnerID:       row.OwnerID,
		OwnerName:     row.OwnerName.String,
		Source:        row.Source.String,
		Campaign:      row.Campaign.String,
		Notes:         row.Notes.String,
		Tags:          []string(row.Tags),
		ActivityCount: row.ActivityCount,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		CreatedBy:     row.CreatedBy,
		Version:       row.Version,
	}

	// Expected close date
	if row.ExpectedCloseDate.Valid {
		opp.ExpectedCloseDate = &row.ExpectedCloseDate.Time
	}

	// Actual close date
	if row.ActualCloseDate.Valid {
		opp.ActualCloseDate = &row.ActualCloseDate.Time
	}

	// Lead
	if row.LeadID.Valid {
		opp.LeadID = &row.LeadID.UUID
	}

	// Campaign
	if row.CampaignID.Valid {
		opp.CampaignID = &row.CampaignID.UUID
	}

	// Last activity
	if row.LastActivityAt.Valid {
		opp.LastActivityAt = &row.LastActivityAt.Time
	}

	// Close info
	if row.ClosedAt.Valid {
		opp.CloseInfo = &domain.CloseInfo{
			ClosedAt: row.ClosedAt.Time,
			ClosedBy: row.ClosedBy.UUID,
			Reason:   row.CloseReason.String,
			Notes:    row.CloseNotes.String,
		}
		if row.CompetitorID.Valid {
			opp.CloseInfo.CompetitorID = &row.CompetitorID.UUID
		}
		if row.CompetitorName.Valid {
			opp.CloseInfo.CompetitorName = row.CompetitorName.String
		}
	}

	// Custom fields
	if row.CustomFields.Valid {
		var customFields map[string]interface{}
		if err := json.Unmarshal([]byte(row.CustomFields.String), &customFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
		}
		opp.CustomFields = customFields
	}

	return opp, nil
}

// ============================================================================
// Product Operations
// ============================================================================

func (r *OpportunityRepository) insertProducts(ctx context.Context, opportunityID, tenantID uuid.UUID, products []domain.OpportunityProduct) error {
	if len(products) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	query := `
		INSERT INTO sales.opportunity_products (
			id, opportunity_id, tenant_id, product_id, product_name, quantity,
			unit_price_amount, unit_price_currency, discount, total_price_amount,
			total_price_currency, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	for _, p := range products {
		_, err := exec.ExecContext(ctx, query,
			p.ID, opportunityID, tenantID, p.ProductID, p.ProductName, p.Quantity,
			p.UnitPrice.Amount, p.UnitPrice.Currency, p.Discount,
			p.TotalPrice.Amount, p.TotalPrice.Currency,
			nullString(p.Notes), time.Now().UTC(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert product: %w", err)
		}
	}

	return nil
}

func (r *OpportunityRepository) getProducts(ctx context.Context, opportunityID, tenantID uuid.UUID) ([]domain.OpportunityProduct, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, product_id, product_name, quantity,
			unit_price_amount, unit_price_currency, discount,
			total_price_amount, total_price_currency, notes
		FROM sales.opportunity_products
		WHERE opportunity_id = $1 AND tenant_id = $2`

	rows, err := exec.QueryContext(ctx, query, opportunityID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []domain.OpportunityProduct
	for rows.Next() {
		var p domain.OpportunityProduct
		var unitPriceAmount, totalPriceAmount int64
		var unitPriceCurrency, totalPriceCurrency string
		var notes sql.NullString

		if err := rows.Scan(
			&p.ID, &p.ProductID, &p.ProductName, &p.Quantity,
			&unitPriceAmount, &unitPriceCurrency, &p.Discount,
			&totalPriceAmount, &totalPriceCurrency, &notes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		p.UnitPrice = domain.Money{Amount: unitPriceAmount, Currency: unitPriceCurrency}
		p.TotalPrice = domain.Money{Amount: totalPriceAmount, Currency: totalPriceCurrency}
		p.Notes = notes.String

		products = append(products, p)
	}

	return products, nil
}

func (r *OpportunityRepository) deleteProducts(ctx context.Context, opportunityID, tenantID uuid.UUID) error {
	exec := getExecutor(ctx, r.db)

	query := `DELETE FROM sales.opportunity_products WHERE opportunity_id = $1 AND tenant_id = $2`
	_, err := exec.ExecContext(ctx, query, opportunityID, tenantID)
	return err
}

// ============================================================================
// Contact Operations
// ============================================================================

func (r *OpportunityRepository) insertContacts(ctx context.Context, opportunityID, tenantID uuid.UUID, contacts []domain.OpportunityContact) error {
	if len(contacts) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	query := `
		INSERT INTO sales.opportunity_contacts (
			opportunity_id, tenant_id, contact_id, name, email, phone, role, is_primary, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, c := range contacts {
		_, err := exec.ExecContext(ctx, query,
			opportunityID, tenantID, c.ContactID, c.Name, c.Email,
			nullString(c.Phone), c.Role, c.IsPrimary, time.Now().UTC(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert contact: %w", err)
		}
	}

	return nil
}

func (r *OpportunityRepository) getContacts(ctx context.Context, opportunityID, tenantID uuid.UUID) ([]domain.OpportunityContact, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT contact_id, name, email, phone, role, is_primary
		FROM sales.opportunity_contacts
		WHERE opportunity_id = $1 AND tenant_id = $2`

	rows, err := exec.QueryContext(ctx, query, opportunityID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}
	defer rows.Close()

	var contacts []domain.OpportunityContact
	for rows.Next() {
		var c domain.OpportunityContact
		var phone sql.NullString

		if err := rows.Scan(&c.ContactID, &c.Name, &c.Email, &phone, &c.Role, &c.IsPrimary); err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}

		c.Phone = phone.String
		contacts = append(contacts, c)
	}

	return contacts, nil
}

func (r *OpportunityRepository) deleteContacts(ctx context.Context, opportunityID, tenantID uuid.UUID) error {
	exec := getExecutor(ctx, r.db)

	query := `DELETE FROM sales.opportunity_contacts WHERE opportunity_id = $1 AND tenant_id = $2`
	_, err := exec.ExecContext(ctx, query, opportunityID, tenantID)
	return err
}

// ============================================================================
// Stage History Operations
// ============================================================================

func (r *OpportunityRepository) insertStageHistory(ctx context.Context, opportunityID, tenantID uuid.UUID, history []domain.StageHistory) error {
	if len(history) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	query := `
		INSERT INTO sales.opportunity_stage_history (
			id, opportunity_id, tenant_id, stage_id, stage_name,
			entered_at, exited_at, duration_hours, moved_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO NOTHING`

	for _, h := range history {
		id := uuid.New()
		_, err := exec.ExecContext(ctx, query,
			id, opportunityID, tenantID, h.StageID, h.StageName,
			h.EnteredAt, NewNullTime(h.ExitedAt).NullTime, h.Duration,
			h.MovedBy, nullString(h.Notes),
		)
		if err != nil {
			return fmt.Errorf("failed to insert stage history: %w", err)
		}
	}

	return nil
}

func (r *OpportunityRepository) getStageHistory(ctx context.Context, opportunityID, tenantID uuid.UUID) ([]domain.StageHistory, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT stage_id, stage_name, entered_at, exited_at, duration_hours, moved_by, notes
		FROM sales.opportunity_stage_history
		WHERE opportunity_id = $1 AND tenant_id = $2
		ORDER BY entered_at ASC`

	rows, err := exec.QueryContext(ctx, query, opportunityID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage history: %w", err)
	}
	defer rows.Close()

	var history []domain.StageHistory
	for rows.Next() {
		var h domain.StageHistory
		var exitedAt sql.NullTime
		var notes sql.NullString

		if err := rows.Scan(&h.StageID, &h.StageName, &h.EnteredAt, &exitedAt, &h.Duration, &h.MovedBy, &notes); err != nil {
			return nil, fmt.Errorf("failed to scan stage history: %w", err)
		}

		if exitedAt.Valid {
			h.ExitedAt = &exitedAt.Time
		}
		h.Notes = notes.String
		history = append(history, h)
	}

	return history, nil
}
