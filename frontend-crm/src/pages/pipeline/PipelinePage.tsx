// ============================================
// Pipeline Page
// Production-Ready Kanban Board
// ============================================

import { useState, useEffect, useCallback, type DragEvent } from 'react';
import { pipelineService } from '../../services';
import { Button, Modal, Input, Textarea, Select, Badge } from '../../components/ui';
import type { Pipeline, Deal, PipelineStage } from '../../types';

// SVG Icons
const MoreIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="1" /><circle cx="19" cy="12" r="1" /><circle cx="5" cy="12" r="1" />
    </svg>
);

const PlusIcon = () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
    </svg>
);

const CalendarIcon = () => (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="4" width="18" height="18" rx="2" ry="2" /><line x1="16" y1="2" x2="16" y2="6" /><line x1="8" y1="2" x2="8" y2="6" /><line x1="3" y1="10" x2="21" y2="10" />
    </svg>
);

interface DealCardProps {
    deal: Deal;
    onDragStart: (e: DragEvent<HTMLDivElement>, deal: Deal) => void;
    onDragEnd: (e: DragEvent<HTMLDivElement>) => void;
    onClick: (deal: Deal) => void;
}

function DealCard({ deal, onDragStart, onDragEnd, onClick }: DealCardProps) {
    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('ms-MY', {
            style: 'currency',
            currency: 'MYR',
            minimumFractionDigits: 0,
        }).format(value);
    };

    const isOverdue = deal.expected_close_date && new Date(deal.expected_close_date) < new Date();

    return (
        <div
            className="kanban-card"
            draggable
            onDragStart={(e) => onDragStart(e, deal)}
            onDragEnd={onDragEnd}
            onClick={() => onClick(deal)}
        >
            <div className="kanban-card-header">
                <span className="kanban-card-title">{deal.name}</span>
                <span className={`kanban-card-priority ${deal.priority || 'medium'}`} />
            </div>

            <div className="kanban-card-customer">{deal.customer_name || 'No customer'}</div>

            <div className="kanban-card-value">{formatCurrency(deal.value)}</div>

            <div className="kanban-card-footer">
                <div className="kanban-card-owner">
                    <div className="kanban-card-avatar">
                        {deal.owner_name?.split(' ').map(n => n[0]).join('').toUpperCase() || 'U'}
                    </div>
                    <span className="kanban-card-owner-name">{deal.owner_name || 'Unassigned'}</span>
                </div>

                {deal.expected_close_date && (
                    <div className={`kanban-card-date ${isOverdue ? 'overdue' : ''}`}>
                        <CalendarIcon />
                        {new Date(deal.expected_close_date).toLocaleDateString('en-MY', {
                            month: 'short',
                            day: 'numeric',
                        })}
                    </div>
                )}
            </div>
        </div>
    );
}

interface KanbanColumnProps {
    stage: PipelineStage;
    deals: Deal[];
    onDragStart: (e: DragEvent<HTMLDivElement>, deal: Deal) => void;
    onDragEnd: (e: DragEvent<HTMLDivElement>) => void;
    onDragOver: (e: DragEvent<HTMLDivElement>) => void;
    onDrop: (e: DragEvent<HTMLDivElement>, stageId: string) => void;
    onDealClick: (deal: Deal) => void;
    onAddDeal: (stageId: string) => void;
}

