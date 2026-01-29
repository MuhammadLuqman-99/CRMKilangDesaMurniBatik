// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/ports"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// MockCustomerRepository is a mock implementation of domain.CustomerRepository.
type MockCustomerRepository struct {
	customers       map[uuid.UUID]*domain.Customer
	createErr       error
	updateErr       error
	findByIDErr     error
	existsByCodeVal bool
	existsByCodeErr error
	existsByEmailVal bool
	existsByEmailErr error
	findDuplicatesVal []*domain.Customer
}

func NewMockCustomerRepository() *MockCustomerRepository {
	return &MockCustomerRepository{
		customers: make(map[uuid.UUID]*domain.Customer),
	}
}

func (m *MockCustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.customers[customer.ID] = customer
	return nil
}

func (m *MockCustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.customers[customer.ID] = customer
	return nil
}

func (m *MockCustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.customers, id)
	return nil
}

func (m *MockCustomerRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	delete(m.customers, id)
	return nil
}

func (m *MockCustomerRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	customer, ok := m.customers[id]
	if !ok {
		return nil, domain.NewNotFoundError("customer", id.String())
	}
	return customer, nil
}

func (m *MockCustomerRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Customer, error) {
	for _, c := range m.customers {
		if c.TenantID == tenantID && c.Code == code {
			return c, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockCustomerRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.Customer, error) {
	return nil, errors.New("not found")
}

func (m *MockCustomerRepository) List(ctx context.Context, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	return &domain.CustomerList{Customers: []*domain.Customer{}}, nil
}

func (m *MockCustomerRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	return &domain.CustomerList{}, nil
}

func (m *MockCustomerRepository) FindByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	return &domain.CustomerList{}, nil
}

func (m *MockCustomerRepository) FindByTag(ctx context.Context, tenantID uuid.UUID, tag string, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	return &domain.CustomerList{}, nil
}

func (m *MockCustomerRepository) FindBySegment(ctx context.Context, tenantID, segmentID uuid.UUID, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	return &domain.CustomerList{}, nil
}

func (m *MockCustomerRepository) FindByStatus(ctx context.Context, tenantID uuid.UUID, status domain.CustomerStatus, filter domain.CustomerFilter) (*domain.CustomerList, error) {
	return &domain.CustomerList{}, nil
}

func (m *MockCustomerRepository) FindDuplicates(ctx context.Context, tenantID uuid.UUID, email, phone, name string) ([]*domain.Customer, error) {
	return m.findDuplicatesVal, nil
}

func (m *MockCustomerRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return int64(len(m.customers)), nil
}

func (m *MockCustomerRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.CustomerStatus]int64, error) {
	return map[domain.CustomerStatus]int64{}, nil
}

func (m *MockCustomerRepository) GetStats(ctx context.Context, tenantID uuid.UUID) (*domain.CustomerStats, error) {
	return &domain.CustomerStats{}, nil
}

func (m *MockCustomerRepository) BulkCreate(ctx context.Context, customers []*domain.Customer) error {
	return nil
}

func (m *MockCustomerRepository) BulkUpdate(ctx context.Context, ids []uuid.UUID, updates map[string]interface{}) error {
	return nil
}

func (m *MockCustomerRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	return nil
}

func (m *MockCustomerRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.customers[id]
	return ok, nil
}

func (m *MockCustomerRepository) ExistsByCode(ctx context.Context, tenantID uuid.UUID, code string) (bool, error) {
	if m.existsByCodeErr != nil {
		return false, m.existsByCodeErr
	}
	return m.existsByCodeVal, nil
}

func (m *MockCustomerRepository) ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email string) (bool, error) {
	if m.existsByEmailErr != nil {
		return false, m.existsByEmailErr
	}
	return m.existsByEmailVal, nil
}

func (m *MockCustomerRepository) GetVersion(ctx context.Context, id uuid.UUID) (int, error) {
	if c, ok := m.customers[id]; ok {
		return c.Version, nil
	}
	return 0, errors.New("not found")
}

