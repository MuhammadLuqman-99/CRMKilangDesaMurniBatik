// ============================================
// Pages Index
// Export all pages
// ============================================

// Auth Pages
export {
    LoginPage,
    RegisterPage,
    ForgotPasswordPage,
    ResetPasswordPage,
    EmailVerificationPage,
} from './auth';

// Dashboard Pages
export { DashboardPage } from './dashboard';

// Lead Pages
export { LeadListPage, LeadDetailPage, LeadFormPage } from './leads';

// Customer Pages
export { CustomerListPage, CustomerDetailPage } from './customers';

// Opportunity Pages
export { OpportunityListPage, OpportunityDetailPage } from './opportunities';

// Pipeline Pages
export { PipelinePage } from './pipeline';

// Settings Pages
export {
    SettingsLayout,
    ProfileSettingsPage,
    SecuritySettingsPage,
    NotificationSettingsPage,
    TeamManagementPage,
} from './settings';
