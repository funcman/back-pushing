import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Objects from './pages/Objects'
import GraphView from './pages/GraphView'
import Alerts from './pages/Alerts'

function App() {
  return (
    <BrowserRouter>
      <div className="app">
        <nav className="sidebar">
          <h1>Back-Pushing</h1>
          <ul>
            <li><Link to="/">Dashboard</Link></li>
            <li><Link to="/objects">Objects</Link></li>
            <li><Link to="/graph">Graph</Link></li>
            <li><Link to="/alerts">Alerts</Link></li>
          </ul>
        </nav>
        <main className="content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/objects" element={<Objects />} />
            <Route path="/graph" element={<GraphView />} />
            <Route path="/alerts" element={<Alerts />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  )
}

export default App
