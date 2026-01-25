// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application"
	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/ports"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// DeleteCustomerUseCase handles customer deletion.
type DeleteCustomerUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	searchIndex    ports.SearchIndex
	auditLogger    ports.AuditLogger
	config         DeleteCustomerConfig
}

// DeleteCustomerConfig holds configuration for customer deletion.
type DeleteCustomerConfig struct {
	AllowHardDelete        bool
	RequireReason          bool
	CheckActiveDeals       bool
	CheckOutstandingBalance bool
}

// DefaultDeleteCustomerConfig returns default configuration.
func DefaultDeleteCustomerConfig() DeleteCustomerConfig {
	return DeleteCustomerConfig{
		AllowHardDelete:        false,
		RequireReason:          false,
		CheckActiveDeals:       true,
		CheckOutstandingBalance: true,
	}
}

// NewDeleteCustomerUseCase creates a new DeleteCustomerUseCase.
func NewDeleteCustomerUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	searchIndex ports.SearchIndex,
	auditLogger ports.AuditLogger,
	config DeleteCustomerConfig,
) *DeleteCustomerUseCase {
	return &DeleteCustomerUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		searchIndex:    searchIndex,
		auditLogger:    auditLogger,
		config:         config,
	}
}

// DeleteCustomerInput holds input for customer deletion.
type DeleteCustomerInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	HardDelete bool
	Reason     string
	IPAddress  string
	UserAgent  string
}

// DeleteCustomerOutput holds the result of customer deletion.
type DeleteCustomerOutput struct {
	Success    bool   `json:"success"`
	CustomerID uuid.UUID `json:"customer_id"`
	HardDelete bool   `json:"hard_delete"`
}

// Execute deletes a customer.
func (uc *DeleteCustomerUseCase) Execute(ctx context.Context, input DeleteCustomerInput) (*DeleteCustomerOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
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

	// Check if can be deleted
	if err := uc.checkDeletionAllowed(ctx, customer); err != nil {
		return nil, err
	}

	// Store customer data for audit before deletion
	customerData := customerToAuditMap(customer)

	// Begin transaction
	txCtx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, application.ErrInternalError("failed to begin transaction", err)
	}
	defer uc.uow.Rollback(txCtx)

	// Perform deletion
	if input.HardDelete && uc.config.AllowHardDelete {
		// Hard delete
		if err := uc.uow.Customers().HardDelete(txCtx, input.CustomerID); err != nil {
			return nil, application.ErrInternalError("failed to hard delete customer", err)
		}
	} else {
		// Soft delete
		customer.SoftDelete()
		if err := uc.uow.Customers().Delete(txCtx, input.CustomerID); err != nil {
			return nil, application.ErrInternalError("failed to soft delete customer", err)
		}

		// Save deletion event to outbox
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
	}

	// Commit transaction
	if err := uc.uow.Commit(txCtx); err != nil {
		return nil, application.ErrInternalError("failed to commit transaction", err)
	}

	// Remove from search index
	if uc.searchIndex != nil {
		_ = uc.searchIndex.RemoveCustomer(ctx, input.CustomerID)
	}

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, "customer", input.CustomerID)
		_ = uc.cache.InvalidateByTenant(ctx, input.TenantID)
	}

	// Audit log
	if uc.auditLogger != nil {
		action := "customer.deleted"
		if input.HardDelete {
			action = "customer.hard_deleted"
		}
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     action,
			EntityType: "customer",
			EntityID:   input.CustomerID,
			OldValue:   customerData,
			Metadata:   map[string]interface{}{"reason": input.Reason, "hard_delete": input.HardDelete},
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return &DeleteCustomerOutput{
		Success:    true,
		CustomerID: input.CustomerID,
		HardDelete: input.HardDelete && uc.config.AllowHardDelete,
	}, nil
}

// validateInput validates the input for customer deletion.
func (uc *DeleteCustomerUseCase) validateInput(input DeleteCustomerInput) error {
	if input.TenantID == uuid.Nil {
		return application.ErrInvalidInput("tenant_id is required")
	}
	if input.UserID == uuid.Nil {
		return application.ErrInvalidInput("user_id is required")
	}
	if input.CustomerID == uuid.Nil {
		return application.ErrInvalidInput("customer_id is required")
	}
	if uc.config.RequireReason && input.Reason == "" {
		return application.ErrInvalidInput("reason is required for deletion")
	}
	if input.HardDelete && !uc.config.AllowHardDelete {
		return application.ErrInvalidInput("hard delete is not allowed")
	}
	return nil
}

// checkDeletionAllowed checks if the customer can be deleted.
func (uc *DeleteCustomerUseCase) checkDeletionAllowed(ctx context.Context, customer *domain.Customer) error {
	// Check for outstanding balance
	if uc.config.CheckOutstandingBalance {
		if customer.Financials.CurrentBalance != nil && customer.Financials.CurrentBalance.Amount > 0 {
			return application.ErrCustomerCannotDelete(customer.ID, "customer has outstanding balance")
		}
	}

	// Check for active deals (would be checked via Sales service in real implementation)
	if uc.config.CheckActiveDeals {
		if customer.Stats.DealCount > customer.Stats.WonDealCount+customer.Stats.LostDealCount {
			return application.ErrCustomerCannotDelete(customer.ID, "customer has active deals")
		}
	}

	return nil
}

// ============================================================================
// Bulk Delete Use Case
// ============================================================================

