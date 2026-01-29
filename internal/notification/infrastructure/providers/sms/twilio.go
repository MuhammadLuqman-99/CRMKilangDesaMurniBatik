// Package sms provides SMS provider implementations for the notification service.
package sms

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kilang-desa-murni/crm/internal/notification/application/ports"
)

// ============================================================================
// Twilio Provider
// ============================================================================

// TwilioConfig holds configuration for Twilio.
type TwilioConfig struct {
	AccountSID        string
	AuthToken         string
	FromNumber        string
	MessagingServiceSID string
	StatusCallbackURL string
	Timeout           time.Duration
}

// DefaultTwilioConfig returns default Twilio configuration.
func DefaultTwilioConfig() TwilioConfig {
	return TwilioConfig{
		Timeout: 30 * time.Second,
	}
}

// TwilioProvider implements SMSProvider using Twilio.
type TwilioProvider struct {
	config     TwilioConfig
	httpClient *http.Client
	baseURL    string
}

// NewTwilioProvider creates a new Twilio SMS provider.
func NewTwilioProvider(config TwilioConfig) *TwilioProvider {
	return &TwilioProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL: fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s", config.AccountSID),
	}
}

// twilioMessageResponse represents a Twilio message response.
type twilioMessageResponse struct {
	SID         string `json:"sid"`
	AccountSID  string `json:"account_sid"`
	To          string `json:"to"`
	From        string `json:"from"`
	Body        string `json:"body"`
	Status      string `json:"status"`
	NumSegments string `json:"num_segments"`
	DateCreated string `json:"date_created"`
	DateSent    string `json:"date_sent"`
	ErrorCode   int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// twilioErrorResponse represents a Twilio error response.
type twilioErrorResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

// SendSMS sends an SMS using Twilio.
func (p *TwilioProvider) SendSMS(ctx context.Context, request ports.SMSRequest) (*ports.SMSResponse, error) {
	startTime := time.Now()

	// Build form data
	formData := url.Values{}
	formData.Set("To", request.To)
	formData.Set("Body", request.Body)

	// Set from number or messaging service SID
	if p.config.MessagingServiceSID != "" {
		formData.Set("MessagingServiceSid", p.config.MessagingServiceSID)
	} else {
		from := request.From
		if from == "" {
			from = p.config.FromNumber
		}
		formData.Set("From", from)
	}

	// Set status callback
	if p.config.StatusCallbackURL != "" {
		formData.Set("StatusCallback", p.config.StatusCallbackURL)
	}

	// Set scheduled time if provided
	if request.ScheduledAt != nil {
		formData.Set("SendAt", request.ScheduledAt.Format(time.RFC3339))
		formData.Set("ScheduleType", "fixed")
	}

	// Create HTTP request
	endpoint := p.baseURL + "/Messages.json"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(p.config.AccountSID+":"+p.config.AuthToken)))

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ports.SMSResponse{
			MessageID:    request.MessageID,
			Provider:     "twilio",
			Status:       "failed",
			StatusCode:   0,
			ErrorMessage: fmt.Sprintf("request failed: %v", err),
			SentAt:       startTime,
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)

	// Check response status
	if resp.StatusCode >= 400 {
		var errResp twilioErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Message != "" {
			return &ports.SMSResponse{
				MessageID:    request.MessageID,
				Provider:     "twilio",
				Status:       "failed",
				StatusCode:   resp.StatusCode,
				ErrorMessage: errResp.Message,
				SentAt:       startTime,
			}, fmt.Errorf("twilio error: %s (code: %d)", errResp.Message, errResp.Code)
		}
		return &ports.SMSResponse{
			MessageID:    request.MessageID,
			Provider:     "twilio",
			Status:       "failed",
			StatusCode:   resp.StatusCode,
			ErrorMessage: string(respBody),
			SentAt:       startTime,
		}, fmt.Errorf("twilio error: status %d", resp.StatusCode)
	}

	// Parse success response
	var msgResp twilioMessageResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse segment count
	segmentCount := 1
	if msgResp.NumSegments != "" {
		fmt.Sscanf(msgResp.NumSegments, "%d", &segmentCount)
	}

	return &ports.SMSResponse{
		MessageID:    request.MessageID,
		ProviderID:   msgResp.SID,
		Provider:     "twilio",
		Status:       msgResp.Status,
		StatusCode:   resp.StatusCode,
		SegmentCount: segmentCount,
		SentAt:       startTime,
	}, nil
}

// ValidatePhoneNumber validates a phone number using Twilio Lookup API.
func (p *TwilioProvider) ValidatePhoneNumber(ctx context.Context, phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}

	// Use Twilio Lookup API
	endpoint := fmt.Sprintf("https://lookups.twilio.com/v2/PhoneNumbers/%s", url.PathEscape(phone))
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(p.config.AccountSID+":"+p.config.AuthToken)))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		// Fallback to basic validation
		return isValidPhoneFormat(phone), nil
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// isValidPhoneFormat performs basic phone number format validation.
func isValidPhoneFormat(phone string) bool {
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}
	// Remove common formatting characters
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' || r == '+' {
			return r
		}
		return -1
	}, phone)
	return len(cleaned) >= 10
}

