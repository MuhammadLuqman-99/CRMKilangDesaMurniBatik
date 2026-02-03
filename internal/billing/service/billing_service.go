// Package service provides the billing business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kilang-desa-murni/crm/internal/billing/domain"
)

// BillingService manages billing operations.
type BillingService struct {
	mu              sync.RWMutex
	plans           map[uuid.UUID]*domain.Plan
	subscriptions   map[uuid.UUID]*domain.Subscription
	customers       map[uuid.UUID]*domain.BillingCustomer
	paymentMethods  map[uuid.UUID]*domain.CustomerPaymentMethod
	invoices        map[uuid.UUID]*domain.Invoice
	payments        map[uuid.UUID]*domain.Payment
	coupons         map[string]*domain.Coupon
	providers       map[domain.PaymentProvider]PaymentProvider
}

// PaymentProvider interface for payment gateway integrations.
type PaymentProvider interface {
	// Customer operations
	CreateCustomer(ctx context.Context, customer *domain.BillingCustomer) (string, error)
	UpdateCustomer(ctx context.Context, customer *domain.BillingCustomer) error
	DeleteCustomer(ctx context.Context, externalID string) error

	// Payment method operations
	AttachPaymentMethod(ctx context.Context, customerID string, paymentMethodID string) error
	DetachPaymentMethod(ctx context.Context, paymentMethodID string) error
	ListPaymentMethods(ctx context.Context, customerID string) ([]domain.CustomerPaymentMethod, error)

	// Subscription operations
	CreateSubscription(ctx context.Context, customerID string, priceID string, trialDays int) (string, error)
	UpdateSubscription(ctx context.Context, subscriptionID string, priceID string) error
	CancelSubscription(ctx context.Context, subscriptionID string, cancelAtPeriodEnd bool) error

	// Payment operations
	CreatePaymentIntent(ctx context.Context, amount int64, currency domain.Currency, customerID string) (string, string, error) // returns ID, client_secret
	ConfirmPayment(ctx context.Context, paymentIntentID string, paymentMethodID string) error
	RefundPayment(ctx context.Context, paymentID string, amount int64) error

	// Invoice operations
	CreateInvoice(ctx context.Context, customerID string, items []domain.InvoiceItem) (string, error)
	FinalizeInvoice(ctx context.Context, invoiceID string) error
	VoidInvoice(ctx context.Context, invoiceID string) error

	// Webhook handling
	HandleWebhook(ctx context.Context, payload []byte, signature string) (*domain.WebhookEvent, error)

	// Provider info
	Name() domain.PaymentProvider
}

// NewBillingService creates a new BillingService.
func NewBillingService() *BillingService {
	bs := &BillingService{
		plans:          make(map[uuid.UUID]*domain.Plan),
		subscriptions:  make(map[uuid.UUID]*domain.Subscription),
		customers:      make(map[uuid.UUID]*domain.BillingCustomer),
		paymentMethods: make(map[uuid.UUID]*domain.CustomerPaymentMethod),
		invoices:       make(map[uuid.UUID]*domain.Invoice),
		payments:       make(map[uuid.UUID]*domain.Payment),
		coupons:        make(map[string]*domain.Coupon),
		providers:      make(map[domain.PaymentProvider]PaymentProvider),
	}

	// Initialize default plans
	bs.initializeDefaultPlans()

	return bs
}

// RegisterProvider registers a payment provider.
func (s *BillingService) RegisterProvider(provider PaymentProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[provider.Name()] = provider
}

