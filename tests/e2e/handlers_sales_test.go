// Package e2e contains sales handlers for E2E tests.
package e2e

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ============================================================================
// Lead Handlers
// ============================================================================

func listLeadsHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var leads []map[string]interface{}
	for _, l := range dataStore.leads {
		if l.TenantID == tenantID {
			leads = append(leads, map[string]interface{}{
				"id":            l.ID,
				"tenant_id":     l.TenantID,
				"company_name":  l.CompanyName,
				"contact_name":  l.ContactName,
				"contact_email": l.ContactEmail,
				"source":        l.Source,
				"status":        l.Status,
				"score":         l.Score,
				"assigned_to":   l.AssignedTo,
				"created_at":    l.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  leads,
		"total": len(leads),
	})
}

func createLeadHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	var req struct {
		CompanyName  string `json:"company_name"`
		ContactName  string `json:"contact_name"`
		ContactEmail string `json:"contact_email"`
		Source       string `json:"source"`
		Score        int    `json:"score"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	lead := &LeadData{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		CompanyName:  req.CompanyName,
		ContactName:  req.ContactName,
		ContactEmail: req.ContactEmail,
		Source:       req.Source,
		Status:       "new",
		Score:        req.Score,
		CreatedAt:    time.Now(),
	}

	dataStore.leads[lead.ID] = lead

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":            lead.ID,
		"tenant_id":     lead.TenantID,
		"company_name":  lead.CompanyName,
		"contact_name":  lead.ContactName,
		"contact_email": lead.ContactEmail,
		"source":        lead.Source,
		"status":        lead.Status,
		"score":         lead.Score,
		"created_at":    lead.CreatedAt.Format(time.RFC3339),
	})
}

func getLeadHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	lead, ok := dataStore.leads[id]
	if !ok || lead.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "lead not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":            lead.ID,
		"tenant_id":     lead.TenantID,
		"company_name":  lead.CompanyName,
		"contact_name":  lead.ContactName,
		"contact_email": lead.ContactEmail,
		"source":        lead.Source,
		"status":        lead.Status,
		"score":         lead.Score,
		"assigned_to":   lead.AssignedTo,
		"customer_id":   lead.CustomerID,
		"created_at":    lead.CreatedAt.Format(time.RFC3339),
	})
}

func updateLeadHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	lead, ok := dataStore.leads[id]
	if !ok || lead.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "lead not found")
		return
	}

	if companyName, ok := req["company_name"].(string); ok {
		lead.CompanyName = companyName
	}
	if contactName, ok := req["contact_name"].(string); ok {
		lead.ContactName = contactName
	}
	if contactEmail, ok := req["contact_email"].(string); ok {
		lead.ContactEmail = contactEmail
	}
	if score, ok := req["score"].(float64); ok {
		lead.Score = int(score)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":            lead.ID,
		"tenant_id":     lead.TenantID,
		"company_name":  lead.CompanyName,
		"contact_name":  lead.ContactName,
		"contact_email": lead.ContactEmail,
		"source":        lead.Source,
		"status":        lead.Status,
		"score":         lead.Score,
	})
}

func deleteLeadHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	lead, ok := dataStore.leads[id]
	if !ok || lead.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "lead not found")
		return
	}

	delete(dataStore.leads, id)
	w.WriteHeader(http.StatusNoContent)
}

func qualifyLeadHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	lead, ok := dataStore.leads[id]
	if !ok || lead.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "lead not found")
		return
	}

	lead.Status = "qualified"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     lead.ID,
		"status": lead.Status,
	})
}

func disqualifyLeadHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	lead, ok := dataStore.leads[id]
	if !ok || lead.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "lead not found")
		return
	}

	lead.Status = "disqualified"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     lead.ID,
		"status": lead.Status,
	})
}

func convertLeadHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req struct {
		PipelineID    string `json:"pipeline_id"`
		StageID       string `json:"stage_id"`
		CustomerID    string `json:"customer_id"`
		ValueAmount   int64  `json:"value_amount"`
		ValueCurrency string `json:"value_currency"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	lead, ok := dataStore.leads[id]
	if !ok || lead.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "lead not found")
		return
	}

	// Create a new customer if not provided
	customerID := req.CustomerID
	if customerID == "" {
		customer := &CustomerData{
			ID:        uuid.New().String(),
			TenantID:  tenantID,
			Code:      "CUST-" + uuid.New().String()[:8],
			Name:      lead.CompanyName,
			Type:      "business",
			Status:    "active",
			Email:     map[string]interface{}{"address": lead.ContactEmail},
			CreatedAt: time.Now(),
		}
		dataStore.customers[customer.ID] = customer
		customerID = customer.ID
	}

	// Find or use provided pipeline
	pipelineID := req.PipelineID
	stageID := req.StageID
	if pipelineID == "" {
		// Find default pipeline
		for _, p := range dataStore.pipelines {
			if p.TenantID == tenantID && p.IsDefault {
				pipelineID = p.ID
				if len(p.Stages) > 0 {
					stageID = p.Stages[0].ID
				}
				break
			}
		}
	}

	// Create opportunity
	opportunity := &OpportunityData{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		CustomerID:    customerID,
		LeadID:        lead.ID,
		PipelineID:    pipelineID,
		StageID:       stageID,
		Name:          "Opportunity from " + lead.CompanyName,
		ValueAmount:   req.ValueAmount,
		ValueCurrency: req.ValueCurrency,
		Probability:   25,
		Status:        "open",
		CreatedAt:     time.Now(),
	}

	if opportunity.ValueCurrency == "" {
		opportunity.ValueCurrency = "USD"
	}

	dataStore.opportunities[opportunity.ID] = opportunity

	// Update lead status
	lead.Status = "converted"
	lead.CustomerID = customerID
	lead.ConvertedOpportunityID = opportunity.ID

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"lead": map[string]interface{}{
			"id":     lead.ID,
			"status": lead.Status,
		},
		"customer": map[string]interface{}{
			"id": customerID,
		},
		"opportunity": map[string]interface{}{
			"id":            opportunity.ID,
			"name":          opportunity.Name,
			"pipeline_id":   opportunity.PipelineID,
			"stage_id":      opportunity.StageID,
			"value_amount":  opportunity.ValueAmount,
			"value_currency": opportunity.ValueCurrency,
		},
	})
}

func assignLeadHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req struct {
		UserID string `json:"user_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	lead, ok := dataStore.leads[id]
	if !ok || lead.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "lead not found")
		return
	}

	lead.AssignedTo = req.UserID

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          lead.ID,
		"assigned_to": lead.AssignedTo,
	})
}

func getLeadActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

// ============================================================================
// Opportunity Handlers
// ============================================================================

func listOpportunitiesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var opportunities []map[string]interface{}
	for _, o := range dataStore.opportunities {
		if o.TenantID == tenantID {
			opportunities = append(opportunities, map[string]interface{}{
				"id":             o.ID,
				"tenant_id":      o.TenantID,
				"customer_id":    o.CustomerID,
				"lead_id":        o.LeadID,
				"pipeline_id":    o.PipelineID,
				"stage_id":       o.StageID,
				"name":           o.Name,
				"value_amount":   o.ValueAmount,
				"value_currency": o.ValueCurrency,
				"probability":    o.Probability,
				"status":         o.Status,
				"created_at":     o.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  opportunities,
		"total": len(opportunities),
	})
}

func createOpportunityHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	var req struct {
		CustomerID    string `json:"customer_id"`
		PipelineID    string `json:"pipeline_id"`
		StageID       string `json:"stage_id"`
		Name          string `json:"name"`
		ValueAmount   int64  `json:"value_amount"`
		ValueCurrency string `json:"value_currency"`
		Probability   int    `json:"probability"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	opportunity := &OpportunityData{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		CustomerID:    req.CustomerID,
		PipelineID:    req.PipelineID,
		StageID:       req.StageID,
		Name:          req.Name,
		ValueAmount:   req.ValueAmount,
		ValueCurrency: req.ValueCurrency,
		Probability:   req.Probability,
		Status:        "open",
		CreatedAt:     time.Now(),
	}

	if opportunity.ValueCurrency == "" {
		opportunity.ValueCurrency = "USD"
	}

	dataStore.opportunities[opportunity.ID] = opportunity

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             opportunity.ID,
		"tenant_id":      opportunity.TenantID,
		"customer_id":    opportunity.CustomerID,
		"pipeline_id":    opportunity.PipelineID,
		"stage_id":       opportunity.StageID,
		"name":           opportunity.Name,
		"value_amount":   opportunity.ValueAmount,
		"value_currency": opportunity.ValueCurrency,
		"probability":    opportunity.Probability,
		"status":         opportunity.Status,
		"created_at":     opportunity.CreatedAt.Format(time.RFC3339),
	})
}

func getOpportunityHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	opp, ok := dataStore.opportunities[id]
	if !ok || opp.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":             opp.ID,
		"tenant_id":      opp.TenantID,
		"customer_id":    opp.CustomerID,
		"lead_id":        opp.LeadID,
		"pipeline_id":    opp.PipelineID,
		"stage_id":       opp.StageID,
		"name":           opp.Name,
		"value_amount":   opp.ValueAmount,
		"value_currency": opp.ValueCurrency,
		"probability":    opp.Probability,
		"status":         opp.Status,
		"created_at":     opp.CreatedAt.Format(time.RFC3339),
	})
}

func updateOpportunityHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	opp, ok := dataStore.opportunities[id]
	if !ok || opp.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}

	if name, ok := req["name"].(string); ok {
		opp.Name = name
	}
	if amount, ok := req["value_amount"].(float64); ok {
		opp.ValueAmount = int64(amount)
	}
	if prob, ok := req["probability"].(float64); ok {
		opp.Probability = int(prob)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":             opp.ID,
		"name":           opp.Name,
		"value_amount":   opp.ValueAmount,
		"value_currency": opp.ValueCurrency,
		"probability":    opp.Probability,
		"status":         opp.Status,
	})
}

func deleteOpportunityHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	opp, ok := dataStore.opportunities[id]
	if !ok || opp.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}

	delete(dataStore.opportunities, id)
	w.WriteHeader(http.StatusNoContent)
}

func moveStageHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req struct {
		StageID string `json:"stage_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	opp, ok := dataStore.opportunities[id]
	if !ok || opp.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}

	opp.StageID = req.StageID

	// Update probability based on stage
	for _, p := range dataStore.pipelines {
		if p.ID == opp.PipelineID {
			for _, s := range p.Stages {
				if s.ID == req.StageID {
					opp.Probability = s.Probability
					break
				}
			}
			break
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          opp.ID,
		"stage_id":    opp.StageID,
		"probability": opp.Probability,
	})
}

func winOpportunityHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	opp, ok := dataStore.opportunities[id]
	if !ok || opp.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}

	opp.Status = "won"
	opp.Probability = 100

	// Create deal
	deal := &DealData{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		OpportunityID: opp.ID,
		CustomerID:    opp.CustomerID,
		Name:          opp.Name,
		ValueAmount:   opp.ValueAmount,
		ValueCurrency: opp.ValueCurrency,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	dataStore.deals[deal.ID] = deal

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"opportunity": map[string]interface{}{
			"id":     opp.ID,
			"status": opp.Status,
		},
		"deal": map[string]interface{}{
			"id":            deal.ID,
			"name":          deal.Name,
			"value_amount":  deal.ValueAmount,
			"value_currency": deal.ValueCurrency,
			"status":        deal.Status,
		},
	})
}

func loseOpportunityHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	opp, ok := dataStore.opportunities[id]
	if !ok || opp.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}

	opp.Status = "lost"
	opp.Probability = 0

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     opp.ID,
		"status": opp.Status,
	})
}

func reopenOpportunityHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	opp, ok := dataStore.opportunities[id]
	if !ok || opp.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "opportunity not found")
		return
	}

	opp.Status = "open"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     opp.ID,
		"status": opp.Status,
	})
}

func getOpportunityHistoryHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

// ============================================================================
// Deal Handlers
// ============================================================================

func listDealsHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var deals []map[string]interface{}
	for _, d := range dataStore.deals {
		if d.TenantID == tenantID {
			deals = append(deals, map[string]interface{}{
				"id":             d.ID,
				"tenant_id":      d.TenantID,
				"opportunity_id": d.OpportunityID,
				"customer_id":    d.CustomerID,
				"name":           d.Name,
				"value_amount":   d.ValueAmount,
				"value_currency": d.ValueCurrency,
				"status":         d.Status,
				"created_at":     d.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  deals,
		"total": len(deals),
	})
}

func createDealHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	var req struct {
		OpportunityID string `json:"opportunity_id"`
		CustomerID    string `json:"customer_id"`
		Name          string `json:"name"`
		ValueAmount   int64  `json:"value_amount"`
		ValueCurrency string `json:"value_currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	deal := &DealData{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		OpportunityID: req.OpportunityID,
		CustomerID:    req.CustomerID,
		Name:          req.Name,
		ValueAmount:   req.ValueAmount,
		ValueCurrency: req.ValueCurrency,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	if deal.ValueCurrency == "" {
		deal.ValueCurrency = "USD"
	}

	dataStore.deals[deal.ID] = deal

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             deal.ID,
		"tenant_id":      deal.TenantID,
		"opportunity_id": deal.OpportunityID,
		"customer_id":    deal.CustomerID,
		"name":           deal.Name,
		"value_amount":   deal.ValueAmount,
		"value_currency": deal.ValueCurrency,
		"status":         deal.Status,
		"created_at":     deal.CreatedAt.Format(time.RFC3339),
	})
}

func getDealHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	deal, ok := dataStore.deals[id]
	if !ok || deal.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "deal not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":             deal.ID,
		"tenant_id":      deal.TenantID,
		"opportunity_id": deal.OpportunityID,
		"customer_id":    deal.CustomerID,
		"name":           deal.Name,
		"value_amount":   deal.ValueAmount,
		"value_currency": deal.ValueCurrency,
		"status":         deal.Status,
		"created_at":     deal.CreatedAt.Format(time.RFC3339),
	})
}

func updateDealHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	deal, ok := dataStore.deals[id]
	if !ok || deal.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "deal not found")
		return
	}

	if name, ok := req["name"].(string); ok {
		deal.Name = name
	}
	if amount, ok := req["value_amount"].(float64); ok {
		deal.ValueAmount = int64(amount)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":           deal.ID,
		"name":         deal.Name,
		"value_amount": deal.ValueAmount,
		"status":       deal.Status,
	})
}

func deleteDealHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	deal, ok := dataStore.deals[id]
	if !ok || deal.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "deal not found")
		return
	}

	delete(dataStore.deals, id)
	w.WriteHeader(http.StatusNoContent)
}

func activateDealHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	deal, ok := dataStore.deals[id]
	if !ok || deal.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "deal not found")
		return
	}

	deal.Status = "active"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     deal.ID,
		"status": deal.Status,
	})
}

func completeDealHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	deal, ok := dataStore.deals[id]
	if !ok || deal.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "deal not found")
		return
	}

	deal.Status = "completed"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     deal.ID,
		"status": deal.Status,
	})
}

func cancelDealHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	deal, ok := dataStore.deals[id]
	if !ok || deal.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "deal not found")
		return
	}

	deal.Status = "cancelled"

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     deal.ID,
		"status": deal.Status,
	})
}

func listLineItemsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

func addLineItemHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":           uuid.New().String(),
		"product_name": "Test Product",
	})
}

func updateLineItemHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id": chi.URLParam(r, "itemId"),
	})
}

func removeLineItemHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Pipeline Handlers
// ============================================================================

func listPipelinesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var pipelines []map[string]interface{}
	for _, p := range dataStore.pipelines {
		if p.TenantID == tenantID {
			stages := make([]map[string]interface{}, 0)
			for _, s := range p.Stages {
				stages = append(stages, map[string]interface{}{
					"id":          s.ID,
					"pipeline_id": s.PipelineID,
					"name":        s.Name,
					"type":        s.Type,
					"stage_order": s.Order,
					"probability": s.Probability,
				})
			}

			pipelines = append(pipelines, map[string]interface{}{
				"id":         p.ID,
				"tenant_id":  p.TenantID,
				"name":       p.Name,
				"is_default": p.IsDefault,
				"status":     p.Status,
				"stages":     stages,
				"created_at": p.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  pipelines,
		"total": len(pipelines),
	})
}

func createPipelineHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	var req struct {
		Name      string `json:"name"`
		IsDefault bool   `json:"is_default"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	pipeline := &PipelineData{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      req.Name,
		IsDefault: req.IsDefault,
		Status:    "active",
		Stages:    make([]*StageData, 0),
		CreatedAt: time.Now(),
	}

	// If this is default, unset others
	if pipeline.IsDefault {
		for _, p := range dataStore.pipelines {
			if p.TenantID == tenantID {
				p.IsDefault = false
			}
		}
	}

	// Add default stages
	stages := []struct {
		Name        string
		Type        string
		Order       int
		Probability int
	}{
		{"Prospecting", "open", 1, 10},
		{"Qualification", "open", 2, 25},
		{"Proposal", "open", 3, 50},
		{"Negotiation", "open", 4, 75},
		{"Closed Won", "won", 5, 100},
		{"Closed Lost", "lost", 6, 0},
	}

	for _, s := range stages {
		stage := &StageData{
			ID:          uuid.New().String(),
			PipelineID:  pipeline.ID,
			Name:        s.Name,
			Type:        s.Type,
			Order:       s.Order,
			Probability: s.Probability,
		}
		pipeline.Stages = append(pipeline.Stages, stage)
	}

	dataStore.pipelines[pipeline.ID] = pipeline

	stagesResp := make([]map[string]interface{}, 0)
	for _, s := range pipeline.Stages {
		stagesResp = append(stagesResp, map[string]interface{}{
			"id":          s.ID,
			"pipeline_id": s.PipelineID,
			"name":        s.Name,
			"type":        s.Type,
			"stage_order": s.Order,
			"probability": s.Probability,
		})
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         pipeline.ID,
		"tenant_id":  pipeline.TenantID,
		"name":       pipeline.Name,
		"is_default": pipeline.IsDefault,
		"status":     pipeline.Status,
		"stages":     stagesResp,
		"created_at": pipeline.CreatedAt.Format(time.RFC3339),
	})
}

func getPipelineHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	pipeline, ok := dataStore.pipelines[id]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	stages := make([]map[string]interface{}, 0)
	for _, s := range pipeline.Stages {
		stages = append(stages, map[string]interface{}{
			"id":          s.ID,
			"pipeline_id": s.PipelineID,
			"name":        s.Name,
			"type":        s.Type,
			"stage_order": s.Order,
			"probability": s.Probability,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         pipeline.ID,
		"tenant_id":  pipeline.TenantID,
		"name":       pipeline.Name,
		"is_default": pipeline.IsDefault,
		"status":     pipeline.Status,
		"stages":     stages,
		"created_at": pipeline.CreatedAt.Format(time.RFC3339),
	})
}

func updatePipelineHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	pipeline, ok := dataStore.pipelines[id]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	if name, ok := req["name"].(string); ok {
		pipeline.Name = name
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         pipeline.ID,
		"name":       pipeline.Name,
		"is_default": pipeline.IsDefault,
		"status":     pipeline.Status,
	})
}

func deletePipelineHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	pipeline, ok := dataStore.pipelines[id]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	if pipeline.IsDefault {
		writeError(w, http.StatusBadRequest, "cannot delete default pipeline")
		return
	}

	delete(dataStore.pipelines, id)
	w.WriteHeader(http.StatusNoContent)
}

func getPipelineAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	pipeline, ok := dataStore.pipelines[id]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	// Count opportunities by stage
	stageCounts := make(map[string]int)
	var totalValue int64
	for _, o := range dataStore.opportunities {
		if o.PipelineID == id && o.TenantID == tenantID {
			stageCounts[o.StageID]++
			totalValue += o.ValueAmount
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pipeline_id":  id,
		"stage_counts": stageCounts,
		"total_value":  totalValue,
	})
}

func listStagesHandler(w http.ResponseWriter, r *http.Request) {
	pipelineID := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	pipeline, ok := dataStore.pipelines[pipelineID]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	stages := make([]map[string]interface{}, 0)
	for _, s := range pipeline.Stages {
		stages = append(stages, map[string]interface{}{
			"id":          s.ID,
			"pipeline_id": s.PipelineID,
			"name":        s.Name,
			"type":        s.Type,
			"stage_order": s.Order,
			"probability": s.Probability,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": stages})
}

func createStageHandler(w http.ResponseWriter, r *http.Request) {
	pipelineID := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Order       int    `json:"stage_order"`
		Probability int    `json:"probability"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	pipeline, ok := dataStore.pipelines[pipelineID]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	stage := &StageData{
		ID:          uuid.New().String(),
		PipelineID:  pipelineID,
		Name:        req.Name,
		Type:        req.Type,
		Order:       req.Order,
		Probability: req.Probability,
	}

	if stage.Type == "" {
		stage.Type = "open"
	}

	pipeline.Stages = append(pipeline.Stages, stage)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          stage.ID,
		"pipeline_id": stage.PipelineID,
		"name":        stage.Name,
		"type":        stage.Type,
		"stage_order": stage.Order,
		"probability": stage.Probability,
	})
}

