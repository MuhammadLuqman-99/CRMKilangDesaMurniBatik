// ============================================
// Authentication Service
// Production-Ready Auth API Methods
// ============================================

import { api, tokenManager } from './api';
import type {
    LoginRequest,
    LoginResponse,
    RegisterRequest,
    ForgotPasswordRequest,
    ResetPasswordRequest,
    ChangePasswordRequest,
    User,
} from '../types';

export const authService = {
    /**
     * Login with email and password
     */
    login: async (credentials: LoginRequest): Promise<LoginResponse> => {
        const response = await api.post<LoginResponse>('/auth/login', credentials);
        tokenManager.setTokens(response.access_token, response.refresh_token);
        return response;
    },

    /**
     * Register a new user
     */
    register: async (data: RegisterRequest): Promise<{ message: string }> => {
        return api.post<{ message: string }>('/auth/register', data);
    },

    /**
     * Logout the current user
     */
    logout: async (): Promise<void> => {
        try {
            await api.post('/auth/logout');
        } finally {
            tokenManager.clearTokens();
        }
    },

    /**
     * Get current user profile
     */
    getProfile: async (): Promise<User> => {
        return api.get<User>('/auth/me');
    },

    /**
     * Update current user profile
     */
    updateProfile: async (data: Partial<User>): Promise<User> => {
        return api.put<User>('/auth/me', data);
    },

    /**
     * Request password reset email
     */
    forgotPassword: async (data: ForgotPasswordRequest): Promise<{ message: string }> => {
        return api.post<{ message: string }>('/auth/forgot-password', data);
    },

    /**
     * Reset password with token
     */
    resetPassword: async (data: ResetPasswordRequest): Promise<{ message: string }> => {
        return api.post<{ message: string }>('/auth/reset-password', data);
    },

    /**
     * Change current user's password
     */
    changePassword: async (data: ChangePasswordRequest): Promise<{ message: string }> => {
        return api.put<{ message: string }>('/auth/password', data);
    },

    /**
     * Verify email with token
     */
    verifyEmail: async (token: string): Promise<{ message: string }> => {
        return api.post<{ message: string }>('/auth/verify-email', { token });
    },

    /**
     * Resend verification email
     */
    resendVerification: async (): Promise<{ message: string }> => {
        return api.post<{ message: string }>('/auth/resend-verification');
    },

    /**
     * Check if user is authenticated
     */
    isAuthenticated: (): boolean => {
        return tokenManager.isAuthenticated();
    },

    /**
     * Clear authentication tokens
     */
    clearAuth: (): void => {
        tokenManager.clearTokens();
    },
};

export default authService;
