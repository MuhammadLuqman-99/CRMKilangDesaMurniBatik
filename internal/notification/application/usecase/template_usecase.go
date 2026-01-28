// Package usecase contains the application use cases for the Notification service.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kilang-desa-murni/crm/internal/notification/application"
	"github.com/kilang-desa-murni/crm/internal/notification/application/dto"
	"github.com/kilang-desa-murni/crm/internal/notification/application/mapper"
	"github.com/kilang-desa-murni/crm/internal/notification/application/ports"
	"github.com/kilang-desa-murni/crm/internal/notification/domain"
)

// TemplateUseCase defines the interface for template use cases.
type TemplateUseCase interface {
	// CreateTemplate creates a new notification template.
	CreateTemplate(ctx context.Context, req *dto.CreateTemplateRequest) (*dto.CreateTemplateResponse, error)
	// UpdateTemplate updates an existing notification template.
	UpdateTemplate(ctx context.Context, req *dto.UpdateTemplateRequest) (*dto.UpdateTemplateResponse, error)
	// PublishTemplate publishes a draft template.
	PublishTemplate(ctx context.Context, req *dto.PublishTemplateRequest) (*dto.PublishTemplateResponse, error)
	// ArchiveTemplate archives a template.
	ArchiveTemplate(ctx context.Context, req *dto.ArchiveTemplateRequest) (*dto.ArchiveTemplateResponse, error)
	// RestoreTemplate restores an archived template.
	RestoreTemplate(ctx context.Context, req *dto.RestoreTemplateRequest) (*dto.RestoreTemplateResponse, error)
	// DeleteTemplate deletes a template.
	DeleteTemplate(ctx context.Context, req *dto.DeleteTemplateRequest) (*dto.DeleteTemplateResponse, error)
	// GetTemplate retrieves a template by ID.
	GetTemplate(ctx context.Context, req *dto.GetTemplateRequest) (*dto.TemplateDTO, error)
	// GetTemplateByCode retrieves a template by code.
	GetTemplateByCode(ctx context.Context, req *dto.GetTemplateByCodeRequest) (*dto.TemplateDTO, error)
	// ListTemplates lists templates with filtering and pagination.
	ListTemplates(ctx context.Context, req *dto.ListTemplatesRequest) (*dto.TemplateListDTO, error)
	// CloneTemplate clones an existing template.
	CloneTemplate(ctx context.Context, req *dto.CloneTemplateRequest) (*dto.CloneTemplateResponse, error)
	// RenderTemplate renders a template with variables.
	RenderTemplate(ctx context.Context, req *dto.RenderTemplateRequest) (*dto.RenderTemplateResponse, error)
	// ValidateTemplate validates template syntax.
	ValidateTemplate(ctx context.Context, req *dto.ValidateTemplateRequest) (*dto.ValidateTemplateResponse, error)
}

// templateUseCase implements the TemplateUseCase interface.
type templateUseCase struct {
	templateRepo     domain.TemplateRepository
	notificationRepo domain.NotificationRepository
	eventPublisher   ports.EventPublisher
	cache            ports.CacheService
	idGenerator      ports.IdGenerator
	timeProvider     ports.TimeProvider
	metrics          ports.MetricsCollector
	logger           ports.Logger

	mapper *mapper.TemplateMapper
}

// TemplateUseCaseConfig holds configuration for the template use case.
type TemplateUseCaseConfig struct {
	TemplateRepo     domain.TemplateRepository
	NotificationRepo domain.NotificationRepository
	EventPublisher   ports.EventPublisher
	Cache            ports.CacheService
	IdGenerator      ports.IdGenerator
	TimeProvider     ports.TimeProvider
	Metrics          ports.MetricsCollector
	Logger           ports.Logger
}

// NewTemplateUseCase creates a new TemplateUseCase.
func NewTemplateUseCase(cfg TemplateUseCaseConfig) TemplateUseCase {
	return &templateUseCase{
		templateRepo:     cfg.TemplateRepo,
		notificationRepo: cfg.NotificationRepo,
		eventPublisher:   cfg.EventPublisher,
		cache:            cfg.Cache,
		idGenerator:      cfg.IdGenerator,
		timeProvider:     cfg.TimeProvider,
		metrics:          cfg.Metrics,
		logger:           cfg.Logger,

		mapper: mapper.NewTemplateMapper(),
	}
}

