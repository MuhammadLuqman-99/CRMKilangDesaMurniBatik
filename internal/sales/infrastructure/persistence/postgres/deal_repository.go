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
// Deal Repository
// ============================================================================

// dealRow represents a deal database row.
type dealRow struct {
	ID                   uuid.UUID      `db:"id"`
	TenantID             uuid.UUID      `db:"tenant_id"`
	Code                 string         `db:"code"`
	Name                 string         `db:"name"`
	Description          sql.NullString `db:"description"`
	Status               string         `db:"status"`
	OpportunityID        uuid.UUID      `db:"opportunity_id"`
	CustomerID           uuid.UUID      `db:"customer_id"`
	CustomerName         sql.NullString `db:"customer_name"`
	PrimaryContactID     uuid.NullUUID  `db:"primary_contact_id"`
	PrimaryContactName   sql.NullString `db:"primary_contact_name"`
	OwnerID              uuid.UUID      `db:"owner_id"`
	OwnerName            sql.NullString `db:"owner_name"`
	Currency             string         `db:"currency"`
	Subtotal             int64          `db:"subtotal"`
	TotalDiscount        int64          `db:"total_discount"`
	TotalTax             int64          `db:"total_tax"`
	TotalAmount          int64          `db:"total_amount"`
	PaidAmount           int64          `db:"paid_amount"`
	OutstandingAmount    int64          `db:"outstanding_amount"`
	PaymentTerm          string         `db:"payment_term"`
	ContractURL          sql.NullString `db:"contract_url"`
	Notes                sql.NullString `db:"notes"`
	Tags                 StringArray    `db:"tags"`
	CustomFields         NullableJSON   `db:"custom_fields"`
	WonAt                time.Time      `db:"won_at"`
	ContractDate         sql.NullTime   `db:"contract_date"`
	StartDate            sql.NullTime   `db:"start_date"`
	EndDate              sql.NullTime   `db:"end_date"`
	CancelledAt          sql.NullTime   `db:"cancelled_at"`
	CancelledBy          uuid.NullUUID  `db:"cancelled_by"`
	CancelReason         sql.NullString `db:"cancel_reason"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
	CreatedBy            uuid.UUID      `db:"created_by"`
	UpdatedBy            uuid.UUID      `db:"updated_by"`
	DeletedAt            sql.NullTime   `db:"deleted_at"`
	Version              int            `db:"version"`
}

// DealRepository implements domain.DealRepository for PostgreSQL.
type DealRepository struct {
	db *sqlx.DB
}

// NewDealRepository creates a new DealRepository.
func NewDealRepository(db *sqlx.DB) *DealRepository {
	return &DealRepository{db: db}
}

// allowedDealSortColumns maps API sort fields to database columns.
var allowedDealSortColumns = map[string]string{
	"created_at":   "d.created_at",
	"updated_at":   "d.updated_at",
	"won_at":       "d.won_at",
	"signed_date":  "d.contract_date",
	"total_amount": "d.total_amount",
	"deal_number":  "d.code",
	"name":         "d.name",
}

// Create inserts a new deal into the database.
func (r *DealRepository) Create(ctx context.Context, deal *domain.Deal) error {
	exec := getExecutor(ctx, r.db)

	customFieldsJSON, err := ToJSON(deal.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		INSERT INTO sales.deals (
			id, tenant_id, code, name, description, status,
			opportunity_id, customer_id, customer_name,
			primary_contact_id, primary_contact_name, owner_id, owner_name,
			currency, subtotal, total_discount, total_tax, total_amount,
			paid_amount, outstanding_amount, payment_term, contract_url, notes,
			tags, custom_fields, won_at, contract_date, start_date, end_date,
			created_at, updated_at, created_by, updated_by, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23,
			$24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34
		)`

	_, err = exec.ExecContext(ctx, query,
		deal.ID,
		deal.TenantID,
		deal.Code,
		deal.Name,
		nullString(deal.Description),
		string(deal.Status),
		deal.OpportunityID,
		deal.CustomerID,
		nullString(deal.CustomerName),
		nullUUID(deal.PrimaryContactID),
		nullString(deal.PrimaryContactName),
		deal.OwnerID,
		nullString(deal.OwnerName),
		deal.Currency,
		deal.Subtotal.Amount,
		deal.TotalDiscount.Amount,
		deal.TotalTax.Amount,
		deal.TotalAmount.Amount,
		deal.PaidAmount.Amount,
		deal.OutstandingAmount.Amount,
		string(deal.PaymentTerm),
		nullString(deal.ContractURL),
		nullString(deal.Notes),
		deal.Tags,
		customFieldsJSON,
		deal.WonAt,
		NewNullTime(deal.Timeline.ContractDate).NullTime,
		NewNullTime(deal.Timeline.StartDate).NullTime,
		NewNullTime(deal.Timeline.EndDate).NullTime,
		deal.CreatedAt,
		deal.UpdatedAt,
		deal.CreatedBy,
		deal.CreatedBy,
		deal.Version,
	)

	if err != nil {
		if IsUniqueViolation(err) {
			return fmt.Errorf("deal with code already exists: %w", err)
		}
		return fmt.Errorf("failed to create deal: %w", err)
	}

	// Insert line items
	if err := r.insertLineItems(ctx, deal.ID, deal.TenantID, deal.LineItems); err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a deal by ID.
func (r *DealRepository) GetByID(ctx context.Context, tenantID, dealID uuid.UUID) (*domain.Deal, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT d.id, d.tenant_id, d.code, d.name, d.description, d.status,
			d.opportunity_id, d.customer_id, d.customer_name,
			d.primary_contact_id, d.primary_contact_name, d.owner_id, d.owner_name,
			d.currency, d.subtotal, d.total_discount, d.total_tax, d.total_amount,
			d.paid_amount, d.outstanding_amount, d.payment_term, d.contract_url, d.notes,
			d.tags, d.custom_fields, d.won_at, d.contract_date, d.start_date, d.end_date,
			d.cancelled_at, d.cancelled_by, d.cancel_reason,
			d.created_at, d.updated_at, d.created_by, d.updated_by, d.deleted_at, d.version
		FROM sales.deals d
		WHERE d.tenant_id = $1 AND d.id = $2 AND d.deleted_at IS NULL`

	var row dealRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, dealID); err != nil {
		if IsNotFoundError(err) {
			return nil, fmt.Errorf("deal not found")
		}
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	deal, err := r.toDomain(&row)
	if err != nil {
		return nil, err
	}

	// Load line items
	lineItems, err := r.getLineItems(ctx, dealID, tenantID)
	if err != nil {
		return nil, err
	}
	deal.LineItems = lineItems

	// Load invoices
	invoices, err := r.getInvoices(ctx, dealID, tenantID)
	if err != nil {
		return nil, err
	}
	deal.Invoices = invoices

	// Load payments
	payments, err := r.getPayments(ctx, dealID, tenantID)
	if err != nil {
		return nil, err
	}
	deal.Payments = payments

	return deal, nil
}

// Update updates an existing deal.
func (r *DealRepository) Update(ctx context.Context, deal *domain.Deal) error {
	exec := getExecutor(ctx, r.db)

	customFieldsJSON, err := ToJSON(deal.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	query := `
		UPDATE sales.deals SET
			name = $3, description = $4, status = $5,
			customer_name = $6, primary_contact_id = $7, primary_contact_name = $8,
			owner_id = $9, owner_name = $10,
			subtotal = $11, total_discount = $12, total_tax = $13, total_amount = $14,
			paid_amount = $15, outstanding_amount = $16, payment_term = $17,
			contract_url = $18, notes = $19, tags = $20, custom_fields = $21,
			contract_date = $22, start_date = $23, end_date = $24,
			cancelled_at = $25, cancelled_by = $26, cancel_reason = $27,
			updated_at = $28, updated_by = $29, version = version + 1
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL AND version = $30`

	result, err := exec.ExecContext(ctx, query,
		deal.TenantID,
		deal.ID,
		deal.Name,
		nullString(deal.Description),
		string(deal.Status),
		nullString(deal.CustomerName),
		nullUUID(deal.PrimaryContactID),
		nullString(deal.PrimaryContactName),
		deal.OwnerID,
		nullString(deal.OwnerName),
		deal.Subtotal.Amount,
		deal.TotalDiscount.Amount,
		deal.TotalTax.Amount,
		deal.TotalAmount.Amount,
		deal.PaidAmount.Amount,
		deal.OutstandingAmount.Amount,
		string(deal.PaymentTerm),
		nullString(deal.ContractURL),
		nullString(deal.Notes),
		deal.Tags,
		customFieldsJSON,
		NewNullTime(deal.Timeline.ContractDate).NullTime,
		NewNullTime(deal.Timeline.StartDate).NullTime,
		NewNullTime(deal.Timeline.EndDate).NullTime,
		NewNullTime(deal.CancelledAt).NullTime,
		nullUUID(deal.CancelledBy),
		nullString(deal.CancelReason),
		time.Now().UTC(),
		deal.CreatedBy,
		deal.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update deal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deal not found or version mismatch")
	}

	// Sync line items
	if err := r.syncLineItems(ctx, deal.ID, deal.TenantID, deal.LineItems); err != nil {
		return err
	}

	// Sync invoices
	if err := r.syncInvoices(ctx, deal.ID, deal.TenantID, deal.Invoices); err != nil {
		return err
	}

	// Sync payments
	if err := r.syncPayments(ctx, deal.ID, deal.TenantID, deal.Payments); err != nil {
		return err
	}

	deal.Version++
	return nil
}

// Delete soft-deletes a deal.
func (r *DealRepository) Delete(ctx context.Context, tenantID, dealID uuid.UUID) error {
	exec := getExecutor(ctx, r.db)

	query := `
		UPDATE sales.deals
		SET deleted_at = $3, updated_at = $3
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`

	result, err := exec.ExecContext(ctx, query, tenantID, dealID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete deal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deal not found")
	}

	return nil
}

// List retrieves deals with filtering and pagination.
func (r *DealRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.DealFilter, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	exec := getExecutor(ctx, r.db)

	baseQuery := `
		SELECT d.id, d.tenant_id, d.code, d.name, d.description, d.status,
			d.opportunity_id, d.customer_id, d.customer_name,
			d.primary_contact_id, d.primary_contact_name, d.owner_id, d.owner_name,
			d.currency, d.subtotal, d.total_discount, d.total_tax, d.total_amount,
			d.paid_amount, d.outstanding_amount, d.payment_term, d.contract_url, d.notes,
			d.tags, d.custom_fields, d.won_at, d.contract_date, d.start_date, d.end_date,
			d.cancelled_at, d.cancelled_by, d.cancel_reason,
			d.created_at, d.updated_at, d.created_by, d.updated_by, d.deleted_at, d.version
		FROM sales.deals d
		WHERE d.tenant_id = $1 AND d.deleted_at IS NULL`

	qb := NewQueryBuilder(baseQuery)
	qb.args = append(qb.args, tenantID)

	// Apply filters
	r.applyFilters(qb, filter)

	// Get total count
	countQuery, countArgs := qb.BuildCount()
	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, countArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to count deals: %w", err)
	}

	// Apply sorting and pagination
	sortColumn := ValidateSortColumn(opts.SortBy, allowedDealSortColumns)
	sortOrder := ValidateSortOrder(opts.SortOrder)
	qb.OrderBy(sortColumn, sortOrder)
	qb.Limit(opts.Limit())
	qb.Offset(opts.Offset())

	query, args := qb.Build()

	var rows []dealRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to list deals: %w", err)
	}

	deals := make([]*domain.Deal, 0, len(rows))
	for _, row := range rows {
		deal, err := r.toDomain(&row)
		if err != nil {
			return nil, 0, err
		}
		deals = append(deals, deal)
	}

	return deals, total, nil
}

// GetByNumber retrieves a deal by its code/number.
func (r *DealRepository) GetByNumber(ctx context.Context, tenantID uuid.UUID, dealNumber string) (*domain.Deal, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT d.id, d.tenant_id, d.code, d.name, d.description, d.status,
			d.opportunity_id, d.customer_id, d.customer_name,
			d.primary_contact_id, d.primary_contact_name, d.owner_id, d.owner_name,
			d.currency, d.subtotal, d.total_discount, d.total_tax, d.total_amount,
			d.paid_amount, d.outstanding_amount, d.payment_term, d.contract_url, d.notes,
			d.tags, d.custom_fields, d.won_at, d.contract_date, d.start_date, d.end_date,
			d.cancelled_at, d.cancelled_by, d.cancel_reason,
			d.created_at, d.updated_at, d.created_by, d.updated_by, d.deleted_at, d.version
		FROM sales.deals d
		WHERE d.tenant_id = $1 AND d.code = $2 AND d.deleted_at IS NULL`

	var row dealRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, dealNumber); err != nil {
		if IsNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get deal by number: %w", err)
	}

	return r.toDomain(&row)
}

