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
// Extended Mock Implementations for Opportunity Tests
// ============================================================================

// ExtendedMockOpportunityRepository extends MockOpportunityRepository with full functionality.
type ExtendedMockOpportunityRepository struct {
	opportunities     map[uuid.UUID]*domain.Opportunity
	createErr         error
	updateErr         error
	deleteErr         error
	getByIDErr        error
	listErr           error
	countByStatusErr  error
	bulkUpdateOwnerErr error
	bulkUpdateStageErr error
}

func NewExtendedMockOpportunityRepository() *ExtendedMockOpportunityRepository {
	return &ExtendedMockOpportunityRepository{
		opportunities: make(map[uuid.UUID]*domain.Opportunity),
	}
}

func (m *ExtendedMockOpportunityRepository) Create(ctx context.Context, opp *domain.Opportunity) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.opportunities[opp.ID] = opp
	return nil
}

func (m *ExtendedMockOpportunityRepository) Update(ctx context.Context, opp *domain.Opportunity) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.opportunities[opp.ID] = opp
	return nil
}

func (m *ExtendedMockOpportunityRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.opportunities, id)
	return nil
}

func (m *ExtendedMockOpportunityRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Opportunity, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	opp, ok := m.opportunities[id]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	return opp, nil
}

func (m *ExtendedMockOpportunityRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.OpportunityFilter, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.OpportunityStatus, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.Status == status {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetOpenOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return m.GetByStatus(ctx, tenantID, domain.OpportunityStatusOpen, opts)
}

func (m *ExtendedMockOpportunityRepository) GetWonOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return m.GetByStatus(ctx, tenantID, domain.OpportunityStatusWon, opts)
}

func (m *ExtendedMockOpportunityRepository) GetLostOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return m.GetByStatus(ctx, tenantID, domain.OpportunityStatusLost, opts)
}

func (m *ExtendedMockOpportunityRepository) GetByPipeline(ctx context.Context, tenantID, pipelineID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.PipelineID == pipelineID {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetByStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.PipelineID == pipelineID && opp.StageID == stageID {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.CustomerID == customerID {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetByContact(ctx context.Context, tenantID, contactID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID {
			for _, c := range opp.Contacts {
				if c.ContactID == contactID {
					result = append(result, opp)
					break
				}
			}
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetByLead(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.Opportunity, error) {
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.LeadID != nil && *opp.LeadID == leadID {
			return opp, nil
		}
	}
	return nil, errors.New("opportunity not found")
}

func (m *ExtendedMockOpportunityRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.OwnerID == ownerID {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetOverdueOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	now := time.Now()
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.ExpectedCloseDate != nil && opp.ExpectedCloseDate.Before(now) && opp.Status == domain.OpportunityStatusOpen {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetByExpectedCloseDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.ExpectedCloseDate != nil {
			if (opp.ExpectedCloseDate.After(start) || opp.ExpectedCloseDate.Equal(start)) &&
				(opp.ExpectedCloseDate.Before(end) || opp.ExpectedCloseDate.Equal(end)) {
				result = append(result, opp)
			}
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetHighValueOpportunities(ctx context.Context, tenantID uuid.UUID, minAmount int64, currency string, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.Amount.Amount >= minAmount && opp.Amount.Currency == currency {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) CountByStage(ctx context.Context, tenantID, pipelineID uuid.UUID) (map[uuid.UUID]int64, error) {
	counts := make(map[uuid.UUID]int64)
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.PipelineID == pipelineID {
			counts[opp.StageID]++
		}
	}
	return counts, nil
}

func (m *ExtendedMockOpportunityRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.OpportunityStatus]int64, error) {
	if m.countByStatusErr != nil {
		return nil, m.countByStatusErr
	}
	counts := make(map[domain.OpportunityStatus]int64)
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID {
			counts[opp.Status]++
		}
	}
	return counts, nil
}


func (m *ExtendedMockOpportunityRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, filter domain.OpportunityFilter, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *ExtendedMockOpportunityRepository) BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, ownerID uuid.UUID) error {
	if m.bulkUpdateOwnerErr != nil {
		return m.bulkUpdateOwnerErr
	}
	for _, id := range opportunityIDs {
		if opp, ok := m.opportunities[id]; ok {
			opp.OwnerID = ownerID
		}
	}
	return nil
}

func (m *ExtendedMockOpportunityRepository) BulkUpdateStage(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, stageID uuid.UUID) error {
	if m.bulkUpdateStageErr != nil {
		return m.bulkUpdateStageErr
	}
	for _, id := range opportunityIDs {
		if opp, ok := m.opportunities[id]; ok {
			opp.StageID = stageID
		}
	}
	return nil
}

func (m *ExtendedMockOpportunityRepository) GetTotalPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	var total int64
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.Status == domain.OpportunityStatusOpen {
			total += opp.Amount.Amount
		}
	}
	return total, nil
}

func (m *ExtendedMockOpportunityRepository) GetWeightedPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	var total int64
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.Status == domain.OpportunityStatusOpen {
			total += opp.WeightedAmount.Amount
		}
	}
	return total, nil
}

func (m *ExtendedMockOpportunityRepository) GetWinRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	var won, total int64
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID {
			if opp.Status == domain.OpportunityStatusWon {
				won++
			}
			if opp.Status.IsClosed() {
				total++
			}
		}
	}
	if total == 0 {
		return 0, nil
	}
	return float64(won) / float64(total) * 100, nil
}