// CreateTemplate creates a new notification template.
func (uc *templateUseCase) CreateTemplate(ctx context.Context, req *dto.CreateTemplateRequest) (*dto.CreateTemplateResponse, error) {
	uc.logger.WithContext(ctx).Info("CreateTemplate started", map[string]interface{}{
		"tenant_id": req.TenantID,
		"name":      req.Name,
		"code":      req.Code,
		"channel":   req.Channel,
	})

	// Validate request
	if err := uc.validateCreateTemplateRequest(req); err != nil {
		return nil, err
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Check if template with same code already exists
	exists, err := uc.templateRepo.ExistsByCode(ctx, tenantID, req.Code)
	if err != nil {
		return nil, application.NewInternalError("failed to check template existence", err)
	}
	if exists {
		return nil, application.NewTemplateAlreadyExistsError(req.Code)
	}

	// Parse notification type
	notificationType, err := domain.ParseType(req.Type)
	if err != nil {
		return nil, application.NewInvalidInputError(fmt.Sprintf("invalid notification type: %s", req.Type))
	}

	// Create template entity
	template, err := domain.NewNotificationTemplate(
		tenantID,
		req.Code,
		req.Name,
		notificationType,
	)
	if err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set description if provided
	if req.Description != "" {
		template.Description = req.Description
	}

	// Set category if provided
	if req.Category != "" {
		template.Category = req.Category
	}

	// Set default locale
	if req.DefaultLocale != "" {
		template.DefaultLocale = req.DefaultLocale
	}

	// Set tags
	for _, tag := range req.Tags {
		template.AddTag(tag)
	}

	// Set channel-specific template content based on channel
	channel, err := domain.ParseChannel(req.Channel)
	if err != nil {
		return nil, application.NewInvalidInputError(fmt.Sprintf("invalid channel: %s", req.Channel))
	}

	switch channel {
	case domain.ChannelEmail:
		emailContent := &domain.EmailTemplateContent{
			Subject:  req.Subject,
			Body:     req.Body,
			HTMLBody: req.HTMLBody,
		}
		if err := template.SetEmailTemplate(emailContent); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	case domain.ChannelSMS:
		smsContent := &domain.SMSTemplateContent{
			Body: req.Body,
		}
		if err := template.SetSMSTemplate(smsContent); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	case domain.ChannelPush:
		pushContent := &domain.PushTemplateContent{
			Title: req.Subject,
			Body:  req.Body,
		}
		if err := template.SetPushTemplate(pushContent); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	case domain.ChannelInApp:
		inAppContent := &domain.InAppTemplateContent{
			Title: req.Subject,
			Body:  req.Body,
		}
		if err := template.SetInAppTemplate(inAppContent); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Add variables
	for _, v := range req.Variables {
		variable := domain.TemplateVariable{
			Name:         v.Name,
			Type:         v.Type,
			Required:     v.Required,
			DefaultValue: v.DefaultValue,
			Description:  v.Description,
		}
		if err := template.AddVariable(variable); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Set created by
	if req.CreatedBy != "" {
		createdByID, err := uuid.Parse(req.CreatedBy)
		if err == nil {
			template.CreatedBy = &createdByID
		}
	}

	// Publish if not a draft
	if !req.IsDraft {
		template.Publish()
	}

	// Save template
	if err := uc.templateRepo.Create(ctx, template); err != nil {
		return nil, application.NewInternalError("failed to create template", err)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, template)

	// Invalidate cache
	uc.invalidateTemplateCache(ctx, req.TenantID, template.ID.String(), req.Code)

	uc.metrics.IncrementCounter(ctx, "template.created", map[string]string{
		"tenant_id": req.TenantID,
		"channel":   req.Channel,
	})

	status := "active"
	if template.HasDraft() {
		status = "draft"
	}

	return &dto.CreateTemplateResponse{
		TemplateID: template.ID.String(),
		Name:       req.Name,
		Version:    template.TemplateVersion,
		Status:     status,
		CreatedAt:  template.CreatedAt,
		Message:    "Template created successfully",
	}, nil
}

// UpdateTemplate updates an existing notification template.
func (uc *templateUseCase) UpdateTemplate(ctx context.Context, req *dto.UpdateTemplateRequest) (*dto.UpdateTemplateResponse, error) {
	uc.logger.WithContext(ctx).Info("UpdateTemplate started", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"template_id": req.TemplateID,
	})

	// Parse template ID
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Get existing template
	template, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(req.TemplateID)
	}

	// Verify tenant access
	if template.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	// Check if template is locked
	if template.IsLocked {
		return nil, application.NewInvalidStateError("cannot update a locked template")
	}

	// Check if template is archived
	if template.DeletedAt != nil {
		return nil, application.NewInvalidStateError("cannot update an archived template")
	}

	// Store old code for cache invalidation
	oldCode := template.Code

	// Apply updates
	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Description != nil {
		template.Description = *req.Description
	}
	if req.Category != nil {
		template.Category = *req.Category
	}

	// Update channel-specific content based on the template's channels
	if req.Subject != nil || req.Body != nil || req.HTMLBody != nil {
		if template.EmailTemplate != nil {
			if req.Subject != nil {
				template.EmailTemplate.Subject = *req.Subject
			}
			if req.Body != nil {
				template.EmailTemplate.Body = *req.Body
			}
			if req.HTMLBody != nil {
				template.EmailTemplate.HTMLBody = *req.HTMLBody
			}
		}
		if template.SMSTemplate != nil && req.Body != nil {
			template.SMSTemplate.Body = *req.Body
		}
		if template.PushTemplate != nil {
			if req.Subject != nil {
				template.PushTemplate.Title = *req.Subject
			}
			if req.Body != nil {
				template.PushTemplate.Body = *req.Body
			}
		}
		if template.InAppTemplate != nil {
			if req.Subject != nil {
				template.InAppTemplate.Title = *req.Subject
			}
			if req.Body != nil {
				template.InAppTemplate.Body = *req.Body
			}
		}
	}

	// Set updated by
	if req.UpdatedBy != "" {
		updatedByID, err := uuid.Parse(req.UpdatedBy)
		if err == nil {
			template.UpdatedBy = &updatedByID
		}
	}

	// Mark as updated
	template.MarkUpdated()

	// Save template
	if err := uc.templateRepo.Update(ctx, template); err != nil {
		return nil, application.NewInternalError("failed to update template", err)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, template)

	// Invalidate cache
	uc.invalidateTemplateCache(ctx, req.TenantID, req.TemplateID, oldCode)
	uc.invalidateTemplateCache(ctx, req.TenantID, req.TemplateID, template.Code)

	uc.metrics.IncrementCounter(ctx, "template.updated", map[string]string{
		"tenant_id": req.TenantID,
	})

	status := "active"
	if template.HasDraft() {
		status = "draft"
	}
	if !template.IsActive {
		status = "inactive"
	}

	return &dto.UpdateTemplateResponse{
		TemplateID: req.TemplateID,
		Version:    template.TemplateVersion,
		Status:     status,
		UpdatedAt:  template.UpdatedAt,
		Message:    "Template updated successfully",
	}, nil
}

// PublishTemplate publishes a draft template.
func (uc *templateUseCase) PublishTemplate(ctx context.Context, req *dto.PublishTemplateRequest) (*dto.PublishTemplateResponse, error) {
	uc.logger.WithContext(ctx).Info("PublishTemplate started", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"template_id": req.TemplateID,
	})

	// Parse template ID
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Get template
	template, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(req.TemplateID)
	}

	// Verify tenant access
	if template.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	// Check if template is archived
	if template.DeletedAt != nil {
		return nil, application.NewInvalidStateError("cannot publish an archived template")
	}

	// Validate template before publishing
	if err := template.Validate(); err != nil {
		return nil, application.NewAppError(application.ErrCodeTemplateInvalid, fmt.Sprintf("cannot publish: %s", err.Error()))
	}

	// Publish template
	template.Publish()

	// Set published by
	if req.PublishedBy != "" {
		publishedByID, err := uuid.Parse(req.PublishedBy)
		if err == nil {
			template.UpdatedBy = &publishedByID
		}
	}

	// Save template
	if err := uc.templateRepo.Update(ctx, template); err != nil {
		return nil, application.NewInternalError("failed to publish template", err)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, template)

	// Invalidate cache
	uc.invalidateTemplateCache(ctx, req.TenantID, req.TemplateID, template.Code)

	uc.metrics.IncrementCounter(ctx, "template.published", map[string]string{
		"tenant_id": req.TenantID,
	})

	return &dto.PublishTemplateResponse{
		TemplateID:  req.TemplateID,
		Version:     template.TemplateVersion,
		PublishedAt: *template.PublishedAt,
		Message:     "Template published successfully",
	}, nil
}

