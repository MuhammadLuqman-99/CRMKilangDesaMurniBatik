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
// Deal Use Case Interface
// ============================================================================

// DealUseCase defines the interface for deal operations.
type DealUseCase interface {
	// CRUD operations
	Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateDealRequest) (*dto.DealResponse, error)
	GetByID(ctx context.Context, tenantID, dealID uuid.UUID) (*dto.DealResponse, error)
	GetByNumber(ctx context.Context, tenantID uuid.UUID, dealNumber string) (*dto.DealResponse, error)
	Update(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.UpdateDealRequest) (*dto.DealResponse, error)
	Delete(ctx context.Context, tenantID, dealID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *dto.DealFilterRequest) (*dto.DealListResponse, error)

	// Line item operations
	AddLineItem(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.AddLineItemRequest) (*dto.DealResponse, error)
	UpdateLineItem(ctx context.Context, tenantID, dealID, lineItemID, userID uuid.UUID, req *dto.UpdateLineItemRequest) (*dto.DealResponse, error)
	RemoveLineItem(ctx context.Context, tenantID, dealID, lineItemID, userID uuid.UUID) (*dto.DealResponse, error)

	// Invoice operations
	GenerateInvoice(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.GenerateInvoiceRequest) (*dto.DealResponse, error)

	// Payment operations
	RecordPayment(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.RecordPaymentRequest) (*dto.DealResponse, error)

	// Fulfillment operations
	UpdateFulfillment(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.UpdateFulfillmentRequest) (*dto.DealResponse, error)
	BulkUpdateFulfillment(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.BulkUpdateFulfillmentRequest) (*dto.DealResponse, error)

	// Status operations
	Activate(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error)
	Complete(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error)
	Cancel(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.CancelDealRequest) (*dto.DealResponse, error)

	// Assignment
	Assign(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.AssignDealRequest) (*dto.DealResponse, error)

	// Statistics
	GetStatistics(ctx context.Context, tenantID uuid.UUID) (*dto.DealStatisticsResponse, error)
	GetRevenueReport(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) (*dto.RevenueReportResponse, error)
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

// Create creates a new deal.
func (uc *dealUseCase) Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateDealRequest) (*dto.DealResponse, error) {
	// Parse customer ID
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return nil, application.ErrValidation("invalid customer_id format")
	}

	// Verify customer exists
	customer, err := uc.customerService.GetCustomer(ctx, tenantID, customerID)
	if err != nil {
		return nil, application.ErrCustomerNotFound(customerID)
	}

	// Parse owner ID
	ownerID := userID
	if req.OwnerID != nil {
		parsedOwnerID, err := uuid.Parse(*req.OwnerID)
		if err != nil {
			return nil, application.ErrValidation("invalid owner_id format")
		}
		exists, _ := uc.userService.UserExists(ctx, tenantID, parsedOwnerID)
		if !exists {
			return nil, application.ErrUserNotFound(parsedOwnerID)
		}
		ownerID = parsedOwnerID
	}

	// Generate deal number
	dealNumber, err := uc.idGenerator.GenerateDealNumber(ctx, tenantID)
	if err != nil {
		return nil, application.WrapError(application.ErrCodeDealNumberGeneration, "failed to generate deal number", err)
	}

	// Create total amount
	totalAmount, err := domain.NewMoney(req.TotalAmount, req.Currency)
	if err != nil {
		return nil, application.ErrValidation("invalid total_amount or currency")
	}

	// Create deal
	dealID := uc.idGenerator.GenerateID()
	deal := &domain.Deal{
		ID:          dealID,
		TenantID:    tenantID,
		DealNumber:  dealNumber,
		Name:        req.Name,
		Status:      domain.DealStatusDraft,
		CustomerID:  customerID,
		TotalAmount: totalAmount,
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   userID,
		UpdatedBy:   userID,
		Version:     1,
	}

	// Set optional fields
	if req.Description != nil {
		deal.Description = req.Description
	}
	if req.OpportunityID != nil {
		opportunityID, _ := uuid.Parse(*req.OpportunityID)
		deal.OpportunityID = &opportunityID
	}
	if req.ClosedDate != nil {
		closedDate, _ := time.Parse("2006-01-02", *req.ClosedDate)
		deal.ClosedDate = &closedDate
	}
	if req.SignedDate != nil {
		signedDate, _ := time.Parse("2006-01-02", *req.SignedDate)
		deal.SignedDate = &signedDate
	}
	if req.StartDate != nil {
		startDate, _ := time.Parse("2006-01-02", *req.StartDate)
		deal.StartDate = &startDate
	}
	if req.EndDate != nil {
		endDate, _ := time.Parse("2006-01-02", *req.EndDate)
		deal.EndDate = &endDate
	}
	if req.ContractNumber != nil {
		deal.ContractNumber = req.ContractNumber
	}
	if req.ContractTerms != nil {
		deal.ContractTerms = req.ContractTerms
	}
	if req.PaymentTerms != nil {
		deal.PaymentTerms = *req.PaymentTerms
	} else {
		deal.PaymentTerms = "net_30"
	}
	if req.PaymentMethod != nil {
		deal.PaymentMethod = req.PaymentMethod
	}
	if req.BillingContactID != nil {
		billingContactID, _ := uuid.Parse(*req.BillingContactID)
		deal.BillingContactID = &billingContactID
	}
	if req.BillingAddress != nil {
		deal.BillingAddress = mapAddressDTOToDomain(req.BillingAddress)
	}
	if req.ShippingContactID != nil {
		shippingContactID, _ := uuid.Parse(*req.ShippingContactID)
		deal.ShippingContactID = &shippingContactID
	}
	if req.ShippingAddress != nil {
		deal.ShippingAddress = mapAddressDTOToDomain(req.ShippingAddress)
	}
	if req.ShippingMethod != nil {
		deal.ShippingMethod = req.ShippingMethod
	}
	if req.ShippingCost != nil {
		shippingCost, _ := domain.NewMoney(*req.ShippingCost, req.Currency)
		deal.ShippingCost = shippingCost
	}
	if req.TaxRate != nil {
		deal.TaxRate = *req.TaxRate
	}
	if req.TaxAmount != nil {
		taxAmount, _ := domain.NewMoney(*req.TaxAmount, req.Currency)
		deal.TaxAmount = taxAmount
	}
	if req.DiscountPercent != nil {
		deal.DiscountPercent = *req.DiscountPercent
	}
	if req.DiscountAmount != nil {
		discountAmount, _ := domain.NewMoney(*req.DiscountAmount, req.Currency)
		deal.DiscountAmount = discountAmount
	}
	if req.Tags != nil {
		deal.Tags = req.Tags
	}
	if req.CustomFields != nil {
		deal.CustomFields = req.CustomFields
	}
	if req.Notes != nil {
		deal.Notes = req.Notes
	}

	// Add line items
	if len(req.LineItems) > 0 {
		deal.LineItems = make([]*domain.LineItem, len(req.LineItems))
		for i, itemReq := range req.LineItems {
			productID, _ := uuid.Parse(itemReq.ProductID)
			unitPrice, _ := domain.NewMoney(itemReq.UnitPrice, itemReq.Currency)

			lineItem := &domain.LineItem{
				ID:          uc.idGenerator.GenerateID(),
				ProductID:   productID,
				ProductName: itemReq.ProductName,
				ProductSKU:  itemReq.ProductSKU,
				Description: itemReq.Description,
				Quantity:    itemReq.Quantity,
				UnitPrice:   unitPrice,
				Notes:       itemReq.Notes,
			}

			if itemReq.DiscountPercent != nil {
				lineItem.DiscountPercent = *itemReq.DiscountPercent
			}
			if itemReq.DiscountAmount != nil {
				discountAmt, _ := domain.NewMoney(*itemReq.DiscountAmount, itemReq.Currency)
				lineItem.DiscountAmount = discountAmt
			}
			if itemReq.TaxRate != nil {
				lineItem.TaxRate = *itemReq.TaxRate
			}

			lineItem.CalculateTotal()
			deal.LineItems[i] = lineItem
		}
	}

	// Calculate totals
	deal.RecalculateTotals()

	// Initialize payment tracking
	deal.TotalPaid, _ = domain.NewMoney(0, req.Currency)
	deal.Payments = []*domain.Payment{}
	deal.Invoices = []*domain.Invoice{}

	// Add created event
	deal.AddEvent(domain.NewDealCreatedEvent(deal, userID))

	// Save deal
	if err := uc.dealRepo.Create(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to save deal", err)
	}

	// Publish events
	for _, event := range deal.Events() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	// Index for search
	if uc.searchService != nil {
		go uc.indexDeal(context.Background(), deal, customer)
	}

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

