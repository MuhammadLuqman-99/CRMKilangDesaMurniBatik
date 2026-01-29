// Package gateway provides API gateway functionality for the CRM application.
package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Request Aggregation
// ============================================================================

// AggregatorConfig holds configuration for request aggregation.
type AggregatorConfig struct {
	Timeout         time.Duration
	MaxConcurrent   int
	EnableCaching   bool
	CacheTTL        time.Duration
	RetryAttempts   int
	RetryDelay      time.Duration
}

// DefaultAggregatorConfig returns default aggregator configuration.
func DefaultAggregatorConfig() AggregatorConfig {
	return AggregatorConfig{
		Timeout:       30 * time.Second,
		MaxConcurrent: 10,
		EnableCaching: true,
		CacheTTL:      5 * time.Minute,
		RetryAttempts: 2,
		RetryDelay:    100 * time.Millisecond,
	}
}

// ServiceRequest represents a request to a backend service.
type ServiceRequest struct {
	ID          string            `json:"id"`
	Service     string            `json:"service"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
	Body        json.RawMessage   `json:"body,omitempty"`
	DependsOn   []string          `json:"depends_on,omitempty"`
	Required    bool              `json:"required"`
}

// ServiceResponse represents a response from a backend service.
type ServiceResponse struct {
	ID         string            `json:"id"`
	Service    string            `json:"service"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       json.RawMessage   `json:"body,omitempty"`
	Error      string            `json:"error,omitempty"`
	Duration   time.Duration     `json:"duration_ms"`
	FromCache  bool              `json:"from_cache"`
}

// AggregatedRequest represents a batch of requests to aggregate.
type AggregatedRequest struct {
	RequestID string           `json:"request_id"`
	Requests  []ServiceRequest `json:"requests"`
}

// AggregatedResponse represents the aggregated response.
type AggregatedResponse struct {
	RequestID  string            `json:"request_id"`
	Responses  []ServiceResponse `json:"responses"`
	TotalTime  time.Duration     `json:"total_time_ms"`
	HasErrors  bool              `json:"has_errors"`
	ErrorCount int               `json:"error_count"`
}

// RequestAggregator aggregates multiple service requests.
type RequestAggregator struct {
	config     AggregatorConfig
	services   map[string]ServiceEndpoint
	httpClient *http.Client
	cache      Cache
	mu         sync.RWMutex
}

// ServiceEndpoint represents a backend service endpoint.
type ServiceEndpoint struct {
	Name      string
	BaseURL   string
	Timeout   time.Duration
	Headers   map[string]string
	HealthURL string
	Healthy   bool
}

// Cache interface for response caching.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// NewRequestAggregator creates a new request aggregator.
func NewRequestAggregator(config AggregatorConfig, cache Cache) *RequestAggregator {
	return &RequestAggregator{
		config:   config,
		services: make(map[string]ServiceEndpoint),
		httpClient: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		cache: cache,
	}
}

// RegisterService registers a backend service.
func (a *RequestAggregator) RegisterService(service ServiceEndpoint) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.services[service.Name] = service
}

// UnregisterService unregisters a backend service.
func (a *RequestAggregator) UnregisterService(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.services, name)
}

// GetService returns a service endpoint.
func (a *RequestAggregator) GetService(name string) (ServiceEndpoint, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	service, ok := a.services[name]
	return service, ok
}

// Aggregate executes multiple service requests and aggregates the responses.
func (a *RequestAggregator) Aggregate(ctx context.Context, req AggregatedRequest) (*AggregatedResponse, error) {
	startTime := time.Now()

	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// Build dependency graph
	graph := a.buildDependencyGraph(req.Requests)

	// Execute requests respecting dependencies
	responses := a.executeWithDependencies(ctx, req.Requests, graph)

	// Count errors
	errorCount := 0
	for _, resp := range responses {
		if resp.Error != "" || resp.StatusCode >= 400 {
			errorCount++
		}
	}

	return &AggregatedResponse{
		RequestID:  req.RequestID,
		Responses:  responses,
		TotalTime:  time.Since(startTime),
		HasErrors:  errorCount > 0,
		ErrorCount: errorCount,
	}, nil
}

// dependencyNode represents a node in the dependency graph.
type dependencyNode struct {
	request     ServiceRequest
	dependencies []*dependencyNode
	dependents   []*dependencyNode
	executed    bool
	response    *ServiceResponse
}

