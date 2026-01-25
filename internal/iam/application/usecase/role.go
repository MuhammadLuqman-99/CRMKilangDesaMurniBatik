// Package usecase contains the application use cases for the IAM service.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/application/mapper"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ============================================================================
// Role Query Use Cases
// ============================================================================

// GetRoleUseCase handles retrieving a role by ID.
type GetRoleUseCase struct {
	roleRepo domain.RoleRepository
	userRepo domain.UserRepository
}

// NewGetRoleUseCase creates a new GetRoleUseCase.
func NewGetRoleUseCase(
	roleRepo domain.RoleRepository,
	userRepo domain.UserRepository,
) *GetRoleUseCase {
	return &GetRoleUseCase{
		roleRepo: roleRepo,
		userRepo: userRepo,
	}
}

// Execute retrieves a role by ID.
func (uc *GetRoleUseCase) Execute(ctx context.Context, roleID, tenantID uuid.UUID) (*dto.RoleDTO, error) {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, application.ErrNotFound("role", roleID)
	}

	// Verify role is accessible (system role or belongs to tenant)
	if !role.IsSystem() && role.TenantID() != nil && *role.TenantID() != tenantID {
		return nil, application.ErrForbidden("role does not belong to this tenant")
	}

	roleDTO := mapper.RoleToDTO(role)

	// Get user count for this role
	users, err := uc.userRepo.FindByRoleID(ctx, roleID)
	if err == nil {
		roleDTO.UserCount = int64(len(users))
	}

	return roleDTO, nil
}

// ListRolesUseCase handles listing roles with pagination.
type ListRolesUseCase struct {
	roleRepo domain.RoleRepository
}

// NewListRolesUseCase creates a new ListRolesUseCase.
func NewListRolesUseCase(roleRepo domain.RoleRepository) *ListRolesUseCase {
	return &ListRolesUseCase{
		roleRepo: roleRepo,
	}
}

// Execute lists roles with pagination.
func (uc *ListRolesUseCase) Execute(ctx context.Context, req *dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "name"
	}
	if req.SortDirection == "" {
		req.SortDirection = "asc"
	}

	opts := domain.RoleQueryOptions{
		Page:          req.Page,
		PageSize:      req.PageSize,
		SortBy:        req.SortBy,
		SortDirection: req.SortDirection,
		IncludeSystem: req.IncludeSystem,
		Search:        req.Search,
	}

	roles, total, err := uc.roleRepo.FindByTenant(ctx, req.TenantID, opts)
	if err != nil {
		return nil, application.ErrInternal("failed to list roles", err)
	}

	return &dto.ListRolesResponse{
		Roles:      mapper.RolesToDTO(roles),
		Pagination: dto.NewPaginationDTO(req.Page, req.PageSize, total),
	}, nil
}

// ListSystemRolesUseCase handles listing system roles.
type ListSystemRolesUseCase struct {
	roleRepo domain.RoleRepository
}

// NewListSystemRolesUseCase creates a new ListSystemRolesUseCase.
func NewListSystemRolesUseCase(roleRepo domain.RoleRepository) *ListSystemRolesUseCase {
	return &ListSystemRolesUseCase{
		roleRepo: roleRepo,
	}
}

// Execute lists all system roles.
func (uc *ListSystemRolesUseCase) Execute(ctx context.Context) ([]*dto.RoleDTO, error) {
	roles, err := uc.roleRepo.FindSystemRoles(ctx)
	if err != nil {
		return nil, application.ErrInternal("failed to list system roles", err)
	}

	return mapper.RolesToDTO(roles), nil
}

// ListAvailablePermissionsUseCase handles listing available permissions.
type ListAvailablePermissionsUseCase struct{}

// NewListAvailablePermissionsUseCase creates a new ListAvailablePermissionsUseCase.
func NewListAvailablePermissionsUseCase() *ListAvailablePermissionsUseCase {
	return &ListAvailablePermissionsUseCase{}
}

