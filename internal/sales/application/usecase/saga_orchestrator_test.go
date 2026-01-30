package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application/ports"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Mock Implementations for Saga Orchestrator Testing
// ============================================================================

// MockSagaRepository is a mock implementation of domain.SagaRepository.
type MockSagaRepository struct {
	mu            sync.RWMutex
	sagas         map[uuid.UUID]*domain.LeadConversionSaga
	sagasByLead   map[uuid.UUID]*domain.LeadConversionSaga
	sagasByKey    map[string]*domain.LeadConversionSaga
	createErr     error
	updateErr     error
	getByIDErr    error
	getByLeadErr  error
	getByKeyErr   error
}

func NewMockSagaRepository() *MockSagaRepository {
	return &MockSagaRepository{
		sagas:       make(map[uuid.UUID]*domain.LeadConversionSaga),
		sagasByLead: make(map[uuid.UUID]*domain.LeadConversionSaga),
		sagasByKey:  make(map[string]*domain.LeadConversionSaga),
	}
}

func (m *MockSagaRepository) Create(ctx context.Context, saga *domain.LeadConversionSaga) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createErr != nil {
		return m.createErr
	}
	m.sagas[saga.ID] = saga
	m.sagasByLead[saga.LeadID] = saga
	m.sagasByKey[saga.IdempotencyKey] = saga
	return nil
}

func (m *MockSagaRepository) GetByID(ctx context.Context, tenantID, sagaID uuid.UUID) (*domain.LeadConversionSaga, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	saga, ok := m.sagas[sagaID]
	if !ok {
		return nil, domain.ErrSagaNotFound
	}
	return saga, nil
}

func (m *MockSagaRepository) GetByLeadID(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.LeadConversionSaga, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getByLeadErr != nil {
		return nil, m.getByLeadErr
	}
	saga, ok := m.sagasByLead[leadID]
	if !ok {
		return nil, domain.ErrSagaNotFound
	}
	return saga, nil
}

func (m *MockSagaRepository) GetByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*domain.LeadConversionSaga, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getByKeyErr != nil {
		return nil, m.getByKeyErr
	}
	saga, ok := m.sagasByKey[key]
	if !ok {
		return nil, domain.ErrSagaNotFound
	}
	return saga, nil
}

func (m *MockSagaRepository) Update(ctx context.Context, saga *domain.LeadConversionSaga) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateErr != nil {
		return m.updateErr
	}
	m.sagas[saga.ID] = saga
	m.sagasByLead[saga.LeadID] = saga
	m.sagasByKey[saga.IdempotencyKey] = saga
	return nil
}

