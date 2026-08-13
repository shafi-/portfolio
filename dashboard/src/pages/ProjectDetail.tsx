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
        setLoading(prev => ({ ...prev, project: true }))
        const projectData = await api.getProject(id)
        setProject(projectData)
      } catch (err) {
        setErrors(prev => ({ ...prev, project: err as Error }))
      } finally {
        setLoading(prev => ({ ...prev, project: false }))
      }
    }

    async function loadDocuments() {
      if (!id) return

      try {
        setLoading(prev => ({ ...prev, documents: true }))
        const docs = await api.getProjectDocuments(id)
        setDocuments(docs)
      } catch (err) {
        setErrors(prev => ({ ...prev, documents: err as Error }))
      } finally {
        setLoading(prev => ({ ...prev, documents: false }))
      }
    }

    async function loadAnalysis() {
      if (!id) return

      try {
        setLoading(prev => ({ ...prev, analysis: true }))
        const analysisData = await api.getProjectAnalysis(id)
        setAnalysis(analysisData)
      } catch (err) {
        // Analysis not found is acceptable (progressive enhancement)
        setAnalysis(null)
      } finally {
        setLoading(prev => ({ ...prev, analysis: false }))
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
    loadDocuments()
    loadAnalysis()
    loadRelationships()
  }, [id])

  if (loading.project && !project) {
    return <LoadingSpinner />
  }

  if (errors.project && !project) {
    return (
      <div className="text-center py-12">
        <h2 className="text-xl font-semibold text-red-600 mb-4">
          Error loading project
        </h2>
        <p className="text-gray-600 mb-4">{errors.project?.message}</p>
        <button
          onClick={() => navigate('/projects')}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Back to Projects
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Back Navigation */}
      <button
        onClick={() => navigate('/projects')}
        className="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
      >
        ← Back to Projects
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