func (m *MockCustomerRepository) FindNeedingFollowUp(ctx context.Context, tenantID uuid.UUID, before time.Time) ([]*domain.Customer, error) {
	return nil, nil
}

func (m *MockCustomerRepository) FindInactive(ctx context.Context, tenantID uuid.UUID, lastContactBefore time.Time) ([]*domain.Customer, error) {
	return nil, nil
}

func (m *MockCustomerRepository) FindRecentlyCreated(ctx context.Context, tenantID uuid.UUID, since time.Time, limit int) ([]*domain.Customer, error) {
	return nil, nil
}

func (m *MockCustomerRepository) FindRecentlyUpdated(ctx context.Context, tenantID uuid.UUID, since time.Time, limit int) ([]*domain.Customer, error) {
	return nil, nil
}

// MockContactRepository is a mock implementation of domain.ContactRepository.
type MockContactRepository struct{}

func NewMockContactRepository() *MockContactRepository {
	return &MockContactRepository{}
}

func (m *MockContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	return nil, errors.New("not found")
}

func (m *MockContactRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Contact, error) {
	return nil, nil
}

func (m *MockContactRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]*domain.Contact, error) {
	return nil, nil
}

func (m *MockContactRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, filter domain.ContactFilter) (*domain.ContactList, error) {
	return &domain.ContactList{}, nil
}

func (m *MockContactRepository) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error) {
	return 0, nil
}

// MockOutboxRepository is a mock implementation of domain.OutboxRepository.
type MockCustomerOutboxRepository struct {
	entries   []*domain.OutboxEntry
	createErr error
}

func NewMockCustomerOutboxRepository() *MockCustomerOutboxRepository {
	return &MockCustomerOutboxRepository{
		entries: make([]*domain.OutboxEntry, 0),
	}
}

func (m *MockCustomerOutboxRepository) Create(ctx context.Context, entry *domain.OutboxEntry) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockCustomerOutboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockCustomerOutboxRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, err string) error {
	return nil
}

func (m *MockCustomerOutboxRepository) FindPending(ctx context.Context, limit int) ([]*domain.OutboxEntry, error) {
	return m.entries, nil
}

func (m *MockCustomerOutboxRepository) FindFailed(ctx context.Context, limit int) ([]*domain.OutboxEntry, error) {
	return nil, nil
}

func (m *MockCustomerOutboxRepository) DeleteOld(ctx context.Context, before time.Time) error {
	return nil
}

// MockNoteRepository is a mock implementation
type MockNoteRepository struct{}

func NewMockNoteRepository() *MockNoteRepository {
	return &MockNoteRepository{}
}

func (m *MockNoteRepository) Create(ctx context.Context, note *domain.Note) error { return nil }
func (m *MockNoteRepository) Update(ctx context.Context, note *domain.Note) error { return nil }
func (m *MockNoteRepository) Delete(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *MockNoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	return nil, errors.New("not found")
}
func (m *MockNoteRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID, filter domain.NoteFilter) ([]*domain.Note, error) {
	return nil, nil
}
func (m *MockNoteRepository) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error) {
	return 0, nil
}

// MockActivityRepository is a mock implementation
type MockActivityRepository struct{}

func NewMockActivityRepository() *MockActivityRepository {
	return &MockActivityRepository{}
}

func (m *MockActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	return nil
}
func (m *MockActivityRepository) Update(ctx context.Context, activity *domain.Activity) error {
	return nil
}
func (m *MockActivityRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *MockActivityRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Activity, error) {
	return nil, errors.New("not found")
}
func (m *MockActivityRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	return &domain.ActivityList{}, nil
}
func (m *MockActivityRepository) FindByContactID(ctx context.Context, contactID uuid.UUID, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	return &domain.ActivityList{}, nil
}
func (m *MockActivityRepository) FindByPerformedBy(ctx context.Context, userID uuid.UUID, filter domain.ActivityFilter) (*domain.ActivityList, error) {
	return &domain.ActivityList{}, nil
}
func (m *MockActivityRepository) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error) {
	return 0, nil
}
func (m *MockActivityRepository) GetActivitySummary(ctx context.Context, customerID uuid.UUID) (map[domain.ActivityType]int, error) {
	return nil, nil
}

