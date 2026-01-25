// Package usecase contains the application use cases for the Customer service.
package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/customer/application"
	"github.com/kilang-desa-murni/crm/internal/customer/application/dto"
	"github.com/kilang-desa-murni/crm/internal/customer/application/mapper"
	"github.com/kilang-desa-murni/crm/internal/customer/application/ports"
	"github.com/kilang-desa-murni/crm/internal/customer/domain"
)

// ============================================================================
// Add Contact Use Case
// ============================================================================

// AddContactUseCase handles adding a contact to a customer.
type AddContactUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
	contactMapper  *mapper.ContactMapper
	config         ContactConfig
}

// ContactConfig holds configuration for contact operations.
type ContactConfig struct {
	MaxContactsPerCustomer int
	DuplicateCheckEnabled  bool
}

// DefaultContactConfig returns default configuration.
func DefaultContactConfig() ContactConfig {
	return ContactConfig{
		MaxContactsPerCustomer: 100,
		DuplicateCheckEnabled:  true,
	}
}

// NewAddContactUseCase creates a new AddContactUseCase.
func NewAddContactUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
	config ContactConfig,
) *AddContactUseCase {
	return &AddContactUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
		contactMapper:  mapper.NewContactMapper(),
		config:         config,
	}
}

// AddContactInput holds input for adding a contact.
type AddContactInput struct {
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Request   *dto.CreateContactRequest
	IPAddress string
	UserAgent string
}

// Execute adds a contact to a customer.
func (uc *AddContactUseCase) Execute(ctx context.Context, input AddContactInput) (*dto.ContactResponse, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Find customer
	customer, err := uc.uow.Customers().FindByID(ctx, input.Request.CustomerID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, application.ErrCustomerNotFound(input.Request.CustomerID)
		}
		return nil, application.ErrInternalError("failed to find customer", err)
	}

	// Verify tenant
	if customer.TenantID != input.TenantID {
		return nil, application.ErrTenantMismatch(input.TenantID, customer.TenantID)
	}

	// Check contact limit
	contactCount, err := uc.uow.Contacts().CountByCustomer(ctx, customer.ID)
	if err != nil {
		return nil, application.ErrInternalError("failed to count contacts", err)
	}
	if contactCount >= uc.config.MaxContactsPerCustomer {
		return nil, application.ErrCustomerMaxContactsReached(customer.ID, uc.config.MaxContactsPerCustomer)
	}

	// Check for duplicate email if enabled
	if uc.config.DuplicateCheckEnabled && input.Request.Email != "" {
		existingContacts, err := uc.uow.Contacts().FindByEmail(ctx, input.TenantID, input.Request.Email)
		if err == nil && len(existingContacts) > 0 {
			for _, existing := range existingContacts {
				if existing.CustomerID == customer.ID {
					return nil, application.ErrContactAlreadyExists("email", input.Request.Email)
				}
			}
		}
	}

	// Create contact domain entity
	contact, err := uc.contactMapper.ToDomain(input.TenantID, input.UserID, input.Request)
	if err != nil {
		if verrs, ok := err.(domain.ValidationErrors); ok {
			return nil, application.ErrContactValidation("validation failed", map[string]interface{}{
				"errors": verrs.Error(),
			})
		}
		return nil, application.ErrInternalError("failed to create contact entity", err)
	}

	// Begin transaction
	txCtx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, application.ErrInternalError("failed to begin transaction", err)
	}
	defer uc.uow.Rollback(txCtx)

	// Add contact to customer
	if err := customer.AddContact(contact); err != nil {
		return nil, application.ErrInternalError("failed to add contact", err)
	}

	// Update customer
	if err := uc.uow.Customers().Update(txCtx, customer); err != nil {
		return nil, application.ErrInternalError("failed to update customer", err)
	}

	// Save domain events to outbox
	for _, event := range customer.DomainEvents() {
		payload, err := json.Marshal(event)
		if err != nil {
			continue
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
			Action:     "contact.added",
			EntityType: "contact",
			EntityID:   contact.ID,
			NewValue:   contactToAuditMap(contact),
			Metadata:   map[string]interface{}{"customer_id": customer.ID},
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return uc.contactMapper.ToResponse(contact), nil
}

// validateInput validates the input for adding a contact.
func (uc *AddContactUseCase) validateInput(input AddContactInput) error {
	if input.TenantID == uuid.Nil {
		return application.ErrInvalidInput("tenant_id is required")
	}
	if input.UserID == uuid.Nil {
		return application.ErrInvalidInput("user_id is required")
	}
	if input.Request == nil {
		return application.ErrInvalidInput("request is required")
	}
	if input.Request.CustomerID == uuid.Nil {
		return application.ErrInvalidInput("customer_id is required")
	}
	if input.Request.FirstName == "" || input.Request.LastName == "" {
		return application.ErrInvalidInput("first_name and last_name are required")
	}
	if input.Request.Email == "" && len(input.Request.PhoneNumbers) == 0 {
		return application.ErrInvalidInput("at least email or phone number is required")
	}
	return nil
}

// contactToAuditMap converts a contact to a map for audit logging.
func contactToAuditMap(contact *domain.Contact) map[string]interface{} {
	return map[string]interface{}{
		"id":          contact.ID,
		"customer_id": contact.CustomerID,
		"name":        contact.FullName(),
		"email":       contact.Email.String(),
		"role":        contact.Role,
		"status":      contact.Status,
		"is_primary":  contact.IsPrimary,
	}
}

// ============================================================================
// Update Contact Use Case
// ============================================================================

// UpdateContactUseCase handles contact updates.
type UpdateContactUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
	contactMapper  *mapper.ContactMapper
}

