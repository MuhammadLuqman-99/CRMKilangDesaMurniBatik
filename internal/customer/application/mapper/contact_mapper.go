// Package mapper provides functions to map between domain entities and DTOs.
package mapper

import (
	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/dto"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
)

// ContactMapper provides mapping functions for Contact entities.
type ContactMapper struct{}

// NewContactMapper creates a new ContactMapper.
func NewContactMapper() *ContactMapper {
	return &ContactMapper{}
}

// ============================================================================
// Domain to DTO Mappings
// ============================================================================

// ToResponse maps a Contact domain entity to ContactResponse DTO.
func (m *ContactMapper) ToResponse(contact *domain.Contact) *dto.ContactResponse {
	if contact == nil {
		return nil
	}

	return &dto.ContactResponse{
		ID:                contact.ID,
		CustomerID:        contact.CustomerID,
		TenantID:          contact.TenantID,
		Name:              m.personNameToResponse(contact.Name),
		Email:             contact.Email.String(),
		PhoneNumbers:      m.phonesToResponse(contact.PhoneNumbers),
		Addresses:         m.addressesToResponse(contact.Addresses),
		SocialProfiles:    m.socialProfilesToResponse(contact.SocialProfiles),
		JobTitle:          contact.JobTitle,
		Department:        contact.Department,
		Role:              contact.Role,
		Status:            contact.Status,
		IsPrimary:         contact.IsPrimary,
		ReportsTo:         contact.ReportsTo,
		CommPreference:    contact.CommPreference,
		OptedOutMarketing: contact.OptedOutMarketing,
		MarketingConsent:  contact.MarketingConsent,
		Birthday:          contact.Birthday,
		Notes:             contact.Notes,
		Tags:              contact.Tags,
		CustomFields:      contact.CustomFields,
		LastContactedAt:   contact.LastContactedAt,
		NextFollowUpAt:    contact.NextFollowUpAt,
		EngagementScore:   contact.EngagementScore,
		LinkedInURL:       contact.LinkedInURL,
		ProfilePhotoURL:   contact.ProfilePhotoURL,
		Version:           contact.Version,
		CreatedAt:         contact.CreatedAt,
		UpdatedAt:         contact.UpdatedAt,
		CreatedBy:         contact.AuditInfo.CreatedBy,
		UpdatedBy:         contact.AuditInfo.UpdatedBy,
	}
}

// ToSummaryResponse maps a Contact to ContactSummaryResponse.
func (m *ContactMapper) ToSummaryResponse(contact *domain.Contact) *dto.ContactSummaryResponse {
	if contact == nil {
		return nil
	}

	var phone string
	if primaryPhone := contact.GetPrimaryPhone(); primaryPhone != nil {
		phone = primaryPhone.Formatted()
	}

	return &dto.ContactSummaryResponse{
		ID:              contact.ID,
		CustomerID:      contact.CustomerID,
		FullName:        contact.FullName(),
		Email:           contact.Email.String(),
		Phone:           phone,
		JobTitle:        contact.JobTitle,
		Department:      contact.Department,
		Role:            contact.Role,
		Status:          contact.Status,
		IsPrimary:       contact.IsPrimary,
		EngagementScore: contact.EngagementScore,
		LastContactedAt: contact.LastContactedAt,
		ProfilePhotoURL: contact.ProfilePhotoURL,
	}
}

// ToListResponse maps a slice of Contacts to ContactListResponse.
func (m *ContactMapper) ToListResponse(contacts []*domain.Contact, total int64, offset, limit int) *dto.ContactListResponse {
	summaries := make([]dto.ContactSummaryResponse, len(contacts))
	for i, contact := range contacts {
		if summary := m.ToSummaryResponse(contact); summary != nil {
			summaries[i] = *summary
		}
	}

	return &dto.ContactListResponse{
		Contacts: summaries,
		Total:    total,
		Offset:   offset,
		Limit:    limit,
		HasMore:  int64(offset+len(contacts)) < total,
	}
}

