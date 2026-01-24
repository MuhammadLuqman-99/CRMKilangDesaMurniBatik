// Package mapper provides mapping functions between domain entities and DTOs.
package mapper

import (
	"time"

	"github.com/google/uuid"

	"github.com/DesmondSanctworker/CRMKilangDesaMurniBatik/internal/sales/application/dto"
	"github.com/DesmondSanctworker/CRMKilangDesaMurniBatik/internal/sales/domain"
)

// ============================================================================
// Deal Mappers
// ============================================================================

// DealMapper handles mapping between Deal entities and DTOs.
type DealMapper struct{}

// NewDealMapper creates a new DealMapper instance.
func NewDealMapper() *DealMapper {
	return &DealMapper{}
}

// ToResponse maps a Deal entity to DealResponse DTO.
func (m *DealMapper) ToResponse(deal *domain.Deal) *dto.DealResponse {
	if deal == nil {
		return nil
	}

	response := &dto.DealResponse{
		ID:         deal.ID.String(),
		TenantID:   deal.TenantID.String(),
		DealNumber: deal.Code,
		Name:       deal.Name,
		Status:     string(deal.Status),
		CustomerID: deal.CustomerID.String(),
		OwnerID:    deal.OwnerID.String(),
		Subtotal: dto.MoneyDTO{
			Amount:   deal.Subtotal.Amount,
			Currency: deal.Subtotal.Currency,
			Display:  formatMoney(deal.Subtotal),
		},
		DiscountAmount: dto.MoneyDTO{
			Amount:   deal.TotalDiscount.Amount,
			Currency: deal.TotalDiscount.Currency,
			Display:  formatMoney(deal.TotalDiscount),
		},
		TaxAmount: dto.MoneyDTO{
			Amount:   deal.TotalTax.Amount,
			Currency: deal.TotalTax.Currency,
			Display:  formatMoney(deal.TotalTax),
		},
		TotalAmount: dto.MoneyDTO{
			Amount:   deal.TotalAmount.Amount,
			Currency: deal.TotalAmount.Currency,
			Display:  formatMoney(deal.TotalAmount),
		},
		TotalPaid: dto.MoneyDTO{
			Amount:   deal.PaidAmount.Amount,
			Currency: deal.PaidAmount.Currency,
			Display:  formatMoney(deal.PaidAmount),
		},
		TotalPending: dto.MoneyDTO{
			Amount:   deal.OutstandingAmount.Amount,
			Currency: deal.OutstandingAmount.Currency,
			Display:  formatMoney(deal.OutstandingAmount),
		},
		TotalInvoiced: m.calculateTotalInvoiced(deal),
		PaymentTerms:  string(deal.PaymentTerm),
		FulfillmentProgress: int(deal.FulfillmentProgress()),
		FulfillmentStatus:   m.getFulfillmentStatus(deal),
		PaymentStatus:       m.getPaymentStatus(deal),
		Tags:                deal.Tags,
		CustomFields:        deal.CustomFields,
		CreatedAt:           deal.CreatedAt,
		UpdatedAt:           deal.UpdatedAt,
		CreatedBy:           deal.CreatedBy.String(),
		Version:             deal.Version,
	}

	// Description
	if deal.Description != "" {
		response.Description = dto.StringPtr(deal.Description)
	}

	// Notes
	if deal.Notes != "" {
		response.Notes = dto.StringPtr(deal.Notes)
	}

	// Opportunity
	oppIDStr := deal.OpportunityID.String()
	response.OpportunityID = &oppIDStr

	// Customer
	response.Customer = &dto.CustomerBriefDTO{
		ID:   deal.CustomerID.String(),
		Name: deal.CustomerName,
	}

	// Owner
	response.Owner = &dto.UserBriefDTO{
		ID:   deal.OwnerID.String(),
		Name: deal.OwnerName,
	}

	// Primary contact
	if deal.PrimaryContactID != nil {
		contactIDStr := deal.PrimaryContactID.String()
		response.BillingContactID = &contactIDStr
		if deal.PrimaryContactName != "" {
			response.BillingContact = &dto.ContactBriefDTO{
				ID:       deal.PrimaryContactID.String(),
				FullName: deal.PrimaryContactName,
			}
		}
	}

	// Contract details
	if deal.ContractURL != "" {
		response.ContractNumber = dto.StringPtr(deal.ContractURL)
	}

	// Line items
	response.LineItems = m.mapLineItems(deal.LineItems, deal.Currency)
	response.LineItemCount = len(deal.LineItems)

	// Invoices
	response.Invoices = m.mapInvoices(deal.Invoices, deal.Currency)
	response.InvoiceCount = len(deal.Invoices)

	// Payments
	response.Payments = m.mapPayments(deal.Payments, deal.Currency)
	response.PaymentCount = len(deal.Payments)

	// Shipping cost (use zero if not tracked separately)
	zero, _ := domain.Zero(deal.Currency)
	response.ShippingCost = dto.MoneyDTO{
		Amount:   zero.Amount,
		Currency: zero.Currency,
		Display:  formatMoney(zero),
	}

	// Timeline dates
	response.ClosedDate = &deal.WonAt
	response.SignedDate = deal.Timeline.ContractDate
	response.StartDate = deal.Timeline.StartDate
	response.EndDate = deal.Timeline.EndDate

	// Cancellation info
	if deal.CancelledAt != nil {
		response.CancelledAt = deal.CancelledAt
		if deal.Status == domain.DealStatusCancelled {
			response.CancelNotes = dto.StringPtr(deal.Notes)
		}
	}

	return response
}

