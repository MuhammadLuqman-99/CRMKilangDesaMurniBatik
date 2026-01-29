// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOpportunityStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   OpportunityStatus
		expected bool
	}{
		{OpportunityStatusOpen, true},
		{OpportunityStatusWon, true},
		{OpportunityStatusLost, true},
		{OpportunityStatus("invalid"), false},
		{OpportunityStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsValid() != tt.expected {
				t.Errorf("OpportunityStatus.IsValid() = %v, want %v", tt.status.IsValid(), tt.expected)
			}
		})
	}
}

func TestOpportunityStatus_IsClosed(t *testing.T) {
	tests := []struct {
		status   OpportunityStatus
		expected bool
	}{
		{OpportunityStatusOpen, false},
		{OpportunityStatusWon, true},
		{OpportunityStatusLost, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if tt.status.IsClosed() != tt.expected {
				t.Errorf("OpportunityStatus.IsClosed() = %v, want %v", tt.status.IsClosed(), tt.expected)
			}
		})
	}
}

func TestValidOpportunityStatuses(t *testing.T) {
	statuses := ValidOpportunityStatuses()
	if len(statuses) != 3 {
		t.Errorf("ValidOpportunityStatuses() len = %v, want 3", len(statuses))
	}
}

func TestOpportunityProduct_CalculateTotalPrice(t *testing.T) {
	unitPrice, _ := NewMoneyFromFloat(100, "USD")
	product := OpportunityProduct{
		ID:          uuid.New(),
		ProductID:   uuid.New(),
		ProductName: "Test Product",
		Quantity:    10,
		UnitPrice:   unitPrice,
		Discount:    10,  // 10%
		Tax:         8,   // 8%
	}

	product.CalculateTotalPrice()

	// Base: 100 * 10 = 1000
	// Discount: 1000 * 10% = 100
	// After discount: 900
	// Tax: 900 * 8% = 72
	// Total: 900 + 72 = 972
	expectedTotal := int64(97200) // cents
	if product.TotalPrice.Amount != expectedTotal {
		t.Errorf("OpportunityProduct.CalculateTotalPrice() = %v, want %v", product.TotalPrice.Amount, expectedTotal)
	}
}

func createTestOpportunityPipeline(t *testing.T) *Pipeline {
	tenantID := uuid.New()
	createdBy := uuid.New()
	pipeline, err := NewPipeline(tenantID, "Test Pipeline", "USD", createdBy)
	if err != nil {
		t.Fatalf("createTestOpportunityPipeline() failed: %v", err)
	}
	pipeline.EnsureClosedStages()
	return pipeline
}

func TestNewOpportunity(t *testing.T) {
	pipeline := createTestOpportunityPipeline(t)
	tenantID := uuid.New()
	customerID := uuid.New()
	ownerID := uuid.New()
	createdBy := uuid.New()
	amount, _ := NewMoneyFromFloat(50000, "USD")

	tests := []struct {
		name         string
		tenantID     uuid.UUID
		oppName      string
		pipeline     *Pipeline
		customerID   uuid.UUID
		customerName string
		amount       Money
		ownerID      uuid.UUID
		ownerName    string
		createdBy    uuid.UUID
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid opportunity",
			tenantID:     tenantID,
			oppName:      "Test Opportunity",
			pipeline:     pipeline,
			customerID:   customerID,
			customerName: "Test Customer",
			amount:       amount,
			ownerID:      ownerID,
			ownerName:    "John Owner",
			createdBy:    createdBy,
			wantErr:      false,
		},
		{
			name:         "missing name",
			tenantID:     tenantID,
			oppName:      "",
			pipeline:     pipeline,
			customerID:   customerID,
			customerName: "Test Customer",
			amount:       amount,
			ownerID:      ownerID,
			ownerName:    "John Owner",
			createdBy:    createdBy,
			wantErr:      true,
			errMsg:       "opportunity name is required",
		},
		{
			name:         "nil pipeline",
			tenantID:     tenantID,
			oppName:      "Test",
			pipeline:     nil,
			customerID:   customerID,
			customerName: "Test Customer",
			amount:       amount,
			ownerID:      ownerID,
			ownerName:    "John Owner",
			createdBy:    createdBy,
			wantErr:      true,
			errMsg:       "pipeline is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opp, err := NewOpportunity(tt.tenantID, tt.oppName, tt.pipeline, tt.customerID, tt.customerName, tt.amount, tt.ownerID, tt.ownerName, tt.createdBy)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewOpportunity() expected error, got nil")
				}
				if err != nil && err.Error() != tt.errMsg {
					t.Errorf("NewOpportunity() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("NewOpportunity() unexpected error = %v", err)
				}
				if opp == nil {
					t.Fatal("NewOpportunity() returned nil opportunity")
				}
				if opp.ID == uuid.Nil {
					t.Error("NewOpportunity() should generate ID")
				}
				if opp.Status != OpportunityStatusOpen {
					t.Errorf("NewOpportunity() Status = %v, want %v", opp.Status, OpportunityStatusOpen)
				}
				if opp.PipelineID != pipeline.ID {
					t.Errorf("NewOpportunity() PipelineID = %v, want %v", opp.PipelineID, pipeline.ID)
				}
				if opp.Version != 1 {
					t.Errorf("NewOpportunity() Version = %v, want 1", opp.Version)
				}
				// Should have stage history
				if len(opp.StageHistory) != 1 {
					t.Errorf("NewOpportunity() should have initial stage history, got %d", len(opp.StageHistory))
				}
				// Check domain event was added
				events := opp.GetEvents()
				if len(events) != 1 {
					t.Errorf("NewOpportunity() should add OpportunityCreatedEvent, got %d events", len(events))
				}
			}
		})
	}
}