// NewUpdateContactUseCase creates a new UpdateContactUseCase.
func NewUpdateContactUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
) *UpdateContactUseCase {
	return &UpdateContactUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
		contactMapper:  mapper.NewContactMapper(),
	}
}

// UpdateContactInput holds input for contact update.
type UpdateContactInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	ContactID  uuid.UUID
	Request    *dto.UpdateContactRequest
	IPAddress  string
	UserAgent  string
}

// Execute updates a contact.
func (uc *UpdateContactUseCase) Execute(ctx context.Context, input UpdateContactInput) (*dto.ContactResponse, error) {
	// Validate input
	if input.ContactID == uuid.Nil {
		return nil, application.ErrInvalidInput("contact_id is required")
	}
	if input.Request == nil {
		return nil, application.ErrInvalidInput("request is required")
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

	// Find contact within customer
	contact := customer.GetContactByID(input.ContactID)
	if contact == nil {
		return nil, application.ErrContactNotFound(input.ContactID)
	}

	// Check version
	if contact.Version != input.Request.Version {
		return nil, application.ErrContactVersionConflict(input.ContactID, input.Request.Version, contact.Version)
	}

	// Store old values for audit
	oldValues := contactToAuditMap(contact)

	// Apply updates
	if err := uc.contactMapper.ApplyUpdate(contact, input.Request); err != nil {
		if verrs, ok := err.(domain.ValidationErrors); ok {
			return nil, application.ErrContactValidation("validation failed", map[string]interface{}{
				"errors": verrs.Error(),
			})
		}
		return nil, application.ErrInternalError("failed to apply updates", err)
	}

	// Set updated by
	contact.AuditInfo.SetUpdatedBy(input.UserID)

	// Update customer (which includes the contact)
	if err := uc.uow.Customers().Update(ctx, customer); err != nil {
		return nil, application.ErrInternalError("failed to update contact", err)
	}

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, "customer", customer.ID)
	}

	// Audit log
	if uc.auditLogger != nil {
		newValues := contactToAuditMap(contact)
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "contact.updated",
			EntityType: "contact",
			EntityID:   contact.ID,
			OldValue:   oldValues,
			NewValue:   newValues,
			Metadata:   map[string]interface{}{"customer_id": customer.ID},
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return uc.contactMapper.ToResponse(contact), nil
}

// ============================================================================
// Delete Contact Use Case
// ============================================================================

// DeleteContactUseCase handles contact deletion.
type DeleteContactUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
}

// NewDeleteContactUseCase creates a new DeleteContactUseCase.
func NewDeleteContactUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
) *DeleteContactUseCase {
	return &DeleteContactUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
	}
}

// DeleteContactInput holds input for contact deletion.
type DeleteContactInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	ContactID  uuid.UUID
	IPAddress  string
	UserAgent  string
}

// Execute deletes a contact.
func (uc *DeleteContactUseCase) Execute(ctx context.Context, input DeleteContactInput) error {
	// Validate input
	if input.ContactID == uuid.Nil {
		return application.ErrInvalidInput("contact_id is required")
	}

	// Find customer
	customer, err := uc.uow.Customers().FindByID(ctx, input.CustomerID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return application.ErrCustomerNotFound(input.CustomerID)
		}
		return application.ErrInternalError("failed to find customer", err)
	}

	// Verify tenant
	if customer.TenantID != input.TenantID {
		return application.ErrTenantMismatch(input.TenantID, customer.TenantID)
	}

	// Find contact
	contact := customer.GetContactByID(input.ContactID)
	if contact == nil {
		return application.ErrContactNotFound(input.ContactID)
	}

	// Store contact data for audit
	contactData := contactToAuditMap(contact)

	// Remove contact from customer
	if err := customer.RemoveContact(input.ContactID); err != nil {
		return application.ErrInternalError("failed to remove contact", err)
	}

	// Update customer
	if err := uc.uow.Customers().Update(ctx, customer); err != nil {
		return application.ErrInternalError("failed to update customer", err)
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
			Action:     "contact.deleted",
			EntityType: "contact",
			EntityID:   input.ContactID,
			OldValue:   contactData,
			Metadata:   map[string]interface{}{"customer_id": customer.ID},
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return nil
}

// ============================================================================
// Get Contact Use Case
// ============================================================================

// GetContactUseCase handles getting a contact.
type GetContactUseCase struct {
	uow           domain.UnitOfWork
	contactMapper *mapper.ContactMapper
}

