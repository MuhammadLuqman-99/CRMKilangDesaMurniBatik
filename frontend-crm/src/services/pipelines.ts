// ============================================
// Pipeline Service
// Production-Ready Pipeline & Deal API Methods
// ============================================

import { api } from './api';
import type {
    Pipeline,
    PipelineStage,
    Deal,
    CreatePipelineRequest,
    UpdatePipelineRequest,
    CreatePipelineStageRequest,
    PipelineAnalytics,
    MoveDealRequest,
} from '../types';

interface PipelineListResponse {
    pipelines: Pipeline[];
}

interface DealListResponse {
    deals: Deal[];
}

export const pipelineService = {
    // ============================================
    // Pipelines
    // ============================================

    /**
     * Get all pipelines
     */
    getPipelines: async (): Promise<PipelineListResponse> => {
        return api.get<PipelineListResponse>('/pipelines');
    },

    /**
     * Get a single pipeline by ID
     */
    getPipeline: async (id: string): Promise<Pipeline> => {
        return api.get<Pipeline>(`/pipelines/${id}`);
    },

    /**
     * Create a new pipeline
     */
    createPipeline: async (data: CreatePipelineRequest): Promise<Pipeline> => {
        return api.post<Pipeline>('/pipelines', data);
    },

    /**
     * Update an existing pipeline
     */
    updatePipeline: async (id: string, data: UpdatePipelineRequest): Promise<Pipeline> => {
        return api.put<Pipeline>(`/pipelines/${id}`, data);
    },

    /**
     * Delete a pipeline
     */
    deletePipeline: async (id: string): Promise<void> => {
        return api.delete(`/pipelines/${id}`);
    },

    /**
     * Get pipeline analytics
     */
    getPipelineAnalytics: async (id: string): Promise<PipelineAnalytics> => {
        return api.get<PipelineAnalytics>(`/pipelines/${id}/analytics`);
    },

    // ============================================
    // Pipeline Stages
    // ============================================

    /**
     * Add a stage to a pipeline
     */
    addStage: async (pipelineId: string, data: CreatePipelineStageRequest): Promise<PipelineStage> => {
        return api.post<PipelineStage>(`/pipelines/${pipelineId}/stages`, data);
    },

    /**
     * Update a pipeline stage
     */
    updateStage: async (
        pipelineId: string,
        stageId: string,
        data: Partial<CreatePipelineStageRequest>
    ): Promise<PipelineStage> => {
        return api.put<PipelineStage>(`/pipelines/${pipelineId}/stages/${stageId}`, data);
    },

    /**
     * Delete a pipeline stage
     */
    deleteStage: async (pipelineId: string, stageId: string): Promise<void> => {
        return api.delete(`/pipelines/${pipelineId}/stages/${stageId}`);
    },

    /**
     * Reorder pipeline stages
     */
    reorderStages: async (pipelineId: string, stageIds: string[]): Promise<Pipeline> => {
        return api.post<Pipeline>(`/pipelines/${pipelineId}/stages/reorder`, { stage_ids: stageIds });
    },

    // ============================================
    // Deals (Kanban Board)
    // ============================================

    /**
     * Get all deals for a pipeline
     */
    getDeals: async (pipelineId: string): Promise<DealListResponse> => {
        return api.get<DealListResponse>('/deals', { pipeline_id: pipelineId });
    },

    /**
     * Get a single deal by ID
     */
    getDeal: async (id: string): Promise<Deal> => {
        return api.get<Deal>(`/deals/${id}`);
    },

    /**
     * Move a deal to a different stage
     */
    moveDeal: async (dealId: string, data: MoveDealRequest): Promise<Deal> => {
        return api.post<Deal>(`/deals/${dealId}/move`, data);
    },

    /**
     * Update deal details
     */
    updateDeal: async (id: string, data: Partial<Deal>): Promise<Deal> => {
        return api.put<Deal>(`/deals/${id}`, data);
    },

    /**
     * Generate invoice for a deal
     */
    generateInvoice: async (dealId: string): Promise<{ invoice_url: string }> => {
        return api.post<{ invoice_url: string }>(`/deals/${dealId}/invoice`);
    },

    /**
     * Record payment for a deal
     */
    recordPayment: async (
        dealId: string,
        amount: number,
        paymentMethod: string,
        reference?: string
    ): Promise<void> => {
        return api.post(`/deals/${dealId}/payment`, {
            amount,
            payment_method: paymentMethod,
            reference,
        });
    },
};

export default pipelineService;
