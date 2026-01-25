// Package postgres contains PostgreSQL repository implementations.
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// UnitOfWork implements domain.UnitOfWork for PostgreSQL.
type UnitOfWork struct {
	db                     *sqlx.DB
	userRepo               *UserRepository
	roleRepo               *RoleRepository
	tenantRepo             *TenantRepository
	refreshTokenRepo       *RefreshTokenRepository
	auditLogRepo           *AuditLogRepository
	outboxRepo             *OutboxRepository
}

// NewUnitOfWork creates a new UnitOfWork.
func NewUnitOfWork(db *sqlx.DB) *UnitOfWork {
	return &UnitOfWork{
		db:                     db,
		userRepo:               NewUserRepository(db),
		roleRepo:               NewRoleRepository(db),
		tenantRepo:             NewTenantRepository(db),
		refreshTokenRepo:       NewRefreshTokenRepository(db),
		auditLogRepo:           NewAuditLogRepository(db),
		outboxRepo:             NewOutboxRepository(db),
	}
}

// Begin starts a new transaction and returns a context with the transaction.
func (uow *UnitOfWork) Begin(ctx context.Context) (context.Context, error) {
	// Check if already in transaction
	if getTxFromContext(ctx) != nil {
		return ctx, nil
	}

	tx, err := uow.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return setTxToContext(ctx, tx), nil
}

// Commit commits the current transaction.
func (uow *UnitOfWork) Commit(ctx context.Context) error {
	tx := getTxFromContext(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction in context")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back the current transaction.
func (uow *UnitOfWork) Rollback(ctx context.Context) error {
	tx := getTxFromContext(ctx)
	if tx == nil {
		return fmt.Errorf("no transaction in context")
	}

	if err := tx.Rollback(); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

// UserRepository returns the user repository.
func (uow *UnitOfWork) UserRepository() domain.UserRepository {
	return uow.userRepo
}

// RoleRepository returns the role repository.
func (uow *UnitOfWork) RoleRepository() domain.RoleRepository {
	return uow.roleRepo
}

// TenantRepository returns the tenant repository.
func (uow *UnitOfWork) TenantRepository() domain.TenantRepository {
	return uow.tenantRepo
}

// RefreshTokenRepository returns the refresh token repository.
func (uow *UnitOfWork) RefreshTokenRepository() domain.RefreshTokenRepository {
	return uow.refreshTokenRepo
}

// AuditLogRepository returns the audit log repository.
func (uow *UnitOfWork) AuditLogRepository() domain.AuditLogRepository {
	return uow.auditLogRepo
}

// OutboxRepository returns the outbox repository.
func (uow *UnitOfWork) OutboxRepository() domain.OutboxRepository {
	return uow.outboxRepo
}

// WithTransaction executes a function within a transaction.
func (uow *UnitOfWork) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if already in transaction
	if getTxFromContext(ctx) != nil {
		return fn(ctx)
	}

	tx, err := uow.db.BeginTxx(ctx, &sql.TxOptions{
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

// WithSerializableTransaction executes a function within a serializable transaction.
// This provides the highest isolation level and is useful for critical operations.
func (uow *UnitOfWork) WithSerializableTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := uow.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin serializable transaction: %w", err)
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
		return fmt.Errorf("failed to commit serializable transaction: %w", err)
	}

	return nil
}

// DB returns the underlying database connection for advanced use cases.
func (uow *UnitOfWork) DB() *sqlx.DB {
	return uow.db
}

// Ping checks if the database connection is alive.
func (uow *UnitOfWork) Ping(ctx context.Context) error {
	return uow.db.PingContext(ctx)
}

// Close closes the database connection.
func (uow *UnitOfWork) Close() error {
	return uow.db.Close()
}
