// Package resilience provides resilience patterns for the notification service.
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
// Circuit Breaker Errors
// ============================================================================

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// ErrCircuitBreakerTimeout is returned when an operation times out.
var ErrCircuitBreakerTimeout = errors.New("circuit breaker timeout")

// ============================================================================
// Circuit Breaker States
// ============================================================================

// State represents the state of a circuit breaker.
type State int

const (
	// StateClosed means the circuit is closed and requests can pass through.
	StateClosed State = iota
	// StateOpen means the circuit is open and requests are blocked.
	StateOpen
	// StateHalfOpen means the circuit is testing if requests can pass through.
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

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// Name is the name of the circuit breaker.
	Name string

	// MaxRequests is the maximum number of requests allowed to pass
	// when the circuit is half-open. Default: 1
	MaxRequests uint32

	// Interval is the cyclic period of the closed state. Default: 0 (disabled)
	// The circuit breaker clears the internal counts at the end of each interval.
	Interval time.Duration

	// Timeout is the duration of the open state before transitioning to half-open. Default: 60s
	Timeout time.Duration

	// FailureThreshold is the number of failures before the circuit opens. Default: 5
	FailureThreshold uint32

	// SuccessThreshold is the number of successes in half-open state before closing. Default: 1
	SuccessThreshold uint32

	// FailureRatio is the failure ratio threshold (failures/total) to open. Default: 0.5
	// Only used if MinRequests is met.
	FailureRatio float64

	// MinRequests is the minimum number of requests before failure ratio is considered. Default: 10
	MinRequests uint32

	// OnStateChange is called when the circuit breaker state changes.
	OnStateChange func(name string, from State, to State)

	// IsSuccessful determines if an error should be considered a success.
	// By default, any non-nil error is a failure.
	IsSuccessful func(err error) bool

	// ShouldTrip determines if the circuit should trip based on counts.
	// If nil, uses default logic (threshold or ratio).
	ShouldTrip func(counts Counts) bool
}

// DefaultCircuitBreakerConfig returns a default configuration.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:             "default",
		MaxRequests:      1,
		Timeout:          60 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 1,
		FailureRatio:     0.5,
		MinRequests:      10,
	}
}

// ============================================================================
// Counts
// ============================================================================

// Counts holds request counts for the circuit breaker.
type Counts struct {
	Requests             uint32    `json:"requests"`
	TotalSuccesses       uint32    `json:"total_successes"`
	TotalFailures        uint32    `json:"total_failures"`
	ConsecutiveSuccesses uint32    `json:"consecutive_successes"`
	ConsecutiveFailures  uint32    `json:"consecutive_failures"`
	LastSuccess          time.Time `json:"last_success,omitempty"`
	LastFailure          time.Time `json:"last_failure,omitempty"`
}

// onSuccess updates counts for a success.
func (c *Counts) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
	c.LastSuccess = time.Now()
}

// onFailure updates counts for a failure.
func (c *Counts) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
	c.LastFailure = time.Now()
}

// onRequest updates counts for a request.
func (c *Counts) onRequest() {
	c.Requests++
}

// clear resets all counts.
func (c *Counts) clear() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// FailureRatio returns the failure ratio.
func (c *Counts) FailureRatio() float64 {
	if c.Requests == 0 {
		return 0
	}
	return float64(c.TotalFailures) / float64(c.Requests)
}

// ============================================================================
// Circuit Breaker Implementation
// ============================================================================

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	config CircuitBreakerConfig

	mu          sync.RWMutex
	state       State
	counts      Counts
	expiry      time.Time
	lastFailure time.Time
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.MaxRequests == 0 {
		config.MaxRequests = 1
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 1
	}
	if config.IsSuccessful == nil {
		config.IsSuccessful = func(err error) bool { return err == nil }
	}

	cb := &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}

	if config.Interval > 0 {
		go cb.intervalCleaner()
	}

	return cb
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.currentState()
}

