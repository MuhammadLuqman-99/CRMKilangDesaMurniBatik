package domain

import (
	"testing"

	"github.com/google/uuid"
)

// ============================================================================
// NotificationTemplate Creation Tests
// ============================================================================

func TestNewNotificationTemplate_Success(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name      string
		code      string
		tmplName  string
		tmplType  NotificationType
	}{
		{"welcome template", "welcome_email", "Welcome Email", TypeWelcome},
		{"order confirmation", "order_confirmation", "Order Confirmation", TypeTransactional},
		{"password reset", "password_reset", "Password Reset", TypePasswordReset},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := NewNotificationTemplate(tenantID, tt.code, tt.tmplName, tt.tmplType)

			if err != nil {
				t.Fatalf("NewNotificationTemplate() error = %v", err)
			}

			if tmpl == nil {
				t.Fatal("NewNotificationTemplate() returned nil")
			}

			if tmpl.TenantID != tenantID {
				t.Errorf("TenantID = %v, want %v", tmpl.TenantID, tenantID)
			}
			if tmpl.Code != tt.code {
				t.Errorf("Code = %s, want %s", tmpl.Code, tt.code)
			}
			if tmpl.Name != tt.tmplName {
				t.Errorf("Name = %s, want %s", tmpl.Name, tt.tmplName)
			}
			if tmpl.Type != tt.tmplType {
				t.Errorf("Type = %v, want %v", tmpl.Type, tt.tmplType)
			}
			if !tmpl.IsActive {
				t.Error("Template should be active by default")
			}
			if tmpl.DefaultLocale != "en" {
				t.Errorf("DefaultLocale = %s, want 'en'", tmpl.DefaultLocale)
			}
			if tmpl.RenderEngine != "go-template" {
				t.Errorf("RenderEngine = %s, want 'go-template'", tmpl.RenderEngine)
			}
			if tmpl.TemplateVersion != 1 {
				t.Errorf("TemplateVersion = %d, want 1", tmpl.TemplateVersion)
			}
		})
	}
}

func TestNewNotificationTemplate_EmptyCode(t *testing.T) {
	tenantID := uuid.New()

	_, err := NewNotificationTemplate(tenantID, "", "Name", TypeTransactional)

	if err == nil {
		t.Error("NewNotificationTemplate() should return error for empty code")
	}
}

func TestNewNotificationTemplate_InvalidCodeFormat(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name string
		code string
	}{
		{"starts with number", "1invalid"},
		{"has uppercase", "InvalidCode"},
		{"has spaces", "invalid code"},
		{"has dashes", "invalid-code"},
		{"has special chars", "invalid@code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewNotificationTemplate(tenantID, tt.code, "Name", TypeTransactional)

			if err == nil {
				t.Errorf("NewNotificationTemplate() should return error for code %q", tt.code)
			}
		})
	}
}

func TestNewNotificationTemplate_CodeTooLong(t *testing.T) {
	tenantID := uuid.New()
	longCode := "a"
	for i := 0; i < 101; i++ {
		longCode += "a"
	}

	_, err := NewNotificationTemplate(tenantID, longCode, "Name", TypeTransactional)

	if err == nil {
		t.Error("NewNotificationTemplate() should return error for code too long")
	}
}

func TestNewNotificationTemplate_EmptyName(t *testing.T) {
	tenantID := uuid.New()

	_, err := NewNotificationTemplate(tenantID, "valid_code", "", TypeTransactional)

	if err == nil {
		t.Error("NewNotificationTemplate() should return error for empty name")
	}
}

func TestNewNotificationTemplate_NameTooLong(t *testing.T) {
	tenantID := uuid.New()
	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}

	_, err := NewNotificationTemplate(tenantID, "valid_code", longName, TypeTransactional)

	if err == nil {
		t.Error("NewNotificationTemplate() should return error for name too long")
	}
}

