# API Overview

> Comprehensive API documentation for CRM Kilang Desa Murni Batik

---

## Base URL

```
Production: https://api.your-domain.com
Development: http://localhost:8080
```

## Authentication

All API requests (except `/auth/login` and `/auth/register`) require authentication via JWT Bearer token.

### Login Example

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your-password"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600,
    "token_type": "Bearer"
  }
}
```

### Using the Token

```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

---

## IAM Service Endpoints

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/auth/register` | Register new user |
| `POST` | `/auth/login` | User login |
| `POST` | `/auth/refresh` | Refresh access token |
| `POST` | `/auth/logout` | Logout user |
| `GET` | `/auth/me` | Get current user profile |
| `PUT` | `/auth/password` | Change password |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/users` | List all users (paginated) |
| `POST` | `/users` | Create new user |
| `GET` | `/users/{id}` | Get user by ID |
| `PUT` | `/users/{id}` | Update user |
| `DELETE` | `/users/{id}` | Delete user |
| `POST` | `/users/{id}/roles` | Assign role to user |
| `DELETE` | `/users/{id}/roles` | Remove role from user |
| `GET` | `/users/{id}/permissions` | Get user permissions |

### Roles

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/roles` | List all roles |
| `POST` | `/roles` | Create new role |
| `GET` | `/roles/{id}` | Get role by ID |
| `PUT` | `/roles/{id}` | Update role |
| `DELETE` | `/roles/{id}` | Delete role |
| `POST` | `/roles/{id}/permissions` | Add permission to role |
| `DELETE` | `/roles/{id}/permissions` | Remove permission from role |
| `GET` | `/roles/{id}/users` | Get users with role |
| `GET` | `/roles/system` | List system roles |

### Tenants

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/tenants` | List all tenants |
| `POST` | `/tenants` | Create new tenant |
| `GET` | `/tenants/{id}` | Get tenant by ID |
| `PUT` | `/tenants/{id}` | Update tenant |
| `DELETE` | `/tenants/{id}` | Delete tenant |
| `PUT` | `/tenants/{id}/status` | Update tenant status |
| `PUT` | `/tenants/{id}/plan` | Update tenant plan |
| `GET` | `/tenants/{id}/stats` | Get tenant statistics |
| `GET` | `/tenants/check-slug` | Check slug availability |
| `GET` | `/tenants/by-slug/{slug}` | Get tenant by slug |

---

## Customer Service Endpoints

### Customers

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/customers` | List/search customers |
| `POST` | `/customers` | Create customer |
| `GET` | `/customers/{id}` | Get customer by ID |
| `PUT` | `/customers/{id}` | Update customer |
| `DELETE` | `/customers/{id}` | Delete customer |
| `POST` | `/customers/{id}/restore` | Restore deleted customer |
| `POST` | `/customers/{id}/activate` | Activate customer |
| `POST` | `/customers/{id}/deactivate` | Deactivate customer |
| `POST` | `/customers/{id}/block` | Block customer |
| `POST` | `/customers/{id}/unblock` | Unblock customer |

### Contacts

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/customers/{id}/contacts` | List customer contacts |
| `POST` | `/customers/{id}/contacts` | Add contact |
| `GET` | `/customers/{id}/contacts/{contactId}` | Get contact |
| `PUT` | `/customers/{id}/contacts/{contactId}` | Update contact |
| `DELETE` | `/customers/{id}/contacts/{contactId}` | Delete contact |
| `POST` | `/customers/{id}/contacts/{contactId}/primary` | Set as primary |

### Notes & Activities

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/customers/{id}/notes` | List customer notes |
| `POST` | `/customers/{id}/notes` | Add note |
| `DELETE` | `/customers/{id}/notes/{noteId}` | Delete note |
| `POST` | `/customers/{id}/notes/{noteId}/pin` | Pin/unpin note |
| `GET` | `/customers/{id}/activities` | Get activity log |
| `POST` | `/customers/{id}/activities` | Log activity |

### Import/Export

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/customers/import` | Import customers (CSV/XLSX/JSON) |
| `GET` | `/customers/export` | Export customers |
| `GET` | `/imports` | List import jobs |
| `GET` | `/imports/{id}` | Get import status |
| `DELETE` | `/imports/{id}` | Cancel import |

