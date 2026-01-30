// Package usecase contains tests for the notification template use cases.
package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kilang-desa-murni/crm/internal/notification/application"
	"github.com/kilang-desa-murni/crm/internal/notification/application/dto"
	"github.com/kilang-desa-murni/crm/internal/notification/application/ports"
	"github.com/kilang-desa-murni/crm/internal/notification/domain"
)

// =============================================================================
// Mock Implementations
// =============================================================================

// mockTemplateRepository is a mock implementation of domain.TemplateRepository.
type mockTemplateRepository struct {
	mu             sync.RWMutex
	templates      map[uuid.UUID]*domain.NotificationTemplate
	templateByCode map[string]*domain.NotificationTemplate
	createErr      error
	updateErr      error
	deleteErr      error
	hardDeleteErr  error
	findByIDErr    error
	findByCodeErr  error
	listErr        error
	existsErr      error
}

func newMockTemplateRepository() *mockTemplateRepository {
	return &mockTemplateRepository{
		templates:      make(map[uuid.UUID]*domain.NotificationTemplate),
		templateByCode: make(map[string]*domain.NotificationTemplate),
	}
}

func (m *mockTemplateRepository) Create(ctx context.Context, template *domain.NotificationTemplate) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[template.ID] = template
	m.templateByCode[template.Code] = template
	return nil
}

func (m *mockTemplateRepository) Update(ctx context.Context, template *domain.NotificationTemplate) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[template.ID] = template
	m.templateByCode[template.Code] = template
	return nil
}

func (m *mockTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.templates[id]; ok {
		now := time.Now().UTC()
		t.DeletedAt = &now
	}
	return nil
}

func (m *mockTemplateRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	if m.hardDeleteErr != nil {
		return m.hardDeleteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.templates[id]; ok {
		delete(m.templateByCode, t.Code)
		delete(m.templates, id)
	}
	return nil
}

func (m *mockTemplateRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.NotificationTemplate, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.templates[id]; ok {
		return t, nil
	}
	return nil, errors.New("template not found")
}

func (m *mockTemplateRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.NotificationTemplate, error) {
	if m.findByCodeErr != nil {
		return nil, m.findByCodeErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.templateByCode[code]; ok && t.TenantID == tenantID {
		return t, nil
	}
	return nil, errors.New("template not found")
}

func (m *mockTemplateRepository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.NotificationTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.templates {
		if t.Name == name && t.TenantID == tenantID {
			return t, nil
		}
	}
	return nil, errors.New("template not found")
}

func (m *mockTemplateRepository) List(ctx context.Context, filter domain.TemplateFilter) (*domain.TemplateList, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var templates []*domain.NotificationTemplate
	for _, t := range m.templates {
		if filter.TenantID != nil && t.TenantID != *filter.TenantID {
			continue
		}
		if t.DeletedAt != nil && !filter.IncludeDeleted {
			continue
		}
		templates = append(templates, t)
	}

	return &domain.TemplateList{
		Templates: templates,
		Total:     int64(len(templates)),
		Offset:    filter.Offset,
		Limit:     filter.Limit,
		HasMore:   false,
	}, nil
}

func (m *mockTemplateRepository) FindByType(ctx context.Context, tenantID uuid.UUID, notifType domain.NotificationType) ([]*domain.NotificationTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var templates []*domain.NotificationTemplate
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.Type == notifType {
			templates = append(templates, t)
		}
	}
	return templates, nil
}

func (m *mockTemplateRepository) FindByChannel(ctx context.Context, tenantID uuid.UUID, channel domain.NotificationChannel) ([]*domain.NotificationTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var templates []*domain.NotificationTemplate
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.SupportsChannel(channel) {
			templates = append(templates, t)
		}
	}
	return templates, nil
}

func (m *mockTemplateRepository) FindDefault(ctx context.Context, tenantID uuid.UUID, notifType domain.NotificationType) (*domain.NotificationTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.Type == notifType && t.IsDefault {
			return t, nil
		}
	}
	return nil, errors.New("default template not found")
}

func (m *mockTemplateRepository) FindActive(ctx context.Context, tenantID uuid.UUID) ([]*domain.NotificationTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var templates []*domain.NotificationTemplate
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.IsActive && t.DeletedAt == nil {
			templates = append(templates, t)
		}
	}
	return templates, nil
}

func (m *mockTemplateRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var count int64
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.DeletedAt == nil {
			count++
		}
	}
	return count, nil
}

func (m *mockTemplateRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.templates[id]
	return ok, nil
}

func (m *mockTemplateRepository) ExistsByCode(ctx context.Context, tenantID uuid.UUID, code string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.templateByCode[code]; ok && t.TenantID == tenantID {
		return true, nil
	}
	return false, nil
}

func (m *mockTemplateRepository) GetVersion(ctx context.Context, id uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.templates[id]; ok {
		return t.Version, nil
	}
	return 0, errors.New("template not found")
}

func (m *mockTemplateRepository) IncrementUsageCount(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.templates[id]; ok {
		t.UsageCount++
		return nil
	}
	return errors.New("template not found")
}

func (m *mockTemplateRepository) FindByTag(ctx context.Context, tenantID uuid.UUID, tag string) ([]*domain.NotificationTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var templates []*domain.NotificationTemplate
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.HasTag(tag) {
			templates = append(templates, t)
		}
	}
	return templates, nil
}

func (m *mockTemplateRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, filter domain.TemplateFilter) (*domain.TemplateList, error) {
	return m.List(ctx, filter)
}