// ToBriefResponse maps a Deal entity to DealBriefResponse DTO.
func (m *DealMapper) ToBriefResponse(deal *domain.Deal) *dto.DealBriefResponse {
	if deal == nil {
		return nil
	}

	return &dto.DealBriefResponse{
		ID:         deal.ID.String(),
		DealNumber: deal.Code,
		Name:       deal.Name,
		Status:     string(deal.Status),
		TotalAmount: dto.MoneyDTO{
			Amount:   deal.TotalAmount.Amount,
			Currency: deal.TotalAmount.Currency,
			Display:  formatMoney(deal.TotalAmount),
		},
		CustomerID:          deal.CustomerID.String(),
		CustomerName:        deal.CustomerName,
		OwnerID:             deal.OwnerID.String(),
		OwnerName:           deal.OwnerName,
		PaymentStatus:       m.getPaymentStatus(deal),
		FulfillmentProgress: int(deal.FulfillmentProgress()),
		ClosedDate:          &deal.WonAt,
		CreatedAt:           deal.CreatedAt,
	}
}

// ToListResponse maps a slice of Deal entities to DealListResponse DTO.
func (m *DealMapper) ToListResponse(
	deals []*domain.Deal,
	page, pageSize int,
	totalItems int64,
	summary *dto.DealSummaryDTO,
) *dto.DealListResponse {
	briefResponses := make([]*dto.DealBriefResponse, 0, len(deals))
	for _, deal := range deals {
		briefResponses = append(briefResponses, m.ToBriefResponse(deal))
	}

	return &dto.DealListResponse{
		Deals:      briefResponses,
		Pagination: dto.NewPaginationResponse(page, pageSize, totalItems),
		Summary:    summary,
	}
}

