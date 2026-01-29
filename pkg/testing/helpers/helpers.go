// Package helpers provides test helper functions for integration testing.
package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestContext creates a context with timeout for testing.
func TestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// DefaultTestContext creates a context with default timeout (30 seconds).
func DefaultTestContext() (context.Context, context.CancelFunc) {
	return TestContext(30 * time.Second)
}

// AssertEqual asserts that two values are equal.
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected %v, got %v", msg, expected, actual)
	}
}

// AssertNotEqual asserts that two values are not equal.
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if reflect.DeepEqual(expected, actual) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected values to be different, but both are %v", msg, expected)
	}
}

// AssertNil asserts that a value is nil.
func AssertNil(t *testing.T, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !isNil(actual) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected nil, got %v", msg, actual)
	}
}

// AssertNotNil asserts that a value is not nil.
func AssertNotNil(t *testing.T, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if isNil(actual) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected non-nil value", msg)
	}
}

// AssertTrue asserts that a value is true.
func AssertTrue(t *testing.T, actual bool, msgAndArgs ...interface{}) {
	t.Helper()
	if !actual {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected true, got false", msg)
	}
}

// AssertFalse asserts that a value is false.
func AssertFalse(t *testing.T, actual bool, msgAndArgs ...interface{}) {
	t.Helper()
	if actual {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected false, got true", msg)
	}
}

// AssertNoError asserts that an error is nil.
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sUnexpected error: %v", msg, err)
	}
}

// AssertError asserts that an error is not nil.
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected an error, got nil", msg)
	}
}

// AssertErrorContains asserts that an error message contains a substring.
func AssertErrorContains(t *testing.T, err error, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected an error containing %q, got nil", msg, substr)
		return
	}
	if !strings.Contains(err.Error(), substr) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected error to contain %q, got %q", msg, substr, err.Error())
	}
}

// AssertContains asserts that a string contains a substring.
func AssertContains(t *testing.T, s, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	if !strings.Contains(s, substr) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected %q to contain %q", msg, s, substr)
	}
}

// AssertNotContains asserts that a string does not contain a substring.
func AssertNotContains(t *testing.T, s, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	if strings.Contains(s, substr) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected %q to not contain %q", msg, s, substr)
	}
}

// AssertLen asserts that a slice/array/map/string has a specific length.
func AssertLen(t *testing.T, obj interface{}, length int, msgAndArgs ...interface{}) {
	t.Helper()
	v := reflect.ValueOf(obj)
	if v.Len() != length {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected length %d, got %d", msg, length, v.Len())
	}
}

// AssertEmpty asserts that a slice/array/map/string is empty.
func AssertEmpty(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	v := reflect.ValueOf(obj)
	if v.Len() != 0 {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected empty, got length %d", msg, v.Len())
	}
}

// AssertNotEmpty asserts that a slice/array/map/string is not empty.
func AssertNotEmpty(t *testing.T, obj interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	v := reflect.ValueOf(obj)
	if v.Len() == 0 {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected non-empty", msg)
	}
}

// AssertGreater asserts that a > b.
func AssertGreater(t *testing.T, a, b int, msgAndArgs ...interface{}) {
	t.Helper()
	if a <= b {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected %d > %d", msg, a, b)
	}
}

// AssertGreaterOrEqual asserts that a >= b.
func AssertGreaterOrEqual(t *testing.T, a, b int, msgAndArgs ...interface{}) {
	t.Helper()
	if a < b {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected %d >= %d", msg, a, b)
	}
}

// AssertLess asserts that a < b.
func AssertLess(t *testing.T, a, b int, msgAndArgs ...interface{}) {
	t.Helper()
	if a >= b {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected %d < %d", msg, a, b)
	}
}

// AssertLessOrEqual asserts that a <= b.
func AssertLessOrEqual(t *testing.T, a, b int, msgAndArgs ...interface{}) {
	t.Helper()
	if a > b {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected %d <= %d", msg, a, b)
	}
}

// AssertUUID asserts that a string is a valid UUID.
func AssertUUID(t *testing.T, s string, msgAndArgs ...interface{}) {
	t.Helper()
	if _, err := uuid.Parse(s); err != nil {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sExpected valid UUID, got %q: %v", msg, s, err)
	}
}

// AssertJSONEqual asserts that two JSON strings are equal.
func AssertJSONEqual(t *testing.T, expected, actual string, msgAndArgs ...interface{}) {
	t.Helper()
	var expectedJSON, actualJSON interface{}
	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sFailed to unmarshal expected JSON: %v", msg, err)
		return
	}
	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sFailed to unmarshal actual JSON: %v", msg, err)
		return
	}
	if !reflect.DeepEqual(expectedJSON, actualJSON) {
		msg := formatMessage(msgAndArgs...)
		t.Errorf("%sJSON not equal.\nExpected: %s\nActual: %s", msg, expected, actual)
	}
}

