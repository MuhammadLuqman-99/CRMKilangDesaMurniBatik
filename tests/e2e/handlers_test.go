// Package e2e contains handler implementations for E2E tests.
package e2e

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================================
// In-Memory Data Store for E2E Tests
// ============================================================================

var (
	dataStore = &TestDataStore{
		tenants:       make(map[string]*TenantData),
		users:         make(map[string]*UserData),
		customers:     make(map[string]*CustomerData),
		leads:         make(map[string]*LeadData),
		opportunities: make(map[string]*OpportunityData),
		deals:         make(map[string]*DealData),
		pipelines:     make(map[string]*PipelineData),
		roles:         make(map[string]*RoleData),
		tokens:        make(map[string]*TokenData),
	}
)

type TestDataStore struct {
	mu            sync.RWMutex
	tenants       map[string]*TenantData
	users         map[string]*UserData
	customers     map[string]*CustomerData
	leads         map[string]*LeadData
	opportunities map[string]*OpportunityData
	deals         map[string]*DealData
	pipelines     map[string]*PipelineData
	roles         map[string]*RoleData
	tokens        map[string]*TokenData
}

type TenantData struct {
	ID        string
	Name      string
	Slug      string
	Status    string
	Plan      string
	Settings  map[string]interface{}
	CreatedAt time.Time
}

type UserData struct {
	ID              string
	TenantID        string
	Email           string
	PasswordHash    string
	FirstName       string
	LastName        string
	Status          string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	Roles           []string
}

type CustomerData struct {
	ID        string
	TenantID  string
	Code      string
	Name      string
	Type      string
	Status    string
	Email     map[string]interface{}
	Contacts  []*ContactData
	Version   int
	CreatedAt time.Time
}

type ContactData struct {
	ID         string
	CustomerID string
	FirstName  string
	LastName   string
	Email      map[string]interface{}
	IsPrimary  bool
}

type LeadData struct {
	ID                    string
	TenantID              string
	CustomerID            string
	CompanyName           string
	ContactName           string
	ContactEmail          string
	Source                string
	Status                string
	Score                 int
	AssignedTo            string
	ConvertedOpportunityID string
	CreatedAt             time.Time
}

type OpportunityData struct {
	ID            string
	TenantID      string
	CustomerID    string
	LeadID        string
	PipelineID    string
	StageID       string
	Name          string
	ValueAmount   int64
	ValueCurrency string
	Probability   int
	Status        string
	CreatedAt     time.Time
}

type DealData struct {
	ID            string
	TenantID      string
	OpportunityID string
	CustomerID    string
	Name          string
	ValueAmount   int64
	ValueCurrency string
	Status        string
	CreatedAt     time.Time
}

type PipelineData struct {
	ID        string
	TenantID  string
	Name      string
	IsDefault bool
	Status    string
	Stages    []*StageData
	CreatedAt time.Time
}

type StageData struct {
	ID          string
	PipelineID  string
	Name        string
	Type        string
	Order       int
	Probability int
}

type RoleData struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Permissions []string
	IsSystem    bool
}

type TokenData struct {
	UserID    string
	TenantID  string
	Token     string
	ExpiresAt time.Time
}

// ============================================================================
// Health Handler
// ============================================================================

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// ============================================================================
// Auth Handlers
// ============================================================================

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		TenantID  string `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.TenantID == "" {
		writeError(w, http.StatusBadRequest, "email, password, and tenant_id are required")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	// Check tenant exists
	if _, ok := dataStore.tenants[req.TenantID]; !ok {
		writeError(w, http.StatusBadRequest, "tenant not found")
		return
	}

	// Check email uniqueness within tenant
	for _, u := range dataStore.users {
		if u.TenantID == req.TenantID && u.Email == req.Email {
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
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       "pending_verification",
		CreatedAt:    time.Now(),
	}

	dataStore.users[user.ID] = user

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"status":     user.Status,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TenantID string `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var foundUser *UserData
	for _, u := range dataStore.users {
		if u.TenantID == req.TenantID && u.Email == req.Email {
			foundUser = u
			break
		}
	}

	if foundUser == nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken := generateToken()
	refreshToken := generateToken()

	dataStore.tokens[accessToken] = &TokenData{
		UserID:    foundUser.ID,
		TenantID:  foundUser.TenantID,
		Token:     accessToken,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"user": map[string]interface{}{
			"id":         foundUser.ID,
			"email":      foundUser.Email,
			"first_name": foundUser.FirstName,
			"last_name":  foundUser.LastName,
		},
	})
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := generateToken()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func verifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserFromToken(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	if user, ok := dataStore.users[userID]; ok {
		now := time.Now()
		user.EmailVerifiedAt = &now
		user.Status = "active"
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "email verified"})
}

func resendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "verification email sent"})
}

func forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "reset email sent"})
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset"})
}

func getMeHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserFromToken(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	user, ok := dataStore.users[userID]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"status":     user.Status,
	})
}

func updateMeHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserFromToken(r)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	user, ok := dataStore.users[userID]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if fn, ok := req["first_name"].(string); ok {
		user.FirstName = fn
	}
	if ln, ok := req["last_name"].(string); ok {
		user.LastName = ln
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"status":     user.Status,
	})
}

func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "password changed"})
}

// ============================================================================
// Helper Functions
// ============================================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func getUserFromToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	token := strings.TrimPrefix(auth, "Bearer ")

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	if td, ok := dataStore.tokens[token]; ok && td.ExpiresAt.After(time.Now()) {
		return td.UserID
	}

	return ""
}

func getTenantFromHeader(r *http.Request) string {
	return r.Header.Get("X-Tenant-ID")
}

func getIDParam(r *http.Request) string {
	return chi.URLParam(r, "id")
}
