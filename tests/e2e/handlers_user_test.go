// Package e2e contains user and tenant handlers for E2E tests.
package e2e

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ============================================================================
// User Handlers
// ============================================================================

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var users []map[string]interface{}
	for _, u := range dataStore.users {
		if u.TenantID == tenantID {
			users = append(users, map[string]interface{}{
				"id":         u.ID,
				"email":      u.Email,
				"first_name": u.FirstName,
				"last_name":  u.LastName,
				"status":     u.Status,
				"created_at": u.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  users,
		"total": len(users),
	})
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	user := &UserData{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	dataStore.users[user.ID] = user

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"status":     user.Status,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	})
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	user, ok := dataStore.users[id]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"status":     user.Status,
		"created_at": user.CreatedAt.Format(time.RFC3339),
	})
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	user, ok := dataStore.users[id]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if fn, ok := req["first_name"].(string); ok {
		user.FirstName = fn
	}
	if ln, ok := req["last_name"].(string); ok {
		user.LastName = ln
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"status":     user.Status,
	})
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	if _, ok := dataStore.users[id]; !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	delete(dataStore.users, id)
	w.WriteHeader(http.StatusNoContent)
}

func activateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	user, ok := dataStore.users[id]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	user.Status = "active"
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     user.ID,
		"status": user.Status,
	})
}

func suspendUserHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	user, ok := dataStore.users[id]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	user.Status = "suspended"
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     user.ID,
		"status": user.Status,
	})
}

func getUserRolesHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	user, ok := dataStore.users[id]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var roles []map[string]interface{}
	for _, roleID := range user.Roles {
		if role, ok := dataStore.roles[roleID]; ok {
			roles = append(roles, map[string]interface{}{
				"id":   role.ID,
				"name": role.Name,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"roles": roles})
}

func assignRoleHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	var req struct {
		RoleID string `json:"role_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	user, ok := dataStore.users[id]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	user.Roles = append(user.Roles, req.RoleID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "role assigned"})
}

func removeRoleHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)
	roleID := chi.URLParam(r, "roleId")

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	user, ok := dataStore.users[id]
	if !ok {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	var newRoles []string
	for _, rid := range user.Roles {
		if rid != roleID {
			newRoles = append(newRoles, rid)
		}
	}
	user.Roles = newRoles

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Role Handlers
// ============================================================================

func listRolesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var roles []map[string]interface{}
	for _, role := range dataStore.roles {
		if role.TenantID == tenantID || role.IsSystem {
			roles = append(roles, map[string]interface{}{
				"id":          role.ID,
				"name":        role.Name,
				"description": role.Description,
				"is_system":   role.IsSystem,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": roles})
}

func createRoleHandler(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantFromHeader(r)

	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	role := &RoleData{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
		IsSystem:    false,
	}

	dataStore.roles[role.ID] = role

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description,
	})
}

func getSystemRolesHandler(w http.ResponseWriter, r *http.Request) {
	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var roles []map[string]interface{}
	for _, role := range dataStore.roles {
		if role.IsSystem {
			roles = append(roles, map[string]interface{}{
				"id":          role.ID,
				"name":        role.Name,
				"description": role.Description,
				"is_system":   true,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": roles})
}

func getRoleHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	role, ok := dataStore.roles[id]
	if !ok {
		writeError(w, http.StatusNotFound, "role not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description,
		"permissions": role.Permissions,
		"is_system":   role.IsSystem,
	})
}

func updateRoleHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	role, ok := dataStore.roles[id]
	if !ok {
		writeError(w, http.StatusNotFound, "role not found")
		return
	}

	if role.IsSystem {
		writeError(w, http.StatusForbidden, "cannot modify system role")
		return
	}

	if name, ok := req["name"].(string); ok {
		role.Name = name
	}
	if desc, ok := req["description"].(string); ok {
		role.Description = desc
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description,
	})
}

func deleteRoleHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	role, ok := dataStore.roles[id]
	if !ok {
		writeError(w, http.StatusNotFound, "role not found")
		return
	}

	if role.IsSystem {
		writeError(w, http.StatusForbidden, "cannot delete system role")
		return
	}

	delete(dataStore.roles, id)
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Tenant Handlers
// ============================================================================

func listTenantsHandler(w http.ResponseWriter, r *http.Request) {
	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	var tenants []map[string]interface{}
	for _, t := range dataStore.tenants {
		tenants = append(tenants, map[string]interface{}{
			"id":         t.ID,
			"name":       t.Name,
			"slug":       t.Slug,
			"status":     t.Status,
			"plan":       t.Plan,
			"created_at": t.CreatedAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": tenants})
}

func createTenantHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string                 `json:"name"`
		Slug     string                 `json:"slug"`
		Plan     string                 `json:"plan"`
		Settings map[string]interface{} `json:"settings"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	// Check slug uniqueness
	for _, t := range dataStore.tenants {
		if t.Slug == req.Slug {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
	}

	tenant := &TenantData{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Slug:      req.Slug,
		Status:    "active",
		Plan:      req.Plan,
		Settings:  req.Settings,
		CreatedAt: time.Now(),
	}

	if tenant.Plan == "" {
		tenant.Plan = "free"
	}

	dataStore.tenants[tenant.ID] = tenant

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"slug":       tenant.Slug,
		"status":     tenant.Status,
		"plan":       tenant.Plan,
		"settings":   tenant.Settings,
		"created_at": tenant.CreatedAt.Format(time.RFC3339),
	})
}

func getTenantHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	tenant, ok := dataStore.tenants[id]
	if !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"slug":       tenant.Slug,
		"status":     tenant.Status,
		"plan":       tenant.Plan,
		"settings":   tenant.Settings,
		"created_at": tenant.CreatedAt.Format(time.RFC3339),
	})
}

func updateTenantHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	tenant, ok := dataStore.tenants[id]
	if !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	if name, ok := req["name"].(string); ok {
		tenant.Name = name
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     tenant.ID,
		"name":   tenant.Name,
		"slug":   tenant.Slug,
		"status": tenant.Status,
		"plan":   tenant.Plan,
	})
}

func deleteTenantHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	if _, ok := dataStore.tenants[id]; !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	delete(dataStore.tenants, id)
	w.WriteHeader(http.StatusNoContent)
}

func updateTenantStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	var req struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	tenant, ok := dataStore.tenants[id]
	if !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	tenant.Status = req.Status
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":     tenant.ID,
		"status": tenant.Status,
	})
}

func updateTenantPlanHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	var req struct {
		Plan string `json:"plan"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dataStore.mu.Lock()
	defer dataStore.mu.Unlock()

	tenant, ok := dataStore.tenants[id]
	if !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	tenant.Plan = req.Plan
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":   tenant.ID,
		"plan": tenant.Plan,
	})
}

func getTenantStatsHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDParam(r)

	dataStore.mu.RLock()
	defer dataStore.mu.RUnlock()

	if _, ok := dataStore.tenants[id]; !ok {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	// Count users, customers, leads for this tenant
	var userCount, customerCount, leadCount int
	for _, u := range dataStore.users {
		if u.TenantID == id {
			userCount++
		}
	}
	for _, c := range dataStore.customers {
		if c.TenantID == id {
			customerCount++
		}
	}
	for _, l := range dataStore.leads {
		if l.TenantID == id {
			leadCount++
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users":     userCount,
		"customers": customerCount,
		"leads":     leadCount,
	})
}
