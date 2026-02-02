// ============================================
// Lead Service
// Production-Ready Lead API Methods
// ============================================

import { api } from './api';
import type {
    Lead,
    CreateLeadRequest,
    UpdateLeadRequest,
    ConvertLeadRequest,
    QualifyLeadRequest,
    DisqualifyLeadRequest,
    LeadFilters,
    PaginationMeta,
} from '../types';

interface LeadListResponse {
    leads: Lead[];
    meta: PaginationMeta;
}

export const leadService = {
    /**
     * Get paginated list of leads with filters
     */
    getLeads: async (filters?: LeadFilters): Promise<LeadListResponse> => {
        const params: Record<string, string | number | boolean | undefined> = {
            page: filters?.page,
            per_page: filters?.per_page,
            sort: filters?.sort,
            order: filters?.order,
            status: filters?.status,
            source: filters?.source,
            score_label: filters?.score_label,
            owner_id: filters?.owner_id,
            search: filters?.search,
            created_from: filters?.created_from,
            created_to: filters?.created_to,
        };

        return api.get<LeadListResponse>('/leads', params);
    },

    /**
     * Get a single lead by ID
     */
    getLead: async (id: string): Promise<Lead> => {
        return api.get<Lead>(`/leads/${id}`);
    },

    /**
     * Create a new lead
     */
    createLead: async (data: CreateLeadRequest): Promise<Lead> => {
        return api.post<Lead>('/leads', data);
    },

    /**
     * Update an existing lead
     */
    updateLead: async (id: string, data: UpdateLeadRequest): Promise<Lead> => {
        return api.put<Lead>(`/leads/${id}`, data);
    },

    /**
     * Delete a lead
     */
    deleteLead: async (id: string): Promise<void> => {
        return api.delete(`/leads/${id}`);
    },

    /**
     * Convert lead to opportunity and/or customer
     */
    convertLead: async (id: string, data: ConvertLeadRequest): Promise<{ opportunity_id?: string; customer_id?: string }> => {
        return api.post<{ opportunity_id?: string; customer_id?: string }>(`/leads/${id}/convert`, data);
    },

    /**
     * Qualify a lead
     */
    qualifyLead: async (id: string, data: QualifyLeadRequest): Promise<Lead> => {
        return api.post<Lead>(`/leads/${id}/qualify`, data);
    },

    /**
     * Disqualify a lead
     */
    disqualifyLead: async (id: string, data: DisqualifyLeadRequest): Promise<Lead> => {
        return api.post<Lead>(`/leads/${id}/disqualify`, data);
    },

    /**
     * Bulk import leads
     */
    importLeads: async (file: File, mappings: Record<string, string>): Promise<{ import_id: string }> => {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('mappings', JSON.stringify(mappings));

        const response = await fetch('/api/v1/leads/import', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('crm_access_token')}`,
            },
            body: formData,
        });

        if (!response.ok) {
            throw new Error('Failed to import leads');
        }

        const data = await response.json();
        return data.data;
    },

    /**
     * Export leads
     */
    exportLeads: async (filters?: LeadFilters, format: 'csv' | 'xlsx' = 'csv'): Promise<Blob> => {
        const params = new URLSearchParams();
        params.append('format', format);

        if (filters?.status) params.append('status', filters.status);
        if (filters?.source) params.append('source', filters.source);
        if (filters?.search) params.append('search', filters.search);

        const response = await fetch(`/api/v1/leads/export?${params.toString()}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('crm_access_token')}`,
            },
        });

        if (!response.ok) {
            throw new Error('Failed to export leads');
        }

        return response.blob();
    },
};

export default leadService;
