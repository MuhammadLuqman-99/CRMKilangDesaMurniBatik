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
// Extended Mock Implementations for Assign Role Tests
// ============================================================================

// ExtendedMockRoleRepository extends MockRoleRepository with additional functions.
type ExtendedMockRoleRepository struct {
	MockRoleRepository
	FindByIDFn             func(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	AssignRoleToUserFn     func(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error
	RemoveRoleFromUserFn   func(ctx context.Context, userID, roleID uuid.UUID) error
}

func (m *ExtendedMockRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *ExtendedMockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
	if m.AssignRoleToUserFn != nil {
		return m.AssignRoleToUserFn(ctx, userID, roleID, assignedBy)
	}
	return nil
}

func (m *ExtendedMockRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.RemoveRoleFromUserFn != nil {
		return m.RemoveRoleFromUserFn(ctx, userID, roleID)
	}
	return nil
}

// MockOutboxRepository is a mock implementation of domain.OutboxRepository.
type MockOutboxRepository struct {
	CreateFn func(ctx context.Context, entry *domain.OutboxEntry) error
}

func (m *MockOutboxRepository) Create(ctx context.Context, entry *domain.OutboxEntry) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, entry)
	}
	return nil
}

func (m *MockOutboxRepository) MarkAsPublished(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockOutboxRepository) FindUnpublished(ctx context.Context, limit int) ([]*domain.OutboxEntry, error) {
	return nil, nil
}

func (m *MockOutboxRepository) DeletePublished(ctx context.Context, olderThan interface{}) (int64, error) {
	return 0, nil
}

// ============================================================================
// Test Helpers for Assign Role Tests
// ============================================================================

func createTestRole(t *testing.T, tenantID *uuid.UUID, name string) *domain.Role {
	t.Helper()
	permissions := domain.NewPermissionSet()
	permissions.Add(domain.PermissionUsersRead)
	role, err := domain.NewRole(tenantID, name, "Test role description", permissions)
	if err != nil {
		t.Fatalf("Failed to create test role: %v", err)
	}
	return role
}

func createSystemRole(t *testing.T) *domain.Role {
	t.Helper()
	return domain.CreateAdminRole()
}

// ============================================================================
// AssignRoleUseCase Tests
// ============================================================================

func TestAssignRoleUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	assignedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			if id == user.GetID() {
				return user, nil
			}
			return nil, errors.New("not found")
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			if id == role.GetID() {
				return role, nil
			}
			return nil, errors.New("not found")
		},
		MockRoleRepository: MockRoleRepository{
			FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
				return []*domain.Role{}, nil
			},
		},
		AssignRoleToUserFn: func(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.AssignRoleRequest{
		UserID: user.GetID(),
		RoleID: role.GetID(),
	}

	resp, err := useCase.Execute(ctx, req, &assignedBy, tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("Execute() response should not be nil")
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestAssignRoleUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	assignedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &ExtendedMockRoleRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.AssignRoleRequest{
		UserID: uuid.New(),
		RoleID: uuid.New(),
	}

	_, err := useCase.Execute(ctx, req, &assignedBy, tenant.GetID())

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

func TestAssignRoleUseCase_Execute_UserDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	differentTenantID := uuid.New()
	user := createTestUser(t, differentTenantID) // User belongs to different tenant
	assignedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.AssignRoleRequest{
		UserID: user.GetID(),
		RoleID: uuid.New(),
	}

	_, err := useCase.Execute(ctx, req, &assignedBy, tenant.GetID())

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

func TestAssignRoleUseCase_Execute_RoleNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	assignedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.AssignRoleRequest{
		UserID: user.GetID(),
		RoleID: uuid.New(),
	}

	_, err := useCase.Execute(ctx, req, &assignedBy, tenant.GetID())

	if err == nil {
		t.Fatal("Execute() should return error when role not found")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

func TestAssignRoleUseCase_Execute_RoleDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	differentTenantID := uuid.New()
	role := createTestRole(t, &differentTenantID, "test_role") // Role belongs to different tenant
	assignedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.AssignRoleRequest{
		UserID: user.GetID(),
		RoleID: role.GetID(),
	}

	_, err := useCase.Execute(ctx, req, &assignedBy, tenant.GetID())

	if err == nil {
		t.Fatal("Execute() should return error when role belongs to different tenant")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("Execute() error code = %s, want FORBIDDEN", appErr.Code)
	}
}

func TestAssignRoleUseCase_Execute_RoleAlreadyAssigned(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	assignedBy := uuid.New()

	// Pre-assign the role to the user
	user.SetRoles([]*domain.Role{role})

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.AssignRoleRequest{
		UserID: user.GetID(),
		RoleID: role.GetID(),
	}

	_, err := useCase.Execute(ctx, req, &assignedBy, tenant.GetID())

	if err == nil {
		t.Fatal("Execute() should return error when role is already assigned")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "CONFLICT" {
		t.Errorf("Execute() error code = %s, want CONFLICT", appErr.Code)
	}
}

func TestAssignRoleUseCase_Execute_SystemRole(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	systemRole := createSystemRole(t) // System role (no tenant)
	assignedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return systemRole, nil
		},
		MockRoleRepository: MockRoleRepository{
			FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
				return []*domain.Role{}, nil
			},
		},
		AssignRoleToUserFn: func(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.AssignRoleRequest{
		UserID: user.GetID(),
		RoleID: systemRole.GetID(),
	}

	// System roles should be assignable to any tenant's users
	resp, err := useCase.Execute(ctx, req, &assignedBy, tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("Execute() response should not be nil")
	}
}

func TestAssignRoleUseCase_Execute_TransactionError(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	assignedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
		MockRoleRepository: MockRoleRepository{
			FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
				return []*domain.Role{}, nil
			},
		},
		AssignRoleToUserFn: func(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
			return errors.New("database error")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.AssignRoleRequest{
		UserID: user.GetID(),
		RoleID: role.GetID(),
	}

	_, err := useCase.Execute(ctx, req, &assignedBy, tenant.GetID())

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
// RemoveRoleUseCase Tests
// ============================================================================

func TestRemoveRoleUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	removedBy := uuid.New()

	// Pre-assign the role to the user
	user.SetRoles([]*domain.Role{role})

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
		MockRoleRepository: MockRoleRepository{
			FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
				return []*domain.Role{role}, nil
			},
		},
		RemoveRoleFromUserFn: func(ctx context.Context, userID, roleID uuid.UUID) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRemoveRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.RemoveRoleRequest{
		UserID: user.GetID(),
		RoleID: role.GetID(),
	}

	resp, err := useCase.Execute(ctx, req, &removedBy, tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if resp == nil {
		t.Fatal("Execute() response should not be nil")
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestRemoveRoleUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	removedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &ExtendedMockRoleRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRemoveRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.RemoveRoleRequest{
		UserID: uuid.New(),
		RoleID: uuid.New(),
	}

	_, err := useCase.Execute(ctx, req, &removedBy, tenant.GetID())

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

func TestRemoveRoleUseCase_Execute_RoleNotAssigned(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	removedBy := uuid.New()

	// User has no roles assigned
	user.SetRoles([]*domain.Role{})

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRemoveRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.RemoveRoleRequest{
		UserID: user.GetID(),
		RoleID: role.GetID(),
	}

	_, err := useCase.Execute(ctx, req, &removedBy, tenant.GetID())

	if err == nil {
		t.Fatal("Execute() should return error when role is not assigned")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "NOT_FOUND" {
		t.Errorf("Execute() error code = %s, want NOT_FOUND", appErr.Code)
	}
}

func TestRemoveRoleUseCase_Execute_UserDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	differentTenantID := uuid.New()
	user := createTestUser(t, differentTenantID)
	removedBy := uuid.New()

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRemoveRoleUseCase(
		userRepo,
		roleRepo,
		outboxRepo,
		txManager,
		auditLogger,
	)

	req := &dto.RemoveRoleRequest{
		UserID: user.GetID(),
		RoleID: uuid.New(),
	}

	_, err := useCase.Execute(ctx, req, &removedBy, tenant.GetID())

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
// GetUserRolesUseCase Tests
// ============================================================================

func TestGetUserRolesUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	tenantID := tenant.GetID()
	role1 := createTestRole(t, &tenantID, "role1")
	role2 := createTestRole(t, &tenantID, "role2")

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role1, role2}, nil
		},
	}

	useCase := NewGetUserRolesUseCase(userRepo, roleRepo)

	roles, err := useCase.Execute(ctx, user.GetID(), tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("Execute() returned %d roles, want 2", len(roles))
	}
}

func TestGetUserRolesUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &MockRoleRepository{}

	useCase := NewGetUserRolesUseCase(userRepo, roleRepo)

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

func TestGetUserRolesUseCase_Execute_UserDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	differentTenantID := uuid.New()
	user := createTestUser(t, differentTenantID)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{}

	useCase := NewGetUserRolesUseCase(userRepo, roleRepo)

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
// GetUserPermissionsUseCase Tests
// ============================================================================

func TestGetUserPermissionsUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	user := createTestUser(t, tenant.GetID())
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role}, nil
		},
	}

	useCase := NewGetUserPermissionsUseCase(userRepo, roleRepo)

	permissions, err := useCase.Execute(ctx, user.GetID(), tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(permissions) == 0 {
		t.Error("Execute() should return permissions")
	}
}

