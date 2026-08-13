import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ApiClient, ApiError } from '../../services/api'
import { Project, Statistics, Relationship } from '../../types'

describe('API Client Contract Tests', () => {
  let client: ApiClient
  let mockFetch: ReturnType<typeof vi.fn>

  beforeEach(() => {
    // Mock fetch globally
    mockFetch = vi.fn()
    global.fetch = mockFetch
    client = new ApiClient('')
  })

  describe('getProjects', () => {
    it('should extract projects array from API response', async () => {
      const mockResponse = {
        projects: [
          {
            id: 'proj-1',
            name: 'Test Project',
            root_path: '/path/to/project',
            repository_type: 'git',
            discovered_at: '2024-08-13T10:00:00Z',
            updated_at: '2024-08-13T10:00:00Z',
            metadata: {
              project_id: 'proj-1',
              git_head: 'abc123',
              default_branch: 'main',
              last_commit_at: '2024-08-13T09:00:00Z',
              language_summary: 'TypeScript',
              framework_summary: 'React',
              dependency_summary: '',
              documentation_hash: 'hash',
              readme_exists: true,
              contributing_exists: false,
              license_exists: true,
            },
          },
        ],
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response)

      const result = await client.getProjects()

      expect(result).toHaveLength(1)
      expect(result[0]).toMatchObject({
        id: 'proj-1',
        name: 'Test Project',
        root_path: '/path/to/project',
      })
    })

    it('should handle empty projects array', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ projects: [] }),
      } as Response)

      const result = await client.getProjects()

      expect(result).toEqual([])
    })

    it('should throw ApiError on failed request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({ message: 'Internal Server Error' }),
      } as Response)

      await expect(client.getProjects()).rejects.toThrow(ApiError)
    })
  })

  describe('getStatistics', () => {
    it('should parse statistics API response correctly', async () => {
      const mockResponse = {
        total_projects: 10,
        projects_with_metadata: 8,
        projects_with_analysis: 5,
        projects_with_documents: 7,
        total_documents: 25,
        total_dependencies: 150,
        total_relationships: 12,
        language_counts: {
          'TypeScript, JavaScript': 5,
        },
        framework_counts: {
          'React': 4,
        },
        technology_counts: {
          'TypeScript': 5,
          'React': 4,
        },
        document_kind_counts: {
          'README': 8,
          'LICENSE': 7,
        },
        maturity_distribution: {
          low: 3,
          medium: 5,
          high: 2,
        },
        recent_activity: [],
        top_technologies: [],
      }

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response)

      const result = await client.getStatistics()

      expect(result).toMatchObject({
        totalProjects: 10,
        analyzedProjectCount: 5,
        maturityDistribution: {
          low: 5,
          medium: 3,
          high: 2,
        },
      })
    })
  })

  describe('getRelationships', () => {
    it('should parse relationships API response correctly', async () => {
      const mockResponse = [
        {
          source_project: 'proj-1',
          target_project: 'proj-2',
          type: 'Evolution',
          description: 'Test relationship',
          confidence: 0.85,
          created_at: '2024-08-13T10:00:00Z',
        },
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response)

      const result = await client.getRelationships()

      expect(result).toHaveLength(1)
      expect(result[0]).toMatchObject({
        source_project: 'proj-1',
        target_project: 'proj-2',
        type: 'Evolution',
        confidence: 0.85,
      })
    })

    it('should handle empty relationships array', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      } as Response)

      const result = await client.getRelationships()

      expect(result).toEqual([])
    })
  })

  describe('ApiError handling', () => {
    it('should create ApiError with correct properties', () => {
      const error = new ApiError(404, 'NOT_FOUND', 'Resource not found')

      expect(error.status).toBe(404)
      expect(error.code).toBe('NOT_FOUND')
      expect(error.message).toBe('Resource not found')
      expect(error.name).toBe('ApiError')
    })

    it('should handle network errors', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      await expect(client.getProjects()).rejects.toThrow()
    })
  })

  describe('Caching behavior', () => {
    it('should cache successful responses', async () => {
      const mockResponse = { projects: [] }

      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => mockResponse,
      } as Response)

      // First call
      await client.getProjects()

      // Second call should use cache
      await client.getProjects()

      // Should only call fetch once due to caching
      expect(mockFetch).toHaveBeenCalledTimes(1)
    })

    it('should not cache errors', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => ({ message: 'Server Error' }),
      } as Response)

      // First call fails
      await expect(client.getProjects()).rejects.toThrow()

      // Reset mock for second call
      mockFetch.mockResolvedValue({
        ok: true,
        json: async () => ({ projects: [] }),
      } as Response)

      // Second call should succeed (not cached error)
      await client.getProjects()

      expect(mockFetch).toHaveBeenCalledTimes(2)
    })
  })
})
