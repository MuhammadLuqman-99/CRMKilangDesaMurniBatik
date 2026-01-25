package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application"
	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
	"github.com/kilang-desa-murni/crm/internal/sales/application/ports"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Lead Use Case Interface
// ============================================================================

// LeadUseCase defines the interface for lead operations.
type LeadUseCase interface {
	// CRUD operations
	Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateLeadRequest) (*dto.LeadResponse, error)
	GetByID(ctx context.Context, tenantID, leadID uuid.UUID) (*dto.LeadResponse, error)
	Update(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.UpdateLeadRequest) (*dto.LeadResponse, error)
	Delete(ctx context.Context, tenantID, leadID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *dto.LeadFilterRequest) (*dto.LeadListResponse, error)

	// Status operations
	Qualify(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.QualifyLeadRequest) (*dto.LeadResponse, error)
	Disqualify(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.DisqualifyLeadRequest) (*dto.LeadResponse, error)
	Convert(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.ConvertLeadRequest) (*dto.LeadConversionResponse, error)
	Nurture(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.NurtureLeadRequest) (*dto.LeadResponse, error)

	// Assignment
	Assign(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.AssignLeadRequest) (*dto.LeadResponse, error)
	BulkAssign(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkAssignLeadsRequest) (int, error)

	// Scoring
	UpdateScore(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.ScoreLeadRequest) (*dto.LeadResponse, error)

	// Statistics
	GetStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.LeadStatisticsResponse, error)
}

// ============================================================================
// Lead Use Case Implementation
// ============================================================================

// leadUseCase implements LeadUseCase.
type leadUseCase struct {
	leadRepo        domain.LeadRepository
	opportunityRepo domain.OpportunityRepository
	pipelineRepo    domain.PipelineRepository
	eventPublisher  ports.EventPublisher
	customerService ports.CustomerService
	userService     ports.UserService
	cacheService    ports.CacheService
	searchService   ports.SearchService
	idGenerator     ports.IDGenerator
}

// NewLeadUseCase creates a new lead use case.
func NewLeadUseCase(
	leadRepo domain.LeadRepository,
	opportunityRepo domain.OpportunityRepository,
	pipelineRepo domain.PipelineRepository,
	eventPublisher ports.EventPublisher,
	customerService ports.CustomerService,
	userService ports.UserService,
	cacheService ports.CacheService,
	searchService ports.SearchService,
	idGenerator ports.IDGenerator,
) LeadUseCase {
	return &leadUseCase{
		leadRepo:        leadRepo,
		opportunityRepo: opportunityRepo,
		pipelineRepo:    pipelineRepo,
		eventPublisher:  eventPublisher,
		customerService: customerService,
		userService:     userService,
		cacheService:    cacheService,
		searchService:   searchService,
		idGenerator:     idGenerator,
	}
}

// Create creates a new lead.
func (uc *leadUseCase) Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateLeadRequest) (*dto.LeadResponse, error) {
	// Build contact information
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

	// Build company information
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
		if req.Address.State != nil {
			company.State = *req.Address.State
		}
		company.Country = req.Address.Country
		if req.Address.PostalCode != nil {
			company.PostalCode = *req.Address.PostalCode
		}
	}

	// Parse source
	source := domain.LeadSource(req.Source)
	if !source.IsValid() {
		source = domain.LeadSourceOther
	}

	// Create lead
	lead, err := domain.NewLead(tenantID, contact, company, source, userID)
	if err != nil {
		return nil, application.ErrValidation(err.Error())
	}

	// Set optional fields
	if req.Description != nil {
		lead.Description = *req.Description
	}
	if req.Tags != nil {
		lead.Tags = req.Tags
	}
	if req.CustomFields != nil {
		lead.CustomFields = req.CustomFields
	}
	if req.Budget != nil && req.BudgetCurrency != nil {
		money, err := domain.NewMoney(*req.Budget, *req.BudgetCurrency)
		if err != nil {
			return nil, application.ErrValidation("invalid budget currency: " + err.Error())
		}
		lead.EstimatedValue = money
	}

	// Handle owner assignment
	if req.OwnerID != nil {
		ownerID, err := uuid.Parse(*req.OwnerID)
		if err == nil {
			// Get owner name from user service
			ownerName := ""
			if uc.userService != nil {
				user, err := uc.userService.GetUser(ctx, tenantID, ownerID)
				if err == nil && user != nil {
					ownerName = user.FullName
				}
			}
			lead.AssignOwner(ownerID, ownerName)
		}
	}

	// Generate code
	if uc.idGenerator != nil {
		code, err := uc.idGenerator.GenerateLeadNumber(ctx, tenantID)
		if err == nil {
			lead.Code = code
		}
	}

	// Save lead
	if err := uc.leadRepo.Create(ctx, lead); err != nil {
		return nil, application.ErrInternal("failed to create lead", err)
	}

	// Publish events
	uc.publishDomainEvents(ctx, lead.GetEvents())
	lead.ClearEvents()

	return uc.mapLeadToResponse(lead), nil
}

