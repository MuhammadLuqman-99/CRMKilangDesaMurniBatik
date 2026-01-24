// Package dto contains Data Transfer Objects for the Customer application layer.
package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
)

// ============================================================================
// Contact Input DTOs
// ============================================================================

// CreateContactInput represents contact creation input (embedded in customer).
type CreateContactInput struct {
	FirstName      string                        `json:"first_name" validate:"required,min=1,max=100"`
	LastName       string                        `json:"last_name" validate:"required,min=1,max=100"`
	MiddleName     string                        `json:"middle_name,omitempty" validate:"omitempty,max=100"`
	Title          string                        `json:"title,omitempty" validate:"omitempty,max=20"`
	Suffix         string                        `json:"suffix,omitempty" validate:"omitempty,max=20"`
	Email          string                        `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumbers   []PhoneInput                  `json:"phone_numbers,omitempty" validate:"omitempty,max=5,dive"`
	Addresses      []AddressInput                `json:"addresses,omitempty" validate:"omitempty,max=3,dive"`
	SocialProfiles []SocialProfileInput          `json:"social_profiles,omitempty" validate:"omitempty,max=5,dive"`
	JobTitle       string                        `json:"job_title,omitempty" validate:"omitempty,max=100"`
	Department     string                        `json:"department,omitempty" validate:"omitempty,max=100"`
	Role           domain.ContactRole            `json:"role,omitempty" validate:"omitempty,oneof=decision_maker influencer user technical finance procurement executive other"`
	IsPrimary      bool                          `json:"is_primary,omitempty"`
	CommPreference domain.CommunicationPreference `json:"comm_preference,omitempty" validate:"omitempty,oneof=email phone whatsapp sms mail any"`
	Birthday       *time.Time                    `json:"birthday,omitempty"`
	Notes          string                        `json:"notes,omitempty" validate:"omitempty,max=2000"`
	Tags           []string                      `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields   map[string]interface{}        `json:"custom_fields,omitempty"`
	LinkedInURL    string                        `json:"linkedin_url,omitempty" validate:"omitempty,url"`
	ProfilePhotoURL string                       `json:"profile_photo_url,omitempty" validate:"omitempty,url"`
}

// CreateContactRequest represents a standalone contact creation request.
type CreateContactRequest struct {
	CustomerID     uuid.UUID                     `json:"customer_id" validate:"required"`
	FirstName      string                        `json:"first_name" validate:"required,min=1,max=100"`
	LastName       string                        `json:"last_name" validate:"required,min=1,max=100"`
	MiddleName     string                        `json:"middle_name,omitempty" validate:"omitempty,max=100"`
	Title          string                        `json:"title,omitempty" validate:"omitempty,max=20"`
	Suffix         string                        `json:"suffix,omitempty" validate:"omitempty,max=20"`
	Email          string                        `json:"email,omitempty" validate:"omitempty,email"`
	PhoneNumbers   []PhoneInput                  `json:"phone_numbers,omitempty" validate:"omitempty,max=5,dive"`
	Addresses      []AddressInput                `json:"addresses,omitempty" validate:"omitempty,max=3,dive"`
	SocialProfiles []SocialProfileInput          `json:"social_profiles,omitempty" validate:"omitempty,max=5,dive"`
	JobTitle       string                        `json:"job_title,omitempty" validate:"omitempty,max=100"`
	Department     string                        `json:"department,omitempty" validate:"omitempty,max=100"`
	Role           domain.ContactRole            `json:"role,omitempty" validate:"omitempty,oneof=decision_maker influencer user technical finance procurement executive other"`
	IsPrimary      bool                          `json:"is_primary,omitempty"`
	ReportsTo      *uuid.UUID                    `json:"reports_to,omitempty"`
	CommPreference domain.CommunicationPreference `json:"comm_preference,omitempty" validate:"omitempty,oneof=email phone whatsapp sms mail any"`
	OptedOutMarketing bool                       `json:"opted_out_marketing,omitempty"`
	Birthday       *time.Time                    `json:"birthday,omitempty"`
	Notes          string                        `json:"notes,omitempty" validate:"omitempty,max=2000"`
	Tags           []string                      `json:"tags,omitempty" validate:"omitempty,max=20,dive,max=50"`
	CustomFields   map[string]interface{}        `json:"custom_fields,omitempty"`
	LinkedInURL    string                        `json:"linkedin_url,omitempty" validate:"omitempty,url"`
	ProfilePhotoURL string                       `json:"profile_photo_url,omitempty" validate:"omitempty,url"`
}

