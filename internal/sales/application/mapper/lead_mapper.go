// Package mapper provides mapping functions between domain entities and DTOs.
package mapper

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Lead Mappers
// ============================================================================

// LeadMapper handles mapping between Lead entities and DTOs.
type LeadMapper struct{}

// NewLeadMapper creates a new LeadMapper instance.
func NewLeadMapper() *LeadMapper {
	return &LeadMapper{}
}

// ToResponse maps a Lead entity to LeadResponse DTO.
func (m *LeadMapper) ToResponse(lead *domain.Lead) *dto.LeadResponse {
	if lead == nil {
		return nil
	}

	response := &dto.LeadResponse{
		ID:        lead.ID.String(),
		TenantID:  lead.TenantID.String(),
		FirstName: lead.Contact.FirstName,
		LastName:  lead.Contact.LastName,
		FullName:  lead.Contact.FullName(),
		Email:     lead.Contact.Email,
		Status:    string(lead.Status),
		Source:    string(lead.Source),
		Rating:    string(lead.Rating),
		Score:     lead.Score.Score,
		DemographicScore: lead.Score.Demographic,
		BehavioralScore:  lead.Score.Behavioral,
		Tags:             lead.Tags,
		CustomFields:     lead.CustomFields,
		CreatedAt:        lead.CreatedAt,
		UpdatedAt:        lead.UpdatedAt,
		CreatedBy:        lead.CreatedBy.String(),
		Version:          lead.Version,
	}

	// Contact details
	if lead.Contact.Phone != "" {
		response.Phone = dto.StringPtr(lead.Contact.Phone)
	}
	if lead.Contact.Mobile != "" {
		response.Mobile = dto.StringPtr(lead.Contact.Mobile)
	}
	if lead.Contact.JobTitle != "" {
		response.JobTitle = dto.StringPtr(lead.Contact.JobTitle)
	}
	if lead.Contact.Department != "" {
		response.Department = dto.StringPtr(lead.Contact.Department)
	}

	// Company details
	if lead.Company.Name != "" {
		response.Company = dto.StringPtr(lead.Company.Name)
	}
	if lead.Company.Size != "" {
		response.CompanySize = dto.StringPtr(lead.Company.Size)
	}
	if lead.Company.Industry != "" {
		response.Industry = dto.StringPtr(lead.Company.Industry)
	}
	if lead.Company.Website != "" {
		response.Website = dto.StringPtr(lead.Company.Website)
	}

	// Address
	if lead.Company.Address != "" || lead.Company.City != "" || lead.Company.Country != "" {
		response.Address = &dto.AddressDTO{
			Street1: lead.Company.Address,
			City:    lead.Company.City,
			Country: lead.Company.Country,
		}
		if lead.Company.State != "" {
			response.Address.State = dto.StringPtr(lead.Company.State)
		}
		if lead.Company.PostalCode != "" {
			response.Address.PostalCode = dto.StringPtr(lead.Company.PostalCode)
		}
	}

	// Owner
	if lead.OwnerID != nil {
		ownerIDStr := lead.OwnerID.String()
		response.OwnerID = &ownerIDStr
	}

	// Campaign
	if lead.CampaignID != nil {
		campaignIDStr := lead.CampaignID.String()
		response.CampaignID = &campaignIDStr
	}

	// Description and Notes
	if lead.Description != "" {
		response.Description = dto.StringPtr(lead.Description)
	}

	// Budget/Estimated Value
	if lead.EstimatedValue.Amount > 0 {
		response.Budget = &dto.MoneyDTO{
			Amount:   lead.EstimatedValue.Amount,
			Currency: lead.EstimatedValue.Currency,
			Display:  formatMoney(lead.EstimatedValue),
		}
	}

	// Activity information
	if lead.LastContactedAt != nil {
		response.LastContactedAt = lead.LastContactedAt
	}

	// Engagement metrics as activity count
	response.ActivityCount = lead.Engagement.EmailsOpened +
		lead.Engagement.EmailsClicked +
		lead.Engagement.WebVisits +
		lead.Engagement.FormSubmissions

	if lead.Engagement.LastEngagement != nil {
		response.LastActivityAt = lead.Engagement.LastEngagement
	}

	// Conversion info
	if lead.ConversionInfo != nil {
		response.ConvertedAt = &lead.ConversionInfo.ConvertedAt
		convertedByStr := lead.ConversionInfo.ConvertedBy.String()
		response.ConvertedBy = &convertedByStr
		oppIDStr := lead.ConversionInfo.OpportunityID.String()
		response.OpportunityID = &oppIDStr
		if lead.ConversionInfo.CustomerID != nil {
			custIDStr := lead.ConversionInfo.CustomerID.String()
			response.CustomerID = &custIDStr
		}
		if lead.ConversionInfo.ContactID != nil {
			contIDStr := lead.ConversionInfo.ContactID.String()
			response.ContactID = &contIDStr
		}
	}

	// Disqualification info
	if lead.DisqualifyInfo != nil {
		response.DisqualifiedAt = &lead.DisqualifyInfo.DisqualifiedAt
		disqualifiedByStr := lead.DisqualifyInfo.DisqualifiedBy.String()
		response.DisqualifiedBy = &disqualifiedByStr
		response.DisqualifyReason = dto.StringPtr(lead.DisqualifyInfo.Reason)
	}

	return response
}

