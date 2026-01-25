// Package mongodb provides MongoDB implementations for customer repositories.
package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

const (
	customersCollection = "customers"
)

// CustomerRepository implements domain.CustomerRepository using MongoDB.
type CustomerRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewCustomerRepository creates a new CustomerRepository.
func NewCustomerRepository(db *mongo.Database) *CustomerRepository {
	return &CustomerRepository{
		db:         db,
		collection: db.Collection(customersCollection),
	}
}

// Create creates a new customer.
func (r *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	customer.CreatedAt = time.Now().UTC()
	customer.UpdatedAt = customer.CreatedAt
	customer.Version = 1

	_, err := r.collection.InsertOne(ctx, customer)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrCustomerAlreadyExists
		}
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// Update updates an existing customer.
func (r *CustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	customer.UpdatedAt = time.Now().UTC()
	previousVersion := customer.Version
	customer.Version++

	filter := bson.M{
		"_id":     customer.ID,
		"version": previousVersion,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, customer)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	if result.MatchedCount == 0 {
		// Check if customer exists
		var existing domain.Customer
		err := r.collection.FindOne(ctx, bson.M{"_id": customer.ID}).Decode(&existing)
		if err == mongo.ErrNoDocuments {
			return domain.ErrCustomerNotFound
		}
		return domain.ErrVersionConflict
	}

	return nil
}

// Delete soft deletes a customer.
func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "deleted_at": nil}, update)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

// HardDelete permanently deletes a customer.
func (r *CustomerRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to hard delete customer: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrCustomerNotFound
	}

	return nil
}

// FindByID finds a customer by ID.
func (r *CustomerRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "deleted_at": nil}).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to find customer: %w", err)
	}

	return &customer, nil
}

// FindByCode finds a customer by code.
func (r *CustomerRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Customer, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"code":       code,
		"deleted_at": nil,
	}

	var customer domain.Customer
	err := r.collection.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to find customer by code: %w", err)
	}

	return &customer, nil
}

// FindByEmail finds a customer by email.
func (r *CustomerRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.Customer, error) {
	filter := bson.M{
		"tenant_id":     tenantID,
		"email.address": strings.ToLower(email),
		"deleted_at":    nil,
	}

	var customer domain.Customer
	err := r.collection.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to find customer by email: %w", err)
	}

	return &customer, nil
}

// List lists customers with filtering and pagination.
func (r *CustomerRepository) List(ctx context.Context, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	mongoFilter := r.buildFilter(filter)

	// Count total
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	// Build options
	findOpts := options.Find().
		SetSkip(int64(filter.Offset)).
		SetLimit(int64(filter.Limit))

	// Sorting
	sortField := "created_at"
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
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("failed to decode customers: %w", err)
	}

	return &domain.CustomerList{
		Customers: customers,
		Total:     total,
		Offset:    filter.Offset,
		Limit:     filter.Limit,
		HasMore:   int64(filter.Offset+len(customers)) < total,
	}, nil
}

// Search performs full-text search on customers.
func (r *CustomerRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	mongoFilter := r.buildFilter(filter)
	mongoFilter["tenant_id"] = tenantID

	// Add text search
	if query != "" {
		mongoFilter["$text"] = bson.M{"$search": query}
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Build options
	findOpts := options.Find().
		SetSkip(int64(filter.Offset)).
		SetLimit(int64(filter.Limit))

	// Add text score for relevance sorting if searching
	if query != "" {
		findOpts.SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}})
		findOpts.SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}})
	} else {
		sortField := "created_at"
		if filter.SortBy != "" {
			sortField = filter.SortBy
		}
		sortOrder := -1
		if filter.SortOrder == "asc" {
			sortOrder = 1
		}
		findOpts.SetSort(bson.D{{Key: sortField, Value: sortOrder}})
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("failed to decode customers: %w", err)
	}

	return &domain.CustomerList{
		Customers: customers,
		Total:     total,
		Offset:    filter.Offset,
		Limit:     filter.Limit,
		HasMore:   int64(filter.Offset+len(customers)) < total,
	}, nil
}

