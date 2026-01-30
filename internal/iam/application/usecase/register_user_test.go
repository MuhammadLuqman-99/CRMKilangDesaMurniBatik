package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ============================================================================
// Mock Implementations for Register User Tests
// ============================================================================

// MockEmailService is a mock implementation of ports.EmailService.
type MockEmailService struct {
	SendWelcomeEmailFn           func(ctx context.Context, email, firstName, token string) error
	SendPasswordChangedEmailFn   func(ctx context.Context, email, firstName string) error
	SendEmailVerificationEmailFn func(ctx context.Context, email, firstName, token string) error
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, email, firstName, token string) error {
	if m.SendWelcomeEmailFn != nil {
		return m.SendWelcomeEmailFn(ctx, email, firstName, token)
	}
	return nil
}

func (m *MockEmailService) SendPasswordChangedEmail(ctx context.Context, email, firstName string) error {
	if m.SendPasswordChangedEmailFn != nil {
		return m.SendPasswordChangedEmailFn(ctx, email, firstName)
	}
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, email, firstName, token string) error {
	return nil
}

func (m *MockEmailService) SendEmailVerificationEmail(ctx context.Context, email, firstName, token string) error {
	if m.SendEmailVerificationEmailFn != nil {
		return m.SendEmailVerificationEmailFn(ctx, email, firstName, token)
	}
	return nil
}

// MockVerificationTokenService is a mock implementation of ports.VerificationTokenService.
type MockVerificationTokenService struct {
	GenerateEmailVerificationTokenFn  func(ctx context.Context, userID uuid.UUID) (string, error)
	ValidateEmailVerificationTokenFn  func(ctx context.Context, token string) (uuid.UUID, error)
	GeneratePasswordResetTokenFn      func(ctx context.Context, userID uuid.UUID) (string, error)
	ValidatePasswordResetTokenFn      func(ctx context.Context, token string) (uuid.UUID, error)
	InvalidateTokenFn                 func(ctx context.Context, token string) error
}

func (m *MockVerificationTokenService) GenerateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	if m.GenerateEmailVerificationTokenFn != nil {
		return m.GenerateEmailVerificationTokenFn(ctx, userID)
	}
	return "verification_token_" + userID.String(), nil
}

func (m *MockVerificationTokenService) ValidateEmailVerificationToken(ctx context.Context, token string) (uuid.UUID, error) {
	if m.ValidateEmailVerificationTokenFn != nil {
		return m.ValidateEmailVerificationTokenFn(ctx, token)
	}
	return uuid.New(), nil
}

func (m *MockVerificationTokenService) GeneratePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error) {
	if m.GeneratePasswordResetTokenFn != nil {
		return m.GeneratePasswordResetTokenFn(ctx, userID)
	}
	return "reset_token_" + userID.String(), nil
}

func (m *MockVerificationTokenService) ValidatePasswordResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	if m.ValidatePasswordResetTokenFn != nil {
		return m.ValidatePasswordResetTokenFn(ctx, token)
	}
	return uuid.New(), nil
}

func (m *MockVerificationTokenService) InvalidateToken(ctx context.Context, token string) error {
	if m.InvalidateTokenFn != nil {
		return m.InvalidateTokenFn(ctx, token)
	}
	return nil
}

// MockEventPublisher is a mock implementation of ports.EventPublisher.
type MockEventPublisher struct {
	PublishFn     func(ctx context.Context, event ports.DomainEventMessage) error
	PublishManyFn func(ctx context.Context, events []ports.DomainEventMessage) error
}

func (m *MockEventPublisher) Publish(ctx context.Context, event ports.DomainEventMessage) error {
	if m.PublishFn != nil {
		return m.PublishFn(ctx, event)
	}
	return nil
}

func (m *MockEventPublisher) PublishMany(ctx context.Context, events []ports.DomainEventMessage) error {
	if m.PublishManyFn != nil {
		return m.PublishManyFn(ctx, events)
	}
	return nil
}

// ExtendedMockUserRepository adds CountByTenant and ExistsByEmail functionality.
type ExtendedMockUserRepository struct {
	MockUserRepository
	CountByTenantFn  func(ctx context.Context, tenantID uuid.UUID) (int64, error)
	ExistsByEmailFn  func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error)
}

func (m *ExtendedMockUserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	if m.CountByTenantFn != nil {
		return m.CountByTenantFn(ctx, tenantID)
	}
	return 0, nil
}

func (m *ExtendedMockUserRepository) ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
	if m.ExistsByEmailFn != nil {
		return m.ExistsByEmailFn(ctx, tenantID, email)
	}
	return false, nil
}

// ExtendedMockRoleRepositoryForRegister adds FindByName functionality.
type ExtendedMockRoleRepositoryForRegister struct {
	MockRoleRepository
	FindByNameFn func(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error)
}

