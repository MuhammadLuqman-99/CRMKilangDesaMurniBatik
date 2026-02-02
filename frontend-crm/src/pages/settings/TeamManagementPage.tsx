// ============================================
// Team Management Page
// Team members and permissions management
// ============================================

import { useState } from 'react';
import { Button, Input, Select, Badge, Card, Modal } from '../../components/ui';
import { useToast } from '../../components/ui/Toast';

// SVG Icons
const SearchIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
);

const PlusIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
    </svg>
);

const MoreIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="1" /><circle cx="19" cy="12" r="1" /><circle cx="5" cy="12" r="1" />
    </svg>
);

interface TeamMember {
    id: string;
    name: string;
    email: string;
    avatar?: string;
    role: string;
    status: 'active' | 'pending' | 'inactive';
    leads_count: number;
    deals_count: number;
    deals_value: number;
    last_active: string;
}

const roleOptions = [
    { value: '', label: 'All Roles' },
    { value: 'admin', label: 'Admin' },
    { value: 'manager', label: 'Sales Manager' },
    { value: 'rep', label: 'Sales Rep' },
    { value: 'viewer', label: 'Viewer' },
];

const rolePermissions = {
    admin: {
        label: 'Admin',
        color: '#8b5cf6',
        permissions: ['Full access to all features and settings'],
    },
    manager: {
        label: 'Sales Manager',
        color: '#3b82f6',
        permissions: ['View and manage team', 'Access reports', 'Manage pipelines'],
    },
    rep: {
        label: 'Sales Rep',
        color: '#10b981',
        permissions: ['Manage own leads and deals', 'View team calendar'],
    },
    viewer: {
        label: 'Viewer',
        color: '#94a3b8',
        permissions: ['View-only access to CRM data'],
    },
};

