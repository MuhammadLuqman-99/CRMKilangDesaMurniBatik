package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ============================================================================
// Mock Implementations for Validate Permission Tests
// ============================================================================

// MockCacheService is a mock implementation of ports.CacheService.
type MockCacheService struct {
	GetFn       func(ctx context.Context, key string) ([]byte, error)
	SetFn       func(ctx context.Context, key string, value []byte, expiration time.Duration) error
	DeleteFn    func(ctx context.Context, key string) error
	ExistsFn    func(ctx context.Context, key string) (bool, error)
	IncrementFn func(ctx context.Context, key string) (int64, error)
}

func (m *MockCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, key)
	}
	return nil, errors.New("cache miss")
}

func (m *MockCacheService) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if m.SetFn != nil {
		return m.SetFn(ctx, key, value, expiration)
	}
	return nil
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, key)
	}
	return nil
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	if m.ExistsFn != nil {
		return m.ExistsFn(ctx, key)
	}
	return false, nil
}

func (m *MockCacheService) Increment(ctx context.Context, key string) (int64, error) {
	if m.IncrementFn != nil {
		return m.IncrementFn(ctx, key)
	}
	return 0, nil
}

func (m *MockCacheService) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	return true, nil
}

// Helper to create user with roles
func createUserWithRoles(t *testing.T, tenantID uuid.UUID, roles []*domain.Role) *domain.User {
	t.Helper()
	email, _ := domain.NewEmail("test@example.com")
	user, err := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	user.Activate()
	user.SetRoles(roles)
	return user
}

// Helper to create role with permissions
func createRoleWithPermissions(t *testing.T, tenantID *uuid.UUID, permissions ...string) *domain.Role {
	t.Helper()
	permSet := domain.NewPermissionSet()
	for _, p := range permissions {
		perm, _ := domain.ParsePermission(p)
		permSet.Add(perm)
	}
	role, err := domain.NewRole(tenantID, "test_role", "Test role", permSet)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}
	return role
}

// ============================================================================
// ValidatePermissionUseCase Tests
// ============================================================================

func TestValidatePermissionUseCase_Execute_Success_SinglePermission(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read", "customers:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      user.GetID(),
		TenantID:    tenant.GetID(),
		Permissions: []string{"users:read"},
		RequireAll:  true,
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("Execute() response should not be nil")
	}
	if !resp.Allowed {
		t.Error("Execute() should allow user with required permission")
	}
	if len(resp.MissingPermissions) > 0 {
		t.Errorf("Execute() should not have missing permissions, got %v", resp.MissingPermissions)
	}
}

func TestValidatePermissionUseCase_Execute_Success_MultiplePermissions_RequireAll(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read", "customers:read", "customers:create")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      user.GetID(),
		TenantID:    tenant.GetID(),
		Permissions: []string{"users:read", "customers:read"},
		RequireAll:  true,
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if !resp.Allowed {
		t.Error("Execute() should allow user with all required permissions")
	}
}

func TestValidatePermissionUseCase_Execute_Success_MultiplePermissions_RequireAny(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      user.GetID(),
		TenantID:    tenant.GetID(),
		Permissions: []string{"users:read", "customers:read"}, // Only has users:read
		RequireAll:  false,                                    // Any permission is sufficient
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if !resp.Allowed {
		t.Error("Execute() should allow user with at least one permission")
	}
}

func TestValidatePermissionUseCase_Execute_Denied_MissingPermission(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      user.GetID(),
		TenantID:    tenant.GetID(),
		Permissions: []string{"customers:delete"}, // User doesn't have this
		RequireAll:  true,
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp.Allowed {
		t.Error("Execute() should deny user without required permission")
	}
	if len(resp.MissingPermissions) == 0 {
		t.Error("Execute() should report missing permissions")
	}
}

func TestValidatePermissionUseCase_Execute_Denied_InactiveUser(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read")
	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Suspend("Test suspension") // Make user inactive
	user.SetRoles([]*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      user.GetID(),
		TenantID:    tenant.GetID(),
		Permissions: []string{"users:read"},
		RequireAll:  true,
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp.Allowed {
		t.Error("Execute() should deny inactive user")
	}
}

func TestValidatePermissionUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &MockRoleRepository{}
	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      uuid.New(),
		TenantID:    tenant.GetID(),
		Permissions: []string{"users:read"},
		RequireAll:  true,
	}

	_, err := useCase.Execute(ctx, req)

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

