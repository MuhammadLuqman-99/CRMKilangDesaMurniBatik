// ============================================
// Input Component
// Production-Ready Form Input
// ============================================

import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
    label?: string;
    error?: string;
    hint?: string;
    leftIcon?: ReactNode;
    rightIcon?: ReactNode;
    inputSize?: 'sm' | 'md' | 'lg';
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
    (
        {
            label,
            error,
            hint,
            leftIcon,
            rightIcon,
            inputSize = 'md',
            className = '',
            id,
            ...props
        },
        ref
    ) => {
        const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;

        return (
            <div className={`form-group ${className}`}>
                {label && (
                    <label htmlFor={inputId} className="form-label">
                        {label}
                        {props.required && <span className="text-danger"> *</span>}
                    </label>
                )}
                <div className={`input-wrapper ${leftIcon ? 'has-left-icon' : ''} ${rightIcon ? 'has-right-icon' : ''}`}>
                    {leftIcon && <span className="input-icon left">{leftIcon}</span>}
                    <input
                        ref={ref}
                        id={inputId}
                        className={`form-input ${inputSize !== 'md' ? `form-input-${inputSize}` : ''} ${error ? 'error' : ''}`}
                        {...props}
                    />
                    {rightIcon && <span className="input-icon right">{rightIcon}</span>}
                </div>
                {error && <span className="form-error">{error}</span>}
                {hint && !error && <span className="form-hint">{hint}</span>}
            </div>
        );
    }
);

Input.displayName = 'Input';

export default Input;