func TestNewNotificationTemplate_InvalidType(t *testing.T) {
	tenantID := uuid.New()

	_, err := NewNotificationTemplate(tenantID, "valid_code", "Name", NotificationType("invalid"))

	if err == nil {
		t.Error("NewNotificationTemplate() should return error for invalid type")
	}
}

// ============================================================================
// Template Content Tests
// ============================================================================

func createTestTemplate(t *testing.T) *NotificationTemplate {
	t.Helper()
	tmpl, err := NewNotificationTemplate(uuid.New(), "test_template", "Test Template", TypeTransactional)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}
	return tmpl
}

func TestNotificationTemplate_SetEmailTemplate(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &EmailTemplateContent{
		Subject:  "Welcome, {{.Name}}!",
		Body:     "Hello {{.Name}}",
		HTMLBody: "<h1>Hello {{.Name}}</h1>",
	}

	err := tmpl.SetEmailTemplate(content)

	if err != nil {
		t.Fatalf("SetEmailTemplate() error = %v", err)
	}
	if tmpl.EmailTemplate == nil {
		t.Fatal("EmailTemplate should be set")
	}
	if !tmpl.SupportsChannel(ChannelEmail) {
		t.Error("Template should support email channel")
	}
}

func TestNotificationTemplate_SetEmailTemplate_MissingSubject(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &EmailTemplateContent{
		Body: "Hello",
	}

	err := tmpl.SetEmailTemplate(content)

	if err == nil {
		t.Error("SetEmailTemplate() should return error for missing subject")
	}
}

func TestNotificationTemplate_SetEmailTemplate_MissingBody(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &EmailTemplateContent{
		Subject: "Subject",
	}

	err := tmpl.SetEmailTemplate(content)

	if err == nil {
		t.Error("SetEmailTemplate() should return error for missing body")
	}
}

func TestNotificationTemplate_SetSMSTemplate(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &SMSTemplateContent{
		Body:     "Your code is {{.Code}}",
		SenderID: "MySender",
	}

	err := tmpl.SetSMSTemplate(content)

	if err != nil {
		t.Fatalf("SetSMSTemplate() error = %v", err)
	}
	if tmpl.SMSTemplate == nil {
		t.Fatal("SMSTemplate should be set")
	}
	if !tmpl.SupportsChannel(ChannelSMS) {
		t.Error("Template should support SMS channel")
	}
}

func TestNotificationTemplate_SetSMSTemplate_EmptyBody(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &SMSTemplateContent{
		Body: "",
	}

	err := tmpl.SetSMSTemplate(content)

	if err != ErrSMSBodyRequired {
		t.Errorf("SetSMSTemplate() error = %v, want ErrSMSBodyRequired", err)
	}
}

func TestNotificationTemplate_SetSMSTemplate_BodyTooLong(t *testing.T) {
	tmpl := createTestTemplate(t)
	longBody := ""
	for i := 0; i < 200; i++ {
		longBody += "a"
	}
	content := &SMSTemplateContent{
		Body:      longBody,
		MaxLength: 160,
	}

	err := tmpl.SetSMSTemplate(content)

	if err != ErrSMSBodyTooLong {
		t.Errorf("SetSMSTemplate() error = %v, want ErrSMSBodyTooLong", err)
	}
}

func TestNotificationTemplate_SetPushTemplate(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &PushTemplateContent{
		Title: "New Message",
		Body:  "You have a new message",
	}

	err := tmpl.SetPushTemplate(content)

	if err != nil {
		t.Fatalf("SetPushTemplate() error = %v", err)
	}
	if tmpl.PushTemplate == nil {
		t.Fatal("PushTemplate should be set")
	}
	if !tmpl.SupportsChannel(ChannelPush) {
		t.Error("Template should support push channel")
	}
}

func TestNotificationTemplate_SetPushTemplate_Empty(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &PushTemplateContent{
		Title: "",
		Body:  "",
	}

	err := tmpl.SetPushTemplate(content)

	if err == nil {
		t.Error("SetPushTemplate() should return error for empty content")
	}
}

