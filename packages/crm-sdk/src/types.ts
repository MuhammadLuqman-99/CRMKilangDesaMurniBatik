// ============================================
// CRM SDK Type Definitions
// All TypeScript types for the API
// ============================================

// -------------------- Configuration --------------------

export interface CRMClientConfig {
    /** Base URL of the CRM API (e.g., https://api.example.com) */
    baseUrl: string;
    /** API version (default: 'v1') */
    apiVersion?: string;
    /** Request timeout in milliseconds (default: 30000) */
    timeout?: number;
    /** Custom headers to include with every request */
    headers?: Record<string, string>;
    /** Callback when token refresh is needed */
    onTokenRefresh?: (newTokens: TokenPair) => void;
    /** Callback when authentication fails */
    onAuthError?: (error: CRMError) => void;
}

export interface TokenPair {
    accessToken: string;
    refreshToken: string;
    expiresIn?: number;
}

// -------------------- Common Types --------------------

export interface PaginationMeta {
    total: number;
    page: number;
    perPage: number;
    totalPages: number;
}

export interface PaginatedResponse<T> {
    data: T[];
    meta: PaginationMeta;
}

export interface PaginationParams {
    page?: number;
    perPage?: number;
    sort?: string;
    order?: 'asc' | 'desc';
}

// -------------------- Error Types --------------------

export interface CRMErrorResponse {
    code: string;
    message: string;
    details?: Record<string, unknown>;
    requestId?: string;
}

export class CRMError extends Error {
    code: string;
    status: number;
    details?: Record<string, unknown>;
    requestId?: string;

    constructor(message: string, code: string, status: number, details?: Record<string, unknown>, requestId?: string) {
        super(message);
        this.name = 'CRMError';
        this.code = code;
        this.status = status;
        this.details = details;
        this.requestId = requestId;
    }
}

// -------------------- User & Auth Types --------------------

export interface User {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
    role: string;
    avatarUrl?: string;
    phone?: string;
    title?: string;
    department?: string;
    timezone?: string;
    createdAt: string;
    updatedAt?: string;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    user: User;
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
}

export interface RegisterRequest {
    firstName: string;
    lastName: string;
    email: string;
    password: string;
}

export interface ForgotPasswordRequest {
    email: string;
}

export interface ResetPasswordRequest {
    token: string;
    password: string;
}

export interface ChangePasswordRequest {
    currentPassword: string;
    newPassword: string;
}

export interface ProfileUpdateRequest {
    firstName?: string;
    lastName?: string;
    phone?: string;
    title?: string;
    department?: string;
    timezone?: string;
    avatarUrl?: string;
}

// -------------------- Lead Types --------------------

export interface Lead {
    id: string;
    firstName: string;
    lastName: string;
    email: string;
    phone?: string;
    company?: string;
    title?: string;
    status: LeadStatus;
    source?: string;
    score?: number;
    scoreLabel?: 'hot' | 'warm' | 'cold';
    industry?: string;
    website?: string;
    address?: string;
    notes?: string;
    ownerId?: string;
    ownerName?: string;
    createdAt: string;
    updatedAt?: string;
}

export type LeadStatus = 'new' | 'contacted' | 'qualified' | 'unqualified' | 'converted';

export interface LeadFilters extends PaginationParams {
    status?: LeadStatus;
    source?: string;
    search?: string;
    ownerId?: string;
    scoreLabel?: 'hot' | 'warm' | 'cold';
    createdFrom?: string;
    createdTo?: string;
}

export interface CreateLeadRequest {
    firstName: string;
    lastName: string;
    email: string;
    phone?: string;
    company?: string;
    title?: string;
    source?: string;
    industry?: string;
    website?: string;
    address?: string;
    notes?: string;
    ownerId?: string;
}

export interface UpdateLeadRequest extends Partial<CreateLeadRequest> {}

export interface ConvertLeadRequest {
    createOpportunity?: boolean;
    createCustomer?: boolean;
    opportunityName?: string;
    opportunityValue?: number;
}

export interface ConvertLeadResponse {
    leadId: string;
    opportunityId?: string;
    customerId?: string;
}

export interface QualifyLeadRequest {
    score?: number;
    notes?: string;
}

export interface DisqualifyLeadRequest {
    reason: string;
    notes?: string;
}

// -------------------- Customer Types --------------------

export interface Customer {
    id: string;
    name: string;
    email: string;
    phone?: string;
    status: CustomerStatus;
    segment?: string;
    industry?: string;
    website?: string;
    address?: string;
    description?: string;
    totalValue?: number;
    dealsCount?: number;
    ownerId?: string;
    ownerName?: string;
    createdAt: string;
    updatedAt?: string;
}

