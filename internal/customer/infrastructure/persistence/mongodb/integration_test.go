// Package mongodb contains MongoDB repository integration tests.
package mongodb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/kilang-desa-murni/crm/internal/customer/domain"
	"github.com/kilang-desa-murni/crm/pkg/testing/containers"
	"github.com/kilang-desa-murni/crm/pkg/testing/fixtures"
	"github.com/kilang-desa-murni/crm/pkg/testing/helpers"
)

var (
	testMongoDB   *containers.MongoDBContainer
	customerRepo  *CustomerRepository
	contactRepo   *ContactRepository
	testCtx       context.Context
	testCtxCancel context.CancelFunc
)

// TestMain sets up and tears down the test database.
func TestMain(m *testing.M) {
	// Skip if in short mode
	if testing.Short() {
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup MongoDB container
	var err error
	testMongoDB, err = containers.NewMongoDBContainer(ctx, containers.DefaultMongoDBConfig())
	if err != nil {
		panic("failed to create MongoDB container: " + err.Error())
	}

	// Setup collections and indexes
	if err := testMongoDB.SetupCustomerCollections(ctx); err != nil {
		panic("failed to setup collections: " + err.Error())
	}

	// Initialize repositories
	customerRepo = NewCustomerRepository(testMongoDB.GetDB())
	contactRepo = NewContactRepository(testMongoDB.GetDB())

	// Run tests
	code := m.Run()

	// Cleanup
	if testMongoDB != nil {
		testMongoDB.Close(context.Background())
	}

	os.Exit(code)
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
	// Clean up test data
	testMongoDB.CleanAllCollections(context.Background())
}

// ============================================================================
// Customer Repository Integration Tests
// ============================================================================

func TestCustomerRepository_Create(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("successfully creates a new customer", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "CUST-001", "Test Customer 1")

		err := customerRepo.Create(testCtx, customer)
		helpers.AssertNoError(t, err)

		// Verify customer was created
		found, err := customerRepo.FindByID(testCtx, customer.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, customer.ID, found.ID)
		helpers.AssertEqual(t, "Test Customer 1", found.Name)
		helpers.AssertEqual(t, 1, found.Version)
	})

	t.Run("fails with duplicate code in same tenant", func(t *testing.T) {
		customer1 := createTestCustomer(tenantID, "DUP-CODE", "Customer 1")
		err := customerRepo.Create(testCtx, customer1)
		helpers.AssertNoError(t, err)

		customer2 := createTestCustomer(tenantID, "DUP-CODE", "Customer 2")
		err = customerRepo.Create(testCtx, customer2)
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrCustomerAlreadyExists, err)
	})

	t.Run("allows same code in different tenants", func(t *testing.T) {
		tenantID2 := fixtures.TestIDs.TenantID2

		customer1 := createTestCustomer(tenantID, "SAME-CODE", "Tenant 1 Customer")
		err := customerRepo.Create(testCtx, customer1)
		helpers.AssertNoError(t, err)

		customer2 := createTestCustomer(tenantID2, "SAME-CODE", "Tenant 2 Customer")
		err = customerRepo.Create(testCtx, customer2)
		helpers.AssertNoError(t, err)
	})
}

func TestCustomerRepository_Update(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("successfully updates customer", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "UPDATE-001", "Original Name")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		// Update customer
		customer.Name = "Updated Name"
		err = customerRepo.Update(testCtx, customer)
		helpers.AssertNoError(t, err)

		// Verify update
		found, err := customerRepo.FindByID(testCtx, customer.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "Updated Name", found.Name)
		helpers.AssertEqual(t, 2, found.Version)
	})

	t.Run("fails with optimistic locking conflict", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "CONFLICT-001", "Conflict Test")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		// Simulate concurrent update
		customer1, _ := customerRepo.FindByID(testCtx, customer.ID)
		customer2, _ := customerRepo.FindByID(testCtx, customer.ID)

		// Update customer1
		customer1.Name = "Updated by 1"
		err = customerRepo.Update(testCtx, customer1)
		helpers.AssertNoError(t, err)

		// Update customer2 (should fail due to version conflict)
		customer2.Name = "Updated by 2"
		err = customerRepo.Update(testCtx, customer2)
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrVersionConflict, err)
	})

	t.Run("fails to update non-existent customer", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "NOEXIST-001", "Non-existent")
		err := customerRepo.Update(testCtx, customer)
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrCustomerNotFound, err)
	})
}

