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

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/domain"
)

// UserRow represents a user database row.
type UserRow struct {
	ID              uuid.UUID       `db:"id"`
	TenantID        uuid.UUID       `db:"tenant_id"`
	Email           string          `db:"email"`
	PasswordHash    string          `db:"password_hash"`
	FirstName       sql.NullString  `db:"first_name"`
	LastName        sql.NullString  `db:"last_name"`
	AvatarURL       sql.NullString  `db:"avatar_url"`
	Phone           sql.NullString  `db:"phone"`
	Status          string          `db:"status"`
	EmailVerifiedAt *time.Time      `db:"email_verified_at"`
	LastLoginAt     *time.Time      `db:"last_login_at"`
	Metadata        json.RawMessage `db:"metadata"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
	DeletedAt       *time.Time      `db:"deleted_at"`
}

// ToEntity converts a UserRow to a User domain entity.
func (r *UserRow) ToEntity() *domain.User {
	email, _ := domain.NewEmail(r.Email)
	password := domain.NewPasswordFromHash(r.PasswordHash)

	var metadata map[string]interface{}
	if len(r.Metadata) > 0 {
		_ = json.Unmarshal(r.Metadata, &metadata)
	}

	return domain.ReconstructUser(
		r.ID,
		r.TenantID,
		email,
		password,
		r.FirstName.String,
		r.LastName.String,
		r.AvatarURL.String,
		r.Phone.String,
		domain.UserStatus(r.Status),
		r.EmailVerifiedAt,
		r.LastLoginAt,
		metadata,
		r.CreatedAt,
		r.UpdatedAt,
		r.DeletedAt,
	)
}

// UserRepository implements domain.UserRepository using PostgreSQL.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	metadata, err := json.Marshal(user.Metadata())
	if err != nil {
		metadata = []byte("{}")
	}

	query := `
		INSERT INTO users (
			id, tenant_id, email, password_hash, first_name, last_name,
			avatar_url, phone, status, email_verified_at, last_login_at,
			metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	_, err = r.getDB(ctx).ExecContext(ctx, query,
		user.GetID(),
		user.TenantID(),
		user.Email().String(),
		user.PasswordHash().Hash(),
		nullString(user.FirstName()),
		nullString(user.LastName()),
		nullString(user.AvatarURL()),
		nullString(user.Phone()),
		user.Status().String(),
		user.EmailVerifiedAt(),
		user.LastLoginAt(),
		metadata,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Update updates an existing user.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	metadata, err := json.Marshal(user.Metadata())
	if err != nil {
		metadata = []byte("{}")
	}

	query := `
		UPDATE users SET
			email = $1,
			password_hash = $2,
			first_name = $3,
			last_name = $4,
			avatar_url = $5,
			phone = $6,
			status = $7,
			email_verified_at = $8,
			last_login_at = $9,
			metadata = $10,
			updated_at = $11
		WHERE id = $12 AND deleted_at IS NULL`

	result, err := r.getDB(ctx).ExecContext(ctx, query,
		user.Email().String(),
		user.PasswordHash().Hash(),
		nullString(user.FirstName()),
		nullString(user.LastName()),
		nullString(user.AvatarURL()),
		nullString(user.Phone()),
		user.Status().String(),
		user.EmailVerifiedAt(),
		user.LastLoginAt(),
		metadata,
		user.UpdatedAt,
		user.GetID(),
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete soft deletes a user.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users SET
			deleted_at = $1,
			updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// FindByID finds a user by ID.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name,
			avatar_url, phone, status, email_verified_at, last_login_at,
			metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	var row UserRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return row.ToEntity(), nil
}

// FindByEmail finds a user by email within a tenant.
func (r *UserRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, password_hash, first_name, last_name,
			avatar_url, phone, status, email_verified_at, last_login_at,
			metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND LOWER(email) = LOWER($2) AND deleted_at IS NULL`

	var row UserRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, tenantID, email.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return row.ToEntity(), nil
}

// FindByTenant finds all users for a tenant with pagination.
func (r *UserRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, opts domain.UserQueryOptions) ([]*domain.User, int64, error) {
	// Build WHERE clause
	where := []string{"tenant_id = $1", "deleted_at IS NULL"}
	args := []interface{}{tenantID}
	argIndex := 2

	if opts.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, opts.Status.String())
		argIndex++
	}

	if opts.Search != "" {
		where = append(where, fmt.Sprintf("(LOWER(email) LIKE $%d OR LOWER(first_name) LIKE $%d OR LOWER(last_name) LIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+strings.ToLower(opts.Search)+"%")
		argIndex++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	var total int64
	err := r.getDB(ctx).GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
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

	// Query users
	query := fmt.Sprintf(`
		SELECT id, tenant_id, email, password_hash, first_name, last_name,
			avatar_url, phone, status, email_verified_at, last_login_at,
			metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, orderDir, argIndex, argIndex+1)

	args = append(args, opts.PageSize, offset)

	var rows []UserRow
	err = r.getDB(ctx).SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find users: %w", err)
	}

	users := make([]*domain.User, len(rows))
	for i, row := range rows {
		users[i] = row.ToEntity()
	}

	return users, total, nil
}

// FindByRoleID finds all users with a specific role.
func (r *UserRepository) FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
	query := `
		SELECT u.id, u.tenant_id, u.email, u.password_hash, u.first_name, u.last_name,
			u.avatar_url, u.phone, u.status, u.email_verified_at, u.last_login_at,
			u.metadata, u.created_at, u.updated_at, u.deleted_at
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		WHERE ur.role_id = $1 AND u.deleted_at IS NULL`

	var rows []UserRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by role: %w", err)
	}

	users := make([]*domain.User, len(rows))
	for i, row := range rows {
		users[i] = row.ToEntity()
	}

	return users, nil
}

// ExistsByEmail checks if a user with the email exists in the tenant.
func (r *UserRepository) ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE tenant_id = $1 AND LOWER(email) = LOWER($2) AND deleted_at IS NULL
		)`

	var exists bool
	err := r.getDB(ctx).GetContext(ctx, &exists, query, tenantID, email.String())
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// CountByTenant returns the number of users for a tenant.
func (r *UserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND deleted_at IS NULL`

	var count int64
	err := r.getDB(ctx).GetContext(ctx, &count, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// getDB returns the database connection, checking for transaction in context.
func (r *UserRepository) getDB(ctx context.Context) sqlx.ExtContext {
	if tx := getTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}

// Helper functions

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "unique") ||
		strings.Contains(err.Error(), "duplicate") ||
		strings.Contains(err.Error(), "23505")
}
