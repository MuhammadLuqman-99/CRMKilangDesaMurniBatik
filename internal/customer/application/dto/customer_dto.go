// Package dto contains Data Transfer Objects for the Customer application layer.
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Customer DTOs
// ============================================================================

// CreateCustomerRequest represents a request to create a customer.
type CreateCustomerRequest struct {
	Name         string                    `json:"name" validate:"required,min=1,max=255"`
	Type         domain.CustomerType       `json:"type" validate:"required,oneof=individual company partner reseller"`
	Code         string                    `json:"code,omitempty" validate:"omitempty,max=50"`
	Email        string                    `json:"email,omitempty" validate:"omitempty,email"`
	Phone        *PhoneInput               `json:"phone,omitempty"`
	Website      string                    `json:"website,omitempty" validate:"omitempty,url"`
	Address      *AddressInput             `json:"address,omitempty"`
	Source       domain.CustomerSource     `json:"source,omitempty" validate:"omitempty,oneof=direct referral website social_media event partner cold_call import other"`
	Tier         domain.CustomerTier       `json:"tier,omitempty" validate:"omitempty,oneof=standard bronze silver gold platinum enterprise"`
	OwnerID      *uuid.UUID                `json:"owner_id,omitempty"`
	CompanyInfo  *CompanyInfoInput         `json:"company_info,omitempty"`
	Preferences  *CustomerPreferencesInput `json:"preferences,omitempty"`
	Tags         []string                  `json:"tags,omitempty" validate:"omitempty,max=50,dive,max=50"`
	Notes        string                    `json:"notes,omitempty" validate:"omitempty,max=5000"`
	CustomFields map[string]interface{}    `json:"custom_fields,omitempty"`
	Contacts     []CreateContactInput      `json:"contacts,omitempty" validate:"omitempty,max=10,dive"`
}