// mockNotificationRepository is a mock implementation of domain.NotificationRepository.
type mockNotificationRepository struct{}

func (m *mockNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	return nil
}
func (m *mockNotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	return nil
}
func (m *mockNotificationRepository) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockNotificationRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockNotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepository) List(ctx context.Context, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindByRecipient(ctx context.Context, tenantID uuid.UUID, recipientID uuid.UUID, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindByStatus(ctx context.Context, tenantID uuid.UUID, status domain.NotificationStatus, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindPending(ctx context.Context, limit int) ([]*domain.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindScheduled(ctx context.Context, before time.Time, limit int) ([]*domain.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindRetryable(ctx context.Context, before time.Time, limit int) ([]*domain.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindByBatch(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepository) FindBySourceEvent(ctx context.Context, tenantID uuid.UUID, sourceEvent string, sourceEntityID uuid.UUID) ([]*domain.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockNotificationRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.NotificationStatus]int64, error) {
	return nil, nil
}
func (m *mockNotificationRepository) CountByChannel(ctx context.Context, tenantID uuid.UUID) (map[domain.NotificationChannel]int64, error) {
	return nil, nil
}
func (m *mockNotificationRepository) GetStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*domain.NotificationStats, error) {
	return nil, nil
}
func (m *mockNotificationRepository) BulkCreate(ctx context.Context, notifications []*domain.Notification) error {
	return nil
}
func (m *mockNotificationRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.NotificationStatus) error {
	return nil
}
func (m *mockNotificationRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockNotificationRepository) GetVersion(ctx context.Context, id uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockNotificationRepository) FindUnread(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*domain.Notification, error) {
	return nil, nil
}
func (m *mockNotificationRepository) CountUnread(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockNotificationRepository) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	return nil
}
func (m *mockNotificationRepository) DeleteOld(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

// mockEventPublisher is a mock implementation of ports.EventPublisher.
type mockEventPublisher struct {
	events []domain.DomainEvent
	err    error
}

func (m *mockEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, events...)
	return nil
}

// mockCacheService is a mock implementation of ports.CacheService.
type mockCacheService struct {
	data map[string][]byte
	err  error
}

func newMockCacheService() *mockCacheService {
	return &mockCacheService{
		data: make(map[string][]byte),
	}
}

func (m *mockCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[key], nil
}

func (m *mockCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.err != nil {
		return m.err
	}
	m.data[key] = value
	return nil
}

func (m *mockCacheService) Delete(ctx context.Context, key string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.data, key)
	return nil
}

func (m *mockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	_, ok := m.data[key]
	return ok, nil
}

func (m *mockCacheService) GetOrSet(ctx context.Context, key string, loader func() ([]byte, error), ttl time.Duration) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	if v, ok := m.data[key]; ok {
		return v, nil
	}
	v, err := loader()
	if err != nil {
		return nil, err
	}
	m.data[key] = v
	return v, nil
}

func (m *mockCacheService) Invalidate(ctx context.Context, pattern string) error {
	return nil
}

// mockIdGenerator is a mock implementation of ports.IdGenerator.
type mockIdGenerator struct {
	id string
}

func (m *mockIdGenerator) Generate() string {
	if m.id != "" {
		return m.id
	}
	return uuid.New().String()
}

func (m *mockIdGenerator) GenerateWithPrefix(prefix string) string {
	return prefix + "_" + m.Generate()
}

// mockTimeProvider is a mock implementation of ports.TimeProvider.
type mockTimeProvider struct {
	time time.Time
}

func (m *mockTimeProvider) Now() time.Time {
	if m.time.IsZero() {
		return time.Now()
	}
	return m.time
}

func (m *mockTimeProvider) NowUTC() time.Time {
	if m.time.IsZero() {
		return time.Now().UTC()
	}
	return m.time.UTC()
}

// mockMetricsCollector is a mock implementation of ports.MetricsCollector.
type mockMetricsCollector struct {
	counters   map[string]int
	histograms map[string][]float64
	gauges     map[string]float64
	durations  map[string][]time.Duration
}

func newMockMetricsCollector() *mockMetricsCollector {
	return &mockMetricsCollector{
		counters:   make(map[string]int),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
		durations:  make(map[string][]time.Duration),
	}
}

func (m *mockMetricsCollector) IncrementCounter(ctx context.Context, name string, tags map[string]string) {
	m.counters[name]++
}

func (m *mockMetricsCollector) RecordHistogram(ctx context.Context, name string, value float64, tags map[string]string) {
	m.histograms[name] = append(m.histograms[name], value)
}

func (m *mockMetricsCollector) RecordGauge(ctx context.Context, name string, value float64, tags map[string]string) {
	m.gauges[name] = value
}

func (m *mockMetricsCollector) RecordDuration(ctx context.Context, name string, duration time.Duration, tags map[string]string) {
	m.durations[name] = append(m.durations[name], duration)
}

// mockLogger is a mock implementation of ports.Logger.
type mockLogger struct {
	logs []map[string]interface{}
}

func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "debug", "msg": msg, "fields": fields})
}

func (m *mockLogger) Info(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "info", "msg": msg, "fields": fields})
}

func (m *mockLogger) Warn(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "warn", "msg": msg, "fields": fields})
}

func (m *mockLogger) Error(msg string, err error, fields map[string]interface{}) {
	m.logs = append(m.logs, map[string]interface{}{"level": "error", "msg": msg, "error": err, "fields": fields})
}