// ToBriefResponse maps a Lead entity to LeadBriefResponse DTO.
func (m *LeadMapper) ToBriefResponse(lead *domain.Lead) *dto.LeadBriefResponse {
	if lead == nil {
		return nil
	}

	response := &dto.LeadBriefResponse{
		ID:        lead.ID.String(),
		FirstName: lead.Contact.FirstName,
		LastName:  lead.Contact.LastName,
		FullName:  lead.Contact.FullName(),
		Email:     lead.Contact.Email,
		Status:    string(lead.Status),
		Score:     lead.Score.Score,
		Rating:    string(lead.Rating),
		CreatedAt: lead.CreatedAt,
	}

	if lead.Company.Name != "" {
		response.Company = dto.StringPtr(lead.Company.Name)
	}

	if lead.OwnerID != nil {
		ownerIDStr := lead.OwnerID.String()
		response.OwnerID = &ownerIDStr
	}

	return response
}

// ToListResponse maps a slice of Lead entities to LeadListResponse DTO.
func (m *LeadMapper) ToListResponse(leads []*domain.Lead, page, pageSize int, totalItems int64) *dto.LeadListResponse {
	briefResponses := make([]*dto.LeadBriefResponse, 0, len(leads))
	for _, lead := range leads {
		briefResponses = append(briefResponses, m.ToBriefResponse(lead))
	}

	return &dto.LeadListResponse{
		Leads:      briefResponses,
		Pagination: dto.NewPaginationResponse(page, pageSize, totalItems),
	}
}

