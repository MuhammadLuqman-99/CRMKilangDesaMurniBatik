// ============================================
// Lead Form Page
// Production-Ready Lead Create/Edit Form
// ============================================

import { useState, useEffect, type FormEvent } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { leadService } from '../../services';
import { Button, Input, Select, Textarea, Card } from '../../components/ui';
import { useToast } from '../../components/ui/Toast';
import type { Lead, CreateLeadRequest } from '../../types';

// SVG Icons
const ArrowLeftIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="19" y1="12" x2="5" y2="12" /><polyline points="12 19 5 12 12 5" />
    </svg>
);

const sourceOptions = [
    { value: 'website', label: 'Website' },
    { value: 'referral', label: 'Referral' },
    { value: 'social_media', label: 'Social Media' },
    { value: 'event', label: 'Event' },
    { value: 'cold_call', label: 'Cold Call' },
    { value: 'advertisement', label: 'Advertisement' },
    { value: 'other', label: 'Other' },
];

const industryOptions = [
    { value: 'manufacturing', label: 'Manufacturing' },
    { value: 'retail', label: 'Retail' },
    { value: 'hospitality', label: 'Hospitality' },
    { value: 'corporate', label: 'Corporate' },
    { value: 'government', label: 'Government' },
    { value: 'education', label: 'Education' },
    { value: 'healthcare', label: 'Healthcare' },
    { value: 'other', label: 'Other' },
];

