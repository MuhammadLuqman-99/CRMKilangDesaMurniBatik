// Package discovery provides service discovery abstractions for the CRM application.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Common Errors
// ============================================================================

var (
	// ErrServiceNotFound is returned when a service is not found.
	ErrServiceNotFound = errors.New("service not found")

	// ErrNoHealthyInstances is returned when no healthy instances are available.
	ErrNoHealthyInstances = errors.New("no healthy instances available")

	// ErrRegistrationFailed is returned when service registration fails.
	ErrRegistrationFailed = errors.New("service registration failed")

	// ErrDeregistrationFailed is returned when service deregistration fails.
	ErrDeregistrationFailed = errors.New("service deregistration failed")
)

// ============================================================================
// Service Instance
// ============================================================================

// ServiceInstance represents a registered service instance.
type ServiceInstance struct {
	ID          string            `json:"id"`
	ServiceName string            `json:"service_name"`
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Protocol    string            `json:"protocol"` // http, https, grpc
	Metadata    map[string]string `json:"metadata"`
	Tags        []string          `json:"tags"`
	Health      HealthStatus      `json:"health"`
	Weight      int               `json:"weight"` // For weighted load balancing
	Zone        string            `json:"zone"`   // For zone-aware routing
	Version     string            `json:"version"`
	RegisteredAt time.Time        `json:"registered_at"`
	LastHeartbeat time.Time       `json:"last_heartbeat"`
}

// Address returns the full address of the instance.
func (s *ServiceInstance) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// URL returns the full URL of the instance.
func (s *ServiceInstance) URL() string {
	protocol := s.Protocol
	if protocol == "" {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s:%d", protocol, s.Host, s.Port)
}

// IsHealthy returns true if the instance is healthy.
func (s *ServiceInstance) IsHealthy() bool {
	return s.Health == HealthStatusHealthy || s.Health == HealthStatusWarning
}

// HealthStatus represents the health status of a service instance.
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusWarning   HealthStatus = "warning"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ============================================================================
// Health Check Configuration
// ============================================================================

// HealthCheckConfig configures health checks for a service.
type HealthCheckConfig struct {
	// Type of health check (http, tcp, grpc, script)
	Type string `json:"type"`

	// Endpoint for HTTP/gRPC health checks
	Endpoint string `json:"endpoint"`

	// Interval between health checks
	Interval time.Duration `json:"interval"`

	// Timeout for health check requests
	Timeout time.Duration `json:"timeout"`

	// DeregisterAfter removes unhealthy service after this duration
	DeregisterAfter time.Duration `json:"deregister_after"`

	// SuccessThreshold is the number of consecutive successes to be considered healthy
	SuccessThreshold int `json:"success_threshold"`

	// FailureThreshold is the number of consecutive failures to be considered unhealthy
	FailureThreshold int `json:"failure_threshold"`
}

// DefaultHealthCheckConfig returns default health check configuration.
func DefaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		Type:             "http",
		Endpoint:         "/health",
		Interval:         10 * time.Second,
		Timeout:          5 * time.Second,
		DeregisterAfter:  60 * time.Second,
		SuccessThreshold: 2,
		FailureThreshold: 3,
	}
}

// ============================================================================
// Service Registration
// ============================================================================

// ServiceRegistration contains information for registering a service.
type ServiceRegistration struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	Protocol     string            `json:"protocol"`
	Tags         []string          `json:"tags"`
	Metadata     map[string]string `json:"metadata"`
	HealthCheck  *HealthCheckConfig `json:"health_check"`
	Weight       int               `json:"weight"`
	Zone         string            `json:"zone"`
	Version      string            `json:"version"`
	TTL          time.Duration     `json:"ttl"` // For TTL-based registration
}

// ============================================================================
// Service Discovery Interface
// ============================================================================

// ServiceDiscovery defines the interface for service discovery.
type ServiceDiscovery interface {
	// Register registers a service instance.
	Register(ctx context.Context, registration *ServiceRegistration) error

	// Deregister removes a service instance.
	Deregister(ctx context.Context, serviceID string) error

	// GetService retrieves all instances of a service.
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)

	// GetHealthyService retrieves only healthy instances of a service.
	GetHealthyService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)

	// GetServiceInstance retrieves a specific service instance.
	GetServiceInstance(ctx context.Context, serviceID string) (*ServiceInstance, error)

	// ListServices lists all registered services.
	ListServices(ctx context.Context) ([]string, error)

	// Watch watches for changes to a service.
	Watch(ctx context.Context, serviceName string, callback func([]*ServiceInstance)) error

	// Heartbeat sends a heartbeat for a service instance.
	Heartbeat(ctx context.Context, serviceID string) error

	// Close closes the service discovery connection.
	Close() error
}

// ============================================================================
// Service Registry (Local Implementation)
// ============================================================================

