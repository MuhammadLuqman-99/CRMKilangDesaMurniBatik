package domain

import (
	"encoding/json"
	"testing"
)

// ============================================================================
// NotificationChannel Tests
// ============================================================================

func TestNotificationChannel_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		channel  NotificationChannel
		expected bool
	}{
		{"email is valid", ChannelEmail, true},
		{"sms is valid", ChannelSMS, true},
		{"push is valid", ChannelPush, true},
		{"in_app is valid", ChannelInApp, true},
		{"webhook is valid", ChannelWebhook, true},
		{"slack is valid", ChannelSlack, true},
		{"whatsapp is valid", ChannelWhatsApp, true},
		{"telegram is valid", ChannelTelegram, true},
		{"invalid channel", NotificationChannel("invalid"), false},
		{"empty channel", NotificationChannel(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.channel.IsValid()
			if result != tt.expected {
				t.Errorf("NotificationChannel(%q).IsValid() = %v, want %v", tt.channel, result, tt.expected)
			}
		})
	}
}

func TestNotificationChannel_String(t *testing.T) {
	channel := ChannelEmail
	if channel.String() != "email" {
		t.Errorf("NotificationChannel.String() = %s, want 'email'", channel.String())
	}
}

func TestNotificationChannel_DisplayName(t *testing.T) {
	tests := []struct {
		channel  NotificationChannel
		expected string
	}{
		{ChannelEmail, "Email"},
		{ChannelSMS, "SMS"},
		{ChannelPush, "Push Notification"},
		{ChannelInApp, "In-App Notification"},
		{ChannelWebhook, "Webhook"},
		{ChannelSlack, "Slack"},
		{ChannelWhatsApp, "WhatsApp"},
		{ChannelTelegram, "Telegram"},
		{NotificationChannel("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			result := tt.channel.DisplayName()
			if result != tt.expected {
				t.Errorf("DisplayName() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestNotificationChannel_RequiresDeviceToken(t *testing.T) {
	tests := []struct {
		channel  NotificationChannel
		expected bool
	}{
		{ChannelPush, true},
		{ChannelEmail, false},
		{ChannelSMS, false},
		{ChannelInApp, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			if tt.channel.RequiresDeviceToken() != tt.expected {
				t.Errorf("RequiresDeviceToken() = %v, want %v", tt.channel.RequiresDeviceToken(), tt.expected)
			}
		})
	}
}

func TestNotificationChannel_RequiresPhoneNumber(t *testing.T) {
	tests := []struct {
		channel  NotificationChannel
		expected bool
	}{
		{ChannelSMS, true},
		{ChannelWhatsApp, true},
		{ChannelEmail, false},
		{ChannelPush, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			if tt.channel.RequiresPhoneNumber() != tt.expected {
				t.Errorf("RequiresPhoneNumber() = %v, want %v", tt.channel.RequiresPhoneNumber(), tt.expected)
			}
		})
	}
}

func TestNotificationChannel_RequiresEmail(t *testing.T) {
	tests := []struct {
		channel  NotificationChannel
		expected bool
	}{
		{ChannelEmail, true},
		{ChannelSMS, false},
		{ChannelPush, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			if tt.channel.RequiresEmail() != tt.expected {
				t.Errorf("RequiresEmail() = %v, want %v", tt.channel.RequiresEmail(), tt.expected)
			}
		})
	}
}

func TestNotificationChannel_RequiresUserID(t *testing.T) {
	tests := []struct {
		channel  NotificationChannel
		expected bool
	}{
		{ChannelInApp, true},
		{ChannelEmail, false},
		{ChannelSMS, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			if tt.channel.RequiresUserID() != tt.expected {
				t.Errorf("RequiresUserID() = %v, want %v", tt.channel.RequiresUserID(), tt.expected)
			}
		})
	}
}

func TestParseChannel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  NotificationChannel
		expectErr bool
	}{
		{"valid email", "email", ChannelEmail, false},
		{"valid SMS uppercase", "SMS", ChannelSMS, false},
		{"valid with spaces", "  push  ", ChannelPush, false},
		{"invalid channel", "invalid", "", true},
		{"empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseChannel(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseChannel(%q) error = %v, expectErr %v", tt.input, err, tt.expectErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseChannel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAllChannels(t *testing.T) {
	channels := AllChannels()
	if len(channels) != 8 {
		t.Errorf("AllChannels() returned %d channels, want 8", len(channels))
	}
}

// ============================================================================
// NotificationPriority Tests
// ============================================================================

func TestNotificationPriority_IsValid(t *testing.T) {
	tests := []struct {
		priority NotificationPriority
		expected bool
	}{
		{PriorityLow, true},
		{PriorityNormal, true},
		{PriorityHigh, true},
		{PriorityCritical, true},
		{NotificationPriority("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if tt.priority.IsValid() != tt.expected {
				t.Errorf("IsValid() = %v, want %v", tt.priority.IsValid(), tt.expected)
			}
		})
	}
}

func TestNotificationPriority_Weight(t *testing.T) {
	tests := []struct {
		priority NotificationPriority
		expected int
	}{
		{PriorityLow, 1},
		{PriorityNormal, 2},
		{PriorityHigh, 3},
		{PriorityCritical, 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if tt.priority.Weight() != tt.expected {
				t.Errorf("Weight() = %d, want %d", tt.priority.Weight(), tt.expected)
			}
		})
	}
}

