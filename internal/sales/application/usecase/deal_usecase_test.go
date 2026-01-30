package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
	"github.com/kilang-desa-murni/crm/internal/sales/application/ports"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Mock Implementations for Deal Tests
// ============================================================================

// DealMockDealRepository is a mock implementation of domain.DealRepository.
type DealMockDealRepository struct {
	deals      map[uuid.UUID]*domain.Deal
	createErr  error
	updateErr  error
	deleteErr  error
	getByIDErr error
}

func NewDealMockDealRepository() *DealMockDealRepository {
	return &DealMockDealRepository{
		deals: make(map[uuid.UUID]*domain.Deal),
	}
}

func (m *DealMockDealRepository) Create(ctx context.Context, deal *domain.Deal) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.deals[deal.ID] = deal
	return nil
}

func (m *DealMockDealRepository) Update(ctx context.Context, deal *domain.Deal) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.deals[deal.ID] = deal
	return nil
}

func (m *DealMockDealRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.deals, id)
	return nil
}

func (m *DealMockDealRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Deal, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	deal, ok := m.deals[id]
	if !ok {
		return nil, errors.New("deal not found")
	}
	return deal, nil
}

func (m *DealMockDealRepository) GetByNumber(ctx context.Context, tenantID uuid.UUID, number string) (*domain.Deal, error) {
	for _, deal := range m.deals {
		if deal.Code == number && deal.TenantID == tenantID {
			return deal, nil
		}
	}
	return nil, errors.New("deal not found")
}

func (m *DealMockDealRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.DealFilter, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	var result []*domain.Deal
	for _, deal := range m.deals {
		if deal.TenantID == tenantID {
			result = append(result, deal)
		}
	}
	return result, int64(len(result)), nil
}

func (m *DealMockDealRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.DealStatus, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	var result []*domain.Deal
	for _, deal := range m.deals {
		if deal.TenantID == tenantID && deal.Status == status {
			result = append(result, deal)
		}
	}
	return result, int64(len(result)), nil
}

func (m *DealMockDealRepository) GetActiveDeals(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return m.GetByStatus(ctx, tenantID, domain.DealStatusActive, opts)
}

func (m *DealMockDealRepository) GetCompletedDeals(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return m.GetByStatus(ctx, tenantID, domain.DealStatusFulfilled, opts)
}

func (m *DealMockDealRepository) GetByOpportunity(ctx context.Context, tenantID, opportunityID uuid.UUID) (*domain.Deal, error) {
	for _, deal := range m.deals {
		if deal.TenantID == tenantID && deal.OpportunityID == opportunityID {
			return deal, nil
		}
	}
	return nil, errors.New("deal not found")
}

func (m *DealMockDealRepository) GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	var result []*domain.Deal
	for _, deal := range m.deals {
		if deal.TenantID == tenantID && deal.CustomerID == customerID {
			result = append(result, deal)
		}
	}
	return result, int64(len(result)), nil
}

func (m *DealMockDealRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	var result []*domain.Deal
	for _, deal := range m.deals {
		if deal.TenantID == tenantID && deal.OwnerID == ownerID {
			result = append(result, deal)
		}
	}
	return result, int64(len(result)), nil
}

func (m *DealMockDealRepository) GetDealsWithPendingPayments(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return []*domain.Deal{}, 0, nil
}

func (m *DealMockDealRepository) GetOverduePayments(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return []*domain.Deal{}, 0, nil
}

func (m *DealMockDealRepository) GetFullyPaidDeals(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return []*domain.Deal{}, 0, nil
}

func (m *DealMockDealRepository) GetDealsForFulfillment(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return []*domain.Deal{}, 0, nil
}

func (m *DealMockDealRepository) GetPartiallyFulfilledDeals(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return []*domain.Deal{}, 0, nil
}

func (m *DealMockDealRepository) GetByClosedDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return []*domain.Deal{}, 0, nil
}

func (m *DealMockDealRepository) GetBySignedDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Deal, int64, error) {
	return []*domain.Deal{}, 0, nil
}

func (m *DealMockDealRepository) GetTotalRevenue(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	return 0, nil
}

func (m *DealMockDealRepository) GetTotalReceivedPayments(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	return 0, nil
}

func (m *DealMockDealRepository) GetOutstandingAmount(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	return 0, nil
}

func (m *DealMockDealRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.DealStatus]int64, error) {
	counts := make(map[domain.DealStatus]int64)
	for _, deal := range m.deals {
		if deal.TenantID == tenantID {
			counts[deal.Status]++
		}
	}
	return counts, nil
}

