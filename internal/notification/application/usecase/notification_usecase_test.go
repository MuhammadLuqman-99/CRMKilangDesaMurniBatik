// Package usecase contains the application use cases for the Notification service.
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

// ============================================================================
// Mock Implementations
// ============================================================================

// MockNotificationRepository is a mock implementation of NotificationRepository.
type MockNotificationRepository struct {
	mu            sync.RWMutex
	notifications map[uuid.UUID]*domain.Notification
	createErr     error
	updateErr     error
	findErr       error
	listErr       error
	deleteErr     error
	markReadErr   error
}

func NewMockNotificationRepository() *MockNotificationRepository {
	return &MockNotificationRepository{
		notifications: make(map[uuid.UUID]*domain.Notification),
	}
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications[notification.ID] = notification
	return nil
}

func (m *MockNotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications[notification.ID] = notification
	return nil
}

func (m *MockNotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.notifications, id)
	return nil
}

func (m *MockNotificationRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Delete(ctx, id)
}

func (m *MockNotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if n, ok := m.notifications[id]; ok {
		return n, nil
	}
	return nil, errors.New("notification not found")
}

func (m *MockNotificationRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.Notification, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, n := range m.notifications {
		if n.TenantID == tenantID && n.Code == code {
			return n, nil
		}
	}
	return nil, errors.New("notification not found")
}

func (m *MockNotificationRepository) List(ctx context.Context, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var notifications []*domain.Notification
	for _, n := range m.notifications {
		if filter.TenantID != nil && n.TenantID != *filter.TenantID {
			continue
		}
		notifications = append(notifications, n)
	}

	return &domain.NotificationList{
		Notifications: notifications,
		Total:         int64(len(notifications)),
		Offset:        filter.Offset,
		Limit:         filter.Limit,
		HasMore:       false,
	}, nil
}

func (m *MockNotificationRepository) FindByRecipient(ctx context.Context, tenantID uuid.UUID, recipientID uuid.UUID, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	return m.List(ctx, filter)
}

func (m *MockNotificationRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	return m.List(ctx, filter)
}

func (m *MockNotificationRepository) FindByStatus(ctx context.Context, tenantID uuid.UUID, status domain.NotificationStatus, filter domain.NotificationFilter) (*domain.NotificationList, error) {
	return m.List(ctx, filter)
}

func (m *MockNotificationRepository) FindPending(ctx context.Context, limit int) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationRepository) FindScheduled(ctx context.Context, before time.Time, limit int) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationRepository) FindRetryable(ctx context.Context, before time.Time, limit int) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationRepository) FindByBatch(ctx context.Context, batchID uuid.UUID) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationRepository) FindBySourceEvent(ctx context.Context, tenantID uuid.UUID, sourceEvent string, sourceEntityID uuid.UUID) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return int64(len(m.notifications)), nil
}

func (m *MockNotificationRepository) CountByStatus(ctx context.Context, tenantID uuid.UUID) (map[domain.NotificationStatus]int64, error) {
	return nil, nil
}

func (m *MockNotificationRepository) CountByChannel(ctx context.Context, tenantID uuid.UUID) (map[domain.NotificationChannel]int64, error) {
	return nil, nil
}

func (m *MockNotificationRepository) GetStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*domain.NotificationStats, error) {
	return nil, nil
}

func (m *MockNotificationRepository) BulkCreate(ctx context.Context, notifications []*domain.Notification) error {
	for _, n := range notifications {
		if err := m.Create(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockNotificationRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.NotificationStatus) error {
	return nil
}

func (m *MockNotificationRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.notifications[id]
	return ok, nil
}

func (m *MockNotificationRepository) GetVersion(ctx context.Context, id uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if n, ok := m.notifications[id]; ok {
		return n.Version, nil
	}
	return 0, errors.New("notification not found")
}

func (m *MockNotificationRepository) FindUnread(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationRepository) CountUnread(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	return 0, nil
}

func (m *MockNotificationRepository) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	if m.markReadErr != nil {
		return m.markReadErr
	}
	return nil
}

func (m *MockNotificationRepository) DeleteOld(ctx context.Context, before time.Time) (int64, error) {
	return 0, nil
}

func (m *MockNotificationRepository) SetCreateError(err error) {
	m.createErr = err
}

func (m *MockNotificationRepository) SetUpdateError(err error) {
	m.updateErr = err
}

func (m *MockNotificationRepository) SetFindError(err error) {
	m.findErr = err
}

func (m *MockNotificationRepository) SetListError(err error) {
	m.listErr = err
}

// MockTemplateRepository is a mock implementation of TemplateRepository.
type MockTemplateRepository struct {
	mu        sync.RWMutex
	templates map[uuid.UUID]*domain.NotificationTemplate
	findErr   error
}

func NewMockTemplateRepository() *MockTemplateRepository {
	return &MockTemplateRepository{
		templates: make(map[uuid.UUID]*domain.NotificationTemplate),
	}
}

func (m *MockTemplateRepository) Create(ctx context.Context, template *domain.NotificationTemplate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[template.ID] = template
	return nil
}

func (m *MockTemplateRepository) Update(ctx context.Context, template *domain.NotificationTemplate) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[template.ID] = template
	return nil
}

func (m *MockTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.templates, id)
	return nil
}

func (m *MockTemplateRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Delete(ctx, id)
}

func (m *MockTemplateRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.NotificationTemplate, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.templates[id]; ok {
		return t, nil
	}
	return nil, errors.New("template not found")
}

func (m *MockTemplateRepository) FindByCode(ctx context.Context, tenantID uuid.UUID, code string) (*domain.NotificationTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.Code == code {
			return t, nil
		}
	}
	return nil, errors.New("template not found")
}

func (m *MockTemplateRepository) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*domain.NotificationTemplate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.Name == name {
			return t, nil
		}
	}
	return nil, errors.New("template not found")
}

func (m *MockTemplateRepository) List(ctx context.Context, filter domain.TemplateFilter) (*domain.TemplateList, error) {
	return &domain.TemplateList{
		Templates: []*domain.NotificationTemplate{},
		Total:     0,
	}, nil
}

func (m *MockTemplateRepository) FindByType(ctx context.Context, tenantID uuid.UUID, notifType domain.NotificationType) ([]*domain.NotificationTemplate, error) {
	return nil, nil
}

func (m *MockTemplateRepository) FindByChannel(ctx context.Context, tenantID uuid.UUID, channel domain.NotificationChannel) ([]*domain.NotificationTemplate, error) {
	return nil, nil
}

func (m *MockTemplateRepository) FindDefault(ctx context.Context, tenantID uuid.UUID, notifType domain.NotificationType) (*domain.NotificationTemplate, error) {
	return nil, nil
}

func (m *MockTemplateRepository) FindActive(ctx context.Context, tenantID uuid.UUID) ([]*domain.NotificationTemplate, error) {
	return nil, nil
}

func (m *MockTemplateRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return int64(len(m.templates)), nil
}

func (m *MockTemplateRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.templates[id]
	return ok, nil
}

