// ============================================
// Type Definitions
// Production-Ready TypeScript Types
// ============================================

// -------------------- Common Types --------------------

export interface PaginationMeta {
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
}

export interface ApiResponse<T> {
    data: T;
    meta?: PaginationMeta;
    message?: string;
}

// -------------------- User Types --------------------

export interface User {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    role: string;
    avatar_url?: string;
    phone?: string;
    title?: string;
    department?: string;
    timezone?: string;
    created_at: string;
    updated_at?: string;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface LoginResponse {
    user: User;
    access_token: string;
    refresh_token: string;
    expires_in: number;
}

export interface RegisterRequest {
    first_name: string;
    last_name: string;
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
    current_password: string;
    new_password: string;
}

// -------------------- Lead Types --------------------

export interface Lead {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    phone?: string;
    company?: string;
    title?: string;
    status: 'new' | 'contacted' | 'qualified' | 'unqualified' | 'converted';
    source?: string;
    score?: number;
    score_label?: 'hot' | 'warm' | 'cold';
    industry?: string;
    website?: string;
    address?: string;
    notes?: string;
    owner_id?: string;
    owner_name?: string;
    created_at: string;
    updated_at?: string;
}

export interface LeadFilters {
    page?: number;
    per_page?: number;
    status?: string;
    source?: string;
    search?: string;
    owner_id?: string;
    score_label?: string;
    sort?: string;
    order?: 'asc' | 'desc';
    sort_by?: string;
    sort_order?: 'asc' | 'desc';
    created_from?: string;
    created_to?: string;
}

export interface CreateLeadRequest {
    first_name: string;
    last_name: string;
    email: string;
    phone?: string;
    company?: string;
    title?: string;
    source?: string;
    industry?: string;
    website?: string;
    address?: string;
    notes?: string;
    owner_id?: string;
}

export interface UpdateLeadRequest extends Partial<CreateLeadRequest> { }

export interface ConvertLeadRequest {
    create_opportunity?: boolean;
    create_customer?: boolean;
    opportunity_name?: string;
    opportunity_value?: number;
}

export interface ConvertLeadResponse {
    lead_id: string;
    opportunity_id?: string;
    customer_id?: string;
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
    status: 'active' | 'inactive' | 'prospect' | 'churned';
    segment?: string;
    industry?: string;
    website?: string;
    address?: string;
    description?: string;
    total_value?: number;
    deals_count?: number;
    owner_id?: string;
    owner_name?: string;
    created_at: string;
    updated_at?: string;
}

export interface CustomerFilters {
    page?: number;
    per_page?: number;
    status?: string;
    segment?: string;
    search?: string;
    owner_id?: string;
    type?: string;
    industry?: string;
    sort?: string;
    order?: 'asc' | 'desc';
    created_from?: string;
    created_to?: string;
}

export interface CreateCustomerRequest {
    name: string;
    email: string;
    phone?: string;
    status?: string;
    segment?: string;
    industry?: string;
    website?: string;
    address?: string;
    description?: string;
}

export interface UpdateCustomerRequest extends Partial<CreateCustomerRequest> { }

export interface CustomerContact {
    id: string;
    customer_id?: string;
    name: string;
    email: string;
    phone?: string;
    title?: string;
    department?: string;
    is_primary: boolean;
    created_at?: string;
}

// Alias for backwards compatibility
export type Contact = CustomerContact;

export interface CreateContactRequest {
    name: string;
    email: string;
    phone?: string;
    title?: string;
    department?: string;
    is_primary?: boolean;
}

export interface UpdateContactRequest extends Partial<CreateContactRequest> { }

export interface CustomerNote {
    id: string;
    customer_id?: string;
    content: string;
    created_by: string;
    created_at: string;
    is_pinned?: boolean;
}

// Alias for backwards compatibility
export type Note = CustomerNote;

export interface CreateNoteRequest {
    content: string;
}

export interface CustomerActivity {
    id: string;
    customer_id?: string;
    type: string;
    title: string;
    description?: string;
    user_name: string;
    created_at: string;
    status?: string;
}

// Alias for backwards compatibility
export type Activity = CustomerActivity;

export interface ActivityFilters {
    page?: number;
    per_page?: number;
    type?: string;
    status?: string;
}

// -------------------- Opportunity Types --------------------

export interface Opportunity {
    id: string;
    name: string;
    customer_id?: string;
    customer_name?: string;
    value: number;
    stage_id: string;
    stage_name?: string;
    probability?: number;
    expected_close_date?: string;
    actual_close_date?: string;
    status: 'open' | 'won' | 'lost';
    notes?: string;
    owner_id?: string;
    owner_name?: string;
    created_at: string;
    updated_at?: string;
}

export interface OpportunityFilters {
    page?: number;
    per_page?: number;
    status?: string;
    stage_id?: string;
    customer_id?: string;
    search?: string;
    pipeline_id?: string;
    owner_id?: string;
    sort?: string;
    order?: 'asc' | 'desc';
    value_min?: number;
    value_max?: number;
    expected_close_from?: string;
    expected_close_to?: string;
}

export interface CreateOpportunityRequest {
    name: string;
    customer_id?: string;
    value: number;
    stage_id: string;
    probability?: number;
    expected_close_date?: string;
    notes?: string;
    owner_id?: string;
}

export interface UpdateOpportunityRequest extends Partial<CreateOpportunityRequest> { }

export interface MoveStageRequest {
    stage_id: string;
    position?: number;
}

export interface WinOpportunityRequest {
    actual_close_date?: string;
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
    is_default?: boolean;
    stages: PipelineStage[];
    created_at?: string;
}

export interface PipelineStage {
    id: string;
    pipeline_id?: string;
    name: string;
    position: number;
    color?: string;
    probability?: number;
    created_at?: string;
}

export interface Deal {
    id: string;
    name: string;
    customer_id?: string;
    customer_name?: string;
    value: number;
    stage_id: string;
    priority?: 'low' | 'medium' | 'high' | 'urgent';
    expected_close_date?: string;
    owner_id?: string;
    owner_name?: string;
    probability?: number;
    notes?: string;
    created_at?: string;
    updated_at?: string;
}

export interface MoveDealRequest {
    stage_id: string;
    position?: number;
}

export interface CreatePipelineRequest {
    name: string;
    description?: string;
    is_default?: boolean;
}

export interface UpdatePipelineRequest extends Partial<CreatePipelineRequest> { }

export interface CreatePipelineStageRequest {
    name: string;
    position?: number;
    color?: string;
    probability?: number;
}

export interface PipelineAnalytics {
    total_deals: number;
    total_value: number;
    avg_deal_value: number;
    avg_time_in_stage: Record<string, number>;
    conversion_rate: number;
    deals_by_stage: { stage_id: string; stage_name: string; count: number; value: number }[];
}

// -------------------- Dashboard Types --------------------

export interface DashboardStats {
    total_leads: number;
    leads_change: number;
    total_opportunities: number;
    opportunities_change: number;
    pipeline_value: number;
    pipeline_change: number;
    won_deals: number;
    deals_change: number;
}

export interface PipelineOverview {
    stages: PipelineStageStats[];
    total_value: number;
    total_deals: number;
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

export interface UpcomingTask {
    id: string;
    title: string;
    description?: string;
    due_date: string;
    priority: 'low' | 'medium' | 'high';
    related_to?: {
        type: string;
        id: string;
        name: string;
    };
    assigned_to: string;
}

export interface DashboardData {
    stats: DashboardStats;
    pipeline_overview: PipelineOverview;
    recent_activities: RecentActivity[];
    upcoming_tasks?: UpcomingTask[];
}

export interface DealsClosingSoon {
    deals: Deal[];
    total: number;
}

// -------------------- Settings Types --------------------

export interface UserSettings {
    notifications_enabled: boolean;
    email_notifications: boolean;
    dark_mode: boolean;
    language: string;
    timezone: string;
}

export interface ProfileUpdateRequest {
    first_name?: string;
    last_name?: string;
    phone?: string;
    title?: string;
    department?: string;
    timezone?: string;
    avatar_url?: string;
}

// -------------------- Export/Import Types --------------------

export interface ExportOptions {
    format: 'csv' | 'xlsx';
    fields?: string[];
    filters?: Record<string, string>;
}

export interface ImportResult {
    total: number;
    imported: number;
    failed: number;
    errors: ImportError[];
}

export interface ImportError {
    row: number;
    field: string;
    message: string;
}