// mapLineItems maps domain DealLineItem to DealLineItemResponseDTO.
func (m *DealMapper) mapLineItems(items []domain.DealLineItem, currency string) []*dto.DealLineItemResponseDTO {
	if len(items) == 0 {
		return nil
	}

	result := make([]*dto.DealLineItemResponseDTO, 0, len(items))
	for _, item := range items {
		lineItemDTO := &dto.DealLineItemResponseDTO{
			ID:          item.ID.String(),
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice: dto.MoneyDTO{
				Amount:   item.UnitPrice.Amount,
				Currency: item.UnitPrice.Currency,
				Display:  formatMoney(item.UnitPrice),
			},
			DiscountPercent: int(item.Discount),
			TaxRate:         int(item.Tax * 100), // Convert to basis points
			TaxAmount: dto.MoneyDTO{
				Amount:   item.TaxAmount.Amount,
				Currency: item.TaxAmount.Currency,
				Display:  formatMoney(item.TaxAmount),
			},
			TotalPrice: dto.MoneyDTO{
				Amount:   item.Total.Amount,
				Currency: item.Total.Currency,
				Display:  formatMoney(item.Total),
			},
			FulfilledQty: item.FulfilledQty,
			PendingQty:   item.RemainingQuantity(),
		}

		// Calculate discount amount
		basePrice := item.UnitPrice.Multiply(float64(item.Quantity))
		discountAmount, _ := basePrice.Subtract(item.Subtotal)
		lineItemDTO.DiscountAmount = dto.MoneyDTO{
			Amount:   discountAmount.Amount,
			Currency: discountAmount.Currency,
			Display:  formatMoney(discountAmount),
		}

		if item.ProductSKU != "" {
			lineItemDTO.ProductSKU = dto.StringPtr(item.ProductSKU)
		}

		if item.Description != "" {
			lineItemDTO.Description = dto.StringPtr(item.Description)
		}

		if item.Notes != "" {
			lineItemDTO.Notes = dto.StringPtr(item.Notes)
		}

		result = append(result, lineItemDTO)
	}

	return result
}

// mapInvoices maps domain Invoice to InvoiceResponseDTO.
func (m *DealMapper) mapInvoices(invoices []domain.Invoice, currency string) []*dto.InvoiceResponseDTO {
	if len(invoices) == 0 {
		return nil
	}

	result := make([]*dto.InvoiceResponseDTO, 0, len(invoices))
	for _, inv := range invoices {
		invoiceDTO := &dto.InvoiceResponseDTO{
			ID:            inv.ID.String(),
			InvoiceNumber: inv.InvoiceNumber,
			Amount: dto.MoneyDTO{
				Amount:   inv.Amount.Amount,
				Currency: inv.Amount.Currency,
				Display:  formatMoney(inv.Amount),
			},
			Status:     inv.Status,
			IssuedDate: inv.CreatedAt,
			DueDate:    inv.DueDate,
			PaidDate:   inv.PaidAt,
			PaidAmount: dto.MoneyDTO{
				Amount:   inv.PaidAmount.Amount,
				Currency: inv.PaidAmount.Currency,
				Display:  formatMoney(inv.PaidAmount),
			},
		}

		if inv.Notes != "" {
			invoiceDTO.Notes = dto.StringPtr(inv.Notes)
		}

		result = append(result, invoiceDTO)
	}

	return result
}

// mapPayments maps domain Payment to PaymentResponseDTO.
func (m *DealMapper) mapPayments(payments []domain.Payment, currency string) []*dto.PaymentResponseDTO {
	if len(payments) == 0 {
		return nil
	}

	result := make([]*dto.PaymentResponseDTO, 0, len(payments))
	for _, pmt := range payments {
		paymentDTO := &dto.PaymentResponseDTO{
			ID: pmt.ID.String(),
			Amount: dto.MoneyDTO{
				Amount:   pmt.Amount.Amount,
				Currency: pmt.Amount.Currency,
				Display:  formatMoney(pmt.Amount),
			},
			PaymentDate:   pmt.ReceivedAt,
			PaymentMethod: pmt.PaymentMethod,
			Status:        "completed",
			RecordedAt:    pmt.ReceivedAt,
			RecordedBy:    pmt.ReceivedBy.String(),
		}

		if pmt.InvoiceID != nil {
			invoiceIDStr := pmt.InvoiceID.String()
			paymentDTO.InvoiceID = &invoiceIDStr
		}

		if pmt.Reference != "" {
			paymentDTO.ReferenceNumber = dto.StringPtr(pmt.Reference)
		}

		if pmt.Notes != "" {
			paymentDTO.Notes = dto.StringPtr(pmt.Notes)
		}

		result = append(result, paymentDTO)
	}

	return result
}