func TestValidatePermissionUseCase_Execute_UserDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	differentTenantID := uuid.New()
	user := createTestUser(t, differentTenantID)

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{}
	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      user.GetID(),
		TenantID:    tenant.GetID(),
		Permissions: []string{"users:read"},
		RequireAll:  true,
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when user belongs to different tenant")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("Execute() error code = %s, want FORBIDDEN", appErr.Code)
	}
}

func TestValidatePermissionUseCase_Execute_InvalidPermissionFormat(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      user.GetID(),
		TenantID:    tenant.GetID(),
		Permissions: []string{"invalid_permission"}, // Invalid format
		RequireAll:  true,
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error for invalid permission format")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("Execute() error code = %s, want VALIDATION_ERROR", appErr.Code)
	}
}

func TestValidatePermissionUseCase_CheckPermission_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	allowed, err := useCase.CheckPermission(ctx, user.GetID(), tenant.GetID(), "users:read")

	if err != nil {
		t.Fatalf("CheckPermission() unexpected error = %v", err)
	}
	if !allowed {
		t.Error("CheckPermission() should return true for user with permission")
	}
}

func TestValidatePermissionUseCase_CheckPermissions_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read", "customers:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	allowed, missing, err := useCase.CheckPermissions(ctx, user.GetID(), tenant.GetID(), []string{"users:read", "customers:read"}, true)

	if err != nil {
		t.Fatalf("CheckPermissions() unexpected error = %v", err)
	}
	if !allowed {
		t.Error("CheckPermissions() should return true for user with all permissions")
	}
	if len(missing) > 0 {
		t.Errorf("CheckPermissions() should not have missing permissions, got %v", missing)
	}
}

// ============================================================================
// AuthorizeUseCase Tests
// ============================================================================

func TestAuthorizeUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewAuthorizeUseCase(userRepo, roleRepo, tenantRepo)

	req := &AuthorizeRequest{
		UserID:   user.GetID(),
		TenantID: tenant.GetID(),
		Resource: "users",
		Action:   "read",
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("Execute() response should not be nil")
	}
	if !resp.Allowed {
		t.Error("Execute() should allow authorized request")
	}
}

