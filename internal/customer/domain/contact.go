// Package domain contains the domain layer for the Customer service.
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContactStatus represents the status of a contact.
type ContactStatus string

const (
	ContactStatusActive   ContactStatus = "active"
	ContactStatusInactive ContactStatus = "inactive"
	ContactStatusBlocked  ContactStatus = "blocked"
)

// ContactRole represents the role of a contact in an organization.
type ContactRole string

const (
	ContactRoleDecisionMaker ContactRole = "decision_maker"
	ContactRoleInfluencer    ContactRole = "influencer"
	ContactRoleUser          ContactRole = "user"
	ContactRoleTechnical     ContactRole = "technical"
	ContactRoleFinance       ContactRole = "finance"
	ContactRoleProcurement   ContactRole = "procurement"
	ContactRoleExecutive     ContactRole = "executive"
	ContactRoleOther         ContactRole = "other"
)

// CommunicationPreference represents preferred communication channel.
type CommunicationPreference string

const (
	CommPrefEmail    CommunicationPreference = "email"
	CommPrefPhone    CommunicationPreference = "phone"
	CommPrefWhatsApp CommunicationPreference = "whatsapp"
	CommPrefSMS      CommunicationPreference = "sms"
	CommPrefMail     CommunicationPreference = "mail"
	CommPrefAny      CommunicationPreference = "any"
)