### Segments

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/segments` | List segments |
| `POST` | `/segments` | Create segment |
| `GET` | `/segments/{id}` | Get segment |
| `PUT` | `/segments/{id}` | Update segment |
| `DELETE` | `/segments/{id}` | Delete segment |
| `POST` | `/segments/{id}/refresh` | Refresh dynamic segment |
| `GET` | `/segments/{id}/customers` | Get segment customers |

---

## Sales Service Endpoints

### Leads

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/leads` | List leads |
| `POST` | `/leads` | Create lead |
| `GET` | `/leads/{id}` | Get lead |
| `PUT` | `/leads/{id}` | Update lead |
| `DELETE` | `/leads/{id}` | Delete lead |
| `POST` | `/leads/{id}/convert` | Convert lead to opportunity |
| `POST` | `/leads/{id}/qualify` | Qualify lead |
| `POST` | `/leads/{id}/disqualify` | Disqualify lead |

### Opportunities

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/opportunities` | List opportunities |
| `POST` | `/opportunities` | Create opportunity |
| `GET` | `/opportunities/{id}` | Get opportunity |
| `PUT` | `/opportunities/{id}` | Update opportunity |
| `DELETE` | `/opportunities/{id}` | Delete opportunity |
| `POST` | `/opportunities/{id}/move-stage` | Move to stage |
| `POST` | `/opportunities/{id}/win` | Mark as won |
| `POST` | `/opportunities/{id}/lose` | Mark as lost |
| `POST` | `/opportunities/{id}/reopen` | Reopen opportunity |

### Pipelines

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/pipelines` | List pipelines |
| `POST` | `/pipelines` | Create pipeline |
| `GET` | `/pipelines/{id}` | Get pipeline |
| `PUT` | `/pipelines/{id}` | Update pipeline |
| `DELETE` | `/pipelines/{id}` | Delete pipeline |
| `GET` | `/pipelines/{id}/analytics` | Get pipeline analytics |

### Deals

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/deals` | List deals |
| `GET` | `/deals/{id}` | Get deal |
| `PUT` | `/deals/{id}` | Update deal |
| `POST` | `/deals/{id}/invoice` | Generate invoice |
| `POST` | `/deals/{id}/payment` | Record payment |

---

## Error Handling

All errors follow a consistent format using the `pkg/errors` package:

```json
{
  "success": false,
  "error": {
    "code": "ERR_VALIDATION",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `ERR_VALIDATION` | 400 | Request validation failed |
| `ERR_UNAUTHORIZED` | 401 | Authentication required |
| `ERR_FORBIDDEN` | 403 | Insufficient permissions |
| `ERR_NOT_FOUND` | 404 | Resource not found |
| `ERR_CONFLICT` | 409 | Resource conflict |
| `ERR_RATE_LIMITED` | 429 | Too many requests |
| `ERR_INTERNAL` | 500 | Internal server error |

---

## Pagination

List endpoints support pagination:

```
GET /api/v1/customers?page=1&per_page=20&sort=created_at&order=desc
```

**Response:**
```json
{
  "success": true,
  "data": [...],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

## Rate Limiting

API requests are rate-limited per tenant:

| Tier | Requests/min | Burst |
|------|-------------|-------|
| Free | 60 | 10 |
| Pro | 300 | 50 |
| Enterprise | 1000 | 100 |

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1706630400
```

---

## Swagger/OpenAPI

Interactive API documentation is available at:

```
http://localhost:8080/swagger
```

Features:
- Interactive "Try it out" functionality
- Request/response examples
- Authentication testing
- Schema definitions

---

## Request Headers

| Header | Required | Description |
|--------|----------|-------------|
| `Authorization` | Yes* | Bearer token (*except auth endpoints) |
| `Content-Type` | Yes | `application/json` |
| `X-Tenant-ID` | No | Override tenant (admin only) |
| `X-Request-ID` | No | Trace ID for debugging |
| `Accept-Language` | No | Preferred language (e.g., `ms-MY`) |