func TestNotificationPriority_IsHigherThan(t *testing.T) {
	tests := []struct {
		name     string
		p1       NotificationPriority
		p2       NotificationPriority
		expected bool
	}{
		{"critical higher than high", PriorityCritical, PriorityHigh, true},
		{"high higher than normal", PriorityHigh, PriorityNormal, true},
		{"normal higher than low", PriorityNormal, PriorityLow, true},
		{"low not higher than normal", PriorityLow, PriorityNormal, false},
		{"same priority", PriorityNormal, PriorityNormal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.p1.IsHigherThan(tt.p2) != tt.expected {
				t.Errorf("IsHigherThan() = %v, want %v", tt.p1.IsHigherThan(tt.p2), tt.expected)
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  NotificationPriority
		expectErr bool
	}{
		{"valid low", "low", PriorityLow, false},
		{"valid HIGH uppercase", "HIGH", PriorityHigh, false},
		{"invalid", "urgent", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePriority(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParsePriority(%q) error = %v, expectErr %v", tt.input, err, tt.expectErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParsePriority(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// NotificationStatus Tests
// ============================================================================

func TestNotificationStatus_IsValid(t *testing.T) {
	validStatuses := []NotificationStatus{
		StatusPending, StatusQueued, StatusSending, StatusSent,
		StatusDelivered, StatusRead, StatusFailed, StatusCancelled,
		StatusScheduled, StatusRetrying, StatusBounced, StatusComplained,
	}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			if !status.IsValid() {
				t.Errorf("Status %q should be valid", status)
			}
		})
	}

	if NotificationStatus("invalid").IsValid() {
		t.Error("Invalid status should return false")
	}
}

func TestNotificationStatus_IsFinal(t *testing.T) {
	tests := []struct {
		status   NotificationStatus
		expected bool
	}{
		{StatusPending, false},
		{StatusQueued, false},
		{StatusSending, false},
		{StatusSent, false},
		{StatusDelivered, true},
		{StatusRead, true},
		{StatusFailed, true},
		{StatusCancelled, true},
		{StatusScheduled, false},
		{StatusRetrying, false},
		{StatusBounced, true},
		{StatusComplained, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsFinal() != tt.expected {
				t.Errorf("IsFinal() = %v, want %v", tt.status.IsFinal(), tt.expected)
			}
		})
	}
}

func TestNotificationStatus_IsSuccess(t *testing.T) {
	tests := []struct {
		status   NotificationStatus
		expected bool
	}{
		{StatusSent, true},
		{StatusDelivered, true},
		{StatusRead, true},
		{StatusPending, false},
		{StatusFailed, false},
		{StatusBounced, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsSuccess() != tt.expected {
				t.Errorf("IsSuccess() = %v, want %v", tt.status.IsSuccess(), tt.expected)
			}
		})
	}
}

func TestNotificationStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     NotificationStatus
		to       NotificationStatus
		expected bool
	}{
		{"pending to queued", StatusPending, StatusQueued, true},
		{"pending to scheduled", StatusPending, StatusScheduled, true},
		{"pending to cancelled", StatusPending, StatusCancelled, true},
		{"pending to sent", StatusPending, StatusSent, false},
		{"scheduled to queued", StatusScheduled, StatusQueued, true},
		{"scheduled to cancelled", StatusScheduled, StatusCancelled, true},
		{"queued to sending", StatusQueued, StatusSending, true},
		{"sending to sent", StatusSending, StatusSent, true},
		{"sending to failed", StatusSending, StatusFailed, true},
		{"sending to retrying", StatusSending, StatusRetrying, true},
		{"retrying to sending", StatusRetrying, StatusSending, true},
		{"sent to delivered", StatusSent, StatusDelivered, true},
		{"sent to bounced", StatusSent, StatusBounced, true},
		{"delivered to read", StatusDelivered, StatusRead, true},
		{"failed to sent", StatusFailed, StatusSent, false},
		{"cancelled to anything", StatusCancelled, StatusSent, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.from.CanTransitionTo(tt.to) != tt.expected {
				t.Errorf("CanTransitionTo() = %v, want %v", tt.from.CanTransitionTo(tt.to), tt.expected)
			}
		})
	}
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		input     string
		expected  NotificationStatus
		expectErr bool
	}{
		{"pending", StatusPending, false},
		{"DELIVERED", StatusDelivered, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseStatus(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseStatus() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// NotificationType Tests
// ============================================================================

func TestNotificationType_IsValid(t *testing.T) {
	validTypes := []NotificationType{
		TypeSystem, TypeMarketing, TypeTransactional, TypeAlert,
		TypeReminder, TypeWelcome, TypeVerification, TypePasswordReset,
		TypeInvoice, TypePromotion, TypeUpdate, TypeComment,
		TypeMention, TypeAssignment, TypeDeadline,
	}

	for _, nt := range validTypes {
		t.Run(string(nt), func(t *testing.T) {
			if !nt.IsValid() {
				t.Errorf("Type %q should be valid", nt)
			}
		})
	}

	if NotificationType("invalid").IsValid() {
		t.Error("Invalid type should return false")
	}
}

func TestNotificationType_RequiresOptIn(t *testing.T) {
	tests := []struct {
		notifType NotificationType
		expected  bool
	}{
		{TypeMarketing, true},
		{TypePromotion, true},
		{TypeSystem, false},
		{TypeTransactional, false},
		{TypeVerification, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.notifType), func(t *testing.T) {
			if tt.notifType.RequiresOptIn() != tt.expected {
				t.Errorf("RequiresOptIn() = %v, want %v", tt.notifType.RequiresOptIn(), tt.expected)
			}
		})
	}
}

func TestNotificationType_CanUnsubscribe(t *testing.T) {
	tests := []struct {
		notifType NotificationType
		expected  bool
	}{
		{TypeMarketing, true},
		{TypePromotion, true},
		{TypeReminder, true},
		{TypeSystem, false},
		{TypeVerification, false},
		{TypePasswordReset, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.notifType), func(t *testing.T) {
			if tt.notifType.CanUnsubscribe() != tt.expected {
				t.Errorf("CanUnsubscribe() = %v, want %v", tt.notifType.CanUnsubscribe(), tt.expected)
			}
		})
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		input     string
		expected  NotificationType
		expectErr bool
	}{
		{"system", TypeSystem, false},
		{"MARKETING", TypeMarketing, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseType(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseType() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// EmailAddress Tests
// ============================================================================

func TestNewEmailAddress_Success(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"simple email", "user@example.com", "user@example.com"},
		{"email with subdomain", "user@mail.example.com", "user@mail.example.com"},
		{"email with plus", "user+tag@example.com", "user+tag@example.com"},
		{"email with dots", "first.last@example.com", "first.last@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmailAddress(tt.email)
			if err != nil {
				t.Fatalf("NewEmailAddress() error = %v", err)
			}
			if email.String() != tt.expected {
				t.Errorf("String() = %s, want %s", email.String(), tt.expected)
			}
		})
	}
}

func TestNewEmailAddress_Errors(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"empty string", ""},
		{"no @ symbol", "userexample.com"},
		{"multiple @", "user@@example.com"},
		{"no domain", "user@"},
		{"no local", "@example.com"},
		{"spaces", "user @example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEmailAddress(tt.email)
			if err == nil {
				t.Errorf("NewEmailAddress(%q) should return error", tt.email)
			}
		})
	}
}

