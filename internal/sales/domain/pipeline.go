// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

// Pipeline errors
var (
	ErrPipelineNotFound       = errors.New("pipeline not found")
	ErrPipelineAlreadyExists  = errors.New("pipeline already exists")
	ErrPipelineInactive       = errors.New("pipeline is inactive")
	ErrPipelineHasOpportunities = errors.New("pipeline has active opportunities")
	ErrStageNotFound          = errors.New("stage not found")
	ErrStageAlreadyExists     = errors.New("stage already exists")
	ErrStageInUse             = errors.New("stage is in use")
	ErrInvalidStageOrder      = errors.New("invalid stage order")
	ErrMinimumStagesRequired  = errors.New("minimum 2 stages required")
	ErrDefaultPipelineRequired = errors.New("at least one default pipeline required")
	ErrCannotDeleteDefaultPipeline = errors.New("cannot delete default pipeline")
)

// StageType represents the type of a pipeline stage.
type StageType string

const (
	StageTypeOpen       StageType = "open"       // Active, in progress
	StageTypeWon        StageType = "won"        // Successfully closed
	StageTypeLost       StageType = "lost"       // Lost/abandoned
	StageTypeQualifying StageType = "qualifying" // Qualification stage
	StageTypeNegotiating StageType = "negotiating" // Negotiation stage
)

// ValidStageTypes returns all valid stage types.
func ValidStageTypes() []StageType {
	return []StageType{
		StageTypeOpen,
		StageTypeWon,
		StageTypeLost,
		StageTypeQualifying,
		StageTypeNegotiating,
	}
}

