// Package mapper provides functions to map between domain entities and DTOs.
package mapper

import (
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// CustomerMapper provides mapping functions for Customer entities.
type CustomerMapper struct{}

// NewCustomerMapper creates a new CustomerMapper.
func NewCustomerMapper() *CustomerMapper {
	return &CustomerMapper{}
}

// ============================================================================
// Domain to DTO Mappings
// ============================================================================

// ToResponse maps a Customer domain entity to CustomerResponse DTO.
func (m *CustomerMapper) ToResponse(customer *domain.Customer) *dto.CustomerResponse {
	if customer == nil {
		return nil
	}

	response := &dto.CustomerResponse{
		ID:              customer.ID,
		TenantID:        customer.TenantID,
		Code:            customer.Code,
		Name:            customer.Name,
		Type:            customer.Type,
		Status:          customer.Status,
		Tier:            customer.Tier,
		Source:          customer.Source,
		Email:           customer.Email.String(),
		PhoneNumbers:    m.phonesToResponse(customer.PhoneNumbers),
		Website:         customer.Website.String(),
		Addresses:       m.addressesToResponse(customer.Addresses),
		SocialProfiles:  m.socialProfilesToResponse(customer.SocialProfiles),
		CompanyInfo:     m.companyInfoToResponse(&customer.CompanyInfo),
		Financials:      m.financialsToResponse(&customer.Financials),
		Preferences:     m.preferencesToResponse(&customer.Preferences),
		Stats:           m.statsToResponse(&customer.Stats),
		Contacts:        m.contactSummariesToResponse(customer.Contacts),
		OwnerID:         customer.OwnerID,
		AssignedTeam:    customer.AssignedTeam,
		Tags:            customer.Tags,
		Segments:        customer.Segments,
		CustomFields:    customer.CustomFields,
		Notes:           customer.Notes,
		LogoURL:         customer.LogoURL,
		LastContactedAt: customer.LastContactedAt,
		NextFollowUpAt:  customer.NextFollowUpAt,
		ConvertedAt:     customer.ConvertedAt,
		ChurnedAt:       customer.ChurnedAt,
		ChurnReason:     customer.ChurnReason,
		Version:         customer.Version,
		CreatedAt:       customer.CreatedAt,
		UpdatedAt:       customer.UpdatedAt,
		CreatedBy:       customer.AuditInfo.CreatedBy,
		UpdatedBy:       customer.AuditInfo.UpdatedBy,
	}

	return response
}

// ToSummaryResponse maps a Customer to CustomerSummaryResponse.
func (m *CustomerMapper) ToSummaryResponse(customer *domain.Customer) *dto.CustomerSummaryResponse {
	if customer == nil {
		return nil
	}

	var phone string
	if primaryPhone := customer.GetPrimaryPhone(); primaryPhone != nil {
		phone = primaryPhone.Formatted()
	}

	return &dto.CustomerSummaryResponse{
		ID:              customer.ID,
		Code:            customer.Code,
		Name:            customer.Name,
		Type:            customer.Type,
		Status:          customer.Status,
		Tier:            customer.Tier,
		Email:           customer.Email.String(),
		Phone:           phone,
		OwnerID:         customer.OwnerID,
		ContactCount:    len(customer.Contacts),
		Tags:            customer.Tags,
		LastContactedAt: customer.LastContactedAt,
		CreatedAt:       customer.CreatedAt,
	}
}

// ToListResponse maps a slice of Customers to CustomerListResponse.
func (m *CustomerMapper) ToListResponse(customers []*domain.Customer, total int64, offset, limit int) *dto.CustomerListResponse {
	summaries := make([]dto.CustomerSummaryResponse, len(customers))
	for i, customer := range customers {
		if summary := m.ToSummaryResponse(customer); summary != nil {
			summaries[i] = *summary
		}
	}

	return &dto.CustomerListResponse{
		Customers: summaries,
		Total:     total,
		Offset:    offset,
		Limit:     limit,
		HasMore:   int64(offset+len(customers)) < total,
	}
}

// ============================================================================
// Value Object to DTO Mappings
// ============================================================================

func (m *CustomerMapper) phonesToResponse(phones []domain.PhoneNumber) []dto.PhoneResponse {
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

func (m *CustomerMapper) addressesToResponse(addresses []domain.Address) []dto.AddressResponse {
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

func (m *CustomerMapper) socialProfilesToResponse(profiles []domain.SocialProfile) []dto.SocialProfileResponse {
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

func (m *CustomerMapper) companyInfoToResponse(info *domain.CompanyInfo) *dto.CompanyInfoResponse {
	if info == nil || info.LegalName == "" {
		return nil
	}

	response := &dto.CompanyInfoResponse{
		LegalName:          info.LegalName,
		TradingName:        info.TradingName,
		RegistrationNumber: info.RegistrationNumber,
		TaxID:              info.TaxID,
		Industry:           info.Industry,
		Size:               info.Size,
		EmployeeCount:      info.EmployeeCount,
		FoundedYear:        info.FoundedYear,
		Description:        info.Description,
		ParentCompanyID:    info.ParentCompanyID,
	}

	if info.AnnualRevenue != nil && info.AnnualRevenue.Amount > 0 {
		response.AnnualRevenue = &dto.MoneyResponse{
			Amount:   info.AnnualRevenue.Amount,
			Currency: info.AnnualRevenue.Currency,
			Display:  info.AnnualRevenue.Display(),
		}
	}

	return response
}

func (m *CustomerMapper) financialsToResponse(financials *domain.CustomerFinancials) dto.CustomerFinancialsResponse {
	response := dto.CustomerFinancialsResponse{
		Currency:           financials.Currency,
		PaymentTerms:       financials.PaymentTerms,
		TaxExempt:          financials.TaxExempt,
		TaxExemptionID:     financials.TaxExemptionID,
		BillingEmail:       financials.BillingEmail.String(),
		DefaultDiscountPct: financials.DefaultDiscountPct,
		LastPaymentAt:      financials.LastPaymentAt,
		LastPurchaseAt:     financials.LastPurchaseAt,
		TotalPurchases:     financials.TotalPurchases,
	}

	if financials.CreditLimit != nil {
		response.CreditLimit = &dto.MoneyResponse{
			Amount:   financials.CreditLimit.Amount,
			Currency: financials.CreditLimit.Currency,
			Display:  financials.CreditLimit.Display(),
		}
	}

	if financials.CurrentBalance != nil {
		response.CurrentBalance = &dto.MoneyResponse{
			Amount:   financials.CurrentBalance.Amount,
			Currency: financials.CurrentBalance.Currency,
			Display:  financials.CurrentBalance.Display(),
		}
	}

	if financials.LifetimeValue != nil {
		response.LifetimeValue = &dto.MoneyResponse{
			Amount:   financials.LifetimeValue.Amount,
			Currency: financials.LifetimeValue.Currency,
			Display:  financials.LifetimeValue.Display(),
		}
	}

	if financials.TotalSpent != nil {
		response.TotalSpent = &dto.MoneyResponse{
			Amount:   financials.TotalSpent.Amount,
			Currency: financials.TotalSpent.Currency,
			Display:  financials.TotalSpent.Display(),
		}
	}

	return response
}

func (m *CustomerMapper) preferencesToResponse(prefs *domain.CustomerPreferences) dto.CustomerPreferencesResponse {
	return dto.CustomerPreferencesResponse{
		Language:          prefs.Language,
		Timezone:          prefs.Timezone,
		Currency:          prefs.Currency,
		DateFormat:        prefs.DateFormat,
		CommPreference:    prefs.CommPreference,
		OptedOutMarketing: prefs.OptedOutMarketing,
		MarketingConsent:  prefs.MarketingConsent,
		NewsletterOptIn:   prefs.NewsletterOptIn,
		SMSOptIn:          prefs.SMSOptIn,
	}
}

func (m *CustomerMapper) statsToResponse(stats *domain.CustomerStats) dto.CustomerStatsResponse {
	response := dto.CustomerStatsResponse{
		ContactCount:         stats.ContactCount,
		ActiveContactCount:   stats.ActiveContactCount,
		NoteCount:            stats.NoteCount,
		ActivityCount:        stats.ActivityCount,
		DealCount:            stats.DealCount,
		WonDealCount:         stats.WonDealCount,
		LostDealCount:        stats.LostDealCount,
		DaysSinceLastContact: stats.DaysSinceLastContact,
		EngagementScore:      stats.EngagementScore,
		HealthScore:          stats.HealthScore,
		LastCalculatedAt:     stats.LastCalculatedAt,
	}

	if stats.OpenDealValue != nil {
		response.OpenDealValue = &dto.MoneyResponse{
			Amount:   stats.OpenDealValue.Amount,
			Currency: stats.OpenDealValue.Currency,
			Display:  stats.OpenDealValue.Display(),
		}
	}

	if stats.AvgDealSize != nil {
		response.AvgDealSize = &dto.MoneyResponse{
			Amount:   stats.AvgDealSize.Amount,
			Currency: stats.AvgDealSize.Currency,
			Display:  stats.AvgDealSize.Display(),
		}
	}

	return response
}

func (m *CustomerMapper) contactSummariesToResponse(contacts []domain.Contact) []dto.ContactSummaryResponse {
	if len(contacts) == 0 {
		return nil
	}

	responses := make([]dto.ContactSummaryResponse, len(contacts))
	contactMapper := NewContactMapper()
	for i, contact := range contacts {
		if summary := contactMapper.ToSummaryResponse(&contact); summary != nil {
			responses[i] = *summary
		}
	}
	return responses
}

// ============================================================================
// DTO to Domain Mappings
// ============================================================================

// ToDomain maps CreateCustomerRequest to a domain Customer using the builder.
func (m *CustomerMapper) ToDomain(tenantID, creatorID uuid.UUID, req *dto.CreateCustomerRequest) (*domain.Customer, error) {
	builder := domain.NewCustomerBuilder(tenantID, req.Name, req.Type).
		WithSource(req.Source).
		WithTier(req.Tier)

	// Set code if provided
	if req.Code != "" {
		builder.WithCode(req.Code)
	}

	// Set email
	if req.Email != "" {
		builder.WithEmail(req.Email)
	}

	// Set phone
	if req.Phone != nil {
		builder.WithPhone(req.Phone.Number, req.Phone.Type, req.Phone.IsPrimary)
	}

	// Set website
	if req.Website != "" {
		builder.WithWebsite(req.Website)
	}

	// Set address
	if req.Address != nil {
		builder.WithAddress(
			req.Address.Line1,
			req.Address.City,
			req.Address.PostalCode,
			req.Address.CountryCode,
			req.Address.AddressType,
			req.Address.IsPrimary,
		)
	}

	// Set owner
	if req.OwnerID != nil {
		builder.WithOwner(*req.OwnerID)
	}

	// Set company info
	if req.CompanyInfo != nil {
		builder.WithCompanyInfo(domain.CompanyInfo{
			LegalName:          req.CompanyInfo.LegalName,
			TradingName:        req.CompanyInfo.TradingName,
			RegistrationNumber: req.CompanyInfo.RegistrationNumber,
			TaxID:              req.CompanyInfo.TaxID,
			Industry:           req.CompanyInfo.Industry,
			Size:               req.CompanyInfo.Size,
			EmployeeCount:      req.CompanyInfo.EmployeeCount,
			FoundedYear:        req.CompanyInfo.FoundedYear,
			Description:        req.CompanyInfo.Description,
			ParentCompanyID:    req.CompanyInfo.ParentCompanyID,
		})
	}

	// Set preferences
	if req.Preferences != nil {
		prefs := domain.CustomerPreferences{
			Language:       req.Preferences.Language,
			Timezone:       req.Preferences.Timezone,
			Currency:       req.Preferences.Currency,
			DateFormat:     req.Preferences.DateFormat,
			CommPreference: req.Preferences.CommPreference,
		}
		if req.Preferences.OptedOutMarketing != nil {
			prefs.OptedOutMarketing = *req.Preferences.OptedOutMarketing
		}
		if req.Preferences.NewsletterOptIn != nil {
			prefs.NewsletterOptIn = *req.Preferences.NewsletterOptIn
		}
		if req.Preferences.SMSOptIn != nil {
			prefs.SMSOptIn = *req.Preferences.SMSOptIn
		}
		builder.WithPreferences(prefs)
	}

	// Set tags
	if len(req.Tags) > 0 {
		builder.WithTags(req.Tags...)
	}

	// Set notes
	if req.Notes != "" {
		builder.WithNotes(req.Notes)
	}

	// Set custom fields
	for key, value := range req.CustomFields {
		builder.WithCustomField(key, value)
	}

	// Set creator
	builder.WithCreatedBy(creatorID)

	return builder.Build()
}

// ApplyUpdate applies UpdateCustomerRequest to an existing Customer.
func (m *CustomerMapper) ApplyUpdate(customer *domain.Customer, req *dto.UpdateCustomerRequest) error {
	if req.Name != nil {
		customer.UpdateName(*req.Name)
	}

	if req.Type != nil {
		customer.UpdateType(*req.Type)
	}

	if req.Email != nil {
		email, err := domain.NewEmail(*req.Email)
		if err != nil {
			return err
		}
		customer.UpdateEmail(email)
	}

	if req.Phone != nil {
		phone, err := domain.NewPhoneNumberWithPrimary(req.Phone.Number, req.Phone.Type, req.Phone.IsPrimary)
		if err != nil {
			return err
		}
		customer.AddPhoneNumber(phone)
	}

	if req.Website != nil {
		website, err := domain.NewWebsite(*req.Website)
		if err != nil {
			return err
		}
		customer.UpdateWebsite(website)
	}

	if req.Source != nil {
		customer.UpdateSource(*req.Source)
	}

	if req.Tier != nil {
		customer.UpdateTier(*req.Tier)
	}

	if req.OwnerID != nil {
		customer.AssignOwner(*req.OwnerID)
	}

	if req.Notes != nil {
		customer.UpdateNotes(*req.Notes)
	}

	if req.CompanyInfo != nil {
		customer.UpdateCompanyInfo(domain.CompanyInfo{
			LegalName:          req.CompanyInfo.LegalName,
			TradingName:        req.CompanyInfo.TradingName,
			RegistrationNumber: req.CompanyInfo.RegistrationNumber,
			TaxID:              req.CompanyInfo.TaxID,
			Industry:           req.CompanyInfo.Industry,
			Size:               req.CompanyInfo.Size,
			EmployeeCount:      req.CompanyInfo.EmployeeCount,
			FoundedYear:        req.CompanyInfo.FoundedYear,
			Description:        req.CompanyInfo.Description,
			ParentCompanyID:    req.CompanyInfo.ParentCompanyID,
		})
	}

	if req.Preferences != nil {
		prefs := customer.Preferences
		if req.Preferences.Language != "" {
			prefs.Language = req.Preferences.Language
		}
		if req.Preferences.Timezone != "" {
			prefs.Timezone = req.Preferences.Timezone
		}
		if req.Preferences.Currency != "" {
			prefs.Currency = req.Preferences.Currency
		}
		if req.Preferences.DateFormat != "" {
			prefs.DateFormat = req.Preferences.DateFormat
		}
		if req.Preferences.CommPreference != "" {
			prefs.CommPreference = req.Preferences.CommPreference
		}
		if req.Preferences.OptedOutMarketing != nil {
			prefs.OptedOutMarketing = *req.Preferences.OptedOutMarketing
		}
		if req.Preferences.NewsletterOptIn != nil {
			prefs.NewsletterOptIn = *req.Preferences.NewsletterOptIn
		}
		if req.Preferences.SMSOptIn != nil {
			prefs.SMSOptIn = *req.Preferences.SMSOptIn
		}
		customer.UpdatePreferences(prefs)
	}

	for key, value := range req.CustomFields {
		customer.SetCustomField(key, value)
	}

	return nil
}

// PhoneInputToDomain converts PhoneInput DTO to domain PhoneNumber.
func (m *CustomerMapper) PhoneInputToDomain(input *dto.PhoneInput) (domain.PhoneNumber, error) {
	return domain.NewPhoneNumberWithPrimary(input.Number, input.Type, input.IsPrimary)
}

// AddressInputToDomain converts AddressInput DTO to domain Address.
func (m *CustomerMapper) AddressInputToDomain(input *dto.AddressInput) (domain.Address, error) {
	addr, err := domain.NewAddress(input.Line1, input.City, input.PostalCode, input.CountryCode, input.AddressType)
	if err != nil {
		return domain.Address{}, err
	}

	// Set optional fields
	addr.Line2 = input.Line2
	addr.Line3 = input.Line3
	addr.State = input.State
	addr.IsPrimary = input.IsPrimary
	addr.Label = input.Label

	return addr, nil
}

// SocialProfileInputToDomain converts SocialProfileInput DTO to domain SocialProfile.
func (m *CustomerMapper) SocialProfileInputToDomain(input *dto.SocialProfileInput) (domain.SocialProfile, error) {
	profile, err := domain.NewSocialProfile(input.Platform, input.URL)
	if err != nil {
		return domain.SocialProfile{}, err
	}

	if input.DisplayName != "" {
		profile.DisplayName = input.DisplayName
	}

	return profile, nil
}

// SearchRequestToFilter converts SearchCustomersRequest to domain CustomerFilter.
func (m *CustomerMapper) SearchRequestToFilter(req *dto.SearchCustomersRequest) domain.CustomerFilter {
	filter := domain.CustomerFilter{
		Query:          req.Query,
		Types:          req.Types,
		Statuses:       req.Statuses,
		Tiers:          req.Tiers,
		Sources:        req.Sources,
		Tags:           req.Tags,
		OwnerIDs:       req.OwnerIDs,
		SegmentIDs:     req.SegmentIDs,
		Industries:     req.Industries,
		Countries:      req.Countries,
		CreatedAfter:   req.CreatedAfter,
		CreatedBefore:  req.CreatedBefore,
		IncludeDeleted: req.IncludeDeleted,
	}

	return filter
}

// PaginationFromRequest extracts pagination from search request.
func (m *CustomerMapper) PaginationFromRequest(req *dto.SearchCustomersRequest) (offset, limit int, sortBy, sortOrder string) {
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

// BulkOperationResultToResponse maps bulk operation results to response.
func (m *CustomerMapper) BulkOperationResultToResponse(processed, succeeded, failed int, errors []string) *dto.BulkOperationResponse {
	return &dto.BulkOperationResponse{
		Processed: processed,
		Succeeded: succeeded,
		Failed:    failed,
		Errors:    errors,
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// MoneyInputToDomain converts MoneyInput DTO to domain Money.
func MoneyInputToDomain(input *dto.MoneyInput) *domain.Money {
	if input == nil {
		return nil
	}
	money, _ := domain.NewMoney(input.Amount, input.Currency)
	return &money
}

// MoneyToResponse converts domain Money to MoneyResponse DTO.
func MoneyToResponse(money *domain.Money) *dto.MoneyResponse {
	if money == nil {
		return nil
	}
	return &dto.MoneyResponse{
		Amount:   money.Amount,
		Currency: money.Currency,
		Display:  money.Display(),
	}
}

// TimePtr returns a pointer to the given time.
func TimePtr(t time.Time) *time.Time {
	return &t
}

// IntPtr returns a pointer to the given int.
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to the given float64.
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr returns a pointer to the given bool.
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string {
	return &s
}
