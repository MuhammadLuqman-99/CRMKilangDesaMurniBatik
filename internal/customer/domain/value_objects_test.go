// Package domain contains the domain layer for the Customer service.
package domain

import (
	"testing"
)

// ============================================================================
// Email Tests
// ============================================================================

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email uppercase", "TEST@EXAMPLE.COM", false},
		{"valid email with subdomain", "test@mail.example.com", false},
		{"valid email with plus", "test+alias@example.com", false},
		{"empty email", "", true},
		{"invalid format", "notanemail", true},
		{"missing domain", "test@", true},
		{"missing local", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.address)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEmail(%q) expected error, got nil", tt.address)
				}
				return
			}

			if err != nil {
				t.Errorf("NewEmail(%q) unexpected error: %v", tt.address, err)
				return
			}

			if email.String() == "" {
				t.Error("Email string should not be empty")
			}
		})
	}
}

func TestEmail_IsDisposable(t *testing.T) {
	disposable, _ := NewEmail("test@tempmail.com")
	if !disposable.IsDisposable() {
		t.Error("tempmail.com should be disposable")
	}

	regular, _ := NewEmail("test@gmail.com")
	if regular.IsDisposable() {
		t.Error("gmail.com should not be disposable")
	}
}

func TestEmail_Parts(t *testing.T) {
	email, _ := NewEmail("Test@Example.com")

	if email.Local() != "Test" {
		t.Errorf("Local() = %s, want Test", email.Local())
	}
	if email.Domain() != "example.com" {
		t.Errorf("Domain() = %s, want example.com", email.Domain())
	}
	if email.Normalized() != "test@example.com" {
		t.Errorf("Normalized() = %s, want test@example.com", email.Normalized())
	}
}

func TestEmail_Equals(t *testing.T) {
	e1, _ := NewEmail("test@example.com")
	e2, _ := NewEmail("TEST@EXAMPLE.COM")
	e3, _ := NewEmail("other@example.com")

	if !e1.Equals(e2) {
		t.Error("Same emails (case-insensitive) should be equal")
	}
	if e1.Equals(e3) {
		t.Error("Different emails should not be equal")
	}
}

// ============================================================================
// PhoneNumber Tests
// ============================================================================

func TestNewPhoneNumber(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		phoneType PhoneType
		wantErr   bool
	}{
		{"valid US number", "+1234567890", PhoneTypeMobile, false},
		{"valid number with plus", "+60123456789", PhoneTypeMobile, false},
		{"valid number without plus", "60123456789", PhoneTypeMobile, false},
		{"number with extension", "+1234567890 ext 123", PhoneTypeWork, false},
		{"empty number", "", PhoneTypeMobile, true},
		{"invalid number", "abc", PhoneTypeMobile, true},
		{"starts with zero", "+0123456789", PhoneTypeMobile, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone, err := NewPhoneNumber(tt.raw, tt.phoneType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPhoneNumber(%q) expected error, got nil", tt.raw)
				}
				return
			}

			if err != nil {
				t.Errorf("NewPhoneNumber(%q) unexpected error: %v", tt.raw, err)
				return
			}

			if phone.IsEmpty() {
				t.Error("Phone should not be empty")
			}
		})
	}
}

func TestPhoneNumber_E164(t *testing.T) {
	phone, _ := NewPhoneNumber("60123456789", PhoneTypeMobile)
	e164 := phone.E164()

	if e164 == "" {
		t.Error("E164() should not be empty")
	}
	if e164[0] != '+' {
		t.Error("E164 should start with +")
	}
}

func TestPhoneNumber_Type(t *testing.T) {
	phone, _ := NewPhoneNumber("+60123456789", PhoneTypeWork)
	if phone.Type() != PhoneTypeWork {
		t.Errorf("Type() = %s, want work", phone.Type())
	}
}

func TestPhoneNumber_Primary(t *testing.T) {
	phone, _ := NewPhoneNumberWithPrimary("+60123456789", PhoneTypeMobile, true)
	if !phone.IsPrimary() {
		t.Error("IsPrimary() should be true")
	}

	phone.SetPrimary(false)
	if phone.IsPrimary() {
		t.Error("IsPrimary() should be false after SetPrimary(false)")
	}
}

