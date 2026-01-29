package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Contact Creation Tests
// ============================================================================

func TestNewContact_Success(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	contact, err := NewContact(customerID, tenantID, "John", "Doe", "john.doe@example.com")

	if err != nil {
		t.Fatalf("NewContact() error = %v", err)
	}

	if contact == nil {
		t.Fatal("NewContact() returned nil")
	}

	if contact.CustomerID != customerID {
		t.Errorf("CustomerID = %v, want %v", contact.CustomerID, customerID)
	}
	if contact.TenantID != tenantID {
		t.Errorf("TenantID = %v, want %v", contact.TenantID, tenantID)
	}
	if contact.Name.FirstName != "John" {
		t.Errorf("FirstName = %s, want 'John'", contact.Name.FirstName)
	}
	if contact.Name.LastName != "Doe" {
		t.Errorf("LastName = %s, want 'Doe'", contact.Name.LastName)
	}
	if contact.Email.String() != "john.doe@example.com" {
		t.Errorf("Email = %s, want 'john.doe@example.com'", contact.Email.String())
	}
	if contact.Status != ContactStatusActive {
		t.Errorf("Status = %v, want %v", contact.Status, ContactStatusActive)
	}
	if contact.Role != ContactRoleOther {
		t.Errorf("Role = %v, want %v", contact.Role, ContactRoleOther)
	}
	if contact.CommPreference != CommPrefEmail {
		t.Errorf("CommPreference = %v, want %v", contact.CommPreference, CommPrefEmail)
	}
}

func TestNewContact_InvalidEmail(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	_, err := NewContact(customerID, tenantID, "John", "Doe", "invalid-email")

	if err == nil {
		t.Error("NewContact() should return error for invalid email")
	}
}

func TestNewContact_EmptyName(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	_, err := NewContact(customerID, tenantID, "", "", "john@example.com")

	if err == nil {
		t.Error("NewContact() should return error for empty name")
	}
}

// ============================================================================
// ContactBuilder Tests
// ============================================================================

func TestContactBuilder_Success(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()
	createdBy := uuid.New()

	contact, err := NewContactBuilder(customerID, tenantID).
		WithName("John", "Doe").
		WithEmail("john.doe@example.com").
		WithPhone("+60123456789", PhoneTypeMobile, true).
		WithJobTitle("Manager").
		WithDepartment("Sales").
		WithRole(ContactRoleDecisionMaker).
		AsPrimary().
		WithCommPreference(CommPrefPhone).
		WithNotes("Important contact").
		WithTags("vip", "enterprise").
		WithCreatedBy(createdBy).
		Build()

	if err != nil {
		t.Fatalf("ContactBuilder.Build() error = %v", err)
	}

	if contact.JobTitle != "Manager" {
		t.Errorf("JobTitle = %s, want 'Manager'", contact.JobTitle)
	}
	if contact.Department != "Sales" {
		t.Errorf("Department = %s, want 'Sales'", contact.Department)
	}
	if contact.Role != ContactRoleDecisionMaker {
		t.Errorf("Role = %v, want %v", contact.Role, ContactRoleDecisionMaker)
	}
	if !contact.IsPrimary {
		t.Error("IsPrimary should be true")
	}
	if contact.CommPreference != CommPrefPhone {
		t.Errorf("CommPreference = %v, want %v", contact.CommPreference, CommPrefPhone)
	}
	if contact.Notes != "Important contact" {
		t.Errorf("Notes = %s, want 'Important contact'", contact.Notes)
	}
	if len(contact.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(contact.Tags))
	}
	if len(contact.PhoneNumbers) != 1 {
		t.Errorf("PhoneNumbers count = %d, want 1", len(contact.PhoneNumbers))
	}
}

func TestContactBuilder_WithFullName(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	contact, err := NewContactBuilder(customerID, tenantID).
		WithFullName("Dr.", "John", "Robert", "Doe", "Jr.").
		WithEmail("john.doe@example.com").
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if contact.Name.Title != "Dr." {
		t.Errorf("Title = %s, want 'Dr.'", contact.Name.Title)
	}
	if contact.Name.MiddleName != "Robert" {
		t.Errorf("MiddleName = %s, want 'Robert'", contact.Name.MiddleName)
	}
	if contact.Name.Suffix != "Jr." {
		t.Errorf("Suffix = %s, want 'Jr.'", contact.Name.Suffix)
	}
}

