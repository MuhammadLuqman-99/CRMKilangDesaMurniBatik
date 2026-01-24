// Package domain contains the domain layer for the Customer service.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// CustomerType represents the type of customer.
type CustomerType string

const (
	CustomerTypeIndividual CustomerType = "individual"
	CustomerTypeCompany    CustomerType = "company"
	CustomerTypePartner    CustomerType = "partner"
	CustomerTypeReseller   CustomerType = "reseller"
)

// CustomerStatus represents the status of a customer.
type CustomerStatus string

const (
	CustomerStatusLead      CustomerStatus = "lead"
	CustomerStatusProspect  CustomerStatus = "prospect"
	CustomerStatusActive    CustomerStatus = "active"
	CustomerStatusInactive  CustomerStatus = "inactive"
	CustomerStatusChurned   CustomerStatus = "churned"
	CustomerStatusBlocked   CustomerStatus = "blocked"
)

// CustomerTier represents the customer tier/segment.
type CustomerTier string

const (
	CustomerTierStandard   CustomerTier = "standard"
	CustomerTierBronze     CustomerTier = "bronze"
	CustomerTierSilver     CustomerTier = "silver"
	CustomerTierGold       CustomerTier = "gold"
	CustomerTierPlatinum   CustomerTier = "platinum"
	CustomerTierEnterprise CustomerTier = "enterprise"
)

// CustomerSource represents how the customer was acquired.
type CustomerSource string

const (
	CustomerSourceDirect     CustomerSource = "direct"
	CustomerSourceReferral   CustomerSource = "referral"
	CustomerSourceWebsite    CustomerSource = "website"
	CustomerSourceSocialMedia CustomerSource = "social_media"
	CustomerSourceEvent      CustomerSource = "event"
	CustomerSourcePartner    CustomerSource = "partner"
	CustomerSourceColdCall   CustomerSource = "cold_call"
	CustomerSourceImport     CustomerSource = "import"
	CustomerSourceOther      CustomerSource = "other"
)

// Industry represents the customer's industry.
type Industry string

const (
	IndustryTechnology    Industry = "technology"
	IndustryFinance       Industry = "finance"
	IndustryHealthcare    Industry = "healthcare"
	IndustryManufacturing Industry = "manufacturing"
	IndustryRetail        Industry = "retail"
	IndustryEducation     Industry = "education"
	IndustryGovernment    Industry = "government"
	IndustryNonProfit     Industry = "non_profit"
	IndustryRealEstate    Industry = "real_estate"
	IndustryHospitality   Industry = "hospitality"
	IndustryMedia         Industry = "media"
	IndustryAgriculture   Industry = "agriculture"
	IndustryConstruction  Industry = "construction"
	IndustryTransportation Industry = "transportation"
	IndustryOther         Industry = "other"
)

// CompanySize represents the size of a company.
type CompanySize string

const (
	CompanySizeSolo       CompanySize = "solo"        // 1 person
	CompanySizeMicro      CompanySize = "micro"       // 2-10
	CompanySizeSmall      CompanySize = "small"       // 11-50
	CompanySizeMedium     CompanySize = "medium"      // 51-200
	CompanySizeLarge      CompanySize = "large"       // 201-1000
	CompanySizeEnterprise CompanySize = "enterprise"  // 1000+
)

// CompanyInfo holds company-specific information.
type CompanyInfo struct {
	LegalName          string      `json:"legal_name,omitempty" bson:"legal_name,omitempty"`
	TradingName        string      `json:"trading_name,omitempty" bson:"trading_name,omitempty"`
	RegistrationNumber string      `json:"registration_number,omitempty" bson:"registration_number,omitempty"`
	TaxID              string      `json:"tax_id,omitempty" bson:"tax_id,omitempty"`
	Industry           Industry    `json:"industry,omitempty" bson:"industry,omitempty"`
	Size               CompanySize `json:"size,omitempty" bson:"size,omitempty"`
	EmployeeCount      *int        `json:"employee_count,omitempty" bson:"employee_count,omitempty"`
	AnnualRevenue      *Money      `json:"annual_revenue,omitempty" bson:"annual_revenue,omitempty"`
	FoundedYear        *int        `json:"founded_year,omitempty" bson:"founded_year,omitempty"`
	Description        string      `json:"description,omitempty" bson:"description,omitempty"`
	ParentCompanyID    *uuid.UUID  `json:"parent_company_id,omitempty" bson:"parent_company_id,omitempty"`
}

