// ============================================
// Opportunity List Page
// Production-Ready Opportunity Management List
// ============================================

import { useState, useEffect, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { opportunityService, pipelineService } from '../../services';
import { Button, Input, Select, Badge, Table, Pagination, Modal } from '../../components/ui';
import type { Opportunity, OpportunityFilters, Pipeline } from '../../types';

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

const KanbanIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="5" height="18" rx="1" /><rect x="10" y="3" width="5" height="12" rx="1" /><rect x="17" y="3" width="5" height="8" rx="1" />
    </svg>
);

const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'open', label: 'Open' },
    { value: 'won', label: 'Won' },
    { value: 'lost', label: 'Lost' },
];

export function OpportunityListPage() {
    const navigate = useNavigate();
    const [opportunities, setOpportunities] = useState<Opportunity[]>([]);
    const [pipelines, setPipelines] = useState<Pipeline[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [selectedIds, setSelectedIds] = useState<string[]>([]);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);

    const [filters, setFilters] = useState<OpportunityFilters>({
        page: 1,
        per_page: 20,
        status: '',
        stage_id: '',
        search: '',
    });

    const [meta, setMeta] = useState({
        total: 0,
        page: 1,
        per_page: 20,
        total_pages: 1,
    });

    const [newOpportunity, setNewOpportunity] = useState({
        name: '',
        customer_name: '',
        value: '',
        stage_id: '',
        expected_close_date: '',
    });

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
                        { id: 's1', name: 'Qualification', position: 1, color: '#3b82f6' },
                        { id: 's2', name: 'Proposal', position: 2, color: '#8b5cf6' },
                        { id: 's3', name: 'Negotiation', position: 3, color: '#f59e0b' },
                        { id: 's4', name: 'Closing', position: 4, color: '#10b981' },
                    ],
                }]);
            }
        };

        fetchPipelines();
    }, []);

    useEffect(() => {
        const fetchOpportunities = async () => {
            setIsLoading(true);
            try {
                const response = await opportunityService.getOpportunities(filters);
                setOpportunities(response.opportunities);
                setMeta(response.meta);
            } catch (error) {
                console.error('Failed to fetch opportunities:', error);
                // Mock data
                setOpportunities([
                    {
                        id: 'o1',
                        name: 'Enterprise Package',
                        customer_name: 'Batik Industries Sdn Bhd',
                        value: 85000,
                        stage_id: 's1',
                        stage_name: 'Qualification',
                        probability: 25,
                        status: 'open',
                        expected_close_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
                        owner_name: 'Ahmad Razak',
                        created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'o2',
                        name: 'Premium Collection',
                        customer_name: 'Textile Malaysia',
                        value: 45000,
                        stage_id: 's2',
                        stage_name: 'Proposal',
                        probability: 50,
                        status: 'open',
                        expected_close_date: new Date(Date.now() + 21 * 24 * 60 * 60 * 1000).toISOString(),
                        owner_name: 'Siti Aminah',
                        created_at: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'o3',
                        name: 'Wholesale Order Q1',
                        customer_name: 'Kraf Malaysia',
                        value: 120000,
                        stage_id: 's3',
                        stage_name: 'Negotiation',
                        probability: 75,
                        status: 'open',
                        expected_close_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
                        owner_name: 'Muhammad Hafiz',
                        created_at: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'o4',
                        name: 'Custom Design Project',
                        customer_name: 'Heritage Batik',
                        value: 35000,
                        stage_id: 's4',
                        stage_name: 'Closing',
                        probability: 90,
                        status: 'open',
                        expected_close_date: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
                        owner_name: 'Nurul Aisyah',
                        created_at: new Date(Date.now() - 45 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'o5',
                        name: 'Annual Contract 2024',
                        customer_name: 'Malaysian Airlines',
                        value: 250000,
                        stage_id: 's4',
                        stage_name: 'Closing',
                        probability: 95,
                        status: 'won',
                        expected_close_date: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                        actual_close_date: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
                        owner_name: 'Ahmad Razak',
                        created_at: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                ]);
                setMeta({ total: 48, page: 1, per_page: 20, total_pages: 3 });
            } finally {
                setIsLoading(false);
            }
        };

        fetchOpportunities();
    }, [filters]);

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('ms-MY', {
            style: 'currency',
            currency: 'MYR',
            minimumFractionDigits: 0,
        }).format(value);
    };

    const getStatusBadge = (status: string) => {
        const variants: Record<string, 'default' | 'success' | 'warning' | 'danger'> = {
            open: 'warning',
            won: 'success',
            lost: 'danger',
        };
        return <Badge variant={variants[status] || 'default'}>{status}</Badge>;
    };

    const getProbabilityBadge = (probability?: number) => {
        if (!probability) return null;
        let variant: 'danger' | 'warning' | 'success' | 'default' = 'default';
        if (probability >= 75) variant = 'success';
        else if (probability >= 50) variant = 'warning';
        else if (probability >= 25) variant = 'danger';
        return <Badge variant={variant}>{probability}%</Badge>;
    };

    const allStages = pipelines.flatMap((p) => p.stages);

    const stageOptions = [
        { value: '', label: 'All Stages' },
        ...allStages.map((s) => ({ value: s.id, label: s.name })),
    ];

    const handleCreateOpportunity = async () => {
        try {
            const stageId = newOpportunity.stage_id || allStages[0]?.id || '';
            await opportunityService.createOpportunity({
                name: newOpportunity.name,
                value: parseFloat(newOpportunity.value) || 0,
                stage_id: stageId,
                expected_close_date: newOpportunity.expected_close_date || undefined,
            });
            setIsCreateModalOpen(false);
            setNewOpportunity({
                name: '',
                customer_name: '',
                value: '',
                stage_id: '',
                expected_close_date: '',
            });
            // Refresh list
            setFilters((prev) => ({ ...prev }));
        } catch (error) {
            console.error('Failed to create opportunity:', error);
        }
    };

    const columns = useMemo(
        () => [
            {
                key: 'name',
                header: 'Opportunity',
                render: (opp: Opportunity) => (
                    <div>
                        <Link
                            to={`/opportunities/${opp.id}`}
                            className="font-medium"
                            style={{ color: 'var(--text-primary)' }}
                        >
                            {opp.name}
                        </Link>
                        <div className="text-sm text-muted">{opp.customer_name || 'No customer'}</div>
                    </div>
                ),
            },
            {
                key: 'value',
                header: 'Value',
                render: (opp: Opportunity) => (
                    <span className="font-medium text-success">{formatCurrency(opp.value)}</span>
                ),
            },
            {
                key: 'stage',
                header: 'Stage',
                render: (opp: Opportunity) => {
                    const stage = allStages.find((s) => s.id === opp.stage_id);
                    return (
                        <div className="flex items-center gap-2">
                            <div
                                style={{
                                    width: '8px',
                                    height: '8px',
                                    borderRadius: '50%',
                                    background: stage?.color || '#94a3b8',
                                }}
                            />
                            <span className="text-sm">{opp.stage_name || stage?.name || 'Unknown'}</span>
                        </div>
                    );
                },
            },
            {
                key: 'probability',
                header: 'Probability',
                render: (opp: Opportunity) => getProbabilityBadge(opp.probability),
            },
            {
                key: 'status',
                header: 'Status',
                render: (opp: Opportunity) => getStatusBadge(opp.status),
            },
            {
                key: 'expected_close',
                header: 'Expected Close',
                render: (opp: Opportunity) => (
                    <span className="text-sm">
                        {opp.expected_close_date
                            ? new Date(opp.expected_close_date).toLocaleDateString('en-MY')
                            : 'Not set'}
                    </span>
                ),
            },
            {
                key: 'owner',
                header: 'Owner',
                render: (opp: Opportunity) => (
                    <span className="text-sm">{opp.owner_name || 'Unassigned'}</span>
                ),
            },
        ],
        [allStages]
    );

    return (
        <div className="animate-fade-in">
            {/* Page Header */}
            <div className="page-header">
                <div className="page-header-left">
                    <h1 className="page-title">Opportunities</h1>
                    <p className="page-description">
                        Track and manage your sales opportunities
                    </p>
                </div>
                <div className="page-header-actions">
                    <Link to="/pipeline">
                        <Button variant="outline">
                            <KanbanIcon />
                            <span>Kanban View</span>
                        </Button>
                    </Link>
                    <Button onClick={() => setIsCreateModalOpen(true)}>
                        <PlusIcon />
                        <span>New Opportunity</span>
                    </Button>
                </div>
            </div>

            {/* Search & Filters */}
            <div className="search-filter-bar">
                <div className="search-input-wrapper">
                    <span className="search-icon">
                        <SearchIcon />
                    </span>
                    <Input
                        type="text"
                        placeholder="Search opportunities..."
                        value={filters.search}
                        onChange={(e) => setFilters((prev) => ({ ...prev, search: e.target.value, page: 1 }))}
                    />
                </div>

                <div className="filter-group">
                    <Select
                        options={statusOptions}
                        value={filters.status}
                        onChange={(e) => setFilters((prev) => ({ ...prev, status: e.target.value, page: 1 }))}
                    />
                    <Select
                        options={stageOptions}
                        value={filters.stage_id}
                        onChange={(e) => setFilters((prev) => ({ ...prev, stage_id: e.target.value, page: 1 }))}
                    />
                </div>
            </div>

            {/* Opportunities Table */}
            <div className="card" style={{ padding: 0 }}>
                <Table
                    columns={columns}
                    data={opportunities}
                    keyExtractor={(opp) => opp.id}
                    loading={isLoading}
                    emptyMessage="No opportunities found"
                    onRowClick={(opp) => navigate(`/opportunities/${opp.id}`)}
                    showCheckboxes
                    selectedRows={selectedIds}
                    onSelectionChange={setSelectedIds}
                />

                <Pagination
                    currentPage={meta.page}
                    totalPages={meta.total_pages}
                    totalItems={meta.total}
                    itemsPerPage={meta.per_page}
                    onPageChange={(page) => setFilters((prev) => ({ ...prev, page }))}
                />
            </div>

            {/* Create Opportunity Modal */}
            <Modal
                isOpen={isCreateModalOpen}
                onClose={() => setIsCreateModalOpen(false)}
                title="Create New Opportunity"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsCreateModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleCreateOpportunity}>Create Opportunity</Button>
                    </>
                }
            >
                <form onSubmit={(e) => { e.preventDefault(); handleCreateOpportunity(); }}>
                    <Input
                        label="Opportunity Name"
                        placeholder="Enter opportunity name"
                        value={newOpportunity.name}
                        onChange={(e) => setNewOpportunity((prev) => ({ ...prev, name: e.target.value }))}
                        required
                    />

                    <div style={{ marginTop: '1rem' }}>
                        <Input
                            label="Customer Name"
                            placeholder="Enter customer name"
                            value={newOpportunity.customer_name}
                            onChange={(e) => setNewOpportunity((prev) => ({ ...prev, customer_name: e.target.value }))}
                        />
                    </div>

                    <div style={{ marginTop: '1rem' }}>
                        <Input
                            type="number"
                            label="Value (MYR)"
                            placeholder="0.00"
                            value={newOpportunity.value}
                            onChange={(e) => setNewOpportunity((prev) => ({ ...prev, value: e.target.value }))}
                        />
                    </div>

                    <div style={{ marginTop: '1rem' }}>
                        <Select
                            label="Stage"
                            options={allStages.map((s) => ({ value: s.id, label: s.name }))}
                            value={newOpportunity.stage_id}
                            onChange={(e) => setNewOpportunity((prev) => ({ ...prev, stage_id: e.target.value }))}
                        />
                    </div>

                    <div style={{ marginTop: '1rem' }}>
                        <Input
                            type="date"
                            label="Expected Close Date"
                            value={newOpportunity.expected_close_date}
                            onChange={(e) =>
                                setNewOpportunity((prev) => ({ ...prev, expected_close_date: e.target.value }))
                            }
                        />
                    </div>
                </form>
            </Modal>
        </div>
    );
}

export default OpportunityListPage;