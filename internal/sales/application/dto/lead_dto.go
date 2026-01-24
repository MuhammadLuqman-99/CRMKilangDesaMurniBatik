package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Lead Request DTOs
// ============================================================================

// CreateLeadRequest represents a request to create a new lead.
type CreateLeadRequest struct {
	// Basic Information
	FirstName   string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName    string  `json:"last_name" validate:"required,min=1,max=100"`
	Email       string  `json:"email" validate:"required,email,max=255"`
	Phone       *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Mobile      *string `json:"mobile,omitempty" validate:"omitempty,max=50"`
	JobTitle    *string `json:"job_title,omitempty" validate:"omitempty,max=100"`
	Department  *string `json:"department,omitempty" validate:"omitempty,max=100"`

	// Company Information
	Company         *string `json:"company,omitempty" validate:"omitempty,max=200"`
	CompanySize     *string `json:"company_size,omitempty" validate:"omitempty,oneof=1-10 11-50 51-200 201-500 501-1000 1001-5000 5000+"`
	Industry        *string `json:"industry,omitempty" validate:"omitempty,max=100"`
	Website         *string `json:"website,omitempty" validate:"omitempty,url,max=500"`
	AnnualRevenue   *int64  `json:"annual_revenue,omitempty" validate:"omitempty,min=0"`
	NumberEmployees *int    `json:"number_employees,omitempty" validate:"omitempty,min=0"`

	// Address
	Address *AddressDTO `json:"address,omitempty"`

	// Lead Source and Attribution
	Source         string  `json:"source" validate:"required,oneof=website referral social_media email_campaign cold_call trade_show advertisement partner other"`
	SourceDetails  *string `json:"source_details,omitempty" validate:"omitempty,max=500"`
	CampaignID     *string `json:"campaign_id,omitempty" validate:"omitempty,uuid"`
	ReferralSource *string `json:"referral_source,omitempty" validate:"omitempty,max=200"`
	UTMSource      *string `json:"utm_source,omitempty" validate:"omitempty,max=100"`
	UTMMedium      *string `json:"utm_medium,omitempty" validate:"omitempty,max=100"`
	UTMCampaign    *string `json:"utm_campaign,omitempty" validate:"omitempty,max=100"`
	UTMTerm        *string `json:"utm_term,omitempty" validate:"omitempty,max=100"`
	UTMContent     *string `json:"utm_content,omitempty" validate:"omitempty,max=100"`

	// Assignment
	OwnerID *string `json:"owner_id,omitempty" validate:"omitempty,uuid"`

	// Additional Information
	Description *string           `json:"description,omitempty" validate:"omitempty,max=5000"`
	Tags        []string          `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`

	// Interest and Requirements
	ProductInterest []string `json:"product_interest,omitempty" validate:"omitempty,max=10,dive,max=100"`
	Budget          *int64   `json:"budget,omitempty" validate:"omitempty,min=0"`
	BudgetCurrency  *string  `json:"budget_currency,omitempty" validate:"omitempty,len=3"`
	Timeline        *string  `json:"timeline,omitempty" validate:"omitempty,oneof=immediate 1_month 3_months 6_months 1_year no_timeline"`
	Requirements    *string  `json:"requirements,omitempty" validate:"omitempty,max=2000"`

	// Consent
	MarketingConsent *bool `json:"marketing_consent,omitempty"`
	PrivacyConsent   *bool `json:"privacy_consent,omitempty"`
}