// GetByNumber retrieves a deal by deal number.
func (uc *dealUseCase) GetByNumber(ctx context.Context, tenantID uuid.UUID, dealNumber string) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByNumber(ctx, tenantID, dealNumber)
	if err != nil {
		return nil, application.ErrDealNotFound(dealNumber)
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
	if deal.Status == domain.DealStatusCancelled {
		return nil, application.ErrDealCancelled(dealID)
	}
	if deal.Status == domain.DealStatusCompleted {
		return nil, application.ErrDealCompleted(dealID)
	}

	// Update fields
	if req.Name != nil {
		deal.Name = *req.Name
	}
	if req.Description != nil {
		deal.Description = req.Description
	}
	if req.SignedDate != nil {
		signedDate, _ := time.Parse("2006-01-02", *req.SignedDate)
		deal.SignedDate = &signedDate
	}
	if req.StartDate != nil {
		startDate, _ := time.Parse("2006-01-02", *req.StartDate)
		deal.StartDate = &startDate
	}
	if req.EndDate != nil {
		endDate, _ := time.Parse("2006-01-02", *req.EndDate)
		deal.EndDate = &endDate
	}
	if req.ContractNumber != nil {
		deal.ContractNumber = req.ContractNumber
	}
	if req.ContractTerms != nil {
		deal.ContractTerms = req.ContractTerms
	}
	if req.PaymentTerms != nil {
		deal.PaymentTerms = *req.PaymentTerms
	}
	if req.PaymentMethod != nil {
		deal.PaymentMethod = req.PaymentMethod
	}
	if req.BillingContactID != nil {
		billingContactID, _ := uuid.Parse(*req.BillingContactID)
		deal.BillingContactID = &billingContactID
	}
	if req.BillingAddress != nil {
		deal.BillingAddress = mapAddressDTOToDomain(req.BillingAddress)
	}
	if req.ShippingContactID != nil {
		shippingContactID, _ := uuid.Parse(*req.ShippingContactID)
		deal.ShippingContactID = &shippingContactID
	}
	if req.ShippingAddress != nil {
		deal.ShippingAddress = mapAddressDTOToDomain(req.ShippingAddress)
	}
	if req.ShippingMethod != nil {
		deal.ShippingMethod = req.ShippingMethod
	}
	if req.Tags != nil {
		deal.Tags = req.Tags
	}
	if req.CustomFields != nil {
		deal.CustomFields = req.CustomFields
	}
	if req.Notes != nil {
		deal.Notes = req.Notes
	}

	// Update metadata
	deal.UpdatedAt = time.Now()
	deal.UpdatedBy = userID
	deal.Version++

	// Add update event
	deal.AddEvent(domain.NewDealUpdatedEvent(deal, userID))

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.Events() {
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

	// Check if deal can be deleted
	if deal.Status == domain.DealStatusActive || deal.Status == domain.DealStatusCompleted {
		if deal.TotalPaid != nil && deal.TotalPaid.Amount > 0 {
			return application.ErrDealCannotCancel(dealID, "deal has payments")
		}
	}

	// Delete deal
	if err := uc.dealRepo.Delete(ctx, tenantID, dealID); err != nil {
		return application.WrapError(application.ErrCodeInternal, "failed to delete deal", err)
	}

	// Remove from search index
	if uc.searchService != nil {
		go uc.searchService.DeleteIndex(context.Background(), tenantID, "deal", dealID)
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
		dealResponses[i] = uc.mapDealToBriefResponse(ctx, deal)
	}

	// Calculate summary
	summary := uc.calculateListSummary(deals)

	return &dto.DealListResponse{
		Deals:      dealResponses,
		Pagination: dto.NewPaginationResponse(opts.Page, opts.PageSize, total),
		Summary:    summary,
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

	// Check if deal can be modified
	if deal.Status == domain.DealStatusCancelled {
		return nil, application.ErrDealCancelled(dealID)
	}
	if deal.Status == domain.DealStatusCompleted {
		return nil, application.ErrDealCompleted(dealID)
	}

	// Parse product ID
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, application.ErrValidation("invalid product_id format")
	}

	// Create line item
	unitPrice, _ := domain.NewMoney(req.UnitPrice, req.Currency)
	lineItem := &domain.LineItem{
		ID:          uc.idGenerator.GenerateID(),
		ProductID:   productID,
		ProductName: req.ProductName,
		ProductSKU:  req.ProductSKU,
		Description: req.Description,
		Quantity:    req.Quantity,
		UnitPrice:   unitPrice,
		Notes:       req.Notes,
	}

	if req.DiscountPercent != nil {
		lineItem.DiscountPercent = *req.DiscountPercent
	}
	if req.DiscountAmount != nil {
		discountAmt, _ := domain.NewMoney(*req.DiscountAmount, req.Currency)
		lineItem.DiscountAmount = discountAmt
	}
	if req.TaxRate != nil {
		lineItem.TaxRate = *req.TaxRate
	}

	lineItem.CalculateTotal()

	// Add line item
	if err := deal.AddLineItem(lineItem); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, err.Error(), err)
	}

	// Update metadata
	deal.UpdatedAt = time.Now()
	deal.UpdatedBy = userID
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

	// Check if deal can be modified
	if deal.Status == domain.DealStatusCancelled {
		return nil, application.ErrDealCancelled(dealID)
	}
	if deal.Status == domain.DealStatusCompleted {
		return nil, application.ErrDealCompleted(dealID)
	}

	// Find and update line item
	found := false
	for _, item := range deal.LineItems {
		if item.ID == lineItemID {
			found = true
			if req.ProductName != nil {
				item.ProductName = *req.ProductName
			}
			if req.Description != nil {
				item.Description = req.Description
			}
			if req.Quantity != nil {
				item.Quantity = *req.Quantity
			}
			if req.UnitPrice != nil {
				item.UnitPrice.Amount = *req.UnitPrice
			}
			if req.DiscountPercent != nil {
				item.DiscountPercent = *req.DiscountPercent
			}
			if req.DiscountAmount != nil {
				item.DiscountAmount.Amount = *req.DiscountAmount
			}
			if req.TaxRate != nil {
				item.TaxRate = *req.TaxRate
			}
			if req.Notes != nil {
				item.Notes = req.Notes
			}
			item.CalculateTotal()
			break
		}
	}

	if !found {
		return nil, application.ErrDealLineItemNotFound(dealID, lineItemID)
	}

	// Recalculate totals
	deal.RecalculateTotals()

	// Update metadata
	deal.UpdatedAt = time.Now()
	deal.UpdatedBy = userID
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

	// Check if deal can be modified
	if deal.Status == domain.DealStatusCancelled {
		return nil, application.ErrDealCancelled(dealID)
	}
	if deal.Status == domain.DealStatusCompleted {
		return nil, application.ErrDealCompleted(dealID)
	}

	// Remove line item
	if err := deal.RemoveLineItem(lineItemID); err != nil {
		return nil, application.ErrDealLineItemNotFound(dealID, lineItemID)
	}

	// Update metadata
	deal.UpdatedAt = time.Now()
	deal.UpdatedBy = userID
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

