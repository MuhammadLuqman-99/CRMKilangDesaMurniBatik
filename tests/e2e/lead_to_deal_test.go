// Package e2e contains E2E tests for the complete lead-to-deal flow.
package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLeadToDealFlow tests the complete sales pipeline from lead to closed deal.
func TestLeadToDealFlow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup: Create tenant, user, and login
	suite.CreateTenant("Sales Flow Test Company", "sales-flow-test")
	suite.RegisterUser("sales.rep@test.com", "Password123!", "Sales", "Rep")
	suite.Login("sales.rep@test.com", "Password123!")

	t.Run("Complete Lead to Deal Flow", func(t *testing.T) {
		// Step 1: Create a sales pipeline
		pipelineResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name":       "Standard Sales Pipeline",
			"is_default": true,
		})
		suite.AssertStatus(pipelineResp, http.StatusCreated)

		var pipeline PipelineResponse
		suite.DecodeResponse(pipelineResp, &pipeline)
		require.NotEmpty(t, pipeline.ID, "pipeline ID should not be empty")
		assert.Equal(t, "Standard Sales Pipeline", pipeline.Name)
		assert.True(t, pipeline.IsDefault)
		assert.GreaterOrEqual(t, len(pipeline.Stages), 4) // Default stages should be created

		// Get first stage for lead conversion
		firstStageID := ""
		if len(pipeline.Stages) > 0 {
			firstStageID = pipeline.Stages[0].ID
		}

		// Step 2: Create a lead
		lead := suite.CreateLead("Potential Corp", "Alice Johnson", "alice@potentialcorp.com")
		require.NotEmpty(t, lead.ID, "lead ID should not be empty")
		assert.Equal(t, "Potential Corp", lead.CompanyName)
		assert.Equal(t, "new", lead.Status)
		assert.Equal(t, "website", lead.Source)
		assert.Equal(t, 50, lead.Score)

		// Step 3: Update lead score and information
		updateResp := suite.DoRequestAuth("PUT", salesServer.URL+"/api/v1/leads/"+lead.ID, map[string]interface{}{
			"score":        75,
			"company_name": "Potential Corporation",
		})
		suite.AssertStatus(updateResp, http.StatusOK)

		var updatedLead struct {
			Score       int    `json:"score"`
			CompanyName string `json:"company_name"`
		}
		suite.DecodeResponse(updateResp, &updatedLead)
		assert.Equal(t, 75, updatedLead.Score)
		assert.Equal(t, "Potential Corporation", updatedLead.CompanyName)

		// Step 4: Qualify the lead
		qualifyResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/leads/"+lead.ID+"/qualify", nil)
		suite.AssertStatus(qualifyResp, http.StatusOK)

		var qualifiedLead struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(qualifyResp, &qualifiedLead)
		assert.Equal(t, "qualified", qualifiedLead.Status)

		// Step 5: Convert lead to opportunity (creates customer automatically)
		convertResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/leads/"+lead.ID+"/convert", map[string]interface{}{
			"pipeline_id":    pipeline.ID,
			"stage_id":       firstStageID,
			"value_amount":   50000,
			"value_currency": "USD",
		})
		suite.AssertStatus(convertResp, http.StatusOK)

		var conversionResult struct {
			Lead struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"lead"`
			Customer struct {
				ID string `json:"id"`
			} `json:"customer"`
			Opportunity struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				PipelineID    string `json:"pipeline_id"`
				StageID       string `json:"stage_id"`
				ValueAmount   int64  `json:"value_amount"`
				ValueCurrency string `json:"value_currency"`
			} `json:"opportunity"`
		}
		suite.DecodeResponse(convertResp, &conversionResult)
		assert.Equal(t, "converted", conversionResult.Lead.Status)
		assert.NotEmpty(t, conversionResult.Customer.ID)
		assert.NotEmpty(t, conversionResult.Opportunity.ID)
		assert.Equal(t, int64(50000), conversionResult.Opportunity.ValueAmount)

		opportunityID := conversionResult.Opportunity.ID

		// Step 6: Move opportunity through pipeline stages
		// Find negotiation stage
		var negotiationStageID string
		for _, stage := range pipeline.Stages {
			if stage.Name == "Negotiation" {
				negotiationStageID = stage.ID
				break
			}
		}

		if negotiationStageID != "" {
			moveResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities/"+opportunityID+"/move-stage", map[string]interface{}{
				"stage_id": negotiationStageID,
			})
			suite.AssertStatus(moveResp, http.StatusOK)

			var movedOpp struct {
				StageID     string `json:"stage_id"`
				Probability int    `json:"probability"`
			}
			suite.DecodeResponse(moveResp, &movedOpp)
			assert.Equal(t, negotiationStageID, movedOpp.StageID)
		}

		// Step 7: Update opportunity value
		updateOppResp := suite.DoRequestAuth("PUT", salesServer.URL+"/api/v1/opportunities/"+opportunityID, map[string]interface{}{
			"value_amount": 75000,
			"name":         "Potential Corporation - Enterprise Deal",
		})
		suite.AssertStatus(updateOppResp, http.StatusOK)

		var updatedOpp struct {
			Name        string `json:"name"`
			ValueAmount int64  `json:"value_amount"`
		}
		suite.DecodeResponse(updateOppResp, &updatedOpp)
		assert.Equal(t, int64(75000), updatedOpp.ValueAmount)

		// Step 8: Win the opportunity (creates deal)
		winResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities/"+opportunityID+"/win", map[string]interface{}{
			"reason": "Client accepted proposal with minor modifications",
		})
		suite.AssertStatus(winResp, http.StatusOK)

		var winResult struct {
			Opportunity struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"opportunity"`
			Deal struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				ValueAmount   int64  `json:"value_amount"`
				ValueCurrency string `json:"value_currency"`
				Status        string `json:"status"`
			} `json:"deal"`
		}
		suite.DecodeResponse(winResp, &winResult)
		assert.Equal(t, "won", winResult.Opportunity.Status)
		assert.NotEmpty(t, winResult.Deal.ID)
		assert.Equal(t, "pending", winResult.Deal.Status)

		dealID := winResult.Deal.ID

		// Step 9: Activate the deal
		activateResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/deals/"+dealID+"/activate", nil)
		suite.AssertStatus(activateResp, http.StatusOK)

		var activatedDeal struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(activateResp, &activatedDeal)
		assert.Equal(t, "active", activatedDeal.Status)

		// Step 10: Complete the deal
		completeResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/deals/"+dealID+"/complete", nil)
		suite.AssertStatus(completeResp, http.StatusOK)

		var completedDeal struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(completeResp, &completedDeal)
		assert.Equal(t, "completed", completedDeal.Status)
	})
}