// GetByStatus retrieves deals by status.
func (r *DealRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.DealStatus, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	filter := domain.DealFilter{
		Statuses: []domain.DealStatus{status},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetActiveDeals retrieves active deals.
func (r *DealRepository) GetActiveDeals(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return r.GetByStatus(ctx, tenantID, domain.DealStatusActive, opts)
}

// GetCompletedDeals retrieves completed deals.
func (r *DealRepository) GetCompletedDeals(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return r.GetByStatus(ctx, tenantID, domain.DealStatusCompleted, opts)
}

// GetByOpportunity retrieves the deal created from an opportunity.
func (r *DealRepository) GetByOpportunity(ctx context.Context, tenantID, opportunityID uuid.UUID) (*domain.Deal, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT d.id, d.tenant_id, d.code, d.name, d.description, d.status,
			d.opportunity_id, d.customer_id, d.customer_name,
			d.primary_contact_id, d.primary_contact_name, d.owner_id, d.owner_name,
			d.currency, d.subtotal, d.total_discount, d.total_tax, d.total_amount,
			d.paid_amount, d.outstanding_amount, d.payment_term, d.contract_url, d.notes,
			d.tags, d.custom_fields, d.won_at, d.contract_date, d.start_date, d.end_date,
			d.cancelled_at, d.cancelled_by, d.cancel_reason,
			d.created_at, d.updated_at, d.created_by, d.updated_by, d.deleted_at, d.version
		FROM sales.deals d
		WHERE d.tenant_id = $1 AND d.opportunity_id = $2 AND d.deleted_at IS NULL`

	var row dealRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, opportunityID); err != nil {
		if IsNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get deal by opportunity: %w", err)
	}

	return r.toDomain(&row)
}

// GetByCustomer retrieves deals for a customer.
func (r *DealRepository) GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	filter := domain.DealFilter{
		CustomerIDs: []uuid.UUID{customerID},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetByOwner retrieves deals for an owner.
func (r *DealRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	filter := domain.DealFilter{
		OwnerIDs: []uuid.UUID{ownerID},
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetDealsWithPendingPayments retrieves deals with pending payments.
func (r *DealRepository) GetDealsWithPendingPayments(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	hasPending := true
	filter := domain.DealFilter{
		HasPendingPayments: &hasPending,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetOverduePayments retrieves deals with overdue invoices.
func (r *DealRepository) GetOverduePayments(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT DISTINCT d.id, d.tenant_id, d.code, d.name, d.description, d.status,
			d.opportunity_id, d.customer_id, d.customer_name,
			d.primary_contact_id, d.primary_contact_name, d.owner_id, d.owner_name,
			d.currency, d.subtotal, d.total_discount, d.total_tax, d.total_amount,
			d.paid_amount, d.outstanding_amount, d.payment_term, d.contract_url, d.notes,
			d.tags, d.custom_fields, d.won_at, d.contract_date, d.start_date, d.end_date,
			d.cancelled_at, d.cancelled_by, d.cancel_reason,
			d.created_at, d.updated_at, d.created_by, d.updated_by, d.deleted_at, d.version
		FROM sales.deals d
		JOIN sales.deal_invoices i ON i.deal_id = d.id AND i.tenant_id = d.tenant_id
		WHERE d.tenant_id = $1
			AND d.deleted_at IS NULL
			AND i.status NOT IN ('paid', 'cancelled')
			AND i.due_date < $2
		ORDER BY d.created_at DESC
		LIMIT $3 OFFSET $4`

	var rows []dealRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, tenantID, time.Now().UTC(), opts.Limit(), opts.Offset()); err != nil {
		return nil, 0, fmt.Errorf("failed to get overdue deals: %w", err)
	}

	// Count
	countQuery := `
		SELECT COUNT(DISTINCT d.id)
		FROM sales.deals d
		JOIN sales.deal_invoices i ON i.deal_id = d.id AND i.tenant_id = d.tenant_id
		WHERE d.tenant_id = $1
			AND d.deleted_at IS NULL
			AND i.status NOT IN ('paid', 'cancelled')
			AND i.due_date < $2`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, countQuery, tenantID, time.Now().UTC()); err != nil {
		return nil, 0, fmt.Errorf("failed to count overdue deals: %w", err)
	}

	deals := make([]*domain.Deal, 0, len(rows))
	for _, row := range rows {
		deal, err := r.toDomain(&row)
		if err != nil {
			return nil, 0, err
		}
		deals = append(deals, deal)
	}

	return deals, total, nil
}