func (m *ExtendedMockOpportunityRepository) GetAverageDealSize(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	var total int64
	var count int64
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.Status == domain.OpportunityStatusWon {
			total += opp.Amount.Amount
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return total / count, nil
}

func (m *ExtendedMockOpportunityRepository) GetAverageSalesCycle(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (int, error) {
	return 30, nil
}

func (m *ExtendedMockOpportunityRepository) GetClosingThisMonth(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	now := time.Now()
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.ExpectedCloseDate != nil {
			if opp.ExpectedCloseDate.Month() == now.Month() && opp.ExpectedCloseDate.Year() == now.Year() {
				result = append(result, opp)
			}
		}
	}
	return result, int64(len(result)), nil
}

func (m *ExtendedMockOpportunityRepository) GetClosingThisQuarter(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.Status == domain.OpportunityStatusOpen {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

// NOTE: MockPipelineRepository is defined in lead_usecase_test.go and reused here

// ExtendedMockUserService extends MockUserService for opportunity tests.
type ExtendedMockUserService struct {
	users       map[uuid.UUID]*ports.UserInfo
	getErr      error
	existsError error
}

func NewExtendedMockUserService() *ExtendedMockUserService {
	return &ExtendedMockUserService{
		users: make(map[uuid.UUID]*ports.UserInfo),
	}
}

func (m *ExtendedMockUserService) GetUser(ctx context.Context, tenantID, userID uuid.UUID) (*ports.UserInfo, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	user, ok := m.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *ExtendedMockUserService) UserExists(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	if m.existsError != nil {
		return false, m.existsError
	}
	_, ok := m.users[userID]
	return ok, nil
}

func (m *ExtendedMockUserService) GetUsersByIDs(ctx context.Context, tenantID uuid.UUID, userIDs []uuid.UUID) ([]*ports.UserInfo, error) {
	var result []*ports.UserInfo
	for _, id := range userIDs {
		if user, ok := m.users[id]; ok {
			result = append(result, user)
		}
	}
	return result, nil
}

func (m *ExtendedMockUserService) GetUsersByRole(ctx context.Context, tenantID uuid.UUID, roleName string) ([]*ports.UserInfo, error) {
	return []*ports.UserInfo{}, nil
}

func (m *ExtendedMockUserService) HasPermission(ctx context.Context, tenantID, userID uuid.UUID, permission string) (bool, error) {
	return true, nil
}

// ExtendedMockCustomerService extends MockCustomerService for opportunity tests.
type ExtendedMockCustomerService struct {
	customers     map[uuid.UUID]*ports.CustomerInfo
	contacts      map[uuid.UUID]*ports.ContactInfo
	getErr        error
	existsErr     error
	contactExistsErr error
}

func NewExtendedMockCustomerService() *ExtendedMockCustomerService {
	return &ExtendedMockCustomerService{
		customers: make(map[uuid.UUID]*ports.CustomerInfo),
		contacts:  make(map[uuid.UUID]*ports.ContactInfo),
	}
}

func (m *ExtendedMockCustomerService) GetCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*ports.CustomerInfo, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	customer, ok := m.customers[customerID]
	if !ok {
		return nil, errors.New("customer not found")
	}
	return customer, nil
}

func (m *ExtendedMockCustomerService) GetCustomerByCode(ctx context.Context, tenantID uuid.UUID, code string) (*ports.CustomerInfo, error) {
	for _, customer := range m.customers {
		if customer.Code == code {
			return customer, nil
		}
	}
	return nil, errors.New("customer not found")
}

func (m *ExtendedMockCustomerService) CustomerExists(ctx context.Context, tenantID, customerID uuid.UUID) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	_, ok := m.customers[customerID]
	return ok, nil
}

func (m *ExtendedMockCustomerService) CreateCustomer(ctx context.Context, tenantID uuid.UUID, req ports.CreateCustomerRequest) (*ports.CustomerInfo, error) {
	customer := &ports.CustomerInfo{
		ID:   uuid.New(),
		Name: req.Name,
	}
	m.customers[customer.ID] = customer
	return customer, nil
}

func (m *ExtendedMockCustomerService) GetContact(ctx context.Context, tenantID, contactID uuid.UUID) (*ports.ContactInfo, error) {
	contact, ok := m.contacts[contactID]
	if !ok {
		return nil, errors.New("contact not found")
	}
	return contact, nil
}

func (m *ExtendedMockCustomerService) ContactExists(ctx context.Context, tenantID, contactID uuid.UUID) (bool, error) {
	if m.contactExistsErr != nil {
		return false, m.contactExistsErr
	}
	_, ok := m.contacts[contactID]
	return ok, nil
}

func (m *ExtendedMockCustomerService) CreateContact(ctx context.Context, tenantID, customerID uuid.UUID, req ports.CreateContactRequest) (*ports.ContactInfo, error) {
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

// ExtendedMockProductService extends MockProductService for opportunity tests.
type ExtendedMockProductService struct {
	products    map[uuid.UUID]*ports.ProductInfo
	getErr      error
	existsErr   error
}

func NewExtendedMockProductService() *ExtendedMockProductService {
	return &ExtendedMockProductService{
		products: make(map[uuid.UUID]*ports.ProductInfo),
	}
}

func (m *ExtendedMockProductService) GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (*ports.ProductInfo, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	product, ok := m.products[productID]
	if !ok {
		return nil, errors.New("product not found")
	}
	return product, nil
}

func (m *ExtendedMockProductService) ProductExists(ctx context.Context, tenantID, productID uuid.UUID) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	_, ok := m.products[productID]
	return ok, nil
}

func (m *ExtendedMockProductService) GetProductsByIDs(ctx context.Context, tenantID uuid.UUID, productIDs []uuid.UUID) ([]*ports.ProductInfo, error) {
	var result []*ports.ProductInfo
	for _, id := range productIDs {
		if product, ok := m.products[id]; ok {
			result = append(result, product)
		}
	}
	return result, nil
}

func (m *ExtendedMockProductService) SearchProducts(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]*ports.ProductInfo, error) {
	return []*ports.ProductInfo{}, nil
}

