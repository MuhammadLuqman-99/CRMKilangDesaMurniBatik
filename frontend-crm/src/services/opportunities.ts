// ============================================
// Opportunity Service
// Production-Ready Opportunity API Methods
// ============================================

import { api } from './api';
import type {
    Opportunity,
    CreateOpportunityRequest,
    UpdateOpportunityRequest,
    MoveStageRequest,
    WinOpportunityRequest,
    LoseOpportunityRequest,
    OpportunityFilters,
    PaginationMeta,
} from '../types';

interface OpportunityListResponse {
    opportunities: Opportunity[];
    meta: PaginationMeta;
}

export const opportunityService = {
    /**
     * Get paginated list of opportunities with filters
     */
    getOpportunities: async (filters?: OpportunityFilters): Promise<OpportunityListResponse> => {
        const params: Record<string, string | number | boolean | undefined> = {
            page: filters?.page,
            per_page: filters?.per_page,
            sort: filters?.sort,
            order: filters?.order,
            status: filters?.status,
            pipeline_id: filters?.pipeline_id,
            stage_id: filters?.stage_id,
            owner_id: filters?.owner_id,
            customer_id: filters?.customer_id,
            search: filters?.search,
            value_min: filters?.value_min,
            value_max: filters?.value_max,
            expected_close_from: filters?.expected_close_from,
            expected_close_to: filters?.expected_close_to,
        };

        return api.get<OpportunityListResponse>('/opportunities', params);
    },

    /**
     * Get a single opportunity by ID
     */
    getOpportunity: async (id: string): Promise<Opportunity> => {
        return api.get<Opportunity>(`/opportunities/${id}`);
    },

    /**
     * Create a new opportunity
     */
    createOpportunity: async (data: CreateOpportunityRequest): Promise<Opportunity> => {
        return api.post<Opportunity>('/opportunities', data);
    },

    /**
     * Update an existing opportunity
     */
    updateOpportunity: async (id: string, data: UpdateOpportunityRequest): Promise<Opportunity> => {
        return api.put<Opportunity>(`/opportunities/${id}`, data);
    },

    /**
     * Delete an opportunity
     */
    deleteOpportunity: async (id: string): Promise<void> => {
        return api.delete(`/opportunities/${id}`);
    },

    /**
     * Move opportunity to a different stage
     */
    moveStage: async (id: string, data: MoveStageRequest): Promise<Opportunity> => {
        return api.post<Opportunity>(`/opportunities/${id}/move-stage`, data);
    },

    /**
     * Mark opportunity as won
     */
    winOpportunity: async (id: string, data?: WinOpportunityRequest): Promise<Opportunity> => {
        return api.post<Opportunity>(`/opportunities/${id}/win`, data);
    },

    /**
     * Mark opportunity as lost
     */
    loseOpportunity: async (id: string, data: LoseOpportunityRequest): Promise<Opportunity> => {
        return api.post<Opportunity>(`/opportunities/${id}/lose`, data);
    },

    /**
     * Reopen a closed opportunity
     */
    reopenOpportunity: async (id: string): Promise<Opportunity> => {
        return api.post<Opportunity>(`/opportunities/${id}/reopen`);
    },

    /**
     * Add a contact to an opportunity
     */
    addContact: async (opportunityId: string, contactId: string): Promise<void> => {
        return api.post(`/opportunities/${opportunityId}/contacts`, { contact_id: contactId });
    },

    /**
     * Remove a contact from an opportunity
     */
    removeContact: async (opportunityId: string, contactId: string): Promise<void> => {
        return api.delete(`/opportunities/${opportunityId}/contacts/${contactId}`);
    },

    /**
     * Add a product to an opportunity
     */
    addProduct: async (
        opportunityId: string,
        productId: string,
        quantity: number,
        unitPrice: number,
        discountPercent?: number
    ): Promise<void> => {
        return api.post(`/opportunities/${opportunityId}/products`, {
            product_id: productId,
            quantity,
            unit_price: unitPrice,
            discount_percent: discountPercent || 0,
        });
    },

    /**
     * Remove a product from an opportunity
     */
    removeProduct: async (opportunityId: string, productId: string): Promise<void> => {
        return api.delete(`/opportunities/${opportunityId}/products/${productId}`);
    },
};

export default opportunityService;