// MockSegmentRepository is a mock implementation
type MockSegmentRepository struct{}

func NewMockSegmentRepository() *MockSegmentRepository {
	return &MockSegmentRepository{}
}

func (m *MockSegmentRepository) Create(ctx context.Context, segment *domain.Segment) error { return nil }
func (m *MockSegmentRepository) Update(ctx context.Context, segment *domain.Segment) error { return nil }
func (m *MockSegmentRepository) Delete(ctx context.Context, id uuid.UUID) error            { return nil }
func (m *MockSegmentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Segment, error) {
	return nil, errors.New("not found")
}
func (m *MockSegmentRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.Segment, error) {
	return nil, nil
}
func (m *MockSegmentRepository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.Segment, error) {
	return nil, errors.New("not found")
}
func (m *MockSegmentRepository) GetCustomerCount(ctx context.Context, segmentID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockSegmentRepository) RefreshDynamic(ctx context.Context, segmentID uuid.UUID) error {
	return nil
}

// MockImportRepository is a mock implementation
type MockImportRepository struct{}

func NewMockImportRepository() *MockImportRepository {
	return &MockImportRepository{}
}

func (m *MockImportRepository) CreateImport(ctx context.Context, imp *domain.Import) error { return nil }
func (m *MockImportRepository) UpdateImport(ctx context.Context, imp *domain.Import) error { return nil }
func (m *MockImportRepository) FindImportByID(ctx context.Context, id uuid.UUID) (*domain.Import, error) {
	return nil, errors.New("not found")
}
func (m *MockImportRepository) FindImportsByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*domain.Import, error) {
	return nil, nil
}
func (m *MockImportRepository) CreateImportError(ctx context.Context, err *domain.ImportError) error {
	return nil
}
func (m *MockImportRepository) FindImportErrors(ctx context.Context, importID uuid.UUID) ([]*domain.ImportError, error) {
	return nil, nil
}

// MockUnitOfWork is a mock implementation of domain.UnitOfWork.
type MockUnitOfWork struct {
	customerRepo *MockCustomerRepository
	contactRepo  *MockContactRepository
	noteRepo     *MockNoteRepository
	activityRepo *MockActivityRepository
	segmentRepo  *MockSegmentRepository
	outboxRepo   *MockCustomerOutboxRepository
	importRepo   *MockImportRepository
	beginErr     error
	commitErr    error
}

func NewMockUnitOfWork() *MockUnitOfWork {
	return &MockUnitOfWork{
		customerRepo: NewMockCustomerRepository(),
		contactRepo:  NewMockContactRepository(),
		noteRepo:     NewMockNoteRepository(),
		activityRepo: NewMockActivityRepository(),
		segmentRepo:  NewMockSegmentRepository(),
		outboxRepo:   NewMockCustomerOutboxRepository(),
		importRepo:   NewMockImportRepository(),
	}
}

func (m *MockUnitOfWork) Begin(ctx context.Context) (context.Context, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	return ctx, nil
}

func (m *MockUnitOfWork) Commit(ctx context.Context) error {
	return m.commitErr
}

func (m *MockUnitOfWork) Rollback(ctx context.Context) error {
	return nil
}

func (m *MockUnitOfWork) Customers() domain.CustomerRepository {
	return m.customerRepo
}

func (m *MockUnitOfWork) Contacts() domain.ContactRepository {
	return m.contactRepo
}

func (m *MockUnitOfWork) Notes() domain.NoteRepository {
	return m.noteRepo
}

func (m *MockUnitOfWork) Activities() domain.ActivityRepository {
	return m.activityRepo
}

func (m *MockUnitOfWork) Segments() domain.SegmentRepository {
	return m.segmentRepo
}

func (m *MockUnitOfWork) Outbox() domain.OutboxRepository {
	return m.outboxRepo
}

func (m *MockUnitOfWork) Imports() domain.ImportRepository {
	return m.importRepo
}

