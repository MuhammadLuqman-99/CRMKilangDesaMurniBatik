// ============================================
// Header Component
// Production-Ready App Header
// ============================================

import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

// SVG Icons
const MenuIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="3" y1="12" x2="21" y2="12" /><line x1="3" y1="6" x2="21" y2="6" /><line x1="3" y1="18" x2="21" y2="18" />
    </svg>
);

const SearchIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
);

const BellIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" /><path d="M13.73 21a2 2 0 0 1-3.46 0" />
    </svg>
);

const LogOutIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><polyline points="16 17 21 12 16 7" /><line x1="21" y1="12" x2="9" y2="12" />
    </svg>
);

const UserIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
    </svg>
);

const SettingsIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" /><circle cx="12" cy="12" r="3" />
    </svg>
);

interface HeaderProps {
    isSidebarCollapsed: boolean;
    onToggleSidebar: () => void;
    onMobileMenuOpen: () => void;
}

export function Header({ isSidebarCollapsed, onMobileMenuOpen }: HeaderProps) {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const dropdownRef = useRef<HTMLDivElement>(null);

    // Close dropdown when clicking outside
    useEffect(() => {
        function handleClickOutside(event: MouseEvent) {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsDropdownOpen(false);
            }
        }

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleLogout = async () => {
        await logout();
        navigate('/login');
    };

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        if (searchQuery.trim()) {
            navigate(`/search?q=${encodeURIComponent(searchQuery)}`);
        }
    };

    const getInitials = (firstName?: string, lastName?: string) => {
        const first = firstName?.charAt(0) || '';
        const last = lastName?.charAt(0) || '';
        return (first + last).toUpperCase() || 'U';
    };

    return (
        <header className={`header ${isSidebarCollapsed ? 'sidebar-collapsed' : ''}`}>
            <div className="header-left">
                <button className="header-toggle" onClick={onMobileMenuOpen}>
                    <MenuIcon />
                </button>

                <form onSubmit={handleSearch} className="header-search">
                    <span className="header-search-icon">
                        <SearchIcon />
                    </span>
                    <input
                        type="text"
                        placeholder="Search leads, customers, opportunities..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="form-input"
                    />
                </form>
            </div>

            <div className="header-right">
                {/* Notifications */}
                <button className="header-icon-btn" title="Notifications">
                    <BellIcon />
                    <span className="badge">3</span>
                </button>

                {/* User Menu */}
                <div className="dropdown" ref={dropdownRef}>
                    <button
                        className="sidebar-user"
                        onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                    >
                        <div className="sidebar-user-avatar">
                            {getInitials(user?.first_name, user?.last_name)}
                        </div>
                        <div className="sidebar-user-info">
                            <div className="sidebar-user-name">{user?.first_name ? `${user.first_name} ${user.last_name || ''}`.trim() : 'User'}</div>
                            <div className="sidebar-user-email">{user?.email || ''}</div>
                        </div>
                    </button>

                    {isDropdownOpen && (
                        <div className="dropdown-menu" style={{ opacity: 1, visibility: 'visible', transform: 'translateY(0)' }}>
                            <div
                                className="dropdown-item"
                                onClick={() => {
                                    navigate('/settings/profile');
                                    setIsDropdownOpen(false);
                                }}
                            >
                                <UserIcon />
                                <span>Profile</span>
                            </div>
                            <div
                                className="dropdown-item"
                                onClick={() => {
                                    navigate('/settings');
                                    setIsDropdownOpen(false);
                                }}
                            >
                                <SettingsIcon />
                                <span>Settings</span>
                            </div>
                            <div className="dropdown-divider" />
                            <div className="dropdown-item danger" onClick={handleLogout}>
                                <LogOutIcon />
                                <span>Log out</span>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </header>
    );
}

export default Header;
