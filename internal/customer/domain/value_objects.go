// Package domain contains the domain layer for the Customer service.
package domain

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// ============================================================================
// Email Value Object
// ============================================================================

// Email represents a validated email address.
type Email struct {
	address    string
	local      string
	domain     string
	normalized string
}

// Common email regex pattern - comprehensive validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// Disposable email domains (common ones)
var disposableDomains = map[string]bool{
	"tempmail.com": true, "throwaway.com": true, "guerrillamail.com": true,
	"mailinator.com": true, "10minutemail.com": true, "temp-mail.org": true,
	"fakeinbox.com": true, "trashmail.com": true, "yopmail.com": true,
}

// NewEmail creates a new Email value object.
func NewEmail(address string) (Email, error) {
	address = strings.TrimSpace(address)

	if address == "" {
		return Email{}, ErrInvalidEmail
	}

	// Length check
	if len(address) > 254 {
		return Email{}, NewValidationError("email", "email address too long (max 254 characters)", "EMAIL_TOO_LONG")
	}

	// Format validation
	if !emailRegex.MatchString(address) {
		return Email{}, ErrInvalidEmail
	}

	// Split into local and domain parts
	parts := strings.SplitN(address, "@", 2)
	if len(parts) != 2 {
		return Email{}, ErrInvalidEmail
	}

	local := parts[0]
	domain := strings.ToLower(parts[1])

	// Local part length check
	if len(local) > 64 {
		return Email{}, NewValidationError("email", "local part too long (max 64 characters)", "EMAIL_LOCAL_TOO_LONG")
	}

	// Domain validation
	if len(domain) > 255 {
		return Email{}, NewValidationError("email", "domain too long (max 255 characters)", "EMAIL_DOMAIN_TOO_LONG")
	}

	// Normalize - lowercase the entire address
	normalized := strings.ToLower(address)

	return Email{
		address:    address,
		local:      local,
		domain:     domain,
		normalized: normalized,
	}, nil
}

// MustNewEmail creates a new Email, panics on error.
func MustNewEmail(address string) Email {
	email, err := NewEmail(address)
	if err != nil {
		panic(err)
	}
	return email
}

// String returns the original email address.
func (e Email) String() string {
	return e.address
}

// Normalized returns the normalized (lowercase) email.
func (e Email) Normalized() string {
	return e.normalized
}

// Local returns the local part (before @).
func (e Email) Local() string {
	return e.local
}

// Domain returns the domain part (after @).
func (e Email) Domain() string {
	return e.domain
}

// IsEmpty returns true if the email is empty.
func (e Email) IsEmpty() bool {
	return e.address == ""
}

// IsDisposable checks if the email is from a disposable domain.
func (e Email) IsDisposable() bool {
	return disposableDomains[e.domain]
}

// Equals checks if two emails are equal (case-insensitive).
func (e Email) Equals(other Email) bool {
	return e.normalized == other.normalized
}

// MarshalJSON implements json.Marshaler.
func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.address)
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Email) UnmarshalJSON(data []byte) error {
	var address string
	if err := json.Unmarshal(data, &address); err != nil {
		return err
	}
	email, err := NewEmail(address)
	if err != nil {
		return err
	}
	*e = email
	return nil
}

// ============================================================================
// PhoneNumber Value Object
// ============================================================================

// PhoneType represents the type of phone number.
type PhoneType string

const (
	PhoneTypeMobile   PhoneType = "mobile"
	PhoneTypeWork     PhoneType = "work"
	PhoneTypeHome     PhoneType = "home"
	PhoneTypeFax      PhoneType = "fax"
	PhoneTypeWhatsApp PhoneType = "whatsapp"
	PhoneTypeOther    PhoneType = "other"
)

// PhoneNumber represents a validated phone number.
type PhoneNumber struct {
	raw         string
	countryCode string
	number      string
	extension   string
	formatted   string
	e164        string
	phoneType   PhoneType
	isPrimary   bool
}

