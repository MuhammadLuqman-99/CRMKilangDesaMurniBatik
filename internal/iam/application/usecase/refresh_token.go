// Package usecase contains the application use cases for the IAM service.
package usecase

import (
	"context"
	"time"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application/dto"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application/ports"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/domain"
)

// RefreshTokenUseCase handles token refresh operations.
type RefreshTokenUseCase struct {
	userRepo         domain.UserRepository
	tenantRepo       domain.TenantRepository
	roleRepo         domain.RoleRepository
	refreshTokenRepo domain.RefreshTokenRepository
	tokenService     ports.TokenService
	txManager        ports.TransactionManager
	rateLimiter      ports.RateLimiter
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase.
func NewRefreshTokenUseCase(
	userRepo domain.UserRepository,
	tenantRepo domain.TenantRepository,
	roleRepo domain.RoleRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	tokenService ports.TokenService,
	txManager ports.TransactionManager,
	rateLimiter ports.RateLimiter,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userRepo:         userRepo,
		tenantRepo:       tenantRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenService:     tokenService,
		txManager:        txManager,
		rateLimiter:      rateLimiter,
	}
}

// Execute refreshes the access token using a refresh token.
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req *dto.RefreshTokenRequest, ipAddress, userAgent string) (*dto.RefreshTokenResponse, error) {
	// Rate limiting by IP
	rateLimitKey := "refresh:" + ipAddress
	allowed, err := uc.rateLimiter.Allow(ctx, rateLimitKey)
	if err != nil || !allowed {
		return nil, application.ErrRateLimited()
	}

	// Hash the provided token
	tokenHash := domain.HashPlainToken(req.RefreshToken)

	// Find refresh token by hash
	storedToken, err := uc.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, application.ErrTokenInvalid()
	}

	// Validate token
	if !storedToken.IsValid() {
		if storedToken.IsRevoked() {
			// Token reuse detected - potential attack
			// Revoke all tokens for this user (security measure)
			_ = uc.refreshTokenRepo.RevokeByUserID(ctx, storedToken.UserID())
			return nil, application.NewAppError(
				application.ErrCodeTokenInvalid,
				"refresh token has been revoked - all sessions terminated for security",
				domain.ErrRefreshTokenReused,
			)
		}
		if storedToken.IsExpired() {
			return nil, application.ErrTokenExpired()
		}
		return nil, application.ErrTokenInvalid()
	}

	// Verify the plain token matches
	if !storedToken.Verify(req.RefreshToken) {
		return nil, application.ErrTokenInvalid()
	}

	// Get user
	user, err := uc.userRepo.FindByID(ctx, storedToken.UserID())
	if err != nil {
		return nil, application.ErrNotFound("user", storedToken.UserID())
	}

	// Check user can still login
	if !user.CanLogin() {
		return nil, application.ErrUserInactive()
	}

	// Get tenant
	tenant, err := uc.tenantRepo.FindByID(ctx, user.TenantID())
	if err != nil {
		return nil, application.ErrNotFound("tenant", user.TenantID())
	}

	// Check tenant is active
	if !tenant.IsActive() {
		return nil, application.ErrTenantInactive()
	}

	// Load user roles
	roles, err := uc.roleRepo.FindByUserID(ctx, user.GetID())
	if err == nil {
		user.SetRoles(roles)
	}

	// Rotate refresh token (create new, revoke old)
	expiresAt := time.Now().UTC().Add(uc.tokenService.GetRefreshTokenExpiry())
	newRefreshToken, newPlainToken, err := storedToken.Rotate(expiresAt)
	if err != nil {
		return nil, application.ErrInternal("failed to rotate refresh token", err)
	}

	// Generate new access token
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

	// Execute in transaction
	err = uc.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Update old token (mark as revoked)
		if err := uc.refreshTokenRepo.Update(txCtx, storedToken); err != nil {
			return err
		}

		// Save new refresh token
		if err := uc.refreshTokenRepo.Create(txCtx, newRefreshToken); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, application.ErrInternal("token refresh failed", err)
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newPlainToken,
		ExpiresAt:    time.Now().Add(uc.tokenService.GetAccessTokenExpiry()).Unix(),
		TokenType:    "Bearer",
	}, nil
}

// getRoleNames extracts role names from roles.
func (uc *RefreshTokenUseCase) getRoleNames(roles []*domain.Role) []string {
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = r.Name()
	}
	return names
}

// CleanupExpiredTokensUseCase handles cleanup of expired refresh tokens.
type CleanupExpiredTokensUseCase struct {
	refreshTokenRepo domain.RefreshTokenRepository
}

// NewCleanupExpiredTokensUseCase creates a new CleanupExpiredTokensUseCase.
func NewCleanupExpiredTokensUseCase(refreshTokenRepo domain.RefreshTokenRepository) *CleanupExpiredTokensUseCase {
	return &CleanupExpiredTokensUseCase{
		refreshTokenRepo: refreshTokenRepo,
	}
}

// Execute deletes all expired refresh tokens.
func (uc *CleanupExpiredTokensUseCase) Execute(ctx context.Context) (int64, error) {
	count, err := uc.refreshTokenRepo.DeleteExpired(ctx)
	if err != nil {
		return 0, application.ErrInternal("failed to cleanup expired tokens", err)
	}
	return count, nil
}
