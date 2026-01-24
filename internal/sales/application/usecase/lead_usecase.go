package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"crm-kilang-desa-murni-batik/internal/sales/application"
	"crm-kilang-desa-murni-batik/internal/sales/application/dto"
	"crm-kilang-desa-murni-batik/internal/sales/application/ports"
	"crm-kilang-desa-murni-batik/internal/sales/domain"
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
	Contact(ctx context.Context, tenantID, leadID, userID uuid.UUID) (*dto.LeadResponse, error)
	Nurture(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.NurtureLeadRequest) (*dto.LeadResponse, error)

	// Assignment
	Assign(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.AssignLeadRequest) (*dto.LeadResponse, error)
	BulkAssign(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkAssignLeadsRequest) error

	// Scoring
	UpdateScore(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.ScoreLeadRequest) (*dto.LeadResponse, error)

	// Conversion
	ConvertToOpportunity(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.ConvertLeadRequest) (*dto.LeadConversionResponse, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkUpdateLeadStatusRequest) error

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

// ============================================================================
// CRUD Operations
// ============================================================================

// Create creates a new lead.
func (uc *leadUseCase) Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateLeadRequest) (*dto.LeadResponse, error) {
	// Check for duplicate email
	existingLead, err := uc.leadRepo.GetByEmail(ctx, tenantID, req.Email)
	if err == nil && existingLead != nil {
		return nil, application.ErrLeadDuplicateEmail(req.Email)
	}

	// Parse owner ID if provided
	var ownerID *uuid.UUID
	if req.OwnerID != nil {
		parsedOwnerID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return nil, application.ErrValidation("invalid owner_id format")
		}
		// Verify owner exists
		exists, err := uc.userService.UserExists(ctx, tenantID, parsedOwnerID)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeUserServiceError, "failed to verify owner", err)
		}
		if !exists {
			return nil, application.ErrUserNotFound(parsedOwnerID)
		}
		ownerID = &parsedOwnerID
	}

	// Parse campaign ID if provided
	var campaignID *uuid.UUID
	if req.CampaignID != nil {
		parsedCampaignID, err := uuid.Parse(*req.CampaignID)
		if err != nil {
			return nil, application.ErrValidation("invalid campaign_id format")
		}
		campaignID = &parsedCampaignID
	}

	// Map source
	source := mapLeadSource(req.Source)

	// Create lead entity
	lead, err := domain.NewLead(
		tenantID,
		req.FirstName,
		req.LastName,
		req.Email,
		source,
		userID,
	)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeValidation, "failed to create lead", err)
	}

	// Set optional fields
	if req.Phone != nil {
		lead.Phone = req.Phone
	}
	if req.Mobile != nil {
		lead.Mobile = req.Mobile
	}
	if req.JobTitle != nil {
		lead.JobTitle = req.JobTitle
	}
	if req.Department != nil {
		lead.Department = req.Department
	}
	if req.Company != nil {
		lead.Company = req.Company
	}
	if req.CompanySize != nil {
		lead.CompanySize = req.CompanySize
	}
	if req.Industry != nil {
		lead.Industry = req.Industry
	}
	if req.Website != nil {
		lead.Website = req.Website
	}
	if req.AnnualRevenue != nil {
		lead.AnnualRevenue = req.AnnualRevenue
	}
	if req.NumberEmployees != nil {
		lead.NumberEmployees = req.NumberEmployees
	}
	if req.Address != nil {
		lead.Address = mapAddressDTOToDomain(req.Address)
	}
	if req.SourceDetails != nil {
		lead.SourceDetails = req.SourceDetails
	}
	if campaignID != nil {
		lead.CampaignID = campaignID
	}
	if req.ReferralSource != nil {
		lead.ReferralSource = req.ReferralSource
	}

	// Set UTM parameters
	if req.UTMSource != nil {
		lead.UTMSource = req.UTMSource
	}
	if req.UTMMedium != nil {
		lead.UTMMedium = req.UTMMedium
	}
	if req.UTMCampaign != nil {
		lead.UTMCampaign = req.UTMCampaign
	}
	if req.UTMTerm != nil {
		lead.UTMTerm = req.UTMTerm
	}
	if req.UTMContent != nil {
		lead.UTMContent = req.UTMContent
	}

	// Set owner
	if ownerID != nil {
		lead.OwnerID = ownerID
	}

	// Set additional fields
	if req.Description != nil {
		lead.Description = req.Description
	}
	if req.Tags != nil {
		lead.Tags = req.Tags
	}
	if req.CustomFields != nil {
		lead.CustomFields = req.CustomFields
	}
	if req.ProductInterest != nil {
		lead.ProductInterest = req.ProductInterest
	}
	if req.Budget != nil && req.BudgetCurrency != nil {
		budget, _ := domain.NewMoney(*req.Budget, *req.BudgetCurrency)
		lead.Budget = budget
	}
	if req.Timeline != nil {
		lead.Timeline = req.Timeline
	}
	if req.Requirements != nil {
		lead.Requirements = req.Requirements
	}
	if req.MarketingConsent != nil {
		lead.MarketingConsent = *req.MarketingConsent
	}
	if req.PrivacyConsent != nil {
		lead.PrivacyConsent = *req.PrivacyConsent
	}

	// Save lead
	if err := uc.leadRepo.Create(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to save lead", err)
	}

	// Publish domain events
	for _, event := range lead.Events() {
		if err := uc.publishEvent(ctx, event); err != nil {
			// Log error but don't fail the operation
		}
	}
	lead.ClearEvents()

	// Index for search
	if uc.searchService != nil {
		go uc.indexLead(context.Background(), lead)
	}

	// Invalidate cache
	uc.invalidateLeadCache(ctx, tenantID)

	return uc.mapLeadToResponse(ctx, lead), nil
}

