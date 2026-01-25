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
	activitiesCollection = "customer_activities"
)

// ActivityRepository implements domain.ActivityRepository using MongoDB.
type ActivityRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewActivityRepository creates a new ActivityRepository.
func NewActivityRepository(db *mongo.Database) *ActivityRepository {
	return &ActivityRepository{
		db:         db,
		collection: db.Collection(activitiesCollection),
	}
}

// Create creates a new activity.
func (r *ActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	activity.CreatedAt = time.Now().UTC()
	activity.UpdatedAt = activity.CreatedAt
	activity.Version = 1

	if activity.OccurredAt.IsZero() {
		activity.OccurredAt = activity.CreatedAt
	}

	_, err := r.collection.InsertOne(ctx, activity)
	if err != nil {
		return fmt.Errorf("failed to create activity: %w", err)
	}

	return nil
}

// Update updates an activity.
func (r *ActivityRepository) Update(ctx context.Context, activity *domain.Activity) error {
	activity.UpdatedAt = time.Now().UTC()
	previousVersion := activity.Version
	activity.Version++

	filter := bson.M{
		"_id":     activity.ID,
		"version": previousVersion,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, activity)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}

	if result.MatchedCount == 0 {
		var existing domain.Activity
		err := r.collection.FindOne(ctx, bson.M{"_id": activity.ID}).Decode(&existing)
		if err == mongo.ErrNoDocuments {
			return domain.ErrActivityNotFound
		}
		return domain.ErrVersionConflict
	}

	return nil
}

// Delete deletes an activity.
func (r *ActivityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrActivityNotFound
	}

	return nil
}

// FindByID finds an activity by ID.
func (r *ActivityRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Activity, error) {
	var activity domain.Activity
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&activity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrActivityNotFound
		}
		return nil, fmt.Errorf("failed to find activity: %w", err)
	}

	return &activity, nil
}

// FindByCustomerID finds all activities for a customer.
func (r *ActivityRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	mongoFilter := r.buildFilter(filter)
	mongoFilter["customer_id"] = customerID

	return r.findActivities(ctx, mongoFilter, filter)
}

// FindByContactID finds all activities for a contact.
func (r *ActivityRepository) FindByContactID(ctx context.Context, contactID uuid.UUID, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	mongoFilter := r.buildFilter(filter)
	mongoFilter["contact_id"] = contactID

	return r.findActivities(ctx, mongoFilter, filter)
}

// FindByPerformedBy finds activities performed by a user.
func (r *ActivityRepository) FindByPerformedBy(ctx context.Context, userID uuid.UUID, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	mongoFilter := r.buildFilter(filter)
	mongoFilter["performed_by"] = userID

	return r.findActivities(ctx, mongoFilter, filter)
}

// findActivities is a helper to find activities with a filter.
func (r *ActivityRepository) findActivities(ctx context.Context, mongoFilter bson.M, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	// Count total
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count activities: %w", err)
	}

	// Build options
	findOpts := options.Find().
		SetSkip(int64(filter.Offset)).
		SetLimit(int64(filter.Limit))

	// Sorting
	sortField := "occurred_at"
	if filter.SortBy != "" {
		sortField = filter.SortBy
	}
	sortOrder := -1 // desc
	if filter.SortOrder == "asc" {
		sortOrder = 1
	}
	findOpts.SetSort(bson.D{{Key: sortField, Value: sortOrder}})

	// Execute query
	cursor, err := r.collection.Find(ctx, mongoFilter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to find activities: %w", err)
	}
	defer cursor.Close(ctx)

	var activities []*domain.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, fmt.Errorf("failed to decode activities: %w", err)
	}

	return &domain.ActivityList{
		Activities: activities,
		Total:      total,
		Offset:     filter.Offset,
		Limit:      filter.Limit,
		HasMore:    int64(filter.Offset+len(activities)) < total,
	}, nil
}

// CountByCustomer counts activities for a customer.
func (r *ActivityRepository) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return 0, fmt.Errorf("failed to count activities: %w", err)
	}
	return int(count), nil
}