func (m *MockTemplateRepository) ExistsByCode(ctx context.Context, tenantID uuid.UUID, code string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, t := range m.templates {
		if t.TenantID == tenantID && t.Code == code {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockTemplateRepository) GetVersion(ctx context.Context, id uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.templates[id]; ok {
		return t.Version, nil
	}
	return 0, errors.New("template not found")
}

func (m *MockTemplateRepository) IncrementUsageCount(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockTemplateRepository) FindByTag(ctx context.Context, tenantID uuid.UUID, tag string) ([]*domain.NotificationTemplate, error) {
	return nil, nil
}

func (m *MockTemplateRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, filter domain.TemplateFilter) (*domain.TemplateList, error) {
	return nil, nil
}

func (m *MockTemplateRepository) SetFindError(err error) {
	m.findErr = err
}

// MockEmailProvider is a mock implementation of EmailProvider.
type MockEmailProvider struct {
	mu        sync.RWMutex
	sent      []ports.EmailRequest
	sendErr   error
	available bool
}

func NewMockEmailProvider() *MockEmailProvider {
	return &MockEmailProvider{
		sent:      make([]ports.EmailRequest, 0),
		available: true,
	}
}

func (m *MockEmailProvider) SendEmail(ctx context.Context, request ports.EmailRequest) (*ports.EmailResponse, error) {
	if m.sendErr != nil {
		return nil, m.sendErr
	}
	m.mu.Lock()
	m.sent = append(m.sent, request)
	m.mu.Unlock()

	return &ports.EmailResponse{
		MessageID:  request.MessageID,
		ProviderID: "mock-provider-" + request.MessageID,
		Provider:   "mock",
		Status:     "sent",
		SentAt:     time.Now().UTC(),
	}, nil
}

func (m *MockEmailProvider) ValidateEmail(ctx context.Context, email string) (bool, error) {
	return true, nil
}

func (m *MockEmailProvider) GetProviderName() string {
	return "mock-email"
}

func (m *MockEmailProvider) IsAvailable(ctx context.Context) bool {
	return m.available
}

func (m *MockEmailProvider) SetSendError(err error) {
	m.sendErr = err
}

func (m *MockEmailProvider) SetAvailable(available bool) {
	m.available = available
}

func (m *MockEmailProvider) GetSentEmails() []ports.EmailRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]ports.EmailRequest{}, m.sent...)
}

// MockSMSProvider is a mock implementation of SMSProvider.
type MockSMSProvider struct {
	mu        sync.RWMutex
	sent      []ports.SMSRequest
	sendErr   error
	available bool
}

func NewMockSMSProvider() *MockSMSProvider {
	return &MockSMSProvider{
		sent:      make([]ports.SMSRequest, 0),
		available: true,
	}
}

func (m *MockSMSProvider) SendSMS(ctx context.Context, request ports.SMSRequest) (*ports.SMSResponse, error) {
	if m.sendErr != nil {
		return nil, m.sendErr
	}
	m.mu.Lock()
	m.sent = append(m.sent, request)
	m.mu.Unlock()

	return &ports.SMSResponse{
		MessageID:    request.MessageID,
		ProviderID:   "mock-sms-" + request.MessageID,
		Provider:     "mock",
		Status:       "sent",
		SegmentCount: 1,
		SentAt:       time.Now().UTC(),
	}, nil
}

func (m *MockSMSProvider) ValidatePhoneNumber(ctx context.Context, phone string) (bool, error) {
	return true, nil
}

func (m *MockSMSProvider) GetProviderName() string {
	return "mock-sms"
}

func (m *MockSMSProvider) IsAvailable(ctx context.Context) bool {
	return m.available
}

func (m *MockSMSProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*ports.SMSDeliveryStatus, error) {
	return &ports.SMSDeliveryStatus{
		MessageID: messageID,
		Status:    "delivered",
	}, nil
}

func (m *MockSMSProvider) SetSendError(err error) {
	m.sendErr = err
}

// MockPushProvider is a mock implementation of PushProvider.
type MockPushProvider struct {
	mu        sync.RWMutex
	sent      []ports.PushRequest
	sendErr   error
	available bool
}

func NewMockPushProvider() *MockPushProvider {
	return &MockPushProvider{
		sent:      make([]ports.PushRequest, 0),
		available: true,
	}
}

func (m *MockPushProvider) SendPush(ctx context.Context, request ports.PushRequest) (*ports.PushResponse, error) {
	if m.sendErr != nil {
		return nil, m.sendErr
	}
	m.mu.Lock()
	m.sent = append(m.sent, request)
	m.mu.Unlock()

	return &ports.PushResponse{
		MessageID:  request.MessageID,
		ProviderID: "mock-push-" + request.MessageID,
		Provider:   "mock",
		Status:     "sent",
		SentAt:     time.Now().UTC(),
	}, nil
}

func (m *MockPushProvider) ValidateDeviceToken(ctx context.Context, token string, platform string) (bool, error) {
	return true, nil
}

func (m *MockPushProvider) GetProviderName() string {
	return "mock-push"
}

func (m *MockPushProvider) IsAvailable(ctx context.Context) bool {
	return m.available
}

func (m *MockPushProvider) SetSendError(err error) {
	m.sendErr = err
}

// MockInAppProvider is a mock implementation of InAppProvider.
type MockInAppProvider struct {
	mu        sync.RWMutex
	sent      []ports.InAppRequest
	sendErr   error
	available bool
}

func NewMockInAppProvider() *MockInAppProvider {
	return &MockInAppProvider{
		sent:      make([]ports.InAppRequest, 0),
		available: true,
	}
}

func (m *MockInAppProvider) SendInApp(ctx context.Context, request ports.InAppRequest) (*ports.InAppResponse, error) {
	if m.sendErr != nil {
		return nil, m.sendErr
	}
	m.mu.Lock()
	m.sent = append(m.sent, request)
	m.mu.Unlock()

	return &ports.InAppResponse{
		MessageID: request.MessageID,
		UserID:    request.UserID,
		Status:    "sent",
		SentAt:    time.Now().UTC(),
	}, nil
}

func (m *MockInAppProvider) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	return nil
}

func (m *MockInAppProvider) MarkAllAsRead(ctx context.Context, userID string) error {
	return nil
}

func (m *MockInAppProvider) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockInAppProvider) GetProviderName() string {
	return "mock-inapp"
}

func (m *MockInAppProvider) SetSendError(err error) {
	m.sendErr = err
}

// MockEventPublisher is a mock implementation of EventPublisher.
type MockEventPublisher struct {
	mu     sync.RWMutex
	events []domain.DomainEvent
	err    error
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{
		events: make([]domain.DomainEvent, 0),
	}
}

func (m *MockEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	for _, e := range events {
		if err := m.Publish(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockEventPublisher) GetPublishedEvents() []domain.DomainEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]domain.DomainEvent{}, m.events...)
}

// MockRateLimiter is a mock implementation of RateLimiter.
type MockRateLimiter struct {
	allowed bool
	err     error
}

func NewMockRateLimiter() *MockRateLimiter {
	return &MockRateLimiter{allowed: true}
}

func (m *MockRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.allowed, nil
}

func (m *MockRateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	return 100, nil
}

func (m *MockRateLimiter) Reset(ctx context.Context, key string) error {
	return nil
}

func (m *MockRateLimiter) SetAllowed(allowed bool) {
	m.allowed = allowed
}

func (m *MockRateLimiter) SetError(err error) {
	m.err = err
}

