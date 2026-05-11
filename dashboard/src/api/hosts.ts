import { getToken } from './auth'

const BASE = '/api/v1'

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(BASE + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? `HTTP ${res.status}`)
  return data
}

export interface Host {
  id: string
  org_id: string
  name: string
  region: string
  created_at: string
}

export interface CreateHostResponse {
  host: Host
  token: string
  install_linux: string
  install_windows: string
}

export const listHosts = () => req<{ hosts: Host[] }>('GET', '/hosts')
export const createHost = (name: string, region: string) =>
  req<CreateHostResponse>('POST', '/hosts', { name, region })
export const deleteHost = (id: string) => req<{ status: string }>('DELETE', `/hosts/${id}`)
export const rotateToken = (id: string) => req<{ token: string; note: string }>('POST', `/hosts/${id}/rotate`)
export const getMetrics = (id: string, limit = 60) =>
  req<{ host_id: string; points: Array<{ timestamp: string; cpu_percent: number; mem_percent: number }> }>(
    'GET', `/hosts/${id}/metrics?limit=${limit}`
  )