// GetByID retrieves a lead by ID.
func (uc *leadUseCase) GetByID(ctx context.Context, tenantID, leadID uuid.UUID) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	return uc.mapLeadToResponse(lead), nil
}

// Update updates a lead.
func (uc *leadUseCase) Update(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.UpdateLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	if lead.IsConverted() {
		return nil, application.ErrConflict("lead is already converted")
	}

	// Check version
	if lead.Version != req.Version {
		return nil, application.ErrVersionMismatch(req.Version, lead.Version)
	}

	// Update contact
	contact := lead.Contact
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

	// Update company
	company := lead.Company
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
		if req.Address.State != nil {
			company.State = *req.Address.State
		}
		company.Country = req.Address.Country
		if req.Address.PostalCode != nil {
			company.PostalCode = *req.Address.PostalCode
		}
	}

	// Update lead
	lead.Update(contact, company, lead.Source, lead.Rating)

	// Update other fields
	if req.Description != nil {
		lead.Description = *req.Description
	}
	if req.Tags != nil {
		lead.Tags = req.Tags
	}
	if req.CustomFields != nil {
		lead.CustomFields = req.CustomFields
	}
	if req.Budget != nil && req.BudgetCurrency != nil {
		money, err := domain.NewMoney(*req.Budget, *req.BudgetCurrency)
		if err != nil {
			return nil, application.ErrValidation("invalid budget currency: " + err.Error())
		}
		lead.SetEstimatedValue(money)
	}

	// Save
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.ErrInternal("failed to update lead", err)
	}

	// Publish events
	uc.publishDomainEvents(ctx, lead.GetEvents())
	lead.ClearEvents()

	return uc.mapLeadToResponse(lead), nil
}

// Delete deletes a lead.
func (uc *leadUseCase) Delete(ctx context.Context, tenantID, leadID, userID uuid.UUID) error {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return application.ErrLeadNotFound(leadID)
	}

	if err := lead.Delete(); err != nil {
		return application.ErrConflict(err.Error())
	}

	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return application.ErrInternal("failed to delete lead", err)
	}

	// Publish events
	uc.publishDomainEvents(ctx, lead.GetEvents())

	return nil
}

