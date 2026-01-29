// Package e2e contains end-to-end integration tests for the CRM system.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kilang-desa-murni/crm/pkg/testing/containers"
	"github.com/kilang-desa-murni/crm/pkg/testing/fixtures"
	"github.com/kilang-desa-murni/crm/pkg/testing/helpers"
)

// ============================================================================
// Global Test Infrastructure
// ============================================================================

var (
	// Database containers
	postgresContainer *containers.PostgresContainer
	mongoContainer    *containers.MongoDBContainer
	redisContainer    *containers.RedisContainer
	rabbitContainer   *containers.RabbitMQContainer

	// Test servers
	iamServer      *httptest.Server
	customerServer *httptest.Server
	salesServer    *httptest.Server

	// Test client
	httpClient *http.Client

	// Synchronization
	setupOnce sync.Once
	setupErr  error
)

// TestMain sets up all test infrastructure for E2E tests.
func TestMain(m *testing.M) {
	if testing.Short() {
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Setup all containers
	if err := setupTestInfrastructure(ctx); err != nil {
		fmt.Printf("Failed to setup test infrastructure: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	teardownTestInfrastructure()

	os.Exit(code)
}

func setupTestInfrastructure(ctx context.Context) error {
	setupOnce.Do(func() {
		// Setup PostgreSQL
		postgresContainer, setupErr = containers.NewPostgresContainer(ctx, containers.DefaultPostgresConfig())
		if setupErr != nil {
			return
		}

		// Run IAM migrations
		if setupErr = runIAMMigrations(ctx, postgresContainer.DB); setupErr != nil {
			return
		}

		// Run Sales migrations
		if setupErr = runSalesMigrations(ctx, postgresContainer.DB); setupErr != nil {
			return
		}

		// Setup MongoDB
		mongoContainer, setupErr = containers.NewMongoDBContainer(ctx, containers.DefaultMongoDBConfig())
		if setupErr != nil {
			return
		}

		// Setup MongoDB collections
		if setupErr = mongoContainer.SetupCustomerCollections(ctx); setupErr != nil {
			return
		}

		// Setup Redis
		redisContainer, setupErr = containers.NewRedisContainer(ctx, containers.DefaultRedisConfig())
		if setupErr != nil {
			return
		}

		// Setup RabbitMQ
		rabbitContainer, setupErr = containers.NewRabbitMQContainer(ctx, containers.DefaultRabbitMQConfig())
		if setupErr != nil {
			return
		}

		if setupErr = rabbitContainer.SetupTestQueues(); setupErr != nil {
			return
		}

		// Setup HTTP servers
		iamServer = setupIAMServer()
		customerServer = setupCustomerServer()
		salesServer = setupSalesServer()

		// Setup HTTP client
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	})

	return setupErr
}

func teardownTestInfrastructure() {
	if iamServer != nil {
		iamServer.Close()
	}
	if customerServer != nil {
		customerServer.Close()
	}
	if salesServer != nil {
		salesServer.Close()
	}
	if postgresContainer != nil {
		postgresContainer.Close()
	}
	if mongoContainer != nil {
		mongoContainer.Close(context.Background())
	}
	if redisContainer != nil {
		redisContainer.Close()
	}
	if rabbitContainer != nil {
		rabbitContainer.Close()
	}
}

// ============================================================================
// Database Migrations
// ============================================================================

func runIAMMigrations(ctx context.Context, db *sqlx.DB) error {
	migration := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	-- Tenants
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

	-- Roles
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

	-- Users
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		email VARCHAR(255) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		first_name VARCHAR(100),
		last_name VARCHAR(100),
		avatar_url VARCHAR(500),
		phone VARCHAR(50),
		status VARCHAR(50) NOT NULL DEFAULT 'pending_verification',
		email_verified_at TIMESTAMP WITH TIME ZONE,
		last_login_at TIMESTAMP WITH TIME ZONE,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE,
		UNIQUE(tenant_id, email)
	);

	-- User Roles
	CREATE TABLE IF NOT EXISTS user_roles (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
		assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		assigned_by UUID REFERENCES users(id),
		PRIMARY KEY (user_id, role_id)
	);

	-- Refresh Tokens
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

	-- Email Verification Tokens
	CREATE TABLE IF NOT EXISTS email_verification_tokens (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash VARCHAR(255) NOT NULL UNIQUE,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		used_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	-- Password Reset Tokens
	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash VARCHAR(255) NOT NULL UNIQUE,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		used_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	-- Audit Logs
	CREATE TABLE IF NOT EXISTS audit_logs (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		user_id UUID REFERENCES users(id) ON DELETE SET NULL,
		action VARCHAR(100) NOT NULL,
		entity_type VARCHAR(100) NOT NULL,
		entity_id UUID,
		old_values JSONB,
		new_values JSONB,
		ip_address VARCHAR(50),
		user_agent TEXT,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	-- Sessions
	CREATE TABLE IF NOT EXISTS sessions (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash VARCHAR(255) NOT NULL UNIQUE,
		ip_address VARCHAR(50),
		user_agent TEXT,
		last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	-- Insert default system roles
	INSERT INTO roles (id, name, description, permissions, is_system) VALUES
		(uuid_generate_v4(), 'super_admin', 'Super Administrator', '["*"]', true),
		(uuid_generate_v4(), 'admin', 'Administrator', '["users:*", "roles:*", "settings:*"]', true),
		(uuid_generate_v4(), 'manager', 'Manager', '["users:read", "customers:*", "sales:*"]', true),
		(uuid_generate_v4(), 'sales_rep', 'Sales Representative', '["customers:read", "customers:create", "customers:update", "leads:*", "opportunities:*"]', true),
		(uuid_generate_v4(), 'viewer', 'Read-only access', '["*:read"]', true)
	ON CONFLICT DO NOTHING;

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
	CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
	CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
	CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	`

	_, err := db.ExecContext(ctx, migration)
	return err
}

func runSalesMigrations(ctx context.Context, db *sqlx.DB) error {
	migration := `
	-- Pipelines
	CREATE TABLE IF NOT EXISTS pipelines (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		is_default BOOLEAN NOT NULL DEFAULT false,
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		settings JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	-- Pipeline Stages
	CREATE TABLE IF NOT EXISTS pipeline_stages (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		type VARCHAR(50) NOT NULL DEFAULT 'open',
		stage_order INT NOT NULL,
		probability INT NOT NULL DEFAULT 0,
		color VARCHAR(20),
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	-- Leads
	CREATE TABLE IF NOT EXISTS leads (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		customer_id UUID,
		source VARCHAR(100),
		status VARCHAR(50) NOT NULL DEFAULT 'new',
		score INT DEFAULT 0,
		company_name VARCHAR(255),
		contact_name VARCHAR(255),
		contact_email VARCHAR(255),
		contact_phone VARCHAR(100),
		website VARCHAR(500),
		industry VARCHAR(100),
		company_size VARCHAR(50),
		budget VARCHAR(50),
		timeline VARCHAR(50),
		notes TEXT,
		assigned_to UUID,
		qualified_at TIMESTAMP WITH TIME ZONE,
		qualified_by UUID,
		disqualified_at TIMESTAMP WITH TIME ZONE,
		disqualified_reason TEXT,
		converted_at TIMESTAMP WITH TIME ZONE,
		converted_by UUID,
		converted_opportunity_id UUID,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	-- Opportunities
	CREATE TABLE IF NOT EXISTS opportunities (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		customer_id UUID NOT NULL,
		lead_id UUID REFERENCES leads(id),
		pipeline_id UUID NOT NULL REFERENCES pipelines(id),
		stage_id UUID NOT NULL REFERENCES pipeline_stages(id),
		name VARCHAR(255) NOT NULL,
		description TEXT,
		value_amount BIGINT NOT NULL DEFAULT 0,
		value_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
		probability INT NOT NULL DEFAULT 0,
		expected_close TIMESTAMP WITH TIME ZONE,
		actual_close TIMESTAMP WITH TIME ZONE,
		assigned_to UUID,
		status VARCHAR(50) NOT NULL DEFAULT 'open',
		won_at TIMESTAMP WITH TIME ZONE,
		won_reason TEXT,
		lost_at TIMESTAMP WITH TIME ZONE,
		lost_reason TEXT,
		competitor TEXT,
		next_step TEXT,
		notes TEXT,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	-- Opportunity Stage History
	CREATE TABLE IF NOT EXISTS opportunity_stage_history (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		opportunity_id UUID NOT NULL REFERENCES opportunities(id) ON DELETE CASCADE,
		from_stage_id UUID REFERENCES pipeline_stages(id),
		to_stage_id UUID NOT NULL REFERENCES pipeline_stages(id),
		changed_by UUID,
		notes TEXT,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	-- Deals
	CREATE TABLE IF NOT EXISTS deals (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		opportunity_id UUID NOT NULL REFERENCES opportunities(id),
		customer_id UUID NOT NULL,
		deal_number VARCHAR(50),
		name VARCHAR(255) NOT NULL,
		description TEXT,
		value_amount BIGINT NOT NULL DEFAULT 0,
		value_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		closed_at TIMESTAMP WITH TIME ZONE,
		assigned_to UUID,
		contract_start TIMESTAMP WITH TIME ZONE,
		contract_end TIMESTAMP WITH TIME ZONE,
		payment_terms VARCHAR(100),
		notes TEXT,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	-- Deal Line Items
	CREATE TABLE IF NOT EXISTS deal_line_items (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
		product_name VARCHAR(255) NOT NULL,
		description TEXT,
		quantity INT NOT NULL DEFAULT 1,
		unit_price BIGINT NOT NULL DEFAULT 0,
		discount_percent INT DEFAULT 0,
		tax_percent INT DEFAULT 0,
		total_amount BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	-- Outbox
	CREATE TABLE IF NOT EXISTS outbox (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		event_type VARCHAR(100) NOT NULL,
		aggregate_id UUID NOT NULL,
		aggregate_type VARCHAR(100) NOT NULL,
		payload JSONB NOT NULL,
		published BOOLEAN NOT NULL DEFAULT false,
		published_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_pipelines_tenant_id ON pipelines(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_leads_tenant_id ON leads(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
	CREATE INDEX IF NOT EXISTS idx_leads_assigned_to ON leads(assigned_to);
	CREATE INDEX IF NOT EXISTS idx_opportunities_tenant_id ON opportunities(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_opportunities_pipeline_id ON opportunities(pipeline_id);
	CREATE INDEX IF NOT EXISTS idx_opportunities_stage_id ON opportunities(stage_id);
	CREATE INDEX IF NOT EXISTS idx_opportunities_status ON opportunities(status);
	CREATE INDEX IF NOT EXISTS idx_deals_tenant_id ON deals(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_deals_opportunity_id ON deals(opportunity_id);
	CREATE INDEX IF NOT EXISTS idx_deals_status ON deals(status);
	`

	_, err := db.ExecContext(ctx, migration)
	return err
}

// ============================================================================
// HTTP Server Setup
// ============================================================================

func setupIAMServer() *httptest.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", healthHandler)

	// Auth routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", registerHandler)
		r.Post("/login", loginHandler)
		r.Post("/refresh", refreshHandler)
		r.Post("/logout", logoutHandler)
		r.Post("/verify-email", verifyEmailHandler)
		r.Post("/resend-verification", resendVerificationHandler)
		r.Post("/forgot-password", forgotPasswordHandler)
		r.Post("/reset-password", resetPasswordHandler)
		r.Get("/me", getMeHandler)
		r.Put("/me", updateMeHandler)
		r.Put("/me/password", changePasswordHandler)
	})

	// User routes
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Get("/", listUsersHandler)
		r.Post("/", createUserHandler)
		r.Get("/{id}", getUserHandler)
		r.Put("/{id}", updateUserHandler)
		r.Delete("/{id}", deleteUserHandler)
		r.Post("/{id}/activate", activateUserHandler)
		r.Post("/{id}/suspend", suspendUserHandler)
		r.Get("/{id}/roles", getUserRolesHandler)
		r.Post("/{id}/roles", assignRoleHandler)
		r.Delete("/{id}/roles/{roleId}", removeRoleHandler)
	})

	// Role routes
	r.Route("/api/v1/roles", func(r chi.Router) {
		r.Get("/", listRolesHandler)
		r.Post("/", createRoleHandler)
		r.Get("/system", getSystemRolesHandler)
		r.Get("/{id}", getRoleHandler)
		r.Put("/{id}", updateRoleHandler)
		r.Delete("/{id}", deleteRoleHandler)
	})

	// Tenant routes
	r.Route("/api/v1/tenants", func(r chi.Router) {
		r.Get("/", listTenantsHandler)
		r.Post("/", createTenantHandler)
		r.Get("/{id}", getTenantHandler)
		r.Put("/{id}", updateTenantHandler)
		r.Delete("/{id}", deleteTenantHandler)
		r.Put("/{id}/status", updateTenantStatusHandler)
		r.Put("/{id}/plan", updateTenantPlanHandler)
		r.Get("/{id}/stats", getTenantStatsHandler)
	})

	return httptest.NewServer(r)
}

func setupCustomerServer() *httptest.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", healthHandler)

	// Customer routes
	r.Route("/api/v1/customers", func(r chi.Router) {
		r.Get("/", listCustomersHandler)
		r.Post("/", createCustomerHandler)
		r.Get("/search", searchCustomersHandler)
		r.Post("/import", importCustomersHandler)
		r.Get("/export", exportCustomersHandler)
		r.Get("/{id}", getCustomerHandler)
		r.Put("/{id}", updateCustomerHandler)
		r.Delete("/{id}", deleteCustomerHandler)
		r.Post("/{id}/restore", restoreCustomerHandler)
		r.Post("/{id}/activate", activateCustomerHandler)
		r.Post("/{id}/deactivate", deactivateCustomerHandler)

		// Contact routes
		r.Get("/{id}/contacts", listContactsHandler)
		r.Post("/{id}/contacts", createContactHandler)
		r.Get("/{id}/contacts/{contactId}", getContactHandler)
		r.Put("/{id}/contacts/{contactId}", updateContactHandler)
		r.Delete("/{id}/contacts/{contactId}", deleteContactHandler)
		r.Post("/{id}/contacts/{contactId}/primary", setPrimaryContactHandler)

		// Activity routes
		r.Get("/{id}/activities", listActivitiesHandler)
		r.Post("/{id}/activities", createActivityHandler)

		// Note routes
		r.Get("/{id}/notes", listNotesHandler)
		r.Post("/{id}/notes", createNoteHandler)
		r.Put("/{id}/notes/{noteId}", updateNoteHandler)
		r.Delete("/{id}/notes/{noteId}", deleteNoteHandler)
		r.Post("/{id}/notes/{noteId}/pin", pinNoteHandler)
	})

	// Segment routes
	r.Route("/api/v1/segments", func(r chi.Router) {
		r.Get("/", listSegmentsHandler)
		r.Post("/", createSegmentHandler)
		r.Get("/{id}", getSegmentHandler)
		r.Put("/{id}", updateSegmentHandler)
		r.Delete("/{id}", deleteSegmentHandler)
		r.Get("/{id}/customers", getSegmentCustomersHandler)
		r.Post("/{id}/refresh", refreshSegmentHandler)
	})

	return httptest.NewServer(r)
}

func setupSalesServer() *httptest.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", healthHandler)

	// Lead routes
	r.Route("/api/v1/leads", func(r chi.Router) {
		r.Get("/", listLeadsHandler)
		r.Post("/", createLeadHandler)
		r.Get("/{id}", getLeadHandler)
		r.Put("/{id}", updateLeadHandler)
		r.Delete("/{id}", deleteLeadHandler)
		r.Post("/{id}/qualify", qualifyLeadHandler)
		r.Post("/{id}/disqualify", disqualifyLeadHandler)
		r.Post("/{id}/convert", convertLeadHandler)
		r.Post("/{id}/assign", assignLeadHandler)
		r.Get("/{id}/activities", getLeadActivitiesHandler)
	})

	// Opportunity routes
	r.Route("/api/v1/opportunities", func(r chi.Router) {
		r.Get("/", listOpportunitiesHandler)
		r.Post("/", createOpportunityHandler)
		r.Get("/{id}", getOpportunityHandler)
		r.Put("/{id}", updateOpportunityHandler)
		r.Delete("/{id}", deleteOpportunityHandler)
		r.Post("/{id}/move-stage", moveStageHandler)
		r.Post("/{id}/win", winOpportunityHandler)
		r.Post("/{id}/lose", loseOpportunityHandler)
		r.Post("/{id}/reopen", reopenOpportunityHandler)
		r.Get("/{id}/history", getOpportunityHistoryHandler)
	})

	// Deal routes
	r.Route("/api/v1/deals", func(r chi.Router) {
		r.Get("/", listDealsHandler)
		r.Post("/", createDealHandler)
		r.Get("/{id}", getDealHandler)
		r.Put("/{id}", updateDealHandler)
		r.Delete("/{id}", deleteDealHandler)
		r.Post("/{id}/activate", activateDealHandler)
		r.Post("/{id}/complete", completeDealHandler)
		r.Post("/{id}/cancel", cancelDealHandler)
		r.Get("/{id}/line-items", listLineItemsHandler)
		r.Post("/{id}/line-items", addLineItemHandler)
		r.Put("/{id}/line-items/{itemId}", updateLineItemHandler)
		r.Delete("/{id}/line-items/{itemId}", removeLineItemHandler)
	})

	// Pipeline routes
	r.Route("/api/v1/pipelines", func(r chi.Router) {
		r.Get("/", listPipelinesHandler)
		r.Post("/", createPipelineHandler)
		r.Get("/{id}", getPipelineHandler)
		r.Put("/{id}", updatePipelineHandler)
		r.Delete("/{id}", deletePipelineHandler)
		r.Get("/{id}/analytics", getPipelineAnalyticsHandler)
		r.Get("/{id}/stages", listStagesHandler)
		r.Post("/{id}/stages", createStageHandler)
		r.Put("/{id}/stages/{stageId}", updateStageHandler)
		r.Delete("/{id}/stages/{stageId}", deleteStageHandler)
		r.Post("/{id}/stages/reorder", reorderStagesHandler)
	})

	return httptest.NewServer(r)
}

// ============================================================================
// Test Helper Functions
// ============================================================================

// E2ETestSuite provides shared setup for E2E tests.
type E2ETestSuite struct {
	t            *testing.T
	ctx          context.Context
	cancel       context.CancelFunc
	tenantID     uuid.UUID
	tenantSlug   string
	accessToken  string
	refreshToken string
	userID       uuid.UUID
	customerID   uuid.UUID
}

// NewE2ETestSuite creates a new E2E test suite.
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	t.Helper()
	helpers.SkipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	return &E2ETestSuite{
		t:      t,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Cleanup cleans up test resources.
func (s *E2ETestSuite) Cleanup() {
	s.cancel()
	// Clean database
	cleanupTestData(s.ctx)
}

// CreateTenant creates a test tenant.
func (s *E2ETestSuite) CreateTenant(name, slug string) *TenantResponse {
	s.t.Helper()

	body := map[string]interface{}{
		"name": name,
		"slug": slug,
		"plan": "professional",
	}

	resp := s.DoRequest("POST", iamServer.URL+"/api/v1/tenants", body, nil)
	s.AssertStatus(resp, http.StatusCreated)

	var tenantResp TenantResponse
	s.DecodeResponse(resp, &tenantResp)

	s.tenantID = uuid.MustParse(tenantResp.ID)
	s.tenantSlug = tenantResp.Slug

	return &tenantResp
}

// RegisterUser registers a new user.
func (s *E2ETestSuite) RegisterUser(email, password, firstName, lastName string) *UserResponse {
	s.t.Helper()

	body := map[string]interface{}{
		"email":      email,
		"password":   password,
		"first_name": firstName,
		"last_name":  lastName,
		"tenant_id":  s.tenantID.String(),
	}

	resp := s.DoRequest("POST", iamServer.URL+"/api/v1/auth/register", body, nil)
	s.AssertStatus(resp, http.StatusCreated)

	var userResp UserResponse
	s.DecodeResponse(resp, &userResp)

	s.userID = uuid.MustParse(userResp.ID)

	return &userResp
}

// Login performs user login.
func (s *E2ETestSuite) Login(email, password string) *LoginResponse {
	s.t.Helper()

	body := map[string]interface{}{
		"email":     email,
		"password":  password,
		"tenant_id": s.tenantID.String(),
	}

	resp := s.DoRequest("POST", iamServer.URL+"/api/v1/auth/login", body, nil)
	s.AssertStatus(resp, http.StatusOK)

	var loginResp LoginResponse
	s.DecodeResponse(resp, &loginResp)

	s.accessToken = loginResp.AccessToken
	s.refreshToken = loginResp.RefreshToken

	return &loginResp
}

// CreateCustomer creates a test customer.
func (s *E2ETestSuite) CreateCustomer(name, code, email string) *CustomerResponse {
	s.t.Helper()

	body := map[string]interface{}{
		"name":   name,
		"code":   code,
		"type":   "business",
		"status": "active",
		"email": map[string]interface{}{
			"address": email,
		},
	}

	resp := s.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers", body)
	s.AssertStatus(resp, http.StatusCreated)

	var customerResp CustomerResponse
	s.DecodeResponse(resp, &customerResp)

	s.customerID = uuid.MustParse(customerResp.ID)

	return &customerResp
}

// CreateLead creates a test lead.
func (s *E2ETestSuite) CreateLead(companyName, contactName, contactEmail string) *LeadResponse {
	s.t.Helper()

	body := map[string]interface{}{
		"company_name":  companyName,
		"contact_name":  contactName,
		"contact_email": contactEmail,
		"source":        "website",
		"score":         50,
	}

	resp := s.DoRequestAuth("POST", salesServer.URL+"/api/v1/leads", body)
	s.AssertStatus(resp, http.StatusCreated)

	var leadResp LeadResponse
	s.DecodeResponse(resp, &leadResp)

	return &leadResp
}

// DoRequest performs an HTTP request.
func (s *E2ETestSuite) DoRequest(method, url string, body interface{}, headers map[string]string) *http.Response {
	s.t.Helper()

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		helpers.RequireNoError(s.t, err, "failed to marshal request body")
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(s.ctx, method, url, bodyReader)
	helpers.RequireNoError(s.t, err, "failed to create request")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if s.tenantID != uuid.Nil {
		req.Header.Set("X-Tenant-ID", s.tenantID.String())
	}

	resp, err := httpClient.Do(req)
	helpers.RequireNoError(s.t, err, "failed to execute request")

	return resp
}

// DoRequestAuth performs an authenticated HTTP request.
func (s *E2ETestSuite) DoRequestAuth(method, url string, body interface{}) *http.Response {
	s.t.Helper()

	headers := map[string]string{
		"Authorization": "Bearer " + s.accessToken,
	}

	return s.DoRequest(method, url, body, headers)
}

// AssertStatus asserts the response status code.
func (s *E2ETestSuite) AssertStatus(resp *http.Response, expected int) {
	s.t.Helper()

	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		s.t.Errorf("Expected status %d, got %d. Body: %s", expected, resp.StatusCode, string(body))
	}
}

// DecodeResponse decodes the response body.
func (s *E2ETestSuite) DecodeResponse(resp *http.Response, v interface{}) {
	s.t.Helper()

	body, err := io.ReadAll(resp.Body)
	helpers.RequireNoError(s.t, err, "failed to read response body")
	resp.Body.Close()

	err = json.Unmarshal(body, v)
	helpers.RequireNoError(s.t, err, "failed to decode response: "+string(body))
}

// ============================================================================
// Response Types
// ============================================================================

type TenantResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Status    string                 `json:"status"`
	Plan      string                 `json:"plan"`
	Settings  map[string]interface{} `json:"settings"`
	CreatedAt string                 `json:"created_at"`
}

type UserResponse struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Status          string `json:"status"`
	EmailVerifiedAt string `json:"email_verified_at,omitempty"`
	CreatedAt       string `json:"created_at"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"user"`
}

type CustomerResponse struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	Code      string                 `json:"code"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Email     map[string]interface{} `json:"email"`
	CreatedAt string                 `json:"created_at"`
}

type LeadResponse struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	CompanyName  string `json:"company_name"`
	ContactName  string `json:"contact_name"`
	ContactEmail string `json:"contact_email"`
	Source       string `json:"source"`
	Status       string `json:"status"`
	Score        int    `json:"score"`
	CreatedAt    string `json:"created_at"`
}

type OpportunityResponse struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	CustomerID    string `json:"customer_id"`
	LeadID        string `json:"lead_id,omitempty"`
	PipelineID    string `json:"pipeline_id"`
	StageID       string `json:"stage_id"`
	Name          string `json:"name"`
	ValueAmount   int64  `json:"value_amount"`
	ValueCurrency string `json:"value_currency"`
	Probability   int    `json:"probability"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type DealResponse struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	OpportunityID string `json:"opportunity_id"`
	CustomerID    string `json:"customer_id"`
	Name          string `json:"name"`
	ValueAmount   int64  `json:"value_amount"`
	ValueCurrency string `json:"value_currency"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type PipelineResponse struct {
	ID        string          `json:"id"`
	TenantID  string          `json:"tenant_id"`
	Name      string          `json:"name"`
	IsDefault bool            `json:"is_default"`
	Status    string          `json:"status"`
	Stages    []StageResponse `json:"stages"`
	CreatedAt string          `json:"created_at"`
}

type StageResponse struct {
	ID          string `json:"id"`
	PipelineID  string `json:"pipeline_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Order       int    `json:"stage_order"`
	Probability int    `json:"probability"`
}

type ContactResponse struct {
	ID         string                 `json:"id"`
	CustomerID string                 `json:"customer_id"`
	FirstName  string                 `json:"first_name"`
	LastName   string                 `json:"last_name"`
	Email      map[string]interface{} `json:"email"`
	IsPrimary  bool                   `json:"is_primary"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ============================================================================
// Cleanup Functions
// ============================================================================

func cleanupTestData(ctx context.Context) {
	if postgresContainer != nil {
		postgresContainer.TruncateTables(ctx,
			"deal_line_items", "deals", "opportunity_stage_history", "opportunities",
			"leads", "pipeline_stages", "pipelines",
			"sessions", "password_reset_tokens", "email_verification_tokens",
			"refresh_tokens", "user_roles", "users", "roles", "audit_logs", "tenants",
		)
	}

	if mongoContainer != nil {
		mongoContainer.CleanAllCollections(ctx)
	}

	if redisContainer != nil {
		redisContainer.FlushDB(ctx)
	}
}
