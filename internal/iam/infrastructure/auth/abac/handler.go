// Package abac provides Attribute-Based Access Control infrastructure.
package abac

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// PolicyHandler handles ABAC policy management HTTP requests.
type PolicyHandler struct {
	policyRepo PolicyRepository
	engine     *Engine
}

// NewPolicyHandler creates a new policy handler.
func NewPolicyHandler(policyRepo PolicyRepository, engine *Engine) *PolicyHandler {
	return &PolicyHandler{
		policyRepo: policyRepo,
		engine:     engine,
	}
}

// CreatePolicyRequest represents a request to create a policy.
type CreatePolicyRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Rules       []PolicyRule `json:"rules"`
	Targets     PolicyTarget `json:"targets,omitempty"`
	Priority    int          `json:"priority"`
	Enabled     bool         `json:"enabled"`
	Metadata    PolicyMeta   `json:"metadata,omitempty"`
}

// UpdatePolicyRequest represents a request to update a policy.
type UpdatePolicyRequest struct {
	Name        *string       `json:"name,omitempty"`
	Description *string       `json:"description,omitempty"`
	Rules       []PolicyRule  `json:"rules,omitempty"`
	Targets     *PolicyTarget `json:"targets,omitempty"`
	Priority    *int          `json:"priority,omitempty"`
	Enabled     *bool         `json:"enabled,omitempty"`
	Metadata    *PolicyMeta   `json:"metadata,omitempty"`
}

