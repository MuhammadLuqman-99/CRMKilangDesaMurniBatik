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
// Lead Repository
// ============================================================================

// leadRow represents a lead database row.
type leadRow struct {
	ID               uuid.UUID      `db:"id"`
	TenantID         uuid.UUID      `db:"tenant_id"`
	FirstName        string         `db:"first_name"`
	LastName         string         `db:"last_name"`
	Email            string         `db:"email"`
	Phone            sql.NullString `db:"phone"`
	Mobile           sql.NullString `db:"mobile"`
	JobTitle         sql.NullString `db:"job_title"`
	Department       sql.NullString `db:"department"`
	CompanyName      sql.NullString `db:"company_name"`
	CompanySize      sql.NullString `db:"company_size"`
	Industry         sql.NullString `db:"industry"`
	Website          sql.NullString `db:"website"`
	Address          sql.NullString `db:"address"`
	City             sql.NullString `db:"city"`
	State            sql.NullString `db:"state"`
	PostalCode       sql.NullString `db:"postal_code"`
	Country          sql.NullString `db:"country"`
	Status           string         `db:"status"`
	Source           string         `db:"source"`
	Rating           string         `db:"rating"`
	Score            int            `db:"score"`
	DemographicScore int            `db:"demographic_score"`
	BehavioralScore  int            `db:"behavioral_score"`
	OwnerID          uuid.NullUUID  `db:"owner_id"`
	CampaignID       uuid.NullUUID  `db:"campaign_id"`
	Description      sql.NullString `db:"description"`
	EstimatedAmount  int64          `db:"estimated_amount"`
	EstimatedCurrency string        `db:"estimated_currency"`
	Tags             StringArray    `db:"tags"`
	CustomFields     NullableJSON   `db:"custom_fields"`
	LastContactedAt  sql.NullTime   `db:"last_contacted_at"`
	ConvertedAt      sql.NullTime   `db:"converted_at"`
	ConvertedBy      uuid.NullUUID  `db:"converted_by"`
	OpportunityID    uuid.NullUUID  `db:"opportunity_id"`
	CustomerID       uuid.NullUUID  `db:"customer_id"`
	ContactID        uuid.NullUUID  `db:"contact_id"`
	DisqualifiedAt   sql.NullTime   `db:"disqualified_at"`
	DisqualifiedBy   uuid.NullUUID  `db:"disqualified_by"`
	DisqualifyReason sql.NullString `db:"disqualify_reason"`
	EmailsOpened     int            `db:"emails_opened"`
	EmailsClicked    int            `db:"emails_clicked"`
	WebVisits        int            `db:"web_visits"`
	FormSubmissions  int            `db:"form_submissions"`
	LastEngagement   sql.NullTime   `db:"last_engagement"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	CreatedBy        uuid.UUID      `db:"created_by"`
	UpdatedBy        uuid.UUID      `db:"updated_by"`
	DeletedAt        sql.NullTime   `db:"deleted_at"`
	Version          int            `db:"version"`
}

// LeadRepository implements domain.LeadRepository for PostgreSQL.
type LeadRepository struct {
	db *sqlx.DB
}

// NewLeadRepository creates a new LeadRepository.
func NewLeadRepository(db *sqlx.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

// allowedLeadSortColumns maps API sort fields to database columns.
var allowedLeadSortColumns = map[string]string{
	"created_at":  "created_at",
	"updated_at":  "updated_at",
	"score":       "score",
	"first_name":  "first_name",
	"last_name":   "last_name",
	"company":     "company_name",
	"status":      "status",
	"source":      "source",
}

// Create inserts a new lead into the database.
func (r *LeadRepository) Create(ctx context.Context, lead *domain.Lead) error {
	exec := getExecutor(ctx, r.db)

	customFieldsJSON, err := ToJSON(lead.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		INSERT INTO sales.leads (
			id, tenant_id, first_name, last_name, email, phone, mobile,
			job_title, department, company_name, company_size, industry, website,
			address, city, state, postal_code, country,
			status, source, rating, score, demographic_score, behavioral_score,
			owner_id, campaign_id, description, estimated_amount, estimated_currency,
			tags, custom_fields, last_contacted_at,
			emails_opened, emails_clicked, web_visits, form_submissions, last_engagement,
			created_at, updated_at, created_by, updated_by, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24,
			$25, $26, $27, $28, $29, $30, $31, $32,
			$33, $34, $35, $36, $37, $38, $39, $40, $41, $42
		)`

	_, err = exec.ExecContext(ctx, query,
		lead.ID,
		lead.TenantID,
		lead.Contact.FirstName,
		lead.Contact.LastName,
		lead.Contact.Email,
		nullString(lead.Contact.Phone),
		nullString(lead.Contact.Mobile),
		nullString(lead.Contact.JobTitle),
		nullString(lead.Contact.Department),
		nullString(lead.Company.Name),
		nullString(lead.Company.Size),
		nullString(lead.Company.Industry),
		nullString(lead.Company.Website),
		nullString(lead.Company.Address),
		nullString(lead.Company.City),
		nullString(lead.Company.State),
		nullString(lead.Company.PostalCode),
		nullString(lead.Company.Country),
		string(lead.Status),
		string(lead.Source),
		string(lead.Rating),
		lead.Score.Score,
		lead.Score.Demographic,
		lead.Score.Behavioral,
		nullUUID(lead.OwnerID),
		nullUUID(lead.CampaignID),
		nullString(lead.Description),
		lead.EstimatedValue.Amount,
		lead.EstimatedValue.Currency,
		lead.Tags,
		customFieldsJSON,
		NewNullTime(lead.LastContactedAt).NullTime,
		lead.Engagement.EmailsOpened,
		lead.Engagement.EmailsClicked,
		lead.Engagement.WebVisits,
		lead.Engagement.FormSubmissions,
		NewNullTime(lead.Engagement.LastEngagement).NullTime,
		lead.CreatedAt,
		lead.UpdatedAt,
		lead.CreatedBy,
		lead.CreatedBy,
		lead.Version,
	)

	if err != nil {
		if IsUniqueViolation(err) {
			return fmt.Errorf("lead with email already exists: %w", err)
		}
		return fmt.Errorf("failed to create lead: %w", err)
	}

	return nil
}

