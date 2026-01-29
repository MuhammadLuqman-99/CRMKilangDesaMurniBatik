// Package postgres contains PostgreSQL repository implementations for the notification service.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kilang-desa-murni/crm/internal/notification/domain"
)

// ============================================================================
// Notification Repository Implementation
// ============================================================================

// NotificationRepository implements domain.NotificationRepository using PostgreSQL.
type NotificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository creates a new NotificationRepository instance.
func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// notificationRow represents the database row structure for notifications.
type notificationRow struct {
	ID               uuid.UUID      `db:"id"`
	TenantID         uuid.UUID      `db:"tenant_id"`
	Code             string         `db:"code"`
	Type             string         `db:"type"`
	Channel          string         `db:"channel"`
	Priority         string         `db:"priority"`
	Status           string         `db:"status"`
	TemplateID       sql.NullString `db:"template_id"`
	TemplateName     sql.NullString `db:"template_name"`
	RecipientID      sql.NullString `db:"recipient_id"`
	RecipientEmail   sql.NullString `db:"recipient_email"`
	RecipientPhone   sql.NullString `db:"recipient_phone"`
	RecipientName    sql.NullString `db:"recipient_name"`
	DeviceToken      sql.NullString `db:"device_token"`
	Subject          sql.NullString `db:"subject"`
	Body             string         `db:"body"`
	HTMLBody         sql.NullString `db:"html_body"`
	Data             []byte         `db:"data"`
	Metadata         []byte         `db:"metadata"`
	FromAddress      sql.NullString `db:"from_address"`
	FromName         sql.NullString `db:"from_name"`
	ReplyTo          sql.NullString `db:"reply_to"`
	ScheduledAt      sql.NullTime   `db:"scheduled_at"`
	SentAt           sql.NullTime   `db:"sent_at"`
	DeliveredAt      sql.NullTime   `db:"delivered_at"`
	ReadAt           sql.NullTime   `db:"read_at"`
	FailedAt         sql.NullTime   `db:"failed_at"`
	CancelledAt      sql.NullTime   `db:"cancelled_at"`
	AttemptCount     int            `db:"attempt_count"`
	LastAttemptAt    sql.NullTime   `db:"last_attempt_at"`
	NextRetryAt      sql.NullTime   `db:"next_retry_at"`
	ErrorCode        sql.NullString `db:"error_code"`
	ErrorMessage     sql.NullString `db:"error_message"`
	ProviderError    sql.NullString `db:"provider_error"`
	Provider         sql.NullString `db:"provider"`
	ProviderMessageID sql.NullString `db:"provider_message_id"`
	TrackOpens       bool           `db:"track_opens"`
	TrackClicks      bool           `db:"track_clicks"`
	OpenCount        int            `db:"open_count"`
	ClickCount       int            `db:"click_count"`
	SourceEvent      sql.NullString `db:"source_event"`
	SourceEntityID   sql.NullString `db:"source_entity_id"`
	SourceEntityType sql.NullString `db:"source_entity_type"`
	CorrelationID    sql.NullString `db:"correlation_id"`
	BatchID          sql.NullString `db:"batch_id"`
	BatchIndex       int            `db:"batch_index"`
	CreatedBy        sql.NullString `db:"created_by"`
	UpdatedBy        sql.NullString `db:"updated_by"`
	Version          int            `db:"version"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	DeletedAt        sql.NullTime   `db:"deleted_at"`
}

// Create creates a new notification.
func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	executor := getExecutor(ctx, r.db)

	row := r.toRow(notification)
	query := `
		INSERT INTO notifications (
			id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36,
			$37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51
		)`

	_, err := executor.ExecContext(ctx, query,
		row.ID, row.TenantID, row.Code, row.Type, row.Channel, row.Priority, row.Status,
		row.TemplateID, row.TemplateName, row.RecipientID, row.RecipientEmail, row.RecipientPhone,
		row.RecipientName, row.DeviceToken, row.Subject, row.Body, row.HTMLBody, row.Data, row.Metadata,
		row.FromAddress, row.FromName, row.ReplyTo, row.ScheduledAt, row.SentAt, row.DeliveredAt,
		row.ReadAt, row.FailedAt, row.CancelledAt, row.AttemptCount, row.LastAttemptAt, row.NextRetryAt,
		row.ErrorCode, row.ErrorMessage, row.ProviderError, row.Provider, row.ProviderMessageID,
		row.TrackOpens, row.TrackClicks, row.OpenCount, row.ClickCount,
		row.SourceEvent, row.SourceEntityID, row.SourceEntityType, row.CorrelationID,
		row.BatchID, row.BatchIndex, row.CreatedBy, row.UpdatedBy, row.Version, row.CreatedAt, row.UpdatedAt,
	)
	if err != nil {
		if IsUniqueViolation(err) {
			return domain.ErrNotificationAlreadyExists
		}
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// Update updates an existing notification.
func (r *NotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	executor := getExecutor(ctx, r.db)

	row := r.toRow(notification)
	query := `
		UPDATE notifications SET
			status = $3, template_id = $4, template_name = $5,
			recipient_id = $6, recipient_email = $7, recipient_phone = $8,
			recipient_name = $9, device_token = $10, subject = $11, body = $12,
			html_body = $13, data = $14, metadata = $15,
			from_address = $16, from_name = $17, reply_to = $18,
			scheduled_at = $19, sent_at = $20, delivered_at = $21,
			read_at = $22, failed_at = $23, cancelled_at = $24,
			attempt_count = $25, last_attempt_at = $26, next_retry_at = $27,
			error_code = $28, error_message = $29, provider_error = $30,
			provider = $31, provider_message_id = $32,
			track_opens = $33, track_clicks = $34, open_count = $35, click_count = $36,
			source_event = $37, source_entity_id = $38, source_entity_type = $39,
			correlation_id = $40, batch_id = $41, batch_index = $42,
			updated_by = $43, version = version + 1, updated_at = $44
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`

	result, err := executor.ExecContext(ctx, query,
		row.ID, row.TenantID, row.Status, row.TemplateID, row.TemplateName,
		row.RecipientID, row.RecipientEmail, row.RecipientPhone,
		row.RecipientName, row.DeviceToken, row.Subject, row.Body,
		row.HTMLBody, row.Data, row.Metadata,
		row.FromAddress, row.FromName, row.ReplyTo,
		row.ScheduledAt, row.SentAt, row.DeliveredAt,
		row.ReadAt, row.FailedAt, row.CancelledAt,
		row.AttemptCount, row.LastAttemptAt, row.NextRetryAt,
		row.ErrorCode, row.ErrorMessage, row.ProviderError,
		row.Provider, row.ProviderMessageID,
		row.TrackOpens, row.TrackClicks, row.OpenCount, row.ClickCount,
		row.SourceEvent, row.SourceEntityID, row.SourceEntityType,
		row.CorrelationID, row.BatchID, row.BatchIndex,
		row.UpdatedBy, row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotificationNotFound
	}

	return nil
}

// Delete soft deletes a notification.
func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	executor := getExecutor(ctx, r.db)

	query := `UPDATE notifications SET deleted_at = $2, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := executor.ExecContext(ctx, query, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotificationNotFound
	}

	return nil
}

