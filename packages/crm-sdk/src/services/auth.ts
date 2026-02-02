// ============================================
// Authentication Service
// Login, registration, password management
// ============================================

import { HttpClient } from '../http';
import type {
    User,
    LoginRequest,
    LoginResponse,
    RegisterRequest,
    ForgotPasswordRequest,
    ResetPasswordRequest,
    ChangePasswordRequest,
    ProfileUpdateRequest,
} from '../types';

export class AuthService {
    constructor(private http: HttpClient) {}

    /**
     * Login with email and password
     * @returns User data and authentication tokens
     */
    async login(credentials: LoginRequest): Promise<LoginResponse> {
        const response = await this.http.request<LoginResponse>({
            method: 'POST',
            path: '/auth/login',
            body: credentials,
            skipAuth: true,
        });

        // Store tokens in HTTP client
        this.http.setTokens(response.accessToken, response.refreshToken);

        return response;
    }

    /**
     * Register a new user account
     */
    async register(data: RegisterRequest): Promise<{ message: string }> {
        return this.http.request<{ message: string }>({
            method: 'POST',
            path: '/auth/register',
            body: data,
            skipAuth: true,
        });
    }

    /**
     * Logout and invalidate tokens
     */
    async logout(): Promise<void> {
        try {
            await this.http.post('/auth/logout');
        } finally {
            this.http.clearTokens();
        }
    }

    /**
     * Get the current user's profile
     */
    async getProfile(): Promise<User> {
        return this.http.get<User>('/auth/me');
    }

    /**
     * Update the current user's profile
     */
    async updateProfile(data: ProfileUpdateRequest): Promise<User> {
        return this.http.put<User>('/auth/me', data);
    }

    /**
     * Request a password reset email
     */
    async forgotPassword(data: ForgotPasswordRequest): Promise<{ message: string }> {
        return this.http.request<{ message: string }>({
            method: 'POST',
            path: '/auth/forgot-password',
            body: data,
            skipAuth: true,
        });
    }

    /**
     * Reset password using token from email
     */
    async resetPassword(data: ResetPasswordRequest): Promise<{ message: string }> {
        return this.http.request<{ message: string }>({
            method: 'POST',
            path: '/auth/reset-password',
            body: data,
            skipAuth: true,
        });
    }

    /**
     * Change current user's password
     */
    async changePassword(data: ChangePasswordRequest): Promise<{ message: string }> {
        return this.http.put<{ message: string }>('/auth/password', data);
    }

    /**
     * Verify email address using token
     */
    async verifyEmail(token: string): Promise<{ message: string }> {
        return this.http.request<{ message: string }>({
            method: 'POST',
            path: '/auth/verify-email',
            body: { token },
            skipAuth: true,
        });
    }

    /**
     * Resend email verification
     */
    async resendVerification(): Promise<{ message: string }> {
        return this.http.post<{ message: string }>('/auth/resend-verification');
    }

    /**
     * Check if the client is authenticated
     */
    isAuthenticated(): boolean {
        return this.http.isAuthenticated();
    }

    /**
     * Set authentication tokens manually (e.g., from storage)
     */
    setTokens(accessToken: string, refreshToken?: string): void {
        this.http.setTokens(accessToken, refreshToken);
    }

    /**
     * Clear authentication tokens
     */
    clearTokens(): void {
        this.http.clearTokens();
    }
}
