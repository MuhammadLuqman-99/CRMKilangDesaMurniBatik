// Package http provides HTTP handlers for the Customer service.
package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/usecase"
)

// ============================================================================
// Contact CRUD Handlers
// ============================================================================

// AddContact handles POST /api/v1/customers/{customer_id}/contacts
func (h *Handler) AddContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customer_id")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.CreateContactRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.AddContactInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Request:    &req,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	contact, err := h.addContact.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    contact,
	})
}

// GetContact handles GET /api/v1/contacts/{id}
func (h *Handler) GetContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	contactID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.GetContactInput{
		TenantID:  tenantID,
		ContactID: contactID,
	}

	contact, err := h.getContact.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    contact,
	})
}

// UpdateContact handles PUT /api/v1/contacts/{id}
func (h *Handler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	contactID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.UpdateContactRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.UpdateContactInput{
		TenantID:  tenantID,
		UserID:    userID,
		ContactID: contactID,
		Request:   &req,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	contact, err := h.updateContact.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    contact,
	})
}

// DeleteContact handles DELETE /api/v1/contacts/{id}
func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	contactID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.DeleteContactInput{
		TenantID:  tenantID,
		UserID:    userID,
		ContactID: contactID,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	result, err := h.deleteContact.Execute(ctx, input)
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
// Contact List/Search Handlers
// ============================================================================

// ListContacts handles GET /api/v1/customers/{customer_id}/contacts
func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customer_id")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.ListContactsInput{
		TenantID:   tenantID,
		CustomerID: customerID,
		Offset:     getQueryInt(r, "offset", 0),
		Limit:      getQueryInt(r, "limit", 20),
	}

	result, err := h.listContacts.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Contacts,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// SearchContacts handles GET /api/v1/contacts/search
func (h *Handler) SearchContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	input := usecase.SearchContactsInput{
		TenantID: tenantID,
		Request:  buildContactSearchRequest(r),
	}

	result, err := h.searchContacts.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Contacts,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// ============================================================================
// Contact Action Handlers
// ============================================================================

// SetPrimaryContact handles POST /api/v1/contacts/{id}/set-primary
func (h *Handler) SetPrimaryContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)
	userID := getUserID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	contactID, err := getUUIDParam(r, "id")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.SetPrimaryContactInput{
		TenantID:  tenantID,
		UserID:    userID,
		ContactID: contactID,
		IPAddress: getClientIP(r),
		UserAgent: getUserAgent(r),
	}

	contact, err := h.setPrimaryContact.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    contact,
	})
}
