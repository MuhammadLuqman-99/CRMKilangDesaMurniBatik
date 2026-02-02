// ============================================
// Customer Service
// Production-Ready Customer API Methods
// ============================================

import { api } from './api';
import type {
    Customer,
    Contact,
    CreateCustomerRequest,
    UpdateCustomerRequest,
    CreateContactRequest,
    UpdateContactRequest,
    CustomerFilters,
    Note,
    CreateNoteRequest,
    Activity,
    ActivityFilters,
    PaginationMeta,
} from '../types';

interface CustomerListResponse {
    customers: Customer[];
    meta: PaginationMeta;
}

interface ContactListResponse {
    contacts: Contact[];
}

interface NoteListResponse {
    notes: Note[];
}

interface ActivityListResponse {
    activities: Activity[];
    meta: PaginationMeta;
}

export const customerService = {
    // ============================================
    // Customer CRUD
    // ============================================

    /**
     * Get paginated list of customers with filters
     */
    getCustomers: async (filters?: CustomerFilters): Promise<CustomerListResponse> => {
        const params: Record<string, string | number | boolean | undefined> = {
            page: filters?.page,
            per_page: filters?.per_page,
            sort: filters?.sort,
            order: filters?.order,
            type: filters?.type,
            status: filters?.status,
            owner_id: filters?.owner_id,
            search: filters?.search,
            industry: filters?.industry,
            created_from: filters?.created_from,
            created_to: filters?.created_to,
        };

        return api.get<CustomerListResponse>('/customers', params);
    },

    /**
     * Get a single customer by ID
     */
    getCustomer: async (id: string): Promise<Customer> => {
        return api.get<Customer>(`/customers/${id}`);
    },

    /**
     * Create a new customer
     */
    createCustomer: async (data: CreateCustomerRequest): Promise<Customer> => {
        return api.post<Customer>('/customers', data);
    },

    /**
     * Update an existing customer
     */
    updateCustomer: async (id: string, data: UpdateCustomerRequest): Promise<Customer> => {
        return api.put<Customer>(`/customers/${id}`, data);
    },

    /**
     * Delete a customer
     */
    deleteCustomer: async (id: string): Promise<void> => {
        return api.delete(`/customers/${id}`);
    },

    /**
     * Restore a deleted customer
     */
    restoreCustomer: async (id: string): Promise<Customer> => {
        return api.post<Customer>(`/customers/${id}/restore`);
    },

    /**
     * Activate a customer
     */
    activateCustomer: async (id: string): Promise<Customer> => {
        return api.post<Customer>(`/customers/${id}/activate`);
    },

    /**
     * Deactivate a customer
     */
    deactivateCustomer: async (id: string): Promise<Customer> => {
        return api.post<Customer>(`/customers/${id}/deactivate`);
    },

    /**
     * Block a customer
     */
    blockCustomer: async (id: string): Promise<Customer> => {
        return api.post<Customer>(`/customers/${id}/block`);
    },

    /**
     * Unblock a customer
     */
    unblockCustomer: async (id: string): Promise<Customer> => {
        return api.post<Customer>(`/customers/${id}/unblock`);
    },

    // ============================================
    // Contacts
    // ============================================

    /**
     * Get contacts for a customer
     */
    getContacts: async (customerId: string): Promise<ContactListResponse> => {
        return api.get<ContactListResponse>(`/customers/${customerId}/contacts`);
    },

    /**
     * Get a single contact
     */
    getContact: async (customerId: string, contactId: string): Promise<Contact> => {
        return api.get<Contact>(`/customers/${customerId}/contacts/${contactId}`);
    },

    /**
     * Add a contact to a customer
     */
    addContact: async (customerId: string, data: CreateContactRequest): Promise<Contact> => {
        return api.post<Contact>(`/customers/${customerId}/contacts`, data);
    },

    /**
     * Update a contact
     */
    updateContact: async (customerId: string, contactId: string, data: UpdateContactRequest): Promise<Contact> => {
        return api.put<Contact>(`/customers/${customerId}/contacts/${contactId}`, data);
    },

    /**
     * Delete a contact
     */
    deleteContact: async (customerId: string, contactId: string): Promise<void> => {
        return api.delete(`/customers/${customerId}/contacts/${contactId}`);
    },

    /**
     * Set contact as primary
     */
    setPrimaryContact: async (customerId: string, contactId: string): Promise<Contact> => {
        return api.post<Contact>(`/customers/${customerId}/contacts/${contactId}/primary`);
    },

    // ============================================
    // Notes
    // ============================================

    /**
     * Get notes for a customer
     */
    getNotes: async (customerId: string): Promise<NoteListResponse> => {
        return api.get<NoteListResponse>(`/customers/${customerId}/notes`);
    },

    /**
     * Add a note to a customer
     */
    addNote: async (customerId: string, data: CreateNoteRequest): Promise<Note> => {
        return api.post<Note>(`/customers/${customerId}/notes`, data);
    },

    /**
     * Delete a note
     */
    deleteNote: async (customerId: string, noteId: string): Promise<void> => {
        return api.delete(`/customers/${customerId}/notes/${noteId}`);
    },

    /**
     * Pin/unpin a note
     */
    togglePinNote: async (customerId: string, noteId: string): Promise<Note> => {
        return api.post<Note>(`/customers/${customerId}/notes/${noteId}/pin`);
    },

    // ============================================
    // Activities
    // ============================================

    /**
     * Get activities for a customer
     */
    getActivities: async (customerId: string, filters?: ActivityFilters): Promise<ActivityListResponse> => {
        const params: Record<string, string | number | boolean | undefined> = {
            page: filters?.page,
            per_page: filters?.per_page,
            type: filters?.type,
            status: filters?.status,
        };

        return api.get<ActivityListResponse>(`/customers/${customerId}/activities`, params);
    },

    /**
     * Log an activity for a customer
     */
    logActivity: async (customerId: string, data: Partial<Activity>): Promise<Activity> => {
        return api.post<Activity>(`/customers/${customerId}/activities`, data);
    },

    // ============================================
    // Import/Export
    // ============================================

    /**
     * Import customers from file
     */
    importCustomers: async (file: File, mappings: Record<string, string>): Promise<{ import_id: string }> => {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('mappings', JSON.stringify(mappings));

        const response = await fetch('/api/v1/customers/import', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('crm_access_token')}`,
            },
            body: formData,
        });

        if (!response.ok) {
            throw new Error('Failed to import customers');
        }

        const data = await response.json();
        return data.data;
    },

    /**
     * Export customers
     */
    exportCustomers: async (filters?: CustomerFilters, format: 'csv' | 'xlsx' = 'csv'): Promise<Blob> => {
        const params = new URLSearchParams();
        params.append('format', format);

        if (filters?.type) params.append('type', filters.type);
        if (filters?.status) params.append('status', filters.status);
        if (filters?.search) params.append('search', filters.search);

        const response = await fetch(`/api/v1/customers/export?${params.toString()}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('crm_access_token')}`,
            },
        });

        if (!response.ok) {
            throw new Error('Failed to export customers');
        }

        return response.blob();
    },
};

export default customerService;
