package domain

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Customer Creation Tests
// ============================================================================

func TestNewCustomer_Success(t *testing.T) {
	tenantID := uuid.New()

	tests := []struct {
		name         string
		customerName string
		customerType CustomerType
	}{
		{"individual customer", "John Doe", CustomerTypeIndividual},
		{"company customer", "ACME Corp", CustomerTypeCompany},
		{"partner customer", "Partner Inc", CustomerTypePartner},
		{"reseller customer", "Reseller Ltd", CustomerTypeReseller},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := NewCustomer(tenantID, tt.customerName, tt.customerType)

			if err != nil {
				t.Fatalf("NewCustomer() error = %v", err)
			}

			if customer == nil {
				t.Fatal("NewCustomer() returned nil")
			}

			if customer.TenantID != tenantID {
				t.Errorf("TenantID = %v, want %v", customer.TenantID, tenantID)
			}
			if customer.Name != tt.customerName {
				t.Errorf("Name = %s, want %s", customer.Name, tt.customerName)
			}
			if customer.Type != tt.customerType {
				t.Errorf("Type = %v, want %v", customer.Type, tt.customerType)
			}
			if customer.Status != CustomerStatusLead {
				t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusLead)
			}
			if customer.Tier != CustomerTierStandard {
				t.Errorf("Tier = %v, want %v", customer.Tier, CustomerTierStandard)
			}
			if customer.Source != CustomerSourceDirect {
				t.Errorf("Source = %v, want %v", customer.Source, CustomerSourceDirect)
			}
			if customer.Code == "" {
				t.Error("Code should be generated")
			}
			if customer.Version != 1 {
				t.Errorf("Version = %d, want 1", customer.Version)
			}
		})
	}
}

func TestNewCustomer_EmptyName(t *testing.T) {
	tenantID := uuid.New()

	_, err := NewCustomer(tenantID, "", CustomerTypeIndividual)

	if err == nil {
		t.Error("NewCustomer() should return error for empty name")
	}
}

func TestNewCustomer_NameTooLong(t *testing.T) {
	tenantID := uuid.New()
	longName := ""
	for i := 0; i < 300; i++ {
		longName += "a"
	}

	_, err := NewCustomer(tenantID, longName, CustomerTypeIndividual)

	if err == nil {
		t.Error("NewCustomer() should return error for name too long")
	}
}

func TestNewCustomer_NilTenantID(t *testing.T) {
	_, err := NewCustomer(uuid.Nil, "Test Customer", CustomerTypeIndividual)

	if err == nil {
		t.Error("NewCustomer() should return error for nil tenant ID")
	}
}

// ============================================================================
// CustomerBuilder Tests
// ============================================================================

func TestCustomerBuilder_Success(t *testing.T) {
	tenantID := uuid.New()
	ownerID := uuid.New()
	createdBy := uuid.New()

	customer, err := NewCustomerBuilder(tenantID, "Test Company", CustomerTypeCompany).
		WithCode("TEST001").
		WithEmail("contact@test.com").
		WithPhone("+60123456789", PhoneTypeMobile, true).
		WithWebsite("https://test.com").
		WithSource(CustomerSourceWebsite).
		WithTier(CustomerTierGold).
		WithOwner(ownerID).
		WithTags("vip", "enterprise").
		WithNotes("Important customer").
		WithCreatedBy(createdBy).
		Build()

	if err != nil {
		t.Fatalf("CustomerBuilder.Build() error = %v", err)
	}

	if customer.Code != "TEST001" {
		t.Errorf("Code = %s, want 'TEST001'", customer.Code)
	}
	if customer.Email.String() != "contact@test.com" {
		t.Errorf("Email = %s, want 'contact@test.com'", customer.Email.String())
	}
	if customer.Source != CustomerSourceWebsite {
		t.Errorf("Source = %v, want %v", customer.Source, CustomerSourceWebsite)
	}
	if customer.Tier != CustomerTierGold {
		t.Errorf("Tier = %v, want %v", customer.Tier, CustomerTierGold)
	}
	if customer.OwnerID == nil || *customer.OwnerID != ownerID {
		t.Errorf("OwnerID = %v, want %v", customer.OwnerID, ownerID)
	}
	if len(customer.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(customer.Tags))
	}
	if customer.Notes != "Important customer" {
		t.Errorf("Notes = %s, want 'Important customer'", customer.Notes)
	}
}

