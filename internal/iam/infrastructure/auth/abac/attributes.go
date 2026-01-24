// Package abac provides Attribute-Based Access Control infrastructure.
package abac

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AttributeProvider provides additional attributes for ABAC evaluation.
type AttributeProvider interface {
	Name() string
	GetAttributes(ctx context.Context, request *EvaluationRequest) (map[AttributeSource]map[string]interface{}, error)
}

// SubjectAttributeProvider provides subject (user) attributes.
type SubjectAttributeProvider interface {
	GetSubjectAttributes(ctx context.Context, subjectID uuid.UUID) (map[string]interface{}, error)
}

// ResourceAttributeProvider provides resource attributes.
type ResourceAttributeProvider interface {
	GetResourceAttributes(ctx context.Context, resourceType string, resourceID uuid.UUID) (map[string]interface{}, error)
}

// EnvironmentAttributeProvider provides environment attributes.
type EnvironmentAttributeProvider interface {
	GetEnvironmentAttributes(ctx context.Context) (map[string]interface{}, error)
}

// CompositeAttributeProvider combines multiple attribute providers.
type CompositeAttributeProvider struct {
	name      string
	providers []AttributeProvider
	logger    *slog.Logger
}

// NewCompositeAttributeProvider creates a new composite provider.
func NewCompositeAttributeProvider(name string, logger *slog.Logger, providers ...AttributeProvider) *CompositeAttributeProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &CompositeAttributeProvider{
		name:      name,
		providers: providers,
		logger:    logger,
	}
}

// Name returns the provider name.
func (p *CompositeAttributeProvider) Name() string {
	return p.name
}

// GetAttributes gets attributes from all providers.
func (p *CompositeAttributeProvider) GetAttributes(ctx context.Context, request *EvaluationRequest) (map[AttributeSource]map[string]interface{}, error) {
	result := make(map[AttributeSource]map[string]interface{})
	result[SourceSubject] = make(map[string]interface{})
	result[SourceResource] = make(map[string]interface{})
	result[SourceAction] = make(map[string]interface{})
	result[SourceEnvironment] = make(map[string]interface{})
	result[SourceContext] = make(map[string]interface{})

	for _, provider := range p.providers {
		attrs, err := provider.GetAttributes(ctx, request)
		if err != nil {
			p.logger.Warn("attribute provider failed",
				slog.String("provider", provider.Name()),
				slog.Any("error", err),
			)
			continue
		}

		for source, values := range attrs {
			if result[source] == nil {
				result[source] = make(map[string]interface{})
			}
			for k, v := range values {
				result[source][k] = v
			}
		}
	}

	return result, nil
}

// AddProvider adds a provider to the composite.
func (p *CompositeAttributeProvider) AddProvider(provider AttributeProvider) {
	p.providers = append(p.providers, provider)
}

// UserAttributeProvider provides user-related attributes.
type UserAttributeProvider struct {
	name          string
	userStore     UserStore
	roleStore     RoleStore
	tenantStore   TenantStore
	cache         *attributeCache
	logger        *slog.Logger
}

// UserStore interface for fetching user data.
type UserStore interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*UserData, error)
}

// RoleStore interface for fetching role data.
type RoleStore interface {
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]RoleData, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)
}

// TenantStore interface for fetching tenant data.
type TenantStore interface {
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*TenantData, error)
}

// UserData represents user data for attribute provision.
type UserData struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Email         string
	Name          string
	Department    string
	Title         string
	Level         int
	Attributes    map[string]interface{}
	CreatedAt     time.Time
	IsActive      bool
}

// RoleData represents role data.
type RoleData struct {
	ID          uuid.UUID
	Name        string
	Type        string
	Level       int
	Permissions []string
}

// TenantData represents tenant data.
type TenantData struct {
	ID         uuid.UUID
	Name       string
	Plan       string
	Features   []string
	MaxUsers   int
	Attributes map[string]interface{}
}

// attributeCache caches attributes with TTL.
type attributeCache struct {
	data   map[string]cachedAttribute
	ttl    time.Duration
	mu     sync.RWMutex
}

type cachedAttribute struct {
	value     map[string]interface{}
	expiresAt time.Time
}

func newAttributeCache(ttl time.Duration) *attributeCache {
	cache := &attributeCache{
		data: make(map[string]cachedAttribute),
		ttl:  ttl,
	}
	go cache.cleanup()
	return cache
}

func (c *attributeCache) get(key string) (map[string]interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.data[key]
	if !exists || time.Now().After(cached.expiresAt) {
		return nil, false
	}

	return cached.value, true
}

