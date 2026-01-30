// ============================================
// Tenant List Page
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React, { useState, useEffect, useCallback } from 'react';
import { tenantApi } from '../../services/api';
import type { Tenant, TenantPlan, TenantStatus, Pagination } from '../../types';

// ============================================
// Icons
// ============================================

const SearchIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <circle cx="11" cy="11" r="8" />
        <path d="m21 21-4.35-4.35" />
    </svg>
);

const PlusIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 5v14M5 12h14" />
    </svg>
);

const EditIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
        <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
    </svg>
);

const TrashIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
    </svg>
);

const StatsIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M3 3v18h18" />
        <path d="M18 17V9M13 17V5M8 17v-3" />
    </svg>
);

const CloseIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M18 6 6 18M6 6l12 12" />
    </svg>
);

// ============================================
// Modal Components
// ============================================

interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
    size?: 'sm' | 'md' | 'lg';
}

const Modal: React.FC<ModalProps> = ({ isOpen, onClose, title, children, size = 'md' }) => {
    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div
                className={`modal ${size === 'lg' ? 'modal-lg' : ''}`}
                onClick={e => e.stopPropagation()}
            >
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
// Tenant Form Modal
// ============================================

interface TenantFormData {
    name: string;
    slug: string;
    plan: TenantPlan;
}

interface TenantFormModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSubmit: (data: TenantFormData) => Promise<void>;
    tenant?: Tenant;
}

const TenantFormModal: React.FC<TenantFormModalProps> = ({
    isOpen,
    onClose,
    onSubmit,
    tenant,
}) => {
    const [formData, setFormData] = useState<TenantFormData>({
        name: tenant?.name || '',
        slug: tenant?.slug || '',
        plan: tenant?.plan || 'free',
    });
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        if (tenant) {
            setFormData({
                name: tenant.name,
                slug: tenant.slug,
                plan: tenant.plan,
            });
        } else {
            setFormData({ name: '', slug: '', plan: 'free' });
        }
    }, [tenant, isOpen]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);
        try {
            await onSubmit(formData);
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to save tenant');
        } finally {
            setIsLoading(false);
        }
    };

    // Auto-generate slug from name
    const handleNameChange = (name: string) => {
        setFormData(prev => ({
            ...prev,
            name,
            slug: !tenant ? name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') : prev.slug,
        }));
    };

    return (
        <Modal
            isOpen={isOpen}
            onClose={onClose}
            title={tenant ? 'Edit Tenant' : 'Create New Tenant'}
        >
            <form onSubmit={handleSubmit}>
                {error && (
                    <div className="login-error" style={{ marginBottom: 'var(--space-4)' }}>
                        {error}
                    </div>
                )}

                <div className="form-group">
                    <label className="form-label">Tenant Name</label>
                    <input
                        type="text"
                        className="form-input"
                        value={formData.name}
                        onChange={e => handleNameChange(e.target.value)}
                        placeholder="Enter tenant name"
                        required
                        minLength={2}
                        maxLength={100}
                    />
                </div>

                <div className="form-group">
                    <label className="form-label">Slug</label>
                    <input
                        type="text"
                        className="form-input"
                        value={formData.slug}
                        onChange={e => setFormData(prev => ({ ...prev, slug: e.target.value }))}
                        placeholder="tenant-slug"
                        required
                        pattern="[a-z0-9-]+"
                        disabled={!!tenant}
                    />
                    <p className="form-help">Lowercase letters, numbers, and hyphens only</p>
                </div>

                <div className="form-group">
                    <label className="form-label">Subscription Plan</label>
                    <select
                        className="form-select"
                        value={formData.plan}
                        onChange={e => setFormData(prev => ({ ...prev, plan: e.target.value as TenantPlan }))}
                    >
                        <option value="free">Free</option>
                        <option value="starter">Starter</option>
                        <option value="pro">Professional</option>
                        <option value="enterprise">Enterprise</option>
                    </select>
                </div>

                <div className="modal-footer" style={{ margin: '0 calc(-1 * var(--space-5))', marginBottom: 'calc(-1 * var(--space-5))' }}>
                    <button type="button" className="btn btn-secondary" onClick={onClose}>
                        Cancel
                    </button>
                    <button type="submit" className="btn btn-primary" disabled={isLoading}>
                        {isLoading ? 'Saving...' : tenant ? 'Update Tenant' : 'Create Tenant'}
                    </button>
                </div>
            </form>
        </Modal>
    );
};

