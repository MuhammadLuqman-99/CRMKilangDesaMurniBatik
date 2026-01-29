package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// MockUserRepository is a mock implementation of domain.UserRepository.
type MockUserRepository struct {
	FindByEmailFn func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error)
	UpdateFn      func(ctx context.Context, user *domain.User) error
	CreateFn      func(ctx context.Context, user *domain.User) error
	FindByIDFn    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(ctx, tenantID, email)
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, opts domain.UserQueryOptions) ([]*domain.User, int64, error) {
	return nil, 0, nil
}

func (m *MockUserRepository) FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
	return nil, nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
	return false, nil
}

func (m *MockUserRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return 0, nil
}

// MockTenantRepository is a mock implementation of domain.TenantRepository.
type MockTenantRepository struct {
	FindBySlugFn func(ctx context.Context, slug string) (*domain.Tenant, error)
	FindByIDFn   func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
}

func (m *MockTenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	return nil
}

func (m *MockTenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	return nil
}

func (m *MockTenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockTenantRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	if m.FindBySlugFn != nil {
		return m.FindBySlugFn(ctx, slug)
	}
	return nil, errors.New("not found")
}

func (m *MockTenantRepository) FindAll(ctx context.Context, opts domain.TenantQueryOptions) ([]*domain.Tenant, int64, error) {
	return nil, 0, nil
}

func (m *MockTenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return false, nil
}

func (m *MockTenantRepository) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

// MockRoleRepository is a mock implementation of domain.RoleRepository.
type MockRoleRepository struct {
	FindByUserIDFn func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
}

func (m *MockRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	return nil
}

func (m *MockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return nil
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return nil, errors.New("not found")
}

func (m *MockRoleRepository) FindByName(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error) {
	return nil, errors.New("not found")
}

func (m *MockRoleRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, opts domain.RoleQueryOptions) ([]*domain.Role, int64, error) {
	return nil, 0, nil
}

func (m *MockRoleRepository) FindSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	return nil, nil
}

func (m *MockRoleRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	if m.FindByUserIDFn != nil {
		return m.FindByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockRoleRepository) ExistsByName(ctx context.Context, tenantID *uuid.UUID, name string) (bool, error) {
	return false, nil
}

func (m *MockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
	return nil
}

func (m *MockRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return nil
}

// MockRefreshTokenRepository is a mock implementation of domain.RefreshTokenRepository.
type MockRefreshTokenRepository struct {
	CreateFn              func(ctx context.Context, token *domain.RefreshToken) error
	CountActiveByUserIDFn func(ctx context.Context, userID uuid.UUID) (int64, error)
	FindActiveByUserIDFn  func(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error)
	RevokeByIDFn          func(ctx context.Context, id uuid.UUID) error
	FindByTokenHashFn     func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeByUserIDFn      func(ctx context.Context, userID uuid.UUID) error
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, token)
	}
	return nil
}

func (m *MockRefreshTokenRepository) Update(ctx context.Context, token *domain.RefreshToken) error {
	return nil
}

func (m *MockRefreshTokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockRefreshTokenRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.RefreshToken, error) {
	return nil, errors.New("not found")
}

func (m *MockRefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	if m.FindByTokenHashFn != nil {
		return m.FindByTokenHashFn(ctx, tokenHash)
	}
	return nil, errors.New("not found")
}

func (m *MockRefreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error) {
	return nil, nil
}

