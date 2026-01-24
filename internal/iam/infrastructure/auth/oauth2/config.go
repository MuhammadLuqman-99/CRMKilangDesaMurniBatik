// Package oauth2 provides OAuth2 and OIDC authentication infrastructure.
package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ProviderType represents the type of OAuth2 provider.
type ProviderType string

const (
	ProviderTypeGoogle    ProviderType = "google"
	ProviderTypeMicrosoft ProviderType = "microsoft"
	ProviderTypeGitHub    ProviderType = "github"
	ProviderTypeCustom    ProviderType = "custom"
)

// ProviderConfig holds OAuth2 provider configuration.
type ProviderConfig struct {
	Type            ProviderType `json:"type"`
	ClientID        string       `json:"client_id"`
	ClientSecret    string       `json:"client_secret"`
	RedirectURL     string       `json:"redirect_url"`
	Scopes          []string     `json:"scopes"`
	AuthURL         string       `json:"auth_url,omitempty"`
	TokenURL        string       `json:"token_url,omitempty"`
	UserInfoURL     string       `json:"user_info_url,omitempty"`
	Issuer          string       `json:"issuer,omitempty"`
	JWKSURL         string       `json:"jwks_url,omitempty"`
	Enabled         bool         `json:"enabled"`
	AllowSignup     bool         `json:"allow_signup"`
	AutoLinkAccount bool         `json:"auto_link_account"`
}

// DefaultProviderConfigs returns default configurations for known providers.
func DefaultProviderConfigs() map[ProviderType]ProviderConfig {
	return map[ProviderType]ProviderConfig{
		ProviderTypeGoogle: {
			Type:        ProviderTypeGoogle,
			AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:    "https://oauth2.googleapis.com/token",
			UserInfoURL: "https://openidconnect.googleapis.com/v1/userinfo",
			Issuer:      "https://accounts.google.com",
			JWKSURL:     "https://www.googleapis.com/oauth2/v3/certs",
			Scopes:      []string{"openid", "email", "profile"},
		},
		ProviderTypeMicrosoft: {
			Type:        ProviderTypeMicrosoft,
			AuthURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			UserInfoURL: "https://graph.microsoft.com/oidc/userinfo",
			Issuer:      "https://login.microsoftonline.com/common/v2.0",
			JWKSURL:     "https://login.microsoftonline.com/common/discovery/v2.0/keys",
			Scopes:      []string{"openid", "email", "profile"},
		},
		ProviderTypeGitHub: {
			Type:        ProviderTypeGitHub,
			AuthURL:     "https://github.com/login/oauth/authorize",
			TokenURL:    "https://github.com/login/oauth/access_token",
			UserInfoURL: "https://api.github.com/user",
			Scopes:      []string{"user:email", "read:user"},
		},
	}
}

// OAuth2Config holds the complete OAuth2 configuration.
type OAuth2Config struct {
	Providers          map[ProviderType]*ProviderConfig `json:"providers"`
	StateTokenSecret   string                           `json:"state_token_secret"`
	StateTokenExpiry   time.Duration                    `json:"state_token_expiry"`
	DefaultRedirectURL string                           `json:"default_redirect_url"`
}

// DefaultOAuth2Config returns default OAuth2 configuration.
func DefaultOAuth2Config() OAuth2Config {
	return OAuth2Config{
		Providers:        make(map[ProviderType]*ProviderConfig),
		StateTokenExpiry: 10 * time.Minute,
	}
}

// Token represents an OAuth2 token.
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
}

// IsExpired checks if the token has expired.
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(t.ExpiresAt.Add(-30 * time.Second)) // 30 second buffer
}

// UserInfo represents user information from OAuth2 provider.
type UserInfo struct {
	ID            string                 `json:"sub,omitempty"`
	Email         string                 `json:"email,omitempty"`
	EmailVerified bool                   `json:"email_verified,omitempty"`
	Name          string                 `json:"name,omitempty"`
	GivenName     string                 `json:"given_name,omitempty"`
	FamilyName    string                 `json:"family_name,omitempty"`
	Picture       string                 `json:"picture,omitempty"`
	Locale        string                 `json:"locale,omitempty"`
	Provider      ProviderType           `json:"provider"`
	Raw           map[string]interface{} `json:"raw,omitempty"`
}

