// Package discovery provides service discovery abstractions for the CRM application.
package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ============================================================================
// Consul Client
// ============================================================================

// ConsulConfig configures the Consul client.
type ConsulConfig struct {
	Address     string        // Consul agent address (default: localhost:8500)
	Scheme      string        // http or https
	Token       string        // ACL token
	Datacenter  string        // Datacenter name
	HTTPTimeout time.Duration // HTTP client timeout
	WaitTime    time.Duration // Long-poll wait time for watches
}

// DefaultConsulConfig returns default Consul configuration.
func DefaultConsulConfig() ConsulConfig {
	return ConsulConfig{
		Address:     "localhost:8500",
		Scheme:      "http",
		HTTPTimeout: 30 * time.Second,
		WaitTime:    5 * time.Minute,
	}
}

// ConsulClient implements ServiceDiscovery using Consul.
type ConsulClient struct {
	config     ConsulConfig
	httpClient *http.Client
	watchers   map[string]context.CancelFunc
	mu         sync.RWMutex
}

// NewConsulClient creates a new Consul client.
func NewConsulClient(config ConsulConfig) *ConsulClient {
	if config.Address == "" {
		config.Address = "localhost:8500"
	}
	if config.Scheme == "" {
		config.Scheme = "http"
	}
	if config.HTTPTimeout == 0 {
		config.HTTPTimeout = 30 * time.Second
	}

	return &ConsulClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		watchers: make(map[string]context.CancelFunc),
	}
}

// baseURL returns the base URL for Consul API calls.
func (c *ConsulClient) baseURL() string {
	return fmt.Sprintf("%s://%s", c.config.Scheme, c.config.Address)
}

// doRequest performs an HTTP request to Consul.
func (c *ConsulClient) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
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
	if c.config.Token != "" {
		req.Header.Set("X-Consul-Token", c.config.Token)
	}

	return c.httpClient.Do(req)
}

// Register registers a service instance with Consul.
func (c *ConsulClient) Register(ctx context.Context, registration *ServiceRegistration) error {
	consulReg := &consulServiceRegistration{
		ID:      registration.ID,
		Name:    registration.Name,
		Address: registration.Host,
		Port:    registration.Port,
		Tags:    registration.Tags,
		Meta:    registration.Metadata,
	}

	if registration.HealthCheck != nil {
		check := &consulHealthCheck{
			DeregisterCriticalServiceAfter: registration.HealthCheck.DeregisterAfter.String(),
			Interval:                       registration.HealthCheck.Interval.String(),
			Timeout:                        registration.HealthCheck.Timeout.String(),
		}

		switch registration.HealthCheck.Type {
		case "http":
			check.HTTP = fmt.Sprintf("%s://%s:%d%s",
				registration.Protocol, registration.Host, registration.Port, registration.HealthCheck.Endpoint)
		case "tcp":
			check.TCP = fmt.Sprintf("%s:%d", registration.Host, registration.Port)
		case "grpc":
			check.GRPC = fmt.Sprintf("%s:%d", registration.Host, registration.Port)
		case "ttl":
			check.TTL = registration.TTL.String()
		}

		consulReg.Check = check
	}

	resp, err := c.doRequest(ctx, http.MethodPut, "/v1/agent/service/register", consulReg)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRegistrationFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d: %s", ErrRegistrationFailed, resp.StatusCode, string(body))
	}

	return nil
}

// Deregister removes a service instance from Consul.
func (c *ConsulClient) Deregister(ctx context.Context, serviceID string) error {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/v1/agent/service/deregister/%s", serviceID), nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDeregistrationFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d: %s", ErrDeregistrationFailed, resp.StatusCode, string(body))
	}

	return nil
}

// GetService retrieves all instances of a service from Consul.
func (c *ConsulClient) GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	path := fmt.Sprintf("/v1/catalog/service/%s", serviceName)
	if c.config.Datacenter != "" {
		path += fmt.Sprintf("?dc=%s", c.config.Datacenter)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrServiceNotFound
	}

	var services []consulCatalogService
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	instances := make([]*ServiceInstance, 0, len(services))
	for _, svc := range services {
		instances = append(instances, &ServiceInstance{
			ID:          svc.ServiceID,
			ServiceName: svc.ServiceName,
			Host:        svc.ServiceAddress,
			Port:        svc.ServicePort,
			Metadata:    svc.ServiceMeta,
			Tags:        svc.ServiceTags,
			Health:      HealthStatusUnknown, // Will be updated by health check
		})
	}

	return instances, nil
}

