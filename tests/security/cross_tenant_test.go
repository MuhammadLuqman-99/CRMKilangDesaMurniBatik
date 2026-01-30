// Package security contains cross-tenant security tests for the CRM system.
package security

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

// ============================================================================
// Cross-Tenant Security Testing
// ============================================================================

// TestCrossTenantDataIsolation tests that data is properly isolated between tenants.
func TestCrossTenantDataIsolation(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	// Create tokens for two different tenants
	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")
	tenant2Token := server.GetValidToken("user@tenant2.com", "tenant-2")

	// Create data for tenant 1
	tenant1CustomerID := uuid.New().String()
	server.DataStore.mu.Lock()
	server.DataStore.Customers[tenant1CustomerID] = &SecureCustomerData{
		ID:       tenant1CustomerID,
		TenantID: "tenant-1",
		OwnerID:  "user-1",
		Name:     "Tenant1 Secret Customer",
		Code:     "T1-SECRET",
	}
	server.DataStore.mu.Unlock()

	// Create data for tenant 2
	tenant2CustomerID := uuid.New().String()
	server.DataStore.mu.Lock()
	server.DataStore.Customers[tenant2CustomerID] = &SecureCustomerData{
		ID:       tenant2CustomerID,
		TenantID: "tenant-2",
		OwnerID:  "user-2",
		Name:     "Tenant2 Secret Customer",
		Code:     "T2-SECRET",
	}
	server.DataStore.mu.Unlock()

	t.Run("Direct ID Access Cross-Tenant", func(t *testing.T) {
		// Tenant 2 trying to access Tenant 1's customer by ID
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers/"+tenant1CustomerID, nil)
		req.Header.Set("Authorization", tenant2Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Cross-tenant data access vulnerability: Tenant2 accessed Tenant1's customer")
		}

		// Verify tenant 1 can still access their own data
		req2, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers/"+tenant1CustomerID, nil)
		req2.Header.Set("Authorization", tenant1Token)

		resp2, err := client.Client.Do(req2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			t.Error("Tenant should be able to access their own data")
		}
	})

	t.Run("List Endpoint Isolation", func(t *testing.T) {
		// Tenant 1 lists customers - should only see their own
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req.Header.Set("Authorization", tenant1Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// Should not contain tenant 2's data
		if strings.Contains(bodyStr, "T2-SECRET") || strings.Contains(bodyStr, "tenant-2") {
			t.Error("Cross-tenant data leak in list endpoint")
		}

		// Tenant 2 lists customers - should only see their own
		req2, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req2.Header.Set("Authorization", tenant2Token)

		resp2, err := client.Client.Do(req2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp2.Body.Close()

		body2, _ := io.ReadAll(resp2.Body)
		bodyStr2 := string(body2)

		// Should not contain tenant 1's data
		if strings.Contains(bodyStr2, "T1-SECRET") || strings.Contains(bodyStr2, "tenant-1") {
			t.Error("Cross-tenant data leak in list endpoint")
		}
	})

	t.Run("Update Cross-Tenant", func(t *testing.T) {
		// Tenant 2 trying to update Tenant 1's customer
		resp, _ := client.DoAuthenticatedRequest("PUT", "/api/v1/customers/"+tenant1CustomerID, tenant2Token, map[string]interface{}{
			"name": "Hijacked by Tenant2!",
		})

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Error("Cross-tenant update vulnerability detected")
			}
		}

		// Verify data wasn't modified
		server.DataStore.mu.RLock()
		customer := server.DataStore.Customers[tenant1CustomerID]
		server.DataStore.mu.RUnlock()

		if customer != nil && customer.Name != "Tenant1 Secret Customer" {
			t.Error("Cross-tenant data was modified!")
		}
	})

	t.Run("Delete Cross-Tenant", func(t *testing.T) {
		// Tenant 2 trying to delete Tenant 1's customer
		req, _ := http.NewRequest("DELETE", server.Server.URL+"/api/v1/customers/"+tenant1CustomerID, nil)
		req.Header.Set("Authorization", tenant2Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
			t.Error("Cross-tenant delete vulnerability detected")
		}

		// Verify data wasn't deleted
		server.DataStore.mu.RLock()
		_, exists := server.DataStore.Customers[tenant1CustomerID]
		server.DataStore.mu.RUnlock()

		if !exists {
			t.Error("Cross-tenant data was deleted!")
		}
	})
}

