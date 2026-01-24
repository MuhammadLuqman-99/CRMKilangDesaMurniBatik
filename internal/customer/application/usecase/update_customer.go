// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/dto"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/mapper"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/application/ports"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/customer/domain"
)

// UpdateCustomerUseCase handles customer updates.
type UpdateCustomerUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
	customerMapper *mapper.CustomerMapper
}

// NewUpdateCustomerUseCase creates a new UpdateCustomerUseCase.
func NewUpdateCustomerUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
) *UpdateCustomerUseCase {
	return &UpdateCustomerUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// UpdateCustomerInput holds input for customer update.
type UpdateCustomerInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	Request    *dto.UpdateCustomerRequest
	IPAddress  string
	UserAgent  string
}

// Execute updates a customer.
func (uc *UpdateCustomerUseCase) Execute(ctx context.Context, input UpdateCustomerInput) (*dto.CustomerResponse, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Find existing customer
	customer, err := uc.uow.Customers().FindByID(ctx, input.CustomerID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, application.ErrCustomerNotFound(input.CustomerID)
		}
		return nil, application.ErrInternalError("failed to find customer", err)
	}

	// Verify tenant
	if customer.TenantID != input.TenantID {
		return nil, application.ErrTenantMismatch(input.TenantID, customer.TenantID)
	}

	// Check version for optimistic locking
	if customer.Version != input.Request.Version {
		return nil, application.ErrCustomerVersionConflict(input.CustomerID, input.Request.Version, customer.Version)
	}

	// Store old values for audit
	oldValues := customerToAuditMap(customer)

	// Apply updates
	if err := uc.customerMapper.ApplyUpdate(customer, input.Request); err != nil {
		if verrs, ok := err.(domain.ValidationErrors); ok {
			return nil, application.ErrCustomerValidation("validation failed", map[string]interface{}{
				"errors": verrs.Error(),
			})
		}
		return nil, application.ErrInternalError("failed to apply updates", err)
	}

	// Set updated by
	customer.AuditInfo.SetUpdatedBy(input.UserID)

	// Begin transaction
	txCtx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, application.ErrInternalError("failed to begin transaction", err)
	}
	defer uc.uow.Rollback(txCtx)

	// Update customer
	if err := uc.uow.Customers().Update(txCtx, customer); err != nil {
		if domain.IsVersionConflictError(err) {
			return nil, application.ErrCustomerVersionConflict(input.CustomerID, input.Request.Version, customer.Version)
		}
		return nil, application.ErrInternalError("failed to update customer", err)
	}

	// Save domain events to outbox
	for _, event := range customer.DomainEvents() {
		payload, err := json.Marshal(event)
		if err != nil {
			return nil, application.ErrInternalError("failed to serialize event", err)
		}

		outboxEntry := &domain.OutboxEntry{
			ID:          uc.idGenerator.NewID(),
			TenantID:    input.TenantID,
			EventType:   event.EventType(),
			AggregateID: customer.ID,
			Payload:     payload,
			CreatedAt:   time.Now().UTC(),
		}

		if err := uc.uow.Outbox().Create(txCtx, outboxEntry); err != nil {
			return nil, application.ErrInternalError("failed to save outbox entry", err)
		}
	}

	// Commit transaction
	if err := uc.uow.Commit(txCtx); err != nil {
		return nil, application.ErrInternalError("failed to commit transaction", err)
	}

	// Clear domain events
	customer.ClearDomainEvents()

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, "customer", customer.ID)
	}

	// Log audit entry
	if uc.auditLogger != nil {
		newValues := customerToAuditMap(customer)
		changes := computeChanges(oldValues, newValues)
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "customer.updated",
			EntityType: "customer",
			EntityID:   customer.ID,
			OldValue:   oldValues,
			NewValue:   newValues,
			Changes:    changes,
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return uc.customerMapper.ToResponse(customer), nil
}

// validateInput validates the input for customer update.
func (uc *UpdateCustomerUseCase) validateInput(input UpdateCustomerInput) error {
	if input.TenantID == uuid.Nil {
		return application.ErrInvalidInput("tenant_id is required")
	}
	if input.UserID == uuid.Nil {
		return application.ErrInvalidInput("user_id is required")
	}
	if input.CustomerID == uuid.Nil {
		return application.ErrInvalidInput("customer_id is required")
	}
	if input.Request == nil {
		return application.ErrInvalidInput("request is required")
	}
	if input.Request.Version < 1 {
		return application.ErrInvalidInput("version is required")
	}
	return nil
}