// ============================================================================
// Helper Functions for Opportunity Tests
// ============================================================================

// createOpportunityTestPipeline creates a pipeline with all necessary stages for opportunity testing.
// NOTE: This is different from createTestPipeline in lead_usecase_test.go as it includes
// explicit Won/Lost stages needed for Win/Lose operations.
func createOpportunityTestPipeline(tenantID uuid.UUID) *domain.Pipeline {
	pipelineID := uuid.New()
	now := time.Now().UTC()

	// Create stages
	stages := []*domain.Stage{
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Qualified",
			Type:        domain.StageTypeQualifying,
			Order:       1,
			Probability: 20,
			IsActive:    true,
			Color:       "#3498db",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Proposal",
			Type:        domain.StageTypeNegotiating,
			Order:       2,
			Probability: 60,
			IsActive:    true,
			Color:       "#f39c12",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Won",
			Type:        domain.StageTypeWon,
			Order:       3,
			Probability: 100,
			IsActive:    true,
			Color:       "#27ae60",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Lost",
			Type:        domain.StageTypeLost,
			Order:       4,
			Probability: 0,
			IsActive:    true,
			Color:       "#e74c3c",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	totalValue, _ := domain.Zero("USD")
	wonValue, _ := domain.Zero("USD")

	return &domain.Pipeline{
		ID:          pipelineID,
		TenantID:    tenantID,
		Name:        "Default Pipeline",
		IsDefault:   true,
		IsActive:    true,
		Currency:    "USD",
		Stages:      stages,
		TotalValue:  totalValue,
		WonValue:    wonValue,
		CreatedBy:   uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
}

func createTestOpportunityWithPipeline(tenantID uuid.UUID, pipeline *domain.Pipeline) *domain.Opportunity {
	amount, _ := domain.NewMoney(10000, "USD")
	weightedAmount, _ := domain.NewMoney(2000, "USD")
	now := time.Now().UTC()
	expectedClose := now.AddDate(0, 1, 0)

	return &domain.Opportunity{
		ID:                uuid.New(),
		TenantID:          tenantID,
		Code:              "OPP-001",
		Name:              "Test Opportunity",
		Description:       "Test opportunity description",
		Status:            domain.OpportunityStatusOpen,
		Priority:          domain.OpportunityPriorityMedium,
		PipelineID:        pipeline.ID,
		PipelineName:      pipeline.Name,
		StageID:           pipeline.Stages[0].ID,
		StageName:         pipeline.Stages[0].Name,
		StageEnteredAt:    now,
		StageHistory:      []domain.StageHistory{},
		CustomerID:        uuid.New(),
		CustomerName:      "Test Customer",
		Contacts:          []domain.OpportunityContact{},
		Amount:            amount,
		WeightedAmount:    weightedAmount,
		Probability:       20,
		Products:          []domain.OpportunityProduct{},
		ExpectedCloseDate: &expectedClose,
		OwnerID:           uuid.New(),
		OwnerName:         "Test Owner",
		Tags:              []string{"test"},
		CreatedBy:         uuid.New(),
		CreatedAt:         now,
		UpdatedAt:         now,
		Version:           1,
	}
}

func setupOpportunityTestDependencies() (
	*ExtendedMockOpportunityRepository,
	*MockPipelineRepository,
	*MockDealRepository,
	*MockSalesEventPublisher,
	*ExtendedMockCustomerService,
	*ExtendedMockUserService,
	*ExtendedMockProductService,
	*MockSalesCacheService,
	*MockSearchService,
	*MockSalesIDGenerator,
) {
	oppRepo := NewExtendedMockOpportunityRepository()
	pipelineRepo := NewMockPipelineRepository()
	dealRepo := NewMockDealRepository()
	eventPublisher := NewMockSalesEventPublisher()
	customerService := NewExtendedMockCustomerService()
	userService := NewExtendedMockUserService()
	productService := NewExtendedMockProductService()
	cacheService := NewMockSalesCacheService()
	searchService := NewMockSearchService()
	idGenerator := NewMockSalesIDGenerator()

	return oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator
}

// ============================================================================
// OpportunityUseCase Tests - Create
// ============================================================================

func TestOpportunityUseCase_Create_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	customerID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Add customer
	customerService.customers[customerID] = &ports.CustomerInfo{
		ID:   customerID,
		Name: "Test Customer",
		Code: "CUST-001",
	}

	// Add user
	userService.users[userID] = &ports.UserInfo{
		ID:       userID,
		FullName: "Test User",
		Email:    "test@example.com",
	}

	customerIDStr := customerID.String()
	req := &dto.CreateOpportunityRequest{
		Name:              "New Opportunity",
		PipelineID:        pipeline.ID.String(),
		Amount:            5000000,
		Currency:          "USD",
		ExpectedCloseDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		CustomerID:        &customerIDStr,
	}

	// Act
	result, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, result.Name)
	}
	if result.Status != "open" {
		t.Errorf("Expected status 'open', got %s", result.Status)
	}
}