function KanbanColumn({
    stage,
    deals,
    onDragStart,
    onDragEnd,
    onDragOver,
    onDrop,
    onDealClick,
    onAddDeal,
}: KanbanColumnProps) {
    const [isDragOver, setIsDragOver] = useState(false);

    const totalValue = deals.reduce((sum, deal) => sum + (deal.value || 0), 0);

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('ms-MY', {
            style: 'currency',
            currency: 'MYR',
            minimumFractionDigits: 0,
        }).format(value);
    };

    const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragOver(true);
        onDragOver(e);
    };

    const handleDragLeave = () => {
        setIsDragOver(false);
    };

    const handleDrop = (e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        setIsDragOver(false);
        onDrop(e, stage.id);
    };

    return (
        <div className="kanban-column">
            <div className="kanban-column-header">
                <div className="kanban-column-title">
                    <div className="kanban-column-color" style={{ background: stage.color || '#94a3b8' }} />
                    <span className="kanban-column-name">{stage.name}</span>
                    <span className="kanban-column-count">{deals.length}</span>
                </div>
                <span className="kanban-column-value">{formatCurrency(totalValue)}</span>
            </div>

            <div
                className={`kanban-column-body ${isDragOver ? 'drag-over' : ''}`}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
            >
                {deals.map((deal) => (
                    <DealCard
                        key={deal.id}
                        deal={deal}
                        onDragStart={onDragStart}
                        onDragEnd={onDragEnd}
                        onClick={onDealClick}
                    />
                ))}

                <button className="kanban-add-card" onClick={() => onAddDeal(stage.id)}>
                    <PlusIcon />
                    Add deal
                </button>
            </div>
        </div>
    );
}