// LocalRegistry provides an in-memory service registry for development/testing.
type LocalRegistry struct {
	instances map[string]*ServiceInstance // instanceID -> instance
	services  map[string][]string          // serviceName -> []instanceID
	watchers  map[string][]chan []*ServiceInstance
	mu        sync.RWMutex
	healthMu  sync.RWMutex
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// NewLocalRegistry creates a new local registry.
func NewLocalRegistry() *LocalRegistry {
	r := &LocalRegistry{
		instances: make(map[string]*ServiceInstance),
		services:  make(map[string][]string),
		watchers:  make(map[string][]chan []*ServiceInstance),
		stopCh:    make(chan struct{}),
	}

	// Start health check goroutine
	r.wg.Add(1)
	go r.runHealthChecks()

	return r
}

// Register registers a service instance.
func (r *LocalRegistry) Register(ctx context.Context, registration *ServiceRegistration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance := &ServiceInstance{
		ID:           registration.ID,
		ServiceName:  registration.Name,
		Host:         registration.Host,
		Port:         registration.Port,
		Protocol:     registration.Protocol,
		Metadata:     registration.Metadata,
		Tags:         registration.Tags,
		Health:       HealthStatusHealthy,
		Weight:       registration.Weight,
		Zone:         registration.Zone,
		Version:      registration.Version,
		RegisteredAt: time.Now(),
		LastHeartbeat: time.Now(),
	}

	if instance.Weight == 0 {
		instance.Weight = 1
	}

	r.instances[registration.ID] = instance

	// Add to service list
	if _, ok := r.services[registration.Name]; !ok {
		r.services[registration.Name] = make([]string, 0)
	}
	r.services[registration.Name] = append(r.services[registration.Name], registration.ID)

	// Notify watchers
	go r.notifyWatchers(registration.Name)

	return nil
}

// Deregister removes a service instance.
func (r *LocalRegistry) Deregister(ctx context.Context, serviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, ok := r.instances[serviceID]
	if !ok {
		return ErrServiceNotFound
	}

	serviceName := instance.ServiceName

	// Remove from instances
	delete(r.instances, serviceID)

	// Remove from service list
	if ids, ok := r.services[serviceName]; ok {
		newIDs := make([]string, 0, len(ids)-1)
		for _, id := range ids {
			if id != serviceID {
				newIDs = append(newIDs, id)
			}
		}
		if len(newIDs) == 0 {
			delete(r.services, serviceName)
		} else {
			r.services[serviceName] = newIDs
		}
	}

	// Notify watchers
	go r.notifyWatchers(serviceName)

	return nil
}

// GetService retrieves all instances of a service.
func (r *LocalRegistry) GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, ok := r.services[serviceName]
	if !ok {
		return nil, ErrServiceNotFound
	}

	instances := make([]*ServiceInstance, 0, len(ids))
	for _, id := range ids {
		if instance, ok := r.instances[id]; ok {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// GetHealthyService retrieves only healthy instances of a service.
func (r *LocalRegistry) GetHealthyService(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	instances, err := r.GetService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	healthy := make([]*ServiceInstance, 0, len(instances))
	for _, instance := range instances {
		if instance.IsHealthy() {
			healthy = append(healthy, instance)
		}
	}

	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}

	return healthy, nil
}

// GetServiceInstance retrieves a specific service instance.
func (r *LocalRegistry) GetServiceInstance(ctx context.Context, serviceID string) (*ServiceInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, ok := r.instances[serviceID]
	if !ok {
		return nil, ErrServiceNotFound
	}

	return instance, nil
}

// ListServices lists all registered services.
func (r *LocalRegistry) ListServices(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]string, 0, len(r.services))
	for name := range r.services {
		services = append(services, name)
	}

	return services, nil
}

// Watch watches for changes to a service.
func (r *LocalRegistry) Watch(ctx context.Context, serviceName string, callback func([]*ServiceInstance)) error {
	ch := make(chan []*ServiceInstance, 10)

	r.mu.Lock()
	if _, ok := r.watchers[serviceName]; !ok {
		r.watchers[serviceName] = make([]chan []*ServiceInstance, 0)
	}
	r.watchers[serviceName] = append(r.watchers[serviceName], ch)
	r.mu.Unlock()

	// Send initial state
	instances, _ := r.GetService(ctx, serviceName)
	callback(instances)

	go func() {
		for {
			select {
			case <-ctx.Done():
				r.removeWatcher(serviceName, ch)
				return
			case instances := <-ch:
				callback(instances)
			}
		}
	}()

	return nil
}

// removeWatcher removes a watcher channel.
func (r *LocalRegistry) removeWatcher(serviceName string, ch chan []*ServiceInstance) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if watchers, ok := r.watchers[serviceName]; ok {
		newWatchers := make([]chan []*ServiceInstance, 0, len(watchers)-1)
		for _, w := range watchers {
			if w != ch {
				newWatchers = append(newWatchers, w)
			}
		}
		if len(newWatchers) == 0 {
			delete(r.watchers, serviceName)
		} else {
			r.watchers[serviceName] = newWatchers
		}
	}

	close(ch)
}

// notifyWatchers notifies all watchers of a service change.
func (r *LocalRegistry) notifyWatchers(serviceName string) {
	r.mu.RLock()
	watchers := r.watchers[serviceName]
	r.mu.RUnlock()

	if len(watchers) == 0 {
		return
	}

	instances, _ := r.GetService(context.Background(), serviceName)

	for _, ch := range watchers {
		select {
		case ch <- instances:
		default:
			// Skip if channel is full
		}
	}
}

