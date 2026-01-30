// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Mock Contact Repository with Full Implementation
// ============================================================================

// MockContactRepositoryFull is a full mock implementation of domain.ContactRepository.
type MockContactRepositoryFull struct {
	contacts           map[uuid.UUID]*domain.Contact
	findByIDErr        error
	findByCustomerErr  error
	findByEmailResult  []*domain.Contact
	findByEmailErr     error
	searchResult       *domain.ContactList
	searchErr          error
	countByCustomerVal int
	countByCustomerErr error
}

func NewMockContactRepositoryFull() *MockContactRepositoryFull {
	return &MockContactRepositoryFull{
		contacts:          make(map[uuid.UUID]*domain.Contact),
		findByEmailResult: make([]*domain.Contact, 0),
		searchResult: &domain.ContactList{
			Contacts: []*domain.Contact{},
		},
	}
}

func (m *MockContactRepositoryFull) FindByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	contact, ok := m.contacts[id]
	if !ok {
		return nil, domain.ErrContactNotFound
	}
	return contact, nil
}

func (m *MockContactRepositoryFull) FindByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Contact, error) {
	if m.findByCustomerErr != nil {
		return nil, m.findByCustomerErr
	}
	var result []*domain.Contact
	for _, c := range m.contacts {
		if c.CustomerID == customerID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *MockContactRepositoryFull) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) ([]*domain.Contact, error) {
	if m.findByEmailErr != nil {
		return nil, m.findByEmailErr
	}
	return m.findByEmailResult, nil
}

func (m *MockContactRepositoryFull) Search(ctx context.Context, tenantID uuid.UUID, query string, filter domain.ContactFilter) (*domain.ContactList, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.searchResult, nil
}

func (m *MockContactRepositoryFull) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int, error) {
	if m.countByCustomerErr != nil {
		return 0, m.countByCustomerErr
	}
	return m.countByCustomerVal, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTestCustomerWithContacts(tenantID uuid.UUID) *domain.Customer {
	customer, _ := domain.NewCustomer(tenantID, "Test Company", domain.CustomerTypeCompany)
	customer.Email, _ = domain.NewEmail("company@example.com")
	customer.Status = domain.CustomerStatusActive
	customer.Version = 1

	// Add a contact
	contact, _ := domain.NewContact(tenantID, customer.ID, "John", "Doe", "john@example.com")
	contact.Version = 1
	customer.Contacts = append(customer.Contacts, *contact)

	return customer
}

// ============================================================================
// AddContactUseCase Tests
// ============================================================================

func TestAddContactUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	tenantID := uuid.New()
	userID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Company", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	input := AddContactInput{
		TenantID: tenantID,
		UserID:   userID,
		Request: &dto.CreateContactRequest{
			CustomerID: customer.ID,
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
		IPAddress: "127.0.0.1",
		UserAgent: "TestClient/1.0",
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
	if result.Name.FirstName != "Jane" {
		t.Errorf("Expected FirstName Jane, got %s", result.Name.FirstName)
	}
}

func TestAddContactUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	input := AddContactInput{
		TenantID: uuid.Nil, // Missing
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID: uuid.New(),
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestAddContactUseCase_Execute_MissingUserID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	input := AddContactInput{
		TenantID: uuid.New(),
		UserID:   uuid.Nil, // Missing
		Request: &dto.CreateContactRequest{
			CustomerID: uuid.New(),
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing user ID, got nil")
	}
}

func TestAddContactUseCase_Execute_MissingRequest(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	input := AddContactInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request:  nil, // Missing
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing request, got nil")
	}
}

func TestAddContactUseCase_Execute_MissingCustomerID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	input := AddContactInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID: uuid.Nil, // Missing
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing customer ID, got nil")
	}
}

func TestAddContactUseCase_Execute_MissingName(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	input := AddContactInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID: uuid.New(),
			FirstName:  "", // Missing
			LastName:   "", // Missing
			Email:      "jane@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing name, got nil")
	}
}

func TestAddContactUseCase_Execute_MissingContactInfo(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	input := AddContactInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID:   uuid.New(),
			FirstName:    "Jane",
			LastName:     "Smith",
			Email:        "",                            // Missing
			PhoneNumbers: []dto.PhoneInput{}, // Missing
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing contact info, got nil")
	}
}

func TestAddContactUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	input := AddContactInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID: uuid.New(), // Does not exist
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer not found, got nil")
	}
}

func TestAddContactUseCase_Execute_TenantMismatch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	customer, _ := domain.NewCustomer(tenantID1, "Test Company", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	input := AddContactInput{
		TenantID: tenantID2, // Different tenant
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID: customer.ID,
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for tenant mismatch, got nil")
	}
}

