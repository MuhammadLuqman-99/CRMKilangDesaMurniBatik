// Package resilience provides resilience patterns for the CRM application.
package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Errors
// ============================================================================

var (
	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrCircuitTimeout is returned when a request times out.
	ErrCircuitTimeout = errors.New("circuit breaker timeout")

	// ErrTooManyRequests is returned when too many requests are in flight.
	ErrTooManyRequests = errors.New("too many requests")
)

// ============================================================================
// Circuit Breaker States
// ============================================================================

// State represents the circuit breaker state.
type State int32

const (
	// StateClosed allows requests to pass through.
	StateClosed State = iota

	// StateOpen rejects all requests.
	StateOpen

	// StateHalfOpen allows limited requests to test recovery.
	StateHalfOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ============================================================================
// Circuit Breaker Configuration
// ============================================================================

// CircuitBreakerConfig configures the circuit breaker.
type CircuitBreakerConfig struct {
	// Name of the circuit breaker for identification.
	Name string

	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open.
	MaxRequests uint32

	// Interval is the cyclic period of the closed state for clearing counts.
	Interval time.Duration

	// Timeout is the period of the open state after which the circuit
	// breaker transitions to half-open state.
	Timeout time.Duration

	// ReadyToTrip is called when a request fails in the closed state.
	// If it returns true, the circuit breaker will trip to open.
	ReadyToTrip func(counts Counts) bool

	// OnStateChange is called when the circuit state changes.
	OnStateChange func(name string, from State, to State)

	// IsSuccessful is called to determine if a result is successful.
	// If nil, all nil errors are considered successful.
	IsSuccessful func(err error) bool
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration.
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:        name,
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	}
}

// ============================================================================
// Circuit Breaker Counts
// ============================================================================

// Counts holds the numbers of requests and their successes/failures.
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// ============================================================================
// Circuit Breaker
// ============================================================================

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	onStateChange func(name string, from State, to State)
	isSuccessful  func(err error) bool

	state      State
	generation uint64
	counts     Counts
	expiry     time.Time
	mu         sync.Mutex
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:          config.Name,
		maxRequests:   config.MaxRequests,
		interval:      config.Interval,
		timeout:       config.Timeout,
		readyToTrip:   config.ReadyToTrip,
		onStateChange: config.OnStateChange,
		isSuccessful:  config.IsSuccessful,
		state:         StateClosed,
	}

	if cb.maxRequests == 0 {
		cb.maxRequests = 5
	}
	if cb.interval == 0 {
		cb.interval = 60 * time.Second
	}
	if cb.timeout == 0 {
		cb.timeout = 30 * time.Second
	}
	if cb.readyToTrip == nil {
		cb.readyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		}
	}
	if cb.isSuccessful == nil {
		cb.isSuccessful = func(err error) bool {
			return err == nil
		}
	}

	cb.expiry = time.Now().Add(cb.interval)

	return cb
}

// Execute runs the given function if the circuit breaker allows it.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result := fn()
	cb.afterRequest(generation, cb.isSuccessful(result))

	return result
}

// ExecuteWithContext runs the given function with context.
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	done := make(chan error, 1)

	go func() {
		defer func() {
			if e := recover(); e != nil {
				done <- fmt.Errorf("panic: %v", e)
			}
		}()
		done <- fn(ctx)
	}()

	select {
	case <-ctx.Done():
		cb.afterRequest(generation, false)
		return ctx.Err()
	case result := <-done:
		cb.afterRequest(generation, cb.isSuccessful(result))
		return result
	}
}

// beforeRequest is called before each request.
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	switch state {
	case StateOpen:
		return generation, ErrCircuitOpen
	case StateHalfOpen:
		if cb.counts.Requests >= cb.maxRequests {
			return generation, ErrTooManyRequests
		}
		cb.counts.Requests++
	default: // StateClosed
		cb.counts.Requests++
	}

	return generation, nil
}

// afterRequest is called after each request.
func (cb *CircuitBreaker) afterRequest(generation uint64, success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, currentGeneration := cb.currentState(now)

	if generation != currentGeneration {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// onSuccess handles successful requests.
func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0
	case StateHalfOpen:
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0
		if cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
			cb.setState(StateClosed, now)
		}
	}
}

// onFailure handles failed requests.
func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	switch state {
	case StateClosed:
		cb.counts.TotalFailures++
		cb.counts.ConsecutiveFailures++
		cb.counts.ConsecutiveSuccesses = 0
		if cb.readyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	case StateHalfOpen:
		cb.setState(StateOpen, now)
	}
}

// currentState returns the current state of the circuit breaker.
func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// setState sets the state of the circuit breaker.
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state
	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

// toNewGeneration resets counts and sets new expiry.
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var interval time.Duration
	switch cb.state {
	case StateClosed:
		interval = cb.interval
	case StateOpen:
		interval = cb.timeout
	default: // StateHalfOpen
		interval = 0
	}

	if interval == 0 {
		cb.expiry = time.Time{}
	} else {
		cb.expiry = now.Add(interval)
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Counts returns the current counts.
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return cb.counts
}

