package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Notification Creation Tests
// ============================================================================

func TestNewNotification_Success(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name         string
		notifType    NotificationType
		channel      NotificationChannel
		body         string
	}{
		{"email notification", TypeTransactional, ChannelEmail, "Hello, World!"},
		{"SMS notification", TypeAlert, ChannelSMS, "Alert message"},
		{"push notification", TypeReminder, ChannelPush, "Reminder body"},
		{"in-app notification", TypeComment, ChannelInApp, "New comment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notif, err := NewNotification(tenantID, tt.notifType, tt.channel, tt.body)

			if err != nil {
				t.Fatalf("NewNotification() error = %v", err)
			}

			if notif == nil {
				t.Fatal("NewNotification() returned nil")
			}

			if notif.TenantID != tenantID {
				t.Errorf("TenantID = %v, want %v", notif.TenantID, tenantID)
			}
			if notif.Type != tt.notifType {
				t.Errorf("Type = %v, want %v", notif.Type, tt.notifType)
			}
			if notif.Channel != tt.channel {
				t.Errorf("Channel = %v, want %v", notif.Channel, tt.channel)
			}
			if notif.Body != tt.body {
				t.Errorf("Body = %s, want %s", notif.Body, tt.body)
			}
			if notif.Status != StatusPending {
				t.Errorf("Status = %v, want %v", notif.Status, StatusPending)
			}
			if notif.Priority != PriorityNormal {
				t.Errorf("Priority = %v, want %v", notif.Priority, PriorityNormal)
			}
			if notif.Code == "" {
				t.Error("Code should be generated")
			}
			if notif.RetryPolicy == nil {
				t.Error("RetryPolicy should be initialized")
			}
			if notif.Data == nil {
				t.Error("Data map should be initialized")
			}
			if notif.Metadata == nil {
				t.Error("Metadata map should be initialized")
			}
		})
	}
}

func TestNewNotification_InvalidType(t *testing.T) {
	tenantID := uuid.New()

	_, err := NewNotification(tenantID, NotificationType("invalid"), ChannelEmail, "body")

	if err == nil {
		t.Error("NewNotification() should return error for invalid type")
	}
}

func TestNewNotification_InvalidChannel(t *testing.T) {
	tenantID := uuid.New()

	_, err := NewNotification(tenantID, TypeTransactional, NotificationChannel("invalid"), "body")

	if err == nil {
		t.Error("NewNotification() should return error for invalid channel")
	}
}

func TestNewNotification_EmptyBody(t *testing.T) {
	tenantID := uuid.New()

	_, err := NewNotification(tenantID, TypeTransactional, ChannelEmail, "")

	if err == nil {
		t.Error("NewNotification() should return error for empty body")
	}
}

// ============================================================================
// Recipient Setting Tests
// ============================================================================

func createTestNotification(t *testing.T, channel NotificationChannel) *Notification {
	t.Helper()
	notif, err := NewNotification(uuid.New(), TypeTransactional, channel, "Test body")
	if err != nil {
		t.Fatalf("Failed to create test notification: %v", err)
	}
	return notif
}

func TestNotification_SetRecipientEmail(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetRecipientEmail("user@example.com", "John Doe")

	if err != nil {
		t.Fatalf("SetRecipientEmail() error = %v", err)
	}
	if notif.RecipientEmail != "user@example.com" {
		t.Errorf("RecipientEmail = %s, want 'user@example.com'", notif.RecipientEmail)
	}
	if notif.RecipientName != "John Doe" {
		t.Errorf("RecipientName = %s, want 'John Doe'", notif.RecipientName)
	}
}

func TestNotification_SetRecipientEmail_WrongChannel(t *testing.T) {
	notif := createTestNotification(t, ChannelSMS)

	err := notif.SetRecipientEmail("user@example.com", "John")

	if err != ErrInvalidChannel {
		t.Errorf("SetRecipientEmail() error = %v, want ErrInvalidChannel", err)
	}
}

func TestNotification_SetRecipientEmail_InvalidEmail(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetRecipientEmail("invalid-email", "John")

	if err == nil {
		t.Error("SetRecipientEmail() should return error for invalid email")
	}
}