// NewGetContactUseCase creates a new GetContactUseCase.
func NewGetContactUseCase(uow domain.UnitOfWork) *GetContactUseCase {
	return &GetContactUseCase{
		uow:           uow,
		contactMapper: mapper.NewContactMapper(),
	}
}

// GetContactInput holds input for getting a contact.
type GetContactInput struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	ContactID  uuid.UUID
}

// Execute gets a contact by ID.
func (uc *GetContactUseCase) Execute(ctx context.Context, input GetContactInput) (*dto.ContactResponse, error) {
	// Find contact
	contact, err := uc.uow.Contacts().FindByID(ctx, input.ContactID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, application.ErrContactNotFound(input.ContactID)
		}
		return nil, application.ErrInternalError("failed to find contact", err)
	}

	// Verify tenant
	if contact.TenantID != input.TenantID {
		return nil, application.ErrTenantMismatch(input.TenantID, contact.TenantID)
	}

	// Verify customer if provided
	if input.CustomerID != uuid.Nil && contact.CustomerID != input.CustomerID {
		return nil, application.ErrContactNotFound(input.ContactID)
	}

	return uc.contactMapper.ToResponse(contact), nil
}

// ============================================================================
// List Contacts Use Case
// ============================================================================

// ListContactsUseCase handles listing contacts for a customer.
type ListContactsUseCase struct {
	uow           domain.UnitOfWork
	contactMapper *mapper.ContactMapper
}

// NewListContactsUseCase creates a new ListContactsUseCase.
func NewListContactsUseCase(uow domain.UnitOfWork) *ListContactsUseCase {
	return &ListContactsUseCase{
		uow:           uow,
		contactMapper: mapper.NewContactMapper(),
	}
}

// ListContactsInput holds input for listing contacts.
type ListContactsInput struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	Offset     int
	Limit      int
}

// Execute lists contacts for a customer.
func (uc *ListContactsUseCase) Execute(ctx context.Context, input ListContactsInput) (*dto.ContactListResponse, error) {
	// Find customer to verify it exists and tenant matches
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

	// Get contacts
	contacts, err := uc.uow.Contacts().FindByCustomerID(ctx, input.CustomerID)
	if err != nil {
		return nil, application.ErrInternalError("failed to list contacts", err)
	}

	// Apply pagination manually (or use filter in repository)
	total := int64(len(contacts))
	if input.Limit == 0 {
		input.Limit = 20
	}

	start := input.Offset
	end := input.Offset + input.Limit
	if start >= len(contacts) {
		contacts = []*domain.Contact{}
	} else {
		if end > len(contacts) {
			end = len(contacts)
		}
		contacts = contacts[start:end]
	}

	return uc.contactMapper.ToListResponse(contacts, total, input.Offset, input.Limit), nil
}

// ============================================================================
// Set Primary Contact Use Case
// ============================================================================

// SetPrimaryContactUseCase handles setting a contact as primary.
type SetPrimaryContactUseCase struct {
	uow            domain.UnitOfWork
	eventPublisher ports.EventPublisher
	idGenerator    ports.IDGenerator
	cache          ports.CacheService
	auditLogger    ports.AuditLogger
	contactMapper  *mapper.ContactMapper
}

// NewSetPrimaryContactUseCase creates a new SetPrimaryContactUseCase.
func NewSetPrimaryContactUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
) *SetPrimaryContactUseCase {
	return &SetPrimaryContactUseCase{
		uow:            uow,
		eventPublisher: eventPublisher,
		idGenerator:    idGenerator,
		cache:          cache,
		auditLogger:    auditLogger,
		contactMapper:  mapper.NewContactMapper(),
	}
}

// SetPrimaryContactInput holds input for setting primary contact.
type SetPrimaryContactInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	CustomerID uuid.UUID
	Request    *dto.SetPrimaryContactRequest
	IPAddress  string
	UserAgent  string
}

// Execute sets a contact as primary.
func (uc *SetPrimaryContactUseCase) Execute(ctx context.Context, input SetPrimaryContactInput) (*dto.ContactResponse, error) {
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

	// Set primary contact
	if err := customer.SetPrimaryContact(input.Request.ContactID); err != nil {
		return nil, application.ErrContactNotFound(input.Request.ContactID)
	}

	// Update customer
	if err := uc.uow.Customers().Update(ctx, customer); err != nil {
		return nil, application.ErrInternalError("failed to update customer", err)
	}

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, "customer", customer.ID)
	}

	// Get the updated contact
	contact := customer.GetContactByID(input.Request.ContactID)

	// Audit log
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:         uc.idGenerator.NewID(),
			TenantID:   input.TenantID,
			UserID:     &input.UserID,
			Action:     "contact.set_primary",
			EntityType: "contact",
			EntityID:   input.Request.ContactID,
			Metadata:   map[string]interface{}{"customer_id": customer.ID},
			IPAddress:  input.IPAddress,
			UserAgent:  input.UserAgent,
			Timestamp:  time.Now().UTC(),
		})
	}

	return uc.contactMapper.ToResponse(contact), nil
}