func TestContactBuilder_WithAddress(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	contact, err := NewContactBuilder(customerID, tenantID).
		WithName("John", "Doe").
		WithEmail("john@example.com").
		WithAddress("123 Main St", "Kuala Lumpur", "50000", "MY", AddressTypeOffice, true).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(contact.Addresses) != 1 {
		t.Errorf("Addresses count = %d, want 1", len(contact.Addresses))
	}
	if contact.Addresses[0].City != "Kuala Lumpur" {
		t.Errorf("City = %s, want 'Kuala Lumpur'", contact.Addresses[0].City)
	}
}

func TestContactBuilder_WithSocialProfile(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	contact, err := NewContactBuilder(customerID, tenantID).
		WithName("John", "Doe").
		WithEmail("john@example.com").
		WithSocialProfile(SocialPlatformLinkedIn, "https://linkedin.com/in/johndoe").
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(contact.SocialProfiles) != 1 {
		t.Errorf("SocialProfiles count = %d, want 1", len(contact.SocialProfiles))
	}
}

func TestContactBuilder_WithBirthday(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()
	birthday := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	contact, err := NewContactBuilder(customerID, tenantID).
		WithName("John", "Doe").
		WithEmail("john@example.com").
		WithBirthday(birthday).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if contact.Birthday == nil {
		t.Fatal("Birthday should be set")
	}
	if *contact.Birthday != birthday {
		t.Errorf("Birthday = %v, want %v", *contact.Birthday, birthday)
	}
}

func TestContactBuilder_WithCustomField(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	contact, err := NewContactBuilder(customerID, tenantID).
		WithName("John", "Doe").
		WithEmail("john@example.com").
		WithCustomField("preferred_time", "morning").
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if contact.CustomFields["preferred_time"] != "morning" {
		t.Errorf("CustomField = %v, want 'morning'", contact.CustomFields["preferred_time"])
	}
}

func TestContactBuilder_ValidationErrors(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	// Missing name and contact info
	_, err := NewContactBuilder(customerID, tenantID).Build()

	if err == nil {
		t.Error("Build() should return error for missing required fields")
	}
}

func TestContactBuilder_RequiresEmailOrPhone(t *testing.T) {
	customerID := uuid.New()
	tenantID := uuid.New()

	// Name only, no email or phone
	_, err := NewContactBuilder(customerID, tenantID).
		WithName("John", "Doe").
		Build()

	if err == nil {
		t.Error("Build() should return error when no email or phone provided")
	}

	// With phone only (no email) - should succeed
	contact, err := NewContactBuilder(customerID, tenantID).
		WithName("John", "Doe").
		WithPhone("+60123456789", PhoneTypeMobile, true).
		Build()

	if err != nil {
		t.Errorf("Build() with phone only should succeed, got error: %v", err)
	}
	if contact == nil {
		t.Error("Contact should not be nil")
	}
}

// ============================================================================
// Contact Behavior Tests
// ============================================================================

func createTestContact(t *testing.T) *Contact {
	t.Helper()

	contact, err := NewContactBuilder(uuid.New(), uuid.New()).
		WithName("John", "Doe").
		WithEmail("john.doe@example.com").
		WithPhone("+60123456789", PhoneTypeMobile, true).
		Build()

	if err != nil {
		t.Fatalf("Failed to create test contact: %v", err)
	}

	return contact
}

func TestContact_UpdateName(t *testing.T) {
	contact := createTestContact(t)
	newName, _ := NewPersonName("Jane", "Smith")

	contact.UpdateName(newName)

	if contact.Name.FirstName != "Jane" {
		t.Errorf("FirstName = %s, want 'Jane'", contact.Name.FirstName)
	}
	if contact.Name.LastName != "Smith" {
		t.Errorf("LastName = %s, want 'Smith'", contact.Name.LastName)
	}
}

func TestContact_UpdateEmail(t *testing.T) {
	contact := createTestContact(t)
	newEmail, _ := NewEmail("jane@example.com")

	err := contact.UpdateEmail(newEmail)

	if err != nil {
		t.Fatalf("UpdateEmail() error = %v", err)
	}
	if contact.Email.String() != "jane@example.com" {
		t.Errorf("Email = %s, want 'jane@example.com'", contact.Email.String())
	}
}

