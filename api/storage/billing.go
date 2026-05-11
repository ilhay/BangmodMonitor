package storage

import (
	"context"
	"database/sql"
	"time"
)

// ── Plans ─────────────────────────────────────────────────────────────────────

type Plan struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	BasePriceCents    int    `json:"base_price_cents"`
	HostLimit         int    `json:"host_limit"`
	RegionsIncluded   int    `json:"regions_included"`
	RegionPriceCents  int    `json:"region_price_cents"`
	StripePriceID     string `json:"stripe_price_id"`
}

func (m *Maria) GetPlans(ctx context.Context) ([]Plan, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT id, name, description, base_price_cents, host_limit, regions_included, region_price_cents, stripe_price_id
		 FROM plans WHERE is_active = 1 ORDER BY base_price_cents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.BasePriceCents, &p.HostLimit, &p.RegionsIncluded, &p.RegionPriceCents, &p.StripePriceID); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (m *Maria) GetPlan(ctx context.Context, planID string) (*Plan, error) {
	p := &Plan{}
	err := m.db.QueryRowContext(ctx,
		`SELECT id, name, description, base_price_cents, host_limit, regions_included, region_price_cents, stripe_price_id
		 FROM plans WHERE id = ?`, planID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.BasePriceCents, &p.HostLimit, &p.RegionsIncluded, &p.RegionPriceCents, &p.StripePriceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

// ── Subscriptions ─────────────────────────────────────────────────────────────

type Subscription struct {
	ID                   string     `json:"id"`
	OrgID                string     `json:"org_id"`
	PlanID               string     `json:"plan_id"`
	StripeCustomerID     string     `json:"stripe_customer_id"`
	StripeSubscriptionID string     `json:"stripe_subscription_id"`
	Status               string     `json:"status"` // active | past_due | canceled | trialing
	CurrentPeriodEnd     *time.Time `json:"current_period_end"`
	CancelAtPeriodEnd    bool       `json:"cancel_at_period_end"`
}

func (m *Maria) CreateSubscription(ctx context.Context, id, orgID, planID string) error {
	_, err := m.db.ExecContext(ctx,
		`INSERT INTO subscriptions (id, org_id, plan_id, status) VALUES (?, ?, ?, 'active')`,
		id, orgID, planID)
	return err
}

func (m *Maria) GetSubscription(ctx context.Context, orgID string) (*Subscription, error) {
	s := &Subscription{}
	var periodEnd sql.NullTime
	err := m.db.QueryRowContext(ctx,
		`SELECT id, org_id, plan_id, stripe_customer_id, stripe_subscription_id,
		        status, current_period_end, cancel_at_period_end
		 FROM subscriptions WHERE org_id = ?`, orgID,
	).Scan(&s.ID, &s.OrgID, &s.PlanID, &s.StripeCustomerID, &s.StripeSubscriptionID,
		&s.Status, &periodEnd, &s.CancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if periodEnd.Valid {
		s.CurrentPeriodEnd = &periodEnd.Time
	}
	return s, nil
}

func (m *Maria) UpdateSubscriptionStripe(ctx context.Context, orgID, stripeCustomerID, stripeSubID, status string, periodEnd *time.Time, cancelAtEnd bool) error {
	_, err := m.db.ExecContext(ctx,
		`UPDATE subscriptions SET stripe_customer_id=?, stripe_subscription_id=?, status=?,
		 current_period_end=?, cancel_at_period_end=? WHERE org_id=?`,
		stripeCustomerID, stripeSubID, status, periodEnd, cancelAtEnd, orgID)
	return err
}

func (m *Maria) UpdateSubscriptionStatus(ctx context.Context, stripeSubID, status string) error {
	_, err := m.db.ExecContext(ctx,
		`UPDATE subscriptions SET status=? WHERE stripe_subscription_id=?`, status, stripeSubID)
	return err
}

func (m *Maria) UpdateSubscriptionPlan(ctx context.Context, orgID, planID string) error {
	_, err := m.db.ExecContext(ctx,
		`UPDATE subscriptions SET plan_id=? WHERE org_id=?`, planID, orgID)
	return err
}

func (m *Maria) GetSubscriptionByStripeID(ctx context.Context, stripeSubID string) (*Subscription, error) {
	s := &Subscription{}
	var periodEnd sql.NullTime
	err := m.db.QueryRowContext(ctx,
		`SELECT id, org_id, plan_id, stripe_customer_id, stripe_subscription_id,
		        status, current_period_end, cancel_at_period_end
		 FROM subscriptions WHERE stripe_subscription_id = ?`, stripeSubID,
	).Scan(&s.ID, &s.OrgID, &s.PlanID, &s.StripeCustomerID, &s.StripeSubscriptionID,
		&s.Status, &periodEnd, &s.CancelAtPeriodEnd)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if periodEnd.Valid {
		s.CurrentPeriodEnd = &periodEnd.Time
	}
	return s, err
}

// ── Org Regions ───────────────────────────────────────────────────────────────

func (m *Maria) GetOrgRegions(ctx context.Context, orgID string) ([]string, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT region FROM org_regions WHERE org_id = ? ORDER BY added_at`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var regions []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, rows.Err()
}

func (m *Maria) AddOrgRegion(ctx context.Context, id, orgID, region string) error {
	_, err := m.db.ExecContext(ctx,
		`INSERT IGNORE INTO org_regions (id, org_id, region) VALUES (?, ?, ?)`, id, orgID, region)
	return err
}

func (m *Maria) RemoveOrgRegion(ctx context.Context, orgID, region string) error {
	_, err := m.db.ExecContext(ctx,
		`DELETE FROM org_regions WHERE org_id = ? AND region = ?`, orgID, region)
	return err
}

func (m *Maria) CountOrgRegions(ctx context.Context, orgID string) (int, error) {
	var count int
	err := m.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM org_regions WHERE org_id = ?`, orgID).Scan(&count)
	return count, err
}

// ── Invoices ──────────────────────────────────────────────────────────────────

type Invoice struct {
	ID              string    `json:"id"`
	OrgID           string    `json:"org_id"`
	StripeInvoiceID string    `json:"stripe_invoice_id"`
	AmountCents     int       `json:"amount_cents"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	CreatedAt       time.Time `json:"created_at"`
}

func (m *Maria) UpsertInvoice(ctx context.Context, inv Invoice) error {
	_, err := m.db.ExecContext(ctx,
		`INSERT INTO invoices (id, org_id, stripe_invoice_id, amount_cents, currency, status, period_start, period_end)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE status=VALUES(status)`,
		inv.ID, inv.OrgID, inv.StripeInvoiceID, inv.AmountCents, inv.Currency,
		inv.Status, inv.PeriodStart, inv.PeriodEnd)
	return err
}

func (m *Maria) ListInvoices(ctx context.Context, orgID string, limit int) ([]Invoice, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT id, org_id, stripe_invoice_id, amount_cents, currency, status, period_start, period_end, created_at
		 FROM invoices WHERE org_id = ? ORDER BY created_at DESC LIMIT ?`, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invs []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.OrgID, &inv.StripeInvoiceID, &inv.AmountCents, &inv.Currency,
			&inv.Status, &inv.PeriodStart, &inv.PeriodEnd, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	return invs, rows.Err()
}

// ── Usage ─────────────────────────────────────────────────────────────────────

func (m *Maria) CountActiveHosts(ctx context.Context, orgID string) (int, error) {
	var count int
	err := m.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM hosts WHERE org_id = ?`, orgID).Scan(&count)
	return count, err
}

// ── Admin ─────────────────────────────────────────────────────────────────────

type OrgSummary struct {
	OrgID      string     `json:"org_id"`
	OrgName    string     `json:"org_name"`
	PlanName   string     `json:"plan_name"`
	SubStatus  string     `json:"status"`
	HostCount  int        `json:"host_count"`
	RegionCount int       `json:"region_count"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (m *Maria) ListAllOrgs(ctx context.Context) ([]OrgSummary, error) {
	rows, err := m.db.QueryContext(ctx,
		`SELECT o.id, o.name, COALESCE(p.name,'none'), COALESCE(s.status,'none'),
		        COUNT(DISTINCT h.id), COUNT(DISTINCT r.region), o.created_at
		 FROM orgs o
		 LEFT JOIN subscriptions s ON s.org_id = o.id
		 LEFT JOIN plans p ON p.id = s.plan_id
		 LEFT JOIN hosts h ON h.org_id = o.id
		 LEFT JOIN org_regions r ON r.org_id = o.id
		 GROUP BY o.id, o.name, p.name, s.status, o.created_at
		 ORDER BY o.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orgs []OrgSummary
	for rows.Next() {
		var o OrgSummary
		if err := rows.Scan(&o.OrgID, &o.OrgName, &o.PlanName, &o.SubStatus,
			&o.HostCount, &o.RegionCount, &o.CreatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	return orgs, rows.Err()
}

func (m *Maria) GetRevenueStats(ctx context.Context) (totalCents int, paidOrgs int, err error) {
	err = m.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_cents),0), COUNT(DISTINCT org_id)
		 FROM invoices WHERE status = 'paid'`).Scan(&totalCents, &paidOrgs)
	return
}