// TestLeadManagement tests lead CRUD operations.
func TestLeadManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Lead Management Test", "lead-mgmt-test")
	suite.RegisterUser("lead.manager@test.com", "Password123!", "Lead", "Manager")
	suite.Login("lead.manager@test.com", "Password123!")

	t.Run("Create Lead", func(t *testing.T) {
		lead := suite.CreateLead("New Company", "John Doe", "john@newcompany.com")
		assert.NotEmpty(t, lead.ID)
		assert.Equal(t, "New Company", lead.CompanyName)
		assert.Equal(t, "John Doe", lead.ContactName)
		assert.Equal(t, "new", lead.Status)
	})

	t.Run("List Leads", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", salesServer.URL+"/api/v1/leads", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		suite.DecodeResponse(resp, &listResp)
		assert.GreaterOrEqual(t, listResp.Total, 1)
	})

	t.Run("Update Lead", func(t *testing.T) {
		lead := suite.CreateLead("Update Test", "Jane Doe", "jane@updatetest.com")

		resp := suite.DoRequestAuth("PUT", salesServer.URL+"/api/v1/leads/"+lead.ID, map[string]interface{}{
			"company_name":  "Updated Company",
			"contact_email": "updated@company.com",
			"score":         80,
		})
		suite.AssertStatus(resp, http.StatusOK)

		var updatedLead struct {
			CompanyName string `json:"company_name"`
			Score       int    `json:"score"`
		}
		suite.DecodeResponse(resp, &updatedLead)
		assert.Equal(t, "Updated Company", updatedLead.CompanyName)
		assert.Equal(t, 80, updatedLead.Score)
	})

	t.Run("Assign Lead", func(t *testing.T) {
		lead := suite.CreateLead("Assign Test", "Bob Smith", "bob@assigntest.com")

		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/leads/"+lead.ID+"/assign", map[string]interface{}{
			"user_id": suite.userID.String(),
		})
		suite.AssertStatus(resp, http.StatusOK)

		var assignResp struct {
			AssignedTo string `json:"assigned_to"`
		}
		suite.DecodeResponse(resp, &assignResp)
		assert.Equal(t, suite.userID.String(), assignResp.AssignedTo)
	})

	t.Run("Disqualify Lead", func(t *testing.T) {
		lead := suite.CreateLead("Disqualify Test", "Not Interested", "not@interested.com")

		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/leads/"+lead.ID+"/disqualify", map[string]interface{}{
			"reason": "No budget available",
		})
		suite.AssertStatus(resp, http.StatusOK)

		var disqualifyResp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &disqualifyResp)
		assert.Equal(t, "disqualified", disqualifyResp.Status)
	})

	t.Run("Delete Lead", func(t *testing.T) {
		lead := suite.CreateLead("Delete Test", "Delete Me", "delete@test.com")

		resp := suite.DoRequestAuth("DELETE", salesServer.URL+"/api/v1/leads/"+lead.ID, nil)
		suite.AssertStatus(resp, http.StatusNoContent)

		// Verify lead is deleted
		resp = suite.DoRequestAuth("GET", salesServer.URL+"/api/v1/leads/"+lead.ID, nil)
		suite.AssertStatus(resp, http.StatusNotFound)
	})
}