// GetHealthyService retrieves only healthy instances of a service from Consul.
func (c *ConsulClient) GetHealthyService(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	path := fmt.Sprintf("/v1/health/service/%s?passing=true", serviceName)
	if c.config.Datacenter != "" {
		path += fmt.Sprintf("&dc=%s", c.config.Datacenter)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get healthy service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrServiceNotFound
	}

	var healthServices []consulHealthService
	if err := json.NewDecoder(resp.Body).Decode(&healthServices); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(healthServices) == 0 {
		return nil, ErrNoHealthyInstances
	}

	instances := make([]*ServiceInstance, 0, len(healthServices))
	for _, hs := range healthServices {
		host := hs.Service.Address
		if host == "" {
			host = hs.Node.Address
		}

		instances = append(instances, &ServiceInstance{
			ID:          hs.Service.ID,
			ServiceName: hs.Service.Service,
			Host:        host,
			Port:        hs.Service.Port,
			Metadata:    hs.Service.Meta,
			Tags:        hs.Service.Tags,
			Health:      HealthStatusHealthy,
			Weight:      getWeight(hs.Service.Weights),
		})
	}

	return instances, nil
}

// getWeight extracts weight from Consul service weights.
func getWeight(weights *consulServiceWeights) int {
	if weights == nil {
		return 1
	}
	return weights.Passing
}

// GetServiceInstance retrieves a specific service instance.
func (c *ConsulClient) GetServiceInstance(ctx context.Context, serviceID string) (*ServiceInstance, error) {
	// Consul doesn't have a direct API for getting a specific instance
	// We need to get the agent services and filter
	resp, err := c.doRequest(ctx, http.MethodGet, "/v1/agent/services", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent services: %w", err)
	}
	defer resp.Body.Close()

	var services map[string]consulAgentService
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	svc, ok := services[serviceID]
	if !ok {
		return nil, ErrServiceNotFound
	}

	return &ServiceInstance{
		ID:          svc.ID,
		ServiceName: svc.Service,
		Host:        svc.Address,
		Port:        svc.Port,
		Metadata:    svc.Meta,
		Tags:        svc.Tags,
		Health:      HealthStatusUnknown,
	}, nil
}

// ListServices lists all registered services from Consul.
func (c *ConsulClient) ListServices(ctx context.Context) ([]string, error) {
	path := "/v1/catalog/services"
	if c.config.Datacenter != "" {
		path += fmt.Sprintf("?dc=%s", c.config.Datacenter)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer resp.Body.Close()

	var services map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&services); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}

	return names, nil
}

// Watch watches for changes to a service.
func (c *ConsulClient) Watch(ctx context.Context, serviceName string, callback func([]*ServiceInstance)) error {
	watchCtx, cancel := context.WithCancel(ctx)

	c.mu.Lock()
	c.watchers[serviceName] = cancel
	c.mu.Unlock()

	go func() {
		var lastIndex uint64 = 0

		for {
			select {
			case <-watchCtx.Done():
				return
			default:
				instances, index, err := c.watchService(watchCtx, serviceName, lastIndex)
				if err != nil {
					// Sleep and retry on error
					time.Sleep(5 * time.Second)
					continue
				}

				if index > lastIndex {
					lastIndex = index
					callback(instances)
				}
			}
		}
	}()

	return nil
}

// watchService performs a blocking query for service changes.
func (c *ConsulClient) watchService(ctx context.Context, serviceName string, lastIndex uint64) ([]*ServiceInstance, uint64, error) {
	path := fmt.Sprintf("/v1/health/service/%s?passing=true&index=%d&wait=%s",
		serviceName, lastIndex, c.config.WaitTime.String())

	if c.config.Datacenter != "" {
		path += fmt.Sprintf("&dc=%s", c.config.Datacenter)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, lastIndex, err
	}
	defer resp.Body.Close()

	// Get new index from header
	var newIndex uint64
	if indexStr := resp.Header.Get("X-Consul-Index"); indexStr != "" {
		fmt.Sscanf(indexStr, "%d", &newIndex)
	}

	var healthServices []consulHealthService
	if err := json.NewDecoder(resp.Body).Decode(&healthServices); err != nil {
		return nil, lastIndex, err
	}

	instances := make([]*ServiceInstance, 0, len(healthServices))
	for _, hs := range healthServices {
		host := hs.Service.Address
		if host == "" {
			host = hs.Node.Address
		}

		instances = append(instances, &ServiceInstance{
			ID:          hs.Service.ID,
			ServiceName: hs.Service.Service,
			Host:        host,
			Port:        hs.Service.Port,
			Metadata:    hs.Service.Meta,
			Tags:        hs.Service.Tags,
			Health:      HealthStatusHealthy,
			Weight:      getWeight(hs.Service.Weights),
		})
	}

	return instances, newIndex, nil
}