func TestAuthorizeUseCase_Execute_Denied_NoPermission(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewAuthorizeUseCase(userRepo, roleRepo, tenantRepo)

	req := &AuthorizeRequest{
		UserID:   user.GetID(),
		TenantID: tenant.GetID(),
		Resource: "customers",
		Action:   "delete",
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp.Allowed {
		t.Error("Execute() should deny unauthorized request")
	}
	if resp.Reason == "" {
		t.Error("Execute() should provide denial reason")
	}
}

func TestAuthorizeUseCase_Execute_Denied_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &MockRoleRepository{}
	tenantRepo := &FullMockTenantRepository{}

	useCase := NewAuthorizeUseCase(userRepo, roleRepo, tenantRepo)

	req := &AuthorizeRequest{
		UserID:   uuid.New(),
		TenantID: tenant.GetID(),
		Resource: "users",
		Action:   "read",
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp.Allowed {
		t.Error("Execute() should deny when user not found")
	}
	if resp.Reason != "user not found" {
		t.Errorf("Execute() reason = %s, want 'user not found'", resp.Reason)
	}
}

func TestAuthorizeUseCase_Execute_Denied_InactiveUser(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Suspend("Test suspension")

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{}
	tenantRepo := &FullMockTenantRepository{}

	useCase := NewAuthorizeUseCase(userRepo, roleRepo, tenantRepo)

	req := &AuthorizeRequest{
		UserID:   user.GetID(),
		TenantID: tenant.GetID(),
		Resource: "users",
		Action:   "read",
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp.Allowed {
		t.Error("Execute() should deny inactive user")
	}
}

func TestAuthorizeUseCase_Execute_Denied_InactiveTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenant.Suspend("Test suspension")
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{}
	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewAuthorizeUseCase(userRepo, roleRepo, tenantRepo)

	req := &AuthorizeRequest{
		UserID:   user.GetID(),
		TenantID: tenant.GetID(),
		Resource: "users",
		Action:   "read",
	}

	resp, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp.Allowed {
		t.Error("Execute() should deny when tenant is inactive")
	}
}

func TestAuthorizeUseCase_Authorize_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createRoleWithPermissions(t, &tenantID, "users:read")
	user := createUserWithRoles(t, tenantID, []*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewAuthorizeUseCase(userRepo, roleRepo, tenantRepo)

	allowed, err := useCase.Authorize(ctx, user.GetID(), tenant.GetID(), "users", "read")

	if err != nil {
		t.Fatalf("Authorize() unexpected error = %v", err)
	}
	if !allowed {
		t.Error("Authorize() should return true for authorized request")
	}
}

// ============================================================================
// ValidateTokenClaimsUseCase Tests
// ============================================================================

func TestValidateTokenClaimsUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewValidateTokenClaimsUseCase(userRepo, tenantRepo)

	claims := &ports.TokenClaims{
		UserID:   user.GetID(),
		TenantID: tenant.GetID(),
		Email:    user.Email().String(),
	}

	err := useCase.Execute(ctx, claims)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestValidateTokenClaimsUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	tenantRepo := &FullMockTenantRepository{}

	useCase := NewValidateTokenClaimsUseCase(userRepo, tenantRepo)

	claims := &ports.TokenClaims{
		UserID:   uuid.New(),
		TenantID: tenant.GetID(),
	}

	err := useCase.Execute(ctx, claims)

	if err == nil {
		t.Fatal("Execute() should return error when user not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "UNAUTHORIZED" {
		t.Errorf("Execute() error code = %s, want UNAUTHORIZED", appErr.Code)
	}
}

func TestValidateTokenClaimsUseCase_Execute_UserInactive(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	user.Suspend("Test suspension")

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}

	useCase := NewValidateTokenClaimsUseCase(userRepo, tenantRepo)

	claims := &ports.TokenClaims{
		UserID:   user.GetID(),
		TenantID: tenant.GetID(),
	}

	err := useCase.Execute(ctx, claims)

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

func TestValidateTokenClaimsUseCase_Execute_TenantNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return nil, errors.New("not found")
		},
	}

	useCase := NewValidateTokenClaimsUseCase(userRepo, tenantRepo)

	claims := &ports.TokenClaims{
		UserID:   user.GetID(),
		TenantID: uuid.New(),
	}

	err := useCase.Execute(ctx, claims)

	if err == nil {
		t.Fatal("Execute() should return error when tenant not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "UNAUTHORIZED" {
		t.Errorf("Execute() error code = %s, want UNAUTHORIZED", appErr.Code)
	}
}

