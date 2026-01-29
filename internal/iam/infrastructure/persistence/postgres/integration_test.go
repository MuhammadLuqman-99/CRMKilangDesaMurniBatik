// Package postgres contains PostgreSQL repository integration tests.
package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/kilang-desa-murni/crm/internal/iam/domain"
	"github.com/kilang-desa-murni/crm/pkg/testing/containers"
	"github.com/kilang-desa-murni/crm/pkg/testing/fixtures"
	"github.com/kilang-desa-murni/crm/pkg/testing/helpers"
)

var (
	testDB        *containers.PostgresContainer
	userRepo      *UserRepository
	roleRepo      *RoleRepository
	tenantRepo    *TenantRepository
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

	// Initialize repositories
	userRepo = NewUserRepository(testDB.DB)
	roleRepo = NewRoleRepository(testDB.DB)
	tenantRepo = NewTenantRepository(testDB.DB)

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
	-- Enable UUID extension
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	-- Create tenants table
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

	CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
	CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
	CREATE INDEX IF NOT EXISTS idx_tenants_deleted_at ON tenants(deleted_at);

	-- Create roles table
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

	CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_tenant_name ON roles(tenant_id, name) WHERE tenant_id IS NOT NULL;

	-- Create users table
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

	CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
	CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

	-- Create user_roles junction table
	CREATE TABLE IF NOT EXISTS user_roles (
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
		assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		assigned_by UUID REFERENCES users(id),
		PRIMARY KEY (user_id, role_id)
	);

	CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);

	-- Create refresh_tokens table
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

	CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
	CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
	CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

	-- Create audit_logs table
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

	CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

	-- Create outbox table for transactional outbox pattern
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

	CREATE INDEX IF NOT EXISTS idx_outbox_published ON outbox(published) WHERE NOT published;
	CREATE INDEX IF NOT EXISTS idx_outbox_created_at ON outbox(created_at);
	`

	_, err := db.ExecContext(ctx, migration)
	return err
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
	testDB.TruncateTables(context.Background(), "user_roles", "refresh_tokens", "audit_logs", "users", "roles", "tenants")
}

func createTestTenant(t *testing.T) uuid.UUID {
	t.Helper()
	tenantID := uuid.New()
	_, err := testDB.DB.ExecContext(testCtx, `
		INSERT INTO tenants (id, name, slug, status, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, tenantID, "Test Tenant", "test-tenant-"+tenantID.String()[:8], "active", "free", time.Now(), time.Now())
	helpers.RequireNoError(t, err, "failed to create test tenant")
	return tenantID
}

// ============================================================================
// User Repository Integration Tests
// ============================================================================

func TestUserRepository_Create(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("successfully creates a new user", func(t *testing.T) {
		email, _ := domain.NewEmail("test@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")

		user := domain.NewUser(tenantID, email, password, "John", "Doe")

		err := userRepo.Create(testCtx, user)
		helpers.AssertNoError(t, err)

		// Verify user was created
		found, err := userRepo.FindByID(testCtx, user.GetID())
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, user.GetID(), found.GetID())
		helpers.AssertEqual(t, email.String(), found.Email().String())
		helpers.AssertEqual(t, "John", found.FirstName())
		helpers.AssertEqual(t, "Doe", found.LastName())
	})

	t.Run("fails with duplicate email in same tenant", func(t *testing.T) {
		email, _ := domain.NewEmail("duplicate@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")

		user1 := domain.NewUser(tenantID, email, password, "User", "One")
		err := userRepo.Create(testCtx, user1)
		helpers.AssertNoError(t, err)

		user2 := domain.NewUser(tenantID, email, password, "User", "Two")
		err = userRepo.Create(testCtx, user2)
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrUserAlreadyExists, err)
	})

	t.Run("allows same email in different tenants", func(t *testing.T) {
		tenantID2 := createTestTenant(t)
		email, _ := domain.NewEmail("sametenant@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")

		user1 := domain.NewUser(tenantID, email, password, "User", "One")
		err := userRepo.Create(testCtx, user1)
		helpers.AssertNoError(t, err)

		user2 := domain.NewUser(tenantID2, email, password, "User", "Two")
		err = userRepo.Create(testCtx, user2)
		helpers.AssertNoError(t, err)
	})
}