func createTestOpportunity(t *testing.T) (*Opportunity, *Pipeline) {
	pipeline := createTestOpportunityPipeline(t)
	tenantID := uuid.New()
	customerID := uuid.New()
	ownerID := uuid.New()
	createdBy := uuid.New()
	amount, _ := NewMoneyFromFloat(50000, "USD")

	opp, err := NewOpportunity(tenantID, "Test Opportunity", pipeline, customerID, "Test Customer", amount, ownerID, "John Owner", createdBy)
	if err != nil {
		t.Fatalf("createTestOpportunity() failed: %v", err)
	}
	opp.ClearEvents()
	return opp, pipeline
}

func TestOpportunity_Update(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	opp.Update("Updated Name", "New description", OpportunityPriorityHigh)

	if opp.Name != "Updated Name" {
		t.Errorf("Opportunity.Update() Name = %v", opp.Name)
	}
	if opp.Description != "New description" {
		t.Errorf("Opportunity.Update() Description = %v", opp.Description)
	}
	if opp.Priority != OpportunityPriorityHigh {
		t.Errorf("Opportunity.Update() Priority = %v", opp.Priority)
	}

	events := opp.GetEvents()
	if len(events) != 1 {
		t.Errorf("Opportunity.Update() should add event, got %d events", len(events))
	}
}

func TestOpportunity_Update_EmptyName(t *testing.T) {
	opp, _ := createTestOpportunity(t)
	originalName := opp.Name

	opp.Update("", "Description", OpportunityPriorityLow)

	if opp.Name != originalName {
		t.Error("Opportunity.Update() with empty name should keep original")
	}
}

func TestOpportunity_SetAmount(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	newAmount, _ := NewMoneyFromFloat(75000, "USD")
	opp.SetAmount(newAmount)

	if opp.Amount.Amount != newAmount.Amount {
		t.Errorf("Opportunity.SetAmount() Amount = %v, want %v", opp.Amount.Amount, newAmount.Amount)
	}
	// Weighted amount should be recalculated
	if opp.WeightedAmount.Amount == 0 {
		t.Error("Opportunity.SetAmount() should recalculate weighted amount")
	}

	events := opp.GetEvents()
	if len(events) != 1 {
		t.Errorf("Opportunity.SetAmount() should add event, got %d events", len(events))
	}
}

