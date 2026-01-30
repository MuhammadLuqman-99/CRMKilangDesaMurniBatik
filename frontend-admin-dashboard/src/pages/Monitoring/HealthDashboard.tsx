// ============================================
// System Monitoring Page
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React, { useState, useEffect, useCallback } from 'react';
import { healthApi } from '../../services/api';
import type { ServiceHealth, DatabaseStatus, QueueStatus } from '../../types';

// ============================================
// Icons
// ============================================

const RefreshIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
        <path d="M3 3v5h5" />
        <path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16" />
        <path d="M16 16h5v5" />
    </svg>
);

const ServerIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="2" y="2" width="20" height="8" rx="2" ry="2" />
        <rect x="2" y="14" width="20" height="8" rx="2" ry="2" />
        <line x1="6" y1="6" x2="6.01" y2="6" />
        <line x1="6" y1="18" x2="6.01" y2="18" />
    </svg>
);

const DatabaseIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <ellipse cx="12" cy="5" rx="9" ry="3" />
        <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
        <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
    </svg>
);

const QueueIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <line x1="8" y1="6" x2="21" y2="6" />
        <line x1="8" y1="12" x2="21" y2="12" />
        <line x1="8" y1="18" x2="21" y2="18" />
        <line x1="3" y1="6" x2="3.01" y2="6" />
        <line x1="3" y1="12" x2="3.01" y2="12" />
        <line x1="3" y1="18" x2="3.01" y2="18" />
    </svg>
);

// ============================================
// Service Health Card Component
// ============================================

interface ServiceHealthCardProps {
    service: ServiceHealth;
}

const ServiceHealthCard: React.FC<ServiceHealthCardProps> = ({ service }) => {
    return (
        <div className={`health-card ${service.status}`}>
            <div className="health-card-header">
                <h3 className="health-service-name">{service.name}</h3>
                <div className={`health-status-indicator ${service.status}`} />
            </div>
            <div className="health-details">
                <div className="health-detail-row">
                    <span className="health-detail-label">Status</span>
                    <span className={`badge badge-${service.status}`}>{service.status}</span>
                </div>
                <div className="health-detail-row">
                    <span className="health-detail-label">Port</span>
                    <span className="health-detail-value">{service.port}</span>
                </div>
                {service.responseTime !== undefined && (
                    <div className="health-detail-row">
                        <span className="health-detail-label">Response Time</span>
                        <span className="health-detail-value">{service.responseTime}ms</span>
                    </div>
                )}
                <div className="health-detail-row">
                    <span className="health-detail-label">Last Checked</span>
                    <span className="health-detail-value">
                        {new Date(service.lastChecked).toLocaleTimeString()}
                    </span>
                </div>
            </div>
        </div>
    );
};

// ============================================
// Database Status Card Component
// ============================================

interface DatabaseStatusCardProps {
    database: DatabaseStatus;
}

const DatabaseStatusCard: React.FC<DatabaseStatusCardProps> = ({ database }) => {
    const getTypeColor = (type: string) => {
        switch (type) {
            case 'postgresql': return 'var(--primary)';
            case 'mongodb': return 'var(--success)';
            case 'redis': return 'var(--danger)';
            default: return 'var(--text-muted)';
        }
    };

    return (
        <div className={`health-card ${database.status}`}>
            <div className="health-card-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                    <div style={{
                        width: 32,
                        height: 32,
                        borderRadius: 'var(--radius-md)',
                        background: `${getTypeColor(database.type)}20`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                    }}>
                        <DatabaseIcon />
                    </div>
                    <h3 className="health-service-name">{database.name}</h3>
                </div>
                <div className={`health-status-indicator ${database.status}`} />
            </div>
            <div className="health-details">
                <div className="health-detail-row">
                    <span className="health-detail-label">Type</span>
                    <span className="health-detail-value" style={{ textTransform: 'capitalize' }}>
                        {database.type}
                    </span>
                </div>
                {database.connectionPool && (
                    <>
                        <div className="health-detail-row">
                            <span className="health-detail-label">Active Connections</span>
                            <span className="health-detail-value">{database.connectionPool.active}</span>
                        </div>
                        <div className="health-detail-row">
                            <span className="health-detail-label">Idle Connections</span>
                            <span className="health-detail-value">{database.connectionPool.idle}</span>
                        </div>
                        <div className="health-detail-row">
                            <span className="health-detail-label">Total Pool</span>
                            <span className="health-detail-value">{database.connectionPool.total}</span>
                        </div>
                    </>
                )}
                {database.latency !== undefined && (
                    <div className="health-detail-row">
                        <span className="health-detail-label">Latency</span>
                        <span className="health-detail-value">{database.latency}ms</span>
                    </div>
                )}
            </div>

            {/* Connection Pool Progress */}
            {database.connectionPool && (
                <div style={{ marginTop: 'var(--space-3)' }}>
                    <div style={{
                        height: 6,
                        background: 'var(--bg-tertiary)',
                        borderRadius: 'var(--radius-full)',
                        overflow: 'hidden',
                    }}>
                        <div style={{
                            height: '100%',
                            width: `${(database.connectionPool.active / database.connectionPool.total) * 100}%`,
                            background: 'var(--primary)',
                            borderRadius: 'var(--radius-full)',
                            transition: 'width 0.3s ease',
                        }} />
                    </div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: 'var(--space-1)' }}>
                        {Math.round((database.connectionPool.active / database.connectionPool.total) * 100)}% pool utilization
                    </div>
                </div>
            )}
        </div>
    );
};

