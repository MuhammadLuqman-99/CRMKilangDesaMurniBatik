# Service Level Agreement (SLA)

**Last Updated: February 2026**

**Effective Date: February 3, 2026**

---

## 1. Introduction

This Service Level Agreement ("SLA") is a policy governing the use of the CRM Platform ("Service") under the terms of the Terms of Service between Kilang Desa Murni Batik ("Company", "we", "us") and users of the Service ("Customer", "you").

This SLA applies to all paid subscription plans. Free tier accounts are not covered by this SLA.

---

## 2. Service Commitment

### 2.1 Uptime Guarantee

We commit to maintaining the following service availability levels:

| Plan | Monthly Uptime Target | Annual Uptime Target |
|------|----------------------|---------------------|
| Professional | 99.9% | 99.9% |
| Business | 99.95% | 99.95% |
| Enterprise | 99.99% | 99.99% |

### 2.2 Uptime Calculation

**Monthly Uptime Percentage** is calculated as:

```
Uptime % = ((Total Minutes in Month - Downtime Minutes) / Total Minutes in Month) × 100
```

**Example:**
- Total minutes in a 30-day month: 43,200 minutes
- 99.9% uptime allows: 43.2 minutes of downtime per month
- 99.95% uptime allows: 21.6 minutes of downtime per month
- 99.99% uptime allows: 4.32 minutes of downtime per month

### 2.3 What Counts as Downtime

**Included in Downtime:**
- Service unavailability (HTTP 5xx errors for >1 minute)
- API response time exceeding 30 seconds
- Login/authentication failures
- Data access failures
- Core feature unavailability

**Excluded from Downtime:**
- Scheduled maintenance (with advance notice)
- Force majeure events
- Customer's internet connectivity issues
- Issues caused by customer's equipment or software
- Third-party service outages (payment processors, etc.)
- Beta/preview features
- API rate limiting
- Customer-initiated actions

---

## 3. Scheduled Maintenance

### 3.1 Maintenance Windows

**Regular Maintenance:**
- **Window:** Sundays, 2:00 AM - 6:00 AM MYT (Malaysia Time)
- **Frequency:** Weekly or as needed
- **Notice:** Minimum 48 hours advance notice

**Emergency Maintenance:**
- For critical security patches or urgent fixes
- **Notice:** As much advance notice as reasonably possible
- Will be communicated via status page and email

### 3.2 Maintenance Communication

Maintenance will be announced through:
1. **Status Page:** status.crmplatform.my
2. **Email:** To account administrators
3. **In-App Banner:** 24 hours before scheduled maintenance

### 3.3 Maintenance Commitments

- We will minimize maintenance duration
- We will attempt to maintain read-only access when possible
- Major updates will include rollback plans
- Post-maintenance validation will be performed

---

## 4. Support Response Times

### 4.1 Support Tiers

| Severity | Description | Examples |
|----------|-------------|----------|
| Critical (P1) | Service completely unavailable | Complete outage, data loss, security breach |
| High (P2) | Major feature not working | Cannot create leads, login issues for multiple users |
| Medium (P3) | Feature partially impaired | Slow performance, minor feature bugs |
| Low (P4) | Minor issues/questions | UI issues, how-to questions, feature requests |

### 4.2 Response Time Targets

| Plan | P1 (Critical) | P2 (High) | P3 (Medium) | P4 (Low) |
|------|---------------|-----------|-------------|----------|
| Professional | 4 hours | 8 hours | 24 hours | 48 hours |
| Business | 2 hours | 4 hours | 12 hours | 24 hours |
| Enterprise | 1 hour | 2 hours | 8 hours | 24 hours |

**Response Time** = Time from ticket submission to first meaningful response from support team.

### 4.3 Resolution Time Targets

| Plan | P1 (Critical) | P2 (High) | P3 (Medium) | P4 (Low) |
|------|---------------|-----------|-------------|----------|
| Professional | 8 hours | 24 hours | 5 days | 10 days |
| Business | 4 hours | 12 hours | 3 days | 7 days |
| Enterprise | 2 hours | 8 hours | 2 days | 5 days |

