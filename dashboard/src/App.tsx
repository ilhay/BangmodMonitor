import { useState, useEffect, useCallback } from 'react'
import MetricsChart from './components/MetricsChart'
import { fetchMetrics, type MetricPoint } from './api/metrics'

export default function App() {
  const [hostId, setHostId] = useState('')
  const [inputValue, setInputValue] = useState('')
  const [points, setPoints] = useState<MetricPoint[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (id: string) => {
    if (!id) return
    setLoading(true)
    setError('')
    try {
      const data = await fetchMetrics(id, 60)
      setPoints(data.points ?? [])
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!hostId) return
    load(hostId)
    const interval = setInterval(() => load(hostId), 30_000)
    return () => clearInterval(interval)
  }, [hostId, load])

  const latest = points[0]
  const cpuColor = (v: number) => v > 90 ? '#f87171' : v > 70 ? '#fbbf24' : '#4ade80'
  const memColor = (v: number) => v > 90 ? '#f87171' : v > 75 ? '#fbbf24' : '#38bdf8'

  return (
    <div className="container">
      <div className="header">
        <h1>BangmodMonitor</h1>
        <span style={{ color: '#475569', fontSize: '0.75rem' }}>Phase 1 — MVP</span>
      </div>

      <div className="host-input">
        <input
          type="text"
          placeholder="Enter Host ID (UUID from database)"
          value={inputValue}
          onChange={e => setInputValue(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && setHostId(inputValue.trim())}
        />
        <button onClick={() => setHostId(inputValue.trim())}>Load</button>
      </div>

      {error && <p style={{ color: '#f87171', marginBottom: 16 }}>{error}</p>}
      {loading && <p style={{ color: '#94a3b8', marginBottom: 16 }}>Loading…</p>}

      {latest && (
        <>
          <div className="stat-grid">
            <div className="stat-card">
              <div className="label">CPU Usage</div>
              <div className="value" style={{ color: cpuColor(latest.cpu_percent) }}>
                {latest.cpu_percent.toFixed(1)}%
              </div>
            </div>
            <div className="stat-card">
              <div className="label">Memory Usage</div>
              <div className="value" style={{ color: memColor(latest.mem_percent) }}>
                {latest.mem_percent.toFixed(1)}%
              </div>
            </div>
            <div className="stat-card">
              <div className="label">Data Points</div>
              <div className="value ok">{points.length}</div>
            </div>
            <div className="stat-card">
              <div className="label">Last Update</div>
              <div style={{ fontSize: '0.875rem', color: '#94a3b8', marginTop: 8 }}>
                {new Date(latest.timestamp).toLocaleTimeString()}
              </div>
            </div>
          </div>

          <div className="chart-card">
            <h2>CPU Usage (last {points.length} samples)</h2>
            <MetricsChart points={points} title="CPU %" field="cpu_percent" color="#4ade80" />
          </div>

          <div className="chart-card">
            <h2>Memory Usage (last {points.length} samples)</h2>
            <MetricsChart points={points} title="Mem %" field="mem_percent" color="#38bdf8" />
          </div>
        </>
      )}

      {!hostId && !loading && (
        <div style={{ textAlign: 'center', color: '#475569', marginTop: 80 }}>
          <p style={{ fontSize: '1.25rem', marginBottom: 8 }}>Enter a Host ID to view metrics</p>
          <p style={{ fontSize: '0.875rem' }}>
            Install the agent: <code style={{ color: '#38bdf8' }}>AGENT_TOKEN=&lt;token&gt; AGENT_REGION=th ./bangmod-agent</code>
          </p>
        </div>
      )}
    </div>
  )
}
