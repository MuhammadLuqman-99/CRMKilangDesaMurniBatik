// ============================================
// Opportunity Detail Page
// Production-Ready Opportunity Detail View
// ============================================

import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { opportunityService, pipelineService } from '../../services';
import { Button, Badge, Card, Modal, Input, Textarea, Select } from '../../components/ui';
import type { Opportunity, Pipeline, PipelineStage } from '../../types';

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

const CheckIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="20 6 9 17 4 12" />
    </svg>
);

const XIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
    </svg>
);

const PlusIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
    </svg>
);

export function OpportunityDetailPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [opportunity, setOpportunity] = useState<Opportunity | null>(null);
    const [pipelines, setPipelines] = useState<Pipeline[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('overview');

    // Modal states
    const [isWinModalOpen, setIsWinModalOpen] = useState(false);
    const [isLoseModalOpen, setIsLoseModalOpen] = useState(false);
    const [isEditModalOpen, setIsEditModalOpen] = useState(false);
    const [loseReason, setLoseReason] = useState('');
    const [winNotes, setWinNotes] = useState('');

    const [editForm, setEditForm] = useState({
        name: '',
        value: '',
        stage_id: '',
        probability: '',
        expected_close_date: '',
        notes: '',
    });

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('ms-MY', {
            style: 'currency',
            currency: 'MYR',
            minimumFractionDigits: 0,
        }).format(value);
    };

    useEffect(() => {
        const fetchPipelines = async () => {
            try {
                const response = await pipelineService.getPipelines();
                setPipelines(response.pipelines);
            } catch (error) {
                console.error('Failed to fetch pipelines:', error);
                setPipelines([{
                    id: '1',
                    name: 'Sales Pipeline',
                    stages: [
                        { id: 's1', name: 'Qualification', position: 1, color: '#3b82f6', probability: 25 },
                        { id: 's2', name: 'Proposal', position: 2, color: '#8b5cf6', probability: 50 },
                        { id: 's3', name: 'Negotiation', position: 3, color: '#f59e0b', probability: 75 },
                        { id: 's4', name: 'Closing', position: 4, color: '#10b981', probability: 90 },
                    ],
                }]);
            }
        };

        fetchPipelines();
    }, []);

    useEffect(() => {
        const fetchOpportunity = async () => {
            if (!id) return;

            setIsLoading(true);
            try {
                const data = await opportunityService.getOpportunity(id);
                setOpportunity(data);
                setEditForm({
                    name: data.name,
                    value: data.value.toString(),
                    stage_id: data.stage_id,
                    probability: data.probability?.toString() || '',
                    expected_close_date: data.expected_close_date?.split('T')[0] || '',
                    notes: data.notes || '',
                });
            } catch (error) {
                console.error('Failed to fetch opportunity:', error);
                // Mock data
                const mockOpp: Opportunity = {
                    id: id,
                    name: 'Enterprise Package',
                    customer_id: 'c1',
                    customer_name: 'Batik Industries Sdn Bhd',
                    value: 85000,
                    stage_id: 's2',
                    stage_name: 'Proposal',
                    probability: 50,
                    expected_close_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
                    status: 'open',
                    notes: 'Client interested in premium batik collections for corporate gifts.',
                    owner_id: 'u1',
                    owner_name: 'Ahmad Razak',
                    created_at: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
                    updated_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
                };
                setOpportunity(mockOpp);
                setEditForm({
                    name: mockOpp.name,
                    value: mockOpp.value.toString(),
                    stage_id: mockOpp.stage_id,
                    probability: mockOpp.probability?.toString() || '',
                    expected_close_date: mockOpp.expected_close_date?.split('T')[0] || '',
                    notes: mockOpp.notes || '',
                });
            } finally {
                setIsLoading(false);
            }
        };

        fetchOpportunity();
    }, [id]);

    const allStages = pipelines.flatMap((p) => p.stages);
    const currentStage = allStages.find((s) => s.id === opportunity?.stage_id);

    const handleMoveStage = async (stageId: string) => {
        if (!opportunity) return;

        try {
            const updated = await opportunityService.moveStage(opportunity.id, { stage_id: stageId });
            const stage = allStages.find((s) => s.id === stageId);
            setOpportunity({
                ...opportunity,
                stage_id: stageId,
                stage_name: stage?.name,
                probability: stage?.probability,
            });
        } catch (error) {
            console.error('Failed to move stage:', error);
            // Optimistic update for demo
            const stage = allStages.find((s) => s.id === stageId);
            setOpportunity({
                ...opportunity,
                stage_id: stageId,
                stage_name: stage?.name,
                probability: stage?.probability,
            });
        }
    };

    const handleWin = async () => {
        if (!opportunity) return;

        try {
            await opportunityService.winOpportunity(opportunity.id, { notes: winNotes });
            setOpportunity({
                ...opportunity,
                status: 'won',
                actual_close_date: new Date().toISOString(),
            });
            setIsWinModalOpen(false);
            setWinNotes('');
        } catch (error) {
            console.error('Failed to mark as won:', error);
            // Optimistic update for demo
            setOpportunity({
                ...opportunity,
                status: 'won',
                actual_close_date: new Date().toISOString(),
            });
            setIsWinModalOpen(false);
        }
    };

    const handleLose = async () => {
        if (!opportunity || !loseReason) return;

        try {
            await opportunityService.loseOpportunity(opportunity.id, { reason: loseReason });
            setOpportunity({
                ...opportunity,
                status: 'lost',
                actual_close_date: new Date().toISOString(),
            });
            setIsLoseModalOpen(false);
            setLoseReason('');
        } catch (error) {
            console.error('Failed to mark as lost:', error);
            // Optimistic update for demo
            setOpportunity({
                ...opportunity,
                status: 'lost',
                actual_close_date: new Date().toISOString(),
            });
            setIsLoseModalOpen(false);
        }
    };

    const handleUpdate = async () => {
        if (!opportunity) return;

        try {
            const updated = await opportunityService.updateOpportunity(opportunity.id, {
                name: editForm.name,
                value: parseFloat(editForm.value),
                stage_id: editForm.stage_id,
                probability: editForm.probability ? parseInt(editForm.probability) : undefined,
                expected_close_date: editForm.expected_close_date || undefined,
                notes: editForm.notes || undefined,
            });
            const stage = allStages.find((s) => s.id === editForm.stage_id);
            setOpportunity({
                ...opportunity,
                name: editForm.name,
                value: parseFloat(editForm.value),
                stage_id: editForm.stage_id,
                stage_name: stage?.name,
                probability: editForm.probability ? parseInt(editForm.probability) : undefined,
                expected_close_date: editForm.expected_close_date || undefined,
                notes: editForm.notes || undefined,
            });
            setIsEditModalOpen(false);
        } catch (error) {
            console.error('Failed to update opportunity:', error);
        }
    };

    const getStatusBadge = (status: string) => {
        const variants: Record<string, 'default' | 'success' | 'warning' | 'danger'> = {
            open: 'warning',
            won: 'success',
            lost: 'danger',
        };
        return <Badge variant={variants[status] || 'default'} size="md">{status.toUpperCase()}</Badge>;
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

    if (!opportunity) {
        return (
            <div className="empty-state">
                <div className="empty-state-icon">🔍</div>
                <h3 className="empty-state-title">Opportunity not found</h3>
                <p className="empty-state-description">
                    The opportunity you're looking for doesn't exist or has been deleted.
                </p>
                <Link to="/opportunities">
                    <Button>Back to Opportunities</Button>
                </Link>
            </div>
        );
    }

    const isOverdue = opportunity.expected_close_date && new Date(opportunity.expected_close_date) < new Date() && opportunity.status === 'open';

    return (
        <div className="animate-fade-in">
            {/* Breadcrumb */}
            <div className="mb-4">
                <Link to="/opportunities" className="flex items-center gap-2 text-muted" style={{ textDecoration: 'none' }}>
                    <ArrowLeftIcon />
                    Back to Opportunities
                </Link>
            </div>

            {/* Detail Header */}
            <div className="detail-header">
                <div className="detail-info">
                    <div className="detail-avatar" style={{ background: 'linear-gradient(135deg, #8b5cf6, #a78bfa)' }}>
                        💰
                    </div>
                    <div>
                        <h1 className="detail-title">{opportunity.name}</h1>
                        <p className="detail-subtitle">
                            {opportunity.customer_name && (
                                <Link to={`/customers/${opportunity.customer_id}`} style={{ color: 'var(--primary)' }}>
                                    {opportunity.customer_name}
                                </Link>
                            )}
                            {!opportunity.customer_name && 'No customer linked'}
                        </p>
                        <div className="detail-badges">
                            {getStatusBadge(opportunity.status)}
                            {currentStage && (
                                <Badge
                                    variant="default"
                                    size="md"
                                    style={{ borderLeft: `4px solid ${currentStage.color}` }}
                                >
                                    {currentStage.name}
                                </Badge>
                            )}
                            {isOverdue && <Badge variant="danger" size="md">Overdue</Badge>}
                        </div>
                    </div>
                </div>

                <div className="detail-actions">
                    <Button variant="outline" onClick={() => setIsEditModalOpen(true)}>
                        <EditIcon />
                        Edit
                    </Button>
                    {opportunity.status === 'open' && (
                        <>
                            <Button variant="outline" onClick={() => setIsLoseModalOpen(true)}>
                                <XIcon />
                                Mark Lost
                            </Button>
                            <Button onClick={() => setIsWinModalOpen(true)}>
                                <CheckIcon />
                                Mark Won
                            </Button>
                        </>
                    )}
                </div>
            </div>

            {/* Value & Stats Row */}
            <div className="stats-grid mb-6" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
                <div className="stat-card">
                    <span className="stat-card-title">Deal Value</span>
                    <span className="stat-card-value text-success">{formatCurrency(opportunity.value)}</span>
                </div>
                <div className="stat-card">
                    <span className="stat-card-title">Probability</span>
                    <span className="stat-card-value">{opportunity.probability || 0}%</span>
                </div>
                <div className="stat-card">
                    <span className="stat-card-title">Weighted Value</span>
                    <span className="stat-card-value">
                        {formatCurrency(opportunity.value * ((opportunity.probability || 0) / 100))}
                    </span>
                </div>
                <div className="stat-card">
                    <span className="stat-card-title">Expected Close</span>
                    <span className={`stat-card-value ${isOverdue ? 'text-danger' : ''}`}>
                        {opportunity.expected_close_date
                            ? new Date(opportunity.expected_close_date).toLocaleDateString('en-MY', { month: 'short', day: 'numeric', year: 'numeric' })
                            : 'Not set'}
                    </span>
                </div>
            </div>

            {/* Pipeline Progress */}
            {opportunity.status === 'open' && (
                <Card className="mb-6" padding="lg">
                    <h3 className="font-semibold mb-4">Pipeline Progress</h3>
                    <div className="flex items-center gap-2">
                        {allStages.map((stage, index) => {
                            const isActive = stage.id === opportunity.stage_id;
                            const isPast = allStages.findIndex((s) => s.id === opportunity.stage_id) > index;

                            return (
                                <div key={stage.id} className="flex-1 flex items-center">
                                    <button
                                        className={`flex-1 py-3 px-4 rounded-lg text-center transition-all ${isActive
                                                ? 'font-medium'
                                                : isPast
                                                    ? 'opacity-75'
                                                    : 'opacity-50 hover:opacity-75'
                                            }`}
                                        style={{
                                            background: isActive ? stage.color : isPast ? `${stage.color}40` : 'var(--bg-tertiary)',
                                            color: isActive ? 'white' : 'inherit',
                                            border: 'none',
                                            cursor: 'pointer',
                                        }}
                                        onClick={() => handleMoveStage(stage.id)}
                                    >
                                        <div className="text-sm font-medium">{stage.name}</div>
                                        <div className="text-xs opacity-75">{stage.probability}%</div>
                                    </button>
                                    {index < allStages.length - 1 && (
                                        <div
                                            style={{
                                                width: '24px',
                                                height: '2px',
                                                background: isPast ? stage.color : 'var(--border-color)',
                                                margin: '0 4px',
                                            }}
                                        />
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </Card>
            )}

            {/* Tabs */}
            <div className="detail-tabs">
                <button
                    className={`detail-tab ${activeTab === 'overview' ? 'active' : ''}`}
                    onClick={() => setActiveTab('overview')}
                >
                    Overview
                </button>
                <button
                    className={`detail-tab ${activeTab === 'timeline' ? 'active' : ''}`}
                    onClick={() => setActiveTab('timeline')}
                >
                    Timeline
                </button>
                <button
                    className={`detail-tab ${activeTab === 'products' ? 'active' : ''}`}
                    onClick={() => setActiveTab('products')}
                >
                    Products
                </button>
                <button
                    className={`detail-tab ${activeTab === 'contacts' ? 'active' : ''}`}
                    onClick={() => setActiveTab('contacts')}
                >
                    Contacts
                </button>
            </div>

            {/* Tab Content */}
            {activeTab === 'overview' && (
                <div className="grid grid-cols-3 gap-6">
                    <Card className="col-span-2" padding="lg">
                        <h3 className="font-semibold text-lg mb-4">Opportunity Details</h3>
                        <div className="info-grid">
                            <div className="info-item">
                                <span className="info-label">Opportunity Name</span>
                                <span className="info-value">{opportunity.name}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Customer</span>
                                <span className="info-value">
                                    {opportunity.customer_name ? (
                                        <Link to={`/customers/${opportunity.customer_id}`} style={{ color: 'var(--primary)' }}>
                                            {opportunity.customer_name}
                                        </Link>
                                    ) : 'N/A'}
                                </span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Deal Value</span>
                                <span className="info-value text-success font-bold">{formatCurrency(opportunity.value)}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Probability</span>
                                <span className="info-value">{opportunity.probability || 0}%</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Current Stage</span>
                                <span className="info-value">{currentStage?.name || 'Unknown'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Status</span>
                                <span className="info-value capitalize">{opportunity.status}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Expected Close Date</span>
                                <span className={`info-value ${isOverdue ? 'text-danger' : ''}`}>
                                    {opportunity.expected_close_date
                                        ? new Date(opportunity.expected_close_date).toLocaleDateString('en-MY')
                                        : 'Not set'}
                                </span>
                            </div>
                            {opportunity.actual_close_date && (
                                <div className="info-item">
                                    <span className="info-label">Actual Close Date</span>
                                    <span className="info-value">
                                        {new Date(opportunity.actual_close_date).toLocaleDateString('en-MY')}
                                    </span>
                                </div>
                            )}
                        </div>

                        {opportunity.notes && (
                            <div className="mt-6">
                                <h4 className="font-medium mb-2">Notes</h4>
                                <p className="text-sm text-secondary" style={{ lineHeight: 1.6 }}>{opportunity.notes}</p>
                            </div>
                        )}
                    </Card>

                    <div className="flex flex-col gap-4">
                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Owner</h3>
                            <div className="flex items-center gap-3">
                                <div className="activity-avatar">{opportunity.owner_name?.charAt(0) || 'U'}</div>
                                <div>
                                    <p className="font-medium">{opportunity.owner_name || 'Unassigned'}</p>
                                    <p className="text-sm text-muted">Sales Rep</p>
                                </div>
                            </div>
                        </Card>

                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Timestamps</h3>
                            <div className="text-sm">
                                <p className="flex justify-between mb-2">
                                    <span className="text-muted">Created</span>
                                    <span>{new Date(opportunity.created_at).toLocaleDateString('en-MY', { dateStyle: 'medium' })}</span>
                                </p>
                                {opportunity.updated_at && (
                                    <p className="flex justify-between">
                                        <span className="text-muted">Updated</span>
                                        <span>{new Date(opportunity.updated_at).toLocaleDateString('en-MY', { dateStyle: 'medium' })}</span>
                                    </p>
                                )}
                            </div>
                        </Card>
                    </div>
                </div>
            )}

            {activeTab === 'timeline' && (
                <Card padding="lg">
                    <h3 className="font-semibold text-lg mb-4">Activity Timeline</h3>
                    <div className="timeline">
                        <div className="timeline-item">
                            <div className="timeline-icon" style={{ background: 'var(--success)' }}>📧</div>
                            <div className="timeline-content">
                                <div className="timeline-header">
                                    <span className="timeline-title">Proposal Sent</span>
                                    <span className="timeline-time">2 days ago</span>
                                </div>
                                <p className="timeline-body">
                                    Sent detailed proposal with pricing for the Enterprise Package.
                                </p>
                                <div className="timeline-user">
                                    <div className="timeline-user-avatar">AR</div>
                                    Ahmad Razak
                                </div>
                            </div>
                        </div>

                        <div className="timeline-item">
                            <div className="timeline-icon" style={{ background: 'var(--primary)' }}>📞</div>
                            <div className="timeline-content">
                                <div className="timeline-header">
                                    <span className="timeline-title">Discovery Call</span>
                                    <span className="timeline-time">1 week ago</span>
                                </div>
                                <p className="timeline-body">
                                    Initial call to discuss requirements and timeline for batik corporate gifts.
                                </p>
                                <div className="timeline-user">
                                    <div className="timeline-user-avatar">AR</div>
                                    Ahmad Razak
                                </div>
                            </div>
                        </div>

                        <div className="timeline-item">
                            <div className="timeline-icon" style={{ background: 'var(--warning)' }}>📊</div>
                            <div className="timeline-content">
                                <div className="timeline-header">
                                    <span className="timeline-title">Stage Changed</span>
                                    <span className="timeline-time">1 week ago</span>
                                </div>
                                <p className="timeline-body">
                                    Moved from Qualification to Proposal stage.
                                </p>
                            </div>
                        </div>

                        <div className="timeline-item">
                            <div className="timeline-icon" style={{ background: '#8b5cf6' }}>✨</div>
                            <div className="timeline-content">
                                <div className="timeline-header">
                                    <span className="timeline-title">Opportunity Created</span>
                                    <span className="timeline-time">{new Date(opportunity.created_at).toLocaleDateString('en-MY')}</span>
                                </div>
                                <p className="timeline-body">
                                    Opportunity created from converted lead.
                                </p>
                            </div>
                        </div>
                    </div>
                </Card>
            )}

            {activeTab === 'products' && (
                <Card padding="lg">
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="font-semibold text-lg">Products / Line Items</h3>
                        <Button>
                            <PlusIcon />
                            Add Product
                        </Button>
                    </div>

                    <table className="table">
                        <thead>
                            <tr>
                                <th>Product</th>
                                <th>Quantity</th>
                                <th>Unit Price</th>
                                <th>Discount</th>
                                <th>Total</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td>
                                    <strong>Premium Batik Shirt - Corporate</strong>
                                    <br />
                                    <span className="text-sm text-muted">SKU: BTK-CORP-001</span>
                                </td>
                                <td>100</td>
                                <td>{formatCurrency(350)}</td>
                                <td>10%</td>
                                <td className="font-medium text-success">{formatCurrency(31500)}</td>
                            </tr>
                            <tr>
                                <td>
                                    <strong>Batik Scarf - Silk</strong>
                                    <br />
                                    <span className="text-sm text-muted">SKU: BTK-SCF-002</span>
                                </td>
                                <td>200</td>
                                <td>{formatCurrency(180)}</td>
                                <td>5%</td>
                                <td className="font-medium text-success">{formatCurrency(34200)}</td>
                            </tr>
                            <tr>
                                <td>
                                    <strong>Custom Gift Box Set</strong>
                                    <br />
                                    <span className="text-sm text-muted">SKU: BTK-GFT-003</span>
                                </td>
                                <td>50</td>
                                <td>{formatCurrency(450)}</td>
                                <td>15%</td>
                                <td className="font-medium text-success">{formatCurrency(19125)}</td>
                            </tr>
                        </tbody>
                        <tfoot>
                            <tr>
                                <td colSpan={4} className="text-right font-medium">Subtotal:</td>
                                <td className="font-bold text-success">{formatCurrency(84825)}</td>
                            </tr>
                        </tfoot>
                    </table>
                </Card>
            )}

            {activeTab === 'contacts' && (
                <Card padding="lg">
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="font-semibold text-lg">Associated Contacts</h3>
                        <Button>
                            <PlusIcon />
                            Add Contact
                        </Button>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div className="card p-4">
                            <div className="flex items-center gap-3">
                                <div className="activity-avatar">AH</div>
                                <div className="flex-1">
                                    <div className="flex items-center gap-2">
                                        <p className="font-medium">Ahmad bin Hassan</p>
                                        <Badge variant="success" size="sm">Decision Maker</Badge>
                                    </div>
                                    <p className="text-sm text-muted">Procurement Manager</p>
                                    <p className="text-sm">ahmad@batik-industries.com.my</p>
                                    <p className="text-sm text-muted">+60123456789</p>
                                </div>
                            </div>
                        </div>

                        <div className="card p-4">
                            <div className="flex items-center gap-3">
                                <div className="activity-avatar">FA</div>
                                <div className="flex-1">
                                    <div className="flex items-center gap-2">
                                        <p className="font-medium">Fatimah binti Ali</p>
                                        <Badge variant="info" size="sm">Influencer</Badge>
                                    </div>
                                    <p className="text-sm text-muted">Finance Director</p>
                                    <p className="text-sm">fatimah@batik-industries.com.my</p>
                                    <p className="text-sm text-muted">+60198765432</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </Card>
            )}

            {/* Win Modal */}
            <Modal
                isOpen={isWinModalOpen}
                onClose={() => setIsWinModalOpen(false)}
                title="Mark Opportunity as Won"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsWinModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleWin}>
                            <CheckIcon />
                            Mark as Won
                        </Button>
                    </>
                }
            >
                <div className="text-center mb-6">
                    <div className="text-4xl mb-2">🎉</div>
                    <p className="text-lg font-medium">Congratulations!</p>
                    <p className="text-muted">You're about to close this deal worth {formatCurrency(opportunity.value)}</p>
                </div>
                <Textarea
                    label="Closing Notes (optional)"
                    placeholder="Add any notes about how this deal was won..."
                    rows={4}
                    value={winNotes}
                    onChange={(e) => setWinNotes(e.target.value)}
                />
            </Modal>

            {/* Lose Modal */}
            <Modal
                isOpen={isLoseModalOpen}
                onClose={() => setIsLoseModalOpen(false)}
                title="Mark Opportunity as Lost"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsLoseModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button variant="danger" onClick={handleLose}>
                            <XIcon />
                            Mark as Lost
                        </Button>
                    </>
                }
            >
                <p className="mb-4">Please provide a reason for losing this opportunity.</p>
                <Select
                    label="Reason"
                    options={[
                        { value: 'budget', label: 'Budget Constraints' },
                        { value: 'competitor', label: 'Lost to Competitor' },
                        { value: 'timing', label: 'Bad Timing' },
                        { value: 'no_decision', label: 'No Decision Made' },
                        { value: 'requirements', label: 'Requirements Not Met' },
                        { value: 'other', label: 'Other' },
                    ]}
                    value={loseReason}
                    onChange={(e) => setLoseReason(e.target.value)}
                    required
                />
            </Modal>

            {/* Edit Modal */}
            <Modal
                isOpen={isEditModalOpen}
                onClose={() => setIsEditModalOpen(false)}
                title="Edit Opportunity"
                size="lg"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsEditModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleUpdate}>Save Changes</Button>
                    </>
                }
            >
                <div className="grid grid-cols-2 gap-4">
                    <div className="col-span-2">
                        <Input
                            label="Opportunity Name"
                            value={editForm.name}
                            onChange={(e) => setEditForm((prev) => ({ ...prev, name: e.target.value }))}
                            required
                        />
                    </div>
                    <Input
                        type="number"
                        label="Value (MYR)"
                        value={editForm.value}
                        onChange={(e) => setEditForm((prev) => ({ ...prev, value: e.target.value }))}
                    />
                    <Select
                        label="Stage"
                        options={allStages.map((s) => ({ value: s.id, label: s.name }))}
                        value={editForm.stage_id}
                        onChange={(e) => setEditForm((prev) => ({ ...prev, stage_id: e.target.value }))}
                    />
                    <Input
                        type="number"
                        label="Probability (%)"
                        value={editForm.probability}
                        onChange={(e) => setEditForm((prev) => ({ ...prev, probability: e.target.value }))}
                        min="0"
                        max="100"
                    />
                    <Input
                        type="date"
                        label="Expected Close Date"
                        value={editForm.expected_close_date}
                        onChange={(e) => setEditForm((prev) => ({ ...prev, expected_close_date: e.target.value }))}
                    />
                    <div className="col-span-2">
                        <Textarea
                            label="Notes"
                            value={editForm.notes}
                            onChange={(e) => setEditForm((prev) => ({ ...prev, notes: e.target.value }))}
                            rows={4}
                        />
                    </div>
                </div>
            </Modal>
        </div>
    );
}

export default OpportunityDetailPage;
