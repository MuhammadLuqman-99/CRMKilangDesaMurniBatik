// Package handler contains HTTP handlers for the IAM service.
package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application/usecase"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/domain"
	iamhttp "github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/interfaces/http"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/interfaces/http/middleware"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	getUserUC    *usecase.GetUserUseCase
	listUsersUC  *usecase.ListUsersUseCase
	updateUserUC *usecase.UpdateUserUseCase
	deleteUserUC *usecase.DeleteUserUseCase
	assignRoleUC *usecase.AssignRoleUseCase
	decoder      *iamhttp.RequestDecoder
	getPathParam func(*http.Request, string) string
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(
	getUserUC *usecase.GetUserUseCase,
	listUsersUC *usecase.ListUsersUseCase,
	updateUserUC *usecase.UpdateUserUseCase,
	deleteUserUC *usecase.DeleteUserUseCase,
	assignRoleUC *usecase.AssignRoleUseCase,
	getPathParam func(*http.Request, string) string,
) *UserHandler {
	return &UserHandler{
		getUserUC:    getUserUC,
		listUsersUC:  listUsersUC,
		updateUserUC: updateUserUC,
		deleteUserUC: deleteUserUC,
		assignRoleUC: assignRoleUC,
		decoder:      iamhttp.NewRequestDecoder(),
		getPathParam: getPathParam,
	}
}

// List handles listing users with pagination.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "tenant context required", nil)
		return
	}

	query := iamhttp.NewQueryParams(r)
	pagination := query.GetPagination()

	// Parse status filter
	var status *domain.UserStatus
	if statusStr := query.String("status", ""); statusStr != "" {
		s := domain.UserStatus(statusStr)
		status = &s
	}

	result, err := h.listUsersUC.Execute(r.Context(), &usecase.ListUsersRequest{
		TenantID:      tenantID,
		Page:          pagination.Page,
		PageSize:      pagination.PageSize,
		SortBy:        pagination.SortBy,
		SortDirection: pagination.SortDirection,
		Status:        status,
		Search:        query.String("search", ""),
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

// Get handles getting a single user by ID.
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getPathParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid user ID", nil)
		return
	}

	result, err := h.getUserUC.Execute(r.Context(), &usecase.GetUserRequest{
		UserID: userID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.User)
}

// UpdateUserRequest represents a user update request.
type UpdateUserRequest struct {
	FirstName string `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName  string `json:"last_name,omitempty" validate:"omitempty,max=100"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,max=20"`
}

// Update handles updating a user.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getPathParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid user ID", nil)
		return
	}

	var req UpdateUserRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	// Get performer ID from context
	performerID := middleware.GetUserID(r.Context())

	result, err := h.updateUserUC.Execute(r.Context(), &usecase.UpdateUserRequest{
		UserID:      userID,
		FirstName:   &req.FirstName,
		LastName:    &req.LastName,
		Phone:       &req.Phone,
		PerformerID: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.User)
}

// UpdateStatusRequest represents a user status update request.
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended pending_verification"`
}

// UpdateStatus handles updating a user's status.
func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getPathParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid user ID", nil)
		return
	}

	var req UpdateStatusRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	status := domain.UserStatus(req.Status)
	performerID := middleware.GetUserID(r.Context())

	result, err := h.updateUserUC.Execute(r.Context(), &usecase.UpdateUserRequest{
		UserID:      userID,
		Status:      &status,
		PerformerID: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.User)
}

// Delete handles deleting a user.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getPathParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid user ID", nil)
		return
	}

	performerID := middleware.GetUserID(r.Context())

	err = h.deleteUserUC.Execute(r.Context(), &usecase.DeleteUserRequest{
		UserID:      userID,
		PerformerID: performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "user deleted successfully",
	})
}

// AssignRoleRequest represents a role assignment request.
type AssignRoleRequest struct {
	RoleID string `json:"role_id" validate:"required,uuid"`
}

// AssignRole handles assigning a role to a user.
func (h *UserHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getPathParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid user ID", nil)
		return
	}

	var req AssignRoleRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	roleID, _ := uuid.Parse(req.RoleID)
	performerID := middleware.GetUserID(r.Context())

	err = h.assignRoleUC.Execute(r.Context(), &usecase.AssignRoleRequest{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "role assigned successfully",
	})
}

// RemoveRoleRequest represents a role removal request.
type RemoveRoleRequest struct {
	RoleID string `json:"role_id" validate:"required,uuid"`
}

// RemoveRole handles removing a role from a user.
func (h *UserHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getPathParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid user ID", nil)
		return
	}

	var req RemoveRoleRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	roleID, _ := uuid.Parse(req.RoleID)
	performerID := middleware.GetUserID(r.Context())

	err = h.assignRoleUC.RemoveRole(r.Context(), &usecase.RemoveRoleRequest{
		UserID:    userID,
		RoleID:    roleID,
		RemovedBy: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "role removed successfully",
	})
}

// GetRoles handles getting roles for a user.
func (h *UserHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getPathParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid user ID", nil)
		return
	}

	result, err := h.getUserUC.GetUserRoles(r.Context(), userID)
	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result)
}

// GetPermissions handles getting permissions for a user.
func (h *UserHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	userIDStr := h.getPathParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid user ID", nil)
		return
	}

	result, err := h.getUserUC.GetUserPermissions(r.Context(), userID)
	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result)
}