// Reset resets the circuit breaker to the closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed, time.Now())
}

// ============================================================================
// Circuit Breaker Registry
// ============================================================================

// CircuitBreakerRegistry manages multiple circuit breakers.
type CircuitBreakerRegistry struct {
	breakers map[string]*CircuitBreaker
	config   CircuitBreakerConfig
	mu       sync.RWMutex
}

// NewCircuitBreakerRegistry creates a new registry.
func NewCircuitBreakerRegistry(defaultConfig CircuitBreakerConfig) *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
		config:   defaultConfig,
	}
}

// Get returns the circuit breaker for the given name.
func (r *CircuitBreakerRegistry) Get(name string) *CircuitBreaker {
	r.mu.RLock()
	cb, ok := r.breakers[name]
	r.mu.RUnlock()

	if ok {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double check
	if cb, ok := r.breakers[name]; ok {
		return cb
	}

	// Create new circuit breaker
	config := r.config
	config.Name = name
	cb = NewCircuitBreaker(config)
	r.breakers[name] = cb

	return cb
}

// Remove removes a circuit breaker from the registry.
func (r *CircuitBreakerRegistry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.breakers, name)
}

// List returns all circuit breakers.
func (r *CircuitBreakerRegistry) List() map[string]*CircuitBreaker {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*CircuitBreaker, len(r.breakers))
	for k, v := range r.breakers {
		result[k] = v
	}

	return result
}

// ResetAll resets all circuit breakers.
func (r *CircuitBreakerRegistry) ResetAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cb := range r.breakers {
		cb.Reset()
	}
}

// ============================================================================
// Circuit Breaker Metrics
// ============================================================================

// CircuitBreakerMetrics provides metrics for circuit breakers.
type CircuitBreakerMetrics struct {
	TotalRequests    int64
	TotalSuccesses   int64
	TotalFailures    int64
	TotalRejected    int64
	StateChanges     int64
	LastStateChange  time.Time
	CurrentState     State
}

// MetricsCollector collects circuit breaker metrics.
type MetricsCollector struct {
	metrics map[string]*CircuitBreakerMetrics
	mu      sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*CircuitBreakerMetrics),
	}
}

// RecordRequest records a request.
func (c *MetricsCollector) RecordRequest(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := c.getOrCreate(name)
	atomic.AddInt64(&m.TotalRequests, 1)
}

// RecordSuccess records a successful request.
func (c *MetricsCollector) RecordSuccess(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := c.getOrCreate(name)
	atomic.AddInt64(&m.TotalSuccesses, 1)
}

// RecordFailure records a failed request.
func (c *MetricsCollector) RecordFailure(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := c.getOrCreate(name)
	atomic.AddInt64(&m.TotalFailures, 1)
}

// RecordRejected records a rejected request.
func (c *MetricsCollector) RecordRejected(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := c.getOrCreate(name)
	atomic.AddInt64(&m.TotalRejected, 1)
}

// RecordStateChange records a state change.
func (c *MetricsCollector) RecordStateChange(name string, state State) {
	c.mu.Lock()
	defer c.mu.Unlock()

	m := c.getOrCreate(name)
	atomic.AddInt64(&m.StateChanges, 1)
	m.LastStateChange = time.Now()
	m.CurrentState = state
}

// GetMetrics returns metrics for a circuit breaker.
func (c *MetricsCollector) GetMetrics(name string) *CircuitBreakerMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.metrics[name]
}

// GetAllMetrics returns all metrics.
func (c *MetricsCollector) GetAllMetrics() map[string]*CircuitBreakerMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*CircuitBreakerMetrics, len(c.metrics))
	for k, v := range c.metrics {
		result[k] = v
	}

	return result
}

// getOrCreate gets or creates metrics for a name.
func (c *MetricsCollector) getOrCreate(name string) *CircuitBreakerMetrics {
	m, ok := c.metrics[name]
	if !ok {
		m = &CircuitBreakerMetrics{}
		c.metrics[name] = m
	}
	return m
}

// ============================================================================
// Service Circuit Breaker
// ============================================================================

// ServiceCircuitBreaker wraps a service call with circuit breaker protection.
type ServiceCircuitBreaker struct {
	registry *CircuitBreakerRegistry
	metrics  *MetricsCollector
}

// NewServiceCircuitBreaker creates a new service circuit breaker.
func NewServiceCircuitBreaker(config CircuitBreakerConfig) *ServiceCircuitBreaker {
	return &ServiceCircuitBreaker{
		registry: NewCircuitBreakerRegistry(config),
		metrics:  NewMetricsCollector(),
	}
}

// Call executes a service call with circuit breaker protection.
func (s *ServiceCircuitBreaker) Call(ctx context.Context, serviceName string, fn func(context.Context) error) error {
	cb := s.registry.Get(serviceName)

	s.metrics.RecordRequest(serviceName)

	err := cb.ExecuteWithContext(ctx, fn)
	if err != nil {
		if errors.Is(err, ErrCircuitOpen) || errors.Is(err, ErrTooManyRequests) {
			s.metrics.RecordRejected(serviceName)
		} else {
			s.metrics.RecordFailure(serviceName)
		}
		return err
	}

	s.metrics.RecordSuccess(serviceName)
	return nil
}