// ============================================================================
// Value Object to DTO Mappings
// ============================================================================

func (m *ContactMapper) personNameToResponse(name domain.PersonName) dto.PersonNameResponse {
	return dto.PersonNameResponse{
		Title:       name.Title,
		FirstName:   name.FirstName,
		MiddleName:  name.MiddleName,
		LastName:    name.LastName,
		Suffix:      name.Suffix,
		FullName:    name.FullName(),
		DisplayName: name.DisplayName(),
		Initials:    name.Initials(),
	}
}

func (m *ContactMapper) phonesToResponse(phones []domain.PhoneNumber) []dto.PhoneResponse {
	if len(phones) == 0 {
		return nil
	}

	responses := make([]dto.PhoneResponse, len(phones))
	for i, phone := range phones {
		responses[i] = dto.PhoneResponse{
			Raw:         phone.Raw(),
			Formatted:   phone.Formatted(),
			E164:        phone.E164(),
			CountryCode: phone.CountryCode(),
			Extension:   phone.Extension(),
			Type:        phone.Type(),
			IsPrimary:   phone.IsPrimary(),
		}
	}
	return responses
}

func (m *ContactMapper) addressesToResponse(addresses []domain.Address) []dto.AddressResponse {
	if len(addresses) == 0 {
		return nil
	}

	responses := make([]dto.AddressResponse, len(addresses))
	for i, addr := range addresses {
		responses[i] = dto.AddressResponse{
			Line1:       addr.Line1,
			Line2:       addr.Line2,
			Line3:       addr.Line3,
			City:        addr.City,
			State:       addr.State,
			PostalCode:  addr.PostalCode,
			Country:     addr.Country,
			CountryCode: addr.CountryCode,
			AddressType: addr.AddressType,
			IsPrimary:   addr.IsPrimary,
			IsVerified:  addr.IsVerified,
			Latitude:    addr.Latitude,
			Longitude:   addr.Longitude,
			Label:       addr.Label,
			Formatted:   addr.Format(),
		}
	}
	return responses
}

func (m *ContactMapper) socialProfilesToResponse(profiles []domain.SocialProfile) []dto.SocialProfileResponse {
	if len(profiles) == 0 {
		return nil
	}

	responses := make([]dto.SocialProfileResponse, len(profiles))
	for i, profile := range profiles {
		responses[i] = dto.SocialProfileResponse{
			Platform:    profile.Platform,
			URL:         profile.URL,
			Username:    profile.Username,
			DisplayName: profile.DisplayName,
			Followers:   profile.Followers,
			IsVerified:  profile.IsVerified,
		}
	}
	return responses
}

// ============================================================================
// DTO to Domain Mappings
// ============================================================================

