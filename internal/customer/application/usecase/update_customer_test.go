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
// UpdateCustomerUseCase Tests
// ============================================================================

func createTestCustomerForUpdate(tenantID uuid.UUID) *domain.Customer {
	customer, _ := domain.NewCustomer(tenantID, "Test Customer", domain.CustomerTypeCompany)
	customer.Email, _ = domain.NewEmail("test@example.com")
	customer.Status = domain.CustomerStatusActive
	customer.Version = 1
	return customer
}

func TestUpdateCustomerUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	userID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	newName := "Updated Customer Name"
	input := UpdateCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customer.ID,
		Request: &dto.UpdateCustomerRequest{
			Name:    &newName,
			Version: 1,
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
	if result.Name != newName {
		t.Errorf("Expected Name %s, got %s", newName, result.Name)
	}
}

func TestUpdateCustomerUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateCustomerInput{
		TenantID:   uuid.Nil, // Missing
		UserID:     uuid.New(),
		CustomerID: uuid.New(),
		Request: &dto.UpdateCustomerRequest{
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestUpdateCustomerUseCase_Execute_MissingUserID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.Nil, // Missing
		CustomerID: uuid.New(),
		Request: &dto.UpdateCustomerRequest{
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing user ID, got nil")
	}
}

func TestUpdateCustomerUseCase_Execute_MissingCustomerID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.Nil, // Missing
		Request: &dto.UpdateCustomerRequest{
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing customer ID, got nil")
	}
}

func TestUpdateCustomerUseCase_Execute_MissingRequest(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(),
		Request:    nil, // Missing
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing request, got nil")
	}
}

func TestUpdateCustomerUseCase_Execute_InvalidVersion(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(),
		Request: &dto.UpdateCustomerRequest{
			Version: 0, // Invalid
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid version, got nil")
	}
}

func TestUpdateCustomerUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := UpdateCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(), // Does not exist
		Request: &dto.UpdateCustomerRequest{
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

func TestUpdateCustomerUseCase_Execute_TenantMismatch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	customer := createTestCustomerForUpdate(tenantID1)
	uow.customerRepo.customers[customer.ID] = customer

	input := UpdateCustomerInput{
		TenantID:   tenantID2, // Different tenant
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.UpdateCustomerRequest{
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for tenant mismatch, got nil")
	}
}

func TestUpdateCustomerUseCase_Execute_VersionConflict(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Version = 5 // Current version is 5
	uow.customerRepo.customers[customer.ID] = customer

	input := UpdateCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.UpdateCustomerRequest{
			Version: 3, // Outdated version
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for version conflict, got nil")
	}
}

func TestUpdateCustomerUseCase_Execute_TransactionBeginError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.beginErr = errors.New("transaction begin failed")
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	input := UpdateCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.UpdateCustomerRequest{
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for transaction begin failure, got nil")
	}
}

