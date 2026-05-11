import { useEffect, useRef } from 'react'
import uPlot from 'uplot'
import 'uplot/dist/uPlot.min.css'
import type { MetricPoint } from '../api/metrics'

interface Props {
  points: MetricPoint[]
  title: string
  field: 'cpu_percent' | 'mem_percent'
  color: string
}

export default function MetricsChart({ points, title, field, color }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const plotRef = useRef<uPlot | null>(null)

  useEffect(() => {
    if (!containerRef.current || points.length === 0) return

    const sorted = [...points].reverse()
    const xs = sorted.map(p => new Date(p.timestamp).getTime() / 1000)
    const ys = sorted.map(p => p[field])

    const opts: uPlot.Options = {
      width: containerRef.current.clientWidth,
      height: 160,
      series: [
        {},
        {
          label: title,
          stroke: color,
          fill: color + '22',
          width: 2,
        },
      ],
      axes: [
        { stroke: '#475569', ticks: { stroke: '#334155' }, grid: { stroke: '#1e293b' } },
        {
          stroke: '#475569',
          ticks: { stroke: '#334155' },
          grid: { stroke: '#334155' },
          values: (_, vals) => vals.map(v => (v != null ? v.toFixed(1) + '%' : '')),
        },
      ],
      scales: { y: { range: [0, 100] } },
      cursor: { show: false },
      legend: { show: false },
    }

    plotRef.current?.destroy()
    plotRef.current = new uPlot(opts, [xs, ys], containerRef.current)

    return () => { plotRef.current?.destroy() }
  }, [points, field, color, title])

  return <div ref={containerRef} />
}
