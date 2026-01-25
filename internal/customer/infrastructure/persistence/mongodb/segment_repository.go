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

	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

const (
	segmentsCollection = "customer_segments"
)

// SegmentRepository implements domain.SegmentRepository using MongoDB.
type SegmentRepository struct {
	db                 *mongo.Database
	collection         *mongo.Collection
	customerCollection *mongo.Collection
}

// NewSegmentRepository creates a new SegmentRepository.
func NewSegmentRepository(db *mongo.Database) *SegmentRepository {
	return &SegmentRepository{
		db:                 db,
		collection:         db.Collection(segmentsCollection),
		customerCollection: db.Collection(customersCollection),
	}
}

// Create creates a new segment.
func (r *SegmentRepository) Create(ctx context.Context, segment *domain.Segment) error {
	segment.CreatedAt = time.Now().UTC()
	segment.UpdatedAt = segment.CreatedAt
	segment.Version = 1

	_, err := r.collection.InsertOne(ctx, segment)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrSegmentAlreadyExists
		}
		return fmt.Errorf("failed to create segment: %w", err)
	}

	return nil
}

// Update updates a segment.
func (r *SegmentRepository) Update(ctx context.Context, segment *domain.Segment) error {
	segment.UpdatedAt = time.Now().UTC()
	previousVersion := segment.Version
	segment.Version++

	filter := bson.M{
		"_id":     segment.ID,
		"version": previousVersion,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, segment)
	if err != nil {
		return fmt.Errorf("failed to update segment: %w", err)
	}

	if result.MatchedCount == 0 {
		var existing domain.Segment
		err := r.collection.FindOne(ctx, bson.M{"_id": segment.ID}).Decode(&existing)
		if err == mongo.ErrNoDocuments {
			return domain.ErrSegmentNotFound
		}
		return domain.ErrVersionConflict
	}

	return nil
}

// Delete deletes a segment.
func (r *SegmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete segment: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrSegmentNotFound
	}

	// Remove segment from customers
	_, err = r.customerCollection.UpdateMany(
		ctx,
		bson.M{"segments": id},
		bson.M{"$pull": bson.M{"segments": id}},
	)
	if err != nil {
		return fmt.Errorf("failed to remove segment from customers: %w", err)
	}

	return nil
}

// FindByID finds a segment by ID.
func (r *SegmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Segment, error) {
	var segment domain.Segment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&segment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrSegmentNotFound
		}
		return nil, fmt.Errorf("failed to find segment: %w", err)
	}

	return &segment, nil
}

// FindByTenantID finds all segments for a tenant.
func (r *SegmentRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Segment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"tenant_id": tenantID},
		options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to find segments by tenant: %w", err)
	}
	defer cursor.Close(ctx)

	var segments []*domain.Segment
	if err := cursor.All(ctx, &segments); err != nil {
		return nil, fmt.Errorf("failed to decode segments: %w", err)
	}

	return segments, nil
}

// FindByName finds a segment by name.
func (r *SegmentRepository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.Segment, error) {
	filter := bson.M{
		"tenant_id": tenantID,
		"name":      name,
	}

	var segment domain.Segment
	err := r.collection.FindOne(ctx, filter).Decode(&segment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrSegmentNotFound
		}
		return nil, fmt.Errorf("failed to find segment by name: %w", err)
	}

	return &segment, nil
}