func (m *ExtendedMockRoleRepositoryForRegister) FindByName(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error) {
	if m.FindByNameFn != nil {
		return m.FindByNameFn(ctx, tenantID, name)
	}
	return nil, errors.New("not found")
}

// ============================================================================
// RegisterUserUseCase Tests
// ============================================================================

func TestRegisterUserUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &ExtendedMockUserRepository{
		MockUserRepository: MockUserRepository{
			CreateFn: func(ctx context.Context, user *domain.User) error {
				return nil
			},
			FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				email, _ := domain.NewEmail("newuser@example.com")
				user, _ := domain.NewUser(tenant.GetID(), email, domain.NewPasswordFromHash("hash"), "New", "User")
				return user, nil
			},
		},
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 1, nil
		},
		ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
			return false, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	viewerRole := domain.CreateViewerRole()
	roleRepo := &ExtendedMockRoleRepositoryForRegister{
		FindByNameFn: func(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error) {
			if name == domain.RoleNameViewer {
				return viewerRole, nil
			}
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "newuser@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("Execute() response should not be nil")
	}
	if resp.User == nil {
		t.Error("Execute() User should not be nil")
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestRegisterUserUseCase_Execute_TenantNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := &ExtendedMockUserRepository{}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &ExtendedMockRoleRepositoryForRegister{}
	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  uuid.New(),
		Email:     "newuser@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	_, err := useCase.Execute(ctx, req)

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

func TestRegisterUserUseCase_Execute_TenantInactive(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenant.Suspend("Test suspension") // Make tenant inactive

	userRepo := &ExtendedMockUserRepository{}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepositoryForRegister{}
	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "newuser@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when tenant is inactive")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "TENANT_INACTIVE" {
		t.Errorf("Execute() error code = %s, want TENANT_INACTIVE", appErr.Code)
	}
}

func TestRegisterUserUseCase_Execute_UserLimitReached(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	// Tenant is on free plan with max 3 users
	userRepo := &ExtendedMockUserRepository{
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 3, nil // Already at limit for free plan
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepositoryForRegister{}
	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "newuser@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when user limit reached")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "CONFLICT" {
		t.Errorf("Execute() error code = %s, want CONFLICT", appErr.Code)
	}
}

func TestRegisterUserUseCase_Execute_InvalidEmail(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &ExtendedMockUserRepository{
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 0, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepositoryForRegister{}
	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "invalid-email",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error for invalid email")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("Execute() error code = %s, want VALIDATION_ERROR", appErr.Code)
	}
}

func TestRegisterUserUseCase_Execute_EmailAlreadyExists(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &ExtendedMockUserRepository{
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 0, nil
		},
		ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
			return true, nil // Email already exists
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepositoryForRegister{}
	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "existing@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when email already exists")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "CONFLICT" {
		t.Errorf("Execute() error code = %s, want CONFLICT", appErr.Code)
	}
}

func TestRegisterUserUseCase_Execute_WeakPassword(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &ExtendedMockUserRepository{
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 0, nil
		},
		ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
			return false, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepositoryForRegister{}
	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "newuser@example.com",
		Password:  "weak", // Too short and simple
		FirstName: "New",
		LastName:  "User",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error for weak password")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("Execute() error code = %s, want VALIDATION_ERROR", appErr.Code)
	}
}

