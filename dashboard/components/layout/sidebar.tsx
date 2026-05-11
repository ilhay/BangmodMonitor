'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { LayoutGrid, CreditCard, Bell, Globe, LogOut, Activity } from 'lucide-react'
import { cn } from '@/lib/utils'
import { clearToken } from '@/lib/api'

const nav = [
  { href: '/hosts',   label: 'Hosts',    icon: LayoutGrid },
  { href: '/alerts',  label: 'Alerts',   icon: Bell },
  { href: '/billing', label: 'Billing',  icon: CreditCard },
]

export function Sidebar() {
  const pathname = usePathname()
  const router = useRouter()

  function signOut() {
    clearToken()
    router.push('/login')
  }

  return (
    <aside className="fixed inset-y-0 left-0 z-50 flex w-[220px] flex-col border-r border-border bg-bg-sidebar">
      {/* Logo */}
      <div className="flex h-16 items-center gap-3 border-b border-border px-5">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent/10 ring-1 ring-accent/30">
          <Activity className="h-4 w-4 text-accent" />
        </div>
        <div>
          <div className="text-sm font-bold tracking-tight text-slate-100">BangmodMonitor</div>
          <div className="text-[10px] text-slate-600">Multi-region monitoring</div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3 py-4 space-y-1">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = pathname.startsWith(href)
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all',
                active
                  ? 'bg-accent/10 text-accent ring-1 ring-accent/20'
                  : 'text-slate-500 hover:bg-bg-hover hover:text-slate-300'
              )}
            >
              <Icon className={cn('h-4 w-4 flex-shrink-0', active ? 'text-accent' : 'text-slate-600')} />
              {label}
              {label === 'Alerts' && (
                <span className="ml-auto flex h-5 w-5 items-center justify-center rounded-full bg-danger/15 text-[10px] font-bold text-danger">
                  2
                </span>
              )}
            </Link>
          )
        })}
      </nav>

      {/* Status */}
      <div className="border-t border-border px-4 py-3">
        <div className="mb-3 flex items-center gap-2">
          <div className="h-2 w-2 rounded-full bg-success animate-pulse-dot" />
          <span className="text-xs text-slate-500">3 hosts online</span>
        </div>
        <div className="flex items-center gap-2">
          <Globe className="h-3.5 w-3.5 text-slate-600" />
          <span className="text-xs text-slate-500">TH · SG · HK</span>
        </div>
      </div>

      {/* Sign out */}
      <div className="border-t border-border p-3">
        <button
          onClick={signOut}
          className="flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-sm text-slate-600 hover:bg-bg-hover hover:text-slate-400 transition-all"
        >
          <LogOut className="h-4 w-4" />
          Sign out
        </button>
      </div>
    </aside>
  )
}