// EvaluateRequest represents a request to evaluate ABAC policies.
type EvaluateRequest struct {
	Subject     map[string]interface{} `json:"subject"`
	Resource    map[string]interface{} `json:"resource"`
	Action      map[string]interface{} `json:"action"`
	Environment map[string]interface{} `json:"environment,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// HandleCreatePolicy handles policy creation.
func (h *PolicyHandler) HandleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "tenant context required")
		return
	}

	// Get user ID for audit
	var createdBy *uuid.UUID
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		createdBy = &userID
	}

	// Validate request
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if len(req.Rules) == 0 {
		h.writeError(w, http.StatusBadRequest, "at least one rule is required")
		return
	}

	// Create policy
	policy := &Policy{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Rules:       req.Rules,
		Targets:     req.Targets,
		Priority:    req.Priority,
		Enabled:     req.Enabled,
		Metadata:    req.Metadata,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		CreatedBy:   createdBy,
		Version:     1,
	}

	// Assign rule IDs if not provided
	for i := range policy.Rules {
		if policy.Rules[i].ID == "" {
			policy.Rules[i].ID = uuid.New().String()
		}
	}

	if err := h.policyRepo.Create(r.Context(), policy); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate cache
	h.engine.InvalidateCache(tenantID)

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    policy,
	})
}

// HandleGetPolicy handles getting a single policy.
func (h *PolicyHandler) HandleGetPolicy(w http.ResponseWriter, r *http.Request) {
	policyIDStr := r.URL.Query().Get("id")
	if policyIDStr == "" {
		// Try to get from path
		policyIDStr = getPathParam(r, "id")
	}

	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid policy ID")
		return
	}

	policy, err := h.policyRepo.FindByID(r.Context(), policyID)
	if err != nil {
		if err == ErrPolicyNotFound {
			h.writeError(w, http.StatusNotFound, "policy not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    policy,
	})
}

// HandleListPolicies handles listing policies.
func (h *PolicyHandler) HandleListPolicies(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "tenant context required")
		return
	}

	// Parse query parameters
	filter := PolicyFilter{
		TenantID: &tenantID,
	}

	if name := r.URL.Query().Get("name"); name != "" {
		filter.Name = &name
	}

	if enabledStr := r.URL.Query().Get("enabled"); enabledStr != "" {
		enabled := enabledStr == "true"
		filter.Enabled = &enabled
	}

	if tags := r.URL.Query()["tags"]; len(tags) > 0 {
		filter.Tags = tags
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 50
	}

	filter.SortBy = r.URL.Query().Get("sort_by")
	filter.SortOrder = r.URL.Query().Get("sort_order")

	policies, total, err := h.policyRepo.List(r.Context(), filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"policies": policies,
			"total":    total,
			"offset":   filter.Offset,
			"limit":    filter.Limit,
		},
	})
}

// HandleUpdatePolicy handles policy updates.
func (h *PolicyHandler) HandleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	policyIDStr := getPathParam(r, "id")
	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid policy ID")
		return
	}

	var req UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get existing policy
	policy, err := h.policyRepo.FindByID(r.Context(), policyID)
	if err != nil {
		if err == ErrPolicyNotFound {
			h.writeError(w, http.StatusNotFound, "policy not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get user ID for audit
	var updatedBy *uuid.UUID
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		updatedBy = &userID
	}

	// Apply updates
	if req.Name != nil {
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}
	if req.Rules != nil {
		policy.Rules = req.Rules
		// Assign rule IDs if not provided
		for i := range policy.Rules {
			if policy.Rules[i].ID == "" {
				policy.Rules[i].ID = uuid.New().String()
			}
		}
	}
	if req.Targets != nil {
		policy.Targets = *req.Targets
	}
	if req.Priority != nil {
		policy.Priority = *req.Priority
	}
	if req.Enabled != nil {
		policy.Enabled = *req.Enabled
	}
	if req.Metadata != nil {
		policy.Metadata = *req.Metadata
	}

	policy.UpdatedAt = time.Now().UTC()
	policy.UpdatedBy = updatedBy
	policy.Version++

	if err := h.policyRepo.Update(r.Context(), policy); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate cache
	h.engine.InvalidateCache(policy.TenantID)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    policy,
	})
}

// HandleDeletePolicy handles policy deletion.
func (h *PolicyHandler) HandleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	policyIDStr := getPathParam(r, "id")
	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid policy ID")
		return
	}

	// Get policy to get tenant ID
	policy, err := h.policyRepo.FindByID(r.Context(), policyID)
	if err != nil {
		if err == ErrPolicyNotFound {
			h.writeError(w, http.StatusNotFound, "policy not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.policyRepo.Delete(r.Context(), policyID); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Invalidate cache
	h.engine.InvalidateCache(policy.TenantID)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "policy deleted successfully",
	})
}

// HandleEvaluate handles ABAC policy evaluation.
func (h *PolicyHandler) HandleEvaluate(w http.ResponseWriter, r *http.Request) {
	var req EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "tenant context required")
		return
	}

	evalRequest := &EvaluationRequest{
		TenantID:    tenantID,
		Subject:     req.Subject,
		Resource:    req.Resource,
		Action:      req.Action,
		Environment: req.Environment,
		Context:     req.Context,
	}

	result, err := h.engine.Evaluate(r.Context(), evalRequest)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// HandleBatchEvaluate handles batch ABAC policy evaluation.
func (h *PolicyHandler) HandleBatchEvaluate(w http.ResponseWriter, r *http.Request) {
	var requests []EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(requests) == 0 {
		h.writeError(w, http.StatusBadRequest, "at least one request is required")
		return
	}

	if len(requests) > 100 {
		h.writeError(w, http.StatusBadRequest, "maximum 100 requests per batch")
		return
	}

	// Get tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "tenant context required")
		return
	}

	results := make([]*EvaluationResult, len(requests))
	for i, req := range requests {
		evalRequest := &EvaluationRequest{
			TenantID:    tenantID,
			Subject:     req.Subject,
			Resource:    req.Resource,
			Action:      req.Action,
			Environment: req.Environment,
			Context:     req.Context,
		}

		result, err := h.engine.Evaluate(r.Context(), evalRequest)
		if err != nil {
			results[i] = &EvaluationResult{
				Decision: DecisionIndeterminate,
				Reason:   err.Error(),
			}
		} else {
			results[i] = result
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    results,
	})
}

// HandleTestPolicy handles policy testing without persistence.
func (h *PolicyHandler) HandleTestPolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Policy  Policy         `json:"policy"`
		Request EvaluateRequest `json:"request"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "tenant context required")
		return
	}

	// Create a temporary in-memory store for testing
	tempStore := NewInMemoryPolicyStore()
	req.Policy.TenantID = tenantID
	req.Policy.ID = uuid.New()
	req.Policy.Enabled = true

	if err := tempStore.Create(r.Context(), &req.Policy); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create temporary engine
	tempEngine := NewEngine(tempStore, EngineConfig{
		Algorithm:   AlgorithmDenyOverrides,
		EnableTrace: true,
	})

	evalRequest := &EvaluationRequest{
		TenantID:    tenantID,
		Subject:     req.Request.Subject,
		Resource:    req.Request.Resource,
		Action:      req.Request.Action,
		Environment: req.Request.Environment,
		Context:     req.Request.Context,
	}

	result, err := tempEngine.Evaluate(r.Context(), evalRequest)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// HandleClonePolicy handles policy cloning.
func (h *PolicyHandler) HandleClonePolicy(w http.ResponseWriter, r *http.Request) {
	policyIDStr := getPathParam(r, "id")
	policyID, err := uuid.Parse(policyIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid policy ID")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Get existing policy
	policy, err := h.policyRepo.FindByID(r.Context(), policyID)
	if err != nil {
		if err == ErrPolicyNotFound {
			h.writeError(w, http.StatusNotFound, "policy not found")
			return
		}
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get user ID for audit
	var createdBy *uuid.UUID
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		createdBy = &userID
	}

	// Clone policy
	cloned := &Policy{
		ID:          uuid.New(),
		TenantID:    policy.TenantID,
		Name:        req.Name,
		Description: policy.Description + " (cloned)",
		Rules:       make([]PolicyRule, len(policy.Rules)),
		Targets:     policy.Targets,
		Priority:    policy.Priority,
		Enabled:     false, // Cloned policies start disabled
		Metadata:    policy.Metadata,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		CreatedBy:   createdBy,
		Version:     1,
	}

	// Clone rules with new IDs
	for i, rule := range policy.Rules {
		cloned.Rules[i] = rule
		cloned.Rules[i].ID = uuid.New().String()
	}

	if err := h.policyRepo.Create(r.Context(), cloned); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    cloned,
	})
}

