// ============================================
// HTTP Client
// Core HTTP layer with authentication
// ============================================

import { CRMClientConfig, CRMError, CRMErrorResponse, TokenPair } from './types';

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

interface RequestOptions {
    method: HttpMethod;
    path: string;
    body?: unknown;
    params?: Record<string, string | number | boolean | undefined>;
    headers?: Record<string, string>;
    skipAuth?: boolean;
}

/**
 * HTTP client for making authenticated API requests
 */
export class HttpClient {
    private config: Required<Omit<CRMClientConfig, 'onTokenRefresh' | 'onAuthError'>> &
        Pick<CRMClientConfig, 'onTokenRefresh' | 'onAuthError'>;
    private accessToken: string | null = null;
    private refreshToken: string | null = null;
    private isRefreshing = false;
    private refreshQueue: Array<{
        resolve: (token: string) => void;
        reject: (error: Error) => void;
    }> = [];

    constructor(config: CRMClientConfig) {
        this.config = {
            baseUrl: config.baseUrl.replace(/\/$/, ''),
            apiVersion: config.apiVersion || 'v1',
            timeout: config.timeout || 30000,
            headers: config.headers || {},
            onTokenRefresh: config.onTokenRefresh,
            onAuthError: config.onAuthError,
        };
    }

    /**
     * Set authentication tokens
     */
    setTokens(accessToken: string, refreshToken?: string): void {
        this.accessToken = accessToken;
        if (refreshToken) {
            this.refreshToken = refreshToken;
        }
    }

    /**
     * Get current access token
     */
    getAccessToken(): string | null {
        return this.accessToken;
    }

    /**
     * Clear authentication tokens
     */
    clearTokens(): void {
        this.accessToken = null;
        this.refreshToken = null;
    }

    /**
     * Check if client is authenticated
     */
    isAuthenticated(): boolean {
        return this.accessToken !== null;
    }

    /**
     * Build full URL with query parameters
     */
    private buildUrl(path: string, params?: Record<string, string | number | boolean | undefined>): string {
        const url = new URL(`${this.config.baseUrl}/api/${this.config.apiVersion}${path}`);

        if (params) {
            Object.entries(params).forEach(([key, value]) => {
                if (value !== undefined && value !== null && value !== '') {
                    url.searchParams.append(this.toSnakeCase(key), String(value));
                }
            });
        }

        return url.toString();
    }

    /**
     * Convert camelCase to snake_case
     */
    private toSnakeCase(str: string): string {
        return str.replace(/[A-Z]/g, letter => `_${letter.toLowerCase()}`);
    }

    /**
     * Convert snake_case to camelCase
     */
    private toCamelCase(str: string): string {
        return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
    }

    /**
     * Deep convert object keys to camelCase
     */
    private keysToCamelCase<T>(obj: unknown): T {
        if (Array.isArray(obj)) {
            return obj.map(item => this.keysToCamelCase(item)) as T;
        }

        if (obj !== null && typeof obj === 'object') {
            return Object.keys(obj).reduce((acc, key) => {
                const camelKey = this.toCamelCase(key);
                (acc as Record<string, unknown>)[camelKey] = this.keysToCamelCase(
                    (obj as Record<string, unknown>)[key]
                );
                return acc;
            }, {} as T);
        }

        return obj as T;
    }

    /**
     * Deep convert object keys to snake_case
     */
    private keysToSnakeCase(obj: unknown): unknown {
        if (Array.isArray(obj)) {
            return obj.map(item => this.keysToSnakeCase(item));
        }

        if (obj !== null && typeof obj === 'object') {
            return Object.keys(obj).reduce((acc, key) => {
                const snakeKey = this.toSnakeCase(key);
                acc[snakeKey] = this.keysToSnakeCase((obj as Record<string, unknown>)[key]);
                return acc;
            }, {} as Record<string, unknown>);
        }

        return obj;
    }

