// ============================================
// CRM Web Client - Main Application
// Production-Ready React Application
// ============================================

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ToastProvider } from './components/ui/Toast';
import { MainLayout } from './components/layout/MainLayout';
import {
  LoginPage,
  RegisterPage,
  ForgotPasswordPage,
  ResetPasswordPage,
  EmailVerificationPage,
  DashboardPage,
  LeadListPage,
  LeadDetailPage,
  LeadFormPage,
  CustomerListPage,
  CustomerDetailPage,
  OpportunityListPage,
  OpportunityDetailPage,
  PipelinePage,
  SettingsLayout,
  ProfileSettingsPage,
  SecuritySettingsPage,
  NotificationSettingsPage,
  TeamManagementPage,
} from './pages';

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
          <Routes>
            {/* Auth Routes - No Layout */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/forgot-password" element={<ForgotPasswordPage />} />
            <Route path="/reset-password" element={<ResetPasswordPage />} />
            <Route path="/verify-email" element={<EmailVerificationPage />} />

            {/* Protected Routes - With MainLayout */}
            <Route element={<MainLayout />}>
              {/* Dashboard */}
              <Route path="/" element={<DashboardPage />} />
              <Route path="/dashboard" element={<Navigate to="/" replace />} />

              {/* Leads */}
              <Route path="/leads" element={<LeadListPage />} />
              <Route path="/leads/new" element={<LeadFormPage />} />
              <Route path="/leads/:id" element={<LeadDetailPage />} />
              <Route path="/leads/:id/edit" element={<LeadFormPage />} />

              {/* Customers */}
              <Route path="/customers" element={<CustomerListPage />} />
              <Route path="/customers/new" element={<CustomerDetailPage />} />
              <Route path="/customers/:id" element={<CustomerDetailPage />} />
              <Route path="/customers/:id/edit" element={<CustomerDetailPage />} />

              {/* Opportunities */}
              <Route path="/opportunities" element={<OpportunityListPage />} />
              <Route path="/opportunities/:id" element={<OpportunityDetailPage />} />

              {/* Pipeline */}
              <Route path="/pipeline" element={<PipelinePage />} />

              {/* Settings - Nested Routes */}
              <Route path="/settings" element={<SettingsLayout />}>
                <Route index element={<Navigate to="/settings/profile" replace />} />
                <Route path="profile" element={<ProfileSettingsPage />} />
                <Route path="security" element={<SecuritySettingsPage />} />
                <Route path="notifications" element={<NotificationSettingsPage />} />
                <Route path="team" element={<TeamManagementPage />} />
              </Route>

              {/* Catch all - Redirect to dashboard */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
