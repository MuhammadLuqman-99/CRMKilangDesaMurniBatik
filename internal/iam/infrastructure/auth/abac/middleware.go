// Package abac provides Attribute-Based Access Control infrastructure.
package abac

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContextKey type for context keys.
type ContextKey string

const (
	// ContextKeySubject is the context key for subject attributes.
	ContextKeySubject ContextKey = "abac_subject"
	// ContextKeyResource is the context key for resource attributes.
	ContextKeyResource ContextKey = "abac_resource"
	// ContextKeyAction is the context key for action attributes.
	ContextKeyAction ContextKey = "abac_action"
	// ContextKeyDecision is the context key for ABAC decision.
	ContextKeyDecision ContextKey = "abac_decision"
	// ContextKeyEvaluationResult is the context key for evaluation result.
	ContextKeyEvaluationResult ContextKey = "abac_evaluation_result"
)

// Middleware provides ABAC middleware for HTTP handlers.
type Middleware struct {
	engine        *Engine
	logger        *slog.Logger
	resourceMapper ResourceMapper
	actionMapper   ActionMapper
	subjectExtractor SubjectExtractor
	config        MiddlewareConfig
}

// MiddlewareConfig holds middleware configuration.
type MiddlewareConfig struct {
	// DefaultEffect is the effect when no policies apply.
	DefaultEffect Effect
	// SkipPaths are paths that bypass ABAC checks.
	SkipPaths []string
	// EnforceMode determines whether to block or just log denials.
	EnforceMode bool
	// AuditEnabled enables audit logging of decisions.
	AuditEnabled bool
	// CacheDecisions enables decision caching.
	CacheDecisions bool
	// DecisionCacheTTL is the TTL for cached decisions.
	DecisionCacheTTL time.Duration
}

// DefaultMiddlewareConfig returns default middleware configuration.
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		DefaultEffect:    EffectDeny,
		SkipPaths:        []string{"/health", "/ready", "/metrics"},
		EnforceMode:      true,
		AuditEnabled:     true,
		CacheDecisions:   false,
		DecisionCacheTTL: 30 * time.Second,
	}
}

// ResourceMapper maps HTTP requests to resource attributes.
type ResourceMapper interface {
	MapResource(r *http.Request) map[string]interface{}
}

// ActionMapper maps HTTP requests to action attributes.
type ActionMapper interface {
	MapAction(r *http.Request) map[string]interface{}
}

// SubjectExtractor extracts subject attributes from HTTP requests.
type SubjectExtractor interface {
	ExtractSubject(r *http.Request) (map[string]interface{}, error)
}

// NewMiddleware creates a new ABAC middleware.
func NewMiddleware(engine *Engine, config MiddlewareConfig, logger *slog.Logger) *Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return &Middleware{
		engine:        engine,
		logger:        logger,
		config:        config,
	}
}

// WithResourceMapper sets the resource mapper.
func (m *Middleware) WithResourceMapper(mapper ResourceMapper) *Middleware {
	m.resourceMapper = mapper
	return m
}

// WithActionMapper sets the action mapper.
func (m *Middleware) WithActionMapper(mapper ActionMapper) *Middleware {
	m.actionMapper = mapper
	return m
}

// WithSubjectExtractor sets the subject extractor.
func (m *Middleware) WithSubjectExtractor(extractor SubjectExtractor) *Middleware {
	m.subjectExtractor = extractor
	return m
}