// GetByID retrieves a lead by ID.
func (uc *leadUseCase) GetByID(ctx context.Context, tenantID, leadID uuid.UUID) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	return uc.mapLeadToResponse(ctx, lead), nil
}

// Update updates a lead.
func (uc *leadUseCase) Update(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.UpdateLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	// Check version for optimistic locking
	if lead.Version != req.Version {
		return nil, application.ErrVersionMismatch(req.Version, lead.Version)
	}

	// Check if lead can be updated (not converted)
	if lead.Status == domain.LeadStatusConverted {
		return nil, application.ErrLeadAlreadyConverted(leadID)
	}

	// Update fields
	if req.FirstName != nil {
		lead.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		lead.LastName = *req.LastName
	}
	if req.Email != nil && *req.Email != lead.Email {
		// Check for duplicate email
		existingLead, err := uc.leadRepo.GetByEmail(ctx, tenantID, *req.Email)
		if err == nil && existingLead != nil && existingLead.ID != leadID {
			return nil, application.ErrLeadDuplicateEmail(*req.Email)
		}
		lead.Email = *req.Email
	}
	if req.Phone != nil {
		lead.Phone = req.Phone
	}
	if req.Mobile != nil {
		lead.Mobile = req.Mobile
	}
	if req.JobTitle != nil {
		lead.JobTitle = req.JobTitle
	}
	if req.Department != nil {
		lead.Department = req.Department
	}
	if req.Company != nil {
		lead.Company = req.Company
	}
	if req.CompanySize != nil {
		lead.CompanySize = req.CompanySize
	}
	if req.Industry != nil {
		lead.Industry = req.Industry
	}
	if req.Website != nil {
		lead.Website = req.Website
	}
	if req.AnnualRevenue != nil {
		lead.AnnualRevenue = req.AnnualRevenue
	}
	if req.NumberEmployees != nil {
		lead.NumberEmployees = req.NumberEmployees
	}
	if req.Address != nil {
		lead.Address = mapAddressDTOToDomain(req.Address)
	}
	if req.SourceDetails != nil {
		lead.SourceDetails = req.SourceDetails
	}
	if req.ReferralSource != nil {
		lead.ReferralSource = req.ReferralSource
	}
	if req.Description != nil {
		lead.Description = req.Description
	}
	if req.Tags != nil {
		lead.Tags = req.Tags
	}
	if req.CustomFields != nil {
		lead.CustomFields = req.CustomFields
	}
	if req.ProductInterest != nil {
		lead.ProductInterest = req.ProductInterest
	}
	if req.Budget != nil && req.BudgetCurrency != nil {
		budget, _ := domain.NewMoney(*req.Budget, *req.BudgetCurrency)
		lead.Budget = budget
	}
	if req.Timeline != nil {
		lead.Timeline = req.Timeline
	}
	if req.Requirements != nil {
		lead.Requirements = req.Requirements
	}
	if req.MarketingConsent != nil {
		lead.MarketingConsent = *req.MarketingConsent
	}
	if req.PrivacyConsent != nil {
		lead.PrivacyConsent = *req.PrivacyConsent
	}

	// Update metadata
	lead.UpdatedAt = time.Now()
	lead.UpdatedBy = userID
	lead.Version++

	// Add update event
	lead.AddEvent(domain.NewLeadUpdatedEvent(lead, userID))

	// Save changes
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update lead", err)
	}

	// Publish events
	for _, event := range lead.Events() {
		if err := uc.publishEvent(ctx, event); err != nil {
			// Log error but don't fail
		}
	}
	lead.ClearEvents()

	// Update search index
	if uc.searchService != nil {
		go uc.indexLead(context.Background(), lead)
	}

	// Invalidate cache
	uc.invalidateLeadCache(ctx, tenantID)

	return uc.mapLeadToResponse(ctx, lead), nil
}

// Delete deletes a lead.
func (uc *leadUseCase) Delete(ctx context.Context, tenantID, leadID, userID uuid.UUID) error {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return application.ErrLeadNotFound(leadID)
	}

	// Check if lead can be deleted (not converted)
	if lead.Status == domain.LeadStatusConverted {
		return application.ErrLeadAlreadyConverted(leadID)
	}

	// Delete lead
	if err := uc.leadRepo.Delete(ctx, tenantID, leadID); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to delete lead", err)
	}

	// Publish delete event
	event := domain.NewLeadDeletedEvent(lead, userID)
	uc.publishEvent(ctx, event)

	// Remove from search index
	if uc.searchService != nil {
		go uc.searchService.DeleteIndex(context.Background(), tenantID, "lead", leadID)
	}

	// Invalidate cache
	uc.invalidateLeadCache(ctx, tenantID)

	return nil
}