// GetProviderName returns the provider name.
func (p *TwilioProvider) GetProviderName() string {
	return "twilio"
}

// IsAvailable checks if Twilio is available.
func (p *TwilioProvider) IsAvailable(ctx context.Context) bool {
	if p.config.AccountSID == "" || p.config.AuthToken == "" {
		return false
	}

	// Check account status
	endpoint := p.baseURL + ".json"
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(p.config.AccountSID+":"+p.config.AuthToken)))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetDeliveryStatus retrieves the delivery status for a message.
func (p *TwilioProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*ports.SMSDeliveryStatus, error) {
	endpoint := fmt.Sprintf("%s/Messages/%s.json", p.baseURL, messageID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(p.config.AccountSID+":"+p.config.AuthToken)))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get status: HTTP %d", resp.StatusCode)
	}

	var msgResp twilioMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	status := &ports.SMSDeliveryStatus{
		MessageID:  messageID,
		ProviderID: msgResp.SID,
		Status:     msgResp.Status,
	}

	// Parse delivered time if available
	if msgResp.DateSent != "" {
		if t, err := time.Parse(time.RFC1123Z, msgResp.DateSent); err == nil {
			status.DeliveredAt = &t
		}
	}

	// Add error info if present
	if msgResp.ErrorCode != 0 {
		status.ErrorCode = fmt.Sprintf("%d", msgResp.ErrorCode)
		status.ErrorMessage = msgResp.ErrorMessage
	}

	return status, nil
}

// Ensure TwilioProvider implements SMSProvider
var _ ports.SMSProvider = (*TwilioProvider)(nil)

// ============================================================================
// Vonage (Nexmo) Provider
// ============================================================================

// VonageConfig holds configuration for Vonage (Nexmo).
type VonageConfig struct {
	APIKey       string
	APISecret    string
	FromNumber   string
	CallbackURL  string
	Timeout      time.Duration
}

// DefaultVonageConfig returns default Vonage configuration.
func DefaultVonageConfig() VonageConfig {
	return VonageConfig{
		Timeout: 30 * time.Second,
	}
}

// VonageProvider implements SMSProvider using Vonage.
type VonageProvider struct {
	config     VonageConfig
	httpClient *http.Client
}

// NewVonageProvider creates a new Vonage SMS provider.
func NewVonageProvider(config VonageConfig) *VonageProvider {
	return &VonageProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// vonageRequest represents a Vonage SMS request.
type vonageRequest struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	From      string `json:"from"`
	To        string `json:"to"`
	Text      string `json:"text"`
	Type      string `json:"type,omitempty"`
	Callback  string `json:"callback,omitempty"`
}

// vonageResponse represents a Vonage SMS response.
type vonageResponse struct {
	Messages []vonageMessage `json:"messages"`
}

type vonageMessage struct {
	Status          string `json:"status"`
	MessageID       string `json:"message-id"`
	To              string `json:"to"`
	ClientRef       string `json:"client-ref,omitempty"`
	RemainingBalance string `json:"remaining-balance"`
	MessagePrice    string `json:"message-price"`
	Network         string `json:"network"`
	ErrorText       string `json:"error-text,omitempty"`
}