// Enforce returns a middleware that enforces ABAC policies.
func (m *Middleware) Enforce() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be skipped
			if m.shouldSkipPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Build evaluation request
			evalRequest, err := m.buildEvaluationRequest(r)
			if err != nil {
				m.logger.Error("failed to build evaluation request",
					slog.Any("error", err),
				)
				if m.config.EnforceMode {
					m.writeError(w, http.StatusForbidden, "access denied")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Evaluate policies
			result, err := m.engine.Evaluate(r.Context(), evalRequest)
			if err != nil {
				m.logger.Error("ABAC evaluation failed",
					slog.Any("error", err),
				)
				if m.config.EnforceMode {
					m.writeError(w, http.StatusInternalServerError, "authorization error")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Store result in context
			ctx := context.WithValue(r.Context(), ContextKeyDecision, result.Decision)
			ctx = context.WithValue(ctx, ContextKeyEvaluationResult, result)

			// Log decision
			m.logDecision(r, evalRequest, result)

			// Check decision
			if result.Decision == DecisionDeny {
				if m.config.EnforceMode {
					m.writeError(w, http.StatusForbidden, "access denied")
					return
				}
			}

			// Handle not applicable based on default effect
			if result.Decision == DecisionNotApplicable {
				if m.config.DefaultEffect == EffectDeny && m.config.EnforceMode {
					m.writeError(w, http.StatusForbidden, "access denied - no applicable policy")
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission returns a middleware that requires a specific permission.
func (m *Middleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			evalRequest, err := m.buildEvaluationRequest(r)
			if err != nil {
				m.writeError(w, http.StatusForbidden, "access denied")
				return
			}

			// Override action with specific permission
			evalRequest.Action["permission"] = permission

			result, err := m.engine.Evaluate(r.Context(), evalRequest)
			if err != nil || result.Decision != DecisionAllow {
				m.writeError(w, http.StatusForbidden, "access denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns a middleware that requires a specific role.
func (m *Middleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subject, err := m.subjectExtractor.ExtractSubject(r)
			if err != nil {
				m.writeError(w, http.StatusForbidden, "access denied")
				return
			}

			roles, ok := subject["roles"].([]string)
			if !ok {
				m.writeError(w, http.StatusForbidden, "access denied")
				return
			}

			for _, r := range roles {
				if r == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			m.writeError(w, http.StatusForbidden, "access denied - role required: "+role)
		})
	}
}

// RequireAnyRole returns a middleware that requires any of the specified roles.
func (m *Middleware) RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subject, err := m.subjectExtractor.ExtractSubject(r)
			if err != nil {
				m.writeError(w, http.StatusForbidden, "access denied")
				return
			}

			userRoles, ok := subject["roles"].([]string)
			if !ok {
				m.writeError(w, http.StatusForbidden, "access denied")
				return
			}

			for _, role := range userRoles {
				if roleSet[role] {
					next.ServeHTTP(w, r)
					return
				}
			}

			m.writeError(w, http.StatusForbidden, "access denied - required role not found")
		})
	}
}

// RequireOwner returns a middleware that requires resource ownership.
func (m *Middleware) RequireOwner(resourceType string, idParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			evalRequest, err := m.buildEvaluationRequest(r)
			if err != nil {
				m.writeError(w, http.StatusForbidden, "access denied")
				return
			}

			// Set resource type and ID
			evalRequest.Resource["type"] = resourceType

			// Check if user is owner via ABAC
			evalRequest.Action["requires_ownership"] = true

			result, err := m.engine.Evaluate(r.Context(), evalRequest)
			if err != nil || result.Decision != DecisionAllow {
				m.writeError(w, http.StatusForbidden, "access denied - ownership required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// buildEvaluationRequest builds an ABAC evaluation request from HTTP request.
func (m *Middleware) buildEvaluationRequest(r *http.Request) (*EvaluationRequest, error) {
	// Extract tenant ID from context or header
	var tenantID uuid.UUID
	if tid, ok := r.Context().Value("tenant_id").(uuid.UUID); ok {
		tenantID = tid
	} else if tidStr := r.Header.Get("X-Tenant-ID"); tidStr != "" {
		var err error
		tenantID, err = uuid.Parse(tidStr)
		if err != nil {
			return nil, err
		}
	}

	// Extract subject
	var subject map[string]interface{}
	if m.subjectExtractor != nil {
		var err error
		subject, err = m.subjectExtractor.ExtractSubject(r)
		if err != nil {
			return nil, err
		}
	} else {
		subject = m.defaultSubjectExtract(r)
	}

	// Map resource
	var resource map[string]interface{}
	if m.resourceMapper != nil {
		resource = m.resourceMapper.MapResource(r)
	} else {
		resource = m.defaultResourceMap(r)
	}

	// Map action
	var action map[string]interface{}
	if m.actionMapper != nil {
		action = m.actionMapper.MapAction(r)
	} else {
		action = m.defaultActionMap(r)
	}

	// Build environment attributes
	environment := map[string]interface{}{
		"time":        time.Now().Format("15:04"),
		"date":        time.Now().Format("2006-01-02"),
		"day_of_week": int(time.Now().Weekday()),
	}

	// Build context attributes
	ctx := map[string]interface{}{
		"ip_address":  getClientIP(r),
		"user_agent":  r.UserAgent(),
		"request_id":  r.Header.Get("X-Request-ID"),
		"http_method": r.Method,
		"path":        r.URL.Path,
	}

	return &EvaluationRequest{
		TenantID:    tenantID,
		Subject:     subject,
		Resource:    resource,
		Action:      action,
		Environment: environment,
		Context:     ctx,
	}, nil
}

// defaultSubjectExtract extracts subject from JWT claims in context.
func (m *Middleware) defaultSubjectExtract(r *http.Request) map[string]interface{} {
	subject := make(map[string]interface{})

	// Extract from context (set by auth middleware)
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		subject["id"] = userID.String()
	}

	if email, ok := r.Context().Value("email").(string); ok {
		subject["email"] = email
	}

	if roles, ok := r.Context().Value("roles").([]string); ok {
		subject["roles"] = roles
	}

	if permissions, ok := r.Context().Value("permissions").([]string); ok {
		subject["permissions"] = permissions
	}

	return subject
}

// defaultResourceMap maps HTTP request to resource.
func (m *Middleware) defaultResourceMap(r *http.Request) map[string]interface{} {
	resource := make(map[string]interface{})

	// Extract resource type from path
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) >= 1 {
		// Skip "api" and version prefix
		startIdx := 0
		for i, part := range parts {
			if part != "api" && !strings.HasPrefix(part, "v") {
				startIdx = i
				break
			}
		}

		if startIdx < len(parts) {
			resource["type"] = parts[startIdx]
		}

		// Extract resource ID if present
		if startIdx+1 < len(parts) {
			if id, err := uuid.Parse(parts[startIdx+1]); err == nil {
				resource["id"] = id.String()
			}
		}
	}

	resource["path"] = r.URL.Path

	return resource
}

// defaultActionMap maps HTTP method to action.
func (m *Middleware) defaultActionMap(r *http.Request) map[string]interface{} {
	action := make(map[string]interface{})

	// Map HTTP methods to standard actions
	methodActions := map[string]string{
		"GET":     "read",
		"HEAD":    "read",
		"OPTIONS": "read",
		"POST":    "create",
		"PUT":     "update",
		"PATCH":   "update",
		"DELETE":  "delete",
	}

	action["name"] = methodActions[r.Method]
	action["method"] = r.Method

	return action
}

// shouldSkipPath checks if the path should bypass ABAC.
func (m *Middleware) shouldSkipPath(path string) bool {
	for _, skip := range m.config.SkipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// logDecision logs the ABAC decision.
func (m *Middleware) logDecision(r *http.Request, request *EvaluationRequest, result *EvaluationResult) {
	if !m.config.AuditEnabled {
		return
	}

	attrs := []any{
		slog.String("decision", string(result.Decision)),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("client_ip", getClientIP(r)),
		slog.Duration("evaluation_time", result.EvaluationTime),
	}

	if result.PolicyID != nil {
		attrs = append(attrs, slog.String("policy_id", result.PolicyID.String()))
	}

	if result.RuleID != nil {
		attrs = append(attrs, slog.String("rule_id", *result.RuleID))
	}

	if subjectID, ok := request.Subject["id"].(string); ok {
		attrs = append(attrs, slog.String("subject_id", subjectID))
	}

	if result.Decision == DecisionDeny {
		m.logger.Warn("ABAC access denied", attrs...)
	} else {
		m.logger.Debug("ABAC decision", attrs...)
	}
}

// writeError writes an error response.
func (m *Middleware) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"message": message,
		},
	})
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// DefaultResourceMapper is a default implementation of ResourceMapper.
type DefaultResourceMapper struct {
	pathPrefix string
}

// NewDefaultResourceMapper creates a new default resource mapper.
func NewDefaultResourceMapper(pathPrefix string) *DefaultResourceMapper {
	return &DefaultResourceMapper{pathPrefix: pathPrefix}
}

// MapResource maps the HTTP request to resource attributes.
func (m *DefaultResourceMapper) MapResource(r *http.Request) map[string]interface{} {
	resource := make(map[string]interface{})

	path := r.URL.Path
	if m.pathPrefix != "" {
		path = strings.TrimPrefix(path, m.pathPrefix)
	}
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) >= 1 && parts[0] != "" {
		resource["type"] = parts[0]
	}

	if len(parts) >= 2 {
		if id, err := uuid.Parse(parts[1]); err == nil {
			resource["id"] = id.String()
		}
	}

	// Add sub-resource if present
	if len(parts) >= 3 {
		resource["sub_resource"] = parts[2]
	}

	resource["path"] = r.URL.Path

	return resource
}

// DefaultActionMapper is a default implementation of ActionMapper.
type DefaultActionMapper struct {
	customActions map[string]map[string]string // path -> method -> action
}

// NewDefaultActionMapper creates a new default action mapper.
func NewDefaultActionMapper() *DefaultActionMapper {
	return &DefaultActionMapper{
		customActions: make(map[string]map[string]string),
	}
}

// AddCustomAction adds a custom action mapping.
func (m *DefaultActionMapper) AddCustomAction(pathPattern, method, action string) {
	if m.customActions[pathPattern] == nil {
		m.customActions[pathPattern] = make(map[string]string)
	}
	m.customActions[pathPattern][method] = action
}

// MapAction maps the HTTP request to action attributes.
func (m *DefaultActionMapper) MapAction(r *http.Request) map[string]interface{} {
	action := make(map[string]interface{})

	// Check custom actions first
	for pathPattern, methods := range m.customActions {
		if matched, _ := matchPath(r.URL.Path, pathPattern); matched {
			if actionName, ok := methods[r.Method]; ok {
				action["name"] = actionName
				action["method"] = r.Method
				return action
			}
		}
	}

	// Default method to action mapping
	methodActions := map[string]string{
		"GET":     "read",
		"HEAD":    "read",
		"OPTIONS": "read",
		"POST":    "create",
		"PUT":     "update",
		"PATCH":   "update",
		"DELETE":  "delete",
	}

	action["name"] = methodActions[r.Method]
	action["method"] = r.Method

	return action
}

func matchPath(path, pattern string) (bool, map[string]string) {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(pathParts) != len(patternParts) {
		return false, nil
	}

	params := make(map[string]string)
	for i, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, ":") {
			params[patternPart[1:]] = pathParts[i]
		} else if patternPart != pathParts[i] {
			return false, nil
		}
	}

	return true, params
}