**Resolution Time** = Time to resolve or provide a workaround.

### 4.4 Support Hours

| Plan | Support Hours |
|------|---------------|
| Professional | Business hours (9 AM - 6 PM MYT, Mon-Fri) |
| Business | Extended hours (8 AM - 10 PM MYT, Mon-Sat) |
| Enterprise | 24/7/365 |

### 4.5 Support Channels

| Plan | Channels Available |
|------|-------------------|
| Professional | Email, Help Center |
| Business | Email, Help Center, Live Chat |
| Enterprise | Email, Help Center, Live Chat, Phone, Dedicated Account Manager |

---

## 5. Incident Notification Process

### 5.1 Incident Classification

| Level | Impact | Notification Time |
|-------|--------|-------------------|
| Major | Service-wide outage | Within 15 minutes |
| Significant | Partial outage affecting multiple customers | Within 30 minutes |
| Minor | Limited impact on few customers | Within 1 hour |

### 5.2 Notification Channels

1. **Status Page:** status.crmplatform.my (real-time updates)
2. **Email:** Automatic alerts to account administrators
3. **In-App:** Banner notifications for active users
4. **SMS:** For Enterprise customers (P1 incidents only)

### 5.3 Incident Updates

- **During Incident:** Updates every 30 minutes (or more frequently for P1)
- **Root Cause Analysis:** Provided within 5 business days for P1/P2 incidents
- **Post-Incident Report:** Available upon request for Enterprise customers

### 5.4 Incident Communication Content

Each notification will include:
- Current status and impact
- Affected services/regions
- Estimated time to resolution (when known)
- Workarounds (if available)
- Next update time

---

## 6. Service Credits for Downtime

### 6.1 Credit Eligibility

If we fail to meet our uptime commitment, you may be eligible for service credits.

### 6.2 Credit Calculation

| Monthly Uptime | Service Credit (% of Monthly Fee) |
|----------------|-----------------------------------|
| 99.0% - 99.9% (Professional) | 10% |
| 98.0% - 99.0% | 25% |
| 95.0% - 98.0% | 50% |
| < 95.0% | 100% |

**For Business/Enterprise plans:**

| Monthly Uptime | Business Credit | Enterprise Credit |
|----------------|-----------------|-------------------|
| Below target to 99.0% | 15% | 20% |
| 98.0% - 99.0% | 30% | 40% |
| 95.0% - 98.0% | 50% | 75% |
| < 95.0% | 100% | 100% |

### 6.3 Credit Request Process

To request credits:

1. **Submit Request:** Email sla@crmplatform.my within 30 days of incident
2. **Include:**
   - Account ID and company name
   - Date(s) and time(s) of downtime
   - Description of the issue
   - Impact on your business

3. **Review:** We will review within 10 business days
4. **Credit Application:** Credits applied to next billing cycle

### 6.4 Credit Limitations

- Maximum credit per month: 100% of monthly fee
- Credits are non-transferable and have no cash value
- Credits cannot be carried over beyond 12 months
- Credits do not apply to one-time fees or add-ons

### 6.5 Exclusions

Credits do not apply when downtime is caused by:
- Scheduled maintenance
- Customer actions or configurations
- Force majeure events
- Third-party services
- Beta/preview features
- Accounts with overdue payments

---

## 7. Performance Benchmarks

### 7.1 API Response Times

| Endpoint Type | Target (P95) | Maximum |
|---------------|--------------|---------|
| Read (GET) | 200ms | 1 second |
| Write (POST/PUT) | 500ms | 2 seconds |
| Search/List | 500ms | 3 seconds |
| Reports/Analytics | 2 seconds | 10 seconds |
| Bulk Operations | 5 seconds | 30 seconds |

### 7.2 Web Application Performance

| Metric | Target |
|--------|--------|
| Initial Page Load | < 3 seconds |
| Subsequent Navigation | < 1 second |
| Search Results | < 2 seconds |
| Dashboard Rendering | < 2 seconds |

### 7.3 Data Processing

| Operation | Target |
|-----------|--------|
| Import (1,000 records) | < 30 seconds |
| Export (10,000 records) | < 60 seconds |
| Report Generation | < 30 seconds |
| Backup Completion | < 1 hour |