func TestEmailAddress_Normalized(t *testing.T) {
	email, _ := NewEmailAddress("User@Example.COM")
	if email.Normalized() != "user@example.com" {
		t.Errorf("Normalized() = %s, want 'user@example.com'", email.Normalized())
	}
}

func TestEmailAddress_LocalAndDomain(t *testing.T) {
	email, _ := NewEmailAddress("user@example.com")

	if email.Local() != "user" {
		t.Errorf("Local() = %s, want 'user'", email.Local())
	}
	if email.Domain() != "example.com" {
		t.Errorf("Domain() = %s, want 'example.com'", email.Domain())
	}
}

func TestEmailAddress_IsEmpty(t *testing.T) {
	email, _ := NewEmailAddress("user@example.com")
	if email.IsEmpty() {
		t.Error("IsEmpty() should return false for valid email")
	}

	emptyEmail := EmailAddress{}
	if !emptyEmail.IsEmpty() {
		t.Error("IsEmpty() should return true for empty email")
	}
}

func TestEmailAddress_Equals(t *testing.T) {
	email1, _ := NewEmailAddress("user@example.com")
	email2, _ := NewEmailAddress("USER@EXAMPLE.COM")
	email3, _ := NewEmailAddress("other@example.com")

	if !email1.Equals(email2) {
		t.Error("Equals() should return true for case-insensitive match")
	}
	if email1.Equals(email3) {
		t.Error("Equals() should return false for different emails")
	}
}