func (m *mockLogger) WithContext(ctx context.Context) ports.Logger {
	return m
}

func (m *mockLogger) WithFields(fields map[string]interface{}) ports.Logger {
	return m
}

// =============================================================================
// Test Helpers
// =============================================================================

// testFixtures contains common test data.
type testFixtures struct {
	tenantID   uuid.UUID
	userID     uuid.UUID
	templateID uuid.UUID
}

func newTestFixtures() *testFixtures {
	return &testFixtures{
		tenantID:   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		userID:     uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		templateID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
	}
}

// createTestTemplate creates a test template with default values.
func createTestTemplate(tenantID uuid.UUID, code, name string, channel domain.NotificationChannel) (*domain.NotificationTemplate, error) {
	template, err := domain.NewNotificationTemplate(tenantID, code, name, domain.TypeTransactional)
	if err != nil {
		return nil, err
	}

	switch channel {
	case domain.ChannelEmail:
		err = template.SetEmailTemplate(&domain.EmailTemplateContent{
			Subject:  "Test Subject {{.Name}}",
			Body:     "Hello {{.Name}}, this is a test.",
			HTMLBody: "<p>Hello {{.Name}}, this is a test.</p>",
		})
	case domain.ChannelSMS:
		err = template.SetSMSTemplate(&domain.SMSTemplateContent{
			Body: "Hello {{.Name}}, this is a test SMS.",
		})
	case domain.ChannelPush:
		err = template.SetPushTemplate(&domain.PushTemplateContent{
			Title: "Test Notification",
			Body:  "Hello {{.Name}}",
		})
	case domain.ChannelInApp:
		err = template.SetInAppTemplate(&domain.InAppTemplateContent{
			Title: "In-App Test",
			Body:  "Hello {{.Name}}",
		})
	}

	if err != nil {
		return nil, err
	}

	return template, nil
}

// createTemplateTestUseCase creates a template use case with mock dependencies.
func createTemplateTestUseCase() (TemplateUseCase, *mockTemplateRepository, *testFixtures) {
	repo := newMockTemplateRepository()
	fixtures := newTestFixtures()

	cfg := TemplateUseCaseConfig{
		TemplateRepo:     repo,
		NotificationRepo: &mockNotificationRepository{},
		EventPublisher:   &mockEventPublisher{},
		Cache:            newMockCacheService(),
		IdGenerator:      &mockIdGenerator{},
		TimeProvider:     &mockTimeProvider{},
		Metrics:          newMockMetricsCollector(),
		Logger:           &mockLogger{},
	}

	uc := NewTemplateUseCase(cfg)
	return uc, repo, fixtures
}

// createValidCreateTemplateRequest creates a valid create template request.
func createValidCreateTemplateRequest(tenantID, userID uuid.UUID) *dto.CreateTemplateRequest {
	return &dto.CreateTemplateRequest{
		TenantID:      tenantID.String(),
		Code:          "welcome_email",
		Name:          "Welcome Email Template",
		Description:   "Template for welcome emails",
		Channel:       "email",
		Type:          "transactional",
		Category:      "onboarding",
		Subject:       "Welcome {{.Name}}!",
		Body:          "Hello {{.Name}}, welcome to our platform!",
		HTMLBody:      "<p>Hello {{.Name}}, welcome to our platform!</p>",
		DefaultLocale: "en",
		Tags:          []string{"welcome", "onboarding"},
		IsDraft:       false,
		CreatedBy:     userID.String(),
	}
}

// =============================================================================
// CreateTemplate Tests
// =============================================================================

func TestCreateTemplate_Success(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := createValidCreateTemplateRequest(fixtures.tenantID, fixtures.userID)

	resp, err := uc.CreateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.TemplateID == "" {
		t.Error("expected template ID to be set")
	}

	if resp.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, resp.Name)
	}

	if resp.Version != 1 {
		t.Errorf("expected version 1, got %d", resp.Version)
	}

	if resp.Status != "active" {
		t.Errorf("expected status 'active', got %s", resp.Status)
	}

	if resp.Message != "Template created successfully" {
		t.Errorf("expected success message, got %s", resp.Message)
	}
}

func TestCreateTemplate_AsDraft(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := createValidCreateTemplateRequest(fixtures.tenantID, fixtures.userID)
	req.IsDraft = true

	resp, err := uc.CreateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp.Status != "draft" {
		t.Errorf("expected status 'draft', got %s", resp.Status)
	}
}