// HandleExportPolicies handles exporting policies.
func (h *PolicyHandler) HandleExportPolicies(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "tenant context required")
		return
	}

	policies, err := h.policyRepo.FindByTenantID(r.Context(), tenantID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set headers for download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=policies.json")

	export := map[string]interface{}{
		"version":    "1.0",
		"exported_at": time.Now().UTC(),
		"tenant_id":  tenantID,
		"policies":   policies,
	}

	json.NewEncoder(w).Encode(export)
}

// HandleImportPolicies handles importing policies.
func (h *PolicyHandler) HandleImportPolicies(w http.ResponseWriter, r *http.Request) {
	// Get tenant ID from context
	tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
	if !ok {
		h.writeError(w, http.StatusBadRequest, "tenant context required")
		return
	}

	var importData struct {
		Policies []Policy `json:"policies"`
	}

	if err := json.NewDecoder(r.Body).Decode(&importData); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user ID for audit
	var createdBy *uuid.UUID
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		createdBy = &userID
	}

	imported := 0
	var errors []string

	for _, policy := range importData.Policies {
		// Assign new ID and tenant
		policy.ID = uuid.New()
		policy.TenantID = tenantID
		policy.CreatedAt = time.Now().UTC()
		policy.UpdatedAt = time.Now().UTC()
		policy.CreatedBy = createdBy
		policy.Version = 1

		// Assign new rule IDs
		for i := range policy.Rules {
			policy.Rules[i].ID = uuid.New().String()
		}

		if err := h.policyRepo.Create(r.Context(), &policy); err != nil {
			errors = append(errors, policy.Name+": "+err.Error())
		} else {
			imported++
		}
	}

	// Invalidate cache
	h.engine.InvalidateCache(tenantID)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  len(errors) == 0,
		"imported": imported,
		"errors":   errors,
	})
}

// RegisterRoutes registers ABAC policy routes.
func (h *PolicyHandler) RegisterRoutes(mux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}, prefix string) {
	mux.HandleFunc(prefix+"/policies", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleListPolicies(w, r)
		case http.MethodPost:
			h.HandleCreatePolicy(w, r)
		default:
			h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc(prefix+"/policies/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGetPolicy(w, r)
		case http.MethodPut:
			h.HandleUpdatePolicy(w, r)
		case http.MethodDelete:
			h.HandleDeletePolicy(w, r)
		default:
			h.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc(prefix+"/policies/clone/", h.HandleClonePolicy)
	mux.HandleFunc(prefix+"/evaluate", h.HandleEvaluate)
	mux.HandleFunc(prefix+"/evaluate/batch", h.HandleBatchEvaluate)
	mux.HandleFunc(prefix+"/test", h.HandleTestPolicy)
	mux.HandleFunc(prefix+"/export", h.HandleExportPolicies)
	mux.HandleFunc(prefix+"/import", h.HandleImportPolicies)
}

// Helper functions

func (h *PolicyHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *PolicyHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"message": message,
		},
	})
}

func getPathParam(r *http.Request, name string) string {
	// Try to get from context (set by router)
	if v, ok := r.Context().Value(name).(string); ok {
		return v
	}

	// Try to parse from URL path
	// This is a fallback - ideally the router should set this
	path := r.URL.Path
	// Simple extraction for /policies/{id} pattern
	parts := splitPath(path)
	for i, part := range parts {
		if part == "policies" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// DecisionChecker is a helper for checking ABAC decisions in handlers.
type DecisionChecker struct {
	engine *Engine
}

// NewDecisionChecker creates a new decision checker.
func NewDecisionChecker(engine *Engine) *DecisionChecker {
	return &DecisionChecker{engine: engine}
}

// CanPerform checks if the subject can perform an action on a resource.
func (c *DecisionChecker) CanPerform(ctx context.Context, tenantID, subjectID uuid.UUID, action, resourceType string, resourceID *uuid.UUID) (bool, error) {
	request := &EvaluationRequest{
		TenantID: tenantID,
		Subject: map[string]interface{}{
			"id": subjectID.String(),
		},
		Action: map[string]interface{}{
			"name": action,
		},
		Resource: map[string]interface{}{
			"type": resourceType,
		},
	}

	if resourceID != nil {
		request.Resource["id"] = resourceID.String()
	}

	result, err := c.engine.Evaluate(ctx, request)
	if err != nil {
		return false, err
	}

	return result.Decision == DecisionAllow, nil
}

// FilterAllowed filters a list of resource IDs to only those the subject can access.
func (c *DecisionChecker) FilterAllowed(ctx context.Context, tenantID, subjectID uuid.UUID, action, resourceType string, resourceIDs []uuid.UUID) ([]uuid.UUID, error) {
	var allowed []uuid.UUID

	for _, resourceID := range resourceIDs {
		rid := resourceID
		can, err := c.CanPerform(ctx, tenantID, subjectID, action, resourceType, &rid)
		if err != nil {
			return nil, err
		}
		if can {
			allowed = append(allowed, resourceID)
		}
	}

	return allowed, nil
}
