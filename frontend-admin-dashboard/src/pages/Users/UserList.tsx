// ============================================
// User List Page
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React, { useState, useEffect, useCallback } from 'react';
import { userApi, tenantApi } from '../../services/api';
import type { User, Tenant, UserStatus, Pagination } from '../../types';

// ============================================
// Icons
// ============================================

const SearchIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <circle cx="11" cy="11" r="8" />
        <path d="m21 21-4.35-4.35" />
    </svg>
);

const KeyIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <circle cx="7.5" cy="15.5" r="5.5" />
        <path d="m21 2-9.6 9.6" />
        <path d="m15.5 7.5 3 3L22 7l-3-3" />
    </svg>
);

const UserIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <circle cx="12" cy="8" r="5" />
        <path d="M20 21a8 8 0 0 0-16 0" />
    </svg>
);

const CloseIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M18 6 6 18M6 6l12 12" />
    </svg>
);

// ============================================
// Modal Component
// ============================================

interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
}

const Modal: React.FC<ModalProps> = ({ isOpen, onClose, title, children }) => {
    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal" onClick={e => e.stopPropagation()}>
                <div className="modal-header">
                    <h2 className="modal-title">{title}</h2>
                    <button className="modal-close" onClick={onClose}>
                        <CloseIcon />
                    </button>
                </div>
                <div className="modal-body">
                    {children}
                </div>
            </div>
        </div>
    );
};

// ============================================
// User Details Modal
// ============================================

interface UserDetailsModalProps {
    isOpen: boolean;
    onClose: () => void;
    user: User | null;
}

const UserDetailsModal: React.FC<UserDetailsModalProps> = ({ isOpen, onClose, user }) => {
    if (!user) return null;

    return (
        <Modal isOpen={isOpen} onClose={onClose} title="User Details">
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)', marginBottom: 'var(--space-6)' }}>
                <div style={{
                    width: 64,
                    height: 64,
                    borderRadius: 'var(--radius-full)',
                    background: 'linear-gradient(135deg, var(--primary), var(--primary-light))',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '1.5rem',
                    fontWeight: 700,
                    color: 'white',
                }}>
                    {user.firstName?.charAt(0) || ''}{user.lastName?.charAt(0) || ''}
                </div>
                <div>
                    <h3 style={{ fontSize: '1.125rem', fontWeight: 600, color: 'var(--text-primary)' }}>
                        {user.firstName} {user.lastName}
                    </h3>
                    <p style={{ color: 'var(--text-muted)', fontSize: '0.875rem' }}>{user.email}</p>
                </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--space-4)' }}>
                <div>
                    <div className="text-xs text-muted mb-2">Status</div>
                    <span className={`badge badge-${user.status}`}>{user.status}</span>
                </div>
                <div>
                    <div className="text-xs text-muted mb-2">Tenant</div>
                    <span className="text-primary">{user.tenantName || user.tenantId}</span>
                </div>
                <div>
                    <div className="text-xs text-muted mb-2">Phone</div>
                    <span>{user.phone || 'Not provided'}</span>
                </div>
                <div>
                    <div className="text-xs text-muted mb-2">Last Login</div>
                    <span>{user.lastLoginAt ? new Date(user.lastLoginAt).toLocaleString() : 'Never'}</span>
                </div>
            </div>

            {user.roles && user.roles.length > 0 && (
                <div style={{ marginTop: 'var(--space-6)' }}>
                    <div className="text-xs text-muted mb-2">Roles</div>
                    <div style={{ display: 'flex', gap: 'var(--space-2)', flexWrap: 'wrap' }}>
                        {user.roles.map(role => (
                            <span key={role.id} className="badge badge-pro">{role.name}</span>
                        ))}
                    </div>
                </div>
            )}

            <div style={{ marginTop: 'var(--space-4)', padding: 'var(--space-4)', background: 'var(--bg-tertiary)', borderRadius: 'var(--radius-lg)' }}>
                <div className="text-xs text-muted">Created</div>
                <div className="text-sm">{new Date(user.createdAt).toLocaleString()}</div>
                <div className="text-xs text-muted mt-2">Last Updated</div>
                <div className="text-sm">{new Date(user.updatedAt).toLocaleString()}</div>
            </div>
        </Modal>
    );
};

