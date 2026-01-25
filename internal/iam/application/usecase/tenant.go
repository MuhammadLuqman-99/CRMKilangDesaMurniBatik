// Package usecase contains the application use cases for the IAM service.
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/application/mapper"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// CreateTenantUseCase handles tenant creation.
type CreateTenantUseCase struct {
	tenantRepo     domain.TenantRepository
	outboxRepo     domain.OutboxRepository
	txManager      ports.TransactionManager
	eventPublisher ports.EventPublisher
	auditLogger    ports.AuditLogger
}

// NewCreateTenantUseCase creates a new CreateTenantUseCase.
func NewCreateTenantUseCase(
	tenantRepo domain.TenantRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	eventPublisher ports.EventPublisher,
	auditLogger ports.AuditLogger,
) *CreateTenantUseCase {
	return &CreateTenantUseCase{
		tenantRepo:     tenantRepo,
		outboxRepo:     outboxRepo,
		txManager:      txManager,
		eventPublisher: eventPublisher,
		auditLogger:    auditLogger,
	}
}

// Execute creates a new tenant.
func (uc *CreateTenantUseCase) Execute(ctx context.Context, req *dto.CreateTenantRequest) (*dto.CreateTenantResponse, error) {
	// Check if slug already exists
	exists, err := uc.tenantRepo.ExistsBySlug(ctx, req.Slug)
	if err != nil {
		return nil, application.ErrInternal("failed to check slug", err)
	}

	if exists {
		return nil, application.ErrConflict("tenant slug already exists")
	}

	// Create tenant with plan
	var tenant *domain.Tenant
	if req.Plan != "" {
		plan := domain.TenantPlan(req.Plan)
		if !plan.IsValid() {
			return nil, application.ErrValidation("invalid plan", map[string]interface{}{
				"plan": "must be one of: free, starter, pro, enterprise",
			})
		}
		tenant, err = domain.NewTenantWithPlan(req.Name, req.Slug, plan)
	} else {
		tenant, err = domain.NewTenant(req.Name, req.Slug)
	}

	if err != nil {
		return nil, application.ErrValidation("invalid tenant data", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Save tenant
		if err := uc.tenantRepo.Create(txCtx, tenant); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range tenant.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to create tenant", err)
	}

	tenant.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenant.GetID(),
		Action:     ports.AuditActionCreate,
		EntityType: "tenant",
		EntityID:   ptrToUUID(tenant.GetID()),
		NewValues: map[string]interface{}{
			"name": tenant.Name(),
			"slug": tenant.Slug(),
			"plan": tenant.Plan().String(),
		},
	})

	return &dto.CreateTenantResponse{
		Tenant: mapper.TenantToDTO(tenant),
	}, nil
}

// CreateTenantWithAdminUseCase handles creating a tenant with an admin user.
type CreateTenantWithAdminUseCase struct {
	tenantRepo     domain.TenantRepository
	userRepo       domain.UserRepository
	roleRepo       domain.RoleRepository
	outboxRepo     domain.OutboxRepository
	passwordHasher ports.PasswordHasher
	tokenService   ports.TokenService
	txManager      ports.TransactionManager
	auditLogger    ports.AuditLogger
}

