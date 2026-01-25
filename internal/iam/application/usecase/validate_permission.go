// Package usecase contains the application use cases for the IAM service.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ValidatePermissionRequest represents a permission validation request.
type ValidatePermissionRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
	Permissions []string  `json:"permissions" validate:"required,dive,required"`
	RequireAll  bool      `json:"require_all"` // If true, all permissions must be present
}

// ValidatePermissionResponse represents a permission validation response.
type ValidatePermissionResponse struct {
	Allowed            bool              `json:"allowed"`
	MissingPermissions []string          `json:"missing_permissions,omitempty"`
	UserPermissions    []string          `json:"user_permissions,omitempty"`
}

// ValidatePermissionUseCase handles permission validation.
type ValidatePermissionUseCase struct {
	userRepo   domain.UserRepository
	roleRepo   domain.RoleRepository
	tenantRepo domain.TenantRepository
	cache      ports.CacheService
}

// NewValidatePermissionUseCase creates a new ValidatePermissionUseCase.
func NewValidatePermissionUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	tenantRepo domain.TenantRepository,
	cache ports.CacheService,
) *ValidatePermissionUseCase {
	return &ValidatePermissionUseCase{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		tenantRepo: tenantRepo,
		cache:      cache,
	}
}

// Execute validates if a user has the required permissions.
func (uc *ValidatePermissionUseCase) Execute(ctx context.Context, req *ValidatePermissionRequest) (*ValidatePermissionResponse, error) {
	// Get user
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, application.ErrNotFound("user", req.UserID)
	}

	// Verify user belongs to the tenant
	if user.TenantID() != req.TenantID {
		return nil, application.ErrForbidden("user does not belong to this tenant")
	}

	// Check user status
	if !user.IsActive() {
		return &ValidatePermissionResponse{
			Allowed:            false,
			MissingPermissions: req.Permissions,
		}, nil
	}

	// Load user roles
	roles, err := uc.roleRepo.FindByUserID(ctx, user.GetID())
	if err != nil {
		return nil, application.ErrInternal("failed to load user roles", err)
	}
	user.SetRoles(roles)

	// Get user's permission set
	userPermissions := user.GetPermissions()

	// Parse required permissions
	requiredPerms := make([]domain.Permission, 0, len(req.Permissions))
	for _, p := range req.Permissions {
		perm, err := domain.ParsePermission(p)
		if err != nil {
			return nil, application.ErrValidation("invalid permission format", map[string]interface{}{
				"permission": p,
				"error":      err.Error(),
			})
		}
		requiredPerms = append(requiredPerms, perm)
	}

	// Check permissions
	var allowed bool
	var missingPerms []string

	if req.RequireAll {
		// All permissions must be present
		allowed = userPermissions.HasAllPermissions(requiredPerms...)
		if !allowed {
			for _, perm := range requiredPerms {
				if !userPermissions.HasPermission(perm) {
					missingPerms = append(missingPerms, perm.String())
				}
			}
		}
	} else {
		// Any permission is sufficient
		allowed = userPermissions.HasAnyPermission(requiredPerms...)
		if !allowed {
			missingPerms = req.Permissions
		}
	}

	return &ValidatePermissionResponse{
		Allowed:            allowed,
		MissingPermissions: missingPerms,
		UserPermissions:    userPermissions.Strings(),
	}, nil
}

// CheckPermission is a quick check for a single permission.
func (uc *ValidatePermissionUseCase) CheckPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error) {
	resp, err := uc.Execute(ctx, &ValidatePermissionRequest{
		UserID:      userID,
		TenantID:    tenantID,
		Permissions: []string{permission},
		RequireAll:  true,
	})
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

// CheckPermissions checks multiple permissions with require all behavior.
func (uc *ValidatePermissionUseCase) CheckPermissions(ctx context.Context, userID, tenantID uuid.UUID, permissions []string, requireAll bool) (bool, []string, error) {
	resp, err := uc.Execute(ctx, &ValidatePermissionRequest{
		UserID:      userID,
		TenantID:    tenantID,
		Permissions: permissions,
		RequireAll:  requireAll,
	})
	if err != nil {
		return false, nil, err
	}
	return resp.Allowed, resp.MissingPermissions, nil
}

// AuthorizeRequest represents an authorization request for resource access.
type AuthorizeRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Resource   string    `json:"resource"`
	Action     string    `json:"action"`
	ResourceID *uuid.UUID `json:"resource_id,omitempty"` // For resource-level checks
}

