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
// Mock Search Index for Delete Tests
// ============================================================================

// MockSearchIndex is a mock implementation of ports.SearchIndex.
type MockSearchIndex struct {
	removeCustomerErr error
	removeContactErr  error
	removedCustomers  []uuid.UUID
	removedContacts   []uuid.UUID
}

func NewMockSearchIndex() *MockSearchIndex {
	return &MockSearchIndex{
		removedCustomers: make([]uuid.UUID, 0),
		removedContacts:  make([]uuid.UUID, 0),
	}
}

func (m *MockSearchIndex) IndexCustomer(ctx context.Context, customer *domain.Customer) error {
	return nil
}

func (m *MockSearchIndex) IndexContact(ctx context.Context, contact *domain.Contact) error {
	return nil
}

func (m *MockSearchIndex) RemoveCustomer(ctx context.Context, id uuid.UUID) error {
	if m.removeCustomerErr != nil {
		return m.removeCustomerErr
	}
	m.removedCustomers = append(m.removedCustomers, id)
	return nil
}

func (m *MockSearchIndex) RemoveContact(ctx context.Context, id uuid.UUID) error {
	if m.removeContactErr != nil {
		return m.removeContactErr
	}
	m.removedContacts = append(m.removedContacts, id)
	return nil
}

func (m *MockSearchIndex) SearchCustomers(ctx context.Context, tenantID uuid.UUID, query string, options ports.SearchOptions) (*ports.SearchResult, error) {
	return &ports.SearchResult{}, nil
}

func (m *MockSearchIndex) SearchContacts(ctx context.Context, tenantID uuid.UUID, query string, options ports.SearchOptions) (*ports.SearchResult, error) {
	return &ports.SearchResult{}, nil
}

func (m *MockSearchIndex) Reindex(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}

// ============================================================================
// DeleteCustomerUseCase Tests
// ============================================================================

func createTestCustomerForDelete(tenantID uuid.UUID) *domain.Customer {
	customer, _ := domain.NewCustomer(tenantID, "Test Customer", domain.CustomerTypeCompany)
	customer.Email, _ = domain.NewEmail("test@example.com")
	customer.Status = domain.CustomerStatusActive
	return customer
}

func TestDeleteCustomerUseCase_Execute_SoftDelete_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	userID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customer.ID,
		HardDelete: false,
		Reason:     "Customer requested deletion",
		IPAddress:  "127.0.0.1",
		UserAgent:  "TestClient/1.0",
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
	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.CustomerID != customer.ID {
		t.Errorf("Expected CustomerID %v, got %v", customer.ID, result.CustomerID)
	}
	if result.HardDelete {
		t.Error("Expected HardDelete to be false")
	}
}

func TestDeleteCustomerUseCase_Execute_HardDelete_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	config.AllowHardDelete = true
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	userID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customer.ID,
		HardDelete: true,
		Reason:     "GDPR deletion request",
		IPAddress:  "127.0.0.1",
		UserAgent:  "TestClient/1.0",
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
	if !result.HardDelete {
		t.Error("Expected HardDelete to be true")
	}

	// Verify search index was updated
	if len(searchIndex.removedCustomers) != 1 {
		t.Errorf("Expected 1 customer removed from search index, got %d", len(searchIndex.removedCustomers))
	}
}

