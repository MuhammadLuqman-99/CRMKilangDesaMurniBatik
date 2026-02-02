// ============================================
// Select Component
// Production-Ready Form Select
// ============================================

import { forwardRef, type SelectHTMLAttributes, type ReactNode } from 'react';

interface SelectOption {
    value: string;
    label: string;
    disabled?: boolean;
}

interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
    label?: string;
    error?: string;
    hint?: string;
    options: SelectOption[];
    placeholder?: string;
    selectSize?: 'sm' | 'md' | 'lg';
    leftIcon?: ReactNode;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
    (
        {
            label,
            error,
            hint,
            options,
            placeholder,
            selectSize = 'md',
            leftIcon,
            className = '',
            id,
            ...props
        },
        ref
    ) => {
        const selectId = id || `select-${Math.random().toString(36).substr(2, 9)}`;

        return (
            <div className={`form-group ${className}`}>
                {label && (
                    <label htmlFor={selectId} className="form-label">
                        {label}
                        {props.required && <span className="text-danger"> *</span>}
                    </label>
                )}
                <div className={`input-wrapper ${leftIcon ? 'has-left-icon' : ''}`}>
                    {leftIcon && <span className="input-icon left">{leftIcon}</span>}
                    <select
                        ref={ref}
                        id={selectId}
                        className={`form-select ${selectSize !== 'md' ? `form-select-${selectSize}` : ''} ${error ? 'error' : ''}`}
                        {...props}
                    >
                        {placeholder && (
                            <option value="" disabled>
                                {placeholder}
                            </option>
                        )}
                        {options.map((option) => (
                            <option key={option.value} value={option.value} disabled={option.disabled}>
                                {option.label}
                            </option>
                        ))}
                    </select>
                </div>
                {error && <span className="form-error">{error}</span>}
                {hint && !error && <span className="form-hint">{hint}</span>}
            </div>
        );
    }
);

Select.displayName = 'Select';

export default Select;