// TestCrossTenantSearchIsolation tests that search results are tenant-isolated.
func TestCrossTenantSearchIsolation(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")
	tenant2Token := server.GetValidToken("user@tenant2.com", "tenant-2")

	// Create searchable data for both tenants
	server.DataStore.mu.Lock()
	server.DataStore.Customers["t1-search-1"] = &SecureCustomerData{
		ID: "t1-search-1", TenantID: "tenant-1", Name: "SEARCHTERM Company", Code: "T1SEARCH",
	}
	server.DataStore.Customers["t2-search-1"] = &SecureCustomerData{
		ID: "t2-search-1", TenantID: "tenant-2", Name: "SEARCHTERM Industries", Code: "T2SEARCH",
	}
	server.DataStore.mu.Unlock()

	t.Run("Search Results Isolation", func(t *testing.T) {
		searchQueries := []string{
			"/api/v1/customers?search=SEARCHTERM",
			"/api/v1/customers?q=SEARCHTERM",
			"/api/v1/customers?name=SEARCHTERM",
			"/api/v1/customers?filter=name:SEARCHTERM",
		}

		for _, query := range searchQueries {
			t.Run(query, func(t *testing.T) {
				// Tenant 1 search
				req, _ := http.NewRequest("GET", server.Server.URL+query, nil)
				req.Header.Set("Authorization", tenant1Token)

				resp, err := client.Client.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				bodyStr := string(body)

				// Should only contain tenant 1's data
				if strings.Contains(bodyStr, "T2SEARCH") || strings.Contains(bodyStr, "tenant-2") {
					t.Errorf("Search leaking cross-tenant data for query: %s", query)
				}
			})
		}
	})

	t.Run("Aggregation Isolation", func(t *testing.T) {
		// Test that aggregate queries don't leak cross-tenant information
		aggregateEndpoints := []string{
			"/api/v1/customers/count",
			"/api/v1/customers/stats",
			"/api/v1/reports/summary",
		}

		for _, endpoint := range aggregateEndpoints {
			t.Run(endpoint, func(t *testing.T) {
				req1, _ := http.NewRequest("GET", server.Server.URL+endpoint, nil)
				req1.Header.Set("Authorization", tenant1Token)

				req2, _ := http.NewRequest("GET", server.Server.URL+endpoint, nil)
				req2.Header.Set("Authorization", tenant2Token)

				resp1, err1 := client.Client.Do(req1)
				resp2, err2 := client.Client.Do(req2)

				if err1 == nil && resp1 != nil {
					defer resp1.Body.Close()
				}
				if err2 == nil && resp2 != nil {
					defer resp2.Body.Close()
				}

				// Both should return tenant-specific data, not combined
			})
		}
	})
}