func TestOpportunity_MoveToStage(t *testing.T) {
	opp, pipeline := createTestOpportunity(t)
	opp.ClearEvents()
	originalStageID := opp.StageID

	// Get a different stage
	stages := pipeline.GetActiveStages()
	var newStage *Stage
	for _, s := range stages {
		if s.ID != originalStageID && !s.Type.IsClosedType() {
			newStage = s
			break
		}
	}

	if newStage == nil {
		t.Skip("No alternative stage found")
	}

	err := opp.MoveToStage(newStage, uuid.New(), "Moving to next stage")
	if err != nil {
		t.Fatalf("Opportunity.MoveToStage() unexpected error = %v", err)
	}

	if opp.StageID != newStage.ID {
		t.Errorf("Opportunity.MoveToStage() StageID = %v, want %v", opp.StageID, newStage.ID)
	}
	if opp.StageName != newStage.Name {
		t.Errorf("Opportunity.MoveToStage() StageName = %v, want %v", opp.StageName, newStage.Name)
	}
	if opp.Probability != newStage.Probability {
		t.Errorf("Opportunity.MoveToStage() Probability = %v, want %v", opp.Probability, newStage.Probability)
	}

	// Stage history should have new entry
	if len(opp.StageHistory) != 2 {
		t.Errorf("Opportunity.MoveToStage() StageHistory len = %v, want 2", len(opp.StageHistory))
	}

	events := opp.GetEvents()
	if len(events) != 1 {
		t.Errorf("Opportunity.MoveToStage() should add event, got %d events", len(events))
	}
}

func TestOpportunity_MoveToStage_Closed(t *testing.T) {
	opp, pipeline := createTestOpportunity(t)
	opp.Status = OpportunityStatusWon

	newStage, _ := NewStage(pipeline.ID, "New Stage", StageTypeOpen, 10, 50)

	err := opp.MoveToStage(newStage, uuid.New(), "Test")
	if err != ErrOpportunityAlreadyClosed {
		t.Errorf("Opportunity.MoveToStage() on closed opportunity should return ErrOpportunityAlreadyClosed, got %v", err)
	}
}

func TestOpportunity_Win(t *testing.T) {
	opp, pipeline := createTestOpportunity(t)
	opp.ClearEvents()
	wonStage := pipeline.GetWonStage()
	wonBy := uuid.New()

	err := opp.Win(wonStage, "Customer chose us", "Additional notes", wonBy)
	if err != nil {
		t.Fatalf("Opportunity.Win() unexpected error = %v", err)
	}

	if opp.Status != OpportunityStatusWon {
		t.Errorf("Opportunity.Win() Status = %v, want %v", opp.Status, OpportunityStatusWon)
	}
	if opp.Probability != 100 {
		t.Errorf("Opportunity.Win() Probability = %v, want 100", opp.Probability)
	}
	if opp.ActualCloseDate == nil {
		t.Error("Opportunity.Win() should set ActualCloseDate")
	}
	if opp.CloseInfo == nil {
		t.Fatal("Opportunity.Win() should set CloseInfo")
	}
	if opp.CloseInfo.Reason != "Customer chose us" {
		t.Errorf("Opportunity.Win() CloseInfo.Reason = %v", opp.CloseInfo.Reason)
	}

	events := opp.GetEvents()
	if len(events) < 1 {
		t.Errorf("Opportunity.Win() should add events, got %d events", len(events))
	}
}

func TestOpportunity_Win_AlreadyClosed(t *testing.T) {
	opp, pipeline := createTestOpportunity(t)
	opp.Status = OpportunityStatusLost
	wonStage := pipeline.GetWonStage()

	err := opp.Win(wonStage, "Reason", "Notes", uuid.New())
	if err != ErrOpportunityAlreadyClosed {
		t.Errorf("Opportunity.Win() on closed opportunity should return ErrOpportunityAlreadyClosed, got %v", err)
	}
}

func TestOpportunity_Lose(t *testing.T) {
	opp, pipeline := createTestOpportunity(t)
	opp.ClearEvents()
	lostStage := pipeline.GetLostStage()
	lostBy := uuid.New()
	competitorID := uuid.New()

	err := opp.Lose(lostStage, "Price too high", "Notes", &competitorID, "Competitor Inc", lostBy)
	if err != nil {
		t.Fatalf("Opportunity.Lose() unexpected error = %v", err)
	}

	if opp.Status != OpportunityStatusLost {
		t.Errorf("Opportunity.Lose() Status = %v, want %v", opp.Status, OpportunityStatusLost)
	}
	if opp.Probability != 0 {
		t.Errorf("Opportunity.Lose() Probability = %v, want 0", opp.Probability)
	}
	if opp.ActualCloseDate == nil {
		t.Error("Opportunity.Lose() should set ActualCloseDate")
	}
	if opp.CloseInfo == nil {
		t.Fatal("Opportunity.Lose() should set CloseInfo")
	}
	if opp.CloseInfo.CompetitorName != "Competitor Inc" {
		t.Errorf("Opportunity.Lose() CloseInfo.CompetitorName = %v", opp.CloseInfo.CompetitorName)
	}
}

