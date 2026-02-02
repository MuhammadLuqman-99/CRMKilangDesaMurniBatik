// ============================================
// Card Component
// Production-Ready Card Container
// ============================================

import { type ReactNode } from 'react';

interface CardProps {
    children: ReactNode;
    className?: string;
    padding?: 'none' | 'sm' | 'md' | 'lg';
    hoverable?: boolean;
    onClick?: () => void;
}

interface CardHeaderProps {
    children: ReactNode;
    className?: string;
    action?: ReactNode;
}

interface CardBodyProps {
    children: ReactNode;
    className?: string;
}

interface CardFooterProps {
    children: ReactNode;
    className?: string;
}

export function Card({
    children,
    className = '',
    padding = 'md',
    hoverable = false,
    onClick,
}: CardProps) {
    const paddingClass = {
        none: '',
        sm: 'p-3',
        md: 'p-4',
        lg: 'p-6',
    }[padding];

    return (
        <div
            className={`card ${paddingClass} ${hoverable ? 'cursor-pointer transition' : ''} ${className}`}
            onClick={onClick}
            style={hoverable ? { cursor: 'pointer' } : undefined}
        >
            {children}
        </div>
    );
}

export function CardHeader({ children, className = '', action }: CardHeaderProps) {
    return (
        <div className={`card-header ${className}`}>
            <div className="card-header-content">{children}</div>
            {action && <div className="card-header-action">{action}</div>}
        </div>
    );
}

export function CardBody({ children, className = '' }: CardBodyProps) {
    return <div className={`card-body ${className}`}>{children}</div>;
}

export function CardFooter({ children, className = '' }: CardFooterProps) {
    return <div className={`card-footer ${className}`}>{children}</div>;
}

Card.Header = CardHeader;
Card.Body = CardBody;
Card.Footer = CardFooter;

export default Card;
