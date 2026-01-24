// Package domain contains the domain layer for the Customer service.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CustomerRepository defines the interface for customer persistence.
type CustomerRepository interface {
	// Create creates a new customer.
	Create(ctx context.Context, customer *Customer) error

	// Update updates an existing customer.
	Update(ctx context.Context, customer *Customer) error

	// Delete soft deletes a customer.
	Delete(ctx context.Context, id uuid.UUID) error

	// HardDelete permanently deletes a customer.
	HardDelete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a customer by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Customer, error)

	// FindByCode finds a customer by code.
	FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Customer, error)

	// FindByEmail finds a customer by email.
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*Customer, error)

	// List lists customers with filtering and pagination.
	List(ctx context.Context, filter CustomerFilter) (*CustomerList, error)

	// Search performs full-text search on customers.
	Search(ctx context.Context, tenantID uuid.UUID, query string, filter CustomerFilter) (*CustomerList, error)

	// FindByOwner finds customers assigned to an owner.
	FindByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, filter CustomerFilter) (*CustomerList, error)

	// FindByTag finds customers with a specific tag.
	FindByTag(ctx context.Context, tenantID uuid.UUID, tag string, filter CustomerFilter) (*CustomerList, error)

	// FindBySegment finds customers in a segment.
	FindBySegment(ctx context.Context, tenantID, segmentID uuid.UUID, filter CustomerFilter) (*CustomerList, error)

	// FindByStatus finds customers by status.
	FindByStatus(ctx context.Context, tenantID uuid.UUID, status CustomerStatus, filter CustomerFilter) (*CustomerList, error)

	// FindDuplicates finds potential duplicate customers.
	FindDuplicates(ctx context.Context, tenantID uuid.UUID, email, phone, name string) ([]*Customer, error)

	// CountByTenant counts customers for a tenant.
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)

	// CountByStatus counts customers by status.
	CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[CustomerStatus]int64, error)

	// GetStats gets customer statistics.
	GetStats(ctx context.Context, tenantID uuid.UUID) (*CustomerStats, error)

	// BulkCreate creates multiple customers.
	BulkCreate(ctx context.Context, customers []*Customer) error

	// BulkUpdate updates multiple customers.
	BulkUpdate(ctx context.Context, ids []uuid.UUID, updates map[string]interface{}) error

	// BulkDelete soft deletes multiple customers.
	BulkDelete(ctx context.Context, ids []uuid.UUID) error

	// Exists checks if a customer exists.
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// ExistsByCode checks if a customer code exists.
	ExistsByCode(ctx context.Context, tenantID uuid.UUID, code string) (bool, error)

	// ExistsByEmail checks if a customer email exists.
	ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email string) (bool, error)

	// GetVersion gets the current version for optimistic locking.
	GetVersion(ctx context.Context, id uuid.UUID) (int, error)

	// FindNeedingFollowUp finds customers needing follow-up.
	FindNeedingFollowUp(ctx context.Context, tenantID uuid.UUID, before time.Time) ([]*Customer, error)

	// FindInactive finds inactive customers.
	FindInactive(ctx context.Context, tenantID uuid.UUID, lastContactBefore time.Time) ([]*Customer, error)

	// FindRecentlyCreated finds recently created customers.
	FindRecentlyCreated(ctx context.Context, tenantID uuid.UUID, since time.Time, limit int) ([]*Customer, error)

	// FindRecentlyUpdated finds recently updated customers.
	FindRecentlyUpdated(ctx context.Context, tenantID uuid.UUID, since time.Time, limit int) ([]*Customer, error)
}

