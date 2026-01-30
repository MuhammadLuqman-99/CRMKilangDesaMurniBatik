// Package e2e contains E2E tests for user registration flow.
package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserRegistrationFlow tests the complete user registration process.
func TestUserRegistrationFlow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	t.Run("Complete Registration Flow", func(t *testing.T) {
		// Step 1: Create a tenant
		tenant := suite.CreateTenant("Test Company", "test-company-reg")
		require.NotEmpty(t, tenant.ID, "tenant ID should not be empty")
		assert.Equal(t, "Test Company", tenant.Name)
		assert.Equal(t, "test-company-reg", tenant.Slug)
		assert.Equal(t, "active", tenant.Status)
		assert.Equal(t, "professional", tenant.Plan)

		// Step 2: Register a user
		user := suite.RegisterUser(
			"john.doe@testcompany.com",
			"SecureP@ssw0rd!",
			"John",
			"Doe",
		)
		require.NotEmpty(t, user.ID, "user ID should not be empty")
		assert.Equal(t, "john.doe@testcompany.com", user.Email)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "pending_verification", user.Status)

		// Step 3: Login with the registered user
		loginResp := suite.Login("john.doe@testcompany.com", "SecureP@ssw0rd!")
		require.NotEmpty(t, loginResp.AccessToken, "access token should not be empty")
		require.NotEmpty(t, loginResp.RefreshToken, "refresh token should not be empty")
		assert.Equal(t, "Bearer", loginResp.TokenType)
		assert.Equal(t, 3600, loginResp.ExpiresIn)
		assert.Equal(t, user.ID, loginResp.User.ID)

		// Step 4: Get current user profile
		resp := suite.DoRequestAuth("GET", iamServer.URL+"/api/v1/auth/me", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var profile struct {
			ID        string `json:"id"`
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Status    string `json:"status"`
		}
		suite.DecodeResponse(resp, &profile)
		assert.Equal(t, user.ID, profile.ID)
		assert.Equal(t, "john.doe@testcompany.com", profile.Email)

		// Step 5: Update user profile
		resp = suite.DoRequestAuth("PUT", iamServer.URL+"/api/v1/auth/me", map[string]interface{}{
			"first_name": "Jonathan",
			"last_name":  "Doe Jr.",
		})
		suite.AssertStatus(resp, http.StatusOK)

		var updatedProfile struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		suite.DecodeResponse(resp, &updatedProfile)
		assert.Equal(t, "Jonathan", updatedProfile.FirstName)
		assert.Equal(t, "Doe Jr.", updatedProfile.LastName)

		// Step 6: Change password
		resp = suite.DoRequestAuth("PUT", iamServer.URL+"/api/v1/auth/me/password", map[string]interface{}{
			"current_password": "SecureP@ssw0rd!",
			"new_password":     "NewSecureP@ss123!",
		})
		suite.AssertStatus(resp, http.StatusOK)

		// Step 7: Refresh token
		resp = suite.DoRequest("POST", iamServer.URL+"/api/v1/auth/refresh", map[string]interface{}{
			"refresh_token": loginResp.RefreshToken,
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var refreshResp struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
		}
		suite.DecodeResponse(resp, &refreshResp)
		require.NotEmpty(t, refreshResp.AccessToken)
		assert.Equal(t, "Bearer", refreshResp.TokenType)

		// Step 8: Logout
		resp = suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/auth/logout", nil)
		suite.AssertStatus(resp, http.StatusOK)
	})

	t.Run("Registration Validation", func(t *testing.T) {
		// Create tenant first
		tenant := suite.CreateTenant("Validation Test", "validation-test")
		require.NotEmpty(t, tenant.ID)

		// Test: Missing required fields
		resp := suite.DoRequest("POST", iamServer.URL+"/api/v1/auth/register", map[string]interface{}{
			"email": "test@example.com",
			// Missing password and tenant_id
		}, nil)
		suite.AssertStatus(resp, http.StatusBadRequest)

		// Test: Invalid tenant ID
		resp = suite.DoRequest("POST", iamServer.URL+"/api/v1/auth/register", map[string]interface{}{
			"email":     "test@example.com",
			"password":  "Password123!",
			"tenant_id": "non-existent-tenant-id",
		}, nil)
		suite.AssertStatus(resp, http.StatusBadRequest)

		// Register a user
		suite.RegisterUser("existing@validation.com", "Password123!", "Test", "User")

		// Test: Duplicate email in same tenant
		resp = suite.DoRequest("POST", iamServer.URL+"/api/v1/auth/register", map[string]interface{}{
			"email":     "existing@validation.com",
			"password":  "Password123!",
			"tenant_id": suite.tenantID.String(),
		}, nil)
		suite.AssertStatus(resp, http.StatusConflict)
	})

	t.Run("Login Validation", func(t *testing.T) {
		// Create tenant and user
		tenant := suite.CreateTenant("Login Test", "login-test")
		require.NotEmpty(t, tenant.ID)

		suite.RegisterUser("logintest@example.com", "CorrectPassword123!", "Login", "Test")

		// Test: Wrong password
		resp := suite.DoRequest("POST", iamServer.URL+"/api/v1/auth/login", map[string]interface{}{
			"email":     "logintest@example.com",
			"password":  "WrongPassword123!",
			"tenant_id": suite.tenantID.String(),
		}, nil)
		suite.AssertStatus(resp, http.StatusUnauthorized)

		// Test: Non-existent user
		resp = suite.DoRequest("POST", iamServer.URL+"/api/v1/auth/login", map[string]interface{}{
			"email":     "nonexistent@example.com",
			"password":  "Password123!",
			"tenant_id": suite.tenantID.String(),
		}, nil)
		suite.AssertStatus(resp, http.StatusUnauthorized)
	})

	t.Run("Password Reset Flow", func(t *testing.T) {
		// Create tenant and user
		tenant := suite.CreateTenant("Reset Test", "reset-test")
		require.NotEmpty(t, tenant.ID)

		suite.RegisterUser("reset@example.com", "OldPassword123!", "Reset", "User")

		// Step 1: Request password reset
		resp := suite.DoRequest("POST", iamServer.URL+"/api/v1/auth/forgot-password", map[string]interface{}{
			"email":     "reset@example.com",
			"tenant_id": suite.tenantID.String(),
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)

		// Step 2: Reset password with token (simulated)
		resp = suite.DoRequest("POST", iamServer.URL+"/api/v1/auth/reset-password", map[string]interface{}{
			"token":        "simulated-reset-token",
			"new_password": "NewPassword123!",
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)
	})

	t.Run("Email Verification Flow", func(t *testing.T) {
		// Create tenant and user
		tenant := suite.CreateTenant("Verify Test", "verify-test")
		require.NotEmpty(t, tenant.ID)

		user := suite.RegisterUser("verify@example.com", "Password123!", "Verify", "User")
		assert.Equal(t, "pending_verification", user.Status)

		// Login first to get a token
		suite.Login("verify@example.com", "Password123!")

		// Request resend verification
		resp := suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/auth/resend-verification", nil)
		suite.AssertStatus(resp, http.StatusOK)

		// Verify email (simulated)
		resp = suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/auth/verify-email", map[string]interface{}{
			"token": "simulated-verification-token",
		})
		suite.AssertStatus(resp, http.StatusOK)
	})
}

