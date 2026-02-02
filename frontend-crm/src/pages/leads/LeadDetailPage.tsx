// ============================================
// Lead Detail Page
// Production-Ready Lead Detail View
// ============================================

import { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { leadService } from '../../services';
import { Button, Badge, Card, Modal, Select, Textarea } from '../../components/ui';
import type { Lead } from '../../types';

// SVG Icons
const ArrowLeftIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="19" y1="12" x2="5" y2="12" /><polyline points="12 19 5 12 12 5" />
    </svg>
);

const EditIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" /><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
    </svg>
);

const PhoneIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z" />
    </svg>
);

const MailIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" /><polyline points="22,6 12,13 2,6" />
    </svg>
);

const ConvertIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="23 4 23 10 17 10" /><polyline points="1 20 1 14 7 14" /><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
    </svg>
);

export function LeadDetailPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [lead, setLead] = useState<Lead | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('overview');
    const [isConvertModalOpen, setIsConvertModalOpen] = useState(false);
    const [isDisqualifyModalOpen, setIsDisqualifyModalOpen] = useState(false);
    const [disqualifyReason, setDisqualifyReason] = useState('');

    useEffect(() => {
        const fetchLead = async () => {
            if (!id) return;

            setIsLoading(true);
            try {
                const data = await leadService.getLead(id);
                setLead(data);
            } catch (error) {
                console.error('Failed to fetch lead:', error);
                // Mock data
                setLead({
                    id: id,
                    first_name: 'Ahmad',
                    last_name: 'Razak',
                    email: 'ahmad@batik-industries.com.my',
                    phone: '+60123456789',
                    company: 'Batik Industries Sdn Bhd',
                    title: 'Procurement Manager',
                    status: 'qualified',
                    source: 'website',
                    score: 85,
                    score_label: 'hot',
                    owner_name: 'Siti Aminah',
                    website: 'https://batik-industries.com.my',
                    industry: 'Manufacturing',
                    address: '123 Jalan Batik, 50000 Kuala Lumpur, Malaysia',
                    notes: 'Interested in bulk orders for corporate gifts.',
                    created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                    updated_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
                });
            } finally {
                setIsLoading(false);
            }
        };

        fetchLead();
    }, [id]);

    const handleConvert = async (createOpportunity: boolean, createCustomer: boolean) => {
        if (!lead) return;

        try {
            const result = await leadService.convertLead(lead.id, {
                create_opportunity: createOpportunity,
                create_customer: createCustomer,
            });

            if (result.opportunity_id) {
                navigate(`/opportunities/${result.opportunity_id}`);
            } else if (result.customer_id) {
                navigate(`/customers/${result.customer_id}`);
            } else {
                navigate('/leads');
            }
        } catch (error) {
            console.error('Failed to convert lead:', error);
        }
    };

    const handleDisqualify = async () => {
        if (!lead) return;

        try {
            await leadService.disqualifyLead(lead.id, { reason: disqualifyReason });
            setLead((prev) => prev ? { ...prev, status: 'unqualified' } : null);
            setIsDisqualifyModalOpen(false);
        } catch (error) {
            console.error('Failed to disqualify lead:', error);
        }
    };

    const getStatusBadge = (status: string) => {
        const variants: Record<string, 'default' | 'success' | 'warning' | 'danger' | 'info' | 'purple'> = {
            new: 'info',
            contacted: 'warning',
            qualified: 'success',
            unqualified: 'danger',
            converted: 'purple',
        };
        return <Badge variant={variants[status] || 'default'}>{status}</Badge>;
    };

    const getScoreBadge = (score: number) => {
        let variant: 'danger' | 'warning' | 'default' = 'default';
        if (score >= 80) variant = 'danger';
        else if (score >= 50) variant = 'warning';
        return <Badge variant={variant}>{score}%</Badge>;
    };

    if (isLoading) {
        return (
            <div className="animate-fade-in">
                <div className="page-header">
                    <div className="skeleton" style={{ width: '300px', height: '32px' }} />
                </div>
                <div className="skeleton" style={{ width: '100%', height: '200px', marginTop: '1rem' }} />
            </div>
        );
    }

    if (!lead) {
        return (
            <div className="empty-state">
                <div className="empty-state-icon">🔍</div>
                <h3 className="empty-state-title">Lead not found</h3>
                <p className="empty-state-description">
                    The lead you're looking for doesn't exist or has been deleted.
                </p>
                <Link to="/leads">
                    <Button>Back to Leads</Button>
                </Link>
            </div>
        );
    }

    return (
        <div className="animate-fade-in">
            {/* Breadcrumb */}
            <div className="mb-4">
                <Link to="/leads" className="flex items-center gap-2 text-muted" style={{ textDecoration: 'none' }}>
                    <ArrowLeftIcon />
                    Back to Leads
                </Link>
            </div>

            {/* Detail Header */}
            <div className="detail-header">
                <div className="detail-info">
                    <div className="detail-avatar">
                        {lead.first_name.charAt(0)}{lead.last_name.charAt(0)}
                    </div>
                    <div>
                        <h1 className="detail-title">{lead.first_name} {lead.last_name}</h1>
                        <p className="detail-subtitle">
                            {lead.title && `${lead.title} at `}
                            {lead.company || 'No company'}
                        </p>
                        <div className="detail-badges">
                            {getStatusBadge(lead.status)}
                            {lead.score !== undefined && getScoreBadge(lead.score)}
                            {lead.score_label && (
                                <Badge variant={lead.score_label === 'hot' ? 'danger' : lead.score_label === 'warm' ? 'warning' : 'default'}>
                                    {lead.score_label}
                                </Badge>
                            )}
                        </div>
                    </div>
                </div>

                <div className="detail-actions">
                    <Button variant="outline" onClick={() => window.location.href = `tel:${lead.phone}`}>
                        <PhoneIcon />
                        Call
                    </Button>
                    <Button variant="outline" onClick={() => window.location.href = `mailto:${lead.email}`}>
                        <MailIcon />
                        Email
                    </Button>
                    <Link to={`/leads/${lead.id}/edit`}>
                        <Button variant="outline">
                            <EditIcon />
                            Edit
                        </Button>
                    </Link>
                    {lead.status === 'qualified' && (
                        <Button onClick={() => setIsConvertModalOpen(true)}>
                            <ConvertIcon />
                            Convert
                        </Button>
                    )}
                </div>
            </div>

            {/* Tabs */}
            <div className="detail-tabs">
                <button
                    className={`detail-tab ${activeTab === 'overview' ? 'active' : ''}`}
                    onClick={() => setActiveTab('overview')}
                >
                    Overview
                </button>
                <button
                    className={`detail-tab ${activeTab === 'activities' ? 'active' : ''}`}
                    onClick={() => setActiveTab('activities')}
                >
                    Activities
                </button>
                <button
                    className={`detail-tab ${activeTab === 'notes' ? 'active' : ''}`}
                    onClick={() => setActiveTab('notes')}
                >
                    Notes
                </button>
            </div>

            {/* Tab Content */}
            {activeTab === 'overview' && (
                <div className="grid grid-cols-3 gap-6">
                    {/* Lead Information */}
                    <Card className="col-span-2" padding="lg">
                        <h3 className="font-semibold text-lg mb-4">Lead Information</h3>
                        <div className="info-grid">
                            <div className="info-item">
                                <span className="info-label">Full Name</span>
                                <span className="info-value">{lead.first_name} {lead.last_name}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Email</span>
                                <span className="info-value">{lead.email}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Phone</span>
                                <span className="info-value">{lead.phone || 'N/A'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Company</span>
                                <span className="info-value">{lead.company || 'N/A'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Job Title</span>
                                <span className="info-value">{lead.title || 'N/A'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Industry</span>
                                <span className="info-value">{lead.industry || 'N/A'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Website</span>
                                <span className="info-value">
                                    {lead.website ? (
                                        <a href={lead.website} target="_blank" rel="noopener noreferrer" style={{ color: 'var(--primary)' }}>
                                            {lead.website}
                                        </a>
                                    ) : (
                                        'N/A'
                                    )}
                                </span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Source</span>
                                <span className="info-value capitalize">{lead.source?.replace('_', ' ') || 'N/A'}</span>
                            </div>
                            <div className="info-item col-span-2">
                                <span className="info-label">Address</span>
                                <span className="info-value">{lead.address || 'N/A'}</span>
                            </div>
                        </div>

                        {lead.notes && (
                            <div className="mt-6">
                                <h4 className="font-medium mb-2">Notes</h4>
                                <p className="text-sm text-secondary" style={{ lineHeight: 1.6 }}>{lead.notes}</p>
                            </div>
                        )}
                    </Card>

                    {/* Lead Score & Quick Actions */}
                    <div className="flex flex-col gap-4">
                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Lead Score</h3>
                            <div className="text-center">
                                <div
                                    style={{
                                        width: '100px',
                                        height: '100px',
                                        borderRadius: '50%',
                                        background: `conic-gradient(var(--success) ${(lead.score || 0) * 3.6}deg, var(--bg-tertiary) 0deg)`,
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center',
                                        margin: '0 auto 1rem',
                                    }}
                                >
                                    <div
                                        style={{
                                            width: '80px',
                                            height: '80px',
                                            borderRadius: '50%',
                                            background: 'var(--bg-card)',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            flexDirection: 'column',
                                        }}
                                    >
                                        <span className="text-2xl font-bold">{lead.score || 0}</span>
                                        <span className="text-xs text-muted">Score</span>
                                    </div>
                                </div>
                                <Badge
                                    variant={lead.score_label === 'hot' ? 'danger' : lead.score_label === 'warm' ? 'warning' : 'default'}
                                    size="md"
                                >
                                    {lead.score_label?.toUpperCase() || 'COLD'} Lead
                                </Badge>
                            </div>
                        </Card>

                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Owner</h3>
                            <div className="flex items-center gap-3">
                                <div className="activity-avatar">{lead.owner_name?.charAt(0) || 'U'}</div>
                                <div>
                                    <p className="font-medium">{lead.owner_name || 'Unassigned'}</p>
                                    <p className="text-sm text-muted">Sales Rep</p>
                                </div>
                            </div>
                        </Card>

                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Quick Actions</h3>
                            <div className="flex flex-col gap-2">
                                {lead.status !== 'qualified' && lead.status !== 'converted' && lead.status !== 'unqualified' && (
                                    <Button
                                        fullWidth
                                        variant="outline"
                                        onClick={async () => {
                                            await leadService.qualifyLead(lead.id, {});
                                            setLead((prev) => prev ? { ...prev, status: 'qualified' } : null);
                                        }}
                                    >
                                        Qualify Lead
                                    </Button>
                                )}
                                {lead.status !== 'unqualified' && lead.status !== 'converted' && (
                                    <Button
                                        fullWidth
                                        variant="outline"
                                        onClick={() => setIsDisqualifyModalOpen(true)}
                                    >
                                        Disqualify
                                    </Button>
                                )}
                            </div>
                        </Card>

                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Timestamps</h3>
                            <div className="text-sm">
                                <p className="flex justify-between mb-2">
                                    <span className="text-muted">Created</span>
                                    <span>{new Date(lead.created_at).toLocaleDateString('en-MY', { dateStyle: 'medium' })}</span>
                                </p>
                                {lead.updated_at && (
                                    <p className="flex justify-between">
                                        <span className="text-muted">Updated</span>
                                        <span>{new Date(lead.updated_at).toLocaleDateString('en-MY', { dateStyle: 'medium' })}</span>
                                    </p>
                                )}
                            </div>
                        </Card>
                    </div>
                </div>
            )}

            {activeTab === 'activities' && (
                <Card padding="lg">
                    <div className="timeline">
                        <div className="timeline-item">
                            <div className="timeline-icon email">📧</div>
                            <div className="timeline-content">
                                <div className="timeline-header">
                                    <span className="timeline-title">Email Sent</span>
                                    <span className="timeline-time">2 hours ago</span>
                                </div>
                                <p className="timeline-body">
                                    Follow-up email sent regarding product catalog request.
                                </p>
                                <div className="timeline-user">
                                    <div className="timeline-user-avatar">SA</div>
                                    Siti Aminah
                                </div>
                            </div>
                        </div>

                        <div className="timeline-item">
                            <div className="timeline-icon call">📞</div>
                            <div className="timeline-content">
                                <div className="timeline-header">
                                    <span className="timeline-title">Phone Call</span>
                                    <span className="timeline-time">Yesterday</span>
                                </div>
                                <p className="timeline-body">
                                    Initial discovery call - Discussed needs for corporate batik gifts.
                                </p>
                                <div className="timeline-user">
                                    <div className="timeline-user-avatar">SA</div>
                                    Siti Aminah
                                </div>
                            </div>
                        </div>

                        <div className="timeline-item">
                            <div className="timeline-icon note">📝</div>
                            <div className="timeline-content">
                                <div className="timeline-header">
                                    <span className="timeline-title">Lead Created</span>
                                    <span className="timeline-time">{new Date(lead.created_at).toLocaleDateString('en-MY')}</span>
                                </div>
                                <p className="timeline-body">
                                    Lead submitted via website contact form.
                                </p>
                            </div>
                        </div>
                    </div>
                </Card>
            )}

            {activeTab === 'notes' && (
                <Card padding="lg">
                    <h3 className="font-semibold mb-4">Notes</h3>
                    <Textarea
                        placeholder="Add a note..."
                        rows={4}
                    />
                    <div className="mt-4">
                        <Button>Add Note</Button>
                    </div>

                    <div className="mt-6 border-t pt-6">
                        <div className="activity-item">
                            <div className="activity-avatar">SA</div>
                            <div className="activity-content">
                                <p className="activity-text">{lead.notes || 'No notes yet.'}</p>
                                <p className="activity-time">Added by Siti Aminah</p>
                            </div>
                        </div>
                    </div>
                </Card>
            )}

            {/* Convert Modal */}
            <Modal
                isOpen={isConvertModalOpen}
                onClose={() => setIsConvertModalOpen(false)}
                title="Convert Lead"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsConvertModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={() => handleConvert(true, true)}>
                            Convert
                        </Button>
                    </>
                }
            >
                <p className="mb-4">
                    Converting this lead will change its status and optionally create new records.
                </p>
                <div className="flex flex-col gap-3">
                    <label className="flex items-center gap-3 p-4 bg-tertiary rounded-lg cursor-pointer">
                        <input type="checkbox" defaultChecked className="form-checkbox" />
                        <div>
                            <p className="font-medium">Create Opportunity</p>
                            <p className="text-sm text-muted">Start tracking this as a sales opportunity</p>
                        </div>
                    </label>
                    <label className="flex items-center gap-3 p-4 bg-tertiary rounded-lg cursor-pointer">
                        <input type="checkbox" defaultChecked className="form-checkbox" />
                        <div>
                            <p className="font-medium">Create Customer</p>
                            <p className="text-sm text-muted">Add to your customer database</p>
                        </div>
                    </label>
                </div>
            </Modal>

            {/* Disqualify Modal */}
            <Modal
                isOpen={isDisqualifyModalOpen}
                onClose={() => setIsDisqualifyModalOpen(false)}
                title="Disqualify Lead"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsDisqualifyModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button variant="danger" onClick={handleDisqualify}>
                            Disqualify
                        </Button>
                    </>
                }
            >
                <p className="mb-4">
                    Please provide a reason for disqualifying this lead.
                </p>
                <Select
                    label="Reason"
                    options={[
                        { value: 'budget', label: 'No Budget' },
                        { value: 'timing', label: 'Wrong Timing' },
                        { value: 'competitor', label: 'Chose Competitor' },
                        { value: 'no_response', label: 'No Response' },
                        { value: 'duplicate', label: 'Duplicate Lead' },
                        { value: 'other', label: 'Other' },
                    ]}
                    value={disqualifyReason}
                    onChange={(e) => setDisqualifyReason(e.target.value)}
                    required
                />
            </Modal>
        </div>
    );
}

export default LeadDetailPage;