func TestCreateTemplate_SMSChannel(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.CreateTemplateRequest{
		TenantID:      fixtures.tenantID.String(),
		Code:          "otp_sms",
		Name:          "OTP SMS Template",
		Channel:       "sms",
		Type:          "verification",
		Body:          "Your OTP is {{.Code}}",
		DefaultLocale: "en",
		CreatedBy:     fixtures.userID.String(),
	}

	resp, err := uc.CreateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestCreateTemplate_PushChannel(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.CreateTemplateRequest{
		TenantID:      fixtures.tenantID.String(),
		Code:          "new_message_push",
		Name:          "New Message Push",
		Channel:       "push",
		Type:          "alert",
		Subject:       "New Message",
		Body:          "You have a new message from {{.Sender}}",
		DefaultLocale: "en",
		CreatedBy:     fixtures.userID.String(),
	}

	resp, err := uc.CreateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestCreateTemplate_InAppChannel(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.CreateTemplateRequest{
		TenantID:      fixtures.tenantID.String(),
		Code:          "welcome_inapp",
		Name:          "Welcome In-App",
		Channel:       "in_app",
		Type:          "welcome",
		Subject:       "Welcome!",
		Body:          "Welcome to our platform, {{.Name}}!",
		DefaultLocale: "en",
		CreatedBy:     fixtures.userID.String(),
	}

	resp, err := uc.CreateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestCreateTemplate_DuplicateCode(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create first template
	template, _ := createTestTemplate(fixtures.tenantID, "duplicate_code", "First Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode["duplicate_code"] = template

	// Try to create another with same code
	req := createValidCreateTemplateRequest(fixtures.tenantID, fixtures.userID)
	req.Code = "duplicate_code"

	_, err := uc.CreateTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for duplicate code")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeTemplateAlreadyExists {
		t.Errorf("expected error code %s, got %s", application.ErrCodeTemplateAlreadyExists, appErr.Code)
	}
}

func TestCreateTemplate_InvalidTenantID(t *testing.T) {
	uc, _, _ := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.CreateTemplateRequest{
		TenantID:      "invalid-uuid",
		Code:          "test_template",
		Name:          "Test Template",
		Channel:       "email",
		Type:          "transactional",
		Subject:       "Test",
		Body:          "Test body",
		DefaultLocale: "en",
	}

	_, err := uc.CreateTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for invalid tenant ID")
	}
}

func TestCreateTemplate_MissingRequiredFields(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	tests := []struct {
		name    string
		modify  func(*dto.CreateTemplateRequest)
		wantErr string
	}{
		{
			name:    "missing tenant_id",
			modify:  func(req *dto.CreateTemplateRequest) { req.TenantID = "" },
			wantErr: "tenant_id is required",
		},
		{
			name:    "missing code",
			modify:  func(req *dto.CreateTemplateRequest) { req.Code = "" },
			wantErr: "code is required",
		},
		{
			name:    "missing name",
			modify:  func(req *dto.CreateTemplateRequest) { req.Name = "" },
			wantErr: "name is required",
		},
		{
			name:    "missing channel",
			modify:  func(req *dto.CreateTemplateRequest) { req.Channel = "" },
			wantErr: "channel is required",
		},
		{
			name:    "missing type",
			modify:  func(req *dto.CreateTemplateRequest) { req.Type = "" },
			wantErr: "type is required",
		},
		{
			name:    "missing body",
			modify:  func(req *dto.CreateTemplateRequest) { req.Body = "" },
			wantErr: "body is required",
		},
		{
			name:    "email without subject",
			modify:  func(req *dto.CreateTemplateRequest) { req.Subject = "" },
			wantErr: "subject is required for email templates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createValidCreateTemplateRequest(fixtures.tenantID, fixtures.userID)
			tt.modify(req)

			_, err := uc.CreateTemplate(ctx, req)
			if err == nil {
				t.Fatalf("expected error containing '%s'", tt.wantErr)
			}

			if !containsString(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing '%s', got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestCreateTemplate_CodeTooLong(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := createValidCreateTemplateRequest(fixtures.tenantID, fixtures.userID)
	req.Code = string(make([]byte, 101)) // 101 characters

	_, err := uc.CreateTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for code too long")
	}
}

func TestCreateTemplate_NameTooLong(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := createValidCreateTemplateRequest(fixtures.tenantID, fixtures.userID)
	req.Name = string(make([]byte, 256)) // 256 characters

	_, err := uc.CreateTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for name too long")
	}
}

func TestCreateTemplate_WithVariables(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := createValidCreateTemplateRequest(fixtures.tenantID, fixtures.userID)
	req.Variables = []dto.TemplateVariableDTO{
		{
			Name:        "Name",
			Type:        "string",
			Required:    true,
			Description: "User's name",
		},
		{
			Name:         "Company",
			Type:         "string",
			Required:     false,
			DefaultValue: "Acme Inc",
		},
	}

	resp, err := uc.CreateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

// =============================================================================
// GetTemplate Tests
// =============================================================================

func TestGetTemplate_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "test_template", "Test Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.GetTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
	}

	result, err := uc.GetTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.ID != template.ID.String() {
		t.Errorf("expected ID %s, got %s", template.ID.String(), result.ID)
	}

	if result.Code != template.Code {
		t.Errorf("expected code %s, got %s", template.Code, result.Code)
	}

	if result.Name != template.Name {
		t.Errorf("expected name %s, got %s", template.Name, result.Name)
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.GetTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: uuid.New().String(),
	}

	_, err := uc.GetTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for not found template")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeTemplateNotFound {
		t.Errorf("expected error code %s, got %s", application.ErrCodeTemplateNotFound, appErr.Code)
	}
}

func TestGetTemplate_InvalidTemplateID(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.GetTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: "invalid-uuid",
	}

	_, err := uc.GetTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for invalid template ID")
	}
}

func TestGetTemplate_WrongTenant(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template for one tenant
	template, _ := createTestTemplate(fixtures.tenantID, "test_template", "Test Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	// Try to get with different tenant
	req := &dto.GetTemplateRequest{
		TenantID:   uuid.New().String(),
		TemplateID: template.ID.String(),
	}

	_, err := uc.GetTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for wrong tenant")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeForbidden {
		t.Errorf("expected error code %s, got %s", application.ErrCodeForbidden, appErr.Code)
	}
}

