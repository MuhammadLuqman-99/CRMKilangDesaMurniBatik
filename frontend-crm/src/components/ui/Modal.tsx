// ============================================
// Modal Component
// Production-Ready Modal Dialog
// ============================================

import { useEffect, useCallback, type ReactNode } from 'react';
import { createPortal } from 'react-dom';

// SVG Close Icon
const CloseIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
    </svg>
);

interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title?: string;
    size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
    showCloseButton?: boolean;
    closeOnOverlay?: boolean;
    closeOnEscape?: boolean;
    children: ReactNode;
    footer?: ReactNode;
}

export function Modal({
    isOpen,
    onClose,
    title,
    size = 'md',
    showCloseButton = true,
    closeOnOverlay = true,
    closeOnEscape = true,
    children,
    footer,
}: ModalProps) {
    // Handle escape key press
    const handleEscape = useCallback(
        (event: KeyboardEvent) => {
            if (closeOnEscape && event.key === 'Escape') {
                onClose();
            }
        },
        [closeOnEscape, onClose]
    );

    // Add/remove event listeners
    useEffect(() => {
        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
            document.body.style.overflow = 'hidden';
        }

        return () => {
            document.removeEventListener('keydown', handleEscape);
            document.body.style.overflow = '';
        };
    }, [isOpen, handleEscape]);

    // Handle overlay click
    const handleOverlayClick = () => {
        if (closeOnOverlay) {
            onClose();
        }
    };

    // Prevent click propagation from modal content
    const handleContentClick = (e: React.MouseEvent) => {
        e.stopPropagation();
    };

    if (!isOpen) {
        return null;
    }

    const sizeClass = {
        sm: 'modal-sm',
        md: '',
        lg: 'modal-lg',
        xl: 'modal-xl',
        full: 'modal-full',
    }[size];

    const modalContent = (
        <div className="modal-overlay animate-fade-in" onClick={handleOverlayClick}>
            <div
                className={`modal ${sizeClass} animate-slide-up`}
                onClick={handleContentClick}
                role="dialog"
                aria-modal="true"
                aria-labelledby={title ? 'modal-title' : undefined}
            >
                {(title || showCloseButton) && (
                    <div className="modal-header">
                        {title && (
                            <h2 id="modal-title" className="modal-title">
                                {title}
                            </h2>
                        )}
                        {showCloseButton && (
                            <button
                                type="button"
                                className="modal-close"
                                onClick={onClose}
                                aria-label="Close modal"
                            >
                                <CloseIcon />
                            </button>
                        )}
                    </div>
                )}

                <div className="modal-body">{children}</div>

                {footer && <div className="modal-footer">{footer}</div>}
            </div>
        </div>
    );

    return createPortal(modalContent, document.body);
}

export default Modal;