func TestNotification_SetRecipientPhone(t *testing.T) {
	notif := createTestNotification(t, ChannelSMS)

	err := notif.SetRecipientPhone("+60123456789", "John Doe")

	if err != nil {
		t.Fatalf("SetRecipientPhone() error = %v", err)
	}
	if notif.RecipientPhone != "+60123456789" {
		t.Errorf("RecipientPhone = %s, want '+60123456789'", notif.RecipientPhone)
	}
}

func TestNotification_SetRecipientPhone_WhatsApp(t *testing.T) {
	notif := createTestNotification(t, ChannelWhatsApp)

	err := notif.SetRecipientPhone("+60123456789", "John")

	if err != nil {
		t.Fatalf("SetRecipientPhone() error = %v", err)
	}
}

func TestNotification_SetRecipientPhone_WrongChannel(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetRecipientPhone("+60123456789", "John")

	if err != ErrInvalidChannel {
		t.Errorf("SetRecipientPhone() error = %v, want ErrInvalidChannel", err)
	}
}

func TestNotification_SetRecipientUser(t *testing.T) {
	notif := createTestNotification(t, ChannelInApp)
	userID := uuid.New()

	err := notif.SetRecipientUser(userID, "John Doe")

	if err != nil {
		t.Fatalf("SetRecipientUser() error = %v", err)
	}
	if notif.RecipientID == nil || *notif.RecipientID != userID {
		t.Errorf("RecipientID = %v, want %v", notif.RecipientID, userID)
	}
}

func TestNotification_SetRecipientUser_WrongChannel(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetRecipientUser(uuid.New(), "John")

	if err != ErrInvalidChannel {
		t.Errorf("SetRecipientUser() error = %v, want ErrInvalidChannel", err)
	}
}

func TestNotification_SetRecipientDevice(t *testing.T) {
	notif := createTestNotification(t, ChannelPush)
	userID := uuid.New()

	err := notif.SetRecipientDevice(userID, "device-token-123", "John Doe")

	if err != nil {
		t.Fatalf("SetRecipientDevice() error = %v", err)
	}
	if notif.DeviceToken != "device-token-123" {
		t.Errorf("DeviceToken = %s, want 'device-token-123'", notif.DeviceToken)
	}
}

func TestNotification_SetRecipientDevice_EmptyToken(t *testing.T) {
	notif := createTestNotification(t, ChannelPush)

	err := notif.SetRecipientDevice(uuid.New(), "", "John")

	if err != ErrDeviceTokenRequired {
		t.Errorf("SetRecipientDevice() error = %v, want ErrDeviceTokenRequired", err)
	}
}

func TestNotification_SetRecipientDevice_WrongChannel(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetRecipientDevice(uuid.New(), "token", "John")

	if err != ErrInvalidChannel {
		t.Errorf("SetRecipientDevice() error = %v, want ErrInvalidChannel", err)
	}
}

// ============================================================================
// Content Methods Tests
// ============================================================================

func TestNotification_SetSubject(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetSubject("Test Subject")

	if err != nil {
		t.Fatalf("SetSubject() error = %v", err)
	}
	if notif.Subject != "Test Subject" {
		t.Errorf("Subject = %s, want 'Test Subject'", notif.Subject)
	}
}

func TestNotification_SetSubject_TooLong(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	longSubject := make([]byte, 300)
	for i := range longSubject {
		longSubject[i] = 'a'
	}

	err := notif.SetSubject(string(longSubject))

	if err != ErrEmailSubjectTooLong {
		t.Errorf("SetSubject() error = %v, want ErrEmailSubjectTooLong", err)
	}
}

func TestNotification_SetHTMLBody(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	notif.SetHTMLBody("<html><body>Hello</body></html>")

	if notif.HTMLBody != "<html><body>Hello</body></html>" {
		t.Errorf("HTMLBody = %s, want '<html><body>Hello</body></html>'", notif.HTMLBody)
	}
}

func TestNotification_SetData(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	data := map[string]interface{}{
		"name":  "John",
		"count": 5,
	}

	notif.SetData(data)

	if notif.Data["name"] != "John" {
		t.Errorf("Data[name] = %v, want 'John'", notif.Data["name"])
	}
	if notif.Data["count"] != 5 {
		t.Errorf("Data[count] = %v, want 5", notif.Data["count"])
	}
}