// NewCreateTenantWithAdminUseCase creates a new CreateTenantWithAdminUseCase.
func NewCreateTenantWithAdminUseCase(
	tenantRepo domain.TenantRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	outboxRepo domain.OutboxRepository,
	passwordHasher ports.PasswordHasher,
	tokenService ports.TokenService,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *CreateTenantWithAdminUseCase {
	return &CreateTenantWithAdminUseCase{
		tenantRepo:     tenantRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		outboxRepo:     outboxRepo,
		passwordHasher: passwordHasher,
		tokenService:   tokenService,
		txManager:      txManager,
		auditLogger:    auditLogger,
	}
}

// Execute creates a new tenant with an admin user.
func (uc *CreateTenantWithAdminUseCase) Execute(ctx context.Context, req *dto.CreateTenantWithAdminRequest) (*dto.CreateTenantWithAdminResponse, error) {
	// Check if slug already exists
	exists, err := uc.tenantRepo.ExistsBySlug(ctx, req.TenantSlug)
	if err != nil {
		return nil, application.ErrInternal("failed to check slug", err)
	}

	if exists {
		return nil, application.ErrConflict("tenant slug already exists")
	}

	// Create tenant
	var tenant *domain.Tenant
	if req.Plan != "" {
		plan := domain.TenantPlan(req.Plan)
		if !plan.IsValid() {
			return nil, application.ErrValidation("invalid plan", map[string]interface{}{
				"plan": "must be one of: free, starter, pro, enterprise",
			})
		}
		tenant, err = domain.NewTenantWithPlan(req.TenantName, req.TenantSlug, plan)
	} else {
		tenant, err = domain.NewTenant(req.TenantName, req.TenantSlug)
	}

	if err != nil {
		return nil, application.ErrValidation("invalid tenant data", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Activate tenant immediately for self-service signup
	_ = tenant.Activate()

	// Create admin email
	email, err := domain.NewEmail(req.AdminEmail)
	if err != nil {
		return nil, application.ErrValidation("invalid email format", map[string]interface{}{
			"email": err.Error(),
		})
	}

	// Hash password
	passwordHash, err := uc.passwordHasher.Hash(req.AdminPassword)
	if err != nil {
		return nil, application.ErrInternal("failed to hash password", err)
	}

	password := domain.NewPasswordFromHash(passwordHash)

	// Create admin user
	adminUser, err := domain.NewUser(tenant.GetID(), email, password, req.AdminFirstName, req.AdminLastName)
	if err != nil {
		return nil, application.ErrInternal("failed to create admin user", err)
	}

	// Activate and verify admin immediately
	_ = adminUser.Activate()
	adminUser.VerifyEmail()

	// Get admin role
	adminRole, err := uc.roleRepo.FindByName(ctx, nil, domain.RoleNameAdmin)
	if err != nil {
		return nil, application.ErrInternal("admin role not found", err)
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Save tenant
		if err := uc.tenantRepo.Create(txCtx, tenant); err != nil {
			return err
		}

		// Save admin user
		if err := uc.userRepo.Create(txCtx, adminUser); err != nil {
			return err
		}

		// Assign admin role
		if err := uc.roleRepo.AssignRoleToUser(txCtx, adminUser.GetID(), adminRole.GetID(), nil); err != nil {
			return err
		}

		// Save domain events to outbox
		allEvents := append(tenant.GetDomainEvents(), adminUser.GetDomainEvents()...)
		for _, event := range allEvents {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to create tenant", err)
	}

	tenant.ClearDomainEvents()
	adminUser.ClearDomainEvents()

	// Reload admin user with roles
	adminUser, _ = uc.userRepo.FindByID(ctx, adminUser.GetID())
	roles, _ := uc.roleRepo.FindByUserID(ctx, adminUser.GetID())
	adminUser.SetRoles(roles)

	// Generate tokens
	claims := &ports.TokenClaims{
		UserID:      adminUser.GetID(),
		TenantID:    tenant.GetID(),
		Email:       adminUser.Email().String(),
		Roles:       []string{adminRole.Name()},
		Permissions: adminUser.GetPermissions().Strings(),
	}

	accessToken, _ := uc.tokenService.GenerateAccessToken(claims)
	refreshToken, _ := uc.tokenService.GenerateRefreshToken()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenant.GetID(),
		Action:     ports.AuditActionCreate,
		EntityType: "tenant",
		EntityID:   ptrToUUID(tenant.GetID()),
		NewValues: map[string]interface{}{
			"name":        tenant.Name(),
			"slug":        tenant.Slug(),
			"admin_email": email.String(),
		},
	})

	return &dto.CreateTenantWithAdminResponse{
		Tenant:       mapper.TenantToDTO(tenant),
		AdminUser:    mapper.UserToDTO(adminUser),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(uc.tokenService.GetAccessTokenExpiry()).Unix(),
	}, nil
}

// GetTenantUseCase handles retrieving a tenant.
type GetTenantUseCase struct {
	tenantRepo domain.TenantRepository
	userRepo   domain.UserRepository
}

// NewGetTenantUseCase creates a new GetTenantUseCase.
func NewGetTenantUseCase(
	tenantRepo domain.TenantRepository,
	userRepo domain.UserRepository,
) *GetTenantUseCase {
	return &GetTenantUseCase{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
	}
}

// Execute retrieves a tenant by ID.
func (uc *GetTenantUseCase) Execute(ctx context.Context, tenantID uuid.UUID) (*dto.GetTenantResponse, error) {
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, application.ErrNotFound("tenant", tenantID)
	}

	// Get user count for usage stats
	userCount, _ := uc.userRepo.CountByTenant(ctx, tenantID)

	tenantDTO := mapper.TenantWithUsageToDTO(tenant, userCount, 0) // Contact count from other service

	return &dto.GetTenantResponse{
		Tenant: tenantDTO,
	}, nil
}

// GetTenantBySlugUseCase handles retrieving a tenant by slug.
type GetTenantBySlugUseCase struct {
	tenantRepo domain.TenantRepository
}

// NewGetTenantBySlugUseCase creates a new GetTenantBySlugUseCase.
func NewGetTenantBySlugUseCase(tenantRepo domain.TenantRepository) *GetTenantBySlugUseCase {
	return &GetTenantBySlugUseCase{
		tenantRepo: tenantRepo,
	}
}

// Execute retrieves a tenant by slug.
func (uc *GetTenantBySlugUseCase) Execute(ctx context.Context, slug string) (*dto.GetTenantResponse, error) {
	tenant, err := uc.tenantRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, application.ErrNotFound("tenant", slug)
	}

	return &dto.GetTenantResponse{
		Tenant: mapper.TenantToDTO(tenant),
	}, nil
}

