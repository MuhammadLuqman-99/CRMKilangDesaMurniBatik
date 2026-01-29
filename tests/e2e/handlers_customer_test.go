// Package e2e contains customer handlers for E2E tests.
package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ============================================================================
// Customer Handlers
// ============================================================================

func listCustomersHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var customers []map[string]interface{}
	for _, c := range dataStore.customers {
		if c.TenantID == tenantID {
			customers = append(customers, map[string]interface{}{
				"id":         c.ID,
				"tenant_id":  c.TenantID,
				"code":       c.Code,
				"name":       c.Name,
				"type":       c.Type,
				"status":     c.Status,
				"email":      c.Email,
				"created_at": c.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  customers,
		"total": len(customers),
	})
}

func createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	var req struct {
		Code   string                 `json:"code"`
		Name   string                 `json:"name"`
		Type   string                 `json:"type"`
		Status string                 `json:"status"`
		Email  map[string]interface{} `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	// Check code uniqueness within tenant
	for _, c := range dataStore.customers {
		if c.TenantID == tenantID && c.Code == req.Code {
			writeError(w, http.StatusConflict, "customer code already exists")
			return
		}
	}

	customer := &CustomerData{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Code:      req.Code,
		Name:      req.Name,
		Type:      req.Type,
		Status:    req.Status,
		Email:     req.Email,
		Version:   1,
		CreatedAt: time.Now(),
	}

	if customer.Status == "" {
		customer.Status = "active"
	}

	dataStore.customers[customer.ID] = customer

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         customer.ID,
		"tenant_id":  customer.TenantID,
		"code":       customer.Code,
		"name":       customer.Name,
		"type":       customer.Type,
		"status":     customer.Status,
		"email":      customer.Email,
		"created_at": customer.CreatedAt.Format(time.RFC3339),
	})
}

func searchCustomersHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)
	query := r.URL.Query().Get("q")

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var customers []map[string]interface{}
	for _, c := range dataStore.customers {
		if c.TenantID == tenantID {
			if query == "" || strings.Contains(strings.ToLower(c.Name), strings.ToLower(query)) {
				customers = append(customers, map[string]interface{}{
					"id":     c.ID,
					"code":   c.Code,
					"name":   c.Name,
					"type":   c.Type,
					"status": c.Status,
				})
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  customers,
		"total": len(customers),
	})
}

func importCustomersHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"imported": 0,
		"failed":   0,
	})
}

func exportCustomersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	w.Write([]byte("id,name,code,type,status\n"))
}

func getCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	customer, ok := dataStore.customers[id]
	if !ok || customer.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         customer.ID,
		"tenant_id":  customer.TenantID,
		"code":       customer.Code,
		"name":       customer.Name,
		"type":       customer.Type,
		"status":     customer.Status,
		"email":      customer.Email,
		"version":    customer.Version,
		"created_at": customer.CreatedAt.Format(time.RFC3339),
	})
}

func updateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[id]
	if !ok || customer.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	// Optimistic locking check
	if version, ok := req["version"].(float64); ok {
		if int(version) != customer.Version {
			writeError(w, http.StatusConflict, "version mismatch")
			return
		}
	}

	if name, ok := req["name"].(string); ok {
		customer.Name = name
	}
	if status, ok := req["status"].(string); ok {
		customer.Status = status
	}
	if email, ok := req["email"].(map[string]interface{}); ok {
		customer.Email = email
	}

	customer.Version++

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         customer.ID,
		"tenant_id":  customer.TenantID,
		"code":       customer.Code,
		"name":       customer.Name,
		"type":       customer.Type,
		"status":     customer.Status,
		"email":      customer.Email,
		"version":    customer.Version,
		"created_at": customer.CreatedAt.Format(time.RFC3339),
	})
}

func deleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[id]
	if !ok || customer.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	delete(dataStore.customers, id)
	w.WriteHeader(http.StatusNoContent)
}

func restoreCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[id]
	if !ok {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	customer.Status = "active"
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     customer.ID,
		"status": customer.Status,
	})
}

func activateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[id]
	if !ok || customer.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	customer.Status = "active"
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     customer.ID,
		"status": customer.Status,
	})
}

func deactivateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	tenantID := getTenantFromHeader(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[id]
	if !ok || customer.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	customer.Status = "inactive"
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     customer.ID,
		"status": customer.Status,
	})
}

// ============================================================================
// Contact Handlers
// ============================================================================

func listContactsHandler(w http.ResponseWriter, r *http.Request) {
	customerID := getIDParam(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	customer, ok := dataStore.customers[customerID]
	if !ok {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	var contacts []map[string]interface{}
	for _, c := range customer.Contacts {
		contacts = append(contacts, map[string]interface{}{
			"id":          c.ID,
			"customer_id": c.CustomerID,
			"first_name":  c.FirstName,
			"last_name":   c.LastName,
			"email":       c.Email,
			"is_primary":  c.IsPrimary,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": contacts})
}

func createContactHandler(w http.ResponseWriter, r *http.Request) {
	customerID := getIDParam(r)

	var req struct {
		FirstName string                 `json:"first_name"`
		LastName  string                 `json:"last_name"`
		Email     map[string]interface{} `json:"email"`
		IsPrimary bool                   `json:"is_primary"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[customerID]
	if !ok {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	contact := &ContactData{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		IsPrimary:  req.IsPrimary,
	}

	// If this is primary, unset others
	if contact.IsPrimary {
		for _, c := range customer.Contacts {
			c.IsPrimary = false
		}
	}

	customer.Contacts = append(customer.Contacts, contact)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          contact.ID,
		"customer_id": contact.CustomerID,
		"first_name":  contact.FirstName,
		"last_name":   contact.LastName,
		"email":       contact.Email,
		"is_primary":  contact.IsPrimary,
	})
}