// List lists leads.
func (uc *leadUseCase) List(ctx context.Context, tenantID uuid.UUID, filter *dto.LeadFilterRequest) (*dto.LeadListResponse, error) {
	// Set defaults
	page := filter.Page
	pageSize := filter.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Build domain filter
	domainFilter := domain.LeadFilter{}
	if filter.Statuses != nil {
		for _, s := range filter.Statuses {
			domainFilter.Statuses = append(domainFilter.Statuses, domain.LeadStatus(s))
		}
	}
	if filter.Sources != nil {
		for _, s := range filter.Sources {
			domainFilter.Sources = append(domainFilter.Sources, domain.LeadSource(s))
		}
	}
	if filter.OwnerIDs != nil {
		for _, id := range filter.OwnerIDs {
			if ownerID, err := uuid.Parse(id); err == nil {
				domainFilter.OwnerIDs = append(domainFilter.OwnerIDs, ownerID)
			}
		}
	}
	if filter.MinScore != nil {
		domainFilter.MinScore = filter.MinScore
	}
	if filter.MaxScore != nil {
		domainFilter.MaxScore = filter.MaxScore
	}
	if filter.SearchQuery != "" {
		domainFilter.SearchQuery = filter.SearchQuery
	}

	// Build list options
	opts := domain.ListOptions{
		Page:     page,
		PageSize: pageSize,
		SortBy:   filter.SortBy,
		SortOrder: filter.SortOrder,
	}
	if opts.SortBy == "" {
		opts.SortBy = "created_at"
	}
	if opts.SortOrder == "" {
		opts.SortOrder = "desc"
	}

	// Get leads
	leads, total, err := uc.leadRepo.List(ctx, tenantID, domainFilter, opts)
	if err != nil {
		return nil, application.ErrInternal("failed to list leads", err)
	}

	// Map to response
	response := &dto.LeadListResponse{
		Leads:      make([]*dto.LeadBriefResponse, 0, len(leads)),
		Pagination: dto.NewPaginationResponse(page, pageSize, total),
	}

	for _, lead := range leads {
		response.Leads = append(response.Leads, uc.mapLeadToBriefResponse(lead))
	}

	return response, nil
}

// Qualify qualifies a lead.
func (uc *leadUseCase) Qualify(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.QualifyLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	if err := lead.Qualify(); err != nil {
		return nil, application.ErrConflict(err.Error())
	}

	// Update notes if provided
	if req.Notes != "" {
		lead.Notes = req.Notes
	}

	// Save
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.ErrInternal("failed to qualify lead", err)
	}

	// Publish events
	uc.publishDomainEvents(ctx, lead.GetEvents())
	lead.ClearEvents()

	return uc.mapLeadToResponse(lead), nil
}

// Disqualify disqualifies a lead.
func (uc *leadUseCase) Disqualify(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.DisqualifyLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	if err := lead.Disqualify(req.Reason, req.Notes, userID); err != nil {
		return nil, application.ErrConflict(err.Error())
	}

	// Save
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.ErrInternal("failed to disqualify lead", err)
	}

	// Publish events
	uc.publishDomainEvents(ctx, lead.GetEvents())
	lead.ClearEvents()

	return uc.mapLeadToResponse(lead), nil
}

