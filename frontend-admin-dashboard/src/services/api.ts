// ============================================
// API Client Service
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Token management
let accessToken: string | null = null;
let refreshToken: string | null = null;

export const setTokens = (access: string, refresh: string) => {
    accessToken = access;
    refreshToken = refresh;
    localStorage.setItem('admin_access_token', access);
    localStorage.setItem('admin_refresh_token', refresh);
};

export const getAccessToken = (): string | null => {
    if (!accessToken) {
        accessToken = localStorage.getItem('admin_access_token');
    }
    return accessToken;
};

export const getRefreshToken = (): string | null => {
    if (!refreshToken) {
        refreshToken = localStorage.getItem('admin_refresh_token');
    }
    return refreshToken;
};

export const clearTokens = () => {
    accessToken = null;
    refreshToken = null;
    localStorage.removeItem('admin_access_token');
    localStorage.removeItem('admin_refresh_token');
};

// API Error class
export class ApiError extends Error {
    status: number;
    code: string;
    details?: Array<{ field: string; message: string }>;

    constructor(
        status: number,
        code: string,
        message: string,
        details?: Array<{ field: string; message: string }>
    ) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
        this.code = code;
        this.details = details;
    }
}

// Generic fetch wrapper with authentication
async function apiFetch<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const token = getAccessToken();

    const headers: HeadersInit = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    if (token) {
        (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
    }

    try {
        const response = await fetch(url, {
            ...options,
            headers,
        });

        // Handle 401 - attempt token refresh
        if (response.status === 401 && refreshToken) {
            const refreshed = await refreshAccessToken();
            if (refreshed) {
                // Retry the request with new token
                (headers as Record<string, string>)['Authorization'] = `Bearer ${getAccessToken()}`;
                const retryResponse = await fetch(url, { ...options, headers });
                if (retryResponse.ok) {
                    return retryResponse.json();
                }
            }
            // Refresh failed, clear tokens and redirect
            clearTokens();
            window.location.href = '/login';
            throw new ApiError(401, 'ERR_UNAUTHORIZED', 'Session expired');
        }

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new ApiError(
                response.status,
                errorData.error?.code || 'ERR_UNKNOWN',
                errorData.error?.message || 'An error occurred',
                errorData.error?.details
            );
        }

        return response.json();
    } catch (error) {
        if (error instanceof ApiError) {
            throw error;
        }
        throw new ApiError(0, 'ERR_NETWORK', 'Network error occurred');
    }
}

async function refreshAccessToken(): Promise<boolean> {
    const refresh = getRefreshToken();
    if (!refresh) return false;

    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refresh }),
        });

        if (response.ok) {
            const data = await response.json();
            setTokens(data.data.access_token, data.data.refresh_token);
            return true;
        }
    } catch {
        // Refresh failed
    }
    return false;
}

// ============================================
// Authentication API
// ============================================

export const authApi = {
    login: (email: string, password: string) =>
        apiFetch<{ success: boolean; data: { access_token: string; refresh_token: string; expires_in: number } }>(
            '/api/v1/auth/login',
            {
                method: 'POST',
                body: JSON.stringify({ email, password }),
            }
        ),

    logout: () =>
        apiFetch<{ success: boolean }>('/api/v1/auth/logout', {
            method: 'POST',
        }),

    getProfile: () =>
        apiFetch<{ success: boolean; data: { user: import('../types').User } }>('/api/v1/auth/me'),

    changePassword: (currentPassword: string, newPassword: string) =>
        apiFetch<{ success: boolean }>('/api/v1/auth/password', {
            method: 'PUT',
            body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
        }),
};

// ============================================
// Tenant API
// ============================================