// =============================================================================
// GetTemplateByCode Tests
// =============================================================================

func TestGetTemplateByCode_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "welcome_email", "Welcome Email", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.GetTemplateByCodeRequest{
		TenantID: fixtures.tenantID.String(),
		Code:     "welcome_email",
	}

	result, err := uc.GetTemplateByCode(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Code != "welcome_email" {
		t.Errorf("expected code 'welcome_email', got %s", result.Code)
	}
}

func TestGetTemplateByCode_NotFound(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.GetTemplateByCodeRequest{
		TenantID: fixtures.tenantID.String(),
		Code:     "nonexistent_template",
	}

	_, err := uc.GetTemplateByCode(ctx, req)
	if err == nil {
		t.Fatal("expected error for not found template")
	}
}

// =============================================================================
// UpdateTemplate Tests
// =============================================================================

func TestUpdateTemplate_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "test_template", "Test Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	newName := "Updated Template Name"
	newDescription := "Updated description"

	req := &dto.UpdateTemplateRequest{
		TenantID:    fixtures.tenantID.String(),
		TemplateID:  template.ID.String(),
		Name:        &newName,
		Description: &newDescription,
		UpdatedBy:   fixtures.userID.String(),
	}

	resp, err := uc.UpdateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Message != "Template updated successfully" {
		t.Errorf("expected success message, got %s", resp.Message)
	}

	// Verify update
	updated := repo.templates[template.ID]
	if updated.Name != newName {
		t.Errorf("expected name %s, got %s", newName, updated.Name)
	}

	if updated.Description != newDescription {
		t.Errorf("expected description %s, got %s", newDescription, updated.Description)
	}
}

func TestUpdateTemplate_NotFound(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	newName := "New Name"
	req := &dto.UpdateTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: uuid.New().String(),
		Name:       &newName,
	}

	_, err := uc.UpdateTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for not found template")
	}
}

func TestUpdateTemplate_LockedTemplate(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create locked template
	template, _ := createTestTemplate(fixtures.tenantID, "locked_template", "Locked Template", domain.ChannelEmail)
	template.IsLocked = true
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	newName := "New Name"
	req := &dto.UpdateTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Name:       &newName,
	}

	_, err := uc.UpdateTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for locked template")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeInvalidState {
		t.Errorf("expected error code %s, got %s", application.ErrCodeInvalidState, appErr.Code)
	}
}

func TestUpdateTemplate_ArchivedTemplate(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create archived template
	template, _ := createTestTemplate(fixtures.tenantID, "archived_template", "Archived Template", domain.ChannelEmail)
	now := time.Now().UTC()
	template.DeletedAt = &now
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	newName := "New Name"
	req := &dto.UpdateTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Name:       &newName,
	}

	_, err := uc.UpdateTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for archived template")
	}
}

func TestUpdateTemplate_ContentUpdate(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "test_template", "Test Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	newSubject := "New Subject {{.Name}}"
	newBody := "New body {{.Name}}"
	newHTMLBody := "<p>New HTML body {{.Name}}</p>"

	req := &dto.UpdateTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Subject:    &newSubject,
		Body:       &newBody,
		HTMLBody:   &newHTMLBody,
	}

	resp, err := uc.UpdateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	// Verify content update
	updated := repo.templates[template.ID]
	if updated.EmailTemplate.Subject != newSubject {
		t.Errorf("expected subject %s, got %s", newSubject, updated.EmailTemplate.Subject)
	}
}

// =============================================================================
// DeleteTemplate Tests
// =============================================================================

func TestDeleteTemplate_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "test_template", "Test Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.DeleteTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
	}

	resp, err := uc.DeleteTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Message != "Template deleted successfully" {
		t.Errorf("expected success message, got %s", resp.Message)
	}

	// Verify deletion
	if _, ok := repo.templates[template.ID]; ok {
		t.Error("expected template to be deleted")
	}
}

func TestDeleteTemplate_NotFound(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.DeleteTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: uuid.New().String(),
	}

	_, err := uc.DeleteTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for not found template")
	}
}

func TestDeleteTemplate_LockedWithoutForce(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create locked template
	template, _ := createTestTemplate(fixtures.tenantID, "locked_template", "Locked Template", domain.ChannelEmail)
	template.IsLocked = true
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.DeleteTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Force:      false,
	}

	_, err := uc.DeleteTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for locked template without force")
	}
}

func TestDeleteTemplate_LockedWithForce(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create locked template
	template, _ := createTestTemplate(fixtures.tenantID, "locked_template", "Locked Template", domain.ChannelEmail)
	template.IsLocked = true
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.DeleteTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Force:      true,
	}

	resp, err := uc.DeleteTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error with force, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

// =============================================================================
// ListTemplates Tests
// =============================================================================

func TestListTemplates_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create templates
	for i := 0; i < 5; i++ {
		template, _ := createTestTemplate(fixtures.tenantID, "template_"+string(rune('a'+i)), "Template "+string(rune('A'+i)), domain.ChannelEmail)
		repo.templates[template.ID] = template
		repo.templateByCode[template.Code] = template
	}

	req := &dto.ListTemplatesRequest{
		TenantID: fixtures.tenantID.String(),
		Page:     1,
		PageSize: 10,
	}

	result, err := uc.ListTemplates(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if len(result.Items) != 5 {
		t.Errorf("expected 5 items, got %d", len(result.Items))
	}

	if result.TotalCount != 5 {
		t.Errorf("expected total count 5, got %d", result.TotalCount)
	}
}