// HardDelete permanently deletes a notification.
func (r *NotificationRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	executor := getExecutor(ctx, r.db)

	query := `DELETE FROM notifications WHERE id = $1`
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete notification: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotificationNotFound
	}

	return nil
}

// FindByID finds a notification by ID.
func (r *NotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications WHERE id = $1 AND deleted_at IS NULL`

	var row notificationRow
	err := sqlx.GetContext(ctx, executor, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}

	return r.toEntity(&row), nil
}

// FindByCode finds a notification by code.
func (r *NotificationRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Notification, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications WHERE tenant_id = $1 AND code = $2 AND deleted_at IS NULL`

	var row notificationRow
	err := sqlx.GetContext(ctx, executor, &row, query, tenantID, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}

	return r.toEntity(&row), nil
}

// List lists notifications with filtering and pagination.
func (r *NotificationRepository) List(ctx context.Context, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	executor := getExecutor(ctx, r.db)

	baseQuery := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications`

	qb := NewQueryBuilder(baseQuery)
	r.applyFilters(qb, filter)

	// Get total count
	countQuery, countArgs := qb.BuildCount()
	var total int64
	if err := sqlx.GetContext(ctx, executor, &total, countQuery, countArgs...); err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Apply sorting and pagination
	sortColumn := ValidateSortColumn(filter.SortBy, notificationSortColumns)
	sortOrder := ValidateSortOrder(filter.SortOrder)
	qb.OrderBy(sortColumn, sortOrder)

	if filter.Limit > 0 {
		qb.Limit(filter.Limit)
	} else {
		qb.Limit(50)
	}
	qb.Offset(filter.Offset)

	query, args := qb.Build()

	var rows []notificationRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		notifications[i] = r.toEntity(&row)
	}

	return &domain.NotificationList{
		Notifications: notifications,
		Total:         total,
		Offset:        filter.Offset,
		Limit:         filter.Limit,
		HasMore:       int64(filter.Offset+len(notifications)) < total,
	}, nil
}

// FindByRecipient finds notifications for a recipient.
func (r *NotificationRepository) FindByRecipient(ctx context.Context, tenantID uuid.UUID, recipientID uuid.UUID, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	filter.TenantID = &tenantID
	filter.RecipientID = &recipientID
	return r.List(ctx, filter)
}

// FindByEmail finds notifications sent to an email address.
func (r *NotificationRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	filter.TenantID = &tenantID
	filter.RecipientEmail = email
	return r.List(ctx, filter)
}

// FindByStatus finds notifications by status.
func (r *NotificationRepository) FindByStatus(ctx context.Context, tenantID uuid.UUID, status domain.NotificationStatus, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	filter.TenantID = &tenantID
	filter.Statuses = []domain.NotificationStatus{status}
	return r.List(ctx, filter)
}

// FindPending finds pending notifications ready to send.
func (r *NotificationRepository) FindPending(ctx context.Context, limit int) ([]*domain.Notification, error) {
	executor := getExecutor(ctx, r.db)

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications
		WHERE status IN ('pending', 'queued') AND deleted_at IS NULL
		ORDER BY priority DESC, created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	var rows []notificationRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("failed to find pending notifications: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		notifications[i] = r.toEntity(&row)
	}

	return notifications, nil
}

// FindScheduled finds scheduled notifications that are due.
func (r *NotificationRepository) FindScheduled(ctx context.Context, before time.Time, limit int) ([]*domain.Notification, error) {
	executor := getExecutor(ctx, r.db)

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications
		WHERE status = 'scheduled' AND scheduled_at <= $1 AND deleted_at IS NULL
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED`

	var rows []notificationRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, before, limit); err != nil {
		return nil, fmt.Errorf("failed to find scheduled notifications: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		notifications[i] = r.toEntity(&row)
	}

	return notifications, nil
}