// currentState returns the current state (must be called with lock held).
func (cb *CircuitBreaker) currentState() State {
	now := time.Now()

	switch cb.state {
	case StateClosed:
		if cb.config.Interval > 0 && cb.expiry.Before(now) {
			cb.toNewGeneration()
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen)
		}
	}

	return cb.state
}

// Execute executes a function within the circuit breaker.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	return cb.ExecuteContext(context.Background(), func(ctx context.Context) error {
		return fn()
	})
}

// ExecuteContext executes a function within the circuit breaker with context.
func (cb *CircuitBreaker) ExecuteContext(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if we can execute
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case <-ctx.Done():
		cb.afterRequest(ctx.Err())
		return ctx.Err()
	case err := <-done:
		cb.afterRequest(err)
		return err
	}
}

// ExecuteWithFallback executes a function with a fallback if the circuit is open.
func (cb *CircuitBreaker) ExecuteWithFallback(fn func() error, fallback func(error) error) error {
	err := cb.Execute(fn)
	if err != nil && errors.Is(err, ErrCircuitOpen) {
		return fallback(err)
	}
	return err
}

// beforeRequest handles pre-request logic.
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.currentState()

	switch state {
	case StateClosed:
		cb.counts.onRequest()
		return nil
	case StateOpen:
		return ErrCircuitOpen
	case StateHalfOpen:
		if cb.counts.Requests >= cb.config.MaxRequests {
			return ErrCircuitOpen
		}
		cb.counts.onRequest()
		return nil
	default:
		return ErrCircuitOpen
	}
}

// afterRequest handles post-request logic.
func (cb *CircuitBreaker) afterRequest(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.config.IsSuccessful(err) {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

// onSuccess handles successful request.
func (cb *CircuitBreaker) onSuccess() {
	cb.counts.onSuccess()

	switch cb.state {
	case StateClosed:
		// Stay closed
	case StateHalfOpen:
		if cb.counts.ConsecutiveSuccesses >= cb.config.SuccessThreshold {
			cb.setState(StateClosed)
		}
	}
}

// onFailure handles failed request.
func (cb *CircuitBreaker) onFailure() {
	cb.counts.onFailure()
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.shouldTrip() {
			cb.setState(StateOpen)
		}
	case StateHalfOpen:
		cb.setState(StateOpen)
	}
}

// shouldTrip determines if the circuit should trip.
func (cb *CircuitBreaker) shouldTrip() bool {
	if cb.config.ShouldTrip != nil {
		return cb.config.ShouldTrip(cb.counts)
	}

	// Default logic: trip on consecutive failures threshold
	if cb.counts.ConsecutiveFailures >= cb.config.FailureThreshold {
		return true
	}

	// Also trip on failure ratio if min requests met
	if cb.config.MinRequests > 0 && cb.counts.Requests >= cb.config.MinRequests {
		if cb.counts.FailureRatio() >= cb.config.FailureRatio {
			return true
		}
	}

	return false
}

// setState changes the circuit breaker state.
func (cb *CircuitBreaker) setState(state State) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration()

	if cb.config.OnStateChange != nil {
		cb.config.OnStateChange(cb.config.Name, prev, state)
	}
}

// toNewGeneration resets counts for a new generation.
func (cb *CircuitBreaker) toNewGeneration() {
	cb.counts.clear()

	var expiry time.Time
	switch cb.state {
	case StateClosed:
		if cb.config.Interval > 0 {
			expiry = time.Now().Add(cb.config.Interval)
		}
	case StateOpen:
		expiry = time.Now().Add(cb.config.Timeout)
	}
	cb.expiry = expiry
}

// intervalCleaner periodically clears counts in closed state.
func (cb *CircuitBreaker) intervalCleaner() {
	ticker := time.NewTicker(cb.config.Interval)
	defer ticker.Stop()

	for range ticker.C {
		cb.mu.Lock()
		if cb.state == StateClosed && cb.expiry.Before(time.Now()) {
			cb.toNewGeneration()
		}
		cb.mu.Unlock()
	}
}

// Counts returns a copy of the current counts.
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.counts
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateClosed)
}

// ============================================================================
// Circuit Breaker Registry
// ============================================================================

