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
// Deal Use Case Interface
// ============================================================================

// DealUseCase defines the interface for deal operations.
type DealUseCase interface {
	// CRUD operations
	Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateDealRequest) (*dto.DealResponse, error)
	GetByID(ctx context.Context, tenantID, dealID uuid.UUID) (*dto.DealResponse, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*dto.DealResponse, error)
	Update(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.UpdateDealRequest) (*dto.DealResponse, error)
	Delete(ctx context.Context, tenantID, dealID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *dto.DealFilterRequest) (*dto.DealListResponse, error)

	// Line item operations
	AddLineItem(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.AddLineItemRequest) (*dto.DealResponse, error)
	UpdateLineItem(ctx context.Context, tenantID, dealID, lineItemID, userID uuid.UUID, req *dto.UpdateLineItemRequest) (*dto.DealResponse, error)
	RemoveLineItem(ctx context.Context, tenantID, dealID, lineItemID, userID uuid.UUID) (*dto.DealResponse, error)
	FulfillLineItem(ctx context.Context, tenantID, dealID, lineItemID, userID uuid.UUID, quantity int) (*dto.DealResponse, error)

	// Invoice operations
	CreateInvoice(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.CreateInvoiceRequest) (*dto.DealResponse, error)
	SendInvoice(ctx context.Context, tenantID, dealID, invoiceID, userID uuid.UUID) (*dto.DealResponse, error)

	// Payment operations
	RecordPayment(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.RecordPaymentRequest) (*dto.DealResponse, error)

	// Status operations
	Activate(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error)
	Fulfill(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error)
	Cancel(ctx context.Context, tenantID, dealID, userID uuid.UUID, reason string) (*dto.DealResponse, error)
	PutOnHold(ctx context.Context, tenantID, dealID, userID uuid.UUID, reason string) (*dto.DealResponse, error)
	Resume(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error)

	// Statistics
	GetStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.DealStatisticsResponse, error)
}

// ============================================================================
// Deal Use Case Implementation
// ============================================================================

// dealUseCase implements DealUseCase.
type dealUseCase struct {
	dealRepo        domain.DealRepository
	opportunityRepo domain.OpportunityRepository
	eventPublisher  ports.EventPublisher
	customerService ports.CustomerService
	userService     ports.UserService
	productService  ports.ProductService
	cacheService    ports.CacheService
	searchService   ports.SearchService
	idGenerator     ports.IDGenerator
	notificationSvc ports.NotificationService
}

// NewDealUseCase creates a new deal use case.
func NewDealUseCase(
	dealRepo domain.DealRepository,
	opportunityRepo domain.OpportunityRepository,
	eventPublisher ports.EventPublisher,
	customerService ports.CustomerService,
	userService ports.UserService,
	productService ports.ProductService,
	cacheService ports.CacheService,
	searchService ports.SearchService,
	idGenerator ports.IDGenerator,
	notificationSvc ports.NotificationService,
) DealUseCase {
	return &dealUseCase{
		dealRepo:        dealRepo,
		opportunityRepo: opportunityRepo,
		eventPublisher:  eventPublisher,
		customerService: customerService,
		userService:     userService,
		productService:  productService,
		cacheService:    cacheService,
		searchService:   searchService,
		idGenerator:     idGenerator,
		notificationSvc: notificationSvc,
	}
}

// ============================================================================
// CRUD Operations
// ============================================================================

