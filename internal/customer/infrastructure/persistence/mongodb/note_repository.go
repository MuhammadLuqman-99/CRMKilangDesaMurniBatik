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
	notesCollection = "customer_notes"
)

// NoteRepository implements domain.NoteRepository using MongoDB.
type NoteRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewNoteRepository creates a new NoteRepository.
func NewNoteRepository(db *mongo.Database) *NoteRepository {
	return &NoteRepository{
		db:         db,
		collection: db.Collection(notesCollection),
	}
}

// Create creates a new note.
func (r *NoteRepository) Create(ctx context.Context, note *domain.Note) error {
	note.CreatedAt = time.Now().UTC()
	note.UpdatedAt = note.CreatedAt
	note.Version = 1

	_, err := r.collection.InsertOne(ctx, note)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

// Update updates a note.
func (r *NoteRepository) Update(ctx context.Context, note *domain.Note) error {
	note.UpdatedAt = time.Now().UTC()
	previousVersion := note.Version
	note.Version++

	filter := bson.M{
		"_id":     note.ID,
		"version": previousVersion,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, note)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	if result.MatchedCount == 0 {
		var existing domain.Note
		err := r.collection.FindOne(ctx, bson.M{"_id": note.ID}).Decode(&existing)
		if err == mongo.ErrNoDocuments {
			return domain.ErrNoteNotFound
		}
		return domain.ErrVersionConflict
	}

	return nil
}

// Delete deletes a note.
func (r *NoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrNoteNotFound
	}

	return nil
}

// FindByID finds a note by ID.
func (r *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	var note domain.Note
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&note)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrNoteNotFound
		}
		return nil, fmt.Errorf("failed to find note: %w", err)
	}

	return &note, nil
}

// FindByCustomerID finds all notes for a customer.
func (r *NoteRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID, filter domain.NoteFilter) ([]*domain.Note, error) {
	mongoFilter := bson.M{"customer_id": customerID}

	if filter.IsPinned != nil {
		mongoFilter["is_pinned"] = *filter.IsPinned
	}

	findOpts := options.Find().
		SetSort(bson.D{
			{Key: "is_pinned", Value: -1},
			{Key: "created_at", Value: -1},
		})

	if filter.Limit > 0 {
		findOpts.SetSkip(int64(filter.Offset)).SetLimit(int64(filter.Limit))
	}

	cursor, err := r.collection.Find(ctx, mongoFilter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to find notes by customer: %w", err)
	}
	defer cursor.Close(ctx)

	var notes []*domain.Note
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, fmt.Errorf("failed to decode notes: %w", err)
	}

	return notes, nil
}

// CountByCustomer counts notes for a customer.
func (r *NoteRepository) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return 0, fmt.Errorf("failed to count notes: %w", err)
	}
	return int(count), nil
}

// DeleteByCustomer deletes all notes for a customer.
func (r *NoteRepository) DeleteByCustomer(ctx context.Context, customerID uuid.UUID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"customer_id": customerID})
	if err != nil {
		return fmt.Errorf("failed to delete notes by customer: %w", err)
	}
	return nil
}

// Pin pins a note.
func (r *NoteRepository) Pin(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_pinned":  true,
			"updated_at": time.Now().UTC(),
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to pin note: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNoteNotFound
	}

	return nil
}

// Unpin unpins a note.
func (r *NoteRepository) Unpin(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_pinned":  false,
			"updated_at": time.Now().UTC(),
		},
		"$inc": bson.M{"version": 1},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to unpin note: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrNoteNotFound
	}

	return nil
}

// FindPinned finds all pinned notes for a customer.
func (r *NoteRepository) FindPinned(ctx context.Context, customerID uuid.UUID) ([]*domain.Note, error) {
	filter := bson.M{
		"customer_id": customerID,
		"is_pinned":   true,
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("failed to find pinned notes: %w", err)
	}
	defer cursor.Close(ctx)

	var notes []*domain.Note
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, fmt.Errorf("failed to decode notes: %w", err)
	}

	return notes, nil
}
