const API_BASE = 'http://localhost:8080'

export interface SystemStatus {
  status: string
  uptime: number
  version: string
}

export interface Alert {
  id: string
  severity: 'info' | 'warning' | 'error'
  message: string
  timestamp: string
}

export interface GraphNode {
  id: string
  type: string
  label: string
}

export interface GraphEdge {
  source: string
  target: string
}

export interface GraphData {
  nodes: GraphNode[]
  edges: GraphEdge[]
}

export interface ObjectData {
  id: string
  type: string
  name: string
  created: string
}

export const api = {
  async getStatus(): Promise<SystemStatus> {
    const res = await fetch(`${API_BASE}/api/status`)
    if (!res.ok) throw new Error('Failed to fetch status')
    return res.json()
  },

  async getAlerts(): Promise<Alert[]> {
    const res = await fetch(`${API_BASE}/api/alerts`)
    if (!res.ok) throw new Error('Failed to fetch alerts')
    return res.json()
  },

  async getGraph(): Promise<GraphData> {
    const res = await fetch(`${API_BASE}/api/graph`)
    if (!res.ok) throw new Error('Failed to fetch graph')
    return res.json()
  },

  async getObjects(): Promise<ObjectData[]> {
    const res = await fetch(`${API_BASE}/api/objects`)
    if (!res.ok) throw new Error('Failed to fetch objects')
    return res.json()
  }
}
