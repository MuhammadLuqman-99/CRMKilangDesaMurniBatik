package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ============================================================================
// RefreshTokenUseCase Tests
// ============================================================================

func TestRefreshTokenUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	// Create a valid refresh token
	refreshToken, plainToken, err := domain.NewRefreshToken(
		user.GetID(),
		time.Now().Add(24*time.Hour),
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			if id == user.GetID() {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			if id == tenant.GetID() {
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
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
		CreateFn: func(ctx context.Context, token *domain.RefreshToken) error {
			return nil
		},
	}

	tokenService := &MockTokenService{}
	txManager := &MockTransactionManager{}
	rateLimiter := &MockRateLimiter{}

	useCase := NewRefreshTokenUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		refreshTokenRepo,
		tokenService,
		txManager,
		rateLimiter,
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
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
	if resp.TokenType != "Bearer" {
		t.Errorf("Execute() TokenType = %s, want Bearer", resp.TokenType)
	}
}

func TestRefreshTokenUseCase_Execute_RateLimited(t *testing.T) {
	ctx := context.Background()

	rateLimiter := &MockRateLimiter{
		AllowFn: func(ctx context.Context, key string) (bool, error) {
			return false, nil
		},
	}

	useCase := NewRefreshTokenUseCase(
		&MockUserRepository{},
		&MockTenantRepository{},
		&MockRoleRepository{},
		&MockRefreshTokenRepository{},
		&MockTokenService{},
		&MockTransactionManager{},
		rateLimiter,
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: "some_token",
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

func TestRefreshTokenUseCase_Execute_TokenNotFound(t *testing.T) {
	ctx := context.Background()

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return nil, errors.New("not found")
		},
	}

	useCase := NewRefreshTokenUseCase(
		&MockUserRepository{},
		&MockTenantRepository{},
		&MockRoleRepository{},
		refreshTokenRepo,
		&MockTokenService{},
		&MockTransactionManager{},
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: "invalid_token",
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when token not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "TOKEN_INVALID" {
		t.Errorf("Execute() error code = %s, want TOKEN_INVALID", appErr.Code)
	}
}

func TestRefreshTokenUseCase_Execute_TokenExpired(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	// Create an expired token
	expiredToken, plainToken, err := domain.NewRefreshToken(
		userID,
		time.Now().Add(-1*time.Hour), // Expired
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return expiredToken, nil
		},
	}

	useCase := NewRefreshTokenUseCase(
		&MockUserRepository{},
		&MockTenantRepository{},
		&MockRoleRepository{},
		refreshTokenRepo,
		&MockTokenService{},
		&MockTransactionManager{},
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
	}

	_, err = useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when token is expired")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "TOKEN_EXPIRED" {
		t.Errorf("Execute() error code = %s, want TOKEN_EXPIRED", appErr.Code)
	}
}

func TestRefreshTokenUseCase_Execute_TokenRevoked(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	// Create and revoke a token
	revokedToken, plainToken, err := domain.NewRefreshToken(
		userID,
		time.Now().Add(24*time.Hour),
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	revokedToken.Revoke() // Revoke the token

	revokeByUserIDCalled := false
	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return revokedToken, nil
		},
		RevokeByUserIDFn: func(ctx context.Context, uid uuid.UUID) error {
			revokeByUserIDCalled = true
			return nil
		},
	}

	useCase := NewRefreshTokenUseCase(
		&MockUserRepository{},
		&MockTenantRepository{},
		&MockRoleRepository{},
		refreshTokenRepo,
		&MockTokenService{},
		&MockTransactionManager{},
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
	}

	_, err = useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when token is revoked")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "TOKEN_INVALID" {
		t.Errorf("Execute() error code = %s, want TOKEN_INVALID", appErr.Code)
	}

	// Token reuse should trigger revoking all user tokens
	if !revokeByUserIDCalled {
		t.Error("Execute() should revoke all user tokens when reused token detected")
	}
}

func TestRefreshTokenUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	refreshToken, plainToken, _ := domain.NewRefreshToken(
		userID,
		time.Now().Add(24*time.Hour),
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
	}

	useCase := NewRefreshTokenUseCase(
		userRepo,
		&MockTenantRepository{},
		&MockRoleRepository{},
		refreshTokenRepo,
		&MockTokenService{},
		&MockTransactionManager{},
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when user not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

func TestRefreshTokenUseCase_Execute_UserInactive(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	user.Suspend("Test suspension") // Make user inactive

	refreshToken, plainToken, _ := domain.NewRefreshToken(
		user.GetID(),
		time.Now().Add(24*time.Hour),
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
	}

	useCase := NewRefreshTokenUseCase(
		userRepo,
		&MockTenantRepository{},
		&MockRoleRepository{},
		refreshTokenRepo,
		&MockTokenService{},
		&MockTransactionManager{},
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

	if err == nil {
		t.Fatal("Execute() should return error when user is inactive")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "USER_INACTIVE" {
		t.Errorf("Execute() error code = %s, want USER_INACTIVE", appErr.Code)
	}
}

func TestRefreshTokenUseCase_Execute_TenantNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	refreshToken, plainToken, _ := domain.NewRefreshToken(
		user.GetID(),
		time.Now().Add(24*time.Hour),
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
	}

	useCase := NewRefreshTokenUseCase(
		userRepo,
		tenantRepo,
		&MockRoleRepository{},
		refreshTokenRepo,
		&MockTokenService{},
		&MockTransactionManager{},
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

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

func TestRefreshTokenUseCase_Execute_TenantInactive(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenant.Suspend("Test suspension") // Make tenant inactive
	user := createTestUser(t, tenant.GetID())

	refreshToken, plainToken, _ := domain.NewRefreshToken(
		user.GetID(),
		time.Now().Add(24*time.Hour),
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
	}

	useCase := NewRefreshTokenUseCase(
		userRepo,
		tenantRepo,
		&MockRoleRepository{},
		refreshTokenRepo,
		&MockTokenService{},
		&MockTransactionManager{},
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
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

func TestRefreshTokenUseCase_Execute_TransactionError(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	refreshToken, plainToken, _ := domain.NewRefreshToken(
		user.GetID(),
		time.Now().Add(24*time.Hour),
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		},
	}

	txManager := &MockTransactionManager{
		WithTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
			return errors.New("transaction error")
		},
	}

	useCase := NewRefreshTokenUseCase(
		userRepo,
		tenantRepo,
		&MockRoleRepository{},
		refreshTokenRepo,
		&MockTokenService{},
		txManager,
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
	}

	_, err := useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")

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
// CleanupExpiredTokensUseCase Tests
// ============================================================================

func TestCleanupExpiredTokensUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()

	deletedCount := int64(10)

	// Use ExtendedRefreshTokenRepo with DeleteExpiredFn
	extendedRepo := &ExtendedRefreshTokenRepo{
		DeleteExpiredFn: func(ctx context.Context) (int64, error) {
			return deletedCount, nil
		},
	}

	useCase := NewCleanupExpiredTokensUseCase(extendedRepo)

	count, err := useCase.Execute(ctx)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if count != deletedCount {
		t.Errorf("Execute() count = %d, want %d", count, deletedCount)
	}
}

func TestCleanupExpiredTokensUseCase_Execute_Error(t *testing.T) {
	ctx := context.Background()

	type ExtendedRefreshTokenRepo struct {
		MockRefreshTokenRepository
		DeleteExpiredFn func(ctx context.Context) (int64, error)
	}

	extendedRepo := &ExtendedRefreshTokenRepo{
		DeleteExpiredFn: func(ctx context.Context) (int64, error) {
			return 0, errors.New("database error")
		},
	}

	useCase := NewCleanupExpiredTokensUseCase(extendedRepo)

	_, err := useCase.Execute(ctx)

	if err == nil {
		t.Fatal("Execute() should return error when cleanup fails")
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
// Table-Driven Tests for RefreshToken Validation
// ============================================================================

func TestRefreshTokenUseCase_ValidationScenarios(t *testing.T) {
	tests := []struct {
		name            string
		setupMocks      func() (*MockRefreshTokenRepository, *MockRateLimiter)
		request         *dto.RefreshTokenRequest
		wantErr         bool
		expectedErrCode string
	}{
		{
			name: "empty_refresh_token",
			setupMocks: func() (*MockRefreshTokenRepository, *MockRateLimiter) {
				return &MockRefreshTokenRepository{
					FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
						return nil, errors.New("not found")
					},
				}, &MockRateLimiter{}
			},
			request: &dto.RefreshTokenRequest{
				RefreshToken: "",
			},
			wantErr:         true,
			expectedErrCode: "TOKEN_INVALID",
		},
		{
			name: "rate_limiter_error",
			setupMocks: func() (*MockRefreshTokenRepository, *MockRateLimiter) {
				return &MockRefreshTokenRepository{}, &MockRateLimiter{
					AllowFn: func(ctx context.Context, key string) (bool, error) {
						return false, errors.New("rate limiter error")
					},
				}
			},
			request: &dto.RefreshTokenRequest{
				RefreshToken: "some_token",
			},
			wantErr:         true,
			expectedErrCode: "RATE_LIMITED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshTokenRepo, rateLimiter := tt.setupMocks()

			useCase := NewRefreshTokenUseCase(
				&MockUserRepository{},
				&MockTenantRepository{},
				&MockRoleRepository{},
				refreshTokenRepo,
				&MockTokenService{},
				&MockTransactionManager{},
				rateLimiter,
			)

			_, err := useCase.Execute(context.Background(), tt.request, "192.168.1.1", "Mozilla/5.0")

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

func BenchmarkRefreshTokenUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenant, _ := domain.NewTenant("Test Company", "test-company")
	tenant.Activate()
	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenant.GetID(), email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Activate()

	refreshToken, plainToken, _ := domain.NewRefreshToken(
		user.GetID(),
		time.Now().Add(24*time.Hour),
		"192.168.1.1",
		"Mozilla/5.0",
		domain.DeviceInfo{},
	)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &MockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{}, nil
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		FindByTokenHashFn: func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			// Return fresh token each time
			newToken, _, _ := domain.NewRefreshToken(
				user.GetID(),
				time.Now().Add(24*time.Hour),
				"192.168.1.1",
				"Mozilla/5.0",
				domain.DeviceInfo{},
			)
			return newToken, nil
		},
		CreateFn: func(ctx context.Context, token *domain.RefreshToken) error {
			return nil
		},
	}

	useCase := NewRefreshTokenUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		refreshTokenRepo,
		&MockTokenService{},
		&MockTransactionManager{},
		&MockRateLimiter{},
	)

	req := &dto.RefreshTokenRequest{
		RefreshToken: plainToken,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		refreshTokenRepo.FindByTokenHashFn = func(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
			return refreshToken, nil
		}
		_, _ = useCase.Execute(ctx, req, "192.168.1.1", "Mozilla/5.0")
	}
}

// ExtendedRefreshTokenRepo for cleanup tests
type ExtendedRefreshTokenRepo struct {
	MockRefreshTokenRepository
	DeleteExpiredFn func(ctx context.Context) (int64, error)
}

func (m *ExtendedRefreshTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	if m.DeleteExpiredFn != nil {
		return m.DeleteExpiredFn(ctx)
	}
	return 0, nil
}
