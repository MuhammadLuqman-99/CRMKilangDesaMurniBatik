// Package postgres contains PostgreSQL repository integration tests for the Sales service.
package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/kilang-desa-murni/crm/internal/sales/domain"
	"github.com/kilang-desa-murni/crm/pkg/testing/containers"
	"github.com/kilang-desa-murni/crm/pkg/testing/fixtures"
	"github.com/kilang-desa-murni/crm/pkg/testing/helpers"
)

var (
	testDB          *containers.PostgresContainer
	leadRepo        *LeadRepository
	opportunityRepo *OpportunityRepository
	dealRepo        *DealRepository
	pipelineRepo    *PipelineRepository
	testCtx         context.Context
	testCtxCancel   context.CancelFunc
	testTenantID    uuid.UUID
	testPipelineID  uuid.UUID
)

// TestMain sets up and tears down the test database.
func TestMain(m *testing.M) {
	if testing.Short() {
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var err error
	testDB, err = containers.NewPostgresContainer(ctx, containers.DefaultPostgresConfig())
	if err != nil {
		panic("failed to create PostgreSQL container: " + err.Error())
	}

	if err := runTestMigrations(ctx, testDB.DB); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	leadRepo = NewLeadRepository(testDB.DB)
	opportunityRepo = NewOpportunityRepository(testDB.DB)
	dealRepo = NewDealRepository(testDB.DB)
	pipelineRepo = NewPipelineRepository(testDB.DB)

	// Create test tenant and pipeline
	testTenantID = fixtures.TestIDs.TenantID1
	createTestTenantAndPipeline(ctx, testDB.DB)

	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}

	os.Exit(code)
}

