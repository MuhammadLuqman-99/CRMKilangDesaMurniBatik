// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStageType_IsValid(t *testing.T) {
	tests := []struct {
		stageType StageType
		expected  bool
	}{
		{StageTypeOpen, true},
		{StageTypeWon, true},
		{StageTypeLost, true},
		{StageTypeQualifying, true},
		{StageTypeNegotiating, true},
		{StageType("invalid"), false},
		{StageType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.stageType), func(t *testing.T) {
			if tt.stageType.IsValid() != tt.expected {
				t.Errorf("StageType.IsValid() = %v, want %v", tt.stageType.IsValid(), tt.expected)
			}
		})
	}
}

func TestStageType_IsClosedType(t *testing.T) {
	tests := []struct {
		stageType StageType
		expected  bool
	}{
		{StageTypeWon, true},
		{StageTypeLost, true},
		{StageTypeOpen, false},
		{StageTypeQualifying, false},
		{StageTypeNegotiating, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.stageType), func(t *testing.T) {
			if tt.stageType.IsClosedType() != tt.expected {
				t.Errorf("StageType.IsClosedType() = %v, want %v", tt.stageType.IsClosedType(), tt.expected)
			}
		})
	}
}

func TestValidStageTypes(t *testing.T) {
	types := ValidStageTypes()
	if len(types) != 5 {
		t.Errorf("ValidStageTypes() len = %v, want 5", len(types))
	}
}

func TestNewStage(t *testing.T) {
	pipelineID := uuid.New()

	tests := []struct {
		name        string
		pipelineID  uuid.UUID
		stageName   string
		stageType   StageType
		order       int
		probability int
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid stage",
			pipelineID:  pipelineID,
			stageName:   "Qualification",
			stageType:   StageTypeQualifying,
			order:       1,
			probability: 20,
			wantErr:     false,
		},
		{
			name:        "missing name",
			pipelineID:  pipelineID,
			stageName:   "",
			stageType:   StageTypeOpen,
			order:       1,
			probability: 20,
			wantErr:     true,
			errMsg:      "stage name is required",
		},
		{
			name:        "invalid stage type",
			pipelineID:  pipelineID,
			stageName:   "Stage",
			stageType:   StageType("invalid"),
			order:       1,
			probability: 20,
			wantErr:     true,
			errMsg:      "invalid stage type",
		},
		{
			name:        "probability below 0",
			pipelineID:  pipelineID,
			stageName:   "Stage",
			stageType:   StageTypeOpen,
			order:       1,
			probability: -10,
			wantErr:     true,
			errMsg:      "probability must be between 0 and 100",
		},
		{
			name:        "probability above 100",
			pipelineID:  pipelineID,
			stageName:   "Stage",
			stageType:   StageTypeOpen,
			order:       1,
			probability: 110,
			wantErr:     true,
			errMsg:      "probability must be between 0 and 100",
		},
		{
			name:        "probability at 0",
			pipelineID:  pipelineID,
			stageName:   "Lost Stage",
			stageType:   StageTypeLost,
			order:       99,
			probability: 0,
			wantErr:     false,
		},
		{
			name:        "probability at 100",
			pipelineID:  pipelineID,
			stageName:   "Won Stage",
			stageType:   StageTypeWon,
			order:       100,
			probability: 100,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage, err := NewStage(tt.pipelineID, tt.stageName, tt.stageType, tt.order, tt.probability)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewStage() expected error, got nil")
				}
				if err != nil && err.Error() != tt.errMsg {
					t.Errorf("NewStage() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("NewStage() unexpected error = %v", err)
				}
				if stage == nil {
					t.Fatal("NewStage() returned nil stage")
				}
				if stage.ID == uuid.Nil {
					t.Error("NewStage() should generate ID")
				}
				if stage.PipelineID != tt.pipelineID {
					t.Errorf("NewStage() PipelineID = %v, want %v", stage.PipelineID, tt.pipelineID)
				}
				if stage.Name != tt.stageName {
					t.Errorf("NewStage() Name = %v, want %v", stage.Name, tt.stageName)
				}
				if stage.Type != tt.stageType {
					t.Errorf("NewStage() Type = %v, want %v", stage.Type, tt.stageType)
				}
				if stage.Order != tt.order {
					t.Errorf("NewStage() Order = %v, want %v", stage.Order, tt.order)
				}
				if stage.Probability != tt.probability {
					t.Errorf("NewStage() Probability = %v, want %v", stage.Probability, tt.probability)
				}
				if !stage.IsActive {
					t.Error("NewStage() should be active by default")
				}
			}
		})
	}
}