func TestCustomerBuilder_WithAddress(t *testing.T) {
	tenantID := uuid.New()

	customer, err := NewCustomerBuilder(tenantID, "Test", CustomerTypeCompany).
		WithAddress("123 Main St", "Kuala Lumpur", "50000", "MY", AddressTypeBilling).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(customer.Addresses) != 1 {
		t.Errorf("Addresses count = %d, want 1", len(customer.Addresses))
	}
}

func TestCustomerBuilder_WithCompanyInfo(t *testing.T) {
	tenantID := uuid.New()
	companyInfo := CompanyInfo{
		LegalName:  "Test Corp Sdn Bhd",
		TaxID:      "123456789",
		Industry:   IndustryTechnology,
		Size:       CompanySizeMedium,
	}

	customer, err := NewCustomerBuilder(tenantID, "Test Corp", CustomerTypeCompany).
		WithCompanyInfo(companyInfo).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if customer.CompanyInfo == nil {
		t.Fatal("CompanyInfo should be set")
	}
	if customer.CompanyInfo.LegalName != "Test Corp Sdn Bhd" {
		t.Errorf("LegalName = %s", customer.CompanyInfo.LegalName)
	}
}

func TestCustomerBuilder_WithPreferences(t *testing.T) {
	tenantID := uuid.New()
	prefs := CustomerPreferences{
		Language:       "ms",
		Timezone:       "Asia/Kuala_Lumpur",
		Currency:       CurrencyMYR,
		CommPreference: CommPrefWhatsApp,
	}

	customer, err := NewCustomerBuilder(tenantID, "Test", CustomerTypeIndividual).
		WithPreferences(prefs).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if customer.Preferences.Language != "ms" {
		t.Errorf("Language = %s, want 'ms'", customer.Preferences.Language)
	}
}

func TestCustomerBuilder_WithCustomField(t *testing.T) {
	tenantID := uuid.New()

	customer, err := NewCustomerBuilder(tenantID, "Test", CustomerTypeIndividual).
		WithCustomField("account_manager", "John").
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if customer.CustomFields["account_manager"] != "John" {
		t.Errorf("CustomField = %v", customer.CustomFields["account_manager"])
	}
}

func TestCustomerBuilder_CodeTooLong(t *testing.T) {
	tenantID := uuid.New()
	longCode := ""
	for i := 0; i < 60; i++ {
		longCode += "a"
	}

	_, err := NewCustomerBuilder(tenantID, "Test", CustomerTypeIndividual).
		WithCode(longCode).
		Build()

	if err == nil {
		t.Error("Build() should return error for code too long")
	}
}

// ============================================================================
// Customer Behavior Tests
// ============================================================================

func createTestCustomer(t *testing.T) *Customer {
	t.Helper()

	customer, err := NewCustomerBuilder(uuid.New(), "Test Company", CustomerTypeCompany).
		WithEmail("contact@test.com").
		WithPhone("+60123456789", PhoneTypeMobile, true).
		Build()

	if err != nil {
		t.Fatalf("Failed to create test customer: %v", err)
	}

	customer.ClearDomainEvents()
	return customer
}

func TestCustomer_UpdateBasicInfo(t *testing.T) {
	customer := createTestCustomer(t)

	err := customer.UpdateBasicInfo("New Name", CustomerTypePartner)

	if err != nil {
		t.Fatalf("UpdateBasicInfo() error = %v", err)
	}
	if customer.Name != "New Name" {
		t.Errorf("Name = %s, want 'New Name'", customer.Name)
	}
	if customer.Type != CustomerTypePartner {
		t.Errorf("Type = %v, want %v", customer.Type, CustomerTypePartner)
	}

	events := customer.GetDomainEvents()
	if len(events) == 0 {
		t.Error("Expected domain event")
	}
}

func TestCustomer_UpdateBasicInfo_EmptyName(t *testing.T) {
	customer := createTestCustomer(t)

	err := customer.UpdateBasicInfo("", CustomerTypeIndividual)

	if err == nil {
		t.Error("UpdateBasicInfo() should return error for empty name")
	}
}

func TestCustomer_UpdateEmail(t *testing.T) {
	customer := createTestCustomer(t)
	newEmail, _ := NewEmail("new@example.com")

	customer.UpdateEmail(newEmail)

	if customer.Email.String() != "new@example.com" {
		t.Errorf("Email = %s, want 'new@example.com'", customer.Email.String())
	}

	events := customer.GetDomainEvents()
	if len(events) == 0 {
		t.Error("Expected domain event")
	}
}

