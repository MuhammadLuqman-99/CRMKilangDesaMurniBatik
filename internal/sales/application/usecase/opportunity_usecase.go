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
// Opportunity Use Case Interface
// ============================================================================

// OpportunityUseCase defines the interface for opportunity operations.
type OpportunityUseCase interface {
	// CRUD operations
	Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateOpportunityRequest) (*dto.OpportunityResponse, error)
	GetByID(ctx context.Context, tenantID, opportunityID uuid.UUID) (*dto.OpportunityResponse, error)
	Update(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.UpdateOpportunityRequest) (*dto.OpportunityResponse, error)
	Delete(ctx context.Context, tenantID, opportunityID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *dto.OpportunityFilterRequest) (*dto.OpportunityListResponse, error)

	// Stage operations
	MoveStage(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.MoveStageRequest) (*dto.OpportunityResponse, error)
	Win(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.WinOpportunityRequest) (*dto.OpportunityWinResponse, error)
	Lose(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.LoseOpportunityRequest) (*dto.OpportunityLoseResponse, error)
	Reopen(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.ReopenOpportunityRequest) (*dto.OpportunityResponse, error)

	// Product operations
	AddProduct(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.AddProductRequest) (*dto.OpportunityResponse, error)
	UpdateProduct(ctx context.Context, tenantID, opportunityID, productID, userID uuid.UUID, req *dto.UpdateProductRequest) (*dto.OpportunityResponse, error)
	RemoveProduct(ctx context.Context, tenantID, opportunityID, productID, userID uuid.UUID) (*dto.OpportunityResponse, error)

	// Contact operations
	AddContact(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.AddContactRequest) (*dto.OpportunityResponse, error)
	UpdateContact(ctx context.Context, tenantID, opportunityID, contactID, userID uuid.UUID, req *dto.UpdateContactRequest) (*dto.OpportunityResponse, error)
	RemoveContact(ctx context.Context, tenantID, opportunityID, contactID, userID uuid.UUID) (*dto.OpportunityResponse, error)

	// Competitor operations
	AddCompetitor(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.AddCompetitorRequest) (*dto.OpportunityResponse, error)

	// Assignment
	Assign(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.AssignOpportunityRequest) (*dto.OpportunityResponse, error)
	BulkAssign(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkAssignOpportunitiesRequest) error

	// Bulk operations
	BulkMoveStage(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkMoveStageRequest) error

	// Analytics
	GetStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.OpportunityStatisticsResponse, error)
	GetPipelineAnalytics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*dto.PipelineAnalyticsResponse, error)
}

// ============================================================================
// Opportunity Use Case Implementation
// ============================================================================

// opportunityUseCase implements OpportunityUseCase.
type opportunityUseCase struct {
	opportunityRepo domain.OpportunityRepository
	pipelineRepo    domain.PipelineRepository
	dealRepo        domain.DealRepository
	eventPublisher  ports.EventPublisher
	customerService ports.CustomerService
	userService     ports.UserService
	productService  ports.ProductService
	cacheService    ports.CacheService
	searchService   ports.SearchService
	idGenerator     ports.IDGenerator
}

// NewOpportunityUseCase creates a new opportunity use case.
func NewOpportunityUseCase(
	opportunityRepo domain.OpportunityRepository,
	pipelineRepo domain.PipelineRepository,
	dealRepo domain.DealRepository,
	eventPublisher ports.EventPublisher,
	customerService ports.CustomerService,
	userService ports.UserService,
	productService ports.ProductService,
	cacheService ports.CacheService,
	searchService ports.SearchService,
	idGenerator ports.IDGenerator,
) OpportunityUseCase {
	return &opportunityUseCase{
		opportunityRepo: opportunityRepo,
		pipelineRepo:    pipelineRepo,
		dealRepo:        dealRepo,
		eventPublisher:  eventPublisher,
		customerService: customerService,
		userService:     userService,
		productService:  productService,
		cacheService:    cacheService,
		searchService:   searchService,
		idGenerator:     idGenerator,
	}
}

// ============================================================================
// CRUD Operations
// ============================================================================