// Heartbeat sends a heartbeat for a service instance.
func (r *LocalRegistry) Heartbeat(ctx context.Context, serviceID string) error {
	r.healthMu.Lock()
	defer r.healthMu.Unlock()

	instance, ok := r.instances[serviceID]
	if !ok {
		return ErrServiceNotFound
	}

	instance.LastHeartbeat = time.Now()
	instance.Health = HealthStatusHealthy

	return nil
}

// runHealthChecks periodically checks instance health.
func (r *LocalRegistry) runHealthChecks() {
	defer r.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.checkHealth()
		}
	}
}

// checkHealth checks the health of all instances.
func (r *LocalRegistry) checkHealth() {
	r.healthMu.Lock()
	defer r.healthMu.Unlock()

	now := time.Now()
	unhealthyThreshold := 30 * time.Second
	removeThreshold := 60 * time.Second

	for id, instance := range r.instances {
		elapsed := now.Sub(instance.LastHeartbeat)

		if elapsed > removeThreshold {
			// Remove stale instance
			go r.Deregister(context.Background(), id)
		} else if elapsed > unhealthyThreshold {
			// Mark as unhealthy
			if instance.Health != HealthStatusUnhealthy {
				instance.Health = HealthStatusUnhealthy
				go r.notifyWatchers(instance.ServiceName)
			}
		}
	}
}

// Close closes the local registry.
func (r *LocalRegistry) Close() error {
	close(r.stopCh)
	r.wg.Wait()
	return nil
}

// ============================================================================
// Service Discovery Client
// ============================================================================

// DiscoveryClient provides high-level service discovery functionality.
type DiscoveryClient struct {
	discovery    ServiceDiscovery
	cache        *ServiceCache
	cacheEnabled bool
	cacheTTL     time.Duration
}

// DiscoveryClientConfig configures the discovery client.
type DiscoveryClientConfig struct {
	CacheEnabled bool
	CacheTTL     time.Duration
}

// NewDiscoveryClient creates a new discovery client.
func NewDiscoveryClient(discovery ServiceDiscovery, config DiscoveryClientConfig) *DiscoveryClient {
	client := &DiscoveryClient{
		discovery:    discovery,
		cacheEnabled: config.CacheEnabled,
		cacheTTL:     config.CacheTTL,
	}

	if config.CacheEnabled {
		client.cache = NewServiceCache(config.CacheTTL)
	}

	return client
}

// GetService retrieves service instances with caching.
func (c *DiscoveryClient) GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	if c.cacheEnabled {
		if instances, ok := c.cache.Get(serviceName); ok {
			return instances, nil
		}
	}

	instances, err := c.discovery.GetHealthyService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	if c.cacheEnabled {
		c.cache.Set(serviceName, instances)
	}

	return instances, nil
}

// GetInstance retrieves a single instance using load balancing.
func (c *DiscoveryClient) GetInstance(ctx context.Context, serviceName string, lb LoadBalancer) (*ServiceInstance, error) {
	instances, err := c.GetService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	return lb.Select(instances)
}

// Watch watches for service changes and updates cache.
func (c *DiscoveryClient) Watch(ctx context.Context, serviceName string) error {
	return c.discovery.Watch(ctx, serviceName, func(instances []*ServiceInstance) {
		if c.cacheEnabled {
			healthy := make([]*ServiceInstance, 0)
			for _, i := range instances {
				if i.IsHealthy() {
					healthy = append(healthy, i)
				}
			}
			c.cache.Set(serviceName, healthy)
		}
	})
}

// ============================================================================
// Service Cache
// ============================================================================

// ServiceCache caches service instances.
type ServiceCache struct {
	entries map[string]*cacheEntry
	ttl     time.Duration
	mu      sync.RWMutex
}

type cacheEntry struct {
	instances []*ServiceInstance
	expiresAt time.Time
}

// NewServiceCache creates a new service cache.
func NewServiceCache(ttl time.Duration) *ServiceCache {
	cache := &ServiceCache{
		entries: make(map[string]*cacheEntry),
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves instances from cache.
func (c *ServiceCache) Get(serviceName string) ([]*ServiceInstance, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[serviceName]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.instances, true
}

// Set stores instances in cache.
func (c *ServiceCache) Set(serviceName string, instances []*ServiceInstance) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[serviceName] = &cacheEntry{
		instances: instances,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a service from cache.
func (c *ServiceCache) Invalidate(serviceName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, serviceName)
}

// cleanup removes expired entries periodically.
func (c *ServiceCache) cleanup() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for name, entry := range c.entries {
			if now.After(entry.expiresAt) {
				delete(c.entries, name)
			}
		}
		c.mu.Unlock()
	}
}

// ============================================================================
// Load Balancer Interface
// ============================================================================

// LoadBalancer selects a service instance from available instances.
type LoadBalancer interface {
	Select(instances []*ServiceInstance) (*ServiceInstance, error)
}