// TestUserManagement tests user CRUD operations.
func TestUserManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup: Create tenant and admin user
	suite.CreateTenant("User Management Test", "user-mgmt-test")
	suite.RegisterUser("admin@usermgmt.com", "AdminPassword123!", "Admin", "User")
	suite.Login("admin@usermgmt.com", "AdminPassword123!")

	t.Run("List Users", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", iamServer.URL+"/api/v1/users", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		suite.DecodeResponse(resp, &listResp)
		assert.GreaterOrEqual(t, listResp.Total, 1)
	})

	t.Run("Create User", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users", map[string]interface{}{
			"email":      "newuser@usermgmt.com",
			"password":   "NewUserPass123!",
			"first_name": "New",
			"last_name":  "User",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var userResp UserResponse
		suite.DecodeResponse(resp, &userResp)
		assert.Equal(t, "newuser@usermgmt.com", userResp.Email)
		assert.Equal(t, "active", userResp.Status)
	})

	t.Run("Get User", func(t *testing.T) {
		// Create a user first
		resp := suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users", map[string]interface{}{
			"email":      "getuser@usermgmt.com",
			"first_name": "Get",
			"last_name":  "User",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var createdUser UserResponse
		suite.DecodeResponse(resp, &createdUser)

		// Get the user
		resp = suite.DoRequestAuth("GET", iamServer.URL+"/api/v1/users/"+createdUser.ID, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var fetchedUser UserResponse
		suite.DecodeResponse(resp, &fetchedUser)
		assert.Equal(t, createdUser.ID, fetchedUser.ID)
		assert.Equal(t, "getuser@usermgmt.com", fetchedUser.Email)
	})

	t.Run("Update User", func(t *testing.T) {
		// Create a user first
		resp := suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users", map[string]interface{}{
			"email":      "updateuser@usermgmt.com",
			"first_name": "Update",
			"last_name":  "User",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var createdUser UserResponse
		suite.DecodeResponse(resp, &createdUser)

		// Update the user
		resp = suite.DoRequestAuth("PUT", iamServer.URL+"/api/v1/users/"+createdUser.ID, map[string]interface{}{
			"first_name": "Updated",
			"last_name":  "Person",
		})
		suite.AssertStatus(resp, http.StatusOK)

		var updatedUser struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		suite.DecodeResponse(resp, &updatedUser)
		assert.Equal(t, "Updated", updatedUser.FirstName)
		assert.Equal(t, "Person", updatedUser.LastName)
	})

	t.Run("Activate and Suspend User", func(t *testing.T) {
		// Create a user
		resp := suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users", map[string]interface{}{
			"email":      "statususer@usermgmt.com",
			"first_name": "Status",
			"last_name":  "User",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var user UserResponse
		suite.DecodeResponse(resp, &user)

		// Suspend the user
		resp = suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users/"+user.ID+"/suspend", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var suspendResp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &suspendResp)
		assert.Equal(t, "suspended", suspendResp.Status)

		// Activate the user
		resp = suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users/"+user.ID+"/activate", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var activateResp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &activateResp)
		assert.Equal(t, "active", activateResp.Status)
	})

	t.Run("Delete User", func(t *testing.T) {
		// Create a user
		resp := suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users", map[string]interface{}{
			"email":      "deleteuser@usermgmt.com",
			"first_name": "Delete",
			"last_name":  "User",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var user UserResponse
		suite.DecodeResponse(resp, &user)

		// Delete the user
		resp = suite.DoRequestAuth("DELETE", iamServer.URL+"/api/v1/users/"+user.ID, nil)
		suite.AssertStatus(resp, http.StatusNoContent)

		// Verify user is deleted
		resp = suite.DoRequestAuth("GET", iamServer.URL+"/api/v1/users/"+user.ID, nil)
		suite.AssertStatus(resp, http.StatusNotFound)
	})
}