// Convert converts a lead to an opportunity.
func (uc *leadUseCase) Convert(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.ConvertLeadRequest) (*dto.LeadConversionResponse, error) {
	// Get lead
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	if !lead.CanConvert() {
		return nil, application.ErrConflict("lead cannot be converted - must be qualified first")
	}

	// Get pipeline
	pipelineID, err := uuid.Parse(req.PipelineID)
	if err != nil {
		return nil, application.ErrValidation("invalid pipeline ID")
	}

	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Determine customer
	var customerID uuid.UUID
	customerName := lead.Company.Name
	if req.CustomerID != nil {
		customerID, err = uuid.Parse(*req.CustomerID)
		if err != nil {
			return nil, application.ErrValidation("invalid customer ID")
		}
		if req.CustomerName != nil {
			customerName = *req.CustomerName
		}
	} else if req.CreateNewCustomer && uc.customerService != nil {
		// Create new customer - use pointer types for optional fields
		email := lead.Contact.Email
		phone := lead.Contact.Phone
		website := lead.Company.Website

		createReq := ports.CreateCustomerRequest{
			Name:   lead.Company.Name,
			Type:   "business",
			Source: "lead_conversion",
		}
		if email != "" {
			createReq.Email = &email
		}
		if phone != "" {
			createReq.Phone = &phone
		}
		if website != "" {
			createReq.Website = &website
		}

		newCustomer, err := uc.customerService.CreateCustomer(ctx, tenantID, createReq)
		if err != nil {
			return nil, application.ErrInternal("failed to create customer", err)
		}
		customerID = newCustomer.ID
		customerName = newCustomer.Name
	}

	// Get owner
	ownerID := userID
	ownerName := ""
	if req.OwnerID != nil {
		ownerID, _ = uuid.Parse(*req.OwnerID)
	} else if lead.OwnerID != nil {
		ownerID = *lead.OwnerID
		ownerName = lead.OwnerName
	}

	// Create opportunity from lead
	opportunity, err := domain.NewOpportunityFromLead(lead, pipeline, customerID, customerName, userID)
	if err != nil {
		return nil, application.ErrInternal("failed to create opportunity from lead", err)
	}

	// Set owner
	opportunity.AssignOwner(ownerID, ownerName)

	// Set description if provided
	if req.Description != nil {
		opportunity.Description = *req.Description
	}

	// Save opportunity
	if err := uc.opportunityRepo.Create(ctx, opportunity); err != nil {
		return nil, application.ErrInternal("failed to create opportunity", err)
	}

	// Mark lead as converted
	var contactID *uuid.UUID
	if err := lead.ConvertToOpportunity(opportunity.ID, userID, &customerID, contactID); err != nil {
		return nil, application.ErrInternal("failed to convert lead", err)
	}

	// Save lead
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.ErrInternal("failed to update lead", err)
	}

	// Publish events
	uc.publishDomainEvents(ctx, lead.GetEvents())
	uc.publishDomainEvents(ctx, opportunity.GetEvents())

	return &dto.LeadConversionResponse{
		LeadID:        lead.ID.String(),
		OpportunityID: opportunity.ID.String(),
		CustomerID:    dto.StringPtr(customerID.String()),
		Message:       "Lead converted successfully",
	}, nil
}

// Nurture moves a lead to nurturing status.
func (uc *leadUseCase) Nurture(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.NurtureLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	if err := lead.StartNurturing(); err != nil {
		return nil, application.ErrConflict(err.Error())
	}

	// Set campaign if provided
	if req.NurtureCampaignID != nil {
		campaignID, err := uuid.Parse(*req.NurtureCampaignID)
		if err == nil {
			lead.CampaignID = &campaignID
		}
	}

	// Set reengage date if provided
	if req.ReengageDate != nil {
		t, err := time.Parse("2006-01-02", *req.ReengageDate)
		if err == nil {
			lead.SetNextFollowUp(t)
		}
	}

	// Save
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.ErrInternal("failed to nurture lead", err)
	}

	return uc.mapLeadToResponse(lead), nil
}

// Assign assigns a lead to an owner.
func (uc *leadUseCase) Assign(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.AssignLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, application.ErrValidation("invalid owner ID")
	}

	// Get owner name
	ownerName := ""
	if uc.userService != nil {
		user, err := uc.userService.GetUser(ctx, tenantID, ownerID)
		if err == nil && user != nil {
			ownerName = user.FullName
		}
	}

	lead.AssignOwner(ownerID, ownerName)

	// Save
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.ErrInternal("failed to assign lead", err)
	}

	// Publish events
	uc.publishDomainEvents(ctx, lead.GetEvents())
	lead.ClearEvents()

	return uc.mapLeadToResponse(lead), nil
}

// BulkAssign bulk assigns leads.
func (uc *leadUseCase) BulkAssign(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkAssignLeadsRequest) (int, error) {
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return 0, application.ErrValidation("invalid owner ID")
	}

	// Get owner name
	ownerName := ""
	if uc.userService != nil {
		user, err := uc.userService.GetUser(ctx, tenantID, ownerID)
		if err == nil && user != nil {
			ownerName = user.FullName
		}
	}

	updated := 0
	for _, leadIDStr := range req.LeadIDs {
		leadID, err := uuid.Parse(leadIDStr)
		if err != nil {
			continue
		}

		lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
		if err != nil {
			continue
		}

		lead.AssignOwner(ownerID, ownerName)

		if err := uc.leadRepo.Update(ctx, lead); err != nil {
			continue
		}

		updated++
	}

	return updated, nil
}