// Phone number regex patterns
var (
	digitsOnlyRegex = regexp.MustCompile(`[^\d+]`)
	phoneRegex      = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`) // E.164 format
)

// Country code to minimum/maximum length mapping
var countryPhoneLengths = map[string]struct{ min, max int }{
	"1":  {10, 11}, // US, Canada
	"44": {10, 11}, // UK
	"60": {9, 11},  // Malaysia
	"62": {9, 13},  // Indonesia
	"65": {8, 8},   // Singapore
	"81": {10, 11}, // Japan
	"86": {11, 11}, // China
	"91": {10, 10}, // India
}

// NewPhoneNumber creates a new PhoneNumber value object.
func NewPhoneNumber(raw string, phoneType PhoneType) (PhoneNumber, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return PhoneNumber{}, ErrInvalidPhoneNumber
	}

	// Extract extension if present
	var extension string
	if idx := strings.Index(strings.ToLower(raw), "ext"); idx != -1 {
		extension = strings.TrimSpace(digitsOnlyRegex.ReplaceAllString(raw[idx:], ""))
		raw = strings.TrimSpace(raw[:idx])
	} else if idx := strings.Index(raw, "x"); idx != -1 {
		extension = strings.TrimSpace(digitsOnlyRegex.ReplaceAllString(raw[idx:], ""))
		raw = strings.TrimSpace(raw[:idx])
	}

	// Clean the number - keep only digits and leading +
	hasPlus := strings.HasPrefix(raw, "+")
	cleaned := digitsOnlyRegex.ReplaceAllString(raw, "")

	if hasPlus {
		cleaned = "+" + cleaned
	}

	// Validate E.164 format (or close to it)
	if !phoneRegex.MatchString(cleaned) {
		return PhoneNumber{}, ErrInvalidPhoneNumber
	}

	// Extract country code (simplified - first 1-3 digits after +)
	e164 := cleaned
	if !hasPlus {
		e164 = "+" + cleaned
	}

	countryCode := ""
	number := strings.TrimPrefix(e164, "+")

	// Try to identify country code
	for code := range countryPhoneLengths {
		if strings.HasPrefix(number, code) {
			countryCode = code
			number = strings.TrimPrefix(number, code)
			break
		}
	}

	// Format for display
	formatted := formatPhoneNumber(e164)

	if phoneType == "" {
		phoneType = PhoneTypeOther
	}

	return PhoneNumber{
		raw:         raw,
		countryCode: countryCode,
		number:      number,
		extension:   extension,
		formatted:   formatted,
		e164:        e164,
		phoneType:   phoneType,
		isPrimary:   false,
	}, nil
}

// NewPhoneNumberWithPrimary creates a primary phone number.
func NewPhoneNumberWithPrimary(raw string, phoneType PhoneType, isPrimary bool) (PhoneNumber, error) {
	phone, err := NewPhoneNumber(raw, phoneType)
	if err != nil {
		return PhoneNumber{}, err
	}
	phone.isPrimary = isPrimary
	return phone, nil
}

func formatPhoneNumber(e164 string) string {
	// Simple formatting - can be enhanced with libphonenumber
	number := strings.TrimPrefix(e164, "+")

	switch {
	case strings.HasPrefix(number, "1") && len(number) == 11:
		// US/Canada: +1 (XXX) XXX-XXXX
		return fmt.Sprintf("+1 (%s) %s-%s", number[1:4], number[4:7], number[7:11])
	case strings.HasPrefix(number, "60") && len(number) >= 10:
		// Malaysia: +60 XX-XXX XXXX
		return fmt.Sprintf("+60 %s-%s %s", number[2:4], number[4:7], number[7:])
	default:
		return e164
	}
}

// String returns the formatted phone number.
func (p PhoneNumber) String() string {
	if p.extension != "" {
		return fmt.Sprintf("%s ext %s", p.formatted, p.extension)
	}
	return p.formatted
}

// E164 returns the E.164 formatted number.
func (p PhoneNumber) E164() string {
	return p.e164
}

// Raw returns the original input.
func (p PhoneNumber) Raw() string {
	return p.raw
}

// CountryCode returns the country code.
func (p PhoneNumber) CountryCode() string {
	return p.countryCode
}

// Extension returns the extension.
func (p PhoneNumber) Extension() string {
	return p.extension
}

// Type returns the phone type.
func (p PhoneNumber) Type() PhoneType {
	return p.phoneType
}

// IsPrimary returns true if this is the primary number.
func (p PhoneNumber) IsPrimary() bool {
	return p.isPrimary
}

// SetPrimary sets the primary flag.
func (p *PhoneNumber) SetPrimary(isPrimary bool) {
	p.isPrimary = isPrimary
}

// IsEmpty returns true if the phone number is empty.
func (p PhoneNumber) IsEmpty() bool {
	return p.e164 == ""
}

// Equals checks if two phone numbers are equal.
func (p PhoneNumber) Equals(other PhoneNumber) bool {
	return p.e164 == other.e164
}

// MarshalJSON implements json.Marshaler.
func (p PhoneNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"raw":          p.raw,
		"formatted":    p.formatted,
		"e164":         p.e164,
		"country_code": p.countryCode,
		"extension":    p.extension,
		"type":         p.phoneType,
		"is_primary":   p.isPrimary,
	})
}

// ============================================================================
// Address Value Object
// ============================================================================

// AddressType represents the type of address.
type AddressType string

const (
	AddressTypeBilling  AddressType = "billing"
	AddressTypeShipping AddressType = "shipping"
	AddressTypeOffice   AddressType = "office"
	AddressTypeHome     AddressType = "home"
	AddressTypeOther    AddressType = "other"
)

// Address represents a physical address.
type Address struct {
	Line1       string      `json:"line1" bson:"line1"`
	Line2       string      `json:"line2,omitempty" bson:"line2,omitempty"`
	Line3       string      `json:"line3,omitempty" bson:"line3,omitempty"`
	City        string      `json:"city" bson:"city"`
	State       string      `json:"state,omitempty" bson:"state,omitempty"`
	PostalCode  string      `json:"postal_code" bson:"postal_code"`
	Country     string      `json:"country" bson:"country"`
	CountryCode string      `json:"country_code" bson:"country_code"`
	AddressType AddressType `json:"address_type" bson:"address_type"`
	IsPrimary   bool        `json:"is_primary" bson:"is_primary"`
	IsVerified  bool        `json:"is_verified" bson:"is_verified"`
	Latitude    *float64    `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude   *float64    `json:"longitude,omitempty" bson:"longitude,omitempty"`
	Label       string      `json:"label,omitempty" bson:"label,omitempty"`
}