func TestOpportunityUseCase_Create_PipelineNotFound(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()

	req := &dto.CreateOpportunityRequest{
		Name:              "New Opportunity",
		PipelineID:        uuid.New().String(),
		Amount:            5000000,
		Currency:          "USD",
		ExpectedCloseDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

func TestOpportunityUseCase_Create_InvalidPipelineID(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()

	req := &dto.CreateOpportunityRequest{
		Name:              "New Opportunity",
		PipelineID:        "invalid-uuid",
		Amount:            5000000,
		Currency:          "USD",
		ExpectedCloseDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid pipeline ID, got nil")
	}
}

func TestOpportunityUseCase_Create_CustomerNotFound(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	customerID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Add user but not customer
	userService.users[userID] = &ports.UserInfo{
		ID:       userID,
		FullName: "Test User",
	}

	customerIDStr := customerID.String()
	req := &dto.CreateOpportunityRequest{
		Name:              "New Opportunity",
		PipelineID:        pipeline.ID.String(),
		Amount:            5000000,
		Currency:          "USD",
		ExpectedCloseDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		CustomerID:        &customerIDStr,
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer not found, got nil")
	}
}

// ============================================================================
// OpportunityUseCase Tests - GetByID
// ============================================================================

func TestOpportunityUseCase_GetByID_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	// Add customer
	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{
		ID:   opp.CustomerID,
		Name: opp.CustomerName,
	}

	// Add owner
	userService.users[opp.OwnerID] = &ports.UserInfo{
		ID:       opp.OwnerID,
		FullName: opp.OwnerName,
	}

	// Act
	result, err := uc.GetByID(context.Background(), tenantID, opp.ID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.ID != opp.ID.String() {
		t.Errorf("Expected ID %s, got %s", opp.ID.String(), result.ID)
	}
	if result.Name != opp.Name {
		t.Errorf("Expected name %s, got %s", opp.Name, result.Name)
	}
}

func TestOpportunityUseCase_GetByID_NotFound(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	// Act
	_, err := uc.GetByID(context.Background(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for opportunity not found, got nil")
	}
}

// ============================================================================
// OpportunityUseCase Tests - Update
// ============================================================================

func TestOpportunityUseCase_Update_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{
		ID:   opp.CustomerID,
		Name: opp.CustomerName,
	}
	userService.users[opp.OwnerID] = &ports.UserInfo{
		ID:       opp.OwnerID,
		FullName: opp.OwnerName,
	}

	newName := "Updated Opportunity Name"
	req := &dto.UpdateOpportunityRequest{
		Name:    &newName,
		Version: 1,
	}

	// Act
	result, err := uc.Update(context.Background(), tenantID, opp.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, result.Name)
	}
}

func TestOpportunityUseCase_Update_NotFound(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	newName := "Updated Name"
	req := &dto.UpdateOpportunityRequest{
		Name:    &newName,
		Version: 1,
	}

	// Act
	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for opportunity not found, got nil")
	}
}

func TestOpportunityUseCase_Update_VersionMismatch(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp.Version = 5
	oppRepo.opportunities[opp.ID] = opp

	newName := "Updated Name"
	req := &dto.UpdateOpportunityRequest{
		Name:    &newName,
		Version: 3, // Outdated version
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for version mismatch, got nil")
	}
}

func TestOpportunityUseCase_Update_ClosedOpportunity(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp.Status = domain.OpportunityStatusWon
	oppRepo.opportunities[opp.ID] = opp

	newName := "Updated Name"
	req := &dto.UpdateOpportunityRequest{
		Name:    &newName,
		Version: 1,
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for updating closed opportunity, got nil")
	}
}

// ============================================================================
// OpportunityUseCase Tests - Delete
// ============================================================================

func TestOpportunityUseCase_Delete_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	// Act
	err := uc.Delete(context.Background(), tenantID, opp.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify deleted
	if _, exists := oppRepo.opportunities[opp.ID]; exists {
		t.Error("Expected opportunity to be deleted")
	}
}

