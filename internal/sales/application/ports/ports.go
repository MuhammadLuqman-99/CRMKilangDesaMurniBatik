package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Event Publisher Port
// ============================================================================

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	// Publish publishes a single event to the event bus.
	Publish(ctx context.Context, event Event) error

	// PublishBatch publishes multiple events to the event bus.
	PublishBatch(ctx context.Context, events []Event) error

	// PublishAsync publishes an event asynchronously.
	PublishAsync(ctx context.Context, event Event) error
}

// Event represents a domain event to be published.
type Event struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	TenantID      string                 `json:"tenant_id"`
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]string      `json:"metadata"`
	OccurredAt    time.Time              `json:"occurred_at"`
	Version       int                    `json:"version"`
}

// ============================================================================
// Cache Port
// ============================================================================

// CacheService defines the interface for caching operations.
type CacheService interface {
	// Get retrieves a value from the cache.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache with optional TTL.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a value from the cache.
	Delete(ctx context.Context, key string) error

	// DeletePattern removes all values matching a pattern.
	DeletePattern(ctx context.Context, pattern string) error

	// Exists checks if a key exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// GetMulti retrieves multiple values from the cache.
	GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)

	// SetMulti stores multiple values in the cache.
	SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error

	// Increment increments a numeric value.
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// SetNX sets a value only if it doesn't exist (for distributed locks).
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
}

// ============================================================================
// Customer Service Port
// ============================================================================

// CustomerService defines the interface for customer service operations.
type CustomerService interface {
	// GetCustomer retrieves a customer by ID.
	GetCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*CustomerInfo, error)

	// GetCustomerByCode retrieves a customer by code.
	GetCustomerByCode(ctx context.Context, tenantID uuid.UUID, code string) (*CustomerInfo, error)

	// CustomerExists checks if a customer exists.
	CustomerExists(ctx context.Context, tenantID, customerID uuid.UUID) (bool, error)

	// CreateCustomer creates a new customer.
	CreateCustomer(ctx context.Context, tenantID uuid.UUID, req CreateCustomerRequest) (*CustomerInfo, error)

	// GetContact retrieves a contact by ID.
	GetContact(ctx context.Context, tenantID, contactID uuid.UUID) (*ContactInfo, error)

	// ContactExists checks if a contact exists.
	ContactExists(ctx context.Context, tenantID, contactID uuid.UUID) (bool, error)

	// CreateContact creates a new contact for a customer.
	CreateContact(ctx context.Context, tenantID, customerID uuid.UUID, req CreateContactRequest) (*ContactInfo, error)
}