// calculateTotalInvoiced calculates total invoiced amount.
func (m *DealMapper) calculateTotalInvoiced(deal *domain.Deal) dto.MoneyDTO {
	total, _ := domain.Zero(deal.Currency)
	for _, inv := range deal.Invoices {
		total, _ = total.Add(inv.Amount)
	}
	return dto.MoneyDTO{
		Amount:   total.Amount,
		Currency: total.Currency,
		Display:  formatMoney(total),
	}
}

// getPaymentStatus determines the payment status of a deal.
func (m *DealMapper) getPaymentStatus(deal *domain.Deal) string {
	if deal.IsFullyPaid() {
		return "paid"
	}
	if deal.PaidAmount.Amount > 0 {
		return "partial"
	}
	return "unpaid"
}

// getFulfillmentStatus determines the fulfillment status of a deal.
func (m *DealMapper) getFulfillmentStatus(deal *domain.Deal) string {
	progress := deal.FulfillmentProgress()
	if progress >= 100 {
		return "fulfilled"
	}
	if progress > 0 {
		return "partial"
	}
	return "unfulfilled"
}

// ToLineItem maps AddLineItemRequest to DealLineItem domain type.
func (m *DealMapper) ToLineItem(req *dto.AddLineItemRequest) (domain.DealLineItem, error) {
	productID, err := dto.ParseUUIDRequired(req.ProductID)
	if err != nil {
		return domain.DealLineItem{}, err
	}

	unitPrice, err := domain.NewMoney(req.UnitPrice, req.Currency)
	if err != nil {
		return domain.DealLineItem{}, err
	}

	lineItem := domain.DealLineItem{
		ID:           uuid.New(),
		ProductID:    productID,
		ProductName:  req.ProductName,
		Quantity:     req.Quantity,
		UnitPrice:    unitPrice,
		DiscountType: "percentage",
		TaxType:      "percentage",
	}

	if req.ProductSKU != nil {
		lineItem.ProductSKU = *req.ProductSKU
	}

	if req.Description != nil {
		lineItem.Description = *req.Description
	}

	if req.DiscountPercent != nil {
		lineItem.Discount = float64(*req.DiscountPercent)
	}

	if req.TaxRate != nil {
		lineItem.Tax = float64(*req.TaxRate) / 100 // Convert basis points to percentage
	}

	if req.Notes != nil {
		lineItem.Notes = *req.Notes
	}

	// Calculate totals
	lineItem.Calculate()

	return lineItem, nil
}

// ToPayment maps RecordPaymentRequest to Payment domain type.
func (m *DealMapper) ToPayment(req *dto.RecordPaymentRequest, receivedBy uuid.UUID) (domain.Payment, error) {
	amount, err := domain.NewMoney(req.Amount, req.Currency)
	if err != nil {
		return domain.Payment{}, err
	}

	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return domain.Payment{}, err
	}

	payment := domain.Payment{
		ID:            uuid.New(),
		Amount:        amount,
		PaymentMethod: req.PaymentMethod,
		ReceivedAt:    paymentDate,
		ReceivedBy:    receivedBy,
	}

	if req.InvoiceID != nil {
		invoiceID, err := dto.ParseUUIDRequired(*req.InvoiceID)
		if err != nil {
			return domain.Payment{}, err
		}
		payment.InvoiceID = &invoiceID
	}

	if req.ReferenceNumber != nil {
		payment.Reference = *req.ReferenceNumber
	}

	if req.Notes != nil {
		payment.Notes = *req.Notes
	}

	return payment, nil
}