// CustomerFilter defines filtering options for customer queries.
type CustomerFilter struct {
	TenantID       *uuid.UUID       `json:"tenant_id,omitempty"`
	IDs            []uuid.UUID      `json:"ids,omitempty"`
	Codes          []string         `json:"codes,omitempty"`
	Types          []CustomerType   `json:"types,omitempty"`
	Statuses       []CustomerStatus `json:"statuses,omitempty"`
	Tiers          []CustomerTier   `json:"tiers,omitempty"`
	Sources        []CustomerSource `json:"sources,omitempty"`
	Tags           []string         `json:"tags,omitempty"`
	OwnerIDs       []uuid.UUID      `json:"owner_ids,omitempty"`
	SegmentIDs     []uuid.UUID      `json:"segment_ids,omitempty"`
	Industries     []Industry       `json:"industries,omitempty"`
	Countries      []string         `json:"country_codes,omitempty"`
	Query          string           `json:"query,omitempty"` // Full-text search
	Email          string           `json:"email,omitempty"`
	Phone          string           `json:"phone,omitempty"`
	HasDeals       *bool            `json:"has_deals,omitempty"`
	HasOpenDeals   *bool            `json:"has_open_deals,omitempty"`
	CreatedAfter   *time.Time       `json:"created_after,omitempty"`
	CreatedBefore  *time.Time       `json:"created_before,omitempty"`
	UpdatedAfter   *time.Time       `json:"updated_after,omitempty"`
	UpdatedBefore  *time.Time       `json:"updated_before,omitempty"`
	LastContactAfter  *time.Time    `json:"last_contact_after,omitempty"`
	LastContactBefore *time.Time    `json:"last_contact_before,omitempty"`
	IncludeDeleted bool             `json:"include_deleted,omitempty"`
	Offset         int              `json:"offset"`
	Limit          int              `json:"limit"`
	SortBy         string           `json:"sort_by,omitempty"`
	SortOrder      string           `json:"sort_order,omitempty"` // "asc" or "desc"
}

// CustomerList represents a paginated list of customers.
type CustomerList struct {
	Customers  []*Customer `json:"customers"`
	Total      int64       `json:"total"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
	HasMore    bool        `json:"has_more"`
}

// ContactRepository defines the interface for contact persistence.
type ContactRepository interface {
	// FindByID finds a contact by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Contact, error)

	// FindByCustomerID finds all contacts for a customer.
	FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*Contact, error)

	// FindByEmail finds contacts by email.
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]*Contact, error)

	// Search searches contacts.
	Search(ctx context.Context, tenantID uuid.UUID, query string, filter ContactFilter) (*ContactList, error)

	// CountByCustomer counts contacts for a customer.
	CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error)
}

// ContactFilter defines filtering options for contact queries.
type ContactFilter struct {
	TenantID    *uuid.UUID      `json:"tenant_id,omitempty"`
	CustomerIDs []uuid.UUID     `json:"customer_ids,omitempty"`
	Statuses    []ContactStatus `json:"statuses,omitempty"`
	Roles       []ContactRole   `json:"roles,omitempty"`
	IsPrimary   *bool           `json:"is_primary,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Query       string          `json:"query,omitempty"`
	Offset      int             `json:"offset"`
	Limit       int             `json:"limit"`
	SortBy      string          `json:"sort_by,omitempty"`
	SortOrder   string          `json:"sort_order,omitempty"`
}

// ContactList represents a paginated list of contacts.
type ContactList struct {
	Contacts []*Contact `json:"contacts"`
	Total    int64      `json:"total"`
	Offset   int        `json:"offset"`
	Limit    int        `json:"limit"`
	HasMore  bool       `json:"has_more"`
}

// NoteRepository defines the interface for customer notes.
type NoteRepository interface {
	// Create creates a new note.
	Create(ctx context.Context, note *Note) error

	// Update updates a note.
	Update(ctx context.Context, note *Note) error

	// Delete deletes a note.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a note by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Note, error)

	// FindByCustomerID finds all notes for a customer.
	FindByCustomerID(ctx context.Context, customerID uuid.UUID, filter NoteFilter) ([]*Note, error)

	// CountByCustomer counts notes for a customer.
	CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error)
}