// AuthorizeResponse represents an authorization response.
type AuthorizeResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// AuthorizeUseCase handles resource access authorization.
type AuthorizeUseCase struct {
	userRepo   domain.UserRepository
	roleRepo   domain.RoleRepository
	tenantRepo domain.TenantRepository
}

// NewAuthorizeUseCase creates a new AuthorizeUseCase.
func NewAuthorizeUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	tenantRepo domain.TenantRepository,
) *AuthorizeUseCase {
	return &AuthorizeUseCase{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		tenantRepo: tenantRepo,
	}
}

// Execute authorizes a user for resource access.
func (uc *AuthorizeUseCase) Execute(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	// Get user
	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return &AuthorizeResponse{Allowed: false, Reason: "user not found"}, nil
	}

	// Verify tenant
	if user.TenantID() != req.TenantID {
		return &AuthorizeResponse{Allowed: false, Reason: "user does not belong to tenant"}, nil
	}

	// Check user status
	if !user.IsActive() {
		return &AuthorizeResponse{Allowed: false, Reason: "user is not active"}, nil
	}

	// Verify tenant is active
	tenant, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return &AuthorizeResponse{Allowed: false, Reason: "tenant not found"}, nil
	}

	if !tenant.IsActive() {
		return &AuthorizeResponse{Allowed: false, Reason: "tenant is not active"}, nil
	}

	// Load user roles
	roles, err := uc.roleRepo.FindByUserID(ctx, user.GetID())
	if err != nil {
		return nil, application.ErrInternal("failed to load user roles", err)
	}
	user.SetRoles(roles)

	// Create permission to check
	permission, err := domain.NewPermission(req.Resource, req.Action)
	if err != nil {
		return &AuthorizeResponse{Allowed: false, Reason: "invalid permission format"}, nil
	}

	// Check permission
	if user.HasPermission(permission) {
		return &AuthorizeResponse{Allowed: true}, nil
	}

	return &AuthorizeResponse{
		Allowed: false,
		Reason:  "insufficient permissions",
	}, nil
}

// Authorize is a shorthand for authorization check.
func (uc *AuthorizeUseCase) Authorize(ctx context.Context, userID, tenantID uuid.UUID, resource, action string) (bool, error) {
	resp, err := uc.Execute(ctx, &AuthorizeRequest{
		UserID:   userID,
		TenantID: tenantID,
		Resource: resource,
		Action:   action,
	})
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

// ValidateTokenClaimsUseCase validates token claims are still valid.
type ValidateTokenClaimsUseCase struct {
	userRepo   domain.UserRepository
	tenantRepo domain.TenantRepository
}

// NewValidateTokenClaimsUseCase creates a new ValidateTokenClaimsUseCase.
func NewValidateTokenClaimsUseCase(
	userRepo domain.UserRepository,
	tenantRepo domain.TenantRepository,
) *ValidateTokenClaimsUseCase {
	return &ValidateTokenClaimsUseCase{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// Execute validates that token claims are still valid (user active, tenant active).
func (uc *ValidateTokenClaimsUseCase) Execute(ctx context.Context, claims *ports.TokenClaims) error {
	// Validate user
	user, err := uc.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return application.ErrUnauthorized("user not found")
	}

	if !user.CanLogin() {
		return application.ErrUserInactive()
	}

	// Validate tenant
	tenant, err := uc.tenantRepo.FindByID(ctx, claims.TenantID)
	if err != nil {
		return application.ErrUnauthorized("tenant not found")
	}

	if !tenant.IsActive() {
		return application.ErrTenantInactive()
	}

	return nil
}