func TestNotification_AddData(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Data = nil // Reset

	notif.AddData("key", "value")

	if notif.Data == nil {
		t.Fatal("Data should be initialized")
	}
	if notif.Data["key"] != "value" {
		t.Errorf("Data[key] = %v, want 'value'", notif.Data["key"])
	}
}

func TestNotification_SetTemplate(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	templateID := uuid.New()

	notif.SetTemplate(templateID, "welcome_email")

	if notif.TemplateID == nil || *notif.TemplateID != templateID {
		t.Errorf("TemplateID = %v, want %v", notif.TemplateID, templateID)
	}
	if notif.TemplateName != "welcome_email" {
		t.Errorf("TemplateName = %s, want 'welcome_email'", notif.TemplateName)
	}
}

func TestNotification_AddAttachment(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	attachment := Attachment{
		Filename:    "document.pdf",
		ContentType: "application/pdf",
		Size:        1024,
	}

	err := notif.AddAttachment(attachment)

	if err != nil {
		t.Fatalf("AddAttachment() error = %v", err)
	}
	if len(notif.Attachments) != 1 {
		t.Errorf("Attachments count = %d, want 1", len(notif.Attachments))
	}
	if notif.Attachments[0].ID == uuid.Nil {
		t.Error("Attachment should be assigned an ID")
	}
}

func TestNotification_AddAttachment_WrongChannel(t *testing.T) {
	notif := createTestNotification(t, ChannelSMS)

	attachment := Attachment{Filename: "file.txt", Size: 100}

	err := notif.AddAttachment(attachment)

	if err == nil {
		t.Error("AddAttachment() should return error for non-email channel")
	}
}

func TestNotification_AddAttachment_TooLarge(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	attachment := Attachment{
		Filename: "large.bin",
		Size:     30 * 1024 * 1024, // 30 MB
	}

	err := notif.AddAttachment(attachment)

	if err != ErrAttachmentTooLarge {
		t.Errorf("AddAttachment() error = %v, want ErrAttachmentTooLarge", err)
	}
}

func TestNotification_AddAttachment_TooMany(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	for i := 0; i < 10; i++ {
		notif.AddAttachment(Attachment{Filename: "file.txt", Size: 100})
	}

	err := notif.AddAttachment(Attachment{Filename: "extra.txt", Size: 100})

	if err != ErrTooManyAttachments {
		t.Errorf("AddAttachment() error = %v, want ErrTooManyAttachments", err)
	}
}

// ============================================================================
// Email Configuration Tests
// ============================================================================

func TestNotification_SetFromAddress(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetFromAddress("sender@example.com", "Sender Name")

	if err != nil {
		t.Fatalf("SetFromAddress() error = %v", err)
	}
	if notif.FromAddress != "sender@example.com" {
		t.Errorf("FromAddress = %s, want 'sender@example.com'", notif.FromAddress)
	}
	if notif.FromName != "Sender Name" {
		t.Errorf("FromName = %s, want 'Sender Name'", notif.FromName)
	}
}

func TestNotification_SetFromAddress_WrongChannel(t *testing.T) {
	notif := createTestNotification(t, ChannelSMS)

	err := notif.SetFromAddress("sender@example.com", "Sender")

	if err != ErrInvalidChannel {
		t.Errorf("SetFromAddress() error = %v, want ErrInvalidChannel", err)
	}
}

func TestNotification_SetReplyTo(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetReplyTo("reply@example.com")

	if err != nil {
		t.Fatalf("SetReplyTo() error = %v", err)
	}
	if notif.ReplyTo != "reply@example.com" {
		t.Errorf("ReplyTo = %s, want 'reply@example.com'", notif.ReplyTo)
	}
}

func TestNotification_AddCC(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.AddCC("cc@example.com")

	if err != nil {
		t.Fatalf("AddCC() error = %v", err)
	}
	if len(notif.CC) != 1 || notif.CC[0] != "cc@example.com" {
		t.Errorf("CC = %v, want ['cc@example.com']", notif.CC)
	}
}

func TestNotification_AddCC_WrongChannel(t *testing.T) {
	notif := createTestNotification(t, ChannelSMS)

	err := notif.AddCC("cc@example.com")

	if err != ErrInvalidChannel {
		t.Errorf("AddCC() error = %v, want ErrInvalidChannel", err)
	}
}