// UpdateContactRequest represents a contact update request.
type UpdateContactRequest struct {
	FirstName       *string                        `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName        *string                        `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	MiddleName      *string                        `json:"middle_name,omitempty" validate:"omitempty,max=100"`
	Title           *string                        `json:"title,omitempty" validate:"omitempty,max=20"`
	Suffix          *string                        `json:"suffix,omitempty" validate:"omitempty,max=20"`
	Email           *string                        `json:"email,omitempty" validate:"omitempty,email"`
	JobTitle        *string                        `json:"job_title,omitempty" validate:"omitempty,max=100"`
	Department      *string                        `json:"department,omitempty" validate:"omitempty,max=100"`
	Role            *domain.ContactRole            `json:"role,omitempty"`
	IsPrimary       *bool                          `json:"is_primary,omitempty"`
	ReportsTo       *uuid.UUID                     `json:"reports_to,omitempty"`
	CommPreference  *domain.CommunicationPreference `json:"comm_preference,omitempty"`
	OptedOutMarketing *bool                        `json:"opted_out_marketing,omitempty"`
	Birthday        *time.Time                     `json:"birthday,omitempty"`
	Notes           *string                        `json:"notes,omitempty" validate:"omitempty,max=2000"`
	LinkedInURL     *string                        `json:"linkedin_url,omitempty" validate:"omitempty,url"`
	ProfilePhotoURL *string                        `json:"profile_photo_url,omitempty" validate:"omitempty,url"`
	CustomFields    map[string]interface{}         `json:"custom_fields,omitempty"`
	Version         int                            `json:"version" validate:"required,min=1"`
}

// ============================================================================
// Contact Response DTOs
// ============================================================================

// ContactResponse represents a full contact in API responses.
type ContactResponse struct {
	ID                uuid.UUID                      `json:"id"`
	CustomerID        uuid.UUID                      `json:"customer_id"`
	TenantID          uuid.UUID                      `json:"tenant_id"`
	Name              PersonNameResponse             `json:"name"`
	Email             string                         `json:"email,omitempty"`
	PhoneNumbers      []PhoneResponse                `json:"phone_numbers,omitempty"`
	Addresses         []AddressResponse              `json:"addresses,omitempty"`
	SocialProfiles    []SocialProfileResponse        `json:"social_profiles,omitempty"`
	JobTitle          string                         `json:"job_title,omitempty"`
	Department        string                         `json:"department,omitempty"`
	Role              domain.ContactRole             `json:"role"`
	Status            domain.ContactStatus           `json:"status"`
	IsPrimary         bool                           `json:"is_primary"`
	ReportsTo         *uuid.UUID                     `json:"reports_to,omitempty"`
	CommPreference    domain.CommunicationPreference `json:"comm_preference"`
	OptedOutMarketing bool                           `json:"opted_out_marketing"`
	MarketingConsent  *time.Time                     `json:"marketing_consent,omitempty"`
	Birthday          *time.Time                     `json:"birthday,omitempty"`
	Notes             string                         `json:"notes,omitempty"`
	Tags              []string                       `json:"tags,omitempty"`
	CustomFields      map[string]interface{}         `json:"custom_fields,omitempty"`
	LastContactedAt   *time.Time                     `json:"last_contacted_at,omitempty"`
	NextFollowUpAt    *time.Time                     `json:"next_follow_up_at,omitempty"`
	EngagementScore   int                            `json:"engagement_score"`
	LinkedInURL       string                         `json:"linkedin_url,omitempty"`
	ProfilePhotoURL   string                         `json:"profile_photo_url,omitempty"`
	Version           int                            `json:"version"`
	CreatedAt         time.Time                      `json:"created_at"`
	UpdatedAt         time.Time                      `json:"updated_at"`
	CreatedBy         *uuid.UUID                     `json:"created_by,omitempty"`
	UpdatedBy         *uuid.UUID                     `json:"updated_by,omitempty"`
}

