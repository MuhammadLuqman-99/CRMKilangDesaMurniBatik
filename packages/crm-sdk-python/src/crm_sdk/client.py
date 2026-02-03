"""
CRM Platform Python SDK Client.

Main client class for interacting with the CRM Platform API.
"""

from typing import Any, Dict, List, Optional, Type, TypeVar
from datetime import datetime, timedelta
import httpx

from .models import (
    AuthTokens,
    LoginResponse,
    User,
    UserCreate,
    UserUpdate,
    Tenant,
    TenantCreate,
    Customer,
    CustomerCreate,
    CustomerUpdate,
    Contact,
    ContactCreate,
    Lead,
    LeadCreate,
    LeadUpdate,
    Opportunity,
    OpportunityCreate,
    Deal,
    Pipeline,
    PaginatedResponse,
)
from .exceptions import (
    AuthenticationError,
    NetworkError,
    TimeoutError as SDKTimeoutError,
    raise_for_status,
)

T = TypeVar("T")


class CRMClient:
    """
    Main client for CRM Platform API.

    Example:
        >>> client = CRMClient(base_url="https://api.crmplatform.my")
        >>> client.auth.login("user@example.com", "password")
        >>> customers = client.customers.list()
    """

    def __init__(
        self,
        base_url: str = "https://api.crmplatform.my",
        api_key: Optional[str] = None,
        access_token: Optional[str] = None,
        tenant_id: Optional[str] = None,
        timeout: float = 30.0,
        auto_refresh: bool = True,
    ):
        """
        Initialize the CRM client.

        Args:
            base_url: Base URL of the CRM API.
            api_key: API key for authentication (alternative to login).
            access_token: Pre-existing access token.
            tenant_id: Tenant ID for multi-tenant requests.
            timeout: Request timeout in seconds.
            auto_refresh: Automatically refresh expired tokens.
        """
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.tenant_id = tenant_id
        self.timeout = timeout
        self.auto_refresh = auto_refresh

        self._access_token = access_token
        self._refresh_token: Optional[str] = None
        self._token_expires_at: Optional[datetime] = None

        self._client = httpx.Client(
            base_url=self.base_url,
            timeout=timeout,
        )

        # Initialize service clients
        self.auth = AuthService(self)
        self.users = UserService(self)
        self.tenants = TenantService(self)
        self.customers = CustomerService(self)
        self.contacts = ContactService(self)
        self.leads = LeadService(self)
        self.opportunities = OpportunityService(self)
        self.deals = DealService(self)
        self.pipelines = PipelineService(self)

    def close(self) -> None:
        """Close the HTTP client."""
        self._client.close()

    def __enter__(self) -> "CRMClient":
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()

    def _get_headers(self) -> Dict[str, str]:
        """Get headers for API requests."""
        headers: Dict[str, str] = {
            "Content-Type": "application/json",
            "Accept": "application/json",
        }

        if self.api_key:
            headers["X-API-Key"] = self.api_key
        elif self._access_token:
            headers["Authorization"] = f"Bearer {self._access_token}"

        if self.tenant_id:
            headers["X-Tenant-ID"] = self.tenant_id

        return headers

    def _should_refresh_token(self) -> bool:
        """Check if token should be refreshed."""
        if not self.auto_refresh or not self._refresh_token:
            return False
        if not self._token_expires_at:
            return False
        # Refresh 5 minutes before expiry
        return datetime.now() >= self._token_expires_at - timedelta(minutes=5)

    def _refresh_access_token(self) -> None:
        """Refresh the access token."""
        if not self._refresh_token:
            raise AuthenticationError("No refresh token available")

        response = self._client.post(
            "/api/v1/auth/refresh",
            json={"refresh_token": self._refresh_token},
        )

        if response.status_code != 200:
            raise AuthenticationError("Failed to refresh token")

        data = response.json()
        self._access_token = data["access_token"]
        self._token_expires_at = datetime.now() + timedelta(seconds=data.get("expires_in", 3600))

    def request(
        self,
        method: str,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        json: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> httpx.Response:
        """
        Make an HTTP request to the API.

        Args:
            method: HTTP method (GET, POST, PUT, DELETE, etc.).
            path: API endpoint path.
            params: Query parameters.
            json: JSON body data.
            **kwargs: Additional arguments passed to httpx.

        Returns:
            httpx.Response object.

        Raises:
            CRMError: On API errors.
            NetworkError: On network issues.
            TimeoutError: On request timeout.
        """
        if self._should_refresh_token():
            self._refresh_access_token()

        try:
            response = self._client.request(
                method=method,
                url=path,
                params=params,
                json=json,
                headers=self._get_headers(),
                **kwargs,
            )
        except httpx.TimeoutException as e:
            raise SDKTimeoutError(f"Request timed out: {e}", timeout=self.timeout)
        except httpx.RequestError as e:
            raise NetworkError(f"Network error: {e}", original_error=e)

        if response.status_code >= 400:
            try:
                error_data = response.json()
            except Exception:
                error_data = {"error": response.text}
            raise_for_status(response.status_code, error_data)

        return response

    def get(self, path: str, params: Optional[Dict[str, Any]] = None, **kwargs: Any) -> Any:
        """Make a GET request."""
        response = self.request("GET", path, params=params, **kwargs)
        return response.json()

    def post(self, path: str, json: Optional[Dict[str, Any]] = None, **kwargs: Any) -> Any:
        """Make a POST request."""
        response = self.request("POST", path, json=json, **kwargs)
        return response.json()

    def put(self, path: str, json: Optional[Dict[str, Any]] = None, **kwargs: Any) -> Any:
        """Make a PUT request."""
        response = self.request("PUT", path, json=json, **kwargs)
        return response.json()

    def delete(self, path: str, **kwargs: Any) -> None:
        """Make a DELETE request."""
        self.request("DELETE", path, **kwargs)


# ============================================================================
# Service Classes
# ============================================================================

class BaseService:
    """Base class for API services."""

    def __init__(self, client: CRMClient):
        self.client = client


class AuthService(BaseService):
    """Authentication service."""

    def login(self, email: str, password: str, tenant_id: Optional[str] = None) -> LoginResponse:
        """
        Login with email and password.

        Args:
            email: User's email address.
            password: User's password.
            tenant_id: Optional tenant ID.

        Returns:
            LoginResponse with tokens and user info.
        """
        data = self.client.post(
            "/api/v1/auth/login",
            json={
                "email": email,
                "password": password,
                "tenant_id": tenant_id or self.client.tenant_id,
            },
        )

        # Store tokens
        self.client._access_token = data["access_token"]
        self.client._refresh_token = data["refresh_token"]
        self.client._token_expires_at = datetime.now() + timedelta(
            seconds=data.get("expires_in", 3600)
        )

        return LoginResponse(**data)

    def logout(self) -> None:
        """Logout and invalidate tokens."""
        try:
            self.client.post("/api/v1/auth/logout")
        finally:
            self.client._access_token = None
            self.client._refresh_token = None
            self.client._token_expires_at = None

    def register(
        self,
        email: str,
        password: str,
        first_name: Optional[str] = None,
        last_name: Optional[str] = None,
        tenant_id: Optional[str] = None,
    ) -> User:
        """Register a new user."""
        data = self.client.post(
            "/api/v1/auth/register",
            json={
                "email": email,
                "password": password,
                "first_name": first_name,
                "last_name": last_name,
                "tenant_id": tenant_id or self.client.tenant_id,
            },
        )
        return User(**data)

    def me(self) -> User:
        """Get current user info."""
        data = self.client.get("/api/v1/auth/me")
        return User(**data)

    def refresh(self) -> AuthTokens:
        """Refresh access token."""
        self.client._refresh_access_token()
        return AuthTokens(
            access_token=self.client._access_token or "",
            refresh_token=self.client._refresh_token or "",
            token_type="Bearer",
            expires_in=3600,
        )


class UserService(BaseService):
    """User management service."""

    def list(
        self,
        page: int = 1,
        per_page: int = 20,
        status: Optional[str] = None,
    ) -> PaginatedResponse[User]:
        """List users."""
        params = {"page": page, "per_page": per_page}
        if status:
            params["status"] = status

        data = self.client.get("/api/v1/users", params=params)
        return PaginatedResponse[User](
            data=[User(**u) for u in data.get("data", [])],
            total=data.get("total", 0),
            page=page,
            per_page=per_page,
            total_pages=data.get("total_pages", 1),
        )

    def get(self, user_id: str) -> User:
        """Get user by ID."""
        data = self.client.get(f"/api/v1/users/{user_id}")
        return User(**data)

    def create(self, user: UserCreate) -> User:
        """Create a new user."""
        data = self.client.post("/api/v1/users", json=user.model_dump())
        return User(**data)

    def update(self, user_id: str, user: UserUpdate) -> User:
        """Update a user."""
        data = self.client.put(
            f"/api/v1/users/{user_id}",
            json=user.model_dump(exclude_unset=True),
        )
        return User(**data)

    def delete(self, user_id: str) -> None:
        """Delete a user."""
        self.client.delete(f"/api/v1/users/{user_id}")


class TenantService(BaseService):
    """Tenant management service."""

    def list(self, page: int = 1, per_page: int = 20) -> PaginatedResponse[Tenant]:
        """List tenants."""
        data = self.client.get("/api/v1/tenants", params={"page": page, "per_page": per_page})
        return PaginatedResponse[Tenant](
            data=[Tenant(**t) for t in data.get("data", [])],
            total=data.get("total", 0),
            page=page,
            per_page=per_page,
            total_pages=data.get("total_pages", 1),
        )

    def get(self, tenant_id: str) -> Tenant:
        """Get tenant by ID."""
        data = self.client.get(f"/api/v1/tenants/{tenant_id}")
        return Tenant(**data)

    def create(self, tenant: TenantCreate) -> Tenant:
        """Create a new tenant."""
        data = self.client.post("/api/v1/tenants", json=tenant.model_dump())
        return Tenant(**data)


class CustomerService(BaseService):
    """Customer management service."""

    def list(
        self,
        page: int = 1,
        per_page: int = 20,
        status: Optional[str] = None,
        search: Optional[str] = None,
    ) -> PaginatedResponse[Customer]:
        """List customers."""
        params: Dict[str, Any] = {"page": page, "per_page": per_page}
        if status:
            params["status"] = status
        if search:
            params["search"] = search

        data = self.client.get("/api/v1/customers", params=params)
        return PaginatedResponse[Customer](
            data=[Customer(**c) for c in data.get("data", [])],
            total=data.get("total", 0),
            page=page,
            per_page=per_page,
            total_pages=data.get("total_pages", 1),
        )

    def get(self, customer_id: str) -> Customer:
        """Get customer by ID."""
        data = self.client.get(f"/api/v1/customers/{customer_id}")
        return Customer(**data)

    def create(self, customer: CustomerCreate) -> Customer:
        """Create a new customer."""
        data = self.client.post("/api/v1/customers", json=customer.model_dump())
        return Customer(**data)

    def update(self, customer_id: str, customer: CustomerUpdate) -> Customer:
        """Update a customer."""
        data = self.client.put(
            f"/api/v1/customers/{customer_id}",
            json=customer.model_dump(exclude_unset=True),
        )
        return Customer(**data)

    def delete(self, customer_id: str) -> None:
        """Delete a customer."""
        self.client.delete(f"/api/v1/customers/{customer_id}")

    def search(self, query: str, limit: int = 20) -> List[Customer]:
        """Search customers."""
        data = self.client.get("/api/v1/customers/search", params={"q": query, "limit": limit})
        return [Customer(**c) for c in data.get("data", [])]


class ContactService(BaseService):
    """Contact management service."""

    def list(self, customer_id: str) -> List[Contact]:
        """List contacts for a customer."""
        data = self.client.get(f"/api/v1/customers/{customer_id}/contacts")
        return [Contact(**c) for c in data.get("data", [])]

    def get(self, customer_id: str, contact_id: str) -> Contact:
        """Get contact by ID."""
        data = self.client.get(f"/api/v1/customers/{customer_id}/contacts/{contact_id}")
        return Contact(**data)

    def create(self, customer_id: str, contact: ContactCreate) -> Contact:
        """Create a new contact."""
        data = self.client.post(
            f"/api/v1/customers/{customer_id}/contacts",
            json=contact.model_dump(),
        )
        return Contact(**data)

    def delete(self, customer_id: str, contact_id: str) -> None:
        """Delete a contact."""
        self.client.delete(f"/api/v1/customers/{customer_id}/contacts/{contact_id}")


class LeadService(BaseService):
    """Lead management service."""

    def list(
        self,
        page: int = 1,
        per_page: int = 20,
        status: Optional[str] = None,
    ) -> PaginatedResponse[Lead]:
        """List leads."""
        params: Dict[str, Any] = {"page": page, "per_page": per_page}
        if status:
            params["status"] = status

        data = self.client.get("/api/v1/leads", params=params)
        return PaginatedResponse[Lead](
            data=[Lead(**l) for l in data.get("data", [])],
            total=data.get("total", 0),
            page=page,
            per_page=per_page,
            total_pages=data.get("total_pages", 1),
        )

    def get(self, lead_id: str) -> Lead:
        """Get lead by ID."""
        data = self.client.get(f"/api/v1/leads/{lead_id}")
        return Lead(**data)

    def create(self, lead: LeadCreate) -> Lead:
        """Create a new lead."""
        data = self.client.post("/api/v1/leads", json=lead.model_dump())
        return Lead(**data)

    def update(self, lead_id: str, lead: LeadUpdate) -> Lead:
        """Update a lead."""
        data = self.client.put(
            f"/api/v1/leads/{lead_id}",
            json=lead.model_dump(exclude_unset=True),
        )
        return Lead(**data)

    def delete(self, lead_id: str) -> None:
        """Delete a lead."""
        self.client.delete(f"/api/v1/leads/{lead_id}")

    def qualify(self, lead_id: str) -> Lead:
        """Qualify a lead."""
        data = self.client.post(f"/api/v1/leads/{lead_id}/qualify")
        return Lead(**data)

    def convert(self, lead_id: str, opportunity_data: Dict[str, Any]) -> Dict[str, Any]:
        """Convert a lead to an opportunity."""
        return self.client.post(f"/api/v1/leads/{lead_id}/convert", json=opportunity_data)


class OpportunityService(BaseService):
    """Opportunity management service."""

    def list(
        self,
        page: int = 1,
        per_page: int = 20,
        status: Optional[str] = None,
        pipeline_id: Optional[str] = None,
    ) -> PaginatedResponse[Opportunity]:
        """List opportunities."""
        params: Dict[str, Any] = {"page": page, "per_page": per_page}
        if status:
            params["status"] = status
        if pipeline_id:
            params["pipeline_id"] = pipeline_id

        data = self.client.get("/api/v1/opportunities", params=params)
        return PaginatedResponse[Opportunity](
            data=[Opportunity(**o) for o in data.get("data", [])],
            total=data.get("total", 0),
            page=page,
            per_page=per_page,
            total_pages=data.get("total_pages", 1),
        )

    def get(self, opportunity_id: str) -> Opportunity:
        """Get opportunity by ID."""
        data = self.client.get(f"/api/v1/opportunities/{opportunity_id}")
        return Opportunity(**data)

    def create(self, opportunity: OpportunityCreate) -> Opportunity:
        """Create a new opportunity."""
        data = self.client.post("/api/v1/opportunities", json=opportunity.model_dump())
        return Opportunity(**data)

    def win(self, opportunity_id: str, reason: Optional[str] = None) -> Opportunity:
        """Mark opportunity as won."""
        data = self.client.post(
            f"/api/v1/opportunities/{opportunity_id}/win",
            json={"reason": reason} if reason else None,
        )
        return Opportunity(**data)

    def lose(self, opportunity_id: str, reason: str) -> Opportunity:
        """Mark opportunity as lost."""
        data = self.client.post(
            f"/api/v1/opportunities/{opportunity_id}/lose",
            json={"reason": reason},
        )
        return Opportunity(**data)


class DealService(BaseService):
    """Deal management service."""

    def list(
        self,
        page: int = 1,
        per_page: int = 20,
        status: Optional[str] = None,
    ) -> PaginatedResponse[Deal]:
        """List deals."""
        params: Dict[str, Any] = {"page": page, "per_page": per_page}
        if status:
            params["status"] = status

        data = self.client.get("/api/v1/deals", params=params)
        return PaginatedResponse[Deal](
            data=[Deal(**d) for d in data.get("data", [])],
            total=data.get("total", 0),
            page=page,
            per_page=per_page,
            total_pages=data.get("total_pages", 1),
        )

    def get(self, deal_id: str) -> Deal:
        """Get deal by ID."""
        data = self.client.get(f"/api/v1/deals/{deal_id}")
        return Deal(**data)


class PipelineService(BaseService):
    """Pipeline management service."""

    def list(self) -> List[Pipeline]:
        """List pipelines."""
        data = self.client.get("/api/v1/pipelines")
        return [Pipeline(**p) for p in data.get("data", [])]

    def get(self, pipeline_id: str) -> Pipeline:
        """Get pipeline by ID."""
        data = self.client.get(f"/api/v1/pipelines/{pipeline_id}")
        return Pipeline(**data)