func TestUserRepository_Update(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("successfully updates user", func(t *testing.T) {
		email, _ := domain.NewEmail("update@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Original", "Name")

		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		// Update user
		user.SetName("Updated", "Name")
		user.UpdatedAt = time.Now().UTC()

		err = userRepo.Update(testCtx, user)
		helpers.AssertNoError(t, err)

		// Verify update
		found, err := userRepo.FindByID(testCtx, user.GetID())
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "Updated", found.FirstName())
		helpers.AssertEqual(t, "Name", found.LastName())
	})

	t.Run("fails to update non-existent user", func(t *testing.T) {
		email, _ := domain.NewEmail("nonexistent@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Non", "Existent")

		err := userRepo.Update(testCtx, user)
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrUserNotFound, err)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("successfully soft deletes user", func(t *testing.T) {
		email, _ := domain.NewEmail("delete@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "To", "Delete")

		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		err = userRepo.Delete(testCtx, user.GetID())
		helpers.AssertNoError(t, err)

		// User should not be found (soft deleted)
		_, err = userRepo.FindByID(testCtx, user.GetID())
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrUserNotFound, err)
	})

	t.Run("fails to delete non-existent user", func(t *testing.T) {
		err := userRepo.Delete(testCtx, uuid.New())
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrUserNotFound, err)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("finds user by email", func(t *testing.T) {
		email, _ := domain.NewEmail("findbyemail@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Find", "Me")

		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		found, err := userRepo.FindByEmail(testCtx, tenantID, email)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, user.GetID(), found.GetID())
	})

	t.Run("case insensitive email search", func(t *testing.T) {
		email, _ := domain.NewEmail("caseinsensitive@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Case", "Test")

		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		upperEmail, _ := domain.NewEmail("CASEINSENSITIVE@example.com")
		found, err := userRepo.FindByEmail(testCtx, tenantID, upperEmail)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		email, _ := domain.NewEmail("nonexistent@example.com")
		_, err := userRepo.FindByEmail(testCtx, tenantID, email)
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrUserNotFound, err)
	})
}

func TestUserRepository_FindByTenant(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("finds users with pagination", func(t *testing.T) {
		// Create multiple users
		for i := 0; i < 25; i++ {
			email, _ := domain.NewEmail(helpers.GenerateRandomEmail())
			password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
			user := domain.NewUser(tenantID, email, password, "User", "Test")
			err := userRepo.Create(testCtx, user)
			helpers.RequireNoError(t, err)
		}

		// Test pagination
		opts := domain.DefaultUserQueryOptions()
		opts.Page = 1
		opts.PageSize = 10

		users, total, err := userRepo.FindByTenant(testCtx, tenantID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(25), total)
		helpers.AssertLen(t, users, 10)

		// Second page
		opts.Page = 2
		users, _, err = userRepo.FindByTenant(testCtx, tenantID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, users, 10)

		// Third page
		opts.Page = 3
		users, _, err = userRepo.FindByTenant(testCtx, tenantID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, users, 5)
	})

	t.Run("filters by status", func(t *testing.T) {
		// Create suspended user
		email, _ := domain.NewEmail(helpers.GenerateRandomEmail())
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Suspended", "User")
		user.Suspend("Test suspension")
		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		opts := domain.DefaultUserQueryOptions()
		suspended := domain.UserStatusSuspended
		opts.Status = &suspended

		users, _, err := userRepo.FindByTenant(testCtx, tenantID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertGreater(t, len(users), 0)
		for _, u := range users {
			helpers.AssertEqual(t, domain.UserStatusSuspended, u.Status())
		}
	})

	t.Run("searches by term", func(t *testing.T) {
		email, _ := domain.NewEmail("searchterm@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Searchable", "User")
		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		opts := domain.DefaultUserQueryOptions()
		opts.Search = "searchable"

		users, _, err := userRepo.FindByTenant(testCtx, tenantID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertGreater(t, len(users), 0)
	})
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("returns true for existing email", func(t *testing.T) {
		email, _ := domain.NewEmail("exists@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Exists", "User")
		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		exists, err := userRepo.ExistsByEmail(testCtx, tenantID, email)
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, exists)
	})

	t.Run("returns false for non-existing email", func(t *testing.T) {
		email, _ := domain.NewEmail("doesnotexist@example.com")
		exists, err := userRepo.ExistsByEmail(testCtx, tenantID, email)
		helpers.AssertNoError(t, err)
		helpers.AssertFalse(t, exists)
	})
}

func TestUserRepository_CountByTenant(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("counts users correctly", func(t *testing.T) {
		// Create 5 users
		for i := 0; i < 5; i++ {
			email, _ := domain.NewEmail(helpers.GenerateRandomEmail())
			password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
			user := domain.NewUser(tenantID, email, password, "Count", "User")
			err := userRepo.Create(testCtx, user)
			helpers.RequireNoError(t, err)
		}

		count, err := userRepo.CountByTenant(testCtx, tenantID)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(5), count)
	})
}

