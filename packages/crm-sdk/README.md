# @kilangbatik/crm-sdk

Official TypeScript/JavaScript SDK for the CRM Kilang Desa Murni Batik API.

## Installation

```bash
npm install @kilangbatik/crm-sdk
# or
yarn add @kilangbatik/crm-sdk
# or
pnpm add @kilangbatik/crm-sdk
```

## Quick Start

```typescript
import { createClient } from '@kilangbatik/crm-sdk';

// Create a client instance
const crm = createClient({
    baseUrl: 'https://api.example.com',
});

// Login
const { user, accessToken } = await crm.auth.login({
    email: 'user@example.com',
    password: 'password123',
});

// Fetch leads
const leads = await crm.leads.list({ status: 'new', perPage: 10 });
console.log(leads.data);

// Create a customer
const customer = await crm.customers.create({
    name: 'Acme Corp',
    email: 'contact@acme.com',
});
```

## Configuration

```typescript
import { createClient, CRMClient } from '@kilangbatik/crm-sdk';

const crm = createClient({
    // Required: Base URL of the CRM API
    baseUrl: 'https://api.example.com',

    // Optional: API version (default: 'v1')
    apiVersion: 'v1',

    // Optional: Request timeout in ms (default: 30000)
    timeout: 30000,

    // Optional: Custom headers for every request
    headers: {
        'X-Custom-Header': 'value',
    },

    // Optional: Callback when tokens are refreshed
    onTokenRefresh: (tokens) => {
        // Store new tokens in your storage
        localStorage.setItem('accessToken', tokens.accessToken);
        localStorage.setItem('refreshToken', tokens.refreshToken);
    },

    // Optional: Callback when authentication fails
    onAuthError: (error) => {
        // Redirect to login page
        window.location.href = '/login';
    },
});
```

## Authentication

### Login

```typescript
const response = await crm.auth.login({
    email: 'user@example.com',
    password: 'password123',
});

// Response includes user data and tokens
console.log(response.user);
console.log(response.accessToken);
console.log(response.refreshToken);
```

### Register

```typescript
await crm.auth.register({
    firstName: 'John',
    lastName: 'Doe',
    email: 'john@example.com',
    password: 'securePassword123',
});
```

### Set Tokens Manually

If you have stored tokens, you can set them directly:

```typescript
crm.setTokens('access-token', 'refresh-token');
```

### Check Authentication Status

```typescript
if (crm.isAuthenticated()) {
    // User is logged in
}
```

### Logout

```typescript
await crm.auth.logout();
```

### Password Management

```typescript
// Forgot password
await crm.auth.forgotPassword({ email: 'user@example.com' });

// Reset password
await crm.auth.resetPassword({
    token: 'reset-token',
    password: 'newPassword123',
});

// Change password
await crm.auth.changePassword({
    currentPassword: 'oldPassword',
    newPassword: 'newPassword123',
});
```

## Leads

### List Leads

```typescript
const leads = await crm.leads.list({
    page: 1,
    perPage: 20,
    status: 'qualified',
    search: 'john',
    scoreLabel: 'hot',
    sort: 'createdAt',
    order: 'desc',
});

console.log(leads.data);     // Lead[]
console.log(leads.meta);     // { total, page, perPage, totalPages }
```

### Get Lead

```typescript
const lead = await crm.leads.get('lead-id');
```

### Create Lead

```typescript
const lead = await crm.leads.create({
    firstName: 'Jane',
    lastName: 'Smith',
    email: 'jane@company.com',
    phone: '+60123456789',
    company: 'Tech Corp',
    source: 'website',
});
```

### Update Lead

```typescript
const lead = await crm.leads.update('lead-id', {
    status: 'contacted',
    notes: 'Followed up via email',
});
```

### Delete Lead

```typescript
await crm.leads.delete('lead-id');
```

### Qualify/Disqualify Lead

```typescript
// Qualify
await crm.leads.qualify('lead-id', {
    score: 85,
    notes: 'High potential customer',
});

// Disqualify
await crm.leads.disqualify('lead-id', {
    reason: 'Budget constraints',
    notes: 'Not a fit at this time',
});
```

### Convert Lead

```typescript
const result = await crm.leads.convert('lead-id', {
    createCustomer: true,
    createOpportunity: true,
    opportunityName: 'New Deal',
    opportunityValue: 50000,
});

console.log(result.customerId);
console.log(result.opportunityId);
```

## Customers

### List Customers

```typescript
const customers = await crm.customers.list({
    status: 'active',
    segment: 'enterprise',
    search: 'acme',
});
```

### Get Customer

```typescript
const customer = await crm.customers.get('customer-id');
```

### Create Customer

```typescript
const customer = await crm.customers.create({
    name: 'Acme Corporation',
    email: 'sales@acme.com',
    phone: '+60123456789',
    industry: 'Technology',
    website: 'https://acme.com',
});
```

### Update Customer

```typescript
const customer = await crm.customers.update('customer-id', {
    status: 'active',
    segment: 'enterprise',
});
```

### Delete Customer

```typescript
await crm.customers.delete('customer-id');
```

### Customer Contacts

```typescript
// List contacts
const contacts = await crm.customers.listContacts('customer-id');

// Add contact
const contact = await crm.customers.addContact('customer-id', {
    name: 'John Smith',
    email: 'john@acme.com',
    title: 'CTO',
    isPrimary: true,
});

// Update contact
await crm.customers.updateContact('customer-id', 'contact-id', {
    phone: '+60987654321',
});

// Delete contact
await crm.customers.deleteContact('customer-id', 'contact-id');
```