func (m *DealMockDealRepository) GetAverageDealValue(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	return 0, nil
}

func (m *DealMockDealRepository) GetMonthlyRevenue(ctx context.Context, tenantID uuid.UUID, currency string, year int) (map[int]int64, error) {
	return map[int]int64{}, nil
}

// DealMockOpportunityRepository is a mock implementation of domain.OpportunityRepository.
type DealMockOpportunityRepository struct {
	opportunities map[uuid.UUID]*domain.Opportunity
	getByIDErr    error
}

func NewDealMockOpportunityRepository() *DealMockOpportunityRepository {
	return &DealMockOpportunityRepository{
		opportunities: make(map[uuid.UUID]*domain.Opportunity),
	}
}

func (m *DealMockOpportunityRepository) Create(ctx context.Context, opp *domain.Opportunity) error {
	m.opportunities[opp.ID] = opp
	return nil
}

func (m *DealMockOpportunityRepository) Update(ctx context.Context, opp *domain.Opportunity) error {
	m.opportunities[opp.ID] = opp
	return nil
}

func (m *DealMockOpportunityRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	delete(m.opportunities, id)
	return nil
}

func (m *DealMockOpportunityRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Opportunity, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	opp, ok := m.opportunities[id]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	return opp, nil
}

func (m *DealMockOpportunityRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.OpportunityFilter, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.OpportunityStatus, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetOpenOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetWonOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetLostOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetByPipeline(ctx context.Context, tenantID, pipelineID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetByStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetByContact(ctx context.Context, tenantID, contactID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetByLead(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.Opportunity, error) {
	return nil, errors.New("not found")
}

func (m *DealMockOpportunityRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetClosingThisMonth(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetClosingThisQuarter(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetOverdueOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetByExpectedCloseDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetHighValueOpportunities(ctx context.Context, tenantID uuid.UUID, minAmount int64, currency string, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *DealMockOpportunityRepository) GetTotalPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	return 0, nil
}

func (m *DealMockOpportunityRepository) GetWeightedPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	return 0, nil
}

func (m *DealMockOpportunityRepository) BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, ownerID uuid.UUID) error {
	return nil
}

func (m *DealMockOpportunityRepository) BulkUpdateStage(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, stageID uuid.UUID) error {
	return nil
}

func (m *DealMockOpportunityRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.OpportunityStatus]int64, error) {
	return map[domain.OpportunityStatus]int64{}, nil
}

func (m *DealMockOpportunityRepository) CountByStage(ctx context.Context, tenantID, pipelineID uuid.UUID) (map[uuid.UUID]int64, error) {
	return map[uuid.UUID]int64{}, nil
}

func (m *DealMockOpportunityRepository) GetWinRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	return 0, nil
}

func (m *DealMockOpportunityRepository) GetAverageDealSize(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	return 0, nil
}

func (m *DealMockOpportunityRepository) GetAverageSalesCycle(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (int, error) {
	return 0, nil
}

// DealMockEventPublisher is a mock implementation of ports.EventPublisher.
type DealMockEventPublisher struct {
	events []ports.Event
}

func NewDealMockEventPublisher() *DealMockEventPublisher {
	return &DealMockEventPublisher{
		events: make([]ports.Event, 0),
	}
}

func (m *DealMockEventPublisher) Publish(ctx context.Context, event ports.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *DealMockEventPublisher) PublishBatch(ctx context.Context, events []ports.Event) error {
	m.events = append(m.events, events...)
	return nil
}

func (m *DealMockEventPublisher) PublishAsync(ctx context.Context, event ports.Event) error {
	m.events = append(m.events, event)
	return nil
}

// DealMockCustomerService is a mock implementation of ports.CustomerService.
type DealMockCustomerService struct {
	customers map[uuid.UUID]*ports.CustomerInfo
	contacts  map[uuid.UUID]*ports.ContactInfo
}

func NewDealMockCustomerService() *DealMockCustomerService {
	return &DealMockCustomerService{
		customers: make(map[uuid.UUID]*ports.CustomerInfo),
		contacts:  make(map[uuid.UUID]*ports.ContactInfo),
	}
}

func (m *DealMockCustomerService) GetCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*ports.CustomerInfo, error) {
	customer, ok := m.customers[customerID]
	if !ok {
		return nil, errors.New("customer not found")
	}
	return customer, nil
}

func (m *DealMockCustomerService) GetCustomerByCode(ctx context.Context, tenantID uuid.UUID, code string) (*ports.CustomerInfo, error) {
	for _, customer := range m.customers {
		if customer.Code == code {
			return customer, nil
		}
	}
	return nil, errors.New("customer not found")
}

func (m *DealMockCustomerService) CustomerExists(ctx context.Context, tenantID, customerID uuid.UUID) (bool, error) {
	_, ok := m.customers[customerID]
	return ok, nil
}

func (m *DealMockCustomerService) CreateCustomer(ctx context.Context, tenantID uuid.UUID, req ports.CreateCustomerRequest) (*ports.CustomerInfo, error) {
	customer := &ports.CustomerInfo{
		ID:   uuid.New(),
		Name: req.Name,
	}
	m.customers[customer.ID] = customer
	return customer, nil
}

func (m *DealMockCustomerService) GetContact(ctx context.Context, tenantID, contactID uuid.UUID) (*ports.ContactInfo, error) {
	contact, ok := m.contacts[contactID]
	if !ok {
		return nil, errors.New("contact not found")
	}
	return contact, nil
}

func (m *DealMockCustomerService) ContactExists(ctx context.Context, tenantID, contactID uuid.UUID) (bool, error) {
	_, ok := m.contacts[contactID]
	return ok, nil
}

func (m *DealMockCustomerService) CreateContact(ctx context.Context, tenantID, customerID uuid.UUID, req ports.CreateContactRequest) (*ports.ContactInfo, error) {
	contact := &ports.ContactInfo{
		ID:         uuid.New(),
		CustomerID: customerID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
	}
	m.contacts[contact.ID] = contact
	return contact, nil
}

// DealMockUserService is a mock implementation of ports.UserService.
type DealMockUserService struct {
	users map[uuid.UUID]*ports.UserInfo
}

func NewDealMockUserService() *DealMockUserService {
	return &DealMockUserService{
		users: make(map[uuid.UUID]*ports.UserInfo),
	}
}

func (m *DealMockUserService) GetUser(ctx context.Context, tenantID, userID uuid.UUID) (*ports.UserInfo, error) {
	user, ok := m.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *DealMockUserService) UserExists(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	_, ok := m.users[userID]
	return ok, nil
}

func (m *DealMockUserService) GetUsersByIDs(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID) ([]*ports.UserInfo, error) {
	var result []*ports.UserInfo
	for _, id := range userIDs {
		if user, ok := m.users[id]; ok {
			result = append(result, user)
		}
	}
	return result, nil
}

func (m *DealMockUserService) GetUsersByRole(ctx context.Context, tenantID uuid.UUID, roleName string) ([]*ports.UserInfo, error) {
	return []*ports.UserInfo{}, nil
}

func (m *DealMockUserService) HasPermission(ctx context.Context, tenantID, userID uuid.UUID, permission string) (bool, error) {
	return true, nil
}

// DealMockProductService is a mock implementation of ports.ProductService.
type DealMockProductService struct {
	products map[uuid.UUID]*ports.ProductInfo
}

func NewDealMockProductService() *DealMockProductService {
	return &DealMockProductService{
		products: make(map[uuid.UUID]*ports.ProductInfo),
	}
}

func (m *DealMockProductService) GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (*ports.ProductInfo, error) {
	product, ok := m.products[productID]
	if !ok {
		return nil, errors.New("product not found")
	}
	return product, nil
}

func (m *DealMockProductService) ProductExists(ctx context.Context, tenantID, productID uuid.UUID) (bool, error) {
	_, ok := m.products[productID]
	return ok, nil
}

func (m *DealMockProductService) GetProductsByIDs(ctx context.Context, tenantID uuid.UUID, productIDs []uuid.UUID) ([]*ports.ProductInfo, error) {
	var result []*ports.ProductInfo
	for _, id := range productIDs {
		if product, ok := m.products[id]; ok {
			result = append(result, product)
		}
	}
	return result, nil
}

func (m *DealMockProductService) SearchProducts(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]*ports.ProductInfo, error) {
	return []*ports.ProductInfo{}, nil
}

