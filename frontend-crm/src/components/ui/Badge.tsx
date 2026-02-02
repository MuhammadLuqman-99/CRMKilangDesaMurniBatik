// ============================================
// Badge Component
// Production-Ready Status Badge
// ============================================

import { type ReactNode } from 'react';

interface BadgeProps {
    variant?: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'purple';
    size?: 'sm' | 'md';
    children: ReactNode;
    className?: string;
}

export function Badge({ variant = 'default', size = 'md', children, className = '' }: BadgeProps) {
    const variantClass = {
        default: 'badge',
        success: 'badge badge-success',
        warning: 'badge badge-warning',
        danger: 'badge badge-danger',
        info: 'badge badge-info',
        purple: 'badge badge-purple',
    }[variant];

    return (
        <span className={`${variantClass} ${size === 'sm' ? 'badge-sm' : ''} ${className}`}>
            {children}
        </span>
    );
}

export default Badge;