// GenerateInvoice generates an invoice for a deal.
func (uc *dealUseCase) GenerateInvoice(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.GenerateInvoiceRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Check if deal can have invoices
	if deal.Status == domain.DealStatusCancelled {
		return nil, application.ErrDealCancelled(dealID)
	}
	if deal.Status == domain.DealStatusDraft {
		return nil, application.NewAppError(application.ErrCodeDealInvalidStatus, "cannot generate invoice for draft deal")
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return nil, application.ErrValidation("invalid due_date format")
	}

	// Determine invoice amount
	var invoiceAmount int64
	currency := deal.TotalAmount.Currency
	if req.Amount != nil {
		invoiceAmount = *req.Amount
	} else {
		// Invoice remaining balance
		totalInvoiced := deal.GetTotalInvoiced()
		invoiceAmount = deal.TotalAmount.Amount - totalInvoiced
	}
	if req.Currency != nil {
		currency = *req.Currency
	}

	amount, _ := domain.NewMoney(invoiceAmount, currency)

	// Create invoice
	invoice := &domain.Invoice{
		ID:            uc.idGenerator.GenerateID(),
		InvoiceNumber: req.InvoiceNumber,
		Amount:        amount,
		Status:        domain.InvoiceStatusDraft,
		IssuedDate:    time.Now(),
		DueDate:       dueDate,
		Notes:         req.Notes,
		PaidAmount:    &domain.Money{Amount: 0, Currency: currency},
	}

	// Add invoice to deal
	if err := deal.GenerateInvoice(invoice); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, err.Error(), err)
	}

	// Update metadata
	deal.UpdatedAt = time.Now()
	deal.UpdatedBy = userID
	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Send invoice notification if requested
	if req.SendToCustomer && uc.notificationSvc != nil {
		// Send invoice email asynchronously
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

	// Check if deal can accept payments
	if deal.Status == domain.DealStatusCancelled {
		return nil, application.ErrDealCancelled(dealID)
	}
	if deal.Status == domain.DealStatusDraft || deal.Status == domain.DealStatusPending {
		return nil, application.NewAppError(application.ErrCodeDealInvalidStatus, "cannot record payment for draft/pending deal")
	}

	// Check payment amount
	remainingBalance := deal.GetRemainingBalance()
	if req.Amount > remainingBalance {
		return nil, application.ErrDealPaymentExceedsBalance(req.Amount, remainingBalance)
	}

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, application.ErrValidation("invalid payment_date format")
	}

	// Create payment
	amount, _ := domain.NewMoney(req.Amount, req.Currency)
	payment := domain.Payment{
		ID:              uc.idGenerator.GenerateID(),
		Amount:          amount,
		PaymentDate:     paymentDate,
		PaymentMethod:   req.PaymentMethod,
		ReferenceNumber: req.ReferenceNumber,
		Status:          domain.PaymentStatusCompleted,
		Notes:           req.Notes,
		RecordedAt:      time.Now(),
		RecordedBy:      userID,
	}

	// Link to invoice if provided
	if req.InvoiceID != nil {
		invoiceID, _ := uuid.Parse(*req.InvoiceID)
		payment.InvoiceID = &invoiceID
	}

	// Record payment
	if err := deal.RecordPayment(payment); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, err.Error(), err)
	}

	// Update metadata
	deal.UpdatedAt = time.Now()
	deal.UpdatedBy = userID
	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish payment event
	for _, event := range deal.Events() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// ============================================================================