// CustomerInfo represents customer information from the customer service.
type CustomerInfo struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Email       *string   `json:"email,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	Industry    *string   `json:"industry,omitempty"`
	OwnerID     *uuid.UUID `json:"owner_id,omitempty"`
}

// ContactInfo represents contact information from the customer service.
type ContactInfo struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Email      string     `json:"email"`
	Phone      *string    `json:"phone,omitempty"`
	JobTitle   *string    `json:"job_title,omitempty"`
	IsPrimary  bool       `json:"is_primary"`
}

// CreateCustomerRequest represents a request to create a customer.
type CreateCustomerRequest struct {
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Email     *string            `json:"email,omitempty"`
	Phone     *string            `json:"phone,omitempty"`
	Industry  *string            `json:"industry,omitempty"`
	Website   *string            `json:"website,omitempty"`
	Address   *AddressInfo       `json:"address,omitempty"`
	OwnerID   *uuid.UUID         `json:"owner_id,omitempty"`
	Source    string             `json:"source"`
}

// CreateContactRequest represents a request to create a contact.
type CreateContactRequest struct {
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Email       string   `json:"email"`
	Phone       *string  `json:"phone,omitempty"`
	Mobile      *string  `json:"mobile,omitempty"`
	JobTitle    *string  `json:"job_title,omitempty"`
	Department  *string  `json:"department,omitempty"`
	IsPrimary   bool     `json:"is_primary"`
}

// AddressInfo represents address information.
type AddressInfo struct {
	Street1    string  `json:"street1"`
	Street2    *string `json:"street2,omitempty"`
	City       string  `json:"city"`
	State      *string `json:"state,omitempty"`
	PostalCode *string `json:"postal_code,omitempty"`
	Country    string  `json:"country"`
}

// ============================================================================
// User Service Port
// ============================================================================

// UserService defines the interface for user service operations.
type UserService interface {
	// GetUser retrieves a user by ID.
	GetUser(ctx context.Context, tenantID, userID uuid.UUID) (*UserInfo, error)

	// UserExists checks if a user exists.
	UserExists(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// GetUsersByIDs retrieves multiple users by IDs.
	GetUsersByIDs(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID) ([]*UserInfo, error)

	// GetUsersByRole retrieves users by role.
	GetUsersByRole(ctx context.Context, tenantID uuid.UUID, roleName string) ([]*UserInfo, error)

	// HasPermission checks if a user has a specific permission.
	HasPermission(ctx context.Context, tenantID, userID uuid.UUID, permission string) (bool, error)
}

// UserInfo represents user information from the user service.
type UserInfo struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Status    string    `json:"status"`
	Roles     []string  `json:"roles"`
}

// ============================================================================
// Product Service Port
// ============================================================================

// ProductService defines the interface for product service operations.
type ProductService interface {
	// GetProduct retrieves a product by ID.
	GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (*ProductInfo, error)

	// ProductExists checks if a product exists.
	ProductExists(ctx context.Context, tenantID, productID uuid.UUID) (bool, error)

	// GetProductsByIDs retrieves multiple products by IDs.
	GetProductsByIDs(ctx context.Context, tenantID uuid.UUID, productIDs []uuid.UUID) ([]*ProductInfo, error)

	// SearchProducts searches for products.
	SearchProducts(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]*ProductInfo, error)
}

// ProductInfo represents product information from the product service.
type ProductInfo struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	UnitPrice   int64     `json:"unit_price"`
	Currency    string    `json:"currency"`
	Category    *string   `json:"category,omitempty"`
	Status      string    `json:"status"`
	ImageURL    *string   `json:"image_url,omitempty"`
}

// ============================================================================
// Notification Service Port
// ============================================================================

// NotificationService defines the interface for notification operations.
type NotificationService interface {
	// SendEmail sends an email notification.
	SendEmail(ctx context.Context, req EmailRequest) error

	// SendInApp sends an in-app notification.
	SendInApp(ctx context.Context, req InAppNotificationRequest) error

	// SendSMS sends an SMS notification.
	SendSMS(ctx context.Context, req SMSRequest) error

	// SendBatch sends multiple notifications.
	SendBatch(ctx context.Context, notifications []NotificationRequest) error
}

// EmailRequest represents an email notification request.
type EmailRequest struct {
	TenantID    uuid.UUID         `json:"tenant_id"`
	To          []string          `json:"to"`
	CC          []string          `json:"cc,omitempty"`
	BCC         []string          `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	TemplateName string           `json:"template_name,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	HTMLBody    *string           `json:"html_body,omitempty"`
	TextBody    *string           `json:"text_body,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	Priority    string            `json:"priority,omitempty"` // high, normal, low
	ReplyTo     *string           `json:"reply_to,omitempty"`
}

// EmailAttachment represents an email attachment.
type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content"`
}

// InAppNotificationRequest represents an in-app notification request.
type InAppNotificationRequest struct {
	TenantID    uuid.UUID              `json:"tenant_id"`
	UserIDs     []uuid.UUID            `json:"user_ids"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Type        string                 `json:"type"` // info, warning, success, error
	ActionURL   *string                `json:"action_url,omitempty"`
	ActionLabel *string                `json:"action_label,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
}

// SMSRequest represents an SMS notification request.
type SMSRequest struct {
	TenantID uuid.UUID `json:"tenant_id"`
	To       string    `json:"to"`
	Message  string    `json:"message"`
}

// NotificationRequest represents a generic notification request.
type NotificationRequest struct {
	Type    string      `json:"type"` // email, in_app, sms
	Payload interface{} `json:"payload"`
}

// ============================================================================
// Search Service Port
// ============================================================================