func TestListTemplates_WithChannelFilter(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create email template
	emailTemplate, _ := createTestTemplate(fixtures.tenantID, "email_template", "Email Template", domain.ChannelEmail)
	repo.templates[emailTemplate.ID] = emailTemplate
	repo.templateByCode[emailTemplate.Code] = emailTemplate

	// Create SMS template
	smsTemplate, _ := createTestTemplate(fixtures.tenantID, "sms_template", "SMS Template", domain.ChannelSMS)
	repo.templates[smsTemplate.ID] = smsTemplate
	repo.templateByCode[smsTemplate.Code] = smsTemplate

	req := &dto.ListTemplatesRequest{
		TenantID: fixtures.tenantID.String(),
		Channel:  "email",
		Page:     1,
		PageSize: 10,
	}

	result, err := uc.ListTemplates(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// The mock doesn't filter by channel, but verify no error
	if result == nil {
		t.Fatal("expected result, got nil")
	}
}

func TestListTemplates_EmptyResult(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.ListTemplatesRequest{
		TenantID: fixtures.tenantID.String(),
		Page:     1,
		PageSize: 10,
	}

	result, err := uc.ListTemplates(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if len(result.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(result.Items))
	}
}

// =============================================================================
// PublishTemplate Tests
// =============================================================================

func TestPublishTemplate_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template with draft
	template, _ := createTestTemplate(fixtures.tenantID, "draft_template", "Draft Template", domain.ChannelEmail)
	template.SaveDraft(nil)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.PublishTemplateRequest{
		TenantID:    fixtures.tenantID.String(),
		TemplateID:  template.ID.String(),
		PublishedBy: fixtures.userID.String(),
	}

	resp, err := uc.PublishTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Message != "Template published successfully" {
		t.Errorf("expected success message, got %s", resp.Message)
	}
}

func TestPublishTemplate_NotFound(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.PublishTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: uuid.New().String(),
	}

	_, err := uc.PublishTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for not found template")
	}
}

func TestPublishTemplate_ArchivedTemplate(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create archived template
	template, _ := createTestTemplate(fixtures.tenantID, "archived_template", "Archived Template", domain.ChannelEmail)
	now := time.Now().UTC()
	template.DeletedAt = &now
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.PublishTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
	}

	_, err := uc.PublishTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for archived template")
	}
}

// =============================================================================
// ArchiveTemplate Tests
// =============================================================================

func TestArchiveTemplate_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "test_template", "Test Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.ArchiveTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
	}

	resp, err := uc.ArchiveTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Message != "Template archived successfully" {
		t.Errorf("expected success message, got %s", resp.Message)
	}

	// Verify archive
	archived := repo.templates[template.ID]
	if archived.DeletedAt == nil {
		t.Error("expected template to be archived (DeletedAt set)")
	}
}

func TestArchiveTemplate_AlreadyArchived(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create already archived template
	template, _ := createTestTemplate(fixtures.tenantID, "archived_template", "Archived Template", domain.ChannelEmail)
	now := time.Now().UTC()
	template.DeletedAt = &now
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.ArchiveTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
	}

	_, err := uc.ArchiveTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for already archived template")
	}
}

// =============================================================================
// RestoreTemplate Tests
// =============================================================================

func TestRestoreTemplate_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create archived template
	template, _ := createTestTemplate(fixtures.tenantID, "archived_template", "Archived Template", domain.ChannelEmail)
	now := time.Now().UTC()
	template.DeletedAt = &now
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.RestoreTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		RestoredBy: fixtures.userID.String(),
	}

	resp, err := uc.RestoreTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Message != "Template restored successfully" {
		t.Errorf("expected success message, got %s", resp.Message)
	}
}

func TestRestoreTemplate_NotArchived(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create non-archived template
	template, _ := createTestTemplate(fixtures.tenantID, "active_template", "Active Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.RestoreTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
	}

	_, err := uc.RestoreTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for non-archived template")
	}
}

// =============================================================================
// CloneTemplate Tests
// =============================================================================

func TestCloneTemplate_Success(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create source template
	source, _ := createTestTemplate(fixtures.tenantID, "source_template", "Source Template", domain.ChannelEmail)
	source.Description = "Original description"
	source.Category = "marketing"
	source.AddTag("important")
	repo.templates[source.ID] = source
	repo.templateByCode[source.Code] = source

	req := &dto.CloneTemplateRequest{
		TenantID:         fixtures.tenantID.String(),
		SourceTemplateID: source.ID.String(),
		NewCode:          "cloned_template",
		NewName:          "Cloned Template",
		ClonedBy:         fixtures.userID.String(),
	}

	resp, err := uc.CloneTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Name != "Cloned Template" {
		t.Errorf("expected name 'Cloned Template', got %s", resp.Name)
	}

	if resp.Message != "Template cloned successfully" {
		t.Errorf("expected success message, got %s", resp.Message)
	}
}

func TestCloneTemplate_SourceNotFound(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.CloneTemplateRequest{
		TenantID:         fixtures.tenantID.String(),
		SourceTemplateID: uuid.New().String(),
		NewCode:          "cloned_template",
		NewName:          "Cloned Template",
	}

	_, err := uc.CloneTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for not found source template")
	}
}

