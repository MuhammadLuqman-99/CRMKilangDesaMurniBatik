// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/dto"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/mapper"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/ports"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
)

// ============================================================================
// Search Customers Use Case
// ============================================================================

// SearchCustomersUseCase handles customer searching.
type SearchCustomersUseCase struct {
	uow            domain.UnitOfWork
	searchIndex    ports.SearchIndex
	cache          ports.CacheService
	customerMapper *mapper.CustomerMapper
	config         SearchConfig
}

// SearchConfig holds configuration for search operations.
type SearchConfig struct {
	DefaultLimit     int
	MaxLimit         int
	CacheTTLSeconds  int
	EnableFullText   bool
	FuzzyMatchEnabled bool
}

// DefaultSearchConfig returns default configuration.
func DefaultSearchConfig() SearchConfig {
	return SearchConfig{
		DefaultLimit:     20,
		MaxLimit:         100,
		CacheTTLSeconds:  60,
		EnableFullText:   true,
		FuzzyMatchEnabled: true,
	}
}

// NewSearchCustomersUseCase creates a new SearchCustomersUseCase.
func NewSearchCustomersUseCase(
	uow domain.UnitOfWork,
	searchIndex ports.SearchIndex,
	cache ports.CacheService,
	config SearchConfig,
) *SearchCustomersUseCase {
	return &SearchCustomersUseCase{
		uow:            uow,
		searchIndex:    searchIndex,
		cache:          cache,
		customerMapper: mapper.NewCustomerMapper(),
		config:         config,
	}
}

// SearchCustomersInput holds input for customer search.
type SearchCustomersInput struct {
	TenantID uuid.UUID
	Request  *dto.SearchCustomersRequest
}

// Execute searches for customers.
func (uc *SearchCustomersUseCase) Execute(ctx context.Context, input SearchCustomersInput) (*dto.CustomerListResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id is required")
	}
	if input.Request == nil {
		input.Request = &dto.SearchCustomersRequest{}
	}

	// Normalize pagination
	offset, limit, sortBy, sortOrder := uc.customerMapper.PaginationFromRequest(input.Request)
	if limit > uc.config.MaxLimit {
		limit = uc.config.MaxLimit
	}

	// Build filter
	filter := uc.customerMapper.SearchRequestToFilter(input.Request)
	filter.TenantID = &input.TenantID
	filter.Offset = offset
	filter.Limit = limit
	filter.SortBy = sortBy
	filter.SortOrder = sortOrder

	// Try full-text search if query is provided and search index is available
	if input.Request.Query != "" && uc.searchIndex != nil && uc.config.EnableFullText {
		return uc.fullTextSearch(ctx, input.TenantID, input.Request.Query, filter)
	}

	// Use repository search
	return uc.repositorySearch(ctx, input.TenantID, input.Request.Query, filter)
}

// fullTextSearch performs full-text search using the search index.
func (uc *SearchCustomersUseCase) fullTextSearch(ctx context.Context, tenantID uuid.UUID, query string, filter domain.CustomerFilter) (*dto.CustomerListResponse, error) {
	searchOpts := ports.SearchOptions{
		Offset:        filter.Offset,
		Limit:         filter.Limit,
		SortBy:        filter.SortBy,
		SortOrder:     filter.SortOrder,
		FuzzyMatching: uc.config.FuzzyMatchEnabled,
	}

	// Add filters
	searchOpts.Filters = make(map[string]interface{})
	if len(filter.Types) > 0 {
		searchOpts.Filters["type"] = filter.Types
	}
	if len(filter.Statuses) > 0 {
		searchOpts.Filters["status"] = filter.Statuses
	}
	if len(filter.Tiers) > 0 {
		searchOpts.Filters["tier"] = filter.Tiers
	}
	if len(filter.Tags) > 0 {
		searchOpts.Filters["tags"] = filter.Tags
	}
	if len(filter.OwnerIDs) > 0 {
		searchOpts.Filters["owner_id"] = filter.OwnerIDs
	}

	result, err := uc.searchIndex.SearchCustomers(ctx, tenantID, query, searchOpts)
	if err != nil {
		// Fallback to repository search
		return uc.repositorySearch(ctx, tenantID, query, filter)
	}

	// Fetch full customer data for each hit
	customerIDs := make([]uuid.UUID, len(result.Hits))
	for i, hit := range result.Hits {
		customerIDs[i] = hit.ID
	}

	// Fetch customers from repository
	filter.IDs = customerIDs
	filter.Query = "" // Already searched
	customerList, err := uc.uow.Customers().List(ctx, filter)
	if err != nil {
		return nil, application.ErrInternalError("failed to fetch customers", err)
	}

	// Sort by search result order
	customerMap := make(map[uuid.UUID]*domain.Customer)
	for _, customer := range customerList.Customers {
		customerMap[customer.ID] = customer
	}

	orderedCustomers := make([]*domain.Customer, 0, len(customerIDs))
	for _, id := range customerIDs {
		if customer, ok := customerMap[id]; ok {
			orderedCustomers = append(orderedCustomers, customer)
		}
	}

	return uc.customerMapper.ToListResponse(orderedCustomers, result.TotalHits, filter.Offset, filter.Limit), nil
}