// ============================================
// Delete Confirm Modal
// ============================================

interface ConfirmModalProps {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => Promise<void>;
    title: string;
    message: string;
}

const ConfirmModal: React.FC<ConfirmModalProps> = ({
    isOpen,
    onClose,
    onConfirm,
    title,
    message,
}) => {
    const [isLoading, setIsLoading] = useState(false);

    const handleConfirm = async () => {
        setIsLoading(true);
        try {
            await onConfirm();
            onClose();
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={title} size="sm">
            <p style={{ color: 'var(--text-secondary)', marginBottom: 'var(--space-4)' }}>
                {message}
            </p>
            <div className="modal-footer" style={{ margin: '0 calc(-1 * var(--space-5))', marginBottom: 'calc(-1 * var(--space-5))' }}>
                <button type="button" className="btn btn-secondary" onClick={onClose}>
                    Cancel
                </button>
                <button className="btn btn-danger" onClick={handleConfirm} disabled={isLoading}>
                    {isLoading ? 'Deleting...' : 'Delete'}
                </button>
            </div>
        </Modal>
    );
};

// ============================================
// Stats Modal
// ============================================

interface StatsModalProps {
    isOpen: boolean;
    onClose: () => void;
    tenant: Tenant | null;
}

const StatsModal: React.FC<StatsModalProps> = ({ isOpen, onClose, tenant }) => {
    if (!tenant) return null;

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={`Statistics: ${tenant.name}`}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--space-4)' }}>
                <div className="stats-card" style={{ border: 'none', padding: 'var(--space-4)', background: 'var(--bg-tertiary)' }}>
                    <div className="stats-label">Users</div>
                    <div className="stats-value">{tenant.usage?.userCount || 0}</div>
                    <div className="text-xs text-muted">of {tenant.limits?.maxUsers || '∞'} max</div>
                </div>
                <div className="stats-card" style={{ border: 'none', padding: 'var(--space-4)', background: 'var(--bg-tertiary)' }}>
                    <div className="stats-label">Contacts</div>
                    <div className="stats-value">{tenant.usage?.contactCount || 0}</div>
                    <div className="text-xs text-muted">of {tenant.limits?.maxContacts || '∞'} max</div>
                </div>
            </div>

            {tenant.trialInfo?.isTrialing && (
                <div style={{ marginTop: 'var(--space-4)', padding: 'var(--space-4)', background: 'var(--warning-bg)', borderRadius: 'var(--radius-lg)', border: '1px solid var(--warning-border)' }}>
                    <div style={{ fontWeight: 600, color: 'var(--warning)', marginBottom: 'var(--space-2)' }}>
                        Trial Active
                    </div>
                    <div style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                        {tenant.trialInfo.daysLeft} days remaining
                    </div>
                </div>
            )}

            <div style={{ marginTop: 'var(--space-4)' }}>
                <h4 style={{ marginBottom: 'var(--space-2)', fontSize: '0.875rem', fontWeight: 600 }}>Settings</h4>
                <div style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                    <div>Timezone: {tenant.settings?.timezone || 'UTC'}</div>
                    <div>Currency: {tenant.settings?.currency || 'USD'}</div>
                    <div>Language: {tenant.settings?.language || 'en'}</div>
                </div>
            </div>
        </Modal>
    );
};

// ============================================
// Main Component
// ============================================