func TestCloneTemplate_DuplicateNewCode(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create source template
	source, _ := createTestTemplate(fixtures.tenantID, "source_template", "Source Template", domain.ChannelEmail)
	repo.templates[source.ID] = source
	repo.templateByCode[source.Code] = source

	// Create existing template with target code
	existing, _ := createTestTemplate(fixtures.tenantID, "existing_template", "Existing Template", domain.ChannelEmail)
	repo.templates[existing.ID] = existing
	repo.templateByCode[existing.Code] = existing

	req := &dto.CloneTemplateRequest{
		TenantID:         fixtures.tenantID.String(),
		SourceTemplateID: source.ID.String(),
		NewCode:          "existing_template",
		NewName:          "Clone Attempt",
	}

	_, err := uc.CloneTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for duplicate new code")
	}
}

// =============================================================================
// RenderTemplate Tests
// =============================================================================

func TestRenderTemplate_EmailSuccess(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "render_template", "Render Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.RenderTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Channel:    "email",
		Variables: map[string]interface{}{
			"Name": "John Doe",
		},
	}

	resp, err := uc.RenderTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if !containsString(resp.Subject, "John Doe") {
		t.Errorf("expected subject to contain 'John Doe', got: %s", resp.Subject)
	}

	if !containsString(resp.Body, "John Doe") {
		t.Errorf("expected body to contain 'John Doe', got: %s", resp.Body)
	}
}

func TestRenderTemplate_SMSSuccess(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create SMS template
	template, _ := createTestTemplate(fixtures.tenantID, "sms_template", "SMS Template", domain.ChannelSMS)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.RenderTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Channel:    "sms",
		Variables: map[string]interface{}{
			"Name": "Jane Doe",
		},
	}

	resp, err := uc.RenderTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if !containsString(resp.Body, "Jane Doe") {
		t.Errorf("expected body to contain 'Jane Doe', got: %s", resp.Body)
	}
}

func TestRenderTemplate_NotFound(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.RenderTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: uuid.New().String(),
		Channel:    "email",
		Variables:  map[string]interface{}{},
	}

	_, err := uc.RenderTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for not found template")
	}
}

func TestRenderTemplate_NoRenderableContent(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template without any channel content
	template, _ := domain.NewNotificationTemplate(fixtures.tenantID, "empty_template", "Empty Template", domain.TypeTransactional)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.RenderTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Channel:    "webhook", // No webhook template content
		Variables:  map[string]interface{}{},
	}

	_, err := uc.RenderTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error for no renderable content")
	}
}

// =============================================================================
// ValidateTemplate Tests
// =============================================================================

func TestValidateTemplate_ValidSyntax(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.ValidateTemplateRequest{
		TenantID: fixtures.tenantID.String(),
		Channel:  "email",
		Subject:  "Welcome {{.Name}}",
		Body:     "Hello {{.Name}}, welcome to {{.Company}}!",
		HTMLBody: "<p>Hello {{.Name}}, welcome to {{.Company}}!</p>",
	}

	resp, err := uc.ValidateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if !resp.Valid {
		t.Error("expected template to be valid")
	}

	if len(resp.Errors) > 0 {
		t.Errorf("expected no errors, got: %v", resp.Errors)
	}

	if len(resp.DetectedVariables) == 0 {
		t.Error("expected detected variables")
	}
}

func TestValidateTemplate_InvalidSyntax(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.ValidateTemplateRequest{
		TenantID: fixtures.tenantID.String(),
		Channel:  "email",
		Subject:  "Test",
		Body:     "Hello {{.Name}", // Missing closing braces
	}

	resp, err := uc.ValidateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Valid {
		t.Error("expected template to be invalid")
	}

	if len(resp.Errors) == 0 {
		t.Error("expected validation errors")
	}
}

