import { useEffect, useState } from 'react'
import { api, SystemStatus } from '../services/api'

function Dashboard() {
  const [status, setStatus] = useState<SystemStatus | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.getStatus()
      .then(setStatus)
      .catch(err => setError(err.message))
  }, [])

  if (error) return <div className="error">Error: {error}</div>
  if (!status) return <div className="loading">Loading...</div>

  return (
    <div className="dashboard">
      <h2>System Dashboard</h2>
      <div className="stats">
        <div className="stat">
          <span className="label">Status</span>
          <span className="value">{status.status}</span>
        </div>
        <div className="stat">
          <span className="label">Uptime</span>
          <span className="value">{status.uptime}s</span>
        </div>
        <div className="stat">
          <span className="label">Version</span>
          <span className="value">{status.version}</span>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