func TestEmailAddress_JSON(t *testing.T) {
	email, _ := NewEmailAddress("user@example.com")

	// Marshal
	data, err := json.Marshal(email)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(data) != `"user@example.com"` {
		t.Errorf("JSON = %s, want '\"user@example.com\"'", string(data))
	}

	// Unmarshal
	var unmarshaled EmailAddress
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !email.Equals(unmarshaled) {
		t.Errorf("Unmarshaled email doesn't match original")
	}
}

// ============================================================================
// PhoneNumber Tests
// ============================================================================

func TestNewPhoneNumber_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"international format", "+60123456789", "+60123456789"},
		{"without plus", "60123456789", "+60123456789"},
		{"with dashes", "+60-12-345-6789", "+60123456789"},
		{"with spaces", "+60 12 345 6789", "+60123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone, err := NewPhoneNumber(tt.input)
			if err != nil {
				t.Fatalf("NewPhoneNumber() error = %v", err)
			}
			if phone.E164() != tt.expected {
				t.Errorf("E164() = %s, want %s", phone.E164(), tt.expected)
			}
		})
	}
}

func TestNewPhoneNumber_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only letters", "abcdefghij"},
		{"starts with 0", "+0123456789"},
		{"all zeros", "0000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPhoneNumber(tt.input)
			if err == nil {
				t.Errorf("NewPhoneNumber(%q) should return error", tt.input)
			}
		})
	}
}

func TestPhoneNumber_Raw(t *testing.T) {
	phone, _ := NewPhoneNumber("+60-12-345-6789")
	if phone.Raw() != "+60-12-345-6789" {
		t.Errorf("Raw() = %s, want '+60-12-345-6789'", phone.Raw())
	}
}

