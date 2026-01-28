// Package postgres contains PostgreSQL repository implementations for the sales service.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Idempotency Repository
// ============================================================================

// idempotencyKeyRow represents an idempotency key database row.
type idempotencyKeyRow struct {
	Key        string    `db:"key"`
	TenantID   uuid.UUID `db:"tenant_id"`
	ResourceID uuid.UUID `db:"resource_id"`
	ExpiresAt  time.Time `db:"expires_at"`
	CreatedAt  time.Time `db:"created_at"`
}

// IdempotencyRepository implements domain.IdempotencyRepository for PostgreSQL.
type IdempotencyRepository struct {
	db *sqlx.DB
}

// NewIdempotencyRepository creates a new IdempotencyRepository.
func NewIdempotencyRepository(db *sqlx.DB) *IdempotencyRepository {
	return &IdempotencyRepository{db: db}
}

// Store saves an idempotency key with its associated resource ID.
func (r *IdempotencyRepository) Store(ctx context.Context, key *domain.IdempotencyKey) error {
	exec := getExecutor(ctx, r.db)

	query := `
		INSERT INTO sales.idempotency_keys (key, tenant_id, resource_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, key) DO NOTHING`

	_, err := exec.ExecContext(ctx, query,
		key.Key,
		key.TenantID,
		key.ResourceID,
		key.ExpiresAt,
		key.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store idempotency key: %w", err)
	}

	return nil
}

// Get retrieves the idempotency key entry.
func (r *IdempotencyRepository) Get(ctx context.Context, tenantID uuid.UUID, key string) (*domain.IdempotencyKey, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT key, tenant_id, resource_id, expires_at, created_at
		FROM sales.idempotency_keys
		WHERE tenant_id = $1 AND key = $2 AND expires_at > $3`

	var row idempotencyKeyRow
	if err := sqlx.GetContext(ctx, exec, &row, query, tenantID, key, time.Now().UTC()); err != nil {
		if IsNotFoundError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get idempotency key: %w", err)
	}

	return &domain.IdempotencyKey{
		Key:        row.Key,
		TenantID:   row.TenantID,
		ResourceID: row.ResourceID,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
	}, nil
}

// Exists checks if an idempotency key exists and is not expired.
func (r *IdempotencyRepository) Exists(ctx context.Context, tenantID uuid.UUID, key string) (bool, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM sales.idempotency_keys
			WHERE tenant_id = $1 AND key = $2 AND expires_at > $3
		)`

	var exists bool
	if err := sqlx.GetContext(ctx, exec, &exists, query, tenantID, key, time.Now().UTC()); err != nil {
		return false, fmt.Errorf("failed to check idempotency key: %w", err)
	}

	return exists, nil
}

// Delete removes an idempotency key.
func (r *IdempotencyRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	exec := getExecutor(ctx, r.db)

	query := `
		DELETE FROM sales.idempotency_keys
		WHERE tenant_id = $1 AND key = $2`

	_, err := exec.ExecContext(ctx, query, tenantID, key)
	if err != nil {
		return fmt.Errorf("failed to delete idempotency key: %w", err)
	}

	return nil
}

// DeleteExpired removes all expired idempotency keys.
func (r *IdempotencyRepository) DeleteExpired(ctx context.Context) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		DELETE FROM sales.idempotency_keys
		WHERE expires_at < $1`

	result, err := exec.ExecContext(ctx, query, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired idempotency keys: %w", err)
	}

	return result.RowsAffected()
}

// Extend extends the expiration time of an idempotency key.
func (r *IdempotencyRepository) Extend(ctx context.Context, tenantID uuid.UUID, key string, newExpiry time.Time) error {
	exec := getExecutor(ctx, r.db)

	query := `
		UPDATE sales.idempotency_keys
		SET expires_at = $3
		WHERE tenant_id = $1 AND key = $2`

	result, err := exec.ExecContext(ctx, query, tenantID, key, newExpiry)
	if err != nil {
		return fmt.Errorf("failed to extend idempotency key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("idempotency key not found")
	}

	return nil
}

// GetOrCreate attempts to get an existing idempotency key or create a new one atomically.
// Returns the idempotency key and a boolean indicating if it was newly created.
func (r *IdempotencyRepository) GetOrCreate(ctx context.Context, key *domain.IdempotencyKey) (*domain.IdempotencyKey, bool, error) {
	exec := getExecutor(ctx, r.db)

	// Try to insert, if conflict return existing
	query := `
		INSERT INTO sales.idempotency_keys (key, tenant_id, resource_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, key) DO UPDATE SET key = EXCLUDED.key
		RETURNING key, tenant_id, resource_id, expires_at, created_at,
			(xmax = 0) as is_new`

	var row struct {
		idempotencyKeyRow
		IsNew bool `db:"is_new"`
	}

	err := sqlx.GetContext(ctx, exec, &row, query,
		key.Key,
		key.TenantID,
		key.ResourceID,
		key.ExpiresAt,
		key.CreatedAt,
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get or create idempotency key: %w", err)
	}

	result := &domain.IdempotencyKey{
		Key:        row.Key,
		TenantID:   row.TenantID,
		ResourceID: row.ResourceID,
		ExpiresAt:  row.ExpiresAt,
		CreatedAt:  row.CreatedAt,
	}

	return result, row.IsNew, nil
}

// CountActive returns the number of active (non-expired) idempotency keys for a tenant.
func (r *IdempotencyRepository) CountActive(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	exec := getExecutor(ctx, r.db)

	query := `
		SELECT COUNT(*)
		FROM sales.idempotency_keys
		WHERE tenant_id = $1 AND expires_at > $2`

	var count int64
	if err := sqlx.GetContext(ctx, exec, &count, query, tenantID, time.Now().UTC()); err != nil {
		return 0, fmt.Errorf("failed to count active idempotency keys: %w", err)
	}

	return count, nil
}

// CleanupOld removes old expired idempotency keys in batches to prevent long-running transactions.
func (r *IdempotencyRepository) CleanupOld(ctx context.Context, batchSize int) (int64, error) {
	exec := getExecutor(ctx, r.db)

	var totalDeleted int64

	for {
		query := `
			DELETE FROM sales.idempotency_keys
			WHERE ctid IN (
				SELECT ctid FROM sales.idempotency_keys
				WHERE expires_at < $1
				LIMIT $2
			)`

		result, err := exec.ExecContext(ctx, query, time.Now().UTC(), batchSize)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to cleanup old idempotency keys: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalDeleted += rowsAffected

		// If we deleted fewer than batch size, we're done
		if rowsAffected < int64(batchSize) {
			break
		}
	}

	return totalDeleted, nil
}
