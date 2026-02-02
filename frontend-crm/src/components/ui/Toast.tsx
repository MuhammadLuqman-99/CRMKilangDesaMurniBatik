// ============================================
// Toast Component
// Production-Ready Toast Notifications
// ============================================

import { useState, useEffect, useCallback, createContext, useContext, type ReactNode } from 'react';
import { createPortal } from 'react-dom';

// Icons
const CheckIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="20 6 9 17 4 12" />
    </svg>
);

const AlertIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" /><line x1="12" y1="9" x2="12" y2="13" /><line x1="12" y1="17" x2="12.01" y2="17" />
    </svg>
);

const InfoIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" /><line x1="12" y1="16" x2="12" y2="12" /><line x1="12" y1="8" x2="12.01" y2="8" />
    </svg>
);

const CloseIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
    </svg>
);

type ToastType = 'success' | 'error' | 'warning' | 'info';

interface ToastItem {
    id: string;
    type: ToastType;
    title: string;
    message?: string;
    duration?: number;
}

interface ToastContextType {
    showToast: (toast: Omit<ToastItem, 'id'>) => void;
    success: (title: string, message?: string) => void;
    error: (title: string, message?: string) => void;
    warning: (title: string, message?: string) => void;
    info: (title: string, message?: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

interface ToastProviderProps {
    children: ReactNode;
}

export function ToastProvider({ children }: ToastProviderProps) {
    const [toasts, setToasts] = useState<ToastItem[]>([]);

    const removeToast = useCallback((id: string) => {
        setToasts((prev) => prev.filter((toast) => toast.id !== id));
    }, []);

    const showToast = useCallback((toast: Omit<ToastItem, 'id'>) => {
        const id = Math.random().toString(36).substring(2, 11);
        const newToast: ToastItem = { ...toast, id };

        setToasts((prev) => [...prev, newToast]);

        // Auto remove after duration
        const duration = toast.duration || 5000;
        setTimeout(() => {
            removeToast(id);
        }, duration);
    }, [removeToast]);

    const success = useCallback(
        (title: string, message?: string) => showToast({ type: 'success', title, message }),
        [showToast]
    );

    const error = useCallback(
        (title: string, message?: string) => showToast({ type: 'error', title, message }),
        [showToast]
    );

    const warning = useCallback(
        (title: string, message?: string) => showToast({ type: 'warning', title, message }),
        [showToast]
    );

    const info = useCallback(
        (title: string, message?: string) => showToast({ type: 'info', title, message }),
        [showToast]
    );

    const value: ToastContextType = {
        showToast,
        success,
        error,
        warning,
        info,
    };

    return (
        <ToastContext.Provider value={value}>
            {children}
            {createPortal(
                <ToastContainer toasts={toasts} onRemove={removeToast} />,
                document.body
            )}
        </ToastContext.Provider>
    );
}

interface ToastContainerProps {
    toasts: ToastItem[];
    onRemove: (id: string) => void;
}

function ToastContainer({ toasts, onRemove }: ToastContainerProps) {
    if (toasts.length === 0) {
        return null;
    }

    return (
        <div className="toast-container">
            {toasts.map((toast) => (
                <Toast key={toast.id} toast={toast} onRemove={onRemove} />
            ))}
        </div>
    );
}

interface ToastComponentProps {
    toast: ToastItem;
    onRemove: (id: string) => void;
}

function Toast({ toast, onRemove }: ToastComponentProps) {
    const [isExiting, setIsExiting] = useState(false);

    const handleRemove = useCallback(() => {
        setIsExiting(true);
        setTimeout(() => {
            onRemove(toast.id);
        }, 200);
    }, [onRemove, toast.id]);

    useEffect(() => {
        // Ensure cleanup on unmount
        return () => {
            setIsExiting(false);
        };
    }, []);

    const icons = {
        success: <CheckIcon />,
        error: <AlertIcon />,
        warning: <AlertIcon />,
        info: <InfoIcon />,
    };

    return (
        <div
            className={`toast ${toast.type} ${isExiting ? 'exiting' : ''}`}
            style={{
                opacity: isExiting ? 0 : 1,
                transform: isExiting ? 'translateX(100%)' : 'translateX(0)',
                transition: 'all 0.2s ease',
            }}
        >
            <div className="toast-icon">{icons[toast.type]}</div>
            <div className="toast-content">
                <div className="toast-title">{toast.title}</div>
                {toast.message && <div className="toast-message">{toast.message}</div>}
            </div>
            <button className="toast-close" onClick={handleRemove}>
                <CloseIcon />
            </button>
        </div>
    );
}

export function useToast(): ToastContextType {
    const context = useContext(ToastContext);
    if (context === undefined) {
        throw new Error('useToast must be used within a ToastProvider');
    }
    return context;
}

export default ToastProvider;