func TestNotificationTemplate_SetInAppTemplate(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &InAppTemplateContent{
		Title:       "Notification",
		Body:        "You have an update",
		Style:       "banner",
		Dismissable: true,
	}

	err := tmpl.SetInAppTemplate(content)

	if err != nil {
		t.Fatalf("SetInAppTemplate() error = %v", err)
	}
	if tmpl.InAppTemplate == nil {
		t.Fatal("InAppTemplate should be set")
	}
	if !tmpl.SupportsChannel(ChannelInApp) {
		t.Error("Template should support in-app channel")
	}
}

func TestNotificationTemplate_SetInAppTemplate_Empty(t *testing.T) {
	tmpl := createTestTemplate(t)
	content := &InAppTemplateContent{}

	err := tmpl.SetInAppTemplate(content)

	if err == nil {
		t.Error("SetInAppTemplate() should return error for empty content")
	}
}

func TestNotificationTemplate_SupportsChannel(t *testing.T) {
	tmpl := createTestTemplate(t)

	// No channels initially
	if tmpl.SupportsChannel(ChannelEmail) {
		t.Error("Template should not support email initially")
	}

	// Add email support
	tmpl.SetEmailTemplate(&EmailTemplateContent{Subject: "Test", Body: "Test"})

	if !tmpl.SupportsChannel(ChannelEmail) {
		t.Error("Template should support email after SetEmailTemplate")
	}
}

// ============================================================================
// Variable Tests
// ============================================================================

func TestNotificationTemplate_AddVariable(t *testing.T) {
	tmpl := createTestTemplate(t)
	variable := TemplateVariable{
		Name:     "username",
		Type:     "string",
		Required: true,
	}

	err := tmpl.AddVariable(variable)

	if err != nil {
		t.Fatalf("AddVariable() error = %v", err)
	}
	if len(tmpl.Variables) != 1 {
		t.Errorf("Variables count = %d, want 1", len(tmpl.Variables))
	}
}

func TestNotificationTemplate_AddVariable_Duplicate(t *testing.T) {
	tmpl := createTestTemplate(t)
	variable := TemplateVariable{Name: "username", Type: "string"}

	tmpl.AddVariable(variable)
	err := tmpl.AddVariable(variable)

	if err == nil {
		t.Error("AddVariable() should return error for duplicate variable")
	}
}

func TestNotificationTemplate_RemoveVariable(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.AddVariable(TemplateVariable{Name: "var1", Type: "string"})
	tmpl.AddVariable(TemplateVariable{Name: "var2", Type: "number"})

	tmpl.RemoveVariable("var1")

	if len(tmpl.Variables) != 1 {
		t.Errorf("Variables count = %d, want 1", len(tmpl.Variables))
	}
	if tmpl.Variables[0].Name != "var2" {
		t.Errorf("Remaining variable = %s, want 'var2'", tmpl.Variables[0].Name)
	}
}

func TestNotificationTemplate_GetVariable(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.AddVariable(TemplateVariable{Name: "username", Type: "string", Required: true})

	variable := tmpl.GetVariable("username")

	if variable == nil {
		t.Fatal("GetVariable() returned nil")
	}
	if variable.Name != "username" {
		t.Errorf("Name = %s, want 'username'", variable.Name)
	}

	// Non-existent variable
	missing := tmpl.GetVariable("nonexistent")
	if missing != nil {
		t.Error("GetVariable() should return nil for non-existent variable")
	}
}

func TestNotificationTemplate_GetRequiredVariables(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.AddVariable(TemplateVariable{Name: "required1", Type: "string", Required: true})
	tmpl.AddVariable(TemplateVariable{Name: "optional", Type: "string", Required: false})
	tmpl.AddVariable(TemplateVariable{Name: "required2", Type: "number", Required: true})

	required := tmpl.GetRequiredVariables()

	if len(required) != 2 {
		t.Errorf("Required variables count = %d, want 2", len(required))
	}
}

