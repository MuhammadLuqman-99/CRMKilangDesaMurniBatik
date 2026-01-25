// Package ports defines the application layer interfaces (ports) for the Customer service.
// These interfaces follow the hexagonal architecture pattern (ports and adapters).
package ports

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// Cache errors
var (
	ErrCacheMiss = errors.New("cache miss")
)

// ============================================================================
// Event Publishing Ports
// ============================================================================

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	// Publish publishes a single domain event.
	Publish(ctx context.Context, event domain.DomainEvent) error

	// PublishAll publishes multiple domain events.
	PublishAll(ctx context.Context, events []domain.DomainEvent) error

	// PublishAsync publishes events asynchronously.
	PublishAsync(ctx context.Context, events []domain.DomainEvent) error
}

// EventSubscriber defines the interface for subscribing to domain events.
type EventSubscriber interface {
	// Subscribe subscribes to events of a specific type.
	Subscribe(ctx context.Context, eventType string, handler EventHandler) error

	// SubscribeAll subscribes to all events.
	SubscribeAll(ctx context.Context, handler EventHandler) error

	// Unsubscribe unsubscribes from events.
	Unsubscribe(ctx context.Context, eventType string) error
}

// EventHandler defines the interface for handling domain events.
type EventHandler interface {
	// Handle handles a domain event.
	Handle(ctx context.Context, event domain.DomainEvent) error

	// EventTypes returns the event types this handler handles.
	EventTypes() []string
}

// ============================================================================
// Search Ports
// ============================================================================

// SearchIndex defines the interface for full-text search indexing.
type SearchIndex interface {
	// IndexCustomer indexes a customer for full-text search.
	IndexCustomer(ctx context.Context, customer *domain.Customer) error

	// IndexContact indexes a contact for full-text search.
	IndexContact(ctx context.Context, contact *domain.Contact) error

	// RemoveCustomer removes a customer from the search index.
	RemoveCustomer(ctx context.Context, id uuid.UUID) error

	// RemoveContact removes a contact from the search index.
	RemoveContact(ctx context.Context, id uuid.UUID) error

	// SearchCustomers performs full-text search on customers.
	SearchCustomers(ctx context.Context, tenantID uuid.UUID, query string, options SearchOptions) (*SearchResult, error)

	// SearchContacts performs full-text search on contacts.
	SearchContacts(ctx context.Context, tenantID uuid.UUID, query string, options SearchOptions) (*SearchResult, error)

	// Reindex triggers a full reindex of all data.
	Reindex(ctx context.Context, tenantID uuid.UUID) error
}

// SearchOptions contains search configuration options.
type SearchOptions struct {
	Offset        int
	Limit         int
	Fields        []string
	Filters       map[string]interface{}
	SortBy        string
	SortOrder     string
	Highlight     bool
	FuzzyMatching bool
	MinScore      float64
}

// SearchResult contains search results with metadata.
type SearchResult struct {
	TotalHits int64
	MaxScore  float64
	Hits      []SearchHit
	Took      time.Duration
}

// SearchHit represents a single search hit.
type SearchHit struct {
	ID         uuid.UUID
	Score      float64
	Highlights map[string][]string
	Source     map[string]interface{}
}

// ============================================================================
// Cache Ports
// ============================================================================

// CacheService defines the interface for caching operations.
type CacheService interface {
	// Get retrieves a value from cache.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with optional TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from cache.
	Delete(ctx context.Context, key string) error

	// DeletePattern removes all keys matching a pattern.
	DeletePattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists in cache.
	Exists(ctx context.Context, key string) (bool, error)

	// GetOrSet gets a value or sets it if not present.
	GetOrSet(ctx context.Context, key string, ttl time.Duration, getter func() ([]byte, error)) ([]byte, error)

	// Invalidate invalidates cache for a specific entity.
	Invalidate(ctx context.Context, entityType string, entityID uuid.UUID) error

	// InvalidateByTenant invalidates all cache for a tenant.
	InvalidateByTenant(ctx context.Context, tenantID uuid.UUID) error
}

