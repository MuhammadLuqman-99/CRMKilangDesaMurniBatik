// Package resilience provides resilience patterns for the CRM application.
package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ============================================================================
// Bulkhead Errors
// ============================================================================

var (
	// ErrBulkheadFull is returned when the bulkhead is full.
	ErrBulkheadFull = errors.New("bulkhead full")

	// ErrBulkheadTimeout is returned when waiting for a slot times out.
	ErrBulkheadTimeout = errors.New("bulkhead timeout")
)

// ============================================================================
// Bulkhead Configuration
// ============================================================================

// BulkheadConfig configures the bulkhead.
type BulkheadConfig struct {
	// Name of the bulkhead for identification.
	Name string

	// MaxConcurrent is the maximum number of concurrent calls.
	MaxConcurrent int

	// MaxWait is the maximum time to wait for a slot.
	MaxWait time.Duration

	// OnFull is called when the bulkhead is full.
	OnFull func(name string)

	// OnAcquire is called when a slot is acquired.
	OnAcquire func(name string)

	// OnRelease is called when a slot is released.
	OnRelease func(name string)
}

// DefaultBulkheadConfig returns default bulkhead configuration.
func DefaultBulkheadConfig(name string) BulkheadConfig {
	return BulkheadConfig{
		Name:          name,
		MaxConcurrent: 10,
		MaxWait:       0, // No waiting by default
	}
}

// ============================================================================
// Bulkhead
// ============================================================================

// Bulkhead limits concurrent access to a resource.
type Bulkhead struct {
	name          string
	maxConcurrent int
	maxWait       time.Duration
	semaphore     chan struct{}
	onFull        func(string)
	onAcquire     func(string)
	onRelease     func(string)
	mu            sync.Mutex
	active        int
	waiting       int
}

// NewBulkhead creates a new bulkhead.
func NewBulkhead(config BulkheadConfig) *Bulkhead {
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 10
	}

	return &Bulkhead{
		name:          config.Name,
		maxConcurrent: config.MaxConcurrent,
		maxWait:       config.MaxWait,
		semaphore:     make(chan struct{}, config.MaxConcurrent),
		onFull:        config.OnFull,
		onAcquire:     config.OnAcquire,
		onRelease:     config.OnRelease,
	}
}

// Execute runs the function with bulkhead protection.
func (b *Bulkhead) Execute(fn func() error) error {
	if err := b.acquire(); err != nil {
		return err
	}
	defer b.release()

	return fn()
}

// ExecuteWithContext runs the function with context and bulkhead protection.
func (b *Bulkhead) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	if err := b.acquireWithContext(ctx); err != nil {
		return err
	}
	defer b.release()

	return fn(ctx)
}

// acquire acquires a slot from the bulkhead.
func (b *Bulkhead) acquire() error {
	if b.maxWait == 0 {
		// Non-blocking acquire
		select {
		case b.semaphore <- struct{}{}:
			b.mu.Lock()
			b.active++
			b.mu.Unlock()

			if b.onAcquire != nil {
				b.onAcquire(b.name)
			}
			return nil
		default:
			if b.onFull != nil {
				b.onFull(b.name)
			}
			return ErrBulkheadFull
		}
	}

	// Blocking acquire with timeout
	b.mu.Lock()
	b.waiting++
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		b.waiting--
		b.mu.Unlock()
	}()

	select {
	case b.semaphore <- struct{}{}:
		b.mu.Lock()
		b.active++
		b.mu.Unlock()

		if b.onAcquire != nil {
			b.onAcquire(b.name)
		}
		return nil
	case <-time.After(b.maxWait):
		if b.onFull != nil {
			b.onFull(b.name)
		}
		return ErrBulkheadTimeout
	}
}

// acquireWithContext acquires a slot with context cancellation.
func (b *Bulkhead) acquireWithContext(ctx context.Context) error {
	if b.maxWait == 0 {
		// Non-blocking acquire
		select {
		case b.semaphore <- struct{}{}:
			b.mu.Lock()
			b.active++
			b.mu.Unlock()

			if b.onAcquire != nil {
				b.onAcquire(b.name)
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			if b.onFull != nil {
				b.onFull(b.name)
			}
			return ErrBulkheadFull
		}
	}

	// Blocking acquire with timeout and context
	b.mu.Lock()
	b.waiting++
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		b.waiting--
		b.mu.Unlock()
	}()

	var timer <-chan time.Time
	if b.maxWait > 0 {
		timer = time.After(b.maxWait)
	}

	select {
	case b.semaphore <- struct{}{}:
		b.mu.Lock()
		b.active++
		b.mu.Unlock()

		if b.onAcquire != nil {
			b.onAcquire(b.name)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer:
		if b.onFull != nil {
			b.onFull(b.name)
		}
		return ErrBulkheadTimeout
	}
}

// release releases a slot back to the bulkhead.
func (b *Bulkhead) release() {
	<-b.semaphore

	b.mu.Lock()
	b.active--
	b.mu.Unlock()

	if b.onRelease != nil {
		b.onRelease(b.name)
	}
}

// ActiveCount returns the number of active calls.
func (b *Bulkhead) ActiveCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active
}