// Heartbeat sends a TTL check pass for a service instance.
func (c *ConsulClient) Heartbeat(ctx context.Context, serviceID string) error {
	checkID := fmt.Sprintf("service:%s", serviceID)
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/v1/agent/check/pass/%s", checkID), nil)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed: status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close closes the Consul client and stops all watchers.
func (c *ConsulClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cancel := range c.watchers {
		cancel()
	}
	c.watchers = make(map[string]context.CancelFunc)

	return nil
}

// ============================================================================
// Consul API Structures
// ============================================================================

type consulServiceRegistration struct {
	ID      string                 `json:"ID"`
	Name    string                 `json:"Name"`
	Address string                 `json:"Address"`
	Port    int                    `json:"Port"`
	Tags    []string               `json:"Tags,omitempty"`
	Meta    map[string]string      `json:"Meta,omitempty"`
	Check   *consulHealthCheck     `json:"Check,omitempty"`
	Weights *consulServiceWeights  `json:"Weights,omitempty"`
}

type consulHealthCheck struct {
	HTTP                           string `json:"HTTP,omitempty"`
	TCP                            string `json:"TCP,omitempty"`
	GRPC                           string `json:"GRPC,omitempty"`
	TTL                            string `json:"TTL,omitempty"`
	Interval                       string `json:"Interval,omitempty"`
	Timeout                        string `json:"Timeout,omitempty"`
	DeregisterCriticalServiceAfter string `json:"DeregisterCriticalServiceAfter,omitempty"`
}

type consulServiceWeights struct {
	Passing int `json:"Passing"`
	Warning int `json:"Warning"`
}

type consulCatalogService struct {
	ServiceID      string            `json:"ServiceID"`
	ServiceName    string            `json:"ServiceName"`
	ServiceAddress string            `json:"ServiceAddress"`
	ServicePort    int               `json:"ServicePort"`
	ServiceMeta    map[string]string `json:"ServiceMeta"`
	ServiceTags    []string          `json:"ServiceTags"`
}

type consulHealthService struct {
	Node    consulNode    `json:"Node"`
	Service consulService `json:"Service"`
}

type consulNode struct {
	Address string `json:"Address"`
}

type consulService struct {
	ID      string                `json:"ID"`
	Service string                `json:"Service"`
	Address string                `json:"Address"`
	Port    int                   `json:"Port"`
	Meta    map[string]string     `json:"Meta"`
	Tags    []string              `json:"Tags"`
	Weights *consulServiceWeights `json:"Weights"`
}

type consulAgentService struct {
	ID      string            `json:"ID"`
	Service string            `json:"Service"`
	Address string            `json:"Address"`
	Port    int               `json:"Port"`
	Meta    map[string]string `json:"Meta"`
	Tags    []string          `json:"Tags"`
}

// ============================================================================
// Consul Health Checker
// ============================================================================

// ConsulHealthChecker provides health check utilities for Consul.
type ConsulHealthChecker struct {
	client *ConsulClient
}

// NewConsulHealthChecker creates a new health checker.
func NewConsulHealthChecker(client *ConsulClient) *ConsulHealthChecker {
	return &ConsulHealthChecker{client: client}
}

// StartTTLHeartbeat starts periodic TTL heartbeats for a service.
func (h *ConsulHealthChecker) StartTTLHeartbeat(ctx context.Context, serviceID string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := h.client.Heartbeat(ctx, serviceID); err != nil {
					fmt.Printf("Heartbeat failed for %s: %v\n", serviceID, err)
				}
			}
		}
	}()
}
