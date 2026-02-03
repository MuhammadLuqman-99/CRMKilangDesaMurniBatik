"""
Data models for CRM Platform SDK.

All models use Pydantic for validation and serialization.
"""

from datetime import datetime
from typing import Any, Generic, List, Optional, TypeVar
from pydantic import BaseModel, Field
from enum import Enum


# ============================================================================
# Enums
# ============================================================================

class UserStatus(str, Enum):
    ACTIVE = "active"
    PENDING_VERIFICATION = "pending_verification"
    SUSPENDED = "suspended"
    INACTIVE = "inactive"


class CustomerType(str, Enum):
    INDIVIDUAL = "individual"
    BUSINESS = "business"


class CustomerStatus(str, Enum):
    ACTIVE = "active"
    INACTIVE = "inactive"
    CHURNED = "churned"


class LeadStatus(str, Enum):
    NEW = "new"
    CONTACTED = "contacted"
    QUALIFIED = "qualified"
    UNQUALIFIED = "unqualified"
    CONVERTED = "converted"


class OpportunityStatus(str, Enum):
    OPEN = "open"
    WON = "won"
    LOST = "lost"


class DealStatus(str, Enum):
    PENDING = "pending"
    ACTIVE = "active"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


# ============================================================================
# Base Models
# ============================================================================

class BaseEntity(BaseModel):
    """Base model for all entities."""
    id: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


# ============================================================================
# Auth Models
# ============================================================================

class AuthTokens(BaseModel):
    """Authentication tokens."""
    access_token: str
    refresh_token: str
    token_type: str = "Bearer"
    expires_in: int


class UserInfo(BaseModel):
    """Basic user info returned with auth."""
    id: str
    email: str
    first_name: Optional[str] = None
    last_name: Optional[str] = None


class LoginResponse(BaseModel):
    """Response from login endpoint."""
    access_token: str
    refresh_token: str
    token_type: str = "Bearer"
    expires_in: int
    user: UserInfo


# ============================================================================
# User Models
# ============================================================================

class User(BaseEntity):
    """User entity."""
    tenant_id: str
    email: str
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    phone: Optional[str] = None
    avatar_url: Optional[str] = None
    status: UserStatus = UserStatus.ACTIVE
    email_verified_at: Optional[datetime] = None
    last_login_at: Optional[datetime] = None
    metadata: dict = Field(default_factory=dict)

    @property
    def full_name(self) -> str:
        """Get user's full name."""
        parts = [self.first_name, self.last_name]
        return " ".join(p for p in parts if p) or self.email


class UserCreate(BaseModel):
    """Data for creating a user."""
    email: str
    password: str
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    phone: Optional[str] = None


class UserUpdate(BaseModel):
    """Data for updating a user."""
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    phone: Optional[str] = None
    avatar_url: Optional[str] = None


# ============================================================================
# Tenant Models
# ============================================================================

class Tenant(BaseEntity):
    """Tenant entity."""
    name: str
    slug: str
    status: str = "active"
    plan: str = "free"
    settings: dict = Field(default_factory=dict)
    metadata: dict = Field(default_factory=dict)
    trial_ends_at: Optional[datetime] = None


class TenantCreate(BaseModel):
    """Data for creating a tenant."""
    name: str
    slug: str
    plan: str = "free"


# ============================================================================
# Customer Models
# ============================================================================

class Email(BaseModel):
    """Email address with metadata."""
    address: str
    type: str = "work"
    is_primary: bool = False


class Phone(BaseModel):
    """Phone number with metadata."""
    number: str
    type: str = "work"
    is_primary: bool = False


class Address(BaseModel):
    """Physical address."""
    street: Optional[str] = None
    city: Optional[str] = None
    state: Optional[str] = None
    postal_code: Optional[str] = None
    country: Optional[str] = None


class Customer(BaseEntity):
    """Customer entity."""
    tenant_id: str
    code: str
    name: str
    type: CustomerType = CustomerType.BUSINESS
    status: CustomerStatus = CustomerStatus.ACTIVE
    email: Optional[Email] = None
    phone: Optional[Phone] = None
    website: Optional[str] = None
    industry: Optional[str] = None
    description: Optional[str] = None
    billing_address: Optional[Address] = None
    shipping_address: Optional[Address] = None
    tags: List[str] = Field(default_factory=list)
    metadata: dict = Field(default_factory=dict)
    version: int = 1


class CustomerCreate(BaseModel):
    """Data for creating a customer."""
    code: str
    name: str
    type: CustomerType = CustomerType.BUSINESS
    email: Optional[Email] = None
    phone: Optional[Phone] = None
    website: Optional[str] = None
    industry: Optional[str] = None
    description: Optional[str] = None


class CustomerUpdate(BaseModel):
    """Data for updating a customer."""
    name: Optional[str] = None
    type: Optional[CustomerType] = None
    status: Optional[CustomerStatus] = None
    email: Optional[Email] = None
    phone: Optional[Phone] = None
    website: Optional[str] = None
    industry: Optional[str] = None
    description: Optional[str] = None
    version: Optional[int] = None  # For optimistic locking


# ============================================================================
# Contact Models
# ============================================================================

class Contact(BaseEntity):
    """Contact entity (person associated with a customer)."""
    customer_id: str
    first_name: str
    last_name: Optional[str] = None
    title: Optional[str] = None
    email: Optional[Email] = None
    phone: Optional[Phone] = None
    is_primary: bool = False
    metadata: dict = Field(default_factory=dict)

    @property
    def full_name(self) -> str:
        """Get contact's full name."""
        parts = [self.first_name, self.last_name]
        return " ".join(p for p in parts if p)