func TestPhoneNumber_IsEmpty(t *testing.T) {
	phone, _ := NewPhoneNumber("+60123456789")
	if phone.IsEmpty() {
		t.Error("IsEmpty() should return false for valid phone")
	}

	emptyPhone := PhoneNumber{}
	if !emptyPhone.IsEmpty() {
		t.Error("IsEmpty() should return true for empty phone")
	}
}

func TestPhoneNumber_Equals(t *testing.T) {
	phone1, _ := NewPhoneNumber("+60123456789")
	phone2, _ := NewPhoneNumber("60-12-345-6789")
	phone3, _ := NewPhoneNumber("+60987654321")

	if !phone1.Equals(phone2) {
		t.Error("Equals() should return true for same number different format")
	}
	if phone1.Equals(phone3) {
		t.Error("Equals() should return false for different numbers")
	}
}

func TestPhoneNumber_JSON(t *testing.T) {
	phone, _ := NewPhoneNumber("+60123456789")

	data, err := json.Marshal(phone)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(data) != `"+60123456789"` {
		t.Errorf("JSON = %s, want '\"+60123456789\"'", string(data))
	}

	var unmarshaled PhoneNumber
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !phone.Equals(unmarshaled) {
		t.Error("Unmarshaled phone doesn't match original")
	}
}

// ============================================================================
// DeviceToken Tests
// ============================================================================

func TestNewDeviceToken_Success(t *testing.T) {
	// iOS token (64 characters)
	iosToken := "1234567890123456789012345678901234567890123456789012345678901234"
	token, err := NewDeviceToken(iosToken, PlatformIOS)
	if err != nil {
		t.Fatalf("NewDeviceToken() error = %v", err)
	}
	if token.Token() != iosToken {
		t.Errorf("Token() = %s, want %s", token.Token(), iosToken)
	}
	if token.Platform() != PlatformIOS {
		t.Errorf("Platform() = %s, want %s", token.Platform(), PlatformIOS)
	}
}

func TestNewDeviceToken_Errors(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		platform DevicePlatform
	}{
		{"empty token", "", PlatformAndroid},
		{"invalid platform", "validtoken", DevicePlatform("invalid")},
		{"wrong iOS length", "short", PlatformIOS},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDeviceToken(tt.token, tt.platform)
			if err == nil {
				t.Error("NewDeviceToken() should return error")
			}
		})
	}
}

func TestDeviceToken_IsEmpty(t *testing.T) {
	token, _ := NewDeviceToken("androidtoken123", PlatformAndroid)
	if token.IsEmpty() {
		t.Error("IsEmpty() should return false for valid token")
	}

	emptyToken := DeviceToken{}
	if !emptyToken.IsEmpty() {
		t.Error("IsEmpty() should return true for empty token")
	}
}

func TestDeviceToken_Equals(t *testing.T) {
	token1, _ := NewDeviceToken("token123", PlatformAndroid)
	token2, _ := NewDeviceToken("token123", PlatformAndroid)
	token3, _ := NewDeviceToken("token123", PlatformWeb)
	token4, _ := NewDeviceToken("different", PlatformAndroid)

	if !token1.Equals(token2) {
		t.Error("Equals() should return true for same token and platform")
	}
	if token1.Equals(token3) {
		t.Error("Equals() should return false for different platform")
	}
	if token1.Equals(token4) {
		t.Error("Equals() should return false for different token")
	}
}

// ============================================================================
// Recipient Tests
// ============================================================================

