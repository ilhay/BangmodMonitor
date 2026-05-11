import { cn } from '@/lib/utils'

type Variant = 'success' | 'warning' | 'danger' | 'accent' | 'muted'

const variants: Record<Variant, string> = {
  success: 'bg-success/10 text-success border-success/20',
  warning: 'bg-warning/10 text-warning border-warning/20',
  danger:  'bg-danger/10  text-danger  border-danger/20',
  accent:  'bg-accent/10  text-accent  border-accent/20',
  muted:   'bg-bg-hover   text-slate-400 border-border',
}

export function Badge({ variant = 'muted', className, children }: {
  variant?: Variant
  className?: string
  children: React.ReactNode
}) {
  return (
    <span className={cn(
      'inline-flex items-center rounded-full border px-2.5 py-0.5 text-[11px] font-semibold tracking-wide uppercase',
      variants[variant], className
    )}>
      {children}
    </span>
  )
}