// Create creates a new opportunity.
func (uc *opportunityUseCase) Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateOpportunityRequest) (*dto.OpportunityResponse, error) {
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

	// Determine stage
	var stage *domain.Stage
	if req.StageID != nil {
		stageID, _ := uuid.Parse(*req.StageID)
		stage = pipeline.GetStageByID(stageID)
		if stage == nil {
			return nil, application.ErrPipelineStageNotFound(pipelineID, stageID)
		}
	} else {
		stage = pipeline.GetFirstOpenStage()
		if stage == nil {
			return nil, application.NewAppError(application.ErrCodePipelineStageNotFound, "pipeline has no open stages")
		}
	}

	// Parse owner ID
	ownerID := userID
	if req.OwnerID != nil {
		parsedOwnerID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return nil, application.ErrValidation("invalid owner_id format")
		}
		exists, err := uc.userService.UserExists(ctx, tenantID, parsedOwnerID)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeUserServiceError, "failed to verify owner", err)
		}
		if !exists {
			return nil, application.ErrUserNotFound(parsedOwnerID)
		}
		ownerID = parsedOwnerID
	}

	// Parse expected close date
	expectedCloseDate, err := time.Parse("2006-01-02", req.ExpectedCloseDate)
	if err != nil {
		return nil, application.ErrValidation("invalid expected_close_date format")
	}

	// Create amount
	amount, err := domain.NewMoney(req.Amount, req.Currency)
	if err != nil {
		return nil, application.ErrValidation("invalid amount or currency")
	}

	// Create opportunity
	opportunityID := uc.idGenerator.GenerateID()
	opportunity, err := domain.NewOpportunity(
		opportunityID,
		tenantID,
		req.Name,
		pipelineID,
		stage.ID,
		amount,
		expectedCloseDate,
		ownerID,
		userID,
	)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeValidation, "failed to create opportunity", err)
	}

	// Set probability
	if req.Probability != nil {
		opportunity.Probability = *req.Probability
	} else {
		opportunity.Probability = stage.Probability
	}

	// Set optional fields
	if req.Description != nil {
		opportunity.Description = req.Description
	}
	if req.CustomerID != nil {
		customerID, _ := uuid.Parse(*req.CustomerID)
		exists, err := uc.customerService.CustomerExists(ctx, tenantID, customerID)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to verify customer", err)
		}
		if !exists {
			return nil, application.ErrCustomerNotFound(customerID)
		}
		opportunity.CustomerID = &customerID
	}
	if req.LeadID != nil {
		leadID, _ := uuid.Parse(*req.LeadID)
		opportunity.LeadID = &leadID
	}
	if req.PrimaryContactID != nil {
		contactID, _ := uuid.Parse(*req.PrimaryContactID)
		exists, err := uc.customerService.ContactExists(ctx, tenantID, contactID)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to verify contact", err)
		}
		if !exists {
			return nil, application.ErrContactNotFound(contactID)
		}
		opportunity.PrimaryContactID = &contactID
	}
	if req.Source != "" {
		opportunity.Source = req.Source
	}
	if req.SourceDetails != nil {
		opportunity.SourceDetails = req.SourceDetails
	}
	if req.CampaignID != nil {
		campaignID, _ := uuid.Parse(*req.CampaignID)
		opportunity.CampaignID = &campaignID
	}
	if req.Tags != nil {
		opportunity.Tags = req.Tags
	}
	if req.CustomFields != nil {
		opportunity.CustomFields = req.CustomFields
	}
	if req.Notes != nil {
		opportunity.Notes = req.Notes
	}

	// Add contacts
	if len(req.Contacts) > 0 {
		for _, contactReq := range req.Contacts {
			contactID, _ := uuid.Parse(contactReq.ContactID)
			role := domain.ContactRole(contactReq.Role)
			opportunity.AddContact(contactID, role, contactReq.IsPrimary, userID)
		}
	}

	// Add products
	if len(req.Products) > 0 {
		for _, productReq := range req.Products {
			productID, _ := uuid.Parse(productReq.ProductID)
			unitPrice, _ := domain.NewMoney(productReq.UnitPrice, productReq.Currency)
			discountPct := 0
			if productReq.DiscountPercent != nil {
				discountPct = *productReq.DiscountPercent
			}
			var discountAmt *domain.Money
			if productReq.DiscountAmount != nil {
				discountAmt, _ = domain.NewMoney(*productReq.DiscountAmount, productReq.Currency)
			}
			opportunity.AddProduct(productID, productReq.ProductName, productReq.Quantity, unitPrice, discountPct, discountAmt, productReq.Description)
		}
	}

	// Add competitors
	if len(req.Competitors) > 0 {
		for _, compReq := range req.Competitors {
			competitor := &domain.Competitor{
				ID:          uc.idGenerator.GenerateID(),
				Name:        compReq.Name,
				Website:     compReq.Website,
				Strengths:   compReq.Strengths,
				Weaknesses:  compReq.Weaknesses,
				ThreatLevel: domain.ThreatLevel(compReq.ThreatLevel),
				Notes:       compReq.Notes,
			}
			opportunity.Competitors = append(opportunity.Competitors, competitor)
		}
	}

	// Calculate weighted amount
	opportunity.CalculateWeightedAmount()

	// Save opportunity
	if err := uc.opportunityRepo.Create(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to save opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.Events() {
		uc.publishEvent(ctx, event)
	}
	opportunity.ClearEvents()

	// Index for search
	if uc.searchService != nil {
		go uc.indexOpportunity(context.Background(), opportunity, pipeline)
	}

	// Invalidate cache
	uc.invalidateOpportunityCache(ctx, tenantID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// GetByID retrieves an opportunity by ID.
func (uc *opportunityUseCase) GetByID(ctx context.Context, tenantID, opportunityID uuid.UUID) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// Update updates an opportunity.
func (uc *opportunityUseCase) Update(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.UpdateOpportunityRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	// Check version
	if opportunity.Version != req.Version {
		return nil, application.ErrVersionMismatch(req.Version, opportunity.Version)
	}

	// Check if opportunity can be updated
	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Update fields
	if req.Name != nil {
		opportunity.Name = *req.Name
	}
	if req.Description != nil {
		opportunity.Description = req.Description
	}
	if req.Amount != nil && req.Currency != nil {
		amount, _ := domain.NewMoney(*req.Amount, *req.Currency)
		opportunity.Amount = amount
	}
	if req.Probability != nil {
		opportunity.Probability = *req.Probability
	}
	if req.ExpectedCloseDate != nil {
		expectedCloseDate, _ := time.Parse("2006-01-02", *req.ExpectedCloseDate)
		opportunity.ExpectedCloseDate = expectedCloseDate
	}
	if req.CustomerID != nil {
		customerID, _ := uuid.Parse(*req.CustomerID)
		exists, _ := uc.customerService.CustomerExists(ctx, tenantID, customerID)
		if !exists {
			return nil, application.ErrCustomerNotFound(customerID)
		}
		opportunity.CustomerID = &customerID
	}
	if req.PrimaryContactID != nil {
		contactID, _ := uuid.Parse(*req.PrimaryContactID)
		exists, _ := uc.customerService.ContactExists(ctx, tenantID, contactID)
		if !exists {
			return nil, application.ErrContactNotFound(contactID)
		}
		opportunity.PrimaryContactID = &contactID
	}
	if req.Source != nil {
		opportunity.Source = *req.Source
	}
	if req.SourceDetails != nil {
		opportunity.SourceDetails = req.SourceDetails
	}
	if req.Tags != nil {
		opportunity.Tags = req.Tags
	}
	if req.CustomFields != nil {
		opportunity.CustomFields = req.CustomFields
	}
	if req.Notes != nil {
		opportunity.Notes = req.Notes
	}

	// Recalculate weighted amount
	opportunity.CalculateWeightedAmount()

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.UpdatedBy = userID
	opportunity.Version++

	// Add update event
	opportunity.AddEvent(domain.NewOpportunityUpdatedEvent(opportunity, userID))

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.Events() {
		uc.publishEvent(ctx, event)
	}
	opportunity.ClearEvents()

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	// Update search index
	if uc.searchService != nil {
		go uc.indexOpportunity(context.Background(), opportunity, pipeline)
	}

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// Delete deletes an opportunity.
func (uc *opportunityUseCase) Delete(ctx context.Context, tenantID, opportunityID, userID uuid.UUID) error {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return application.ErrOpportunityNotFound(opportunityID)
	}

	// Check if opportunity can be deleted (no deal created)
	if opportunity.DealID != nil {
		return application.NewAppError(application.ErrCodeConflict, "cannot delete opportunity with associated deal")
	}

	// Delete opportunity
	if err := uc.opportunityRepo.Delete(ctx, tenantID, opportunityID); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to delete opportunity", err)
	}

	// Publish delete event
	event := domain.NewOpportunityDeletedEvent(opportunity, userID)
	uc.publishEvent(ctx, event)

	// Remove from search index
	if uc.searchService != nil {
		go uc.searchService.DeleteIndex(context.Background(), tenantID, "opportunity", opportunityID)
	}

	// Invalidate cache
	uc.invalidateOpportunityCache(ctx, tenantID)

	return nil
}

// List lists opportunities with filtering.
func (uc *opportunityUseCase) List(ctx context.Context, tenantID uuid.UUID, filter *dto.OpportunityFilterRequest) (*dto.OpportunityListResponse, error) {
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

	// Get opportunities
	opportunities, total, err := uc.opportunityRepo.List(ctx, tenantID, domainFilter, opts)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to list opportunities", err)
	}

	// Get pipelines for stage names
	pipelineMap := make(map[uuid.UUID]*domain.Pipeline)

	// Map to response
	opportunityResponses := make([]*dto.OpportunityBriefResponse, len(opportunities))
	for i, opp := range opportunities {
		// Get pipeline if not cached
		if _, ok := pipelineMap[opp.PipelineID]; !ok {
			pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opp.PipelineID)
			pipelineMap[opp.PipelineID] = pipeline
		}
		opportunityResponses[i] = uc.mapOpportunityToBriefResponse(ctx, opp, pipelineMap[opp.PipelineID])
	}

	// Calculate summary
	summary := uc.calculateListSummary(opportunities)

	return &dto.OpportunityListResponse{
		Opportunities: opportunityResponses,
		Pagination:    dto.NewPaginationResponse(opts.Page, opts.PageSize, total),
		Summary:       summary,
	}, nil
}