// ToDomain maps CreateContactRequest to a domain Contact using the builder.
func (m *ContactMapper) ToDomain(tenantID, creatorID uuid.UUID, req *dto.CreateContactRequest) (*domain.Contact, error) {
	builder := domain.NewContactBuilder(req.CustomerID, tenantID).
		WithCreatedBy(creatorID)

	// Set name
	if req.MiddleName != "" || req.Title != "" || req.Suffix != "" {
		builder.WithFullName(req.Title, req.FirstName, req.MiddleName, req.LastName, req.Suffix)
	} else {
		builder.WithName(req.FirstName, req.LastName)
	}

	// Set email
	if req.Email != "" {
		builder.WithEmail(req.Email)
	}

	// Set phone numbers
	for _, phone := range req.PhoneNumbers {
		builder.WithPhone(phone.Number, phone.Type, phone.IsPrimary)
	}

	// Set addresses
	for _, addr := range req.Addresses {
		builder.WithAddress(addr.Line1, addr.City, addr.PostalCode, addr.CountryCode, addr.AddressType, addr.IsPrimary)
	}

	// Set social profiles
	for _, profile := range req.SocialProfiles {
		builder.WithSocialProfile(profile.Platform, profile.URL)
	}

	// Set job info
	if req.JobTitle != "" || req.Department != "" || req.Role != "" {
		builder.WithJobTitle(req.JobTitle).WithDepartment(req.Department).WithRole(req.Role)
	}

	// Set primary
	if req.IsPrimary {
		builder.AsPrimary()
	}

	// Set communication preference
	if req.CommPreference != "" {
		builder.WithCommPreference(req.CommPreference)
	}

	// Set birthday
	if req.Birthday != nil {
		builder.WithBirthday(*req.Birthday)
	}

	// Set notes
	if req.Notes != "" {
		builder.WithNotes(req.Notes)
	}

	// Set tags
	if len(req.Tags) > 0 {
		builder.WithTags(req.Tags...)
	}

	// Set custom fields
	for key, value := range req.CustomFields {
		builder.WithCustomField(key, value)
	}

	contact, err := builder.Build()
	if err != nil {
		return nil, err
	}

	// Set additional fields not covered by builder
	if req.ReportsTo != nil {
		contact.SetReportsTo(req.ReportsTo)
	}
	if req.LinkedInURL != "" {
		contact.SetLinkedIn(req.LinkedInURL)
	}
	if req.ProfilePhotoURL != "" {
		contact.SetProfilePhoto(req.ProfilePhotoURL)
	}
	if req.OptedOutMarketing {
		contact.OptOutMarketing()
	}

	return contact, nil
}

// FromCreateInput maps CreateContactInput (embedded) to a domain Contact.
func (m *ContactMapper) FromCreateInput(customerID, tenantID, creatorID uuid.UUID, input *dto.CreateContactInput) (*domain.Contact, error) {
	builder := domain.NewContactBuilder(customerID, tenantID).
		WithCreatedBy(creatorID)

	// Set name
	if input.MiddleName != "" || input.Title != "" || input.Suffix != "" {
		builder.WithFullName(input.Title, input.FirstName, input.MiddleName, input.LastName, input.Suffix)
	} else {
		builder.WithName(input.FirstName, input.LastName)
	}

	// Set email
	if input.Email != "" {
		builder.WithEmail(input.Email)
	}

	// Set phone numbers
	for _, phone := range input.PhoneNumbers {
		builder.WithPhone(phone.Number, phone.Type, phone.IsPrimary)
	}

	// Set addresses
	for _, addr := range input.Addresses {
		builder.WithAddress(addr.Line1, addr.City, addr.PostalCode, addr.CountryCode, addr.AddressType, addr.IsPrimary)
	}

	// Set social profiles
	for _, profile := range input.SocialProfiles {
		builder.WithSocialProfile(profile.Platform, profile.URL)
	}

	// Set job info
	if input.JobTitle != "" {
		builder.WithJobTitle(input.JobTitle)
	}
	if input.Department != "" {
		builder.WithDepartment(input.Department)
	}
	if input.Role != "" {
		builder.WithRole(input.Role)
	}

	// Set primary
	if input.IsPrimary {
		builder.AsPrimary()
	}

	// Set communication preference
	if input.CommPreference != "" {
		builder.WithCommPreference(input.CommPreference)
	}

	// Set birthday
	if input.Birthday != nil {
		builder.WithBirthday(*input.Birthday)
	}

	// Set notes
	if input.Notes != "" {
		builder.WithNotes(input.Notes)
	}

	// Set tags
	if len(input.Tags) > 0 {
		builder.WithTags(input.Tags...)
	}

	// Set custom fields
	for key, value := range input.CustomFields {
		builder.WithCustomField(key, value)
	}

	contact, err := builder.Build()
	if err != nil {
		return nil, err
	}

	// Set additional fields
	if input.LinkedInURL != "" {
		contact.SetLinkedIn(input.LinkedInURL)
	}
	if input.ProfilePhotoURL != "" {
		contact.SetProfilePhoto(input.ProfilePhotoURL)
	}

	return contact, nil
}