func (c *attributeCache) set(key string, value map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cachedAttribute{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *attributeCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, cached := range c.data {
			if now.After(cached.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// UserAttributeProviderConfig holds configuration.
type UserAttributeProviderConfig struct {
	Name        string
	CacheTTL    time.Duration
	Logger      *slog.Logger
}

// NewUserAttributeProvider creates a new user attribute provider.
func NewUserAttributeProvider(
	config UserAttributeProviderConfig,
	userStore UserStore,
	roleStore RoleStore,
	tenantStore TenantStore,
) *UserAttributeProvider {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	cacheTTL := config.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute
	}

	return &UserAttributeProvider{
		name:        config.Name,
		userStore:   userStore,
		roleStore:   roleStore,
		tenantStore: tenantStore,
		cache:       newAttributeCache(cacheTTL),
		logger:      logger,
	}
}

// Name returns the provider name.
func (p *UserAttributeProvider) Name() string {
	return p.name
}

// GetAttributes gets user-related attributes.
func (p *UserAttributeProvider) GetAttributes(ctx context.Context, request *EvaluationRequest) (map[AttributeSource]map[string]interface{}, error) {
	result := make(map[AttributeSource]map[string]interface{})

	// Get subject ID from request
	subjectIDStr, ok := request.Subject["id"].(string)
	if !ok {
		return result, nil
	}

	subjectID, err := uuid.Parse(subjectIDStr)
	if err != nil {
		return result, nil
	}

	// Check cache
	cacheKey := "user:" + subjectID.String()
	if cached, found := p.cache.get(cacheKey); found {
		result[SourceSubject] = cached
		return result, nil
	}

	// Fetch user data
	attrs := make(map[string]interface{})

	if p.userStore != nil {
		user, err := p.userStore.GetUser(ctx, subjectID)
		if err == nil && user != nil {
			attrs["email"] = user.Email
			attrs["name"] = user.Name
			attrs["department"] = user.Department
			attrs["title"] = user.Title
			attrs["level"] = user.Level
			attrs["is_active"] = user.IsActive
			attrs["created_at"] = user.CreatedAt

			for k, v := range user.Attributes {
				attrs[k] = v
			}
		}
	}

	// Fetch role data
	if p.roleStore != nil {
		roles, err := p.roleStore.GetUserRoles(ctx, subjectID)
		if err == nil {
			roleNames := make([]string, 0, len(roles))
			permissions := make([]string, 0)
			maxLevel := 0

			for _, role := range roles {
				roleNames = append(roleNames, role.Name)
				permissions = append(permissions, role.Permissions...)
				if role.Level > maxLevel {
					maxLevel = role.Level
				}
			}

			attrs["roles"] = roleNames
			attrs["permissions"] = unique(permissions)
			attrs["max_role_level"] = maxLevel
		}
	}

	// Fetch tenant data
	if p.tenantStore != nil {
		tenant, err := p.tenantStore.GetTenant(ctx, request.TenantID)
		if err == nil && tenant != nil {
			attrs["tenant_name"] = tenant.Name
			attrs["tenant_plan"] = tenant.Plan
			attrs["tenant_features"] = tenant.Features
		}
	}

	// Cache the result
	p.cache.set(cacheKey, attrs)

	result[SourceSubject] = attrs
	return result, nil
}

// EnvironmentProvider provides environment attributes.
type EnvironmentProvider struct {
	name       string
	location   string
	timezone   *time.Location
	attributes map[string]interface{}
}

// EnvironmentProviderConfig holds configuration.
type EnvironmentProviderConfig struct {
	Name       string
	Location   string
	Timezone   string
	Attributes map[string]interface{}
}

// NewEnvironmentProvider creates a new environment provider.
func NewEnvironmentProvider(config EnvironmentProviderConfig) *EnvironmentProvider {
	loc := time.UTC
	if config.Timezone != "" {
		if l, err := time.LoadLocation(config.Timezone); err == nil {
			loc = l
		}
	}

	return &EnvironmentProvider{
		name:       config.Name,
		location:   config.Location,
		timezone:   loc,
		attributes: config.Attributes,
	}
}

// Name returns the provider name.
func (p *EnvironmentProvider) Name() string {
	return p.name
}

// GetAttributes gets environment attributes.
func (p *EnvironmentProvider) GetAttributes(ctx context.Context, request *EvaluationRequest) (map[AttributeSource]map[string]interface{}, error) {
	now := time.Now().In(p.timezone)

	attrs := map[string]interface{}{
		"time":           now.Format("15:04"),
		"date":           now.Format("2006-01-02"),
		"datetime":       now.Format(time.RFC3339),
		"day_of_week":    int(now.Weekday()),
		"day_of_month":   now.Day(),
		"month":          int(now.Month()),
		"year":           now.Year(),
		"hour":           now.Hour(),
		"minute":         now.Minute(),
		"is_weekend":     now.Weekday() == time.Saturday || now.Weekday() == time.Sunday,
		"is_business_hours": now.Hour() >= 9 && now.Hour() < 17 && now.Weekday() >= time.Monday && now.Weekday() <= time.Friday,
		"timezone":       p.timezone.String(),
		"location":       p.location,
	}

	// Add custom attributes
	for k, v := range p.attributes {
		attrs[k] = v
	}

	return map[AttributeSource]map[string]interface{}{
		SourceEnvironment: attrs,
	}, nil
}

// RequestContextProvider provides request context attributes.
type RequestContextProvider struct {
	name string
}

// NewRequestContextProvider creates a new request context provider.
func NewRequestContextProvider(name string) *RequestContextProvider {
	return &RequestContextProvider{name: name}
}

// Name returns the provider name.
func (p *RequestContextProvider) Name() string {
	return p.name
}

// GetAttributes gets request context attributes from the existing context.
func (p *RequestContextProvider) GetAttributes(ctx context.Context, request *EvaluationRequest) (map[AttributeSource]map[string]interface{}, error) {
	// Context attributes are already in request.Context
	// This provider can extract additional context values

	attrs := make(map[string]interface{})

	// Copy existing context attributes
	for k, v := range request.Context {
		attrs[k] = v
	}

	// Add any values from Go context
	if ip, ok := ctx.Value("client_ip").(string); ok {
		attrs["client_ip"] = ip
	}

	if ua, ok := ctx.Value("user_agent").(string); ok {
		attrs["user_agent"] = ua
	}

	if reqID, ok := ctx.Value("request_id").(string); ok {
		attrs["request_id"] = reqID
	}

	return map[AttributeSource]map[string]interface{}{
		SourceContext: attrs,
	}, nil
}

// ResourceOwnershipProvider provides resource ownership attributes.
type ResourceOwnershipProvider struct {
	name          string
	ownershipStore ResourceOwnershipStore
	cache         *attributeCache
	logger        *slog.Logger
}

// ResourceOwnershipStore interface for resource ownership lookups.
type ResourceOwnershipStore interface {
	GetOwner(ctx context.Context, resourceType string, resourceID uuid.UUID) (*uuid.UUID, error)
	GetSharedWith(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]uuid.UUID, error)
	GetResourceTenant(ctx context.Context, resourceType string, resourceID uuid.UUID) (*uuid.UUID, error)
}

// NewResourceOwnershipProvider creates a new resource ownership provider.
func NewResourceOwnershipProvider(name string, store ResourceOwnershipStore, logger *slog.Logger) *ResourceOwnershipProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &ResourceOwnershipProvider{
		name:           name,
		ownershipStore: store,
		cache:          newAttributeCache(5 * time.Minute),
		logger:         logger,
	}
}

// Name returns the provider name.
func (p *ResourceOwnershipProvider) Name() string {
	return p.name
}

// GetAttributes gets resource ownership attributes.
func (p *ResourceOwnershipProvider) GetAttributes(ctx context.Context, request *EvaluationRequest) (map[AttributeSource]map[string]interface{}, error) {
	result := make(map[AttributeSource]map[string]interface{})

	resourceType, ok := request.Resource["type"].(string)
	if !ok {
		return result, nil
	}

	resourceIDStr, ok := request.Resource["id"].(string)
	if !ok {
		return result, nil
	}

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		return result, nil
	}

	// Check cache
	cacheKey := resourceType + ":" + resourceID.String()
	if cached, found := p.cache.get(cacheKey); found {
		result[SourceResource] = cached
		return result, nil
	}

	attrs := make(map[string]interface{})

	if p.ownershipStore != nil {
		// Get owner
		ownerID, err := p.ownershipStore.GetOwner(ctx, resourceType, resourceID)
		if err == nil && ownerID != nil {
			attrs["owner_id"] = ownerID.String()

			// Check if current user is owner
			if subjectID, ok := request.Subject["id"].(string); ok {
				attrs["is_owner"] = subjectID == ownerID.String()
			}
		}

		// Get shared users
		sharedWith, err := p.ownershipStore.GetSharedWith(ctx, resourceType, resourceID)
		if err == nil {
			sharedIDs := make([]string, len(sharedWith))
			for i, id := range sharedWith {
				sharedIDs[i] = id.String()
			}
			attrs["shared_with"] = sharedIDs

			// Check if current user has access
			if subjectID, ok := request.Subject["id"].(string); ok {
				for _, id := range sharedIDs {
					if id == subjectID {
						attrs["is_shared_with_me"] = true
						break
					}
				}
			}
		}

		// Get resource tenant
		tenantID, err := p.ownershipStore.GetResourceTenant(ctx, resourceType, resourceID)
		if err == nil && tenantID != nil {
			attrs["resource_tenant_id"] = tenantID.String()
			attrs["same_tenant"] = request.TenantID.String() == tenantID.String()
		}
	}

	// Cache the result
	p.cache.set(cacheKey, attrs)

	result[SourceResource] = attrs
	return result, nil
}

