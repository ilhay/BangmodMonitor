import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { register, saveToken } from '../api/auth'

export default function Register() {
  const nav = useNavigate()
  const [orgName, setOrgName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await register(orgName, email, password)
      saveToken(res.token)
      nav('/hosts')
    } catch (err) {
      setError(String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="auth-wrap">
      <div className="auth-card">
        <h1>BangmodMonitor</h1>
        <p className="auth-sub">Create your account</p>
        <form onSubmit={submit}>
          <label>Organization name</label>
          <input value={orgName} onChange={e => setOrgName(e.target.value)} required autoFocus placeholder="Acme Corp" />
          <label>Email</label>
          <input type="email" value={email} onChange={e => setEmail(e.target.value)} required />
          <label>Password <span style={{color:'#64748b',fontSize:'0.75rem'}}>(min 8 chars)</span></label>
          <input type="password" value={password} onChange={e => setPassword(e.target.value)} required minLength={8} />
          {error && <p className="error">{error}</p>}
          <button type="submit" disabled={loading}>{loading ? 'Creating account…' : 'Create Account'}</button>
        </form>
        <p className="auth-link">Already have an account? <Link to="/login">Sign in</Link></p>
      </div>
    </div>
  )
}