func TestUpdateCustomerUseCase_Execute_TransactionCommitError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.commitErr = errors.New("transaction commit failed")
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	newName := "Updated Name"
	input := UpdateCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.UpdateCustomerRequest{
			Name:    &newName,
			Version: 1,
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
// ChangeCustomerStatusUseCase Tests
// ============================================================================

func TestChangeCustomerStatusUseCase_Execute_Activate_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewChangeCustomerStatusUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Status = domain.CustomerStatusInactive
	uow.customerRepo.customers[customer.ID] = customer

	input := ChangeStatusInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ChangeStatusRequest{
			Status:  domain.CustomerStatusActive,
			Version: 1,
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

func TestChangeCustomerStatusUseCase_Execute_Deactivate_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewChangeCustomerStatusUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Status = domain.CustomerStatusActive
	uow.customerRepo.customers[customer.ID] = customer

	input := ChangeStatusInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ChangeStatusRequest{
			Status:  domain.CustomerStatusInactive,
			Version: 1,
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
}

func TestChangeCustomerStatusUseCase_Execute_MarkChurned_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewChangeCustomerStatusUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Status = domain.CustomerStatusActive
	uow.customerRepo.customers[customer.ID] = customer

	input := ChangeStatusInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ChangeStatusRequest{
			Status:  domain.CustomerStatusChurned,
			Reason:  "Customer moved to competitor",
			Version: 1,
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
}

func TestChangeCustomerStatusUseCase_Execute_Block_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewChangeCustomerStatusUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Status = domain.CustomerStatusActive
	uow.customerRepo.customers[customer.ID] = customer

	input := ChangeStatusInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ChangeStatusRequest{
			Status:  domain.CustomerStatusBlocked,
			Reason:  "Fraudulent activity",
			Version: 1,
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
}

func TestChangeCustomerStatusUseCase_Execute_MissingCustomerID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewChangeCustomerStatusUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := ChangeStatusInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.Nil, // Missing
		Request: &dto.ChangeStatusRequest{
			Status:  domain.CustomerStatusActive,
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing customer ID, got nil")
	}
}

func TestChangeCustomerStatusUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewChangeCustomerStatusUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := ChangeStatusInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(), // Does not exist
		Request: &dto.ChangeStatusRequest{
			Status:  domain.CustomerStatusActive,
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

func TestChangeCustomerStatusUseCase_Execute_VersionConflict(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewChangeCustomerStatusUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Version = 5
	uow.customerRepo.customers[customer.ID] = customer

	input := ChangeStatusInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ChangeStatusRequest{
			Status:  domain.CustomerStatusActive,
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
// AssignCustomerOwnerUseCase Tests
// ============================================================================

func TestAssignCustomerOwnerUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewAssignCustomerOwnerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	newOwnerID := uuid.New()
	input := AssignOwnerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.AssignOwnerRequest{
			OwnerID: newOwnerID,
			Version: 1,
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

func TestAssignCustomerOwnerUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewAssignCustomerOwnerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := AssignOwnerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(), // Does not exist
		Request: &dto.AssignOwnerRequest{
			OwnerID: uuid.New(),
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

func TestAssignCustomerOwnerUseCase_Execute_TenantMismatch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewAssignCustomerOwnerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	customer := createTestCustomerForUpdate(tenantID1)
	uow.customerRepo.customers[customer.ID] = customer

	input := AssignOwnerInput{
		TenantID:   tenantID2, // Different tenant
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.AssignOwnerRequest{
			OwnerID: uuid.New(),
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for tenant mismatch, got nil")
	}
}

func TestAssignCustomerOwnerUseCase_Execute_VersionConflict(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewAssignCustomerOwnerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Version = 5
	uow.customerRepo.customers[customer.ID] = customer

	input := AssignOwnerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.AssignOwnerRequest{
			OwnerID: uuid.New(),
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
// ConvertCustomerUseCase Tests
// ============================================================================

func TestConvertCustomerUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewConvertCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Status = domain.CustomerStatusLead // Lead status
	uow.customerRepo.customers[customer.ID] = customer

	input := ConvertCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ConvertCustomerRequest{
			Reason:  "Lead qualified and ready to convert",
			Version: 1,
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

func TestConvertCustomerUseCase_Execute_CustomerNotFound(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewConvertCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	input := ConvertCustomerInput{
		TenantID:   uuid.New(),
		UserID:     uuid.New(),
		CustomerID: uuid.New(), // Does not exist
		Request: &dto.ConvertCustomerRequest{
			Reason:  "Test",
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

func TestConvertCustomerUseCase_Execute_TenantMismatch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewConvertCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	customer := createTestCustomerForUpdate(tenantID1)
	customer.Status = domain.CustomerStatusLead
	uow.customerRepo.customers[customer.ID] = customer

	input := ConvertCustomerInput{
		TenantID:   tenantID2, // Different tenant
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ConvertCustomerRequest{
			Reason:  "Test",
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for tenant mismatch, got nil")
	}
}

func TestConvertCustomerUseCase_Execute_VersionConflict(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewConvertCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Status = domain.CustomerStatusLead
	customer.Version = 5
	uow.customerRepo.customers[customer.ID] = customer

	input := ConvertCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ConvertCustomerRequest{
			Reason:  "Test",
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

func TestConvertCustomerUseCase_Execute_TransactionBeginError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uow.beginErr = errors.New("transaction begin failed")
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewConvertCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	customer.Status = domain.CustomerStatusLead
	uow.customerRepo.customers[customer.ID] = customer

	input := ConvertCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.ConvertCustomerRequest{
			Reason:  "Test",
			Version: 1,
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for transaction begin failure, got nil")
	}
}

// ============================================================================
// Table-Driven Tests
// ============================================================================

func TestUpdateCustomerUseCase_Execute_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		input     UpdateCustomerInput
		setupUow  func(*MockUnitOfWork)
		expectErr bool
	}{
		{
			name: "valid update",
			input: UpdateCustomerInput{
				TenantID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				UserID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				CustomerID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				Request: &dto.UpdateCustomerRequest{
					Version: 1,
				},
			},
			setupUow: func(uow *MockUnitOfWork) {
				customer := createTestCustomerForUpdate(uuid.MustParse("11111111-1111-1111-1111-111111111111"))
				customer.ID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
				uow.customerRepo.customers[customer.ID] = customer
			},
			expectErr: false,
		},
		{
			name: "missing tenant ID",
			input: UpdateCustomerInput{
				TenantID:   uuid.Nil,
				UserID:     uuid.New(),
				CustomerID: uuid.New(),
				Request: &dto.UpdateCustomerRequest{
					Version: 1,
				},
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "missing user ID",
			input: UpdateCustomerInput{
				TenantID:   uuid.New(),
				UserID:     uuid.Nil,
				CustomerID: uuid.New(),
				Request: &dto.UpdateCustomerRequest{
					Version: 1,
				},
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "missing customer ID",
			input: UpdateCustomerInput{
				TenantID:   uuid.New(),
				UserID:     uuid.New(),
				CustomerID: uuid.Nil,
				Request: &dto.UpdateCustomerRequest{
					Version: 1,
				},
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "nil request",
			input: UpdateCustomerInput{
				TenantID:   uuid.New(),
				UserID:     uuid.New(),
				CustomerID: uuid.New(),
				Request:    nil,
			},
			setupUow:  func(uow *MockUnitOfWork) {},
			expectErr: true,
		},
		{
			name: "zero version",
			input: UpdateCustomerInput{
				TenantID:   uuid.New(),
				UserID:     uuid.New(),
				CustomerID: uuid.New(),
				Request: &dto.UpdateCustomerRequest{
					Version: 0,
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

			uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

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

func BenchmarkUpdateCustomerUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)
	ctx := context.Background()

	tenantID := uuid.New()
	userID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	newName := "Updated Name"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := UpdateCustomerInput{
			TenantID:   tenantID,
			UserID:     userID,
			CustomerID: customer.ID,
			Request: &dto.UpdateCustomerRequest{
				Name:    &newName,
				Version: customer.Version,
			},
		}
		_, _ = uc.Execute(ctx, input)
	}
}

func BenchmarkChangeCustomerStatusUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewChangeCustomerStatusUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)
	ctx := context.Background()

	tenantID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		customer := createTestCustomerForUpdate(tenantID)
		customer.Status = domain.CustomerStatusInactive
		uow.customerRepo.customers[customer.ID] = customer

		input := ChangeStatusInput{
			TenantID:   tenantID,
			UserID:     uuid.New(),
			CustomerID: customer.ID,
			Request: &dto.ChangeStatusRequest{
				Status:  domain.CustomerStatusActive,
				Version: 1,
			},
		}
		_, _ = uc.Execute(ctx, input)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestUpdateCustomerUseCase_Execute_ContextTimeout(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	uc := NewUpdateCustomerUseCase(uow, eventPublisher, idGenerator, cache, auditLogger)

	tenantID := uuid.New()
	customer := createTestCustomerForUpdate(tenantID)
	uow.customerRepo.customers[customer.ID] = customer

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	newName := "Updated Name"
	input := UpdateCustomerInput{
		TenantID:   tenantID,
		UserID:     uuid.New(),
		CustomerID: customer.ID,
		Request: &dto.UpdateCustomerRequest{
			Name:    &newName,
			Version: 1,
		},
	}

	// Act - just ensure no panic
	_, _ = uc.Execute(ctx, input)
}
