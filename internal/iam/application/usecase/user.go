// Package usecase contains the application use cases for the IAM service.
package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/kilang-desa-murni/crm/internal/iam/application"
	"github.com/kilang-desa-murni/crm/internal/iam/application/dto"
	"github.com/kilang-desa-murni/crm/internal/iam/application/mapper"
	"github.com/kilang-desa-murni/crm/internal/iam/application/ports"
	"github.com/kilang-desa-murni/crm/internal/iam/domain"
)

// ============================================================================
// Query Use Cases
// ============================================================================

// GetUserUseCase handles retrieving a user by ID.
type GetUserUseCase struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
}

// NewGetUserUseCase creates a new GetUserUseCase.
func NewGetUserUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute retrieves a user by ID.
func (uc *GetUserUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID) (*dto.UserDTO, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, application.ErrNotFound("user", userID)
	}

	// Verify user belongs to tenant
	if user.TenantID() != tenantID {
		return nil, application.ErrForbidden("user does not belong to this tenant")
	}

	// Load roles
	roles, err := uc.roleRepo.FindByUserID(ctx, userID)
	if err == nil {
		user.SetRoles(roles)
	}

	return mapper.UserToDTO(user), nil
}

// GetUserByEmailUseCase handles retrieving a user by email.
type GetUserByEmailUseCase struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
}