// ListTenantsUseCase handles listing all tenants.
type ListTenantsUseCase struct {
	tenantRepo domain.TenantRepository
}

// NewListTenantsUseCase creates a new ListTenantsUseCase.
func NewListTenantsUseCase(tenantRepo domain.TenantRepository) *ListTenantsUseCase {
	return &ListTenantsUseCase{
		tenantRepo: tenantRepo,
	}
}

// Execute lists all tenants with pagination.
func (uc *ListTenantsUseCase) Execute(ctx context.Context, req *dto.ListTenantsRequest) (*dto.ListTenantsResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDirection == "" {
		req.SortDirection = "desc"
	}

	opts := domain.TenantQueryOptions{
		Page:          req.Page,
		PageSize:      req.PageSize,
		SortBy:        req.SortBy,
		SortDirection: req.SortDirection,
		Search:        req.Search,
	}

	if req.Status != "" {
		status := domain.TenantStatus(req.Status)
		opts.Status = &status
	}

	if req.Plan != "" {
		plan := domain.TenantPlan(req.Plan)
		opts.Plan = &plan
	}

	tenants, total, err := uc.tenantRepo.FindAll(ctx, opts)
	if err != nil {
		return nil, application.ErrInternal("failed to list tenants", err)
	}

	return &dto.ListTenantsResponse{
		Tenants:    mapper.TenantsToDTO(tenants),
		Pagination: dto.NewPaginationDTO(req.Page, req.PageSize, total),
	}, nil
}

