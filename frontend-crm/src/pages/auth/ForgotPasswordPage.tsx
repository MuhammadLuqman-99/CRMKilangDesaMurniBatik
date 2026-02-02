// ============================================
// Forgot Password Page
// Production-Ready Password Recovery Screen
// ============================================

import { useState, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { authService } from '../../services';
import { Button, Input } from '../../components/ui';

const ArrowLeftIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="19" y1="12" x2="5" y2="12" /><polyline points="12 19 5 12 12 5" />
    </svg>
);

export function ForgotPasswordPage() {
    const [email, setEmail] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            await authService.forgotPassword({ email });
            setSuccess(true);
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to send reset email';
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
                                    ✉️
                                </div>
                                <h1 className="auth-title">Check your email</h1>
                                <p className="auth-subtitle">
                                    We've sent password reset instructions to <strong>{email}</strong>.
                                    Please check your inbox and follow the link to reset your password.
                                </p>
                            </div>

                            <div style={{ marginTop: '1.5rem' }}>
                                <Link to="/login">
                                    <Button fullWidth variant="outline" leftIcon={<ArrowLeftIcon />}>
                                        Back to login
                                    </Button>
                                </Link>
                            </div>

                            <p className="auth-footer" style={{ marginTop: '1.5rem' }}>
                                Didn't receive the email?{' '}
                                <button
                                    type="button"
                                    onClick={() => setSuccess(false)}
                                    style={{ background: 'none', border: 'none', color: 'var(--primary)', cursor: 'pointer' }}
                                >
                                    Try again
                                </button>
                            </p>
                        </div>
                    </div>
                </div>
                <div className="auth-right">
                    <div className="auth-branding">
                        <h2 className="auth-branding-title">CRM Kilang Desa Murni Batik</h2>
                        <p className="auth-branding-subtitle">
                            Your password reset is on its way. Check your email and follow the instructions.
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
                            <div className="auth-logo">🔑</div>
                            <h1 className="auth-title">Forgot password?</h1>
                            <p className="auth-subtitle">
                                No worries, we'll send you instructions to reset your password.
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
                                type="email"
                                label="Email address"
                                placeholder="Enter your email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                required
                                autoComplete="email"
                            />

                            <div style={{ marginTop: '1.5rem' }}>
                                <Button type="submit" fullWidth isLoading={isLoading} size="lg">
                                    Send reset link
                                </Button>
                            </div>
                        </form>

                        <div className="auth-footer" style={{ marginTop: '1.5rem' }}>
                            <Link to="/login" style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem' }}>
                                <ArrowLeftIcon />
                                Back to login
                            </Link>
                        </div>
                    </div>
                </div>
            </div>

            <div className="auth-right">
                <div className="auth-branding">
                    <h2 className="auth-branding-title">CRM Kilang Desa Murni Batik</h2>
                    <p className="auth-branding-subtitle">
                        We'll help you regain access to your account quickly and securely.
                    </p>
                </div>
            </div>
        </div>
    );
}

export default ForgotPasswordPage;
