'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Activity, ArrowRight, Loader2 } from 'lucide-react'
import { register, saveToken } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export default function RegisterPage() {
  const router = useRouter()
  const [orgName, setOrgName] = useState('')
  const [email, setEmail]     = useState('')
  const [password, setPassword] = useState('')
  const [error, setError]     = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await register(orgName, email, password)
      saveToken(res.token)
      router.push('/hosts')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="w-full max-w-sm animate-fade-in">
      <div className="mb-8 text-center">
        <div className="inline-flex h-12 w-12 items-center justify-center rounded-xl bg-accent/10 ring-1 ring-accent/30 mb-4">
          <Activity className="h-6 w-6 text-accent" />
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-100">Create account</h1>
        <p className="mt-1 text-sm text-slate-500">Start monitoring in minutes — free</p>
      </div>

      <div className="rounded-2xl border border-border bg-bg-card p-6 shadow-[0_0_40px_rgba(0,0,0,0.4)]">
        <form onSubmit={submit} className="space-y-4">
          <Input label="Organization name" placeholder="Acme Corp" value={orgName} onChange={e => setOrgName(e.target.value)} required autoFocus />
          <Input label="Email" type="email" placeholder="you@company.com" value={email} onChange={e => setEmail(e.target.value)} required />
          <Input label="Password" type="password" placeholder="min 8 characters" value={password} onChange={e => setPassword(e.target.value)} required minLength={8} />
          {error && <div className="rounded-lg border border-danger/20 bg-danger/5 px-3.5 py-2.5 text-xs text-danger">{error}</div>}
          <Button variant="primary" size="lg" type="submit" disabled={loading} className="w-full mt-2">
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <>Create Account <ArrowRight className="h-4 w-4" /></>}
          </Button>
        </form>
      </div>
      <p className="mt-5 text-center text-sm text-slate-600">
        Already have an account?{' '}
        <Link href="/login" className="font-medium text-accent hover:text-accent-hover transition-colors">Sign in →</Link>
      </p>
    </div>
  )
}