func TestNewRecipient(t *testing.T) {
	recipient := NewRecipient()

	if recipient == nil {
		t.Fatal("NewRecipient() returned nil")
	}
	if recipient.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestRecipient_FluentAPI(t *testing.T) {
	recipient := NewRecipient().
		WithUserID("user123").
		WithEmail("user@example.com").
		WithPhone("+60123456789").
		WithDeviceToken("token", "android").
		WithName("John Doe").
		WithLocale("en-US").
		WithTimezone("Asia/Kuala_Lumpur").
		WithMetadata("key", "value")

	if recipient.UserID != "user123" {
		t.Errorf("UserID = %s, want 'user123'", recipient.UserID)
	}
	if recipient.Email != "user@example.com" {
		t.Errorf("Email = %s, want 'user@example.com'", recipient.Email)
	}
	if recipient.Phone != "+60123456789" {
		t.Errorf("Phone = %s, want '+60123456789'", recipient.Phone)
	}
	if recipient.DeviceToken != "token" {
		t.Errorf("DeviceToken = %s, want 'token'", recipient.DeviceToken)
	}
	if recipient.Platform != "android" {
		t.Errorf("Platform = %s, want 'android'", recipient.Platform)
	}
	if recipient.Name != "John Doe" {
		t.Errorf("Name = %s, want 'John Doe'", recipient.Name)
	}
	if recipient.Locale != "en-US" {
		t.Errorf("Locale = %s, want 'en-US'", recipient.Locale)
	}
	if recipient.Timezone != "Asia/Kuala_Lumpur" {
		t.Errorf("Timezone = %s, want 'Asia/Kuala_Lumpur'", recipient.Timezone)
	}
	if recipient.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %v, want 'value'", recipient.Metadata["key"])
	}
}

func TestRecipient_ValidateForChannel(t *testing.T) {
	tests := []struct {
		name      string
		recipient *Recipient
		channel   NotificationChannel
		expectErr bool
	}{
		{
			name:      "valid email recipient",
			recipient: NewRecipient().WithEmail("user@example.com"),
			channel:   ChannelEmail,
			expectErr: false,
		},
		{
			name:      "missing email for email channel",
			recipient: NewRecipient(),
			channel:   ChannelEmail,
			expectErr: true,
		},
		{
			name:      "valid phone for SMS",
			recipient: NewRecipient().WithPhone("+60123456789"),
			channel:   ChannelSMS,
			expectErr: false,
		},
		{
			name:      "missing phone for SMS",
			recipient: NewRecipient(),
			channel:   ChannelSMS,
			expectErr: true,
		},
		{
			name:      "valid device token for push",
			recipient: NewRecipient().WithDeviceToken("token123", "android"),
			channel:   ChannelPush,
			expectErr: false,
		},
		{
			name:      "missing device token for push",
			recipient: NewRecipient(),
			channel:   ChannelPush,
			expectErr: true,
		},
		{
			name:      "valid user ID for in-app",
			recipient: NewRecipient().WithUserID("user123"),
			channel:   ChannelInApp,
			expectErr: false,
		},
		{
			name:      "missing user ID for in-app",
			recipient: NewRecipient(),
			channel:   ChannelInApp,
			expectErr: true,
		},
		{
			name:      "valid webhook URL",
			recipient: NewRecipient().WithWebhookURL("https://example.com/webhook"),
			channel:   ChannelWebhook,
			expectErr: false,
		},
		{
			name:      "missing webhook URL",
			recipient: NewRecipient(),
			channel:   ChannelWebhook,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.recipient.ValidateForChannel(tt.channel)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateForChannel() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestRecipient_IsEmpty(t *testing.T) {
	emptyRecipient := NewRecipient()
	if !emptyRecipient.IsEmpty() {
		t.Error("IsEmpty() should return true for empty recipient")
	}

	withEmail := NewRecipient().WithEmail("user@example.com")
	if withEmail.IsEmpty() {
		t.Error("IsEmpty() should return false for recipient with email")
	}
}

// ============================================================================
// TemplateVariable Tests
// ============================================================================

func TestNewTemplateVariable_Success(t *testing.T) {
	variable, err := NewTemplateVariable("username", "string", true)

	if err != nil {
		t.Fatalf("NewTemplateVariable() error = %v", err)
	}
	if variable.Name != "username" {
		t.Errorf("Name = %s, want 'username'", variable.Name)
	}
	if variable.Type != "string" {
		t.Errorf("Type = %s, want 'string'", variable.Type)
	}
	if !variable.Required {
		t.Error("Required should be true")
	}
}

func TestNewTemplateVariable_Errors(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		varType  string
		required bool
	}{
		{"empty name", "", "string", true},
		{"invalid type", "varname", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTemplateVariable(tt.varName, tt.varType, tt.required)
			if err == nil {
				t.Error("NewTemplateVariable() should return error")
			}
		})
	}
}

func TestTemplateVariable_FluentMethods(t *testing.T) {
	variable, _ := NewTemplateVariable("amount", "number", true)
	variable.
		WithDefaultValue(0).
		WithDescription("The payment amount").
		WithExample(100.50)

	if variable.DefaultValue != 0 {
		t.Errorf("DefaultValue = %v, want 0", variable.DefaultValue)
	}
	if variable.Description != "The payment amount" {
		t.Errorf("Description = %s, want 'The payment amount'", variable.Description)
	}
	if variable.Example != 100.50 {
		t.Errorf("Example = %v, want 100.50", variable.Example)
	}
}

// ============================================================================
// RetryPolicy Tests
// ============================================================================

func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy()

	if policy.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", policy.MaxAttempts)
	}
	if policy.InitialInterval != 60 {
		t.Errorf("InitialInterval = %d, want 60", policy.InitialInterval)
	}
	if policy.MaxInterval != 3600 {
		t.Errorf("MaxInterval = %d, want 3600", policy.MaxInterval)
	}
	if policy.Multiplier != 2.0 {
		t.Errorf("Multiplier = %f, want 2.0", policy.Multiplier)
	}
}