// SearchService defines the interface for search operations.
type SearchService interface {
	// IndexLead indexes a lead for search.
	IndexLead(ctx context.Context, lead SearchableLead) error

	// IndexOpportunity indexes an opportunity for search.
	IndexOpportunity(ctx context.Context, opportunity SearchableOpportunity) error

	// IndexDeal indexes a deal for search.
	IndexDeal(ctx context.Context, deal SearchableDeal) error

	// Search performs a full-text search.
	Search(ctx context.Context, tenantID uuid.UUID, query SearchQuery) (*SearchResult, error)

	// DeleteIndex removes an entity from the search index.
	DeleteIndex(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) error

	// BulkIndex indexes multiple entities.
	BulkIndex(ctx context.Context, entities []SearchableEntity) error
}

// SearchableLead represents a lead in search index.
type SearchableLead struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Company   *string   `json:"company,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Status    string    `json:"status"`
	Source    string    `json:"source"`
	Score     int       `json:"score"`
	Tags      []string  `json:"tags,omitempty"`
	OwnerID   *uuid.UUID `json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchableOpportunity represents an opportunity in search index.
type SearchableOpportunity struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	Name              string     `json:"name"`
	Status            string     `json:"status"`
	Amount            int64      `json:"amount"`
	Currency          string     `json:"currency"`
	Probability       int        `json:"probability"`
	CustomerID        *uuid.UUID `json:"customer_id,omitempty"`
	CustomerName      *string    `json:"customer_name,omitempty"`
	PipelineID        uuid.UUID  `json:"pipeline_id"`
	StageID           uuid.UUID  `json:"stage_id"`
	StageName         string     `json:"stage_name"`
	OwnerID           uuid.UUID  `json:"owner_id"`
	ExpectedCloseDate time.Time  `json:"expected_close_date"`
	Tags              []string   `json:"tags,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// SearchableDeal represents a deal in search index.
type SearchableDeal struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	DealNumber   string     `json:"deal_number"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	TotalAmount  int64      `json:"total_amount"`
	Currency     string     `json:"currency"`
	CustomerID   uuid.UUID  `json:"customer_id"`
	CustomerName string     `json:"customer_name"`
	OwnerID      uuid.UUID  `json:"owner_id"`
	ClosedDate   *time.Time `json:"closed_date,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// SearchableEntity represents any searchable entity.
type SearchableEntity struct {
	Type   string      `json:"type"` // lead, opportunity, deal
	Entity interface{} `json:"entity"`
}

// SearchQuery represents a search query.
type SearchQuery struct {
	Query      string            `json:"query"`
	EntityType *string           `json:"entity_type,omitempty"` // lead, opportunity, deal, or nil for all
	Filters    map[string]interface{} `json:"filters,omitempty"`
	Sort       *SearchSort       `json:"sort,omitempty"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// SearchSort represents sorting options.
