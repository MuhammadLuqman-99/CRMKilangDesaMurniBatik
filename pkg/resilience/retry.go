// Package resilience provides resilience patterns for the CRM application.
package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// ============================================================================
// Retry Configuration
// ============================================================================

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	Multiplier      float64       // Multiplier for exponential backoff
	Jitter          float64       // Jitter factor (0-1) for randomization
	RetryOn         []error       // Specific errors to retry on (empty = retry all)
	DoNotRetryOn    []error       // Errors that should not be retried
}

// DefaultRetryConfig returns a default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.2,
	}
}

// RetryConfigOption is a function that modifies RetryConfig.
type RetryConfigOption func(*RetryConfig)

// WithRetryMaxAttempts sets the maximum number of retry attempts.
func WithRetryMaxAttempts(attempts int) RetryConfigOption {
	return func(c *RetryConfig) {
		c.MaxAttempts = attempts
	}
}

// WithRetryInitialDelay sets the initial delay between retries.
func WithRetryInitialDelay(delay time.Duration) RetryConfigOption {
	return func(c *RetryConfig) {
		c.InitialDelay = delay
	}
}

// WithRetryMaxDelay sets the maximum delay between retries.
func WithRetryMaxDelay(delay time.Duration) RetryConfigOption {
	return func(c *RetryConfig) {
		c.MaxDelay = delay
	}
}

// WithRetryMultiplier sets the multiplier for exponential backoff.
func WithRetryMultiplier(multiplier float64) RetryConfigOption {
	return func(c *RetryConfig) {
		c.Multiplier = multiplier
	}
}

// WithRetryJitter sets the jitter factor for randomization.
func WithRetryJitter(jitter float64) RetryConfigOption {
	return func(c *RetryConfig) {
		c.Jitter = jitter
	}
}

// WithRetryOnErrors sets specific errors to retry on.
func WithRetryOnErrors(errs ...error) RetryConfigOption {
	return func(c *RetryConfig) {
		c.RetryOn = errs
	}
}

// WithDoNotRetryOnErrors sets errors that should not be retried.
func WithDoNotRetryOnErrors(errs ...error) RetryConfigOption {
	return func(c *RetryConfig) {
		c.DoNotRetryOn = errs
	}
}

// ============================================================================
// Retry Errors
// ============================================================================

// RetryError represents an error that occurred during retry.
type RetryError struct {
	Attempts int
	LastErr  error
	Errors   []error
}

// Error returns the error message.
func (e *RetryError) Error() string {
	return fmt.Sprintf("failed after %d attempts: %v", e.Attempts, e.LastErr)
}

// Unwrap returns the last error.
func (e *RetryError) Unwrap() error {
	return e.LastErr
}

// PermanentError is an error that should not be retried.
type PermanentError struct {
	Err error
}

// Error returns the error message.
func (e *PermanentError) Error() string {
	return fmt.Sprintf("permanent error: %v", e.Err)
}

// Unwrap returns the underlying error.
func (e *PermanentError) Unwrap() error {
	return e.Err
}

// MarkPermanent marks an error as permanent (non-retryable).
func MarkPermanent(err error) error {
	return &PermanentError{Err: err}
}

// IsPermanent checks if an error is permanent.
func IsPermanent(err error) bool {
	var permanent *PermanentError
	return errors.As(err, &permanent)
}

// ============================================================================
// Retryer
// ============================================================================

// Retryer provides retry functionality with exponential backoff.
type Retryer struct {
	config RetryConfig
	rand   *rand.Rand
	mu     sync.Mutex
}

// NewRetryer creates a new Retryer with the given options.
func NewRetryer(opts ...RetryConfigOption) *Retryer {
	config := DefaultRetryConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return &Retryer{
		config: config,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Do executes a function with retry logic.
func (r *Retryer) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	var allErrors []error
	var lastErr error

	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err
		allErrors = append(allErrors, err)

		// Check if error should not be retried
		if !r.shouldRetry(err) {
			return &RetryError{
				Attempts: attempt + 1,
				LastErr:  err,
				Errors:   allErrors,
			}
		}

		// Don't wait after the last attempt
		if attempt == r.config.MaxAttempts-1 {
			break
		}

		// Calculate delay with exponential backoff
		delay := r.calculateDelay(attempt)

		// Wait for delay or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return &RetryError{
		Attempts: r.config.MaxAttempts,
		LastErr:  lastErr,
		Errors:   allErrors,
	}
}