// buildDependencyGraph builds a dependency graph from requests.
func (a *RequestAggregator) buildDependencyGraph(requests []ServiceRequest) map[string]*dependencyNode {
	nodes := make(map[string]*dependencyNode)

	// Create nodes
	for _, req := range requests {
		nodes[req.ID] = &dependencyNode{
			request:     req,
			dependencies: make([]*dependencyNode, 0),
			dependents:   make([]*dependencyNode, 0),
		}
	}

	// Link dependencies
	for _, req := range requests {
		node := nodes[req.ID]
		for _, depID := range req.DependsOn {
			if depNode, ok := nodes[depID]; ok {
				node.dependencies = append(node.dependencies, depNode)
				depNode.dependents = append(depNode.dependents, node)
			}
		}
	}

	return nodes
}

// executeWithDependencies executes requests respecting their dependencies.
func (a *RequestAggregator) executeWithDependencies(ctx context.Context, requests []ServiceRequest, graph map[string]*dependencyNode) []ServiceResponse {
	responses := make([]ServiceResponse, 0, len(requests))
	responseMu := sync.Mutex{}

	// Find nodes with no dependencies (root nodes)
	var rootNodes []*dependencyNode
	for _, node := range graph {
		if len(node.dependencies) == 0 {
			rootNodes = append(rootNodes, node)
		}
	}

	// Use semaphore to limit concurrency
	sem := make(chan struct{}, a.config.MaxConcurrent)
	var wg sync.WaitGroup

	// Execute root nodes
	for _, node := range rootNodes {
		wg.Add(1)
		go func(n *dependencyNode) {
			defer wg.Done()
			a.executeNode(ctx, n, sem, &responses, &responseMu, graph)
		}(node)
	}

	wg.Wait()

	return responses
}

// executeNode executes a single node and its dependents.
func (a *RequestAggregator) executeNode(ctx context.Context, node *dependencyNode, sem chan struct{}, responses *[]ServiceResponse, mu *sync.Mutex, graph map[string]*dependencyNode) {
	// Wait for all dependencies
	for _, dep := range node.dependencies {
		for !dep.executed {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Millisecond):
			}
		}
	}

	// Acquire semaphore
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		return
	}

	// Execute request
	resp := a.executeRequest(ctx, node.request)

	// Store response
	mu.Lock()
	*responses = append(*responses, resp)
	node.response = &resp
	node.executed = true
	mu.Unlock()

	// Execute dependents
	var wg sync.WaitGroup
	for _, dep := range node.dependents {
		// Check if all dependencies of the dependent are satisfied
		allDone := true
		for _, d := range dep.dependencies {
			if !d.executed {
				allDone = false
				break
			}
		}
		if allDone && !dep.executed {
			wg.Add(1)
			go func(n *dependencyNode) {
				defer wg.Done()
				a.executeNode(ctx, n, sem, responses, mu, graph)
			}(dep)
		}
	}
	wg.Wait()
}