func TestNotificationTemplate_ValidateVariables(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.AddVariable(TemplateVariable{Name: "name", Type: "string", Required: true})
	tmpl.AddVariable(TemplateVariable{Name: "optional", Type: "string", Required: false})

	// Valid data
	data := map[string]interface{}{"name": "John"}
	err := tmpl.ValidateVariables(data)
	if err != nil {
		t.Errorf("ValidateVariables() error = %v", err)
	}

	// Missing required variable
	data = map[string]interface{}{"optional": "value"}
	err = tmpl.ValidateVariables(data)
	if err == nil {
		t.Error("ValidateVariables() should return error for missing required variable")
	}
}

// ============================================================================
// Localization Tests
// ============================================================================

func TestNotificationTemplate_AddLocalization(t *testing.T) {
	tmpl := createTestTemplate(t)
	localization := &TemplateLocalization{
		Locale: "ms",
		EmailTemplate: &EmailTemplateContent{
			Subject: "Selamat Datang",
			Body:    "Hello in Malay",
		},
	}

	err := tmpl.AddLocalization("ms", localization)

	if err != nil {
		t.Fatalf("AddLocalization() error = %v", err)
	}
	if tmpl.Localizations["ms"] == nil {
		t.Error("Localization should be added")
	}
}

func TestNotificationTemplate_AddLocalization_EmptyLocale(t *testing.T) {
	tmpl := createTestTemplate(t)

	err := tmpl.AddLocalization("", &TemplateLocalization{})

	if err == nil {
		t.Error("AddLocalization() should return error for empty locale")
	}
}

func TestNotificationTemplate_RemoveLocalization(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.AddLocalization("ms", &TemplateLocalization{Locale: "ms"})

	tmpl.RemoveLocalization("ms")

	if tmpl.Localizations["ms"] != nil {
		t.Error("Localization should be removed")
	}
}

func TestNotificationTemplate_GetLocalization(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.AddLocalization("ms", &TemplateLocalization{Locale: "ms"})

	loc := tmpl.GetLocalization("ms")
	if loc == nil {
		t.Error("GetLocalization() should return localization")
	}

	missing := tmpl.GetLocalization("fr")
	if missing != nil {
		t.Error("GetLocalization() should return nil for missing locale")
	}
}

// ============================================================================
// Status Tests
// ============================================================================

func TestNotificationTemplate_Activate(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.IsActive = false

	tmpl.Activate()

	if !tmpl.IsActive {
		t.Error("Template should be active")
	}

	events := tmpl.GetDomainEvents()
	if len(events) == 0 {
		t.Error("Should have domain event")
	}
}

func TestNotificationTemplate_Deactivate(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.IsActive = true

	tmpl.Deactivate()

	if tmpl.IsActive {
		t.Error("Template should be inactive")
	}
}

func TestNotificationTemplate_SetAsDefault(t *testing.T) {
	tmpl := createTestTemplate(t)

	tmpl.SetAsDefault()

	if !tmpl.IsDefault {
		t.Error("Template should be default")
	}
}

func TestNotificationTemplate_ClearDefault(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.IsDefault = true

	tmpl.ClearDefault()

	if tmpl.IsDefault {
		t.Error("Template should not be default")
	}
}

func TestNotificationTemplate_Lock(t *testing.T) {
	tmpl := createTestTemplate(t)

	tmpl.Lock()

	if !tmpl.IsLocked {
		t.Error("Template should be locked")
	}
}

// ============================================================================
// Draft/Publish Tests
// ============================================================================

func TestNotificationTemplate_SaveDraft(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetEmailTemplate(&EmailTemplateContent{Subject: "Test", Body: "Test"})
	userID := uuid.New()

	tmpl.SaveDraft(&userID)

	if tmpl.DraftContent == nil {
		t.Fatal("DraftContent should be set")
	}
	if tmpl.DraftContent.EmailTemplate == nil {
		t.Error("Draft should have email template")
	}
	if tmpl.DraftContent.UpdatedBy == nil || *tmpl.DraftContent.UpdatedBy != userID {
		t.Error("UpdatedBy should be set")
	}
}