// repositorySearch performs search using the repository.
func (uc *SearchCustomersUseCase) repositorySearch(ctx context.Context, tenantID uuid.UUID, query string, filter domain.CustomerFilter) (*dto.CustomerListResponse, error) {
	var customerList *domain.CustomerList
	var err error

	if query != "" {
		customerList, err = uc.uow.Customers().Search(ctx, tenantID, query, filter)
	} else {
		customerList, err = uc.uow.Customers().List(ctx, filter)
	}

	if err != nil {
		return nil, application.ErrInternalError("failed to search customers", err)
	}

	return uc.customerMapper.ToListResponse(customerList.Customers, customerList.Total, customerList.Offset, customerList.Limit), nil
}

// ============================================================================
// List Customers Use Case
// ============================================================================

// ListCustomersUseCase handles listing customers.
type ListCustomersUseCase struct {
	uow            domain.UnitOfWork
	cache          ports.CacheService
	customerMapper *mapper.CustomerMapper
	config         SearchConfig
}

// NewListCustomersUseCase creates a new ListCustomersUseCase.
func NewListCustomersUseCase(
	uow domain.UnitOfWork,
	cache ports.CacheService,
	config SearchConfig,
) *ListCustomersUseCase {
	return &ListCustomersUseCase{
		uow:            uow,
		cache:          cache,
		customerMapper: mapper.NewCustomerMapper(),
		config:         config,
	}
}

// ListCustomersInput holds input for listing customers.
type ListCustomersInput struct {
	TenantID  uuid.UUID
	Offset    int
	Limit     int
	SortBy    string
	SortOrder string
}

// Execute lists customers.
func (uc *ListCustomersUseCase) Execute(ctx context.Context, input ListCustomersInput) (*dto.CustomerListResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id is required")
	}

	// Normalize pagination
	if input.Limit == 0 {
		input.Limit = uc.config.DefaultLimit
	}
	if input.Limit > uc.config.MaxLimit {
		input.Limit = uc.config.MaxLimit
	}
	if input.SortBy == "" {
		input.SortBy = "created_at"
	}
	if input.SortOrder == "" {
		input.SortOrder = "desc"
	}

	// Build filter
	filter := domain.CustomerFilter{
		TenantID:  &input.TenantID,
		Offset:    input.Offset,
		Limit:     input.Limit,
		SortBy:    input.SortBy,
		SortOrder: input.SortOrder,
	}

	// List customers
	customerList, err := uc.uow.Customers().List(ctx, filter)
	if err != nil {
		return nil, application.ErrInternalError("failed to list customers", err)
	}

	return uc.customerMapper.ToListResponse(customerList.Customers, customerList.Total, customerList.Offset, customerList.Limit), nil
}

// ============================================================================
// List Customers By Owner Use Case
// ============================================================================

// ListCustomersByOwnerUseCase handles listing customers by owner.
type ListCustomersByOwnerUseCase struct {
	uow            domain.UnitOfWork
	customerMapper *mapper.CustomerMapper
}