export const tenantApi = {
    list: (params?: {
        page?: number;
        pageSize?: number;
        status?: string;
        plan?: string;
        search?: string;
        sortBy?: string;
        sortDirection?: string;
    }) => {
        const searchParams = new URLSearchParams();
        if (params?.page) searchParams.set('page', params.page.toString());
        if (params?.pageSize) searchParams.set('per_page', params.pageSize.toString());
        if (params?.status) searchParams.set('status', params.status);
        if (params?.plan) searchParams.set('plan', params.plan);
        if (params?.search) searchParams.set('search', params.search);
        if (params?.sortBy) searchParams.set('sort', params.sortBy);
        if (params?.sortDirection) searchParams.set('order', params.sortDirection);

        return apiFetch<{
            success: boolean;
            data: import('../types').Tenant[];
            meta: import('../types').Pagination;
        }>(`/api/v1/tenants?${searchParams.toString()}`);
    },

    get: (id: string) =>
        apiFetch<{ success: boolean; data: { tenant: import('../types').Tenant } }>(
            `/api/v1/tenants/${id}`
        ),

    create: (data: import('../types').CreateTenantRequest) =>
        apiFetch<{ success: boolean; data: { tenant: import('../types').Tenant } }>(
            '/api/v1/tenants',
            {
                method: 'POST',
                body: JSON.stringify(data),
            }
        ),

    update: (id: string, data: import('../types').UpdateTenantRequest) =>
        apiFetch<{ success: boolean; data: { tenant: import('../types').Tenant } }>(
            `/api/v1/tenants/${id}`,
            {
                method: 'PUT',
                body: JSON.stringify(data),
            }
        ),

    delete: (id: string) =>
        apiFetch<{ success: boolean }>(`/api/v1/tenants/${id}`, {
            method: 'DELETE',
        }),

    updateStatus: (id: string, status: string) =>
        apiFetch<{ success: boolean; data: { tenant: import('../types').Tenant } }>(
            `/api/v1/tenants/${id}/status`,
            {
                method: 'PUT',
                body: JSON.stringify({ status }),
            }
        ),

    updatePlan: (id: string, plan: string) =>
        apiFetch<{ success: boolean; data: { tenant: import('../types').Tenant } }>(
            `/api/v1/tenants/${id}/plan`,
            {
                method: 'PUT',
                body: JSON.stringify({ plan }),
            }
        ),

    getStats: (id: string) =>
        apiFetch<{ success: boolean; data: import('../types').TenantStats }>(
            `/api/v1/tenants/${id}/stats`
        ),

    checkSlug: (slug: string) =>
        apiFetch<{ success: boolean; data: { available: boolean; suggestions?: string[] } }>(
            `/api/v1/tenants/check-slug?slug=${encodeURIComponent(slug)}`
        ),
};

// ============================================
// User API
// ============================================

export const userApi = {
    list: (params?: import('../types').UserListFilters) => {
        const searchParams = new URLSearchParams();
        if (params?.page) searchParams.set('page', params.page.toString());
        if (params?.pageSize) searchParams.set('per_page', params.pageSize.toString());
        if (params?.search) searchParams.set('search', params.search);
        if (params?.status) searchParams.set('status', params.status);
        if (params?.tenantId) searchParams.set('tenant_id', params.tenantId);
        if (params?.sortBy) searchParams.set('sort', params.sortBy);
        if (params?.sortDirection) searchParams.set('order', params.sortDirection);

        return apiFetch<{
            success: boolean;
            data: import('../types').User[];
            meta: import('../types').Pagination;
        }>(`/api/v1/users?${searchParams.toString()}`);
    },

    get: (id: string) =>
        apiFetch<{ success: boolean; data: { user: import('../types').User } }>(
            `/api/v1/users/${id}`
        ),

    update: (id: string, data: Partial<import('../types').User>) =>
        apiFetch<{ success: boolean; data: { user: import('../types').User } }>(
            `/api/v1/users/${id}`,
            {
                method: 'PUT',
                body: JSON.stringify(data),
            }
        ),

    delete: (id: string) =>
        apiFetch<{ success: boolean }>(`/api/v1/users/${id}`, {
            method: 'DELETE',
        }),

    updateStatus: (id: string, status: string) =>
        apiFetch<{ success: boolean; data: { user: import('../types').User } }>(
            `/api/v1/users/${id}/status`,
            {
                method: 'PUT',
                body: JSON.stringify({ status }),
            }
        ),

    assignRole: (id: string, roleId: string) =>
        apiFetch<{ success: boolean }>(`/api/v1/users/${id}/roles`, {
            method: 'POST',
            body: JSON.stringify({ role_id: roleId }),
        }),

    removeRole: (id: string, roleId: string) =>
        apiFetch<{ success: boolean }>(`/api/v1/users/${id}/roles`, {
            method: 'DELETE',
            body: JSON.stringify({ role_id: roleId }),
        }),

    getRoles: (id: string) =>
        apiFetch<{ success: boolean; data: { roles: import('../types').Role[] } }>(
            `/api/v1/users/${id}/roles`
        ),

    getPermissions: (id: string) =>
        apiFetch<{ success: boolean; data: { permissions: string[] } }>(
            `/api/v1/users/${id}/permissions`
        ),

    resetPassword: (id: string, options?: { newPassword?: string; sendEmail?: boolean }) =>
        apiFetch<{ success: boolean; data: { temporaryPassword?: string } }>(
            `/api/v1/users/${id}/reset-password`,
            {
                method: 'POST',
                body: JSON.stringify(options || {}),
            }
        ),
};