func TestCustomerRepository_Delete(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("successfully soft deletes customer", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "DELETE-001", "To Delete")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		err = customerRepo.Delete(testCtx, customer.ID)
		helpers.AssertNoError(t, err)

		// Customer should not be found (soft deleted)
		_, err = customerRepo.FindByID(testCtx, customer.ID)
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrCustomerNotFound, err)
	})

	t.Run("hard delete permanently removes customer", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "HARDDELETE-001", "Hard Delete")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		err = customerRepo.HardDelete(testCtx, customer.ID)
		helpers.AssertNoError(t, err)

		// Verify no trace remains
		var result domain.Customer
		err = testMongoDB.Collection("customers").FindOne(testCtx, bson.M{"_id": customer.ID}).Decode(&result)
		helpers.AssertError(t, err)
	})
}

func TestCustomerRepository_FindByCode(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("finds customer by code", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "FIND-BY-CODE", "Find By Code Test")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		found, err := customerRepo.FindByCode(testCtx, tenantID, "FIND-BY-CODE")
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, customer.ID, found.ID)
	})

	t.Run("returns error for non-existent code", func(t *testing.T) {
		_, err := customerRepo.FindByCode(testCtx, tenantID, "NONEXISTENT")
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrCustomerNotFound, err)
	})
}

func TestCustomerRepository_FindByEmail(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("finds customer by email", func(t *testing.T) {
		customer := createTestCustomerWithEmail(tenantID, "EMAIL-001", "Email Test", "findme@example.com")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		found, err := customerRepo.FindByEmail(testCtx, tenantID, "findme@example.com")
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, customer.ID, found.ID)
	})

	t.Run("case insensitive email search", func(t *testing.T) {
		customer := createTestCustomerWithEmail(tenantID, "EMAIL-002", "Case Test", "casetest@example.com")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		found, err := customerRepo.FindByEmail(testCtx, tenantID, "CASETEST@example.com")
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
	})
}

func TestCustomerRepository_List(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("lists customers with pagination", func(t *testing.T) {
		// Create 25 customers
		for i := 0; i < 25; i++ {
			customer := createTestCustomer(tenantID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(20))
			err := customerRepo.Create(testCtx, customer)
			helpers.RequireNoError(t, err)
		}

		// First page
		filter := domain.CustomerFilter{
			TenantID: &tenantID,
			Offset:   0,
			Limit:    10,
		}
		result, err := customerRepo.List(testCtx, filter)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(25), result.Total)
		helpers.AssertLen(t, result.Customers, 10)
		helpers.AssertTrue(t, result.HasMore)

		// Second page
		filter.Offset = 10
		result, err = customerRepo.List(testCtx, filter)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, result.Customers, 10)

		// Third page
		filter.Offset = 20
		result, err = customerRepo.List(testCtx, filter)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, result.Customers, 5)
		helpers.AssertFalse(t, result.HasMore)
	})

	t.Run("filters by status", func(t *testing.T) {
		// Create customers with different statuses
		activeCustomer := createTestCustomer(tenantID, "ACTIVE-001", "Active Customer")
		activeCustomer.Status = domain.CustomerStatusActive
		err := customerRepo.Create(testCtx, activeCustomer)
		helpers.RequireNoError(t, err)

		inactiveCustomer := createTestCustomer(tenantID, "INACTIVE-001", "Inactive Customer")
		inactiveCustomer.Status = domain.CustomerStatusInactive
		err = customerRepo.Create(testCtx, inactiveCustomer)
		helpers.RequireNoError(t, err)

		filter := domain.CustomerFilter{
			TenantID: &tenantID,
			Statuses: []domain.CustomerStatus{domain.CustomerStatusActive},
			Offset:   0,
			Limit:    100,
		}
		result, err := customerRepo.List(testCtx, filter)
		helpers.AssertNoError(t, err)

		for _, c := range result.Customers {
			helpers.AssertEqual(t, domain.CustomerStatusActive, c.Status)
		}
	})

	t.Run("filters by tags", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "TAGGED-001", "Tagged Customer")
		customer.Tags = []string{"vip", "priority"}
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		filter := domain.CustomerFilter{
			TenantID: &tenantID,
			Tags:     []string{"vip"},
			Offset:   0,
			Limit:    100,
		}
		result, err := customerRepo.List(testCtx, filter)
		helpers.AssertNoError(t, err)
		helpers.AssertGreater(t, len(result.Customers), 0)

		for _, c := range result.Customers {
			helpers.AssertContains(t, stringSliceToString(c.Tags), "vip")
		}
	})

	t.Run("sorts results", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			customer := createTestCustomer(tenantID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(20))
			err := customerRepo.Create(testCtx, customer)
			helpers.RequireNoError(t, err)
			time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		}

		// Sort by created_at ascending
		filter := domain.CustomerFilter{
			TenantID:  &tenantID,
			SortBy:    "created_at",
			SortOrder: "asc",
			Offset:    0,
			Limit:     100,
		}
		result, err := customerRepo.List(testCtx, filter)
		helpers.AssertNoError(t, err)

		for i := 1; i < len(result.Customers); i++ {
			helpers.AssertTrue(t, result.Customers[i-1].CreatedAt.Before(result.Customers[i].CreatedAt) ||
				result.Customers[i-1].CreatedAt.Equal(result.Customers[i].CreatedAt))
		}
	})
}