// List lists leads with filtering.
func (uc *leadUseCase) List(ctx context.Context, tenantID uuid.UUID, filter *dto.LeadFilterRequest) (*dto.LeadListResponse, error) {
	// Map filter to domain filter
	domainFilter := uc.mapFilterToDomain(filter)

	// Set pagination defaults
	opts := domain.DefaultListOptions()
	if filter.Page > 0 {
		opts.Page = filter.Page
	}
	if filter.PageSize > 0 {
		opts.PageSize = filter.PageSize
	}
	if filter.SortBy != "" {
		opts.SortBy = filter.SortBy
	}
	if filter.SortOrder != "" {
		opts.SortOrder = filter.SortOrder
	}

	// Get leads
	leads, total, err := uc.leadRepo.List(ctx, tenantID, domainFilter, opts)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to list leads", err)
	}

	// Map to response
	leadResponses := make([]*dto.LeadBriefResponse, len(leads))
	for i, lead := range leads {
		leadResponses[i] = uc.mapLeadToBriefResponse(lead)
	}

	return &dto.LeadListResponse{
		Leads:      leadResponses,
		Pagination: dto.NewPaginationResponse(opts.Page, opts.PageSize, total),
	}, nil
}

// ============================================================================
// Status Operations
// ============================================================================

// Qualify qualifies a lead.
func (uc *leadUseCase) Qualify(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.QualifyLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	// Qualify the lead
	if err := lead.Qualify(userID); err != nil {
		return nil, application.WrapError(application.ErrCodeLeadInvalidTransition, err.Error(), err)
	}

	// Update qualification details
	if req.Budget != nil && req.BudgetCurrency != nil {
		budget, _ := domain.NewMoney(*req.Budget, *req.BudgetCurrency)
		lead.Budget = budget
	}
	if req.Authority != nil {
		lead.Authority = req.Authority
	}
	if req.Need != nil {
		lead.Need = req.Need
	}
	if req.Timeline != nil {
		lead.Timeline = req.Timeline
	}
	if req.Notes != "" {
		lead.QualificationNotes = &req.Notes
	}
	if req.QualificationCriteria != nil {
		if lead.CustomFields == nil {
			lead.CustomFields = make(map[string]interface{})
		}
		lead.CustomFields["qualification_criteria"] = req.QualificationCriteria
	}

	// Save changes
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to qualify lead", err)
	}

	// Publish events
	for _, event := range lead.Events() {
		uc.publishEvent(ctx, event)
	}
	lead.ClearEvents()

	// Update search index
	if uc.searchService != nil {
		go uc.indexLead(context.Background(), lead)
	}

	return uc.mapLeadToResponse(ctx, lead), nil
}

// Disqualify disqualifies a lead.
func (uc *leadUseCase) Disqualify(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.DisqualifyLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	// Disqualify the lead
	if err := lead.Disqualify(req.Reason, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeLeadInvalidTransition, err.Error(), err)
	}

	// Set notes
	if req.Notes != "" {
		lead.DisqualifyNotes = &req.Notes
	}

	// Save changes
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to disqualify lead", err)
	}

	// Publish events
	for _, event := range lead.Events() {
		uc.publishEvent(ctx, event)
	}
	lead.ClearEvents()

	// Update search index
	if uc.searchService != nil {
		go uc.indexLead(context.Background(), lead)
	}

	return uc.mapLeadToResponse(ctx, lead), nil
}

// Contact marks a lead as contacted.
func (uc *leadUseCase) Contact(ctx context.Context, tenantID, leadID, userID uuid.UUID) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	// Mark as contacted
	if err := lead.Contact(userID); err != nil {
		return nil, application.WrapError(application.ErrCodeLeadInvalidTransition, err.Error(), err)
	}

	// Save changes
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update lead", err)
	}

	// Publish events
	for _, event := range lead.Events() {
		uc.publishEvent(ctx, event)
	}
	lead.ClearEvents()

	return uc.mapLeadToResponse(ctx, lead), nil
}

// Nurture moves a lead to nurturing status.
func (uc *leadUseCase) Nurture(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.NurtureLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	// Move to nurturing
	if err := lead.Nurture(userID); err != nil {
		return nil, application.WrapError(application.ErrCodeLeadInvalidTransition, err.Error(), err)
	}

	// Set nurturing details
	if req.NurtureCampaignID != nil {
		campaignID, _ := uuid.Parse(*req.NurtureCampaignID)
		lead.NurtureCampaignID = &campaignID
	}
	if req.Reason != "" {
		lead.NurtureReason = &req.Reason
	}
	if req.ReengageDate != nil {
		reengageDate, _ := time.Parse("2006-01-02", *req.ReengageDate)
		lead.ReengageDate = &reengageDate
	}

	// Save changes
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update lead", err)
	}

	// Publish events
	for _, event := range lead.Events() {
		uc.publishEvent(ctx, event)
	}
	lead.ClearEvents()

	return uc.mapLeadToResponse(ctx, lead), nil
}