func TestContact_AddPhoneNumber(t *testing.T) {
	contact := createTestContact(t)
	initialCount := len(contact.PhoneNumbers)
	newPhone, _ := NewPhoneNumber("+60198765432", PhoneTypeWork)

	contact.AddPhoneNumber(newPhone)

	if len(contact.PhoneNumbers) != initialCount+1 {
		t.Errorf("PhoneNumbers count = %d, want %d", len(contact.PhoneNumbers), initialCount+1)
	}
}

func TestContact_AddPhoneNumber_PrimaryOverride(t *testing.T) {
	contact := createTestContact(t)
	// First phone should be primary
	if len(contact.PhoneNumbers) > 0 && !contact.PhoneNumbers[0].IsPrimary() {
		contact.PhoneNumbers[0].SetPrimary(true)
	}

	newPhone, _ := NewPhoneNumberWithPrimary("+60198765432", PhoneTypeWork, true)
	contact.AddPhoneNumber(newPhone)

	// Old primary should be unset
	if contact.PhoneNumbers[0].IsPrimary() {
		t.Error("Old primary should be unset")
	}
	// New phone should be primary
	lastPhone := contact.PhoneNumbers[len(contact.PhoneNumbers)-1]
	if !lastPhone.IsPrimary() {
		t.Error("New phone should be primary")
	}
}

func TestContact_RemovePhoneNumber(t *testing.T) {
	contact := createTestContact(t)
	if len(contact.PhoneNumbers) == 0 {
		t.Skip("No phone numbers to remove")
	}

	e164 := contact.PhoneNumbers[0].E164()
	initialCount := len(contact.PhoneNumbers)

	result := contact.RemovePhoneNumber(e164)

	if !result {
		t.Error("RemovePhoneNumber() should return true")
	}
	if len(contact.PhoneNumbers) != initialCount-1 {
		t.Errorf("PhoneNumbers count = %d, want %d", len(contact.PhoneNumbers), initialCount-1)
	}
}

func TestContact_RemovePhoneNumber_NotFound(t *testing.T) {
	contact := createTestContact(t)

	result := contact.RemovePhoneNumber("+00000000000")

	if result {
		t.Error("RemovePhoneNumber() should return false for non-existent number")
	}
}

func TestContact_SetPrimaryPhone(t *testing.T) {
	contact := createTestContact(t)
	newPhone, _ := NewPhoneNumber("+60198765432", PhoneTypeWork)
	contact.AddPhoneNumber(newPhone)

	result := contact.SetPrimaryPhone(newPhone.E164())

	if !result {
		t.Error("SetPrimaryPhone() should return true")
	}

	// Find the phone and check it's primary
	for _, phone := range contact.PhoneNumbers {
		if phone.E164() == newPhone.E164() {
			if !phone.IsPrimary() {
				t.Error("Phone should be primary")
			}
		} else {
			if phone.IsPrimary() {
				t.Error("Other phones should not be primary")
			}
		}
	}
}

func TestContact_GetPrimaryPhone(t *testing.T) {
	contact := createTestContact(t)

	// Set first phone as primary
	if len(contact.PhoneNumbers) > 0 {
		contact.PhoneNumbers[0].SetPrimary(true)
	}

	primary := contact.GetPrimaryPhone()

	if primary == nil && len(contact.PhoneNumbers) > 0 {
		t.Error("GetPrimaryPhone() should return a phone")
	}
	if primary != nil && !primary.IsPrimary() {
		t.Error("Returned phone should be primary")
	}
}

func TestContact_AddAddress(t *testing.T) {
	contact := createTestContact(t)
	address, _ := NewAddress("123 Main St", "Kuala Lumpur", "50000", "MY", AddressTypeOffice)

	contact.AddAddress(address)

	if len(contact.Addresses) != 1 {
		t.Errorf("Addresses count = %d, want 1", len(contact.Addresses))
	}
}