// ContactSummaryResponse represents a summarized contact for listings.
type ContactSummaryResponse struct {
	ID              uuid.UUID             `json:"id"`
	CustomerID      uuid.UUID             `json:"customer_id"`
	FullName        string                `json:"full_name"`
	Email           string                `json:"email,omitempty"`
	Phone           string                `json:"phone,omitempty"`
	JobTitle        string                `json:"job_title,omitempty"`
	Department      string                `json:"department,omitempty"`
	Role            domain.ContactRole    `json:"role"`
	Status          domain.ContactStatus  `json:"status"`
	IsPrimary       bool                  `json:"is_primary"`
	EngagementScore int                   `json:"engagement_score"`
	LastContactedAt *time.Time            `json:"last_contacted_at,omitempty"`
	ProfilePhotoURL string                `json:"profile_photo_url,omitempty"`
}

// ContactListResponse represents a paginated list of contacts.
type ContactListResponse struct {
	Contacts []ContactSummaryResponse `json:"contacts"`
	Total    int64                    `json:"total"`
	Offset   int                      `json:"offset"`
	Limit    int                      `json:"limit"`
	HasMore  bool                     `json:"has_more"`
}

// PersonNameResponse represents a person name in responses.
type PersonNameResponse struct {
	Title       string `json:"title,omitempty"`
	FirstName   string `json:"first_name"`
	MiddleName  string `json:"middle_name,omitempty"`
	LastName    string `json:"last_name"`
	Suffix      string `json:"suffix,omitempty"`
	FullName    string `json:"full_name"`
	DisplayName string `json:"display_name"`
	Initials    string `json:"initials"`
}

// ============================================================================
// Contact Operation DTOs
// ============================================================================

// ChangeContactStatusRequest represents a contact status change request.
type ChangeContactStatusRequest struct {
	Status  domain.ContactStatus `json:"status" validate:"required,oneof=active inactive blocked"`
	Reason  string               `json:"reason,omitempty" validate:"omitempty,max=500"`
	Version int                  `json:"version" validate:"required,min=1"`
}

// SetPrimaryContactRequest represents a request to set a contact as primary.
type SetPrimaryContactRequest struct {
	ContactID uuid.UUID `json:"contact_id" validate:"required"`
	Version   int       `json:"version" validate:"required,min=1"`
}

// AssignContactManagerRequest represents a request to assign a manager.
type AssignContactManagerRequest struct {
	ManagerContactID *uuid.UUID `json:"manager_contact_id"`
	Version          int        `json:"version" validate:"required,min=1"`
}

// ContactPhoneRequest represents a phone number operation request.
type ContactPhoneRequest struct {
	Number    string           `json:"number" validate:"required"`
	Type      domain.PhoneType `json:"type,omitempty" validate:"omitempty,oneof=mobile work home fax whatsapp other"`
	IsPrimary bool             `json:"is_primary,omitempty"`
}

// RemovePhoneRequest represents a phone removal request.
type RemovePhoneRequest struct {
	E164 string `json:"e164" validate:"required"`
}