// ============================================
// Reset Password Modal
// ============================================

interface ResetPasswordModalProps {
    isOpen: boolean;
    onClose: () => void;
    user: User | null;
    onReset: (userId: string) => Promise<void>;
}

const ResetPasswordModal: React.FC<ResetPasswordModalProps> = ({ isOpen, onClose, user, onReset }) => {
    const [isLoading, setIsLoading] = useState(false);
    const [success, setSuccess] = useState(false);

    const handleReset = async () => {
        if (!user) return;
        setIsLoading(true);
        try {
            await onReset(user.id);
            setSuccess(true);
        } finally {
            setIsLoading(false);
        }
    };

    if (!user) return null;

    return (
        <Modal isOpen={isOpen} onClose={() => { onClose(); setSuccess(false); }} title="Reset Password">
            {success ? (
                <div style={{ textAlign: 'center', padding: 'var(--space-4)' }}>
                    <div style={{ fontSize: '3rem', marginBottom: 'var(--space-4)' }}>✓</div>
                    <h3 style={{ color: 'var(--success)', marginBottom: 'var(--space-2)' }}>Password Reset Successful</h3>
                    <p style={{ color: 'var(--text-muted)', fontSize: '0.875rem' }}>
                        A password reset email has been sent to {user.email}
                    </p>
                    <button className="btn btn-primary" onClick={() => { onClose(); setSuccess(false); }} style={{ marginTop: 'var(--space-4)' }}>
                        Done
                    </button>
                </div>
            ) : (
                <>
                    <p style={{ color: 'var(--text-secondary)', marginBottom: 'var(--space-4)' }}>
                        Are you sure you want to reset the password for <strong>{user.firstName} {user.lastName}</strong>?
                    </p>
                    <p style={{ color: 'var(--text-muted)', fontSize: '0.875rem', marginBottom: 'var(--space-4)' }}>
                        They will receive an email with instructions to set a new password.
                    </p>
                    <div className="modal-footer" style={{ margin: '0 calc(-1 * var(--space-5))', marginBottom: 'calc(-1 * var(--space-5))' }}>
                        <button className="btn btn-secondary" onClick={onClose}>
                            Cancel
                        </button>
                        <button className="btn btn-warning" onClick={handleReset} disabled={isLoading}>
                            {isLoading ? 'Sending...' : 'Reset Password'}
                        </button>
                    </div>
                </>
            )}
        </Modal>
    );
};

// ============================================
// Main Component
// ============================================