func TestNotification_AddBCC(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.AddBCC("bcc@example.com")

	if err != nil {
		t.Fatalf("AddBCC() error = %v", err)
	}
	if len(notif.BCC) != 1 || notif.BCC[0] != "bcc@example.com" {
		t.Errorf("BCC = %v, want ['bcc@example.com']", notif.BCC)
	}
}

func TestNotification_SetHeader(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	notif.SetHeader("X-Custom-Header", "value")

	if notif.Headers == nil {
		t.Fatal("Headers should be initialized")
	}
	if notif.Headers["X-Custom-Header"] != "value" {
		t.Errorf("Headers[X-Custom-Header] = %s, want 'value'", notif.Headers["X-Custom-Header"])
	}
}

// ============================================================================
// Status Transition Tests
// ============================================================================

func TestNotification_Queue(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.Queue()

	if err != nil {
		t.Fatalf("Queue() error = %v", err)
	}
	if notif.Status != StatusQueued {
		t.Errorf("Status = %v, want %v", notif.Status, StatusQueued)
	}

	// Check domain event
	events := notif.GetDomainEvents()
	if len(events) == 0 {
		t.Error("Expected domain event to be added")
	}
}

func TestNotification_Queue_InvalidTransition(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSent // Not valid to queue from sent

	err := notif.Queue()

	if err == nil {
		t.Error("Queue() should return error for invalid transition")
	}
}

func TestNotification_Schedule(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	scheduleTime := time.Now().Add(time.Hour)

	err := notif.Schedule(scheduleTime)

	if err != nil {
		t.Fatalf("Schedule() error = %v", err)
	}
	if notif.Status != StatusScheduled {
		t.Errorf("Status = %v, want %v", notif.Status, StatusScheduled)
	}
	if notif.ScheduledAt == nil {
		t.Error("ScheduledAt should be set")
	}
}

func TestNotification_Schedule_PastTime(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	pastTime := time.Now().Add(-time.Hour)

	err := notif.Schedule(pastTime)

	if err != ErrScheduledTimeInPast {
		t.Errorf("Schedule() error = %v, want ErrScheduledTimeInPast", err)
	}
}

func TestNotification_MarkSending(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusQueued

	err := notif.MarkSending()

	if err != nil {
		t.Fatalf("MarkSending() error = %v", err)
	}
	if notif.Status != StatusSending {
		t.Errorf("Status = %v, want %v", notif.Status, StatusSending)
	}
	if notif.AttemptCount != 1 {
		t.Errorf("AttemptCount = %d, want 1", notif.AttemptCount)
	}
	if notif.LastAttemptAt == nil {
		t.Error("LastAttemptAt should be set")
	}
}

func TestNotification_MarkSent(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSending

	err := notif.MarkSent("provider-msg-123")

	if err != nil {
		t.Fatalf("MarkSent() error = %v", err)
	}
	if notif.Status != StatusSent {
		t.Errorf("Status = %v, want %v", notif.Status, StatusSent)
	}
	if notif.SentAt == nil {
		t.Error("SentAt should be set")
	}
	if notif.ProviderMessageID != "provider-msg-123" {
		t.Errorf("ProviderMessageID = %s, want 'provider-msg-123'", notif.ProviderMessageID)
	}
}

func TestNotification_MarkDelivered(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSent

	err := notif.MarkDelivered()

	if err != nil {
		t.Fatalf("MarkDelivered() error = %v", err)
	}
	if notif.Status != StatusDelivered {
		t.Errorf("Status = %v, want %v", notif.Status, StatusDelivered)
	}
	if notif.DeliveredAt == nil {
		t.Error("DeliveredAt should be set")
	}
}

func TestNotification_MarkRead(t *testing.T) {
	notif := createTestNotification(t, ChannelInApp)
	notif.Status = StatusDelivered

	err := notif.MarkRead()

	if err != nil {
		t.Fatalf("MarkRead() error = %v", err)
	}
	if notif.Status != StatusRead {
		t.Errorf("Status = %v, want %v", notif.Status, StatusRead)
	}
	if notif.ReadAt == nil {
		t.Error("ReadAt should be set")
	}
}