const TenantList: React.FC = () => {
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
    const [planFilter, setPlanFilter] = useState<string>('');

    // Modal states
    const [isFormOpen, setIsFormOpen] = useState(false);
    const [isDeleteOpen, setIsDeleteOpen] = useState(false);
    const [isStatsOpen, setIsStatsOpen] = useState(false);
    const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null);

    const fetchTenants = useCallback(async () => {
        setIsLoading(true);
        try {
            const response = await tenantApi.list({
                page: pagination.page,
                pageSize: pagination.perPage,
                search: search || undefined,
                status: statusFilter || undefined,
                plan: planFilter || undefined,
            });

            if (response.success) {
                setTenants(response.data);
                setPagination(response.meta);
            }
        } catch (error) {
            console.error('Failed to fetch tenants:', error);
        } finally {
            setIsLoading(false);
        }
    }, [pagination.page, pagination.perPage, search, statusFilter, planFilter]);

    useEffect(() => {
        fetchTenants();
    }, [fetchTenants]);

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        setPagination(prev => ({ ...prev, page: 1 }));
        fetchTenants();
    };

    const handleCreate = async (data: TenantFormData) => {
        await tenantApi.create(data);
        fetchTenants();
    };

    const handleUpdate = async (data: TenantFormData) => {
        if (!selectedTenant) return;
        await tenantApi.update(selectedTenant.id, { name: data.name });
        if (data.plan !== selectedTenant.plan) {
            await tenantApi.updatePlan(selectedTenant.id, data.plan);
        }
        fetchTenants();
    };

    const handleDelete = async () => {
        if (!selectedTenant) return;
        await tenantApi.delete(selectedTenant.id);
        setSelectedTenant(null);
        fetchTenants();
    };

    const handleStatusChange = async (tenant: Tenant, newStatus: TenantStatus) => {
        await tenantApi.updateStatus(tenant.id, newStatus);
        fetchTenants();
    };

    const openEdit = (tenant: Tenant) => {
        setSelectedTenant(tenant);
        setIsFormOpen(true);
    };

    const openDelete = (tenant: Tenant) => {
        setSelectedTenant(tenant);
        setIsDeleteOpen(true);
    };

    const openStats = (tenant: Tenant) => {
        setSelectedTenant(tenant);
        setIsStatsOpen(true);
    };

    return (
        <div>
            {/* Page Header */}
            <div className="page-header">
                <div>
                    <h1 className="page-title">Tenant Management</h1>
                    <p className="page-description">Manage all tenants, their subscription plans, and status</p>
                </div>
                <button className="btn btn-primary" onClick={() => { setSelectedTenant(null); setIsFormOpen(true); }}>
                    <PlusIcon />
                    Create Tenant
                </button>
            </div>

            {/* Search & Filters */}
            <div className="search-filter-bar">
                <form className="search-input-wrapper" onSubmit={handleSearch}>
                    <SearchIcon />
                    <input
                        type="text"
                        className="form-input"
                        placeholder="Search by name or slug..."
                        value={search}
                        onChange={e => setSearch(e.target.value)}
                        style={{ paddingLeft: '2.5rem' }}
                    />
                </form>

                <div className="filter-group">
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
                        <option value="trial">Trial</option>
                        <option value="pending">Pending</option>
                    </select>

                    <select
                        className="form-select"
                        value={planFilter}
                        onChange={e => { setPlanFilter(e.target.value); setPagination(prev => ({ ...prev, page: 1 })); }}
                        style={{ minWidth: 140 }}
                    >
                        <option value="">All Plans</option>
                        <option value="free">Free</option>
                        <option value="starter">Starter</option>
                        <option value="pro">Professional</option>
                        <option value="enterprise">Enterprise</option>
                    </select>
                </div>
            </div>

            {/* Tenants Table */}
            <div className="card">
                <div className="table-container">
                    <table className="table">
                        <thead>
                            <tr>
                                <th>Tenant Name</th>
                                <th>Slug</th>
                                <th>Status</th>
                                <th>Plan</th>
                                <th>Users</th>
                                <th>Created</th>
                                <th style={{ textAlign: 'right' }}>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {isLoading ? (
                                <tr>
                                    <td colSpan={7} style={{ textAlign: 'center', padding: 'var(--space-8)' }}>
                                        <div className="loading-spinner" style={{ margin: '0 auto' }} />
                                    </td>
                                </tr>
                            ) : tenants.length === 0 ? (
                                <tr>
                                    <td colSpan={7}>
                                        <div className="empty-state">
                                            <div className="empty-state-icon">🏢</div>
                                            <div className="empty-state-title">No tenants found</div>
                                            <div className="empty-state-description">
                                                Create your first tenant to get started
                                            </div>
                                            <button className="btn btn-primary" onClick={() => { setSelectedTenant(null); setIsFormOpen(true); }}>
                                                <PlusIcon />
                                                Create Tenant
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                tenants.map(tenant => (
                                    <tr key={tenant.id}>
                                        <td className="text-primary">{tenant.name}</td>
                                        <td>
                                            <code style={{ fontSize: '0.75rem', color: 'var(--text-muted)', background: 'var(--bg-tertiary)', padding: '2px 6px', borderRadius: 4 }}>
                                                {tenant.slug}
                                            </code>
                                        </td>
                                        <td>
                                            <select
                                                className="form-select"
                                                value={tenant.status}
                                                onChange={e => handleStatusChange(tenant, e.target.value as TenantStatus)}
                                                style={{
                                                    padding: '4px 8px',
                                                    fontSize: '0.75rem',
                                                    minWidth: 110,
                                                    background: tenant.status === 'active' ? 'var(--success-bg)' :
                                                        tenant.status === 'suspended' ? 'var(--danger-bg)' : 'var(--bg-tertiary)'
                                                }}
                                            >
                                                <option value="active">Active</option>
                                                <option value="inactive">Inactive</option>
                                                <option value="suspended">Suspended</option>
                                                <option value="pending">Pending</option>
                                                <option value="trial">Trial</option>
                                            </select>
                                        </td>
                                        <td>
                                            <span className={`badge badge-${tenant.plan}`}>
                                                {tenant.plan}
                                            </span>
                                        </td>
                                        <td>{tenant.usage?.userCount || 0}</td>
                                        <td>{new Date(tenant.createdAt).toLocaleDateString()}</td>
                                        <td>
                                            <div style={{ display: 'flex', gap: 'var(--space-2)', justifyContent: 'flex-end' }}>
                                                <button className="btn btn-ghost btn-icon" onClick={() => openStats(tenant)} title="View Statistics">
                                                    <StatsIcon />
                                                </button>
                                                <button className="btn btn-ghost btn-icon" onClick={() => openEdit(tenant)} title="Edit">
                                                    <EditIcon />
                                                </button>
                                                <button className="btn btn-ghost btn-icon" onClick={() => openDelete(tenant)} title="Delete" style={{ color: 'var(--danger)' }}>
                                                    <TrashIcon />
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
                {!isLoading && tenants.length > 0 && (
                    <div className="pagination">
                        <div className="pagination-info">
                            Showing {((pagination.page - 1) * pagination.perPage) + 1} to {Math.min(pagination.page * pagination.perPage, pagination.total)} of {pagination.total} tenants
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
            <TenantFormModal
                isOpen={isFormOpen}
                onClose={() => { setIsFormOpen(false); setSelectedTenant(null); }}
                onSubmit={selectedTenant ? handleUpdate : handleCreate}
                tenant={selectedTenant || undefined}
            />

            <ConfirmModal
                isOpen={isDeleteOpen}
                onClose={() => { setIsDeleteOpen(false); setSelectedTenant(null); }}
                onConfirm={handleDelete}
                title="Delete Tenant"
                message={`Are you sure you want to delete "${selectedTenant?.name}"? This action cannot be undone and will delete all associated data.`}
            />

            <StatsModal
                isOpen={isStatsOpen}
                onClose={() => { setIsStatsOpen(false); setSelectedTenant(null); }}
                tenant={selectedTenant}
            />
        </div>
    );
};

export default TenantList;