// CircuitBreakerRegistry manages multiple circuit breakers.
type CircuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	config   CircuitBreakerConfig
}

// NewCircuitBreakerRegistry creates a new registry.
func NewCircuitBreakerRegistry(defaultConfig CircuitBreakerConfig) *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
		config:   defaultConfig,
	}
}

// Get returns the circuit breaker for the given name, creating it if needed.
func (r *CircuitBreakerRegistry) Get(name string) *CircuitBreaker {
	r.mu.RLock()
	cb, ok := r.breakers[name]
	r.mu.RUnlock()

	if ok {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, ok := r.breakers[name]; ok {
		return cb
	}

	config := r.config
	config.Name = name
	cb = NewCircuitBreaker(config)
	r.breakers[name] = cb

	return cb
}

// GetOrCreate gets or creates a circuit breaker with custom config.
func (r *CircuitBreakerRegistry) GetOrCreate(name string, config CircuitBreakerConfig) *CircuitBreaker {
	r.mu.RLock()
	cb, ok := r.breakers[name]
	r.mu.RUnlock()

	if ok {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if cb, ok := r.breakers[name]; ok {
		return cb
	}

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

// List returns all circuit breaker names.
func (r *CircuitBreakerRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.breakers))
	for name := range r.breakers {
		names = append(names, name)
	}
	return names
}

// States returns the states of all circuit breakers.
func (r *CircuitBreakerRegistry) States() map[string]State {
	r.mu.RLock()
	defer r.mu.RUnlock()

	states := make(map[string]State, len(r.breakers))
	for name, cb := range r.breakers {
		states[name] = cb.State()
	}
	return states
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
// Two-Level Circuit Breaker
// ============================================================================

// TwoLevelCircuitBreaker provides per-operation and per-service circuit breakers.
type TwoLevelCircuitBreaker struct {
	service    *CircuitBreaker
	operations *CircuitBreakerRegistry
}

// NewTwoLevelCircuitBreaker creates a new two-level circuit breaker.
func NewTwoLevelCircuitBreaker(serviceConfig, operationConfig CircuitBreakerConfig) *TwoLevelCircuitBreaker {
	return &TwoLevelCircuitBreaker{
		service:    NewCircuitBreaker(serviceConfig),
		operations: NewCircuitBreakerRegistry(operationConfig),
	}
}

// Execute executes a function with both levels of protection.
func (cb *TwoLevelCircuitBreaker) Execute(operation string, fn func() error) error {
	// Check service-level breaker first
	if cb.service.State() == StateOpen {
		return ErrCircuitOpen
	}

	// Get operation-level breaker
	opBreaker := cb.operations.Get(operation)

	// Execute through operation breaker
	err := opBreaker.Execute(fn)

	// Update service-level breaker
	if err != nil {
		cb.service.afterRequest(err)
	} else {
		cb.service.afterRequest(nil)
	}

	return err
}

// ServiceState returns the service-level circuit state.
func (cb *TwoLevelCircuitBreaker) ServiceState() State {
	return cb.service.State()
}

// OperationState returns the operation-level circuit state.
func (cb *TwoLevelCircuitBreaker) OperationState(operation string) State {
	return cb.operations.Get(operation).State()
}

// ============================================================================
// Metrics
// ============================================================================

// CircuitBreakerMetrics holds metrics for a circuit breaker.
type CircuitBreakerMetrics struct {
	Name                string        `json:"name"`
	State               string        `json:"state"`
	TotalRequests       uint64        `json:"total_requests"`
	TotalSuccesses      uint64        `json:"total_successes"`
	TotalFailures       uint64        `json:"total_failures"`
	ConsecutiveFailures uint32        `json:"consecutive_failures"`
	LastFailure         *time.Time    `json:"last_failure,omitempty"`
	OpenDuration        time.Duration `json:"open_duration"`
}

// CircuitBreakerWithMetrics wraps a circuit breaker with metrics.
type CircuitBreakerWithMetrics struct {
	*CircuitBreaker
	totalRequests  uint64
	totalSuccesses uint64
	totalFailures  uint64
	openTime       time.Time
	totalOpenTime  time.Duration
	mu             sync.Mutex
}

// NewCircuitBreakerWithMetrics creates a circuit breaker with metrics.
func NewCircuitBreakerWithMetrics(config CircuitBreakerConfig) *CircuitBreakerWithMetrics {
	// Wrap the state change callback to track metrics
	originalOnChange := config.OnStateChange
	cb := &CircuitBreakerWithMetrics{}

	config.OnStateChange = func(name string, from, to State) {
		cb.mu.Lock()
		if to == StateOpen {
			cb.openTime = time.Now()
		} else if from == StateOpen {
			cb.totalOpenTime += time.Since(cb.openTime)
		}
		cb.mu.Unlock()

		if originalOnChange != nil {
			originalOnChange(name, from, to)
		}
	}

	cb.CircuitBreaker = NewCircuitBreaker(config)
	return cb
}

// Execute executes a function and updates metrics.
func (cb *CircuitBreakerWithMetrics) Execute(fn func() error) error {
	atomic.AddUint64(&cb.totalRequests, 1)
	err := cb.CircuitBreaker.Execute(fn)
	if err == nil {
		atomic.AddUint64(&cb.totalSuccesses, 1)
	} else if !errors.Is(err, ErrCircuitOpen) {
		atomic.AddUint64(&cb.totalFailures, 1)
	}
	return err
}

// Metrics returns the current metrics.
func (cb *CircuitBreakerWithMetrics) Metrics() CircuitBreakerMetrics {
	cb.mu.Lock()
	openDuration := cb.totalOpenTime
	if cb.State() == StateOpen {
		openDuration += time.Since(cb.openTime)
	}
	cb.mu.Unlock()

	counts := cb.Counts()
	var lastFailure *time.Time
	if !counts.LastFailure.IsZero() {
		lastFailure = &counts.LastFailure
	}

	return CircuitBreakerMetrics{
		Name:                cb.config.Name,
		State:               cb.State().String(),
		TotalRequests:       atomic.LoadUint64(&cb.totalRequests),
		TotalSuccesses:      atomic.LoadUint64(&cb.totalSuccesses),
		TotalFailures:       atomic.LoadUint64(&cb.totalFailures),
		ConsecutiveFailures: counts.ConsecutiveFailures,
		LastFailure:         lastFailure,
		OpenDuration:        openDuration,
	}
}

// ============================================================================
// Provider Circuit Breaker
// ============================================================================

// ProviderCircuitBreaker is a circuit breaker for notification providers.
type ProviderCircuitBreaker struct {
	breaker *CircuitBreakerWithMetrics
	name    string
}

// NewProviderCircuitBreaker creates a circuit breaker for a provider.
func NewProviderCircuitBreaker(providerName string) *ProviderCircuitBreaker {
	config := CircuitBreakerConfig{
		Name:             fmt.Sprintf("provider:%s", providerName),
		MaxRequests:      3,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		FailureRatio:     0.5,
		MinRequests:      10,
		Interval:         10 * time.Second,
		OnStateChange: func(name string, from, to State) {
			// Log state changes in production
		},
	}

	return &ProviderCircuitBreaker{
		breaker: NewCircuitBreakerWithMetrics(config),
		name:    providerName,
	}
}

// Execute executes a function within the provider circuit breaker.
func (cb *ProviderCircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	return cb.breaker.ExecuteContext(ctx, fn)
}

// State returns the current state.
func (cb *ProviderCircuitBreaker) State() State {
	return cb.breaker.State()
}

// IsAvailable returns true if the circuit is not open.
func (cb *ProviderCircuitBreaker) IsAvailable() bool {
	return cb.breaker.State() != StateOpen
}

// Metrics returns the circuit breaker metrics.
func (cb *ProviderCircuitBreaker) Metrics() CircuitBreakerMetrics {
	return cb.breaker.Metrics()
}

// Reset resets the circuit breaker.
func (cb *ProviderCircuitBreaker) Reset() {
	cb.breaker.Reset()
}
