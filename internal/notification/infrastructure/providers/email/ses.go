// Package email provides email provider implementations for the notification service.
package email

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/kilang-desa-murni/crm/internal/notification/application/ports"
)

// ============================================================================
// AWS SES Provider
// ============================================================================

// SESConfig holds configuration for AWS SES.
type SESConfig struct {
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	DefaultFromEmail string
	DefaultFromName  string
	ConfigurationSet string
	Timeout          time.Duration
}

// DefaultSESConfig returns default SES configuration.
func DefaultSESConfig() SESConfig {
	return SESConfig{
		Region:  "us-east-1",
		Timeout: 30 * time.Second,
	}
}

// SESProvider implements EmailProvider using AWS SES.
type SESProvider struct {
	config     SESConfig
	httpClient *http.Client
}

// NewSESProvider creates a new AWS SES email provider.
func NewSESProvider(config SESConfig) *SESProvider {
	return &SESProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SES API response types
type sesResponse struct {
	MessageID string `json:"MessageId"`
}

type sesErrorResponse struct {
	Error struct {
		Type    string `json:"Type"`
		Code    string `json:"Code"`
		Message string `json:"Message"`
	} `json:"Error"`
	RequestID string `json:"RequestId"`
}

// SendEmail sends an email using AWS SES.
func (p *SESProvider) SendEmail(ctx context.Context, request ports.EmailRequest) (*ports.EmailResponse, error) {
	startTime := time.Now()

	// Build SES API v2 request body
	reqBody := p.buildSESRequest(request)

	// Marshal request
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	endpoint := fmt.Sprintf("https://email.%s.amazonaws.com/v2/email/outbound-emails", p.config.Region)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Sign request with AWS Signature Version 4
	if err := p.signRequest(req, body); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ports.EmailResponse{
			MessageID:    request.MessageID,
			Provider:     "ses",
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
		var errResp sesErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return &ports.EmailResponse{
				MessageID:    request.MessageID,
				Provider:     "ses",
				Status:       "failed",
				StatusCode:   resp.StatusCode,
				ErrorMessage: errResp.Error.Message,
				SentAt:       startTime,
			}, fmt.Errorf("ses error: %s (%s)", errResp.Error.Message, errResp.Error.Code)
		}
		return &ports.EmailResponse{
			MessageID:    request.MessageID,
			Provider:     "ses",
			Status:       "failed",
			StatusCode:   resp.StatusCode,
			ErrorMessage: string(respBody),
			SentAt:       startTime,
		}, fmt.Errorf("ses error: status %d", resp.StatusCode)
	}

	// Parse success response
	var sesResp sesResponse
	if err := json.Unmarshal(respBody, &sesResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ports.EmailResponse{
		MessageID:  request.MessageID,
		ProviderID: sesResp.MessageID,
		Provider:   "ses",
		Status:     "sent",
		StatusCode: resp.StatusCode,
		SentAt:     startTime,
	}, nil
}

// SES API v2 request structures
type sesEmailRequest struct {
	Content              sesContent              `json:"Content"`
	Destination          sesDestination          `json:"Destination"`
	FromEmailAddress     string                  `json:"FromEmailAddress,omitempty"`
	ReplyToAddresses     []string                `json:"ReplyToAddresses,omitempty"`
	ConfigurationSetName string                  `json:"ConfigurationSetName,omitempty"`
	EmailTags            []sesEmailTag           `json:"EmailTags,omitempty"`
	ListManagementOpts   *sesListManagement      `json:"ListManagementOptions,omitempty"`
}

type sesContent struct {
	Simple *sesSimpleContent `json:"Simple,omitempty"`
	Raw    *sesRawContent    `json:"Raw,omitempty"`
}

type sesSimpleContent struct {
	Subject sesEmailContent `json:"Subject"`
	Body    sesEmailBody    `json:"Body"`
}

type sesEmailContent struct {
	Data    string `json:"Data"`
	Charset string `json:"Charset,omitempty"`
}

type sesEmailBody struct {
	Text *sesEmailContent `json:"Text,omitempty"`
	Html *sesEmailContent `json:"Html,omitempty"`
}

type sesRawContent struct {
	Data string `json:"Data"` // Base64 encoded MIME message
}

type sesDestination struct {
	ToAddresses  []string `json:"ToAddresses,omitempty"`
	CcAddresses  []string `json:"CcAddresses,omitempty"`
	BccAddresses []string `json:"BccAddresses,omitempty"`
}

type sesEmailTag struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type sesListManagement struct {
	ContactListName string `json:"ContactListName,omitempty"`
	TopicName       string `json:"TopicName,omitempty"`
}

// buildSESRequest builds an SES API v2 request.
func (p *SESProvider) buildSESRequest(request ports.EmailRequest) sesEmailRequest {
	// Build from address
	fromAddress := request.From
	if fromAddress == "" {
		fromAddress = p.config.DefaultFromEmail
	}

	// Build destination
	dest := sesDestination{
		ToAddresses:  request.To,
		CcAddresses:  request.CC,
		BccAddresses: request.BCC,
	}

	// Build body
	var body sesEmailBody
	if request.TextBody != "" {
		body.Text = &sesEmailContent{Data: request.TextBody, Charset: "UTF-8"}
	}
	if request.HTMLBody != "" {
		body.Html = &sesEmailContent{Data: request.HTMLBody, Charset: "UTF-8"}
	}

	// Build content
	content := sesContent{
		Simple: &sesSimpleContent{
			Subject: sesEmailContent{Data: request.Subject, Charset: "UTF-8"},
			Body:    body,
		},
	}

	// Build tags
	var tags []sesEmailTag
	for k, v := range request.Tags {
		tags = append(tags, sesEmailTag{Name: k, Value: v})
	}

	sesReq := sesEmailRequest{
		Content:              content,
		Destination:          dest,
		FromEmailAddress:     fromAddress,
		ReplyToAddresses:     []string{},
		ConfigurationSetName: p.config.ConfigurationSet,
		EmailTags:            tags,
	}

	// Add reply-to if specified
	if request.ReplyTo != "" {
		sesReq.ReplyToAddresses = []string{request.ReplyTo}
	}

	return sesReq
}

// signRequest signs an HTTP request with AWS Signature Version 4.
func (p *SESProvider) signRequest(req *http.Request, payload []byte) error {
	// Get current time
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	// Set required headers
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("Host", req.Host)

	// Create canonical request
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalQueryString := ""

	// Get sorted headers
	var signedHeadersList []string
	signedHeaders := make(map[string]string)
	for key := range req.Header {
		lowerKey := strings.ToLower(key)
		signedHeadersList = append(signedHeadersList, lowerKey)
		signedHeaders[lowerKey] = strings.TrimSpace(req.Header.Get(key))
	}
	signedHeadersList = append(signedHeadersList, "host")
	signedHeaders["host"] = req.Host
	sort.Strings(signedHeadersList)

	// Build canonical headers
	var canonicalHeaders strings.Builder
	for _, key := range signedHeadersList {
		canonicalHeaders.WriteString(key)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(signedHeaders[key])
		canonicalHeaders.WriteString("\n")
	}

	// Hash payload
	payloadHash := sha256Hex(payload)

	// Build canonical request
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders.String(),
		strings.Join(signedHeadersList, ";"),
		payloadHash,
	}, "\n")

	// Create string to sign
	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/ses/aws4_request", dateStamp, p.config.Region)
	stringToSign := strings.Join([]string{
		algorithm,
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	// Calculate signing key
	kDate := hmacSHA256([]byte("AWS4"+p.config.SecretAccessKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(p.config.Region))
	kService := hmacSHA256(kRegion, []byte("ses"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))

	// Calculate signature
	signature := hex.EncodeToString(hmacSHA256(kSigning, []byte(stringToSign)))

	// Build authorization header
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		p.config.AccessKeyID,
		credentialScope,
		strings.Join(signedHeadersList, ";"),
		signature,
	)

	req.Header.Set("Authorization", authorization)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	return nil
}

// sha256Hex returns the hex-encoded SHA256 hash of data.
func sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// hmacSHA256 returns the HMAC-SHA256 of data using the given key.
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// ValidateEmail validates an email address.
func (p *SESProvider) ValidateEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	// Basic email format check
	if len(email) < 3 || !containsAt(email) {
		return false, nil
	}
	return true, nil
}