// IsValid checks if the stage type is valid.
func (t StageType) IsValid() bool {
	for _, valid := range ValidStageTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// IsClosedType returns true if this is a closed stage type (won or lost).
func (t StageType) IsClosedType() bool {
	return t == StageTypeWon || t == StageTypeLost
}

// Stage represents a stage in a sales pipeline.
type Stage struct {
	ID           uuid.UUID  `json:"id" bson:"_id"`
	PipelineID   uuid.UUID  `json:"pipeline_id" bson:"pipeline_id"`
	Name         string     `json:"name" bson:"name"`
	Description  string     `json:"description,omitempty" bson:"description,omitempty"`
	Type         StageType  `json:"type" bson:"type"`
	Order        int        `json:"order" bson:"order"`
	Probability  int        `json:"probability" bson:"probability"` // 0-100
	Color        string     `json:"color,omitempty" bson:"color,omitempty"`
	IsActive     bool       `json:"is_active" bson:"is_active"`
	RottenDays   int        `json:"rotten_days,omitempty" bson:"rotten_days,omitempty"` // Days until opportunity is considered stale
	AutoActions  []AutoAction `json:"auto_actions,omitempty" bson:"auto_actions,omitempty"`
	CreatedAt    time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" bson:"updated_at"`
}

// AutoAction represents an automatic action triggered by stage entry.
type AutoAction struct {
	Type       string                 `json:"type" bson:"type"` // email, notification, task, webhook
	Config     map[string]interface{} `json:"config" bson:"config"`
	DelayHours int                    `json:"delay_hours,omitempty" bson:"delay_hours,omitempty"`
}

// NewStage creates a new stage.
func NewStage(pipelineID uuid.UUID, name string, stageType StageType, order int, probability int) (*Stage, error) {
	if name == "" {
		return nil, errors.New("stage name is required")
	}
	if !stageType.IsValid() {
		return nil, errors.New("invalid stage type")
	}
	if probability < 0 || probability > 100 {
		return nil, errors.New("probability must be between 0 and 100")
	}

	now := time.Now().UTC()
	return &Stage{
		ID:          uuid.New(),
		PipelineID:  pipelineID,
		Name:        name,
		Type:        stageType,
		Order:       order,
		Probability: probability,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update updates stage details.
func (s *Stage) Update(name, description string, probability int, color string, rottenDays int) error {
	if name != "" {
		s.Name = name
	}
	s.Description = description
	if probability >= 0 && probability <= 100 {
		s.Probability = probability
	}
	s.Color = color
	s.RottenDays = rottenDays
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// Deactivate deactivates the stage.
func (s *Stage) Deactivate() {
	s.IsActive = false
	s.UpdatedAt = time.Now().UTC()
}

// Activate activates the stage.
func (s *Stage) Activate() {
	s.IsActive = true
	s.UpdatedAt = time.Now().UTC()
}

// AddAutoAction adds an automatic action to the stage.
func (s *Stage) AddAutoAction(action AutoAction) {
	s.AutoActions = append(s.AutoActions, action)
	s.UpdatedAt = time.Now().UTC()
}

// RemoveAutoAction removes an automatic action by index.
func (s *Stage) RemoveAutoAction(index int) {
	if index >= 0 && index < len(s.AutoActions) {
		s.AutoActions = append(s.AutoActions[:index], s.AutoActions[index+1:]...)
		s.UpdatedAt = time.Now().UTC()
	}
}

// Pipeline represents a sales pipeline configuration.
type Pipeline struct {
	ID                uuid.UUID  `json:"id" bson:"_id"`
	TenantID          uuid.UUID  `json:"tenant_id" bson:"tenant_id"`
	Name              string     `json:"name" bson:"name"`
	Description       string     `json:"description,omitempty" bson:"description,omitempty"`
	IsDefault         bool       `json:"is_default" bson:"is_default"`
	IsActive          bool       `json:"is_active" bson:"is_active"`
	Currency          string     `json:"currency" bson:"currency"`
	Stages            []*Stage   `json:"stages" bson:"stages"`
	WinReasons        []string   `json:"win_reasons,omitempty" bson:"win_reasons,omitempty"`
	LossReasons       []string   `json:"loss_reasons,omitempty" bson:"loss_reasons,omitempty"`
	RequiredFields    []string   `json:"required_fields,omitempty" bson:"required_fields,omitempty"`
	CustomFields      []CustomFieldDef `json:"custom_fields,omitempty" bson:"custom_fields,omitempty"`
	OpportunityCount  int64      `json:"opportunity_count" bson:"opportunity_count"`
	TotalValue        Money      `json:"total_value" bson:"total_value"`
	WonValue          Money      `json:"won_value" bson:"won_value"`
	CreatedBy         uuid.UUID  `json:"created_by" bson:"created_by"`
	CreatedAt         time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" bson:"updated_at"`
	Version           int        `json:"version" bson:"version"`
}

// CustomFieldDef defines a custom field for opportunities.
type CustomFieldDef struct {
	Name        string   `json:"name" bson:"name"`
	Type        string   `json:"type" bson:"type"` // text, number, date, select, multiselect
	Label       string   `json:"label" bson:"label"`
	Required    bool     `json:"required" bson:"required"`
	Options     []string `json:"options,omitempty" bson:"options,omitempty"` // For select types
	Default     string   `json:"default,omitempty" bson:"default,omitempty"`
	Placeholder string   `json:"placeholder,omitempty" bson:"placeholder,omitempty"`
}

// NewPipeline creates a new pipeline with default stages.
func NewPipeline(tenantID uuid.UUID, name, currency string, createdBy uuid.UUID) (*Pipeline, error) {
	if name == "" {
		return nil, errors.New("pipeline name is required")
	}
	if !IsSupportedCurrency(currency) {
		return nil, ErrInvalidCurrency
	}

	now := time.Now().UTC()
	pipelineID := uuid.New()

	// Create default stages
	stages := []*Stage{
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Qualified",
			Type:        StageTypeQualifying,
			Order:       1,
			Probability: 10,
			IsActive:    true,
			Color:       "#3498db",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Contact Made",
			Type:        StageTypeOpen,
			Order:       2,
			Probability: 20,
			IsActive:    true,
			Color:       "#9b59b6",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Demo Scheduled",
			Type:        StageTypeOpen,
			Order:       3,
			Probability: 40,
			IsActive:    true,
			Color:       "#1abc9c",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Proposal Made",
			Type:        StageTypeNegotiating,
			Order:       4,
			Probability: 60,
			IsActive:    true,
			Color:       "#f39c12",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          uuid.New(),
			PipelineID:  pipelineID,
			Name:        "Negotiations Started",
			Type:        StageTypeNegotiating,
			Order:       5,
			Probability: 80,
			IsActive:    true,
			Color:       "#e67e22",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	totalValue, _ := Zero(currency)
	wonValue, _ := Zero(currency)

	return &Pipeline{
		ID:           pipelineID,
		TenantID:     tenantID,
		Name:         name,
		IsDefault:    false,
		IsActive:     true,
		Currency:     currency,
		Stages:       stages,
		WinReasons:   []string{"Price", "Product Fit", "Relationship", "Timing", "Other"},
		LossReasons:  []string{"Price Too High", "Competitor Won", "No Budget", "No Decision", "Lost Contact", "Other"},
		TotalValue:   totalValue,
		WonValue:     wonValue,
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}, nil
}

// Update updates pipeline details.
func (p *Pipeline) Update(name, description string) {
	if name != "" {
		p.Name = name
	}
	p.Description = description
	p.UpdatedAt = time.Now().UTC()
}

// SetAsDefault sets this pipeline as the default.
func (p *Pipeline) SetAsDefault() {
	p.IsDefault = true
	p.UpdatedAt = time.Now().UTC()
}

// UnsetDefault removes the default flag.
func (p *Pipeline) UnsetDefault() {
	p.IsDefault = false
	p.UpdatedAt = time.Now().UTC()
}

// Activate activates the pipeline.
func (p *Pipeline) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now().UTC()
}

// Deactivate deactivates the pipeline.
func (p *Pipeline) Deactivate() error {
	if p.IsDefault {
		return ErrCannotDeleteDefaultPipeline
	}
	if p.OpportunityCount > 0 {
		return ErrPipelineHasOpportunities
	}
	p.IsActive = false
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// AddStage adds a new stage to the pipeline.
func (p *Pipeline) AddStage(name string, stageType StageType, probability int) (*Stage, error) {
	// Check for duplicate name
	for _, s := range p.Stages {
		if s.Name == name && s.IsActive {
			return nil, ErrStageAlreadyExists
		}
	}

	// Calculate order (add at the end of open stages, before closed stages)
	order := 1
	for _, s := range p.Stages {
		if s.IsActive && !s.Type.IsClosedType() && s.Order >= order {
			order = s.Order + 1
		}
	}

	stage, err := NewStage(p.ID, name, stageType, order, probability)
	if err != nil {
		return nil, err
	}

	p.Stages = append(p.Stages, stage)
	p.sortStages()
	p.UpdatedAt = time.Now().UTC()

	return stage, nil
}

// UpdateStage updates an existing stage.
func (p *Pipeline) UpdateStage(stageID uuid.UUID, name, description string, probability int, color string, rottenDays int) error {
	stage := p.GetStage(stageID)
	if stage == nil {
		return ErrStageNotFound
	}

	// Check for duplicate name
	for _, s := range p.Stages {
		if s.ID != stageID && s.Name == name && s.IsActive {
			return ErrStageAlreadyExists
		}
	}

	if err := stage.Update(name, description, probability, color, rottenDays); err != nil {
		return err
	}

	p.UpdatedAt = time.Now().UTC()
	return nil
}

// RemoveStage removes a stage from the pipeline.
func (p *Pipeline) RemoveStage(stageID uuid.UUID) error {
	stage := p.GetStage(stageID)
	if stage == nil {
		return ErrStageNotFound
	}

	// Count active stages
	activeCount := 0
	for _, s := range p.Stages {
		if s.IsActive {
			activeCount++
		}
	}

	if activeCount <= 2 {
		return ErrMinimumStagesRequired
	}

	// Deactivate instead of delete
	stage.Deactivate()
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// ReorderStages reorders stages based on the provided order.
func (p *Pipeline) ReorderStages(stageIDs []uuid.UUID) error {
	if len(stageIDs) == 0 {
		return ErrInvalidStageOrder
	}

	// Verify all stage IDs exist
	stageMap := make(map[uuid.UUID]*Stage)
	for _, s := range p.Stages {
		stageMap[s.ID] = s
	}

	for _, id := range stageIDs {
		if _, exists := stageMap[id]; !exists {
			return ErrStageNotFound
		}
	}

	// Update order
	for i, id := range stageIDs {
		if stage, exists := stageMap[id]; exists {
			stage.Order = i + 1
			stage.UpdatedAt = time.Now().UTC()
		}
	}

	p.sortStages()
	p.UpdatedAt = time.Now().UTC()
	return nil
}

// GetStage returns a stage by ID.
func (p *Pipeline) GetStage(stageID uuid.UUID) *Stage {
	for _, s := range p.Stages {
		if s.ID == stageID {
			return s
		}
	}
	return nil
}

// GetStageByName returns a stage by name.
func (p *Pipeline) GetStageByName(name string) *Stage {
	for _, s := range p.Stages {
		if s.Name == name && s.IsActive {
			return s
		}
	}
	return nil
}

// GetActiveStages returns all active stages sorted by order.
func (p *Pipeline) GetActiveStages() []*Stage {
	stages := make([]*Stage, 0)
	for _, s := range p.Stages {
		if s.IsActive {
			stages = append(stages, s)
		}
	}
	sort.Slice(stages, func(i, j int) bool {
		return stages[i].Order < stages[j].Order
	})
	return stages
}

// GetFirstStage returns the first active stage.
func (p *Pipeline) GetFirstStage() *Stage {
	stages := p.GetActiveStages()
	if len(stages) > 0 {
		return stages[0]
	}
	return nil
}

// GetWonStage returns the won stage (creates one if not exists).
func (p *Pipeline) GetWonStage() *Stage {
	for _, s := range p.Stages {
		if s.Type == StageTypeWon && s.IsActive {
			return s
		}
	}
	return nil
}

// GetLostStage returns the lost stage (creates one if not exists).
func (p *Pipeline) GetLostStage() *Stage {
	for _, s := range p.Stages {
		if s.Type == StageTypeLost && s.IsActive {
			return s
		}
	}
	return nil
}

// EnsureClosedStages ensures won and lost stages exist.
func (p *Pipeline) EnsureClosedStages() {
	now := time.Now().UTC()

	if p.GetWonStage() == nil {
		maxOrder := 0
		for _, s := range p.Stages {
			if s.Order > maxOrder {
				maxOrder = s.Order
			}
		}

		wonStage := &Stage{
			ID:          uuid.New(),
			PipelineID:  p.ID,
			Name:        "Won",
			Type:        StageTypeWon,
			Order:       maxOrder + 1,
			Probability: 100,
			IsActive:    true,
			Color:       "#27ae60",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		p.Stages = append(p.Stages, wonStage)
	}

	if p.GetLostStage() == nil {
		maxOrder := 0
		for _, s := range p.Stages {
			if s.Order > maxOrder {
				maxOrder = s.Order
			}
		}

		lostStage := &Stage{
			ID:          uuid.New(),
			PipelineID:  p.ID,
			Name:        "Lost",
			Type:        StageTypeLost,
			Order:       maxOrder + 1,
			Probability: 0,
			IsActive:    true,
			Color:       "#e74c3c",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		p.Stages = append(p.Stages, lostStage)
	}

	p.sortStages()
}

// sortStages sorts stages by order.
func (p *Pipeline) sortStages() {
	sort.Slice(p.Stages, func(i, j int) bool {
		// Closed stages always last
		if p.Stages[i].Type.IsClosedType() && !p.Stages[j].Type.IsClosedType() {
			return false
		}
		if !p.Stages[i].Type.IsClosedType() && p.Stages[j].Type.IsClosedType() {
			return true
		}
		return p.Stages[i].Order < p.Stages[j].Order
	})
}

// SetWinReasons sets the win reasons.
func (p *Pipeline) SetWinReasons(reasons []string) {
	p.WinReasons = reasons
	p.UpdatedAt = time.Now().UTC()
}

// SetLossReasons sets the loss reasons.
func (p *Pipeline) SetLossReasons(reasons []string) {
	p.LossReasons = reasons
	p.UpdatedAt = time.Now().UTC()
}

// AddCustomField adds a custom field definition.
func (p *Pipeline) AddCustomField(field CustomFieldDef) {
	p.CustomFields = append(p.CustomFields, field)
	p.UpdatedAt = time.Now().UTC()
}

// RemoveCustomField removes a custom field by name.
func (p *Pipeline) RemoveCustomField(name string) {
	for i, f := range p.CustomFields {
		if f.Name == name {
			p.CustomFields = append(p.CustomFields[:i], p.CustomFields[i+1:]...)
			p.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// UpdateMetrics updates pipeline metrics.
func (p *Pipeline) UpdateMetrics(opportunityCount int64, totalValue, wonValue Money) {
	p.OpportunityCount = opportunityCount
	p.TotalValue = totalValue
	p.WonValue = wonValue
	p.UpdatedAt = time.Now().UTC()
}

// CanDelete checks if the pipeline can be deleted.
func (p *Pipeline) CanDelete() error {
	if p.IsDefault {
		return ErrCannotDeleteDefaultPipeline
	}
	if p.OpportunityCount > 0 {
		return ErrPipelineHasOpportunities
	}
	return nil
}
