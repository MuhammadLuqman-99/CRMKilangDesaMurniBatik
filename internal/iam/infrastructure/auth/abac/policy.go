// Package abac provides Attribute-Based Access Control infrastructure.
package abac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Common errors.
var (
	ErrPolicyNotFound      = errors.New("policy not found")
	ErrInvalidPolicy       = errors.New("invalid policy")
	ErrInvalidCondition    = errors.New("invalid condition")
	ErrEvaluationFailed    = errors.New("policy evaluation failed")
	ErrAttributeNotFound   = errors.New("attribute not found")
	ErrInvalidAttributeType = errors.New("invalid attribute type")
)

// Effect represents the policy effect.
type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

// ConditionOperator represents comparison operators.
type ConditionOperator string

const (
	OpEquals          ConditionOperator = "equals"
	OpNotEquals       ConditionOperator = "not_equals"
	OpContains        ConditionOperator = "contains"
	OpNotContains     ConditionOperator = "not_contains"
	OpStartsWith      ConditionOperator = "starts_with"
	OpEndsWith        ConditionOperator = "ends_with"
	OpMatches         ConditionOperator = "matches" // regex
	OpIn              ConditionOperator = "in"
	OpNotIn           ConditionOperator = "not_in"
	OpGreaterThan     ConditionOperator = "greater_than"
	OpGreaterOrEqual  ConditionOperator = "greater_or_equal"
	OpLessThan        ConditionOperator = "less_than"
	OpLessOrEqual     ConditionOperator = "less_or_equal"
	OpBetween         ConditionOperator = "between"
	OpExists          ConditionOperator = "exists"
	OpNotExists       ConditionOperator = "not_exists"
	OpIsEmpty         ConditionOperator = "is_empty"
	OpIsNotEmpty      ConditionOperator = "is_not_empty"
	OpIpInCIDR        ConditionOperator = "ip_in_cidr"
	OpTimeInRange     ConditionOperator = "time_in_range"
	OpDayOfWeekIn     ConditionOperator = "day_of_week_in"
)

// LogicalOperator for combining conditions.
type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "and"
	LogicalOr  LogicalOperator = "or"
	LogicalNot LogicalOperator = "not"
)

// AttributeSource indicates where an attribute comes from.
type AttributeSource string

const (
	SourceSubject     AttributeSource = "subject"     // User attributes
	SourceResource    AttributeSource = "resource"    // Resource attributes
	SourceAction      AttributeSource = "action"      // Action attributes
	SourceEnvironment AttributeSource = "environment" // Environmental attributes
	SourceContext     AttributeSource = "context"     // Request context attributes
)

// AttributeReference references an attribute for evaluation.
type AttributeReference struct {
	Source AttributeSource `json:"source"`
	Key    string          `json:"key"`
}

// Condition represents a single condition in a policy rule.
type Condition struct {
	Attribute AttributeReference `json:"attribute"`
	Operator  ConditionOperator  `json:"operator"`
	Value     interface{}        `json:"value"`
}

// ConditionGroup groups conditions with a logical operator.
type ConditionGroup struct {
	Operator   LogicalOperator  `json:"operator"`
	Conditions []Condition      `json:"conditions,omitempty"`
	Groups     []ConditionGroup `json:"groups,omitempty"`
}

// PolicyRule defines a single rule within a policy.
type PolicyRule struct {
	ID          string          `json:"id"`
	Description string          `json:"description,omitempty"`
	Effect      Effect          `json:"effect"`
	Actions     []string        `json:"actions"`     // Action patterns (supports wildcards)
	Resources   []string        `json:"resources"`   // Resource patterns (supports wildcards)
	Conditions  *ConditionGroup `json:"conditions,omitempty"`
	Priority    int             `json:"priority"` // Higher priority = evaluated first
}

// Policy represents an ABAC policy.
type Policy struct {
	ID          uuid.UUID    `json:"id"`
	TenantID    uuid.UUID    `json:"tenant_id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Version     int          `json:"version"`
	Rules       []PolicyRule `json:"rules"`
	Targets     PolicyTarget `json:"targets,omitempty"` // Pre-filter for policy applicability
	Enabled     bool         `json:"enabled"`
	Priority    int          `json:"priority"` // Policy-level priority
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CreatedBy   *uuid.UUID   `json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID   `json:"updated_by,omitempty"`
	Metadata    PolicyMeta   `json:"metadata,omitempty"`
}