func TestNotification_MarkFailed(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSending

	err := notif.MarkFailed("SMTP_ERROR", "Connection refused", "provider details")

	if err != nil {
		t.Fatalf("MarkFailed() error = %v", err)
	}
	if notif.Status != StatusFailed {
		t.Errorf("Status = %v, want %v", notif.Status, StatusFailed)
	}
	if notif.ErrorCode != "SMTP_ERROR" {
		t.Errorf("ErrorCode = %s, want 'SMTP_ERROR'", notif.ErrorCode)
	}
	if notif.ErrorMessage != "Connection refused" {
		t.Errorf("ErrorMessage = %s, want 'Connection refused'", notif.ErrorMessage)
	}
	if notif.ProviderError != "provider details" {
		t.Errorf("ProviderError = %s, want 'provider details'", notif.ProviderError)
	}
	if notif.FailedAt == nil {
		t.Error("FailedAt should be set")
	}
}

func TestNotification_MarkRetrying(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSending
	nextRetry := time.Now().Add(time.Minute)

	err := notif.MarkRetrying(nextRetry)

	if err != nil {
		t.Fatalf("MarkRetrying() error = %v", err)
	}
	if notif.Status != StatusRetrying {
		t.Errorf("Status = %v, want %v", notif.Status, StatusRetrying)
	}
	if notif.NextRetryAt == nil {
		t.Error("NextRetryAt should be set")
	}
}

func TestNotification_MarkRetrying_MaxRetriesExceeded(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSending
	notif.AttemptCount = 10 // Exceeds max
	notif.RetryPolicy = &RetryPolicy{MaxAttempts: 3}

	err := notif.MarkRetrying(time.Now().Add(time.Minute))

	if err != ErrMaxRetriesExceeded {
		t.Errorf("MarkRetrying() error = %v, want ErrMaxRetriesExceeded", err)
	}
}

func TestNotification_Cancel(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusScheduled

	err := notif.Cancel()

	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if notif.Status != StatusCancelled {
		t.Errorf("Status = %v, want %v", notif.Status, StatusCancelled)
	}
	if notif.CancelledAt == nil {
		t.Error("CancelledAt should be set")
	}
}

func TestNotification_Cancel_FinalStatus(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusDelivered

	err := notif.Cancel()

	if err != ErrCannotCancel {
		t.Errorf("Cancel() error = %v, want ErrCannotCancel", err)
	}
}

func TestNotification_Cancel_Sending(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSending

	err := notif.Cancel()

	if err != ErrCannotCancel {
		t.Errorf("Cancel() error = %v, want ErrCannotCancel", err)
	}
}

func TestNotification_MarkBounced(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSent

	err := notif.MarkBounced("hard", "Mailbox not found")

	if err != nil {
		t.Fatalf("MarkBounced() error = %v", err)
	}
	if notif.Status != StatusBounced {
		t.Errorf("Status = %v, want %v", notif.Status, StatusBounced)
	}
	if notif.ErrorCode != "BOUNCED" {
		t.Errorf("ErrorCode = %s, want 'BOUNCED'", notif.ErrorCode)
	}
}

func TestNotification_MarkBounced_WrongChannel(t *testing.T) {
	notif := createTestNotification(t, ChannelSMS)
	notif.Status = StatusSent

	err := notif.MarkBounced("hard", "message")

	if err != ErrInvalidChannel {
		t.Errorf("MarkBounced() error = %v, want ErrInvalidChannel", err)
	}
}

func TestNotification_MarkComplained(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusSent

	err := notif.MarkComplained()

	if err != nil {
		t.Fatalf("MarkComplained() error = %v", err)
	}
	if notif.Status != StatusComplained {
		t.Errorf("Status = %v, want %v", notif.Status, StatusComplained)
	}
	if notif.ErrorCode != "COMPLAINED" {
		t.Errorf("ErrorCode = %s, want 'COMPLAINED'", notif.ErrorCode)
	}
}

// ============================================================================
// Tracking Methods Tests
// ============================================================================

func TestNotification_RecordOpen(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.OpenCount = 0

	notif.RecordOpen()
	notif.RecordOpen()

	if notif.OpenCount != 2 {
		t.Errorf("OpenCount = %d, want 2", notif.OpenCount)
	}
}

func TestNotification_RecordClick(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.ClickCount = 0

	notif.RecordClick()
	notif.RecordClick()
	notif.RecordClick()

	if notif.ClickCount != 3 {
		t.Errorf("ClickCount = %d, want 3", notif.ClickCount)
	}
}

