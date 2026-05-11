'use client'

import { useState, useEffect, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Link from 'next/link'
import { ArrowLeft, RotateCcw, Cpu, MemoryStick, Clock, Loader2, AlertTriangle } from 'lucide-react'
import dynamic from 'next/dynamic'
import { getMetrics, rotateToken, clearToken } from '@/lib/api'
import type { MetricPoint } from '@/lib/types'
import { fmtPct, statusColor, timeAgo } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'

const MetricChart = dynamic(() => import('@/components/charts/metric-chart').then(m => m.MetricChart), { ssr: false })

function KpiCard({ label, value, color, icon: Icon, sub }: {
  label: string; value: string; color: string; icon: React.ElementType; sub?: string
}) {
  return (
    <Card>
      <CardContent className="py-4">
        <div className="flex items-center justify-between mb-3">
          <p className="text-xs text-slate-500 uppercase tracking-wider">{label}</p>
          <Icon className={`h-4 w-4 ${color}`} />
        </div>
        <p className={`font-mono text-3xl font-bold leading-none ${color}`}>{value}</p>
        {sub && <p className="mt-1.5 text-xs text-slate-600">{sub}</p>}
      </CardContent>
    </Card>
  )
}

export default function HostDetailPage() {
  const { id } = useParams<{ id: string }>()
  const router  = useRouter()
  const [points, setPoints] = useState<MetricPoint[]>([])
  const [loading, setLoading] = useState(true)
  const [rotating, setRotating] = useState(false)
  const [newToken, setNewToken] = useState('')

  const load = useCallback(async () => {
    try {
      const res = await getMetrics(id, 60)
      setPoints(res.points ?? [])
    } catch {
      clearToken(); router.push('/login')
    } finally {
      setLoading(false)
    }
  }, [id, router])

  useEffect(() => {
    load()
    const t = setInterval(load, 30_000)
    return () => clearInterval(t)
  }, [load])

  async function handleRotate() {
    if (!confirm('Rotate token? Old token is revoked immediately.')) return
    setRotating(true)
    try {
      const res = await rotateToken(id)
      setNewToken(res.token)
    } finally {
      setRotating(false)
    }
  }

  const latest = points[0]
  const cpuPct = latest?.cpu_percent ?? 0
  const memPct = latest?.mem_percent ?? 0

  return (
    <div className="p-6 max-w-5xl mx-auto animate-fade-in">
      {/* Breadcrumb */}
      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href="/hosts" className="flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-300 transition-colors">
            <ArrowLeft className="h-3.5 w-3.5" /> Hosts
          </Link>
          <span className="text-slate-700">/</span>
          <span className="text-sm font-medium text-slate-200 font-mono">{id.slice(0,8)}…</span>
          <Badge variant="success">Online</Badge>
        </div>
        <Button variant="warning" size="sm" onClick={handleRotate} disabled={rotating}>
          {rotating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <RotateCcw className="h-3.5 w-3.5" />}
          Rotate Token
        </Button>
      </div>

      {/* Rotated token warning */}
      {newToken && (
        <Card className="mb-6 border-warning/30">
          <CardContent className="py-3">
            <div className="flex items-start gap-2 text-xs text-warning mb-2">
              <AlertTriangle className="h-3.5 w-3.5 shrink-0 mt-0.5" />
              New token — save now, shown once only.
            </div>
            <code className="block break-all font-mono text-xs text-slate-300">{newToken}</code>
          </CardContent>
        </Card>
      )}

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-slate-600" />
        </div>
      ) : (
        <>
          {/* KPI cards */}
          <div className="mb-6 grid grid-cols-3 gap-4">
            <KpiCard
              label="CPU Usage" icon={Cpu}
              value={latest ? fmtPct(cpuPct) : '—'}
              color={latest ? statusColor(cpuPct) : 'text-slate-500'}
              sub={latest ? 'Last reading' : 'No data yet'}
            />
            <KpiCard
              label="Memory Usage" icon={MemoryStick}
              value={latest ? fmtPct(memPct) : '—'}
              color={latest ? statusColor(memPct) : 'text-slate-500'}
              sub={latest ? 'Used / Total' : 'No data yet'}
            />
            <KpiCard
              label="Last Update" icon={Clock}
              value={latest ? timeAgo(latest.timestamp) : '—'}
              color="text-slate-300"
              sub={`${points.length} samples loaded`}
            />
          </div>

          {/* Charts */}
          {points.length === 0 ? (
            <Card>
              <CardContent className="py-16 text-center">
                <p className="text-sm text-slate-500">No metrics yet. Make sure the agent is running.</p>
                <code className="mt-3 block rounded-lg bg-bg-input px-4 py-2 font-mono text-xs text-slate-500">
                  AGENT_TOKEN=your-token go run ./agent
                </code>
              </CardContent>
            </Card>
          ) : (
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>CPU Usage</CardTitle>
                  <span className={`font-mono text-sm font-bold ${statusColor(cpuPct)}`}>
                    {fmtPct(cpuPct)}
                  </span>
                </CardHeader>
                <CardContent>
                  <MetricChart data={points} field="cpu_percent" color="#22c55e" label="CPU %" />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Memory Usage</CardTitle>
                  <span className={`font-mono text-sm font-bold ${statusColor(memPct)}`}>
                    {fmtPct(memPct)}
                  </span>
                </CardHeader>
                <CardContent>
                  <MetricChart data={points} field="mem_percent" color="#f59e0b" label="Memory %" />
                </CardContent>
              </Card>
            </div>
          )}
        </>
      )}
    </div>
  )
}