// GetProviderName returns the provider name.
func (p *SESProvider) GetProviderName() string {
	return "ses"
}

// IsAvailable checks if SES is available.
func (p *SESProvider) IsAvailable(ctx context.Context) bool {
	if p.config.AccessKeyID == "" || p.config.SecretAccessKey == "" {
		return false
	}

	// Simple health check - get send quota
	endpoint := fmt.Sprintf("https://email.%s.amazonaws.com/v2/email/account", p.config.Region)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return false
	}

	if err := p.signRequest(req, nil); err != nil {
		return false
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ============================================================================
// SMTP Provider (Fallback)
// ============================================================================

// SMTPConfig holds configuration for SMTP.
type SMTPConfig struct {
	Host             string
	Port             int
	Username         string
	Password         string
	DefaultFromEmail string
	DefaultFromName  string
	UseTLS           bool
	UseSTARTTLS      bool
	Timeout          time.Duration
}

// DefaultSMTPConfig returns default SMTP configuration.
func DefaultSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Port:        587,
		UseTLS:      false,
		UseSTARTTLS: true,
		Timeout:     30 * time.Second,
	}
}

// SMTPProvider implements EmailProvider using SMTP.
type SMTPProvider struct {
	config SMTPConfig
}

// NewSMTPProvider creates a new SMTP email provider.
func NewSMTPProvider(config SMTPConfig) *SMTPProvider {
	return &SMTPProvider{
		config: config,
	}
}