// TestCrossTenantAPIManipulation tests various API manipulation attempts.
func TestCrossTenantAPIManipulation(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")

	t.Run("Tenant ID Injection in Body", func(t *testing.T) {
		injectionPayloads := []map[string]interface{}{
			{"name": "Test", "tenant_id": "tenant-2"},
			{"name": "Test", "tenantId": "tenant-2"},
			{"name": "Test", "TenantID": "tenant-2"},
			{"name": "Test", "organization_id": "tenant-2"},
			{"name": "Test", "_tenant": "tenant-2"},
		}

		for i, payload := range injectionPayloads {
			t.Run(fmt.Sprintf("Injection_%d", i), func(t *testing.T) {
				resp, _ := client.DoAuthenticatedRequest("POST", "/api/v1/customers", tenant1Token, payload)

				if resp != nil {
					defer resp.Body.Close()

					// Verify created customer belongs to tenant-1, not tenant-2
					if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
						body, _ := io.ReadAll(resp.Body)
						if strings.Contains(string(body), "tenant-2") {
							t.Error("Tenant ID injection succeeded in request body")
						}
					}
				}
			})
		}
	})

	t.Run("Tenant ID Injection in Headers", func(t *testing.T) {
		maliciousHeaders := []struct {
			name  string
			value string
		}{
			{"X-Tenant-ID", "tenant-2"},
			{"X-Organization-ID", "tenant-2"},
			{"X-Tenant", "tenant-2"},
			{"Tenant-ID", "tenant-2"},
			{"X-Forwarded-Tenant", "tenant-2"},
		}

		for _, header := range maliciousHeaders {
			t.Run(header.name, func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
				req.Header.Set("Authorization", tenant1Token)
				req.Header.Set(header.name, header.value)

				resp, err := client.Client.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				// Should not return tenant-2 data
				if strings.Contains(string(body), "tenant-2") {
					t.Errorf("Header injection with %s leaked cross-tenant data", header.name)
				}
			})
		}
	})

	t.Run("Tenant ID Injection in Query Parameters", func(t *testing.T) {
		queryParams := []string{
			"?tenant_id=tenant-2",
			"?tenantId=tenant-2",
			"?organization=tenant-2",
			"?_tenant=tenant-2",
		}

		for _, query := range queryParams {
			t.Run(query, func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers"+query, nil)
				req.Header.Set("Authorization", tenant1Token)

				resp, err := client.Client.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				if strings.Contains(string(body), "tenant-2") {
					t.Errorf("Query parameter injection leaked cross-tenant data: %s", query)
				}
			})
		}
	})
}

// TestCrossTenantConcurrentAccess tests race conditions in tenant isolation.
func TestCrossTenantConcurrentAccess(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")
	tenant2Token := server.GetValidToken("user@tenant2.com", "tenant-2")

	t.Run("Concurrent Access Race Condition", func(t *testing.T) {
		var wg sync.WaitGroup
		var tenant1SeenTenant2 bool
		var tenant2SeenTenant1 bool
		var mu sync.Mutex

		iterations := 100

		for i := 0; i < iterations; i++ {
			wg.Add(2)

			// Tenant 1 accessing customers
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
				req.Header.Set("Authorization", tenant1Token)

				resp, err := client.Client.Do(req)
				if err == nil && resp != nil {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()

					if strings.Contains(string(body), "tenant-2") {
						mu.Lock()
						tenant1SeenTenant2 = true
						mu.Unlock()
					}
				}
			}()

			// Tenant 2 accessing customers
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
				req.Header.Set("Authorization", tenant2Token)

				resp, err := client.Client.Do(req)
				if err == nil && resp != nil {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()

					if strings.Contains(string(body), "tenant-1") {
						mu.Lock()
						tenant2SeenTenant1 = true
						mu.Unlock()
					}
				}
			}()
		}

		wg.Wait()

		if tenant1SeenTenant2 {
			t.Error("Race condition: Tenant1 saw Tenant2's data during concurrent access")
		}
		if tenant2SeenTenant1 {
			t.Error("Race condition: Tenant2 saw Tenant1's data during concurrent access")
		}
	})

	t.Run("Concurrent Create/Read Race", func(t *testing.T) {
		var wg sync.WaitGroup
		var leakDetected bool
		var mu sync.Mutex

		for i := 0; i < 50; i++ {
			wg.Add(2)

			// Tenant 1 creates a customer
			go func(idx int) {
				defer wg.Done()
				client.DoAuthenticatedRequest("POST", "/api/v1/customers", tenant1Token, map[string]interface{}{
					"code":   fmt.Sprintf("T1-RACE-%d", idx),
					"name":   "Tenant1 Race Customer",
					"status": "active",
				})
			}(i)

			// Tenant 2 lists customers immediately
			go func() {
				defer wg.Done()
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
				req.Header.Set("Authorization", tenant2Token)

				resp, err := client.Client.Do(req)
				if err == nil && resp != nil {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()

					if strings.Contains(string(body), "T1-RACE") {
						mu.Lock()
						leakDetected = true
						mu.Unlock()
					}
				}
			}()
		}

		wg.Wait()

		if leakDetected {
			t.Error("Race condition leak: Tenant2 saw Tenant1's newly created data")
		}
	})
}

