// Package handler contains HTTP handlers for the IAM service.
package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application/usecase"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
	iamhttp "github.com/kilang-desa-murni/crm/internal/iam/interfaces/http"
	"github.com/kilang-desa-murni/crm/internal/iam/interfaces/http/middleware"
)

// TenantHandler handles tenant-related HTTP requests.
type TenantHandler struct {
	createTenantUC *usecase.CreateTenantUseCase
	getTenantUC    *usecase.GetTenantUseCase
	listTenantsUC  *usecase.ListTenantsUseCase
	updateTenantUC *usecase.UpdateTenantUseCase
	decoder        *iamhttp.RequestDecoder
	getPathParam   func(*http.Request, string) string
}

// NewTenantHandler creates a new TenantHandler.
func NewTenantHandler(
	createTenantUC *usecase.CreateTenantUseCase,
	getTenantUC *usecase.GetTenantUseCase,
	listTenantsUC *usecase.ListTenantsUseCase,
	updateTenantUC *usecase.UpdateTenantUseCase,
	getPathParam func(*http.Request, string) string,
) *TenantHandler {
	return &TenantHandler{
		createTenantUC: createTenantUC,
		getTenantUC:    getTenantUC,
		listTenantsUC:  listTenantsUC,
		updateTenantUC: updateTenantUC,
		decoder:        iamhttp.NewRequestDecoder(),
		getPathParam:   getPathParam,
	}
}

// CreateTenantRequest represents a tenant creation request.
type CreateTenantRequest struct {
	Name     string                 `json:"name" validate:"required,min=2,max=100"`
	Slug     string                 `json:"slug" validate:"required,slug,min=2,max=50"`
	Plan     string                 `json:"plan,omitempty" validate:"omitempty,oneof=free starter professional enterprise"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Create handles creating a new tenant.
func (h *TenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	performerID := middleware.GetUserID(r.Context())

	result, err := h.createTenantUC.Execute(r.Context(), &usecase.CreateTenantRequest{
		Name:      req.Name,
		Slug:      req.Slug,
		Plan:      req.Plan,
		Settings:  req.Settings,
		Metadata:  req.Metadata,
		CreatedBy: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusCreated, result.Tenant)
}

// List handles listing tenants with pagination.
func (h *TenantHandler) List(w http.ResponseWriter, r *http.Request) {
	query := iamhttp.NewQueryParams(r)
	pagination := query.GetPagination()

	// Parse status filter
	var status *domain.TenantStatus
	if statusStr := query.String("status", ""); statusStr != "" {
		s := domain.TenantStatus(statusStr)
		status = &s
	}

	// Parse plan filter
	var plan *domain.TenantPlan
	if planStr := query.String("plan", ""); planStr != "" {
		p := domain.TenantPlan(planStr)
		plan = &p
	}

	result, err := h.listTenantsUC.Execute(r.Context(), &usecase.ListTenantsRequest{
		Page:          pagination.Page,
		PageSize:      pagination.PageSize,
		SortBy:        pagination.SortBy,
		SortDirection: pagination.SortDirection,
		Status:        status,
		Plan:          plan,
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

	iamhttp.WriteSuccessWithMeta(w, http.StatusOK, result.Tenants, meta)
}

// Get handles getting a single tenant by ID.
func (h *TenantHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := h.getPathParam(r, "id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid tenant ID", nil)
		return
	}

	result, err := h.getTenantUC.Execute(r.Context(), &usecase.GetTenantRequest{
		TenantID: tenantID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.Tenant)
}

// GetBySlug handles getting a tenant by slug.
func (h *TenantHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := h.getPathParam(r, "slug")
	if slug == "" {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "slug is required", nil)
		return
	}

	result, err := h.getTenantUC.GetBySlug(r.Context(), slug)

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result)
}

// UpdateTenantRequest represents a tenant update request.
type UpdateTenantRequest struct {
	Name     string                 `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Update handles updating a tenant.
func (h *TenantHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := h.getPathParam(r, "id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid tenant ID", nil)
		return
	}

	var req UpdateTenantRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	performerID := middleware.GetUserID(r.Context())

	result, err := h.updateTenantUC.Execute(r.Context(), &usecase.UpdateTenantRequest{
		TenantID:  tenantID,
		Name:      &req.Name,
		Settings:  req.Settings,
		Metadata:  req.Metadata,
		UpdatedBy: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.Tenant)
}

// UpdateStatusRequest represents a tenant status update request.
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended provisioning"`
}

// UpdateStatus handles updating a tenant's status.
func (h *TenantHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := h.getPathParam(r, "id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid tenant ID", nil)
		return
	}

	var req UpdateStatusRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	status := domain.TenantStatus(req.Status)
	performerID := middleware.GetUserID(r.Context())

	result, err := h.updateTenantUC.Execute(r.Context(), &usecase.UpdateTenantRequest{
		TenantID:  tenantID,
		Status:    &status,
		UpdatedBy: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.Tenant)
}

// UpdatePlanRequest represents a tenant plan update request.
type UpdatePlanRequest struct {
	Plan string `json:"plan" validate:"required,oneof=free starter professional enterprise"`
}

// UpdatePlan handles updating a tenant's plan.
func (h *TenantHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := h.getPathParam(r, "id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid tenant ID", nil)
		return
	}

	var req UpdatePlanRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	plan := domain.TenantPlan(req.Plan)
	performerID := middleware.GetUserID(r.Context())

	result, err := h.updateTenantUC.Execute(r.Context(), &usecase.UpdateTenantRequest{
		TenantID:  tenantID,
		Plan:      &plan,
		UpdatedBy: &performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result.Tenant)
}

// Delete handles soft-deleting a tenant.
func (h *TenantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := h.getPathParam(r, "id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid tenant ID", nil)
		return
	}

	performerID := middleware.GetUserID(r.Context())

	err = h.updateTenantUC.Delete(r.Context(), &usecase.DeleteTenantRequest{
		TenantID:  tenantID,
		DeletedBy: performerID,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "tenant deleted successfully",
	})
}

// GetStats handles getting tenant statistics.
func (h *TenantHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := h.getPathParam(r, "id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "invalid tenant ID", nil)
		return
	}

	result, err := h.getTenantUC.GetStats(r.Context(), tenantID)

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, result)
}

// CheckSlugAvailability handles checking if a slug is available.
func (h *TenantHandler) CheckSlugAvailability(w http.ResponseWriter, r *http.Request) {
	slug := iamhttp.NewQueryParams(r).String("slug", "")
	if slug == "" {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeBadRequest, "slug is required", nil)
		return
	}

	available, err := h.getTenantUC.CheckSlugAvailability(r.Context(), slug)

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"slug":      slug,
		"available": available,
	})
}