// HierarchyProvider provides organizational hierarchy attributes.
type HierarchyProvider struct {
	name           string
	hierarchyStore HierarchyStore
	cache          *attributeCache
	logger         *slog.Logger
}

// HierarchyStore interface for organizational hierarchy lookups.
type HierarchyStore interface {
	GetManager(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
	GetSubordinates(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetDepartment(ctx context.Context, userID uuid.UUID) (string, error)
	GetDepartmentMembers(ctx context.Context, department string) ([]uuid.UUID, error)
	GetHierarchyPath(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// NewHierarchyProvider creates a new hierarchy provider.
func NewHierarchyProvider(name string, store HierarchyStore, logger *slog.Logger) *HierarchyProvider {
	if logger == nil {
		logger = slog.Default()
	}
	return &HierarchyProvider{
		name:           name,
		hierarchyStore: store,
		cache:          newAttributeCache(10 * time.Minute),
		logger:         logger,
	}
}

// Name returns the provider name.
func (p *HierarchyProvider) Name() string {
	return p.name
}

// GetAttributes gets hierarchy attributes.
func (p *HierarchyProvider) GetAttributes(ctx context.Context, request *EvaluationRequest) (map[AttributeSource]map[string]interface{}, error) {
	result := make(map[AttributeSource]map[string]interface{})

	subjectIDStr, ok := request.Subject["id"].(string)
	if !ok {
		return result, nil
	}

	subjectID, err := uuid.Parse(subjectIDStr)
	if err != nil {
		return result, nil
	}

	// Check cache
	cacheKey := "hierarchy:" + subjectID.String()
	if cached, found := p.cache.get(cacheKey); found {
		result[SourceSubject] = cached
		return result, nil
	}

	attrs := make(map[string]interface{})

	if p.hierarchyStore != nil {
		// Get manager
		managerID, err := p.hierarchyStore.GetManager(ctx, subjectID)
		if err == nil && managerID != nil {
			attrs["manager_id"] = managerID.String()
		}

		// Get subordinates
		subordinates, err := p.hierarchyStore.GetSubordinates(ctx, subjectID)
		if err == nil {
			subIDs := make([]string, len(subordinates))
			for i, id := range subordinates {
				subIDs[i] = id.String()
			}
			attrs["subordinate_ids"] = subIDs
			attrs["is_manager"] = len(subordinates) > 0
		}

		// Get department
		department, err := p.hierarchyStore.GetDepartment(ctx, subjectID)
		if err == nil {
			attrs["department"] = department
		}

		// Get hierarchy path (for hierarchical permissions)
		path, err := p.hierarchyStore.GetHierarchyPath(ctx, subjectID)
		if err == nil {
			pathIDs := make([]string, len(path))
			for i, id := range path {
				pathIDs[i] = id.String()
			}
			attrs["hierarchy_path"] = pathIDs
			attrs["hierarchy_level"] = len(path)
		}

		// Check resource owner relationship
		if resourceOwnerID, ok := request.Resource["owner_id"].(string); ok {
			// Check if subject is manager of resource owner
			ownerID, _ := uuid.Parse(resourceOwnerID)
			ownerPath, err := p.hierarchyStore.GetHierarchyPath(ctx, ownerID)
			if err == nil {
				for _, id := range ownerPath {
					if id == subjectID {
						attrs["is_manager_of_owner"] = true
						break
					}
				}
			}
		}
	}

	// Cache the result
	p.cache.set(cacheKey, attrs)

	result[SourceSubject] = attrs
	return result, nil
}

// Helper functions

func unique(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// StaticAttributeProvider provides static attributes.
type StaticAttributeProvider struct {
	name       string
	attributes map[AttributeSource]map[string]interface{}
}

// NewStaticAttributeProvider creates a new static attribute provider.
func NewStaticAttributeProvider(name string) *StaticAttributeProvider {
	return &StaticAttributeProvider{
		name:       name,
		attributes: make(map[AttributeSource]map[string]interface{}),
	}
}

// Name returns the provider name.
func (p *StaticAttributeProvider) Name() string {
	return p.name
}

// SetAttribute sets a static attribute.
func (p *StaticAttributeProvider) SetAttribute(source AttributeSource, key string, value interface{}) {
	if p.attributes[source] == nil {
		p.attributes[source] = make(map[string]interface{})
	}
	p.attributes[source][key] = value
}

// GetAttributes returns the static attributes.
func (p *StaticAttributeProvider) GetAttributes(ctx context.Context, request *EvaluationRequest) (map[AttributeSource]map[string]interface{}, error) {
	result := make(map[AttributeSource]map[string]interface{})
	for source, attrs := range p.attributes {
		result[source] = make(map[string]interface{})
		for k, v := range attrs {
			result[source][k] = v
		}
	}
	return result, nil
}