func getContactHandler(w http.ResponseWriter, r *http.Request) {
	customerID := getIDParam(r)
	contactID := chi.URLParam(r, "contactId")

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	customer, ok := dataStore.customers[customerID]
	if !ok {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	for _, c := range customer.Contacts {
		if c.ID == contactID {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"id":          c.ID,
				"customer_id": c.CustomerID,
				"first_name":  c.FirstName,
				"last_name":   c.LastName,
				"email":       c.Email,
				"is_primary":  c.IsPrimary,
			})
			return
		}
	}

	writeError(w, http.StatusNotFound, "contact not found")
}

func updateContactHandler(w http.ResponseWriter, r *http.Request) {
	customerID := getIDParam(r)
	contactID := chi.URLParam(r, "contactId")

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[customerID]
	if !ok {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	for _, c := range customer.Contacts {
		if c.ID == contactID {
			if fn, ok := req["first_name"].(string); ok {
				c.FirstName = fn
			}
			if ln, ok := req["last_name"].(string); ok {
				c.LastName = ln
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"id":          c.ID,
				"customer_id": c.CustomerID,
				"first_name":  c.FirstName,
				"last_name":   c.LastName,
				"email":       c.Email,
				"is_primary":  c.IsPrimary,
			})
			return
		}
	}

	writeError(w, http.StatusNotFound, "contact not found")
}

func deleteContactHandler(w http.ResponseWriter, r *http.Request) {
	customerID := getIDParam(r)
	contactID := chi.URLParam(r, "contactId")

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[customerID]
	if !ok {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	var newContacts []*ContactData
	for _, c := range customer.Contacts {
		if c.ID != contactID {
			newContacts = append(newContacts, c)
		}
	}
	customer.Contacts = newContacts

	w.WriteHeader(http.StatusNoContent)
}

func setPrimaryContactHandler(w http.ResponseWriter, r *http.Request) {
	customerID := getIDParam(r)
	contactID := chi.URLParam(r, "contactId")

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	customer, ok := dataStore.customers[customerID]
	if !ok {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}

	for _, c := range customer.Contacts {
		c.IsPrimary = c.ID == contactID
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "primary contact set"})
}

// ============================================================================
// Activity and Note Handlers
// ============================================================================

func listActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

func createActivityHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":   uuid.New().String(),
		"type": "note",
	})
}

func listNotesHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

func createNoteHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      uuid.New().String(),
		"content": "",
	})
}

func updateNoteHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id": chi.URLParam(r, "noteId"),
	})
}

func deleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func pinNoteHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "note pinned"})
}

// ============================================================================
// Segment Handlers
// ============================================================================

func listSegmentsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

func createSegmentHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":   uuid.New().String(),
		"name": "test-segment",
	})
}

func getSegmentHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":   getIDParam(r),
		"name": "test-segment",
	})
}

func updateSegmentHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id": getIDParam(r),
	})
}

func deleteSegmentHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func getSegmentCustomersHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": []interface{}{}})
}

func refreshSegmentHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "segment refreshed"})
}
