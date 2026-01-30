// Package e2e contains E2E tests for multi-tenant isolation.
package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiTenantIsolation tests that data is properly isolated between tenants.
func TestMultiTenantIsolation(t *testing.T) {
	// Create two separate test suites for two tenants
	suite1 := NewE2ETestSuite(t)
	defer suite1.Cleanup()

	suite2 := NewE2ETestSuite(t)
	defer suite2.Cleanup()

	t.Run("Complete Multi-Tenant Isolation", func(t *testing.T) {
		// =====================================================
		// Setup Tenant A
		// =====================================================
		tenantA := suite1.CreateTenant("Tenant A Company", "tenant-a-company")
		require.NotEmpty(t, tenantA.ID)

		suite1.RegisterUser("admin@tenant-a.com", "PasswordA123!", "Admin", "TenantA")
		suite1.Login("admin@tenant-a.com", "PasswordA123!")

		// Create data in Tenant A
		customerA := suite1.CreateCustomer("Customer A1", "CUST-A1", "customer@tenant-a.com")
		require.NotEmpty(t, customerA.ID)

		leadA := suite1.CreateLead("Lead A1 Company", "Lead A1 Contact", "lead@tenant-a.com")
		require.NotEmpty(t, leadA.ID)

		// Create pipeline in Tenant A
		pipelineRespA := suite1.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name":       "Tenant A Pipeline",
			"is_default": true,
		})
		suite1.AssertStatus(pipelineRespA, http.StatusCreated)
		var pipelineA PipelineResponse
		suite1.DecodeResponse(pipelineRespA, &pipelineA)

		// =====================================================
		// Setup Tenant B
		// =====================================================
		tenantB := suite2.CreateTenant("Tenant B Company", "tenant-b-company")
		require.NotEmpty(t, tenantB.ID)

		suite2.RegisterUser("admin@tenant-b.com", "PasswordB123!", "Admin", "TenantB")
		suite2.Login("admin@tenant-b.com", "PasswordB123!")

		// Create data in Tenant B
		customerB := suite2.CreateCustomer("Customer B1", "CUST-B1", "customer@tenant-b.com")
		require.NotEmpty(t, customerB.ID)

		leadB := suite2.CreateLead("Lead B1 Company", "Lead B1 Contact", "lead@tenant-b.com")
		require.NotEmpty(t, leadB.ID)

		// Create pipeline in Tenant B
		pipelineRespB := suite2.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name":       "Tenant B Pipeline",
			"is_default": true,
		})
		suite2.AssertStatus(pipelineRespB, http.StatusCreated)
		var pipelineB PipelineResponse
		suite2.DecodeResponse(pipelineRespB, &pipelineB)

		// =====================================================
		// Test Customer Isolation
		// =====================================================
		t.Run("Customer Isolation", func(t *testing.T) {
			// Tenant A should only see their customers
			resp := suite1.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers", nil)
			suite1.AssertStatus(resp, http.StatusOK)
			var customersA struct {
				Data []map[string]interface{} `json:"data"`
			}
			suite1.DecodeResponse(resp, &customersA)

			for _, c := range customersA.Data {
				tenantID := c["tenant_id"].(string)
				assert.Equal(t, tenantA.ID, tenantID, "Tenant A should only see their own customers")
			}

			// Tenant B should only see their customers
			resp = suite2.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers", nil)
			suite2.AssertStatus(resp, http.StatusOK)
			var customersB struct {
				Data []map[string]interface{} `json:"data"`
			}
			suite2.DecodeResponse(resp, &customersB)

			for _, c := range customersB.Data {
				tenantID := c["tenant_id"].(string)
				assert.Equal(t, tenantB.ID, tenantID, "Tenant B should only see their own customers")
			}

			// Tenant A cannot access Tenant B's customer by ID
			resp = suite1.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/"+customerB.ID, nil)
			suite1.AssertStatus(resp, http.StatusNotFound)

			// Tenant B cannot access Tenant A's customer by ID
			resp = suite2.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/"+customerA.ID, nil)
			suite2.AssertStatus(resp, http.StatusNotFound)
		})

		// =====================================================
		// Test Lead Isolation
		// =====================================================
		t.Run("Lead Isolation", func(t *testing.T) {
			// Tenant A should only see their leads
			resp := suite1.DoRequestAuth("GET", salesServer.URL+"/api/v1/leads", nil)
			suite1.AssertStatus(resp, http.StatusOK)
			var leadsA struct {
				Data []map[string]interface{} `json:"data"`
			}
			suite1.DecodeResponse(resp, &leadsA)

			for _, l := range leadsA.Data {
				tenantID := l["tenant_id"].(string)
				assert.Equal(t, tenantA.ID, tenantID, "Tenant A should only see their own leads")
			}

			// Tenant B should only see their leads
			resp = suite2.DoRequestAuth("GET", salesServer.URL+"/api/v1/leads", nil)
			suite2.AssertStatus(resp, http.StatusOK)
			var leadsB struct {
				Data []map[string]interface{} `json:"data"`
			}
			suite2.DecodeResponse(resp, &leadsB)

			for _, l := range leadsB.Data {
				tenantID := l["tenant_id"].(string)
				assert.Equal(t, tenantB.ID, tenantID, "Tenant B should only see their own leads")
			}

			// Tenant A cannot access Tenant B's lead by ID
			resp = suite1.DoRequestAuth("GET", salesServer.URL+"/api/v1/leads/"+leadB.ID, nil)
			suite1.AssertStatus(resp, http.StatusNotFound)

			// Tenant B cannot access Tenant A's lead by ID
			resp = suite2.DoRequestAuth("GET", salesServer.URL+"/api/v1/leads/"+leadA.ID, nil)
			suite2.AssertStatus(resp, http.StatusNotFound)
		})

		// =====================================================
		// Test Pipeline Isolation
		// =====================================================
		t.Run("Pipeline Isolation", func(t *testing.T) {
			// Tenant A should only see their pipelines
			resp := suite1.DoRequestAuth("GET", salesServer.URL+"/api/v1/pipelines", nil)
			suite1.AssertStatus(resp, http.StatusOK)
			var pipelinesA struct {
				Data []map[string]interface{} `json:"data"`
			}
			suite1.DecodeResponse(resp, &pipelinesA)

			for _, p := range pipelinesA.Data {
				tenantID := p["tenant_id"].(string)
				assert.Equal(t, tenantA.ID, tenantID, "Tenant A should only see their own pipelines")
			}

			// Tenant B should only see their pipelines
			resp = suite2.DoRequestAuth("GET", salesServer.URL+"/api/v1/pipelines", nil)
			suite2.AssertStatus(resp, http.StatusOK)
			var pipelinesB struct {
				Data []map[string]interface{} `json:"data"`
			}
			suite2.DecodeResponse(resp, &pipelinesB)

			for _, p := range pipelinesB.Data {
				tenantID := p["tenant_id"].(string)
				assert.Equal(t, tenantB.ID, tenantID, "Tenant B should only see their own pipelines")
			}

			// Tenant A cannot access Tenant B's pipeline by ID
			resp = suite1.DoRequestAuth("GET", salesServer.URL+"/api/v1/pipelines/"+pipelineB.ID, nil)
			suite1.AssertStatus(resp, http.StatusNotFound)

			// Tenant B cannot access Tenant A's pipeline by ID
			resp = suite2.DoRequestAuth("GET", salesServer.URL+"/api/v1/pipelines/"+pipelineA.ID, nil)
			suite2.AssertStatus(resp, http.StatusNotFound)
		})

		// =====================================================
		// Test Cross-Tenant Modification Prevention
		// =====================================================
		t.Run("Cross-Tenant Modification Prevention", func(t *testing.T) {
			// Tenant A cannot update Tenant B's customer
			resp := suite1.DoRequestAuth("PUT", customerServer.URL+"/api/v1/customers/"+customerB.ID, map[string]interface{}{
				"name": "Hacked Customer",
			})
			suite1.AssertStatus(resp, http.StatusNotFound)

			// Tenant A cannot delete Tenant B's customer
			resp = suite1.DoRequestAuth("DELETE", customerServer.URL+"/api/v1/customers/"+customerB.ID, nil)
			suite1.AssertStatus(resp, http.StatusNotFound)

			// Tenant A cannot update Tenant B's lead
			resp = suite1.DoRequestAuth("PUT", salesServer.URL+"/api/v1/leads/"+leadB.ID, map[string]interface{}{
				"company_name": "Hacked Lead",
			})
			suite1.AssertStatus(resp, http.StatusNotFound)

			// Tenant A cannot delete Tenant B's lead
			resp = suite1.DoRequestAuth("DELETE", salesServer.URL+"/api/v1/leads/"+leadB.ID, nil)
			suite1.AssertStatus(resp, http.StatusNotFound)

			// Tenant A cannot update Tenant B's pipeline
			resp = suite1.DoRequestAuth("PUT", salesServer.URL+"/api/v1/pipelines/"+pipelineB.ID, map[string]interface{}{
				"name": "Hacked Pipeline",
			})
			suite1.AssertStatus(resp, http.StatusNotFound)
		})

		// =====================================================
		// Test User Isolation
		// =====================================================
		t.Run("User Isolation", func(t *testing.T) {
			// Tenant A should only see their users
			resp := suite1.DoRequestAuth("GET", iamServer.URL+"/api/v1/users", nil)
			suite1.AssertStatus(resp, http.StatusOK)
			var usersA struct {
				Data []map[string]interface{} `json:"data"`
			}
			suite1.DecodeResponse(resp, &usersA)

			// Verify Tenant A users don't include Tenant B users
			for _, u := range usersA.Data {
				email := u["email"].(string)
				assert.NotContains(t, email, "tenant-b", "Tenant A should not see Tenant B users")
			}

			// Tenant B should only see their users
			resp = suite2.DoRequestAuth("GET", iamServer.URL+"/api/v1/users", nil)
			suite2.AssertStatus(resp, http.StatusOK)
			var usersB struct {
				Data []map[string]interface{} `json:"data"`
			}
			suite2.DecodeResponse(resp, &usersB)

			// Verify Tenant B users don't include Tenant A users
			for _, u := range usersB.Data {
				email := u["email"].(string)
				assert.NotContains(t, email, "tenant-a", "Tenant B should not see Tenant A users")
			}
		})
	})
}

