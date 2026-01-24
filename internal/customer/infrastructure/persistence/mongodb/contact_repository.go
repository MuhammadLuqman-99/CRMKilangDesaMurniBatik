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

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
)

const (
	contactsCollection = "contacts"
)

// ContactRepository implements domain.ContactRepository using MongoDB.
type ContactRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewContactRepository creates a new ContactRepository.
func NewContactRepository(db *mongo.Database) *ContactRepository {
	return &ContactRepository{
		db:         db,
		collection: db.Collection(contactsCollection),
	}
}

// Create creates a new contact.
func (r *ContactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	contact.CreatedAt = time.Now().UTC()
	contact.UpdatedAt = contact.CreatedAt
	contact.Version = 1

	_, err := r.collection.InsertOne(ctx, contact)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrContactAlreadyExists
		}
		return fmt.Errorf("failed to create contact: %w", err)
	}

	return nil
}

// Update updates an existing contact.
func (r *ContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	contact.UpdatedAt = time.Now().UTC()
	previousVersion := contact.Version
	contact.Version++

	filter := bson.M{
		"_id":     contact.ID,
		"version": previousVersion,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, contact)
	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}

	if result.MatchedCount == 0 {
		var existing domain.Contact
		err := r.collection.FindOne(ctx, bson.M{"_id": contact.ID}).Decode(&existing)
		if err == mongo.ErrNoDocuments {
			return domain.ErrContactNotFound
		}
		return domain.ErrVersionConflict
	}

	return nil
}

// Delete soft deletes a contact.
func (r *ContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
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
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrContactNotFound
	}

	return nil
}

// HardDelete permanently deletes a contact.
func (r *ContactRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to hard delete contact: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrContactNotFound
	}

	return nil
}

// FindByID finds a contact by ID.
func (r *ContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	var contact domain.Contact
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "deleted_at": nil}).Decode(&contact)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrContactNotFound
		}
		return nil, fmt.Errorf("failed to find contact: %w", err)
	}

	return &contact, nil
}

// FindByCustomerID finds all contacts for a customer.
func (r *ContactRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Contact, error) {
	filter := bson.M{
		"customer_id": customerID,
		"deleted_at":  nil,
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{
		{Key: "is_primary", Value: -1},
		{Key: "created_at", Value: 1},
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to find contacts by customer: %w", err)
	}
	defer cursor.Close(ctx)

	var contacts []*domain.Contact
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("failed to decode contacts: %w", err)
	}

	return contacts, nil
}

// FindByEmail finds contacts by email.
func (r *ContactRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]*domain.Contact, error) {
	filter := bson.M{
		"tenant_id":     tenantID,
		"email.address": strings.ToLower(email),
		"deleted_at":    nil,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find contacts by email: %w", err)
	}
	defer cursor.Close(ctx)

	var contacts []*domain.Contact
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("failed to decode contacts: %w", err)
	}

	return contacts, nil
}

// Search searches contacts.
func (r *ContactRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, filter domain.ContactFilter) (*domain.ContactList, error) {
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
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}
	defer cursor.Close(ctx)

	var contacts []*domain.Contact
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("failed to decode contacts: %w", err)
	}

	return &domain.ContactList{
		Contacts: contacts,
		Total:    total,
		Offset:   filter.Offset,
		Limit:    filter.Limit,
		HasMore:  int64(filter.Offset+len(contacts)) < total,
	}, nil
}

// CountByCustomer counts contacts for a customer.
func (r *ContactRepository) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"customer_id": customerID,
		"deleted_at":  nil,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count contacts: %w", err)
	}
	return int(count), nil
}

// List lists contacts with filtering and pagination.
func (r *ContactRepository) List(ctx context.Context, filter domain.ContactFilter) (*domain.ContactList, error) {
	mongoFilter := r.buildFilter(filter)

	// Count total
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count contacts: %w", err)
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
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}
	defer cursor.Close(ctx)

	var contacts []*domain.Contact
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("failed to decode contacts: %w", err)
	}

	return &domain.ContactList{
		Contacts: contacts,
		Total:    total,
		Offset:   filter.Offset,
		Limit:    filter.Limit,
		HasMore:  int64(filter.Offset+len(contacts)) < total,
	}, nil
}

// FindPrimaryByCustomer finds the primary contact for a customer.
func (r *ContactRepository) FindPrimaryByCustomer(ctx context.Context, customerID uuid.UUID) (*domain.Contact, error) {
	filter := bson.M{
		"customer_id": customerID,
		"is_primary":  true,
		"deleted_at":  nil,
	}

	var contact domain.Contact
	err := r.collection.FindOne(ctx, filter).Decode(&contact)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find primary contact: %w", err)
	}

	return &contact, nil
}