// JWTSubjectExtractor extracts subject from JWT claims.
type JWTSubjectExtractor struct {
	claimsKey string
}

// NewJWTSubjectExtractor creates a new JWT subject extractor.
func NewJWTSubjectExtractor(claimsKey string) *JWTSubjectExtractor {
	if claimsKey == "" {
		claimsKey = "claims"
	}
	return &JWTSubjectExtractor{claimsKey: claimsKey}
}

// ExtractSubject extracts subject attributes from JWT claims in context.
func (e *JWTSubjectExtractor) ExtractSubject(r *http.Request) (map[string]interface{}, error) {
	subject := make(map[string]interface{})

	// Get claims from context
	claims, ok := r.Context().Value(e.claimsKey).(map[string]interface{})
	if !ok {
		// Try to get individual values
		if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
			subject["id"] = userID.String()
		} else if userIDStr, ok := r.Context().Value("user_id").(string); ok {
			subject["id"] = userIDStr
		}

		if email, ok := r.Context().Value("email").(string); ok {
			subject["email"] = email
		}

		if roles, ok := r.Context().Value("roles").([]string); ok {
			subject["roles"] = roles
		}

		if tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID); ok {
			subject["tenant_id"] = tenantID.String()
		}

		return subject, nil
	}

	// Extract from claims
	if sub, ok := claims["sub"].(string); ok {
		subject["id"] = sub
	}

	if email, ok := claims["email"].(string); ok {
		subject["email"] = email
	}

	if roles, ok := claims["roles"].([]interface{}); ok {
		roleStrs := make([]string, len(roles))
		for i, r := range roles {
			if rs, ok := r.(string); ok {
				roleStrs[i] = rs
			}
		}
		subject["roles"] = roleStrs
	}

	if permissions, ok := claims["permissions"].([]interface{}); ok {
		permStrs := make([]string, len(permissions))
		for i, p := range permissions {
			if ps, ok := p.(string); ok {
				permStrs[i] = ps
			}
		}
		subject["permissions"] = permStrs
	}

	if tenantID, ok := claims["tenant_id"].(string); ok {
		subject["tenant_id"] = tenantID
	}

	return subject, nil
}

// GetDecision retrieves the ABAC decision from context.
func GetDecision(ctx context.Context) Decision {
	if decision, ok := ctx.Value(ContextKeyDecision).(Decision); ok {
		return decision
	}
	return DecisionNotApplicable
}

// GetEvaluationResult retrieves the full evaluation result from context.
func GetEvaluationResult(ctx context.Context) *EvaluationResult {
	if result, ok := ctx.Value(ContextKeyEvaluationResult).(*EvaluationResult); ok {
		return result
	}
	return nil
}
