// ============================================
// Register Page
// Production-Ready Registration Screen
// ============================================

import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { Button, Input, Checkbox } from '../../components/ui';

// SVG Icons
const EyeIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
    </svg>
);

const EyeOffIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" /><line x1="1" y1="1" x2="23" y2="23" />
    </svg>
);

const CheckIcon = () => (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="20 6 9 17 4 12" />
    </svg>
);

interface PasswordStrength {
    score: number;
    label: string;
    color: string;
}

function getPasswordStrength(password: string): PasswordStrength {
    let score = 0;

    if (password.length >= 8) score++;
    if (password.length >= 12) score++;
    if (/[a-z]/.test(password)) score++;
    if (/[A-Z]/.test(password)) score++;
    if (/[0-9]/.test(password)) score++;
    if (/[^a-zA-Z0-9]/.test(password)) score++;

    if (score <= 2) return { score, label: 'Weak', color: 'var(--danger)' };
    if (score <= 4) return { score, label: 'Fair', color: 'var(--warning)' };
    if (score <= 5) return { score, label: 'Good', color: 'var(--info)' };
    return { score, label: 'Strong', color: 'var(--success)' };
}

export function RegisterPage() {
    const navigate = useNavigate();
    const { register, error, isLoading, clearError } = useAuth();

    const [formData, setFormData] = useState({
        first_name: '',
        last_name: '',
        email: '',
        password: '',
        confirm_password: '',
    });
    const [showPassword, setShowPassword] = useState(false);
    const [acceptTerms, setAcceptTerms] = useState(false);
    const [success, setSuccess] = useState(false);
    const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

    const passwordStrength = getPasswordStrength(formData.password);

    const passwordChecks = [
        { label: 'At least 8 characters', valid: formData.password.length >= 8 },
        { label: 'Contains lowercase letter', valid: /[a-z]/.test(formData.password) },
        { label: 'Contains uppercase letter', valid: /[A-Z]/.test(formData.password) },
        { label: 'Contains a number', valid: /[0-9]/.test(formData.password) },
        { label: 'Contains special character', valid: /[^a-zA-Z0-9]/.test(formData.password) },
    ];

    const handleChange = (field: string, value: string) => {
        setFormData((prev) => ({ ...prev, [field]: value }));
        if (validationErrors[field]) {
            setValidationErrors((prev) => ({ ...prev, [field]: '' }));
        }
    };

    const validate = (): boolean => {
        const errors: Record<string, string> = {};

        if (!formData.first_name.trim()) {
            errors.first_name = 'First name is required';
        }
        if (!formData.last_name.trim()) {
            errors.last_name = 'Last name is required';
        }
        if (!formData.email.trim()) {
            errors.email = 'Email is required';
        } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
            errors.email = 'Please enter a valid email address';
        }
        if (!formData.password) {
            errors.password = 'Password is required';
        } else if (formData.password.length < 8) {
            errors.password = 'Password must be at least 8 characters';
        }
        if (formData.password !== formData.confirm_password) {
            errors.confirm_password = 'Passwords do not match';
        }
        if (!acceptTerms) {
            errors.terms = 'You must accept the terms and conditions';
        }

        setValidationErrors(errors);
        return Object.keys(errors).length === 0;
    };

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        clearError();

        if (!validate()) return;

        try {
            await register({
                first_name: formData.first_name,
                last_name: formData.last_name,
                email: formData.email,
                password: formData.password,
            });
            setSuccess(true);
        } catch {
            // Error is handled by AuthContext
        }
    };

    if (success) {
        return (
            <div className="auth-page">
                <div className="auth-left">
                    <div className="auth-container">
                        <div className="auth-card text-center">
                            <div className="auth-header">
                                <div
                                    className="auth-logo"
                                    style={{ background: 'var(--success)', margin: '0 auto 1rem' }}
                                >
                                    ✓
                                </div>
                                <h1 className="auth-title">Check your email</h1>
                                <p className="auth-subtitle">
                                    We've sent a verification link to <strong>{formData.email}</strong>.
                                    Please check your inbox and click the link to verify your account.
                                </p>
                            </div>

                            <Button
                                fullWidth
                                size="lg"
                                onClick={() => navigate('/login')}
                            >
                                Go to Login
                            </Button>

                            <p className="auth-footer" style={{ marginTop: '1.5rem' }}>
                                Didn't receive the email?{' '}
                                <button
                                    type="button"
                                    style={{ background: 'none', border: 'none', color: 'var(--primary)', cursor: 'pointer' }}
                                >
                                    Resend verification
                                </button>
                            </p>
                        </div>
                    </div>
                </div>
                <div className="auth-right">
                    <div className="auth-branding">
                        <h2 className="auth-branding-title">Welcome to CRM Kilang Desa Murni Batik</h2>
                        <p className="auth-branding-subtitle">
                            You're just one step away from managing your business more effectively.
                        </p>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="auth-page">
            <div className="auth-left">
                <div className="auth-container">
                    <div className="auth-card">
                        <div className="auth-header">
                            <div className="auth-logo">CRM</div>
                            <h1 className="auth-title">Create an account</h1>
                            <p className="auth-subtitle">Start your 14-day free trial, no credit card required</p>
                        </div>

                        {error && (
                            <div className="auth-error">
                                <span>⚠️</span>
                                <span>{error}</span>
                            </div>
                        )}

                        <form onSubmit={handleSubmit}>
                            <div className="grid grid-cols-2 gap-4">
                                <Input
                                    label="First name"
                                    placeholder="John"
                                    value={formData.first_name}
                                    onChange={(e) => handleChange('first_name', e.target.value)}
                                    error={validationErrors.first_name}
                                    required
                                />
                                <Input
                                    label="Last name"
                                    placeholder="Doe"
                                    value={formData.last_name}
                                    onChange={(e) => handleChange('last_name', e.target.value)}
                                    error={validationErrors.last_name}
                                    required
                                />
                            </div>

                            <div style={{ marginTop: '1rem' }}>
                                <Input
                                    type="email"
                                    label="Email address"
                                    placeholder="john@company.com"
                                    value={formData.email}
                                    onChange={(e) => handleChange('email', e.target.value)}
                                    error={validationErrors.email}
                                    required
                                    autoComplete="email"
                                />
                            </div>

                            <div style={{ marginTop: '1rem' }}>
                                <Input
                                    type={showPassword ? 'text' : 'password'}
                                    label="Password"
                                    placeholder="Create a password"
                                    value={formData.password}
                                    onChange={(e) => handleChange('password', e.target.value)}
                                    error={validationErrors.password}
                                    required
                                    autoComplete="new-password"
                                    rightIcon={
                                        <button
                                            type="button"
                                            onClick={() => setShowPassword(!showPassword)}
                                            style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'inherit' }}
                                        >
                                            {showPassword ? <EyeOffIcon /> : <EyeIcon />}
                                        </button>
                                    }
                                />

                                {/* Password strength indicator */}
                                {formData.password && (
                                    <div style={{ marginTop: '0.5rem' }}>
                                        <div style={{ display: 'flex', gap: '4px', marginBottom: '0.5rem' }}>
                                            {[1, 2, 3, 4].map((i) => (
                                                <div
                                                    key={i}
                                                    style={{
                                                        flex: 1,
                                                        height: '4px',
                                                        borderRadius: '2px',
                                                        background:
                                                            i <= Math.ceil(passwordStrength.score / 1.5)
                                                                ? passwordStrength.color
                                                                : 'var(--border-color)',
                                                    }}
                                                />
                                            ))}
                                        </div>
                                        <span style={{ fontSize: '0.75rem', color: passwordStrength.color }}>
                                            {passwordStrength.label}
                                        </span>
                                    </div>
                                )}

                                {/* Password requirements */}
                                {formData.password && (
                                    <ul style={{ marginTop: '0.75rem', listStyle: 'none', padding: 0 }}>
                                        {passwordChecks.map((check) => (
                                            <li
                                                key={check.label}
                                                style={{
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    gap: '0.5rem',
                                                    fontSize: '0.8125rem',
                                                    color: check.valid ? 'var(--success)' : 'var(--text-muted)',
                                                    marginBottom: '0.25rem',
                                                }}
                                            >
                                                {check.valid ? (
                                                    <CheckIcon />
                                                ) : (
                                                    <span style={{ width: '14px', height: '14px', display: 'inline-block' }}>○</span>
                                                )}
                                                {check.label}
                                            </li>
                                        ))}
                                    </ul>
                                )}
                            </div>

                            <div style={{ marginTop: '1rem' }}>
                                <Input
                                    type={showPassword ? 'text' : 'password'}
                                    label="Confirm password"
                                    placeholder="Confirm your password"
                                    value={formData.confirm_password}
                                    onChange={(e) => handleChange('confirm_password', e.target.value)}
                                    error={validationErrors.confirm_password}
                                    required
                                    autoComplete="new-password"
                                />
                            </div>

                            <div style={{ marginTop: '1.5rem' }}>
                                <Checkbox
                                    label={
                                        <>
                                            I agree to the <a href="/terms" style={{ color: 'var(--primary)' }}>Terms of Service</a> and{' '}
                                            <a href="/privacy" style={{ color: 'var(--primary)' }}>Privacy Policy</a>
                                        </>
                                    }
                                    checked={acceptTerms}
                                    onChange={(e) => setAcceptTerms(e.target.checked)}
                                    error={validationErrors.terms}
                                />
                            </div>

                            <div style={{ marginTop: '1.5rem' }}>
                                <Button type="submit" fullWidth isLoading={isLoading} size="lg">
                                    Create account
                                </Button>
                            </div>
                        </form>

                        <div className="auth-footer">
                            Already have an account? <Link to="/login">Sign in</Link>
                        </div>
                    </div>
                </div>
            </div>

            <div className="auth-right">
                <div className="auth-branding">
                    <h2 className="auth-branding-title">CRM Kilang Desa Murni Batik</h2>
                    <p className="auth-branding-subtitle">
                        Join thousands of businesses already using our CRM to
                        streamline their sales process and grow their revenue.
                    </p>
                </div>
            </div>
        </div>
    );
}

export default RegisterPage;