// executeRequest executes a single service request.
func (a *RequestAggregator) executeRequest(ctx context.Context, req ServiceRequest) ServiceResponse {
	startTime := time.Now()
	response := ServiceResponse{
		ID:      req.ID,
		Service: req.Service,
	}

	// Check cache first
	if a.config.EnableCaching && req.Method == "GET" && a.cache != nil {
		cacheKey := a.getCacheKey(req)
		if data, ok := a.cache.Get(ctx, cacheKey); ok {
			response.StatusCode = http.StatusOK
			response.Body = data
			response.FromCache = true
			response.Duration = time.Since(startTime)
			return response
		}
	}

	// Get service endpoint
	service, ok := a.GetService(req.Service)
	if !ok {
		response.StatusCode = http.StatusBadGateway
		response.Error = fmt.Sprintf("service not found: %s", req.Service)
		response.Duration = time.Since(startTime)
		return response
	}

	// Build URL
	url := service.BaseURL + req.Path
	if len(req.QueryParams) > 0 {
		url += "?"
		first := true
		for k, v := range req.QueryParams {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
	}

	// Create request
	var body io.Reader
	if req.Body != nil {
		body = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.Error = fmt.Sprintf("failed to create request: %v", err)
		response.Duration = time.Since(startTime)
		return response
	}

	// Add headers
	for k, v := range service.Headers {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Execute with retries
	var httpResp *http.Response
	for attempt := 0; attempt <= a.config.RetryAttempts; attempt++ {
		httpResp, err = a.httpClient.Do(httpReq)
		if err == nil {
			break
		}
		if attempt < a.config.RetryAttempts {
			time.Sleep(a.config.RetryDelay * time.Duration(attempt+1))
		}
	}

	if err != nil {
		response.StatusCode = http.StatusBadGateway
		response.Error = fmt.Sprintf("request failed: %v", err)
		response.Duration = time.Since(startTime)
		return response
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.Error = fmt.Sprintf("failed to read response: %v", err)
		response.Duration = time.Since(startTime)
		return response
	}

	response.StatusCode = httpResp.StatusCode
	response.Body = respBody
	response.Headers = make(map[string]string)
	for k := range httpResp.Header {
		response.Headers[k] = httpResp.Header.Get(k)
	}
	response.Duration = time.Since(startTime)

	// Cache successful GET responses
	if a.config.EnableCaching && req.Method == "GET" && httpResp.StatusCode == http.StatusOK && a.cache != nil {
		cacheKey := a.getCacheKey(req)
		a.cache.Set(ctx, cacheKey, respBody, a.config.CacheTTL)
	}

	return response
}

// getCacheKey generates a cache key for a request.
func (a *RequestAggregator) getCacheKey(req ServiceRequest) string {
	return fmt.Sprintf("agg:%s:%s:%s", req.Service, req.Method, req.Path)
}

// ============================================================================
// Aggregation Handler
// ============================================================================

// AggregationHandler handles HTTP requests for request aggregation.
type AggregationHandler struct {
	aggregator *RequestAggregator
}

// NewAggregationHandler creates a new aggregation handler.
func NewAggregationHandler(aggregator *RequestAggregator) *AggregationHandler {
	return &AggregationHandler{aggregator: aggregator}
}

// ServeHTTP handles aggregation requests.
func (h *AggregationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AggregatedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	resp, err := h.aggregator.Aggregate(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Aggregation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Set appropriate status code
	if resp.HasErrors {
		// Check if any required request failed
		for _, svcResp := range resp.Responses {
			for _, svcReq := range req.Requests {
				if svcReq.ID == svcResp.ID && svcReq.Required && (svcResp.Error != "" || svcResp.StatusCode >= 400) {
					w.WriteHeader(http.StatusBadGateway)
					json.NewEncoder(w).Encode(resp)
					return
				}
			}
		}
		w.WriteHeader(http.StatusMultiStatus)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(resp)
}

// ============================================================================
// GraphQL-like Query Aggregation
// ============================================================================

// QueryField represents a field in an aggregation query.
type QueryField struct {
	Name      string            `json:"name"`
	Service   string            `json:"service"`
	Path      string            `json:"path"`
	Params    map[string]string `json:"params,omitempty"`
	Fields    []QueryField      `json:"fields,omitempty"`
	Transform string            `json:"transform,omitempty"`
}

// QueryRequest represents an aggregation query.
type QueryRequest struct {
	Fields []QueryField `json:"fields"`
}

// QueryResponse represents the query result.
type QueryResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []QueryError               `json:"errors,omitempty"`
}

// QueryError represents an error in query execution.
type QueryError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ExecuteQuery executes an aggregation query.
func (a *RequestAggregator) ExecuteQuery(ctx context.Context, query QueryRequest) (*QueryResponse, error) {
	// Convert query fields to service requests
	requests := make([]ServiceRequest, 0)
	for i, field := range query.Fields {
		req := ServiceRequest{
			ID:          fmt.Sprintf("field_%d", i),
			Service:     field.Service,
			Method:      "GET",
			Path:        field.Path,
			QueryParams: field.Params,
		}
		requests = append(requests, req)
	}

	// Execute aggregated request
	aggReq := AggregatedRequest{
		RequestID: uuid.New().String(),
		Requests:  requests,
	}

	aggResp, err := a.Aggregate(ctx, aggReq)
	if err != nil {
		return nil, err
	}

	// Build query response
	response := &QueryResponse{
		Data:   make(map[string]json.RawMessage),
		Errors: make([]QueryError, 0),
	}

	for i, svcResp := range aggResp.Responses {
		fieldName := query.Fields[i].Name
		if svcResp.Error != "" || svcResp.StatusCode >= 400 {
			response.Errors = append(response.Errors, QueryError{
				Field:   fieldName,
				Message: svcResp.Error,
			})
		} else {
			response.Data[fieldName] = svcResp.Body
		}
	}

	return response, nil
}

// ============================================================================
// In-Memory Cache Implementation
// ============================================================================

// MemoryCache provides an in-memory cache implementation.
type MemoryCache struct {
	data map[string]cacheEntry
	mu   sync.RWMutex
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		data: make(map[string]cacheEntry),
	}
	go cache.cleanup()
	return cache
}

// Get retrieves a value from cache.
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.value, true
}

// Set stores a value in cache.
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// Delete removes a value from cache.
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// cleanup periodically removes expired entries.
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// Ensure MemoryCache implements Cache
var _ Cache = (*MemoryCache)(nil)
