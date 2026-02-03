# CRM Platform Python SDK

Official Python SDK for the CRM Platform API.

## Installation

```bash
pip install crm-platform-sdk
```

Or install from source:

```bash
git clone https://github.com/kilang-desa-murni/crm-sdk-python.git
cd crm-sdk-python
pip install -e .
```

## Quick Start

```python
from crm_sdk import CRMClient

# Initialize client
client = CRMClient(
    base_url="https://api.crmplatform.my",
    tenant_id="your-tenant-id"
)

# Login
response = client.auth.login("user@example.com", "password")
print(f"Logged in as: {response.user.email}")

# List customers
customers = client.customers.list(page=1, per_page=20)
for customer in customers.data:
    print(f"- {customer.name} ({customer.code})")

# Create a lead
from crm_sdk import LeadCreate

lead = client.leads.create(LeadCreate(
    company_name="Acme Corp",
    contact_name="John Doe",
    contact_email="john@acme.com",
    source="website",
    score=75
))
print(f"Created lead: {lead.id}")

# Close connection
client.close()
```

## Using Context Manager

```python
from crm_sdk import CRMClient

with CRMClient(base_url="https://api.crmplatform.my") as client:
    client.auth.login("user@example.com", "password")
    customers = client.customers.list()
```

## Authentication

### Email/Password Login

```python
client = CRMClient(base_url="https://api.crmplatform.my")
client.auth.login("user@example.com", "password")
```

### API Key

```python
client = CRMClient(
    base_url="https://api.crmplatform.my",
    api_key="your-api-key"
)
```

### Pre-existing Token

```python
client = CRMClient(
    base_url="https://api.crmplatform.my",
    access_token="your-access-token"
)
```

## Available Services

### Auth Service

```python
# Login
client.auth.login(email, password)

# Register
client.auth.register(email, password, first_name, last_name)

# Get current user
user = client.auth.me()

# Logout
client.auth.logout()

# Refresh token
client.auth.refresh()
```

### Customers Service

```python
# List customers
customers = client.customers.list(page=1, per_page=20, status="active")

# Search customers
results = client.customers.search("acme")

# Get customer
customer = client.customers.get("customer-id")

# Create customer
customer = client.customers.create(CustomerCreate(
    code="CUST-001",
    name="Acme Corp",
    type="business"
))

# Update customer
customer = client.customers.update("customer-id", CustomerUpdate(
    name="Acme Corporation",
    version=1  # For optimistic locking
))

# Delete customer
client.customers.delete("customer-id")
```

### Contacts Service

```python
# List contacts for a customer
contacts = client.contacts.list("customer-id")

# Create contact
contact = client.contacts.create("customer-id", ContactCreate(
    first_name="John",
    last_name="Doe",
    email=Email(address="john@example.com"),
    is_primary=True
))

# Delete contact
client.contacts.delete("customer-id", "contact-id")
```

### Leads Service

```python
# List leads
leads = client.leads.list(status="new")

# Create lead
lead = client.leads.create(LeadCreate(
    company_name="Tech Startup",
    contact_name="Jane Smith",
    contact_email="jane@startup.com",
    source="referral",
    score=80
))

# Qualify lead
lead = client.leads.qualify("lead-id")

# Convert lead to opportunity
result = client.leads.convert("lead-id", {
    "pipeline_id": "pipeline-id",
    "stage_id": "stage-id",
    "name": "New Opportunity",
    "value_amount": 50000
})
```

### Opportunities Service

```python
# List opportunities
opportunities = client.opportunities.list(
    status="open",
    pipeline_id="pipeline-id"
)

# Create opportunity
opportunity = client.opportunities.create(OpportunityCreate(
    customer_id="customer-id",
    pipeline_id="pipeline-id",
    stage_id="stage-id",
    name="Enterprise Deal",
    value_amount=100000,
    value_currency="MYR"
))

# Win opportunity
opportunity = client.opportunities.win("opportunity-id", reason="Best value")

# Lose opportunity
opportunity = client.opportunities.lose("opportunity-id", reason="Budget constraints")
```

### Pipelines Service

```python
# List pipelines
pipelines = client.pipelines.list()

# Get pipeline with stages
pipeline = client.pipelines.get("pipeline-id")
for stage in pipeline.stages:
    print(f"- {stage.name} ({stage.probability}%)")
```

### Deals Service

```python
# List deals
deals = client.deals.list(status="active")

# Get deal
deal = client.deals.get("deal-id")
```

## Error Handling

```python
from crm_sdk import CRMClient
from crm_sdk.exceptions import (
    AuthenticationError,
    NotFoundError,
    ValidationError,
    RateLimitError,
    ServerError,
)

client = CRMClient(base_url="https://api.crmplatform.my")

try:
    client.auth.login("user@example.com", "wrong-password")
except AuthenticationError as e:
    print(f"Login failed: {e.message}")

try:
    customer = client.customers.get("non-existent-id")
except NotFoundError as e:
    print(f"Customer not found: {e.message}")

try:
    client.customers.create(CustomerCreate(code="", name=""))
except ValidationError as e:
    print(f"Validation failed: {e.message}")
    print(f"Errors: {e.details.get('validation_errors')}")

try:
    # Make many requests
    for i in range(1000):
        client.customers.list()
except RateLimitError as e:
    print(f"Rate limited. Retry after: {e.retry_after} seconds")
```

## Pagination

```python
# Using paginated responses
customers = client.customers.list(page=1, per_page=50)

print(f"Total customers: {customers.total}")
print(f"Page {customers.page} of {customers.total_pages}")
print(f"Has next page: {customers.has_next}")
print(f"Has previous page: {customers.has_prev}")

# Iterate through all pages
page = 1
while True:
    customers = client.customers.list(page=page)
    for customer in customers.data:
        print(customer.name)

    if not customers.has_next:
        break
    page += 1
```

## Type Hints

All models and methods are fully typed for excellent IDE support:

```python
from crm_sdk import CRMClient, Customer, Lead

client = CRMClient(base_url="...")
client.auth.login("...", "...")

# IDE will autocomplete customer attributes
customer: Customer = client.customers.get("id")
print(customer.name)  # IDE knows this is str
print(customer.email.address)  # IDE knows email structure

# IDE will show LeadCreate required/optional fields
lead: Lead = client.leads.create(...)
```

## Configuration

```python
client = CRMClient(
    base_url="https://api.crmplatform.my",  # API base URL
    api_key="...",                           # Optional: API key auth
    access_token="...",                      # Optional: Pre-existing token
    tenant_id="...",                         # Optional: Tenant ID for multi-tenant
    timeout=30.0,                            # Request timeout (seconds)
    auto_refresh=True,                       # Auto-refresh expired tokens
)
```

## Requirements

- Python 3.8+
- httpx >= 0.24.0
- pydantic >= 2.0.0

## License

MIT License - see [LICENSE](LICENSE) file.

## Support

- Documentation: https://docs.crmplatform.my/sdk/python
- Issues: https://github.com/kilang-desa-murni/crm-sdk-python/issues
- Email: dev@crmplatform.my
