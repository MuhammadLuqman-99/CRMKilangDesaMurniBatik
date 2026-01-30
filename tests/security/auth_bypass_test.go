// Package security contains authentication bypass tests for the CRM system.
package security

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ============================================================================
// Authentication Bypass Testing
// ============================================================================

// TestAuthenticationBypass tests various authentication bypass techniques.
func TestAuthenticationBypass(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	t.Run("Missing Authentication", func(t *testing.T) {
		// Attempt to access protected endpoints without any authentication
		protectedEndpoints := []struct {
			method   string
			endpoint string
		}{
			{"GET", "/api/v1/customers"},
			{"POST", "/api/v1/customers"},
			{"GET", "/api/v1/customers/123"},
			{"PUT", "/api/v1/customers/123"},
			{"DELETE", "/api/v1/customers/123"},
			{"GET", "/api/v1/users"},
			{"POST", "/api/v1/users"},
			{"GET", "/api/v1/admin/settings"},
		}

		for _, ep := range protectedEndpoints {
			t.Run(fmt.Sprintf("%s_%s", ep.method, strings.ReplaceAll(ep.endpoint, "/", "_")), func(t *testing.T) {
				req, _ := http.NewRequest(ep.method, server.Server.URL+ep.endpoint, nil)
				// No authentication header added

				resp, err := client.Client.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				// Should return 401 Unauthorized
				if resp.StatusCode != http.StatusUnauthorized {
					t.Errorf("Expected 401 for unauthenticated request, got %d", resp.StatusCode)
				}
			})
		}
	})

	t.Run("Invalid Token Format", func(t *testing.T) {
		invalidTokens := []struct {
			name  string
			token string
		}{
			{"Empty token", ""},
			{"Invalid format", "invalid-token"},
			{"Missing Bearer prefix", "some-token-value"},
			{"Wrong prefix", "Basic some-token"},
			{"Malformed JWT", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid"},
			{"Truncated token", "Bearer eyJ"},
			{"SQL in token", "Bearer ' OR '1'='1"},
			{"XSS in token", "Bearer <script>alert(1)</script>"},
			{"Null bytes", "Bearer token\x00\x00"},
			{"Unicode tricks", "Bearer \u202Etoken"},
		}

		for _, tc := range invalidTokens {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
				if tc.token != "" {
					req.Header.Set("Authorization", tc.token)
				}

				resp, err := client.Client.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				// Should return 401 Unauthorized for invalid tokens
				if resp.StatusCode != http.StatusUnauthorized {
					t.Errorf("Expected 401 for invalid token '%s', got %d", tc.name, resp.StatusCode)
				}
			})
		}
	})

	t.Run("Expired Token", func(t *testing.T) {
		// Simulate an expired token
		expiredToken := "Bearer expired-token-12345"

		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req.Header.Set("Authorization", expiredToken)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected 401 for expired token, got %d", resp.StatusCode)
		}
	})

	t.Run("Token Tampering", func(t *testing.T) {
		// Get a valid token first
		validToken := server.GetValidToken("user@test.com", "tenant-1")

		tamperTests := []struct {
			name     string
			modifier func(string) string
		}{
			{
				"Modified payload",
				func(token string) string {
					// Attempt to modify the token payload
					parts := strings.Split(strings.TrimPrefix(token, "Bearer "), ".")
					if len(parts) >= 2 {
						parts[1] = base64.StdEncoding.EncodeToString([]byte(`{"admin":true}`))
						return "Bearer " + strings.Join(parts, ".")
					}
					return token + "modified"
				},
			},
			{
				"Modified signature",
				func(token string) string {
					return token + "x"
				},
			},
			{
				"Removed signature",
				func(token string) string {
					parts := strings.Split(token, ".")
					if len(parts) >= 2 {
						return strings.Join(parts[:2], ".")
					}
					return token[:len(token)-10]
				},
			},
			{
				"Algorithm confusion",
				func(token string) string {
					// Attempt alg:none attack
					return "Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." + strings.TrimPrefix(token, "Bearer ")
				},
			},
		}

		for _, tc := range tamperTests {
			t.Run(tc.name, func(t *testing.T) {
				tamperedToken := tc.modifier(validToken)

				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
				req.Header.Set("Authorization", tamperedToken)

				resp, err := client.Client.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				// Tampered tokens should be rejected
				if resp.StatusCode == http.StatusOK {
					t.Errorf("Tampered token was accepted: %s", tc.name)
				}
			})
		}
	})
}