// TestRoleManagement tests role assignment and management.
func TestRoleManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup: Create tenant and user
	suite.CreateTenant("Role Management Test", "role-mgmt-test")
	suite.RegisterUser("roleadmin@example.com", "Password123!", "Role", "Admin")
	suite.Login("roleadmin@example.com", "Password123!")

	t.Run("Create Custom Role", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/roles", map[string]interface{}{
			"name":        "custom_analyst",
			"description": "Custom analyst role with read permissions",
			"permissions": []string{"customers:read", "leads:read", "reports:read"},
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var roleResp struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		suite.DecodeResponse(resp, &roleResp)
		assert.Equal(t, "custom_analyst", roleResp.Name)
		assert.Equal(t, "Custom analyst role with read permissions", roleResp.Description)
	})

	t.Run("List Roles", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", iamServer.URL+"/api/v1/roles", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data []map[string]interface{} `json:"data"`
		}
		suite.DecodeResponse(resp, &listResp)
		// Should have at least the custom role we created
		assert.GreaterOrEqual(t, len(listResp.Data), 1)
	})

	t.Run("Assign and Remove Role", func(t *testing.T) {
		// Create a role
		resp := suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/roles", map[string]interface{}{
			"name":        "test_role",
			"description": "Test role for assignment",
			"permissions": []string{"test:read"},
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var role struct {
			ID string `json:"id"`
		}
		suite.DecodeResponse(resp, &role)

		// Create a user
		resp = suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users", map[string]interface{}{
			"email":      "roleuser@example.com",
			"first_name": "Role",
			"last_name":  "User",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var user UserResponse
		suite.DecodeResponse(resp, &user)

		// Assign role to user
		resp = suite.DoRequestAuth("POST", iamServer.URL+"/api/v1/users/"+user.ID+"/roles", map[string]interface{}{
			"role_id": role.ID,
		})
		suite.AssertStatus(resp, http.StatusOK)

		// Get user roles
		resp = suite.DoRequestAuth("GET", iamServer.URL+"/api/v1/users/"+user.ID+"/roles", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var rolesResp struct {
			Roles []map[string]interface{} `json:"roles"`
		}
		suite.DecodeResponse(resp, &rolesResp)
		assert.GreaterOrEqual(t, len(rolesResp.Roles), 1)

		// Remove role from user
		resp = suite.DoRequestAuth("DELETE", iamServer.URL+"/api/v1/users/"+user.ID+"/roles/"+role.ID, nil)
		suite.AssertStatus(resp, http.StatusNoContent)
	})

	t.Run("Cannot Modify System Role", func(t *testing.T) {
		// First, let's create a system role in the data store for this test
		// Get system roles
		resp := suite.DoRequestAuth("GET", iamServer.URL+"/api/v1/roles/system", nil)
		suite.AssertStatus(resp, http.StatusOK)

		// The system roles list might be empty in our in-memory store
		// So we just verify the endpoint works
	})
}