// DoWithResult executes a function that returns a value with retry logic.
func (r *Retryer) DoWithResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var allErrors []error
	var lastErr error
	var result interface{}

	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		res, err := fn(ctx)
		if err == nil {
			return res, nil
		}

		lastErr = err
		allErrors = append(allErrors, err)

		if !r.shouldRetry(err) {
			return nil, &RetryError{
				Attempts: attempt + 1,
				LastErr:  err,
				Errors:   allErrors,
			}
		}

		if attempt == r.config.MaxAttempts-1 {
			break
		}

		delay := r.calculateDelay(attempt)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return result, &RetryError{
		Attempts: r.config.MaxAttempts,
		LastErr:  lastErr,
		Errors:   allErrors,
	}
}

// shouldRetry determines if an error should be retried.
func (r *Retryer) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Check for permanent error marker
	if IsPermanent(err) {
		return false
	}

	// Check do-not-retry list
	for _, noRetry := range r.config.DoNotRetryOn {
		if errors.Is(err, noRetry) {
			return false
		}
	}

	// If specific retry errors are configured, only retry on those
	if len(r.config.RetryOn) > 0 {
		for _, retryErr := range r.config.RetryOn {
			if errors.Is(err, retryErr) {
				return true
			}
		}
		return false
	}

	// Default: retry on all errors
	return true
}

// calculateDelay calculates the delay for the given attempt.
func (r *Retryer) calculateDelay(attempt int) time.Duration {
	// Calculate base delay with exponential backoff
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.Multiplier, float64(attempt))

	// Apply jitter
	if r.config.Jitter > 0 {
		r.mu.Lock()
		jitterRange := delay * r.config.Jitter
		jitter := (r.rand.Float64() * 2 * jitterRange) - jitterRange
		r.mu.Unlock()
		delay += jitter
	}

	// Cap at max delay
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	// Ensure minimum delay
	if delay < 0 {
		delay = float64(r.config.InitialDelay)
	}

	return time.Duration(delay)
}

// ============================================================================
// Retry Helpers
// ============================================================================

// Retry executes a function with default retry configuration.
func Retry(ctx context.Context, fn func(ctx context.Context) error) error {
	return NewRetryer().Do(ctx, fn)
}

// RetryWithConfig executes a function with custom retry configuration.
func RetryWithConfig(ctx context.Context, config RetryConfig, fn func(ctx context.Context) error) error {
	retryer := &Retryer{
		config: config,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return retryer.Do(ctx, fn)
}

// RetryN executes a function with a specific number of attempts.
func RetryN(ctx context.Context, attempts int, fn func(ctx context.Context) error) error {
	return NewRetryer(WithRetryMaxAttempts(attempts)).Do(ctx, fn)
}

// ============================================================================
// Retry with Circuit Breaker
// ============================================================================

// RetryWithCircuitBreaker combines retry with circuit breaker.
type RetryWithCircuitBreaker struct {
	retryer        *Retryer
	circuitBreaker *CircuitBreaker
}

// NewRetryWithCircuitBreaker creates a new retry with circuit breaker.
func NewRetryWithCircuitBreaker(retryConfig RetryConfig, cbConfig CircuitBreakerConfig) *RetryWithCircuitBreaker {
	return &RetryWithCircuitBreaker{
		retryer:        NewRetryer(withRetryConfig(retryConfig)),
		circuitBreaker: NewCircuitBreaker(cbConfig),
	}
}

// withRetryConfig converts RetryConfig to options.
func withRetryConfig(config RetryConfig) RetryConfigOption {
	return func(c *RetryConfig) {
		c.MaxAttempts = config.MaxAttempts
		c.InitialDelay = config.InitialDelay
		c.MaxDelay = config.MaxDelay
		c.Multiplier = config.Multiplier
		c.Jitter = config.Jitter
		c.RetryOn = config.RetryOn
		c.DoNotRetryOn = config.DoNotRetryOn
	}
}

// Execute executes a function with retry and circuit breaker.
func (r *RetryWithCircuitBreaker) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.retryer.Do(ctx, func(ctx context.Context) error {
		return r.circuitBreaker.ExecuteWithContext(ctx, fn)
	})
}

// ============================================================================
// Backoff Strategies
// ============================================================================

