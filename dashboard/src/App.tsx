import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Login from './pages/Login'
import Register from './pages/Register'
import Hosts from './pages/Hosts'
import HostDetail from './pages/HostDetail'
import Billing from './pages/Billing'
import { getToken } from './api/auth'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  return getToken() ? <>{children}</> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login"     element={<Login />} />
        <Route path="/register"  element={<Register />} />
        <Route path="/hosts"     element={<PrivateRoute><Hosts /></PrivateRoute>} />
        <Route path="/hosts/:id" element={<PrivateRoute><HostDetail /></PrivateRoute>} />
        <Route path="/billing"   element={<PrivateRoute><Billing /></PrivateRoute>} />
        <Route path="*"          element={<Navigate to="/hosts" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