// GetCustomerCount gets the number of customers in a segment.
func (r *SegmentRepository) GetCustomerCount(ctx context.Context, segmentID uuid.UUID) (int64, error) {
	count, err := r.customerCollection.CountDocuments(ctx, bson.M{
		"segments":   segmentID,
		"deleted_at": nil,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count segment customers: %w", err)
	}

	return count, nil
}

// RefreshDynamic refreshes dynamic segment membership.
func (r *SegmentRepository) RefreshDynamic(ctx context.Context, segmentID uuid.UUID) error {
	// Find the segment
	segment, err := r.FindByID(ctx, segmentID)
	if err != nil {
		return err
	}

	if segment.Type != domain.SegmentTypeDynamic {
		return fmt.Errorf("segment is not dynamic")
	}

	if segment.Criteria == nil || len(segment.Criteria.Rules) == 0 {
		return fmt.Errorf("segment has no criteria")
	}

	// Build filter from criteria
	mongoFilter := r.buildFilterFromCriteria(segment.TenantID, segment.Criteria)

	// First, remove this segment from all customers
	_, err = r.customerCollection.UpdateMany(
		ctx,
		bson.M{"tenant_id": segment.TenantID, "deleted_at": nil},
		bson.M{"$pull": bson.M{"segments": segmentID}},
	)
	if err != nil {
		return fmt.Errorf("failed to clear segment membership: %w", err)
	}

	// Add segment to matching customers
	_, err = r.customerCollection.UpdateMany(
		ctx,
		mongoFilter,
		bson.M{"$addToSet": bson.M{"segments": segmentID}},
	)
	if err != nil {
		return fmt.Errorf("failed to update segment membership: %w", err)
	}

	// Update customer count
	count, err := r.GetCustomerCount(ctx, segmentID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": segmentID},
		bson.M{
			"$set": bson.M{
				"customer_count": count,
				"last_refreshed": now,
				"updated_at":     now,
			},
			"$inc": bson.M{"version": 1},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update segment count: %w", err)
	}

	return nil
}

// buildFilterFromCriteria builds a MongoDB filter from segment criteria.
func (r *SegmentRepository) buildFilterFromCriteria(tenantID uuid.UUID, criteria *domain.SegmentCriteria) bson.M {
	if criteria == nil || len(criteria.Rules) == 0 {
		return bson.M{"tenant_id": tenantID, "deleted_at": nil}
	}

	var conditions []bson.M
	for _, rule := range criteria.Rules {
		condition := r.buildConditionFromRule(rule)
		if condition != nil {
			conditions = append(conditions, condition)
		}
	}

	mongoFilter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
	}

	if len(conditions) > 0 {
		if criteria.Operator == "or" {
			mongoFilter["$or"] = conditions
		} else {
			// Default to AND
			for _, cond := range conditions {
				for k, v := range cond {
					mongoFilter[k] = v
				}
			}
		}
	}

	return mongoFilter
}

// buildConditionFromRule builds a single MongoDB condition from a segment rule.
func (r *SegmentRepository) buildConditionFromRule(rule domain.SegmentRule) bson.M {
	switch rule.Operator {
	case "eq", "equals":
		return bson.M{rule.Field: rule.Value}
	case "ne", "not_equals":
		return bson.M{rule.Field: bson.M{"$ne": rule.Value}}
	case "gt", "greater_than":
		return bson.M{rule.Field: bson.M{"$gt": rule.Value}}
	case "gte", "greater_than_or_equals":
		return bson.M{rule.Field: bson.M{"$gte": rule.Value}}
	case "lt", "less_than":
		return bson.M{rule.Field: bson.M{"$lt": rule.Value}}
	case "lte", "less_than_or_equals":
		return bson.M{rule.Field: bson.M{"$lte": rule.Value}}
	case "in":
		return bson.M{rule.Field: bson.M{"$in": rule.Value}}
	case "nin", "not_in":
		return bson.M{rule.Field: bson.M{"$nin": rule.Value}}
	case "contains":
		if str, ok := rule.Value.(string); ok {
			return bson.M{rule.Field: bson.M{"$regex": str, "$options": "i"}}
		}
	case "starts_with":
		if str, ok := rule.Value.(string); ok {
			return bson.M{rule.Field: bson.M{"$regex": "^" + str, "$options": "i"}}
		}
	case "ends_with":
		if str, ok := rule.Value.(string); ok {
			return bson.M{rule.Field: bson.M{"$regex": str + "$", "$options": "i"}}
		}
	case "exists":
		return bson.M{rule.Field: bson.M{"$exists": true, "$ne": nil}}
	case "not_exists":
		return bson.M{rule.Field: bson.M{"$exists": false}}
	case "is_null":
		return bson.M{rule.Field: nil}
	case "is_not_null":
		return bson.M{rule.Field: bson.M{"$ne": nil}}
	}

	return nil
}

// AddCustomerToSegment adds a customer to a static segment.
func (r *SegmentRepository) AddCustomerToSegment(ctx context.Context, segmentID, customerID uuid.UUID) error {
	// Verify segment exists and is static
	segment, err := r.FindByID(ctx, segmentID)
	if err != nil {
		return err
	}

	if segment.Type != domain.SegmentTypeStatic {
		return fmt.Errorf("cannot manually add customers to dynamic segment")
	}

	// Add segment to customer
	result, err := r.customerCollection.UpdateOne(
		ctx,
		bson.M{"_id": customerID, "deleted_at": nil},
		bson.M{"$addToSet": bson.M{"segments": segmentID}},
	)
	if err != nil {
		return fmt.Errorf("failed to add customer to segment: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	// Update count if customer was added
	if result.ModifiedCount > 0 {
		_, err = r.collection.UpdateOne(
			ctx,
			bson.M{"_id": segmentID},
			bson.M{
				"$inc": bson.M{"customer_count": 1},
				"$set": bson.M{"updated_at": time.Now().UTC()},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to update segment count: %w", err)
		}
	}

	return nil
}

// RemoveCustomerFromSegment removes a customer from a static segment.
func (r *SegmentRepository) RemoveCustomerFromSegment(ctx context.Context, segmentID, customerID uuid.UUID) error {
	// Verify segment exists and is static
	segment, err := r.FindByID(ctx, segmentID)
	if err != nil {
		return err
	}

	if segment.Type != domain.SegmentTypeStatic {
		return fmt.Errorf("cannot manually remove customers from dynamic segment")
	}

	// Remove segment from customer
	result, err := r.customerCollection.UpdateOne(
		ctx,
		bson.M{"_id": customerID},
		bson.M{"$pull": bson.M{"segments": segmentID}},
	)
	if err != nil {
		return fmt.Errorf("failed to remove customer from segment: %w", err)
	}

	// Update count if customer was removed
	if result.ModifiedCount > 0 {
		_, err = r.collection.UpdateOne(
			ctx,
			bson.M{"_id": segmentID},
			bson.M{
				"$inc": bson.M{"customer_count": -1},
				"$set": bson.M{"updated_at": time.Now().UTC()},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to update segment count: %w", err)
		}
	}

	return nil
}

// FindActiveByTenant finds all active segments for a tenant.
func (r *SegmentRepository) FindActiveByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Segment, error) {
	filter := bson.M{
		"tenant_id": tenantID,
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to find active segments: %w", err)
	}
	defer cursor.Close(ctx)

	var segments []*domain.Segment
	if err := cursor.All(ctx, &segments); err != nil {
		return nil, fmt.Errorf("failed to decode segments: %w", err)
	}

	return segments, nil
}

// Exists checks if a segment exists.
func (r *SegmentRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id})
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}

// CountByTenant counts segments for a tenant.
func (r *SegmentRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"tenant_id": tenantID})
}