func (m *MockSagaRepository) GetPendingSagas(ctx context.Context, olderThan time.Duration, limit int) ([]*domain.LeadConversionSaga, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*domain.LeadConversionSaga
	cutoff := time.Now().Add(-olderThan)
	for _, saga := range m.sagas {
		if !saga.State.IsTerminal() && saga.UpdatedAt.Before(cutoff) {
			result = append(result, saga)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockSagaRepository) GetByState(ctx context.Context, tenantID uuid.UUID, state domain.SagaState, opts domain.ListOptions) ([]*domain.LeadConversionSaga, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*domain.LeadConversionSaga
	for _, saga := range m.sagas {
		if saga.TenantID == tenantID && saga.State == state {
			result = append(result, saga)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockSagaRepository) GetCompensatingSagas(ctx context.Context, limit int) ([]*domain.LeadConversionSaga, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*domain.LeadConversionSaga
	for _, saga := range m.sagas {
		if saga.State == domain.SagaStateCompensating {
			result = append(result, saga)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockSagaRepository) GetFailedSagas(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.LeadConversionSaga, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*domain.LeadConversionSaga
	for _, saga := range m.sagas {
		if saga.TenantID == tenantID && saga.State == domain.SagaStateFailed {
			result = append(result, saga)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockSagaRepository) DeleteOldCompletedSagas(ctx context.Context, olderThan time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	var count int64
	for id, saga := range m.sagas {
		if saga.State == domain.SagaStateCompleted && saga.CompletedAt != nil && saga.CompletedAt.Before(cutoff) {
			delete(m.sagas, id)
			delete(m.sagasByLead, saga.LeadID)
			delete(m.sagasByKey, saga.IdempotencyKey)
			count++
		}
	}
	return count, nil
}

func (m *MockSagaRepository) CountByState(ctx context.Context, tenantID uuid.UUID) (map[domain.SagaState]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[domain.SagaState]int64)
	for _, saga := range m.sagas {
		if saga.TenantID == tenantID {
			result[saga.State]++
		}
	}
	return result, nil
}

// SagaMockLeadRepo is a mock implementation of domain.LeadRepository for saga tests.
type SagaMockLeadRepo struct {
	mu        sync.RWMutex
	leads     map[uuid.UUID]*domain.Lead
	getErr    error
	updateErr error
}

func NewSagaMockLeadRepo() *SagaMockLeadRepo {
	return &SagaMockLeadRepo{
		leads: make(map[uuid.UUID]*domain.Lead),
	}
}

func (m *SagaMockLeadRepo) Create(ctx context.Context, lead *domain.Lead) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.leads[lead.ID] = lead
	return nil
}

func (m *SagaMockLeadRepo) GetByID(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.Lead, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getErr != nil {
		return nil, m.getErr
	}
	lead, ok := m.leads[leadID]
	if !ok {
		return nil, domain.ErrLeadNotFound
	}
	return lead, nil
}

func (m *SagaMockLeadRepo) Update(ctx context.Context, lead *domain.Lead) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateErr != nil {
		return m.updateErr
	}
	m.leads[lead.ID] = lead
	return nil
}

func (m *SagaMockLeadRepo) Delete(ctx context.Context, tenantID, leadID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.leads, leadID)
	return nil
}

func (m *SagaMockLeadRepo) List(ctx context.Context, tenantID uuid.UUID, filter domain.LeadFilter, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.Lead, error) {
	return nil, domain.ErrLeadNotFound
}

func (m *SagaMockLeadRepo) GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*domain.Lead, error) {
	return nil, domain.ErrLeadNotFound
}

func (m *SagaMockLeadRepo) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.LeadStatus, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetQualifiedLeads(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetUnassignedLeads(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetBySource(ctx context.Context, tenantID uuid.UUID, source domain.LeadSource, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetHighScoreLeads(ctx context.Context, tenantID uuid.UUID, minScore int, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetLeadsForNurturing(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetCreatedBetween(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetUpdatedSince(ctx context.Context, tenantID uuid.UUID, since time.Time, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) GetStaleLeads(ctx context.Context, tenantID uuid.UUID, staleDays int, opts domain.ListOptions) ([]*domain.Lead, int64, error) {
	return []*domain.Lead{}, 0, nil
}

func (m *SagaMockLeadRepo) BulkCreate(ctx context.Context, leads []*domain.Lead) error {
	return nil
}

func (m *SagaMockLeadRepo) BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, leadIDs []uuid.UUID, newOwnerID uuid.UUID) error {
	return nil
}

func (m *SagaMockLeadRepo) BulkUpdateStatus(ctx context.Context, tenantID uuid.UUID, leadIDs []uuid.UUID, status domain.LeadStatus) error {
	return nil
}

func (m *SagaMockLeadRepo) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.LeadStatus]int64, error) {
	return map[domain.LeadStatus]int64{}, nil
}

func (m *SagaMockLeadRepo) CountBySource(ctx context.Context, tenantID uuid.UUID) (map[domain.LeadSource]int64, error) {
	return map[domain.LeadSource]int64{}, nil
}

func (m *SagaMockLeadRepo) GetConversionRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	return 0, nil
}

// MockSagaOpportunityRepository is a mock implementation of domain.OpportunityRepository for saga tests.
type MockSagaOpportunityRepository struct {
	mu            sync.RWMutex
	opportunities map[uuid.UUID]*domain.Opportunity
	createErr     error
	updateErr     error
	deleteErr     error
	getByIDErr    error
}

func NewMockSagaOpportunityRepository() *MockSagaOpportunityRepository {
	return &MockSagaOpportunityRepository{
		opportunities: make(map[uuid.UUID]*domain.Opportunity),
	}
}

func (m *MockSagaOpportunityRepository) Create(ctx context.Context, opp *domain.Opportunity) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createErr != nil {
		return m.createErr
	}
	m.opportunities[opp.ID] = opp
	return nil
}

func (m *MockSagaOpportunityRepository) GetByID(ctx context.Context, tenantID, opportunityID uuid.UUID) (*domain.Opportunity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	opp, ok := m.opportunities[opportunityID]
	if !ok {
		return nil, domain.ErrOpportunityNotFound
	}
	return opp, nil
}

func (m *MockSagaOpportunityRepository) Update(ctx context.Context, opp *domain.Opportunity) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateErr != nil {
		return m.updateErr
	}
	m.opportunities[opp.ID] = opp
	return nil
}

func (m *MockSagaOpportunityRepository) Delete(ctx context.Context, tenantID, opportunityID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.opportunities, opportunityID)
	return nil
}

func (m *MockSagaOpportunityRepository) List(ctx context.Context, tenantID uuid.UUID, filter domain.OpportunityFilter, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status domain.OpportunityStatus, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetOpenOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetWonOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetLostOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetByPipeline(ctx context.Context, tenantID, pipelineID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetByStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetByContact(ctx context.Context, tenantID, contactID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetByLead(ctx context.Context, tenantID, leadID uuid.UUID) (*domain.Opportunity, error) {
	return nil, domain.ErrOpportunityNotFound
}

func (m *MockSagaOpportunityRepository) GetByOwner(ctx context.Context, tenantID, ownerID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetClosingThisMonth(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetClosingThisQuarter(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetOverdueOpportunities(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetByExpectedCloseDate(ctx context.Context, tenantID uuid.UUID, start, end time.Time, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetHighValueOpportunities(ctx context.Context, tenantID uuid.UUID, minAmount int64, currency string, opts domain.ListOptions) ([]*domain.Opportunity, int64, error) {
	return []*domain.Opportunity{}, 0, nil
}

func (m *MockSagaOpportunityRepository) GetTotalPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	return 0, nil
}

func (m *MockSagaOpportunityRepository) GetWeightedPipelineValue(ctx context.Context, tenantID uuid.UUID, currency string) (int64, error) {
	return 0, nil
}

func (m *MockSagaOpportunityRepository) BulkUpdateOwner(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, newOwnerID uuid.UUID) error {
	return nil
}

func (m *MockSagaOpportunityRepository) BulkUpdateStage(ctx context.Context, tenantID uuid.UUID, opportunityIDs []uuid.UUID, stageID uuid.UUID) error {
	return nil
}

func (m *MockSagaOpportunityRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.OpportunityStatus]int64, error) {
	return map[domain.OpportunityStatus]int64{}, nil
}

func (m *MockSagaOpportunityRepository) CountByStage(ctx context.Context, tenantID, pipelineID uuid.UUID) (map[uuid.UUID]int64, error) {
	return map[uuid.UUID]int64{}, nil
}

func (m *MockSagaOpportunityRepository) GetWinRate(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (float64, error) {
	return 0, nil
}

func (m *MockSagaOpportunityRepository) GetAverageDealSize(ctx context.Context, tenantID uuid.UUID, currency string, start, end time.Time) (int64, error) {
	return 0, nil
}

func (m *MockSagaOpportunityRepository) GetAverageSalesCycle(ctx context.Context, tenantID uuid.UUID, start, end time.Time) (int, error) {
	return 0, nil
}

// MockSagaPipelineRepository is a mock implementation of domain.PipelineRepository for saga tests.
type MockSagaPipelineRepository struct {
	mu        sync.RWMutex
	pipelines map[uuid.UUID]*domain.Pipeline
	getErr    error
}

func NewMockSagaPipelineRepository() *MockSagaPipelineRepository {
	return &MockSagaPipelineRepository{
		pipelines: make(map[uuid.UUID]*domain.Pipeline),
	}
}

func (m *MockSagaPipelineRepository) Create(ctx context.Context, pipeline *domain.Pipeline) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pipelines[pipeline.ID] = pipeline
	return nil
}

func (m *MockSagaPipelineRepository) GetByID(ctx context.Context, tenantID, pipelineID uuid.UUID) (*domain.Pipeline, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getErr != nil {
		return nil, m.getErr
	}
	pipeline, ok := m.pipelines[pipelineID]
	if !ok {
		return nil, domain.ErrPipelineNotFound
	}
	return pipeline, nil
}

func (m *MockSagaPipelineRepository) Update(ctx context.Context, pipeline *domain.Pipeline) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pipelines[pipeline.ID] = pipeline
	return nil
}

func (m *MockSagaPipelineRepository) Delete(ctx context.Context, tenantID, pipelineID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.pipelines, pipelineID)
	return nil
}

func (m *MockSagaPipelineRepository) List(ctx context.Context, tenantID uuid.UUID, opts domain.ListOptions) ([]*domain.Pipeline, int64, error) {
	return []*domain.Pipeline{}, 0, nil
}

func (m *MockSagaPipelineRepository) GetActivePipelines(ctx context.Context, tenantID uuid.UUID) ([]*domain.Pipeline, error) {
	return []*domain.Pipeline{}, nil
}

func (m *MockSagaPipelineRepository) GetDefaultPipeline(ctx context.Context, tenantID uuid.UUID) (*domain.Pipeline, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.pipelines {
		if p.TenantID == tenantID && p.IsDefault {
			return p, nil
		}
	}
	return nil, domain.ErrPipelineNotFound
}

func (m *MockSagaPipelineRepository) GetStageByID(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*domain.Stage, error) {
	return nil, domain.ErrStageNotFound
}

func (m *MockSagaPipelineRepository) AddStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *domain.Stage) error {
	return nil
}

func (m *MockSagaPipelineRepository) UpdateStage(ctx context.Context, tenantID, pipelineID uuid.UUID, stage *domain.Stage) error {
	return nil
}

func (m *MockSagaPipelineRepository) RemoveStage(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) error {
	return nil
}

func (m *MockSagaPipelineRepository) ReorderStages(ctx context.Context, tenantID, pipelineID uuid.UUID, stageIDs []uuid.UUID) error {
	return nil
}

func (m *MockSagaPipelineRepository) GetPipelineStatistics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*domain.PipelineStatistics, error) {
	return &domain.PipelineStatistics{}, nil
}

func (m *MockSagaPipelineRepository) GetStageStatistics(ctx context.Context, tenantID, pipelineID, stageID uuid.UUID) (*domain.StageStatistics, error) {
	return &domain.StageStatistics{}, nil
}

// MockSagaUnitOfWork is a mock implementation of domain.UnitOfWork for saga tests.
type MockSagaUnitOfWork struct {
	leadRepo        *SagaMockLeadRepo
	opportunityRepo *MockSagaOpportunityRepository
	pipelineRepo    *MockSagaPipelineRepository
	sagaRepo        *MockSagaRepository
	beginErr        error
	commitErr       error
	rollbackErr     error
	inTransaction   bool
}

func NewMockSagaUnitOfWork(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, sagaRepo *MockSagaRepository) *MockSagaUnitOfWork {
	return &MockSagaUnitOfWork{
		leadRepo:        leadRepo,
		opportunityRepo: oppRepo,
		pipelineRepo:    pipelineRepo,
		sagaRepo:        sagaRepo,
	}
}

func (m *MockSagaUnitOfWork) Begin(ctx context.Context) (domain.UnitOfWork, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	m.inTransaction = true
	return m, nil
}

func (m *MockSagaUnitOfWork) Commit() error {
	if m.commitErr != nil {
		return m.commitErr
	}
	m.inTransaction = false
	return nil
}

func (m *MockSagaUnitOfWork) Rollback() error {
	if m.rollbackErr != nil {
		return m.rollbackErr
	}
	m.inTransaction = false
	return nil
}

func (m *MockSagaUnitOfWork) Leads() domain.LeadRepository {
	return m.leadRepo
}

func (m *MockSagaUnitOfWork) Opportunities() domain.OpportunityRepository {
	return m.opportunityRepo
}

func (m *MockSagaUnitOfWork) Deals() domain.DealRepository {
	return nil
}

func (m *MockSagaUnitOfWork) Pipelines() domain.PipelineRepository {
	return m.pipelineRepo
}

func (m *MockSagaUnitOfWork) Events() domain.EventStore {
	return nil
}

func (m *MockSagaUnitOfWork) Sagas() domain.SagaRepository {
	return m.sagaRepo
}

func (m *MockSagaUnitOfWork) IdempotencyKeys() domain.IdempotencyRepository {
	return nil
}

// MockSagaEventPublisher is a mock implementation of ports.EventPublisher for saga tests.
type MockSagaEventPublisher struct {
	mu         sync.Mutex
	events     []ports.Event
	publishErr error
}

func NewMockSagaEventPublisher() *MockSagaEventPublisher {
	return &MockSagaEventPublisher{
		events: make([]ports.Event, 0),
	}
}

func (m *MockSagaEventPublisher) Publish(ctx context.Context, event ports.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.publishErr != nil {
		return m.publishErr
	}
	m.events = append(m.events, event)
	return nil
}

func (m *MockSagaEventPublisher) PublishBatch(ctx context.Context, events []ports.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.publishErr != nil {
		return m.publishErr
	}
	m.events = append(m.events, events...)
	return nil
}

func (m *MockSagaEventPublisher) PublishAsync(ctx context.Context, event ports.Event) error {
	return m.Publish(ctx, event)
}

// MockSagaCustomerService is a mock implementation of ports.CustomerService for saga tests.
type MockSagaCustomerService struct {
	mu              sync.RWMutex
	customers       map[uuid.UUID]*ports.CustomerInfo
	createErr       error
	getErr          error
	createCustomerID uuid.UUID
}

func NewMockSagaCustomerService() *MockSagaCustomerService {
	return &MockSagaCustomerService{
		customers: make(map[uuid.UUID]*ports.CustomerInfo),
	}
}

func (m *MockSagaCustomerService) GetCustomer(ctx context.Context, tenantID, customerID uuid.UUID) (*ports.CustomerInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getErr != nil {
		return nil, m.getErr
	}
	customer, ok := m.customers[customerID]
	if !ok {
		return nil, errors.New("customer not found")
	}
	return customer, nil
}

func (m *MockSagaCustomerService) GetCustomerByCode(ctx context.Context, tenantID uuid.UUID, code string) (*ports.CustomerInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.customers {
		if c.Code == code {
			return c, nil
		}
	}
	return nil, errors.New("customer not found")
}

func (m *MockSagaCustomerService) CustomerExists(ctx context.Context, tenantID, customerID uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.customers[customerID]
	return ok, nil
}

func (m *MockSagaCustomerService) CreateCustomer(ctx context.Context, tenantID uuid.UUID, req ports.CreateCustomerRequest) (*ports.CustomerInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createErr != nil {
		return nil, m.createErr
	}

	customerID := m.createCustomerID
	if customerID == uuid.Nil {
		customerID = uuid.New()
	}

	customer := &ports.CustomerInfo{
		ID:       customerID,
		TenantID: tenantID,
		Code:     "CUST-" + customerID.String()[:8],
		Name:     req.Name,
		Type:     req.Type,
		Status:   "active",
		Email:    req.Email,
		Phone:    req.Phone,
	}
	m.customers[customerID] = customer
	return customer, nil
}

func (m *MockSagaCustomerService) GetContact(ctx context.Context, tenantID, contactID uuid.UUID) (*ports.ContactInfo, error) {
	return nil, errors.New("contact not found")
}

func (m *MockSagaCustomerService) ContactExists(ctx context.Context, tenantID, contactID uuid.UUID) (bool, error) {
	return false, nil
}

func (m *MockSagaCustomerService) CreateContact(ctx context.Context, tenantID, customerID uuid.UUID, req ports.CreateContactRequest) (*ports.ContactInfo, error) {
	return &ports.ContactInfo{
		ID:         uuid.New(),
		TenantID:   tenantID,
		CustomerID: customerID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
	}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTestQualifiedLead(tenantID uuid.UUID) *domain.Lead {
	ownerID := uuid.New()
	lead, _ := domain.NewLead(
		tenantID,
		domain.LeadContact{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "+1234567890",
		},
		domain.LeadCompany{
			Name:    "Acme Corporation",
			Website: "https://acme.com",
		},
		domain.LeadSourceWebsite,
		uuid.New(),
	)
	lead.AssignOwner(ownerID, "Test Owner")
	_ = lead.Qualify()
	return lead
}

func createSagaTestPipeline(tenantID uuid.UUID) *domain.Pipeline {
	pipeline, _ := domain.NewPipeline(tenantID, "Test Sales Pipeline", "USD", uuid.New())
	pipeline.EnsureClosedStages()
	return pipeline
}

func createTestConversionRequest(lead *domain.Lead, pipeline *domain.Pipeline, createCustomer bool) *domain.LeadConversionRequest {
	req := &domain.LeadConversionRequest{
		LeadID:            lead.ID,
		PipelineID:        pipeline.ID,
		CreateNewCustomer: createCustomer,
	}

	if !createCustomer {
		customerID := uuid.New()
		req.CustomerID = &customerID
	}

	return req
}

func createTestSaga(tenantID uuid.UUID, lead *domain.Lead, request *domain.LeadConversionRequest) *domain.LeadConversionSaga {
	idempotencyKey := domain.GenerateIdempotencyKey(tenantID, lead.ID, uuid.New())
	return domain.NewLeadConversionSaga(tenantID, lead.ID, idempotencyKey, uuid.New(), request)
}

func setupSagaOrchestratorTest() (*leadConversionOrchestrator, *SagaMockLeadRepo, *MockSagaOpportunityRepository, *MockSagaPipelineRepository, *MockSagaRepository, *MockSagaCustomerService) {
	leadRepo := NewSagaMockLeadRepo()
	oppRepo := NewMockSagaOpportunityRepository()
	pipelineRepo := NewMockSagaPipelineRepository()
	sagaRepo := NewMockSagaRepository()
	eventPublisher := NewMockSagaEventPublisher()
	customerService := NewMockSagaCustomerService()

	uow := NewMockSagaUnitOfWork(leadRepo, oppRepo, pipelineRepo, sagaRepo)

	orchestrator := NewLeadConversionOrchestrator(
		uow,
		leadRepo,
		oppRepo,
		pipelineRepo,
		eventPublisher,
		customerService,
	).(*leadConversionOrchestrator)

	return orchestrator, leadRepo, oppRepo, pipelineRepo, sagaRepo, customerService
}

// ============================================================================
// Test Cases - ConvertLeadToOpportunity Success Scenarios
// ============================================================================

func TestSagaOrchestrator_Execute_Success_WithNewCustomer(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if saga.State != domain.SagaStateCompleted {
		t.Errorf("Expected saga state to be completed, got: %s", saga.State)
	}

	if saga.OpportunityID == nil {
		t.Error("Expected opportunity ID to be set")
	}

	if saga.CustomerID == nil {
		t.Error("Expected customer ID to be set")
	}

	if !saga.CustomerCreated {
		t.Error("Expected CustomerCreated to be true")
	}

	// Verify customer was created
	if len(customerService.customers) == 0 {
		t.Error("Expected customer to be created in customer service")
	}
}

func TestSagaOrchestrator_Execute_Success_WithExistingCustomer(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	existingCustomerID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	customerService.customers[existingCustomerID] = &ports.CustomerInfo{
		ID:       existingCustomerID,
		TenantID: tenantID,
		Name:     "Existing Customer",
		Code:     "EXIST-001",
	}

	request := &domain.LeadConversionRequest{
		LeadID:            lead.ID,
		PipelineID:        pipeline.ID,
		CustomerID:        &existingCustomerID,
		CreateNewCustomer: false,
	}
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if saga.State != domain.SagaStateCompleted {
		t.Errorf("Expected saga state to be completed, got: %s", saga.State)
	}

	if saga.CustomerID == nil || *saga.CustomerID != existingCustomerID {
		t.Errorf("Expected customer ID to be %s, got: %v", existingCustomerID, saga.CustomerID)
	}

	if saga.CustomerCreated {
		t.Error("Expected CustomerCreated to be false for existing customer")
	}
}

// ============================================================================
// Test Cases - ConvertLeadToOpportunity Failure Scenarios
// ============================================================================

func TestSagaOrchestrator_Execute_Failure_LeadNotFound(t *testing.T) {
	// Arrange
	orchestrator, _, _, pipelineRepo, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	// Don't add lead to repo - simulating not found
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error for lead not found, got nil")
	}

	if saga.State != domain.SagaStateCompensated && saga.State != domain.SagaStateFailed {
		t.Errorf("Expected saga state to be compensated or failed, got: %s", saga.State)
	}
}

func TestSagaOrchestrator_Execute_Failure_LeadNotQualified(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	lead.Status = domain.LeadStatusNew // Not qualified
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error for lead not qualified, got nil")
	}
}

func TestSagaOrchestrator_Execute_Failure_PipelineNotFound(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, _, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	// Don't add pipeline to repo - simulating not found

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error for pipeline not found, got nil")
	}
}

func TestSagaOrchestrator_Execute_Failure_CustomerCreationFailed(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	customerService.createErr = errors.New("customer service unavailable")

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error for customer creation failure, got nil")
	}
}

func TestSagaOrchestrator_Execute_Failure_OpportunityCreationFailed(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, oppRepo, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	existingCustomerID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	customerService.customers[existingCustomerID] = &ports.CustomerInfo{
		ID:       existingCustomerID,
		TenantID: tenantID,
		Name:     "Existing Customer",
	}
	oppRepo.createErr = errors.New("database error")

	request := &domain.LeadConversionRequest{
		LeadID:            lead.ID,
		PipelineID:        pipeline.ID,
		CustomerID:        &existingCustomerID,
		CreateNewCustomer: false,
	}
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error for opportunity creation failure, got nil")
	}
}

