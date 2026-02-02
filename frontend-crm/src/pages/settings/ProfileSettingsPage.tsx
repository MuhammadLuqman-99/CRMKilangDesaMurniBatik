// ============================================
// Profile Settings Page
// User profile management
// ============================================

import { useState, useEffect, type FormEvent } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { Button, Input, Select, Card } from '../../components/ui';
import { useToast } from '../../components/ui/Toast';

const timezones = [
    { value: 'Asia/Kuala_Lumpur', label: 'Malaysia (GMT+8)' },
    { value: 'Asia/Singapore', label: 'Singapore (GMT+8)' },
    { value: 'Asia/Jakarta', label: 'Indonesia - Jakarta (GMT+7)' },
    { value: 'Asia/Bangkok', label: 'Thailand (GMT+7)' },
    { value: 'Asia/Tokyo', label: 'Japan (GMT+9)' },
    { value: 'Asia/Hong_Kong', label: 'Hong Kong (GMT+8)' },
    { value: 'Asia/Shanghai', label: 'China (GMT+8)' },
    { value: 'Australia/Sydney', label: 'Australia - Sydney (GMT+10)' },
    { value: 'Europe/London', label: 'United Kingdom (GMT)' },
    { value: 'America/New_York', label: 'US - Eastern (GMT-5)' },
    { value: 'America/Los_Angeles', label: 'US - Pacific (GMT-8)' },
];

export function ProfileSettingsPage() {
    const { user, updateProfile } = useAuth();
    const { showToast } = useToast();
    const [isLoading, setIsLoading] = useState(false);
    const [avatarPreview, setAvatarPreview] = useState<string | null>(null);

    const [form, setForm] = useState({
        first_name: '',
        last_name: '',
        email: '',
        phone: '',
        title: '',
        department: '',
        timezone: 'Asia/Kuala_Lumpur',
    });

    useEffect(() => {
        if (user) {
            setForm({
                first_name: user.first_name || '',
                last_name: user.last_name || '',
                email: user.email || '',
                phone: user.phone || '',
                title: user.title || '',
                department: user.department || '',
                timezone: user.timezone || 'Asia/Kuala_Lumpur',
            });
            setAvatarPreview(user.avatar_url || null);
        }
    }, [user]);

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setIsLoading(true);

        try {
            await updateProfile?.({
                first_name: form.first_name,
                last_name: form.last_name,
                phone: form.phone || undefined,
                title: form.title || undefined,
                department: form.department || undefined,
                timezone: form.timezone,
            });
            showToast('Profile updated successfully', 'success');
        } catch (error) {
            showToast('Failed to update profile', 'error');
        } finally {
            setIsLoading(false);
        }
    };

    const handleAvatarChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = () => {
                setAvatarPreview(reader.result as string);
            };
            reader.readAsDataURL(file);
        }
    };

    const getInitials = () => {
        return `${form.first_name?.charAt(0) || ''}${form.last_name?.charAt(0) || ''}`.toUpperCase() || 'U';
    };

    return (
        <div>
            <Card padding="lg">
                <h2 className="text-lg font-semibold mb-6">Profile Information</h2>

                <form onSubmit={handleSubmit}>
                    {/* Avatar Section */}
                    <div className="flex items-center gap-6 mb-8 pb-6 border-b">
                        <div
                            className="profile-avatar"
                            style={{
                                width: '96px',
                                height: '96px',
                                borderRadius: '50%',
                                background: avatarPreview
                                    ? `url(${avatarPreview}) center/cover`
                                    : 'linear-gradient(135deg, var(--primary), #60a5fa)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                fontSize: '2rem',
                                fontWeight: 600,
                                color: 'white',
                            }}
                        >
                            {!avatarPreview && getInitials()}
                        </div>
                        <div>
                            <h3 className="font-medium mb-1">Profile Photo</h3>
                            <p className="text-sm text-muted mb-3">
                                JPG, PNG or GIF. Max size 2MB.
                            </p>
                            <div className="flex gap-2">
                                <label className="btn btn-outline btn-sm" style={{ cursor: 'pointer' }}>
                                    Upload Photo
                                    <input
                                        type="file"
                                        accept="image/*"
                                        onChange={handleAvatarChange}
                                        style={{ display: 'none' }}
                                    />
                                </label>
                                {avatarPreview && (
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => setAvatarPreview(null)}
                                    >
                                        Remove
                                    </Button>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Form Fields */}
                    <div className="grid grid-cols-2 gap-4">
                        <Input
                            label="First Name"
                            value={form.first_name}
                            onChange={(e) => setForm((prev) => ({ ...prev, first_name: e.target.value }))}
                            required
                        />
                        <Input
                            label="Last Name"
                            value={form.last_name}
                            onChange={(e) => setForm((prev) => ({ ...prev, last_name: e.target.value }))}
                            required
                        />
                        <div className="col-span-2">
                            <Input
                                type="email"
                                label="Email Address"
                                value={form.email}
                                disabled
                                hint="Contact your administrator to change your email address."
                            />
                        </div>
                        <Input
                            type="tel"
                            label="Phone Number"
                            placeholder="+60123456789"
                            value={form.phone}
                            onChange={(e) => setForm((prev) => ({ ...prev, phone: e.target.value }))}
                        />
                        <Input
                            label="Job Title"
                            placeholder="e.g. Sales Manager"
                            value={form.title}
                            onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
                        />
                        <Input
                            label="Department"
                            placeholder="e.g. Sales"
                            value={form.department}
                            onChange={(e) => setForm((prev) => ({ ...prev, department: e.target.value }))}
                        />
                        <Select
                            label="Timezone"
                            options={timezones}
                            value={form.timezone}
                            onChange={(e) => setForm((prev) => ({ ...prev, timezone: e.target.value }))}
                        />
                    </div>

                    <div className="flex justify-end gap-3 mt-6 pt-6 border-t">
                        <Button type="button" variant="outline">
                            Cancel
                        </Button>
                        <Button type="submit" isLoading={isLoading}>
                            Save Changes
                        </Button>
                    </div>
                </form>
            </Card>

            {/* Danger Zone */}
            <Card padding="lg" className="mt-6">
                <h2 className="text-lg font-semibold mb-2 text-danger">Danger Zone</h2>
                <p className="text-sm text-muted mb-4">
                    Once you delete your account, there is no going back. Please be certain.
                </p>
                <Button variant="danger" disabled>
                    Delete Account
                </Button>
            </Card>
        </div>
    );
}

export default ProfileSettingsPage;
