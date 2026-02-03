// Package http provides HTTP handlers for the billing service.
package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kilang-desa-murni/crm/internal/billing/domain"
	"github.com/kilang-desa-murni/crm/internal/billing/service"
)

// BillingHandler handles billing HTTP requests.
type BillingHandler struct {
	billingService *service.BillingService
}

// NewBillingHandler creates a new BillingHandler.
func NewBillingHandler(billingService *service.BillingService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
	}
}

// RegisterRoutes registers all billing routes.
func (h *BillingHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/billing", func(r chi.Router) {
		// Public endpoints
		r.Get("/plans", h.HandleListPlans)
		r.Get("/plans/{id}", h.HandleGetPlan)
		r.Post("/coupons/validate", h.HandleValidateCoupon)

		// Customer endpoints (authenticated)
		r.Post("/customers", h.HandleCreateCustomer)
		r.Get("/customers/me", h.HandleGetCustomer)

		// Subscription endpoints
		r.Post("/subscriptions", h.HandleCreateSubscription)
		r.Get("/subscriptions/current", h.HandleGetCurrentSubscription)
		r.Put("/subscriptions/current", h.HandleUpdateSubscription)
		r.Post("/subscriptions/current/cancel", h.HandleCancelSubscription)
		r.Post("/subscriptions/current/reactivate", h.HandleReactivateSubscription)

		// Payment method endpoints
		r.Get("/payment-methods", h.HandleListPaymentMethods)
		r.Post("/payment-methods", h.HandleAddPaymentMethod)
		r.Put("/payment-methods/{id}/default", h.HandleSetDefaultPaymentMethod)
		r.Delete("/payment-methods/{id}", h.HandleRemovePaymentMethod)

		// Invoice endpoints
		r.Get("/invoices", h.HandleListInvoices)
		r.Get("/invoices/{id}", h.HandleGetInvoice)
		r.Get("/invoices/{id}/pdf", h.HandleDownloadInvoice)

		// Checkout endpoints
		r.Post("/checkout/session", h.HandleCreateCheckoutSession)
		r.Get("/checkout/success", h.HandleCheckoutSuccess)
		r.Get("/checkout/cancel", h.HandleCheckoutCancel)

		// Webhook endpoints
		r.Post("/webhooks/stripe", h.HandleStripeWebhook)
		r.Post("/webhooks/toyyibpay", h.HandleToyyibPayWebhook)
		r.Post("/webhooks/billplz", h.HandleBillplzWebhook)

		// Usage endpoints
		r.Get("/usage", h.HandleGetUsage)

		// Admin endpoints
		r.Route("/admin", func(r chi.Router) {
			r.Get("/subscriptions", h.HandleAdminListSubscriptions)
			r.Get("/invoices", h.HandleAdminListInvoices)
			r.Post("/plans", h.HandleAdminCreatePlan)
			r.Put("/plans/{id}", h.HandleAdminUpdatePlan)
			r.Post("/coupons", h.HandleAdminCreateCoupon)
		})
	})
}

// ============================================================================
// Plan Handlers
// ============================================================================

// HandleListPlans returns all available plans.
func (h *BillingHandler) HandleListPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.billingService.GetPlans(r.Context())

	// Format for response
	response := make([]map[string]interface{}, len(plans))
	for i, plan := range plans {
		response[i] = map[string]interface{}{
			"id":            plan.ID,
			"name":          plan.Name,
			"code":          plan.Code,
			"description":   plan.Description,
			"price_monthly": plan.PriceMonthly,
			"price_yearly":  plan.PriceYearly,
			"currency":      plan.Currency,
			"features":      plan.Features,
			"limits":        plan.Limits,
			"trial_days":    plan.TrialDays,
			"order":         plan.Order,
			// Format prices for display
			"formatted_monthly": formatPrice(plan.PriceMonthly, plan.Currency),
			"formatted_yearly":  formatPrice(plan.PriceYearly, plan.Currency),
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"plans": response,
	})
}

// HandleGetPlan returns a specific plan.
func (h *BillingHandler) HandleGetPlan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid plan ID")
		return
	}

	plan, err := h.billingService.GetPlan(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Plan not found")
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

// HandleValidateCoupon validates a coupon code.
func (h *BillingHandler) HandleValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	coupon, err := h.billingService.ValidateCoupon(r.Context(), req.Code)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":           true,
		"coupon":          coupon,
		"discount_amount": coupon.DiscountAmount,
		"discount_type":   coupon.Type,
	})
}

// ============================================================================
// Customer Handlers
// ============================================================================