// TestOpportunityManagement tests opportunity operations.
func TestOpportunityManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Opportunity Test", "opp-test")
	suite.RegisterUser("opp.manager@test.com", "Password123!", "Opp", "Manager")
	suite.Login("opp.manager@test.com", "Password123!")

	// Create prerequisite data
	customer := suite.CreateCustomer("Opp Test Corp", "OPP-001", "opp@test.com")
	pipelineResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
		"name":       "Test Pipeline",
		"is_default": true,
	})
	suite.AssertStatus(pipelineResp, http.StatusCreated)
	var pipeline PipelineResponse
	suite.DecodeResponse(pipelineResp, &pipeline)

	stageID := ""
	if len(pipeline.Stages) > 0 {
		stageID = pipeline.Stages[0].ID
	}

	t.Run("Create Opportunity", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities", map[string]interface{}{
			"customer_id":    customer.ID,
			"pipeline_id":    pipeline.ID,
			"stage_id":       stageID,
			"name":           "New Opportunity",
			"value_amount":   25000,
			"value_currency": "USD",
			"probability":    25,
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var opp OpportunityResponse
		suite.DecodeResponse(resp, &opp)
		assert.NotEmpty(t, opp.ID)
		assert.Equal(t, "New Opportunity", opp.Name)
		assert.Equal(t, int64(25000), opp.ValueAmount)
		assert.Equal(t, "open", opp.Status)
	})

	t.Run("List Opportunities", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", salesServer.URL+"/api/v1/opportunities", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		suite.DecodeResponse(resp, &listResp)
		assert.GreaterOrEqual(t, listResp.Total, 1)
	})

	t.Run("Move Stage", func(t *testing.T) {
		// Create an opportunity
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities", map[string]interface{}{
			"customer_id":    customer.ID,
			"pipeline_id":    pipeline.ID,
			"stage_id":       stageID,
			"name":           "Stage Move Test",
			"value_amount":   10000,
			"value_currency": "USD",
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var opp OpportunityResponse
		suite.DecodeResponse(resp, &opp)

		// Move to next stage if available
		if len(pipeline.Stages) > 1 {
			nextStageID := pipeline.Stages[1].ID
			resp = suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities/"+opp.ID+"/move-stage", map[string]interface{}{
				"stage_id": nextStageID,
			})
			suite.AssertStatus(resp, http.StatusOK)

			var movedOpp struct {
				StageID string `json:"stage_id"`
			}
			suite.DecodeResponse(resp, &movedOpp)
			assert.Equal(t, nextStageID, movedOpp.StageID)
		}
	})

	t.Run("Lose Opportunity", func(t *testing.T) {
		// Create an opportunity
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities", map[string]interface{}{
			"customer_id":    customer.ID,
			"pipeline_id":    pipeline.ID,
			"stage_id":       stageID,
			"name":           "Lost Opportunity",
			"value_amount":   15000,
			"value_currency": "USD",
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var opp OpportunityResponse
		suite.DecodeResponse(resp, &opp)

		// Lose the opportunity
		resp = suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities/"+opp.ID+"/lose", map[string]interface{}{
			"reason": "Competitor offered better price",
		})
		suite.AssertStatus(resp, http.StatusOK)

		var lostOpp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &lostOpp)
		assert.Equal(t, "lost", lostOpp.Status)
	})

	t.Run("Reopen Opportunity", func(t *testing.T) {
		// Create and lose an opportunity
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities", map[string]interface{}{
			"customer_id":    customer.ID,
			"pipeline_id":    pipeline.ID,
			"stage_id":       stageID,
			"name":           "Reopen Test",
			"value_amount":   20000,
			"value_currency": "USD",
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var opp OpportunityResponse
		suite.DecodeResponse(resp, &opp)

		// Lose it
		suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities/"+opp.ID+"/lose", map[string]interface{}{
			"reason": "Temporary close",
		})

		// Reopen it
		resp = suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities/"+opp.ID+"/reopen", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var reopenedOpp struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &reopenedOpp)
		assert.Equal(t, "open", reopenedOpp.Status)
	})
}

