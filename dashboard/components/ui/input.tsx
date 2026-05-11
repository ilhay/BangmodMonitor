import { cn } from '@/lib/utils'
import { type InputHTMLAttributes } from 'react'

export function Input({ className, label, ...props }: InputHTMLAttributes<HTMLInputElement> & { label?: string }) {
  return (
    <div className="space-y-1.5">
      {label && (
        <label className="block text-xs font-medium uppercase tracking-wider text-slate-500">
          {label}
        </label>
      )}
      <input
        className={cn(
          'w-full rounded-lg border border-border bg-bg-input px-3.5 py-2.5 text-sm text-slate-200',
          'placeholder:text-slate-600 outline-none',
          'focus:border-accent focus:ring-2 focus:ring-accent/20',
          'transition-colors duration-150',
          className
        )}
        {...props}
      />
    </div>
  )
}

export function Select({ className, label, children, ...props }: React.SelectHTMLAttributes<HTMLSelectElement> & { label?: string }) {
  return (
    <div className="space-y-1.5">
      {label && (
        <label className="block text-xs font-medium uppercase tracking-wider text-slate-500">
          {label}
        </label>
      )}
      <select
        className={cn(
          'w-full rounded-lg border border-border bg-bg-input px-3.5 py-2.5 text-sm text-slate-200',
          'outline-none focus:border-accent focus:ring-2 focus:ring-accent/20',
          'transition-colors duration-150 cursor-pointer',
          className
        )}
        {...props}
      >
        {children}
      </select>
    </div>
  )
}
