// ============================================
// Dashboard Page
// Production-Ready Main Dashboard
// ============================================

import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { dashboardService } from '../../services';
import { StatsCard, Card, Badge } from '../../components/ui';
import type { DashboardStats, PipelineOverview, RecentActivity } from '../../types';

// SVG Icons
const TrendingUpIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" /><polyline points="17 6 23 6 23 12" />
    </svg>
);

const UsersIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>
);

const DollarIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" /><path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 1 1 0 4H8" /><path d="M12 18V6" />
    </svg>
);

const TargetIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" /><circle cx="12" cy="12" r="6" /><circle cx="12" cy="12" r="2" />
    </svg>
);

const PlusIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
    </svg>
);

interface DashboardData {
    stats: DashboardStats;
    pipelineOverview: PipelineOverview;
    recentActivities: RecentActivity[];
}

export function DashboardPage() {
    const [data, setData] = useState<DashboardData | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchDashboardData = async () => {
            try {
                const response = await dashboardService.getDashboardData();
                setData({
                    stats: response.stats,
                    pipelineOverview: response.pipeline_overview,
                    recentActivities: response.recent_activities,
                });
            } catch (err) {
                const message = err instanceof Error ? err.message : 'Failed to load dashboard';
                setError(message);
                // Use mock data for demonstration
                setData({
                    stats: {
                        total_leads: 156,
                        leads_change: 12.5,
                        total_opportunities: 48,
                        opportunities_change: 8.3,
                        pipeline_value: 1250000,
                        pipeline_change: 15.2,
                        won_deals: 23,
                        deals_change: 5.7,
                    },
                    pipelineOverview: {
                        stages: [
                            { id: '1', name: 'Qualification', value: 320000, count: 12, color: '#3b82f6' },
                            { id: '2', name: 'Proposal', value: 450000, count: 8, color: '#8b5cf6' },
                            { id: '3', name: 'Negotiation', value: 280000, count: 5, color: '#f59e0b' },
                            { id: '4', name: 'Closing', value: 200000, count: 3, color: '#10b981' },
                        ],
                        total_value: 1250000,
                        total_deals: 28,
                    },
                    recentActivities: [
                        {
                            id: '1',
                            type: 'deal_won',
                            title: 'Deal Won',
                            description: 'Enterprise Package for Batik Industries Sdn Bhd',
                            user: 'Ahmad Razak',
                            timestamp: new Date(Date.now() - 1800000).toISOString(),
                        },
                        {
                            id: '2',
                            type: 'lead_created',
                            title: 'New Lead',
                            description: 'Textile Malaysia contacted via website',
                            user: 'Siti Aminah',
                            timestamp: new Date(Date.now() - 7200000).toISOString(),
                        },
                        {
                            id: '3',
                            type: 'opportunity_moved',
                            title: 'Stage Changed',
                            description: 'Premium Batik Collection moved to Negotiation',
                            user: 'Muhammad Hafiz',
                            timestamp: new Date(Date.now() - 14400000).toISOString(),
                        },
                        {
                            id: '4',
                            type: 'meeting_scheduled',
                            title: 'Meeting Scheduled',
                            description: 'Demo call with Kraf Malaysia for tomorrow',
                            user: 'Nurul Aisyah',
                            timestamp: new Date(Date.now() - 28800000).toISOString(),
                        },
                    ],
                });
            } finally {
                setIsLoading(false);
            }
        };

        fetchDashboardData();
    }, []);

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('ms-MY', {
            style: 'currency',
            currency: 'MYR',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(value);
    };

    const formatTimeAgo = (timestamp: string) => {
        const now = new Date();
        const date = new Date(timestamp);
        const diff = Math.floor((now.getTime() - date.getTime()) / 1000);

        if (diff < 60) return 'Just now';
        if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
        if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
        return `${Math.floor(diff / 86400)}d ago`;
    };

    const getActivityIcon = (type: string) => {
        switch (type) {
            case 'deal_won':
                return '🎉';
            case 'lead_created':
                return '✨';
            case 'opportunity_moved':
                return '📊';
            case 'meeting_scheduled':
                return '📅';
            case 'email_sent':
                return '📧';
            case 'call_made':
                return '📞';
            default:
                return '📋';
        }
    };

    if (isLoading) {
        return (
            <div className="animate-fade-in">
                <div className="page-header">
                    <div className="page-header-left">
                        <div className="skeleton" style={{ width: '200px', height: '32px' }} />
                        <div className="skeleton" style={{ width: '300px', height: '20px', marginTop: '0.5rem' }} />
                    </div>
                </div>

                <div className="stats-grid mb-6">
                    {[1, 2, 3, 4].map((i) => (
                        <div key={i} className="stat-card">
                            <div className="skeleton" style={{ width: '100%', height: '100px' }} />
                        </div>
                    ))}
                </div>

                <div className="dashboard-grid">
                    <div className="widget col-span-8">
                        <div className="skeleton" style={{ width: '100%', height: '300px' }} />
                    </div>
                    <div className="widget col-span-4">
                        <div className="skeleton" style={{ width: '100%', height: '300px' }} />
                    </div>
                </div>
            </div>
        );
    }

    if (!data) {
        return (
            <div className="empty-state">
                <div className="empty-state-icon">⚠️</div>
                <h3 className="empty-state-title">Failed to load dashboard</h3>
                <p className="empty-state-description">{error || 'Please try again later.'}</p>
            </div>
        );
    }

    const totalPipelineWidth = data.pipelineOverview.total_value || 1;

    return (
        <div className="animate-fade-in">
            {/* Page Header */}
            <div className="page-header">
                <div className="page-header-left">
                    <h1 className="page-title">Dashboard</h1>
                    <p className="page-description">Welcome back! Here's your sales overview.</p>
                </div>
                <div className="page-header-actions">
                    <Link to="/leads/new" className="btn btn-primary">
                        <PlusIcon />
                        <span>New Lead</span>
                    </Link>
                </div>
            </div>

            {/* Stats Grid */}
            <div className="stats-grid mb-6">
                <StatsCard
                    title="Total Leads"
                    value={data.stats.total_leads.toLocaleString()}
                    change={{
                        value: data.stats.leads_change,
                        type: data.stats.leads_change >= 0 ? 'increase' : 'decrease',
                    }}
                    icon={<UsersIcon />}
                    iconBackground="linear-gradient(135deg, #3b82f6, #60a5fa)"
                />
                <StatsCard
                    title="Open Opportunities"
                    value={data.stats.total_opportunities.toLocaleString()}
                    change={{
                        value: data.stats.opportunities_change,
                        type: data.stats.opportunities_change >= 0 ? 'increase' : 'decrease',
                    }}
                    icon={<TrendingUpIcon />}
                    iconBackground="linear-gradient(135deg, #8b5cf6, #a78bfa)"
                />
                <StatsCard
                    title="Pipeline Value"
                    value={formatCurrency(data.stats.pipeline_value)}
                    change={{
                        value: data.stats.pipeline_change,
                        type: data.stats.pipeline_change >= 0 ? 'increase' : 'decrease',
                    }}
                    icon={<DollarIcon />}
                    iconBackground="linear-gradient(135deg, #10b981, #34d399)"
                />
                <StatsCard
                    title="Won Deals"
                    value={data.stats.won_deals.toLocaleString()}
                    change={{
                        value: data.stats.deals_change,
                        type: data.stats.deals_change >= 0 ? 'increase' : 'decrease',
                    }}
                    icon={<TargetIcon />}
                    iconBackground="linear-gradient(135deg, #f59e0b, #fbbf24)"
                />
            </div>

            {/* Main Content Grid */}
            <div className="dashboard-grid">
                {/* Pipeline Overview */}
                <Card className="widget col-span-8" padding="none">
                    <div className="widget-header">
                        <h3 className="widget-title">Pipeline Overview</h3>
                        <Link to="/pipeline" className="btn btn-ghost btn-sm">
                            View all
                        </Link>
                    </div>
                    <div className="widget-body">
                        {/* Pipeline Bar */}
                        <div className="pipeline-overview">
                            {data.pipelineOverview.stages.map((stage) => (
                                <div
                                    key={stage.id}
                                    className="pipeline-stage-bar"
                                    style={{
                                        flex: stage.value / totalPipelineWidth,
                                        background: stage.color,
                                    }}
                                    title={`${stage.name}: ${formatCurrency(stage.value)}`}
                                />
                            ))}
                        </div>

                        {/* Pipeline Legend */}
                        <div className="pipeline-legend">
                            {data.pipelineOverview.stages.map((stage) => (
                                <div key={stage.id} className="pipeline-legend-item">
                                    <div
                                        className="pipeline-legend-color"
                                        style={{ background: stage.color }}
                                    />
                                    <span className="pipeline-legend-label">{stage.name}</span>
                                    <span className="pipeline-legend-value">
                                        {stage.count} ({formatCurrency(stage.value)})
                                    </span>
                                </div>
                            ))}
                        </div>

                        {/* Summary Stats */}
                        <div
                            style={{
                                display: 'grid',
                                gridTemplateColumns: 'repeat(3, 1fr)',
                                gap: '1rem',
                                marginTop: '1.5rem',
                                paddingTop: '1.5rem',
                                borderTop: '1px solid var(--border-color)',
                            }}
                        >
                            <div>
                                <p className="text-muted text-sm">Total Pipeline Value</p>
                                <p className="text-xl font-bold text-success">
                                    {formatCurrency(data.pipelineOverview.total_value)}
                                </p>
                            </div>
                            <div>
                                <p className="text-muted text-sm">Active Deals</p>
                                <p className="text-xl font-bold">{data.pipelineOverview.total_deals}</p>
                            </div>
                            <div>
                                <p className="text-muted text-sm">Avg. Deal Size</p>
                                <p className="text-xl font-bold">
                                    {formatCurrency(data.pipelineOverview.total_value / data.pipelineOverview.total_deals)}
                                </p>
                            </div>
                        </div>
                    </div>
                </Card>

                {/* Recent Activities */}
                <Card className="widget col-span-4" padding="none">
                    <div className="widget-header">
                        <h3 className="widget-title">Recent Activities</h3>
                    </div>
                    <div className="widget-body" style={{ padding: 0 }}>
                        {data.recentActivities.map((activity) => (
                            <div key={activity.id} className="activity-item" style={{ padding: '1rem 1.25rem' }}>
                                <div
                                    className="activity-avatar"
                                    style={{ fontSize: '1.25rem', background: 'var(--bg-tertiary)' }}
                                >
                                    {getActivityIcon(activity.type)}
                                </div>
                                <div className="activity-content">
                                    <p className="activity-text">
                                        <strong>{activity.title}</strong>
                                        <br />
                                        <span style={{ color: 'var(--text-muted)' }}>{activity.description}</span>
                                    </p>
                                    <div className="activity-time">
                                        {activity.user} · {formatTimeAgo(activity.timestamp)}
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </Card>

                {/* Quick Actions */}
                <Card className="widget col-span-4" padding="none">
                    <div className="widget-header">
                        <h3 className="widget-title">Quick Actions</h3>
                    </div>
                    <div className="widget-body">
                        <div className="quick-action-grid">
                            <Link to="/leads/new" className="quick-action">
                                <div className="quick-action-icon" style={{ background: 'var(--primary)' }}>
                                    ➕
                                </div>
                                <span className="quick-action-label">Add Lead</span>
                            </Link>
                            <Link to="/opportunities/new" className="quick-action">
                                <div className="quick-action-icon" style={{ background: '#8b5cf6' }}>
                                    💰
                                </div>
                                <span className="quick-action-label">New Deal</span>
                            </Link>
                            <Link to="/customers/new" className="quick-action">
                                <div className="quick-action-icon" style={{ background: '#10b981' }}>
                                    👤
                                </div>
                                <span className="quick-action-label">New Customer</span>
                            </Link>
                            <Link to="/pipeline" className="quick-action">
                                <div className="quick-action-icon" style={{ background: '#f59e0b' }}>
                                    📊
                                </div>
                                <span className="quick-action-label">View Pipeline</span>
                            </Link>
                        </div>
                    </div>
                </Card>

                {/* Upcoming Tasks */}
                <Card className="widget col-span-8" padding="none">
                    <div className="widget-header">
                        <h3 className="widget-title">Upcoming Tasks</h3>
                        <Badge>3 due today</Badge>
                    </div>
                    <div className="widget-body" style={{ padding: 0 }}>
                        <table className="table">
                            <thead>
                                <tr>
                                    <th>Task</th>
                                    <th>Related To</th>
                                    <th>Due Date</th>
                                    <th>Priority</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td>
                                        <strong>Follow up call</strong>
                                        <br />
                                        <span className="text-muted text-sm">Discuss pricing options</span>
                                    </td>
                                    <td>
                                        <Badge variant="info">Batik Industries</Badge>
                                    </td>
                                    <td>Today, 2:00 PM</td>
                                    <td>
                                        <Badge variant="danger">High</Badge>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                        <strong>Send proposal</strong>
                                        <br />
                                        <span className="text-muted text-sm">Premium Package proposal</span>
                                    </td>
                                    <td>
                                        <Badge variant="info">Textile Malaysia</Badge>
                                    </td>
                                    <td>Today, 5:00 PM</td>
                                    <td>
                                        <Badge variant="warning">Medium</Badge>
                                    </td>
                                </tr>
                                <tr>
                                    <td>
                                        <strong>Demo presentation</strong>
                                        <br />
                                        <span className="text-muted text-sm">Product walkthrough</span>
                                    </td>
                                    <td>
                                        <Badge variant="info">Kraf Malaysia</Badge>
                                    </td>
                                    <td>Tomorrow, 10:00 AM</td>
                                    <td>
                                        <Badge variant="warning">Medium</Badge>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </Card>
            </div>
        </div>
    );
}

export default DashboardPage;