// ============================================================================
// Stage Operations
// ============================================================================

// MoveStage moves an opportunity to a different stage.
func (uc *opportunityUseCase) MoveStage(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.MoveStageRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	// Check if opportunity is open
	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Get pipeline
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(opportunity.PipelineID)
	}

	// Parse stage ID
	stageID, err := uuid.Parse(req.StageID)
	if err != nil {
		return nil, application.ErrValidation("invalid stage_id format")
	}

	// Get stage
	stage := pipeline.GetStageByID(stageID)
	if stage == nil {
		return nil, application.ErrPipelineStageNotFound(pipeline.ID, stageID)
	}
	if !stage.IsActive {
		return nil, application.ErrPipelineStageInactive(stageID)
	}

	// Move stage
	if err := opportunity.MoveToStage(stage, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeOpportunityInvalidTransition, err.Error(), err)
	}

	// Update probability if provided
	if req.Probability != nil {
		opportunity.Probability = *req.Probability
	}

	// Recalculate weighted amount
	opportunity.CalculateWeightedAmount()

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.Events() {
		uc.publishEvent(ctx, event)
	}
	opportunity.ClearEvents()

	// Update search index
	if uc.searchService != nil {
		go uc.indexOpportunity(context.Background(), opportunity, pipeline)
	}

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// Win marks an opportunity as won.
func (uc *opportunityUseCase) Win(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.WinOpportunityRequest) (*dto.OpportunityWinResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	// Check if opportunity is open
	if opportunity.Status == domain.OpportunityStatusWon {
		return nil, application.ErrOpportunityAlreadyWon(opportunityID)
	}
	if opportunity.Status == domain.OpportunityStatusLost {
		return nil, application.ErrOpportunityAlreadyLost(opportunityID)
	}

	// Get pipeline
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(opportunity.PipelineID)
	}

	// Get won stage
	var wonStage *domain.Stage
	if req.WonStageID != nil {
		stageID, _ := uuid.Parse(*req.WonStageID)
		wonStage = pipeline.GetStageByID(stageID)
	} else {
		// Get first won stage
		for _, s := range pipeline.Stages {
			if s.Type == domain.StageTypeWon && s.IsActive {
				wonStage = s
				break
			}
		}
	}

	if wonStage == nil {
		return nil, application.NewAppError(application.ErrCodePipelineStageNotFound, "no won stage found in pipeline")
	}

	// Win the opportunity
	notes := ""
	if req.WonNotes != nil {
		notes = *req.WonNotes
	}
	if err := opportunity.Win(wonStage, req.WonReason, notes, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeOpportunityWinFailed, err.Error(), err)
	}

	// Update actual amount if provided
	if req.ActualAmount != nil {
		actualAmount, _ := domain.NewMoney(*req.ActualAmount, opportunity.Amount.Currency)
		opportunity.Amount = actualAmount
		opportunity.CalculateWeightedAmount()
	}

	// Update actual close date if provided
	if req.ActualCloseDate != nil {
		actualCloseDate, _ := time.Parse("2006-01-02", *req.ActualCloseDate)
		opportunity.ActualCloseDate = &actualCloseDate
	}

	// Save opportunity
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Create deal if requested
	var dealID *string
	if req.CreateDeal {
		deal, err := uc.createDealFromOpportunity(ctx, opportunity, userID, req)
		if err != nil {
			// Log error but don't fail
		} else {
			// Update opportunity with deal ID
			opportunity.DealID = &deal.ID
			uc.opportunityRepo.Update(ctx, opportunity)
			s := deal.ID.String()
			dealID = &s
		}
	}

	// Publish events
	for _, event := range opportunity.Events() {
		uc.publishEvent(ctx, event)
	}
	opportunity.ClearEvents()

	// Update search index
	if uc.searchService != nil {
		go uc.indexOpportunity(context.Background(), opportunity, pipeline)
	}

	// Invalidate cache
	uc.invalidateOpportunityCache(ctx, tenantID)

	return &dto.OpportunityWinResponse{
		OpportunityID: opportunityID.String(),
		Status:        string(opportunity.Status),
		DealID:        dealID,
		Message:       "Opportunity marked as won successfully",
	}, nil
}

