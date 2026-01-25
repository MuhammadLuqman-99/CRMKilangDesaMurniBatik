// Package http provides HTTP handlers for the Sales Pipeline service.
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
	"github.com/kilang-desa-murni/crm/internal/sales/application/usecase"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Handler Structure
// ============================================================================

// Handler holds all HTTP handlers for the Sales Pipeline service.
type Handler struct {
	// Lead use cases
	leadUseCase usecase.LeadUseCase

	// Opportunity use cases
	opportunityUseCase usecase.OpportunityUseCase

	// Deal use cases
	dealUseCase usecase.DealUseCase

	// Pipeline use cases
	pipelineUseCase usecase.PipelineUseCase

	// Middleware configuration
	middlewareConfig MiddlewareConfig
}

// HandlerDependencies contains all dependencies needed to create handlers.
type HandlerDependencies struct {
	LeadUseCase        usecase.LeadUseCase
	OpportunityUseCase usecase.OpportunityUseCase
	DealUseCase        usecase.DealUseCase
	PipelineUseCase    usecase.PipelineUseCase
	MiddlewareConfig   MiddlewareConfig
}

// NewHandler creates a new handler with all dependencies.
func NewHandler(deps HandlerDependencies) *Handler {
	config := deps.MiddlewareConfig
	if config.JWTSecret == "" {
		config = DefaultMiddlewareConfig()
	}

	return &Handler{
		leadUseCase:        deps.LeadUseCase,
		opportunityUseCase: deps.OpportunityUseCase,
		dealUseCase:        deps.DealUseCase,
		pipelineUseCase:    deps.PipelineUseCase,
		middlewareConfig:   config,
	}
}

// ============================================================================
// Context Helpers
// ============================================================================

// getTenantID extracts the tenant ID from the context.
func (h *Handler) getTenantID(ctx context.Context) (uuid.UUID, error) {
	if id, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok && id != uuid.Nil {
		return id, nil
	}
	return uuid.Nil, ErrUnauthorized("tenant identification required")
}

// getUserID extracts the user ID from the context.
func (h *Handler) getUserID(ctx context.Context) (*uuid.UUID, error) {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok && id != uuid.Nil {
		return &id, nil
	}
	return nil, nil
}

// getRequestID extracts the request ID from the context.
func (h *Handler) getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// ============================================================================
// URL Parameter Helpers
// ============================================================================

// getUUIDParam extracts a UUID parameter from the URL.
func (h *Handler) getUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	param := chi.URLParam(r, name)
	if param == "" {
		return uuid.Nil, ErrMissingParameter(name)
	}
	id, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, ErrInvalidParameter(name, "invalid UUID format")
	}
	return id, nil
}

// getQueryInt extracts an integer query parameter with default value.
func (h *Handler) getQueryInt(r *http.Request, name string, defaultValue int) int {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// getQueryInt64 extracts an int64 query parameter.
func (h *Handler) getQueryInt64(r *http.Request, name string) *int64 {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

// getQueryString extracts a string query parameter.
func (h *Handler) getQueryString(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// getQueryStringPtr extracts a string query parameter as pointer.
func (h *Handler) getQueryStringPtr(r *http.Request, name string) *string {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	return &value
}

// getQueryStringSlice extracts a comma-separated string slice query parameter.
func (h *Handler) getQueryStringSlice(r *http.Request, name string) []string {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// getQueryUUID extracts a UUID query parameter.
func (h *Handler) getQueryUUID(r *http.Request, name string) *uuid.UUID {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil
	}
	return &id
}

// getQueryUUIDSlice extracts a comma-separated UUID slice query parameter.
func (h *Handler) getQueryUUIDSlice(r *http.Request, name string) []uuid.UUID {
	values := h.getQueryStringSlice(r, name)
	if len(values) == 0 {
		return nil
	}
	result := make([]uuid.UUID, 0, len(values))
	for _, v := range values {
		if id, err := uuid.Parse(v); err == nil {
			result = append(result, id)
		}
	}
	return result
}

// getQueryTime extracts a time query parameter (RFC3339 format).
func (h *Handler) getQueryTime(r *http.Request, name string) *time.Time {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		// Try date only format
		t, err = time.Parse("2006-01-02", value)
		if err != nil {
			return nil
		}
	}
	return &t
}

// getQueryDate extracts a date query parameter (YYYY-MM-DD format).
func (h *Handler) getQueryDate(r *http.Request, name string) *time.Time {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &t
}

// getQueryBool extracts a boolean query parameter.
func (h *Handler) getQueryBool(r *http.Request, name string) *bool {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}
	return &b
}

// ============================================================================
// Request Helpers
// ============================================================================

// getClientIP extracts the client IP address from the request.
func (h *Handler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to remote address
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// getUserAgent extracts the user agent from the request.
func (h *Handler) getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// decodeJSON decodes JSON from the request body.
func (h *Handler) decodeJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return ErrInvalidJSON(err.Error())
	}
	return nil
}