export type CustomerStatus = 'active' | 'inactive' | 'prospect' | 'churned';

export interface CustomerFilters extends PaginationParams {
    status?: CustomerStatus;
    segment?: string;
    search?: string;
    ownerId?: string;
    industry?: string;
    createdFrom?: string;
    createdTo?: string;
}

export interface CreateCustomerRequest {
    name: string;
    email: string;
    phone?: string;
    status?: CustomerStatus;
    segment?: string;
    industry?: string;
    website?: string;
    address?: string;
    description?: string;
}

export interface UpdateCustomerRequest extends Partial<CreateCustomerRequest> {}

export interface CustomerContact {
    id: string;
    customerId?: string;
    name: string;
    email: string;
    phone?: string;
    title?: string;
    department?: string;
    isPrimary: boolean;
    createdAt?: string;
}

export interface CreateContactRequest {
    name: string;
    email: string;
    phone?: string;
    title?: string;
    department?: string;
    isPrimary?: boolean;
}

export interface CustomerNote {
    id: string;
    customerId?: string;
    content: string;
    createdBy: string;
    createdAt: string;
    isPinned?: boolean;
}

export interface CreateNoteRequest {
    content: string;
}

export interface CustomerActivity {
    id: string;
    customerId?: string;
    type: string;
    title: string;
    description?: string;
    userName: string;
    createdAt: string;
    status?: string;
}

// -------------------- Opportunity Types --------------------

export interface Opportunity {
    id: string;
    name: string;
    customerId?: string;
    customerName?: string;
    value: number;
    stageId: string;
    stageName?: string;
    probability?: number;
    expectedCloseDate?: string;
    actualCloseDate?: string;
    status: OpportunityStatus;
    notes?: string;
    ownerId?: string;
    ownerName?: string;
    createdAt: string;
    updatedAt?: string;
}

export type OpportunityStatus = 'open' | 'won' | 'lost';

export interface OpportunityFilters extends PaginationParams {
    status?: OpportunityStatus;
    stageId?: string;
    customerId?: string;
    search?: string;
    pipelineId?: string;
    ownerId?: string;
    valueMin?: number;
    valueMax?: number;
    expectedCloseFrom?: string;
    expectedCloseTo?: string;
}

export interface CreateOpportunityRequest {
    name: string;
    customerId?: string;
    value: number;
    stageId: string;
    probability?: number;
    expectedCloseDate?: string;
    notes?: string;
    ownerId?: string;
}

export interface UpdateOpportunityRequest extends Partial<CreateOpportunityRequest> {}

export interface MoveStageRequest {
    stageId: string;
    position?: number;
}

export interface WinOpportunityRequest {
    actualCloseDate?: string;
    notes?: string;
}

export interface LoseOpportunityRequest {
    reason: string;
    notes?: string;
}

// -------------------- Pipeline Types --------------------

export interface Pipeline {
    id: string;
    name: string;
    description?: string;
    isDefault?: boolean;
    stages: PipelineStage[];
    createdAt?: string;
}

export interface PipelineStage {
    id: string;
    pipelineId?: string;
    name: string;
    position: number;
    color?: string;
    probability?: number;
    createdAt?: string;
}

export interface Deal {
    id: string;
    name: string;
    customerId?: string;
    customerName?: string;
    value: number;
    stageId: string;
    priority?: 'low' | 'medium' | 'high' | 'urgent';
    expectedCloseDate?: string;
    ownerId?: string;
    ownerName?: string;
    probability?: number;
    notes?: string;
    createdAt?: string;
    updatedAt?: string;
}

export interface CreatePipelineRequest {
    name: string;
    description?: string;
    isDefault?: boolean;
}

export interface CreateStageRequest {
    name: string;
    position?: number;
    color?: string;
    probability?: number;
}

// -------------------- Dashboard Types --------------------

export interface DashboardStats {
    totalLeads: number;
    leadsChange: number;
    totalOpportunities: number;
    opportunitiesChange: number;
    pipelineValue: number;
    pipelineChange: number;
    wonDeals: number;
    dealsChange: number;
}

export interface PipelineOverview {
    stages: PipelineStageStats[];
    totalValue: number;
    totalDeals: number;
}

export interface PipelineStageStats {
    id: string;
    name: string;
    value: number;
    count: number;
    color: string;
}

export interface RecentActivity {
    id: string;
    type: string;
    title: string;
    description: string;
    user?: string;
    timestamp: string;
}

export interface DashboardData {
    stats: DashboardStats;
    pipelineOverview: PipelineOverview;
    recentActivities: RecentActivity[];
}