func TestValidateTemplate_UnusedVariables(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.ValidateTemplateRequest{
		TenantID: fixtures.tenantID.String(),
		Channel:  "email",
		Subject:  "Test",
		Body:     "Hello {{.Name}}",
		Variables: map[string]interface{}{
			"Name":    "John",
			"Company": "Acme", // Not used in template
		},
	}

	resp, err := uc.ValidateTemplate(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if len(resp.Warnings) == 0 {
		t.Error("expected warnings for unused variables")
	}
}

// =============================================================================
// Table-Driven Validation Tests
// =============================================================================

func TestCreateTemplateRequest_Validation(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	tests := []struct {
		name        string
		req         *dto.CreateTemplateRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid email template",
			req: &dto.CreateTemplateRequest{
				TenantID:      fixtures.tenantID.String(),
				Code:          "valid_template",
				Name:          "Valid Template",
				Channel:       "email",
				Type:          "transactional",
				Subject:       "Test Subject",
				Body:          "Test Body",
				DefaultLocale: "en",
			},
			wantErr: false,
		},
		{
			name: "valid SMS template without subject",
			req: &dto.CreateTemplateRequest{
				TenantID:      fixtures.tenantID.String(),
				Code:          "sms_template",
				Name:          "SMS Template",
				Channel:       "sms",
				Type:          "transactional",
				Body:          "Test Body",
				DefaultLocale: "en",
			},
			wantErr: false,
		},
		{
			name: "empty tenant ID",
			req: &dto.CreateTemplateRequest{
				TenantID: "",
				Code:     "test",
				Name:     "Test",
				Channel:  "email",
				Type:     "transactional",
				Subject:  "Test",
				Body:     "Body",
			},
			wantErr:     true,
			errContains: "tenant_id",
		},
		{
			name: "empty code",
			req: &dto.CreateTemplateRequest{
				TenantID: fixtures.tenantID.String(),
				Code:     "",
				Name:     "Test",
				Channel:  "email",
				Type:     "transactional",
				Subject:  "Test",
				Body:     "Body",
			},
			wantErr:     true,
			errContains: "code",
		},
		{
			name: "empty name",
			req: &dto.CreateTemplateRequest{
				TenantID: fixtures.tenantID.String(),
				Code:     "test_code",
				Name:     "",
				Channel:  "email",
				Type:     "transactional",
				Subject:  "Test",
				Body:     "Body",
			},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "invalid channel",
			req: &dto.CreateTemplateRequest{
				TenantID: fixtures.tenantID.String(),
				Code:     "test_code",
				Name:     "Test",
				Channel:  "invalid_channel",
				Type:     "transactional",
				Subject:  "Test",
				Body:     "Body",
			},
			wantErr:     true,
			errContains: "channel",
		},
		{
			name: "invalid type",
			req: &dto.CreateTemplateRequest{
				TenantID: fixtures.tenantID.String(),
				Code:     "test_code",
				Name:     "Test",
				Channel:  "email",
				Type:     "invalid_type",
				Subject:  "Test",
				Body:     "Body",
			},
			wantErr:     true,
			errContains: "type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.CreateTemplate(ctx, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', got: %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// =============================================================================
// Context Timeout Tests
// =============================================================================

func TestCreateTemplate_ContextTimeout(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()

	// Set up repo to return error simulating timeout
	repo.createErr = errors.New("context deadline exceeded")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(2 * time.Millisecond)

	req := createValidCreateTemplateRequest(fixtures.tenantID, fixtures.userID)
	req.Code = "timeout_test"

	_, err := uc.CreateTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error due to context timeout")
	}
}

func TestGetTemplate_ContextCancellation(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "test_template", "Test Template", domain.ChannelEmail)
	repo.templates[template.ID] = template

	// Set up repo to return error
	repo.findByIDErr = errors.New("context canceled")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &dto.GetTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
	}

	_, err := uc.GetTemplate(ctx, req)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkCreateTemplate(b *testing.B) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &dto.CreateTemplateRequest{
			TenantID:      fixtures.tenantID.String(),
			Code:          "benchmark_template_" + uuid.New().String()[:8],
			Name:          "Benchmark Template",
			Channel:       "email",
			Type:          "transactional",
			Subject:       "Test Subject",
			Body:          "Test Body",
			DefaultLocale: "en",
		}
		_, _ = uc.CreateTemplate(ctx, req)
	}
}

func BenchmarkGetTemplate(b *testing.B) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "benchmark_template", "Benchmark Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.GetTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetTemplate(ctx, req)
	}
}

func BenchmarkListTemplates(b *testing.B) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create multiple templates
	for i := 0; i < 100; i++ {
		template, _ := createTestTemplate(fixtures.tenantID, "template_"+uuid.New().String()[:8], "Template", domain.ChannelEmail)
		repo.templates[template.ID] = template
		repo.templateByCode[template.Code] = template
	}

	req := &dto.ListTemplatesRequest{
		TenantID: fixtures.tenantID.String(),
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.ListTemplates(ctx, req)
	}
}

func BenchmarkRenderTemplate(b *testing.B) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "render_benchmark", "Render Benchmark", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	req := &dto.RenderTemplateRequest{
		TenantID:   fixtures.tenantID.String(),
		TemplateID: template.ID.String(),
		Channel:    "email",
		Variables: map[string]interface{}{
			"Name":    "John Doe",
			"Company": "Acme Inc",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.RenderTemplate(ctx, req)
	}
}

func BenchmarkValidateTemplate(b *testing.B) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	req := &dto.ValidateTemplateRequest{
		TenantID: fixtures.tenantID.String(),
		Channel:  "email",
		Subject:  "Welcome {{.Name}} to {{.Company}}",
		Body:     "Hello {{.Name}}, welcome to {{.Company}}! Your account ID is {{.AccountID}}.",
		HTMLBody: "<p>Hello {{.Name}}, welcome to {{.Company}}!</p><p>Your account ID is {{.AccountID}}.</p>",
		Variables: map[string]interface{}{
			"Name":      "John",
			"Company":   "Acme",
			"AccountID": "12345",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.ValidateTemplate(ctx, req)
	}
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestCreateTemplate_Concurrent(t *testing.T) {
	uc, _, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := &dto.CreateTemplateRequest{
				TenantID:      fixtures.tenantID.String(),
				Code:          "concurrent_template_" + uuid.New().String()[:8],
				Name:          "Concurrent Template",
				Channel:       "email",
				Type:          "transactional",
				Subject:       "Test Subject",
				Body:          "Test Body",
				DefaultLocale: "en",
			}
			_, err := uc.CreateTemplate(ctx, req)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent create failed: %v", err)
	}
}

func TestGetTemplate_Concurrent(t *testing.T) {
	uc, repo, fixtures := createTemplateTestUseCase()
	ctx := context.Background()

	// Create template
	template, _ := createTestTemplate(fixtures.tenantID, "concurrent_get", "Concurrent Get Template", domain.ChannelEmail)
	repo.templates[template.ID] = template
	repo.templateByCode[template.Code] = template

	const numGoroutines = 50
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := &dto.GetTemplateRequest{
				TenantID:   fixtures.tenantID.String(),
				TemplateID: template.ID.String(),
			}
			_, err := uc.GetTemplate(ctx, req)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent get failed: %v", err)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