// ArchiveTemplate archives a template.
func (uc *templateUseCase) ArchiveTemplate(ctx context.Context, req *dto.ArchiveTemplateRequest) (*dto.ArchiveTemplateResponse, error) {
	uc.logger.WithContext(ctx).Info("ArchiveTemplate started", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"template_id": req.TemplateID,
	})

	// Parse template ID
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Get template
	template, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(req.TemplateID)
	}

	// Verify tenant access
	if template.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	// Check if already archived
	if template.DeletedAt != nil {
		return nil, application.NewInvalidStateError("template is already archived")
	}

	// Archive template (soft delete)
	if err := uc.templateRepo.Delete(ctx, templateID); err != nil {
		return nil, application.NewInternalError("failed to archive template", err)
	}

	// Invalidate cache
	uc.invalidateTemplateCache(ctx, req.TenantID, req.TemplateID, template.Code)

	uc.metrics.IncrementCounter(ctx, "template.archived", map[string]string{
		"tenant_id": req.TenantID,
	})

	return &dto.ArchiveTemplateResponse{
		TemplateID: req.TemplateID,
		ArchivedAt: time.Now().UTC(),
		Message:    "Template archived successfully",
	}, nil
}

// RestoreTemplate restores an archived template.
func (uc *templateUseCase) RestoreTemplate(ctx context.Context, req *dto.RestoreTemplateRequest) (*dto.RestoreTemplateResponse, error) {
	uc.logger.WithContext(ctx).Info("RestoreTemplate started", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"template_id": req.TemplateID,
	})

	// Parse template ID
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Get template (including archived)
	template, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(req.TemplateID)
	}

	// Verify tenant access
	if template.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	// Check if archived
	if template.DeletedAt == nil {
		return nil, application.NewInvalidStateError("template is not archived")
	}

	// Restore template
	now := uc.timeProvider.NowUTC()
	template.DeletedAt = nil
	template.UpdatedAt = now

	if req.RestoredBy != "" {
		restoredByID, err := uuid.Parse(req.RestoredBy)
		if err == nil {
			template.UpdatedBy = &restoredByID
		}
	}

	// Save template
	if err := uc.templateRepo.Update(ctx, template); err != nil {
		return nil, application.NewInternalError("failed to restore template", err)
	}

	// Invalidate cache
	uc.invalidateTemplateCache(ctx, req.TenantID, req.TemplateID, template.Code)

	uc.metrics.IncrementCounter(ctx, "template.restored", map[string]string{
		"tenant_id": req.TenantID,
	})

	status := "active"
	if template.HasDraft() {
		status = "draft"
	}
	if !template.IsActive {
		status = "inactive"
	}

	return &dto.RestoreTemplateResponse{
		TemplateID: req.TemplateID,
		Status:     status,
		RestoredAt: now,
		Message:    "Template restored successfully",
	}, nil
}