// WaitingCount returns the number of waiting calls.
func (b *Bulkhead) WaitingCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.waiting
}

// AvailableSlots returns the number of available slots.
func (b *Bulkhead) AvailableSlots() int {
	return b.maxConcurrent - b.ActiveCount()
}

// ============================================================================
// Bulkhead Registry
// ============================================================================

// BulkheadRegistry manages multiple bulkheads.
type BulkheadRegistry struct {
	bulkheads map[string]*Bulkhead
	config    BulkheadConfig
	mu        sync.RWMutex
}

// NewBulkheadRegistry creates a new registry.
func NewBulkheadRegistry(defaultConfig BulkheadConfig) *BulkheadRegistry {
	return &BulkheadRegistry{
		bulkheads: make(map[string]*Bulkhead),
		config:    defaultConfig,
	}
}

// Get returns the bulkhead for the given name.
func (r *BulkheadRegistry) Get(name string) *Bulkhead {
	r.mu.RLock()
	bh, ok := r.bulkheads[name]
	r.mu.RUnlock()

	if ok {
		return bh
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double check
	if bh, ok := r.bulkheads[name]; ok {
		return bh
	}

	// Create new bulkhead
	config := r.config
	config.Name = name
	bh = NewBulkhead(config)
	r.bulkheads[name] = bh

	return bh
}

// Remove removes a bulkhead from the registry.
func (r *BulkheadRegistry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.bulkheads, name)
}

// List returns all bulkheads.
func (r *BulkheadRegistry) List() map[string]*Bulkhead {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*Bulkhead, len(r.bulkheads))
	for k, v := range r.bulkheads {
		result[k] = v
	}

	return result
}

// ============================================================================
// Thread Pool Bulkhead
// ============================================================================

// ThreadPoolBulkhead implements bulkhead using a worker pool.
type ThreadPoolBulkhead struct {
	name        string
	maxWorkers  int
	maxQueue    int
	taskQueue   chan func()
	stopCh      chan struct{}
	stoppedCh   chan struct{}
	activeCount int32
	queuedCount int32
	wg          sync.WaitGroup
}

// ThreadPoolConfig configures the thread pool bulkhead.
type ThreadPoolConfig struct {
	Name       string
	MaxWorkers int
	MaxQueue   int
}

// NewThreadPoolBulkhead creates a new thread pool bulkhead.
func NewThreadPoolBulkhead(config ThreadPoolConfig) *ThreadPoolBulkhead {
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 10
	}
	if config.MaxQueue <= 0 {
		config.MaxQueue = 100
	}

	pool := &ThreadPoolBulkhead{
		name:       config.Name,
		maxWorkers: config.MaxWorkers,
		maxQueue:   config.MaxQueue,
		taskQueue:  make(chan func(), config.MaxQueue),
		stopCh:     make(chan struct{}),
		stoppedCh:  make(chan struct{}),
	}

	// Start workers
	for i := 0; i < config.MaxWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker()
	}

	return pool
}

// worker is the goroutine that processes tasks.
func (p *ThreadPoolBulkhead) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.stopCh:
			return
		case task := <-p.taskQueue:
			p.activeCount++
			p.queuedCount--
			task()
			p.activeCount--
		}
	}
}

// Submit submits a task to the pool.
func (p *ThreadPoolBulkhead) Submit(task func()) error {
	select {
	case <-p.stopCh:
		return errors.New("pool stopped")
	default:
	}

	select {
	case p.taskQueue <- task:
		p.queuedCount++
		return nil
	default:
		return ErrBulkheadFull
	}
}

