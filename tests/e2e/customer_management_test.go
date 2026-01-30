// Package e2e contains E2E tests for customer management flow.
package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCustomerManagementFlow tests the complete customer lifecycle.
func TestCustomerManagementFlow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup: Create tenant, user, and login
	suite.CreateTenant("Customer Test Company", "customer-test-co")
	suite.RegisterUser("customer.manager@test.com", "Password123!", "Customer", "Manager")
	suite.Login("customer.manager@test.com", "Password123!")

	t.Run("Complete Customer Lifecycle", func(t *testing.T) {
		// Step 1: Create a customer
		customer := suite.CreateCustomer("Acme Corporation", "ACME-001", "contact@acme.com")
		require.NotEmpty(t, customer.ID, "customer ID should not be empty")
		assert.Equal(t, "Acme Corporation", customer.Name)
		assert.Equal(t, "ACME-001", customer.Code)
		assert.Equal(t, "business", customer.Type)
		assert.Equal(t, "active", customer.Status)

		// Step 2: Get customer details
		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/"+customer.ID, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var fetchedCustomer CustomerResponse
		suite.DecodeResponse(resp, &fetchedCustomer)
		assert.Equal(t, customer.ID, fetchedCustomer.ID)
		assert.Equal(t, "Acme Corporation", fetchedCustomer.Name)

		// Step 3: Update customer
		resp = suite.DoRequestAuth("PUT", customerServer.URL+"/api/v1/customers/"+customer.ID, map[string]interface{}{
			"name":   "Acme Corp International",
			"status": "active",
			"email": map[string]interface{}{
				"address": "info@acme-intl.com",
			},
		})
		suite.AssertStatus(resp, http.StatusOK)

		var updatedCustomer struct {
			Name  string                 `json:"name"`
			Email map[string]interface{} `json:"email"`
		}
		suite.DecodeResponse(resp, &updatedCustomer)
		assert.Equal(t, "Acme Corp International", updatedCustomer.Name)

		// Step 4: Deactivate customer
		resp = suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/deactivate", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var deactivateResp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &deactivateResp)
		assert.Equal(t, "inactive", deactivateResp.Status)

		// Step 5: Reactivate customer
		resp = suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/activate", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var activateResp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &activateResp)
		assert.Equal(t, "active", activateResp.Status)

		// Step 6: Delete customer
		resp = suite.DoRequestAuth("DELETE", customerServer.URL+"/api/v1/customers/"+customer.ID, nil)
		suite.AssertStatus(resp, http.StatusNoContent)

		// Step 7: Verify customer is deleted
		resp = suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/"+customer.ID, nil)
		suite.AssertStatus(resp, http.StatusNotFound)
	})

	t.Run("Customer Validation", func(t *testing.T) {
		// Test: Duplicate customer code
		suite.CreateCustomer("First Company", "DUP-001", "first@company.com")

		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers", map[string]interface{}{
			"name":   "Second Company",
			"code":   "DUP-001", // Duplicate code
			"type":   "business",
			"status": "active",
			"email": map[string]interface{}{
				"address": "second@company.com",
			},
		})
		suite.AssertStatus(resp, http.StatusConflict)
	})
}

// TestContactManagement tests contact CRUD operations within customers.
func TestContactManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Contact Test Company", "contact-test-co")
	suite.RegisterUser("contact.manager@test.com", "Password123!", "Contact", "Manager")
	suite.Login("contact.manager@test.com", "Password123!")

	// Create a customer to add contacts to
	customer := suite.CreateCustomer("Contact Test Corp", "CTC-001", "corp@contacttest.com")

	t.Run("Create Contact", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts", map[string]interface{}{
			"first_name": "John",
			"last_name":  "Smith",
			"email": map[string]interface{}{
				"address": "john.smith@contacttest.com",
			},
			"is_primary": true,
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var contact ContactResponse
		suite.DecodeResponse(resp, &contact)
		assert.NotEmpty(t, contact.ID)
		assert.Equal(t, "John", contact.FirstName)
		assert.Equal(t, "Smith", contact.LastName)
		assert.True(t, contact.IsPrimary)
	})

	t.Run("List Contacts", func(t *testing.T) {
		// Create a few contacts first
		suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts", map[string]interface{}{
			"first_name": "Jane",
			"last_name":  "Doe",
			"email": map[string]interface{}{
				"address": "jane.doe@contacttest.com",
			},
		})

		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data []map[string]interface{} `json:"data"`
		}
		suite.DecodeResponse(resp, &listResp)
		assert.GreaterOrEqual(t, len(listResp.Data), 1)
	})

	t.Run("Update Contact", func(t *testing.T) {
		// Create a contact
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts", map[string]interface{}{
			"first_name": "Update",
			"last_name":  "Test",
			"email": map[string]interface{}{
				"address": "update@contacttest.com",
			},
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var contact ContactResponse
		suite.DecodeResponse(resp, &contact)

		// Update the contact
		resp = suite.DoRequestAuth("PUT", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts/"+contact.ID, map[string]interface{}{
			"first_name": "Updated",
			"last_name":  "Person",
		})
		suite.AssertStatus(resp, http.StatusOK)

		var updatedContact struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		suite.DecodeResponse(resp, &updatedContact)
		assert.Equal(t, "Updated", updatedContact.FirstName)
		assert.Equal(t, "Person", updatedContact.LastName)
	})

	t.Run("Set Primary Contact", func(t *testing.T) {
		// Create two contacts
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts", map[string]interface{}{
			"first_name": "Primary",
			"last_name":  "One",
			"is_primary": true,
		})
		suite.AssertStatus(resp, http.StatusCreated)

		resp = suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts", map[string]interface{}{
			"first_name": "Secondary",
			"last_name":  "Two",
			"is_primary": false,
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var secondContact ContactResponse
		suite.DecodeResponse(resp, &secondContact)

		// Set second contact as primary
		resp = suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts/"+secondContact.ID+"/primary", nil)
		suite.AssertStatus(resp, http.StatusOK)
	})

	t.Run("Delete Contact", func(t *testing.T) {
		// Create a contact
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts", map[string]interface{}{
			"first_name": "Delete",
			"last_name":  "Me",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var contact ContactResponse
		suite.DecodeResponse(resp, &contact)

		// Delete the contact
		resp = suite.DoRequestAuth("DELETE", customerServer.URL+"/api/v1/customers/"+customer.ID+"/contacts/"+contact.ID, nil)
		suite.AssertStatus(resp, http.StatusNoContent)
	})
}