func TestCustomer_AddPhone(t *testing.T) {
	customer := createTestCustomer(t)
	initialCount := len(customer.PhoneNumbers)
	newPhone, _ := NewPhoneNumber("+60198765432", PhoneTypeWork)

	customer.AddPhone(newPhone)

	if len(customer.PhoneNumbers) != initialCount+1 {
		t.Errorf("PhoneNumbers count = %d, want %d", len(customer.PhoneNumbers), initialCount+1)
	}
}

func TestCustomer_RemovePhone(t *testing.T) {
	customer := createTestCustomer(t)
	if len(customer.PhoneNumbers) == 0 {
		t.Skip("No phone numbers")
	}

	e164 := customer.PhoneNumbers[0].E164()
	initialCount := len(customer.PhoneNumbers)

	result := customer.RemovePhone(e164)

	if !result {
		t.Error("RemovePhone() should return true")
	}
	if len(customer.PhoneNumbers) != initialCount-1 {
		t.Errorf("PhoneNumbers count = %d", len(customer.PhoneNumbers))
	}
}

func TestCustomer_GetPrimaryPhone(t *testing.T) {
	customer := createTestCustomer(t)
	if len(customer.PhoneNumbers) > 0 {
		customer.PhoneNumbers[0].SetPrimary(true)
	}

	primary := customer.GetPrimaryPhone()

	if len(customer.PhoneNumbers) > 0 && primary == nil {
		t.Error("GetPrimaryPhone() should return a phone")
	}
}

func TestCustomer_AddAddress(t *testing.T) {
	customer := createTestCustomer(t)
	addr, _ := NewAddress("123 Main St", "KL", "50000", "MY", AddressTypeBilling)

	customer.AddAddress(addr)

	if len(customer.Addresses) != 1 {
		t.Errorf("Addresses count = %d, want 1", len(customer.Addresses))
	}
}

func TestCustomer_GetBillingAddress(t *testing.T) {
	customer := createTestCustomer(t)
	addr, _ := NewAddress("123 Main St", "KL", "50000", "MY", AddressTypeBilling)
	addr.IsPrimary = true
	customer.AddAddress(addr)

	billing := customer.GetBillingAddress()

	if billing == nil {
		t.Fatal("GetBillingAddress() should return address")
	}
	if billing.City != "KL" {
		t.Errorf("City = %s, want 'KL'", billing.City)
	}
}

func TestCustomer_GetShippingAddress(t *testing.T) {
	customer := createTestCustomer(t)
	addr, _ := NewAddress("456 Ship St", "KL", "50001", "MY", AddressTypeShipping)
	customer.AddAddress(addr)

	shipping := customer.GetShippingAddress()

	if shipping == nil {
		t.Fatal("GetShippingAddress() should return address")
	}
	if shipping.Line1 != "456 Ship St" {
		t.Errorf("Line1 = %s", shipping.Line1)
	}
}

// ============================================================================
// Contact Management Tests
// ============================================================================

func TestCustomer_AddContact(t *testing.T) {
	customer := createTestCustomer(t)
	contact, _ := NewContact(customer.ID, customer.TenantID, "John", "Doe", "john@example.com")

	err := customer.AddContact(contact)

	if err != nil {
		t.Fatalf("AddContact() error = %v", err)
	}
	if len(customer.Contacts) != 1 {
		t.Errorf("Contacts count = %d, want 1", len(customer.Contacts))
	}
	if customer.Stats.ContactCount != 1 {
		t.Errorf("ContactCount = %d, want 1", customer.Stats.ContactCount)
	}
	// First contact should be primary
	if !customer.Contacts[0].IsPrimary {
		t.Error("First contact should be primary")
	}
}

func TestCustomer_AddContact_MaxExceeded(t *testing.T) {
	customer := createTestCustomer(t)

	// Add max contacts
	for i := 0; i < MaxContacts; i++ {
		contact, err := NewContact(customer.ID, customer.TenantID, "Contact", "Name", fmt.Sprintf("contact%d@example.com", i))
		if err != nil {
			t.Fatalf("Failed to create contact %d: %v", i, err)
		}
		customer.AddContact(contact)
	}

	// Try to add one more
	extraContact, err := NewContact(customer.ID, customer.TenantID, "Extra", "Contact", "extra@example.com")
	if err != nil {
		t.Fatalf("Failed to create extra contact: %v", err)
	}
	err = customer.AddContact(extraContact)

	if err != ErrMaxContactsExceeded {
		t.Errorf("AddContact() error = %v, want ErrMaxContactsExceeded", err)
	}
}

