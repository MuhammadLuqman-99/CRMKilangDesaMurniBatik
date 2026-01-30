# Phase 9: User Interface & Productization 🚀

> The next phase of development to make CRM Kilang Desa Murni Batik production-ready for end users.

---

## Overview

Phase 9 focuses on building the missing user-facing components that transform this backend infrastructure into a complete, marketable SaaS product.

---

## 9.1 Frontend Applications (The Missing Link)

### Admin Dashboard (Internal Tool) ✅

- [x] **Setup admin dashboard project** (Vite + React + TypeScript)
- [x] Tenant management UI
  - [x] List all tenants with search/filter
  - [x] Create/edit/delete tenants
  - [x] View tenant statistics
  - [x] Manage tenant subscription plans
- [x] User management UI
  - [x] View all users across tenants
  - [x] Suspend/activate users
  - [x] Reset user passwords
- [x] System monitoring UI
  - [x] API health dashboard
  - [x] Database connection status
  - [x] Queue depth monitoring

> **Implementation Details**: See `frontend-admin-dashboard/` directory. Built with Vite + React + TypeScript, featuring a professional dark theme, JWT authentication, and complete CRUD operations for tenants and users.

### CRM Web Client (Main Application)

- [ ] **Initialize frontend project** (React / Vue / Next.js)
- [ ] **Login & Authentication Screens**
  - [ ] Login page with email/password
  - [ ] OAuth2 login buttons (Google, Microsoft, GitHub)
  - [ ] Forgot password flow
  - [ ] Registration page
  - [ ] Email verification screen
- [ ] **Dashboard**
  - [ ] Sales pipeline overview
  - [ ] Recent activities
  - [ ] Key metrics widgets
  - [ ] Quick actions
- [ ] **Pipeline/Kanban Board**
  - [ ] Drag & drop deals between stages
  - [ ] Filter by pipeline
  - [ ] Deal quick view popup
  - [ ] Create deal from card
- [ ] **Lead Management**
  - [ ] Lead list with filters
  - [ ] Lead detail view
  - [ ] Lead scoring display
  - [ ] Convert lead to opportunity
  - [ ] Bulk lead import
- [ ] **Opportunity Management**
  - [ ] Opportunity list with status filters
  - [ ] Opportunity detail with timeline
  - [ ] Add products/contacts
  - [ ] Win/lose opportunity actions
- [ ] **Customer List & Detail View**
  - [ ] Customer search and filter
  - [ ] Customer 360° detail page
  - [ ] Contact management
  - [ ] Activity history
  - [ ] Notes and attachments
- [ ] **Settings**
  - [ ] Profile settings
  - [ ] Password change
  - [ ] Notification preferences
  - [ ] Team management (for managers)

### API Client SDK

- [ ] **JavaScript/TypeScript SDK**
  - [ ] npm package setup
  - [ ] Authentication helpers
  - [ ] Typed API methods
  - [ ] Error handling
  - [ ] README & docs
- [ ] **Python SDK** (optional)
  - [ ] PyPI package setup
  - [ ] Typed client using dataclasses
  - [ ] Authentication handling

---

## 9.2 Onboarding & Documentation

### Postman Collection

- [ ] Export complete API collection
- [ ] Organize by service (IAM, Customer, Sales, Notification)
- [ ] Include environment variables template
- [ ] Add request examples with sample data
- [ ] Publish to Postman Public Workspace

### User Manual

- [ ] **Getting Started Guide**
  - [ ] First-time login
  - [ ] Setting up your profile
  - [ ] Understanding the dashboard
- [ ] **Lead Management Guide**
  - [ ] Creating a new lead
  - [ ] Qualifying leads
  - [ ] Converting leads to opportunities
- [ ] **Customer Management Guide**
  - [ ] Adding customers
  - [ ] Managing contacts
  - [ ] Importing bulk data
- [ ] **Sales Pipeline Guide**
  - [ ] Understanding pipeline stages
  - [ ] Moving deals through pipeline
  - [ ] Closing deals
- [ ] **Reporting Guide**
  - [ ] Available reports
  - [ ] Exporting data
- [ ] Format: PDF or Web (GitBook/Docusaurus)

### Landing Page

- [ ] **Marketing landing page**
  - [ ] Hero section with value proposition
  - [ ] Feature highlights
  - [ ] Pricing table
  - [ ] Testimonials section
  - [ ] FAQ section
  - [ ] Sign up / Login buttons
- [ ] SEO optimization
- [ ] Mobile responsive
- [ ] Contact form integration