// DeleteTemplate permanently deletes a template.
func (uc *templateUseCase) DeleteTemplate(ctx context.Context, req *dto.DeleteTemplateRequest) (*dto.DeleteTemplateResponse, error) {
	uc.logger.WithContext(ctx).Info("DeleteTemplate started", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"template_id": req.TemplateID,
		"force":       req.Force,
	})

	// Parse template ID
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Get template
	template, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(req.TemplateID)
	}

	// Verify tenant access
	if template.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	// Check if template is locked
	if template.IsLocked && !req.Force {
		return nil, application.NewInvalidStateError("cannot delete a locked template without force flag")
	}

	// Hard delete template
	if err := uc.templateRepo.HardDelete(ctx, templateID); err != nil {
		return nil, application.NewInternalError("failed to delete template", err)
	}

	// Invalidate cache
	uc.invalidateTemplateCache(ctx, req.TenantID, req.TemplateID, template.Code)

	uc.metrics.IncrementCounter(ctx, "template.deleted", map[string]string{
		"tenant_id": req.TenantID,
	})

	return &dto.DeleteTemplateResponse{
		TemplateID: req.TemplateID,
		Message:    "Template deleted successfully",
	}, nil
}

// GetTemplate retrieves a template by ID.
func (uc *templateUseCase) GetTemplate(ctx context.Context, req *dto.GetTemplateRequest) (*dto.TemplateDTO, error) {
	// Parse template ID
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	template, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(req.TemplateID)
	}

	// Verify tenant access
	if template.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	return uc.mapper.ToDTO(template), nil
}