// ============================================================================
// Response Helpers
// ============================================================================

// APIResponse is the standard API response wrapper.
type APIResponse struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
	Meta    *MetaResponse  `json:"meta,omitempty"`
}

// MetaResponse contains pagination metadata.
type MetaResponse struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int64 `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

// respondJSON writes a JSON response.
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Log error but don't fail - headers already sent
		}
	}
}

// respondSuccess writes a success response with data.
func (h *Handler) respondSuccess(w http.ResponseWriter, status int, data interface{}) {
	h.respondJSON(w, status, APIResponse{
		Success: true,
		Data:    data,
	})
}

// respondCreated writes a created response with data.
func (h *Handler) respondCreated(w http.ResponseWriter, data interface{}) {
	h.respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

// respondList writes a paginated list response.
func (h *Handler) respondList(w http.ResponseWriter, data interface{}, total int64, page, pageSize int) {
	totalPages := total / int64(pageSize)
	if total%int64(pageSize) > 0 {
		totalPages++
	}

	h.respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta: &MetaResponse{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			HasMore:    int64(page) < totalPages,
		},
	})
}

// respondNoContent writes a no content response.
func (h *Handler) respondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// respondError writes an error response.
func (h *Handler) respondError(w http.ResponseWriter, err error) {
	httpErr := toHTTPError(err)
	h.respondJSON(w, httpErr.StatusCode, APIResponse{
		Success: false,
		Error:   httpErr,
	})
}

// mapAppError maps application/domain errors to HTTP errors
func (h *Handler) mapAppError(err error) error {
	return mapAppError(err)
}

// ============================================================================
// Filter Builders
// ============================================================================

// buildLeadFilter builds a LeadFilter from query parameters.
func (h *Handler) buildLeadFilter(r *http.Request) dto.LeadFilter {
	filter := dto.LeadFilter{}

	// Status filters
	if statuses := h.getQueryStringSlice(r, "statuses"); len(statuses) > 0 {
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, domain.LeadStatus(s))
		}
	}

	// Source filters
	if sources := h.getQueryStringSlice(r, "sources"); len(sources) > 0 {
		for _, s := range sources {
			filter.Sources = append(filter.Sources, domain.LeadSource(s))
		}
	}

	// Assignment filters
	filter.OwnerIDs = h.getQueryUUIDSlice(r, "owner_ids")
	filter.Unassigned = h.getQueryBool(r, "unassigned")

	// Score filters
	if minScore := h.getQueryInt(r, "min_score", -1); minScore >= 0 {
		filter.MinScore = &minScore
	}
	if maxScore := h.getQueryInt(r, "max_score", -1); maxScore >= 0 {
		filter.MaxScore = &maxScore
	}

	// Time filters
	filter.CreatedAfter = h.getQueryTime(r, "created_after")
	filter.CreatedBefore = h.getQueryTime(r, "created_before")
	filter.UpdatedAfter = h.getQueryTime(r, "updated_after")

	// Search and tags
	filter.SearchQuery = h.getQueryString(r, "q")
	filter.Tags = h.getQueryStringSlice(r, "tags")

	// Campaign filter
	filter.CampaignID = h.getQueryUUID(r, "campaign_id")

	return filter
}

// buildOpportunityFilter builds an OpportunityFilter from query parameters.
func (h *Handler) buildOpportunityFilter(r *http.Request) dto.OpportunityFilter {
	filter := dto.OpportunityFilter{}

	// Status filters
	if statuses := h.getQueryStringSlice(r, "statuses"); len(statuses) > 0 {
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, domain.OpportunityStatus(s))
		}
	}

	// Pipeline filters
	filter.PipelineIDs = h.getQueryUUIDSlice(r, "pipeline_ids")
	filter.StageIDs = h.getQueryUUIDSlice(r, "stage_ids")

	// Relationship filters
	filter.CustomerIDs = h.getQueryUUIDSlice(r, "customer_ids")
	filter.ContactIDs = h.getQueryUUIDSlice(r, "contact_ids")
	filter.OwnerIDs = h.getQueryUUIDSlice(r, "owner_ids")
	filter.LeadID = h.getQueryUUID(r, "lead_id")

	// Value filters
	filter.MinAmount = h.getQueryInt64(r, "min_amount")
	filter.MaxAmount = h.getQueryInt64(r, "max_amount")
	filter.Currency = h.getQueryStringPtr(r, "currency")

	// Probability filter
	if minProb := h.getQueryInt(r, "min_probability", -1); minProb >= 0 {
		filter.MinProbability = &minProb
	}
	if maxProb := h.getQueryInt(r, "max_probability", -1); maxProb >= 0 {
		filter.MaxProbability = &maxProb
	}

	// Date filters
	filter.ExpectedCloseDateAfter = h.getQueryDate(r, "expected_close_after")
	filter.ExpectedCloseDateBefore = h.getQueryDate(r, "expected_close_before")
	filter.CreatedAfter = h.getQueryTime(r, "created_after")
	filter.CreatedBefore = h.getQueryTime(r, "created_before")

	// Search
	filter.SearchQuery = h.getQueryString(r, "q")

	// Product and source filters
	filter.ProductIDs = h.getQueryUUIDSlice(r, "product_ids")
	filter.Sources = h.getQueryStringSlice(r, "sources")

	return filter
}

// buildDealFilter builds a DealFilter from query parameters.
func (h *Handler) buildDealFilter(r *http.Request) dto.DealFilter {
	filter := dto.DealFilter{}

	// Status filters
	if statuses := h.getQueryStringSlice(r, "statuses"); len(statuses) > 0 {
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, domain.DealStatus(s))
		}
	}

	// Relationship filters
	filter.CustomerIDs = h.getQueryUUIDSlice(r, "customer_ids")
	filter.OpportunityID = h.getQueryUUID(r, "opportunity_id")
	filter.OwnerIDs = h.getQueryUUIDSlice(r, "owner_ids")

	// Value filters
	filter.MinAmount = h.getQueryInt64(r, "min_amount")
	filter.MaxAmount = h.getQueryInt64(r, "max_amount")
	filter.Currency = h.getQueryStringPtr(r, "currency")

	// Payment status
	filter.HasPendingPayments = h.getQueryBool(r, "has_pending_payments")
	filter.FullyPaid = h.getQueryBool(r, "fully_paid")

	// Fulfillment filter
	if progress := h.getQueryInt(r, "min_fulfillment", -1); progress >= 0 {
		filter.FulfillmentProgress = &progress
	}

	// Date filters
	filter.ClosedDateAfter = h.getQueryTime(r, "closed_after")
	filter.ClosedDateBefore = h.getQueryTime(r, "closed_before")
	filter.SignedDateAfter = h.getQueryTime(r, "signed_after")
	filter.SignedDateBefore = h.getQueryTime(r, "signed_before")

	// Search
	filter.SearchQuery = h.getQueryString(r, "q")
	filter.DealNumber = h.getQueryStringPtr(r, "deal_number")

	return filter
}

// buildListOptions builds ListOptions from query parameters.
func (h *Handler) buildListOptions(r *http.Request) dto.ListOptions {
	includeDeleted := h.getQueryBool(r, "include_deleted")
	return dto.ListOptions{
		Page:           h.getQueryInt(r, "page", 1),
		PageSize:       h.getQueryInt(r, "page_size", 20),
		SortBy:         h.getQueryString(r, "sort_by"),
		SortOrder:      h.getQueryString(r, "sort_order"),
		IncludeDeleted: includeDeleted != nil && *includeDeleted,
	}
}
