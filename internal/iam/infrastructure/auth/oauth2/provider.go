// Package oauth2 provides OAuth2 and OIDC authentication infrastructure.
package oauth2

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Provider implements OAuth2 provider operations.
type Provider struct {
	config     *ProviderConfig
	httpClient HTTPClient
}

// NewProvider creates a new OAuth2 provider.
func NewProvider(config *ProviderConfig, httpClient HTTPClient) *Provider {
	if httpClient == nil {
		httpClient = DefaultHTTPClient()
	}
	return &Provider{
		config:     config,
		httpClient: httpClient,
	}
}

// GetAuthorizationURL generates the authorization URL.
func (p *Provider) GetAuthorizationURL(state, nonce, codeChallenge string, extraParams map[string]string) string {
	params := url.Values{}
	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(p.config.Scopes, " "))
	params.Set("state", state)

	if nonce != "" {
		params.Set("nonce", nonce)
	}

	// PKCE support
	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}

	// Provider-specific parameters
	switch p.config.Type {
	case ProviderTypeGoogle:
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
	case ProviderTypeMicrosoft:
		params.Set("response_mode", "query")
	}

	// Extra parameters
	for key, value := range extraParams {
		params.Set(key, value)
	}

	return fmt.Sprintf("%s?%s", p.config.AuthURL, params.Encode())
}

// ExchangeCode exchanges an authorization code for tokens.
func (p *Provider) ExchangeCode(ctx context.Context, code, codeVerifier string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("redirect_uri", p.config.RedirectURL)

	// PKCE support
	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	headers := map[string]string{
		"Accept": "application/json",
	}

	body, err := PostForm(ctx, p.httpClient, p.config.TokenURL, data, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Try JSON first
	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		// Try form-encoded response (GitHub)
		formData, err := ParseFormResponse(body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token response: %w", err)
		}

		token = Token{
			AccessToken:  formData["access_token"],
			TokenType:    formData["token_type"],
			RefreshToken: formData["refresh_token"],
			Scope:        formData["scope"],
		}
	}

	// Check for error response
	if token.AccessToken == "" {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("oauth2 error: %s - %s", errResp.Error, errResp.ErrorDescription)
		}
		return nil, fmt.Errorf("failed to get access token")
	}

	// Calculate expiry time
	if token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	return &token, nil
}

// RefreshAccessToken refreshes an access token using a refresh token.
func (p *Provider) RefreshAccessToken(ctx context.Context, refreshToken string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	headers := map[string]string{
		"Accept": "application/json",
	}

	body, err := PostForm(ctx, p.httpClient, p.config.TokenURL, data, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("failed to refresh access token")
	}

	if token.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	return &token, nil
}

// GetUserInfo fetches user information using an access token.
func (p *Provider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.config.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	userInfo := p.parseUserInfo(raw)
	userInfo.Provider = p.config.Type
	userInfo.Raw = raw

	return userInfo, nil
}

// parseUserInfo parses user info from different providers.
func (p *Provider) parseUserInfo(raw map[string]interface{}) *UserInfo {
	userInfo := &UserInfo{}

	switch p.config.Type {
	case ProviderTypeGoogle, ProviderTypeMicrosoft:
		userInfo.ID = getString(raw, "sub")
		userInfo.Email = getString(raw, "email")
		userInfo.EmailVerified = getBool(raw, "email_verified")
		userInfo.Name = getString(raw, "name")
		userInfo.GivenName = getString(raw, "given_name")
		userInfo.FamilyName = getString(raw, "family_name")
		userInfo.Picture = getString(raw, "picture")
		userInfo.Locale = getString(raw, "locale")

	case ProviderTypeGitHub:
		userInfo.ID = fmt.Sprintf("%v", raw["id"])
		userInfo.Email = getString(raw, "email")
		userInfo.Name = getString(raw, "name")
		if userInfo.Name == "" {
			userInfo.Name = getString(raw, "login")
		}
		userInfo.Picture = getString(raw, "avatar_url")
		// GitHub doesn't verify email in the same way
		userInfo.EmailVerified = getString(raw, "email") != ""

	default:
		// Generic parsing
		userInfo.ID = getString(raw, "sub")
		if userInfo.ID == "" {
			userInfo.ID = getString(raw, "id")
		}
		userInfo.Email = getString(raw, "email")
		userInfo.Name = getString(raw, "name")
		userInfo.Picture = getString(raw, "picture")
	}

	return userInfo
}