// ContactAddressRequest represents an address operation request.
type ContactAddressRequest struct {
	Line1       string             `json:"line1" validate:"required,max=200"`
	Line2       string             `json:"line2,omitempty" validate:"omitempty,max=200"`
	Line3       string             `json:"line3,omitempty" validate:"omitempty,max=200"`
	City        string             `json:"city" validate:"required,max=100"`
	State       string             `json:"state,omitempty" validate:"omitempty,max=100"`
	PostalCode  string             `json:"postal_code" validate:"required,max=20"`
	CountryCode string             `json:"country_code" validate:"required,len=2"`
	AddressType domain.AddressType `json:"address_type,omitempty" validate:"omitempty,oneof=billing shipping office home other"`
	IsPrimary   bool               `json:"is_primary,omitempty"`
}

// RemoveAddressRequest represents an address removal request.
type RemoveAddressRequest struct {
	Line1      string `json:"line1" validate:"required"`
	PostalCode string `json:"postal_code" validate:"required"`
}

// ContactSocialProfileRequest represents a social profile operation request.
type ContactSocialProfileRequest struct {
	Platform    domain.SocialPlatform `json:"platform" validate:"required"`
	URL         string                `json:"url" validate:"required,url"`
	DisplayName string                `json:"display_name,omitempty"`
}

// RemoveSocialProfileRequest represents a social profile removal request.
type RemoveSocialProfileRequest struct {
	Platform domain.SocialPlatform `json:"platform" validate:"required"`
}

// UpdateContactJobInfoRequest represents a job info update request.
type UpdateContactJobInfoRequest struct {
	JobTitle   string             `json:"job_title,omitempty" validate:"omitempty,max=100"`
	Department string             `json:"department,omitempty" validate:"omitempty,max=100"`
	Role       domain.ContactRole `json:"role,omitempty" validate:"omitempty,oneof=decision_maker influencer user technical finance procurement executive other"`
	Version    int                `json:"version" validate:"required,min=1"`
}

// UpdateCommPreferenceRequest represents a communication preference update.
type UpdateCommPreferenceRequest struct {
	CommPreference domain.CommunicationPreference `json:"comm_preference" validate:"required,oneof=email phone whatsapp sms mail any"`
	Version        int                            `json:"version" validate:"required,min=1"`
}

// MarketingOptRequest represents a marketing opt-in/out request.
type MarketingOptRequest struct {
	OptIn   bool `json:"opt_in"`
	Version int  `json:"version" validate:"required,min=1"`
}

// RecordContactInteractionRequest represents a contact interaction recording.
type RecordContactInteractionRequest struct {
	InteractionType string    `json:"interaction_type" validate:"required,oneof=call email meeting sms whatsapp other"`
	Notes           string    `json:"notes,omitempty" validate:"omitempty,max=2000"`
	FollowUpAt      *time.Time `json:"follow_up_at,omitempty"`
}

// SetFollowUpRequest represents a follow-up scheduling request.
type SetFollowUpRequest struct {
	FollowUpAt time.Time `json:"follow_up_at" validate:"required"`
	Version    int       `json:"version" validate:"required,min=1"`
}

// UpdateEngagementScoreRequest represents an engagement score update.
type UpdateEngagementScoreRequest struct {
	Score   int    `json:"score" validate:"min=0,max=100"`
	Reason  string `json:"reason,omitempty" validate:"omitempty,max=200"`
	Version int    `json:"version" validate:"required,min=1"`
}

// AddContactTagRequest represents a tag addition request.
type AddContactTagRequest struct {
	Tag string `json:"tag" validate:"required,max=50"`
}

// RemoveContactTagRequest represents a tag removal request.
type RemoveContactTagRequest struct {
	Tag string `json:"tag" validate:"required,max=50"`
}

// SetContactCustomFieldRequest represents a custom field set request.
type SetContactCustomFieldRequest struct {
	Key   string      `json:"key" validate:"required,max=50"`
	Value interface{} `json:"value" validate:"required"`
}

// ============================================================================
// Contact Search and Filter DTOs
// ============================================================================