// NewGetUserByEmailUseCase creates a new GetUserByEmailUseCase.
func NewGetUserByEmailUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
) *GetUserByEmailUseCase {
	return &GetUserByEmailUseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute retrieves a user by email.
func (uc *GetUserByEmailUseCase) Execute(ctx context.Context, email string, tenantID uuid.UUID) (*dto.UserDTO, error) {
	emailVO, err := domain.NewEmail(email)
	if err != nil {
		return nil, application.ErrValidation("invalid email", map[string]interface{}{
			"email": err.Error(),
		})
	}

	user, err := uc.userRepo.FindByEmail(ctx, tenantID, emailVO)
	if err != nil {
		return nil, application.ErrNotFound("user", email)
	}

	// Load roles
	roles, err := uc.roleRepo.FindByUserID(ctx, user.GetID())
	if err == nil {
		user.SetRoles(roles)
	}

	return mapper.UserToDTO(user), nil
}

// ListUsersUseCase handles listing users with pagination.
type ListUsersUseCase struct {
	userRepo domain.UserRepository
	roleRepo domain.RoleRepository
}

// NewListUsersUseCase creates a new ListUsersUseCase.
func NewListUsersUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
) *ListUsersUseCase {
	return &ListUsersUseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Execute lists users with pagination.
func (uc *ListUsersUseCase) Execute(ctx context.Context, req *dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
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

	opts := domain.UserQueryOptions{
		Page:          req.Page,
		PageSize:      req.PageSize,
		SortBy:        req.SortBy,
		SortDirection: req.SortDirection,
		Search:        req.Search,
		IncludeRoles:  req.IncludeRoles,
	}

	if req.Status != "" {
		status := domain.UserStatus(req.Status)
		opts.Status = &status
	}

	users, total, err := uc.userRepo.FindByTenant(ctx, req.TenantID, opts)
	if err != nil {
		return nil, application.ErrInternal("failed to list users", err)
	}

	// Load roles if requested
	if req.IncludeRoles {
		for _, user := range users {
			roles, err := uc.roleRepo.FindByUserID(ctx, user.GetID())
			if err == nil {
				user.SetRoles(roles)
			}
		}
	}

	return &dto.ListUsersResponse{
		Users:      mapper.UsersToDTO(users),
		Pagination: dto.NewPaginationDTO(req.Page, req.PageSize, total),
	}, nil
}

// ============================================================================
// Mutation Use Cases
// ============================================================================

// UpdateUserUseCase handles updating a user's profile.
type UpdateUserUseCase struct {
	userRepo    domain.UserRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewUpdateUserUseCase creates a new UpdateUserUseCase.
func NewUpdateUserUseCase(
	userRepo domain.UserRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *UpdateUserUseCase {
	return &UpdateUserUseCase{
		userRepo:    userRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute updates a user's profile.
func (uc *UpdateUserUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID, req *dto.UpdateUserRequest, updatedBy *uuid.UUID) (*dto.UserDTO, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, application.ErrNotFound("user", userID)
	}

	// Verify user belongs to tenant
	if user.TenantID() != tenantID {
		return nil, application.ErrForbidden("user does not belong to this tenant")
	}

	oldValues := map[string]interface{}{
		"first_name": user.FirstName(),
		"last_name":  user.LastName(),
		"phone":      user.Phone(),
		"avatar_url": user.AvatarURL(),
	}

	// Update profile
	user.UpdateProfile(req.FirstName, req.LastName, req.Phone, req.AvatarURL)

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return nil, application.ErrInternal("failed to update user", err)
	}

	user.ClearDomainEvents()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     updatedBy,
		Action:     ports.AuditActionUpdate,
		EntityType: "user",
		EntityID:   ptrToUUID(userID),
		OldValues:  oldValues,
		NewValues: map[string]interface{}{
			"first_name": req.FirstName,
			"last_name":  req.LastName,
			"phone":      req.Phone,
			"avatar_url": req.AvatarURL,
		},
	})

	return mapper.UserToDTO(user), nil
}

// UpdateUserEmailUseCase handles updating a user's email.
type UpdateUserEmailUseCase struct {
	userRepo       domain.UserRepository
	outboxRepo     domain.OutboxRepository
	passwordHasher ports.PasswordHasher
	verifyService  ports.VerificationTokenService
	emailService   ports.EmailService
	txManager      ports.TransactionManager
	auditLogger    ports.AuditLogger
}

// NewUpdateUserEmailUseCase creates a new UpdateUserEmailUseCase.
func NewUpdateUserEmailUseCase(
	userRepo domain.UserRepository,
	outboxRepo domain.OutboxRepository,
	passwordHasher ports.PasswordHasher,
	verifyService ports.VerificationTokenService,
	emailService ports.EmailService,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *UpdateUserEmailUseCase {
	return &UpdateUserEmailUseCase{
		userRepo:       userRepo,
		outboxRepo:     outboxRepo,
		passwordHasher: passwordHasher,
		verifyService:  verifyService,
		emailService:   emailService,
		txManager:      txManager,
		auditLogger:    auditLogger,
	}
}

// Execute updates a user's email.
func (uc *UpdateUserEmailUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID, req *dto.UpdateUserEmailRequest) (*dto.UserDTO, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, application.ErrNotFound("user", userID)
	}

	// Verify user belongs to tenant
	if user.TenantID() != tenantID {
		return nil, application.ErrForbidden("user does not belong to this tenant")
	}

	// Verify password
	valid, err := uc.passwordHasher.Verify(req.Password, user.PasswordHash().Hash())
	if err != nil || !valid {
		return nil, application.ErrForbidden("invalid password")
	}

	// Create new email value object
	newEmail, err := domain.NewEmail(req.NewEmail)
	if err != nil {
		return nil, application.ErrValidation("invalid email format", map[string]interface{}{
			"email": err.Error(),
		})
	}

	// Check if new email already exists
	exists, err := uc.userRepo.ExistsByEmail(ctx, tenantID, newEmail)
	if err != nil {
		return nil, application.ErrInternal("failed to check email", err)
	}

	if exists {
		return nil, application.ErrConflict("email already in use")
	}

	oldEmail := user.Email()

	// Update email
	if err := user.UpdateEmail(newEmail); err != nil {
		return nil, application.ErrInternal("failed to update email", err)
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return nil, application.ErrInternal("failed to update email", err)
	}

	user.ClearDomainEvents()

	// Generate verification token and send email
	verificationToken, err := uc.verifyService.GenerateEmailVerificationToken(ctx, userID)
	if err == nil {
		go func() {
			_ = uc.emailService.SendEmailVerificationEmail(context.Background(), newEmail.String(), user.FirstName(), verificationToken)
		}()
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     ptrToUUID(userID),
		Action:     ports.AuditActionUpdate,
		EntityType: "user_email",
		EntityID:   ptrToUUID(userID),
		OldValues:  map[string]interface{}{"email": oldEmail.String()},
		NewValues:  map[string]interface{}{"email": newEmail.String()},
	})

	return mapper.UserToDTO(user), nil
}