// Create creates a new deal from an opportunity.
func (uc *dealUseCase) Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateDealRequest) (*dto.DealResponse, error) {
	// Parse opportunity ID
	opportunityID, err := uuid.Parse(req.OpportunityID)
	if err != nil {
		return nil, application.ErrValidation("invalid opportunity_id format")
	}

	// Get opportunity
	opportunity, err := uc.opportunityRepo.GetByID(ctx, tenantID, opportunityID)
	if err != nil {
		return nil, application.ErrOpportunityNotFound(opportunityID)
	}

	// Verify opportunity is won
	if !opportunity.IsWon() {
		return nil, application.NewAppError(application.ErrCodeValidation, "opportunity must be won to create a deal")
	}

	// Create deal from opportunity
	deal, err := domain.NewDealFromOpportunity(opportunity, userID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to create deal from opportunity", err)
	}

	// Generate deal code
	if uc.idGenerator != nil {
		code, err := uc.idGenerator.GenerateDealNumber(ctx, tenantID)
		if err == nil {
			deal.Code = code
		}
	}

	// Override with request values if provided
	if req.Name != "" {
		deal.Name = req.Name
	}
	if req.Description != "" {
		deal.Description = req.Description
	}
	if req.PaymentTerm != "" {
		deal.PaymentTerm = domain.PaymentTerm(req.PaymentTerm)
		deal.PaymentTermDays = deal.PaymentTerm.DaysUntilDue()
	}
	if req.PaymentTermDays > 0 {
		deal.PaymentTerm = domain.PaymentTermCustom
		deal.PaymentTermDays = req.PaymentTermDays
	}

	// Save deal
	if err := uc.dealRepo.Create(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to save deal", err)
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	// Invalidate cache
	uc.invalidateDealCache(ctx, tenantID)

	return uc.mapDealToResponse(ctx, deal), nil
}

// GetByID retrieves a deal by ID.
func (uc *dealUseCase) GetByID(ctx context.Context, tenantID, dealID uuid.UUID) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// GetByCode retrieves a deal by code.
func (uc *dealUseCase) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByNumber(ctx, tenantID, code)
	if err != nil {
		return nil, application.ErrDealNotFound(code)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// Update updates a deal.
func (uc *dealUseCase) Update(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.UpdateDealRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Check version
	if deal.Version != req.Version {
		return nil, application.ErrVersionMismatch(req.Version, deal.Version)
	}

	// Check if deal can be updated
	if deal.Status.IsClosed() {
		return nil, application.NewAppError(application.ErrCodeDealCancelled, "deal is closed")
	}

	// Build update parameters
	name := deal.Name
	description := deal.Description
	paymentTerm := deal.PaymentTerm
	paymentTermDays := deal.PaymentTermDays

	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.PaymentTerm != nil {
		paymentTerm = domain.PaymentTerm(*req.PaymentTerm)
		paymentTermDays = paymentTerm.DaysUntilDue()
	}
	if req.PaymentTermDays != nil {
		paymentTerm = domain.PaymentTermCustom
		paymentTermDays = *req.PaymentTermDays
	}

	// Update deal
	deal.Update(name, description, paymentTerm, paymentTermDays)
	deal.Version++

	// Update tags if provided
	if req.Tags != nil {
		deal.Tags = req.Tags
	}

	// Update custom fields if provided
	if req.CustomFields != nil {
		deal.CustomFields = req.CustomFields
	}

	// Update notes if provided
	if req.Notes != nil {
		deal.Notes = *req.Notes
	}

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// Delete deletes a deal.
func (uc *dealUseCase) Delete(ctx context.Context, tenantID, dealID, userID uuid.UUID) error {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return application.ErrDealNotFound(dealID)
	}

	// Delete deal
	if err := deal.Delete(); err != nil {
		return application.WrapError(application.ErrCodeDealCannotCancel, err.Error(), err)
	}

	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to delete deal", err)
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}

	// Invalidate cache
	uc.invalidateDealCache(ctx, tenantID)

	return nil
}

// List lists deals with filtering.
func (uc *dealUseCase) List(ctx context.Context, tenantID uuid.UUID, filter *dto.DealFilterRequest) (*dto.DealListResponse, error) {
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

	// Get deals
	deals, total, err := uc.dealRepo.List(ctx, tenantID, domainFilter, opts)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to list deals", err)
	}

	// Map to response
	dealResponses := make([]*dto.DealBriefResponse, len(deals))
	for i, deal := range deals {
		dealResponses[i] = uc.mapDealToBriefResponse(deal)
	}

	return &dto.DealListResponse{
		Deals:      dealResponses,
		Pagination: dto.NewPaginationResponse(opts.Page, opts.PageSize, total),
	}, nil
}

