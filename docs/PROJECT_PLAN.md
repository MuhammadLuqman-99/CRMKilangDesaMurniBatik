# CRM Kilang Desa Murni Batik - Project Implementation Plan

> **Architecture**: Clean Architecture + Domain-Driven Design (DDD)
> **Type**: Multi-Tenant SaaS CRM
> **Primary Language**: Go (Golang)

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Phase 1: Foundation & Infrastructure](#phase-1-foundation--infrastructure)
3. [Phase 2: IAM Service](#phase-2-iam-service)
4. [Phase 3: Customer/Contact Service](#phase-3-customercontact-service)
5. [Phase 4: Sales Pipeline Service](#phase-4-sales-pipeline-service)
6. [Phase 5: Notification Service](#phase-5-notification-service)
7. [Phase 6: Integration & API Gateway](#phase-6-integration--api-gateway)
8. [Phase 7: Testing & Quality Assurance](#phase-7-testing--quality-assurance)
9. [Phase 8: Deployment & DevOps](#phase-8-deployment--devops)
10. [Technical Decisions](#technical-decisions)

---

## Project Overview

### Business Context
CRM system for Kilang Desa Murni Batik to manage:
- Customer relationships and contacts
- Sales pipeline (leads → opportunities → deals)
- Multi-tenant support for potential B2B expansion
- Automated notifications and communications

### Bounded Contexts (Microservices)
```
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway                               │
│              (Authentication, Rate Limiting, Routing)            │
└─────────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐    ┌───────────────────┐    ┌─────────────────┐
│  IAM Service  │    │ Customer/Contact  │    │ Sales Pipeline  │
│               │    │     Service       │    │    Service      │
│ - Auth/OAuth2 │    │ - Contact CRUD    │    │ - Leads         │
│ - RBAC/ABAC   │    │ - Customer Mgmt   │    │ - Opportunities │
│ - JWT Tokens  │    │ - Tenant Isolation│    │ - Deals         │
└───────────────┘    └───────────────────┘    └─────────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                │
                    ┌───────────────────┐
                    │   Event Bus       │
                    │ (RabbitMQ/Kafka)  │
                    └───────────────────┘
                                │
                    ┌───────────────────┐
                    │ Notification Svc  │
                    │ - Email/SMS       │
                    │ - Domain Events   │
                    └───────────────────┘
```

---

## Phase 1: Foundation & Infrastructure

### 1.1 Project Structure Setup
- [ ] Initialize Go workspace with modules
- [ ] Create monorepo structure for microservices
- [ ] Setup shared libraries (pkg/)
- [ ] Configure .gitignore and .editorconfig
- [ ] Setup Makefile for common tasks

**Directory Structure:**
```
CRMKilangDesaMurniBatik/
├── cmd/                          # Application entrypoints
│   ├── iam-service/
│   ├── customer-service/
│   ├── sales-service/
│   ├── notification-service/
│   └── api-gateway/
├── internal/                     # Private application code
│   ├── iam/
│   │   ├── domain/              # Entities, Value Objects, Aggregates
│   │   ├── application/         # Use Cases, DTOs
│   │   ├── infrastructure/      # Repositories, External Services
│   │   └── interfaces/          # HTTP Handlers, gRPC
│   ├── customer/
│   ├── sales/
│   └── notification/
├── pkg/                          # Shared libraries
│   ├── auth/                    # JWT utilities
│   ├── database/                # DB connections
│   ├── events/                  # Event bus abstractions
│   ├── middleware/              # Common middleware
│   └── errors/                  # Custom error types
├── api/                          # API specifications
│   ├── proto/                   # gRPC definitions
│   └── openapi/                 # OpenAPI/Swagger specs
├── deployments/                  # Docker, K8s configs
├── scripts/                      # Utility scripts
├── configs/                      # Configuration files
├── migrations/                   # Database migrations
└── docs/                         # Documentation
```

### 1.2 Development Environment
- [ ] Setup Docker Compose for local development
- [ ] Configure PostgreSQL container (IAM, Sales)
- [ ] Configure MongoDB container (Customer/Contact)
- [ ] Configure RabbitMQ/Kafka container
- [ ] Configure Redis for caching/sessions
- [ ] Setup hot-reload for development (Air)

### 1.3 Shared Libraries (pkg/)
- [ ] Implement custom error handling package
- [ ] Create database connection utilities
- [ ] Build JWT token utilities
- [ ] Implement tenant context middleware
- [ ] Create event bus abstraction layer
- [ ] Build logging infrastructure (structured logs)
- [ ] Implement distributed tracing (OpenTelemetry)

---

## Phase 2: IAM Service

### 2.1 Domain Layer
- [ ] Define `User` entity with behaviors
- [ ] Define `Role` entity
- [ ] Define `Permission` value object
- [ ] Define `Tenant` aggregate root
- [ ] Define `RefreshToken` entity
- [ ] Implement domain events (UserCreated, UserRoleChanged)

**Domain Models:**
```go
// User Aggregate
type User struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    Email       Email           // Value Object
    Password    HashedPassword  // Value Object
    Roles       []Role
    Status      UserStatus
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (u *User) AssignRole(role Role) error { ... }
func (u *User) Deactivate() error { ... }
```

### 2.2 Application Layer (Use Cases)
- [ ] RegisterUser use case
- [ ] AuthenticateUser use case
- [ ] RefreshToken use case
- [ ] AssignRole use case
- [ ] RevokeAccess use case
- [ ] CreateTenant use case
- [ ] ValidatePermission use case (ABAC)

### 2.3 Infrastructure Layer
- [ ] Implement PostgreSQL UserRepository
- [ ] Implement PostgreSQL RoleRepository
- [ ] Implement Redis TokenStore
- [ ] Implement password hashing (bcrypt/argon2)
- [ ] Implement JWT token generation/validation
- [ ] Setup database migrations

### 2.4 Interface Layer (API)
- [ ] POST /api/v1/auth/register
- [ ] POST /api/v1/auth/login
- [ ] POST /api/v1/auth/refresh
- [ ] POST /api/v1/auth/logout
- [ ] GET /api/v1/users
- [ ] GET /api/v1/users/{id}
- [ ] PUT /api/v1/users/{id}
- [ ] DELETE /api/v1/users/{id}
- [ ] GET /api/v1/roles
- [ ] POST /api/v1/roles
- [ ] POST /api/v1/users/{id}/roles

### 2.5 Security Implementation
- [ ] Implement OAuth2/OIDC provider integration
- [ ] Setup RBAC middleware
- [ ] Implement ABAC policy engine
- [ ] Add rate limiting per tenant
- [ ] Implement audit logging

---

## Phase 3: Customer/Contact Service

### 3.1 Domain Layer
- [ ] Define `Customer` aggregate root
- [ ] Define `Contact` entity
- [ ] Define `Address` value object
- [ ] Define `PhoneNumber` value object
- [ ] Define `Email` value object
- [ ] Define `SocialProfile` value object
- [ ] Implement domain events (CustomerCreated, ContactAdded)
- [ ] Enforce invariants (unique email per tenant)

**Domain Models:**
```go
// Customer Aggregate
type Customer struct {
    ID            uuid.UUID
    TenantID      uuid.UUID
    Name          string
    Type          CustomerType  // Individual, Business
    Contacts      []Contact
    Addresses     []Address
    Tags          []string
    CustomFields  map[string]interface{}
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

func (c *Customer) AddContact(contact Contact) error { ... }
func (c *Customer) SetPrimaryContact(contactID uuid.UUID) error { ... }
```

### 3.2 Application Layer (Use Cases)
- [ ] CreateCustomer use case
- [ ] UpdateCustomer use case
- [ ] DeleteCustomer use case
- [ ] AddContact use case
- [ ] UpdateContact use case
- [ ] RemoveContact use case
- [ ] SearchCustomers use case
- [ ] ImportCustomers use case (bulk)
- [ ] ExportCustomers use case

### 3.3 Infrastructure Layer
- [ ] Implement MongoDB CustomerRepository
- [ ] Implement MongoDB ContactRepository
- [ ] Setup MongoDB indexes for search
- [ ] Implement full-text search
- [ ] Setup database migrations/seeding
- [ ] Implement event publisher

### 3.4 Interface Layer (API)
- [ ] GET /api/v1/customers
- [ ] POST /api/v1/customers
- [ ] GET /api/v1/customers/{id}
- [ ] PUT /api/v1/customers/{id}
- [ ] DELETE /api/v1/customers/{id}
- [ ] GET /api/v1/customers/{id}/contacts
- [ ] POST /api/v1/customers/{id}/contacts
- [ ] PUT /api/v1/customers/{id}/contacts/{contactId}
- [ ] DELETE /api/v1/customers/{id}/contacts/{contactId}
- [ ] POST /api/v1/customers/import
- [ ] GET /api/v1/customers/export
- [ ] GET /api/v1/customers/search

---

## Phase 4: Sales Pipeline Service

### 4.1 Domain Layer
- [ ] Define `Lead` aggregate root
- [ ] Define `Opportunity` aggregate root
- [ ] Define `Deal` aggregate root
- [ ] Define `Stage` entity
- [ ] Define `Pipeline` entity
- [ ] Define `Money` value object
- [ ] Define `Probability` value object
- [ ] Implement `Lead.ConvertToDeal()` behavior
- [ ] Implement domain events (LeadCreated, LeadConverted, DealWon, DealLost)

**Domain Models:**
```go
// Lead Aggregate
type Lead struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    CustomerID   uuid.UUID
    Source       LeadSource
    Status       LeadStatus
    Score        int
    AssignedTo   uuid.UUID
    CreatedAt    time.Time
}

func (l *Lead) ConvertToOpportunity(value Money) (*Opportunity, error) { ... }
func (l *Lead) Qualify(score int) error { ... }

// Opportunity Aggregate
type Opportunity struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    CustomerID      uuid.UUID
    LeadID          *uuid.UUID
    PipelineID      uuid.UUID
    StageID         uuid.UUID
    Value           Money
    Probability     Probability
    ExpectedClose   time.Time
    AssignedTo      uuid.UUID
}

func (o *Opportunity) MoveToStage(stageID uuid.UUID) error { ... }
func (o *Opportunity) Win() (*Deal, error) { ... }
func (o *Opportunity) Lose(reason string) error { ... }
```

### 4.2 Application Layer (Use Cases)
- [ ] CreateLead use case
- [ ] QualifyLead use case
- [ ] ConvertLeadToOpportunity use case
- [ ] CreateOpportunity use case
- [ ] UpdateOpportunityStage use case
- [ ] WinOpportunity use case
- [ ] LoseOpportunity use case
- [ ] CreatePipeline use case
- [ ] GetPipelineAnalytics use case
- [ ] ForecastRevenue use case

### 4.3 Infrastructure Layer
- [ ] Implement PostgreSQL LeadRepository
- [ ] Implement PostgreSQL OpportunityRepository
- [ ] Implement PostgreSQL DealRepository
- [ ] Implement PostgreSQL PipelineRepository
- [ ] Setup database migrations
- [ ] Implement Transactional Outbox pattern
- [ ] Implement event publisher

### 4.4 Interface Layer (API)
- [ ] GET /api/v1/leads
- [ ] POST /api/v1/leads
- [ ] GET /api/v1/leads/{id}
- [ ] PUT /api/v1/leads/{id}
- [ ] POST /api/v1/leads/{id}/convert
- [ ] GET /api/v1/opportunities
- [ ] POST /api/v1/opportunities
- [ ] GET /api/v1/opportunities/{id}
- [ ] PUT /api/v1/opportunities/{id}
- [ ] POST /api/v1/opportunities/{id}/move-stage
- [ ] POST /api/v1/opportunities/{id}/win
- [ ] POST /api/v1/opportunities/{id}/lose
- [ ] GET /api/v1/pipelines
- [ ] POST /api/v1/pipelines
- [ ] GET /api/v1/pipelines/{id}/analytics
- [ ] GET /api/v1/deals
- [ ] GET /api/v1/deals/{id}

### 4.5 Saga Implementation
- [ ] Implement LeadConversion Saga
- [ ] Implement compensating transactions
- [ ] Setup saga orchestrator
- [ ] Implement idempotency checks

---

## Phase 5: Notification Service

### 5.1 Domain Layer
- [ ] Define `Notification` entity
- [ ] Define `NotificationTemplate` entity
- [ ] Define `NotificationChannel` value object
- [ ] Define `NotificationStatus` value object
- [ ] Define event handlers for domain events

### 5.2 Application Layer (Use Cases)
- [ ] SendEmail use case
- [ ] SendSMS use case
- [ ] SendInAppNotification use case
- [ ] CreateTemplate use case
- [ ] ProcessDomainEvent use case
- [ ] RetryFailedNotification use case

### 5.3 Infrastructure Layer
- [ ] Implement Email provider adapter (SendGrid/SES)
- [ ] Implement SMS provider adapter (Twilio)
- [ ] Implement PostgreSQL NotificationRepository
- [ ] Implement event bus consumer (RabbitMQ/Kafka)
- [ ] Implement retry mechanism with backoff
- [ ] Implement Circuit Breaker pattern

### 5.4 Interface Layer
- [ ] Event consumers for:
  - [ ] UserCreated → Send welcome email
  - [ ] LeadCreated → Notify sales team
  - [ ] DealWon → Send confirmation to customer
  - [ ] DealLost → Send follow-up survey
- [ ] GET /api/v1/notifications
- [ ] GET /api/v1/notifications/{id}
- [ ] POST /api/v1/notifications/templates
- [ ] GET /api/v1/notifications/templates

---

## Phase 6: Integration & API Gateway

### 6.1 API Gateway Setup
- [ ] Setup Kong/Traefik/Custom gateway
- [ ] Configure routing rules
- [ ] Implement request aggregation
- [ ] Setup SSL/TLS termination
- [ ] Configure CORS policies

### 6.2 Cross-Cutting Concerns
- [ ] Implement centralized authentication
- [ ] Implement rate limiting per tenant
- [ ] Implement request/response logging
- [ ] Setup distributed tracing propagation
- [ ] Implement request validation

### 6.3 Event Bus Integration
- [ ] Setup RabbitMQ/Kafka cluster
- [ ] Define event schemas
- [ ] Implement dead letter queues
- [ ] Setup event replay capability
- [ ] Implement event versioning

### 6.4 Service Discovery
- [ ] Setup Consul/etcd for service registry
- [ ] Implement health check endpoints
- [ ] Configure load balancing
- [ ] Setup circuit breakers between services

---

## Phase 7: Testing & Quality Assurance

### 7.1 Unit Testing
- [ ] Domain layer unit tests (all services)
- [ ] Use case unit tests (all services)
- [ ] Repository unit tests with mocks
- [ ] Achieve >80% code coverage

### 7.2 Integration Testing
- [ ] Database integration tests
- [ ] Event bus integration tests
- [ ] External service integration tests
- [ ] API endpoint integration tests

### 7.3 End-to-End Testing
- [ ] Complete user registration flow
- [ ] Customer creation and management flow
- [ ] Lead to deal conversion flow
- [ ] Multi-tenant isolation verification

### 7.4 Performance Testing
- [ ] Load testing with k6/Locust
- [ ] Stress testing endpoints
- [ ] Database query optimization
- [ ] Memory leak detection

### 7.5 Security Testing
- [ ] OWASP vulnerability scanning
- [ ] Penetration testing
- [ ] SQL/NoSQL injection testing
- [ ] Authentication bypass testing
- [ ] Cross-tenant access testing

---

## Phase 8: Deployment & DevOps

### 8.1 Containerization
- [ ] Create Dockerfile for each service
- [ ] Optimize Docker images (multi-stage builds)
- [ ] Setup Docker Compose for local dev
- [ ] Configure container health checks

### 8.2 CI/CD Pipeline
- [ ] Setup GitHub Actions/GitLab CI
- [ ] Configure automated testing
- [ ] Setup code quality gates (SonarQube)
- [ ] Configure automated deployments
- [ ] Setup semantic versioning

### 8.3 Kubernetes Deployment
- [ ] Create Kubernetes manifests
- [ ] Setup Helm charts
- [ ] Configure horizontal pod autoscaling
- [ ] Setup ingress controllers
- [ ] Configure secrets management
- [ ] Setup persistent volume claims

### 8.4 Monitoring & Observability
- [ ] Setup Prometheus metrics
- [ ] Configure Grafana dashboards
- [ ] Implement structured logging (ELK/Loki)
- [ ] Setup distributed tracing (Jaeger)
- [ ] Configure alerting rules
- [ ] Setup uptime monitoring

### 8.5 Database Operations
- [ ] Setup automated backups
- [ ] Configure point-in-time recovery
- [ ] Implement blue-green deployments
- [ ] Setup database replication
- [ ] Configure connection pooling

---

## Technical Decisions

### Technology Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Language | Go 1.21+ | Performance, concurrency, strong typing |
| IAM Database | PostgreSQL | ACID compliance, role hierarchies |
| Customer Database | MongoDB | Flexible schema, document structure |
| Sales Database | PostgreSQL | Transactional integrity |
| Message Broker | RabbitMQ | Reliable messaging, DLQ support |
| Cache | Redis | Session storage, caching |
| API Gateway | Kong/Traefik | Industry standard, plugin ecosystem |
| Container Runtime | Docker | Standard containerization |
| Orchestration | Kubernetes | Production-grade orchestration |
| Monitoring | Prometheus + Grafana | Industry standard observability |
| Tracing | OpenTelemetry + Jaeger | Distributed tracing |
| CI/CD | GitHub Actions | Integrated with repository |

### Multi-Tenancy Implementation

```go
// Tenant Context Middleware
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := r.Header.Get("X-Tenant-ID")
        if tenantID == "" {
            // Extract from JWT claims
            claims := r.Context().Value("claims").(jwt.Claims)
            tenantID = claims.TenantID
        }

        ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Row-Level Security (PostgreSQL)
// CREATE POLICY tenant_isolation ON customers
//     USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

### Event-Driven Communication

```go
// Domain Event
type LeadConverted struct {
    EventID      uuid.UUID `json:"event_id"`
    LeadID       uuid.UUID `json:"lead_id"`
    CustomerID   uuid.UUID `json:"customer_id"`
    TenantID     uuid.UUID `json:"tenant_id"`
    OpportunityID uuid.UUID `json:"opportunity_id"`
    ConvertedAt  time.Time `json:"converted_at"`
    ConvertedBy  uuid.UUID `json:"converted_by"`
}

// Event Publisher Interface
type EventPublisher interface {
    Publish(ctx context.Context, event DomainEvent) error
}

// Transactional Outbox
type OutboxEntry struct {
    ID          uuid.UUID
    EventType   string
    Payload     []byte
    Published   bool
    CreatedAt   time.Time
}
```

---

## Milestones Summary

| Phase | Description | Priority |
|-------|-------------|----------|
| Phase 1 | Foundation & Infrastructure | Critical |
| Phase 2 | IAM Service | Critical |
| Phase 3 | Customer/Contact Service | High |
| Phase 4 | Sales Pipeline Service | High |
| Phase 5 | Notification Service | Medium |
| Phase 6 | Integration & API Gateway | High |
| Phase 7 | Testing & QA | Critical |
| Phase 8 | Deployment & DevOps | High |

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Cross-tenant data leakage | Critical | Row-level security, ABAC, extensive testing |
| Service cascade failure | High | Circuit breakers, async communication |
| Database performance | Medium | Polyglot persistence, proper indexing |
| Event ordering issues | Medium | Saga pattern, idempotency keys |
| Authentication bypass | Critical | OAuth2/OIDC standards, security audits |

---

## Quick Start Commands

```bash
# Initialize Go modules
go mod init github.com/kilang-desa-murni/crm

# Start development environment
docker-compose up -d

# Run migrations
make migrate-up

# Run all services
make run-all

# Run tests
make test

# Build all services
make build-all
```

---

**Document Version**: 1.0
**Last Updated**: 2026-01-22
**Author**: System Architect