// SendSMS sends an SMS using Vonage.
func (p *VonageProvider) SendSMS(ctx context.Context, request ports.SMSRequest) (*ports.SMSResponse, error) {
	startTime := time.Now()

	// Build request
	vonageReq := vonageRequest{
		APIKey:    p.config.APIKey,
		APISecret: p.config.APISecret,
		From:      p.config.FromNumber,
		To:        request.To,
		Text:      request.Body,
	}

	if request.From != "" {
		vonageReq.From = request.From
	}

	if request.Unicode {
		vonageReq.Type = "unicode"
	}

	if p.config.CallbackURL != "" {
		vonageReq.Callback = p.config.CallbackURL
	}

	// Marshal request
	body, err := json.Marshal(vonageReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://rest.nexmo.com/sms/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ports.SMSResponse{
			MessageID:    request.MessageID,
			Provider:     "vonage",
			Status:       "failed",
			StatusCode:   0,
			ErrorMessage: fmt.Sprintf("request failed: %v", err),
			SentAt:       startTime,
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)

	// Parse response
	var vonageResp vonageResponse
	if err := json.Unmarshal(respBody, &vonageResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(vonageResp.Messages) == 0 {
		return &ports.SMSResponse{
			MessageID:    request.MessageID,
			Provider:     "vonage",
			Status:       "failed",
			StatusCode:   resp.StatusCode,
			ErrorMessage: "no messages in response",
			SentAt:       startTime,
		}, fmt.Errorf("vonage error: no messages in response")
	}

	msg := vonageResp.Messages[0]
	if msg.Status != "0" {
		return &ports.SMSResponse{
			MessageID:    request.MessageID,
			Provider:     "vonage",
			Status:       "failed",
			StatusCode:   resp.StatusCode,
			ErrorMessage: msg.ErrorText,
			SentAt:       startTime,
		}, fmt.Errorf("vonage error: %s (status: %s)", msg.ErrorText, msg.Status)
	}

	return &ports.SMSResponse{
		MessageID:    request.MessageID,
		ProviderID:   msg.MessageID,
		Provider:     "vonage",
		Status:       "sent",
		StatusCode:   resp.StatusCode,
		SegmentCount: 1,
		SentAt:       startTime,
	}, nil
}

// ValidatePhoneNumber validates a phone number.
func (p *VonageProvider) ValidatePhoneNumber(ctx context.Context, phone string) (bool, error) {
	return isValidPhoneFormat(phone), nil
}

// GetProviderName returns the provider name.
func (p *VonageProvider) GetProviderName() string {
	return "vonage"
}

// IsAvailable checks if Vonage is available.
func (p *VonageProvider) IsAvailable(ctx context.Context) bool {
	return p.config.APIKey != "" && p.config.APISecret != ""
}

// GetDeliveryStatus retrieves the delivery status for a message.
func (p *VonageProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*ports.SMSDeliveryStatus, error) {
	// Vonage delivery status is typically received via callback
	return nil, fmt.Errorf("delivery status not implemented for vonage")
}

// Ensure VonageProvider implements SMSProvider
var _ ports.SMSProvider = (*VonageProvider)(nil)

// ============================================================================
// SMS Provider Factory
// ============================================================================

// SMSProviderType represents an SMS provider type.
type SMSProviderType string

const (
	ProviderTwilio SMSProviderType = "twilio"
	ProviderVonage SMSProviderType = "vonage"
)

// SMSProviderFactory creates SMS providers.
type SMSProviderFactory struct {
	twilioConfig TwilioConfig
	vonageConfig VonageConfig
}

// NewSMSProviderFactory creates a new SMS provider factory.
func NewSMSProviderFactory(twilioConfig TwilioConfig, vonageConfig VonageConfig) *SMSProviderFactory {
	return &SMSProviderFactory{
		twilioConfig: twilioConfig,
		vonageConfig: vonageConfig,
	}
}

// CreateProvider creates an SMS provider by type.
func (f *SMSProviderFactory) CreateProvider(providerType SMSProviderType) ports.SMSProvider {
	switch providerType {
	case ProviderTwilio:
		return NewTwilioProvider(f.twilioConfig)
	case ProviderVonage:
		return NewVonageProvider(f.vonageConfig)
	default:
		return NewTwilioProvider(f.twilioConfig)
	}
}

// CreateMultiProvider creates a multi-provider that falls back to other providers on failure.
func (f *SMSProviderFactory) CreateMultiProvider(primaryType SMSProviderType, fallbackTypes ...SMSProviderType) *MultiSMSProvider {
	providers := []ports.SMSProvider{f.CreateProvider(primaryType)}
	for _, t := range fallbackTypes {
		providers = append(providers, f.CreateProvider(t))
	}
	return NewMultiSMSProvider(providers)
}

// ============================================================================
// Multi-Provider (Fallback Support)
// ============================================================================

// MultiSMSProvider wraps multiple SMS providers with fallback support.
type MultiSMSProvider struct {
	providers []ports.SMSProvider
}

// NewMultiSMSProvider creates a new multi-provider.
func NewMultiSMSProvider(providers []ports.SMSProvider) *MultiSMSProvider {
	return &MultiSMSProvider{providers: providers}
}

// SendSMS sends an SMS, falling back to other providers on failure.
func (m *MultiSMSProvider) SendSMS(ctx context.Context, request ports.SMSRequest) (*ports.SMSResponse, error) {
	var lastErr error
	for _, provider := range m.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}
		resp, err := provider.SendSMS(ctx, request)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed: %w", lastErr)
	}
	return nil, fmt.Errorf("no available providers")
}

// ValidatePhoneNumber validates using the first available provider.
func (m *MultiSMSProvider) ValidatePhoneNumber(ctx context.Context, phone string) (bool, error) {
	for _, provider := range m.providers {
		if provider.IsAvailable(ctx) {
			return provider.ValidatePhoneNumber(ctx, phone)
		}
	}
	return isValidPhoneFormat(phone), nil
}

// GetProviderName returns "multi".
func (m *MultiSMSProvider) GetProviderName() string {
	return "multi"
}

// IsAvailable returns true if any provider is available.
func (m *MultiSMSProvider) IsAvailable(ctx context.Context) bool {
	for _, provider := range m.providers {
		if provider.IsAvailable(ctx) {
			return true
		}
	}
	return false
}

// GetDeliveryStatus retrieves delivery status from the first provider that supports it.
func (m *MultiSMSProvider) GetDeliveryStatus(ctx context.Context, messageID string) (*ports.SMSDeliveryStatus, error) {
	for _, provider := range m.providers {
		if provider.IsAvailable(ctx) {
			status, err := provider.GetDeliveryStatus(ctx, messageID)
			if err == nil {
				return status, nil
			}
		}
	}
	return nil, fmt.Errorf("no provider could retrieve delivery status")
}

var _ ports.SMSProvider = (*MultiSMSProvider)(nil)
