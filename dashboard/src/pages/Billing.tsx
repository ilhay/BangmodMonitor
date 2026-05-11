import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  getBillingOverview, getRegions, addRegion, removeRegion, getInvoices,
  createCheckout, createPortal,
  type BillingOverview, type RegionInfo, type Invoice,
} from '../api/billing'
import { clearToken } from '../api/auth'

export default function Billing() {
  const nav = useNavigate()
  const [overview, setOverview] = useState<BillingOverview | null>(null)
  const [regions, setRegions] = useState<RegionInfo[]>([])
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [regionLoading, setRegionLoading] = useState<string | null>(null)

  const load = async () => {
    try {
      const [ov, reg, inv] = await Promise.all([getBillingOverview(), getRegions(), getInvoices()])
      setOverview(ov)
      setRegions(reg.regions)
      setInvoices(inv.invoices ?? [])
    } catch {
      clearToken(); nav('/login')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleRegionToggle = async (region: string, enabled: boolean) => {
    setRegionLoading(region)
    try {
      if (enabled) {
        await removeRegion(region)
      } else {
        await addRegion(region)
      }
      await load()
    } catch (err) {
      setError(String(err))
    } finally {
      setRegionLoading(null)
    }
  }

  const handleUpgrade = async (planId: string) => {
    try {
      const res = await createCheckout(planId)
      window.location.href = res.checkout_url
    } catch (err) {
      setError(String(err))
    }
  }

  const handlePortal = async () => {
    try {
      const res = await createPortal()
      window.location.href = res.portal_url
    } catch (err) {
      setError(String(err))
    }
  }

  const cents = (c: number) => `$${(c / 100).toFixed(2)}`
  const statusColor = (s: string) => s === 'active' ? '#4ade80' : s === 'past_due' ? '#f87171' : '#fbbf24'

  if (loading) return <div className="container"><p style={{ color: '#94a3b8', marginTop: 40 }}>Loading…</p></div>

  return (
    <div className="container">
      <div className="header">
        <Link to="/hosts" style={{ color: '#38bdf8', textDecoration: 'none', fontSize: '0.875rem' }}>← Hosts</Link>
        <h1 style={{ marginLeft: 12 }}>Billing</h1>
      </div>

      {error && <p style={{ color: '#f87171', marginBottom: 16 }}>{error}</p>}

      {/* Current Plan */}
      {overview?.plan && overview.subscription && (
        <div className="chart-card" style={{ marginBottom: 16 }}>
          <h2>Current Plan</h2>
          <div style={{ display: 'flex', gap: 24, alignItems: 'center', flexWrap: 'wrap' }}>
            <div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#38bdf8' }}>{overview.plan.name}</div>
              <div style={{ color: '#94a3b8', fontSize: '0.875rem' }}>{overview.plan.description}</div>
            </div>
            <div style={{ color: statusColor(overview.subscription.status), fontWeight: 600 }}>
              {overview.subscription.status.toUpperCase()}
            </div>
            <div style={{ color: '#94a3b8', fontSize: '0.875rem' }}>
              {overview.subscription.current_period_end
                ? `Renews ${new Date(overview.subscription.current_period_end).toLocaleDateString()}`
                : 'Free plan'}
            </div>
            {overview.subscription.stripe_customer_id && (
              <button onClick={handlePortal} style={{ marginLeft: 'auto', background: 'none', border: '1px solid #334155', borderRadius: 6, color: '#94a3b8', padding: '6px 14px', cursor: 'pointer', fontSize: '0.8rem' }}>
                Manage Billing ↗
              </button>
            )}
          </div>

          {/* Usage */}
          <div style={{ display: 'flex', gap: 24, marginTop: 20 }}>
            <div className="stat-card" style={{ flex: 1 }}>
              <div className="label">Hosts</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700 }}>
                <span style={{ color: overview.usage.hosts >= overview.usage.host_limit ? '#f87171' : '#4ade80' }}>
                  {overview.usage.hosts}
                </span>
                <span style={{ color: '#475569', fontSize: '1rem' }}> / {overview.usage.host_limit}</span>
              </div>
            </div>
            <div className="stat-card" style={{ flex: 1 }}>
              <div className="label">Active Regions</div>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: '#38bdf8' }}>
                {overview.usage.regions}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Upgrade Plans */}
      {overview?.available_plans && overview.plan?.id === 'plan-free' && (
        <div className="chart-card" style={{ marginBottom: 16 }}>
          <h2>Upgrade Plan</h2>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginTop: 8 }}>
            {overview.available_plans.filter(p => p.base_price_cents > 0).map(p => (
              <div key={p.id} style={{ background: '#0f1117', border: '1px solid #334155', borderRadius: 8, padding: 16, flex: 1, minWidth: 180 }}>
                <div style={{ fontWeight: 700, fontSize: '1rem', marginBottom: 4 }}>{p.name}</div>
                <div style={{ color: '#38bdf8', fontSize: '1.5rem', fontWeight: 700 }}>{cents(p.base_price_cents)}<span style={{ color: '#64748b', fontSize: '0.75rem' }}>/mo</span></div>
                <div style={{ color: '#94a3b8', fontSize: '0.8rem', marginTop: 4 }}>{p.description}</div>
                <div style={{ color: '#64748b', fontSize: '0.75rem', marginTop: 4 }}>+{cents(p.region_price_cents)}/extra region</div>
                <button onClick={() => handleUpgrade(p.id)} style={{ width: '100%', marginTop: 12, background: '#38bdf8', color: '#0f172a', border: 'none', borderRadius: 6, padding: '8px', fontWeight: 600, cursor: 'pointer', fontSize: '0.8rem' }}>
                  Upgrade
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Regions */}
      <div className="chart-card" style={{ marginBottom: 16 }}>
        <h2>Probe Regions</h2>
        <p style={{ color: '#64748b', fontSize: '0.8rem', marginBottom: 12 }}>
          Enable regions to probe your hosts from multiple locations. Each region adds {overview?.plan ? cents(overview.plan.region_price_cents) : '$10'}/month.
        </p>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
          {regions.map(r => (
            <button
              key={r.region}
              onClick={() => handleRegionToggle(r.region, r.enabled)}
              disabled={regionLoading === r.region}
              style={{
                padding: '6px 14px', borderRadius: 9999, fontSize: '0.8rem', fontWeight: 600,
                cursor: 'pointer', border: 'none',
                background: r.enabled ? '#166534' : '#1e293b',
                color: r.enabled ? '#4ade80' : '#64748b',
                border: r.enabled ? '1px solid #166534' : '1px solid #334155',
                opacity: regionLoading === r.region ? 0.5 : 1,
              } as React.CSSProperties}
            >
              {r.region.toUpperCase()} {r.enabled ? '✓' : '+'}
            </button>
          ))}
        </div>
      </div>

      {/* Invoices */}
      {invoices.length > 0 && (
        <div className="chart-card">
          <h2>Invoice History</h2>
          <table style={{ width: '100%', borderCollapse: 'collapse', marginTop: 8 }}>
            <thead>
              <tr style={{ color: '#64748b', fontSize: '0.75rem', textAlign: 'left' }}>
                <th style={{ padding: '6px 0' }}>Period</th>
                <th>Amount</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {invoices.map(inv => (
                <tr key={inv.id} style={{ borderTop: '1px solid #1e293b', fontSize: '0.875rem' }}>
                  <td style={{ padding: '8px 0', color: '#94a3b8' }}>
                    {new Date(inv.period_start).toLocaleDateString()} – {new Date(inv.period_end).toLocaleDateString()}
                  </td>
                  <td style={{ color: '#e2e8f0' }}>{cents(inv.amount_cents)}</td>
                  <td><span className={`badge ${inv.status === 'paid' ? 'up' : 'down'}`}>{inv.status}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
