// Package performance contains performance tests for the CRM system.
package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// ============================================================================
// Performance Test Configuration
// ============================================================================

// Config holds performance test configuration.
type Config struct {
	// Load test settings
	ConcurrentUsers   int           // Number of concurrent users
	RequestsPerUser   int           // Number of requests per user
	RampUpDuration    time.Duration // Time to ramp up to full load
	TestDuration      time.Duration // Total test duration
	ThinkTime         time.Duration // Pause between requests

	// Thresholds
	MaxResponseTime   time.Duration // Maximum acceptable response time
	MaxErrorRate      float64       // Maximum acceptable error rate (0.0-1.0)
	MinThroughput     float64       // Minimum requests per second

	// Memory settings
	MaxMemoryMB       int64         // Maximum memory usage in MB
	MaxGoroutines     int           // Maximum number of goroutines
}

// DefaultConfig returns default performance test configuration.
func DefaultConfig() *Config {
	return &Config{
		ConcurrentUsers:   100,
		RequestsPerUser:   50,
		RampUpDuration:    10 * time.Second,
		TestDuration:      60 * time.Second,
		ThinkTime:         100 * time.Millisecond,
		MaxResponseTime:   2 * time.Second,
		MaxErrorRate:      0.01, // 1%
		MinThroughput:     100,  // requests per second
		MaxMemoryMB:       512,
		MaxGoroutines:     10000,
	}
}

// LightConfig returns a lighter configuration for CI/CD environments.
func LightConfig() *Config {
	return &Config{
		ConcurrentUsers:   10,
		RequestsPerUser:   10,
		RampUpDuration:    2 * time.Second,
		TestDuration:      10 * time.Second,
		ThinkTime:         50 * time.Millisecond,
		MaxResponseTime:   5 * time.Second,
		MaxErrorRate:      0.05, // 5%
		MinThroughput:     10,   // requests per second
		MaxMemoryMB:       256,
		MaxGoroutines:     1000,
	}
}

// ============================================================================
// Metrics Collection
// ============================================================================

// Metrics collects performance metrics during tests.
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests     int64
	SuccessfulRequests int64
	FailedRequests    int64

	// Response times
	ResponseTimes     []time.Duration

	// Throughput
	StartTime         time.Time
	EndTime           time.Time

	// Memory metrics
	InitialMemory     runtime.MemStats
	PeakMemory        uint64
	FinalMemory       runtime.MemStats

	// Goroutine metrics
	InitialGoroutines int
	PeakGoroutines    int
	FinalGoroutines   int

	// Error tracking
	Errors            map[string]int
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	m := &Metrics{
		ResponseTimes: make([]time.Duration, 0, 10000),
		Errors:        make(map[string]int),
	}

	// Capture initial state
	runtime.ReadMemStats(&m.InitialMemory)
	m.InitialGoroutines = runtime.NumGoroutine()
	m.StartTime = time.Now()

	return m
}

// RecordRequest records a request result.
func (m *Metrics) RecordRequest(duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.TotalRequests, 1)

	if err != nil {
		atomic.AddInt64(&m.FailedRequests, 1)
		m.Errors[err.Error()]++
	} else {
		atomic.AddInt64(&m.SuccessfulRequests, 1)
	}

	m.ResponseTimes = append(m.ResponseTimes, duration)
}

// UpdatePeaks updates peak memory and goroutine counts.
func (m *Metrics) UpdatePeaks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	if memStats.Alloc > m.PeakMemory {
		m.PeakMemory = memStats.Alloc
	}

	goroutines := runtime.NumGoroutine()
	if goroutines > m.PeakGoroutines {
		m.PeakGoroutines = goroutines
	}
}

// Finalize captures final state.
func (m *Metrics) Finalize() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.EndTime = time.Now()
	runtime.ReadMemStats(&m.FinalMemory)
	m.FinalGoroutines = runtime.NumGoroutine()
}