// GetState returns the current state of a service's circuit breaker.
func (s *ServiceCircuitBreaker) GetState(serviceName string) State {
	return s.registry.Get(serviceName).State()
}

// GetMetrics returns metrics for a service.
func (s *ServiceCircuitBreaker) GetMetrics(serviceName string) *CircuitBreakerMetrics {
	return s.metrics.GetMetrics(serviceName)
}

// Reset resets the circuit breaker for a service.
func (s *ServiceCircuitBreaker) Reset(serviceName string) {
	s.registry.Get(serviceName).Reset()
}

// ============================================================================
// Two-Phase Circuit Breaker
// ============================================================================

// TwoPhaseCircuitBreaker implements a two-phase circuit breaker.
// The first phase handles transient failures, the second handles persistent failures.
type TwoPhaseCircuitBreaker struct {
	primary   *CircuitBreaker
	secondary *CircuitBreaker
}

// NewTwoPhaseCircuitBreaker creates a new two-phase circuit breaker.
func NewTwoPhaseCircuitBreaker(name string, primaryConfig, secondaryConfig CircuitBreakerConfig) *TwoPhaseCircuitBreaker {
	primaryConfig.Name = name + "-primary"
	secondaryConfig.Name = name + "-secondary"

	return &TwoPhaseCircuitBreaker{
		primary:   NewCircuitBreaker(primaryConfig),
		secondary: NewCircuitBreaker(secondaryConfig),
	}
}

// Execute runs the function with two-phase circuit breaker protection.
func (cb *TwoPhaseCircuitBreaker) Execute(fn func() error) error {
	// Check secondary (persistent failure protection) first
	secondaryState := cb.secondary.State()
	if secondaryState == StateOpen {
		return ErrCircuitOpen
	}

	// Execute with primary circuit breaker
	err := cb.primary.Execute(fn)

	// If primary trips, increment secondary failure count
	if cb.primary.State() == StateOpen {
		cb.secondary.Execute(func() error {
			return errors.New("primary circuit open")
		})
	}

	return err
}

// State returns the combined state.
func (cb *TwoPhaseCircuitBreaker) State() State {
	if cb.secondary.State() == StateOpen {
		return StateOpen
	}
	return cb.primary.State()
}

// ============================================================================
// Fallback Circuit Breaker
// ============================================================================

// FallbackCircuitBreaker provides fallback functionality.
type FallbackCircuitBreaker struct {
	*CircuitBreaker
	fallback func(error) error
}

// NewFallbackCircuitBreaker creates a circuit breaker with fallback.
func NewFallbackCircuitBreaker(config CircuitBreakerConfig, fallback func(error) error) *FallbackCircuitBreaker {
	return &FallbackCircuitBreaker{
		CircuitBreaker: NewCircuitBreaker(config),
		fallback:       fallback,
	}
}

// ExecuteWithFallback executes with fallback on circuit open.
func (cb *FallbackCircuitBreaker) ExecuteWithFallback(fn func() error) error {
	err := cb.Execute(fn)
	if err != nil && cb.fallback != nil {
		if errors.Is(err, ErrCircuitOpen) || errors.Is(err, ErrTooManyRequests) {
			return cb.fallback(err)
		}
	}
	return err
}

// ============================================================================
// Circuit Breaker Middleware
// ============================================================================

// CircuitBreakerMiddleware provides middleware for HTTP handlers.
type CircuitBreakerMiddleware struct {
	cb      *CircuitBreaker
	handler func() error
}

// NewCircuitBreakerMiddleware creates a new middleware.
func NewCircuitBreakerMiddleware(cb *CircuitBreaker, handler func() error) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		cb:      cb,
		handler: handler,
	}
}

// Execute executes the handler with circuit breaker protection.
func (m *CircuitBreakerMiddleware) Execute() error {
	return m.cb.Execute(m.handler)
}

// ============================================================================
// Circuit Breaker Health Check
// ============================================================================

// CircuitBreakerHealthCheck provides health check functionality.
type CircuitBreakerHealthCheck struct {
	registry *CircuitBreakerRegistry
}

// NewCircuitBreakerHealthCheck creates a new health check.
func NewCircuitBreakerHealthCheck(registry *CircuitBreakerRegistry) *CircuitBreakerHealthCheck {
	return &CircuitBreakerHealthCheck{
		registry: registry,
	}
}

// Check returns the health status of circuit breakers.
func (h *CircuitBreakerHealthCheck) Check() map[string]string {
	result := make(map[string]string)

	for name, cb := range h.registry.List() {
		result[name] = cb.State().String()
	}

	return result
}

// IsHealthy returns true if all circuit breakers are closed.
func (h *CircuitBreakerHealthCheck) IsHealthy() bool {
	for _, cb := range h.registry.List() {
		if cb.State() != StateClosed {
			return false
		}
	}
	return true
}