// ============================================
// Queue Status Card Component
// ============================================

interface QueueStatusCardProps {
    queue: QueueStatus;
}

const QueueStatusCard: React.FC<QueueStatusCardProps> = ({ queue }) => {
    return (
        <div className={`health-card ${queue.status}`}>
            <div className="health-card-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                    <div style={{
                        width: 32,
                        height: 32,
                        borderRadius: 'var(--radius-md)',
                        background: 'var(--warning-bg)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: 'var(--warning)',
                    }}>
                        <QueueIcon />
                    </div>
                    <h3 className="health-service-name">{queue.name}</h3>
                </div>
                <div className={`health-status-indicator ${queue.status}`} />
            </div>
            <div className="health-details">
                <div className="health-detail-row">
                    <span className="health-detail-label">Messages</span>
                    <span className="health-detail-value" style={{
                        color: queue.messageCount > 100 ? 'var(--warning)' : 'var(--text-primary)',
                        fontWeight: 600,
                    }}>
                        {queue.messageCount}
                    </span>
                </div>
                <div className="health-detail-row">
                    <span className="health-detail-label">Consumers</span>
                    <span className="health-detail-value">{queue.consumerCount}</span>
                </div>
                {queue.publishRate !== undefined && (
                    <div className="health-detail-row">
                        <span className="health-detail-label">Publish Rate</span>
                        <span className="health-detail-value">{queue.publishRate}/s</span>
                    </div>
                )}
                {queue.consumeRate !== undefined && (
                    <div className="health-detail-row">
                        <span className="health-detail-label">Consume Rate</span>
                        <span className="health-detail-value">{queue.consumeRate}/s</span>
                    </div>
                )}
            </div>

            {/* Throughput indicator */}
            {queue.publishRate !== undefined && queue.consumeRate !== undefined && (
                <div style={{ marginTop: 'var(--space-3)', display: 'flex', gap: 'var(--space-2)', alignItems: 'center' }}>
                    <span style={{ fontSize: '0.75rem', color: 'var(--success)' }}>↑ {queue.publishRate}/s</span>
                    <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>|</span>
                    <span style={{ fontSize: '0.75rem', color: 'var(--primary)' }}>↓ {queue.consumeRate}/s</span>
                </div>
            )}
        </div>
    );
};

// ============================================
// Main Component
// ============================================