type SearchSort struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// SearchResult represents search results.
type SearchResult struct {
	Hits       []SearchHit `json:"hits"`
	TotalHits  int64       `json:"total_hits"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// SearchHit represents a single search result.
type SearchHit struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Score      float64                `json:"score"`
	Source     map[string]interface{} `json:"source"`
	Highlights map[string][]string    `json:"highlights,omitempty"`
}

// ============================================================================
// File Storage Port
// ============================================================================

// FileStorageService defines the interface for file storage operations.
type FileStorageService interface {
	// Upload uploads a file.
	Upload(ctx context.Context, req FileUploadRequest) (*FileInfo, error)

	// Download downloads a file.
	Download(ctx context.Context, tenantID uuid.UUID, fileID string) ([]byte, error)

	// Delete deletes a file.
	Delete(ctx context.Context, tenantID uuid.UUID, fileID string) error

	// GetURL gets a presigned URL for a file.
	GetURL(ctx context.Context, tenantID uuid.UUID, fileID string, expiresIn time.Duration) (string, error)

	// ListFiles lists files for an entity.
	ListFiles(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) ([]*FileInfo, error)
}

// FileUploadRequest represents a file upload request.
type FileUploadRequest struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	EntityType  string    `json:"entity_type"` // lead, opportunity, deal
	EntityID    uuid.UUID `json:"entity_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Content     []byte    `json:"content"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FileInfo represents file information.
type FileInfo struct {
	ID          string            `json:"id"`
	TenantID    uuid.UUID         `json:"tenant_id"`
	EntityType  string            `json:"entity_type"`
	EntityID    uuid.UUID         `json:"entity_id"`
	Filename    string            `json:"filename"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	URL         string            `json:"url"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	UploadedAt  time.Time         `json:"uploaded_at"`
	UploadedBy  uuid.UUID         `json:"uploaded_by"`
}

// ============================================================================
// Audit Log Port
// ============================================================================

// AuditLogService defines the interface for audit logging.
type AuditLogService interface {
	// Log logs an audit event.
	Log(ctx context.Context, entry AuditEntry) error

	// LogBatch logs multiple audit events.
	LogBatch(ctx context.Context, entries []AuditEntry) error
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	TenantID      uuid.UUID              `json:"tenant_id"`
	UserID        uuid.UUID              `json:"user_id"`
	Action        string                 `json:"action"`
	EntityType    string                 `json:"entity_type"`
	EntityID      uuid.UUID              `json:"entity_id"`
	OldValues     map[string]interface{} `json:"old_values,omitempty"`
	NewValues     map[string]interface{} `json:"new_values,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ============================================================================
// Transaction Manager Port
// ============================================================================

// TransactionManager defines the interface for managing database transactions.
type TransactionManager interface {
	// Begin starts a new transaction.
	Begin(ctx context.Context) (Transaction, error)

	// WithTransaction executes a function within a transaction.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// Transaction represents a database transaction.
type Transaction interface {
	// Commit commits the transaction.
	Commit() error

	// Rollback rolls back the transaction.
	Rollback() error

	// Context returns the transaction context.
	Context() context.Context
}

// ============================================================================
// ID Generator Port
// ============================================================================

// IDGenerator defines the interface for generating IDs.
type IDGenerator interface {
	// GenerateID generates a new UUID.
	GenerateID() uuid.UUID

	// GenerateDealNumber generates a deal number.
	GenerateDealNumber(ctx context.Context, tenantID uuid.UUID) (string, error)

	// GenerateLeadNumber generates a lead number.
	GenerateLeadNumber(ctx context.Context, tenantID uuid.UUID) (string, error)
}

// ============================================================================
// Webhook Service Port
// ============================================================================

// WebhookService defines the interface for webhook operations.
type WebhookService interface {
	// Send sends a webhook notification.
	Send(ctx context.Context, req WebhookRequest) error

	// SendAsync sends a webhook notification asynchronously.
	SendAsync(ctx context.Context, req WebhookRequest) error
}

// WebhookRequest represents a webhook request.
type WebhookRequest struct {
	TenantID  uuid.UUID              `json:"tenant_id"`
	URL       string                 `json:"url"`
	Method    string                 `json:"method"` // POST, PUT
	Headers   map[string]string      `json:"headers,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
	RetryCount int                   `json:"retry_count,omitempty"`
}

// ============================================================================
// Analytics Service Port
// ============================================================================

// AnalyticsService defines the interface for analytics operations.
type AnalyticsService interface {
	// TrackEvent tracks an analytics event.
	TrackEvent(ctx context.Context, event AnalyticsEvent) error

	// GetMetrics retrieves metrics for a tenant.
	GetMetrics(ctx context.Context, tenantID uuid.UUID, query MetricsQuery) (*MetricsResult, error)
}

// AnalyticsEvent represents an analytics event.
type AnalyticsEvent struct {
	TenantID   uuid.UUID              `json:"tenant_id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	EventName  string                 `json:"event_name"`
	EntityType string                 `json:"entity_type"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// MetricsQuery represents a metrics query.
type MetricsQuery struct {
	MetricNames []string          `json:"metric_names"`
	StartDate   time.Time         `json:"start_date"`
	EndDate     time.Time         `json:"end_date"`
	Granularity string            `json:"granularity"` // hour, day, week, month
	Dimensions  []string          `json:"dimensions,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
}

// MetricsResult represents metrics query results.
type MetricsResult struct {
	Metrics map[string][]MetricDataPoint `json:"metrics"`
}

// MetricDataPoint represents a single metric data point.
type MetricDataPoint struct {
	Timestamp  time.Time              `json:"timestamp"`
	Value      float64                `json:"value"`
	Dimensions map[string]string      `json:"dimensions,omitempty"`
}
