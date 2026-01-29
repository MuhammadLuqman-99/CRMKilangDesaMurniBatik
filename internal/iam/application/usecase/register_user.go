// Package usecase contains the application use cases for the IAM service.
package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/application/mapper"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// RegisterUserUseCase handles user registration.
type RegisterUserUseCase struct {
	userRepo       domain.UserRepository
	tenantRepo     domain.TenantRepository
	roleRepo       domain.RoleRepository
	outboxRepo     domain.OutboxRepository
	passwordHasher ports.PasswordHasher
	tokenService   ports.TokenService
	emailService   ports.EmailService
	verifyService  ports.VerificationTokenService
	txManager      ports.TransactionManager
	eventPublisher ports.EventPublisher
	auditLogger    ports.AuditLogger
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase.
func NewRegisterUserUseCase(
	userRepo domain.UserRepository,
	tenantRepo domain.TenantRepository,
	roleRepo domain.RoleRepository,
	outboxRepo domain.OutboxRepository,
	passwordHasher ports.PasswordHasher,
	tokenService ports.TokenService,
	emailService ports.EmailService,
	verifyService ports.VerificationTokenService,
	txManager ports.TransactionManager,
	eventPublisher ports.EventPublisher,
	auditLogger ports.AuditLogger,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		roleRepo:       roleRepo,
		outboxRepo:     outboxRepo,
		passwordHasher: passwordHasher,
		tokenService:   tokenService,
		emailService:   emailService,
		verifyService:  verifyService,
		txManager:      txManager,
		eventPublisher: eventPublisher,
		auditLogger:    auditLogger,
	}
}

// Execute registers a new user.
func (uc *RegisterUserUseCase) Execute(ctx context.Context, req *dto.RegisterUserRequest) (*dto.RegisterUserResponse, error) {
	// Validate tenant exists and is active
	tenant, err := uc.tenantRepo.FindByID(ctx, req.TenantID)
	if err != nil {
		return nil, application.ErrNotFound("tenant", req.TenantID)
	}

	if !tenant.IsActive() {
		return nil, application.ErrTenantInactive()
	}

	// Check user limit for tenant
	userCount, err := uc.userRepo.CountByTenant(ctx, req.TenantID)
	if err != nil {
		return nil, application.ErrInternal("failed to check user count", err)
	}

	if !tenant.CanAddUser(int(userCount)) {
		return nil, application.ErrConflict("user limit reached for current plan")
	}

	// Create email value object
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, application.ErrValidation("invalid email format", map[string]interface{}{
			"email": err.Error(),
		})
	}

	// Check if email already exists in tenant
	exists, err := uc.userRepo.ExistsByEmail(ctx, req.TenantID, email)
	if err != nil {
		return nil, application.ErrInternal("failed to check email existence", err)
	}

	if exists {
		return nil, application.ErrConflict("email already registered in this tenant")
	}

	// Validate password policy
	policy := domain.DefaultPasswordPolicy()
	if validationErrors := domain.ValidatePasswordStrength(req.Password, policy); len(validationErrors) > 0 {
		details := make(map[string]interface{})
		for i, e := range validationErrors {
			details[fmt.Sprintf("password_%d", i)] = e.Error()
		}
		return nil, application.ErrValidation("password does not meet requirements", details)
	}

	// Hash password
	passwordHash, err := uc.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, application.ErrInternal("failed to hash password", err)
	}

	password := domain.NewPasswordFromHash(passwordHash)

	// Create user entity
	user, err := domain.NewUser(req.TenantID, email, password, req.FirstName, req.LastName)
	if err != nil {
		return nil, application.ErrInternal("failed to create user", err)
	}

	// Assign default viewer role
	viewerRole, err := uc.roleRepo.FindByName(ctx, nil, domain.RoleNameViewer)
	if err == nil && viewerRole != nil {
		_ = user.AssignRole(viewerRole)
	}

	var verificationToken string

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Save user
		if err := uc.userRepo.Create(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
			outboxEntry := &domain.OutboxEntry{
				EventType:     event.EventType(),
				AggregateID:   event.AggregateID(),
				AggregateType: event.AggregateType(),
				Payload:       nil, // Will be serialized by repository
			}
			if err := uc.outboxRepo.Create(txCtx, outboxEntry); err != nil {
				return fmt.Errorf("failed to save outbox entry: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("registration failed", err)
	}

	// Clear domain events after successful save
	user.ClearDomainEvents()

	// Generate email verification token (async-safe)
	verificationToken, err = uc.verifyService.GenerateEmailVerificationToken(ctx, user.GetID())
	if err != nil {
		// Log error but don't fail registration
		// The user can request a new verification email
	}

	// Send welcome email with verification link (async)
	go func() {
		_ = uc.emailService.SendWelcomeEmail(context.Background(), email.String(), req.FirstName, verificationToken)
	}()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   req.TenantID,
		UserID:     nil,
		Action:     ports.AuditActionCreate,
		EntityType: "user",
		EntityID:   ptrUUID(user.GetID()),
		NewValues: map[string]interface{}{
			"email":      email.String(),
			"first_name": req.FirstName,
			"last_name":  req.LastName,
		},
	})

	// Reload user with roles for response
	user, _ = uc.userRepo.FindByID(ctx, user.GetID())

	return &dto.RegisterUserResponse{
		User: mapper.UserToDTO(user),
	}, nil
}

// RegisterUserWithTokens registers a user and returns auth tokens.
func (uc *RegisterUserUseCase) RegisterUserWithTokens(ctx context.Context, req *dto.RegisterUserRequest, ipAddress, userAgent string) (*dto.RegisterUserResponse, error) {
	// First register the user
	response, err := uc.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	// Generate tokens for immediate login
	claims := &ports.TokenClaims{
		UserID:      response.User.ID,
		TenantID:    response.User.TenantID,
		Email:       response.User.Email,
		Roles:       getRoleNames(response.User.Roles),
		Permissions: response.User.Permissions,
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(claims)
	if err != nil {
		// User is registered but token generation failed
		// Return user without tokens
		return response, nil
	}

	refreshToken, err := uc.tokenService.GenerateRefreshToken()
	if err != nil {
		return response, nil
	}

	response.AccessToken = accessToken
	response.RefreshToken = refreshToken
	response.ExpiresAt = claims.ExpiresAt

	return response, nil
}

// Helper functions

func ptrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}

func getRoleNames(roles []*dto.RoleDTO) []string {
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name
	}
	return names
}