// ============================================
// Role API
// ============================================

export const roleApi = {
    list: () =>
        apiFetch<{ success: boolean; data: import('../types').Role[] }>('/api/v1/roles'),

    get: (id: string) =>
        apiFetch<{ success: boolean; data: { role: import('../types').Role } }>(
            `/api/v1/roles/${id}`
        ),

    getSystemRoles: () =>
        apiFetch<{ success: boolean; data: import('../types').Role[] }>('/api/v1/roles/system'),
};

// ============================================
// Health API
// ============================================

const SERVICES = [
    { name: 'API Gateway', port: 8080, url: 'http://localhost:8080' },
    { name: 'IAM Service', port: 8081, url: 'http://localhost:8081' },
    { name: 'Customer Service', port: 8082, url: 'http://localhost:8082' },
    { name: 'Sales Service', port: 8083, url: 'http://localhost:8083' },
    { name: 'Notification Service', port: 8084, url: 'http://localhost:8084' },
];

export const healthApi = {
    checkService: async (serviceUrl: string): Promise<import('../types').HealthCheckResponse | null> => {
        try {
            const response = await fetch(`${serviceUrl}/health`, {
                method: 'GET',
                signal: AbortSignal.timeout(5000),
            });
            if (response.ok) {
                return response.json();
            }
            return null;
        } catch {
            return null;
        }
    },

    checkAllServices: async (): Promise<import('../types').ServiceHealth[]> => {
        const results: import('../types').ServiceHealth[] = [];

        for (const service of SERVICES) {
            const startTime = Date.now();
            try {
                const health = await healthApi.checkService(service.url);
                const responseTime = Date.now() - startTime;

                results.push({
                    name: service.name,
                    status: health?.status === 'healthy' ? 'healthy' : 'unhealthy',
                    url: service.url,
                    port: service.port,
                    lastChecked: new Date().toISOString(),
                    responseTime,
                    details: health || undefined,
                });
            } catch {
                results.push({
                    name: service.name,
                    status: 'unhealthy',
                    url: service.url,
                    port: service.port,
                    lastChecked: new Date().toISOString(),
                });
            }
        }

        return results;
    },

    getDatabaseStatus: async (): Promise<import('../types').DatabaseStatus[]> => {
        // In a real implementation, these would come from dedicated health endpoints
        // For now, we'll derive from service health
        return [
            {
                name: 'PostgreSQL',
                type: 'postgresql',
                status: 'healthy',
                connectionPool: { active: 5, idle: 15, total: 20 },
                latency: 2,
            },
            {
                name: 'MongoDB',
                type: 'mongodb',
                status: 'healthy',
                connectionPool: { active: 3, idle: 7, total: 10 },
                latency: 3,
            },
            {
                name: 'Redis',
                type: 'redis',
                status: 'healthy',
                connectionPool: { active: 2, idle: 8, total: 10 },
                latency: 1,
            },
        ];
    },

    getQueueStatus: async (): Promise<import('../types').QueueStatus[]> => {
        // In a real implementation, this would query RabbitMQ management API
        return [
            {
                name: 'notifications',
                status: 'healthy',
                messageCount: 0,
                consumerCount: 2,
                publishRate: 10,
                consumeRate: 10,
            },
            {
                name: 'events',
                status: 'healthy',
                messageCount: 5,
                consumerCount: 3,
                publishRate: 25,
                consumeRate: 24,
            },
            {
                name: 'dead-letter',
                status: 'healthy',
                messageCount: 0,
                consumerCount: 1,
                publishRate: 0,
                consumeRate: 0,
            },
        ];
    },
};

export default {
    auth: authApi,
    tenants: tenantApi,
    users: userApi,
    roles: roleApi,
    health: healthApi,
};