// GetTemplateByCode retrieves a template by code.
func (uc *templateUseCase) GetTemplateByCode(ctx context.Context, req *dto.GetTemplateByCodeRequest) (*dto.TemplateDTO, error) {
	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	template, err := uc.templateRepo.FindByCode(ctx, tenantID, req.Code)
	if err != nil {
		return nil, application.NewNotFoundError("template", req.Code)
	}

	return uc.mapper.ToDTO(template), nil
}

// ListTemplates lists templates with filtering and pagination.
func (uc *templateUseCase) ListTemplates(ctx context.Context, req *dto.ListTemplatesRequest) (*dto.TemplateListDTO, error) {
	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Build filter
	filter := domain.TemplateFilter{
		TenantID:  &tenantID,
		Query:     req.Search,
		Offset:    (req.Page - 1) * req.PageSize,
		Limit:     req.PageSize,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}

	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	if req.Channel != "" {
		channel, err := domain.ParseChannel(req.Channel)
		if err == nil {
			filter.Channels = []domain.NotificationChannel{channel}
		}
	}

	if req.Type != "" {
		notificationType, err := domain.ParseType(req.Type)
		if err == nil {
			filter.Types = []domain.NotificationType{notificationType}
		}
	}

	if req.Category != "" {
		filter.Categories = []string{req.Category}
	}

	if len(req.Tags) > 0 {
		filter.Tags = req.Tags
	}

	// Handle status filter
	if req.Status != "" {
		switch req.Status {
		case "active":
			isActive := true
			filter.IsActive = &isActive
		case "inactive":
			isActive := false
			filter.IsActive = &isActive
		case "archived":
			filter.IncludeDeleted = true
		}
	}

	// Execute query
	templateList, err := uc.templateRepo.List(ctx, filter)
	if err != nil {
		return nil, application.NewInternalError("failed to list templates", err)
	}

	return uc.mapper.TemplateListToDTO(templateList.Templates, templateList.Total, req.Page, req.PageSize), nil
}

// CloneTemplate clones an existing template.
func (uc *templateUseCase) CloneTemplate(ctx context.Context, req *dto.CloneTemplateRequest) (*dto.CloneTemplateResponse, error) {
	uc.logger.WithContext(ctx).Info("CloneTemplate started", map[string]interface{}{
		"tenant_id":          req.TenantID,
		"source_template_id": req.SourceTemplateID,
		"new_code":           req.NewCode,
		"new_name":           req.NewName,
	})

	// Parse source template ID
	sourceTemplateID, err := uuid.Parse(req.SourceTemplateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid source template ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Get source template
	sourceTemplate, err := uc.templateRepo.FindByID(ctx, sourceTemplateID)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(req.SourceTemplateID)
	}

	// Verify tenant access
	if sourceTemplate.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to source template")
	}

	// Check if template with new code already exists
	exists, err := uc.templateRepo.ExistsByCode(ctx, tenantID, req.NewCode)
	if err != nil {
		return nil, application.NewInternalError("failed to check template existence", err)
	}
	if exists {
		return nil, application.NewTemplateAlreadyExistsError(req.NewCode)
	}

	// Clone the template
	clonedTemplate, err := sourceTemplate.Clone(req.NewCode, req.NewName)
	if err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set cloned by
	if req.ClonedBy != "" {
		clonedByID, err := uuid.Parse(req.ClonedBy)
		if err == nil {
			clonedTemplate.CreatedBy = &clonedByID
		}
	}

	// Save new template
	if err := uc.templateRepo.Create(ctx, clonedTemplate); err != nil {
		return nil, application.NewInternalError("failed to create cloned template", err)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, clonedTemplate)

	uc.metrics.IncrementCounter(ctx, "template.cloned", map[string]string{
		"tenant_id": req.TenantID,
	})

	status := "active"
	if clonedTemplate.HasDraft() {
		status = "draft"
	}

	return &dto.CloneTemplateResponse{
		TemplateID: clonedTemplate.ID.String(),
		Name:       req.NewName,
		Version:    clonedTemplate.TemplateVersion,
		Status:     status,
		CreatedAt:  clonedTemplate.CreatedAt,
		Message:    "Template cloned successfully",
	}, nil
}

