// ============================================
// Reset Password Page
// Production-Ready Password Reset Screen
// ============================================

import { useState, type FormEvent } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { authService } from '../../services';
import { Button, Input } from '../../components/ui';

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

export function ResetPasswordPage() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const token = searchParams.get('token');

    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);

    if (!token) {
        return (
            <div className="auth-page">
                <div className="auth-left">
                    <div className="auth-container">
                        <div className="auth-card text-center">
                            <div className="auth-header">
                                <div
                                    className="auth-logo"
                                    style={{ background: 'var(--danger)', margin: '0 auto 1rem' }}
                                >
                                    ⚠️
                                </div>
                                <h1 className="auth-title">Invalid Link</h1>
                                <p className="auth-subtitle">
                                    This password reset link is invalid or has expired.
                                    Please request a new password reset.
                                </p>
                            </div>

                            <div style={{ marginTop: '1.5rem' }}>
                                <Link to="/forgot-password">
                                    <Button fullWidth size="lg">
                                        Request new link
                                    </Button>
                                </Link>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="auth-right">
                    <div className="auth-branding">
                        <h2 className="auth-branding-title">CRM Kilang Desa Murni Batik</h2>
                        <p className="auth-branding-subtitle">
                            Don't worry, you can request a new password reset link.
                        </p>
                    </div>
                </div>
            </div>
        );
    }

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setError('');

        if (password !== confirmPassword) {
            setError('Passwords do not match');
            return;
        }

        if (password.length < 8) {
            setError('Password must be at least 8 characters');
            return;
        }

        setIsLoading(true);

        try {
            await authService.resetPassword({ token, password });
            setSuccess(true);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to reset password';
            setError(message);
        } finally {
            setIsLoading(false);
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
                                <h1 className="auth-title">Password reset successful</h1>
                                <p className="auth-subtitle">
                                    Your password has been successfully reset.
                                    You can now sign in with your new password.
                                </p>
                            </div>

                            <div style={{ marginTop: '1.5rem' }}>
                                <Button fullWidth size="lg" onClick={() => navigate('/login')}>
                                    Go to login
                                </Button>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="auth-right">
                    <div className="auth-branding">
                        <h2 className="auth-branding-title">CRM Kilang Desa Murni Batik</h2>
                        <p className="auth-branding-subtitle">
                            Your account is secure. You can now continue using the CRM.
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
                            <div className="auth-logo">🔐</div>
                            <h1 className="auth-title">Reset your password</h1>
                            <p className="auth-subtitle">
                                Enter a new password for your account.
                            </p>
                        </div>

                        {error && (
                            <div className="auth-error">
                                <span>⚠️</span>
                                <span>{error}</span>
                            </div>
                        )}

                        <form onSubmit={handleSubmit}>
                            <Input
                                type={showPassword ? 'text' : 'password'}
                                label="New password"
                                placeholder="Enter new password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
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

                            <div style={{ marginTop: '1rem' }}>
                                <Input
                                    type={showPassword ? 'text' : 'password'}
                                    label="Confirm new password"
                                    placeholder="Confirm new password"
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    required
                                    autoComplete="new-password"
                                />
                            </div>

                            <div style={{ marginTop: '1.5rem' }}>
                                <Button type="submit" fullWidth isLoading={isLoading} size="lg">
                                    Reset password
                                </Button>
                            </div>
                        </form>

                        <div className="auth-footer" style={{ marginTop: '1.5rem' }}>
                            Remember your password? <Link to="/login">Sign in</Link>
                        </div>
                    </div>
                </div>
            </div>

            <div className="auth-right">
                <div className="auth-branding">
                    <h2 className="auth-branding-title">CRM Kilang Desa Murni Batik</h2>
                    <p className="auth-branding-subtitle">
                        Choose a strong password to keep your account secure.
                    </p>
                </div>
            </div>
        </div>
    );
}

export default ResetPasswordPage;