// ============================================================================
// Assignment
// ============================================================================

// Assign assigns a lead to an owner.
func (uc *leadUseCase) Assign(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.AssignLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	// Parse owner ID
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, application.ErrValidation("invalid owner_id format")
	}

	// Verify owner exists
	exists, err := uc.userService.UserExists(ctx, tenantID, ownerID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeUserServiceError, "failed to verify owner", err)
	}
	if !exists {
		return nil, application.ErrUserNotFound(ownerID)
	}

	// Assign lead
	previousOwner := lead.OwnerID
	if err := lead.Assign(ownerID, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeLeadAssignmentFailed, err.Error(), err)
	}

	// Add assignment notes
	if req.Notes != "" {
		if lead.CustomFields == nil {
			lead.CustomFields = make(map[string]interface{})
		}
		lead.CustomFields["assignment_notes"] = req.Notes
	}

	// Save changes
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to assign lead", err)
	}

	// Publish events
	for _, event := range lead.Events() {
		uc.publishEvent(ctx, event)
	}
	lead.ClearEvents()

	// Notify new owner (if different from previous)
	if previousOwner == nil || *previousOwner != ownerID {
		// Send notification asynchronously
	}

	return uc.mapLeadToResponse(ctx, lead), nil
}

// BulkAssign bulk assigns leads to an owner.
func (uc *leadUseCase) BulkAssign(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkAssignLeadsRequest) error {
	// Parse owner ID
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return application.ErrValidation("invalid owner_id format")
	}

	// Verify owner exists
	exists, err := uc.userService.UserExists(ctx, tenantID, ownerID)
	if err != nil {
		return application.WrapError(application.ErrCodeUserServiceError, "failed to verify owner", err)
	}
	if !exists {
		return application.ErrUserNotFound(ownerID)
	}

	// Parse lead IDs
	leadIDs := make([]uuid.UUID, len(req.LeadIDs))
	for i, id := range req.LeadIDs {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return application.ErrValidation("invalid lead_id format")
		}
		leadIDs[i] = parsedID
	}

	// Bulk update
	if err := uc.leadRepo.BulkUpdateOwner(ctx, tenantID, leadIDs, ownerID); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to bulk assign leads", err)
	}

	// Invalidate cache
	uc.invalidateLeadCache(ctx, tenantID)

	return nil
}

// ============================================================================
// Scoring
// ============================================================================

// UpdateScore updates a lead's score.
func (uc *leadUseCase) UpdateScore(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.ScoreLeadRequest) (*dto.LeadResponse, error) {
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	// Update scores
	if req.DemographicScore != nil {
		lead.DemographicScore = *req.DemographicScore
	}
	if req.BehavioralScore != nil {
		lead.BehavioralScore = *req.BehavioralScore
	}

	// Recalculate total score
	lead.CalculateScore()

	// Process activity if provided
	if req.Activity != nil {
		lead.BehavioralScore += req.Activity.Score
		lead.CalculateScore()
		lead.LastActivityAt = timePtr(time.Now())
		lead.ActivityCount++
	}

	// Update metadata
	lead.UpdatedAt = time.Now()
	lead.UpdatedBy = userID
	lead.Version++

	// Add score updated event
	lead.AddEvent(domain.NewLeadScoreUpdatedEvent(lead, lead.Score, userID))

	// Save changes
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update lead score", err)
	}

	// Publish events
	for _, event := range lead.Events() {
		uc.publishEvent(ctx, event)
	}
	lead.ClearEvents()

	return uc.mapLeadToResponse(ctx, lead), nil
}

// ============================================================================
// Conversion
// ============================================================================