// CustomerFinancials holds financial information.
type CustomerFinancials struct {
	Currency           Currency   `json:"currency" bson:"currency"`
	CreditLimit        *Money     `json:"credit_limit,omitempty" bson:"credit_limit,omitempty"`
	CurrentBalance     *Money     `json:"current_balance,omitempty" bson:"current_balance,omitempty"`
	LifetimeValue      *Money     `json:"lifetime_value,omitempty" bson:"lifetime_value,omitempty"`
	PaymentTerms       int        `json:"payment_terms" bson:"payment_terms"` // Days
	TaxExempt          bool       `json:"tax_exempt" bson:"tax_exempt"`
	TaxExemptionID     string     `json:"tax_exemption_id,omitempty" bson:"tax_exemption_id,omitempty"`
	BillingEmail       string     `json:"billing_email,omitempty" bson:"billing_email,omitempty"`
	DefaultDiscountPct float64    `json:"default_discount_pct" bson:"default_discount_pct"`
	LastPaymentAt      *time.Time `json:"last_payment_at,omitempty" bson:"last_payment_at,omitempty"`
	LastPurchaseAt     *time.Time `json:"last_purchase_at,omitempty" bson:"last_purchase_at,omitempty"`
	TotalPurchases     int        `json:"total_purchases" bson:"total_purchases"`
	TotalSpent         *Money     `json:"total_spent,omitempty" bson:"total_spent,omitempty"`
}

// CustomerPreferences holds customer preferences.
type CustomerPreferences struct {
	Language           string                  `json:"language" bson:"language"`
	Timezone           string                  `json:"timezone" bson:"timezone"`
	Currency           Currency                `json:"currency" bson:"currency"`
	DateFormat         string                  `json:"date_format" bson:"date_format"`
	CommPreference     CommunicationPreference `json:"comm_preference" bson:"comm_preference"`
	OptedOutMarketing  bool                    `json:"opted_out_marketing" bson:"opted_out_marketing"`
	MarketingConsent   *time.Time              `json:"marketing_consent,omitempty" bson:"marketing_consent,omitempty"`
	NewsletterOptIn    bool                    `json:"newsletter_opt_in" bson:"newsletter_opt_in"`
	SMSOptIn           bool                    `json:"sms_opt_in" bson:"sms_opt_in"`
}

// CustomerStats holds computed statistics.
type CustomerStats struct {
	ContactCount         int        `json:"contact_count" bson:"contact_count"`
	ActiveContactCount   int        `json:"active_contact_count" bson:"active_contact_count"`
	NoteCount            int        `json:"note_count" bson:"note_count"`
	ActivityCount        int        `json:"activity_count" bson:"activity_count"`
	DealCount            int        `json:"deal_count" bson:"deal_count"`
	WonDealCount         int        `json:"won_deal_count" bson:"won_deal_count"`
	LostDealCount        int        `json:"lost_deal_count" bson:"lost_deal_count"`
	OpenDealValue        *Money     `json:"open_deal_value,omitempty" bson:"open_deal_value,omitempty"`
	DaysSinceLastContact *int       `json:"days_since_last_contact,omitempty" bson:"days_since_last_contact,omitempty"`
	AvgDealSize          *Money     `json:"avg_deal_size,omitempty" bson:"avg_deal_size,omitempty"`
	EngagementScore      int        `json:"engagement_score" bson:"engagement_score"`
	HealthScore          int        `json:"health_score" bson:"health_score"`
	LastCalculatedAt     *time.Time `json:"last_calculated_at,omitempty" bson:"last_calculated_at,omitempty"`
}

