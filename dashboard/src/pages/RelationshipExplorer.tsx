import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import CytoscapeComponent from 'react-cytoscapejs'
import { api } from '../services/api'
import { Relationship, Project } from '../types'
import RelationshipList from '../components/RelationshipList'
import TypeFilter from '../components/TypeFilter'
import ErrorState from '../components/ErrorState'

export default function RelationshipExplorer() {
  const navigate = useNavigate()
  const [relationships, setRelationships] = useState<Relationship[]>([])
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [viewMode, setViewMode] = useState<'graph' | 'list'>('graph')
  const [filterType, setFilterType] = useState<string>('all')

  useEffect(() => {
    async function loadData() {
      try {
        setLoading(true)
        const [rels, projs] = await Promise.all([
          api.getRelationships(),
          api.getProjects(),
        ])
        setRelationships(rels)
        setProjects(projs)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    loadData()
  }, [])

  const filteredRelationships = useCallback(() => {
    if (filterType === 'all') return relationships
    return relationships.filter(rel => rel.type.toLowerCase() === filterType.toLowerCase())
  }, [relationships, filterType])

  const prepareGraphData = useCallback(() => {
    const nodes = projects.map(project => ({
      data: {
        id: project.id,
        label: project.name,
        weight: 1,
      },
    }))

    const edges = filteredRelationships().map(rel => ({
      data: {
        source: rel.source_project,
        target: rel.target_project,
        label: rel.type,
        weight: rel.confidence || 1,
      },
    }))

    return { elements: [...nodes, ...edges] }
  }, [projects, filteredRelationships])

  const handleNodeClick = (event: any) => {
    const projectId = event.target.data().id
    navigate(`/projects/${projectId}`)
  }

  if (loading) {
    return (
      <div className="text-center py-16">
        <div className="inline-block relative">
          <div className="w-16 h-16 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
          <div className="absolute inset-0 w-16 h-16 border-4 border-purple-200 border-r-purple-600 rounded-full animate-spin" style={{ animationDuration: '1.5s' }}></div>
        </div>
        <p className="mt-6 text-slate-600 dark:text-slate-400 font-medium">Loading relationships...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <ErrorState error={error} onRetry={() => window.location.reload()} />
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">
            Relationship Explorer
          </h1>
          <p className="text-slate-600 dark:text-slate-400 text-sm mt-1">
            {relationships.length} connections between projects
          </p>
        </div>
        <div className="flex flex-col md:flex-row gap-4">
          <TypeFilter
            value={filterType}
            onChange={setFilterType}
          />
          <button
            onClick={() => setViewMode(viewMode === 'graph' ? 'list' : 'graph')}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            aria-label={`Switch to ${viewMode === 'graph' ? 'list' : 'graph'} view`}
          >
            {viewMode === 'graph' ? '📋 View as List' : '🕸️ View as Graph'}
          </button>
        </div>
      </div>

      {viewMode === 'graph' ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6" style={{ height: '600px' }}>
          <CytoscapeComponent
            elements={prepareGraphData().elements}
            style={{ width: '100%', height: '100%' }}
            layout={{
              name: 'cose',
              animate: true,
              animationDuration: 500,
              nodeRepulsion: 8000,
              idealEdgeLength: 100,
              edgeElasticity: 100,
              nestingFactor: 5,
            }}
            stylesheet={[
              {
                selector: 'node',
                style: {
                  'background-color': '#3B82F6',
                  'label': 'data(label)',
                  'width': '30px',
                  'height': '30px',
                  'font-size': '12px',
                  'color': '#FFFFFF',
                  'text-valign': 'center',
                  'text-halign': 'center',
                },
              },
              {
                selector: 'edge',
                style: {
                  'width': 2,
                  'line-color': '#94A3B8',
                  'target-arrow-color': '#94A3B8',
                  'target-arrow-shape': 'triangle',
                  'curve-style': 'bezier',
                  'label': 'data(label)',
                  'font-size': '10px',
                  'color': '#64748B',
                  'text-rotation': 'autorotate',
                  'text-margin-y': -10,
                },
              },
            ]}
            cy={(cy) => {
              cy.on('tap', 'node', handleNodeClick)
            }}
          />
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <RelationshipList relationships={filteredRelationships()} />
        </div>
      )}
    </div>
  )
}