// Lose marks an opportunity as lost.
func (uc *opportunityUseCase) Lose(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.LoseOpportunityRequest) (*dto.OpportunityLoseResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	// Check if opportunity is open
	if opportunity.Status == domain.OpportunityStatusWon {
		return nil, application.ErrOpportunityAlreadyWon(opportunityID)
	}
	if opportunity.Status == domain.OpportunityStatusLost {
		return nil, application.ErrOpportunityAlreadyLost(opportunityID)
	}

	// Get pipeline
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(opportunity.PipelineID)
	}

	// Get lost stage
	var lostStage *domain.Stage
	if req.LostStageID != nil {
		stageID, _ := uuid.Parse(*req.LostStageID)
		lostStage = pipeline.GetStageByID(stageID)
	} else {
		// Get first lost stage
		for _, s := range pipeline.Stages {
			if s.Type == domain.StageTypeLost && s.IsActive {
				lostStage = s
				break
			}
		}
	}

	if lostStage == nil {
		return nil, application.NewAppError(application.ErrCodePipelineStageNotFound, "no lost stage found in pipeline")
	}

	// Prepare competitor info
	var competitorID *uuid.UUID
	if req.CompetitorID != nil {
		id, _ := uuid.Parse(*req.CompetitorID)
		competitorID = &id
	}
	competitorName := ""
	if req.CompetitorName != nil {
		competitorName = *req.CompetitorName
	}

	notes := ""
	if req.LostNotes != nil {
		notes = *req.LostNotes
	}

	// Lose the opportunity
	if err := opportunity.Lose(lostStage, req.LostReason, notes, competitorID, competitorName, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeOpportunityLoseFailed, err.Error(), err)
	}

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.Events() {
		uc.publishEvent(ctx, event)
	}
	opportunity.ClearEvents()

	// Update search index
	if uc.searchService != nil {
		go uc.indexOpportunity(context.Background(), opportunity, pipeline)
	}

	// Invalidate cache
	uc.invalidateOpportunityCache(ctx, tenantID)

	return &dto.OpportunityLoseResponse{
		OpportunityID: opportunityID.String(),
		Status:        string(opportunity.Status),
		Message:       "Opportunity marked as lost",
	}, nil
}

// Reopen reopens a closed opportunity.
func (uc *opportunityUseCase) Reopen(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.ReopenOpportunityRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	// Get pipeline
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(opportunity.PipelineID)
	}

	// Parse stage ID
	stageID, err := uuid.Parse(req.StageID)
	if err != nil {
		return nil, application.ErrValidation("invalid stage_id format")
	}

	// Get stage
	stage := pipeline.GetStageByID(stageID)
	if stage == nil {
		return nil, application.ErrPipelineStageNotFound(pipeline.ID, stageID)
	}

	// Parse expected close date
	expectedCloseDate, err := time.Parse("2006-01-02", req.ExpectedCloseDate)
	if err != nil {
		return nil, application.ErrValidation("invalid expected_close_date format")
	}

	// Reopen the opportunity
	if err := opportunity.Reopen(stage, expectedCloseDate, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeOpportunityInvalidTransition, err.Error(), err)
	}

	// Update probability if provided
	if req.Probability != nil {
		opportunity.Probability = *req.Probability
	}

	// Recalculate weighted amount
	opportunity.CalculateWeightedAmount()

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.Events() {
		uc.publishEvent(ctx, event)
	}
	opportunity.ClearEvents()

	// Update search index
	if uc.searchService != nil {
		go uc.indexOpportunity(context.Background(), opportunity, pipeline)
	}

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// ============================================================================
// Product Operations
// ============================================================================

// AddProduct adds a product to an opportunity.
func (uc *opportunityUseCase) AddProduct(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.AddProductRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Parse product ID
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, application.ErrValidation("invalid product_id format")
	}

	// Verify product exists
	exists, err := uc.productService.ProductExists(ctx, tenantID, productID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeProductServiceError, "failed to verify product", err)
	}
	if !exists {
		return nil, application.ErrProductNotFound(productID)
	}

	// Create money
	unitPrice, err := domain.NewMoney(req.UnitPrice, req.Currency)
	if err != nil {
		return nil, application.ErrValidation("invalid unit_price or currency")
	}

	// Prepare discount
	discountPct := 0
	if req.DiscountPercent != nil {
		discountPct = *req.DiscountPercent
	}
	var discountAmt *domain.Money
	if req.DiscountAmount != nil {
		discountAmt, _ = domain.NewMoney(*req.DiscountAmount, req.Currency)
	}

	// Add product
	if err := opportunity.AddProduct(productID, req.ProductName, req.Quantity, unitPrice, discountPct, discountAmt, req.Description); err != nil {
		return nil, application.WrapError(application.ErrCodeOpportunityProductDuplicate, err.Error(), err)
	}

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.UpdatedBy = userID
	opportunity.Version++

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// UpdateProduct updates a product in an opportunity.
func (uc *opportunityUseCase) UpdateProduct(ctx context.Context, tenantID, opportunityID, productID, userID uuid.UUID, req *dto.UpdateProductRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Find and update product
	found := false
	for _, product := range opportunity.Products {
		if product.ProductID == productID {
			found = true
			if req.Quantity != nil {
				product.Quantity = *req.Quantity
			}
			if req.UnitPrice != nil {
				product.UnitPrice.Amount = *req.UnitPrice
			}
			if req.DiscountPercent != nil {
				product.DiscountPercent = *req.DiscountPercent
			}
			if req.DiscountAmount != nil {
				product.DiscountAmount.Amount = *req.DiscountAmount
			}
			if req.Description != nil {
				product.Description = req.Description
			}
			product.CalculateTotal()
			break
		}
	}

	if !found {
		return nil, application.ErrOpportunityProductNotFound(opportunityID, productID)
	}

	// Recalculate opportunity total
	opportunity.RecalculateAmount()

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.UpdatedBy = userID
	opportunity.Version++

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// RemoveProduct removes a product from an opportunity.
func (uc *opportunityUseCase) RemoveProduct(ctx context.Context, tenantID, opportunityID, productID, userID uuid.UUID) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Remove product
	if err := opportunity.RemoveProduct(productID); err != nil {
		return nil, application.ErrOpportunityProductNotFound(opportunityID, productID)
	}

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.UpdatedBy = userID
	opportunity.Version++

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// ============================================================================
// Contact Operations
// ============================================================================

// AddContact adds a contact to an opportunity.
func (uc *opportunityUseCase) AddContact(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.AddContactRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Parse contact ID
	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		return nil, application.ErrValidation("invalid contact_id format")
	}

	// Verify contact exists
	exists, err := uc.customerService.ContactExists(ctx, tenantID, contactID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to verify contact", err)
	}
	if !exists {
		return nil, application.ErrContactNotFound(contactID)
	}

	// Add contact
	role := domain.ContactRole(req.Role)
	if err := opportunity.AddContact(contactID, role, req.IsPrimary, userID); err != nil {
		return nil, application.ErrOpportunityContactDuplicate(opportunityID, contactID)
	}

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.UpdatedBy = userID
	opportunity.Version++

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// UpdateContact updates a contact in an opportunity.
func (uc *opportunityUseCase) UpdateContact(ctx context.Context, tenantID, opportunityID, contactID, userID uuid.UUID, req *dto.UpdateContactRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Find and update contact
	found := false
	for _, contact := range opportunity.Contacts {
		if contact.ContactID == contactID {
			found = true
			if req.Role != nil {
				contact.Role = domain.ContactRole(*req.Role)
			}
			if req.IsPrimary != nil && *req.IsPrimary {
				// Remove primary from others
				for _, c := range opportunity.Contacts {
					c.IsPrimary = false
				}
				contact.IsPrimary = true
				opportunity.PrimaryContactID = &contactID
			}
			if req.Notes != nil {
				contact.Notes = req.Notes
			}
			break
		}
	}

	if !found {
		return nil, application.ErrOpportunityContactNotFound(opportunityID, contactID)
	}

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.UpdatedBy = userID
	opportunity.Version++

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// RemoveContact removes a contact from an opportunity.
func (uc *opportunityUseCase) RemoveContact(ctx context.Context, tenantID, opportunityID, contactID, userID uuid.UUID) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Remove contact
	if err := opportunity.RemoveContact(contactID); err != nil {
		return nil, application.ErrOpportunityContactNotFound(opportunityID, contactID)
	}

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.UpdatedBy = userID
	opportunity.Version++

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// ============================================================================
// Competitor Operations
// ============================================================================

// AddCompetitor adds a competitor to an opportunity.
func (uc *opportunityUseCase) AddCompetitor(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.AddCompetitorRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	if opportunity.Status != domain.OpportunityStatusOpen {
		return nil, application.ErrOpportunityClosed(opportunityID)
	}

	// Create competitor
	competitor := &domain.Competitor{
		ID:          uc.idGenerator.GenerateID(),
		Name:        req.Name,
		Website:     req.Website,
		Strengths:   req.Strengths,
		Weaknesses:  req.Weaknesses,
		ThreatLevel: domain.ThreatLevel(req.ThreatLevel),
		Notes:       req.Notes,
	}

	opportunity.Competitors = append(opportunity.Competitors, competitor)

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.UpdatedBy = userID
	opportunity.Version++

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// ============================================================================
// Assignment
// ============================================================================

// Assign assigns an opportunity to an owner.
func (uc *opportunityUseCase) Assign(ctx context.Context, tenantID, opportunityID, userID uuid.UUID, req *dto.AssignOpportunityRequest) (*dto.OpportunityResponse, error) {
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
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

	// Assign
	if err := opportunity.Assign(ownerID, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeOpportunityAssignmentFailed, err.Error(), err)
	}

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.Events() {
		uc.publishEvent(ctx, event)
	}
	opportunity.ClearEvents()

	// Get pipeline
	pipeline, _ := uc.pipelineRepo.GetByID(ctx, tenantID, opportunity.PipelineID)

	return uc.mapOpportunityToResponse(ctx, opportunity, pipeline), nil
}

// BulkAssign bulk assigns opportunities.
func (uc *opportunityUseCase) BulkAssign(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkAssignOpportunitiesRequest) error {
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

	// Parse opportunity IDs
	opportunityIDs := make([]uuid.UUID, len(req.OpportunityIDs))
	for i, id := range req.OpportunityIDs {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return application.ErrValidation("invalid opportunity_id format")
		}
		opportunityIDs[i] = parsedID
	}

	// Bulk update
	if err := uc.opportunityRepo.BulkUpdateOwner(ctx, tenantID, opportunityIDs, ownerID); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to bulk assign opportunities", err)
	}

	// Invalidate cache
	uc.invalidateOpportunityCache(ctx, tenantID)

	return nil
}

