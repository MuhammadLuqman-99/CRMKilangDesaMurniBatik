// ============================================
// Login Page
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const LoginPage: React.FC = () => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    const { login, isAuthenticated } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();

    // Redirect if already authenticated
    React.useEffect(() => {
        if (isAuthenticated) {
            const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/';
            navigate(from, { replace: true });
        }
    }, [isAuthenticated, navigate, location]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            await login(email, password);
            const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/';
            navigate(from, { replace: true });
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Invalid email or password');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="login-page">
            <div className="login-container">
                <div className="login-card">
                    {/* Header */}
                    <div className="login-header">
                        <div className="login-logo">CRM</div>
                        <h1 className="login-title">Admin Dashboard</h1>
                        <p className="login-subtitle">Sign in to manage your CRM platform</p>
                    </div>

                    {/* Error Message */}
                    {error && (
                        <div className="login-error">
                            {error}
                        </div>
                    )}

                    {/* Login Form */}
                    <form onSubmit={handleSubmit}>
                        <div className="form-group">
                            <label className="form-label" htmlFor="email">
                                Email Address
                            </label>
                            <input
                                id="email"
                                type="email"
                                className="form-input"
                                placeholder="admin@example.com"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                required
                                autoComplete="email"
                                autoFocus
                            />
                        </div>

                        <div className="form-group">
                            <label className="form-label" htmlFor="password">
                                Password
                            </label>
                            <input
                                id="password"
                                type="password"
                                className="form-input"
                                placeholder="Enter your password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                required
                                autoComplete="current-password"
                            />
                        </div>

                        <button
                            type="submit"
                            className="btn btn-primary btn-lg w-full"
                            disabled={isLoading}
                            style={{ marginTop: 'var(--space-4)' }}
                        >
                            {isLoading ? (
                                <>
                                    <span className="loading-spinner" style={{ width: 18, height: 18 }} />
                                    Signing in...
                                </>
                            ) : (
                                'Sign In'
                            )}
                        </button>
                    </form>

                    {/* Demo Credentials */}
                    <div style={{
                        marginTop: 'var(--space-6)',
                        padding: 'var(--space-4)',
                        background: 'var(--bg-tertiary)',
                        borderRadius: 'var(--radius-lg)',
                        fontSize: '0.75rem',
                        color: 'var(--text-muted)',
                    }}>
                        <strong style={{ color: 'var(--text-secondary)' }}>Demo Credentials:</strong>
                        <div style={{ marginTop: 'var(--space-2)' }}>
                            Email: <code style={{ color: 'var(--primary)' }}>admin@example.com</code>
                        </div>
                        <div>
                            Password: <code style={{ color: 'var(--primary)' }}>admin123</code>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default LoginPage;