// UpdateTenantUseCase handles updating a tenant.
type UpdateTenantUseCase struct {
	tenantRepo  domain.TenantRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewUpdateTenantUseCase creates a new UpdateTenantUseCase.
func NewUpdateTenantUseCase(
	tenantRepo domain.TenantRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *UpdateTenantUseCase {
	return &UpdateTenantUseCase{
		tenantRepo:  tenantRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute updates a tenant.
func (uc *UpdateTenantUseCase) Execute(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateTenantRequest) (*dto.UpdateTenantResponse, error) {
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, application.ErrNotFound("tenant", tenantID)
	}

	oldName := tenant.Name()

	if req.Name != "" {
		if err := tenant.UpdateName(req.Name); err != nil {
			return nil, application.ErrValidation("invalid tenant name", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.tenantRepo.Update(txCtx, tenant); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range tenant.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to update tenant", err)
	}

	tenant.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		Action:     ports.AuditActionUpdate,
		EntityType: "tenant",
		EntityID:   ptrToUUID(tenantID),
		OldValues:  map[string]interface{}{"name": oldName},
		NewValues:  map[string]interface{}{"name": tenant.Name()},
	})

	return &dto.UpdateTenantResponse{
		Tenant: mapper.TenantToDTO(tenant),
	}, nil
}

// UpdateTenantSettingsUseCase handles updating tenant settings.
type UpdateTenantSettingsUseCase struct {
	tenantRepo  domain.TenantRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewUpdateTenantSettingsUseCase creates a new UpdateTenantSettingsUseCase.
func NewUpdateTenantSettingsUseCase(
	tenantRepo domain.TenantRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *UpdateTenantSettingsUseCase {
	return &UpdateTenantSettingsUseCase{
		tenantRepo:  tenantRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute updates tenant settings.
func (uc *UpdateTenantSettingsUseCase) Execute(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateTenantSettingsRequest) (*dto.UpdateTenantResponse, error) {
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, application.ErrNotFound("tenant", tenantID)
	}

	oldSettings := tenant.Settings()

	// Build updates map
	updates := make(map[string]interface{})
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.DateFormat != nil {
		updates["date_format"] = *req.DateFormat
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}
	if req.NotificationsEmail != nil {
		updates["notifications_email"] = *req.NotificationsEmail
	}

	if len(updates) > 0 {
		tenant.UpdateSettingsPartial(updates)
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.tenantRepo.Update(txCtx, tenant); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range tenant.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to update tenant settings", err)
	}

	tenant.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		Action:     ports.AuditActionUpdate,
		EntityType: "tenant_settings",
		EntityID:   ptrToUUID(tenantID),
		OldValues: map[string]interface{}{
			"timezone":   oldSettings.Timezone,
			"currency":   oldSettings.Currency,
			"language":   oldSettings.Language,
		},
		NewValues: updates,
	})

	return &dto.UpdateTenantResponse{
		Tenant: mapper.TenantToDTO(tenant),
	}, nil
}

// ChangeTenantPlanUseCase handles changing a tenant's plan.
type ChangeTenantPlanUseCase struct {
	tenantRepo  domain.TenantRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewChangeTenantPlanUseCase creates a new ChangeTenantPlanUseCase.
func NewChangeTenantPlanUseCase(
	tenantRepo domain.TenantRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *ChangeTenantPlanUseCase {
	return &ChangeTenantPlanUseCase{
		tenantRepo:  tenantRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute changes a tenant's plan.
func (uc *ChangeTenantPlanUseCase) Execute(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateTenantPlanRequest) (*dto.UpdateTenantResponse, error) {
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return nil, application.ErrNotFound("tenant", tenantID)
	}

	oldPlan := tenant.Plan()
	newPlan := domain.TenantPlan(req.Plan)

	if !newPlan.IsValid() {
		return nil, application.ErrValidation("invalid plan", map[string]interface{}{
			"plan": "must be one of: free, starter, pro, enterprise",
		})
	}

	if err := tenant.UpgradePlan(newPlan); err != nil {
		return nil, application.ErrValidation("invalid plan change", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.tenantRepo.Update(txCtx, tenant); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range tenant.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("failed to change tenant plan", err)
	}

	tenant.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		Action:     ports.AuditActionUpdate,
		EntityType: "tenant_plan",
		EntityID:   ptrToUUID(tenantID),
		OldValues:  map[string]interface{}{"plan": oldPlan.String()},
		NewValues:  map[string]interface{}{"plan": newPlan.String()},
	})

	return &dto.UpdateTenantResponse{
		Tenant: mapper.TenantToDTO(tenant),
	}, nil
}

// SuspendTenantUseCase handles suspending a tenant.
type SuspendTenantUseCase struct {
	tenantRepo  domain.TenantRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewSuspendTenantUseCase creates a new SuspendTenantUseCase.
func NewSuspendTenantUseCase(
	tenantRepo domain.TenantRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *SuspendTenantUseCase {
	return &SuspendTenantUseCase{
		tenantRepo:  tenantRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute suspends a tenant.
func (uc *SuspendTenantUseCase) Execute(ctx context.Context, req *dto.SuspendTenantRequest) error {
	tenant, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return application.ErrNotFound("tenant", req.TenantID)
	}

	if err := tenant.Suspend(req.Reason); err != nil {
		return application.ErrConflict(err.Error())
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.tenantRepo.Update(txCtx, tenant); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range tenant.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return application.ErrInternal("failed to suspend tenant", err)
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   req.TenantID,
		Action:     "suspend",
		EntityType: "tenant",
		EntityID:   ptrToUUID(req.TenantID),
		NewValues:  map[string]interface{}{"reason": req.Reason},
	})

	return nil
}

// ActivateTenantUseCase handles activating a tenant.
type ActivateTenantUseCase struct {
	tenantRepo  domain.TenantRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewActivateTenantUseCase creates a new ActivateTenantUseCase.
func NewActivateTenantUseCase(
	tenantRepo domain.TenantRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *ActivateTenantUseCase {
	return &ActivateTenantUseCase{
		tenantRepo:  tenantRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute activates a tenant.
func (uc *ActivateTenantUseCase) Execute(ctx context.Context, tenantID uuid.UUID) error {
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return application.ErrNotFound("tenant", tenantID)
	}

	if err := tenant.Activate(); err != nil {
		return application.ErrConflict(err.Error())
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.tenantRepo.Update(txCtx, tenant); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range tenant.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return application.ErrInternal("failed to activate tenant", err)
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		Action:     "activate",
		EntityType: "tenant",
		EntityID:   ptrToUUID(tenantID),
	})

	return nil
}

// DeleteTenantUseCase handles soft deleting a tenant.
type DeleteTenantUseCase struct {
	tenantRepo  domain.TenantRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewDeleteTenantUseCase creates a new DeleteTenantUseCase.
func NewDeleteTenantUseCase(
	tenantRepo domain.TenantRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *DeleteTenantUseCase {
	return &DeleteTenantUseCase{
		tenantRepo:  tenantRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute soft deletes a tenant.
func (uc *DeleteTenantUseCase) Execute(ctx context.Context, tenantID uuid.UUID) error {
	tenant, err := uc.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return application.ErrNotFound("tenant", tenantID)
	}

	tenant.Delete()

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.tenantRepo.Delete(txCtx, tenantID); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range tenant.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return application.ErrInternal("failed to delete tenant", err)
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		Action:     ports.AuditActionDelete,
		EntityType: "tenant",
		EntityID:   ptrToUUID(tenantID),
	})

	return nil
}