// ClearPrimaryForCustomer clears the primary flag for all contacts of a customer.
func (r *ContactRepository) ClearPrimaryForCustomer(ctx context.Context, customerID uuid.UUID) error {
	filter := bson.M{
		"customer_id": customerID,
		"is_primary":  true,
		"deleted_at":  nil,
	}
	update := bson.M{
		"$set": bson.M{
			"is_primary": false,
			"updated_at": time.Now().UTC(),
		},
		"$inc": bson.M{"version": 1},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to clear primary contacts: %w", err)
	}

	return nil
}

// SetPrimary sets a contact as primary.
func (r *ContactRepository) SetPrimary(ctx context.Context, contactID uuid.UUID) error {
	now := time.Now().UTC()

	// Get the contact to find customer ID
	contact, err := r.FindByID(ctx, contactID)
	if err != nil {
		return err
	}

	// Clear existing primary
	if err := r.ClearPrimaryForCustomer(ctx, contact.CustomerID); err != nil {
		return err
	}

	// Set new primary
	filter := bson.M{"_id": contactID, "deleted_at": nil}
	update := bson.M{
		"$set": bson.M{
			"is_primary": true,
			"updated_at": now,
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to set primary contact: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrContactNotFound
	}

	return nil
}

// FindByPhone finds contacts by phone number.
func (r *ContactRepository) FindByPhone(ctx context.Context, tenantID uuid.UUID, phone string) ([]*domain.Contact, error) {
	filter := bson.M{
		"tenant_id":          tenantID,
		"phone_numbers.e164": phone,
		"deleted_at":         nil,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find contacts by phone: %w", err)
	}
	defer cursor.Close(ctx)

	var contacts []*domain.Contact
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("failed to decode contacts: %w", err)
	}

	return contacts, nil
}

// FindDuplicates finds potential duplicate contacts.
func (r *ContactRepository) FindDuplicates(ctx context.Context, tenantID uuid.UUID, email, phone, name string) ([]*domain.Contact, error) {
	var orConditions []bson.M

	if email != "" {
		orConditions = append(orConditions, bson.M{"email.address": strings.ToLower(email)})
	}
	if phone != "" {
		orConditions = append(orConditions, bson.M{"phone_numbers.e164": phone})
	}
	if name != "" {
		orConditions = append(orConditions, bson.M{
			"$or": []bson.M{
				{"name.first_name": primitive.Regex{Pattern: "^" + name + "$", Options: "i"}},
				{"name.last_name": primitive.Regex{Pattern: "^" + name + "$", Options: "i"}},
			},
		})
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

	var contacts []*domain.Contact
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("failed to decode duplicates: %w", err)
	}

	return contacts, nil
}

// CountByTenant counts contacts for a tenant.
func (r *ContactRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"tenant_id": tenantID, "deleted_at": nil})
}