func TestOpportunity_Reopen(t *testing.T) {
	opp, pipeline := createTestOpportunity(t)
	wonStage := pipeline.GetWonStage()
	opp.Win(wonStage, "Reason", "Notes", uuid.New())
	opp.ClearEvents()

	reopenedBy := uuid.New()
	err := opp.Reopen(pipeline, reopenedBy, "Customer wants to continue")
	if err != nil {
		t.Fatalf("Opportunity.Reopen() unexpected error = %v", err)
	}

	if opp.Status != OpportunityStatusOpen {
		t.Errorf("Opportunity.Reopen() Status = %v, want %v", opp.Status, OpportunityStatusOpen)
	}
	if opp.ActualCloseDate != nil {
		t.Error("Opportunity.Reopen() should clear ActualCloseDate")
	}
	if opp.CloseInfo != nil {
		t.Error("Opportunity.Reopen() should clear CloseInfo")
	}
}

func TestOpportunity_Reopen_NotClosed(t *testing.T) {
	opp, pipeline := createTestOpportunity(t)

	err := opp.Reopen(pipeline, uuid.New(), "Notes")
	if err != ErrOpportunityNotClosed {
		t.Errorf("Opportunity.Reopen() on open opportunity should return ErrOpportunityNotClosed, got %v", err)
	}
}

func TestOpportunity_Reopen_TooLate(t *testing.T) {
	opp, pipeline := createTestOpportunity(t)
	wonStage := pipeline.GetWonStage()
	opp.Win(wonStage, "Reason", "Notes", uuid.New())

	// Set close date to more than 30 days ago
	closeDate := time.Now().Add(-35 * 24 * time.Hour)
	opp.ActualCloseDate = &closeDate

	err := opp.Reopen(pipeline, uuid.New(), "Notes")
	if err != ErrCannotReopenAfterDays {
		t.Errorf("Opportunity.Reopen() after 30 days should return ErrCannotReopenAfterDays, got %v", err)
	}
}

func TestOpportunity_Contacts(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	contact1 := OpportunityContact{
		ContactID:  uuid.New(),
		CustomerID: opp.CustomerID,
		Name:       "John Doe",
		Email:      "john@example.com",
		Role:       "decision_maker",
		IsPrimary:  true,
	}

	contact2 := OpportunityContact{
		ContactID:  uuid.New(),
		CustomerID: opp.CustomerID,
		Name:       "Jane Smith",
		Email:      "jane@example.com",
		Role:       "influencer",
		IsPrimary:  false,
	}

	// Add contacts
	opp.AddContact(contact1)
	opp.AddContact(contact2)

	if len(opp.Contacts) != 2 {
		t.Errorf("Opportunity.AddContact() len = %v, want 2", len(opp.Contacts))
	}

	// Get primary contact
	primary := opp.GetPrimaryContact()
	if primary == nil {
		t.Fatal("Opportunity.GetPrimaryContact() should return primary contact")
	}
	if primary.Name != "John Doe" {
		t.Errorf("Opportunity.GetPrimaryContact() Name = %v, want John Doe", primary.Name)
	}

	// Set different primary
	opp.SetPrimaryContact(contact2.ContactID)
	primary = opp.GetPrimaryContact()
	if primary.Name != "Jane Smith" {
		t.Errorf("Opportunity.SetPrimaryContact() primary should be Jane Smith")
	}

	// Remove contact
	opp.RemoveContact(contact1.ContactID)
	if len(opp.Contacts) != 1 {
		t.Errorf("Opportunity.RemoveContact() len = %v, want 1", len(opp.Contacts))
	}
}

func TestOpportunity_GetPrimaryContact_Empty(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	primary := opp.GetPrimaryContact()
	if primary != nil {
		t.Error("Opportunity.GetPrimaryContact() should return nil for empty contacts")
	}
}