// UpdateLeadRequest represents a request to update an existing lead.
type UpdateLeadRequest struct {
	// Basic Information
	FirstName   *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName    *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Email       *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Phone       *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Mobile      *string `json:"mobile,omitempty" validate:"omitempty,max=50"`
	JobTitle    *string `json:"job_title,omitempty" validate:"omitempty,max=100"`
	Department  *string `json:"department,omitempty" validate:"omitempty,max=100"`

	// Company Information
	Company         *string `json:"company,omitempty" validate:"omitempty,max=200"`
	CompanySize     *string `json:"company_size,omitempty" validate:"omitempty,oneof=1-10 11-50 51-200 201-500 501-1000 1001-5000 5000+"`
	Industry        *string `json:"industry,omitempty" validate:"omitempty,max=100"`
	Website         *string `json:"website,omitempty" validate:"omitempty,url,max=500"`
	AnnualRevenue   *int64  `json:"annual_revenue,omitempty" validate:"omitempty,min=0"`
	NumberEmployees *int    `json:"number_employees,omitempty" validate:"omitempty,min=0"`

	// Address
	Address *AddressDTO `json:"address,omitempty"`

	// Lead Source and Attribution
	SourceDetails  *string `json:"source_details,omitempty" validate:"omitempty,max=500"`
	ReferralSource *string `json:"referral_source,omitempty" validate:"omitempty,max=200"`

	// Additional Information
	Description  *string                `json:"description,omitempty" validate:"omitempty,max=5000"`
	Tags         []string               `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`

	// Interest and Requirements
	ProductInterest []string `json:"product_interest,omitempty" validate:"omitempty,max=10,dive,max=100"`
	Budget          *int64   `json:"budget,omitempty" validate:"omitempty,min=0"`
	BudgetCurrency  *string  `json:"budget_currency,omitempty" validate:"omitempty,len=3"`
	Timeline        *string  `json:"timeline,omitempty" validate:"omitempty,oneof=immediate 1_month 3_months 6_months 1_year no_timeline"`
	Requirements    *string  `json:"requirements,omitempty" validate:"omitempty,max=2000"`

	// Consent
	MarketingConsent *bool `json:"marketing_consent,omitempty"`
	PrivacyConsent   *bool `json:"privacy_consent,omitempty"`

	// Version for optimistic locking
	Version int `json:"version" validate:"required,min=1"`
}

