import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { listHosts, createHost, deleteHost, type Host, type CreateHostResponse } from '../api/hosts'
import { clearToken } from '../api/auth'

const REGIONS = ['th', 'sg', 'hk', 'jp', 'tw', 'de', 'fr', 'uk', 'us-east', 'us-west']

export default function Hosts() {
  const nav = useNavigate()
  const [hosts, setHosts] = useState<Host[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showAdd, setShowAdd] = useState(false)
  const [newName, setNewName] = useState('')
  const [newRegion, setNewRegion] = useState('th')
  const [creating, setCreating] = useState(false)
  const [created, setCreated] = useState<CreateHostResponse | null>(null)

  const load = async () => {
    try {
      const res = await listHosts()
      setHosts(res.hosts)
    } catch {
      clearToken(); nav('/login')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setCreating(true)
    try {
      const res = await createHost(newName, newRegion)
      setCreated(res)
      setNewName('')
      load()
    } catch (err) {
      setError(String(err))
    } finally {
      setCreating(false)
    }
  }

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete host "${name}"? This cannot be undone.`)) return
    try {
      await deleteHost(id)
      load()
    } catch (err) {
      setError(String(err))
    }
  }

  return (
    <div className="container">
      <div className="header">
        <h1>BangmodMonitor</h1>
        <Link to="/billing" style={{ marginLeft: 'auto', background: 'none', border: '1px solid #334155', borderRadius: 6, color: '#94a3b8', padding: '6px 12px', textDecoration: 'none', fontSize: '0.875rem' }}>
          Billing
        </Link>
        <button onClick={() => { clearToken(); nav('/login') }} style={{ marginLeft: 8, background: 'none', border: '1px solid #334155', borderRadius: 6, color: '#94a3b8', padding: '6px 12px', cursor: 'pointer' }}>
          Sign out
        </button>
      </div>

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h2 style={{ fontSize: '1.125rem', fontWeight: 600 }}>Hosts</h2>
        <button onClick={() => { setShowAdd(true); setCreated(null) }} style={{ background: '#38bdf8', color: '#0f172a', border: 'none', borderRadius: 6, padding: '8px 16px', fontWeight: 600, cursor: 'pointer' }}>
          + Add Host
        </button>
      </div>

      {error && <p style={{ color: '#f87171', marginBottom: 16 }}>{error}</p>}
      {loading && <p style={{ color: '#94a3b8' }}>Loading…</p>}

      {/* Add host modal */}
      {showAdd && (
        <div style={{ background: '#1e293b', border: '1px solid #334155', borderRadius: 8, padding: 24, marginBottom: 24 }}>
          {!created ? (
            <>
              <h3 style={{ marginBottom: 16 }}>Add New Host</h3>
              <form onSubmit={handleCreate} style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', color: '#94a3b8', marginBottom: 4 }}>Host name</label>
                  <input value={newName} onChange={e => setNewName(e.target.value)} required placeholder="web-server-01"
                    style={{ background: '#0f1117', border: '1px solid #334155', borderRadius: 6, padding: '8px 12px', color: '#e2e8f0', width: 200 }} />
                </div>
                <div>
                  <label style={{ display: 'block', fontSize: '0.75rem', color: '#94a3b8', marginBottom: 4 }}>Region</label>
                  <select value={newRegion} onChange={e => setNewRegion(e.target.value)}
                    style={{ background: '#0f1117', border: '1px solid #334155', borderRadius: 6, padding: '8px 12px', color: '#e2e8f0' }}>
                    {REGIONS.map(r => <option key={r} value={r}>{r.toUpperCase()}</option>)}
                  </select>
                </div>
                <button type="submit" disabled={creating} style={{ background: '#38bdf8', color: '#0f172a', border: 'none', borderRadius: 6, padding: '8px 16px', fontWeight: 600, cursor: 'pointer' }}>
                  {creating ? 'Creating…' : 'Create'}
                </button>
                <button type="button" onClick={() => setShowAdd(false)} style={{ background: 'none', border: '1px solid #334155', borderRadius: 6, padding: '8px 16px', color: '#94a3b8', cursor: 'pointer' }}>
                  Cancel
                </button>
              </form>
            </>
          ) : (
            <>
              <h3 style={{ marginBottom: 8, color: '#4ade80' }}>Host created!</h3>
              <p style={{ color: '#fbbf24', marginBottom: 12, fontSize: '0.875rem' }}>Save this token now — it will NOT be shown again.</p>
              <code style={{ display: 'block', background: '#0f1117', padding: 12, borderRadius: 6, wordBreak: 'break-all', color: '#38bdf8', marginBottom: 16 }}>
                {created.token}
              </code>
              <p style={{ fontSize: '0.75rem', color: '#94a3b8', marginBottom: 8 }}>Linux install:</p>
              <code style={{ display: 'block', background: '#0f1117', padding: 8, borderRadius: 6, fontSize: '0.75rem', color: '#e2e8f0', marginBottom: 16 }}>
                {created.install_linux}
              </code>
              <p style={{ fontSize: '0.75rem', color: '#94a3b8', marginBottom: 8 }}>Windows install:</p>
              <code style={{ display: 'block', background: '#0f1117', padding: 8, borderRadius: 6, fontSize: '0.75rem', color: '#e2e8f0', marginBottom: 16 }}>
                {created.install_windows}
              </code>
              <button onClick={() => { setShowAdd(false); setCreated(null) }} style={{ background: '#38bdf8', color: '#0f172a', border: 'none', borderRadius: 6, padding: '8px 16px', fontWeight: 600, cursor: 'pointer' }}>
                Done
              </button>
            </>
          )}
        </div>
      )}

      {/* Host list */}
      {hosts.length === 0 && !loading ? (
        <p style={{ color: '#475569', textAlign: 'center', marginTop: 64 }}>No hosts yet. Add your first host above.</p>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {hosts.map(h => (
            <div key={h.id} style={{ background: '#1e293b', border: '1px solid #334155', borderRadius: 8, padding: '16px 20px', display: 'flex', alignItems: 'center', gap: 16 }}>
              <div style={{ flex: 1 }}>
                <Link to={`/hosts/${h.id}`} style={{ color: '#e2e8f0', fontWeight: 600, textDecoration: 'none', fontSize: '1rem' }}>
                  {h.name}
                </Link>
                <div style={{ fontSize: '0.75rem', color: '#94a3b8', marginTop: 2 }}>
                  Region: <span style={{ color: '#38bdf8' }}>{h.region.toUpperCase()}</span>
                  &nbsp;·&nbsp;ID: {h.id.substring(0, 8)}…
                </div>
              </div>
              <Link to={`/hosts/${h.id}`} style={{ background: '#0f1117', border: '1px solid #334155', borderRadius: 6, padding: '6px 12px', color: '#94a3b8', textDecoration: 'none', fontSize: '0.75rem' }}>
                View metrics
              </Link>
              <button onClick={() => handleDelete(h.id, h.name)} style={{ background: 'none', border: '1px solid #7f1d1d', borderRadius: 6, padding: '6px 12px', color: '#f87171', cursor: 'pointer', fontSize: '0.75rem' }}>
                Delete
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
