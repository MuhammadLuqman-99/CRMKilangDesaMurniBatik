// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/ports"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Mock Import Service
// ============================================================================

// MockImportService is a mock implementation of ports.ImportService.
type MockImportService struct {
	parseCustomersResult     []*ports.CustomerImportRow
	parseCustomersErr        error
	parseContactsResult      []*ports.ContactImportRow
	parseContactsErr         error
	validateResult           []ports.ImportValidationError
	validateErr              error
}

func NewMockImportService() *MockImportService {
	return &MockImportService{
		parseCustomersResult: make([]*ports.CustomerImportRow, 0),
		validateResult:       make([]ports.ImportValidationError, 0),
	}
}

func (m *MockImportService) ParseCustomers(ctx context.Context, data io.Reader, format ports.ExportFormat, fieldMapping map[string]string) ([]*ports.CustomerImportRow, error) {
	if m.parseCustomersErr != nil {
		return nil, m.parseCustomersErr
	}
	return m.parseCustomersResult, nil
}

func (m *MockImportService) ParseContacts(ctx context.Context, data io.Reader, format ports.ExportFormat, fieldMapping map[string]string) ([]*ports.ContactImportRow, error) {
	if m.parseContactsErr != nil {
		return nil, m.parseContactsErr
	}
	return m.parseContactsResult, nil
}

func (m *MockImportService) ValidateImportData(ctx context.Context, rows interface{}) ([]ports.ImportValidationError, error) {
	if m.validateErr != nil {
		return nil, m.validateErr
	}
	return m.validateResult, nil
}

// ============================================================================
// Mock Export Service
// ============================================================================

// MockExportService is a mock implementation of ports.ExportService.
type MockExportService struct {
	exportCustomersResult []byte
	exportCustomersErr    error
	exportContactsResult  []byte
	exportContactsErr     error
	streamExportErr       error
}

func NewMockExportService() *MockExportService {
	return &MockExportService{
		exportCustomersResult: []byte("id,name,email\n1,Test,test@example.com"),
	}
}

func (m *MockExportService) ExportCustomers(ctx context.Context, tenantID uuid.UUID, customers []*domain.Customer, format ports.ExportFormat, fields []string) ([]byte, error) {
	if m.exportCustomersErr != nil {
		return nil, m.exportCustomersErr
	}
	return m.exportCustomersResult, nil
}

func (m *MockExportService) ExportContacts(ctx context.Context, tenantID uuid.UUID, contacts []*domain.Contact, format ports.ExportFormat, fields []string) ([]byte, error) {
	if m.exportContactsErr != nil {
		return nil, m.exportContactsErr
	}
	return m.exportContactsResult, nil
}

func (m *MockExportService) StreamExport(ctx context.Context, tenantID uuid.UUID, query string, format ports.ExportFormat, writer io.Writer) error {
	if m.streamExportErr != nil {
		return m.streamExportErr
	}
	_, err := writer.Write([]byte("streamed data"))
	return err
}

// ============================================================================
// Mock Import Repository with Full Implementation
// ============================================================================

// MockImportRepositoryFull is a more complete mock implementation.
type MockImportRepositoryFull struct {
	imports     map[uuid.UUID]*domain.Import
	importErrs  []*domain.ImportError
	createErr   error
	updateErr   error
	findByIDErr error
}

func NewMockImportRepositoryFull() *MockImportRepositoryFull {
	return &MockImportRepositoryFull{
		imports:    make(map[uuid.UUID]*domain.Import),
		importErrs: make([]*domain.ImportError, 0),
	}
}

func (m *MockImportRepositoryFull) CreateImport(ctx context.Context, imp *domain.Import) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.imports[imp.ID] = imp
	return nil
}

func (m *MockImportRepositoryFull) UpdateImport(ctx context.Context, imp *domain.Import) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.imports[imp.ID] = imp
	return nil
}

func (m *MockImportRepositoryFull) FindImportByID(ctx context.Context, id uuid.UUID) (*domain.Import, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	imp, ok := m.imports[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return imp, nil
}

func (m *MockImportRepositoryFull) FindImportsByTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]*domain.Import, error) {
	var result []*domain.Import
	for _, imp := range m.imports {
		if imp.TenantID == tenantID {
			result = append(result, imp)
		}
	}
	return result, nil
}