func TestStage_Update(t *testing.T) {
	pipelineID := uuid.New()
	stage, _ := NewStage(pipelineID, "Original", StageTypeOpen, 1, 20)

	err := stage.Update("Updated", "New description", 50, "#FF0000", 14)
	if err != nil {
		t.Fatalf("Stage.Update() unexpected error = %v", err)
	}

	if stage.Name != "Updated" {
		t.Errorf("Stage.Update() Name = %v, want Updated", stage.Name)
	}
	if stage.Description != "New description" {
		t.Errorf("Stage.Update() Description = %v", stage.Description)
	}
	if stage.Probability != 50 {
		t.Errorf("Stage.Update() Probability = %v, want 50", stage.Probability)
	}
	if stage.Color != "#FF0000" {
		t.Errorf("Stage.Update() Color = %v", stage.Color)
	}
	if stage.RottenDays != 14 {
		t.Errorf("Stage.Update() RottenDays = %v, want 14", stage.RottenDays)
	}
}

func TestStage_Update_EmptyName(t *testing.T) {
	pipelineID := uuid.New()
	stage, _ := NewStage(pipelineID, "Original", StageTypeOpen, 1, 20)

	err := stage.Update("", "Description", 50, "", 0)
	if err != nil {
		t.Fatalf("Stage.Update() unexpected error = %v", err)
	}

	// Empty name should keep original
	if stage.Name != "Original" {
		t.Errorf("Stage.Update() with empty name should keep original, got %v", stage.Name)
	}
}

func TestStage_DeactivateActivate(t *testing.T) {
	pipelineID := uuid.New()
	stage, _ := NewStage(pipelineID, "Test", StageTypeOpen, 1, 20)

	stage.Deactivate()
	if stage.IsActive {
		t.Error("Stage.Deactivate() should set IsActive to false")
	}

	stage.Activate()
	if !stage.IsActive {
		t.Error("Stage.Activate() should set IsActive to true")
	}
}

func TestStage_AutoActions(t *testing.T) {
	pipelineID := uuid.New()
	stage, _ := NewStage(pipelineID, "Test", StageTypeOpen, 1, 20)

	action := AutoAction{
		Type:       "email",
		Config:     map[string]interface{}{"template": "welcome"},
		DelayHours: 24,
	}

	stage.AddAutoAction(action)
	if len(stage.AutoActions) != 1 {
		t.Errorf("Stage.AddAutoAction() len = %v, want 1", len(stage.AutoActions))
	}

	stage.RemoveAutoAction(0)
	if len(stage.AutoActions) != 0 {
		t.Errorf("Stage.RemoveAutoAction() len = %v, want 0", len(stage.AutoActions))
	}

	// Remove invalid index should not panic
	stage.RemoveAutoAction(-1)
	stage.RemoveAutoAction(100)
}

