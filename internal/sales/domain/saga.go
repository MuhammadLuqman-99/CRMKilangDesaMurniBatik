// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Saga Errors
// ============================================================================

var (
	ErrSagaNotFound           = errors.New("saga not found")
	ErrSagaAlreadyCompleted   = errors.New("saga already completed")
	ErrSagaAlreadyCompensated = errors.New("saga already compensated")
	ErrSagaStepFailed         = errors.New("saga step failed")
	ErrSagaCompensationFailed = errors.New("saga compensation failed")
	ErrIdempotencyKeyExists   = errors.New("idempotency key already exists")
	ErrInvalidSagaState       = errors.New("invalid saga state")
	ErrInvalidSagaTransition  = errors.New("invalid saga state transition")
)

// ============================================================================
// Saga State
// ============================================================================

// SagaState represents the current state of a saga.
type SagaState string

const (
	// SagaStateStarted indicates the saga has been initiated but not completed.
	SagaStateStarted SagaState = "started"

	// SagaStateRunning indicates the saga is actively executing steps.
	SagaStateRunning SagaState = "running"

	// SagaStateCompleted indicates all saga steps completed successfully.
	SagaStateCompleted SagaState = "completed"

	// SagaStateCompensating indicates the saga is executing compensation steps.
	SagaStateCompensating SagaState = "compensating"

	// SagaStateCompensated indicates all compensation steps completed.
	SagaStateCompensated SagaState = "compensated"

	// SagaStateFailed indicates the saga failed and could not be compensated.
	SagaStateFailed SagaState = "failed"
)

