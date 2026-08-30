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
      <div className="text-center py-16">
        <div className="inline-block relative">
          <div className="w-16 h-16 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
          <div className="absolute inset-0 w-16 h-16 border-4 border-purple-200 border-r-purple-600 rounded-full animate-spin" style={{ animationDuration: '1.5s' }}></div>
        </div>
        <p className="mt-6 text-slate-600 dark:text-slate-400 font-medium">Loading projects...</p>
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

  if (projects.length === 0) {
    return (
      <div>
        <EmptyState
          message="No projects found matching your filters"
          onClearFilters={() => setFilters({ technologies: [], frameworks: [], repositoryType: 'all' })}
        />
      </div>
    )
  }

  return (
    <div className="space-y-8">
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