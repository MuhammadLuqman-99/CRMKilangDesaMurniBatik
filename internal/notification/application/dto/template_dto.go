// Package dto contains the Data Transfer Objects for the Notification service.
package dto

import (
	"time"
)

// TemplateDTO represents a notification template data transfer object.
type TemplateDTO struct {
	ID              string                 `json:"id"`
	TenantID        string                 `json:"tenant_id"`
	Code            string                 `json:"code"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Channel         string                 `json:"channel"`
	Type            string                 `json:"type"`
	Category        string                 `json:"category,omitempty"`
	Subject         string                 `json:"subject,omitempty"`
	Body            string                 `json:"body"`
	HTMLBody        string                 `json:"html_body,omitempty"`
	TextBody        string                 `json:"text_body,omitempty"`
	Variables       []TemplateVariableDTO  `json:"variables,omitempty"`
	Localizations   []LocalizationDTO      `json:"localizations,omitempty"`
	DefaultLocale   string                 `json:"default_locale"`
	Tags            []string               `json:"tags,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Version         int                    `json:"version"`
	Status          string                 `json:"status"` // draft, active, archived
	IsDraft         bool                   `json:"is_draft"`
	PublishedAt     *time.Time             `json:"published_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       string                 `json:"created_by,omitempty"`
	UpdatedBy       string                 `json:"updated_by,omitempty"`
}

// TemplateVariableDTO represents a template variable.
type TemplateVariableDTO struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // string, number, boolean, date, array, object
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Description  string      `json:"description,omitempty"`
	Format       string      `json:"format,omitempty"` // For date: RFC3339, custom format
	Validation   string      `json:"validation,omitempty"` // Regex or other validation rules
}

// LocalizationDTO represents a template localization.
type LocalizationDTO struct {
	Locale   string `json:"locale"`
	Subject  string `json:"subject,omitempty"`
	Body     string `json:"body"`
	HTMLBody string `json:"html_body,omitempty"`
	TextBody string `json:"text_body,omitempty"`
}

// TemplateListDTO represents a paginated list of templates.
type TemplateListDTO struct {
	Items      []TemplateSummaryDTO `json:"items"`
	TotalCount int64                `json:"total_count"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// TemplateSummaryDTO represents a summary of a template.
type TemplateSummaryDTO struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Channel     string    `json:"channel"`
	Type        string    `json:"type"`
	Category    string    `json:"category,omitempty"`
	Version     int       `json:"version"`
	Status      string    `json:"status"`
	IsDraft     bool      `json:"is_draft"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TemplateVersionDTO represents a template version.
type TemplateVersionDTO struct {
	Version      int                   `json:"version"`
	Subject      string                `json:"subject,omitempty"`
	Body         string                `json:"body"`
	HTMLBody     string                `json:"html_body,omitempty"`
	TextBody     string                `json:"text_body,omitempty"`
	Variables    []TemplateVariableDTO `json:"variables,omitempty"`
	ChangeSummary string               `json:"change_summary,omitempty"`
	PublishedAt  *time.Time            `json:"published_at,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	CreatedBy    string                `json:"created_by,omitempty"`
}

// TemplateVersionListDTO represents a list of template versions.
type TemplateVersionListDTO struct {
	TemplateID string               `json:"template_id"`
	Versions   []TemplateVersionDTO `json:"versions"`
	TotalCount int                  `json:"total_count"`
}

// === Request DTOs ===

// CreateTemplateRequest represents a request to create a notification template.
type CreateTemplateRequest struct {
	TenantID      string                 `json:"tenant_id" validate:"required,uuid"`
	Code          string                 `json:"code" validate:"required,min=1,max=100"` // Unique code like "welcome_email"
	Name          string                 `json:"name" validate:"required,min=1,max=255"`
	Description   string                 `json:"description,omitempty" validate:"omitempty,max=500"`
	Channel       string                 `json:"channel" validate:"required,oneof=email sms push in_app webhook slack whatsapp telegram"`
	Type          string                 `json:"type" validate:"required"`
	Category      string                 `json:"category,omitempty" validate:"omitempty,max=50"`
	Subject       string                 `json:"subject,omitempty" validate:"required_if=Channel email,max=500"`
	Body          string                 `json:"body" validate:"required,max=100000"`
	HTMLBody      string                 `json:"html_body,omitempty" validate:"omitempty,max=500000"`
	TextBody      string                 `json:"text_body,omitempty" validate:"omitempty,max=100000"`
	Variables     []TemplateVariableDTO  `json:"variables,omitempty" validate:"omitempty,dive"`
	Localizations []LocalizationDTO      `json:"localizations,omitempty" validate:"omitempty,dive"`
	DefaultLocale string                 `json:"default_locale" validate:"required,max=10"`
	Tags          []string               `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	IsDraft       bool                   `json:"is_draft"`
	CreatedBy     string                 `json:"created_by,omitempty" validate:"omitempty,uuid"`
}

// UpdateTemplateRequest represents a request to update a notification template.
type UpdateTemplateRequest struct {
	TenantID      string                 `json:"tenant_id" validate:"required,uuid"`
	TemplateID    string                 `json:"template_id" validate:"required,uuid"`
	Name          *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description   *string                `json:"description,omitempty" validate:"omitempty,max=500"`
	Category      *string                `json:"category,omitempty" validate:"omitempty,max=50"`
	Subject       *string                `json:"subject,omitempty" validate:"omitempty,max=500"`
	Body          *string                `json:"body,omitempty" validate:"omitempty,max=100000"`
	HTMLBody      *string                `json:"html_body,omitempty" validate:"omitempty,max=500000"`
	TextBody      *string                `json:"text_body,omitempty" validate:"omitempty,max=100000"`
	Variables     []TemplateVariableDTO  `json:"variables,omitempty" validate:"omitempty,dive"`
	Localizations []LocalizationDTO      `json:"localizations,omitempty" validate:"omitempty,dive"`
	DefaultLocale *string                `json:"default_locale,omitempty" validate:"omitempty,max=10"`
	Tags          []string               `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ChangeSummary string                 `json:"change_summary,omitempty" validate:"omitempty,max=500"`
	UpdatedBy     string                 `json:"updated_by,omitempty" validate:"omitempty,uuid"`
}