// MockCustomerEventPublisher is a mock implementation of ports.EventPublisher.
type MockCustomerEventPublisher struct {
	publishedEvents []domain.DomainEvent
	publishErr      error
}

func NewMockCustomerEventPublisher() *MockCustomerEventPublisher {
	return &MockCustomerEventPublisher{
		publishedEvents: make([]domain.DomainEvent, 0),
	}
}

func (m *MockCustomerEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedEvents = append(m.publishedEvents, event)
	return nil
}

func (m *MockCustomerEventPublisher) PublishAll(ctx context.Context, events []domain.DomainEvent) error {
	if m.publishErr != nil {
		return m.publishErr
	}
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

func (m *MockCustomerEventPublisher) PublishAsync(ctx context.Context, events []domain.DomainEvent) error {
	return nil
}

// MockDuplicateDetector is a mock implementation of ports.DuplicateDetector.
type MockDuplicateDetector struct {
	matches      []ports.DuplicateMatch
	findErr      error
}

func NewMockDuplicateDetector() *MockDuplicateDetector {
	return &MockDuplicateDetector{
		matches: make([]ports.DuplicateMatch, 0),
	}
}

func (m *MockDuplicateDetector) FindDuplicateCustomers(ctx context.Context, tenantID uuid.UUID, customer *domain.Customer) ([]ports.DuplicateMatch, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.matches, nil
}

func (m *MockDuplicateDetector) FindDuplicateContacts(ctx context.Context, tenantID uuid.UUID, contact *domain.Contact) ([]ports.DuplicateMatch, error) {
	return nil, nil
}

func (m *MockDuplicateDetector) CalculateSimilarity(ctx context.Context, entity1, entity2 interface{}) (float64, []string, error) {
	return 0, nil, nil
}

// MockIDGenerator is a mock implementation of ports.IDGenerator.
type MockIDGenerator struct {
	codeSeq int
}

func NewMockIDGenerator() *MockIDGenerator {
	return &MockIDGenerator{}
}

func (m *MockIDGenerator) NewID() uuid.UUID {
	return uuid.New()
}

func (m *MockIDGenerator) NewCode(ctx context.Context, tenantID uuid.UUID, entityType string) (string, error) {
	m.codeSeq++
	return "CUST-" + string(rune('0'+m.codeSeq)), nil
}

func (m *MockIDGenerator) ValidateID(id string) error {
	return nil
}

// MockCustomerCacheService is a mock implementation of ports.CacheService.
type MockCustomerCacheService struct {
	cache map[string][]byte
}

func NewMockCustomerCacheService() *MockCustomerCacheService {
	return &MockCustomerCacheService{
		cache: make(map[string][]byte),
	}
}

func (m *MockCustomerCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	v, ok := m.cache[key]
	if !ok {
		return nil, ports.ErrCacheMiss
	}
	return v, nil
}

func (m *MockCustomerCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.cache[key] = value
	return nil
}

func (m *MockCustomerCacheService) Delete(ctx context.Context, key string) error {
	delete(m.cache, key)
	return nil
}

func (m *MockCustomerCacheService) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockCustomerCacheService) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.cache[key]
	return ok, nil
}

func (m *MockCustomerCacheService) GetOrSet(ctx context.Context, key string, ttl time.Duration, getter func() ([]byte, error)) ([]byte, error) {
	if v, ok := m.cache[key]; ok {
		return v, nil
	}
	v, err := getter()
	if err != nil {
		return nil, err
	}
	m.cache[key] = v
	return v, nil
}

func (m *MockCustomerCacheService) Invalidate(ctx context.Context, entityType string, entityID uuid.UUID) error {
	return nil
}

func (m *MockCustomerCacheService) InvalidateByTenant(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}

// MockCustomerAuditLogger is a mock implementation of ports.AuditLogger.
type MockCustomerAuditLogger struct {
	entries []ports.AuditEntry
	logErr  error
}

func NewMockCustomerAuditLogger() *MockCustomerAuditLogger {
	return &MockCustomerAuditLogger{
		entries: make([]ports.AuditEntry, 0),
	}
}