// Fulfillment Operations
// ============================================================================

// UpdateFulfillment updates fulfillment for a line item.
func (uc *dealUseCase) UpdateFulfillment(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.UpdateFulfillmentRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Check if deal can be fulfilled
	if deal.Status == domain.DealStatusCancelled {
		return nil, application.ErrDealCancelled(dealID)
	}
	if deal.Status != domain.DealStatusActive {
		return nil, application.NewAppError(application.ErrCodeDealInvalidStatus, "can only fulfill active deals")
	}

	// Parse line item ID
	lineItemID, err := uuid.Parse(req.LineItemID)
	if err != nil {
		return nil, application.ErrValidation("invalid line_item_id format")
	}

	// Find line item
	var lineItem *domain.LineItem
	for _, item := range deal.LineItems {
		if item.ID == lineItemID {
			lineItem = item
			break
		}
	}
	if lineItem == nil {
		return nil, application.ErrDealLineItemNotFound(dealID, lineItemID)
	}

	// Check fulfillment quantity
	remainingQty := lineItem.Quantity - lineItem.FulfilledQty
	if req.Quantity > remainingQty {
		return nil, application.ErrDealFulfillmentExceedsQuantity(req.Quantity, remainingQty)
	}

	// Parse fulfilled date
	fulfilledDate := time.Now()
	if req.FulfilledDate != nil {
		fulfilledDate, _ = time.Parse("2006-01-02", *req.FulfilledDate)
	}

	// Create fulfillment record
	fulfillment := &domain.Fulfillment{
		ID:             uc.idGenerator.GenerateID(),
		LineItemID:     lineItemID,
		Quantity:       req.Quantity,
		FulfilledDate:  fulfilledDate,
		TrackingNumber: req.TrackingNumber,
		CarrierName:    req.CarrierName,
		Notes:          req.Notes,
		FulfilledBy:    userID,
	}

	// Update fulfillment
	if err := deal.UpdateFulfillment(lineItemID, req.Quantity, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, err.Error(), err)
	}

	// Add fulfillment record
	deal.Fulfillments = append(deal.Fulfillments, fulfillment)

	// Update metadata
	deal.UpdatedAt = time.Now()
	deal.UpdatedBy = userID
	deal.Version++

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.Events() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// BulkUpdateFulfillment bulk updates fulfillment for multiple line items.
func (uc *dealUseCase) BulkUpdateFulfillment(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.BulkUpdateFulfillmentRequest) (*dto.DealResponse, error) {
	for _, item := range req.Items {
		_, err := uc.UpdateFulfillment(ctx, tenantID, dealID, userID, &item)
		if err != nil {
			return nil, err
		}
	}

	return uc.GetByID(ctx, tenantID, dealID)
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
	if err := deal.Activate(userID); err != nil {
		return nil, application.WrapError(application.ErrCodeDealInvalidTransition, err.Error(), err)
	}

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.Events() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// Complete completes a deal.
func (uc *dealUseCase) Complete(ctx context.Context, tenantID, dealID, userID uuid.UUID) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Complete deal
	if err := deal.Complete(userID); err != nil {
		return nil, application.WrapError(application.ErrCodeDealInvalidTransition, err.Error(), err)
	}

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.Events() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// Cancel cancels a deal.
func (uc *dealUseCase) Cancel(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.CancelDealRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Check if deal can be cancelled
	if deal.TotalPaid != nil && deal.TotalPaid.Amount > 0 {
		return nil, application.ErrDealCannotCancel(dealID, "deal has payments recorded")
	}

	// Cancel deal
	if err := deal.Cancel(req.Reason, userID); err != nil {
		return nil, application.WrapError(application.ErrCodeDealInvalidTransition, err.Error(), err)
	}

	// Set cancel notes
	if req.Notes != nil {
		deal.CancelNotes = req.Notes
	}

	// Save changes
	if err := uc.dealRepo.Update(ctx, deal); err != nil {
		return nil, application.WrapError(application.ErrCodeInternal, "failed to update deal", err)
	}

	// Publish events
	for _, event := range deal.Events() {
		uc.publishEvent(ctx, event)
	}
	deal.ClearEvents()

	return uc.mapDealToResponse(ctx, deal), nil
}

// ============================================================================
// Assignment
// ============================================================================

// Assign assigns a deal to an owner.
func (uc *dealUseCase) Assign(ctx context.Context, tenantID, dealID, userID uuid.UUID, req *dto.AssignDealRequest) (*dto.DealResponse, error) {
	deal, err := uc.dealRepo.GetByID(ctx, tenantID, dealID)
	if err != nil {
		return nil, application.ErrDealNotFound(dealID)
	}

	// Parse owner ID
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, application.ErrValidation("invalid owner_id format")
	}

	// Verify owner exists
	exists, _ := uc.userService.UserExists(ctx, tenantID, ownerID)
	if !exists {
		return nil, application.ErrUserNotFound(ownerID)
	}

	// Assign
	deal.OwnerID = ownerID
	deal.UpdatedAt = time.Now()
	deal.UpdatedBy = userID
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

	// Get revenue data
	currency := "USD"
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	totalRevenue, _ := uc.dealRepo.GetTotalRevenue(ctx, tenantID, currency, startOfYear, now)
	totalCollected, _ := uc.dealRepo.GetTotalReceivedPayments(ctx, tenantID, currency, startOfYear, now)
	totalOutstanding, _ := uc.dealRepo.GetOutstandingAmount(ctx, tenantID, currency)
	avgDealSize, _ := uc.dealRepo.GetAverageDealValue(ctx, tenantID, currency, startOfYear, now)
	revenueThisMonth, _ := uc.dealRepo.GetTotalRevenue(ctx, tenantID, currency, startOfMonth, now)
	collectedThisMonth, _ := uc.dealRepo.GetTotalReceivedPayments(ctx, tenantID, currency, startOfMonth, now)

	// Get deal counts
	fullyPaidDeals, _, _ := uc.dealRepo.GetFullyPaidDeals(ctx, tenantID, domain.ListOptions{PageSize: 1})
	partiallyPaidDeals, _, _ := uc.dealRepo.GetDealsWithPendingPayments(ctx, tenantID, domain.ListOptions{PageSize: 1})

	dealsThisMonth, _, _ := uc.dealRepo.GetByClosedDate(ctx, tenantID, startOfMonth, now, domain.ListOptions{PageSize: 1})

	// Map status to string
	statusMap := make(map[string]int64)
	for status, count := range byStatus {
		statusMap[string(status)] = count
	}

	return &dto.DealStatisticsResponse{
		TotalDeals: total,
		ByStatus:   statusMap,
		TotalRevenue: dto.MoneyDTO{
			Amount:   totalRevenue,
			Currency: currency,
		},
		TotalCollected: dto.MoneyDTO{
			Amount:   totalCollected,
			Currency: currency,
		},
		TotalOutstanding: dto.MoneyDTO{
			Amount:   totalOutstanding,
			Currency: currency,
		},
		AverageDealSize: dto.MoneyDTO{
			Amount:   avgDealSize,
			Currency: currency,
		},
		FullyPaidDeals:     int64(len(fullyPaidDeals)),
		PartiallyPaidDeals: int64(len(partiallyPaidDeals)),
		DealsThisMonth:     int64(len(dealsThisMonth)),
		RevenueThisMonth: dto.MoneyDTO{
			Amount:   revenueThisMonth,
			Currency: currency,
		},
		CollectedThisMonth: dto.MoneyDTO{
			Amount:   collectedThisMonth,
			Currency: currency,
		},
	}, nil
}