func TestCustomer_AddContact_DuplicateEmail(t *testing.T) {
	customer := createTestCustomer(t)
	contact1, _ := NewContact(customer.ID, customer.TenantID, "John", "Doe", "john@example.com")
	contact2, _ := NewContact(customer.ID, customer.TenantID, "Jane", "Doe", "john@example.com")

	customer.AddContact(contact1)
	err := customer.AddContact(contact2)

	if err != ErrDuplicateContactEmail {
		t.Errorf("AddContact() error = %v, want ErrDuplicateContactEmail", err)
	}
}

func TestCustomer_UpdateContact(t *testing.T) {
	customer := createTestCustomer(t)
	contact, _ := NewContact(customer.ID, customer.TenantID, "John", "Doe", "john@example.com")
	customer.AddContact(contact)

	err := customer.UpdateContact(contact.ID, func(c *Contact) error {
		c.JobTitle = "Manager"
		return nil
	})

	if err != nil {
		t.Fatalf("UpdateContact() error = %v", err)
	}
	if customer.Contacts[0].JobTitle != "Manager" {
		t.Errorf("JobTitle = %s, want 'Manager'", customer.Contacts[0].JobTitle)
	}
}

func TestCustomer_UpdateContact_NotFound(t *testing.T) {
	customer := createTestCustomer(t)

	err := customer.UpdateContact(uuid.New(), func(c *Contact) error {
		return nil
	})

	if err != ErrContactNotFound {
		t.Errorf("UpdateContact() error = %v, want ErrContactNotFound", err)
	}
}

func TestCustomer_RemoveContact(t *testing.T) {
	customer := createTestCustomer(t)
	contact, _ := NewContact(customer.ID, customer.TenantID, "John", "Doe", "john@example.com")
	customer.AddContact(contact)

	err := customer.RemoveContact(contact.ID)

	if err != nil {
		t.Fatalf("RemoveContact() error = %v", err)
	}
	if len(customer.Contacts) != 0 {
		t.Errorf("Contacts count = %d, want 0", len(customer.Contacts))
	}
}

func TestCustomer_RemoveContact_Primary(t *testing.T) {
	customer := createTestCustomer(t)
	contact1, _ := NewContact(customer.ID, customer.TenantID, "John", "Doe", "john@example.com")
	contact2, _ := NewContact(customer.ID, customer.TenantID, "Jane", "Doe", "jane@example.com")
	customer.AddContact(contact1)
	customer.AddContact(contact2)

	// contact1 is primary, try to remove it
	err := customer.RemoveContact(contact1.ID)

	if err != ErrCannotDeletePrimaryContact {
		t.Errorf("RemoveContact() error = %v, want ErrCannotDeletePrimaryContact", err)
	}
}

func TestCustomer_GetContact(t *testing.T) {
	customer := createTestCustomer(t)
	contact, _ := NewContact(customer.ID, customer.TenantID, "John", "Doe", "john@example.com")
	customer.AddContact(contact)

	found := customer.GetContact(contact.ID)
	missing := customer.GetContact(uuid.New())

	if found == nil {
		t.Error("GetContact() should return contact")
	}
	if missing != nil {
		t.Error("GetContact() should return nil for missing ID")
	}
}

func TestCustomer_GetPrimaryContact(t *testing.T) {
	customer := createTestCustomer(t)
	contact, _ := NewContact(customer.ID, customer.TenantID, "John", "Doe", "john@example.com")
	customer.AddContact(contact)

	primary := customer.GetPrimaryContact()

	if primary == nil {
		t.Fatal("GetPrimaryContact() should return contact")
	}
	if !primary.IsPrimary {
		t.Error("Contact should be primary")
	}
}

