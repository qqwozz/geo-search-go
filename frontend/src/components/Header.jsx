import { useState, useEffect } from 'react'
import { health } from '../api'
import './Header.css'

export default function Header() {
  const [status, setStatus] = useState(null)

  useEffect(() => {
    let active = true
    const check = () => health().then(r => active && setStatus(r.status === 'ok' ? 'ok' : 'degraded')).catch(() => active && setStatus('error'))
    check()
    const id = setInterval(check, 30000)
    return () => { active = false; clearInterval(id) }
  }, [])

  return (
    <header className="header">
      <h1 className="header-title">Geo Search</h1>
      <span className="header-subtitle">Умный поиск мест</span>
      <span className={`health-dot health-${status || 'unknown'}`} title={`API: ${status || 'проверка...'}`} />
    </header>
  )
}