func TestNotification_EnableTracking(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	notif.EnableTracking(true, true)

	if !notif.TrackOpens {
		t.Error("TrackOpens should be true")
	}
	if !notif.TrackClicks {
		t.Error("TrackClicks should be true")
	}
}

// ============================================================================
// Source Reference Tests
// ============================================================================

func TestNotification_SetSourceEvent(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	entityID := uuid.New()

	notif.SetSourceEvent("order.created", &entityID, "Order")

	if notif.SourceEvent != "order.created" {
		t.Errorf("SourceEvent = %s, want 'order.created'", notif.SourceEvent)
	}
	if notif.SourceEntityID == nil || *notif.SourceEntityID != entityID {
		t.Errorf("SourceEntityID = %v, want %v", notif.SourceEntityID, entityID)
	}
	if notif.SourceEntityType != "Order" {
		t.Errorf("SourceEntityType = %s, want 'Order'", notif.SourceEntityType)
	}
}

func TestNotification_SetCorrelationID(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	notif.SetCorrelationID("correlation-123")

	if notif.CorrelationID != "correlation-123" {
		t.Errorf("CorrelationID = %s, want 'correlation-123'", notif.CorrelationID)
	}
}

func TestNotification_SetBatch(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	batchID := uuid.New()

	notif.SetBatch(batchID, 5)

	if notif.BatchID == nil || *notif.BatchID != batchID {
		t.Errorf("BatchID = %v, want %v", notif.BatchID, batchID)
	}
	if notif.BatchIndex != 5 {
		t.Errorf("BatchIndex = %d, want 5", notif.BatchIndex)
	}
}

// ============================================================================
// Priority Methods Tests
// ============================================================================

func TestNotification_SetPriority(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetPriority(PriorityCritical)

	if err != nil {
		t.Fatalf("SetPriority() error = %v", err)
	}
	if notif.Priority != PriorityCritical {
		t.Errorf("Priority = %v, want %v", notif.Priority, PriorityCritical)
	}
}

func TestNotification_SetPriority_Invalid(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)

	err := notif.SetPriority(NotificationPriority("invalid"))

	if err != ErrInvalidPriority {
		t.Errorf("SetPriority() error = %v, want ErrInvalidPriority", err)
	}
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestNotification_Validate_Email(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.SetRecipientEmail("user@example.com", "User")
	notif.SetSubject("Test Subject")

	err := notif.Validate()

	if err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestNotification_Validate_Email_MissingRecipient(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.SetSubject("Test Subject")

	err := notif.Validate()

	if err == nil {
		t.Error("Validate() should return error for missing recipient")
	}
}

func TestNotification_Validate_Email_MissingSubject(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.SetRecipientEmail("user@example.com", "User")

	err := notif.Validate()

	if err == nil {
		t.Error("Validate() should return error for missing subject")
	}
}

func TestNotification_Validate_SMS_MissingPhone(t *testing.T) {
	notif := createTestNotification(t, ChannelSMS)

	err := notif.Validate()

	if err == nil {
		t.Error("Validate() should return error for missing phone")
	}
}

func TestNotification_Validate_Push_MissingToken(t *testing.T) {
	notif := createTestNotification(t, ChannelPush)

	err := notif.Validate()

	if err == nil {
		t.Error("Validate() should return error for missing device token")
	}
}

func TestNotification_Validate_InApp_MissingUserID(t *testing.T) {
	notif := createTestNotification(t, ChannelInApp)

	err := notif.Validate()

	if err == nil {
		t.Error("Validate() should return error for missing user ID")
	}
}

// ============================================================================
// Retry Logic Tests
// ============================================================================

func TestNotification_CanRetry(t *testing.T) {
	tests := []struct {
		name         string
		status       NotificationStatus
		attemptCount int
		maxAttempts  int
		expected     bool
	}{
		{"failed with attempts remaining", StatusFailed, 1, 3, true},
		{"failed at max attempts", StatusFailed, 3, 3, false},
		{"retrying with attempts remaining", StatusRetrying, 2, 5, true},
		{"pending cannot retry", StatusPending, 0, 3, false},
		{"delivered cannot retry", StatusDelivered, 1, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notif := createTestNotification(t, ChannelEmail)
			notif.Status = tt.status
			notif.AttemptCount = tt.attemptCount
			notif.RetryPolicy = &RetryPolicy{MaxAttempts: tt.maxAttempts}

			if notif.CanRetry() != tt.expected {
				t.Errorf("CanRetry() = %v, want %v", notif.CanRetry(), tt.expected)
			}
		})
	}
}