// UpdateCustomerRequest represents a request to update a customer.
type UpdateCustomerRequest struct {
	Name         *string                    `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Type         *domain.CustomerType       `json:"type,omitempty" validate:"omitempty,oneof=individual company partner reseller"`
	Email        *string                    `json:"email,omitempty" validate:"omitempty,email"`
	Phone        *PhoneInput                `json:"phone,omitempty"`
	Website      *string                    `json:"website,omitempty" validate:"omitempty,url"`
	Address      *AddressInput              `json:"address,omitempty"`
	Source       *domain.CustomerSource     `json:"source,omitempty"`
	Tier         *domain.CustomerTier       `json:"tier,omitempty"`
	OwnerID      *uuid.UUID                 `json:"owner_id,omitempty"`
	CompanyInfo  *CompanyInfoInput          `json:"company_info,omitempty"`
	Preferences  *CustomerPreferencesInput  `json:"preferences,omitempty"`
	Financials   *CustomerFinancialsInput   `json:"financials,omitempty"`
	Notes        *string                    `json:"notes,omitempty" validate:"omitempty,max=5000"`
	CustomFields map[string]interface{}     `json:"custom_fields,omitempty"`
	Version      int                        `json:"version" validate:"required,min=1"`
}

// CustomerResponse represents a customer in API responses.
type CustomerResponse struct {
	ID              uuid.UUID                     `json:"id"`
	TenantID        uuid.UUID                     `json:"tenant_id"`
	Code            string                        `json:"code"`
	Name            string                        `json:"name"`
	Type            domain.CustomerType           `json:"type"`
	Status          domain.CustomerStatus         `json:"status"`
	Tier            domain.CustomerTier           `json:"tier"`
	Source          domain.CustomerSource         `json:"source"`
	Email           string                        `json:"email,omitempty"`
	PhoneNumbers    []PhoneResponse               `json:"phone_numbers,omitempty"`
	Website         string                        `json:"website,omitempty"`
	Addresses       []AddressResponse             `json:"addresses,omitempty"`
	SocialProfiles  []SocialProfileResponse       `json:"social_profiles,omitempty"`
	CompanyInfo     *CompanyInfoResponse          `json:"company_info,omitempty"`
	Financials      CustomerFinancialsResponse    `json:"financials"`
	Preferences     CustomerPreferencesResponse   `json:"preferences"`
	Stats           CustomerStatsResponse         `json:"stats"`
	Contacts        []ContactSummaryResponse      `json:"contacts,omitempty"`
	OwnerID         *uuid.UUID                    `json:"owner_id,omitempty"`
	AssignedTeam    []uuid.UUID                   `json:"assigned_team,omitempty"`
	Tags            []string                      `json:"tags,omitempty"`
	Segments        []uuid.UUID                   `json:"segments,omitempty"`
	CustomFields    map[string]interface{}        `json:"custom_fields,omitempty"`
	Notes           string                        `json:"notes,omitempty"`
	LogoURL         string                        `json:"logo_url,omitempty"`
	LastContactedAt *time.Time                    `json:"last_contacted_at,omitempty"`
	NextFollowUpAt  *time.Time                    `json:"next_follow_up_at,omitempty"`
	ConvertedAt     *time.Time                    `json:"converted_at,omitempty"`
	ChurnedAt       *time.Time                    `json:"churned_at,omitempty"`
	ChurnReason     string                        `json:"churn_reason,omitempty"`
	Version         int                           `json:"version"`
	CreatedAt       time.Time                     `json:"created_at"`
	UpdatedAt       time.Time                     `json:"updated_at"`
	CreatedBy       *uuid.UUID                    `json:"created_by,omitempty"`
	UpdatedBy       *uuid.UUID                    `json:"updated_by,omitempty"`
}

// CustomerSummaryResponse represents a summarized customer.
type CustomerSummaryResponse struct {
	ID             uuid.UUID             `json:"id"`
	Code           string                `json:"code"`
	Name           string                `json:"name"`
	Type           domain.CustomerType   `json:"type"`
	Status         domain.CustomerStatus `json:"status"`
	Tier           domain.CustomerTier   `json:"tier"`
	Email          string                `json:"email,omitempty"`
	Phone          string                `json:"phone,omitempty"`
	OwnerID        *uuid.UUID            `json:"owner_id,omitempty"`
	ContactCount   int                   `json:"contact_count"`
	Tags           []string              `json:"tags,omitempty"`
	LastContactedAt *time.Time           `json:"last_contacted_at,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
}

// CustomerListResponse represents a paginated list of customers.
type CustomerListResponse struct {
	Customers  []CustomerSummaryResponse `json:"customers"`
	Total      int64                     `json:"total"`
	Offset     int                       `json:"offset"`
	Limit      int                       `json:"limit"`
	HasMore    bool                      `json:"has_more"`
}

// ============================================================================
// Phone DTOs
// ============================================================================

// PhoneInput represents phone number input.
type PhoneInput struct {
	Number    string           `json:"number" validate:"required"`
	Type      domain.PhoneType `json:"type,omitempty" validate:"omitempty,oneof=mobile work home fax whatsapp other"`
	IsPrimary bool             `json:"is_primary,omitempty"`
}

// PhoneResponse represents a phone number in responses.
type PhoneResponse struct {
	Raw         string           `json:"raw"`
	Formatted   string           `json:"formatted"`
	E164        string           `json:"e164"`
	CountryCode string           `json:"country_code,omitempty"`
	Extension   string           `json:"extension,omitempty"`
	Type        domain.PhoneType `json:"type"`
	IsPrimary   bool             `json:"is_primary"`
}

// ============================================================================
// Address DTOs
// ============================================================================