func TestNewPipeline(t *testing.T) {
	tenantID := uuid.New()
	createdBy := uuid.New()

	tests := []struct {
		name      string
		tenantID  uuid.UUID
		pipeName  string
		currency  string
		createdBy uuid.UUID
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid pipeline",
			tenantID:  tenantID,
			pipeName:  "Sales Pipeline",
			currency:  "USD",
			createdBy: createdBy,
			wantErr:   false,
		},
		{
			name:      "missing name",
			tenantID:  tenantID,
			pipeName:  "",
			currency:  "USD",
			createdBy: createdBy,
			wantErr:   true,
			errMsg:    "pipeline name is required",
		},
		{
			name:      "invalid currency",
			tenantID:  tenantID,
			pipeName:  "Pipeline",
			currency:  "INVALID",
			createdBy: createdBy,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := NewPipeline(tt.tenantID, tt.pipeName, tt.currency, tt.createdBy)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPipeline() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewPipeline() unexpected error = %v", err)
				}
				if pipeline == nil {
					t.Fatal("NewPipeline() returned nil pipeline")
				}
				if pipeline.ID == uuid.Nil {
					t.Error("NewPipeline() should generate ID")
				}
				if pipeline.TenantID != tt.tenantID {
					t.Errorf("NewPipeline() TenantID = %v, want %v", pipeline.TenantID, tt.tenantID)
				}
				if pipeline.Name != tt.pipeName {
					t.Errorf("NewPipeline() Name = %v, want %v", pipeline.Name, tt.pipeName)
				}
				if pipeline.Currency != tt.currency {
					t.Errorf("NewPipeline() Currency = %v, want %v", pipeline.Currency, tt.currency)
				}
				if !pipeline.IsActive {
					t.Error("NewPipeline() should be active by default")
				}
				if pipeline.IsDefault {
					t.Error("NewPipeline() should not be default by default")
				}
				// Should have default stages
				if len(pipeline.Stages) < 2 {
					t.Errorf("NewPipeline() should have default stages, got %d", len(pipeline.Stages))
				}
				if pipeline.Version != 1 {
					t.Errorf("NewPipeline() Version = %v, want 1", pipeline.Version)
				}
			}
		})
	}
}

func createTestPipeline(t *testing.T) *Pipeline {
	tenantID := uuid.New()
	createdBy := uuid.New()

	pipeline, err := NewPipeline(tenantID, "Test Pipeline", "USD", createdBy)
	if err != nil {
		t.Fatalf("createTestPipeline() failed: %v", err)
	}
	return pipeline
}

func TestPipeline_Update(t *testing.T) {
	pipeline := createTestPipeline(t)

	pipeline.Update("Updated Pipeline", "New description")

	if pipeline.Name != "Updated Pipeline" {
		t.Errorf("Pipeline.Update() Name = %v", pipeline.Name)
	}
	if pipeline.Description != "New description" {
		t.Errorf("Pipeline.Update() Description = %v", pipeline.Description)
	}
}

func TestPipeline_Update_EmptyName(t *testing.T) {
	pipeline := createTestPipeline(t)
	originalName := pipeline.Name

	pipeline.Update("", "Description")

	// Empty name should keep original
	if pipeline.Name != originalName {
		t.Errorf("Pipeline.Update() with empty name should keep original, got %v", pipeline.Name)
	}
}

func TestPipeline_SetAsDefault(t *testing.T) {
	pipeline := createTestPipeline(t)

	pipeline.SetAsDefault()

	if !pipeline.IsDefault {
		t.Error("Pipeline.SetAsDefault() should set IsDefault to true")
	}
}

func TestPipeline_UnsetDefault(t *testing.T) {
	pipeline := createTestPipeline(t)
	pipeline.SetAsDefault()

	pipeline.UnsetDefault()

	if pipeline.IsDefault {
		t.Error("Pipeline.UnsetDefault() should set IsDefault to false")
	}
}

func TestPipeline_ActivateDeactivate(t *testing.T) {
	pipeline := createTestPipeline(t)

	// Deactivate
	err := pipeline.Deactivate()
	if err != nil {
		t.Fatalf("Pipeline.Deactivate() unexpected error = %v", err)
	}
	if pipeline.IsActive {
		t.Error("Pipeline.Deactivate() should set IsActive to false")
	}

	// Activate
	pipeline.Activate()
	if !pipeline.IsActive {
		t.Error("Pipeline.Activate() should set IsActive to true")
	}
}