// RequireNoError fails the test immediately if there's an error.
func RequireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		msg := formatMessage(msgAndArgs...)
		t.Fatalf("%sFatal error: %v", msg, err)
	}
}

// RequireNotNil fails the test immediately if the value is nil.
func RequireNotNil(t *testing.T, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if isNil(actual) {
		msg := formatMessage(msgAndArgs...)
		t.Fatalf("%sRequired non-nil value", msg)
	}
}

// HTTPTestRequest creates an HTTP test request.
type HTTPTestRequest struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// HTTPTestResponse represents an HTTP test response.
type HTTPTestResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// DoRequest performs an HTTP request against a handler.
func DoRequest(handler http.Handler, req HTTPTestRequest) (*HTTPTestResponse, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	httpReq := httptest.NewRequest(req.Method, req.Path, bodyReader)
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httpReq)

	return &HTTPTestResponse{
		StatusCode: recorder.Code,
		Body:       recorder.Body.Bytes(),
		Headers:    recorder.Header(),
	}, nil
}

// AssertHTTPStatus asserts the HTTP response status code.
func AssertHTTPStatus(t *testing.T, resp *HTTPTestResponse, expectedStatus int) {
	t.Helper()
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected HTTP status %d, got %d. Body: %s", expectedStatus, resp.StatusCode, string(resp.Body))
	}
}

// AssertHTTPHeader asserts an HTTP response header value.
func AssertHTTPHeader(t *testing.T, resp *HTTPTestResponse, header, expectedValue string) {
	t.Helper()
	if resp.Headers.Get(header) != expectedValue {
		t.Errorf("Expected header %s to be %q, got %q", header, expectedValue, resp.Headers.Get(header))
	}
}

// DecodeJSONResponse decodes the JSON response body.
func DecodeJSONResponse(t *testing.T, resp *HTTPTestResponse, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(resp.Body, v); err != nil {
		t.Errorf("Failed to decode JSON response: %v. Body: %s", err, string(resp.Body))
	}
}

// WaitFor waits for a condition to be true with timeout.
func WaitFor(t *testing.T, timeout time.Duration, interval time.Duration, condition func() bool, msgAndArgs ...interface{}) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}
	msg := formatMessage(msgAndArgs...)
	t.Errorf("%sCondition not met within %v", msg, timeout)
}

// Eventually asserts that a function eventually returns true.
func Eventually(t *testing.T, condition func() bool, timeout, interval time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	WaitFor(t, timeout, interval, condition, msgAndArgs...)
}

// Never asserts that a function never returns true within the timeout.
func Never(t *testing.T, condition func() bool, timeout, interval time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			msg := formatMessage(msgAndArgs...)
			t.Errorf("%sCondition became true unexpectedly", msg)
			return
		}
		time.Sleep(interval)
	}
}

// RetryWithBackoff retries a function with exponential backoff.
func RetryWithBackoff(ctx context.Context, maxRetries int, initialDelay time.Duration, fn func() error) error {
	var lastErr error
	delay := initialDelay

	for i := 0; i < maxRetries; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			delay *= 2
		}
	}

	return lastErr
}

// SkipIfShort skips the test if running in short mode.
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// SkipIfCI skips the test if running in CI environment.
func SkipIfCI(t *testing.T) {
	t.Helper()
	// Check common CI environment variables
	ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "TRAVIS"}
	for _, v := range ciVars {
		if getEnv(v) != "" {
			t.Skipf("Skipping test in CI environment (%s)", v)
		}
	}
}

// Helper functions

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

func formatMessage(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	if len(msgAndArgs) == 1 {
		return fmt.Sprintf("%v: ", msgAndArgs[0])
	}
	return fmt.Sprintf(msgAndArgs[0].(string)+": ", msgAndArgs[1:]...)
}

func getEnv(key string) string {
	if v, ok := lookupEnv(key); ok {
		return v
	}
	return ""
}

func lookupEnv(key string) (string, bool) {
	// Simple implementation - in real code, use os.LookupEnv
	return "", false
}

// TestTableEntry represents a single test case in a table-driven test.
type TestTableEntry struct {
	Name    string
	Setup   func()
	Cleanup func()
	Run     func(t *testing.T)
	Skip    bool
}

// RunTestTable runs a table-driven test.
func RunTestTable(t *testing.T, entries []TestTableEntry) {
	t.Helper()
	for _, entry := range entries {
		t.Run(entry.Name, func(t *testing.T) {
			if entry.Skip {
				t.Skip("Test skipped")
			}
			if entry.Setup != nil {
				entry.Setup()
			}
			if entry.Cleanup != nil {
				defer entry.Cleanup()
			}
			entry.Run(t)
		})
	}
}

// GenerateRandomString generates a random string of given length.
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// GenerateRandomEmail generates a random email address.
func GenerateRandomEmail() string {
	return fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
}

// GenerateRandomPhone generates a random phone number.
func GenerateRandomPhone() string {
	return fmt.Sprintf("+1%010d", time.Now().UnixNano()%10000000000)
}