// Valid ISO 3166-1 alpha-2 country codes (subset)
var validCountryCodes = map[string]string{
	"MY": "Malaysia", "SG": "Singapore", "ID": "Indonesia", "TH": "Thailand",
	"VN": "Vietnam", "PH": "Philippines", "US": "United States", "GB": "United Kingdom",
	"AU": "Australia", "NZ": "New Zealand", "JP": "Japan", "KR": "South Korea",
	"CN": "China", "HK": "Hong Kong", "TW": "Taiwan", "IN": "India",
	"DE": "Germany", "FR": "France", "IT": "Italy", "ES": "Spain",
	"NL": "Netherlands", "BE": "Belgium", "CH": "Switzerland", "AT": "Austria",
	"CA": "Canada", "MX": "Mexico", "BR": "Brazil", "AR": "Argentina",
	"AE": "United Arab Emirates", "SA": "Saudi Arabia", "QA": "Qatar",
}

// NewAddress creates a new Address value object.
func NewAddress(line1, city, postalCode, countryCode string, addressType AddressType) (Address, error) {
	line1 = strings.TrimSpace(line1)
	city = strings.TrimSpace(city)
	postalCode = strings.TrimSpace(postalCode)
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))

	var errs ValidationErrors

	if line1 == "" {
		errs.AddField("line1", "address line 1 is required", "REQUIRED")
	} else if len(line1) > 200 {
		errs.AddField("line1", "address line 1 too long (max 200 characters)", "TOO_LONG")
	}

	if city == "" {
		errs.AddField("city", "city is required", "REQUIRED")
	} else if len(city) > 100 {
		errs.AddField("city", "city name too long (max 100 characters)", "TOO_LONG")
	}

	if postalCode == "" {
		errs.AddField("postal_code", "postal code is required", "REQUIRED")
	} else if !isValidPostalCode(postalCode, countryCode) {
		errs.AddField("postal_code", "invalid postal code format for country", "INVALID_FORMAT")
	}

	country, ok := validCountryCodes[countryCode]
	if !ok {
		errs.AddField("country_code", "invalid country code", "INVALID_COUNTRY_CODE")
	}

	if addressType == "" {
		addressType = AddressTypeOther
	}

	if errs.HasErrors() {
		return Address{}, errs
	}

	return Address{
		Line1:       line1,
		City:        city,
		PostalCode:  postalCode,
		Country:     country,
		CountryCode: countryCode,
		AddressType: addressType,
	}, nil
}