// ChangePasswordUseCase handles changing a user's password.
type ChangePasswordUseCase struct {
	userRepo       domain.UserRepository
	outboxRepo     domain.OutboxRepository
	passwordHasher ports.PasswordHasher
	emailService   ports.EmailService
	txManager      ports.TransactionManager
	auditLogger    ports.AuditLogger
}

// NewChangePasswordUseCase creates a new ChangePasswordUseCase.
func NewChangePasswordUseCase(
	userRepo domain.UserRepository,
	outboxRepo domain.OutboxRepository,
	passwordHasher ports.PasswordHasher,
	emailService ports.EmailService,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		userRepo:       userRepo,
		outboxRepo:     outboxRepo,
		passwordHasher: passwordHasher,
		emailService:   emailService,
		txManager:      txManager,
		auditLogger:    auditLogger,
	}
}

// Execute changes a user's password.
func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID, req *dto.ChangePasswordRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return application.ErrNotFound("user", userID)
	}

	// Verify user belongs to tenant
	if user.TenantID() != tenantID {
		return application.ErrForbidden("user does not belong to this tenant")
	}

	// Verify current password
	valid, err := uc.passwordHasher.Verify(req.CurrentPassword, user.PasswordHash().Hash())
	if err != nil || !valid {
		return application.ErrForbidden("current password is incorrect")
	}

	// Validate new password policy
	policy := domain.DefaultPasswordPolicy()
	if validationErrors := domain.ValidatePasswordStrength(req.NewPassword, policy); len(validationErrors) > 0 {
		details := make(map[string]interface{})
		for i, e := range validationErrors {
			details[string(rune('a'+i))] = e.Error()
		}
		return application.ErrValidation("password does not meet requirements", details)
	}

	// Hash new password
	newPasswordHash, err := uc.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		return application.ErrInternal("failed to hash password", err)
	}

	newPassword := domain.NewPasswordFromHash(newPasswordHash)

	// Change password
	if err := user.ChangePassword(newPassword); err != nil {
		return application.ErrInternal("failed to change password", err)
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return application.ErrInternal("failed to change password", err)
	}

	user.ClearDomainEvents()

	// Send notification email
	go func() {
		_ = uc.emailService.SendPasswordChangedEmail(context.Background(), user.Email().String(), user.FirstName())
	}()

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     ptrToUUID(userID),
		Action:     ports.AuditActionPasswordChanged,
		EntityType: "user",
		EntityID:   ptrToUUID(userID),
	})

	return nil
}

// VerifyEmailUseCase handles email verification.
type VerifyEmailUseCase struct {
	userRepo      domain.UserRepository
	outboxRepo    domain.OutboxRepository
	verifyService ports.VerificationTokenService
	txManager     ports.TransactionManager
	auditLogger   ports.AuditLogger
}

// NewVerifyEmailUseCase creates a new VerifyEmailUseCase.
func NewVerifyEmailUseCase(
	userRepo domain.UserRepository,
	outboxRepo domain.OutboxRepository,
	verifyService ports.VerificationTokenService,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *VerifyEmailUseCase {
	return &VerifyEmailUseCase{
		userRepo:      userRepo,
		outboxRepo:    outboxRepo,
		verifyService: verifyService,
		txManager:     txManager,
		auditLogger:   auditLogger,
	}
}

// Execute verifies a user's email.
func (uc *VerifyEmailUseCase) Execute(ctx context.Context, req *dto.VerifyEmailRequest) error {
	// Validate token and get user ID
	userID, err := uc.verifyService.ValidateEmailVerificationToken(ctx, req.Token)
	if err != nil {
		return application.ErrTokenInvalid()
	}

	// Get user
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return application.ErrNotFound("user", userID)
	}

	// Check if already verified
	if user.IsEmailVerified() {
		return nil
	}

	// Verify email
	user.VerifyEmail()

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		// Invalidate token
		_ = uc.verifyService.InvalidateToken(txCtx, req.Token)

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return application.ErrInternal("failed to verify email", err)
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   user.TenantID(),
		UserID:     ptrToUUID(userID),
		Action:     "email_verified",
		EntityType: "user",
		EntityID:   ptrToUUID(userID),
	})

	return nil
}

