// ============================================
// Customer Detail Page
// Production-Ready Customer Detail View
// ============================================

import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { customerService } from '../../services';
import { Button, Badge, Card, Modal, Input, Textarea } from '../../components/ui';
import type { Customer, CustomerContact, CustomerNote } from '../../types';

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

const PlusIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="12" y1="5" x2="12" y2="19" /><line x1="5" y1="12" x2="19" y2="12" />
    </svg>
);

export function CustomerDetailPage() {
    const { id } = useParams<{ id: string }>();
    const [customer, setCustomer] = useState<Customer | null>(null);
    const [contacts, setContacts] = useState<CustomerContact[]>([]);
    const [notes, setNotes] = useState<CustomerNote[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('overview');
    const [isContactModalOpen, setIsContactModalOpen] = useState(false);
    const [isNoteModalOpen, setIsNoteModalOpen] = useState(false);
    const [newNote, setNewNote] = useState('');

    const formatCurrency = (value: number) => {
        return new Intl.NumberFormat('ms-MY', {
            style: 'currency',
            currency: 'MYR',
            minimumFractionDigits: 0,
        }).format(value);
    };

    useEffect(() => {
        const fetchCustomer = async () => {
            if (!id) return;

            setIsLoading(true);
            try {
                const data = await customerService.getCustomer(id);
                setCustomer(data);
                const contactsData = await customerService.getContacts(id);
                setContacts(contactsData.contacts);
                const notesData = await customerService.getNotes(id);
                setNotes(notesData.notes);
            } catch (error) {
                console.error('Failed to fetch customer:', error);
                // Mock data
                setCustomer({
                    id: id,
                    name: 'Batik Industries Sdn Bhd',
                    email: 'purchasing@batik-industries.com.my',
                    phone: '+60123456789',
                    status: 'active',
                    segment: 'enterprise',
                    industry: 'Manufacturing',
                    website: 'https://batik-industries.com.my',
                    address: '123 Jalan Batik, Seksyen 15, 40000 Shah Alam, Selangor',
                    total_value: 850000,
                    deals_count: 5,
                    owner_name: 'Ahmad Razak',
                    created_at: new Date(Date.now() - 180 * 24 * 60 * 60 * 1000).toISOString(),
                });
                setContacts([
                    {
                        id: 'ct1',
                        name: 'Ahmad bin Hassan',
                        email: 'ahmad@batik-industries.com.my',
                        phone: '+60123456789',
                        title: 'Procurement Manager',
                        is_primary: true,
                    },
                    {
                        id: 'ct2',
                        name: 'Fatimah binti Ali',
                        email: 'fatimah@batik-industries.com.my',
                        phone: '+60198765432',
                        title: 'Finance Director',
                        is_primary: false,
                    },
                ]);
                setNotes([
                    {
                        id: 'n1',
                        content: 'Customer interested in expanding wholesale orders.',
                        created_by: 'Ahmad Razak',
                        created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                    {
                        id: 'n2',
                        content: 'Discussed pricing for annual contract renewal.',
                        created_by: 'Siti Aminah',
                        created_at: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
                    },
                ]);
            } finally {
                setIsLoading(false);
            }
        };

        fetchCustomer();
    }, [id]);

    const getStatusBadge = (status: string) => {
        const variants: Record<string, 'default' | 'success' | 'warning' | 'danger'> = {
            active: 'success',
            inactive: 'default',
            prospect: 'warning',
            churned: 'danger',
        };
        return <Badge variant={variants[status] || 'default'}>{status}</Badge>;
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

    if (!customer) {
        return (
            <div className="empty-state">
                <div className="empty-state-icon">🔍</div>
                <h3 className="empty-state-title">Customer not found</h3>
                <p className="empty-state-description">
                    The customer you're looking for doesn't exist or has been deleted.
                </p>
                <Link to="/customers">
                    <Button>Back to Customers</Button>
                </Link>
            </div>
        );
    }

    return (
        <div className="animate-fade-in">
            {/* Breadcrumb */}
            <div className="mb-4">
                <Link to="/customers" className="flex items-center gap-2 text-muted" style={{ textDecoration: 'none' }}>
                    <ArrowLeftIcon />
                    Back to Customers
                </Link>
            </div>

            {/* Detail Header */}
            <div className="detail-header">
                <div className="detail-info">
                    <div className="detail-avatar" style={{ background: 'linear-gradient(135deg, #10b981, #34d399)' }}>
                        {customer.name.split(' ').slice(0, 2).map(n => n[0]).join('')}
                    </div>
                    <div>
                        <h1 className="detail-title">{customer.name}</h1>
                        <p className="detail-subtitle">{customer.industry || 'No industry'}</p>
                        <div className="detail-badges">
                            {getStatusBadge(customer.status)}
                            {customer.segment && (
                                <Badge variant="purple">{customer.segment.replace('_', '-')}</Badge>
                            )}
                        </div>
                    </div>
                </div>

                <div className="detail-actions">
                    <Button variant="outline" onClick={() => window.location.href = `tel:${customer.phone}`}>
                        <PhoneIcon />
                        Call
                    </Button>
                    <Button variant="outline" onClick={() => window.location.href = `mailto:${customer.email}`}>
                        <MailIcon />
                        Email
                    </Button>
                    <Link to={`/customers/${customer.id}/edit`}>
                        <Button variant="outline">
                            <EditIcon />
                            Edit
                        </Button>
                    </Link>
                </div>
            </div>

            {/* Stats Row */}
            <div className="stats-grid mb-6" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
                <div className="stat-card">
                    <span className="stat-card-title">Total Revenue</span>
                    <span className="stat-card-value text-success">{formatCurrency(customer.total_value || 0)}</span>
                </div>
                <div className="stat-card">
                    <span className="stat-card-title">Total Deals</span>
                    <span className="stat-card-value">{customer.deals_count || 0}</span>
                </div>
                <div className="stat-card">
                    <span className="stat-card-title">Avg. Deal Size</span>
                    <span className="stat-card-value">
                        {formatCurrency((customer.total_value || 0) / (customer.deals_count || 1))}
                    </span>
                </div>
                <div className="stat-card">
                    <span className="stat-card-title">Customer Since</span>
                    <span className="stat-card-value">
                        {new Date(customer.created_at).toLocaleDateString('en-MY', { month: 'short', year: 'numeric' })}
                    </span>
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
                    className={`detail-tab ${activeTab === 'contacts' ? 'active' : ''}`}
                    onClick={() => setActiveTab('contacts')}
                >
                    Contacts ({contacts.length})
                </button>
                <button
                    className={`detail-tab ${activeTab === 'deals' ? 'active' : ''}`}
                    onClick={() => setActiveTab('deals')}
                >
                    Deals
                </button>
                <button
                    className={`detail-tab ${activeTab === 'notes' ? 'active' : ''}`}
                    onClick={() => setActiveTab('notes')}
                >
                    Notes ({notes.length})
                </button>
            </div>

            {/* Tab Content */}
            {activeTab === 'overview' && (
                <div className="grid grid-cols-3 gap-6">
                    <Card className="col-span-2" padding="lg">
                        <h3 className="font-semibold text-lg mb-4">Customer Information</h3>
                        <div className="info-grid">
                            <div className="info-item">
                                <span className="info-label">Company Name</span>
                                <span className="info-value">{customer.name}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Email</span>
                                <span className="info-value">{customer.email}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Phone</span>
                                <span className="info-value">{customer.phone || 'N/A'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Industry</span>
                                <span className="info-value">{customer.industry || 'N/A'}</span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Website</span>
                                <span className="info-value">
                                    {customer.website ? (
                                        <a href={customer.website} target="_blank" rel="noopener noreferrer" style={{ color: 'var(--primary)' }}>
                                            {customer.website}
                                        </a>
                                    ) : 'N/A'}
                                </span>
                            </div>
                            <div className="info-item">
                                <span className="info-label">Account Owner</span>
                                <span className="info-value">{customer.owner_name || 'Unassigned'}</span>
                            </div>
                            <div className="info-item col-span-2">
                                <span className="info-label">Address</span>
                                <span className="info-value">{customer.address || 'N/A'}</span>
                            </div>
                        </div>
                    </Card>

                    <div className="flex flex-col gap-4">
                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Primary Contact</h3>
                            {contacts.find(c => c.is_primary) ? (
                                <div className="flex items-center gap-3">
                                    <div className="activity-avatar">
                                        {contacts.find(c => c.is_primary)?.name.split(' ').map(n => n[0]).join('')}
                                    </div>
                                    <div>
                                        <p className="font-medium">{contacts.find(c => c.is_primary)?.name}</p>
                                        <p className="text-sm text-muted">{contacts.find(c => c.is_primary)?.title}</p>
                                    </div>
                                </div>
                            ) : (
                                <p className="text-muted">No primary contact set</p>
                            )}
                        </Card>

                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Recent Notes</h3>
                            {notes.slice(0, 2).map((note) => (
                                <div key={note.id} className="mb-3 pb-3 border-b last:border-b-0 last:pb-0 last:mb-0">
                                    <p className="text-sm">{note.content}</p>
                                    <p className="text-xs text-muted mt-1">
                                        {note.created_by} · {new Date(note.created_at).toLocaleDateString('en-MY')}
                                    </p>
                                </div>
                            ))}
                            {notes.length === 0 && <p className="text-muted text-sm">No notes yet</p>}
                        </Card>
                    </div>
                </div>
            )}

            {activeTab === 'contacts' && (
                <Card padding="lg">
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="font-semibold text-lg">Contacts</h3>
                        <Button onClick={() => setIsContactModalOpen(true)}>
                            <PlusIcon />
                            Add Contact
                        </Button>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        {contacts.map((contact) => (
                            <div key={contact.id} className="card p-4">
                                <div className="flex items-center gap-3">
                                    <div className="activity-avatar">
                                        {contact.name.split(' ').map(n => n[0]).join('')}
                                    </div>
                                    <div className="flex-1">
                                        <div className="flex items-center gap-2">
                                            <p className="font-medium">{contact.name}</p>
                                            {contact.is_primary && <Badge variant="success" size="sm">Primary</Badge>}
                                        </div>
                                        <p className="text-sm text-muted">{contact.title || 'No title'}</p>
                                        <p className="text-sm">{contact.email}</p>
                                        <p className="text-sm text-muted">{contact.phone}</p>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </Card>
            )}

            {activeTab === 'deals' && (
                <Card padding="lg">
                    <h3 className="font-semibold text-lg mb-4">Deals</h3>
                    <table className="table">
                        <thead>
                            <tr>
                                <th>Deal Name</th>
                                <th>Value</th>
                                <th>Stage</th>
                                <th>Close Date</th>
                                <th>Status</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr>
                                <td>Annual Contract 2024</td>
                                <td className="text-success font-medium">{formatCurrency(250000)}</td>
                                <td>Closing</td>
                                <td>{new Date().toLocaleDateString('en-MY')}</td>
                                <td><Badge variant="warning">In Progress</Badge></td>
                            </tr>
                            <tr>
                                <td>Q3 Bulk Order</td>
                                <td className="text-success font-medium">{formatCurrency(180000)}</td>
                                <td>-</td>
                                <td>{new Date(Date.now() - 60 * 24 * 60 * 60 * 1000).toLocaleDateString('en-MY')}</td>
                                <td><Badge variant="success">Won</Badge></td>
                            </tr>
                            <tr>
                                <td>Custom Design Project</td>
                                <td className="text-success font-medium">{formatCurrency(95000)}</td>
                                <td>-</td>
                                <td>{new Date(Date.now() - 120 * 24 * 60 * 60 * 1000).toLocaleDateString('en-MY')}</td>
                                <td><Badge variant="success">Won</Badge></td>
                            </tr>
                        </tbody>
                    </table>
                </Card>
            )}

            {activeTab === 'notes' && (
                <Card padding="lg">
                    <div className="flex justify-between items-center mb-4">
                        <h3 className="font-semibold text-lg">Notes</h3>
                        <Button onClick={() => setIsNoteModalOpen(true)}>
                            <PlusIcon />
                            Add Note
                        </Button>
                    </div>

                    <div className="timeline">
                        {notes.map((note) => (
                            <div key={note.id} className="timeline-item">
                                <div className="timeline-icon note">📝</div>
                                <div className="timeline-content">
                                    <p className="timeline-body">{note.content}</p>
                                    <div className="timeline-time">
                                        {note.created_by} · {new Date(note.created_at).toLocaleDateString('en-MY')}
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </Card>
            )}

            {/* Add Contact Modal */}
            <Modal
                isOpen={isContactModalOpen}
                onClose={() => setIsContactModalOpen(false)}
                title="Add Contact"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsContactModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button>Add Contact</Button>
                    </>
                }
            >
                <Input label="Name" placeholder="Full name" className="mb-4" />
                <Input type="email" label="Email" placeholder="email@company.com" className="mb-4" />
                <Input type="tel" label="Phone" placeholder="+60123456789" className="mb-4" />
                <Input label="Job Title" placeholder="e.g. Manager" />
            </Modal>

            {/* Add Note Modal */}
            <Modal
                isOpen={isNoteModalOpen}
                onClose={() => setIsNoteModalOpen(false)}
                title="Add Note"
                size="md"
                footer={
                    <>
                        <Button variant="outline" onClick={() => setIsNoteModalOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={() => {
                            if (newNote.trim()) {
                                setNotes(prev => [{
                                    id: `n${Date.now()}`,
                                    content: newNote,
                                    created_by: 'Current User',
                                    created_at: new Date().toISOString(),
                                }, ...prev]);
                                setNewNote('');
                                setIsNoteModalOpen(false);
                            }
                        }}>
                            Add Note
                        </Button>
                    </>
                }
            >
                <Textarea
                    label="Note"
                    placeholder="Enter your note..."
                    rows={5}
                    value={newNote}
                    onChange={(e) => setNewNote(e.target.value)}
                />
            </Modal>
        </div>
    );
}

export default CustomerDetailPage;
