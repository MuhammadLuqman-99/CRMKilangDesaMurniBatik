// Package http provides HTTP handlers for the Customer service.
package http

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/usecase"
)

// ============================================================================
// Note CRUD Handlers
// ============================================================================

// AddNote handles POST /api/v1/customers/{customerId}/notes
func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
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

	var req dto.CreateNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.AddNoteInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		Request:    &req,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	note, err := h.addNote.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    note,
	})
}

// GetNote handles GET /api/v1/customers/{customerId}/notes/{noteId}
func (h *Handler) GetNote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	noteID, err := getUUIDParam(r, "noteId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.GetNoteInput{
		TenantID:   tenantID,
		CustomerID: customerID,
		NoteID:     noteID,
	}

	note, err := h.getNote.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    note,
	})
}

// UpdateNote handles PUT /api/v1/customers/{customerId}/notes/{noteId}
func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
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

	noteID, err := getUUIDParam(r, "noteId")
	if err != nil {
		respondError(w, err)
		return
	}

	var req dto.UpdateNoteRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}

	input := usecase.UpdateNoteInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		NoteID:     noteID,
		Request:    &req,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	note, err := h.updateNote.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    note,
	})
}

// DeleteNote handles DELETE /api/v1/customers/{customerId}/notes/{noteId}
func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
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

	noteID, err := getUUIDParam(r, "noteId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.DeleteNoteInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		NoteID:     noteID,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	err = h.deleteNote.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "note deleted"},
	})
}

// ============================================================================
// Note List Handlers
// ============================================================================

// ListNotes handles GET /api/v1/customers/{customerId}/notes
func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantID(ctx)

	if tenantID == uuid.Nil {
		respondError(w, ErrUnauthorized("tenant_id is required"))
		return
	}

	customerID, err := getUUIDParam(r, "customerId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.ListNotesInput{
		TenantID:   tenantID,
		CustomerID: customerID,
		Offset:     getQueryInt(r, "offset", 0),
		Limit:      getQueryInt(r, "limit", 20),
		SortBy:     getQueryString(r, "sort_by"),
		SortOrder:  getQueryString(r, "sort_order"),
		PinnedOnly: getQueryBool(r, "pinned_only"),
	}

	result, err := h.listNotes.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    result.Notes,
		Meta: &MetaResponse{
			Total:   result.Total,
			Offset:  result.Offset,
			Limit:   result.Limit,
			HasMore: result.HasMore,
		},
	})
}

// ============================================================================
// Note Action Handlers
// ============================================================================

// PinNote handles POST /api/v1/customers/{customerId}/notes/{noteId}/pin
func (h *Handler) PinNote(w http.ResponseWriter, r *http.Request) {
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

	noteID, err := getUUIDParam(r, "noteId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.PinNoteInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		NoteID:     noteID,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	note, err := h.pinNote.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    note,
	})
}

// UnpinNote handles DELETE /api/v1/customers/{customerId}/notes/{noteId}/pin
func (h *Handler) UnpinNote(w http.ResponseWriter, r *http.Request) {
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

	noteID, err := getUUIDParam(r, "noteId")
	if err != nil {
		respondError(w, err)
		return
	}

	input := usecase.UnpinNoteInput{
		TenantID:   tenantID,
		UserID:     userID,
		CustomerID: customerID,
		NoteID:     noteID,
		IPAddress:  getClientIP(r),
		UserAgent:  getUserAgent(r),
	}

	note, err := h.unpinNote.Execute(ctx, input)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    note,
	})
}
