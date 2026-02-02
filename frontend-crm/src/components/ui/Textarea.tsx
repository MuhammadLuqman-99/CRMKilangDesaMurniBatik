// ============================================
// Textarea Component
// Production-Ready Form Textarea
// ============================================

import { forwardRef, type TextareaHTMLAttributes } from 'react';

interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
    label?: string;
    error?: string;
    hint?: string;
    resize?: 'none' | 'vertical' | 'horizontal' | 'both';
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
    ({ label, error, hint, resize = 'vertical', className = '', id, ...props }, ref) => {
        const textareaId = id || `textarea-${Math.random().toString(36).substr(2, 9)}`;

        return (
            <div className={`form-group ${className}`}>
                {label && (
                    <label htmlFor={textareaId} className="form-label">
                        {label}
                        {props.required && <span className="text-danger"> *</span>}
                    </label>
                )}
                <textarea
                    ref={ref}
                    id={textareaId}
                    className={`form-textarea ${error ? 'error' : ''}`}
                    style={{ resize }}
                    {...props}
                />
                {error && <span className="form-error">{error}</span>}
                {hint && !error && <span className="form-hint">{hint}</span>}
            </div>
        );
    }
);

Textarea.displayName = 'Textarea';

export default Textarea;