func (m *MockRefreshTokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error) {
	if m.FindActiveByUserIDFn != nil {
		return m.FindActiveByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockRefreshTokenRepository) RevokeByUserID(ctx context.Context, userID uuid.UUID) error {
	if m.RevokeByUserIDFn != nil {
		return m.RevokeByUserIDFn(ctx, userID)
	}
	return nil
}

func (m *MockRefreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	if m.RevokeByIDFn != nil {
		return m.RevokeByIDFn(ctx, id)
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockRefreshTokenRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.CountActiveByUserIDFn != nil {
		return m.CountActiveByUserIDFn(ctx, userID)
	}
	return 0, nil
}

// MockPasswordHasher is a mock implementation of ports.PasswordHasher.
type MockPasswordHasher struct {
	HashFn   func(password string) (string, error)
	VerifyFn func(password, hash string) (bool, error)
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	if m.HashFn != nil {
		return m.HashFn(password)
	}
	return "hashed_" + password, nil
}

func (m *MockPasswordHasher) Verify(password, hash string) (bool, error) {
	if m.VerifyFn != nil {
		return m.VerifyFn(password, hash)
	}
	return hash == "hashed_"+password, nil
}

// MockTokenService is a mock implementation of ports.TokenService.
type MockTokenService struct {
	GenerateAccessTokenFn  func(claims *ports.TokenClaims) (string, error)
	ValidateAccessTokenFn  func(token string) (*ports.TokenClaims, error)
	GetAccessTokenExpiryFn func() time.Duration
	GetRefreshTokenExpiryFn func() time.Duration
}

func (m *MockTokenService) GenerateAccessToken(claims *ports.TokenClaims) (string, error) {
	if m.GenerateAccessTokenFn != nil {
		return m.GenerateAccessTokenFn(claims)
	}
	return "access_token_" + claims.UserID.String(), nil
}

func (m *MockTokenService) GenerateRefreshToken() (string, error) {
	return "refresh_token_" + uuid.New().String(), nil
}

func (m *MockTokenService) ValidateAccessToken(token string) (*ports.TokenClaims, error) {
	if m.ValidateAccessTokenFn != nil {
		return m.ValidateAccessTokenFn(token)
	}
	return nil, errors.New("invalid token")
}

func (m *MockTokenService) GetAccessTokenExpiry() time.Duration {
	if m.GetAccessTokenExpiryFn != nil {
		return m.GetAccessTokenExpiryFn()
	}
	return 15 * time.Minute
}

func (m *MockTokenService) GetRefreshTokenExpiry() time.Duration {
	if m.GetRefreshTokenExpiryFn != nil {
		return m.GetRefreshTokenExpiryFn()
	}
	return 7 * 24 * time.Hour
}

// MockRateLimiter is a mock implementation of ports.RateLimiter.
type MockRateLimiter struct {
	AllowFn func(ctx context.Context, key string) (bool, error)
}

func (m *MockRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if m.AllowFn != nil {
		return m.AllowFn(ctx, key)
	}
	return true, nil
}

func (m *MockRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	return true, nil
}

func (m *MockRateLimiter) Reset(ctx context.Context, key string) error {
	return nil
}

// MockTransactionManager is a mock implementation of ports.TransactionManager.
type MockTransactionManager struct {
	WithTransactionFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *MockTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.WithTransactionFn != nil {
		return m.WithTransactionFn(ctx, fn)
	}
	return fn(ctx)
}

func (m *MockTransactionManager) BeginTx(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (m *MockTransactionManager) CommitTx(ctx context.Context) error {
	return nil
}

func (m *MockTransactionManager) RollbackTx(ctx context.Context) error {
	return nil
}

// MockAuditLogger is a mock implementation of ports.AuditLogger.
type MockAuditLogger struct {
	LogFn func(ctx context.Context, entry ports.AuditEntry) error
	Calls []ports.AuditEntry
}

func (m *MockAuditLogger) Log(ctx context.Context, entry ports.AuditEntry) error {
	m.Calls = append(m.Calls, entry)
	if m.LogFn != nil {
		return m.LogFn(ctx, entry)
	}
	return nil
}

// MockTokenBlacklist is a mock implementation of ports.TokenBlacklist.
type MockTokenBlacklist struct {
	AddFn           func(ctx context.Context, token string, expiration time.Duration) error
	IsBlacklistedFn func(ctx context.Context, token string) (bool, error)
}

func (m *MockTokenBlacklist) Add(ctx context.Context, token string, expiration time.Duration) error {
	if m.AddFn != nil {
		return m.AddFn(ctx, token, expiration)
	}
	return nil
}

func (m *MockTokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if m.IsBlacklistedFn != nil {
		return m.IsBlacklistedFn(ctx, token)
	}
	return false, nil
}

// ============================================================================
// Test Helpers
// ============================================================================

func createTestTenant(t *testing.T) *domain.Tenant {
	t.Helper()
	tenant, err := domain.NewTenant("Test Company", "test-company")
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}
	// Activate tenant so it's ready for use
	if err := tenant.Activate(); err != nil {
		t.Fatalf("Failed to activate test tenant: %v", err)
	}
	return tenant
}

func createTestUser(t *testing.T, tenantID uuid.UUID) *domain.User {
	t.Helper()
	email, _ := domain.NewEmail("test@example.com")
	passwordHash := domain.NewPasswordFromHash("hashed_password123")
	user, err := domain.NewUser(
		tenantID,
		email,
		passwordHash,
		"John",
		"Doe",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	user.Activate() // Ensure user is active
	return user
}

// ============================================================================
// AuthenticateUserUseCase Tests
// ============================================================================

func TestAuthenticateUserUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &MockUserRepository{
		FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
			if tenantID == tenant.GetID() && email.String() == "test@example.com" {
				return user, nil
			}
			return nil, errors.New("not found")
		},
		UpdateFn: func(ctx context.Context, u *domain.User) error {
			return nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			if slug == "test-company" {
				return tenant, nil
			}
			return nil, errors.New("not found")
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{}, nil
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		CountActiveByUserIDFn: func(ctx context.Context, userID uuid.UUID) (int64, error) {
			return 0, nil
		},
		CreateFn: func(ctx context.Context, token *domain.RefreshToken) error {
			return nil
		},
	}

	passwordHasher := &MockPasswordHasher{
		VerifyFn: func(password, hash string) (bool, error) {
			return password == "password123", nil
		},
	}

	tokenService := &MockTokenService{}
	rateLimiter := &MockRateLimiter{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAuthenticateUserUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		refreshTokenRepo,
		passwordHasher,
		tokenService,
		rateLimiter,
		txManager,
		auditLogger,
	)

	req := &dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantSlug: "test-company",
	}

	resp, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("Execute() response should not be nil")
	}
	if resp.AccessToken == "" {
		t.Error("Execute() AccessToken should not be empty")
	}
	if resp.RefreshToken == "" {
		t.Error("Execute() RefreshToken should not be empty")
	}
	if resp.User == nil {
		t.Error("Execute() User should not be nil")
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("Execute() TokenType = %s, want Bearer", resp.TokenType)
	}
}