// computeChanges computes the differences between old and new values.
func computeChanges(oldValues, newValues map[string]interface{}) []ports.FieldChange {
	var changes []ports.FieldChange

	for key, newVal := range newValues {
		oldVal, exists := oldValues[key]
		if !exists || oldVal != newVal {
			changes = append(changes, ports.FieldChange{
				Field:    key,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}

	return changes
}

// ============================================================================
// Change Status Use Case
// ============================================================================

// ChangeCustomerStatusUseCase handles customer status changes.
type ChangeCustomerStatusUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
	customerMapper *mapper.CustomerMapper
}

// NewChangeCustomerStatusUseCase creates a new ChangeCustomerStatusUseCase.
func NewChangeCustomerStatusUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
) *ChangeCustomerStatusUseCase {
	return &ChangeCustomerStatusUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// ChangeStatusInput holds input for status change.
type ChangeStatusInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	Request    *dto.ChangeStatusRequest
	IPAddress  string
	UserAgent  string
}

// Execute changes a customer's status.
func (uc *ChangeCustomerStatusUseCase) Execute(ctx context.Context, input ChangeStatusInput) (*dto.CustomerResponse, error) {
	// Validate input
	if input.CustomerID == uuid.Nil {
		return nil, application.ErrInvalidInput("customer_id is required")
	}

	// Find customer
	customer, err := uc.uow.Customers().FindByID(ctx, input.CustomerID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, application.ErrCustomerNotFound(input.CustomerID)
		}
		return nil, application.ErrInternalError("failed to find customer", err)
	}

	// Verify tenant
	if customer.TenantID != input.TenantID {
		return nil, application.ErrTenantMismatch(input.TenantID, customer.TenantID)
	}

	// Check version
	if customer.Version != input.Request.Version {
		return nil, application.ErrCustomerVersionConflict(input.CustomerID, input.Request.Version, customer.Version)
	}

	oldStatus := customer.Status

	// Apply status change based on target status
	var statusErr error
	switch input.Request.Status {
	case domain.CustomerStatusActive:
		statusErr = customer.Activate()
	case domain.CustomerStatusInactive:
		statusErr = customer.Deactivate()
	case domain.CustomerStatusChurned:
		statusErr = customer.MarkAsChurned(input.Request.Reason)
	case domain.CustomerStatusBlocked:
		statusErr = customer.Block(input.Request.Reason)
	case domain.CustomerStatusProspect:
		statusErr = customer.ConvertToProspect()
	default:
		return nil, application.ErrCustomerInvalidStatus(string(oldStatus), string(input.Request.Status))
	}

	if statusErr != nil {
		return nil, application.ErrCustomerInvalidStatus(string(oldStatus), string(input.Request.Status))
	}

	// Update in database
	txCtx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, application.ErrInternalError("failed to begin transaction", err)
	}
	defer uc.uow.Rollback(txCtx)

	if err := uc.uow.Customers().Update(txCtx, customer); err != nil {
		return nil, application.ErrInternalError("failed to update customer", err)
	}

	// Save events to outbox
	for _, event := range customer.DomainEvents() {
		payload, _ := json.Marshal(event)
		outboxEntry := &domain.OutboxEntry{
			ID:          uc.idGenerator.NewID(),
			TenantID:    input.TenantID,
			EventType:   event.EventType(),
			AggregateID: customer.ID,
			Payload:     payload,
			CreatedAt:   time.Now().UTC(),
		}
		if err := uc.uow.Outbox().Create(txCtx, outboxEntry); err != nil {
			return nil, application.ErrInternalError("failed to save outbox entry", err)
		}
	}

	if err := uc.uow.Commit(txCtx); err != nil {
		return nil, application.ErrInternalError("failed to commit transaction", err)
	}

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, "customer", customer.ID)
	}

	// Audit log
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "customer.status_changed",
			EntityType: "customer",
			EntityID:   customer.ID,
			OldValue:   map[string]interface{}{"status": oldStatus},
			NewValue:   map[string]interface{}{"status": customer.Status, "reason": input.Request.Reason},
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return uc.customerMapper.ToResponse(customer), nil
}

// ============================================================================
// Assign Owner Use Case
// ============================================================================

// AssignCustomerOwnerUseCase handles owner assignment.
type AssignCustomerOwnerUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
	customerMapper *mapper.CustomerMapper
}

// NewAssignCustomerOwnerUseCase creates a new AssignCustomerOwnerUseCase.
func NewAssignCustomerOwnerUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
) *AssignCustomerOwnerUseCase {
	return &AssignCustomerOwnerUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// AssignOwnerInput holds input for owner assignment.
type AssignOwnerInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	Request    *dto.AssignOwnerRequest
	IPAddress  string
	UserAgent  string
}

