// Package handler contains HTTP handlers for the IAM service.
package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application/dto"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/application/usecase"
	iamhttp "github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/interfaces/http"
	"github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik/internal/iam/interfaces/http/middleware"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	registerUC     *usecase.RegisterUserUseCase
	authenticateUC *usecase.AuthenticateUserUseCase
	refreshTokenUC *usecase.RefreshTokenUseCase
	decoder        *iamhttp.RequestDecoder
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	registerUC *usecase.RegisterUserUseCase,
	authenticateUC *usecase.AuthenticateUserUseCase,
	refreshTokenUC *usecase.RefreshTokenUseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUC:     registerUC,
		authenticateUC: authenticateUC,
		refreshTokenUC: refreshTokenUC,
		decoder:        iamhttp.NewRequestDecoder(),
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	TenantID  string `json:"tenant_id" validate:"required,uuid"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,max=20"`
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	tenantID, _ := uuid.Parse(req.TenantID)

	result, err := h.registerUC.Execute(r.Context(), &usecase.RegisterUserRequest{
		TenantID:  tenantID,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	response := map[string]interface{}{
		"user":          result.User,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
	}

	iamhttp.WriteSuccess(w, http.StatusCreated, response)
}

// LoginRequest represents a login request.
type LoginRequest struct {
	TenantID string `json:"tenant_id" validate:"required,uuid"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	tenantID, _ := uuid.Parse(req.TenantID)

	// Get device info from request
	deviceInfo := extractDeviceInfo(r)

	result, err := h.authenticateUC.Execute(r.Context(), &usecase.AuthenticateUserRequest{
		TenantID:   tenantID,
		Email:      req.Email,
		Password:   req.Password,
		IPAddress:  getClientIP(r),
		UserAgent:  r.UserAgent(),
		DeviceInfo: deviceInfo,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	response := map[string]interface{}{
		"user":          result.User,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
	}

	iamhttp.WriteSuccess(w, http.StatusOK, response)
}

// RefreshTokenRequest represents a token refresh request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken handles token refresh.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	result, err := h.refreshTokenUC.Execute(r.Context(), &usecase.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
		IPAddress:    getClientIP(r),
		UserAgent:    r.UserAgent(),
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	response := map[string]interface{}{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
	}

	iamhttp.WriteSuccess(w, http.StatusOK, response)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		iamhttp.WriteError(w, http.StatusUnauthorized, iamhttp.ErrCodeUnauthorized, "authentication required", nil)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token,omitempty"`
		AllDevices   bool   `json:"all_devices,omitempty"`
	}

	// Decode request body (optional)
	_ = h.decoder.Decode(r, &req)

	err := h.authenticateUC.Logout(r.Context(), &usecase.LogoutRequest{
		UserID:       userID,
		RefreshToken: req.RefreshToken,
		AllDevices:   req.AllDevices,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

// Me returns the current authenticated user's information.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetTokenClaims(r.Context())
	if claims == nil {
		iamhttp.WriteError(w, http.StatusUnauthorized, iamhttp.ErrCodeUnauthorized, "authentication required", nil)
		return
	}

	response := dto.UserDTO{
		ID:    claims.UserID,
		Email: claims.Email,
	}

	iamhttp.WriteSuccess(w, http.StatusOK, response)
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=128"`
}

// ChangePassword handles password change.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		iamhttp.WriteError(w, http.StatusUnauthorized, iamhttp.ErrCodeUnauthorized, "authentication required", nil)
		return
	}

	var req ChangePasswordRequest
	if err := h.decoder.Decode(r, &req); err != nil {
		iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, "invalid request", iamhttp.ValidationErrors(err))
		return
	}

	err := h.authenticateUC.ChangePassword(r.Context(), &usecase.ChangePasswordRequest{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})

	if err != nil {
		handleApplicationError(w, err)
		return
	}

	iamhttp.WriteSuccess(w, http.StatusOK, map[string]string{
		"message": "password changed successfully",
	})
}

// Helper functions

func extractDeviceInfo(r *http.Request) map[string]string {
	deviceInfo := make(map[string]string)

	if deviceID := r.Header.Get("X-Device-ID"); deviceID != "" {
		deviceInfo["device_id"] = deviceID
	}
	if deviceType := r.Header.Get("X-Device-Type"); deviceType != "" {
		deviceInfo["device_type"] = deviceType
	}
	if deviceName := r.Header.Get("X-Device-Name"); deviceName != "" {
		deviceInfo["device_name"] = deviceName
	}

	return deviceInfo
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func handleApplicationError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *application.AppError:
		switch e.Type {
		case application.ErrTypeValidation:
			iamhttp.WriteError(w, http.StatusBadRequest, iamhttp.ErrCodeValidation, e.Message, e.Details)
		case application.ErrTypeNotFound:
			iamhttp.WriteError(w, http.StatusNotFound, iamhttp.ErrCodeNotFound, e.Message, e.Details)
		case application.ErrTypeUnauthorized:
			iamhttp.WriteError(w, http.StatusUnauthorized, iamhttp.ErrCodeUnauthorized, e.Message, e.Details)
		case application.ErrTypeForbidden:
			iamhttp.WriteError(w, http.StatusForbidden, iamhttp.ErrCodeForbidden, e.Message, e.Details)
		case application.ErrTypeConflict:
			iamhttp.WriteError(w, http.StatusConflict, iamhttp.ErrCodeConflict, e.Message, e.Details)
		default:
			iamhttp.WriteError(w, http.StatusInternalServerError, iamhttp.ErrCodeInternalServer, "internal server error", nil)
		}
	default:
		iamhttp.WriteError(w, http.StatusInternalServerError, iamhttp.ErrCodeInternalServer, "internal server error", nil)
	}
}