// FindRetryable finds failed notifications that can be retried.
func (r *NotificationRepository) FindRetryable(ctx context.Context, before time.Time, limit int) ([]*domain.Notification, error) {
	executor := getExecutor(ctx, r.db)

	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications
		WHERE status = 'retrying' AND next_retry_at <= $1 AND deleted_at IS NULL
		ORDER BY priority DESC, next_retry_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED`

	var rows []notificationRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, before, limit); err != nil {
		return nil, fmt.Errorf("failed to find retryable notifications: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		notifications[i] = r.toEntity(&row)
	}

	return notifications, nil
}

// FindByBatch finds notifications in a batch.
func (r *NotificationRepository) FindByBatch(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications
		WHERE batch_id = $1 AND deleted_at IS NULL
		ORDER BY batch_index ASC`

	var rows []notificationRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, batchID); err != nil {
		return nil, fmt.Errorf("failed to find batch notifications: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		notifications[i] = r.toEntity(&row)
	}

	return notifications, nil
}

// FindBySourceEvent finds notifications triggered by a source event.
func (r *NotificationRepository) FindBySourceEvent(ctx context.Context, tenantID uuid.UUID, sourceEvent string, sourceEntityID uuid.UUID) ([]*domain.Notification, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications
		WHERE tenant_id = $1 AND source_event = $2 AND source_entity_id = $3 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	var rows []notificationRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, tenantID, sourceEvent, sourceEntityID); err != nil {
		return nil, fmt.Errorf("failed to find notifications by source event: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		notifications[i] = r.toEntity(&row)
	}

	return notifications, nil
}

// CountByTenant counts notifications for a tenant.
func (r *NotificationRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	executor := getExecutor(ctx, r.db)

	query := `SELECT COUNT(*) FROM notifications WHERE tenant_id = $1 AND deleted_at IS NULL`
	var count int64
	if err := sqlx.GetContext(ctx, executor, &count, query, tenantID); err != nil {
		return 0, fmt.Errorf("failed to count notifications: %w", err)
	}
	return count, nil
}

// CountByStatus counts notifications by status for a tenant.
func (r *NotificationRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.NotificationStatus]int64, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT status, COUNT(*) as count
		FROM notifications
		WHERE tenant_id = $1 AND deleted_at IS NULL
		GROUP BY status`

	type statusCount struct {
		Status string `db:"status"`
		Count  int64  `db:"count"`
	}

	var counts []statusCount
	if err := sqlx.SelectContext(ctx, executor, &counts, query, tenantID); err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}

	result := make(map[domain.NotificationStatus]int64)
	for _, sc := range counts {
		result[domain.NotificationStatus(sc.Status)] = sc.Count
	}
	return result, nil
}

// CountByChannel counts notifications by channel for a tenant.
func (r *NotificationRepository) CountByChannel(ctx context.Context, tenantID uuid.UUID) (map[domain.NotificationChannel]int64, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT channel, COUNT(*) as count
		FROM notifications
		WHERE tenant_id = $1 AND deleted_at IS NULL
		GROUP BY channel`

	type channelCount struct {
		Channel string `db:"channel"`
		Count   int64  `db:"count"`
	}

	var counts []channelCount
	if err := sqlx.SelectContext(ctx, executor, &counts, query, tenantID); err != nil {
		return nil, fmt.Errorf("failed to count by channel: %w", err)
	}

	result := make(map[domain.NotificationChannel]int64)
	for _, cc := range counts {
		result[domain.NotificationChannel(cc.Channel)] = cc.Count
	}
	return result, nil
}

