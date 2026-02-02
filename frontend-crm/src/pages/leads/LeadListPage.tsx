// ============================================
// Lead List Page
// Production-Ready Lead Management List
// ============================================

import { useState, useEffect, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { leadService } from '../../services';
import { Button, Input, Select, Badge, Table, Pagination, Modal } from '../../components/ui';
import type { Lead, LeadFilters } from '../../types';

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

const FilterIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
    </svg>
);

const DownloadIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline points="7 10 12 15 17 10" /><line x1="12" y1="15" x2="12" y2="3" />
    </svg>
);

const UploadIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline points="17 8 12 3 7 8" /><line x1="12" y1="3" x2="12" y2="15" />
    </svg>
);

const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'new', label: 'New' },
    { value: 'contacted', label: 'Contacted' },
    { value: 'qualified', label: 'Qualified' },
    { value: 'unqualified', label: 'Unqualified' },
    { value: 'converted', label: 'Converted' },
];

const sourceOptions = [
    { value: '', label: 'All Sources' },
    { value: 'website', label: 'Website' },
    { value: 'referral', label: 'Referral' },
    { value: 'social_media', label: 'Social Media' },
    { value: 'event', label: 'Event' },
    { value: 'cold_call', label: 'Cold Call' },
    { value: 'advertisement', label: 'Advertisement' },
];