func TestValidateTokenClaimsUseCase_Execute_TenantInactive(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenant.Suspend("Test suspension")
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewValidateTokenClaimsUseCase(userRepo, tenantRepo)

	claims := &ports.TokenClaims{
		UserID:   user.GetID(),
		TenantID: tenant.GetID(),
	}

	err := useCase.Execute(ctx, claims)

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

// ============================================================================
// Table-Driven Tests for Permission Validation
// ============================================================================

func TestValidatePermissionUseCase_PermissionScenarios(t *testing.T) {
	tenant, _ := domain.NewTenant("Test", "test")
	tenant.Activate()
	tenantID := tenant.GetID()

	tests := []struct {
		name               string
		userPermissions    []string
		requestedPerms     []string
		requireAll         bool
		expectedAllowed    bool
		expectedMissingLen int
	}{
		{
			name:               "single_permission_granted",
			userPermissions:    []string{"users:read"},
			requestedPerms:     []string{"users:read"},
			requireAll:         true,
			expectedAllowed:    true,
			expectedMissingLen: 0,
		},
		{
			name:               "single_permission_denied",
			userPermissions:    []string{"users:read"},
			requestedPerms:     []string{"users:delete"},
			requireAll:         true,
			expectedAllowed:    false,
			expectedMissingLen: 1,
		},
		{
			name:               "multiple_all_granted",
			userPermissions:    []string{"users:read", "users:update", "customers:read"},
			requestedPerms:     []string{"users:read", "users:update"},
			requireAll:         true,
			expectedAllowed:    true,
			expectedMissingLen: 0,
		},
		{
			name:               "multiple_all_partial_denied",
			userPermissions:    []string{"users:read"},
			requestedPerms:     []string{"users:read", "users:update"},
			requireAll:         true,
			expectedAllowed:    false,
			expectedMissingLen: 1,
		},
		{
			name:               "multiple_any_granted",
			userPermissions:    []string{"users:read"},
			requestedPerms:     []string{"users:read", "users:delete"},
			requireAll:         false,
			expectedAllowed:    true,
			expectedMissingLen: 0,
		},
		{
			name:               "multiple_any_denied",
			userPermissions:    []string{"customers:read"},
			requestedPerms:     []string{"users:read", "users:delete"},
			requireAll:         false,
			expectedAllowed:    false,
			expectedMissingLen: 2,
		},
		{
			name:               "wildcard_permission",
			userPermissions:    []string{"users:*"},
			requestedPerms:     []string{"users:read", "users:delete"},
			requireAll:         true,
			expectedAllowed:    true,
			expectedMissingLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := createRoleWithPermissions(t, &tenantID, tt.userPermissions...)
			user := createUserWithRoles(t, tenantID, []*domain.Role{role})

			userRepo := &FullMockUserRepositoryForUserTests{
				FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return user, nil
				},
			}

			roleRepo := &MockRoleRepository{
				FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
					return []*domain.Role{role}, nil
				},
			}

			tenantRepo := &FullMockTenantRepository{}
			cacheService := &MockCacheService{}

			useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

			req := &ValidatePermissionRequest{
				UserID:      user.GetID(),
				TenantID:    tenantID,
				Permissions: tt.requestedPerms,
				RequireAll:  tt.requireAll,
			}

			resp, err := useCase.Execute(context.Background(), req)

			if err != nil {
				t.Fatalf("Execute() unexpected error = %v", err)
			}
			if resp.Allowed != tt.expectedAllowed {
				t.Errorf("Execute() Allowed = %v, want %v", resp.Allowed, tt.expectedAllowed)
			}
			if len(resp.MissingPermissions) != tt.expectedMissingLen {
				t.Errorf("Execute() MissingPermissions len = %d, want %d", len(resp.MissingPermissions), tt.expectedMissingLen)
			}
		})
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkValidatePermissionUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	permSet := domain.NewPermissionSet()
	permSet.Add(domain.PermissionUsersRead)
	permSet.Add(domain.PermissionCustomersRead)
	role, _ := domain.NewRole(&tenantID, "test_role", "Test", permSet)

	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Activate()
	user.SetRoles([]*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{}
	cacheService := &MockCacheService{}

	useCase := NewValidatePermissionUseCase(userRepo, roleRepo, tenantRepo, cacheService)

	req := &ValidatePermissionRequest{
		UserID:      user.GetID(),
		TenantID:    tenantID,
		Permissions: []string{"users:read"},
		RequireAll:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, req)
	}
}

func BenchmarkAuthorizeUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenant, _ := domain.NewTenant("Test", "test")
	tenant.Activate()
	tenantID := tenant.GetID()

	permSet := domain.NewPermissionSet()
	permSet.Add(domain.PermissionUsersRead)
	role, _ := domain.NewRole(&tenantID, "test_role", "Test", permSet)

	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Activate()
	user.SetRoles([]*domain.Role{role})

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewAuthorizeUseCase(userRepo, roleRepo, tenantRepo)

	req := &AuthorizeRequest{
		UserID:   user.GetID(),
		TenantID: tenantID,
		Resource: "users",
		Action:   "read",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, req)
	}
}

func BenchmarkValidateTokenClaimsUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenant, _ := domain.NewTenant("Test", "test")
	tenant.Activate()

	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenant.GetID(), email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Activate()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	tenantRepo := &FullMockTenantRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
			return tenant, nil
		},
	}

	useCase := NewValidateTokenClaimsUseCase(userRepo, tenantRepo)

	claims := &ports.TokenClaims{
		UserID:   user.GetID(),
		TenantID: tenant.GetID(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = useCase.Execute(ctx, claims)
	}
}