// Customer is the aggregate root for customer management.
type Customer struct {
	BaseAggregateRoot
	TenantID           uuid.UUID           `json:"tenant_id" bson:"tenant_id"`
	Code               string              `json:"code" bson:"code"` // Unique customer code
	Name               string              `json:"name" bson:"name"`
	Type               CustomerType        `json:"type" bson:"type"`
	Status             CustomerStatus      `json:"status" bson:"status"`
	Tier               CustomerTier        `json:"tier" bson:"tier"`
	Source             CustomerSource      `json:"source" bson:"source"`
	Email              Email               `json:"email" bson:"email"`
	PhoneNumbers       []PhoneNumber       `json:"phone_numbers" bson:"phone_numbers"`
	Website            Website             `json:"website,omitempty" bson:"website,omitempty"`
	Addresses          []Address           `json:"addresses" bson:"addresses"`
	SocialProfiles     []SocialProfile     `json:"social_profiles,omitempty" bson:"social_profiles,omitempty"`
	CompanyInfo        *CompanyInfo        `json:"company_info,omitempty" bson:"company_info,omitempty"`
	Financials         CustomerFinancials  `json:"financials" bson:"financials"`
	Preferences        CustomerPreferences `json:"preferences" bson:"preferences"`
	Stats              CustomerStats       `json:"stats" bson:"stats"`
	Contacts           []Contact           `json:"contacts" bson:"contacts"`
	OwnerID            *uuid.UUID          `json:"owner_id,omitempty" bson:"owner_id,omitempty"` // Sales owner
	AssignedTeam       []uuid.UUID         `json:"assigned_team,omitempty" bson:"assigned_team,omitempty"`
	ReferredBy         *uuid.UUID          `json:"referred_by,omitempty" bson:"referred_by,omitempty"`
	Tags               []string            `json:"tags" bson:"tags"`
	Segments           []uuid.UUID         `json:"segments,omitempty" bson:"segments,omitempty"`
	CustomFields       map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
	Notes              string              `json:"notes,omitempty" bson:"notes,omitempty"`
	LogoURL            string              `json:"logo_url,omitempty" bson:"logo_url,omitempty"`
	AuditInfo          AuditInfo           `json:"audit_info" bson:"audit_info"`
	Metadata           Metadata            `json:"metadata" bson:"metadata"`
	LastContactedAt    *time.Time          `json:"last_contacted_at,omitempty" bson:"last_contacted_at,omitempty"`
	NextFollowUpAt     *time.Time          `json:"next_follow_up_at,omitempty" bson:"next_follow_up_at,omitempty"`
	ConvertedAt        *time.Time          `json:"converted_at,omitempty" bson:"converted_at,omitempty"`
	ChurnedAt          *time.Time          `json:"churned_at,omitempty" bson:"churned_at,omitempty"`
	ChurnReason        string              `json:"churn_reason,omitempty" bson:"churn_reason,omitempty"`
}

// MaxContacts is the maximum number of contacts per customer.
const MaxContacts = 100

// CustomerBuilder provides a fluent API for building Customer aggregates.
type CustomerBuilder struct {
	customer Customer
	errors   ValidationErrors
}

// NewCustomerBuilder creates a new CustomerBuilder.
func NewCustomerBuilder(tenantID uuid.UUID, name string, customerType CustomerType) *CustomerBuilder {
	return &CustomerBuilder{
		customer: Customer{
			BaseAggregateRoot: NewBaseAggregateRoot(),
			TenantID:          tenantID,
			Name:              strings.TrimSpace(name),
			Type:              customerType,
			Status:            CustomerStatusLead,
			Tier:              CustomerTierStandard,
			Source:            CustomerSourceDirect,
			PhoneNumbers:      make([]PhoneNumber, 0),
			Addresses:         make([]Address, 0),
			SocialProfiles:    make([]SocialProfile, 0),
			Contacts:          make([]Contact, 0),
			Tags:              make([]string, 0),
			Segments:          make([]uuid.UUID, 0),
			AssignedTeam:      make([]uuid.UUID, 0),
			CustomFields:      make(map[string]interface{}),
			Preferences: CustomerPreferences{
				Language:       "en",
				Currency:       CurrencyMYR,
				CommPreference: CommPrefEmail,
			},
			Financials: CustomerFinancials{
				Currency:     CurrencyMYR,
				PaymentTerms: 30,
			},
		},
	}
}

// WithCode sets the customer code.
func (b *CustomerBuilder) WithCode(code string) *CustomerBuilder {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code != "" && len(code) > 50 {
		b.errors.AddField("code", "code too long (max 50 characters)", "TOO_LONG")
	}
	b.customer.Code = code
	return b
}

// WithEmail sets the customer email.
func (b *CustomerBuilder) WithEmail(email string) *CustomerBuilder {
	e, err := NewEmail(email)
	if err != nil {
		b.errors.AddField("email", err.Error(), "INVALID_EMAIL")
	} else {
		b.customer.Email = e
	}
	return b
}

