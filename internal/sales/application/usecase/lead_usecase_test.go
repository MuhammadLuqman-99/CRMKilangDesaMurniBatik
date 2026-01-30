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
// Mock Implementations for Lead Use Case
// ============================================================================

// MockLeadRepository is a mock implementation of domain.LeadRepository.
type MockLeadRepository struct {
	leads           map[uuid.UUID]*domain.Lead
	createErr       error
	updateErr       error
	deleteErr       error
	getByIDErr      error
	listErr         error
	countByStatusErr error
	countBySourceErr error
	conversionRateErr error
}

func NewMockLeadRepository() *MockLeadRepository {
	return &MockLeadRepository{
		leads: make(map[uuid.UUID]*domain.Lead),
	}
}

func (m *MockLeadRepository) Create(ctx context.Context, lead *domain.Lead) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.leads[lead.ID] = lead
	return nil
}

func (m *MockLeadRepository) GetByID(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.Lead, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	lead, ok := m.leads[leadID]
	if !ok {
		return nil, errors.New("lead not found")
	}
	if lead.TenantID != tenantID {
		return nil, errors.New("lead not found")
	}
	return lead, nil
}

func (m *MockLeadRepository) Update(ctx context.Context, lead *domain.Lead) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.leads[lead.ID] = lead
	return nil
}

func (m *MockLeadRepository) Delete(ctx context.Context, tenantID, leadID uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.leads, leadID)
	return nil
}

func (m *MockLeadRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.LeadFilter, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var result []*domain.Lead
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.DeletedAt == nil {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.Lead, error) {
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.Contact.Email == email {
			return lead, nil
		}
	}
	return nil, errors.New("lead not found")
}

func (m *MockLeadRepository) GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*domain.Lead, error) {
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.Contact.Phone == phone {
			return lead, nil
		}
	}
	return nil, errors.New("lead not found")
}