func TestOpportunity_Products(t *testing.T) {
	opp, _ := createTestOpportunity(t)
	unitPrice, _ := NewMoneyFromFloat(100, "USD")

	product := OpportunityProduct{
		ProductID:   uuid.New(),
		ProductName: "Product A",
		Quantity:    5,
		UnitPrice:   unitPrice,
		Discount:    10,
		Tax:         8,
	}

	// Add product
	opp.AddProduct(product)
	if len(opp.Products) != 1 {
		t.Errorf("Opportunity.AddProduct() len = %v, want 1", len(opp.Products))
	}
	if opp.Products[0].ID == uuid.Nil {
		t.Error("Opportunity.AddProduct() should generate product ID")
	}

	// Update product
	productID := opp.Products[0].ID
	newPrice, _ := NewMoneyFromFloat(150, "USD")
	err := opp.UpdateProduct(productID, 10, newPrice, 15, 10)
	if err != nil {
		t.Fatalf("Opportunity.UpdateProduct() unexpected error = %v", err)
	}
	if opp.Products[0].Quantity != 10 {
		t.Errorf("Opportunity.UpdateProduct() Quantity = %v, want 10", opp.Products[0].Quantity)
	}

	// Update non-existent product
	err = opp.UpdateProduct(uuid.New(), 1, newPrice, 0, 0)
	if err == nil {
		t.Error("Opportunity.UpdateProduct() should return error for non-existent product")
	}

	// Remove product
	opp.RemoveProduct(productID)
	if len(opp.Products) != 0 {
		t.Errorf("Opportunity.RemoveProduct() len = %v, want 0", len(opp.Products))
	}
}

func TestOpportunity_AssignOwner(t *testing.T) {
	opp, _ := createTestOpportunity(t)
	opp.ClearEvents()

	newOwnerID := uuid.New()
	opp.AssignOwner(newOwnerID, "New Owner")

	if opp.OwnerID != newOwnerID {
		t.Errorf("Opportunity.AssignOwner() OwnerID = %v, want %v", opp.OwnerID, newOwnerID)
	}
	if opp.OwnerName != "New Owner" {
		t.Errorf("Opportunity.AssignOwner() OwnerName = %v", opp.OwnerName)
	}

	events := opp.GetEvents()
	if len(events) != 1 {
		t.Errorf("Opportunity.AssignOwner() should add event, got %d events", len(events))
	}
}

func TestOpportunity_ExpectedCloseDate(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	closeDate := time.Now().Add(30 * 24 * time.Hour)
	opp.SetExpectedCloseDate(closeDate)

	if opp.ExpectedCloseDate == nil {
		t.Fatal("Opportunity.SetExpectedCloseDate() should set date")
	}
	if !opp.ExpectedCloseDate.Equal(closeDate) {
		t.Errorf("Opportunity.SetExpectedCloseDate() date = %v, want %v", opp.ExpectedCloseDate, closeDate)
	}
}

func TestOpportunity_Activity(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	activityTime := time.Now()
	opp.RecordActivity(activityTime)

	if opp.LastActivityAt == nil {
		t.Fatal("Opportunity.RecordActivity() should set LastActivityAt")
	}
	if opp.ActivityCount != 1 {
		t.Errorf("Opportunity.RecordActivity() ActivityCount = %v, want 1", opp.ActivityCount)
	}

	nextActivity := time.Now().Add(7 * 24 * time.Hour)
	opp.SetNextActivity(nextActivity)
	if opp.NextActivityAt == nil {
		t.Error("Opportunity.SetNextActivity() should set NextActivityAt")
	}
}

func TestOpportunity_Tags(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	opp.AddTag("hot")
	opp.AddTag("enterprise")
	opp.AddTag("hot") // Duplicate

	if len(opp.Tags) != 2 {
		t.Errorf("Opportunity.AddTag() len = %v, want 2", len(opp.Tags))
	}

	opp.RemoveTag("hot")
	if len(opp.Tags) != 1 {
		t.Errorf("Opportunity.RemoveTag() len = %v, want 1", len(opp.Tags))
	}
}

func TestOpportunity_CustomFields(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	opp.SetCustomField("industry", "Technology")
	opp.SetCustomField("deal_type", "New Business")

	if opp.CustomFields["industry"] != "Technology" {
		t.Errorf("Opportunity.SetCustomField() industry = %v", opp.CustomFields["industry"])
	}
}

func TestOpportunity_Delete(t *testing.T) {
	opp, _ := createTestOpportunity(t)
	opp.ClearEvents()

	opp.Delete()

	if !opp.IsDeleted() {
		t.Error("Opportunity.Delete() should mark as deleted")
	}
	if opp.DeletedAt == nil {
		t.Error("Opportunity.Delete() should set DeletedAt")
	}

	events := opp.GetEvents()
	if len(events) != 1 {
		t.Errorf("Opportunity.Delete() should add event, got %d events", len(events))
	}
}

