// Package http provides HTTP handlers for the Customer service.
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

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/dto"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/usecase"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
)

// Handler holds all HTTP handlers for the Customer service.
type Handler struct {
	// Customer use cases
	createCustomer       *usecase.CreateCustomerUseCase
	updateCustomer       *usecase.UpdateCustomerUseCase
	deleteCustomer       *usecase.DeleteCustomerUseCase
	bulkDeleteCustomers  *usecase.BulkDeleteCustomersUseCase
	getCustomer          *usecase.GetCustomerUseCase
	getCustomerByCode    *usecase.GetCustomerByCodeUseCase
	searchCustomers      *usecase.SearchCustomersUseCase
	listCustomers        *usecase.ListCustomersUseCase
	listByOwner          *usecase.ListCustomersByOwnerUseCase
	listByStatus         *usecase.ListCustomersByStatusUseCase
	listByTag            *usecase.ListCustomersByTagUseCase
	changeStatus         *usecase.ChangeCustomerStatusUseCase
	assignOwner          *usecase.AssignOwnerUseCase
	convertCustomer      *usecase.ConvertCustomerUseCase
	importCustomers      *usecase.ImportCustomersUseCase
	exportCustomers      *usecase.ExportCustomersUseCase
	restoreCustomer      *usecase.RestoreCustomerUseCase
	activateCustomer     *usecase.ActivateCustomerUseCase
	deactivateCustomer   *usecase.DeactivateCustomerUseCase
	blockCustomer        *usecase.BlockCustomerUseCase
	unblockCustomer      *usecase.UnblockCustomerUseCase

	// Contact use cases
	addContact           *usecase.AddContactUseCase
	updateContact        *usecase.UpdateContactUseCase
	deleteContact        *usecase.DeleteContactUseCase
	getContact           *usecase.GetContactUseCase
	listContacts         *usecase.ListContactsUseCase
	setPrimaryContact    *usecase.SetPrimaryContactUseCase
	searchContacts       *usecase.SearchContactsUseCase

	// Note use cases
	addNote              *usecase.AddNoteUseCase
	getNote              *usecase.GetNoteUseCase
	updateNote           *usecase.UpdateNoteUseCase
	deleteNote           *usecase.DeleteNoteUseCase
	listNotes            *usecase.ListNotesUseCase
	pinNote              *usecase.PinNoteUseCase
	unpinNote            *usecase.UnpinNoteUseCase

	// Activity use cases
	logActivity          *usecase.LogActivityUseCase
	getActivity          *usecase.GetActivityUseCase
	listActivities       *usecase.ListActivitiesUseCase

	// Segment use cases
	createSegment        *usecase.CreateSegmentUseCase
	getSegment           *usecase.GetSegmentUseCase
	updateSegment        *usecase.UpdateSegmentUseCase
	deleteSegment        *usecase.DeleteSegmentUseCase
	listSegments         *usecase.ListSegmentsUseCase
	refreshSegment       *usecase.RefreshSegmentUseCase
	getSegmentCustomers  *usecase.GetSegmentCustomersUseCase
	addToSegment         *usecase.AddToSegmentUseCase
	removeFromSegment    *usecase.RemoveFromSegmentUseCase

	// Import use cases
	getImportStatus      *usecase.GetImportStatusUseCase
	getImportErrors      *usecase.GetImportErrorsUseCase
	listImports          *usecase.ListImportsUseCase
	cancelImport         *usecase.CancelImportUseCase
}

// NewHandler is now defined in routes.go using HandlerDependencies pattern.

// ============================================================================
// Helper Functions
// ============================================================================

// contextKey is a type for context keys.
type contextKey string

const (
	contextKeyTenantID contextKey = "tenant_id"
	contextKeyUserID   contextKey = "user_id"
)

