// ============================================
// Leads Service
// Lead management operations
// ============================================

import { HttpClient } from '../http';
import type {
    Lead,
    LeadFilters,
    CreateLeadRequest,
    UpdateLeadRequest,
    ConvertLeadRequest,
    ConvertLeadResponse,
    QualifyLeadRequest,
    DisqualifyLeadRequest,
    PaginatedResponse,
    PaginationMeta,
} from '../types';

interface LeadListResponse {
    leads: Lead[];
    meta: PaginationMeta;
}

export class LeadsService {
    constructor(private http: HttpClient) {}

    /**
     * Get a paginated list of leads
     */
    async list(filters?: LeadFilters): Promise<PaginatedResponse<Lead>> {
        const response = await this.http.get<LeadListResponse>('/sales/leads', filters as Record<string, string | number | boolean | undefined>);
        return {
            data: response.leads,
            meta: response.meta,
        };
    }

    /**
     * Get a single lead by ID
     */
    async get(id: string): Promise<Lead> {
        return this.http.get<Lead>(`/sales/leads/${id}`);
    }

    /**
     * Create a new lead
     */
    async create(data: CreateLeadRequest): Promise<Lead> {
        return this.http.post<Lead>('/sales/leads', data);
    }

    /**
     * Update an existing lead
     */
    async update(id: string, data: UpdateLeadRequest): Promise<Lead> {
        return this.http.put<Lead>(`/sales/leads/${id}`, data);
    }

    /**
     * Delete a lead
     */
    async delete(id: string): Promise<void> {
        await this.http.delete(`/sales/leads/${id}`);
    }

    /**
     * Qualify a lead
     */
    async qualify(id: string, data?: QualifyLeadRequest): Promise<Lead> {
        return this.http.post<Lead>(`/sales/leads/${id}/qualify`, data);
    }

    /**
     * Disqualify a lead
     */
    async disqualify(id: string, data: DisqualifyLeadRequest): Promise<Lead> {
        return this.http.post<Lead>(`/sales/leads/${id}/disqualify`, data);
    }

    /**
     * Convert a lead to opportunity/customer
     */
    async convert(id: string, data: ConvertLeadRequest): Promise<ConvertLeadResponse> {
        return this.http.post<ConvertLeadResponse>(`/sales/leads/${id}/convert`, data);
    }

    /**
     * Assign a lead to a user
     */
    async assign(id: string, userId: string): Promise<Lead> {
        return this.http.post<Lead>(`/sales/leads/${id}/assign`, { userId });
    }

    /**
     * Get lead statistics
     */
    async getStatistics(): Promise<{
        total: number;
        byStatus: Record<string, number>;
        bySource: Record<string, number>;
        conversionRate: number;
    }> {
        return this.http.get('/sales/leads/statistics');
    }

    /**
     * Import leads from file (returns import job ID)
     */
    async import(file: Blob, mappings: Record<string, string>): Promise<{ jobId: string }> {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('mappings', JSON.stringify(mappings));

        // Note: This would need special handling for FormData
        return this.http.post<{ jobId: string }>('/sales/leads/import', {
            mappings,
        });
    }

    /**
     * Export leads to file
     */
    async export(filters?: LeadFilters, format: 'csv' | 'xlsx' = 'csv'): Promise<Blob> {
        const response = await fetch(
            `${this.http['config'].baseUrl}/api/${this.http['config'].apiVersion}/sales/leads/export?format=${format}`,
            {
                method: 'GET',
                headers: {
                    Authorization: `Bearer ${this.http.getAccessToken()}`,
                },
            }
        );

        if (!response.ok) {
            throw new Error('Export failed');
        }

        return response.blob();
    }
}