// CountByStatus counts contacts by status.
func (r *ContactRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.ContactStatus]int64, error) {
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

	result := make(map[domain.ContactStatus]int64)
	for cursor.Next(ctx) {
		var item struct {
			ID    domain.ContactStatus `bson:"_id"`
			Count int64                `bson:"count"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, fmt.Errorf("failed to decode count: %w", err)
		}
		result[item.ID] = item.Count
	}

	return result, nil
}

// BulkCreate creates multiple contacts.
func (r *ContactRepository) BulkCreate(ctx context.Context, contacts []*domain.Contact) error {
	if len(contacts) == 0 {
		return nil
	}

	now := time.Now().UTC()
	docs := make([]interface{}, len(contacts))
	for i, contact := range contacts {
		contact.CreatedAt = now
		contact.UpdatedAt = now
		contact.Version = 1
		docs[i] = contact
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to bulk create contacts: %w", err)
	}

	return nil
}

// BulkDelete soft deletes multiple contacts.
func (r *ContactRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
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
		return fmt.Errorf("failed to bulk delete contacts: %w", err)
	}

	return nil
}

// DeleteByCustomer soft deletes all contacts for a customer.
func (r *ContactRepository) DeleteByCustomer(ctx context.Context, customerID uuid.UUID) error {
	now := time.Now().UTC()
	filter := bson.M{
		"customer_id": customerID,
		"deleted_at":  nil,
	}
	update := bson.M{
		"$set": bson.M{"deleted_at": now, "updated_at": now},
		"$inc": bson.M{"version": 1},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete contacts by customer: %w", err)
	}

	return nil
}

// FindNeedingFollowUp finds contacts needing follow-up.
func (r *ContactRepository) FindNeedingFollowUp(ctx context.Context, tenantID uuid.UUID, before time.Time) ([]*domain.Contact, error) {
	filter := bson.M{
		"tenant_id":        tenantID,
		"deleted_at":       nil,
		"next_follow_up_at": bson.M{"$lte": before},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetLimit(100))
	if err != nil {
		return nil, fmt.Errorf("failed to find contacts needing follow-up: %w", err)
	}
	defer cursor.Close(ctx)

	var contacts []*domain.Contact
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("failed to decode contacts: %w", err)
	}

	return contacts, nil
}

// FindMarketingOptIn finds contacts opted in for marketing.
func (r *ContactRepository) FindMarketingOptIn(ctx context.Context, tenantID uuid.UUID, filter domain.ContactFilter) (*domain.ContactList, error) {
	mongoFilter := r.buildFilter(filter)
	mongoFilter["tenant_id"] = tenantID
	mongoFilter["opted_out_marketing"] = false
	mongoFilter["status"] = domain.ContactStatusActive

	// Count total
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count marketing contacts: %w", err)
	}

	// Build options
	findOpts := options.Find().
		SetSkip(int64(filter.Offset)).
		SetLimit(int64(filter.Limit))

	cursor, err := r.collection.Find(ctx, mongoFilter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to find marketing contacts: %w", err)
	}
	defer cursor.Close(ctx)

	var contacts []*domain.Contact
	if err := cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("failed to decode contacts: %w", err)
	}

	return &domain.ContactList{
		Contacts: contacts,
		Total:    total,
		Offset:   filter.Offset,
		Limit:    filter.Limit,
		HasMore:  int64(filter.Offset+len(contacts)) < total,
	}, nil
}

// Exists checks if a contact exists.
func (r *ContactRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"_id": id, "deleted_at": nil})
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}

// GetVersion gets the current version for optimistic locking.
func (r *ContactRepository) GetVersion(ctx context.Context, id uuid.UUID) (int, error) {
	var result struct {
		Version int `bson:"version"`
	}
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, domain.ErrContactNotFound
		}
		return 0, fmt.Errorf("failed to get version: %w", err)
	}
	return result.Version, nil
}

// buildFilter builds a MongoDB filter from ContactFilter.
func (r *ContactRepository) buildFilter(filter domain.ContactFilter) bson.M {
	mongoFilter := bson.M{}

	if !filter.IncludeDeleted {
		mongoFilter["deleted_at"] = nil
	}

	if filter.TenantID != nil {
		mongoFilter["tenant_id"] = *filter.TenantID
	}

	if filter.CustomerID != nil {
		mongoFilter["customer_id"] = *filter.CustomerID
	}

	if len(filter.CustomerIDs) > 0 {
		mongoFilter["customer_id"] = bson.M{"$in": filter.CustomerIDs}
	}

	if len(filter.Statuses) > 0 {
		mongoFilter["status"] = bson.M{"$in": filter.Statuses}
	}

	if len(filter.Roles) > 0 {
		mongoFilter["role"] = bson.M{"$in": filter.Roles}
	}

	if filter.IsPrimary != nil {
		mongoFilter["is_primary"] = *filter.IsPrimary
	}

	if filter.HasEmail != nil && *filter.HasEmail {
		mongoFilter["email.address"] = bson.M{"$ne": ""}
	}

	if filter.HasPhone != nil && *filter.HasPhone {
		mongoFilter["phone_numbers.0"] = bson.M{"$exists": true}
	}

	if filter.OptedInMarketing != nil {
		mongoFilter["opted_out_marketing"] = !*filter.OptedInMarketing
	}

	if len(filter.Tags) > 0 {
		mongoFilter["tags"] = bson.M{"$all": filter.Tags}
	}

	if filter.MinEngagement != nil {
		mongoFilter["engagement_score"] = bson.M{"$gte": *filter.MinEngagement}
	}

	if filter.MaxEngagement != nil {
		if _, exists := mongoFilter["engagement_score"]; exists {
			mongoFilter["engagement_score"].(bson.M)["$lte"] = *filter.MaxEngagement
		} else {
			mongoFilter["engagement_score"] = bson.M{"$lte": *filter.MaxEngagement}
		}
	}

	if filter.NeedsFollowUp != nil && *filter.NeedsFollowUp {
		mongoFilter["next_follow_up_at"] = bson.M{"$lte": time.Now().UTC()}
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

	if filter.LastContactedAfter != nil {
		mongoFilter["last_contacted_at"] = bson.M{"$gte": *filter.LastContactedAfter}
	}

	if filter.LastContactedBefore != nil {
		if _, exists := mongoFilter["last_contacted_at"]; exists {
			mongoFilter["last_contacted_at"].(bson.M)["$lte"] = *filter.LastContactedBefore
		} else {
			mongoFilter["last_contacted_at"] = bson.M{"$lte": *filter.LastContactedBefore}
		}
	}

	return mongoFilter
}