func TestAuthenticateUserUseCase_Execute_RateLimited(t *testing.T) {
	ctx := context.Background()

	rateLimiter := &MockRateLimiter{
		AllowFn: func(ctx context.Context, key string) (bool, error) {
			return false, nil
		},
	}

	useCase := NewAuthenticateUserUseCase(
		&MockUserRepository{},
		&MockTenantRepository{},
		&MockRoleRepository{},
		&MockRefreshTokenRepository{},
		&MockPasswordHasher{},
		&MockTokenService{},
		rateLimiter,
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantSlug: "test-company",
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when rate limited")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "RATE_LIMITED" {
		t.Errorf("Execute() error code = %s, want RATE_LIMITED", appErr.Code)
	}
}

func TestAuthenticateUserUseCase_Execute_TenantNotFound(t *testing.T) {
	ctx := context.Background()

	tenantRepo := &MockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	auditLogger := &MockAuditLogger{}

	useCase := NewAuthenticateUserUseCase(
		&MockUserRepository{},
		tenantRepo,
		&MockRoleRepository{},
		&MockRefreshTokenRepository{},
		&MockPasswordHasher{},
		&MockTokenService{},
		&MockRateLimiter{},
		&MockTransactionManager{},
		auditLogger,
	)

	req := &dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantSlug: "nonexistent",
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when tenant not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "INVALID_CREDENTIALS" {
		t.Errorf("Execute() error code = %s, want INVALID_CREDENTIALS", appErr.Code)
	}

	// Verify audit log was called
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log failed login attempt")
	}
}

