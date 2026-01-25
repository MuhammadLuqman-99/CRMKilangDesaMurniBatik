// Package audit provides audit logging infrastructure.
package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// PostgresAuditLogger implements ports.AuditLogger using PostgreSQL.
type PostgresAuditLogger struct {
	repo   domain.AuditLogRepository
	logger *slog.Logger
	async  bool
}

// AuditLoggerConfig holds configuration for the audit logger.
type AuditLoggerConfig struct {
	Async  bool
	Logger *slog.Logger
}

// NewPostgresAuditLogger creates a new PostgreSQL audit logger.
func NewPostgresAuditLogger(repo domain.AuditLogRepository, config AuditLoggerConfig) *PostgresAuditLogger {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &PostgresAuditLogger{
		repo:   repo,
		logger: logger,
		async:  config.Async,
	}
}

// Log logs an audit event.
func (l *PostgresAuditLogger) Log(ctx context.Context, entry ports.AuditEntry) error {
	dbEntry := &domain.AuditLogEntry{
		ID:         uuid.New(),
		TenantID:   entry.TenantID,
		UserID:     entry.UserID,
		Action:     entry.Action,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		OldValues:  entry.OldValues,
		NewValues:  entry.NewValues,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		CreatedAt:  time.Now().UTC(),
	}

	if l.async {
		go l.logAsync(dbEntry)
		return nil
	}

	return l.repo.Create(ctx, dbEntry)
}

func (l *PostgresAuditLogger) logAsync(entry *domain.AuditLogEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := l.repo.Create(ctx, entry); err != nil {
		l.logger.Error("failed to log audit entry",
			slog.String("action", entry.Action),
			slog.String("entity_type", entry.EntityType),
			slog.Any("error", err),
		)
	}
}

// LoggingAuditLogger wraps another audit logger and adds structured logging.
type LoggingAuditLogger struct {
	inner  ports.AuditLogger
	logger *slog.Logger
}

// NewLoggingAuditLogger creates a new logging audit logger.
func NewLoggingAuditLogger(inner ports.AuditLogger, logger *slog.Logger) *LoggingAuditLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggingAuditLogger{
		inner:  inner,
		logger: logger,
	}
}

// Log logs an audit event with additional structured logging.
func (l *LoggingAuditLogger) Log(ctx context.Context, entry ports.AuditEntry) error {
	// Log to structured logger
	l.logger.Info("audit event",
		slog.String("tenant_id", entry.TenantID.String()),
		slog.Any("user_id", entry.UserID),
		slog.String("action", entry.Action),
		slog.String("entity_type", entry.EntityType),
		slog.Any("entity_id", entry.EntityID),
		slog.String("ip_address", entry.IPAddress),
	)

	// Forward to inner logger
	if l.inner != nil {
		return l.inner.Log(ctx, entry)
	}

	return nil
}

// BufferedAuditLogger buffers audit entries and flushes them in batches.
type BufferedAuditLogger struct {
	repo      domain.AuditLogRepository
	buffer    chan *domain.AuditLogEntry
	batchSize int
	interval  time.Duration
	logger    *slog.Logger
	done      chan struct{}
}

// BufferedAuditLoggerConfig holds configuration for the buffered audit logger.
type BufferedAuditLoggerConfig struct {
	BufferSize    int
	BatchSize     int
	FlushInterval time.Duration
	Logger        *slog.Logger
}

// DefaultBufferedAuditLoggerConfig returns default configuration.
func DefaultBufferedAuditLoggerConfig() BufferedAuditLoggerConfig {
	return BufferedAuditLoggerConfig{
		BufferSize:    1000,
		BatchSize:     100,
		FlushInterval: 5 * time.Second,
		Logger:        slog.Default(),
	}
}

// NewBufferedAuditLogger creates a new buffered audit logger.
func NewBufferedAuditLogger(repo domain.AuditLogRepository, config BufferedAuditLoggerConfig) *BufferedAuditLogger {
	logger := &BufferedAuditLogger{
		repo:      repo,
		buffer:    make(chan *domain.AuditLogEntry, config.BufferSize),
		batchSize: config.BatchSize,
		interval:  config.FlushInterval,
		logger:    config.Logger,
		done:      make(chan struct{}),
	}

	go logger.run()

	return logger
}

