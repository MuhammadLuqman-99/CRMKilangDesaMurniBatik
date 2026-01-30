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
// Mock Implementations for Pipeline Tests
// ============================================================================

// PipelineMockRepo is a mock implementation of domain.PipelineRepository.
type PipelineMockRepo struct {
	pipelines        map[uuid.UUID]*domain.Pipeline
	createErr        error
	updateErr        error
	deleteErr        error
	getByIDErr       error
	getDefaultErr    error
	getStatisticsErr error
}

func NewPipelineMockRepo() *PipelineMockRepo {
	return &PipelineMockRepo{
		pipelines: make(map[uuid.UUID]*domain.Pipeline),
	}
}

func (m *PipelineMockRepo) Create(ctx context.Context, pipeline *domain.Pipeline) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.pipelines[pipeline.ID] = pipeline
	return nil
}

func (m *PipelineMockRepo) Update(ctx context.Context, pipeline *domain.Pipeline) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.pipelines[pipeline.ID] = pipeline
	return nil
}

func (m *PipelineMockRepo) Delete(ctx context.Context, tenantID, pipelineID uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.pipelines, pipelineID)
	return nil
}

func (m *PipelineMockRepo) GetByID(ctx context.Context, tenantID, pipelineID uuid.UUID) (*domain.Pipeline, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return nil, errors.New("pipeline not found")
	}
	return pipeline, nil
}

func (m *PipelineMockRepo) List(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Pipeline, int64, error) {
	var result []*domain.Pipeline
	for _, p := range m.pipelines {
		if p.TenantID == tenantID {
			result = append(result, p)
		}
	}
	return result, int64(len(result)), nil
}

func (m *PipelineMockRepo) GetActivePipelines(ctx context.Context, tenantID uuid.UUID) ([]*domain.Pipeline, error) {
	var result []*domain.Pipeline
	for _, p := range m.pipelines {
		if p.TenantID == tenantID && p.IsActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *PipelineMockRepo) GetDefaultPipeline(ctx context.Context, tenantID uuid.UUID) (*domain.Pipeline, error) {
	if m.getDefaultErr != nil {
		return nil, m.getDefaultErr
	}
	for _, p := range m.pipelines {
		if p.TenantID == tenantID && p.IsDefault {
			return p, nil
		}
	}
	return nil, errors.New("no default pipeline found")
}

func (m *PipelineMockRepo) GetStageByID(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*domain.Stage, error) {
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return nil, errors.New("pipeline not found")
	}
	for _, s := range pipeline.Stages {
		if s.ID == stageID {
			return s, nil
		}
	}
	return nil, errors.New("stage not found")
}

func (m *PipelineMockRepo) AddStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *domain.Stage) error {
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return errors.New("pipeline not found")
	}
	pipeline.Stages = append(pipeline.Stages, stage)
	return nil
}

func (m *PipelineMockRepo) UpdateStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *domain.Stage) error {
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return errors.New("pipeline not found")
	}
	for i, s := range pipeline.Stages {
		if s.ID == stage.ID {
			pipeline.Stages[i] = stage
			return nil
		}
	}
	return errors.New("stage not found")
}

func (m *PipelineMockRepo) RemoveStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) error {
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return errors.New("pipeline not found")
	}
	for i, s := range pipeline.Stages {
		if s.ID == stageID {
			pipeline.Stages = append(pipeline.Stages[:i], pipeline.Stages[i+1:]...)
			return nil
		}
	}
	return errors.New("stage not found")
}

func (m *PipelineMockRepo) ReorderStages(ctx context.Context, tenantID, pipelineID uuid.UUID, stageIDs []uuid.UUID) error {
	_, ok := m.pipelines[pipelineID]
	if !ok {
		return errors.New("pipeline not found")
	}
	return nil
}

func (m *PipelineMockRepo) GetPipelineStatistics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*domain.PipelineStatistics, error) {
	if m.getStatisticsErr != nil {
		return nil, m.getStatisticsErr
	}
	totalValue, _ := domain.NewMoney(100000, "USD")
	weightedValue, _ := domain.NewMoney(50000, "USD")
	return &domain.PipelineStatistics{
		PipelineID:         pipelineID,
		TotalOpportunities: 10,
		OpenOpportunities:  5,
		WonOpportunities:   3,
		LostOpportunities:  2,
		TotalValue:         totalValue,
		WeightedValue:      weightedValue,
		WinRate:            60.0,
		AverageSalesCycle:  30,
		StageDistribution:  make(map[uuid.UUID]int64),
		ConversionRates:    make(map[uuid.UUID]float64),
	}, nil
}

func (m *PipelineMockRepo) GetStageStatistics(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*domain.StageStatistics, error) {
	totalValue, _ := domain.NewMoney(25000, "USD")
	return &domain.StageStatistics{
		StageID:            stageID,
		TotalOpportunities: 5,
		TotalValue:         totalValue,
		AverageTimeInStage: 7,
		ConversionRate:     80.0,
	}, nil
}

// MockPipelineOpportunityRepository is a mock for opportunity queries in pipeline tests.
type MockPipelineOpportunityRepository struct {
	opportunities map[uuid.UUID]*domain.Opportunity
}