export function LeadFormPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { success, error } = useToast();
    const isEditing = !!id;

    const [isLoading, setIsLoading] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [formData, setFormData] = useState<CreateLeadRequest>({
        first_name: '',
        last_name: '',
        email: '',
        phone: '',
        company: '',
        title: '',
        source: 'website',
        industry: '',
        website: '',
        address: '',
        notes: '',
    });
    const [errors, setErrors] = useState<Record<string, string>>({});

    useEffect(() => {
        const fetchLead = async () => {
            if (!id) return;

            setIsLoading(true);
            try {
                const lead = await leadService.getLead(id);
                setFormData({
                    first_name: lead.first_name,
                    last_name: lead.last_name,
                    email: lead.email,
                    phone: lead.phone || '',
                    company: lead.company || '',
                    title: lead.title || '',
                    source: lead.source || 'website',
                    industry: lead.industry || '',
                    website: lead.website || '',
                    address: lead.address || '',
                    notes: lead.notes || '',
                });
            } catch (err) {
                console.error('Failed to fetch lead:', err);
                error('Error', 'Failed to load lead data');
            } finally {
                setIsLoading(false);
            }
        };

        fetchLead();
    }, [id, error]);

    const handleChange = (field: keyof CreateLeadRequest, value: string) => {
        setFormData((prev) => ({ ...prev, [field]: value }));
        if (errors[field]) {
            setErrors((prev) => ({ ...prev, [field]: '' }));
        }
    };

    const validate = (): boolean => {
        const newErrors: Record<string, string> = {};

        if (!formData.first_name.trim()) {
            newErrors.first_name = 'First name is required';
        }
        if (!formData.last_name.trim()) {
            newErrors.last_name = 'Last name is required';
        }
        if (!formData.email.trim()) {
            newErrors.email = 'Email is required';
        } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
            newErrors.email = 'Please enter a valid email address';
        }
        if (formData.website && !/^https?:\/\/.+/.test(formData.website)) {
            newErrors.website = 'Please enter a valid URL (starting with http:// or https://)';
        }

        setErrors(newErrors);
        return Object.keys(newErrors).length === 0;
    };

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault();

        if (!validate()) return;

        setIsSaving(true);
        try {
            if (isEditing && id) {
                await leadService.updateLead(id, formData);
                success('Success', 'Lead updated successfully');
            } else {
                const newLead = await leadService.createLead(formData);
                success('Success', 'Lead created successfully');
                navigate(`/leads/${newLead.id}`);
                return;
            }
            navigate(`/leads/${id}`);
        } catch (err) {
            console.error('Failed to save lead:', err);
            error('Error', 'Failed to save lead');
        } finally {
            setIsSaving(false);
        }
    };

    if (isLoading) {
        return (
            <div className="animate-fade-in">
                <div className="page-header">
                    <div className="skeleton" style={{ width: '200px', height: '32px' }} />
                </div>
                <div className="skeleton" style={{ width: '100%', height: '600px', marginTop: '1rem' }} />
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

            {/* Page Header */}
            <div className="page-header">
                <div className="page-header-left">
                    <h1 className="page-title">{isEditing ? 'Edit Lead' : 'Create New Lead'}</h1>
                    <p className="page-description">
                        {isEditing ? 'Update the lead information below' : 'Fill in the details to add a new lead'}
                    </p>
                </div>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit}>
                <div className="grid grid-cols-3 gap-6">
                    {/* Main Information */}
                    <Card className="col-span-2" padding="lg">
                        <h3 className="font-semibold text-lg mb-4">Personal Information</h3>

                        <div className="grid grid-cols-2 gap-4">
                            <Input
                                label="First Name"
                                placeholder="Enter first name"
                                value={formData.first_name}
                                onChange={(e) => handleChange('first_name', e.target.value)}
                                error={errors.first_name}
                                required
                            />
                            <Input
                                label="Last Name"
                                placeholder="Enter last name"
                                value={formData.last_name}
                                onChange={(e) => handleChange('last_name', e.target.value)}
                                error={errors.last_name}
                                required
                            />
                        </div>

                        <div className="grid grid-cols-2 gap-4 mt-4">
                            <Input
                                type="email"
                                label="Email"
                                placeholder="email@company.com"
                                value={formData.email}
                                onChange={(e) => handleChange('email', e.target.value)}
                                error={errors.email}
                                required
                            />
                            <Input
                                type="tel"
                                label="Phone"
                                placeholder="+60123456789"
                                value={formData.phone}
                                onChange={(e) => handleChange('phone', e.target.value)}
                            />
                        </div>

                        <h3 className="font-semibold text-lg mt-8 mb-4">Company Information</h3>

                        <div className="grid grid-cols-2 gap-4">
                            <Input
                                label="Company"
                                placeholder="Company name"
                                value={formData.company}
                                onChange={(e) => handleChange('company', e.target.value)}
                            />
                            <Input
                                label="Job Title"
                                placeholder="e.g. Procurement Manager"
                                value={formData.title}
                                onChange={(e) => handleChange('title', e.target.value)}
                            />
                        </div>

                        <div className="grid grid-cols-2 gap-4 mt-4">
                            <Select
                                label="Industry"
                                options={industryOptions}
                                value={formData.industry}
                                onChange={(e) => handleChange('industry', e.target.value)}
                                placeholder="Select industry"
                            />
                            <Input
                                type="url"
                                label="Website"
                                placeholder="https://company.com"
                                value={formData.website}
                                onChange={(e) => handleChange('website', e.target.value)}
                                error={errors.website}
                            />
                        </div>

                        <div className="mt-4">
                            <Textarea
                                label="Address"
                                placeholder="Full address"
                                value={formData.address}
                                onChange={(e) => handleChange('address', e.target.value)}
                                rows={2}
                            />
                        </div>

                        <h3 className="font-semibold text-lg mt-8 mb-4">Additional Information</h3>

                        <Textarea
                            label="Notes"
                            placeholder="Add any additional notes about this lead..."
                            value={formData.notes}
                            onChange={(e) => handleChange('notes', e.target.value)}
                            rows={4}
                        />
                    </Card>

                    {/* Sidebar */}
                    <div className="flex flex-col gap-4">
                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Lead Source</h3>
                            <Select
                                label="How did they find us?"
                                options={sourceOptions}
                                value={formData.source}
                                onChange={(e) => handleChange('source', e.target.value)}
                                required
                            />
                        </Card>

                        <Card padding="lg">
                            <h3 className="font-semibold mb-4">Actions</h3>
                            <div className="flex flex-col gap-3">
                                <Button type="submit" fullWidth isLoading={isSaving}>
                                    {isEditing ? 'Update Lead' : 'Create Lead'}
                                </Button>
                                <Button
                                    type="button"
                                    variant="outline"
                                    fullWidth
                                    onClick={() => navigate(-1)}
                                >
                                    Cancel
                                </Button>
                            </div>
                        </Card>

                        {isEditing && (
                            <Card padding="lg">
                                <h3 className="font-semibold mb-4 text-danger">Danger Zone</h3>
                                <Button
                                    type="button"
                                    variant="danger"
                                    fullWidth
                                    onClick={async () => {
                                        if (confirm('Are you sure you want to delete this lead?')) {
                                            await leadService.deleteLead(id);
                                            success('Success', 'Lead deleted');
                                            navigate('/leads');
                                        }
                                    }}
                                >
                                    Delete Lead
                                </Button>
                            </Card>
                        )}
                    </div>
                </div>
            </form>
        </div>
    );
}

export default LeadFormPage;