// DealMockCacheService is a mock implementation of ports.CacheService.
type DealMockCacheService struct {
	cache map[string][]byte
}

func NewDealMockCacheService() *DealMockCacheService {
	return &DealMockCacheService{
		cache: make(map[string][]byte),
	}
}

func (m *DealMockCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	v, ok := m.cache[key]
	if !ok {
		return nil, errors.New("cache miss")
	}
	return v, nil
}

func (m *DealMockCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.cache[key] = value
	return nil
}

func (m *DealMockCacheService) Delete(ctx context.Context, key string) error {
	delete(m.cache, key)
	return nil
}

func (m *DealMockCacheService) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *DealMockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.cache[key]
	return ok, nil
}

func (m *DealMockCacheService) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	for _, key := range keys {
		if v, ok := m.cache[key]; ok {
			result[key] = v
		}
	}
	return result, nil
}

func (m *DealMockCacheService) SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	for k, v := range items {
		m.cache[k] = v
	}
	return nil
}

func (m *DealMockCacheService) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return delta, nil
}

func (m *DealMockCacheService) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	if _, ok := m.cache[key]; ok {
		return false, nil
	}
	m.cache[key] = value
	return true, nil
}

// DealMockSearchService is a mock implementation of ports.SearchService.
type DealMockSearchService struct{}