// FindByOwner finds customers assigned to an owner.
func (r *CustomerRepository) FindByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	filter.TenantID = &tenantID
	filter.OwnerIDs = []uuid.UUID{ownerID}
	return r.List(ctx, filter)
}

// FindByTag finds customers with a specific tag.
func (r *CustomerRepository) FindByTag(ctx context.Context, tenantID uuid.UUID, tag string, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	filter.TenantID = &tenantID
	filter.Tags = []string{tag}
	return r.List(ctx, filter)
}

// FindBySegment finds customers in a segment.
func (r *CustomerRepository) FindBySegment(ctx context.Context, tenantID, segmentID uuid.UUID, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	filter.TenantID = &tenantID
	filter.SegmentIDs = []uuid.UUID{segmentID}
	return r.List(ctx, filter)
}

// FindByStatus finds customers by status.
func (r *CustomerRepository) FindByStatus(ctx context.Context, tenantID uuid.UUID, status domain.CustomerStatus, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	filter.TenantID = &tenantID
	filter.Statuses = []domain.CustomerStatus{status}
	return r.List(ctx, filter)
}

// FindDuplicates finds potential duplicate customers.
func (r *CustomerRepository) FindDuplicates(ctx context.Context, tenantID uuid.UUID, email, phone, name string) ([]*domain.Customer, error) {
	var orConditions []bson.M

	if email != "" {
		orConditions = append(orConditions, bson.M{"email.address": strings.ToLower(email)})
	}
	if phone != "" {
		orConditions = append(orConditions, bson.M{"phone_numbers.e164": phone})
	}
	if name != "" {
		orConditions = append(orConditions, bson.M{"name": primitive.Regex{Pattern: "^" + name + "$", Options: "i"}})
	}

	if len(orConditions) == 0 {
		return nil, nil
	}

	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"$or":        orConditions,
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetLimit(10))
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicates: %w", err)
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("failed to decode duplicates: %w", err)
	}

	return customers, nil
}

// CountByTenant counts customers for a tenant.
func (r *CustomerRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"tenant_id": tenantID, "deleted_at": nil})
}

// CountByStatus counts customers by status.
func (r *CustomerRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.CustomerStatus]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"tenant_id": tenantID, "deleted_at": nil}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$status",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to count by status: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[domain.CustomerStatus]int64)
	for cursor.Next(ctx) {
		var item struct {
			ID    domain.CustomerStatus `bson:"_id"`
			Count int64                 `bson:"count"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("failed to decode count: %w", err)
		}
		result[item.ID] = item.Count
	}

	return result, nil
}

// GetStats gets customer statistics.
func (r *CustomerRepository) GetStats(ctx context.Context, tenantID uuid.UUID) (*domain.CustomerStats, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"tenant_id": tenantID, "deleted_at": nil}}},
		{{Key: "$group", Value: bson.M{
			"_id":                  nil,
			"total":                bson.M{"$sum": 1},
			"total_contacts":       bson.M{"$sum": bson.M{"$size": bson.M{"$ifNull": bson.A{"$contacts", bson.A{}}}}},
			"total_engagement":     bson.M{"$sum": "$stats.engagement_score"},
			"total_health":         bson.M{"$sum": "$stats.health_score"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer cursor.Close(ctx)

	stats := &domain.CustomerStats{}
	if cursor.Next(ctx) {
		var result struct {
			Total            int `bson:"total"`
			TotalContacts    int `bson:"total_contacts"`
			TotalEngagement  int `bson:"total_engagement"`
			TotalHealth      int `bson:"total_health"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode stats: %w", err)
		}

		if result.Total > 0 {
			stats.ContactCount = result.TotalContacts
			stats.EngagementScore = result.TotalEngagement / result.Total
			stats.HealthScore = result.TotalHealth / result.Total
		}
	}

	now := time.Now().UTC()
	stats.LastCalculatedAt = &now

	return stats, nil
}