func TestGetUserPermissionsUseCase_Execute_UserNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return nil, errors.New("not found")
		},
	}

	roleRepo := &MockRoleRepository{}

	useCase := NewGetUserPermissionsUseCase(userRepo, roleRepo)

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

// ============================================================================
// Table-Driven Tests for Assign Role Validation
// ============================================================================

func TestAssignRoleUseCase_ValidationScenarios(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func() (*MockUserRepository, *ExtendedMockRoleRepository)
		tenantID       uuid.UUID
		request        *dto.AssignRoleRequest
		wantErr        bool
		expectedErrCode string
	}{
		{
			name: "empty_user_id",
			setupMocks: func() (*MockUserRepository, *ExtendedMockRoleRepository) {
				return &MockUserRepository{
					FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
						return nil, errors.New("not found")
					},
				}, &ExtendedMockRoleRepository{}
			},
			tenantID: uuid.New(),
			request: &dto.AssignRoleRequest{
				UserID: uuid.Nil,
				RoleID: uuid.New(),
			},
			wantErr:        true,
			expectedErrCode: "NOT_FOUND",
		},
		{
			name: "empty_role_id",
			setupMocks: func() (*MockUserRepository, *ExtendedMockRoleRepository) {
				tenant := uuid.New()
				email, _ := domain.NewEmail("test@example.com")
				user, _ := domain.NewUser(tenant, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
				user.Activate()
				return &MockUserRepository{
					FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
						return user, nil
					},
				}, &ExtendedMockRoleRepository{
					FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
						return nil, errors.New("not found")
					},
				}
			},
			tenantID: uuid.New(),
			request: &dto.AssignRoleRequest{
				UserID: uuid.New(),
				RoleID: uuid.Nil,
			},
			wantErr:        true,
			expectedErrCode: "FORBIDDEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo, roleRepo := tt.setupMocks()

			useCase := NewAssignRoleUseCase(
				userRepo,
				roleRepo,
				&MockOutboxRepository{},
				&MockTransactionManager{},
				&MockAuditLogger{},
			)

			assignedBy := uuid.New()
			_, err := useCase.Execute(context.Background(), tt.request, &assignedBy, tt.tenantID)

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

func BenchmarkAssignRoleUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Activate()

	permissions := domain.NewPermissionSet()
	role, _ := domain.NewRole(&tenantID, "test_role", "Test", permissions)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &ExtendedMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
		MockRoleRepository: MockRoleRepository{
			FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
				return []*domain.Role{}, nil
			},
		},
		AssignRoleToUserFn: func(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
			return nil
		},
	}

	useCase := NewAssignRoleUseCase(
		userRepo,
		roleRepo,
		&MockOutboxRepository{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	assignedBy := uuid.New()
	req := &dto.AssignRoleRequest{
		UserID: user.GetID(),
		RoleID: role.GetID(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset user roles for each iteration
		user.SetRoles([]*domain.Role{})
		_, _ = useCase.Execute(ctx, req, &assignedBy, tenantID)
	}
}

func BenchmarkGetUserRolesUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	email, _ := domain.NewEmail("test@example.com")
	user, _ := domain.NewUser(tenantID, email, domain.NewPasswordFromHash("hash"), "John", "Doe")
	user.Activate()

	permissions := domain.NewPermissionSet()
	role1, _ := domain.NewRole(&tenantID, "role1", "Test", permissions)
	role2, _ := domain.NewRole(&tenantID, "role2", "Test", permissions)

	userRepo := &MockUserRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return user, nil
		},
	}

	roleRepo := &MockRoleRepository{
		FindByUserIDFn: func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
			return []*domain.Role{role1, role2}, nil
		},
	}

	useCase := NewGetUserRolesUseCase(userRepo, roleRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, user.GetID(), tenantID)
	}
}