// ============================================================================
// File Storage Ports
// ============================================================================

// FileStorage defines the interface for file storage operations.
type FileStorage interface {
	// Upload uploads a file and returns its URL.
	Upload(ctx context.Context, tenantID uuid.UUID, filename string, content io.Reader, contentType string) (string, error)

	// Download downloads a file.
	Download(ctx context.Context, url string) (io.ReadCloser, error)

	// Delete deletes a file.
	Delete(ctx context.Context, url string) error

	// GetSignedURL generates a signed URL for temporary access.
	GetSignedURL(ctx context.Context, url string, expiry time.Duration) (string, error)
}

// ============================================================================
// Export/Import Ports
// ============================================================================

// ExportFormat represents supported export formats.
type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatXLSX ExportFormat = "xlsx"
	ExportFormatJSON ExportFormat = "json"
	ExportFormatPDF  ExportFormat = "pdf"
)

// ExportService defines the interface for data export operations.
type ExportService interface {
	// ExportCustomers exports customers to the specified format.
	ExportCustomers(ctx context.Context, tenantID uuid.UUID, customers []*domain.Customer, format ExportFormat, fields []string) ([]byte, error)

	// ExportContacts exports contacts to the specified format.
	ExportContacts(ctx context.Context, tenantID uuid.UUID, contacts []*domain.Contact, format ExportFormat, fields []string) ([]byte, error)

	// StreamExport streams large exports to a writer.
	StreamExport(ctx context.Context, tenantID uuid.UUID, query string, format ExportFormat, writer io.Writer) error
}

// ImportService defines the interface for data import operations.
type ImportService interface {
	// ParseCustomers parses customer data from various formats.
	ParseCustomers(ctx context.Context, data io.Reader, format ExportFormat, fieldMapping map[string]string) ([]*CustomerImportRow, error)

	// ParseContacts parses contact data from various formats.
	ParseContacts(ctx context.Context, data io.Reader, format ExportFormat, fieldMapping map[string]string) ([]*ContactImportRow, error)

	// ValidateImportData validates imported data before processing.
	ValidateImportData(ctx context.Context, rows interface{}) ([]ImportValidationError, error)
}

// CustomerImportRow represents a parsed customer row for import.
type CustomerImportRow struct {
	RowNumber    int
	Name         string
	Type         string
	Email        string
	Phone        string
	Website      string
	Address      ImportAddress
	CompanyInfo  ImportCompanyInfo
	Tags         []string
	Notes        string
	CustomFields map[string]interface{}
	RawData      map[string]string
	Errors       []string
	Warnings     []string
}

// ContactImportRow represents a parsed contact row for import.
type ContactImportRow struct {
	RowNumber    int
	FirstName    string
	LastName     string
	Email        string
	Phone        string
	JobTitle     string
	Department   string
	Role         string
	CustomerRef  string
	Tags         []string
	Notes        string
	CustomFields map[string]interface{}
	RawData      map[string]string
	Errors       []string
	Warnings     []string
}

// ImportAddress represents address data in import.
type ImportAddress struct {
	Line1       string
	Line2       string
	City        string
	State       string
	PostalCode  string
	CountryCode string
}

// ImportCompanyInfo represents company info in import.
type ImportCompanyInfo struct {
	LegalName          string
	RegistrationNumber string
	TaxID              string
	Industry           string
	Size               string
}

// ImportValidationError represents an import validation error.
type ImportValidationError struct {
	RowNumber int
	Field     string
	Value     string
	Error     string
	Severity  string // "error" or "warning"
}

// ============================================================================
// Notification Ports
// ============================================================================

