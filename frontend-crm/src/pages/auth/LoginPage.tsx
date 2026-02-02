// ============================================
// Login Page
// Production-Ready Login Screen
// ============================================

import { useState, type FormEvent } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { Button, Input, Checkbox } from '../../components/ui';

// SVG Icons
const GoogleIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24">
        <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
        <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
        <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
        <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
    </svg>
);

const MicrosoftIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24">
        <path fill="#F25022" d="M1 1h10v10H1z" />
        <path fill="#00A4EF" d="M1 13h10v10H1z" />
        <path fill="#7FBA00" d="M13 1h10v10H13z" />
        <path fill="#FFB900" d="M13 13h10v10H13z" />
    </svg>
);

const GithubIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
    </svg>
);

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

export function LoginPage() {
    const navigate = useNavigate();
    const location = useLocation();
    const { login, error, isLoading, clearError } = useAuth();

    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    const [rememberMe, setRememberMe] = useState(false);

    const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/';

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        clearError();

        try {
            await login({ email, password });
            navigate(from, { replace: true });
        } catch {
            // Error is handled by AuthContext
        }
    };

    const handleOAuthLogin = (provider: string) => {
        // Redirect to OAuth endpoint
        window.location.href = `/api/v1/auth/oauth/${provider}`;
    };

    return (
        <div className="auth-page">
            <div className="auth-left">
                <div className="auth-container">
                    <div className="auth-card">
                        <div className="auth-header">
                            <div className="auth-logo">CRM</div>
                            <h1 className="auth-title">Welcome back</h1>
                            <p className="auth-subtitle">Sign in to your account to continue</p>
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

                            <div style={{ marginTop: '1rem' }}>
                                <Input
                                    type={showPassword ? 'text' : 'password'}
                                    label="Password"
                                    placeholder="Enter your password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    required
                                    autoComplete="current-password"
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
                            </div>

                            <div className="auth-remember">
                                <Checkbox
                                    label="Remember me"
                                    checked={rememberMe}
                                    onChange={(e) => setRememberMe(e.target.checked)}
                                />
                                <Link to="/forgot-password">Forgot password?</Link>
                            </div>

                            <Button
                                type="submit"
                                fullWidth
                                isLoading={isLoading}
                                size="lg"
                            >
                                Sign in
                            </Button>
                        </form>

                        <div className="auth-divider">
                            <span>or continue with</span>
                        </div>

                        <div className="flex gap-3">
                            <Button
                                type="button"
                                variant="outline"
                                fullWidth
                                onClick={() => handleOAuthLogin('google')}
                                leftIcon={<GoogleIcon />}
                            >
                                Google
                            </Button>
                            <Button
                                type="button"
                                variant="outline"
                                fullWidth
                                onClick={() => handleOAuthLogin('microsoft')}
                                leftIcon={<MicrosoftIcon />}
                            >
                                Microsoft
                            </Button>
                            <Button
                                type="button"
                                variant="outline"
                                fullWidth
                                onClick={() => handleOAuthLogin('github')}
                                leftIcon={<GithubIcon />}
                            >
                                GitHub
                            </Button>
                        </div>

                        <div className="auth-footer">
                            Don't have an account? <Link to="/register">Sign up</Link>
                        </div>
                    </div>
                </div>
            </div>

            <div className="auth-right">
                <div className="auth-branding">
                    <h2 className="auth-branding-title">
                        CRM Kilang Desa Murni Batik
                    </h2>
                    <p className="auth-branding-subtitle">
                        Manage your leads, customers, and sales pipeline with our powerful
                        and intuitive CRM solution designed for Malaysian batik businesses.
                    </p>
                </div>
            </div>
        </div>
    );
}

export default LoginPage;