// GetFullyPaidDeals retrieves fully paid deals.
func (r *DealRepository) GetFullyPaidDeals(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	fullyPaid := true
	filter := domain.DealFilter{
		FullyPaid: &fullyPaid,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetDealsForFulfillment retrieves deals needing fulfillment.
func (r *DealRepository) GetDealsForFulfillment(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	progress := 0
	filter := domain.DealFilter{
		FulfillmentProgress: &progress,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetPartiallyFulfilledDeals retrieves partially fulfilled deals.
func (r *DealRepository) GetPartiallyFulfilledDeals(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	exec := getExecutor(ctx, r.db)

	// Deals where some but not all line items are fulfilled
	query := `
		SELECT d.id, d.tenant_id, d.code, d.name, d.description, d.status,
			d.opportunity_id, d.customer_id, d.customer_name,
			d.primary_contact_id, d.primary_contact_name, d.owner_id, d.owner_name,
			d.currency, d.subtotal, d.total_discount, d.total_tax, d.total_amount,
			d.paid_amount, d.outstanding_amount, d.payment_term, d.contract_url, d.notes,
			d.tags, d.custom_fields, d.won_at, d.contract_date, d.start_date, d.end_date,
			d.cancelled_at, d.cancelled_by, d.cancel_reason,
			d.created_at, d.updated_at, d.created_by, d.updated_by, d.deleted_at, d.version
		FROM sales.deals d
		WHERE d.tenant_id = $1
			AND d.deleted_at IS NULL
			AND d.status = 'active'
			AND EXISTS (
				SELECT 1 FROM sales.deal_line_items li
				WHERE li.deal_id = d.id AND li.fulfilled_qty > 0 AND li.fulfilled_qty < li.quantity
			)
		ORDER BY d.created_at DESC
		LIMIT $2 OFFSET $3`

	var rows []dealRow
	if err := sqlx.SelectContext(ctx, exec, &rows, query, tenantID, opts.Limit(), opts.Offset()); err != nil {
		return nil, 0, fmt.Errorf("failed to get partially fulfilled deals: %w", err)
	}

	deals := make([]*domain.Deal, 0, len(rows))
	for _, row := range rows {
		deal, err := r.toDomain(&row)
		if err != nil {
			return nil, 0, err
		}
		deals = append(deals, deal)
	}

	return deals, int64(len(deals)), nil
}

// GetByClosedDate retrieves deals by closed/won date range.
func (r *DealRepository) GetByClosedDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	filter := domain.DealFilter{
		ClosedDateAfter:  &start,
		ClosedDateBefore: &end,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetBySignedDate retrieves deals by signed date range.
func (r *DealRepository) GetBySignedDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	filter := domain.DealFilter{
		SignedDateAfter:  &start,
		SignedDateBefore: &end,
	}
	return r.List(ctx, tenantID, filter, opts)
}

// GetTotalRevenue returns total revenue in a period.
func (r *DealRepository) GetTotalRevenue(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(SUM(total_amount), 0)
		FROM sales.deals
		WHERE tenant_id = $1
			AND currency = $2
			AND deleted_at IS NULL
			AND status NOT IN ('cancelled', 'draft')
			AND won_at >= $3
			AND won_at < $4`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, query, tenantID, currency, start, end); err != nil {
		return 0, fmt.Errorf("failed to get total revenue: %w", err)
	}

	return total, nil
}

// GetTotalReceivedPayments returns total payments received in a period.
func (r *DealRepository) GetTotalReceivedPayments(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(SUM(p.amount), 0)
		FROM sales.deal_payments p
		JOIN sales.deals d ON d.id = p.deal_id AND d.tenant_id = p.tenant_id
		WHERE p.tenant_id = $1
			AND p.currency = $2
			AND p.received_at >= $3
			AND p.received_at < $4`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, query, tenantID, currency, start, end); err != nil {
		return 0, fmt.Errorf("failed to get total payments: %w", err)
	}

	return total, nil
}

// GetOutstandingAmount returns total outstanding amount.
func (r *DealRepository) GetOutstandingAmount(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(SUM(outstanding_amount), 0)
		FROM sales.deals
		WHERE tenant_id = $1
			AND currency = $2
			AND deleted_at IS NULL
			AND status NOT IN ('cancelled', 'completed')`

	var total int64
	if err := sqlx.GetContext(ctx, exec, &total, query, tenantID, currency); err != nil {
		return 0, fmt.Errorf("failed to get outstanding amount: %w", err)
	}

	return total, nil
}

// CountByStatus counts deals by status.
func (r *DealRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.DealStatus]int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT status, COUNT(*) as count
		FROM sales.deals
		WHERE tenant_id = $1 AND deleted_at IS NULL
		GROUP BY status`

	rows, err := exec.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	defer rows.Close()

	result := make(map[domain.DealStatus]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[domain.DealStatus(status)] = count
	}

	return result, nil
}

// GetAverageDealValue calculates the average deal value.
func (r *DealRepository) GetAverageDealValue(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(AVG(total_amount), 0)::bigint
		FROM sales.deals
		WHERE tenant_id = $1
			AND currency = $2
			AND deleted_at IS NULL
			AND status NOT IN ('cancelled', 'draft')
			AND won_at >= $3
			AND won_at < $4`

	var avg int64
	if err := sqlx.GetContext(ctx, exec, &avg, query, tenantID, currency, start, end); err != nil {
		return 0, fmt.Errorf("failed to get average deal value: %w", err)
	}

	return avg, nil
}

// GetMonthlyRevenue returns revenue by month for a year.
func (r *DealRepository) GetMonthlyRevenue(ctx context.Context, tenantID uuid.UUID, currency string, year int) (map[int]int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT EXTRACT(MONTH FROM won_at)::int as month, COALESCE(SUM(total_amount), 0) as total
		FROM sales.deals
		WHERE tenant_id = $1
			AND currency = $2
			AND deleted_at IS NULL
			AND status NOT IN ('cancelled', 'draft')
			AND EXTRACT(YEAR FROM won_at) = $3
		GROUP BY EXTRACT(MONTH FROM won_at)
		ORDER BY month`

	rows, err := exec.QueryContext(ctx, query, tenantID, currency, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}
	defer rows.Close()

	result := make(map[int]int64)
	for rows.Next() {
		var month int
		var total int64
		if err := rows.Scan(&month, &total); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[month] = total
	}

	return result, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func (r *DealRepository) applyFilters(qb *QueryBuilder, filter domain.DealFilter) {
	// Status filter
	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, s := range filter.Statuses {
			statuses[i] = string(s)
		}
		qb.WhereInStrings("d.status", statuses)
	}

	// Customer filter
	if len(filter.CustomerIDs) > 0 {
		qb.WhereIn("d.customer_id", filter.CustomerIDs)
	}

	// Opportunity filter
	if filter.OpportunityID != nil {
		qb.Where(fmt.Sprintf("d.opportunity_id = $%d", qb.NextParam()), *filter.OpportunityID)
	}

	// Owner filter
	if len(filter.OwnerIDs) > 0 {
		qb.WhereIn("d.owner_id", filter.OwnerIDs)
	}

	// Amount filters
	if filter.MinAmount != nil {
		qb.Where(fmt.Sprintf("d.total_amount >= $%d", qb.NextParam()), *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		qb.Where(fmt.Sprintf("d.total_amount <= $%d", qb.NextParam()), *filter.MaxAmount)
	}
	if filter.Currency != nil {
		qb.Where(fmt.Sprintf("d.currency = $%d", qb.NextParam()), *filter.Currency)
	}

	// Payment status
	if filter.HasPendingPayments != nil && *filter.HasPendingPayments {
		qb.Where("d.outstanding_amount > 0")
	}
	if filter.FullyPaid != nil && *filter.FullyPaid {
		qb.Where("d.outstanding_amount = 0 AND d.paid_amount > 0")
	}

	// Date filters
	if filter.ClosedDateAfter != nil {
		qb.Where(fmt.Sprintf("d.won_at >= $%d", qb.NextParam()), *filter.ClosedDateAfter)
	}
	if filter.ClosedDateBefore != nil {
		qb.Where(fmt.Sprintf("d.won_at < $%d", qb.NextParam()), *filter.ClosedDateBefore)
	}
	if filter.SignedDateAfter != nil {
		qb.Where(fmt.Sprintf("d.contract_date >= $%d", qb.NextParam()), *filter.SignedDateAfter)
	}
	if filter.SignedDateBefore != nil {
		qb.Where(fmt.Sprintf("d.contract_date < $%d", qb.NextParam()), *filter.SignedDateBefore)
	}

	// Search
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		qb.Where(fmt.Sprintf(`(
			d.name ILIKE $%d OR
			d.code ILIKE $%d OR
			d.customer_name ILIKE $%d
		)`, qb.NextParam(), qb.NextParam(), qb.NextParam()), searchPattern, searchPattern, searchPattern)
	}

	// Deal number
	if filter.DealNumber != nil {
		qb.Where(fmt.Sprintf("d.code = $%d", qb.NextParam()), *filter.DealNumber)
	}
}

func (r *DealRepository) toDomain(row *dealRow) (*domain.Deal, error) {
	deal := &domain.Deal{
		ID:            row.ID,
		TenantID:      row.TenantID,
		Code:          row.Code,
		Name:          row.Name,
		Description:   row.Description.String,
		Status:        domain.DealStatus(row.Status),
		OpportunityID: row.OpportunityID,
		CustomerID:    row.CustomerID,
		CustomerName:  row.CustomerName.String,
		OwnerID:       row.OwnerID,
		OwnerName:     row.OwnerName.String,
		Currency:      row.Currency,
		Subtotal:      domain.Money{Amount: row.Subtotal, Currency: row.Currency},
		TotalDiscount: domain.Money{Amount: row.TotalDiscount, Currency: row.Currency},
		TotalTax:      domain.Money{Amount: row.TotalTax, Currency: row.Currency},
		TotalAmount:   domain.Money{Amount: row.TotalAmount, Currency: row.Currency},
		PaidAmount:    domain.Money{Amount: row.PaidAmount, Currency: row.Currency},
		OutstandingAmount: domain.Money{Amount: row.OutstandingAmount, Currency: row.Currency},
		PaymentTerm:       domain.PaymentTerm(row.PaymentTerm),
		ContractURL:       row.ContractURL.String,
		Notes:             row.Notes.String,
		Tags:              []string(row.Tags),
		WonAt:             row.WonAt,
		Timeline: domain.DealTimeline{
			ContractDate: NullTime{row.ContractDate}.TimePtr(),
			StartDate:    NullTime{row.StartDate}.TimePtr(),
			EndDate:      NullTime{row.EndDate}.TimePtr(),
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		CreatedBy: row.CreatedBy,
		Version:   row.Version,
	}

	// Primary contact
	if row.PrimaryContactID.Valid {
		deal.PrimaryContactID = &row.PrimaryContactID.UUID
		deal.PrimaryContactName = row.PrimaryContactName.String
	}

	// Cancellation
	if row.CancelledAt.Valid {
		deal.CancelledAt = &row.CancelledAt.Time
	}
	if row.CancelledBy.Valid {
		deal.CancelledBy = &row.CancelledBy.UUID
	}
	deal.CancelReason = row.CancelReason.String

	// Custom fields
	if row.CustomFields.Valid {
		var customFields map[string]interface{}
		if err := json.Unmarshal([]byte(row.CustomFields.String), &customFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
		}
		deal.CustomFields = customFields
	}

	return deal, nil
}

// ============================================================================
// Line Item Operations
// ============================================================================

func (r *DealRepository) insertLineItems(ctx context.Context, dealID, tenantID uuid.UUID, items []domain.DealLineItem) error {
	if len(items) == 0 {
		return nil
	}

	exec := getExecutor(ctx, r.db)

	query := `
		INSERT INTO sales.deal_line_items (
			id, deal_id, tenant_id, product_id, product_name, product_sku, description,
			quantity, unit_price, currency, discount, discount_type, subtotal,
			tax, tax_type, tax_amount, total, fulfilled_qty, notes, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`

	for _, item := range items {
		_, err := exec.ExecContext(ctx, query,
			item.ID, dealID, tenantID, item.ProductID, item.ProductName,
			nullString(item.ProductSKU), nullString(item.Description),
			item.Quantity, item.UnitPrice.Amount, item.UnitPrice.Currency,
			item.Discount, item.DiscountType, item.Subtotal.Amount,
			item.Tax, item.TaxType, item.TaxAmount.Amount, item.Total.Amount,
			item.FulfilledQty, nullString(item.Notes), time.Now().UTC(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert line item: %w", err)
		}
	}

	return nil
}

func (r *DealRepository) getLineItems(ctx context.Context, dealID, tenantID uuid.UUID) ([]domain.DealLineItem, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, product_id, product_name, product_sku, description,
			quantity, unit_price, currency, discount, discount_type, subtotal,
			tax, tax_type, tax_amount, total, fulfilled_qty, notes
		FROM sales.deal_line_items
		WHERE deal_id = $1 AND tenant_id = $2
		ORDER BY created_at ASC`

	rows, err := exec.QueryContext(ctx, query, dealID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items: %w", err)
	}
	defer rows.Close()

	var items []domain.DealLineItem
	for rows.Next() {
		var item domain.DealLineItem
		var unitPrice, subtotal, taxAmount, total int64
		var currency string
		var sku, description, notes sql.NullString

		if err := rows.Scan(
			&item.ID, &item.ProductID, &item.ProductName, &sku, &description,
			&item.Quantity, &unitPrice, &currency, &item.Discount, &item.DiscountType,
			&subtotal, &item.Tax, &item.TaxType, &taxAmount, &total,
			&item.FulfilledQty, &notes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan line item: %w", err)
		}

		item.ProductSKU = sku.String
		item.Description = description.String
		item.Notes = notes.String
		item.UnitPrice = domain.Money{Amount: unitPrice, Currency: currency}
		item.Subtotal = domain.Money{Amount: subtotal, Currency: currency}
		item.TaxAmount = domain.Money{Amount: taxAmount, Currency: currency}
		item.Total = domain.Money{Amount: total, Currency: currency}

		items = append(items, item)
	}

	return items, nil
}

func (r *DealRepository) syncLineItems(ctx context.Context, dealID, tenantID uuid.UUID, items []domain.DealLineItem) error {
	exec := getExecutor(ctx, r.db)

	// Delete existing and re-insert
	_, err := exec.ExecContext(ctx, "DELETE FROM sales.deal_line_items WHERE deal_id = $1 AND tenant_id = $2", dealID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete line items: %w", err)
	}

	return r.insertLineItems(ctx, dealID, tenantID, items)
}

// ============================================================================
// Invoice Operations
// ============================================================================

func (r *DealRepository) getInvoices(ctx context.Context, dealID, tenantID uuid.UUID) ([]domain.Invoice, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, invoice_number, amount, currency, status, due_date,
			paid_amount, paid_at, notes, created_at
		FROM sales.deal_invoices
		WHERE deal_id = $1 AND tenant_id = $2
		ORDER BY created_at ASC`

	rows, err := exec.QueryContext(ctx, query, dealID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		var amount, paidAmount int64
		var currency string
		var paidAt sql.NullTime
		var notes sql.NullString

		if err := rows.Scan(
			&inv.ID, &inv.InvoiceNumber, &amount, &currency, &inv.Status,
			&inv.DueDate, &paidAmount, &paidAt, &notes, &inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}

		inv.Amount = domain.Money{Amount: amount, Currency: currency}
		inv.PaidAmount = domain.Money{Amount: paidAmount, Currency: currency}
		if paidAt.Valid {
			inv.PaidAt = &paidAt.Time
		}
		inv.Notes = notes.String

		invoices = append(invoices, inv)
	}

	return invoices, nil
}

func (r *DealRepository) syncInvoices(ctx context.Context, dealID, tenantID uuid.UUID, invoices []domain.Invoice) error {
	exec := getExecutor(ctx, r.db)

	for _, inv := range invoices {
		query := `
			INSERT INTO sales.deal_invoices (
				id, deal_id, tenant_id, invoice_number, amount, currency,
				status, due_date, paid_amount, paid_at, notes, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (id) DO UPDATE SET
				status = EXCLUDED.status,
				paid_amount = EXCLUDED.paid_amount,
				paid_at = EXCLUDED.paid_at`

		_, err := exec.ExecContext(ctx, query,
			inv.ID, dealID, tenantID, inv.InvoiceNumber,
			inv.Amount.Amount, inv.Amount.Currency, inv.Status, inv.DueDate,
			inv.PaidAmount.Amount, NewNullTime(inv.PaidAt).NullTime,
			nullString(inv.Notes), inv.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to sync invoice: %w", err)
		}
	}

	return nil
}

// ============================================================================
// Payment Operations
// ============================================================================

func (r *DealRepository) getPayments(ctx context.Context, dealID, tenantID uuid.UUID) ([]domain.Payment, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT id, invoice_id, amount, currency, payment_method,
			reference, notes, received_at, received_by
		FROM sales.deal_payments
		WHERE deal_id = $1 AND tenant_id = $2
		ORDER BY received_at ASC`

	rows, err := exec.QueryContext(ctx, query, dealID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var pmt domain.Payment
		var amount int64
		var currency string
		var invoiceID uuid.NullUUID
		var reference, notes sql.NullString

		if err := rows.Scan(
			&pmt.ID, &invoiceID, &amount, &currency, &pmt.PaymentMethod,
			&reference, &notes, &pmt.ReceivedAt, &pmt.ReceivedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		pmt.Amount = domain.Money{Amount: amount, Currency: currency}
		if invoiceID.Valid {
			pmt.InvoiceID = &invoiceID.UUID
		}
		pmt.Reference = reference.String
		pmt.Notes = notes.String

		payments = append(payments, pmt)
	}

	return payments, nil
}

func (r *DealRepository) syncPayments(ctx context.Context, dealID, tenantID uuid.UUID, payments []domain.Payment) error {
	exec := getExecutor(ctx, r.db)

	for _, pmt := range payments {
		query := `
			INSERT INTO sales.deal_payments (
				id, deal_id, tenant_id, invoice_id, amount, currency,
				payment_method, reference, notes, received_at, received_by
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO NOTHING`

		_, err := exec.ExecContext(ctx, query,
			pmt.ID, dealID, tenantID, nullUUID(pmt.InvoiceID),
			pmt.Amount.Amount, pmt.Amount.Currency, pmt.PaymentMethod,
			nullString(pmt.Reference), nullString(pmt.Notes),
			pmt.ReceivedAt, pmt.ReceivedBy,
		)
		if err != nil {
			return fmt.Errorf("failed to sync payment: %w", err)
		}
	}

	return nil
}
