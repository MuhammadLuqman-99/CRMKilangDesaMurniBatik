// Package usecase contains the application use cases for the IAM service.
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application/dto"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application/mapper"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application/ports"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/domain"
)

// AuthenticateUserUseCase handles user authentication.
type AuthenticateUserUseCase struct {
	userRepo           domain.UserRepository
	tenantRepo         domain.TenantRepository
	roleRepo           domain.RoleRepository
	refreshTokenRepo   domain.RefreshTokenRepository
	passwordHasher     ports.PasswordHasher
	tokenService       ports.TokenService
	rateLimiter        ports.RateLimiter
	txManager          ports.TransactionManager
	auditLogger        ports.AuditLogger
	maxActiveTokens    int
}

// NewAuthenticateUserUseCase creates a new AuthenticateUserUseCase.
func NewAuthenticateUserUseCase(
	userRepo domain.UserRepository,
	tenantRepo domain.TenantRepository,
	roleRepo domain.RoleRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	passwordHasher ports.PasswordHasher,
	tokenService ports.TokenService,
	rateLimiter ports.RateLimiter,
	txManager ports.TransactionManager,
	auditLogger ports.AuditLogger,
) *AuthenticateUserUseCase {
	return &AuthenticateUserUseCase{
		userRepo:         userRepo,
		tenantRepo:       tenantRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordHasher:   passwordHasher,
		tokenService:     tokenService,
		rateLimiter:      rateLimiter,
		txManager:        txManager,
		auditLogger:      auditLogger,
		maxActiveTokens:  5, // Maximum active sessions per user
	}
}

// Execute authenticates a user and returns tokens.
func (uc *AuthenticateUserUseCase) Execute(ctx context.Context, req *dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error) {
	// Rate limiting by IP + tenant
	rateLimitKey := "auth:" + ipAddress + ":" + req.TenantSlug
	allowed, err := uc.rateLimiter.Allow(ctx, rateLimitKey)
	if err != nil || !allowed {
		return nil, application.ErrRateLimited()
	}

	// Find tenant by slug
	tenant, err := uc.tenantRepo.FindBySlug(ctx, req.TenantSlug)
	if err != nil {
		uc.logFailedLogin(ctx, uuid.Nil, nil, req.Email, ipAddress, userAgent, "tenant_not_found")
		return nil, application.ErrInvalidCredentials()
	}

	if !tenant.IsActive() {
		uc.logFailedLogin(ctx, tenant.GetID(), nil, req.Email, ipAddress, userAgent, "tenant_inactive")
		return nil, application.ErrTenantInactive()
	}

	// Create email value object
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		uc.logFailedLogin(ctx, tenant.GetID(), nil, req.Email, ipAddress, userAgent, "invalid_email_format")
		return nil, application.ErrInvalidCredentials()
	}

	// Find user by email in tenant
	user, err := uc.userRepo.FindByEmail(ctx, tenant.GetID(), email)
	if err != nil {
		uc.logFailedLogin(ctx, tenant.GetID(), nil, req.Email, ipAddress, userAgent, "user_not_found")
		return nil, application.ErrInvalidCredentials()
	}

	// Check user status
	if !user.CanLogin() {
		uc.logFailedLogin(ctx, tenant.GetID(), &user, req.Email, ipAddress, userAgent, "user_cannot_login")

		if user.IsDeleted() {
			return nil, application.ErrInvalidCredentials()
		}
		if user.Status() == domain.UserStatusSuspended {
			return nil, application.ErrUserInactive().WithDetail("reason", "account_suspended")
		}
		return nil, application.ErrUserInactive()
	}

	// Verify password
	valid, err := uc.passwordHasher.Verify(req.Password, user.PasswordHash().Hash())
	if err != nil || !valid {
		uc.logFailedLogin(ctx, tenant.GetID(), &user, req.Email, ipAddress, userAgent, "invalid_password")
		return nil, application.ErrInvalidCredentials()
	}

	// Load user roles
	roles, err := uc.roleRepo.FindByUserID(ctx, user.GetID())
	if err == nil {
		user.SetRoles(roles)
	}

	// Generate access token
	claims := &ports.TokenClaims{
		UserID:      user.GetID(),
		TenantID:    tenant.GetID(),
		Email:       user.Email().String(),
		Roles:       uc.getRoleNames(user.Roles()),
		Permissions: user.GetPermissions().Strings(),
	}

	accessToken, err := uc.tokenService.GenerateAccessToken(claims)
	if err != nil {
		return nil, application.ErrInternal("failed to generate access token", err)
	}

	// Create refresh token
	expiresAt := time.Now().UTC().Add(uc.tokenService.GetRefreshTokenExpiry())
	deviceInfo := mapper.DeviceInfoDTOToDomain(req.DeviceInfo)

	refreshToken, plainToken, err := domain.NewRefreshToken(
		user.GetID(),
		expiresAt,
		ipAddress,
		userAgent,
		deviceInfo,
	)
	if err != nil {
		return nil, application.ErrInternal("failed to generate refresh token", err)
	}

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Check active token count and revoke oldest if limit exceeded
		activeCount, err := uc.refreshTokenRepo.CountActiveByUserID(txCtx, user.GetID())
		if err != nil {
			return err
		}

		if activeCount >= int64(uc.maxActiveTokens) {
			// Get active tokens and revoke the oldest ones
			activeTokens, err := uc.refreshTokenRepo.FindActiveByUserID(txCtx, user.GetID())
			if err != nil {
				return err
			}

			// Revoke oldest tokens to make room
			tokensToRevoke := int(activeCount) - uc.maxActiveTokens + 1
			for i := 0; i < tokensToRevoke && i < len(activeTokens); i++ {
				if err := uc.refreshTokenRepo.RevokeByID(txCtx, activeTokens[i].GetID()); err != nil {
					return err
				}
			}
		}

		// Save refresh token
		if err := uc.refreshTokenRepo.Create(txCtx, refreshToken); err != nil {
			return err
		}

		// Record login
		user.RecordLogin()
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("authentication failed", err)
	}

	// Log successful login
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenant.GetID(),
		UserID:     ptrToUUID(user.GetID()),
		Action:     ports.AuditActionLogin,
		EntityType: "user",
		EntityID:   ptrToUUID(user.GetID()),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})

	return &dto.LoginResponse{
		User:         mapper.UserToDTO(user),
		AccessToken:  accessToken,
		RefreshToken: plainToken,
		ExpiresAt:    time.Now().Add(uc.tokenService.GetAccessTokenExpiry()).Unix(),
		TokenType:    "Bearer",
	}, nil
}

