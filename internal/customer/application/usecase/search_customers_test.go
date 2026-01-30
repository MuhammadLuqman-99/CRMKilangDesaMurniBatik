// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/ports"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Mock Search Index with Full Implementation
// ============================================================================

// MockSearchIndexFull is a full mock implementation of ports.SearchIndex.
type MockSearchIndexFull struct {
	searchCustomersResult *ports.SearchResult
	searchCustomersErr    error
	searchContactsResult  *ports.SearchResult
	searchContactsErr     error
}

func NewMockSearchIndexFull() *MockSearchIndexFull {
	return &MockSearchIndexFull{
		searchCustomersResult: &ports.SearchResult{
			TotalHits: 0,
			Hits:      []ports.SearchHit{},
		},
		searchContactsResult: &ports.SearchResult{
			TotalHits: 0,
			Hits:      []ports.SearchHit{},
		},
	}
}

func (m *MockSearchIndexFull) IndexCustomer(ctx context.Context, customer *domain.Customer) error {
	return nil
}

func (m *MockSearchIndexFull) IndexContact(ctx context.Context, contact *domain.Contact) error {
	return nil
}

func (m *MockSearchIndexFull) RemoveCustomer(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockSearchIndexFull) RemoveContact(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockSearchIndexFull) SearchCustomers(ctx context.Context, tenantID uuid.UUID, query string, options ports.SearchOptions) (*ports.SearchResult, error) {
	if m.searchCustomersErr != nil {
		return nil, m.searchCustomersErr
	}
	return m.searchCustomersResult, nil
}

func (m *MockSearchIndexFull) SearchContacts(ctx context.Context, tenantID uuid.UUID, query string, options ports.SearchOptions) (*ports.SearchResult, error) {
	if m.searchContactsErr != nil {
		return nil, m.searchContactsErr
	}
	return m.searchContactsResult, nil
}

func (m *MockSearchIndexFull) Reindex(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}

// ============================================================================
// SearchCustomersUseCase Tests
// ============================================================================

func createTestCustomersForSearch(tenantID uuid.UUID, count int) []*domain.Customer {
	customers := make([]*domain.Customer, count)
	for i := 0; i < count; i++ {
		customer, _ := domain.NewCustomer(tenantID, "Customer "+string(rune('A'+i)), domain.CustomerTypeCompany)
		customer.Email, _ = domain.NewEmail("customer" + string(rune('a'+i)) + "@example.com")
		customer.Status = domain.CustomerStatusActive
		customers[i] = customer
	}
	return customers
}

func TestSearchCustomersUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

	tenantID := uuid.New()
	customers := createTestCustomersForSearch(tenantID, 3)
	for _, c := range customers {
		uow.customerRepo.customers[c.ID] = c
	}

	input := SearchCustomersInput{
		TenantID: tenantID,
		Request: &dto.SearchCustomersRequest{
			Query: "Customer",
			Limit: 10,
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

func TestSearchCustomersUseCase_Execute_WithoutQuery(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

	tenantID := uuid.New()
	customers := createTestCustomersForSearch(tenantID, 5)
	for _, c := range customers {
		uow.customerRepo.customers[c.ID] = c
	}

	input := SearchCustomersInput{
		TenantID: tenantID,
		Request: &dto.SearchCustomersRequest{
			Limit: 10,
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

func TestSearchCustomersUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

	input := SearchCustomersInput{
		TenantID: uuid.Nil, // Missing
		Request: &dto.SearchCustomersRequest{
			Query: "test",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestSearchCustomersUseCase_Execute_NilRequest(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

	tenantID := uuid.New()
	customers := createTestCustomersForSearch(tenantID, 2)
	for _, c := range customers {
		uow.customerRepo.customers[c.ID] = c
	}

	input := SearchCustomersInput{
		TenantID: tenantID,
		Request:  nil, // Nil request should use defaults
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

func TestSearchCustomersUseCase_Execute_LimitExceedsMax(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	config.MaxLimit = 50
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

	tenantID := uuid.New()

	input := SearchCustomersInput{
		TenantID: tenantID,
		Request: &dto.SearchCustomersRequest{
			Limit: 1000, // Exceeds max
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

func TestSearchCustomersUseCase_Execute_WithFullTextSearch(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Acme Corporation", domain.CustomerTypeCompany)
	uow.customerRepo.customers[customer.ID] = customer

	// Set up search index to return results
	searchIndex.searchCustomersResult = &ports.SearchResult{
		TotalHits: 1,
		Hits: []ports.SearchHit{
			{
				ID:    customer.ID,
				Score: 0.95,
			},
		},
	}

	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	config.EnableFullText = true
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

	input := SearchCustomersInput{
		TenantID: tenantID,
		Request: &dto.SearchCustomersRequest{
			Query: "Acme",
			Limit: 10,
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

func TestSearchCustomersUseCase_Execute_WithFilters(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

	tenantID := uuid.New()
	customer, _ := domain.NewCustomer(tenantID, "Active Customer", domain.CustomerTypeCompany)
	customer.Status = domain.CustomerStatusActive
	uow.customerRepo.customers[customer.ID] = customer

	input := SearchCustomersInput{
		TenantID: tenantID,
		Request: &dto.SearchCustomersRequest{
			Statuses: []domain.CustomerStatus{domain.CustomerStatusActive},
			Types:    []domain.CustomerType{domain.CustomerTypeCompany},
			Limit:    10,
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

// ============================================================================
// ListCustomersUseCase Tests
// ============================================================================

func TestListCustomersUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewListCustomersUseCase(uow, cache, config)

	tenantID := uuid.New()
	customers := createTestCustomersForSearch(tenantID, 5)
	for _, c := range customers {
		uow.customerRepo.customers[c.ID] = c
	}

	input := ListCustomersInput{
		TenantID: tenantID,
		Offset:   0,
		Limit:    10,
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

func TestListCustomersUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewListCustomersUseCase(uow, cache, config)

	input := ListCustomersInput{
		TenantID: uuid.Nil, // Missing
		Offset:   0,
		Limit:    10,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestListCustomersUseCase_Execute_DefaultPagination(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	config.DefaultLimit = 20
	uc := NewListCustomersUseCase(uow, cache, config)

	tenantID := uuid.New()

	input := ListCustomersInput{
		TenantID: tenantID,
		Offset:   0,
		Limit:    0, // Should use default
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

func TestListCustomersUseCase_Execute_LimitExceedsMax(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	config.MaxLimit = 100
	uc := NewListCustomersUseCase(uow, cache, config)

	tenantID := uuid.New()

	input := ListCustomersInput{
		TenantID: tenantID,
		Offset:   0,
		Limit:    500, // Exceeds max
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

func TestListCustomersUseCase_Execute_WithSorting(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewListCustomersUseCase(uow, cache, config)

	tenantID := uuid.New()

	input := ListCustomersInput{
		TenantID:  tenantID,
		Offset:    0,
		Limit:     10,
		SortBy:    "name",
		SortOrder: "asc",
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
// ListCustomersByOwnerUseCase Tests
// ============================================================================

func TestListCustomersByOwnerUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByOwnerUseCase(uow)

	tenantID := uuid.New()
	ownerID := uuid.New()

	input := ListByOwnerInput{
		TenantID: tenantID,
		OwnerID:  ownerID,
		Offset:   0,
		Limit:    10,
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

func TestListCustomersByOwnerUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByOwnerUseCase(uow)

	input := ListByOwnerInput{
		TenantID: uuid.Nil, // Missing
		OwnerID:  uuid.New(),
		Offset:   0,
		Limit:    10,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestListCustomersByOwnerUseCase_Execute_MissingOwnerID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByOwnerUseCase(uow)

	input := ListByOwnerInput{
		TenantID: uuid.New(),
		OwnerID:  uuid.Nil, // Missing
		Offset:   0,
		Limit:    10,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing owner ID, got nil")
	}
}

func TestListCustomersByOwnerUseCase_Execute_DefaultLimit(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByOwnerUseCase(uow)

	tenantID := uuid.New()
	ownerID := uuid.New()

	input := ListByOwnerInput{
		TenantID: tenantID,
		OwnerID:  ownerID,
		Offset:   0,
		Limit:    0, // Should use default
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
// ListCustomersByStatusUseCase Tests
// ============================================================================

func TestListCustomersByStatusUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByStatusUseCase(uow)

	tenantID := uuid.New()

	input := ListByStatusInput{
		TenantID: tenantID,
		Status:   domain.CustomerStatusActive,
		Offset:   0,
		Limit:    10,
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

func TestListCustomersByStatusUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByStatusUseCase(uow)

	input := ListByStatusInput{
		TenantID: uuid.Nil, // Missing
		Status:   domain.CustomerStatusActive,
		Offset:   0,
		Limit:    10,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestListCustomersByStatusUseCase_Execute_DefaultLimit(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByStatusUseCase(uow)

	tenantID := uuid.New()

	input := ListByStatusInput{
		TenantID: tenantID,
		Status:   domain.CustomerStatusActive,
		Offset:   0,
		Limit:    0, // Should use default
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
// ListCustomersByTagUseCase Tests
// ============================================================================

func TestListCustomersByTagUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByTagUseCase(uow)

	tenantID := uuid.New()

	input := ListByTagInput{
		TenantID: tenantID,
		Tag:      "vip",
		Offset:   0,
		Limit:    10,
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

func TestListCustomersByTagUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByTagUseCase(uow)

	input := ListByTagInput{
		TenantID: uuid.Nil, // Missing
		Tag:      "vip",
		Offset:   0,
		Limit:    10,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestListCustomersByTagUseCase_Execute_MissingTag(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByTagUseCase(uow)

	input := ListByTagInput{
		TenantID: uuid.New(),
		Tag:      "", // Missing
		Offset:   0,
		Limit:    10,
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tag, got nil")
	}
}

func TestListCustomersByTagUseCase_Execute_DefaultLimit(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	uc := NewListCustomersByTagUseCase(uow)

	tenantID := uuid.New()

	input := ListByTagInput{
		TenantID: tenantID,
		Tag:      "vip",
		Offset:   0,
		Limit:    0, // Should use default
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
// SearchContactsUseCase Tests
// ============================================================================

func TestSearchContactsUseCase_Execute_Success(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()

	config := DefaultSearchConfig()
	uc := NewSearchContactsUseCase(uow, searchIndex, config)

	tenantID := uuid.New()

	input := SearchContactsInput{
		TenantID: tenantID,
		Request: &dto.SearchContactsRequest{
			Query: "John",
			Limit: 10,
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

func TestSearchContactsUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()

	config := DefaultSearchConfig()
	uc := NewSearchContactsUseCase(uow, searchIndex, config)

	input := SearchContactsInput{
		TenantID: uuid.Nil, // Missing
		Request: &dto.SearchContactsRequest{
			Query: "John",
		},
	}

	// Act
	_, err := uc.Execute(context.Background(), input)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing tenant ID, got nil")
	}
}

func TestSearchContactsUseCase_Execute_NilRequest(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()

	config := DefaultSearchConfig()
	uc := NewSearchContactsUseCase(uow, searchIndex, config)

	tenantID := uuid.New()

	input := SearchContactsInput{
		TenantID: tenantID,
		Request:  nil, // Nil request
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

func TestSearchContactsUseCase_Execute_LimitExceedsMax(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()

	config := DefaultSearchConfig()
	config.MaxLimit = 50
	uc := NewSearchContactsUseCase(uow, searchIndex, config)

	tenantID := uuid.New()

	input := SearchContactsInput{
		TenantID: tenantID,
		Request: &dto.SearchContactsRequest{
			Query: "John",
			Limit: 1000, // Exceeds max
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

// ============================================================================
// Table-Driven Tests
// ============================================================================

func TestSearchCustomersUseCase_Execute_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		input     SearchCustomersInput
		expectErr bool
	}{
		{
			name: "valid search with query",
			input: SearchCustomersInput{
				TenantID: uuid.New(),
				Request: &dto.SearchCustomersRequest{
					Query: "test",
					Limit: 10,
				},
			},
			expectErr: false,
		},
		{
			name: "valid search without query",
			input: SearchCustomersInput{
				TenantID: uuid.New(),
				Request: &dto.SearchCustomersRequest{
					Limit: 10,
				},
			},
			expectErr: false,
		},
		{
			name: "valid search with nil request",
			input: SearchCustomersInput{
				TenantID: uuid.New(),
				Request:  nil,
			},
			expectErr: false,
		},
		{
			name: "missing tenant ID",
			input: SearchCustomersInput{
				TenantID: uuid.Nil,
				Request: &dto.SearchCustomersRequest{
					Query: "test",
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := NewMockUnitOfWork()
			searchIndex := NewMockSearchIndexFull()
			cache := NewMockCustomerCacheService()

			config := DefaultSearchConfig()
			uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

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

func TestListCustomersUseCase_Execute_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		input     ListCustomersInput
		expectErr bool
	}{
		{
			name: "valid list",
			input: ListCustomersInput{
				TenantID: uuid.New(),
				Offset:   0,
				Limit:    10,
			},
			expectErr: false,
		},
		{
			name: "valid list with sorting",
			input: ListCustomersInput{
				TenantID:  uuid.New(),
				Offset:    0,
				Limit:     10,
				SortBy:    "name",
				SortOrder: "asc",
			},
			expectErr: false,
		},
		{
			name: "missing tenant ID",
			input: ListCustomersInput{
				TenantID: uuid.Nil,
				Offset:   0,
				Limit:    10,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := NewMockUnitOfWork()
			cache := NewMockCustomerCacheService()

			config := DefaultSearchConfig()
			uc := NewListCustomersUseCase(uow, cache, config)

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

func BenchmarkSearchCustomersUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()
	cache := NewMockCustomerCacheService()

	tenantID := uuid.New()
	for i := 0; i < 100; i++ {
		customer, _ := domain.NewCustomer(tenantID, "Customer", domain.CustomerTypeCompany)
		uow.customerRepo.customers[customer.ID] = customer
	}

	config := DefaultSearchConfig()
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := SearchCustomersInput{
			TenantID: tenantID,
			Request: &dto.SearchCustomersRequest{
				Query: "Customer",
				Limit: 10,
			},
		}
		_, _ = uc.Execute(ctx, input)
	}
}

func BenchmarkListCustomersUseCase_Execute(b *testing.B) {
	uow := NewMockUnitOfWork()
	cache := NewMockCustomerCacheService()

	tenantID := uuid.New()
	for i := 0; i < 100; i++ {
		customer, _ := domain.NewCustomer(tenantID, "Customer", domain.CustomerTypeCompany)
		uow.customerRepo.customers[customer.ID] = customer
	}

	config := DefaultSearchConfig()
	uc := NewListCustomersUseCase(uow, cache, config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := ListCustomersInput{
			TenantID: tenantID,
			Offset:   0,
			Limit:    10,
		}
		_, _ = uc.Execute(ctx, input)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestSearchCustomersUseCase_Execute_ContextTimeout(t *testing.T) {
	// Arrange
	uow := NewMockUnitOfWork()
	searchIndex := NewMockSearchIndexFull()
	cache := NewMockCustomerCacheService()

	config := DefaultSearchConfig()
	uc := NewSearchCustomersUseCase(uow, searchIndex, cache, config)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	input := SearchCustomersInput{
		TenantID: uuid.New(),
		Request: &dto.SearchCustomersRequest{
			Query: "test",
		},
	}

	// Act - just ensure no panic
	_, _ = uc.Execute(ctx, input)
}
