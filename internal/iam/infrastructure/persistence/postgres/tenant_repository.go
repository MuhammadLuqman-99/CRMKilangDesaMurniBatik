// Package postgres contains PostgreSQL repository implementations.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// TenantRow represents a tenant database row.
type TenantRow struct {
	ID        uuid.UUID       `db:"id"`
	Name      string          `db:"name"`
	Slug      string          `db:"slug"`
	Status    string          `db:"status"`
	Plan      string          `db:"plan"`
	Settings  json.RawMessage `db:"settings"`
	Metadata  json.RawMessage `db:"metadata"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
	DeletedAt *time.Time      `db:"deleted_at"`
}

// ToEntity converts a TenantRow to a Tenant domain entity.
func (r *TenantRow) ToEntity() *domain.Tenant {
	var settings domain.TenantSettings
	if len(r.Settings) > 0 {
		_ = json.Unmarshal(r.Settings, &settings)
	} else {
		settings = domain.DefaultTenantSettings()
	}

	var metadata map[string]interface{}
	if len(r.Metadata) > 0 {
		_ = json.Unmarshal(r.Metadata, &metadata)
	}

	return domain.ReconstructTenant(
		r.ID,
		r.Name,
		r.Slug,
		domain.TenantStatus(r.Status),
		domain.TenantPlan(r.Plan),
		settings,
		metadata,
		r.CreatedAt,
		r.UpdatedAt,
		r.DeletedAt,
	)
}

// TenantRepository implements domain.TenantRepository using PostgreSQL.
type TenantRepository struct {
	db *sqlx.DB
}

// NewTenantRepository creates a new TenantRepository.
func NewTenantRepository(db *sqlx.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create creates a new tenant.
func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	settings, err := json.Marshal(tenant.Settings())
	if err != nil {
		settings = []byte("{}")
	}

	metadata, err := json.Marshal(tenant.Metadata())
	if err != nil {
		metadata = []byte("{}")
	}

	query := `
		INSERT INTO tenants (id, name, slug, status, plan, settings, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = r.getDB(ctx).ExecContext(ctx, query,
		tenant.GetID(),
		tenant.Name(),
		tenant.Slug(),
		tenant.Status().String(),
		tenant.Plan().String(),
		settings,
		metadata,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrTenantSlugExists
		}
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// Update updates an existing tenant.
func (r *TenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	settings, err := json.Marshal(tenant.Settings())
	if err != nil {
		settings = []byte("{}")
	}

	metadata, err := json.Marshal(tenant.Metadata())
	if err != nil {
		metadata = []byte("{}")
	}

	query := `
		UPDATE tenants SET
			name = $1,
			status = $2,
			plan = $3,
			settings = $4,
			metadata = $5,
			updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL`

	result, err := r.getDB(ctx).ExecContext(ctx, query,
		tenant.Name(),
		tenant.Status().String(),
		tenant.Plan().String(),
		settings,
		metadata,
		tenant.UpdatedAt,
		tenant.GetID(),
	)

	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrTenantNotFound
	}

	return nil
}

// Delete soft deletes a tenant.
func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE tenants SET
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrTenantNotFound
	}

	return nil
}

// FindByID finds a tenant by ID.
func (r *TenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, status, plan, settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL`

	var row TenantRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to find tenant: %w", err)
	}

	return row.ToEntity(), nil
}

// FindBySlug finds a tenant by slug.
func (r *TenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, slug, status, plan, settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE LOWER(slug) = LOWER($1) AND deleted_at IS NULL`

	var row TenantRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to find tenant by slug: %w", err)
	}

	return row.ToEntity(), nil
}

// FindAll finds all tenants with pagination.
func (r *TenantRepository) FindAll(ctx context.Context, opts domain.TenantQueryOptions) ([]*domain.Tenant, int64, error) {
	// Build WHERE clause
	where := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if opts.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, opts.Status.String())
		argIndex++
	}

	if opts.Plan != nil {
		where = append(where, fmt.Sprintf("plan = $%d", argIndex))
		args = append(args, opts.Plan.String())
		argIndex++
	}

	if opts.Search != "" {
		where = append(where, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(slug) LIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+strings.ToLower(opts.Search)+"%")
		argIndex++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tenants WHERE %s", whereClause)
	var total int64
	err := r.getDB(ctx).GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Build ORDER BY
	orderBy := "created_at"
	if opts.SortBy != "" {
		orderBy = opts.SortBy
	}
	orderDir := "DESC"
	if strings.ToUpper(opts.SortDirection) == "ASC" {
		orderDir = "ASC"
	}

	// Calculate offset
	offset := (opts.Page - 1) * opts.PageSize

	// Query tenants
	query := fmt.Sprintf(`
		SELECT id, name, slug, status, plan, settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, orderDir, argIndex, argIndex+1)

	args = append(args, opts.PageSize, offset)

	var rows []TenantRow
	err = r.getDB(ctx).SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find tenants: %w", err)
	}

	tenants := make([]*domain.Tenant, len(rows))
	for i, row := range rows {
		tenants[i] = row.ToEntity()
	}

	return tenants, total, nil
}

// ExistsBySlug checks if a tenant with the slug exists.
func (r *TenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE LOWER(slug) = LOWER($1) AND deleted_at IS NULL)`

	var exists bool
	err := r.getDB(ctx).GetContext(ctx, &exists, query, slug)
	if err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}

	return exists, nil
}

// Count returns the total number of tenants.
func (r *TenantRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL`

	var count int64
	err := r.getDB(ctx).GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	return count, nil
}

// getDB returns the database connection, checking for transaction in context.
func (r *TenantRepository) getDB(ctx context.Context) sqlx.ExtContext {
	if tx := getTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}