// initializeDefaultPlans sets up the default subscription plans.
func (s *BillingService) initializeDefaultPlans() {
	plans := []*domain.Plan{
		{
			ID:           uuid.New(),
			Name:         "Free",
			Code:         "free",
			Description:  "Perfect for trying out CRM Platform",
			PriceMonthly: 0,
			PriceYearly:  0,
			Currency:     domain.CurrencyMYR,
			Features: []string{
				"1 User",
				"100 Customers",
				"100 Leads",
				"Basic Pipeline",
				"Email Support",
			},
			Limits: domain.PlanLimits{
				MaxUsers:       1,
				MaxCustomers:   100,
				MaxLeads:       100,
				MaxDeals:       50,
				MaxStorage:     100 * 1024 * 1024, // 100 MB
				MaxAPIRequests: 100,
				Features: map[string]bool{
					"api_access":   false,
					"custom_fields": false,
					"reports":      false,
					"integrations": false,
				},
			},
			TrialDays: 0,
			IsActive:  true,
			IsPublic:  true,
			Order:     1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:           uuid.New(),
			Name:         "Professional",
			Code:         "professional",
			Description:  "For growing teams who need more power",
			PriceMonthly: 9900, // RM 99.00
			PriceYearly:  99000, // RM 990.00 (2 months free)
			Currency:     domain.CurrencyMYR,
			Features: []string{
				"Up to 5 Users",
				"Unlimited Customers",
				"Unlimited Leads",
				"Multiple Pipelines",
				"Email Integration",
				"Basic Reports",
				"Priority Email Support",
			},
			Limits: domain.PlanLimits{
				MaxUsers:       5,
				MaxCustomers:   -1, // Unlimited
				MaxLeads:       -1,
				MaxDeals:       -1,
				MaxStorage:     1 * 1024 * 1024 * 1024, // 1 GB
				MaxAPIRequests: 5000,
				Features: map[string]bool{
					"api_access":    false,
					"custom_fields": true,
					"reports":       true,
					"integrations":  true,
				},
			},
			TrialDays: 14,
			IsActive:  true,
			IsPublic:  true,
			Order:     2,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:           uuid.New(),
			Name:         "Business",
			Code:         "business",
			Description:  "For established businesses requiring advanced features",
			PriceMonthly: 24900, // RM 249.00
			PriceYearly:  249000, // RM 2,490.00 (2 months free)
			Currency:     domain.CurrencyMYR,
			Features: []string{
				"Up to 25 Users",
				"Unlimited Everything",
				"API Access",
				"Advanced Analytics",
				"Custom Fields",
				"All Integrations",
				"Phone Support",
				"8-hour Response SLA",
			},
			Limits: domain.PlanLimits{
				MaxUsers:       25,
				MaxCustomers:   -1,
				MaxLeads:       -1,
				MaxDeals:       -1,
				MaxStorage:     10 * 1024 * 1024 * 1024, // 10 GB
				MaxAPIRequests: 50000,
				Features: map[string]bool{
					"api_access":     true,
					"custom_fields":  true,
					"reports":        true,
					"integrations":   true,
					"advanced_analytics": true,
				},
			},
			TrialDays: 14,
			IsActive:  true,
			IsPublic:  true,
			Order:     3,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:           uuid.New(),
			Name:         "Enterprise",
			Code:         "enterprise",
			Description:  "Custom solutions for large organizations",
			PriceMonthly: 0, // Custom pricing
			PriceYearly:  0,
			Currency:     domain.CurrencyMYR,
			Features: []string{
				"Unlimited Users",
				"Unlimited Everything",
				"Full API Access",
				"Custom Integrations",
				"Dedicated Account Manager",
				"24/7 Support",
				"4-hour Response SLA",
				"Custom Training",
				"On-premise Option",
			},
			Limits: domain.PlanLimits{
				MaxUsers:       -1,
				MaxCustomers:   -1,
				MaxLeads:       -1,
				MaxDeals:       -1,
				MaxStorage:     -1,
				MaxAPIRequests: -1,
				Features: map[string]bool{
					"api_access":         true,
					"custom_fields":      true,
					"reports":            true,
					"integrations":       true,
					"advanced_analytics": true,
					"custom_branding":    true,
					"sso":                true,
					"audit_logs":         true,
				},
			},
			TrialDays: 30,
			IsActive:  true,
			IsPublic:  true,
			Order:     4,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, plan := range plans {
		s.plans[plan.ID] = plan
	}
}

// ============================================================================
// Plan Operations
// ============================================================================

// GetPlans returns all active public plans.
func (s *BillingService) GetPlans(ctx context.Context) []*domain.Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var plans []*domain.Plan
	for _, plan := range s.plans {
		if plan.IsActive && plan.IsPublic {
			plans = append(plans, plan)
		}
	}

	// Sort by order
	for i := 0; i < len(plans)-1; i++ {
		for j := 0; j < len(plans)-i-1; j++ {
			if plans[j].Order > plans[j+1].Order {
				plans[j], plans[j+1] = plans[j+1], plans[j]
			}
		}
	}

	return plans
}

// GetPlan returns a plan by ID.
func (s *BillingService) GetPlan(ctx context.Context, planID uuid.UUID) (*domain.Plan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plan, ok := s.plans[planID]
	if !ok {
		return nil, ErrPlanNotFound
	}
	return plan, nil
}

// GetPlanByCode returns a plan by code.
func (s *BillingService) GetPlanByCode(ctx context.Context, code string) (*domain.Plan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, plan := range s.plans {
		if plan.Code == code {
			return plan, nil
		}
	}
	return nil, ErrPlanNotFound
}

// ============================================================================
// Customer Operations
// ============================================================================

// CreateCustomer creates a billing customer.
func (s *BillingService) CreateCustomer(ctx context.Context, tenantID uuid.UUID, email, name string, provider domain.PaymentProvider) (*domain.BillingCustomer, error) {
	customer := &domain.BillingCustomer{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Email:     email,
		Name:      name,
		Provider:  provider,
		Currency:  domain.CurrencyMYR,
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create in payment provider if available
	if p, ok := s.providers[provider]; ok {
		externalID, err := p.CreateCustomer(ctx, customer)
		if err != nil {
			return nil, fmt.Errorf("failed to create customer in %s: %w", provider, err)
		}
		customer.ExternalID = externalID
	}

	s.mu.Lock()
	s.customers[customer.ID] = customer
	s.mu.Unlock()

	return customer, nil
}

// GetCustomer returns a billing customer by tenant ID.
func (s *BillingService) GetCustomer(ctx context.Context, tenantID uuid.UUID) (*domain.BillingCustomer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, customer := range s.customers {
		if customer.TenantID == tenantID {
			return customer, nil
		}
	}
	return nil, ErrCustomerNotFound
}

// ============================================================================
// Subscription Operations
// ============================================================================

// CreateSubscription creates a new subscription.
func (s *BillingService) CreateSubscription(ctx context.Context, tenantID, planID uuid.UUID, billingCycle string, couponCode string) (*domain.Subscription, error) {
	plan, err := s.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	subscription := domain.NewSubscription(tenantID, planID, billingCycle, plan.TrialDays)

	// Apply coupon if provided
	if couponCode != "" {
		coupon, err := s.ValidateCoupon(ctx, couponCode)
		if err == nil {
			subscription.Metadata["coupon_code"] = couponCode
			subscription.Metadata["coupon_id"] = coupon.ID.String()
		}
	}

	s.mu.Lock()
	s.subscriptions[subscription.ID] = subscription
	s.mu.Unlock()

	return subscription, nil
}

// GetSubscription returns a subscription by ID.
func (s *BillingService) GetSubscription(ctx context.Context, subscriptionID uuid.UUID) (*domain.Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sub, ok := s.subscriptions[subscriptionID]
	if !ok {
		return nil, ErrSubscriptionNotFound
	}
	return sub, nil
}

// GetSubscriptionByTenant returns a tenant's active subscription.
func (s *BillingService) GetSubscriptionByTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Subscription, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, sub := range s.subscriptions {
		if sub.TenantID == tenantID && sub.Status != domain.SubscriptionCanceled {
			return sub, nil
		}
	}
	return nil, ErrSubscriptionNotFound
}

// UpdateSubscription updates a subscription (upgrade/downgrade).
func (s *BillingService) UpdateSubscription(ctx context.Context, subscriptionID, newPlanID uuid.UUID) (*domain.Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscriptions[subscriptionID]
	if !ok {
		return nil, ErrSubscriptionNotFound
	}

	if _, ok := s.plans[newPlanID]; !ok {
		return nil, ErrPlanNotFound
	}

	sub.PlanID = newPlanID
	sub.UpdatedAt = time.Now()

	return sub, nil
}

// CancelSubscription cancels a subscription.
func (s *BillingService) CancelSubscription(ctx context.Context, subscriptionID uuid.UUID, immediately bool, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscriptions[subscriptionID]
	if !ok {
		return ErrSubscriptionNotFound
	}

	if immediately {
		sub.CancelImmediately(reason)
	} else {
		sub.Cancel(reason)
	}

	return nil
}

// ReactivateSubscription reactivates a canceled subscription.
func (s *BillingService) ReactivateSubscription(ctx context.Context, subscriptionID uuid.UUID) (*domain.Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscriptions[subscriptionID]
	if !ok {
		return nil, ErrSubscriptionNotFound
	}

	if sub.Status == domain.SubscriptionCanceled && time.Now().After(sub.CurrentPeriodEnd) {
		return nil, ErrSubscriptionExpired
	}

	sub.CancelAtPeriodEnd = false
	sub.CanceledAt = nil
	sub.CancellationReason = ""
	sub.Status = domain.SubscriptionActive
	sub.UpdatedAt = time.Now()

	return sub, nil
}

// ============================================================================
// Payment Method Operations
// ============================================================================

// AddPaymentMethod adds a payment method for a customer.
func (s *BillingService) AddPaymentMethod(ctx context.Context, tenantID uuid.UUID, pm *domain.CustomerPaymentMethod) (*domain.CustomerPaymentMethod, error) {
	pm.ID = uuid.New()
	pm.TenantID = tenantID
	pm.CreatedAt = time.Now()
	pm.UpdatedAt = time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	// If this is the first payment method, make it default
	hasExisting := false
	for _, existing := range s.paymentMethods {
		if existing.TenantID == tenantID {
			hasExisting = true
			break
		}
	}
	if !hasExisting {
		pm.IsDefault = true
	}

	s.paymentMethods[pm.ID] = pm
	return pm, nil
}

// GetPaymentMethods returns all payment methods for a tenant.
func (s *BillingService) GetPaymentMethods(ctx context.Context, tenantID uuid.UUID) ([]*domain.CustomerPaymentMethod, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var methods []*domain.CustomerPaymentMethod
	for _, pm := range s.paymentMethods {
		if pm.TenantID == tenantID {
			methods = append(methods, pm)
		}
	}
	return methods, nil
}

// SetDefaultPaymentMethod sets a payment method as default.
func (s *BillingService) SetDefaultPaymentMethod(ctx context.Context, tenantID, paymentMethodID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for _, pm := range s.paymentMethods {
		if pm.TenantID == tenantID {
			pm.IsDefault = pm.ID == paymentMethodID
			if pm.ID == paymentMethodID {
				found = true
			}
		}
	}

	if !found {
		return ErrPaymentMethodNotFound
	}
	return nil
}

// RemovePaymentMethod removes a payment method.
func (s *BillingService) RemovePaymentMethod(ctx context.Context, paymentMethodID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pm, ok := s.paymentMethods[paymentMethodID]
	if !ok {
		return ErrPaymentMethodNotFound
	}

	if pm.IsDefault {
		return ErrCannotRemoveDefault
	}

	delete(s.paymentMethods, paymentMethodID)
	return nil
}

// ============================================================================
// Invoice Operations
// ============================================================================

// CreateInvoice creates a new invoice.
func (s *BillingService) CreateInvoice(ctx context.Context, tenantID, subscriptionID uuid.UUID, items []domain.InvoiceItem) (*domain.Invoice, error) {
	var subtotal int64
	for _, item := range items {
		subtotal += item.Amount
	}

	// Calculate tax (SST 6%)
	tax := subtotal * 6 / 100
	total := subtotal + tax

	invoice := &domain.Invoice{
		ID:             uuid.New(),
		TenantID:       tenantID,
		SubscriptionID: subscriptionID,
		InvoiceNumber:  fmt.Sprintf("INV-%s-%04d", time.Now().Format("200601"), len(s.invoices)+1),
		Status:         "draft",
		Currency:       domain.CurrencyMYR,
		Subtotal:       subtotal,
		Tax:            tax,
		Total:          total,
		AmountDue:      total,
		LineItems:      items,
		DueDate:        time.Now().AddDate(0, 0, 14), // Due in 14 days
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	s.mu.Lock()
	s.invoices[invoice.ID] = invoice
	s.mu.Unlock()

	return invoice, nil
}

// GetInvoice returns an invoice by ID.
func (s *BillingService) GetInvoice(ctx context.Context, invoiceID uuid.UUID) (*domain.Invoice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	invoice, ok := s.invoices[invoiceID]
	if !ok {
		return nil, ErrInvoiceNotFound
	}
	return invoice, nil
}

// GetInvoices returns all invoices for a tenant.
func (s *BillingService) GetInvoices(ctx context.Context, tenantID uuid.UUID) ([]*domain.Invoice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var invoices []*domain.Invoice
	for _, inv := range s.invoices {
		if inv.TenantID == tenantID {
			invoices = append(invoices, inv)
		}
	}
	return invoices, nil
}

// ============================================================================
// Coupon Operations
// ============================================================================

// ValidateCoupon validates a coupon code.
func (s *BillingService) ValidateCoupon(ctx context.Context, code string) (*domain.Coupon, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	coupon, ok := s.coupons[code]
	if !ok {
		return nil, ErrCouponNotFound
	}

	if !coupon.IsActive {
		return nil, ErrCouponInactive
	}

	now := time.Now()
	if now.Before(coupon.ValidFrom) {
		return nil, ErrCouponNotYetValid
	}

	if coupon.ValidUntil != nil && now.After(*coupon.ValidUntil) {
		return nil, ErrCouponExpired
	}

	if coupon.MaxRedemptions > 0 && coupon.TimesRedeemed >= coupon.MaxRedemptions {
		return nil, ErrCouponMaxRedemptions
	}

	return coupon, nil
}

// ============================================================================
// Errors
// ============================================================================

var (
	ErrPlanNotFound          = errors.New("plan not found")
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrSubscriptionNotFound  = errors.New("subscription not found")
	ErrSubscriptionExpired   = errors.New("subscription has expired")
	ErrPaymentMethodNotFound = errors.New("payment method not found")
	ErrCannotRemoveDefault   = errors.New("cannot remove default payment method")
	ErrInvoiceNotFound       = errors.New("invoice not found")
	ErrCouponNotFound        = errors.New("coupon not found")
	ErrCouponInactive        = errors.New("coupon is inactive")
	ErrCouponNotYetValid     = errors.New("coupon is not yet valid")
	ErrCouponExpired         = errors.New("coupon has expired")
	ErrCouponMaxRedemptions  = errors.New("coupon has reached maximum redemptions")
)