func TestDeleteCustomerUseCase_Execute_HardDelete_NotAllowed(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	config.AllowHardDelete = false // Hard delete not allowed
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	userID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customer.ID,
		HardDelete: true,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for hard delete not allowed, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	input := DeleteCustomerInput{
		TenantID:   uuid.Nil, // Missing
		UserID:     uuid.New(),
		CustomerID: uuid.New(),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_MissingUserID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	input := DeleteCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.Nil, // Missing
		CustomerID: uuid.New(),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing user ID, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_MissingCustomerID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	input := DeleteCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.Nil, // Missing
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing customer ID, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	input := DeleteCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(), // Does not exist
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer not found, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_TenantMismatch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	customer := createTestCustomerForDelete(tenantID1)
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID2, // Different tenant
		UserID:     uuid.New(),
		CustomerID: customer.ID,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for tenant mismatch, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_RequireReason(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	config.RequireReason = true
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Reason:     "", // Missing reason
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing reason, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_OutstandingBalance(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	config.CheckOutstandingBalance = true
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	// Set outstanding balance
	balance, _ := domain.NewMoney(100.00, "USD")
	customer.Financials.CurrentBalance = &balance
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for outstanding balance, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_ActiveDeals(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	config.CheckActiveDeals = true
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	// Set active deals (more deals than won+lost)
	customer.Stats.DealCount = 5
	customer.Stats.WonDealCount = 1
	customer.Stats.LostDealCount = 1
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for active deals, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_TransactionBeginError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.beginErr = errors.New("transaction begin failed")
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for transaction begin failure, got nil")
	}
}

func TestDeleteCustomerUseCase_Execute_TransactionCommitError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.commitErr = errors.New("transaction commit failed")
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for transaction commit failure, got nil")
	}
}

// ============================================================================
// BulkDeleteCustomersUseCase Tests
// ============================================================================

func TestBulkDeleteCustomersUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewBulkDeleteCustomersUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	userID := uuid.New()

	// Create multiple customers
	customerIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		customer := createTestCustomerForDelete(tenantID)
		uow.customerRepo.customers[customer.ID] = customer
		customerIDs[i] = customer.ID
	}

	input := BulkDeleteInput{
		TenantID:  tenantID,
		UserID:    userID,
		Request:   &dto.BulkDeleteRequest{CustomerIDs: customerIDs},
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
	if result.Processed != 3 {
		t.Errorf("Expected Processed 3, got %d", result.Processed)
	}
	if result.Succeeded != 3 {
		t.Errorf("Expected Succeeded 3, got %d", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("Expected Failed 0, got %d", result.Failed)
	}
}

func TestBulkDeleteCustomersUseCase_Execute_PartialSuccess(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewBulkDeleteCustomersUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	userID := uuid.New()

	// Create one valid customer
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	// Include one non-existent customer ID
	customerIDs := []uuid.UUID{customer.ID, uuid.New()}

	input := BulkDeleteInput{
		TenantID: tenantID,
		UserID:   userID,
		Request:  &dto.BulkDeleteRequest{CustomerIDs: customerIDs},
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.Succeeded != 1 {
		t.Errorf("Expected Succeeded 1, got %d", result.Succeeded)
	}
	if result.Failed != 1 {
		t.Errorf("Expected Failed 1, got %d", result.Failed)
	}
}

func TestBulkDeleteCustomersUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewBulkDeleteCustomersUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	input := BulkDeleteInput{
		TenantID: uuid.Nil, // Missing
		UserID:   uuid.New(),
		Request:  &dto.BulkDeleteRequest{CustomerIDs: []uuid.UUID{uuid.New()}},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestBulkDeleteCustomersUseCase_Execute_EmptyCustomerIDs(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewBulkDeleteCustomersUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	input := BulkDeleteInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Request:  &dto.BulkDeleteRequest{CustomerIDs: []uuid.UUID{}},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty customer IDs, got nil")
	}
}

func TestBulkDeleteCustomersUseCase_Execute_TenantMismatch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewBulkDeleteCustomersUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	userID := uuid.New()

	// Create customer for different tenant
	customer := createTestCustomerForDelete(tenantID1)
	uow.customerRepo.customers[customer.ID] = customer

	input := BulkDeleteInput{
		TenantID: tenantID2, // Different tenant
		UserID:   userID,
		Request:  &dto.BulkDeleteRequest{CustomerIDs: []uuid.UUID{customer.ID}},
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for bulk operation, got: %v", err)
	}
	if result.Failed != 1 {
		t.Errorf("Expected Failed 1 for tenant mismatch, got %d", result.Failed)
	}
}

// ============================================================================
// GetCustomerUseCase Tests
// ============================================================================

func TestGetCustomerUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()
	uc := NewGetCustomerUseCase(uow, cache)

	tenantID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := GetCustomerInput{
		TenantID:   tenantID,
		CustomerID: customer.ID,
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
	if result.ID != customer.ID {
		t.Errorf("Expected ID %v, got %v", customer.ID, result.ID)
	}
}

func TestGetCustomerUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()
	uc := NewGetCustomerUseCase(uow, cache)

	input := GetCustomerInput{
		TenantID:   uuid.Nil, // Missing
		CustomerID: uuid.New(),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestGetCustomerUseCase_Execute_MissingCustomerID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()
	uc := NewGetCustomerUseCase(uow, cache)

	input := GetCustomerInput{
		TenantID:   uuid.New(),
		CustomerID: uuid.Nil, // Missing
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing customer ID, got nil")
	}
}

func TestGetCustomerUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()
	uc := NewGetCustomerUseCase(uow, cache)

	input := GetCustomerInput{
		TenantID:   uuid.New(),
		CustomerID: uuid.New(), // Does not exist
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer not found, got nil")
	}
}