// Note represents a customer note.
type Note struct {
	BaseEntity
	CustomerID uuid.UUID  `json:"customer_id" bson:"customer_id"`
	TenantID   uuid.UUID  `json:"tenant_id" bson:"tenant_id"`
	Content    string     `json:"content" bson:"content"`
	IsPinned   bool       `json:"is_pinned" bson:"is_pinned"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty" bson:"created_by,omitempty"`
	UpdatedBy  *uuid.UUID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
}

// NoteFilter defines filtering options for notes.
type NoteFilter struct {
	IsPinned *bool  `json:"is_pinned,omitempty"`
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
}

// ActivityRepository defines the interface for customer activities.
type ActivityRepository interface {
	// Create creates a new activity.
	Create(ctx context.Context, activity *Activity) error

	// Update updates an activity.
	Update(ctx context.Context, activity *Activity) error

	// Delete deletes an activity.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID finds an activity by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Activity, error)

	// FindByCustomerID finds all activities for a customer.
	FindByCustomerID(ctx context.Context, customerID uuid.UUID, filter ActivityFilter) (*ActivityList, error)

	// FindByContactID finds all activities for a contact.
	FindByContactID(ctx context.Context, contactID uuid.UUID, filter ActivityFilter) (*ActivityList, error)

	// FindByPerformedBy finds activities performed by a user.
	FindByPerformedBy(ctx context.Context, userID uuid.UUID, filter ActivityFilter) (*ActivityList, error)

	// CountByCustomer counts activities for a customer.
	CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error)

	// GetActivitySummary gets activity summary for a customer.
	GetActivitySummary(ctx context.Context, customerID uuid.UUID) (map[ActivityType]int, error)
}

// Activity represents a customer activity.
type Activity struct {
	BaseEntity
	CustomerID   uuid.UUID    `json:"customer_id" bson:"customer_id"`
	ContactID    *uuid.UUID   `json:"contact_id,omitempty" bson:"contact_id,omitempty"`
	TenantID     uuid.UUID    `json:"tenant_id" bson:"tenant_id"`
	Type         ActivityType `json:"type" bson:"type"`
	Subject      string       `json:"subject" bson:"subject"`
	Description  string       `json:"description,omitempty" bson:"description,omitempty"`
	OccurredAt   time.Time    `json:"occurred_at" bson:"occurred_at"`
	Duration     *int         `json:"duration,omitempty" bson:"duration,omitempty"` // Minutes
	Outcome      string       `json:"outcome,omitempty" bson:"outcome,omitempty"`
	PerformedBy  *uuid.UUID   `json:"performed_by,omitempty" bson:"performed_by,omitempty"`
	DealID       *uuid.UUID   `json:"deal_id,omitempty" bson:"deal_id,omitempty"`
	LeadID       *uuid.UUID   `json:"lead_id,omitempty" bson:"lead_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// ActivityFilter defines filtering options for activities.
type ActivityFilter struct {
	Types       []ActivityType `json:"types,omitempty"`
	PerformedBy []uuid.UUID    `json:"performed_by,omitempty"`
	After       *time.Time     `json:"after,omitempty"`
	Before      *time.Time     `json:"before,omitempty"`
	Offset      int            `json:"offset"`
	Limit       int            `json:"limit"`
	SortBy      string         `json:"sort_by,omitempty"`
	SortOrder   string         `json:"sort_order,omitempty"`
}

// ActivityList represents a paginated list of activities.
type ActivityList struct {
	Activities []*Activity `json:"activities"`
	Total      int64       `json:"total"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
	HasMore    bool        `json:"has_more"`
}

