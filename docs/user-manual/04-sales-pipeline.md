# Sales Pipeline

Master the visual Kanban board for managing your sales opportunities.

---

## Understanding the Pipeline

### What is a Pipeline?

A **sales pipeline** visualizes your sales process from initial contact to closed deal. It shows where each opportunity stands and helps you forecast revenue.

### Pipeline Stages

Default stages in your pipeline:

```
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Lead      │ │  Qualified  │ │  Proposal   │ │ Negotiation │ │   Closed    │
│             │ │             │ │             │ │             │ │   Won 🎉    │
│  [Card 1]   │ │  [Card 4]   │ │  [Card 7]   │ │  [Card 9]   │ │  [Card 11]  │
│  [Card 2]   │ │  [Card 5]   │ │  [Card 8]   │ │             │ │             │
│  [Card 3]   │ │  [Card 6]   │ │             │ │             │ │             │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
     10%            25%             50%             75%             100%
```

| Stage | Probability | Description |
|-------|-------------|-------------|
| Lead | 10% | Initial interest identified |
| Qualified | 25% | Need and budget confirmed |
| Proposal | 50% | Proposal sent to customer |
| Negotiation | 75% | Active discussions ongoing |
| Closed Won | 100% | Deal successfully closed |

---

## The Kanban Board

### Accessing the Pipeline

1. Click **Pipeline** in the sidebar
2. Select your pipeline (if you have multiple)

### Board Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  [Pipeline Selector ▼]   Total: RM 1,250,000   Deals: 24       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Stage 1          Stage 2          Stage 3          Stage 4    │
│  ───────          ───────          ───────          ───────    │
│  RM 150K (5)      RM 300K (8)      RM 500K (6)      RM 300K (5)│
│                                                                 │
│  ┌─────────┐      ┌─────────┐      ┌─────────┐      ┌─────────┐│
│  │ Deal A  │      │ Deal D  │      │ Deal G  │      │ Deal J  ││
│  │ RM 30K  │      │ RM 45K  │      │ Deal 80K│      │ RM 60K  ││
│  │ Acme Co │      │ Tech Ltd│      │ BigCorp │      │ Star Inc││
│  └─────────┘      └─────────┘      └─────────┘      └─────────┘│
│                                                                 │
│  ┌─────────┐      ┌─────────┐      ┌─────────┐                 │
│  │ Deal B  │      │ Deal E  │      │ Deal H  │                 │
│  └─────────┘      └─────────┘      └─────────┘                 │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Deal Cards

Each card displays:
- Deal name
- Value (RM amount)
- Customer name
- Expected close date
- Assigned owner
- Priority indicator

---

## Drag-and-Drop Functionality

### Moving Deals Between Stages

1. Click and hold a deal card
2. Drag it to the target stage
3. Drop to move the deal

The deal's probability automatically updates based on the new stage.

### Reordering Within a Stage

Drag cards up or down within a column to prioritize:
- Top = Higher priority
- Bottom = Lower priority

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `→` | Move deal to next stage |
| `←` | Move deal to previous stage |
| `↑` | Move deal up in stage |
| `↓` | Move deal down in stage |

---

## Creating Opportunities

### From the Pipeline View

1. Click **+ Add Deal** at the top of any stage
2. Fill in the quick-create form:

| Field | Description |
|-------|-------------|
| Deal Name | Descriptive name |
| Customer | Select or create |
| Value | Expected deal value |
| Expected Close | When you expect to close |

3. Click **Create**

### Detailed Opportunity Creation

For more complete information:

1. Click **+ New Opportunity** (main button)
2. Fill all fields:

| Field | Required | Description |
|-------|----------|-------------|
| Name | Yes | Deal name |
| Customer | No* | Associated customer |
| Value | Yes | Deal value in RM |
| Stage | Yes | Starting pipeline stage |
| Probability | Auto | Based on stage |
| Expected Close | Recommended | Target close date |
| Owner | Auto | Assigned salesperson |
| Notes | No | Additional context |

*Recommended to link to a customer

---

## Managing Opportunities

### Quick View Popup

Click a deal card to open quick view:
- Key information at a glance
- Quick actions (Win, Lose, Move)
- Click **View Full Details** for complete page

### Full Detail View

Access complete opportunity information:

#### Overview Tab
- Deal summary
- Customer info
- Timeline progress
- Value and probability

#### Products Tab
Add products/services to the deal:
1. Click **+ Add Product**
2. Select product or enter custom
3. Set quantity and price
4. Total auto-calculates

