// Package mapper contains the mappers for converting between domain entities and DTOs.
package mapper

import (
	"github.com/kilang-desa-murni/crm/internal/notification/application/dto"
	"github.com/kilang-desa-murni/crm/internal/notification/domain"
)

// TemplateMapper provides mapping between NotificationTemplate entities and DTOs.
type TemplateMapper struct{}

// NewTemplateMapper creates a new TemplateMapper.
func NewTemplateMapper() *TemplateMapper {
	return &TemplateMapper{}
}

// ToDTO converts a NotificationTemplate entity to a TemplateDTO.
func (m *TemplateMapper) ToDTO(entity *domain.NotificationTemplate) *dto.TemplateDTO {
	if entity == nil {
		return nil
	}

	// Determine channel - pick first from Channels array
	channel := ""
	if len(entity.Channels) > 0 {
		channel = entity.Channels[0].String()
	}

	// Extract body based on available templates
	subject := ""
	body := ""
	htmlBody := ""

	if entity.EmailTemplate != nil {
		subject = entity.EmailTemplate.Subject
		body = entity.EmailTemplate.Body
		htmlBody = entity.EmailTemplate.HTMLBody
	} else if entity.SMSTemplate != nil {
		body = entity.SMSTemplate.Body
	} else if entity.PushTemplate != nil {
		subject = entity.PushTemplate.Title
		body = entity.PushTemplate.Body
	} else if entity.InAppTemplate != nil {
		subject = entity.InAppTemplate.Title
		body = entity.InAppTemplate.Body
	}

	result := &dto.TemplateDTO{
		ID:            entity.ID.String(),
		TenantID:      entity.TenantID.String(),
		Code:          entity.Code,
		Name:          entity.Name,
		Description:   entity.Description,
		Channel:       channel,
		Type:          entity.Type.String(),
		Category:      entity.Category,
		Subject:       subject,
		Body:          body,
		HTMLBody:      htmlBody,
		Variables:     m.variablesToDTO(entity.Variables),
		Localizations: m.localizationsToDTO(entity.Localizations),
		DefaultLocale: entity.DefaultLocale,
		Tags:          entity.Tags,
		Version:       entity.TemplateVersion,
		Status:        m.getTemplateStatus(entity),
		IsDraft:       entity.HasDraft(),
		PublishedAt:   entity.PublishedAt,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}

	if entity.CreatedBy != nil {
		result.CreatedBy = entity.CreatedBy.String()
	}
	if entity.UpdatedBy != nil {
		result.UpdatedBy = entity.UpdatedBy.String()
	}

	return result
}

// ToDTOList converts a list of NotificationTemplate entities to a list of TemplateDTOs.
func (m *TemplateMapper) ToDTOList(entities []*domain.NotificationTemplate) []dto.TemplateDTO {
	result := make([]dto.TemplateDTO, 0, len(entities))
	for _, entity := range entities {
		if entity != nil {
			result = append(result, *m.ToDTO(entity))
		}
	}
	return result
}

