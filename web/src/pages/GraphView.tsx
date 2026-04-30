import { useEffect, useRef, useState } from 'react'
import * as d3 from 'd3'
import { api, GraphData } from '../services/api'

function GraphView() {
  const svgRef = useRef<SVGSVGElement>(null)
  const [graphData, setGraphData] = useState<GraphData | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.getGraph()
      .then(setGraphData)
      .catch(err => setError(err.message))
  }, [])

  useEffect(() => {
    if (!graphData || !svgRef.current) return

    const svg = d3.select(svgRef.current)
    svg.selectAll('*').remove()

    const width = 800
    const height = 600

    const simulation = d3.forceSimulation(graphData.nodes as d3.SimulationNodeDatum[])
      .force('link', d3.forceLink(graphData.edges).id((d: any) => d.id))
      .force('charge', d3.forceManyBody().strength(-400))
      .force('center', d3.forceCenter(width / 2, height / 2))

    const link = svg.append('g')
      .selectAll('line')
      .data(graphData.edges)
      .join('line')
      .attr('stroke', '#999')
      .attr('stroke-opacity', 0.6)

    const node = svg.append('g')
      .selectAll('circle')
      .data(graphData.nodes)
      .join('circle')
      .attr('r', 10)
      .attr('fill', '#69b3a2')

    const label = svg.append('g')
      .selectAll('text')
      .data(graphData.nodes)
      .join('text')
      .text((d: any) => d.label)
      .attr('font-size', 10)
      .attr('text-anchor', 'middle')
      .attr('dy', 20)

    simulation.on('tick', () => {
      link
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y)

      node
        .attr('cx', (d: any) => d.x)
        .attr('cy', (d: any) => d.y)

      label
        .attr('x', (d: any) => d.x)
        .attr('y', (d: any) => d.y)
    })
  }, [graphData])

  if (error) return <div className="error">Error: {error}</div>
  if (!graphData) return <div className="loading">Loading...</div>

  return (
    <div className="graph-view">
      <h2>Graph View</h2>
      <svg ref={svgRef} width="800" height="600" />
    </div>
  )
}

export default GraphView
