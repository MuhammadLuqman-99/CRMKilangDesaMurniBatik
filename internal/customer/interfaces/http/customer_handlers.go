// Package http provides HTTP handlers for the Customer service.
package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/usecase"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Customer CRUD Handlers
// ============================================================================

// CreateCustomer handles POST /api/v1/customers
func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	var req dto.CreateCustomerRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.CreateCustomerInput{
		TenantID:  tenantID,
		UserID:    userID,
		Request:   &req,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	customer, err := h.createCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// GetCustomer handles GET /api/v1/customers/{id}
func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.GetCustomerInput{
		TenantID:   tenantID,
		CustomerID: customerID,
	}

	customer, err := h.getCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// GetCustomerByCode handles GET /api/v1/customers/code/{code}
func (h *Handler) GetCustomerByCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	code := getQueryString(r, "code")
	if code == "" {
		respondError(w, ErrMissingParameter("code"))
		return
	}

	input := usecase.GetCustomerByCodeInput{
		TenantID: tenantID,
		Code:     code,
	}

	customer, err := h.getCustomerByCode.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// UpdateCustomer handles PUT /api/v1/customers/{id}
func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.UpdateCustomerRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.UpdateCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Request:    &req,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.updateCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// DeleteCustomer handles DELETE /api/v1/customers/{id}
func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	hardDelete := getQueryBool(r, "hard")
	reason := getQueryString(r, "reason")

	input := usecase.DeleteCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		HardDelete: hardDelete != nil && *hardDelete,
		Reason:     reason,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	result, err := h.deleteCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result,
	})
}

// BulkDeleteCustomers handles POST /api/v1/customers/bulk-delete
func (h *Handler) BulkDeleteCustomers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	var req dto.BulkDeleteRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.BulkDeleteInput{
		TenantID:  tenantID,
		UserID:    userID,
		Request:   &req,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	result, err := h.bulkDeleteCustomers.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result,
	})
}

// ============================================================================
// Customer List/Search Handlers
// ============================================================================

// ListCustomers handles GET /api/v1/customers
func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	// Check if this is a search request (has query parameter)
	if q := getQueryString(r, "q"); q != "" {
		h.searchCustomers.Execute(ctx, usecase.SearchCustomersInput{
			TenantID: tenantID,
			Request:  buildSearchRequest(r),
		})
		return
	}

	input := usecase.ListCustomersInput{
		TenantID:  tenantID,
		Offset:    getQueryInt(r, "offset", 0),
		Limit:     getQueryInt(r, "limit", 20),
		SortBy:    getQueryString(r, "sort_by"),
		SortOrder: getQueryString(r, "sort_order"),
	}

	result, err := h.listCustomers.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Customers,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// SearchCustomers handles GET /api/v1/customers/search
func (h *Handler) SearchCustomers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	input := usecase.SearchCustomersInput{
		TenantID: tenantID,
		Request:  buildSearchRequest(r),
	}

	result, err := h.searchCustomers.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Customers,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// ListCustomersByOwner handles GET /api/v1/customers/by-owner/{owner_id}
func (h *Handler) ListCustomersByOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	ownerID, err := getUUIDParam(r, "owner_id")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.ListByOwnerInput{
		TenantID: tenantID,
		OwnerID:  ownerID,
		Offset:   getQueryInt(r, "offset", 0),
		Limit:    getQueryInt(r, "limit", 20),
	}

	result, err := h.listByOwner.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Customers,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// ListCustomersByStatus handles GET /api/v1/customers/by-status/{status}
func (h *Handler) ListCustomersByStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	statusParam := getQueryString(r, "status")
	if statusParam == "" {
		respondError(w, ErrMissingParameter("status"))
		return
	}

	input := usecase.ListByStatusInput{
		TenantID: tenantID,
		Status:   domain.CustomerStatus(statusParam),
		Offset:   getQueryInt(r, "offset", 0),
		Limit:    getQueryInt(r, "limit", 20),
	}

	result, err := h.listByStatus.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Customers,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// ListCustomersByTag handles GET /api/v1/customers/by-tag/{tag}
