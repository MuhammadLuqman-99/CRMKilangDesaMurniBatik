// ============================================
// Email Verification Page
// Production-Ready Email Verification Screen
// ============================================

import { useState, useEffect } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { authService } from '../../services';
import { Button } from '../../components/ui';

export function EmailVerificationPage() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const token = searchParams.get('token');

    const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
    const [errorMessage, setErrorMessage] = useState('');

    useEffect(() => {
        const verifyEmail = async () => {
            if (!token) {
                setStatus('error');
                setErrorMessage('Invalid verification link. Please check your email for the correct link.');
                return;
            }

            try {
                await authService.verifyEmail(token);
                setStatus('success');
            } catch (err) {
                setStatus('error');
                const message = err instanceof Error ? err.message : 'Failed to verify email';
                setErrorMessage(message);
            }
        };

        verifyEmail();
    }, [token]);

    if (status === 'loading') {
        return (
            <div className="auth-page">
                <div className="auth-left">
                    <div className="auth-container">
                        <div className="auth-card text-center">
                            <div className="auth-header">
                                <div className="loading-spinner lg" style={{ margin: '0 auto 1rem' }} />
                                <h1 className="auth-title">Verifying your email...</h1>
                                <p className="auth-subtitle">Please wait while we verify your email address.</p>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="auth-right">
                    <div className="auth-branding">
                        <h2 className="auth-branding-title">CRM Kilang Desa Murni Batik</h2>
                        <p className="auth-branding-subtitle">Almost there! We're verifying your email address.</p>
                    </div>
                </div>
            </div>
        );
    }

    if (status === 'error') {
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
                                    ✕
                                </div>
                                <h1 className="auth-title">Verification failed</h1>
                                <p className="auth-subtitle">{errorMessage}</p>
                            </div>

                            <div style={{ marginTop: '1.5rem', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                                <Link to="/login">
                                    <Button fullWidth variant="outline">
                                        Go to login
                                    </Button>
                                </Link>
                                <Button
                                    fullWidth
                                    onClick={async () => {
                                        try {
                                            await authService.resendVerification();
                                            alert('Verification email sent! Please check your inbox.');
                                        } catch {
                                            alert('Failed to resend verification email. Please try again later.');
                                        }
                                    }}
                                >
                                    Resend verification email
                                </Button>
                            </div>
                        </div>
                    </div>
                </div>
                <div className="auth-right">
                    <div className="auth-branding">
                        <h2 className="auth-branding-title">CRM Kilang Desa Murni Batik</h2>
                        <p className="auth-branding-subtitle">
                            Don't worry, you can request a new verification email.
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
                    <div className="auth-card text-center">
                        <div className="auth-header">
                            <div
                                className="auth-logo"
                                style={{ background: 'var(--success)', margin: '0 auto 1rem' }}
                            >
                                ✓
                            </div>
                            <h1 className="auth-title">Email verified!</h1>
                            <p className="auth-subtitle">
                                Your email has been successfully verified. You can now sign in to your account.
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
                    <h2 className="auth-branding-title">Welcome aboard!</h2>
                    <p className="auth-branding-subtitle">
                        Your account is now verified and ready to use. Start exploring our CRM features!
                    </p>
                </div>
            </div>
        </div>
    );
}

export default EmailVerificationPage;
