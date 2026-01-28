// Package mapper contains the mappers for converting between domain entities and DTOs.
package mapper

import (
	"github.com/kilang-desa-murni/crm/internal/notification/application/dto"
	"github.com/kilang-desa-murni/crm/internal/notification/domain"
)

// NotificationMapper provides mapping between Notification entities and DTOs.
type NotificationMapper struct{}

// NewNotificationMapper creates a new NotificationMapper.
func NewNotificationMapper() *NotificationMapper {
	return &NotificationMapper{}
}

// ToDTO converts a Notification entity to a NotificationDTO.
func (m *NotificationMapper) ToDTO(entity *domain.Notification) *dto.NotificationDTO {
	if entity == nil {
		return nil
	}

	result := &dto.NotificationDTO{
		ID:          entity.ID.String(),
		TenantID:    entity.TenantID.String(),
		Type:        entity.Type.String(),
		Channel:     entity.Channel.String(),
		Priority:    entity.Priority.String(),
		Status:      entity.Status.String(),
		Subject:     entity.Subject,
		Body:        entity.Body,
		HTMLBody:    entity.HTMLBody,
		Recipient:   m.recipientToDTO(entity),
		Variables:   entity.Data,
		Metadata:    entity.Metadata,
		ScheduledAt: entity.ScheduledAt,
		SentAt:      entity.SentAt,
		DeliveredAt: entity.DeliveredAt,
		ReadAt:      entity.ReadAt,
		FailedAt:    entity.FailedAt,
		CancelledAt: entity.CancelledAt,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
		Version:     entity.Version,
	}

	// Map template ID
	if entity.TemplateID != nil {
		result.TemplateID = entity.TemplateID.String()
	}

	// Map sender from flat fields
	if entity.FromAddress != "" || entity.FromName != "" {
		result.Sender = &dto.SenderDTO{
			Email:   entity.FromAddress,
			Name:    entity.FromName,
			ReplyTo: entity.ReplyTo,
		}
	}

	// Map source event
	if entity.SourceEvent != "" {
		result.SourceEvent = &dto.SourceEventDTO{
			EventType:     entity.SourceEvent,
			AggregateType: entity.SourceEntityType,
		}
		if entity.SourceEntityID != nil {
			result.SourceEvent.AggregateID = entity.SourceEntityID.String()
		}
	}

	// Map delivery info from flat fields
	if entity.Provider != "" || entity.ErrorCode != "" {
		result.DeliveryInfo = &dto.DeliveryInfoDTO{
			Provider:     entity.Provider,
			ProviderID:   entity.ProviderMessageID,
			ErrorCode:    entity.ErrorCode,
			ErrorMessage: entity.ErrorMessage,
		}
	}

	// Map retry info
	if entity.RetryPolicy != nil || entity.AttemptCount > 0 {
		result.RetryInfo = &dto.RetryInfoDTO{
			RetryCount:  entity.AttemptCount,
			LastRetryAt: entity.LastAttemptAt,
			NextRetryAt: entity.NextRetryAt,
		}
		if entity.RetryPolicy != nil {
			result.RetryInfo.MaxRetries = entity.RetryPolicy.MaxAttempts
		}
	}

	// Map tracking info for email
	if entity.Channel == domain.ChannelEmail {
		result.TrackingInfo = &dto.TrackingInfoDTO{
			TrackOpens:  entity.TrackOpens,
			TrackClicks: entity.TrackClicks,
			OpenCount:   entity.OpenCount,
			ClickCount:  entity.ClickCount,
		}
	}

	return result
}

// ToDTOList converts a list of Notification entities to a list of NotificationDTOs.
func (m *NotificationMapper) ToDTOList(entities []*domain.Notification) []dto.NotificationDTO {
	result := make([]dto.NotificationDTO, 0, len(entities))
	for _, entity := range entities {
		if entity != nil {
			result = append(result, *m.ToDTO(entity))
		}
	}
	return result
}

// ToSummaryDTO converts a Notification entity to a NotificationSummaryDTO.
func (m *NotificationMapper) ToSummaryDTO(entity *domain.Notification) *dto.NotificationSummaryDTO {
	if entity == nil {
		return nil
	}

	recipientID := ""
	if entity.RecipientID != nil {
		recipientID = entity.RecipientID.String()
	}

	return &dto.NotificationSummaryDTO{
		ID:          entity.ID.String(),
		TenantID:    entity.TenantID.String(),
		Type:        entity.Type.String(),
		Channel:     entity.Channel.String(),
		Priority:    entity.Priority.String(),
		Status:      entity.Status.String(),
		Subject:     entity.Subject,
		RecipientID: recipientID,
		SentAt:      entity.SentAt,
		CreatedAt:   entity.CreatedAt,
	}
}

// ToSummaryDTOList converts a list of Notification entities to a list of NotificationSummaryDTOs.
func (m *NotificationMapper) ToSummaryDTOList(entities []*domain.Notification) []dto.NotificationSummaryDTO {
	result := make([]dto.NotificationSummaryDTO, 0, len(entities))
	for _, entity := range entities {
		if entity != nil {
			result = append(result, *m.ToSummaryDTO(entity))
		}
	}
	return result
}

// recipientToDTO extracts recipient information from a Notification to a RecipientDTO.
func (m *NotificationMapper) recipientToDTO(entity *domain.Notification) dto.RecipientDTO {
	recipientDTO := dto.RecipientDTO{
		Email:       entity.RecipientEmail,
		Phone:       entity.RecipientPhone,
		DeviceToken: entity.DeviceToken,
		Name:        entity.RecipientName,
	}

	if entity.RecipientID != nil {
		recipientDTO.UserID = entity.RecipientID.String()
		recipientDTO.ID = entity.RecipientID.String()
	}

	// Determine type based on channel
	switch entity.Channel {
	case domain.ChannelEmail:
		recipientDTO.Type = "email"
	case domain.ChannelSMS, domain.ChannelWhatsApp:
		recipientDTO.Type = "phone"
	case domain.ChannelPush:
		recipientDTO.Type = "device"
	case domain.ChannelInApp:
		recipientDTO.Type = "user"
	case domain.ChannelWebhook:
		recipientDTO.Type = "webhook"
	default:
		recipientDTO.Type = "external"
	}

	return recipientDTO
}

// NotificationListToDTO converts a list of notifications with pagination info to a NotificationListDTO.
func (m *NotificationMapper) NotificationListToDTO(
	notifications []*domain.Notification,
	totalCount int64,
	page int,
	pageSize int,
) *dto.NotificationListDTO {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return &dto.NotificationListDTO{
		Items:      m.ToDTOList(notifications),
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// RecipientFromDTO converts a RecipientDTO to a domain Recipient.
func (m *NotificationMapper) RecipientFromDTO(d *dto.RecipientDTO) *domain.Recipient {
	if d == nil {
		return nil
	}
	return domain.NewRecipient().
		WithUserID(d.UserID).
		WithEmail(d.Email).
		WithPhone(d.Phone).
		WithDeviceToken(d.DeviceToken, d.Locale).
		WithName(d.Name).
		WithLocale(d.Locale).
		WithTimezone(d.Timezone)
}
