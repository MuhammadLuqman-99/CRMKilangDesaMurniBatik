package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application"
	"github.com/kilang-desa-murni/crm/internal/sales/application/dto"
	"github.com/kilang-desa-murni/crm/internal/sales/application/ports"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
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

	// Parse owner ID and get owner name
	ownerID := userID
	ownerName := ""
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
	// Get owner name
	if uc.userService != nil {
		if user, err := uc.userService.GetUser(ctx, tenantID, ownerID); err == nil && user != nil {
			ownerName = user.FullName
		}
	}

	// Parse customer ID and name (required for NewOpportunity)
	var customerID uuid.UUID
	customerName := ""
	if req.CustomerID != nil {
		customerID, _ = uuid.Parse(*req.CustomerID)
		exists, err := uc.customerService.CustomerExists(ctx, tenantID, customerID)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to verify customer", err)
		}
		if !exists {
			return nil, application.ErrCustomerNotFound(customerID)
		}
		// Get customer name
		if customer, err := uc.customerService.GetCustomer(ctx, tenantID, customerID); err == nil && customer != nil {
			customerName = customer.Name
		}
	}

	// Create amount
	amount, err := domain.NewMoney(req.Amount, req.Currency)
	if err != nil {
		return nil, application.ErrValidation("invalid amount or currency")
	}

	// Create opportunity using domain factory
	opportunity, err := domain.NewOpportunity(
		tenantID,
		req.Name,
		pipeline,
		customerID,
		customerName,
		amount,
		ownerID,
		ownerName,
		userID,
	)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeValidation, "failed to create opportunity", err)
	}

	// Set probability
	if req.Probability != nil {
		opportunity.Probability = *req.Probability
	}

	// Set optional fields
	if req.Description != nil {
		opportunity.Description = *req.Description
	}
	if req.LeadID != nil {
		leadID, _ := uuid.Parse(*req.LeadID)
		opportunity.LeadID = &leadID
	}
	// Add primary contact if specified
	if req.PrimaryContactID != nil {
		contactID, _ := uuid.Parse(*req.PrimaryContactID)
		exists, err := uc.customerService.ContactExists(ctx, tenantID, contactID)
		if err != nil {
			return nil, application.WrapError(application.ErrCodeCustomerServiceError, "failed to verify contact", err)
		}
		if !exists {
			return nil, application.ErrContactNotFound(contactID)
		}
		// Add as primary contact
		opportunity.AddContact(domain.OpportunityContact{
			ContactID:  contactID,
			CustomerID: customerID,
			Role:       "decision_maker",
			IsPrimary:  true,
		})
	}
	if req.Source != "" {
		opportunity.Source = req.Source
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
		opportunity.Notes = *req.Notes
	}

	// Add contacts
	if len(req.Contacts) > 0 {
		for _, contactReq := range req.Contacts {
			contactID, _ := uuid.Parse(contactReq.ContactID)
			opportunity.AddContact(domain.OpportunityContact{
				ContactID:  contactID,
				CustomerID: customerID,
				Role:       contactReq.Role,
				IsPrimary:  contactReq.IsPrimary,
			})
		}
	}

	// Add products
	if len(req.Products) > 0 {
		for _, productReq := range req.Products {
			productID, _ := uuid.Parse(productReq.ProductID)
			unitPrice, _ := domain.NewMoney(productReq.UnitPrice, productReq.Currency)
			discount := float64(0)
			if productReq.DiscountPercent != nil {
				discount = float64(*productReq.DiscountPercent)
			}
			product := domain.OpportunityProduct{
				ProductID:   productID,
				ProductName: productReq.ProductName,
				Quantity:    productReq.Quantity,
				UnitPrice:   unitPrice,
				Discount:    discount,
			}
			if productReq.Description != nil {
				product.Notes = *productReq.Description
			}
			opportunity.AddProduct(product)
		}
	}

	// Note: Competitors list not supported in domain - use CloseInfo.CompetitorID/Name when losing

	// Save opportunity
	if err := uc.opportunityRepo.Create(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to save opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.GetEvents() {
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
		opportunity.Description = *req.Description
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
		opportunity.ExpectedCloseDate = &expectedCloseDate
	}
	if req.CustomerID != nil {
		customerID, _ := uuid.Parse(*req.CustomerID)
		exists, _ := uc.customerService.CustomerExists(ctx, tenantID, customerID)
		if !exists {
			return nil, application.ErrCustomerNotFound(customerID)
		}
		opportunity.CustomerID = customerID
		// Get customer name
		if customer, err := uc.customerService.GetCustomer(ctx, tenantID, customerID); err == nil && customer != nil {
			opportunity.CustomerName = customer.Name
		}
	}
	if req.Source != nil {
		opportunity.Source = *req.Source
	}
	if req.Tags != nil {
		opportunity.Tags = req.Tags
	}
	if req.CustomFields != nil {
		opportunity.CustomFields = req.CustomFields
	}
	if req.Notes != nil {
		opportunity.Notes = *req.Notes
	}

	// Update metadata
	opportunity.UpdatedAt = time.Now()
	opportunity.Version++

	// Add update event
	opportunity.AddEvent(domain.NewOpportunityUpdatedEvent(opportunity))

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.GetEvents() {
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

	// Check if opportunity is closed
	if opportunity.Status.IsClosed() {
		return application.NewAppError(application.ErrCodeConflict, "cannot delete closed opportunity")
	}

	// Delete opportunity
	if err := uc.opportunityRepo.Delete(ctx, tenantID, opportunityID); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to delete opportunity", err)
	}

	// Publish delete event
	event := domain.NewOpportunityDeletedEvent(opportunity)
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
	stage := pipeline.GetStage(stageID)
	if stage == nil {
		return nil, application.ErrPipelineStageNotFound(pipeline.ID, stageID)
	}
	if !stage.IsActive {
		return nil, application.ErrPipelineStageInactive(stageID)
	}

	// Move stage
	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}
	if err := opportunity.MoveToStage(stage, userID, notes); err != nil {
		return nil, application.WrapError(application.ErrCodeOpportunityInvalidTransition, err.Error(), err)
	}

	// Update probability if provided
	if req.Probability != nil {
		opportunity.Probability = *req.Probability
	}

	// Note: weighted amount is calculated automatically in domain

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.GetEvents() {
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
		wonStage = pipeline.GetStage(stageID)
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
			s := deal.ID.String()
			dealID = &s
		}
	}

	// Publish events
	for _, event := range opportunity.GetEvents() {
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
		lostStage = pipeline.GetStage(stageID)
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
	for _, event := range opportunity.GetEvents() {
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
	stage := pipeline.GetStage(stageID)
	if stage == nil {
		return nil, application.ErrPipelineStageNotFound(pipeline.ID, stageID)
	}

	// Parse expected close date
	expectedCloseDate, err := time.Parse("2006-01-02", req.ExpectedCloseDate)
	if err != nil {
		return nil, application.ErrValidation("invalid expected_close_date format")
	}

	// Reopen the opportunity
	notes := ""
	if req.Notes != nil {
		notes = *req.Notes
	}
	if err := opportunity.Reopen(pipeline, userID, notes); err != nil {
		return nil, application.WrapError(application.ErrCodeOpportunityInvalidTransition, err.Error(), err)
	}

	// Set expected close date
	opportunity.ExpectedCloseDate = &expectedCloseDate

	// Update probability if provided
	if req.Probability != nil {
		opportunity.Probability = *req.Probability
	}

	// Note: weighted amount is calculated automatically in domain

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.GetEvents() {
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
	discount := float64(0)
	if req.DiscountPercent != nil {
		discount = float64(*req.DiscountPercent)
	}

	// Add product using OpportunityProduct struct
	product := domain.OpportunityProduct{
		ProductID:   productID,
		ProductName: req.ProductName,
		Quantity:    req.Quantity,
		UnitPrice:   unitPrice,
		Discount:    discount,
	}
	if req.Description != nil {
		product.Notes = *req.Description
	}
	opportunity.AddProduct(product)

	// Update metadata
	opportunity.UpdatedAt = time.Now()
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

	// Find the product first
	var existingProduct *domain.OpportunityProduct
	for i := range opportunity.Products {
		if opportunity.Products[i].ProductID == productID {
			existingProduct = &opportunity.Products[i]
			break
		}
	}

	if existingProduct == nil {
		return nil, application.ErrOpportunityProductNotFound(opportunityID, productID)
	}

	// Prepare update values
	quantity := existingProduct.Quantity
	unitPrice := existingProduct.UnitPrice
	discount := existingProduct.Discount
	tax := existingProduct.Tax

	if req.Quantity != nil {
		quantity = *req.Quantity
	}
	if req.UnitPrice != nil {
		unitPrice, _ = domain.NewMoney(*req.UnitPrice, opportunity.Amount.Currency)
	}
	if req.DiscountPercent != nil {
		discount = float64(*req.DiscountPercent)
	}

	// Use domain's UpdateProduct method
	if err := opportunity.UpdateProduct(existingProduct.ID, quantity, unitPrice, discount, tax); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update product", err)
	}

	// Update metadata
	opportunity.UpdatedAt = time.Now()
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

	// Check if product exists and remove
	found := false
	for _, p := range opportunity.Products {
		if p.ID == productID || p.ProductID == productID {
			found = true
			break
		}
	}
	if !found {
		return nil, application.ErrOpportunityProductNotFound(opportunityID, productID)
	}
	opportunity.RemoveProduct(productID)

	// Update metadata
	opportunity.UpdatedAt = time.Now()
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
	opportunity.AddContact(domain.OpportunityContact{
		ContactID:  contactID,
		CustomerID: opportunity.CustomerID,
		Role:       req.Role,
		IsPrimary:  req.IsPrimary,
	})

	// Update metadata
	opportunity.UpdatedAt = time.Now()
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
	contactIdx := -1
	for i, contact := range opportunity.Contacts {
		if contact.ContactID == contactID {
			contactIdx = i
			break
		}
	}

	if contactIdx == -1 {
		return nil, application.ErrOpportunityContactNotFound(opportunityID, contactID)
	}

	// Update role
	if req.Role != nil {
		opportunity.Contacts[contactIdx].Role = *req.Role
	}
	// Handle primary contact
	if req.IsPrimary != nil && *req.IsPrimary {
		// Remove primary from others first
		for i := range opportunity.Contacts {
			opportunity.Contacts[i].IsPrimary = false
		}
		opportunity.Contacts[contactIdx].IsPrimary = true
	}

	// Update metadata
	opportunity.UpdatedAt = time.Now()
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

	// Check if contact exists and remove
	contactFound := false
	for _, c := range opportunity.Contacts {
		if c.ContactID == contactID {
			contactFound = true
			break
		}
	}
	if !contactFound {
		return nil, application.ErrOpportunityContactNotFound(opportunityID, contactID)
	}
	opportunity.RemoveContact(contactID)

	// Update metadata
	opportunity.UpdatedAt = time.Now()
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

	// Note: Competitor tracking is not supported in the current domain model.
	// Competitor information is only stored when the opportunity is lost (CloseInfo).
	return nil, application.NewAppError(application.ErrCodeValidation, "competitor tracking is not supported - use loss reason instead")
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

	// Verify owner exists and get name
	exists, err := uc.userService.UserExists(ctx, tenantID, ownerID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeUserServiceError, "failed to verify owner", err)
	}
	if !exists {
		return nil, application.ErrUserNotFound(ownerID)
	}

	// Get owner name
	ownerName := ""
	if user, err := uc.userService.GetUser(ctx, tenantID, ownerID); err == nil && user != nil {
		ownerName = user.FullName
	}

	// Assign using domain method
	opportunity.AssignOwner(ownerID, ownerName)

	// Save changes
	if err := uc.opportunityRepo.Update(ctx, opportunity); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update opportunity", err)
	}

	// Publish events
	for _, event := range opportunity.GetEvents() {
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
		deal.PaymentTerm = domain.PaymentTerm(*req.PaymentTerms)
	}
	if req.DealOwnerID != nil {
		ownerID, _ := uuid.Parse(*req.DealOwnerID)
		deal.OwnerID = ownerID
	}

	// Generate deal code
	dealCode, err := uc.idGenerator.GenerateDealNumber(ctx, opportunity.TenantID)
	if err != nil {
		dealCode = "DL-" + deal.ID.String()[:8]
	}
	deal.Code = dealCode

	// Save deal
	if err := uc.dealRepo.Create(ctx, deal); err != nil {
		return nil, err
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return deal, nil
}

