// Package domain contains the domain layer for the Notification service.
package domain

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

// NotificationTemplate represents a reusable notification template.
type NotificationTemplate struct {
	BaseAggregateRoot

	// Core identification
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Code     string    `json:"code" db:"code"` // Unique template code (e.g., "welcome_email", "order_confirmation")

	// Template metadata
	Name        string           `json:"name" db:"name"`
	Description string           `json:"description,omitempty" db:"description"`
	Type        NotificationType `json:"type" db:"type"`
	Category    string           `json:"category,omitempty" db:"category"` // For organization

	// Supported channels
	Channels []NotificationChannel `json:"channels" db:"-"`

	// Template content per channel
	EmailTemplate *EmailTemplateContent `json:"email_template,omitempty" db:"-"`
	SMSTemplate   *SMSTemplateContent   `json:"sms_template,omitempty" db:"-"`
	PushTemplate  *PushTemplateContent  `json:"push_template,omitempty" db:"-"`
	InAppTemplate *InAppTemplateContent `json:"in_app_template,omitempty" db:"-"`

	// Template variables
	Variables []TemplateVariable `json:"variables" db:"-"`

	// Localization
	DefaultLocale string                                    `json:"default_locale" db:"default_locale"`
	Localizations map[string]*TemplateLocalization `json:"localizations,omitempty" db:"-"` // locale -> localized content

	// Rendering settings
	RenderEngine string `json:"render_engine" db:"render_engine"` // "go-template", "handlebars", etc.

	// Template status
	IsActive    bool `json:"is_active" db:"is_active"`
	IsDefault   bool `json:"is_default" db:"is_default"` // Default template for type
	IsLocked    bool `json:"is_locked" db:"is_locked"`   // System template, cannot be edited

	// Usage tracking
	UsageCount   int64      `json:"usage_count" db:"usage_count"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`

	// Versioning
	TemplateVersion int       `json:"template_version" db:"template_version"`
	PublishedAt     *time.Time `json:"published_at,omitempty" db:"published_at"`
	DraftContent    *TemplateDraft `json:"draft_content,omitempty" db:"-"`

	// Audit
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by,omitempty" db:"updated_by"`

	// Tags for organization
	Tags []string `json:"tags,omitempty" db:"-"`
}

