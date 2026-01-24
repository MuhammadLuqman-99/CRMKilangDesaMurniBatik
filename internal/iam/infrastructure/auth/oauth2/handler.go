// Package oauth2 provides OAuth2 and OIDC authentication infrastructure.
package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// OAuth2Handler handles OAuth2 authentication HTTP requests.
type OAuth2Handler struct {
	providerManager    *ProviderManager
	linkedAccountStore LinkedAccountStore
	authCallback       AuthCallback
}

// AuthCallback is called when OAuth2 authentication is complete.
type AuthCallback interface {
	// OnOAuth2Login handles the OAuth2 login callback.
	OnOAuth2Login(ctx context.Context, request *OAuth2LoginRequest) (*OAuth2LoginResponse, error)
}

// OAuth2LoginRequest represents an OAuth2 login request.
type OAuth2LoginRequest struct {
	TenantID     uuid.UUID    `json:"tenant_id"`
	Provider     ProviderType `json:"provider"`
	ProviderID   string       `json:"provider_id"`
	Email        string       `json:"email"`
	Name         string       `json:"name"`
	Picture      string       `json:"picture,omitempty"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	ExpiresAt    *time.Time   `json:"expires_at,omitempty"`
	IsNewUser    bool         `json:"is_new_user"`
}

// OAuth2LoginResponse represents an OAuth2 login response.
type OAuth2LoginResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	IsNewUser    bool      `json:"is_new_user"`
}

// OAuth2HandlerConfig holds configuration for the OAuth2 handler.
type OAuth2HandlerConfig struct {
	ProviderManager    *ProviderManager
	LinkedAccountStore LinkedAccountStore
	AuthCallback       AuthCallback
}

// NewOAuth2Handler creates a new OAuth2 handler.
func NewOAuth2Handler(config OAuth2HandlerConfig) *OAuth2Handler {
	return &OAuth2Handler{
		providerManager:    config.ProviderManager,
		linkedAccountStore: config.LinkedAccountStore,
		authCallback:       config.AuthCallback,
	}
}

// HandleAuthorize initiates the OAuth2 authorization flow.
func (h *OAuth2Handler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	providerType := ProviderType(r.URL.Query().Get("provider"))
	if providerType == "" {
		h.writeError(w, http.StatusBadRequest, "missing provider")
		return
	}

	tenantIDStr := r.URL.Query().Get("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid tenant_id")
		return
	}

	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = h.providerManager.config.DefaultRedirectURL
	}

	authURL, err := h.providerManager.StartAuthorization(r.Context(), providerType, tenantID, redirectURL)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Redirect to OAuth2 provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleCallback handles the OAuth2 callback.
func (h *OAuth2Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorCode := r.URL.Query().Get("error")

	if errorCode != "" {
		errorDesc := r.URL.Query().Get("error_description")
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("%s: %s", errorCode, errorDesc))
		return
	}

	if code == "" || state == "" {
		h.writeError(w, http.StatusBadRequest, "missing code or state")
		return
	}

	// Complete authorization
	token, userInfo, request, err := h.providerManager.CompleteAuthorization(r.Context(), code, state)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check if account is already linked
	var isNewUser bool
	linkedAccount, err := h.linkedAccountStore.FindByProviderID(r.Context(), request.Provider, userInfo.ID)
	if err != nil {
		// New user
		isNewUser = true
	}

	// Prepare login request
	loginRequest := &OAuth2LoginRequest{
		TenantID:   request.TenantID,
		Provider:   request.Provider,
		ProviderID: userInfo.ID,
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		Picture:    userInfo.Picture,
		AccessToken: token.AccessToken,
		RefreshToken: token.RefreshToken,
		IsNewUser:  isNewUser,
	}

	if !token.ExpiresAt.IsZero() {
		loginRequest.ExpiresAt = &token.ExpiresAt
	}

	// Call auth callback
	loginResponse, err := h.authCallback.OnOAuth2Login(r.Context(), loginRequest)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update or create linked account
	if linkedAccount == nil {
		linkedAccount = &LinkedAccount{
			ID:             uuid.New(),
			UserID:         loginResponse.UserID,
			TenantID:       request.TenantID,
			Provider:       request.Provider,
			ProviderUserID: userInfo.ID,
			Email:          userInfo.Email,
			Name:           userInfo.Name,
			Picture:        userInfo.Picture,
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
			TokenExpiresAt: loginRequest.ExpiresAt,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
	} else {
		linkedAccount.AccessToken = token.AccessToken
		if token.RefreshToken != "" {
			linkedAccount.RefreshToken = token.RefreshToken
		}
		linkedAccount.TokenExpiresAt = loginRequest.ExpiresAt
		linkedAccount.UpdatedAt = time.Now().UTC()
	}

	if err := h.linkedAccountStore.Save(r.Context(), linkedAccount); err != nil {
		// Log error but don't fail the login
		_ = err
	}

	// Redirect with tokens
	if request.RedirectURL != "" {
		redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s&expires_in=%d",
			request.RedirectURL,
			loginResponse.AccessToken,
			loginResponse.RefreshToken,
			loginResponse.ExpiresIn,
		)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Return JSON response
	h.writeJSON(w, http.StatusOK, loginResponse)
}

// HandleListProviders lists available OAuth2 providers.
func (h *OAuth2Handler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.providerManager.ListProviders()

	type ProviderInfo struct {
		Type    ProviderType `json:"type"`
		Enabled bool         `json:"enabled"`
	}

	var providerInfos []ProviderInfo
	for _, p := range providers {
		providerInfos = append(providerInfos, ProviderInfo{
			Type:    p,
			Enabled: true,
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"providers": providerInfos,
	})
}

// HandleUnlinkAccount unlinks an OAuth2 account.
func (h *OAuth2Handler) HandleUnlinkAccount(w http.ResponseWriter, r *http.Request) {
	accountIDStr := r.URL.Query().Get("account_id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid account_id")
		return
	}

	if err := h.linkedAccountStore.Delete(r.Context(), accountID); err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": "account unlinked successfully",
	})
}

// HandleListLinkedAccounts lists linked OAuth2 accounts for a user.
func (h *OAuth2Handler) HandleListLinkedAccounts(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	accounts, err := h.linkedAccountStore.FindByUserID(r.Context(), userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Don't expose tokens in response
	type SafeAccount struct {
		ID             uuid.UUID    `json:"id"`
		Provider       ProviderType `json:"provider"`
		ProviderUserID string       `json:"provider_user_id"`
		Email          string       `json:"email"`
		Name           string       `json:"name"`
		Picture        string       `json:"picture,omitempty"`
		CreatedAt      time.Time    `json:"created_at"`
	}

	var safeAccounts []SafeAccount
	for _, a := range accounts {
		safeAccounts = append(safeAccounts, SafeAccount{
			ID:             a.ID,
			Provider:       a.Provider,
			ProviderUserID: a.ProviderUserID,
			Email:          a.Email,
			Name:           a.Name,
			Picture:        a.Picture,
			CreatedAt:      a.CreatedAt,
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"accounts": safeAccounts,
	})
}

// Helper methods

func (h *OAuth2Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *OAuth2Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"message": message,
		},
	})
}

// RegisterRoutes registers OAuth2 routes.
func (h *OAuth2Handler) RegisterRoutes(mux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}, prefix string) {
	mux.HandleFunc(prefix+"/authorize", h.HandleAuthorize)
	mux.HandleFunc(prefix+"/callback", h.HandleCallback)
	mux.HandleFunc(prefix+"/providers", h.HandleListProviders)
	mux.HandleFunc(prefix+"/accounts", h.HandleListLinkedAccounts)
	mux.HandleFunc(prefix+"/accounts/unlink", h.HandleUnlinkAccount)
}