// ApplyUpdate applies UpdateContactRequest to an existing Contact.
func (m *ContactMapper) ApplyUpdate(contact *domain.Contact, req *dto.UpdateContactRequest) error {
	// Update name if any name field is provided
	if req.FirstName != nil || req.LastName != nil || req.MiddleName != nil || req.Title != nil || req.Suffix != nil {
		name := contact.Name
		if req.FirstName != nil {
			name.FirstName = *req.FirstName
		}
		if req.LastName != nil {
			name.LastName = *req.LastName
		}
		if req.MiddleName != nil {
			name.MiddleName = *req.MiddleName
		}
		if req.Title != nil {
			name.Title = *req.Title
		}
		if req.Suffix != nil {
			name.Suffix = *req.Suffix
		}
		contact.UpdateName(name)
	}

	// Update email
	if req.Email != nil {
		email, err := domain.NewEmail(*req.Email)
		if err != nil {
			return err
		}
		if err := contact.UpdateEmail(email); err != nil {
			return err
		}
	}

	// Update job info
	if req.JobTitle != nil || req.Department != nil || req.Role != nil {
		jobTitle := contact.JobTitle
		department := contact.Department
		role := contact.Role
		if req.JobTitle != nil {
			jobTitle = *req.JobTitle
		}
		if req.Department != nil {
			department = *req.Department
		}
		if req.Role != nil {
			role = *req.Role
		}
		contact.UpdateJobInfo(jobTitle, department, role)
	}

	// Update primary status
	if req.IsPrimary != nil {
		contact.SetPrimary(*req.IsPrimary)
	}

	// Update reports to
	if req.ReportsTo != nil {
		contact.SetReportsTo(req.ReportsTo)
	}

	// Update communication preference
	if req.CommPreference != nil {
		contact.SetCommunicationPreference(*req.CommPreference)
	}

	// Update marketing opt-out
	if req.OptedOutMarketing != nil {
		if *req.OptedOutMarketing {
			contact.OptOutMarketing()
		} else {
			contact.OptInMarketing()
		}
	}

	// Update birthday
	if req.Birthday != nil {
		contact.Birthday = req.Birthday
		contact.MarkUpdated()
	}

	// Update notes
	if req.Notes != nil {
		contact.Notes = *req.Notes
		contact.MarkUpdated()
	}

	// Update LinkedIn URL
	if req.LinkedInURL != nil {
		contact.SetLinkedIn(*req.LinkedInURL)
	}

	// Update profile photo
	if req.ProfilePhotoURL != nil {
		contact.SetProfilePhoto(*req.ProfilePhotoURL)
	}

	// Update custom fields
	for key, value := range req.CustomFields {
		contact.SetCustomField(key, value)
	}

	return nil
}

// SearchRequestToFilter converts SearchContactsRequest to domain ContactFilter.
func (m *ContactMapper) SearchRequestToFilter(req *dto.SearchContactsRequest) domain.ContactFilter {
	return domain.ContactFilter{
		Query:               req.Query,
		CustomerID:          req.CustomerID,
		Statuses:            req.Statuses,
		Roles:               req.Roles,
		IsPrimary:           req.IsPrimary,
		HasEmail:            req.HasEmail,
		HasPhone:            req.HasPhone,
		OptedInMarketing:    req.OptedInMarketing,
		Tags:                req.Tags,
		MinEngagement:       req.MinEngagement,
		MaxEngagement:       req.MaxEngagement,
		NeedsFollowUp:       req.NeedsFollowUp,
		CreatedAfter:        req.CreatedAfter,
		CreatedBefore:       req.CreatedBefore,
		LastContactedAfter:  req.LastContactedAfter,
		LastContactedBefore: req.LastContactedBefore,
		IncludeDeleted:      req.IncludeDeleted,
	}
}

// PaginationFromSearchRequest extracts pagination from search request.
func (m *ContactMapper) PaginationFromSearchRequest(req *dto.SearchContactsRequest) (offset, limit int, sortBy, sortOrder string) {
	offset = req.Offset
	limit = req.Limit
	if limit == 0 {
		limit = 20 // default limit
	}
	sortBy = req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder = req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	return
}