// SubmitWait submits a task and waits for the result.
func (p *ThreadPoolBulkhead) SubmitWait(ctx context.Context, task func() error) error {
	done := make(chan error, 1)

	wrapper := func() {
		done <- task()
	}

	select {
	case <-p.stopCh:
		return errors.New("pool stopped")
	case <-ctx.Done():
		return ctx.Err()
	case p.taskQueue <- wrapper:
		p.queuedCount++
	default:
		return ErrBulkheadFull
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// ActiveCount returns the number of active workers.
func (p *ThreadPoolBulkhead) ActiveCount() int32 {
	return p.activeCount
}

// QueuedCount returns the number of queued tasks.
func (p *ThreadPoolBulkhead) QueuedCount() int32 {
	return p.queuedCount
}

// Stop stops the pool.
func (p *ThreadPoolBulkhead) Stop() {
	close(p.stopCh)
	p.wg.Wait()
	close(p.stoppedCh)
}

// ============================================================================
// Rate Limiter
// ============================================================================

// RateLimiter limits the rate of operations.
type RateLimiter struct {
	name       string
	rate       int           // Operations per second
	burst      int           // Maximum burst size
	tokens     float64       // Current tokens
	lastUpdate time.Time     // Last update time
	mu         sync.Mutex
}

// RateLimiterConfig configures the rate limiter.
type RateLimiterConfig struct {
	Name  string
	Rate  int // Operations per second
	Burst int // Maximum burst size
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.Rate <= 0 {
		config.Rate = 10
	}
	if config.Burst <= 0 {
		config.Burst = config.Rate
	}

	return &RateLimiter{
		name:       config.Name,
		rate:       config.Rate,
		burst:      config.Burst,
		tokens:     float64(config.Burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if an operation is allowed.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens >= 1 {
		r.tokens--
		return true
	}

	return false
}

// Wait waits until an operation is allowed.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		r.refill()

		if r.tokens >= 1 {
			r.tokens--
			r.mu.Unlock()
			return nil
		}

		// Calculate wait time
		waitTime := time.Duration(float64(time.Second) / float64(r.rate))
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}
}

// refill refills tokens based on elapsed time.
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastUpdate)
	r.lastUpdate = now

	// Add tokens based on elapsed time
	tokensToAdd := float64(r.rate) * elapsed.Seconds()
	r.tokens += tokensToAdd

	// Cap at burst size
	if r.tokens > float64(r.burst) {
		r.tokens = float64(r.burst)
	}
}

// Execute executes a function with rate limiting.
func (r *RateLimiter) Execute(ctx context.Context, fn func() error) error {
	if err := r.Wait(ctx); err != nil {
		return err
	}
	return fn()
}

// ============================================================================
// Resilience Decorator
// ============================================================================

// ResilienceDecorator combines multiple resilience patterns.
type ResilienceDecorator struct {
	circuitBreaker *CircuitBreaker
	bulkhead       *Bulkhead
	rateLimiter    *RateLimiter
	retryer        *Retryer
}

// ResilienceConfig configures the resilience decorator.
type ResilienceConfig struct {
	CircuitBreaker *CircuitBreakerConfig
	Bulkhead       *BulkheadConfig
	RateLimiter    *RateLimiterConfig
	Retry          *RetryConfig
}

// NewResilienceDecorator creates a new resilience decorator.
func NewResilienceDecorator(config ResilienceConfig) *ResilienceDecorator {
	d := &ResilienceDecorator{}

	if config.CircuitBreaker != nil {
		d.circuitBreaker = NewCircuitBreaker(*config.CircuitBreaker)
	}

	if config.Bulkhead != nil {
		d.bulkhead = NewBulkhead(*config.Bulkhead)
	}

	if config.RateLimiter != nil {
		d.rateLimiter = NewRateLimiter(*config.RateLimiter)
	}

	if config.Retry != nil {
		d.retryer = NewRetryer(withRetryConfig(*config.Retry))
	}

	return d
}

// Execute executes a function with all configured resilience patterns.
func (d *ResilienceDecorator) Execute(ctx context.Context, fn func(context.Context) error) error {
	var wrappedFn func(context.Context) error = fn

	// Wrap with circuit breaker (innermost)
	if d.circuitBreaker != nil {
		originalFn := wrappedFn
		wrappedFn = func(ctx context.Context) error {
			return d.circuitBreaker.ExecuteWithContext(ctx, originalFn)
		}
	}

	// Wrap with bulkhead
	if d.bulkhead != nil {
		originalFn := wrappedFn
		wrappedFn = func(ctx context.Context) error {
			return d.bulkhead.ExecuteWithContext(ctx, originalFn)
		}
	}

	// Wrap with rate limiter
	if d.rateLimiter != nil {
		originalFn := wrappedFn
		wrappedFn = func(ctx context.Context) error {
			if err := d.rateLimiter.Wait(ctx); err != nil {
				return err
			}
			return originalFn(ctx)
		}
	}

	// Wrap with retry (outermost)
	if d.retryer != nil {
		return d.retryer.Do(ctx, wrappedFn)
	}

	return wrappedFn(ctx)
}