// NotificationService defines the interface for sending notifications.
type NotificationService interface {
	// SendEmail sends an email notification.
	SendEmail(ctx context.Context, to, subject, body string, opts EmailOptions) error

	// SendSMS sends an SMS notification.
	SendSMS(ctx context.Context, to, message string) error

	// SendPushNotification sends a push notification.
	SendPushNotification(ctx context.Context, userID uuid.UUID, title, body string, data map[string]interface{}) error

	// SendInAppNotification sends an in-app notification.
	SendInAppNotification(ctx context.Context, userID uuid.UUID, notification InAppNotification) error
}

// EmailOptions contains email sending options.
type EmailOptions struct {
	CC          []string
	BCC         []string
	ReplyTo     string
	Attachments []EmailAttachment
	Template    string
	TemplateData map[string]interface{}
	Priority    string
}

// EmailAttachment represents an email attachment.
type EmailAttachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// InAppNotification represents an in-app notification.
type InAppNotification struct {
	ID        uuid.UUID
	Type      string
	Title     string
	Body      string
	Link      string
	Data      map[string]interface{}
	Priority  string
	ExpiresAt *time.Time
}

// ============================================================================
// Audit Ports
// ============================================================================

// AuditLogger defines the interface for audit logging.
type AuditLogger interface {
	// LogAction logs a user action.
	LogAction(ctx context.Context, entry AuditEntry) error

	// GetAuditLog retrieves audit log entries.
	GetAuditLog(ctx context.Context, filter AuditFilter) ([]AuditEntry, error)

	// GetEntityHistory retrieves the audit history for a specific entity.
	GetEntityHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]AuditEntry, error)
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Action       string
	EntityType   string
	EntityID     uuid.UUID
	OldValue     map[string]interface{}
	NewValue     map[string]interface{}
	Changes      []FieldChange
	IPAddress    string
	UserAgent    string
	Metadata     map[string]interface{}
	Timestamp    time.Time
}

// FieldChange represents a single field change in audit.
type FieldChange struct {
	Field    string
	OldValue interface{}
	NewValue interface{}
}

// AuditFilter contains filters for querying audit logs.
type AuditFilter struct {
	TenantID    uuid.UUID
	UserID      *uuid.UUID
	EntityType  string
	EntityID    *uuid.UUID
	Actions     []string
	StartTime   *time.Time
	EndTime     *time.Time
	Offset      int
	Limit       int
}

// ============================================================================
// Duplicate Detection Ports
// ============================================================================

// DuplicateDetector defines the interface for duplicate detection.
type DuplicateDetector interface {
	// FindDuplicateCustomers finds potential duplicate customers.
	FindDuplicateCustomers(ctx context.Context, tenantID uuid.UUID, customer *domain.Customer) ([]DuplicateMatch, error)

	// FindDuplicateContacts finds potential duplicate contacts.
	FindDuplicateContacts(ctx context.Context, tenantID uuid.UUID, contact *domain.Contact) ([]DuplicateMatch, error)

	// CalculateSimilarity calculates similarity between two entities.
	CalculateSimilarity(ctx context.Context, entity1, entity2 interface{}) (float64, []string, error)
}

// DuplicateMatch represents a potential duplicate match.
type DuplicateMatch struct {
	EntityID    uuid.UUID
	Score       float64
	MatchFields []string
	Reason      string
}

// ============================================================================
// Geocoding Ports
// ============================================================================

// GeocodingService defines the interface for geocoding operations.
type GeocodingService interface {
	// Geocode converts an address to coordinates.
	Geocode(ctx context.Context, address string) (*GeoLocation, error)

	// ReverseGeocode converts coordinates to an address.
	ReverseGeocode(ctx context.Context, lat, lng float64) (*GeoAddress, error)

	// ValidateAddress validates and standardizes an address.
	ValidateAddress(ctx context.Context, address domain.Address) (*ValidatedAddress, error)
}

// GeoLocation represents geographic coordinates.
type GeoLocation struct {
	Latitude  float64
	Longitude float64
	Accuracy  string
}

