// Package mongodb provides MongoDB implementations for customer repositories.
package mongodb

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// sessionContextKey is the key used to store the MongoDB session in context.
type sessionContextKey struct{}

// UnitOfWork implements domain.UnitOfWork using MongoDB transactions.
type UnitOfWork struct {
	client             *mongo.Client
	db                 *mongo.Database
	customerRepo       *CustomerRepository
	contactRepo        *ContactRepository
	noteRepo           *NoteRepository
	activityRepo       *ActivityRepository
	segmentRepo        *SegmentRepository
	importRepo         *ImportRepository
	outboxRepo         *OutboxRepository
	mu                 sync.RWMutex
}

// NewUnitOfWork creates a new UnitOfWork.
func NewUnitOfWork(client *mongo.Client, db *mongo.Database) *UnitOfWork {
	return &UnitOfWork{
		client:       client,
		db:           db,
		customerRepo: NewCustomerRepository(db),
		contactRepo:  NewContactRepository(db),
		noteRepo:     NewNoteRepository(db),
		activityRepo: NewActivityRepository(db),
		segmentRepo:  NewSegmentRepository(db),
		importRepo:   NewImportRepository(db),
		outboxRepo:   NewOutboxRepository(db),
	}
}

// Begin begins a new transaction and returns a context with the session.
func (uow *UnitOfWork) Begin(ctx context.Context) (context.Context, error) {
	// Start a new session
	session, err := uow.client.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	// Start transaction
	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Store session in context
	txCtx := context.WithValue(ctx, sessionContextKey{}, session)

	// Use session context for MongoDB operations
	return mongo.NewSessionContext(txCtx, session), nil
}

// Commit commits the transaction.
func (uow *UnitOfWork) Commit(ctx context.Context) error {
	session := uow.getSession(ctx)
	if session == nil {
		return fmt.Errorf("no active transaction")
	}

	if err := session.CommitTransaction(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	session.EndSession(ctx)
	return nil
}

// Rollback rolls back the transaction.
func (uow *UnitOfWork) Rollback(ctx context.Context) error {
	session := uow.getSession(ctx)
	if session == nil {
		return nil // No active transaction, nothing to rollback
	}

	if err := session.AbortTransaction(ctx); err != nil {
		session.EndSession(ctx)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	session.EndSession(ctx)
	return nil
}

// Customers returns the customer repository.
func (uow *UnitOfWork) Customers() domain.CustomerRepository {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.customerRepo
}

// Contacts returns the contact repository.
func (uow *UnitOfWork) Contacts() domain.ContactRepository {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.contactRepo
}

// Notes returns the note repository.
func (uow *UnitOfWork) Notes() domain.NoteRepository {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.noteRepo
}

// Activities returns the activity repository.
func (uow *UnitOfWork) Activities() domain.ActivityRepository {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.activityRepo
}

// Segments returns the segment repository.
func (uow *UnitOfWork) Segments() domain.SegmentRepository {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.segmentRepo
}

// Outbox returns the outbox repository.
func (uow *UnitOfWork) Outbox() domain.OutboxRepository {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.outboxRepo
}

// Imports returns the import repository.
func (uow *UnitOfWork) Imports() domain.ImportRepository {
	uow.mu.RLock()
	defer uow.mu.RUnlock()
	return uow.importRepo
}

// getSession extracts the MongoDB session from the context.
func (uow *UnitOfWork) getSession(ctx context.Context) mongo.Session {
	session, _ := ctx.Value(sessionContextKey{}).(mongo.Session)
	return session
}

// Database returns the MongoDB database for direct access if needed.
func (uow *UnitOfWork) Database() *mongo.Database {
	return uow.db
}

// Client returns the MongoDB client for direct access if needed.
func (uow *UnitOfWork) Client() *mongo.Client {
	return uow.client
}

// WithTransaction executes a function within a transaction.
func (uow *UnitOfWork) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(txCtx); err != nil {
		_ = uow.Rollback(txCtx)
		return err
	}

	return uow.Commit(txCtx)
}

// Ping checks if the database connection is alive.
func (uow *UnitOfWork) Ping(ctx context.Context) error {
	return uow.client.Ping(ctx, nil)
}

// Close closes the database connection.
func (uow *UnitOfWork) Close(ctx context.Context) error {
	return uow.client.Disconnect(ctx)
}

// EnsureIndexes ensures all required indexes exist.
func (uow *UnitOfWork) EnsureIndexes(ctx context.Context) error {
	indexManager := NewIndexManager(uow.db)
	return indexManager.EnsureIndexes(ctx)
}