func NewMockPipelineOpportunityRepository() *MockPipelineOpportunityRepository {
	return &MockPipelineOpportunityRepository{
		opportunities: make(map[uuid.UUID]*domain.Opportunity),
	}
}

func (m *MockPipelineOpportunityRepository) Create(ctx context.Context, opp *domain.Opportunity) error {
	m.opportunities[opp.ID] = opp
	return nil
}

func (m *MockPipelineOpportunityRepository) Update(ctx context.Context, opp *domain.Opportunity) error {
	m.opportunities[opp.ID] = opp
	return nil
}

func (m *MockPipelineOpportunityRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	delete(m.opportunities, id)
	return nil
}

func (m *MockPipelineOpportunityRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Opportunity, error) {
	opp, ok := m.opportunities[id]
	if !ok {
		return nil, errors.New("opportunity not found")
	}
	return opp, nil
}

func (m *MockPipelineOpportunityRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.OpportunityFilter, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockPipelineOpportunityRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.OpportunityStatus, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetOpenOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.Status == domain.OpportunityStatusOpen {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockPipelineOpportunityRepository) GetWonOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetLostOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetByPipeline(ctx context.Context, tenantID, pipelineID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.PipelineID == pipelineID {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockPipelineOpportunityRepository) GetByStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	var result []*domain.Opportunity
	for _, opp := range m.opportunities {
		if opp.TenantID == tenantID && opp.PipelineID == pipelineID && opp.StageID == stageID {
			result = append(result, opp)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockPipelineOpportunityRepository) GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetByContact(ctx context.Context, tenantID, contactID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetByLead(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.Opportunity, error) {
	return nil, errors.New("not found")
}

func (m *MockPipelineOpportunityRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetClosingThisMonth(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetClosingThisQuarter(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetOverdueOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetByExpectedCloseDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetHighValueOpportunities(ctx context.Context, tenantID uuid.UUID, minAmount int64, currency string, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockPipelineOpportunityRepository) GetTotalPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	return 0, nil
}

func (m *MockPipelineOpportunityRepository) GetWeightedPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	return 0, nil
}

func (m *MockPipelineOpportunityRepository) BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, newOwnerID uuid.UUID) error {
	return nil
}

func (m *MockPipelineOpportunityRepository) BulkUpdateStage(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, stageID uuid.UUID) error {
	return nil
}

func (m *MockPipelineOpportunityRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.OpportunityStatus]int64, error) {
	return make(map[domain.OpportunityStatus]int64), nil
}

func (m *MockPipelineOpportunityRepository) CountByStage(ctx context.Context, tenantID, pipelineID uuid.UUID) (map[uuid.UUID]int64, error) {
	return make(map[uuid.UUID]int64), nil
}

func (m *MockPipelineOpportunityRepository) GetWinRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	return 0, nil
}

func (m *MockPipelineOpportunityRepository) GetAverageDealSize(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	return 0, nil
}

func (m *MockPipelineOpportunityRepository) GetAverageSalesCycle(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (int, error) {
	return 0, nil
}

// MockPipelineEventPublisher is a mock implementation of ports.EventPublisher for pipeline tests.
type MockPipelineEventPublisher struct {
	events []ports.Event
}

func NewMockPipelineEventPublisher() *MockPipelineEventPublisher {
	return &MockPipelineEventPublisher{
		events: make([]ports.Event, 0),
	}
}

func (m *MockPipelineEventPublisher) Publish(ctx context.Context, event ports.Event) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockPipelineEventPublisher) PublishBatch(ctx context.Context, events []ports.Event) error {
	m.events = append(m.events, events...)
	return nil
}

func (m *MockPipelineEventPublisher) PublishAsync(ctx context.Context, event ports.Event) error {
	m.events = append(m.events, event)
	return nil
}

// MockPipelineCacheService is a mock implementation of ports.CacheService for pipeline tests.
type MockPipelineCacheService struct {
	cache map[string][]byte
}

func NewMockPipelineCacheService() *MockPipelineCacheService {
	return &MockPipelineCacheService{
		cache: make(map[string][]byte),
	}
}

func (m *MockPipelineCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	v, ok := m.cache[key]
	if !ok {
		return nil, errors.New("cache miss")
	}
	return v, nil
}

func (m *MockPipelineCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.cache[key] = value
	return nil
}

func (m *MockPipelineCacheService) Delete(ctx context.Context, key string) error {
	delete(m.cache, key)
	return nil
}

func (m *MockPipelineCacheService) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockPipelineCacheService) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.cache[key]
	return ok, nil
}

func (m *MockPipelineCacheService) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	for _, key := range keys {
		if v, ok := m.cache[key]; ok {
			result[key] = v
		}
	}
	return result, nil
}

func (m *MockPipelineCacheService) SetMulti(ctx context.Context, items map[string][]byte, ttl time.Duration) error {
	for k, v := range items {
		m.cache[k] = v
	}
	return nil
}

func (m *MockPipelineCacheService) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return delta, nil
}

func (m *MockPipelineCacheService) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	if _, ok := m.cache[key]; ok {
		return false, nil
	}
	m.cache[key] = value
	return true, nil
}

