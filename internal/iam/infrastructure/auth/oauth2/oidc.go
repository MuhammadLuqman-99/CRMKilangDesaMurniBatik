// Package oauth2 provides OAuth2 and OIDC authentication infrastructure.
package oauth2

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// OIDCDiscoveryDocument represents an OIDC discovery document.
type OIDCDiscoveryDocument struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserInfoEndpoint                 string   `json:"userinfo_endpoint"`
	JwksURI                          string   `json:"jwks_uri"`
	RegistrationEndpoint             string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                  []string `json:"scopes_supported,omitempty"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	ResponseModesSupported           []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported              []string `json:"grant_types_supported,omitempty"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                  []string `json:"claims_supported,omitempty"`
	CodeChallengeMethodsSupported    []string `json:"code_challenge_methods_supported,omitempty"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key.
type JWK struct {
	Kty string `json:"kty"` // Key type (RSA, EC, etc.)
	Use string `json:"use"` // Key use (sig, enc)
	Kid string `json:"kid"` // Key ID
	Alg string `json:"alg"` // Algorithm
	N   string `json:"n"`   // RSA modulus
	E   string `json:"e"`   // RSA exponent
	X   string `json:"x"`   // EC x coordinate
	Y   string `json:"y"`   // EC y coordinate
	Crv string `json:"crv"` // EC curve
}

// ToRSAPublicKey converts a JWK to an RSA public key.
func (j *JWK) ToRSAPublicKey() (*rsa.PublicKey, error) {
	if j.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", j.Kty)
	}

	// Decode modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// IDTokenClaims represents claims in an OIDC ID token.
type IDTokenClaims struct {
	Issuer          string   `json:"iss"`
	Subject         string   `json:"sub"`
	Audience        Audience `json:"aud"`
	ExpiresAt       int64    `json:"exp"`
	IssuedAt        int64    `json:"iat"`
	AuthTime        int64    `json:"auth_time,omitempty"`
	Nonce           string   `json:"nonce,omitempty"`
	ACR             string   `json:"acr,omitempty"`
	AMR             []string `json:"amr,omitempty"`
	AZP             string   `json:"azp,omitempty"`
	AtHash          string   `json:"at_hash,omitempty"`
	CHash           string   `json:"c_hash,omitempty"`
	Email           string   `json:"email,omitempty"`
	EmailVerified   bool     `json:"email_verified,omitempty"`
	Name            string   `json:"name,omitempty"`
	GivenName       string   `json:"given_name,omitempty"`
	FamilyName      string   `json:"family_name,omitempty"`
	Picture         string   `json:"picture,omitempty"`
	Profile         string   `json:"profile,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
}

// Audience can be a string or array of strings.
type Audience []string

// UnmarshalJSON handles both string and array audience values.
func (a *Audience) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		*a = multiple
		return nil
	}

	return fmt.Errorf("invalid audience format")
}

// Contains checks if the audience contains a specific value.
func (a Audience) Contains(value string) bool {
	for _, v := range a {
		if v == value {
			return true
		}
	}
	return false
}

// JWTHeader represents a JWT header.
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid,omitempty"`
}

// OIDCValidator validates OIDC ID tokens.
type OIDCValidator struct {
	issuer       string
	clientID     string
	jwksURL      string
	httpClient   HTTPClient
	keyCache     map[string]*rsa.PublicKey
	keyCacheMu   sync.RWMutex
	cacheExpiry  time.Duration
	lastFetched  time.Time
	discovery    *OIDCDiscoveryDocument
	discoveryMu  sync.RWMutex
}

// OIDCValidatorConfig holds configuration for the OIDC validator.
type OIDCValidatorConfig struct {
	Issuer      string
	ClientID    string
	JwksURL     string
	HTTPClient  HTTPClient
	CacheExpiry time.Duration
}

// NewOIDCValidator creates a new OIDC validator.
func NewOIDCValidator(config OIDCValidatorConfig) *OIDCValidator {
	if config.HTTPClient == nil {
		config.HTTPClient = DefaultHTTPClient()
	}
	if config.CacheExpiry == 0 {
		config.CacheExpiry = time.Hour
	}

	return &OIDCValidator{
		issuer:      config.Issuer,
		clientID:    config.ClientID,
		jwksURL:     config.JwksURL,
		httpClient:  config.HTTPClient,
		keyCache:    make(map[string]*rsa.PublicKey),
		cacheExpiry: config.CacheExpiry,
	}
}

// DiscoverConfiguration fetches the OIDC discovery document.
func (v *OIDCValidator) DiscoverConfiguration(ctx context.Context) (*OIDCDiscoveryDocument, error) {
	v.discoveryMu.RLock()
	if v.discovery != nil {
		v.discoveryMu.RUnlock()
		return v.discovery, nil
	}
	v.discoveryMu.RUnlock()

	v.discoveryMu.Lock()
	defer v.discoveryMu.Unlock()

	// Double-check after acquiring write lock
	if v.discovery != nil {
		return v.discovery, nil
	}

	discoveryURL := strings.TrimSuffix(v.issuer, "/") + "/.well-known/openid-configuration"

	var discovery OIDCDiscoveryDocument
	if err := FetchJSON(ctx, v.httpClient, discoveryURL, nil, &discovery); err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}

	v.discovery = &discovery

	// Update JWKS URL if not set
	if v.jwksURL == "" {
		v.jwksURL = discovery.JwksURI
	}

	return v.discovery, nil
}

// FetchJWKS fetches the JSON Web Key Set.
func (v *OIDCValidator) FetchJWKS(ctx context.Context) (*JWKS, error) {
	if v.jwksURL == "" {
		// Try to discover
		discovery, err := v.DiscoverConfiguration(ctx)
		if err != nil {
			return nil, err
		}
		v.jwksURL = discovery.JwksURI
	}

	var jwks JWKS
	if err := FetchJSON(ctx, v.httpClient, v.jwksURL, nil, &jwks); err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return &jwks, nil
}

