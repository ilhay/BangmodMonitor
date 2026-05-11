import { cn } from '@/lib/utils'
import { type ButtonHTMLAttributes } from 'react'

type Variant = 'primary' | 'ghost' | 'danger' | 'warning'
type Size    = 'sm' | 'md' | 'lg'

const variants: Record<Variant, string> = {
  primary: 'bg-accent hover:bg-accent-hover text-white shadow-glow-sm hover:shadow-glow',
  ghost:   'border border-border hover:border-border-focus/50 hover:bg-bg-hover text-slate-300',
  danger:  'bg-danger/10 border border-danger/30 text-danger hover:bg-danger/20',
  warning: 'bg-warning/10 border border-warning/30 text-warning hover:bg-warning/20',
}

const sizes: Record<Size, string> = {
  sm: 'h-7 px-3 text-xs gap-1.5',
  md: 'h-9 px-4 text-sm gap-2',
  lg: 'h-11 px-6 text-base gap-2',
}

export function Button({
  variant = 'ghost', size = 'md', className, children, ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant; size?: Size }) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center rounded-lg font-medium transition-all duration-150',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        variants[variant], sizes[size], className
      )}
      {...props}
    >
      {children}
    </button>
  )
}
