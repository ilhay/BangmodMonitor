'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Plus, Server, Cpu, MemoryStick, Globe, Copy, Check, X, Loader2 } from 'lucide-react'
import { listHosts, createHost, deleteHost, clearToken } from '@/lib/api'
import type { Host } from '@/lib/types'
import { REGIONS } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input, Select } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'

function StatusDot({ online = true }: { online?: boolean }) {
  return (
    <span className={`inline-block h-2 w-2 rounded-full ${online ? 'bg-success animate-pulse-dot' : 'bg-danger'}`} />
  )
}

export default function HostsPage() {
  const router = useRouter()
  const [hosts, setHosts]     = useState<Host[]>([])
  const [loading, setLoading] = useState(true)
  const [showAdd, setShowAdd] = useState(false)
  const [newName, setNewName] = useState('')
  const [newRegion, setNewRegion] = useState('th')
  const [creating, setCreating]  = useState(false)
  const [created, setCreated]    = useState<{ token: string; linux: string; windows: string } | null>(null)
  const [copied, setCopied]      = useState(false)
  const [error, setError]        = useState('')

  async function load() {
    try {
      const res = await listHosts()
      setHosts(res.hosts ?? [])
    } catch {
      clearToken(); router.push('/login')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreating(true)
    setError('')
    try {
      const res = await createHost(newName, newRegion)
      setCreated({ token: res.token, linux: res.install_linux, windows: res.install_windows })
      setNewName('')
      load()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed')
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(`Delete "${name}"? This cannot be undone.`)) return
    try { await deleteHost(id); load() } catch {}
  }

  async function copyToken(text: string) {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="p-6 max-w-5xl mx-auto animate-fade-in">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Hosts</h1>
          <p className="mt-0.5 text-sm text-slate-500">Manage your monitored servers</p>
        </div>
        <Button variant="primary" onClick={() => { setShowAdd(true); setCreated(null) }}>
          <Plus className="h-4 w-4" /> Add Host
        </Button>
      </div>

      {/* Stats row */}
      <div className="mb-6 grid grid-cols-3 gap-4">
        {[
          { label: 'Total Hosts', value: hosts.length, color: 'text-slate-100' },
          { label: 'Online',      value: hosts.length, color: 'text-success' },
          { label: 'Active Alerts', value: 0,          color: 'text-slate-100' },
        ].map(s => (
          <Card key={s.label}>
            <CardContent className="py-4">
              <p className="text-xs text-slate-500 uppercase tracking-wider mb-1">{s.label}</p>
              <p className={`font-mono text-2xl font-bold ${s.color}`}>{s.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Add Host form */}
      {showAdd && (
        <Card className="mb-6 border-accent/20">
          <CardHeader>
            <CardTitle>{created ? '✓ Host Created' : 'New Host'}</CardTitle>
            <button onClick={() => { setShowAdd(false); setCreated(null) }} className="text-slate-600 hover:text-slate-400">
              <X className="h-4 w-4" />
            </button>
          </CardHeader>
          <CardContent>
            {!created ? (
              <form onSubmit={handleCreate} className="flex items-end gap-3 flex-wrap">
                <div className="flex-1 min-w-48">
                  <Input label="Host name" placeholder="prod-web-01" value={newName} onChange={e => setNewName(e.target.value)} required />
                </div>
                <div className="w-48">
                  <Select label="Region" value={newRegion} onChange={e => setNewRegion(e.target.value)}>
                    {REGIONS.map(r => <option key={r} value={r}>{r.toUpperCase()}</option>)}
                  </Select>
                </div>
                <Button variant="primary" type="submit" disabled={creating}>
                  {creating ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Create'}
                </Button>
                {error && <p className="w-full text-xs text-danger">{error}</p>}
              </form>
            ) : (
              <div className="space-y-3">
                <div className="flex items-start gap-2 rounded-lg border border-warning/20 bg-warning/5 p-3 text-xs text-warning">
                  ⚠ Save this token now — it will NOT be shown again.
                </div>
                <div className="rounded-lg border border-border bg-bg-input p-3">
                  <div className="mb-1 text-[10px] uppercase tracking-wider text-accent">Agent Token</div>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 break-all font-mono text-xs text-slate-300">{created.token}</code>
                    <button onClick={() => copyToken(created.token)} className="shrink-0 text-slate-500 hover:text-accent">
                      {copied ? <Check className="h-3.5 w-3.5 text-success" /> : <Copy className="h-3.5 w-3.5" />}
                    </button>
                  </div>
                </div>
                <div>
                  <p className="mb-1 text-xs text-slate-500">Linux:</p>
                  <code className="block rounded-lg bg-bg-input px-3 py-2 font-mono text-[11px] text-slate-400">{created.linux}</code>
                </div>
                <div>
                  <p className="mb-1 text-xs text-slate-500">Windows:</p>
                  <code className="block rounded-lg bg-bg-input px-3 py-2 font-mono text-[11px] text-slate-400">{created.windows}</code>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Host list */}
      <Card>
        <CardHeader>
          <CardTitle>All Hosts</CardTitle>
          <span className="text-xs text-slate-600">{hosts.length} total</span>
        </CardHeader>
        {loading ? (
          <CardContent className="py-12 text-center">
            <Loader2 className="h-6 w-6 animate-spin text-slate-600 mx-auto" />
          </CardContent>
        ) : hosts.length === 0 ? (
          <CardContent className="py-16 text-center">
            <Server className="h-10 w-10 text-slate-700 mx-auto mb-3" />
            <p className="text-sm text-slate-500">No hosts yet. Add your first server above.</p>
          </CardContent>
        ) : (
          <div className="divide-y divide-border">
            {hosts.map(h => (
              <div key={h.id} className="flex items-center gap-4 px-5 py-4 hover:bg-bg-hover/50 transition-colors">
                <StatusDot />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-slate-200 truncate">{h.name}</span>
                    <Badge variant="accent">{h.region.toUpperCase()}</Badge>
                  </div>
                  <p className="mt-0.5 font-mono text-xs text-slate-600 truncate">{h.id}</p>
                </div>
                <div className="hidden sm:flex items-center gap-6 text-xs">
                  <div className="text-right">
                    <div className="font-mono font-semibold text-success">—</div>
                    <div className="text-slate-600 flex items-center gap-1"><Cpu className="h-3 w-3" /> CPU</div>
                  </div>
                  <div className="text-right">
                    <div className="font-mono font-semibold text-slate-400">—</div>
                    <div className="text-slate-600 flex items-center gap-1"><MemoryStick className="h-3 w-3" /> MEM</div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button size="sm" onClick={() => router.push(`/hosts/${h.id}`)}>
                    <Globe className="h-3.5 w-3.5" /> Metrics
                  </Button>
                  <Button size="sm" variant="danger" onClick={() => handleDelete(h.id, h.name)}>
                    Remove
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
