// ============================================
// Header Component
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

interface HeaderProps {
    title: string;
    isSidebarCollapsed: boolean;
}

const LogoutIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
        <polyline points="16 17 21 12 16 7" />
        <line x1="21" y1="12" x2="9" y2="12" />
    </svg>
);

const ChevronDownIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="m6 9 6 6 6-6" />
    </svg>
);

const Header: React.FC<HeaderProps> = ({ title, isSidebarCollapsed }) => {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);

    const handleLogout = async () => {
        await logout();
        navigate('/login');
    };

    const getInitials = (firstName?: string, lastName?: string): string => {
        const first = firstName?.charAt(0)?.toUpperCase() || '';
        const last = lastName?.charAt(0)?.toUpperCase() || '';
        return first + last || 'AD';
    };

    return (
        <header className={`header ${isSidebarCollapsed ? 'sidebar-collapsed' : ''}`}>
            <h1 className="header-title">{title}</h1>

            <div className="header-actions">
                {/* User Menu */}
                <div
                    className="header-user"
                    onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                    style={{ position: 'relative' }}
                >
                    <div className="header-avatar">
                        {getInitials(user?.firstName, user?.lastName)}
                    </div>
                    <div className="header-user-info">
                        <span className="header-user-name">
                            {user?.firstName || 'Admin'} {user?.lastName || 'User'}
                        </span>
                        <span className="header-user-role">Administrator</span>
                    </div>
                    <ChevronDownIcon />

                    {/* Dropdown Menu */}
                    {isDropdownOpen && (
                        <div
                            style={{
                                position: 'absolute',
                                top: '100%',
                                right: 0,
                                marginTop: '8px',
                                background: 'var(--bg-secondary)',
                                border: '1px solid var(--border-color)',
                                borderRadius: 'var(--radius-lg)',
                                minWidth: '180px',
                                boxShadow: 'var(--shadow-lg)',
                                zIndex: 100,
                                overflow: 'hidden',
                            }}
                        >
                            <button
                                onClick={handleLogout}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '8px',
                                    width: '100%',
                                    padding: '12px 16px',
                                    background: 'none',
                                    border: 'none',
                                    color: 'var(--danger)',
                                    cursor: 'pointer',
                                    fontSize: '0.875rem',
                                    textAlign: 'left',
                                }}
                            >
                                <LogoutIcon />
                                Sign Out
                            </button>
                        </div>
                    )}
                </div>
            </div>
        </header>
    );
};

export default Header;