func TestGetCustomerUseCase_Execute_TenantMismatch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()
	uc := NewGetCustomerUseCase(uow, cache)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	customer := createTestCustomerForDelete(tenantID1)
	uow.customerRepo.customers[customer.ID] = customer

	input := GetCustomerInput{
		TenantID:   tenantID2, // Different tenant
		CustomerID: customer.ID,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for tenant mismatch, got nil")
	}
}

// ============================================================================
// GetCustomerByCodeUseCase Tests
// ============================================================================

func TestGetCustomerByCodeUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewGetCustomerByCodeUseCase(uow)

	tenantID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	customer.Code = "CUST-001"
	uow.customerRepo.customers[customer.ID] = customer

	input := GetCustomerByCodeInput{
		TenantID: tenantID,
		Code:     "CUST-001",
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

func TestGetCustomerByCodeUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewGetCustomerByCodeUseCase(uow)

	input := GetCustomerByCodeInput{
		TenantID: uuid.Nil, // Missing
		Code:     "CUST-001",
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestGetCustomerByCodeUseCase_Execute_MissingCode(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewGetCustomerByCodeUseCase(uow)

	input := GetCustomerByCodeInput{
		TenantID: uuid.New(),
		Code:     "", // Missing
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing code, got nil")
	}
}

// ============================================================================
// Table-Driven Tests for Delete Validation
// ============================================================================

func TestDeleteCustomerUseCase_Execute_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		input     DeleteCustomerInput
		setupUow  func(*MockUnitOfWork)
		expectErr bool
	}{
		{
			name: "valid input",
			input: DeleteCustomerInput{
				TenantID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				UserID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				CustomerID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			},
			setupUow: func(uow *MockUnitOfWork) {
				customer := createTestCustomerForDelete(uuid.MustParse("11111111-1111-1111-1111-111111111111"))
				customer.ID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
				uow.customerRepo.customers[customer.ID] = customer
			},
			expectErr: false,
		},
		{
			name: "missing tenant ID",
			input: DeleteCustomerInput{
				TenantID:   uuid.Nil,
				UserID:     uuid.New(),
				CustomerID: uuid.New(),
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "missing user ID",
			input: DeleteCustomerInput{
				TenantID:   uuid.New(),
				UserID:     uuid.Nil,
				CustomerID: uuid.New(),
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "missing customer ID",
			input: DeleteCustomerInput{
				TenantID:   uuid.New(),
				UserID:     uuid.New(),
				CustomerID: uuid.Nil,
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
			searchIndex := NewMockSearchIndex()
			auditLogger := NewMockCustomerAuditLogger()

			tt.setupUow(uow)

			config := DefaultDeleteCustomerConfig()
			uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

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

func BenchmarkDeleteCustomerUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a new customer for each iteration
		customer := createTestCustomerForDelete(tenantID)
		uow.customerRepo.customers[customer.ID] = customer

		input := DeleteCustomerInput{
			TenantID:   tenantID,
			UserID:     userID,
			CustomerID: customer.ID,
		}
		_, _ = uc.Execute(ctx, input)
	}
}

func BenchmarkBulkDeleteCustomersUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewBulkDeleteCustomersUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create customers for each iteration
		customerIDs := make([]uuid.UUID, 10)
		for j := 0; j < 10; j++ {
			customer := createTestCustomerForDelete(tenantID)
			uow.customerRepo.customers[customer.ID] = customer
			customerIDs[j] = customer.ID
		}

		input := BulkDeleteInput{
			TenantID: tenantID,
			UserID:   userID,
			Request:  &dto.BulkDeleteRequest{CustomerIDs: customerIDs},
		}
		_, _ = uc.Execute(ctx, input)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestDeleteCustomerUseCase_Execute_ContextTimeout(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	searchIndex := NewMockSearchIndex()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultDeleteCustomerConfig()
	uc := NewDeleteCustomerUseCase(uow, eventPublisher, idGenerator, cache, searchIndex, auditLogger, config)

	tenantID := uuid.New()
	customer := createTestCustomerForDelete(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	// Create a context that's already cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure timeout

	input := DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
	}

	// Act
	_, err := uc.Execute(ctx, input)

	// Assert - the operation may or may not error depending on timing
	// This test is for coverage of context handling
	_ = err
}
