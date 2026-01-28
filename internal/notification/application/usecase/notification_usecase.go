// Package usecase contains the application use cases for the Notification service.
package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kilang-desa-murni/crm/internal/notification/application"
	"github.com/kilang-desa-murni/crm/internal/notification/application/dto"
	"github.com/kilang-desa-murni/crm/internal/notification/application/mapper"
	"github.com/kilang-desa-murni/crm/internal/notification/application/ports"
	"github.com/kilang-desa-murni/crm/internal/notification/domain"
)

// NotificationUseCase defines the interface for notification use cases.
type NotificationUseCase interface {
	// SendEmail sends an email notification.
	SendEmail(ctx context.Context, req *dto.SendEmailRequest) (*dto.SendNotificationResponse, error)
	// SendSMS sends an SMS notification.
	SendSMS(ctx context.Context, req *dto.SendSMSRequest) (*dto.SendNotificationResponse, error)
	// SendInAppNotification sends an in-app notification.
	SendInAppNotification(ctx context.Context, req *dto.SendInAppRequest) (*dto.SendNotificationResponse, error)
	// SendPushNotification sends a push notification.
	SendPushNotification(ctx context.Context, req *dto.SendPushRequest) (*dto.SendNotificationResponse, error)
	// RetryNotification retries a failed notification.
	RetryNotification(ctx context.Context, req *dto.RetryNotificationRequest) (*dto.RetryNotificationResponse, error)
	// CancelNotification cancels a scheduled notification.
	CancelNotification(ctx context.Context, req *dto.CancelNotificationRequest) (*dto.CancelNotificationResponse, error)
	// GetNotification retrieves a notification by ID.
	GetNotification(ctx context.Context, req *dto.GetNotificationRequest) (*dto.NotificationDTO, error)
	// ListNotifications lists notifications with filtering and pagination.
	ListNotifications(ctx context.Context, req *dto.ListNotificationsRequest) (*dto.NotificationListDTO, error)
}

// notificationUseCase implements the NotificationUseCase interface.
type notificationUseCase struct {
	notificationRepo domain.NotificationRepository
	templateRepo     domain.TemplateRepository

	emailProvider ports.EmailProvider
	smsProvider   ports.SMSProvider
	pushProvider  ports.PushProvider
	inAppProvider ports.InAppProvider

	eventPublisher     ports.EventPublisher
	rateLimiter        ports.RateLimiter
	quotaManager       ports.QuotaManager
	userService        ports.UserService
	scheduler          ports.Scheduler
	suppressionService ports.SuppressionService
	idGenerator        ports.IdGenerator
	timeProvider       ports.TimeProvider
	metrics            ports.MetricsCollector
	logger             ports.Logger

	mapper *mapper.NotificationMapper
}

// NotificationUseCaseConfig holds configuration for the notification use case.
type NotificationUseCaseConfig struct {
	NotificationRepo domain.NotificationRepository
	TemplateRepo     domain.TemplateRepository

	EmailProvider ports.EmailProvider
	SMSProvider   ports.SMSProvider
	PushProvider  ports.PushProvider
	InAppProvider ports.InAppProvider

	EventPublisher     ports.EventPublisher
	RateLimiter        ports.RateLimiter
	QuotaManager       ports.QuotaManager
	UserService        ports.UserService
	Scheduler          ports.Scheduler
	SuppressionService ports.SuppressionService
	IdGenerator        ports.IdGenerator
	TimeProvider       ports.TimeProvider
	Metrics            ports.MetricsCollector
	Logger             ports.Logger
}

// NewNotificationUseCase creates a new NotificationUseCase.
func NewNotificationUseCase(cfg NotificationUseCaseConfig) NotificationUseCase {
	return &notificationUseCase{
		notificationRepo: cfg.NotificationRepo,
		templateRepo:     cfg.TemplateRepo,

		emailProvider: cfg.EmailProvider,
		smsProvider:   cfg.SMSProvider,
		pushProvider:  cfg.PushProvider,
		inAppProvider: cfg.InAppProvider,

		eventPublisher:     cfg.EventPublisher,
		rateLimiter:        cfg.RateLimiter,
		quotaManager:       cfg.QuotaManager,
		userService:        cfg.UserService,
		scheduler:          cfg.Scheduler,
		suppressionService: cfg.SuppressionService,
		idGenerator:        cfg.IdGenerator,
		timeProvider:       cfg.TimeProvider,
		metrics:            cfg.Metrics,
		logger:             cfg.Logger,

		mapper: mapper.NewNotificationMapper(),
	}
}