func TestPhoneNumber_Equals(t *testing.T) {
	p1, _ := NewPhoneNumber("+60123456789", PhoneTypeMobile)
	p2, _ := NewPhoneNumber("+60123456789", PhoneTypeWork)
	p3, _ := NewPhoneNumber("+60198765432", PhoneTypeMobile)

	if !p1.Equals(p2) {
		t.Error("Same E164 numbers should be equal regardless of type")
	}
	if p1.Equals(p3) {
		t.Error("Different numbers should not be equal")
	}
}

// ============================================================================
// Address Tests
// ============================================================================

func TestNewAddress(t *testing.T) {
	tests := []struct {
		name        string
		line1       string
		city        string
		postalCode  string
		countryCode string
		addressType AddressType
		wantErr     bool
	}{
		{"valid MY address", "123 Jalan Test", "Kuala Lumpur", "50000", "MY", AddressTypeBilling, false},
		{"valid SG address", "123 Test Street", "Singapore", "123456", "SG", AddressTypeShipping, false},
		{"valid US address", "123 Main St", "New York", "10001", "US", AddressTypeOffice, false},
		{"empty line1", "", "City", "12345", "MY", AddressTypeBilling, true},
		{"empty city", "123 Test", "", "12345", "MY", AddressTypeBilling, true},
		{"empty postal", "123 Test", "City", "", "MY", AddressTypeBilling, true},
		{"invalid country", "123 Test", "City", "12345", "XX", AddressTypeBilling, true},
		{"invalid MY postal", "123 Test", "City", "123", "MY", AddressTypeBilling, true},
		{"invalid SG postal", "123 Test", "City", "12345", "SG", AddressTypeBilling, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := NewAddress(tt.line1, tt.city, tt.postalCode, tt.countryCode, tt.addressType)

			if tt.wantErr {
				if err == nil {
					t.Error("NewAddress() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewAddress() unexpected error: %v", err)
				return
			}

			if address.IsEmpty() {
				t.Error("Address should not be empty")
			}
		})
	}
}

func TestAddress_WithDetails(t *testing.T) {
	addr, _ := NewAddress("123 Main St", "City", "50000", "MY", AddressTypeBilling)
	addr = addr.WithDetails("Apt 4B", "Building A", "Selangor")

	if addr.Line2 != "Apt 4B" {
		t.Errorf("Line2 = %s, want Apt 4B", addr.Line2)
	}
	if addr.Line3 != "Building A" {
		t.Errorf("Line3 = %s, want Building A", addr.Line3)
	}
	if addr.State != "Selangor" {
		t.Errorf("State = %s, want Selangor", addr.State)
	}
}

func TestAddress_WithCoordinates(t *testing.T) {
	addr, _ := NewAddress("123 Main St", "City", "50000", "MY", AddressTypeBilling)
	addr = addr.WithCoordinates(3.1390, 101.6869)

	if !addr.HasCoordinates() {
		t.Error("HasCoordinates() should be true")
	}
	if *addr.Latitude != 3.1390 {
		t.Errorf("Latitude = %f, want 3.1390", *addr.Latitude)
	}
	if *addr.Longitude != 101.6869 {
		t.Errorf("Longitude = %f, want 101.6869", *addr.Longitude)
	}
}

func TestAddress_String(t *testing.T) {
	addr, _ := NewAddress("123 Main St", "Kuala Lumpur", "50000", "MY", AddressTypeBilling)
	str := addr.String()

	if str == "" {
		t.Error("String() should not be empty")
	}
}

func TestAddress_Equals(t *testing.T) {
	a1, _ := NewAddress("123 Main St", "City", "50000", "MY", AddressTypeBilling)
	a2, _ := NewAddress("123 Main St", "City", "50000", "MY", AddressTypeShipping)
	a3, _ := NewAddress("456 Other St", "City", "50000", "MY", AddressTypeBilling)

	if !a1.Equals(a2) {
		t.Error("Same address details should be equal regardless of type")
	}
	if a1.Equals(a3) {
		t.Error("Different addresses should not be equal")
	}
}