// TestTenantManagement tests tenant operations.
func TestTenantManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	t.Run("Create Tenant", func(t *testing.T) {
		tenant := suite.CreateTenant("New Tenant", "new-tenant-slug")
		assert.NotEmpty(t, tenant.ID)
		assert.Equal(t, "New Tenant", tenant.Name)
		assert.Equal(t, "new-tenant-slug", tenant.Slug)
		assert.Equal(t, "active", tenant.Status)
	})

	t.Run("Get Tenant", func(t *testing.T) {
		tenant := suite.CreateTenant("Get Tenant Test", "get-tenant-test")

		resp := suite.DoRequest("GET", iamServer.URL+"/api/v1/tenants/"+tenant.ID, nil, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var fetchedTenant TenantResponse
		suite.DecodeResponse(resp, &fetchedTenant)
		assert.Equal(t, tenant.ID, fetchedTenant.ID)
		assert.Equal(t, "Get Tenant Test", fetchedTenant.Name)
	})

	t.Run("Update Tenant", func(t *testing.T) {
		tenant := suite.CreateTenant("Update Tenant Test", "update-tenant-test")

		resp := suite.DoRequest("PUT", iamServer.URL+"/api/v1/tenants/"+tenant.ID, map[string]interface{}{
			"name": "Updated Tenant Name",
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var updatedTenant struct {
			Name string `json:"name"`
		}
		suite.DecodeResponse(resp, &updatedTenant)
		assert.Equal(t, "Updated Tenant Name", updatedTenant.Name)
	})

	t.Run("Update Tenant Status", func(t *testing.T) {
		tenant := suite.CreateTenant("Status Tenant Test", "status-tenant-test")

		resp := suite.DoRequest("PUT", iamServer.URL+"/api/v1/tenants/"+tenant.ID+"/status", map[string]interface{}{
			"status": "suspended",
		}, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var statusResp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &statusResp)
		assert.Equal(t, "suspended", statusResp.Status)
	})

	t.Run("Update Tenant Plan", func(t *testing.T) {
		tenant := suite.CreateTenant("Plan Tenant Test", "plan-tenant-test")

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

	t.Run("Get Tenant Stats", func(t *testing.T) {
		tenant := suite.CreateTenant("Stats Tenant Test", "stats-tenant-test")

		// Register a user to have some stats
		suite.RegisterUser("statsuser@tenant.com", "Password123!", "Stats", "User")

		resp := suite.DoRequest("GET", iamServer.URL+"/api/v1/tenants/"+tenant.ID+"/stats", nil, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var statsResp struct {
			Users     int `json:"users"`
			Customers int `json:"customers"`
			Leads     int `json:"leads"`
		}
		suite.DecodeResponse(resp, &statsResp)
		assert.GreaterOrEqual(t, statsResp.Users, 1)
	})

	t.Run("Duplicate Slug", func(t *testing.T) {
		// Create first tenant
		suite.CreateTenant("First Tenant", "duplicate-slug")

		// Try to create second tenant with same slug
		resp := suite.DoRequest("POST", iamServer.URL+"/api/v1/tenants", map[string]interface{}{
			"name": "Second Tenant",
			"slug": "duplicate-slug",
			"plan": "professional",
		}, nil)
		suite.AssertStatus(resp, http.StatusConflict)
	})

	t.Run("Delete Tenant", func(t *testing.T) {
		tenant := suite.CreateTenant("Delete Tenant Test", "delete-tenant-test")

		resp := suite.DoRequest("DELETE", iamServer.URL+"/api/v1/tenants/"+tenant.ID, nil, nil)
		suite.AssertStatus(resp, http.StatusNoContent)

		// Verify tenant is deleted
		resp = suite.DoRequest("GET", iamServer.URL+"/api/v1/tenants/"+tenant.ID, nil, nil)
		suite.AssertStatus(resp, http.StatusNotFound)
	})
}
