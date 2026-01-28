// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Lead errors
var (
	ErrLeadNotFound            = errors.New("lead not found")
	ErrLeadAlreadyExists       = errors.New("lead already exists")
	ErrLeadAlreadyConverted    = errors.New("lead is already converted")
	ErrLeadAlreadyQualified    = errors.New("lead is already qualified")
	ErrLeadAlreadyDisqualified = errors.New("lead is already disqualified")
	ErrLeadNotQualified        = errors.New("lead is not qualified")
	ErrLeadNotConverted        = errors.New("lead is not converted")
	ErrInvalidLeadStatus       = errors.New("invalid lead status")
	ErrInvalidLeadSource       = errors.New("invalid lead source")
	ErrInvalidLeadRating       = errors.New("invalid lead rating")
	ErrLeadVersionMismatch     = errors.New("lead version mismatch")
)

// LeadStatus represents the status of a lead.
type LeadStatus string

const (
	LeadStatusNew         LeadStatus = "new"
	LeadStatusContacted   LeadStatus = "contacted"
	LeadStatusQualified   LeadStatus = "qualified"
	LeadStatusUnqualified LeadStatus = "unqualified"
	LeadStatusConverted   LeadStatus = "converted"
	LeadStatusNurturing   LeadStatus = "nurturing"
)

// ValidLeadStatuses returns all valid lead statuses.
func ValidLeadStatuses() []LeadStatus {
	return []LeadStatus{
		LeadStatusNew,
		LeadStatusContacted,
		LeadStatusQualified,
		LeadStatusUnqualified,
		LeadStatusConverted,
		LeadStatusNurturing,
	}
}