// GetStats gets notification statistics.
func (r *NotificationRepository) GetStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*domain.NotificationStats, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT
			COUNT(*) as total_count,
			COALESCE(SUM(CASE WHEN status IN ('sent', 'delivered', 'read') THEN 1 ELSE 0 END), 0) as sent_count,
			COALESCE(SUM(CASE WHEN status IN ('delivered', 'read') THEN 1 ELSE 0 END), 0) as delivered_count,
			COALESCE(SUM(CASE WHEN status IN ('failed', 'bounced', 'complained') THEN 1 ELSE 0 END), 0) as failed_count,
			COALESCE(SUM(CASE WHEN status IN ('pending', 'queued', 'scheduled') THEN 1 ELSE 0 END), 0) as pending_count,
			COALESCE(SUM(open_count), 0) as total_opens,
			COALESCE(SUM(click_count), 0) as total_clicks,
			COALESCE(SUM(CASE WHEN status = 'bounced' THEN 1 ELSE 0 END), 0) as bounced_count,
			COALESCE(SUM(CASE WHEN status = 'complained' THEN 1 ELSE 0 END), 0) as complained_count,
			COALESCE(AVG(CASE WHEN delivered_at IS NOT NULL AND sent_at IS NOT NULL
				THEN EXTRACT(EPOCH FROM (delivered_at - sent_at)) ELSE NULL END), 0) as avg_delivery_time
		FROM notifications
		WHERE tenant_id = $1 AND created_at >= $2 AND created_at <= $3 AND deleted_at IS NULL`

	type statsResult struct {
		TotalCount      int64   `db:"total_count"`
		SentCount       int64   `db:"sent_count"`
		DeliveredCount  int64   `db:"delivered_count"`
		FailedCount     int64   `db:"failed_count"`
		PendingCount    int64   `db:"pending_count"`
		TotalOpens      int64   `db:"total_opens"`
		TotalClicks     int64   `db:"total_clicks"`
		BouncedCount    int64   `db:"bounced_count"`
		ComplainedCount int64   `db:"complained_count"`
		AvgDeliveryTime float64 `db:"avg_delivery_time"`
	}

	var result statsResult
	if err := sqlx.GetContext(ctx, executor, &result, query, tenantID, startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Calculate rates
	var openRate, clickRate, bounceRate, complaintRate float64
	if result.DeliveredCount > 0 {
		openRate = float64(result.TotalOpens) / float64(result.DeliveredCount) * 100
		clickRate = float64(result.TotalClicks) / float64(result.DeliveredCount) * 100
	}
	if result.SentCount > 0 {
		bounceRate = float64(result.BouncedCount) / float64(result.SentCount) * 100
		complaintRate = float64(result.ComplainedCount) / float64(result.SentCount) * 100
	}

	// Get counts by channel and status
	byChannel, _ := r.CountByChannel(ctx, tenantID)
	byStatus, _ := r.CountByStatus(ctx, tenantID)

	return &domain.NotificationStats{
		TotalCount:          result.TotalCount,
		SentCount:           result.SentCount,
		DeliveredCount:      result.DeliveredCount,
		FailedCount:         result.FailedCount,
		PendingCount:        result.PendingCount,
		ByChannel:           byChannel,
		ByStatus:            byStatus,
		OpenRate:            openRate,
		ClickRate:           clickRate,
		BounceRate:          bounceRate,
		ComplaintRate:       complaintRate,
		AverageDeliveryTime: result.AvgDeliveryTime,
		StartDate:           startDate,
		EndDate:             endDate,
	}, nil
}

// BulkCreate creates multiple notifications.
func (r *NotificationRepository) BulkCreate(ctx context.Context, notifications []*domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	executor := getExecutor(ctx, r.db)

	for _, notification := range notifications {
		row := r.toRow(notification)
		query := `
			INSERT INTO notifications (
				id, tenant_id, code, type, channel, priority, status,
				template_id, template_name, recipient_id, recipient_email, recipient_phone,
				recipient_name, device_token, subject, body, html_body, data, metadata,
				from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
				read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
				error_code, error_message, provider_error, provider, provider_message_id,
				track_opens, track_clicks, open_count, click_count,
				source_event, source_entity_id, source_entity_type, correlation_id,
				batch_id, batch_index, created_by, updated_by, version, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
				$20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36,
				$37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51
			)`

		_, err := executor.ExecContext(ctx, query,
			row.ID, row.TenantID, row.Code, row.Type, row.Channel, row.Priority, row.Status,
			row.TemplateID, row.TemplateName, row.RecipientID, row.RecipientEmail, row.RecipientPhone,
			row.RecipientName, row.DeviceToken, row.Subject, row.Body, row.HTMLBody, row.Data, row.Metadata,
			row.FromAddress, row.FromName, row.ReplyTo, row.ScheduledAt, row.SentAt, row.DeliveredAt,
			row.ReadAt, row.FailedAt, row.CancelledAt, row.AttemptCount, row.LastAttemptAt, row.NextRetryAt,
			row.ErrorCode, row.ErrorMessage, row.ProviderError, row.Provider, row.ProviderMessageID,
			row.TrackOpens, row.TrackClicks, row.OpenCount, row.ClickCount,
			row.SourceEvent, row.SourceEntityID, row.SourceEntityType, row.CorrelationID,
			row.BatchID, row.BatchIndex, row.CreatedBy, row.UpdatedBy, row.Version, row.CreatedAt, row.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to bulk create notification %s: %w", notification.ID, err)
		}
	}

	return nil
}

// BulkUpdateStatus updates status for multiple notifications.
func (r *NotificationRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.NotificationStatus) error {
	if len(ids) == 0 {
		return nil
	}

	executor := getExecutor(ctx, r.db)
	now := time.Now().UTC()

	query := `UPDATE notifications SET status = $1, updated_at = $2 WHERE id = ANY($3)`
	_, err := executor.ExecContext(ctx, query, string(status), now, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to bulk update status: %w", err)
	}

	return nil
}

// Exists checks if a notification exists.
func (r *NotificationRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	executor := getExecutor(ctx, r.db)

	query := `SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1 AND deleted_at IS NULL)`
	var exists bool
	if err := sqlx.GetContext(ctx, executor, &exists, query, id); err != nil {
		return false, fmt.Errorf("failed to check notification exists: %w", err)
	}
	return exists, nil
}

// GetVersion gets the current version for optimistic locking.
func (r *NotificationRepository) GetVersion(ctx context.Context, id uuid.UUID) (int, error) {
	executor := getExecutor(ctx, r.db)

	query := `SELECT version FROM notifications WHERE id = $1 AND deleted_at IS NULL`
	var version int
	if err := sqlx.GetContext(ctx, executor, &version, query, id); err != nil {
		if err == sql.ErrNoRows {
			return 0, domain.ErrNotificationNotFound
		}
		return 0, fmt.Errorf("failed to get version: %w", err)
	}
	return version, nil
}

// FindUnread finds unread in-app notifications for a user.
func (r *NotificationRepository) FindUnread(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*domain.Notification, error) {
	executor := getExecutor(ctx, r.db)

	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, tenant_id, code, type, channel, priority, status,
			template_id, template_name, recipient_id, recipient_email, recipient_phone,
			recipient_name, device_token, subject, body, html_body, data, metadata,
			from_address, from_name, reply_to, scheduled_at, sent_at, delivered_at,
			read_at, failed_at, cancelled_at, attempt_count, last_attempt_at, next_retry_at,
			error_code, error_message, provider_error, provider, provider_message_id,
			track_opens, track_clicks, open_count, click_count,
			source_event, source_entity_id, source_entity_type, correlation_id,
			batch_id, batch_index, created_by, updated_by, version, created_at, updated_at, deleted_at
		FROM notifications
		WHERE tenant_id = $1 AND recipient_id = $2 AND channel = 'in_app'
			AND status = 'delivered' AND read_at IS NULL AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3`

	var rows []notificationRow
	if err := sqlx.SelectContext(ctx, executor, &rows, query, tenantID, userID, limit); err != nil {
		return nil, fmt.Errorf("failed to find unread notifications: %w", err)
	}

	notifications := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		notifications[i] = r.toEntity(&row)
	}

	return notifications, nil
}