// UpdateScore updates lead scoring.
func (uc *leadUseCase) UpdateScore(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.ScoreLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	demographic := lead.Score.Demographic
	behavioral := lead.Score.Behavioral
	reason := req.Reason

	if req.DemographicScore != nil {
		demographic = *req.DemographicScore
	}
	if req.BehavioralScore != nil {
		behavioral = *req.BehavioralScore
	}
	if reason == "" {
		reason = "Manual score update"
	}

	lead.UpdateScore(demographic, behavioral, reason)

	// Save
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.ErrInternal("failed to update lead score", err)
	}

	// Publish events
	uc.publishDomainEvents(ctx, lead.GetEvents())
	lead.ClearEvents()

	return uc.mapLeadToResponse(lead), nil
}

// GetStatistics returns lead statistics.
func (uc *leadUseCase) GetStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.LeadStatisticsResponse, error) {
	// Get counts by status
	statusCounts, err := uc.leadRepo.CountByStatus(ctx, tenantID)
	if err != nil {
		return nil, application.ErrInternal("failed to get lead statistics", err)
	}

	// Get counts by source
	sourceCounts, err := uc.leadRepo.CountBySource(ctx, tenantID)
	if err != nil {
		return nil, application.ErrInternal("failed to get lead statistics", err)
	}

	// Get conversion rate for last 30 days
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	conversionRate, err := uc.leadRepo.GetConversionRate(ctx, tenantID, start, end)
	if err != nil {
		conversionRate = 0 // Default to 0 if error
	}

	// Calculate totals
	var totalLeads int64
	byStatus := make(map[string]int64)
	for status, count := range statusCounts {
		totalLeads += count
		byStatus[string(status)] = count
	}

	bySource := make(map[string]int64)
	for source, count := range sourceCounts {
		bySource[string(source)] = count
	}

	return &dto.LeadStatisticsResponse{
		TotalLeads:     totalLeads,
		ByStatus:       byStatus,
		BySource:       bySource,
		ByRating:       make(map[string]int64), // Would need additional repo method
		ConversionRate: conversionRate,
		AverageScore:   0, // Would need additional repo method
		NewLeadsToday:  0, // Would need additional repo method
		NewLeadsWeek:   0, // Would need additional repo method
		NewLeadsMonth:  0, // Would need additional repo method
	}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// publishDomainEvents converts domain events to ports.Event and publishes them.
func (uc *leadUseCase) publishDomainEvents(ctx context.Context, events []domain.DomainEvent) {
	if uc.eventPublisher == nil {
		return
	}

	for _, domainEvent := range events {
		event := ports.Event{
			ID:            domainEvent.EventID().String(),
			Type:          domainEvent.EventType(),
			AggregateID:   domainEvent.AggregateID().String(),
			AggregateType: domainEvent.AggregateType(),
			TenantID:      domainEvent.TenantID().String(),
			Payload:       make(map[string]interface{}),
			Metadata:      make(map[string]string),
			OccurredAt:    domainEvent.OccurredAt(),
			Version:       domainEvent.Version(),
		}
		_ = uc.eventPublisher.Publish(ctx, event)
	}
}

// ============================================================================
// Mapping Functions
// ============================================================================

func (uc *leadUseCase) mapLeadToResponse(lead *domain.Lead) *dto.LeadResponse {
	resp := &dto.LeadResponse{
		ID:               lead.ID.String(),
		TenantID:         lead.TenantID.String(),
		FirstName:        lead.Contact.FirstName,
		LastName:         lead.Contact.LastName,
		FullName:         lead.Contact.FullName(),
		Email:            lead.Contact.Email,
		Phone:            dto.StringPtr(lead.Contact.Phone),
		Mobile:           dto.StringPtr(lead.Contact.Mobile),
		JobTitle:         dto.StringPtr(lead.Contact.JobTitle),
		Department:       dto.StringPtr(lead.Contact.Department),
		Company:          dto.StringPtr(lead.Company.Name),
		CompanySize:      dto.StringPtr(lead.Company.Size),
		Industry:         dto.StringPtr(lead.Company.Industry),
		Website:          dto.StringPtr(lead.Company.Website),
		Status:           string(lead.Status),
		Score:            lead.Score.Score,
		DemographicScore: lead.Score.Demographic,
		BehavioralScore:  lead.Score.Behavioral,
		Rating:           string(lead.Rating),
		Source:           string(lead.Source),
		Description:      dto.StringPtr(lead.Description),
		Tags:             lead.Tags,
		CustomFields:     lead.CustomFields,
		CreatedAt:        lead.CreatedAt,
		UpdatedAt:        lead.UpdatedAt,
		CreatedBy:        lead.CreatedBy.String(),
		UpdatedBy:        lead.CreatedBy.String(),
		Version:          lead.Version,
	}

	if lead.OwnerID != nil {
		resp.OwnerID = dto.StringPtr(lead.OwnerID.String())
	}

	if lead.EstimatedValue.Amount > 0 {
		resp.Budget = &dto.MoneyDTO{
			Amount:   lead.EstimatedValue.Amount,
			Currency: lead.EstimatedValue.Currency,
			Display:  fmt.Sprintf("%s %d", lead.EstimatedValue.Currency, lead.EstimatedValue.Amount/100),
		}
	}

	if lead.Company.Address != "" {
		resp.Address = &dto.AddressDTO{
			Street1: lead.Company.Address,
			City:    lead.Company.City,
			State:   dto.StringPtr(lead.Company.State),
			Country: lead.Company.Country,
		}
	}

	// Conversion info
	if lead.ConversionInfo != nil {
		resp.ConvertedAt = &lead.ConversionInfo.ConvertedAt
		resp.ConvertedBy = dto.StringPtr(lead.ConversionInfo.ConvertedBy.String())
		resp.OpportunityID = dto.StringPtr(lead.ConversionInfo.OpportunityID.String())
		if lead.ConversionInfo.CustomerID != nil {
			resp.CustomerID = dto.StringPtr(lead.ConversionInfo.CustomerID.String())
		}
		if lead.ConversionInfo.ContactID != nil {
			resp.ContactID = dto.StringPtr(lead.ConversionInfo.ContactID.String())
		}
	}

	// Disqualify info
	if lead.DisqualifyInfo != nil {
		resp.DisqualifiedAt = &lead.DisqualifyInfo.DisqualifiedAt
		resp.DisqualifiedBy = dto.StringPtr(lead.DisqualifyInfo.DisqualifiedBy.String())
		resp.DisqualifyReason = dto.StringPtr(lead.DisqualifyInfo.Reason)
	}

	// Activity info
	resp.LastContactedAt = lead.LastContactedAt
	if lead.Engagement.LastEngagement != nil {
		resp.LastActivityAt = lead.Engagement.LastEngagement
	}
	resp.ActivityCount = lead.Engagement.EmailsOpened + lead.Engagement.EmailsClicked + lead.Engagement.WebVisits + lead.Engagement.FormSubmissions

	return resp
}

func (uc *leadUseCase) mapLeadToBriefResponse(lead *domain.Lead) *dto.LeadBriefResponse {
	resp := &dto.LeadBriefResponse{
		ID:        lead.ID.String(),
		FirstName: lead.Contact.FirstName,
		LastName:  lead.Contact.LastName,
		FullName:  lead.Contact.FullName(),
		Email:     lead.Contact.Email,
		Company:   dto.StringPtr(lead.Company.Name),
		Status:    string(lead.Status),
		Score:     lead.Score.Score,
		Rating:    string(lead.Rating),
		CreatedAt: lead.CreatedAt,
	}

	if lead.OwnerID != nil {
		resp.OwnerID = dto.StringPtr(lead.OwnerID.String())
	}

	return resp
}
