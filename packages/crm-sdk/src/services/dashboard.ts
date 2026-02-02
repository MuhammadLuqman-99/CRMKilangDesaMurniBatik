// ============================================
// Dashboard Service
// Dashboard data and analytics
// ============================================

import { HttpClient } from '../http';
import type { DashboardData, DashboardStats, PipelineOverview, RecentActivity } from '../types';

export class DashboardService {
    constructor(private http: HttpClient) {}

    /**
     * Get complete dashboard data
     */
    async getData(): Promise<DashboardData> {
        return this.http.get<DashboardData>('/dashboard');
    }

    /**
     * Get dashboard statistics
     */
    async getStats(): Promise<DashboardStats> {
        return this.http.get<DashboardStats>('/dashboard/stats');
    }

    /**
     * Get pipeline overview
     */
    async getPipelineOverview(): Promise<PipelineOverview> {
        return this.http.get<PipelineOverview>('/dashboard/pipeline-overview');
    }

    /**
     * Get recent activities
     */
    async getRecentActivities(limit?: number): Promise<RecentActivity[]> {
        const response = await this.http.get<{ activities: RecentActivity[] }>(
            '/dashboard/recent-activities',
            { limit }
        );
        return response.activities;
    }

    /**
     * Get deals closing soon
     */
    async getDealsClosingSoon(days?: number): Promise<{
        deals: Array<{
            id: string;
            name: string;
            value: number;
            expectedCloseDate: string;
            customerName?: string;
            stageName?: string;
        }>;
        total: number;
    }> {
        return this.http.get('/dashboard/deals-closing-soon', { days });
    }

    /**
     * Get sales forecast
     */
    async getSalesForecast(months?: number): Promise<{
        forecast: Array<{
            month: string;
            projected: number;
            committed: number;
            closed: number;
        }>;
        total: {
            projected: number;
            committed: number;
            closed: number;
        };
    }> {
        return this.http.get('/dashboard/sales-forecast', { months });
    }

    /**
     * Get team performance metrics
     */
    async getTeamPerformance(): Promise<{
        members: Array<{
            userId: string;
            userName: string;
            leadsCount: number;
            dealsCount: number;
            dealsValue: number;
            winRate: number;
        }>;
        totals: {
            leadsCount: number;
            dealsCount: number;
            dealsValue: number;
            avgWinRate: number;
        };
    }> {
        return this.http.get('/dashboard/team-performance');
    }
}
