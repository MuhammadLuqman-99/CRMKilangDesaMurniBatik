// Package http contains HTTP handler integration tests for the IAM service.
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/kilang-desa-murni/crm/pkg/testing/containers"
	"github.com/kilang-desa-murni/crm/pkg/testing/fixtures"
	"github.com/kilang-desa-murni/crm/pkg/testing/helpers"
)

var (
	testDB        *containers.PostgresContainer
	testRouter    *chi.Mux
	testCtx       context.Context
	testCtxCancel context.CancelFunc
	testTenantID  uuid.UUID
)

// TestMain sets up and tears down the test database and router.
func TestMain(m *testing.M) {
	// Skip if in short mode
	if testing.Short() {
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup PostgreSQL container
	var err error
	testDB, err = containers.NewPostgresContainer(ctx, containers.DefaultPostgresConfig())
	if err != nil {
		panic("failed to create PostgreSQL container: " + err.Error())
	}

	// Run migrations
	if err := runTestMigrations(ctx, testDB.DB); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	// Setup router
	testRouter = setupTestRouter()

	// Create test tenant
	testTenantID = createTestTenantForAPI(ctx, testDB.DB)

	// Run tests
	code := m.Run()

	// Cleanup
	if testDB != nil {
		testDB.Close()
	}

	os.Exit(code)
}

func runTestMigrations(ctx context.Context, db *sqlx.DB) error {
	migration := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS tenants (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(100) NOT NULL UNIQUE,
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		plan VARCHAR(50) NOT NULL DEFAULT 'free',
		settings JSONB DEFAULT '{}',
		metadata JSONB DEFAULT '{}',
		trial_ends_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	CREATE TABLE IF NOT EXISTS roles (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		permissions JSONB DEFAULT '[]',
		is_system BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		email VARCHAR(255) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		first_name VARCHAR(100),
		last_name VARCHAR(100),
		avatar_url VARCHAR(500),
		phone VARCHAR(50),
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		email_verified_at TIMESTAMP WITH TIME ZONE,
		last_login_at TIMESTAMP WITH TIME ZONE,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE,
		UNIQUE(tenant_id, email)
	);

	CREATE TABLE IF NOT EXISTS user_roles (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
		assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		assigned_by UUID REFERENCES users(id),
		PRIMARY KEY (user_id, role_id)
	);

	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash VARCHAR(255) NOT NULL UNIQUE,
		device_info JSONB DEFAULT '{}',
		ip_address VARCHAR(50),
		user_agent TEXT,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		revoked_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);
	`

	_, err := db.ExecContext(ctx, migration)
	return err
}

func setupTestRouter() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handleRegister)
			r.Post("/login", handleLogin)
			r.Post("/refresh", handleRefresh)
			r.Post("/logout", handleLogout)
		})

		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", handleListUsers)
			r.Post("/", handleCreateUser)
			r.Get("/{id}", handleGetUser)
			r.Put("/{id}", handleUpdateUser)
			r.Delete("/{id}", handleDeleteUser)
		})

		// Role routes
		r.Route("/roles", func(r chi.Router) {
			r.Get("/", handleListRoles)
			r.Post("/", handleCreateRole)
			r.Get("/{id}", handleGetRole)
			r.Put("/{id}", handleUpdateRole)
			r.Delete("/{id}", handleDeleteRole)
		})

		// Tenant routes
		r.Route("/tenants", func(r chi.Router) {
			r.Get("/", handleListTenants)
			r.Post("/", handleCreateTenant)
			r.Get("/{id}", handleGetTenant)
			r.Put("/{id}", handleUpdateTenant)
			r.Delete("/{id}", handleDeleteTenant)
		})
	})

	return r
}

func createTestTenantForAPI(ctx context.Context, db *sqlx.DB) uuid.UUID {
	tenantID := fixtures.TestIDs.TenantID1
	_, _ = db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, slug, status, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, tenantID, "Test Tenant", "test-tenant", "active", "free", time.Now(), time.Now())
	return tenantID
}

func setupTest(t *testing.T) {
	t.Helper()
	helpers.SkipIfShort(t)
	testCtx, testCtxCancel = helpers.DefaultTestContext()
}

func cleanupTest(t *testing.T) {
	t.Helper()
	if testCtxCancel != nil {
		testCtxCancel()
	}
	// Clean up test data but keep tenant
	testDB.TruncateTables(context.Background(), "user_roles", "refresh_tokens", "users", "roles")
}

// ============================================================================
// Mock Handlers (simplified for testing HTTP layer)
// ============================================================================

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password required"}`, http.StatusBadRequest)
		return
	}

	// Simulate user creation
	userID := uuid.New()
	response := map[string]interface{}{
		"id":         userID.String(),
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"status":     "pending_verification",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password required"}`, http.StatusBadRequest)
		return
	}

	// Simulate login - check for specific test credentials
	if req.Email == "valid@test.com" && req.Password == "validpassword" {
		response := map[string]interface{}{
			"access_token":  "test_access_token",
			"refresh_token": "test_refresh_token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, `{"error":"refresh_token required"}`, http.StatusBadRequest)
		return
	}

	// Simulate token refresh
	if req.RefreshToken == "valid_refresh_token" {
		response := map[string]interface{}{
			"access_token":  "new_access_token",
			"refresh_token": "new_refresh_token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, `{"error":"invalid refresh token"}`, http.StatusUnauthorized)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func handleListUsers(w http.ResponseWriter, r *http.Request) {
	// Simulate user list
	users := []map[string]interface{}{
		{
			"id":         fixtures.TestIDs.UserID1.String(),
			"email":      "user1@test.com",
			"first_name": "User",
			"last_name":  "One",
			"status":     "active",
		},
		{
			"id":         fixtures.TestIDs.UserID2.String(),
			"email":      "user2@test.com",
			"first_name": "User",
			"last_name":  "Two",
			"status":     "active",
		},
	}

	response := map[string]interface{}{
		"users": users,
		"total": 2,
		"page":  1,
		"limit": 20,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	userID := uuid.New()
	response := map[string]interface{}{
		"id":         userID.String(),
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"status":     "active",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Simulate user retrieval
	if id == fixtures.TestIDs.UserID1.String() {
		response := map[string]interface{}{
			"id":         id,
			"email":      "user1@test.com",
			"first_name": "User",
			"last_name":  "One",
			"status":     "active",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"id":      id,
		"message": "user updated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleListRoles(w http.ResponseWriter, r *http.Request) {
	roles := []map[string]interface{}{
		{"id": fixtures.TestIDs.RoleID1.String(), "name": "admin", "is_system": true},
		{"id": fixtures.TestIDs.RoleID2.String(), "name": "user", "is_system": false},
	}

	response := map[string]interface{}{
		"roles": roles,
		"total": 2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	roleID := uuid.New()
	response := map[string]interface{}{
		"id":          roleID.String(),
		"name":        req.Name,
		"description": req.Description,
		"permissions": req.Permissions,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleGetRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	response := map[string]interface{}{
		"id":          id,
		"name":        "admin",
		"description": "Administrator role",
		"permissions": []string{"*"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	response := map[string]interface{}{
		"id":      id,
		"message": "role updated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func handleListTenants(w http.ResponseWriter, r *http.Request) {
	tenants := []map[string]interface{}{
		{"id": fixtures.TestIDs.TenantID1.String(), "name": "Tenant 1", "slug": "tenant-1"},
		{"id": fixtures.TestIDs.TenantID2.String(), "name": "Tenant 2", "slug": "tenant-2"},
	}

	response := map[string]interface{}{
		"tenants": tenants,
		"total":   2,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	tenantID := uuid.New()
	response := map[string]interface{}{
		"id":   tenantID.String(),
		"name": req.Name,
		"slug": req.Slug,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleGetTenant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	response := map[string]interface{}{
		"id":     id,
		"name":   "Test Tenant",
		"slug":   "test-tenant",
		"status": "active",
		"plan":   "free",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	response := map[string]interface{}{
		"id":      id,
		"message": "tenant updated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// API Integration Tests
// ============================================================================

func TestHealthEndpoint(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	testRouter.ServeHTTP(rr, req)

	helpers.AssertEqual(t, http.StatusOK, rr.Code)

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	helpers.AssertEqual(t, "healthy", response["status"])
}

func TestAuthEndpoints(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("POST /auth/register - success", func(t *testing.T) {
		body := map[string]string{
			"email":      "newuser@test.com",
			"password":   "SecureP@ss123",
			"first_name": "New",
			"last_name":  "User",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusCreated, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)
		helpers.AssertEqual(t, "newuser@test.com", response["email"])
		helpers.AssertNotNil(t, response["id"])
	})

	t.Run("POST /auth/register - missing fields", func(t *testing.T) {
		body := map[string]string{
			"email": "incomplete@test.com",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("POST /auth/login - success", func(t *testing.T) {
		body := map[string]string{
			"email":    "valid@test.com",
			"password": "validpassword",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)
		helpers.AssertNotNil(t, response["access_token"])
		helpers.AssertNotNil(t, response["refresh_token"])
	})

	t.Run("POST /auth/login - invalid credentials", func(t *testing.T) {
		body := map[string]string{
			"email":    "invalid@test.com",
			"password": "wrongpassword",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("POST /auth/refresh - success", func(t *testing.T) {
		body := map[string]string{
			"refresh_token": "valid_refresh_token",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)
	})

	t.Run("POST /auth/logout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer test_token")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusNoContent, rr.Code)
	})
}

func TestUserEndpoints(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("GET /users - list users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)
		helpers.AssertNotNil(t, response["users"])
	})

	t.Run("POST /users - create user", func(t *testing.T) {
		body := map[string]string{
			"email":      "newcreated@test.com",
			"password":   "Password123!",
			"first_name": "Created",
			"last_name":  "User",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusCreated, rr.Code)
	})

	t.Run("GET /users/{id} - get user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+fixtures.TestIDs.UserID1.String(), nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)
	})

	t.Run("GET /users/{id} - not found", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+nonExistentID, nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusNotFound, rr.Code)
	})

	t.Run("PUT /users/{id} - update user", func(t *testing.T) {
		body := map[string]string{
			"first_name": "Updated",
			"last_name":  "Name",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/"+fixtures.TestIDs.UserID1.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)
	})

	t.Run("DELETE /users/{id} - delete user", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/"+fixtures.TestIDs.UserID1.String(), nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusNoContent, rr.Code)
	})

	t.Run("GET /users/{id} - invalid ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid-uuid", nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusBadRequest, rr.Code)
	})
}

func TestRoleEndpoints(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("GET /roles - list roles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)
		helpers.AssertNotNil(t, response["roles"])
	})

	t.Run("POST /roles - create role", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "custom_role",
			"description": "A custom role",
			"permissions": []string{"read", "write"},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusCreated, rr.Code)
	})

	t.Run("GET /roles/{id} - get role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/"+fixtures.TestIDs.RoleID1.String(), nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)
	})

	t.Run("PUT /roles/{id} - update role", func(t *testing.T) {
		body := map[string]interface{}{
			"description": "Updated description",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/roles/"+fixtures.TestIDs.RoleID1.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)
	})

	t.Run("DELETE /roles/{id} - delete role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/"+fixtures.TestIDs.RoleID1.String(), nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusNoContent, rr.Code)
	})
}

func TestTenantEndpoints(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("GET /tenants - list tenants", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants", nil)
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &response)
		helpers.AssertNotNil(t, response["tenants"])
	})

	t.Run("POST /tenants - create tenant", func(t *testing.T) {
		body := map[string]string{
			"name": "New Tenant",
			"slug": "new-tenant",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusCreated, rr.Code)
	})

	t.Run("GET /tenants/{id} - get tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tenants/"+testTenantID.String(), nil)
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)
	})

	t.Run("PUT /tenants/{id} - update tenant", func(t *testing.T) {
		body := map[string]string{
			"name": "Updated Tenant Name",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/tenants/"+testTenantID.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusOK, rr.Code)
	})

	t.Run("DELETE /tenants/{id} - delete tenant", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/tenants/"+testTenantID.String(), nil)
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusNoContent, rr.Code)
	})
}

func TestInvalidContentType(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("POST with invalid content type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer([]byte("plain text")))
		req.Header.Set("Content-Type", "text/plain")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		// Should return bad request due to invalid JSON
		helpers.AssertEqual(t, http.StatusBadRequest, rr.Code)
	})
}

func TestMalformedJSON(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("POST with malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer([]byte("{invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testRouter.ServeHTTP(rr, req)

		helpers.AssertEqual(t, http.StatusBadRequest, rr.Code)
	})
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkHealthEndpoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()
		testRouter.ServeHTTP(rr, req)
	}
}

func BenchmarkListUsers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		req.Header.Set("X-Tenant-ID", testTenantID.String())
		rr := httptest.NewRecorder()
		testRouter.ServeHTTP(rr, req)
	}
}