// HandleCreateCustomer creates a billing customer.
func (h *BillingHandler) HandleCreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID uuid.UUID             `json:"tenant_id"`
		Email    string                `json:"email"`
		Name     string                `json:"name"`
		Provider domain.PaymentProvider `json:"provider"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Provider == "" {
		req.Provider = domain.ProviderStripe
	}

	customer, err := h.billingService.CreateCustomer(r.Context(), req.TenantID, req.Email, req.Name, req.Provider)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, customer)
}

// HandleGetCustomer returns the current tenant's billing customer.
func (h *BillingHandler) HandleGetCustomer(w http.ResponseWriter, r *http.Request) {
	// In production, extract tenant ID from auth context
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	customer, err := h.billingService.GetCustomer(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Billing customer not found")
		return
	}

	writeJSON(w, http.StatusOK, customer)
}

// ============================================================================
// Subscription Handlers
// ============================================================================

// HandleCreateSubscription creates a new subscription.
func (h *BillingHandler) HandleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID     uuid.UUID `json:"tenant_id"`
		PlanID       uuid.UUID `json:"plan_id"`
		BillingCycle string    `json:"billing_cycle"`
		CouponCode   string    `json:"coupon_code,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.BillingCycle == "" {
		req.BillingCycle = "monthly"
	}

	subscription, err := h.billingService.CreateSubscription(r.Context(), req.TenantID, req.PlanID, req.BillingCycle, req.CouponCode)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, subscription)
}

// HandleGetCurrentSubscription returns the tenant's current subscription.
func (h *BillingHandler) HandleGetCurrentSubscription(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	subscription, err := h.billingService.GetSubscriptionByTenant(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	// Get plan details
	plan, _ := h.billingService.GetPlan(r.Context(), subscription.PlanID)

	response := map[string]interface{}{
		"subscription":       subscription,
		"plan":               plan,
		"days_until_renewal": subscription.DaysUntilRenewal(),
		"is_in_trial":        subscription.IsInTrial(),
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleUpdateSubscription updates the subscription (upgrade/downgrade).
func (h *BillingHandler) HandleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var req struct {
		PlanID uuid.UUID `json:"plan_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	subscription, err := h.billingService.GetSubscriptionByTenant(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	updated, err := h.billingService.UpdateSubscription(r.Context(), subscription.ID, req.PlanID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// HandleCancelSubscription cancels the subscription.
func (h *BillingHandler) HandleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var req struct {
		Immediately bool   `json:"immediately"`
		Reason      string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	subscription, err := h.billingService.GetSubscriptionByTenant(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "No active subscription found")
		return
	}

	if err := h.billingService.CancelSubscription(r.Context(), subscription.ID, req.Immediately, req.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription canceled successfully",
	})
}

// HandleReactivateSubscription reactivates a canceled subscription.
func (h *BillingHandler) HandleReactivateSubscription(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	subscription, err := h.billingService.GetSubscriptionByTenant(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "No subscription found")
		return
	}

	reactivated, err := h.billingService.ReactivateSubscription(r.Context(), subscription.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, reactivated)
}

// ============================================================================
// Payment Method Handlers
// ============================================================================

// HandleListPaymentMethods returns all payment methods.
func (h *BillingHandler) HandleListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	methods, err := h.billingService.GetPaymentMethods(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"payment_methods": methods,
	})
}

// HandleAddPaymentMethod adds a new payment method.
func (h *BillingHandler) HandleAddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	var req struct {
		Type         domain.PaymentMethod   `json:"type"`
		Provider     domain.PaymentProvider `json:"provider"`
		CardBrand    string                 `json:"card_brand,omitempty"`
		CardLast4    string                 `json:"card_last4,omitempty"`
		CardExpMonth int                    `json:"card_exp_month,omitempty"`
		CardExpYear  int                    `json:"card_exp_year,omitempty"`
		BankCode     string                 `json:"bank_code,omitempty"`
		BankName     string                 `json:"bank_name,omitempty"`
		IsDefault    bool                   `json:"is_default"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pm := &domain.CustomerPaymentMethod{
		Type:         req.Type,
		Provider:     req.Provider,
		CardBrand:    req.CardBrand,
		CardLast4:    req.CardLast4,
		CardExpMonth: req.CardExpMonth,
		CardExpYear:  req.CardExpYear,
		BankCode:     req.BankCode,
		BankName:     req.BankName,
		IsDefault:    req.IsDefault,
	}

	method, err := h.billingService.AddPaymentMethod(r.Context(), tenantID, pm)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, method)
}

// HandleSetDefaultPaymentMethod sets a payment method as default.
func (h *BillingHandler) HandleSetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	pmIDStr := chi.URLParam(r, "id")
	pmID, err := uuid.Parse(pmIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payment method ID")
		return
	}

	if err := h.billingService.SetDefaultPaymentMethod(r.Context(), tenantID, pmID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Default payment method updated",
	})
}