func (m *MockCustomerAuditLogger) LogAction(ctx context.Context, entry ports.AuditEntry) error {
	if m.logErr != nil {
		return m.logErr
	}
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockCustomerAuditLogger) GetAuditLog(ctx context.Context, filter ports.AuditFilter) ([]ports.AuditEntry, error) {
	return m.entries, nil
}

func (m *MockCustomerAuditLogger) GetEntityHistory(ctx context.Context, entityType string, entityID uuid.UUID) ([]ports.AuditEntry, error) {
	return nil, nil
}

// ============================================================================
// CreateCustomerUseCase Tests
// ============================================================================

func TestCreateCustomerUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateCustomerRequest{
			Name:  "Test Customer",
			Type:  domain.CustomerTypeCompany,
			Email: "test@example.com",
		},
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if !result.Created {
		t.Error("Expected Created to be true")
	}
	if result.Customer == nil {
		t.Fatal("Expected Customer, got nil")
	}
}

func TestCreateCustomerUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID: uuid.Nil, // Missing
		UserID:   uuid.New(),
		Request: &dto.CreateCustomerRequest{
			Name: "Test Customer",
			Type: domain.CustomerTypeCompany,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestCreateCustomerUseCase_Execute_MissingUserID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID: uuid.New(),
		UserID:   uuid.Nil, // Missing
		Request: &dto.CreateCustomerRequest{
			Name: "Test Customer",
			Type: domain.CustomerTypeCompany,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing user ID, got nil")
	}
}

func TestCreateCustomerUseCase_Execute_MissingName(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateCustomerRequest{
			Name: "", // Missing
			Type: domain.CustomerTypeCompany,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing name, got nil")
	}
}

func TestCreateCustomerUseCase_Execute_DuplicateCodeFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.customerRepo.existsByCodeVal = true // Code already exists
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	config.AutoGenerateCode = false // Manual code
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateCustomerRequest{
			Name: "Test Customer",
			Type: domain.CustomerTypeCompany,
			Code: "EXISTING-CODE",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for duplicate code, got nil")
	}
}

func TestCreateCustomerUseCase_Execute_DuplicateEmailFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.customerRepo.existsByEmailVal = true // Email already exists
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateCustomerRequest{
			Name:  "Test Customer",
			Type:  domain.CustomerTypeCompany,
			Email: "existing@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for duplicate email, got nil")
	}
}

func TestCreateCustomerUseCase_Execute_DuplicateDetection(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	duplicateDetector.matches = []ports.DuplicateMatch{
		{
			EntityID:    uuid.New(),
			Score:       0.95,
			MatchFields: []string{"email", "name"},
			Reason:      "Potential duplicate",
		},
	}
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	// Add a customer to be found as duplicate
	existingCustomer := &domain.Customer{
		ID:       duplicateDetector.matches[0].EntityID,
		TenantID: uuid.New(),
		Name:     "Existing Customer",
	}
	existingCustomer.Email, _ = domain.NewEmail("existing@example.com")
	uow.customerRepo.customers[existingCustomer.ID] = existingCustomer

	config := DefaultCreateCustomerConfig()
	config.DuplicateCheckEnabled = true
	config.DuplicateThreshold = 0.9
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID:       uuid.New(),
		UserID:         uuid.New(),
		SkipDuplicates: false,
		Request: &dto.CreateCustomerRequest{
			Name:  "Existing Customer",
			Type:  domain.CustomerTypeCompany,
			Email: "existing@example.com",
		},
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for duplicate detection, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Created {
		t.Error("Expected Created to be false when duplicates found")
	}
	if len(result.Duplicates) == 0 {
		t.Error("Expected Duplicates to be populated")
	}
}

