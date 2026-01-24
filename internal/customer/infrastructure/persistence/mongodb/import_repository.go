// Package mongodb provides MongoDB implementations for customer repositories.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
)

const (
	importsCollection      = "customer_imports"
	importErrorsCollection = "customer_import_errors"
)

// ImportRepository implements domain.ImportRepository using MongoDB.
type ImportRepository struct {
	db              *mongo.Database
	collection      *mongo.Collection
	errorsCollection *mongo.Collection
}

// NewImportRepository creates a new ImportRepository.
func NewImportRepository(db *mongo.Database) *ImportRepository {
	return &ImportRepository{
		db:              db,
		collection:      db.Collection(importsCollection),
		errorsCollection: db.Collection(importErrorsCollection),
	}
}

// CreateImport creates a new import record.
func (r *ImportRepository) CreateImport(ctx context.Context, imp *domain.Import) error {
	imp.CreatedAt = time.Now().UTC()
	imp.UpdatedAt = imp.CreatedAt
	imp.Version = 1

	_, err := r.collection.InsertOne(ctx, imp)
	if err != nil {
		return fmt.Errorf("failed to create import: %w", err)
	}

	return nil
}

// UpdateImport updates an import record.
func (r *ImportRepository) UpdateImport(ctx context.Context, imp *domain.Import) error {
	imp.UpdatedAt = time.Now().UTC()
	previousVersion := imp.Version
	imp.Version++

	filter := bson.M{
		"_id":     imp.ID,
		"version": previousVersion,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, imp)
	if err != nil {
		return fmt.Errorf("failed to update import: %w", err)
	}

	if result.MatchedCount == 0 {
		var existing domain.Import
		err := r.collection.FindOne(ctx, bson.M{"_id": imp.ID}).Decode(&existing)
		if err == mongo.ErrNoDocuments {
			return domain.ErrImportNotFound
		}
		return domain.ErrVersionConflict
	}

	return nil
}

// FindImportByID finds an import by ID.
func (r *ImportRepository) FindImportByID(ctx context.Context, id uuid.UUID) (*domain.Import, error) {
	var imp domain.Import
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&imp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrImportNotFound
		}
		return nil, fmt.Errorf("failed to find import: %w", err)
	}

	return &imp, nil
}

// FindImportsByTenant finds imports for a tenant.
func (r *ImportRepository) FindImportsByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*domain.Import, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{"tenant_id": tenantID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find imports: %w", err)
	}
	defer cursor.Close(ctx)

	var imports []*domain.Import
	if err := cursor.All(ctx, &imports); err != nil {
		return nil, fmt.Errorf("failed to decode imports: %w", err)
	}

	return imports, nil
}

// CreateImportError creates an import error record.
func (r *ImportRepository) CreateImportError(ctx context.Context, err *domain.ImportError) error {
	err.CreatedAt = time.Now().UTC()
	err.UpdatedAt = err.CreatedAt
	err.Version = 1

	_, insertErr := r.errorsCollection.InsertOne(ctx, err)
	if insertErr != nil {
		return fmt.Errorf("failed to create import error: %w", insertErr)
	}

	return nil
}

// FindImportErrors finds errors for an import.
func (r *ImportRepository) FindImportErrors(ctx context.Context, importID uuid.UUID) ([]*domain.ImportError, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "row_number", Value: 1}})

	cursor, err := r.errorsCollection.Find(ctx, bson.M{"import_id": importID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find import errors: %w", err)
	}
	defer cursor.Close(ctx)

	var errors []*domain.ImportError
	if err := cursor.All(ctx, &errors); err != nil {
		return nil, fmt.Errorf("failed to decode import errors: %w", err)
	}

	return errors, nil
}

// BulkCreateImportErrors creates multiple import error records.
func (r *ImportRepository) BulkCreateImportErrors(ctx context.Context, errors []*domain.ImportError) error {
	if len(errors) == 0 {
		return nil
	}

	now := time.Now().UTC()
	docs := make([]interface{}, len(errors))
	for i, err := range errors {
		err.CreatedAt = now
		err.UpdatedAt = now
		err.Version = 1
		docs[i] = err
	}

	_, err := r.errorsCollection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to bulk create import errors: %w", err)
	}

	return nil
}

// DeleteImportErrors deletes all errors for an import.
func (r *ImportRepository) DeleteImportErrors(ctx context.Context, importID uuid.UUID) error {
	_, err := r.errorsCollection.DeleteMany(ctx, bson.M{"import_id": importID})
	if err != nil {
		return fmt.Errorf("failed to delete import errors: %w", err)
	}
	return nil
}

