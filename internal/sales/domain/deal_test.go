package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// DealStatus Tests
// ============================================================================

func TestDealStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   DealStatus
		expected bool
	}{
		{"draft is valid", DealStatusDraft, true},
		{"pending is valid", DealStatusPending, true},
		{"active is valid", DealStatusActive, true},
		{"on_hold is valid", DealStatusOnHold, true},
		{"fulfilled is valid", DealStatusFulfilled, true},
		{"cancelled is valid", DealStatusCancelled, true},
		{"invalid status", DealStatus("invalid"), false},
		{"empty status", DealStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsValid()
			if result != tt.expected {
				t.Errorf("DealStatus(%q).IsValid() = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestDealStatus_IsClosed(t *testing.T) {
	tests := []struct {
		name     string
		status   DealStatus
		expected bool
	}{
		{"draft is not closed", DealStatusDraft, false},
		{"pending is not closed", DealStatusPending, false},
		{"active is not closed", DealStatusActive, false},
		{"on_hold is not closed", DealStatusOnHold, false},
		{"fulfilled is closed", DealStatusFulfilled, true},
		{"cancelled is closed", DealStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.IsClosed()
			if result != tt.expected {
				t.Errorf("DealStatus(%q).IsClosed() = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestValidDealStatuses(t *testing.T) {
	statuses := ValidDealStatuses()
	if len(statuses) != 6 {
		t.Errorf("ValidDealStatuses() returned %d statuses, want 6", len(statuses))
	}

	expectedStatuses := []DealStatus{
		DealStatusDraft,
		DealStatusPending,
		DealStatusActive,
		DealStatusOnHold,
		DealStatusFulfilled,
		DealStatusCancelled,
	}

	for _, expected := range expectedStatuses {
		found := false
		for _, status := range statuses {
			if status == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ValidDealStatuses() missing status %q", expected)
		}
	}
}

// ============================================================================
// PaymentTerm Tests
// ============================================================================

func TestPaymentTerm_DaysUntilDue(t *testing.T) {
	tests := []struct {
		name     string
		term     PaymentTerm
		expected int
	}{
		{"immediate", PaymentTermImmediate, 0},
		{"net_7", PaymentTermNet7, 7},
		{"net_15", PaymentTermNet15, 15},
		{"net_30", PaymentTermNet30, 30},
		{"net_45", PaymentTermNet45, 45},
		{"net_60", PaymentTermNet60, 60},
		{"net_90", PaymentTermNet90, 90},
		{"custom defaults to 30", PaymentTermCustom, 30},
		{"unknown defaults to 30", PaymentTerm("unknown"), 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.term.DaysUntilDue()
			if result != tt.expected {
				t.Errorf("PaymentTerm(%q).DaysUntilDue() = %d, want %d", tt.term, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// DealLineItem Tests
// ============================================================================

func TestDealLineItem_Calculate(t *testing.T) {
	tests := []struct {
		name              string
		quantity          int
		unitPrice         float64
		discount          float64
		discountType      string
		tax               float64
		taxType           string
		expectedSubtotal  int64
		expectedTaxAmount int64
		expectedTotal     int64
	}{
		{
			name:              "percentage discount and tax",
			quantity:          10,
			unitPrice:         100.00,
			discount:          10, // 10%
			discountType:      "percentage",
			tax:               6, // 6%
			taxType:           "percentage",
			expectedSubtotal:  90000,  // 1000 - 100 = 900 (in cents)
			expectedTaxAmount: 5400,   // 900 * 6% = 54
			expectedTotal:     95400,  // 900 + 54 = 954
		},
		{
			name:              "fixed discount and percentage tax",
			quantity:          5,
			unitPrice:         200.00,
			discount:          50.00, // $50 fixed
			discountType:      "fixed",
			tax:               10, // 10%
			taxType:           "percentage",
			expectedSubtotal:  95000, // 1000 - 50 = 950
			expectedTaxAmount: 9500,  // 950 * 10% = 95
			expectedTotal:     104500,
		},
		{
			name:              "no discount with fixed tax",
			quantity:          2,
			unitPrice:         500.00,
			discount:          0,
			discountType:      "percentage",
			tax:               25.00, // $25 fixed tax
			taxType:           "fixed",
			expectedSubtotal:  100000, // 1000
			expectedTaxAmount: 2500,   // $25 fixed
			expectedTotal:     102500,
		},
		{
			name:              "no discount no tax",
			quantity:          1,
			unitPrice:         150.00,
			discount:          0,
			discountType:      "percentage",
			tax:               0,
			taxType:           "percentage",
			expectedSubtotal:  15000,
			expectedTaxAmount: 0,
			expectedTotal:     15000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unitPrice, _ := NewMoneyFromFloat(tt.unitPrice, "MYR")
			li := &DealLineItem{
				ID:           uuid.New(),
				ProductID:    uuid.New(),
				ProductName:  "Test Product",
				Quantity:     tt.quantity,
				UnitPrice:    unitPrice,
				Discount:     tt.discount,
				DiscountType: tt.discountType,
				Tax:          tt.tax,
				TaxType:      tt.taxType,
			}

			li.Calculate()

			if li.Subtotal.Amount != tt.expectedSubtotal {
				t.Errorf("Subtotal = %d, want %d", li.Subtotal.Amount, tt.expectedSubtotal)
			}
			if li.TaxAmount.Amount != tt.expectedTaxAmount {
				t.Errorf("TaxAmount = %d, want %d", li.TaxAmount.Amount, tt.expectedTaxAmount)
			}
			if li.Total.Amount != tt.expectedTotal {
				t.Errorf("Total = %d, want %d", li.Total.Amount, tt.expectedTotal)
			}
		})
	}
}

func TestDealLineItem_IsFulfilled(t *testing.T) {
	tests := []struct {
		name         string
		quantity     int
		fulfilledQty int
		expected     bool
	}{
		{"fully fulfilled", 10, 10, true},
		{"over fulfilled", 10, 15, true},
		{"partially fulfilled", 10, 5, false},
		{"not fulfilled", 10, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			li := &DealLineItem{
				Quantity:     tt.quantity,
				FulfilledQty: tt.fulfilledQty,
			}

			result := li.IsFulfilled()
			if result != tt.expected {
				t.Errorf("IsFulfilled() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDealLineItem_RemainingQuantity(t *testing.T) {
	tests := []struct {
		name         string
		quantity     int
		fulfilledQty int
		expected     int
	}{
		{"none fulfilled", 10, 0, 10},
		{"partially fulfilled", 10, 3, 7},
		{"fully fulfilled", 10, 10, 0},
		{"over fulfilled returns negative", 10, 15, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			li := &DealLineItem{
				Quantity:     tt.quantity,
				FulfilledQty: tt.fulfilledQty,
			}

			result := li.RemainingQuantity()
			if result != tt.expected {
				t.Errorf("RemainingQuantity() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Invoice Tests
// ============================================================================

func TestInvoice_IsPaid(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"paid invoice", "paid", true},
		{"draft invoice", "draft", false},
		{"sent invoice", "sent", false},
		{"overdue invoice", "overdue", false},
		{"cancelled invoice", "cancelled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := &Invoice{Status: tt.status}
			if invoice.IsPaid() != tt.expected {
				t.Errorf("IsPaid() = %v, want %v", invoice.IsPaid(), tt.expected)
			}
		})
	}
}

func TestInvoice_IsOverdue(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		dueDate  time.Time
		expected bool
	}{
		{"paid invoice not overdue", "paid", time.Now().AddDate(0, 0, -10), false},
		{"past due invoice is overdue", "sent", time.Now().AddDate(0, 0, -1), true},
		{"future due invoice not overdue", "sent", time.Now().AddDate(0, 0, 10), false},
		{"draft past due is overdue", "draft", time.Now().AddDate(0, 0, -5), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := &Invoice{
				Status:  tt.status,
				DueDate: tt.dueDate,
			}
			if invoice.IsOverdue() != tt.expected {
				t.Errorf("IsOverdue() = %v, want %v", invoice.IsOverdue(), tt.expected)
			}
		})
	}
}

func TestInvoice_OutstandingAmount(t *testing.T) {
	tests := []struct {
		name       string
		amount     int64
		paidAmount int64
		expected   int64
	}{
		{"nothing paid", 100000, 0, 100000},
		{"partially paid", 100000, 40000, 60000},
		{"fully paid", 100000, 100000, 0},
		{"overpaid returns negative", 100000, 120000, -20000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := &Invoice{
				Amount:     Money{Amount: tt.amount, Currency: "MYR"},
				PaidAmount: Money{Amount: tt.paidAmount, Currency: "MYR"},
			}

			result := invoice.OutstandingAmount()
			if result.Amount != tt.expected {
				t.Errorf("OutstandingAmount() = %d, want %d", result.Amount, tt.expected)
			}
		})
	}
}

// ============================================================================
// Deal Creation Tests
// ============================================================================

func createTestWonOpportunity(t *testing.T) (*Opportunity, *Pipeline) {
	t.Helper()

	tenantID := uuid.New()
	ownerID := uuid.New()
	customerID := uuid.New()

	pipeline := createTestPipelineForDeal(t, tenantID)
	pipeline.EnsureClosedStages()
	wonStage := pipeline.GetWonStage()

	opp := &Opportunity{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Name:         "Test Won Opportunity",
		Description:  "Test Description",
		CustomerID:   customerID,
		CustomerName: "Test Customer",
		PipelineID:   pipeline.ID,
		StageID:      wonStage.ID,
		Status:       OpportunityStatusWon,
		Amount:       Money{Amount: 1000000, Currency: "MYR"}, // 10,000 MYR
		OwnerID:      ownerID,
		OwnerName:    "Test Owner",
		Tags:         []string{"important"},
		Products:     []OpportunityProduct{},
		Contacts:     []OpportunityContact{},
		CloseInfo: &CloseInfo{
			Reason:    "Best price",
			Notes:     "Customer was satisfied",
			ClosedBy:  ownerID,
			ClosedAt:  time.Now().UTC(),
		},
	}

	closeDate := time.Now().UTC()
	opp.ActualCloseDate = &closeDate

	// Add a product
	unitPrice, _ := NewMoneyFromFloat(500.00, "MYR")
	opp.Products = append(opp.Products, OpportunityProduct{
		ID:          uuid.New(),
		ProductID:   uuid.New(),
		ProductName: "Test Product",
		SKU:         "PROD-001",
		Quantity:    10,
		UnitPrice:   unitPrice,
		Discount:    10,
		Tax:         6,
	})

	// Add a contact
	contactID := uuid.New()
	opp.Contacts = append(opp.Contacts, OpportunityContact{
		ContactID: contactID,
		Name:      "Primary Contact",
		Email:     "contact@test.com",
		Role:      "decision_maker",
		IsPrimary: true,
	})

	return opp, pipeline
}

func createTestPipelineForDeal(t *testing.T, tenantID uuid.UUID) *Pipeline {
	t.Helper()

	pipeline := &Pipeline{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "Deal Test Pipeline",
		Stages:   make([]*Stage, 0),
	}

	return pipeline
}

func TestNewDealFromOpportunity_Success(t *testing.T) {
	opp, _ := createTestWonOpportunity(t)
	createdBy := uuid.New()

	deal, err := NewDealFromOpportunity(opp, createdBy)

	if err != nil {
		t.Fatalf("NewDealFromOpportunity() error = %v", err)
	}

	if deal == nil {
		t.Fatal("NewDealFromOpportunity() returned nil deal")
	}

	// Verify deal properties
	if deal.TenantID != opp.TenantID {
		t.Errorf("TenantID = %v, want %v", deal.TenantID, opp.TenantID)
	}
	if deal.Name != opp.Name {
		t.Errorf("Name = %s, want %s", deal.Name, opp.Name)
	}
	if deal.Status != DealStatusDraft {
		t.Errorf("Status = %s, want %s", deal.Status, DealStatusDraft)
	}
	if deal.OpportunityID != opp.ID {
		t.Errorf("OpportunityID = %v, want %v", deal.OpportunityID, opp.ID)
	}
	if deal.CustomerID != opp.CustomerID {
		t.Errorf("CustomerID = %v, want %v", deal.CustomerID, opp.CustomerID)
	}
	if deal.WonReason != opp.CloseInfo.Reason {
		t.Errorf("WonReason = %s, want %s", deal.WonReason, opp.CloseInfo.Reason)
	}
	if deal.PaymentTerm != PaymentTermNet30 {
		t.Errorf("PaymentTerm = %s, want %s", deal.PaymentTerm, PaymentTermNet30)
	}
	if deal.PaymentTermDays != 30 {
		t.Errorf("PaymentTermDays = %d, want 30", deal.PaymentTermDays)
	}
	if deal.CreatedBy != createdBy {
		t.Errorf("CreatedBy = %v, want %v", deal.CreatedBy, createdBy)
	}
	if deal.Version != 1 {
		t.Errorf("Version = %d, want 1", deal.Version)
	}

	// Verify primary contact was set
	if deal.PrimaryContactID == nil {
		t.Error("PrimaryContactID should be set")
	}
	if deal.PrimaryContactName != "Primary Contact" {
		t.Errorf("PrimaryContactName = %s, want 'Primary Contact'", deal.PrimaryContactName)
	}

	// Verify line items were created from products
	if len(deal.LineItems) != len(opp.Products) {
		t.Errorf("LineItems count = %d, want %d", len(deal.LineItems), len(opp.Products))
	}

	// Verify domain event was added
	events := deal.GetEvents()
	if len(events) == 0 {
		t.Error("Expected DealCreated event to be added")
	}
}

func TestNewDealFromOpportunity_NotWon(t *testing.T) {
	opp, _ := createTestWonOpportunity(t)
	opp.Status = OpportunityStatusOpen // Not won
	createdBy := uuid.New()

	deal, err := NewDealFromOpportunity(opp, createdBy)

	if err == nil {
		t.Error("NewDealFromOpportunity() should return error for non-won opportunity")
	}
	if deal != nil {
		t.Error("NewDealFromOpportunity() should return nil deal for non-won opportunity")
	}
}

// ============================================================================
// Deal Update Tests
// ============================================================================

func createTestDeal(t *testing.T) *Deal {
	t.Helper()

	opp, _ := createTestWonOpportunity(t)
	deal, err := NewDealFromOpportunity(opp, uuid.New())
	if err != nil {
		t.Fatalf("Failed to create test deal: %v", err)
	}
	deal.ClearEvents() // Clear creation event for cleaner tests
	return deal
}

func TestDeal_Update(t *testing.T) {
	deal := createTestDeal(t)
	originalUpdatedAt := deal.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	deal.Update("New Name", "New Description", PaymentTermNet60, 0)

	if deal.Name != "New Name" {
		t.Errorf("Name = %s, want 'New Name'", deal.Name)
	}
	if deal.Description != "New Description" {
		t.Errorf("Description = %s, want 'New Description'", deal.Description)
	}
	if deal.PaymentTerm != PaymentTermNet60 {
		t.Errorf("PaymentTerm = %s, want %s", deal.PaymentTerm, PaymentTermNet60)
	}
	if deal.PaymentTermDays != 60 {
		t.Errorf("PaymentTermDays = %d, want 60", deal.PaymentTermDays)
	}
	if !deal.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}

	// Verify event
	events := deal.GetEvents()
	if len(events) == 0 {
		t.Error("Expected DealUpdated event")
	}
}

func TestDeal_Update_CustomPaymentTerm(t *testing.T) {
	deal := createTestDeal(t)

	deal.Update("Name", "Description", PaymentTermCustom, 45)

	if deal.PaymentTerm != PaymentTermCustom {
		t.Errorf("PaymentTerm = %s, want %s", deal.PaymentTerm, PaymentTermCustom)
	}
	if deal.PaymentTermDays != 45 {
		t.Errorf("PaymentTermDays = %d, want 45", deal.PaymentTermDays)
	}
}

func TestDeal_Update_EmptyName(t *testing.T) {
	deal := createTestDeal(t)
	originalName := deal.Name

	deal.Update("", "New Description", PaymentTermNet30, 0)

	// Name should not change if empty
	if deal.Name != originalName {
		t.Errorf("Name should remain %s when empty string provided", originalName)
	}
}

// ============================================================================
// Deal Status Transition Tests
// ============================================================================

func TestDeal_Activate(t *testing.T) {
	deal := createTestDeal(t)

	err := deal.Activate()

	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if deal.Status != DealStatusActive {
		t.Errorf("Status = %s, want %s", deal.Status, DealStatusActive)
	}
	if deal.ActivatedAt == nil {
		t.Error("ActivatedAt should be set")
	}

	// Verify event
	events := deal.GetEvents()
	if len(events) == 0 {
		t.Error("Expected DealActivated event")
	}
}

func TestDeal_Activate_AlreadyActive(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusActive

	err := deal.Activate()

	if err != nil {
		t.Errorf("Activate() on already active deal should return nil, got error: %v", err)
	}
}

func TestDeal_Activate_Closed(t *testing.T) {
	tests := []struct {
		name   string
		status DealStatus
	}{
		{"fulfilled", DealStatusFulfilled},
		{"cancelled", DealStatusCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deal := createTestDeal(t)
			deal.Status = tt.status

			err := deal.Activate()

			if err != ErrDealAlreadyClosed {
				t.Errorf("Activate() error = %v, want ErrDealAlreadyClosed", err)
			}
		})
	}
}

func TestDeal_PutOnHold(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusActive

	err := deal.PutOnHold("Payment issues")

	if err != nil {
		t.Fatalf("PutOnHold() error = %v", err)
	}
	if deal.Status != DealStatusOnHold {
		t.Errorf("Status = %s, want %s", deal.Status, DealStatusOnHold)
	}
	if deal.Notes != "Payment issues" {
		t.Errorf("Notes = %s, want 'Payment issues'", deal.Notes)
	}
}

func TestDeal_PutOnHold_Closed(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusFulfilled

	err := deal.PutOnHold("Some reason")

	if err != ErrDealAlreadyClosed {
		t.Errorf("PutOnHold() error = %v, want ErrDealAlreadyClosed", err)
	}
}

func TestDeal_Resume(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusOnHold

	err := deal.Resume()

	if err != nil {
		t.Fatalf("Resume() error = %v", err)
	}
	if deal.Status != DealStatusActive {
		t.Errorf("Status = %s, want %s", deal.Status, DealStatusActive)
	}
}

func TestDeal_Resume_NotOnHold(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusActive

	err := deal.Resume()

	if err == nil {
		t.Error("Resume() should return error when not on hold")
	}
}

func TestDeal_Fulfill_Success(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusActive

	// Fulfill all line items
	for i := range deal.LineItems {
		deal.LineItems[i].FulfilledQty = deal.LineItems[i].Quantity
	}

	err := deal.Fulfill()

	if err != nil {
		t.Fatalf("Fulfill() error = %v", err)
	}
	if deal.Status != DealStatusFulfilled {
		t.Errorf("Status = %s, want %s", deal.Status, DealStatusFulfilled)
	}
	if deal.FulfilledAt == nil {
		t.Error("FulfilledAt should be set")
	}
}

func TestDeal_Fulfill_NotAllItemsFulfilled(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusActive

	// Don't fulfill line items

	err := deal.Fulfill()

	if err == nil {
		t.Error("Fulfill() should return error when line items not fulfilled")
	}
}

func TestDeal_Fulfill_AlreadyFulfilled(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusFulfilled

	err := deal.Fulfill()

	if err != ErrDealAlreadyFulfilled {
		t.Errorf("Fulfill() error = %v, want ErrDealAlreadyFulfilled", err)
	}
}

func TestDeal_Fulfill_Cancelled(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusCancelled

	err := deal.Fulfill()

	if err != ErrDealAlreadyClosed {
		t.Errorf("Fulfill() error = %v, want ErrDealAlreadyClosed", err)
	}
}

func TestDeal_Cancel_Success(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusActive

	err := deal.Cancel("Customer changed mind")

	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if deal.Status != DealStatusCancelled {
		t.Errorf("Status = %s, want %s", deal.Status, DealStatusCancelled)
	}
	if deal.CancelledAt == nil {
		t.Error("CancelledAt should be set")
	}
	if deal.Notes != "Customer changed mind" {
		t.Errorf("Notes = %s, want 'Customer changed mind'", deal.Notes)
	}
}

func TestDeal_Cancel_HasPaidInvoices(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusActive
	deal.Invoices = append(deal.Invoices, Invoice{
		ID:     uuid.New(),
		Status: "paid",
	})

	err := deal.Cancel("Some reason")

	if err != ErrDealCannotBeCancelled {
		t.Errorf("Cancel() error = %v, want ErrDealCannotBeCancelled", err)
	}
}

func TestDeal_Cancel_AlreadyClosed(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusFulfilled

	err := deal.Cancel("Some reason")

	if err != ErrDealAlreadyClosed {
		t.Errorf("Cancel() error = %v, want ErrDealAlreadyClosed", err)
	}
}

// ============================================================================
// Deal Line Item Management Tests
// ============================================================================

func TestDeal_AddLineItem(t *testing.T) {
	deal := createTestDeal(t)
	initialLineItemCount := len(deal.LineItems)

	unitPrice, _ := NewMoneyFromFloat(200.00, "MYR")
	item := DealLineItem{
		ProductID:    uuid.New(),
		ProductName:  "New Product",
		Quantity:     5,
		UnitPrice:    unitPrice,
		DiscountType: "percentage",
		TaxType:      "percentage",
	}

	err := deal.AddLineItem(item)

	if err != nil {
		t.Fatalf("AddLineItem() error = %v", err)
	}
	if len(deal.LineItems) != initialLineItemCount+1 {
		t.Errorf("LineItems count = %d, want %d", len(deal.LineItems), initialLineItemCount+1)
	}
	// Verify the item was assigned an ID
	lastItem := deal.LineItems[len(deal.LineItems)-1]
	if lastItem.ID == uuid.Nil {
		t.Error("LineItem should be assigned an ID")
	}
}

func TestDeal_AddLineItem_ClosedDeal(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusFulfilled

	unitPrice, _ := NewMoneyFromFloat(200.00, "MYR")
	item := DealLineItem{
		ProductID: uuid.New(),
		UnitPrice: unitPrice,
	}

	err := deal.AddLineItem(item)

	if err != ErrDealAlreadyClosed {
		t.Errorf("AddLineItem() error = %v, want ErrDealAlreadyClosed", err)
	}
}

func TestDeal_UpdateLineItem(t *testing.T) {
	deal := createTestDeal(t)
	if len(deal.LineItems) == 0 {
		t.Skip("No line items to update")
	}

	itemID := deal.LineItems[0].ID
	newUnitPrice, _ := NewMoneyFromFloat(300.00, "MYR")

	err := deal.UpdateLineItem(itemID, 20, newUnitPrice, 15, 8)

	if err != nil {
		t.Fatalf("UpdateLineItem() error = %v", err)
	}
	if deal.LineItems[0].Quantity != 20 {
		t.Errorf("Quantity = %d, want 20", deal.LineItems[0].Quantity)
	}
	if deal.LineItems[0].UnitPrice.Amount != newUnitPrice.Amount {
		t.Errorf("UnitPrice = %d, want %d", deal.LineItems[0].UnitPrice.Amount, newUnitPrice.Amount)
	}
	if deal.LineItems[0].Discount != 15 {
		t.Errorf("Discount = %f, want 15", deal.LineItems[0].Discount)
	}
	if deal.LineItems[0].Tax != 8 {
		t.Errorf("Tax = %f, want 8", deal.LineItems[0].Tax)
	}
}

func TestDeal_UpdateLineItem_NotFound(t *testing.T) {
	deal := createTestDeal(t)
	nonExistentID := uuid.New()
	unitPrice, _ := NewMoneyFromFloat(100.00, "MYR")

	err := deal.UpdateLineItem(nonExistentID, 10, unitPrice, 0, 0)

	if err == nil {
		t.Error("UpdateLineItem() should return error for non-existent item")
	}
}

func TestDeal_UpdateLineItem_ClosedDeal(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusCancelled

	if len(deal.LineItems) == 0 {
		t.Skip("No line items to update")
	}

	itemID := deal.LineItems[0].ID
	unitPrice, _ := NewMoneyFromFloat(100.00, "MYR")

	err := deal.UpdateLineItem(itemID, 10, unitPrice, 0, 0)

	if err != ErrDealAlreadyClosed {
		t.Errorf("UpdateLineItem() error = %v, want ErrDealAlreadyClosed", err)
	}
}

func TestDeal_RemoveLineItem(t *testing.T) {
	deal := createTestDeal(t)
	if len(deal.LineItems) == 0 {
		t.Skip("No line items to remove")
	}

	initialCount := len(deal.LineItems)
	itemID := deal.LineItems[0].ID

	err := deal.RemoveLineItem(itemID)

	if err != nil {
		t.Fatalf("RemoveLineItem() error = %v", err)
	}
	if len(deal.LineItems) != initialCount-1 {
		t.Errorf("LineItems count = %d, want %d", len(deal.LineItems), initialCount-1)
	}
}

func TestDeal_RemoveLineItem_NotFound(t *testing.T) {
	deal := createTestDeal(t)

	err := deal.RemoveLineItem(uuid.New())

	if err == nil {
		t.Error("RemoveLineItem() should return error for non-existent item")
	}
}

func TestDeal_FulfillLineItem(t *testing.T) {
	deal := createTestDeal(t)
	if len(deal.LineItems) == 0 {
		t.Skip("No line items")
	}

	itemID := deal.LineItems[0].ID
	quantity := deal.LineItems[0].Quantity / 2 // Partial fulfillment

	err := deal.FulfillLineItem(itemID, quantity)

	if err != nil {
		t.Fatalf("FulfillLineItem() error = %v", err)
	}
	if deal.LineItems[0].FulfilledQty != quantity {
		t.Errorf("FulfilledQty = %d, want %d", deal.LineItems[0].FulfilledQty, quantity)
	}
}

func TestDeal_FulfillLineItem_ExceedsQuantity(t *testing.T) {
	deal := createTestDeal(t)
	if len(deal.LineItems) == 0 {
		t.Skip("No line items")
	}

	itemID := deal.LineItems[0].ID
	quantity := deal.LineItems[0].Quantity + 10 // More than ordered

	err := deal.FulfillLineItem(itemID, quantity)

	if err == nil {
		t.Error("FulfillLineItem() should return error when exceeding quantity")
	}
}

func TestDeal_FulfillLineItem_NotFound(t *testing.T) {
	deal := createTestDeal(t)

	err := deal.FulfillLineItem(uuid.New(), 5)

	if err == nil {
		t.Error("FulfillLineItem() should return error for non-existent item")
	}
}

// ============================================================================
// Deal Invoice Tests
// ============================================================================

func TestDeal_CreateInvoice(t *testing.T) {
	deal := createTestDeal(t)
	invoiceAmount, _ := NewMoneyFromFloat(500.00, "MYR")
	dueDate := time.Now().AddDate(0, 0, 30)

	invoice, err := deal.CreateInvoice("INV-001", invoiceAmount, dueDate)

	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
	if invoice == nil {
		t.Fatal("CreateInvoice() returned nil invoice")
	}
	if invoice.InvoiceNumber != "INV-001" {
		t.Errorf("InvoiceNumber = %s, want 'INV-001'", invoice.InvoiceNumber)
	}
	if invoice.Status != "draft" {
		t.Errorf("Status = %s, want 'draft'", invoice.Status)
	}
	if len(deal.Invoices) != 1 {
		t.Errorf("Invoices count = %d, want 1", len(deal.Invoices))
	}
}

func TestDeal_CreateInvoice_ExceedsOutstanding(t *testing.T) {
	deal := createTestDeal(t)
	// Create invoice for more than outstanding
	invoiceAmount, _ := NewMoneyFromFloat(999999999.00, "MYR")
	dueDate := time.Now().AddDate(0, 0, 30)

	_, err := deal.CreateInvoice("INV-001", invoiceAmount, dueDate)

	if err == nil {
		t.Error("CreateInvoice() should return error when amount exceeds outstanding")
	}
}

func TestDeal_CreateInvoice_DuplicateNumber(t *testing.T) {
	deal := createTestDeal(t)
	invoiceAmount, _ := NewMoneyFromFloat(100.00, "MYR")
	dueDate := time.Now().AddDate(0, 0, 30)

	// Create first invoice
	deal.CreateInvoice("INV-001", invoiceAmount, dueDate)

	// Try to create duplicate
	_, err := deal.CreateInvoice("INV-001", invoiceAmount, dueDate)

	if err != ErrInvoiceAlreadyExists {
		t.Errorf("CreateInvoice() error = %v, want ErrInvoiceAlreadyExists", err)
	}
}

func TestDeal_SendInvoice(t *testing.T) {
	deal := createTestDeal(t)
	invoiceAmount, _ := NewMoneyFromFloat(100.00, "MYR")
	invoice, _ := deal.CreateInvoice("INV-001", invoiceAmount, time.Now().AddDate(0, 0, 30))

	err := deal.SendInvoice(invoice.ID)

	if err != nil {
		t.Fatalf("SendInvoice() error = %v", err)
	}

	// Find the invoice and check status
	for _, inv := range deal.Invoices {
		if inv.ID == invoice.ID {
			if inv.Status != "sent" {
				t.Errorf("Status = %s, want 'sent'", inv.Status)
			}
			if inv.SentAt == nil {
				t.Error("SentAt should be set")
			}
			break
		}
	}
}

func TestDeal_SendInvoice_NotFound(t *testing.T) {
	deal := createTestDeal(t)

	err := deal.SendInvoice(uuid.New())

	if err == nil {
		t.Error("SendInvoice() should return error for non-existent invoice")
	}
}

// ============================================================================
// Deal Payment Tests
// ============================================================================

func TestDeal_RecordPayment(t *testing.T) {
	deal := createTestDeal(t)
	paymentAmount, _ := NewMoneyFromFloat(100.00, "MYR")
	payment := Payment{
		Amount:        paymentAmount,
		PaymentMethod: "bank_transfer",
		Reference:     "PAY-001",
		ReceivedAt:    time.Now(),
		ReceivedBy:    uuid.New(),
	}

	originalPaidAmount := deal.PaidAmount.Amount
	err := deal.RecordPayment(payment)

	if err != nil {
		t.Fatalf("RecordPayment() error = %v", err)
	}
	if len(deal.Payments) != 1 {
		t.Errorf("Payments count = %d, want 1", len(deal.Payments))
	}
	if deal.PaidAmount.Amount != originalPaidAmount+paymentAmount.Amount {
		t.Errorf("PaidAmount = %d, want %d", deal.PaidAmount.Amount, originalPaidAmount+paymentAmount.Amount)
	}
}

func TestDeal_RecordPayment_ExceedsBalance(t *testing.T) {
	deal := createTestDeal(t)
	// Create payment exceeding outstanding balance
	paymentAmount, _ := NewMoneyFromFloat(999999999.00, "MYR")
	payment := Payment{
		Amount:     paymentAmount,
		ReceivedBy: uuid.New(),
	}

	err := deal.RecordPayment(payment)

	if err != ErrPaymentExceedsBalance {
		t.Errorf("RecordPayment() error = %v, want ErrPaymentExceedsBalance", err)
	}
}

func TestDeal_RecordPayment_WithInvoice(t *testing.T) {
	deal := createTestDeal(t)
	invoiceAmount, _ := NewMoneyFromFloat(100.00, "MYR")
	invoice, _ := deal.CreateInvoice("INV-001", invoiceAmount, time.Now().AddDate(0, 0, 30))

	payment := Payment{
		InvoiceID:     &invoice.ID,
		Amount:        invoiceAmount,
		PaymentMethod: "credit_card",
		ReceivedAt:    time.Now(),
		ReceivedBy:    uuid.New(),
	}

	err := deal.RecordPayment(payment)

	if err != nil {
		t.Fatalf("RecordPayment() error = %v", err)
	}

	// Find the invoice and check it's marked as paid
	for _, inv := range deal.Invoices {
		if inv.ID == invoice.ID {
			if inv.Status != "paid" {
				t.Errorf("Invoice status = %s, want 'paid'", inv.Status)
			}
			if inv.PaidAt == nil {
				t.Error("Invoice PaidAt should be set")
			}
			break
		}
	}
}

// ============================================================================
// Deal Query Methods Tests
// ============================================================================

func TestDeal_IsFullyPaid(t *testing.T) {
	tests := []struct {
		name              string
		outstandingAmount int64
		expected          bool
	}{
		{"outstanding positive", 10000, false},
		{"outstanding zero", 0, true},
		{"outstanding negative", -100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deal := &Deal{
				OutstandingAmount: Money{Amount: tt.outstandingAmount, Currency: "MYR"},
			}

			if deal.IsFullyPaid() != tt.expected {
				t.Errorf("IsFullyPaid() = %v, want %v", deal.IsFullyPaid(), tt.expected)
			}
		})
	}
}

func TestDeal_GetOverdueInvoices(t *testing.T) {
	deal := createTestDeal(t)

	// Add some invoices
	deal.Invoices = []Invoice{
		{ID: uuid.New(), InvoiceNumber: "INV-001", Status: "sent", DueDate: time.Now().AddDate(0, 0, -10)}, // Overdue
		{ID: uuid.New(), InvoiceNumber: "INV-002", Status: "paid", DueDate: time.Now().AddDate(0, 0, -5)}, // Not overdue (paid)
		{ID: uuid.New(), InvoiceNumber: "INV-003", Status: "sent", DueDate: time.Now().AddDate(0, 0, 10)}, // Not overdue
		{ID: uuid.New(), InvoiceNumber: "INV-004", Status: "draft", DueDate: time.Now().AddDate(0, 0, -3)}, // Overdue
	}

	overdue := deal.GetOverdueInvoices()

	if len(overdue) != 2 {
		t.Errorf("GetOverdueInvoices() count = %d, want 2", len(overdue))
	}
}

func TestDeal_FulfillmentProgress(t *testing.T) {
	tests := []struct {
		name           string
		lineItems      []DealLineItem
		expectedResult float64
	}{
		{
			name:           "no line items",
			lineItems:      []DealLineItem{},
			expectedResult: 0,
		},
		{
			name: "zero quantity",
			lineItems: []DealLineItem{
				{Quantity: 0, FulfilledQty: 0},
			},
			expectedResult: 0,
		},
		{
			name: "50% fulfilled",
			lineItems: []DealLineItem{
				{Quantity: 10, FulfilledQty: 5},
			},
			expectedResult: 50,
		},
		{
			name: "100% fulfilled",
			lineItems: []DealLineItem{
				{Quantity: 10, FulfilledQty: 10},
			},
			expectedResult: 100,
		},
		{
			name: "multiple items 75% fulfilled",
			lineItems: []DealLineItem{
				{Quantity: 10, FulfilledQty: 10},  // 100%
				{Quantity: 10, FulfilledQty: 5},   // 50%
			},
			expectedResult: 75, // 15/20 = 75%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deal := &Deal{LineItems: tt.lineItems}

			result := deal.FulfillmentProgress()
			if result != tt.expectedResult {
				t.Errorf("FulfillmentProgress() = %f, want %f", result, tt.expectedResult)
			}
		})
	}
}

func TestDeal_PaymentProgress(t *testing.T) {
	tests := []struct {
		name           string
		totalAmount    int64
		paidAmount     int64
		expectedResult float64
	}{
		{"zero total", 0, 0, 0},
		{"nothing paid", 100000, 0, 0},
		{"50% paid", 100000, 50000, 50},
		{"fully paid", 100000, 100000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deal := &Deal{
				TotalAmount: Money{Amount: tt.totalAmount, Currency: "MYR"},
				PaidAmount:  Money{Amount: tt.paidAmount, Currency: "MYR"},
			}

			result := deal.PaymentProgress()
			if result != tt.expectedResult {
				t.Errorf("PaymentProgress() = %f, want %f", result, tt.expectedResult)
			}
		})
	}
}

// ============================================================================
// Deal Metadata Tests
// ============================================================================

func TestDeal_SetTimeline(t *testing.T) {
	deal := createTestDeal(t)
	now := time.Now().UTC()
	timeline := DealTimeline{
		QuoteDate:    &now,
		ContractDate: &now,
		StartDate:    &now,
	}

	deal.SetTimeline(timeline)

	if deal.Timeline.QuoteDate == nil {
		t.Error("QuoteDate should be set")
	}
	if deal.Timeline.ContractDate == nil {
		t.Error("ContractDate should be set")
	}
	if deal.Timeline.StartDate == nil {
		t.Error("StartDate should be set")
	}
}

func TestDeal_SetContractURL(t *testing.T) {
	deal := createTestDeal(t)

	deal.SetContractURL("https://example.com/contract.pdf")

	if deal.ContractURL != "https://example.com/contract.pdf" {
		t.Errorf("ContractURL = %s, want 'https://example.com/contract.pdf'", deal.ContractURL)
	}
}

func TestDeal_AddTag(t *testing.T) {
	deal := createTestDeal(t)
	initialTagCount := len(deal.Tags)

	deal.AddTag("new-tag")

	if len(deal.Tags) != initialTagCount+1 {
		t.Errorf("Tags count = %d, want %d", len(deal.Tags), initialTagCount+1)
	}
}

func TestDeal_AddTag_Duplicate(t *testing.T) {
	deal := createTestDeal(t)
	deal.AddTag("existing-tag")
	tagCount := len(deal.Tags)

	deal.AddTag("existing-tag")

	if len(deal.Tags) != tagCount {
		t.Errorf("Tags count = %d, should not change for duplicate", len(deal.Tags))
	}
}

func TestDeal_RemoveTag(t *testing.T) {
	deal := createTestDeal(t)
	deal.AddTag("tag-to-remove")
	tagCount := len(deal.Tags)

	deal.RemoveTag("tag-to-remove")

	if len(deal.Tags) != tagCount-1 {
		t.Errorf("Tags count = %d, want %d", len(deal.Tags), tagCount-1)
	}
}

func TestDeal_RemoveTag_NotFound(t *testing.T) {
	deal := createTestDeal(t)
	tagCount := len(deal.Tags)

	deal.RemoveTag("non-existent")

	if len(deal.Tags) != tagCount {
		t.Errorf("Tags count should not change when removing non-existent tag")
	}
}

func TestDeal_SetCustomField(t *testing.T) {
	deal := createTestDeal(t)

	deal.SetCustomField("custom_key", "custom_value")

	if deal.CustomFields == nil {
		t.Fatal("CustomFields should be initialized")
	}
	if deal.CustomFields["custom_key"] != "custom_value" {
		t.Errorf("CustomField = %v, want 'custom_value'", deal.CustomFields["custom_key"])
	}
}

func TestDeal_SetCustomField_InitializesMap(t *testing.T) {
	deal := &Deal{CustomFields: nil}

	deal.SetCustomField("key", "value")

	if deal.CustomFields == nil {
		t.Fatal("CustomFields should be initialized")
	}
}

// ============================================================================
// Deal Delete/Restore Tests
// ============================================================================

func TestDeal_Delete_Draft(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusDraft

	err := deal.Delete()

	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if deal.DeletedAt == nil {
		t.Error("DeletedAt should be set")
	}
}

func TestDeal_Delete_Active(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusActive

	err := deal.Delete()

	if err == nil {
		t.Error("Delete() should return error for active deal")
	}
}

func TestDeal_Delete_Fulfilled(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusFulfilled

	err := deal.Delete()

	if err == nil {
		t.Error("Delete() should return error for fulfilled deal")
	}
}

func TestDeal_Restore(t *testing.T) {
	deal := createTestDeal(t)
	deal.Status = DealStatusDraft
	deal.Delete()

	deal.Restore()

	if deal.DeletedAt != nil {
		t.Error("DeletedAt should be nil after restore")
	}
}

func TestDeal_IsDeleted(t *testing.T) {
	deal := createTestDeal(t)

	if deal.IsDeleted() {
		t.Error("IsDeleted() should return false for non-deleted deal")
	}

	deal.Status = DealStatusDraft
	deal.Delete()

	if !deal.IsDeleted() {
		t.Error("IsDeleted() should return true for deleted deal")
	}
}

// ============================================================================
// Deal Event Tests
// ============================================================================

func TestDeal_AddEvent(t *testing.T) {
	deal := createTestDeal(t)
	deal.ClearEvents()

	event := &BaseEvent{
		ID:   uuid.New(),
		Type: "test_event",
	}

	deal.AddEvent(event)

	events := deal.GetEvents()
	if len(events) != 1 {
		t.Errorf("Events count = %d, want 1", len(events))
	}
	if events[0].EventType() != "test_event" {
		t.Errorf("Event type = %s, want 'test_event'", events[0].EventType())
	}
}

func TestDeal_ClearEvents(t *testing.T) {
	deal := createTestDeal(t)
	deal.AddEvent(&BaseEvent{ID: uuid.New(), Type: "test"})

	deal.ClearEvents()

	events := deal.GetEvents()
	if len(events) != 0 {
		t.Errorf("Events count = %d, want 0", len(events))
	}
}