// Execute lists all available permissions grouped by resource.
func (uc *ListAvailablePermissionsUseCase) Execute(ctx context.Context) *dto.ListAvailablePermissionsResponse {
	groups := []*dto.PermissionGroupDTO{
		{
			Resource:    "users",
			Description: "User management permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "users", Action: "create", Description: "Create new users"},
				{Resource: "users", Action: "read", Description: "View user details"},
				{Resource: "users", Action: "update", Description: "Update user profiles"},
				{Resource: "users", Action: "delete", Description: "Delete users"},
				{Resource: "users", Action: "list", Description: "List all users"},
				{Resource: "users", Action: "*", Description: "Full access to user management"},
			},
		},
		{
			Resource:    "roles",
			Description: "Role management permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "roles", Action: "create", Description: "Create new roles"},
				{Resource: "roles", Action: "read", Description: "View role details"},
				{Resource: "roles", Action: "update", Description: "Update role permissions"},
				{Resource: "roles", Action: "delete", Description: "Delete roles"},
				{Resource: "roles", Action: "list", Description: "List all roles"},
				{Resource: "roles", Action: "*", Description: "Full access to role management"},
			},
		},
		{
			Resource:    "customers",
			Description: "Customer management permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "customers", Action: "create", Description: "Create new customers"},
				{Resource: "customers", Action: "read", Description: "View customer details"},
				{Resource: "customers", Action: "update", Description: "Update customer information"},
				{Resource: "customers", Action: "delete", Description: "Delete customers"},
				{Resource: "customers", Action: "list", Description: "List all customers"},
				{Resource: "customers", Action: "*", Description: "Full access to customer management"},
			},
		},
		{
			Resource:    "contacts",
			Description: "Contact management permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "contacts", Action: "create", Description: "Create new contacts"},
				{Resource: "contacts", Action: "read", Description: "View contact details"},
				{Resource: "contacts", Action: "update", Description: "Update contact information"},
				{Resource: "contacts", Action: "delete", Description: "Delete contacts"},
				{Resource: "contacts", Action: "list", Description: "List all contacts"},
				{Resource: "contacts", Action: "*", Description: "Full access to contact management"},
			},
		},
		{
			Resource:    "leads",
			Description: "Lead management permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "leads", Action: "create", Description: "Create new leads"},
				{Resource: "leads", Action: "read", Description: "View lead details"},
				{Resource: "leads", Action: "update", Description: "Update lead information"},
				{Resource: "leads", Action: "delete", Description: "Delete leads"},
				{Resource: "leads", Action: "convert", Description: "Convert leads to opportunities"},
				{Resource: "leads", Action: "*", Description: "Full access to lead management"},
			},
		},
		{
			Resource:    "opportunities",
			Description: "Opportunity management permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "opportunities", Action: "create", Description: "Create new opportunities"},
				{Resource: "opportunities", Action: "read", Description: "View opportunity details"},
				{Resource: "opportunities", Action: "update", Description: "Update opportunity information"},
				{Resource: "opportunities", Action: "delete", Description: "Delete opportunities"},
				{Resource: "opportunities", Action: "list", Description: "List all opportunities"},
				{Resource: "opportunities", Action: "*", Description: "Full access to opportunity management"},
			},
		},
		{
			Resource:    "deals",
			Description: "Deal management permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "deals", Action: "create", Description: "Create new deals"},
				{Resource: "deals", Action: "read", Description: "View deal details"},
				{Resource: "deals", Action: "update", Description: "Update deal information"},
				{Resource: "deals", Action: "delete", Description: "Delete deals"},
				{Resource: "deals", Action: "*", Description: "Full access to deal management"},
			},
		},
		{
			Resource:    "settings",
			Description: "Settings management permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "settings", Action: "read", Description: "View settings"},
				{Resource: "settings", Action: "update", Description: "Update settings"},
				{Resource: "settings", Action: "*", Description: "Full access to settings"},
			},
		},
		{
			Resource:    "reports",
			Description: "Reporting permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "reports", Action: "read", Description: "View reports"},
				{Resource: "reports", Action: "create", Description: "Create reports"},
				{Resource: "reports", Action: "export", Description: "Export reports"},
				{Resource: "reports", Action: "*", Description: "Full access to reports"},
			},
		},
		{
			Resource:    "*",
			Description: "Wildcard permissions",
			Permissions: []*dto.PermissionDTO{
				{Resource: "*", Action: "read", Description: "Read access to all resources"},
				{Resource: "*", Action: "*", Description: "Full access to everything (super admin)"},
			},
		},
	}

	return &dto.ListAvailablePermissionsResponse{
		Groups: groups,
	}
}

// ============================================================================
// Role Mutation Use Cases
// ============================================================================