// WithPhone adds a phone number.
func (b *CustomerBuilder) WithPhone(number string, phoneType PhoneType, isPrimary bool) *CustomerBuilder {
	phone, err := NewPhoneNumberWithPrimary(number, phoneType, isPrimary)
	if err != nil {
		b.errors.AddField("phone", err.Error(), "INVALID_PHONE")
	} else {
		b.customer.PhoneNumbers = append(b.customer.PhoneNumbers, phone)
	}
	return b
}

// WithWebsite sets the website.
func (b *CustomerBuilder) WithWebsite(url string) *CustomerBuilder {
	website, err := NewWebsite(url)
	if err != nil {
		b.errors.AddField("website", err.Error(), "INVALID_WEBSITE")
	} else {
		b.customer.Website = website
	}
	return b
}

// WithAddress adds an address.
func (b *CustomerBuilder) WithAddress(line1, city, postalCode, countryCode string, addressType AddressType) *CustomerBuilder {
	addr, err := NewAddress(line1, city, postalCode, countryCode, addressType)
	if err != nil {
		if verrs, ok := err.(ValidationErrors); ok {
			for _, e := range verrs {
				b.errors.Add(e)
			}
		}
	} else {
		b.customer.Addresses = append(b.customer.Addresses, addr)
	}
	return b
}

// WithSource sets the customer source.
func (b *CustomerBuilder) WithSource(source CustomerSource) *CustomerBuilder {
	b.customer.Source = source
	return b
}

// WithTier sets the customer tier.
func (b *CustomerBuilder) WithTier(tier CustomerTier) *CustomerBuilder {
	b.customer.Tier = tier
	return b
}

// WithOwner sets the sales owner.
func (b *CustomerBuilder) WithOwner(ownerID uuid.UUID) *CustomerBuilder {
	b.customer.OwnerID = &ownerID
	return b
}

// WithCompanyInfo sets company information.
func (b *CustomerBuilder) WithCompanyInfo(info CompanyInfo) *CustomerBuilder {
	b.customer.CompanyInfo = &info
	return b
}

// WithTags adds tags.
func (b *CustomerBuilder) WithTags(tags ...string) *CustomerBuilder {
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" {
			b.customer.Tags = append(b.customer.Tags, tag)
		}
	}
	return b
}

// WithNotes sets notes.
func (b *CustomerBuilder) WithNotes(notes string) *CustomerBuilder {
	b.customer.Notes = strings.TrimSpace(notes)
	return b
}

// WithPreferences sets preferences.
func (b *CustomerBuilder) WithPreferences(prefs CustomerPreferences) *CustomerBuilder {
	b.customer.Preferences = prefs
	return b
}

// WithCustomField adds a custom field.
func (b *CustomerBuilder) WithCustomField(key string, value interface{}) *CustomerBuilder {
	b.customer.CustomFields[key] = value
	return b
}

// WithCreatedBy sets the creator.
func (b *CustomerBuilder) WithCreatedBy(userID uuid.UUID) *CustomerBuilder {
	b.customer.AuditInfo.SetCreatedBy(userID)
	return b
}

// WithMetadata sets metadata.
func (b *CustomerBuilder) WithMetadata(meta Metadata) *CustomerBuilder {
	b.customer.Metadata = meta
	return b
}

// Build creates the Customer aggregate.
func (b *CustomerBuilder) Build() (*Customer, error) {
	// Validate required fields
	if b.customer.Name == "" {
		b.errors.AddField("name", "name is required", "REQUIRED")
	} else if len(b.customer.Name) > 255 {
		b.errors.AddField("name", "name too long (max 255 characters)", "TOO_LONG")
	}

	if b.customer.TenantID == uuid.Nil {
		b.errors.AddField("tenant_id", "tenant ID is required", "REQUIRED")
	}

	if b.errors.HasErrors() {
		return nil, b.errors
	}

	// Generate code if not provided
	if b.customer.Code == "" {
		b.customer.Code = generateCustomerCode(b.customer.Type)
	}

	// Add CustomerCreated event
	b.customer.AddDomainEvent(NewCustomerCreatedEvent(&b.customer))

	return &b.customer, nil
}

// generateCustomerCode generates a unique customer code.
func generateCustomerCode(customerType CustomerType) string {
	prefix := "C"
	switch customerType {
	case CustomerTypeCompany:
		prefix = "CO"
	case CustomerTypePartner:
		prefix = "PA"
	case CustomerTypeReseller:
		prefix = "RS"
	case CustomerTypeIndividual:
		prefix = "IN"
	}
	return prefix + strings.ToUpper(uuid.New().String()[:8])
}