// GetActivitySummary gets activity summary for a customer.
func (r *ActivityRepository) GetActivitySummary(ctx context.Context, customerID uuid.UUID) (map[domain.ActivityType]int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"customer_id": customerID}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$type",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity summary: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[domain.ActivityType]int)
	for cursor.Next(ctx) {
		var item struct {
			ID    domain.ActivityType `bson:"_id"`
			Count int                 `bson:"count"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("failed to decode summary: %w", err)
		}
		result[item.ID] = item.Count
	}

	return result, nil
}

// DeleteByCustomer deletes all activities for a customer.
func (r *ActivityRepository) DeleteByCustomer(ctx context.Context, customerID uuid.UUID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return fmt.Errorf("failed to delete activities by customer: %w", err)
	}
	return nil
}

// DeleteByContact deletes all activities for a contact.
func (r *ActivityRepository) DeleteByContact(ctx context.Context, contactID uuid.UUID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"contact_id": contactID})
	if err != nil {
		return fmt.Errorf("failed to delete activities by contact: %w", err)
	}
	return nil
}

// FindByDealID finds activities for a deal.
func (r *ActivityRepository) FindByDealID(ctx context.Context, dealID uuid.UUID, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	mongoFilter := r.buildFilter(filter)
	mongoFilter["deal_id"] = dealID

	return r.findActivities(ctx, mongoFilter, filter)
}

// FindByLeadID finds activities for a lead.
func (r *ActivityRepository) FindByLeadID(ctx context.Context, leadID uuid.UUID, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	mongoFilter := r.buildFilter(filter)
	mongoFilter["lead_id"] = leadID

	return r.findActivities(ctx, mongoFilter, filter)
}

// FindRecent finds recent activities for a tenant.
func (r *ActivityRepository) FindRecent(ctx context.Context, tenantID uuid.UUID, limit int) ([]*domain.Activity, error) {
	filter := bson.M{"tenant_id": tenantID}

	cursor, err := r.collection.Find(ctx, filter,
		options.Find().
			SetSort(bson.D{{Key: "occurred_at", Value: -1}}).
			SetLimit(int64(limit)))
	if err != nil {
		return nil, fmt.Errorf("failed to find recent activities: %w", err)
	}
	defer cursor.Close(ctx)

	var activities []*domain.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, fmt.Errorf("failed to decode activities: %w", err)
	}

	return activities, nil
}

// FindByDateRange finds activities within a date range.
func (r *ActivityRepository) FindByDateRange(ctx context.Context, tenantID uuid.UUID, start, end time.Time, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	mongoFilter := r.buildFilter(filter)
	mongoFilter["tenant_id"] = tenantID
	mongoFilter["occurred_at"] = bson.M{
		"$gte": start,
		"$lte": end,
	}

	return r.findActivities(ctx, mongoFilter, filter)
}

// GetDailyActivityCount gets daily activity counts for a customer.
func (r *ActivityRepository) GetDailyActivityCount(ctx context.Context, customerID uuid.UUID, days int) (map[string]int, error) {
	startDate := time.Now().UTC().AddDate(0, 0, -days)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"customer_id": customerID,
			"occurred_at": bson.M{"$gte": startDate},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$occurred_at",
				},
			},
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily counts: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]int)
	for cursor.Next(ctx) {
		var item struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("failed to decode count: %w", err)
		}
		result[item.ID] = item.Count
	}

	return result, nil
}

// BulkCreate creates multiple activities.
func (r *ActivityRepository) BulkCreate(ctx context.Context, activities []*domain.Activity) error {
	if len(activities) == 0 {
		return nil
	}

	now := time.Now().UTC()
	docs := make([]interface{}, len(activities))
	for i, activity := range activities {
		activity.CreatedAt = now
		activity.UpdatedAt = now
		activity.Version = 1
		if activity.OccurredAt.IsZero() {
			activity.OccurredAt = now
		}
		docs[i] = activity
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to bulk create activities: %w", err)
	}

	return nil
}

// buildFilter builds a MongoDB filter from ActivityFilter.
func (r *ActivityRepository) buildFilter(filter domain.ActivityFilter) bson.M {
	mongoFilter := bson.M{}

	if len(filter.Types) > 0 {
		mongoFilter["type"] = bson.M{"$in": filter.Types}
	}

	if len(filter.PerformedBy) > 0 {
		mongoFilter["performed_by"] = bson.M{"$in": filter.PerformedBy}
	}

	if filter.After != nil {
		mongoFilter["occurred_at"] = bson.M{"$gte": *filter.After}
	}

	if filter.Before != nil {
		if _, exists := mongoFilter["occurred_at"]; exists {
			mongoFilter["occurred_at"].(bson.M)["$lte"] = *filter.Before
		} else {
			mongoFilter["occurred_at"] = bson.M{"$lte": *filter.Before}
		}
	}

	return mongoFilter
}