// PublishTemplateRequest represents a request to publish a template.
type PublishTemplateRequest struct {
	TenantID      string `json:"tenant_id" validate:"required,uuid"`
	TemplateID    string `json:"template_id" validate:"required,uuid"`
	ChangeSummary string `json:"change_summary,omitempty" validate:"omitempty,max=500"`
	PublishedBy   string `json:"published_by,omitempty" validate:"omitempty,uuid"`
}

// ArchiveTemplateRequest represents a request to archive a template.
type ArchiveTemplateRequest struct {
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	TemplateID string `json:"template_id" validate:"required,uuid"`
	ArchivedBy string `json:"archived_by,omitempty" validate:"omitempty,uuid"`
}

// RestoreTemplateRequest represents a request to restore an archived template.
type RestoreTemplateRequest struct {
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	TemplateID string `json:"template_id" validate:"required,uuid"`
	RestoredBy string `json:"restored_by,omitempty" validate:"omitempty,uuid"`
}

// DeleteTemplateRequest represents a request to delete a template.
type DeleteTemplateRequest struct {
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	TemplateID string `json:"template_id" validate:"required,uuid"`
	Force      bool   `json:"force,omitempty"` // Force delete even if in use
}

// GetTemplateRequest represents a request to get a template.
type GetTemplateRequest struct {
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	TemplateID string `json:"template_id" validate:"required,uuid"`
	Version    *int   `json:"version,omitempty"` // If not specified, returns the active version
}

// GetTemplateByNameRequest represents a request to get a template by name.
type GetTemplateByNameRequest struct {
	TenantID string `json:"tenant_id" validate:"required,uuid"`
	Name     string `json:"name" validate:"required,max=100"`
	Channel  string `json:"channel,omitempty"`
}

// GetTemplateByCodeRequest represents a request to get a template by code.
type GetTemplateByCodeRequest struct {
	TenantID string `json:"tenant_id" validate:"required,uuid"`
	Code     string `json:"code" validate:"required,max=100"`
}

// ListTemplatesRequest represents a request to list templates.
type ListTemplatesRequest struct {
	TenantID   string   `json:"tenant_id" validate:"required,uuid"`
	Channel    string   `json:"channel,omitempty" validate:"omitempty,oneof=email sms push in_app webhook slack whatsapp telegram"`
	Type       string   `json:"type,omitempty"`
	Category   string   `json:"category,omitempty"`
	Status     string   `json:"status,omitempty" validate:"omitempty,oneof=draft active archived"`
	Tags       []string `json:"tags,omitempty"`
	Search     string   `json:"search,omitempty" validate:"omitempty,max=100"`
	Page       int      `json:"page" validate:"min=1"`
	PageSize   int      `json:"page_size" validate:"min=1,max=100"`
	SortBy     string   `json:"sort_by,omitempty" validate:"omitempty,oneof=name created_at updated_at version"`
	SortOrder  string   `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// GetTemplateVersionsRequest represents a request to get template versions.
type GetTemplateVersionsRequest struct {
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	TemplateID string `json:"template_id" validate:"required,uuid"`
	Page       int    `json:"page" validate:"min=1"`
	PageSize   int    `json:"page_size" validate:"min=1,max=50"`
}

// RevertTemplateVersionRequest represents a request to revert to a previous version.
type RevertTemplateVersionRequest struct {
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	TemplateID string `json:"template_id" validate:"required,uuid"`
	Version    int    `json:"version" validate:"required,min=1"`
	RevertedBy string `json:"reverted_by,omitempty" validate:"omitempty,uuid"`
}

// CloneTemplateRequest represents a request to clone a template.
type CloneTemplateRequest struct {
	TenantID         string `json:"tenant_id" validate:"required,uuid"`
	SourceTemplateID string `json:"source_template_id" validate:"required,uuid"`
	NewCode          string `json:"new_code" validate:"required,min=1,max=100"` // Unique code for cloned template
	NewName          string `json:"new_name" validate:"required,min=1,max=255"`
	NewChannel       string `json:"new_channel,omitempty" validate:"omitempty,oneof=email sms push in_app webhook slack whatsapp telegram"`
	ClonedBy         string `json:"cloned_by,omitempty" validate:"omitempty,uuid"`
}

// RenderTemplateRequest represents a request to render a template.
type RenderTemplateRequest struct {
	TenantID   string                 `json:"tenant_id" validate:"required,uuid"`
	TemplateID string                 `json:"template_id" validate:"required,uuid"`
	Channel    string                 `json:"channel,omitempty" validate:"omitempty,oneof=email sms push in_app"`
	Version    *int                   `json:"version,omitempty"`
	Locale     string                 `json:"locale,omitempty"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
}