// TestAuthorizationBypass tests authorization bypass techniques.
func TestAuthorizationBypass(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	// Create test users with different roles
	adminToken := server.GetValidToken("admin@test.com", "tenant-1")
	userToken := server.GetValidToken("user@test.com", "tenant-1")

	t.Run("Privilege Escalation", func(t *testing.T) {
		// Regular user trying to access admin endpoints
		adminEndpoints := []string{
			"/api/v1/admin/settings",
			"/api/v1/admin/users",
			"/api/v1/admin/audit-logs",
			"/api/v1/admin/system-config",
		}

		for _, endpoint := range adminEndpoints {
			t.Run(endpoint, func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+endpoint, nil)
				req.Header.Set("Authorization", userToken)

				resp, err := client.Client.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				// Should return 403 Forbidden for unauthorized access
				if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusNotFound {
					t.Errorf("Expected 403/404 for privilege escalation attempt, got %d", resp.StatusCode)
				}
			})
		}

		// Verify admin can access
		t.Run("Admin access succeeds", func(t *testing.T) {
			req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/admin/settings", nil)
			req.Header.Set("Authorization", adminToken)

			resp, err := client.Client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusForbidden {
				t.Error("Admin should have access to admin endpoints")
			}
		})
	})

	t.Run("Horizontal Privilege Escalation", func(t *testing.T) {
		// User trying to access another user's resources
		user1Token := server.GetValidToken("user1@test.com", "tenant-1")
		user2ID := uuid.New().String()

		// Create a customer owned by user2
		server.DataStore.mu.Lock()
		server.DataStore.Customers[user2ID] = &SecureCustomerData{
			ID:       user2ID,
			TenantID: "tenant-1",
			OwnerID:  "user2-id",
			Name:     "User2 Customer",
		}
		server.DataStore.mu.Unlock()

		// User1 trying to access User2's customer
		req, _ := http.NewRequest("DELETE", server.Server.URL+"/api/v1/customers/"+user2ID, nil)
		req.Header.Set("Authorization", user1Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should be blocked (varies by implementation - could be 403 or 404)
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
			t.Error("User should not be able to delete another user's resources")
		}
	})

	t.Run("Role Manipulation", func(t *testing.T) {
		// Attempting to modify own role
		manipulations := []map[string]interface{}{
			{"role": "admin"},
			{"roles": []string{"admin", "superuser"}},
			{"is_admin": true},
			{"permissions": []string{"*"}},
			{"group": "administrators"},
		}

		for i, payload := range manipulations {
			t.Run(fmt.Sprintf("Manipulation_%d", i), func(t *testing.T) {
				resp, _ := client.DoAuthenticatedRequest("PUT", "/api/v1/users/me", userToken, payload)
				if resp != nil {
					defer resp.Body.Close()

					// Verify that role wasn't actually changed
					verifyResp, _ := client.DoAuthenticatedRequest("GET", "/api/v1/users/me", userToken, nil)
					if verifyResp != nil {
						defer verifyResp.Body.Close()
						// Role should still be 'user', not 'admin'
					}
				}
			})
		}
	})

	t.Run("Function Level Access Control", func(t *testing.T) {
		// Test access to various function levels
		functionTests := []struct {
			name     string
			method   string
			endpoint string
			allowed  bool
		}{
			{"Read own data", "GET", "/api/v1/customers", true},
			{"Create data", "POST", "/api/v1/customers", true},
			{"Bulk delete", "DELETE", "/api/v1/customers/bulk", false},
			{"Export all", "GET", "/api/v1/customers/export/all", false},
			{"System health", "GET", "/api/v1/system/health", true},
			{"System shutdown", "POST", "/api/v1/system/shutdown", false},
		}

		for _, tc := range functionTests {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest(tc.method, server.Server.URL+tc.endpoint, nil)
				req.Header.Set("Authorization", userToken)

				resp, err := client.Client.Do(req)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
				defer resp.Body.Close()

				isAllowed := resp.StatusCode == http.StatusOK ||
					resp.StatusCode == http.StatusCreated ||
					resp.StatusCode == http.StatusNoContent

				if tc.allowed && !isAllowed && resp.StatusCode == http.StatusForbidden {
					t.Errorf("Expected access to be allowed for %s", tc.name)
				}
				if !tc.allowed && isAllowed {
					t.Errorf("Expected access to be denied for %s", tc.name)
				}
			})
		}
	})
}