func TestContact_AddAddress_PrimaryOverride(t *testing.T) {
	contact := createTestContact(t)
	addr1, _ := NewAddress("First St", "KL", "50000", "MY", AddressTypeOffice)
	addr1.IsPrimary = true
	contact.AddAddress(addr1)

	addr2, _ := NewAddress("Second St", "KL", "50001", "MY", AddressTypeOffice)
	addr2.IsPrimary = true
	contact.AddAddress(addr2)

	// First address should no longer be primary for same type
	if contact.Addresses[0].IsPrimary {
		t.Error("First address should not be primary after adding new primary of same type")
	}
}

func TestContact_RemoveAddress(t *testing.T) {
	contact := createTestContact(t)
	address, _ := NewAddress("123 Main St", "Kuala Lumpur", "50000", "MY", AddressTypeOffice)
	contact.AddAddress(address)

	result := contact.RemoveAddress("123 Main St", "50000")

	if !result {
		t.Error("RemoveAddress() should return true")
	}
	if len(contact.Addresses) != 0 {
		t.Errorf("Addresses count = %d, want 0", len(contact.Addresses))
	}
}

func TestContact_GetPrimaryAddress(t *testing.T) {
	contact := createTestContact(t)
	addr1, _ := NewAddress("First St", "KL", "50000", "MY", AddressTypeOffice)
	addr2, _ := NewAddress("Second St", "KL", "50001", "MY", AddressTypeOffice)
	addr2.IsPrimary = true
	contact.AddAddress(addr1)
	contact.AddAddress(addr2)

	primary := contact.GetPrimaryAddress(AddressTypeOffice)

	if primary == nil {
		t.Fatal("GetPrimaryAddress() should return address")
	}
	if primary.Line1 != "Second St" {
		t.Errorf("Primary address Line1 = %s, want 'Second St'", primary.Line1)
	}
}

func TestContact_AddSocialProfile(t *testing.T) {
	contact := createTestContact(t)
	profile, _ := NewSocialProfile(SocialPlatformLinkedIn, "https://linkedin.com/in/johndoe")

	contact.AddSocialProfile(profile)

	if len(contact.SocialProfiles) != 1 {
		t.Errorf("SocialProfiles count = %d, want 1", len(contact.SocialProfiles))
	}
}

func TestContact_AddSocialProfile_ReplacesSamePlatform(t *testing.T) {
	contact := createTestContact(t)
	profile1, _ := NewSocialProfile(SocialPlatformLinkedIn, "https://linkedin.com/in/johndoe")
	profile2, _ := NewSocialProfile(SocialPlatformLinkedIn, "https://linkedin.com/in/johndoe2")

	contact.AddSocialProfile(profile1)
	contact.AddSocialProfile(profile2)

	if len(contact.SocialProfiles) != 1 {
		t.Errorf("SocialProfiles count = %d, want 1 (should replace)", len(contact.SocialProfiles))
	}
	if contact.SocialProfiles[0].URL != profile2.URL {
		t.Error("Should have the second profile URL")
	}
}

func TestContact_RemoveSocialProfile(t *testing.T) {
	contact := createTestContact(t)
	profile, _ := NewSocialProfile(SocialPlatformLinkedIn, "https://linkedin.com/in/johndoe")
	contact.AddSocialProfile(profile)

	result := contact.RemoveSocialProfile(SocialPlatformLinkedIn)

	if !result {
		t.Error("RemoveSocialProfile() should return true")
	}
	if len(contact.SocialProfiles) != 0 {
		t.Errorf("SocialProfiles count = %d, want 0", len(contact.SocialProfiles))
	}
}

func TestContact_GetSocialProfile(t *testing.T) {
	contact := createTestContact(t)
	profile, _ := NewSocialProfile(SocialPlatformLinkedIn, "https://linkedin.com/in/johndoe")
	contact.AddSocialProfile(profile)

	found := contact.GetSocialProfile(SocialPlatformLinkedIn)
	missing := contact.GetSocialProfile(SocialPlatformFacebook)

	if found == nil {
		t.Error("GetSocialProfile() should return profile")
	}
	if missing != nil {
		t.Error("GetSocialProfile() should return nil for missing platform")
	}
}

func TestContact_UpdateJobInfo(t *testing.T) {
	contact := createTestContact(t)

	contact.UpdateJobInfo("Senior Manager", "Engineering", ContactRoleTechnical)

	if contact.JobTitle != "Senior Manager" {
		t.Errorf("JobTitle = %s, want 'Senior Manager'", contact.JobTitle)
	}
	if contact.Department != "Engineering" {
		t.Errorf("Department = %s, want 'Engineering'", contact.Department)
	}
	if contact.Role != ContactRoleTechnical {
		t.Errorf("Role = %v, want %v", contact.Role, ContactRoleTechnical)
	}
}