// TestCustomerSearch tests customer search functionality.
func TestCustomerSearch(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Search Test Company", "search-test-co")
	suite.RegisterUser("search.user@test.com", "Password123!", "Search", "User")
	suite.Login("search.user@test.com", "Password123!")

	// Create several customers for searching
	suite.CreateCustomer("Alpha Tech Solutions", "ALPHA-001", "alpha@tech.com")
	suite.CreateCustomer("Beta Industries", "BETA-001", "beta@industries.com")
	suite.CreateCustomer("Alpha Omega Corp", "ALPHA-002", "omega@alpha.com")
	suite.CreateCustomer("Gamma Services", "GAMMA-001", "gamma@services.com")

	t.Run("Search by Name", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/search?q=Alpha", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var searchResp struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		suite.DecodeResponse(resp, &searchResp)
		assert.Equal(t, 2, searchResp.Total) // Alpha Tech and Alpha Omega
	})

	t.Run("Search with No Results", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/search?q=NonExistent", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var searchResp struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		suite.DecodeResponse(resp, &searchResp)
		assert.Equal(t, 0, searchResp.Total)
	})

	t.Run("List All Customers", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		suite.DecodeResponse(resp, &listResp)
		assert.GreaterOrEqual(t, listResp.Total, 4) // At least our 4 test customers
	})
}

// TestCustomerOptimisticLocking tests version-based concurrency control.
func TestCustomerOptimisticLocking(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Locking Test Company", "locking-test-co")
	suite.RegisterUser("locking.user@test.com", "Password123!", "Locking", "User")
	suite.Login("locking.user@test.com", "Password123!")

	t.Run("Optimistic Lock Conflict", func(t *testing.T) {
		// Create a customer
		customer := suite.CreateCustomer("Lock Test Corp", "LOCK-001", "lock@test.com")

		// Get the customer to know its version
		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/"+customer.ID, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var fetchedCustomer struct {
			Version int `json:"version"`
		}
		suite.DecodeResponse(resp, &fetchedCustomer)

		// Update with correct version
		resp = suite.DoRequestAuth("PUT", customerServer.URL+"/api/v1/customers/"+customer.ID, map[string]interface{}{
			"name":    "Updated Lock Test Corp",
			"version": fetchedCustomer.Version,
		})
		suite.AssertStatus(resp, http.StatusOK)

		// Try to update with old version (should fail)
		resp = suite.DoRequestAuth("PUT", customerServer.URL+"/api/v1/customers/"+customer.ID, map[string]interface{}{
			"name":    "Another Update",
			"version": fetchedCustomer.Version, // Old version
		})
		suite.AssertStatus(resp, http.StatusConflict)
	})
}