// BulkCreate creates multiple customers.
func (r *CustomerRepository) BulkCreate(ctx context.Context, customers []*domain.Customer) error {
	if len(customers) == 0 {
		return nil
	}

	now := time.Now().UTC()
	docs := make([]interface{}, len(customers))
	for i, customer := range customers {
		customer.CreatedAt = now
		customer.UpdatedAt = now
		customer.Version = 1
		docs[i] = customer
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to bulk create customers: %w", err)
	}

	return nil
}

// BulkUpdate updates multiple customers.
func (r *CustomerRepository) BulkUpdate(ctx context.Context, ids []uuid.UUID, updates map[string]interface{}) error {
	if len(ids) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now().UTC()

	filter := bson.M{"_id": bson.M{"$in": ids}, "deleted_at": nil}
	update := bson.M{
		"$set": updates,
		"$inc": bson.M{"version": 1},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to bulk update customers: %w", err)
	}

	return nil
}

// BulkDelete soft deletes multiple customers.
func (r *CustomerRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	filter := bson.M{"_id": bson.M{"$in": ids}, "deleted_at": nil}
	update := bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
		"$inc": bson.M{"version": 1},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to bulk delete customers: %w", err)
	}

	return nil
}

// Exists checks if a customer exists.
func (r *CustomerRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id, "deleted_at": nil})
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByCode checks if a customer code exists.
func (r *CustomerRepository) ExistsByCode(ctx context.Context, tenantID uuid.UUID, code string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"tenant_id":  tenantID,
		"code":       code,
		"deleted_at": nil,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByEmail checks if a customer email exists.
func (r *CustomerRepository) ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"tenant_id":     tenantID,
		"email.address": strings.ToLower(email),
		"deleted_at":    nil,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}

// GetVersion gets the current version for optimistic locking.
func (r *CustomerRepository) GetVersion(ctx context.Context, id uuid.UUID) (int, error) {
	var result struct {
		Version int `bson:"version"`
	}
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, domain.ErrCustomerNotFound
		}
		return 0, fmt.Errorf("failed to get version: %w", err)
	}
	return result.Version, nil
}

// FindNeedingFollowUp finds customers needing follow-up.
func (r *CustomerRepository) FindNeedingFollowUp(ctx context.Context, tenantID uuid.UUID, before time.Time) ([]*domain.Customer, error) {
	filter := bson.M{
		"tenant_id":        tenantID,
		"deleted_at":       nil,
		"next_follow_up_at": bson.M{"$lte": before},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetLimit(100))
	if err != nil {
		return nil, fmt.Errorf("failed to find customers needing follow-up: %w", err)
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("failed to decode customers: %w", err)
	}

	return customers, nil
}

// FindInactive finds inactive customers.
func (r *CustomerRepository) FindInactive(ctx context.Context, tenantID uuid.UUID, lastContactBefore time.Time) ([]*domain.Customer, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"$or": []bson.M{
			{"last_contacted_at": bson.M{"$lt": lastContactBefore}},
			{"last_contacted_at": nil},
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetLimit(100))
	if err != nil {
		return nil, fmt.Errorf("failed to find inactive customers: %w", err)
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("failed to decode customers: %w", err)
	}

	return customers, nil
}

// FindRecentlyCreated finds recently created customers.
func (r *CustomerRepository) FindRecentlyCreated(ctx context.Context, tenantID uuid.UUID, since time.Time, limit int) ([]*domain.Customer, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"created_at": bson.M{"$gte": since},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find recently created customers: %w", err)
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("failed to decode customers: %w", err)
	}

	return customers, nil
}

// FindRecentlyUpdated finds recently updated customers.
func (r *CustomerRepository) FindRecentlyUpdated(ctx context.Context, tenantID uuid.UUID, since time.Time, limit int) ([]*domain.Customer, error) {
	filter := bson.M{
		"tenant_id":  tenantID,
		"deleted_at": nil,
		"updated_at": bson.M{"$gte": since},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find recently updated customers: %w", err)
	}
	defer cursor.Close(ctx)

	var customers []*domain.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		return nil, fmt.Errorf("failed to decode customers: %w", err)
	}

	return customers, nil
}

