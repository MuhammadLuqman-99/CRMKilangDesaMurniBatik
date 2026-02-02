// ============================================
// Stats Card Component
// Production-Ready Dashboard Stats Card
// ============================================

import { type ReactNode } from 'react';

interface StatsCardProps {
    title: string;
    value: string | number;
    change?: {
        value: number;
        type: 'increase' | 'decrease';
        label?: string;
    };
    icon?: ReactNode;
    iconBackground?: string;
    subtitle?: string;
}

export function StatsCard({
    title,
    value,
    change,
    icon,
    iconBackground = 'var(--primary)',
    subtitle,
}: StatsCardProps) {
    return (
        <div className="stat-card">
            <div className="stat-card-header">
                <div className="stat-card-info">
                    <span className="stat-card-title">{title}</span>
                    <span className="stat-card-value">{value}</span>
                    {subtitle && <span className="text-sm text-muted">{subtitle}</span>}
                </div>
                {icon && (
                    <div className="stat-card-icon" style={{ background: iconBackground }}>
                        {icon}
                    </div>
                )}
            </div>
            {change && (
                <div className="stat-card-footer">
                    <span className={`stat-card-change ${change.type === 'increase' ? 'positive' : 'negative'}`}>
                        {change.type === 'increase' ? '↑' : '↓'} {Math.abs(change.value)}%
                    </span>
                    <span className="stat-card-period">{change.label || 'vs last month'}</span>
                </div>
            )}
        </div>
    );
}

export default StatsCard;
