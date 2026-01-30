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
// Mock Implementations for Role Tests
// ============================================================================

// FullMockRoleRepository provides complete mock for role repository.
type FullMockRoleRepository struct {
	CreateFn             func(ctx context.Context, role *domain.Role) error
	UpdateFn             func(ctx context.Context, role *domain.Role) error
	DeleteFn             func(ctx context.Context, id uuid.UUID) error
	FindByIDFn           func(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	FindByNameFn         func(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error)
	FindByTenantFn       func(ctx context.Context, tenantID uuid.UUID, opts domain.RoleQueryOptions) ([]*domain.Role, int64, error)
	FindSystemRolesFn    func(ctx context.Context) ([]*domain.Role, error)
	FindByUserIDFn       func(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
	ExistsByNameFn       func(ctx context.Context, tenantID *uuid.UUID, name string) (bool, error)
	AssignRoleToUserFn   func(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error
	RemoveRoleFromUserFn func(ctx context.Context, userID, roleID uuid.UUID) error
}

func (m *FullMockRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, role)
	}
	return nil
}

func (m *FullMockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, role)
	}
	return nil
}

func (m *FullMockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *FullMockRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *FullMockRoleRepository) FindByName(ctx context.Context, tenantID *uuid.UUID, name string) (*domain.Role, error) {
	if m.FindByNameFn != nil {
		return m.FindByNameFn(ctx, tenantID, name)
	}
	return nil, errors.New("not found")
}

func (m *FullMockRoleRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID, opts domain.RoleQueryOptions) ([]*domain.Role, int64, error) {
	if m.FindByTenantFn != nil {
		return m.FindByTenantFn(ctx, tenantID, opts)
	}
	return nil, 0, nil
}

func (m *FullMockRoleRepository) FindSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	if m.FindSystemRolesFn != nil {
		return m.FindSystemRolesFn(ctx)
	}
	return nil, nil
}

func (m *FullMockRoleRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	if m.FindByUserIDFn != nil {
		return m.FindByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *FullMockRoleRepository) ExistsByName(ctx context.Context, tenantID *uuid.UUID, name string) (bool, error) {
	if m.ExistsByNameFn != nil {
		return m.ExistsByNameFn(ctx, tenantID, name)
	}
	return false, nil
}

func (m *FullMockRoleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy *uuid.UUID) error {
	if m.AssignRoleToUserFn != nil {
		return m.AssignRoleToUserFn(ctx, userID, roleID, assignedBy)
	}
	return nil
}

func (m *FullMockRoleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if m.RemoveRoleFromUserFn != nil {
		return m.RemoveRoleFromUserFn(ctx, userID, roleID)
	}
	return nil
}

// FullMockUserRepository for role tests.
type FullMockUserRepository struct {
	MockUserRepository
	FindByRoleIDFn func(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error)
}

func (m *FullMockUserRepository) FindByRoleID(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
	if m.FindByRoleIDFn != nil {
		return m.FindByRoleIDFn(ctx, roleID)
	}
	return nil, nil
}

// ============================================================================
// GetRoleUseCase Tests
// ============================================================================

func TestGetRoleUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	userRepo := &FullMockUserRepository{
		FindByRoleIDFn: func(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
			return []*domain.User{}, nil
		},
	}

	useCase := NewGetRoleUseCase(roleRepo, userRepo)

	result, err := useCase.Execute(ctx, role.GetID(), tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if result.Name != role.Name() {
		t.Errorf("Execute() role name = %s, want %s", result.Name, role.Name())
	}
}

func TestGetRoleUseCase_Execute_RoleNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return nil, errors.New("not found")
		},
	}

	userRepo := &FullMockUserRepository{}

	useCase := NewGetRoleUseCase(roleRepo, userRepo)

	_, err := useCase.Execute(ctx, uuid.New(), tenant.GetID())

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

func TestGetRoleUseCase_Execute_RoleDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	differentTenantID := uuid.New()
	role := createTestRole(t, &differentTenantID, "test_role")

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	userRepo := &FullMockUserRepository{}

	useCase := NewGetRoleUseCase(roleRepo, userRepo)

	_, err := useCase.Execute(ctx, role.GetID(), tenant.GetID())

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

func TestGetRoleUseCase_Execute_SystemRole(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	systemRole := createSystemRole(t)

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return systemRole, nil
		},
	}

	userRepo := &FullMockUserRepository{
		FindByRoleIDFn: func(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
			return []*domain.User{}, nil
		},
	}

	useCase := NewGetRoleUseCase(roleRepo, userRepo)

	// System roles should be accessible by any tenant
	result, err := useCase.Execute(ctx, systemRole.GetID(), tenant.GetID())

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if !result.IsSystem {
		t.Error("Execute() IsSystem should be true")
	}
}