// getTenantID extracts the tenant ID from the context.
func getTenantID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(contextKeyTenantID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// getUserID extracts the user ID from the context.
func getUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(contextKeyUserID).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

// getUUIDParam extracts a UUID parameter from the URL.
func getUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
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
func getQueryInt(r *http.Request, name string, defaultValue int) int {
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

// getQueryString extracts a string query parameter.
func getQueryString(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// getQueryStringSlice extracts a comma-separated string slice query parameter.
func getQueryStringSlice(r *http.Request, name string) []string {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

// getQueryUUID extracts a UUID query parameter.
func getQueryUUID(r *http.Request, name string) *uuid.UUID {
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

// getQueryTime extracts a time query parameter (RFC3339 format).
func getQueryTime(r *http.Request, name string) *time.Time {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &t
}

// getQueryBool extracts a boolean query parameter.
func getQueryBool(r *http.Request, name string) *bool {
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

// getClientIP extracts the client IP address from the request.
func getClientIP(r *http.Request) string {
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
	return strings.Split(r.RemoteAddr, ":")[0]
}

// getUserAgent extracts the user agent from the request.
func getUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// decodeJSON decodes JSON from the request body.
func decodeJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		return ErrInvalidJSON(err.Error())
	}
	return nil
}

// respondJSON writes a JSON response.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Log error but don't fail - headers already sent
		}
	}
}

// respondError writes an error response.
func respondError(w http.ResponseWriter, err error) {
	httpErr := toHTTPError(err)
	respondJSON(w, httpErr.StatusCode, httpErr)
}

// ============================================================================
// Request/Response Types
// ============================================================================

// APIResponse is the standard API response wrapper.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
	Meta    *MetaResponse  `json:"meta,omitempty"`
}

// MetaResponse contains pagination metadata.
type MetaResponse struct {
	Total   int64  `json:"total"`
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	HasMore bool   `json:"has_more"`
}

// Note: Health check endpoints are defined in routes.go

// ============================================================================
// Search Request Builder
// ============================================================================

// buildSearchRequest builds a SearchCustomersRequest from query parameters.
func buildSearchRequest(r *http.Request) *dto.SearchCustomersRequest {
	req := &dto.SearchCustomersRequest{
		Query:     getQueryString(r, "q"),
		Offset:    getQueryInt(r, "offset", 0),
		Limit:     getQueryInt(r, "limit", 20),
		SortBy:    getQueryString(r, "sort_by"),
		SortOrder: getQueryString(r, "sort_order"),
	}

	// Types
	if types := getQueryStringSlice(r, "types"); len(types) > 0 {
		for _, t := range types {
			req.Types = append(req.Types, domain.CustomerType(t))
		}
	}

	// Statuses
	if statuses := getQueryStringSlice(r, "statuses"); len(statuses) > 0 {
		for _, s := range statuses {
			req.Statuses = append(req.Statuses, domain.CustomerStatus(s))
		}
	}

	// Tiers
	if tiers := getQueryStringSlice(r, "tiers"); len(tiers) > 0 {
		for _, t := range tiers {
			req.Tiers = append(req.Tiers, domain.CustomerTier(t))
		}
	}

	// Tags
	req.Tags = getQueryStringSlice(r, "tags")

	// Owner IDs
	if ownerIDs := getQueryStringSlice(r, "owner_ids"); len(ownerIDs) > 0 {
		for _, id := range ownerIDs {
			if uid, err := uuid.Parse(id); err == nil {
				req.OwnerIDs = append(req.OwnerIDs, uid)
			}
		}
	}

	// Date filters
	if after := getQueryTime(r, "created_after"); after != nil {
		req.CreatedAfter = after
	}
	if before := getQueryTime(r, "created_before"); before != nil {
		req.CreatedBefore = before
	}

	req.HasDeals = getQueryBool(r, "has_deals")
	req.HasOpenDeals = getQueryBool(r, "has_open_deals")
	req.IncludeDeleted = getQueryBool(r, "include_deleted") != nil && *getQueryBool(r, "include_deleted")

	return req
}

// buildContactSearchRequest builds a SearchContactsRequest from query parameters.
func buildContactSearchRequest(r *http.Request) *dto.SearchContactsRequest {
	req := &dto.SearchContactsRequest{
		Query:     getQueryString(r, "q"),
		Offset:    getQueryInt(r, "offset", 0),
		Limit:     getQueryInt(r, "limit", 20),
		SortBy:    getQueryString(r, "sort_by"),
		SortOrder: getQueryString(r, "sort_order"),
	}

	if customerID := getQueryUUID(r, "customer_id"); customerID != nil {
		req.CustomerID = customerID
	}

	// Statuses
	if statuses := getQueryStringSlice(r, "statuses"); len(statuses) > 0 {
		for _, s := range statuses {
			req.Statuses = append(req.Statuses, domain.ContactStatus(s))
		}
	}

	// Roles
	if roles := getQueryStringSlice(r, "roles"); len(roles) > 0 {
		for _, role := range roles {
			req.Roles = append(req.Roles, domain.ContactRole(role))
		}
	}

	req.IsPrimary = getQueryBool(r, "is_primary")
	req.Tags = getQueryStringSlice(r, "tags")

	return req
}