// ============================================================================
// Contact Status Tests
// ============================================================================

func TestContact_Activate(t *testing.T) {
	contact := createTestContact(t)
	contact.Status = ContactStatusInactive

	err := contact.Activate()

	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if contact.Status != ContactStatusActive {
		t.Errorf("Status = %v, want %v", contact.Status, ContactStatusActive)
	}
}

func TestContact_Activate_Blocked(t *testing.T) {
	contact := createTestContact(t)
	contact.Status = ContactStatusBlocked

	err := contact.Activate()

	if err != ErrContactBlocked {
		t.Errorf("Activate() error = %v, want ErrContactBlocked", err)
	}
}

func TestContact_Deactivate(t *testing.T) {
	contact := createTestContact(t)

	contact.Deactivate()

	if contact.Status != ContactStatusInactive {
		t.Errorf("Status = %v, want %v", contact.Status, ContactStatusInactive)
	}
}

func TestContact_Block(t *testing.T) {
	contact := createTestContact(t)

	contact.Block()

	if contact.Status != ContactStatusBlocked {
		t.Errorf("Status = %v, want %v", contact.Status, ContactStatusBlocked)
	}
}

func TestContact_Unblock(t *testing.T) {
	contact := createTestContact(t)
	contact.Status = ContactStatusBlocked

	contact.Unblock()

	if contact.Status != ContactStatusInactive {
		t.Errorf("Status = %v, want %v", contact.Status, ContactStatusInactive)
	}
}

// ============================================================================
// Contact Marketing Preferences Tests
// ============================================================================

func TestContact_OptOutMarketing(t *testing.T) {
	contact := createTestContact(t)
	contact.OptedOutMarketing = false
	now := time.Now()
	contact.MarketingConsent = &now

	contact.OptOutMarketing()

	if !contact.OptedOutMarketing {
		t.Error("OptedOutMarketing should be true")
	}
	if contact.MarketingConsent != nil {
		t.Error("MarketingConsent should be nil after opt out")
	}
}

func TestContact_OptInMarketing(t *testing.T) {
	contact := createTestContact(t)
	contact.OptedOutMarketing = true

	contact.OptInMarketing()

	if contact.OptedOutMarketing {
		t.Error("OptedOutMarketing should be false")
	}
	if contact.MarketingConsent == nil {
		t.Error("MarketingConsent should be set")
	}
}

func TestContact_SetCommunicationPreference(t *testing.T) {
	contact := createTestContact(t)

	contact.SetCommunicationPreference(CommPrefWhatsApp)

	if contact.CommPreference != CommPrefWhatsApp {
		t.Errorf("CommPreference = %v, want %v", contact.CommPreference, CommPrefWhatsApp)
	}
}

// ============================================================================
// Contact Engagement Tests
// ============================================================================

func TestContact_RecordContact(t *testing.T) {
	contact := createTestContact(t)
	contact.LastContactedAt = nil

	contact.RecordContact()

	if contact.LastContactedAt == nil {
		t.Error("LastContactedAt should be set")
	}
}

func TestContact_SetNextFollowUp(t *testing.T) {
	contact := createTestContact(t)
	followUpAt := time.Now().Add(24 * time.Hour)

	contact.SetNextFollowUp(followUpAt)

	if contact.NextFollowUpAt == nil {
		t.Fatal("NextFollowUpAt should be set")
	}
	if *contact.NextFollowUpAt != followUpAt {
		t.Errorf("NextFollowUpAt = %v, want %v", *contact.NextFollowUpAt, followUpAt)
	}
}

func TestContact_ClearNextFollowUp(t *testing.T) {
	contact := createTestContact(t)
	now := time.Now()
	contact.NextFollowUpAt = &now

	contact.ClearNextFollowUp()

	if contact.NextFollowUpAt != nil {
		t.Error("NextFollowUpAt should be nil")
	}
}