// IsValid checks if the saga state is valid.
func (s SagaState) IsValid() bool {
	switch s {
	case SagaStateStarted, SagaStateRunning, SagaStateCompleted,
		SagaStateCompensating, SagaStateCompensated, SagaStateFailed:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the saga is in a terminal state.
func (s SagaState) IsTerminal() bool {
	return s == SagaStateCompleted || s == SagaStateCompensated || s == SagaStateFailed
}

// CanTransitionTo checks if a transition to the target state is valid.
func (s SagaState) CanTransitionTo(target SagaState) bool {
	switch s {
	case SagaStateStarted:
		return target == SagaStateRunning || target == SagaStateFailed
	case SagaStateRunning:
		return target == SagaStateCompleted || target == SagaStateCompensating || target == SagaStateFailed
	case SagaStateCompensating:
		return target == SagaStateCompensated || target == SagaStateFailed
	case SagaStateCompleted, SagaStateCompensated, SagaStateFailed:
		return false // Terminal states
	default:
		return false
	}
}

// ============================================================================
// Saga Step Types
// ============================================================================

// SagaStepType identifies the type of saga step.
type SagaStepType string

const (
	StepTypeValidateLead      SagaStepType = "validate_lead"
	StepTypeCreateCustomer    SagaStepType = "create_customer"
	StepTypeLookupCustomer    SagaStepType = "lookup_customer"
	StepTypeCreateOpportunity SagaStepType = "create_opportunity"
	StepTypeMarkLeadConverted SagaStepType = "mark_lead_converted"
	StepTypePublishEvents     SagaStepType = "publish_events"
)

// SagaStepStatus represents the execution status of a step.
type SagaStepStatus string

const (
	StepStatusPending     SagaStepStatus = "pending"
	StepStatusRunning     SagaStepStatus = "running"
	StepStatusCompleted   SagaStepStatus = "completed"
	StepStatusFailed      SagaStepStatus = "failed"
	StepStatusCompensated SagaStepStatus = "compensated"
	StepStatusSkipped     SagaStepStatus = "skipped"
)

// ============================================================================
// Saga Step
// ============================================================================

// SagaStep represents a single step in the saga workflow.
type SagaStep struct {
	// ID is the unique identifier for this step instance.
	ID uuid.UUID `json:"id"`

	// Type identifies what kind of step this is.
	Type SagaStepType `json:"type"`

	// Order determines the execution sequence.
	Order int `json:"order"`

	// Status is the current execution status.
	Status SagaStepStatus `json:"status"`

	// Input contains the data needed to execute the step.
	Input map[string]interface{} `json:"input,omitempty"`

	// Output contains the result data from execution.
	Output map[string]interface{} `json:"output,omitempty"`

	// Error contains error details if the step failed.
	Error *string `json:"error,omitempty"`

	// StartedAt is when execution began.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is when execution finished.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// CompensatedAt is when compensation was executed.
	CompensatedAt *time.Time `json:"compensated_at,omitempty"`

	// Compensatable indicates if this step can be rolled back.
	Compensatable bool `json:"compensatable"`

	// RetryCount tracks how many times this step has been retried.
	RetryCount int `json:"retry_count"`

	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int `json:"max_retries"`
}

// NewSagaStep creates a new saga step.
func NewSagaStep(stepType SagaStepType, order int, compensatable bool) *SagaStep {
	return &SagaStep{
		ID:            uuid.New(),
		Type:          stepType,
		Order:         order,
		Status:        StepStatusPending,
		Input:         make(map[string]interface{}),
		Output:        make(map[string]interface{}),
		Compensatable: compensatable,
		RetryCount:    0,
		MaxRetries:    3,
	}
}

// Start marks the step as running.
func (s *SagaStep) Start() {
	now := time.Now().UTC()
	s.Status = StepStatusRunning
	s.StartedAt = &now
}

// Complete marks the step as successfully completed.
func (s *SagaStep) Complete(output map[string]interface{}) {
	now := time.Now().UTC()
	s.Status = StepStatusCompleted
	s.CompletedAt = &now
	if output != nil {
		s.Output = output
	}
}

// Fail marks the step as failed.
func (s *SagaStep) Fail(err error) {
	now := time.Now().UTC()
	s.Status = StepStatusFailed
	s.CompletedAt = &now
	errStr := err.Error()
	s.Error = &errStr
}

// Compensate marks the step as compensated.
func (s *SagaStep) Compensate() {
	now := time.Now().UTC()
	s.Status = StepStatusCompensated
	s.CompensatedAt = &now
}

// Skip marks the step as skipped.
func (s *SagaStep) Skip() {
	s.Status = StepStatusSkipped
}

// CanRetry returns true if the step can be retried.
func (s *SagaStep) CanRetry() bool {
	return s.Status == StepStatusFailed && s.RetryCount < s.MaxRetries
}

// IncrementRetry increments the retry counter.
func (s *SagaStep) IncrementRetry() {
	s.RetryCount++
}

// ============================================================================
// Lead Conversion Saga
// ============================================================================

// LeadConversionSaga tracks the complete lead-to-opportunity conversion workflow.
type LeadConversionSaga struct {
	// ID is the unique identifier for this saga instance.
	ID uuid.UUID `json:"id"`

	// TenantID is the tenant this saga belongs to.
	TenantID uuid.UUID `json:"tenant_id"`

	// LeadID is the lead being converted.
	LeadID uuid.UUID `json:"lead_id"`

	// IdempotencyKey prevents duplicate conversions.
	IdempotencyKey string `json:"idempotency_key"`

	// State is the current saga execution state.
	State SagaState `json:"state"`

	// CurrentStepIndex tracks which step is currently executing.
	CurrentStepIndex int `json:"current_step_index"`

	// Steps contains all steps in execution order.
	Steps []*SagaStep `json:"steps"`

	// Request contains the original conversion request data.
	Request *LeadConversionRequest `json:"request"`

	// Result contains the conversion result if successful.
	Result *LeadConversionResult `json:"result,omitempty"`

	// OpportunityID is the created opportunity if conversion succeeded.
	OpportunityID *uuid.UUID `json:"opportunity_id,omitempty"`

	// CustomerID is the created or linked customer.
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`

	// ContactID is the created or linked contact.
	ContactID *uuid.UUID `json:"contact_id,omitempty"`

	// CustomerCreated indicates if a new customer was created (for compensation).
	CustomerCreated bool `json:"customer_created"`

	// Error contains the error message if the saga failed.
	Error *string `json:"error,omitempty"`

	// ErrorCode contains a structured error code if available.
	ErrorCode *string `json:"error_code,omitempty"`

	// FailedStepType identifies which step caused the failure.
	FailedStepType *SagaStepType `json:"failed_step_type,omitempty"`

	// InitiatedBy is the user who initiated the conversion.
	InitiatedBy uuid.UUID `json:"initiated_by"`

	// StartedAt is when the saga began.
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when the saga finished (success, failure, or compensation).
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Version is used for optimistic locking.
	Version int `json:"version"`

	// Metadata contains additional context data.
	Metadata map[string]string `json:"metadata,omitempty"`

	// CreatedAt is the creation timestamp.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt time.Time `json:"updated_at"`

	// Events accumulated during saga execution.
	events []DomainEvent
}

// LeadConversionRequest contains all data needed to execute the conversion.
type LeadConversionRequest struct {
	LeadID            uuid.UUID  `json:"lead_id"`
	PipelineID        uuid.UUID  `json:"pipeline_id"`
	CustomerID        *uuid.UUID `json:"customer_id,omitempty"`
	CustomerName      *string    `json:"customer_name,omitempty"`
	CreateNewCustomer bool       `json:"create_new_customer"`
	OwnerID           *uuid.UUID `json:"owner_id,omitempty"`
	OwnerName         *string    `json:"owner_name,omitempty"`
	Description       *string    `json:"description,omitempty"`
	ExpectedCloseDate *time.Time `json:"expected_close_date,omitempty"`
	Amount            *Money     `json:"amount,omitempty"`
	Probability       *int       `json:"probability,omitempty"`
}

// LeadConversionResult contains the successful conversion output.
type LeadConversionResult struct {
	LeadID           uuid.UUID  `json:"lead_id"`
	OpportunityID    uuid.UUID  `json:"opportunity_id"`
	OpportunityCode  string     `json:"opportunity_code"`
	CustomerID       *uuid.UUID `json:"customer_id,omitempty"`
	CustomerCode     *string    `json:"customer_code,omitempty"`
	ContactID        *uuid.UUID `json:"contact_id,omitempty"`
	ConvertedAt      time.Time  `json:"converted_at"`
	ConvertedBy      uuid.UUID  `json:"converted_by"`
	CustomerCreated  bool       `json:"customer_created"`
	ConversionTimeMs int64      `json:"conversion_time_ms"`
}

// NewLeadConversionSaga creates a new lead conversion saga.
func NewLeadConversionSaga(
	tenantID uuid.UUID,
	leadID uuid.UUID,
	idempotencyKey string,
	initiatedBy uuid.UUID,
	request *LeadConversionRequest,
) *LeadConversionSaga {
	now := time.Now().UTC()
	saga := &LeadConversionSaga{
		ID:               uuid.New(),
		TenantID:         tenantID,
		LeadID:           leadID,
		IdempotencyKey:   idempotencyKey,
		State:            SagaStateStarted,
		CurrentStepIndex: 0,
		Steps:            make([]*SagaStep, 0),
		Request:          request,
		InitiatedBy:      initiatedBy,
		StartedAt:        now,
		Version:          1,
		Metadata:         make(map[string]string),
		CreatedAt:        now,
		UpdatedAt:        now,
		events:           make([]DomainEvent, 0),
	}

	// Build the step workflow based on request
	saga.buildSteps(request)

	// Raise saga started event
	saga.addEvent(NewLeadConversionSagaStartedEvent(saga))

	return saga
}

// buildSteps constructs the workflow steps based on the conversion request.
func (s *LeadConversionSaga) buildSteps(request *LeadConversionRequest) {
	order := 1

	// Step 1: Validate lead (always required, not compensatable)
	s.Steps = append(s.Steps, NewSagaStep(StepTypeValidateLead, order, false))
	order++

	// Step 2: Customer handling (depends on request)
	if request.CreateNewCustomer {
		// Create new customer (compensatable - can delete)
		s.Steps = append(s.Steps, NewSagaStep(StepTypeCreateCustomer, order, true))
	} else if request.CustomerID != nil {
		// Lookup existing customer (not compensatable)
		s.Steps = append(s.Steps, NewSagaStep(StepTypeLookupCustomer, order, false))
	}
	order++

	// Step 3: Create opportunity (compensatable - can delete)
	s.Steps = append(s.Steps, NewSagaStep(StepTypeCreateOpportunity, order, true))
	order++

	// Step 4: Mark lead as converted (compensatable - can revert status)
	s.Steps = append(s.Steps, NewSagaStep(StepTypeMarkLeadConverted, order, true))
	order++

	// Step 5: Publish events via outbox (compensatable - can remove from outbox)
	s.Steps = append(s.Steps, NewSagaStep(StepTypePublishEvents, order, true))
}

// ============================================================================
// Saga State Transitions
// ============================================================================

// Start transitions the saga to running state.
func (s *LeadConversionSaga) Start() error {
	if !s.State.CanTransitionTo(SagaStateRunning) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidSagaTransition, s.State, SagaStateRunning)
	}
	s.State = SagaStateRunning
	s.UpdatedAt = time.Now().UTC()
	return nil
}

// Complete transitions the saga to completed state.
func (s *LeadConversionSaga) Complete(result *LeadConversionResult) error {
	if !s.State.CanTransitionTo(SagaStateCompleted) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidSagaTransition, s.State, SagaStateCompleted)
	}

	now := time.Now().UTC()
	s.State = SagaStateCompleted
	s.Result = result
	s.CompletedAt = &now
	s.UpdatedAt = now

	// Calculate conversion time
	if result != nil {
		result.ConversionTimeMs = now.Sub(s.StartedAt).Milliseconds()
	}

	s.addEvent(NewLeadConversionSagaCompletedEvent(s))
	return nil
}

// StartCompensation transitions the saga to compensating state.
func (s *LeadConversionSaga) StartCompensation(err error, failedStepType SagaStepType) error {
	if !s.State.CanTransitionTo(SagaStateCompensating) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidSagaTransition, s.State, SagaStateCompensating)
	}

	s.State = SagaStateCompensating
	errStr := err.Error()
	s.Error = &errStr
	s.FailedStepType = &failedStepType
	s.UpdatedAt = time.Now().UTC()

	s.addEvent(NewLeadConversionSagaCompensatingEvent(s, err.Error()))
	return nil
}

// Compensated transitions the saga to compensated state.
func (s *LeadConversionSaga) Compensated() error {
	if !s.State.CanTransitionTo(SagaStateCompensated) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidSagaTransition, s.State, SagaStateCompensated)
	}

	now := time.Now().UTC()
	s.State = SagaStateCompensated
	s.CompletedAt = &now
	s.UpdatedAt = now

	s.addEvent(NewLeadConversionSagaCompensatedEvent(s))
	return nil
}

// Fail transitions the saga to failed state (unrecoverable).
func (s *LeadConversionSaga) Fail(err error, failedStepType *SagaStepType) error {
	// Can fail from multiple states
	if s.State.IsTerminal() {
		return fmt.Errorf("%w: saga is already in terminal state %s", ErrInvalidSagaTransition, s.State)
	}

	now := time.Now().UTC()
	s.State = SagaStateFailed
	errStr := err.Error()
	s.Error = &errStr
	s.FailedStepType = failedStepType
	s.CompletedAt = &now
	s.UpdatedAt = now

	s.addEvent(NewLeadConversionSagaFailedEvent(s, err.Error()))
	return nil
}

// ============================================================================
// Step Management
// ============================================================================

// GetCurrentStep returns the currently executing step.
func (s *LeadConversionSaga) GetCurrentStep() *SagaStep {
	if s.CurrentStepIndex >= 0 && s.CurrentStepIndex < len(s.Steps) {
		return s.Steps[s.CurrentStepIndex]
	}
	return nil
}

// GetStepByType returns the step of the given type.
func (s *LeadConversionSaga) GetStepByType(stepType SagaStepType) *SagaStep {
	for _, step := range s.Steps {
		if step.Type == stepType {
			return step
		}
	}
	return nil
}

// AdvanceToNextStep moves to the next step in the sequence.
func (s *LeadConversionSaga) AdvanceToNextStep() bool {
	s.CurrentStepIndex++
	s.UpdatedAt = time.Now().UTC()
	return s.CurrentStepIndex < len(s.Steps)
}

// HasMoreSteps returns true if there are more steps to execute.
func (s *LeadConversionSaga) HasMoreSteps() bool {
	return s.CurrentStepIndex < len(s.Steps)
}

// GetCompletedCompensatableSteps returns completed steps that can be compensated, in reverse order.
func (s *LeadConversionSaga) GetCompletedCompensatableSteps() []*SagaStep {
	var steps []*SagaStep
	// Iterate in reverse order for compensation
	for i := len(s.Steps) - 1; i >= 0; i-- {
		step := s.Steps[i]
		if step.Status == StepStatusCompleted && step.Compensatable {
			steps = append(steps, step)
		}
	}
	return steps
}

// SetOpportunityID records the created opportunity.
func (s *LeadConversionSaga) SetOpportunityID(id uuid.UUID) {
	s.OpportunityID = &id
	s.UpdatedAt = time.Now().UTC()
}

// SetCustomerID records the customer (created or linked).
func (s *LeadConversionSaga) SetCustomerID(id uuid.UUID, created bool) {
	s.CustomerID = &id
	s.CustomerCreated = created
	s.UpdatedAt = time.Now().UTC()
}

// SetContactID records the contact.
func (s *LeadConversionSaga) SetContactID(id uuid.UUID) {
	s.ContactID = &id
	s.UpdatedAt = time.Now().UTC()
}

// ============================================================================
// Domain Events
// ============================================================================

func (s *LeadConversionSaga) addEvent(event DomainEvent) {
	s.events = append(s.events, event)
}

// GetEvents returns all accumulated domain events.
func (s *LeadConversionSaga) GetEvents() []DomainEvent {
	return s.events
}

// ClearEvents clears the accumulated events.
func (s *LeadConversionSaga) ClearEvents() {
	s.events = make([]DomainEvent, 0)
}

// ============================================================================
// Serialization Helpers
// ============================================================================

// StepsJSON returns the steps as JSON bytes for database storage.
func (s *LeadConversionSaga) StepsJSON() ([]byte, error) {
	return json.Marshal(s.Steps)
}

// SetStepsFromJSON populates steps from JSON bytes.
func (s *LeadConversionSaga) SetStepsFromJSON(data []byte) error {
	return json.Unmarshal(data, &s.Steps)
}

// RequestJSON returns the request as JSON bytes.
func (s *LeadConversionSaga) RequestJSON() ([]byte, error) {
	return json.Marshal(s.Request)
}

// SetRequestFromJSON populates request from JSON bytes.
func (s *LeadConversionSaga) SetRequestFromJSON(data []byte) error {
	return json.Unmarshal(data, &s.Request)
}

// ResultJSON returns the result as JSON bytes.
func (s *LeadConversionSaga) ResultJSON() ([]byte, error) {
	if s.Result == nil {
		return nil, nil
	}
	return json.Marshal(s.Result)
}

// SetResultFromJSON populates result from JSON bytes.
func (s *LeadConversionSaga) SetResultFromJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	s.Result = &LeadConversionResult{}
	return json.Unmarshal(data, s.Result)
}

// MetadataJSON returns metadata as JSON bytes.
func (s *LeadConversionSaga) MetadataJSON() ([]byte, error) {
	return json.Marshal(s.Metadata)
}

// SetMetadataFromJSON populates metadata from JSON bytes.
func (s *LeadConversionSaga) SetMetadataFromJSON(data []byte) error {
	if len(data) == 0 {
		s.Metadata = make(map[string]string)
		return nil
	}
	return json.Unmarshal(data, &s.Metadata)
}

// ============================================================================
// Saga Summary
// ============================================================================

// Summary returns a summary of the saga for logging/debugging.
func (s *LeadConversionSaga) Summary() string {
	completedSteps := 0
	failedSteps := 0
	for _, step := range s.Steps {
		switch step.Status {
		case StepStatusCompleted:
			completedSteps++
		case StepStatusFailed:
			failedSteps++
		}
	}

	return fmt.Sprintf(
		"Saga[%s] Lead[%s] State[%s] Steps[%d/%d completed, %d failed]",
		s.ID.String()[:8],
		s.LeadID.String()[:8],
		s.State,
		completedSteps,
		len(s.Steps),
		failedSteps,
	)
}

// ============================================================================
// Idempotency Key
// ============================================================================

// IdempotencyStatus represents the processing status of an idempotency key.
type IdempotencyStatus string

const (
	IdempotencyStatusPending    IdempotencyStatus = "pending"
	IdempotencyStatusProcessing IdempotencyStatus = "processing"
	IdempotencyStatusCompleted  IdempotencyStatus = "completed"
	IdempotencyStatusFailed     IdempotencyStatus = "failed"
)

// IdempotencyKey represents a stored idempotency key for duplicate detection.
type IdempotencyKey struct {
	Key          string            `json:"key"`
	TenantID     uuid.UUID         `json:"tenant_id"`
	ResourceID   uuid.UUID         `json:"resource_id,omitempty"`
	Status       IdempotencyStatus `json:"status"`
	ResponseCode *int              `json:"response_code,omitempty"`
	ResponseBody []byte            `json:"response_body,omitempty"`
	ExpiresAt    time.Time         `json:"expires_at"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// NewIdempotencyKey creates a new idempotency key entry.
func NewIdempotencyKey(tenantID uuid.UUID, key string, ttl time.Duration) *IdempotencyKey {
	now := time.Now().UTC()
	return &IdempotencyKey{
		Key:       key,
		TenantID:  tenantID,
		Status:    IdempotencyStatusPending,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewIdempotencyKeyWithResource creates a new idempotency key with resource ID.
func NewIdempotencyKeyWithResource(key string, tenantID, resourceID uuid.UUID, ttl time.Duration) *IdempotencyKey {
	idemKey := NewIdempotencyKey(tenantID, key, ttl)
	idemKey.ResourceID = resourceID
	return idemKey
}

// IsExpired returns true if the idempotency key has expired.
func (k *IdempotencyKey) IsExpired() bool {
	return time.Now().UTC().After(k.ExpiresAt)
}

// MarkProcessing marks the key as currently being processed.
func (k *IdempotencyKey) MarkProcessing() {
	k.Status = IdempotencyStatusProcessing
	k.UpdatedAt = time.Now().UTC()
}

// MarkCompleted marks the key as completed with the response.
func (k *IdempotencyKey) MarkCompleted(responseCode *int, responseBody []byte) {
	k.Status = IdempotencyStatusCompleted
	k.ResponseCode = responseCode
	k.ResponseBody = responseBody
	k.UpdatedAt = time.Now().UTC()
}

// MarkFailed marks the key as failed with the response.
func (k *IdempotencyKey) MarkFailed(responseCode *int, responseBody []byte) {
	k.Status = IdempotencyStatusFailed
	k.ResponseCode = responseCode
	k.ResponseBody = responseBody
	k.UpdatedAt = time.Now().UTC()
}

// SetResourceID sets the resource ID created by this request.
func (k *IdempotencyKey) SetResourceID(resourceID uuid.UUID) {
	k.ResourceID = resourceID
	k.UpdatedAt = time.Now().UTC()
}

// GenerateIdempotencyKey generates a deterministic idempotency key for lead conversion.
func GenerateIdempotencyKey(tenantID, leadID, userID uuid.UUID) string {
	// Use a combination that ensures uniqueness per conversion attempt
	return fmt.Sprintf("lead_conversion:%s:%s:%s:%d",
		tenantID.String(),
		leadID.String(),
		userID.String(),
		time.Now().UTC().Unix()/300, // 5-minute window
	)
}