// AssignLeadRequest represents a request to assign a lead to an owner.
type AssignLeadRequest struct {
	OwnerID string `json:"owner_id" validate:"required,uuid"`
	Notes   string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// QualifyLeadRequest represents a request to qualify a lead.
type QualifyLeadRequest struct {
	// BANT Criteria
	Budget        *int64  `json:"budget,omitempty" validate:"omitempty,min=0"`
	BudgetCurrency *string `json:"budget_currency,omitempty" validate:"omitempty,len=3"`
	Authority     *string `json:"authority,omitempty" validate:"omitempty,oneof=decision_maker influencer evaluator user other"`
	Need          *string `json:"need,omitempty" validate:"omitempty,max=1000"`
	Timeline      *string `json:"timeline,omitempty" validate:"omitempty,oneof=immediate 1_month 3_months 6_months 1_year no_timeline"`

	// Qualification Notes
	Notes string `json:"notes,omitempty" validate:"omitempty,max=2000"`

	// Custom qualification criteria
	QualificationCriteria map[string]interface{} `json:"qualification_criteria,omitempty"`
}

// DisqualifyLeadRequest represents a request to disqualify a lead.
type DisqualifyLeadRequest struct {
	Reason string `json:"reason" validate:"required,oneof=no_budget no_authority no_need no_timeline competitor duplicate invalid_contact unresponsive other"`
	Notes  string `json:"notes,omitempty" validate:"omitempty,max=2000"`
}

// ConvertLeadRequest represents a request to convert a lead to an opportunity.
type ConvertLeadRequest struct {
	// Opportunity Details
	OpportunityName     string  `json:"opportunity_name" validate:"required,min=1,max=200"`
	PipelineID          string  `json:"pipeline_id" validate:"required,uuid"`
	StageID             *string `json:"stage_id,omitempty" validate:"omitempty,uuid"`
	ExpectedAmount      *int64  `json:"expected_amount,omitempty" validate:"omitempty,min=0"`
	Currency            *string `json:"currency,omitempty" validate:"omitempty,len=3"`
	ExpectedCloseDate   *string `json:"expected_close_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Probability         *int    `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`

	// Customer Linking
	CustomerID          *string `json:"customer_id,omitempty" validate:"omitempty,uuid"`
	CreateNewCustomer   bool    `json:"create_new_customer"`
	CustomerName        *string `json:"customer_name,omitempty" validate:"omitempty,max=200"`

	// Contact Linking
	ContactID           *string `json:"contact_id,omitempty" validate:"omitempty,uuid"`
	CreateNewContact    bool    `json:"create_new_contact"`

	// Assignment
	OwnerID             *string `json:"owner_id,omitempty" validate:"omitempty,uuid"`

	// Additional Details
	Description         *string `json:"description,omitempty" validate:"omitempty,max=5000"`
	Source              *string `json:"source,omitempty" validate:"omitempty,max=100"`

	// Notes
	ConversionNotes     string  `json:"conversion_notes,omitempty" validate:"omitempty,max=2000"`
}

// ScoreLeadRequest represents a request to update lead scoring.
type ScoreLeadRequest struct {
	// Demographic Score Adjustments
	DemographicScore *int `json:"demographic_score,omitempty" validate:"omitempty,min=0,max=100"`

	// Behavioral Score Adjustments
	BehavioralScore *int `json:"behavioral_score,omitempty" validate:"omitempty,min=0,max=100"`

	// Score Reason
	Reason string `json:"reason,omitempty" validate:"omitempty,max=500"`

	// Activity-based scoring
	Activity *LeadActivityScoreDTO `json:"activity,omitempty"`
}

// LeadActivityScoreDTO represents an activity that affects lead scoring.
type LeadActivityScoreDTO struct {
	Type        string `json:"type" validate:"required,oneof=email_open email_click website_visit form_submit content_download webinar_attend meeting_scheduled demo_requested pricing_viewed trial_started"`
	Score       int    `json:"score" validate:"required,min=-50,max=50"`
	Description string `json:"description,omitempty" validate:"omitempty,max=200"`
}

// NurtureLeadRequest represents a request to move a lead to nurturing.
type NurtureLeadRequest struct {
	NurtureCampaignID *string `json:"nurture_campaign_id,omitempty" validate:"omitempty,uuid"`
	Reason            string  `json:"reason,omitempty" validate:"omitempty,max=500"`
	ReengageDate      *string `json:"reengage_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// BulkAssignLeadsRequest represents a request to bulk assign leads.
type BulkAssignLeadsRequest struct {
	LeadIDs []string `json:"lead_ids" validate:"required,min=1,max=100,dive,uuid"`
	OwnerID string   `json:"owner_id" validate:"required,uuid"`
	Notes   string   `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// BulkUpdateLeadStatusRequest represents a request to bulk update lead statuses.
type BulkUpdateLeadStatusRequest struct {
	LeadIDs []string `json:"lead_ids" validate:"required,min=1,max=100,dive,uuid"`
	Status  string   `json:"status" validate:"required,oneof=new contacted qualified unqualified converted nurturing"`
	Notes   string   `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// LeadFilterRequest represents filter options for listing leads.
type LeadFilterRequest struct {
	// Status filters
	Statuses []string `json:"statuses,omitempty" validate:"omitempty,dive,oneof=new contacted qualified unqualified converted nurturing"`

	// Source filters
	Sources []string `json:"sources,omitempty" validate:"omitempty,dive,oneof=website referral social_media email_campaign cold_call trade_show advertisement partner other"`

	// Assignment filters
	OwnerIDs   []string `json:"owner_ids,omitempty" validate:"omitempty,dive,uuid"`
	Unassigned *bool    `json:"unassigned,omitempty"`

	// Score filters
	MinScore *int `json:"min_score,omitempty" validate:"omitempty,min=0,max=100"`
	MaxScore *int `json:"max_score,omitempty" validate:"omitempty,min=0,max=100"`

	// Time filters
	CreatedAfter  *string `json:"created_after,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	CreatedBefore *string `json:"created_before,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	UpdatedAfter  *string `json:"updated_after,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`

	// Search
	SearchQuery string `json:"search_query,omitempty" validate:"omitempty,max=200"`

	// Tags
	Tags []string `json:"tags,omitempty" validate:"omitempty,max=10,dive,max=50"`

	// Campaign
	CampaignID *string `json:"campaign_id,omitempty" validate:"omitempty,uuid"`

	// Company filters
	Companies  []string `json:"companies,omitempty" validate:"omitempty,max=20,dive,max=200"`
	Industries []string `json:"industries,omitempty" validate:"omitempty,max=20,dive,max=100"`

	// Pagination
	Page     int `json:"page,omitempty" validate:"omitempty,min=1"`
	PageSize int `json:"page_size,omitempty" validate:"omitempty,min=1,max=100"`

	// Sorting
	SortBy    string `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at updated_at score first_name last_name company"`
	SortOrder string `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// ============================================================================
// Lead Response DTOs
// ============================================================================

// LeadResponse represents a lead in API responses.
type LeadResponse struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`

	// Basic Information
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	FullName    string  `json:"full_name"`
	Email       string  `json:"email"`
	Phone       *string `json:"phone,omitempty"`
	Mobile      *string `json:"mobile,omitempty"`
	JobTitle    *string `json:"job_title,omitempty"`
	Department  *string `json:"department,omitempty"`

	// Company Information
	Company         *string `json:"company,omitempty"`
	CompanySize     *string `json:"company_size,omitempty"`
	Industry        *string `json:"industry,omitempty"`
	Website         *string `json:"website,omitempty"`
	AnnualRevenue   *int64  `json:"annual_revenue,omitempty"`
	NumberEmployees *int    `json:"number_employees,omitempty"`

	// Address
	Address *AddressDTO `json:"address,omitempty"`

	// Status and Scoring
	Status           string `json:"status"`
	Score            int    `json:"score"`
	DemographicScore int    `json:"demographic_score"`
	BehavioralScore  int    `json:"behavioral_score"`
	Rating           string `json:"rating"` // hot, warm, cold

	// Lead Source and Attribution
	Source         string  `json:"source"`
	SourceDetails  *string `json:"source_details,omitempty"`
	CampaignID     *string `json:"campaign_id,omitempty"`
	ReferralSource *string `json:"referral_source,omitempty"`
	UTMParams      *UTMParamsDTO `json:"utm_params,omitempty"`

	// Assignment
	OwnerID   *string       `json:"owner_id,omitempty"`
	Owner     *UserBriefDTO `json:"owner,omitempty"`

	// Interest and Requirements
	ProductInterest []string `json:"product_interest,omitempty"`
	Budget          *MoneyDTO `json:"budget,omitempty"`
	Timeline        *string   `json:"timeline,omitempty"`
	Requirements    *string   `json:"requirements,omitempty"`

	// Qualification
	QualifiedAt   *time.Time `json:"qualified_at,omitempty"`
	QualifiedBy   *string    `json:"qualified_by,omitempty"`
	QualificationNotes *string `json:"qualification_notes,omitempty"`

	// Disqualification
	DisqualifiedAt    *time.Time `json:"disqualified_at,omitempty"`
	DisqualifiedBy    *string    `json:"disqualified_by,omitempty"`
	DisqualifyReason  *string    `json:"disqualify_reason,omitempty"`

	// Conversion
	ConvertedAt     *time.Time `json:"converted_at,omitempty"`
	ConvertedBy     *string    `json:"converted_by,omitempty"`
	OpportunityID   *string    `json:"opportunity_id,omitempty"`
	CustomerID      *string    `json:"customer_id,omitempty"`
	ContactID       *string    `json:"contact_id,omitempty"`

	// Nurturing
	NurturingAt       *time.Time `json:"nurturing_at,omitempty"`
	NurtureCampaignID *string    `json:"nurture_campaign_id,omitempty"`

	// Additional Information
	Description  *string                `json:"description,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`

	// Consent
	MarketingConsent bool `json:"marketing_consent"`
	PrivacyConsent   bool `json:"privacy_consent"`

	// Activity Summary
	LastActivityAt    *time.Time `json:"last_activity_at,omitempty"`
	LastContactedAt   *time.Time `json:"last_contacted_at,omitempty"`
	ActivityCount     int        `json:"activity_count"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy string     `json:"created_by"`
	UpdatedBy string     `json:"updated_by"`
	Version   int        `json:"version"`
}

// LeadBriefResponse represents a brief lead summary.
type LeadBriefResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	FullName  string  `json:"full_name"`
	Email     string  `json:"email"`
	Company   *string `json:"company,omitempty"`
	Status    string  `json:"status"`
	Score     int     `json:"score"`
	Rating    string  `json:"rating"`
	OwnerID   *string `json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// LeadListResponse represents a paginated list of leads.
type LeadListResponse struct {
	Leads      []*LeadBriefResponse `json:"leads"`
	Pagination PaginationResponse   `json:"pagination"`
}

// LeadConversionResponse represents the result of converting a lead.
type LeadConversionResponse struct {
	LeadID        string `json:"lead_id"`
	OpportunityID string `json:"opportunity_id"`
	CustomerID    *string `json:"customer_id,omitempty"`
	ContactID     *string `json:"contact_id,omitempty"`
	Message       string `json:"message"`
}

// LeadStatisticsResponse represents lead statistics.
type LeadStatisticsResponse struct {
	TotalLeads     int64            `json:"total_leads"`
	ByStatus       map[string]int64 `json:"by_status"`
	BySource       map[string]int64 `json:"by_source"`
	ByRating       map[string]int64 `json:"by_rating"`
	ConversionRate float64          `json:"conversion_rate"`
	AverageScore   float64          `json:"average_score"`
	NewLeadsToday  int64            `json:"new_leads_today"`
	NewLeadsWeek   int64            `json:"new_leads_week"`
	NewLeadsMonth  int64            `json:"new_leads_month"`
}

// ============================================================================
// Supporting DTOs
// ============================================================================

// AddressDTO represents an address.
type AddressDTO struct {
	Street1    string  `json:"street1,omitempty" validate:"omitempty,max=200"`
	Street2    *string `json:"street2,omitempty" validate:"omitempty,max=200"`
	City       string  `json:"city,omitempty" validate:"omitempty,max=100"`
	State      *string `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode *string `json:"postal_code,omitempty" validate:"omitempty,max=20"`
	Country    string  `json:"country,omitempty" validate:"omitempty,len=2"` // ISO 3166-1 alpha-2
}

// UTMParamsDTO represents UTM tracking parameters.
type UTMParamsDTO struct {
	Source   *string `json:"source,omitempty"`
	Medium   *string `json:"medium,omitempty"`
	Campaign *string `json:"campaign,omitempty"`
	Term     *string `json:"term,omitempty"`
	Content  *string `json:"content,omitempty"`
}

// MoneyDTO represents a monetary amount.
type MoneyDTO struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Display  string `json:"display"` // Formatted display string
}

// UserBriefDTO represents a brief user summary.
type UserBriefDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// PaginationResponse represents pagination metadata.
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewPaginationResponse creates a new pagination response.
func NewPaginationResponse(page, pageSize int, totalItems int64) PaginationResponse {
	totalPages := int(totalItems) / pageSize
	if int(totalItems)%pageSize > 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// ParseUUID parses a string to UUID, returning nil if empty or invalid.
func ParseUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

// ParseUUIDRequired parses a string to UUID, returning error if invalid.
func ParseUUIDRequired(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// FormatUUID formats a UUID to string, returning empty string if nil.
func FormatUUID(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

// FormatUUIDValue formats a UUID value to string.
func FormatUUIDValue(id uuid.UUID) string {
	return id.String()
}

// StringPtr returns a pointer to the string.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// IntPtr returns a pointer to the int.
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr returns a pointer to the int64.
func Int64Ptr(i int64) *int64 {
	return &i
}

// BoolPtr returns a pointer to the bool.
func BoolPtr(b bool) *bool {
	return &b
}

// TimePtr returns a pointer to the time.
func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