// ============================================================================
// SocialProfile Tests
// ============================================================================

func TestNewSocialProfile(t *testing.T) {
	tests := []struct {
		name     string
		platform SocialPlatform
		url      string
		wantErr  bool
	}{
		{"valid LinkedIn", SocialPlatformLinkedIn, "https://linkedin.com/in/johndoe", false},
		{"valid Twitter", SocialPlatformTwitter, "https://twitter.com/johndoe", false},
		{"valid X", SocialPlatformTwitter, "https://x.com/johndoe", false},
		{"valid Facebook", SocialPlatformFacebook, "https://facebook.com/johndoe", false},
		{"valid Instagram", SocialPlatformInstagram, "https://instagram.com/johndoe", false},
		{"valid GitHub", SocialPlatformGitHub, "https://github.com/johndoe", false},
		{"valid website", SocialPlatformWebsite, "https://example.com", false},
		{"empty url", SocialPlatformLinkedIn, "", true},
		{"invalid url", SocialPlatformLinkedIn, "not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := NewSocialProfile(tt.platform, tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("NewSocialProfile() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewSocialProfile() unexpected error: %v", err)
				return
			}

			if profile.IsEmpty() {
				t.Error("Profile should not be empty")
			}
		})
	}
}

func TestSocialProfile_Username(t *testing.T) {
	profile, _ := NewSocialProfile(SocialPlatformTwitter, "https://twitter.com/johndoe")
	if profile.Username != "johndoe" {
		t.Errorf("Username = %s, want johndoe", profile.Username)
	}
}

func TestSocialProfile_WithDisplayName(t *testing.T) {
	profile, _ := NewSocialProfile(SocialPlatformTwitter, "https://twitter.com/johndoe")
	profile = profile.WithDisplayName("John Doe")

	if profile.DisplayName != "John Doe" {
		t.Errorf("DisplayName = %s, want John Doe", profile.DisplayName)
	}
}

func TestSocialProfile_WithFollowers(t *testing.T) {
	profile, _ := NewSocialProfile(SocialPlatformTwitter, "https://twitter.com/johndoe")
	profile = profile.WithFollowers(1000)

	if profile.Followers == nil || *profile.Followers != 1000 {
		t.Error("Followers should be 1000")
	}
}

func TestSocialProfile_Equals(t *testing.T) {
	p1, _ := NewSocialProfile(SocialPlatformTwitter, "https://twitter.com/johndoe")
	p2, _ := NewSocialProfile(SocialPlatformTwitter, "https://twitter.com/johndoe")
	p3, _ := NewSocialProfile(SocialPlatformTwitter, "https://twitter.com/janedoe")

	if !p1.Equals(p2) {
		t.Error("Same profiles should be equal")
	}
	if p1.Equals(p3) {
		t.Error("Different profiles should not be equal")
	}
}

// ============================================================================
// Money Tests (Customer domain)
// ============================================================================

func TestNewMoney_Customer(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency Currency
		wantErr  bool
	}{
		{"valid MYR", 1000, CurrencyMYR, false},
		{"valid USD", 500, CurrencyUSD, false},
		{"invalid currency", 100, Currency("XXX"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := NewMoney(tt.amount, tt.currency)

			if tt.wantErr {
				if err == nil {
					t.Error("NewMoney() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewMoney() unexpected error: %v", err)
				return
			}

			if money.Amount != tt.amount {
				t.Errorf("Amount = %d, want %d", money.Amount, tt.amount)
			}
		})
	}
}

func TestMoney_Add(t *testing.T) {
	m1, _ := NewMoney(1000, CurrencyMYR)
	m2, _ := NewMoney(500, CurrencyMYR)
	m3, _ := NewMoney(100, CurrencyUSD)

	sum, err := m1.Add(m2)
	if err != nil {
		t.Errorf("Add() error: %v", err)
	}
	if sum.Amount != 1500 {
		t.Errorf("Add() = %d, want 1500", sum.Amount)
	}

	// Different currency
	_, err = m1.Add(m3)
	if err == nil {
		t.Error("Add() should error with different currencies")
	}
}

