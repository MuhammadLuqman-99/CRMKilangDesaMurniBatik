// Package postgres contains PostgreSQL repository implementations.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TxKey is the context key for database transactions.
type txKey struct{}

// getTxFromContext retrieves a transaction from context if present.
func getTxFromContext(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

// setTxToContext stores a transaction in context.
func setTxToContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TransactionManager manages database transactions.
type TransactionManager struct {
	db *sqlx.DB
}

// NewTransactionManager creates a new TransactionManager.
func NewTransactionManager(db *sqlx.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a transaction.
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if already in transaction
	if getTxFromContext(ctx) != nil {
		return fn(ctx)
	}

	tx, err := tm.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := setTxToContext(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// BeginTx starts a new transaction.
func (tm *TransactionManager) BeginTx(ctx context.Context) (context.Context, error) {
	tx, err := tm.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return setTxToContext(ctx, tx), nil
}

// CommitTx commits the current transaction.
func (tm *TransactionManager) CommitTx(ctx context.Context) error {
	tx := getTxFromContext(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction in context")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RollbackTx rolls back the current transaction.
func (tm *TransactionManager) RollbackTx(ctx context.Context) error {
	tx := getTxFromContext(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction in context")
	}

	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

// TenantContextKey is the context key for tenant ID.
type tenantContextKey struct{}

// SetTenantContext sets the tenant context for row-level security.
func SetTenantContext(ctx context.Context, db sqlx.ExtContext, tenantID string) error {
	_, err := db.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
	return err
}

// GetTenantFromContext retrieves the tenant ID from context.
func GetTenantFromContext(ctx context.Context) string {
	if tid, ok := ctx.Value(tenantContextKey{}).(string); ok {
		return tid
	}
	return ""
}

// WithTenantContext adds tenant ID to context.
func WithTenantContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenantID)
}