// ============================================================================
// Activity Mappings
// ============================================================================

// ActivityToResponse maps contact activity to response.
func (m *ContactMapper) ActivityToResponse(activity *domain.ContactActivity) *dto.ContactActivityResponse {
	if activity == nil {
		return nil
	}

	return &dto.ContactActivityResponse{
		ID:          activity.ID,
		ContactID:   activity.ContactID,
		Type:        activity.Type,
		Description: activity.Description,
		Metadata:    activity.Metadata,
		PerformedBy: activity.PerformedBy,
		PerformedAt: activity.PerformedAt,
	}
}

// ActivitiesToListResponse maps activities to list response.
func (m *ContactMapper) ActivitiesToListResponse(activities []*domain.ContactActivity, total int64, offset, limit int) *dto.ContactActivityListResponse {
	responses := make([]dto.ContactActivityResponse, len(activities))
	for i, activity := range activities {
		if resp := m.ActivityToResponse(activity); resp != nil {
			responses[i] = *resp
		}
	}

	return &dto.ContactActivityListResponse{
		Activities: responses,
		Total:      total,
		Offset:     offset,
		Limit:      limit,
		HasMore:    int64(offset+len(activities)) < total,
	}
}

// ============================================================================
// Statistics Mappings
// ============================================================================

// StatsToResponse maps contact statistics to response.
func (m *ContactMapper) StatsToResponse(stats *domain.ContactStats) *dto.ContactStatsResponse {
	if stats == nil {
		return nil
	}

	return &dto.ContactStatsResponse{
		TotalContacts:         stats.TotalContacts,
		ActiveContacts:        stats.ActiveContacts,
		InactiveContacts:      stats.InactiveContacts,
		BlockedContacts:       stats.BlockedContacts,
		PrimaryContacts:       stats.PrimaryContacts,
		MarketingOptInCount:   stats.MarketingOptInCount,
		MarketingOptOutCount:  stats.MarketingOptOutCount,
		AvgEngagementScore:    stats.AvgEngagementScore,
		ContactsNeedingFollowUp: stats.ContactsNeedingFollowUp,
		RoleDistribution:      stats.RoleDistribution,
		DepartmentDistribution: stats.DepartmentDistribution,
		LastCalculatedAt:      stats.LastCalculatedAt,
	}
}

// ============================================================================
// Duplicate Detection Mappings
// ============================================================================

// DuplicatesToResponse maps duplicate matches to response.
func (m *ContactMapper) DuplicatesToResponse(duplicates []domain.DuplicateContactMatch) *dto.FindDuplicateContactsResponse {
	responses := make([]dto.DuplicateContactResponse, len(duplicates))
	for i, dup := range duplicates {
		responses[i] = dto.DuplicateContactResponse{
			Contact:     *m.ToSummaryResponse(dup.Contact),
			MatchScore:  dup.Score,
			MatchFields: dup.MatchFields,
			MatchReason: dup.Reason,
		}
	}

	return &dto.FindDuplicateContactsResponse{
		Duplicates: responses,
		Total:      len(responses),
	}
}

// ============================================================================
// Import/Export Mappings
// ============================================================================

// ImportResultToResponse maps import result to response.
func (m *ContactMapper) ImportResultToResponse(result *domain.ContactImportResult) *dto.ContactImportResponse {
	if result == nil {
		return nil
	}

	results := make([]dto.ContactImportResult, len(result.Results))
	for i, r := range result.Results {
		results[i] = dto.ContactImportResult{
			Row:       r.Row,
			Success:   r.Success,
			ContactID: r.ContactID,
			Errors:    r.Errors,
			Warnings:  r.Warnings,
		}
	}

	return &dto.ContactImportResponse{
		TotalRows:    result.TotalRows,
		SuccessCount: result.SuccessCount,
		FailureCount: result.FailureCount,
		SkippedCount: result.SkippedCount,
		UpdatedCount: result.UpdatedCount,
		Results:      results,
		Errors:       result.Errors,
	}
}