export function TeamManagementPage() {
    const { showToast } = useToast();
    const [searchQuery, setSearchQuery] = useState('');
    const [roleFilter, setRoleFilter] = useState('');
    const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
    const [inviteEmail, setInviteEmail] = useState('');
    const [inviteRole, setInviteRole] = useState('rep');

    const [teamMembers] = useState<TeamMember[]>([
        {
            id: '1',
            name: 'Ahmad Razak',
            email: 'ahmad@kilangbatik.com.my',
            role: 'admin',
            status: 'active',
            leads_count: 45,
            deals_count: 12,
            deals_value: 580000,
            last_active: 'Active now',
        },
        {
            id: '2',
            name: 'Siti Aminah',
            email: 'siti@kilangbatik.com.my',
            role: 'manager',
            status: 'active',
            leads_count: 38,
            deals_count: 8,
            deals_value: 320000,
            last_active: '2 hours ago',
        },
        {
            id: '3',
            name: 'Muhammad Hafiz',
            email: 'hafiz@kilangbatik.com.my',
            role: 'rep',
            status: 'active',
            leads_count: 52,
            deals_count: 15,
            deals_value: 420000,
            last_active: '1 hour ago',
        },
        {
            id: '4',
            name: 'Nurul Aisyah',
            email: 'nurul@kilangbatik.com.my',
            role: 'rep',
            status: 'active',
            leads_count: 41,
            deals_count: 10,
            deals_value: 280000,
            last_active: 'Yesterday',
        },
        {
            id: '5',
            name: 'Ali Hassan',
            email: 'ali@kilangbatik.com.my',
            role: 'rep',
            status: 'pending',
            leads_count: 0,
            deals_count: 0,
            deals_value: 0,
            last_active: 'Never',
        },
    ]);

    const filteredMembers = teamMembers.filter((member) => {
        const matchesSearch =
            member.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            member.email.toLowerCase().includes(searchQuery.toLowerCase());
        const matchesRole = !roleFilter || member.role === roleFilter;
        return matchesSearch && matchesRole;
    });

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('ms-MY', {
            style: 'currency',
            currency: 'MYR',
            minimumFractionDigits: 0,
        }).format(value);
    };

    const handleInvite = () => {
        showToast(`Invitation sent to ${inviteEmail}`, 'success');
        setIsInviteModalOpen(false);
        setInviteEmail('');
        setInviteRole('rep');
    };

    const getStatusBadge = (status: string) => {
        const variants: Record<string, 'success' | 'warning' | 'default'> = {
            active: 'success',
            pending: 'warning',
            inactive: 'default',
        };
        return <Badge variant={variants[status] || 'default'} size="sm">{status}</Badge>;
    };

    const getRoleBadge = (role: string) => {
        const roleInfo = rolePermissions[role as keyof typeof rolePermissions];
        return (
            <Badge
                variant="default"
                size="sm"
                style={{ borderLeft: `4px solid ${roleInfo?.color || '#94a3b8'}` }}
            >
                {roleInfo?.label || role}
            </Badge>
        );
    };

    return (
        <div>
            <Card padding="lg">
                <div className="flex justify-between items-center mb-6">
                    <div>
                        <h2 className="text-lg font-semibold mb-1">Team Members</h2>
                        <p className="text-sm text-muted">
                            Manage your team members and their roles.
                        </p>
                    </div>
                    <Button onClick={() => setIsInviteModalOpen(true)}>
                        <PlusIcon />
                        Invite Member
                    </Button>
                </div>

                {/* Search & Filter */}
                <div className="flex gap-4 mb-6">
                    <div className="flex-1 relative">
                        <span style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }}>
                            <SearchIcon />
                        </span>
                        <Input
                            type="text"
                            placeholder="Search team members..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            style={{ paddingLeft: '40px' }}
                        />
                    </div>
                    <Select
                        options={roleOptions}
                        value={roleFilter}
                        onChange={(e) => setRoleFilter(e.target.value)}
                        style={{ width: '200px' }}
                    />
                </div>

                {/* Team Table */}
                <div className="overflow-x-auto">
                    <table className="table">
                        <thead>
                            <tr>
                                <th>Member</th>
                                <th>Role</th>
                                <th>Status</th>
                                <th>Leads</th>
                                <th>Deals</th>
                                <th>Deal Value</th>
                                <th>Last Active</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredMembers.map((member) => (
                                <tr key={member.id}>
                                    <td>
                                        <div className="flex items-center gap-3">
                                            <div
                                                style={{
                                                    width: '40px',
                                                    height: '40px',
                                                    borderRadius: '50%',
                                                    background: 'linear-gradient(135deg, var(--primary), #60a5fa)',
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    justifyContent: 'center',
                                                    color: 'white',
                                                    fontWeight: 600,
                                                    fontSize: '0.875rem',
                                                }}
                                            >
                                                {member.name.split(' ').map((n) => n[0]).join('')}
                                            </div>
                                            <div>
                                                <p className="font-medium">{member.name}</p>
                                                <p className="text-sm text-muted">{member.email}</p>
                                            </div>
                                        </div>
                                    </td>
                                    <td>{getRoleBadge(member.role)}</td>
                                    <td>{getStatusBadge(member.status)}</td>
                                    <td>{member.leads_count}</td>
                                    <td>{member.deals_count}</td>
                                    <td className="text-success font-medium">{formatCurrency(member.deals_value)}</td>
                                    <td className="text-muted">{member.last_active}</td>
                                    <td>
                                        <Button variant="ghost" size="sm">
                                            <MoreIcon />
                                        </Button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                {filteredMembers.length === 0 && (
                    <div className="text-center py-8">
                        <p className="text-muted">No team members found matching your criteria.</p>
                    </div>
                )}
            </Card>

            {/* Roles & Permissions */}
            <Card padding="lg" className="mt-6">
                <h2 className="text-lg font-semibold mb-4">Roles & Permissions</h2>
                <div className="grid grid-cols-2 gap-4">
                    {Object.entries(rolePermissions).map(([key, role]) => (
                        <div
                            key={key}
                            className="p-4 rounded-lg border"
                            style={{ borderLeft: `4px solid ${role.color}` }}
                        >
                            <h3 className="font-medium mb-2">{role.label}</h3>
                            <ul className="text-sm text-muted space-y-1">
                                {role.permissions.map((perm, index) => (
                                    <li key={index}>• {perm}</li>
                                ))}
                            </ul>
                        </div>
                    ))}
                </div>
            </Card>

            {/* Invite Modal */}
            <Modal
                isOpen={isInviteModalOpen}
                onClose={() => setIsInviteModalOpen(false)}
                title="Invite Team Member"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsInviteModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleInvite} disabled={!inviteEmail}>
                            Send Invitation
                        </Button>
                    </>
                }
            >
                <div className="space-y-4">
                    <Input
                        type="email"
                        label="Email Address"
                        placeholder="colleague@company.com"
                        value={inviteEmail}
                        onChange={(e) => setInviteEmail(e.target.value)}
                        required
                    />
                    <Select
                        label="Role"
                        options={[
                            { value: 'admin', label: 'Admin' },
                            { value: 'manager', label: 'Sales Manager' },
                            { value: 'rep', label: 'Sales Rep' },
                            { value: 'viewer', label: 'Viewer' },
                        ]}
                        value={inviteRole}
                        onChange={(e) => setInviteRole(e.target.value)}
                    />

                    <div className="p-4 bg-tertiary rounded-lg">
                        <h4 className="font-medium mb-2">
                            {rolePermissions[inviteRole as keyof typeof rolePermissions]?.label} Permissions:
                        </h4>
                        <ul className="text-sm text-muted space-y-1">
                            {rolePermissions[inviteRole as keyof typeof rolePermissions]?.permissions.map((perm, index) => (
                                <li key={index}>• {perm}</li>
                            ))}
                        </ul>
                    </div>
                </div>
            </Modal>
        </div>
    );
}

export default TeamManagementPage;