func TestAuthenticateUserUseCase_Execute_TenantInactive(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenant.Suspend("Test suspension reason") // Make tenant inactive

	tenantRepo := &MockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewAuthenticateUserUseCase(
		&MockUserRepository{},
		tenantRepo,
		&MockRoleRepository{},
		&MockRefreshTokenRepository{},
		&MockPasswordHasher{},
		&MockTokenService{},
		&MockRateLimiter{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantSlug: "test-company",
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

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

func TestAuthenticateUserUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	tenantRepo := &MockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	userRepo := &MockUserRepository{
		FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	useCase := NewAuthenticateUserUseCase(
		userRepo,
		tenantRepo,
		&MockRoleRepository{},
		&MockRefreshTokenRepository{},
		&MockPasswordHasher{},
		&MockTokenService{},
		&MockRateLimiter{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.LoginRequest{
		Email:      "nonexistent@example.com",
		Password:   "password123",
		TenantSlug: "test-company",
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when user not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "INVALID_CREDENTIALS" {
		t.Errorf("Execute() error code = %s, want INVALID_CREDENTIALS", appErr.Code)
	}
}

func TestAuthenticateUserUseCase_Execute_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	tenantRepo := &MockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	userRepo := &MockUserRepository{
		FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
			return user, nil
		},
	}

	passwordHasher := &MockPasswordHasher{
		VerifyFn: func(password, hash string) (bool, error) {
			return false, nil // Password doesn't match
		},
	}

	useCase := NewAuthenticateUserUseCase(
		userRepo,
		tenantRepo,
		&MockRoleRepository{},
		&MockRefreshTokenRepository{},
		passwordHasher,
		&MockTokenService{},
		&MockRateLimiter{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "wrong_password",
		TenantSlug: "test-company",
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when password is invalid")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "INVALID_CREDENTIALS" {
		t.Errorf("Execute() error code = %s, want INVALID_CREDENTIALS", appErr.Code)
	}
}

func TestAuthenticateUserUseCase_Execute_UserSuspended(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	user.Suspend("Test suspension reason") // Suspend user

	tenantRepo := &MockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	userRepo := &MockUserRepository{
		FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
			return user, nil
		},
	}

	useCase := NewAuthenticateUserUseCase(
		userRepo,
		tenantRepo,
		&MockRoleRepository{},
		&MockRefreshTokenRepository{},
		&MockPasswordHasher{},
		&MockTokenService{},
		&MockRateLimiter{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantSlug: "test-company",
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when user is suspended")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "USER_INACTIVE" {
		t.Errorf("Execute() error code = %s, want USER_INACTIVE", appErr.Code)
	}
}

func TestAuthenticateUserUseCase_Execute_TokenGenerationFailure(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &MockUserRepository{
		FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	passwordHasher := &MockPasswordHasher{
		VerifyFn: func(password, hash string) (bool, error) {
			return true, nil
		},
	}

	tokenService := &MockTokenService{
		GenerateAccessTokenFn: func(claims *ports.TokenClaims) (string, error) {
			return "", errors.New("token generation failed")
		},
	}

	useCase := NewAuthenticateUserUseCase(
		userRepo,
		tenantRepo,
		&MockRoleRepository{},
		&MockRefreshTokenRepository{},
		passwordHasher,
		tokenService,
		&MockRateLimiter{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantSlug: "test-company",
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when token generation fails")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "INTERNAL_ERROR" {
		t.Errorf("Execute() error code = %s, want INTERNAL_ERROR", appErr.Code)
	}
}

func TestAuthenticateUserUseCase_Execute_MaxActiveTokensReached(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &MockUserRepository{
		FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
			return user, nil
		},
		UpdateFn: func(ctx context.Context, u *domain.User) error {
			return nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindBySlugFn: func(ctx context.Context, slug string) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	passwordHasher := &MockPasswordHasher{
		VerifyFn: func(password, hash string) (bool, error) {
			return true, nil
		},
	}

	// Create mock active tokens
	activeTokens := make([]*domain.RefreshToken, 6)
	for i := range activeTokens {
		token, _, _ := domain.NewRefreshToken(user.GetID(), time.Now().Add(24*time.Hour), "127.0.0.1", "Test", domain.DeviceInfo{})
		activeTokens[i] = token
	}

	revokedTokenIDs := make([]uuid.UUID, 0)
	refreshTokenRepo := &MockRefreshTokenRepository{
		CountActiveByUserIDFn: func(ctx context.Context, userID uuid.UUID) (int64, error) {
			return 6, nil // More than max allowed
		},
		FindActiveByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error) {
			return activeTokens, nil
		},
		RevokeByIDFn: func(ctx context.Context, id uuid.UUID) error {
			revokedTokenIDs = append(revokedTokenIDs, id)
			return nil
		},
		CreateFn: func(ctx context.Context, token *domain.RefreshToken) error {
			return nil
		},
	}

	useCase := NewAuthenticateUserUseCase(
		userRepo,
		tenantRepo,
		&MockRoleRepository{},
		refreshTokenRepo,
		passwordHasher,
		&MockTokenService{},
		&MockRateLimiter{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantSlug: "test-company",
	}

	resp, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("Execute() response should not be nil")
	}

	// Should have revoked some tokens
	if len(revokedTokenIDs) == 0 {
		t.Error("Execute() should revoke oldest tokens when max active tokens exceeded")
	}
}

// ============================================================================
// LogoutUseCase Tests
// ============================================================================

func TestLogoutUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	// Create a valid refresh token
	refreshToken, plainToken, _ := domain.NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", domain.DeviceInfo{})

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
		RevokeByIDFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	useCase := NewLogoutUseCase(
		refreshTokenRepo,
		&MockTokenBlacklist{},
		&MockTokenService{},
		&MockAuditLogger{},
	)

	req := &dto.LogoutRequest{
		RefreshToken: plainToken,
		AllDevices:   false,
	}

	err := useCase.Execute(ctx, req, "access_token", userID, tenantID)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestLogoutUseCase_Execute_AllDevices(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	refreshToken, plainToken, _ := domain.NewRefreshToken(userID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", domain.DeviceInfo{})

	revokeByUserIDCalled := false
	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
		RevokeByUserIDFn: func(ctx context.Context, uid uuid.UUID) error {
			revokeByUserIDCalled = true
			return nil
		},
	}

	useCase := NewLogoutUseCase(
		refreshTokenRepo,
		&MockTokenBlacklist{},
		&MockTokenService{},
		&MockAuditLogger{},
	)

	req := &dto.LogoutRequest{
		RefreshToken: plainToken,
		AllDevices:   true,
	}

	err := useCase.Execute(ctx, req, "access_token", userID, tenantID)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if !revokeByUserIDCalled {
		t.Error("Execute() should revoke all tokens when AllDevices is true")
	}
}

func TestLogoutUseCase_Execute_InvalidToken(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return nil, errors.New("not found")
		},
	}

	useCase := NewLogoutUseCase(
		refreshTokenRepo,
		&MockTokenBlacklist{},
		&MockTokenService{},
		&MockAuditLogger{},
	)

	req := &dto.LogoutRequest{
		RefreshToken: "invalid_token",
		AllDevices:   false,
	}

	err := useCase.Execute(ctx, req, "access_token", userID, tenantID)

	if err == nil {
		t.Fatal("Execute() should return error when token is invalid")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "TOKEN_INVALID" {
		t.Errorf("Execute() error code = %s, want TOKEN_INVALID", appErr.Code)
	}
}

func TestLogoutUseCase_Execute_TokenNotBelongToUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()
	tenantID := uuid.New()

	// Token belongs to different user
	refreshToken, plainToken, _ := domain.NewRefreshToken(otherUserID, time.Now().Add(24*time.Hour), "192.168.1.1", "Mozilla/5.0", domain.DeviceInfo{})

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
	}

	useCase := NewLogoutUseCase(
		refreshTokenRepo,
		&MockTokenBlacklist{},
		&MockTokenService{},
		&MockAuditLogger{},
	)

	req := &dto.LogoutRequest{
		RefreshToken: plainToken,
		AllDevices:   false,
	}

	err := useCase.Execute(ctx, req, "access_token", userID, tenantID)

	if err == nil {
		t.Fatal("Execute() should return error when token doesn't belong to user")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("Execute() error code = %s, want FORBIDDEN", appErr.Code)
	}
}