// SegmentRepository defines the interface for customer segments.
type SegmentRepository interface {
	// Create creates a new segment.
	Create(ctx context.Context, segment *Segment) error

	// Update updates a segment.
	Update(ctx context.Context, segment *Segment) error

	// Delete deletes a segment.
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID finds a segment by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Segment, error)

	// FindByTenantID finds all segments for a tenant.
	FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*Segment, error)

	// FindByName finds a segment by name.
	FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*Segment, error)

	// GetCustomerCount gets the number of customers in a segment.
	GetCustomerCount(ctx context.Context, segmentID uuid.UUID) (int64, error)

	// RefreshDynamic refreshes dynamic segment membership.
	RefreshDynamic(ctx context.Context, segmentID uuid.UUID) error
}

// SegmentType represents the type of segment.
type SegmentType string

const (
	SegmentTypeStatic  SegmentType = "static"  // Manually managed
	SegmentTypeDynamic SegmentType = "dynamic" // Rule-based
)

// Segment represents a customer segment.
type Segment struct {
	BaseEntity
	TenantID      uuid.UUID                `json:"tenant_id" bson:"tenant_id"`
	Name          string                   `json:"name" bson:"name"`
	Description   string                   `json:"description,omitempty" bson:"description,omitempty"`
	Type          SegmentType              `json:"type" bson:"type"`
	Criteria      *SegmentCriteria         `json:"criteria,omitempty" bson:"criteria,omitempty"`
	CustomerCount int64                    `json:"customer_count" bson:"customer_count"`
	Color         string                   `json:"color,omitempty" bson:"color,omitempty"`
	IsActive      bool                     `json:"is_active" bson:"is_active"`
	CreatedBy     *uuid.UUID               `json:"created_by,omitempty" bson:"created_by,omitempty"`
	LastRefreshed *time.Time               `json:"last_refreshed,omitempty" bson:"last_refreshed,omitempty"`
}

// SegmentCriteria defines rules for dynamic segments.
type SegmentCriteria struct {
	Rules    []SegmentRule `json:"rules" bson:"rules"`
	Operator string        `json:"operator" bson:"operator"` // "and" or "or"
}

// SegmentRule defines a single segment rule.
type SegmentRule struct {
	Field    string      `json:"field" bson:"field"`
	Operator string      `json:"operator" bson:"operator"`
	Value    interface{} `json:"value" bson:"value"`
}

// ImportRepository defines the interface for import operations.
type ImportRepository interface {
	// CreateImport creates a new import record.
	CreateImport(ctx context.Context, imp *Import) error

	// UpdateImport updates an import record.
	UpdateImport(ctx context.Context, imp *Import) error

	// FindImportByID finds an import by ID.
	FindImportByID(ctx context.Context, id uuid.UUID) (*Import, error)

	// FindImportsByTenant finds imports for a tenant.
	FindImportsByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*Import, error)

	// CreateImportError creates an import error record.
	CreateImportError(ctx context.Context, err *ImportError) error

	// FindImportErrors finds errors for an import.
	FindImportErrors(ctx context.Context, importID uuid.UUID) ([]*ImportError, error)
}

// ImportStatus represents the status of an import.
type ImportStatus string

const (
	ImportStatusPending    ImportStatus = "pending"
	ImportStatusProcessing ImportStatus = "processing"
	ImportStatusCompleted  ImportStatus = "completed"
	ImportStatusFailed     ImportStatus = "failed"
	ImportStatusCancelled  ImportStatus = "cancelled"
)

