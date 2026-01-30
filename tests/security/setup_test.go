// Package security contains security tests for the CRM system.
package security

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// Security Test Configuration
// ============================================================================

// SecurityTestConfig holds security test configuration.
type SecurityTestConfig struct {
	// Rate limiting
	MaxRequestsPerMinute int
	MaxLoginAttempts     int
	LockoutDuration      time.Duration

	// Token settings
	TokenExpiration      time.Duration
	RefreshExpiration    time.Duration

	// Password policy
	MinPasswordLength    int
	RequireUppercase     bool
	RequireLowercase     bool
	RequireNumbers       bool
	RequireSpecialChars  bool
}

// DefaultSecurityConfig returns default security test configuration.
func DefaultSecurityConfig() *SecurityTestConfig {
	return &SecurityTestConfig{
		MaxRequestsPerMinute: 100,
		MaxLoginAttempts:     5,
		LockoutDuration:      15 * time.Minute,
		TokenExpiration:      1 * time.Hour,
		RefreshExpiration:    7 * 24 * time.Hour,
		MinPasswordLength:    8,
		RequireUppercase:     true,
		RequireLowercase:     true,
		RequireNumbers:       true,
		RequireSpecialChars:  true,
	}
}

// ============================================================================
// Security Test Results
// ============================================================================

// VulnerabilityReport contains security test results.
type VulnerabilityReport struct {
	TestName        string
	Category        string
	Severity        string // Critical, High, Medium, Low, Info
	Passed          bool
	Description     string
	Recommendation  string
	Evidence        string
}

// SecurityReport contains overall security test report.
type SecurityReport struct {
	mu              sync.Mutex
	Vulnerabilities []VulnerabilityReport
	TotalTests      int
	PassedTests     int
	FailedTests     int
	CriticalCount   int
	HighCount       int
	MediumCount     int
	LowCount        int
}

// NewSecurityReport creates a new security report.
func NewSecurityReport() *SecurityReport {
	return &SecurityReport{
		Vulnerabilities: make([]VulnerabilityReport, 0),
	}
}

// AddResult adds a vulnerability result to the report.
func (r *SecurityReport) AddResult(result VulnerabilityReport) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Vulnerabilities = append(r.Vulnerabilities, result)
	r.TotalTests++

	if result.Passed {
		r.PassedTests++
	} else {
		r.FailedTests++
		switch result.Severity {
		case "Critical":
			r.CriticalCount++
		case "High":
			r.HighCount++
		case "Medium":
			r.MediumCount++
		case "Low":
			r.LowCount++
		}
	}
}