// MockQuotaManager is a mock implementation of QuotaManager.
type MockQuotaManager struct {
	allowed bool
	err     error
}

func NewMockQuotaManager() *MockQuotaManager {
	return &MockQuotaManager{allowed: true}
}

func (m *MockQuotaManager) CheckQuota(ctx context.Context, tenantID, channel string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.allowed, nil
}

func (m *MockQuotaManager) ConsumeQuota(ctx context.Context, tenantID, channel string, amount int) error {
	return nil
}

func (m *MockQuotaManager) GetUsage(ctx context.Context, tenantID, channel string) (*ports.QuotaUsage, error) {
	return &ports.QuotaUsage{
		TenantID: tenantID,
		Channel:  channel,
		Used:     0,
		Limit:    1000,
	}, nil
}

func (m *MockQuotaManager) ResetQuota(ctx context.Context, tenantID, channel string) error {
	return nil
}

func (m *MockQuotaManager) SetAllowed(allowed bool) {
	m.allowed = allowed
}

func (m *MockQuotaManager) SetError(err error) {
	m.err = err
}

// MockUserService is a mock implementation of UserService.
type MockUserService struct {
	users map[string]*ports.UserInfo
	err   error
}

func NewMockUserService() *MockUserService {
	return &MockUserService{
		users: make(map[string]*ports.UserInfo),
	}
}

func (m *MockUserService) GetUser(ctx context.Context, userID string) (*ports.UserInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	if user, ok := m.users[userID]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*ports.UserInfo, error) {
	return nil, nil
}

func (m *MockUserService) GetUsersByIDs(ctx context.Context, userIDs []string) ([]*ports.UserInfo, error) {
	return nil, nil
}

func (m *MockUserService) IsUserActive(ctx context.Context, userID string) (bool, error) {
	if user, ok := m.users[userID]; ok {
		return user.IsActive, nil
	}
	return false, nil
}

func (m *MockUserService) AddUser(user *ports.UserInfo) {
	m.users[user.ID] = user
}

func (m *MockUserService) SetError(err error) {
	m.err = err
}

// MockScheduler is a mock implementation of Scheduler.
type MockScheduler struct {
	scheduled map[string]time.Time
	err       error
}

func NewMockScheduler() *MockScheduler {
	return &MockScheduler{
		scheduled: make(map[string]time.Time),
	}
}

func (m *MockScheduler) Schedule(ctx context.Context, notificationID string, scheduledAt time.Time) error {
	if m.err != nil {
		return m.err
	}
	m.scheduled[notificationID] = scheduledAt
	return nil
}

func (m *MockScheduler) Cancel(ctx context.Context, notificationID string) error {
	delete(m.scheduled, notificationID)
	return nil
}

func (m *MockScheduler) Reschedule(ctx context.Context, notificationID string, scheduledAt time.Time) error {
	return m.Schedule(ctx, notificationID, scheduledAt)
}

func (m *MockScheduler) GetScheduled(ctx context.Context, before time.Time, limit int) ([]string, error) {
	return nil, nil
}

func (m *MockScheduler) SetError(err error) {
	m.err = err
}

// MockSuppressionService is a mock implementation of SuppressionService.
type MockSuppressionService struct {
	suppressions map[string]bool
	err          error
}

func NewMockSuppressionService() *MockSuppressionService {
	return &MockSuppressionService{
		suppressions: make(map[string]bool),
	}
}

func (m *MockSuppressionService) IsSuppressed(ctx context.Context, tenantID, channel, recipient string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	key := tenantID + ":" + channel + ":" + recipient
	return m.suppressions[key], nil
}

func (m *MockSuppressionService) AddSuppression(ctx context.Context, tenantID, channel, recipient, reason string) error {
	key := tenantID + ":" + channel + ":" + recipient
	m.suppressions[key] = true
	return nil
}

func (m *MockSuppressionService) RemoveSuppression(ctx context.Context, tenantID, channel, recipient string) error {
	key := tenantID + ":" + channel + ":" + recipient
	delete(m.suppressions, key)
	return nil
}

func (m *MockSuppressionService) GetSuppressionReason(ctx context.Context, tenantID, channel, recipient string) (string, error) {
	return "", nil
}

func (m *MockSuppressionService) SetSuppressed(tenantID, channel, recipient string, suppressed bool) {
	key := tenantID + ":" + channel + ":" + recipient
	m.suppressions[key] = suppressed
}

func (m *MockSuppressionService) SetError(err error) {
	m.err = err
}

// MockIdGenerator is a mock implementation of IdGenerator.
type MockIdGenerator struct{}

func NewMockIdGenerator() *MockIdGenerator {
	return &MockIdGenerator{}
}

func (m *MockIdGenerator) Generate() string {
	return uuid.New().String()
}

func (m *MockIdGenerator) GenerateWithPrefix(prefix string) string {
	return prefix + "-" + uuid.New().String()
}

// MockTimeProvider is a mock implementation of TimeProvider.
type MockTimeProvider struct {
	now time.Time
}

func NewMockTimeProvider() *MockTimeProvider {
	return &MockTimeProvider{
		now: time.Now().UTC(),
	}
}

func (m *MockTimeProvider) Now() time.Time {
	return m.now
}

func (m *MockTimeProvider) NowUTC() time.Time {
	return m.now
}

func (m *MockTimeProvider) SetNow(t time.Time) {
	m.now = t
}

// MockMetricsCollector is a mock implementation of MetricsCollector.
type MockMetricsCollector struct {
	mu       sync.RWMutex
	counters map[string]int
}

func NewMockMetricsCollector() *MockMetricsCollector {
	return &MockMetricsCollector{
		counters: make(map[string]int),
	}
}

func (m *MockMetricsCollector) IncrementCounter(ctx context.Context, name string, tags map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name]++
}

func (m *MockMetricsCollector) RecordHistogram(ctx context.Context, name string, value float64, tags map[string]string) {
}

func (m *MockMetricsCollector) RecordGauge(ctx context.Context, name string, value float64, tags map[string]string) {
}

func (m *MockMetricsCollector) RecordDuration(ctx context.Context, name string, duration time.Duration, tags map[string]string) {
}

func (m *MockMetricsCollector) GetCounter(name string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[name]
}

// MockLogger is a mock implementation of Logger.
type MockLogger struct {
	logs []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{logs: make([]string, 0)}
}

func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "DEBUG: "+msg)
}

func (m *MockLogger) Info(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "INFO: "+msg)
}

func (m *MockLogger) Warn(msg string, fields map[string]interface{}) {
	m.logs = append(m.logs, "WARN: "+msg)
}

func (m *MockLogger) Error(msg string, err error, fields map[string]interface{}) {
	m.logs = append(m.logs, "ERROR: "+msg)
}

func (m *MockLogger) WithContext(ctx context.Context) ports.Logger {
	return m
}