// ToContact maps CreateLeadRequest to LeadContact value object.
func (m *LeadMapper) ToContact(req *dto.CreateLeadRequest) domain.LeadContact {
	contact := domain.LeadContact{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	if req.Phone != nil {
		contact.Phone = *req.Phone
	}
	if req.Mobile != nil {
		contact.Mobile = *req.Mobile
	}
	if req.JobTitle != nil {
		contact.JobTitle = *req.JobTitle
	}
	if req.Department != nil {
		contact.Department = *req.Department
	}

	return contact
}

// ToCompany maps CreateLeadRequest to LeadCompany value object.
func (m *LeadMapper) ToCompany(req *dto.CreateLeadRequest) domain.LeadCompany {
	company := domain.LeadCompany{}

	if req.Company != nil {
		company.Name = *req.Company
	}
	if req.CompanySize != nil {
		company.Size = *req.CompanySize
	}
	if req.Industry != nil {
		company.Industry = *req.Industry
	}
	if req.Website != nil {
		company.Website = *req.Website
	}

	if req.Address != nil {
		company.Address = req.Address.Street1
		company.City = req.Address.City
		company.Country = req.Address.Country
		if req.Address.State != nil {
			company.State = *req.Address.State
		}
		if req.Address.PostalCode != nil {
			company.PostalCode = *req.Address.PostalCode
		}
	}

	return company
}

// ToSource maps source string to LeadSource domain type.
func (m *LeadMapper) ToSource(source string) domain.LeadSource {
	switch source {
	case "website":
		return domain.LeadSourceWebsite
	case "referral":
		return domain.LeadSourceReferral
	case "social_media":
		return domain.LeadSourceSocialMedia
	case "advertising":
		return domain.LeadSourceAdvertising
	case "trade_show":
		return domain.LeadSourceTradeShow
	case "cold_call":
		return domain.LeadSourceColdCall
	case "email":
		return domain.LeadSourceEmail
	case "email_campaign":
		return domain.LeadSourceEmail
	case "partner":
		return domain.LeadSourcePartner
	default:
		return domain.LeadSourceOther
	}
}

// ToRating maps rating string to LeadRating domain type.
func (m *LeadMapper) ToRating(rating string) domain.LeadRating {
	switch rating {
	case "hot":
		return domain.LeadRatingHot
	case "warm":
		return domain.LeadRatingWarm
	case "cold":
		return domain.LeadRatingCold
	default:
		return domain.LeadRatingCold
	}
}

// UpdateContactFromRequest updates a LeadContact from UpdateLeadRequest.
func (m *LeadMapper) UpdateContactFromRequest(contact *domain.LeadContact, req *dto.UpdateLeadRequest) {
	if req.FirstName != nil {
		contact.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		contact.LastName = *req.LastName
	}
	if req.Email != nil {
		contact.Email = *req.Email
	}
	if req.Phone != nil {
		contact.Phone = *req.Phone
	}
	if req.Mobile != nil {
		contact.Mobile = *req.Mobile
	}
	if req.JobTitle != nil {
		contact.JobTitle = *req.JobTitle
	}
	if req.Department != nil {
		contact.Department = *req.Department
	}
}

// UpdateCompanyFromRequest updates a LeadCompany from UpdateLeadRequest.
func (m *LeadMapper) UpdateCompanyFromRequest(company *domain.LeadCompany, req *dto.UpdateLeadRequest) {
	if req.Company != nil {
		company.Name = *req.Company
	}
	if req.CompanySize != nil {
		company.Size = *req.CompanySize
	}
	if req.Industry != nil {
		company.Industry = *req.Industry
	}
	if req.Website != nil {
		company.Website = *req.Website
	}

	if req.Address != nil {
		company.Address = req.Address.Street1
		company.City = req.Address.City
		company.Country = req.Address.Country
		if req.Address.State != nil {
			company.State = *req.Address.State
		}
		if req.Address.PostalCode != nil {
			company.PostalCode = *req.Address.PostalCode
		}
	}
}

// ToEstimatedValue maps budget fields to Money value object.
func (m *LeadMapper) ToEstimatedValue(amount *int64, currency *string) (domain.Money, error) {
	if amount == nil || *amount == 0 {
		return domain.Money{}, nil
	}

	curr := "USD"
	if currency != nil && *currency != "" {
		curr = *currency
	}

	return domain.NewMoney(*amount, curr)
}

// ============================================================================
// Helper Functions
// ============================================================================

// formatMoney formats a Money value object to display string.
func formatMoney(money domain.Money) string {
	// Simple formatting - can be enhanced based on currency
	majorUnits := float64(money.Amount) / 100
	return fmt.Sprintf("%s %.2f", money.Currency, majorUnits)
}

// parseUUID parses a string to UUID, returning nil UUID if empty.
func parseUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// parseUUIDPtr parses a string to UUID pointer, returning nil if empty.
func parseUUIDPtr(s *string) *uuid.UUID {
	if s == nil || *s == "" {
		return nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil
	}
	return &id
}

// uuidToString converts a UUID to string, returning empty string for nil UUID.
func uuidToString(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

// uuidPtrToString converts a UUID pointer to string pointer.
func uuidPtrToString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
