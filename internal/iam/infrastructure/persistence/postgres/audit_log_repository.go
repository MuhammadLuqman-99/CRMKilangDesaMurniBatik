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

// AuditLogRow represents an audit log database row.
type AuditLogRow struct {
	ID         uuid.UUID       `db:"id"`
	TenantID   uuid.UUID       `db:"tenant_id"`
	UserID     *uuid.UUID      `db:"user_id"`
	Action     string          `db:"action"`
	EntityType string          `db:"entity_type"`
	EntityID   *uuid.UUID      `db:"entity_id"`
	OldValues  json.RawMessage `db:"old_values"`
	NewValues  json.RawMessage `db:"new_values"`
	IPAddress  sql.NullString  `db:"ip_address"`
	UserAgent  sql.NullString  `db:"user_agent"`
	CreatedAt  time.Time       `db:"created_at"`
}

// ToEntry converts an AuditLogRow to an AuditLogEntry.
func (r *AuditLogRow) ToEntry() *domain.AuditLogEntry {
	var oldValues map[string]interface{}
	if len(r.OldValues) > 0 && string(r.OldValues) != "null" {
		_ = json.Unmarshal(r.OldValues, &oldValues)
	}

	var newValues map[string]interface{}
	if len(r.NewValues) > 0 && string(r.NewValues) != "null" {
		_ = json.Unmarshal(r.NewValues, &newValues)
	}

	return &domain.AuditLogEntry{
		ID:         r.ID,
		TenantID:   r.TenantID,
		UserID:     r.UserID,
		Action:     r.Action,
		EntityType: r.EntityType,
		EntityID:   r.EntityID,
		OldValues:  oldValues,
		NewValues:  newValues,
		IPAddress:  r.IPAddress.String,
		UserAgent:  r.UserAgent.String,
		CreatedAt:  r.CreatedAt,
	}
}

// AuditLogRepository implements domain.AuditLogRepository using PostgreSQL.
type AuditLogRepository struct {
	db *sqlx.DB
}

