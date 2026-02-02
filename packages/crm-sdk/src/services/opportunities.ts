// ============================================
// Opportunities Service
// Opportunity management operations
// ============================================

import { HttpClient } from '../http';
import type {
    Opportunity,
    OpportunityFilters,
    CreateOpportunityRequest,
    UpdateOpportunityRequest,
    MoveStageRequest,
    WinOpportunityRequest,
    LoseOpportunityRequest,
    PaginatedResponse,
    PaginationMeta,
} from '../types';

interface OpportunityListResponse {
    opportunities: Opportunity[];
    meta: PaginationMeta;
}

export class OpportunitiesService {
    constructor(private http: HttpClient) {}

    /**
     * Get a paginated list of opportunities
     */
    async list(filters?: OpportunityFilters): Promise<PaginatedResponse<Opportunity>> {
        const response = await this.http.get<OpportunityListResponse>(
            '/sales/opportunities',
            filters as Record<string, string | number | boolean | undefined>
        );
        return {
            data: response.opportunities,
            meta: response.meta,
        };
    }

    /**
     * Get a single opportunity by ID
     */
    async get(id: string): Promise<Opportunity> {
        return this.http.get<Opportunity>(`/sales/opportunities/${id}`);
    }

    /**
     * Create a new opportunity
     */
    async create(data: CreateOpportunityRequest): Promise<Opportunity> {
        return this.http.post<Opportunity>('/sales/opportunities', data);
    }

    /**
     * Update an existing opportunity
     */
    async update(id: string, data: UpdateOpportunityRequest): Promise<Opportunity> {
        return this.http.put<Opportunity>(`/sales/opportunities/${id}`, data);
    }

    /**
     * Delete an opportunity
     */
    async delete(id: string): Promise<void> {
        await this.http.delete(`/sales/opportunities/${id}`);
    }

    /**
     * Move opportunity to a different stage
     */
    async moveStage(id: string, data: MoveStageRequest): Promise<Opportunity> {
        return this.http.post<Opportunity>(`/sales/opportunities/${id}/move-stage`, data);
    }

    /**
     * Mark opportunity as won
     */
    async win(id: string, data?: WinOpportunityRequest): Promise<Opportunity> {
        return this.http.post<Opportunity>(`/sales/opportunities/${id}/win`, data);
    }

    /**
     * Mark opportunity as lost
     */
    async lose(id: string, data: LoseOpportunityRequest): Promise<Opportunity> {
        return this.http.post<Opportunity>(`/sales/opportunities/${id}/lose`, data);
    }

    /**
     * Reopen a closed opportunity
     */
    async reopen(id: string): Promise<Opportunity> {
        return this.http.post<Opportunity>(`/sales/opportunities/${id}/reopen`);
    }

    /**
     * Add a contact to an opportunity
     */
    async addContact(opportunityId: string, contactId: string, role?: string): Promise<void> {
        await this.http.post(`/sales/opportunities/${opportunityId}/contacts`, {
            contactId,
            role,
        });
    }

    /**
     * Remove a contact from an opportunity
     */
    async removeContact(opportunityId: string, contactId: string): Promise<void> {
        await this.http.delete(`/sales/opportunities/${opportunityId}/contacts/${contactId}`);
    }

    /**
     * Add a product to an opportunity
     */
    async addProduct(
        opportunityId: string,
        data: {
            productId: string;
            quantity: number;
            unitPrice: number;
            discountPercent?: number;
        }
    ): Promise<void> {
        await this.http.post(`/sales/opportunities/${opportunityId}/products`, data);
    }

    /**
     * Update a product in an opportunity
     */
    async updateProduct(
        opportunityId: string,
        productId: string,
        data: {
            quantity?: number;
            unitPrice?: number;
            discountPercent?: number;
        }
    ): Promise<void> {
        await this.http.put(`/sales/opportunities/${opportunityId}/products/${productId}`, data);
    }

    /**
     * Remove a product from an opportunity
     */
    async removeProduct(opportunityId: string, productId: string): Promise<void> {
        await this.http.delete(`/sales/opportunities/${opportunityId}/products/${productId}`);
    }

    /**
     * Assign opportunity to a user
     */
    async assign(id: string, userId: string): Promise<Opportunity> {
        return this.http.post<Opportunity>(`/sales/opportunities/${id}/assign`, { userId });
    }

    /**
     * Get opportunity statistics
     */
    async getStatistics(): Promise<{
        total: number;
        totalValue: number;
        byStatus: Record<string, { count: number; value: number }>;
        byStage: Record<string, { count: number; value: number }>;
        winRate: number;
        avgDealSize: number;
        avgCycleTime: number;
    }> {
        return this.http.get('/sales/opportunities/statistics');
    }
}
