import {
  Project,
  Document,
  Analysis,
  Relationship,
  SearchRequest,
  SearchResponse,
  Statistics,
  Configuration,
  HealthResponse,
} from '../types'
import { parseStatistics, ApiStatisticsResponse } from '../utils/statisticsParser'
import { parseRelationships, ApiRelationshipResponse } from '../utils/relationshipsParser'

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: any
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export class ApiClient {
  private baseUrl: string
  private cache: Map<string, { data: any; timestamp: number }>
  private CACHE_TTL = 5 * 60 * 1000 // 5 minutes

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl
    this.cache = new Map()
  }

  private async get<T>(path: string, useCache = true): Promise<T> {
    // Check cache first
    if (useCache) {
      const cached = this.cache.get(path)
      if (cached && Date.now() - cached.timestamp < this.CACHE_TTL) {
        return cached.data as T
      }
    }

    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!response.ok) {
      let errorMessage = 'Request failed'
      let errorCode = 'UNKNOWN'
      let errorDetails: any

      try {
        const error = await response.json()
        errorMessage = error.message || errorMessage
        errorCode = error.code || errorCode
        errorDetails = error.details
      } catch {
        // If JSON parsing fails, use status text
        errorMessage = response.statusText || errorMessage
        errorCode = response.status.toString()
      }

      throw new ApiError(
        response.status,
        errorCode,
        `${errorMessage} (${response.status})`,
        errorDetails
      )
    }

    let data: any
    try {
      data = await response.json()
    } catch (parseError) {
      throw new ApiError(
        500,
        'PARSE_ERROR',
        'Failed to parse API response',
        { originalError: parseError }
      )
    }

    // Cache successful responses
    if (useCache) {
      this.cache.set(path, { data, timestamp: Date.now() })
    }

    return data
  }

  // Dashboard-specific endpoints
  async getProjects(): Promise<Project[]> {
    return this.get<Project[]>('/projects')
  }

  async getProject(id: string): Promise<Project> {
    return this.get<Project>(`/projects/${id}`)
  }

  async getProjectDocuments(id: string): Promise<Document[]> {
    return this.get<Document[]>(`/projects/${id}/documents`)
  }

  async getProjectAnalysis(id: string): Promise<Analysis | null> {
    try {
      return await this.get<Analysis>(`/projects/${id}/analysis`)
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) {
        return null
      }
      throw error
    }
  }

  async searchProjects(request: SearchRequest): Promise<SearchResponse> {
    const params = new URLSearchParams()
    params.append('q', request.query)
    request.technologies?.forEach(t => params.append('technology', t))
    request.frameworks?.forEach(f => params.append('framework', f))
    params.append('page', request.page.toString())
    params.append('page_size', request.pageSize.toString())

    return this.get<SearchResponse>(`/search?${params.toString()}`, false)
  }

  async getRelationships(projectId?: string): Promise<Relationship[]> {
    const path = projectId ? `/relationships/${projectId}` : '/relationships'

    if (projectId) {
      // For specific project relationships, the API returns { project_id, relationships: [...] }
      const response = await this.get<{ project_id: string; relationships: ApiRelationshipResponse[] }>(path)
      return parseRelationships(response.relationships)
    } else {
      // For all relationships, the API returns array directly
      const apiRels = await this.get<ApiRelationshipResponse[]>(path)
      return parseRelationships(apiRels)
    }
  }

  async getStatistics(): Promise<Statistics> {
    const apiStats = await this.get<ApiStatisticsResponse>('/statistics', false)
    return parseStatistics(apiStats)
  }

  async getConfiguration(): Promise<Configuration> {
    return this.get<Configuration>('/configuration')
  }

  async getHealth(): Promise<HealthResponse> {
    return this.get<HealthResponse>('/health', false)
  }

  // Clear cache
  clearCache(): void {
    this.cache.clear()
  }
}

// Determine base URL based on environment
const baseUrl = import.meta.env.DEV
  ? 'http://localhost:3000'
  : '' // Production: same-origin

export const api = new ApiClient(baseUrl)