func TestCustomer_SetPrimaryContact(t *testing.T) {
	customer := createTestCustomer(t)
	contact1, _ := NewContact(customer.ID, customer.TenantID, "John", "Doe", "john@example.com")
	contact2, _ := NewContact(customer.ID, customer.TenantID, "Jane", "Doe", "jane@example.com")
	customer.AddContact(contact1)
	customer.AddContact(contact2)

	err := customer.SetPrimaryContact(contact2.ID)

	if err != nil {
		t.Fatalf("SetPrimaryContact() error = %v", err)
	}

	for _, c := range customer.Contacts {
		if c.ID == contact2.ID && !c.IsPrimary {
			t.Error("Contact2 should be primary")
		}
		if c.ID == contact1.ID && c.IsPrimary {
			t.Error("Contact1 should not be primary")
		}
	}
}

func TestCustomer_SetPrimaryContact_NotFound(t *testing.T) {
	customer := createTestCustomer(t)

	err := customer.SetPrimaryContact(uuid.New())

	if err != ErrContactNotFound {
		t.Errorf("SetPrimaryContact() error = %v, want ErrContactNotFound", err)
	}
}

func TestCustomer_GetActiveContacts(t *testing.T) {
	customer := createTestCustomer(t)
	contact1, _ := NewContact(customer.ID, customer.TenantID, "Active", "Contact", "active@example.com")
	contact2, _ := NewContact(customer.ID, customer.TenantID, "Inactive", "Contact", "inactive@example.com")
	customer.AddContact(contact1)
	customer.AddContact(contact2)
	customer.Contacts[1].Status = ContactStatusInactive

	active := customer.GetActiveContacts()

	if len(active) != 1 {
		t.Errorf("Active contacts count = %d, want 1", len(active))
	}
}

// ============================================================================
// Status Management Tests
// ============================================================================

func TestCustomer_Activate(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusInactive

	err := customer.Activate()

	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if customer.Status != CustomerStatusActive {
		t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusActive)
	}
}

func TestCustomer_Activate_Blocked(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusBlocked

	err := customer.Activate()

	if err != ErrCustomerBlocked {
		t.Errorf("Activate() error = %v, want ErrCustomerBlocked", err)
	}
}

func TestCustomer_ConvertToCustomer(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusLead

	err := customer.ConvertToCustomer()

	if err != nil {
		t.Fatalf("ConvertToCustomer() error = %v", err)
	}
	if customer.Status != CustomerStatusActive {
		t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusActive)
	}
	if customer.ConvertedAt == nil {
		t.Error("ConvertedAt should be set")
	}
}

func TestCustomer_ConvertToCustomer_InvalidStatus(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusActive

	err := customer.ConvertToCustomer()

	if err == nil {
		t.Error("ConvertToCustomer() should return error for non-lead/prospect")
	}
}

func TestCustomer_Deactivate(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusActive

	customer.Deactivate()

	if customer.Status != CustomerStatusInactive {
		t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusInactive)
	}
}

func TestCustomer_MarkAsChurned(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusActive

	customer.MarkAsChurned("Competitor pricing")

	if customer.Status != CustomerStatusChurned {
		t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusChurned)
	}
	if customer.ChurnReason != "Competitor pricing" {
		t.Errorf("ChurnReason = %s", customer.ChurnReason)
	}
	if customer.ChurnedAt == nil {
		t.Error("ChurnedAt should be set")
	}
}

func TestCustomer_Block(t *testing.T) {
	customer := createTestCustomer(t)

	customer.Block("Payment fraud")

	if customer.Status != CustomerStatusBlocked {
		t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusBlocked)
	}
	if customer.Notes != "Payment fraud" {
		t.Errorf("Notes = %s", customer.Notes)
	}
}

func TestCustomer_Unblock(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusBlocked

	customer.Unblock()

	if customer.Status != CustomerStatusInactive {
		t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusInactive)
	}
}

func TestCustomer_PromoteToProspect(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusLead

	err := customer.PromoteToProspect()

	if err != nil {
		t.Fatalf("PromoteToProspect() error = %v", err)
	}
	if customer.Status != CustomerStatusProspect {
		t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusProspect)
	}
}

func TestCustomer_PromoteToProspect_InvalidStatus(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusActive

	err := customer.PromoteToProspect()

	if err == nil {
		t.Error("PromoteToProspect() should return error for non-lead")
	}
}

// ============================================================================
// Tier Management Tests
// ============================================================================

func TestCustomer_UpgradeTier(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Tier = CustomerTierStandard

	customer.UpgradeTier(CustomerTierGold)

	if customer.Tier != CustomerTierGold {
		t.Errorf("Tier = %v, want %v", customer.Tier, CustomerTierGold)
	}

	events := customer.GetDomainEvents()
	if len(events) == 0 {
		t.Error("Expected domain event for tier change")
	}
}

