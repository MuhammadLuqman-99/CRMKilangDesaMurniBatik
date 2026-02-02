// ============================================
// Notification Settings Page
// Notification preferences management
// ============================================

import { useState, type FormEvent } from 'react';
import { Button, Card, Checkbox } from '../../components/ui';
import { useToast } from '../../components/ui/Toast';

interface NotificationSetting {
    id: string;
    title: string;
    description: string;
    email: boolean;
    push: boolean;
    inApp: boolean;
}

export function NotificationSettingsPage() {
    const { showToast } = useToast();
    const [isLoading, setIsLoading] = useState(false);

    const [settings, setSettings] = useState<NotificationSetting[]>([
        {
            id: 'new_lead',
            title: 'New Lead Assigned',
            description: 'When a new lead is assigned to you',
            email: true,
            push: true,
            inApp: true,
        },
        {
            id: 'deal_update',
            title: 'Deal Stage Changed',
            description: 'When a deal moves to a new stage',
            email: true,
            push: false,
            inApp: true,
        },
        {
            id: 'deal_won',
            title: 'Deal Won',
            description: 'When a deal is marked as won',
            email: true,
            push: true,
            inApp: true,
        },
        {
            id: 'deal_lost',
            title: 'Deal Lost',
            description: 'When a deal is marked as lost',
            email: true,
            push: true,
            inApp: true,
        },
        {
            id: 'task_reminder',
            title: 'Task Reminders',
            description: 'Reminders for upcoming and overdue tasks',
            email: true,
            push: true,
            inApp: true,
        },
        {
            id: 'meeting_reminder',
            title: 'Meeting Reminders',
            description: 'Reminders for upcoming meetings',
            email: true,
            push: true,
            inApp: true,
        },
        {
            id: 'mention',
            title: 'Mentions',
            description: 'When someone mentions you in a comment or note',
            email: false,
            push: true,
            inApp: true,
        },
        {
            id: 'weekly_report',
            title: 'Weekly Summary',
            description: 'Weekly summary of your sales performance',
            email: true,
            push: false,
            inApp: false,
        },
        {
            id: 'team_activity',
            title: 'Team Activity',
            description: 'Updates on team activities and achievements',
            email: false,
            push: false,
            inApp: true,
        },
        {
            id: 'system_updates',
            title: 'System Updates',
            description: 'Important system updates and maintenance notices',
            email: true,
            push: false,
            inApp: true,
        },
    ]);

    const handleToggle = (settingId: string, channel: 'email' | 'push' | 'inApp') => {
        setSettings((prev) =>
            prev.map((setting) =>
                setting.id === settingId
                    ? { ...setting, [channel]: !setting[channel] }
                    : setting
            )
        );
    };

    const handleEnableAll = (channel: 'email' | 'push' | 'inApp') => {
        setSettings((prev) =>
            prev.map((setting) => ({ ...setting, [channel]: true }))
        );
    };

    const handleDisableAll = (channel: 'email' | 'push' | 'inApp') => {
        setSettings((prev) =>
            prev.map((setting) => ({ ...setting, [channel]: false }))
        );
    };

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();
        setIsLoading(true);

        try {
            // API call would go here
            await new Promise((resolve) => setTimeout(resolve, 1000));
            showToast('Notification preferences saved', 'success');
        } catch (error) {
            showToast('Failed to save preferences', 'error');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div>
            <Card padding="lg">
                <div className="flex justify-between items-start mb-6">
                    <div>
                        <h2 className="text-lg font-semibold mb-1">Notification Preferences</h2>
                        <p className="text-sm text-muted">
                            Choose how you want to be notified about activity in your CRM.
                        </p>
                    </div>
                </div>

                <form onSubmit={handleSubmit}>
                    {/* Table Header */}
                    <div
                        className="grid gap-4 pb-4 border-b font-medium text-sm"
                        style={{ gridTemplateColumns: '1fr 100px 100px 100px' }}
                    >
                        <div>Notification Type</div>
                        <div className="text-center">Email</div>
                        <div className="text-center">Push</div>
                        <div className="text-center">In-App</div>
                    </div>

                    {/* Quick Actions */}
                    <div
                        className="grid gap-4 py-3 border-b bg-tertiary -mx-6 px-6"
                        style={{ gridTemplateColumns: '1fr 100px 100px 100px' }}
                    >
                        <div className="text-sm text-muted">Quick actions:</div>
                        <div className="text-center">
                            <button
                                type="button"
                                onClick={() => handleEnableAll('email')}
                                className="text-xs text-primary hover:underline"
                            >
                                Enable all
                            </button>
                        </div>
                        <div className="text-center">
                            <button
                                type="button"
                                onClick={() => handleEnableAll('push')}
                                className="text-xs text-primary hover:underline"
                            >
                                Enable all
                            </button>
                        </div>
                        <div className="text-center">
                            <button
                                type="button"
                                onClick={() => handleEnableAll('inApp')}
                                className="text-xs text-primary hover:underline"
                            >
                                Enable all
                            </button>
                        </div>
                    </div>

                    {/* Notification Settings */}
                    {settings.map((setting) => (
                        <div
                            key={setting.id}
                            className="grid gap-4 py-4 border-b items-center"
                            style={{ gridTemplateColumns: '1fr 100px 100px 100px' }}
                        >
                            <div>
                                <p className="font-medium">{setting.title}</p>
                                <p className="text-sm text-muted">{setting.description}</p>
                            </div>
                            <div className="flex justify-center">
                                <Checkbox
                                    checked={setting.email}
                                    onChange={() => handleToggle(setting.id, 'email')}
                                />
                            </div>
                            <div className="flex justify-center">
                                <Checkbox
                                    checked={setting.push}
                                    onChange={() => handleToggle(setting.id, 'push')}
                                />
                            </div>
                            <div className="flex justify-center">
                                <Checkbox
                                    checked={setting.inApp}
                                    onChange={() => handleToggle(setting.id, 'inApp')}
                                />
                            </div>
                        </div>
                    ))}

                    <div className="flex justify-end gap-3 mt-6 pt-6 border-t">
                        <Button type="button" variant="outline">
                            Reset to Default
                        </Button>
                        <Button type="submit" isLoading={isLoading}>
                            Save Preferences
                        </Button>
                    </div>
                </form>
            </Card>

            {/* Email Digest Settings */}
            <Card padding="lg" className="mt-6">
                <h2 className="text-lg font-semibold mb-2">Email Digest</h2>
                <p className="text-sm text-muted mb-4">
                    Configure your email digest frequency and content.
                </p>

                <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 bg-tertiary rounded-lg">
                        <div>
                            <p className="font-medium">Daily Digest</p>
                            <p className="text-sm text-muted">Receive a daily summary of activities</p>
                        </div>
                        <Checkbox defaultChecked />
                    </div>

                    <div className="flex items-center justify-between p-4 bg-tertiary rounded-lg">
                        <div>
                            <p className="font-medium">Weekly Report</p>
                            <p className="text-sm text-muted">Receive a weekly performance report</p>
                        </div>
                        <Checkbox defaultChecked />
                    </div>

                    <div className="flex items-center justify-between p-4 bg-tertiary rounded-lg">
                        <div>
                            <p className="font-medium">Monthly Summary</p>
                            <p className="text-sm text-muted">Receive a monthly summary of achievements</p>
                        </div>
                        <Checkbox />
                    </div>
                </div>
            </Card>

            {/* Do Not Disturb */}
            <Card padding="lg" className="mt-6">
                <h2 className="text-lg font-semibold mb-2">Do Not Disturb</h2>
                <p className="text-sm text-muted mb-4">
                    Set quiet hours to pause all non-critical notifications.
                </p>

                <div className="flex items-center justify-between p-4 bg-tertiary rounded-lg">
                    <div>
                        <p className="font-medium">Enable Quiet Hours</p>
                        <p className="text-sm text-muted">10:00 PM - 8:00 AM (local time)</p>
                    </div>
                    <Checkbox />
                </div>
            </Card>
        </div>
    );
}

export default NotificationSettingsPage;