// Execute assigns an owner to a customer.
func (uc *AssignCustomerOwnerUseCase) Execute(ctx context.Context, input AssignOwnerInput) (*dto.CustomerResponse, error) {
	// Find customer
	customer, err := uc.uow.Customers().FindByID(ctx, input.CustomerID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, application.ErrCustomerNotFound(input.CustomerID)
		}
		return nil, application.ErrInternalError("failed to find customer", err)
	}

	// Verify tenant
	if customer.TenantID != input.TenantID {
		return nil, application.ErrTenantMismatch(input.TenantID, customer.TenantID)
	}

	// Check version
	if customer.Version != input.Request.Version {
		return nil, application.ErrCustomerVersionConflict(input.CustomerID, input.Request.Version, customer.Version)
	}

	oldOwner := customer.OwnerID

	// Assign owner
	customer.AssignOwner(input.Request.OwnerID)

	// Update in database
	if err := uc.uow.Customers().Update(ctx, customer); err != nil {
		return nil, application.ErrInternalError("failed to update customer", err)
	}

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, "customer", customer.ID)
	}

	// Audit log
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "customer.owner_assigned",
			EntityType: "customer",
			EntityID:   customer.ID,
			OldValue:   map[string]interface{}{"owner_id": oldOwner},
			NewValue:   map[string]interface{}{"owner_id": customer.OwnerID},
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return uc.customerMapper.ToResponse(customer), nil
}

// ============================================================================
// Convert Customer Use Case
// ============================================================================

// ConvertCustomerUseCase handles customer conversion (lead -> customer).
type ConvertCustomerUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
	customerMapper *mapper.CustomerMapper
}

// NewConvertCustomerUseCase creates a new ConvertCustomerUseCase.
func NewConvertCustomerUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
) *ConvertCustomerUseCase {
	return &ConvertCustomerUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// ConvertCustomerInput holds input for customer conversion.
type ConvertCustomerInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	Request    *dto.ConvertCustomerRequest
	IPAddress  string
	UserAgent  string
}

// Execute converts a lead/prospect to a customer.
func (uc *ConvertCustomerUseCase) Execute(ctx context.Context, input ConvertCustomerInput) (*dto.CustomerResponse, error) {
	// Find customer
	customer, err := uc.uow.Customers().FindByID(ctx, input.CustomerID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, application.ErrCustomerNotFound(input.CustomerID)
		}
		return nil, application.ErrInternalError("failed to find customer", err)
	}

	// Verify tenant
	if customer.TenantID != input.TenantID {
		return nil, application.ErrTenantMismatch(input.TenantID, customer.TenantID)
	}

	// Check version
	if customer.Version != input.Request.Version {
		return nil, application.ErrCustomerVersionConflict(input.CustomerID, input.Request.Version, customer.Version)
	}

	oldStatus := customer.Status

	// Convert to customer
	if err := customer.ConvertToCustomer(); err != nil {
		return nil, application.ErrCustomerInvalidStatus(string(oldStatus), string(domain.CustomerStatusActive))
	}

	// Begin transaction
	txCtx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, application.ErrInternalError("failed to begin transaction", err)
	}
	defer uc.uow.Rollback(txCtx)

	// Update customer
	if err := uc.uow.Customers().Update(txCtx, customer); err != nil {
		return nil, application.ErrInternalError("failed to update customer", err)
	}

	// Save events to outbox
	for _, event := range customer.DomainEvents() {
		payload, _ := json.Marshal(event)
		outboxEntry := &domain.OutboxEntry{
			ID:          uc.idGenerator.NewID(),
			TenantID:    input.TenantID,
			EventType:   event.EventType(),
			AggregateID: customer.ID,
			Payload:     payload,
			CreatedAt:   time.Now().UTC(),
		}
		if err := uc.uow.Outbox().Create(txCtx, outboxEntry); err != nil {
			return nil, application.ErrInternalError("failed to save outbox entry", err)
		}
	}

	if err := uc.uow.Commit(txCtx); err != nil {
		return nil, application.ErrInternalError("failed to commit transaction", err)
	}

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, "customer", customer.ID)
	}

	// Audit log
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "customer.converted",
			EntityType: "customer",
			EntityID:   customer.ID,
			OldValue:   map[string]interface{}{"status": oldStatus},
			NewValue:   map[string]interface{}{"status": customer.Status, "converted_at": customer.ConvertedAt},
			Metadata:   map[string]interface{}{"reason": input.Request.Reason},
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return uc.customerMapper.ToResponse(customer), nil
}