func TestCustomer_DowngradeTier(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Tier = CustomerTierGold

	customer.DowngradeTier(CustomerTierBronze)

	if customer.Tier != CustomerTierBronze {
		t.Errorf("Tier = %v, want %v", customer.Tier, CustomerTierBronze)
	}
}

// ============================================================================
// Owner and Assignment Tests
// ============================================================================

func TestCustomer_AssignOwner(t *testing.T) {
	customer := createTestCustomer(t)
	ownerID := uuid.New()

	customer.AssignOwner(ownerID)

	if customer.OwnerID == nil || *customer.OwnerID != ownerID {
		t.Errorf("OwnerID = %v, want %v", customer.OwnerID, ownerID)
	}
}

func TestCustomer_UnassignOwner(t *testing.T) {
	customer := createTestCustomer(t)
	ownerID := uuid.New()
	customer.OwnerID = &ownerID

	customer.UnassignOwner()

	if customer.OwnerID != nil {
		t.Error("OwnerID should be nil")
	}
}

func TestCustomer_AddToTeam(t *testing.T) {
	customer := createTestCustomer(t)
	userID := uuid.New()

	customer.AddToTeam(userID)

	if len(customer.AssignedTeam) != 1 {
		t.Errorf("AssignedTeam count = %d, want 1", len(customer.AssignedTeam))
	}
	if customer.AssignedTeam[0] != userID {
		t.Errorf("AssignedTeam[0] = %v, want %v", customer.AssignedTeam[0], userID)
	}
}

func TestCustomer_AddToTeam_Duplicate(t *testing.T) {
	customer := createTestCustomer(t)
	userID := uuid.New()

	customer.AddToTeam(userID)
	customer.AddToTeam(userID)

	if len(customer.AssignedTeam) != 1 {
		t.Errorf("AssignedTeam count = %d, want 1 (no duplicates)", len(customer.AssignedTeam))
	}
}

func TestCustomer_RemoveFromTeam(t *testing.T) {
	customer := createTestCustomer(t)
	userID := uuid.New()
	customer.AddToTeam(userID)

	customer.RemoveFromTeam(userID)

	if len(customer.AssignedTeam) != 0 {
		t.Errorf("AssignedTeam count = %d, want 0", len(customer.AssignedTeam))
	}
}

func TestCustomer_IsAssignedTo(t *testing.T) {
	customer := createTestCustomer(t)
	ownerID := uuid.New()
	teamMemberID := uuid.New()
	randomID := uuid.New()

	customer.AssignOwner(ownerID)
	customer.AddToTeam(teamMemberID)

	if !customer.IsAssignedTo(ownerID) {
		t.Error("IsAssignedTo() should return true for owner")
	}
	if !customer.IsAssignedTo(teamMemberID) {
		t.Error("IsAssignedTo() should return true for team member")
	}
	if customer.IsAssignedTo(randomID) {
		t.Error("IsAssignedTo() should return false for unassigned user")
	}
}

// ============================================================================
// Tags and Segments Tests
// ============================================================================

func TestCustomer_AddTag(t *testing.T) {
	customer := createTestCustomer(t)

	customer.AddTag("enterprise")

	if len(customer.Tags) != 1 {
		t.Errorf("Tags count = %d, want 1", len(customer.Tags))
	}
	if !customer.HasTag("enterprise") {
		t.Error("HasTag() should return true")
	}
}

func TestCustomer_AddTag_Duplicate(t *testing.T) {
	customer := createTestCustomer(t)

	customer.AddTag("vip")
	customer.AddTag("vip")

	if len(customer.Tags) != 1 {
		t.Errorf("Tags count = %d, want 1 (no duplicates)", len(customer.Tags))
	}
}

func TestCustomer_RemoveTag(t *testing.T) {
	customer := createTestCustomer(t)
	customer.AddTag("vip")

	customer.RemoveTag("vip")

	if len(customer.Tags) != 0 {
		t.Errorf("Tags count = %d, want 0", len(customer.Tags))
	}
}

func TestCustomer_AddToSegment(t *testing.T) {
	customer := createTestCustomer(t)
	segmentID := uuid.New()

	customer.AddToSegment(segmentID)

	if len(customer.Segments) != 1 {
		t.Errorf("Segments count = %d, want 1", len(customer.Segments))
	}
}

