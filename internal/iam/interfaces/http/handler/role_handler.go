// Package handler contains HTTP handlers for the IAM service.
package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application/usecase"
	iamhttp "github.com/kilang-desa-murni/crm/internal/iam/interfaces/http"
	"github.com/kilang-desa-murni/crm/internal/iam/interfaces/http/middleware"
)

// RoleHandler handles role-related HTTP requests.
type RoleHandler struct {
	createRoleUC *usecase.CreateRoleUseCase
	getRoleUC    *usecase.GetRoleUseCase
	listRolesUC  *usecase.ListRolesUseCase
	updateRoleUC *usecase.UpdateRoleUseCase
	deleteRoleUC *usecase.DeleteRoleUseCase
	decoder      *iamhttp.RequestDecoder
	getPathParam func(*http.Request, string) string
}

// NewRoleHandler creates a new RoleHandler.
func NewRoleHandler(
	createRoleUC *usecase.CreateRoleUseCase,
	getRoleUC *usecase.GetRoleUseCase,
	listRolesUC *usecase.ListRolesUseCase,
	updateRoleUC *usecase.UpdateRoleUseCase,
	deleteRoleUC *usecase.DeleteRoleUseCase,
	getPathParam func(*http.Request, string) string,
) *RoleHandler {
	return &RoleHandler{
		createRoleUC: createRoleUC,
		getRoleUC:    getRoleUC,
		listRolesUC:  listRolesUC,
		updateRoleUC: updateRoleUC,
		deleteRoleUC: deleteRoleUC,
		decoder:      iamhttp.NewRequestDecoder(),
		getPathParam: getPathParam,
	}
}

// CreateRoleRequest represents a role creation request.
type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=100"`
	Description string   `json:"description,omitempty" validate:"max=500"`
	Permissions []string `json:"permissions" validate:"required,min=1,dive,permission"`
}

// Create handles creating a new role.
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "tenant context required", nil)
		return
	}

	var req CreateRoleRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	performerID := middleware.GetUserID(r.Context())

	result, err := h.createRoleUC.Execute(r.Context(), &usecase.CreateRoleRequest{
		TenantID:    &tenantID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		CreatedBy:   &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusCreated, result.Role)
}

// List handles listing roles with pagination.
func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "tenant context required", nil)
		return
	}

	query := iamhttp.NewQueryParams(r)
	pagination := query.GetPagination()

	result, err := h.listRolesUC.Execute(r.Context(), &usecase.ListRolesRequest{
		TenantID:      tenantID,
		Page:          pagination.Page,
		PageSize:      pagination.PageSize,
		SortBy:        pagination.SortBy,
		SortDirection: pagination.SortDirection,
		Search:        query.String("search", ""),
		IncludeSystem: query.Bool("include_system", true),
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	meta := &iamhttp.MetaInfo{
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: result.Total,
		TotalPages: iamhttp.CalculateTotalPages(result.Total, pagination.PageSize),
	}

	iamhttp.WriteSuccessWithMeta(w, http.StatusOK, result.Roles, meta)
}

// Get handles getting a single role by ID.
func (h *RoleHandler) Get(w http.ResponseWriter, r *http.Request) {
	roleIDStr := h.getPathParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid role ID", nil)
		return
	}

	result, err := h.getRoleUC.Execute(r.Context(), &usecase.GetRoleRequest{
		RoleID: roleID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.Role)
}

// UpdateRoleRequest represents a role update request.
type UpdateRoleRequest struct {
	Name        string   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description string   `json:"description,omitempty" validate:"max=500"`
	Permissions []string `json:"permissions,omitempty" validate:"omitempty,dive,permission"`
}

// Update handles updating a role.
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	roleIDStr := h.getPathParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid role ID", nil)
		return
	}

	var req UpdateRoleRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	performerID := middleware.GetUserID(r.Context())

	result, err := h.updateRoleUC.Execute(r.Context(), &usecase.UpdateRoleRequest{
		RoleID:      roleID,
		Name:        &req.Name,
		Description: &req.Description,
		Permissions: req.Permissions,
		UpdatedBy:   &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.Role)
}

// Delete handles deleting a role.
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	roleIDStr := h.getPathParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid role ID", nil)
		return
	}

	performerID := middleware.GetUserID(r.Context())

	err = h.deleteRoleUC.Execute(r.Context(), &usecase.DeleteRoleRequest{
		RoleID:    roleID,
		DeletedBy: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "role deleted successfully",
	})
}

// GetSystemRoles handles getting all system roles.
func (h *RoleHandler) GetSystemRoles(w http.ResponseWriter, r *http.Request) {
	result, err := h.listRolesUC.GetSystemRoles(r.Context())
	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result)
}

// AddPermissionRequest represents a request to add permissions to a role.
type AddPermissionRequest struct {
	Permissions []string `json:"permissions" validate:"required,min=1,dive,permission"`
}

// AddPermissions handles adding permissions to a role.
func (h *RoleHandler) AddPermissions(w http.ResponseWriter, r *http.Request) {
	roleIDStr := h.getPathParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid role ID", nil)
		return
	}

	var req AddPermissionRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	performerID := middleware.GetUserID(r.Context())

	result, err := h.updateRoleUC.AddPermissions(r.Context(), &usecase.AddPermissionsRequest{
		RoleID:      roleID,
		Permissions: req.Permissions,
		UpdatedBy:   &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result)
}

// RemovePermissionRequest represents a request to remove permissions from a role.
type RemovePermissionRequest struct {
	Permissions []string `json:"permissions" validate:"required,min=1,dive,permission"`
}

// RemovePermissions handles removing permissions from a role.
func (h *RoleHandler) RemovePermissions(w http.ResponseWriter, r *http.Request) {
	roleIDStr := h.getPathParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid role ID", nil)
		return
	}

	var req RemovePermissionRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	performerID := middleware.GetUserID(r.Context())

	result, err := h.updateRoleUC.RemovePermissions(r.Context(), &usecase.RemovePermissionsRequest{
		RoleID:      roleID,
		Permissions: req.Permissions,
		UpdatedBy:   &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result)
}

// GetRoleUsers handles getting users assigned to a role.
func (h *RoleHandler) GetRoleUsers(w http.ResponseWriter, r *http.Request) {
	roleIDStr := h.getPathParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid role ID", nil)
		return
	}

	query := iamhttp.NewQueryParams(r)
	pagination := query.GetPagination()

	result, err := h.getRoleUC.GetRoleUsers(r.Context(), &usecase.GetRoleUsersRequest{
		RoleID:   roleID,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	meta := &iamhttp.MetaInfo{
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: result.Total,
		TotalPages: iamhttp.CalculateTotalPages(result.Total, pagination.PageSize),
	}

	iamhttp.WriteSuccessWithMeta(w, http.StatusOK, result.Users, meta)
}