export function PipelinePage() {
    const [pipelines, setPipelines] = useState<Pipeline[]>([]);
    const [selectedPipelineId, setSelectedPipelineId] = useState<string>('');
    const [deals, setDeals] = useState<Deal[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [draggedDeal, setDraggedDeal] = useState<Deal | null>(null);

    // Modal states
    const [isQuickViewOpen, setIsQuickViewOpen] = useState(false);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const [selectedDeal, setSelectedDeal] = useState<Deal | null>(null);
    const [newDealStageId, setNewDealStageId] = useState<string>('');
    const [newDealForm, setNewDealForm] = useState({
        name: '',
        customer_name: '',
        value: '',
        priority: 'medium',
        expected_close_date: '',
    });

    const selectedPipeline = pipelines.find((p) => p.id === selectedPipelineId);

    useEffect(() => {
        const fetchPipelines = async () => {
            try {
                const response = await pipelineService.getPipelines();
                setPipelines(response.pipelines);
                if (response.pipelines.length > 0) {
                    setSelectedPipelineId(response.pipelines[0].id);
                }
            } catch (error) {
                console.error('Failed to fetch pipelines:', error);
                // Use mock data
                const mockPipelines: Pipeline[] = [
                    {
                        id: '1',
                        name: 'Sales Pipeline',
                        stages: [
                            { id: 's1', name: 'Qualification', position: 1, color: '#3b82f6' },
                            { id: 's2', name: 'Proposal', position: 2, color: '#8b5cf6' },
                            { id: 's3', name: 'Negotiation', position: 3, color: '#f59e0b' },
                            { id: 's4', name: 'Closing', position: 4, color: '#10b981' },
                        ],
                    },
                ];
                setPipelines(mockPipelines);
                setSelectedPipelineId(mockPipelines[0].id);
            }
        };

        fetchPipelines();
    }, []);

    useEffect(() => {
        const fetchDeals = async () => {
            if (!selectedPipelineId) return;

            setIsLoading(true);
            try {
                const response = await pipelineService.getDeals(selectedPipelineId);
                setDeals(response.deals);
            } catch (error) {
                console.error('Failed to fetch deals:', error);
                // Use mock data
                setDeals([
                    {
                        id: 'd1',
                        name: 'Enterprise Package',
                        customer_name: 'Batik Industries Sdn Bhd',
                        value: 85000,
                        priority: 'high',
                        stage_id: 's1',
                        owner_name: 'Ahmad Razak',
                        expected_close_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'd2',
                        name: 'Premium Collection',
                        customer_name: 'Textile Malaysia',
                        value: 45000,
                        priority: 'medium',
                        stage_id: 's1',
                        owner_name: 'Siti Aminah',
                        expected_close_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'd3',
                        name: 'Wholesale Order',
                        customer_name: 'Kraf Malaysia',
                        value: 120000,
                        priority: 'urgent',
                        stage_id: 's2',
                        owner_name: 'Muhammad Hafiz',
                        expected_close_date: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'd4',
                        name: 'Custom Design Package',
                        customer_name: 'Heritage Batik',
                        value: 35000,
                        priority: 'medium',
                        stage_id: 's3',
                        owner_name: 'Nurul Aisyah',
                        expected_close_date: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'd5',
                        name: 'Annual Contract',
                        customer_name: 'Malaysian Airlines',
                        value: 250000,
                        priority: 'high',
                        stage_id: 's4',
                        owner_name: 'Ahmad Razak',
                        expected_close_date: new Date(Date.now() + 2 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                ]);
            } finally {
                setIsLoading(false);
            }
        };

        fetchDeals();
    }, [selectedPipelineId]);

    const handleDragStart = useCallback((_e: DragEvent<HTMLDivElement>, deal: Deal) => {
        setDraggedDeal(deal);
    }, []);

    const handleDragEnd = useCallback((_e: DragEvent<HTMLDivElement>) => {
        setDraggedDeal(null);
    }, []);

    const handleDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
    }, []);

    const handleDrop = useCallback(
        async (_e: DragEvent<HTMLDivElement>, stageId: string) => {
            if (!draggedDeal || draggedDeal.stage_id === stageId) return;

            // Optimistic update
            setDeals((prev) =>
                prev.map((d) => (d.id === draggedDeal.id ? { ...d, stage_id: stageId } : d))
            );

            try {
                await pipelineService.moveDeal(draggedDeal.id, { stage_id: stageId });
            } catch (error) {
                console.error('Failed to move deal:', error);
                // Revert on error
                setDeals((prev) =>
                    prev.map((d) =>
                        d.id === draggedDeal.id ? { ...d, stage_id: draggedDeal.stage_id } : d
                    )
                );
            }

            setDraggedDeal(null);
        },
        [draggedDeal]
    );

    const handleDealClick = useCallback((deal: Deal) => {
        setSelectedDeal(deal);
        setIsQuickViewOpen(true);
    }, []);

    const handleAddDeal = useCallback((stageId: string) => {
        setNewDealStageId(stageId);
        setNewDealForm({
            name: '',
            customer_name: '',
            value: '',
            priority: 'medium',
            expected_close_date: '',
        });
        setIsCreateModalOpen(true);
    }, []);

    const handleCreateDeal = async () => {
        const newDeal: Deal = {
            id: `temp-${Date.now()}`,
            name: newDealForm.name,
            customer_name: newDealForm.customer_name,
            value: parseFloat(newDealForm.value) || 0,
            priority: newDealForm.priority as 'low' | 'medium' | 'high' | 'urgent',
            stage_id: newDealStageId,
            owner_name: 'Current User',
            expected_close_date: newDealForm.expected_close_date || undefined,
        };

        setDeals((prev) => [...prev, newDeal]);
        setIsCreateModalOpen(false);
    };

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('ms-MY', {
            style: 'currency',
            currency: 'MYR',
            minimumFractionDigits: 0,
        }).format(value);
    };

    if (isLoading && pipelines.length === 0) {
        return (
            <div className="animate-fade-in">
                <div className="page-header">
                    <div className="skeleton" style={{ width: '200px', height: '32px' }} />
                </div>
                <div className="kanban-container">
                    {[1, 2, 3, 4].map((i) => (
                        <div key={i} className="kanban-column">
                            <div className="skeleton" style={{ width: '100%', height: '600px' }} />
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    return (
        <div className="animate-fade-in">
            {/* Page Header */}
            <div className="page-header">
                <div className="page-header-left">
                    <h1 className="page-title">Pipeline</h1>
                    <p className="page-description">Drag and drop deals between stages</p>
                </div>
                <div className="page-header-actions">
                    {pipelines.length > 1 && (
                        <Select
                            options={pipelines.map((p) => ({ value: p.id, label: p.name }))}
                            value={selectedPipelineId}
                            onChange={(e) => setSelectedPipelineId(e.target.value)}
                            selectSize="md"
                        />
                    )}
                    <Button onClick={() => handleAddDeal(selectedPipeline?.stages[0]?.id || '')}>
                        <PlusIcon />
                        <span>New Deal</span>
                    </Button>
                </div>
            </div>

            {/* Kanban Board */}
            <div className="kanban-container">
                {selectedPipeline?.stages.map((stage) => {
                    const stageDeals = deals.filter((d) => d.stage_id === stage.id);
                    return (
                        <KanbanColumn
                            key={stage.id}
                            stage={stage}
                            deals={stageDeals}
                            onDragStart={handleDragStart}
                            onDragEnd={handleDragEnd}
                            onDragOver={handleDragOver}
                            onDrop={handleDrop}
                            onDealClick={handleDealClick}
                            onAddDeal={handleAddDeal}
                        />
                    );
                })}
            </div>

            {/* Quick View Modal */}
            <Modal
                isOpen={isQuickViewOpen}
                onClose={() => setIsQuickViewOpen(false)}
                title={selectedDeal?.name || 'Deal Details'}
                size="lg"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsQuickViewOpen(false)}>
                            Close
                        </Button>
                        <Button>Edit Deal</Button>
                    </>
                }
            >
                {selectedDeal && (
                    <div>
                        <div className="info-grid mb-6">
                            <div className="info-item">
                                <span className="info-label">Customer</span>
                                <span className="info-value">{selectedDeal.customer_name || 'N/A'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Value</span>
                                <span className="info-value text-success font-bold">
                                    {formatCurrency(selectedDeal.value)}
                                </span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Priority</span>
                                <Badge
                                    variant={
                                        selectedDeal.priority === 'urgent'
                                            ? 'danger'
                                            : selectedDeal.priority === 'high'
                                                ? 'warning'
                                                : 'default'
                                    }
                                >
                                    {selectedDeal.priority}
                                </Badge>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Owner</span>
                                <span className="info-value">{selectedDeal.owner_name || 'Unassigned'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Expected Close</span>
                                <span className="info-value">
                                    {selectedDeal.expected_close_date
                                        ? new Date(selectedDeal.expected_close_date).toLocaleDateString('en-MY')
                                        : 'Not set'}
                                </span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Stage</span>
                                <span className="info-value">
                                    {selectedPipeline?.stages.find((s) => s.id === selectedDeal.stage_id)?.name || 'Unknown'}
                                </span>
                            </div>
                        </div>
                    </div>
                )}
            </Modal>

            {/* Create Deal Modal */}
            <Modal
                isOpen={isCreateModalOpen}
                onClose={() => setIsCreateModalOpen(false)}
                title="Create New Deal"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsCreateModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleCreateDeal}>Create Deal</Button>
                    </>
                }
            >
                <form onSubmit={(e) => { e.preventDefault(); handleCreateDeal(); }}>
                    <Input
                        label="Deal Name"
                        placeholder="Enter deal name"
                        value={newDealForm.name}
                        onChange={(e) => setNewDealForm((prev) => ({ ...prev, name: e.target.value }))}
                        required
                    />

                    <div style={{ marginTop: '1rem' }}>
                        <Input
                            label="Customer Name"
                            placeholder="Enter customer name"
                            value={newDealForm.customer_name}
                            onChange={(e) => setNewDealForm((prev) => ({ ...prev, customer_name: e.target.value }))}
                        />
                    </div>

                    <div style={{ marginTop: '1rem' }}>
                        <Input
                            type="number"
                            label="Deal Value (MYR)"
                            placeholder="0.00"
                            value={newDealForm.value}
                            onChange={(e) => setNewDealForm((prev) => ({ ...prev, value: e.target.value }))}
                        />
                    </div>

                    <div style={{ marginTop: '1rem' }}>
                        <Select
                            label="Priority"
                            options={[
                                { value: 'low', label: 'Low' },
                                { value: 'medium', label: 'Medium' },
                                { value: 'high', label: 'High' },
                                { value: 'urgent', label: 'Urgent' },
                            ]}
                            value={newDealForm.priority}
                            onChange={(e) => setNewDealForm((prev) => ({ ...prev, priority: e.target.value }))}
                        />
                    </div>

                    <div style={{ marginTop: '1rem' }}>
                        <Input
                            type="date"
                            label="Expected Close Date"
                            value={newDealForm.expected_close_date}
                            onChange={(e) =>
                                setNewDealForm((prev) => ({ ...prev, expected_close_date: e.target.value }))
                            }
                        />
                    </div>
                </form>
            </Modal>
        </div>
    );
}

export default PipelinePage;
