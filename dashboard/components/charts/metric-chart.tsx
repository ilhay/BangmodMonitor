'use client'

import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import type { MetricPoint } from '@/lib/types'
import { fmtTime } from '@/lib/utils'

interface Props {
  data: MetricPoint[]
  field: 'cpu_percent' | 'mem_percent'
  color: string
  label: string
}

function CustomTooltip({ active, payload, label }: any) {
  if (!active || !payload?.length) return null
  return (
    <div className="rounded-lg border border-border bg-bg-card px-3 py-2 shadow-xl text-xs">
      <p className="text-slate-500 mb-1">{label}</p>
      <p className="font-mono font-semibold text-slate-200">{payload[0].value.toFixed(1)}%</p>
    </div>
  )
}

export function MetricChart({ data, field, color, label }: Props) {
  const chartData = [...data].reverse().map(p => ({
    time: fmtTime(p.timestamp),
    value: field === 'cpu_percent' ? p.cpu_percent : p.mem_percent,
  }))

  return (
    <div className="h-48 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={chartData} margin={{ top: 4, right: 4, bottom: 0, left: -20 }}>
          <defs>
            <linearGradient id={`grad-${field}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={color} stopOpacity={0.25} />
              <stop offset="100%" stopColor={color} stopOpacity={0.02} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#1a2d4a" vertical={false} />
          <XAxis
            dataKey="time"
            tick={{ fill: '#475569', fontSize: 10, fontFamily: 'var(--font-geist-mono)' }}
            tickLine={false} axisLine={false}
            interval="preserveStartEnd"
          />
          <YAxis
            domain={[0, 100]}
            tick={{ fill: '#475569', fontSize: 10, fontFamily: 'var(--font-geist-mono)' }}
            tickLine={false} axisLine={false}
            tickFormatter={v => `${v}%`}
          />
          <Tooltip content={<CustomTooltip />} />
          <Area
            type="monotone" dataKey="value"
            stroke={color} strokeWidth={2}
            fill={`url(#grad-${field})`}
            dot={false} activeDot={{ r: 4, fill: color, strokeWidth: 0 }}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}