// GetPublicKey gets a public key by key ID.
func (v *OIDCValidator) GetPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	// Check cache
	v.keyCacheMu.RLock()
	if key, ok := v.keyCache[keyID]; ok && time.Since(v.lastFetched) < v.cacheExpiry {
		v.keyCacheMu.RUnlock()
		return key, nil
	}
	v.keyCacheMu.RUnlock()

	// Fetch JWKS
	jwks, err := v.FetchJWKS(ctx)
	if err != nil {
		return nil, err
	}

	// Update cache
	v.keyCacheMu.Lock()
	v.keyCache = make(map[string]*rsa.PublicKey)
	v.lastFetched = time.Now()

	for _, jwk := range jwks.Keys {
		if jwk.Kty == "RSA" {
			key, err := jwk.ToRSAPublicKey()
			if err == nil {
				v.keyCache[jwk.Kid] = key
			}
		}
	}
	v.keyCacheMu.Unlock()

	v.keyCacheMu.RLock()
	defer v.keyCacheMu.RUnlock()

	if key, ok := v.keyCache[keyID]; ok {
		return key, nil
	}

	return nil, fmt.Errorf("key not found: %s", keyID)
}

// ValidateIDToken validates an OIDC ID token.
func (v *OIDCValidator) ValidateIDToken(ctx context.Context, idToken string) (*IDTokenClaims, error) {
	// Split token
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Decode claims
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims IDTokenClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Validate claims
	if err := v.validateClaims(&claims); err != nil {
		return nil, err
	}

	// Verify signature
	if err := v.verifySignature(ctx, parts, &header); err != nil {
		return nil, err
	}

	return &claims, nil
}

// validateClaims validates the token claims.
func (v *OIDCValidator) validateClaims(claims *IDTokenClaims) error {
	now := time.Now().Unix()

	// Validate issuer
	if claims.Issuer != v.issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", v.issuer, claims.Issuer)
	}

	// Validate audience
	if !claims.Audience.Contains(v.clientID) {
		return fmt.Errorf("invalid audience: %v does not contain %s", claims.Audience, v.clientID)
	}

	// Validate expiration
	if claims.ExpiresAt < now {
		return fmt.Errorf("token expired at %d, current time is %d", claims.ExpiresAt, now)
	}

	// Validate issued at
	if claims.IssuedAt > now+60 { // 60 second clock skew allowance
		return fmt.Errorf("token issued in the future")
	}

	return nil
}

// verifySignature verifies the token signature.
func (v *OIDCValidator) verifySignature(ctx context.Context, parts []string, header *JWTHeader) error {
	// Get public key
	publicKey, err := v.GetPublicKey(ctx, header.KeyID)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Prepare message
	message := parts[0] + "." + parts[1]

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify based on algorithm
	var hashFunc crypto.Hash
	switch header.Algorithm {
	case "RS256":
		hashFunc = crypto.SHA256
	case "RS384":
		hashFunc = crypto.SHA384
	case "RS512":
		hashFunc = crypto.SHA512
	default:
		return fmt.Errorf("unsupported algorithm: %s", header.Algorithm)
	}

	h := hashFunc.New()
	h.Write([]byte(message))
	hashed := h.Sum(nil)

	if err := rsa.VerifyPKCS1v15(publicKey, hashFunc, hashed, signature); err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	return nil
}

// ValidateNonce validates the nonce claim.
func (v *OIDCValidator) ValidateNonce(claims *IDTokenClaims, expectedNonce string) error {
	if expectedNonce == "" {
		return nil // Nonce not required
	}

	if claims.Nonce != expectedNonce {
		return fmt.Errorf("invalid nonce: expected %s, got %s", expectedNonce, claims.Nonce)
	}

	return nil
}

// OIDCProviderFactory creates OIDC validators for different providers.
type OIDCProviderFactory struct {
	validators map[ProviderType]*OIDCValidator
	mu         sync.RWMutex
}

// NewOIDCProviderFactory creates a new OIDC provider factory.
func NewOIDCProviderFactory() *OIDCProviderFactory {
	return &OIDCProviderFactory{
		validators: make(map[ProviderType]*OIDCValidator),
	}
}

// RegisterProvider registers an OIDC provider.
func (f *OIDCProviderFactory) RegisterProvider(providerType ProviderType, config OIDCValidatorConfig) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.validators[providerType] = NewOIDCValidator(config)
}

// GetValidator gets a validator for a provider.
func (f *OIDCProviderFactory) GetValidator(providerType ProviderType) (*OIDCValidator, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	validator, ok := f.validators[providerType]
	if !ok {
		return nil, fmt.Errorf("validator not found for provider: %s", providerType)
	}

	return validator, nil
}

// DefaultOIDCProviderFactory creates a factory with default provider configurations.
func DefaultOIDCProviderFactory(googleClientID, microsoftClientID string) *OIDCProviderFactory {
	factory := NewOIDCProviderFactory()

	if googleClientID != "" {
		factory.RegisterProvider(ProviderTypeGoogle, OIDCValidatorConfig{
			Issuer:   "https://accounts.google.com",
			ClientID: googleClientID,
			JwksURL:  "https://www.googleapis.com/oauth2/v3/certs",
		})
	}

	if microsoftClientID != "" {
		factory.RegisterProvider(ProviderTypeMicrosoft, OIDCValidatorConfig{
			Issuer:   "https://login.microsoftonline.com/common/v2.0",
			ClientID: microsoftClientID,
			JwksURL:  "https://login.microsoftonline.com/common/discovery/v2.0/keys",
		})
	}

	return factory
}
