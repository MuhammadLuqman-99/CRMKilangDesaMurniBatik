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
// Extended Mock Implementations for User Tests
// ============================================================================

// FullMockUserRepositoryForUserTests provides complete mock for user repository.
type FullMockUserRepositoryForUserTests struct {
	CreateFn        func(ctx context.Context, user *domain.User) error
	UpdateFn        func(ctx context.Context, user *domain.User) error
	DeleteFn        func(ctx context.Context, id uuid.UUID) error
	FindByIDFn      func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmailFn   func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error)
	FindByTenantFn  func(ctx context.Context, tenantID uuid.UUID, opts domain.UserQueryOptions) ([]*domain.User, int64, error)
	FindByRoleIDFn  func(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error)
	ExistsByEmailFn func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error)
	CountByTenantFn func(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

func (m *FullMockUserRepositoryForUserTests) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

func (m *FullMockUserRepositoryForUserTests) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return nil
}

func (m *FullMockUserRepositoryForUserTests) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *FullMockUserRepositoryForUserTests) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *FullMockUserRepositoryForUserTests) FindByEmail(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(ctx, tenantID, email)
	}
	return nil, errors.New("not found")
}

func (m *FullMockUserRepositoryForUserTests) FindByTenant(ctx context.Context, tenantID uuid.UUID, opts domain.UserQueryOptions) ([]*domain.User, int64, error) {
	if m.FindByTenantFn != nil {
		return m.FindByTenantFn(ctx, tenantID, opts)
	}
	return nil, 0, nil
}

func (m *FullMockUserRepositoryForUserTests) FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
	if m.FindByRoleIDFn != nil {
		return m.FindByRoleIDFn(ctx, roleID)
	}
	return nil, nil
}

func (m *FullMockUserRepositoryForUserTests) ExistsByEmail(ctx context.Context, tenantID uuid.UUID, email domain.Email) (bool, error) {
	if m.ExistsByEmailFn != nil {
		return m.ExistsByEmailFn(ctx, tenantID, email)
	}
	return false, nil
}

func (m *FullMockUserRepositoryForUserTests) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	if m.CountByTenantFn != nil {
		return m.CountByTenantFn(ctx, tenantID)
	}
	return 0, nil
}

// ============================================================================
// GetUserUseCase Tests
// ============================================================================

func TestGetUserUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{}, nil
		},
	}

	useCase := NewGetUserUseCase(userRepo, roleRepo)

	result, err := useCase.Execute(ctx, user.GetID(), tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if result.Email != user.Email().String() {
		t.Errorf("Execute() email = %s, want %s", result.Email, user.Email().String())
	}
}

func TestGetUserUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &MockRoleRepository{}

	useCase := NewGetUserUseCase(userRepo, roleRepo)

	_, err := useCase.Execute(ctx, uuid.New(), tenant.GetID())

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

func TestGetUserUseCase_Execute_UserDifferentTenant(t *testing.T) {
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

	useCase := NewGetUserUseCase(userRepo, roleRepo)

	_, err := useCase.Execute(ctx, user.GetID(), tenant.GetID())

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

// ============================================================================
// GetUserByEmailUseCase Tests
// ============================================================================

func TestGetUserByEmailUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{}, nil
		},
	}

	useCase := NewGetUserByEmailUseCase(userRepo, roleRepo)

	result, err := useCase.Execute(ctx, "test@example.com", tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
}

func TestGetUserByEmailUseCase_Execute_InvalidEmail(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &FullMockUserRepositoryForUserTests{}
	roleRepo := &MockRoleRepository{}

	useCase := NewGetUserByEmailUseCase(userRepo, roleRepo)

	_, err := useCase.Execute(ctx, "invalid-email", tenant.GetID())

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

func TestGetUserByEmailUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email domain.Email) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &MockRoleRepository{}

	useCase := NewGetUserByEmailUseCase(userRepo, roleRepo)

	_, err := useCase.Execute(ctx, "notfound@example.com", tenant.GetID())

	if err == nil {
		t.Fatal("Execute() should return error when user not found")
	}
}

// ============================================================================
// ListUsersUseCase Tests
// ============================================================================

func TestListUsersUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user1 := createTestUser(t, tenant.GetID())
	email2, _ := domain.NewEmail("user2@example.com")
	user2, _ := domain.NewUser(tenant.GetID(), email2, domain.NewPasswordFromHash("hash"), "Jane", "Doe")
	user2.Activate()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByTenantFn: func(ctx context.Context, tenantID uuid.UUID, opts domain.UserQueryOptions) ([]*domain.User, int64, error) {
			return []*domain.User{user1, user2}, 2, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{}, nil
		},
	}

	useCase := NewListUsersUseCase(userRepo, roleRepo)

	req := &dto.ListUsersRequest{
		TenantID:     tenant.GetID(),
		Page:         1,
		PageSize:     20,
		IncludeRoles: true,
	}

	result, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if len(result.Users) != 2 {
		t.Errorf("Execute() returned %d users, want 2", len(result.Users))
	}
	if result.Pagination.TotalItems != 2 {
		t.Errorf("Execute() TotalItems = %d, want 2", result.Pagination.TotalItems)
	}
}

func TestListUsersUseCase_Execute_WithDefaults(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByTenantFn: func(ctx context.Context, tenantID uuid.UUID, opts domain.UserQueryOptions) ([]*domain.User, int64, error) {
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
			return []*domain.User{}, 0, nil
		},
	}

	roleRepo := &MockRoleRepository{}

	useCase := NewListUsersUseCase(userRepo, roleRepo)

	req := &dto.ListUsersRequest{
		TenantID: tenant.GetID(),
	}

	_, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestListUsersUseCase_Execute_RepositoryError(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByTenantFn: func(ctx context.Context, tenantID uuid.UUID, opts domain.UserQueryOptions) ([]*domain.User, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	roleRepo := &MockRoleRepository{}

	useCase := NewListUsersUseCase(userRepo, roleRepo)

	req := &dto.ListUsersRequest{
		TenantID: tenant.GetID(),
		Page:     1,
		PageSize: 20,
	}

	_, err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error when repository fails")
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
// UpdateUserUseCase Tests
// ============================================================================

func TestUpdateUserUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	updatedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
		UpdateFn: func(ctx context.Context, u *domain.User) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateUserUseCase(userRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateUserRequest{
		FirstName: "Updated",
		LastName:  "Name",
		Phone:     "+1234567890",
	}

	result, err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), req, &updatedBy)

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

func TestUpdateUserUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	updatedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateUserUseCase(userRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateUserRequest{
		FirstName: "Updated",
	}

	_, err := useCase.Execute(ctx, uuid.New(), tenant.GetID(), req, &updatedBy)

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

func TestUpdateUserUseCase_Execute_UserDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	differentTenantID := uuid.New()
	user := createTestUser(t, differentTenantID)
	updatedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateUserUseCase(userRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateUserRequest{
		FirstName: "Updated",
	}

	_, err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), req, &updatedBy)

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

// ============================================================================
// ChangePasswordUseCase Tests
// ============================================================================

func TestChangePasswordUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
		UpdateFn: func(ctx context.Context, u *domain.User) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{
		VerifyFn: func(password, hash string) (bool, error) {
			return password == "password123", nil
		},
		HashFn: func(password string) (string, error) {
			return "hashed_" + password, nil
		},
	}
	emailService := &MockEmailService{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewChangePasswordUseCase(
		userRepo,
		outboxRepo,
		passwordHasher,
		emailService,
		txManager,
		auditLogger,
	)

	req := &dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "NewPassword123!",
	}

	err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestChangePasswordUseCase_Execute_WrongCurrentPassword(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{
		VerifyFn: func(password, hash string) (bool, error) {
			return false, nil // Password doesn't match
		},
	}
	emailService := &MockEmailService{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewChangePasswordUseCase(
		userRepo,
		outboxRepo,
		passwordHasher,
		emailService,
		txManager,
		auditLogger,
	)

	req := &dto.ChangePasswordRequest{
		CurrentPassword: "wrong_password",
		NewPassword:     "NewPassword123!",
	}

	err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), req)

	if err == nil {
		t.Fatal("Execute() should return error for wrong current password")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("Execute() error code = %s, want FORBIDDEN", appErr.Code)
	}
}

func TestChangePasswordUseCase_Execute_WeakNewPassword(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	passwordHasher := &MockPasswordHasher{
		VerifyFn: func(password, hash string) (bool, error) {
			return true, nil
		},
	}
	emailService := &MockEmailService{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewChangePasswordUseCase(
		userRepo,
		outboxRepo,
		passwordHasher,
		emailService,
		txManager,
		auditLogger,
	)

	req := &dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "weak", // Too weak
	}

	err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), req)

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

// ============================================================================
// DeleteUserUseCase Tests
// ============================================================================

func TestDeleteUserUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	deletedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		RevokeByUserIDFn: func(ctx context.Context, userID uuid.UUID) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteUserUseCase(userRepo, refreshTokenRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), &deletedBy)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestDeleteUserUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	deletedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteUserUseCase(userRepo, refreshTokenRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, uuid.New(), tenant.GetID(), &deletedBy)

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