// SendEmail sends an email notification.
func (uc *notificationUseCase) SendEmail(ctx context.Context, req *dto.SendEmailRequest) (*dto.SendNotificationResponse, error) {
	uc.logger.WithContext(ctx).Info("SendEmail started", map[string]interface{}{
		"tenant_id": req.TenantID,
		"to":        req.To,
	})

	// Validate request
	if err := uc.validateSendEmailRequest(req); err != nil {
		return nil, err
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Check rate limit
	if allowed, err := uc.rateLimiter.Allow(ctx, fmt.Sprintf("email:%s", req.TenantID), 100, time.Minute); err != nil {
		return nil, application.NewInternalError("failed to check rate limit", err)
	} else if !allowed {
		return nil, application.NewRateLimitExceededError("email", 100, "minute")
	}

	// Check quota
	if allowed, err := uc.quotaManager.CheckQuota(ctx, req.TenantID, "email"); err != nil {
		return nil, application.NewInternalError("failed to check quota", err)
	} else if !allowed {
		return nil, application.NewQuotaExceededError("email", 0, "month")
	}

	// Parse notification type
	notificationType, err := domain.ParseType(req.Type)
	if err != nil {
		return nil, application.NewInvalidInputError(fmt.Sprintf("invalid notification type: %s", req.Type))
	}

	// Resolve template if provided
	subject := req.Subject
	body := req.Body
	htmlBody := req.HTMLBody

	if req.TemplateID != "" {
		renderedContent, err := uc.renderEmailTemplate(ctx, req.TenantID, req.TemplateID, req.Variables)
		if err != nil {
			return nil, err
		}
		subject = renderedContent.Subject
		body = renderedContent.Body
		htmlBody = renderedContent.HTMLBody
	}

	// Create notification entity
	notification, err := domain.NewNotification(
		tenantID,
		notificationType,
		domain.ChannelEmail,
		body,
	)
	if err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set email recipient
	if err := notification.SetRecipientEmail(req.To[0], ""); err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set subject
	if err := notification.SetSubject(subject); err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set HTML body if provided
	if htmlBody != "" {
		notification.SetHTMLBody(htmlBody)
	}

	// Set sender info
	if req.From != "" {
		if err := notification.SetFromAddress(req.From, req.FromName); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}
	if req.ReplyTo != "" {
		if err := notification.SetReplyTo(req.ReplyTo); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Add CC/BCC
	for _, cc := range req.CC {
		if err := notification.AddCC(cc); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}
	for _, bcc := range req.BCC {
		if err := notification.AddBCC(bcc); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Set priority
	if req.Priority != "" {
		priority, err := domain.ParsePriority(req.Priority)
		if err != nil {
			return nil, application.NewInvalidInputError(fmt.Sprintf("invalid priority: %s", req.Priority))
		}
		if err := notification.SetPriority(priority); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Set tracking options
	if req.TrackOpens != nil || req.TrackClicks != nil {
		trackOpens := req.TrackOpens != nil && *req.TrackOpens
		trackClicks := req.TrackClicks != nil && *req.TrackClicks
		notification.EnableTracking(trackOpens, trackClicks)
	}

	// Set template data
	if req.Variables != nil {
		notification.SetData(req.Variables)
	}

	// Set template reference
	if req.TemplateID != "" {
		templateID, err := uuid.Parse(req.TemplateID)
		if err == nil {
			notification.SetTemplate(templateID, "")
		}
	}

	// Set source event
	if req.SourceEvent != nil {
		var sourceEntityID *uuid.UUID
		if req.SourceEvent.AggregateID != "" {
			id, err := uuid.Parse(req.SourceEvent.AggregateID)
			if err == nil {
				sourceEntityID = &id
			}
		}
		notification.SetSourceEvent(req.SourceEvent.EventType, sourceEntityID, req.SourceEvent.AggregateType)
	}

	// Check if scheduled for future
	if req.ScheduledAt != nil && req.ScheduledAt.After(uc.timeProvider.NowUTC()) {
		if err := notification.Schedule(*req.ScheduledAt); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}

		// Save and schedule
		if err := uc.saveAndScheduleNotification(ctx, notification); err != nil {
			return nil, err
		}

		return &dto.SendNotificationResponse{
			NotificationID: notification.ID.String(),
			Status:         notification.Status.String(),
			ScheduledAt:    req.ScheduledAt,
			Message:        "Notification scheduled for future delivery",
		}, nil
	}

	// Queue for immediate delivery
	if err := notification.Queue(); err != nil {
		return nil, application.NewInternalError("failed to queue notification", err)
	}

	// Create notification
	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return nil, application.NewInternalError("failed to save notification", err)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, notification)

	// Send email asynchronously
	go uc.deliverEmail(context.Background(), notification, req)

	uc.metrics.IncrementCounter(ctx, "notification.email.queued", map[string]string{
		"tenant_id": req.TenantID,
		"type":      req.Type,
	})

	return &dto.SendNotificationResponse{
		NotificationID: notification.ID.String(),
		Status:         notification.Status.String(),
		Message:        "Email notification queued for delivery",
	}, nil
}

// SendSMS sends an SMS notification.
func (uc *notificationUseCase) SendSMS(ctx context.Context, req *dto.SendSMSRequest) (*dto.SendNotificationResponse, error) {
	uc.logger.WithContext(ctx).Info("SendSMS started", map[string]interface{}{
		"tenant_id": req.TenantID,
		"to":        req.To,
	})

	// Validate request
	if err := uc.validateSendSMSRequest(req); err != nil {
		return nil, err
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Check rate limit
	if allowed, err := uc.rateLimiter.Allow(ctx, fmt.Sprintf("sms:%s", req.TenantID), 50, time.Minute); err != nil {
		return nil, application.NewInternalError("failed to check rate limit", err)
	} else if !allowed {
		return nil, application.NewRateLimitExceededError("sms", 50, "minute")
	}

	// Check suppression list
	if suppressed, err := uc.suppressionService.IsSuppressed(ctx, req.TenantID, "sms", req.To); err != nil {
		return nil, application.NewInternalError("failed to check suppression list", err)
	} else if suppressed {
		return nil, application.NewSMSOptedOutError(req.To)
	}

	// Parse notification type
	notificationType, err := domain.ParseType(req.Type)
	if err != nil {
		return nil, application.NewInvalidInputError(fmt.Sprintf("invalid notification type: %s", req.Type))
	}

	// Resolve template if provided
	body := req.Body
	if req.TemplateID != "" {
		renderedContent, err := uc.renderSMSTemplate(ctx, req.TenantID, req.TemplateID, req.Variables)
		if err != nil {
			return nil, err
		}
		body = renderedContent.Body
	}

	// Create notification entity
	notification, err := domain.NewNotification(
		tenantID,
		notificationType,
		domain.ChannelSMS,
		body,
	)
	if err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set SMS recipient
	if err := notification.SetRecipientPhone(req.To, ""); err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set priority
	if req.Priority != "" {
		priority, err := domain.ParsePriority(req.Priority)
		if err != nil {
			return nil, application.NewInvalidInputError(fmt.Sprintf("invalid priority: %s", req.Priority))
		}
		if err := notification.SetPriority(priority); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Set template data
	if req.Variables != nil {
		notification.SetData(req.Variables)
	}

	// Set template reference
	if req.TemplateID != "" {
		templateID, err := uuid.Parse(req.TemplateID)
		if err == nil {
			notification.SetTemplate(templateID, "")
		}
	}

	// Set source event
	if req.SourceEvent != nil {
		var sourceEntityID *uuid.UUID
		if req.SourceEvent.AggregateID != "" {
			id, err := uuid.Parse(req.SourceEvent.AggregateID)
			if err == nil {
				sourceEntityID = &id
			}
		}
		notification.SetSourceEvent(req.SourceEvent.EventType, sourceEntityID, req.SourceEvent.AggregateType)
	}

	// Check if scheduled for future
	if req.ScheduledAt != nil && req.ScheduledAt.After(uc.timeProvider.NowUTC()) {
		if err := notification.Schedule(*req.ScheduledAt); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}

		if err := uc.saveAndScheduleNotification(ctx, notification); err != nil {
			return nil, err
		}

		return &dto.SendNotificationResponse{
			NotificationID: notification.ID.String(),
			Status:         notification.Status.String(),
			ScheduledAt:    req.ScheduledAt,
			Message:        "SMS notification scheduled for future delivery",
		}, nil
	}

	// Queue for immediate delivery
	if err := notification.Queue(); err != nil {
		return nil, application.NewInternalError("failed to queue notification", err)
	}

	// Create notification
	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return nil, application.NewInternalError("failed to save notification", err)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, notification)

	// Send SMS asynchronously
	go uc.deliverSMS(context.Background(), notification, req)

	uc.metrics.IncrementCounter(ctx, "notification.sms.queued", map[string]string{
		"tenant_id": req.TenantID,
		"type":      req.Type,
	})

	return &dto.SendNotificationResponse{
		NotificationID: notification.ID.String(),
		Status:         notification.Status.String(),
		Message:        "SMS notification queued for delivery",
	}, nil
}

// SendInAppNotification sends an in-app notification.
func (uc *notificationUseCase) SendInAppNotification(ctx context.Context, req *dto.SendInAppRequest) (*dto.SendNotificationResponse, error) {
	uc.logger.WithContext(ctx).Info("SendInAppNotification started", map[string]interface{}{
		"tenant_id": req.TenantID,
		"user_id":   req.UserID,
	})

	// Validate request
	if err := uc.validateSendInAppRequest(req); err != nil {
		return nil, err
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid user ID format")
	}

	// Verify user exists and is active
	user, err := uc.userService.GetUser(ctx, req.UserID)
	if err != nil {
		return nil, application.NewUserNotFoundError(req.UserID)
	}
	if !user.IsActive {
		return nil, application.NewAppError(application.ErrCodeUserInactive, "user is inactive")
	}

	// Parse notification type
	notificationType, err := domain.ParseType(req.Type)
	if err != nil {
		return nil, application.NewInvalidInputError(fmt.Sprintf("invalid notification type: %s", req.Type))
	}

	// Resolve template if provided
	title := req.Title
	body := req.Body
	if req.TemplateID != "" {
		renderedContent, err := uc.renderInAppTemplate(ctx, req.TenantID, req.TemplateID, req.Variables, user.Locale)
		if err != nil {
			return nil, err
		}
		title = renderedContent.Title
		body = renderedContent.Body
	}

	// Create notification entity
	notification, err := domain.NewNotification(
		tenantID,
		notificationType,
		domain.ChannelInApp,
		body,
	)
	if err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set in-app recipient
	if err := notification.SetRecipientUser(userID, user.DisplayName); err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set subject (title)
	if title != "" {
		if err := notification.SetSubject(title); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Set priority
	if req.Priority != "" {
		priority, err := domain.ParsePriority(req.Priority)
		if err != nil {
			return nil, application.NewInvalidInputError(fmt.Sprintf("invalid priority: %s", req.Priority))
		}
		if err := notification.SetPriority(priority); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Set template data and additional metadata
	data := make(map[string]interface{})
	if req.Variables != nil {
		for k, v := range req.Variables {
			data[k] = v
		}
	}
	if req.ActionURL != "" {
		data["action_url"] = req.ActionURL
	}
	if req.ActionText != "" {
		data["action_text"] = req.ActionText
	}
	if req.ImageURL != "" {
		data["image_url"] = req.ImageURL
	}
	if req.Category != "" {
		data["category"] = req.Category
	}
	notification.SetData(data)

	// Set template reference
	if req.TemplateID != "" {
		templateID, err := uuid.Parse(req.TemplateID)
		if err == nil {
			notification.SetTemplate(templateID, "")
		}
	}

	// Set source event
	if req.SourceEvent != nil {
		var sourceEntityID *uuid.UUID
		if req.SourceEvent.AggregateID != "" {
			id, err := uuid.Parse(req.SourceEvent.AggregateID)
			if err == nil {
				sourceEntityID = &id
			}
		}
		notification.SetSourceEvent(req.SourceEvent.EventType, sourceEntityID, req.SourceEvent.AggregateType)
	}

	// Check if scheduled for future
	if req.ScheduledAt != nil && req.ScheduledAt.After(uc.timeProvider.NowUTC()) {
		if err := notification.Schedule(*req.ScheduledAt); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}

		if err := uc.saveAndScheduleNotification(ctx, notification); err != nil {
			return nil, err
		}

		return &dto.SendNotificationResponse{
			NotificationID: notification.ID.String(),
			Status:         notification.Status.String(),
			ScheduledAt:    req.ScheduledAt,
			Message:        "In-app notification scheduled for future delivery",
		}, nil
	}

	// Queue for immediate delivery
	if err := notification.Queue(); err != nil {
		return nil, application.NewInternalError("failed to queue notification", err)
	}

	// Create notification
	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return nil, application.NewInternalError("failed to save notification", err)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, notification)

	// Send in-app notification asynchronously
	go uc.deliverInApp(context.Background(), notification, req)

	uc.metrics.IncrementCounter(ctx, "notification.in_app.queued", map[string]string{
		"tenant_id": req.TenantID,
		"type":      req.Type,
	})

	return &dto.SendNotificationResponse{
		NotificationID: notification.ID.String(),
		Status:         notification.Status.String(),
		Message:        "In-app notification queued for delivery",
	}, nil
}

// SendPushNotification sends a push notification.
func (uc *notificationUseCase) SendPushNotification(ctx context.Context, req *dto.SendPushRequest) (*dto.SendNotificationResponse, error) {
	uc.logger.WithContext(ctx).Info("SendPushNotification started", map[string]interface{}{
		"tenant_id": req.TenantID,
		"user_id":   req.UserID,
	})

	// Validate request
	if err := uc.validateSendPushRequest(req); err != nil {
		return nil, err
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid user ID format")
	}

	// Verify user exists and get device tokens
	user, err := uc.userService.GetUser(ctx, req.UserID)
	if err != nil {
		return nil, application.NewUserNotFoundError(req.UserID)
	}
	if !user.IsActive {
		return nil, application.NewAppError(application.ErrCodeUserInactive, "user is inactive")
	}

	// Determine device token
	deviceToken := req.DeviceToken
	if deviceToken == "" && len(user.DeviceTokens) > 0 {
		deviceToken = user.DeviceTokens[0]
	}
	if deviceToken == "" {
		return nil, application.NewDeviceNotRegisteredError(req.UserID)
	}

	// Parse notification type
	notificationType, err := domain.ParseType(req.Type)
	if err != nil {
		return nil, application.NewInvalidInputError(fmt.Sprintf("invalid notification type: %s", req.Type))
	}

	// Resolve template if provided
	title := req.Title
	body := req.Body
	if req.TemplateID != "" {
		renderedContent, err := uc.renderPushTemplate(ctx, req.TenantID, req.TemplateID, req.Variables, user.Locale)
		if err != nil {
			return nil, err
		}
		title = renderedContent.Title
		body = renderedContent.Body
	}

	// Create notification entity
	notification, err := domain.NewNotification(
		tenantID,
		notificationType,
		domain.ChannelPush,
		body,
	)
	if err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set push recipient
	if err := notification.SetRecipientDevice(userID, deviceToken, user.DisplayName); err != nil {
		return nil, application.NewInvalidInputError(err.Error())
	}

	// Set subject (title)
	if title != "" {
		if err := notification.SetSubject(title); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Set priority
	if req.Priority != "" {
		priority, err := domain.ParsePriority(req.Priority)
		if err != nil {
			return nil, application.NewInvalidInputError(fmt.Sprintf("invalid priority: %s", req.Priority))
		}
		if err := notification.SetPriority(priority); err != nil {
			return nil, application.NewInvalidInputError(err.Error())
		}
	}

	// Set template data
	data := make(map[string]interface{})
	if req.Variables != nil {
		for k, v := range req.Variables {
			data[k] = v
		}
	}
	if req.Data != nil {
		for k, v := range req.Data {
			data[k] = v
		}
	}
	notification.SetData(data)

	// Set template reference
	if req.TemplateID != "" {
		templateID, err := uuid.Parse(req.TemplateID)
		if err == nil {
			notification.SetTemplate(templateID, "")
		}
	}

	// Set source event
	if req.SourceEvent != nil {
		var sourceEntityID *uuid.UUID
		if req.SourceEvent.AggregateID != "" {
			id, err := uuid.Parse(req.SourceEvent.AggregateID)
			if err == nil {
				sourceEntityID = &id
			}
		}
		notification.SetSourceEvent(req.SourceEvent.EventType, sourceEntityID, req.SourceEvent.AggregateType)
	}

	// Queue for immediate delivery
	if err := notification.Queue(); err != nil {
		return nil, application.NewInternalError("failed to queue notification", err)
	}

	// Create notification
	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return nil, application.NewInternalError("failed to save notification", err)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, notification)

	// Send push notification asynchronously
	go uc.deliverPush(context.Background(), notification, req, deviceToken)

	uc.metrics.IncrementCounter(ctx, "notification.push.queued", map[string]string{
		"tenant_id": req.TenantID,
		"type":      req.Type,
	})

	return &dto.SendNotificationResponse{
		NotificationID: notification.ID.String(),
		Status:         notification.Status.String(),
		Message:        "Push notification queued for delivery",
	}, nil
}

// RetryNotification retries a failed notification.
func (uc *notificationUseCase) RetryNotification(ctx context.Context, req *dto.RetryNotificationRequest) (*dto.RetryNotificationResponse, error) {
	uc.logger.WithContext(ctx).Info("RetryNotification started", map[string]interface{}{
		"notification_id": req.NotificationID,
	})

	// Parse notification ID
	notificationID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid notification ID format")
	}

	// Get notification
	notification, err := uc.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return nil, application.NewNotificationNotFoundError(req.NotificationID)
	}

	// Check if retry is allowed
	if !req.Force && !notification.CanRetry() {
		return nil, application.NewCannotRetryError(req.NotificationID, notification.Status.String())
	}

	// Calculate next retry time
	nextRetryAt := time.Now().UTC().Add(time.Duration(notification.GetNextRetryInterval()) * time.Second)

	// Mark for retry
	if err := notification.MarkRetrying(nextRetryAt); err != nil {
		return nil, application.NewInternalError("failed to mark notification for retry", err)
	}

	// Save updated notification
	if err := uc.notificationRepo.Update(ctx, notification); err != nil {
		return nil, application.NewInternalError("failed to update notification", err)
	}

	// Retry delivery based on channel
	switch notification.Channel {
	case domain.ChannelEmail:
		go uc.retryEmailDelivery(context.Background(), notification)
	case domain.ChannelSMS:
		go uc.retrySMSDelivery(context.Background(), notification)
	case domain.ChannelInApp:
		go uc.retryInAppDelivery(context.Background(), notification)
	case domain.ChannelPush:
		go uc.retryPushDelivery(context.Background(), notification)
	}

	uc.metrics.IncrementCounter(ctx, "notification.retry.initiated", map[string]string{
		"channel": notification.Channel.String(),
	})

	return &dto.RetryNotificationResponse{
		NotificationID: req.NotificationID,
		Status:         notification.Status.String(),
		RetryCount:     notification.AttemptCount,
		Message:        fmt.Sprintf("Notification retry initiated (attempt %d)", notification.AttemptCount),
	}, nil
}