func NewDealMockSearchService() *DealMockSearchService {
	return &DealMockSearchService{}
}

func (m *DealMockSearchService) IndexLead(ctx context.Context, lead ports.SearchableLead) error {
	return nil
}

func (m *DealMockSearchService) IndexOpportunity(ctx context.Context, opp ports.SearchableOpportunity) error {
	return nil
}

func (m *DealMockSearchService) IndexDeal(ctx context.Context, deal ports.SearchableDeal) error {
	return nil
}

func (m *DealMockSearchService) Search(ctx context.Context, tenantID uuid.UUID, query ports.SearchQuery) (*ports.SearchResult, error) {
	return &ports.SearchResult{}, nil
}

func (m *DealMockSearchService) DeleteIndex(ctx context.Context, tenantID uuid.UUID, entityType string, entityID uuid.UUID) error {
	return nil
}

func (m *DealMockSearchService) BulkIndex(ctx context.Context, entities []ports.SearchableEntity) error {
	return nil
}

// DealMockIDGenerator is a mock implementation of ports.IDGenerator.
type DealMockIDGenerator struct {
	codeSeq int
}

func NewDealMockIDGenerator() *DealMockIDGenerator {
	return &DealMockIDGenerator{}
}

func (m *DealMockIDGenerator) GenerateID() uuid.UUID {
	return uuid.New()
}

func (m *DealMockIDGenerator) GenerateDealNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	m.codeSeq++
	return "DEAL-001", nil
}

func (m *DealMockIDGenerator) GenerateLeadNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	m.codeSeq++
	return "LEAD-001", nil
}

// DealMockNotificationService is a mock implementation of ports.NotificationService.
type DealMockNotificationService struct{}

func NewDealMockNotificationService() *DealMockNotificationService {
	return &DealMockNotificationService{}
}

func (m *DealMockNotificationService) SendEmail(ctx context.Context, req ports.EmailRequest) error {
	return nil
}

func (m *DealMockNotificationService) SendInApp(ctx context.Context, req ports.InAppNotificationRequest) error {
	return nil
}

func (m *DealMockNotificationService) SendSMS(ctx context.Context, req ports.SMSRequest) error {
	return nil
}

