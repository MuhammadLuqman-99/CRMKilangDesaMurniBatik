// Package discovery provides service discovery abstractions for the CRM application.
package discovery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// etcd Client
// ============================================================================

// EtcdConfig configures the etcd client.
type EtcdConfig struct {
	Endpoints   []string      // etcd endpoints (default: localhost:2379)
	Prefix      string        // Key prefix for services (default: /services)
	Username    string        // Username for authentication
	Password    string        // Password for authentication
	HTTPTimeout time.Duration // HTTP client timeout
	LeaseTTL    int64         // Lease TTL in seconds
}

// DefaultEtcdConfig returns default etcd configuration.
func DefaultEtcdConfig() EtcdConfig {
	return EtcdConfig{
		Endpoints:   []string{"localhost:2379"},
		Prefix:      "/services",
		HTTPTimeout: 30 * time.Second,
		LeaseTTL:    30,
	}
}

// EtcdClient implements ServiceDiscovery using etcd.
type EtcdClient struct {
	config     EtcdConfig
	httpClient *http.Client
	leases     map[string]int64 // serviceID -> leaseID
	watchers   map[string]context.CancelFunc
	mu         sync.RWMutex
}

// NewEtcdClient creates a new etcd client.
func NewEtcdClient(config EtcdConfig) *EtcdClient {
	if len(config.Endpoints) == 0 {
		config.Endpoints = []string{"localhost:2379"}
	}
	if config.Prefix == "" {
		config.Prefix = "/services"
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 30 * time.Second
	}
	if config.LeaseTTL == 0 {
		config.LeaseTTL = 30
	}

	return &EtcdClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		leases:   make(map[string]int64),
		watchers: make(map[string]context.CancelFunc),
	}
}

// baseURL returns the base URL for etcd API calls.
func (c *EtcdClient) baseURL() string {
	return fmt.Sprintf("http://%s", c.config.Endpoints[0])
}

// doRequest performs an HTTP request to etcd.
func (c *EtcdClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL() + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add basic auth if credentials provided
	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	return c.httpClient.Do(req)
}

// serviceKey returns the etcd key for a service instance.
func (c *EtcdClient) serviceKey(serviceName, instanceID string) string {
	return path.Join(c.config.Prefix, serviceName, instanceID)
}

// servicePrefixKey returns the etcd prefix key for a service.
func (c *EtcdClient) servicePrefixKey(serviceName string) string {
	return path.Join(c.config.Prefix, serviceName) + "/"
}

