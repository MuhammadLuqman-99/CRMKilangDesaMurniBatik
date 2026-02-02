// ============================================
// Sidebar Component
// Production-Ready Navigation Sidebar
// ============================================

import { NavLink, useLocation } from 'react-router-dom';

// SVG Icons as components
const DashboardIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="7" height="7" /><rect x="14" y="3" width="7" height="7" /><rect x="14" y="14" width="7" height="7" /><rect x="3" y="14" width="7" height="7" />
    </svg>
);

const PipelineIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3 3v18h18" /><path d="m19 9-5 5-4-4-3 3" />
    </svg>
);

const LeadsIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
);

const OpportunitiesIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" /><path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8" /><path d="M12 18V6" />
    </svg>
);

const CustomersIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
    </svg>
);

const SettingsIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" /><circle cx="12" cy="12" r="3" />
    </svg>
);

interface NavItemProps {
    to: string;
    icon: React.ReactNode;
    label: string;
    badge?: number;
    isCollapsed: boolean;
}

function NavItem({ to, icon, label, badge, isCollapsed }: NavItemProps) {
    return (
        <NavLink
            to={to}
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
        >
            <span className="nav-item-icon">{icon}</span>
            {!isCollapsed && <span className="nav-item-text">{label}</span>}
            {!isCollapsed && badge !== undefined && badge > 0 && (
                <span className="nav-item-badge">{badge > 99 ? '99+' : badge}</span>
            )}
        </NavLink>
    );
}

interface SidebarProps {
    isCollapsed: boolean;
    onToggle: () => void;
    isMobileOpen: boolean;
    onMobileClose: () => void;
}

export function Sidebar({ isCollapsed, isMobileOpen, onMobileClose }: SidebarProps) {
    const location = useLocation();

    // Close mobile sidebar on route change
    if (isMobileOpen && location.pathname) {
        // Using a slight delay to allow animation
        setTimeout(onMobileClose, 100);
    }

    return (
        <>
            {/* Mobile overlay */}
            {isMobileOpen && (
                <div
                    className="modal-overlay"
                    style={{ zIndex: 90 }}
                    onClick={onMobileClose}
                />
            )}

            <aside className={`sidebar ${isCollapsed ? 'collapsed' : ''} ${isMobileOpen ? 'open' : ''}`}>
                {/* Header */}
                <div className="sidebar-header">
                    <a href="/" className="sidebar-logo">
                        <div className="sidebar-logo-icon">CRM</div>
                        <span className="sidebar-logo-text">Kilang Batik</span>
                    </a>
                </div>

                {/* Navigation */}
                <nav className="sidebar-nav">
                    <div className="nav-section">
                        <div className="nav-section-title">Main</div>
                        <NavItem
                            to="/"
                            icon={<DashboardIcon />}
                            label="Dashboard"
                            isCollapsed={isCollapsed}
                        />
                        <NavItem
                            to="/pipeline"
                            icon={<PipelineIcon />}
                            label="Pipeline"
                            isCollapsed={isCollapsed}
                        />
                    </div>

                    <div className="nav-section">
                        <div className="nav-section-title">Sales</div>
                        <NavItem
                            to="/leads"
                            icon={<LeadsIcon />}
                            label="Leads"
                            isCollapsed={isCollapsed}
                        />
                        <NavItem
                            to="/opportunities"
                            icon={<OpportunitiesIcon />}
                            label="Opportunities"
                            isCollapsed={isCollapsed}
                        />
                    </div>

                    <div className="nav-section">
                        <div className="nav-section-title">Customers</div>
                        <NavItem
                            to="/customers"
                            icon={<CustomersIcon />}
                            label="Customers"
                            isCollapsed={isCollapsed}
                        />
                    </div>

                    <div className="nav-section">
                        <div className="nav-section-title">Settings</div>
                        <NavItem
                            to="/settings"
                            icon={<SettingsIcon />}
                            label="Settings"
                            isCollapsed={isCollapsed}
                        />
                    </div>
                </nav>
            </aside>
        </>
    );
}

export default Sidebar;
