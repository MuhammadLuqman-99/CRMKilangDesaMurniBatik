// ============================================
// Customers Service
// Customer management operations
// ============================================

import { HttpClient } from '../http';
import type {
    Customer,
    CustomerFilters,
    CreateCustomerRequest,
    UpdateCustomerRequest,
    CustomerContact,
    CreateContactRequest,
    CustomerNote,
    CreateNoteRequest,
    CustomerActivity,
    PaginatedResponse,
    PaginationMeta,
    PaginationParams,
} from '../types';

interface CustomerListResponse {
    customers: Customer[];
    meta: PaginationMeta;
}

interface ContactListResponse {
    contacts: CustomerContact[];
    meta: PaginationMeta;
}

interface NoteListResponse {
    notes: CustomerNote[];
    meta: PaginationMeta;
}

interface ActivityListResponse {
    activities: CustomerActivity[];
    meta: PaginationMeta;
}

export class CustomersService {
    constructor(private http: HttpClient) {}

    // -------------------- Customer CRUD --------------------

    /**
     * Get a paginated list of customers
     */
    async list(filters?: CustomerFilters): Promise<PaginatedResponse<Customer>> {
        const response = await this.http.get<CustomerListResponse>(
            '/customers',
            filters as Record<string, string | number | boolean | undefined>
        );
        return {
            data: response.customers,
            meta: response.meta,
        };
    }

    /**
     * Get a single customer by ID
     */
    async get(id: string): Promise<Customer> {
        return this.http.get<Customer>(`/customers/${id}`);
    }

    /**
     * Create a new customer
     */
    async create(data: CreateCustomerRequest): Promise<Customer> {
        return this.http.post<Customer>('/customers', data);
    }

    /**
     * Update an existing customer
     */
    async update(id: string, data: UpdateCustomerRequest): Promise<Customer> {
        return this.http.put<Customer>(`/customers/${id}`, data);
    }

    /**
     * Delete a customer
     */
    async delete(id: string): Promise<void> {
        await this.http.delete(`/customers/${id}`);
    }

    /**
     * Restore a deleted customer
     */
    async restore(id: string): Promise<Customer> {
        return this.http.post<Customer>(`/customers/${id}/restore`);
    }

    /**
     * Activate a customer
     */
    async activate(id: string): Promise<Customer> {
        return this.http.post<Customer>(`/customers/${id}/activate`);
    }

    /**
     * Deactivate a customer
     */
    async deactivate(id: string): Promise<Customer> {
        return this.http.post<Customer>(`/customers/${id}/deactivate`);
    }

    // -------------------- Contacts --------------------

    /**
     * Get contacts for a customer
     */
    async getContacts(customerId: string, params?: PaginationParams): Promise<PaginatedResponse<CustomerContact>> {
        const response = await this.http.get<ContactListResponse>(
            `/customers/${customerId}/contacts`,
            params as Record<string, string | number | boolean | undefined>
        );
        return {
            data: response.contacts,
            meta: response.meta,
        };
    }

    /**
     * Get a single contact
     */
    async getContact(customerId: string, contactId: string): Promise<CustomerContact> {
        return this.http.get<CustomerContact>(`/customers/${customerId}/contacts/${contactId}`);
    }

    /**
     * Add a contact to a customer
     */
    async createContact(customerId: string, data: CreateContactRequest): Promise<CustomerContact> {
        return this.http.post<CustomerContact>(`/customers/${customerId}/contacts`, data);
    }

    /**
     * Update a contact
     */
    async updateContact(
        customerId: string,
        contactId: string,
        data: Partial<CreateContactRequest>
    ): Promise<CustomerContact> {
        return this.http.put<CustomerContact>(`/customers/${customerId}/contacts/${contactId}`, data);
    }

    /**
     * Delete a contact
     */
    async deleteContact(customerId: string, contactId: string): Promise<void> {
        await this.http.delete(`/customers/${customerId}/contacts/${contactId}`);
    }

    /**
     * Set a contact as primary
     */
    async setPrimaryContact(customerId: string, contactId: string): Promise<CustomerContact> {
        return this.http.post<CustomerContact>(`/customers/${customerId}/contacts/${contactId}/primary`);
    }

    // -------------------- Notes --------------------

    /**
     * Get notes for a customer
     */
    async getNotes(customerId: string, params?: PaginationParams): Promise<PaginatedResponse<CustomerNote>> {
        const response = await this.http.get<NoteListResponse>(
            `/customers/${customerId}/notes`,
            params as Record<string, string | number | boolean | undefined>
        );
        return {
            data: response.notes,
            meta: response.meta,
        };
    }

    /**
     * Add a note to a customer
     */
    async createNote(customerId: string, data: CreateNoteRequest): Promise<CustomerNote> {
        return this.http.post<CustomerNote>(`/customers/${customerId}/notes`, data);
    }

    /**
     * Update a note
     */
    async updateNote(customerId: string, noteId: string, data: CreateNoteRequest): Promise<CustomerNote> {
        return this.http.put<CustomerNote>(`/customers/${customerId}/notes/${noteId}`, data);
    }

    /**
     * Delete a note
     */
    async deleteNote(customerId: string, noteId: string): Promise<void> {
        await this.http.delete(`/customers/${customerId}/notes/${noteId}`);
    }

    /**
     * Pin a note
     */
    async pinNote(customerId: string, noteId: string): Promise<CustomerNote> {
        return this.http.post<CustomerNote>(`/customers/${customerId}/notes/${noteId}/pin`);
    }

    /**
     * Unpin a note
     */
    async unpinNote(customerId: string, noteId: string): Promise<CustomerNote> {
        return this.http.post<CustomerNote>(`/customers/${customerId}/notes/${noteId}/unpin`);
    }

    // -------------------- Activities --------------------

    /**
     * Get activities for a customer
     */
    async getActivities(
        customerId: string,
        params?: PaginationParams & { type?: string }
    ): Promise<PaginatedResponse<CustomerActivity>> {
        const response = await this.http.get<ActivityListResponse>(
            `/customers/${customerId}/activities`,
            params as Record<string, string | number | boolean | undefined>
        );
        return {
            data: response.activities,
            meta: response.meta,
        };
    }

    /**
     * Log an activity for a customer
     */
    async logActivity(
        customerId: string,
        data: { type: string; title: string; description?: string }
    ): Promise<CustomerActivity> {
        return this.http.post<CustomerActivity>(`/customers/${customerId}/activities`, data);
    }

    // -------------------- Export --------------------

    /**
     * Export customers to file
     */
    async export(filters?: CustomerFilters, format: 'csv' | 'xlsx' = 'csv'): Promise<Blob> {
        const params = new URLSearchParams();
        params.append('format', format);

        if (filters) {
            Object.entries(filters).forEach(([key, value]) => {
                if (value !== undefined) {
                    params.append(key, String(value));
                }
            });
        }

        const response = await fetch(
            `${this.http['config'].baseUrl}/api/${this.http['config'].apiVersion}/customers/export?${params}`,
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