// ============================================================================
// Test Cases - Saga Rollback and Compensation
// ============================================================================

func TestSagaOrchestrator_Compensate_Success(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, oppRepo, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)

	// Simulate a partially completed saga
	saga.Start()

	// Complete validate lead step
	validateStep := saga.GetStepByType(domain.StepTypeValidateLead)
	validateStep.Start()
	validateStep.Complete(map[string]interface{}{"lead_id": lead.ID.String()})

	// Complete create customer step
	customerStep := saga.GetStepByType(domain.StepTypeCreateCustomer)
	customerStep.Start()
	customerID := uuid.New()
	saga.SetCustomerID(customerID, true)
	customerService.customers[customerID] = &ports.CustomerInfo{ID: customerID, Name: "Test Customer"}
	customerStep.Complete(map[string]interface{}{"customer_id": customerID.String()})

	// Complete create opportunity step
	oppStep := saga.GetStepByType(domain.StepTypeCreateOpportunity)
	oppStep.Start()
	opportunityID := uuid.New()
	saga.SetOpportunityID(opportunityID)
	oppRepo.opportunities[opportunityID] = &domain.Opportunity{ID: opportunityID, TenantID: tenantID}
	oppStep.Complete(map[string]interface{}{"opportunity_id": opportunityID.String()})

	// Start compensation
	saga.StartCompensation(errors.New("simulated failure"), domain.StepTypeMarkLeadConverted)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Compensate(context.Background(), saga)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if saga.State != domain.SagaStateCompensated {
		t.Errorf("Expected saga state to be compensated, got: %s", saga.State)
	}

	// Verify opportunity was deleted
	if _, exists := oppRepo.opportunities[opportunityID]; exists {
		t.Error("Expected opportunity to be deleted during compensation")
	}
}