### Customer Notes

```typescript
// List notes
const notes = await crm.customers.listNotes('customer-id');

// Add note
const note = await crm.customers.addNote('customer-id', {
    content: 'Had a productive meeting about expansion plans.',
});

// Delete note
await crm.customers.deleteNote('customer-id', 'note-id');
```

### Customer Activities

```typescript
const activities = await crm.customers.getActivities('customer-id');
```

## Opportunities

### List Opportunities

```typescript
const opportunities = await crm.opportunities.list({
    status: 'open',
    stageId: 'stage-id',
    valueMin: 10000,
    valueMax: 100000,
});
```

### Get Opportunity

```typescript
const opportunity = await crm.opportunities.get('opportunity-id');
```

### Create Opportunity

```typescript
const opportunity = await crm.opportunities.create({
    name: 'Enterprise Deal',
    customerId: 'customer-id',
    value: 75000,
    stageId: 'stage-id',
    probability: 60,
    expectedCloseDate: '2024-06-30',
});
```

### Update Opportunity

```typescript
const opportunity = await crm.opportunities.update('opportunity-id', {
    value: 85000,
    probability: 75,
});
```

### Delete Opportunity

```typescript
await crm.opportunities.delete('opportunity-id');
```

### Move Stage

```typescript
await crm.opportunities.moveStage('opportunity-id', {
    stageId: 'new-stage-id',
    position: 2,
});
```

### Win/Lose Opportunity

```typescript
// Win
await crm.opportunities.win('opportunity-id', {
    actualCloseDate: '2024-05-15',
    notes: 'Closed after successful demo',
});

// Lose
await crm.opportunities.lose('opportunity-id', {
    reason: 'Competitor selected',
    notes: 'Lost to competitor X due to pricing',
});
```

## Pipelines

### List Pipelines

```typescript
const pipelines = await crm.pipelines.list();
```

### Get Pipeline

```typescript
const pipeline = await crm.pipelines.get('pipeline-id');
```

### Create Pipeline

```typescript
const pipeline = await crm.pipelines.create({
    name: 'Enterprise Sales',
    description: 'Pipeline for enterprise deals',
    isDefault: false,
});
```

### Update Pipeline

```typescript
const pipeline = await crm.pipelines.update('pipeline-id', {
    name: 'Enterprise Sales v2',
});
```

### Delete Pipeline

```typescript
await crm.pipelines.delete('pipeline-id');
```

### Pipeline Stages

```typescript
// Add stage
const stage = await crm.pipelines.addStage('pipeline-id', {
    name: 'Negotiation',
    position: 3,
    color: '#F59E0B',
    probability: 60,
});

// Update stage
await crm.pipelines.updateStage('pipeline-id', 'stage-id', {
    name: 'Contract Review',
});

// Delete stage
await crm.pipelines.deleteStage('pipeline-id', 'stage-id');

// Reorder stages
await crm.pipelines.reorderStages('pipeline-id', [
    'stage-1-id',
    'stage-2-id',
    'stage-3-id',
]);
```

### Deals (Pipeline View)

```typescript
// Get all deals in pipeline
const deals = await crm.pipelines.getDeals('pipeline-id');

// Move deal between stages
await crm.pipelines.moveDeal('pipeline-id', 'deal-id', 'new-stage-id');
```

## Dashboard

### Get Dashboard Data

```typescript
const dashboard = await crm.dashboard.getData();

console.log(dashboard.stats);           // DashboardStats
console.log(dashboard.pipelineOverview); // PipelineOverview
console.log(dashboard.recentActivities); // RecentActivity[]
```

### Get Individual Sections

```typescript
// Get stats only
const stats = await crm.dashboard.getStats();

// Get pipeline overview
const pipeline = await crm.dashboard.getPipelineOverview();

// Get recent activities
const activities = await crm.dashboard.getRecentActivities(10);
```

## Error Handling

The SDK throws `CRMError` for API errors:

```typescript
import { CRMError } from '@kilangbatik/crm-sdk';

try {
    await crm.leads.get('non-existent-id');
} catch (error) {
    if (error instanceof CRMError) {
        console.log(error.message);   // Human-readable message
        console.log(error.code);      // Error code (e.g., 'NOT_FOUND')
        console.log(error.status);    // HTTP status (e.g., 404)
        console.log(error.details);   // Additional details
        console.log(error.requestId); // Request ID for support
    }
}
```

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `UNAUTHORIZED` | 401 | Invalid or expired credentials |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 422 | Invalid request data |
| `TIMEOUT` | 408 | Request timed out |
| `NETWORK_ERROR` | 0 | Network connectivity issue |

## TypeScript Support

The SDK is fully typed. Import types as needed:

```typescript
import type {
    // Configuration
    CRMClientConfig,
    TokenPair,

    // Common
    PaginatedResponse,
    PaginationParams,

    // User & Auth
    User,
    LoginRequest,
    LoginResponse,

    // Leads
    Lead,
    LeadStatus,
    LeadFilters,
    CreateLeadRequest,

    // Customers
    Customer,
    CustomerStatus,
    CustomerContact,

    // Opportunities
    Opportunity,
    OpportunityStatus,

    // Pipelines
    Pipeline,
    PipelineStage,
    Deal,

    // Dashboard
    DashboardData,
    DashboardStats,
} from '@kilangbatik/crm-sdk';
```

## Browser & Node.js Support

The SDK works in both browser and Node.js environments. It uses the native `fetch` API, which is available in:

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Node.js 18+
- Deno
- Bun

For older Node.js versions, you may need a fetch polyfill like `node-fetch`.

## License

MIT