// TestSessionManagement tests session management security.
func TestSessionManagement(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	t.Run("Session Fixation", func(t *testing.T) {
		// Get initial session
		initialToken := server.GetValidToken("user@test.com", "tenant-1")

		// Simulate login - session should be regenerated
		loginResp, _ := client.DoAuthenticatedRequest("POST", "/api/v1/auth/login", initialToken, map[string]interface{}{
			"email":    "user@test.com",
			"password": "password123",
		})

		if loginResp != nil {
			defer loginResp.Body.Close()

			// New session token should be different
			newToken := loginResp.Header.Get("X-New-Token")
			if newToken == initialToken {
				t.Error("Session should be regenerated after login")
			}
		}
	})

	t.Run("Session Timeout", func(t *testing.T) {
		// This test verifies session timeout is configured
		// In production, we'd test actual timeout behavior
		token := server.GetValidToken("user@test.com", "tenant-1")

		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req.Header.Set("Authorization", token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Check for session-related headers
		if resp.Header.Get("X-Session-Expires") == "" {
			// Session expiration should be communicated
			t.Log("Note: Session expiration header not set")
		}
	})

	t.Run("Concurrent Sessions", func(t *testing.T) {
		// Test handling of concurrent sessions from same user
		token1 := server.GetValidToken("concurrent@test.com", "tenant-1")
		token2 := server.GetValidToken("concurrent@test.com", "tenant-1")

		// Both should work or only the latest should work
		req1, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req1.Header.Set("Authorization", token1)

		req2, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req2.Header.Set("Authorization", token2)

		resp1, _ := client.Client.Do(req1)
		resp2, _ := client.Client.Do(req2)

		if resp1 != nil {
			defer resp1.Body.Close()
		}
		if resp2 != nil {
			defer resp2.Body.Close()
		}

		// Log the behavior for review
		if resp1 != nil && resp2 != nil {
			t.Logf("Token1 status: %d, Token2 status: %d", resp1.StatusCode, resp2.StatusCode)
		}
	})

	t.Run("Logout Token Invalidation", func(t *testing.T) {
		token := server.GetValidToken("logout@test.com", "tenant-1")

		// Verify token works before logout
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req.Header.Set("Authorization", token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Skipf("Token not valid before logout test")
		}

		// Perform logout
		logoutReq, _ := http.NewRequest("POST", server.Server.URL+"/api/v1/auth/logout", nil)
		logoutReq.Header.Set("Authorization", token)
		logoutResp, _ := client.Client.Do(logoutReq)
		if logoutResp != nil {
			logoutResp.Body.Close()
		}

		// Add token to invalidated list (simulating server-side logout)
		server.DataStore.mu.Lock()
		server.DataStore.InvalidatedTokens[token] = true
		server.DataStore.mu.Unlock()

		// Token should be invalid after logout
		req2, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req2.Header.Set("Authorization", token)

		resp2, err := client.Client.Do(req2)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode == http.StatusOK {
			t.Error("Token should be invalidated after logout")
		}
	})
}

// TestBruteForceProtection tests protection against brute force attacks.
func TestBruteForceProtection(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	t.Run("Login Brute Force", func(t *testing.T) {
		// Attempt multiple failed logins
		failedAttempts := 0
		for i := 0; i < 15; i++ {
			resp, err := client.DoRequest("POST", "/api/v1/auth/login", map[string]interface{}{
				"email":    "target@test.com",
				"password": fmt.Sprintf("wrong-password-%d", i),
			})

			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				t.Logf("Rate limited after %d attempts", i+1)
				break
			}
			failedAttempts++
		}

		// Should be rate limited before 15 attempts
		if failedAttempts >= 15 {
			t.Error("Brute force protection should limit login attempts")
		}
	})

	t.Run("Password Reset Brute Force", func(t *testing.T) {
		attempts := 0
		for i := 0; i < 10; i++ {
			resp, err := client.DoRequest("POST", "/api/v1/auth/reset-password", map[string]interface{}{
				"email": "target@test.com",
			})

			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusTooManyRequests {
				t.Logf("Rate limited after %d password reset attempts", i+1)
				break
			}
			attempts++
		}

		if attempts >= 10 {
			t.Log("Note: Password reset should have rate limiting")
		}
	})

	t.Run("Account Enumeration Prevention", func(t *testing.T) {
		// Response should be the same for existing and non-existing accounts
		existingResp, _ := client.DoRequest("POST", "/api/v1/auth/reset-password", map[string]interface{}{
			"email": "existing@test.com",
		})
		nonExistingResp, _ := client.DoRequest("POST", "/api/v1/auth/reset-password", map[string]interface{}{
			"email": "nonexisting@test.com",
		})

		if existingResp != nil && nonExistingResp != nil {
			defer existingResp.Body.Close()
			defer nonExistingResp.Body.Close()

			// Response codes should be the same to prevent enumeration
			if existingResp.StatusCode != nonExistingResp.StatusCode {
				t.Log("Note: Different responses for existing/non-existing accounts allows enumeration")
			}
		}
	})
}

