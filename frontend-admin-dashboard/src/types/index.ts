// ============================================
// TypeScript Type Definitions
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

// ============================================
// Authentication Types
// ============================================

export interface User {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
    phone?: string;
    status: UserStatus;
    tenantId: string;
    tenantName?: string;
    roles: Role[];
    permissions: string[];
    createdAt: string;
    updatedAt: string;
    lastLoginAt?: string;
}

export type UserStatus = 'active' | 'inactive' | 'suspended' | 'pending_verification';

export interface Role {
    id: string;
    name: string;
    description?: string;
    permissions: Permission[];
    isSystem: boolean;
}

export interface Permission {
    id: string;
    name: string;
    resource: string;
    action: string;
    description?: string;
}

export interface AuthTokens {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
    tokenType: string;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    success: boolean;
    data: {
        access_token: string;
        refresh_token: string;
        expires_in: number;
        token_type: string;
        user?: User;
    };
}

// ============================================
// Tenant Types
// ============================================

export interface Tenant {
    id: string;
    name: string;
    slug: string;
    status: TenantStatus;
    plan: TenantPlan;
    settings?: TenantSettings;
    limits?: TenantLimits;
    usage?: TenantUsage;
    trialInfo?: TrialInfo;
    createdAt: string;
    updatedAt: string;
}

export type TenantStatus = 'active' | 'inactive' | 'suspended' | 'pending' | 'trial';

export type TenantPlan = 'free' | 'starter' | 'pro' | 'enterprise';

export interface TenantSettings {
    timezone: string;
    dateFormat: string;
    currency: string;
    language: string;
    notificationsEmail: boolean;
}

export interface TenantLimits {
    maxUsers: number;
    maxContacts: number;
}

export interface TenantUsage {
    userCount: number;
    contactCount: number;
}

export interface TrialInfo {
    isTrialing: boolean;
    trialStarted?: string;
    trialEnds?: string;
    daysLeft?: number;
}

export interface TenantStats {
    totalTenants: number;
    activeTenants: number;
    trialTenants: number;
    pendingTenants: number;
    planBreakdown: Record<string, number>;
}

export interface CreateTenantRequest {
    name: string;
    slug: string;
    plan?: TenantPlan;
}

export interface UpdateTenantRequest {
    name?: string;
    settings?: Partial<TenantSettings>;
}

export interface UpdateTenantStatusRequest {
    status: TenantStatus;
}

export interface UpdateTenantPlanRequest {
    plan: TenantPlan;
}

// ============================================
// User Management Types
// ============================================

export interface UserListFilters {
    page?: number;
    pageSize?: number;
    search?: string;
    status?: UserStatus;
    tenantId?: string;
    sortBy?: string;
    sortDirection?: 'asc' | 'desc';
}

export interface UpdateUserStatusRequest {
    status: UserStatus;
}

export interface AssignRoleRequest {
    roleId: string;
}

export interface ResetPasswordRequest {
    newPassword?: string;
    sendEmail?: boolean;
}

// ============================================
// API Response Types
// ============================================

export interface ApiResponse<T> {
    success: boolean;
    data: T;
    error?: ApiError;
}

export interface ApiError {
    code: string;
    message: string;
    details?: Array<{
        field: string;
        message: string;
    }>;
}

export interface PaginatedResponse<T> {
    success: boolean;
    data: T[];
    meta: {
        page: number;
        perPage: number;
        total: number;
        totalPages: number;
    };
}

export interface Pagination {
    page: number;
    perPage: number;
    total: number;
    totalPages: number;
}

// ============================================
// Health & Monitoring Types
// ============================================

export interface ServiceHealth {
    name: string;
    status: HealthStatus;
    url: string;
    port: number;
    lastChecked: string;
    responseTime?: number;
    details?: HealthCheckResponse;
}

export type HealthStatus = 'healthy' | 'unhealthy' | 'degraded' | 'unknown';

export interface HealthCheckResponse {
    status: string;
    service?: string;
    version?: string;
    uptime?: number;
    checks?: ComponentCheck[];
}

export interface ComponentCheck {
    name: string;
    status: string;
    duration?: string;
}

export interface DatabaseStatus {
    name: string;
    type: 'postgresql' | 'mongodb' | 'redis';
    status: HealthStatus;
    connectionPool?: {
        active: number;
        idle: number;
        total: number;
    };
    latency?: number;
}

export interface QueueStatus {
    name: string;
    status: HealthStatus;
    messageCount: number;
    consumerCount: number;
    publishRate?: number;
    consumeRate?: number;
}

export interface SystemMetrics {
    services: ServiceHealth[];
    databases: DatabaseStatus[];
    queues: QueueStatus[];
    lastUpdated: string;
}

// ============================================
// Dashboard Types
// ============================================

export interface DashboardStats {
    totalTenants: number;
    activeTenants: number;
    totalUsers: number;
    activeUsers: number;
    servicesHealthy: number;
    servicesTotal: number;
}

// ============================================
// Common Types
// ============================================

export interface SelectOption {
    value: string;
    label: string;
}

export interface TableColumn<T> {
    key: keyof T | string;
    header: string;
    width?: string;
    render?: (item: T) => React.ReactNode;
}

export interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
    size?: 'sm' | 'md' | 'lg';
}

export interface ConfirmDialogProps {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => void;
    title: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    variant?: 'danger' | 'warning' | 'primary';
}