// DeleteUserUseCase handles soft deleting a user.
type DeleteUserUseCase struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	outboxRepo       domain.OutboxRepository
	txManager        ports.TransactionManager
	auditLogger      ports.AuditLogger
}

// NewDeleteUserUseCase creates a new DeleteUserUseCase.
func NewDeleteUserUseCase(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *DeleteUserUseCase {
	return &DeleteUserUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		outboxRepo:       outboxRepo,
		txManager:        txManager,
		auditLogger:      auditLogger,
	}
}

// Execute soft deletes a user.
func (uc *DeleteUserUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID, deletedBy *uuid.UUID) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return application.ErrNotFound("user", userID)
	}

	// Verify user belongs to tenant
	if user.TenantID() != tenantID {
		return application.ErrForbidden("user does not belong to this tenant")
	}

	// Delete user
	user.Delete()

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Revoke all refresh tokens
		if err := uc.refreshTokenRepo.RevokeByUserID(txCtx, userID); err != nil {
			return err
		}

		// Soft delete user
		if err := uc.userRepo.Delete(txCtx, userID); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return application.ErrInternal("failed to delete user", err)
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     deletedBy,
		Action:     ports.AuditActionDelete,
		EntityType: "user",
		EntityID:   ptrToUUID(userID),
		OldValues: map[string]interface{}{
			"email": user.Email().String(),
		},
	})

	return nil
}

// SuspendUserUseCase handles suspending a user.
type SuspendUserUseCase struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	outboxRepo       domain.OutboxRepository
	txManager        ports.TransactionManager
	auditLogger      ports.AuditLogger
}

// NewSuspendUserUseCase creates a new SuspendUserUseCase.
func NewSuspendUserUseCase(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *SuspendUserUseCase {
	return &SuspendUserUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		outboxRepo:       outboxRepo,
		txManager:        txManager,
		auditLogger:      auditLogger,
	}
}

// Execute suspends a user.
func (uc *SuspendUserUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID, reason string, suspendedBy *uuid.UUID) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return application.ErrNotFound("user", userID)
	}

	// Verify user belongs to tenant
	if user.TenantID() != tenantID {
		return application.ErrForbidden("user does not belong to this tenant")
	}

	// Suspend user
	if err := user.Suspend(reason); err != nil {
		return application.ErrConflict(err.Error())
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Revoke all refresh tokens
		if err := uc.refreshTokenRepo.RevokeByUserID(txCtx, userID); err != nil {
			return err
		}

		// Update user
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return application.ErrInternal("failed to suspend user", err)
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     suspendedBy,
		Action:     "suspend",
		EntityType: "user",
		EntityID:   ptrToUUID(userID),
		NewValues:  map[string]interface{}{"reason": reason},
	})

	return nil
}

// ActivateUserUseCase handles activating a user.
type ActivateUserUseCase struct {
	userRepo    domain.UserRepository
	outboxRepo  domain.OutboxRepository
	txManager   ports.TransactionManager
	auditLogger ports.AuditLogger
}

// NewActivateUserUseCase creates a new ActivateUserUseCase.
func NewActivateUserUseCase(
	userRepo domain.UserRepository,
	outboxRepo domain.OutboxRepository,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *ActivateUserUseCase {
	return &ActivateUserUseCase{
		userRepo:    userRepo,
		outboxRepo:  outboxRepo,
		txManager:   txManager,
		auditLogger: auditLogger,
	}
}

// Execute activates a user.
func (uc *ActivateUserUseCase) Execute(ctx context.Context, userID, tenantID uuid.UUID, activatedBy *uuid.UUID) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return application.ErrNotFound("user", userID)
	}

	// Verify user belongs to tenant
	if user.TenantID() != tenantID {
		return application.ErrForbidden("user does not belong to this tenant")
	}

	// Activate user
	if err := user.Activate(); err != nil {
		return application.ErrConflict(err.Error())
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		// Save domain events to outbox
		for _, event := range user.GetDomainEvents() {
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
		return application.ErrInternal("failed to activate user", err)
	}

	// Log audit
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     activatedBy,
		Action:     "activate",
		EntityType: "user",
		EntityID:   ptrToUUID(userID),
	})

	return nil
}