func TestSagaOrchestrator_Compensate_NotInCompensatingState(t *testing.T) {
	// Arrange
	orchestrator, _, _, _, _, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	// Saga is in started state, not compensating

	// Act
	err := orchestrator.Compensate(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error for saga not in compensating state, got nil")
	}
}

func TestSagaOrchestrator_Compensate_OpportunityDeletionFailed(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, oppRepo, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, false)
	customerID := uuid.New()
	request.CustomerID = &customerID
	customerService.customers[customerID] = &ports.CustomerInfo{ID: customerID, Name: "Test Customer"}

	saga := createTestSaga(tenantID, lead, request)
	saga.Start()

	// Complete steps up to opportunity creation
	validateStep := saga.GetStepByType(domain.StepTypeValidateLead)
	validateStep.Start()
	validateStep.Complete(map[string]interface{}{})

	lookupStep := saga.GetStepByType(domain.StepTypeLookupCustomer)
	if lookupStep != nil {
		lookupStep.Start()
		saga.SetCustomerID(customerID, false)
		lookupStep.Complete(map[string]interface{}{})
	}

	oppStep := saga.GetStepByType(domain.StepTypeCreateOpportunity)
	oppStep.Start()
	opportunityID := uuid.New()
	saga.SetOpportunityID(opportunityID)
	oppRepo.opportunities[opportunityID] = &domain.Opportunity{ID: opportunityID, TenantID: tenantID}
	oppStep.Complete(map[string]interface{}{"opportunity_id": opportunityID.String()})

	// Start compensation but set delete error
	saga.StartCompensation(errors.New("simulated failure"), domain.StepTypeMarkLeadConverted)
	oppRepo.deleteErr = errors.New("database error during delete")
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Compensate(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error for opportunity deletion failure, got nil")
	}

	if saga.State != domain.SagaStateFailed {
		t.Errorf("Expected saga state to be failed after compensation failure, got: %s", saga.State)
	}
}

