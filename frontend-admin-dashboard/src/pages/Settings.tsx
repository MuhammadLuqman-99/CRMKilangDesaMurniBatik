// ============================================
// Settings Page
// Admin Dashboard - CRM Kilang Desa Murni Batik
// ============================================

import React from 'react';

const Settings: React.FC = () => {
    return (
        <div>
            {/* Page Header */}
            <div className="page-header">
                <div>
                    <h1 className="page-title">Settings</h1>
                    <p className="page-description">Configure admin dashboard preferences</p>
                </div>
            </div>

            <div className="card">
                <div className="card-body">
                    <div className="empty-state">
                        <div className="empty-state-icon">⚙️</div>
                        <div className="empty-state-title">Settings Coming Soon</div>
                        <div className="empty-state-description">
                            Additional configuration options will be available in a future update.
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Settings;