func TestNewRetryPolicy_Success(t *testing.T) {
	policy, err := NewRetryPolicy(5, 30, 1800, 1.5)

	if err != nil {
		t.Fatalf("NewRetryPolicy() error = %v", err)
	}
	if policy.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", policy.MaxAttempts)
	}
	if policy.InitialInterval != 30 {
		t.Errorf("InitialInterval = %d, want 30", policy.InitialInterval)
	}
}

func TestNewRetryPolicy_Errors(t *testing.T) {
	tests := []struct {
		name            string
		maxAttempts     int
		initialInterval int
		maxInterval     int
		multiplier      float64
	}{
		{"negative attempts", -1, 60, 3600, 2.0},
		{"negative initial interval", 3, -1, 3600, 2.0},
		{"max < initial", 3, 100, 50, 2.0},
		{"multiplier < 1", 3, 60, 3600, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRetryPolicy(tt.maxAttempts, tt.initialInterval, tt.maxInterval, tt.multiplier)
			if err == nil {
				t.Error("NewRetryPolicy() should return error")
			}
		})
	}
}

func TestRetryPolicy_GetInterval(t *testing.T) {
	policy := RetryPolicy{
		MaxAttempts:     5,
		InitialInterval: 60,
		MaxInterval:     300,
		Multiplier:      2.0,
	}

	tests := []struct {
		attempt  int
		expected int
	}{
		{0, 60},  // Initial
		{1, 60},  // First retry
		{2, 120}, // Second retry (60 * 2)
		{3, 240}, // Third retry (120 * 2)
		{4, 300}, // Fourth retry (capped at max)
		{5, 300}, // Fifth retry (capped at max)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := policy.GetInterval(tt.attempt)
			if result != tt.expected {
				t.Errorf("GetInterval(%d) = %d, want %d", tt.attempt, result, tt.expected)
			}
		})
	}
}

func TestRetryPolicy_ShouldRetry(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: 3}

	tests := []struct {
		attempt  int
		expected bool
	}{
		{0, true},
		{1, true},
		{2, true},
		{3, false},
		{4, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := policy.ShouldRetry(tt.attempt)
			if result != tt.expected {
				t.Errorf("ShouldRetry(%d) = %v, want %v", tt.attempt, result, tt.expected)
			}
		})
	}
}