func TestCustomerRepository_Search(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("searches by text", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "SEARCH-001", "Searchable Batik Company")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		filter := domain.CustomerFilter{
			Offset: 0,
			Limit:  100,
		}
		result, err := customerRepo.Search(testCtx, tenantID, "Batik", filter)
		helpers.AssertNoError(t, err)
		helpers.AssertGreater(t, len(result.Customers), 0)
	})
}

func TestCustomerRepository_BulkOperations(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("bulk creates customers", func(t *testing.T) {
		customers := make([]*domain.Customer, 10)
		for i := 0; i < 10; i++ {
			customers[i] = createTestCustomer(tenantID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(20))
		}

		err := customerRepo.BulkCreate(testCtx, customers)
		helpers.AssertNoError(t, err)

		// Verify all were created
		count, err := customerRepo.CountByTenant(testCtx, tenantID)
		helpers.AssertNoError(t, err)
		helpers.AssertGreaterOrEqual(t, int(count), 10)
	})

	t.Run("bulk updates customers", func(t *testing.T) {
		// Create customers
		ids := make([]uuid.UUID, 5)
		for i := 0; i < 5; i++ {
			customer := createTestCustomer(tenantID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(20))
			err := customerRepo.Create(testCtx, customer)
			helpers.RequireNoError(t, err)
			ids[i] = customer.ID
		}

		// Bulk update
		updates := map[string]interface{}{
			"tier": "premium",
		}
		err := customerRepo.BulkUpdate(testCtx, ids, updates)
		helpers.AssertNoError(t, err)

		// Verify updates
		for _, id := range ids {
			customer, err := customerRepo.FindByID(testCtx, id)
			helpers.AssertNoError(t, err)
			helpers.AssertEqual(t, domain.CustomerTier("premium"), customer.Tier)
		}
	})

	t.Run("bulk deletes customers", func(t *testing.T) {
		// Create customers
		ids := make([]uuid.UUID, 5)
		for i := 0; i < 5; i++ {
			customer := createTestCustomer(tenantID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(20))
			err := customerRepo.Create(testCtx, customer)
			helpers.RequireNoError(t, err)
			ids[i] = customer.ID
		}

		// Bulk delete
		err := customerRepo.BulkDelete(testCtx, ids)
		helpers.AssertNoError(t, err)

		// Verify deletions
		for _, id := range ids {
			_, err := customerRepo.FindByID(testCtx, id)
			helpers.AssertError(t, err)
			helpers.AssertEqual(t, domain.ErrCustomerNotFound, err)
		}
	})
}

func TestCustomerRepository_FindDuplicates(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("finds duplicates by email", func(t *testing.T) {
		customer := createTestCustomerWithEmail(tenantID, "DUP-EMAIL", "Duplicate Check", "duplicate@example.com")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		duplicates, err := customerRepo.FindDuplicates(testCtx, tenantID, "duplicate@example.com", "", "")
		helpers.AssertNoError(t, err)
		helpers.AssertGreater(t, len(duplicates), 0)
	})

	t.Run("finds duplicates by name", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "DUP-NAME", "Exact Name Match")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		duplicates, err := customerRepo.FindDuplicates(testCtx, tenantID, "", "", "Exact Name Match")
		helpers.AssertNoError(t, err)
		helpers.AssertGreater(t, len(duplicates), 0)
	})
}