// ============================================================================
// Test Cases - Resume
// ============================================================================

func TestSagaOrchestrator_Resume_RunningState(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	saga.Start() // Move to running state

	sagaRepo.sagas[saga.ID] = saga

	// Pre-create customer ID
	customerService.createCustomerID = uuid.New()

	// Act
	err := orchestrator.Resume(context.Background(), saga)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if saga.State != domain.SagaStateCompleted {
		t.Errorf("Expected saga state to be completed after resume, got: %s", saga.State)
	}
}

func TestSagaOrchestrator_Resume_CompensatingState(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, oppRepo, pipelineRepo, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, false)
	customerID := uuid.New()
	request.CustomerID = &customerID

	saga := createTestSaga(tenantID, lead, request)
	saga.Start()

	// Simulate completed opportunity step
	oppStep := saga.GetStepByType(domain.StepTypeCreateOpportunity)
	oppStep.Start()
	opportunityID := uuid.New()
	saga.SetOpportunityID(opportunityID)
	oppRepo.opportunities[opportunityID] = &domain.Opportunity{ID: opportunityID, TenantID: tenantID}
	oppStep.Complete(map[string]interface{}{})

	saga.StartCompensation(errors.New("simulated failure"), domain.StepTypeMarkLeadConverted)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Resume(context.Background(), saga)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if saga.State != domain.SagaStateCompensated {
		t.Errorf("Expected saga state to be compensated after resume, got: %s", saga.State)
	}
}