// ============================================================================
// Role Repository Integration Tests
// ============================================================================

func TestRoleRepository_Create(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("successfully creates a new role", func(t *testing.T) {
		role := domain.NewRole(&tenantID, "test_role", "Test Role Description", []string{"users:read", "users:write"})

		err := roleRepo.Create(testCtx, role)
		helpers.AssertNoError(t, err)

		// Verify role was created
		found, err := roleRepo.FindByID(testCtx, role.GetID())
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, role.GetID(), found.GetID())
		helpers.AssertEqual(t, "test_role", found.Name())
	})

	t.Run("creates system role without tenant", func(t *testing.T) {
		role := domain.NewSystemRole("system_test_role", "System Test Role", []string{"*"})

		err := roleRepo.Create(testCtx, role)
		helpers.AssertNoError(t, err)

		found, err := roleRepo.FindByID(testCtx, role.GetID())
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, found.IsSystem())
	})
}

func TestRoleRepository_FindByTenant(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("finds roles for tenant", func(t *testing.T) {
		// Create some roles
		for i := 0; i < 3; i++ {
			role := domain.NewRole(&tenantID, helpers.GenerateRandomString(10), "Description", []string{"read"})
			err := roleRepo.Create(testCtx, role)
			helpers.RequireNoError(t, err)
		}

		opts := domain.DefaultRoleQueryOptions()
		roles, total, err := roleRepo.FindByTenant(testCtx, tenantID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertGreaterOrEqual(t, int(total), 3)
		helpers.AssertNotEmpty(t, roles)
	})
}

func TestRoleRepository_AssignRoleToUser(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("assigns role to user", func(t *testing.T) {
		// Create user
		email, _ := domain.NewEmail(helpers.GenerateRandomEmail())
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Role", "User")
		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		// Create role
		role := domain.NewRole(&tenantID, "user_role_"+helpers.GenerateRandomString(5), "User Role", []string{"read"})
		err = roleRepo.Create(testCtx, role)
		helpers.RequireNoError(t, err)

		// Assign role
		err = roleRepo.AssignRoleToUser(testCtx, user.GetID(), role.GetID(), nil)
		helpers.AssertNoError(t, err)

		// Verify assignment
		roles, err := roleRepo.FindByUserID(testCtx, user.GetID())
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, roles, 1)
		helpers.AssertEqual(t, role.GetID(), roles[0].GetID())
	})

	t.Run("removes role from user", func(t *testing.T) {
		// Create user
		email, _ := domain.NewEmail(helpers.GenerateRandomEmail())
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Remove", "Role")
		err := userRepo.Create(testCtx, user)
		helpers.RequireNoError(t, err)

		// Create and assign role
		role := domain.NewRole(&tenantID, "remove_role_"+helpers.GenerateRandomString(5), "Remove Role", []string{"read"})
		err = roleRepo.Create(testCtx, role)
		helpers.RequireNoError(t, err)

		err = roleRepo.AssignRoleToUser(testCtx, user.GetID(), role.GetID(), nil)
		helpers.RequireNoError(t, err)

		// Remove role
		err = roleRepo.RemoveRoleFromUser(testCtx, user.GetID(), role.GetID())
		helpers.AssertNoError(t, err)

		// Verify removal
		roles, err := roleRepo.FindByUserID(testCtx, user.GetID())
		helpers.AssertNoError(t, err)
		helpers.AssertEmpty(t, roles)
	})
}

// ============================================================================
// Tenant Repository Integration Tests
// ============================================================================

func TestTenantRepository_Create(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully creates a new tenant", func(t *testing.T) {
		tenant := domain.NewTenant("New Tenant", "new-tenant-"+helpers.GenerateRandomString(5))

		err := tenantRepo.Create(testCtx, tenant)
		helpers.AssertNoError(t, err)

		// Verify tenant was created
		found, err := tenantRepo.FindByID(testCtx, tenant.GetID())
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, tenant.GetID(), found.GetID())
	})

	t.Run("fails with duplicate slug", func(t *testing.T) {
		slug := "duplicate-slug-" + helpers.GenerateRandomString(5)
		tenant1 := domain.NewTenant("Tenant 1", slug)
		err := tenantRepo.Create(testCtx, tenant1)
		helpers.RequireNoError(t, err)

		tenant2 := domain.NewTenant("Tenant 2", slug)
		err = tenantRepo.Create(testCtx, tenant2)
		helpers.AssertError(t, err)
	})
}