// GetByID retrieves a lead by ID.
func (r *LeadRepository) GetByID(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.Lead, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, first_name, last_name, email, phone, mobile,
			job_title, department, company_name, company_size, industry, website,
			address, city, state, postal_code, country,
			status, source, rating, score, demographic_score, behavioral_score,
			owner_id, campaign_id, description, estimated_amount, estimated_currency,
			tags, custom_fields, last_contacted_at,
			converted_at, converted_by, opportunity_id, customer_id, contact_id,
			disqualified_at, disqualified_by, disqualify_reason,
			emails_opened, emails_clicked, web_visits, form_submissions, last_engagement,
			created_at, updated_at, created_by, updated_by, deleted_at, version
		FROM sales.leads
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`

	var row leadRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, leadID); err != nil {
		if IsNotFoundError(err) {
			return nil, fmt.Errorf("lead not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get lead: %w", err)
	}

	return r.toDomain(&row)
}

// Update updates an existing lead.
func (r *LeadRepository) Update(ctx context.Context, lead *domain.Lead) error {
	exec := getExecutor(ctx, r.db)

	customFieldsJSON, err := ToJSON(lead.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		UPDATE sales.leads SET
			first_name = $3, last_name = $4, email = $5, phone = $6, mobile = $7,
			job_title = $8, department = $9, company_name = $10, company_size = $11,
			industry = $12, website = $13, address = $14, city = $15, state = $16,
			postal_code = $17, country = $18, status = $19, source = $20, rating = $21,
			score = $22, demographic_score = $23, behavioral_score = $24,
			owner_id = $25, campaign_id = $26, description = $27,
			estimated_amount = $28, estimated_currency = $29,
			tags = $30, custom_fields = $31, last_contacted_at = $32,
			converted_at = $33, converted_by = $34, opportunity_id = $35,
			customer_id = $36, contact_id = $37,
			disqualified_at = $38, disqualified_by = $39, disqualify_reason = $40,
			emails_opened = $41, emails_clicked = $42, web_visits = $43,
			form_submissions = $44, last_engagement = $45,
			updated_at = $46, updated_by = $47, version = version + 1
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL AND version = $48`

	var convertedAt, convertedBy, opportunityID, customerID, contactID interface{}
	var disqualifiedAt, disqualifiedBy, disqualifyReason interface{}

	if lead.ConversionInfo != nil {
		convertedAt = lead.ConversionInfo.ConvertedAt
		convertedBy = lead.ConversionInfo.ConvertedBy
		opportunityID = lead.ConversionInfo.OpportunityID
		if lead.ConversionInfo.CustomerID != nil {
			customerID = *lead.ConversionInfo.CustomerID
		}
		if lead.ConversionInfo.ContactID != nil {
			contactID = *lead.ConversionInfo.ContactID
		}
	}

	if lead.DisqualifyInfo != nil {
		disqualifiedAt = lead.DisqualifyInfo.DisqualifiedAt
		disqualifiedBy = lead.DisqualifyInfo.DisqualifiedBy
		disqualifyReason = lead.DisqualifyInfo.Reason
	}

	result, err := exec.ExecContext(ctx, query,
		lead.TenantID,
		lead.ID,
		lead.Contact.FirstName,
		lead.Contact.LastName,
		lead.Contact.Email,
		nullString(lead.Contact.Phone),
		nullString(lead.Contact.Mobile),
		nullString(lead.Contact.JobTitle),
		nullString(lead.Contact.Department),
		nullString(lead.Company.Name),
		nullString(lead.Company.Size),
		nullString(lead.Company.Industry),
		nullString(lead.Company.Website),
		nullString(lead.Company.Address),
		nullString(lead.Company.City),
		nullString(lead.Company.State),
		nullString(lead.Company.PostalCode),
		nullString(lead.Company.Country),
		string(lead.Status),
		string(lead.Source),
		string(lead.Rating),
		lead.Score.Score,
		lead.Score.Demographic,
		lead.Score.Behavioral,
		nullUUID(lead.OwnerID),
		nullUUID(lead.CampaignID),
		nullString(lead.Description),
		lead.EstimatedValue.Amount,
		lead.EstimatedValue.Currency,
		lead.Tags,
		customFieldsJSON,
		NewNullTime(lead.LastContactedAt).NullTime,
		convertedAt,
		convertedBy,
		opportunityID,
		customerID,
		contactID,
		disqualifiedAt,
		disqualifiedBy,
		disqualifyReason,
		lead.Engagement.EmailsOpened,
		lead.Engagement.EmailsClicked,
		lead.Engagement.WebVisits,
		lead.Engagement.FormSubmissions,
		NewNullTime(lead.Engagement.LastEngagement).NullTime,
		time.Now().UTC(),
		lead.CreatedBy,
		lead.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("lead not found or version mismatch")
	}

	lead.Version++
	return nil
}