func TestPipeline_Deactivate_DefaultPipeline(t *testing.T) {
	pipeline := createTestPipeline(t)
	pipeline.SetAsDefault()

	err := pipeline.Deactivate()
	if err != ErrCannotDeleteDefaultPipeline {
		t.Errorf("Pipeline.Deactivate() on default pipeline should return ErrCannotDeleteDefaultPipeline, got %v", err)
	}
}

func TestPipeline_Deactivate_HasOpportunities(t *testing.T) {
	pipeline := createTestPipeline(t)
	pipeline.OpportunityCount = 5

	err := pipeline.Deactivate()
	if err != ErrPipelineHasOpportunities {
		t.Errorf("Pipeline.Deactivate() with opportunities should return ErrPipelineHasOpportunities, got %v", err)
	}
}

func TestPipeline_AddStage(t *testing.T) {
	pipeline := createTestPipeline(t)
	initialStageCount := len(pipeline.Stages)

	stage, err := pipeline.AddStage("New Stage", StageTypeOpen, 40)
	if err != nil {
		t.Fatalf("Pipeline.AddStage() unexpected error = %v", err)
	}

	if stage == nil {
		t.Fatal("Pipeline.AddStage() returned nil stage")
	}
	if len(pipeline.Stages) != initialStageCount+1 {
		t.Errorf("Pipeline.AddStage() len = %v, want %v", len(pipeline.Stages), initialStageCount+1)
	}
}

func TestPipeline_AddStage_DuplicateName(t *testing.T) {
	pipeline := createTestPipeline(t)
	existingName := pipeline.Stages[0].Name

	_, err := pipeline.AddStage(existingName, StageTypeOpen, 40)
	if err != ErrStageAlreadyExists {
		t.Errorf("Pipeline.AddStage() with duplicate name should return ErrStageAlreadyExists, got %v", err)
	}
}

func TestPipeline_UpdateStage(t *testing.T) {
	pipeline := createTestPipeline(t)
	stageID := pipeline.Stages[0].ID

	err := pipeline.UpdateStage(stageID, "Updated Stage", "Description", 60, "#00FF00", 7)
	if err != nil {
		t.Fatalf("Pipeline.UpdateStage() unexpected error = %v", err)
	}

	stage := pipeline.GetStage(stageID)
	if stage.Name != "Updated Stage" {
		t.Errorf("Pipeline.UpdateStage() Name = %v", stage.Name)
	}
	if stage.Probability != 60 {
		t.Errorf("Pipeline.UpdateStage() Probability = %v", stage.Probability)
	}
}

func TestPipeline_UpdateStage_NotFound(t *testing.T) {
	pipeline := createTestPipeline(t)

	err := pipeline.UpdateStage(uuid.New(), "Updated", "Desc", 50, "", 0)
	if err != ErrStageNotFound {
		t.Errorf("Pipeline.UpdateStage() with invalid ID should return ErrStageNotFound, got %v", err)
	}
}

func TestPipeline_UpdateStage_DuplicateName(t *testing.T) {
	pipeline := createTestPipeline(t)
	stageID := pipeline.Stages[0].ID
	existingName := pipeline.Stages[1].Name

	err := pipeline.UpdateStage(stageID, existingName, "Desc", 50, "", 0)
	if err != ErrStageAlreadyExists {
		t.Errorf("Pipeline.UpdateStage() with duplicate name should return ErrStageAlreadyExists, got %v", err)
	}
}

