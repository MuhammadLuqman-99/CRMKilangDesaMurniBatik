// Package email provides email provider implementations for the notification service.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kilang-desa-murni/crm/internal/notification/application/ports"
)

// ============================================================================
// SendGrid Provider
// ============================================================================

// SendGridConfig holds configuration for SendGrid.
type SendGridConfig struct {
	APIKey           string
	BaseURL          string
	DefaultFromEmail string
	DefaultFromName  string
	Sandbox          bool
	Timeout          time.Duration
}

// DefaultSendGridConfig returns default SendGrid configuration.
func DefaultSendGridConfig() SendGridConfig {
	return SendGridConfig{
		BaseURL: "https://api.sendgrid.com/v3",
		Timeout: 30 * time.Second,
		Sandbox: false,
	}
}

// SendGridProvider implements EmailProvider using SendGrid.
type SendGridProvider struct {
	config     SendGridConfig
	httpClient *http.Client
}

// NewSendGridProvider creates a new SendGrid email provider.
func NewSendGridProvider(config SendGridConfig) *SendGridProvider {
	return &SendGridProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// sendGridRequest represents a SendGrid API mail send request.
type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	ReplyTo          *sendGridEmail            `json:"reply_to,omitempty"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
	Attachments      []sendGridAttachment      `json:"attachments,omitempty"`
	Headers          map[string]string         `json:"headers,omitempty"`
	Categories       []string                  `json:"categories,omitempty"`
	CustomArgs       map[string]string         `json:"custom_args,omitempty"`
	SendAt           int64                     `json:"send_at,omitempty"`
	TrackingSettings *sendGridTracking         `json:"tracking_settings,omitempty"`
	MailSettings     *sendGridMailSettings     `json:"mail_settings,omitempty"`
}

type sendGridPersonalization struct {
	To      []sendGridEmail   `json:"to"`
	CC      []sendGridEmail   `json:"cc,omitempty"`
	BCC     []sendGridEmail   `json:"bcc,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sendGridAttachment struct {
	Content     string `json:"content"` // Base64 encoded
	Type        string `json:"type"`
	Filename    string `json:"filename"`
	Disposition string `json:"disposition,omitempty"`
	ContentID   string `json:"content_id,omitempty"`
}

type sendGridTracking struct {
	ClickTracking        *sendGridClickTracking `json:"click_tracking,omitempty"`
	OpenTracking         *sendGridOpenTracking  `json:"open_tracking,omitempty"`
	SubscriptionTracking *sendGridSubTracking   `json:"subscription_tracking,omitempty"`
}

type sendGridClickTracking struct {
	Enable     bool `json:"enable"`
	EnableText bool `json:"enable_text"`
}

type sendGridOpenTracking struct {
	Enable bool `json:"enable"`
}

type sendGridSubTracking struct {
	Enable bool `json:"enable"`
}

type sendGridMailSettings struct {
	SandboxMode *sendGridSandbox `json:"sandbox_mode,omitempty"`
}

type sendGridSandbox struct {
	Enable bool `json:"enable"`
}

// sendGridResponse represents a SendGrid API response.
type sendGridResponse struct {
	MessageID string `json:"-"` // Extracted from headers
}

// sendGridErrorResponse represents a SendGrid API error response.
type sendGridErrorResponse struct {
	Errors []sendGridError `json:"errors"`
}

type sendGridError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Help    string `json:"help"`
}

