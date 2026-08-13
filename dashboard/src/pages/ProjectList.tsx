import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../services/api'
import { Project, SearchResult } from '../types'
import SearchBar from '../components/SearchBar'
import FilterDropdown from '../components/FilterDropdown'
import SortControl from '../components/SortControl'
import Pagination from '../components/Pagination'
import ProjectTable from '../components/ProjectTable'
import EmptyState from '../components/EmptyState'
import ErrorState from '../components/ErrorState'

type SortOption = 'name' | 'date' | 'size'

interface FilterState {
  technologies: string[]
  frameworks: string[]
  repositoryType: string
}

interface PaginationState {
  page: number
  pageSize: number
  totalCount: number
  totalPages?: number
}

export default function ProjectList() {
  const navigate = useNavigate()
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Search and filter state
  const [searchQuery, setSearchQuery] = useState('')
  const [filters, setFilters] = useState<FilterState>({
    technologies: [],
    frameworks: [],
    repositoryType: 'all',
  })
  const [sortBy, setSortBy] = useState<SortOption>('name')
  const [pagination, setPagination] = useState<PaginationState>({
    page: 1,
    pageSize: 20,
    totalCount: 0,
  })

  useEffect(() => {
    async function loadProjects() {
      try {
        setLoading(true)

        // Use search endpoint for filtering
        const response = await api.searchProjects({
          query: searchQuery,
          technologies: filters.technologies,
          frameworks: filters.frameworks,
          page: pagination.page,
          pageSize: pagination.pageSize,
        })

        // Convert search results to projects (filter only project results)
        const projectResults = response.results.filter(r => r.type === 'project') as SearchResult[]

        // Fetch full project data for each result
        const projectPromises = projectResults.map(result =>
          api.getProject(result.id)
        )

        const fullProjects = await Promise.all(projectPromises)
        setProjects(fullProjects)
        setPagination(prev => ({
          ...prev,
          totalCount: response.totalCount,
          totalPages: response.totalPages,
        }))
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    loadProjects()
  }, [searchQuery, filters, sortBy, pagination.page, pagination.pageSize])

  const handleProjectClick = (project: Project) => {
    navigate(`/projects/${project.id}`)
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Projects
        </h1>
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">Loading projects...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Projects
        </h1>
        <ErrorState error={error} onRetry={() => window.location.reload()} />
      </div>
    )
  }

  if (projects.length === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Projects
        </h1>
        <EmptyState
          message="No projects found matching your filters"
          onClearFilters={() => setFilters({ technologies: [], frameworks: [], repositoryType: 'all' })}
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
        Projects
      </h1>

      {/* Search and Filters */}
      <div className="flex flex-col md:flex-row gap-4">
        <SearchBar
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder="Search projects..."
        />
        <FilterDropdown
          filters={filters}
          onChange={setFilters}
        />
        <SortControl
          value={sortBy}
          onChange={setSortBy}
        />
      </div>

      {/* Project Table */}
      <ProjectTable
        projects={projects}
        onProjectClick={handleProjectClick}
      />

      {/* Pagination */}
      <Pagination
        pagination={pagination}
        onPageChange={(page) => setPagination(prev => ({ ...prev, page }))}
      />
    </div>
  )
}