// ============================================================================
// Performance Report
// ============================================================================

// Report contains the performance test report.
type Report struct {
	// Summary
	TestName          string
	Duration          time.Duration
	TotalRequests     int64
	SuccessfulRequests int64
	FailedRequests    int64

	// Response time statistics
	MinResponseTime   time.Duration
	MaxResponseTime   time.Duration
	AvgResponseTime   time.Duration
	P50ResponseTime   time.Duration
	P90ResponseTime   time.Duration
	P95ResponseTime   time.Duration
	P99ResponseTime   time.Duration

	// Throughput
	RequestsPerSecond float64

	// Error rate
	ErrorRate         float64

	// Memory statistics
	InitialMemoryMB   float64
	PeakMemoryMB      float64
	FinalMemoryMB     float64
	MemoryGrowthMB    float64

	// Goroutine statistics
	InitialGoroutines int
	PeakGoroutines    int
	FinalGoroutines   int
	GoroutineLeaks    int

	// Errors breakdown
	Errors            map[string]int

	// Pass/Fail
	Passed            bool
	FailureReasons    []string
}

// GenerateReport generates a report from metrics.
func GenerateReport(name string, metrics *Metrics, config *Config) *Report {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	report := &Report{
		TestName:           name,
		Duration:           metrics.EndTime.Sub(metrics.StartTime),
		TotalRequests:      metrics.TotalRequests,
		SuccessfulRequests: metrics.SuccessfulRequests,
		FailedRequests:     metrics.FailedRequests,
		Errors:             metrics.Errors,
		InitialGoroutines:  metrics.InitialGoroutines,
		PeakGoroutines:     metrics.PeakGoroutines,
		FinalGoroutines:    metrics.FinalGoroutines,
	}

	// Calculate response time statistics
	if len(metrics.ResponseTimes) > 0 {
		times := make([]time.Duration, len(metrics.ResponseTimes))
		copy(times, metrics.ResponseTimes)
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })

		report.MinResponseTime = times[0]
		report.MaxResponseTime = times[len(times)-1]
		report.AvgResponseTime = calculateAvg(times)
		report.P50ResponseTime = percentile(times, 50)
		report.P90ResponseTime = percentile(times, 90)
		report.P95ResponseTime = percentile(times, 95)
		report.P99ResponseTime = percentile(times, 99)
	}

	// Calculate throughput
	if report.Duration > 0 {
		report.RequestsPerSecond = float64(report.TotalRequests) / report.Duration.Seconds()
	}

	// Calculate error rate
	if report.TotalRequests > 0 {
		report.ErrorRate = float64(report.FailedRequests) / float64(report.TotalRequests)
	}

	// Memory statistics
	report.InitialMemoryMB = float64(metrics.InitialMemory.Alloc) / 1024 / 1024
	report.PeakMemoryMB = float64(metrics.PeakMemory) / 1024 / 1024
	report.FinalMemoryMB = float64(metrics.FinalMemory.Alloc) / 1024 / 1024
	report.MemoryGrowthMB = report.FinalMemoryMB - report.InitialMemoryMB

	// Goroutine leaks
	report.GoroutineLeaks = report.FinalGoroutines - report.InitialGoroutines

	// Validate against thresholds
	report.Passed = true

	if report.P95ResponseTime > config.MaxResponseTime {
		report.Passed = false
		report.FailureReasons = append(report.FailureReasons,
			fmt.Sprintf("P95 response time %.2fms exceeds threshold %.2fms",
				float64(report.P95ResponseTime.Milliseconds()),
				float64(config.MaxResponseTime.Milliseconds())))
	}

	if report.ErrorRate > config.MaxErrorRate {
		report.Passed = false
		report.FailureReasons = append(report.FailureReasons,
			fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%",
				report.ErrorRate*100, config.MaxErrorRate*100))
	}

	if report.RequestsPerSecond < config.MinThroughput {
		report.Passed = false
		report.FailureReasons = append(report.FailureReasons,
			fmt.Sprintf("Throughput %.2f req/s below threshold %.2f req/s",
				report.RequestsPerSecond, config.MinThroughput))
	}

	if report.PeakMemoryMB > float64(config.MaxMemoryMB) {
		report.Passed = false
		report.FailureReasons = append(report.FailureReasons,
			fmt.Sprintf("Peak memory %.2fMB exceeds threshold %dMB",
				report.PeakMemoryMB, config.MaxMemoryMB))
	}

	if report.PeakGoroutines > config.MaxGoroutines {
		report.Passed = false
		report.FailureReasons = append(report.FailureReasons,
			fmt.Sprintf("Peak goroutines %d exceeds threshold %d",
				report.PeakGoroutines, config.MaxGoroutines))
	}

	return report
}