// Print prints the security report.
func (r *SecurityReport) Print() {
	fmt.Println("================================================================================")
	fmt.Println("                        SECURITY TEST REPORT")
	fmt.Println("================================================================================")
	fmt.Println()

	fmt.Printf("Total Tests:    %d\n", r.TotalTests)
	fmt.Printf("Passed:         %d\n", r.PassedTests)
	fmt.Printf("Failed:         %d\n", r.FailedTests)
	fmt.Println()

	if r.FailedTests > 0 {
		fmt.Println("Vulnerability Summary:")
		fmt.Printf("  Critical:     %d\n", r.CriticalCount)
		fmt.Printf("  High:         %d\n", r.HighCount)
		fmt.Printf("  Medium:       %d\n", r.MediumCount)
		fmt.Printf("  Low:          %d\n", r.LowCount)
		fmt.Println()

		fmt.Println("Failed Tests:")
		fmt.Println(strings.Repeat("-", 80))

		for _, v := range r.Vulnerabilities {
			if !v.Passed {
				fmt.Printf("\n[%s] %s - %s\n", v.Severity, v.Category, v.TestName)
				fmt.Printf("  Description:    %s\n", v.Description)
				fmt.Printf("  Recommendation: %s\n", v.Recommendation)
				if v.Evidence != "" {
					fmt.Printf("  Evidence:       %s\n", v.Evidence)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("================================================================================")
	if r.CriticalCount > 0 || r.HighCount > 0 {
		fmt.Println("  RESULT: FAILED - Critical/High vulnerabilities found")
	} else if r.MediumCount > 0 {
		fmt.Println("  RESULT: WARNING - Medium vulnerabilities found")
	} else if r.FailedTests > 0 {
		fmt.Println("  RESULT: PASSED WITH WARNINGS - Low vulnerabilities found")
	} else {
		fmt.Println("  RESULT: PASSED - No vulnerabilities found")
	}
	fmt.Println("================================================================================")
}

// ============================================================================
// Security Test Server
// ============================================================================

// SecurityTestServer provides a test server with security features.
type SecurityTestServer struct {
	Server          *httptest.Server
	Router          *chi.Mux
	DataStore       *SecureDataStore
	Config          *SecurityTestConfig
	RateLimiter     *RateLimiter
	LoginAttempts   map[string]int
	LockedAccounts  map[string]time.Time
	mu              sync.RWMutex
}

// SecureDataStore provides secure in-memory storage.
type SecureDataStore struct {
	mu                sync.RWMutex
	tenants           map[string]*TenantData
	users             map[string]*UserData
	customers         map[string]*CustomerData
	Customers         map[string]*SecureCustomerData // For cross-tenant tests
	tokens            map[string]*TokenData
	sessions          map[string]*SessionData
	InvalidatedTokens map[string]bool // For tracking invalidated tokens
}

// SecureCustomerData represents customer data with owner tracking.
type SecureCustomerData struct {
	ID        string
	TenantID  string
	OwnerID   string
	Code      string
	Name      string
	Email     string
	Status    string
	CreatedAt time.Time
}

// TenantData represents tenant data.
type TenantData struct {
	ID        string
	Name      string
	Slug      string
	Status    string
	CreatedAt time.Time
}

// UserData represents user data.
type UserData struct {
	ID           string
	TenantID     string
	Email        string
	PasswordHash string
	Status       string
	Role         string
	FailedLogins int
	LockedUntil  *time.Time
	CreatedAt    time.Time
}

// CustomerData represents customer data.
type CustomerData struct {
	ID        string
	TenantID  string
	Code      string
	Name      string
	Email     string
	Status    string
	CreatedAt time.Time
}

// TokenData represents token data.
type TokenData struct {
	Token     string
	UserID    string
	TenantID  string
	ExpiresAt time.Time
	Revoked   bool
}

// SessionData represents session data.
type SessionData struct {
	ID        string
	UserID    string
	TenantID  string
	IPAddress string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// RateLimiter provides simple rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request should be allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Clean old requests
	if times, ok := rl.requests[key]; ok {
		var valid []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		rl.requests[key] = valid
	}

	// Check limit
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// Add request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// NewSecurityTestServer creates a new security test server.
func NewSecurityTestServer() *SecurityTestServer {
	config := DefaultSecurityConfig()

	store := &SecureDataStore{
		tenants:           make(map[string]*TenantData),
		users:             make(map[string]*UserData),
		customers:         make(map[string]*CustomerData),
		Customers:         make(map[string]*SecureCustomerData),
		tokens:            make(map[string]*TokenData),
		sessions:          make(map[string]*SessionData),
		InvalidatedTokens: make(map[string]bool),
	}

	server := &SecurityTestServer{
		DataStore:      store,
		Config:         config,
		RateLimiter:    NewRateLimiter(config.MaxRequestsPerMinute, time.Minute),
		LoginAttempts:  make(map[string]int),
		LockedAccounts: make(map[string]time.Time),
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(server.securityHeadersMiddleware)
	r.Use(server.rateLimitMiddleware)

	// Auth routes
	r.Post("/api/v1/auth/register", server.registerHandler)
	r.Post("/api/v1/auth/login", server.loginHandler)
	r.Post("/api/v1/auth/logout", server.logoutHandler)
	r.Get("/api/v1/auth/me", server.getMeHandler)
	r.Post("/api/v1/auth/reset-password", server.resetPasswordHandler)

	// User routes
	r.Get("/api/v1/users", server.listUsersHandler)
	r.Get("/api/v1/users/{id}", server.getUserHandler)
	r.Put("/api/v1/users/{id}", server.updateUserHandler)
	r.Delete("/api/v1/users/{id}", server.deleteUserHandler)
	r.Get("/api/v1/users/me", server.getMeHandler)
	r.Put("/api/v1/users/me", server.updateMeHandler)

	// Customer routes
	r.Get("/api/v1/customers", server.listCustomersHandler)
	r.Post("/api/v1/customers", server.createCustomerHandler)
	r.Get("/api/v1/customers/{id}", server.getCustomerHandler)
	r.Put("/api/v1/customers/{id}", server.updateCustomerHandler)
	r.Delete("/api/v1/customers/{id}", server.deleteCustomerHandler)
	r.Get("/api/v1/customers/search", server.searchCustomersHandler)
	r.Post("/api/v1/customers/bulk-delete", server.bulkDeleteHandler)
	r.Put("/api/v1/customers/bulk-update", server.bulkUpdateHandler)
	r.Post("/api/v1/customers/export", server.exportHandler)
	r.Post("/api/v1/customers/{id}/contacts", server.createContactHandler)
	r.Get("/api/v1/customers/{id}/contacts", server.listContactsHandler)
	r.Get("/api/v1/customers/{id}/leads", server.listCustomerLeadsHandler)
	r.Post("/api/v1/customers/batch", server.batchGetHandler)
	r.Get("/api/v1/customers/count", server.countHandler)
	r.Get("/api/v1/customers/stats", server.statsHandler)

	// Admin routes
	r.Get("/api/v1/admin/users", server.adminListUsersHandler)
	r.Post("/api/v1/admin/impersonate/{id}", server.impersonateHandler)
	r.Post("/api/v1/admin/impersonate", server.impersonateByBodyHandler)
	r.Get("/api/v1/admin/settings", server.adminSettingsHandler)
	r.Get("/api/v1/admin/audit-logs", server.auditLogsHandler)
	r.Get("/api/v1/admin/system-config", server.systemConfigHandler)

	// System routes
	r.Get("/api/v1/system/health", server.healthHandler)
	r.Post("/api/v1/system/shutdown", server.shutdownHandler)

	// Leads routes
	r.Get("/api/v1/leads", server.listLeadsHandler)

	// Reports and dashboard
	r.Get("/api/v1/reports/customers", server.reportsHandler)
	r.Get("/api/v1/reports/summary", server.reportsHandler)
	r.Get("/api/v1/reports/analytics", server.reportsHandler)
	r.Get("/api/v1/dashboard/stats", server.reportsHandler)

	// Webhooks
	r.Get("/api/v1/webhooks", server.webhooksHandler)
	r.Post("/api/v1/webhooks/{id}/trigger", server.triggerWebhookHandler)

	// Audit logs
	r.Get("/api/v1/audit-logs", server.auditLogsHandler)

	// Settings and config
	r.Get("/api/v1/settings", server.settingsHandler)
	r.Get("/api/v1/tenants/{tenantId}/settings", server.tenantSettingsHandler)
	r.Put("/api/v1/tenants/{tenantId}/settings", server.updateTenantSettingsHandler)
	r.Get("/api/v1/config/{tenantId}", server.tenantConfigHandler)

	server.Router = r
	server.Server = httptest.NewServer(r)

	return server
}

// Close closes the test server.
func (s *SecurityTestServer) Close() {
	s.Server.Close()
}

// GetValidToken generates a valid token for a user.
func (s *SecurityTestServer) GetValidToken(email, tenantID string) string {
	token := generateSecureToken()

	// Find or create user
	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	var userID string
	var role string = "user"

	// Check if user exists
	for _, u := range s.DataStore.users {
		if u.Email == email && u.TenantID == tenantID {
			userID = u.ID
			role = u.Role
			break
		}
	}

	// Create user if not found
	if userID == "" {
		userID = uuid.New().String()
		if strings.Contains(email, "admin") {
			role = "admin"
		}
		hash, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
		s.DataStore.users[userID] = &UserData{
			ID:           userID,
			TenantID:     tenantID,
			Email:        email,
			PasswordHash: string(hash),
			Status:       "active",
			Role:         role,
			CreatedAt:    time.Now(),
		}
	}

	// Create token
	s.DataStore.tokens[token] = &TokenData{
		Token:     token,
		UserID:    userID,
		TenantID:  tenantID,
		ExpiresAt: time.Now().Add(s.Config.TokenExpiration),
	}

	return "Bearer " + token
}

// Seed seeds the data store with test data.
func (s *SecurityTestServer) Seed() {
	// Create tenants
	tenant1ID := uuid.New().String()
	tenant2ID := uuid.New().String()

	s.DataStore.tenants[tenant1ID] = &TenantData{
		ID:        tenant1ID,
		Name:      "Tenant One",
		Slug:      "tenant-one",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	s.DataStore.tenants[tenant2ID] = &TenantData{
		ID:        tenant2ID,
		Name:      "Tenant Two",
		Slug:      "tenant-two",
		Status:    "active",
		CreatedAt: time.Now(),
	}

	// Create users
	hash1, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	hash2, _ := bcrypt.GenerateFromPassword([]byte("Password456!"), bcrypt.DefaultCost)

	user1ID := uuid.New().String()
	user2ID := uuid.New().String()
	adminID := uuid.New().String()

	s.DataStore.users[user1ID] = &UserData{
		ID:           user1ID,
		TenantID:     tenant1ID,
		Email:        "user1@tenant1.com",
		PasswordHash: string(hash1),
		Status:       "active",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	s.DataStore.users[user2ID] = &UserData{
		ID:           user2ID,
		TenantID:     tenant2ID,
		Email:        "user2@tenant2.com",
		PasswordHash: string(hash2),
		Status:       "active",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	s.DataStore.users[adminID] = &UserData{
		ID:           adminID,
		TenantID:     tenant1ID,
		Email:        "admin@tenant1.com",
		PasswordHash: string(hash1),
		Status:       "active",
		Role:         "admin",
		CreatedAt:    time.Now(),
	}

	// Create customers
	for i := 0; i < 5; i++ {
		custID := uuid.New().String()
		s.DataStore.customers[custID] = &CustomerData{
			ID:        custID,
			TenantID:  tenant1ID,
			Code:      fmt.Sprintf("CUST1-%03d", i),
			Name:      fmt.Sprintf("Tenant1 Customer %d", i),
			Email:     fmt.Sprintf("customer%d@tenant1.com", i),
			Status:    "active",
			CreatedAt: time.Now(),
		}
	}

	for i := 0; i < 5; i++ {
		custID := uuid.New().String()
		s.DataStore.customers[custID] = &CustomerData{
			ID:        custID,
			TenantID:  tenant2ID,
			Code:      fmt.Sprintf("CUST2-%03d", i),
			Name:      fmt.Sprintf("Tenant2 Customer %d", i),
			Email:     fmt.Sprintf("customer%d@tenant2.com", i),
			Status:    "active",
			CreatedAt: time.Now(),
		}
	}
}

// ============================================================================
// Middleware
// ============================================================================

func (s *SecurityTestServer) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func (s *SecurityTestServer) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !s.RateLimiter.Allow(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// Handlers
// ============================================================================

func (s *SecurityTestServer) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TenantID string `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Validate password
	if len(req.Password) < s.Config.MinPasswordLength {
		writeError(w, http.StatusBadRequest, "password too short")
		return
	}

	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	// Check for existing user
	for _, u := range s.DataStore.users {
		if u.Email == req.Email && u.TenantID == req.TenantID {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := &UserData{
		ID:           uuid.New().String(),
		TenantID:     req.TenantID,
		Email:        req.Email,
		PasswordHash: string(hash),
		Status:       "active",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	s.DataStore.users[user.ID] = user

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

func (s *SecurityTestServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TenantID string `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	s.DataStore.mu.RLock()
	var foundUser *UserData
	for _, u := range s.DataStore.users {
		if u.Email == req.Email && u.TenantID == req.TenantID {
			foundUser = u
			break
		}
	}
	s.DataStore.mu.RUnlock()

	if foundUser == nil {
		// Use same error message to prevent user enumeration
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check if account is locked
	s.mu.RLock()
	if lockedUntil, ok := s.LockedAccounts[foundUser.ID]; ok && time.Now().Before(lockedUntil) {
		s.mu.RUnlock()
		writeError(w, http.StatusTooManyRequests, "account locked")
		return
	}
	s.mu.RUnlock()

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(req.Password)); err != nil {
		s.mu.Lock()
		s.LoginAttempts[foundUser.ID]++
		if s.LoginAttempts[foundUser.ID] >= s.Config.MaxLoginAttempts {
			s.LockedAccounts[foundUser.ID] = time.Now().Add(s.Config.LockoutDuration)
			s.LoginAttempts[foundUser.ID] = 0
		}
		s.mu.Unlock()
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Reset login attempts on successful login
	s.mu.Lock()
	delete(s.LoginAttempts, foundUser.ID)
	delete(s.LockedAccounts, foundUser.ID)
	s.mu.Unlock()

	// Generate token
	token := generateSecureToken()

	s.DataStore.mu.Lock()
	s.DataStore.tokens[token] = &TokenData{
		Token:     token,
		UserID:    foundUser.ID,
		TenantID:  foundUser.TenantID,
		ExpiresAt: time.Now().Add(s.Config.TokenExpiration),
	}
	s.DataStore.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   int(s.Config.TokenExpiration.Seconds()),
	})
}

func (s *SecurityTestServer) logoutHandler(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	s.DataStore.mu.Lock()
	if td, ok := s.DataStore.tokens[token]; ok {
		td.Revoked = true
	}
	s.DataStore.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (s *SecurityTestServer) getMeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":        user.ID,
		"email":     user.Email,
		"tenant_id": user.TenantID,
		"role":      user.Role,
	})
}

func (s *SecurityTestServer) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	var users []map[string]interface{}
	for _, u := range s.DataStore.users {
		// Only show users from same tenant
		if u.TenantID == user.TenantID {
			users = append(users, map[string]interface{}{
				"id":    u.ID,
				"email": u.Email,
				"role":  u.Role,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": users})
}

func (s *SecurityTestServer) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	targetUser, ok := s.DataStore.users[id]
	if !ok || targetUser.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":    targetUser.ID,
		"email": targetUser.Email,
		"role":  targetUser.Role,
	})
}

func (s *SecurityTestServer) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")

	// Only admins can update other users
	if id != user.ID && user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	targetUser, ok := s.DataStore.users[id]
	if !ok || targetUser.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	// Prevent role escalation
	if role, ok := req["role"].(string); ok && user.Role != "admin" {
		if role == "admin" {
			writeError(w, http.StatusForbidden, "cannot escalate privileges")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":    targetUser.ID,
		"email": targetUser.Email,
	})
}

func (s *SecurityTestServer) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	id := chi.URLParam(r, "id")

	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	targetUser, ok := s.DataStore.users[id]
	if !ok || targetUser.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	delete(s.DataStore.users, id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *SecurityTestServer) listCustomersHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	var customers []map[string]interface{}
	for _, c := range s.DataStore.customers {
		if c.TenantID == user.TenantID {
			customers = append(customers, map[string]interface{}{
				"id":    c.ID,
				"code":  c.Code,
				"name":  c.Name,
				"email": c.Email,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": customers})
}

func (s *SecurityTestServer) createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Code  string `json:"code"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	customer := &CustomerData{
		ID:        uuid.New().String(),
		TenantID:  user.TenantID,
		Code:      req.Code,
		Name:      req.Name,
		Email:     req.Email,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	s.DataStore.customers[customer.ID] = customer

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":   customer.ID,
		"code": customer.Code,
		"name": customer.Name,
	})
}

func (s *SecurityTestServer) getCustomerHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	customer, ok := s.DataStore.customers[id]
	if !ok || customer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	writeJSON(w, http.StatusOK, customer)
}

func (s *SecurityTestServer) updateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")

	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	customer, ok := s.DataStore.customers[id]
	if !ok || customer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	if name, ok := req["name"].(string); ok {
		customer.Name = name
	}

	writeJSON(w, http.StatusOK, customer)
}

func (s *SecurityTestServer) deleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := chi.URLParam(r, "id")

	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	customer, ok := s.DataStore.customers[id]
	if !ok || customer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	delete(s.DataStore.customers, id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *SecurityTestServer) searchCustomersHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query().Get("q")

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	var customers []map[string]interface{}
	for _, c := range s.DataStore.customers {
		if c.TenantID == user.TenantID {
			if query == "" || strings.Contains(strings.ToLower(c.Name), strings.ToLower(query)) {
				customers = append(customers, map[string]interface{}{
					"id":   c.ID,
					"name": c.Name,
				})
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": customers})
}

func (s *SecurityTestServer) adminListUsersHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	var users []map[string]interface{}
	for _, u := range s.DataStore.users {
		if u.TenantID == user.TenantID {
			users = append(users, map[string]interface{}{
				"id":     u.ID,
				"email":  u.Email,
				"role":   u.Role,
				"status": u.Status,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": users})
}

func (s *SecurityTestServer) impersonateHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	targetID := chi.URLParam(r, "id")

	s.DataStore.mu.RLock()
	targetUser, ok := s.DataStore.users[targetID]
	s.DataStore.mu.RUnlock()

	if !ok || targetUser.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// Generate impersonation token
	token := generateSecureToken()

	s.DataStore.mu.Lock()
	s.DataStore.tokens[token] = &TokenData{
		Token:     token,
		UserID:    targetUser.ID,
		TenantID:  targetUser.TenantID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	s.DataStore.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"impersonation_token": token,
		"expires_in":          900,
	})
}

func (s *SecurityTestServer) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Always return same response to prevent user enumeration
	writeJSON(w, http.StatusOK, map[string]string{"message": "If the email exists, a reset link will be sent"})
}

func (s *SecurityTestServer) updateMeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	// Prevent role escalation via self-update
	if _, hasRole := req["role"]; hasRole {
		writeError(w, http.StatusForbidden, "cannot modify own role")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

func (s *SecurityTestServer) bulkDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	deleted := 0
	for _, id := range req.IDs {
		if customer, ok := s.DataStore.customers[id]; ok && customer.TenantID == user.TenantID {
			delete(s.DataStore.customers, id)
			deleted++
		}
		if customer, ok := s.DataStore.Customers[id]; ok && customer.TenantID == user.TenantID {
			delete(s.DataStore.Customers, id)
			deleted++
		}
	}

	writeJSON(w, http.StatusOK, map[string]int{"deleted": deleted})
}

func (s *SecurityTestServer) bulkUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		IDs    []string `json:"ids"`
		Status string   `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.DataStore.mu.Lock()
	defer s.DataStore.mu.Unlock()

	updated := 0
	for _, id := range req.IDs {
		if customer, ok := s.DataStore.customers[id]; ok && customer.TenantID == user.TenantID {
			customer.Status = req.Status
			updated++
		}
		if customer, ok := s.DataStore.Customers[id]; ok && customer.TenantID == user.TenantID {
			customer.Status = req.Status
			updated++
		}
	}

	writeJSON(w, http.StatusOK, map[string]int{"updated": updated})
}

func (s *SecurityTestServer) exportHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	var result []map[string]interface{}
	for _, id := range req.IDs {
		if customer, ok := s.DataStore.customers[id]; ok && customer.TenantID == user.TenantID {
			result = append(result, map[string]interface{}{
				"id":   customer.ID,
				"name": customer.Name,
			})
		}
		if customer, ok := s.DataStore.Customers[id]; ok && customer.TenantID == user.TenantID {
			result = append(result, map[string]interface{}{
				"id":   customer.ID,
				"name": customer.Name,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func (s *SecurityTestServer) createContactHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	customerID := chi.URLParam(r, "id")

	s.DataStore.mu.RLock()
	customer, ok := s.DataStore.customers[customerID]
	secureCustomer, secureOk := s.DataStore.Customers[customerID]
	s.DataStore.mu.RUnlock()

	if ok && customer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	if secureOk && secureCustomer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	if !ok && !secureOk {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"id": uuid.New().String()})
}

func (s *SecurityTestServer) listContactsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	customerID := chi.URLParam(r, "id")

	s.DataStore.mu.RLock()
	customer, ok := s.DataStore.customers[customerID]
	secureCustomer, secureOk := s.DataStore.Customers[customerID]
	s.DataStore.mu.RUnlock()

	if ok && customer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	if secureOk && secureCustomer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	if !ok && !secureOk {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

func (s *SecurityTestServer) listCustomerLeadsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	customerID := chi.URLParam(r, "id")

	s.DataStore.mu.RLock()
	customer, ok := s.DataStore.customers[customerID]
	secureCustomer, secureOk := s.DataStore.Customers[customerID]
	s.DataStore.mu.RUnlock()

	if ok && customer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	if secureOk && secureCustomer.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	if !ok && !secureOk {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

func (s *SecurityTestServer) batchGetHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	var result []map[string]interface{}
	for _, id := range req.IDs {
		if customer, ok := s.DataStore.customers[id]; ok && customer.TenantID == user.TenantID {
			result = append(result, map[string]interface{}{
				"id":   customer.ID,
				"name": customer.Name,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func (s *SecurityTestServer) countHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	count := 0
	for _, c := range s.DataStore.customers {
		if c.TenantID == user.TenantID {
			count++
		}
	}

	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (s *SecurityTestServer) statsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	count := 0
	for _, c := range s.DataStore.customers {
		if c.TenantID == user.TenantID {
			count++
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_customers": count,
		"tenant_id":       user.TenantID,
	})
}

func (s *SecurityTestServer) impersonateByBodyHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		TenantID string `json:"tenant_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Cross-tenant impersonation is not allowed
	if req.TenantID != "" && req.TenantID != user.TenantID {
		writeError(w, http.StatusForbidden, "cross-tenant impersonation not allowed")
		return
	}

	s.DataStore.mu.RLock()
	targetUser, ok := s.DataStore.users[req.UserID]
	s.DataStore.mu.RUnlock()

	if !ok || targetUser.TenantID != user.TenantID {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	token := generateSecureToken()

	s.DataStore.mu.Lock()
	s.DataStore.tokens[token] = &TokenData{
		Token:     token,
		UserID:    targetUser.ID,
		TenantID:  targetUser.TenantID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	s.DataStore.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"impersonation_token": token,
	})
}

func (s *SecurityTestServer) adminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings": map[string]interface{}{
			"max_users":     100,
			"features":      []string{"crm", "sales"},
			"tenant_id":     user.TenantID,
		},
	})
}

func (s *SecurityTestServer) auditLogsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Only return audit logs for user's tenant
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": []map[string]interface{}{
			{"action": "login", "tenant_id": user.TenantID},
		},
	})
}

func (s *SecurityTestServer) systemConfigHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"config": "system config",
	})
}

func (s *SecurityTestServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (s *SecurityTestServer) shutdownHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if user.Role != "admin" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Don't actually shutdown, just return forbidden for non-super-admins
	writeError(w, http.StatusForbidden, "operation not permitted")
}

func (s *SecurityTestServer) listLeadsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":      []interface{}{},
		"tenant_id": user.TenantID,
	})
}

func (s *SecurityTestServer) reportsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	// Only return data for user's tenant
	var customers []map[string]interface{}
	for _, c := range s.DataStore.customers {
		if c.TenantID == user.TenantID {
			customers = append(customers, map[string]interface{}{
				"id":   c.ID,
				"name": c.Name,
			})
		}
	}
	for _, c := range s.DataStore.Customers {
		if c.TenantID == user.TenantID {
			customers = append(customers, map[string]interface{}{
				"id":   c.ID,
				"name": c.Name,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":      customers,
		"tenant_id": user.TenantID,
	})
}

func (s *SecurityTestServer) webhooksHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Only return webhooks for user's tenant
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":      []interface{}{},
		"tenant_id": user.TenantID,
	})
}

func (s *SecurityTestServer) triggerWebhookHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	webhookID := chi.URLParam(r, "id")

	// Check if webhook belongs to user's tenant (simplified)
	if strings.Contains(webhookID, "tenant-2") && user.TenantID != "tenant-2" {
		writeError(w, http.StatusNotFound, "webhook not found")
		return
	}
	if strings.Contains(webhookID, "tenant-1") && user.TenantID != "tenant-1" {
		writeError(w, http.StatusNotFound, "webhook not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "triggered"})
}

func (s *SecurityTestServer) settingsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings":  map[string]string{"theme": "light"},
		"tenant_id": user.TenantID,
	})
}

func (s *SecurityTestServer) tenantSettingsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tenantID := chi.URLParam(r, "tenantId")

	// Only allow access to own tenant's settings
	if tenantID != user.TenantID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"settings":  map[string]string{"theme": "light"},
		"tenant_id": tenantID,
	})
}

func (s *SecurityTestServer) updateTenantSettingsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tenantID := chi.URLParam(r, "tenantId")

	// Only allow access to own tenant's settings
	if tenantID != user.TenantID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *SecurityTestServer) tenantConfigHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.authenticateRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tenantID := chi.URLParam(r, "tenantId")

	// Only allow access to own tenant's config
	if tenantID != user.TenantID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"config":    map[string]string{"feature": "enabled"},
		"tenant_id": tenantID,
	})
}

// ============================================================================
// Helper Functions
// ============================================================================

func (s *SecurityTestServer) authenticateRequest(r *http.Request) (*UserData, error) {
	token := extractToken(r)
	if token == "" {
		return nil, fmt.Errorf("no token")
	}

	s.DataStore.mu.RLock()
	defer s.DataStore.mu.RUnlock()

	td, ok := s.DataStore.tokens[token]
	if !ok || td.Revoked || time.Now().After(td.ExpiresAt) {
		return nil, fmt.Errorf("invalid token")
	}

	user, ok := s.DataStore.users[td.UserID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

func generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// ============================================================================
// Security Test Client
// ============================================================================

// SecurityTestClient provides HTTP client for security testing.
type SecurityTestClient struct {
	Client  *http.Client
	baseURL string
	token   string
}

// NewSecurityTestClient creates a new security test client.
func NewSecurityTestClient(baseURL string) *SecurityTestClient {
	return &SecurityTestClient{
		Client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
		},
		baseURL: baseURL,
	}
}

// SetToken sets the authentication token.
func (c *SecurityTestClient) SetToken(token string) {
	c.token = token
}

// DoRequest performs an HTTP request (simplified version for injection tests).
func (c *SecurityTestClient) DoRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	return c.Client.Do(req)
}

// DoRequestWithHeaders performs an HTTP request with custom headers.
func (c *SecurityTestClient) DoRequestWithHeaders(method, path string, body interface{}, headers map[string]string) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp, respBody, nil
}

// DoAuthenticatedRequest performs an HTTP request with a specific token.
func (c *SecurityTestClient) DoAuthenticatedRequest(method, path, token string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", token)
	}

	return c.Client.Do(req)
}

// ============================================================================
// Test Main
// ============================================================================

func TestMain(m *testing.M) {
	if testing.Short() {
		fmt.Println("Skipping security tests in short mode")
		os.Exit(0)
	}

	os.Exit(m.Run())
}
