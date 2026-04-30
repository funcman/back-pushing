import { useEffect, useState } from 'react'
import { api, Alert } from '../services/api'

function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.getAlerts()
      .then(setAlerts)
      .catch(err => setError(err.message))
  }, [])

  if (error) return <div className="error">Error: {error}</div>
  if (!alerts.length) return <div className="loading">Loading...</div>

  return (
    <div className="alerts">
      <h2>Alerts</h2>
      <ul className="alert-list">
        {alerts.map(alert => (
          <li key={alert.id} className={`alert alert-${alert.severity}`}>
            <span className="timestamp">{alert.timestamp}</span>
            <span className="message">{alert.message}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

export default Alerts