// PolicyTarget defines quick-match targets for policy applicability.
type PolicyTarget struct {
	Subjects   []string `json:"subjects,omitempty"`   // User/role patterns
	Resources  []string `json:"resources,omitempty"`  // Resource type patterns
	Actions    []string `json:"actions,omitempty"`    // Action patterns
	Tenants    []string `json:"tenants,omitempty"`    // Tenant patterns
}

// PolicyMeta holds policy metadata.
type PolicyMeta struct {
	Tags       []string          `json:"tags,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Inheritable bool             `json:"inheritable"` // Can be inherited by child resources
	Delegatable bool             `json:"delegatable"` // Can be delegated to other policies
}

// PolicySet represents a collection of policies with combining algorithm.
type PolicySet struct {
	ID          uuid.UUID           `json:"id"`
	TenantID    uuid.UUID           `json:"tenant_id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Algorithm   CombiningAlgorithm  `json:"algorithm"`
	Policies    []uuid.UUID         `json:"policies"` // Policy IDs
	PolicySets  []uuid.UUID         `json:"policy_sets,omitempty"` // Nested policy sets
	Enabled     bool                `json:"enabled"`
	Priority    int                 `json:"priority"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// CombiningAlgorithm defines how multiple policy decisions are combined.
type CombiningAlgorithm string

const (
	// DenyOverrides: If any policy denies, result is deny.
	AlgorithmDenyOverrides CombiningAlgorithm = "deny_overrides"
	// PermitOverrides: If any policy permits, result is permit.
	AlgorithmPermitOverrides CombiningAlgorithm = "permit_overrides"
	// FirstApplicable: Use the first applicable policy's decision.
	AlgorithmFirstApplicable CombiningAlgorithm = "first_applicable"
	// OnlyOneApplicable: Exactly one policy must be applicable.
	AlgorithmOnlyOneApplicable CombiningAlgorithm = "only_one_applicable"
	// HighestPriority: Use highest priority policy's decision.
	AlgorithmHighestPriority CombiningAlgorithm = "highest_priority"
)

// PolicyRepository defines storage operations for policies.
type PolicyRepository interface {
	Create(ctx context.Context, policy *Policy) error
	Update(ctx context.Context, policy *Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Policy, error)
	FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*Policy, error)
	FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*Policy, error)
	FindApplicable(ctx context.Context, tenantID uuid.UUID, resource, action string) ([]*Policy, error)
	List(ctx context.Context, filter PolicyFilter) ([]*Policy, int, error)
}

// PolicyFilter for querying policies.
type PolicyFilter struct {
	TenantID  *uuid.UUID
	Name      *string
	Enabled   *bool
	Tags      []string
	Offset    int
	Limit     int
	SortBy    string
	SortOrder string
}

// PolicySetRepository defines storage operations for policy sets.
type PolicySetRepository interface {
	Create(ctx context.Context, policySet *PolicySet) error
	Update(ctx context.Context, policySet *PolicySet) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*PolicySet, error)
	FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*PolicySet, error)
}

// InMemoryPolicyStore provides in-memory policy storage.
type InMemoryPolicyStore struct {
	policies    map[uuid.UUID]*Policy
	policySets  map[uuid.UUID]*PolicySet
	byTenant    map[uuid.UUID][]uuid.UUID
	byName      map[string]uuid.UUID // tenantID:name -> policyID
	mu          sync.RWMutex
}

// NewInMemoryPolicyStore creates a new in-memory policy store.
func NewInMemoryPolicyStore() *InMemoryPolicyStore {
	return &InMemoryPolicyStore{
		policies:   make(map[uuid.UUID]*Policy),
		policySets: make(map[uuid.UUID]*PolicySet),
		byTenant:   make(map[uuid.UUID][]uuid.UUID),
		byName:     make(map[string]uuid.UUID),
	}
}

// Create creates a new policy.
func (s *InMemoryPolicyStore) Create(ctx context.Context, policy *Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policy.ID]; exists {
		return fmt.Errorf("policy already exists: %s", policy.ID)
	}

	nameKey := fmt.Sprintf("%s:%s", policy.TenantID, policy.Name)
	if _, exists := s.byName[nameKey]; exists {
		return fmt.Errorf("policy with name already exists: %s", policy.Name)
	}

	policy.CreatedAt = time.Now().UTC()
	policy.UpdatedAt = policy.CreatedAt
	policy.Version = 1

	s.policies[policy.ID] = policy
	s.byTenant[policy.TenantID] = append(s.byTenant[policy.TenantID], policy.ID)
	s.byName[nameKey] = policy.ID

	return nil
}

// Update updates a policy.
func (s *InMemoryPolicyStore) Update(ctx context.Context, policy *Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.policies[policy.ID]
	if !exists {
		return ErrPolicyNotFound
	}

	// Update name index if name changed
	if existing.Name != policy.Name {
		oldKey := fmt.Sprintf("%s:%s", existing.TenantID, existing.Name)
		newKey := fmt.Sprintf("%s:%s", policy.TenantID, policy.Name)

		if _, exists := s.byName[newKey]; exists {
			return fmt.Errorf("policy with name already exists: %s", policy.Name)
		}

		delete(s.byName, oldKey)
		s.byName[newKey] = policy.ID
	}

	policy.UpdatedAt = time.Now().UTC()
	policy.Version = existing.Version + 1
	s.policies[policy.ID] = policy

	return nil
}

// Delete deletes a policy.
func (s *InMemoryPolicyStore) Delete(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[id]
	if !exists {
		return ErrPolicyNotFound
	}

	delete(s.policies, id)

	// Remove from tenant index
	tenantPolicies := s.byTenant[policy.TenantID]
	for i, pid := range tenantPolicies {
		if pid == id {
			s.byTenant[policy.TenantID] = append(tenantPolicies[:i], tenantPolicies[i+1:]...)
			break
		}
	}

	// Remove from name index
	nameKey := fmt.Sprintf("%s:%s", policy.TenantID, policy.Name)
	delete(s.byName, nameKey)

	return nil
}

// FindByID finds a policy by ID.
func (s *InMemoryPolicyStore) FindByID(ctx context.Context, id uuid.UUID) (*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.policies[id]
	if !exists {
		return nil, ErrPolicyNotFound
	}

	return policy, nil
}

// FindByTenantID finds all policies for a tenant.
func (s *InMemoryPolicyStore) FindByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policyIDs := s.byTenant[tenantID]
	policies := make([]*Policy, 0, len(policyIDs))

	for _, id := range policyIDs {
		if policy, exists := s.policies[id]; exists {
			policies = append(policies, policy)
		}
	}

	return policies, nil
}

// FindByName finds a policy by name within a tenant.
func (s *InMemoryPolicyStore) FindByName(ctx context.Context, tenantID uuid.UUID, name string) (*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nameKey := fmt.Sprintf("%s:%s", tenantID, name)
	id, exists := s.byName[nameKey]
	if !exists {
		return nil, ErrPolicyNotFound
	}

	return s.policies[id], nil
}

// FindApplicable finds all applicable policies for a resource and action.
func (s *InMemoryPolicyStore) FindApplicable(ctx context.Context, tenantID uuid.UUID, resource, action string) ([]*Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policyIDs := s.byTenant[tenantID]
	var applicable []*Policy

	for _, id := range policyIDs {
		policy, exists := s.policies[id]
		if !exists || !policy.Enabled {
			continue
		}

		if s.policyApplies(policy, resource, action) {
			applicable = append(applicable, policy)
		}
	}

	return applicable, nil
}

// policyApplies checks if a policy applies to the given resource and action.
func (s *InMemoryPolicyStore) policyApplies(policy *Policy, resource, action string) bool {
	// Check targets for quick filtering
	if len(policy.Targets.Resources) > 0 {
		if !matchesAny(resource, policy.Targets.Resources) {
			return false
		}
	}

	if len(policy.Targets.Actions) > 0 {
		if !matchesAny(action, policy.Targets.Actions) {
			return false
		}
	}

	// Check if any rule matches
	for _, rule := range policy.Rules {
		if matchesAny(resource, rule.Resources) && matchesAny(action, rule.Actions) {
			return true
		}
	}

	return false
}

// List lists policies with filtering.
func (s *InMemoryPolicyStore) List(ctx context.Context, filter PolicyFilter) ([]*Policy, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var policies []*Policy

	for _, policy := range s.policies {
		if filter.TenantID != nil && policy.TenantID != *filter.TenantID {
			continue
		}
		if filter.Name != nil && policy.Name != *filter.Name {
			continue
		}
		if filter.Enabled != nil && policy.Enabled != *filter.Enabled {
			continue
		}
		if len(filter.Tags) > 0 && !hasAllTags(policy.Metadata.Tags, filter.Tags) {
			continue
		}
		policies = append(policies, policy)
	}

	total := len(policies)

	// Apply pagination
	if filter.Offset >= len(policies) {
		return []*Policy{}, total, nil
	}

	end := filter.Offset + filter.Limit
	if end > len(policies) || filter.Limit == 0 {
		end = len(policies)
	}

	return policies[filter.Offset:end], total, nil
}

// Helper functions

func matchesAny(value string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchPattern(value, pattern) {
			return true
		}
	}
	return false
}

func matchPattern(value, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	// Check for wildcards
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(value) >= len(prefix) && value[:len(prefix)] == prefix
	}

	if len(pattern) > 0 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix
	}

	// Check for glob pattern
	if containsWildcard(pattern) {
		matched, _ := globMatch(pattern, value)
		return matched
	}

	return value == pattern
}

func containsWildcard(pattern string) bool {
	for _, c := range pattern {
		if c == '*' || c == '?' {
			return true
		}
	}
	return false
}

func globMatch(pattern, value string) (bool, error) {
	// Convert glob to regex
	regexPattern := "^"
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				regexPattern += ".*"
				i++
			} else {
				regexPattern += "[^/]*"
			}
		case '?':
			regexPattern += "."
		case '.', '+', '^', '$', '[', ']', '(', ')', '{', '}', '|', '\\':
			regexPattern += "\\" + string(pattern[i])
		default:
			regexPattern += string(pattern[i])
		}
	}
	regexPattern += "$"

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, err
	}

	return re.MatchString(value), nil
}

func hasAllTags(policyTags, requiredTags []string) bool {
	tagSet := make(map[string]bool)
	for _, tag := range policyTags {
		tagSet[tag] = true
	}

	for _, tag := range requiredTags {
		if !tagSet[tag] {
			return false
		}
	}

	return true
}

// PolicyBuilder helps construct policies.
type PolicyBuilder struct {
	policy Policy
}

// NewPolicyBuilder creates a new policy builder.
func NewPolicyBuilder(tenantID uuid.UUID, name string) *PolicyBuilder {
	return &PolicyBuilder{
		policy: Policy{
			ID:       uuid.New(),
			TenantID: tenantID,
			Name:     name,
			Enabled:  true,
			Rules:    make([]PolicyRule, 0),
		},
	}
}

// WithDescription sets the description.
func (b *PolicyBuilder) WithDescription(description string) *PolicyBuilder {
	b.policy.Description = description
	return b
}

// WithPriority sets the priority.
func (b *PolicyBuilder) WithPriority(priority int) *PolicyBuilder {
	b.policy.Priority = priority
	return b
}

// WithTargets sets the policy targets.
func (b *PolicyBuilder) WithTargets(targets PolicyTarget) *PolicyBuilder {
	b.policy.Targets = targets
	return b
}

// WithMetadata sets the metadata.
func (b *PolicyBuilder) WithMetadata(metadata PolicyMeta) *PolicyBuilder {
	b.policy.Metadata = metadata
	return b
}

// AddRule adds a rule to the policy.
func (b *PolicyBuilder) AddRule(rule PolicyRule) *PolicyBuilder {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	b.policy.Rules = append(b.policy.Rules, rule)
	return b
}

// Build returns the constructed policy.
func (b *PolicyBuilder) Build() *Policy {
	return &b.policy
}

// RuleBuilder helps construct policy rules.
type RuleBuilder struct {
	rule PolicyRule
}

// NewRuleBuilder creates a new rule builder.
func NewRuleBuilder(effect Effect) *RuleBuilder {
	return &RuleBuilder{
		rule: PolicyRule{
			ID:     uuid.New().String(),
			Effect: effect,
		},
	}
}

// WithDescription sets the description.
func (b *RuleBuilder) WithDescription(description string) *RuleBuilder {
	b.rule.Description = description
	return b
}

// WithActions sets the actions.
func (b *RuleBuilder) WithActions(actions ...string) *RuleBuilder {
	b.rule.Actions = actions
	return b
}

// WithResources sets the resources.
func (b *RuleBuilder) WithResources(resources ...string) *RuleBuilder {
	b.rule.Resources = resources
	return b
}

// WithConditions sets the conditions.
func (b *RuleBuilder) WithConditions(conditions *ConditionGroup) *RuleBuilder {
	b.rule.Conditions = conditions
	return b
}

// WithPriority sets the priority.
func (b *RuleBuilder) WithPriority(priority int) *RuleBuilder {
	b.rule.Priority = priority
	return b
}

// Build returns the constructed rule.
func (b *RuleBuilder) Build() PolicyRule {
	return b.rule
}

// ConditionBuilder helps construct conditions.
type ConditionBuilder struct {
	group ConditionGroup
}

// NewConditionBuilder creates a new condition builder.
func NewConditionBuilder(operator LogicalOperator) *ConditionBuilder {
	return &ConditionBuilder{
		group: ConditionGroup{
			Operator: operator,
		},
	}
}

// And creates an AND condition group.
func And(conditions ...Condition) *ConditionGroup {
	return &ConditionGroup{
		Operator:   LogicalAnd,
		Conditions: conditions,
	}
}

// Or creates an OR condition group.
func Or(conditions ...Condition) *ConditionGroup {
	return &ConditionGroup{
		Operator:   LogicalOr,
		Conditions: conditions,
	}
}

// Not creates a NOT condition group (negates the first condition).
func Not(condition Condition) *ConditionGroup {
	return &ConditionGroup{
		Operator:   LogicalNot,
		Conditions: []Condition{condition},
	}
}

// SubjectAttr creates a subject attribute reference.
func SubjectAttr(key string) AttributeReference {
	return AttributeReference{Source: SourceSubject, Key: key}
}

// ResourceAttr creates a resource attribute reference.
func ResourceAttr(key string) AttributeReference {
	return AttributeReference{Source: SourceResource, Key: key}
}

// ActionAttr creates an action attribute reference.
func ActionAttr(key string) AttributeReference {
	return AttributeReference{Source: SourceAction, Key: key}
}

// EnvAttr creates an environment attribute reference.
func EnvAttr(key string) AttributeReference {
	return AttributeReference{Source: SourceEnvironment, Key: key}
}

// ContextAttr creates a context attribute reference.
func ContextAttr(key string) AttributeReference {
	return AttributeReference{Source: SourceContext, Key: key}
}

// Equals creates an equals condition.
func Equals(attr AttributeReference, value interface{}) Condition {
	return Condition{Attribute: attr, Operator: OpEquals, Value: value}
}

// NotEquals creates a not equals condition.
func NotEquals(attr AttributeReference, value interface{}) Condition {
	return Condition{Attribute: attr, Operator: OpNotEquals, Value: value}
}

// Contains creates a contains condition.
func Contains(attr AttributeReference, value interface{}) Condition {
	return Condition{Attribute: attr, Operator: OpContains, Value: value}
}

// In creates an in condition.
func In(attr AttributeReference, values ...interface{}) Condition {
	return Condition{Attribute: attr, Operator: OpIn, Value: values}
}

// NotIn creates a not in condition.
func NotIn(attr AttributeReference, values ...interface{}) Condition {
	return Condition{Attribute: attr, Operator: OpNotIn, Value: values}
}

// GreaterThan creates a greater than condition.
func GreaterThan(attr AttributeReference, value interface{}) Condition {
	return Condition{Attribute: attr, Operator: OpGreaterThan, Value: value}
}

// LessThan creates a less than condition.
func LessThan(attr AttributeReference, value interface{}) Condition {
	return Condition{Attribute: attr, Operator: OpLessThan, Value: value}
}

// Exists creates an exists condition.
func Exists(attr AttributeReference) Condition {
	return Condition{Attribute: attr, Operator: OpExists, Value: nil}
}

// IpInCIDR creates an IP in CIDR condition.
func IpInCIDR(attr AttributeReference, cidr string) Condition {
	return Condition{Attribute: attr, Operator: OpIpInCIDR, Value: cidr}
}

// Serialize serializes a policy to JSON.
func (p *Policy) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// Deserialize deserializes a policy from JSON.
func DeserializePolicy(data []byte) (*Policy, error) {
	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to deserialize policy: %w", err)
	}
	return &policy, nil
}