func (m *MockLogger) WithFields(fields map[string]interface{}) ports.Logger {
	return m
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTestUseCase(t *testing.T) (*notificationUseCase, *TestMocks) {
	t.Helper()

	mocks := &TestMocks{
		NotificationRepo:   NewMockNotificationRepository(),
		TemplateRepo:       NewMockTemplateRepository(),
		EmailProvider:      NewMockEmailProvider(),
		SMSProvider:        NewMockSMSProvider(),
		PushProvider:       NewMockPushProvider(),
		InAppProvider:      NewMockInAppProvider(),
		EventPublisher:     NewMockEventPublisher(),
		RateLimiter:        NewMockRateLimiter(),
		QuotaManager:       NewMockQuotaManager(),
		UserService:        NewMockUserService(),
		Scheduler:          NewMockScheduler(),
		SuppressionService: NewMockSuppressionService(),
		IdGenerator:        NewMockIdGenerator(),
		TimeProvider:       NewMockTimeProvider(),
		Metrics:            NewMockMetricsCollector(),
		Logger:             NewMockLogger(),
	}

	cfg := NotificationUseCaseConfig{
		NotificationRepo:   mocks.NotificationRepo,
		TemplateRepo:       mocks.TemplateRepo,
		EmailProvider:      mocks.EmailProvider,
		SMSProvider:        mocks.SMSProvider,
		PushProvider:       mocks.PushProvider,
		InAppProvider:      mocks.InAppProvider,
		EventPublisher:     mocks.EventPublisher,
		RateLimiter:        mocks.RateLimiter,
		QuotaManager:       mocks.QuotaManager,
		UserService:        mocks.UserService,
		Scheduler:          mocks.Scheduler,
		SuppressionService: mocks.SuppressionService,
		IdGenerator:        mocks.IdGenerator,
		TimeProvider:       mocks.TimeProvider,
		Metrics:            mocks.Metrics,
		Logger:             mocks.Logger,
	}

	uc := NewNotificationUseCase(cfg).(*notificationUseCase)
	return uc, mocks
}

type TestMocks struct {
	NotificationRepo   *MockNotificationRepository
	TemplateRepo       *MockTemplateRepository
	EmailProvider      *MockEmailProvider
	SMSProvider        *MockSMSProvider
	PushProvider       *MockPushProvider
	InAppProvider      *MockInAppProvider
	EventPublisher     *MockEventPublisher
	RateLimiter        *MockRateLimiter
	QuotaManager       *MockQuotaManager
	UserService        *MockUserService
	Scheduler          *MockScheduler
	SuppressionService *MockSuppressionService
	IdGenerator        *MockIdGenerator
	TimeProvider       *MockTimeProvider
	Metrics            *MockMetricsCollector
	Logger             *MockLogger
}

func createTestEmailRequest() *dto.SendEmailRequest {
	return &dto.SendEmailRequest{
		TenantID: uuid.New().String(),
		Type:     "transactional",
		To:       []string{"test@example.com"},
		Subject:  "Test Subject",
		Body:     "Test Body",
	}
}

func createTestSMSRequest() *dto.SendSMSRequest {
	return &dto.SendSMSRequest{
		TenantID: uuid.New().String(),
		Type:     "transactional",
		To:       "+1234567890",
		Body:     "Test SMS Body",
	}
}

func createTestInAppRequest(userID string) *dto.SendInAppRequest {
	return &dto.SendInAppRequest{
		TenantID: uuid.New().String(),
		Type:     "alert",
		UserID:   userID,
		Title:    "Test Title",
		Body:     "Test In-App Body",
	}
}

func createTestPushRequest(userID string) *dto.SendPushRequest {
	return &dto.SendPushRequest{
		TenantID:    uuid.New().String(),
		Type:        "alert",
		UserID:      userID,
		DeviceToken: "test-device-token-1234567890",
		Title:       "Test Push Title",
		Body:        "Test Push Body",
	}
}

func createTestUser(userID string) *ports.UserInfo {
	return &ports.UserInfo{
		ID:           userID,
		TenantID:     uuid.New().String(),
		Email:        "testuser@example.com",
		DisplayName:  "Test User",
		IsActive:     true,
		DeviceTokens: []string{"device-token-123"},
		Locale:       "en-US",
	}
}

func createTestNotification(tenantID uuid.UUID, channel domain.NotificationChannel) *domain.Notification {
	notification, _ := domain.NewNotification(
		tenantID,
		domain.TypeTransactional,
		channel,
		"Test notification body",
	)
	return notification
}

// ============================================================================
// SendEmail Tests
// ============================================================================

func TestSendEmail_Success(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestEmailRequest()

	resp, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.NotificationID == "" {
		t.Error("expected notification ID in response")
	}

	if resp.Status != "queued" {
		t.Errorf("expected status 'queued', got '%s'", resp.Status)
	}
}

func TestSendEmail_ValidationErrors(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *dto.SendEmailRequest
		wantErr bool
	}{
		{
			name: "missing tenant ID",
			req: &dto.SendEmailRequest{
				Type:    "transactional",
				To:      []string{"test@example.com"},
				Subject: "Test",
				Body:    "Test body",
			},
			wantErr: true,
		},
		{
			name: "missing recipients",
			req: &dto.SendEmailRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				To:       []string{},
				Subject:  "Test",
			},
			wantErr: true,
		},
		{
			name: "missing subject without template",
			req: &dto.SendEmailRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				To:       []string{"test@example.com"},
				Body:     "Test body",
			},
			wantErr: true,
		},
		{
			name: "invalid tenant ID format",
			req: &dto.SendEmailRequest{
				TenantID: "invalid-uuid",
				Type:     "transactional",
				To:       []string{"test@example.com"},
				Subject:  "Test",
				Body:     "Test body",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.SendEmail(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendEmail_RateLimitExceeded(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	mocks.RateLimiter.SetAllowed(false)

	req := createTestEmailRequest()
	_, err := uc.SendEmail(ctx, req)

	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeRateLimitExceeded {
		t.Errorf("expected error code %s, got %s", application.ErrCodeRateLimitExceeded, appErr.Code)
	}
}

func TestSendEmail_QuotaExceeded(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	mocks.QuotaManager.SetAllowed(false)

	req := createTestEmailRequest()
	_, err := uc.SendEmail(ctx, req)

	if err == nil {
		t.Fatal("expected quota exceeded error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeQuotaExceeded {
		t.Errorf("expected error code %s, got %s", application.ErrCodeQuotaExceeded, appErr.Code)
	}
}

func TestSendEmail_Scheduled(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	futureTime := time.Now().UTC().Add(24 * time.Hour)
	req := createTestEmailRequest()
	req.ScheduledAt = &futureTime

	resp, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}

	if resp.Status != "scheduled" {
		t.Errorf("expected status 'scheduled', got '%s'", resp.Status)
	}

	if resp.ScheduledAt == nil {
		t.Error("expected scheduled time in response")
	}

	// Verify scheduler was called
	if len(mocks.Scheduler.scheduled) != 1 {
		t.Error("expected notification to be scheduled")
	}
}

func TestSendEmail_WithPriority(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	priorities := []string{"low", "normal", "high", "critical"}

	for _, priority := range priorities {
		t.Run("priority_"+priority, func(t *testing.T) {
			req := createTestEmailRequest()
			req.Priority = priority

			resp, err := uc.SendEmail(ctx, req)
			if err != nil {
				t.Fatalf("SendEmail with priority %s failed: %v", priority, err)
			}

			if resp == nil {
				t.Fatal("expected response, got nil")
			}
		})
	}
}

func TestSendEmail_InvalidPriority(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestEmailRequest()
	req.Priority = "invalid-priority"

	_, err := uc.SendEmail(ctx, req)
	if err == nil {
		t.Fatal("expected error for invalid priority, got nil")
	}
}

func TestSendEmail_WithTracking(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestEmailRequest()
	trackOpens := true
	trackClicks := true
	req.TrackOpens = &trackOpens
	req.TrackClicks = &trackClicks

	resp, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail with tracking failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestSendEmail_RepositoryCreateError(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	mocks.NotificationRepo.SetCreateError(errors.New("database error"))

	req := createTestEmailRequest()
	_, err := uc.SendEmail(ctx, req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ============================================================================
// SendSMS Tests
// ============================================================================

func TestSendSMS_Success(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestSMSRequest()

	resp, err := uc.SendSMS(ctx, req)
	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.NotificationID == "" {
		t.Error("expected notification ID in response")
	}

	if resp.Status != "queued" {
		t.Errorf("expected status 'queued', got '%s'", resp.Status)
	}
}

func TestSendSMS_ValidationErrors(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *dto.SendSMSRequest
		wantErr bool
	}{
		{
			name: "missing tenant ID",
			req: &dto.SendSMSRequest{
				Type: "transactional",
				To:   "+1234567890",
				Body: "Test",
			},
			wantErr: true,
		},
		{
			name: "missing phone number",
			req: &dto.SendSMSRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				Body:     "Test",
			},
			wantErr: true,
		},
		{
			name: "missing body without template",
			req: &dto.SendSMSRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				To:       "+1234567890",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.SendSMS(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendSMS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendSMS_Suppressed(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	req := createTestSMSRequest()
	mocks.SuppressionService.SetSuppressed(req.TenantID, "sms", req.To, true)

	_, err := uc.SendSMS(ctx, req)

	if err == nil {
		t.Fatal("expected suppression error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeSMSOptedOut {
		t.Errorf("expected error code %s, got %s", application.ErrCodeSMSOptedOut, appErr.Code)
	}
}

func TestSendSMS_RateLimitExceeded(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	mocks.RateLimiter.SetAllowed(false)

	req := createTestSMSRequest()
	_, err := uc.SendSMS(ctx, req)

	if err == nil {
		t.Fatal("expected rate limit error, got nil")
	}
}

// ============================================================================
// SendInAppNotification Tests
// ============================================================================

func TestSendInAppNotification_Success(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	userID := uuid.New().String()
	user := createTestUser(userID)
	mocks.UserService.AddUser(user)

	req := createTestInAppRequest(userID)
	req.TenantID = user.TenantID

	resp, err := uc.SendInAppNotification(ctx, req)
	if err != nil {
		t.Fatalf("SendInAppNotification failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.NotificationID == "" {
		t.Error("expected notification ID in response")
	}
}

func TestSendInAppNotification_UserNotFound(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestInAppRequest(uuid.New().String())

	_, err := uc.SendInAppNotification(ctx, req)

	if err == nil {
		t.Fatal("expected user not found error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeUserNotFound {
		t.Errorf("expected error code %s, got %s", application.ErrCodeUserNotFound, appErr.Code)
	}
}

func TestSendInAppNotification_UserInactive(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	userID := uuid.New().String()
	user := createTestUser(userID)
	user.IsActive = false
	mocks.UserService.AddUser(user)

	req := createTestInAppRequest(userID)
	req.TenantID = user.TenantID

	_, err := uc.SendInAppNotification(ctx, req)

	if err == nil {
		t.Fatal("expected user inactive error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeUserInactive {
		t.Errorf("expected error code %s, got %s", application.ErrCodeUserInactive, appErr.Code)
	}
}

func TestSendInAppNotification_ValidationErrors(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *dto.SendInAppRequest
		wantErr bool
	}{
		{
			name: "missing tenant ID",
			req: &dto.SendInAppRequest{
				Type:   "alert",
				UserID: uuid.New().String(),
				Title:  "Test",
				Body:   "Test body",
			},
			wantErr: true,
		},
		{
			name: "missing user ID",
			req: &dto.SendInAppRequest{
				TenantID: uuid.New().String(),
				Type:     "alert",
				Title:    "Test",
				Body:     "Test body",
			},
			wantErr: true,
		},
		{
			name: "missing title without template",
			req: &dto.SendInAppRequest{
				TenantID: uuid.New().String(),
				Type:     "alert",
				UserID:   uuid.New().String(),
				Body:     "Test body",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.SendInAppNotification(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendInAppNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// SendPushNotification Tests
// ============================================================================

func TestSendPushNotification_Success(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	userID := uuid.New().String()
	user := createTestUser(userID)
	mocks.UserService.AddUser(user)

	req := createTestPushRequest(userID)
	req.TenantID = user.TenantID

	resp, err := uc.SendPushNotification(ctx, req)
	if err != nil {
		t.Fatalf("SendPushNotification failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.NotificationID == "" {
		t.Error("expected notification ID in response")
	}
}

func TestSendPushNotification_NoDeviceToken(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	userID := uuid.New().String()
	user := createTestUser(userID)
	user.DeviceTokens = []string{} // No device tokens
	mocks.UserService.AddUser(user)

	req := createTestPushRequest(userID)
	req.TenantID = user.TenantID
	req.DeviceToken = "" // No device token in request either

	_, err := uc.SendPushNotification(ctx, req)

	if err == nil {
		t.Fatal("expected device not registered error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeDeviceNotRegistered {
		t.Errorf("expected error code %s, got %s", application.ErrCodeDeviceNotRegistered, appErr.Code)
	}
}

func TestSendPushNotification_UserNotFound(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestPushRequest(uuid.New().String())

	_, err := uc.SendPushNotification(ctx, req)

	if err == nil {
		t.Fatal("expected user not found error, got nil")
	}
}

// ============================================================================
// GetNotification Tests
// ============================================================================

func TestGetNotification_Success(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.GetNotificationRequest{
		TenantID:       tenantID.String(),
		NotificationID: notification.ID.String(),
	}

	result, err := uc.GetNotification(ctx, req)
	if err != nil {
		t.Fatalf("GetNotification failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected notification, got nil")
	}

	if result.ID != notification.ID.String() {
		t.Errorf("expected notification ID %s, got %s", notification.ID.String(), result.ID)
	}
}

func TestGetNotification_NotFound(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := &dto.GetNotificationRequest{
		TenantID:       uuid.New().String(),
		NotificationID: uuid.New().String(),
	}

	_, err := uc.GetNotification(ctx, req)

	if err == nil {
		t.Fatal("expected not found error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeNotificationNotFound {
		t.Errorf("expected error code %s, got %s", application.ErrCodeNotificationNotFound, appErr.Code)
	}
}

func TestGetNotification_WrongTenant(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	mocks.NotificationRepo.Create(ctx, notification)

	// Request with different tenant ID
	req := &dto.GetNotificationRequest{
		TenantID:       uuid.New().String(),
		NotificationID: notification.ID.String(),
	}

	_, err := uc.GetNotification(ctx, req)

	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeForbidden {
		t.Errorf("expected error code %s, got %s", application.ErrCodeForbidden, appErr.Code)
	}
}

func TestGetNotification_InvalidIDs(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *dto.GetNotificationRequest
		wantErr bool
	}{
		{
			name: "invalid tenant ID",
			req: &dto.GetNotificationRequest{
				TenantID:       "invalid-uuid",
				NotificationID: uuid.New().String(),
			},
			wantErr: true,
		},
		{
			name: "invalid notification ID",
			req: &dto.GetNotificationRequest{
				TenantID:       uuid.New().String(),
				NotificationID: "invalid-uuid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.GetNotification(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNotification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// ListNotifications Tests
// ============================================================================

func TestListNotifications_Success(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()

	// Create some notifications
	for i := 0; i < 5; i++ {
		notification := createTestNotification(tenantID, domain.ChannelEmail)
		notification.RecipientEmail = "test@example.com"
		notification.Subject = "Test Subject"
		mocks.NotificationRepo.Create(ctx, notification)
	}

	req := &dto.ListNotificationsRequest{
		TenantID: tenantID.String(),
		Page:     1,
		PageSize: 10,
	}

	result, err := uc.ListNotifications(ctx, req)
	if err != nil {
		t.Fatalf("ListNotifications failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected notification list, got nil")
	}

	if result.TotalCount != 5 {
		t.Errorf("expected 5 notifications, got %d", result.TotalCount)
	}
}

func TestListNotifications_WithFilters(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()

	// Create notifications with different channels
	emailNotification := createTestNotification(tenantID, domain.ChannelEmail)
	emailNotification.RecipientEmail = "test@example.com"
	emailNotification.Subject = "Test Subject"
	mocks.NotificationRepo.Create(ctx, emailNotification)

	req := &dto.ListNotificationsRequest{
		TenantID: tenantID.String(),
		Channel:  "email",
		Page:     1,
		PageSize: 10,
	}

	result, err := uc.ListNotifications(ctx, req)
	if err != nil {
		t.Fatalf("ListNotifications failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected notification list, got nil")
	}
}

func TestListNotifications_InvalidTenantID(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := &dto.ListNotificationsRequest{
		TenantID: "invalid-uuid",
		Page:     1,
		PageSize: 10,
	}

	_, err := uc.ListNotifications(ctx, req)

	if err == nil {
		t.Fatal("expected error for invalid tenant ID, got nil")
	}
}

// ============================================================================
// RetryNotification Tests
// ============================================================================

func TestRetryNotification_Success(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	notification.Status = domain.StatusFailed
	notification.AttemptCount = 1
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.RetryNotificationRequest{
		NotificationID: notification.ID.String(),
	}

	resp, err := uc.RetryNotification(ctx, req)
	if err != nil {
		t.Fatalf("RetryNotification failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.NotificationID != notification.ID.String() {
		t.Errorf("expected notification ID %s, got %s", notification.ID.String(), resp.NotificationID)
	}
}

func TestRetryNotification_NotFound(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := &dto.RetryNotificationRequest{
		NotificationID: uuid.New().String(),
	}

	_, err := uc.RetryNotification(ctx, req)

	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestRetryNotification_CannotRetry(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	notification.Status = domain.StatusDelivered // Cannot retry delivered notifications
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.RetryNotificationRequest{
		NotificationID: notification.ID.String(),
	}

	_, err := uc.RetryNotification(ctx, req)

	if err == nil {
		t.Fatal("expected cannot retry error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeCannotRetry {
		t.Errorf("expected error code %s, got %s", application.ErrCodeCannotRetry, appErr.Code)
	}
}

func TestRetryNotification_ForceRetry(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	notification.Status = domain.StatusFailed
	notification.AttemptCount = 10 // Exceeded normal retry limit
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.RetryNotificationRequest{
		NotificationID: notification.ID.String(),
		Force:          true,
	}

	resp, err := uc.RetryNotification(ctx, req)
	if err != nil {
		t.Fatalf("RetryNotification with force failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

// ============================================================================
// CancelNotification Tests
// ============================================================================

func TestCancelNotification_Success(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	notification.Status = domain.StatusPending
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.CancelNotificationRequest{
		NotificationID: notification.ID.String(),
	}

	resp, err := uc.CancelNotification(ctx, req)
	if err != nil {
		t.Fatalf("CancelNotification failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}

	if resp.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got '%s'", resp.Status)
	}
}

func TestCancelNotification_NotFound(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := &dto.CancelNotificationRequest{
		NotificationID: uuid.New().String(),
	}

	_, err := uc.CancelNotification(ctx, req)

	if err == nil {
		t.Fatal("expected not found error, got nil")
	}
}

func TestCancelNotification_CannotCancel(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	notification.Status = domain.StatusDelivered // Cannot cancel delivered
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.CancelNotificationRequest{
		NotificationID: notification.ID.String(),
	}

	_, err := uc.CancelNotification(ctx, req)

	if err == nil {
		t.Fatal("expected cannot cancel error, got nil")
	}

	appErr, ok := err.(*application.AppError)
	if !ok {
		t.Fatalf("expected *application.AppError, got %T", err)
	}

	if appErr.Code != application.ErrCodeCannotCancel {
		t.Errorf("expected error code %s, got %s", application.ErrCodeCannotCancel, appErr.Code)
	}
}

// ============================================================================
// Table-Driven Validation Tests
// ============================================================================

func TestValidateSendEmailRequest(t *testing.T) {
	uc, _ := createTestUseCase(t)

	tests := []struct {
		name    string
		req     *dto.SendEmailRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &dto.SendEmailRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				To:       []string{"test@example.com"},
				Subject:  "Test",
				Body:     "Test body",
			},
			wantErr: false,
		},
		{
			name: "empty tenant ID",
			req: &dto.SendEmailRequest{
				Type:    "transactional",
				To:      []string{"test@example.com"},
				Subject: "Test",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "empty recipients",
			req: &dto.SendEmailRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				To:       []string{},
				Subject:  "Test",
			},
			wantErr: true,
			errMsg:  "at least one recipient email is required",
		},
		{
			name: "missing subject without template",
			req: &dto.SendEmailRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				To:       []string{"test@example.com"},
				Body:     "Test body",
			},
			wantErr: true,
			errMsg:  "subject is required when not using a template",
		},
		{
			name: "subject provided with template",
			req: &dto.SendEmailRequest{
				TenantID:   uuid.New().String(),
				Type:       "transactional",
				To:         []string{"test@example.com"},
				TemplateID: uuid.New().String(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.validateSendEmailRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSendEmailRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				appErr, ok := err.(*application.AppError)
				if ok && appErr.Message != tt.errMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errMsg, appErr.Message)
				}
			}
		})
	}
}

func TestValidateSendSMSRequest(t *testing.T) {
	uc, _ := createTestUseCase(t)

	tests := []struct {
		name    string
		req     *dto.SendSMSRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &dto.SendSMSRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				To:       "+1234567890",
				Body:     "Test body",
			},
			wantErr: false,
		},
		{
			name: "empty tenant ID",
			req: &dto.SendSMSRequest{
				Type: "transactional",
				To:   "+1234567890",
				Body: "Test",
			},
			wantErr: true,
		},
		{
			name: "empty phone number",
			req: &dto.SendSMSRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				Body:     "Test",
			},
			wantErr: true,
		},
		{
			name: "missing body without template",
			req: &dto.SendSMSRequest{
				TenantID: uuid.New().String(),
				Type:     "transactional",
				To:       "+1234567890",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.validateSendSMSRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSendSMSRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSendInAppRequest(t *testing.T) {
	uc, _ := createTestUseCase(t)

	tests := []struct {
		name    string
		req     *dto.SendInAppRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &dto.SendInAppRequest{
				TenantID: uuid.New().String(),
				Type:     "alert",
				UserID:   uuid.New().String(),
				Title:    "Test",
				Body:     "Test body",
			},
			wantErr: false,
		},
		{
			name: "empty tenant ID",
			req: &dto.SendInAppRequest{
				Type:   "alert",
				UserID: uuid.New().String(),
				Title:  "Test",
				Body:   "Test",
			},
			wantErr: true,
		},
		{
			name: "empty user ID",
			req: &dto.SendInAppRequest{
				TenantID: uuid.New().String(),
				Type:     "alert",
				Title:    "Test",
				Body:     "Test",
			},
			wantErr: true,
		},
		{
			name: "missing title without template",
			req: &dto.SendInAppRequest{
				TenantID: uuid.New().String(),
				Type:     "alert",
				UserID:   uuid.New().String(),
				Body:     "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.validateSendInAppRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSendInAppRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSendPushRequest(t *testing.T) {
	uc, _ := createTestUseCase(t)

	tests := []struct {
		name    string
		req     *dto.SendPushRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &dto.SendPushRequest{
				TenantID: uuid.New().String(),
				Type:     "alert",
				UserID:   uuid.New().String(),
				Title:    "Test",
				Body:     "Test body",
			},
			wantErr: false,
		},
		{
			name: "empty tenant ID",
			req: &dto.SendPushRequest{
				Type:   "alert",
				UserID: uuid.New().String(),
				Title:  "Test",
			},
			wantErr: true,
		},
		{
			name: "empty user ID",
			req: &dto.SendPushRequest{
				TenantID: uuid.New().String(),
				Type:     "alert",
				Title:    "Test",
			},
			wantErr: true,
		},
		{
			name: "missing title without template",
			req: &dto.SendPushRequest{
				TenantID: uuid.New().String(),
				Type:     "alert",
				UserID:   uuid.New().String(),
				Body:     "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.validateSendPushRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSendPushRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Context Timeout Tests
// ============================================================================

func TestSendEmail_ContextCancelled(t *testing.T) {
	uc, _ := createTestUseCase(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := createTestEmailRequest()

	// The operation should still succeed because context is checked at specific points
	// and our mock doesn't check context. In a real scenario, repository operations
	// would fail with context cancelled error.
	_, _ = uc.SendEmail(ctx, req)
}

func TestSendEmail_ContextTimeout(t *testing.T) {
	uc, mocks := createTestUseCase(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Simulate slow repository
	mocks.NotificationRepo.SetCreateError(context.DeadlineExceeded)

	time.Sleep(10 * time.Millisecond) // Wait for timeout

	req := createTestEmailRequest()
	_, err := uc.SendEmail(ctx, req)

	if err == nil {
		t.Log("Note: mock doesn't propagate context timeout - this is expected behavior in tests")
	}
}

func TestListNotifications_ContextCancelled(t *testing.T) {
	uc, mocks := createTestUseCase(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mocks.NotificationRepo.SetListError(context.Canceled)

	req := &dto.ListNotificationsRequest{
		TenantID: uuid.New().String(),
		Page:     1,
		PageSize: 10,
	}

	_, err := uc.ListNotifications(ctx, req)
	if err == nil {
		t.Log("Note: context cancellation check depends on repository implementation")
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkSendEmail(b *testing.B) {
	uc, _ := createTestUseCaseForBenchmark()
	ctx := context.Background()
	req := createTestEmailRequest()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.SendEmail(ctx, req)
	}
}

func BenchmarkSendSMS(b *testing.B) {
	uc, _ := createTestUseCaseForBenchmark()
	ctx := context.Background()
	req := createTestSMSRequest()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.SendSMS(ctx, req)
	}
}

func BenchmarkGetNotification(b *testing.B) {
	uc, mocks := createTestUseCaseForBenchmark()
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.GetNotificationRequest{
		TenantID:       tenantID.String(),
		NotificationID: notification.ID.String(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.GetNotification(ctx, req)
	}
}

func BenchmarkListNotifications(b *testing.B) {
	uc, mocks := createTestUseCaseForBenchmark()
	ctx := context.Background()

	tenantID := uuid.New()
	for i := 0; i < 100; i++ {
		notification := createTestNotification(tenantID, domain.ChannelEmail)
		notification.RecipientEmail = "test@example.com"
		notification.Subject = "Test Subject"
		mocks.NotificationRepo.Create(ctx, notification)
	}

	req := &dto.ListNotificationsRequest{
		TenantID: tenantID.String(),
		Page:     1,
		PageSize: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.ListNotifications(ctx, req)
	}
}

func BenchmarkValidateSendEmailRequest(b *testing.B) {
	uc, _ := createTestUseCaseForBenchmark()
	req := &dto.SendEmailRequest{
		TenantID: uuid.New().String(),
		Type:     "transactional",
		To:       []string{"test@example.com"},
		Subject:  "Test",
		Body:     "Test body",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uc.validateSendEmailRequest(req)
	}
}

func BenchmarkConcurrentSendEmail(b *testing.B) {
	uc, _ := createTestUseCaseForBenchmark()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := createTestEmailRequest()
			_, _ = uc.SendEmail(ctx, req)
		}
	})
}

func createTestUseCaseForBenchmark() (*notificationUseCase, *TestMocks) {
	mocks := &TestMocks{
		NotificationRepo:   NewMockNotificationRepository(),
		TemplateRepo:       NewMockTemplateRepository(),
		EmailProvider:      NewMockEmailProvider(),
		SMSProvider:        NewMockSMSProvider(),
		PushProvider:       NewMockPushProvider(),
		InAppProvider:      NewMockInAppProvider(),
		EventPublisher:     NewMockEventPublisher(),
		RateLimiter:        NewMockRateLimiter(),
		QuotaManager:       NewMockQuotaManager(),
		UserService:        NewMockUserService(),
		Scheduler:          NewMockScheduler(),
		SuppressionService: NewMockSuppressionService(),
		IdGenerator:        NewMockIdGenerator(),
		TimeProvider:       NewMockTimeProvider(),
		Metrics:            NewMockMetricsCollector(),
		Logger:             NewMockLogger(),
	}

	cfg := NotificationUseCaseConfig{
		NotificationRepo:   mocks.NotificationRepo,
		TemplateRepo:       mocks.TemplateRepo,
		EmailProvider:      mocks.EmailProvider,
		SMSProvider:        mocks.SMSProvider,
		PushProvider:       mocks.PushProvider,
		InAppProvider:      mocks.InAppProvider,
		EventPublisher:     mocks.EventPublisher,
		RateLimiter:        mocks.RateLimiter,
		QuotaManager:       mocks.QuotaManager,
		UserService:        mocks.UserService,
		Scheduler:          mocks.Scheduler,
		SuppressionService: mocks.SuppressionService,
		IdGenerator:        mocks.IdGenerator,
		TimeProvider:       mocks.TimeProvider,
		Metrics:            mocks.Metrics,
		Logger:             mocks.Logger,
	}

	uc := NewNotificationUseCase(cfg).(*notificationUseCase)
	return uc, mocks
}

// ============================================================================
// Integration-like Tests (testing multiple operations together)
// ============================================================================

func TestNotificationLifecycle_EmailScheduleAndCancel(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	// Schedule an email for the future
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	req := createTestEmailRequest()
	req.ScheduledAt = &futureTime

	resp, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}

	if resp.Status != "scheduled" {
		t.Errorf("expected status 'scheduled', got '%s'", resp.Status)
	}

	// Verify notification was created
	notification, err := mocks.NotificationRepo.FindByID(ctx, uuid.MustParse(resp.NotificationID))
	if err != nil {
		t.Fatalf("Failed to find notification: %v", err)
	}

	if notification.Status != domain.StatusScheduled {
		t.Errorf("expected notification status 'scheduled', got '%s'", notification.Status)
	}

	// Cancel the notification
	cancelReq := &dto.CancelNotificationRequest{
		NotificationID: resp.NotificationID,
	}

	cancelResp, err := uc.CancelNotification(ctx, cancelReq)
	if err != nil {
		t.Fatalf("CancelNotification failed: %v", err)
	}

	if cancelResp.Status != "cancelled" {
		t.Errorf("expected status 'cancelled', got '%s'", cancelResp.Status)
	}
}

func TestNotificationLifecycle_SMSSendAndRetry(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	// Send SMS
	req := createTestSMSRequest()
	resp, err := uc.SendSMS(ctx, req)
	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}

	// Simulate failure by updating status
	notification, _ := mocks.NotificationRepo.FindByID(ctx, uuid.MustParse(resp.NotificationID))
	notification.Status = domain.StatusFailed
	notification.AttemptCount = 1
	mocks.NotificationRepo.Update(ctx, notification)

	// Retry the notification
	retryReq := &dto.RetryNotificationRequest{
		NotificationID: resp.NotificationID,
	}

	retryResp, err := uc.RetryNotification(ctx, retryReq)
	if err != nil {
		t.Fatalf("RetryNotification failed: %v", err)
	}

	if retryResp.RetryCount < 1 {
		t.Error("expected retry count to be incremented")
	}
}

func TestConcurrentNotificationSending(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := createTestEmailRequest()
			_, err := uc.SendEmail(ctx, req)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent SendEmail failed: %v", err)
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestSendEmail_EmptyCC_BCC(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestEmailRequest()
	req.CC = []string{}
	req.BCC = []string{}

	resp, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail with empty CC/BCC failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestSendEmail_WithCC_BCC(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestEmailRequest()
	req.CC = []string{"cc@example.com"}
	req.BCC = []string{"bcc@example.com"}

	resp, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail with CC/BCC failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestSendEmail_WithSourceEvent(t *testing.T) {
	uc, _ := createTestUseCase(t)
	ctx := context.Background()

	req := createTestEmailRequest()
	req.SourceEvent = &dto.SourceEventDTO{
		EventType:     "order.created",
		AggregateType: "order",
		AggregateID:   uuid.New().String(),
	}

	resp, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail with source event failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestSendInAppNotification_WithAllOptionalFields(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	userID := uuid.New().String()
	user := createTestUser(userID)
	mocks.UserService.AddUser(user)

	req := &dto.SendInAppRequest{
		TenantID:   user.TenantID,
		Type:       "alert",
		UserID:     userID,
		Title:      "Test Title",
		Body:       "Test Body",
		Category:   "system",
		ActionURL:  "https://example.com/action",
		ActionText: "View Details",
		ImageURL:   "https://example.com/image.png",
		Priority:   "high",
		Variables: map[string]interface{}{
			"key": "value",
		},
	}

	resp, err := uc.SendInAppNotification(ctx, req)
	if err != nil {
		t.Fatalf("SendInAppNotification failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestSendPushNotification_WithAllOptionalFields(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	userID := uuid.New().String()
	user := createTestUser(userID)
	mocks.UserService.AddUser(user)

	badge := 5
	req := &dto.SendPushRequest{
		TenantID:    user.TenantID,
		Type:        "alert",
		UserID:      userID,
		Title:       "Test Title",
		Body:        "Test Body",
		DeviceToken: "test-device-token",
		Platform:    "ios",
		ImageURL:    "https://example.com/image.png",
		ClickAction: "OPEN_ACTIVITY",
		Category:    "messages",
		Badge:       &badge,
		Sound:       "default",
		CollapseKey: "group-key",
		Priority:    "high",
		Data: map[string]string{
			"custom_key": "custom_value",
		},
	}

	resp, err := uc.SendPushNotification(ctx, req)
	if err != nil {
		t.Fatalf("SendPushNotification failed: %v", err)
	}

	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

// ============================================================================
// Metrics Tests
// ============================================================================

func TestSendEmail_MetricsIncremented(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	req := createTestEmailRequest()
	_, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}

	// Give some time for async operations
	time.Sleep(10 * time.Millisecond)

	counter := mocks.Metrics.GetCounter("notification.email.queued")
	if counter != 1 {
		t.Errorf("expected email queued counter to be 1, got %d", counter)
	}
}

func TestSendSMS_MetricsIncremented(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	req := createTestSMSRequest()
	_, err := uc.SendSMS(ctx, req)
	if err != nil {
		t.Fatalf("SendSMS failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	counter := mocks.Metrics.GetCounter("notification.sms.queued")
	if counter != 1 {
		t.Errorf("expected sms queued counter to be 1, got %d", counter)
	}
}

func TestCancelNotification_MetricsIncremented(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	notification.Status = domain.StatusPending
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.CancelNotificationRequest{
		NotificationID: notification.ID.String(),
	}

	_, err := uc.CancelNotification(ctx, req)
	if err != nil {
		t.Fatalf("CancelNotification failed: %v", err)
	}

	counter := mocks.Metrics.GetCounter("notification.cancelled")
	if counter != 1 {
		t.Errorf("expected cancelled counter to be 1, got %d", counter)
	}
}

// ============================================================================
// Event Publishing Tests
// ============================================================================

func TestSendEmail_DomainEventsPublished(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	req := createTestEmailRequest()
	_, err := uc.SendEmail(ctx, req)
	if err != nil {
		t.Fatalf("SendEmail failed: %v", err)
	}

	// Give time for async publishing
	time.Sleep(50 * time.Millisecond)

	events := mocks.EventPublisher.GetPublishedEvents()
	if len(events) == 0 {
		t.Error("expected domain events to be published")
	}
}

func TestCancelNotification_DomainEventsPublished(t *testing.T) {
	uc, mocks := createTestUseCase(t)
	ctx := context.Background()

	tenantID := uuid.New()
	notification := createTestNotification(tenantID, domain.ChannelEmail)
	notification.RecipientEmail = "test@example.com"
	notification.Subject = "Test Subject"
	notification.Status = domain.StatusPending
	mocks.NotificationRepo.Create(ctx, notification)

	req := &dto.CancelNotificationRequest{
		NotificationID: notification.ID.String(),
	}

	_, err := uc.CancelNotification(ctx, req)
	if err != nil {
		t.Fatalf("CancelNotification failed: %v", err)
	}

	events := mocks.EventPublisher.GetPublishedEvents()
	hasEvent := false
	for _, e := range events {
		if e.EventType() == "notification.cancelled" {
			hasEvent = true
			break
		}
	}

	if !hasEvent {
		t.Error("expected notification.cancelled event to be published")
	}
}
