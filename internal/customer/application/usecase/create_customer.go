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

// CreateCustomerUseCase handles customer creation.
type CreateCustomerUseCase struct {
	uow               domain.UnitOfWork
	eventPublisher    ports.EventPublisher
	duplicateDetector ports.DuplicateDetector
	idGenerator       ports.IDGenerator
	cache             ports.CacheService
	auditLogger       ports.AuditLogger
	customerMapper    *mapper.CustomerMapper
	contactMapper     *mapper.ContactMapper
	config            CreateCustomerConfig
}

// CreateCustomerConfig holds configuration for customer creation.
type CreateCustomerConfig struct {
	MaxContactsPerCustomer int
	DuplicateCheckEnabled  bool
	DuplicateThreshold     float64
	AutoGenerateCode       bool
	DefaultStatus          domain.CustomerStatus
	DefaultTier            domain.CustomerTier
}

// DefaultCreateCustomerConfig returns default configuration.
func DefaultCreateCustomerConfig() CreateCustomerConfig {
	return CreateCustomerConfig{
		MaxContactsPerCustomer: 100,
		DuplicateCheckEnabled:  true,
		DuplicateThreshold:     0.8,
		AutoGenerateCode:       true,
		DefaultStatus:          domain.CustomerStatusLead,
		DefaultTier:            domain.CustomerTierStandard,
	}
}

// NewCreateCustomerUseCase creates a new CreateCustomerUseCase.
func NewCreateCustomerUseCase(
	uow domain.UnitOfWork,
	eventPublisher ports.EventPublisher,
	duplicateDetector ports.DuplicateDetector,
	idGenerator ports.IDGenerator,
	cache ports.CacheService,
	auditLogger ports.AuditLogger,
	config CreateCustomerConfig,
) *CreateCustomerUseCase {
	return &CreateCustomerUseCase{
		uow:               uow,
		eventPublisher:    eventPublisher,
		duplicateDetector: duplicateDetector,
		idGenerator:       idGenerator,
		cache:             cache,
		auditLogger:       auditLogger,
		customerMapper:    mapper.NewCustomerMapper(),
		contactMapper:     mapper.NewContactMapper(),
		config:            config,
	}
}

// CreateCustomerInput holds input for customer creation.
type CreateCustomerInput struct {
	TenantID       uuid.UUID
	UserID         uuid.UUID
	Request        *dto.CreateCustomerRequest
	SkipDuplicates bool
	IPAddress      string
	UserAgent      string
}

// CreateCustomerOutput holds the result of customer creation.
type CreateCustomerOutput struct {
	Customer   *dto.CustomerResponse
	Duplicates []DuplicateInfo
	Created    bool
}

// DuplicateInfo contains information about a duplicate match.
type DuplicateInfo struct {
	CustomerID  uuid.UUID `json:"customer_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email,omitempty"`
	MatchScore  float64   `json:"match_score"`
	MatchFields []string  `json:"match_fields"`
}

// Execute creates a new customer.
func (uc *CreateCustomerUseCase) Execute(ctx context.Context, input CreateCustomerInput) (*CreateCustomerOutput, error) {
	// Validate input
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	// Check for duplicates if enabled
	if uc.config.DuplicateCheckEnabled && !input.SkipDuplicates {
		duplicates, err := uc.checkDuplicates(ctx, input)
		if err != nil {
			return nil, err
		}
		if len(duplicates) > 0 {
			return &CreateCustomerOutput{
				Duplicates: duplicates,
				Created:    false,
			}, nil
		}
	}

	// Generate customer code if needed
	code := input.Request.Code
	if code == "" && uc.config.AutoGenerateCode {
		var err error
		code, err = uc.idGenerator.NewCode(ctx, input.TenantID, "customer")
		if err != nil {
			return nil, application.ErrInternalError("failed to generate customer code", err)
		}
	}

	// Check if code already exists
	if code != "" {
		exists, err := uc.uow.Customers().ExistsByCode(ctx, input.TenantID, code)
		if err != nil {
			return nil, application.ErrInternalError("failed to check customer code", err)
		}
		if exists {
			return nil, application.ErrCustomerAlreadyExists("code", code)
		}
	}

	// Check if email already exists (if provided)
	if input.Request.Email != "" {
		exists, err := uc.uow.Customers().ExistsByEmail(ctx, input.TenantID, input.Request.Email)
		if err != nil {
			return nil, application.ErrInternalError("failed to check customer email", err)
		}
		if exists {
			return nil, application.ErrCustomerAlreadyExists("email", input.Request.Email)
		}
	}

	// Create customer domain entity
	customer, err := uc.customerMapper.ToDomain(input.TenantID, input.UserID, input.Request)
	if err != nil {
		if verrs, ok := err.(domain.ValidationErrors); ok {
			return nil, application.ErrCustomerValidation("validation failed", map[string]interface{}{
				"errors": verrs.Error(),
			})
		}
		return nil, application.ErrInternalError("failed to create customer entity", err)
	}

	// Set generated code
	if code != "" {
		customer.Code = code
	}

	// Begin transaction
	txCtx, err := uc.uow.Begin(ctx)
	if err != nil {
		return nil, application.ErrInternalError("failed to begin transaction", err)
	}
	defer uc.uow.Rollback(txCtx)

	// Create customer
	if err := uc.uow.Customers().Create(txCtx, customer); err != nil {
		return nil, application.ErrInternalError("failed to create customer", err)
	}

	// Create contacts if provided
	for i, contactInput := range input.Request.Contacts {
		contact, err := uc.contactMapper.FromCreateInput(customer.ID, input.TenantID, input.UserID, &contactInput)
		if err != nil {
			return nil, application.ErrContactValidation(
				"failed to create contact",
				map[string]interface{}{"index": i, "error": err.Error()},
			)
		}
		if err := customer.AddContact(contact); err != nil {
			return nil, application.ErrCustomerMaxContactsReached(customer.ID, uc.config.MaxContactsPerCustomer)
		}
	}

	// Update customer with contacts
	if len(input.Request.Contacts) > 0 {
		if err := uc.uow.Customers().Update(txCtx, customer); err != nil {
			return nil, application.ErrInternalError("failed to update customer with contacts", err)
		}
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

	// Clear domain events after successful commit
	customer.ClearDomainEvents()

	// Publish events asynchronously (best effort)
	go func() {
		for _, event := range customer.DomainEvents() {
			_ = uc.eventPublisher.Publish(context.Background(), event)
		}
	}()

	// Invalidate cache
	if uc.cache != nil {
		_ = uc.cache.InvalidateByTenant(ctx, input.TenantID)
	}

	// Log audit entry
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogAction(ctx, ports.AuditEntry{
			ID:          uc.idGenerator.NewID(),
			TenantID:    input.TenantID,
			UserID:      &input.UserID,
			Action:      "customer.created",
			EntityType:  "customer",
			EntityID:    customer.ID,
			NewValue:    customerToAuditMap(customer),
			IPAddress:   input.IPAddress,
			UserAgent:   input.UserAgent,
			Timestamp:   time.Now().UTC(),
		})
	}

	// Map to response
	response := uc.customerMapper.ToResponse(customer)

	return &CreateCustomerOutput{
		Customer: response,
		Created:  true,
	}, nil
}

