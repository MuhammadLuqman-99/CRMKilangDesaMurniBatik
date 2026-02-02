// ============================================
// API Service - Base Configuration
// Production-Ready HTTP Client
// ============================================

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

interface RequestConfig extends RequestInit {
    params?: Record<string, string | number | boolean | undefined>;
}

interface ApiErrorResponse {
    success: false;
    error: {
        code: string;
        message: string;
        details?: Array<{ field: string; message: string }>;
    };
}

export class ApiError extends Error {
    code: string;
    status: number;
    details?: Array<{ field: string; message: string }>;

    constructor(message: string, code: string, status: number, details?: Array<{ field: string; message: string }>) {
        super(message);
        this.name = 'ApiError';
        this.code = code;
        this.status = status;
        this.details = details;
    }
}

// Token management
const TOKEN_KEY = 'crm_access_token';
const REFRESH_TOKEN_KEY = 'crm_refresh_token';

export const tokenManager = {
    getAccessToken: (): string | null => {
        return localStorage.getItem(TOKEN_KEY);
    },

    getRefreshToken: (): string | null => {
        return localStorage.getItem(REFRESH_TOKEN_KEY);
    },

    setTokens: (accessToken: string, refreshToken: string): void => {
        localStorage.setItem(TOKEN_KEY, accessToken);
        localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
    },

    clearTokens: (): void => {
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(REFRESH_TOKEN_KEY);
    },

    isAuthenticated: (): boolean => {
        return !!localStorage.getItem(TOKEN_KEY);
    },
};

// Build URL with query parameters
function buildUrl(endpoint: string, params?: Record<string, string | number | boolean | undefined>): string {
    const url = new URL(`${API_BASE_URL}${endpoint}`);

    if (params) {
        Object.entries(params).forEach(([key, value]) => {
            if (value !== undefined && value !== null && value !== '') {
                url.searchParams.append(key, String(value));
            }
        });
    }

    return url.toString();
}

// Refresh token logic
let isRefreshing = false;
let refreshSubscribers: Array<(token: string) => void> = [];

function subscribeTokenRefresh(callback: (token: string) => void): void {
    refreshSubscribers.push(callback);
}

function onTokenRefreshed(token: string): void {
    refreshSubscribers.forEach((callback) => callback(token));
    refreshSubscribers = [];
}

async function refreshAccessToken(): Promise<string> {
    const refreshToken = tokenManager.getRefreshToken();

    if (!refreshToken) {
        throw new ApiError('No refresh token available', 'ERR_NO_REFRESH_TOKEN', 401);
    }

    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
        tokenManager.clearTokens();
        throw new ApiError('Session expired', 'ERR_SESSION_EXPIRED', 401);
    }

    const data = await response.json();
    tokenManager.setTokens(data.data.access_token, data.data.refresh_token);

    return data.data.access_token;
}

// Main request function
async function request<T>(endpoint: string, config: RequestConfig = {}): Promise<T> {
    const { params, ...fetchConfig } = config;
    const url = buildUrl(endpoint, params);

    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(config.headers as Record<string, string>),
    };

    const accessToken = tokenManager.getAccessToken();
    if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
    }

    let response = await fetch(url, {
        ...fetchConfig,
        headers,
    });

    // Handle token refresh on 401
    if (response.status === 401 && accessToken && !endpoint.includes('/auth/')) {
        if (!isRefreshing) {
            isRefreshing = true;

            try {
                const newToken = await refreshAccessToken();
                isRefreshing = false;
                onTokenRefreshed(newToken);

                // Retry original request with new token
                headers['Authorization'] = `Bearer ${newToken}`;
                response = await fetch(url, {
                    ...fetchConfig,
                    headers,
                });
            } catch (error) {
                isRefreshing = false;
                tokenManager.clearTokens();
                window.location.href = '/login';
                throw error;
            }
        } else {
            // Wait for token refresh
            return new Promise((resolve, reject) => {
                subscribeTokenRefresh(async (newToken: string) => {
                    try {
                        headers['Authorization'] = `Bearer ${newToken}`;
                        const retryResponse = await fetch(url, {
                            ...fetchConfig,
                            headers,
                        });
                        const data = await retryResponse.json();
                        resolve(data.data);
                    } catch (error) {
                        reject(error);
                    }
                });
            });
        }
    }

    const data = await response.json();

    if (!response.ok) {
        const errorData = data as ApiErrorResponse;
        throw new ApiError(
            errorData.error?.message || 'An error occurred',
            errorData.error?.code || 'ERR_UNKNOWN',
            response.status,
            errorData.error?.details
        );
    }

    return data.data as T;
}

// HTTP method helpers
export const api = {
    get: <T>(endpoint: string, params?: Record<string, string | number | boolean | undefined>): Promise<T> => {
        return request<T>(endpoint, { method: 'GET', params });
    },

    post: <T>(endpoint: string, body?: unknown): Promise<T> => {
        return request<T>(endpoint, {
            method: 'POST',
            body: body ? JSON.stringify(body) : undefined,
        });
    },

    put: <T>(endpoint: string, body?: unknown): Promise<T> => {
        return request<T>(endpoint, {
            method: 'PUT',
            body: body ? JSON.stringify(body) : undefined,
        });
    },

    patch: <T>(endpoint: string, body?: unknown): Promise<T> => {
        return request<T>(endpoint, {
            method: 'PATCH',
            body: body ? JSON.stringify(body) : undefined,
        });
    },

    delete: <T>(endpoint: string): Promise<T> => {
        return request<T>(endpoint, { method: 'DELETE' });
    },
};

export { API_BASE_URL };
