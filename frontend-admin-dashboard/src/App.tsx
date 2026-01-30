// ============================================
// Main App Component with Routing
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { MainLayout } from './components/Layout';
import {
  LoginPage,
  DashboardPage,
  TenantList,
  UserList,
  HealthDashboard,
  Settings,
} from './pages';

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          {/* Public Route - Login */}
          <Route path="/login" element={<LoginPage />} />

          {/* Protected Routes - Wrapped in MainLayout */}
          <Route element={<MainLayout />}>
            {/* Dashboard */}
            <Route path="/" element={<DashboardPage />} />

            {/* Tenant Management */}
            <Route path="/tenants" element={<TenantList />} />

            {/* User Management */}
            <Route path="/users" element={<UserList />} />

            {/* System Monitoring */}
            <Route path="/monitoring" element={<HealthDashboard />} />

            {/* Settings */}
            <Route path="/settings" element={<Settings />} />

            {/* Catch all - redirect to dashboard */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