func updateStageHandler(w http.ResponseWriter, r *http.Request) {
	pipelineID := getIDParam(r)
	stageID := chi.URLParam(r, "stageId")
	tenantID := getTenantFromHeader(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	pipeline, ok := dataStore.pipelines[pipelineID]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	for _, s := range pipeline.Stages {
		if s.ID == stageID {
			if name, ok := req["name"].(string); ok {
				s.Name = name
			}
			if prob, ok := req["probability"].(float64); ok {
				s.Probability = int(prob)
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"id":          s.ID,
				"pipeline_id": s.PipelineID,
				"name":        s.Name,
				"type":        s.Type,
				"stage_order": s.Order,
				"probability": s.Probability,
			})
			return
		}
	}

	writeError(w, http.StatusNotFound, "stage not found")
}

func deleteStageHandler(w http.ResponseWriter, r *http.Request) {
	pipelineID := getIDParam(r)
	stageID := chi.URLParam(r, "stageId")
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	pipeline, ok := dataStore.pipelines[pipelineID]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	var newStages []*StageData
	for _, s := range pipeline.Stages {
		if s.ID != stageID {
			newStages = append(newStages, s)
		}
	}
	pipeline.Stages = newStages

	w.WriteHeader(http.StatusNoContent)
}

func reorderStagesHandler(w http.ResponseWriter, r *http.Request) {
	pipelineID := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req struct {
		StageIDs []string `json:"stage_ids"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	pipeline, ok := dataStore.pipelines[pipelineID]
	if !ok || pipeline.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	// Reorder stages based on provided IDs
	stageMap := make(map[string]*StageData)
	for _, s := range pipeline.Stages {
		stageMap[s.ID] = s
	}

	newStages := make([]*StageData, 0, len(req.StageIDs))
	for i, id := range req.StageIDs {
		if s, ok := stageMap[id]; ok {
			s.Order = i + 1
			newStages = append(newStages, s)
		}
	}
	pipeline.Stages = newStages

	writeJSON(w, http.StatusOK, map[string]string{"message": "stages reordered"})
}