// BulkDeleteCustomersUseCase handles bulk customer deletion.
type BulkDeleteCustomersUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	searchIndex    ports.SearchIndex
	auditLogger    ports.AuditLogger
	config         DeleteCustomerConfig
}

// NewBulkDeleteCustomersUseCase creates a new BulkDeleteCustomersUseCase.
func NewBulkDeleteCustomersUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	searchIndex ports.SearchIndex,
	auditLogger ports.AuditLogger,
	config DeleteCustomerConfig,
) *BulkDeleteCustomersUseCase {
	return &BulkDeleteCustomersUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		searchIndex:    searchIndex,
		auditLogger:    auditLogger,
		config:         config,
	}
}

// BulkDeleteInput holds input for bulk deletion.
type BulkDeleteInput struct {
	TenantID    uuid.UUID
	UserID      uuid.UUID
	Request     *dto.BulkDeleteRequest
	IPAddress   string
	UserAgent   string
}

// Execute performs bulk deletion.
func (uc *BulkDeleteCustomersUseCase) Execute(ctx context.Context, input BulkDeleteInput) (*dto.BulkOperationResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil || input.UserID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id and user_id are required")
	}
	if input.Request == nil || len(input.Request.CustomerIDs) == 0 {
		return nil, application.ErrInvalidInput("customer_ids are required")
	}

	var processed, succeeded, failed int
	var errors []string

	// Begin transaction
	txCtx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, application.ErrInternalError("failed to begin transaction", err)
	}
	defer uc.uow.Rollback(txCtx)

	for _, customerID := range input.Request.CustomerIDs {
		processed++

		// Find customer
		customer, err := uc.uow.Customers().FindByID(txCtx, customerID)
		if err != nil {
			failed++
			errors = append(errors, "customer "+customerID.String()+": not found")
			continue
		}

		// Verify tenant
		if customer.TenantID != input.TenantID {
			failed++
			errors = append(errors, "customer "+customerID.String()+": tenant mismatch")
			continue
		}

		// Soft delete
		if err := uc.uow.Customers().Delete(txCtx, customerID); err != nil {
			failed++
			errors = append(errors, "customer "+customerID.String()+": "+err.Error())
			continue
		}

		succeeded++

		// Remove from search index
		if uc.searchIndex != nil {
			_ = uc.searchIndex.RemoveCustomer(ctx, customerID)
		}
	}

	// Commit transaction
	if err := uc.uow.Commit(txCtx); err != nil {
		return nil, application.ErrInternalError("failed to commit transaction", err)
	}

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.InvalidateByTenant(ctx, input.TenantID)
	}

	// Audit log
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "customers.bulk_deleted",
			EntityType: "customer",
			EntityID:   uuid.Nil,
			Metadata: map[string]interface{}{
				"customer_ids": input.Request.CustomerIDs,
				"processed":    processed,
				"succeeded":    succeeded,
				"failed":       failed,
			},
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
			Timestamp: time.Now().UTC(),
		})
	}

	return &dto.BulkOperationResponse{
		Processed: processed,
		Succeeded: succeeded,
		Failed:    failed,
		Errors:    errors,
	}, nil
}

// ============================================================================
// Get Customer Use Case
// ============================================================================

// GetCustomerUseCase handles getting a single customer.
type GetCustomerUseCase struct {
	uow            domain.UnitOfWork
	cache          ports.CacheService
	customerMapper *mapper.CustomerMapper
}

// NewGetCustomerUseCase creates a new GetCustomerUseCase.
func NewGetCustomerUseCase(
	uow domain.UnitOfWork,
	cache ports.CacheService,
) *GetCustomerUseCase {
	return &GetCustomerUseCase{
		uow:            uow,
		cache:          cache,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// GetCustomerInput holds input for getting a customer.
type GetCustomerInput struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
}

// Execute gets a customer by ID.
func (uc *GetCustomerUseCase) Execute(ctx context.Context, input GetCustomerInput) (*dto.CustomerResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id is required")
	}
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

	return uc.customerMapper.ToResponse(customer), nil
}

// ============================================================================
// Get Customer By Code Use Case
// ============================================================================

// GetCustomerByCodeUseCase handles getting a customer by code.
type GetCustomerByCodeUseCase struct {
	uow            domain.UnitOfWork
	customerMapper *mapper.CustomerMapper
}

// NewGetCustomerByCodeUseCase creates a new GetCustomerByCodeUseCase.
func NewGetCustomerByCodeUseCase(uow domain.UnitOfWork) *GetCustomerByCodeUseCase {
	return &GetCustomerByCodeUseCase{
		uow:            uow,
		customerMapper: mapper.NewCustomerMapper(),
	}
}

// GetCustomerByCodeInput holds input for getting a customer by code.
type GetCustomerByCodeInput struct {
	TenantID uuid.UUID
	Code     string
}

// Execute gets a customer by code.
func (uc *GetCustomerByCodeUseCase) Execute(ctx context.Context, input GetCustomerByCodeInput) (*dto.CustomerResponse, error) {
	// Validate input
	if input.TenantID == uuid.Nil {
		return nil, application.ErrInvalidInput("tenant_id is required")
	}
	if input.Code == "" {
		return nil, application.ErrInvalidInput("code is required")
	}

	// Find customer
	customer, err := uc.uow.Customers().FindByCode(ctx, input.TenantID, input.Code)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, application.ErrCustomerNotFound(uuid.Nil).WithDetail("code", input.Code)
		}
		return nil, application.ErrInternalError("failed to find customer", err)
	}

	return uc.customerMapper.ToResponse(customer), nil
}