func TestCustomer_RemoveFromSegment(t *testing.T) {
	customer := createTestCustomer(t)
	segmentID := uuid.New()
	customer.AddToSegment(segmentID)

	customer.RemoveFromSegment(segmentID)

	if len(customer.Segments) != 0 {
		t.Errorf("Segments count = %d, want 0", len(customer.Segments))
	}
}

// ============================================================================
// Financial Tests
// ============================================================================

func TestCustomer_UpdateFinancials(t *testing.T) {
	customer := createTestCustomer(t)
	financials := CustomerFinancials{
		Currency:     CurrencyUSD,
		PaymentTerms: 60,
		TaxExempt:    true,
	}

	customer.UpdateFinancials(financials)

	if customer.Financials.Currency != CurrencyUSD {
		t.Errorf("Currency = %v, want %v", customer.Financials.Currency, CurrencyUSD)
	}
	if customer.Financials.PaymentTerms != 60 {
		t.Errorf("PaymentTerms = %d, want 60", customer.Financials.PaymentTerms)
	}
	if !customer.Financials.TaxExempt {
		t.Error("TaxExempt should be true")
	}
}

func TestCustomer_RecordPurchase(t *testing.T) {
	customer := createTestCustomer(t)
	amount, _ := NewMoney(10000, CurrencyMYR)

	customer.RecordPurchase(amount)

	if customer.Financials.TotalPurchases != 1 {
		t.Errorf("TotalPurchases = %d, want 1", customer.Financials.TotalPurchases)
	}
	if customer.Financials.LastPurchaseAt == nil {
		t.Error("LastPurchaseAt should be set")
	}
	if customer.Financials.TotalSpent == nil {
		t.Error("TotalSpent should be set")
	}
}

func TestCustomer_RecordPayment(t *testing.T) {
	customer := createTestCustomer(t)
	balance, _ := NewMoney(50000, CurrencyMYR)
	customer.Financials.CurrentBalance = &balance

	payment, _ := NewMoney(20000, CurrencyMYR)
	customer.RecordPayment(payment)

	if customer.Financials.LastPaymentAt == nil {
		t.Error("LastPaymentAt should be set")
	}
	if customer.Financials.CurrentBalance.Amount != 30000 {
		t.Errorf("CurrentBalance = %d, want 30000", customer.Financials.CurrentBalance.Amount)
	}
}

// ============================================================================
// Activity Tracking Tests
// ============================================================================

func TestCustomer_RecordContact(t *testing.T) {
	customer := createTestCustomer(t)

	customer.RecordContact()

	if customer.LastContactedAt == nil {
		t.Error("LastContactedAt should be set")
	}
}

func TestCustomer_SetNextFollowUp(t *testing.T) {
	customer := createTestCustomer(t)
	followUp := time.Now().Add(24 * time.Hour)

	customer.SetNextFollowUp(followUp)

	if customer.NextFollowUpAt == nil {
		t.Fatal("NextFollowUpAt should be set")
	}
}

func TestCustomer_ClearNextFollowUp(t *testing.T) {
	customer := createTestCustomer(t)
	now := time.Now()
	customer.NextFollowUpAt = &now

	customer.ClearNextFollowUp()

	if customer.NextFollowUpAt != nil {
		t.Error("NextFollowUpAt should be nil")
	}
}

func TestCustomer_UpdateEngagementScore(t *testing.T) {
	customer := createTestCustomer(t)

	customer.UpdateEngagementScore(75)

	if customer.Stats.EngagementScore != 75 {
		t.Errorf("EngagementScore = %d, want 75", customer.Stats.EngagementScore)
	}
}

func TestCustomer_UpdateHealthScore(t *testing.T) {
	customer := createTestCustomer(t)

	customer.UpdateHealthScore(85)

	if customer.Stats.HealthScore != 85 {
		t.Errorf("HealthScore = %d, want 85", customer.Stats.HealthScore)
	}
}

// ============================================================================
// Query Methods Tests
// ============================================================================

func TestCustomer_IsActive(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusActive

	if !customer.IsActive() {
		t.Error("IsActive() should return true")
	}

	customer.Status = CustomerStatusChurned
	if customer.IsActive() {
		t.Error("IsActive() should return false")
	}
}