func (m *DealMockNotificationService) SendBatch(ctx context.Context, notifications []ports.NotificationRequest) error {
	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func createDealTestDeal(tenantID uuid.UUID) *domain.Deal {
	deal := &domain.Deal{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Code:          "DEAL-001",
		Name:          "Test Deal",
		CustomerID:    uuid.New(),
		OpportunityID: uuid.New(),
		OwnerID:       uuid.New(),
		Status:        domain.DealStatusDraft,
		Version:       1,
	}
	return deal
}

func createDealTestOpportunity(tenantID uuid.UUID) *domain.Opportunity {
	amount, _ := domain.NewMoney(10000, "USD")
	opp := &domain.Opportunity{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Code:       "OPP-001",
		Name:       "Test Opportunity",
		CustomerID: uuid.New(),
		OwnerID:    uuid.New(),
		Status:     domain.OpportunityStatusWon,
		Amount:     amount,
		Version:    1,
	}
	return opp
}

// ============================================================================
// DealUseCase Tests - GetByID
// ============================================================================

func TestDealUseCase_GetByID_Success(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	tenantID := uuid.New()
	deal := createDealTestDeal(tenantID)
	dealRepo.deals[deal.ID] = deal

	// Add customer info
	customerService.customers[deal.CustomerID] = &ports.CustomerInfo{
		ID:   deal.CustomerID,
		Name: "Test Customer",
	}

	// Act
	result, err := uc.GetByID(context.Background(), tenantID, deal.ID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.ID != deal.ID.String() {
		t.Errorf("Expected ID %s, got %s", deal.ID.String(), result.ID)
	}
}

func TestDealUseCase_GetByID_NotFound(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	// Act
	_, err := uc.GetByID(context.Background(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for deal not found, got nil")
	}
}

// ============================================================================
// DealUseCase Tests - GetByCode
// ============================================================================

func TestDealUseCase_GetByCode_Success(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	tenantID := uuid.New()
	deal := createDealTestDeal(tenantID)
	dealRepo.deals[deal.ID] = deal

	customerService.customers[deal.CustomerID] = &ports.CustomerInfo{
		ID:   deal.CustomerID,
		Name: "Test Customer",
	}

	// Act
	result, err := uc.GetByCode(context.Background(), tenantID, deal.Code)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestDealUseCase_GetByCode_NotFound(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	// Act
	_, err := uc.GetByCode(context.Background(), uuid.New(), "NONEXISTENT")

	// Assert
	if err == nil {
		t.Fatal("Expected error for deal not found, got nil")
	}
}

// ============================================================================
// DealUseCase Tests - List
// ============================================================================

func TestDealUseCase_List_Success(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	tenantID := uuid.New()
	deal1 := createDealTestDeal(tenantID)
	deal2 := createDealTestDeal(tenantID)
	dealRepo.deals[deal1.ID] = deal1
	dealRepo.deals[deal2.ID] = deal2

	customerService.customers[deal1.CustomerID] = &ports.CustomerInfo{ID: deal1.CustomerID, Name: "Customer 1"}
	customerService.customers[deal2.CustomerID] = &ports.CustomerInfo{ID: deal2.CustomerID, Name: "Customer 2"}

	filter := &dto.DealFilterRequest{
		Page:     1,
		PageSize: 10,
	}

	// Act
	result, err := uc.List(context.Background(), tenantID, filter)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

// ============================================================================
// DealUseCase Tests - Update
// ============================================================================

func TestDealUseCase_Update_Success(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	tenantID := uuid.New()
	userID := uuid.New()
	deal := createDealTestDeal(tenantID)
	dealRepo.deals[deal.ID] = deal

	customerService.customers[deal.CustomerID] = &ports.CustomerInfo{
		ID:   deal.CustomerID,
		Name: "Test Customer",
	}

	newName := "Updated Deal Name"
	req := &dto.UpdateDealRequest{
		Name:    &newName,
		Version: 1,
	}

	// Act
	result, err := uc.Update(context.Background(), tenantID, deal.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestDealUseCase_Update_DealNotFound(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	req := &dto.UpdateDealRequest{
		Version: 1,
	}

	// Act
	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for deal not found, got nil")
	}
}

func TestDealUseCase_Update_VersionMismatch(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	tenantID := uuid.New()
	deal := createDealTestDeal(tenantID)
	deal.Version = 5
	dealRepo.deals[deal.ID] = deal

	req := &dto.UpdateDealRequest{
		Version: 3, // Outdated version
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, deal.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for version mismatch, got nil")
	}
}

// ============================================================================
// DealUseCase Tests - Delete
// ============================================================================

func TestDealUseCase_Delete_Success(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	tenantID := uuid.New()
	userID := uuid.New()
	deal := createDealTestDeal(tenantID)
	dealRepo.deals[deal.ID] = deal

	// Act
	err := uc.Delete(context.Background(), tenantID, deal.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestDealUseCase_Delete_DealNotFound(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	// Act
	err := uc.Delete(context.Background(), uuid.New(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for deal not found, got nil")
	}
}

// ============================================================================
// Table-Driven Tests
// ============================================================================

func TestDealUseCase_GetByID_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  uuid.UUID
		dealID    uuid.UUID
		setupRepo func(*DealMockDealRepository)
		expectErr bool
	}{
		{
			name:     "valid deal",
			tenantID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			dealID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			setupRepo: func(repo *DealMockDealRepository) {
				deal := createDealTestDeal(uuid.MustParse("11111111-1111-1111-1111-111111111111"))
				deal.ID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
				repo.deals[deal.ID] = deal
			},
			expectErr: false,
		},
		{
			name:      "deal not found",
			tenantID:  uuid.New(),
			dealID:    uuid.New(),
			setupRepo: func(repo *DealMockDealRepository) {},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dealRepo := NewDealMockDealRepository()
			oppRepo := NewDealMockOpportunityRepository()
			eventPublisher := NewDealMockEventPublisher()
			customerService := NewDealMockCustomerService()
			userService := NewDealMockUserService()
			productService := NewDealMockProductService()
			cacheService := NewDealMockCacheService()
			searchService := NewDealMockSearchService()
			idGenerator := NewDealMockIDGenerator()
			notificationSvc := NewDealMockNotificationService()

			tt.setupRepo(dealRepo)

			// Add customer for valid deal
			if !tt.expectErr {
				for _, deal := range dealRepo.deals {
					customerService.customers[deal.CustomerID] = &ports.CustomerInfo{
						ID:   deal.CustomerID,
						Name: "Test Customer",
					}
				}
			}

			uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

			_, err := uc.GetByID(context.Background(), tt.tenantID, tt.dealID)

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

func BenchmarkDealUseCase_GetByID(b *testing.B) {
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	tenantID := uuid.New()
	deal := createDealTestDeal(tenantID)
	dealRepo.deals[deal.ID] = deal
	customerService.customers[deal.CustomerID] = &ports.CustomerInfo{ID: deal.CustomerID, Name: "Customer"}

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetByID(ctx, tenantID, deal.ID)
	}
}

func BenchmarkDealUseCase_List(b *testing.B) {
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	tenantID := uuid.New()
	for i := 0; i < 10; i++ {
		deal := createDealTestDeal(tenantID)
		dealRepo.deals[deal.ID] = deal
		customerService.customers[deal.CustomerID] = &ports.CustomerInfo{ID: deal.CustomerID, Name: "Customer"}
	}

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)
	ctx := context.Background()

	filter := &dto.DealFilterRequest{Page: 1, PageSize: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.List(ctx, tenantID, filter)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestDealUseCase_GetByID_ContextTimeout(t *testing.T) {
	// Arrange
	dealRepo := NewDealMockDealRepository()
	oppRepo := NewDealMockOpportunityRepository()
	eventPublisher := NewDealMockEventPublisher()
	customerService := NewDealMockCustomerService()
	userService := NewDealMockUserService()
	productService := NewDealMockProductService()
	cacheService := NewDealMockCacheService()
	searchService := NewDealMockSearchService()
	idGenerator := NewDealMockIDGenerator()
	notificationSvc := NewDealMockNotificationService()

	uc := NewDealUseCase(dealRepo, oppRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator, notificationSvc)

	tenantID := uuid.New()
	deal := createDealTestDeal(tenantID)
	dealRepo.deals[deal.ID] = deal
	customerService.customers[deal.CustomerID] = &ports.CustomerInfo{ID: deal.CustomerID, Name: "Customer"}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - just ensure no panic
	_, _ = uc.GetByID(ctx, tenantID, deal.ID)
}

// ============================================================================
// Type Aliases for Shared Mocks (used by other test files in package)
// ============================================================================

// Type aliases for mocks used by opportunity_usecase_test.go and lead_usecase_test.go
type MockDealRepository = DealMockDealRepository
type MockSalesEventPublisher = DealMockEventPublisher
type MockSalesCacheService = DealMockCacheService
type MockSearchService = DealMockSearchService
type MockSalesIDGenerator = DealMockIDGenerator
type MockOpportunityRepository = DealMockOpportunityRepository
type MockCustomerService = DealMockCustomerService
type MockUserService = DealMockUserService
type MockProductService = DealMockProductService
type MockNotificationService = DealMockNotificationService

// Constructor aliases
func NewMockDealRepository() *MockDealRepository {
	return NewDealMockDealRepository()
}

func NewMockSalesEventPublisher() *MockSalesEventPublisher {
	return NewDealMockEventPublisher()
}

func NewMockSalesCacheService() *MockSalesCacheService {
	return NewDealMockCacheService()
}

func NewMockSearchService() *MockSearchService {
	return NewDealMockSearchService()
}

func NewMockSalesIDGenerator() *MockSalesIDGenerator {
	return NewDealMockIDGenerator()
}

func NewMockOpportunityRepository() *MockOpportunityRepository {
	return NewDealMockOpportunityRepository()
}

func NewMockCustomerService() *MockCustomerService {
	return NewDealMockCustomerService()
}

func NewMockUserService() *MockUserService {
	return NewDealMockUserService()
}

func NewMockProductService() *MockProductService {
	return NewDealMockProductService()
}

func NewMockNotificationService() *MockNotificationService {
	return NewDealMockNotificationService()
}
