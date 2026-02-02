// ============================================
// Checkbox Component
// Production-Ready Form Checkbox
// ============================================

import { forwardRef, type InputHTMLAttributes } from 'react';

interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
    label?: string;
    error?: string;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
    ({ label, error, className = '', id, ...props }, ref) => {
        const checkboxId = id || `checkbox-${Math.random().toString(36).substr(2, 9)}`;

        return (
            <div className={`form-checkbox-wrapper ${className}`}>
                <label htmlFor={checkboxId} className="form-checkbox-label">
                    <input
                        ref={ref}
                        type="checkbox"
                        id={checkboxId}
                        className={`form-checkbox ${error ? 'error' : ''}`}
                        {...props}
                    />
                    <span className="form-checkbox-custom" />
                    {label && <span className="form-checkbox-text">{label}</span>}
                </label>
                {error && <span className="form-error">{error}</span>}
            </div>
        );
    }
);

Checkbox.displayName = 'Checkbox';

export default Checkbox;