// Import represents an import operation.
type Import struct {
	BaseEntity
	TenantID      uuid.UUID    `json:"tenant_id" bson:"tenant_id"`
	FileName      string       `json:"file_name" bson:"file_name"`
	FileSize      int64        `json:"file_size" bson:"file_size"`
	FileType      string       `json:"file_type" bson:"file_type"`
	Status        ImportStatus `json:"status" bson:"status"`
	TotalRows     int          `json:"total_rows" bson:"total_rows"`
	ProcessedRows int          `json:"processed_rows" bson:"processed_rows"`
	SuccessRows   int          `json:"success_rows" bson:"success_rows"`
	FailedRows    int          `json:"failed_rows" bson:"failed_rows"`
	DuplicateRows int          `json:"duplicate_rows" bson:"duplicate_rows"`
	ErrorMessage  string       `json:"error_message,omitempty" bson:"error_message,omitempty"`
	StartedAt     *time.Time   `json:"started_at,omitempty" bson:"started_at,omitempty"`
	CompletedAt   *time.Time   `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	CreatedBy     *uuid.UUID   `json:"created_by,omitempty" bson:"created_by,omitempty"`
	Options       ImportOptions `json:"options" bson:"options"`
}

// ImportOptions defines import configuration.
type ImportOptions struct {
	SkipDuplicates   bool              `json:"skip_duplicates" bson:"skip_duplicates"`
	UpdateExisting   bool              `json:"update_existing" bson:"update_existing"`
	DefaultOwner     *uuid.UUID        `json:"default_owner,omitempty" bson:"default_owner,omitempty"`
	DefaultStatus    CustomerStatus    `json:"default_status" bson:"default_status"`
	DefaultType      CustomerType      `json:"default_type" bson:"default_type"`
	DefaultSource    CustomerSource    `json:"default_source" bson:"default_source"`
	DefaultTags      []string          `json:"default_tags,omitempty" bson:"default_tags,omitempty"`
	FieldMapping     map[string]string `json:"field_mapping" bson:"field_mapping"`
}

// ImportError represents an import error.
type ImportError struct {
	BaseEntity
	ImportID   uuid.UUID `json:"import_id" bson:"import_id"`
	RowNumber  int       `json:"row_number" bson:"row_number"`
	Field      string    `json:"field,omitempty" bson:"field,omitempty"`
	Value      string    `json:"value,omitempty" bson:"value,omitempty"`
	Error      string    `json:"error" bson:"error"`
	ErrorCode  string    `json:"error_code" bson:"error_code"`
	RawData    string    `json:"raw_data,omitempty" bson:"raw_data,omitempty"`
}

// OutboxRepository defines the interface for the transactional outbox pattern.
type OutboxRepository interface {
	// Create creates an outbox entry.
	Create(ctx context.Context, entry *OutboxEntry) error

	// MarkAsProcessed marks an entry as processed.
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error

	// MarkAsFailed marks an entry as failed.
	MarkAsFailed(ctx context.Context, id uuid.UUID, err string) error

	// FindPending finds pending outbox entries.
	FindPending(ctx context.Context, limit int) ([]*OutboxEntry, error)

	// FindFailed finds failed outbox entries.
	FindFailed(ctx context.Context, limit int) ([]*OutboxEntry, error)

	// DeleteOld deletes old processed entries.
	DeleteOld(ctx context.Context, before time.Time) error
}

// OutboxEntry represents an outbox entry.
type OutboxEntry struct {
	ID          uuid.UUID  `json:"id" bson:"_id"`
	TenantID    uuid.UUID  `json:"tenant_id" bson:"tenant_id"`
	EventType   string     `json:"event_type" bson:"event_type"`
	AggregateID uuid.UUID  `json:"aggregate_id" bson:"aggregate_id"`
	Payload     []byte     `json:"payload" bson:"payload"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty" bson:"failed_at,omitempty"`
	Error       string     `json:"error,omitempty" bson:"error,omitempty"`
	RetryCount  int        `json:"retry_count" bson:"retry_count"`
}

// UnitOfWork defines the interface for transaction management.
type UnitOfWork interface {
	// Begin begins a new transaction.
	Begin(ctx context.Context) (context.Context, error)

	// Commit commits the transaction.
	Commit(ctx context.Context) error

	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error

	// Customers returns the customer repository.
	Customers() CustomerRepository

	// Contacts returns the contact repository.
	Contacts() ContactRepository

	// Notes returns the note repository.
	Notes() NoteRepository

	// Activities returns the activity repository.
	Activities() ActivityRepository

	// Segments returns the segment repository.
	Segments() SegmentRepository

	// Outbox returns the outbox repository.
	Outbox() OutboxRepository
}