// Delete soft-deletes a lead.
func (r *LeadRepository) Delete(ctx context.Context, tenantID, leadID uuid.UUID) error {
	exec := getExecutor(ctx, r.db)

	query := `
		UPDATE sales.leads
		SET deleted_at = $3, updated_at = $3
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`

	result, err := exec.ExecContext(ctx, query, tenantID, leadID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete lead: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("lead not found")
	}

	return nil
}

// List retrieves leads with filtering and pagination.
func (r *LeadRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.LeadFilter, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	exec := getExecutor(ctx, r.db)

	baseQuery := `
		SELECT id, tenant_id, first_name, last_name, email, phone, mobile,
			job_title, department, company_name, company_size, industry, website,
			address, city, state, postal_code, country,
			status, source, rating, score, demographic_score, behavioral_score,
			owner_id, campaign_id, description, estimated_amount, estimated_currency,
			tags, custom_fields, last_contacted_at,
			converted_at, converted_by, opportunity_id, customer_id, contact_id,
			disqualified_at, disqualified_by, disqualify_reason,
			emails_opened, emails_clicked, web_visits, form_submissions, last_engagement,
			created_at, updated_at, created_by, updated_by, deleted_at, version
		FROM sales.leads
		WHERE tenant_id = $1 AND deleted_at IS NULL`

	qb := NewQueryBuilder(baseQuery)
	qb.args = append(qb.args, tenantID)

	// Apply filters
	r.applyFilters(qb, filter)

	// Get total count
	countQuery, countArgs := qb.BuildCount()
	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, countArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to count leads: %w", err)
	}

	// Apply sorting and pagination
	sortColumn := ValidateSortColumn(opts.SortBy, allowedLeadSortColumns)
	sortOrder := ValidateSortOrder(opts.SortOrder)
	qb.OrderBy(sortColumn, sortOrder)
	qb.Limit(opts.Limit())
	qb.Offset(opts.Offset())

	query, args := qb.Build()

	var rows []leadRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list leads: %w", err)
	}

	leads := make([]*domain.Lead, 0, len(rows))
	for _, row := range rows {
		lead, err := r.toDomain(&row)
		if err != nil {
			return nil, 0, err
		}
		leads = append(leads, lead)
	}

	return leads, total, nil
}

