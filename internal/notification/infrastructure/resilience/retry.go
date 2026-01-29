// Package resilience provides resilience patterns for the notification service.
package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
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

// WithMaxAttempts sets the maximum number of retry attempts.
func WithMaxAttempts(attempts int) RetryConfigOption {
	return func(c *RetryConfig) {
		c.MaxAttempts = attempts
	}
}

// WithInitialDelay sets the initial delay between retries.
func WithInitialDelay(delay time.Duration) RetryConfigOption {
	return func(c *RetryConfig) {
		c.InitialDelay = delay
	}
}

// WithMaxDelay sets the maximum delay between retries.
func WithMaxDelay(delay time.Duration) RetryConfigOption {
	return func(c *RetryConfig) {
		c.MaxDelay = delay
	}
}

// WithMultiplier sets the multiplier for exponential backoff.
func WithMultiplier(multiplier float64) RetryConfigOption {
	return func(c *RetryConfig) {
		c.Multiplier = multiplier
	}
}

// WithJitter sets the jitter factor for randomization.
func WithJitter(jitter float64) RetryConfigOption {
	return func(c *RetryConfig) {
		c.Jitter = jitter
	}
}

// WithRetryOn sets specific errors to retry on.
func WithRetryOn(errors ...error) RetryConfigOption {
	return func(c *RetryConfig) {
		c.RetryOn = errors
	}
}

// WithDoNotRetryOn sets errors that should not be retried.
func WithDoNotRetryOn(errors ...error) RetryConfigOption {
	return func(c *RetryConfig) {
		c.DoNotRetryOn = errors
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

// NonRetryableError is an error that should not be retried.
type NonRetryableError struct {
	Err error
}

// Error returns the error message.
func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("non-retryable: %v", e.Err)
}

// Unwrap returns the underlying error.
func (e *NonRetryableError) Unwrap() error {
	return e.Err
}

// MarkNonRetryable marks an error as non-retryable.
func MarkNonRetryable(err error) error {
	return &NonRetryableError{Err: err}
}

// IsNonRetryable checks if an error is non-retryable.
func IsNonRetryable(err error) bool {
	var nonRetryable *NonRetryableError
	return errors.As(err, &nonRetryable)
}

// ============================================================================
// Retryer
// ============================================================================

// Retryer provides retry functionality with exponential backoff.
type Retryer struct {
	config RetryConfig
	rand   *rand.Rand
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
func (r *Retryer) DoWithResult[T any](ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T
	var allErrors []error
	var lastErr error

	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		res, err := fn(ctx)
		if err == nil {
			return res, nil
		}

		lastErr = err
		allErrors = append(allErrors, err)

		if !r.shouldRetry(err) {
			return result, &RetryError{
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
			return result, ctx.Err()
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

	// Check for non-retryable marker
	if IsNonRetryable(err) {
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

// calculateDelay calculates the delay for the given attempt with exponential backoff and jitter.
func (r *Retryer) calculateDelay(attempt int) time.Duration {
	// Calculate base delay with exponential backoff
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.Multiplier, float64(attempt))

	// Apply jitter
	if r.config.Jitter > 0 {
		jitterRange := delay * r.config.Jitter
		jitter := (r.rand.Float64() * 2 * jitterRange) - jitterRange
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
// Retry Strategies
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
// This is recommended by AWS for better distribution of retries.
type DecorrelatedJitterBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	rand      *rand.Rand
	lastDelay time.Duration
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
	b.lastDelay = b.BaseDelay
}

// ============================================================================
// Retry Helpers
// ============================================================================

// Retry executes a function with default retry configuration.
func Retry(ctx context.Context, fn func(ctx context.Context) error) error {
	return NewRetryer().Do(ctx, fn)
}

// RetryWithOptions executes a function with custom retry options.
func RetryWithOptions(ctx context.Context, fn func(ctx context.Context) error, opts ...RetryConfigOption) error {
	return NewRetryer(opts...).Do(ctx, fn)
}

// RetryN executes a function with a specific number of attempts.
func RetryN(ctx context.Context, attempts int, fn func(ctx context.Context) error) error {
	return NewRetryer(WithMaxAttempts(attempts)).Do(ctx, fn)
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

		if IsNonRetryable(err) {
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

// ============================================================================
// Notification-Specific Retry
// ============================================================================

// NotificationRetryConfig provides notification-specific retry configuration.
type NotificationRetryConfig struct {
	RetryConfig

	// RetryableStatusCodes are HTTP status codes that should be retried
	RetryableStatusCodes []int

	// RetryableProviderCodes are provider-specific error codes that should be retried
	RetryableProviderCodes []string
}

// DefaultNotificationRetryConfig returns default notification retry configuration.
func DefaultNotificationRetryConfig() NotificationRetryConfig {
	return NotificationRetryConfig{
		RetryConfig: RetryConfig{
			MaxAttempts:  3,
			InitialDelay: 5 * time.Second,
			MaxDelay:     5 * time.Minute,
			Multiplier:   2.0,
			Jitter:       0.3,
		},
		RetryableStatusCodes: []int{
			408, // Request Timeout
			429, // Too Many Requests
			500, // Internal Server Error
			502, // Bad Gateway
			503, // Service Unavailable
			504, // Gateway Timeout
		},
	}
}

// NotificationRetryer provides notification-specific retry functionality.
type NotificationRetryer struct {
	*Retryer
	config NotificationRetryConfig
}

// NewNotificationRetryer creates a new notification retryer.
func NewNotificationRetryer(config NotificationRetryConfig) *NotificationRetryer {
	return &NotificationRetryer{
		Retryer: &Retryer{
			config: config.RetryConfig,
			rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		},
		config: config,
	}
}

// ShouldRetryStatusCode checks if a status code should be retried.
func (r *NotificationRetryer) ShouldRetryStatusCode(statusCode int) bool {
	for _, code := range r.config.RetryableStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// ShouldRetryProviderCode checks if a provider code should be retried.
func (r *NotificationRetryer) ShouldRetryProviderCode(code string) bool {
	for _, c := range r.config.RetryableProviderCodes {
		if c == code {
			return true
		}
	}
	return false
}

// CalculateNextRetryTime calculates the next retry time based on attempt count.
func (r *NotificationRetryer) CalculateNextRetryTime(attemptCount int) time.Time {
	delay := r.calculateDelay(attemptCount)
	return time.Now().Add(delay)
}