// TestDealManagement tests deal operations.
func TestDealManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Deal Management Test", "deal-mgmt-test")
	suite.RegisterUser("deal.manager@test.com", "Password123!", "Deal", "Manager")
	suite.Login("deal.manager@test.com", "Password123!")

	// Create prerequisite data
	customer := suite.CreateCustomer("Deal Test Corp", "DEAL-001", "deal@test.com")
	pipelineResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
		"name":       "Deal Test Pipeline",
		"is_default": true,
	})
	suite.AssertStatus(pipelineResp, http.StatusCreated)
	var pipeline PipelineResponse
	suite.DecodeResponse(pipelineResp, &pipeline)

	stageID := ""
	if len(pipeline.Stages) > 0 {
		stageID = pipeline.Stages[0].ID
	}

	// Create and win an opportunity to get a deal
	oppResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities", map[string]interface{}{
		"customer_id":    customer.ID,
		"pipeline_id":    pipeline.ID,
		"stage_id":       stageID,
		"name":           "Deal Source Opportunity",
		"value_amount":   30000,
		"value_currency": "USD",
	})
	suite.AssertStatus(oppResp, http.StatusCreated)
	var opp OpportunityResponse
	suite.DecodeResponse(oppResp, &opp)

	// Win the opportunity
	winResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities/"+opp.ID+"/win", nil)
	suite.AssertStatus(winResp, http.StatusOK)
	var winResult struct {
		Deal struct {
			ID string `json:"id"`
		} `json:"deal"`
	}
	suite.DecodeResponse(winResp, &winResult)
	dealID := winResult.Deal.ID

	t.Run("Get Deal", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", salesServer.URL+"/api/v1/deals/"+dealID, nil)
		suite.AssertStatus(resp, http.StatusOK)

		var deal DealResponse
		suite.DecodeResponse(resp, &deal)
		assert.Equal(t, dealID, deal.ID)
		assert.Equal(t, int64(30000), deal.ValueAmount)
	})

	t.Run("List Deals", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", salesServer.URL+"/api/v1/deals", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		suite.DecodeResponse(resp, &listResp)
		assert.GreaterOrEqual(t, listResp.Total, 1)
	})

	t.Run("Update Deal", func(t *testing.T) {
		resp := suite.DoRequestAuth("PUT", salesServer.URL+"/api/v1/deals/"+dealID, map[string]interface{}{
			"name":         "Updated Deal Name",
			"value_amount": 35000,
		})
		suite.AssertStatus(resp, http.StatusOK)

		var updatedDeal struct {
			Name        string `json:"name"`
			ValueAmount int64  `json:"value_amount"`
		}
		suite.DecodeResponse(resp, &updatedDeal)
		assert.Equal(t, "Updated Deal Name", updatedDeal.Name)
		assert.Equal(t, int64(35000), updatedDeal.ValueAmount)
	})

	t.Run("Cancel Deal", func(t *testing.T) {
		// Create another opportunity and deal for this test
		oppResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities", map[string]interface{}{
			"customer_id":    customer.ID,
			"pipeline_id":    pipeline.ID,
			"stage_id":       stageID,
			"name":           "Cancel Test Opportunity",
			"value_amount":   5000,
			"value_currency": "USD",
		})
		suite.AssertStatus(oppResp, http.StatusCreated)
		var opp2 OpportunityResponse
		suite.DecodeResponse(oppResp, &opp2)

		winResp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/opportunities/"+opp2.ID+"/win", nil)
		suite.AssertStatus(winResp, http.StatusOK)
		var winResult2 struct {
			Deal struct {
				ID string `json:"id"`
			} `json:"deal"`
		}
		suite.DecodeResponse(winResp, &winResult2)

		// Cancel the deal
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/deals/"+winResult2.Deal.ID+"/cancel", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var cancelledDeal struct {
			Status string `json:"status"`
		}
		suite.DecodeResponse(resp, &cancelledDeal)
		assert.Equal(t, "cancelled", cancelledDeal.Status)
	})
}

