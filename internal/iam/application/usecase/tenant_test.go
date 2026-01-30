package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ============================================================================
// Mock Implementations for Tenant Tests
// ============================================================================

// FullMockTenantRepository provides complete mock for tenant repository.
type FullMockTenantRepository struct {
	CreateFn       func(ctx context.Context, tenant *domain.Tenant) error
	UpdateFn       func(ctx context.Context, tenant *domain.Tenant) error
	DeleteFn       func(ctx context.Context, id uuid.UUID) error
	FindByIDFn     func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	FindBySlugFn   func(ctx context.Context, slug string) (*domain.Tenant, error)
	FindAllFn      func(ctx context.Context, opts domain.TenantQueryOptions) ([]*domain.Tenant, int64, error)
	ExistsBySlugFn func(ctx context.Context, slug string) (bool, error)
	CountFn        func(ctx context.Context) (int64, error)
}

func (m *FullMockTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, tenant)
	}
	return nil
}

func (m *FullMockTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, tenant)
	}
	return nil
}

func (m *FullMockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *FullMockTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *FullMockTenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	if m.FindBySlugFn != nil {
		return m.FindBySlugFn(ctx, slug)
	}
	return nil, errors.New("not found")
}

func (m *FullMockTenantRepository) FindAll(ctx context.Context, opts domain.TenantQueryOptions) ([]*domain.Tenant, int64, error) {
	if m.FindAllFn != nil {
		return m.FindAllFn(ctx, opts)
	}
	return nil, 0, nil
}

func (m *FullMockTenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	if m.ExistsBySlugFn != nil {
		return m.ExistsBySlugFn(ctx, slug)
	}
	return false, nil
}

func (m *FullMockTenantRepository) Count(ctx context.Context) (int64, error) {
	if m.CountFn != nil {
		return m.CountFn(ctx)
	}
	return 0, nil
}

// ============================================================================
// CreateTenantUseCase Tests
// ============================================================================

func TestCreateTenantUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
			return false, nil
		},
		CreateFn: func(ctx context.Context, tenant *domain.Tenant) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewCreateTenantUseCase(
		tenantRepo,
		outboxRepo,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.CreateTenantRequest{
		Name: "Test Company",
		Slug: "test-company",
		Plan: "starter",
	}

	result, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if result.Tenant.Name != "Test Company" {
		t.Errorf("Execute() tenant name = %s, want Test Company", result.Tenant.Name)
	}
	if result.Tenant.Plan != "starter" {
		t.Errorf("Execute() tenant plan = %s, want starter", result.Tenant.Plan)
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestCreateTenantUseCase_Execute_SlugAlreadyExists(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
			return true, nil // Slug already exists
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewCreateTenantUseCase(
		tenantRepo,
		outboxRepo,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.CreateTenantRequest{
		Name: "Test Company",
		Slug: "existing-slug",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when slug already exists")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "CONFLICT" {
		t.Errorf("Execute() error code = %s, want CONFLICT", appErr.Code)
	}
}

func TestCreateTenantUseCase_Execute_InvalidPlan(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
			return false, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewCreateTenantUseCase(
		tenantRepo,
		outboxRepo,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.CreateTenantRequest{
		Name: "Test Company",
		Slug: "test-company",
		Plan: "invalid_plan",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error for invalid plan")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("Execute() error code = %s, want VALIDATION_ERROR", appErr.Code)
	}
}

func TestCreateTenantUseCase_Execute_DefaultPlan(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
			return false, nil
		},
		CreateFn: func(ctx context.Context, tenant *domain.Tenant) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewCreateTenantUseCase(
		tenantRepo,
		outboxRepo,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.CreateTenantRequest{
		Name: "Test Company",
		Slug: "test-company",
		// No plan specified - should default to free
	}

	result, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result.Tenant.Plan != "free" {
		t.Errorf("Execute() tenant plan = %s, want free", result.Tenant.Plan)
	}
}

func TestCreateTenantUseCase_Execute_TransactionError(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
			return false, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{
		WithTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return errors.New("transaction error")
		},
	}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewCreateTenantUseCase(
		tenantRepo,
		outboxRepo,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.CreateTenantRequest{
		Name: "Test Company",
		Slug: "test-company",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when transaction fails")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "INTERNAL_ERROR" {
		t.Errorf("Execute() error code = %s, want INTERNAL_ERROR", appErr.Code)
	}
}

// ============================================================================
// GetTenantUseCase Tests
// ============================================================================

func TestGetTenantUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	userRepo := &MockUserRepository{}

	useCase := NewGetTenantUseCase(tenantRepo, userRepo)

	result, err := useCase.Execute(ctx, tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if result.Tenant.Name != tenant.Name() {
		t.Errorf("Execute() tenant name = %s, want %s", result.Tenant.Name, tenant.Name())
	}
}

func TestGetTenantUseCase_Execute_NotFound(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	userRepo := &MockUserRepository{}

	useCase := NewGetTenantUseCase(tenantRepo, userRepo)

	_, err := useCase.Execute(ctx, uuid.New())

	if err == nil {
		t.Fatal("Execute() should return error when tenant not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

// ============================================================================
// GetTenantBySlugUseCase Tests
// ============================================================================

func TestGetTenantBySlugUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &FullMockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewGetTenantBySlugUseCase(tenantRepo)

	result, err := useCase.Execute(ctx, "test-company")

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
}

func TestGetTenantBySlugUseCase_Execute_NotFound(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	useCase := NewGetTenantBySlugUseCase(tenantRepo)

	_, err := useCase.Execute(ctx, "nonexistent")

	if err == nil {
		t.Fatal("Execute() should return error when tenant not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

// ============================================================================
// ListTenantsUseCase Tests
// ============================================================================

func TestListTenantsUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant1 := createTestTenant(t)
	tenant2, _ := domain.NewTenant("Another Company", "another-company")

	tenantRepo := &FullMockTenantRepository{
		FindAllFn: func(ctx context.Context, opts domain.TenantQueryOptions) ([]*domain.Tenant, int64, error) {
			return []*domain.Tenant{tenant1, tenant2}, 2, nil
		},
	}

	useCase := NewListTenantsUseCase(tenantRepo)

	req := &dto.ListTenantsRequest{
		Page:     1,
		PageSize: 20,
	}

	result, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if len(result.Tenants) != 2 {
		t.Errorf("Execute() returned %d tenants, want 2", len(result.Tenants))
	}
	if result.Pagination.TotalItems != 2 {
		t.Errorf("Execute() TotalItems = %d, want 2", result.Pagination.TotalItems)
	}
}

func TestListTenantsUseCase_Execute_WithDefaults(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		FindAllFn: func(ctx context.Context, opts domain.TenantQueryOptions) ([]*domain.Tenant, int64, error) {
			// Verify defaults were applied
			if opts.Page != 1 {
				t.Errorf("Expected page 1, got %d", opts.Page)
			}
			if opts.PageSize != 20 {
				t.Errorf("Expected pageSize 20, got %d", opts.PageSize)
			}
			if opts.SortBy != "created_at" {
				t.Errorf("Expected sortBy 'created_at', got %s", opts.SortBy)
			}
			if opts.SortDirection != "desc" {
				t.Errorf("Expected sortDirection 'desc', got %s", opts.SortDirection)
			}
			return []*domain.Tenant{}, 0, nil
		},
	}

	useCase := NewListTenantsUseCase(tenantRepo)

	req := &dto.ListTenantsRequest{
		// Leave fields at zero values to test defaults
	}

	_, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestListTenantsUseCase_Execute_WithFilters(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		FindAllFn: func(ctx context.Context, opts domain.TenantQueryOptions) ([]*domain.Tenant, int64, error) {
			if opts.Status == nil {
				t.Error("Expected status filter")
			}
			if opts.Plan == nil {
				t.Error("Expected plan filter")
			}
			return []*domain.Tenant{}, 0, nil
		},
	}

	useCase := NewListTenantsUseCase(tenantRepo)

	req := &dto.ListTenantsRequest{
		Page:     1,
		PageSize: 20,
		Status:   "active",
		Plan:     "pro",
	}

	_, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

// ============================================================================
// UpdateTenantUseCase Tests
// ============================================================================

func TestUpdateTenantUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
		UpdateFn: func(ctx context.Context, t *domain.Tenant) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateTenantUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateTenantRequest{
		Name: "Updated Company Name",
	}

	result, err := useCase.Execute(ctx, tenant.GetID(), req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestUpdateTenantUseCase_Execute_NotFound(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateTenantUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateTenantRequest{
		Name: "Updated Company Name",
	}

	_, err := useCase.Execute(ctx, uuid.New(), req)

	if err == nil {
		t.Fatal("Execute() should return error when tenant not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

// ============================================================================
// UpdateTenantSettingsUseCase Tests
// ============================================================================

func TestUpdateTenantSettingsUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
		UpdateFn: func(ctx context.Context, t *domain.Tenant) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateTenantSettingsUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	timezone := "America/New_York"
	currency := "EUR"
	req := &dto.UpdateTenantSettingsRequest{
		Timezone: &timezone,
		Currency: &currency,
	}

	result, err := useCase.Execute(ctx, tenant.GetID(), req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
}

// ============================================================================
// ChangeTenantPlanUseCase Tests
// ============================================================================

func TestChangeTenantPlanUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
		UpdateFn: func(ctx context.Context, t *domain.Tenant) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewChangeTenantPlanUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateTenantPlanRequest{
		Plan: "pro",
	}

	result, err := useCase.Execute(ctx, tenant.GetID(), req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if result.Tenant.Plan != "pro" {
		t.Errorf("Execute() plan = %s, want pro", result.Tenant.Plan)
	}
}

func TestChangeTenantPlanUseCase_Execute_InvalidPlan(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewChangeTenantPlanUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateTenantPlanRequest{
		Plan: "invalid_plan",
	}

	_, err := useCase.Execute(ctx, tenant.GetID(), req)

	if err == nil {
		t.Fatal("Execute() should return error for invalid plan")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("Execute() error code = %s, want VALIDATION_ERROR", appErr.Code)
	}
}

// ============================================================================
// SuspendTenantUseCase Tests
// ============================================================================

func TestSuspendTenantUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
		UpdateFn: func(ctx context.Context, t *domain.Tenant) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewSuspendTenantUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	req := &dto.SuspendTenantRequest{
		TenantID: tenant.GetID(),
		Reason:   "Non-payment",
	}

	err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestSuspendTenantUseCase_Execute_NotFound(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewSuspendTenantUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	req := &dto.SuspendTenantRequest{
		TenantID: uuid.New(),
		Reason:   "Non-payment",
	}

	err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when tenant not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

// ============================================================================
// ActivateTenantUseCase Tests
// ============================================================================

func TestActivateTenantUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenant.Suspend("Test") // Suspend first

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
		UpdateFn: func(ctx context.Context, t *domain.Tenant) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewActivateTenantUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestActivateTenantUseCase_Execute_NotFound(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewActivateTenantUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, uuid.New())

	if err == nil {
		t.Fatal("Execute() should return error when tenant not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

// ============================================================================
// DeleteTenantUseCase Tests
// ============================================================================

func TestDeleteTenantUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteTenantUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestDeleteTenantUseCase_Execute_NotFound(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteTenantUseCase(tenantRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, uuid.New())

	if err == nil {
		t.Fatal("Execute() should return error when tenant not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

// ============================================================================
// Table-Driven Tests for Tenant Validation
// ============================================================================

func TestCreateTenantUseCase_ValidationScenarios(t *testing.T) {
	tests := []struct {
		name            string
		request         *dto.CreateTenantRequest
		setupMocks      func() *FullMockTenantRepository
		wantErr         bool
		expectedErrCode string
	}{
		{
			name: "empty_name",
			request: &dto.CreateTenantRequest{
				Name: "",
				Slug: "test-company",
			},
			setupMocks: func() *FullMockTenantRepository {
				return &FullMockTenantRepository{
					ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
						return false, nil
					},
				}
			},
			wantErr:         true,
			expectedErrCode: "VALIDATION_ERROR",
		},
		{
			name: "empty_slug",
			request: &dto.CreateTenantRequest{
				Name: "Test Company",
				Slug: "",
			},
			setupMocks: func() *FullMockTenantRepository {
				return &FullMockTenantRepository{
					ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
						return false, nil
					},
				}
			},
			wantErr:         true,
			expectedErrCode: "VALIDATION_ERROR",
		},
		{
			name: "invalid_slug_format",
			request: &dto.CreateTenantRequest{
				Name: "Test Company",
				Slug: "Invalid Slug!", // Contains spaces and special chars
			},
			setupMocks: func() *FullMockTenantRepository {
				return &FullMockTenantRepository{
					ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
						return false, nil
					},
				}
			},
			wantErr:         true,
			expectedErrCode: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantRepo := tt.setupMocks()

			useCase := NewCreateTenantUseCase(
				tenantRepo,
				&MockOutboxRepository{},
				&MockTransactionManager{},
				&MockEventPublisher{},
				&MockAuditLogger{},
			)

			_, err := useCase.Execute(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				var appErr *application.AppError
				if errors.As(err, &appErr) {
					if appErr.Code != tt.expectedErrCode {
						t.Errorf("Execute() error code = %s, want %s", appErr.Code, tt.expectedErrCode)
					}
				}
			}
		})
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkCreateTenantUseCase_Execute(b *testing.B) {
	ctx := context.Background()

	tenantRepo := &FullMockTenantRepository{
		ExistsBySlugFn: func(ctx context.Context, slug string) (bool, error) {
			return false, nil
		},
		CreateFn: func(ctx context.Context, tenant *domain.Tenant) error {
			return nil
		},
	}

	useCase := NewCreateTenantUseCase(
		tenantRepo,
		&MockOutboxRepository{},
		&MockTransactionManager{},
		&MockEventPublisher{},
		&MockAuditLogger{},
	)

	req := &dto.CreateTenantRequest{
		Name: "Test Company",
		Slug: "test-company",
		Plan: "starter",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, req)
	}
}

func BenchmarkListTenantsUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenant1, _ := domain.NewTenant("Company 1", "company-1")
	tenant2, _ := domain.NewTenant("Company 2", "company-2")

	tenantRepo := &FullMockTenantRepository{
		FindAllFn: func(ctx context.Context, opts domain.TenantQueryOptions) ([]*domain.Tenant, int64, error) {
			return []*domain.Tenant{tenant1, tenant2}, 2, nil
		},
	}

	useCase := NewListTenantsUseCase(tenantRepo)

	req := &dto.ListTenantsRequest{
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, req)
	}
}

func BenchmarkGetTenantUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenant, _ := domain.NewTenant("Test Company", "test-company")
	tenant.Activate()

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	userRepo := &MockUserRepository{}

	useCase := NewGetTenantUseCase(tenantRepo, userRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, tenant.GetID())
	}
}