func TestSagaOrchestrator_Resume_CompletedState(t *testing.T) {
	// Arrange
	orchestrator, _, _, _, _, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	saga.Start()
	saga.Complete(&domain.LeadConversionResult{})

	// Act
	err := orchestrator.Resume(context.Background(), saga)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for completed saga, got: %v", err)
	}
}

// ============================================================================
// Test Cases - GetSagaStatus
// ============================================================================

func TestSagaOrchestrator_GetSagaStatus_Success(t *testing.T) {
	// Arrange
	orchestrator, _, _, _, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	result, err := orchestrator.GetSagaStatus(context.Background(), tenantID, saga.ID)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected saga result, got nil")
	}

	if result.ID != saga.ID {
		t.Errorf("Expected saga ID %s, got %s", saga.ID, result.ID)
	}
}

func TestSagaOrchestrator_GetSagaStatus_NotFound(t *testing.T) {
	// Arrange
	orchestrator, _, _, _, _, _ := setupSagaOrchestratorTest()

	// Act
	_, err := orchestrator.GetSagaStatus(context.Background(), uuid.New(), uuid.New())

	// Assert
	if err == nil {
		t.Fatal("Expected error for saga not found, got nil")
	}
}

// ============================================================================
// Table-Driven Tests
// ============================================================================

func TestSagaOrchestrator_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		setupFunc         func(*SagaMockLeadRepo, *MockSagaOpportunityRepository, *MockSagaPipelineRepository, *MockSagaCustomerService) (*domain.Lead, *domain.Pipeline, *domain.LeadConversionRequest)
		expectErr         bool
		expectCompensated bool
		expectCompleted   bool
	}{
		{
			name: "success with new customer",
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService) (*domain.Lead, *domain.Pipeline, *domain.LeadConversionRequest) {
				tenantID := uuid.New()
				lead := createTestQualifiedLead(tenantID)
				pipeline := createSagaTestPipeline(tenantID)

				leadRepo.leads[lead.ID] = lead
				pipelineRepo.pipelines[pipeline.ID] = pipeline

				request := &domain.LeadConversionRequest{
					LeadID:            lead.ID,
					PipelineID:        pipeline.ID,
					CreateNewCustomer: true,
				}
				return lead, pipeline, request
			},
			expectErr:       false,
			expectCompleted: true,
		},
		{
			name: "success with existing customer",
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService) (*domain.Lead, *domain.Pipeline, *domain.LeadConversionRequest) {
				tenantID := uuid.New()
				lead := createTestQualifiedLead(tenantID)
				pipeline := createSagaTestPipeline(tenantID)
				customerID := uuid.New()

				leadRepo.leads[lead.ID] = lead
				pipelineRepo.pipelines[pipeline.ID] = pipeline
				customerSvc.customers[customerID] = &ports.CustomerInfo{ID: customerID, TenantID: tenantID, Name: "Existing"}

				request := &domain.LeadConversionRequest{
					LeadID:            lead.ID,
					PipelineID:        pipeline.ID,
					CustomerID:        &customerID,
					CreateNewCustomer: false,
				}
				return lead, pipeline, request
			},
			expectErr:       false,
			expectCompleted: true,
		},
		{
			name: "failure - lead not found",
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService) (*domain.Lead, *domain.Pipeline, *domain.LeadConversionRequest) {
				tenantID := uuid.New()
				lead := createTestQualifiedLead(tenantID)
				pipeline := createSagaTestPipeline(tenantID)

				// Don't add lead to repo
				pipelineRepo.pipelines[pipeline.ID] = pipeline

				request := &domain.LeadConversionRequest{
					LeadID:            lead.ID,
					PipelineID:        pipeline.ID,
					CreateNewCustomer: true,
				}
				return lead, pipeline, request
			},
			expectErr:         true,
			expectCompensated: true,
		},
		{
			name: "failure - pipeline not found",
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService) (*domain.Lead, *domain.Pipeline, *domain.LeadConversionRequest) {
				tenantID := uuid.New()
				lead := createTestQualifiedLead(tenantID)
				pipeline := createSagaTestPipeline(tenantID)

				leadRepo.leads[lead.ID] = lead
				// Don't add pipeline to repo

				request := &domain.LeadConversionRequest{
					LeadID:            lead.ID,
					PipelineID:        pipeline.ID,
					CreateNewCustomer: true,
				}
				return lead, pipeline, request
			},
			expectErr:         true,
			expectCompensated: true,
		},
		{
			name: "failure - customer service unavailable",
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService) (*domain.Lead, *domain.Pipeline, *domain.LeadConversionRequest) {
				tenantID := uuid.New()
				lead := createTestQualifiedLead(tenantID)
				pipeline := createSagaTestPipeline(tenantID)

				leadRepo.leads[lead.ID] = lead
				pipelineRepo.pipelines[pipeline.ID] = pipeline
				customerSvc.createErr = errors.New("service unavailable")

				request := &domain.LeadConversionRequest{
					LeadID:            lead.ID,
					PipelineID:        pipeline.ID,
					CreateNewCustomer: true,
				}
				return lead, pipeline, request
			},
			expectErr:         true,
			expectCompensated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orchestrator, leadRepo, oppRepo, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

			lead, _, request := tt.setupFunc(leadRepo, oppRepo, pipelineRepo, customerService)
			saga := createTestSaga(lead.TenantID, lead, request)
			sagaRepo.sagas[saga.ID] = saga

			err := orchestrator.Execute(context.Background(), saga)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error for %s, got: %v", tt.name, err)
			}
			if tt.expectCompleted && saga.State != domain.SagaStateCompleted {
				t.Errorf("Expected saga to be completed for %s, got: %s", tt.name, saga.State)
			}
			if tt.expectCompensated && saga.State != domain.SagaStateCompensated && saga.State != domain.SagaStateFailed {
				t.Errorf("Expected saga to be compensated or failed for %s, got: %s", tt.name, saga.State)
			}
		})
	}
}

