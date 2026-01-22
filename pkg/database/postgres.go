// Package database provides database connection utilities for the CRM application.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/kilang-desa-murni/crm/pkg/config"
	"github.com/kilang-desa-murni/crm/pkg/logger"
)

// PostgresDB wraps the sql.DB connection pool.
type PostgresDB struct {
	*sql.DB
	config *config.DatabaseConfig
	log    *logger.Logger
}

// NewPostgres creates a new PostgreSQL database connection.
func NewPostgres(cfg *config.DatabaseConfig, log *logger.Logger) (*PostgresDB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.DBName).
		Msg("Connected to PostgreSQL")

	return &PostgresDB{
		DB:     db,
		config: cfg,
		log:    log,
	}, nil
}

// Close closes the database connection.
func (db *PostgresDB) Close() error {
	db.log.Info().Msg("Closing PostgreSQL connection")
	return db.DB.Close()
}

// Health checks the database connection health.
func (db *PostgresDB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Transaction executes a function within a database transaction.
func (db *PostgresDB) Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionWithOptions executes a function within a transaction with custom options.
func (db *PostgresDB) TransactionWithOptions(ctx context.Context, opts *sql.TxOptions, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Stats returns database statistics.
func (db *PostgresDB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// SetTenantContext sets the tenant ID in the PostgreSQL session for row-level security.
func (db *PostgresDB) SetTenantContext(ctx context.Context, tenantID string) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf("SET app.tenant_id = '%s'", tenantID))
	return err
}

// SetTenantContextTx sets the tenant ID in a transaction for row-level security.
func SetTenantContextTx(ctx context.Context, tx *sql.Tx, tenantID string) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantID))
	return err
}
