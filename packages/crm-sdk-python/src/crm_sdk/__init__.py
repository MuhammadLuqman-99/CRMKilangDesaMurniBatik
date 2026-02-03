"""
CRM Platform Python SDK

Official Python SDK for interacting with the CRM Platform API.

Example:
    >>> from crm_sdk import CRMClient
    >>> client = CRMClient(base_url="https://api.crmplatform.my")
    >>> client.auth.login("user@example.com", "password")
    >>> customers = client.customers.list()
"""

from .client import CRMClient
from .models import (
    User,
    Tenant,
    Customer,
    Contact,
    Lead,
    Opportunity,
    Deal,
    Pipeline,
    PipelineStage,
    PaginatedResponse,
    AuthTokens,
)
from .exceptions import (
    CRMError,
    AuthenticationError,
    NotFoundError,
    ValidationError,
    RateLimitError,
    ServerError,
)

__version__ = "1.0.0"
__all__ = [
    # Client
    "CRMClient",
    # Models
    "User",
    "Tenant",
    "Customer",
    "Contact",
    "Lead",
    "Opportunity",
    "Deal",
    "Pipeline",
    "PipelineStage",
    "PaginatedResponse",
    "AuthTokens",
    # Exceptions
    "CRMError",
    "AuthenticationError",
    "NotFoundError",
    "ValidationError",
    "RateLimitError",
    "ServerError",
]
