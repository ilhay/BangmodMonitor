import { getToken } from './auth'

const BASE = '/api/v1/billing'

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(BASE + path, { method, headers, body: body ? JSON.stringify(body) : undefined })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? `HTTP ${res.status}`)
  return data
}

export interface Plan {
  id: string
  name: string
  description: string
  base_price_cents: number
  host_limit: number
  regions_included: number
  region_price_cents: number
}

export interface Subscription {
  id: string
  plan_id: string
  stripe_customer_id: string
  stripe_subscription_id: string
  status: string
  current_period_end: string | null
  cancel_at_period_end: boolean
}

export interface Invoice {
  id: string
  amount_cents: number
  currency: string
  status: string
  period_start: string
  period_end: string
  created_at: string
}

export interface BillingOverview {
  subscription: Subscription | null
  plan: Plan | null
  usage: { hosts: number; host_limit: number; regions: number }
  active_regions: string[]
  available_plans: Plan[]
}

export interface RegionInfo {
  region: string
  enabled: boolean
}

export const getBillingOverview = () => req<BillingOverview>('GET', '/overview')
export const getRegions = () => req<{ regions: RegionInfo[]; active: string[] }>('GET', '/regions')
export const addRegion = (region: string) => req<{ status: string }>('POST', `/regions/${region}`)
export const removeRegion = (region: string) => req<{ status: string }>('DELETE', `/regions/${region}`)
export const getInvoices = () => req<{ invoices: Invoice[] }>('GET', '/invoices')
export const createCheckout = (planId: string) => req<{ checkout_url: string }>('POST', '/checkout', { plan_id: planId })
export const createPortal = () => req<{ portal_url: string }>('POST', '/portal')