// grantLease grants a new lease from etcd.
func (c *EtcdClient) grantLease(ctx context.Context) (int64, error) {
	req := map[string]interface{}{
		"TTL": c.config.LeaseTTL,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/lease/grant", req)
	if err != nil {
		return 0, fmt.Errorf("failed to grant lease: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ID int64 `json:"ID,string"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode lease response: %w", err)
	}

	return result.ID, nil
}

// keepAliveLease keeps a lease alive.
func (c *EtcdClient) keepAliveLease(ctx context.Context, leaseID int64) error {
	req := map[string]interface{}{
		"ID": leaseID,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/lease/keepalive", req)
	if err != nil {
		return fmt.Errorf("failed to keep alive lease: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// revokeLease revokes a lease.
func (c *EtcdClient) revokeLease(ctx context.Context, leaseID int64) error {
	req := map[string]interface{}{
		"ID": leaseID,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/lease/revoke", req)
	if err != nil {
		return fmt.Errorf("failed to revoke lease: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Register registers a service instance with etcd.
func (c *EtcdClient) Register(ctx context.Context, registration *ServiceRegistration) error {
	// Grant a lease
	leaseID, err := c.grantLease(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRegistrationFailed, err)
	}

	// Create service instance
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

	// Encode instance data
	data, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	// Put key with lease
	key := c.serviceKey(registration.Name, registration.ID)
	req := map[string]interface{}{
		"key":   encodeBase64(key),
		"value": encodeBase64(string(data)),
		"lease": leaseID,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/kv/put", req)
	if err != nil {
		// Cleanup lease on failure
		c.revokeLease(ctx, leaseID)
		return fmt.Errorf("%w: %v", ErrRegistrationFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.revokeLease(ctx, leaseID)
		return fmt.Errorf("%w: status %d: %s", ErrRegistrationFailed, resp.StatusCode, string(body))
	}

	// Store lease ID for keepalive
	c.mu.Lock()
	c.leases[registration.ID] = leaseID
	c.mu.Unlock()

	// Start keepalive goroutine
	go c.startKeepAlive(ctx, registration.ID, leaseID)

	return nil
}

// startKeepAlive starts the lease keepalive loop.
func (c *EtcdClient) startKeepAlive(ctx context.Context, serviceID string, leaseID int64) {
	ticker := time.NewTicker(time.Duration(c.config.LeaseTTL/3) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			currentLeaseID, ok := c.leases[serviceID]
			c.mu.RUnlock()

			if !ok || currentLeaseID != leaseID {
				return // Lease no longer valid
			}

			if err := c.keepAliveLease(ctx, leaseID); err != nil {
				fmt.Printf("Keepalive failed for %s: %v\n", serviceID, err)
			}
		}
	}
}

// Deregister removes a service instance from etcd.
func (c *EtcdClient) Deregister(ctx context.Context, serviceID string) error {
	c.mu.Lock()
	leaseID, ok := c.leases[serviceID]
	if ok {
		delete(c.leases, serviceID)
	}
	c.mu.Unlock()

	// Revoke lease (which deletes associated keys)
	if ok {
		if err := c.revokeLease(ctx, leaseID); err != nil {
			return fmt.Errorf("%w: %v", ErrDeregistrationFailed, err)
		}
	}

	return nil
}

// GetService retrieves all instances of a service from etcd.
func (c *EtcdClient) GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	prefix := c.servicePrefixKey(serviceName)
	req := map[string]interface{}{
		"key":        encodeBase64(prefix),
		"range_end":  encodeBase64(getPrefix(prefix)),
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/kv/range", req)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Kvs []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"kvs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Kvs) == 0 {
		return nil, ErrServiceNotFound
	}

	instances := make([]*ServiceInstance, 0, len(result.Kvs))
	for _, kv := range result.Kvs {
		value := decodeBase64(kv.Value)

		var instance ServiceInstance
		if err := json.Unmarshal([]byte(value), &instance); err != nil {
			continue // Skip invalid entries
		}

		instances = append(instances, &instance)
	}

	return instances, nil
}

// GetHealthyService retrieves only healthy instances of a service.
func (c *EtcdClient) GetHealthyService(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	instances, err := c.GetService(ctx, serviceName)
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
func (c *EtcdClient) GetServiceInstance(ctx context.Context, serviceID string) (*ServiceInstance, error) {
	// Need to search through all services
	services, err := c.ListServices(ctx)
	if err != nil {
		return nil, err
	}

	for _, serviceName := range services {
		key := c.serviceKey(serviceName, serviceID)
		req := map[string]interface{}{
			"key": encodeBase64(key),
		}

		resp, err := c.doRequest(ctx, http.MethodPost, "/v3/kv/range", req)
		if err != nil {
			continue
		}

		var result struct {
			Kvs []struct {
				Value string `json:"value"`
			} `json:"kvs"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		if len(result.Kvs) > 0 {
			value := decodeBase64(result.Kvs[0].Value)
			var instance ServiceInstance
			if err := json.Unmarshal([]byte(value), &instance); err != nil {
				continue
			}
			return &instance, nil
		}
	}

	return nil, ErrServiceNotFound
}

// ListServices lists all registered services from etcd.
func (c *EtcdClient) ListServices(ctx context.Context) ([]string, error) {
	prefix := c.config.Prefix + "/"
	req := map[string]interface{}{
		"key":        encodeBase64(prefix),
		"range_end":  encodeBase64(getPrefix(prefix)),
		"keys_only":  true,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v3/kv/range", req)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Kvs []struct {
			Key string `json:"key"`
		} `json:"kvs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract unique service names
	serviceMap := make(map[string]bool)
	for _, kv := range result.Kvs {
		key := decodeBase64(kv.Key)
		// Key format: /services/{serviceName}/{instanceID}
		parts := splitPath(key)
		if len(parts) >= 2 {
			serviceName := parts[1]
			serviceMap[serviceName] = true
		}
	}

	services := make([]string, 0, len(serviceMap))
	for name := range serviceMap {
		services = append(services, name)
	}

	return services, nil
}

// Watch watches for changes to a service.
func (c *EtcdClient) Watch(ctx context.Context, serviceName string, callback func([]*ServiceInstance)) error {
	watchCtx, cancel := context.WithCancel(ctx)

	c.mu.Lock()
	c.watchers[serviceName] = cancel
	c.mu.Unlock()

	go func() {
		prefix := c.servicePrefixKey(serviceName)

		// Create watch request
		req := map[string]interface{}{
			"create_request": map[string]interface{}{
				"key":       encodeBase64(prefix),
				"range_end": encodeBase64(getPrefix(prefix)),
			},
		}

		for {
			select {
			case <-watchCtx.Done():
				return
			default:
				resp, err := c.doRequest(watchCtx, http.MethodPost, "/v3/watch", req)
				if err != nil {
					time.Sleep(5 * time.Second)
					continue
				}

				// Read watch events
				decoder := json.NewDecoder(resp.Body)
				for {
					var watchResp struct {
						Result struct {
							Events []struct {
								Type string `json:"type"`
								Kv   struct {
									Key   string `json:"key"`
									Value string `json:"value"`
								} `json:"kv"`
							} `json:"events"`
						} `json:"result"`
					}

					if err := decoder.Decode(&watchResp); err != nil {
						break
					}

					// Fetch current instances on any change
					instances, _ := c.GetService(watchCtx, serviceName)
					callback(instances)
				}
				resp.Body.Close()
			}
		}
	}()

	// Send initial state
	instances, _ := c.GetService(ctx, serviceName)
	callback(instances)

	return nil
}

// Heartbeat updates the last heartbeat time for a service instance.
func (c *EtcdClient) Heartbeat(ctx context.Context, serviceID string) error {
	c.mu.RLock()
	leaseID, ok := c.leases[serviceID]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no lease found for service: %s", serviceID)
	}

	return c.keepAliveLease(ctx, leaseID)
}

// Close closes the etcd client and stops all watchers.
func (c *EtcdClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel all watchers
	for _, cancel := range c.watchers {
		cancel()
	}
	c.watchers = make(map[string]context.CancelFunc)

	// Revoke all leases
	for serviceID, leaseID := range c.leases {
		c.revokeLease(context.Background(), leaseID)
		delete(c.leases, serviceID)
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// encodeBase64 encodes a string to base64.
func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// decodeBase64 decodes a base64 string.
func decodeBase64(s string) string {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s
	}
	return string(data)
}

// getPrefix returns the prefix end key for range queries.
func getPrefix(prefix string) string {
	end := []byte(prefix)
	for i := len(end) - 1; i >= 0; i-- {
		if end[i] < 0xff {
			end[i]++
			end = end[:i+1]
			break
		}
	}
	return string(end)
}

// splitPath splits a path into components.
func splitPath(p string) []string {
	parts := strings.Split(p, "/")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// ============================================================================
// etcd Service Registry with Watcher
// ============================================================================

// EtcdServiceRegistry provides high-level service registry functionality.
type EtcdServiceRegistry struct {
	client         *EtcdClient
	instances      map[string]map[string]*ServiceInstance // serviceName -> instanceID -> instance
	callbacks      map[string][]func([]*ServiceInstance)
	mu             sync.RWMutex
}

// NewEtcdServiceRegistry creates a new service registry.
func NewEtcdServiceRegistry(client *EtcdClient) *EtcdServiceRegistry {
	return &EtcdServiceRegistry{
		client:    client,
		instances: make(map[string]map[string]*ServiceInstance),
		callbacks: make(map[string][]func([]*ServiceInstance)),
	}
}

// Subscribe subscribes to service changes.
func (r *EtcdServiceRegistry) Subscribe(serviceName string, callback func([]*ServiceInstance)) error {
	r.mu.Lock()
	r.callbacks[serviceName] = append(r.callbacks[serviceName], callback)
	r.mu.Unlock()

	return r.client.Watch(context.Background(), serviceName, func(instances []*ServiceInstance) {
		r.mu.Lock()
		if r.instances[serviceName] == nil {
			r.instances[serviceName] = make(map[string]*ServiceInstance)
		}

		// Update local cache
		newCache := make(map[string]*ServiceInstance)
		for _, inst := range instances {
			newCache[inst.ID] = inst
		}
		r.instances[serviceName] = newCache

		// Notify callbacks
		callbacks := r.callbacks[serviceName]
		r.mu.Unlock()

		for _, cb := range callbacks {
			cb(instances)
		}
	})
}

// GetInstances returns cached instances for a service.
func (r *EtcdServiceRegistry) GetInstances(serviceName string) []*ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instanceMap := r.instances[serviceName]
	if instanceMap == nil {
		return nil
	}

	instances := make([]*ServiceInstance, 0, len(instanceMap))
	for _, inst := range instanceMap {
		instances = append(instances, inst)
	}

	return instances
}