// GeoAddress represents a reverse-geocoded address.
type GeoAddress struct {
	FormattedAddress string
	Line1            string
	Line2            string
	City             string
	State            string
	PostalCode       string
	Country          string
	CountryCode      string
}

// ValidatedAddress represents a validated address.
type ValidatedAddress struct {
	Original    domain.Address
	Standardized domain.Address
	IsValid     bool
	Corrections []AddressCorrection
	Location    *GeoLocation
}

// AddressCorrection represents an address correction suggestion.
type AddressCorrection struct {
	Field       string
	Original    string
	Suggested   string
	Confidence  float64
}

// ============================================================================
// External Integration Ports
// ============================================================================

// CRMIntegration defines the interface for external CRM integrations.
type CRMIntegration interface {
	// SyncCustomer syncs a customer to external CRM.
	SyncCustomer(ctx context.Context, customer *domain.Customer) error

	// SyncContact syncs a contact to external CRM.
	SyncContact(ctx context.Context, contact *domain.Contact) error

	// ImportFromExternal imports data from external CRM.
	ImportFromExternal(ctx context.Context, tenantID uuid.UUID, since *time.Time) (*ImportSummary, error)

	// GetExternalID gets the external system ID for an entity.
	GetExternalID(ctx context.Context, entityType string, entityID uuid.UUID) (string, error)
}

// ImportSummary represents a summary of an import operation.
type ImportSummary struct {
	TotalRecords   int
	ImportedCount  int
	UpdatedCount   int
	SkippedCount   int
	ErrorCount     int
	Errors         []string
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
}

// ============================================================================
// ID Generator Port
// ============================================================================

// IDGenerator defines the interface for generating IDs.
type IDGenerator interface {
	// NewID generates a new UUID.
	NewID() uuid.UUID

	// NewCode generates a new customer/contact code.
	NewCode(ctx context.Context, tenantID uuid.UUID, entityType string) (string, error)

	// ValidateID validates an ID.
	ValidateID(id string) error
}

// ============================================================================
// Clock Port (for testing)
// ============================================================================

// Clock defines the interface for time operations.
type Clock interface {
	// Now returns the current time.
	Now() time.Time

	// Since returns the time elapsed since t.
	Since(t time.Time) time.Duration

	// Until returns the duration until t.
	Until(t time.Time) time.Duration
}

// ============================================================================
// Transaction Port
// ============================================================================

// TransactionManager defines the interface for transaction management.
type TransactionManager interface {
	// Begin starts a new transaction.
	Begin(ctx context.Context) (Transaction, error)

	// WithTransaction executes a function within a transaction.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Transaction defines the interface for a database transaction.
type Transaction interface {
	// Commit commits the transaction.
	Commit() error

	// Rollback rolls back the transaction.
	Rollback() error

	// Context returns the transaction context.
	Context() context.Context
}

// ============================================================================
// Configuration Port
// ============================================================================

// ConfigProvider defines the interface for configuration access.
type ConfigProvider interface {
	// GetString gets a string configuration value.
	GetString(key string) string

	// GetInt gets an integer configuration value.
	GetInt(key string) int

	// GetBool gets a boolean configuration value.
	GetBool(key string) bool

	// GetDuration gets a duration configuration value.
	GetDuration(key string) time.Duration

	// GetTenantConfig gets tenant-specific configuration.
	GetTenantConfig(tenantID uuid.UUID, key string) (interface{}, error)
}

// ============================================================================
// Metrics Port
// ============================================================================

// MetricsCollector defines the interface for metrics collection.
type MetricsCollector interface {
	// IncrementCounter increments a counter metric.
	IncrementCounter(name string, tags map[string]string)

	// RecordGauge records a gauge metric.
	RecordGauge(name string, value float64, tags map[string]string)

	// RecordHistogram records a histogram metric.
	RecordHistogram(name string, value float64, tags map[string]string)

	// RecordTiming records a timing metric.
	RecordTiming(name string, duration time.Duration, tags map[string]string)
}