// Contact represents a person associated with a customer.
type Contact struct {
	BaseEntity
	CustomerID         uuid.UUID               `json:"customer_id" bson:"customer_id"`
	TenantID           uuid.UUID               `json:"tenant_id" bson:"tenant_id"`
	Name               PersonName              `json:"name" bson:"name"`
	Email              Email                   `json:"email" bson:"email"`
	PhoneNumbers       []PhoneNumber           `json:"phone_numbers" bson:"phone_numbers"`
	Addresses          []Address               `json:"addresses,omitempty" bson:"addresses,omitempty"`
	SocialProfiles     []SocialProfile         `json:"social_profiles,omitempty" bson:"social_profiles,omitempty"`
	JobTitle           string                  `json:"job_title,omitempty" bson:"job_title,omitempty"`
	Department         string                  `json:"department,omitempty" bson:"department,omitempty"`
	Role               ContactRole             `json:"role" bson:"role"`
	Status             ContactStatus           `json:"status" bson:"status"`
	IsPrimary          bool                    `json:"is_primary" bson:"is_primary"`
	ReportsTo          *uuid.UUID              `json:"reports_to,omitempty" bson:"reports_to,omitempty"`
	CommPreference     CommunicationPreference `json:"comm_preference" bson:"comm_preference"`
	OptedOutMarketing  bool                    `json:"opted_out_marketing" bson:"opted_out_marketing"`
	MarketingConsent   *time.Time              `json:"marketing_consent,omitempty" bson:"marketing_consent,omitempty"`
	Birthday           *time.Time              `json:"birthday,omitempty" bson:"birthday,omitempty"`
	Notes              string                  `json:"notes,omitempty" bson:"notes,omitempty"`
	Tags               []string                `json:"tags,omitempty" bson:"tags,omitempty"`
	CustomFields       map[string]interface{}  `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
	AuditInfo          AuditInfo               `json:"audit_info" bson:"audit_info"`
	LastContactedAt    *time.Time              `json:"last_contacted_at,omitempty" bson:"last_contacted_at,omitempty"`
	NextFollowUpAt     *time.Time              `json:"next_follow_up_at,omitempty" bson:"next_follow_up_at,omitempty"`
	EngagementScore    int                     `json:"engagement_score" bson:"engagement_score"`
	LinkedInURL        string                  `json:"linkedin_url,omitempty" bson:"linkedin_url,omitempty"`
	ProfilePhotoURL    string                  `json:"profile_photo_url,omitempty" bson:"profile_photo_url,omitempty"`
}

// ContactBuilder provides a fluent API for building Contact entities.
type ContactBuilder struct {
	contact Contact
	errors  ValidationErrors
}

// NewContactBuilder creates a new ContactBuilder.
func NewContactBuilder(customerID, tenantID uuid.UUID) *ContactBuilder {
	return &ContactBuilder{
		contact: Contact{
			BaseEntity: NewBaseEntity(),
			CustomerID: customerID,
			TenantID:   tenantID,
			Status:     ContactStatusActive,
			Role:       ContactRoleOther,
			CommPreference: CommPrefEmail,
			PhoneNumbers:   make([]PhoneNumber, 0),
			Addresses:      make([]Address, 0),
			SocialProfiles: make([]SocialProfile, 0),
			Tags:           make([]string, 0),
			CustomFields:   make(map[string]interface{}),
		},
	}
}

// WithName sets the contact name.
func (b *ContactBuilder) WithName(firstName, lastName string) *ContactBuilder {
	name, err := NewPersonName(firstName, lastName)
	if err != nil {
		b.errors.Add(err.(*ValidationError))
	} else {
		b.contact.Name = name
	}
	return b
}

// WithFullName sets the contact name with all parts.
func (b *ContactBuilder) WithFullName(title, firstName, middleName, lastName, suffix string) *ContactBuilder {
	name, err := NewPersonName(firstName, lastName)
	if err != nil {
		b.errors.Add(err.(*ValidationError))
	} else {
		b.contact.Name = name.WithTitle(title).WithMiddleName(middleName).WithSuffix(suffix)
	}
	return b
}

// WithEmail sets the contact email.
func (b *ContactBuilder) WithEmail(email string) *ContactBuilder {
	e, err := NewEmail(email)
	if err != nil {
		b.errors.AddField("email", err.Error(), "INVALID_EMAIL")
	} else {
		b.contact.Email = e
	}
	return b
}

// WithPhone adds a phone number.
func (b *ContactBuilder) WithPhone(number string, phoneType PhoneType, isPrimary bool) *ContactBuilder {
	phone, err := NewPhoneNumberWithPrimary(number, phoneType, isPrimary)
	if err != nil {
		b.errors.AddField("phone", err.Error(), "INVALID_PHONE")
	} else {
		b.contact.PhoneNumbers = append(b.contact.PhoneNumbers, phone)
	}
	return b
}

// WithAddress adds an address.
func (b *ContactBuilder) WithAddress(line1, city, postalCode, countryCode string, addressType AddressType, isPrimary bool) *ContactBuilder {
	addr, err := NewAddress(line1, city, postalCode, countryCode, addressType)
	if err != nil {
		if verrs, ok := err.(ValidationErrors); ok {
			for _, e := range verrs {
				b.errors.Add(e)
			}
		}
	} else {
		addr.IsPrimary = isPrimary
		b.contact.Addresses = append(b.contact.Addresses, addr)
	}
	return b
}

// WithJobTitle sets the job title.
func (b *ContactBuilder) WithJobTitle(title string) *ContactBuilder {
	b.contact.JobTitle = strings.TrimSpace(title)
	return b
}

// WithDepartment sets the department.
func (b *ContactBuilder) WithDepartment(department string) *ContactBuilder {
	b.contact.Department = strings.TrimSpace(department)
	return b
}

// WithRole sets the contact role.
func (b *ContactBuilder) WithRole(role ContactRole) *ContactBuilder {
	b.contact.Role = role
	return b
}

// AsPrimary marks this as the primary contact.
func (b *ContactBuilder) AsPrimary() *ContactBuilder {
	b.contact.IsPrimary = true
	return b
}

// WithCommPreference sets communication preference.
func (b *ContactBuilder) WithCommPreference(pref CommunicationPreference) *ContactBuilder {
	b.contact.CommPreference = pref
	return b
}

// WithNotes sets notes.
func (b *ContactBuilder) WithNotes(notes string) *ContactBuilder {
	b.contact.Notes = strings.TrimSpace(notes)
	return b
}

// WithTags sets tags.
func (b *ContactBuilder) WithTags(tags ...string) *ContactBuilder {
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" {
			b.contact.Tags = append(b.contact.Tags, tag)
		}
	}
	return b
}

// WithBirthday sets the birthday.
func (b *ContactBuilder) WithBirthday(birthday time.Time) *ContactBuilder {
	b.contact.Birthday = &birthday
	return b
}

// WithSocialProfile adds a social profile.
func (b *ContactBuilder) WithSocialProfile(platform SocialPlatform, url string) *ContactBuilder {
	profile, err := NewSocialProfile(platform, url)
	if err != nil {
		b.errors.AddField("social_profile", err.Error(), "INVALID_SOCIAL_PROFILE")
	} else {
		b.contact.SocialProfiles = append(b.contact.SocialProfiles, profile)
	}
	return b
}

// WithCustomField adds a custom field.
func (b *ContactBuilder) WithCustomField(key string, value interface{}) *ContactBuilder {
	b.contact.CustomFields[key] = value
	return b
}

// WithCreatedBy sets the creator.
func (b *ContactBuilder) WithCreatedBy(userID uuid.UUID) *ContactBuilder {
	b.contact.AuditInfo.SetCreatedBy(userID)
	return b
}

// Build creates the Contact entity.
func (b *ContactBuilder) Build() (*Contact, error) {
	// Validate required fields
	if b.contact.Name.IsEmpty() {
		b.errors.AddField("name", "name is required", "REQUIRED")
	}

	if b.contact.Email.IsEmpty() && len(b.contact.PhoneNumbers) == 0 {
		b.errors.AddField("contact_info", "at least email or phone number is required", "REQUIRED")
	}

	if b.errors.HasErrors() {
		return nil, b.errors
	}

	return &b.contact, nil
}

// NewContact creates a new Contact with basic information.
func NewContact(customerID, tenantID uuid.UUID, firstName, lastName, email string) (*Contact, error) {
	return NewContactBuilder(customerID, tenantID).
		WithName(firstName, lastName).
		WithEmail(email).
		Build()
}

// ============================================================================
// Contact Behaviors
// ============================================================================

// UpdateName updates the contact name.
func (c *Contact) UpdateName(name PersonName) {
	c.Name = name
	c.MarkUpdated()
}

// UpdateEmail updates the contact email.
func (c *Contact) UpdateEmail(email Email) error {
	c.Email = email
	c.MarkUpdated()
	return nil
}

// AddPhoneNumber adds a phone number to the contact.
func (c *Contact) AddPhoneNumber(phone PhoneNumber) {
	// If this is marked as primary, unmark others
	if phone.IsPrimary() {
		for i := range c.PhoneNumbers {
			c.PhoneNumbers[i].SetPrimary(false)
		}
	}
	c.PhoneNumbers = append(c.PhoneNumbers, phone)
	c.MarkUpdated()
}

// RemovePhoneNumber removes a phone number.
func (c *Contact) RemovePhoneNumber(e164 string) bool {
	for i, phone := range c.PhoneNumbers {
		if phone.E164() == e164 {
			c.PhoneNumbers = append(c.PhoneNumbers[:i], c.PhoneNumbers[i+1:]...)
			c.MarkUpdated()
			return true
		}
	}
	return false
}

// SetPrimaryPhone sets a phone number as primary.
func (c *Contact) SetPrimaryPhone(e164 string) bool {
	found := false
	for i := range c.PhoneNumbers {
		if c.PhoneNumbers[i].E164() == e164 {
			c.PhoneNumbers[i].SetPrimary(true)
			found = true
		} else {
			c.PhoneNumbers[i].SetPrimary(false)
		}
	}
	if found {
		c.MarkUpdated()
	}
	return found
}

// GetPrimaryPhone returns the primary phone number.
func (c *Contact) GetPrimaryPhone() *PhoneNumber {
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

// AddAddress adds an address to the contact.
func (c *Contact) AddAddress(address Address) {
	// If this is marked as primary, unmark others of same type
	if address.IsPrimary {
		for i := range c.Addresses {
			if c.Addresses[i].AddressType == address.AddressType {
				c.Addresses[i].IsPrimary = false
			}
		}
	}
	c.Addresses = append(c.Addresses, address)
	c.MarkUpdated()
}

// RemoveAddress removes an address by matching line1 and postal code.
func (c *Contact) RemoveAddress(line1, postalCode string) bool {
	for i, addr := range c.Addresses {
		if addr.Line1 == line1 && addr.PostalCode == postalCode {
			c.Addresses = append(c.Addresses[:i], c.Addresses[i+1:]...)
			c.MarkUpdated()
			return true
		}
	}
	return false
}

// GetPrimaryAddress returns the primary address of a specific type.
func (c *Contact) GetPrimaryAddress(addressType AddressType) *Address {
	for _, addr := range c.Addresses {
		if addr.AddressType == addressType && addr.IsPrimary {
			return &addr
		}
	}
	return nil
}

// AddSocialProfile adds a social profile.
func (c *Contact) AddSocialProfile(profile SocialProfile) {
	// Remove existing profile for same platform
	for i, p := range c.SocialProfiles {
		if p.Platform == profile.Platform {
			c.SocialProfiles = append(c.SocialProfiles[:i], c.SocialProfiles[i+1:]...)
			break
		}
	}
	c.SocialProfiles = append(c.SocialProfiles, profile)
	c.MarkUpdated()
}

// RemoveSocialProfile removes a social profile by platform.
func (c *Contact) RemoveSocialProfile(platform SocialPlatform) bool {
	for i, profile := range c.SocialProfiles {
		if profile.Platform == platform {
			c.SocialProfiles = append(c.SocialProfiles[:i], c.SocialProfiles[i+1:]...)
			c.MarkUpdated()
			return true
		}
	}
	return false
}

// GetSocialProfile gets a social profile by platform.
func (c *Contact) GetSocialProfile(platform SocialPlatform) *SocialProfile {
	for _, profile := range c.SocialProfiles {
		if profile.Platform == platform {
			return &profile
		}
	}
	return nil
}

// UpdateJobInfo updates job-related information.
func (c *Contact) UpdateJobInfo(title, department string, role ContactRole) {
	c.JobTitle = strings.TrimSpace(title)
	c.Department = strings.TrimSpace(department)
	c.Role = role
	c.MarkUpdated()
}

// SetPrimary marks this contact as primary.
func (c *Contact) SetPrimary(isPrimary bool) {
	c.IsPrimary = isPrimary
	c.MarkUpdated()
}

// SetReportsTo sets the manager contact.
func (c *Contact) SetReportsTo(managerID *uuid.UUID) {
	c.ReportsTo = managerID
	c.MarkUpdated()
}

// Activate activates the contact.
func (c *Contact) Activate() error {
	if c.Status == ContactStatusBlocked {
		return ErrContactBlocked
	}
	c.Status = ContactStatusActive
	c.MarkUpdated()
	return nil
}

// Deactivate deactivates the contact.
func (c *Contact) Deactivate() {
	c.Status = ContactStatusInactive
	c.MarkUpdated()
}

// Block blocks the contact.
func (c *Contact) Block() {
	c.Status = ContactStatusBlocked
	c.MarkUpdated()
}

// Unblock unblocks the contact.
func (c *Contact) Unblock() {
	c.Status = ContactStatusInactive
	c.MarkUpdated()
}

// SetCommunicationPreference sets the preferred communication channel.
func (c *Contact) SetCommunicationPreference(pref CommunicationPreference) {
	c.CommPreference = pref
	c.MarkUpdated()
}

// OptOutMarketing opts out of marketing communications.
func (c *Contact) OptOutMarketing() {
	c.OptedOutMarketing = true
	c.MarketingConsent = nil
	c.MarkUpdated()
}

// OptInMarketing opts in to marketing communications.
func (c *Contact) OptInMarketing() {
	c.OptedOutMarketing = false
	now := time.Now().UTC()
	c.MarketingConsent = &now
	c.MarkUpdated()
}

// RecordContact records a contact interaction.
func (c *Contact) RecordContact() {
	now := time.Now().UTC()
	c.LastContactedAt = &now
	c.AuditInfo.RecordContact()
	c.MarkUpdated()
}

// SetNextFollowUp schedules the next follow-up.
func (c *Contact) SetNextFollowUp(followUpAt time.Time) {
	c.NextFollowUpAt = &followUpAt
	c.MarkUpdated()
}

// ClearNextFollowUp clears the next follow-up.
func (c *Contact) ClearNextFollowUp() {
	c.NextFollowUpAt = nil
	c.MarkUpdated()
}

// UpdateEngagementScore updates the engagement score.
func (c *Contact) UpdateEngagementScore(score int) {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	c.EngagementScore = score
	c.MarkUpdated()
}

// IncrementEngagement increments the engagement score.
func (c *Contact) IncrementEngagement(points int) {
	c.UpdateEngagementScore(c.EngagementScore + points)
}

// AddTag adds a tag.
func (c *Contact) AddTag(tag string) {
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
}

// RemoveTag removes a tag.
func (c *Contact) RemoveTag(tag string) {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for i, t := range c.Tags {
		if t == tag {
			c.Tags = append(c.Tags[:i], c.Tags[i+1:]...)
			c.MarkUpdated()
			return
		}
	}
}

// HasTag checks if contact has a tag.
func (c *Contact) HasTag(tag string) bool {
	tag = strings.TrimSpace(strings.ToLower(tag))
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// SetCustomField sets a custom field.
func (c *Contact) SetCustomField(key string, value interface{}) {
	if c.CustomFields == nil {
		c.CustomFields = make(map[string]interface{})
	}
	c.CustomFields[key] = value
	c.MarkUpdated()
}

// GetCustomField gets a custom field value.
func (c *Contact) GetCustomField(key string) (interface{}, bool) {
	if c.CustomFields == nil {
		return nil, false
	}
	val, ok := c.CustomFields[key]
	return val, ok
}

// RemoveCustomField removes a custom field.
func (c *Contact) RemoveCustomField(key string) {
	if c.CustomFields != nil {
		delete(c.CustomFields, key)
		c.MarkUpdated()
	}
}

// SetProfilePhoto sets the profile photo URL.
func (c *Contact) SetProfilePhoto(url string) {
	c.ProfilePhotoURL = strings.TrimSpace(url)
	c.MarkUpdated()
}

// SetLinkedIn sets the LinkedIn URL.
func (c *Contact) SetLinkedIn(url string) {
	c.LinkedInURL = strings.TrimSpace(url)
	c.MarkUpdated()
}

// ============================================================================
// Contact Query Helpers
// ============================================================================

// IsActive returns true if the contact is active.
func (c *Contact) IsActive() bool {
	return c.Status == ContactStatusActive && !c.IsDeleted()
}

// IsBlocked returns true if the contact is blocked.
func (c *Contact) IsBlocked() bool {
	return c.Status == ContactStatusBlocked
}

// CanReceiveMarketing returns true if contact can receive marketing.
func (c *Contact) CanReceiveMarketing() bool {
	return c.IsActive() && !c.OptedOutMarketing && !c.Email.IsEmpty()
}

// NeedsFollowUp returns true if follow-up is due.
func (c *Contact) NeedsFollowUp() bool {
	if c.NextFollowUpAt == nil {
		return false
	}
	return time.Now().After(*c.NextFollowUpAt)
}

// DisplayName returns a display-friendly name.
func (c *Contact) DisplayName() string {
	if !c.Name.IsEmpty() {
		return c.Name.DisplayName()
	}
	if !c.Email.IsEmpty() {
		return c.Email.Local()
	}
	return "Unknown Contact"
}

// FullName returns the full name.
func (c *Contact) FullName() string {
	return c.Name.FullName()
}

// GetBestContactMethod returns the best way to contact this person.
func (c *Contact) GetBestContactMethod() string {
	switch c.CommPreference {
	case CommPrefEmail:
		if !c.Email.IsEmpty() {
			return "email:" + c.Email.String()
		}
	case CommPrefPhone, CommPrefSMS:
		if phone := c.GetPrimaryPhone(); phone != nil {
			return "phone:" + phone.E164()
		}
	case CommPrefWhatsApp:
		for _, phone := range c.PhoneNumbers {
			if phone.Type() == PhoneTypeWhatsApp {
				return "whatsapp:" + phone.E164()
			}
		}
		if phone := c.GetPrimaryPhone(); phone != nil {
			return "whatsapp:" + phone.E164()
		}
	}

	// Fallback
	if !c.Email.IsEmpty() {
		return "email:" + c.Email.String()
	}
	if phone := c.GetPrimaryPhone(); phone != nil {
		return "phone:" + phone.E164()
	}
	return ""
}

// ErrContactBlocked is returned when trying to activate a blocked contact.
var ErrContactBlocked = NewDomainError("CONTACT_BLOCKED", "contact is blocked and cannot be activated", nil)
