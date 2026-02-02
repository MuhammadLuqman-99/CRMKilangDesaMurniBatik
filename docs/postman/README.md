# CRM Kilang Desa Murni Batik - Postman Collection

Complete API collection for testing and exploring the CRM API.

## Contents

- `CRM_Kilang_Desa_Murni_Batik.postman_collection.json` - Full API collection
- `CRM_Environment.postman_environment.json` - Environment variables template

## Quick Start

1. **Import Collection**
   - Open Postman
   - Click "Import" button
   - Select `CRM_Kilang_Desa_Murni_Batik.postman_collection.json`

2. **Import Environment**
   - Click "Import" button
   - Select `CRM_Environment.postman_environment.json`

3. **Configure Environment**
   - Select "CRM Kilang Desa Murni Batik" from the environment dropdown
   - Click the "eye" icon to view/edit variables
   - Update `base_url` to your API gateway URL (default: `http://localhost:8080`)
   - Update `user_email` and `user_password` with valid credentials

4. **Authenticate**
   - Navigate to: IAM Service → Auth → Login
   - Click "Send"
   - Access token is automatically saved to environment

5. **Explore the API**
   - All subsequent requests will use the saved access token
   - IDs from created resources are automatically saved for easy chaining

## Collection Structure

```
CRM Kilang Desa Murni Batik
├── IAM Service
│   ├── Auth (Login, Register, Logout, etc.)
│   ├── Users (CRUD, Roles, Permissions)
│   ├── Roles (CRUD, Permissions)
│   ├── Tenants (Multi-tenant management)
│   └── Health (Health checks)
├── Customer Service
│   ├── Customers (CRUD, Status actions)
│   ├── Contacts (Customer contacts)
│   ├── Notes (Customer notes)
│   ├── Activities (Activity tracking)
│   ├── Segments (Customer segmentation)
│   └── Import (Bulk import)
├── Sales Service
│   ├── Leads (CRUD, Qualify, Convert)
│   ├── Opportunities (CRUD, Win/Lose)
│   └── Pipelines (CRUD, Stages)
└── Notification Service
    ├── Notifications (List, Get)
    ├── Templates (CRUD)
    └── Send (Email, SMS)
```

## Environment Variables

| Variable | Description | Auto-Set |
|----------|-------------|----------|
| `base_url` | API Gateway URL | No |
| `user_email` | Login email | No |
| `user_password` | Login password | No |
| `access_token` | JWT access token | Yes (on login) |
| `refresh_token` | JWT refresh token | Yes (on login) |
| `tenant_id` | Current tenant ID | Yes (on create) |
| `customer_id` | Current customer ID | Yes (on create) |
| `lead_id` | Current lead ID | Yes (on create) |
| `opportunity_id` | Current opportunity ID | Yes (on create) |
| `pipeline_id` | Current pipeline ID | Yes (on create) |
| `stage_id` | Current stage ID | Yes (on create) |

## Features

### Automatic Token Management
- Login request automatically saves tokens to environment
- Refresh token request updates both tokens
- All authenticated requests use Bearer token from environment

### Request Chaining
- Creating resources automatically saves their IDs
- Subsequent requests can use `{{resource_id}}` variables
- Makes it easy to test full workflows

### Test Scripts
- Login includes tests to verify successful authentication
- Create requests include tests to save resource IDs

## API Endpoints Summary

| Service | Endpoints |
|---------|-----------|
| IAM | 32 routes |
| Customer | 67 routes |
| Sales | 104 routes |
| Notification | 11 routes |
| **Total** | **214 routes** |

## Common Workflows

### 1. Lead to Customer Conversion
1. Create Lead → `POST /api/v1/sales/leads`
2. Qualify Lead → `POST /api/v1/sales/leads/{id}/qualify`
3. Convert Lead → `POST /api/v1/sales/leads/{id}/convert`

### 2. Sales Pipeline Flow
1. Create Opportunity → `POST /api/v1/sales/opportunities`
2. Move Stage → `POST /api/v1/sales/opportunities/{id}/move-stage`
3. Win/Lose → `POST /api/v1/sales/opportunities/{id}/win`

### 3. Customer Management
1. Create Customer → `POST /api/v1/customers`
2. Add Contacts → `POST /api/v1/customers/{id}/contacts`
3. Log Activities → `POST /api/v1/customers/{id}/activities`

## Troubleshooting

### 401 Unauthorized
- Token may have expired
- Run "Refresh Token" request or login again

### 403 Forbidden
- User lacks required permissions
- Check user's assigned roles

### 404 Not Found
- Resource ID may be invalid
- Verify the ID exists in your environment

## Support

For issues or questions:
- GitHub: https://github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik
- Email: support@kilangbatik.com
