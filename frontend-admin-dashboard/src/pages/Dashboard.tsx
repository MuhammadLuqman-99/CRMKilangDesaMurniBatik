// ============================================
// Dashboard Page
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { tenantApi, userApi, healthApi } from '../services/api';
import type { ServiceHealth, Tenant } from '../types';

// Icons
const TenantsIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M3 21h18" />
        <path d="M5 21V7l8-4v18" />
        <path d="M19 21V11l-6-4" />
    </svg>
);

const UsersIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
);

const HealthyIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
        <polyline points="22 4 12 14.01 9 11.01" />
    </svg>
);

const AlertIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
        <line x1="12" y1="9" x2="12" y2="13" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
    </svg>
);

interface DashboardStats {
    totalTenants: number;
    activeTenants: number;
    totalUsers: number;
    servicesHealthy: number;
    servicesTotal: number;
}

const DashboardPage: React.FC = () => {
    const [stats, setStats] = useState<DashboardStats>({
        totalTenants: 0,
        activeTenants: 0,
        totalUsers: 0,
        servicesHealthy: 0,
        servicesTotal: 5,
    });
    const [recentTenants, setRecentTenants] = useState<Tenant[]>([]);
    const [services, setServices] = useState<ServiceHealth[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchDashboardData = async () => {
            setIsLoading(true);
            try {
                // Fetch tenants
                const tenantsResponse = await tenantApi.list({ page: 1, pageSize: 5 });
                if (tenantsResponse.success) {
                    setRecentTenants(tenantsResponse.data);
                    setStats(prev => ({
                        ...prev,
                        totalTenants: tenantsResponse.meta.total,
                        activeTenants: tenantsResponse.data.filter(t => t.status === 'active').length,
                    }));
                }

                // Fetch users count
                const usersResponse = await userApi.list({ page: 1, pageSize: 1 });
                if (usersResponse.success) {
                    setStats(prev => ({ ...prev, totalUsers: usersResponse.meta.total }));
                }

                // Fetch service health
                const healthResults = await healthApi.checkAllServices();
                setServices(healthResults);
                setStats(prev => ({
                    ...prev,
                    servicesHealthy: healthResults.filter(s => s.status === 'healthy').length,
                    servicesTotal: healthResults.length,
                }));
            } catch (error) {
                console.error('Failed to fetch dashboard data:', error);
            } finally {
                setIsLoading(false);
            }
        };

        fetchDashboardData();
    }, []);

    if (isLoading) {
        return (
            <div className="flex items-center justify-center" style={{ height: '60vh' }}>
                <div className="loading-spinner" style={{ width: 48, height: 48 }} />
            </div>
        );
    }

    const allServicesHealthy = stats.servicesHealthy === stats.servicesTotal;

    return (
        <div>
            {/* Stats Grid */}
            <div className="grid grid-cols-4 gap-4 mb-6">
                {/* Total Tenants */}
                <div className="stats-card">
                    <div className="stats-icon primary">
                        <TenantsIcon />
                    </div>
                    <div className="stats-content">
                        <div className="stats-label">Total Tenants</div>
                        <div className="stats-value">{stats.totalTenants}</div>
                        <div className="stats-change up">
                            <span>{stats.activeTenants} active</span>
                        </div>
                    </div>
                </div>

                {/* Total Users */}
                <div className="stats-card">
                    <div className="stats-icon success">
                        <UsersIcon />
                    </div>
                    <div className="stats-content">
                        <div className="stats-label">Total Users</div>
                        <div className="stats-value">{stats.totalUsers}</div>
                        <div className="stats-change up">
                            <span>Across all tenants</span>
                        </div>
                    </div>
                </div>

                {/* System Health */}
                <div className="stats-card">
                    <div className={`stats-icon ${allServicesHealthy ? 'success' : 'danger'}`}>
                        {allServicesHealthy ? <HealthyIcon /> : <AlertIcon />}
                    </div>
                    <div className="stats-content">
                        <div className="stats-label">System Health</div>
                        <div className="stats-value">
                            {stats.servicesHealthy}/{stats.servicesTotal}
                        </div>
                        <div className={`stats-change ${allServicesHealthy ? 'up' : 'down'}`}>
                            <span>{allServicesHealthy ? 'All services healthy' : 'Issues detected'}</span>
                        </div>
                    </div>
                </div>

                {/* Quick Actions */}
                <div className="stats-card" style={{ background: 'linear-gradient(135deg, var(--primary), var(--primary-dark))' }}>
                    <div className="stats-content" style={{ width: '100%' }}>
                        <div className="stats-label" style={{ color: 'rgba(255,255,255,0.8)' }}>Quick Actions</div>
                        <div style={{ marginTop: 'var(--space-3)', display: 'flex', gap: 'var(--space-2)' }}>
                            <Link to="/tenants" className="btn btn-secondary btn-sm">
                                View Tenants
                            </Link>
                            <Link to="/monitoring" className="btn btn-secondary btn-sm">
                                Monitoring
                            </Link>
                        </div>
                    </div>
                </div>
            </div>

            {/* Two Column Layout */}
            <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 'var(--space-6)' }}>
                {/* Recent Tenants */}
                <div className="card">
                    <div className="card-header">
                        <h3 className="card-title">Recent Tenants</h3>
                        <Link to="/tenants" className="btn btn-ghost btn-sm">
                            View All →
                        </Link>
                    </div>
                    <div className="table-container">
                        <table className="table">
                            <thead>
                                <tr>
                                    <th>Name</th>
                                    <th>Slug</th>
                                    <th>Status</th>
                                    <th>Plan</th>
                                    <th>Created</th>
                                </tr>
                            </thead>
                            <tbody>
                                {recentTenants.length === 0 ? (
                                    <tr>
                                        <td colSpan={5} style={{ textAlign: 'center', padding: 'var(--space-8)' }}>
                                            No tenants found
                                        </td>
                                    </tr>
                                ) : (
                                    recentTenants.map((tenant) => (
                                        <tr key={tenant.id}>
                                            <td className="text-primary">{tenant.name}</td>
                                            <td>{tenant.slug}</td>
                                            <td>
                                                <span className={`badge badge-${tenant.status}`}>
                                                    {tenant.status}
                                                </span>
                                            </td>
                                            <td>
                                                <span className={`badge badge-${tenant.plan}`}>
                                                    {tenant.plan}
                                                </span>
                                            </td>
                                            <td>{new Date(tenant.createdAt).toLocaleDateString()}</td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>

                {/* Service Status */}
                <div className="card">
                    <div className="card-header">
                        <h3 className="card-title">Service Status</h3>
                        <Link to="/monitoring" className="btn btn-ghost btn-sm">
                            Details →
                        </Link>
                    </div>
                    <div className="card-body" style={{ padding: 0 }}>
                        {services.map((service) => (
                            <div
                                key={service.name}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'space-between',
                                    padding: 'var(--space-3) var(--space-4)',
                                    borderBottom: '1px solid var(--border-color)',
                                }}
                            >
                                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                                    <div
                                        className={`health-status-indicator ${service.status}`}
                                        style={{ width: 8, height: 8 }}
                                    />
                                    <span style={{ fontSize: '0.875rem' }}>{service.name}</span>
                                </div>
                                <span className={`badge badge-${service.status}`}>
                                    {service.status}
                                </span>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default DashboardPage;
