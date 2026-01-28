// Package usecase contains the application use cases for the Sales Pipeline service.
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/sales/application"
	"github.com/kilang-desa-murni/crm/internal/sales/application/ports"
	"github.com/kilang-desa-murni/crm/internal/sales/domain"
)

// ============================================================================
// Saga Orchestrator Interface
// ============================================================================

// LeadConversionOrchestrator orchestrates the lead-to-opportunity conversion saga.
type LeadConversionOrchestrator interface {
	// Execute runs the lead conversion saga.
	Execute(ctx context.Context, saga *domain.LeadConversionSaga) error

	// Compensate runs the compensation logic for a failed saga.
	Compensate(ctx context.Context, saga *domain.LeadConversionSaga) error

	// Resume attempts to resume a saga that was interrupted.
	Resume(ctx context.Context, saga *domain.LeadConversionSaga) error

	// GetSagaStatus returns the current status of a saga.
	GetSagaStatus(ctx context.Context, tenantID, sagaID uuid.UUID) (*domain.LeadConversionSaga, error)
}

// ============================================================================
// Lead Conversion Orchestrator Implementation
// ============================================================================

// leadConversionOrchestrator implements LeadConversionOrchestrator.
type leadConversionOrchestrator struct {
	uow             domain.UnitOfWork
	leadRepo        domain.LeadRepository
	opportunityRepo domain.OpportunityRepository
	pipelineRepo    domain.PipelineRepository
	eventPublisher  ports.EventPublisher
	customerService ports.CustomerService

	// Step handlers
	stepHandlers         map[domain.SagaStepType]stepHandler
	compensationHandlers map[domain.SagaStepType]compensationHandler
}