// ============================================================================
// Bulk Operations
// ============================================================================

// BulkMoveStage bulk moves opportunities to a stage.
func (uc *opportunityUseCase) BulkMoveStage(ctx context.Context, tenantID, userID uuid.UUID, req *dto.BulkMoveStageRequest) error {
	// Parse stage ID
	stageID, err := uuid.Parse(req.StageID)
	if err != nil {
		return application.ErrValidation("invalid stage_id format")
	}

	// Parse opportunity IDs
	opportunityIDs := make([]uuid.UUID, len(req.OpportunityIDs))
	for i, id := range req.OpportunityIDs {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return application.ErrValidation("invalid opportunity_id format")
		}
		opportunityIDs[i] = parsedID
	}

	// Bulk update
	if err := uc.opportunityRepo.BulkUpdateStage(ctx, tenantID, opportunityIDs, stageID); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to bulk move stages", err)
	}

	// Invalidate cache
	uc.invalidateOpportunityCache(ctx, tenantID)

	return nil
}

// ============================================================================
// Analytics
// ============================================================================

// GetStatistics retrieves opportunity statistics.
func (uc *opportunityUseCase) GetStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.OpportunityStatisticsResponse, error) {
	// Get counts by status
	byStatus, err := uc.opportunityRepo.CountByStatus(ctx, tenantID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to get status counts", err)
	}

	// Calculate totals
	var total, open, won, lost int64
	for status, count := range byStatus {
		total += count
		switch status {
		case domain.OpportunityStatusOpen:
			open = count
		case domain.OpportunityStatusWon:
			won = count
		case domain.OpportunityStatusLost:
			lost = count
		}
	}

	// Get pipeline values (using default currency)
	currency := "USD"
	totalPipelineValue, _ := uc.opportunityRepo.GetTotalPipelineValue(ctx, tenantID, currency)
	weightedPipelineValue, _ := uc.opportunityRepo.GetWeightedPipelineValue(ctx, tenantID, currency)

	// Get win rate
	now := time.Now()
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	winRate, _ := uc.opportunityRepo.GetWinRate(ctx, tenantID, startOfYear, now)

	// Get average deal size
	avgDealSize, _ := uc.opportunityRepo.GetAverageDealSize(ctx, tenantID, currency, startOfYear, now)

	// Get average sales cycle
	avgSalesCycle, _ := uc.opportunityRepo.GetAverageSalesCycle(ctx, tenantID, startOfYear, now)

	// Get closing counts
	closingThisMonth, _, _ := uc.opportunityRepo.GetClosingThisMonth(ctx, tenantID, domain.ListOptions{PageSize: 1})
	closingThisQuarter, _, _ := uc.opportunityRepo.GetClosingThisQuarter(ctx, tenantID, domain.ListOptions{PageSize: 1})

	// Map status to string
	statusMap := make(map[string]int64)
	for status, count := range byStatus {
		statusMap[string(status)] = count
	}

	return &dto.OpportunityStatisticsResponse{
		TotalOpportunities: total,
		OpenOpportunities:  open,
		WonOpportunities:   won,
		LostOpportunities:  lost,
		TotalPipelineValue: dto.MoneyDTO{
			Amount:   totalPipelineValue,
			Currency: currency,
		},
		WeightedPipelineValue: dto.MoneyDTO{
			Amount:   weightedPipelineValue,
			Currency: currency,
		},
		WinRate: winRate,
		AverageDealSize: dto.MoneyDTO{
			Amount:   avgDealSize,
			Currency: currency,
		},
		AverageSalesCycle:  avgSalesCycle,
		ByStatus:           statusMap,
		ClosingThisMonth:   int64(len(closingThisMonth)),
		ClosingThisQuarter: int64(len(closingThisQuarter)),
	}, nil
}