// IsValid checks if the lead status is valid.
func (s LeadStatus) IsValid() bool {
	for _, valid := range ValidLeadStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// CanConvert checks if lead with this status can be converted.
func (s LeadStatus) CanConvert() bool {
	return s == LeadStatusQualified
}

// LeadSource represents the source of a lead.
type LeadSource string

const (
	LeadSourceWebsite     LeadSource = "website"
	LeadSourceReferral    LeadSource = "referral"
	LeadSourceSocialMedia LeadSource = "social_media"
	LeadSourceAdvertising LeadSource = "advertising"
	LeadSourceTradeShow   LeadSource = "trade_show"
	LeadSourceColdCall    LeadSource = "cold_call"
	LeadSourceEmail       LeadSource = "email"
	LeadSourcePartner     LeadSource = "partner"
	LeadSourceOther       LeadSource = "other"
)

// ValidLeadSources returns all valid lead sources.
func ValidLeadSources() []LeadSource {
	return []LeadSource{
		LeadSourceWebsite,
		LeadSourceReferral,
		LeadSourceSocialMedia,
		LeadSourceAdvertising,
		LeadSourceTradeShow,
		LeadSourceColdCall,
		LeadSourceEmail,
		LeadSourcePartner,
		LeadSourceOther,
	}
}

// IsValid checks if the lead source is valid.
func (s LeadSource) IsValid() bool {
	for _, valid := range ValidLeadSources() {
		if s == valid {
			return true
		}
	}
	return false
}

// LeadRating represents the rating/quality of a lead.
type LeadRating string

const (
	LeadRatingHot  LeadRating = "hot"
	LeadRatingWarm LeadRating = "warm"
	LeadRatingCold LeadRating = "cold"
)

// ValidLeadRatings returns all valid lead ratings.
func ValidLeadRatings() []LeadRating {
	return []LeadRating{
		LeadRatingHot,
		LeadRatingWarm,
		LeadRatingCold,
	}
}

// IsValid checks if the lead rating is valid.
func (r LeadRating) IsValid() bool {
	for _, valid := range ValidLeadRatings() {
		if r == valid {
			return true
		}
	}
	return false
}

// LeadContact represents contact information for a lead.
type LeadContact struct {
	FirstName  string `json:"first_name" bson:"first_name"`
	LastName   string `json:"last_name" bson:"last_name"`
	Email      string `json:"email" bson:"email"`
	Phone      string `json:"phone,omitempty" bson:"phone,omitempty"`
	Mobile     string `json:"mobile,omitempty" bson:"mobile,omitempty"`
	JobTitle   string `json:"job_title,omitempty" bson:"job_title,omitempty"`
	Department string `json:"department,omitempty" bson:"department,omitempty"`
	LinkedIn   string `json:"linkedin,omitempty" bson:"linkedin,omitempty"`
	Twitter    string `json:"twitter,omitempty" bson:"twitter,omitempty"`
}

// FullName returns the full name of the contact.
func (c LeadContact) FullName() string {
	if c.LastName == "" {
		return c.FirstName
	}
	return c.FirstName + " " + c.LastName
}

// LeadCompany represents company information for a lead.
type LeadCompany struct {
	Name       string `json:"name" bson:"name"`
	Website    string `json:"website,omitempty" bson:"website,omitempty"`
	Industry   string `json:"industry,omitempty" bson:"industry,omitempty"`
	Size       string `json:"size,omitempty" bson:"size,omitempty"` // e.g., "1-10", "11-50", "51-200"
	Revenue    string `json:"revenue,omitempty" bson:"revenue,omitempty"`
	Address    string `json:"address,omitempty" bson:"address,omitempty"`
	City       string `json:"city,omitempty" bson:"city,omitempty"`
	State      string `json:"state,omitempty" bson:"state,omitempty"`
	Country    string `json:"country,omitempty" bson:"country,omitempty"`
	PostalCode string `json:"postal_code,omitempty" bson:"postal_code,omitempty"`
}

// LeadScore represents the scoring for a lead.
type LeadScore struct {
	Score          int          `json:"score" bson:"score"` // 0-100
	Demographic    int          `json:"demographic" bson:"demographic"`
	Behavioral     int          `json:"behavioral" bson:"behavioral"`
	LastCalculated time.Time    `json:"last_calculated" bson:"last_calculated"`
	ScoreHistory   []ScoreEntry `json:"score_history,omitempty" bson:"score_history,omitempty"`
}

// ScoreEntry represents a historical score entry.
type ScoreEntry struct {
	Score     int       `json:"score" bson:"score"`
	Reason    string    `json:"reason" bson:"reason"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

// LeadEngagement tracks lead engagement metrics.
type LeadEngagement struct {
	EmailsOpened    int        `json:"emails_opened" bson:"emails_opened"`
	EmailsClicked   int        `json:"emails_clicked" bson:"emails_clicked"`
	WebVisits       int        `json:"web_visits" bson:"web_visits"`
	FormSubmissions int        `json:"form_submissions" bson:"form_submissions"`
	LastEngagement  *time.Time `json:"last_engagement,omitempty" bson:"last_engagement,omitempty"`
}

// Lead represents a sales lead (potential customer).
type Lead struct {
	ID              uuid.UUID              `json:"id" bson:"_id"`
	TenantID        uuid.UUID              `json:"tenant_id" bson:"tenant_id"`
	Code            string                 `json:"code" bson:"code"` // e.g., "LD-2024-001"
	Status          LeadStatus             `json:"status" bson:"status"`
	Source          LeadSource             `json:"source" bson:"source"`
	Rating          LeadRating             `json:"rating" bson:"rating"`
	Contact         LeadContact            `json:"contact" bson:"contact"`
	Company         LeadCompany            `json:"company" bson:"company"`
	EstimatedValue  Money                  `json:"estimated_value" bson:"estimated_value"`
	Score           LeadScore              `json:"score" bson:"score"`
	Engagement      LeadEngagement         `json:"engagement" bson:"engagement"`
	Campaign        string                 `json:"campaign,omitempty" bson:"campaign,omitempty"`
	CampaignID      *uuid.UUID             `json:"campaign_id,omitempty" bson:"campaign_id,omitempty"`
	OwnerID         *uuid.UUID             `json:"owner_id,omitempty" bson:"owner_id,omitempty"`
	OwnerName       string                 `json:"owner_name,omitempty" bson:"owner_name,omitempty"`
	Tags            []string               `json:"tags,omitempty" bson:"tags,omitempty"`
	Description     string                 `json:"description,omitempty" bson:"description,omitempty"`
	Notes           string                 `json:"notes,omitempty" bson:"notes,omitempty"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
	ConversionInfo  *ConversionInfo        `json:"conversion_info,omitempty" bson:"conversion_info,omitempty"`
	DisqualifyInfo  *DisqualifyInfo        `json:"disqualify_info,omitempty" bson:"disqualify_info,omitempty"`
	NextFollowUp    *time.Time             `json:"next_follow_up,omitempty" bson:"next_follow_up,omitempty"`
	LastContactedAt *time.Time             `json:"last_contacted_at,omitempty" bson:"last_contacted_at,omitempty"`
	CreatedBy       uuid.UUID              `json:"created_by" bson:"created_by"`
	CreatedAt       time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" bson:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Version         int                    `json:"version" bson:"version"`

	// Domain events
	events []DomainEvent `json:"-" bson:"-"`
}

// ConversionInfo contains information about lead conversion.
type ConversionInfo struct {
	ConvertedAt   time.Time  `json:"converted_at" bson:"converted_at"`
	ConvertedBy   uuid.UUID  `json:"converted_by" bson:"converted_by"`
	OpportunityID uuid.UUID  `json:"opportunity_id" bson:"opportunity_id"`
	CustomerID    *uuid.UUID `json:"customer_id,omitempty" bson:"customer_id,omitempty"`
	ContactID     *uuid.UUID `json:"contact_id,omitempty" bson:"contact_id,omitempty"`
}

// DisqualifyInfo contains information about lead disqualification.
type DisqualifyInfo struct {
	DisqualifiedAt time.Time `json:"disqualified_at" bson:"disqualified_at"`
	DisqualifiedBy uuid.UUID `json:"disqualified_by" bson:"disqualified_by"`
	Reason         string    `json:"reason" bson:"reason"`
	Notes          string    `json:"notes,omitempty" bson:"notes,omitempty"`
}

// NewLead creates a new lead.
func NewLead(tenantID uuid.UUID, contact LeadContact, company LeadCompany, source LeadSource, createdBy uuid.UUID) (*Lead, error) {
	if contact.FirstName == "" {
		return nil, errors.New("contact first name is required")
	}
	if contact.Email == "" {
		return nil, errors.New("contact email is required")
	}
	if company.Name == "" {
		return nil, errors.New("company name is required")
	}
	if !source.IsValid() {
		source = LeadSourceOther
	}

	now := time.Now().UTC()
	lead := &Lead{
		ID:       uuid.New(),
		TenantID: tenantID,
		Status:   LeadStatusNew,
		Source:   source,
		Rating:   LeadRatingCold,
		Contact:  contact,
		Company:  company,
		Score: LeadScore{
			Score:          0,
			LastCalculated: now,
		},
		Engagement: LeadEngagement{},
		CreatedBy:  createdBy,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
		events:     make([]DomainEvent, 0),
	}

	lead.AddEvent(NewLeadCreatedEvent(lead))
	return lead, nil
}

// Update updates lead details.
func (l *Lead) Update(contact LeadContact, company LeadCompany, source LeadSource, rating LeadRating) {
	l.Contact = contact
	l.Company = company
	if source.IsValid() {
		l.Source = source
	}
	if rating.IsValid() {
		l.Rating = rating
	}
	l.UpdatedAt = time.Now().UTC()

	l.AddEvent(NewLeadUpdatedEvent(l))
}

// SetEstimatedValue sets the estimated deal value.
func (l *Lead) SetEstimatedValue(value Money) {
	l.EstimatedValue = value
	l.UpdatedAt = time.Now().UTC()
}

// AssignOwner assigns an owner to the lead.
func (l *Lead) AssignOwner(ownerID uuid.UUID, ownerName string) {
	oldOwnerID := l.OwnerID
	l.OwnerID = &ownerID
	l.OwnerName = ownerName
	l.UpdatedAt = time.Now().UTC()

	l.AddEvent(NewLeadOwnerAssignedEvent(l, oldOwnerID))
}

// UnassignOwner removes the owner from the lead.
func (l *Lead) UnassignOwner() {
	l.OwnerID = nil
	l.OwnerName = ""
	l.UpdatedAt = time.Now().UTC()
}

// MarkContacted marks the lead as contacted.
func (l *Lead) MarkContacted() error {
	if l.Status == LeadStatusConverted {
		return ErrLeadAlreadyConverted
	}

	now := time.Now().UTC()
	l.Status = LeadStatusContacted
	l.LastContactedAt = &now
	l.UpdatedAt = now

	l.AddEvent(NewLeadContactedEvent(l))
	return nil
}

// Qualify qualifies the lead.
func (l *Lead) Qualify() error {
	if l.Status == LeadStatusConverted {
		return ErrLeadAlreadyConverted
	}
	if l.Status == LeadStatusQualified {
		return ErrLeadAlreadyQualified
	}

	l.Status = LeadStatusQualified
	l.UpdatedAt = time.Now().UTC()

	l.AddEvent(NewLeadQualifiedEvent(l))
	return nil
}

// Disqualify disqualifies the lead.
func (l *Lead) Disqualify(reason, notes string, disqualifiedBy uuid.UUID) error {
	if l.Status == LeadStatusConverted {
		return ErrLeadAlreadyConverted
	}
	if l.Status == LeadStatusUnqualified {
		return ErrLeadAlreadyDisqualified
	}

	now := time.Now().UTC()
	l.Status = LeadStatusUnqualified
	l.DisqualifyInfo = &DisqualifyInfo{
		DisqualifiedAt: now,
		DisqualifiedBy: disqualifiedBy,
		Reason:         reason,
		Notes:          notes,
	}
	l.UpdatedAt = now

	l.AddEvent(NewLeadDisqualifiedEvent(l, reason))
	return nil
}

// StartNurturing puts the lead in nurturing status.
func (l *Lead) StartNurturing() error {
	if l.Status == LeadStatusConverted {
		return ErrLeadAlreadyConverted
	}

	l.Status = LeadStatusNurturing
	l.UpdatedAt = time.Now().UTC()
	return nil
}

// ConvertToOpportunity converts the lead to an opportunity.
func (l *Lead) ConvertToOpportunity(opportunityID uuid.UUID, convertedBy uuid.UUID, customerID, contactID *uuid.UUID) error {
	if l.Status == LeadStatusConverted {
		return ErrLeadAlreadyConverted
	}
	if l.Status != LeadStatusQualified {
		return ErrLeadNotQualified
	}

	now := time.Now().UTC()
	l.Status = LeadStatusConverted
	l.ConversionInfo = &ConversionInfo{
		ConvertedAt:   now,
		ConvertedBy:   convertedBy,
		OpportunityID: opportunityID,
		CustomerID:    customerID,
		ContactID:     contactID,
	}
	l.UpdatedAt = now

	l.AddEvent(NewLeadConvertedEvent(l, opportunityID))
	return nil
}

// RevertConversion reverts a lead from converted status back to qualified.
// This is used for saga compensation when the conversion workflow fails.
func (l *Lead) RevertConversion() error {
	if l.Status != LeadStatusConverted {
		return ErrLeadNotConverted
	}

	now := time.Now().UTC()
	l.Status = LeadStatusQualified
	l.ConversionInfo = nil
	l.UpdatedAt = now

	// Note: We don't emit an event here as the original conversion event
	// should be handled by the saga's event compensation mechanism
	return nil
}

// UpdateScore updates the lead score.
func (l *Lead) UpdateScore(demographic, behavioral int, reason string) {
	score := demographic + behavioral
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	now := time.Now().UTC()

	// Add to history
	l.Score.ScoreHistory = append(l.Score.ScoreHistory, ScoreEntry{
		Score:     score,
		Reason:    reason,
		Timestamp: now,
	})

	// Keep only last 10 entries
	if len(l.Score.ScoreHistory) > 10 {
		l.Score.ScoreHistory = l.Score.ScoreHistory[len(l.Score.ScoreHistory)-10:]
	}

	l.Score.Score = score
	l.Score.Demographic = demographic
	l.Score.Behavioral = behavioral
	l.Score.LastCalculated = now
	l.UpdatedAt = now

	// Update rating based on score
	if score >= 70 {
		l.Rating = LeadRatingHot
	} else if score >= 40 {
		l.Rating = LeadRatingWarm
	} else {
		l.Rating = LeadRatingCold
	}

	l.AddEvent(NewLeadScoredEvent(l))
}

// RecordEngagement records an engagement activity.
func (l *Lead) RecordEngagement(engagementType string) {
	now := time.Now().UTC()
	l.Engagement.LastEngagement = &now
	l.UpdatedAt = now

	switch engagementType {
	case "email_opened":
		l.Engagement.EmailsOpened++
	case "email_clicked":
		l.Engagement.EmailsClicked++
	case "web_visit":
		l.Engagement.WebVisits++
	case "form_submission":
		l.Engagement.FormSubmissions++
	}
}

// SetNextFollowUp sets the next follow-up date.
func (l *Lead) SetNextFollowUp(followUp time.Time) {
	l.NextFollowUp = &followUp
	l.UpdatedAt = time.Now().UTC()
}

// ClearNextFollowUp clears the next follow-up date.
func (l *Lead) ClearNextFollowUp() {
	l.NextFollowUp = nil
	l.UpdatedAt = time.Now().UTC()
}

// AddTag adds a tag to the lead.
func (l *Lead) AddTag(tag string) {
	for _, t := range l.Tags {
		if t == tag {
			return
		}
	}
	l.Tags = append(l.Tags, tag)
	l.UpdatedAt = time.Now().UTC()
}

// RemoveTag removes a tag from the lead.
func (l *Lead) RemoveTag(tag string) {
	for i, t := range l.Tags {
		if t == tag {
			l.Tags = append(l.Tags[:i], l.Tags[i+1:]...)
			l.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// SetCustomField sets a custom field value.
func (l *Lead) SetCustomField(key string, value interface{}) {
	if l.CustomFields == nil {
		l.CustomFields = make(map[string]interface{})
	}
	l.CustomFields[key] = value
	l.UpdatedAt = time.Now().UTC()
}

// Delete soft deletes the lead.
func (l *Lead) Delete() error {
	if l.Status == LeadStatusConverted {
		return ErrLeadAlreadyConverted
	}

	now := time.Now().UTC()
	l.DeletedAt = &now
	l.UpdatedAt = now

	l.AddEvent(NewLeadDeletedEvent(l))
	return nil
}

// Restore restores a soft-deleted lead.
func (l *Lead) Restore() {
	l.DeletedAt = nil
	l.UpdatedAt = time.Now().UTC()
}

// IsDeleted returns true if the lead is deleted.
func (l *Lead) IsDeleted() bool {
	return l.DeletedAt != nil
}

// IsConverted returns true if the lead is converted.
func (l *Lead) IsConverted() bool {
	return l.Status == LeadStatusConverted
}

// IsQualified returns true if the lead is qualified.
func (l *Lead) IsQualified() bool {
	return l.Status == LeadStatusQualified
}

// CanConvert returns true if the lead can be converted.
func (l *Lead) CanConvert() bool {
	return l.Status.CanConvert() && !l.IsDeleted()
}

// DaysInCurrentStatus returns the number of days in current status.
func (l *Lead) DaysInCurrentStatus() int {
	return int(time.Since(l.UpdatedAt).Hours() / 24)
}

// DaysSinceLastContact returns the number of days since last contact.
func (l *Lead) DaysSinceLastContact() int {
	if l.LastContactedAt == nil {
		return int(time.Since(l.CreatedAt).Hours() / 24)
	}
	return int(time.Since(*l.LastContactedAt).Hours() / 24)
}

// AddEvent adds a domain event.
func (l *Lead) AddEvent(event DomainEvent) {
	l.events = append(l.events, event)
}

// GetEvents returns all domain events.
func (l *Lead) GetEvents() []DomainEvent {
	return l.events
}

// ClearEvents clears all domain events.
func (l *Lead) ClearEvents() {
	l.events = make([]DomainEvent, 0)
}
