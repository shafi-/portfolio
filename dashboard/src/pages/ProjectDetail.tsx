import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../services/api'
import { Project, Document, Analysis, Relationship } from '../types'
import IdentityHeader from '../components/IdentityHeader'
import MetadataSection from '../components/MetadataSection'
import DocumentationSection from '../components/DocumentationSection'
import AnalysisSection from '../components/AnalysisSection'
import RelationshipsSection from '../components/RelationshipsSection'
import LoadingSpinner from '../components/LoadingSpinner'

export default function ProjectDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [project, setProject] = useState<Project | null>(null)
  const [documents, setDocuments] = useState<Document[]>([])
  const [analysis, setAnalysis] = useState<Analysis | null>(null)
  const [relationships, setRelationships] = useState<Relationship[]>([])

  const [loading, setLoading] = useState({
    project: true,
    documents: true,
    analysis: true,
    relationships: true,
  })

  const [errors, setErrors] = useState({
    project: undefined as Error | undefined,
    documents: undefined as Error | undefined,
    analysis: undefined as Error | undefined,
    relationships: undefined as Error | undefined,
  })

  useEffect(() => {
    async function loadProject() {
      if (!id) return

      try {
        setLoading(prev => ({ ...prev, project: true, documents: true, analysis: true }))
        const projectData = await api.getProject(id)
        setProject(projectData)

        // Documents and analyses are included in the project response
        setDocuments(projectData.documents || [])
        setAnalysis(projectData.analyses?.[0] || null)

      } catch (err) {
        setErrors(prev => ({ ...prev, project: err as Error }))
      } finally {
        setLoading(prev => ({ ...prev, project: false, documents: false, analysis: false }))
      }
    }

    async function loadRelationships() {
      if (!id) return

      try {
        setLoading(prev => ({ ...prev, relationships: true }))
        const rels = await api.getRelationships(id)
        setRelationships(rels)
      } catch (err) {
        setErrors(prev => ({ ...prev, relationships: err as Error }))
      } finally {
        setLoading(prev => ({ ...prev, relationships: false }))
      }
    }

    loadProject()
    loadRelationships()
  }, [id])

  if (loading.project && !project) {
    return <LoadingSpinner />
  }

  if (errors.project && !project) {
    return (
      <div className="card text-center py-16">
        <div className="w-20 h-20 bg-gradient-to-br from-red-50 to-orange-50 dark:from-red-900/20 dark:to-orange-900/20 rounded-2xl flex items-center justify-center mx-auto mb-6">
          <span className="text-4xl">❌</span>
        </div>
        <h2 className="text-xl font-semibold text-slate-900 dark:text-white mb-2">
          Error loading project
        </h2>
        <p className="text-slate-600 dark:text-slate-400 mb-6 max-w-md mx-auto">{errors.project?.message}</p>
        <button
          onClick={() => navigate('/projects')}
          className="btn btn-primary inline-flex items-center space-x-2"
        >
          <span>←</span>
          <span>Back to Projects</span>
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Back Navigation */}
      <button
        onClick={() => navigate('/projects')}
        className="btn btn-secondary inline-flex items-center space-x-2"
      >
        <span>←</span>
        <span>Back to Projects</span>
      </button>

      {/* Identity Header */}
      <IdentityHeader project={project} />

      {/* Metadata Section */}
      <MetadataSection
        project={project}
        loading={loading.project}
        error={errors.project}
      />

      {/* Documentation Section */}
      <DocumentationSection
        documents={documents}
        loading={loading.documents}
        error={errors.documents}
      />

      {/* Analysis Section (Progressive Enhancement) */}
      <AnalysisSection
        analysis={analysis}
        loading={loading.analysis}
        error={errors.analysis}
      />

      {/* Relationships Section */}
      <RelationshipsSection
        relationships={relationships}
        loading={loading.relationships}
        error={errors.relationships}
        onProjectClick={(projectId) => navigate(`/projects/${projectId}`)}
      />
    </div>
  )
}