func (uc *opportunityUseCase) mapOpportunityToResponse(ctx context.Context, opportunity *domain.Opportunity, pipeline *domain.Pipeline) *dto.OpportunityResponse {
	resp := &dto.OpportunityResponse{
		ID:         opportunity.ID.String(),
		TenantID:   opportunity.TenantID.String(),
		Name:       opportunity.Name,
		Status:     string(opportunity.Status),
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
		Probability:  opportunity.Probability,
		Source:       opportunity.Source,
		OwnerID:      opportunity.OwnerID.String(),
		Tags:         opportunity.Tags,
		CustomFields: opportunity.CustomFields,
		CreatedAt:    opportunity.CreatedAt,
		UpdatedAt:    opportunity.UpdatedAt,
		CreatedBy:    opportunity.CreatedBy.String(),
		Version:      opportunity.Version,
	}

	// ExpectedCloseDate is *time.Time in domain, time.Time in DTO
	if opportunity.ExpectedCloseDate != nil {
		resp.ExpectedCloseDate = *opportunity.ExpectedCloseDate
	}

	// Map pipeline info
	if pipeline != nil {
		resp.Pipeline = &dto.PipelineBriefDTO{
			ID:        pipeline.ID.String(),
			Name:      pipeline.Name,
			IsDefault: pipeline.IsDefault,
		}
		if stage := pipeline.GetStage(opportunity.StageID); stage != nil {
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

	// Map optional fields - Description is string
	if opportunity.Description != "" {
		desc := opportunity.Description
		resp.Description = &desc
	}
	if opportunity.ActualCloseDate != nil {
		resp.ActualCloseDate = opportunity.ActualCloseDate
	}
	// CustomerID is uuid.UUID, not *uuid.UUID
	if opportunity.CustomerID != uuid.Nil {
		s := opportunity.CustomerID.String()
		resp.CustomerID = &s
		// Fetch customer details if needed
		if uc.customerService != nil {
			if customer, err := uc.customerService.GetCustomer(ctx, opportunity.TenantID, opportunity.CustomerID); err == nil {
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
	// Find primary contact from Contacts array
	for _, c := range opportunity.Contacts {
		if c.IsPrimary {
			s := c.ContactID.String()
			resp.PrimaryContactID = &s
			break
		}
	}
	if opportunity.CampaignID != nil {
		s := opportunity.CampaignID.String()
		resp.CampaignID = &s
	}
	// Notes is string
	if opportunity.Notes != "" {
		notes := opportunity.Notes
		resp.Notes = &notes
	}

	// Map win/loss info from CloseInfo
	if opportunity.CloseInfo != nil {
		closeInfo := opportunity.CloseInfo
		resp.ActualCloseDate = &closeInfo.ClosedAt
		closedByStr := closeInfo.ClosedBy.String()
		// Use opportunity.Status to determine win/loss (CloseInfo doesn't have Won field)
		if opportunity.Status == domain.OpportunityStatusWon {
			resp.WonAt = &closeInfo.ClosedAt
			resp.WonBy = &closedByStr
			if closeInfo.Reason != "" {
				resp.WonReason = &closeInfo.Reason
			}
			if closeInfo.Notes != "" {
				resp.WonNotes = &closeInfo.Notes
			}
		} else {
			resp.LostAt = &closeInfo.ClosedAt
			resp.LostBy = &closedByStr
			if closeInfo.Reason != "" {
				resp.LostReason = &closeInfo.Reason
			}
			if closeInfo.Notes != "" {
				resp.LostNotes = &closeInfo.Notes
			}
			if closeInfo.CompetitorID != nil {
				s := closeInfo.CompetitorID.String()
				resp.CompetitorID = &s
			}
			if closeInfo.CompetitorName != "" {
				resp.CompetitorName = &closeInfo.CompetitorName
			}
		}
	}

	// Map products - use domain fields
	resp.Products = make([]*dto.OpportunityProductResponseDTO, len(opportunity.Products))
	for i, product := range opportunity.Products {
		resp.Products[i] = &dto.OpportunityProductResponseDTO{
			ID:              product.ID.String(),
			ProductID:       product.ProductID.String(),
			ProductName:     product.ProductName,
			Quantity:        product.Quantity,
			UnitPrice: dto.MoneyDTO{
				Amount:   product.UnitPrice.Amount,
				Currency: product.UnitPrice.Currency,
			},
			DiscountPercent: int(product.Discount),
			TotalPrice: dto.MoneyDTO{
				Amount:   product.TotalPrice.Amount,
				Currency: product.TotalPrice.Currency,
			},
		}
		if product.Notes != "" {
			resp.Products[i].Description = &product.Notes
		}
	}
	resp.ProductCount = len(opportunity.Products)

	// Map contacts - domain has simpler OpportunityContact
	resp.Contacts = make([]*dto.OpportunityContactResponseDTO, len(opportunity.Contacts))
	for i, contact := range opportunity.Contacts {
		resp.Contacts[i] = &dto.OpportunityContactResponseDTO{
			ContactID: contact.ContactID.String(),
			Role:      contact.Role,
			IsPrimary: contact.IsPrimary,
		}
	}

	// No Competitors in domain - skip mapping

	// Map stage history
	resp.StageHistory = make([]*dto.StageHistoryDTO, len(opportunity.StageHistory))
	for i, history := range opportunity.StageHistory {
		stageName := history.StageName
		if stageName == "" && pipeline != nil {
			if stage := pipeline.GetStage(history.StageID); stage != nil {
				stageName = stage.Name
			}
		}
		historyDTO := &dto.StageHistoryDTO{
			StageID:   history.StageID.String(),
			StageName: stageName,
			EnteredAt: history.EnteredAt,
			ExitedAt:  history.ExitedAt,
		}
		if history.Notes != "" {
			historyDTO.Notes = &history.Notes
		}
		resp.StageHistory[i] = historyDTO
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
		if stage := pipeline.GetStage(opportunity.StageID); stage != nil {
			stageName = stage.Name
		}
	}

	var customerID, customerName *string
	// CustomerID is uuid.UUID, not *uuid.UUID
	if opportunity.CustomerID != uuid.Nil {
		s := opportunity.CustomerID.String()
		customerID = &s
		if uc.customerService != nil {
			if customer, err := uc.customerService.GetCustomer(ctx, opportunity.TenantID, opportunity.CustomerID); err == nil {
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

	resp := &dto.OpportunityBriefResponse{
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
		Probability: opportunity.Probability,
		StageID:     opportunity.StageID.String(),
		StageName:   stageName,
		CustomerID:  customerID,
		CustomerName: customerName,
		OwnerID:     opportunity.OwnerID.String(),
		OwnerName:   ownerName,
		DaysOpen:    int(time.Since(opportunity.CreatedAt).Hours() / 24),
		CreatedAt:   opportunity.CreatedAt,
	}

	// ExpectedCloseDate is *time.Time in domain, time.Time in DTO
	if opportunity.ExpectedCloseDate != nil {
		resp.ExpectedCloseDate = *opportunity.ExpectedCloseDate
	}

	return resp
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
		// Amount and WeightedAmount are Money values, not pointers
		totalValue += opp.Amount.Amount
		if opp.Amount.Currency != "" {
			currency = opp.Amount.Currency
		}
		weightedValue += opp.WeightedAmount.Amount
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
		if stage := pipeline.GetStage(opportunity.StageID); stage != nil {
			stageName = stage.Name
		}
	}

	searchable := ports.SearchableOpportunity{
		ID:          opportunity.ID,
		TenantID:    opportunity.TenantID,
		Name:        opportunity.Name,
		Status:      string(opportunity.Status),
		Amount:      opportunity.Amount.Amount,
		Currency:    opportunity.Amount.Currency,
		Probability: opportunity.Probability,
		PipelineID:  opportunity.PipelineID,
		StageID:     opportunity.StageID,
		StageName:   stageName,
		OwnerID:     opportunity.OwnerID,
		Tags:        opportunity.Tags,
		CreatedAt:   opportunity.CreatedAt,
		UpdatedAt:   opportunity.UpdatedAt,
	}

	// CustomerID is uuid.UUID in domain, *uuid.UUID in SearchableOpportunity
	if opportunity.CustomerID != uuid.Nil {
		searchable.CustomerID = &opportunity.CustomerID
	}

	// ExpectedCloseDate is *time.Time in domain, time.Time in SearchableOpportunity
	if opportunity.ExpectedCloseDate != nil {
		searchable.ExpectedCloseDate = *opportunity.ExpectedCloseDate
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