#### Contacts Tab
Link relevant customer contacts:
1. Click **+ Add Contact**
2. Select from customer's contacts
3. Set role (Decision Maker, Influencer, etc.)

#### Activity Tab
See complete interaction history.

---

## Pipeline Stages

### Stage Properties

Each stage has:
- **Name**: Display label
- **Position**: Order in pipeline
- **Probability**: Win likelihood (%)
- **Color**: Visual indicator

### Viewing Stage Details

Click the stage header to see:
- Total value in stage
- Number of deals
- Average deal age
- Conversion rate

---

## Closing Deals

### Winning a Deal

When you close a deal successfully:

1. Open the opportunity
2. Click **Win Deal** button
3. Fill in closing details:
   - Actual close date
   - Final value (if different)
   - Closing notes
4. Click **Confirm Win**

What happens:
- Deal moves to "Closed Won"
- Customer status may update
- Revenue is recorded
- Celebration animation! 🎉

### Losing a Deal

When a deal is lost:

1. Open the opportunity
2. Click **Lose Deal** button
3. Select loss reason:
   - Price too high
   - Competitor selected
   - No budget
   - No decision
   - Timing not right
   - Other
4. Add notes explaining
5. Click **Confirm Loss**

### Reopening a Closed Deal

If circumstances change:

1. Open the closed opportunity
2. Click **Reopen** button
3. Select the stage to return to
4. Add a note explaining why
5. Confirm

---

## Filtering and Views

### Filter Options

| Filter | Description |
|--------|-------------|
| Pipeline | Which pipeline to view |
| Owner | Deals by salesperson |
| Customer | Deals for specific customer |
| Value Range | Min/max deal value |
| Close Date | Expected close date range |
| Stage | Specific stages only |

### Saved Views

Create custom views:
1. Apply your filters
2. Click **Save View**
3. Name the view
4. Access from view dropdown

Example views:
- "My deals closing this month"
- "High-value opportunities"
- "Stalled deals (30+ days)"

---

## Pipeline Analytics

### Stage Conversion Rates

```
Lead (100%) → Qualified (60%) → Proposal (40%) → Negotiation (25%) → Won (15%)
```

View conversion rates:
1. Click pipeline settings icon
2. Select **Analytics**
3. See stage-by-stage conversion

### Pipeline Velocity

How fast deals move:
- Average days per stage
- Total cycle time
- Bottleneck identification

### Forecasting

Predict future revenue:
- Weighted pipeline value
- Expected close by period
- Probability-adjusted totals

---

## Multiple Pipelines

### When to Use Multiple Pipelines

- Different products/services
- Different sales processes
- Different teams
- Different customer types

### Switching Pipelines

1. Click the pipeline dropdown
2. Select desired pipeline
3. Board updates to show that pipeline

### Creating a New Pipeline

(Admin permission required)

1. Go to **Settings** → **Pipelines**
2. Click **+ Create Pipeline**
3. Name the pipeline
4. Add stages with:
   - Name
   - Order
   - Probability
   - Color
5. Save

---

## Best Practices

### Pipeline Hygiene

✅ Update deals daily
✅ Move stale deals appropriately
✅ Keep values accurate
✅ Set realistic close dates
✅ Add notes on every update

### Stage Discipline

✅ Clear criteria for each stage
✅ Don't skip stages
✅ Move backward if needed
✅ Close lost deals promptly

### Forecasting Accuracy

✅ Use correct probabilities
✅ Update close dates when they slip
✅ Don't inflate values
✅ Remove dead deals

---

## Common Patterns

### Deal Progression Example

```
Day 1:  Lead created (10%)
Day 3:  Meeting held, qualified (25%)
Day 7:  Proposal sent (50%)
Day 14: Customer reviewing
Day 21: Negotiation started (75%)
Day 28: Contract signed - WON (100%)
```

### Red Flags

⚠️ Deal stuck in stage too long
⚠️ Close date keeps moving
⚠️ No recent activity
⚠️ Missing customer contact
⚠️ Value seems unrealistic

---

## Troubleshooting

### Can't Move Deal

- Check if you're the owner
- Verify stage is not disabled
- May need admin permission

### Deal Disappeared

- Check filters (might be filtered out)
- Look in Closed Won/Lost
- Search by deal name

### Wrong Stage Probability

- Contact admin to update stage settings
- Probabilities are set per-stage, not per-deal

---

[← Previous: Customer Management](./03-customer-management.md) | [Next: Reporting →](./05-reporting.md)