// TestPipelineManagement tests pipeline and stage operations.
func TestPipelineManagement(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.Cleanup()

	// Setup
	suite.CreateTenant("Pipeline Test", "pipeline-test")
	suite.RegisterUser("pipeline.admin@test.com", "Password123!", "Pipeline", "Admin")
	suite.Login("pipeline.admin@test.com", "Password123!")

	t.Run("Create Pipeline", func(t *testing.T) {
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name":       "Custom Pipeline",
			"is_default": false,
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var pipeline PipelineResponse
		suite.DecodeResponse(resp, &pipeline)
		assert.NotEmpty(t, pipeline.ID)
		assert.Equal(t, "Custom Pipeline", pipeline.Name)
		assert.False(t, pipeline.IsDefault)
		assert.GreaterOrEqual(t, len(pipeline.Stages), 4)
	})

	t.Run("List Pipelines", func(t *testing.T) {
		resp := suite.DoRequestAuth("GET", salesServer.URL+"/api/v1/pipelines", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var listResp struct {
			Data  []map[string]interface{} `json:"data"`
			Total int                      `json:"total"`
		}
		suite.DecodeResponse(resp, &listResp)
		assert.GreaterOrEqual(t, listResp.Total, 1)
	})

	t.Run("Update Pipeline", func(t *testing.T) {
		// Create a pipeline
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name": "Update Test Pipeline",
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var pipeline PipelineResponse
		suite.DecodeResponse(resp, &pipeline)

		// Update it
		resp = suite.DoRequestAuth("PUT", salesServer.URL+"/api/v1/pipelines/"+pipeline.ID, map[string]interface{}{
			"name": "Updated Pipeline Name",
		})
		suite.AssertStatus(resp, http.StatusOK)

		var updatedPipeline struct {
			Name string `json:"name"`
		}
		suite.DecodeResponse(resp, &updatedPipeline)
		assert.Equal(t, "Updated Pipeline Name", updatedPipeline.Name)
	})

	t.Run("Create Stage", func(t *testing.T) {
		// Create a pipeline
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name": "Stage Test Pipeline",
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var pipeline PipelineResponse
		suite.DecodeResponse(resp, &pipeline)

		// Create a custom stage
		resp = suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines/"+pipeline.ID+"/stages", map[string]interface{}{
			"name":        "Custom Review Stage",
			"type":        "open",
			"stage_order": 99,
			"probability": 60,
		})
		suite.AssertStatus(resp, http.StatusCreated)

		var stage StageResponse
		suite.DecodeResponse(resp, &stage)
		assert.NotEmpty(t, stage.ID)
		assert.Equal(t, "Custom Review Stage", stage.Name)
		assert.Equal(t, 60, stage.Probability)
	})

	t.Run("Update Stage", func(t *testing.T) {
		// Create a pipeline
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name": "Stage Update Test",
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var pipeline PipelineResponse
		suite.DecodeResponse(resp, &pipeline)

		if len(pipeline.Stages) > 0 {
			stageID := pipeline.Stages[0].ID
			resp = suite.DoRequestAuth("PUT", salesServer.URL+"/api/v1/pipelines/"+pipeline.ID+"/stages/"+stageID, map[string]interface{}{
				"name":        "Renamed Stage",
				"probability": 15,
			})
			suite.AssertStatus(resp, http.StatusOK)

			var updatedStage struct {
				Name        string `json:"name"`
				Probability int    `json:"probability"`
			}
			suite.DecodeResponse(resp, &updatedStage)
			assert.Equal(t, "Renamed Stage", updatedStage.Name)
			assert.Equal(t, 15, updatedStage.Probability)
		}
	})

	t.Run("Delete Stage", func(t *testing.T) {
		// Create a pipeline
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name": "Stage Delete Test",
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var pipeline PipelineResponse
		suite.DecodeResponse(resp, &pipeline)

		// Add a custom stage
		resp = suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines/"+pipeline.ID+"/stages", map[string]interface{}{
			"name":        "To Be Deleted",
			"type":        "open",
			"stage_order": 100,
			"probability": 0,
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var stage StageResponse
		suite.DecodeResponse(resp, &stage)

		// Delete the stage
		resp = suite.DoRequestAuth("DELETE", salesServer.URL+"/api/v1/pipelines/"+pipeline.ID+"/stages/"+stage.ID, nil)
		suite.AssertStatus(resp, http.StatusNoContent)
	})

	t.Run("Reorder Stages", func(t *testing.T) {
		// Create a pipeline
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name": "Reorder Test",
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var pipeline PipelineResponse
		suite.DecodeResponse(resp, &pipeline)

		if len(pipeline.Stages) >= 2 {
			// Reverse the order
			stageIDs := make([]string, len(pipeline.Stages))
			for i, s := range pipeline.Stages {
				stageIDs[len(pipeline.Stages)-1-i] = s.ID
			}

			resp = suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines/"+pipeline.ID+"/stages/reorder", map[string]interface{}{
				"stage_ids": stageIDs,
			})
			suite.AssertStatus(resp, http.StatusOK)
		}
	})

	t.Run("Pipeline Analytics", func(t *testing.T) {
		// Create a pipeline
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name":       "Analytics Test",
			"is_default": true,
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var pipeline PipelineResponse
		suite.DecodeResponse(resp, &pipeline)

		// Get analytics
		resp = suite.DoRequestAuth("GET", salesServer.URL+"/api/v1/pipelines/"+pipeline.ID+"/analytics", nil)
		suite.AssertStatus(resp, http.StatusOK)

		var analytics struct {
			PipelineID  string         `json:"pipeline_id"`
			StageCounts map[string]int `json:"stage_counts"`
			TotalValue  int64          `json:"total_value"`
		}
		suite.DecodeResponse(resp, &analytics)
		assert.Equal(t, pipeline.ID, analytics.PipelineID)
	})

	t.Run("Cannot Delete Default Pipeline", func(t *testing.T) {
		// Create a default pipeline
		resp := suite.DoRequestAuth("POST", salesServer.URL+"/api/v1/pipelines", map[string]interface{}{
			"name":       "Default Pipeline",
			"is_default": true,
		})
		suite.AssertStatus(resp, http.StatusCreated)
		var pipeline PipelineResponse
		suite.DecodeResponse(resp, &pipeline)

		// Try to delete it
		resp = suite.DoRequestAuth("DELETE", salesServer.URL+"/api/v1/pipelines/"+pipeline.ID, nil)
		suite.AssertStatus(resp, http.StatusBadRequest)
	})
}
