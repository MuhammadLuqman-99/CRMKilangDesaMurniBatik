// ============================================
// Security Settings Page
// Password change and security settings
// ============================================

import { useState, type FormEvent } from 'react';
import { Button, Input, Card, Badge } from '../../components/ui';
import { useToast } from '../../components/ui/Toast';
import { authService } from '../../services';

// SVG Icons
const CheckIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="20 6 9 17 4 12" />
    </svg>
);

const XIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
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

export function SecuritySettingsPage() {
    const { showToast } = useToast();
    const [isLoading, setIsLoading] = useState(false);
    const [showCurrentPassword, setShowCurrentPassword] = useState(false);
    const [showNewPassword, setShowNewPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

    const [form, setForm] = useState({
        current_password: '',
        new_password: '',
        confirm_password: '',
    });

    const [errors, setErrors] = useState<Record<string, string>>({});

    // Password strength checks
    const passwordChecks = {
        length: form.new_password.length >= 8,
        uppercase: /[A-Z]/.test(form.new_password),
        lowercase: /[a-z]/.test(form.new_password),
        number: /[0-9]/.test(form.new_password),
        special: /[!@#$%^&*(),.?":{}|<>]/.test(form.new_password),
    };

    const isPasswordStrong = Object.values(passwordChecks).every(Boolean);
    const passwordsMatch = form.new_password === form.confirm_password && form.confirm_password.length > 0;

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setErrors({});

        // Validate
        const newErrors: Record<string, string> = {};

        if (!form.current_password) {
            newErrors.current_password = 'Current password is required';
        }
        if (!form.new_password) {
            newErrors.new_password = 'New password is required';
        } else if (!isPasswordStrong) {
            newErrors.new_password = 'Password does not meet requirements';
        }
        if (form.new_password !== form.confirm_password) {
            newErrors.confirm_password = 'Passwords do not match';
        }

        if (Object.keys(newErrors).length > 0) {
            setErrors(newErrors);
            return;
        }

        setIsLoading(true);

        try {
            await authService.changePassword({
                current_password: form.current_password,
                new_password: form.new_password,
            });
            showToast('Password changed successfully', 'success');
            setForm({ current_password: '', new_password: '', confirm_password: '' });
        } catch (error) {
            showToast('Failed to change password. Please check your current password.', 'error');
        } finally {
            setIsLoading(false);
        }
    };

    // Mock session data
    const activeSessions = [
        {
            id: '1',
            device: 'Chrome on Windows',
            location: 'Kuala Lumpur, Malaysia',
            ip: '175.139.xxx.xxx',
            lastActive: 'Active now',
            current: true,
        },
        {
            id: '2',
            device: 'Safari on iPhone',
            location: 'Shah Alam, Malaysia',
            ip: '60.54.xxx.xxx',
            lastActive: '2 hours ago',
            current: false,
        },
        {
            id: '3',
            device: 'Firefox on MacOS',
            location: 'Singapore',
            ip: '203.116.xxx.xxx',
            lastActive: 'Yesterday',
            current: false,
        },
    ];

    return (
        <div>
            {/* Change Password */}
            <Card padding="lg">
                <h2 className="text-lg font-semibold mb-2">Change Password</h2>
                <p className="text-sm text-muted mb-6">
                    Update your password to keep your account secure.
                </p>

                <form onSubmit={handleSubmit}>
                    <div className="max-w-md">
                        <div className="mb-4">
                            <Input
                                type={showCurrentPassword ? 'text' : 'password'}
                                label="Current Password"
                                placeholder="Enter your current password"
                                value={form.current_password}
                                onChange={(e) => setForm((prev) => ({ ...prev, current_password: e.target.value }))}
                                error={errors.current_password}
                                rightIcon={
                                    <button
                                        type="button"
                                        onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                                        style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'inherit' }}
                                    >
                                        {showCurrentPassword ? <EyeOffIcon /> : <EyeIcon />}
                                    </button>
                                }
                            />
                        </div>

                        <div className="mb-4">
                            <Input
                                type={showNewPassword ? 'text' : 'password'}
                                label="New Password"
                                placeholder="Enter your new password"
                                value={form.new_password}
                                onChange={(e) => setForm((prev) => ({ ...prev, new_password: e.target.value }))}
                                error={errors.new_password}
                                rightIcon={
                                    <button
                                        type="button"
                                        onClick={() => setShowNewPassword(!showNewPassword)}
                                        style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'inherit' }}
                                    >
                                        {showNewPassword ? <EyeOffIcon /> : <EyeIcon />}
                                    </button>
                                }
                            />
                        </div>

                        {/* Password Strength Indicators */}
                        {form.new_password && (
                            <div className="mb-4 p-4 bg-tertiary rounded-lg">
                                <p className="text-sm font-medium mb-3">Password requirements:</p>
                                <div className="grid grid-cols-2 gap-2 text-sm">
                                    <div className={`flex items-center gap-2 ${passwordChecks.length ? 'text-success' : 'text-muted'}`}>
                                        {passwordChecks.length ? <CheckIcon /> : <XIcon />}
                                        At least 8 characters
                                    </div>
                                    <div className={`flex items-center gap-2 ${passwordChecks.uppercase ? 'text-success' : 'text-muted'}`}>
                                        {passwordChecks.uppercase ? <CheckIcon /> : <XIcon />}
                                        One uppercase letter
                                    </div>
                                    <div className={`flex items-center gap-2 ${passwordChecks.lowercase ? 'text-success' : 'text-muted'}`}>
                                        {passwordChecks.lowercase ? <CheckIcon /> : <XIcon />}
                                        One lowercase letter
                                    </div>
                                    <div className={`flex items-center gap-2 ${passwordChecks.number ? 'text-success' : 'text-muted'}`}>
                                        {passwordChecks.number ? <CheckIcon /> : <XIcon />}
                                        One number
                                    </div>
                                    <div className={`flex items-center gap-2 ${passwordChecks.special ? 'text-success' : 'text-muted'}`}>
                                        {passwordChecks.special ? <CheckIcon /> : <XIcon />}
                                        One special character
                                    </div>
                                </div>
                            </div>
                        )}

                        <div className="mb-6">
                            <Input
                                type={showConfirmPassword ? 'text' : 'password'}
                                label="Confirm New Password"
                                placeholder="Confirm your new password"
                                value={form.confirm_password}
                                onChange={(e) => setForm((prev) => ({ ...prev, confirm_password: e.target.value }))}
                                error={errors.confirm_password}
                                rightIcon={
                                    <button
                                        type="button"
                                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                        style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'inherit' }}
                                    >
                                        {showConfirmPassword ? <EyeOffIcon /> : <EyeIcon />}
                                    </button>
                                }
                            />
                            {form.confirm_password && (
                                <div className={`flex items-center gap-2 mt-2 text-sm ${passwordsMatch ? 'text-success' : 'text-danger'}`}>
                                    {passwordsMatch ? <CheckIcon /> : <XIcon />}
                                    {passwordsMatch ? 'Passwords match' : 'Passwords do not match'}
                                </div>
                            )}
                        </div>

                        <Button
                            type="submit"
                            isLoading={isLoading}
                            disabled={!isPasswordStrong || !passwordsMatch}
                        >
                            Update Password
                        </Button>
                    </div>
                </form>
            </Card>

            {/* Two-Factor Authentication */}
            <Card padding="lg" className="mt-6">
                <div className="flex justify-between items-start mb-4">
                    <div>
                        <h2 className="text-lg font-semibold mb-1">Two-Factor Authentication</h2>
                        <p className="text-sm text-muted">
                            Add an extra layer of security to your account.
                        </p>
                    </div>
                    <Badge variant="default">Not Enabled</Badge>
                </div>
                <Button variant="outline">Enable 2FA</Button>
            </Card>

            {/* Active Sessions */}
            <Card padding="lg" className="mt-6">
                <div className="flex justify-between items-center mb-4">
                    <div>
                        <h2 className="text-lg font-semibold mb-1">Active Sessions</h2>
                        <p className="text-sm text-muted">
                            Manage and log out of your active sessions on other browsers and devices.
                        </p>
                    </div>
                    <Button variant="outline" size="sm">
                        Log Out All Other Sessions
                    </Button>
                </div>

                <div className="divide-y">
                    {activeSessions.map((session) => (
                        <div key={session.id} className="flex justify-between items-center py-4">
                            <div className="flex items-center gap-4">
                                <div
                                    style={{
                                        width: '40px',
                                        height: '40px',
                                        borderRadius: '8px',
                                        background: 'var(--bg-tertiary)',
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        fontSize: '1.25rem',
                                    }}
                                >
                                    {session.device.includes('iPhone') ? '📱' : '💻'}
                                </div>
                                <div>
                                    <div className="flex items-center gap-2">
                                        <p className="font-medium">{session.device}</p>
                                        {session.current && <Badge variant="success" size="sm">This device</Badge>}
                                    </div>
                                    <p className="text-sm text-muted">
                                        {session.location} · {session.ip} · {session.lastActive}
                                    </p>
                                </div>
                            </div>
                            {!session.current && (
                                <Button variant="ghost" size="sm">
                                    Log out
                                </Button>
                            )}
                        </div>
                    ))}
                </div>
            </Card>
        </div>
    );
}

export default SecuritySettingsPage;
