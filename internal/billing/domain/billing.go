// Package domain provides domain models for the billing service.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionStatus represents the status of a subscription.
type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionTrialing  SubscriptionStatus = "trialing"
	SubscriptionPastDue   SubscriptionStatus = "past_due"
	SubscriptionCanceled  SubscriptionStatus = "canceled"
	SubscriptionUnpaid    SubscriptionStatus = "unpaid"
	SubscriptionPaused    SubscriptionStatus = "paused"
)

// PaymentStatus represents the status of a payment.
type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentCompleted PaymentStatus = "completed"
	PaymentFailed    PaymentStatus = "failed"
	PaymentRefunded  PaymentStatus = "refunded"
	PaymentCanceled  PaymentStatus = "canceled"
)

// PaymentMethod represents a payment method type.
type PaymentMethod string

const (
	PaymentMethodCard   PaymentMethod = "card"
	PaymentMethodFPX    PaymentMethod = "fpx"
	PaymentMethodEWallet PaymentMethod = "ewallet"
	PaymentMethodBank   PaymentMethod = "bank_transfer"
)

// PaymentProvider represents a payment provider.
type PaymentProvider string

const (
	ProviderStripe       PaymentProvider = "stripe"
	ProviderToyyibPay    PaymentProvider = "toyyibpay"
	ProviderBillplz      PaymentProvider = "billplz"
	ProviderRevenueMonster PaymentProvider = "revenue_monster"
)

// Currency represents supported currencies.
type Currency string

const (
	CurrencyMYR Currency = "MYR"
	CurrencyUSD Currency = "USD"
	CurrencySGD Currency = "SGD"
)

// Plan represents a subscription plan.
type Plan struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Description  string    `json:"description"`
	PriceMonthly int64     `json:"price_monthly"` // In cents/sen
	PriceYearly  int64     `json:"price_yearly"`  // In cents/sen
	Currency     Currency  `json:"currency"`
	Features     []string  `json:"features"`
	Limits       PlanLimits `json:"limits"`
	TrialDays    int       `json:"trial_days"`
	IsActive     bool      `json:"is_active"`
	IsPublic     bool      `json:"is_public"`
	Order        int       `json:"order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PlanLimits defines resource limits for a plan.
type PlanLimits struct {
	MaxUsers       int   `json:"max_users"`
	MaxCustomers   int   `json:"max_customers"`
	MaxLeads       int   `json:"max_leads"`
	MaxDeals       int   `json:"max_deals"`
	MaxStorage     int64 `json:"max_storage_bytes"`
	MaxAPIRequests int   `json:"max_api_requests_per_day"`
	Features       map[string]bool `json:"features"`
}

// Subscription represents a tenant's subscription.
type Subscription struct {
	ID                   uuid.UUID          `json:"id"`
	TenantID             uuid.UUID          `json:"tenant_id"`
	PlanID               uuid.UUID          `json:"plan_id"`
	Status               SubscriptionStatus `json:"status"`
	BillingCycle         string             `json:"billing_cycle"` // monthly, yearly
	CurrentPeriodStart   time.Time          `json:"current_period_start"`
	CurrentPeriodEnd     time.Time          `json:"current_period_end"`
	TrialStart           *time.Time         `json:"trial_start,omitempty"`
	TrialEnd             *time.Time         `json:"trial_end,omitempty"`
	CancelAtPeriodEnd    bool               `json:"cancel_at_period_end"`
	CanceledAt           *time.Time         `json:"canceled_at,omitempty"`
	CancellationReason   string             `json:"cancellation_reason,omitempty"`
	PaymentProvider      PaymentProvider    `json:"payment_provider"`
	ExternalID           string             `json:"external_id"` // Stripe subscription ID, etc.
	DefaultPaymentMethod uuid.UUID          `json:"default_payment_method"`
	Metadata             map[string]string  `json:"metadata"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

// CustomerPaymentMethod represents a saved payment method.
type CustomerPaymentMethod struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	Type           PaymentMethod   `json:"type"`
	Provider       PaymentProvider `json:"provider"`
	ExternalID     string          `json:"external_id"`
	IsDefault      bool            `json:"is_default"`

	// Card details
	CardBrand      string `json:"card_brand,omitempty"`
	CardLast4      string `json:"card_last4,omitempty"`
	CardExpMonth   int    `json:"card_exp_month,omitempty"`
	CardExpYear    int    `json:"card_exp_year,omitempty"`

	// Bank details (FPX)
	BankCode       string `json:"bank_code,omitempty"`
	BankName       string `json:"bank_name,omitempty"`

	// E-wallet details
	WalletType     string `json:"wallet_type,omitempty"`

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Invoice represents a billing invoice.
type Invoice struct {
	ID               uuid.UUID       `json:"id"`
	TenantID         uuid.UUID       `json:"tenant_id"`
	SubscriptionID   uuid.UUID       `json:"subscription_id"`
	InvoiceNumber    string          `json:"invoice_number"`
	Status           string          `json:"status"` // draft, open, paid, void, uncollectible
	Currency         Currency        `json:"currency"`
	Subtotal         int64           `json:"subtotal"`
	Tax              int64           `json:"tax"`
	Total            int64           `json:"total"`
	AmountDue        int64           `json:"amount_due"`
	AmountPaid       int64           `json:"amount_paid"`
	LineItems        []InvoiceItem   `json:"line_items"`
	DueDate          time.Time       `json:"due_date"`
	PaidAt           *time.Time      `json:"paid_at,omitempty"`
	PaymentIntentID  string          `json:"payment_intent_id,omitempty"`
	HostedURL        string          `json:"hosted_url,omitempty"`
	PDFURL           string          `json:"pdf_url,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// InvoiceItem represents a line item on an invoice.
type InvoiceItem struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	UnitPrice   int64     `json:"unit_price"`
	Amount      int64     `json:"amount"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
}