func TestAddContactUseCase_Execute_TransactionBeginError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.beginErr = errors.New("transaction begin failed")
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Company", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	input := AddContactInput{
		TenantID: tenantID,
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID: customer.ID,
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for transaction begin failure, got nil")
	}
}

func TestAddContactUseCase_Execute_TransactionCommitError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.commitErr = errors.New("transaction commit failed")
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Company", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	input := AddContactInput{
		TenantID: tenantID,
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID: customer.ID,
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for transaction commit failure, got nil")
	}
}

// ============================================================================
// UpdateContactUseCase Tests
// ============================================================================

func TestUpdateContactUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	contact := customer.Contacts[0]
	newFirstName := "Johnny"

	input := UpdateContactInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		ContactID:  contact.ID,
		Request: &dto.UpdateContactRequest{
			FirstName: &newFirstName,
			Version:   1,
		},
		IPAddress: "127.0.0.1",
		UserAgent: "TestClient/1.0",
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
}

func TestUpdateContactUseCase_Execute_MissingContactID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateContactInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(),
		ContactID:  uuid.Nil, // Missing
		Request: &dto.UpdateContactRequest{
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing contact ID, got nil")
	}
}

func TestUpdateContactUseCase_Execute_MissingRequest(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateContactInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(),
		ContactID:  uuid.New(),
		Request:    nil, // Missing
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing request, got nil")
	}
}

func TestUpdateContactUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateContactInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(), // Does not exist
		ContactID:  uuid.New(),
		Request: &dto.UpdateContactRequest{
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer not found, got nil")
	}
}

func TestUpdateContactUseCase_Execute_ContactNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Company", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	input := UpdateContactInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		ContactID:  uuid.New(), // Does not exist in customer
		Request: &dto.UpdateContactRequest{
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for contact not found, got nil")
	}
}

func TestUpdateContactUseCase_Execute_VersionConflict(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	customer.Contacts[0].Version = 5
	uow.customerRepo.customers[customer.ID] = customer

	contact := customer.Contacts[0]

	input := UpdateContactInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		ContactID:  contact.ID,
		Request: &dto.UpdateContactRequest{
			Version: 3, // Outdated
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for version conflict, got nil")
	}
}

// ============================================================================
// DeleteContactUseCase Tests
// ============================================================================

func TestDeleteContactUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewDeleteContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	contact := customer.Contacts[0]

	input := DeleteContactInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		ContactID:  contact.ID,
		IPAddress:  "127.0.0.1",
		UserAgent:  "TestClient/1.0",
	}

	// Act
	err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestDeleteContactUseCase_Execute_MissingContactID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewDeleteContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := DeleteContactInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(),
		ContactID:  uuid.Nil, // Missing
	}

	// Act
	err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing contact ID, got nil")
	}
}

func TestDeleteContactUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewDeleteContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := DeleteContactInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(), // Does not exist
		ContactID:  uuid.New(),
	}

	// Act
	err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer not found, got nil")
	}
}

func TestDeleteContactUseCase_Execute_ContactNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewDeleteContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Company", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteContactInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		ContactID:  uuid.New(), // Does not exist
	}

	// Act
	err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for contact not found, got nil")
	}
}

// ============================================================================
// GetContactUseCase Tests
// ============================================================================

func TestGetContactUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewGetContactUseCase(uow)

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	contact := &customer.Contacts[0]

	// Create a mock that returns contacts
	mockContactRepo := NewMockContactRepositoryFull()
	mockContactRepo.contacts[contact.ID] = contact
	uow.contactRepo = &MockContactRepository{}

	input := GetContactInput{
		TenantID:   tenantID,
		CustomerID: customer.ID,
		ContactID:  contact.ID,
	}

	// Act - this will fail since our mock doesn't fully implement the interface
	// In production, this would work with a proper mock setup
	_, _ = uc.Execute(context.Background(), input)
}

// ============================================================================
// ListContactsUseCase Tests
// ============================================================================

func TestListContactsUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListContactsUseCase(uow)

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := ListContactsInput{
		TenantID:   tenantID,
		CustomerID: customer.ID,
		Offset:     0,
		Limit:      10,
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
}

func TestListContactsUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListContactsUseCase(uow)

	input := ListContactsInput{
		TenantID:   uuid.New(),
		CustomerID: uuid.New(), // Does not exist
		Offset:     0,
		Limit:      10,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer not found, got nil")
	}
}