func TestOpportunity_Restore(t *testing.T) {
	opp, _ := createTestOpportunity(t)
	opp.Delete()

	opp.Restore()

	if opp.IsDeleted() {
		t.Error("Opportunity.Restore() should restore opportunity")
	}
	if opp.DeletedAt != nil {
		t.Error("Opportunity.Restore() should clear DeletedAt")
	}
}

func TestOpportunity_StatusChecks(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	// IsOpen
	if !opp.IsOpen() {
		t.Error("Opportunity.IsOpen() should return true for open opportunity")
	}

	// IsWon
	if opp.IsWon() {
		t.Error("Opportunity.IsWon() should return false for open opportunity")
	}
	opp.Status = OpportunityStatusWon
	if !opp.IsWon() {
		t.Error("Opportunity.IsWon() should return true for won opportunity")
	}

	// IsLost
	opp.Status = OpportunityStatusLost
	if !opp.IsLost() {
		t.Error("Opportunity.IsLost() should return true for lost opportunity")
	}
}

func TestOpportunity_DaysCalculations(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	// Set created at to 10 days ago
	opp.CreatedAt = time.Now().Add(-10 * 24 * time.Hour)
	opp.StageEnteredAt = time.Now().Add(-3 * 24 * time.Hour)

	days := opp.DaysInPipeline()
	if days < 9 || days > 11 {
		t.Errorf("Opportunity.DaysInPipeline() = %v, expected ~10", days)
	}

	days = opp.DaysInCurrentStage()
	if days < 2 || days > 4 {
		t.Errorf("Opportunity.DaysInCurrentStage() = %v, expected ~3", days)
	}

	// Expected close date
	closeDate := time.Now().Add(15 * 24 * time.Hour)
	opp.SetExpectedCloseDate(closeDate)
	days = opp.DaysUntilExpectedClose()
	if days < 14 || days > 16 {
		t.Errorf("Opportunity.DaysUntilExpectedClose() = %v, expected ~15", days)
	}
}

func TestOpportunity_IsOverdue(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	// No expected close date
	if opp.IsOverdue() {
		t.Error("Opportunity.IsOverdue() should return false without expected close date")
	}

	// Future close date
	futureDate := time.Now().Add(10 * 24 * time.Hour)
	opp.SetExpectedCloseDate(futureDate)
	if opp.IsOverdue() {
		t.Error("Opportunity.IsOverdue() should return false for future date")
	}

	// Past close date
	pastDate := time.Now().Add(-10 * 24 * time.Hour)
	opp.SetExpectedCloseDate(pastDate)
	if !opp.IsOverdue() {
		t.Error("Opportunity.IsOverdue() should return true for past date")
	}

	// Closed opportunity
	opp.Status = OpportunityStatusWon
	if opp.IsOverdue() {
		t.Error("Opportunity.IsOverdue() should return false for closed opportunity")
	}
}

func TestOpportunity_IsStale(t *testing.T) {
	opp, _ := createTestOpportunity(t)

	// Not stale with recent creation
	if opp.IsStale(30) {
		t.Error("Opportunity.IsStale() should return false for new opportunity")
	}

	// Stale with old creation and no activity
	opp.CreatedAt = time.Now().Add(-60 * 24 * time.Hour)
	if !opp.IsStale(30) {
		t.Error("Opportunity.IsStale() should return true for old opportunity without activity")
	}

	// Not stale with recent activity
	activityTime := time.Now().Add(-5 * 24 * time.Hour)
	opp.LastActivityAt = &activityTime
	if opp.IsStale(30) {
		t.Error("Opportunity.IsStale() should return false with recent activity")
	}

	// Closed opportunity is never stale
	opp.Status = OpportunityStatusWon
	if opp.IsStale(1) {
		t.Error("Opportunity.IsStale() should return false for closed opportunity")
	}
}

func TestOpportunity_Events(t *testing.T) {
	opp, _ := createTestOpportunity(t)
	opp.ClearEvents()

	if len(opp.GetEvents()) != 0 {
		t.Error("Opportunity.GetEvents() should return empty after ClearEvents")
	}

	opp.AddEvent(&OpportunityCreatedEvent{})
	if len(opp.GetEvents()) != 1 {
		t.Error("Opportunity.AddEvent() should add event")
	}

	opp.ClearEvents()
	if len(opp.GetEvents()) != 0 {
		t.Error("Opportunity.ClearEvents() should clear all events")
	}
}
