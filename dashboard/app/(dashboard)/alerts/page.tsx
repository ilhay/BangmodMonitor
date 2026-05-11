'use client'

import { useState } from 'react'
import { Bell, Plus, Slack, Send, Mail, MessageSquare, Cpu, MemoryStick, Globe, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'

const MOCK_RULES = [
  { id: '1', name: 'High CPU on web-server-01', condition: 'cpu_high', threshold: 80, duration_sec: 60, channel: 'telegram', enabled: true },
  { id: '2', name: 'google.com Down',           condition: 'probe_down', threshold: 0, duration_sec: 120, channel: 'slack',    enabled: true },
]

const CHANNELS: Record<string, { label: string; icon: React.ElementType }> = {
  slack:    { label: 'Slack',    icon: Slack },
  discord:  { label: 'Discord',  icon: MessageSquare },
  telegram: { label: 'Telegram', icon: Send },
  email:    { label: 'Email',    icon: Mail },
}

const CONDITIONS: Record<string, { label: string; icon: React.ElementType; color: string }> = {
  cpu_high:    { label: 'CPU > threshold%',      icon: Cpu,         color: 'text-warning' },
  memory_high: { label: 'Memory > threshold%',   icon: MemoryStick, color: 'text-warning' },
  probe_down:  { label: 'URL unreachable',        icon: Globe,       color: 'text-danger' },
  probe_slow:  { label: 'Response > threshold ms', icon: Zap,         color: 'text-accent' },
}

export default function AlertsPage() {
  const [rules] = useState(MOCK_RULES)

  return (
    <div className="p-6 max-w-5xl mx-auto animate-fade-in space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Alert Rules</h1>
          <p className="mt-0.5 text-sm text-slate-500">Get notified when something goes wrong</p>
        </div>
        <Button variant="primary">
          <Plus className="h-4 w-4" /> Add Rule
        </Button>
      </div>

      {/* Rules table */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Bell className="h-4 w-4 text-slate-500" />
            <CardTitle>Active Rules</CardTitle>
          </div>
          <Badge variant="success">{rules.filter(r => r.enabled).length} active</Badge>
        </CardHeader>
        <div className="divide-y divide-border">
          {rules.map(rule => {
            const cond = CONDITIONS[rule.condition]
            const chan = CHANNELS[rule.channel]
            const ChanIcon = chan?.icon ?? Bell
            const CondIcon = cond?.icon ?? Bell
            return (
              <div key={rule.id} className="flex items-center gap-4 px-5 py-4 hover:bg-bg-hover/40 transition-colors">
                <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-bg-input border border-border shrink-0">
                  <CondIcon className={`h-4 w-4 ${cond?.color ?? 'text-slate-500'}`} />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-sm text-slate-200 truncate">{rule.name}</p>
                  <div className="mt-0.5 flex items-center gap-2 flex-wrap">
                    <code className={`text-xs font-mono ${cond?.color ?? 'text-slate-500'}`}>
                      {rule.condition}
                      {rule.threshold > 0 ? ` > ${rule.threshold}${rule.condition.includes('slow') ? 'ms' : '%'}` : ''}
                    </code>
                    <span className="text-slate-700">·</span>
                    <span className="text-xs text-slate-600">for {rule.duration_sec}s</span>
                  </div>
                </div>
                <div className="flex items-center gap-1.5 text-xs text-slate-500 shrink-0">
                  <ChanIcon className="h-3.5 w-3.5" />
                  {chan?.label}
                </div>
                <Badge variant={rule.enabled ? 'success' : 'muted'}>
                  {rule.enabled ? 'Active' : 'Paused'}
                </Badge>
                <Button size="sm" variant="ghost">Edit</Button>
              </div>
            )
          })}
        </div>
      </Card>

      {/* Condition types reference */}
      <Card>
        <CardHeader><CardTitle>Condition Reference</CardTitle></CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-3">
            {Object.entries(CONDITIONS).map(([key, { label, icon: Icon, color }]) => (
              <div key={key} className="flex items-center gap-3 rounded-lg border border-border bg-bg-input px-3 py-2.5">
                <Icon className={`h-4 w-4 shrink-0 ${color}`} />
                <div>
                  <code className={`text-xs font-mono font-semibold ${color}`}>{key}</code>
                  <p className="text-xs text-slate-600 mt-0.5">{label}</p>
                </div>
              </div>
            ))}
          </div>
          <div className="mt-4 pt-4 border-t border-border">
            <p className="text-xs text-slate-500 mb-2">Notification channels:</p>
            <div className="flex flex-wrap gap-2">
              {Object.entries(CHANNELS).map(([key, { label, icon: Icon }]) => (
                <div key={key} className="flex items-center gap-1.5 rounded-full border border-border px-3 py-1 text-xs text-slate-400">
                  <Icon className="h-3 w-3" /> {label}
                </div>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
