// Package abac provides Attribute-Based Access Control infrastructure.
package abac

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Decision represents an ABAC evaluation decision.
type Decision string

const (
	DecisionAllow        Decision = "allow"
	DecisionDeny         Decision = "deny"
	DecisionNotApplicable Decision = "not_applicable"
	DecisionIndeterminate Decision = "indeterminate"
)

// EvaluationResult contains the result of policy evaluation.
type EvaluationResult struct {
	Decision        Decision          `json:"decision"`
	PolicyID        *uuid.UUID        `json:"policy_id,omitempty"`
	RuleID          *string           `json:"rule_id,omitempty"`
	Reason          string            `json:"reason,omitempty"`
	Obligations     []Obligation      `json:"obligations,omitempty"`
	Advice          []Advice          `json:"advice,omitempty"`
	EvaluatedAt     time.Time         `json:"evaluated_at"`
	EvaluationTime  time.Duration     `json:"evaluation_time"`
	MatchedPolicies []uuid.UUID       `json:"matched_policies,omitempty"`
	Trace           *EvaluationTrace  `json:"trace,omitempty"`
}

// Obligation represents an action that must be performed.
type Obligation struct {
	ID         string                 `json:"id"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Advice represents a non-mandatory recommendation.
type Advice struct {
	ID         string                 `json:"id"`
	Message    string                 `json:"message"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// EvaluationTrace provides detailed evaluation information for debugging.
type EvaluationTrace struct {
	Request          *EvaluationRequest      `json:"request"`
	ApplicablePolicies []PolicyTrace         `json:"applicable_policies"`
	CombinedDecision Decision               `json:"combined_decision"`
	Algorithm        CombiningAlgorithm      `json:"algorithm"`
}

// PolicyTrace traces evaluation of a single policy.
type PolicyTrace struct {
	PolicyID   uuid.UUID    `json:"policy_id"`
	PolicyName string       `json:"policy_name"`
	Decision   Decision     `json:"decision"`
	Rules      []RuleTrace  `json:"rules,omitempty"`
}

// RuleTrace traces evaluation of a single rule.
type RuleTrace struct {
	RuleID     string    `json:"rule_id"`
	Decision   Decision  `json:"decision"`
	Matched    bool      `json:"matched"`
	Conditions []string  `json:"conditions,omitempty"`
}

// EvaluationRequest represents an ABAC evaluation request.
type EvaluationRequest struct {
	TenantID    uuid.UUID              `json:"tenant_id"`
	Subject     map[string]interface{} `json:"subject"`
	Resource    map[string]interface{} `json:"resource"`
	Action      map[string]interface{} `json:"action"`
	Environment map[string]interface{} `json:"environment"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// Engine is the ABAC policy evaluation engine.
type Engine struct {
	policyRepo    PolicyRepository
	attributeProviders []AttributeProvider
	algorithm     CombiningAlgorithm
	logger        *slog.Logger
	cache         *policyCache
	enableTrace   bool
	mu            sync.RWMutex
}

// EngineConfig holds configuration for the ABAC engine.
type EngineConfig struct {
	Algorithm    CombiningAlgorithm
	Logger       *slog.Logger
	CacheTTL     time.Duration
	EnableTrace  bool
}

// DefaultEngineConfig returns default engine configuration.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		Algorithm:   AlgorithmDenyOverrides,
		Logger:      slog.Default(),
		CacheTTL:    5 * time.Minute,
		EnableTrace: false,
	}
}

// policyCache caches policies for faster evaluation.
type policyCache struct {
	policies  map[uuid.UUID][]*Policy
	expiry    map[uuid.UUID]time.Time
	ttl       time.Duration
	mu        sync.RWMutex
}

func newPolicyCache(ttl time.Duration) *policyCache {
	return &policyCache{
		policies: make(map[uuid.UUID][]*Policy),
		expiry:   make(map[uuid.UUID]time.Time),
		ttl:      ttl,
	}
}

func (c *policyCache) get(tenantID uuid.UUID) ([]*Policy, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expiry, exists := c.expiry[tenantID]
	if !exists || time.Now().After(expiry) {
		return nil, false
	}

	return c.policies[tenantID], true
}

func (c *policyCache) set(tenantID uuid.UUID, policies []*Policy) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.policies[tenantID] = policies
	c.expiry[tenantID] = time.Now().Add(c.ttl)
}

func (c *policyCache) invalidate(tenantID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.policies, tenantID)
	delete(c.expiry, tenantID)
}

// NewEngine creates a new ABAC evaluation engine.
func NewEngine(policyRepo PolicyRepository, config EngineConfig) *Engine {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &Engine{
		policyRepo:         policyRepo,
		attributeProviders: make([]AttributeProvider, 0),
		algorithm:          config.Algorithm,
		logger:             logger,
		cache:              newPolicyCache(config.CacheTTL),
		enableTrace:        config.EnableTrace,
	}
}

// RegisterAttributeProvider registers an attribute provider.
func (e *Engine) RegisterAttributeProvider(provider AttributeProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.attributeProviders = append(e.attributeProviders, provider)
}

// Evaluate evaluates an ABAC request against policies.
func (e *Engine) Evaluate(ctx context.Context, request *EvaluationRequest) (*EvaluationResult, error) {
	startTime := time.Now()

	result := &EvaluationResult{
		EvaluatedAt: startTime,
	}

	// Enrich attributes from providers
	enrichedRequest, err := e.enrichAttributes(ctx, request)
	if err != nil {
		e.logger.Error("failed to enrich attributes",
			slog.Any("error", err),
		)
		// Continue with original request
		enrichedRequest = request
	}

	// Get applicable policies
	policies, err := e.getApplicablePolicies(ctx, enrichedRequest)
	if err != nil {
		result.Decision = DecisionIndeterminate
		result.Reason = fmt.Sprintf("failed to get policies: %v", err)
		result.EvaluationTime = time.Since(startTime)
		return result, nil
	}

	if len(policies) == 0 {
		result.Decision = DecisionNotApplicable
		result.Reason = "no applicable policies found"
		result.EvaluationTime = time.Since(startTime)
		return result, nil
	}

	// Sort policies by priority (higher first)
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority > policies[j].Priority
	})

	// Evaluate policies
	var trace *EvaluationTrace
	if e.enableTrace {
		trace = &EvaluationTrace{
			Request:            enrichedRequest,
			ApplicablePolicies: make([]PolicyTrace, 0, len(policies)),
			Algorithm:          e.algorithm,
		}
	}

	decisions := make([]policyDecision, 0, len(policies))
	for _, policy := range policies {
		decision, policyTrace := e.evaluatePolicy(ctx, enrichedRequest, policy)
		decisions = append(decisions, policyDecision{
			policyID: policy.ID,
			decision: decision,
			priority: policy.Priority,
		})
		result.MatchedPolicies = append(result.MatchedPolicies, policy.ID)

		if e.enableTrace && policyTrace != nil {
			trace.ApplicablePolicies = append(trace.ApplicablePolicies, *policyTrace)
		}
	}

	// Combine decisions using algorithm
	finalDecision, matchedPolicy, matchedRule := e.combineDecisions(decisions)
	result.Decision = finalDecision
	result.PolicyID = matchedPolicy
	result.RuleID = matchedRule
	result.EvaluationTime = time.Since(startTime)

	if e.enableTrace {
		trace.CombinedDecision = finalDecision
		result.Trace = trace
	}

	e.logger.Debug("ABAC evaluation completed",
		slog.String("tenant_id", request.TenantID.String()),
		slog.String("decision", string(result.Decision)),
		slog.Duration("evaluation_time", result.EvaluationTime),
		slog.Int("policies_evaluated", len(policies)),
	)

	return result, nil
}

// enrichAttributes enriches the request with additional attributes from providers.
func (e *Engine) enrichAttributes(ctx context.Context, request *EvaluationRequest) (*EvaluationRequest, error) {
	e.mu.RLock()
	providers := e.attributeProviders
	e.mu.RUnlock()

	enriched := &EvaluationRequest{
		TenantID:    request.TenantID,
		Subject:     copyMap(request.Subject),
		Resource:    copyMap(request.Resource),
		Action:      copyMap(request.Action),
		Environment: copyMap(request.Environment),
		Context:     copyMap(request.Context),
	}

	for _, provider := range providers {
		attrs, err := provider.GetAttributes(ctx, request)
		if err != nil {
			e.logger.Warn("attribute provider failed",
				slog.String("provider", provider.Name()),
				slog.Any("error", err),
			)
			continue
		}

		for source, values := range attrs {
			target := enriched.getAttributeMap(source)
			for k, v := range values {
				target[k] = v
			}
		}
	}

	return enriched, nil
}

func (r *EvaluationRequest) getAttributeMap(source AttributeSource) map[string]interface{} {
	switch source {
	case SourceSubject:
		return r.Subject
	case SourceResource:
		return r.Resource
	case SourceAction:
		return r.Action
	case SourceEnvironment:
		return r.Environment
	case SourceContext:
		return r.Context
	default:
		return nil
	}
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return make(map[string]interface{})
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// getApplicablePolicies retrieves policies that may apply to the request.
func (e *Engine) getApplicablePolicies(ctx context.Context, request *EvaluationRequest) ([]*Policy, error) {
	// Check cache first
	if policies, found := e.cache.get(request.TenantID); found {
		return e.filterApplicablePolicies(policies, request), nil
	}

	// Load from repository
	policies, err := e.policyRepo.FindByTenantID(ctx, request.TenantID)
	if err != nil {
		return nil, err
	}

	// Cache policies
	e.cache.set(request.TenantID, policies)

	return e.filterApplicablePolicies(policies, request), nil
}

// filterApplicablePolicies filters policies that apply to the request.
func (e *Engine) filterApplicablePolicies(policies []*Policy, request *EvaluationRequest) []*Policy {
	resource := getString(request.Resource, "type")
	action := getString(request.Action, "name")

	var applicable []*Policy
	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		// Check target matching
		if e.policyTargetMatches(policy, request) {
			// Check if any rule might apply
			for _, rule := range policy.Rules {
				if matchesAny(resource, rule.Resources) && matchesAny(action, rule.Actions) {
					applicable = append(applicable, policy)
					break
				}
			}
		}
	}

	return applicable
}

func (e *Engine) policyTargetMatches(policy *Policy, request *EvaluationRequest) bool {
	// If no targets specified, policy applies globally
	if len(policy.Targets.Subjects) == 0 &&
		len(policy.Targets.Resources) == 0 &&
		len(policy.Targets.Actions) == 0 {
		return true
	}

	// Check subject targets
	if len(policy.Targets.Subjects) > 0 {
		subjectID := getString(request.Subject, "id")
		role := getString(request.Subject, "role")
		if !matchesAny(subjectID, policy.Targets.Subjects) &&
			!matchesAny(role, policy.Targets.Subjects) {
			return false
		}
	}

	// Check resource targets
	if len(policy.Targets.Resources) > 0 {
		resourceType := getString(request.Resource, "type")
		if !matchesAny(resourceType, policy.Targets.Resources) {
			return false
		}
	}

	// Check action targets
	if len(policy.Targets.Actions) > 0 {
		actionName := getString(request.Action, "name")
		if !matchesAny(actionName, policy.Targets.Actions) {
			return false
		}
	}

	return true
}

type policyDecision struct {
	policyID uuid.UUID
	decision Decision
	priority int
	ruleID   string
}

// evaluatePolicy evaluates a single policy against the request.
func (e *Engine) evaluatePolicy(ctx context.Context, request *EvaluationRequest, policy *Policy) (Decision, *PolicyTrace) {
	var trace *PolicyTrace
	if e.enableTrace {
		trace = &PolicyTrace{
			PolicyID:   policy.ID,
			PolicyName: policy.Name,
			Rules:      make([]RuleTrace, 0, len(policy.Rules)),
		}
	}

	resource := getString(request.Resource, "type")
	action := getString(request.Action, "name")

	// Sort rules by priority
	rules := make([]PolicyRule, len(policy.Rules))
	copy(rules, policy.Rules)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	var decisions []Decision
	for _, rule := range rules {
		// Check if rule applies to this resource/action
		if !matchesAny(resource, rule.Resources) || !matchesAny(action, rule.Actions) {
			continue
		}

		// Evaluate conditions
		matched, ruleTrace := e.evaluateRuleConditions(ctx, request, rule)

		if e.enableTrace && ruleTrace != nil {
			trace.Rules = append(trace.Rules, *ruleTrace)
		}

		if matched {
			if rule.Effect == EffectDeny {
				decisions = append(decisions, DecisionDeny)
			} else {
				decisions = append(decisions, DecisionAllow)
			}
		}
	}

	// Determine policy decision
	var decision Decision
	if len(decisions) == 0 {
		decision = DecisionNotApplicable
	} else {
		// Use deny overrides for rule combining within a policy
		hasDeny := false
		hasAllow := false
		for _, d := range decisions {
			if d == DecisionDeny {
				hasDeny = true
			} else if d == DecisionAllow {
				hasAllow = true
			}
		}

		if hasDeny {
			decision = DecisionDeny
		} else if hasAllow {
			decision = DecisionAllow
		} else {
			decision = DecisionNotApplicable
		}
	}

	if trace != nil {
		trace.Decision = decision
	}

	return decision, trace
}

// evaluateRuleConditions evaluates all conditions in a rule.
func (e *Engine) evaluateRuleConditions(ctx context.Context, request *EvaluationRequest, rule PolicyRule) (bool, *RuleTrace) {
	var trace *RuleTrace
	if e.enableTrace {
		trace = &RuleTrace{
			RuleID:     rule.ID,
			Conditions: make([]string, 0),
		}
	}

	// No conditions means rule always applies
	if rule.Conditions == nil {
		if trace != nil {
			trace.Matched = true
			trace.Decision = DecisionAllow
		}
		return true, trace
	}

	matched := e.evaluateConditionGroup(request, rule.Conditions)

	if trace != nil {
		trace.Matched = matched
		if matched {
			if rule.Effect == EffectDeny {
				trace.Decision = DecisionDeny
			} else {
				trace.Decision = DecisionAllow
			}
		} else {
			trace.Decision = DecisionNotApplicable
		}
	}

	return matched, trace
}

// evaluateConditionGroup evaluates a group of conditions.
func (e *Engine) evaluateConditionGroup(request *EvaluationRequest, group *ConditionGroup) bool {
	switch group.Operator {
	case LogicalAnd:
		// All conditions must be true
		for _, condition := range group.Conditions {
			if !e.evaluateCondition(request, condition) {
				return false
			}
		}
		for _, subGroup := range group.Groups {
			if !e.evaluateConditionGroup(request, &subGroup) {
				return false
			}
		}
		return true

	case LogicalOr:
		// Any condition must be true
		for _, condition := range group.Conditions {
			if e.evaluateCondition(request, condition) {
				return true
			}
		}
		for _, subGroup := range group.Groups {
			if e.evaluateConditionGroup(request, &subGroup) {
				return true
			}
		}
		return false

	case LogicalNot:
		// Negate the first condition/group
		if len(group.Conditions) > 0 {
			return !e.evaluateCondition(request, group.Conditions[0])
		}
		if len(group.Groups) > 0 {
			return !e.evaluateConditionGroup(request, &group.Groups[0])
		}
		return true

	default:
		return false
	}
}

// evaluateCondition evaluates a single condition.
func (e *Engine) evaluateCondition(request *EvaluationRequest, condition Condition) bool {
	attrValue := e.getAttributeValue(request, condition.Attribute)
	expectedValue := condition.Value

	switch condition.Operator {
	case OpEquals:
		return compareEquals(attrValue, expectedValue)

	case OpNotEquals:
		return !compareEquals(attrValue, expectedValue)

	case OpContains:
		return compareContains(attrValue, expectedValue)

	case OpNotContains:
		return !compareContains(attrValue, expectedValue)

	case OpStartsWith:
		return compareStartsWith(attrValue, expectedValue)

	case OpEndsWith:
		return compareEndsWith(attrValue, expectedValue)

	case OpMatches:
		return compareMatches(attrValue, expectedValue)

	case OpIn:
		return compareIn(attrValue, expectedValue)

	case OpNotIn:
		return !compareIn(attrValue, expectedValue)

	case OpGreaterThan:
		return compareGreaterThan(attrValue, expectedValue)

	case OpGreaterOrEqual:
		return compareGreaterOrEqual(attrValue, expectedValue)

	case OpLessThan:
		return compareLessThan(attrValue, expectedValue)

	case OpLessOrEqual:
		return compareLessOrEqual(attrValue, expectedValue)

	case OpBetween:
		return compareBetween(attrValue, expectedValue)

	case OpExists:
		return attrValue != nil

	case OpNotExists:
		return attrValue == nil

	case OpIsEmpty:
		return isEmpty(attrValue)

	case OpIsNotEmpty:
		return !isEmpty(attrValue)

	case OpIpInCIDR:
		return compareIPInCIDR(attrValue, expectedValue)

	case OpTimeInRange:
		return compareTimeInRange(attrValue, expectedValue)

	case OpDayOfWeekIn:
		return compareDayOfWeekIn(attrValue, expectedValue)

	default:
		return false
	}
}

// getAttributeValue retrieves an attribute value from the request.
func (e *Engine) getAttributeValue(request *EvaluationRequest, ref AttributeReference) interface{} {
	var source map[string]interface{}

	switch ref.Source {
	case SourceSubject:
		source = request.Subject
	case SourceResource:
		source = request.Resource
	case SourceAction:
		source = request.Action
	case SourceEnvironment:
		source = request.Environment
	case SourceContext:
		source = request.Context
	default:
		return nil
	}

	if source == nil {
		return nil
	}

	// Support nested keys with dot notation
	keys := strings.Split(ref.Key, ".")
	var current interface{} = source

	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[key]
		default:
			return nil
		}
	}

	return current
}

// combineDecisions combines multiple policy decisions using the configured algorithm.
func (e *Engine) combineDecisions(decisions []policyDecision) (Decision, *uuid.UUID, *string) {
	if len(decisions) == 0 {
		return DecisionNotApplicable, nil, nil
	}

	switch e.algorithm {
	case AlgorithmDenyOverrides:
		// If any decision is deny, result is deny
		for _, d := range decisions {
			if d.decision == DecisionDeny {
				return DecisionDeny, &d.policyID, &d.ruleID
			}
		}
		// If any decision is allow, result is allow
		for _, d := range decisions {
			if d.decision == DecisionAllow {
				return DecisionAllow, &d.policyID, &d.ruleID
			}
		}
		return DecisionNotApplicable, nil, nil

	case AlgorithmPermitOverrides:
		// If any decision is allow, result is allow
		for _, d := range decisions {
			if d.decision == DecisionAllow {
				return DecisionAllow, &d.policyID, &d.ruleID
			}
		}
		// If any decision is deny, result is deny
		for _, d := range decisions {
			if d.decision == DecisionDeny {
				return DecisionDeny, &d.policyID, &d.ruleID
			}
		}
		return DecisionNotApplicable, nil, nil

	case AlgorithmFirstApplicable:
		// Return first applicable decision
		for _, d := range decisions {
			if d.decision == DecisionAllow || d.decision == DecisionDeny {
				return d.decision, &d.policyID, &d.ruleID
			}
		}
		return DecisionNotApplicable, nil, nil

	case AlgorithmHighestPriority:
		// Sort by priority and return first applicable
		sort.Slice(decisions, func(i, j int) bool {
			return decisions[i].priority > decisions[j].priority
		})
		for _, d := range decisions {
			if d.decision == DecisionAllow || d.decision == DecisionDeny {
				return d.decision, &d.policyID, &d.ruleID
			}
		}
		return DecisionNotApplicable, nil, nil

	case AlgorithmOnlyOneApplicable:
		// Exactly one policy must be applicable
		applicableCount := 0
		var applicable *policyDecision
		for i := range decisions {
			if decisions[i].decision == DecisionAllow || decisions[i].decision == DecisionDeny {
				applicableCount++
				applicable = &decisions[i]
			}
		}
		if applicableCount == 1 {
			return applicable.decision, &applicable.policyID, &applicable.ruleID
		}
		return DecisionIndeterminate, nil, nil

	default:
		return DecisionIndeterminate, nil, nil
	}
}

// InvalidateCache invalidates the policy cache for a tenant.
func (e *Engine) InvalidateCache(tenantID uuid.UUID) {
	e.cache.invalidate(tenantID)
}

// Comparison helper functions

func compareEquals(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func compareContains(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Contains(aStr, bStr)
}

func compareStartsWith(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.HasPrefix(aStr, bStr)
}

func compareEndsWith(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.HasSuffix(aStr, bStr)
}

func compareMatches(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)
	pattern := fmt.Sprintf("%v", b)
	matched, _ := regexp.MatchString(pattern, aStr)
	return matched
}

func compareIn(a, b interface{}) bool {
	aStr := fmt.Sprintf("%v", a)

	switch v := b.(type) {
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == aStr {
				return true
			}
		}
	case []string:
		for _, item := range v {
			if item == aStr {
				return true
			}
		}
	}
	return false
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	}
	return 0, false
}

func compareGreaterThan(a, b interface{}) bool {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if aOk && bOk {
		return aNum > bNum
	}
	return fmt.Sprintf("%v", a) > fmt.Sprintf("%v", b)
}

func compareGreaterOrEqual(a, b interface{}) bool {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if aOk && bOk {
		return aNum >= bNum
	}
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return aStr >= bStr
}

func compareLessThan(a, b interface{}) bool {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if aOk && bOk {
		return aNum < bNum
	}
	return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
}

func compareLessOrEqual(a, b interface{}) bool {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if aOk && bOk {
		return aNum <= bNum
	}
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return aStr <= bStr
}

func compareBetween(a, b interface{}) bool {
	aNum, aOk := toFloat64(a)
	if !aOk {
		return false
	}

	switch v := b.(type) {
	case []interface{}:
		if len(v) != 2 {
			return false
		}
		min, minOk := toFloat64(v[0])
		max, maxOk := toFloat64(v[1])
		if minOk && maxOk {
			return aNum >= min && aNum <= max
		}
	case []float64:
		if len(v) != 2 {
			return false
		}
		return aNum >= v[0] && aNum <= v[1]
	}
	return false
}

func isEmpty(a interface{}) bool {
	if a == nil {
		return true
	}
	switch v := a.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	}
	return false
}

func compareIPInCIDR(a, b interface{}) bool {
	ipStr, ok := a.(string)
	if !ok {
		return false
	}

	cidrStr, ok := b.(string)
	if !ok {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, cidr, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return false
	}

	return cidr.Contains(ip)
}

func compareTimeInRange(a, b interface{}) bool {
	// Expected format: "HH:MM-HH:MM"
	timeStr, ok := a.(string)
	if !ok {
		// Use current time
		timeStr = time.Now().Format("15:04")
	}

	rangeStr, ok := b.(string)
	if !ok {
		return false
	}

	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return false
	}

	current, err := time.Parse("15:04", timeStr)
	if err != nil {
		return false
	}

	start, err := time.Parse("15:04", parts[0])
	if err != nil {
		return false
	}

	end, err := time.Parse("15:04", parts[1])
	if err != nil {
		return false
	}

	currentMinutes := current.Hour()*60 + current.Minute()
	startMinutes := start.Hour()*60 + start.Minute()
	endMinutes := end.Hour()*60 + end.Minute()

	// Handle overnight range (e.g., 22:00-06:00)
	if startMinutes > endMinutes {
		return currentMinutes >= startMinutes || currentMinutes <= endMinutes
	}

	return currentMinutes >= startMinutes && currentMinutes <= endMinutes
}

func compareDayOfWeekIn(a, b interface{}) bool {
	var dayNum int

	switch v := a.(type) {
	case string:
		dayNum = parseDayOfWeek(v)
	case int:
		dayNum = v
	case float64:
		dayNum = int(v)
	default:
		// Use current day
		dayNum = int(time.Now().Weekday())
	}

	switch v := b.(type) {
	case []interface{}:
		for _, day := range v {
			var targetDay int
			switch d := day.(type) {
			case string:
				targetDay = parseDayOfWeek(d)
			case int:
				targetDay = d
			case float64:
				targetDay = int(d)
			}
			if dayNum == targetDay {
				return true
			}
		}
	case []int:
		for _, day := range v {
			if dayNum == day {
				return true
			}
		}
	case []string:
		for _, day := range v {
			if dayNum == parseDayOfWeek(day) {
				return true
			}
		}
	}

	return false
}

func parseDayOfWeek(day string) int {
	day = strings.ToLower(day)
	days := map[string]int{
		"sunday":    0,
		"monday":    1,
		"tuesday":   2,
		"wednesday": 3,
		"thursday":  4,
		"friday":    5,
		"saturday":  6,
		"sun":       0,
		"mon":       1,
		"tue":       2,
		"wed":       3,
		"thu":       4,
		"fri":       5,
		"sat":       6,
	}
	if num, ok := days[day]; ok {
		return num
	}
	return -1
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