func TestNotificationTemplate_Publish(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.DraftContent = &TemplateDraft{}
	originalVersion := tmpl.TemplateVersion

	tmpl.Publish()

	if tmpl.TemplateVersion != originalVersion+1 {
		t.Errorf("TemplateVersion = %d, want %d", tmpl.TemplateVersion, originalVersion+1)
	}
	if tmpl.PublishedAt == nil {
		t.Error("PublishedAt should be set")
	}
	if tmpl.DraftContent != nil {
		t.Error("DraftContent should be nil after publish")
	}
}

func TestNotificationTemplate_DiscardDraft(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.DraftContent = &TemplateDraft{}

	tmpl.DiscardDraft()

	if tmpl.DraftContent != nil {
		t.Error("DraftContent should be nil after discard")
	}
}

func TestNotificationTemplate_HasDraft(t *testing.T) {
	tmpl := createTestTemplate(t)

	if tmpl.HasDraft() {
		t.Error("HasDraft() should return false initially")
	}

	tmpl.DraftContent = &TemplateDraft{}
	if !tmpl.HasDraft() {
		t.Error("HasDraft() should return true")
	}
}

// ============================================================================
// Usage Tracking Tests
// ============================================================================

func TestNotificationTemplate_RecordUsage(t *testing.T) {
	tmpl := createTestTemplate(t)
	originalCount := tmpl.UsageCount

	tmpl.RecordUsage()
	tmpl.RecordUsage()

	if tmpl.UsageCount != originalCount+2 {
		t.Errorf("UsageCount = %d, want %d", tmpl.UsageCount, originalCount+2)
	}
	if tmpl.LastUsedAt == nil {
		t.Error("LastUsedAt should be set")
	}
}

// ============================================================================
// Tag Tests
// ============================================================================

func TestNotificationTemplate_AddTag(t *testing.T) {
	tmpl := createTestTemplate(t)

	tmpl.AddTag("important")
	tmpl.AddTag("  MARKETING  ") // Should be trimmed and lowercased

	if len(tmpl.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(tmpl.Tags))
	}
	if !tmpl.HasTag("important") {
		t.Error("Should have 'important' tag")
	}
	if !tmpl.HasTag("marketing") {
		t.Error("Should have 'marketing' tag")
	}
}

func TestNotificationTemplate_AddTag_Duplicate(t *testing.T) {
	tmpl := createTestTemplate(t)

	tmpl.AddTag("tag1")
	tmpl.AddTag("tag1")

	if len(tmpl.Tags) != 1 {
		t.Errorf("Tags count = %d, want 1 (no duplicates)", len(tmpl.Tags))
	}
}

func TestNotificationTemplate_AddTag_Empty(t *testing.T) {
	tmpl := createTestTemplate(t)

	tmpl.AddTag("")
	tmpl.AddTag("   ")

	if len(tmpl.Tags) != 0 {
		t.Errorf("Tags count = %d, want 0 (empty tags ignored)", len(tmpl.Tags))
	}
}

func TestNotificationTemplate_RemoveTag(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.AddTag("tag1")
	tmpl.AddTag("tag2")

	tmpl.RemoveTag("tag1")

	if len(tmpl.Tags) != 1 {
		t.Errorf("Tags count = %d, want 1", len(tmpl.Tags))
	}
	if tmpl.HasTag("tag1") {
		t.Error("Should not have 'tag1'")
	}
}

func TestNotificationTemplate_HasTag(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.AddTag("existing")

	if !tmpl.HasTag("existing") {
		t.Error("HasTag() should return true for existing tag")
	}
	if !tmpl.HasTag("  EXISTING  ") {
		t.Error("HasTag() should be case-insensitive and trim spaces")
	}
	if tmpl.HasTag("nonexistent") {
		t.Error("HasTag() should return false for non-existent tag")
	}
}

// ============================================================================
// Rendering Tests
// ============================================================================