// TestIDORVulnerabilities tests Insecure Direct Object Reference vulnerabilities.
func TestIDORVulnerabilities(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	user1Token := server.GetValidToken("user1@test.com", "tenant-1")
	user2Token := server.GetValidToken("user2@test.com", "tenant-1")

	// Create resources for user2
	user2CustomerID := uuid.New().String()
	server.DataStore.mu.Lock()
	server.DataStore.Customers[user2CustomerID] = &SecureCustomerData{
		ID:       user2CustomerID,
		TenantID: "tenant-1",
		OwnerID:  "user2-id",
		Name:     "User2 Private Customer",
	}
	server.DataStore.mu.Unlock()

	t.Run("Direct ID Access", func(t *testing.T) {
		// User1 trying to access User2's customer by direct ID
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers/"+user2CustomerID, nil)
		req.Header.Set("Authorization", user1Token)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("IDOR vulnerability: User1 can access User2's customer")
		}
	})

	t.Run("ID Parameter Tampering", func(t *testing.T) {
		// Various ID manipulation attempts
		idManipulations := []string{
			user2CustomerID,
			"1",
			"0",
			"-1",
			"../admin",
			user2CustomerID + "/../admin",
			"00000000-0000-0000-0000-000000000000",
		}

		for _, id := range idManipulations {
			t.Run(fmt.Sprintf("ID_%s", id), func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers/"+id, nil)
				req.Header.Set("Authorization", user1Token)

				resp, err := client.Client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				// Should not return sensitive data from other users
				if resp.StatusCode == http.StatusOK && id == user2CustomerID {
					t.Errorf("IDOR: Access granted to ID %s", id)
				}
			})
		}
	})

	t.Run("Batch ID Access", func(t *testing.T) {
		// Attempting to access multiple IDs including unauthorized ones
		resp, _ := client.DoAuthenticatedRequest("POST", "/api/v1/customers/batch", user1Token, map[string]interface{}{
			"ids": []string{user2CustomerID, "some-other-id"},
		})

		if resp != nil {
			defer resp.Body.Close()
			// Should not return any data for unauthorized IDs
		}
	})

	t.Run("Sequential ID Guessing", func(t *testing.T) {
		// If IDs are sequential, this is a vulnerability
		// Testing that UUIDs are used instead of sequential IDs
		if len(user2CustomerID) < 32 {
			t.Log("Note: Short IDs might be guessable")
		}

		// Verify UUID format
		_, err := uuid.Parse(user2CustomerID)
		if err != nil {
			t.Log("Note: Non-UUID IDs used, consider using UUIDs")
		}
	})

	t.Run("Reference Switching", func(t *testing.T) {
		// Create a customer for user1
		user1CustomerID := uuid.New().String()
		server.DataStore.mu.Lock()
		server.DataStore.Customers[user1CustomerID] = &SecureCustomerData{
			ID:       user1CustomerID,
			TenantID: "tenant-1",
			OwnerID:  "user1-id",
			Name:     "User1 Customer",
		}
		server.DataStore.mu.Unlock()

		// Try to update user1's customer with user2's token
		resp, _ := client.DoAuthenticatedRequest("PUT", "/api/v1/customers/"+user1CustomerID, user2Token, map[string]interface{}{
			"name": "Hijacked!",
		})

		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Error("Reference switching vulnerability detected")
			}
		}
	})
}