func (m *MockImportRepositoryFull) CreateImportError(ctx context.Context, err *domain.ImportError) error {
	m.importErrs = append(m.importErrs, err)
	return nil
}

func (m *MockImportRepositoryFull) FindImportErrors(ctx context.Context, importID uuid.UUID) ([]*domain.ImportError, error) {
	return m.importErrs, nil
}

// ============================================================================
// ImportCustomersUseCase Tests
// ============================================================================

func createValidImportRows() []*ports.CustomerImportRow {
	return []*ports.CustomerImportRow{
		{
			RowNumber: 1,
			Name:      "Customer One",
			Type:      "company",
			Email:     "one@example.com",
			Phone:     "+1234567890",
		},
		{
			RowNumber: 2,
			Name:      "Customer Two",
			Type:      "individual",
			Email:     "two@example.com",
			Phone:     "+0987654321",
		},
	}
}

func TestImportCustomersUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	importService.parseCustomersResult = createValidImportRows()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	tenantID := uuid.New()
	userID := uuid.New()

	input := ImportCustomersInput{
		TenantID:       tenantID,
		UserID:         userID,
		FileName:       "customers.csv",
		FileSize:       1024,
		Format:         "csv",
		Data:           []byte("name,email\nCustomer One,one@example.com"),
		SkipDuplicates: true,
		IPAddress:      "127.0.0.1",
		UserAgent:      "TestClient/1.0",
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
	if result.TotalRows != 2 {
		t.Errorf("Expected TotalRows 2, got %d", result.TotalRows)
	}
	if result.SuccessCount != 2 {
		t.Errorf("Expected SuccessCount 2, got %d", result.SuccessCount)
	}
}

func TestImportCustomersUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID: uuid.Nil, // Missing
		UserID:   uuid.New(),
		FileName: "customers.csv",
		FileSize: 1024,
		Format:   "csv",
		Data:     []byte("test data"),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestImportCustomersUseCase_Execute_MissingUserID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.Nil, // Missing
		FileName: "customers.csv",
		FileSize: 1024,
		Format:   "csv",
		Data:     []byte("test data"),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing user ID, got nil")
	}
}

func TestImportCustomersUseCase_Execute_EmptyData(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "customers.csv",
		FileSize: 0,
		Format:   "csv",
		Data:     []byte{}, // Empty
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty data, got nil")
	}
}

func TestImportCustomersUseCase_Execute_FileTooLarge(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	config.MaxFileSize = 100 // Very small limit
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "customers.csv",
		FileSize: 1000, // Exceeds limit
		Format:   "csv",
		Data:     []byte("test data"),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for file too large, got nil")
	}
}

func TestImportCustomersUseCase_Execute_InvalidFormat(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "customers.txt",
		FileSize: 1024,
		Format:   "txt", // Unsupported format
		Data:     []byte("test data"),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid format, got nil")
	}
}

func TestImportCustomersUseCase_Execute_ParseError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	importService.parseCustomersErr = errors.New("parse error")
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "customers.csv",
		FileSize: 1024,
		Format:   "csv",
		Data:     []byte("invalid data"),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for parse failure, got nil")
	}
}

func TestImportCustomersUseCase_Execute_ValidationErrors(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	importService.parseCustomersResult = createValidImportRows()
	// Add validation errors for more than half the rows
	importService.validateResult = []ports.ImportValidationError{
		{RowNumber: 1, Field: "email", Error: "invalid email", Severity: "error"},
		{RowNumber: 2, Field: "email", Error: "invalid email", Severity: "error"},
	}
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	config.ValidateBeforeImport = true
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "customers.csv",
		FileSize: 1024,
		Format:   "csv",
		Data:     []byte("test data"),
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for too many validation errors, got nil")
	}
}