// CancelNotification cancels a scheduled notification.
func (uc *notificationUseCase) CancelNotification(ctx context.Context, req *dto.CancelNotificationRequest) (*dto.CancelNotificationResponse, error) {
	uc.logger.WithContext(ctx).Info("CancelNotification started", map[string]interface{}{
		"notification_id": req.NotificationID,
	})

	// Parse notification ID
	notificationID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid notification ID format")
	}

	// Get notification
	notification, err := uc.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return nil, application.NewNotificationNotFoundError(req.NotificationID)
	}

	// Cancel the notification
	if err := notification.Cancel(); err != nil {
		return nil, application.NewCannotCancelError(req.NotificationID, notification.Status.String())
	}

	// Save updated notification
	if err := uc.notificationRepo.Update(ctx, notification); err != nil {
		return nil, application.NewInternalError("failed to update notification", err)
	}

	// Cancel scheduled delivery if applicable
	if notification.ScheduledAt != nil {
		if err := uc.scheduler.Cancel(ctx, req.NotificationID); err != nil {
			uc.logger.WithContext(ctx).Warn("failed to cancel scheduled delivery", map[string]interface{}{
				"notification_id": req.NotificationID,
				"error":           err.Error(),
			})
		}
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, notification)

	uc.metrics.IncrementCounter(ctx, "notification.cancelled", map[string]string{
		"channel": notification.Channel.String(),
	})

	return &dto.CancelNotificationResponse{
		NotificationID: req.NotificationID,
		Status:         notification.Status.String(),
		Message:        "Notification cancelled successfully",
	}, nil
}