// buildFilter builds a MongoDB filter from CustomerFilter.
func (r *CustomerRepository) buildFilter(filter domain.CustomerFilter) bson.M {
	mongoFilter := bson.M{}

	if !filter.IncludeDeleted {
		mongoFilter["deleted_at"] = nil
	}

	if filter.TenantID != nil {
		mongoFilter["tenant_id"] = *filter.TenantID
	}

	if len(filter.IDs) > 0 {
		mongoFilter["_id"] = bson.M{"$in": filter.IDs}
	}

	if len(filter.Codes) > 0 {
		mongoFilter["code"] = bson.M{"$in": filter.Codes}
	}

	if len(filter.Types) > 0 {
		mongoFilter["type"] = bson.M{"$in": filter.Types}
	}

	if len(filter.Statuses) > 0 {
		mongoFilter["status"] = bson.M{"$in": filter.Statuses}
	}

	if len(filter.Tiers) > 0 {
		mongoFilter["tier"] = bson.M{"$in": filter.Tiers}
	}

	if len(filter.Sources) > 0 {
		mongoFilter["source"] = bson.M{"$in": filter.Sources}
	}

	if len(filter.Tags) > 0 {
		mongoFilter["tags"] = bson.M{"$all": filter.Tags}
	}

	if len(filter.OwnerIDs) > 0 {
		mongoFilter["owner_id"] = bson.M{"$in": filter.OwnerIDs}
	}

	if len(filter.SegmentIDs) > 0 {
		mongoFilter["segments"] = bson.M{"$in": filter.SegmentIDs}
	}

	if len(filter.Industries) > 0 {
		mongoFilter["company_info.industry"] = bson.M{"$in": filter.Industries}
	}

	if len(filter.Countries) > 0 {
		mongoFilter["addresses.country_code"] = bson.M{"$in": filter.Countries}
	}

	if filter.Email != "" {
		mongoFilter["email.address"] = strings.ToLower(filter.Email)
	}

	if filter.Phone != "" {
		mongoFilter["phone_numbers.e164"] = filter.Phone
	}

	if filter.HasDeals != nil && *filter.HasDeals {
		mongoFilter["stats.deal_count"] = bson.M{"$gt": 0}
	}

	if filter.HasOpenDeals != nil && *filter.HasOpenDeals {
		mongoFilter["$expr"] = bson.M{
			"$gt": bson.A{
				"$stats.deal_count",
				bson.M{"$add": bson.A{"$stats.won_deal_count", "$stats.lost_deal_count"}},
			},
		}
	}

	if filter.CreatedAfter != nil {
		mongoFilter["created_at"] = bson.M{"$gte": *filter.CreatedAfter}
	}

	if filter.CreatedBefore != nil {
		if _, exists := mongoFilter["created_at"]; exists {
			mongoFilter["created_at"].(bson.M)["$lte"] = *filter.CreatedBefore
		} else {
			mongoFilter["created_at"] = bson.M{"$lte": *filter.CreatedBefore}
		}
	}

	if filter.UpdatedAfter != nil {
		mongoFilter["updated_at"] = bson.M{"$gte": *filter.UpdatedAfter}
	}

	if filter.UpdatedBefore != nil {
		if _, exists := mongoFilter["updated_at"]; exists {
			mongoFilter["updated_at"].(bson.M)["$lte"] = *filter.UpdatedBefore
		} else {
			mongoFilter["updated_at"] = bson.M{"$lte": *filter.UpdatedBefore}
		}
	}

	if filter.LastContactAfter != nil {
		mongoFilter["last_contacted_at"] = bson.M{"$gte": *filter.LastContactAfter}
	}

	if filter.LastContactBefore != nil {
		if _, exists := mongoFilter["last_contacted_at"]; exists {
			mongoFilter["last_contacted_at"].(bson.M)["$lte"] = *filter.LastContactBefore
		} else {
			mongoFilter["last_contacted_at"] = bson.M{"$lte": *filter.LastContactBefore}
		}
	}

	return mongoFilter
}
