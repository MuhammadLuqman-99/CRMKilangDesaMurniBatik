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

// RoleRow represents a role database row.
type RoleRow struct {
	ID          uuid.UUID       `db:"id"`
	TenantID    *uuid.UUID      `db:"tenant_id"`
	Name        string          `db:"name"`
	Description sql.NullString  `db:"description"`
	Permissions json.RawMessage `db:"permissions"`
	IsSystem    bool            `db:"is_system"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

// ToEntity converts a RoleRow to a Role domain entity.
func (r *RoleRow) ToEntity() *domain.Role {
	// Parse permissions
	var permStrings []string
	if len(r.Permissions) > 0 {
		_ = json.Unmarshal(r.Permissions, &permStrings)
	}

	permSet, _ := domain.NewPermissionSetFromStrings(permStrings)
	if permSet == nil {
		permSet = domain.NewPermissionSet()
	}

	return domain.ReconstructRole(
		r.ID,
		r.TenantID,
		r.Name,
		r.Description.String,
		permSet,
		r.IsSystem,
		r.CreatedAt,
		r.UpdatedAt,
	)
}

// RoleRepository implements domain.RoleRepository using PostgreSQL.
type RoleRepository struct {
	db *sqlx.DB
}

// NewRoleRepository creates a new RoleRepository.
func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role.
func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	permissions, err := json.Marshal(role.Permissions().Strings())
	if err != nil {
		permissions = []byte("[]")
	}

	query := `
		INSERT INTO roles (id, tenant_id, name, description, permissions, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.getDB(ctx).ExecContext(ctx, query,
		role.GetID(),
		role.TenantID(),
		role.Name(),
		nullString(role.Description()),
		permissions,
		role.IsSystem(),
		role.CreatedAt,
		role.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrRoleDuplicate
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// Update updates an existing role.
func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	permissions, err := json.Marshal(role.Permissions().Strings())
	if err != nil {
		permissions = []byte("[]")
	}

	query := `
		UPDATE roles SET
			name = $1,
			description = $2,
			permissions = $3,
			updated_at = $4
		WHERE id = $5`

	result, err := r.getDB(ctx).ExecContext(ctx, query,
		role.Name(),
		nullString(role.Description()),
		permissions,
		role.UpdatedAt,
		role.GetID(),
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrRoleDuplicate
		}
		return fmt.Errorf("failed to update role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrRoleNotFound
	}

	return nil
}

// Delete deletes a role.
func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1 AND is_system = false`

	result, err := r.getDB(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrRoleNotFound
	}

	return nil
}

// FindByID finds a role by ID.
func (r *RoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE id = $1`

	var row RoleRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to find role: %w", err)
	}

	return row.ToEntity(), nil
}

// FindByName finds a role by name within a tenant.
func (r *RoleRepository) FindByName(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error) {
	var query string
	var args []interface{}

	if tenantID == nil {
		query = `
			SELECT id, tenant_id, name, description, permissions, is_system, created_at, updated_at
			FROM roles
			WHERE tenant_id IS NULL AND LOWER(name) = LOWER($1)`
		args = []interface{}{name}
	} else {
		query = `
			SELECT id, tenant_id, name, description, permissions, is_system, created_at, updated_at
			FROM roles
			WHERE (tenant_id = $1 OR tenant_id IS NULL) AND LOWER(name) = LOWER($2)
			ORDER BY tenant_id DESC NULLS LAST
			LIMIT 1`
		args = []interface{}{*tenantID, name}
	}

	var row RoleRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to find role by name: %w", err)
	}

	return row.ToEntity(), nil
}

// FindByTenant finds all roles for a tenant (including system roles).
func (r *RoleRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, opts domain.RoleQueryOptions) ([]*domain.Role, int64, error) {
	// Build WHERE clause
	where := []string{"(tenant_id = $1 OR tenant_id IS NULL)"}
	args := []interface{}{tenantID}
	argIndex := 2

	if !opts.IncludeSystem {
		where = append(where, "is_system = false")
	}

	if opts.Search != "" {
		where = append(where, fmt.Sprintf("LOWER(name) LIKE $%d", argIndex))
		args = append(args, "%"+strings.ToLower(opts.Search)+"%")
		argIndex++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM roles WHERE %s", whereClause)
	var total int64
	err := r.getDB(ctx).GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// Build ORDER BY
	orderBy := "name"
	if opts.SortBy != "" {
		orderBy = opts.SortBy
	}
	orderDir := "ASC"
	if strings.ToUpper(opts.SortDirection) == "DESC" {
		orderDir = "DESC"
	}

	// Calculate offset
	offset := (opts.Page - 1) * opts.PageSize

	// Query roles
	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE %s
		ORDER BY is_system DESC, %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, orderDir, argIndex, argIndex+1)

	args = append(args, opts.PageSize, offset)

	var rows []RoleRow
	err = r.getDB(ctx).SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find roles: %w", err)
	}

	roles := make([]*domain.Role, len(rows))
	for i, row := range rows {
		roles[i] = row.ToEntity()
	}

	return roles, total, nil
}

// FindSystemRoles finds all system roles.
func (r *RoleRepository) FindSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	query := `
		SELECT id, tenant_id, name, description, permissions, is_system, created_at, updated_at
		FROM roles
		WHERE is_system = true
		ORDER BY name`

	var rows []RoleRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find system roles: %w", err)
	}

	roles := make([]*domain.Role, len(rows))
	for i, row := range rows {
		roles[i] = row.ToEntity()
	}

	return roles, nil
}

// FindByUserID finds all roles assigned to a user.
func (r *RoleRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	query := `
		SELECT r.id, r.tenant_id, r.name, r.description, r.permissions, r.is_system, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.is_system DESC, r.name`

	var rows []RoleRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find roles by user: %w", err)
	}

	roles := make([]*domain.Role, len(rows))
	for i, row := range rows {
		roles[i] = row.ToEntity()
	}

	return roles, nil
}

// ExistsByName checks if a role with the name exists in the tenant.
func (r *RoleRepository) ExistsByName(ctx context.Context, tenantID *uuid.UUID, name string) (bool, error) {
	var query string
	var args []interface{}

	if tenantID == nil {
		query = `SELECT EXISTS(SELECT 1 FROM roles WHERE tenant_id IS NULL AND LOWER(name) = LOWER($1))`
		args = []interface{}{name}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM roles WHERE tenant_id = $1 AND LOWER(name) = LOWER($2))`
		args = []interface{}{*tenantID, name}
	}

	var exists bool
	err := r.getDB(ctx).GetContext(ctx, &exists, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	return exists, nil
}

// AssignRoleToUser assigns a role to a user.
func (r *RoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id) DO NOTHING`

	_, err := r.getDB(ctx).ExecContext(ctx, query, userID, roleID, time.Now().UTC(), assignedBy)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user.
func (r *RoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	result, err := r.getDB(ctx).ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrRoleNotFound
	}

	return nil
}

// getDB returns the database connection, checking for transaction in context.
func (r *RoleRepository) getDB(ctx context.Context) sqlx.ExtContext {
	if tx := getTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}
