const BASE = '/api/v1'

export interface AuthResponse {
  token: string
  user_id: string
  org_id: string
  email: string
}

async function post<T>(path: string, body: unknown, token?: string): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(BASE + path, { method: 'POST', headers, body: JSON.stringify(body) })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? `HTTP ${res.status}`)
  return data
}

export function register(orgName: string, email: string, password: string) {
  return post<AuthResponse>('/auth/register', { org_name: orgName, email, password })
}

export function login(email: string, password: string) {
  return post<AuthResponse>('/auth/login', { email, password })
}

export function getToken(): string | null {
  return localStorage.getItem('bmm_token')
}

export function saveToken(token: string) {
  localStorage.setItem('bmm_token', token)
}

export function clearToken() {
  localStorage.removeItem('bmm_token')
}