// isValidPostalCode validates postal code format based on country.
func isValidPostalCode(postalCode, countryCode string) bool {
	patterns := map[string]*regexp.Regexp{
		"MY": regexp.MustCompile(`^\d{5}$`),                         // Malaysia
		"SG": regexp.MustCompile(`^\d{6}$`),                         // Singapore
		"ID": regexp.MustCompile(`^\d{5}$`),                         // Indonesia
		"US": regexp.MustCompile(`^\d{5}(-\d{4})?$`),                // USA
		"GB": regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}$`), // UK
		"CA": regexp.MustCompile(`^[A-Z]\d[A-Z]\s?\d[A-Z]\d$`),      // Canada
		"AU": regexp.MustCompile(`^\d{4}$`),                         // Australia
		"JP": regexp.MustCompile(`^\d{3}-?\d{4}$`),                  // Japan
		"DE": regexp.MustCompile(`^\d{5}$`),                         // Germany
		"FR": regexp.MustCompile(`^\d{5}$`),                         // France
	}

	if pattern, ok := patterns[countryCode]; ok {
		return pattern.MatchString(strings.ToUpper(postalCode))
	}

	// Default: allow alphanumeric with spaces/hyphens, 3-12 characters
	return regexp.MustCompile(`^[A-Z0-9\s-]{3,12}$`).MatchString(strings.ToUpper(postalCode))
}

// WithDetails adds additional address details.
func (a Address) WithDetails(line2, line3, state string) Address {
	a.Line2 = strings.TrimSpace(line2)
	a.Line3 = strings.TrimSpace(line3)
	a.State = strings.TrimSpace(state)
	return a
}

// WithCoordinates adds GPS coordinates.
func (a Address) WithCoordinates(lat, lng float64) Address {
	a.Latitude = &lat
	a.Longitude = &lng
	return a
}

// WithLabel adds a label to the address.
func (a Address) WithLabel(label string) Address {
	a.Label = strings.TrimSpace(label)
	return a
}

// SetPrimary sets the primary flag.
func (a *Address) SetPrimary(isPrimary bool) {
	a.IsPrimary = isPrimary
}

// SetVerified sets the verified flag.
func (a *Address) SetVerified(isVerified bool) {
	a.IsVerified = isVerified
}

// String returns a formatted address string.
func (a Address) String() string {
	var parts []string

	if a.Line1 != "" {
		parts = append(parts, a.Line1)
	}
	if a.Line2 != "" {
		parts = append(parts, a.Line2)
	}
	if a.Line3 != "" {
		parts = append(parts, a.Line3)
	}

	cityPart := a.City
	if a.PostalCode != "" {
		cityPart = fmt.Sprintf("%s %s", a.City, a.PostalCode)
	}
	if a.State != "" {
		cityPart = fmt.Sprintf("%s, %s", cityPart, a.State)
	}
	parts = append(parts, cityPart)

	if a.Country != "" {
		parts = append(parts, a.Country)
	}

	return strings.Join(parts, ", ")
}

// SingleLine returns the address as a single line.
func (a Address) SingleLine() string {
	return a.String()
}

// IsEmpty returns true if the address is empty.
func (a Address) IsEmpty() bool {
	return a.Line1 == "" && a.City == ""
}

// HasCoordinates returns true if coordinates are set.
func (a Address) HasCoordinates() bool {
	return a.Latitude != nil && a.Longitude != nil
}

// Equals checks if two addresses are equal.
func (a Address) Equals(other Address) bool {
	return a.Line1 == other.Line1 &&
		a.Line2 == other.Line2 &&
		a.City == other.City &&
		a.PostalCode == other.PostalCode &&
		a.CountryCode == other.CountryCode
}

// ============================================================================
// SocialProfile Value Object
// ============================================================================

// SocialPlatform represents a social media platform.
type SocialPlatform string

const (
	SocialPlatformLinkedIn   SocialPlatform = "linkedin"
	SocialPlatformFacebook   SocialPlatform = "facebook"
	SocialPlatformTwitter    SocialPlatform = "twitter"
	SocialPlatformInstagram  SocialPlatform = "instagram"
	SocialPlatformYouTube    SocialPlatform = "youtube"
	SocialPlatformTikTok     SocialPlatform = "tiktok"
	SocialPlatformWhatsApp   SocialPlatform = "whatsapp"
	SocialPlatformTelegram   SocialPlatform = "telegram"
	SocialPlatformGitHub     SocialPlatform = "github"
	SocialPlatformWebsite    SocialPlatform = "website"
	SocialPlatformOther      SocialPlatform = "other"
)

// SocialProfile represents a social media profile.
type SocialProfile struct {
	Platform     SocialPlatform `json:"platform" bson:"platform"`
	URL          string         `json:"url" bson:"url"`
	Username     string         `json:"username,omitempty" bson:"username,omitempty"`
	DisplayName  string         `json:"display_name,omitempty" bson:"display_name,omitempty"`
	Followers    *int           `json:"followers,omitempty" bson:"followers,omitempty"`
	IsVerified   bool           `json:"is_verified" bson:"is_verified"`
	LastSyncedAt *string        `json:"last_synced_at,omitempty" bson:"last_synced_at,omitempty"`
}

// Platform URL patterns for validation
var platformPatterns = map[SocialPlatform]*regexp.Regexp{
	SocialPlatformLinkedIn:  regexp.MustCompile(`^https?://(www\.)?linkedin\.com/(in|company)/[\w-]+/?$`),
	SocialPlatformFacebook:  regexp.MustCompile(`^https?://(www\.)?facebook\.com/[\w.-]+/?$`),
	SocialPlatformTwitter:   regexp.MustCompile(`^https?://(www\.)?(twitter|x)\.com/[\w]+/?$`),
	SocialPlatformInstagram: regexp.MustCompile(`^https?://(www\.)?instagram\.com/[\w.]+/?$`),
	SocialPlatformYouTube:   regexp.MustCompile(`^https?://(www\.)?youtube\.com/(c/|channel/|user/|@)?[\w-]+/?$`),
	SocialPlatformTikTok:    regexp.MustCompile(`^https?://(www\.)?tiktok\.com/@[\w.]+/?$`),
	SocialPlatformGitHub:    regexp.MustCompile(`^https?://(www\.)?github\.com/[\w-]+/?$`),
}