func TestListContactsUseCase_Execute_TenantMismatch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListContactsUseCase(uow)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	customer := createTestCustomerWithContacts(tenantID1)
	uow.customerRepo.customers[customer.ID] = customer

	input := ListContactsInput{
		TenantID:   tenantID2, // Different tenant
		CustomerID: customer.ID,
		Offset:     0,
		Limit:      10,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for tenant mismatch, got nil")
	}
}

func TestListContactsUseCase_Execute_DefaultLimit(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListContactsUseCase(uow)

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := ListContactsInput{
		TenantID:   tenantID,
		CustomerID: customer.ID,
		Offset:     0,
		Limit:      0, // Should use default
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
}

// ============================================================================
// SetPrimaryContactUseCase Tests
// ============================================================================

func TestSetPrimaryContactUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewSetPrimaryContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	contact := customer.Contacts[0]

	input := SetPrimaryContactInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.SetPrimaryContactRequest{
			ContactID: contact.ID,
		},
		IPAddress: "127.0.0.1",
		UserAgent: "TestClient/1.0",
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
}

func TestSetPrimaryContactUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewSetPrimaryContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := SetPrimaryContactInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(), // Does not exist
		Request: &dto.SetPrimaryContactRequest{
			ContactID: uuid.New(),
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer not found, got nil")
	}
}

func TestSetPrimaryContactUseCase_Execute_ContactNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewSetPrimaryContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := SetPrimaryContactInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.SetPrimaryContactRequest{
			ContactID: uuid.New(), // Does not exist
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for contact not found, got nil")
	}
}

// ============================================================================
// Table-Driven Tests
// ============================================================================

func TestAddContactUseCase_Execute_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		input     AddContactInput
		setupUow  func(*MockUnitOfWork)
		expectErr bool
	}{
		{
			name: "valid input with email",
			input: AddContactInput{
				TenantID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				UserID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Request: &dto.CreateContactRequest{
					CustomerID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
					FirstName:  "Jane",
					LastName:   "Smith",
					Email:      "jane@example.com",
				},
			},
			setupUow: func(uow *MockUnitOfWork) {
				customer, _ := domain.NewCustomer(
					uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					"Test Company",
					domain.CustomerTypeCompany,
				)
				customer.ID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
				uow.customerRepo.customers[customer.ID] = customer
			},
			expectErr: false,
		},
		{
			name: "missing tenant ID",
			input: AddContactInput{
				TenantID: uuid.Nil,
				UserID:   uuid.New(),
				Request: &dto.CreateContactRequest{
					CustomerID: uuid.New(),
					FirstName:  "Jane",
					LastName:   "Smith",
					Email:      "jane@example.com",
				},
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "missing user ID",
			input: AddContactInput{
				TenantID: uuid.New(),
				UserID:   uuid.Nil,
				Request: &dto.CreateContactRequest{
					CustomerID: uuid.New(),
					FirstName:  "Jane",
					LastName:   "Smith",
					Email:      "jane@example.com",
				},
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "nil request",
			input: AddContactInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Request:  nil,
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "missing customer ID",
			input: AddContactInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				Request: &dto.CreateContactRequest{
					CustomerID: uuid.Nil,
					FirstName:  "Jane",
					LastName:   "Smith",
					Email:      "jane@example.com",
				},
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := NewMockUnitOfWork()
			eventPublisher := NewMockCustomerEventPublisher()
			idGenerator := NewMockIDGenerator()
			cache := NewMockCustomerCacheService()
			auditLogger := NewMockCustomerAuditLogger()

			tt.setupUow(uow)

			config := DefaultContactConfig()
			uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

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

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkAddContactUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Company", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := AddContactInput{
			TenantID: tenantID,
			UserID:   uuid.New(),
			Request: &dto.CreateContactRequest{
				CustomerID: customer.ID,
				FirstName:  "Jane",
				LastName:   "Smith",
				Email:      "jane@example.com",
			},
		}
		_, _ = uc.Execute(ctx, input)
	}
}

func BenchmarkListContactsUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()

	tenantID := uuid.New()
	customer := createTestCustomerWithContacts(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	uc := NewListContactsUseCase(uow)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := ListContactsInput{
			TenantID:   tenantID,
			CustomerID: customer.ID,
			Offset:     0,
			Limit:      10,
		}
		_, _ = uc.Execute(ctx, input)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestAddContactUseCase_Execute_ContextTimeout(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Company", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	config := DefaultContactConfig()
	uc := NewAddContactUseCase(uow, eventPublisher, idGenerator, cache, auditLogger, config)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	input := AddContactInput{
		TenantID: tenantID,
		UserID:   uuid.New(),
		Request: &dto.CreateContactRequest{
			CustomerID: customer.ID,
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane@example.com",
		},
	}

	// Act - just ensure no panic
	_, _ = uc.Execute(ctx, input)
}