// BackoffStrategy defines a strategy for calculating retry delays.
type BackoffStrategy interface {
	// NextDelay returns the delay for the given attempt.
	NextDelay(attempt int) time.Duration
	// Reset resets the strategy state.
	Reset()
}

// ExponentialBackoff implements exponential backoff strategy.
type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// NewExponentialBackoff creates a new exponential backoff strategy.
func NewExponentialBackoff(initialDelay, maxDelay time.Duration, multiplier float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay: initialDelay,
		MaxDelay:     maxDelay,
		Multiplier:   multiplier,
	}
}

// NextDelay returns the delay for the given attempt.
func (b *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	delay := float64(b.InitialDelay) * math.Pow(b.Multiplier, float64(attempt))
	if delay > float64(b.MaxDelay) {
		delay = float64(b.MaxDelay)
	}
	return time.Duration(delay)
}

// Reset resets the strategy state.
func (b *ExponentialBackoff) Reset() {}

// LinearBackoff implements linear backoff strategy.
type LinearBackoff struct {
	InitialDelay time.Duration
	Increment    time.Duration
	MaxDelay     time.Duration
}

// NewLinearBackoff creates a new linear backoff strategy.
func NewLinearBackoff(initialDelay, increment, maxDelay time.Duration) *LinearBackoff {
	return &LinearBackoff{
		InitialDelay: initialDelay,
		Increment:    increment,
		MaxDelay:     maxDelay,
	}
}

// NextDelay returns the delay for the given attempt.
func (b *LinearBackoff) NextDelay(attempt int) time.Duration {
	delay := b.InitialDelay + time.Duration(attempt)*b.Increment
	if delay > b.MaxDelay {
		delay = b.MaxDelay
	}
	return delay
}

// Reset resets the strategy state.
func (b *LinearBackoff) Reset() {}

// ConstantBackoff implements constant delay strategy.
type ConstantBackoff struct {
	Delay time.Duration
}

// NewConstantBackoff creates a new constant backoff strategy.
func NewConstantBackoff(delay time.Duration) *ConstantBackoff {
	return &ConstantBackoff{Delay: delay}
}

// NextDelay returns the constant delay.
func (b *ConstantBackoff) NextDelay(attempt int) time.Duration {
	return b.Delay
}

// Reset resets the strategy state.
func (b *ConstantBackoff) Reset() {}

// DecorrelatedJitterBackoff implements decorrelated jitter strategy.
type DecorrelatedJitterBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	rand      *rand.Rand
	lastDelay time.Duration
	mu        sync.Mutex
}

// NewDecorrelatedJitterBackoff creates a new decorrelated jitter backoff.
func NewDecorrelatedJitterBackoff(baseDelay, maxDelay time.Duration) *DecorrelatedJitterBackoff {
	return &DecorrelatedJitterBackoff{
		BaseDelay: baseDelay,
		MaxDelay:  maxDelay,
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		lastDelay: baseDelay,
	}
}

// NextDelay returns the next delay with decorrelated jitter.
func (b *DecorrelatedJitterBackoff) NextDelay(attempt int) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	if attempt == 0 {
		b.lastDelay = b.BaseDelay
		return b.BaseDelay
	}

	// delay = min(cap, random_between(base, delay * 3))
	min := float64(b.BaseDelay)
	max := float64(b.lastDelay) * 3
	delay := min + b.rand.Float64()*(max-min)

	if delay > float64(b.MaxDelay) {
		delay = float64(b.MaxDelay)
	}

	b.lastDelay = time.Duration(delay)
	return b.lastDelay
}

// Reset resets the strategy state.
func (b *DecorrelatedJitterBackoff) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lastDelay = b.BaseDelay
}

// RetryWithBackoff executes a function with a custom backoff strategy.
func RetryWithBackoff(ctx context.Context, strategy BackoffStrategy, maxAttempts int, fn func(ctx context.Context) error) error {
	var allErrors []error
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn(ctx)
		if err == nil {
			strategy.Reset()
			return nil
		}

		lastErr = err
		allErrors = append(allErrors, err)

		if IsPermanent(err) {
			return &RetryError{
				Attempts: attempt + 1,
				LastErr:  err,
				Errors:   allErrors,
			}
		}

		if attempt == maxAttempts-1 {
			break
		}

		delay := strategy.NextDelay(attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return &RetryError{
		Attempts: maxAttempts,
		LastErr:  lastErr,
		Errors:   allErrors,
	}
}