// Log adds an audit entry to the buffer.
func (l *BufferedAuditLogger) Log(ctx context.Context, entry ports.AuditEntry) error {
	dbEntry := &domain.AuditLogEntry{
		ID:         uuid.New(),
		TenantID:   entry.TenantID,
		UserID:     entry.UserID,
		Action:     entry.Action,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		OldValues:  entry.OldValues,
		NewValues:  entry.NewValues,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		CreatedAt:  time.Now().UTC(),
	}

	select {
	case l.buffer <- dbEntry:
		return nil
	default:
		// Buffer full, log synchronously
		return l.repo.Create(ctx, dbEntry)
	}
}

// Close stops the buffered logger and flushes remaining entries.
func (l *BufferedAuditLogger) Close() error {
	close(l.done)
	return nil
}

func (l *BufferedAuditLogger) run() {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	var batch []*domain.AuditLogEntry

	for {
		select {
		case entry := <-l.buffer:
			batch = append(batch, entry)
			if len(batch) >= l.batchSize {
				l.flush(batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				l.flush(batch)
				batch = nil
			}

		case <-l.done:
			// Flush remaining entries
			for entry := range l.buffer {
				batch = append(batch, entry)
			}
			if len(batch) > 0 {
				l.flush(batch)
			}
			return
		}
	}
}

func (l *BufferedAuditLogger) flush(entries []*domain.AuditLogEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, entry := range entries {
		if err := l.repo.Create(ctx, entry); err != nil {
			l.logger.Error("failed to flush audit entry",
				slog.String("action", entry.Action),
				slog.Any("error", err),
			)
		}
	}
}

// AuditEntryBuilder helps build audit entries.
type AuditEntryBuilder struct {
	entry ports.AuditEntry
}

// NewAuditEntryBuilder creates a new audit entry builder.
func NewAuditEntryBuilder(tenantID uuid.UUID, action string) *AuditEntryBuilder {
	return &AuditEntryBuilder{
		entry: ports.AuditEntry{
			TenantID: tenantID,
			Action:   action,
		},
	}
}

// WithUserID sets the user ID.
func (b *AuditEntryBuilder) WithUserID(userID uuid.UUID) *AuditEntryBuilder {
	b.entry.UserID = &userID
	return b
}

// WithEntityType sets the entity type.
func (b *AuditEntryBuilder) WithEntityType(entityType string) *AuditEntryBuilder {
	b.entry.EntityType = entityType
	return b
}

// WithEntityID sets the entity ID.
func (b *AuditEntryBuilder) WithEntityID(entityID uuid.UUID) *AuditEntryBuilder {
	b.entry.EntityID = &entityID
	return b
}

// WithOldValues sets the old values.
func (b *AuditEntryBuilder) WithOldValues(oldValues map[string]interface{}) *AuditEntryBuilder {
	b.entry.OldValues = oldValues
	return b
}

// WithNewValues sets the new values.
func (b *AuditEntryBuilder) WithNewValues(newValues map[string]interface{}) *AuditEntryBuilder {
	b.entry.NewValues = newValues
	return b
}

// WithIPAddress sets the IP address.
func (b *AuditEntryBuilder) WithIPAddress(ipAddress string) *AuditEntryBuilder {
	b.entry.IPAddress = ipAddress
	return b
}

// WithUserAgent sets the user agent.
func (b *AuditEntryBuilder) WithUserAgent(userAgent string) *AuditEntryBuilder {
	b.entry.UserAgent = userAgent
	return b
}

// Build returns the built audit entry.
func (b *AuditEntryBuilder) Build() ports.AuditEntry {
	return b.entry
}

// CompareChanges compares two objects and returns the differences.
func CompareChanges(old, new interface{}) (oldValues, newValues map[string]interface{}) {
	oldMap := toMap(old)
	newMap := toMap(new)

	oldValues = make(map[string]interface{})
	newValues = make(map[string]interface{})

	for key, newVal := range newMap {
		oldVal, exists := oldMap[key]
		if !exists || !equal(oldVal, newVal) {
			if exists {
				oldValues[key] = oldVal
			}
			newValues[key] = newVal
		}
	}

	for key, oldVal := range oldMap {
		if _, exists := newMap[key]; !exists {
			oldValues[key] = oldVal
		}
	}

	return oldValues, newValues
}

func toMap(v interface{}) map[string]interface{} {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}

	return result
}

func equal(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}