// GetRevenueReport generates a revenue report.
func (uc *dealUseCase) GetRevenueReport(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time, groupBy string) (*dto.RevenueReportResponse, error) {
	currency := "USD"

	totalRevenue, _ := uc.dealRepo.GetTotalRevenue(ctx, tenantID, currency, startDate, endDate)
	totalCollected, _ := uc.dealRepo.GetTotalReceivedPayments(ctx, tenantID, currency, startDate, endDate)
	totalOutstanding, _ := uc.dealRepo.GetOutstandingAmount(ctx, tenantID, currency)
	avgDealSize, _ := uc.dealRepo.GetAverageDealValue(ctx, tenantID, currency, startDate, endDate)

	// Get deals in range
	deals, total, _ := uc.dealRepo.GetByClosedDate(ctx, tenantID, startDate, endDate, domain.ListOptions{PageSize: 1000})

	return &dto.RevenueReportResponse{
		Period:    groupBy,
		StartDate: startDate,
		EndDate:   endDate,
		TotalRevenue: dto.MoneyDTO{
			Amount:   totalRevenue,
			Currency: currency,
		},
		TotalCollected: dto.MoneyDTO{
			Amount:   totalCollected,
			Currency: currency,
		},
		TotalOutstanding: dto.MoneyDTO{
			Amount:   totalOutstanding,
			Currency: currency,
		},
		DealCount: total,
		AverageDealSize: dto.MoneyDTO{
			Amount:   avgDealSize,
			Currency: currency,
		},
		// Note: Additional breakdowns (ByCustomer, ByProduct, ByOwner, Trend)
		// would require additional repository methods
		Trend: uc.buildRevenueTrend(deals, startDate, endDate, groupBy, currency),
	}, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (uc *dealUseCase) mapDealToResponse(ctx context.Context, deal *domain.Deal) *dto.DealResponse {
	resp := &dto.DealResponse{
		ID:         deal.ID.String(),
		TenantID:   deal.TenantID.String(),
		DealNumber: deal.DealNumber,
		Name:       deal.Name,
		Status:     string(deal.Status),
		CustomerID: deal.CustomerID.String(),
		OwnerID:    deal.OwnerID.String(),
		PaymentTerms: deal.PaymentTerms,
		LineItemCount: len(deal.LineItems),
		InvoiceCount:  len(deal.Invoices),
		PaymentCount:  len(deal.Payments),
		FulfillmentProgress: deal.FulfillmentProgress,
		FulfillmentStatus:   deal.GetFulfillmentStatus(),
		Tags:         deal.Tags,
		CustomFields: deal.CustomFields,
		CreatedAt:    deal.CreatedAt,
		UpdatedAt:    deal.UpdatedAt,
		CreatedBy:    deal.CreatedBy.String(),
		UpdatedBy:    deal.UpdatedBy.String(),
		Version:      deal.Version,
	}

	// Map amounts
	if deal.Subtotal != nil {
		resp.Subtotal = dto.MoneyDTO{
			Amount:   deal.Subtotal.Amount,
			Currency: deal.Subtotal.Currency,
			Display:  deal.Subtotal.Format(),
		}
	}
	if deal.DiscountAmount != nil {
		resp.DiscountAmount = dto.MoneyDTO{
			Amount:   deal.DiscountAmount.Amount,
			Currency: deal.DiscountAmount.Currency,
			Display:  deal.DiscountAmount.Format(),
		}
	}
	if deal.TaxAmount != nil {
		resp.TaxAmount = dto.MoneyDTO{
			Amount:   deal.TaxAmount.Amount,
			Currency: deal.TaxAmount.Currency,
			Display:  deal.TaxAmount.Format(),
		}
	}
	if deal.ShippingCost != nil {
		resp.ShippingCost = dto.MoneyDTO{
			Amount:   deal.ShippingCost.Amount,
			Currency: deal.ShippingCost.Currency,
			Display:  deal.ShippingCost.Format(),
		}
	}
	if deal.TotalAmount != nil {
		resp.TotalAmount = dto.MoneyDTO{
			Amount:   deal.TotalAmount.Amount,
			Currency: deal.TotalAmount.Currency,
			Display:  deal.TotalAmount.Format(),
		}
	}
	if deal.TotalPaid != nil {
		resp.TotalPaid = dto.MoneyDTO{
			Amount:   deal.TotalPaid.Amount,
			Currency: deal.TotalPaid.Currency,
			Display:  deal.TotalPaid.Format(),
		}
	}

	// Calculate total pending
	if deal.TotalAmount != nil && deal.TotalPaid != nil {
		pending := deal.TotalAmount.Amount - deal.TotalPaid.Amount
		resp.TotalPending = dto.MoneyDTO{
			Amount:   pending,
			Currency: deal.TotalAmount.Currency,
		}
	}

	// Calculate total invoiced
	totalInvoiced := deal.GetTotalInvoiced()
	resp.TotalInvoiced = dto.MoneyDTO{
		Amount:   totalInvoiced,
		Currency: deal.TotalAmount.Currency,
	}

	// Payment status
	resp.PaymentStatus = deal.GetPaymentStatus()

	// Map optional fields
	if deal.Description != nil {
		resp.Description = deal.Description
	}
	if deal.OpportunityID != nil {
		s := deal.OpportunityID.String()
		resp.OpportunityID = &s
	}
	if deal.ClosedDate != nil {
		resp.ClosedDate = deal.ClosedDate
	}
	if deal.SignedDate != nil {
		resp.SignedDate = deal.SignedDate
	}
	if deal.StartDate != nil {
		resp.StartDate = deal.StartDate
	}
	if deal.EndDate != nil {
		resp.EndDate = deal.EndDate
	}
	if deal.ContractNumber != nil {
		resp.ContractNumber = deal.ContractNumber
	}
	if deal.ContractTerms != nil {
		resp.ContractTerms = deal.ContractTerms
	}
	if deal.PaymentMethod != nil {
		resp.PaymentMethod = deal.PaymentMethod
	}
	if deal.BillingContactID != nil {
		s := deal.BillingContactID.String()
		resp.BillingContactID = &s
	}
	if deal.BillingAddress != nil {
		resp.BillingAddress = mapAddressDomainToDTO(deal.BillingAddress)
	}
	if deal.ShippingContactID != nil {
		s := deal.ShippingContactID.String()
		resp.ShippingContactID = &s
	}
	if deal.ShippingAddress != nil {
		resp.ShippingAddress = mapAddressDomainToDTO(deal.ShippingAddress)
	}
	if deal.ShippingMethod != nil {
		resp.ShippingMethod = deal.ShippingMethod
	}
	if deal.Notes != nil {
		resp.Notes = deal.Notes
	}

	// Map cancellation info
	if deal.CancelledAt != nil {
		resp.CancelledAt = deal.CancelledAt
	}
	if deal.CancelledBy != nil {
		s := deal.CancelledBy.String()
		resp.CancelledBy = &s
	}
	if deal.CancelReason != nil {
		resp.CancelReason = deal.CancelReason
	}
	if deal.CancelNotes != nil {
		resp.CancelNotes = deal.CancelNotes
	}

	// Map line items
	resp.LineItems = make([]*dto.DealLineItemResponseDTO, len(deal.LineItems))
	for i, item := range deal.LineItems {
		resp.LineItems[i] = &dto.DealLineItemResponseDTO{
			ID:              item.ID.String(),
			ProductID:       item.ProductID.String(),
			ProductName:     item.ProductName,
			ProductSKU:      item.ProductSKU,
			Description:     item.Description,
			Quantity:        item.Quantity,
			DiscountPercent: item.DiscountPercent,
			TaxRate:         item.TaxRate,
			FulfilledQty:    item.FulfilledQty,
			PendingQty:      item.Quantity - item.FulfilledQty,
			Notes:           item.Notes,
		}
		if item.UnitPrice != nil {
			resp.LineItems[i].UnitPrice = dto.MoneyDTO{
				Amount:   item.UnitPrice.Amount,
				Currency: item.UnitPrice.Currency,
			}
		}
		if item.DiscountAmount != nil {
			resp.LineItems[i].DiscountAmount = dto.MoneyDTO{
				Amount:   item.DiscountAmount.Amount,
				Currency: item.DiscountAmount.Currency,
			}
		}
		if item.TaxAmount != nil {
			resp.LineItems[i].TaxAmount = dto.MoneyDTO{
				Amount:   item.TaxAmount.Amount,
				Currency: item.TaxAmount.Currency,
			}
		}
		if item.TotalPrice != nil {
			resp.LineItems[i].TotalPrice = dto.MoneyDTO{
				Amount:   item.TotalPrice.Amount,
				Currency: item.TotalPrice.Currency,
			}
		}
	}

	// Map invoices
	resp.Invoices = make([]*dto.InvoiceResponseDTO, len(deal.Invoices))
	for i, invoice := range deal.Invoices {
		resp.Invoices[i] = &dto.InvoiceResponseDTO{
			ID:            invoice.ID.String(),
			InvoiceNumber: invoice.InvoiceNumber,
			Status:        string(invoice.Status),
			IssuedDate:    invoice.IssuedDate,
			DueDate:       invoice.DueDate,
			PaidDate:      invoice.PaidDate,
			Notes:         invoice.Notes,
		}
		if invoice.Amount != nil {
			resp.Invoices[i].Amount = dto.MoneyDTO{
				Amount:   invoice.Amount.Amount,
				Currency: invoice.Amount.Currency,
			}
		}
		if invoice.PaidAmount != nil {
			resp.Invoices[i].PaidAmount = dto.MoneyDTO{
				Amount:   invoice.PaidAmount.Amount,
				Currency: invoice.PaidAmount.Currency,
			}
		}
	}

	// Map payments
	resp.Payments = make([]*dto.PaymentResponseDTO, len(deal.Payments))
	for i, payment := range deal.Payments {
		resp.Payments[i] = &dto.PaymentResponseDTO{
			ID:              payment.ID.String(),
			PaymentDate:     payment.PaymentDate,
			PaymentMethod:   payment.PaymentMethod,
			ReferenceNumber: payment.ReferenceNumber,
			Status:          string(payment.Status),
			Notes:           payment.Notes,
			RecordedAt:      payment.RecordedAt,
			RecordedBy:      payment.RecordedBy.String(),
		}
		if payment.InvoiceID != nil {
			s := payment.InvoiceID.String()
			resp.Payments[i].InvoiceID = &s
		}
		if payment.Amount != nil {
			resp.Payments[i].Amount = dto.MoneyDTO{
				Amount:   payment.Amount.Amount,
				Currency: payment.Amount.Currency,
			}
		}
	}

	// Map fulfillments
	resp.Fulfillments = make([]*dto.FulfillmentResponseDTO, len(deal.Fulfillments))
	for i, fulfillment := range deal.Fulfillments {
		productName := ""
		for _, item := range deal.LineItems {
			if item.ID == fulfillment.LineItemID {
				productName = item.ProductName
				break
			}
		}
		resp.Fulfillments[i] = &dto.FulfillmentResponseDTO{
			ID:             fulfillment.ID.String(),
			LineItemID:     fulfillment.LineItemID.String(),
			ProductName:    productName,
			Quantity:       fulfillment.Quantity,
			FulfilledDate:  fulfillment.FulfilledDate,
			TrackingNumber: fulfillment.TrackingNumber,
			CarrierName:    fulfillment.CarrierName,
			Notes:          fulfillment.Notes,
			FulfilledBy:    fulfillment.FulfilledBy.String(),
		}
	}

	// Get customer info
	if uc.customerService != nil {
		if customer, err := uc.customerService.GetCustomer(ctx, deal.TenantID, deal.CustomerID); err == nil {
			resp.Customer = &dto.CustomerBriefDTO{
				ID:     customer.ID.String(),
				Name:   customer.Name,
				Code:   customer.Code,
				Type:   customer.Type,
				Status: customer.Status,
			}
		}
	}

	// Get owner info
	if uc.userService != nil {
		if owner, err := uc.userService.GetUser(ctx, deal.TenantID, deal.OwnerID); err == nil {
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

func (uc *dealUseCase) mapDealToBriefResponse(ctx context.Context, deal *domain.Deal) *dto.DealBriefResponse {
	customerName := ""
	if uc.customerService != nil {
		if customer, err := uc.customerService.GetCustomer(ctx, deal.TenantID, deal.CustomerID); err == nil {
			customerName = customer.Name
		}
	}

	ownerName := ""
	if uc.userService != nil {
		if owner, err := uc.userService.GetUser(ctx, deal.TenantID, deal.OwnerID); err == nil {
			ownerName = owner.FullName
		}
	}

	return &dto.DealBriefResponse{
		ID:         deal.ID.String(),
		DealNumber: deal.DealNumber,
		Name:       deal.Name,
		Status:     string(deal.Status),
		TotalAmount: dto.MoneyDTO{
			Amount:   deal.TotalAmount.Amount,
			Currency: deal.TotalAmount.Currency,
			Display:  deal.TotalAmount.Format(),
		},
		CustomerID:          deal.CustomerID.String(),
		CustomerName:        customerName,
		OwnerID:             deal.OwnerID.String(),
		OwnerName:           ownerName,
		PaymentStatus:       deal.GetPaymentStatus(),
		FulfillmentProgress: deal.FulfillmentProgress,
		ClosedDate:          deal.ClosedDate,
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
	domainFilter.HasPendingPayments = filter.HasPendingPayments
	domainFilter.FullyPaid = filter.FullyPaid
	domainFilter.FulfillmentProgress = filter.FulfillmentProgress
	domainFilter.SearchQuery = filter.SearchQuery
	domainFilter.DealNumber = filter.DealNumber

	// Parse dates
	if filter.ClosedDateAfter != nil {
		if t, err := time.Parse("2006-01-02", *filter.ClosedDateAfter); err == nil {
			domainFilter.ClosedDateAfter = &t
		}
	}
	if filter.ClosedDateBefore != nil {
		if t, err := time.Parse("2006-01-02", *filter.ClosedDateBefore); err == nil {
			domainFilter.ClosedDateBefore = &t
		}
	}

	return domainFilter
}

func (uc *dealUseCase) calculateListSummary(deals []*domain.Deal) *dto.DealSummaryDTO {
	if len(deals) == 0 {
		return nil
	}

	var totalValue, totalPaid, totalPending int64
	var fullyPaidCount, partiallyPaidCount, unpaidCount, fullyFulfilledCount int64
	currency := "USD"

	for _, deal := range deals {
		if deal.TotalAmount != nil {
			totalValue += deal.TotalAmount.Amount
			currency = deal.TotalAmount.Currency
		}
		if deal.TotalPaid != nil {
			totalPaid += deal.TotalPaid.Amount
		}

		// Calculate payment status counts
		switch deal.GetPaymentStatus() {
		case "paid":
			fullyPaidCount++
		case "partial":
			partiallyPaidCount++
		case "unpaid":
			unpaidCount++
		}

		// Check fulfillment
		if deal.FulfillmentProgress >= 100 {
			fullyFulfilledCount++
		}
	}

	totalPending = totalValue - totalPaid

	return &dto.DealSummaryDTO{
		TotalCount: int64(len(deals)),
		TotalValue: dto.MoneyDTO{
			Amount:   totalValue,
			Currency: currency,
		},
		TotalPaid: dto.MoneyDTO{
			Amount:   totalPaid,
			Currency: currency,
		},
		TotalPending: dto.MoneyDTO{
			Amount:   totalPending,
			Currency: currency,
		},
		FullyPaidCount:      fullyPaidCount,
		PartiallyPaidCount:  partiallyPaidCount,
		UnpaidCount:         unpaidCount,
		FullyFulfilledCount: fullyFulfilledCount,
	}
}

func (uc *dealUseCase) buildRevenueTrend(deals []*domain.Deal, startDate, endDate time.Time, groupBy, currency string) []*dto.RevenueTrendDTO {
	// Group deals by period
	trends := make(map[string]*dto.RevenueTrendDTO)

	for _, deal := range deals {
		if deal.ClosedDate == nil {
			continue
		}

		var periodKey string
		switch groupBy {
		case "day":
			periodKey = deal.ClosedDate.Format("2006-01-02")
		case "week":
			year, week := deal.ClosedDate.ISOWeek()
			periodKey = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, (week-1)*7).Format("2006-01-02")
		case "month":
			periodKey = deal.ClosedDate.Format("2006-01")
		case "quarter":
			quarter := (deal.ClosedDate.Month()-1)/3 + 1
			periodKey = deal.ClosedDate.Format("2006") + "-Q" + string('0'+rune(quarter))
		default:
			periodKey = deal.ClosedDate.Format("2006-01")
		}

		if _, ok := trends[periodKey]; !ok {
			trends[periodKey] = &dto.RevenueTrendDTO{
				Date: periodKey,
				Revenue: dto.MoneyDTO{
					Amount:   0,
					Currency: currency,
				},
				Collected: dto.MoneyDTO{
					Amount:   0,
					Currency: currency,
				},
			}
		}

		if deal.TotalAmount != nil {
			trends[periodKey].Revenue.Amount += deal.TotalAmount.Amount
		}
		if deal.TotalPaid != nil {
			trends[periodKey].Collected.Amount += deal.TotalPaid.Amount
		}
		trends[periodKey].DealCount++
	}

	// Convert to slice
	result := make([]*dto.RevenueTrendDTO, 0, len(trends))
	for _, trend := range trends {
		result = append(result, trend)
	}

	return result
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

func (uc *dealUseCase) indexDeal(ctx context.Context, deal *domain.Deal, customer *ports.CustomerInfo) {
	if uc.searchService == nil {
		return
	}

	customerName := ""
	if customer != nil {
		customerName = customer.Name
	}

	searchable := ports.SearchableDeal{
		ID:           deal.ID,
		TenantID:     deal.TenantID,
		DealNumber:   deal.DealNumber,
		Name:         deal.Name,
		Status:       string(deal.Status),
		TotalAmount:  deal.TotalAmount.Amount,
		Currency:     deal.TotalAmount.Currency,
		CustomerID:   deal.CustomerID,
		CustomerName: customerName,
		OwnerID:      deal.OwnerID,
		ClosedDate:   deal.ClosedDate,
		Tags:         deal.Tags,
		CreatedAt:    deal.CreatedAt,
		UpdatedAt:    deal.UpdatedAt,
	}

	uc.searchService.IndexDeal(ctx, searchable)
}

func (uc *dealUseCase) invalidateDealCache(ctx context.Context, tenantID uuid.UUID) {
	if uc.cacheService == nil {
		return
	}

	pattern := "deal:" + tenantID.String() + ":*"
	uc.cacheService.DeletePattern(ctx, pattern)
}
