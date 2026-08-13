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
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Relationship Explorer
          </h1>
        </div>
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">Loading relationships...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Relationship Explorer
          </h1>
        </div>
        <ErrorState error={error} onRetry={() => window.location.reload()} />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Relationship Explorer
        </h1>
        <div className="flex flex-col md:flex-row gap-4">
          <TypeFilter
            value={filterType}
            onChange={setFilterType}
          />
          <button
            onClick={() => setViewMode(viewMode === 'graph' ? 'list' : 'graph')}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
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