// GetByEmail retrieves a lead by email.
func (r *LeadRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.Lead, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, first_name, last_name, email, phone, mobile,
			job_title, department, company_name, company_size, industry, website,
			address, city, state, postal_code, country,
			status, source, rating, score, demographic_score, behavioral_score,
			owner_id, campaign_id, description, estimated_amount, estimated_currency,
			tags, custom_fields, last_contacted_at,
			converted_at, converted_by, opportunity_id, customer_id, contact_id,
			disqualified_at, disqualified_by, disqualify_reason,
			emails_opened, emails_clicked, web_visits, form_submissions, last_engagement,
			created_at, updated_at, created_by, updated_by, deleted_at, version
		FROM sales.leads
		WHERE tenant_id = $1 AND LOWER(email) = LOWER($2) AND deleted_at IS NULL`

	var row leadRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, email); err != nil {
		if IsNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lead by email: %w", err)
	}

	return r.toDomain(&row)
}

// GetByPhone retrieves a lead by phone.
func (r *LeadRepository) GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*domain.Lead, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, first_name, last_name, email, phone, mobile,
			job_title, department, company_name, company_size, industry, website,
			address, city, state, postal_code, country,
			status, source, rating, score, demographic_score, behavioral_score,
			owner_id, campaign_id, description, estimated_amount, estimated_currency,
			tags, custom_fields, last_contacted_at,
			converted_at, converted_by, opportunity_id, customer_id, contact_id,
			disqualified_at, disqualified_by, disqualify_reason,
			emails_opened, emails_clicked, web_visits, form_submissions, last_engagement,
			created_at, updated_at, created_by, updated_by, deleted_at, version
		FROM sales.leads
		WHERE tenant_id = $1 AND (phone = $2 OR mobile = $2) AND deleted_at IS NULL`

	var row leadRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, phone); err != nil {
		if IsNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lead by phone: %w", err)
	}

	return r.toDomain(&row)
}

