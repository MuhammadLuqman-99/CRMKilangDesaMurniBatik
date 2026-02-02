# Customer Management

Learn how to manage your customer relationships effectively.

---

## What is a Customer?

A **customer** is an individual or organization that has purchased from you or has a formal business relationship with your company. Customers are created when leads convert or can be added directly.

### Customer vs Lead

| Aspect | Lead | Customer |
|--------|------|----------|
| Status | Potential buyer | Active relationship |
| Created from | Forms, imports, manual | Lead conversion, direct |
| Has deals | No | Yes (opportunities/deals) |
| Multiple contacts | No | Yes |
| Full history | Limited | Complete |

---

## Creating a Customer

### Method 1: From Lead Conversion

See [Lead Management - Converting Leads](./02-lead-management.md#converting-leads-to-opportunities)

### Method 2: Direct Creation

1. Navigate to **Customers** in the sidebar
2. Click **+ New Customer** button
3. Fill in customer information:

| Field | Description | Required |
|-------|-------------|----------|
| Company Name | Official company name | Yes |
| Email | Primary email | Yes |
| Phone | Main phone number | No |
| Industry | Business sector | Recommended |
| Website | Company website | No |
| Address | Physical address | No |
| Description | About the customer | No |

4. Click **Create Customer**

---

## Customer List View

### Viewing Customers

The customer list displays all your customers with key information at a glance.

### Filtering Options

| Filter | Options |
|--------|---------|
| Status | Active, Inactive, Prospect, Churned |
| Segment | Enterprise, SMB, etc. |
| Industry | Technology, Finance, etc. |
| Owner | Assigned account manager |
| Date Created | Date range |

### Search

Search by:
- Company name
- Email address
- Phone number
- Contact names

### Export

Export customers to CSV:
1. Apply desired filters
2. Click **Export** button
3. Download the CSV file

---

## Customer Detail Page (360° View)

The customer detail page provides a complete view of the relationship.

### Overview Section

```
┌──────────────────────────────────────────────────────────────┐
│  Company Logo    ACME Corporation                            │
│                  Technology | Enterprise | Active            │
│  ─────────────────────────────────────────────────────────── │
│  📧 sales@acme.com    📞 +60123456789    🌐 acme.com        │
│  📍 Kuala Lumpur, Malaysia                                   │
├──────────────────────────────────────────────────────────────┤
│  Total Value: RM 250,000   │   Deals: 5   │   Since: 2024   │
└──────────────────────────────────────────────────────────────┘
```

### Tabs

1. **Overview** - Key metrics and quick info
2. **Contacts** - People at the company
3. **Activities** - Interaction history
4. **Notes** - Internal notes
5. **Deals** - Related opportunities/deals

---

## Contact Management

### Adding Contacts

Each customer can have multiple contacts (people at the company).

1. Open customer detail
2. Go to **Contacts** tab
3. Click **+ Add Contact**
4. Fill in contact details:

| Field | Description | Required |
|-------|-------------|----------|
| Name | Full name | Yes |
| Email | Email address | Yes |
| Phone | Direct line | No |
| Title | Job title | Recommended |
| Department | Team/department | No |
| Primary | Is this the main contact? | No |

5. Click **Save**

### Primary Contact

Mark one contact as "Primary":
- Shown first in lists
- Used as default for communications
- Click **Set as Primary** on any contact

### Updating Contacts

1. Click on a contact
2. Edit the fields
3. Click **Save Changes**

### Removing Contacts

1. Click on a contact
2. Click **Delete** button
3. Confirm deletion

---

## Activity History

Track all interactions with the customer.

### Activity Types

| Type | Icon | Description |
|------|------|-------------|
| Call | 📞 | Phone conversations |
| Email | ✉️ | Email correspondence |
| Meeting | 📅 | In-person or virtual meetings |
| Note | 📝 | Internal observations |
| Task | ✓ | Completed tasks |
| Deal Update | 💼 | Opportunity changes |

### Logging an Activity

1. Go to customer **Activities** tab
2. Click **+ Log Activity**
3. Select activity type
4. Fill in details:
   - Title (brief description)
   - Description (details)
   - Duration (for calls/meetings)
   - Outcome (positive/negative/neutral)
5. Click **Save**

### Viewing Activity Timeline

Activities are displayed chronologically:
- Most recent at top
- Filter by activity type
- Search within activities

---

## Customer Notes

### Adding Notes

1. Go to **Notes** tab
2. Click **+ Add Note**
3. Write your note (supports formatting)
4. Click **Save**

### Pinning Important Notes

Pin notes to keep them visible:
1. Click the pin icon on a note
2. Pinned notes appear at the top
3. Click again to unpin

### Note Best Practices

✅ Date your observations
✅ Include context
✅ Mention relevant people
✅ Use for handoff information

---

## Customer Status Management

### Status Types

| Status | Description |
|--------|-------------|
| **Active** | Current paying customer |
| **Inactive** | Customer with no recent activity |
| **Prospect** | Potential customer (not yet purchased) |
| **Churned** | Former customer who left |

### Changing Status

#### Activate Customer
```
Customer Detail → More Actions → Activate
```

#### Deactivate Customer
```
Customer Detail → More Actions → Deactivate
```

#### Block Customer
For problematic customers:
```
Customer Detail → More Actions → Block
- Enter reason
- Confirm
```

#### Unblock Customer
```
Customer Detail → More Actions → Unblock
```

---

## Customer Segments

Organize customers into segments for targeted actions.

### What are Segments?

Segments are groups of customers based on criteria:
- **Static**: Manually assigned
- **Dynamic**: Auto-updated based on rules

### Viewing Segments

1. Go to **Customers** → **Segments**
2. See list of all segments
3. Click a segment to see its members

### Creating a Segment

1. Click **+ Create Segment**
2. Enter segment name and description
3. Choose segment type:

**Static Segment:**
- Manually add/remove customers
- Good for special groups

**Dynamic Segment:**
- Define criteria (rules)
- Customers auto-added when matching
- Auto-removed when no longer matching

4. For dynamic segments, set rules:
   - "Industry equals Technology"
   - "Total value greater than 100000"
   - "Status equals Active"

5. Click **Create**

### Adding Customers to Segments

**For Static Segments:**
1. Open customer detail
2. Go to **More Actions** → **Add to Segment**
3. Select segment(s)
4. Confirm

**For Dynamic Segments:**
- Customers are added automatically based on rules
- Click **Refresh** on segment to update membership

---

## Importing Customers

### Bulk Import

1. Go to **Customers** → **Import**
2. Download CSV template
3. Fill in your data
4. Upload the file
5. Map columns:
   - Match CSV columns to CRM fields
   - Handle duplicates (skip, update, create new)
6. Preview and confirm
7. Monitor import progress

### Import Template

```csv
name,email,phone,industry,website,address,status
"Acme Corp",sales@acme.com,+60123456789,Technology,https://acme.com,"123 Main St, KL",active
```

### Handling Duplicates

When importing, choose:
- **Skip**: Don't import if email exists
- **Update**: Update existing record
- **Create New**: Always create new record

---

## Best Practices

### Customer Data Quality

✅ Keep contact information current
✅ Verify email addresses
✅ Regular data cleanup
✅ Document all interactions

### Relationship Building

✅ Log all touchpoints
✅ Set follow-up reminders
✅ Note personal details (appropriate)
✅ Track preferences

### Account Management

✅ Assign dedicated owners
✅ Regular check-ins
✅ Proactive outreach
✅ Monitor health indicators

---

## Troubleshooting

### Customer Not Found

- Check spelling in search
- Remove filters
- Verify permissions
- Search by email (most reliable)

### Can't Edit Customer

- Check if you're assigned as owner
- Verify edit permissions
- Customer may be archived

### Duplicate Customers

1. Identify the duplicates
2. Decide which to keep (usually older one)
3. Merge contacts and notes manually
4. Archive the duplicate

---

[← Previous: Lead Management](./02-lead-management.md) | [Next: Sales Pipeline →](./04-sales-pipeline.md)