// CountUnread counts unread in-app notifications for a user.
func (r *NotificationRepository) CountUnread(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	executor := getExecutor(ctx, r.db)

	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE tenant_id = $1 AND recipient_id = $2 AND channel = 'in_app'
			AND status = 'delivered' AND read_at IS NULL AND deleted_at IS NULL`

	var count int64
	if err := sqlx.GetContext(ctx, executor, &count, query, tenantID, userID); err != nil {
		return 0, fmt.Errorf("failed to count unread: %w", err)
	}
	return count, nil
}

// MarkAllRead marks all in-app notifications as read for a user.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	executor := getExecutor(ctx, r.db)
	now := time.Now().UTC()

	query := `
		UPDATE notifications SET
			status = 'read', read_at = $3, updated_at = $3
		WHERE tenant_id = $1 AND recipient_id = $2 AND channel = 'in_app'
			AND status = 'delivered' AND read_at IS NULL AND deleted_at IS NULL`

	_, err := executor.ExecContext(ctx, query, tenantID, userID, now)
	if err != nil {
		return fmt.Errorf("failed to mark all read: %w", err)
	}
	return nil
}

// DeleteOld deletes notifications older than a specified date.
func (r *NotificationRepository) DeleteOld(ctx context.Context, before time.Time) (int64, error) {
	executor := getExecutor(ctx, r.db)

	query := `DELETE FROM notifications WHERE created_at < $1`
	result, err := executor.ExecContext(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old notifications: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// notificationSortColumns maps filter sort fields to database columns.
var notificationSortColumns = map[string]string{
	"created_at":   "created_at",
	"updated_at":   "updated_at",
	"sent_at":      "sent_at",
	"scheduled_at": "scheduled_at",
	"priority":     "priority",
	"status":       "status",
}

// applyFilters applies filters to the query builder.
func (r *NotificationRepository) applyFilters(qb *QueryBuilder, filter domain.NotificationFilter) {
	// Always exclude deleted unless specified
	if !filter.IncludeDeleted {
		qb.Where(fmt.Sprintf("deleted_at IS NULL"))
	}

	if filter.TenantID != nil {
		qb.Where(fmt.Sprintf("tenant_id = $%d", qb.NextParam()), *filter.TenantID)
	}

	if len(filter.IDs) > 0 {
		qb.WhereIn("id", filter.IDs)
	}

	if len(filter.Codes) > 0 {
		qb.WhereInStrings("code", filter.Codes)
	}

	if len(filter.Channels) > 0 {
		channels := make([]string, len(filter.Channels))
		for i, c := range filter.Channels {
			channels[i] = string(c)
		}
		qb.WhereInStrings("channel", channels)
	}

	if len(filter.Types) > 0 {
		types := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			types[i] = string(t)
		}
		qb.WhereInStrings("type", types)
	}

	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, s := range filter.Statuses {
			statuses[i] = string(s)
		}
		qb.WhereInStrings("status", statuses)
	}

	if filter.RecipientID != nil {
		qb.Where(fmt.Sprintf("recipient_id = $%d", qb.NextParam()), *filter.RecipientID)
	}

	if filter.RecipientEmail != "" {
		qb.Where(fmt.Sprintf("recipient_email = $%d", qb.NextParam()), filter.RecipientEmail)
	}

	if filter.RecipientPhone != "" {
		qb.Where(fmt.Sprintf("recipient_phone = $%d", qb.NextParam()), filter.RecipientPhone)
	}

	if filter.TemplateID != nil {
		qb.Where(fmt.Sprintf("template_id = $%d", qb.NextParam()), *filter.TemplateID)
	}

	if filter.BatchID != nil {
		qb.Where(fmt.Sprintf("batch_id = $%d", qb.NextParam()), *filter.BatchID)
	}

	if filter.SourceEvent != "" {
		qb.Where(fmt.Sprintf("source_event = $%d", qb.NextParam()), filter.SourceEvent)
	}

	if filter.SourceEntityID != nil {
		qb.Where(fmt.Sprintf("source_entity_id = $%d", qb.NextParam()), *filter.SourceEntityID)
	}

	if filter.CreatedAfter != nil {
		qb.Where(fmt.Sprintf("created_at >= $%d", qb.NextParam()), *filter.CreatedAfter)
	}

	if filter.CreatedBefore != nil {
		qb.Where(fmt.Sprintf("created_at <= $%d", qb.NextParam()), *filter.CreatedBefore)
	}

	if filter.SentAfter != nil {
		qb.Where(fmt.Sprintf("sent_at >= $%d", qb.NextParam()), *filter.SentAfter)
	}

	if filter.SentBefore != nil {
		qb.Where(fmt.Sprintf("sent_at <= $%d", qb.NextParam()), *filter.SentBefore)
	}
}

// toRow converts a domain notification to a database row.
func (r *NotificationRepository) toRow(n *domain.Notification) notificationRow {
	row := notificationRow{
		ID:           n.ID,
		TenantID:     n.TenantID,
		Code:         n.Code,
		Type:         string(n.Type),
		Channel:      string(n.Channel),
		Priority:     string(n.Priority),
		Status:       string(n.Status),
		Body:         n.Body,
		AttemptCount: n.AttemptCount,
		TrackOpens:   n.TrackOpens,
		TrackClicks:  n.TrackClicks,
		OpenCount:    n.OpenCount,
		ClickCount:   n.ClickCount,
		BatchIndex:   n.BatchIndex,
		Version:      n.Version(),
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
	}

	// Optional fields
	if n.TemplateID != nil {
		row.TemplateID = sql.NullString{String: n.TemplateID.String(), Valid: true}
	}
	if n.TemplateName != "" {
		row.TemplateName = sql.NullString{String: n.TemplateName, Valid: true}
	}
	if n.RecipientID != nil {
		row.RecipientID = sql.NullString{String: n.RecipientID.String(), Valid: true}
	}
	if n.RecipientEmail != "" {
		row.RecipientEmail = sql.NullString{String: n.RecipientEmail, Valid: true}
	}
	if n.RecipientPhone != "" {
		row.RecipientPhone = sql.NullString{String: n.RecipientPhone, Valid: true}
	}
	if n.RecipientName != "" {
		row.RecipientName = sql.NullString{String: n.RecipientName, Valid: true}
	}
	if n.DeviceToken != "" {
		row.DeviceToken = sql.NullString{String: n.DeviceToken, Valid: true}
	}
	if n.Subject != "" {
		row.Subject = sql.NullString{String: n.Subject, Valid: true}
	}
	if n.HTMLBody != "" {
		row.HTMLBody = sql.NullString{String: n.HTMLBody, Valid: true}
	}
	if n.Data != nil {
		row.Data, _ = json.Marshal(n.Data)
	}
	if n.Metadata != nil {
		row.Metadata, _ = json.Marshal(n.Metadata)
	}
	if n.FromAddress != "" {
		row.FromAddress = sql.NullString{String: n.FromAddress, Valid: true}
	}
	if n.FromName != "" {
		row.FromName = sql.NullString{String: n.FromName, Valid: true}
	}
	if n.ReplyTo != "" {
		row.ReplyTo = sql.NullString{String: n.ReplyTo, Valid: true}
	}
	if n.ScheduledAt != nil {
		row.ScheduledAt = sql.NullTime{Time: *n.ScheduledAt, Valid: true}
	}
	if n.SentAt != nil {
		row.SentAt = sql.NullTime{Time: *n.SentAt, Valid: true}
	}
	if n.DeliveredAt != nil {
		row.DeliveredAt = sql.NullTime{Time: *n.DeliveredAt, Valid: true}
	}
	if n.ReadAt != nil {
		row.ReadAt = sql.NullTime{Time: *n.ReadAt, Valid: true}
	}
	if n.FailedAt != nil {
		row.FailedAt = sql.NullTime{Time: *n.FailedAt, Valid: true}
	}
	if n.CancelledAt != nil {
		row.CancelledAt = sql.NullTime{Time: *n.CancelledAt, Valid: true}
	}
	if n.LastAttemptAt != nil {
		row.LastAttemptAt = sql.NullTime{Time: *n.LastAttemptAt, Valid: true}
	}
	if n.NextRetryAt != nil {
		row.NextRetryAt = sql.NullTime{Time: *n.NextRetryAt, Valid: true}
	}
	if n.ErrorCode != "" {
		row.ErrorCode = sql.NullString{String: n.ErrorCode, Valid: true}
	}
	if n.ErrorMessage != "" {
		row.ErrorMessage = sql.NullString{String: n.ErrorMessage, Valid: true}
	}
	if n.ProviderError != "" {
		row.ProviderError = sql.NullString{String: n.ProviderError, Valid: true}
	}
	if n.Provider != "" {
		row.Provider = sql.NullString{String: n.Provider, Valid: true}
	}
	if n.ProviderMessageID != "" {
		row.ProviderMessageID = sql.NullString{String: n.ProviderMessageID, Valid: true}
	}
	if n.SourceEvent != "" {
		row.SourceEvent = sql.NullString{String: n.SourceEvent, Valid: true}
	}
	if n.SourceEntityID != nil {
		row.SourceEntityID = sql.NullString{String: n.SourceEntityID.String(), Valid: true}
	}
	if n.SourceEntityType != "" {
		row.SourceEntityType = sql.NullString{String: n.SourceEntityType, Valid: true}
	}
	if n.CorrelationID != "" {
		row.CorrelationID = sql.NullString{String: n.CorrelationID, Valid: true}
	}
	if n.BatchID != nil {
		row.BatchID = sql.NullString{String: n.BatchID.String(), Valid: true}
	}
	if n.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: n.CreatedBy.String(), Valid: true}
	}
	if n.UpdatedBy != nil {
		row.UpdatedBy = sql.NullString{String: n.UpdatedBy.String(), Valid: true}
	}

	return row
}

// toEntity converts a database row to a domain notification.
func (r *NotificationRepository) toEntity(row *notificationRow) *domain.Notification {
	n := &domain.Notification{
		TenantID:     row.TenantID,
		Code:         row.Code,
		Type:         domain.NotificationType(row.Type),
		Channel:      domain.NotificationChannel(row.Channel),
		Priority:     domain.NotificationPriority(row.Priority),
		Status:       domain.NotificationStatus(row.Status),
		Body:         row.Body,
		AttemptCount: row.AttemptCount,
		TrackOpens:   row.TrackOpens,
		TrackClicks:  row.TrackClicks,
		OpenCount:    row.OpenCount,
		ClickCount:   row.ClickCount,
		BatchIndex:   row.BatchIndex,
	}

	// Set base aggregate root fields
	n.BaseAggregateRoot = domain.BaseAggregateRoot{}
	n.ID = row.ID
	n.SetCreatedAt(row.CreatedAt)
	n.SetUpdatedAt(row.UpdatedAt)
	n.SetVersion(row.Version)

	// Optional fields
	if row.TemplateID.Valid {
		id, _ := uuid.Parse(row.TemplateID.String)
		n.TemplateID = &id
	}
	if row.TemplateName.Valid {
		n.TemplateName = row.TemplateName.String
	}
	if row.RecipientID.Valid {
		id, _ := uuid.Parse(row.RecipientID.String)
		n.RecipientID = &id
	}
	if row.RecipientEmail.Valid {
		n.RecipientEmail = row.RecipientEmail.String
	}
	if row.RecipientPhone.Valid {
		n.RecipientPhone = row.RecipientPhone.String
	}
	if row.RecipientName.Valid {
		n.RecipientName = row.RecipientName.String
	}
	if row.DeviceToken.Valid {
		n.DeviceToken = row.DeviceToken.String
	}
	if row.Subject.Valid {
		n.Subject = row.Subject.String
	}
	if row.HTMLBody.Valid {
		n.HTMLBody = row.HTMLBody.String
	}
	if row.Data != nil && len(row.Data) > 0 {
		json.Unmarshal(row.Data, &n.Data)
	}
	if row.Metadata != nil && len(row.Metadata) > 0 {
		json.Unmarshal(row.Metadata, &n.Metadata)
	}
	if row.FromAddress.Valid {
		n.FromAddress = row.FromAddress.String
	}
	if row.FromName.Valid {
		n.FromName = row.FromName.String
	}
	if row.ReplyTo.Valid {
		n.ReplyTo = row.ReplyTo.String
	}
	if row.ScheduledAt.Valid {
		n.ScheduledAt = &row.ScheduledAt.Time
	}
	if row.SentAt.Valid {
		n.SentAt = &row.SentAt.Time
	}
	if row.DeliveredAt.Valid {
		n.DeliveredAt = &row.DeliveredAt.Time
	}
	if row.ReadAt.Valid {
		n.ReadAt = &row.ReadAt.Time
	}
	if row.FailedAt.Valid {
		n.FailedAt = &row.FailedAt.Time
	}
	if row.CancelledAt.Valid {
		n.CancelledAt = &row.CancelledAt.Time
	}
	if row.LastAttemptAt.Valid {
		n.LastAttemptAt = &row.LastAttemptAt.Time
	}
	if row.NextRetryAt.Valid {
		n.NextRetryAt = &row.NextRetryAt.Time
	}
	if row.ErrorCode.Valid {
		n.ErrorCode = row.ErrorCode.String
	}
	if row.ErrorMessage.Valid {
		n.ErrorMessage = row.ErrorMessage.String
	}
	if row.ProviderError.Valid {
		n.ProviderError = row.ProviderError.String
	}
	if row.Provider.Valid {
		n.Provider = row.Provider.String
	}
	if row.ProviderMessageID.Valid {
		n.ProviderMessageID = row.ProviderMessageID.String
	}
	if row.SourceEvent.Valid {
		n.SourceEvent = row.SourceEvent.String
	}
	if row.SourceEntityID.Valid {
		id, _ := uuid.Parse(row.SourceEntityID.String)
		n.SourceEntityID = &id
	}
	if row.SourceEntityType.Valid {
		n.SourceEntityType = row.SourceEntityType.String
	}
	if row.CorrelationID.Valid {
		n.CorrelationID = row.CorrelationID.String
	}
	if row.BatchID.Valid {
		id, _ := uuid.Parse(row.BatchID.String)
		n.BatchID = &id
	}
	if row.CreatedBy.Valid {
		id, _ := uuid.Parse(row.CreatedBy.String)
		n.CreatedBy = &id
	}
	if row.UpdatedBy.Valid {
		id, _ := uuid.Parse(row.UpdatedBy.String)
		n.UpdatedBy = &id
	}

	return n
}

// Ensure NotificationRepository implements the domain interface
var _ domain.NotificationRepository = (*NotificationRepository)(nil)