// SearchContactsRequest represents a contact search request.
type SearchContactsRequest struct {
	Query           string                 `json:"query,omitempty" validate:"omitempty,max=200"`
	CustomerID      *uuid.UUID             `json:"customer_id,omitempty"`
	Statuses        []domain.ContactStatus `json:"statuses,omitempty"`
	Roles           []domain.ContactRole   `json:"roles,omitempty"`
	IsPrimary       *bool                  `json:"is_primary,omitempty"`
	HasEmail        *bool                  `json:"has_email,omitempty"`
	HasPhone        *bool                  `json:"has_phone,omitempty"`
	OptedInMarketing *bool                 `json:"opted_in_marketing,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	MinEngagement   *int                   `json:"min_engagement,omitempty" validate:"omitempty,min=0,max=100"`
	MaxEngagement   *int                   `json:"max_engagement,omitempty" validate:"omitempty,min=0,max=100"`
	NeedsFollowUp   *bool                  `json:"needs_follow_up,omitempty"`
	CreatedAfter    *time.Time             `json:"created_after,omitempty"`
	CreatedBefore   *time.Time             `json:"created_before,omitempty"`
	LastContactedAfter  *time.Time         `json:"last_contacted_after,omitempty"`
	LastContactedBefore *time.Time         `json:"last_contacted_before,omitempty"`
	IncludeDeleted  bool                   `json:"include_deleted,omitempty"`
	Offset          int                    `json:"offset,omitempty" validate:"min=0"`
	Limit           int                    `json:"limit,omitempty" validate:"min=1,max=100"`
	SortBy          string                 `json:"sort_by,omitempty" validate:"omitempty,oneof=name email created_at updated_at last_contacted_at engagement_score"`
	SortOrder       string                 `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// ============================================================================
// Contact Bulk Operations DTOs
// ============================================================================

// BulkContactUpdateRequest represents a bulk contact update request.
type BulkContactUpdateRequest struct {
	ContactIDs []uuid.UUID            `json:"contact_ids" validate:"required,min=1,max=50"`
	Updates    map[string]interface{} `json:"updates" validate:"required"`
}

// BulkContactDeleteRequest represents a bulk contact delete request.
type BulkContactDeleteRequest struct {
	ContactIDs []uuid.UUID `json:"contact_ids" validate:"required,min=1,max=50"`
}

// TransferContactsRequest represents a request to transfer contacts.
type TransferContactsRequest struct {
	ContactIDs       []uuid.UUID `json:"contact_ids" validate:"required,min=1,max=50"`
	TargetCustomerID uuid.UUID   `json:"target_customer_id" validate:"required"`
}

// MergeContactsRequest represents a contact merge request.
type MergeContactsRequest struct {
	TargetID  uuid.UUID   `json:"target_id" validate:"required"`
	SourceIDs []uuid.UUID `json:"source_ids" validate:"required,min=1,max=5"`
}

// ============================================================================
// Contact Activity DTOs
// ============================================================================

// ContactActivityResponse represents a contact activity entry.
type ContactActivityResponse struct {
	ID           uuid.UUID              `json:"id"`
	ContactID    uuid.UUID              `json:"contact_id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	PerformedBy  *uuid.UUID             `json:"performed_by,omitempty"`
	PerformedAt  time.Time              `json:"performed_at"`
}

// ContactActivityListResponse represents a paginated list of contact activities.
type ContactActivityListResponse struct {
	Activities []ContactActivityResponse `json:"activities"`
	Total      int64                     `json:"total"`
	Offset     int                       `json:"offset"`
	Limit      int                       `json:"limit"`
	HasMore    bool                      `json:"has_more"`
}

// GetContactActivitiesRequest represents a request to get contact activities.
type GetContactActivitiesRequest struct {
	ContactID uuid.UUID  `json:"contact_id" validate:"required"`
	Types     []string   `json:"types,omitempty"`
	After     *time.Time `json:"after,omitempty"`
	Before    *time.Time `json:"before,omitempty"`
	Offset    int        `json:"offset,omitempty" validate:"min=0"`
	Limit     int        `json:"limit,omitempty" validate:"min=1,max=100"`
}

// ============================================================================
// Contact Statistics DTOs
// ============================================================================

// ContactStatsResponse represents contact statistics.
type ContactStatsResponse struct {
	TotalContacts         int        `json:"total_contacts"`
	ActiveContacts        int        `json:"active_contacts"`
	InactiveContacts      int        `json:"inactive_contacts"`
	BlockedContacts       int        `json:"blocked_contacts"`
	PrimaryContacts       int        `json:"primary_contacts"`
	MarketingOptInCount   int        `json:"marketing_opt_in_count"`
	MarketingOptOutCount  int        `json:"marketing_opt_out_count"`
	AvgEngagementScore    float64    `json:"avg_engagement_score"`
	ContactsNeedingFollowUp int      `json:"contacts_needing_follow_up"`
	RoleDistribution      map[string]int `json:"role_distribution"`
	DepartmentDistribution map[string]int `json:"department_distribution"`
	LastCalculatedAt      time.Time  `json:"last_calculated_at"`
}

// CustomerContactStatsRequest represents a request for customer contact stats.
type CustomerContactStatsRequest struct {
	CustomerID uuid.UUID `json:"customer_id" validate:"required"`
}

// ============================================================================
// Duplicate Detection DTOs
// ============================================================================

// DuplicateContactResponse represents a potential duplicate contact.
type DuplicateContactResponse struct {
	Contact     ContactSummaryResponse `json:"contact"`
	MatchScore  float64                `json:"match_score"`
	MatchFields []string               `json:"match_fields"`
	MatchReason string                 `json:"match_reason"`
}

// FindDuplicateContactsRequest represents a duplicate search request.
type FindDuplicateContactsRequest struct {
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	Email      string     `json:"email,omitempty"`
	Phone      string     `json:"phone,omitempty"`
	Name       string     `json:"name,omitempty"`
	Threshold  float64    `json:"threshold,omitempty" validate:"omitempty,min=0.5,max=1.0"`
}

// FindDuplicateContactsResponse represents duplicate search results.
type FindDuplicateContactsResponse struct {
	Duplicates []DuplicateContactResponse `json:"duplicates"`
	Total      int                        `json:"total"`
}

// ============================================================================
// Contact Export/Import DTOs
// ============================================================================

// ContactExportRequest represents a contact export request.
type ContactExportRequest struct {
	CustomerID *uuid.UUID             `json:"customer_id,omitempty"`
	Format     string                 `json:"format" validate:"required,oneof=csv xlsx json"`
	Fields     []string               `json:"fields,omitempty"`
	Filter     *SearchContactsRequest `json:"filter,omitempty"`
}

// ContactImportRequest represents a contact import request.
type ContactImportRequest struct {
	CustomerID      uuid.UUID `json:"customer_id" validate:"required"`
	Format          string    `json:"format" validate:"required,oneof=csv xlsx json"`
	Data            []byte    `json:"data" validate:"required"`
	SkipDuplicates  bool      `json:"skip_duplicates,omitempty"`
	UpdateExisting  bool      `json:"update_existing,omitempty"`
	FieldMapping    map[string]string `json:"field_mapping,omitempty"`
}

// ContactImportResult represents a single contact import result.
type ContactImportResult struct {
	Row     int      `json:"row"`
	Success bool     `json:"success"`
	ContactID *uuid.UUID `json:"contact_id,omitempty"`
	Errors  []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ContactImportResponse represents an import operation response.
type ContactImportResponse struct {
	TotalRows      int                   `json:"total_rows"`
	SuccessCount   int                   `json:"success_count"`
	FailureCount   int                   `json:"failure_count"`
	SkippedCount   int                   `json:"skipped_count"`
	UpdatedCount   int                   `json:"updated_count"`
	Results        []ContactImportResult `json:"results,omitempty"`
	Errors         []string              `json:"errors,omitempty"`
}