func TestNotificationTemplate_RenderEmail(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetEmailTemplate(&EmailTemplateContent{
		Subject:     "Hello {{.Name}}",
		Body:        "Dear {{.Name}}, your order #{{.OrderID}} is ready.",
		HTMLBody:    "<p>Dear {{.Name}},</p><p>Your order #{{.OrderID}} is ready.</p>",
		TrackOpens:  true,
		TrackClicks: true,
	})

	data := map[string]interface{}{
		"Name":    "John",
		"OrderID": "12345",
	}

	rendered, err := tmpl.RenderEmail(data, "")

	if err != nil {
		t.Fatalf("RenderEmail() error = %v", err)
	}
	if rendered.Subject != "Hello John" {
		t.Errorf("Subject = %s, want 'Hello John'", rendered.Subject)
	}
	if rendered.Body != "Dear John, your order #12345 is ready." {
		t.Errorf("Body = %s", rendered.Body)
	}
	if !rendered.TrackOpens {
		t.Error("TrackOpens should be true")
	}
}

func TestNotificationTemplate_RenderEmail_MissingRequiredVariable(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetEmailTemplate(&EmailTemplateContent{
		Subject: "Hello {{.Name}}",
		Body:    "Hello",
	})
	tmpl.AddVariable(TemplateVariable{Name: "Name", Type: "string", Required: true})

	_, err := tmpl.RenderEmail(map[string]interface{}{}, "")

	if err == nil {
		t.Error("RenderEmail() should return error for missing required variable")
	}
}

func TestNotificationTemplate_RenderEmail_NoTemplate(t *testing.T) {
	tmpl := createTestTemplate(t)

	_, err := tmpl.RenderEmail(map[string]interface{}{}, "")

	if err == nil {
		t.Error("RenderEmail() should return error when no email template configured")
	}
}

func TestNotificationTemplate_RenderEmail_WithLocalization(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetEmailTemplate(&EmailTemplateContent{
		Subject: "Hello {{.Name}}",
		Body:    "Hello in English",
	})
	tmpl.AddLocalization("ms", &TemplateLocalization{
		Locale: "ms",
		EmailTemplate: &EmailTemplateContent{
			Subject: "Selamat {{.Name}}",
			Body:    "Hello in Malay",
		},
	})

	data := map[string]interface{}{"Name": "John"}

	// Default locale
	rendered, _ := tmpl.RenderEmail(data, "")
	if rendered.Subject != "Hello John" {
		t.Errorf("Default Subject = %s, want 'Hello John'", rendered.Subject)
	}

	// Malay locale
	rendered, _ = tmpl.RenderEmail(data, "ms")
	if rendered.Subject != "Selamat John" {
		t.Errorf("Malay Subject = %s, want 'Selamat John'", rendered.Subject)
	}
}

func TestNotificationTemplate_RenderSMS(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetSMSTemplate(&SMSTemplateContent{
		Body:     "Your OTP is {{.Code}}",
		SenderID: "MySender",
	})

	data := map[string]interface{}{"Code": "123456"}

	rendered, err := tmpl.RenderSMS(data, "")

	if err != nil {
		t.Fatalf("RenderSMS() error = %v", err)
	}
	if rendered.Body != "Your OTP is 123456" {
		t.Errorf("Body = %s, want 'Your OTP is 123456'", rendered.Body)
	}
	if rendered.SenderID != "MySender" {
		t.Errorf("SenderID = %s, want 'MySender'", rendered.SenderID)
	}
}

func TestNotificationTemplate_RenderSMS_NoTemplate(t *testing.T) {
	tmpl := createTestTemplate(t)

	_, err := tmpl.RenderSMS(map[string]interface{}{}, "")

	if err == nil {
		t.Error("RenderSMS() should return error when no SMS template configured")
	}
}