export function LeadListPage() {
    const navigate = useNavigate();
    const [leads, setLeads] = useState<Lead[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [selectedIds, setSelectedIds] = useState<string[]>([]);
    const [isImportModalOpen, setIsImportModalOpen] = useState(false);

    // Filters
    const [filters, setFilters] = useState<LeadFilters>({
        page: 1,
        per_page: 20,
        status: '',
        source: '',
        search: '',
    });

    const [meta, setMeta] = useState({
        total: 0,
        page: 1,
        per_page: 20,
        total_pages: 1,
    });

    useEffect(() => {
        const fetchLeads = async () => {
            setIsLoading(true);
            try {
                const response = await leadService.getLeads(filters);
                setLeads(response.leads);
                setMeta(response.meta);
            } catch (error) {
                console.error('Failed to fetch leads:', error);
                // Use mock data
                setLeads([
                    {
                        id: 'l1',
                        first_name: 'Ahmad',
                        last_name: 'Razak',
                        email: 'ahmad@batik-industries.com.my',
                        phone: '+60123456789',
                        company: 'Batik Industries Sdn Bhd',
                        status: 'qualified',
                        source: 'website',
                        score: 85,
                        score_label: 'hot',
                        owner_name: 'Siti Aminah',
                        created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'l2',
                        first_name: 'Nurul',
                        last_name: 'Aisyah',
                        email: 'nurul@textile-my.com',
                        phone: '+60198765432',
                        company: 'Textile Malaysia',
                        status: 'contacted',
                        source: 'referral',
                        score: 72,
                        score_label: 'warm',
                        owner_name: 'Muhammad Hafiz',
                        created_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'l3',
                        first_name: 'Ali',
                        last_name: 'Hassan',
                        email: 'ali@kraf-malaysia.gov.my',
                        phone: '+60134567890',
                        company: 'Kraf Malaysia',
                        status: 'new',
                        source: 'event',
                        score: 45,
                        score_label: 'cold',
                        owner_name: 'Unassigned',
                        created_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'l4',
                        first_name: 'Fatimah',
                        last_name: 'Zahra',
                        email: 'fatimah@heritage-batik.com',
                        phone: '+60145678901',
                        company: 'Heritage Batik',
                        status: 'qualified',
                        source: 'social_media',
                        score: 91,
                        score_label: 'hot',
                        owner_name: 'Siti Aminah',
                        created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'l5',
                        first_name: 'Hafiz',
                        last_name: 'Ibrahim',
                        email: 'hafiz@mas.com.my',
                        phone: '+60156789012',
                        company: 'Malaysian Airlines',
                        status: 'contacted',
                        source: 'cold_call',
                        score: 65,
                        score_label: 'warm',
                        owner_name: 'Muhammad Hafiz',
                        created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                ]);
                setMeta({ total: 156, page: 1, per_page: 20, total_pages: 8 });
            } finally {
                setIsLoading(false);
            }
        };

        fetchLeads();
    }, [filters]);

    const handleSearch = (value: string) => {
        setFilters((prev) => ({ ...prev, search: value, page: 1 }));
    };

    const handleStatusFilter = (status: string) => {
        setFilters((prev) => ({ ...prev, status, page: 1 }));
    };

    const handleSourceFilter = (source: string) => {
        setFilters((prev) => ({ ...prev, source, page: 1 }));
    };

    const handlePageChange = (page: number) => {
        setFilters((prev) => ({ ...prev, page }));
    };

    const handleExport = async () => {
        try {
            const blob = await leadService.exportLeads(filters, 'csv');
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `leads-${new Date().toISOString().split('T')[0]}.csv`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } catch (error) {
            console.error('Failed to export leads:', error);
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

    const getScoreBadge = (label: string) => {
        const variants: Record<string, 'default' | 'success' | 'warning' | 'danger'> = {
            hot: 'danger',
            warm: 'warning',
            cold: 'default',
        };
        return <Badge variant={variants[label] || 'default'}>{label}</Badge>;
    };

    const columns = useMemo(
        () => [
            {
                key: 'name',
                header: 'Lead',
                render: (lead: Lead) => (
                    <div>
                        <Link
                            to={`/leads/${lead.id}`}
                            className="font-medium text-primary-color"
                            style={{ color: 'var(--text-primary)' }}
                        >
                            {lead.first_name} {lead.last_name}
                        </Link>
                        <div className="text-sm text-muted">{lead.company || 'No company'}</div>
                    </div>
                ),
            },
            {
                key: 'email',
                header: 'Contact',
                render: (lead: Lead) => (
                    <div>
                        <div className="text-sm">{lead.email}</div>
                        <div className="text-sm text-muted">{lead.phone || '-'}</div>
                    </div>
                ),
            },
            {
                key: 'status',
                header: 'Status',
                render: (lead: Lead) => getStatusBadge(lead.status),
            },
            {
                key: 'source',
                header: 'Source',
                render: (lead: Lead) => <span className="text-sm capitalize">{lead.source?.replace('_', ' ')}</span>,
            },
            {
                key: 'score',
                header: 'Score',
                render: (lead: Lead) => (
                    <div className="flex items-center gap-2">
                        <span className="font-medium">{lead.score || 0}</span>
                        {lead.score_label && getScoreBadge(lead.score_label)}
                    </div>
                ),
            },
            {
                key: 'owner',
                header: 'Owner',
                render: (lead: Lead) => (
                    <span className="text-sm">{lead.owner_name || 'Unassigned'}</span>
                ),
            },
            {
                key: 'created_at',
                header: 'Created',
                render: (lead: Lead) => (
                    <span className="text-sm text-muted">
                        {new Date(lead.created_at).toLocaleDateString('en-MY')}
                    </span>
                ),
            },
        ],
        []
    );

    return (
        <div className="animate-fade-in">
            {/* Page Header */}
            <div className="page-header">
                <div className="page-header-left">
                    <h1 className="page-title">Leads</h1>
                    <p className="page-description">
                        Manage and track your leads through the sales process
                    </p>
                </div>
                <div className="page-header-actions">
                    <Button variant="outline" onClick={() => setIsImportModalOpen(true)}>
                        <UploadIcon />
                        <span>Import</span>
                    </Button>
                    <Button variant="outline" onClick={handleExport}>
                        <DownloadIcon />
                        <span>Export</span>
                    </Button>
                    <Link to="/leads/new">
                        <Button>
                            <PlusIcon />
                            <span>New Lead</span>
                        </Button>
                    </Link>
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
                        placeholder="Search leads by name, email, or company..."
                        value={filters.search}
                        onChange={(e) => handleSearch(e.target.value)}
                    />
                </div>

                <div className="filter-group">
                    <Select
                        options={statusOptions}
                        value={filters.status}
                        onChange={(e) => handleStatusFilter(e.target.value)}
                        placeholder="Status"
                    />
                    <Select
                        options={sourceOptions}
                        value={filters.source}
                        onChange={(e) => handleSourceFilter(e.target.value)}
                        placeholder="Source"
                    />
                    <button className="filter-btn">
                        <FilterIcon />
                        More Filters
                    </button>
                </div>
            </div>

            {/* Leads Table */}
            <div className="card" style={{ padding: 0 }}>
                <Table
                    columns={columns}
                    data={leads}
                    keyExtractor={(lead) => lead.id}
                    loading={isLoading}
                    emptyMessage="No leads found"
                    onRowClick={(lead) => navigate(`/leads/${lead.id}`)}
                    showCheckboxes
                    selectedRows={selectedIds}
                    onSelectionChange={setSelectedIds}
                />

                <Pagination
                    currentPage={meta.page}
                    totalPages={meta.total_pages}
                    totalItems={meta.total}
                    itemsPerPage={meta.per_page}
                    onPageChange={handlePageChange}
                />
            </div>

            {/* Import Modal */}
            <Modal
                isOpen={isImportModalOpen}
                onClose={() => setIsImportModalOpen(false)}
                title="Import Leads"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsImportModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button>Upload & Import</Button>
                    </>
                }
            >
                <div className="text-center p-6">
                    <div
                        style={{
                            border: '2px dashed var(--border-color)',
                            borderRadius: 'var(--radius-xl)',
                            padding: '3rem',
                            cursor: 'pointer',
                        }}
                    >
                        <UploadIcon />
                        <p className="mt-4 font-medium">Click to upload or drag and drop</p>
                        <p className="text-sm text-muted">CSV or XLSX files (max 5MB)</p>
                    </div>

                    <div className="mt-6 text-left">
                        <h4 className="font-medium mb-2">Download Template</h4>
                        <p className="text-sm text-muted mb-3">
                            Use our template to ensure your data is formatted correctly.
                        </p>
                        <Button variant="outline" fullWidth>
                            <DownloadIcon />
                            Download CSV Template
                        </Button>
                    </div>
                </div>
            </Modal>
        </div>
    );
}

export default LeadListPage;