// Print prints the report to stdout.
func (r *Report) Print() {
	fmt.Println("================================================================================")
	fmt.Printf("Performance Test Report: %s\n", r.TestName)
	fmt.Println("================================================================================")
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  Duration:           %v\n", r.Duration)
	fmt.Printf("  Total Requests:     %d\n", r.TotalRequests)
	fmt.Printf("  Successful:         %d\n", r.SuccessfulRequests)
	fmt.Printf("  Failed:             %d\n", r.FailedRequests)
	fmt.Printf("  Error Rate:         %.2f%%\n", r.ErrorRate*100)
	fmt.Printf("  Throughput:         %.2f req/s\n", r.RequestsPerSecond)
	fmt.Println()

	fmt.Println("Response Times:")
	fmt.Printf("  Min:                %v\n", r.MinResponseTime)
	fmt.Printf("  Max:                %v\n", r.MaxResponseTime)
	fmt.Printf("  Avg:                %v\n", r.AvgResponseTime)
	fmt.Printf("  P50 (Median):       %v\n", r.P50ResponseTime)
	fmt.Printf("  P90:                %v\n", r.P90ResponseTime)
	fmt.Printf("  P95:                %v\n", r.P95ResponseTime)
	fmt.Printf("  P99:                %v\n", r.P99ResponseTime)
	fmt.Println()

	fmt.Println("Memory:")
	fmt.Printf("  Initial:            %.2f MB\n", r.InitialMemoryMB)
	fmt.Printf("  Peak:               %.2f MB\n", r.PeakMemoryMB)
	fmt.Printf("  Final:              %.2f MB\n", r.FinalMemoryMB)
	fmt.Printf("  Growth:             %.2f MB\n", r.MemoryGrowthMB)
	fmt.Println()

	fmt.Println("Goroutines:")
	fmt.Printf("  Initial:            %d\n", r.InitialGoroutines)
	fmt.Printf("  Peak:               %d\n", r.PeakGoroutines)
	fmt.Printf("  Final:              %d\n", r.FinalGoroutines)
	fmt.Printf("  Potential Leaks:    %d\n", r.GoroutineLeaks)
	fmt.Println()

	if len(r.Errors) > 0 {
		fmt.Println("Errors:")
		for err, count := range r.Errors {
			fmt.Printf("  %s: %d\n", err, count)
		}
		fmt.Println()
	}

	fmt.Println("Result:")
	if r.Passed {
		fmt.Println("  PASSED")
	} else {
		fmt.Println("  FAILED")
		for _, reason := range r.FailureReasons {
			fmt.Printf("    - %s\n", reason)
		}
	}
	fmt.Println("================================================================================")
}

// ============================================================================
// Test Server Setup
// ============================================================================

// TestServer provides a mock server for performance testing.
type TestServer struct {
	Server     *httptest.Server
	Router     *chi.Mux
	DataStore  *InMemoryStore
}