func TestImportCustomersUseCase_Execute_WithDuplicateSkip(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	importService.parseCustomersResult = []*ports.CustomerImportRow{
		{
			RowNumber: 1,
			Name:      "Existing Customer",
			Email:     "existing@example.com",
		},
	}
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	// Add existing customer
	tenantID := uuid.New()
	existingCustomer, _ := domain.NewCustomer(tenantID, "Existing Customer", domain.CustomerTypeCompany)
	existingCustomer.Email, _ = domain.NewEmail("existing@example.com")
	uow.customerRepo.customers[existingCustomer.ID] = existingCustomer
	uow.customerRepo.existsByEmailVal = true

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID:       tenantID,
		UserID:         uuid.New(),
		FileName:       "customers.csv",
		FileSize:       1024,
		Format:         "csv",
		Data:           []byte("test data"),
		SkipDuplicates: true,
		UpdateExisting: false,
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.SkippedCount != 1 {
		t.Errorf("Expected SkippedCount 1, got %d", result.SkippedCount)
	}
}

func TestImportCustomersUseCase_Execute_WithDefaultOwner(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	importService.parseCustomersResult = []*ports.CustomerImportRow{
		{
			RowNumber: 1,
			Name:      "New Customer",
			Type:      "company",
			Email:     "new@example.com",
		},
	}
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	tenantID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()

	input := ImportCustomersInput{
		TenantID:     tenantID,
		UserID:       userID,
		FileName:     "customers.csv",
		FileSize:     1024,
		Format:       "csv",
		Data:         []byte("test data"),
		DefaultOwner: &ownerID,
		DefaultTags:  []string{"imported", "2024"},
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.SuccessCount != 1 {
		t.Errorf("Expected SuccessCount 1, got %d", result.SuccessCount)
	}
}

func TestImportCustomersUseCase_Execute_RowWithErrors(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	importService.parseCustomersResult = []*ports.CustomerImportRow{
		{
			RowNumber: 1,
			Name:      "Valid Customer",
			Type:      "company",
			Email:     "valid@example.com",
		},
		{
			RowNumber: 2,
			Name:      "Invalid Customer",
			Errors:    []string{"missing required field"},
		},
	}
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	config.ValidateBeforeImport = false // Skip validation to test row-level errors
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	input := ImportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "customers.csv",
		FileSize: 1024,
		Format:   "csv",
		Data:     []byte("test data"),
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.SuccessCount != 1 {
		t.Errorf("Expected SuccessCount 1, got %d", result.SuccessCount)
	}
	if result.FailureCount != 1 {
		t.Errorf("Expected FailureCount 1, got %d", result.FailureCount)
	}
}

// ============================================================================
// ExportCustomersUseCase Tests
// ============================================================================

func TestExportCustomersUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	// Add customers to export
	tenantID := uuid.New()
	customer1, _ := domain.NewCustomer(tenantID, "Customer One", domain.CustomerTypeCompany)
	customer2, _ := domain.NewCustomer(tenantID, "Customer Two", domain.CustomerTypeIndividual)
	uow.customerRepo.customers[customer1.ID] = customer1
	uow.customerRepo.customers[customer2.ID] = customer2

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID:  tenantID,
		UserID:    uuid.New(),
		Format:    "csv",
		Fields:    []string{"name", "email"},
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
	if len(result.Data) == 0 {
		t.Error("Expected Data to not be empty")
	}
	if result.ContentType != "text/csv" {
		t.Errorf("Expected ContentType text/csv, got %s", result.ContentType)
	}
}

func TestExportCustomersUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID: uuid.Nil, // Missing
		UserID:   uuid.New(),
		Format:   "csv",
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestExportCustomersUseCase_Execute_MissingUserID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.Nil, // Missing
		Format:   "csv",
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing user ID, got nil")
	}
}

func TestExportCustomersUseCase_Execute_InvalidFormat(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Format:   "txt", // Unsupported
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid format, got nil")
	}
}

func TestExportCustomersUseCase_Execute_NoCustomers(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Format:   "csv",
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for no customers to export, got nil")
	}
}

func TestExportCustomersUseCase_Execute_XLSXFormat(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Customer", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID: tenantID,
		UserID:   uuid.New(),
		Format:   "xlsx",
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.ContentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Errorf("Expected XLSX content type, got %s", result.ContentType)
	}
}

