// ============================================
// Main Layout Component
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React, { useState } from 'react';
import { Outlet, useLocation, Navigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import Sidebar from './Sidebar';
import Header from './Header';

// Page title mapping
const getPageTitle = (pathname: string): string => {
    const titles: Record<string, string> = {
        '/': 'Dashboard',
        '/tenants': 'Tenant Management',
        '/users': 'User Management',
        '/monitoring': 'System Monitoring',
        '/settings': 'Settings',
    };

    // Check for dynamic routes
    if (pathname.startsWith('/tenants/')) return 'Tenant Details';
    if (pathname.startsWith('/users/')) return 'User Details';

    return titles[pathname] || 'Admin Dashboard';
};

const MainLayout: React.FC = () => {
    const { isAuthenticated, isLoading } = useAuth();
    const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
    const location = useLocation();

    // Show loading while checking auth
    if (isLoading) {
        return (
            <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                minHeight: '100vh',
                background: 'var(--bg-primary)',
            }}>
                <div className="loading-spinner" style={{ width: 48, height: 48 }} />
            </div>
        );
    }

    // Redirect to login if not authenticated
    if (!isAuthenticated) {
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    const pageTitle = getPageTitle(location.pathname);

    return (
        <div className="app-layout">
            <Sidebar
                isCollapsed={isSidebarCollapsed}
                onToggle={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
            />

            <div className={`main-wrapper ${isSidebarCollapsed ? 'sidebar-collapsed' : ''}`}>
                <Header title={pageTitle} isSidebarCollapsed={isSidebarCollapsed} />

                <main className="main-content">
                    <Outlet />
                </main>
            </div>
        </div>
    );
};

export default MainLayout;