func TestPipeline_RemoveStage(t *testing.T) {
	pipeline := createTestPipeline(t)
	// Add extra stages to meet minimum requirement
	pipeline.AddStage("Extra Stage 1", StageTypeOpen, 30)
	pipeline.AddStage("Extra Stage 2", StageTypeOpen, 40)

	stageID := pipeline.Stages[0].ID
	initialCount := len(pipeline.GetActiveStages())

	err := pipeline.RemoveStage(stageID)
	if err != nil {
		t.Fatalf("Pipeline.RemoveStage() unexpected error = %v", err)
	}

	// Stage should be deactivated, not deleted
	stage := pipeline.GetStage(stageID)
	if stage == nil {
		t.Error("Pipeline.RemoveStage() should deactivate, not delete stage")
	}
	if stage.IsActive {
		t.Error("Pipeline.RemoveStage() should deactivate stage")
	}

	// Active stages count should decrease
	if len(pipeline.GetActiveStages()) != initialCount-1 {
		t.Errorf("Pipeline.RemoveStage() active count = %v, want %v", len(pipeline.GetActiveStages()), initialCount-1)
	}
}

func TestPipeline_RemoveStage_NotFound(t *testing.T) {
	pipeline := createTestPipeline(t)

	err := pipeline.RemoveStage(uuid.New())
	if err != ErrStageNotFound {
		t.Errorf("Pipeline.RemoveStage() with invalid ID should return ErrStageNotFound, got %v", err)
	}
}

func TestPipeline_RemoveStage_MinimumRequired(t *testing.T) {
	pipeline := createTestPipeline(t)
	// Deactivate all but 2 stages
	activeStages := pipeline.GetActiveStages()
	for i := 2; i < len(activeStages); i++ {
		activeStages[i].Deactivate()
	}

	err := pipeline.RemoveStage(pipeline.GetActiveStages()[0].ID)
	if err != ErrMinimumStagesRequired {
		t.Errorf("Pipeline.RemoveStage() with minimum stages should return ErrMinimumStagesRequired, got %v", err)
	}
}

func TestPipeline_ReorderStages(t *testing.T) {
	pipeline := createTestPipeline(t)

	// Get current stage IDs
	stageIDs := make([]uuid.UUID, len(pipeline.Stages))
	for i, s := range pipeline.Stages {
		stageIDs[i] = s.ID
	}

	// Reverse the order
	for i, j := 0, len(stageIDs)-1; i < j; i, j = i+1, j-1 {
		stageIDs[i], stageIDs[j] = stageIDs[j], stageIDs[i]
	}

	err := pipeline.ReorderStages(stageIDs)
	if err != nil {
		t.Fatalf("Pipeline.ReorderStages() unexpected error = %v", err)
	}

	// First stage should now have highest order number from original
	if pipeline.GetStage(stageIDs[0]).Order != 1 {
		t.Errorf("Pipeline.ReorderStages() first stage order = %v, want 1", pipeline.GetStage(stageIDs[0]).Order)
	}
}

func TestPipeline_ReorderStages_EmptyList(t *testing.T) {
	pipeline := createTestPipeline(t)

	err := pipeline.ReorderStages([]uuid.UUID{})
	if err != ErrInvalidStageOrder {
		t.Errorf("Pipeline.ReorderStages() with empty list should return ErrInvalidStageOrder, got %v", err)
	}
}

func TestPipeline_ReorderStages_InvalidID(t *testing.T) {
	pipeline := createTestPipeline(t)

	err := pipeline.ReorderStages([]uuid.UUID{uuid.New()})
	if err != ErrStageNotFound {
		t.Errorf("Pipeline.ReorderStages() with invalid ID should return ErrStageNotFound, got %v", err)
	}
}

func TestPipeline_GetStage(t *testing.T) {
	pipeline := createTestPipeline(t)
	existingID := pipeline.Stages[0].ID

	stage := pipeline.GetStage(existingID)
	if stage == nil {
		t.Error("Pipeline.GetStage() should return existing stage")
	}

	stage = pipeline.GetStage(uuid.New())
	if stage != nil {
		t.Error("Pipeline.GetStage() should return nil for non-existent ID")
	}
}