func TestExportCustomersUseCase_Execute_JSONFormat(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Customer", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID: tenantID,
		UserID:   uuid.New(),
		Format:   "json",
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.ContentType != "application/json" {
		t.Errorf("Expected JSON content type, got %s", result.ContentType)
	}
}

func TestExportCustomersUseCase_Execute_WithFilter(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Active Customer", domain.CustomerTypeCompany)
	customer.Status = domain.CustomerStatusActive
	uow.customerRepo.customers[customer.ID] = customer

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID: tenantID,
		UserID:   uuid.New(),
		Format:   "csv",
		Filter: &dto.SearchCustomersRequest{
			Statuses: []domain.CustomerStatus{domain.CustomerStatusActive},
		},
	}

	// Act
	result, err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.TotalRows == 0 {
		t.Error("Expected at least one row in export")
	}
}

func TestExportCustomersUseCase_Execute_ExportServiceError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	exportService.exportCustomersErr = errors.New("export failed")
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Test Customer", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := ExportCustomersInput{
		TenantID: tenantID,
		UserID:   uuid.New(),
		Format:   "csv",
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for export service failure, got nil")
	}
}

// ============================================================================
// StreamExportCustomersUseCase Tests
// ============================================================================

func TestStreamExportCustomersUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultExportConfig()
	uc := NewStreamExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	var buf bytes.Buffer
	input := StreamExportInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Format:   "csv",
		Query:    "test",
		Writer:   &buf,
	}

	// Act
	err := uc.Execute(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Expected data written to buffer")
	}
}

func TestStreamExportCustomersUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultExportConfig()
	uc := NewStreamExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	var buf bytes.Buffer
	input := StreamExportInput{
		TenantID: uuid.Nil, // Missing
		UserID:   uuid.New(),
		Format:   "csv",
		Writer:   &buf,
	}

	// Act
	err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestStreamExportCustomersUseCase_Execute_MissingWriter(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultExportConfig()
	uc := NewStreamExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	input := StreamExportInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Format:   "csv",
		Writer:   nil, // Missing
	}

	// Act
	err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing writer, got nil")
	}
}

func TestStreamExportCustomersUseCase_Execute_ExportError(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	exportService.streamExportErr = errors.New("stream export failed")
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultExportConfig()
	uc := NewStreamExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

	var buf bytes.Buffer
	input := StreamExportInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		Format:   "csv",
		Writer:   &buf,
	}

	// Act
	err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for stream export failure, got nil")
	}
}

// ============================================================================
// Table-Driven Tests
// ============================================================================