// MockPipelineIDGenerator is a mock implementation of ports.IDGenerator for pipeline tests.
type MockPipelineIDGenerator struct {
	codeSeq int
}

func NewMockPipelineIDGenerator() *MockPipelineIDGenerator {
	return &MockPipelineIDGenerator{}
}

func (m *MockPipelineIDGenerator) GenerateID() uuid.UUID {
	return uuid.New()
}

func (m *MockPipelineIDGenerator) GenerateDealNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	m.codeSeq++
	return "DEAL-" + string(rune('0'+m.codeSeq)), nil
}

func (m *MockPipelineIDGenerator) GenerateOpportunityNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	m.codeSeq++
	return "OPP-" + string(rune('0'+m.codeSeq)), nil
}

func (m *MockPipelineIDGenerator) GenerateLeadNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	m.codeSeq++
	return "LEAD-" + string(rune('0'+m.codeSeq)), nil
}

func (m *MockPipelineIDGenerator) GenerateInvoiceNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	m.codeSeq++
	return "INV-" + string(rune('0'+m.codeSeq)), nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func createPipelineForTest(tenantID uuid.UUID) *domain.Pipeline {
	pipelineID := uuid.New()
	now := time.Now().UTC()
	totalValue, _ := domain.NewMoney(0, "USD")
	wonValue, _ := domain.NewMoney(0, "USD")

	stages := createDefaultTestStages(pipelineID)

	return &domain.Pipeline{
		ID:          pipelineID,
		TenantID:    tenantID,
		Name:        "Test Pipeline",
		Description: "Test pipeline description",
		IsDefault:   false,
		IsActive:    true,
		Currency:    "USD",
		Stages:      stages,
		WinReasons:  []string{"Price", "Product Fit"},
		LossReasons: []string{"Too Expensive", "Competitor Won"},
		TotalValue:  totalValue,
		WonValue:    wonValue,
		CreatedBy:   uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
}

func createDefaultTestStages(pipelineID uuid.UUID) []*domain.Stage {
	now := time.Now().UTC()
	return []*domain.Stage{
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Qualification",
			Type:        domain.StageTypeQualifying,
			Order:       1,
			Probability: 10,
			Color:       "#3498db",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Proposal",
			Type:        domain.StageTypeOpen,
			Order:       2,
			Probability: 40,
			Color:       "#9b59b6",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Negotiation",
			Type:        domain.StageTypeNegotiating,
			Order:       3,
			Probability: 70,
			Color:       "#f39c12",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Won",
			Type:        domain.StageTypeWon,
			Order:       4,
			Probability: 100,
			Color:       "#27ae60",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Lost",
			Type:        domain.StageTypeLost,
			Order:       5,
			Probability: 0,
			Color:       "#e74c3c",
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

func createTestStage(pipelineID uuid.UUID, name string, stageType domain.StageType, order int, probability int) *domain.Stage {
	now := time.Now().UTC()
	return &domain.Stage{
		ID:          uuid.New(),
		PipelineID:  pipelineID,
		Name:        name,
		Type:        stageType,
		Order:       order,
		Probability: probability,
		Color:       "#3498db",
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func setupPipelineUseCase() (*pipelineUseCase, *PipelineMockRepo, *MockPipelineOpportunityRepository) {
	pipelineRepo := NewPipelineMockRepo()
	oppRepo := NewMockPipelineOpportunityRepository()
	eventPublisher := NewMockPipelineEventPublisher()
	cacheService := NewMockPipelineCacheService()
	idGenerator := NewMockPipelineIDGenerator()

	uc := NewPipelineUseCase(pipelineRepo, oppRepo, eventPublisher, cacheService, idGenerator)
	return uc.(*pipelineUseCase), pipelineRepo, oppRepo
}

// ============================================================================
// PipelineUseCase Tests - Create
// ============================================================================

func TestPipelineUseCase_Create_Success(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Sales Pipeline",
		DefaultCurrency: "USD",
		IsDefault:       false,
		Stages: []*dto.CreateStageRequest{
			{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
			{Name: "Proposal", Type: "open", Order: 2, Probability: 40, Color: "#9b59b6"},
			{Name: "Negotiation", Type: "negotiating", Order: 3, Probability: 70, Color: "#f39c12"},
			{Name: "Won", Type: "won", Order: 4, Probability: 100, Color: "#27ae60"},
			{Name: "Lost", Type: "lost", Order: 5, Probability: 0, Color: "#e74c3c"},
		},
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
	if result.Name != "Sales Pipeline" {
		t.Errorf("Expected name 'Sales Pipeline', got '%s'", result.Name)
	}
	if len(result.Stages) != 5 {
		t.Errorf("Expected 5 stages, got %d", len(result.Stages))
	}
}

func TestPipelineUseCase_Create_WithDefaultStages(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Default Pipeline",
		DefaultCurrency: "USD",
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
	if len(result.Stages) == 0 {
		t.Error("Expected default stages to be created")
	}
}

func TestPipelineUseCase_Create_EmptyName(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "",
		DefaultCurrency: "USD",
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for empty name, got nil")
	}
}

func TestPipelineUseCase_Create_InvalidCurrency(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Test Pipeline",
		DefaultCurrency: "INVALID",
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid currency, got nil")
	}
}

func TestPipelineUseCase_Create_MissingWonStage(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Test Pipeline",
		DefaultCurrency: "USD",
		Stages: []*dto.CreateStageRequest{
			{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
			{Name: "Proposal", Type: "open", Order: 2, Probability: 40, Color: "#9b59b6"},
			{Name: "Lost", Type: "lost", Order: 3, Probability: 0, Color: "#e74c3c"},
		},
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing won stage, got nil")
	}
}

func TestPipelineUseCase_Create_MissingLostStage(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Test Pipeline",
		DefaultCurrency: "USD",
		Stages: []*dto.CreateStageRequest{
			{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
			{Name: "Proposal", Type: "open", Order: 2, Probability: 40, Color: "#9b59b6"},
			{Name: "Won", Type: "won", Order: 3, Probability: 100, Color: "#27ae60"},
		},
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for missing lost stage, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - GetByID
// ============================================================================

func TestPipelineUseCase_GetByID_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	result, err := uc.GetByID(context.Background(), tenantID, pipeline.ID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.ID != pipeline.ID.String() {
		t.Errorf("Expected ID %s, got %s", pipeline.ID.String(), result.ID)
	}
	if result.Name != pipeline.Name {
		t.Errorf("Expected name %s, got %s", pipeline.Name, result.Name)
	}
}

func TestPipelineUseCase_GetByID_NotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	// Act
	_, err := uc.GetByID(context.Background(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - Update
// ============================================================================

func TestPipelineUseCase_Update_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	newName := "Updated Pipeline"
	newDesc := "Updated description"
	req := &dto.UpdatePipelineRequest{
		Name:        &newName,
		Description: &newDesc,
		Version:     1,
	}

	// Act
	result, err := uc.Update(context.Background(), tenantID, pipeline.ID, userID, req)

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

func TestPipelineUseCase_Update_NotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	req := &dto.UpdatePipelineRequest{
		Version: 1,
	}

	// Act
	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

func TestPipelineUseCase_Update_VersionMismatch(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.Version = 5
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	req := &dto.UpdatePipelineRequest{
		Version: 3, // Outdated version
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, pipeline.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for version mismatch, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - Delete
// ============================================================================

func TestPipelineUseCase_Delete_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsDefault = false
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	err := uc.Delete(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestPipelineUseCase_Delete_NotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	// Act
	err := uc.Delete(context.Background(), uuid.New(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

func TestPipelineUseCase_Delete_DefaultPipeline(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsDefault = true
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	err := uc.Delete(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for deleting default pipeline, got nil")
	}
}

func TestPipelineUseCase_Delete_WithOpportunities(t *testing.T) {
	// Arrange
	uc, pipelineRepo, oppRepo := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsDefault = false
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Add opportunity to the pipeline
	amount, _ := domain.NewMoney(10000, "USD")
	opp := &domain.Opportunity{
		ID:         uuid.New(),
		TenantID:   tenantID,
		PipelineID: pipeline.ID,
		Name:       "Test Opportunity",
		Amount:     amount,
	}
	oppRepo.opportunities[opp.ID] = opp

	// Act
	err := uc.Delete(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for deleting pipeline with opportunities, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - List
// ============================================================================

func TestPipelineUseCase_List_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline1 := createPipelineForTest(tenantID)
	pipeline2 := createPipelineForTest(tenantID)
	pipeline2.Name = "Second Pipeline"
	pipelineRepo.pipelines[pipeline1.ID] = pipeline1
	pipelineRepo.pipelines[pipeline2.ID] = pipeline2

	filter := &dto.PipelineFilterRequest{
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
	if len(result.Pipelines) != 2 {
		t.Errorf("Expected 2 pipelines, got %d", len(result.Pipelines))
	}
}

func TestPipelineUseCase_List_EmptyResult(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	filter := &dto.PipelineFilterRequest{
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
	if len(result.Pipelines) != 0 {
		t.Errorf("Expected 0 pipelines, got %d", len(result.Pipelines))
	}
}

// ============================================================================
// PipelineUseCase Tests - GetDefault
// ============================================================================

func TestPipelineUseCase_GetDefault_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsDefault = true
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	result, err := uc.GetDefault(context.Background(), tenantID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if !result.IsDefault {
		t.Error("Expected pipeline to be default")
	}
}

func TestPipelineUseCase_GetDefault_NotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	// Act
	_, err := uc.GetDefault(context.Background(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for no default pipeline, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - SetDefault
// ============================================================================

func TestPipelineUseCase_SetDefault_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()

	// Create existing default pipeline
	existingDefault := createPipelineForTest(tenantID)
	existingDefault.IsDefault = true
	pipelineRepo.pipelines[existingDefault.ID] = existingDefault

	// Create pipeline to set as default
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsDefault = false
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	result, err := uc.SetDefault(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if !result.IsDefault {
		t.Error("Expected pipeline to be set as default")
	}
}

func TestPipelineUseCase_SetDefault_NotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	// Act
	_, err := uc.SetDefault(context.Background(), uuid.New(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

func TestPipelineUseCase_SetDefault_InactivePipeline(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsActive = false
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	_, err := uc.SetDefault(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for inactive pipeline, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - AddStage
// ============================================================================

func TestPipelineUseCase_AddStage_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	req := &dto.AddStageRequest{
		Name:        "Discovery",
		Type:        "open",
		Probability: 25,
		Color:       "#3498db",
	}

	// Act
	result, err := uc.AddStage(context.Background(), tenantID, pipeline.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineUseCase_AddStage_PipelineNotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	req := &dto.AddStageRequest{
		Name:        "Discovery",
		Type:        "open",
		Probability: 25,
		Color:       "#3498db",
	}

	// Act
	_, err := uc.AddStage(context.Background(), uuid.New(), uuid.New(), uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

func TestPipelineUseCase_AddStage_DuplicateName(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Use existing stage name
	req := &dto.AddStageRequest{
		Name:        "Qualification", // Already exists
		Type:        "open",
		Probability: 25,
		Color:       "#3498db",
	}

	// Act
	_, err := uc.AddStage(context.Background(), tenantID, pipeline.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for duplicate stage name, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - UpdateStage
// ============================================================================

func TestPipelineUseCase_UpdateStage_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	stageID := pipeline.Stages[0].ID
	newName := "Updated Stage"
	newProb := 50
	req := &dto.UpdateStageRequest{
		Name:        &newName,
		Probability: &newProb,
	}

	// Act
	result, err := uc.UpdateStage(context.Background(), tenantID, pipeline.ID, stageID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineUseCase_UpdateStage_StageNotFound(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	newName := "Updated Stage"
	req := &dto.UpdateStageRequest{
		Name: &newName,
	}

	// Act
	_, err := uc.UpdateStage(context.Background(), tenantID, pipeline.ID, uuid.New(), userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for stage not found, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - RemoveStage
// ============================================================================

func TestPipelineUseCase_RemoveStage_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Remove a non-closed stage (not Won or Lost)
	stageID := pipeline.Stages[1].ID // Proposal stage

	// Act
	result, err := uc.RemoveStage(context.Background(), tenantID, pipeline.ID, stageID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineUseCase_RemoveStage_PipelineNotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	// Act
	_, err := uc.RemoveStage(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

func TestPipelineUseCase_RemoveStage_WithOpportunities(t *testing.T) {
	// Arrange
	uc, pipelineRepo, oppRepo := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	stageID := pipeline.Stages[0].ID

	// Add opportunity to the stage
	amount, _ := domain.NewMoney(10000, "USD")
	opp := &domain.Opportunity{
		ID:         uuid.New(),
		TenantID:   tenantID,
		PipelineID: pipeline.ID,
		StageID:    stageID,
		Name:       "Test Opportunity",
		Amount:     amount,
	}
	oppRepo.opportunities[opp.ID] = opp

	// Act
	_, err := uc.RemoveStage(context.Background(), tenantID, pipeline.ID, stageID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for removing stage with opportunities, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - ReorderStages
// ============================================================================

func TestPipelineUseCase_ReorderStages_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Reorder stages
	req := &dto.ReorderStagesRequest{
		StageOrders: []dto.StageOrderDTO{
			{StageID: pipeline.Stages[1].ID.String(), Order: 1},
			{StageID: pipeline.Stages[0].ID.String(), Order: 2},
			{StageID: pipeline.Stages[2].ID.String(), Order: 3},
			{StageID: pipeline.Stages[3].ID.String(), Order: 4},
			{StageID: pipeline.Stages[4].ID.String(), Order: 5},
		},
	}

	// Act
	result, err := uc.ReorderStages(context.Background(), tenantID, pipeline.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineUseCase_ReorderStages_PipelineNotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	req := &dto.ReorderStagesRequest{
		StageOrders: []dto.StageOrderDTO{
			{StageID: uuid.New().String(), Order: 1},
		},
	}

	// Act
	_, err := uc.ReorderStages(context.Background(), uuid.New(), uuid.New(), uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - GetStatistics
// ============================================================================

func TestPipelineUseCase_GetStatistics_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	result, err := uc.GetStatistics(context.Background(), tenantID, pipeline.ID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.TotalOpportunities != 10 {
		t.Errorf("Expected 10 total opportunities, got %d", result.TotalOpportunities)
	}
}

func TestPipelineUseCase_GetStatistics_Error(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	pipelineRepo.getStatisticsErr = errors.New("statistics error")

	// Act
	_, err := uc.GetStatistics(context.Background(), tenantID, pipeline.ID)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - Activate/Deactivate
// ============================================================================

func TestPipelineUseCase_Activate_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsActive = false
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	result, err := uc.Activate(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if !result.IsActive {
		t.Error("Expected pipeline to be active")
	}
}

func TestPipelineUseCase_Deactivate_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsActive = true
	pipeline.IsDefault = false
	pipeline.OpportunityCount = 0
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	result, err := uc.Deactivate(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.IsActive {
		t.Error("Expected pipeline to be inactive")
	}
}

func TestPipelineUseCase_Deactivate_DefaultPipeline(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsDefault = true
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Act
	_, err := uc.Deactivate(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for deactivating default pipeline, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - Clone
// ============================================================================

func TestPipelineUseCase_Clone_Success(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	req := &dto.ClonePipelineRequest{
		SourcePipelineID:    pipeline.ID.String(),
		Name:                "Cloned Pipeline",
		IncludeCustomFields: true,
	}

	// Act
	result, err := uc.Clone(context.Background(), tenantID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Name != "Cloned Pipeline" {
		t.Errorf("Expected name 'Cloned Pipeline', got '%s'", result.Name)
	}
	if result.ID == pipeline.ID.String() {
		t.Error("Clone should have different ID from source")
	}
}

func TestPipelineUseCase_Clone_SourceNotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	req := &dto.ClonePipelineRequest{
		SourcePipelineID: uuid.New().String(),
		Name:             "Cloned Pipeline",
	}

	// Act
	_, err := uc.Clone(context.Background(), uuid.New(), uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for source pipeline not found, got nil")
	}
}

// ============================================================================
// PipelineUseCase Tests - Templates
// ============================================================================

func TestPipelineUseCase_GetTemplates_Success(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	// Act
	result, err := uc.GetTemplates(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if len(result) == 0 {
		t.Error("Expected at least one template")
	}
}

func TestPipelineUseCase_CreateFromTemplate_Success(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()

	req := &dto.CreateFromTemplateRequest{
		TemplateID: "b2b-sales",
		Name:       "My B2B Pipeline",
		IsDefault:  false,
	}

	// Act
	result, err := uc.CreateFromTemplate(context.Background(), tenantID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Name != "My B2B Pipeline" {
		t.Errorf("Expected name 'My B2B Pipeline', got '%s'", result.Name)
	}
}

func TestPipelineUseCase_CreateFromTemplate_TemplateNotFound(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	req := &dto.CreateFromTemplateRequest{
		TemplateID: "nonexistent-template",
		Name:       "My Pipeline",
	}

	// Act
	_, err := uc.CreateFromTemplate(context.Background(), uuid.New(), uuid.New(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for template not found, got nil")
	}
}

// ============================================================================
// Table-Driven Tests
// ============================================================================

func TestPipelineUseCase_Create_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		req       *dto.CreatePipelineRequest
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid pipeline with all stages",
			req: &dto.CreatePipelineRequest{
				Name:            "Valid Pipeline",
				DefaultCurrency: "USD",
				Stages: []*dto.CreateStageRequest{
					{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
					{Name: "Proposal", Type: "open", Order: 2, Probability: 40, Color: "#9b59b6"},
					{Name: "Won", Type: "won", Order: 3, Probability: 100, Color: "#27ae60"},
					{Name: "Lost", Type: "lost", Order: 4, Probability: 0, Color: "#e74c3c"},
				},
			},
			expectErr: false,
		},
		{
			name: "empty name",
			req: &dto.CreatePipelineRequest{
				Name:            "",
				DefaultCurrency: "USD",
			},
			expectErr: true,
			errMsg:    "name required",
		},
		{
			name: "invalid currency",
			req: &dto.CreatePipelineRequest{
				Name:            "Test Pipeline",
				DefaultCurrency: "INVALID",
			},
			expectErr: true,
			errMsg:    "invalid currency",
		},
		{
			name: "missing won stage",
			req: &dto.CreatePipelineRequest{
				Name:            "Test Pipeline",
				DefaultCurrency: "USD",
				Stages: []*dto.CreateStageRequest{
					{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
					{Name: "Lost", Type: "lost", Order: 2, Probability: 0, Color: "#e74c3c"},
				},
			},
			expectErr: true,
			errMsg:    "won stage required",
		},
		{
			name: "missing lost stage",
			req: &dto.CreatePipelineRequest{
				Name:            "Test Pipeline",
				DefaultCurrency: "USD",
				Stages: []*dto.CreateStageRequest{
					{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
					{Name: "Won", Type: "won", Order: 2, Probability: 100, Color: "#27ae60"},
				},
			},
			expectErr: true,
			errMsg:    "lost stage required",
		},
		{
			name: "insufficient stages",
			req: &dto.CreatePipelineRequest{
				Name:            "Test Pipeline",
				DefaultCurrency: "USD",
				Stages: []*dto.CreateStageRequest{
					{Name: "Won", Type: "won", Order: 1, Probability: 100, Color: "#27ae60"},
					{Name: "Lost", Type: "lost", Order: 2, Probability: 0, Color: "#e74c3c"},
				},
			},
			expectErr: true,
			errMsg:    "minimum stages required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, _, _ := setupPipelineUseCase()

			_, err := uc.Create(context.Background(), uuid.New(), uuid.New(), tt.req)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestPipelineUseCase_Update_ValidationCases(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(*PipelineMockRepo, uuid.UUID) uuid.UUID
		req         *dto.UpdatePipelineRequest
		expectErr   bool
	}{
		{
			name: "valid update",
			setupRepo: func(repo *PipelineMockRepo, tenantID uuid.UUID) uuid.UUID {
				pipeline := createPipelineForTest(tenantID)
				repo.pipelines[pipeline.ID] = pipeline
				return pipeline.ID
			},
			req: &dto.UpdatePipelineRequest{
				Version: 1,
			},
			expectErr: false,
		},
		{
			name: "version mismatch",
			setupRepo: func(repo *PipelineMockRepo, tenantID uuid.UUID) uuid.UUID {
				pipeline := createPipelineForTest(tenantID)
				pipeline.Version = 5
				repo.pipelines[pipeline.ID] = pipeline
				return pipeline.ID
			},
			req: &dto.UpdatePipelineRequest{
				Version: 1,
			},
			expectErr: true,
		},
		{
			name: "pipeline not found",
			setupRepo: func(repo *PipelineMockRepo, tenantID uuid.UUID) uuid.UUID {
				return uuid.New()
			},
			req: &dto.UpdatePipelineRequest{
				Version: 1,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, pipelineRepo, _ := setupPipelineUseCase()
			tenantID := uuid.New()
			pipelineID := tt.setupRepo(pipelineRepo, tenantID)

			_, err := uc.Update(context.Background(), tenantID, pipelineID, uuid.New(), tt.req)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestPipelineUseCase_AddStage_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		req       *dto.AddStageRequest
		expectErr bool
	}{
		{
			name: "valid stage",
			req: &dto.AddStageRequest{
				Name:        "New Stage",
				Type:        "open",
				Probability: 50,
				Color:       "#3498db",
			},
			expectErr: false,
		},
		{
			name: "duplicate name",
			req: &dto.AddStageRequest{
				Name:        "Qualification", // Exists in default stages
				Type:        "open",
				Probability: 50,
				Color:       "#3498db",
			},
			expectErr: true,
		},
		{
			name: "invalid stage type",
			req: &dto.AddStageRequest{
				Name:        "New Stage",
				Type:        "invalid",
				Probability: 50,
				Color:       "#3498db",
			},
			expectErr: true,
		},
		{
			name: "invalid probability low",
			req: &dto.AddStageRequest{
				Name:        "New Stage",
				Type:        "open",
				Probability: -10,
				Color:       "#3498db",
			},
			expectErr: true,
		},
		{
			name: "invalid probability high",
			req: &dto.AddStageRequest{
				Name:        "New Stage",
				Type:        "open",
				Probability: 150,
				Color:       "#3498db",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, pipelineRepo, _ := setupPipelineUseCase()
			tenantID := uuid.New()
			pipeline := createPipelineForTest(tenantID)
			pipelineRepo.pipelines[pipeline.ID] = pipeline

			_, err := uc.AddStage(context.Background(), tenantID, pipeline.ID, uuid.New(), tt.req)

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

func BenchmarkPipelineUseCase_Create(b *testing.B) {
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Benchmark Pipeline",
		DefaultCurrency: "USD",
		Stages: []*dto.CreateStageRequest{
			{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
			{Name: "Won", Type: "won", Order: 2, Probability: 100, Color: "#27ae60"},
			{Name: "Lost", Type: "lost", Order: 3, Probability: 0, Color: "#e74c3c"},
		},
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.Create(ctx, tenantID, userID, req)
	}
}

func BenchmarkPipelineUseCase_GetByID(b *testing.B) {
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetByID(ctx, tenantID, pipeline.ID)
	}
}

func BenchmarkPipelineUseCase_List(b *testing.B) {
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	for i := 0; i < 10; i++ {
		pipeline := createPipelineForTest(tenantID)
		pipelineRepo.pipelines[pipeline.ID] = pipeline
	}
	ctx := context.Background()
	filter := &dto.PipelineFilterRequest{PageSize: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.List(ctx, tenantID, filter)
	}
}

func BenchmarkPipelineUseCase_GetStatistics(b *testing.B) {
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetStatistics(ctx, tenantID, pipeline.ID)
	}
}

func BenchmarkPipelineUseCase_AddStage(b *testing.B) {
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pipeline := createPipelineForTest(tenantID)
		pipelineRepo.pipelines[pipeline.ID] = pipeline
		req := &dto.AddStageRequest{
			Name:        "Bench Stage",
			Type:        "open",
			Probability: 50,
			Color:       "#3498db",
		}
		b.StartTimer()
		_, _ = uc.AddStage(ctx, tenantID, pipeline.ID, userID, req)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestPipelineUseCase_GetByID_ContextTimeout(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - just ensure no panic
	_, _ = uc.GetByID(ctx, tenantID, pipeline.ID)
}

func TestPipelineUseCase_Create_ContextTimeout(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Test Pipeline",
		DefaultCurrency: "USD",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - just ensure no panic
	_, _ = uc.Create(ctx, tenantID, userID, req)
}

func TestPipelineUseCase_List_ContextTimeout(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	filter := &dto.PipelineFilterRequest{PageSize: 10}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - just ensure no panic
	_, _ = uc.List(ctx, tenantID, filter)
}

func TestPipelineUseCase_GetStatistics_ContextTimeout(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - just ensure no panic
	_, _ = uc.GetStatistics(ctx, tenantID, pipeline.ID)
}

func TestPipelineUseCase_Update_ContextCancelled(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	newName := "Updated"
	req := &dto.UpdatePipelineRequest{
		Name:    &newName,
		Version: 1,
	}

	// Act - just ensure no panic
	_, _ = uc.Update(ctx, tenantID, pipeline.ID, userID, req)
}

func TestPipelineUseCase_Delete_ContextCancelled(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsDefault = false
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act - just ensure no panic
	_ = uc.Delete(ctx, tenantID, pipeline.ID, userID)
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestPipelineUseCase_Create_WithCustomFields(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Pipeline with Custom Fields",
		DefaultCurrency: "USD",
		Stages: []*dto.CreateStageRequest{
			{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
			{Name: "Proposal", Type: "open", Order: 2, Probability: 40, Color: "#9b59b6"},
			{Name: "Won", Type: "won", Order: 3, Probability: 100, Color: "#27ae60"},
			{Name: "Lost", Type: "lost", Order: 4, Probability: 0, Color: "#e74c3c"},
		},
		CustomFieldsSchema: []*dto.CustomFieldSchemaDTO{
			{
				Name:     "budget",
				Label:    "Budget",
				Type:     "currency",
				Required: true,
			},
			{
				Name:     "industry",
				Label:    "Industry",
				Type:     "select",
				Required: false,
				Options: []dto.CustomFieldOptionDTO{
					{Value: "tech", Label: "Technology"},
					{Value: "finance", Label: "Finance"},
				},
			},
		},
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
	if len(result.CustomFieldsSchema) != 2 {
		t.Errorf("Expected 2 custom fields, got %d", len(result.CustomFieldsSchema))
	}
}

func TestPipelineUseCase_Create_WithWinLossReasons(t *testing.T) {
	// Arrange
	uc, _, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Pipeline with Reasons",
		DefaultCurrency: "USD",
		Stages: []*dto.CreateStageRequest{
			{Name: "Lead", Type: "qualifying", Order: 1, Probability: 10, Color: "#3498db"},
			{Name: "Proposal", Type: "open", Order: 2, Probability: 40, Color: "#9b59b6"},
			{Name: "Won", Type: "won", Order: 3, Probability: 100, Color: "#27ae60"},
			{Name: "Lost", Type: "lost", Order: 4, Probability: 0, Color: "#e74c3c"},
		},
		WinReasons:  []string{"Best Price", "Great Product", "Excellent Support"},
		LossReasons: []string{"Too Expensive", "Missing Features", "Competitor Won"},
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
	if len(result.WinReasons) != 3 {
		t.Errorf("Expected 3 win reasons, got %d", len(result.WinReasons))
	}
	if len(result.LossReasons) != 3 {
		t.Errorf("Expected 3 loss reasons, got %d", len(result.LossReasons))
	}
}

func TestPipelineUseCase_UpdateStage_WithAutoActions(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	stageID := pipeline.Stages[0].ID
	delayHours := 24
	req := &dto.UpdateStageRequest{
		AutoActions: []*dto.StageAutoActionDTO{
			{
				Type:       "send_email",
				Trigger:    "on_enter",
				Config:     map[string]interface{}{"template": "welcome"},
				DelayHours: &delayHours,
			},
		},
	}

	// Act
	result, err := uc.UpdateStage(context.Background(), tenantID, pipeline.ID, stageID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestPipelineUseCase_Clone_WithoutCustomFields(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.CustomFields = []domain.CustomFieldDef{
		{Name: "budget", Type: "currency", Label: "Budget"},
	}
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	req := &dto.ClonePipelineRequest{
		SourcePipelineID:    pipeline.ID.String(),
		Name:                "Cloned Without Custom Fields",
		IncludeCustomFields: false,
	}

	// Act
	result, err := uc.Clone(context.Background(), tenantID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if len(result.CustomFieldsSchema) != 0 {
		t.Errorf("Expected no custom fields in clone, got %d", len(result.CustomFieldsSchema))
	}
}

func TestPipelineUseCase_List_WithOpportunityCounts(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	filter := &dto.PipelineFilterRequest{
		PageSize:                 10,
		IncludeOpportunityCounts: true,
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
// Repository Error Tests
// ============================================================================

func TestPipelineUseCase_Create_RepositoryError(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()
	pipelineRepo.createErr = errors.New("database error")

	tenantID := uuid.New()
	userID := uuid.New()
	req := &dto.CreatePipelineRequest{
		Name:            "Test Pipeline",
		DefaultCurrency: "USD",
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository error, got nil")
	}
}

func TestPipelineUseCase_Update_RepositoryError(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	pipelineRepo.updateErr = errors.New("database error")

	newName := "Updated"
	req := &dto.UpdatePipelineRequest{
		Name:    &newName,
		Version: 1,
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, pipeline.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository error, got nil")
	}
}

func TestPipelineUseCase_Delete_RepositoryError(t *testing.T) {
	// Arrange
	uc, pipelineRepo, _ := setupPipelineUseCase()

	tenantID := uuid.New()
	userID := uuid.New()
	pipeline := createPipelineForTest(tenantID)
	pipeline.IsDefault = false
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	pipelineRepo.deleteErr = errors.New("database error")

	// Act
	err := uc.Delete(context.Background(), tenantID, pipeline.ID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for repository error, got nil")
	}
}
