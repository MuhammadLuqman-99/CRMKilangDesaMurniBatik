// ============================================
// Pipelines Service
// Pipeline and deal management
// ============================================

import { HttpClient } from '../http';
import type {
    Pipeline,
    PipelineStage,
    Deal,
    CreatePipelineRequest,
    CreateStageRequest,
    PaginatedResponse,
    PaginationMeta,
    PaginationParams,
} from '../types';

interface PipelineListResponse {
    pipelines: Pipeline[];
    meta: PaginationMeta;
}

interface DealListResponse {
    deals: Deal[];
    meta: PaginationMeta;
}

interface PipelineAnalytics {
    totalDeals: number;
    totalValue: number;
    avgDealValue: number;
    avgTimeInStage: Record<string, number>;
    conversionRate: number;
    dealsByStage: Array<{
        stageId: string;
        stageName: string;
        count: number;
        value: number;
    }>;
}

export class PipelinesService {
    constructor(private http: HttpClient) {}

    // -------------------- Pipelines --------------------

    /**
     * Get all pipelines
     */
    async list(params?: PaginationParams): Promise<PaginatedResponse<Pipeline>> {
        const response = await this.http.get<PipelineListResponse>(
            '/sales/pipelines',
            params as Record<string, string | number | boolean | undefined>
        );
        return {
            data: response.pipelines,
            meta: response.meta,
        };
    }

    /**
     * Get a single pipeline by ID
     */
    async get(id: string): Promise<Pipeline> {
        return this.http.get<Pipeline>(`/sales/pipelines/${id}`);
    }

    /**
     * Create a new pipeline
     */
    async create(data: CreatePipelineRequest): Promise<Pipeline> {
        return this.http.post<Pipeline>('/sales/pipelines', data);
    }

    /**
     * Update a pipeline
     */
    async update(id: string, data: Partial<CreatePipelineRequest>): Promise<Pipeline> {
        return this.http.put<Pipeline>(`/sales/pipelines/${id}`, data);
    }

    /**
     * Delete a pipeline
     */
    async delete(id: string): Promise<void> {
        await this.http.delete(`/sales/pipelines/${id}`);
    }

    /**
     * Set a pipeline as default
     */
    async setDefault(id: string): Promise<Pipeline> {
        return this.http.post<Pipeline>(`/sales/pipelines/${id}/default`);
    }

    /**
     * Get pipeline analytics
     */
    async getAnalytics(id: string): Promise<PipelineAnalytics> {
        return this.http.get<PipelineAnalytics>(`/sales/pipelines/${id}/analytics`);
    }

    // -------------------- Stages --------------------

    /**
     * Add a stage to a pipeline
     */
    async createStage(pipelineId: string, data: CreateStageRequest): Promise<PipelineStage> {
        return this.http.post<PipelineStage>(`/sales/pipelines/${pipelineId}/stages`, data);
    }

    /**
     * Update a stage
     */
    async updateStage(
        pipelineId: string,
        stageId: string,
        data: Partial<CreateStageRequest>
    ): Promise<PipelineStage> {
        return this.http.put<PipelineStage>(`/sales/pipelines/${pipelineId}/stages/${stageId}`, data);
    }

    /**
     * Delete a stage
     */
    async deleteStage(pipelineId: string, stageId: string): Promise<void> {
        await this.http.delete(`/sales/pipelines/${pipelineId}/stages/${stageId}`);
    }

    /**
     * Reorder stages
     */
    async reorderStages(pipelineId: string, stageIds: string[]): Promise<Pipeline> {
        return this.http.post<Pipeline>(`/sales/pipelines/${pipelineId}/stages/reorder`, {
            stageIds,
        });
    }

    // -------------------- Deals --------------------

    /**
     * Get deals for a pipeline
     */
    async getDeals(
        pipelineId: string,
        params?: PaginationParams & { stageId?: string }
    ): Promise<PaginatedResponse<Deal>> {
        const response = await this.http.get<DealListResponse>(
            `/sales/pipelines/${pipelineId}/deals`,
            params as Record<string, string | number | boolean | undefined>
        );
        return {
            data: response.deals,
            meta: response.meta,
        };
    }

    /**
     * Get a single deal
     */
    async getDeal(dealId: string): Promise<Deal> {
        return this.http.get<Deal>(`/sales/deals/${dealId}`);
    }

    /**
     * Create a new deal
     */
    async createDeal(data: {
        name: string;
        customerId?: string;
        value: number;
        stageId: string;
        priority?: 'low' | 'medium' | 'high' | 'urgent';
        expectedCloseDate?: string;
        ownerId?: string;
        notes?: string;
    }): Promise<Deal> {
        return this.http.post<Deal>('/sales/deals', data);
    }

    /**
     * Update a deal
     */
    async updateDeal(
        dealId: string,
        data: Partial<{
            name: string;
            customerId: string;
            value: number;
            stageId: string;
            priority: 'low' | 'medium' | 'high' | 'urgent';
            expectedCloseDate: string;
            ownerId: string;
            notes: string;
        }>
    ): Promise<Deal> {
        return this.http.put<Deal>(`/sales/deals/${dealId}`, data);
    }

    /**
     * Delete a deal
     */
    async deleteDeal(dealId: string): Promise<void> {
        await this.http.delete(`/sales/deals/${dealId}`);
    }

    /**
     * Move a deal to a different stage
     */
    async moveDeal(dealId: string, stageId: string, position?: number): Promise<Deal> {
        return this.http.post<Deal>(`/sales/deals/${dealId}/move`, {
            stageId,
            position,
        });
    }

    /**
     * Mark deal as won
     */
    async winDeal(dealId: string, notes?: string): Promise<Deal> {
        return this.http.post<Deal>(`/sales/deals/${dealId}/win`, { notes });
    }

    /**
     * Mark deal as lost
     */
    async loseDeal(dealId: string, reason: string, notes?: string): Promise<Deal> {
        return this.http.post<Deal>(`/sales/deals/${dealId}/lose`, { reason, notes });
    }
}