func TestCreateCustomerUseCase_Execute_SkipDuplicates(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	duplicateDetector.matches = []ports.DuplicateMatch{
		{
			EntityID:    uuid.New(),
			Score:       0.95,
			MatchFields: []string{"email"},
		},
	}
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID:       uuid.New(),
		UserID:         uuid.New(),
		SkipDuplicates: true, // Skip duplicate check
		Request: &dto.CreateCustomerRequest{
			Name:  "Test Customer",
			Type:  domain.CustomerTypeCompany,
			Email: "test@example.com",
		},
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if !result.Created {
		t.Error("Expected Created to be true when skipping duplicates")
	}
}

func TestCreateCustomerUseCase_Execute_TransactionBeginError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.beginErr = errors.New("transaction begin failed")
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateCustomerRequest{
			Name: "Test Customer",
			Type: domain.CustomerTypeCompany,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for transaction begin failure, got nil")
	}
}

func TestCreateCustomerUseCase_Execute_TransactionCommitError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.commitErr = errors.New("transaction commit failed")
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	input := CreateCustomerInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateCustomerRequest{
			Name: "Test Customer",
			Type: domain.CustomerTypeCompany,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for transaction commit failure, got nil")
	}
}

func TestCreateCustomerUseCase_Execute_TooManyContacts(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	config.MaxContactsPerCustomer = 2 // Only 2 contacts allowed
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

	// Create request with too many contacts
	contacts := make([]dto.CreateContactInput, 5)
	for i := 0; i < 5; i++ {
		contacts[i] = dto.CreateContactInput{
			FirstName: "Contact",
			LastName:  "Name",
			Email:     "contact@example.com",
		}
	}

	input := CreateCustomerInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateCustomerRequest{
			Name:     "Test Customer",
			Type:     domain.CustomerTypeCompany,
			Contacts: contacts,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for too many contacts, got nil")
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkCreateCustomerUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	duplicateDetector := NewMockDuplicateDetector()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultCreateCustomerConfig()
	uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := CreateCustomerInput{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			Request: &dto.CreateCustomerRequest{
				Name: "Test Customer",
				Type: domain.CustomerTypeCompany,
			},
		}
		_, _ = uc.Execute(ctx, input)
	}
}

// ============================================================================
// Table-Driven Tests for Validation
// ============================================================================

func TestCreateCustomerUseCase_Execute_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		input     CreateCustomerInput
		expectErr bool
	}{
		{
			name: "valid input",
			input: CreateCustomerInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Request: &dto.CreateCustomerRequest{
					Name: "Test Customer",
					Type: domain.CustomerTypeCompany,
				},
			},
			expectErr: false,
		},
		{
			name: "missing tenant ID",
			input: CreateCustomerInput{
				TenantID: uuid.Nil,
				UserID:   uuid.New(),
				Request: &dto.CreateCustomerRequest{
					Name: "Test Customer",
					Type: domain.CustomerTypeCompany,
				},
			},
			expectErr: true,
		},
		{
			name: "missing user ID",
			input: CreateCustomerInput{
				TenantID: uuid.New(),
				UserID:   uuid.Nil,
				Request: &dto.CreateCustomerRequest{
					Name: "Test Customer",
					Type: domain.CustomerTypeCompany,
				},
			},
			expectErr: true,
		},
		{
			name: "missing request",
			input: CreateCustomerInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Request:  nil,
			},
			expectErr: true,
		},
		{
			name: "missing name",
			input: CreateCustomerInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Request: &dto.CreateCustomerRequest{
					Name: "",
					Type: domain.CustomerTypeCompany,
				},
			},
			expectErr: true,
		},
		{
			name: "missing type",
			input: CreateCustomerInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Request: &dto.CreateCustomerRequest{
					Name: "Test Customer",
					Type: "",
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := NewMockUnitOfWork()
			eventPublisher := NewMockCustomerEventPublisher()
			duplicateDetector := NewMockDuplicateDetector()
			idGenerator := NewMockIDGenerator()
			cache := NewMockCustomerCacheService()
			auditLogger := NewMockCustomerAuditLogger()

			config := DefaultCreateCustomerConfig()
			uc := NewCreateCustomerUseCase(uow, eventPublisher, duplicateDetector, idGenerator, cache, auditLogger, config)

			_, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}