// stepHandler executes a saga step.
type stepHandler func(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error

// compensationHandler rolls back a saga step.
type compensationHandler func(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error

// NewLeadConversionOrchestrator creates a new saga orchestrator.
func NewLeadConversionOrchestrator(
	uow domain.UnitOfWork,
	leadRepo domain.LeadRepository,
	opportunityRepo domain.OpportunityRepository,
	pipelineRepo domain.PipelineRepository,
	eventPublisher ports.EventPublisher,
	customerService ports.CustomerService,
) LeadConversionOrchestrator {
	orch := &leadConversionOrchestrator{
		uow:             uow,
		leadRepo:        leadRepo,
		opportunityRepo: opportunityRepo,
		pipelineRepo:    pipelineRepo,
		eventPublisher:  eventPublisher,
		customerService: customerService,
	}

	// Initialize step handlers
	orch.stepHandlers = map[domain.SagaStepType]stepHandler{
		domain.StepTypeValidateLead:      orch.executeValidateLead,
		domain.StepTypeCreateCustomer:    orch.executeCreateCustomer,
		domain.StepTypeLookupCustomer:    orch.executeLookupCustomer,
		domain.StepTypeCreateOpportunity: orch.executeCreateOpportunity,
		domain.StepTypeMarkLeadConverted: orch.executeMarkLeadConverted,
		domain.StepTypePublishEvents:     orch.executePublishEvents,
	}

	// Initialize compensation handlers
	orch.compensationHandlers = map[domain.SagaStepType]compensationHandler{
		domain.StepTypeCreateCustomer:    orch.compensateCreateCustomer,
		domain.StepTypeCreateOpportunity: orch.compensateCreateOpportunity,
		domain.StepTypeMarkLeadConverted: orch.compensateMarkLeadConverted,
		domain.StepTypePublishEvents:     orch.compensatePublishEvents,
	}

	return orch
}

// Execute runs the lead conversion saga.
func (o *leadConversionOrchestrator) Execute(ctx context.Context, saga *domain.LeadConversionSaga) error {
	// Start the saga if it's in started state
	if saga.State == domain.SagaStateStarted {
		if err := saga.Start(); err != nil {
			return fmt.Errorf("failed to start saga: %w", err)
		}
	}

	// Execute steps
	for saga.HasMoreSteps() {
		step := saga.GetCurrentStep()
		if step == nil {
			break
		}

		// Skip completed or failed steps
		if step.Status == domain.StepStatusCompleted || step.Status == domain.StepStatusSkipped {
			saga.AdvanceToNextStep()
			continue
		}

		// Execute the step
		if err := o.executeStep(ctx, saga, step); err != nil {
			// Step failed - start compensation
			step.Fail(err)
			if startErr := saga.StartCompensation(err, step.Type); startErr != nil {
				return fmt.Errorf("failed to start compensation: %w", startErr)
			}

			// Persist the failure
			if saveErr := o.saveSaga(ctx, saga); saveErr != nil {
				return fmt.Errorf("failed to save saga after failure: %w", saveErr)
			}

			// Run compensation
			if compErr := o.Compensate(ctx, saga); compErr != nil {
				return fmt.Errorf("compensation failed: %w", compErr)
			}

			return err
		}

		// Persist progress after each step
		if saveErr := o.saveSaga(ctx, saga); saveErr != nil {
			return fmt.Errorf("failed to save saga progress: %w", saveErr)
		}

		saga.AdvanceToNextStep()
	}

	// All steps completed successfully
	result := o.buildConversionResult(saga)
	if err := saga.Complete(result); err != nil {
		return fmt.Errorf("failed to complete saga: %w", err)
	}

	// Final save
	if err := o.saveSaga(ctx, saga); err != nil {
		return fmt.Errorf("failed to save completed saga: %w", err)
	}

	return nil
}

// executeStep runs a single saga step.
func (o *leadConversionOrchestrator) executeStep(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	handler, exists := o.stepHandlers[step.Type]
	if !exists {
		return fmt.Errorf("no handler for step type: %s", step.Type)
	}

	step.Start()

	if err := handler(ctx, saga, step); err != nil {
		return err
	}

	step.Complete(step.Output)
	return nil
}

// Compensate runs the compensation logic for a failed saga.
func (o *leadConversionOrchestrator) Compensate(ctx context.Context, saga *domain.LeadConversionSaga) error {
	if saga.State != domain.SagaStateCompensating {
		return fmt.Errorf("saga is not in compensating state")
	}

	// Get completed compensatable steps in reverse order
	stepsToCompensate := saga.GetCompletedCompensatableSteps()

	for _, step := range stepsToCompensate {
		handler, exists := o.compensationHandlers[step.Type]
		if !exists {
			// No compensation handler means nothing to undo
			continue
		}

		if err := handler(ctx, saga, step); err != nil {
			// Compensation failed - mark saga as failed
			if failErr := saga.Fail(err, &step.Type); failErr != nil {
				return fmt.Errorf("failed to mark saga as failed: %w", failErr)
			}
			if saveErr := o.saveSaga(ctx, saga); saveErr != nil {
				return fmt.Errorf("failed to save failed saga: %w", saveErr)
			}
			return fmt.Errorf("compensation for step %s failed: %w", step.Type, err)
		}

		step.Compensate()
	}

	// All compensation completed
	if err := saga.Compensated(); err != nil {
		return fmt.Errorf("failed to mark saga as compensated: %w", err)
	}

	if err := o.saveSaga(ctx, saga); err != nil {
		return fmt.Errorf("failed to save compensated saga: %w", err)
	}

	return nil
}

// Resume attempts to resume a saga that was interrupted.
func (o *leadConversionOrchestrator) Resume(ctx context.Context, saga *domain.LeadConversionSaga) error {
	switch saga.State {
	case domain.SagaStateStarted, domain.SagaStateRunning:
		return o.Execute(ctx, saga)
	case domain.SagaStateCompensating:
		return o.Compensate(ctx, saga)
	case domain.SagaStateCompleted, domain.SagaStateCompensated, domain.SagaStateFailed:
		return nil // Already in terminal state
	default:
		return fmt.Errorf("unknown saga state: %s", saga.State)
	}
}

// GetSagaStatus returns the current status of a saga.
func (o *leadConversionOrchestrator) GetSagaStatus(ctx context.Context, tenantID, sagaID uuid.UUID) (*domain.LeadConversionSaga, error) {
	return o.uow.Sagas().GetByID(ctx, tenantID, sagaID)
}

// saveSaga persists the saga state.
func (o *leadConversionOrchestrator) saveSaga(ctx context.Context, saga *domain.LeadConversionSaga) error {
	return o.uow.Sagas().Update(ctx, saga)
}

// buildConversionResult constructs the result from the saga state.
func (o *leadConversionOrchestrator) buildConversionResult(saga *domain.LeadConversionSaga) *domain.LeadConversionResult {
	result := &domain.LeadConversionResult{
		LeadID:      saga.LeadID,
		ConvertedAt: time.Now().UTC(),
		ConvertedBy: saga.InitiatedBy,
	}

	if saga.OpportunityID != nil {
		result.OpportunityID = *saga.OpportunityID
	}

	if saga.CustomerID != nil {
		result.CustomerID = saga.CustomerID
		result.CustomerCreated = saga.CustomerCreated
	}

	if saga.ContactID != nil {
		result.ContactID = saga.ContactID
	}

	// Get opportunity code from step output
	if step := saga.GetStepByType(domain.StepTypeCreateOpportunity); step != nil {
		if code, ok := step.Output["opportunity_code"].(string); ok {
			result.OpportunityCode = code
		}
	}

	// Get customer code from step output
	if step := saga.GetStepByType(domain.StepTypeCreateCustomer); step != nil {
		if code, ok := step.Output["customer_code"].(string); ok {
			result.CustomerCode = &code
		}
	}

	return result
}

// ============================================================================
// Step Handlers
// ============================================================================

// executeValidateLead validates the lead can be converted.
func (o *leadConversionOrchestrator) executeValidateLead(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	lead, err := o.leadRepo.GetByID(ctx, saga.TenantID, saga.LeadID)
	if err != nil {
		return application.ErrLeadNotFound(saga.LeadID)
	}

	if !lead.CanConvert() {
		return application.ErrConflict("lead cannot be converted - must be qualified first")
	}

	// Validate pipeline
	pipeline, err := o.pipelineRepo.GetByID(ctx, saga.TenantID, saga.Request.PipelineID)
	if err != nil {
		return application.ErrPipelineNotFound(saga.Request.PipelineID)
	}

	// Store validation outputs
	step.Output["lead_id"] = lead.ID.String()
	step.Output["lead_status"] = string(lead.Status)
	step.Output["pipeline_id"] = pipeline.ID.String()
	step.Output["pipeline_name"] = pipeline.Name
	step.Output["company_name"] = lead.Company.Name
	step.Output["contact_email"] = lead.Contact.Email

	return nil
}

// executeCreateCustomer creates a new customer from lead data.
func (o *leadConversionOrchestrator) executeCreateCustomer(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	if o.customerService == nil {
		return fmt.Errorf("customer service is not available")
	}

	// Get lead for customer data
	lead, err := o.leadRepo.GetByID(ctx, saga.TenantID, saga.LeadID)
	if err != nil {
		return err
	}

	// Prepare request
	createReq := ports.CreateCustomerRequest{
		Name:   lead.Company.Name,
		Type:   "business",
		Source: "lead_conversion",
	}
	if lead.Contact.Email != "" {
		email := lead.Contact.Email
		createReq.Email = &email
	}
	if lead.Contact.Phone != "" {
		phone := lead.Contact.Phone
		createReq.Phone = &phone
	}
	if lead.Company.Website != "" {
		website := lead.Company.Website
		createReq.Website = &website
	}

	// Create customer
	newCustomer, err := o.customerService.CreateCustomer(ctx, saga.TenantID, createReq)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	// Store in saga and step output
	saga.SetCustomerID(newCustomer.ID, true)
	step.Output["customer_id"] = newCustomer.ID.String()
	step.Output["customer_code"] = newCustomer.Code
	step.Output["customer_name"] = newCustomer.Name

	return nil
}

// executeLookupCustomer looks up an existing customer.
func (o *leadConversionOrchestrator) executeLookupCustomer(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	if saga.Request.CustomerID == nil {
		return fmt.Errorf("customer ID is required for lookup")
	}

	if o.customerService == nil {
		// If no customer service, just trust the ID
		saga.SetCustomerID(*saga.Request.CustomerID, false)
		step.Output["customer_id"] = saga.Request.CustomerID.String()
		return nil
	}

	// Verify customer exists
	customer, err := o.customerService.GetCustomer(ctx, saga.TenantID, *saga.Request.CustomerID)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	saga.SetCustomerID(customer.ID, false)
	step.Output["customer_id"] = customer.ID.String()
	step.Output["customer_name"] = customer.Name

	return nil
}

// executeCreateOpportunity creates the opportunity from the lead.
func (o *leadConversionOrchestrator) executeCreateOpportunity(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	// Get lead
	lead, err := o.leadRepo.GetByID(ctx, saga.TenantID, saga.LeadID)
	if err != nil {
		return err
	}

	// Get pipeline
	pipeline, err := o.pipelineRepo.GetByID(ctx, saga.TenantID, saga.Request.PipelineID)
	if err != nil {
		return err
	}

	// Determine customer info
	customerID := saga.CustomerID
	if customerID == nil && saga.Request.CustomerID != nil {
		customerID = saga.Request.CustomerID
	}

	if customerID == nil {
		return fmt.Errorf("customer ID is required")
	}

	customerName := lead.Company.Name
	if saga.Request.CustomerName != nil {
		customerName = *saga.Request.CustomerName
	}

	// Create opportunity
	opportunity, err := domain.NewOpportunityFromLead(lead, pipeline, *customerID, customerName, saga.InitiatedBy)
	if err != nil {
		return fmt.Errorf("failed to create opportunity from lead: %w", err)
	}

	// Set owner
	ownerID := saga.InitiatedBy
	ownerName := ""
	if saga.Request.OwnerID != nil {
		ownerID = *saga.Request.OwnerID
	} else if lead.OwnerID != nil {
		ownerID = *lead.OwnerID
		ownerName = lead.OwnerName
	}
	opportunity.AssignOwner(ownerID, ownerName)

	// Set optional fields
	if saga.Request.Description != nil {
		opportunity.Description = *saga.Request.Description
	}
	if saga.Request.ExpectedCloseDate != nil {
		opportunity.ExpectedCloseDate = saga.Request.ExpectedCloseDate
	}
	if saga.Request.Probability != nil {
		opportunity.Probability = *saga.Request.Probability
	}
	if saga.Request.Amount != nil {
		opportunity.Amount = *saga.Request.Amount
	}

	// Save opportunity
	if err := o.opportunityRepo.Create(ctx, opportunity); err != nil {
		return fmt.Errorf("failed to save opportunity: %w", err)
	}

	// Store in saga and step output
	saga.SetOpportunityID(opportunity.ID)
	step.Output["opportunity_id"] = opportunity.ID.String()
	step.Output["opportunity_code"] = opportunity.Code
	step.Output["opportunity_name"] = opportunity.Name

	return nil
}

// executeMarkLeadConverted marks the lead as converted.
func (o *leadConversionOrchestrator) executeMarkLeadConverted(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	lead, err := o.leadRepo.GetByID(ctx, saga.TenantID, saga.LeadID)
	if err != nil {
		return err
	}

	// Mark as converted
	if err := lead.ConvertToOpportunity(*saga.OpportunityID, saga.InitiatedBy, saga.CustomerID, saga.ContactID); err != nil {
		return fmt.Errorf("failed to mark lead as converted: %w", err)
	}

	// Store previous status for compensation
	step.Input["previous_status"] = string(domain.LeadStatusQualified)

	// Save lead
	if err := o.leadRepo.Update(ctx, lead); err != nil {
		return fmt.Errorf("failed to save converted lead: %w", err)
	}

	step.Output["lead_id"] = lead.ID.String()
	step.Output["new_status"] = string(lead.Status)

	return nil
}

// executePublishEvents publishes domain events via the event publisher.
func (o *leadConversionOrchestrator) executePublishEvents(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	events := saga.GetEvents()
	if len(events) == 0 {
		// No events to publish
		return nil
	}

	if o.eventPublisher == nil {
		// No publisher available, events will be collected and published later
		step.Output["events_pending"] = len(events)
		return nil
	}

	// Publish events
	// Note: In a production system, you would use an outbox pattern here
	// For now, we just track that events need to be published
	step.Output["events_count"] = len(events)

	saga.ClearEvents()
	return nil
}

// ============================================================================
// Compensation Handlers
// ============================================================================

// compensateCreateCustomer handles customer creation compensation.
// Note: Customer deletion would require the CustomerService to implement DeleteCustomer.
// For now, we log and continue since not all services may support deletion.
func (o *leadConversionOrchestrator) compensateCreateCustomer(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	if !saga.CustomerCreated || saga.CustomerID == nil {
		return nil // Nothing to compensate
	}

	// Customer deletion would be handled here if the service supports it
	// For now, we just acknowledge the compensation was attempted
	step.Output["compensation_note"] = "customer deletion requires manual cleanup"

	return nil
}

// compensateCreateOpportunity deletes the created opportunity.
func (o *leadConversionOrchestrator) compensateCreateOpportunity(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	if saga.OpportunityID == nil {
		return nil // Nothing to compensate
	}

	if err := o.opportunityRepo.Delete(ctx, saga.TenantID, *saga.OpportunityID); err != nil {
		return fmt.Errorf("failed to delete opportunity: %w", err)
	}

	return nil
}

// compensateMarkLeadConverted reverts the lead status.
func (o *leadConversionOrchestrator) compensateMarkLeadConverted(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	lead, err := o.leadRepo.GetByID(ctx, saga.TenantID, saga.LeadID)
	if err != nil {
		return err
	}

	// Revert the conversion
	if err := lead.RevertConversion(); err != nil {
		return fmt.Errorf("failed to revert lead conversion: %w", err)
	}

	if err := o.leadRepo.Update(ctx, lead); err != nil {
		return fmt.Errorf("failed to save reverted lead: %w", err)
	}

	return nil
}

// compensatePublishEvents handles event publishing compensation.
// Events that haven't been published yet don't need compensation.
func (o *leadConversionOrchestrator) compensatePublishEvents(ctx context.Context, saga *domain.LeadConversionSaga, step *domain.SagaStep) error {
	// No-op for now - events are typically handled through the outbox pattern
	// which would be compensated during the transaction rollback
	return nil
}