// GetByStatus retrieves leads by status.
func (r *LeadRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.LeadStatus, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	filter := domain.LeadFilter{
		Statuses: []domain.LeadStatus{status},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetQualifiedLeads retrieves qualified leads.
func (r *LeadRepository) GetQualifiedLeads(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return r.GetByStatus(ctx, tenantID, domain.LeadStatusQualified, opts)
}

// GetUnassignedLeads retrieves leads without an owner.
func (r *LeadRepository) GetUnassignedLeads(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	unassigned := true
	filter := domain.LeadFilter{
		Unassigned: &unassigned,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetByOwner retrieves leads by owner.
func (r *LeadRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	filter := domain.LeadFilter{
		OwnerIDs: []uuid.UUID{ownerID},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetBySource retrieves leads by source.
func (r *LeadRepository) GetBySource(ctx context.Context, tenantID uuid.UUID, source domain.LeadSource, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	filter := domain.LeadFilter{
		Sources: []domain.LeadSource{source},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetHighScoreLeads retrieves leads with score >= minScore.
func (r *LeadRepository) GetHighScoreLeads(ctx context.Context, tenantID uuid.UUID, minScore int, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	filter := domain.LeadFilter{
		MinScore: &minScore,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetLeadsForNurturing retrieves leads in nurturing status.
func (r *LeadRepository) GetLeadsForNurturing(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return r.GetByStatus(ctx, tenantID, domain.LeadStatusNurturing, opts)
}

// GetCreatedBetween retrieves leads created within a time range.
func (r *LeadRepository) GetCreatedBetween(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	filter := domain.LeadFilter{
		CreatedAfter:  &start,
		CreatedBefore: &end,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetUpdatedSince retrieves leads updated since a specific time.
func (r *LeadRepository) GetUpdatedSince(ctx context.Context, tenantID uuid.UUID, since time.Time, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	filter := domain.LeadFilter{
		UpdatedAfter: &since,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetStaleLeads retrieves leads not contacted for staleDays.
func (r *LeadRepository) GetStaleLeads(ctx context.Context, tenantID uuid.UUID, staleDays int, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	exec := getExecutor(ctx, r.db)

	staleDate := time.Now().AddDate(0, 0, -staleDays)

	query := `
		SELECT id, tenant_id, first_name, last_name, email, phone, mobile,
			job_title, department, company_name, company_size, industry, website,
			address, city, state, postal_code, country,
			status, source, rating, score, demographic_score, behavioral_score,
			owner_id, campaign_id, description, estimated_amount, estimated_currency,
			tags, custom_fields, last_contacted_at,
			converted_at, converted_by, opportunity_id, customer_id, contact_id,
			disqualified_at, disqualified_by, disqualify_reason,
			emails_opened, emails_clicked, web_visits, form_submissions, last_engagement,
			created_at, updated_at, created_by, updated_by, deleted_at, version
		FROM sales.leads
		WHERE tenant_id = $1
			AND deleted_at IS NULL
			AND status NOT IN ('converted', 'disqualified')
			AND (last_contacted_at IS NULL OR last_contacted_at < $2)
		ORDER BY last_contacted_at ASC NULLS FIRST
		LIMIT $3 OFFSET $4`

	var rows []leadRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, tenantID, staleDate, opts.Limit(), opts.Offset()); err != nil {
		return nil, 0, fmt.Errorf("failed to get stale leads: %w", err)
	}

	// Get count
	countQuery := `
		SELECT COUNT(*)
		FROM sales.leads
		WHERE tenant_id = $1
			AND deleted_at IS NULL
			AND status NOT IN ('converted', 'disqualified')
			AND (last_contacted_at IS NULL OR last_contacted_at < $2)`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, tenantID, staleDate); err != nil {
		return nil, 0, fmt.Errorf("failed to count stale leads: %w", err)
	}

	leads := make([]*domain.Lead, 0, len(rows))
	for _, row := range rows {
		lead, err := r.toDomain(&row)
		if err != nil {
			return nil, 0, err
		}
		leads = append(leads, lead)
	}

	return leads, total, nil
}

// BulkCreate inserts multiple leads.
func (r *LeadRepository) BulkCreate(ctx context.Context, leads []*domain.Lead) error {
	if len(leads) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	for _, lead := range leads {
		customFieldsJSON, err := ToJSON(lead.CustomFields)
		if err != nil {
			return fmt.Errorf("failed to marshal custom fields: %w", err)
		}

		query := `
			INSERT INTO sales.leads (
				id, tenant_id, first_name, last_name, email, phone, mobile,
				job_title, department, company_name, company_size, industry, website,
				address, city, state, postal_code, country,
				status, source, rating, score, demographic_score, behavioral_score,
				owner_id, campaign_id, description, estimated_amount, estimated_currency,
				tags, custom_fields, created_at, updated_at, created_by, updated_by, version
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
				$14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24,
				$25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36
			)`

		_, err = exec.ExecContext(ctx, query,
			lead.ID, lead.TenantID,
			lead.Contact.FirstName, lead.Contact.LastName, lead.Contact.Email,
			nullString(lead.Contact.Phone), nullString(lead.Contact.Mobile),
			nullString(lead.Contact.JobTitle), nullString(lead.Contact.Department),
			nullString(lead.Company.Name), nullString(lead.Company.Size),
			nullString(lead.Company.Industry), nullString(lead.Company.Website),
			nullString(lead.Company.Address), nullString(lead.Company.City),
			nullString(lead.Company.State), nullString(lead.Company.PostalCode),
			nullString(lead.Company.Country),
			string(lead.Status), string(lead.Source), string(lead.Rating),
			lead.Score.Score, lead.Score.Demographic, lead.Score.Behavioral,
			nullUUID(lead.OwnerID), nullUUID(lead.CampaignID),
			nullString(lead.Description),
			lead.EstimatedValue.Amount, lead.EstimatedValue.Currency,
			lead.Tags, customFieldsJSON,
			lead.CreatedAt, lead.UpdatedAt, lead.CreatedBy, lead.CreatedBy, lead.Version,
		)

		if err != nil {
			return fmt.Errorf("failed to bulk create leads: %w", err)
		}
	}

	return nil
}

// BulkUpdateOwner updates owner for multiple leads.
func (r *LeadRepository) BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, leadIDs []uuid.UUID, newOwnerID uuid.UUID) error {
	if len(leadIDs) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	query := `
		UPDATE sales.leads
		SET owner_id = $2, updated_at = $3
		WHERE tenant_id = $1 AND id = ANY($4) AND deleted_at IS NULL`

	_, err := exec.ExecContext(ctx, query, tenantID, newOwnerID, time.Now().UTC(), leadIDs)
	if err != nil {
		return fmt.Errorf("failed to bulk update owner: %w", err)
	}

	return nil
}

// BulkUpdateStatus updates status for multiple leads.
func (r *LeadRepository) BulkUpdateStatus(ctx context.Context, tenantID uuid.UUID, leadIDs []uuid.UUID, status domain.LeadStatus) error {
	if len(leadIDs) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	query := `
		UPDATE sales.leads
		SET status = $2, updated_at = $3
		WHERE tenant_id = $1 AND id = ANY($4) AND deleted_at IS NULL`

	_, err := exec.ExecContext(ctx, query, tenantID, string(status), time.Now().UTC(), leadIDs)
	if err != nil {
		return fmt.Errorf("failed to bulk update status: %w", err)
	}

	return nil
}

// CountByStatus counts leads by status.
func (r *LeadRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.LeadStatus]int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT status, COUNT(*) as count
		FROM sales.leads
		WHERE tenant_id = $1 AND deleted_at IS NULL
		GROUP BY status`

	rows, err := exec.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.LeadStatus]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[domain.LeadStatus(status)] = count
	}

	return result, nil
}

// CountBySource counts leads by source.
func (r *LeadRepository) CountBySource(ctx context.Context, tenantID uuid.UUID) (map[domain.LeadSource]int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT source, COUNT(*) as count
		FROM sales.leads
		WHERE tenant_id = $1 AND deleted_at IS NULL
		GROUP BY source`

	rows, err := exec.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by source: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.LeadSource]int64)
	for rows.Next() {
		var source string
		var count int64
		if err := rows.Scan(&source, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[domain.LeadSource(source)] = count
	}

	return result, nil
}

// GetConversionRate calculates the lead conversion rate.
func (r *LeadRepository) GetConversionRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT
			COALESCE(
				CAST(SUM(CASE WHEN status = 'converted' THEN 1 ELSE 0 END) AS FLOAT) /
				NULLIF(COUNT(*), 0),
				0
			) as conversion_rate
		FROM sales.leads
		WHERE tenant_id = $1
			AND deleted_at IS NULL
			AND created_at >= $2
			AND created_at < $3`

	var rate float64
	if err := sqlx.GetContext(ctx, exec, &rate, query, tenantID, start, end); err != nil {
		return 0, fmt.Errorf("failed to get conversion rate: %w", err)
	}

	return rate, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func (r *LeadRepository) applyFilters(qb *QueryBuilder, filter domain.LeadFilter) {
	// Status filter
	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, s := range filter.Statuses {
			statuses[i] = string(s)
		}
		qb.WhereInStrings("status", statuses)
	}

	// Source filter
	if len(filter.Sources) > 0 {
		sources := make([]string, len(filter.Sources))
		for i, s := range filter.Sources {
			sources[i] = string(s)
		}
		qb.WhereInStrings("source", sources)
	}

	// Owner filter
	if len(filter.OwnerIDs) > 0 {
		qb.WhereIn("owner_id", filter.OwnerIDs)
	}

	// Unassigned filter
	if filter.Unassigned != nil && *filter.Unassigned {
		qb.Where("owner_id IS NULL")
	}

	// Score filters
	if filter.MinScore != nil {
		qb.Where(fmt.Sprintf("score >= $%d", qb.NextParam()), *filter.MinScore)
	}
	if filter.MaxScore != nil {
		qb.Where(fmt.Sprintf("score <= $%d", qb.NextParam()), *filter.MaxScore)
	}

	// Time filters
	if filter.CreatedAfter != nil {
		qb.Where(fmt.Sprintf("created_at >= $%d", qb.NextParam()), *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		qb.Where(fmt.Sprintf("created_at < $%d", qb.NextParam()), *filter.CreatedBefore)
	}
	if filter.UpdatedAfter != nil {
		qb.Where(fmt.Sprintf("updated_at >= $%d", qb.NextParam()), *filter.UpdatedAfter)
	}

	// Search filter
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		qb.Where(fmt.Sprintf(`(
			first_name ILIKE $%d OR
			last_name ILIKE $%d OR
			email ILIKE $%d OR
			company_name ILIKE $%d
		)`, qb.NextParam(), qb.NextParam(), qb.NextParam(), qb.NextParam()),
			searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Campaign filter
	if filter.CampaignID != nil {
		qb.Where(fmt.Sprintf("campaign_id = $%d", qb.NextParam()), *filter.CampaignID)
	}
}

func (r *LeadRepository) toDomain(row *leadRow) (*domain.Lead, error) {
	lead := &domain.Lead{
		ID:       row.ID,
		TenantID: row.TenantID,
		Contact: domain.LeadContact{
			FirstName:  row.FirstName,
			LastName:   row.LastName,
			Email:      row.Email,
			Phone:      row.Phone.String,
			Mobile:     row.Mobile.String,
			JobTitle:   row.JobTitle.String,
			Department: row.Department.String,
		},
		Company: domain.LeadCompany{
			Name:       row.CompanyName.String,
			Size:       row.CompanySize.String,
			Industry:   row.Industry.String,
			Website:    row.Website.String,
			Address:    row.Address.String,
			City:       row.City.String,
			State:      row.State.String,
			PostalCode: row.PostalCode.String,
			Country:    row.Country.String,
		},
		Status: domain.LeadStatus(row.Status),
		Source: domain.LeadSource(row.Source),
		Rating: domain.LeadRating(row.Rating),
		Score: domain.LeadScore{
			Score:       row.Score,
			Demographic: row.DemographicScore,
			Behavioral:  row.BehavioralScore,
		},
		Description: row.Description.String,
		Tags:        []string(row.Tags),
		Engagement: domain.LeadEngagement{
			EmailsOpened:    row.EmailsOpened,
			EmailsClicked:   row.EmailsClicked,
			WebVisits:       row.WebVisits,
			FormSubmissions: row.FormSubmissions,
			LastEngagement:  NullTime{row.LastEngagement}.TimePtr(),
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		CreatedBy: row.CreatedBy,
		Version:   row.Version,
	}

	// Owner
	if row.OwnerID.Valid {
		lead.OwnerID = &row.OwnerID.UUID
	}

	// Campaign
	if row.CampaignID.Valid {
		lead.CampaignID = &row.CampaignID.UUID
	}

	// Estimated value
	if row.EstimatedAmount > 0 {
		lead.EstimatedValue = domain.Money{
			Amount:   row.EstimatedAmount,
			Currency: row.EstimatedCurrency,
		}
	}

	// Last contacted
	if row.LastContactedAt.Valid {
		lead.LastContactedAt = &row.LastContactedAt.Time
	}

	// Conversion info
	if row.ConvertedAt.Valid {
		lead.ConversionInfo = &domain.LeadConversionInfo{
			ConvertedAt:   row.ConvertedAt.Time,
			ConvertedBy:   row.ConvertedBy.UUID,
			OpportunityID: row.OpportunityID.UUID,
		}
		if row.CustomerID.Valid {
			lead.ConversionInfo.CustomerID = &row.CustomerID.UUID
		}
		if row.ContactID.Valid {
			lead.ConversionInfo.ContactID = &row.ContactID.UUID
		}
	}

	// Disqualify info
	if row.DisqualifiedAt.Valid {
		lead.DisqualifyInfo = &domain.LeadDisqualifyInfo{
			DisqualifiedAt: row.DisqualifiedAt.Time,
			DisqualifiedBy: row.DisqualifiedBy.UUID,
			Reason:         row.DisqualifyReason.String,
		}
	}

	// Custom fields
	if row.CustomFields.Valid {
		var customFields map[string]interface{}
		if err := json.Unmarshal([]byte(row.CustomFields.String), &customFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
		}
		lead.CustomFields = customFields
	}

	return lead, nil
}

// nullString converts a string to sql.NullString.
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullUUID converts a UUID pointer to uuid.NullUUID.
func nullUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}
