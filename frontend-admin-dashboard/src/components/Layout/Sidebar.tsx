// ============================================
// Sidebar Component
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React from 'react';
import { NavLink } from 'react-router-dom';

interface SidebarProps {
    isCollapsed: boolean;
    onToggle: () => void;
}

// SVG Icons as components
const DashboardIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="7" height="9" rx="1" />
        <rect x="14" y="3" width="7" height="5" rx="1" />
        <rect x="14" y="12" width="7" height="9" rx="1" />
        <rect x="3" y="16" width="7" height="5" rx="1" />
    </svg>
);

const TenantsIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3 21h18" />
        <path d="M5 21V7l8-4v18" />
        <path d="M19 21V11l-6-4" />
        <path d="M9 9v.01" />
        <path d="M9 12v.01" />
        <path d="M9 15v.01" />
        <path d="M9 18v.01" />
    </svg>
);

const UsersIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
);

const MonitorIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
    </svg>
);

const SettingsIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
        <circle cx="12" cy="12" r="3" />
    </svg>
);

const CollapseIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="m11 17-5-5 5-5" />
        <path d="m18 17-5-5 5-5" />
    </svg>
);

const ExpandIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="m6 17 5-5-5-5" />
        <path d="m13 17 5-5-5-5" />
    </svg>
);

interface NavItem {
    to: string;
    icon: React.ReactNode;
    label: string;
}

const mainNavItems: NavItem[] = [
    { to: '/', icon: <DashboardIcon />, label: 'Dashboard' },
    { to: '/tenants', icon: <TenantsIcon />, label: 'Tenants' },
    { to: '/users', icon: <UsersIcon />, label: 'Users' },
    { to: '/monitoring', icon: <MonitorIcon />, label: 'Monitoring' },
];

const systemNavItems: NavItem[] = [
    { to: '/settings', icon: <SettingsIcon />, label: 'Settings' },
];

const Sidebar: React.FC<SidebarProps> = ({ isCollapsed, onToggle }) => {
    return (
        <aside className={`sidebar ${isCollapsed ? 'collapsed' : ''}`}>
            {/* Header */}
            <div className="sidebar-header">
                <NavLink to="/" className="sidebar-logo">
                    <div className="sidebar-logo-icon">CRM</div>
                    <span className="sidebar-logo-text">Admin Panel</span>
                </NavLink>
            </div>

            {/* Navigation */}
            <nav className="sidebar-nav">
                {/* Main Section */}
                <div className="nav-section">
                    <div className="nav-section-title">Main Menu</div>
                    {mainNavItems.map((item) => (
                        <NavLink
                            key={item.to}
                            to={item.to}
                            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
                            end={item.to === '/'}
                        >
                            <span className="nav-item-icon">{item.icon}</span>
                            <span className="nav-item-text">{item.label}</span>
                        </NavLink>
                    ))}
                </div>

                {/* System Section */}
                <div className="nav-section">
                    <div className="nav-section-title">System</div>
                    {systemNavItems.map((item) => (
                        <NavLink
                            key={item.to}
                            to={item.to}
                            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
                        >
                            <span className="nav-item-icon">{item.icon}</span>
                            <span className="nav-item-text">{item.label}</span>
                        </NavLink>
                    ))}
                </div>
            </nav>

            {/* Footer */}
            <div className="sidebar-footer">
                <button className="sidebar-toggle" onClick={onToggle}>
                    {isCollapsed ? <ExpandIcon /> : <CollapseIcon />}
                    {!isCollapsed && <span>Collapse</span>}
                </button>
            </div>
        </aside>
    );
};

export default Sidebar;
