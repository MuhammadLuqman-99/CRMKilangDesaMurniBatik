// ============================================
// CRM Client
// Main SDK entry point
// ============================================

import { HttpClient } from './http';
import {
    AuthService,
    LeadsService,
    CustomersService,
    OpportunitiesService,
    PipelinesService,
    DashboardService,
} from './services';
import type { CRMClientConfig } from './types';

/**
 * CRM SDK Client
 *
 * Main entry point for interacting with the CRM API.
 *
 * @example
 * ```typescript
 * import { CRMClient } from '@kilangbatik/crm-sdk';
 *
 * const crm = new CRMClient({
 *   baseUrl: 'https://api.example.com',
 * });
 *
 * // Login
 * const { user } = await crm.auth.login({
 *   email: 'user@example.com',
 *   password: 'password123',
 * });
 *
 * // Get leads
 * const { data: leads } = await crm.leads.list({ status: 'qualified' });
 *
 * // Create a customer
 * const customer = await crm.customers.create({
 *   name: 'Acme Corp',
 *   email: 'contact@acme.com',
 * });
 * ```
 */
export class CRMClient {
    private http: HttpClient;

    /** Authentication service */
    public readonly auth: AuthService;

    /** Leads management */
    public readonly leads: LeadsService;

    /** Customers management */
    public readonly customers: CustomersService;

    /** Opportunities management */
    public readonly opportunities: OpportunitiesService;

    /** Pipelines and deals management */
    public readonly pipelines: PipelinesService;

    /** Dashboard data and analytics */
    public readonly dashboard: DashboardService;

    /**
     * Create a new CRM client instance
     *
     * @param config - Client configuration
     */
    constructor(config: CRMClientConfig) {
        this.http = new HttpClient(config);

        this.auth = new AuthService(this.http);
        this.leads = new LeadsService(this.http);
        this.customers = new CustomersService(this.http);
        this.opportunities = new OpportunitiesService(this.http);
        this.pipelines = new PipelinesService(this.http);
        this.dashboard = new DashboardService(this.http);
    }

    /**
     * Check if the client is authenticated
     */
    isAuthenticated(): boolean {
        return this.auth.isAuthenticated();
    }

    /**
     * Set authentication tokens manually
     * Useful for restoring tokens from storage
     *
     * @param accessToken - JWT access token
     * @param refreshToken - Optional refresh token
     */
    setTokens(accessToken: string, refreshToken?: string): void {
        this.auth.setTokens(accessToken, refreshToken);
    }

    /**
     * Clear authentication tokens
     */
    clearTokens(): void {
        this.auth.clearTokens();
    }
}

/**
 * Create a new CRM client instance
 *
 * @param config - Client configuration
 * @returns CRM client instance
 *
 * @example
 * ```typescript
 * import { createClient } from '@kilangbatik/crm-sdk';
 *
 * const crm = createClient({
 *   baseUrl: 'https://api.example.com',
 * });
 * ```
 */
export function createClient(config: CRMClientConfig): CRMClient {
    return new CRMClient(config);
}