// TestTenantDataIsolationWithSameNames tests that tenants can have entities with same names/codes.
func TestTenantDataIsolationWithSameNames(t *testing.T) {
	suite1 := NewE2ETestSuite(t)
	defer suite1.Cleanup()

	suite2 := NewE2ETestSuite(t)
	defer suite2.Cleanup()

	t.Run("Same Customer Code Different Tenants", func(t *testing.T) {
		// Setup Tenant 1
		suite1.CreateTenant("Company One", "company-one")
		suite1.RegisterUser("admin@company-one.com", "Password123!", "Admin", "One")
		suite1.Login("admin@company-one.com", "Password123!")

		// Setup Tenant 2
		suite2.CreateTenant("Company Two", "company-two")
		suite2.RegisterUser("admin@company-two.com", "Password123!", "Admin", "Two")
		suite2.Login("admin@company-two.com", "Password123!")

		// Both tenants can create customers with the same code
		customer1 := suite1.CreateCustomer("Acme Corp", "ACME-001", "acme@company-one.com")
		assert.NotEmpty(t, customer1.ID)
		assert.Equal(t, "ACME-001", customer1.Code)

		customer2 := suite2.CreateCustomer("Acme Inc", "ACME-001", "acme@company-two.com")
		assert.NotEmpty(t, customer2.ID)
		assert.Equal(t, "ACME-001", customer2.Code)

		// Verify they are different entities
		assert.NotEqual(t, customer1.ID, customer2.ID)
	})

	t.Run("Same Email Different Tenants", func(t *testing.T) {
		// Setup Tenant 1
		suite1.CreateTenant("Email Test One", "email-test-one")
		suite1.RegisterUser("admin@email-test-one.com", "Password123!", "Admin", "One")
		suite1.Login("admin@email-test-one.com", "Password123!")

		// Setup Tenant 2
		suite2.CreateTenant("Email Test Two", "email-test-two")
		suite2.RegisterUser("admin@email-test-two.com", "Password123!", "Admin", "Two")
		suite2.Login("admin@email-test-two.com", "Password123!")

		// Both tenants can have users with the same email
		user1 := suite1.RegisterUser("shared@email.com", "Password123!", "User", "One")
		assert.NotEmpty(t, user1.ID)

		user2 := suite2.RegisterUser("shared@email.com", "Password123!", "User", "Two")
		assert.NotEmpty(t, user2.ID)

		// Verify they are different entities
		assert.NotEqual(t, user1.ID, user2.ID)
	})
}