// validateInput validates the input for customer creation.
func (uc *CreateCustomerUseCase) validateInput(input CreateCustomerInput) error {
	if input.TenantID == uuid.Nil {
		return application.ErrInvalidInput("tenant_id is required")
	}
	if input.UserID == uuid.Nil {
		return application.ErrInvalidInput("user_id is required")
	}
	if input.Request == nil {
		return application.ErrInvalidInput("request is required")
	}
	if input.Request.Name == "" {
		return application.ErrInvalidInput("name is required")
	}
	if input.Request.Type == "" {
		return application.ErrInvalidInput("type is required")
	}
	if len(input.Request.Contacts) > uc.config.MaxContactsPerCustomer {
		return application.ErrInvalidInput("too many contacts")
	}
	return nil
}

// checkDuplicates checks for potential duplicate customers.
func (uc *CreateCustomerUseCase) checkDuplicates(ctx context.Context, input CreateCustomerInput) ([]DuplicateInfo, error) {
	if uc.duplicateDetector == nil {
		// Fallback to simple duplicate check
		return uc.simpleDuplicateCheck(ctx, input)
	}

	// Create temporary customer for duplicate detection
	tempCustomer, err := uc.customerMapper.ToDomain(input.TenantID, input.UserID, input.Request)
	if err != nil {
		return nil, nil // Skip duplicate check on validation error
	}

	matches, err := uc.duplicateDetector.FindDuplicateCustomers(ctx, input.TenantID, tempCustomer)
	if err != nil {
		// Log error but don't fail the operation
		return nil, nil
	}

	var duplicates []DuplicateInfo
	for _, match := range matches {
		if match.Score >= uc.config.DuplicateThreshold {
			customer, err := uc.uow.Customers().FindByID(ctx, match.EntityID)
			if err != nil {
				continue
			}
			duplicates = append(duplicates, DuplicateInfo{
				CustomerID:  customer.ID,
				Name:        customer.Name,
				Email:       customer.Email.String(),
				MatchScore:  match.Score,
				MatchFields: match.MatchFields,
			})
		}
	}

	return duplicates, nil
}

// simpleDuplicateCheck performs a simple duplicate check using repository.
func (uc *CreateCustomerUseCase) simpleDuplicateCheck(ctx context.Context, input CreateCustomerInput) ([]DuplicateInfo, error) {
	var phone string
	if input.Request.Phone != nil {
		phone = input.Request.Phone.Number
	}

	customers, err := uc.uow.Customers().FindDuplicates(ctx, input.TenantID, input.Request.Email, phone, input.Request.Name)
	if err != nil {
		return nil, nil // Don't fail on duplicate check error
	}

	var duplicates []DuplicateInfo
	for _, customer := range customers {
		var matchFields []string
		if input.Request.Email != "" && customer.Email.String() == input.Request.Email {
			matchFields = append(matchFields, "email")
		}
		if customer.Name == input.Request.Name {
			matchFields = append(matchFields, "name")
		}

		duplicates = append(duplicates, DuplicateInfo{
			CustomerID:  customer.ID,
			Name:        customer.Name,
			Email:       customer.Email.String(),
			MatchScore:  0.9,
			MatchFields: matchFields,
		})
	}

	return duplicates, nil
}

// customerToAuditMap converts a customer to a map for audit logging.
func customerToAuditMap(customer *domain.Customer) map[string]interface{} {
	return map[string]interface{}{
		"id":       customer.ID,
		"code":     customer.Code,
		"name":     customer.Name,
		"type":     customer.Type,
		"status":   customer.Status,
		"tier":     customer.Tier,
		"source":   customer.Source,
		"email":    customer.Email.String(),
		"owner_id": customer.OwnerID,
	}
}