// NewCustomer creates a new Customer with basic information.
func NewCustomer(tenantID uuid.UUID, name string, customerType CustomerType) (*Customer, error) {
	return NewCustomerBuilder(tenantID, name, customerType).Build()
}

// ============================================================================
// Customer Behaviors
// ============================================================================

// UpdateBasicInfo updates basic customer information.
func (c *Customer) UpdateBasicInfo(name string, customerType CustomerType) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return NewValidationError("name", "name is required", "REQUIRED")
	}
	if len(name) > 255 {
		return NewValidationError("name", "name too long (max 255 characters)", "TOO_LONG")
	}

	oldName := c.Name
	c.Name = name
	c.Type = customerType
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerUpdatedEvent(c, map[string]interface{}{
		"name": map[string]string{"old": oldName, "new": name},
	}))

	return nil
}

// UpdateEmail updates the customer email.
func (c *Customer) UpdateEmail(email Email) {
	oldEmail := c.Email.String()
	c.Email = email
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerUpdatedEvent(c, map[string]interface{}{
		"email": map[string]string{"old": oldEmail, "new": email.String()},
	}))
}

// AddPhone adds a phone number.
func (c *Customer) AddPhone(phone PhoneNumber) {
	if phone.IsPrimary() {
		for i := range c.PhoneNumbers {
			c.PhoneNumbers[i].SetPrimary(false)
		}
	}
	c.PhoneNumbers = append(c.PhoneNumbers, phone)
	c.MarkUpdated()
	c.IncrementVersion()
}

// RemovePhone removes a phone number.
func (c *Customer) RemovePhone(e164 string) bool {
	for i, phone := range c.PhoneNumbers {
		if phone.E164() == e164 {
			c.PhoneNumbers = append(c.PhoneNumbers[:i], c.PhoneNumbers[i+1:]...)
			c.MarkUpdated()
			c.IncrementVersion()
			return true
		}
	}
	return false
}

// GetPrimaryPhone returns the primary phone.
func (c *Customer) GetPrimaryPhone() *PhoneNumber {
	for _, phone := range c.PhoneNumbers {
		if phone.IsPrimary() {
			return &phone
		}
	}
	if len(c.PhoneNumbers) > 0 {
		return &c.PhoneNumbers[0]
	}
	return nil
}

// AddAddress adds an address.
func (c *Customer) AddAddress(address Address) {
	if address.IsPrimary {
		for i := range c.Addresses {
			if c.Addresses[i].AddressType == address.AddressType {
				c.Addresses[i].IsPrimary = false
			}
		}
	}
	c.Addresses = append(c.Addresses, address)
	c.MarkUpdated()
	c.IncrementVersion()
}

// GetBillingAddress returns the billing address.
func (c *Customer) GetBillingAddress() *Address {
	for _, addr := range c.Addresses {
		if addr.AddressType == AddressTypeBilling && addr.IsPrimary {
			return &addr
		}
	}
	for _, addr := range c.Addresses {
		if addr.AddressType == AddressTypeBilling {
			return &addr
		}
	}
	return nil
}

// GetShippingAddress returns the shipping address.
func (c *Customer) GetShippingAddress() *Address {
	for _, addr := range c.Addresses {
		if addr.AddressType == AddressTypeShipping && addr.IsPrimary {
			return &addr
		}
	}
	for _, addr := range c.Addresses {
		if addr.AddressType == AddressTypeShipping {
			return &addr
		}
	}
	return nil
}

// ============================================================================
// Contact Management
// ============================================================================

// AddContact adds a contact to the customer.
func (c *Customer) AddContact(contact *Contact) error {
	if len(c.Contacts) >= MaxContacts {
		return ErrMaxContactsExceeded
	}

	// Check for duplicate email
	for _, existing := range c.Contacts {
		if !contact.Email.IsEmpty() && contact.Email.Equals(existing.Email) {
			return ErrDuplicateContactEmail
		}
	}

	// If this is primary, unmark others
	if contact.IsPrimary {
		for i := range c.Contacts {
			c.Contacts[i].SetPrimary(false)
		}
	}

	// If first contact, make it primary
	if len(c.Contacts) == 0 {
		contact.SetPrimary(true)
	}

	c.Contacts = append(c.Contacts, *contact)
	c.Stats.ContactCount = len(c.Contacts)
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewContactAddedEvent(c.ID, c.TenantID, contact.ID))

	return nil
}