// ToSummaryDTO converts a NotificationTemplate entity to a TemplateSummaryDTO.
func (m *TemplateMapper) ToSummaryDTO(entity *domain.NotificationTemplate) *dto.TemplateSummaryDTO {
	if entity == nil {
		return nil
	}

	// Determine channel - pick first from Channels array
	channel := ""
	if len(entity.Channels) > 0 {
		channel = entity.Channels[0].String()
	}

	return &dto.TemplateSummaryDTO{
		ID:          entity.ID.String(),
		TenantID:    entity.TenantID.String(),
		Code:        entity.Code,
		Name:        entity.Name,
		Channel:     channel,
		Type:        entity.Type.String(),
		Category:    entity.Category,
		Version:     entity.TemplateVersion,
		Status:      m.getTemplateStatus(entity),
		IsDraft:     entity.HasDraft(),
		PublishedAt: entity.PublishedAt,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

// ToSummaryDTOList converts a list of NotificationTemplate entities to a list of TemplateSummaryDTOs.
func (m *TemplateMapper) ToSummaryDTOList(entities []*domain.NotificationTemplate) []dto.TemplateSummaryDTO {
	result := make([]dto.TemplateSummaryDTO, 0, len(entities))
	for _, entity := range entities {
		if entity != nil {
			result = append(result, *m.ToSummaryDTO(entity))
		}
	}
	return result
}

// getTemplateStatus returns the status string for a template.
func (m *TemplateMapper) getTemplateStatus(entity *domain.NotificationTemplate) string {
	if entity.DeletedAt != nil {
		return "archived"
	}
	if !entity.IsActive {
		return "inactive"
	}
	if entity.HasDraft() {
		return "draft"
	}
	return "active"
}

// variablesToDTO converts domain template variables to DTOs.
func (m *TemplateMapper) variablesToDTO(variables []domain.TemplateVariable) []dto.TemplateVariableDTO {
	if len(variables) == 0 {
		return nil
	}

	result := make([]dto.TemplateVariableDTO, len(variables))
	for i, v := range variables {
		result[i] = dto.TemplateVariableDTO{
			Name:         v.Name,
			Type:         v.Type,
			Required:     v.Required,
			DefaultValue: v.DefaultValue,
			Description:  v.Description,
		}
	}
	return result
}

// localizationsToDTO converts domain localizations to DTOs.
func (m *TemplateMapper) localizationsToDTO(localizations map[string]*domain.TemplateLocalization) []dto.LocalizationDTO {
	if len(localizations) == 0 {
		return nil
	}

	result := make([]dto.LocalizationDTO, 0, len(localizations))
	for locale, loc := range localizations {
		if loc == nil {
			continue
		}
		locDTO := dto.LocalizationDTO{
			Locale: locale,
		}
		// Extract content from channel templates
		if loc.EmailTemplate != nil {
			locDTO.Subject = loc.EmailTemplate.Subject
			locDTO.Body = loc.EmailTemplate.Body
			locDTO.HTMLBody = loc.EmailTemplate.HTMLBody
		} else if loc.SMSTemplate != nil {
			locDTO.Body = loc.SMSTemplate.Body
		} else if loc.PushTemplate != nil {
			locDTO.Subject = loc.PushTemplate.Title
			locDTO.Body = loc.PushTemplate.Body
		} else if loc.InAppTemplate != nil {
			locDTO.Subject = loc.InAppTemplate.Title
			locDTO.Body = loc.InAppTemplate.Body
		}
		result = append(result, locDTO)
	}
	return result
}

// VariablesFromDTO converts DTO template variables to domain variables.
func (m *TemplateMapper) VariablesFromDTO(variables []dto.TemplateVariableDTO) []domain.TemplateVariable {
	if len(variables) == 0 {
		return nil
	}

	result := make([]domain.TemplateVariable, len(variables))
	for i, v := range variables {
		result[i] = domain.TemplateVariable{
			Name:         v.Name,
			Type:         v.Type,
			Required:     v.Required,
			DefaultValue: v.DefaultValue,
			Description:  v.Description,
		}
	}
	return result
}

// ToVersionDTO converts a NotificationTemplate to a TemplateVersionDTO.
func (m *TemplateMapper) ToVersionDTO(entity *domain.NotificationTemplate, changeSummary string) *dto.TemplateVersionDTO {
	if entity == nil {
		return nil
	}

	// Extract content
	subject := ""
	body := ""
	htmlBody := ""

	if entity.EmailTemplate != nil {
		subject = entity.EmailTemplate.Subject
		body = entity.EmailTemplate.Body
		htmlBody = entity.EmailTemplate.HTMLBody
	} else if entity.SMSTemplate != nil {
		body = entity.SMSTemplate.Body
	} else if entity.PushTemplate != nil {
		subject = entity.PushTemplate.Title
		body = entity.PushTemplate.Body
	} else if entity.InAppTemplate != nil {
		subject = entity.InAppTemplate.Title
		body = entity.InAppTemplate.Body
	}

	result := &dto.TemplateVersionDTO{
		Version:       entity.TemplateVersion,
		Subject:       subject,
		Body:          body,
		HTMLBody:      htmlBody,
		Variables:     m.variablesToDTO(entity.Variables),
		ChangeSummary: changeSummary,
		PublishedAt:   entity.PublishedAt,
		CreatedAt:     entity.CreatedAt,
	}

	if entity.CreatedBy != nil {
		result.CreatedBy = entity.CreatedBy.String()
	}

	return result
}

// ToVersionDTOList converts a list of templates to a list of TemplateVersionDTOs.
func (m *TemplateMapper) ToVersionDTOList(entities []*domain.NotificationTemplate, changeSummaries map[int]string) []dto.TemplateVersionDTO {
	result := make([]dto.TemplateVersionDTO, 0, len(entities))
	for _, entity := range entities {
		if entity != nil {
			changeSummary := ""
			if changeSummaries != nil {
				changeSummary = changeSummaries[entity.TemplateVersion]
			}
			result = append(result, *m.ToVersionDTO(entity, changeSummary))
		}
	}
	return result
}

// TemplateListToDTO converts a list of templates with pagination info to a TemplateListDTO.
func (m *TemplateMapper) TemplateListToDTO(
	templates []*domain.NotificationTemplate,
	totalCount int64,
	page int,
	pageSize int,
) *dto.TemplateListDTO {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &dto.TemplateListDTO{
		Items:      m.ToSummaryDTOList(templates),
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// RenderResponseFromResult creates a RenderTemplateResponse from rendered content.
func (m *TemplateMapper) RenderResponseFromResult(
	subject string,
	body string,
	htmlBody string,
	textBody string,
) *dto.RenderTemplateResponse {
	return &dto.RenderTemplateResponse{
		Subject:  subject,
		Body:     body,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}
}

// ToValidationResponse creates a ValidateTemplateResponse.
func (m *TemplateMapper) ToValidationResponse(
	valid bool,
	errors []dto.TemplateErrorDTO,
	warnings []dto.TemplateWarningDTO,
	detectedVariables []string,
) *dto.ValidateTemplateResponse {
	return &dto.ValidateTemplateResponse{
		Valid:             valid,
		Errors:            errors,
		Warnings:          warnings,
		DetectedVariables: detectedVariables,
	}
}