// NewSocialProfile creates a new SocialProfile value object.
func NewSocialProfile(platform SocialPlatform, profileURL string) (SocialProfile, error) {
	profileURL = strings.TrimSpace(profileURL)

	if profileURL == "" {
		return SocialProfile{}, NewValidationError("url", "social profile URL is required", "REQUIRED")
	}

	// Validate URL format
	parsedURL, err := url.Parse(profileURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return SocialProfile{}, ErrInvalidURL
	}

	// Validate platform-specific URL pattern
	if pattern, ok := platformPatterns[platform]; ok {
		if !pattern.MatchString(profileURL) {
			return SocialProfile{}, NewValidationError("url", "invalid URL format for platform", "INVALID_PLATFORM_URL")
		}
	}

	// Extract username from URL
	username := extractUsername(platform, profileURL)

	return SocialProfile{
		Platform: platform,
		URL:      profileURL,
		Username: username,
	}, nil
}

// extractUsername extracts username from platform URL.
func extractUsername(platform SocialPlatform, profileURL string) string {
	parsedURL, err := url.Parse(profileURL)
	if err != nil {
		return ""
	}

	path := strings.Trim(parsedURL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return ""
	}

	switch platform {
	case SocialPlatformLinkedIn:
		if len(parts) >= 2 {
			return parts[1]
		}
	case SocialPlatformTwitter, SocialPlatformInstagram, SocialPlatformGitHub:
		return parts[len(parts)-1]
	case SocialPlatformTikTok:
		username := parts[len(parts)-1]
		return strings.TrimPrefix(username, "@")
	case SocialPlatformFacebook:
		return parts[len(parts)-1]
	case SocialPlatformYouTube:
		if len(parts) >= 1 {
			last := parts[len(parts)-1]
			return strings.TrimPrefix(last, "@")
		}
	}

	return ""
}