// AddressInput represents address input.
type AddressInput struct {
	Line1       string             `json:"line1" validate:"required,max=200"`
	Line2       string             `json:"line2,omitempty" validate:"omitempty,max=200"`
	Line3       string             `json:"line3,omitempty" validate:"omitempty,max=200"`
	City        string             `json:"city" validate:"required,max=100"`
	State       string             `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode  string             `json:"postal_code" validate:"required,max=20"`
	CountryCode string             `json:"country_code" validate:"required,len=2"`
	AddressType domain.AddressType `json:"address_type,omitempty" validate:"omitempty,oneof=billing shipping office home other"`
	IsPrimary   bool               `json:"is_primary,omitempty"`
	Label       string             `json:"label,omitempty" validate:"omitempty,max=50"`
}

// AddressResponse represents an address in responses.
type AddressResponse struct {
	Line1       string             `json:"line1"`
	Line2       string             `json:"line2,omitempty"`
	Line3       string             `json:"line3,omitempty"`
	City        string             `json:"city"`
	State       string             `json:"state,omitempty"`
	PostalCode  string             `json:"postal_code"`
	Country     string             `json:"country"`
	CountryCode string             `json:"country_code"`
	AddressType domain.AddressType `json:"address_type"`
	IsPrimary   bool               `json:"is_primary"`
	IsVerified  bool               `json:"is_verified"`
	Latitude    *float64           `json:"latitude,omitempty"`
	Longitude   *float64           `json:"longitude,omitempty"`
	Label       string             `json:"label,omitempty"`
	Formatted   string             `json:"formatted"`
}

// ============================================================================
// Social Profile DTOs
// ============================================================================

// SocialProfileInput represents social profile input.
type SocialProfileInput struct {
	Platform    domain.SocialPlatform `json:"platform" validate:"required"`
	URL         string                `json:"url" validate:"required,url"`
	DisplayName string                `json:"display_name,omitempty"`
}

// SocialProfileResponse represents a social profile in responses.
type SocialProfileResponse struct {
	Platform     domain.SocialPlatform `json:"platform"`
	URL          string                `json:"url"`
	Username     string                `json:"username,omitempty"`
	DisplayName  string                `json:"display_name,omitempty"`
	Followers    *int                  `json:"followers,omitempty"`
	IsVerified   bool                  `json:"is_verified"`
}

// ============================================================================
// Company Info DTOs
// ============================================================================

// CompanyInfoInput represents company information input.
type CompanyInfoInput struct {
	LegalName          string            `json:"legal_name,omitempty" validate:"omitempty,max=255"`
	TradingName        string            `json:"trading_name,omitempty" validate:"omitempty,max=255"`
	RegistrationNumber string            `json:"registration_number,omitempty" validate:"omitempty,max=50"`
	TaxID              string            `json:"tax_id,omitempty" validate:"omitempty,max=50"`
	Industry           domain.Industry   `json:"industry,omitempty"`
	Size               domain.CompanySize `json:"size,omitempty"`
	EmployeeCount      *int              `json:"employee_count,omitempty" validate:"omitempty,min=0"`
	FoundedYear        *int              `json:"founded_year,omitempty" validate:"omitempty,min=1800,max=2100"`
	Description        string            `json:"description,omitempty" validate:"omitempty,max=2000"`
	ParentCompanyID    *uuid.UUID        `json:"parent_company_id,omitempty"`
}

// CompanyInfoResponse represents company information in responses.
type CompanyInfoResponse struct {
	LegalName          string             `json:"legal_name,omitempty"`
	TradingName        string             `json:"trading_name,omitempty"`
	RegistrationNumber string             `json:"registration_number,omitempty"`
	TaxID              string             `json:"tax_id,omitempty"`
	Industry           domain.Industry    `json:"industry,omitempty"`
	Size               domain.CompanySize `json:"size,omitempty"`
	EmployeeCount      *int               `json:"employee_count,omitempty"`
	AnnualRevenue      *MoneyResponse     `json:"annual_revenue,omitempty"`
	FoundedYear        *int               `json:"founded_year,omitempty"`
	Description        string             `json:"description,omitempty"`
	ParentCompanyID    *uuid.UUID         `json:"parent_company_id,omitempty"`
}

// ============================================================================
// Financials DTOs
// ============================================================================

// MoneyInput represents money input.
type MoneyInput struct {
	Amount   float64         `json:"amount" validate:"min=0"`
	Currency domain.Currency `json:"currency" validate:"required"`
}

// MoneyResponse represents money in responses.
type MoneyResponse struct {
	Amount   float64         `json:"amount"`
	Currency domain.Currency `json:"currency"`
	Display  string          `json:"display"`
}

// CustomerFinancialsInput represents financial information input.
type CustomerFinancialsInput struct {
	Currency           domain.Currency `json:"currency,omitempty"`
	CreditLimit        *MoneyInput     `json:"credit_limit,omitempty"`
	PaymentTerms       *int            `json:"payment_terms,omitempty" validate:"omitempty,min=0,max=365"`
	TaxExempt          *bool           `json:"tax_exempt,omitempty"`
	TaxExemptionID     string          `json:"tax_exemption_id,omitempty"`
	BillingEmail       string          `json:"billing_email,omitempty" validate:"omitempty,email"`
	DefaultDiscountPct *float64        `json:"default_discount_pct,omitempty" validate:"omitempty,min=0,max=100"`
}

// CustomerFinancialsResponse represents financial information in responses.
type CustomerFinancialsResponse struct {
	Currency           domain.Currency `json:"currency"`
	CreditLimit        *MoneyResponse  `json:"credit_limit,omitempty"`
	CurrentBalance     *MoneyResponse  `json:"current_balance,omitempty"`
	LifetimeValue      *MoneyResponse  `json:"lifetime_value,omitempty"`
	PaymentTerms       int             `json:"payment_terms"`
	TaxExempt          bool            `json:"tax_exempt"`
	TaxExemptionID     string          `json:"tax_exemption_id,omitempty"`
	BillingEmail       string          `json:"billing_email,omitempty"`
	DefaultDiscountPct float64         `json:"default_discount_pct"`
	LastPaymentAt      *time.Time      `json:"last_payment_at,omitempty"`
	LastPurchaseAt     *time.Time      `json:"last_purchase_at,omitempty"`
	TotalPurchases     int             `json:"total_purchases"`
	TotalSpent         *MoneyResponse  `json:"total_spent,omitempty"`
}

// ============================================================================
// Preferences DTOs
// ============================================================================

// CustomerPreferencesInput represents preferences input.
type CustomerPreferencesInput struct {
	Language          string                        `json:"language,omitempty" validate:"omitempty,len=2"`
	Timezone          string                        `json:"timezone,omitempty"`
	Currency          domain.Currency               `json:"currency,omitempty"`
	DateFormat        string                        `json:"date_format,omitempty"`
	CommPreference    domain.CommunicationPreference `json:"comm_preference,omitempty"`
	OptedOutMarketing *bool                         `json:"opted_out_marketing,omitempty"`
	NewsletterOptIn   *bool                         `json:"newsletter_opt_in,omitempty"`
	SMSOptIn          *bool                         `json:"sms_opt_in,omitempty"`
}

// CustomerPreferencesResponse represents preferences in responses.
type CustomerPreferencesResponse struct {
	Language          string                        `json:"language"`
	Timezone          string                        `json:"timezone"`
	Currency          domain.Currency               `json:"currency"`
	DateFormat        string                        `json:"date_format"`
	CommPreference    domain.CommunicationPreference `json:"comm_preference"`
	OptedOutMarketing bool                          `json:"opted_out_marketing"`
	MarketingConsent  *time.Time                    `json:"marketing_consent,omitempty"`
	NewsletterOptIn   bool                          `json:"newsletter_opt_in"`
	SMSOptIn          bool                          `json:"sms_opt_in"`
}

// ============================================================================
// Stats DTOs
// ============================================================================

// CustomerStatsResponse represents customer statistics.
type CustomerStatsResponse struct {
	ContactCount         int            `json:"contact_count"`
	ActiveContactCount   int            `json:"active_contact_count"`
	NoteCount            int            `json:"note_count"`
	ActivityCount        int            `json:"activity_count"`
	DealCount            int            `json:"deal_count"`
	WonDealCount         int            `json:"won_deal_count"`
	LostDealCount        int            `json:"lost_deal_count"`
	OpenDealValue        *MoneyResponse `json:"open_deal_value,omitempty"`
	DaysSinceLastContact *int           `json:"days_since_last_contact,omitempty"`
	AvgDealSize          *MoneyResponse `json:"avg_deal_size,omitempty"`
	EngagementScore      int            `json:"engagement_score"`
	HealthScore          int            `json:"health_score"`
	LastCalculatedAt     *time.Time     `json:"last_calculated_at,omitempty"`
}

// ============================================================================
// Status Change DTOs
// ============================================================================

// ChangeStatusRequest represents a status change request.
type ChangeStatusRequest struct {
	Status  domain.CustomerStatus `json:"status" validate:"required"`
	Reason  string                `json:"reason,omitempty" validate:"omitempty,max=500"`
	Version int                   `json:"version" validate:"required,min=1"`
}

// ConvertCustomerRequest represents a customer conversion request.
type ConvertCustomerRequest struct {
	Reason  string `json:"reason,omitempty" validate:"omitempty,max=500"`
	Version int    `json:"version" validate:"required,min=1"`
}

// AssignOwnerRequest represents an owner assignment request.
type AssignOwnerRequest struct {
	OwnerID uuid.UUID `json:"owner_id" validate:"required"`
	Version int       `json:"version" validate:"required,min=1"`
}

// AddTagRequest represents a tag addition request.
type AddTagRequest struct {
	Tag string `json:"tag" validate:"required,max=50"`
}

// AddToSegmentRequest represents a segment addition request.
type AddToSegmentRequest struct {
	SegmentID uuid.UUID `json:"segment_id" validate:"required"`
}

// ============================================================================
// Search and Filter DTOs
// ============================================================================

// SearchCustomersRequest represents a customer search request.
type SearchCustomersRequest struct {
	Query         string                  `json:"query,omitempty" validate:"omitempty,max=200"`
	Types         []domain.CustomerType   `json:"types,omitempty"`
	Statuses      []domain.CustomerStatus `json:"statuses,omitempty"`
	Tiers         []domain.CustomerTier   `json:"tiers,omitempty"`
	Sources       []domain.CustomerSource `json:"sources,omitempty"`
	Tags          []string                `json:"tags,omitempty"`
	OwnerIDs      []uuid.UUID             `json:"owner_ids,omitempty"`
	SegmentIDs    []uuid.UUID             `json:"segment_ids,omitempty"`
	Industries    []domain.Industry       `json:"industries,omitempty"`
	Countries     []string                `json:"countries,omitempty"`
	CreatedAfter  *time.Time              `json:"created_after,omitempty"`
	CreatedBefore *time.Time              `json:"created_before,omitempty"`
	IncludeDeleted bool                   `json:"include_deleted,omitempty"`
	Offset        int                     `json:"offset,omitempty" validate:"min=0"`
	Limit         int                     `json:"limit,omitempty" validate:"min=1,max=100"`
	SortBy        string                  `json:"sort_by,omitempty" validate:"omitempty,oneof=name created_at updated_at last_contacted_at"`
	SortOrder     string                  `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// ============================================================================
// Bulk Operations DTOs
// ============================================================================

// BulkUpdateRequest represents a bulk update request.
type BulkUpdateRequest struct {
	CustomerIDs []uuid.UUID            `json:"customer_ids" validate:"required,min=1,max=100"`
	Updates     map[string]interface{} `json:"updates" validate:"required"`
}

// BulkDeleteRequest represents a bulk delete request.
type BulkDeleteRequest struct {
	CustomerIDs []uuid.UUID `json:"customer_ids" validate:"required,min=1,max=100"`
}

// BulkOperationResponse represents a bulk operation result.
type BulkOperationResponse struct {
	Processed int      `json:"processed"`
	Succeeded int      `json:"succeeded"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

// MergeCustomersRequest represents a customer merge request.
type MergeCustomersRequest struct {
	TargetID  uuid.UUID   `json:"target_id" validate:"required"`
	SourceIDs []uuid.UUID `json:"source_ids" validate:"required,min=1,max=10"`
}