// AuthorizationRequest represents an OAuth2 authorization request.
type AuthorizationRequest struct {
	State        string       `json:"state"`
	Nonce        string       `json:"nonce,omitempty"`
	Provider     ProviderType `json:"provider"`
	TenantID     uuid.UUID    `json:"tenant_id"`
	RedirectURL  string       `json:"redirect_url"`
	CodeVerifier string       `json:"code_verifier,omitempty"` // For PKCE
	CreatedAt    time.Time    `json:"created_at"`
	ExpiresAt    time.Time    `json:"expires_at"`
}

// AuthorizationResponse represents an OAuth2 authorization response.
type AuthorizationResponse struct {
	Code  string `json:"code"`
	State string `json:"state"`
	Error string `json:"error,omitempty"`
}

// StateStore interface for storing OAuth2 state tokens.
type StateStore interface {
	Save(ctx context.Context, state string, request *AuthorizationRequest) error
	Get(ctx context.Context, state string) (*AuthorizationRequest, error)
	Delete(ctx context.Context, state string) error
}

// LinkedAccountStore interface for storing linked OAuth2 accounts.
type LinkedAccountStore interface {
	Save(ctx context.Context, account *LinkedAccount) error
	FindByProviderID(ctx context.Context, provider ProviderType, providerUserID string) (*LinkedAccount, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*LinkedAccount, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// LinkedAccount represents a linked OAuth2 account.
type LinkedAccount struct {
	ID             uuid.UUID    `json:"id"`
	UserID         uuid.UUID    `json:"user_id"`
	TenantID       uuid.UUID    `json:"tenant_id"`
	Provider       ProviderType `json:"provider"`
	ProviderUserID string       `json:"provider_user_id"`
	Email          string       `json:"email"`
	Name           string       `json:"name"`
	Picture        string       `json:"picture,omitempty"`
	AccessToken    string       `json:"access_token,omitempty"`
	RefreshToken   string       `json:"refresh_token,omitempty"`
	TokenExpiresAt *time.Time   `json:"token_expires_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// InMemoryStateStore implements StateStore using in-memory storage.
type InMemoryStateStore struct {
	mu     sync.RWMutex
	states map[string]*AuthorizationRequest
}

// NewInMemoryStateStore creates a new in-memory state store.
func NewInMemoryStateStore() *InMemoryStateStore {
	store := &InMemoryStateStore{
		states: make(map[string]*AuthorizationRequest),
	}
	go store.cleanup()
	return store
}

// Save saves a state token.
func (s *InMemoryStateStore) Save(ctx context.Context, state string, request *AuthorizationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = request
	return nil
}

// Get retrieves a state token.
func (s *InMemoryStateStore) Get(ctx context.Context, state string) (*AuthorizationRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	request, ok := s.states[state]
	if !ok {
		return nil, fmt.Errorf("state not found")
	}

	if time.Now().After(request.ExpiresAt) {
		return nil, fmt.Errorf("state expired")
	}

	return request, nil
}

// Delete deletes a state token.
func (s *InMemoryStateStore) Delete(ctx context.Context, state string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, state)
	return nil
}

func (s *InMemoryStateStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for state, request := range s.states {
			if now.After(request.ExpiresAt) {
				delete(s.states, state)
			}
		}
		s.mu.Unlock()
	}
}

// GenerateState generates a secure random state token.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateNonce generates a secure random nonce for OIDC.
func GenerateNonce() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateCodeVerifier generates a PKCE code verifier.
func GenerateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ParseFormResponse parses an OAuth2 form-encoded response.
func ParseFormResponse(body []byte) (map[string]string, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for key := range values {
		result[key] = values.Get(key)
	}
	return result, nil
}

// HTTPClient interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient returns a default HTTP client with timeout.
func DefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// FetchJSON fetches JSON from a URL.
func FetchJSON(ctx context.Context, client HTTPClient, url string, headers map[string]string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// PostForm posts form data and returns the response.
func PostForm(ctx context.Context, client HTTPClient, url string, data url.Values, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