func TestCustomerRepository_CountByStatus(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("counts by status", func(t *testing.T) {
		// Create customers with different statuses
		for i := 0; i < 3; i++ {
			customer := createTestCustomer(tenantID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(20))
			customer.Status = domain.CustomerStatusActive
			err := customerRepo.Create(testCtx, customer)
			helpers.RequireNoError(t, err)
		}

		for i := 0; i < 2; i++ {
			customer := createTestCustomer(tenantID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(20))
			customer.Status = domain.CustomerStatusInactive
			err := customerRepo.Create(testCtx, customer)
			helpers.RequireNoError(t, err)
		}

		counts, err := customerRepo.CountByStatus(testCtx, tenantID)
		helpers.AssertNoError(t, err)
		helpers.AssertGreaterOrEqual(t, int(counts[domain.CustomerStatusActive]), 3)
		helpers.AssertGreaterOrEqual(t, int(counts[domain.CustomerStatusInactive]), 2)
	})
}

func TestCustomerRepository_GetStats(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("gets customer statistics", func(t *testing.T) {
		// Create some customers
		for i := 0; i < 5; i++ {
			customer := createTestCustomer(tenantID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(20))
			err := customerRepo.Create(testCtx, customer)
			helpers.RequireNoError(t, err)
		}

		stats, err := customerRepo.GetStats(testCtx, tenantID)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, stats)
		helpers.AssertNotNil(t, stats.LastCalculatedAt)
	})
}

func TestCustomerRepository_Exists(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("returns true for existing customer", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "EXISTS-001", "Exists Test")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		exists, err := customerRepo.Exists(testCtx, customer.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, exists)
	})

	t.Run("returns false for non-existing customer", func(t *testing.T) {
		exists, err := customerRepo.Exists(testCtx, uuid.New())
		helpers.AssertNoError(t, err)
		helpers.AssertFalse(t, exists)
	})
}

func TestCustomerRepository_GetVersion(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("gets correct version", func(t *testing.T) {
		customer := createTestCustomer(tenantID, "VERSION-001", "Version Test")
		err := customerRepo.Create(testCtx, customer)
		helpers.RequireNoError(t, err)

		version, err := customerRepo.GetVersion(testCtx, customer.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 1, version)

		// Update to increment version
		customer.Name = "Updated"
		err = customerRepo.Update(testCtx, customer)
		helpers.RequireNoError(t, err)

		version, err = customerRepo.GetVersion(testCtx, customer.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 2, version)
	})
}

// ============================================================================
// Contact Repository Integration Tests
// ============================================================================

func TestContactRepository_Create(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	// Create customer first
	customer := createTestCustomer(tenantID, "CONTACT-CUST", "Customer for Contacts")
	err := customerRepo.Create(testCtx, customer)
	helpers.RequireNoError(t, err)

	t.Run("successfully creates a new contact", func(t *testing.T) {
		contact := createTestContact(tenantID, customer.ID, "John", "Doe", "john@example.com")

		err := contactRepo.Create(testCtx, contact)
		helpers.AssertNoError(t, err)

		// Verify contact was created
		found, err := contactRepo.FindByID(testCtx, contact.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, "John", found.FirstName)
		helpers.AssertEqual(t, "Doe", found.LastName)
	})
}