// EmailTemplateContent holds email-specific template content.
type EmailTemplateContent struct {
	Subject       string            `json:"subject"`
	Body          string            `json:"body"`           // Plain text body
	HTMLBody      string            `json:"html_body"`      // HTML body
	PreviewText   string            `json:"preview_text,omitempty"` // Email preview text
	FromName      string            `json:"from_name,omitempty"`
	FromAddress   string            `json:"from_address,omitempty"`
	ReplyTo       string            `json:"reply_to,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Attachments   []TemplateAttachment `json:"attachments,omitempty"`
	TrackOpens    bool              `json:"track_opens"`
	TrackClicks   bool              `json:"track_clicks"`
}

// SMSTemplateContent holds SMS-specific template content.
type SMSTemplateContent struct {
	Body      string `json:"body"`
	SenderID  string `json:"sender_id,omitempty"`
	MaxLength int    `json:"max_length,omitempty"` // Character limit
}

// PushTemplateContent holds push notification-specific template content.
type PushTemplateContent struct {
	Title        string                 `json:"title"`
	Body         string                 `json:"body"`
	ImageURL     string                 `json:"image_url,omitempty"`
	IconURL      string                 `json:"icon_url,omitempty"`
	ActionURL    string                 `json:"action_url,omitempty"`
	Actions      []PushAction           `json:"actions,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Sound        string                 `json:"sound,omitempty"`
	Badge        *int                   `json:"badge,omitempty"`
	TTL          int                    `json:"ttl,omitempty"` // Time to live in seconds
	CollapseKey  string                 `json:"collapse_key,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
	ChannelID    string                 `json:"channel_id,omitempty"` // Android notification channel
}

// PushAction represents an action button on a push notification.
type PushAction struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Action string `json:"action"` // URL or deep link
	Icon   string `json:"icon,omitempty"`
}

// InAppTemplateContent holds in-app notification-specific template content.
type InAppTemplateContent struct {
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	IconURL     string                 `json:"icon_url,omitempty"`
	ImageURL    string                 `json:"image_url,omitempty"`
	ActionURL   string                 `json:"action_url,omitempty"`
	ActionLabel string                 `json:"action_label,omitempty"`
	Style       string                 `json:"style,omitempty"` // "default", "banner", "modal"
	Duration    int                    `json:"duration,omitempty"` // Display duration in seconds
	Data        map[string]interface{} `json:"data,omitempty"`
	Dismissable bool                   `json:"dismissable"`
}

// TemplateAttachment represents a template attachment configuration.
type TemplateAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	URL         string `json:"url,omitempty"`        // URL to fetch attachment
	TemplateKey string `json:"template_key,omitempty"` // Variable key containing attachment
	Inline      bool   `json:"inline"`
	ContentID   string `json:"content_id,omitempty"`
}

// TemplateLocalization holds localized template content.
type TemplateLocalization struct {
	Locale        string                `json:"locale"`
	EmailTemplate *EmailTemplateContent `json:"email_template,omitempty"`
	SMSTemplate   *SMSTemplateContent   `json:"sms_template,omitempty"`
	PushTemplate  *PushTemplateContent  `json:"push_template,omitempty"`
	InAppTemplate *InAppTemplateContent `json:"in_app_template,omitempty"`
}

// TemplateDraft holds draft changes before publishing.
type TemplateDraft struct {
	EmailTemplate *EmailTemplateContent `json:"email_template,omitempty"`
	SMSTemplate   *SMSTemplateContent   `json:"sms_template,omitempty"`
	PushTemplate  *PushTemplateContent  `json:"push_template,omitempty"`
	InAppTemplate *InAppTemplateContent `json:"in_app_template,omitempty"`
	Variables     []TemplateVariable    `json:"variables,omitempty"`
	UpdatedAt     time.Time             `json:"updated_at"`
	UpdatedBy     *uuid.UUID            `json:"updated_by,omitempty"`
}

// NewNotificationTemplate creates a new notification template.
func NewNotificationTemplate(
	tenantID uuid.UUID,
	code, name string,
	notificationType NotificationType,
) (*NotificationTemplate, error) {
	// Validate inputs
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)

	if code == "" {
		return nil, NewValidationError("code", "template code is required", "REQUIRED")
	}
	if len(code) > 100 {
		return nil, NewValidationError("code", "template code too long (max 100 characters)", "TOO_LONG")
	}
	if !isValidTemplateCode(code) {
		return nil, NewValidationError("code", "template code must contain only lowercase letters, numbers, and underscores", "INVALID_FORMAT")
	}
	if name == "" {
		return nil, ErrTemplateNameRequired
	}
	if len(name) > 255 {
		return nil, ErrTemplateNameTooLong
	}
	if !notificationType.IsValid() {
		return nil, NewValidationError("type", "invalid notification type", "INVALID_TYPE")
	}

	t := &NotificationTemplate{
		BaseAggregateRoot: NewBaseAggregateRoot(),
		TenantID:          tenantID,
		Code:              code,
		Name:              name,
		Type:              notificationType,
		Channels:          make([]NotificationChannel, 0),
		Variables:         make([]TemplateVariable, 0),
		DefaultLocale:     "en",
		Localizations:     make(map[string]*TemplateLocalization),
		RenderEngine:      "go-template",
		IsActive:          true,
		TemplateVersion:   1,
		Tags:              make([]string, 0),
	}

	return t, nil
}

// isValidTemplateCode validates the template code format.
func isValidTemplateCode(code string) bool {
	return regexp.MustCompile(`^[a-z][a-z0-9_]*$`).MatchString(code)
}

// ============================================================================
// Template Content Methods
// ============================================================================

// SetEmailTemplate sets the email template content.
func (t *NotificationTemplate) SetEmailTemplate(content *EmailTemplateContent) error {
	if content.Subject == "" {
		return NewValidationError("subject", "email subject is required", "REQUIRED")
	}
	if content.Body == "" && content.HTMLBody == "" {
		return NewValidationError("body", "email body is required", "REQUIRED")
	}

	t.EmailTemplate = content
	if !t.hasChannel(ChannelEmail) {
		t.Channels = append(t.Channels, ChannelEmail)
	}
	t.MarkUpdated()
	return nil
}

// SetSMSTemplate sets the SMS template content.
func (t *NotificationTemplate) SetSMSTemplate(content *SMSTemplateContent) error {
	if content.Body == "" {
		return ErrSMSBodyRequired
	}
	if content.MaxLength > 0 && len(content.Body) > content.MaxLength {
		return ErrSMSBodyTooLong
	}

	t.SMSTemplate = content
	if !t.hasChannel(ChannelSMS) {
		t.Channels = append(t.Channels, ChannelSMS)
	}
	t.MarkUpdated()
	return nil
}

// SetPushTemplate sets the push notification template content.
func (t *NotificationTemplate) SetPushTemplate(content *PushTemplateContent) error {
	if content.Title == "" && content.Body == "" {
		return NewValidationError("content", "push notification requires title or body", "REQUIRED")
	}

	t.PushTemplate = content
	if !t.hasChannel(ChannelPush) {
		t.Channels = append(t.Channels, ChannelPush)
	}
	t.MarkUpdated()
	return nil
}

// SetInAppTemplate sets the in-app notification template content.
func (t *NotificationTemplate) SetInAppTemplate(content *InAppTemplateContent) error {
	if content.Title == "" && content.Body == "" {
		return NewValidationError("content", "in-app notification requires title or body", "REQUIRED")
	}

	t.InAppTemplate = content
	if !t.hasChannel(ChannelInApp) {
		t.Channels = append(t.Channels, ChannelInApp)
	}
	t.MarkUpdated()
	return nil
}

// hasChannel checks if the template supports a channel.
func (t *NotificationTemplate) hasChannel(channel NotificationChannel) bool {
	for _, c := range t.Channels {
		if c == channel {
			return true
		}
	}
	return false
}

// SupportsChannel returns true if the template supports the given channel.
func (t *NotificationTemplate) SupportsChannel(channel NotificationChannel) bool {
	return t.hasChannel(channel)
}

// ============================================================================
// Variable Methods
// ============================================================================

// AddVariable adds a template variable.
func (t *NotificationTemplate) AddVariable(variable TemplateVariable) error {
	// Check for duplicate
	for _, v := range t.Variables {
		if v.Name == variable.Name {
			return NewValidationError("name", "variable with this name already exists", "DUPLICATE")
		}
	}
	t.Variables = append(t.Variables, variable)
	t.MarkUpdated()
	return nil
}

// RemoveVariable removes a template variable.
func (t *NotificationTemplate) RemoveVariable(name string) {
	for i, v := range t.Variables {
		if v.Name == name {
			t.Variables = append(t.Variables[:i], t.Variables[i+1:]...)
			t.MarkUpdated()
			return
		}
	}
}

// GetVariable gets a variable by name.
func (t *NotificationTemplate) GetVariable(name string) *TemplateVariable {
	for i := range t.Variables {
		if t.Variables[i].Name == name {
			return &t.Variables[i]
		}
	}
	return nil
}

// GetRequiredVariables returns all required variables.
func (t *NotificationTemplate) GetRequiredVariables() []TemplateVariable {
	var required []TemplateVariable
	for _, v := range t.Variables {
		if v.Required {
			required = append(required, v)
		}
	}
	return required
}

// ValidateVariables validates that all required variables are provided.
func (t *NotificationTemplate) ValidateVariables(data map[string]interface{}) error {
	var missing []string
	for _, v := range t.Variables {
		if v.Required {
			if _, ok := data[v.Name]; !ok {
				missing = append(missing, v.Name)
			}
		}
	}
	if len(missing) > 0 {
		return NewTemplateError(t.Code, "missing required variables", "MISSING_VARIABLES").WithMissingVars(missing)
	}
	return nil
}

// ============================================================================
// Localization Methods
// ============================================================================

// AddLocalization adds a localized version of the template.
func (t *NotificationTemplate) AddLocalization(locale string, localization *TemplateLocalization) error {
	if locale == "" {
		return NewValidationError("locale", "locale is required", "REQUIRED")
	}
	localization.Locale = locale
	if t.Localizations == nil {
		t.Localizations = make(map[string]*TemplateLocalization)
	}
	t.Localizations[locale] = localization
	t.MarkUpdated()
	return nil
}

// RemoveLocalization removes a localized version.
func (t *NotificationTemplate) RemoveLocalization(locale string) {
	if t.Localizations != nil {
		delete(t.Localizations, locale)
		t.MarkUpdated()
	}
}

// GetLocalization gets a localized version or falls back to default.
func (t *NotificationTemplate) GetLocalization(locale string) *TemplateLocalization {
	if t.Localizations != nil {
		if loc, ok := t.Localizations[locale]; ok {
			return loc
		}
	}
	return nil
}

// ============================================================================
// Status Methods
// ============================================================================

// Activate activates the template.
func (t *NotificationTemplate) Activate() {
	t.IsActive = true
	t.MarkUpdated()
	t.IncrementVersion()
	t.AddDomainEvent(NewTemplateActivatedEvent(t))
}

// Deactivate deactivates the template.
func (t *NotificationTemplate) Deactivate() {
	t.IsActive = false
	t.MarkUpdated()
	t.IncrementVersion()
	t.AddDomainEvent(NewTemplateDeactivatedEvent(t))
}

// SetAsDefault sets this template as the default for its type.
func (t *NotificationTemplate) SetAsDefault() {
	t.IsDefault = true
	t.MarkUpdated()
}

// ClearDefault clears the default flag.
func (t *NotificationTemplate) ClearDefault() {
	t.IsDefault = false
	t.MarkUpdated()
}

// Lock locks the template to prevent modifications.
func (t *NotificationTemplate) Lock() {
	t.IsLocked = true
	t.MarkUpdated()
}

// ============================================================================
// Draft/Publish Methods
// ============================================================================

// SaveDraft saves current changes as a draft.
func (t *NotificationTemplate) SaveDraft(updatedBy *uuid.UUID) {
	t.DraftContent = &TemplateDraft{
		EmailTemplate: t.EmailTemplate,
		SMSTemplate:   t.SMSTemplate,
		PushTemplate:  t.PushTemplate,
		InAppTemplate: t.InAppTemplate,
		Variables:     t.Variables,
		UpdatedAt:     time.Now().UTC(),
		UpdatedBy:     updatedBy,
	}
	t.MarkUpdated()
}

// Publish publishes the template, incrementing the version.
func (t *NotificationTemplate) Publish() {
	now := time.Now().UTC()
	t.PublishedAt = &now
	t.TemplateVersion++
	t.DraftContent = nil
	t.MarkUpdated()
	t.IncrementVersion()
	t.AddDomainEvent(NewTemplatePublishedEvent(t))
}

// DiscardDraft discards the draft changes.
func (t *NotificationTemplate) DiscardDraft() {
	t.DraftContent = nil
	t.MarkUpdated()
}

// HasDraft returns true if there are unsaved draft changes.
func (t *NotificationTemplate) HasDraft() bool {
	return t.DraftContent != nil
}

// ============================================================================
// Usage Tracking Methods
// ============================================================================

// RecordUsage records that the template was used.
func (t *NotificationTemplate) RecordUsage() {
	t.UsageCount++
	now := time.Now().UTC()
	t.LastUsedAt = &now
	t.MarkUpdated()
}

// ============================================================================
// Tag Methods
// ============================================================================

// AddTag adds a tag.
func (t *NotificationTemplate) AddTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return
	}
	for _, existing := range t.Tags {
		if existing == tag {
			return
		}
	}
	t.Tags = append(t.Tags, tag)
	t.MarkUpdated()
}

// RemoveTag removes a tag.
func (t *NotificationTemplate) RemoveTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for i, existing := range t.Tags {
		if existing == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			t.MarkUpdated()
			return
		}
	}
}

// HasTag checks if the template has a tag.
func (t *NotificationTemplate) HasTag(tag string) bool {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for _, existing := range t.Tags {
		if existing == tag {
			return true
		}
	}
	return false
}

// ============================================================================
// Rendering Methods
// ============================================================================

// RenderEmail renders the email template with the provided data.
func (t *NotificationTemplate) RenderEmail(data map[string]interface{}, locale string) (*RenderedEmail, error) {
	var content *EmailTemplateContent

	// Try localized content first
	if locale != "" && locale != t.DefaultLocale {
		if loc := t.GetLocalization(locale); loc != nil && loc.EmailTemplate != nil {
			content = loc.EmailTemplate
		}
	}

	// Fall back to default
	if content == nil {
		content = t.EmailTemplate
	}

	if content == nil {
		return nil, NewTemplateError(t.Code, "email template not configured", "NO_EMAIL_TEMPLATE")
	}

	// Validate required variables
	if err := t.ValidateVariables(data); err != nil {
		return nil, err
	}

	// Render templates
	subject, err := renderText(content.Subject, data)
	if err != nil {
		return nil, NewTemplateError(t.Code, "failed to render subject", "RENDER_ERROR").WithInvalidVars([]string{"subject"})
	}

	body, err := renderText(content.Body, data)
	if err != nil {
		return nil, NewTemplateError(t.Code, "failed to render body", "RENDER_ERROR").WithInvalidVars([]string{"body"})
	}

	var htmlBody string
	if content.HTMLBody != "" {
		htmlBody, err = renderText(content.HTMLBody, data)
		if err != nil {
			return nil, NewTemplateError(t.Code, "failed to render HTML body", "RENDER_ERROR").WithInvalidVars([]string{"html_body"})
		}
	}

	return &RenderedEmail{
		Subject:     subject,
		Body:        body,
		HTMLBody:    htmlBody,
		PreviewText: content.PreviewText,
		FromName:    content.FromName,
		FromAddress: content.FromAddress,
		ReplyTo:     content.ReplyTo,
		Headers:     content.Headers,
		TrackOpens:  content.TrackOpens,
		TrackClicks: content.TrackClicks,
	}, nil
}

// RenderSMS renders the SMS template with the provided data.
func (t *NotificationTemplate) RenderSMS(data map[string]interface{}, locale string) (*RenderedSMS, error) {
	var content *SMSTemplateContent

	// Try localized content first
	if locale != "" && locale != t.DefaultLocale {
		if loc := t.GetLocalization(locale); loc != nil && loc.SMSTemplate != nil {
			content = loc.SMSTemplate
		}
	}

	if content == nil {
		content = t.SMSTemplate
	}

	if content == nil {
		return nil, NewTemplateError(t.Code, "SMS template not configured", "NO_SMS_TEMPLATE")
	}

	if err := t.ValidateVariables(data); err != nil {
		return nil, err
	}

	body, err := renderText(content.Body, data)
	if err != nil {
		return nil, NewTemplateError(t.Code, "failed to render SMS body", "RENDER_ERROR")
	}

	return &RenderedSMS{
		Body:     body,
		SenderID: content.SenderID,
	}, nil
}

// RenderPush renders the push notification template with the provided data.
func (t *NotificationTemplate) RenderPush(data map[string]interface{}, locale string) (*RenderedPush, error) {
	var content *PushTemplateContent

	if locale != "" && locale != t.DefaultLocale {
		if loc := t.GetLocalization(locale); loc != nil && loc.PushTemplate != nil {
			content = loc.PushTemplate
		}
	}

	if content == nil {
		content = t.PushTemplate
	}

	if content == nil {
		return nil, NewTemplateError(t.Code, "push template not configured", "NO_PUSH_TEMPLATE")
	}

	if err := t.ValidateVariables(data); err != nil {
		return nil, err
	}

	title, err := renderText(content.Title, data)
	if err != nil {
		return nil, NewTemplateError(t.Code, "failed to render push title", "RENDER_ERROR")
	}

	body, err := renderText(content.Body, data)
	if err != nil {
		return nil, NewTemplateError(t.Code, "failed to render push body", "RENDER_ERROR")
	}

	return &RenderedPush{
		Title:       title,
		Body:        body,
		ImageURL:    content.ImageURL,
		ActionURL:   content.ActionURL,
		Data:        content.Data,
		Sound:       content.Sound,
		Badge:       content.Badge,
		TTL:         content.TTL,
		CollapseKey: content.CollapseKey,
		Priority:    content.Priority,
	}, nil
}

// RenderInApp renders the in-app notification template with the provided data.
func (t *NotificationTemplate) RenderInApp(data map[string]interface{}, locale string) (*RenderedInApp, error) {
	var content *InAppTemplateContent

	if locale != "" && locale != t.DefaultLocale {
		if loc := t.GetLocalization(locale); loc != nil && loc.InAppTemplate != nil {
			content = loc.InAppTemplate
		}
	}

	if content == nil {
		content = t.InAppTemplate
	}

	if content == nil {
		return nil, NewTemplateError(t.Code, "in-app template not configured", "NO_INAPP_TEMPLATE")
	}

	if err := t.ValidateVariables(data); err != nil {
		return nil, err
	}

	title, err := renderText(content.Title, data)
	if err != nil {
		return nil, NewTemplateError(t.Code, "failed to render in-app title", "RENDER_ERROR")
	}

	body, err := renderText(content.Body, data)
	if err != nil {
		return nil, NewTemplateError(t.Code, "failed to render in-app body", "RENDER_ERROR")
	}

	return &RenderedInApp{
		Title:       title,
		Body:        body,
		IconURL:     content.IconURL,
		ImageURL:    content.ImageURL,
		ActionURL:   content.ActionURL,
		ActionLabel: content.ActionLabel,
		Style:       content.Style,
		Duration:    content.Duration,
		Data:        content.Data,
		Dismissable: content.Dismissable,
	}, nil
}

// renderText renders a Go template string with data.
func renderText(templateStr string, data map[string]interface{}) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	tmpl, err := template.New("notification").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ============================================================================
// Rendered Content Types
// ============================================================================

// RenderedEmail represents rendered email content.
type RenderedEmail struct {
	Subject     string
	Body        string
	HTMLBody    string
	PreviewText string
	FromName    string
	FromAddress string
	ReplyTo     string
	Headers     map[string]string
	TrackOpens  bool
	TrackClicks bool
}

// RenderedSMS represents rendered SMS content.
type RenderedSMS struct {
	Body     string
	SenderID string
}

// RenderedPush represents rendered push notification content.
type RenderedPush struct {
	Title       string
	Body        string
	ImageURL    string
	ActionURL   string
	Data        map[string]interface{}
	Sound       string
	Badge       *int
	TTL         int
	CollapseKey string
	Priority    string
}

// RenderedInApp represents rendered in-app notification content.
type RenderedInApp struct {
	Title       string
	Body        string
	IconURL     string
	ImageURL    string
	ActionURL   string
	ActionLabel string
	Style       string
	Duration    int
	Data        map[string]interface{}
	Dismissable bool
}

// ============================================================================
// Validation Methods
// ============================================================================

// Validate validates the template.
func (t *NotificationTemplate) Validate() error {
	var errs ValidationErrors

	if t.TenantID == uuid.Nil {
		errs.AddField("tenant_id", "tenant ID is required", "REQUIRED")
	}
	if t.Code == "" {
		errs.AddField("code", "template code is required", "REQUIRED")
	}
	if t.Name == "" {
		errs.AddField("name", "template name is required", "REQUIRED")
	}
	if len(t.Channels) == 0 {
		errs.AddField("channels", "at least one channel template is required", "REQUIRED")
	}

	// Validate channel-specific content
	if t.hasChannel(ChannelEmail) && t.EmailTemplate == nil {
		errs.AddField("email_template", "email template content is required", "REQUIRED")
	}
	if t.hasChannel(ChannelSMS) && t.SMSTemplate == nil {
		errs.AddField("sms_template", "SMS template content is required", "REQUIRED")
	}
	if t.hasChannel(ChannelPush) && t.PushTemplate == nil {
		errs.AddField("push_template", "push template content is required", "REQUIRED")
	}
	if t.hasChannel(ChannelInApp) && t.InAppTemplate == nil {
		errs.AddField("in_app_template", "in-app template content is required", "REQUIRED")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Clone creates a copy of the template with a new ID.
func (t *NotificationTemplate) Clone(newCode, newName string) (*NotificationTemplate, error) {
	clone, err := NewNotificationTemplate(t.TenantID, newCode, newName, t.Type)
	if err != nil {
		return nil, err
	}

	clone.Description = t.Description
	clone.Category = t.Category
	clone.Channels = append([]NotificationChannel{}, t.Channels...)
	clone.EmailTemplate = t.EmailTemplate
	clone.SMSTemplate = t.SMSTemplate
	clone.PushTemplate = t.PushTemplate
	clone.InAppTemplate = t.InAppTemplate
	clone.Variables = append([]TemplateVariable{}, t.Variables...)
	clone.DefaultLocale = t.DefaultLocale
	clone.RenderEngine = t.RenderEngine
	clone.Tags = append([]string{}, t.Tags...)

	// Copy localizations
	if t.Localizations != nil {
		clone.Localizations = make(map[string]*TemplateLocalization)
		for k, v := range t.Localizations {
			clone.Localizations[k] = v
		}
	}

	return clone, nil
}