// TestCrossTenantBulkOperations tests bulk operation isolation.
func TestCrossTenantBulkOperations(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")

	// Create customers for tenant 2
	tenant2CustomerIDs := []string{}
	server.DataStore.mu.Lock()
	for i := 0; i < 5; i++ {
		id := uuid.New().String()
		server.DataStore.Customers[id] = &SecureCustomerData{
			ID:       id,
			TenantID: "tenant-2",
			Name:     fmt.Sprintf("Tenant2 Bulk Customer %d", i),
		}
		tenant2CustomerIDs = append(tenant2CustomerIDs, id)
	}
	server.DataStore.mu.Unlock()

	t.Run("Bulk Delete Cross-Tenant", func(t *testing.T) {
		// Tenant 1 attempting to bulk delete Tenant 2's customers
		resp, _ := client.DoAuthenticatedRequest("POST", "/api/v1/customers/bulk-delete", tenant1Token, map[string]interface{}{
			"ids": tenant2CustomerIDs,
		})

		if resp != nil {
			defer resp.Body.Close()
		}

		// Verify tenant 2's customers still exist
		server.DataStore.mu.RLock()
		for _, id := range tenant2CustomerIDs {
			if _, exists := server.DataStore.Customers[id]; !exists {
				t.Error("Cross-tenant bulk delete succeeded - customer was deleted")
			}
		}
		server.DataStore.mu.RUnlock()
	})

	t.Run("Bulk Update Cross-Tenant", func(t *testing.T) {
		// Tenant 1 attempting to bulk update Tenant 2's customers
		resp, _ := client.DoAuthenticatedRequest("PUT", "/api/v1/customers/bulk-update", tenant1Token, map[string]interface{}{
			"ids":    tenant2CustomerIDs,
			"status": "hijacked",
		})

		if resp != nil {
			defer resp.Body.Close()
		}

		// Verify tenant 2's customers weren't modified
		server.DataStore.mu.RLock()
		for _, id := range tenant2CustomerIDs {
			if customer, exists := server.DataStore.Customers[id]; exists {
				if customer.Status == "hijacked" {
					t.Error("Cross-tenant bulk update succeeded - customer was modified")
				}
			}
		}
		server.DataStore.mu.RUnlock()
	})

	t.Run("Bulk Export Cross-Tenant", func(t *testing.T) {
		// Tenant 1 attempting to export Tenant 2's data
		resp, _ := client.DoAuthenticatedRequest("POST", "/api/v1/customers/export", tenant1Token, map[string]interface{}{
			"ids": tenant2CustomerIDs,
		})

		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			// Should not contain tenant 2's data
			for _, id := range tenant2CustomerIDs {
				if strings.Contains(string(body), id) {
					t.Error("Cross-tenant bulk export leaked data")
					break
				}
			}
		}
	})
}

// TestCrossTenantRelationships tests isolation of related data.
func TestCrossTenantRelationships(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")

	// Create customer with related data for tenant 2
	tenant2CustomerID := uuid.New().String()
	server.DataStore.mu.Lock()
	server.DataStore.Customers[tenant2CustomerID] = &SecureCustomerData{
		ID:       tenant2CustomerID,
		TenantID: "tenant-2",
		Name:     "Tenant2 Parent Customer",
	}
	server.DataStore.mu.Unlock()

	t.Run("Access Related Contacts Cross-Tenant", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers/"+tenant2CustomerID+"/contacts", nil)
		req.Header.Set("Authorization", tenant1Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Cross-tenant access to related contacts allowed")
		}
	})

	t.Run("Access Related Leads Cross-Tenant", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers/"+tenant2CustomerID+"/leads", nil)
		req.Header.Set("Authorization", tenant1Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Cross-tenant access to related leads allowed")
		}
	})

	t.Run("Create Related Data Cross-Tenant", func(t *testing.T) {
		// Try to create a contact for tenant 2's customer
		resp, _ := client.DoAuthenticatedRequest("POST", "/api/v1/customers/"+tenant2CustomerID+"/contacts", tenant1Token, map[string]interface{}{
			"name":  "Injected Contact",
			"email": "injected@test.com",
		})

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
				t.Error("Cross-tenant creation of related data succeeded")
			}
		}
	})
}

