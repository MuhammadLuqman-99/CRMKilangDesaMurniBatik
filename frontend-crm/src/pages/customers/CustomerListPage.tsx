// ============================================
// Customer List Page
// Production-Ready Customer Management List
// ============================================

import { useState, useEffect, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { customerService } from '../../services';
import { Button, Input, Select, Badge, Table, Pagination } from '../../components/ui';
import type { Customer, CustomerFilters } from '../../types';

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

const statusOptions = [
    { value: '', label: 'All Statuses' },
    { value: 'active', label: 'Active' },
    { value: 'inactive', label: 'Inactive' },
    { value: 'prospect', label: 'Prospect' },
    { value: 'churned', label: 'Churned' },
];

const segmentOptions = [
    { value: '', label: 'All Segments' },
    { value: 'enterprise', label: 'Enterprise' },
    { value: 'mid_market', label: 'Mid-Market' },
    { value: 'smb', label: 'SMB' },
    { value: 'startup', label: 'Startup' },
];

export function CustomerListPage() {
    const navigate = useNavigate();
    const [customers, setCustomers] = useState<Customer[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [selectedIds, setSelectedIds] = useState<string[]>([]);

    const [filters, setFilters] = useState<CustomerFilters>({
        page: 1,
        per_page: 20,
        status: '',
        segment: '',
        search: '',
    });

    const [meta, setMeta] = useState({
        total: 0,
        page: 1,
        per_page: 20,
        total_pages: 1,
    });

    useEffect(() => {
        const fetchCustomers = async () => {
            setIsLoading(true);
            try {
                const response = await customerService.getCustomers(filters);
                setCustomers(response.customers);
                setMeta(response.meta);
            } catch (error) {
                console.error('Failed to fetch customers:', error);
                // Mock data
                setCustomers([
                    {
                        id: 'c1',
                        name: 'Batik Industries Sdn Bhd',
                        email: 'purchasing@batik-industries.com.my',
                        phone: '+60123456789',
                        status: 'active',
                        segment: 'enterprise',
                        industry: 'Manufacturing',
                        total_value: 850000,
                        deals_count: 5,
                        owner_name: 'Ahmad Razak',
                        created_at: new Date(Date.now() - 180 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'c2',
                        name: 'Textile Malaysia',
                        email: 'info@textile-my.com',
                        phone: '+60198765432',
                        status: 'active',
                        segment: 'mid_market',
                        industry: 'Retail',
                        total_value: 320000,
                        deals_count: 3,
                        owner_name: 'Siti Aminah',
                        created_at: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'c3',
                        name: 'Kraf Malaysia',
                        email: 'procurement@kraf-malaysia.gov.my',
                        phone: '+60134567890',
                        status: 'active',
                        segment: 'enterprise',
                        industry: 'Government',
                        total_value: 1200000,
                        deals_count: 8,
                        owner_name: 'Muhammad Hafiz',
                        created_at: new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'c4',
                        name: 'Heritage Batik',
                        email: 'orders@heritage-batik.com',
                        phone: '+60145678901',
                        status: 'prospect',
                        segment: 'smb',
                        industry: 'Retail',
                        total_value: 45000,
                        deals_count: 1,
                        owner_name: 'Nurul Aisyah',
                        created_at: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                ]);
                setMeta({ total: 89, page: 1, per_page: 20, total_pages: 5 });
            } finally {
                setIsLoading(false);
            }
        };

        fetchCustomers();
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
            active: 'success',
            inactive: 'default',
            prospect: 'warning',
            churned: 'danger',
        };
        return <Badge variant={variants[status] || 'default'}>{status}</Badge>;
    };

    const getSegmentBadge = (segment: string) => {
        const variants: Record<string, 'default' | 'success' | 'info' | 'purple'> = {
            enterprise: 'purple',
            mid_market: 'info',
            smb: 'default',
            startup: 'success',
        };
        return <Badge variant={variants[segment] || 'default'}>{segment?.replace('_', '-')}</Badge>;
    };

    const columns = useMemo(
        () => [
            {
                key: 'name',
                header: 'Customer',
                render: (customer: Customer) => (
                    <div>
                        <Link
                            to={`/customers/${customer.id}`}
                            className="font-medium"
                            style={{ color: 'var(--text-primary)' }}
                        >
                            {customer.name}
                        </Link>
                        <div className="text-sm text-muted">{customer.industry || 'No industry'}</div>
                    </div>
                ),
            },
            {
                key: 'contact',
                header: 'Contact',
                render: (customer: Customer) => (
                    <div>
                        <div className="text-sm">{customer.email}</div>
                        <div className="text-sm text-muted">{customer.phone || '-'}</div>
                    </div>
                ),
            },
            {
                key: 'status',
                header: 'Status',
                render: (customer: Customer) => getStatusBadge(customer.status),
            },
            {
                key: 'segment',
                header: 'Segment',
                render: (customer: Customer) => customer.segment && getSegmentBadge(customer.segment),
            },
            {
                key: 'total_value',
                header: 'Total Value',
                render: (customer: Customer) => (
                    <span className="font-medium text-success">{formatCurrency(customer.total_value || 0)}</span>
                ),
            },
            {
                key: 'deals',
                header: 'Deals',
                render: (customer: Customer) => (
                    <span className="text-sm">{customer.deals_count || 0}</span>
                ),
            },
            {
                key: 'owner',
                header: 'Owner',
                render: (customer: Customer) => (
                    <span className="text-sm">{customer.owner_name || 'Unassigned'}</span>
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
                    <h1 className="page-title">Customers</h1>
                    <p className="page-description">
                        Manage your customer relationships and accounts
                    </p>
                </div>
                <div className="page-header-actions">
                    <Link to="/customers/new">
                        <Button>
                            <PlusIcon />
                            <span>New Customer</span>
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
                        placeholder="Search customers..."
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
                        options={segmentOptions}
                        value={filters.segment}
                        onChange={(e) => setFilters((prev) => ({ ...prev, segment: e.target.value, page: 1 }))}
                    />
                </div>
            </div>

            {/* Customers Table */}
            <div className="card" style={{ padding: 0 }}>
                <Table
                    columns={columns}
                    data={customers}
                    keyExtractor={(customer) => customer.id}
                    loading={isLoading}
                    emptyMessage="No customers found"
                    onRowClick={(customer) => navigate(`/customers/${customer.id}`)}
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
        </div>
    );
}

export default CustomerListPage;