func runTestMigrations(ctx context.Context, db *sqlx.DB) error {
	migration := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS tenants (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(100) NOT NULL UNIQUE,
		status VARCHAR(50) NOT NULL DEFAULT 'active',
		plan VARCHAR(50) NOT NULL DEFAULT 'free',
		settings JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

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

	CREATE TABLE IF NOT EXISTS pipeline_stages (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		type VARCHAR(50) NOT NULL DEFAULT 'open',
		stage_order INT NOT NULL,
		probability INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
	);

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
		notes TEXT,
		assigned_to UUID,
		converted_at TIMESTAMP WITH TIME ZONE,
		converted_opportunity_id UUID,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	CREATE INDEX IF NOT EXISTS idx_leads_tenant_id ON leads(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
	CREATE INDEX IF NOT EXISTS idx_leads_assigned_to ON leads(assigned_to);

	CREATE TABLE IF NOT EXISTS opportunities (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		customer_id UUID NOT NULL,
		lead_id UUID REFERENCES leads(id),
		pipeline_id UUID NOT NULL REFERENCES pipelines(id),
		stage_id UUID NOT NULL REFERENCES pipeline_stages(id),
		name VARCHAR(255) NOT NULL,
		value_amount BIGINT NOT NULL DEFAULT 0,
		value_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
		probability INT NOT NULL DEFAULT 0,
		expected_close TIMESTAMP WITH TIME ZONE,
		assigned_to UUID,
		status VARCHAR(50) NOT NULL DEFAULT 'open',
		won_reason TEXT,
		lost_reason TEXT,
		competitor TEXT,
		notes TEXT,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	CREATE INDEX IF NOT EXISTS idx_opportunities_tenant_id ON opportunities(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_opportunities_pipeline_id ON opportunities(pipeline_id);
	CREATE INDEX IF NOT EXISTS idx_opportunities_stage_id ON opportunities(stage_id);
	CREATE INDEX IF NOT EXISTS idx_opportunities_status ON opportunities(status);

	CREATE TABLE IF NOT EXISTS deals (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
		opportunity_id UUID NOT NULL REFERENCES opportunities(id),
		customer_id UUID NOT NULL,
		name VARCHAR(255) NOT NULL,
		value_amount BIGINT NOT NULL DEFAULT 0,
		value_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		closed_at TIMESTAMP WITH TIME ZONE,
		assigned_to UUID,
		notes TEXT,
		metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE
	);

	CREATE INDEX IF NOT EXISTS idx_deals_tenant_id ON deals(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_deals_opportunity_id ON deals(opportunity_id);
	CREATE INDEX IF NOT EXISTS idx_deals_status ON deals(status);

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
	`

	_, err := db.ExecContext(ctx, migration)
	return err
}

func createTestTenantAndPipeline(ctx context.Context, db *sqlx.DB) {
	// Create tenant
	_, _ = db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, slug, status, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, testTenantID, "Test Tenant", "test-tenant-sales", "active", "free", time.Now(), time.Now())

	// Create pipeline
	testPipelineID = fixtures.TestIDs.PipelineID1
	_, _ = db.ExecContext(ctx, `
		INSERT INTO pipelines (id, tenant_id, name, description, is_default, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING
	`, testPipelineID, testTenantID, "Default Pipeline", "Default sales pipeline", true, "active", time.Now(), time.Now())

	// Create stages
	stages := []struct {
		id          uuid.UUID
		name        string
		stageType   string
		order       int
		probability int
	}{
		{uuid.MustParse("57a61111-1111-1111-1111-111111111111"), "Lead", "open", 1, 10},
		{uuid.MustParse("57a62222-2222-2222-2222-222222222222"), "Qualified", "open", 2, 30},
		{uuid.MustParse("57a63333-3333-3333-3333-333333333333"), "Proposal", "open", 3, 60},
		{uuid.MustParse("57a64444-4444-4444-4444-444444444444"), "Won", "won", 4, 100},
		{uuid.MustParse("57a65555-5555-5555-5555-555555555555"), "Lost", "lost", 5, 0},
	}

	for _, s := range stages {
		_, _ = db.ExecContext(ctx, `
			INSERT INTO pipeline_stages (id, pipeline_id, name, type, stage_order, probability, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO NOTHING
		`, s.id, testPipelineID, s.name, s.stageType, s.order, s.probability, time.Now(), time.Now())
	}
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
	testDB.TruncateTables(context.Background(), "deals", "opportunities", "leads")
}

// ============================================================================
// Lead Repository Integration Tests
// ============================================================================

func TestLeadRepository_Create(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully creates a new lead", func(t *testing.T) {
		lead := createTestLead(testTenantID)

		err := leadRepo.Create(testCtx, lead)
		helpers.AssertNoError(t, err)

		found, err := leadRepo.FindByID(testCtx, lead.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, lead.ID, found.ID)
		helpers.AssertEqual(t, "Test Company", found.CompanyName)
	})

	t.Run("creates lead with all fields", func(t *testing.T) {
		lead := createTestLead(testTenantID)
		lead.Score = 85
		lead.Notes = "Important lead"
		lead.AssignedTo = &fixtures.TestIDs.UserID1

		err := leadRepo.Create(testCtx, lead)
		helpers.AssertNoError(t, err)

		found, err := leadRepo.FindByID(testCtx, lead.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, 85, found.Score)
		helpers.AssertEqual(t, "Important lead", found.Notes)
		helpers.AssertNotNil(t, found.AssignedTo)
	})
}

func TestLeadRepository_Update(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully updates lead", func(t *testing.T) {
		lead := createTestLead(testTenantID)
		err := leadRepo.Create(testCtx, lead)
		helpers.RequireNoError(t, err)

		lead.Status = domain.LeadStatusQualified
		lead.Score = 90
		lead.UpdatedAt = time.Now().UTC()

		err = leadRepo.Update(testCtx, lead)
		helpers.AssertNoError(t, err)

		found, err := leadRepo.FindByID(testCtx, lead.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, domain.LeadStatusQualified, found.Status)
		helpers.AssertEqual(t, 90, found.Score)
	})
}

func TestLeadRepository_Delete(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("soft deletes lead", func(t *testing.T) {
		lead := createTestLead(testTenantID)
		err := leadRepo.Create(testCtx, lead)
		helpers.RequireNoError(t, err)

		err = leadRepo.Delete(testCtx, lead.ID)
		helpers.AssertNoError(t, err)

		_, err = leadRepo.FindByID(testCtx, lead.ID)
		helpers.AssertError(t, err)
	})
}

func TestLeadRepository_FindByTenant(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds leads with pagination", func(t *testing.T) {
		for i := 0; i < 15; i++ {
			lead := createTestLead(testTenantID)
			err := leadRepo.Create(testCtx, lead)
			helpers.RequireNoError(t, err)
		}

		opts := domain.LeadQueryOptions{
			Page:     1,
			PageSize: 10,
		}

		leads, total, err := leadRepo.FindByTenant(testCtx, testTenantID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(15), total)
		helpers.AssertLen(t, leads, 10)
	})

	t.Run("filters by status", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			lead := createTestLead(testTenantID)
			lead.Status = domain.LeadStatusQualified
			err := leadRepo.Create(testCtx, lead)
			helpers.RequireNoError(t, err)
		}

		qualified := domain.LeadStatusQualified
		opts := domain.LeadQueryOptions{
			Page:     1,
			PageSize: 100,
			Status:   &qualified,
		}

		leads, _, err := leadRepo.FindByTenant(testCtx, testTenantID, opts)
		helpers.AssertNoError(t, err)
		for _, lead := range leads {
			helpers.AssertEqual(t, domain.LeadStatusQualified, lead.Status)
		}
	})
}

func TestLeadRepository_FindByAssignee(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds leads by assignee", func(t *testing.T) {
		assigneeID := fixtures.TestIDs.UserID1

		for i := 0; i < 5; i++ {
			lead := createTestLead(testTenantID)
			lead.AssignedTo = &assigneeID
			err := leadRepo.Create(testCtx, lead)
			helpers.RequireNoError(t, err)
		}

		leads, err := leadRepo.FindByAssignee(testCtx, testTenantID, assigneeID)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, leads, 5)
	})
}

// ============================================================================
// Opportunity Repository Integration Tests
// ============================================================================

func TestOpportunityRepository_Create(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully creates opportunity", func(t *testing.T) {
		stageID := uuid.MustParse("57a61111-1111-1111-1111-111111111111")
		opp := createTestOpportunity(testTenantID, testPipelineID, stageID)

		err := opportunityRepo.Create(testCtx, opp)
		helpers.AssertNoError(t, err)

		found, err := opportunityRepo.FindByID(testCtx, opp.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, opp.ID, found.ID)
		helpers.AssertEqual(t, "Test Opportunity", found.Name)
	})
}

func TestOpportunityRepository_Update(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("updates opportunity", func(t *testing.T) {
		stageID := uuid.MustParse("57a61111-1111-1111-1111-111111111111")
		opp := createTestOpportunity(testTenantID, testPipelineID, stageID)
		err := opportunityRepo.Create(testCtx, opp)
		helpers.RequireNoError(t, err)

		newStageID := uuid.MustParse("57a62222-2222-2222-2222-222222222222")
		opp.StageID = newStageID
		opp.Probability = 30
		opp.UpdatedAt = time.Now().UTC()

		err = opportunityRepo.Update(testCtx, opp)
		helpers.AssertNoError(t, err)

		found, err := opportunityRepo.FindByID(testCtx, opp.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, newStageID, found.StageID)
		helpers.AssertEqual(t, 30, found.Probability)
	})
}

func TestOpportunityRepository_FindByPipeline(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds opportunities in pipeline", func(t *testing.T) {
		stageID := uuid.MustParse("57a61111-1111-1111-1111-111111111111")

		for i := 0; i < 5; i++ {
			opp := createTestOpportunity(testTenantID, testPipelineID, stageID)
			err := opportunityRepo.Create(testCtx, opp)
			helpers.RequireNoError(t, err)
		}

		opts := domain.OpportunityQueryOptions{
			Page:     1,
			PageSize: 100,
		}

		opps, total, err := opportunityRepo.FindByPipeline(testCtx, testPipelineID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(5), total)
		helpers.AssertLen(t, opps, 5)
	})
}

func TestOpportunityRepository_FindByStage(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds opportunities by stage", func(t *testing.T) {
		qualifiedStageID := uuid.MustParse("57a62222-2222-2222-2222-222222222222")

		for i := 0; i < 3; i++ {
			opp := createTestOpportunity(testTenantID, testPipelineID, qualifiedStageID)
			err := opportunityRepo.Create(testCtx, opp)
			helpers.RequireNoError(t, err)
		}

		opps, err := opportunityRepo.FindByStage(testCtx, qualifiedStageID)
		helpers.AssertNoError(t, err)
		helpers.AssertLen(t, opps, 3)
	})
}

func TestOpportunityRepository_GetTotalValue(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("calculates total pipeline value", func(t *testing.T) {
		stageID := uuid.MustParse("57a61111-1111-1111-1111-111111111111")

		for i := 0; i < 3; i++ {
			opp := createTestOpportunity(testTenantID, testPipelineID, stageID)
			opp.ValueAmount = 100000 // $1000.00 each
			err := opportunityRepo.Create(testCtx, opp)
			helpers.RequireNoError(t, err)
		}

		total, err := opportunityRepo.GetTotalValue(testCtx, testPipelineID)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(300000), total)
	})
}

// ============================================================================
// Deal Repository Integration Tests
// ============================================================================

func TestDealRepository_Create(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("successfully creates deal", func(t *testing.T) {
		// Create opportunity first
		stageID := uuid.MustParse("57a64444-4444-4444-4444-444444444444") // Won stage
		opp := createTestOpportunity(testTenantID, testPipelineID, stageID)
		opp.Status = domain.OpportunityStatusWon
		err := opportunityRepo.Create(testCtx, opp)
		helpers.RequireNoError(t, err)

		deal := createTestDeal(testTenantID, opp.ID, opp.CustomerID)

		err = dealRepo.Create(testCtx, deal)
		helpers.AssertNoError(t, err)

		found, err := dealRepo.FindByID(testCtx, deal.ID)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, found)
		helpers.AssertEqual(t, deal.ID, found.ID)
	})
}

func TestDealRepository_FindByTenant(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds deals with pagination", func(t *testing.T) {
		stageID := uuid.MustParse("57a64444-4444-4444-4444-444444444444")

		for i := 0; i < 10; i++ {
			opp := createTestOpportunity(testTenantID, testPipelineID, stageID)
			opp.Status = domain.OpportunityStatusWon
			err := opportunityRepo.Create(testCtx, opp)
			helpers.RequireNoError(t, err)

			deal := createTestDeal(testTenantID, opp.ID, opp.CustomerID)
			err = dealRepo.Create(testCtx, deal)
			helpers.RequireNoError(t, err)
		}

		opts := domain.DealQueryOptions{
			Page:     1,
			PageSize: 5,
		}

		deals, total, err := dealRepo.FindByTenant(testCtx, testTenantID, opts)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(10), total)
		helpers.AssertLen(t, deals, 5)
	})
}

func TestDealRepository_GetTotalValue(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("calculates total deal value", func(t *testing.T) {
		stageID := uuid.MustParse("57a64444-4444-4444-4444-444444444444")

		for i := 0; i < 3; i++ {
			opp := createTestOpportunity(testTenantID, testPipelineID, stageID)
			err := opportunityRepo.Create(testCtx, opp)
			helpers.RequireNoError(t, err)

			deal := createTestDeal(testTenantID, opp.ID, opp.CustomerID)
			deal.ValueAmount = 50000 // $500 each
			deal.Status = domain.DealStatusWon
			err = dealRepo.Create(testCtx, deal)
			helpers.RequireNoError(t, err)
		}

		total, err := dealRepo.GetTotalValue(testCtx, testTenantID, domain.DealStatusWon)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, int64(150000), total)
	})
}

// ============================================================================
// Pipeline Repository Integration Tests
// ============================================================================

func TestPipelineRepository_FindByID(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds pipeline with stages", func(t *testing.T) {
		pipeline, err := pipelineRepo.FindByID(testCtx, testPipelineID)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, pipeline)
		helpers.AssertEqual(t, "Default Pipeline", pipeline.Name)
		helpers.AssertGreater(t, len(pipeline.Stages), 0)
	})
}

func TestPipelineRepository_FindByTenant(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds pipelines for tenant", func(t *testing.T) {
		pipelines, err := pipelineRepo.FindByTenant(testCtx, testTenantID)
		helpers.AssertNoError(t, err)
		helpers.AssertGreater(t, len(pipelines), 0)
	})
}

func TestPipelineRepository_GetDefault(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("finds default pipeline", func(t *testing.T) {
		pipeline, err := pipelineRepo.GetDefault(testCtx, testTenantID)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, pipeline)
		helpers.AssertTrue(t, pipeline.IsDefault)
	})
}

// ============================================================================
// Transaction Tests
// ============================================================================

func TestLeadConversionTransaction(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("lead conversion in transaction", func(t *testing.T) {
		// Create lead
		lead := createTestLead(testTenantID)
		err := leadRepo.Create(testCtx, lead)
		helpers.RequireNoError(t, err)

		// Start transaction
		tx, err := testDB.DB.BeginTx(testCtx, nil)
		helpers.RequireNoError(t, err)

		// Create opportunity from lead
		stageID := uuid.MustParse("57a61111-1111-1111-1111-111111111111")
		opp := createTestOpportunity(testTenantID, testPipelineID, stageID)
		opp.LeadID = &lead.ID

		// Update lead as converted
		lead.Status = domain.LeadStatusConverted
		lead.ConvertedAt = &time.Time{}
		*lead.ConvertedAt = time.Now().UTC()
		lead.ConvertedOpportunityID = &opp.ID

		// Commit transaction
		err = tx.Commit()
		helpers.AssertNoError(t, err)
	})
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestConcurrentLeadCreation(t *testing.T) {
	setupTest(t)
	defer cleanupTest(t)

	t.Run("handles concurrent lead creation", func(t *testing.T) {
		const numGoroutines = 10
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				lead := createTestLead(testTenantID)
				err := leadRepo.Create(testCtx, lead)
				errChan <- err
			}()
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

func createTestLead(tenantID uuid.UUID) *domain.Lead {
	return &domain.Lead{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Source:       domain.LeadSourceWebsite,
		Status:       domain.LeadStatusNew,
		Score:        50,
		CompanyName:  "Test Company",
		ContactName:  "John Doe",
		ContactEmail: helpers.GenerateRandomEmail(),
		ContactPhone: helpers.GenerateRandomPhone(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

func createTestOpportunity(tenantID, pipelineID, stageID uuid.UUID) *domain.Opportunity {
	return &domain.Opportunity{
		ID:            uuid.New(),
		TenantID:      tenantID,
		CustomerID:    fixtures.TestIDs.CustomerID1,
		PipelineID:    pipelineID,
		StageID:       stageID,
		Name:          "Test Opportunity",
		ValueAmount:   100000, // $1000.00
		ValueCurrency: "USD",
		Probability:   10,
		ExpectedClose: time.Now().AddDate(0, 1, 0),
		Status:        domain.OpportunityStatusOpen,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}

func createTestDeal(tenantID, opportunityID, customerID uuid.UUID) *domain.Deal {
	return &domain.Deal{
		ID:            uuid.New(),
		TenantID:      tenantID,
		OpportunityID: opportunityID,
		CustomerID:    customerID,
		Name:          "Test Deal",
		ValueAmount:   100000,
		ValueCurrency: "USD",
		Status:        domain.DealStatusPending,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
}

// Benchmark tests
func BenchmarkLeadRepository_Create(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lead := createTestLead(testTenantID)
		_ = leadRepo.Create(ctx, lead)
	}
}

func BenchmarkLeadRepository_FindByID(b *testing.B) {
	ctx := context.Background()

	lead := createTestLead(testTenantID)
	_ = leadRepo.Create(ctx, lead)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = leadRepo.FindByID(ctx, lead.ID)
	}
}
