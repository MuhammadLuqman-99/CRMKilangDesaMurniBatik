// ============================================
// Dashboard Service
// Production-Ready Dashboard API Methods
// ============================================

import { api } from './api';
import type {
    DashboardStats,
    PipelineOverview,
    RecentActivity,
} from '../types';

interface DashboardDataResponse {
    stats: DashboardStats;
    pipeline_overview: PipelineOverview;
    recent_activities: RecentActivity[];
}

export const dashboardService = {
    /**
     * Get complete dashboard data
     */
    getDashboardData: async (): Promise<DashboardDataResponse> => {
        return api.get<DashboardDataResponse>('/dashboard');
    },

    /**
     * Get dashboard statistics
     */
    getStats: async (): Promise<DashboardStats> => {
        return api.get<DashboardStats>('/dashboard/stats');
    },

    /**
     * Get pipeline overview
     */
    getPipelineOverview: async (pipelineId?: string): Promise<PipelineOverview> => {
        const params = pipelineId ? { pipeline_id: pipelineId } : undefined;
        return api.get<PipelineOverview>('/dashboard/pipeline-overview', params);
    },

    /**
     * Get recent activities
     */
    getRecentActivities: async (limit: number = 10): Promise<RecentActivity[]> => {
        return api.get<RecentActivity[]>('/dashboard/activities', { limit });
    },

    /**
     * Get upcoming tasks
     */
    getUpcomingTasks: async (days: number = 7): Promise<RecentActivity[]> => {
        return api.get<RecentActivity[]>('/dashboard/upcoming-tasks', { days });
    },

    /**
     * Get deals closing soon
     */
    getDealsClosingSoon: async (days: number = 30): Promise<{ deals: unknown[] }> => {
        return api.get<{ deals: unknown[] }>('/dashboard/closing-soon', { days });
    },
};

export default dashboardService;