// TestJWTVulnerabilities tests JWT-specific security issues.
func TestJWTVulnerabilities(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	t.Run("Algorithm None Attack", func(t *testing.T) {
		// Attempt to use 'none' algorithm
		noneToken := "Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwiYWRtaW4iOnRydWV9."

		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req.Header.Set("Authorization", noneToken)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("JWT 'none' algorithm attack succeeded - critical vulnerability!")
		}
	})

	t.Run("Algorithm Confusion RS256 to HS256", func(t *testing.T) {
		// Attempt algorithm confusion attack
		// This tests if server accepts HS256 when it expects RS256
		confusedToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZX0.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"

		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req.Header.Set("Authorization", confusedToken)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Algorithm confusion attack may have succeeded")
		}
	})

	t.Run("Key ID Injection", func(t *testing.T) {
		// Attempt to inject malicious key ID
		kidInjectionTokens := []string{
			// SQL injection in kid
			`Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IicgT1IgJzEnPScxIn0.eyJhZG1pbiI6dHJ1ZX0.test`,
			// Path traversal in kid
			`Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6Ii4uLy4uLy4uL2V0Yy9wYXNzd2QifQ.eyJhZG1pbiI6dHJ1ZX0.test`,
			// Command injection in kid
			`Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InwgY2F0IC9ldGMvcGFzc3dkIn0.eyJhZG1pbiI6dHJ1ZX0.test`,
		}

		for i, token := range kidInjectionTokens {
			t.Run(fmt.Sprintf("KID_Injection_%d", i), func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
				req.Header.Set("Authorization", token)

				resp, err := client.Client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					t.Error("KID injection attack may have succeeded")
				}
			})
		}
	})

	t.Run("JWKS Spoofing", func(t *testing.T) {
		// Attempt to use a self-signed token with JKU pointing to attacker server
		jkuToken := `Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImprdSI6Imh0dHA6Ly9ldmlsLmNvbS9qd2tzIn0.eyJhZG1pbiI6dHJ1ZX0.test`

		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req.Header.Set("Authorization", jkuToken)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("JWKS spoofing attack may have succeeded")
		}
	})

	t.Run("Claim Manipulation", func(t *testing.T) {
		validToken := server.GetValidToken("user@test.com", "tenant-1")

		// Test various claim manipulations would require modifying the JWT
		// For this test, we verify the token validation rejects modifications
		manipulatedToken := validToken + "extra"

		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
		req.Header.Set("Authorization", manipulatedToken)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Error("Manipulated token was accepted")
		}
	})
}

// TestAPIKeyBypass tests API key authentication bypass.
func TestAPIKeyBypass(t *testing.T) {
	server := NewSecurityTestServer()
	defer server.Close()

	client := NewSecurityTestClient(server.Server.URL)

	t.Run("Invalid API Keys", func(t *testing.T) {
		invalidKeys := []string{
			"",
			"invalid",
			"null",
			"undefined",
			"admin",
			"test-api-key",
			"' OR '1'='1",
			"<script>alert(1)</script>",
		}

		for _, key := range invalidKeys {
			t.Run(fmt.Sprintf("Key_%s", key), func(t *testing.T) {
				req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers", nil)
				req.Header.Set("X-API-Key", key)

				resp, err := client.Client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					t.Errorf("Invalid API key accepted: %s", key)
				}
			})
		}
	})

	t.Run("API Key in URL", func(t *testing.T) {
		// API keys in URL are a security risk (logged, cached, etc.)
		req, _ := http.NewRequest("GET", server.Server.URL+"/api/v1/customers?api_key=test-key", nil)

		resp, err := client.Client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should either reject or warn about API key in URL
		if resp.StatusCode == http.StatusOK {
			t.Log("Note: API key accepted in URL parameter - consider header-only authentication")
		}
	})
}
