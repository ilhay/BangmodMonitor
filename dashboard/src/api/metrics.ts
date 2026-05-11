export interface MetricPoint {
  timestamp: string
  cpu_percent: number
  mem_percent: number
}

export interface MetricsResponse {
  host_id: string
  points: MetricPoint[]
}

export async function fetchMetrics(hostId: string, limit = 60): Promise<MetricsResponse> {
  const res = await fetch(`/api/v1/metrics/${hostId}?limit=${limit}`)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}