func TestContact_UpdateEngagementScore(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		expected int
	}{
		{"normal score", 50, 50},
		{"negative becomes 0", -10, 0},
		{"over 100 becomes 100", 150, 100},
		{"exact 100", 100, 100},
		{"exact 0", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contact := createTestContact(t)

			contact.UpdateEngagementScore(tt.score)

			if contact.EngagementScore != tt.expected {
				t.Errorf("EngagementScore = %d, want %d", contact.EngagementScore, tt.expected)
			}
		})
	}
}

func TestContact_IncrementEngagement(t *testing.T) {
	contact := createTestContact(t)
	contact.EngagementScore = 50

	contact.IncrementEngagement(10)

	if contact.EngagementScore != 60 {
		t.Errorf("EngagementScore = %d, want 60", contact.EngagementScore)
	}
}

// ============================================================================
// Contact Tag Tests
// ============================================================================

func TestContact_AddTag(t *testing.T) {
	contact := createTestContact(t)

	contact.AddTag("vip")
	contact.AddTag("  ENTERPRISE  ") // Should be normalized

	if len(contact.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(contact.Tags))
	}
	if !contact.HasTag("vip") {
		t.Error("Should have 'vip' tag")
	}
	if !contact.HasTag("enterprise") {
		t.Error("Should have 'enterprise' tag")
	}
}

func TestContact_AddTag_Duplicate(t *testing.T) {
	contact := createTestContact(t)

	contact.AddTag("vip")
	contact.AddTag("vip")

	if len(contact.Tags) != 1 {
		t.Errorf("Tags count = %d, want 1 (no duplicates)", len(contact.Tags))
	}
}

func TestContact_AddTag_Empty(t *testing.T) {
	contact := createTestContact(t)

	contact.AddTag("")
	contact.AddTag("   ")

	if len(contact.Tags) != 0 {
		t.Errorf("Tags count = %d, want 0 (empty tags ignored)", len(contact.Tags))
	}
}

func TestContact_RemoveTag(t *testing.T) {
	contact := createTestContact(t)
	contact.AddTag("vip")
	contact.AddTag("enterprise")

	contact.RemoveTag("vip")

	if len(contact.Tags) != 1 {
		t.Errorf("Tags count = %d, want 1", len(contact.Tags))
	}
	if contact.HasTag("vip") {
		t.Error("Should not have 'vip' tag")
	}
}

func TestContact_HasTag(t *testing.T) {
	contact := createTestContact(t)
	contact.AddTag("vip")

	if !contact.HasTag("vip") {
		t.Error("HasTag() should return true for existing tag")
	}
	if !contact.HasTag("  VIP  ") {
		t.Error("HasTag() should be case-insensitive and trim spaces")
	}
	if contact.HasTag("nonexistent") {
		t.Error("HasTag() should return false for non-existent tag")
	}
}

// ============================================================================
// Contact Custom Fields Tests
// ============================================================================

func TestContact_SetCustomField(t *testing.T) {
	contact := createTestContact(t)

	contact.SetCustomField("preferred_color", "blue")

	if contact.CustomFields["preferred_color"] != "blue" {
		t.Errorf("CustomField = %v, want 'blue'", contact.CustomFields["preferred_color"])
	}
}

func TestContact_SetCustomField_InitializesMap(t *testing.T) {
	contact := &Contact{CustomFields: nil}

	contact.SetCustomField("key", "value")

	if contact.CustomFields == nil {
		t.Fatal("CustomFields should be initialized")
	}
}

func TestContact_GetCustomField(t *testing.T) {
	contact := createTestContact(t)
	contact.SetCustomField("key", "value")

	val, ok := contact.GetCustomField("key")
	if !ok || val != "value" {
		t.Errorf("GetCustomField() = %v, %v, want 'value', true", val, ok)
	}

	val, ok = contact.GetCustomField("nonexistent")
	if ok {
		t.Errorf("GetCustomField() should return false for non-existent key")
	}
}

func TestContact_RemoveCustomField(t *testing.T) {
	contact := createTestContact(t)
	contact.SetCustomField("key", "value")

	contact.RemoveCustomField("key")

	_, ok := contact.GetCustomField("key")
	if ok {
		t.Error("Field should be removed")
	}
}

// ============================================================================
// Contact Query Methods Tests
// ============================================================================