func (h *Handler) ListCustomersByTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	tag := getQueryString(r, "tag")
	if tag == "" {
		respondError(w, ErrMissingParameter("tag"))
		return
	}

	input := usecase.ListByTagInput{
		TenantID: tenantID,
		Tag:      tag,
		Offset:   getQueryInt(r, "offset", 0),
		Limit:    getQueryInt(r, "limit", 20),
	}

	result, err := h.listByTag.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Customers,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// ============================================================================
// Customer Action Handlers
// ============================================================================

// ChangeCustomerStatus handles POST /api/v1/customers/{id}/status
func (h *Handler) ChangeCustomerStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.ChangeStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.ChangeStatusInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Request:    &req,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.changeStatus.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// AssignOwner handles POST /api/v1/customers/{id}/assign
func (h *Handler) AssignOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.AssignOwnerRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.AssignOwnerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Request:    &req,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.assignOwner.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// ConvertCustomer handles POST /api/v1/customers/{id}/convert
func (h *Handler) ConvertCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.ConvertCustomerRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.ConvertCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Request:    &req,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.convertCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// RestoreCustomer handles POST /api/v1/customers/{id}/restore
func (h *Handler) RestoreCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.RestoreCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.restoreCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// ActivateCustomer handles POST /api/v1/customers/{id}/activate
func (h *Handler) ActivateCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.ActivateCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.activateCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// DeactivateCustomer handles POST /api/v1/customers/{id}/deactivate
func (h *Handler) DeactivateCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = decodeJSON(r, &req)

	input := usecase.DeactivateCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Reason:     req.Reason,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.deactivateCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// BlockCustomer handles POST /api/v1/customers/{id}/block
func (h *Handler) BlockCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.BlockCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Reason:     req.Reason,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.blockCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// UnblockCustomer handles POST /api/v1/customers/{id}/unblock
func (h *Handler) UnblockCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.UnblockCustomerInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	customer, err := h.unblockCustomer.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    customer,
	})
}

// AddToSegment handles POST /api/v1/customers/{customerId}/segments/{segmentId}
func (h *Handler) AddToSegment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	segmentID, err := getUUIDParam(r, "segmentId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.AddToSegmentInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		SegmentID:  segmentID,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	err = h.addToSegment.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "customer added to segment"},
	})
}

// RemoveFromSegment handles DELETE /api/v1/customers/{customerId}/segments/{segmentId}
func (h *Handler) RemoveFromSegment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	segmentID, err := getUUIDParam(r, "segmentId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.RemoveFromSegmentInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		SegmentID:  segmentID,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	err = h.removeFromSegment.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "customer removed from segment"},
	})
}

// ============================================================================
// Import/Export Handlers
// ============================================================================

// ImportCustomers handles POST /api/v1/customers/import
func (h *Handler) ImportCustomers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		respondError(w, ErrInvalidRequest("failed to parse multipart form"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, ErrMissingParameter("file"))
		return
	}
	defer file.Close()

	// Get import options from form
	options := dto.ImportOptions{
		SkipDuplicates: r.FormValue("skip_duplicates") == "true",
		UpdateExisting: r.FormValue("update_existing") == "true",
		DefaultStatus:  domain.CustomerStatus(r.FormValue("default_status")),
		DefaultType:    domain.CustomerType(r.FormValue("default_type")),
	}

	input := usecase.ImportCustomersInput{
		TenantID:  tenantID,
		UserID:    userID,
		File:      file,
		FileName:  header.Filename,
		FileSize:  header.Size,
		Options:   options,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	result, err := h.importCustomers.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result,
	})
}

// ExportCustomers handles GET /api/v1/customers/export
func (h *Handler) ExportCustomers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	format := getQueryString(r, "format")
	if format == "" {
		format = "csv"
	}

	fields := getQueryStringSlice(r, "fields")

	input := usecase.ExportCustomersInput{
		TenantID: tenantID,
		Filter:   buildSearchRequest(r),
		Format:   format,
		Fields:   fields,
	}

	result, err := h.exportCustomers.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	// Set content type based on format
	contentType := "text/csv"
	filename := "customers.csv"
	switch format {
	case "xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "customers.xlsx"
	case "json":
		contentType = "application/json"
		filename = "customers.json"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.WriteHeader(http.StatusOK)
	w.Write(result.Data)
}