// ============================================================================
// Line Item Operations
// ============================================================================

// AddLineItem adds a line item to a deal.
func (uc *dealUseCase) AddLineItem(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.AddLineItemRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Parse product ID
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, application.ErrValidation("invalid product_id format")
	}

	// Create unit price
	unitPrice, err := domain.NewMoney(req.UnitPrice, req.Currency)
	if err != nil {
		return nil, application.ErrValidation("invalid unit price")
	}

	// Create line item
	lineItem := domain.DealLineItem{
		ProductID:    productID,
		ProductName:  req.ProductName,
		ProductSKU:   req.ProductSKU,
		Description:  req.Description,
		Quantity:     req.Quantity,
		UnitPrice:    unitPrice,
		Discount:     req.Discount,
		DiscountType: req.DiscountType,
		Tax:          req.Tax,
		TaxType:      req.TaxType,
		Notes:        req.Notes,
	}

	// Add line item
	if err := deal.AddLineItem(lineItem); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// UpdateLineItem updates a line item in a deal.
func (uc *dealUseCase) UpdateLineItem(ctx context.Context, tenantID, dealID, lineItemID, userID uuid.UUID, req *dto.UpdateLineItemRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Get current values
	var quantity int
	var unitPrice domain.Money
	var discount, tax float64

	for _, item := range deal.LineItems {
		if item.ID == lineItemID {
			quantity = item.Quantity
			unitPrice = item.UnitPrice
			discount = item.Discount
			tax = item.Tax
			break
		}
	}

	// Override with request values
	if req.Quantity != nil {
		quantity = *req.Quantity
	}
	if req.UnitPrice != nil {
		unitPrice, _ = domain.NewMoney(*req.UnitPrice, unitPrice.Currency)
	}
	if req.Discount != nil {
		discount = *req.Discount
	}
	if req.Tax != nil {
		tax = *req.Tax
	}

	// Update line item
	if err := deal.UpdateLineItem(lineItemID, quantity, unitPrice, discount, tax); err != nil {
		return nil, application.ErrDealLineItemNotFound(dealID, lineItemID)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// RemoveLineItem removes a line item from a deal.
func (uc *dealUseCase) RemoveLineItem(ctx context.Context, tenantID, dealID, lineItemID, userID uuid.UUID) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Remove line item
	if err := deal.RemoveLineItem(lineItemID); err != nil {
		return nil, application.ErrDealLineItemNotFound(dealID, lineItemID)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// FulfillLineItem marks quantity as fulfilled for a line item.
func (uc *dealUseCase) FulfillLineItem(ctx context.Context, tenantID, dealID, lineItemID, userID uuid.UUID, quantity int) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Fulfill line item
	if err := deal.FulfillLineItem(lineItemID, quantity); err != nil {
		return nil, application.WrapError(application.ErrCodeDealFulfillmentExceeds, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// ============================================================================
// Invoice Operations
// ============================================================================

// CreateInvoice creates a new invoice for a deal.
func (uc *dealUseCase) CreateInvoice(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.CreateInvoiceRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return nil, application.ErrValidation("invalid due_date format")
	}

	// Create amount
	amount, err := domain.NewMoney(req.Amount, deal.Currency)
	if err != nil {
		return nil, application.ErrValidation("invalid amount")
	}

	// Create invoice
	_, err = deal.CreateInvoice(req.InvoiceNumber, amount, dueDate)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// SendInvoice marks an invoice as sent.
func (uc *dealUseCase) SendInvoice(ctx context.Context, tenantID, dealID, invoiceID, userID uuid.UUID) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Send invoice
	if err := deal.SendInvoice(invoiceID); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// ============================================================================
// Payment Operations
// ============================================================================

// RecordPayment records a payment for a deal.
func (uc *dealUseCase) RecordPayment(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.RecordPaymentRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Create amount
	amount, err := domain.NewMoney(req.Amount, req.Currency)
	if err != nil {
		return nil, application.ErrValidation("invalid amount")
	}

	// Parse received date
	receivedAt := time.Now().UTC()
	if req.PaymentDate != "" {
		receivedAt, _ = time.Parse("2006-01-02", req.PaymentDate)
	}

	// Create payment
	payment := domain.Payment{
		Amount:        amount,
		PaymentMethod: req.PaymentMethod,
		ReceivedAt:    receivedAt,
		ReceivedBy:    userID,
	}
	if req.ReferenceNumber != nil {
		payment.Reference = *req.ReferenceNumber
	}
	if req.Notes != nil {
		payment.Notes = *req.Notes
	}

	// Link to invoice if provided
	if req.InvoiceID != nil {
		invoiceID, _ := uuid.Parse(*req.InvoiceID)
		payment.InvoiceID = &invoiceID
	}

	// Record payment
	if err := deal.RecordPayment(payment); err != nil {
		return nil, application.WrapError(application.ErrCodeDealPaymentExceedsBalance, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// ============================================================================
// Status Operations
// ============================================================================

// Activate activates a deal.
func (uc *dealUseCase) Activate(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Activate deal
	if err := deal.Activate(); err != nil {
		return nil, application.WrapError(application.ErrCodeDealInvalidTransition, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// Fulfill marks a deal as fulfilled.
func (uc *dealUseCase) Fulfill(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Fulfill deal
	if err := deal.Fulfill(); err != nil {
		return nil, application.WrapError(application.ErrCodeDealInvalidTransition, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// Cancel cancels a deal.
func (uc *dealUseCase) Cancel(ctx context.Context, tenantID, dealID, userID uuid.UUID, reason string) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Cancel deal
	if err := deal.Cancel(reason); err != nil {
		return nil, application.WrapError(application.ErrCodeDealInvalidTransition, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.GetEvents() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// PutOnHold puts a deal on hold.
func (uc *dealUseCase) PutOnHold(ctx context.Context, tenantID, dealID, userID uuid.UUID, reason string) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Put on hold
	if err := deal.PutOnHold(reason); err != nil {
		return nil, application.WrapError(application.ErrCodeDealInvalidTransition, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// Resume resumes a deal from hold.
func (uc *dealUseCase) Resume(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Resume
	if err := deal.Resume(); err != nil {
		return nil, application.WrapError(application.ErrCodeDealInvalidTransition, err.Error(), err)
	}

	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	return uc.mapDealToResponse(ctx, deal), nil
}

// ============================================================================
// Statistics
// ============================================================================

// GetStatistics retrieves deal statistics.
func (uc *dealUseCase) GetStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.DealStatisticsResponse, error) {
	// Get counts by status
	byStatus, err := uc.dealRepo.CountByStatus(ctx, tenantID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to get status counts", err)
	}

	// Calculate totals
	var total int64
	for _, count := range byStatus {
		total += count
	}

	// Map status to string
	statusMap := make(map[string]int64)
	for status, count := range byStatus {
		statusMap[string(status)] = count
	}

	return &dto.DealStatisticsResponse{
		TotalDeals: total,
		ByStatus:   statusMap,
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (uc *dealUseCase) mapDealToResponse(ctx context.Context, deal *domain.Deal) *dto.DealResponse {
	resp := &dto.DealResponse{
		ID:          deal.ID.String(),
		TenantID:    deal.TenantID.String(),
		Code:        deal.Code,
		Name:        deal.Name,
		Description: deal.Description,
		Status:      string(deal.Status),
		CustomerID:  deal.CustomerID.String(),
		CustomerName: deal.CustomerName,
		OwnerID:     deal.OwnerID.String(),
		OwnerName:   deal.OwnerName,
		Currency:    deal.Currency,
		Subtotal: dto.MoneyDTO{
			Amount:   deal.Subtotal.Amount,
			Currency: deal.Subtotal.Currency,
			Display:  deal.Subtotal.Format(),
		},
		TotalDiscount: dto.MoneyDTO{
			Amount:   deal.TotalDiscount.Amount,
			Currency: deal.TotalDiscount.Currency,
			Display:  deal.TotalDiscount.Format(),
		},
		TotalTax: dto.MoneyDTO{
			Amount:   deal.TotalTax.Amount,
			Currency: deal.TotalTax.Currency,
			Display:  deal.TotalTax.Format(),
		},
		TotalAmount: dto.MoneyDTO{
			Amount:   deal.TotalAmount.Amount,
			Currency: deal.TotalAmount.Currency,
			Display:  deal.TotalAmount.Format(),
		},
		PaidAmount: dto.MoneyDTO{
			Amount:   deal.PaidAmount.Amount,
			Currency: deal.PaidAmount.Currency,
			Display:  deal.PaidAmount.Format(),
		},
		OutstandingAmount: dto.MoneyDTO{
			Amount:   deal.OutstandingAmount.Amount,
			Currency: deal.OutstandingAmount.Currency,
			Display:  deal.OutstandingAmount.Format(),
		},
		PaymentTerm:         string(deal.PaymentTerm),
		PaymentTermDays:     deal.PaymentTermDays,
		FulfillmentProgress: deal.FulfillmentProgress(),
		PaymentProgress:     deal.PaymentProgress(),
		IsFullyPaid:         deal.IsFullyPaid(),
		Tags:                deal.Tags,
		CustomFields:        deal.CustomFields,
		Notes:               deal.Notes,
		ContractURL:         deal.ContractURL,
		WonAt:               &deal.WonAt,
		ActivatedAt:         deal.ActivatedAt,
		FulfilledAt:         deal.FulfilledAt,
		CancelledAt:         deal.CancelledAt,
		CreatedBy:           deal.CreatedBy.String(),
		CreatedAt:           deal.CreatedAt,
		UpdatedAt:           deal.UpdatedAt,
		Version:             deal.Version,
	}

	// Map opportunity info
	resp.OpportunityID = deal.OpportunityID.String()
	resp.PipelineID = deal.PipelineID.String()
	resp.WonReason = deal.WonReason

	// Map primary contact
	if deal.PrimaryContactID != nil {
		s := deal.PrimaryContactID.String()
		resp.PrimaryContactID = &s
		resp.PrimaryContactName = deal.PrimaryContactName
	}

	// Map team
	if deal.TeamID != nil {
		s := deal.TeamID.String()
		resp.TeamID = &s
	}

	// Map timeline
	resp.Timeline = &dto.DealTimelineDTO{
		QuoteDate:       deal.Timeline.QuoteDate,
		ContractDate:    deal.Timeline.ContractDate,
		StartDate:       deal.Timeline.StartDate,
		EndDate:         deal.Timeline.EndDate,
		RenewalDate:     deal.Timeline.RenewalDate,
		FirstPaymentDue: deal.Timeline.FirstPaymentDue,
	}

	// Map line items
	resp.LineItems = make([]*dto.DealLineItemDTO, len(deal.LineItems))
	for i, item := range deal.LineItems {
		resp.LineItems[i] = &dto.DealLineItemDTO{
			ID:          item.ID.String(),
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductName,
			ProductSKU:  item.ProductSKU,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice: dto.MoneyDTO{
				Amount:   item.UnitPrice.Amount,
				Currency: item.UnitPrice.Currency,
			},
			Discount:     item.Discount,
			DiscountType: item.DiscountType,
			Tax:          item.Tax,
			TaxType:      item.TaxType,
			Subtotal: dto.MoneyDTO{
				Amount:   item.Subtotal.Amount,
				Currency: item.Subtotal.Currency,
			},
			TaxAmount: dto.MoneyDTO{
				Amount:   item.TaxAmount.Amount,
				Currency: item.TaxAmount.Currency,
			},
			Total: dto.MoneyDTO{
				Amount:   item.Total.Amount,
				Currency: item.Total.Currency,
			},
			FulfilledQty:      item.FulfilledQty,
			RemainingQuantity: item.RemainingQuantity(),
			IsFulfilled:       item.IsFulfilled(),
			DeliveryDate:      item.DeliveryDate,
			Notes:             item.Notes,
		}
	}

	// Map invoices
	resp.Invoices = make([]*dto.InvoiceDTO, len(deal.Invoices))
	for i, inv := range deal.Invoices {
		resp.Invoices[i] = &dto.InvoiceDTO{
			ID:            inv.ID.String(),
			InvoiceNumber: inv.InvoiceNumber,
			Amount: dto.MoneyDTO{
				Amount:   inv.Amount.Amount,
				Currency: inv.Amount.Currency,
			},
			DueDate:    inv.DueDate,
			Status:     inv.Status,
			SentAt:     inv.SentAt,
			PaidAt:     inv.PaidAt,
			PaidAmount: dto.MoneyDTO{
				Amount:   inv.PaidAmount.Amount,
				Currency: inv.PaidAmount.Currency,
			},
			OutstandingAmount: dto.MoneyDTO{
				Amount:   inv.OutstandingAmount().Amount,
				Currency: inv.Amount.Currency,
			},
			IsOverdue: inv.IsOverdue(),
			Notes:     inv.Notes,
			CreatedAt: inv.CreatedAt,
		}
	}

	// Map payments
	resp.Payments = make([]*dto.PaymentDTO, len(deal.Payments))
	for i, payment := range deal.Payments {
		resp.Payments[i] = &dto.PaymentDTO{
			ID: payment.ID.String(),
			Amount: dto.MoneyDTO{
				Amount:   payment.Amount.Amount,
				Currency: payment.Amount.Currency,
			},
			PaymentMethod: payment.PaymentMethod,
			Reference:     payment.Reference,
			ReceivedAt:    payment.ReceivedAt,
			ReceivedBy:    payment.ReceivedBy.String(),
			Notes:         payment.Notes,
		}
		if payment.InvoiceID != nil {
			s := payment.InvoiceID.String()
			resp.Payments[i].InvoiceID = &s
		}
	}

	// Get overdue invoices count
	overdueInvoices := deal.GetOverdueInvoices()
	resp.OverdueInvoiceCount = len(overdueInvoices)

	return resp
}

func (uc *dealUseCase) mapDealToBriefResponse(deal *domain.Deal) *dto.DealBriefResponse {
	return &dto.DealBriefResponse{
		ID:           deal.ID.String(),
		Code:         deal.Code,
		Name:         deal.Name,
		Status:       string(deal.Status),
		CustomerID:   deal.CustomerID.String(),
		CustomerName: deal.CustomerName,
		OwnerID:      deal.OwnerID.String(),
		OwnerName:    deal.OwnerName,
		TotalAmount: dto.MoneyDTO{
			Amount:   deal.TotalAmount.Amount,
			Currency: deal.TotalAmount.Currency,
			Display:  deal.TotalAmount.Format(),
		},
		OutstandingAmount: dto.MoneyDTO{
			Amount:   deal.OutstandingAmount.Amount,
			Currency: deal.OutstandingAmount.Currency,
			Display:  deal.OutstandingAmount.Format(),
		},
		FulfillmentProgress: deal.FulfillmentProgress(),
		PaymentProgress:     deal.PaymentProgress(),
		IsFullyPaid:         deal.IsFullyPaid(),
		WonAt:               &deal.WonAt,
		CreatedAt:           deal.CreatedAt,
	}
}

func (uc *dealUseCase) mapFilterToDomain(filter *dto.DealFilterRequest) domain.DealFilter {
	domainFilter := domain.DealFilter{}

	if filter == nil {
		return domainFilter
	}

	// Map statuses
	if len(filter.Statuses) > 0 {
		domainFilter.Statuses = make([]domain.DealStatus, len(filter.Statuses))
		for i, s := range filter.Statuses {
			domainFilter.Statuses[i] = domain.DealStatus(s)
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
	domainFilter.SearchQuery = filter.SearchQuery

	return domainFilter
}

func (uc *dealUseCase) publishEvent(ctx context.Context, event domain.DomainEvent) error {
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

func (uc *dealUseCase) invalidateDealCache(ctx context.Context, tenantID uuid.UUID) {
	if uc.cacheService == nil {
		return
	}

	pattern := "deal:" + tenantID.String() + ":*"
	uc.cacheService.DeletePattern(ctx, pattern)
}