// CountImportErrors counts errors for an import.
func (r *ImportRepository) CountImportErrors(ctx context.Context, importID uuid.UUID) (int64, error) {
	count, err := r.errorsCollection.CountDocuments(ctx, bson.M{"import_id": importID})
	if err != nil {
		return 0, fmt.Errorf("failed to count import errors: %w", err)
	}
	return count, nil
}

// UpdateImportProgress updates the progress of an import.
func (r *ImportRepository) UpdateImportProgress(ctx context.Context, id uuid.UUID, processed, success, failed, duplicates int) error {
	update := bson.M{
		"$set": bson.M{
			"processed_rows": processed,
			"success_rows":   success,
			"failed_rows":    failed,
			"duplicate_rows": duplicates,
			"updated_at":     time.Now().UTC(),
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to update import progress: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrImportNotFound
	}

	return nil
}

// StartImport marks an import as processing.
func (r *ImportRepository) StartImport(ctx context.Context, id uuid.UUID, totalRows int) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":     domain.ImportStatusProcessing,
			"started_at": now,
			"total_rows": totalRows,
			"updated_at": now,
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to start import: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrImportNotFound
	}

	return nil
}

// CompleteImport marks an import as completed.
func (r *ImportRepository) CompleteImport(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":       domain.ImportStatusCompleted,
			"completed_at": now,
			"updated_at":   now,
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to complete import: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrImportNotFound
	}

	return nil
}

// FailImport marks an import as failed.
func (r *ImportRepository) FailImport(ctx context.Context, id uuid.UUID, errMsg string) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":        domain.ImportStatusFailed,
			"error_message": errMsg,
			"completed_at":  now,
			"updated_at":    now,
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to mark import as failed: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrImportNotFound
	}

	return nil
}

// CancelImport marks an import as cancelled.
func (r *ImportRepository) CancelImport(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"status":       domain.ImportStatusCancelled,
			"completed_at": now,
			"updated_at":   now,
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("failed to cancel import: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrImportNotFound
	}

	return nil
}

// FindPendingImports finds pending imports.
func (r *ImportRepository) FindPendingImports(ctx context.Context, limit int) ([]*domain.Import, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, bson.M{"status": domain.ImportStatusPending}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending imports: %w", err)
	}
	defer cursor.Close(ctx)

	var imports []*domain.Import
	if err := cursor.All(ctx, &imports); err != nil {
		return nil, fmt.Errorf("failed to decode imports: %w", err)
	}

	return imports, nil
}

// FindProcessingImports finds imports that are currently processing.
func (r *ImportRepository) FindProcessingImports(ctx context.Context) ([]*domain.Import, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": domain.ImportStatusProcessing})
	if err != nil {
		return nil, fmt.Errorf("failed to find processing imports: %w", err)
	}
	defer cursor.Close(ctx)

	var imports []*domain.Import
	if err := cursor.All(ctx, &imports); err != nil {
		return nil, fmt.Errorf("failed to decode imports: %w", err)
	}

	return imports, nil
}

// DeleteOldImports deletes old completed/failed imports.
func (r *ImportRepository) DeleteOldImports(ctx context.Context, before time.Time) error {
	filter := bson.M{
		"status": bson.M{"$in": []domain.ImportStatus{
			domain.ImportStatusCompleted,
			domain.ImportStatusFailed,
			domain.ImportStatusCancelled,
		}},
		"completed_at": bson.M{"$lt": before},
	}

	// First find the imports to delete their errors
	cursor, err := r.collection.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return fmt.Errorf("failed to find old imports: %w", err)
	}
	defer cursor.Close(ctx)

	var ids []uuid.UUID
	for cursor.Next(ctx) {
		var item struct {
			ID uuid.UUID `bson:"_id"`
		}
		if err := cursor.Decode(&item); err != nil {
			continue
		}
		ids = append(ids, item.ID)
	}

	// Delete errors for these imports
	if len(ids) > 0 {
		_, err = r.errorsCollection.DeleteMany(ctx, bson.M{"import_id": bson.M{"$in": ids}})
		if err != nil {
			return fmt.Errorf("failed to delete old import errors: %w", err)
		}
	}

	// Delete the imports
	_, err = r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete old imports: %w", err)
	}

	return nil
}
