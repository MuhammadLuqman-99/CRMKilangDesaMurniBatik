// Package postgres contains PostgreSQL repository implementations.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// RefreshTokenRow represents a refresh token database row.
type RefreshTokenRow struct {
	ID         uuid.UUID       `db:"id"`
	UserID     uuid.UUID       `db:"user_id"`
	TokenHash  string          `db:"token_hash"`
	DeviceInfo json.RawMessage `db:"device_info"`
	IPAddress  sql.NullString  `db:"ip_address"`
	UserAgent  sql.NullString  `db:"user_agent"`
	ExpiresAt  time.Time       `db:"expires_at"`
	RevokedAt  *time.Time      `db:"revoked_at"`
	CreatedAt  time.Time       `db:"created_at"`
}

// ToEntity converts a RefreshTokenRow to a RefreshToken domain entity.
func (r *RefreshTokenRow) ToEntity() *domain.RefreshToken {
	var deviceInfo domain.DeviceInfo
	if len(r.DeviceInfo) > 0 {
		_ = json.Unmarshal(r.DeviceInfo, &deviceInfo)
	}

	return domain.ReconstructRefreshToken(
		r.ID,
		r.UserID,
		r.TokenHash,
		deviceInfo,
		r.IPAddress.String,
		r.UserAgent.String,
		r.ExpiresAt,
		r.RevokedAt,
		r.CreatedAt,
	)
}

// RefreshTokenRepository implements domain.RefreshTokenRepository using PostgreSQL.
type RefreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	deviceInfo, err := json.Marshal(token.DeviceInfo())
	if err != nil {
		deviceInfo = []byte("{}")
	}

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, device_info, ip_address, user_agent, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.getDB(ctx).ExecContext(ctx, query,
		token.GetID(),
		token.UserID(),
		token.TokenHash(),
		deviceInfo,
		nullString(token.IPAddress()),
		nullString(token.UserAgent()),
		token.ExpiresAt(),
		token.CreatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrRefreshTokenInvalid
		}
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// Update updates an existing refresh token.
func (r *RefreshTokenRepository) Update(ctx context.Context, token *domain.RefreshToken) error {
	deviceInfo, err := json.Marshal(token.DeviceInfo())
	if err != nil {
		deviceInfo = []byte("{}")
	}

	query := `
		UPDATE refresh_tokens SET
			device_info = $1,
			ip_address = $2,
			user_agent = $3,
			revoked_at = $4
		WHERE id = $5`

	result, err := r.getDB(ctx).ExecContext(ctx, query,
		deviceInfo,
		nullString(token.IPAddress()),
		nullString(token.UserAgent()),
		token.RevokedAt(),
		token.GetID(),
	)

	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	return nil
}

// Delete deletes a refresh token.
func (r *RefreshTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE id = $1`

	result, err := r.getDB(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	return nil
}

// FindByID finds a refresh token by ID.
func (r *RefreshTokenRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, user_agent, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE id = $1`

	var row RefreshTokenRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	return row.ToEntity(), nil
}

// FindByTokenHash finds a refresh token by its hash.
func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, user_agent, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`

	var row RefreshTokenRow
	err := r.getDB(ctx).GetContext(ctx, &row, query, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token by hash: %w", err)
	}

	return row.ToEntity(), nil
}

// FindByUserID finds all refresh tokens for a user.
func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, user_agent, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC`

	var rows []RefreshTokenRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find refresh tokens by user: %w", err)
	}

	tokens := make([]*domain.RefreshToken, len(rows))
	for i, row := range rows {
		tokens[i] = row.ToEntity()
	}

	return tokens, nil
}

// FindActiveByUserID finds all active (non-revoked, non-expired) refresh tokens for a user.
func (r *RefreshTokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, user_agent, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE user_id = $1
			AND revoked_at IS NULL
			AND expires_at > $2
		ORDER BY created_at DESC`

	var rows []RefreshTokenRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, userID, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to find active refresh tokens: %w", err)
	}

	tokens := make([]*domain.RefreshToken, len(rows))
	for i, row := range rows {
		tokens[i] = row.ToEntity()
	}

	return tokens, nil
}

// RevokeByUserID revokes all refresh tokens for a user.
func (r *RefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens SET
			revoked_at = $1
		WHERE user_id = $2
			AND revoked_at IS NULL`

	_, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens by user: %w", err)
	}

	return nil
}

// RevokeByID revokes a specific refresh token.
func (r *RefreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE refresh_tokens SET
			revoked_at = $1
		WHERE id = $2
			AND revoked_at IS NULL`

	result, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	return nil
}

// DeleteExpired deletes all expired refresh tokens.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1`

	result, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// CountActiveByUserID counts active refresh tokens for a user.
func (r *RefreshTokenRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM refresh_tokens
		WHERE user_id = $1
			AND revoked_at IS NULL
			AND expires_at > $2`

	var count int64
	err := r.getDB(ctx).GetContext(ctx, &count, query, userID, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to count active refresh tokens: %w", err)
	}

	return count, nil
}

// FindByDeviceID finds refresh tokens by device ID.
func (r *RefreshTokenRepository) FindByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) ([]*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, device_info, ip_address, user_agent, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE user_id = $1
			AND device_info->>'device_id' = $2
		ORDER BY created_at DESC`

	var rows []RefreshTokenRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, userID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find refresh tokens by device: %w", err)
	}

	tokens := make([]*domain.RefreshToken, len(rows))
	for i, row := range rows {
		tokens[i] = row.ToEntity()
	}

	return tokens, nil
}

// RevokeByDeviceID revokes all refresh tokens for a specific device.
func (r *RefreshTokenRepository) RevokeByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error {
	query := `
		UPDATE refresh_tokens SET
			revoked_at = $1
		WHERE user_id = $2
			AND device_info->>'device_id' = $3
			AND revoked_at IS NULL`

	_, err := r.getDB(ctx).ExecContext(ctx, query, time.Now().UTC(), userID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens by device: %w", err)
	}

	return nil
}

// CleanupOldTokens removes old revoked or expired tokens.
func (r *RefreshTokenRepository) CleanupOldTokens(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE (expires_at < $1)
			OR (revoked_at IS NOT NULL AND revoked_at < $1)`

	result, err := r.getDB(ctx).ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old refresh tokens: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// getDB returns the database connection, checking for transaction in context.
func (r *RefreshTokenRepository) getDB(ctx context.Context) sqlx.ExtContext {
	if tx := getTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}
