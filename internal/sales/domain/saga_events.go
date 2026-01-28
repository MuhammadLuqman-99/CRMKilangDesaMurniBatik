// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Saga Events
// ============================================================================

// LeadConversionSagaStartedEvent is raised when a lead conversion saga begins.
type LeadConversionSagaStartedEvent struct {
	BaseEvent
	SagaID         uuid.UUID `json:"saga_id"`
	LeadID         uuid.UUID `json:"lead_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	InitiatedBy    uuid.UUID `json:"initiated_by"`
	StepCount      int       `json:"step_count"`
}

// NewLeadConversionSagaStartedEvent creates a new saga started event.
func NewLeadConversionSagaStartedEvent(saga *LeadConversionSaga) *LeadConversionSagaStartedEvent {
	return &LeadConversionSagaStartedEvent{
		BaseEvent:      newBaseEvent("saga.lead_conversion.started", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:         saga.ID,
		LeadID:         saga.LeadID,
		IdempotencyKey: saga.IdempotencyKey,
		InitiatedBy:    saga.InitiatedBy,
		StepCount:      len(saga.Steps),
	}
}

// LeadConversionSagaStepCompletedEvent is raised when a saga step completes.
type LeadConversionSagaStepCompletedEvent struct {
	BaseEvent
	SagaID     uuid.UUID    `json:"saga_id"`
	LeadID     uuid.UUID    `json:"lead_id"`
	StepType   SagaStepType `json:"step_type"`
	StepOrder  int          `json:"step_order"`
	DurationMs int64        `json:"duration_ms"`
	OutputKeys []string     `json:"output_keys,omitempty"`
}

// NewLeadConversionSagaStepCompletedEvent creates a new step completed event.
func NewLeadConversionSagaStepCompletedEvent(saga *LeadConversionSaga, step *SagaStep) *LeadConversionSagaStepCompletedEvent {
	var durationMs int64
	if step.StartedAt != nil && step.CompletedAt != nil {
		durationMs = step.CompletedAt.Sub(*step.StartedAt).Milliseconds()
	}

	var outputKeys []string
	for k := range step.Output {
		outputKeys = append(outputKeys, k)
	}

	return &LeadConversionSagaStepCompletedEvent{
		BaseEvent:  newBaseEvent("saga.lead_conversion.step_completed", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:     saga.ID,
		LeadID:     saga.LeadID,
		StepType:   step.Type,
		StepOrder:  step.Order,
		DurationMs: durationMs,
		OutputKeys: outputKeys,
	}
}

// LeadConversionSagaStepFailedEvent is raised when a saga step fails.
type LeadConversionSagaStepFailedEvent struct {
	BaseEvent
	SagaID     uuid.UUID    `json:"saga_id"`
	LeadID     uuid.UUID    `json:"lead_id"`
	StepType   SagaStepType `json:"step_type"`
	StepOrder  int          `json:"step_order"`
	Error      string       `json:"error"`
	RetryCount int          `json:"retry_count"`
	CanRetry   bool         `json:"can_retry"`
}

// NewLeadConversionSagaStepFailedEvent creates a new step failed event.
func NewLeadConversionSagaStepFailedEvent(saga *LeadConversionSaga, step *SagaStep, err error) *LeadConversionSagaStepFailedEvent {
	return &LeadConversionSagaStepFailedEvent{
		BaseEvent:  newBaseEvent("saga.lead_conversion.step_failed", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:     saga.ID,
		LeadID:     saga.LeadID,
		StepType:   step.Type,
		StepOrder:  step.Order,
		Error:      err.Error(),
		RetryCount: step.RetryCount,
		CanRetry:   step.CanRetry(),
	}
}

// LeadConversionSagaCompletedEvent is raised when a saga completes successfully.
type LeadConversionSagaCompletedEvent struct {
	BaseEvent
	SagaID          uuid.UUID  `json:"saga_id"`
	LeadID          uuid.UUID  `json:"lead_id"`
	OpportunityID   uuid.UUID  `json:"opportunity_id"`
	CustomerID      *uuid.UUID `json:"customer_id,omitempty"`
	CustomerCreated bool       `json:"customer_created"`
	TotalDurationMs int64      `json:"total_duration_ms"`
	StepsCompleted  int        `json:"steps_completed"`
}

// NewLeadConversionSagaCompletedEvent creates a new saga completed event.
func NewLeadConversionSagaCompletedEvent(saga *LeadConversionSaga) *LeadConversionSagaCompletedEvent {
	var durationMs int64
	if saga.CompletedAt != nil {
		durationMs = saga.CompletedAt.Sub(saga.StartedAt).Milliseconds()
	}

	stepsCompleted := 0
	for _, step := range saga.Steps {
		if step.Status == StepStatusCompleted {
			stepsCompleted++
		}
	}

	return &LeadConversionSagaCompletedEvent{
		BaseEvent:       newBaseEvent("saga.lead_conversion.completed", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:          saga.ID,
		LeadID:          saga.LeadID,
		OpportunityID:   *saga.OpportunityID,
		CustomerID:      saga.CustomerID,
		CustomerCreated: saga.CustomerCreated,
		TotalDurationMs: durationMs,
		StepsCompleted:  stepsCompleted,
	}
}

// LeadConversionSagaCompensatingEvent is raised when a saga starts compensating.
type LeadConversionSagaCompensatingEvent struct {
	BaseEvent
	SagaID            uuid.UUID    `json:"saga_id"`
	LeadID            uuid.UUID    `json:"lead_id"`
	FailedStepType    SagaStepType `json:"failed_step_type"`
	Error             string       `json:"error"`
	StepsToCompensate int          `json:"steps_to_compensate"`
}

// NewLeadConversionSagaCompensatingEvent creates a new compensating event.
func NewLeadConversionSagaCompensatingEvent(saga *LeadConversionSaga, err string) *LeadConversionSagaCompensatingEvent {
	stepsToCompensate := 0
	for _, step := range saga.Steps {
		if step.Status == StepStatusCompleted && step.Compensatable {
			stepsToCompensate++
		}
	}

	var failedStepType SagaStepType
	if saga.FailedStepType != nil {
		failedStepType = *saga.FailedStepType
	}

	return &LeadConversionSagaCompensatingEvent{
		BaseEvent:         newBaseEvent("saga.lead_conversion.compensating", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:            saga.ID,
		LeadID:            saga.LeadID,
		FailedStepType:    failedStepType,
		Error:             err,
		StepsToCompensate: stepsToCompensate,
	}
}

// LeadConversionSagaStepCompensatedEvent is raised when a step is compensated.
type LeadConversionSagaStepCompensatedEvent struct {
	BaseEvent
	SagaID   uuid.UUID    `json:"saga_id"`
	LeadID   uuid.UUID    `json:"lead_id"`
	StepType SagaStepType `json:"step_type"`
}

// NewLeadConversionSagaStepCompensatedEvent creates a new step compensated event.
func NewLeadConversionSagaStepCompensatedEvent(saga *LeadConversionSaga, step *SagaStep) *LeadConversionSagaStepCompensatedEvent {
	return &LeadConversionSagaStepCompensatedEvent{
		BaseEvent: newBaseEvent("saga.lead_conversion.step_compensated", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:    saga.ID,
		LeadID:    saga.LeadID,
		StepType:  step.Type,
	}
}

// LeadConversionSagaCompensatedEvent is raised when all compensation is complete.
type LeadConversionSagaCompensatedEvent struct {
	BaseEvent
	SagaID           uuid.UUID `json:"saga_id"`
	LeadID           uuid.UUID `json:"lead_id"`
	OriginalError    string    `json:"original_error"`
	StepsCompensated int       `json:"steps_compensated"`
	TotalDurationMs  int64     `json:"total_duration_ms"`
}

// NewLeadConversionSagaCompensatedEvent creates a new saga compensated event.
func NewLeadConversionSagaCompensatedEvent(saga *LeadConversionSaga) *LeadConversionSagaCompensatedEvent {
	var durationMs int64
	if saga.CompletedAt != nil {
		durationMs = saga.CompletedAt.Sub(saga.StartedAt).Milliseconds()
	}

	stepsCompensated := 0
	for _, step := range saga.Steps {
		if step.Status == StepStatusCompensated {
			stepsCompensated++
		}
	}

	var originalError string
	if saga.Error != nil {
		originalError = *saga.Error
	}

	return &LeadConversionSagaCompensatedEvent{
		BaseEvent:        newBaseEvent("saga.lead_conversion.compensated", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:           saga.ID,
		LeadID:           saga.LeadID,
		OriginalError:    originalError,
		StepsCompensated: stepsCompensated,
		TotalDurationMs:  durationMs,
	}
}

// LeadConversionSagaFailedEvent is raised when a saga fails unrecoverably.
type LeadConversionSagaFailedEvent struct {
	BaseEvent
	SagaID          uuid.UUID     `json:"saga_id"`
	LeadID          uuid.UUID     `json:"lead_id"`
	Error           string        `json:"error"`
	FailedStepType  *SagaStepType `json:"failed_step_type,omitempty"`
	TotalDurationMs int64         `json:"total_duration_ms"`
	StepsCompleted  int           `json:"steps_completed"`
	StepsFailed     int           `json:"steps_failed"`
}

// NewLeadConversionSagaFailedEvent creates a new saga failed event.
func NewLeadConversionSagaFailedEvent(saga *LeadConversionSaga, err string) *LeadConversionSagaFailedEvent {
	var durationMs int64
	if saga.CompletedAt != nil {
		durationMs = saga.CompletedAt.Sub(saga.StartedAt).Milliseconds()
	}

	stepsCompleted := 0
	stepsFailed := 0
	for _, step := range saga.Steps {
		switch step.Status {
		case StepStatusCompleted:
			stepsCompleted++
		case StepStatusFailed:
			stepsFailed++
		}
	}

	return &LeadConversionSagaFailedEvent{
		BaseEvent:       newBaseEvent("saga.lead_conversion.failed", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:          saga.ID,
		LeadID:          saga.LeadID,
		Error:           err,
		FailedStepType:  saga.FailedStepType,
		TotalDurationMs: durationMs,
		StepsCompleted:  stepsCompleted,
		StepsFailed:     stepsFailed,
	}
}

// LeadConversionSagaRetryingEvent is raised when a saga step is being retried.
type LeadConversionSagaRetryingEvent struct {
	BaseEvent
	SagaID      uuid.UUID    `json:"saga_id"`
	LeadID      uuid.UUID    `json:"lead_id"`
	StepType    SagaStepType `json:"step_type"`
	RetryCount  int          `json:"retry_count"`
	MaxRetries  int          `json:"max_retries"`
	LastError   string       `json:"last_error"`
	NextRetryAt time.Time    `json:"next_retry_at"`
}

// NewLeadConversionSagaRetryingEvent creates a new retry event.
func NewLeadConversionSagaRetryingEvent(saga *LeadConversionSaga, step *SagaStep, nextRetryAt time.Time) *LeadConversionSagaRetryingEvent {
	var lastError string
	if step.Error != nil {
		lastError = *step.Error
	}

	return &LeadConversionSagaRetryingEvent{
		BaseEvent:   newBaseEvent("saga.lead_conversion.retrying", "saga", saga.ID, saga.TenantID, saga.Version),
		SagaID:      saga.ID,
		LeadID:      saga.LeadID,
		StepType:    step.Type,
		RetryCount:  step.RetryCount,
		MaxRetries:  step.MaxRetries,
		LastError:   lastError,
		NextRetryAt: nextRetryAt,
	}
}