// HandleRemovePaymentMethod removes a payment method.
func (h *BillingHandler) HandleRemovePaymentMethod(w http.ResponseWriter, r *http.Request) {
	pmIDStr := chi.URLParam(r, "id")
	pmID, err := uuid.Parse(pmIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payment method ID")
		return
	}

	if err := h.billingService.RemovePaymentMethod(r.Context(), pmID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Invoice Handlers
// ============================================================================

// HandleListInvoices returns all invoices for the tenant.
func (h *BillingHandler) HandleListInvoices(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant ID")
		return
	}

	invoices, err := h.billingService.GetInvoices(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"invoices": invoices,
		"total":    len(invoices),
	})
}

// HandleGetInvoice returns a specific invoice.
func (h *BillingHandler) HandleGetInvoice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid invoice ID")
		return
	}

	invoice, err := h.billingService.GetInvoice(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Invoice not found")
		return
	}

	writeJSON(w, http.StatusOK, invoice)
}

// HandleDownloadInvoice returns the invoice as PDF.
func (h *BillingHandler) HandleDownloadInvoice(w http.ResponseWriter, r *http.Request) {
	// In production, generate PDF or return hosted URL
	writeError(w, http.StatusNotImplemented, "PDF generation not implemented")
}

// ============================================================================
// Checkout Handlers
// ============================================================================

// HandleCreateCheckoutSession creates a checkout session.
func (h *BillingHandler) HandleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlanID       uuid.UUID `json:"plan_id"`
		BillingCycle string    `json:"billing_cycle"`
		SuccessURL   string    `json:"success_url"`
		CancelURL    string    `json:"cancel_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// In production, create Stripe checkout session or redirect to payment gateway
	checkoutURL := "/checkout?session=" + uuid.New().String()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"checkout_url": checkoutURL,
		"session_id":   uuid.New().String(),
	})
}

// HandleCheckoutSuccess handles successful checkout.
func (h *BillingHandler) HandleCheckoutSuccess(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"session_id": sessionID,
		"message":    "Payment successful! Your subscription is now active.",
	})
}

// HandleCheckoutCancel handles canceled checkout.
func (h *BillingHandler) HandleCheckoutCancel(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"canceled": true,
		"message":  "Checkout canceled.",
	})
}

// ============================================================================
// Webhook Handlers
// ============================================================================

// HandleStripeWebhook handles Stripe webhook events.
func (h *BillingHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	// In production, verify signature and process events
	writeJSON(w, http.StatusOK, map[string]string{"received": "true"})
}

// HandleToyyibPayWebhook handles ToyyibPay webhook events.
func (h *BillingHandler) HandleToyyibPayWebhook(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"received": "true"})
}

// HandleBillplzWebhook handles Billplz webhook events.
func (h *BillingHandler) HandleBillplzWebhook(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"received": "true"})
}

// ============================================================================
// Usage Handlers
// ============================================================================

// HandleGetUsage returns usage data for the tenant.
func (h *BillingHandler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	// In production, calculate actual usage
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"users":       3,
		"customers":   150,
		"leads":       320,
		"deals":       45,
		"storage_mb":  125,
		"api_calls":   2500,
		"period_start": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
		"period_end":   time.Now().Format(time.RFC3339),
	})
}

// ============================================================================
// Admin Handlers
// ============================================================================

// HandleAdminListSubscriptions returns all subscriptions.
func (h *BillingHandler) HandleAdminListSubscriptions(w http.ResponseWriter, r *http.Request) {
	// In production, return paginated list
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": []interface{}{},
		"total":         0,
	})
}

// HandleAdminListInvoices returns all invoices.
func (h *BillingHandler) HandleAdminListInvoices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"invoices": []interface{}{},
		"total":    0,
	})
}

// HandleAdminCreatePlan creates a new plan.
func (h *BillingHandler) HandleAdminCreatePlan(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not implemented")
}

// HandleAdminUpdatePlan updates a plan.
func (h *BillingHandler) HandleAdminUpdatePlan(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not implemented")
}

// HandleAdminCreateCoupon creates a new coupon.
func (h *BillingHandler) HandleAdminCreateCoupon(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not implemented")
}

// ============================================================================
// Helper Functions
// ============================================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func formatPrice(amount int64, currency domain.Currency) string {
	switch currency {
	case domain.CurrencyMYR:
		return "RM " + strconv.FormatFloat(float64(amount)/100, 'f', 2, 64)
	case domain.CurrencyUSD:
		return "$" + strconv.FormatFloat(float64(amount)/100, 'f', 2, 64)
	case domain.CurrencySGD:
		return "S$" + strconv.FormatFloat(float64(amount)/100, 'f', 2, 64)
	default:
		return strconv.FormatFloat(float64(amount)/100, 'f', 2, 64) + " " + string(currency)
	}
}