func TestNotificationTemplate_RenderPush(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetPushTemplate(&PushTemplateContent{
		Title: "New Message from {{.Sender}}",
		Body:  "{{.Message}}",
		Sound: "default",
	})

	data := map[string]interface{}{
		"Sender":  "John",
		"Message": "Hello!",
	}

	rendered, err := tmpl.RenderPush(data, "")

	if err != nil {
		t.Fatalf("RenderPush() error = %v", err)
	}
	if rendered.Title != "New Message from John" {
		t.Errorf("Title = %s", rendered.Title)
	}
	if rendered.Body != "Hello!" {
		t.Errorf("Body = %s", rendered.Body)
	}
	if rendered.Sound != "default" {
		t.Errorf("Sound = %s", rendered.Sound)
	}
}

func TestNotificationTemplate_RenderPush_NoTemplate(t *testing.T) {
	tmpl := createTestTemplate(t)

	_, err := tmpl.RenderPush(map[string]interface{}{}, "")

	if err == nil {
		t.Error("RenderPush() should return error when no push template configured")
	}
}

func TestNotificationTemplate_RenderInApp(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetInAppTemplate(&InAppTemplateContent{
		Title:       "{{.Title}}",
		Body:        "{{.Body}}",
		ActionLabel: "View",
		ActionURL:   "/orders/{{.OrderID}}",
		Dismissable: true,
	})

	data := map[string]interface{}{
		"Title":   "Order Shipped",
		"Body":    "Your order has been shipped!",
		"OrderID": "12345",
	}

	rendered, err := tmpl.RenderInApp(data, "")

	if err != nil {
		t.Fatalf("RenderInApp() error = %v", err)
	}
	if rendered.Title != "Order Shipped" {
		t.Errorf("Title = %s", rendered.Title)
	}
	if rendered.ActionLabel != "View" {
		t.Errorf("ActionLabel = %s", rendered.ActionLabel)
	}
	if !rendered.Dismissable {
		t.Error("Dismissable should be true")
	}
}

func TestNotificationTemplate_RenderInApp_NoTemplate(t *testing.T) {
	tmpl := createTestTemplate(t)

	_, err := tmpl.RenderInApp(map[string]interface{}{}, "")

	if err == nil {
		t.Error("RenderInApp() should return error when no in-app template configured")
	}
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestNotificationTemplate_Validate(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetEmailTemplate(&EmailTemplateContent{Subject: "Test", Body: "Test"})

	err := tmpl.Validate()

	if err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestNotificationTemplate_Validate_NoChannels(t *testing.T) {
	tmpl := createTestTemplate(t)

	err := tmpl.Validate()

	if err == nil {
		t.Error("Validate() should return error when no channels configured")
	}
}

func TestNotificationTemplate_Validate_MissingContent(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.Channels = []NotificationChannel{ChannelEmail}
	// EmailTemplate is nil

	err := tmpl.Validate()

	if err == nil {
		t.Error("Validate() should return error when channel added but content missing")
	}
}

// ============================================================================
// Clone Tests
// ============================================================================

func TestNotificationTemplate_Clone(t *testing.T) {
	tmpl := createTestTemplate(t)
	tmpl.SetEmailTemplate(&EmailTemplateContent{Subject: "Test", Body: "Test"})
	tmpl.AddVariable(TemplateVariable{Name: "var1", Type: "string"})
	tmpl.AddTag("tag1")
	tmpl.Description = "Original description"

	clone, err := tmpl.Clone("cloned_template", "Cloned Template")

	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}
	if clone.Code != "cloned_template" {
		t.Errorf("Clone Code = %s, want 'cloned_template'", clone.Code)
	}
	if clone.Name != "Cloned Template" {
		t.Errorf("Clone Name = %s, want 'Cloned Template'", clone.Name)
	}
	if clone.ID == tmpl.ID {
		t.Error("Clone should have different ID")
	}
	if clone.Description != "Original description" {
		t.Errorf("Clone Description = %s", clone.Description)
	}
	if clone.EmailTemplate == nil {
		t.Error("Clone should have email template")
	}
	if len(clone.Variables) != 1 {
		t.Errorf("Clone Variables count = %d, want 1", len(clone.Variables))
	}
	if len(clone.Tags) != 1 {
		t.Errorf("Clone Tags count = %d, want 1", len(clone.Tags))
	}
}