// WithDisplayName sets the display name.
func (s SocialProfile) WithDisplayName(name string) SocialProfile {
	s.DisplayName = strings.TrimSpace(name)
	return s
}

// WithFollowers sets the follower count.
func (s SocialProfile) WithFollowers(count int) SocialProfile {
	s.Followers = &count
	return s
}

// SetVerified sets the verified flag.
func (s *SocialProfile) SetVerified(isVerified bool) {
	s.IsVerified = isVerified
}

// String returns the URL.
func (s SocialProfile) String() string {
	return s.URL
}

// IsEmpty returns true if the profile is empty.
func (s SocialProfile) IsEmpty() bool {
	return s.URL == ""
}

// Equals checks if two profiles are equal.
func (s SocialProfile) Equals(other SocialProfile) bool {
	return s.Platform == other.Platform && s.URL == other.URL
}

// ============================================================================
// Money Value Object (for customer financials)
// ============================================================================

// Currency represents a currency code.
type Currency string

const (
	CurrencyMYR Currency = "MYR"
	CurrencySGD Currency = "SGD"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyIDR Currency = "IDR"
	CurrencyJPY Currency = "JPY"
	CurrencyAUD Currency = "AUD"
	CurrencyCNY Currency = "CNY"
	CurrencyINR Currency = "INR"
)

// ValidCurrencies is a set of valid currency codes.
var ValidCurrencies = map[Currency]bool{
	CurrencyMYR: true, CurrencySGD: true, CurrencyUSD: true, CurrencyEUR: true,
	CurrencyGBP: true, CurrencyIDR: true, CurrencyJPY: true, CurrencyAUD: true,
	CurrencyCNY: true, CurrencyINR: true,
}

// Money represents a monetary value with currency.
type Money struct {
	Amount   int64    `json:"amount" bson:"amount"`     // Amount in smallest unit (cents)
	Currency Currency `json:"currency" bson:"currency"`
}

// NewMoney creates a new Money value object.
func NewMoney(amount int64, currency Currency) (Money, error) {
	if !ValidCurrencies[currency] {
		return Money{}, ErrInvalidCurrency
	}
	return Money{Amount: amount, Currency: currency}, nil
}

// NewMoneyFromFloat creates Money from a float amount.
func NewMoneyFromFloat(amount float64, currency Currency) (Money, error) {
	if !ValidCurrencies[currency] {
		return Money{}, ErrInvalidCurrency
	}
	// Convert to smallest unit (assumes 2 decimal places)
	cents := int64(amount * 100)
	return Money{Amount: cents, Currency: currency}, nil
}

// Float returns the amount as a float.
func (m Money) Float() float64 {
	return float64(m.Amount) / 100
}

// String returns formatted money string.
func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.Currency, m.Float())
}

// Add adds two money values.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract subtracts money values.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// Multiply multiplies money by a factor.
func (m Money) Multiply(factor float64) Money {
	return Money{Amount: int64(float64(m.Amount) * factor), Currency: m.Currency}
}

// IsZero returns true if amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive returns true if amount is positive.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// IsNegative returns true if amount is negative.
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// Equals checks if two money values are equal.
func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// ============================================================================
// Website Value Object
// ============================================================================

// Website represents a validated website URL.
type Website struct {
	url        string
	domain     string
	normalized string
}

// NewWebsite creates a new Website value object.
func NewWebsite(urlStr string) (Website, error) {
	urlStr = strings.TrimSpace(urlStr)

	if urlStr == "" {
		return Website{}, nil // Empty is valid
	}

	// Add https if no scheme
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return Website{}, ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return Website{}, ErrInvalidURL
	}

	// Validate host has at least one dot (basic domain check)
	if !strings.Contains(parsedURL.Host, ".") {
		return Website{}, ErrInvalidURL
	}

	domain := parsedURL.Host
	normalized := parsedURL.String()

	return Website{
		url:        urlStr,
		domain:     domain,
		normalized: normalized,
	}, nil
}