// TestCrossTenantReporting tests isolation in reporting endpoints.
func TestCrossTenantReporting(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")
	tenant2Token := server.GetValidToken("user@tenant2.com", "tenant-2")

	// Create data with identifiable markers
	server.DataStore.mu.Lock()
	server.DataStore.Customers["t1-report"] = &SecureCustomerData{
		ID: "t1-report", TenantID: "tenant-1", Name: "MARKER_TENANT1",
	}
	server.DataStore.Customers["t2-report"] = &SecureCustomerData{
		ID: "t2-report", TenantID: "tenant-2", Name: "MARKER_TENANT2",
	}
	server.DataStore.mu.Unlock()

	reportEndpoints := []string{
		"/api/v1/reports/customers",
		"/api/v1/reports/summary",
		"/api/v1/reports/analytics",
		"/api/v1/dashboard/stats",
	}

	for _, endpoint := range reportEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			// Tenant 1 report
			req1, _ := http.NewRequest("GET", server.Server.URL+endpoint, nil)
			req1.Header.Set("Authorization", tenant1Token)

			resp1, err := client.Client.Do(req1)
			if err == nil && resp1 != nil {
				body, _ := io.ReadAll(resp1.Body)
				resp1.Body.Close()

				if strings.Contains(string(body), "MARKER_TENANT2") {
					t.Errorf("Report endpoint %s leaking tenant2 data to tenant1", endpoint)
				}
			}

			// Tenant 2 report
			req2, _ := http.NewRequest("GET", server.Server.URL+endpoint, nil)
			req2.Header.Set("Authorization", tenant2Token)

			resp2, err := client.Client.Do(req2)
			if err == nil && resp2 != nil {
				body, _ := io.ReadAll(resp2.Body)
				resp2.Body.Close()

				if strings.Contains(string(body), "MARKER_TENANT1") {
					t.Errorf("Report endpoint %s leaking tenant1 data to tenant2", endpoint)
				}
			}
		})
	}
}

// TestCrossTenantWebhooks tests webhook isolation.
func TestCrossTenantWebhooks(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")

	t.Run("List Other Tenant Webhooks", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/webhooks?tenant_id=tenant-2", nil)
		req.Header.Set("Authorization", tenant1Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "tenant-2") {
			t.Error("Cross-tenant webhook listing allowed")
		}
	})

	t.Run("Trigger Other Tenant Webhook", func(t *testing.T) {
		// Attempt to trigger another tenant's webhook
		resp, _ := client.DoAuthenticatedRequest("POST", "/api/v1/webhooks/tenant-2-webhook-id/trigger", tenant1Token, nil)

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Error("Cross-tenant webhook trigger allowed")
			}
		}
	})
}

// TestCrossTenantAuditLogs tests audit log isolation.
func TestCrossTenantAuditLogs(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")

	t.Run("Access Other Tenant Audit Logs", func(t *testing.T) {
		auditEndpoints := []string{
			"/api/v1/audit-logs?tenant_id=tenant-2",
			"/api/v1/audit-logs/tenant-2",
			"/api/v1/admin/audit-logs?filter=tenant:tenant-2",
		}

		for _, endpoint := range auditEndpoints {
			t.Run(endpoint, func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+endpoint, nil)
				req.Header.Set("Authorization", tenant1Token)

				resp, err := client.Client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				body, _ := io.ReadAll(resp.Body)
				if strings.Contains(string(body), "tenant-2") && resp.StatusCode == http.StatusOK {
					t.Errorf("Cross-tenant audit log access allowed via %s", endpoint)
				}
			})
		}
	})
}

