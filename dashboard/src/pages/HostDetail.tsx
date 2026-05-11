import { useState, useEffect, useCallback } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { getMetrics, rotateToken } from '../api/hosts'
import MetricsChart from '../components/MetricsChart'

interface MetricPoint {
  timestamp: string
  cpu_percent: number
  mem_percent: number
}

export default function HostDetail() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
  const [points, setPoints] = useState<MetricPoint[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [rotating, setRotating] = useState(false)
  const [newToken, setNewToken] = useState('')

  const load = useCallback(async () => {
    if (!id) return
    try {
      const res = await getMetrics(id, 60)
      setPoints(res.points ?? [])
    } catch {
      nav('/login')
    } finally {
      setLoading(false)
    }
  }, [id, nav])

  useEffect(() => {
    load()
    const interval = setInterval(load, 30_000)
    return () => clearInterval(interval)
  }, [load])

  const handleRotate = async () => {
    if (!id || !confirm('Rotate token? The old token will stop working immediately.')) return
    setRotating(true)
    try {
      const res = await rotateToken(id)
      setNewToken(res.token)
    } catch (err) {
      setError(String(err))
    } finally {
      setRotating(false)
    }
  }

  const latest = points[0]

  return (
    <div className="container">
      <div className="header">
        <Link to="/hosts" style={{ color: '#38bdf8', textDecoration: 'none', fontSize: '0.875rem' }}>← Hosts</Link>
        <h1 style={{ marginLeft: 12 }}>Host: {id?.substring(0, 8)}…</h1>
        <button onClick={handleRotate} disabled={rotating} style={{ marginLeft: 'auto', background: 'none', border: '1px solid #334155', borderRadius: 6, color: '#fbbf24', padding: '6px 12px', cursor: 'pointer', fontSize: '0.75rem' }}>
          {rotating ? 'Rotating…' : 'Rotate Token'}
        </button>
      </div>

      {newToken && (
        <div style={{ background: '#1e293b', border: '1px solid #fbbf24', borderRadius: 8, padding: 16, marginBottom: 24 }}>
          <p style={{ color: '#fbbf24', marginBottom: 8, fontSize: '0.875rem' }}>New token (save now — shown once):</p>
          <code style={{ color: '#38bdf8', wordBreak: 'break-all' }}>{newToken}</code>
        </div>
      )}

      {error && <p style={{ color: '#f87171', marginBottom: 16 }}>{error}</p>}
      {loading && <p style={{ color: '#94a3b8' }}>Loading…</p>}

      {latest && (
        <>
          <div className="stat-grid">
            <div className="stat-card">
              <div className="label">CPU Usage</div>
              <div className="value" style={{ color: latest.cpu_percent > 90 ? '#f87171' : latest.cpu_percent > 70 ? '#fbbf24' : '#4ade80' }}>
                {latest.cpu_percent.toFixed(1)}%
              </div>
            </div>
            <div className="stat-card">
              <div className="label">Memory Usage</div>
              <div className="value" style={{ color: latest.mem_percent > 90 ? '#f87171' : latest.mem_percent > 75 ? '#fbbf24' : '#38bdf8' }}>
                {latest.mem_percent.toFixed(1)}%
              </div>
            </div>
            <div className="stat-card">
              <div className="label">Last Update</div>
              <div style={{ fontSize: '0.875rem', color: '#94a3b8', marginTop: 8 }}>
                {new Date(latest.timestamp).toLocaleTimeString()}
              </div>
            </div>
          </div>

          <div className="chart-card">
            <h2>CPU Usage</h2>
            <MetricsChart points={points} title="CPU %" field="cpu_percent" color="#4ade80" />
          </div>

          <div className="chart-card">
            <h2>Memory Usage</h2>
            <MetricsChart points={points} title="Mem %" field="mem_percent" color="#38bdf8" />
          </div>
        </>
      )}

      {!loading && points.length === 0 && (
        <p style={{ color: '#475569', textAlign: 'center', marginTop: 64 }}>
          No data yet. Make sure the agent is running with this host's token.
        </p>
      )}
    </div>
  )
}