// NewAuditLogRepository creates a new AuditLogRepository.
func NewAuditLogRepository(db *sqlx.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry.
func (r *AuditLogRepository) Create(ctx context.Context, entry *domain.AuditLogEntry) error {
	oldValues, err := json.Marshal(entry.OldValues)
	if err != nil {
		oldValues = []byte("null")
	}

	newValues, err := json.Marshal(entry.NewValues)
	if err != nil {
		newValues = []byte("null")
	}

	// Ensure ID is set
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}

	// Ensure CreatedAt is set
	createdAt, ok := entry.CreatedAt.(time.Time)
	if !ok || createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	query := `
		INSERT INTO audit_logs (id, tenant_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = r.getDB(ctx).ExecContext(ctx, query,
		entry.ID,
		entry.TenantID,
		entry.UserID,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
		oldValues,
		newValues,
		nullString(entry.IPAddress),
		nullString(entry.UserAgent),
		createdAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// FindByTenant finds audit logs for a tenant with pagination.
func (r *AuditLogRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, opts domain.AuditLogQueryOptions) ([]*domain.AuditLogEntry, int64, error) {
	// Build WHERE clause
	where := []string{"tenant_id = $1"}
	args := []interface{}{tenantID}
	argIndex := 2

	if opts.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, opts.Action)
		argIndex++
	}

	if opts.EntityType != "" {
		where = append(where, fmt.Sprintf("entity_type = $%d", argIndex))
		args = append(args, opts.EntityType)
		argIndex++
	}

	if opts.StartDate != nil {
		if startDate, ok := opts.StartDate.(time.Time); ok && !startDate.IsZero() {
			where = append(where, fmt.Sprintf("created_at >= $%d", argIndex))
			args = append(args, startDate)
			argIndex++
		}
	}

	if opts.EndDate != nil {
		if endDate, ok := opts.EndDate.(time.Time); ok && !endDate.IsZero() {
			where = append(where, fmt.Sprintf("created_at <= $%d", argIndex))
			args = append(args, endDate)
			argIndex++
		}
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", whereClause)
	var total int64
	err := r.getDB(ctx).GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Build ORDER BY
	orderDir := "DESC"
	if strings.ToUpper(opts.SortDirection) == "ASC" {
		orderDir = "ASC"
	}

	// Calculate offset
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	// Query audit logs
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY created_at %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderDir, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	var rows []AuditLogRow
	err = r.getDB(ctx).SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find audit logs: %w", err)
	}

	entries := make([]*domain.AuditLogEntry, len(rows))
	for i, row := range rows {
		entries[i] = row.ToEntry()
	}

	return entries, total, nil
}

// FindByEntity finds audit logs for a specific entity.
func (r *AuditLogRepository) FindByEntity(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*domain.AuditLogEntry, error) {
	query := `
		SELECT id, tenant_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at DESC`

	var rows []AuditLogRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, tenantID, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to find audit logs by entity: %w", err)
	}

	entries := make([]*domain.AuditLogEntry, len(rows))
	for i, row := range rows {
		entries[i] = row.ToEntry()
	}

	return entries, nil
}

// FindByUser finds audit logs for a specific user.
func (r *AuditLogRepository) FindByUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, opts domain.AuditLogQueryOptions) ([]*domain.AuditLogEntry, int64, error) {
	// Build WHERE clause
	where := []string{"tenant_id = $1", "user_id = $2"}
	args := []interface{}{tenantID, userID}
	argIndex := 3

	if opts.Action != "" {
		where = append(where, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, opts.Action)
		argIndex++
	}

	if opts.EntityType != "" {
		where = append(where, fmt.Sprintf("entity_type = $%d", argIndex))
		args = append(args, opts.EntityType)
		argIndex++
	}

	if opts.StartDate != nil {
		if startDate, ok := opts.StartDate.(time.Time); ok && !startDate.IsZero() {
			where = append(where, fmt.Sprintf("created_at >= $%d", argIndex))
			args = append(args, startDate)
			argIndex++
		}
	}

	if opts.EndDate != nil {
		if endDate, ok := opts.EndDate.(time.Time); ok && !endDate.IsZero() {
			where = append(where, fmt.Sprintf("created_at <= $%d", argIndex))
			args = append(args, endDate)
			argIndex++
		}
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", whereClause)
	var total int64
	err := r.getDB(ctx).GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Build ORDER BY
	orderDir := "DESC"
	if strings.ToUpper(opts.SortDirection) == "ASC" {
		orderDir = "ASC"
	}

	// Calculate offset
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	// Query audit logs
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY created_at %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderDir, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	var rows []AuditLogRow
	err = r.getDB(ctx).SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find audit logs: %w", err)
	}

	entries := make([]*domain.AuditLogEntry, len(rows))
	for i, row := range rows {
		entries[i] = row.ToEntry()
	}

	return entries, total, nil
}

// DeleteOldEntries deletes audit log entries older than the specified time.
func (r *AuditLogRepository) DeleteOldEntries(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `DELETE FROM audit_logs WHERE created_at < $1`

	result, err := r.getDB(ctx).ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// CountByTenant counts audit logs for a tenant.
func (r *AuditLogRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE tenant_id = $1`

	var count int64
	err := r.getDB(ctx).GetContext(ctx, &count, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	return count, nil
}

// FindRecentByTenant finds recent audit logs for a tenant.
func (r *AuditLogRepository) FindRecentByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*domain.AuditLogEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, tenant_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	var rows []AuditLogRow
	err := r.getDB(ctx).SelectContext(ctx, &rows, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent audit logs: %w", err)
	}

	entries := make([]*domain.AuditLogEntry, len(rows))
	for i, row := range rows {
		entries[i] = row.ToEntry()
	}

	return entries, nil
}

// FindByAction finds audit logs by action type.
func (r *AuditLogRepository) FindByAction(ctx context.Context, tenantID uuid.UUID, action string, opts domain.AuditLogQueryOptions) ([]*domain.AuditLogEntry, int64, error) {
	opts.Action = action
	return r.FindByTenant(ctx, tenantID, opts)
}

// getDB returns the database connection, checking for transaction in context.
func (r *AuditLogRepository) getDB(ctx context.Context) sqlx.ExtContext {
	if tx := getTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}