// ============================================================================
// ListRolesUseCase Tests
// ============================================================================

func TestListRolesUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role1 := createTestRole(t, &tenantID, "role1")
	role2 := createTestRole(t, &tenantID, "role2")

	roleRepo := &FullMockRoleRepository{
		FindByTenantFn: func(ctx context.Context, tid uuid.UUID, opts domain.RoleQueryOptions) ([]*domain.Role, int64, error) {
			return []*domain.Role{role1, role2}, 2, nil
		},
	}

	useCase := NewListRolesUseCase(roleRepo)

	req := &dto.ListRolesRequest{
		TenantID:      tenant.GetID(),
		Page:          1,
		PageSize:      20,
		IncludeSystem: false,
	}

	result, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if len(result.Roles) != 2 {
		t.Errorf("Execute() returned %d roles, want 2", len(result.Roles))
	}
	if result.Pagination.TotalItems != 2 {
		t.Errorf("Execute() TotalItems = %d, want 2", result.Pagination.TotalItems)
	}
}

func TestListRolesUseCase_Execute_WithDefaults(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	roleRepo := &FullMockRoleRepository{
		FindByTenantFn: func(ctx context.Context, tid uuid.UUID, opts domain.RoleQueryOptions) ([]*domain.Role, int64, error) {
			// Verify defaults were applied
			if opts.Page != 1 {
				t.Errorf("Expected page 1, got %d", opts.Page)
			}
			if opts.PageSize != 20 {
				t.Errorf("Expected pageSize 20, got %d", opts.PageSize)
			}
			if opts.SortBy != "name" {
				t.Errorf("Expected sortBy 'name', got %s", opts.SortBy)
			}
			if opts.SortDirection != "asc" {
				t.Errorf("Expected sortDirection 'asc', got %s", opts.SortDirection)
			}
			return []*domain.Role{}, 0, nil
		},
	}

	useCase := NewListRolesUseCase(roleRepo)

	req := &dto.ListRolesRequest{
		TenantID: tenant.GetID(),
		// Leave other fields at zero values
	}

	_, err := useCase.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestListRolesUseCase_Execute_RepositoryError(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)

	roleRepo := &FullMockRoleRepository{
		FindByTenantFn: func(ctx context.Context, tid uuid.UUID, opts domain.RoleQueryOptions) ([]*domain.Role, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	useCase := NewListRolesUseCase(roleRepo)

	req := &dto.ListRolesRequest{
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
// ListSystemRolesUseCase Tests
// ============================================================================

func TestListSystemRolesUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	adminRole := domain.CreateAdminRole()
	viewerRole := domain.CreateViewerRole()

	roleRepo := &FullMockRoleRepository{
		FindSystemRolesFn: func(ctx context.Context) ([]*domain.Role, error) {
			return []*domain.Role{adminRole, viewerRole}, nil
		},
	}

	useCase := NewListSystemRolesUseCase(roleRepo)

	result, err := useCase.Execute(ctx)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Execute() returned %d roles, want 2", len(result))
	}
}

func TestListSystemRolesUseCase_Execute_RepositoryError(t *testing.T) {
	ctx := context.Background()

	roleRepo := &FullMockRoleRepository{
		FindSystemRolesFn: func(ctx context.Context) ([]*domain.Role, error) {
			return nil, errors.New("database error")
		},
	}

	useCase := NewListSystemRolesUseCase(roleRepo)

	_, err := useCase.Execute(ctx)

	if err == nil {
		t.Fatal("Execute() should return error when repository fails")
	}
}

// ============================================================================
// ListAvailablePermissionsUseCase Tests
// ============================================================================

func TestListAvailablePermissionsUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()

	useCase := NewListAvailablePermissionsUseCase()

	result := useCase.Execute(ctx)

	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if len(result.Groups) == 0 {
		t.Error("Execute() should return permission groups")
	}

	// Verify some expected groups exist
	foundUsers := false
	foundRoles := false
	for _, group := range result.Groups {
		if group.Resource == "users" {
			foundUsers = true
		}
		if group.Resource == "roles" {
			foundRoles = true
		}
	}

	if !foundUsers {
		t.Error("Execute() should include 'users' permission group")
	}
	if !foundRoles {
		t.Error("Execute() should include 'roles' permission group")
	}
}

// ============================================================================
// CreateRoleUseCase Tests
// ============================================================================

func TestCreateRoleUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	createdBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		ExistsByNameFn: func(ctx context.Context, tenantID *uuid.UUID, name string) (bool, error) {
			return false, nil
		},
		CreateFn: func(ctx context.Context, role *domain.Role) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewCreateRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.CreateRoleRequest{
		TenantID:    tenant.GetID(),
		Name:        "new_role",
		Description: "A new role",
		Permissions: []string{"users:read", "customers:read"},
	}

	result, err := useCase.Execute(ctx, req, &createdBy)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
	if result.Role.Name != "new_role" {
		t.Errorf("Execute() role name = %s, want new_role", result.Role.Name)
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestCreateRoleUseCase_Execute_NameAlreadyExists(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	createdBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		ExistsByNameFn: func(ctx context.Context, tenantID *uuid.UUID, name string) (bool, error) {
			return true, nil // Name already exists
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewCreateRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.CreateRoleRequest{
		TenantID:    tenant.GetID(),
		Name:        "existing_role",
		Description: "A role",
		Permissions: []string{"users:read"},
	}

	_, err := useCase.Execute(ctx, req, &createdBy)

	if err == nil {
		t.Fatal("Execute() should return error when name already exists")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "CONFLICT" {
		t.Errorf("Execute() error code = %s, want CONFLICT", appErr.Code)
	}
}

func TestCreateRoleUseCase_Execute_InvalidPermissions(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	createdBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		ExistsByNameFn: func(ctx context.Context, tenantID *uuid.UUID, name string) (bool, error) {
			return false, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewCreateRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.CreateRoleRequest{
		TenantID:    tenant.GetID(),
		Name:        "new_role",
		Description: "A role",
		Permissions: []string{"invalid_permission"}, // Invalid format
	}

	_, err := useCase.Execute(ctx, req, &createdBy)

	if err == nil {
		t.Fatal("Execute() should return error for invalid permissions")
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
// UpdateRoleUseCase Tests
// ============================================================================

func TestUpdateRoleUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	updatedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
		ExistsByNameFn: func(ctx context.Context, tid *uuid.UUID, name string) (bool, error) {
			return false, nil
		},
		UpdateFn: func(ctx context.Context, r *domain.Role) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateRoleRequest{
		Name:        "updated_role",
		Description: "Updated description",
		Permissions: []string{"users:read", "users:update"},
	}

	result, err := useCase.Execute(ctx, role.GetID(), tenant.GetID(), req, &updatedBy)

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

func TestUpdateRoleUseCase_Execute_RoleNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	updatedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return nil, errors.New("not found")
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateRoleRequest{
		Name: "updated_role",
	}

	_, err := useCase.Execute(ctx, uuid.New(), tenant.GetID(), req, &updatedBy)

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

func TestUpdateRoleUseCase_Execute_SystemRoleNotModifiable(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	systemRole := createSystemRole(t)
	updatedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return systemRole, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateRoleRequest{
		Name: "try_to_update",
	}

	_, err := useCase.Execute(ctx, systemRole.GetID(), tenant.GetID(), req, &updatedBy)

	if err == nil {
		t.Fatal("Execute() should return error when trying to modify system role")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("Execute() error code = %s, want FORBIDDEN", appErr.Code)
	}
}

func TestUpdateRoleUseCase_Execute_RoleDifferentTenant(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	differentTenantID := uuid.New()
	role := createTestRole(t, &differentTenantID, "test_role")
	updatedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewUpdateRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.UpdateRoleRequest{
		Name: "updated_role",
	}

	_, err := useCase.Execute(ctx, role.GetID(), tenant.GetID(), req, &updatedBy)

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

// ============================================================================
// DeleteRoleUseCase Tests
// ============================================================================

func TestDeleteRoleUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	deletedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}

	userRepo := &FullMockUserRepository{
		FindByRoleIDFn: func(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
			return []*domain.User{}, nil // No users assigned
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteRoleUseCase(roleRepo, userRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, role.GetID(), tenant.GetID(), &deletedBy)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(auditLogger.Calls) == 0 {
		t.Error("Execute() should log audit entry")
	}
}

func TestDeleteRoleUseCase_Execute_RoleNotFound(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	deletedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return nil, errors.New("not found")
		},
	}

	userRepo := &FullMockUserRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteRoleUseCase(roleRepo, userRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, uuid.New(), tenant.GetID(), &deletedBy)

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

func TestDeleteRoleUseCase_Execute_SystemRoleNotDeletable(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	systemRole := createSystemRole(t)
	deletedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return systemRole, nil
		},
	}

	userRepo := &FullMockUserRepository{}
	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteRoleUseCase(roleRepo, userRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, systemRole.GetID(), tenant.GetID(), &deletedBy)

	if err == nil {
		t.Fatal("Execute() should return error when trying to delete system role")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "FORBIDDEN" {
		t.Errorf("Execute() error code = %s, want FORBIDDEN", appErr.Code)
	}
}

func TestDeleteRoleUseCase_Execute_RoleHasUsers(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	deletedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	userRepo := &FullMockUserRepository{
		FindByRoleIDFn: func(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
			user := createTestUser(t, tenantID)
			return []*domain.User{user}, nil // Role has assigned users
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewDeleteRoleUseCase(roleRepo, userRepo, outboxRepo, txManager, auditLogger)

	err := useCase.Execute(ctx, role.GetID(), tenant.GetID(), &deletedBy)

	if err == nil {
		t.Fatal("Execute() should return error when role has users")
	}

	var appErr *application.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Execute() error should be AppError, got %T", err)
	}
	if appErr.Code != "CONFLICT" {
		t.Errorf("Execute() error code = %s, want CONFLICT", appErr.Code)
	}
}

// ============================================================================
// AddPermissionToRoleUseCase Tests
// ============================================================================

func TestAddPermissionToRoleUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	addedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
		UpdateFn: func(ctx context.Context, r *domain.Role) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAddPermissionToRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.AddPermissionRequest{
		Permission: "customers:create",
	}

	result, err := useCase.Execute(ctx, role.GetID(), tenant.GetID(), req, &addedBy)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
}

func TestAddPermissionToRoleUseCase_Execute_InvalidPermission(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	addedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewAddPermissionToRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.AddPermissionRequest{
		Permission: "invalid_permission",
	}

	_, err := useCase.Execute(ctx, role.GetID(), tenant.GetID(), req, &addedBy)

	if err == nil {
		t.Fatal("Execute() should return error for invalid permission")
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
// RemovePermissionFromRoleUseCase Tests
// ============================================================================

func TestRemovePermissionFromRoleUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenant := createTestTenant(t)
	tenantID := tenant.GetID()
	role := createTestRole(t, &tenantID, "test_role")
	removedBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
		UpdateFn: func(ctx context.Context, r *domain.Role) error {
			return nil
		},
	}

	outboxRepo := &MockOutboxRepository{}
	txManager := &MockTransactionManager{}
	auditLogger := &MockAuditLogger{}

	useCase := NewRemovePermissionFromRoleUseCase(roleRepo, outboxRepo, txManager, auditLogger)

	req := &dto.RemovePermissionRequest{
		Permission: "users:read",
	}

	result, err := useCase.Execute(ctx, role.GetID(), tenant.GetID(), req, &removedBy)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil {
		t.Fatal("Execute() result should not be nil")
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkCreateRoleUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	createdBy := uuid.New()

	roleRepo := &FullMockRoleRepository{
		ExistsByNameFn: func(ctx context.Context, tid *uuid.UUID, name string) (bool, error) {
			return false, nil
		},
		CreateFn: func(ctx context.Context, role *domain.Role) error {
			return nil
		},
	}

	useCase := NewCreateRoleUseCase(
		roleRepo,
		&MockOutboxRepository{},
		&MockTransactionManager{},
		&MockAuditLogger{},
	)

	req := &dto.CreateRoleRequest{
		TenantID:    tenantID,
		Name:        "benchmark_role",
		Description: "Benchmark role",
		Permissions: []string{"users:read", "customers:read"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, req, &createdBy)
	}
}

func BenchmarkListRolesUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	permissions := domain.NewPermissionSet()
	role1, _ := domain.NewRole(&tenantID, "role1", "Test", permissions)
	role2, _ := domain.NewRole(&tenantID, "role2", "Test", permissions)

	roleRepo := &FullMockRoleRepository{
		FindByTenantFn: func(ctx context.Context, tid uuid.UUID, opts domain.RoleQueryOptions) ([]*domain.Role, int64, error) {
			return []*domain.Role{role1, role2}, 2, nil
		},
	}

	useCase := NewListRolesUseCase(roleRepo)

	req := &dto.ListRolesRequest{
		TenantID: tenantID,
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, req)
	}
}

func BenchmarkGetRoleUseCase_Execute(b *testing.B) {
	ctx := context.Background()
	tenantID := uuid.New()
	permissions := domain.NewPermissionSet()
	role, _ := domain.NewRole(&tenantID, "test_role", "Test", permissions)

	roleRepo := &FullMockRoleRepository{
		FindByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
			return role, nil
		},
	}

	userRepo := &FullMockUserRepository{
		FindByRoleIDFn: func(ctx context.Context, roleID uuid.UUID) ([]*domain.User, error) {
			return []*domain.User{}, nil
		},
	}

	useCase := NewGetRoleUseCase(roleRepo, userRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = useCase.Execute(ctx, role.GetID(), tenantID)
	}
}
