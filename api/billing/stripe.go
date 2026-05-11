package billing

import (
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/webhook"
)

type StripeService struct {
	secretKey      string
	webhookSecret  string
	regionPriceID  string
	enabled        bool
}

func NewStripe(secretKey, webhookSecret, regionPriceID string) *StripeService {
	svc := &StripeService{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		regionPriceID: regionPriceID,
		enabled:       secretKey != "",
	}
	if svc.enabled {
		stripe.Key = secretKey
	}
	return svc
}

func (s *StripeService) Enabled() bool { return s.enabled }

// CreateCustomer creates a Stripe customer and returns the customer ID.
func (s *StripeService) CreateCustomer(email, orgName string) (string, error) {
	if !s.enabled {
		return "", nil
	}
	c, err := customer.New(&stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(orgName),
	})
	if err != nil {
		return "", fmt.Errorf("stripe create customer: %w", err)
	}
	return c.ID, nil
}

// CreateCheckoutSession creates a Stripe Checkout session for a subscription upgrade.
func (s *StripeService) CreateCheckoutSession(customerID, priceID, successURL, cancelURL string) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("stripe not configured")
	}
	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(customerID),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{"price_id": priceID},
		},
	}
	sess, err := checkoutsession.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe checkout: %w", err)
	}
	return sess.URL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session.
func (s *StripeService) CreatePortalSession(customerID, returnURL string) (string, error) {
	if !s.enabled {
		return "", fmt.Errorf("stripe not configured")
	}
	sess, err := session.New(&stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	})
	if err != nil {
		return "", fmt.Errorf("stripe portal: %w", err)
	}
	return sess.URL, nil
}

// ConstructWebhookEvent verifies and parses a Stripe webhook payload.
func (s *StripeService) ConstructWebhookEvent(payload []byte, sigHeader string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, sigHeader, s.webhookSecret)
}