func TestContactRepository_FindByCustomer(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	// Create customer
	customer := createTestCustomer(tenantID, "MULTI-CONTACT", "Customer with Multiple Contacts")
	err := customerRepo.Create(testCtx, customer)
	helpers.RequireNoError(t, err)

	t.Run("finds all contacts for customer", func(t *testing.T) {
		// Create multiple contacts
		for i := 0; i < 5; i++ {
			contact := createTestContact(tenantID, customer.ID, helpers.GenerateRandomString(10), helpers.GenerateRandomString(10), helpers.GenerateRandomEmail())
			err := contactRepo.Create(testCtx, contact)
			helpers.RequireNoError(t, err)
		}

		contacts, err := contactRepo.FindByCustomer(testCtx, customer.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertGreaterOrEqual(t, len(contacts), 5)
	})
}

func TestContactRepository_SetPrimary(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	// Create customer
	customer := createTestCustomer(tenantID, "PRIMARY-CONTACT", "Customer for Primary Test")
	err := customerRepo.Create(testCtx, customer)
	helpers.RequireNoError(t, err)

	t.Run("sets primary contact", func(t *testing.T) {
		// Create two contacts
		contact1 := createTestContact(tenantID, customer.ID, "First", "Contact", "first@example.com")
		contact1.IsPrimary = true
		err := contactRepo.Create(testCtx, contact1)
		helpers.RequireNoError(t, err)

		contact2 := createTestContact(tenantID, customer.ID, "Second", "Contact", "second@example.com")
		err = contactRepo.Create(testCtx, contact2)
		helpers.RequireNoError(t, err)

		// Set contact2 as primary
		err = contactRepo.SetPrimary(testCtx, customer.ID, contact2.ID)
		helpers.AssertNoError(t, err)

		// Verify contact2 is now primary
		found, err := contactRepo.FindByID(testCtx, contact2.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, found.IsPrimary)

		// Verify contact1 is no longer primary
		found1, err := contactRepo.FindByID(testCtx, contact1.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertFalse(t, found1.IsPrimary)
	})
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestConcurrentCustomerCreation(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := fixtures.TestIDs.TenantID1

	t.Run("handles concurrent customer creation", func(t *testing.T) {
		const numGoroutines = 10
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				customer := createTestCustomer(tenantID, helpers.GenerateRandomString(15), helpers.GenerateRandomString(20))
				err := customerRepo.Create(testCtx, customer)
				errChan <- err
			}(i)
		}

		successCount := 0
		for i := 0; i < numGoroutines; i++ {
			if err := <-errChan; err == nil {
				successCount++
			}
		}

		helpers.AssertEqual(t, numGoroutines, successCount)
	})
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTestCustomer(tenantID uuid.UUID, code, name string) *domain.Customer {
	return &domain.Customer{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Code:      code,
		Name:      name,
		Type:      domain.CustomerTypeBusiness,
		Status:    domain.CustomerStatusActive,
		Tier:      domain.CustomerTierStandard,
		Source:    "test",
		Tags:      []string{},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Version:   0,
	}
}

func createTestCustomerWithEmail(tenantID uuid.UUID, code, name, email string) *domain.Customer {
	customer := createTestCustomer(tenantID, code, name)
	customer.Email = &domain.Email{
		Address:  email,
		Verified: false,
	}
	return customer
}

func createTestContact(tenantID, customerID uuid.UUID, firstName, lastName, email string) *domain.Contact {
	return &domain.Contact{
		ID:         uuid.New(),
		TenantID:   tenantID,
		CustomerID: customerID,
		FirstName:  firstName,
		LastName:   lastName,
		Email: &domain.Email{
			Address: email,
		},
		IsPrimary: false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func stringSliceToString(slice []string) string {
	result := ""
	for _, s := range slice {
		result += s + ","
	}
	return result
}

// Benchmark tests
func BenchmarkCustomerRepository_Create(b *testing.B) {
	ctx := context.Background()
	tenantID := fixtures.TestIDs.TenantID1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		customer := createTestCustomer(tenantID, helpers.GenerateRandomString(15), helpers.GenerateRandomString(20))
		_ = customerRepo.Create(ctx, customer)
	}
}

func BenchmarkCustomerRepository_FindByID(b *testing.B) {
	ctx := context.Background()
	tenantID := fixtures.TestIDs.TenantID1

	// Create a customer to find
	customer := createTestCustomer(tenantID, helpers.GenerateRandomString(15), helpers.GenerateRandomString(20))
	_ = customerRepo.Create(ctx, customer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = customerRepo.FindByID(ctx, customer.ID)
	}
}

func BenchmarkCustomerRepository_List(b *testing.B) {
	ctx := context.Background()
	tenantID := fixtures.TestIDs.TenantID1

	filter := domain.CustomerFilter{
		TenantID: &tenantID,
		Offset:   0,
		Limit:    20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = customerRepo.List(ctx, filter)
	}
}