// CreateRoleUseCase handles creating a new role.
type CreateRoleUseCase struct {
	roleRepo    domain.RoleRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewCreateRoleUseCase creates a new CreateRoleUseCase.
func NewCreateRoleUseCase(
	roleRepo domain.RoleRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *CreateRoleUseCase {
	return &CreateRoleUseCase{
		roleRepo:    roleRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute creates a new role.
func (uc *CreateRoleUseCase) Execute(ctx context.Context, req *dto.CreateRoleRequest, createdBy *uuid.UUID) (*dto.CreateRoleResponse, error) {
	// Check if role name already exists in tenant
	exists, err := uc.roleRepo.ExistsByName(ctx, &req.TenantID, req.Name)
	if err != nil {
		return nil, application.ErrInternal("failed to check role name", err)
	}

	if exists {
		return nil, application.ErrConflict("role with this name already exists")
	}

	// Parse permissions
	permissionSet, err := mapper.StringsToPermissionSet(req.Permissions)
	if err != nil {
		return nil, application.ErrValidation("invalid permissions", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create role
	role, err := domain.NewRole(&req.TenantID, req.Name, req.Description, permissionSet)
	if err != nil {
		return nil, application.ErrValidation("invalid role data", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.roleRepo.Create(txCtx, role); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range role.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to create role", err)
	}

	role.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   req.TenantID,
		UserID:     createdBy,
		Action:     ports.AuditActionCreate,
		EntityType: "role",
		EntityID:   ptrToUUID(role.GetID()),
		NewValues: map[string]interface{}{
			"name":        role.Name(),
			"permissions": req.Permissions,
		},
	})

	return &dto.CreateRoleResponse{
		Role: mapper.RoleToDTO(role),
	}, nil
}

// UpdateRoleUseCase handles updating a role.
type UpdateRoleUseCase struct {
	roleRepo    domain.RoleRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewUpdateRoleUseCase creates a new UpdateRoleUseCase.
func NewUpdateRoleUseCase(
	roleRepo domain.RoleRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *UpdateRoleUseCase {
	return &UpdateRoleUseCase{
		roleRepo:    roleRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute updates a role.
func (uc *UpdateRoleUseCase) Execute(ctx context.Context, roleID, tenantID uuid.UUID, req *dto.UpdateRoleRequest, updatedBy *uuid.UUID) (*dto.UpdateRoleResponse, error) {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, application.ErrNotFound("role", roleID)
	}

	// Verify role belongs to tenant
	if role.TenantID() != nil && *role.TenantID() != tenantID {
		return nil, application.ErrForbidden("role does not belong to this tenant")
	}

	// Cannot modify system roles
	if role.IsSystem() {
		return nil, application.ErrForbidden("cannot modify system role")
	}

	oldName := role.Name()
	oldPermissions := role.Permissions().Strings()

	// Update details
	if req.Name != "" {
		// Check if new name conflicts with existing role
		if req.Name != role.Name() {
			exists, err := uc.roleRepo.ExistsByName(ctx, &tenantID, req.Name)
			if err != nil {
				return nil, application.ErrInternal("failed to check role name", err)
			}
			if exists {
				return nil, application.ErrConflict("role with this name already exists")
			}
		}

		if err := role.UpdateDetails(req.Name, req.Description); err != nil {
			return nil, application.ErrValidation("invalid role data", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Update permissions if provided
	if len(req.Permissions) > 0 {
		permissionSet, err := mapper.StringsToPermissionSet(req.Permissions)
		if err != nil {
			return nil, application.ErrValidation("invalid permissions", map[string]interface{}{
				"error": err.Error(),
			})
		}

		if err := role.SetPermissions(permissionSet); err != nil {
			return nil, application.ErrForbidden(err.Error())
		}
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.roleRepo.Update(txCtx, role); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range role.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to update role", err)
	}

	role.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     updatedBy,
		Action:     ports.AuditActionUpdate,
		EntityType: "role",
		EntityID:   ptrToUUID(roleID),
		OldValues: map[string]interface{}{
			"name":        oldName,
			"permissions": oldPermissions,
		},
		NewValues: map[string]interface{}{
			"name":        role.Name(),
			"permissions": role.Permissions().Strings(),
		},
	})

	return &dto.UpdateRoleResponse{
		Role: mapper.RoleToDTO(role),
	}, nil
}

// DeleteRoleUseCase handles deleting a role.
type DeleteRoleUseCase struct {
	roleRepo    domain.RoleRepository
	userRepo    domain.UserRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewDeleteRoleUseCase creates a new DeleteRoleUseCase.
func NewDeleteRoleUseCase(
	roleRepo domain.RoleRepository,
	userRepo domain.UserRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *DeleteRoleUseCase {
	return &DeleteRoleUseCase{
		roleRepo:    roleRepo,
		userRepo:    userRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute deletes a role.
func (uc *DeleteRoleUseCase) Execute(ctx context.Context, roleID, tenantID uuid.UUID, deletedBy *uuid.UUID) error {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return application.ErrNotFound("role", roleID)
	}

	// Verify role belongs to tenant
	if role.TenantID() != nil && *role.TenantID() != tenantID {
		return application.ErrForbidden("role does not belong to this tenant")
	}

	// Cannot delete system roles
	if role.IsSystem() {
		return application.ErrForbidden("cannot delete system role")
	}

	// Check if role is assigned to any users
	users, err := uc.userRepo.FindByRoleID(ctx, roleID)
	if err != nil {
		return application.ErrInternal("failed to check role assignments", err)
	}

	if len(users) > 0 {
		return application.ErrConflict("cannot delete role that is assigned to users").WithDetail("user_count", len(users))
	}

	// Delete role
	if err := role.Delete(); err != nil {
		return application.ErrForbidden(err.Error())
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.roleRepo.Delete(txCtx, roleID); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range role.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return application.ErrInternal("failed to delete role", err)
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     deletedBy,
		Action:     ports.AuditActionDelete,
		EntityType: "role",
		EntityID:   ptrToUUID(roleID),
		OldValues: map[string]interface{}{
			"name": role.Name(),
		},
	})

	return nil
}

// AddPermissionToRoleUseCase handles adding a permission to a role.
type AddPermissionToRoleUseCase struct {
	roleRepo    domain.RoleRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewAddPermissionToRoleUseCase creates a new AddPermissionToRoleUseCase.
func NewAddPermissionToRoleUseCase(
	roleRepo domain.RoleRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *AddPermissionToRoleUseCase {
	return &AddPermissionToRoleUseCase{
		roleRepo:    roleRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute adds a permission to a role.
func (uc *AddPermissionToRoleUseCase) Execute(ctx context.Context, roleID, tenantID uuid.UUID, req *dto.AddPermissionRequest, addedBy *uuid.UUID) (*dto.RoleDTO, error) {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, application.ErrNotFound("role", roleID)
	}

	// Verify role belongs to tenant
	if role.TenantID() != nil && *role.TenantID() != tenantID {
		return nil, application.ErrForbidden("role does not belong to this tenant")
	}

	// Parse permission
	permission, err := domain.ParsePermission(req.Permission)
	if err != nil {
		return nil, application.ErrValidation("invalid permission format", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Add permission
	if err := role.AddPermission(permission); err != nil {
		return nil, application.ErrForbidden(err.Error())
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.roleRepo.Update(txCtx, role); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range role.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to add permission", err)
	}

	role.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     addedBy,
		Action:     "permission_added",
		EntityType: "role",
		EntityID:   ptrToUUID(roleID),
		NewValues: map[string]interface{}{
			"permission": req.Permission,
		},
	})

	return mapper.RoleToDTO(role), nil
}

// RemovePermissionFromRoleUseCase handles removing a permission from a role.
type RemovePermissionFromRoleUseCase struct {
	roleRepo    domain.RoleRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewRemovePermissionFromRoleUseCase creates a new RemovePermissionFromRoleUseCase.
func NewRemovePermissionFromRoleUseCase(
	roleRepo domain.RoleRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *RemovePermissionFromRoleUseCase {
	return &RemovePermissionFromRoleUseCase{
		roleRepo:    roleRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute removes a permission from a role.
func (uc *RemovePermissionFromRoleUseCase) Execute(ctx context.Context, roleID, tenantID uuid.UUID, req *dto.RemovePermissionRequest, removedBy *uuid.UUID) (*dto.RoleDTO, error) {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, application.ErrNotFound("role", roleID)
	}

	// Verify role belongs to tenant
	if role.TenantID() != nil && *role.TenantID() != tenantID {
		return nil, application.ErrForbidden("role does not belong to this tenant")
	}

	// Parse permission
	permission, err := domain.ParsePermission(req.Permission)
	if err != nil {
		return nil, application.ErrValidation("invalid permission format", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Remove permission
	if err := role.RemovePermission(permission); err != nil {
		return nil, application.ErrForbidden(err.Error())
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.roleRepo.Update(txCtx, role); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range role.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to remove permission", err)
	}

	role.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     removedBy,
		Action:     "permission_removed",
		EntityType: "role",
		EntityID:   ptrToUUID(roleID),
		OldValues: map[string]interface{}{
			"permission": req.Permission,
		},
	})

	return mapper.RoleToDTO(role), nil
}
