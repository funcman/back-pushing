import { useEffect, useState } from 'react'
import { api, ObjectData } from '../services/api'

function Objects() {
  const [objects, setObjects] = useState<ObjectData[]>([])
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.getObjects()
      .then(setObjects)
      .catch(err => setError(err.message))
  }, [])

  if (error) return <div className="error">Error: {error}</div>
  if (!objects.length) return <div className="loading">Loading...</div>

  return (
    <div className="objects">
      <h2>Objects</h2>
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Type</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {objects.map(obj => (
            <tr key={obj.id}>
              <td>{obj.id}</td>
              <td>{obj.name}</td>
              <td>{obj.type}</td>
              <td>{obj.created}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default Objects