// GetPipelineAnalytics retrieves pipeline analytics.
func (uc *opportunityUseCase) GetPipelineAnalytics(ctx context.Context, tenantID, pipelineID uuid.UUID) (*dto.PipelineAnalyticsResponse, error) {
	// Get pipeline
	pipeline, err := uc.pipelineRepo.GetByID(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.ErrPipelineNotFound(pipelineID)
	}

	// Get pipeline statistics
	stats, err := uc.pipelineRepo.GetPipelineStatistics(ctx, tenantID, pipelineID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to get pipeline statistics", err)
	}

	// Build stage analytics
	stageAnalytics := make([]*dto.StageAnalyticsDTO, 0, len(pipeline.Stages))
	for _, stage := range pipeline.Stages {
		if !stage.IsActive {
			continue
		}

		stageStats, _ := uc.pipelineRepo.GetStageStatistics(ctx, tenantID, pipelineID, stage.ID)
		if stageStats == nil {
			continue
		}

		stageAnalytics = append(stageAnalytics, &dto.StageAnalyticsDTO{
			StageID:           stage.ID.String(),
			StageName:         stage.Name,
			StageOrder:        stage.Order,
			OpportunityCount:  stageStats.TotalOpportunities,
			TotalValue: dto.MoneyDTO{
				Amount:   stageStats.TotalValue.Amount,
				Currency: stageStats.TotalValue.Currency,
			},
			AverageTimeInStage: stageStats.AverageTimeInStage,
			ConversionRate:     stageStats.ConversionRate,
		})
	}

	return &dto.PipelineAnalyticsResponse{
		PipelineID:        pipelineID.String(),
		PipelineName:      pipeline.Name,
		TotalOpportunities: stats.TotalOpportunities,
		TotalValue: dto.MoneyDTO{
			Amount:   stats.TotalValue.Amount,
			Currency: stats.TotalValue.Currency,
		},
		WeightedValue: dto.MoneyDTO{
			Amount:   stats.WeightedValue.Amount,
			Currency: stats.WeightedValue.Currency,
		},
		WinRate:           stats.WinRate,
		AverageSalesCycle: stats.AverageSalesCycle,
		Stages:            stageAnalytics,
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (uc *opportunityUseCase) createDealFromOpportunity(ctx context.Context, opportunity *domain.Opportunity, userID uuid.UUID, req *dto.WinOpportunityRequest) (*domain.Deal, error) {
	deal, err := domain.NewDealFromOpportunity(opportunity, userID)
	if err != nil {
		return nil, err
	}

	// Set additional details from request
	if req.PaymentTerms != nil {
		deal.PaymentTerms = *req.PaymentTerms
	}
	if req.ContractTerms != nil {
		deal.ContractTerms = req.ContractTerms
	}
	if req.DealOwnerID != nil {
		ownerID, _ := uuid.Parse(*req.DealOwnerID)
		deal.OwnerID = ownerID
	}

	// Generate deal number
	dealNumber, err := uc.idGenerator.GenerateDealNumber(ctx, opportunity.TenantID)
	if err != nil {
		dealNumber = deal.ID.String()[:8]
	}
	deal.DealNumber = dealNumber

	// Save deal
	if err := uc.dealRepo.Create(ctx, deal); err != nil {
		return nil, err
	}

	// Publish events
	for _, event := range deal.Events() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return deal, nil
}

func (uc *opportunityUseCase) mapOpportunityToResponse(ctx context.Context, opportunity *domain.Opportunity, pipeline *domain.Pipeline) *dto.OpportunityResponse {
	resp := &dto.OpportunityResponse{
		ID:       opportunity.ID.String(),
		TenantID: opportunity.TenantID.String(),
		Name:     opportunity.Name,
		Status:   string(opportunity.Status),
		PipelineID: opportunity.PipelineID.String(),
		StageID:    opportunity.StageID.String(),
		Amount: dto.MoneyDTO{
			Amount:   opportunity.Amount.Amount,
			Currency: opportunity.Amount.Currency,
			Display:  opportunity.Amount.Format(),
		},
		WeightedAmount: dto.MoneyDTO{
			Amount:   opportunity.WeightedAmount.Amount,
			Currency: opportunity.WeightedAmount.Currency,
			Display:  opportunity.WeightedAmount.Format(),
		},
		Probability:       opportunity.Probability,
		ExpectedCloseDate: opportunity.ExpectedCloseDate,
		Source:            opportunity.Source,
		OwnerID:           opportunity.OwnerID.String(),
		Tags:              opportunity.Tags,
		CustomFields:      opportunity.CustomFields,
		CreatedAt:         opportunity.CreatedAt,
		UpdatedAt:         opportunity.UpdatedAt,
		CreatedBy:         opportunity.CreatedBy.String(),
		UpdatedBy:         opportunity.UpdatedBy.String(),
		Version:           opportunity.Version,
	}

	// Map pipeline info
	if pipeline != nil {
		resp.Pipeline = &dto.PipelineBriefDTO{
			ID:        pipeline.ID.String(),
			Name:      pipeline.Name,
			IsDefault: pipeline.IsDefault,
		}
		if stage := pipeline.GetStageByID(opportunity.StageID); stage != nil {
			resp.Stage = &dto.StageBriefDTO{
				ID:          stage.ID.String(),
				Name:        stage.Name,
				Type:        string(stage.Type),
				Order:       stage.Order,
				Probability: stage.Probability,
				Color:       stage.Color,
			}
		}
	}

	// Map optional fields
	if opportunity.Description != nil {
		resp.Description = opportunity.Description
	}
	if opportunity.ActualCloseDate != nil {
		resp.ActualCloseDate = opportunity.ActualCloseDate
	}
	if opportunity.CustomerID != nil {
		s := opportunity.CustomerID.String()
		resp.CustomerID = &s
		// Fetch customer details if needed
		if uc.customerService != nil {
			if customer, err := uc.customerService.GetCustomer(ctx, opportunity.TenantID, *opportunity.CustomerID); err == nil {
				resp.Customer = &dto.CustomerBriefDTO{
					ID:     customer.ID.String(),
					Name:   customer.Name,
					Code:   customer.Code,
					Type:   customer.Type,
					Status: customer.Status,
				}
			}
		}
	}
	if opportunity.LeadID != nil {
		s := opportunity.LeadID.String()
		resp.LeadID = &s
	}
	if opportunity.PrimaryContactID != nil {
		s := opportunity.PrimaryContactID.String()
		resp.PrimaryContactID = &s
	}
	if opportunity.SourceDetails != nil {
		resp.SourceDetails = opportunity.SourceDetails
	}
	if opportunity.CampaignID != nil {
		s := opportunity.CampaignID.String()
		resp.CampaignID = &s
	}
	if opportunity.Notes != nil {
		resp.Notes = opportunity.Notes
	}

	// Map win/loss info
	if opportunity.WonAt != nil {
		resp.WonAt = opportunity.WonAt
	}
	if opportunity.WonBy != nil {
		s := opportunity.WonBy.String()
		resp.WonBy = &s
	}
	if opportunity.WonReason != nil {
		resp.WonReason = opportunity.WonReason
	}
	if opportunity.WonNotes != nil {
		resp.WonNotes = opportunity.WonNotes
	}
	if opportunity.LostAt != nil {
		resp.LostAt = opportunity.LostAt
	}
	if opportunity.LostBy != nil {
		s := opportunity.LostBy.String()
		resp.LostBy = &s
	}
	if opportunity.LostReason != nil {
		resp.LostReason = opportunity.LostReason
	}
	if opportunity.LostNotes != nil {
		resp.LostNotes = opportunity.LostNotes
	}
	if opportunity.LostToCompetitorID != nil {
		s := opportunity.LostToCompetitorID.String()
		resp.CompetitorID = &s
	}
	if opportunity.LostToCompetitorName != "" {
		resp.CompetitorName = &opportunity.LostToCompetitorName
	}

	// Map deal ID
	if opportunity.DealID != nil {
		s := opportunity.DealID.String()
		resp.DealID = &s
	}

	// Map products
	resp.Products = make([]*dto.OpportunityProductResponseDTO, len(opportunity.Products))
	for i, product := range opportunity.Products {
		resp.Products[i] = &dto.OpportunityProductResponseDTO{
			ID:          product.ID.String(),
			ProductID:   product.ProductID.String(),
			ProductName: product.ProductName,
			Quantity:    product.Quantity,
			UnitPrice: dto.MoneyDTO{
				Amount:   product.UnitPrice.Amount,
				Currency: product.UnitPrice.Currency,
			},
			DiscountPercent: product.DiscountPercent,
			DiscountAmount: dto.MoneyDTO{
				Amount:   product.DiscountAmount.Amount,
				Currency: product.DiscountAmount.Currency,
			},
			TotalPrice: dto.MoneyDTO{
				Amount:   product.TotalPrice.Amount,
				Currency: product.TotalPrice.Currency,
			},
			Description: product.Description,
		}
	}
	resp.ProductCount = len(opportunity.Products)

	// Map contacts
	resp.Contacts = make([]*dto.OpportunityContactResponseDTO, len(opportunity.Contacts))
	for i, contact := range opportunity.Contacts {
		resp.Contacts[i] = &dto.OpportunityContactResponseDTO{
			ContactID: contact.ContactID.String(),
			Role:      string(contact.Role),
			IsPrimary: contact.IsPrimary,
			Notes:     contact.Notes,
			AddedAt:   contact.AddedAt,
		}
	}

	// Map competitors
	resp.Competitors = make([]*dto.CompetitorDTO, len(opportunity.Competitors))
	for i, comp := range opportunity.Competitors {
		resp.Competitors[i] = &dto.CompetitorDTO{
			ID:          comp.ID.String(),
			Name:        comp.Name,
			Website:     comp.Website,
			Strengths:   comp.Strengths,
			Weaknesses:  comp.Weaknesses,
			ThreatLevel: string(comp.ThreatLevel),
			Notes:       comp.Notes,
		}
	}

	// Map stage history
	resp.StageHistory = make([]*dto.StageHistoryDTO, len(opportunity.StageHistory))
	for i, history := range opportunity.StageHistory {
		stageName := ""
		if pipeline != nil {
			if stage := pipeline.GetStageByID(history.StageID); stage != nil {
				stageName = stage.Name
			}
		}
		resp.StageHistory[i] = &dto.StageHistoryDTO{
			StageID:   history.StageID.String(),
			StageName: stageName,
			EnteredAt: history.EnteredAt,
			ExitedAt:  history.ExitedAt,
			ChangedBy: history.ChangedBy.String(),
			Notes:     history.Notes,
		}
	}

	// Calculate days
	resp.DaysOpen = int(time.Since(opportunity.CreatedAt).Hours() / 24)
	if len(opportunity.StageHistory) > 0 {
		lastHistory := opportunity.StageHistory[len(opportunity.StageHistory)-1]
		resp.DaysInStage = int(time.Since(lastHistory.EnteredAt).Hours() / 24)
	}

	// Get owner info
	if uc.userService != nil {
		if owner, err := uc.userService.GetUser(ctx, opportunity.TenantID, opportunity.OwnerID); err == nil {
			resp.Owner = &dto.UserBriefDTO{
				ID:        owner.ID.String(),
				Name:      owner.FullName,
				Email:     owner.Email,
				AvatarURL: owner.AvatarURL,
			}
		}
	}

	return resp
}

func (uc *opportunityUseCase) mapOpportunityToBriefResponse(ctx context.Context, opportunity *domain.Opportunity, pipeline *domain.Pipeline) *dto.OpportunityBriefResponse {
	stageName := ""
	if pipeline != nil {
		if stage := pipeline.GetStageByID(opportunity.StageID); stage != nil {
			stageName = stage.Name
		}
	}

	var customerID, customerName *string
	if opportunity.CustomerID != nil {
		s := opportunity.CustomerID.String()
		customerID = &s
		if uc.customerService != nil {
			if customer, err := uc.customerService.GetCustomer(ctx, opportunity.TenantID, *opportunity.CustomerID); err == nil {
				customerName = &customer.Name
			}
		}
	}

	ownerName := ""
	if uc.userService != nil {
		if owner, err := uc.userService.GetUser(ctx, opportunity.TenantID, opportunity.OwnerID); err == nil {
			ownerName = owner.FullName
		}
	}

	return &dto.OpportunityBriefResponse{
		ID:     opportunity.ID.String(),
		Name:   opportunity.Name,
		Status: string(opportunity.Status),
		Amount: dto.MoneyDTO{
			Amount:   opportunity.Amount.Amount,
			Currency: opportunity.Amount.Currency,
			Display:  opportunity.Amount.Format(),
		},
		WeightedAmount: dto.MoneyDTO{
			Amount:   opportunity.WeightedAmount.Amount,
			Currency: opportunity.WeightedAmount.Currency,
			Display:  opportunity.WeightedAmount.Format(),
		},
		Probability:       opportunity.Probability,
		StageID:           opportunity.StageID.String(),
		StageName:         stageName,
		ExpectedCloseDate: opportunity.ExpectedCloseDate,
		CustomerID:        customerID,
		CustomerName:      customerName,
		OwnerID:           opportunity.OwnerID.String(),
		OwnerName:         ownerName,
		DaysOpen:          int(time.Since(opportunity.CreatedAt).Hours() / 24),
		CreatedAt:         opportunity.CreatedAt,
	}
}

func (uc *opportunityUseCase) mapFilterToDomain(filter *dto.OpportunityFilterRequest) domain.OpportunityFilter {
	domainFilter := domain.OpportunityFilter{}

	if filter == nil {
		return domainFilter
	}

	// Map statuses
	if len(filter.Statuses) > 0 {
		domainFilter.Statuses = make([]domain.OpportunityStatus, len(filter.Statuses))
		for i, s := range filter.Statuses {
			domainFilter.Statuses[i] = domain.OpportunityStatus(s)
		}
	}

	// Map pipeline IDs
	if len(filter.PipelineIDs) > 0 {
		domainFilter.PipelineIDs = make([]uuid.UUID, 0, len(filter.PipelineIDs))
		for _, id := range filter.PipelineIDs {
			if parsed, err := uuid.Parse(id); err == nil {
				domainFilter.PipelineIDs = append(domainFilter.PipelineIDs, parsed)
			}
		}
	}

	// Map stage IDs
	if len(filter.StageIDs) > 0 {
		domainFilter.StageIDs = make([]uuid.UUID, 0, len(filter.StageIDs))
		for _, id := range filter.StageIDs {
			if parsed, err := uuid.Parse(id); err == nil {
				domainFilter.StageIDs = append(domainFilter.StageIDs, parsed)
			}
		}
	}

	// Map customer IDs
	if len(filter.CustomerIDs) > 0 {
		domainFilter.CustomerIDs = make([]uuid.UUID, 0, len(filter.CustomerIDs))
		for _, id := range filter.CustomerIDs {
			if parsed, err := uuid.Parse(id); err == nil {
				domainFilter.CustomerIDs = append(domainFilter.CustomerIDs, parsed)
			}
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

	domainFilter.MinAmount = filter.MinAmount
	domainFilter.MaxAmount = filter.MaxAmount
	domainFilter.Currency = filter.Currency
	domainFilter.MinProbability = filter.MinProbability
	domainFilter.MaxProbability = filter.MaxProbability
	domainFilter.SearchQuery = filter.SearchQuery

	// Parse dates
	if filter.ExpectedCloseDateAfter != nil {
		if t, err := time.Parse("2006-01-02", *filter.ExpectedCloseDateAfter); err == nil {
			domainFilter.ExpectedCloseDateAfter = &t
		}
	}
	if filter.ExpectedCloseDateBefore != nil {
		if t, err := time.Parse("2006-01-02", *filter.ExpectedCloseDateBefore); err == nil {
			domainFilter.ExpectedCloseDateBefore = &t
		}
	}
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

	return domainFilter
}

func (uc *opportunityUseCase) calculateListSummary(opportunities []*domain.Opportunity) *dto.OpportunitySummaryDTO {
	if len(opportunities) == 0 {
		return nil
	}

	var totalValue, weightedValue int64
	var totalProbability int
	currency := "USD"

	for _, opp := range opportunities {
		if opp.Amount != nil {
			totalValue += opp.Amount.Amount
			currency = opp.Amount.Currency
		}
		if opp.WeightedAmount != nil {
			weightedValue += opp.WeightedAmount.Amount
		}
		totalProbability += opp.Probability
	}

	avgProbability := float64(totalProbability) / float64(len(opportunities))
	totalDays := 0
	for _, opp := range opportunities {
		totalDays += int(time.Since(opp.CreatedAt).Hours() / 24)
	}
	avgDaysOpen := float64(totalDays) / float64(len(opportunities))

	return &dto.OpportunitySummaryDTO{
		TotalCount: int64(len(opportunities)),
		TotalValue: dto.MoneyDTO{
			Amount:   totalValue,
			Currency: currency,
		},
		WeightedValue: dto.MoneyDTO{
			Amount:   weightedValue,
			Currency: currency,
		},
		AverageProbability: avgProbability,
		AverageDaysOpen:    avgDaysOpen,
	}
}

func (uc *opportunityUseCase) publishEvent(ctx context.Context, event domain.DomainEvent) error {
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

func (uc *opportunityUseCase) indexOpportunity(ctx context.Context, opportunity *domain.Opportunity, pipeline *domain.Pipeline) {
	if uc.searchService == nil {
		return
	}

	stageName := ""
	if pipeline != nil {
		if stage := pipeline.GetStageByID(opportunity.StageID); stage != nil {
			stageName = stage.Name
		}
	}

	searchable := ports.SearchableOpportunity{
		ID:                opportunity.ID,
		TenantID:          opportunity.TenantID,
		Name:              opportunity.Name,
		Status:            string(opportunity.Status),
		Amount:            opportunity.Amount.Amount,
		Currency:          opportunity.Amount.Currency,
		Probability:       opportunity.Probability,
		CustomerID:        opportunity.CustomerID,
		PipelineID:        opportunity.PipelineID,
		StageID:           opportunity.StageID,
		StageName:         stageName,
		OwnerID:           opportunity.OwnerID,
		ExpectedCloseDate: opportunity.ExpectedCloseDate,
		Tags:              opportunity.Tags,
		CreatedAt:         opportunity.CreatedAt,
		UpdatedAt:         opportunity.UpdatedAt,
	}

	uc.searchService.IndexOpportunity(ctx, searchable)
}

func (uc *opportunityUseCase) invalidateOpportunityCache(ctx context.Context, tenantID uuid.UUID) {
	if uc.cacheService == nil {
		return
	}

	pattern := "opportunity:" + tenantID.String() + ":*"
	uc.cacheService.DeletePattern(ctx, pattern)
}
