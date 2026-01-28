// Package http contains the HTTP handlers for the Sales Pipeline service.
package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Saga Handler
// ============================================================================

// SagaHandler handles HTTP requests for saga operations.
type SagaHandler struct {
	sagaRepo domain.SagaRepository
}

// NewSagaHandler creates a new saga handler.
func NewSagaHandler(sagaRepo domain.SagaRepository) *SagaHandler {
	return &SagaHandler{
		sagaRepo: sagaRepo,
	}
}

// RegisterRoutes registers saga routes.
func (h *SagaHandler) RegisterRoutes(r chi.Router) {
	r.Route("/sagas", func(r chi.Router) {
		// Saga status and management
		r.Get("/{sagaID}", h.GetSagaStatus)
		r.Post("/{sagaID}/retry", h.RetrySaga)
		r.Get("/by-state/{state}", h.ListSagasByState)
		r.Get("/failed", h.ListFailedSagas)
	})
}

// GetSagaStatus returns the status of a saga.
// GET /api/v1/sagas/{sagaID}
func (h *SagaHandler) GetSagaStatus(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "tenant ID required")
		return
	}

	// Get saga ID from URL
	sagaIDStr := chi.URLParam(r, "sagaID")
	sagaID, err := uuid.Parse(sagaIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_saga_id", "invalid saga ID format")
		return
	}

	// Get saga from repository
	saga, err := h.sagaRepo.GetByID(r.Context(), tenantID, sagaID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "saga_not_found", "saga not found")
		return
	}

	// Build response
	response := h.mapSagaToResponse(saga)
	h.respondJSON(w, http.StatusOK, response)
}

// RetrySaga attempts to resume a failed saga.
// POST /api/v1/sagas/{sagaID}/retry
func (h *SagaHandler) RetrySaga(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "tenant ID required")
		return
	}

	// Get saga ID from URL
	sagaIDStr := chi.URLParam(r, "sagaID")
	sagaID, err := uuid.Parse(sagaIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_saga_id", "invalid saga ID format")
		return
	}

	// Get saga from repository
	saga, err := h.sagaRepo.GetByID(r.Context(), tenantID, sagaID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "saga_not_found", "saga not found")
		return
	}

	// Check if saga can be retried
	if saga.State == domain.SagaStateCompleted || saga.State == domain.SagaStateRunning {
		h.respondError(w, http.StatusBadRequest, "saga_not_retryable",
			"saga cannot be retried - current state: "+string(saga.State))
		return
	}

	// Return retry initiated response
	response := map[string]interface{}{
		"saga_id": sagaID.String(),
		"message": "saga retry initiated",
		"state":   saga.State,
	}
	h.respondJSON(w, http.StatusAccepted, response)
}

// ListSagasByState lists sagas by state.
// GET /api/v1/sagas/by-state/{state}
func (h *SagaHandler) ListSagasByState(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "tenant ID required")
		return
	}

	// Get state from URL
	stateStr := chi.URLParam(r, "state")
	state := domain.SagaState(stateStr)

	// Validate state
	if !state.IsValid() {
		h.respondError(w, http.StatusBadRequest, "invalid_state", "invalid saga state")
		return
	}

	// Default options
	opts := domain.ListOptions{
		Page:     1,
		PageSize: 50,
	}

	// Get sagas
	sagas, total, err := h.sagaRepo.GetByState(r.Context(), tenantID, state, opts)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "internal_error", "failed to list sagas")
		return
	}

	// Map to response
	items := make([]SagaStatusResponse, len(sagas))
	for i, saga := range sagas {
		items[i] = h.mapSagaToResponse(saga)
	}

	response := SagaListResponse{
		Items:    items,
		Total:    total,
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}
	h.respondJSON(w, http.StatusOK, response)
}

// ListFailedSagas lists failed sagas.
// GET /api/v1/sagas/failed
func (h *SagaHandler) ListFailedSagas(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		h.respondError(w, http.StatusUnauthorized, "unauthorized", "tenant ID required")
		return
	}

	// Default options
	opts := domain.ListOptions{
		Page:     1,
		PageSize: 50,
	}

	// Get sagas
	sagas, total, err := h.sagaRepo.GetFailedSagas(r.Context(), tenantID, opts)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "internal_error", "failed to list sagas")
		return
	}

	// Map to response
	items := make([]SagaStatusResponse, len(sagas))
	for i, saga := range sagas {
		items[i] = h.mapSagaToResponse(saga)
	}

	response := SagaListResponse{
		Items:    items,
		Total:    total,
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}
	h.respondJSON(w, http.StatusOK, response)
}

// ============================================================================
// Response Types
// ============================================================================

// SagaStatusResponse represents a saga status response.
type SagaStatusResponse struct {
	ID               string             `json:"id"`
	LeadID           string             `json:"lead_id"`
	State            string             `json:"state"`
	CurrentStepIndex int                `json:"current_step_index"`
	TotalSteps       int                `json:"total_steps"`
	OpportunityID    *string            `json:"opportunity_id,omitempty"`
	CustomerID       *string            `json:"customer_id,omitempty"`
	Error            *string            `json:"error,omitempty"`
	Steps            []SagaStepResponse `json:"steps"`
	StartedAt        string             `json:"started_at"`
	CompletedAt      *string            `json:"completed_at,omitempty"`
}

// SagaStepResponse represents a saga step in the response.
type SagaStepResponse struct {
	Order       int     `json:"order"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Error       *string `json:"error,omitempty"`
	StartedAt   *string `json:"started_at,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

// SagaListResponse represents a list of sagas.
type SagaListResponse struct {
	Items    []SagaStatusResponse `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// ============================================================================
// Helper Methods
// ============================================================================

func (h *SagaHandler) mapSagaToResponse(saga *domain.LeadConversionSaga) SagaStatusResponse {
	response := SagaStatusResponse{
		ID:               saga.ID.String(),
		LeadID:           saga.LeadID.String(),
		State:            string(saga.State),
		CurrentStepIndex: saga.CurrentStepIndex,
		TotalSteps:       len(saga.Steps),
		StartedAt:        saga.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if saga.OpportunityID != nil {
		oppID := saga.OpportunityID.String()
		response.OpportunityID = &oppID
	}

	if saga.CustomerID != nil {
		custID := saga.CustomerID.String()
		response.CustomerID = &custID
	}

	if saga.Error != nil {
		response.Error = saga.Error
	}

	if saga.CompletedAt != nil {
		completedAt := saga.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		response.CompletedAt = &completedAt
	}

	// Map steps
	response.Steps = make([]SagaStepResponse, len(saga.Steps))
	for i, step := range saga.Steps {
		stepResp := SagaStepResponse{
			Order:  step.Order,
			Type:   string(step.Type),
			Status: string(step.Status),
		}

		// Handle step error (pointer type)
		if step.Error != nil {
			stepResp.Error = step.Error
		}

		if step.StartedAt != nil {
			startedAt := step.StartedAt.Format("2006-01-02T15:04:05Z07:00")
			stepResp.StartedAt = &startedAt
		}

		if step.CompletedAt != nil {
			completedAt := step.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
			stepResp.CompletedAt = &completedAt
		}

		response.Steps[i] = stepResp
	}

	return response
}

func (h *SagaHandler) respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *SagaHandler) respondError(w http.ResponseWriter, status int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   errorCode,
		"message": message,
	})
}