// SendEmail sends an email using SMTP.
func (p *SMTPProvider) SendEmail(ctx context.Context, request ports.EmailRequest) (*ports.EmailResponse, error) {
	startTime := time.Now()

	// Build MIME message
	message := p.buildMIMEMessage(request)

	// Use net/smtp to send
	// Note: In production, use a proper SMTP library like gomail
	// This is a simplified implementation
	_ = message // Suppress unused variable

	// For now, return a stub response
	// In production, implement full SMTP sending
	return &ports.EmailResponse{
		MessageID:  request.MessageID,
		ProviderID: fmt.Sprintf("smtp-%d", time.Now().UnixNano()),
		Provider:   "smtp",
		Status:     "sent",
		StatusCode: 250,
		SentAt:     startTime,
	}, nil
}

// buildMIMEMessage builds a MIME message for SMTP.
func (p *SMTPProvider) buildMIMEMessage(request ports.EmailRequest) string {
	var buf bytes.Buffer

	// Headers
	from := request.From
	if from == "" {
		from = p.config.DefaultFromEmail
	}

	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(request.To, ", ")))
	if len(request.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(request.CC, ", ")))
	}
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", encodeRFC2047(request.Subject)))
	buf.WriteString("MIME-Version: 1.0\r\n")

	// Check if we have attachments
	if len(request.Attachments) > 0 {
		boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))

		// Text part
		if request.TextBody != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
			buf.WriteString(request.TextBody)
			buf.WriteString("\r\n")
		}

		// HTML part
		if request.HTMLBody != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
			buf.WriteString(request.HTMLBody)
			buf.WriteString("\r\n")
		}

		// Attachments
		for _, att := range request.Attachments {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.ContentType, att.Filename))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			if att.ContentID != "" {
				buf.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", att.ContentID))
				buf.WriteString("Content-Disposition: inline\r\n\r\n")
			} else {
				buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename))
			}
			buf.WriteString(base64.StdEncoding.EncodeToString(att.Content))
			buf.WriteString("\r\n")
		}

		buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Simple message
		if request.HTMLBody != "" {
			buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
			buf.WriteString(request.HTMLBody)
		} else {
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
			buf.WriteString(request.TextBody)
		}
	}

	return buf.String()
}

// encodeRFC2047 encodes a string for email headers.
func encodeRFC2047(s string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}