// Payment represents a payment transaction.
type Payment struct {
	ID                uuid.UUID       `json:"id"`
	TenantID          uuid.UUID       `json:"tenant_id"`
	InvoiceID         uuid.UUID       `json:"invoice_id"`
	PaymentMethodID   uuid.UUID       `json:"payment_method_id"`
	Provider          PaymentProvider `json:"provider"`
	ExternalID        string          `json:"external_id"`
	Amount            int64           `json:"amount"`
	Currency          Currency        `json:"currency"`
	Status            PaymentStatus   `json:"status"`
	FailureCode       string          `json:"failure_code,omitempty"`
	FailureMessage    string          `json:"failure_message,omitempty"`
	RefundedAmount    int64           `json:"refunded_amount"`
	Metadata          map[string]string `json:"metadata"`
	ProcessedAt       *time.Time      `json:"processed_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// UsageRecord represents metered usage for a tenant.
type UsageRecord struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	MetricName string    `json:"metric_name"`
	Quantity   int64     `json:"quantity"`
	Timestamp  time.Time `json:"timestamp"`
}

// BillingCustomer represents a billing customer (linked to tenant).
type BillingCustomer struct {
	ID                uuid.UUID       `json:"id"`
	TenantID          uuid.UUID       `json:"tenant_id"`
	Email             string          `json:"email"`
	Name              string          `json:"name"`
	Phone             string          `json:"phone,omitempty"`
	TaxID             string          `json:"tax_id,omitempty"`
	Provider          PaymentProvider `json:"provider"`
	ExternalID        string          `json:"external_id"` // Stripe customer ID
	Balance           int64           `json:"balance"`
	Currency          Currency        `json:"currency"`
	BillingAddress    Address         `json:"billing_address"`
	Metadata          map[string]string `json:"metadata"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// Address represents a billing/shipping address.
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Coupon represents a discount coupon.
type Coupon struct {
	ID             uuid.UUID `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Type           string    `json:"type"` // percent, fixed
	DiscountAmount int64     `json:"discount_amount"` // Percentage or fixed amount
	Currency       Currency  `json:"currency,omitempty"`
	Duration       string    `json:"duration"` // once, repeating, forever
	DurationMonths int       `json:"duration_months,omitempty"`
	MaxRedemptions int       `json:"max_redemptions,omitempty"`
	TimesRedeemed  int       `json:"times_redeemed"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// WebhookEvent represents a payment provider webhook event.
type WebhookEvent struct {
	ID         uuid.UUID       `json:"id"`
	Provider   PaymentProvider `json:"provider"`
	EventType  string          `json:"event_type"`
	ExternalID string          `json:"external_id"`
	Payload    []byte          `json:"payload"`
	Processed  bool            `json:"processed"`
	ProcessedAt *time.Time     `json:"processed_at,omitempty"`
	Error      string          `json:"error,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// NewPlan creates a new Plan with default values.
func NewPlan(name, code string, priceMonthly, priceYearly int64, currency Currency) *Plan {
	return &Plan{
		ID:           uuid.New(),
		Name:         name,
		Code:         code,
		PriceMonthly: priceMonthly,
		PriceYearly:  priceYearly,
		Currency:     currency,
		TrialDays:    14,
		IsActive:     true,
		IsPublic:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// NewSubscription creates a new Subscription.
func NewSubscription(tenantID, planID uuid.UUID, billingCycle string, trialDays int) *Subscription {
	now := time.Now()

	sub := &Subscription{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		PlanID:             planID,
		Status:             SubscriptionActive,
		BillingCycle:       billingCycle,
		CurrentPeriodStart: now,
		Metadata:           make(map[string]string),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Set period end based on billing cycle
	if billingCycle == "yearly" {
		sub.CurrentPeriodEnd = now.AddDate(1, 0, 0)
	} else {
		sub.CurrentPeriodEnd = now.AddDate(0, 1, 0)
	}

	// Set up trial if applicable
	if trialDays > 0 {
		sub.Status = SubscriptionTrialing
		trialStart := now
		trialEnd := now.AddDate(0, 0, trialDays)
		sub.TrialStart = &trialStart
		sub.TrialEnd = &trialEnd
	}

	return sub
}

// IsInTrial returns true if subscription is in trial period.
func (s *Subscription) IsInTrial() bool {
	return s.Status == SubscriptionTrialing && s.TrialEnd != nil && time.Now().Before(*s.TrialEnd)
}

// DaysUntilRenewal returns days until next billing date.
func (s *Subscription) DaysUntilRenewal() int {
	if s.Status == SubscriptionCanceled {
		return -1
	}
	days := int(time.Until(s.CurrentPeriodEnd).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// Cancel cancels the subscription at period end.
func (s *Subscription) Cancel(reason string) {
	now := time.Now()
	s.CancelAtPeriodEnd = true
	s.CanceledAt = &now
	s.CancellationReason = reason
	s.UpdatedAt = now
}

// CancelImmediately cancels the subscription immediately.
func (s *Subscription) CancelImmediately(reason string) {
	now := time.Now()
	s.Status = SubscriptionCanceled
	s.CanceledAt = &now
	s.CancellationReason = reason
	s.CurrentPeriodEnd = now
	s.UpdatedAt = now
}
