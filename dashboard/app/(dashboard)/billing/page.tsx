'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { CreditCard, Globe, Check, Loader2, ExternalLink } from 'lucide-react'
import { getBillingOverview, addRegion, removeRegion, createCheckout, clearToken } from '@/lib/api'
import type { BillingOverview } from '@/lib/types'
import { REGIONS, cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'

export default function BillingPage() {
  const router = useRouter()
  const [overview, setOverview] = useState<BillingOverview | null>(null)
  const [loading, setLoading]   = useState(true)
  const [regionLoading, setRegionLoading] = useState<string | null>(null)

  async function load() {
    try {
      const res = await getBillingOverview()
      setOverview(res)
    } catch {
      clearToken(); router.push('/login')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  async function toggleRegion(region: string, enabled: boolean) {
    setRegionLoading(region)
    try {
      if (enabled) await removeRegion(region)
      else await addRegion(region)
      await load()
    } finally {
      setRegionLoading(null)
    }
  }

  async function upgrade(planId: string) {
    try {
      const res = await createCheckout(planId)
      window.location.href = res.checkout_url
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Checkout failed')
    }
  }

  const activeSet = new Set(overview?.active_regions ?? [])
  const cents = (c: number) => `$${(c / 100).toFixed(0)}`

  if (loading) return (
    <div className="flex h-64 items-center justify-center">
      <Loader2 className="h-8 w-8 animate-spin text-slate-600" />
    </div>
  )

  return (
    <div className="p-6 max-w-4xl mx-auto animate-fade-in space-y-6">
      <div>
        <h1 className="text-xl font-bold text-slate-100">Billing</h1>
        <p className="mt-0.5 text-sm text-slate-500">Manage your plan and probe regions</p>
      </div>

      {/* Current plan */}
      {overview?.plan && overview.subscription && (
        <Card className="border-accent/20">
          <CardHeader>
            <div className="flex items-center gap-3">
              <CreditCard className="h-4 w-4 text-accent" />
              <CardTitle>Current Plan</CardTitle>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant={overview.subscription.status === 'active' ? 'success' : 'warning'}>
                {overview.subscription.status}
              </Badge>
              {overview.subscription.stripe_customer_id && (
                <Button size="sm" onClick={() => {}}>
                  <ExternalLink className="h-3.5 w-3.5" /> Manage
                </Button>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-1 mb-4">
              <span className="font-mono text-4xl font-bold text-slate-100">
                {cents(overview.plan.base_price_cents)}
              </span>
              <span className="text-slate-500">/month</span>
              <span className="ml-3 text-lg font-semibold text-accent">{overview.plan.name}</span>
            </div>
            <div className="grid grid-cols-2 gap-4">
              {[
                { label: 'Hosts', used: overview.usage.hosts, limit: overview.usage.host_limit },
                { label: 'Regions', used: overview.usage.regions, limit: overview.plan.regions_included },
              ].map(u => (
                <div key={u.label}>
                  <div className="mb-1.5 flex items-center justify-between text-xs">
                    <span className="text-slate-500">{u.label}</span>
                    <span className={`font-mono font-semibold ${u.used >= u.limit ? 'text-danger' : 'text-slate-300'}`}>
                      {u.used} / {u.limit}
                    </span>
                  </div>
                  <div className="h-1.5 rounded-full bg-border overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all ${u.used >= u.limit ? 'bg-danger' : 'bg-accent'}`}
                      style={{ width: `${Math.min(100, (u.used / u.limit) * 100)}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Upgrade plans */}
      {overview?.available_plans?.filter(p => p.base_price_cents > 0).length > 0 && (
        <div>
          <h2 className="mb-3 text-sm font-semibold text-slate-400">Upgrade</h2>
          <div className="grid grid-cols-2 gap-4">
            {overview.available_plans.filter(p => p.base_price_cents > 0).map(plan => (
              <Card key={plan.id} className="hover:border-accent/40 transition-colors">
                <CardContent className="py-5">
                  <p className="text-xs text-slate-500 uppercase tracking-wider mb-1">{plan.name}</p>
                  <div className="flex items-baseline gap-1 mb-3">
                    <span className="font-mono text-2xl font-bold text-slate-100">{cents(plan.base_price_cents)}</span>
                    <span className="text-xs text-slate-500">/mo</span>
                  </div>
                  <ul className="space-y-1.5 mb-4">
                    {[
                      `${plan.host_limit} hosts`,
                      `${plan.regions_included} region${plan.regions_included > 1 ? 's' : ''} included`,
                      `+${cents(plan.region_price_cents)}/extra region`,
                    ].map(f => (
                      <li key={f} className="flex items-center gap-2 text-xs text-slate-400">
                        <Check className="h-3 w-3 text-success shrink-0" /> {f}
                      </li>
                    ))}
                  </ul>
                  <Button variant="primary" size="sm" className="w-full" onClick={() => upgrade(plan.id)}>
                    Upgrade to {plan.name}
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}

      {/* Regions */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Globe className="h-4 w-4 text-slate-500" />
            <CardTitle>Probe Regions</CardTitle>
          </div>
          <span className="text-xs text-slate-600">
            +{overview?.plan ? `$${(overview.plan.region_price_cents / 100).toFixed(0)}` : '$10'}/region/month
          </span>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {REGIONS.map(r => {
              const enabled = activeSet.has(r)
              const busy = regionLoading === r
              return (
                <button
                  key={r}
                  onClick={() => toggleRegion(r, enabled)}
                  disabled={busy}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-semibold uppercase tracking-wide transition-all',
                    enabled
                      ? 'border-success/40 bg-success/10 text-success hover:bg-success/15'
                      : 'border-border text-slate-600 hover:border-accent/40 hover:text-accent hover:bg-accent/5',
                    busy && 'opacity-50 cursor-not-allowed'
                  )}
                >
                  {busy ? <Loader2 className="h-3 w-3 animate-spin" /> : enabled ? <Check className="h-3 w-3" /> : null}
                  {r.toUpperCase()}
                </button>
              )
            })}
          </div>
        </CardContent>
      </Card>

      {/* Invoices placeholder */}
      <Card>
        <CardHeader><CardTitle>Invoice History</CardTitle></CardHeader>
        <CardContent className="py-10 text-center text-sm text-slate-600">
          No invoices yet.
        </CardContent>
      </Card>
    </div>
  )
}