// UpdateContact updates a contact.
func (c *Customer) UpdateContact(contactID uuid.UUID, updateFn func(*Contact) error) error {
	for i := range c.Contacts {
		if c.Contacts[i].ID == contactID {
			if err := updateFn(&c.Contacts[i]); err != nil {
				return err
			}
			c.MarkUpdated()
			c.IncrementVersion()
			return nil
		}
	}
	return ErrContactNotFound
}

// RemoveContact removes a contact.
func (c *Customer) RemoveContact(contactID uuid.UUID) error {
	for i, contact := range c.Contacts {
		if contact.ID == contactID {
			// Cannot remove primary if others exist
			if contact.IsPrimary && len(c.Contacts) > 1 {
				return ErrCannotDeletePrimaryContact
			}

			c.Contacts = append(c.Contacts[:i], c.Contacts[i+1:]...)
			c.Stats.ContactCount = len(c.Contacts)

			// If removed was primary and others exist, make first one primary
			if contact.IsPrimary && len(c.Contacts) > 0 {
				c.Contacts[0].SetPrimary(true)
			}

			c.MarkUpdated()
			c.IncrementVersion()

			c.AddDomainEvent(NewContactRemovedEvent(c.ID, c.TenantID, contactID))

			return nil
		}
	}
	return ErrContactNotFound
}

// GetContact returns a contact by ID.
func (c *Customer) GetContact(contactID uuid.UUID) *Contact {
	for i := range c.Contacts {
		if c.Contacts[i].ID == contactID {
			return &c.Contacts[i]
		}
	}
	return nil
}

// GetPrimaryContact returns the primary contact.
func (c *Customer) GetPrimaryContact() *Contact {
	for i := range c.Contacts {
		if c.Contacts[i].IsPrimary {
			return &c.Contacts[i]
		}
	}
	if len(c.Contacts) > 0 {
		return &c.Contacts[0]
	}
	return nil
}

// SetPrimaryContact sets a contact as primary.
func (c *Customer) SetPrimaryContact(contactID uuid.UUID) error {
	found := false
	for i := range c.Contacts {
		if c.Contacts[i].ID == contactID {
			c.Contacts[i].SetPrimary(true)
			found = true
		} else {
			c.Contacts[i].SetPrimary(false)
		}
	}
	if !found {
		return ErrContactNotFound
	}
	c.MarkUpdated()
	c.IncrementVersion()
	return nil
}

// GetActiveContacts returns all active contacts.
func (c *Customer) GetActiveContacts() []Contact {
	var active []Contact
	for _, contact := range c.Contacts {
		if contact.IsActive() {
			active = append(active, contact)
		}
	}
	c.Stats.ActiveContactCount = len(active)
	return active
}

// ============================================================================
// Status Management
// ============================================================================

// Activate activates the customer.
func (c *Customer) Activate() error {
	if c.Status == CustomerStatusBlocked {
		return ErrCustomerBlocked
	}

	oldStatus := c.Status
	c.Status = CustomerStatusActive
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerStatusChangedEvent(c, oldStatus, CustomerStatusActive))

	return nil
}

// ConvertToCustomer converts a lead/prospect to an active customer.
func (c *Customer) ConvertToCustomer() error {
	if c.Status != CustomerStatusLead && c.Status != CustomerStatusProspect {
		return NewDomainError("INVALID_CONVERSION", "can only convert leads or prospects", nil)
	}

	oldStatus := c.Status
	c.Status = CustomerStatusActive
	now := time.Now().UTC()
	c.ConvertedAt = &now
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerConvertedEvent(c, oldStatus))

	return nil
}

// Deactivate deactivates the customer.
func (c *Customer) Deactivate() {
	oldStatus := c.Status
	c.Status = CustomerStatusInactive
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerStatusChangedEvent(c, oldStatus, CustomerStatusInactive))
}

// MarkAsChurned marks the customer as churned.
func (c *Customer) MarkAsChurned(reason string) {
	oldStatus := c.Status
	c.Status = CustomerStatusChurned
	c.ChurnReason = strings.TrimSpace(reason)
	now := time.Now().UTC()
	c.ChurnedAt = &now
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerChurnedEvent(c, reason))
	c.AddDomainEvent(NewCustomerStatusChangedEvent(c, oldStatus, CustomerStatusChurned))
}