// InMemoryStore provides fast in-memory storage for performance tests.
type InMemoryStore struct {
	mu        sync.RWMutex
	customers map[string]*CustomerData
	leads     map[string]*LeadData
	users     map[string]*UserData
}

// CustomerData represents a customer in the store.
type CustomerData struct {
	ID        string
	TenantID  string
	Code      string
	Name      string
	Status    string
	CreatedAt time.Time
}

// LeadData represents a lead in the store.
type LeadData struct {
	ID           string
	TenantID     string
	CompanyName  string
	ContactName  string
	ContactEmail string
	Status       string
	Score        int
	CreatedAt    time.Time
}

// UserData represents a user in the store.
type UserData struct {
	ID        string
	TenantID  string
	Email     string
	FirstName string
	LastName  string
	Status    string
	CreatedAt time.Time
}

// NewTestServer creates a new test server for performance testing.
func NewTestServer() *TestServer {
	store := &InMemoryStore{
		customers: make(map[string]*CustomerData),
		leads:     make(map[string]*LeadData),
		users:     make(map[string]*UserData),
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	// Customer endpoints
	r.Route("/api/v1/customers", func(r chi.Router) {
		r.Get("/", store.listCustomers)
		r.Post("/", store.createCustomer)
		r.Get("/{id}", store.getCustomer)
		r.Put("/{id}", store.updateCustomer)
		r.Delete("/{id}", store.deleteCustomer)
	})

	// Lead endpoints
	r.Route("/api/v1/leads", func(r chi.Router) {
		r.Get("/", store.listLeads)
		r.Post("/", store.createLead)
		r.Get("/{id}", store.getLead)
		r.Put("/{id}", store.updateLead)
		r.Delete("/{id}", store.deleteLead)
	})

	// User endpoints
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Get("/", store.listUsers)
		r.Post("/", store.createUser)
		r.Get("/{id}", store.getUser)
	})

	server := httptest.NewServer(r)

	return &TestServer{
		Server:    server,
		Router:    r,
		DataStore: store,
	}
}

// Close closes the test server.
func (s *TestServer) Close() {
	s.Server.Close()
}

// Seed seeds the data store with test data.
func (s *TestServer) Seed(numCustomers, numLeads, numUsers int) {
	tenantID := uuid.New().String()

	for i := 0; i < numCustomers; i++ {
		id := uuid.New().String()
		s.DataStore.customers[id] = &CustomerData{
			ID:        id,
			TenantID:  tenantID,
			Code:      fmt.Sprintf("CUST-%06d", i),
			Name:      fmt.Sprintf("Customer %d", i),
			Status:    "active",
			CreatedAt: time.Now(),
		}
	}

	for i := 0; i < numLeads; i++ {
		id := uuid.New().String()
		s.DataStore.leads[id] = &LeadData{
			ID:           id,
			TenantID:     tenantID,
			CompanyName:  fmt.Sprintf("Lead Company %d", i),
			ContactName:  fmt.Sprintf("Contact %d", i),
			ContactEmail: fmt.Sprintf("lead%d@example.com", i),
			Status:       "new",
			Score:        50,
			CreatedAt:    time.Now(),
		}
	}

	for i := 0; i < numUsers; i++ {
		id := uuid.New().String()
		s.DataStore.users[id] = &UserData{
			ID:        id,
			TenantID:  tenantID,
			Email:     fmt.Sprintf("user%d@example.com", i),
			FirstName: fmt.Sprintf("User%d", i),
			LastName:  "Test",
			Status:    "active",
			CreatedAt: time.Now(),
		}
	}
}

// Handler implementations
func (s *InMemoryStore) listCustomers(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	customers := make([]*CustomerData, 0, len(s.customers))
	for _, c := range s.customers {
		customers = append(customers, c)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  customers,
		"total": len(customers),
	})
}