// SendEmail sends an email using SendGrid.
func (p *SendGridProvider) SendEmail(ctx context.Context, request ports.EmailRequest) (*ports.EmailResponse, error) {
	// Build SendGrid request
	sgRequest := p.buildSendGridRequest(request)

	// Marshal request
	body, err := json.Marshal(sgRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/mail/send", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	startTime := time.Now()
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &ports.EmailResponse{
			MessageID:    request.MessageID,
			Provider:     "sendgrid",
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
		var errResp sendGridErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && len(errResp.Errors) > 0 {
			return &ports.EmailResponse{
				MessageID:    request.MessageID,
				Provider:     "sendgrid",
				Status:       "failed",
				StatusCode:   resp.StatusCode,
				ErrorMessage: errResp.Errors[0].Message,
				SentAt:       startTime,
			}, fmt.Errorf("sendgrid error: %s", errResp.Errors[0].Message)
		}
		return &ports.EmailResponse{
			MessageID:    request.MessageID,
			Provider:     "sendgrid",
			Status:       "failed",
			StatusCode:   resp.StatusCode,
			ErrorMessage: string(respBody),
			SentAt:       startTime,
		}, fmt.Errorf("sendgrid error: status %d", resp.StatusCode)
	}

	// Extract message ID from headers
	providerID := resp.Header.Get("X-Message-Id")

	return &ports.EmailResponse{
		MessageID:  request.MessageID,
		ProviderID: providerID,
		Provider:   "sendgrid",
		Status:     "sent",
		StatusCode: resp.StatusCode,
		SentAt:     startTime,
	}, nil
}

// buildSendGridRequest builds a SendGrid API request.
func (p *SendGridProvider) buildSendGridRequest(request ports.EmailRequest) sendGridRequest {
	// Build recipients
	toEmails := make([]sendGridEmail, len(request.To))
	for i, email := range request.To {
		toEmails[i] = sendGridEmail{Email: email}
	}

	ccEmails := make([]sendGridEmail, len(request.CC))
	for i, email := range request.CC {
		ccEmails[i] = sendGridEmail{Email: email}
	}

	bccEmails := make([]sendGridEmail, len(request.BCC))
	for i, email := range request.BCC {
		bccEmails[i] = sendGridEmail{Email: email}
	}

	personalization := sendGridPersonalization{
		To:  toEmails,
		CC:  ccEmails,
		BCC: bccEmails,
	}

	// Build from
	from := sendGridEmail{Email: request.From}
	if p.config.DefaultFromEmail != "" && from.Email == "" {
		from.Email = p.config.DefaultFromEmail
		from.Name = p.config.DefaultFromName
	}

	// Build content
	var content []sendGridContent
	if request.TextBody != "" {
		content = append(content, sendGridContent{Type: "text/plain", Value: request.TextBody})
	}
	if request.HTMLBody != "" {
		content = append(content, sendGridContent{Type: "text/html", Value: request.HTMLBody})
	}

	// Build attachments
	var attachments []sendGridAttachment
	for _, att := range request.Attachments {
		disposition := "attachment"
		if att.ContentID != "" {
			disposition = "inline"
		}
		attachments = append(attachments, sendGridAttachment{
			Content:     string(att.Content), // Should be base64 encoded
			Type:        att.ContentType,
			Filename:    att.Filename,
			Disposition: disposition,
			ContentID:   att.ContentID,
		})
	}

	// Build tracking settings
	tracking := &sendGridTracking{
		OpenTracking:  &sendGridOpenTracking{Enable: request.TrackOpens},
		ClickTracking: &sendGridClickTracking{Enable: request.TrackClicks, EnableText: request.TrackClicks},
	}

	// Build mail settings
	var mailSettings *sendGridMailSettings
	if p.config.Sandbox {
		mailSettings = &sendGridMailSettings{
			SandboxMode: &sendGridSandbox{Enable: true},
		}
	}

	sgRequest := sendGridRequest{
		Personalizations: []sendGridPersonalization{personalization},
		From:             from,
		Subject:          request.Subject,
		Content:          content,
		Attachments:      attachments,
		Headers:          request.Headers,
		CustomArgs:       request.Tags,
		TrackingSettings: tracking,
		MailSettings:     mailSettings,
	}

	// Set reply-to
	if request.ReplyTo != "" {
		sgRequest.ReplyTo = &sendGridEmail{Email: request.ReplyTo}
	}

	// Set scheduled time
	if request.ScheduledAt != nil {
		sgRequest.SendAt = request.ScheduledAt.Unix()
	}

	return sgRequest
}

// ValidateEmail validates an email address using SendGrid.
func (p *SendGridProvider) ValidateEmail(ctx context.Context, email string) (bool, error) {
	// SendGrid email validation is a separate API endpoint
	// For simplicity, we'll do basic validation here
	if email == "" {
		return false, nil
	}
	// Basic email format check
	if len(email) < 3 || !containsAt(email) {
		return false, nil
	}
	return true, nil
}

// containsAt checks if a string contains @.
func containsAt(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}

// GetProviderName returns the provider name.
func (p *SendGridProvider) GetProviderName() string {
	return "sendgrid"
}

// IsAvailable checks if SendGrid is available.
func (p *SendGridProvider) IsAvailable(ctx context.Context) bool {
	if p.config.APIKey == "" {
		return false
	}

	// Simple health check - verify API key works
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/scopes", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Ensure SendGridProvider implements EmailProvider
var _ ports.EmailProvider = (*SendGridProvider)(nil)