// GetNotification retrieves a notification by ID.
func (uc *notificationUseCase) GetNotification(ctx context.Context, req *dto.GetNotificationRequest) (*dto.NotificationDTO, error) {
	// Parse notification ID
	notificationID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid notification ID format")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	notification, err := uc.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return nil, application.NewNotificationNotFoundError(req.NotificationID)
	}

	// Verify tenant access
	if notification.TenantID != tenantID {
		return nil, application.NewForbiddenError("access denied to this notification")
	}

	return uc.mapper.ToDTO(notification), nil
}

// ListNotifications lists notifications with filtering and pagination.
func (uc *notificationUseCase) ListNotifications(ctx context.Context, req *dto.ListNotificationsRequest) (*dto.NotificationListDTO, error) {
	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid tenant ID format")
	}

	// Build filter
	filter := domain.NotificationFilter{
		TenantID: &tenantID,
		Offset:   (req.Page - 1) * req.PageSize,
		Limit:    req.PageSize,
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

	if req.Status != "" {
		status, err := domain.ParseStatus(req.Status)
		if err == nil {
			filter.Statuses = []domain.NotificationStatus{status}
		}
	}

	if req.RecipientID != "" {
		recipientID, err := uuid.Parse(req.RecipientID)
		if err == nil {
			filter.RecipientID = &recipientID
		}
	}

	// Execute query
	notificationList, err := uc.notificationRepo.List(ctx, filter)
	if err != nil {
		return nil, application.NewInternalError("failed to list notifications", err)
	}

	return uc.mapper.NotificationListToDTO(notificationList.Notifications, notificationList.Total, req.Page, req.PageSize), nil
}