const UserList: React.FC = () => {
    const [users, setUsers] = useState<User[]>([]);
    const [tenants, setTenants] = useState<Tenant[]>([]);
    const [pagination, setPagination] = useState<Pagination>({
        page: 1,
        perPage: 10,
        total: 0,
        totalPages: 0,
    });
    const [isLoading, setIsLoading] = useState(true);
    const [search, setSearch] = useState('');
    const [statusFilter, setStatusFilter] = useState<string>('');
    const [tenantFilter, setTenantFilter] = useState<string>('');

    // Modal states
    const [isDetailsOpen, setIsDetailsOpen] = useState(false);
    const [isResetOpen, setIsResetOpen] = useState(false);
    const [selectedUser, setSelectedUser] = useState<User | null>(null);

    // Fetch tenants for filter dropdown
    useEffect(() => {
        const fetchTenants = async () => {
            try {
                const response = await tenantApi.list({ pageSize: 100 });
                if (response.success) {
                    setTenants(response.data);
                }
            } catch (error) {
                console.error('Failed to fetch tenants:', error);
            }
        };
        fetchTenants();
    }, []);

    const fetchUsers = useCallback(async () => {
        setIsLoading(true);
        try {
            const response = await userApi.list({
                page: pagination.page,
                pageSize: pagination.perPage,
                search: search || undefined,
                status: statusFilter as UserStatus || undefined,
                tenantId: tenantFilter || undefined,
            });

            if (response.success) {
                // Add tenant name to users
                const usersWithTenantNames = response.data.map(user => {
                    const tenant = tenants.find(t => t.id === user.tenantId);
                    return { ...user, tenantName: tenant?.name };
                });
                setUsers(usersWithTenantNames);
                setPagination(response.meta);
            }
        } catch (error) {
            console.error('Failed to fetch users:', error);
        } finally {
            setIsLoading(false);
        }
    }, [pagination.page, pagination.perPage, search, statusFilter, tenantFilter, tenants]);

    useEffect(() => {
        if (tenants.length > 0 || !tenantFilter) {
            fetchUsers();
        }
    }, [fetchUsers, tenants.length, tenantFilter]);

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        setPagination(prev => ({ ...prev, page: 1 }));
        fetchUsers();
    };

    const handleStatusChange = async (user: User, newStatus: UserStatus) => {
        try {
            await userApi.updateStatus(user.id, newStatus);
            fetchUsers();
        } catch (error) {
            console.error('Failed to update user status:', error);
        }
    };

    const handleResetPassword = async (userId: string) => {
        await userApi.resetPassword(userId, { sendEmail: true });
    };

    const openDetails = (user: User) => {
        setSelectedUser(user);
        setIsDetailsOpen(true);
    };

    const openResetPassword = (user: User) => {
        setSelectedUser(user);
        setIsResetOpen(true);
    };

    return (
        <div>
            {/* Page Header */}
            <div className="page-header">
                <div>
                    <h1 className="page-title">User Management</h1>
                    <p className="page-description">View and manage users across all tenants</p>
                </div>
            </div>

            {/* Search & Filters */}
            <div className="search-filter-bar">
                <form className="search-input-wrapper" onSubmit={handleSearch}>
                    <SearchIcon />
                    <input
                        type="text"
                        className="form-input"
                        placeholder="Search by name or email..."
                        value={search}
                        onChange={e => setSearch(e.target.value)}
                        style={{ paddingLeft: '2.5rem' }}
                    />
                </form>

                <div className="filter-group">
                    <select
                        className="form-select"
                        value={tenantFilter}
                        onChange={e => { setTenantFilter(e.target.value); setPagination(prev => ({ ...prev, page: 1 })); }}
                        style={{ minWidth: 180 }}
                    >
                        <option value="">All Tenants</option>
                        {tenants.map(tenant => (
                            <option key={tenant.id} value={tenant.id}>{tenant.name}</option>
                        ))}
                    </select>

                    <select
                        className="form-select"
                        value={statusFilter}
                        onChange={e => { setStatusFilter(e.target.value); setPagination(prev => ({ ...prev, page: 1 })); }}
                        style={{ minWidth: 140 }}
                    >
                        <option value="">All Status</option>
                        <option value="active">Active</option>
                        <option value="inactive">Inactive</option>
                        <option value="suspended">Suspended</option>
                        <option value="pending_verification">Pending</option>
                    </select>
                </div>
            </div>

            {/* Users Table */}
            <div className="card">
                <div className="table-container">
                    <table className="table">
                        <thead>
                            <tr>
                                <th>User</th>
                                <th>Email</th>
                                <th>Tenant</th>
                                <th>Status</th>
                                <th>Last Login</th>
                                <th style={{ textAlign: 'right' }}>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {isLoading ? (
                                <tr>
                                    <td colSpan={6} style={{ textAlign: 'center', padding: 'var(--space-8)' }}>
                                        <div className="loading-spinner" style={{ margin: '0 auto' }} />
                                    </td>
                                </tr>
                            ) : users.length === 0 ? (
                                <tr>
                                    <td colSpan={6}>
                                        <div className="empty-state">
                                            <div className="empty-state-icon">👤</div>
                                            <div className="empty-state-title">No users found</div>
                                            <div className="empty-state-description">
                                                Adjust your filters or search to find users
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                users.map(user => (
                                    <tr key={user.id}>
                                        <td>
                                            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                                                <div style={{
                                                    width: 36,
                                                    height: 36,
                                                    borderRadius: 'var(--radius-full)',
                                                    background: 'linear-gradient(135deg, var(--primary), var(--primary-light))',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    justifyContent: 'center',
                                                    fontSize: '0.75rem',
                                                    fontWeight: 600,
                                                    color: 'white',
                                                }}>
                                                    {user.firstName?.charAt(0) || ''}{user.lastName?.charAt(0) || ''}
                                                </div>
                                                <div>
                                                    <div className="text-primary">{user.firstName} {user.lastName}</div>
                                                </div>
                                            </div>
                                        </td>
                                        <td>{user.email}</td>
                                        <td>
                                            <span style={{ fontSize: '0.875rem' }}>{user.tenantName || 'Unknown'}</span>
                                        </td>
                                        <td>
                                            <select
                                                className="form-select"
                                                value={user.status}
                                                onChange={e => handleStatusChange(user, e.target.value as UserStatus)}
                                                style={{
                                                    padding: '4px 8px',
                                                    fontSize: '0.75rem',
                                                    minWidth: 120,
                                                    background: user.status === 'active' ? 'var(--success-bg)' :
                                                        user.status === 'suspended' ? 'var(--danger-bg)' : 'var(--bg-tertiary)'
                                                }}
                                            >
                                                <option value="active">Active</option>
                                                <option value="inactive">Inactive</option>
                                                <option value="suspended">Suspended</option>
                                                <option value="pending_verification">Pending</option>
                                            </select>
                                        </td>
                                        <td>
                                            {user.lastLoginAt
                                                ? new Date(user.lastLoginAt).toLocaleDateString()
                                                : <span className="text-muted">Never</span>
                                            }
                                        </td>
                                        <td>
                                            <div style={{ display: 'flex', gap: 'var(--space-2)', justifyContent: 'flex-end' }}>
                                                <button
                                                    className="btn btn-ghost btn-icon"
                                                    onClick={() => openDetails(user)}
                                                    title="View Details"
                                                >
                                                    <UserIcon />
                                                </button>
                                                <button
                                                    className="btn btn-ghost btn-icon"
                                                    onClick={() => openResetPassword(user)}
                                                    title="Reset Password"
                                                    style={{ color: 'var(--warning)' }}
                                                >
                                                    <KeyIcon />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Pagination */}
                {!isLoading && users.length > 0 && (
                    <div className="pagination">
                        <div className="pagination-info">
                            Showing {((pagination.page - 1) * pagination.perPage) + 1} to {Math.min(pagination.page * pagination.perPage, pagination.total)} of {pagination.total} users
                        </div>
                        <div className="pagination-controls">
                            <button
                                className="pagination-btn"
                                disabled={pagination.page <= 1}
                                onClick={() => setPagination(prev => ({ ...prev, page: prev.page - 1 }))}
                            >
                                Previous
                            </button>
                            {Array.from({ length: Math.min(5, pagination.totalPages) }, (_, i) => {
                                const page = i + 1;
                                return (
                                    <button
                                        key={page}
                                        className={`pagination-btn ${page === pagination.page ? 'active' : ''}`}
                                        onClick={() => setPagination(prev => ({ ...prev, page }))}
                                    >
                                        {page}
                                    </button>
                                );
                            })}
                            <button
                                className="pagination-btn"
                                disabled={pagination.page >= pagination.totalPages}
                                onClick={() => setPagination(prev => ({ ...prev, page: prev.page + 1 }))}
                            >
                                Next
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Modals */}
            <UserDetailsModal
                isOpen={isDetailsOpen}
                onClose={() => { setIsDetailsOpen(false); setSelectedUser(null); }}
                user={selectedUser}
            />

            <ResetPasswordModal
                isOpen={isResetOpen}
                onClose={() => { setIsResetOpen(false); setSelectedUser(null); }}
                user={selectedUser}
                onReset={handleResetPassword}
            />
        </div>
    );
};

export default UserList;