// String returns the URL.
func (w Website) String() string {
	return w.normalized
}

// Domain returns the domain.
func (w Website) Domain() string {
	return w.domain
}

// IsEmpty returns true if empty.
func (w Website) IsEmpty() bool {
	return w.url == ""
}

// MarshalJSON implements json.Marshaler.
func (w Website) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.normalized)
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *Website) UnmarshalJSON(data []byte) error {
	var urlStr string
	if err := json.Unmarshal(data, &urlStr); err != nil {
		return err
	}
	website, err := NewWebsite(urlStr)
	if err != nil {
		return err
	}
	*w = website
	return nil
}

// ============================================================================
// Name Value Object (for contact names)
// ============================================================================

// PersonName represents a person's name with proper formatting.
type PersonName struct {
	Title      string `json:"title,omitempty" bson:"title,omitempty"`
	FirstName  string `json:"first_name" bson:"first_name"`
	MiddleName string `json:"middle_name,omitempty" bson:"middle_name,omitempty"`
	LastName   string `json:"last_name" bson:"last_name"`
	Suffix     string `json:"suffix,omitempty" bson:"suffix,omitempty"`
	Nickname   string `json:"nickname,omitempty" bson:"nickname,omitempty"`
}

// Valid titles
var validTitles = map[string]bool{
	"Mr.": true, "Mrs.": true, "Ms.": true, "Miss": true, "Dr.": true,
	"Prof.": true, "Rev.": true, "Sir": true, "Dame": true,
	"Dato'": true, "Datin": true, "Tan Sri": true, "Puan Sri": true,
	"Tun": true, "Toh Puan": true,
}

// NewPersonName creates a new PersonName value object.
func NewPersonName(firstName, lastName string) (PersonName, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if firstName == "" && lastName == "" {
		return PersonName{}, NewValidationError("name", "at least first name or last name is required", "REQUIRED")
	}

	// Capitalize names properly
	firstName = capitalizeName(firstName)
	lastName = capitalizeName(lastName)

	return PersonName{
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

// WithTitle adds a title.
func (n PersonName) WithTitle(title string) PersonName {
	title = strings.TrimSpace(title)
	if validTitles[title] {
		n.Title = title
	}
	return n
}

// WithMiddleName adds a middle name.
func (n PersonName) WithMiddleName(middleName string) PersonName {
	n.MiddleName = capitalizeName(strings.TrimSpace(middleName))
	return n
}

// WithSuffix adds a suffix.
func (n PersonName) WithSuffix(suffix string) PersonName {
	n.Suffix = strings.TrimSpace(suffix)
	return n
}

// WithNickname adds a nickname.
func (n PersonName) WithNickname(nickname string) PersonName {
	n.Nickname = strings.TrimSpace(nickname)
	return n
}

// capitalizeName properly capitalizes a name.
func capitalizeName(name string) string {
	if name == "" {
		return ""
	}

	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// FullName returns the full formatted name.
func (n PersonName) FullName() string {
	var parts []string

	if n.Title != "" {
		parts = append(parts, n.Title)
	}
	if n.FirstName != "" {
		parts = append(parts, n.FirstName)
	}
	if n.MiddleName != "" {
		parts = append(parts, n.MiddleName)
	}
	if n.LastName != "" {
		parts = append(parts, n.LastName)
	}
	if n.Suffix != "" {
		parts = append(parts, n.Suffix)
	}

	return strings.Join(parts, " ")
}

// DisplayName returns a display-friendly name.
func (n PersonName) DisplayName() string {
	if n.Nickname != "" {
		return n.Nickname
	}
	if n.FirstName != "" {
		return n.FirstName
	}
	return n.LastName
}

// Initials returns the initials.
func (n PersonName) Initials() string {
	var initials string
	if n.FirstName != "" {
		initials += string([]rune(n.FirstName)[0])
	}
	if n.LastName != "" {
		initials += string([]rune(n.LastName)[0])
	}
	return strings.ToUpper(initials)
}

// IsEmpty returns true if the name is empty.
func (n PersonName) IsEmpty() bool {
	return n.FirstName == "" && n.LastName == ""
}

// String returns the full name.
func (n PersonName) String() string {
	return n.FullName()
}
