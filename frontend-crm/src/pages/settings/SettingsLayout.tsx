// ============================================
// Settings Layout
// Shared layout for all settings pages
// ============================================

import { NavLink, Outlet } from 'react-router-dom';

// SVG Icons
const UserIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
    </svg>
);

const ShieldIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
    </svg>
);

const BellIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" /><path d="M13.73 21a2 2 0 0 1-3.46 0" />
    </svg>
);

const UsersIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
);

const settingsNav = [
    { path: '/settings/profile', label: 'Profile', icon: UserIcon },
    { path: '/settings/security', label: 'Security', icon: ShieldIcon },
    { path: '/settings/notifications', label: 'Notifications', icon: BellIcon },
    { path: '/settings/team', label: 'Team Management', icon: UsersIcon },
];

export function SettingsLayout() {
    return (
        <div className="animate-fade-in">
            <div className="page-header">
                <div className="page-header-left">
                    <h1 className="page-title">Settings</h1>
                    <p className="page-description">Manage your account settings and preferences</p>
                </div>
            </div>

            <div className="settings-container">
                <nav className="settings-nav">
                    {settingsNav.map((item) => (
                        <NavLink
                            key={item.path}
                            to={item.path}
                            className={({ isActive }) =>
                                `settings-nav-item ${isActive ? 'active' : ''}`
                            }
                        >
                            <item.icon />
                            <span>{item.label}</span>
                        </NavLink>
                    ))}
                </nav>

                <div className="settings-content">
                    <Outlet />
                </div>
            </div>

            <style>{`
                .settings-container {
                    display: grid;
                    grid-template-columns: 250px 1fr;
                    gap: 1.5rem;
                }

                .settings-nav {
                    display: flex;
                    flex-direction: column;
                    gap: 0.25rem;
                    background: var(--bg-card);
                    border-radius: var(--radius-lg);
                    padding: 1rem;
                    height: fit-content;
                    border: 1px solid var(--border-color);
                }

                .settings-nav-item {
                    display: flex;
                    align-items: center;
                    gap: 0.75rem;
                    padding: 0.75rem 1rem;
                    border-radius: var(--radius-md);
                    color: var(--text-secondary);
                    text-decoration: none;
                    font-size: 0.875rem;
                    font-weight: 500;
                    transition: all 0.15s ease;
                }

                .settings-nav-item:hover {
                    background: var(--bg-tertiary);
                    color: var(--text-primary);
                }

                .settings-nav-item.active {
                    background: var(--primary);
                    color: white;
                }

                .settings-content {
                    min-width: 0;
                }

                @media (max-width: 768px) {
                    .settings-container {
                        grid-template-columns: 1fr;
                    }

                    .settings-nav {
                        flex-direction: row;
                        overflow-x: auto;
                        padding: 0.5rem;
                    }

                    .settings-nav-item {
                        white-space: nowrap;
                        padding: 0.5rem 1rem;
                    }

                    .settings-nav-item span {
                        display: none;
                    }
                }
            `}</style>
        </div>
    );
}

export default SettingsLayout;
