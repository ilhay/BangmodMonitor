'use client'

const BASE = '/api/v1'

function getToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('bmm_token')
}

export function saveToken(token: string) {
  localStorage.setItem('bmm_token', token)
}

export function clearToken() {
  localStorage.removeItem('bmm_token')
}

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body != null ? JSON.stringify(body) : undefined,
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? `HTTP ${res.status}`)
  return data as T
}

// Auth
export const login = (email: string, password: string) =>
  req<{ token: string; email: string; org_id: string }>('POST', '/auth/login', { email, password })

export const register = (org_name: string, email: string, password: string) =>
  req<{ token: string; email: string; org_id: string }>('POST', '/auth/register', { org_name, email, password })

export const getMe = () => req<{ email: string; org_id: string; role: string }>('GET', '/me')

// Hosts
export const listHosts = () => req<{ hosts: import('./types').Host[] }>('GET', '/hosts')

export const createHost = (name: string, region: string) =>
  req<{ host: import('./types').Host; token: string; install_linux: string; install_windows: string }>(
    'POST', '/hosts', { name, region }
  )

export const deleteHost = (id: string) => req<{ status: string }>('DELETE', `/hosts/${id}`)

export const rotateToken = (id: string) =>
  req<{ token: string; note: string }>('POST', `/hosts/${id}/rotate`)

export const getMetrics = (id: string, limit = 60) =>
  req<{ host_id: string; points: import('./types').MetricPoint[] }>('GET', `/hosts/${id}/metrics?limit=${limit}`)

// Billing
export const getBillingOverview = () => req<import('./types').BillingOverview>('GET', '/billing/overview')
export const getRegions = () => req<{ regions: { region: string; enabled: boolean }[] }>('GET', '/billing/regions')
export const addRegion = (r: string) => req<{ status: string }>('POST', `/billing/regions/${r}`)
export const removeRegion = (r: string) => req<{ status: string }>('DELETE', `/billing/regions/${r}`)
export const getInvoices = () => req<{ invoices: import('./types').Invoice[] }>('GET', '/billing/invoices')
export const createCheckout = (plan_id: string) =>
  req<{ checkout_url: string }>('POST', '/billing/checkout', { plan_id })

// Probe
export const getProbeResults = (url?: string, region?: string) => {
  const q = new URLSearchParams()
  if (url) q.set('url', url)
  if (region) q.set('region', region)
  return req<{ results: import('./types').ProbeResult[] }>('GET', `/probe/results?${q}`)
}
