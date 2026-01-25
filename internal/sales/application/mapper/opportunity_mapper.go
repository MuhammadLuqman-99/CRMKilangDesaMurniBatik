// Package mapper provides mapping functions between domain entities and DTOs.
package mapper

import (
	"time"

	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Opportunity Mappers
// ============================================================================

// OpportunityMapper handles mapping between Opportunity entities and DTOs.
type OpportunityMapper struct {
	leadMapper *LeadMapper
}

// NewOpportunityMapper creates a new OpportunityMapper instance.
func NewOpportunityMapper() *OpportunityMapper {
	return &OpportunityMapper{
		leadMapper: NewLeadMapper(),
	}
}

// ToResponse maps an Opportunity entity to OpportunityResponse DTO.
func (m *OpportunityMapper) ToResponse(opp *domain.Opportunity) *dto.OpportunityResponse {
	if opp == nil {
		return nil
	}

	response := &dto.OpportunityResponse{
		ID:          opp.ID.String(),
		TenantID:    opp.TenantID.String(),
		Name:        opp.Name,
		Status:      string(opp.Status),
		PipelineID:  opp.PipelineID.String(),
		StageID:     opp.StageID.String(),
		Probability: opp.Probability,
		Amount: dto.MoneyDTO{
			Amount:   opp.Amount.Amount,
			Currency: opp.Amount.Currency,
			Display:  formatMoney(opp.Amount),
		},
		WeightedAmount: dto.MoneyDTO{
			Amount:   opp.WeightedAmount.Amount,
			Currency: opp.WeightedAmount.Currency,
			Display:  formatMoney(opp.WeightedAmount),
		},
		OwnerID:       opp.OwnerID.String(),
		Tags:          opp.Tags,
		CustomFields:  opp.CustomFields,
		ActivityCount: opp.ActivityCount,
		DaysInStage:   opp.DaysInCurrentStage(),
		DaysOpen:      opp.DaysInPipeline(),
		CreatedAt:     opp.CreatedAt,
		UpdatedAt:     opp.UpdatedAt,
		CreatedBy:     opp.CreatedBy.String(),
		Version:       opp.Version,
	}

	// Description
	if opp.Description != "" {
		response.Description = dto.StringPtr(opp.Description)
	}

	// Notes
	if opp.Notes != "" {
		response.Notes = dto.StringPtr(opp.Notes)
	}

	// Expected close date
	if opp.ExpectedCloseDate != nil {
		response.ExpectedCloseDate = *opp.ExpectedCloseDate
	}

	// Actual close date
	response.ActualCloseDate = opp.ActualCloseDate

	// Customer
	customerIDStr := opp.CustomerID.String()
	response.CustomerID = &customerIDStr
	response.Customer = &dto.CustomerBriefDTO{
		ID:   opp.CustomerID.String(),
		Name: opp.CustomerName,
	}

	// Lead
	if opp.LeadID != nil {
		leadIDStr := opp.LeadID.String()
		response.LeadID = &leadIDStr
	}

	// Source
	response.Source = opp.Source
	if opp.Campaign != "" {
		response.SourceDetails = dto.StringPtr(opp.Campaign)
	}
	if opp.CampaignID != nil {
		campaignIDStr := opp.CampaignID.String()
		response.CampaignID = &campaignIDStr
	}

	// Pipeline brief
	response.Pipeline = &dto.PipelineBriefDTO{
		ID:   opp.PipelineID.String(),
		Name: opp.PipelineName,
	}

	// Stage brief
	response.Stage = &dto.StageBriefDTO{
		ID:          opp.StageID.String(),
		Name:        opp.StageName,
		Probability: opp.Probability,
	}

	// Stage history
	response.StageHistory = m.mapStageHistory(opp.StageHistory)

	// Contacts
	response.Contacts = m.mapContacts(opp.Contacts)
	if len(opp.Contacts) > 0 {
		for _, c := range opp.Contacts {
			if c.IsPrimary {
				contactIDStr := c.ContactID.String()
				response.PrimaryContactID = &contactIDStr
				response.PrimaryContact = &dto.ContactBriefDTO{
					ID:       c.ContactID.String(),
					FullName: c.Name,
					Email:    c.Email,
				}
				if c.Phone != "" {
					response.PrimaryContact.Phone = dto.StringPtr(c.Phone)
				}
				break
			}
		}
	}

	// Products
	response.Products = m.mapProducts(opp.Products)
	response.ProductCount = len(opp.Products)

	// Owner
	response.Owner = &dto.UserBriefDTO{
		ID:   opp.OwnerID.String(),
		Name: opp.OwnerName,
	}

	// Activity tracking
	response.LastActivityAt = opp.LastActivityAt

	// Win/Loss information
	if opp.CloseInfo != nil {
		if opp.Status == domain.OpportunityStatusWon {
			response.WonAt = &opp.CloseInfo.ClosedAt
			wonByStr := opp.CloseInfo.ClosedBy.String()
			response.WonBy = &wonByStr
			response.WonReason = dto.StringPtr(opp.CloseInfo.Reason)
			if opp.CloseInfo.Notes != "" {
				response.WonNotes = dto.StringPtr(opp.CloseInfo.Notes)
			}
		} else if opp.Status == domain.OpportunityStatusLost {
			response.LostAt = &opp.CloseInfo.ClosedAt
			lostByStr := opp.CloseInfo.ClosedBy.String()
			response.LostBy = &lostByStr
			response.LostReason = dto.StringPtr(opp.CloseInfo.Reason)
			if opp.CloseInfo.Notes != "" {
				response.LostNotes = dto.StringPtr(opp.CloseInfo.Notes)
			}
			if opp.CloseInfo.CompetitorID != nil {
				competitorIDStr := opp.CloseInfo.CompetitorID.String()
				response.CompetitorID = &competitorIDStr
			}
			if opp.CloseInfo.CompetitorName != "" {
				response.CompetitorName = dto.StringPtr(opp.CloseInfo.CompetitorName)
			}
		}
	}

	return response
}

// ToBriefResponse maps an Opportunity entity to OpportunityBriefResponse DTO.
func (m *OpportunityMapper) ToBriefResponse(opp *domain.Opportunity) *dto.OpportunityBriefResponse {
	if opp == nil {
		return nil
	}

	response := &dto.OpportunityBriefResponse{
		ID:          opp.ID.String(),
		Name:        opp.Name,
		Status:      string(opp.Status),
		Probability: opp.Probability,
		StageID:     opp.StageID.String(),
		StageName:   opp.StageName,
		OwnerID:     opp.OwnerID.String(),
		OwnerName:   opp.OwnerName,
		DaysOpen:    opp.DaysInPipeline(),
		CreatedAt:   opp.CreatedAt,
		Amount: dto.MoneyDTO{
			Amount:   opp.Amount.Amount,
			Currency: opp.Amount.Currency,
			Display:  formatMoney(opp.Amount),
		},
		WeightedAmount: dto.MoneyDTO{
			Amount:   opp.WeightedAmount.Amount,
			Currency: opp.WeightedAmount.Currency,
			Display:  formatMoney(opp.WeightedAmount),
		},
	}

	// Customer
	customerIDStr := opp.CustomerID.String()
	response.CustomerID = &customerIDStr
	response.CustomerName = dto.StringPtr(opp.CustomerName)

	// Expected close date
	if opp.ExpectedCloseDate != nil {
		response.ExpectedCloseDate = *opp.ExpectedCloseDate
	}

	return response
}

// ToListResponse maps a slice of Opportunity entities to OpportunityListResponse DTO.
func (m *OpportunityMapper) ToListResponse(
	opportunities []*domain.Opportunity,
	page, pageSize int,
	totalItems int64,
	summary *dto.OpportunitySummaryDTO,
) *dto.OpportunityListResponse {
	briefResponses := make([]*dto.OpportunityBriefResponse, 0, len(opportunities))
	for _, opp := range opportunities {
		briefResponses = append(briefResponses, m.ToBriefResponse(opp))
	}

	return &dto.OpportunityListResponse{
		Opportunities: briefResponses,
		Pagination:    dto.NewPaginationResponse(page, pageSize, totalItems),
		Summary:       summary,
	}
}

// mapStageHistory maps domain StageHistory to StageHistoryDTO.
func (m *OpportunityMapper) mapStageHistory(history []domain.StageHistory) []*dto.StageHistoryDTO {
	if len(history) == 0 {
		return nil
	}

	result := make([]*dto.StageHistoryDTO, 0, len(history))
	for _, h := range history {
		entry := &dto.StageHistoryDTO{
			StageID:   h.StageID.String(),
			StageName: h.StageName,
			EnteredAt: h.EnteredAt,
			ChangedBy: h.MovedBy.String(),
		}

		if h.ExitedAt != nil {
			entry.ExitedAt = h.ExitedAt
		}

		if h.Duration > 0 {
			durationSeconds := h.Duration * 3600 // Convert hours to seconds
			entry.Duration = &durationSeconds
		}

		if h.Notes != "" {
			entry.Notes = dto.StringPtr(h.Notes)
		}

		result = append(result, entry)
	}

	return result
}

// mapContacts maps domain OpportunityContact to OpportunityContactResponseDTO.
func (m *OpportunityMapper) mapContacts(contacts []domain.OpportunityContact) []*dto.OpportunityContactResponseDTO {
	if len(contacts) == 0 {
		return nil
	}

	result := make([]*dto.OpportunityContactResponseDTO, 0, len(contacts))
	for _, c := range contacts {
		contactDTO := &dto.OpportunityContactResponseDTO{
			ContactID: c.ContactID.String(),
			Role:      c.Role,
			IsPrimary: c.IsPrimary,
			Contact: &dto.ContactBriefDTO{
				ID:       c.ContactID.String(),
				FullName: c.Name,
				Email:    c.Email,
			},
		}

		if c.Phone != "" {
			contactDTO.Contact.Phone = dto.StringPtr(c.Phone)
		}

		result = append(result, contactDTO)
	}

	return result
}

// mapProducts maps domain OpportunityProduct to OpportunityProductResponseDTO.
func (m *OpportunityMapper) mapProducts(products []domain.OpportunityProduct) []*dto.OpportunityProductResponseDTO {
	if len(products) == 0 {
		return nil
	}

	result := make([]*dto.OpportunityProductResponseDTO, 0, len(products))
	for _, p := range products {
		discountAmount := p.UnitPrice.Multiply(float64(p.Quantity))
		discountAmount = discountAmount.Multiply(p.Discount / 100)

		productDTO := &dto.OpportunityProductResponseDTO{
			ID:          p.ID.String(),
			ProductID:   p.ProductID.String(),
			ProductName: p.ProductName,
			Quantity:    p.Quantity,
			UnitPrice: dto.MoneyDTO{
				Amount:   p.UnitPrice.Amount,
				Currency: p.UnitPrice.Currency,
				Display:  formatMoney(p.UnitPrice),
			},
			DiscountPercent: int(p.Discount),
			DiscountAmount: dto.MoneyDTO{
				Amount:   discountAmount.Amount,
				Currency: discountAmount.Currency,
				Display:  formatMoney(discountAmount),
			},
			TotalPrice: dto.MoneyDTO{
				Amount:   p.TotalPrice.Amount,
				Currency: p.TotalPrice.Currency,
				Display:  formatMoney(p.TotalPrice),
			},
		}

		if p.Notes != "" {
			productDTO.Description = dto.StringPtr(p.Notes)
		}

		result = append(result, productDTO)
	}

	return result
}

// ToPriority maps priority string to OpportunityPriority domain type.
func (m *OpportunityMapper) ToPriority(priority string) domain.OpportunityPriority {
	switch priority {
	case "low":
		return domain.OpportunityPriorityLow
	case "medium":
		return domain.OpportunityPriorityMedium
	case "high":
		return domain.OpportunityPriorityHigh
	case "critical":
		return domain.OpportunityPriorityCritical
	default:
		return domain.OpportunityPriorityMedium
	}
}

// ToProduct maps AddProductRequest to OpportunityProduct domain type.
func (m *OpportunityMapper) ToProduct(req *dto.AddProductRequest) (domain.OpportunityProduct, error) {
	productID, err := dto.ParseUUIDRequired(req.ProductID)
	if err != nil {
		return domain.OpportunityProduct{}, err
	}

	unitPrice, err := domain.NewMoney(req.UnitPrice, req.Currency)
	if err != nil {
		return domain.OpportunityProduct{}, err
	}

	product := domain.OpportunityProduct{
		ProductID:   productID,
		ProductName: req.ProductName,
		Quantity:    req.Quantity,
		UnitPrice:   unitPrice,
	}

	if req.DiscountPercent != nil {
		product.Discount = float64(*req.DiscountPercent)
	}

	if req.Description != nil {
		product.Notes = *req.Description
	}

	// Calculate total price
	product.CalculateTotalPrice()

	return product, nil
}

// ToContact maps AddContactRequest to OpportunityContact domain type.
func (m *OpportunityMapper) ToContact(req *dto.AddContactRequest, contactName, contactEmail, contactPhone string) (domain.OpportunityContact, error) {
	contactID, err := dto.ParseUUIDRequired(req.ContactID)
	if err != nil {
		return domain.OpportunityContact{}, err
	}

	return domain.OpportunityContact{
		ContactID: contactID,
		Name:      contactName,
		Email:     contactEmail,
		Phone:     contactPhone,
		Role:      req.Role,
		IsPrimary: req.IsPrimary,
	}, nil
}

// ParseExpectedCloseDate parses expected close date string to time.Time.
func (m *OpportunityMapper) ParseExpectedCloseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// ToWinResponse creates an OpportunityWinResponse DTO.
func (m *OpportunityMapper) ToWinResponse(opp *domain.Opportunity, dealID *string) *dto.OpportunityWinResponse {
	return &dto.OpportunityWinResponse{
		OpportunityID: opp.ID.String(),
		Status:        string(opp.Status),
		DealID:        dealID,
		Message:       "Opportunity marked as won successfully",
	}
}

// ToLoseResponse creates an OpportunityLoseResponse DTO.
func (m *OpportunityMapper) ToLoseResponse(opp *domain.Opportunity) *dto.OpportunityLoseResponse {
	return &dto.OpportunityLoseResponse{
		OpportunityID: opp.ID.String(),
		Status:        string(opp.Status),
		Message:       "Opportunity marked as lost",
	}
}