// RenderTemplate renders a template with variables.
func (uc *templateUseCase) RenderTemplate(ctx context.Context, req *dto.RenderTemplateRequest) (*dto.RenderTemplateResponse, error) {
	// Parse template ID
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Get template
	template, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(req.TemplateID)
	}

	// Verify tenant access
	if template.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	// Determine locale
	locale := req.Locale
	if locale == "" {
		locale = template.DefaultLocale
	}

	// Render based on channel
	channel, _ := domain.ParseChannel(req.Channel)
	switch channel {
	case domain.ChannelEmail:
		rendered, err := template.RenderEmail(req.Variables, locale)
		if err != nil {
			return nil, application.NewTemplateRenderFailedError(req.TemplateID, err.Error())
		}
		return uc.mapper.RenderResponseFromResult(rendered.Subject, rendered.Body, rendered.HTMLBody, rendered.Body), nil

	case domain.ChannelSMS:
		rendered, err := template.RenderSMS(req.Variables, locale)
		if err != nil {
			return nil, application.NewTemplateRenderFailedError(req.TemplateID, err.Error())
		}
		return uc.mapper.RenderResponseFromResult("", rendered.Body, "", rendered.Body), nil

	case domain.ChannelPush:
		rendered, err := template.RenderPush(req.Variables, locale)
		if err != nil {
			return nil, application.NewTemplateRenderFailedError(req.TemplateID, err.Error())
		}
		return uc.mapper.RenderResponseFromResult(rendered.Title, rendered.Body, "", rendered.Body), nil

	case domain.ChannelInApp:
		rendered, err := template.RenderInApp(req.Variables, locale)
		if err != nil {
			return nil, application.NewTemplateRenderFailedError(req.TemplateID, err.Error())
		}
		return uc.mapper.RenderResponseFromResult(rendered.Title, rendered.Body, "", rendered.Body), nil

	default:
		// Try email as default
		if template.EmailTemplate != nil {
			rendered, err := template.RenderEmail(req.Variables, locale)
			if err != nil {
				return nil, application.NewTemplateRenderFailedError(req.TemplateID, err.Error())
			}
			return uc.mapper.RenderResponseFromResult(rendered.Subject, rendered.Body, rendered.HTMLBody, rendered.Body), nil
		}
		return nil, application.NewInvalidInputError("no renderable template content found")
	}
}

// ValidateTemplate validates template syntax.
func (uc *templateUseCase) ValidateTemplate(ctx context.Context, req *dto.ValidateTemplateRequest) (*dto.ValidateTemplateResponse, error) {
	errors := make([]dto.TemplateErrorDTO, 0)
	warnings := make([]dto.TemplateWarningDTO, 0)
	detectedVariables := make([]string, 0)

	// Try to parse the body as a Go template to validate syntax
	if req.Body != "" {
		if err := validateTemplateSyntax(req.Body); err != nil {
			errors = append(errors, dto.TemplateErrorDTO{
				Field:   "body",
				Message: err.Error(),
			})
		}
	}

	// Validate HTML body if provided
	if req.HTMLBody != "" {
		if err := validateTemplateSyntax(req.HTMLBody); err != nil {
			errors = append(errors, dto.TemplateErrorDTO{
				Field:   "html_body",
				Message: err.Error(),
			})
		}
	}

	// Validate subject for email channel
	if req.Channel == "email" && req.Subject != "" {
		if err := validateTemplateSyntax(req.Subject); err != nil {
			errors = append(errors, dto.TemplateErrorDTO{
				Field:   "subject",
				Message: err.Error(),
			})
		}
	}

	// Extract variables from template
	detectedVariables = extractTemplateVariables(req.Body)

	// Add warnings for unused variables
	if req.Variables != nil {
		for varName := range req.Variables {
			found := false
			for _, detected := range detectedVariables {
				if detected == varName {
					found = true
					break
				}
			}
			if !found {
				warnings = append(warnings, dto.TemplateWarningDTO{
					Field:   "variables",
					Message: fmt.Sprintf("variable '%s' is provided but not used in template", varName),
				})
			}
		}
	}

	return uc.mapper.ToValidationResponse(len(errors) == 0, errors, warnings, detectedVariables), nil
}

// === Private helper methods ===