func TestPipeline_GetStageByName(t *testing.T) {
	pipeline := createTestPipeline(t)
	existingName := pipeline.Stages[0].Name

	stage := pipeline.GetStageByName(existingName)
	if stage == nil {
		t.Error("Pipeline.GetStageByName() should return existing stage")
	}

	stage = pipeline.GetStageByName("Non-existent")
	if stage != nil {
		t.Error("Pipeline.GetStageByName() should return nil for non-existent name")
	}
}

func TestPipeline_GetActiveStages(t *testing.T) {
	pipeline := createTestPipeline(t)
	totalStages := len(pipeline.Stages)

	// Deactivate one stage
	pipeline.Stages[0].Deactivate()

	activeStages := pipeline.GetActiveStages()
	if len(activeStages) != totalStages-1 {
		t.Errorf("Pipeline.GetActiveStages() len = %v, want %v", len(activeStages), totalStages-1)
	}

	// Active stages should be sorted by order
	for i := 1; i < len(activeStages); i++ {
		if activeStages[i-1].Order > activeStages[i].Order {
			t.Error("Pipeline.GetActiveStages() should return stages sorted by order")
			break
		}
	}
}

func TestPipeline_GetFirstStage(t *testing.T) {
	pipeline := createTestPipeline(t)

	firstStage := pipeline.GetFirstStage()
	if firstStage == nil {
		t.Fatal("Pipeline.GetFirstStage() should return first stage")
	}

	// First stage should have lowest order
	for _, s := range pipeline.Stages {
		if s.IsActive && s.Order < firstStage.Order {
			t.Error("Pipeline.GetFirstStage() should return stage with lowest order")
			break
		}
	}
}

func TestPipeline_GetWonLostStages(t *testing.T) {
	pipeline := createTestPipeline(t)
	pipeline.EnsureClosedStages()

	wonStage := pipeline.GetWonStage()
	if wonStage == nil {
		t.Error("Pipeline.GetWonStage() should return won stage after EnsureClosedStages")
	}
	if wonStage != nil && wonStage.Type != StageTypeWon {
		t.Error("Pipeline.GetWonStage() should return stage with type Won")
	}

	lostStage := pipeline.GetLostStage()
	if lostStage == nil {
		t.Error("Pipeline.GetLostStage() should return lost stage after EnsureClosedStages")
	}
	if lostStage != nil && lostStage.Type != StageTypeLost {
		t.Error("Pipeline.GetLostStage() should return stage with type Lost")
	}
}

func TestPipeline_EnsureClosedStages(t *testing.T) {
	pipeline := createTestPipeline(t)

	// Remove any existing closed stages
	for _, s := range pipeline.Stages {
		if s.Type == StageTypeWon || s.Type == StageTypeLost {
			s.Deactivate()
		}
	}

	pipeline.EnsureClosedStages()

	wonStage := pipeline.GetWonStage()
	if wonStage == nil {
		t.Error("Pipeline.EnsureClosedStages() should create won stage")
	}

	lostStage := pipeline.GetLostStage()
	if lostStage == nil {
		t.Error("Pipeline.EnsureClosedStages() should create lost stage")
	}
}

func TestPipeline_SetWinLossReasons(t *testing.T) {
	pipeline := createTestPipeline(t)

	winReasons := []string{"Price", "Quality", "Support"}
	pipeline.SetWinReasons(winReasons)
	if len(pipeline.WinReasons) != 3 {
		t.Errorf("Pipeline.SetWinReasons() len = %v, want 3", len(pipeline.WinReasons))
	}

	lossReasons := []string{"Price Too High", "Competitor"}
	pipeline.SetLossReasons(lossReasons)
	if len(pipeline.LossReasons) != 2 {
		t.Errorf("Pipeline.SetLossReasons() len = %v, want 2", len(pipeline.LossReasons))
	}
}