func TestDeleteUserUseCase_Execute_UserDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	differentTenantID := uuid.New()
	user := createTestUser(t, differentTenantID)
	deletedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteUserUseCase(userRepo, refreshTokenRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), &deletedBy)

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

// ============================================================================
// SuspendUserUseCase Tests
// ============================================================================

func TestSuspendUserUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	suspendedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
		UpdateFn: func(ctx context.Context, u *domain.User) error {
			return nil
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{
		RevokeByUserIDFn: func(ctx context.Context, userID uuid.UUID) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewSuspendUserUseCase(userRepo, refreshTokenRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), "Test suspension", &suspendedBy)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestSuspendUserUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	suspendedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	refreshTokenRepo := &MockRefreshTokenRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewSuspendUserUseCase(userRepo, refreshTokenRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, uuid.New(), tenant.GetID(), "Test suspension", &suspendedBy)

	if err == nil {
		t.Fatal("Execute() should return error when user not found")
	}
}

// ============================================================================
// ActivateUserUseCase Tests
// ============================================================================

func TestActivateUserUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	user.Suspend("Test") // Suspend first
	activatedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
		UpdateFn: func(ctx context.Context, u *domain.User) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewActivateUserUseCase(userRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, user.GetID(), tenant.GetID(), &activatedBy)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestActivateUserUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	activatedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewActivateUserUseCase(userRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, uuid.New(), tenant.GetID(), &activatedBy)

	if err == nil {
		t.Fatal("Execute() should return error when user not found")
	}
}

// ============================================================================
// VerifyEmailUseCase Tests
// ============================================================================

func TestVerifyEmailUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
		UpdateFn: func(ctx context.Context, u *domain.User) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	verifyService := &MockVerificationTokenService{
		ValidateEmailVerificationTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			return user.GetID(), nil
		},
	}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewVerifyEmailUseCase(userRepo, outboxRepo, verifyService, txManager, auditLogger)

	req := &dto.VerifyEmailRequest{
		Token: "valid_token",
	}

	err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestVerifyEmailUseCase_Execute_InvalidToken(t *testing.T) {
	ctx := context.Background()

	userRepo := &FullMockUserRepositoryForUserTests{}
	outboxRepo := &MockOutboxRepository{}
	verifyService := &MockVerificationTokenService{
		ValidateEmailVerificationTokenFn: func(ctx context.Context, token string) (uuid.UUID, error) {
			return uuid.Nil, errors.New("invalid token")
		},
	}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewVerifyEmailUseCase(userRepo, outboxRepo, verifyService, txManager, auditLogger)

	req := &dto.VerifyEmailRequest{
		Token: "invalid_token",
	}

	err := useCase.Execute(ctx, req)

	if err == nil {
		t.Fatal("Execute() should return error for invalid token")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "TOKEN_INVALID" {
		t.Errorf("Execute() error code = %s, want TOKEN_INVALID", appErr.Code)
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkGetUserUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Activate()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{}, nil
		},
	}

	useCase := NewGetUserUseCase(userRepo, roleRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, user.GetID(), tenantID)
	}
}

func BenchmarkListUsersUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	email1, _ := domain.NewEmail("user1@example.com")
	user1, _ := domain.NewUser(tenantID, email1, domain.NewPasswordFromHash("hash"), "John", "Doe")
	email2, _ := domain.NewEmail("user2@example.com")
	user2, _ := domain.NewUser(tenantID, email2, domain.NewPasswordFromHash("hash"), "Jane", "Doe")

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByTenantFn: func(ctx context.Context, tid uuid.UUID, opts domain.UserQueryOptions) ([]*domain.User, int64, error) {
			return []*domain.User{user1, user2}, 2, nil
		},
	}

	roleRepo := &MockRoleRepository{}

	useCase := NewListUsersUseCase(userRepo, roleRepo)

	req := &dto.ListUsersRequest{
		TenantID: tenantID,
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, req)
	}
}

func BenchmarkUpdateUserUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Activate()
	updatedBy := uuid.New()

	userRepo := &FullMockUserRepositoryForUserTests{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
		UpdateFn: func(ctx context.Context, u *domain.User) error {
			return nil
		},
	}

	useCase := NewUpdateUserUseCase(
		userRepo,
		&MockOutboxRepository{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.UpdateUserRequest{
		FirstName: "Updated",
		LastName:  "Name",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, user.GetID(), tenantID, req, &updatedBy)
	}
}