func TestRegisterUserUseCase_Execute_TransactionError(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &ExtendedMockUserRepository{
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 0, nil
		},
		ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
			return false, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	viewerRole := domain.CreateViewerRole()
	roleRepo := &ExtendedMockRoleRepositoryForRegister{
		FindByNameFn: func(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error) {
			if name == domain.RoleNameViewer {
				return viewerRole, nil
			}
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{
		WithTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return errors.New("transaction error")
		},
	}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "newuser@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
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

func TestRegisterUserUseCase_Execute_PasswordHashingError(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &ExtendedMockUserRepository{
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 0, nil
		},
		ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
			return false, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepositoryForRegister{}
	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{
		HashFn: func(password string) (string, error) {
			return "", errors.New("hashing error")
		},
	}
	tokenService := &MockTokenService{}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "newuser@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when password hashing fails")
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
// RegisterUserWithTokens Tests
// ============================================================================

func TestRegisterUserUseCase_RegisterUserWithTokens_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &ExtendedMockUserRepository{
		MockUserRepository: MockUserRepository{
			CreateFn: func(ctx context.Context, user *domain.User) error {
				return nil
			},
			FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				email, _ := domain.NewEmail("newuser@example.com")
				user, _ := domain.NewUser(tenant.GetID(), email, domain.NewPasswordFromHash("hash"), "New", "User")
				return user, nil
			},
		},
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 1, nil
		},
		ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
			return false, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	viewerRole := domain.CreateViewerRole()
	roleRepo := &ExtendedMockRoleRepositoryForRegister{
		FindByNameFn: func(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error) {
			if name == domain.RoleNameViewer {
				return viewerRole, nil
			}
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{}
	tokenService := &MockTokenService{
		GenerateAccessTokenFn: func(claims *ports.TokenClaims) (string, error) {
			return "access_token_123", nil
		},
	}
	emailService := &MockEmailService{}
	verifyService := &MockVerificationTokenService{}
	txManager := &MockTransactionManager{}
	eventPublisher := &MockEventPublisher{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		outboxRepo,
		passwordHasher,
		tokenService,
		emailService,
		verifyService,
		txManager,
		eventPublisher,
		auditLogger,
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "newuser@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	resp, err := useCase.RegisterUserWithTokens(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err != nil {
		t.Fatalf("RegisterUserWithTokens() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("RegisterUserWithTokens() response should not be nil")
	}
	if resp.AccessToken == "" {
		t.Error("RegisterUserWithTokens() AccessToken should not be empty")
	}
}

// ============================================================================
// Table-Driven Tests for Registration Validation
// ============================================================================

func TestRegisterUserUseCase_ValidationScenarios(t *testing.T) {
	tenant := createTestTenant(t)

	tests := []struct {
		name            string
		request         *dto.RegisterUserRequest
		setupMocks      func() (*ExtendedMockUserRepository, *MockTenantRepository)
		wantErr         bool
		expectedErrCode string
	}{
		{
			name: "empty_email",
			request: &dto.RegisterUserRequest{
				TenantID:  tenant.GetID(),
				Email:     "",
				Password:  "Password123!",
				FirstName: "New",
				LastName:  "User",
			},
			setupMocks: func() (*ExtendedMockUserRepository, *MockTenantRepository) {
				return &ExtendedMockUserRepository{
					CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
						return 0, nil
					},
				}, &MockTenantRepository{
					FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
						return tenant, nil
					},
				}
			},
			wantErr:         true,
			expectedErrCode: "VALIDATION_ERROR",
		},
		{
			name: "empty_password",
			request: &dto.RegisterUserRequest{
				TenantID:  tenant.GetID(),
				Email:     "test@example.com",
				Password:  "",
				FirstName: "New",
				LastName:  "User",
			},
			setupMocks: func() (*ExtendedMockUserRepository, *MockTenantRepository) {
				return &ExtendedMockUserRepository{
					CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
						return 0, nil
					},
					ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
						return false, nil
					},
				}, &MockTenantRepository{
					FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
						return tenant, nil
					},
				}
			},
			wantErr:         true,
			expectedErrCode: "VALIDATION_ERROR",
		},
		{
			name: "password_too_short",
			request: &dto.RegisterUserRequest{
				TenantID:  tenant.GetID(),
				Email:     "test@example.com",
				Password:  "Ab1!",
				FirstName: "New",
				LastName:  "User",
			},
			setupMocks: func() (*ExtendedMockUserRepository, *MockTenantRepository) {
				return &ExtendedMockUserRepository{
					CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
						return 0, nil
					},
					ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
						return false, nil
					},
				}, &MockTenantRepository{
					FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
						return tenant, nil
					},
				}
			},
			wantErr:         true,
			expectedErrCode: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, tenantRepo := tt.setupMocks()

			useCase := NewRegisterUserUseCase(
				userRepo,
				tenantRepo,
				&ExtendedMockRoleRepositoryForRegister{},
				&MockOutboxRepository{},
				&MockPasswordHasher{},
				&MockTokenService{},
				&MockEmailService{},
				&MockVerificationTokenService{},
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

func BenchmarkRegisterUserUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenant, _ := domain.NewTenant("Test Company", "test-company")
	tenant.Activate()

	userRepo := &ExtendedMockUserRepository{
		MockUserRepository: MockUserRepository{
			CreateFn: func(ctx context.Context, user *domain.User) error {
				return nil
			},
			FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				email, _ := domain.NewEmail("newuser@example.com")
				user, _ := domain.NewUser(tenant.GetID(), email, domain.NewPasswordFromHash("hash"), "New", "User")
				return user, nil
			},
		},
		CountByTenantFn: func(ctx context.Context, tenantID uuid.UUID) (int64, error) {
			return 0, nil
		},
		ExistsByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
			return false, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	viewerRole := domain.CreateViewerRole()
	roleRepo := &ExtendedMockRoleRepositoryForRegister{
		FindByNameFn: func(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error) {
			return viewerRole, nil
		},
	}

	useCase := NewRegisterUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		&MockOutboxRepository{},
		&MockPasswordHasher{},
		&MockTokenService{},
		&MockEmailService{},
		&MockVerificationTokenService{},
		&MockTransactionManager{},
		&MockEventPublisher{},
		&MockAuditLogger{},
	)

	req := &dto.RegisterUserRequest{
		TenantID:  tenant.GetID(),
		Email:     "newuser@example.com",
		Password:  "Password123!",
		FirstName: "New",
		LastName:  "User",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, req)
	}
}