func TestSagaOrchestrator_StepExecution_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		stepType  domain.SagaStepType
		setupFunc func(*SagaMockLeadRepo, *MockSagaOpportunityRepository, *MockSagaPipelineRepository, *MockSagaCustomerService, *domain.LeadConversionSaga)
		expectErr bool
	}{
		{
			name:     "validate_lead - success",
			stepType: domain.StepTypeValidateLead,
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService, saga *domain.LeadConversionSaga) {
				lead := createTestQualifiedLead(saga.TenantID)
				lead.ID = saga.LeadID
				leadRepo.leads[lead.ID] = lead

				pipeline := createSagaTestPipeline(saga.TenantID)
				pipeline.ID = saga.Request.PipelineID
				pipelineRepo.pipelines[pipeline.ID] = pipeline
			},
			expectErr: false,
		},
		{
			name:     "validate_lead - lead not qualified",
			stepType: domain.StepTypeValidateLead,
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService, saga *domain.LeadConversionSaga) {
				lead := createTestQualifiedLead(saga.TenantID)
				lead.ID = saga.LeadID
				lead.Status = domain.LeadStatusNew // Not qualified
				leadRepo.leads[lead.ID] = lead

				pipeline := createSagaTestPipeline(saga.TenantID)
				pipeline.ID = saga.Request.PipelineID
				pipelineRepo.pipelines[pipeline.ID] = pipeline
			},
			expectErr: true,
		},
		{
			name:     "create_customer - success",
			stepType: domain.StepTypeCreateCustomer,
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService, saga *domain.LeadConversionSaga) {
				lead := createTestQualifiedLead(saga.TenantID)
				lead.ID = saga.LeadID
				leadRepo.leads[lead.ID] = lead
			},
			expectErr: false,
		},
		{
			name:     "create_customer - service error",
			stepType: domain.StepTypeCreateCustomer,
			setupFunc: func(leadRepo *SagaMockLeadRepo, oppRepo *MockSagaOpportunityRepository, pipelineRepo *MockSagaPipelineRepository, customerSvc *MockSagaCustomerService, saga *domain.LeadConversionSaga) {
				lead := createTestQualifiedLead(saga.TenantID)
				lead.ID = saga.LeadID
				leadRepo.leads[lead.ID] = lead
				customerSvc.createErr = errors.New("service error")
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orchestrator, leadRepo, oppRepo, pipelineRepo, _, customerService := setupSagaOrchestratorTest()

			tenantID := uuid.New()
			lead := createTestQualifiedLead(tenantID)
			pipeline := createSagaTestPipeline(tenantID)

			request := &domain.LeadConversionRequest{
				LeadID:            lead.ID,
				PipelineID:        pipeline.ID,
				CreateNewCustomer: true,
			}
			saga := createTestSaga(tenantID, lead, request)

			tt.setupFunc(leadRepo, oppRepo, pipelineRepo, customerService, saga)

			step := saga.GetStepByType(tt.stepType)
			if step == nil {
				t.Fatalf("Step type %s not found in saga", tt.stepType)
			}

			err := orchestrator.executeStep(context.Background(), saga, step)

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

func BenchmarkSagaOrchestrator_Execute_Success(b *testing.B) {
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset state for each iteration
		request := createTestConversionRequest(lead, pipeline, true)
		saga := createTestSaga(tenantID, lead, request)
		sagaRepo.sagas[saga.ID] = saga

		// Reset customer service state
		customerService.createErr = nil
		customerService.createCustomerID = uuid.New()

		_ = orchestrator.Execute(ctx, saga)
	}
}

func BenchmarkSagaOrchestrator_Execute_WithExistingCustomer(b *testing.B) {
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	customerID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline
	customerService.customers[customerID] = &ports.CustomerInfo{
		ID:       customerID,
		TenantID: tenantID,
		Name:     "Existing Customer",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := &domain.LeadConversionRequest{
			LeadID:            lead.ID,
			PipelineID:        pipeline.ID,
			CustomerID:        &customerID,
			CreateNewCustomer: false,
		}
		saga := createTestSaga(tenantID, lead, request)
		sagaRepo.sagas[saga.ID] = saga

		_ = orchestrator.Execute(ctx, saga)
	}
}

func BenchmarkSagaOrchestrator_Compensate(b *testing.B) {
	orchestrator, leadRepo, oppRepo, pipelineRepo, sagaRepo, customerService := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Setup saga in compensating state
		request := createTestConversionRequest(lead, pipeline, true)
		saga := createTestSaga(tenantID, lead, request)
		saga.Start()

		// Simulate completed steps
		customerID := uuid.New()
		customerService.customers[customerID] = &ports.CustomerInfo{ID: customerID}
		saga.SetCustomerID(customerID, true)

		opportunityID := uuid.New()
		oppRepo.opportunities[opportunityID] = &domain.Opportunity{ID: opportunityID, TenantID: tenantID}
		saga.SetOpportunityID(opportunityID)

		// Mark steps as completed
		for _, step := range saga.Steps {
			if step.Type != domain.StepTypeMarkLeadConverted && step.Type != domain.StepTypePublishEvents {
				step.Start()
				step.Complete(map[string]interface{}{})
			}
		}

		saga.StartCompensation(errors.New("test failure"), domain.StepTypeMarkLeadConverted)
		sagaRepo.sagas[saga.ID] = saga

		_ = orchestrator.Compensate(ctx, saga)

		// Cleanup
		delete(oppRepo.opportunities, opportunityID)
		delete(customerService.customers, customerID)
	}
}