func TestCustomer_IsLead(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusLead

	if !customer.IsLead() {
		t.Error("IsLead() should return true")
	}
}

func TestCustomer_IsProspect(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusProspect

	if !customer.IsProspect() {
		t.Error("IsProspect() should return true")
	}
}

func TestCustomer_IsChurned(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusChurned

	if !customer.IsChurned() {
		t.Error("IsChurned() should return true")
	}
}

func TestCustomer_IsBlocked(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusBlocked

	if !customer.IsBlocked() {
		t.Error("IsBlocked() should return true")
	}
}

func TestCustomer_IsCompany(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Type = CustomerTypeCompany

	if !customer.IsCompany() {
		t.Error("IsCompany() should return true")
	}

	customer.Type = CustomerTypeIndividual
	if customer.IsCompany() {
		t.Error("IsCompany() should return false")
	}
}

func TestCustomer_CanReceiveMarketing(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Status = CustomerStatusActive
	customer.Preferences.OptedOutMarketing = false

	if !customer.CanReceiveMarketing() {
		t.Error("CanReceiveMarketing() should return true")
	}

	customer.Preferences.OptedOutMarketing = true
	if customer.CanReceiveMarketing() {
		t.Error("CanReceiveMarketing() should return false when opted out")
	}
}

func TestCustomer_NeedsFollowUp(t *testing.T) {
	customer := createTestCustomer(t)

	// No follow-up set
	if customer.NeedsFollowUp() {
		t.Error("NeedsFollowUp() should return false when no follow-up")
	}

	// Past follow-up
	pastTime := time.Now().Add(-time.Hour)
	customer.NextFollowUpAt = &pastTime
	if !customer.NeedsFollowUp() {
		t.Error("NeedsFollowUp() should return true for past follow-up")
	}

	// Future follow-up
	futureTime := time.Now().Add(time.Hour)
	customer.NextFollowUpAt = &futureTime
	if customer.NeedsFollowUp() {
		t.Error("NeedsFollowUp() should return false for future follow-up")
	}
}

func TestCustomer_DaysSinceLastContact(t *testing.T) {
	customer := createTestCustomer(t)

	// No last contact
	days := customer.DaysSinceLastContact()
	if days != -1 {
		t.Errorf("DaysSinceLastContact() = %d, want -1 when no contact", days)
	}

	// 2 days ago
	pastTime := time.Now().Add(-48 * time.Hour)
	customer.LastContactedAt = &pastTime
	days = customer.DaysSinceLastContact()
	if days < 1 || days > 3 {
		t.Errorf("DaysSinceLastContact() = %d, expected ~2", days)
	}
}

// ============================================================================
// Delete/Restore Tests
// ============================================================================

func TestCustomer_Delete(t *testing.T) {
	customer := createTestCustomer(t)
	deletedBy := uuid.New()

	customer.Delete(deletedBy)

	if !customer.IsDeleted() {
		t.Error("IsDeleted() should return true")
	}
	if customer.AuditInfo.DeletedBy == nil || *customer.AuditInfo.DeletedBy != deletedBy {
		t.Errorf("DeletedBy = %v, want %v", customer.AuditInfo.DeletedBy, deletedBy)
	}
}

func TestCustomer_Restore(t *testing.T) {
	customer := createTestCustomer(t)
	customer.Delete(uuid.New())

	customer.Restore()

	if customer.IsDeleted() {
		t.Error("IsDeleted() should return false after restore")
	}
	if customer.Status != CustomerStatusInactive {
		t.Errorf("Status = %v, want %v", customer.Status, CustomerStatusInactive)
	}
}

// ============================================================================
// Customer Code Generation Tests
// ============================================================================

func TestGenerateCustomerCode(t *testing.T) {
	tests := []struct {
		customerType CustomerType
		prefix       string
	}{
		{CustomerTypeIndividual, "IN"},
		{CustomerTypeCompany, "CO"},
		{CustomerTypePartner, "PA"},
		{CustomerTypeReseller, "RS"},
	}

	for _, tt := range tests {
		t.Run(string(tt.customerType), func(t *testing.T) {
			code := generateCustomerCode(tt.customerType)

			if len(code) < 3 {
				t.Errorf("Code too short: %s", code)
			}
			if code[:2] != tt.prefix {
				t.Errorf("Code prefix = %s, want %s", code[:2], tt.prefix)
			}
		})
	}
}