func (m *MockLeadRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.LeadStatus, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	var result []*domain.Lead
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.Status == status {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) GetQualifiedLeads(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return m.GetByStatus(ctx, tenantID, domain.LeadStatusQualified, opts)
}

func (m *MockLeadRepository) GetUnassignedLeads(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	var result []*domain.Lead
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.OwnerID == nil {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	var result []*domain.Lead
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.OwnerID != nil && *lead.OwnerID == ownerID {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) GetBySource(ctx context.Context, tenantID uuid.UUID, source domain.LeadSource, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	var result []*domain.Lead
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.Source == source {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) GetHighScoreLeads(ctx context.Context, tenantID uuid.UUID, minScore int, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	var result []*domain.Lead
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.Score.Score >= minScore {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) GetLeadsForNurturing(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return m.GetByStatus(ctx, tenantID, domain.LeadStatusNurturing, opts)
}

func (m *MockLeadRepository) GetCreatedBetween(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	var result []*domain.Lead
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.CreatedAt.After(start) && lead.CreatedAt.Before(end) {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) GetUpdatedSince(ctx context.Context, tenantID uuid.UUID, since time.Time, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	var result []*domain.Lead
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.UpdatedAt.After(since) {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) GetStaleLeads(ctx context.Context, tenantID uuid.UUID, staleDays int, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	var result []*domain.Lead
	cutoff := time.Now().AddDate(0, 0, -staleDays)
	for _, lead := range m.leads {
		if lead.TenantID == tenantID && lead.UpdatedAt.Before(cutoff) {
			result = append(result, lead)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockLeadRepository) BulkCreate(ctx context.Context, leads []*domain.Lead) error {
	for _, lead := range leads {
		m.leads[lead.ID] = lead
	}
	return nil
}

func (m *MockLeadRepository) BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, leadIDs []uuid.UUID, newOwnerID uuid.UUID) error {
	for _, id := range leadIDs {
		if lead, ok := m.leads[id]; ok && lead.TenantID == tenantID {
			lead.OwnerID = &newOwnerID
		}
	}
	return nil
}

func (m *MockLeadRepository) BulkUpdateStatus(ctx context.Context, tenantID uuid.UUID, leadIDs []uuid.UUID, status domain.LeadStatus) error {
	for _, id := range leadIDs {
		if lead, ok := m.leads[id]; ok && lead.TenantID == tenantID {
			lead.Status = status
		}
	}
	return nil
}

func (m *MockLeadRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.LeadStatus]int64, error) {
	if m.countByStatusErr != nil {
		return nil, m.countByStatusErr
	}
	counts := make(map[domain.LeadStatus]int64)
	for _, lead := range m.leads {
		if lead.TenantID == tenantID {
			counts[lead.Status]++
		}
	}
	return counts, nil
}

func (m *MockLeadRepository) CountBySource(ctx context.Context, tenantID uuid.UUID) (map[domain.LeadSource]int64, error) {
	if m.countBySourceErr != nil {
		return nil, m.countBySourceErr
	}
	counts := make(map[domain.LeadSource]int64)
	for _, lead := range m.leads {
		if lead.TenantID == tenantID {
			counts[lead.Source]++
		}
	}
	return counts, nil
}

func (m *MockLeadRepository) GetConversionRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	if m.conversionRateErr != nil {
		return 0, m.conversionRateErr
	}
	var total, converted int64
	for _, lead := range m.leads {
		if lead.TenantID == tenantID {
			total++
			if lead.Status == domain.LeadStatusConverted {
				converted++
			}
		}
	}
	if total == 0 {
		return 0, nil
	}
	return float64(converted) / float64(total) * 100, nil
}

// MockPipelineRepository is a mock implementation of domain.PipelineRepository.
type MockPipelineRepository struct {
	pipelines  map[uuid.UUID]*domain.Pipeline
	getByIDErr error
}

func NewMockPipelineRepository() *MockPipelineRepository {
	return &MockPipelineRepository{
		pipelines: make(map[uuid.UUID]*domain.Pipeline),
	}
}

func (m *MockPipelineRepository) Create(ctx context.Context, pipeline *domain.Pipeline) error {
	m.pipelines[pipeline.ID] = pipeline
	return nil
}

func (m *MockPipelineRepository) GetByID(ctx context.Context, tenantID, pipelineID uuid.UUID) (*domain.Pipeline, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return nil, errors.New("pipeline not found")
	}
	if pipeline.TenantID != tenantID {
		return nil, errors.New("pipeline not found")
	}
	return pipeline, nil
}

func (m *MockPipelineRepository) Update(ctx context.Context, pipeline *domain.Pipeline) error {
	m.pipelines[pipeline.ID] = pipeline
	return nil
}

func (m *MockPipelineRepository) Delete(ctx context.Context, tenantID, pipelineID uuid.UUID) error {
	delete(m.pipelines, pipelineID)
	return nil
}

func (m *MockPipelineRepository) List(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Pipeline, int64, error) {
	var result []*domain.Pipeline
	for _, p := range m.pipelines {
		if p.TenantID == tenantID {
			result = append(result, p)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockPipelineRepository) GetActivePipelines(ctx context.Context, tenantID uuid.UUID) ([]*domain.Pipeline, error) {
	var result []*domain.Pipeline
	for _, p := range m.pipelines {
		if p.TenantID == tenantID && p.IsActive {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockPipelineRepository) GetDefaultPipeline(ctx context.Context, tenantID uuid.UUID) (*domain.Pipeline, error) {
	for _, p := range m.pipelines {
		if p.TenantID == tenantID && p.IsDefault {
			return p, nil
		}
	}
	return nil, errors.New("no default pipeline")
}

func (m *MockPipelineRepository) GetStageByID(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*domain.Stage, error) {
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return nil, errors.New("pipeline not found")
	}
	return pipeline.GetStage(stageID), nil
}

func (m *MockPipelineRepository) AddStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *domain.Stage) error {
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return errors.New("pipeline not found")
	}
	pipeline.Stages = append(pipeline.Stages, stage)
	return nil
}

func (m *MockPipelineRepository) UpdateStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *domain.Stage) error {
	return nil
}

func (m *MockPipelineRepository) RemoveStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) error {
	return nil
}

func (m *MockPipelineRepository) ReorderStages(ctx context.Context, tenantID, pipelineID uuid.UUID, stageIDs []uuid.UUID) error {
	return nil
}

func (m *MockPipelineRepository) GetPipelineStatistics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*domain.PipelineStatistics, error) {
	return &domain.PipelineStatistics{}, nil
}

func (m *MockPipelineRepository) GetStageStatistics(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*domain.StageStatistics, error) {
	return &domain.StageStatistics{}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTestLead(tenantID uuid.UUID) *domain.Lead {
	contact := domain.LeadContact{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "+1234567890",
		JobTitle:  "CEO",
	}
	company := domain.LeadCompany{
		Name:     "Test Company",
		Industry: "Technology",
		Website:  "https://testcompany.com",
	}
	lead, _ := domain.NewLead(tenantID, contact, company, domain.LeadSourceWebsite, uuid.New())
	lead.Code = "LD-001"
	return lead
}

func createTestPipeline(tenantID uuid.UUID) *domain.Pipeline {
	pipeline, _ := domain.NewPipeline(tenantID, "Test Pipeline", "USD", uuid.New())
	pipeline.IsDefault = true
	pipeline.EnsureClosedStages()
	return pipeline
}

func setupLeadUseCase() (*leadUseCase, *MockLeadRepository, *MockOpportunityRepository, *MockPipelineRepository, *MockCustomerService, *MockUserService) {
	leadRepo := NewMockLeadRepository()
	oppRepo := NewMockOpportunityRepository()
	pipelineRepo := NewMockPipelineRepository()
	eventPublisher := NewMockSalesEventPublisher()
	customerService := NewMockCustomerService()
	userService := NewMockUserService()
	cacheService := NewMockSalesCacheService()
	searchService := NewMockSearchService()
	idGenerator := NewMockSalesIDGenerator()

	uc := NewLeadUseCase(
		leadRepo,
		oppRepo,
		pipelineRepo,
		eventPublisher,
		customerService,
		userService,
		cacheService,
		searchService,
		idGenerator,
	)

	return uc.(*leadUseCase), leadRepo, oppRepo, pipelineRepo, customerService, userService
}

// ============================================================================
// LeadUseCase Tests - Create
// ============================================================================

func TestLeadUseCase_Create_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, userService := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	// Add user for owner assignment
	userService.users[userID] = &ports.UserInfo{
		ID:       userID,
		FullName: "Test User",
		Email:    "test@example.com",
	}

	ownerIDStr := userID.String()
	company := "Test Company"
	req := &dto.CreateLeadRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
		Company:   &company,
		Source:    "website",
		OwnerID:   &ownerIDStr,
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
	if result.FirstName != req.FirstName {
		t.Errorf("Expected FirstName %s, got %s", req.FirstName, result.FirstName)
	}
	if result.Email != req.Email {
		t.Errorf("Expected Email %s, got %s", req.Email, result.Email)
	}
	if len(leadRepo.leads) != 1 {
		t.Errorf("Expected 1 lead in repository, got %d", len(leadRepo.leads))
	}
}

func TestLeadUseCase_Create_ValidationError_MissingFirstName(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	company := "Test Company"
	req := &dto.CreateLeadRequest{
		FirstName: "", // Empty - should fail validation
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
		Company:   &company,
		Source:    "website",
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
}

func TestLeadUseCase_Create_ValidationError_MissingEmail(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	company := "Test Company"
	req := &dto.CreateLeadRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "", // Empty - should fail validation
		Company:   &company,
		Source:    "website",
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
}

func TestLeadUseCase_Create_RepositoryError(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	leadRepo.createErr = errors.New("database error")

	company := "Test Company"
	req := &dto.CreateLeadRequest{
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
		Company:   &company,
		Source:    "website",
	}

	// Act
	_, err := uc.Create(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - GetByID
// ============================================================================

func TestLeadUseCase_GetByID_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	// Act
	result, err := uc.GetByID(context.Background(), tenantID, lead.ID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.ID != lead.ID.String() {
		t.Errorf("Expected ID %s, got %s", lead.ID.String(), result.ID)
	}
}

func TestLeadUseCase_GetByID_NotFound(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	nonExistentID := uuid.New()

	// Act
	_, err := uc.GetByID(context.Background(), tenantID, nonExistentID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for lead not found, got nil")
	}
}

func TestLeadUseCase_GetByID_WrongTenant(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	wrongTenantID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	// Act
	_, err := uc.GetByID(context.Background(), wrongTenantID, lead.ID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for wrong tenant, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - Update
// ============================================================================

func TestLeadUseCase_Update_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	newFirstName := "Updated"
	req := &dto.UpdateLeadRequest{
		FirstName: &newFirstName,
		Version:   1,
	}

	// Act
	result, err := uc.Update(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.FirstName != newFirstName {
		t.Errorf("Expected FirstName %s, got %s", newFirstName, result.FirstName)
	}
}

func TestLeadUseCase_Update_LeadNotFound(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	nonExistentID := uuid.New()

	req := &dto.UpdateLeadRequest{
		Version: 1,
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, nonExistentID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for lead not found, got nil")
	}
}

func TestLeadUseCase_Update_VersionMismatch(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Version = 5
	leadRepo.leads[lead.ID] = lead

	req := &dto.UpdateLeadRequest{
		Version: 3, // Outdated version
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for version mismatch, got nil")
	}
}

func TestLeadUseCase_Update_ConvertedLead(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusConverted
	leadRepo.leads[lead.ID] = lead

	req := &dto.UpdateLeadRequest{
		Version: 1,
	}

	// Act
	_, err := uc.Update(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for converted lead, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - Delete
// ============================================================================

func TestLeadUseCase_Delete_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	// Act
	err := uc.Delete(context.Background(), tenantID, lead.ID, userID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestLeadUseCase_Delete_LeadNotFound(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	nonExistentID := uuid.New()

	// Act
	err := uc.Delete(context.Background(), tenantID, nonExistentID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for lead not found, got nil")
	}
}

func TestLeadUseCase_Delete_ConvertedLead(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusConverted
	leadRepo.leads[lead.ID] = lead

	// Act
	err := uc.Delete(context.Background(), tenantID, lead.ID, userID)

	// Assert
	if err == nil {
		t.Fatal("Expected error for converted lead, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - List
// ============================================================================

func TestLeadUseCase_List_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	lead1 := createTestLead(tenantID)
	lead2 := createTestLead(tenantID)
	lead2.Contact.Email = "another@example.com"
	leadRepo.leads[lead1.ID] = lead1
	leadRepo.leads[lead2.ID] = lead2

	filter := &dto.LeadFilterRequest{
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
	if len(result.Leads) != 2 {
		t.Errorf("Expected 2 leads, got %d", len(result.Leads))
	}
}

func TestLeadUseCase_List_WithStatusFilter(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	lead1 := createTestLead(tenantID)
	lead1.Status = domain.LeadStatusNew
	lead2 := createTestLead(tenantID)
	lead2.Status = domain.LeadStatusQualified
	leadRepo.leads[lead1.ID] = lead1
	leadRepo.leads[lead2.ID] = lead2

	filter := &dto.LeadFilterRequest{
		Statuses: []string{"new"},
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

func TestLeadUseCase_List_EmptyResult(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	filter := &dto.LeadFilterRequest{
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
	if len(result.Leads) != 0 {
		t.Errorf("Expected 0 leads, got %d", len(result.Leads))
	}
}

// ============================================================================
// LeadUseCase Tests - Qualify
// ============================================================================

func TestLeadUseCase_Qualify_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusNew
	leadRepo.leads[lead.ID] = lead

	req := &dto.QualifyLeadRequest{
		Notes: "Lead is qualified",
	}

	// Act
	result, err := uc.Qualify(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Status != string(domain.LeadStatusQualified) {
		t.Errorf("Expected status %s, got %s", domain.LeadStatusQualified, result.Status)
	}
}

func TestLeadUseCase_Qualify_AlreadyQualified(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusQualified
	leadRepo.leads[lead.ID] = lead

	req := &dto.QualifyLeadRequest{}

	// Act
	_, err := uc.Qualify(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already qualified lead, got nil")
	}
}

func TestLeadUseCase_Qualify_ConvertedLead(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusConverted
	leadRepo.leads[lead.ID] = lead

	req := &dto.QualifyLeadRequest{}

	// Act
	_, err := uc.Qualify(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for converted lead, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - Disqualify
// ============================================================================

func TestLeadUseCase_Disqualify_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusNew
	leadRepo.leads[lead.ID] = lead

	req := &dto.DisqualifyLeadRequest{
		Reason: "no_budget",
		Notes:  "Customer has no budget",
	}

	// Act
	result, err := uc.Disqualify(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Status != string(domain.LeadStatusUnqualified) {
		t.Errorf("Expected status %s, got %s", domain.LeadStatusUnqualified, result.Status)
	}
}

func TestLeadUseCase_Disqualify_AlreadyDisqualified(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusUnqualified
	leadRepo.leads[lead.ID] = lead

	req := &dto.DisqualifyLeadRequest{
		Reason: "no_budget",
	}

	// Act
	_, err := uc.Disqualify(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already disqualified lead, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - Convert
// ============================================================================

func TestLeadUseCase_Convert_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, pipelineRepo, customerService, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	// Create qualified lead
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusQualified
	leadRepo.leads[lead.ID] = lead

	// Create pipeline
	pipeline := createTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Create customer
	customerID := uuid.New()
	customerService.customers[customerID] = &ports.CustomerInfo{
		ID:   customerID,
		Name: "Test Customer",
	}

	customerIDStr := customerID.String()
	req := &dto.ConvertLeadRequest{
		OpportunityName: "New Opportunity",
		PipelineID:      pipeline.ID.String(),
		CustomerID:      &customerIDStr,
	}

	// Act
	result, err := uc.Convert(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.LeadID != lead.ID.String() {
		t.Errorf("Expected LeadID %s, got %s", lead.ID.String(), result.LeadID)
	}
}

func TestLeadUseCase_Convert_NotQualified(t *testing.T) {
	// Arrange
	uc, leadRepo, _, pipelineRepo, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	// Create non-qualified lead
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusNew
	leadRepo.leads[lead.ID] = lead

	// Create pipeline
	pipeline := createTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	req := &dto.ConvertLeadRequest{
		OpportunityName: "New Opportunity",
		PipelineID:      pipeline.ID.String(),
	}

	// Act
	_, err := uc.Convert(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for non-qualified lead, got nil")
	}
}

func TestLeadUseCase_Convert_InvalidPipelineID(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	// Create qualified lead
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusQualified
	leadRepo.leads[lead.ID] = lead

	req := &dto.ConvertLeadRequest{
		OpportunityName: "New Opportunity",
		PipelineID:      "invalid-uuid",
	}

	// Act
	_, err := uc.Convert(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid pipeline ID, got nil")
	}
}

func TestLeadUseCase_Convert_PipelineNotFound(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	// Create qualified lead
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusQualified
	leadRepo.leads[lead.ID] = lead

	nonExistentPipelineID := uuid.New()
	req := &dto.ConvertLeadRequest{
		OpportunityName: "New Opportunity",
		PipelineID:      nonExistentPipelineID.String(),
	}

	// Act
	_, err := uc.Convert(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - Nurture
// ============================================================================

func TestLeadUseCase_Nurture_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusNew
	leadRepo.leads[lead.ID] = lead

	req := &dto.NurtureLeadRequest{
		Reason: "Not ready to buy yet",
	}

	// Act
	result, err := uc.Nurture(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Status != string(domain.LeadStatusNurturing) {
		t.Errorf("Expected status %s, got %s", domain.LeadStatusNurturing, result.Status)
	}
}

func TestLeadUseCase_Nurture_ConvertedLead(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusConverted
	leadRepo.leads[lead.ID] = lead

	req := &dto.NurtureLeadRequest{}

	// Act
	_, err := uc.Nurture(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for converted lead, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - Assign
// ============================================================================

func TestLeadUseCase_Assign_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, userService := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	userService.users[ownerID] = &ports.UserInfo{
		ID:       ownerID,
		FullName: "New Owner",
		Email:    "owner@example.com",
	}

	req := &dto.AssignLeadRequest{
		OwnerID: ownerID.String(),
	}

	// Act
	result, err := uc.Assign(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.OwnerID == nil || *result.OwnerID != ownerID.String() {
		t.Error("Expected owner to be assigned")
	}
}

func TestLeadUseCase_Assign_InvalidOwnerID(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	req := &dto.AssignLeadRequest{
		OwnerID: "invalid-uuid",
	}

	// Act
	_, err := uc.Assign(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid owner ID, got nil")
	}
}

func TestLeadUseCase_Assign_LeadNotFound(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()

	req := &dto.AssignLeadRequest{
		OwnerID: ownerID.String(),
	}

	// Act
	_, err := uc.Assign(context.Background(), tenantID, uuid.New(), userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for lead not found, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - BulkAssign
// ============================================================================

func TestLeadUseCase_BulkAssign_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, userService := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()

	lead1 := createTestLead(tenantID)
	lead2 := createTestLead(tenantID)
	leadRepo.leads[lead1.ID] = lead1
	leadRepo.leads[lead2.ID] = lead2

	userService.users[ownerID] = &ports.UserInfo{
		ID:       ownerID,
		FullName: "New Owner",
		Email:    "owner@example.com",
	}

	req := &dto.BulkAssignLeadsRequest{
		LeadIDs: []string{lead1.ID.String(), lead2.ID.String()},
		OwnerID: ownerID.String(),
	}

	// Act
	count, err := uc.BulkAssign(context.Background(), tenantID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 leads assigned, got %d", count)
	}
}

func TestLeadUseCase_BulkAssign_InvalidOwnerID(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	lead1 := createTestLead(tenantID)
	leadRepo.leads[lead1.ID] = lead1

	req := &dto.BulkAssignLeadsRequest{
		LeadIDs: []string{lead1.ID.String()},
		OwnerID: "invalid-uuid",
	}

	// Act
	_, err := uc.BulkAssign(context.Background(), tenantID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for invalid owner ID, got nil")
	}
}

func TestLeadUseCase_BulkAssign_PartialSuccess(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, userService := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	ownerID := uuid.New()

	lead1 := createTestLead(tenantID)
	leadRepo.leads[lead1.ID] = lead1

	userService.users[ownerID] = &ports.UserInfo{
		ID:       ownerID,
		FullName: "New Owner",
		Email:    "owner@example.com",
	}

	nonExistentID := uuid.New()
	req := &dto.BulkAssignLeadsRequest{
		LeadIDs: []string{lead1.ID.String(), nonExistentID.String()},
		OwnerID: ownerID.String(),
	}

	// Act
	count, err := uc.BulkAssign(context.Background(), tenantID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 lead assigned, got %d", count)
	}
}

// ============================================================================
// LeadUseCase Tests - UpdateScore
// ============================================================================

func TestLeadUseCase_UpdateScore_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	demographicScore := 50
	behavioralScore := 30
	req := &dto.ScoreLeadRequest{
		DemographicScore: &demographicScore,
		BehavioralScore:  &behavioralScore,
		Reason:           "Updated based on engagement",
	}

	// Act
	result, err := uc.UpdateScore(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Score != 80 {
		t.Errorf("Expected score 80, got %d", result.Score)
	}
}

func TestLeadUseCase_UpdateScore_LeadNotFound(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	nonExistentID := uuid.New()

	demographicScore := 50
	req := &dto.ScoreLeadRequest{
		DemographicScore: &demographicScore,
	}

	// Act
	_, err := uc.UpdateScore(context.Background(), tenantID, nonExistentID, userID, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for lead not found, got nil")
	}
}

// ============================================================================
// LeadUseCase Tests - GetStatistics
// ============================================================================

func TestLeadUseCase_GetStatistics_Success(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	lead1 := createTestLead(tenantID)
	lead1.Status = domain.LeadStatusNew
	lead1.Source = domain.LeadSourceWebsite
	lead2 := createTestLead(tenantID)
	lead2.Status = domain.LeadStatusQualified
	lead2.Source = domain.LeadSourceReferral
	lead3 := createTestLead(tenantID)
	lead3.Status = domain.LeadStatusConverted
	lead3.Source = domain.LeadSourceWebsite

	leadRepo.leads[lead1.ID] = lead1
	leadRepo.leads[lead2.ID] = lead2
	leadRepo.leads[lead3.ID] = lead3

	// Act
	result, err := uc.GetStatistics(context.Background(), tenantID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.TotalLeads != 3 {
		t.Errorf("Expected 3 total leads, got %d", result.TotalLeads)
	}
}

func TestLeadUseCase_GetStatistics_Empty(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	// Act
	result, err := uc.GetStatistics(context.Background(), tenantID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.TotalLeads != 0 {
		t.Errorf("Expected 0 total leads, got %d", result.TotalLeads)
	}
}

func TestLeadUseCase_GetStatistics_CountByStatusError(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	leadRepo.countByStatusErr = errors.New("database error")

	// Act
	_, err := uc.GetStatistics(context.Background(), tenantID)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// ============================================================================
// Table-Driven Tests - Validation
// ============================================================================

func TestLeadUseCase_Create_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		req       *dto.CreateLeadRequest
		expectErr bool
	}{
		{
			name: "valid lead",
			req: &dto.CreateLeadRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Company:   func() *string { s := "Test Co"; return &s }(),
				Source:    "website",
			},
			expectErr: false,
		},
		{
			name: "missing first name",
			req: &dto.CreateLeadRequest{
				FirstName: "",
				LastName:  "Doe",
				Email:     "john@example.com",
				Company:   func() *string { s := "Test Co"; return &s }(),
				Source:    "website",
			},
			expectErr: true,
		},
		{
			name: "missing email",
			req: &dto.CreateLeadRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
				Company:   func() *string { s := "Test Co"; return &s }(),
				Source:    "website",
			},
			expectErr: true,
		},
		{
			name: "missing company",
			req: &dto.CreateLeadRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Company:   nil,
				Source:    "website",
			},
			expectErr: true,
		},
		{
			name: "invalid source defaults to other",
			req: &dto.CreateLeadRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Company:   func() *string { s := "Test Co"; return &s }(),
				Source:    "invalid_source",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, _, _, _, _, _ := setupLeadUseCase()
			tenantID := uuid.New()
			userID := uuid.New()

			_, err := uc.Create(context.Background(), tenantID, userID, tt.req)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestLeadUseCase_GetByID_ValidationCases(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  uuid.UUID
		leadID    uuid.UUID
		setupRepo func(*MockLeadRepository)
		expectErr bool
	}{
		{
			name:     "valid lead",
			tenantID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			leadID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			setupRepo: func(repo *MockLeadRepository) {
				lead := createTestLead(uuid.MustParse("11111111-1111-1111-1111-111111111111"))
				lead.ID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
				repo.leads[lead.ID] = lead
			},
			expectErr: false,
		},
		{
			name:      "lead not found",
			tenantID:  uuid.New(),
			leadID:    uuid.New(),
			setupRepo: func(repo *MockLeadRepository) {},
			expectErr: true,
		},
		{
			name:     "wrong tenant",
			tenantID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			leadID:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			setupRepo: func(repo *MockLeadRepository) {
				lead := createTestLead(uuid.MustParse("11111111-1111-1111-1111-111111111111"))
				lead.ID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
				repo.leads[lead.ID] = lead
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, leadRepo, _, _, _, _ := setupLeadUseCase()
			tt.setupRepo(leadRepo)

			_, err := uc.GetByID(context.Background(), tt.tenantID, tt.leadID)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestLeadUseCase_StatusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus domain.LeadStatus
		operation     string
		expectErr     bool
	}{
		{"qualify new lead", domain.LeadStatusNew, "qualify", false},
		{"qualify contacted lead", domain.LeadStatusContacted, "qualify", false},
		{"qualify already qualified", domain.LeadStatusQualified, "qualify", true},
		{"qualify converted lead", domain.LeadStatusConverted, "qualify", true},
		{"disqualify new lead", domain.LeadStatusNew, "disqualify", false},
		{"disqualify qualified lead", domain.LeadStatusQualified, "disqualify", false},
		{"disqualify already disqualified", domain.LeadStatusUnqualified, "disqualify", true},
		{"disqualify converted lead", domain.LeadStatusConverted, "disqualify", true},
		{"nurture new lead", domain.LeadStatusNew, "nurture", false},
		{"nurture converted lead", domain.LeadStatusConverted, "nurture", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, leadRepo, _, _, _, _ := setupLeadUseCase()
			tenantID := uuid.New()
			userID := uuid.New()

			lead := createTestLead(tenantID)
			lead.Status = tt.initialStatus
			leadRepo.leads[lead.ID] = lead

			var err error
			switch tt.operation {
			case "qualify":
				_, err = uc.Qualify(context.Background(), tenantID, lead.ID, userID, &dto.QualifyLeadRequest{})
			case "disqualify":
				_, err = uc.Disqualify(context.Background(), tenantID, lead.ID, userID, &dto.DisqualifyLeadRequest{Reason: "no_budget"})
			case "nurture":
				_, err = uc.Nurture(context.Background(), tenantID, lead.ID, userID, &dto.NurtureLeadRequest{})
			}

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

func BenchmarkLeadUseCase_Create(b *testing.B) {
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	company := "Test Company"
	req := &dto.CreateLeadRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Company:   &company,
		Source:    "website",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.Create(ctx, tenantID, userID, req)
	}
}

func BenchmarkLeadUseCase_GetByID(b *testing.B) {
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetByID(ctx, tenantID, lead.ID)
	}
}

func BenchmarkLeadUseCase_List(b *testing.B) {
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	// Add multiple leads
	for i := 0; i < 100; i++ {
		lead := createTestLead(tenantID)
		leadRepo.leads[lead.ID] = lead
	}

	ctx := context.Background()
	filter := &dto.LeadFilterRequest{PageSize: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.List(ctx, tenantID, filter)
	}
}

func BenchmarkLeadUseCase_Update(b *testing.B) {
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead
	ctx := context.Background()

	newFirstName := "Updated"
	req := &dto.UpdateLeadRequest{
		FirstName: &newFirstName,
		Version:   1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset lead version for each iteration
		lead.Version = 1
		_, _ = uc.Update(ctx, tenantID, lead.ID, userID, req)
	}
}

func BenchmarkLeadUseCase_UpdateScore(b *testing.B) {
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead
	ctx := context.Background()

	demographicScore := 50
	behavioralScore := 30
	req := &dto.ScoreLeadRequest{
		DemographicScore: &demographicScore,
		BehavioralScore:  &behavioralScore,
		Reason:           "Benchmark test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.UpdateScore(ctx, tenantID, lead.ID, userID, req)
	}
}

func BenchmarkLeadUseCase_GetStatistics(b *testing.B) {
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	// Add multiple leads with various statuses
	for i := 0; i < 100; i++ {
		lead := createTestLead(tenantID)
		if i%4 == 0 {
			lead.Status = domain.LeadStatusNew
		} else if i%4 == 1 {
			lead.Status = domain.LeadStatusQualified
		} else if i%4 == 2 {
			lead.Status = domain.LeadStatusConverted
		} else {
			lead.Status = domain.LeadStatusNurturing
		}
		leadRepo.leads[lead.ID] = lead
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetStatistics(ctx, tenantID)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestLeadUseCase_GetByID_ContextTimeout(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - just ensure no panic with expired context
	_, _ = uc.GetByID(ctx, tenantID, lead.ID)
}

func TestLeadUseCase_Create_ContextTimeout(t *testing.T) {
	// Arrange
	uc, _, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	company := "Test Company"
	req := &dto.CreateLeadRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Company:   &company,
		Source:    "website",
	}

	// Act - just ensure no panic with expired context
	_, _ = uc.Create(ctx, tenantID, userID, req)
}

func TestLeadUseCase_List_ContextTimeout(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	for i := 0; i < 10; i++ {
		lead := createTestLead(tenantID)
		leadRepo.leads[lead.ID] = lead
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	filter := &dto.LeadFilterRequest{PageSize: 10}

	// Act - just ensure no panic with expired context
	_, _ = uc.List(ctx, tenantID, filter)
}

func TestLeadUseCase_Update_ContextTimeout(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	newFirstName := "Updated"
	req := &dto.UpdateLeadRequest{
		FirstName: &newFirstName,
		Version:   1,
	}

	// Act - just ensure no panic with expired context
	_, _ = uc.Update(ctx, tenantID, lead.ID, userID, req)
}

func TestLeadUseCase_Convert_ContextTimeout(t *testing.T) {
	// Arrange
	uc, leadRepo, _, pipelineRepo, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusQualified
	leadRepo.leads[lead.ID] = lead

	pipeline := createTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	req := &dto.ConvertLeadRequest{
		OpportunityName: "Test Opportunity",
		PipelineID:      pipeline.ID.String(),
	}

	// Act - just ensure no panic with expired context
	_, _ = uc.Convert(ctx, tenantID, lead.ID, userID, req)
}

func TestLeadUseCase_GetStatistics_ContextTimeout(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()

	for i := 0; i < 10; i++ {
		lead := createTestLead(tenantID)
		leadRepo.leads[lead.ID] = lead
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - just ensure no panic with expired context
	_, _ = uc.GetStatistics(ctx, tenantID)
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestLeadUseCase_ConcurrentReads(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	lead := createTestLead(tenantID)
	leadRepo.leads[lead.ID] = lead

	// Act - concurrent reads should not cause race conditions
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = uc.GetByID(context.Background(), tenantID, lead.ID)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLeadUseCase_ConcurrentWrites(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	// Create multiple leads
	for i := 0; i < 10; i++ {
		lead := createTestLead(tenantID)
		leadRepo.leads[lead.ID] = lead
	}

	// Act - concurrent score updates
	done := make(chan bool)
	for _, lead := range leadRepo.leads {
		go func(leadID uuid.UUID) {
			demographicScore := 50
			req := &dto.ScoreLeadRequest{
				DemographicScore: &demographicScore,
			}
			_, _ = uc.UpdateScore(context.Background(), tenantID, leadID, userID, req)
			done <- true
		}(lead.ID)
	}

	// Wait for all goroutines
	for range leadRepo.leads {
		<-done
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestLeadUseCase_Create_WithAllOptionalFields(t *testing.T) {
	// Arrange
	uc, _, _, _, _, userService := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	userService.users[userID] = &ports.UserInfo{
		ID:       userID,
		FullName: "Test User",
		Email:    "test@example.com",
	}

	company := "Test Company"
	phone := "+1234567890"
	mobile := "+0987654321"
	jobTitle := "CEO"
	department := "Executive"
	companySize := "51-200"
	industry := "Technology"
	website := "https://example.com"
	description := "Test lead description"
	budget := int64(100000)
	budgetCurrency := "USD"
	ownerIDStr := userID.String()

	req := &dto.CreateLeadRequest{
		FirstName:      "John",
		LastName:       "Doe",
		Email:          "john@example.com",
		Phone:          &phone,
		Mobile:         &mobile,
		JobTitle:       &jobTitle,
		Department:     &department,
		Company:        &company,
		CompanySize:    &companySize,
		Industry:       &industry,
		Website:        &website,
		Source:         "website",
		Description:    &description,
		Budget:         &budget,
		BudgetCurrency: &budgetCurrency,
		Tags:           []string{"vip", "enterprise"},
		CustomFields:   map[string]interface{}{"priority": "high"},
		OwnerID:        &ownerIDStr,
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
	if result.Phone == nil || *result.Phone != phone {
		t.Error("Expected phone to be set")
	}
	if result.JobTitle == nil || *result.JobTitle != jobTitle {
		t.Error("Expected job title to be set")
	}
	if len(result.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(result.Tags))
	}
}

func TestLeadUseCase_List_WithAllFilters(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	ownerID := uuid.New()

	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusQualified
	lead.Source = domain.LeadSourceWebsite
	lead.OwnerID = &ownerID
	lead.Score.Score = 75
	leadRepo.leads[lead.ID] = lead

	minScore := 50
	maxScore := 100
	filter := &dto.LeadFilterRequest{
		Statuses: []string{"qualified"},
		Sources:  []string{"website"},
		OwnerIDs: []string{ownerID.String()},
		MinScore: &minScore,
		MaxScore: &maxScore,
		Page:     1,
		PageSize: 10,
		SortBy:   "created_at",
		SortOrder: "desc",
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

func TestLeadUseCase_Nurture_WithCampaignAndReengageDate(t *testing.T) {
	// Arrange
	uc, leadRepo, _, _, _, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()
	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusNew
	leadRepo.leads[lead.ID] = lead

	campaignID := uuid.New().String()
	reengageDate := "2024-12-31"
	req := &dto.NurtureLeadRequest{
		NurtureCampaignID: &campaignID,
		Reason:            "Long-term nurturing",
		ReengageDate:      &reengageDate,
	}

	// Act
	result, err := uc.Nurture(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Status != string(domain.LeadStatusNurturing) {
		t.Errorf("Expected status %s, got %s", domain.LeadStatusNurturing, result.Status)
	}
}

func TestLeadUseCase_Convert_WithNewCustomer(t *testing.T) {
	// Arrange
	uc, leadRepo, _, pipelineRepo, customerService, _ := setupLeadUseCase()
	tenantID := uuid.New()
	userID := uuid.New()

	lead := createTestLead(tenantID)
	lead.Status = domain.LeadStatusQualified
	leadRepo.leads[lead.ID] = lead

	pipeline := createTestPipeline(tenantID)
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	// Mock customer creation
	customerService.customers = make(map[uuid.UUID]*ports.CustomerInfo)

	req := &dto.ConvertLeadRequest{
		OpportunityName:   "New Opportunity",
		PipelineID:        pipeline.ID.String(),
		CreateNewCustomer: true,
	}

	// Act
	result, err := uc.Convert(context.Background(), tenantID, lead.ID, userID, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}