// NewListCustomersByOwnerUseCase creates a new ListCustomersByOwnerUseCase.
func NewListCustomersByOwnerUseCase(uow domain.UnitOfWork) *ListCustomersByOwnerUseCase {
	return &ListCustomersByOwnerUseCase{
		uow:            uow,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// ListByOwnerInput holds input for listing by owner.
type ListByOwnerInput struct {
	TenantID uuid.UUID
	OwnerID  uuid.UUID
	Offset   int
	Limit    int
}

// Execute lists customers by owner.
func (uc *ListCustomersByOwnerUseCase) Execute(ctx context.Context, input ListByOwnerInput) (*dto.CustomerListResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil || input.OwnerID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id and owner_id are required")
	}

	if input.Limit == 0 {
		input.Limit = 20
	}

	filter := domain.CustomerFilter{
		Offset: input.Offset,
		Limit:  input.Limit,
	}

	customerList, err := uc.uow.Customers().FindByOwner(ctx, input.TenantID, input.OwnerID, filter)
	if err != nil {
		return nil, application.ErrInternalError("failed to list customers by owner", err)
	}

	return uc.customerMapper.ToListResponse(customerList.Customers, customerList.Total, customerList.Offset, customerList.Limit), nil
}

// ============================================================================
// List Customers By Status Use Case
// ============================================================================

// ListCustomersByStatusUseCase handles listing customers by status.
type ListCustomersByStatusUseCase struct {
	uow            domain.UnitOfWork
	customerMapper *mapper.CustomerMapper
}

// NewListCustomersByStatusUseCase creates a new ListCustomersByStatusUseCase.
func NewListCustomersByStatusUseCase(uow domain.UnitOfWork) *ListCustomersByStatusUseCase {
	return &ListCustomersByStatusUseCase{
		uow:            uow,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// ListByStatusInput holds input for listing by status.
type ListByStatusInput struct {
	TenantID uuid.UUID
	Status   domain.CustomerStatus
	Offset   int
	Limit    int
}

// Execute lists customers by status.
func (uc *ListCustomersByStatusUseCase) Execute(ctx context.Context, input ListByStatusInput) (*dto.CustomerListResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id is required")
	}

	if input.Limit == 0 {
		input.Limit = 20
	}

	filter := domain.CustomerFilter{
		Offset: input.Offset,
		Limit:  input.Limit,
	}

	customerList, err := uc.uow.Customers().FindByStatus(ctx, input.TenantID, input.Status, filter)
	if err != nil {
		return nil, application.ErrInternalError("failed to list customers by status", err)
	}

	return uc.customerMapper.ToListResponse(customerList.Customers, customerList.Total, customerList.Offset, customerList.Limit), nil
}

// ============================================================================
// List Customers By Tag Use Case
// ============================================================================

// ListCustomersByTagUseCase handles listing customers by tag.
type ListCustomersByTagUseCase struct {
	uow            domain.UnitOfWork
	customerMapper *mapper.CustomerMapper
}

// NewListCustomersByTagUseCase creates a new ListCustomersByTagUseCase.
func NewListCustomersByTagUseCase(uow domain.UnitOfWork) *ListCustomersByTagUseCase {
	return &ListCustomersByTagUseCase{
		uow:            uow,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// ListByTagInput holds input for listing by tag.
type ListByTagInput struct {
	TenantID uuid.UUID
	Tag      string
	Offset   int
	Limit    int
}

// Execute lists customers by tag.
func (uc *ListCustomersByTagUseCase) Execute(ctx context.Context, input ListByTagInput) (*dto.CustomerListResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id is required")
	}
	if input.Tag == "" {
		return nil, application.ErrInvalidInput("tag is required")
	}

	if input.Limit == 0 {
		input.Limit = 20
	}

	filter := domain.CustomerFilter{
		Offset: input.Offset,
		Limit:  input.Limit,
	}

	customerList, err := uc.uow.Customers().FindByTag(ctx, input.TenantID, input.Tag, filter)
	if err != nil {
		return nil, application.ErrInternalError("failed to list customers by tag", err)
	}

	return uc.customerMapper.ToListResponse(customerList.Customers, customerList.Total, customerList.Offset, customerList.Limit), nil
}

// ============================================================================
// Search Contacts Use Case
// ============================================================================

// SearchContactsUseCase handles contact searching.
type SearchContactsUseCase struct {
	uow           domain.UnitOfWork
	searchIndex   ports.SearchIndex
	contactMapper *mapper.ContactMapper
	config        SearchConfig
}

// NewSearchContactsUseCase creates a new SearchContactsUseCase.
func NewSearchContactsUseCase(
	uow domain.UnitOfWork,
	searchIndex ports.SearchIndex,
	config SearchConfig,
) *SearchContactsUseCase {
	return &SearchContactsUseCase{
		uow:           uow,
		searchIndex:   searchIndex,
		contactMapper: mapper.NewContactMapper(),
		config:        config,
	}
}

// SearchContactsInput holds input for contact search.
type SearchContactsInput struct {
	TenantID uuid.UUID
	Request  *dto.SearchContactsRequest
}

// Execute searches for contacts.
func (uc *SearchContactsUseCase) Execute(ctx context.Context, input SearchContactsInput) (*dto.ContactListResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id is required")
	}
	if input.Request == nil {
		input.Request = &dto.SearchContactsRequest{}
	}

	// Normalize pagination
	offset, limit, sortBy, sortOrder := uc.contactMapper.PaginationFromSearchRequest(input.Request)
	if limit > uc.config.MaxLimit {
		limit = uc.config.MaxLimit
	}

	// Build filter
	filter := uc.contactMapper.SearchRequestToFilter(input.Request)
	filter.TenantID = &input.TenantID
	filter.Offset = offset
	filter.Limit = limit
	filter.SortBy = sortBy
	filter.SortOrder = sortOrder

	// Search contacts
	contactList, err := uc.uow.Contacts().Search(ctx, input.TenantID, input.Request.Query, filter)
	if err != nil {
		return nil, application.ErrInternalError("failed to search contacts", err)
	}

	return uc.contactMapper.ToListResponse(contactList.Contacts, contactList.Total, contactList.Offset, contactList.Limit), nil
}