// RevokeToken revokes an access or refresh token.
func (p *Provider) RevokeToken(ctx context.Context, token string) error {
	// Not all providers support token revocation
	var revokeURL string

	switch p.config.Type {
	case ProviderTypeGoogle:
		revokeURL = "https://oauth2.googleapis.com/revoke"
	case ProviderTypeMicrosoft:
		// Microsoft doesn't have a standard revocation endpoint
		return nil
	case ProviderTypeGitHub:
		// GitHub revocation requires different approach
		return nil
	default:
		return nil
	}

	data := url.Values{}
	data.Set("token", token)

	_, err := PostForm(ctx, p.httpClient, revokeURL, data, nil)
	return err
}

// Config returns the provider configuration.
func (p *Provider) Config() *ProviderConfig {
	return p.config
}

// Helper functions

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GeneratePKCEChallenge generates a PKCE code challenge from a verifier.
func GeneratePKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// ProviderManager manages multiple OAuth2 providers.
type ProviderManager struct {
	providers  map[ProviderType]*Provider
	stateStore StateStore
	config     OAuth2Config
}

// NewProviderManager creates a new provider manager.
func NewProviderManager(config OAuth2Config, stateStore StateStore) *ProviderManager {
	if stateStore == nil {
		stateStore = NewInMemoryStateStore()
	}

	pm := &ProviderManager{
		providers:  make(map[ProviderType]*Provider),
		stateStore: stateStore,
		config:     config,
	}

	// Initialize configured providers
	for providerType, providerConfig := range config.Providers {
		if providerConfig.Enabled {
			pm.providers[providerType] = NewProvider(providerConfig, nil)
		}
	}

	return pm
}

// GetProvider returns a provider by type.
func (pm *ProviderManager) GetProvider(providerType ProviderType) (*Provider, error) {
	provider, ok := pm.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerType)
	}
	return provider, nil
}

// ListProviders returns all enabled providers.
func (pm *ProviderManager) ListProviders() []ProviderType {
	var providers []ProviderType
	for providerType := range pm.providers {
		providers = append(providers, providerType)
	}
	return providers
}

// StartAuthorization starts the OAuth2 authorization flow.
func (pm *ProviderManager) StartAuthorization(ctx context.Context, providerType ProviderType, tenantID uuid.UUID, redirectURL string) (string, error) {
	provider, err := pm.GetProvider(providerType)
	if err != nil {
		return "", err
	}

	state, err := GenerateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	nonce, err := GenerateNonce()
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Generate PKCE verifier
	codeVerifier, err := GenerateCodeVerifier()
	if err != nil {
		return "", fmt.Errorf("failed to generate code verifier: %w", err)
	}

	codeChallenge := GeneratePKCEChallenge(codeVerifier)

	// Store authorization request
	request := &AuthorizationRequest{
		State:        state,
		Nonce:        nonce,
		Provider:     providerType,
		TenantID:     tenantID,
		RedirectURL:  redirectURL,
		CodeVerifier: codeVerifier,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(pm.config.StateTokenExpiry),
	}

	if err := pm.stateStore.Save(ctx, state, request); err != nil {
		return "", fmt.Errorf("failed to save state: %w", err)
	}

	authURL := provider.GetAuthorizationURL(state, nonce, codeChallenge, nil)
	return authURL, nil
}

// CompleteAuthorization completes the OAuth2 authorization flow.
func (pm *ProviderManager) CompleteAuthorization(ctx context.Context, code, state string) (*Token, *UserInfo, *AuthorizationRequest, error) {
	// Verify state
	request, err := pm.stateStore.Get(ctx, state)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid or expired state: %w", err)
	}

	// Delete state (one-time use)
	_ = pm.stateStore.Delete(ctx, state)

	provider, err := pm.GetProvider(request.Provider)
	if err != nil {
		return nil, nil, nil, err
	}

	// Exchange code for tokens
	token, err := provider.ExchangeCode(ctx, code, request.CodeVerifier)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return token, userInfo, request, nil
}

// IsProviderEnabled checks if a provider is enabled.
func (pm *ProviderManager) IsProviderEnabled(providerType ProviderType) bool {
	_, ok := pm.providers[providerType]
	return ok
}
