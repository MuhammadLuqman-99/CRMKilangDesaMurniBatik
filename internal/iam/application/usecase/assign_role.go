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

// AssignRoleUseCase handles role assignment to users.
type AssignRoleUseCase struct {
	userRepo    domain.UserRepository
	roleRepo    domain.RoleRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewAssignRoleUseCase creates a new AssignRoleUseCase.
func NewAssignRoleUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *AssignRoleUseCase {
	return &AssignRoleUseCase{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute assigns a role to a user.
func (uc *AssignRoleUseCase) Execute(ctx context.Context, req *dto.AssignRoleRequest, assignedBy *uuid.UUID, tenantID uuid.UUID) (*dto.UserDTO, error) {
	// Get user
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, application.ErrNotFound("user", req.UserID)
	}

	// Verify user belongs to the same tenant
	if user.TenantID() != tenantID {
		return nil, application.ErrForbidden("user does not belong to this tenant")
	}

	// Get role
	role, err := uc.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil {
		return nil, application.ErrNotFound("role", req.RoleID)
	}

	// Verify role is accessible (system role or belongs to tenant)
	if !role.IsSystem() && role.TenantID() != nil && *role.TenantID() != tenantID {
		return nil, application.ErrForbidden("role does not belong to this tenant")
	}

	// Check if role is already assigned
	if user.HasRole(req.RoleID) {
		return nil, application.ErrConflict("role is already assigned to user")
	}

	// Load current roles
	currentRoles, err := uc.roleRepo.FindByUserID(ctx, user.GetID())
	if err == nil {
		user.SetRoles(currentRoles)
	}

	// Assign role to user
	if err := user.AssignRole(role); err != nil {
		return nil, application.ErrConflict(err.Error())
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Assign role in database
		if err := uc.roleRepo.AssignRoleToUser(txCtx, req.UserID, req.RoleID, assignedBy); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return nil, application.ErrInternal("failed to assign role", err)
	}

	user.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     assignedBy,
		Action:     ports.AuditActionRoleAssigned,
		EntityType: "user",
		EntityID:   ptrToUUID(user.GetID()),
		NewValues: map[string]interface{}{
			"role_id":   role.GetID(),
			"role_name": role.Name(),
		},
	})

	// Reload user with updated roles
	updatedUser, err := uc.userRepo.FindByID(ctx, user.GetID())
	if err != nil {
		return mapper.UserToDTO(user), nil
	}

	roles, _ := uc.roleRepo.FindByUserID(ctx, updatedUser.GetID())
	updatedUser.SetRoles(roles)

	return mapper.UserToDTO(updatedUser), nil
}

// RemoveRoleUseCase handles role removal from users.
type RemoveRoleUseCase struct {
	userRepo    domain.UserRepository
	roleRepo    domain.RoleRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewRemoveRoleUseCase creates a new RemoveRoleUseCase.
func NewRemoveRoleUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *RemoveRoleUseCase {
	return &RemoveRoleUseCase{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute removes a role from a user.
func (uc *RemoveRoleUseCase) Execute(ctx context.Context, req *dto.RemoveRoleRequest, removedBy *uuid.UUID, tenantID uuid.UUID) (*dto.UserDTO, error) {
	// Get user
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, application.ErrNotFound("user", req.UserID)
	}

	// Verify user belongs to the same tenant
	if user.TenantID() != tenantID {
		return nil, application.ErrForbidden("user does not belong to this tenant")
	}

	// Get role
	role, err := uc.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil {
		return nil, application.ErrNotFound("role", req.RoleID)
	}

	// Check if role is assigned
	if !user.HasRole(req.RoleID) {
		return nil, application.ErrNotFound("role assignment", req.RoleID)
	}

	// Load current roles
	currentRoles, err := uc.roleRepo.FindByUserID(ctx, user.GetID())
	if err == nil {
		user.SetRoles(currentRoles)
	}

	// Remove role from user
	if err := user.RemoveRole(req.RoleID); err != nil {
		return nil, application.ErrNotFound("role assignment", req.RoleID)
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Remove role assignment in database
		if err := uc.roleRepo.RemoveRoleFromUser(txCtx, req.UserID, req.RoleID); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return nil, application.ErrInternal("failed to remove role", err)
	}

	user.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     removedBy,
		Action:     ports.AuditActionRoleRemoved,
		EntityType: "user",
		EntityID:   ptrToUUID(user.GetID()),
		OldValues: map[string]interface{}{
			"role_id":   role.GetID(),
			"role_name": role.Name(),
		},
	})

	// Reload user with updated roles
	updatedUser, err := uc.userRepo.FindByID(ctx, user.GetID())
	if err != nil {
		return mapper.UserToDTO(user), nil
	}

	roles, _ := uc.roleRepo.FindByUserID(ctx, updatedUser.GetID())
	updatedUser.SetRoles(roles)

	return mapper.UserToDTO(updatedUser), nil
}

// GetUserRolesUseCase retrieves all roles assigned to a user.
type GetUserRolesUseCase struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
}

// NewGetUserRolesUseCase creates a new GetUserRolesUseCase.
func NewGetUserRolesUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
) *GetUserRolesUseCase {
	return &GetUserRolesUseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute retrieves all roles for a user.
func (uc *GetUserRolesUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID) ([]*dto.RoleDTO, error) {
	// Get user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, application.ErrNotFound("user", userID)
	}

	// Verify user belongs to the same tenant
	if user.TenantID() != tenantID {
		return nil, application.ErrForbidden("user does not belong to this tenant")
	}

	// Get roles
	roles, err := uc.roleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, application.ErrInternal("failed to get user roles", err)
	}

	return mapper.RolesToDTO(roles), nil
}

// GetUserPermissionsUseCase retrieves all permissions for a user.
type GetUserPermissionsUseCase struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
}

// NewGetUserPermissionsUseCase creates a new GetUserPermissionsUseCase.
func NewGetUserPermissionsUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
) *GetUserPermissionsUseCase {
	return &GetUserPermissionsUseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute retrieves all permissions for a user.
func (uc *GetUserPermissionsUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	// Get user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, application.ErrNotFound("user", userID)
	}

	// Verify user belongs to the same tenant
	if user.TenantID() != tenantID {
		return nil, application.ErrForbidden("user does not belong to this tenant")
	}

	// Get roles
	roles, err := uc.roleRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, application.ErrInternal("failed to get user roles", err)
	}

	user.SetRoles(roles)

	return user.GetPermissions().Strings(), nil
}
