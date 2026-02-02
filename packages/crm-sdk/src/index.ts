// ============================================
// CRM SDK - Main Entry Point
// @kilangbatik/crm-sdk
// ============================================

// Main client
export { CRMClient, createClient } from './client';

// Types
export type {
    // Configuration
    CRMClientConfig,
    TokenPair,

    // Common
    PaginationMeta,
    PaginatedResponse,
    PaginationParams,

    // Errors
    CRMErrorResponse,

    // User & Auth
    User,
    LoginRequest,
    LoginResponse,
    RegisterRequest,
    ForgotPasswordRequest,
    ResetPasswordRequest,
    ChangePasswordRequest,
    ProfileUpdateRequest,

    // Leads
    Lead,
    LeadStatus,
    LeadFilters,
    CreateLeadRequest,
    UpdateLeadRequest,
    ConvertLeadRequest,
    ConvertLeadResponse,
    QualifyLeadRequest,
    DisqualifyLeadRequest,

    // Customers
    Customer,
    CustomerStatus,
    CustomerFilters,
    CreateCustomerRequest,
    UpdateCustomerRequest,
    CustomerContact,
    CreateContactRequest,
    CustomerNote,
    CreateNoteRequest,
    CustomerActivity,

    // Opportunities
    Opportunity,
    OpportunityStatus,
    OpportunityFilters,
    CreateOpportunityRequest,
    UpdateOpportunityRequest,
    MoveStageRequest,
    WinOpportunityRequest,
    LoseOpportunityRequest,

    // Pipelines
    Pipeline,
    PipelineStage,
    Deal,
    CreatePipelineRequest,
    CreateStageRequest,

    // Dashboard
    DashboardData,
    DashboardStats,
    PipelineOverview,
    PipelineStageStats,
    RecentActivity,
} from './types';

// Error class
export { CRMError } from './types';

// Services (for advanced usage)
export {
    AuthService,
    LeadsService,
    CustomersService,
    OpportunitiesService,
    PipelinesService,
    DashboardService,
} from './services';