// logFailedLogin logs a failed login attempt.
func (uc *AuthenticateUserUseCase) logFailedLogin(ctx context.Context, tenantID uuid.UUID, user **domain.User, email, ipAddress, userAgent, reason string) {
	entry := ports.AuditEntry{
		TenantID:   tenantID,
		Action:     ports.AuditActionLoginFailed,
		EntityType: "user",
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		NewValues: map[string]interface{}{
			"email":  email,
			"reason": reason,
		},
	}

	if user != nil && *user != nil {
		userID := (*user).GetID()
		entry.UserID = &userID
		entry.EntityID = ptrToUUID(userID)
	}

	_ = uc.auditLogger.Log(ctx, entry)
}

// getRoleNames extracts role names from roles.
func (uc *AuthenticateUserUseCase) getRoleNames(roles []*domain.Role) []string {
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name()
	}
	return names
}

// ptrToUUID converts UUID to pointer.
func ptrToUUID(id uuid.UUID) *uuid.UUID {
	return &id
}

// LogoutUseCase handles user logout.
type LogoutUseCase struct {
	refreshTokenRepo domain.RefreshTokenRepository
	tokenBlacklist   ports.TokenBlacklist
	tokenService     ports.TokenService
	auditLogger      ports.AuditLogger
}

// NewLogoutUseCase creates a new LogoutUseCase.
func NewLogoutUseCase(
	refreshTokenRepo domain.RefreshTokenRepository,
	tokenBlacklist ports.TokenBlacklist,
	tokenService ports.TokenService,
	auditLogger ports.AuditLogger,
) *LogoutUseCase {
	return &LogoutUseCase{
		refreshTokenRepo: refreshTokenRepo,
		tokenBlacklist:   tokenBlacklist,
		tokenService:     tokenService,
		auditLogger:      auditLogger,
	}
}

// Execute logs out a user.
func (uc *LogoutUseCase) Execute(ctx context.Context, req *dto.LogoutRequest, accessToken string, userID, tenantID uuid.UUID) error {
	// Revoke refresh token
	tokenHash := domain.HashPlainToken(req.RefreshToken)
	refreshToken, err := uc.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return application.ErrTokenInvalid()
	}

	if refreshToken.UserID() != userID {
		return application.ErrForbidden("token does not belong to user")
	}

	if req.AllDevices {
		// Revoke all refresh tokens for user
		if err := uc.refreshTokenRepo.RevokeByUserID(ctx, userID); err != nil {
			return application.ErrInternal("failed to revoke all sessions", err)
		}
	} else {
		// Revoke single refresh token
		if err := uc.refreshTokenRepo.RevokeByID(ctx, refreshToken.GetID()); err != nil {
			return application.ErrInternal("failed to revoke session", err)
		}
	}

	// Blacklist access token
	if accessToken != "" {
		_ = uc.tokenBlacklist.Add(ctx, accessToken, uc.tokenService.GetAccessTokenExpiry())
	}

	// Log logout
	_ = uc.auditLogger.Log(ctx, ports.AuditEntry{
		TenantID:   tenantID,
		UserID:     &userID,
		Action:     ports.AuditActionLogout,
		EntityType: "user",
		EntityID:   &userID,
		NewValues: map[string]interface{}{
			"all_devices": req.AllDevices,
		},
	})

	return nil
}