// ValidateTemplateRequest represents a request to validate a template.
type ValidateTemplateRequest struct {
	TenantID  string                 `json:"tenant_id" validate:"required,uuid"`
	Channel   string                 `json:"channel" validate:"required,oneof=email sms push in_app webhook slack whatsapp telegram"`
	Subject   string                 `json:"subject,omitempty"`
	Body      string                 `json:"body" validate:"required"`
	HTMLBody  string                 `json:"html_body,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// === Response DTOs ===

// CreateTemplateResponse represents a response after creating a template.
type CreateTemplateResponse struct {
	TemplateID string    `json:"template_id"`
	Name       string    `json:"name"`
	Version    int       `json:"version"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	Message    string    `json:"message,omitempty"`
}

// UpdateTemplateResponse represents a response after updating a template.
type UpdateTemplateResponse struct {
	TemplateID string    `json:"template_id"`
	Version    int       `json:"version"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
	Message    string    `json:"message,omitempty"`
}

// PublishTemplateResponse represents a response after publishing a template.
type PublishTemplateResponse struct {
	TemplateID  string    `json:"template_id"`
	Version     int       `json:"version"`
	PublishedAt time.Time `json:"published_at"`
	Message     string    `json:"message,omitempty"`
}

// ArchiveTemplateResponse represents a response after archiving a template.
type ArchiveTemplateResponse struct {
	TemplateID string    `json:"template_id"`
	ArchivedAt time.Time `json:"archived_at"`
	Message    string    `json:"message,omitempty"`
}

// RestoreTemplateResponse represents a response after restoring a template.
type RestoreTemplateResponse struct {
	TemplateID string    `json:"template_id"`
	Status     string    `json:"status"`
	RestoredAt time.Time `json:"restored_at"`
	Message    string    `json:"message,omitempty"`
}

// DeleteTemplateResponse represents a response after deleting a template.
type DeleteTemplateResponse struct {
	TemplateID string `json:"template_id"`
	Message    string `json:"message,omitempty"`
}

// RevertTemplateVersionResponse represents a response after reverting a template version.
type RevertTemplateVersionResponse struct {
	TemplateID string    `json:"template_id"`
	Version    int       `json:"version"`
	RevertedTo int       `json:"reverted_to"`
	RevertedAt time.Time `json:"reverted_at"`
	Message    string    `json:"message,omitempty"`
}

// CloneTemplateResponse represents a response after cloning a template.
type CloneTemplateResponse struct {
	TemplateID string    `json:"template_id"`
	Name       string    `json:"name"`
	Version    int       `json:"version"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	Message    string    `json:"message,omitempty"`
}

// RenderTemplateResponse represents a response after rendering a template.
type RenderTemplateResponse struct {
	Subject  string `json:"subject,omitempty"`
	Body     string `json:"body"`
	HTMLBody string `json:"html_body,omitempty"`
	TextBody string `json:"text_body,omitempty"`
}

// ValidateTemplateResponse represents a response after validating a template.
type ValidateTemplateResponse struct {
	Valid            bool                    `json:"valid"`
	Errors           []TemplateErrorDTO      `json:"errors,omitempty"`
	Warnings         []TemplateWarningDTO    `json:"warnings,omitempty"`
	DetectedVariables []string               `json:"detected_variables,omitempty"`
}

// TemplateErrorDTO represents a template validation error.
type TemplateErrorDTO struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// TemplateWarningDTO represents a template validation warning.
type TemplateWarningDTO struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// TemplateUsageDTO represents template usage statistics.
type TemplateUsageDTO struct {
	TemplateID    string    `json:"template_id"`
	Name          string    `json:"name"`
	TotalUsed     int64     `json:"total_used"`
	TotalSent     int64     `json:"total_sent"`
	TotalDelivered int64    `json:"total_delivered"`
	TotalFailed   int64     `json:"total_failed"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	Period        string    `json:"period"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
}

// GetTemplateUsageRequest represents a request to get template usage.
type GetTemplateUsageRequest struct {
	TenantID   string    `json:"tenant_id" validate:"required,uuid"`
	TemplateID string    `json:"template_id" validate:"required,uuid"`
	StartDate  time.Time `json:"start_date" validate:"required"`
	EndDate    time.Time `json:"end_date" validate:"required"`
}