// === Private helper methods ===

func (uc *notificationUseCase) validateSendEmailRequest(req *dto.SendEmailRequest) error {
	if req.TenantID == "" {
		return application.NewValidationError("tenant_id is required")
	}
	if len(req.To) == 0 {
		return application.NewValidationError("at least one recipient email is required")
	}
	if req.Subject == "" && req.TemplateID == "" {
		return application.NewValidationError("subject is required when not using a template")
	}
	return nil
}

func (uc *notificationUseCase) validateSendSMSRequest(req *dto.SendSMSRequest) error {
	if req.TenantID == "" {
		return application.NewValidationError("tenant_id is required")
	}
	if req.To == "" {
		return application.NewValidationError("recipient phone number is required")
	}
	if req.Body == "" && req.TemplateID == "" {
		return application.NewValidationError("body is required when not using a template")
	}
	return nil
}

func (uc *notificationUseCase) validateSendInAppRequest(req *dto.SendInAppRequest) error {
	if req.TenantID == "" {
		return application.NewValidationError("tenant_id is required")
	}
	if req.UserID == "" {
		return application.NewValidationError("user_id is required")
	}
	if req.Title == "" && req.TemplateID == "" {
		return application.NewValidationError("title is required when not using a template")
	}
	if req.Body == "" && req.TemplateID == "" {
		return application.NewValidationError("body is required when not using a template")
	}
	return nil
}