func (s *InMemoryStore) createCustomer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code   string `json:"code"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.mu.Lock()
	defer s.mu.Unlock()

	customer := &CustomerData{
		ID:        uuid.New().String(),
		TenantID:  r.Header.Get("X-Tenant-ID"),
		Code:      req.Code,
		Name:      req.Name,
		Status:    req.Status,
		CreatedAt: time.Now(),
	}

	s.customers[customer.ID] = customer
	writeJSON(w, http.StatusCreated, customer)
}

func (s *InMemoryStore) getCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	customer, ok := s.customers[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	writeJSON(w, http.StatusOK, customer)
}

func (s *InMemoryStore) updateCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	s.mu.Lock()
	defer s.mu.Unlock()

	customer, ok := s.customers[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	if name, ok := req["name"].(string); ok {
		customer.Name = name
	}

	writeJSON(w, http.StatusOK, customer)
}

func (s *InMemoryStore) deleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.customers, id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *InMemoryStore) listLeads(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	leads := make([]*LeadData, 0, len(s.leads))
	for _, l := range s.leads {
		leads = append(leads, l)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  leads,
		"total": len(leads),
	})
}

func (s *InMemoryStore) createLead(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CompanyName  string `json:"company_name"`
		ContactName  string `json:"contact_name"`
		ContactEmail string `json:"contact_email"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.mu.Lock()
	defer s.mu.Unlock()

	lead := &LeadData{
		ID:           uuid.New().String(),
		TenantID:     r.Header.Get("X-Tenant-ID"),
		CompanyName:  req.CompanyName,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		Status:       "new",
		Score:        50,
		CreatedAt:    time.Now(),
	}

	s.leads[lead.ID] = lead
	writeJSON(w, http.StatusCreated, lead)
}

func (s *InMemoryStore) getLead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	lead, ok := s.leads[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	writeJSON(w, http.StatusOK, lead)
}

func (s *InMemoryStore) updateLead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	s.mu.Lock()
	defer s.mu.Unlock()

	lead, ok := s.leads[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	if name, ok := req["company_name"].(string); ok {
		lead.CompanyName = name
	}

	writeJSON(w, http.StatusOK, lead)
}

func (s *InMemoryStore) deleteLead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.leads, id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *InMemoryStore) listUsers(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*UserData, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  users,
		"total": len(users),
	})
}

func (s *InMemoryStore) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.mu.Lock()
	defer s.mu.Unlock()

	user := &UserData{
		ID:        uuid.New().String(),
		TenantID:  r.Header.Get("X-Tenant-ID"),
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	s.users[user.ID] = user
	writeJSON(w, http.StatusCreated, user)
}

func (s *InMemoryStore) getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// ============================================================================
// Helper Functions
// ============================================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func calculateAvg(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}
	var total time.Duration
	for _, t := range times {
		total += t
	}
	return total / time.Duration(len(times))
}

func percentile(times []time.Duration, p int) time.Duration {
	if len(times) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(p)/100.0*float64(len(times)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(times) {
		idx = len(times) - 1
	}
	return times[idx]
}

// ============================================================================
// Test Main
// ============================================================================

func TestMain(m *testing.M) {
	if testing.Short() {
		fmt.Println("Skipping performance tests in short mode")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

// ============================================================================
// HTTP Client for Load Testing
// ============================================================================

// LoadTestClient provides HTTP client functionality for load testing.
type LoadTestClient struct {
	client  *http.Client
	baseURL string
}

// NewLoadTestClient creates a new load test client.
func NewLoadTestClient(baseURL string) *LoadTestClient {
	return &LoadTestClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL: baseURL,
	}
}

// DoRequest performs an HTTP request and returns the duration.
func (c *LoadTestClient) DoRequest(method, path string, body interface{}) (time.Duration, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return 0, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Tenant-ID", uuid.New().String())

	start := time.Now()
	resp, err := c.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return duration, err
	}
	defer resp.Body.Close()

	// Drain the body to reuse connections
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return duration, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return duration, nil
}