func TestMoney_Subtract(t *testing.T) {
	m1, _ := NewMoney(1000, CurrencyMYR)
	m2, _ := NewMoney(300, CurrencyMYR)

	diff, err := m1.Subtract(m2)
	if err != nil {
		t.Errorf("Subtract() error: %v", err)
	}
	if diff.Amount != 700 {
		t.Errorf("Subtract() = %d, want 700", diff.Amount)
	}
}

func TestMoney_Multiply(t *testing.T) {
	m, _ := NewMoney(1000, CurrencyMYR)
	result := m.Multiply(1.5)

	if result.Amount != 1500 {
		t.Errorf("Multiply() = %d, want 1500", result.Amount)
	}
}

func TestMoney_Float(t *testing.T) {
	m, _ := NewMoney(1050, CurrencyMYR)
	if m.Float() != 10.50 {
		t.Errorf("Float() = %f, want 10.50", m.Float())
	}
}

func TestMoney_String(t *testing.T) {
	m, _ := NewMoney(1050, CurrencyMYR)
	if m.String() != "MYR 10.50" {
		t.Errorf("String() = %s, want MYR 10.50", m.String())
	}
}

// ============================================================================
// Website Tests
// ============================================================================

func TestNewWebsite(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https", "https://example.com", false},
		{"valid http", "http://example.com", false},
		{"valid without scheme", "example.com", false},
		{"valid with path", "https://example.com/path", false},
		{"empty url", "", false}, // Empty is valid
		{"invalid no domain", "notadomain", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			website, err := NewWebsite(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("NewWebsite() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewWebsite() unexpected error: %v", err)
				return
			}

			// Empty input should return empty website
			if tt.url == "" && !website.IsEmpty() {
				t.Error("Empty url should return empty website")
			}
		})
	}
}

func TestWebsite_Domain(t *testing.T) {
	website, _ := NewWebsite("https://www.example.com/path")
	if website.Domain() != "www.example.com" {
		t.Errorf("Domain() = %s, want www.example.com", website.Domain())
	}
}

// ============================================================================
// PersonName Tests
// ============================================================================

func TestNewPersonName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		wantErr   bool
	}{
		{"full name", "John", "Doe", false},
		{"first name only", "John", "", false},
		{"last name only", "", "Doe", false},
		{"empty name", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := NewPersonName(tt.firstName, tt.lastName)

			if tt.wantErr {
				if err == nil {
					t.Error("NewPersonName() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewPersonName() unexpected error: %v", err)
			}

			if !tt.wantErr && name.IsEmpty() {
				t.Error("Name should not be empty")
			}
		})
	}
}

func TestPersonName_FullName(t *testing.T) {
	name, _ := NewPersonName("john", "doe")
	name = name.WithTitle("Mr.")
	name = name.WithMiddleName("william")
	name = name.WithSuffix("Jr.")

	fullName := name.FullName()
	if fullName != "Mr. John William Doe Jr." {
		t.Errorf("FullName() = %s, want Mr. John William Doe Jr.", fullName)
	}
}

func TestPersonName_DisplayName(t *testing.T) {
	name, _ := NewPersonName("John", "Doe")

	if name.DisplayName() != "John" {
		t.Errorf("DisplayName() = %s, want John", name.DisplayName())
	}

	name = name.WithNickname("Johnny")
	if name.DisplayName() != "Johnny" {
		t.Errorf("DisplayName() with nickname = %s, want Johnny", name.DisplayName())
	}
}

func TestPersonName_Initials(t *testing.T) {
	name, _ := NewPersonName("John", "Doe")
	if name.Initials() != "JD" {
		t.Errorf("Initials() = %s, want JD", name.Initials())
	}

	nameFirstOnly, _ := NewPersonName("John", "")
	if nameFirstOnly.Initials() != "J" {
		t.Errorf("Initials() first only = %s, want J", nameFirstOnly.Initials())
	}
}

func TestCapitalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"john", "John"},
		{"JOHN", "John"},
		{"john doe", "John Doe"},
		{"JOHN DOE", "John Doe"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := capitalizeName(tt.input)
			if result != tt.want {
				t.Errorf("capitalizeName(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}
