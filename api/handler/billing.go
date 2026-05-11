package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bangmodmonitor/api/billing"
	"github.com/bangmodmonitor/api/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
)

type BillingHandler struct {
	maria      *storage.Maria
	stripe     *billing.StripeService
	appBaseURL string

	starterPriceID string
	proPriceID     string
}

func NewBilling(maria *storage.Maria, stripe *billing.StripeService, appBaseURL, starterPriceID, proPriceID string) *BillingHandler {
	return &BillingHandler{
		maria:          maria,
		stripe:         stripe,
		appBaseURL:     appBaseURL,
		starterPriceID: starterPriceID,
		proPriceID:     proPriceID,
	}
}

// Overview returns the org's current subscription, plan, usage and regions.
func (h *BillingHandler) Overview(c *gin.Context) {
	orgID := c.GetString("org_id")

	sub, err := h.maria.GetSubscription(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sub == nil {
		c.JSON(http.StatusOK, gin.H{"subscription": nil})
		return
	}

	plan, _ := h.maria.GetPlan(c.Request.Context(), sub.PlanID)
	regions, _ := h.maria.GetOrgRegions(c.Request.Context(), orgID)
	hostCount, _ := h.maria.CountActiveHosts(c.Request.Context(), orgID)
	plans, _ := h.maria.GetPlans(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{
		"subscription": sub,
		"plan":         plan,
		"usage": gin.H{
			"hosts":        hostCount,
			"host_limit":   plan.HostLimit,
			"regions":      len(regions),
		},
		"active_regions": regions,
		"available_plans": plans,
	})
}

// Checkout creates a Stripe Checkout session for upgrading to a paid plan.
func (h *BillingHandler) Checkout(c *gin.Context) {
	orgID := c.GetString("org_id")

	var req struct {
		PlanID string `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	plan, err := h.maria.GetPlan(c.Request.Context(), req.PlanID)
	if err != nil || plan == nil || plan.StripePriceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan or Stripe price not configured"})
		return
	}

	sub, _ := h.maria.GetSubscription(c.Request.Context(), orgID)
	if sub == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no subscription record found"})
		return
	}

	// Ensure org has a Stripe customer
	customerID := sub.StripeCustomerID
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stripe not configured — contact support"})
		return
	}

	url, err := h.stripe.CreateCheckoutSession(
		customerID,
		plan.StripePriceID,
		h.appBaseURL+"/billing?success=1",
		h.appBaseURL+"/billing?canceled=1",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checkout_url": url})
}

// Portal creates a Stripe Customer Portal session.
func (h *BillingHandler) Portal(c *gin.Context) {
	orgID := c.GetString("org_id")

	sub, _ := h.maria.GetSubscription(c.Request.Context(), orgID)
	if sub == nil || sub.StripeCustomerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no Stripe customer linked"})
		return
	}

	url, err := h.stripe.CreatePortalSession(sub.StripeCustomerID, h.appBaseURL+"/billing")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"portal_url": url})
}

// Webhook handles Stripe webhook events.
func (h *BillingHandler) Webhook(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 65536))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body"})
		return
	}

	event, err := h.stripe.ConstructWebhookEvent(body, c.GetHeader("Stripe-Signature"))
	if err != nil {
		log.Printf("webhook sig verify failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	ctx := c.Request.Context()
	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err == nil && sess.Subscription != nil {
			sub, _ := h.maria.GetSubscriptionByStripeID(ctx, "")
			if sub != nil {
				t := time.Unix(sess.Subscription.CurrentPeriodEnd, 0)
				_ = h.maria.UpdateSubscriptionStripe(ctx, sub.OrgID,
					sess.Customer.ID, sess.Subscription.ID,
					string(sess.Subscription.Status), &t, false)
			}
		}

	case "invoice.paid":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err == nil {
			sub, _ := h.maria.GetSubscriptionByStripeID(ctx, inv.Subscription.ID)
			if sub != nil {
				t := time.Unix(inv.PeriodEnd, 0)
				_ = h.maria.UpdateSubscriptionStripe(ctx, sub.OrgID,
					inv.Customer.ID, inv.Subscription.ID, "active", &t, false)
				_ = h.maria.UpsertInvoice(ctx, storage.Invoice{
					ID:              uuid.New().String(),
					OrgID:           sub.OrgID,
					StripeInvoiceID: inv.ID,
					AmountCents:     int(inv.AmountPaid),
					Currency:        string(inv.Currency),
					Status:          "paid",
					PeriodStart:     time.Unix(inv.PeriodStart, 0),
					PeriodEnd:       time.Unix(inv.PeriodEnd, 0),
				})
			}
		}

	case "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err == nil {
			_ = h.maria.UpdateSubscriptionStatus(ctx, inv.Subscription.ID, "past_due")
		}

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err == nil {
			_ = h.maria.UpdateSubscriptionStatus(ctx, sub.ID, "canceled")
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// ── Regions ───────────────────────────────────────────────────────────────────

var availableRegions = []string{"th", "sg", "hk", "jp", "tw", "de", "fr", "uk", "us-east", "us-west", "us-north", "us-south"}

func (h *BillingHandler) GetRegions(c *gin.Context) {
	orgID := c.GetString("org_id")
	active, _ := h.maria.GetOrgRegions(c.Request.Context(), orgID)

	activeSet := map[string]bool{}
	for _, r := range active {
		activeSet[r] = true
	}

	type regionInfo struct {
		Region  string `json:"region"`
		Enabled bool   `json:"enabled"`
	}
	var result []regionInfo
	for _, r := range availableRegions {
		result = append(result, regionInfo{Region: r, Enabled: activeSet[r]})
	}

	c.JSON(http.StatusOK, gin.H{"regions": result, "active": active})
}

func (h *BillingHandler) AddRegion(c *gin.Context) {
	orgID := c.GetString("org_id")
	region := c.Param("region")

	// Validate region
	valid := false
	for _, r := range availableRegions {
		if r == region {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown region"})
		return
	}

	// Check plan limits
	sub, _ := h.maria.GetSubscription(c.Request.Context(), orgID)
	if sub == nil || sub.Status != "active" {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "active subscription required"})
		return
	}
	plan, _ := h.maria.GetPlan(c.Request.Context(), sub.PlanID)
	regionCount, _ := h.maria.CountOrgRegions(c.Request.Context(), orgID)

	if plan != nil && regionCount >= plan.RegionsIncluded+10 {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "region limit reached for your plan"})
		return
	}

	if err := h.maria.AddOrgRegion(c.Request.Context(), uuid.New().String(), orgID, region); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "enabled", "region": region})
}

func (h *BillingHandler) RemoveRegion(c *gin.Context) {
	orgID := c.GetString("org_id")
	region := c.Param("region")

	if err := h.maria.RemoveOrgRegion(c.Request.Context(), orgID, region); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "disabled", "region": region})
}

// Invoices returns billing invoice history.
func (h *BillingHandler) Invoices(c *gin.Context) {
	orgID := c.GetString("org_id")
	invs, err := h.maria.ListInvoices(c.Request.Context(), orgID, 24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if invs == nil {
		invs = []storage.Invoice{}
	}
	c.JSON(http.StatusOK, gin.H{"invoices": invs})
}