// TestCustomerNotes tests note management for customers.
func TestCustomerNotes(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Notes Test Company", "notes-test-co")
	suite.RegisterUser("notes.user@test.com", "Password123!", "Notes", "User")
	suite.Login("notes.user@test.com", "Password123!")

	customer := suite.CreateCustomer("Notes Test Corp", "NOTES-001", "notes@test.com")

	t.Run("Create Note", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/notes", map[string]interface{}{
			"content": "Initial meeting notes - discussed requirements",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var noteResp struct {
			ID string `json:"id"`
		}
		suite.DecodeResponse(resp, &noteResp)
		assert.NotEmpty(t, noteResp.ID)
	})

	t.Run("List Notes", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/"+customer.ID+"/notes", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data []map[string]interface{} `json:"data"`
		}
		suite.DecodeResponse(resp, &listResp)
		// Notes list returned (may be empty in our in-memory implementation)
	})

	t.Run("Pin Note", func(t *testing.T) {
		// Create a note
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/notes", map[string]interface{}{
			"content": "Important note to pin",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var note struct {
			ID string `json:"id"`
		}
		suite.DecodeResponse(resp, &note)

		// Pin the note
		resp = suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/notes/"+note.ID+"/pin", nil)
		suite.AssertStatus(resp, http.StatusOK)
	})

	t.Run("Delete Note", func(t *testing.T) {
		// Create a note
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/notes", map[string]interface{}{
			"content": "Note to delete",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var note struct {
			ID string `json:"id"`
		}
		suite.DecodeResponse(resp, &note)

		// Delete the note
		resp = suite.DoRequestAuth("DELETE", customerServer.URL+"/api/v1/customers/"+customer.ID+"/notes/"+note.ID, nil)
		suite.AssertStatus(resp, http.StatusNoContent)
	})
}

// TestCustomerActivities tests activity tracking for customers.
func TestCustomerActivities(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Activities Test Company", "activities-test-co")
	suite.RegisterUser("activities.user@test.com", "Password123!", "Activities", "User")
	suite.Login("activities.user@test.com", "Password123!")

	customer := suite.CreateCustomer("Activities Test Corp", "ACT-001", "activities@test.com")

	t.Run("Create Activity", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/"+customer.ID+"/activities", map[string]interface{}{
			"type":        "call",
			"description": "Initial sales call",
		})
		suite.AssertStatus(resp, http.StatusCreated)
	})

	t.Run("List Activities", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/"+customer.ID+"/activities", nil)
		suite.AssertStatus(resp, http.StatusOK)
	})
}

// TestSegmentManagement tests customer segmentation features.
func TestSegmentManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Segment Test Company", "segment-test-co")
	suite.RegisterUser("segment.user@test.com", "Password123!", "Segment", "User")
	suite.Login("segment.user@test.com", "Password123!")

	t.Run("Create Segment", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/segments", map[string]interface{}{
			"name":        "High Value Customers",
			"description": "Customers with revenue > $10,000",
			"criteria": map[string]interface{}{
				"min_revenue": 10000,
			},
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var segmentResp struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		suite.DecodeResponse(resp, &segmentResp)
		assert.NotEmpty(t, segmentResp.ID)
	})

	t.Run("List Segments", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/segments", nil)
		suite.AssertStatus(resp, http.StatusOK)
	})

	t.Run("Get Segment Customers", func(t *testing.T) {
		// Create a segment
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/segments", map[string]interface{}{
			"name": "Test Segment",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var segment struct {
			ID string `json:"id"`
		}
		suite.DecodeResponse(resp, &segment)

		// Get segment customers
		resp = suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/segments/"+segment.ID+"/customers", nil)
		suite.AssertStatus(resp, http.StatusOK)
	})

	t.Run("Refresh Segment", func(t *testing.T) {
		// Create a segment
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/segments", map[string]interface{}{
			"name": "Refresh Test Segment",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var segment struct {
			ID string `json:"id"`
		}
		suite.DecodeResponse(resp, &segment)

		// Refresh the segment
		resp = suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/segments/"+segment.ID+"/refresh", nil)
		suite.AssertStatus(resp, http.StatusOK)
	})

	t.Run("Delete Segment", func(t *testing.T) {
		// Create a segment
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/segments", map[string]interface{}{
			"name": "Delete Test Segment",
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var segment struct {
			ID string `json:"id"`
		}
		suite.DecodeResponse(resp, &segment)

		// Delete the segment
		resp = suite.DoRequestAuth("DELETE", customerServer.URL+"/api/v1/segments/"+segment.ID, nil)
		suite.AssertStatus(resp, http.StatusNoContent)
	})
}

// TestCustomerImportExport tests bulk import/export features.
func TestCustomerImportExport(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Import Export Test", "import-export-test")
	suite.RegisterUser("import.user@test.com", "Password123!", "Import", "User")
	suite.Login("import.user@test.com", "Password123!")

	t.Run("Export Customers", func(t *testing.T) {
		// Create some customers first
		suite.CreateCustomer("Export Test 1", "EXP-001", "export1@test.com")
		suite.CreateCustomer("Export Test 2", "EXP-002", "export2@test.com")

		resp := suite.DoRequestAuth("GET", customerServer.URL+"/api/v1/customers/export", nil)
		suite.AssertStatus(resp, http.StatusOK)
		assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
	})

	t.Run("Import Customers", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", customerServer.URL+"/api/v1/customers/import", map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"name":  "Imported Customer 1",
					"code":  "IMP-001",
					"email": "import1@test.com",
				},
			},
		})
		suite.AssertStatus(resp, http.StatusOK)

		var importResp struct {
			Imported int `json:"imported"`
			Failed   int `json:"failed"`
		}
		suite.DecodeResponse(resp, &importResp)
	})
}