    /**
     * Refresh access token
     */
    private async refreshAccessToken(): Promise<string> {
        if (!this.refreshToken) {
            throw new CRMError('No refresh token available', 'NO_REFRESH_TOKEN', 401);
        }

        if (this.isRefreshing) {
            return new Promise((resolve, reject) => {
                this.refreshQueue.push({ resolve, reject });
            });
        }

        this.isRefreshing = true;

        try {
            const response = await fetch(`${this.config.baseUrl}/api/${this.config.apiVersion}/auth/refresh`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ refresh_token: this.refreshToken }),
            });

            if (!response.ok) {
                throw new CRMError('Failed to refresh token', 'TOKEN_REFRESH_FAILED', response.status);
            }

            const data = await response.json();
            const tokens: TokenPair = {
                accessToken: data.access_token,
                refreshToken: data.refresh_token,
                expiresIn: data.expires_in,
            };

            this.accessToken = tokens.accessToken;
            this.refreshToken = tokens.refreshToken;

            // Notify callback
            if (this.config.onTokenRefresh) {
                this.config.onTokenRefresh(tokens);
            }

            // Resolve queued requests
            this.refreshQueue.forEach(({ resolve }) => resolve(tokens.accessToken));
            this.refreshQueue = [];

            return tokens.accessToken;
        } catch (error) {
            // Reject queued requests
            this.refreshQueue.forEach(({ reject }) => reject(error as Error));
            this.refreshQueue = [];

            this.clearTokens();

            if (this.config.onAuthError && error instanceof CRMError) {
                this.config.onAuthError(error);
            }

            throw error;
        } finally {
            this.isRefreshing = false;
        }
    }

    /**
     * Make an HTTP request
     */
    async request<T>(options: RequestOptions): Promise<T> {
        const { method, path, body, params, headers, skipAuth } = options;

        const url = this.buildUrl(path, params);

        const requestHeaders: Record<string, string> = {
            'Content-Type': 'application/json',
            ...this.config.headers,
            ...headers,
        };

        if (!skipAuth && this.accessToken) {
            requestHeaders['Authorization'] = `Bearer ${this.accessToken}`;
        }

        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);

        try {
            const response = await fetch(url, {
                method,
                headers: requestHeaders,
                body: body ? JSON.stringify(this.keysToSnakeCase(body)) : undefined,
                signal: controller.signal,
            });

            clearTimeout(timeoutId);

            // Handle 401 - try to refresh token
            if (response.status === 401 && !skipAuth && this.refreshToken) {
                try {
                    await this.refreshAccessToken();
                    // Retry the request with new token
                    return this.request({ ...options, skipAuth: false });
                } catch {
                    throw new CRMError('Authentication failed', 'UNAUTHORIZED', 401);
                }
            }

            // Handle error responses
            if (!response.ok) {
                let errorData: CRMErrorResponse;
                try {
                    errorData = await response.json();
                } catch {
                    errorData = {
                        code: 'UNKNOWN_ERROR',
                        message: `Request failed with status ${response.status}`,
                    };
                }

                throw new CRMError(
                    errorData.message,
                    errorData.code,
                    response.status,
                    errorData.details,
                    errorData.requestId
                );
            }

            // Handle empty responses
            if (response.status === 204) {
                return undefined as T;
            }

            const data = await response.json();
            return this.keysToCamelCase<T>(data);
        } catch (error) {
            clearTimeout(timeoutId);

            if (error instanceof CRMError) {
                throw error;
            }

            if (error instanceof Error) {
                if (error.name === 'AbortError') {
                    throw new CRMError('Request timeout', 'TIMEOUT', 408);
                }
                throw new CRMError(error.message, 'NETWORK_ERROR', 0);
            }

            throw new CRMError('Unknown error', 'UNKNOWN_ERROR', 0);
        }
    }

    // Convenience methods

    async get<T>(path: string, params?: Record<string, string | number | boolean | undefined>): Promise<T> {
        return this.request<T>({ method: 'GET', path, params });
    }

    async post<T>(path: string, body?: unknown): Promise<T> {
        return this.request<T>({ method: 'POST', path, body });
    }

    async put<T>(path: string, body?: unknown): Promise<T> {
        return this.request<T>({ method: 'PUT', path, body });
    }

    async patch<T>(path: string, body?: unknown): Promise<T> {
        return this.request<T>({ method: 'PATCH', path, body });
    }

    async delete<T>(path: string): Promise<T> {
        return this.request<T>({ method: 'DELETE', path });
    }
}