// Block blocks the customer.
func (c *Customer) Block(reason string) {
	oldStatus := c.Status
	c.Status = CustomerStatusBlocked
	c.Notes = reason
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerStatusChangedEvent(c, oldStatus, CustomerStatusBlocked))
}

// Unblock unblocks the customer.
func (c *Customer) Unblock() {
	oldStatus := c.Status
	c.Status = CustomerStatusInactive
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerStatusChangedEvent(c, oldStatus, CustomerStatusInactive))
}

// PromoteToProspect promotes a lead to prospect.
func (c *Customer) PromoteToProspect() error {
	if c.Status != CustomerStatusLead {
		return NewDomainError("INVALID_PROMOTION", "can only promote leads to prospects", nil)
	}

	oldStatus := c.Status
	c.Status = CustomerStatusProspect
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerStatusChangedEvent(c, oldStatus, CustomerStatusProspect))

	return nil
}

// ============================================================================
// Tier Management
// ============================================================================

// UpgradeTier upgrades the customer tier.
func (c *Customer) UpgradeTier(newTier CustomerTier) {
	oldTier := c.Tier
	c.Tier = newTier
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerTierChangedEvent(c, oldTier, newTier))
}

// DowngradeTier downgrades the customer tier.
func (c *Customer) DowngradeTier(newTier CustomerTier) {
	c.UpgradeTier(newTier) // Same logic
}

// ============================================================================
// Owner and Assignment
// ============================================================================

// AssignOwner assigns a sales owner.
func (c *Customer) AssignOwner(ownerID uuid.UUID) {
	oldOwner := c.OwnerID
	c.OwnerID = &ownerID
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerOwnerAssignedEvent(c, oldOwner, ownerID))
}

// UnassignOwner removes the sales owner.
func (c *Customer) UnassignOwner() {
	c.OwnerID = nil
	c.MarkUpdated()
	c.IncrementVersion()
}

// AddToTeam adds a team member.
func (c *Customer) AddToTeam(userID uuid.UUID) {
	for _, id := range c.AssignedTeam {
		if id == userID {
			return // Already in team
		}
	}
	c.AssignedTeam = append(c.AssignedTeam, userID)
	c.MarkUpdated()
	c.IncrementVersion()
}

// RemoveFromTeam removes a team member.
func (c *Customer) RemoveFromTeam(userID uuid.UUID) {
	for i, id := range c.AssignedTeam {
		if id == userID {
			c.AssignedTeam = append(c.AssignedTeam[:i], c.AssignedTeam[i+1:]...)
			c.MarkUpdated()
			c.IncrementVersion()
			return
		}
	}
}

// IsAssignedTo checks if a user is assigned.
func (c *Customer) IsAssignedTo(userID uuid.UUID) bool {
	if c.OwnerID != nil && *c.OwnerID == userID {
		return true
	}
	for _, id := range c.AssignedTeam {
		if id == userID {
			return true
		}
	}
	return false
}

// ============================================================================
// Tags and Segments
// ============================================================================

// AddTag adds a tag.
func (c *Customer) AddTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	if tag == "" {
		return
	}
	for _, t := range c.Tags {
		if t == tag {
			return
		}
	}
	c.Tags = append(c.Tags, tag)
	c.MarkUpdated()
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerTagAddedEvent(c, tag))
}

// RemoveTag removes a tag.
func (c *Customer) RemoveTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for i, t := range c.Tags {
		if t == tag {
			c.Tags = append(c.Tags[:i], c.Tags[i+1:]...)
			c.MarkUpdated()
			c.IncrementVersion()

			c.AddDomainEvent(NewCustomerTagRemovedEvent(c, tag))
			return
		}
	}
}