func (uc *notificationUseCase) validateSendPushRequest(req *dto.SendPushRequest) error {
	if req.TenantID == "" {
		return application.NewValidationError("tenant_id is required")
	}
	if req.UserID == "" {
		return application.NewValidationError("user_id is required")
	}
	if req.Title == "" && req.TemplateID == "" {
		return application.NewValidationError("title is required when not using a template")
	}
	return nil
}

func (uc *notificationUseCase) renderEmailTemplate(ctx context.Context, tenantID, templateID string, variables map[string]interface{}) (*domain.RenderedEmail, error) {
	tid, err := uuid.Parse(templateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	template, err := uc.templateRepo.FindByID(ctx, tid)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(templateID)
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	if template.TenantID != tenantUUID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	rendered, err := template.RenderEmail(variables, template.DefaultLocale)
	if err != nil {
		return nil, application.NewTemplateRenderFailedError(templateID, err.Error())
	}

	return rendered, nil
}

func (uc *notificationUseCase) renderSMSTemplate(ctx context.Context, tenantID, templateID string, variables map[string]interface{}) (*domain.RenderedSMS, error) {
	tid, err := uuid.Parse(templateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	template, err := uc.templateRepo.FindByID(ctx, tid)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(templateID)
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	if template.TenantID != tenantUUID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	rendered, err := template.RenderSMS(variables, template.DefaultLocale)
	if err != nil {
		return nil, application.NewTemplateRenderFailedError(templateID, err.Error())
	}

	return rendered, nil
}

func (uc *notificationUseCase) renderInAppTemplate(ctx context.Context, tenantID, templateID string, variables map[string]interface{}, locale string) (*domain.RenderedInApp, error) {
	tid, err := uuid.Parse(templateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	template, err := uc.templateRepo.FindByID(ctx, tid)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(templateID)
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	if template.TenantID != tenantUUID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	if locale == "" {
		locale = template.DefaultLocale
	}

	rendered, err := template.RenderInApp(variables, locale)
	if err != nil {
		return nil, application.NewTemplateRenderFailedError(templateID, err.Error())
	}

	return rendered, nil
}

func (uc *notificationUseCase) renderPushTemplate(ctx context.Context, tenantID, templateID string, variables map[string]interface{}, locale string) (*domain.RenderedPush, error) {
	tid, err := uuid.Parse(templateID)
	if err != nil {
		return nil, application.NewInvalidInputError("invalid template ID format")
	}

	template, err := uc.templateRepo.FindByID(ctx, tid)
	if err != nil {
		return nil, application.NewTemplateNotFoundError(templateID)
	}

	tenantUUID, _ := uuid.Parse(tenantID)
	if template.TenantID != tenantUUID {
		return nil, application.NewForbiddenError("access denied to this template")
	}

	if locale == "" {
		locale = template.DefaultLocale
	}

	rendered, err := template.RenderPush(variables, locale)
	if err != nil {
		return nil, application.NewTemplateRenderFailedError(templateID, err.Error())
	}

	return rendered, nil
}

func (uc *notificationUseCase) saveAndScheduleNotification(ctx context.Context, notification *domain.Notification) error {
	// Create notification
	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return application.NewInternalError("failed to save notification", err)
	}

	// Schedule for future delivery
	if notification.ScheduledAt != nil {
		if err := uc.scheduler.Schedule(ctx, notification.ID.String(), *notification.ScheduledAt); err != nil {
			return application.NewInternalError("failed to schedule notification", err)
		}
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, notification)

	return nil
}

func (uc *notificationUseCase) publishDomainEvents(ctx context.Context, notification *domain.Notification) {
	events := notification.GetDomainEvents()
	for _, event := range events {
		if err := uc.eventPublisher.Publish(ctx, event); err != nil {
			uc.logger.WithContext(ctx).Error("failed to publish domain event", err, map[string]interface{}{
				"event_type":      event.EventType(),
				"notification_id": notification.ID.String(),
			})
		}
	}
	notification.ClearDomainEvents()
}

func (uc *notificationUseCase) deliverEmail(ctx context.Context, notification *domain.Notification, req *dto.SendEmailRequest) {
	// Mark as sending
	if err := notification.MarkSending(); err != nil {
		uc.logger.WithContext(ctx).Error("failed to mark notification as sending", err, nil)
		return
	}
	_ = uc.notificationRepo.Update(ctx, notification)

	// Build email request
	emailReq := ports.EmailRequest{
		MessageID:   notification.ID.String(),
		From:        notification.FromAddress,
		ReplyTo:     notification.ReplyTo,
		To:          req.To,
		CC:          notification.CC,
		BCC:         notification.BCC,
		Subject:     notification.Subject,
		HTMLBody:    notification.HTMLBody,
		TextBody:    notification.Body,
		TrackOpens:  notification.TrackOpens,
		TrackClicks: notification.TrackClicks,
	}

	// Add attachments
	if len(req.Attachments) > 0 {
		emailReq.Attachments = make([]ports.EmailAttachment, len(req.Attachments))
		for i, att := range req.Attachments {
			content, _ := base64.StdEncoding.DecodeString(att.Content)
			emailReq.Attachments[i] = ports.EmailAttachment{
				Filename:    att.Filename,
				ContentType: att.ContentType,
				Content:     content,
				ContentID:   att.ContentID,
			}
		}
	}

	// Send email
	response, err := uc.emailProvider.SendEmail(ctx, emailReq)
	uc.handleDeliveryResult(ctx, notification, err, response)
}

func (uc *notificationUseCase) deliverSMS(ctx context.Context, notification *domain.Notification, req *dto.SendSMSRequest) {
	// Mark as sending
	if err := notification.MarkSending(); err != nil {
		uc.logger.WithContext(ctx).Error("failed to mark notification as sending", err, nil)
		return
	}
	_ = uc.notificationRepo.Update(ctx, notification)

	// Build SMS request
	smsReq := ports.SMSRequest{
		MessageID: notification.ID.String(),
		From:      req.From,
		To:        notification.RecipientPhone,
		Body:      notification.Body,
	}

	// Send SMS
	response, err := uc.smsProvider.SendSMS(ctx, smsReq)
	uc.handleSMSDeliveryResult(ctx, notification, err, response)
}

func (uc *notificationUseCase) deliverInApp(ctx context.Context, notification *domain.Notification, req *dto.SendInAppRequest) {
	// Mark as sending
	if err := notification.MarkSending(); err != nil {
		uc.logger.WithContext(ctx).Error("failed to mark notification as sending", err, nil)
		return
	}
	_ = uc.notificationRepo.Update(ctx, notification)

	// Build in-app request
	inAppReq := ports.InAppRequest{
		MessageID:  notification.ID.String(),
		UserID:     notification.RecipientID.String(),
		Title:      notification.Subject,
		Body:       notification.Body,
		Category:   req.Category,
		Priority:   notification.Priority.String(),
		ActionURL:  req.ActionURL,
		ActionText: req.ActionText,
		ImageURL:   req.ImageURL,
	}

	// Send in-app notification
	response, err := uc.inAppProvider.SendInApp(ctx, inAppReq)
	uc.handleInAppDeliveryResult(ctx, notification, err, response)
}

func (uc *notificationUseCase) deliverPush(ctx context.Context, notification *domain.Notification, req *dto.SendPushRequest, deviceToken string) {
	// Mark as sending
	if err := notification.MarkSending(); err != nil {
		uc.logger.WithContext(ctx).Error("failed to mark notification as sending", err, nil)
		return
	}
	_ = uc.notificationRepo.Update(ctx, notification)

	// Build push request
	pushReq := ports.PushRequest{
		MessageID:   notification.ID.String(),
		DeviceToken: deviceToken,
		Platform:    req.Platform,
		Title:       notification.Subject,
		Body:        notification.Body,
		ImageURL:    req.ImageURL,
		Data:        req.Data,
		Badge:       req.Badge,
		Sound:       req.Sound,
		ClickAction: req.ClickAction,
		Category:    req.Category,
		CollapseKey: req.CollapseKey,
	}

	// Send push notification
	response, err := uc.pushProvider.SendPush(ctx, pushReq)
	uc.handlePushDeliveryResult(ctx, notification, err, response)
}

func (uc *notificationUseCase) handleDeliveryResult(ctx context.Context, notification *domain.Notification, err error, response *ports.EmailResponse) {
	if err != nil {
		_ = notification.MarkFailed("DELIVERY_FAILED", err.Error(), "")
		notification.Provider = "email"
	} else if response != nil {
		_ = notification.MarkSent(response.ProviderID)
		notification.Provider = response.Provider
		notification.ProviderMessageID = response.ProviderID
	}

	// Save updated notification
	if saveErr := uc.notificationRepo.Update(ctx, notification); saveErr != nil {
		uc.logger.WithContext(ctx).Error("failed to update notification after delivery", saveErr, nil)
	}

	// Publish domain events
	uc.publishDomainEvents(ctx, notification)
}

func (uc *notificationUseCase) handleSMSDeliveryResult(ctx context.Context, notification *domain.Notification, err error, response *ports.SMSResponse) {
	if err != nil {
		_ = notification.MarkFailed("DELIVERY_FAILED", err.Error(), "")
		notification.Provider = "sms"
	} else if response != nil {
		_ = notification.MarkSent(response.ProviderID)
		notification.Provider = response.Provider
		notification.ProviderMessageID = response.ProviderID
	}

	if saveErr := uc.notificationRepo.Update(ctx, notification); saveErr != nil {
		uc.logger.WithContext(ctx).Error("failed to update notification after delivery", saveErr, nil)
	}
	uc.publishDomainEvents(ctx, notification)
}

func (uc *notificationUseCase) handleInAppDeliveryResult(ctx context.Context, notification *domain.Notification, err error, response *ports.InAppResponse) {
	if err != nil {
		_ = notification.MarkFailed("DELIVERY_FAILED", err.Error(), "")
		notification.Provider = "in_app"
	} else if response != nil {
		_ = notification.MarkSent(response.MessageID)
		notification.Provider = "in_app"
	}

	if saveErr := uc.notificationRepo.Update(ctx, notification); saveErr != nil {
		uc.logger.WithContext(ctx).Error("failed to update notification after delivery", saveErr, nil)
	}
	uc.publishDomainEvents(ctx, notification)
}

func (uc *notificationUseCase) handlePushDeliveryResult(ctx context.Context, notification *domain.Notification, err error, response *ports.PushResponse) {
	if err != nil {
		_ = notification.MarkFailed("DELIVERY_FAILED", err.Error(), "")
		notification.Provider = "push"
	} else if response != nil {
		_ = notification.MarkSent(response.ProviderID)
		notification.Provider = response.Provider
		notification.ProviderMessageID = response.ProviderID
	}

	if saveErr := uc.notificationRepo.Update(ctx, notification); saveErr != nil {
		uc.logger.WithContext(ctx).Error("failed to update notification after delivery", saveErr, nil)
	}
	uc.publishDomainEvents(ctx, notification)
}

func (uc *notificationUseCase) retryEmailDelivery(ctx context.Context, notification *domain.Notification) {
	emailReq := ports.EmailRequest{
		MessageID:   notification.ID.String(),
		To:          []string{notification.RecipientEmail},
		From:        notification.FromAddress,
		ReplyTo:     notification.ReplyTo,
		Subject:     notification.Subject,
		HTMLBody:    notification.HTMLBody,
		TextBody:    notification.Body,
		TrackOpens:  notification.TrackOpens,
		TrackClicks: notification.TrackClicks,
	}

	response, err := uc.emailProvider.SendEmail(ctx, emailReq)
	uc.handleDeliveryResult(ctx, notification, err, response)
}

func (uc *notificationUseCase) retrySMSDelivery(ctx context.Context, notification *domain.Notification) {
	smsReq := ports.SMSRequest{
		MessageID: notification.ID.String(),
		To:        notification.RecipientPhone,
		Body:      notification.Body,
	}

	response, err := uc.smsProvider.SendSMS(ctx, smsReq)
	uc.handleSMSDeliveryResult(ctx, notification, err, response)
}

func (uc *notificationUseCase) retryInAppDelivery(ctx context.Context, notification *domain.Notification) {
	inAppReq := ports.InAppRequest{
		MessageID: notification.ID.String(),
		UserID:    notification.RecipientID.String(),
		Title:     notification.Subject,
		Body:      notification.Body,
		Priority:  notification.Priority.String(),
	}

	response, err := uc.inAppProvider.SendInApp(ctx, inAppReq)
	uc.handleInAppDeliveryResult(ctx, notification, err, response)
}

func (uc *notificationUseCase) retryPushDelivery(ctx context.Context, notification *domain.Notification) {
	pushReq := ports.PushRequest{
		MessageID:   notification.ID.String(),
		DeviceToken: notification.DeviceToken,
		Title:       notification.Subject,
		Body:        notification.Body,
		Priority:    notification.Priority.String(),
	}

	response, err := uc.pushProvider.SendPush(ctx, pushReq)
	uc.handlePushDeliveryResult(ctx, notification, err, response)
}