func TestTenantRepository_FindBySlug(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds tenant by slug", func(t *testing.T) {
		slug := "findable-" + helpers.GenerateRandomString(5)
		tenant := domain.NewTenant("Findable Tenant", slug)
		err := tenantRepo.Create(testCtx, tenant)
		helpers.RequireNoError(t, err)

		found, err := tenantRepo.FindBySlug(testCtx, slug)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, tenant.GetID(), found.GetID())
	})

	t.Run("returns error for non-existent slug", func(t *testing.T) {
		_, err := tenantRepo.FindBySlug(testCtx, "non-existent-slug")
		helpers.AssertError(t, err)
	})
}

func TestTenantRepository_Update(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully updates tenant", func(t *testing.T) {
		tenant := domain.NewTenant("Original Name", "update-test-"+helpers.GenerateRandomString(5))
		err := tenantRepo.Create(testCtx, tenant)
		helpers.RequireNoError(t, err)

		// Update tenant
		tenant.UpdateName("Updated Name")
		tenant.UpdatedAt = time.Now().UTC()

		err = tenantRepo.Update(testCtx, tenant)
		helpers.AssertNoError(t, err)

		// Verify update
		found, err := tenantRepo.FindByID(testCtx, tenant.GetID())
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "Updated Name", found.Name())
	})
}

func TestTenantRepository_ExistsBySlug(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("returns true for existing slug", func(t *testing.T) {
		slug := "exists-" + helpers.GenerateRandomString(5)
		tenant := domain.NewTenant("Exists Tenant", slug)
		err := tenantRepo.Create(testCtx, tenant)
		helpers.RequireNoError(t, err)

		exists, err := tenantRepo.ExistsBySlug(testCtx, slug)
		helpers.AssertNoError(t, err)
		helpers.AssertTrue(t, exists)
	})

	t.Run("returns false for non-existing slug", func(t *testing.T) {
		exists, err := tenantRepo.ExistsBySlug(testCtx, "does-not-exist-"+helpers.GenerateRandomString(5))
		helpers.AssertNoError(t, err)
		helpers.AssertFalse(t, exists)
	})
}

// ============================================================================
// Transaction Tests
// ============================================================================

func TestTransactionRollback(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("rollback on error", func(t *testing.T) {
		email, _ := domain.NewEmail("rollback@example.com")
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Rollback", "Test")

		// Start transaction
		tx, err := testDB.DB.BeginTx(testCtx, nil)
		helpers.RequireNoError(t, err)

		// Create context with transaction
		txCtx := context.WithValue(testCtx, txContextKey{}, tx)

		// Create user in transaction
		txUserRepo := NewUserRepository(testDB.DB)
		err = txUserRepo.Create(txCtx, user)
		helpers.AssertNoError(t, err)

		// Rollback
		err = tx.Rollback()
		helpers.AssertNoError(t, err)

		// User should not exist
		_, err = userRepo.FindByID(testCtx, user.GetID())
		helpers.AssertError(t, err)
		helpers.AssertEqual(t, domain.ErrUserNotFound, err)
	})
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestConcurrentUserCreation(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	tenantID := createTestTenant(t)

	t.Run("handles concurrent user creation", func(t *testing.T) {
		const numGoroutines = 10
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				email, _ := domain.NewEmail(helpers.GenerateRandomEmail())
				password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
				user := domain.NewUser(tenantID, email, password, "Concurrent", "User")
				err := userRepo.Create(testCtx, user)
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

// Benchmark tests
func BenchmarkUserRepository_Create(b *testing.B) {
	ctx := context.Background()
	tenantID := fixtures.TestIDs.TenantID1

	// Ensure tenant exists
	_, _ = testDB.DB.ExecContext(ctx, `
		INSERT INTO tenants (id, name, slug, status, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, tenantID, "Benchmark Tenant", "benchmark-tenant", "active", "free", time.Now(), time.Now())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email, _ := domain.NewEmail(helpers.GenerateRandomEmail())
		password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
		user := domain.NewUser(tenantID, email, password, "Bench", "User")
		_ = userRepo.Create(ctx, user)
	}
}

func BenchmarkUserRepository_FindByID(b *testing.B) {
	ctx := context.Background()
	tenantID := fixtures.TestIDs.TenantID1

	// Create a user to find
	email, _ := domain.NewEmail(helpers.GenerateRandomEmail())
	password := domain.NewPasswordFromHash("$argon2id$v=19$m=65536,t=3,p=4$test$hash")
	user := domain.NewUser(tenantID, email, password, "Find", "User")
	_ = userRepo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = userRepo.FindByID(ctx, user.GetID())
	}
}