func TestPipeline_CustomFields(t *testing.T) {
	pipeline := createTestPipeline(t)

	field := CustomFieldDef{
		Name:     "industry",
		Type:     "select",
		Label:    "Industry",
		Required: true,
		Options:  []string{"Technology", "Healthcare", "Finance"},
	}

	pipeline.AddCustomField(field)
	if len(pipeline.CustomFields) != 1 {
		t.Errorf("Pipeline.AddCustomField() len = %v, want 1", len(pipeline.CustomFields))
	}

	pipeline.RemoveCustomField("industry")
	if len(pipeline.CustomFields) != 0 {
		t.Errorf("Pipeline.RemoveCustomField() len = %v, want 0", len(pipeline.CustomFields))
	}
}

func TestPipeline_UpdateMetrics(t *testing.T) {
	pipeline := createTestPipeline(t)

	totalValue, _ := NewMoneyFromFloat(100000, "USD")
	wonValue, _ := NewMoneyFromFloat(50000, "USD")

	pipeline.UpdateMetrics(10, totalValue, wonValue)

	if pipeline.OpportunityCount != 10 {
		t.Errorf("Pipeline.UpdateMetrics() OpportunityCount = %v, want 10", pipeline.OpportunityCount)
	}
	if pipeline.TotalValue.Amount != totalValue.Amount {
		t.Errorf("Pipeline.UpdateMetrics() TotalValue = %v, want %v", pipeline.TotalValue.Amount, totalValue.Amount)
	}
	if pipeline.WonValue.Amount != wonValue.Amount {
		t.Errorf("Pipeline.UpdateMetrics() WonValue = %v, want %v", pipeline.WonValue.Amount, wonValue.Amount)
	}
}

func TestPipeline_CanDelete(t *testing.T) {
	pipeline := createTestPipeline(t)

	// Can delete initially
	err := pipeline.CanDelete()
	if err != nil {
		t.Errorf("Pipeline.CanDelete() should return nil for deletable pipeline, got %v", err)
	}

	// Cannot delete if default
	pipeline.SetAsDefault()
	err = pipeline.CanDelete()
	if err != ErrCannotDeleteDefaultPipeline {
		t.Errorf("Pipeline.CanDelete() should return ErrCannotDeleteDefaultPipeline for default pipeline, got %v", err)
	}

	// Cannot delete if has opportunities
	pipeline.UnsetDefault()
	pipeline.OpportunityCount = 5
	err = pipeline.CanDelete()
	if err != ErrPipelineHasOpportunities {
		t.Errorf("Pipeline.CanDelete() should return ErrPipelineHasOpportunities, got %v", err)
	}
}

func TestPipeline_StagesSorting(t *testing.T) {
	pipeline := createTestPipeline(t)
	pipeline.EnsureClosedStages()

	activeStages := pipeline.GetActiveStages()

	// Closed stages (Won/Lost) should be last
	var foundClosed bool
	for _, s := range activeStages {
		if s.Type.IsClosedType() {
			foundClosed = true
		} else if foundClosed {
			t.Error("Pipeline stages: closed stages should be at the end")
			break
		}
	}
}

func TestAutoAction(t *testing.T) {
	action := AutoAction{
		Type:       "email",
		Config:     map[string]interface{}{"template": "welcome", "subject": "Welcome!"},
		DelayHours: 24,
	}

	if action.Type != "email" {
		t.Errorf("AutoAction.Type = %v", action.Type)
	}
	if action.Config["template"] != "welcome" {
		t.Errorf("AutoAction.Config[template] = %v", action.Config["template"])
	}
	if action.DelayHours != 24 {
		t.Errorf("AutoAction.DelayHours = %v", action.DelayHours)
	}
}

func TestStage_Timestamps(t *testing.T) {
	pipelineID := uuid.New()
	stage, _ := NewStage(pipelineID, "Test", StageTypeOpen, 1, 20)

	createdAt := stage.CreatedAt
	if createdAt.IsZero() {
		t.Error("NewStage() should set CreatedAt")
	}

	time.Sleep(10 * time.Millisecond)
	stage.Update("Updated", "Desc", 30, "", 0)

	if !stage.UpdatedAt.After(createdAt) {
		t.Error("Stage.Update() should update UpdatedAt")
	}
}