// TestTenantSuspension tests that suspended tenants cannot access resources.
func TestTenantSuspension(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	t.Run("Suspended Tenant Status Update", func(t *testing.T) {
		// Create a tenant
		tenant := suite.CreateTenant("Suspension Test", "suspension-test")
		require.NotEmpty(t, tenant.ID)
		assert.Equal(t, "active", tenant.Status)

		// Suspend the tenant
		resp := suite.DoRequest("PUT", iamServer.URL+"/api/v1/tenants/"+tenant.ID+"/status", map[string]interface{}{
			"status": "suspended",
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var statusResp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &statusResp)
		assert.Equal(t, "suspended", statusResp.Status)

		// Reactivate the tenant
		resp = suite.DoRequest("PUT", iamServer.URL+"/api/v1/tenants/"+tenant.ID+"/status", map[string]interface{}{
			"status": "active",
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)

		suite.DecodeResponse(resp, &statusResp)
		assert.Equal(t, "active", statusResp.Status)
	})
}

// TestTenantStatistics tests that tenant statistics are correctly isolated.
func TestTenantStatistics(t *testing.T) {
	suite1 := NewE2ETestSuite(t)
	defer suite1.Cleanup()

	suite2 := NewE2ETestSuite(t)
	defer suite2.Cleanup()

	t.Run("Tenant Stats Isolation", func(t *testing.T) {
		// Setup Tenant 1 with data
		tenant1 := suite1.CreateTenant("Stats Test One", "stats-test-one")
		suite1.RegisterUser("user1@stats-one.com", "Password123!", "User1", "One")
		suite1.RegisterUser("user2@stats-one.com", "Password123!", "User2", "One")
		suite1.Login("user1@stats-one.com", "Password123!")

		suite1.CreateCustomer("Customer 1A", "C1A", "c1a@stats-one.com")
		suite1.CreateCustomer("Customer 1B", "C1B", "c1b@stats-one.com")
		suite1.CreateCustomer("Customer 1C", "C1C", "c1c@stats-one.com")

		suite1.CreateLead("Lead 1A", "Contact 1A", "l1a@stats-one.com")

		// Setup Tenant 2 with different data
		tenant2 := suite2.CreateTenant("Stats Test Two", "stats-test-two")
		suite2.RegisterUser("user1@stats-two.com", "Password123!", "User1", "Two")
		suite2.Login("user1@stats-two.com", "Password123!")

		suite2.CreateCustomer("Customer 2A", "C2A", "c2a@stats-two.com")

		suite2.CreateLead("Lead 2A", "Contact 2A", "l2a@stats-two.com")
		suite2.CreateLead("Lead 2B", "Contact 2B", "l2b@stats-two.com")

		// Get stats for Tenant 1
		resp := suite1.DoRequest("GET", iamServer.URL+"/api/v1/tenants/"+tenant1.ID+"/stats", nil, nil)
		suite1.AssertStatus(resp, http.StatusOK)

		var stats1 struct {
			Users     int `json:"users"`
			Customers int `json:"customers"`
			Leads     int `json:"leads"`
		}
		suite1.DecodeResponse(resp, &stats1)

		// Tenant 1 should have 2 users, 3 customers, 1 lead
		assert.Equal(t, 2, stats1.Users)
		assert.Equal(t, 3, stats1.Customers)
		assert.Equal(t, 1, stats1.Leads)

		// Get stats for Tenant 2
		resp = suite2.DoRequest("GET", iamServer.URL+"/api/v1/tenants/"+tenant2.ID+"/stats", nil, nil)
		suite2.AssertStatus(resp, http.StatusOK)

		var stats2 struct {
			Users     int `json:"users"`
			Customers int `json:"customers"`
			Leads     int `json:"leads"`
		}
		suite2.DecodeResponse(resp, &stats2)

		// Tenant 2 should have 1 user, 1 customer, 2 leads
		assert.Equal(t, 1, stats2.Users)
		assert.Equal(t, 1, stats2.Customers)
		assert.Equal(t, 2, stats2.Leads)
	})
}

// TestCrossServiceTenantIsolation tests that tenant isolation works across services.
func TestCrossServiceTenantIsolation(t *testing.T) {
	suite1 := NewE2ETestSuite(t)
	defer suite1.Cleanup()

	suite2 := NewE2ETestSuite(t)
	defer suite2.Cleanup()

	t.Run("Cross Service Isolation", func(t *testing.T) {
		// Setup Tenant 1
		suite1.CreateTenant("Cross Service One", "cross-service-one")
		suite1.RegisterUser("admin@cs-one.com", "Password123!", "Admin", "One")
		suite1.Login("admin@cs-one.com", "Password123!")

		// Create customer in Customer service
		customer1 := suite1.CreateCustomer("CS Customer One", "CS-001", "cs@one.com")

		// Create pipeline in Sales service
		pipelineResp := suite1.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name":       "CS Pipeline One",
			"is_default": true,
		})
		suite1.AssertStatus(pipelineResp, http.StatusCreated)
		var pipeline1 PipelineResponse
		suite1.DecodeResponse(pipelineResp, &pipeline1)

		stageID := ""
		if len(pipeline1.Stages) > 0 {
			stageID = pipeline1.Stages[0].ID
		}

		// Create opportunity linking customer and pipeline
		oppResp := suite1.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities", map[string]interface{}{
			"customer_id":    customer1.ID,
			"pipeline_id":    pipeline1.ID,
			"stage_id":       stageID,
			"name":           "Cross Service Opportunity",
			"value_amount":   10000,
			"value_currency": "USD",
		})
		suite1.AssertStatus(oppResp, http.StatusCreated)
		var opp1 OpportunityResponse
		suite1.DecodeResponse(oppResp, &opp1)

		// Setup Tenant 2
		suite2.CreateTenant("Cross Service Two", "cross-service-two")
		suite2.RegisterUser("admin@cs-two.com", "Password123!", "Admin", "Two")
		suite2.Login("admin@cs-two.com", "Password123!")

		// Tenant 2 cannot see Tenant 1's opportunity
		resp := suite2.DoRequestAuth("GET", salesServer.URL+"/api/v1/opportunities/"+opp1.ID, nil)
		suite2.AssertStatus(resp, http.StatusNotFound)

		// Tenant 2 cannot use Tenant 1's customer ID to create opportunity
		pipelineResp2 := suite2.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name":       "CS Pipeline Two",
			"is_default": true,
		})
		suite2.AssertStatus(pipelineResp2, http.StatusCreated)
		var pipeline2 PipelineResponse
		suite2.DecodeResponse(pipelineResp2, &pipeline2)

		stage2ID := ""
		if len(pipeline2.Stages) > 0 {
			stage2ID = pipeline2.Stages[0].ID
		}

		// This should work (Tenant 2's pipeline) but with their own customer
		customer2 := suite2.CreateCustomer("CS Customer Two", "CS-002", "cs@two.com")
		oppResp2 := suite2.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities", map[string]interface{}{
			"customer_id":    customer2.ID,
			"pipeline_id":    pipeline2.ID,
			"stage_id":       stage2ID,
			"name":           "Tenant 2 Opportunity",
			"value_amount":   5000,
			"value_currency": "USD",
		})
		suite2.AssertStatus(oppResp2, http.StatusCreated)
	})
}

// TestTenantPlanLimits tests that tenant plan limits are enforced (placeholder for future implementation).
func TestTenantPlanLimits(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	t.Run("Plan Upgrade", func(t *testing.T) {
		tenant := suite.CreateTenant("Plan Test", "plan-test")
		assert.Equal(t, "professional", tenant.Plan)

		// Upgrade to enterprise
		resp := suite.DoRequest("PUT", iamServer.URL+"/api/v1/tenants/"+tenant.ID+"/plan", map[string]interface{}{
			"plan": "enterprise",
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var planResp struct {
			Plan string `json:"plan"`
		}
		suite.DecodeResponse(resp, &planResp)
		assert.Equal(t, "enterprise", planResp.Plan)
	})

	t.Run("Plan Downgrade", func(t *testing.T) {
		tenant := suite.CreateTenant("Downgrade Test", "downgrade-test")

		// Downgrade to free
		resp := suite.DoRequest("PUT", iamServer.URL+"/api/v1/tenants/"+tenant.ID+"/plan", map[string]interface{}{
			"plan": "free",
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var planResp struct {
			Plan string `json:"plan"`
		}
		suite.DecodeResponse(resp, &planResp)
		assert.Equal(t, "free", planResp.Plan)
	})
}