func TestImportCustomersUseCase_Execute_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		input     ImportCustomersInput
		expectErr bool
	}{
		{
			name: "valid CSV import",
			input: ImportCustomersInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				FileName: "customers.csv",
				FileSize: 1024,
				Format:   "csv",
				Data:     []byte("test data"),
			},
			expectErr: false,
		},
		{
			name: "valid XLSX import",
			input: ImportCustomersInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				FileName: "customers.xlsx",
				FileSize: 1024,
				Format:   "xlsx",
				Data:     []byte("test data"),
			},
			expectErr: false,
		},
		{
			name: "valid JSON import",
			input: ImportCustomersInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				FileName: "customers.json",
				FileSize: 1024,
				Format:   "json",
				Data:     []byte("test data"),
			},
			expectErr: false,
		},
		{
			name: "missing tenant ID",
			input: ImportCustomersInput{
				TenantID: uuid.Nil,
				UserID:   uuid.New(),
				FileName: "customers.csv",
				FileSize: 1024,
				Format:   "csv",
				Data:     []byte("test data"),
			},
			expectErr: true,
		},
		{
			name: "missing user ID",
			input: ImportCustomersInput{
				TenantID: uuid.New(),
				UserID:   uuid.Nil,
				FileName: "customers.csv",
				FileSize: 1024,
				Format:   "csv",
				Data:     []byte("test data"),
			},
			expectErr: true,
		},
		{
			name: "empty data",
			input: ImportCustomersInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				FileName: "customers.csv",
				FileSize: 0,
				Format:   "csv",
				Data:     []byte{},
			},
			expectErr: true,
		},
		{
			name: "unsupported format",
			input: ImportCustomersInput{
				TenantID: uuid.New(),
				UserID:   uuid.New(),
				FileName: "customers.txt",
				FileSize: 1024,
				Format:   "txt",
				Data:     []byte("test data"),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := NewMockUnitOfWork()
			importService := NewMockImportService()
			importService.parseCustomersResult = createValidImportRows()
			eventPublisher := NewMockCustomerEventPublisher()
			idGenerator := NewMockIDGenerator()
			cache := NewMockCustomerCacheService()
			auditLogger := NewMockCustomerAuditLogger()

			config := DefaultImportConfig()
			uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

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

func TestExportCustomersUseCase_Execute_FormatCases(t *testing.T) {
	tests := []struct {
		name            string
		format          string
		expectedContent string
		expectErr       bool
	}{
		{
			name:            "CSV format",
			format:          "csv",
			expectedContent: "text/csv",
			expectErr:       false,
		},
		{
			name:            "XLSX format",
			format:          "xlsx",
			expectedContent: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			expectErr:       false,
		},
		{
			name:            "JSON format",
			format:          "json",
			expectedContent: "application/json",
			expectErr:       false,
		},
		{
			name:            "Unsupported format",
			format:          "txt",
			expectedContent: "",
			expectErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := NewMockUnitOfWork()
			exportService := NewMockExportService()
			idGenerator := NewMockIDGenerator()
			auditLogger := NewMockCustomerAuditLogger()

			tenantID := uuid.New()
			customer, _ := domain.NewCustomer(tenantID, "Test", domain.CustomerTypeCompany)
			uow.customerRepo.customers[customer.ID] = customer

			config := DefaultExportConfig()
			uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)

			input := ExportCustomersInput{
				TenantID: tenantID,
				UserID:   uuid.New(),
				Format:   tt.format,
			}

			result, err := uc.Execute(context.Background(), input)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
			if !tt.expectErr && result.ContentType != tt.expectedContent {
				t.Errorf("Expected content type %s, got %s", tt.expectedContent, result.ContentType)
			}
		})
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkImportCustomersUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	importService.parseCustomersResult = createValidImportRows()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := ImportCustomersInput{
			TenantID: uuid.New(),
			UserID:   uuid.New(),
			FileName: "customers.csv",
			FileSize: 1024,
			Format:   "csv",
			Data:     []byte("test data"),
		}
		_, _ = uc.Execute(ctx, input)
	}
}

func BenchmarkExportCustomersUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	exportService := NewMockExportService()
	idGenerator := NewMockIDGenerator()
	auditLogger := NewMockCustomerAuditLogger()

	tenantID := uuid.New()
	for i := 0; i < 100; i++ {
		customer, _ := domain.NewCustomer(tenantID, "Customer", domain.CustomerTypeCompany)
		uow.customerRepo.customers[customer.ID] = customer
	}

	config := DefaultExportConfig()
	uc := NewExportCustomersUseCase(uow, exportService, idGenerator, auditLogger, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := ExportCustomersInput{
			TenantID: tenantID,
			UserID:   uuid.New(),
			Format:   "csv",
		}
		_, _ = uc.Execute(ctx, input)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestImportCustomersUseCase_Execute_ContextTimeout(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	importService := NewMockImportService()
	importService.parseCustomersResult = createValidImportRows()
	eventPublisher := NewMockCustomerEventPublisher()
	idGenerator := NewMockIDGenerator()
	cache := NewMockCustomerCacheService()
	auditLogger := NewMockCustomerAuditLogger()

	config := DefaultImportConfig()
	uc := NewImportCustomersUseCase(uow, importService, eventPublisher, idGenerator, cache, auditLogger, config)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	input := ImportCustomersInput{
		TenantID: uuid.New(),
		UserID:   uuid.New(),
		FileName: "customers.csv",
		FileSize: 1024,
		Format:   "csv",
		Data:     []byte("test data"),
	}

	// Act - just ensure no panic
	_, _ = uc.Execute(ctx, input)
}