// ConvertToOpportunity converts a lead to an opportunity.
func (uc *leadUseCase) ConvertToOpportunity(ctx context.Context, tenantID, leadID, userID uuid.UUID, req *dto.ConvertLeadRequest) (*dto.LeadConversionResponse, error) {
	// Get the lead
	lead, err := uc.leadRepo.GetByID(ctx, tenantID, leadID)
	if err != nil {
		return nil, application.ErrLeadNotFound(leadID)
	}

	// Check if lead can be converted
	if lead.Status == domain.LeadStatusConverted {
		return nil, application.ErrLeadAlreadyConverted(leadID)
	}
	if lead.Status != domain.LeadStatusQualified {
		return nil, application.ErrLeadNotQualified(leadID)
	}

	// Parse pipeline ID
	pipelineID, err := uuid.Parse(req.PipelineID)
	if err != nil {
		return nil, application.ErrValidation("invalid pipeline_id format")
	}

	// Get pipeline
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}
	if !pipeline.IsActive {
		return nil, application.ErrPipelineInactive(pipelineID)
	}

	// Determine starting stage
	var stage *domain.Stage
	if req.StageID != nil {
		stageID, _ := uuid.Parse(*req.StageID)
		stage = pipeline.GetStageByID(stageID)
		if stage == nil {
			return nil, application.ErrPipelineStageNotFound(pipelineID, stageID)
		}
	} else {
		// Get first open stage
		stage = pipeline.GetFirstOpenStage()
		if stage == nil {
			return nil, application.NewAppError(application.ErrCodePipelineStageNotFound, "pipeline has no open stages")
		}
	}

	// Handle customer
	var customerID *uuid.UUID
	if req.CustomerID != nil {
		parsedCustomerID, _ := uuid.Parse(*req.CustomerID)
		// Verify customer exists
		exists, err := uc.customerService.CustomerExists(ctx, tenantID, parsedCustomerID)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to verify customer", err)
		}
		if !exists {
			return nil, application.ErrCustomerNotFound(parsedCustomerID)
		}
		customerID = &parsedCustomerID
	} else if req.CreateNewCustomer && req.CustomerName != nil {
		// Create new customer
		customerReq := ports.CreateCustomerRequest{
			Name:    *req.CustomerName,
			Type:    "company",
			Email:   &lead.Email,
			Phone:   lead.Phone,
			Website: lead.Website,
			Source:  string(lead.Source),
		}
		if lead.Address != nil {
			customerReq.Address = &ports.AddressInfo{
				Street1: lead.Address.Street1,
				City:    lead.Address.City,
				Country: lead.Address.Country,
			}
		}

		customer, err := uc.customerService.CreateCustomer(ctx, tenantID, customerReq)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to create customer", err)
		}
		customerID = &customer.ID
	}

	// Handle contact
	var contactID *uuid.UUID
	if req.ContactID != nil {
		parsedContactID, _ := uuid.Parse(*req.ContactID)
		exists, err := uc.customerService.ContactExists(ctx, tenantID, parsedContactID)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to verify contact", err)
		}
		if !exists {
			return nil, application.ErrContactNotFound(parsedContactID)
		}
		contactID = &parsedContactID
	} else if req.CreateNewContact && customerID != nil {
		// Create new contact from lead data
		contactReq := ports.CreateContactRequest{
			FirstName:  lead.FirstName,
			LastName:   lead.LastName,
			Email:      lead.Email,
			Phone:      lead.Phone,
			Mobile:     lead.Mobile,
			JobTitle:   lead.JobTitle,
			Department: lead.Department,
			IsPrimary:  true,
		}

		contact, err := uc.customerService.CreateContact(ctx, tenantID, *customerID, contactReq)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to create contact", err)
		}
		contactID = &contact.ID
	}

	// Create opportunity ID
	opportunityID := uc.idGenerator.GenerateID()

	// Convert lead
	if err := lead.ConvertToOpportunity(opportunityID, userID, customerID, contactID); err != nil {
		return nil, application.WrapError(application.ErrCodeLeadConversionFailed, err.Error(), err)
	}

	// Create opportunity
	var amount int64 = 0
	currency := pipeline.DefaultCurrency
	if req.ExpectedAmount != nil {
		amount = *req.ExpectedAmount
	} else if lead.Budget != nil {
		amount = lead.Budget.Amount
		currency = lead.Budget.Currency
	}
	if req.Currency != nil {
		currency = *req.Currency
	}

	opportunityAmount, _ := domain.NewMoney(amount, currency)

	expectedCloseDate := time.Now().AddDate(0, 3, 0) // Default 3 months
	if req.ExpectedCloseDate != nil {
		expectedCloseDate, _ = time.Parse("2006-01-02", *req.ExpectedCloseDate)
	}

	probability := stage.Probability
	if req.Probability != nil {
		probability = *req.Probability
	}

	// Determine owner
	ownerID := userID
	if req.OwnerID != nil {
		ownerID, _ = uuid.Parse(*req.OwnerID)
	} else if lead.OwnerID != nil {
		ownerID = *lead.OwnerID
	}

	opportunity, err := domain.NewOpportunity(
		opportunityID,
		tenantID,
		req.OpportunityName,
		pipelineID,
		stage.ID,
		opportunityAmount,
		expectedCloseDate,
		ownerID,
		userID,
	)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeLeadConversionFailed, "failed to create opportunity", err)
	}

	// Set additional fields
	opportunity.LeadID = &leadID
	opportunity.CustomerID = customerID
	opportunity.PrimaryContactID = contactID
	opportunity.Probability = probability
	if req.Description != nil {
		opportunity.Description = req.Description
	} else if lead.Description != nil {
		opportunity.Description = lead.Description
	}
	if req.Source != nil {
		opportunity.Source = *req.Source
	} else {
		opportunity.Source = string(lead.Source)
	}

	// Save opportunity
	if err := uc.opportunityRepo.Create(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to save opportunity", err)
	}

	// Update lead
	if err := uc.leadRepo.Update(ctx, lead); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update lead", err)
	}

	// Publish events
	for _, event := range lead.Events() {
		uc.publishEvent(ctx, event)
	}
	lead.ClearEvents()

	for _, event := range opportunity.Events() {
		uc.publishEvent(ctx, event)
	}
	opportunity.ClearEvents()

	// Invalidate caches
	uc.invalidateLeadCache(ctx, tenantID)

	var customerIDStr, contactIDStr *string
	if customerID != nil {
		s := customerID.String()
		customerIDStr = &s
	}
	if contactID != nil {
		s := contactID.String()
		contactIDStr = &s
	}

	return &dto.LeadConversionResponse{
		LeadID:        leadID.String(),
		OpportunityID: opportunityID.String(),
		CustomerID:    customerIDStr,
		ContactID:     contactIDStr,
		Message:       "Lead successfully converted to opportunity",
	}, nil
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkUpdateStatus bulk updates lead statuses.
func (uc *leadUseCase) BulkUpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkUpdateLeadStatusRequest) error {
	// Parse lead IDs
	leadIDs := make([]uuid.UUID, len(req.LeadIDs))
	for i, id := range req.LeadIDs {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return application.ErrValidation("invalid lead_id format")
		}
		leadIDs[i] = parsedID
	}

	// Map status
	status := mapLeadStatus(req.Status)

	// Bulk update
	if err := uc.leadRepo.BulkUpdateStatus(ctx, tenantID, leadIDs, status); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to bulk update lead status", err)
	}

	// Invalidate cache
	uc.invalidateLeadCache(ctx, tenantID)

	return nil
}