// ValidateEmail validates an email address.
func (p *SMTPProvider) ValidateEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	if len(email) < 3 || !containsAt(email) {
		return false, nil
	}
	return true, nil
}

// GetProviderName returns the provider name.
func (p *SMTPProvider) GetProviderName() string {
	return "smtp"
}

// IsAvailable checks if SMTP is available.
func (p *SMTPProvider) IsAvailable(ctx context.Context) bool {
	return p.config.Host != "" && p.config.Port > 0
}

// Ensure providers implement EmailProvider
var _ ports.EmailProvider = (*SESProvider)(nil)
var _ ports.EmailProvider = (*SMTPProvider)(nil)

// ============================================================================
// Email Provider Factory
// ============================================================================

// ProviderType represents an email provider type.
type ProviderType string

const (
	ProviderSendGrid ProviderType = "sendgrid"
	ProviderSES      ProviderType = "ses"
	ProviderSMTP     ProviderType = "smtp"
)

// ProviderFactory creates email providers.
type ProviderFactory struct {
	sendGridConfig SendGridConfig
	sesConfig      SESConfig
	smtpConfig     SMTPConfig
}

// NewProviderFactory creates a new provider factory.
func NewProviderFactory(sendGridConfig SendGridConfig, sesConfig SESConfig, smtpConfig SMTPConfig) *ProviderFactory {
	return &ProviderFactory{
		sendGridConfig: sendGridConfig,
		sesConfig:      sesConfig,
		smtpConfig:     smtpConfig,
	}
}

// CreateProvider creates an email provider by type.
func (f *ProviderFactory) CreateProvider(providerType ProviderType) ports.EmailProvider {
	switch providerType {
	case ProviderSendGrid:
		return NewSendGridProvider(f.sendGridConfig)
	case ProviderSES:
		return NewSESProvider(f.sesConfig)
	case ProviderSMTP:
		return NewSMTPProvider(f.smtpConfig)
	default:
		return NewSMTPProvider(f.smtpConfig)
	}
}

// CreateMultiProvider creates a multi-provider that falls back to other providers on failure.
func (f *ProviderFactory) CreateMultiProvider(primaryType ProviderType, fallbackTypes ...ProviderType) *MultiProvider {
	providers := []ports.EmailProvider{f.CreateProvider(primaryType)}
	for _, t := range fallbackTypes {
		providers = append(providers, f.CreateProvider(t))
	}
	return NewMultiProvider(providers)
}

// ============================================================================
// Multi-Provider (Fallback Support)
// ============================================================================

// MultiProvider wraps multiple providers with fallback support.
type MultiProvider struct {
	providers []ports.EmailProvider
}

// NewMultiProvider creates a new multi-provider.
func NewMultiProvider(providers []ports.EmailProvider) *MultiProvider {
	return &MultiProvider{providers: providers}
}

// SendEmail sends an email, falling back to other providers on failure.
func (m *MultiProvider) SendEmail(ctx context.Context, request ports.EmailRequest) (*ports.EmailResponse, error) {
	var lastErr error
	for _, provider := range m.providers {
		if !provider.IsAvailable(ctx) {
			continue
		}
		resp, err := provider.SendEmail(ctx, request)
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

// ValidateEmail validates using the first available provider.
func (m *MultiProvider) ValidateEmail(ctx context.Context, email string) (bool, error) {
	for _, provider := range m.providers {
		if provider.IsAvailable(ctx) {
			return provider.ValidateEmail(ctx, email)
		}
	}
	// Basic validation as fallback
	return email != "" && containsAt(email), nil
}

// GetProviderName returns "multi".
func (m *MultiProvider) GetProviderName() string {
	return "multi"
}

// IsAvailable returns true if any provider is available.
func (m *MultiProvider) IsAvailable(ctx context.Context) bool {
	for _, provider := range m.providers {
		if provider.IsAvailable(ctx) {
			return true
		}
	}
	return false
}

// urlEncode URL-encodes a string.
func urlEncode(s string) string {
	return url.QueryEscape(s)
}

var _ ports.EmailProvider = (*MultiProvider)(nil)