---

## 9.3 Legal & Compliance (Wajib untuk SaaS)

### Terms of Service (ToS)

- [ ] Draft comprehensive terms of service
  - [ ] Service description
  - [ ] User responsibilities
  - [ ] Payment terms
  - [ ] Data usage terms
  - [ ] Termination clause
  - [ ] Limitation of liability
- [ ] Legal review (recommended)
- [ ] Publish to `/terms` route

### Privacy Policy

- [ ] Draft privacy policy
  - [ ] Data collection practices
  - [ ] How data is used
  - [ ] Data retention policy
  - [ ] User rights (access, deletion, portability)
  - [ ] Cookie usage
  - [ ] Third-party services
- [ ] PDPA (Malaysia) compliance
- [ ] GDPR compliance (if serving EU)
- [ ] Publish to `/privacy` route

### Service Level Agreement (SLA)

- [ ] Define SLA document
  - [ ] Uptime guarantee (e.g., 99.9%)
  - [ ] Maintenance windows
  - [ ] Support response times
  - [ ] Incident notification process
  - [ ] Credits/compensation for downtime
- [ ] Publish to `/sla` route

---

## 9.4 Operational Readiness

### Status Page

- [ ] Setup status page (UptimeRobot / Instatus / Betteruptime)
- [ ] Configure service monitors
  - [ ] API Gateway
  - [ ] IAM Service
  - [ ] Customer Service
  - [ ] Sales Service
  - [ ] Notification Service
  - [ ] Database connectivity
- [ ] Incident reporting workflow
- [ ] Subscriber notifications (email/SMS)
- [ ] Custom domain: `status.your-domain.com`

### Support Channel

- [ ] Setup support email (`support@your-domain.com`)
- [ ] Choose ticketing system
  - [ ] Option: Zendesk
  - [ ] Option: Freshdesk
  - [ ] Option: Help Scout
  - [ ] Option: Crisp (chat + tickets)
- [ ] Define support tiers
  - [ ] Free: Email only, 48h response
  - [ ] Pro: Priority email, 24h response
  - [ ] Enterprise: Dedicated support, 4h response
- [ ] Create FAQ / Knowledge base
- [ ] Setup chatbot / canned responses

### Billing Integration

- [ ] **Choose Payment Gateway**
  - [ ] Option: Stripe (international)
  - [ ] Option: ToyyibPay (Malaysia)
  - [ ] Option: Billplz (Malaysia)
  - [ ] Option: Revenue Monster (Malaysia)
- [ ] **Implement subscription billing**
  - [ ] Plan creation API
  - [ ] Subscription management
  - [ ] Payment method storage
  - [ ] Invoice generation
  - [ ] Payment webhooks
- [ ] **Tenant billing integration**
  - [ ] Link tenant to subscription
  - [ ] Plan upgrade/downgrade
  - [ ] Usage-based billing (optional)
- [ ] **Billing portal**
  - [ ] View invoices
  - [ ] Update payment method
  - [ ] Cancel subscription

---

## Progress Tracker

| Section | Total Tasks | Completed | Progress |
|---------|-------------|-----------|----------|
| 9.1 Frontend Applications | 35 | 14 | ✅ 40% |
| 9.2 Onboarding & Documentation | 20 | 0 | ⏳ 0% |
| 9.3 Legal & Compliance | 15 | 0 | ⏳ 0% |
| 9.4 Operational Readiness | 20 | 0 | ⏳ 0% |
| **Total** | **90** | **14** | **16%** |

---

## Priority Order

1. **High Priority** — Complete before public launch
   - CRM Web Client (Login, Dashboard, Basic CRUD)
   - Terms of Service & Privacy Policy
   - Support Email Setup

2. **Medium Priority** — Complete within 2 weeks of launch
   - Pipeline/Kanban Board
   - User Manual
   - Status Page

3. **Low Priority** — Post-launch enhancements
   - Admin Dashboard
   - JavaScript SDK
   - Billing Integration (if starting with free tier)

---

## Notes

> [!IMPORTANT]
> Phase 9 tasks are essential for user adoption. Without a frontend, users cannot interact with the powerful backend APIs built in Phases 1-8.

> [!TIP]
> Consider using low-code tools like Retool or Refine.dev for the Admin Dashboard to speed up development.

> [!CAUTION]
> Legal documents (ToS, Privacy Policy) should be reviewed by a legal professional before publishing.

---

**Last Updated**: 2026-01-30
