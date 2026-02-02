# Lead Management

Learn how to capture, qualify, and convert leads into customers.

---

## What is a Lead?

A **lead** is a potential customer who has shown interest in your products or services but hasn't yet made a purchase decision. In the CRM, leads are tracked separately from customers until they're ready to be converted.

### Lead Lifecycle

```
┌──────────┐    ┌───────────┐    ┌───────────┐    ┌───────────┐
│   New    │ →  │ Contacted │ →  │ Qualified │ →  │ Converted │
└──────────┘    └───────────┘    └───────────┘    └───────────┘
                                        │
                                        ↓
                                 ┌─────────────┐
                                 │ Unqualified │
                                 └─────────────┘
```

---

## Creating a New Lead

### Method 1: Manual Entry

1. Navigate to **Leads** in the sidebar
2. Click **+ New Lead** button
3. Fill in the lead information:

| Field | Description | Required |
|-------|-------------|----------|
| First Name | Lead's first name | Yes |
| Last Name | Lead's last name | Yes |
| Email | Contact email address | Yes |
| Phone | Contact phone number | No |
| Company | Company name | Recommended |
| Title | Job title | No |
| Source | How the lead found you | Recommended |
| Industry | Business sector | No |
| Website | Company website | No |
| Notes | Initial notes or context | No |

4. Click **Create Lead**

### Method 2: Bulk Import

For importing many leads at once:

1. Go to **Leads** → **Import**
2. Download the CSV template
3. Fill in your lead data
4. Upload the completed CSV file
5. Map the columns to CRM fields
6. Review and confirm the import

**CSV Template Columns:**
```
first_name,last_name,email,phone,company,title,source,industry,notes
John,Doe,john@example.com,+60123456789,Acme Corp,CEO,Website,Technology,Interested in enterprise plan
```

### Lead Sources

Track where your leads come from:

| Source | Description |
|--------|-------------|
| Website | Contact form submissions |
| Referral | Word-of-mouth referrals |
| Social Media | LinkedIn, Facebook, etc. |
| Trade Show | Events and exhibitions |
| Cold Call | Outbound prospecting |
| Advertisement | Paid campaigns |
| Partner | Channel partner referrals |

---

## Lead List View

The lead list shows all your leads with filtering and search capabilities.

### Filtering Leads

Use filters to find specific leads:

- **Status**: New, Contacted, Qualified, Unqualified, Converted
- **Source**: Website, Referral, Social Media, etc.
- **Score Label**: Hot, Warm, Cold
- **Owner**: Assigned team member
- **Date Range**: Creation date filter

### Sorting

Click column headers to sort by:
- Name (A-Z or Z-A)
- Company
- Created Date
- Last Updated
- Score

### Search

Use the search box to find leads by:
- Name
- Email
- Company
- Phone number

---

## Lead Detail View

Click on a lead to see the full detail view.

### Overview Tab

- Contact information
- Company details
- Lead score and label
- Current status
- Assigned owner

### Activity Tab

Timeline of all interactions:
- Emails sent
- Calls made
- Notes added
- Status changes

### Notes Tab

Add and view notes about the lead:
1. Click **+ Add Note**
2. Enter your note
3. Click **Save**

Pin important notes to keep them at the top.

---

## Lead Scoring

Lead scoring helps prioritize your leads based on their likelihood to convert.

### Score Labels

| Label | Score Range | Description |
|-------|-------------|-------------|
| 🔥 Hot | 70-100 | High priority, ready to buy |
| 🌡️ Warm | 40-69 | Interested, needs nurturing |
| ❄️ Cold | 0-39 | Low engagement, long-term |

### Automatic Scoring Factors

The system considers:
- Company size
- Industry match
- Engagement level
- Source quality
- Activity recency

### Manual Score Adjustment

1. Open the lead detail
2. Click the score indicator
3. Adjust the score (0-100)
4. Add a reason for the change
5. Save

---

## Qualifying Leads