// ToPaymentTerm maps payment term string to PaymentTerm domain type.
func (m *DealMapper) ToPaymentTerm(term string) domain.PaymentTerm {
	switch term {
	case "due_on_receipt":
		return domain.PaymentTermImmediate
	case "net_7":
		return domain.PaymentTermNet7
	case "net_15":
		return domain.PaymentTermNet15
	case "net_30":
		return domain.PaymentTermNet30
	case "net_45":
		return domain.PaymentTermNet45
	case "net_60":
		return domain.PaymentTermNet60
	case "net_90":
		return domain.PaymentTermNet90
	case "prepaid":
		return domain.PaymentTermImmediate
	case "custom":
		return domain.PaymentTermCustom
	default:
		return domain.PaymentTermNet30
	}
}

// ToTimeline maps deal request to DealTimeline domain type.
func (m *DealMapper) ToTimeline(req *dto.CreateDealRequest) (domain.DealTimeline, error) {
	timeline := domain.DealTimeline{}

	if req.SignedDate != nil {
		signedDate, err := time.Parse("2006-01-02", *req.SignedDate)
		if err != nil {
			return timeline, err
		}
		timeline.ContractDate = &signedDate
	}

	if req.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return timeline, err
		}
		timeline.StartDate = &startDate
	}

	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return timeline, err
		}
		timeline.EndDate = &endDate
	}

	return timeline, nil
}

// ParseInvoiceDueDate parses invoice due date string to time.Time.
func (m *DealMapper) ParseInvoiceDueDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// ToInvoice creates a partial Invoice mapping (for create invoice request).
func (m *DealMapper) ToInvoice(req *dto.GenerateInvoiceRequest, dealCurrency string) (string, domain.Money, time.Time, error) {
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return "", domain.Money{}, time.Time{}, err
	}

	currency := dealCurrency
	if req.Currency != nil {
		currency = *req.Currency
	}

	var amount domain.Money
	if req.Amount != nil {
		amount, err = domain.NewMoney(*req.Amount, currency)
		if err != nil {
			return "", domain.Money{}, time.Time{}, err
		}
	}

	return req.InvoiceNumber, amount, dueDate, nil
}

// ToStatisticsResponse creates a DealStatisticsResponse DTO.
func (m *DealMapper) ToStatisticsResponse(
	totalDeals int64,
	byStatus map[string]int64,
	totalRevenue, totalCollected, totalOutstanding, averageDealSize domain.Money,
	fullyPaid, partiallyPaid, unpaid, fullyFulfilled, dealsThisMonth int64,
	revenueThisMonth, collectedThisMonth domain.Money,
) *dto.DealStatisticsResponse {
	return &dto.DealStatisticsResponse{
		TotalDeals: totalDeals,
		ByStatus:   byStatus,
		TotalRevenue: dto.MoneyDTO{
			Amount:   totalRevenue.Amount,
			Currency: totalRevenue.Currency,
			Display:  formatMoney(totalRevenue),
		},
		TotalCollected: dto.MoneyDTO{
			Amount:   totalCollected.Amount,
			Currency: totalCollected.Currency,
			Display:  formatMoney(totalCollected),
		},
		TotalOutstanding: dto.MoneyDTO{
			Amount:   totalOutstanding.Amount,
			Currency: totalOutstanding.Currency,
			Display:  formatMoney(totalOutstanding),
		},
		AverageDealSize: dto.MoneyDTO{
			Amount:   averageDealSize.Amount,
			Currency: averageDealSize.Currency,
			Display:  formatMoney(averageDealSize),
		},
		FullyPaidDeals:      fullyPaid,
		PartiallyPaidDeals:  partiallyPaid,
		UnpaidDeals:         unpaid,
		FullyFulfilledDeals: fullyFulfilled,
		DealsThisMonth:      dealsThisMonth,
		RevenueThisMonth: dto.MoneyDTO{
			Amount:   revenueThisMonth.Amount,
			Currency: revenueThisMonth.Currency,
			Display:  formatMoney(revenueThisMonth),
		},
		CollectedThisMonth: dto.MoneyDTO{
			Amount:   collectedThisMonth.Amount,
			Currency: collectedThisMonth.Currency,
			Display:  formatMoney(collectedThisMonth),
		},
	}
}