// ============================================================================
// Statistics
// ============================================================================

// GetStatistics retrieves lead statistics.
func (uc *leadUseCase) GetStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.LeadStatisticsResponse, error) {
	// Get counts by status
	byStatus, err := uc.leadRepo.CountByStatus(ctx, tenantID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to get status counts", err)
	}

	// Get counts by source
	bySource, err := uc.leadRepo.CountBySource(ctx, tenantID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to get source counts", err)
	}

	// Get conversion rate
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	conversionRate, err := uc.leadRepo.GetConversionRate(ctx, tenantID, startOfMonth, now)
	if err != nil {
		conversionRate = 0
	}

	// Calculate totals
	var total int64
	for _, count := range byStatus {
		total += count
	}

	// Get new leads counts
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := today.AddDate(0, 0, -int(today.Weekday()))

	newToday, _, _ := uc.leadRepo.GetCreatedBetween(ctx, tenantID, today, now, domain.ListOptions{PageSize: 1})
	newWeek, _, _ := uc.leadRepo.GetCreatedBetween(ctx, tenantID, startOfWeek, now, domain.ListOptions{PageSize: 1})
	newMonth, _, _ := uc.leadRepo.GetCreatedBetween(ctx, tenantID, startOfMonth, now, domain.ListOptions{PageSize: 1})

	// Map status to string
	statusMap := make(map[string]int64)
	for status, count := range byStatus {
		statusMap[string(status)] = count
	}

	// Map source to string
	sourceMap := make(map[string]int64)
	for source, count := range bySource {
		sourceMap[string(source)] = count
	}

	// Calculate by rating (based on score thresholds)
	byRating := map[string]int64{
		"hot":  0,
		"warm": 0,
		"cold": 0,
	}

	return &dto.LeadStatisticsResponse{
		TotalLeads:     total,
		ByStatus:       statusMap,
		BySource:       sourceMap,
		ByRating:       byRating,
		ConversionRate: conversionRate,
		AverageScore:   0, // Calculate from actual data
		NewLeadsToday:  int64(len(newToday)),
		NewLeadsWeek:   int64(len(newWeek)),
		NewLeadsMonth:  int64(len(newMonth)),
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (uc *leadUseCase) mapLeadToResponse(ctx context.Context, lead *domain.Lead) *dto.LeadResponse {
	resp := &dto.LeadResponse{
		ID:               lead.ID.String(),
		TenantID:         lead.TenantID.String(),
		FirstName:        lead.FirstName,
		LastName:         lead.LastName,
		FullName:         lead.FirstName + " " + lead.LastName,
		Email:            lead.Email,
		Phone:            lead.Phone,
		Mobile:           lead.Mobile,
		JobTitle:         lead.JobTitle,
		Department:       lead.Department,
		Company:          lead.Company,
		CompanySize:      lead.CompanySize,
		Industry:         lead.Industry,
		Website:          lead.Website,
		AnnualRevenue:    lead.AnnualRevenue,
		NumberEmployees:  lead.NumberEmployees,
		Status:           string(lead.Status),
		Score:            lead.Score,
		DemographicScore: lead.DemographicScore,
		BehavioralScore:  lead.BehavioralScore,
		Rating:           lead.GetRating(),
		Source:           string(lead.Source),
		SourceDetails:    lead.SourceDetails,
		ReferralSource:   lead.ReferralSource,
		Description:      lead.Description,
		Tags:             lead.Tags,
		CustomFields:     lead.CustomFields,
		ProductInterest:  lead.ProductInterest,
		Timeline:         lead.Timeline,
		Requirements:     lead.Requirements,
		MarketingConsent: lead.MarketingConsent,
		PrivacyConsent:   lead.PrivacyConsent,
		ActivityCount:    lead.ActivityCount,
		CreatedAt:        lead.CreatedAt,
		UpdatedAt:        lead.UpdatedAt,
		CreatedBy:        lead.CreatedBy.String(),
		UpdatedBy:        lead.UpdatedBy.String(),
		Version:          lead.Version,
	}

	// Map address
	if lead.Address != nil {
		resp.Address = mapAddressDomainToDTO(lead.Address)
	}

	// Map budget
	if lead.Budget != nil {
		resp.Budget = &dto.MoneyDTO{
			Amount:   lead.Budget.Amount,
			Currency: lead.Budget.Currency,
			Display:  lead.Budget.Format(),
		}
	}

	// Map campaign ID
	if lead.CampaignID != nil {
		s := lead.CampaignID.String()
		resp.CampaignID = &s
	}

	// Map UTM params
	if lead.UTMSource != nil || lead.UTMMedium != nil || lead.UTMCampaign != nil {
		resp.UTMParams = &dto.UTMParamsDTO{
			Source:   lead.UTMSource,
			Medium:   lead.UTMMedium,
			Campaign: lead.UTMCampaign,
			Term:     lead.UTMTerm,
			Content:  lead.UTMContent,
		}
	}

	// Map owner
	if lead.OwnerID != nil {
		s := lead.OwnerID.String()
		resp.OwnerID = &s
		// Fetch owner details if needed
		if uc.userService != nil {
			if owner, err := uc.userService.GetUser(ctx, lead.TenantID, *lead.OwnerID); err == nil {
				resp.Owner = &dto.UserBriefDTO{
					ID:        owner.ID.String(),
					Name:      owner.FullName,
					Email:     owner.Email,
					AvatarURL: owner.AvatarURL,
				}
			}
		}
	}

	// Map qualification details
	if lead.QualifiedAt != nil {
		resp.QualifiedAt = lead.QualifiedAt
	}
	if lead.QualifiedBy != nil {
		s := lead.QualifiedBy.String()
		resp.QualifiedBy = &s
	}
	resp.QualificationNotes = lead.QualificationNotes

	// Map disqualification details
	if lead.DisqualifiedAt != nil {
		resp.DisqualifiedAt = lead.DisqualifiedAt
	}
	if lead.DisqualifiedBy != nil {
		s := lead.DisqualifiedBy.String()
		resp.DisqualifiedBy = &s
	}
	resp.DisqualifyReason = lead.DisqualifyReason

	// Map conversion details
	if lead.ConvertedAt != nil {
		resp.ConvertedAt = lead.ConvertedAt
	}
	if lead.ConvertedBy != nil {
		s := lead.ConvertedBy.String()
		resp.ConvertedBy = &s
	}
	if lead.OpportunityID != nil {
		s := lead.OpportunityID.String()
		resp.OpportunityID = &s
	}
	if lead.CustomerID != nil {
		s := lead.CustomerID.String()
		resp.CustomerID = &s
	}
	if lead.ContactID != nil {
		s := lead.ContactID.String()
		resp.ContactID = &s
	}

	// Map nurturing details
	if lead.NurturingAt != nil {
		resp.NurturingAt = lead.NurturingAt
	}
	if lead.NurtureCampaignID != nil {
		s := lead.NurtureCampaignID.String()
		resp.NurtureCampaignID = &s
	}

	// Map activity
	resp.LastActivityAt = lead.LastActivityAt
	resp.LastContactedAt = lead.LastContactedAt

	return resp
}

func (uc *leadUseCase) mapLeadToBriefResponse(lead *domain.Lead) *dto.LeadBriefResponse {
	var ownerID *string
	if lead.OwnerID != nil {
		s := lead.OwnerID.String()
		ownerID = &s
	}

	return &dto.LeadBriefResponse{
		ID:        lead.ID.String(),
		FirstName: lead.FirstName,
		LastName:  lead.LastName,
		FullName:  lead.FirstName + " " + lead.LastName,
		Email:     lead.Email,
		Company:   lead.Company,
		Status:    string(lead.Status),
		Score:     lead.Score,
		Rating:    lead.GetRating(),
		OwnerID:   ownerID,
		CreatedAt: lead.CreatedAt,
	}
}

func (uc *leadUseCase) mapFilterToDomain(filter *dto.LeadFilterRequest) domain.LeadFilter {
	domainFilter := domain.LeadFilter{}

	if filter == nil {
		return domainFilter
	}

	// Map statuses
	if len(filter.Statuses) > 0 {
		domainFilter.Statuses = make([]domain.LeadStatus, len(filter.Statuses))
		for i, s := range filter.Statuses {
			domainFilter.Statuses[i] = mapLeadStatus(s)
		}
	}

	// Map sources
	if len(filter.Sources) > 0 {
		domainFilter.Sources = make([]domain.LeadSource, len(filter.Sources))
		for i, s := range filter.Sources {
			domainFilter.Sources[i] = mapLeadSource(s)
		}
	}

	// Map owner IDs
	if len(filter.OwnerIDs) > 0 {
		domainFilter.OwnerIDs = make([]uuid.UUID, 0, len(filter.OwnerIDs))
		for _, id := range filter.OwnerIDs {
			if parsed, err := uuid.Parse(id); err == nil {
				domainFilter.OwnerIDs = append(domainFilter.OwnerIDs, parsed)
			}
		}
	}

	domainFilter.Unassigned = filter.Unassigned
	domainFilter.MinScore = filter.MinScore
	domainFilter.MaxScore = filter.MaxScore
	domainFilter.SearchQuery = filter.SearchQuery
	domainFilter.Tags = filter.Tags

	// Parse dates
	if filter.CreatedAfter != nil {
		if t, err := time.Parse(time.RFC3339, *filter.CreatedAfter); err == nil {
			domainFilter.CreatedAfter = &t
		}
	}
	if filter.CreatedBefore != nil {
		if t, err := time.Parse(time.RFC3339, *filter.CreatedBefore); err == nil {
			domainFilter.CreatedBefore = &t
		}
	}
	if filter.UpdatedAfter != nil {
		if t, err := time.Parse(time.RFC3339, *filter.UpdatedAfter); err == nil {
			domainFilter.UpdatedAfter = &t
		}
	}

	// Parse campaign ID
	if filter.CampaignID != nil {
		if parsed, err := uuid.Parse(*filter.CampaignID); err == nil {
			domainFilter.CampaignID = &parsed
		}
	}

	return domainFilter
}

func (uc *leadUseCase) publishEvent(ctx context.Context, event domain.DomainEvent) error {
	if uc.eventPublisher == nil {
		return nil
	}

	return uc.eventPublisher.Publish(ctx, ports.Event{
		ID:            event.EventID().String(),
		Type:          event.EventType(),
		AggregateID:   event.AggregateID().String(),
		AggregateType: event.AggregateType(),
		TenantID:      event.TenantID().String(),
		OccurredAt:    event.OccurredAt(),
		Version:       event.Version(),
	})
}

func (uc *leadUseCase) indexLead(ctx context.Context, lead *domain.Lead) {
	if uc.searchService == nil {
		return
	}

	searchable := ports.SearchableLead{
		ID:        lead.ID,
		TenantID:  lead.TenantID,
		FirstName: lead.FirstName,
		LastName:  lead.LastName,
		Email:     lead.Email,
		Company:   lead.Company,
		Phone:     lead.Phone,
		Status:    string(lead.Status),
		Source:    string(lead.Source),
		Score:     lead.Score,
		Tags:      lead.Tags,
		OwnerID:   lead.OwnerID,
		CreatedAt: lead.CreatedAt,
		UpdatedAt: lead.UpdatedAt,
	}

	uc.searchService.IndexLead(ctx, searchable)
}

func (uc *leadUseCase) invalidateLeadCache(ctx context.Context, tenantID uuid.UUID) {
	if uc.cacheService == nil {
		return
	}

	pattern := "lead:" + tenantID.String() + ":*"
	uc.cacheService.DeletePattern(ctx, pattern)
}

// ============================================================================
// Mapping Helpers
// ============================================================================

func mapLeadSource(source string) domain.LeadSource {
	switch source {
	case "website":
		return domain.LeadSourceWebsite
	case "referral":
		return domain.LeadSourceReferral
	case "social_media":
		return domain.LeadSourceSocialMedia
	case "email_campaign":
		return domain.LeadSourceEmailCampaign
	case "cold_call":
		return domain.LeadSourceColdCall
	case "trade_show":
		return domain.LeadSourceTradeShow
	case "advertisement":
		return domain.LeadSourceAdvertisement
	case "partner":
		return domain.LeadSourcePartner
	default:
		return domain.LeadSourceOther
	}
}

func mapLeadStatus(status string) domain.LeadStatus {
	switch status {
	case "new":
		return domain.LeadStatusNew
	case "contacted":
		return domain.LeadStatusContacted
	case "qualified":
		return domain.LeadStatusQualified
	case "unqualified":
		return domain.LeadStatusUnqualified
	case "converted":
		return domain.LeadStatusConverted
	case "nurturing":
		return domain.LeadStatusNurturing
	default:
		return domain.LeadStatusNew
	}
}

func mapAddressDTOToDomain(addr *dto.AddressDTO) *domain.Address {
	if addr == nil {
		return nil
	}
	return &domain.Address{
		Street1:    addr.Street1,
		Street2:    addr.Street2,
		City:       addr.City,
		State:      addr.State,
		PostalCode: addr.PostalCode,
		Country:    addr.Country,
	}
}

func mapAddressDomainToDTO(addr *domain.Address) *dto.AddressDTO {
	if addr == nil {
		return nil
	}
	return &dto.AddressDTO{
		Street1:    addr.Street1,
		Street2:    addr.Street2,
		City:       addr.City,
		State:      addr.State,
		PostalCode: addr.PostalCode,
		Country:    addr.Country,
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