func TestContact_IsActive(t *testing.T) {
	contact := createTestContact(t)
	contact.Status = ContactStatusActive

	if !contact.IsActive() {
		t.Error("IsActive() should return true")
	}

	contact.Status = ContactStatusInactive
	if contact.IsActive() {
		t.Error("IsActive() should return false for inactive")
	}
}

func TestContact_IsBlocked(t *testing.T) {
	contact := createTestContact(t)
	contact.Status = ContactStatusBlocked

	if !contact.IsBlocked() {
		t.Error("IsBlocked() should return true")
	}

	contact.Status = ContactStatusActive
	if contact.IsBlocked() {
		t.Error("IsBlocked() should return false")
	}
}

func TestContact_CanReceiveMarketing(t *testing.T) {
	contact := createTestContact(t)
	contact.Status = ContactStatusActive
	contact.OptedOutMarketing = false

	if !contact.CanReceiveMarketing() {
		t.Error("CanReceiveMarketing() should return true")
	}

	contact.OptedOutMarketing = true
	if contact.CanReceiveMarketing() {
		t.Error("CanReceiveMarketing() should return false when opted out")
	}

	contact.OptedOutMarketing = false
	contact.Status = ContactStatusInactive
	if contact.CanReceiveMarketing() {
		t.Error("CanReceiveMarketing() should return false when inactive")
	}
}

func TestContact_NeedsFollowUp(t *testing.T) {
	contact := createTestContact(t)

	// No follow-up set
	if contact.NeedsFollowUp() {
		t.Error("NeedsFollowUp() should return false when no follow-up set")
	}

	// Past follow-up
	pastTime := time.Now().Add(-time.Hour)
	contact.NextFollowUpAt = &pastTime
	if !contact.NeedsFollowUp() {
		t.Error("NeedsFollowUp() should return true for past follow-up")
	}

	// Future follow-up
	futureTime := time.Now().Add(time.Hour)
	contact.NextFollowUpAt = &futureTime
	if contact.NeedsFollowUp() {
		t.Error("NeedsFollowUp() should return false for future follow-up")
	}
}

func TestContact_DisplayName(t *testing.T) {
	contact := createTestContact(t)

	displayName := contact.DisplayName()

	if displayName == "" {
		t.Error("DisplayName() should not be empty")
	}
}

func TestContact_FullName(t *testing.T) {
	contact := createTestContact(t)

	fullName := contact.FullName()

	if fullName == "" {
		t.Error("FullName() should not be empty")
	}
}

func TestContact_GetBestContactMethod(t *testing.T) {
	contact := createTestContact(t)
	contact.CommPreference = CommPrefEmail

	method := contact.GetBestContactMethod()

	if method == "" {
		t.Error("GetBestContactMethod() should return a contact method")
	}
	if contact.CommPreference == CommPrefEmail && !strings.HasPrefix(method, "email:") {
		t.Errorf("Method = %s, expected email prefix for email preference", method)
	}
}

// ============================================================================
// Contact Profile Tests
// ============================================================================

func TestContact_SetProfilePhoto(t *testing.T) {
	contact := createTestContact(t)

	contact.SetProfilePhoto("https://example.com/photo.jpg")

	if contact.ProfilePhotoURL != "https://example.com/photo.jpg" {
		t.Errorf("ProfilePhotoURL = %s", contact.ProfilePhotoURL)
	}
}

func TestContact_SetLinkedIn(t *testing.T) {
	contact := createTestContact(t)

	contact.SetLinkedIn("https://linkedin.com/in/johndoe")

	if contact.LinkedInURL != "https://linkedin.com/in/johndoe" {
		t.Errorf("LinkedInURL = %s", contact.LinkedInURL)
	}
}

func TestContact_SetReportsTo(t *testing.T) {
	contact := createTestContact(t)
	managerID := uuid.New()

	contact.SetReportsTo(&managerID)

	if contact.ReportsTo == nil || *contact.ReportsTo != managerID {
		t.Errorf("ReportsTo = %v, want %v", contact.ReportsTo, managerID)
	}
}

func TestContact_SetPrimary(t *testing.T) {
	contact := createTestContact(t)

	contact.SetPrimary(true)

	if !contact.IsPrimary {
		t.Error("IsPrimary should be true")
	}

	contact.SetPrimary(false)
	if contact.IsPrimary {
		t.Error("IsPrimary should be false")
	}
}