const HealthDashboard: React.FC = () => {
    const [services, setServices] = useState<ServiceHealth[]>([]);
    const [databases, setDatabases] = useState<DatabaseStatus[]>([]);
    const [queues, setQueues] = useState<QueueStatus[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [lastUpdated, setLastUpdated] = useState<Date>(new Date());
    const [isAutoRefresh, setIsAutoRefresh] = useState(true);

    const fetchHealthData = useCallback(async () => {
        setIsLoading(true);
        try {
            const [servicesData, databasesData, queuesData] = await Promise.all([
                healthApi.checkAllServices(),
                healthApi.getDatabaseStatus(),
                healthApi.getQueueStatus(),
            ]);

            setServices(servicesData);
            setDatabases(databasesData);
            setQueues(queuesData);
            setLastUpdated(new Date());
        } catch (error) {
            console.error('Failed to fetch health data:', error);
        } finally {
            setIsLoading(false);
        }
    }, []);

    // Initial fetch
    useEffect(() => {
        fetchHealthData();
    }, [fetchHealthData]);

    // Auto-refresh every 30 seconds
    useEffect(() => {
        if (!isAutoRefresh) return;

        const interval = setInterval(fetchHealthData, 30000);
        return () => clearInterval(interval);
    }, [isAutoRefresh, fetchHealthData]);

    const healthyServicesCount = services.filter(s => s.status === 'healthy').length;
    const healthyDatabasesCount = databases.filter(d => d.status === 'healthy').length;
    const healthyQueuesCount = queues.filter(q => q.status === 'healthy').length;

    const overallHealth =
        healthyServicesCount === services.length &&
        healthyDatabasesCount === databases.length &&
        healthyQueuesCount === queues.length;

    return (
        <div>
            {/* Page Header */}
            <div className="page-header">
                <div>
                    <h1 className="page-title">System Monitoring</h1>
                    <p className="page-description">
                        Real-time health status of all services, databases, and message queues
                    </p>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                    <label style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                        <input
                            type="checkbox"
                            checked={isAutoRefresh}
                            onChange={e => setIsAutoRefresh(e.target.checked)}
                            style={{ accentColor: 'var(--primary)' }}
                        />
                        Auto-refresh
                    </label>
                    <button
                        className="btn btn-secondary"
                        onClick={fetchHealthData}
                        disabled={isLoading}
                    >
                        <RefreshIcon />
                        {isLoading ? 'Refreshing...' : 'Refresh'}
                    </button>
                </div>
            </div>

            {/* Overall Status Banner */}
            <div style={{
                padding: 'var(--space-4) var(--space-5)',
                background: overallHealth ? 'var(--success-bg)' : 'var(--warning-bg)',
                border: `1px solid ${overallHealth ? 'var(--success-border)' : 'var(--warning-border)'}`,
                borderRadius: 'var(--radius-xl)',
                marginBottom: 'var(--space-6)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
            }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                    <div className={`health-status-indicator ${overallHealth ? 'healthy' : 'degraded'}`} />
                    <div>
                        <div style={{ fontWeight: 600, color: overallHealth ? 'var(--success)' : 'var(--warning)' }}>
                            {overallHealth ? 'All Systems Operational' : 'Some Systems Need Attention'}
                        </div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                            Last updated: {lastUpdated.toLocaleTimeString()}
                        </div>
                    </div>
                </div>
                <div style={{ display: 'flex', gap: 'var(--space-4)' }}>
                    <div style={{ textAlign: 'center' }}>
                        <div style={{ fontSize: '1.25rem', fontWeight: 700 }}>{healthyServicesCount}/{services.length}</div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Services</div>
                    </div>
                    <div style={{ textAlign: 'center' }}>
                        <div style={{ fontSize: '1.25rem', fontWeight: 700 }}>{healthyDatabasesCount}/{databases.length}</div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Databases</div>
                    </div>
                    <div style={{ textAlign: 'center' }}>
                        <div style={{ fontSize: '1.25rem', fontWeight: 700 }}>{healthyQueuesCount}/{queues.length}</div>
                        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Queues</div>
                    </div>
                </div>
            </div>

            {/* API Services Section */}
            <div style={{ marginBottom: 'var(--space-8)' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginBottom: 'var(--space-4)' }}>
                    <ServerIcon />
                    <h2 style={{ fontSize: '1.125rem', fontWeight: 600 }}>API Services</h2>
                    <span className={`badge ${healthyServicesCount === services.length ? 'badge-healthy' : 'badge-degraded'}`}>
                        {healthyServicesCount}/{services.length} healthy
                    </span>
                </div>
                {isLoading && services.length === 0 ? (
                    <div className="flex items-center justify-center" style={{ padding: 'var(--space-8)' }}>
                        <div className="loading-spinner" />
                    </div>
                ) : (
                    <div className="health-grid">
                        {services.map(service => (
                            <ServiceHealthCard key={service.name} service={service} />
                        ))}
                    </div>
                )}
            </div>

            {/* Database Section */}
            <div style={{ marginBottom: 'var(--space-8)' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginBottom: 'var(--space-4)' }}>
                    <DatabaseIcon />
                    <h2 style={{ fontSize: '1.125rem', fontWeight: 600 }}>Database Connections</h2>
                    <span className={`badge ${healthyDatabasesCount === databases.length ? 'badge-healthy' : 'badge-degraded'}`}>
                        {healthyDatabasesCount}/{databases.length} healthy
                    </span>
                </div>
                <div className="health-grid">
                    {databases.map(database => (
                        <DatabaseStatusCard key={database.name} database={database} />
                    ))}
                </div>
            </div>

            {/* Queue Section */}
            <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)', marginBottom: 'var(--space-4)' }}>
                    <QueueIcon />
                    <h2 style={{ fontSize: '1.125rem', fontWeight: 600 }}>Message Queues (RabbitMQ)</h2>
                    <span className={`badge ${healthyQueuesCount === queues.length ? 'badge-healthy' : 'badge-degraded'}`}>
                        {healthyQueuesCount}/{queues.length} healthy
                    </span>
                </div>
                <div className="health-grid">
                    {queues.map(queue => (
                        <QueueStatusCard key={queue.name} queue={queue} />
                    ))}
                </div>
            </div>
        </div>
    );
};

export default HealthDashboard;