class ContactCreate(BaseModel):
    """Data for creating a contact."""
    first_name: str
    last_name: Optional[str] = None
    title: Optional[str] = None
    email: Optional[Email] = None
    phone: Optional[Phone] = None
    is_primary: bool = False


# ============================================================================
# Lead Models
# ============================================================================

class Lead(BaseEntity):
    """Lead entity."""
    tenant_id: str
    company_name: Optional[str] = None
    contact_name: Optional[str] = None
    contact_email: Optional[str] = None
    contact_phone: Optional[str] = None
    source: Optional[str] = None
    status: LeadStatus = LeadStatus.NEW
    score: int = 0
    website: Optional[str] = None
    industry: Optional[str] = None
    company_size: Optional[str] = None
    budget: Optional[str] = None
    timeline: Optional[str] = None
    notes: Optional[str] = None
    assigned_to: Optional[str] = None
    qualified_at: Optional[datetime] = None
    converted_at: Optional[datetime] = None
    converted_opportunity_id: Optional[str] = None
    metadata: dict = Field(default_factory=dict)


class LeadCreate(BaseModel):
    """Data for creating a lead."""
    company_name: Optional[str] = None
    contact_name: Optional[str] = None
    contact_email: Optional[str] = None
    contact_phone: Optional[str] = None
    source: Optional[str] = None
    score: int = 0
    notes: Optional[str] = None


class LeadUpdate(BaseModel):
    """Data for updating a lead."""
    company_name: Optional[str] = None
    contact_name: Optional[str] = None
    contact_email: Optional[str] = None
    contact_phone: Optional[str] = None
    source: Optional[str] = None
    status: Optional[LeadStatus] = None
    score: Optional[int] = None
    notes: Optional[str] = None
    assigned_to: Optional[str] = None


# ============================================================================
# Pipeline Models
# ============================================================================

class PipelineStage(BaseModel):
    """Pipeline stage."""
    id: str
    pipeline_id: str
    name: str
    type: str = "open"  # open, won, lost
    order: int
    probability: int = 0
    color: Optional[str] = None


class Pipeline(BaseEntity):
    """Sales pipeline."""
    tenant_id: str
    name: str
    description: Optional[str] = None
    is_default: bool = False
    status: str = "active"
    stages: List[PipelineStage] = Field(default_factory=list)


# ============================================================================
# Opportunity Models
# ============================================================================

class Money(BaseModel):
    """Monetary value."""
    amount: int  # In cents
    currency: str = "MYR"

    @property
    def decimal_amount(self) -> float:
        """Get amount as decimal."""
        return self.amount / 100


class Opportunity(BaseEntity):
    """Sales opportunity."""
    tenant_id: str
    customer_id: str
    lead_id: Optional[str] = None
    pipeline_id: str
    stage_id: str
    name: str
    description: Optional[str] = None
    value: Money
    probability: int = 0
    expected_close: Optional[datetime] = None
    actual_close: Optional[datetime] = None
    assigned_to: Optional[str] = None
    status: OpportunityStatus = OpportunityStatus.OPEN
    won_at: Optional[datetime] = None
    lost_at: Optional[datetime] = None
    lost_reason: Optional[str] = None
    metadata: dict = Field(default_factory=dict)


class OpportunityCreate(BaseModel):
    """Data for creating an opportunity."""
    customer_id: str
    pipeline_id: str
    stage_id: str
    name: str
    description: Optional[str] = None
    value_amount: int
    value_currency: str = "MYR"
    probability: int = 0
    expected_close: Optional[datetime] = None


# ============================================================================
# Deal Models
# ============================================================================

class DealLineItem(BaseModel):
    """Line item in a deal."""
    id: str
    product_name: str
    description: Optional[str] = None
    quantity: int = 1
    unit_price: int  # In cents
    discount_percent: int = 0
    tax_percent: int = 0
    total_amount: int


class Deal(BaseEntity):
    """Closed deal."""
    tenant_id: str
    opportunity_id: str
    customer_id: str
    deal_number: Optional[str] = None
    name: str
    description: Optional[str] = None
    value: Money
    status: DealStatus = DealStatus.PENDING
    closed_at: Optional[datetime] = None
    contract_start: Optional[datetime] = None
    contract_end: Optional[datetime] = None
    payment_terms: Optional[str] = None
    line_items: List[DealLineItem] = Field(default_factory=list)
    metadata: dict = Field(default_factory=dict)


# ============================================================================
# Pagination Models
# ============================================================================

T = TypeVar("T")


class PaginatedResponse(BaseModel, Generic[T]):
    """Paginated response wrapper."""
    data: List[T]
    total: int
    page: int = 1
    per_page: int = 20
    total_pages: int = 1

    @property
    def has_next(self) -> bool:
        """Check if there's a next page."""
        return self.page < self.total_pages

    @property
    def has_prev(self) -> bool:
        """Check if there's a previous page."""
        return self.page > 1


# ============================================================================
# Activity Models
# ============================================================================

class Activity(BaseEntity):
    """Activity/timeline entry."""
    tenant_id: str
    entity_type: str  # customer, lead, opportunity, deal
    entity_id: str
    type: str  # call, email, meeting, note, task
    title: str
    description: Optional[str] = None
    performed_by: Optional[str] = None
    performed_at: datetime
    metadata: dict = Field(default_factory=dict)


# ============================================================================
# Note Models
# ============================================================================

class Note(BaseEntity):
    """Note attached to an entity."""
    tenant_id: str
    entity_type: str
    entity_id: str
    content: str
    is_pinned: bool = False
    created_by: str