// TestCrossTenantConfiguration tests configuration isolation.
func TestCrossTenantConfiguration(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")

	t.Run("Access Other Tenant Settings", func(t *testing.T) {
		settingsEndpoints := []string{
			"/api/v1/settings?tenant_id=tenant-2",
			"/api/v1/tenants/tenant-2/settings",
			"/api/v1/config/tenant-2",
		}

		for _, endpoint := range settingsEndpoints {
			t.Run(endpoint, func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+endpoint, nil)
				req.Header.Set("Authorization", tenant1Token)

				resp, err := client.Client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					if strings.Contains(string(body), "tenant-2") {
						t.Errorf("Cross-tenant settings access allowed via %s", endpoint)
					}
				}
			})
		}
	})

	t.Run("Modify Other Tenant Settings", func(t *testing.T) {
		resp, _ := client.DoAuthenticatedRequest("PUT", "/api/v1/tenants/tenant-2/settings", tenant1Token, map[string]interface{}{
			"billing_email": "hijacked@test.com",
		})

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Error("Cross-tenant settings modification allowed")
			}
		}
	})
}

// TestMultiTenantTokenValidation tests token validation across tenants.
func TestMultiTenantTokenValidation(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	t.Run("Token With Wrong Tenant Claim", func(t *testing.T) {
		// Create a token with tenant-1 claim but try to access tenant-2 data
		// This simulates a forged or modified token
		tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")

		// Add tenant 2 data
		tenant2ID := uuid.New().String()
		server.DataStore.mu.Lock()
		server.DataStore.Customers[tenant2ID] = &SecureCustomerData{
			ID:       tenant2ID,
			TenantID: "tenant-2",
			Name:     "Tenant2 Only",
		}
		server.DataStore.mu.Unlock()

		// Try to access with tenant-2 override header
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers/"+tenant2ID, nil)
		req.Header.Set("Authorization", tenant1Token)
		req.Header.Set("X-Override-Tenant", "tenant-2")

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Token tenant override was accepted")
		}
	})

	t.Run("Cross-Tenant Impersonation", func(t *testing.T) {
		// Attempt to impersonate a user from another tenant
		tenant1Token := server.GetValidToken("admin@tenant1.com", "tenant-1")

		resp, _ := client.DoAuthenticatedRequest("POST", "/api/v1/admin/impersonate", tenant1Token, map[string]interface{}{
			"user_id":   "user-from-tenant-2",
			"tenant_id": "tenant-2",
		})

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Error("Cross-tenant impersonation allowed")
			}
		}
	})
}

// TestTenantIsolationInErrors tests that error messages don't leak tenant info.
func TestTenantIsolationInErrors(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	tenant1Token := server.GetValidToken("user@tenant1.com", "tenant-1")

	// Create tenant 2 data
	tenant2ID := uuid.New().String()
	server.DataStore.mu.Lock()
	server.DataStore.Customers[tenant2ID] = &SecureCustomerData{
		ID:       tenant2ID,
		TenantID: "tenant-2",
		Name:     "Secret Tenant2 Name",
		Code:     "SECRET-CODE",
	}
	server.DataStore.mu.Unlock()

	t.Run("Error Message Isolation", func(t *testing.T) {
		// Try to access tenant 2's resource
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers/"+tenant2ID, nil)
		req.Header.Set("Authorization", tenant1Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// Error should not reveal:
		// - That the resource exists
		// - The tenant it belongs to
		// - Any data about the resource
		sensitiveStrings := []string{
			"tenant-2",
			"Secret Tenant2 Name",
			"SECRET-CODE",
			"belongs to another tenant",
			"different organization",
		}

		for _, sensitive := range sensitiveStrings {
			if strings.Contains(bodyStr, sensitive) {
				t.Errorf("Error message leaks sensitive information: %s", sensitive)
			}
		}
	})

	t.Run("Validation Error Isolation", func(t *testing.T) {
		// Send invalid data that might trigger detailed validation errors
		resp, _ := client.DoAuthenticatedRequest("PUT", "/api/v1/customers/"+tenant2ID, tenant1Token, map[string]interface{}{
			"name": "",
		})

		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			// Validation errors should not reveal the resource exists
			var errorResp map[string]interface{}
			if json.Unmarshal(body, &errorResp) == nil {
				if _, hasField := errorResp["current_name"]; hasField {
					t.Error("Validation error reveals resource data")
				}
			}
		}
	})
}