func (uc *templateUseCase) validateCreateTemplateRequest(req *dto.CreateTemplateRequest) error {
	if req.TenantID == "" {
		return application.NewValidationError("tenant_id is required")
	}
	if req.Code == "" {
		return application.NewValidationError("code is required")
	}
	if len(req.Code) > 100 {
		return application.NewValidationError("code must be 100 characters or less")
	}
	if req.Name == "" {
		return application.NewValidationError("name is required")
	}
	if len(req.Name) > 255 {
		return application.NewValidationError("name must be 255 characters or less")
	}
	if req.Channel == "" {
		return application.NewValidationError("channel is required")
	}
	if req.Type == "" {
		return application.NewValidationError("type is required")
	}
	if req.Body == "" {
		return application.NewValidationError("body is required")
	}
	if req.Channel == "email" && req.Subject == "" {
		return application.NewValidationError("subject is required for email templates")
	}
	return nil
}

func (uc *templateUseCase) publishDomainEvents(ctx context.Context, template *domain.NotificationTemplate) {
	events := template.GetDomainEvents()
	for _, event := range events {
		if err := uc.eventPublisher.Publish(ctx, event); err != nil {
			uc.logger.WithContext(ctx).Error("failed to publish domain event", err, map[string]interface{}{
				"event_type":  event.EventType(),
				"template_id": template.ID.String(),
			})
		}
	}
	template.ClearDomainEvents()
}

func (uc *templateUseCase) invalidateTemplateCache(ctx context.Context, tenantID, templateID, templateCode string) {
	// Invalidate by ID
	cacheKey := fmt.Sprintf("template:%s:%s", tenantID, templateID)
	if err := uc.cache.Delete(ctx, cacheKey); err != nil {
		uc.logger.WithContext(ctx).Warn("failed to invalidate template cache", map[string]interface{}{
			"cache_key": cacheKey,
			"error":     err.Error(),
		})
	}

	// Invalidate by code
	cacheKeyByCode := fmt.Sprintf("template:%s:code:%s", tenantID, templateCode)
	if err := uc.cache.Delete(ctx, cacheKeyByCode); err != nil {
		uc.logger.WithContext(ctx).Warn("failed to invalidate template cache by code", map[string]interface{}{
			"cache_key": cacheKeyByCode,
			"error":     err.Error(),
		})
	}
}

// validateTemplateSyntax validates Go template syntax.
func validateTemplateSyntax(templateStr string) error {
	if templateStr == "" {
		return nil
	}
	_, err := parseGoTemplate(templateStr)
	return err
}

// parseGoTemplate parses a Go template string.
func parseGoTemplate(templateStr string) (interface{}, error) {
	// Using text/template from standard library
	// Import would be: "text/template"
	// For now, simplified validation
	if templateStr == "" {
		return nil, nil
	}
	// Check for balanced braces
	openCount := 0
	for i := 0; i < len(templateStr)-1; i++ {
		if templateStr[i] == '{' && templateStr[i+1] == '{' {
			openCount++
			i++
		} else if templateStr[i] == '}' && templateStr[i+1] == '}' {
			openCount--
			i++
			if openCount < 0 {
				return nil, fmt.Errorf("unbalanced template braces at position %d", i)
			}
		}
	}
	if openCount != 0 {
		return nil, fmt.Errorf("unbalanced template braces: %d unclosed", openCount)
	}
	return templateStr, nil
}

// extractTemplateVariables extracts variable names from a Go template string.
func extractTemplateVariables(templateStr string) []string {
	variables := make([]string, 0)
	// Simple extraction of .VariableName patterns
	// In a real implementation, we would parse the template properly
	i := 0
	for i < len(templateStr)-2 {
		if templateStr[i] == '{' && templateStr[i+1] == '{' {
			// Find the closing }}
			j := i + 2
			for j < len(templateStr)-1 && !(templateStr[j] == '}' && templateStr[j+1] == '}') {
				j++
			}
			if j < len(templateStr)-1 {
				content := templateStr[i+2 : j]
				// Extract variable name after .
				for k := 0; k < len(content); k++ {
					if content[k] == '.' {
						// Find end of variable name
						start := k + 1
						end := start
						for end < len(content) && (content[end] >= 'a' && content[end] <= 'z' ||
							content[end] >= 'A' && content[end] <= 'Z' ||
							content[end] >= '0' && content[end] <= '9' ||
							content[end] == '_') {
							end++
						}
						if end > start {
							varName := content[start:end]
							// Check if already in list
							found := false
							for _, v := range variables {
								if v == varName {
									found = true
									break
								}
							}
							if !found {
								variables = append(variables, varName)
							}
						}
					}
				}
			}
			i = j + 2
		} else {
			i++
		}
	}
	return variables
}
