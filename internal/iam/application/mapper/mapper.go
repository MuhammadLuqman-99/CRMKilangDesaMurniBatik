// Package mapper provides functions to map between domain entities and DTOs.
package mapper

import (
	"time"

	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ============================================================================
// User Mappers
// ============================================================================

// UserToDTO converts a User domain entity to a UserDTO.
func UserToDTO(user *domain.User) *dto.UserDTO {
	if user == nil {
		return nil
	}

	userDTO := &dto.UserDTO{
		ID:              user.GetID(),
		TenantID:        user.TenantID(),
		Email:           user.Email().String(),
		FirstName:       user.FirstName(),
		LastName:        user.LastName(),
		FullName:        user.FullName(),
		AvatarURL:       user.AvatarURL(),
		Phone:           user.Phone(),
		Status:          user.Status().String(),
		EmailVerifiedAt: user.EmailVerifiedAt(),
		LastLoginAt:     user.LastLoginAt(),
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}

	// Map roles if available
	if roles := user.Roles(); len(roles) > 0 {
		userDTO.Roles = make([]*dto.RoleDTO, len(roles))
		for i, role := range roles {
			userDTO.Roles[i] = RoleToDTO(role)
		}
	}

	// Get all permissions from roles
	permissions := user.GetPermissions()
	if permissions != nil && permissions.Len() > 0 {
		userDTO.Permissions = permissions.Strings()
	}

	return userDTO
}

// UsersToDTO converts a slice of User domain entities to UserDTOs.
func UsersToDTO(users []*domain.User) []*dto.UserDTO {
	if users == nil {
		return nil
	}

	result := make([]*dto.UserDTO, len(users))
	for i, user := range users {
		result[i] = UserToDTO(user)
	}
	return result
}

// DeviceInfoDTOToDomain converts a DeviceInfoDTO to domain DeviceInfo.
func DeviceInfoDTOToDomain(d *dto.DeviceInfoDTO) domain.DeviceInfo {
	if d == nil {
		return domain.DeviceInfo{}
	}

	return domain.DeviceInfo{
		DeviceID:       d.DeviceID,
		DeviceType:     d.DeviceType,
		DeviceName:     d.DeviceName,
		OS:             d.OS,
		OSVersion:      d.OSVersion,
		Browser:        d.Browser,
		BrowserVersion: d.BrowserVersion,
	}
}

// ============================================================================
// Role Mappers
// ============================================================================

// RoleToDTO converts a Role domain entity to a RoleDTO.
func RoleToDTO(role *domain.Role) *dto.RoleDTO {
	if role == nil {
		return nil
	}

	roleDTO := &dto.RoleDTO{
		ID:          role.GetID(),
		TenantID:    role.TenantID(),
		Name:        role.Name(),
		Description: role.Description(),
		IsSystem:    role.IsSystem(),
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}

	// Map permissions
	if perms := role.Permissions(); perms != nil {
		roleDTO.Permissions = perms.Strings()
	} else {
		roleDTO.Permissions = []string{}
	}

	return roleDTO
}

// RolesToDTO converts a slice of Role domain entities to RoleDTOs.
func RolesToDTO(roles []*domain.Role) []*dto.RoleDTO {
	if roles == nil {
		return nil
	}

	result := make([]*dto.RoleDTO, len(roles))
	for i, role := range roles {
		result[i] = RoleToDTO(role)
	}
	return result
}

// ============================================================================
// Tenant Mappers
// ============================================================================

// TenantToDTO converts a Tenant domain entity to a TenantDTO.
func TenantToDTO(tenant *domain.Tenant) *dto.TenantDTO {
	if tenant == nil {
		return nil
	}

	settings := tenant.Settings()
	tenantDTO := &dto.TenantDTO{
		ID:     tenant.GetID(),
		Name:   tenant.Name(),
		Slug:   tenant.Slug(),
		Status: tenant.Status().String(),
		Plan:   tenant.Plan().String(),
		Settings: &dto.TenantSettingsDTO{
			Timezone:           settings.Timezone,
			DateFormat:         settings.DateFormat,
			Currency:           settings.Currency,
			Language:           settings.Language,
			NotificationsEmail: settings.NotificationsEmail,
		},
		Limits: &dto.TenantLimitsDTO{
			MaxUsers:    tenant.Plan().MaxUsers(),
			MaxContacts: tenant.Plan().MaxContacts(),
		},
		CreatedAt: tenant.CreatedAt,
		UpdatedAt: tenant.UpdatedAt,
	}

	// Add trial info if applicable
	if tenant.Status() == domain.TenantStatusTrial {
		tenantDTO.TrialInfo = &dto.TrialInfoDTO{
			IsTrialing: true,
		}

		metadata := tenant.Metadata()
		if startedAt, ok := metadata["trial_started_at"].(time.Time); ok {
			tenantDTO.TrialInfo.TrialStarted = &startedAt
		}
		if endsAt, ok := metadata["trial_ends_at"].(time.Time); ok {
			tenantDTO.TrialInfo.TrialEnds = &endsAt
			tenantDTO.TrialInfo.DaysLeft = int(time.Until(endsAt).Hours() / 24)
			if tenantDTO.TrialInfo.DaysLeft < 0 {
				tenantDTO.TrialInfo.DaysLeft = 0
			}
		}
	}

	return tenantDTO
}

// TenantsToDTO converts a slice of Tenant domain entities to TenantDTOs.
func TenantsToDTO(tenants []*domain.Tenant) []*dto.TenantDTO {
	if tenants == nil {
		return nil
	}

	result := make([]*dto.TenantDTO, len(tenants))
	for i, tenant := range tenants {
		result[i] = TenantToDTO(tenant)
	}
	return result
}

// TenantWithUsageToDTO converts a Tenant with usage stats to TenantDTO.
func TenantWithUsageToDTO(tenant *domain.Tenant, userCount, contactCount int64) *dto.TenantDTO {
	tenantDTO := TenantToDTO(tenant)
	if tenantDTO != nil {
		tenantDTO.Usage = &dto.TenantUsageDTO{
			UserCount:    userCount,
			ContactCount: contactCount,
		}
	}
	return tenantDTO
}

// ============================================================================
// Permission Mappers
// ============================================================================

// PermissionToDTO converts a Permission value object to PermissionDTO.
func PermissionToDTO(perm domain.Permission) *dto.PermissionDTO {
	return &dto.PermissionDTO{
		Resource: perm.Resource(),
		Action:   perm.Action(),
	}
}

// PermissionsToDTO converts a slice of Permission value objects to PermissionDTOs.
func PermissionsToDTO(perms []domain.Permission) []*dto.PermissionDTO {
	if perms == nil {
		return nil
	}

	result := make([]*dto.PermissionDTO, len(perms))
	for i, perm := range perms {
		result[i] = PermissionToDTO(perm)
	}
	return result
}

// StringsToPermissions converts string permissions to domain Permissions.
func StringsToPermissions(perms []string) ([]domain.Permission, error) {
	result := make([]domain.Permission, 0, len(perms))
	for _, p := range perms {
		perm, err := domain.ParsePermission(p)
		if err != nil {
			return nil, err
		}
		result = append(result, perm)
	}
	return result, nil
}

// StringsToPermissionSet converts string permissions to a PermissionSet.
func StringsToPermissionSet(perms []string) (*domain.PermissionSet, error) {
	return domain.NewPermissionSetFromStrings(perms)
}