func TestNotification_CanRetry_NoPolicy(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusFailed
	notif.RetryPolicy = nil

	if notif.CanRetry() {
		t.Error("CanRetry() should return false when no retry policy")
	}
}

func TestNotification_GetNextRetryInterval(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.AttemptCount = 2
	notif.RetryPolicy = &RetryPolicy{
		MaxAttempts:     5,
		InitialInterval: 60,
		MaxInterval:     3600,
		Multiplier:      2.0,
	}

	interval := notif.GetNextRetryInterval()

	// For attempt 2, should be: 60 * 2 = 120
	if interval != 120 {
		t.Errorf("GetNextRetryInterval() = %d, want 120", interval)
	}
}

func TestNotification_GetNextRetryInterval_NoPolicy(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.RetryPolicy = nil

	interval := notif.GetNextRetryInterval()

	if interval != 0 {
		t.Errorf("GetNextRetryInterval() = %d, want 0", interval)
	}
}

// ============================================================================
// Query Methods Tests
// ============================================================================

func TestNotification_IsPending(t *testing.T) {
	tests := []struct {
		status   NotificationStatus
		expected bool
	}{
		{StatusPending, true},
		{StatusScheduled, true},
		{StatusQueued, false},
		{StatusSent, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			notif := createTestNotification(t, ChannelEmail)
			notif.Status = tt.status

			if notif.IsPending() != tt.expected {
				t.Errorf("IsPending() = %v, want %v", notif.IsPending(), tt.expected)
			}
		})
	}
}

func TestNotification_IsDelivered(t *testing.T) {
	tests := []struct {
		status   NotificationStatus
		expected bool
	}{
		{StatusDelivered, true},
		{StatusRead, true},
		{StatusSent, false},
		{StatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			notif := createTestNotification(t, ChannelEmail)
			notif.Status = tt.status

			if notif.IsDelivered() != tt.expected {
				t.Errorf("IsDelivered() = %v, want %v", notif.IsDelivered(), tt.expected)
			}
		})
	}
}

func TestNotification_IsFailed(t *testing.T) {
	tests := []struct {
		status   NotificationStatus
		expected bool
	}{
		{StatusFailed, true},
		{StatusBounced, true},
		{StatusComplained, true},
		{StatusDelivered, false},
		{StatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			notif := createTestNotification(t, ChannelEmail)
			notif.Status = tt.status

			if notif.IsFailed() != tt.expected {
				t.Errorf("IsFailed() = %v, want %v", notif.IsFailed(), tt.expected)
			}
		})
	}
}

func TestNotification_IsCancelled(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusCancelled

	if !notif.IsCancelled() {
		t.Error("IsCancelled() should return true")
	}

	notif.Status = StatusSent
	if notif.IsCancelled() {
		t.Error("IsCancelled() should return false")
	}
}

func TestNotification_IsScheduled(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	scheduledTime := time.Now().Add(time.Hour)
	notif.Status = StatusScheduled
	notif.ScheduledAt = &scheduledTime

	if !notif.IsScheduled() {
		t.Error("IsScheduled() should return true")
	}

	notif.ScheduledAt = nil
	if notif.IsScheduled() {
		t.Error("IsScheduled() should return false when ScheduledAt is nil")
	}
}

func TestNotification_IsDue(t *testing.T) {
	notif := createTestNotification(t, ChannelEmail)
	notif.Status = StatusScheduled

	// Past time - should be due
	pastTime := time.Now().Add(-time.Hour)
	notif.ScheduledAt = &pastTime
	if !notif.IsDue() {
		t.Error("IsDue() should return true for past scheduled time")
	}

	// Future time - should not be due
	futureTime := time.Now().Add(time.Hour)
	notif.ScheduledAt = &futureTime
	if notif.IsDue() {
		t.Error("IsDue() should return false for future scheduled time")
	}

	// Not scheduled status
	notif.Status = StatusPending
	if notif.IsDue() {
		t.Error("IsDue() should return false when not scheduled")
	}
}