func BenchmarkSagaOrchestrator_GetSagaStatus(b *testing.B) {
	orchestrator, _, _, _, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orchestrator.GetSagaStatus(ctx, tenantID, saga.ID)
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestSagaOrchestrator_Execute_ContextTimeout(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Create already cancelled context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure timeout has passed

	// Act - should handle cancelled context gracefully without panic
	_ = orchestrator.Execute(ctx, saga)

	// Assert - no panic occurred, test passes
}

func TestSagaOrchestrator_Compensate_ContextTimeout(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, oppRepo, pipelineRepo, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, false)
	customerID := uuid.New()
	request.CustomerID = &customerID

	saga := createTestSaga(tenantID, lead, request)
	saga.Start()

	// Setup completed opportunity
	oppStep := saga.GetStepByType(domain.StepTypeCreateOpportunity)
	oppStep.Start()
	opportunityID := uuid.New()
	saga.SetOpportunityID(opportunityID)
	oppRepo.opportunities[opportunityID] = &domain.Opportunity{ID: opportunityID, TenantID: tenantID}
	oppStep.Complete(map[string]interface{}{})

	saga.StartCompensation(errors.New("test failure"), domain.StepTypeMarkLeadConverted)
	sagaRepo.sagas[saga.ID] = saga

	// Create already cancelled context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	// Act - should handle cancelled context gracefully
	_ = orchestrator.Compensate(ctx, saga)

	// Assert - no panic occurred
}

func TestSagaOrchestrator_GetSagaStatus_ContextCancelled(t *testing.T) {
	// Arrange
	orchestrator, _, _, _, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	_, _ = orchestrator.GetSagaStatus(ctx, tenantID, saga.ID)

	// Assert - no panic occurred
}

// ============================================================================
// Edge Cases and Additional Tests
// ============================================================================

func TestSagaOrchestrator_Execute_AlreadyCompletedSaga(t *testing.T) {
	// Arrange
	orchestrator, _, _, _, _, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	saga.Start()
	saga.Complete(&domain.LeadConversionResult{
		LeadID:        saga.LeadID,
		OpportunityID: uuid.New(),
	})

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert - should return error as saga cannot transition from completed
	if err == nil {
		t.Fatal("Expected error for already completed saga, got nil")
	}
}

func TestSagaOrchestrator_Execute_SaveSagaError(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga
	sagaRepo.updateErr = errors.New("database error")

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error when saving saga fails, got nil")
	}
}

func TestSagaOrchestrator_Execute_ConcurrentExecution(t *testing.T) {
	// This test verifies the orchestrator can handle concurrent saga executions
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()

	// Create multiple leads and pipelines
	numSagas := 5
	sagas := make([]*domain.LeadConversionSaga, numSagas)

	for i := 0; i < numSagas; i++ {
		lead := createTestQualifiedLead(tenantID)
		pipeline := createSagaTestPipeline(tenantID)

		leadRepo.leads[lead.ID] = lead
		pipelineRepo.pipelines[pipeline.ID] = pipeline

		request := createTestConversionRequest(lead, pipeline, true)
		saga := createTestSaga(tenantID, lead, request)
		sagaRepo.sagas[saga.ID] = saga
		sagas[i] = saga
	}

	// Execute all sagas concurrently
	var wg sync.WaitGroup
	errors := make([]error, numSagas)

	for i := 0; i < numSagas; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errors[idx] = orchestrator.Execute(context.Background(), sagas[idx])
		}(i)
	}

	wg.Wait()

	// Verify all completed successfully
	for i, err := range errors {
		if err != nil {
			t.Errorf("Saga %d failed: %v", i, err)
		}
	}

	for i, saga := range sagas {
		if saga.State != domain.SagaStateCompleted {
			t.Errorf("Saga %d not completed, state: %s", i, saga.State)
		}
	}
}

func TestSagaOrchestrator_Execute_LeadAlreadyConverted(t *testing.T) {
	// Arrange
	orchestrator, leadRepo, _, pipelineRepo, sagaRepo, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	lead.Status = domain.LeadStatusConverted // Already converted
	pipeline := createSagaTestPipeline(tenantID)

	leadRepo.leads[lead.ID] = lead
	pipelineRepo.pipelines[pipeline.ID] = pipeline

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)
	sagaRepo.sagas[saga.ID] = saga

	// Act
	err := orchestrator.Execute(context.Background(), saga)

	// Assert
	if err == nil {
		t.Fatal("Expected error for already converted lead, got nil")
	}
}

func TestSagaOrchestrator_Execute_NoStepHandler(t *testing.T) {
	// Test that executing a step without a handler returns an error
	orchestrator, _, _, _, _, _ := setupSagaOrchestratorTest()

	tenantID := uuid.New()
	lead := createTestQualifiedLead(tenantID)
	pipeline := createSagaTestPipeline(tenantID)

	request := createTestConversionRequest(lead, pipeline, true)
	saga := createTestSaga(tenantID, lead, request)

	// Create a step with an unknown type
	unknownStep := &domain.SagaStep{
		ID:     uuid.New(),
		Type:   domain.SagaStepType("unknown_step"),
		Order:  99,
		Status: domain.StepStatusPending,
		Input:  make(map[string]interface{}),
		Output: make(map[string]interface{}),
	}

	// Act
	err := orchestrator.executeStep(context.Background(), saga, unknownStep)

	// Assert
	if err == nil {
		t.Fatal("Expected error for unknown step type, got nil")
	}
}