func TestOpportunityUseCase_Delete_NotFound(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	// Act
	err := uc.Delete(context.Background(), uuid.New(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for opportunity not found, got nil")
	}
}

func TestOpportunityUseCase_Delete_ClosedOpportunity(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp.Status = domain.OpportunityStatusWon
	oppRepo.opportunities[opp.ID] = opp

	// Act
	err := uc.Delete(context.Background(), tenantID, opp.ID, uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for deleting closed opportunity, got nil")
	}
}

// ============================================================================
// OpportunityUseCase Tests - List
// ============================================================================

func TestOpportunityUseCase_List_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp1 := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp2 := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp2.ID = uuid.New()
	opp2.Name = "Second Opportunity"

	oppRepo.opportunities[opp1.ID] = opp1
	oppRepo.opportunities[opp2.ID] = opp2

	customerService.customers[opp1.CustomerID] = &ports.CustomerInfo{ID: opp1.CustomerID, Name: opp1.CustomerName}
	customerService.customers[opp2.CustomerID] = &ports.CustomerInfo{ID: opp2.CustomerID, Name: opp2.CustomerName}
	userService.users[opp1.OwnerID] = &ports.UserInfo{ID: opp1.OwnerID, FullName: opp1.OwnerName}
	userService.users[opp2.OwnerID] = &ports.UserInfo{ID: opp2.OwnerID, FullName: opp2.OwnerName}

	filter := &dto.OpportunityFilterRequest{
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
	if len(result.Opportunities) != 2 {
		t.Errorf("Expected 2 opportunities, got %d", len(result.Opportunities))
	}
}

func TestOpportunityUseCase_List_Empty(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	filter := &dto.OpportunityFilterRequest{
		PageSize: 10,
	}

	// Act
	result, err := uc.List(context.Background(), uuid.New(), filter)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if len(result.Opportunities) != 0 {
		t.Errorf("Expected 0 opportunities, got %d", len(result.Opportunities))
	}
}

// ============================================================================
// OpportunityUseCase Tests - MoveStage
// ============================================================================

func TestOpportunityUseCase_MoveStage_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{ID: opp.CustomerID, Name: opp.CustomerName}
	userService.users[opp.OwnerID] = &ports.UserInfo{ID: opp.OwnerID, FullName: opp.OwnerName}

	// Move to second stage (Proposal)
	newStageID := pipeline.Stages[1].ID
	req := &dto.MoveStageRequest{
		StageID: newStageID.String(),
	}

	// Act
	result, err := uc.MoveStage(context.Background(), tenantID, opp.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.StageID != newStageID.String() {
		t.Errorf("Expected stage ID %s, got %s", newStageID.String(), result.StageID)
	}
}

func TestOpportunityUseCase_MoveStage_OpportunityNotFound(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	req := &dto.MoveStageRequest{
		StageID: uuid.New().String(),
	}

	// Act
	_, err := uc.MoveStage(context.Background(), uuid.New(), uuid.New(), uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for opportunity not found, got nil")
	}
}

func TestOpportunityUseCase_MoveStage_ClosedOpportunity(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp.Status = domain.OpportunityStatusWon
	oppRepo.opportunities[opp.ID] = opp

	req := &dto.MoveStageRequest{
		StageID: pipeline.Stages[1].ID.String(),
	}

	// Act
	_, err := uc.MoveStage(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for moving closed opportunity, got nil")
	}
}

func TestOpportunityUseCase_MoveStage_InvalidStageID(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	req := &dto.MoveStageRequest{
		StageID: "invalid-uuid",
	}

	// Act
	_, err := uc.MoveStage(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid stage ID, got nil")
	}
}

// ============================================================================
// OpportunityUseCase Tests - Win
// ============================================================================

func TestOpportunityUseCase_Win_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{ID: opp.CustomerID, Name: opp.CustomerName}
	userService.users[opp.OwnerID] = &ports.UserInfo{ID: opp.OwnerID, FullName: opp.OwnerName}
	userService.users[userID] = &ports.UserInfo{ID: userID, FullName: "Winner User"}

	req := &dto.WinOpportunityRequest{
		WonReason:  "Great proposal",
		CreateDeal: false,
	}

	// Act
	result, err := uc.Win(context.Background(), tenantID, opp.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Status != "won" {
		t.Errorf("Expected status 'won', got %s", result.Status)
	}
}

func TestOpportunityUseCase_Win_AlreadyWon(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp.Status = domain.OpportunityStatusWon
	oppRepo.opportunities[opp.ID] = opp

	req := &dto.WinOpportunityRequest{
		WonReason: "Great proposal",
	}

	// Act
	_, err := uc.Win(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already won opportunity, got nil")
	}
}

func TestOpportunityUseCase_Win_AlreadyLost(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp.Status = domain.OpportunityStatusLost
	oppRepo.opportunities[opp.ID] = opp

	req := &dto.WinOpportunityRequest{
		WonReason: "Great proposal",
	}

	// Act
	_, err := uc.Win(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already lost opportunity, got nil")
	}
}

// ============================================================================
// OpportunityUseCase Tests - Lose
// ============================================================================

func TestOpportunityUseCase_Lose_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{ID: opp.CustomerID, Name: opp.CustomerName}
	userService.users[opp.OwnerID] = &ports.UserInfo{ID: opp.OwnerID, FullName: opp.OwnerName}
	userService.users[userID] = &ports.UserInfo{ID: userID, FullName: "Loser User"}

	competitorName := "Competitor Inc"
	req := &dto.LoseOpportunityRequest{
		LostReason:     "Price too high",
		CompetitorName: &competitorName,
	}

	// Act
	result, err := uc.Lose(context.Background(), tenantID, opp.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Status != "lost" {
		t.Errorf("Expected status 'lost', got %s", result.Status)
	}
}

func TestOpportunityUseCase_Lose_AlreadyWon(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp.Status = domain.OpportunityStatusWon
	oppRepo.opportunities[opp.ID] = opp

	req := &dto.LoseOpportunityRequest{
		LostReason: "Price too high",
	}

	// Act
	_, err := uc.Lose(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already won opportunity, got nil")
	}
}

func TestOpportunityUseCase_Lose_AlreadyLost(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp.Status = domain.OpportunityStatusLost
	oppRepo.opportunities[opp.ID] = opp

	req := &dto.LoseOpportunityRequest{
		LostReason: "Price too high",
	}

	// Act
	_, err := uc.Lose(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already lost opportunity, got nil")
	}
}

// ============================================================================
// OpportunityUseCase Tests - GetStatistics
// ============================================================================

func TestOpportunityUseCase_GetStatistics_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Add some opportunities
	opp1 := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp1.Status = domain.OpportunityStatusOpen
	opp2 := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp2.ID = uuid.New()
	opp2.Status = domain.OpportunityStatusWon
	opp3 := createTestOpportunityWithPipeline(tenantID, pipeline)
	opp3.ID = uuid.New()
	opp3.Status = domain.OpportunityStatusLost

	oppRepo.opportunities[opp1.ID] = opp1
	oppRepo.opportunities[opp2.ID] = opp2
	oppRepo.opportunities[opp3.ID] = opp3

	// Act
	result, err := uc.GetStatistics(context.Background(), tenantID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.TotalOpportunities != 3 {
		t.Errorf("Expected 3 total opportunities, got %d", result.TotalOpportunities)
	}
	if result.OpenOpportunities != 1 {
		t.Errorf("Expected 1 open opportunity, got %d", result.OpenOpportunities)
	}
	if result.WonOpportunities != 1 {
		t.Errorf("Expected 1 won opportunity, got %d", result.WonOpportunities)
	}
	if result.LostOpportunities != 1 {
		t.Errorf("Expected 1 lost opportunity, got %d", result.LostOpportunities)
	}
}

// ============================================================================
// OpportunityUseCase Tests - Assign
// ============================================================================

func TestOpportunityUseCase_Assign_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	newOwnerID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{ID: opp.CustomerID, Name: opp.CustomerName}
	userService.users[opp.OwnerID] = &ports.UserInfo{ID: opp.OwnerID, FullName: opp.OwnerName}
	userService.users[newOwnerID] = &ports.UserInfo{ID: newOwnerID, FullName: "New Owner"}

	req := &dto.AssignOpportunityRequest{
		OwnerID: newOwnerID.String(),
	}

	// Act
	result, err := uc.Assign(context.Background(), tenantID, opp.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.OwnerID != newOwnerID.String() {
		t.Errorf("Expected owner ID %s, got %s", newOwnerID.String(), result.OwnerID)
	}
}

func TestOpportunityUseCase_Assign_UserNotFound(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	req := &dto.AssignOpportunityRequest{
		OwnerID: uuid.New().String(),
	}

	// Act
	_, err := uc.Assign(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for user not found, got nil")
	}
}

// ============================================================================
// Table-Driven Tests for Validation
// ============================================================================

func TestOpportunityUseCase_Create_ValidationCases(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockPipelineRepository, *ExtendedMockCustomerService, *ExtendedMockUserService)
		request        *dto.CreateOpportunityRequest
		expectErr      bool
		errContains    string
	}{
		{
			name: "valid request",
			setupMocks: func(pipelineRepo *MockPipelineRepository, customerSvc *ExtendedMockCustomerService, userSvc *ExtendedMockUserService) {
				tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
				pipeline := createOpportunityTestPipeline(tenantID)
				pipeline.ID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
				pipelineRepo.pipelines[pipeline.ID] = pipeline

				customerID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
				customerSvc.customers[customerID] = &ports.CustomerInfo{ID: customerID, Name: "Test"}

				userID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
				userSvc.users[userID] = &ports.UserInfo{ID: userID, FullName: "Test User"}
			},
			request: func() *dto.CreateOpportunityRequest {
				customerID := "33333333-3333-3333-3333-333333333333"
				return &dto.CreateOpportunityRequest{
					Name:              "Valid Opportunity",
					PipelineID:        "22222222-2222-2222-2222-222222222222",
					Amount:            5000000,
					Currency:          "USD",
					ExpectedCloseDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
					CustomerID:        &customerID,
				}
			}(),
			expectErr: false,
		},
		{
			name: "invalid pipeline ID format",
			setupMocks: func(pipelineRepo *MockPipelineRepository, customerSvc *ExtendedMockCustomerService, userSvc *ExtendedMockUserService) {
			},
			request: &dto.CreateOpportunityRequest{
				Name:              "Invalid Pipeline",
				PipelineID:        "not-a-uuid",
				Amount:            5000000,
				Currency:          "USD",
				ExpectedCloseDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			},
			expectErr:   true,
			errContains: "invalid",
		},
		{
			name: "pipeline not found",
			setupMocks: func(pipelineRepo *MockPipelineRepository, customerSvc *ExtendedMockCustomerService, userSvc *ExtendedMockUserService) {
			},
			request: &dto.CreateOpportunityRequest{
				Name:              "No Pipeline",
				PipelineID:        uuid.New().String(),
				Amount:            5000000,
				Currency:          "USD",
				ExpectedCloseDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
			},
			expectErr:   true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()
			tt.setupMocks(pipelineRepo, customerService, userService)

			uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

			tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
			userID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

			_, err := uc.Create(context.Background(), tenantID, userID, tt.request)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestOpportunityUseCase_Update_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		setupOpp  func(*ExtendedMockOpportunityRepository, *MockPipelineRepository) (uuid.UUID, uuid.UUID)
		request   *dto.UpdateOpportunityRequest
		expectErr bool
	}{
		{
			name: "valid update",
			setupOpp: func(oppRepo *ExtendedMockOpportunityRepository, pipelineRepo *MockPipelineRepository) (uuid.UUID, uuid.UUID) {
				tenantID := uuid.New()
				pipeline := createOpportunityTestPipeline(tenantID)
				pipelineRepo.pipelines[pipeline.ID] = pipeline

				opp := createTestOpportunityWithPipeline(tenantID, pipeline)
				oppRepo.opportunities[opp.ID] = opp
				return tenantID, opp.ID
			},
			request: func() *dto.UpdateOpportunityRequest {
				name := "Updated Name"
				return &dto.UpdateOpportunityRequest{
					Name:    &name,
					Version: 1,
				}
			}(),
			expectErr: false,
		},
		{
			name: "version mismatch",
			setupOpp: func(oppRepo *ExtendedMockOpportunityRepository, pipelineRepo *MockPipelineRepository) (uuid.UUID, uuid.UUID) {
				tenantID := uuid.New()
				pipeline := createOpportunityTestPipeline(tenantID)
				pipelineRepo.pipelines[pipeline.ID] = pipeline

				opp := createTestOpportunityWithPipeline(tenantID, pipeline)
				opp.Version = 5
				oppRepo.opportunities[opp.ID] = opp
				return tenantID, opp.ID
			},
			request: &dto.UpdateOpportunityRequest{
				Version: 1,
			},
			expectErr: true,
		},
		{
			name: "closed opportunity",
			setupOpp: func(oppRepo *ExtendedMockOpportunityRepository, pipelineRepo *MockPipelineRepository) (uuid.UUID, uuid.UUID) {
				tenantID := uuid.New()
				pipeline := createOpportunityTestPipeline(tenantID)
				pipelineRepo.pipelines[pipeline.ID] = pipeline

				opp := createTestOpportunityWithPipeline(tenantID, pipeline)
				opp.Status = domain.OpportunityStatusWon
				oppRepo.opportunities[opp.ID] = opp
				return tenantID, opp.ID
			},
			request: &dto.UpdateOpportunityRequest{
				Version: 1,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

			tenantID, oppID := tt.setupOpp(oppRepo, pipelineRepo)

			uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

			_, err := uc.Update(context.Background(), tenantID, oppID, uuid.New(), tt.request)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestOpportunityUseCase_GetByID_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  uuid.UUID
		oppID     uuid.UUID
		setupRepo func(*ExtendedMockOpportunityRepository, *MockPipelineRepository)
		expectErr bool
	}{
		{
			name:     "valid opportunity",
			tenantID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			oppID:    uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			setupRepo: func(oppRepo *ExtendedMockOpportunityRepository, pipelineRepo *MockPipelineRepository) {
				tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
				pipeline := createOpportunityTestPipeline(tenantID)
				pipelineRepo.pipelines[pipeline.ID] = pipeline

				opp := createTestOpportunityWithPipeline(tenantID, pipeline)
				opp.ID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
				oppRepo.opportunities[opp.ID] = opp
			},
			expectErr: false,
		},
		{
			name:      "opportunity not found",
			tenantID:  uuid.New(),
			oppID:     uuid.New(),
			setupRepo: func(oppRepo *ExtendedMockOpportunityRepository, pipelineRepo *MockPipelineRepository) {},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

			tt.setupRepo(oppRepo, pipelineRepo)

			// Add mock customer/user for valid opportunities
			if !tt.expectErr {
				for _, opp := range oppRepo.opportunities {
					customerService.customers[opp.CustomerID] = &ports.CustomerInfo{
						ID:   opp.CustomerID,
						Name: opp.CustomerName,
					}
					userService.users[opp.OwnerID] = &ports.UserInfo{
						ID:       opp.OwnerID,
						FullName: opp.OwnerName,
					}
				}
			}

			uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

			_, err := uc.GetByID(context.Background(), tt.tenantID, tt.oppID)

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

func BenchmarkOpportunityUseCase_GetByID(b *testing.B) {
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{ID: opp.CustomerID, Name: opp.CustomerName}
	userService.users[opp.OwnerID] = &ports.UserInfo{ID: opp.OwnerID, FullName: opp.OwnerName}

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetByID(ctx, tenantID, opp.ID)
	}
}

func BenchmarkOpportunityUseCase_List(b *testing.B) {
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Create multiple opportunities
	for i := 0; i < 10; i++ {
		opp := createTestOpportunityWithPipeline(tenantID, pipeline)
		opp.ID = uuid.New()
		oppRepo.opportunities[opp.ID] = opp
		customerService.customers[opp.CustomerID] = &ports.CustomerInfo{ID: opp.CustomerID, Name: opp.CustomerName}
		userService.users[opp.OwnerID] = &ports.UserInfo{ID: opp.OwnerID, FullName: opp.OwnerName}
	}

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)
	ctx := context.Background()
	filter := &dto.OpportunityFilterRequest{PageSize: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.List(ctx, tenantID, filter)
	}
}

func BenchmarkOpportunityUseCase_Update(b *testing.B) {
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{ID: opp.CustomerID, Name: opp.CustomerName}
	userService.users[opp.OwnerID] = &ports.UserInfo{ID: opp.OwnerID, FullName: opp.OwnerName}

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newName := "Updated Name"
		req := &dto.UpdateOpportunityRequest{
			Name:    &newName,
			Version: oppRepo.opportunities[opp.ID].Version,
		}
		_, _ = uc.Update(ctx, tenantID, opp.ID, userID, req)
	}
}

func BenchmarkOpportunityUseCase_GetStatistics(b *testing.B) {
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Create opportunities with various statuses
	for i := 0; i < 50; i++ {
		opp := createTestOpportunityWithPipeline(tenantID, pipeline)
		opp.ID = uuid.New()
		switch i % 3 {
		case 0:
			opp.Status = domain.OpportunityStatusOpen
		case 1:
			opp.Status = domain.OpportunityStatusWon
		case 2:
			opp.Status = domain.OpportunityStatusLost
		}
		oppRepo.opportunities[opp.ID] = opp
	}

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetStatistics(ctx, tenantID)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestOpportunityUseCase_GetByID_ContextTimeout(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	customerService.customers[opp.CustomerID] = &ports.CustomerInfo{ID: opp.CustomerID, Name: opp.CustomerName}
	userService.users[opp.OwnerID] = &ports.UserInfo{ID: opp.OwnerID, FullName: opp.OwnerName}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Let context expire

	// Act - ensure no panic
	_, _ = uc.GetByID(ctx, tenantID, opp.ID)
}

func TestOpportunityUseCase_List_ContextTimeout(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	for i := 0; i < 5; i++ {
		opp := createTestOpportunityWithPipeline(tenantID, pipeline)
		opp.ID = uuid.New()
		oppRepo.opportunities[opp.ID] = opp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	filter := &dto.OpportunityFilterRequest{PageSize: 10}

	// Act - ensure no panic
	_, _ = uc.List(ctx, tenantID, filter)
}

func TestOpportunityUseCase_Update_ContextTimeout(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	newName := "Updated Name"
	req := &dto.UpdateOpportunityRequest{
		Name:    &newName,
		Version: 1,
	}

	// Act - ensure no panic
	_, _ = uc.Update(ctx, tenantID, opp.ID, uuid.New(), req)
}

func TestOpportunityUseCase_Delete_ContextTimeout(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - ensure no panic
	_ = uc.Delete(ctx, tenantID, opp.ID, uuid.New())
}

func TestOpportunityUseCase_Win_ContextTimeout(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	req := &dto.WinOpportunityRequest{
		WonReason: "Test win",
	}

	// Act - ensure no panic
	_, _ = uc.Win(ctx, tenantID, opp.ID, uuid.New(), req)
}

func TestOpportunityUseCase_Lose_ContextTimeout(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	req := &dto.LoseOpportunityRequest{
		LostReason: "Test loss",
	}

	// Act - ensure no panic
	_, _ = uc.Lose(ctx, tenantID, opp.ID, uuid.New(), req)
}

func TestOpportunityUseCase_MoveStage_ContextTimeout(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	req := &dto.MoveStageRequest{
		StageID: pipeline.Stages[1].ID.String(),
	}

	// Act - ensure no panic
	_, _ = uc.MoveStage(ctx, tenantID, opp.ID, uuid.New(), req)
}

func TestOpportunityUseCase_GetStatistics_ContextTimeout(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - ensure no panic
	_, _ = uc.GetStatistics(ctx, tenantID)
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestOpportunityUseCase_Create_RepositoryError(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	userService.users[userID] = &ports.UserInfo{ID: userID, FullName: "Test User"}

	// Set repository error
	oppRepo.createErr = errors.New("database connection failed")

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	req := &dto.CreateOpportunityRequest{
		Name:              "New Opportunity",
		PipelineID:        pipeline.ID.String(),
		Amount:            5000000,
		Currency:          "USD",
		ExpectedCloseDate: time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository failure, got nil")
	}
}

func TestOpportunityUseCase_Update_RepositoryError(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	// Set repository error
	oppRepo.updateErr = errors.New("database connection failed")

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	newName := "Updated Name"
	req := &dto.UpdateOpportunityRequest{
		Name:    &newName,
		Version: 1,
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, opp.ID, uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository failure, got nil")
	}
}

func TestOpportunityUseCase_Delete_RepositoryError(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	opp := createTestOpportunityWithPipeline(tenantID, pipeline)
	oppRepo.opportunities[opp.ID] = opp

	// Set repository error
	oppRepo.deleteErr = errors.New("database connection failed")

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	// Act
	err := uc.Delete(context.Background(), tenantID, opp.ID, uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository failure, got nil")
	}
}

func TestOpportunityUseCase_List_RepositoryError(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	// Set repository error
	oppRepo.listErr = errors.New("database connection failed")

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	filter := &dto.OpportunityFilterRequest{PageSize: 10}

	// Act
	_, err := uc.List(context.Background(), uuid.New(), filter)

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository failure, got nil")
	}
}

func TestOpportunityUseCase_GetStatistics_RepositoryError(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	// Set repository error
	oppRepo.countByStatusErr = errors.New("database connection failed")

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	// Act
	_, err := uc.GetStatistics(context.Background(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository failure, got nil")
	}
}

// ============================================================================
// Additional Edge Case Tests
// ============================================================================

func TestOpportunityUseCase_BulkAssign_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	newOwnerID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Create multiple opportunities
	oppIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		opp := createTestOpportunityWithPipeline(tenantID, pipeline)
		opp.ID = uuid.New()
		oppRepo.opportunities[opp.ID] = opp
		oppIDs[i] = opp.ID.String()
	}

	userService.users[newOwnerID] = &ports.UserInfo{ID: newOwnerID, FullName: "New Owner"}

	req := &dto.BulkAssignOpportunitiesRequest{
		OpportunityIDs: oppIDs,
		OwnerID:        newOwnerID.String(),
	}

	// Act
	err := uc.BulkAssign(context.Background(), tenantID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestOpportunityUseCase_BulkMoveStage_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Create multiple opportunities
	oppIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		opp := createTestOpportunityWithPipeline(tenantID, pipeline)
		opp.ID = uuid.New()
		oppRepo.opportunities[opp.ID] = opp
		oppIDs[i] = opp.ID.String()
	}

	req := &dto.BulkMoveStageRequest{
		OpportunityIDs: oppIDs,
		StageID:        pipeline.Stages[1].ID.String(),
	}

	// Act
	err := uc.BulkMoveStage(context.Background(), tenantID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestOpportunityUseCase_GetPipelineAnalytics_Success(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	tenantID := uuid.New()
	pipeline := createOpportunityTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	result, err := uc.GetPipelineAnalytics(context.Background(), tenantID, pipeline.ID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.PipelineID != pipeline.ID.String() {
		t.Errorf("Expected pipeline ID %s, got %s", pipeline.ID.String(), result.PipelineID)
	}
}

func TestOpportunityUseCase_GetPipelineAnalytics_PipelineNotFound(t *testing.T) {
	// Arrange
	oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator := setupOpportunityTestDependencies()

	uc := NewOpportunityUseCase(oppRepo, pipelineRepo, dealRepo, eventPublisher, customerService, userService, productService, cacheService, searchService, idGenerator)

	// Act
	_, err := uc.GetPipelineAnalytics(context.Background(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}