// HasTag checks if customer has a tag.
func (c *Customer) HasTag(tag string) bool {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddToSegment adds to a segment.
func (c *Customer) AddToSegment(segmentID uuid.UUID) {
	for _, id := range c.Segments {
		if id == segmentID {
			return
		}
	}
	c.Segments = append(c.Segments, segmentID)
	c.MarkUpdated()
	c.IncrementVersion()
}

// RemoveFromSegment removes from a segment.
func (c *Customer) RemoveFromSegment(segmentID uuid.UUID) {
	for i, id := range c.Segments {
		if id == segmentID {
			c.Segments = append(c.Segments[:i], c.Segments[i+1:]...)
			c.MarkUpdated()
			c.IncrementVersion()
			return
		}
	}
}

// ============================================================================
// Financial Updates
// ============================================================================

// UpdateFinancials updates financial information.
func (c *Customer) UpdateFinancials(financials CustomerFinancials) {
	c.Financials = financials
	c.MarkUpdated()
	c.IncrementVersion()
}

// RecordPurchase records a purchase.
func (c *Customer) RecordPurchase(amount Money) {
	c.Financials.TotalPurchases++
	now := time.Now().UTC()
	c.Financials.LastPurchaseAt = &now

	if c.Financials.TotalSpent == nil {
		c.Financials.TotalSpent = &amount
	} else {
		if total, err := c.Financials.TotalSpent.Add(amount); err == nil {
			c.Financials.TotalSpent = &total
		}
	}

	c.MarkUpdated()
	c.IncrementVersion()
}

// RecordPayment records a payment.
func (c *Customer) RecordPayment(amount Money) {
	now := time.Now().UTC()
	c.Financials.LastPaymentAt = &now

	if c.Financials.CurrentBalance != nil {
		if balance, err := c.Financials.CurrentBalance.Subtract(amount); err == nil {
			c.Financials.CurrentBalance = &balance
		}
	}

	c.MarkUpdated()
	c.IncrementVersion()
}

// ============================================================================
// Activity Tracking
// ============================================================================

// RecordContact records a contact interaction.
func (c *Customer) RecordContact() {
	now := time.Now().UTC()
	c.LastContactedAt = &now
	c.AuditInfo.RecordContact()
	c.MarkUpdated()
}

// SetNextFollowUp schedules next follow-up.
func (c *Customer) SetNextFollowUp(followUpAt time.Time) {
	c.NextFollowUpAt = &followUpAt
	c.MarkUpdated()
}

// ClearNextFollowUp clears the follow-up.
func (c *Customer) ClearNextFollowUp() {
	c.NextFollowUpAt = nil
	c.MarkUpdated()
}

// UpdateEngagementScore updates the engagement score.
func (c *Customer) UpdateEngagementScore(score int) {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	c.Stats.EngagementScore = score
	c.MarkUpdated()
}

// UpdateHealthScore updates the health score.
func (c *Customer) UpdateHealthScore(score int) {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	c.Stats.HealthScore = score
	c.MarkUpdated()
}

// ============================================================================
// Query Methods
// ============================================================================

// IsActive returns true if customer is active.
func (c *Customer) IsActive() bool {
	return c.Status == CustomerStatusActive && !c.IsDeleted()
}

// IsLead returns true if customer is a lead.
func (c *Customer) IsLead() bool {
	return c.Status == CustomerStatusLead
}

// IsProspect returns true if customer is a prospect.
func (c *Customer) IsProspect() bool {
	return c.Status == CustomerStatusProspect
}

// IsChurned returns true if customer has churned.
func (c *Customer) IsChurned() bool {
	return c.Status == CustomerStatusChurned
}

// IsBlocked returns true if customer is blocked.
func (c *Customer) IsBlocked() bool {
	return c.Status == CustomerStatusBlocked
}

// IsCompany returns true if customer is a company.
func (c *Customer) IsCompany() bool {
	return c.Type == CustomerTypeCompany
}

// CanReceiveMarketing returns true if customer can receive marketing.
func (c *Customer) CanReceiveMarketing() bool {
	return c.IsActive() && !c.Preferences.OptedOutMarketing && !c.Email.IsEmpty()
}

// NeedsFollowUp returns true if follow-up is due.
func (c *Customer) NeedsFollowUp() bool {
	if c.NextFollowUpAt == nil {
		return false
	}
	return time.Now().After(*c.NextFollowUpAt)
}

// DaysSinceLastContact returns days since last contact.
func (c *Customer) DaysSinceLastContact() int {
	if c.LastContactedAt == nil {
		return -1
	}
	return int(time.Since(*c.LastContactedAt).Hours() / 24)
}

// Delete soft deletes the customer.
func (c *Customer) Delete(deletedBy uuid.UUID) {
	c.MarkDeleted()
	c.AuditInfo.SetDeletedBy(deletedBy)
	c.IncrementVersion()

	c.AddDomainEvent(NewCustomerDeletedEvent(c))
}

// Restore restores a deleted customer.
func (c *Customer) Restore() {
	c.BaseEntity.Restore()
	c.Status = CustomerStatusInactive
	c.IncrementVersion()
}
