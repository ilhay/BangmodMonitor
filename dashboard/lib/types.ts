export interface Host {
  id: string
  org_id: string
  name: string
  region: string
  created_at: string
}

export interface MetricPoint {
  timestamp: string
  cpu_percent: number
  mem_percent: number
}

export interface ProbeResult {
  timestamp: string
  url: string
  region: string
  status_code: number
  response_ms: number
  is_up: boolean
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

export interface BillingOverview {
  subscription: Subscription | null
  plan: Plan | null
  usage: { hosts: number; host_limit: number; regions: number }
  active_regions: string[]
  available_plans: Plan[]
}

export interface AlertRule {
  id: string
  name: string
  host_id: string
  condition_type: 'cpu_high' | 'memory_high' | 'probe_down' | 'probe_slow'
  threshold: number
  duration_sec: number
  target_url: string
  channel: 'slack' | 'discord' | 'telegram' | 'email'
  enabled: boolean
}

export interface User {
  user_id: string
  org_id: string
  email: string
  role: string
}
