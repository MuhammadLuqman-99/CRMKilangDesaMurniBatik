// Package mongodb provides MongoDB implementations for customer repositories.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexManager manages MongoDB indexes for the customer service.
type IndexManager struct {
	db *mongo.Database
}

// NewIndexManager creates a new IndexManager.
func NewIndexManager(db *mongo.Database) *IndexManager {
	return &IndexManager{db: db}
}

// CreateAllIndexes creates all required indexes for the customer service.
func (m *IndexManager) CreateAllIndexes(ctx context.Context) error {
	if err := m.createCustomerIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create customer indexes: %w", err)
	}

	if err := m.createContactIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create contact indexes: %w", err)
	}

	if err := m.createNoteIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create note indexes: %w", err)
	}

	if err := m.createActivityIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create activity indexes: %w", err)
	}

	if err := m.createSegmentIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create segment indexes: %w", err)
	}

	if err := m.createImportIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create import indexes: %w", err)
	}

	if err := m.createOutboxIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create outbox indexes: %w", err)
	}

	return nil
}

// createCustomerIndexes creates indexes for the customers collection.
func (m *IndexManager) createCustomerIndexes(ctx context.Context) error {
	collection := m.db.Collection(customersCollection)

	indexes := []mongo.IndexModel{
		// Unique index on tenant_id + code
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "code", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"deleted_at": nil}).
				SetName("idx_customers_tenant_code_unique"),
		},
		// Unique index on tenant_id + email
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "email.address", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{
					"deleted_at":    nil,
					"email.address": bson.M{"$exists": true, "$ne": ""},
				}).
				SetSparse(true).
				SetName("idx_customers_tenant_email_unique"),
		},
		// Index for listing by tenant with soft delete filter
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "deleted_at", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_customers_tenant_listing"),
		},
		// Index for status filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_tenant_status"),
		},
		// Index for type filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "type", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_tenant_type"),
		},
		// Index for owner filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "owner_id", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_tenant_owner"),
		},
		// Index for tier filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "tier", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_tenant_tier"),
		},
		// Index for tag filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "tags", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_tenant_tags"),
		},
		// Index for segment filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "segments", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_tenant_segments"),
		},
		// Index for follow-up scheduling
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "next_follow_up_at", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_follow_up"),
		},
		// Index for last contacted filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "last_contacted_at", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_last_contact"),
		},
		// Index for phone number lookup
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "phone_numbers.e164", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_phone"),
		},
		// Index for industry filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "company_info.industry", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_industry"),
		},
		// Index for country filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "addresses.country_code", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_customers_country"),
		},
		// Text index for full-text search
		{
			Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "email.address", Value: "text"},
				{Key: "code", Value: "text"},
				{Key: "company_info.company_name", Value: "text"},
				{Key: "tags", Value: "text"},
				{Key: "notes", Value: "text"},
			},
			Options: options.Index().
				SetName("idx_customers_text_search").
				SetWeights(bson.M{
					"name":                     10,
					"email.address":            8,
					"code":                     8,
					"company_info.company_name": 6,
					"tags":                     4,
					"notes":                    2,
				}).
				SetDefaultLanguage("english"),
		},
		// Index for date range queries
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_customers_created"),
		},
		// Index for updated_at queries
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("idx_customers_updated"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// createContactIndexes creates indexes for the contacts collection.
func (m *IndexManager) createContactIndexes(ctx context.Context) error {
	collection := m.db.Collection(contactsCollection)

	indexes := []mongo.IndexModel{
		// Index for customer_id lookup
		{
			Keys: bson.D{
				{Key: "customer_id", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_customer"),
		},
		// Index for tenant + email lookup
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "email.address", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_tenant_email"),
		},
		// Index for phone number lookup
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "phone_numbers.e164", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_phone"),
		},
		// Index for primary contact lookup
		{
			Keys: bson.D{
				{Key: "customer_id", Value: 1},
				{Key: "is_primary", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_primary"),
		},
		// Index for status filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_status"),
		},
		// Index for role filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "role", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_role"),
		},
		// Index for tag filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "tags", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_tags"),
		},
		// Index for follow-up scheduling
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "next_follow_up_at", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_follow_up"),
		},
		// Index for marketing opt-in
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "opted_out_marketing", Value: 1},
				{Key: "status", Value: 1},
				{Key: "deleted_at", Value: 1},
			},
			Options: options.Index().SetName("idx_contacts_marketing"),
		},
		// Text index for full-text search
		{
			Keys: bson.D{
				{Key: "name.first_name", Value: "text"},
				{Key: "name.last_name", Value: "text"},
				{Key: "email.address", Value: "text"},
				{Key: "job_title", Value: "text"},
				{Key: "department", Value: "text"},
				{Key: "tags", Value: "text"},
			},
			Options: options.Index().
				SetName("idx_contacts_text_search").
				SetWeights(bson.M{
					"name.first_name": 10,
					"name.last_name":  10,
					"email.address":   8,
					"job_title":       4,
					"department":      4,
					"tags":            2,
				}).
				SetDefaultLanguage("english"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// createNoteIndexes creates indexes for the notes collection.
func (m *IndexManager) createNoteIndexes(ctx context.Context) error {
	collection := m.db.Collection(notesCollection)

	indexes := []mongo.IndexModel{
		// Index for customer notes
		{
			Keys: bson.D{
				{Key: "customer_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_notes_customer"),
		},
		// Index for pinned notes
		{
			Keys: bson.D{
				{Key: "customer_id", Value: 1},
				{Key: "is_pinned", Value: -1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_notes_pinned"),
		},
		// Index for tenant notes
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_notes_tenant"),
		},
		// Index for notes by creator
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "created_by", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_notes_creator"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// createActivityIndexes creates indexes for the activities collection.
func (m *IndexManager) createActivityIndexes(ctx context.Context) error {
	collection := m.db.Collection(activitiesCollection)

	indexes := []mongo.IndexModel{
		// Index for customer activities
		{
			Keys: bson.D{
				{Key: "customer_id", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("idx_activities_customer"),
		},
		// Index for contact activities
		{
			Keys: bson.D{
				{Key: "contact_id", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("idx_activities_contact"),
		},
		// Index for deal activities
		{
			Keys: bson.D{
				{Key: "deal_id", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("idx_activities_deal"),
		},
		// Index for lead activities
		{
			Keys: bson.D{
				{Key: "lead_id", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("idx_activities_lead"),
		},
		// Index for user activities
		{
			Keys: bson.D{
				{Key: "performed_by", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("idx_activities_performer"),
		},
		// Index for tenant activities
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("idx_activities_tenant"),
		},
		// Index for activity type filtering
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "type", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("idx_activities_type"),
		},
		// Index for date range queries
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "customer_id", Value: 1},
				{Key: "type", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
			Options: options.Index().SetName("idx_activities_customer_type"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// createSegmentIndexes creates indexes for the segments collection.
func (m *IndexManager) createSegmentIndexes(ctx context.Context) error {
	collection := m.db.Collection(segmentsCollection)

	indexes := []mongo.IndexModel{
		// Unique index on tenant_id + name
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "name", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("idx_segments_tenant_name_unique"),
		},
		// Index for tenant segments
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "is_active", Value: 1},
			},
			Options: options.Index().SetName("idx_segments_tenant_active"),
		},
		// Index for segment type
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "type", Value: 1},
			},
			Options: options.Index().SetName("idx_segments_type"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// createImportIndexes creates indexes for the imports collection.
func (m *IndexManager) createImportIndexes(ctx context.Context) error {
	importCollection := m.db.Collection(importsCollection)
	errorCollection := m.db.Collection(importErrorsCollection)

	importIndexes := []mongo.IndexModel{
		// Index for tenant imports
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_imports_tenant"),
		},
		// Index for import status
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "created_at", Value: 1},
			},
			Options: options.Index().SetName("idx_imports_status"),
		},
		// Index for cleanup of old imports
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "completed_at", Value: 1},
			},
			Options: options.Index().SetName("idx_imports_cleanup"),
		},
	}

	errorIndexes := []mongo.IndexModel{
		// Index for import errors
		{
			Keys: bson.D{
				{Key: "import_id", Value: 1},
				{Key: "row_number", Value: 1},
			},
			Options: options.Index().SetName("idx_import_errors_import"),
		},
	}

	if _, err := importCollection.Indexes().CreateMany(ctx, importIndexes); err != nil {
		return err
	}

	_, err := errorCollection.Indexes().CreateMany(ctx, errorIndexes)
	return err
}

// createOutboxIndexes creates indexes for the outbox collection.
func (m *IndexManager) createOutboxIndexes(ctx context.Context) error {
	collection := m.db.Collection(outboxCollection)

	indexes := []mongo.IndexModel{
		// Index for pending entries
		{
			Keys: bson.D{
				{Key: "processed_at", Value: 1},
				{Key: "created_at", Value: 1},
			},
			Options: options.Index().SetName("idx_outbox_pending"),
		},
		// Index for failed entries
		{
			Keys: bson.D{
				{Key: "processed_at", Value: 1},
				{Key: "failed_at", Value: 1},
				{Key: "retry_count", Value: 1},
			},
			Options: options.Index().SetName("idx_outbox_failed"),
		},
		// Index for event type
		{
			Keys: bson.D{
				{Key: "event_type", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_outbox_event_type"),
		},
		// Index for aggregate ID
		{
			Keys: bson.D{
				{Key: "aggregate_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_outbox_aggregate"),
		},
		// Index for cleanup of old processed entries
		{
			Keys: bson.D{
				{Key: "processed_at", Value: 1},
			},
			Options: options.Index().
				SetName("idx_outbox_cleanup").
				SetExpireAfterSeconds(int32(7 * 24 * 60 * 60)), // 7 days TTL
		},
		// Index for locked entries
		{
			Keys: bson.D{
				{Key: "locked_until", Value: 1},
			},
			Options: options.Index().SetName("idx_outbox_locked"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// DropAllIndexes drops all indexes (useful for testing).
func (m *IndexManager) DropAllIndexes(ctx context.Context) error {
	collections := []string{
		customersCollection,
		contactsCollection,
		notesCollection,
		activitiesCollection,
		segmentsCollection,
		importsCollection,
		importErrorsCollection,
		outboxCollection,
	}

	for _, collName := range collections {
		coll := m.db.Collection(collName)
		if _, err := coll.Indexes().DropAll(ctx); err != nil {
			return fmt.Errorf("failed to drop indexes for %s: %w", collName, err)
		}
	}

	return nil
}

// EnsureIndexes creates indexes if they don't exist (idempotent).
func (m *IndexManager) EnsureIndexes(ctx context.Context) error {
	return m.CreateAllIndexes(ctx)
}

// ValidateIndexes validates that all required indexes exist.
func (m *IndexManager) ValidateIndexes(ctx context.Context) error {
	collections := map[string][]string{
		customersCollection:  {"idx_customers_tenant_code_unique", "idx_customers_text_search"},
		contactsCollection:   {"idx_contacts_customer", "idx_contacts_text_search"},
		notesCollection:      {"idx_notes_customer"},
		activitiesCollection: {"idx_activities_customer"},
		segmentsCollection:   {"idx_segments_tenant_name_unique"},
		importsCollection:    {"idx_imports_tenant"},
		outboxCollection:     {"idx_outbox_pending"},
	}

	for collName, requiredIndexes := range collections {
		coll := m.db.Collection(collName)
		cursor, err := coll.Indexes().List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list indexes for %s: %w", collName, err)
		}
		defer cursor.Close(ctx)

		existingIndexes := make(map[string]bool)
		for cursor.Next(ctx) {
			var idx struct {
				Name string `bson:"name"`
			}
			if err := cursor.Decode(&idx); err != nil {
				continue
			}
			existingIndexes[idx.Name] = true
		}

		for _, requiredIdx := range requiredIndexes {
			if !existingIndexes[requiredIdx] {
				return fmt.Errorf("missing required index %s on collection %s", requiredIdx, collName)
			}
		}
	}

	return nil
}

// GetIndexStats returns statistics about indexes for monitoring.
func (m *IndexManager) GetIndexStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	collections := []string{
		customersCollection,
		contactsCollection,
		notesCollection,
		activitiesCollection,
		segmentsCollection,
		importsCollection,
		outboxCollection,
	}

	for _, collName := range collections {
		coll := m.db.Collection(collName)
		cursor, err := coll.Indexes().List(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list indexes for %s: %w", collName, err)
		}

		var indexes []bson.M
		if err := cursor.All(ctx, &indexes); err != nil {
			cursor.Close(ctx)
			return nil, fmt.Errorf("failed to decode indexes for %s: %w", collName, err)
		}
		cursor.Close(ctx)

		collStats := map[string]interface{}{
			"index_count": len(indexes),
			"indexes":     indexes,
		}
		stats[collName] = collStats
	}

	stats["timestamp"] = time.Now().UTC()
	return stats, nil
}