Qualification determines if a lead is a good fit for your business.

### BANT Criteria

Use BANT to qualify leads:

- **B**udget - Do they have budget?
- **A**uthority - Are they a decision-maker?
- **N**eed - Do they have a genuine need?
- **T**imeline - When do they plan to purchase?

### To Qualify a Lead

1. Open the lead detail
2. Click **Qualify Lead** button
3. Optionally adjust the score
4. Add qualification notes
5. Confirm

The lead status changes to "Qualified."

### To Disqualify a Lead

If a lead isn't a good fit:

1. Open the lead detail
2. Click **More Actions** → **Disqualify**
3. Select a reason:
   - No budget
   - Not a decision maker
   - No current need
   - Timeline too long
   - Competitor selected
   - Other
4. Add notes explaining the decision
5. Confirm

---

## Converting Leads to Opportunities

When a qualified lead is ready to move forward, convert them.

### Conversion Process

1. Open the qualified lead
2. Click **Convert Lead** button
3. Choose conversion options:

| Option | Description |
|--------|-------------|
| Create Customer | Creates a new customer record |
| Create Opportunity | Creates a sales opportunity |
| Both | Creates both records (recommended) |

4. If creating an opportunity:
   - Enter opportunity name
   - Set expected value
   - Select pipeline stage

5. Click **Convert**

### What Happens After Conversion

- Lead status changes to "Converted"
- New customer record created (if selected)
- New opportunity created (if selected)
- All lead notes and activities are preserved
- Lead becomes read-only (archived)

### Accessing Converted Records

From the converted lead page:
- Click **View Customer** to see the customer record
- Click **View Opportunity** to see the opportunity

---

## Bulk Lead Actions

Perform actions on multiple leads at once.

### Bulk Select

1. Check the boxes next to leads
2. Or click "Select All" to select visible leads

### Available Bulk Actions

| Action | Description |
|--------|-------------|
| Assign | Assign selected leads to a team member |
| Update Status | Change status of all selected |
| Delete | Remove selected leads |
| Export | Export selected to CSV |

### Bulk Assignment

1. Select leads
2. Click **Bulk Actions** → **Assign**
3. Choose a team member
4. Confirm

---

## Lead Assignment

Leads can be assigned to team members for follow-up.

### Manual Assignment

1. Open the lead
2. Click the owner field
3. Select a team member
4. Save

### Round-Robin Assignment

If enabled by your admin:
- New leads are automatically assigned
- Distribution is balanced across the team
- Assignment rules can be customized

### Claiming Unassigned Leads

1. Go to **Leads** → Filter by "Unassigned"
2. Find leads you want to work
3. Click **Claim** or assign to yourself

---

## Lead Automation

Automate repetitive tasks with lead workflows.

### Auto-Assignment Rules

Set rules like:
- "Assign website leads to Sales Team A"
- "Assign enterprise leads to Senior Reps"

### Lead Nurturing

Set up automated follow-ups:
- Email sequences
- Task reminders
- Status change triggers

Contact your admin to configure automations.

---

## Best Practices

### Do's

✅ Respond to new leads quickly (within 5 minutes if possible)
✅ Always add notes after each interaction
✅ Keep lead information up to date
✅ Qualify leads before spending too much time
✅ Follow up consistently

### Don'ts

❌ Don't let leads go stale (no contact for 30+ days)
❌ Don't skip the qualification process
❌ Don't convert leads prematurely
❌ Don't delete leads without good reason (disqualify instead)

---

## Troubleshooting

### Lead Not Appearing

- Check your filters (might be filtered out)
- Search by email to find it
- Verify you have permission to see it

### Can't Convert Lead

- Lead must be in "Qualified" status
- You need conversion permissions
- Check for required field validation

### Import Errors

- Verify CSV format matches template
- Check for duplicate emails
- Ensure required fields are filled

---

[← Previous: Getting Started](./01-getting-started.md) | [Next: Customer Management →](./03-customer-management.md)