---

## 8. Data Protection

### 8.1 Backup Schedule

| Data Type | Frequency | Retention |
|-----------|-----------|-----------|
| Full Database | Daily | 30 days |
| Transaction Logs | Continuous | 7 days |
| Configuration | Daily | 90 days |
| File Attachments | Daily | 30 days |

### 8.2 Recovery Objectives

| Metric | Target |
|--------|--------|
| Recovery Point Objective (RPO) | < 1 hour |
| Recovery Time Objective (RTO) | < 4 hours |

### 8.3 Disaster Recovery

- Geographically distributed backups
- Annual DR testing
- DR test results available to Enterprise customers upon request

---

## 9. Security Commitments

### 9.1 Security Standards

- Encryption: TLS 1.2+ in transit, AES-256 at rest
- Authentication: MFA available for all accounts
- Access Control: Role-based permissions
- Audit Logging: All data access logged

### 9.2 Vulnerability Management

- Regular security assessments
- Penetration testing: Annual (minimum)
- Critical vulnerabilities: Patched within 24 hours
- High vulnerabilities: Patched within 7 days

### 9.3 Compliance

- PDPA (Malaysia) compliant
- GDPR compliant (for EU customers)
- SOC 2 Type II (in progress)

---

## 10. Monitoring and Reporting

### 10.1 Status Page

**URL:** status.crmplatform.my

Includes:
- Real-time service status
- Incident history (90 days)
- Scheduled maintenance calendar
- Uptime statistics

### 10.2 Monthly Reports (Business/Enterprise)

Available on request:
- Uptime statistics
- Performance metrics
- Support ticket summary
- Security incident summary

### 10.3 Custom Reporting (Enterprise)

- Quarterly business reviews
- Custom SLA dashboards
- API for status integration

---

## 11. Escalation Procedures

### 11.1 Escalation Path

| Level | Contact | Response Time |
|-------|---------|---------------|
| Level 1 | Support Team | Per SLA |
| Level 2 | Support Manager | 2 hours |
| Level 3 | Operations Director | 4 hours |
| Level 4 | CTO | 8 hours |

### 11.2 How to Escalate

1. **Email:** escalation@crmplatform.my
2. **Phone (Enterprise):** [Escalation Hotline]
3. **In Ticket:** Request escalation in support ticket

### 11.3 Executive Escalation (Enterprise)

Enterprise customers may request executive escalation for:
- Repeated SLA breaches
- Critical business impact
- Strategic concerns

---

## 12. SLA Review and Changes

### 12.1 Annual Review

This SLA is reviewed annually and may be updated to:
- Improve service commitments
- Add new services
- Reflect industry standards

### 12.2 Notification of Changes

- 30 days advance notice for changes
- Material changes communicated via email
- Changes posted on our website

### 12.3 Grandfathering

If SLA commitments are reduced:
- Existing customers retain current commitments for contract term
- Option to renew under new or existing terms

---

## 13. Definitions

| Term | Definition |
|------|------------|
| **Downtime** | Period when the Service is unavailable or unresponsive |
| **Scheduled Maintenance** | Pre-announced maintenance window |
| **Response Time** | Time from ticket creation to first substantive response |
| **Resolution Time** | Time to resolve or provide acceptable workaround |
| **Service Credit** | Credit applied to customer account as compensation |
| **P95** | 95th percentile - 95% of requests meet this target |

---

## 14. Contact Information

**General Support:**
- Email: support@crmplatform.my
- Help Center: help.crmplatform.my

**SLA/Credits:**
- Email: sla@crmplatform.my

**Escalations:**
- Email: escalation@crmplatform.my
- Phone (Enterprise): [Escalation Hotline]

**Status Updates:**
- Website: status.crmplatform.my
- Subscribe to updates: status.crmplatform.my/subscribe

---

## 15. Agreement

This SLA is incorporated into and forms part of the Terms of Service. By using the Service, you acknowledge that you have read and agree to this SLA.

---

*Document Version: 1.0*
*Last Review Date: February 2026*
*Next Review Date: February 2027*
