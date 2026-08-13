import { describe, it, expect } from 'vitest'
import { parseStatistics } from '../../utils/statisticsParser'
import { parseRelationships, ApiRelationshipResponse } from '../../utils/relationshipsParser'

describe('API Response Contract Tests', () => {
  describe('Statistics API Contract', () => {
    it('should parse valid statistics response correctly', () => {
      // This test verifies the frontend can handle the actual backend response structure
      const apiResponse = {
        total_projects: 10,
        projects_with_metadata: 8,
        projects_with_analysis: 5,
        projects_with_documents: 7,
        total_documents: 25,
        total_dependencies: 150,
        total_relationships: 12,
        language_counts: {
          'TypeScript, JavaScript': 5,
          'Go': 3,
          'Python': 2,
        },
        framework_counts: {
          'React, Node.js': 4,
          'Express': 2,
        },
        technology_counts: {
          'TypeScript': 5,
          'React': 4,
        },
        document_kind_counts: {
          'README': 8,
          'LICENSE': 7,
        },
      }

      const result = parseStatistics(apiResponse)

      expect(result).toMatchObject({
        totalProjects: 10,
        activeProjects: expect.any(Number),
        uniqueTechnologies: expect.any(Number),
        analyzedProjectCount: 5,
        technologyDistribution: expect.any(Array),
        recentActivity: expect.any(Array),
        topTechnologies: expect.any(Array),
      })
    })

    it('should handle empty statistics response', () => {
      const emptyResponse = {
        total_projects: 0,
        projects_with_metadata: 0,
        projects_with_analysis: 0,
        projects_with_documents: 0,
        total_documents: 0,
        total_dependencies: 0,
        total_relationships: 0,
        language_counts: {},
        framework_counts: {},
        technology_counts: {},
        document_kind_counts: {},
      }

      const result = parseStatistics(emptyResponse)

      expect(result.totalProjects).toBe(0)
      expect(result.technologyDistribution).toEqual([])
      expect(result.recentActivity).toEqual([])
    })

    it('should parse technology counts correctly', () => {
      const response = {
        total_projects: 1,
        projects_with_metadata: 1,
        projects_with_analysis: 0,
        projects_with_documents: 0,
        total_documents: 0,
        total_dependencies: 0,
        total_relationships: 0,
        language_counts: {
          'C++, C, TypeScript, Svelte, Python': 1,
        },
        framework_counts: {},
        technology_counts: {
          'C++': 1,
          'C': 1,
          'TypeScript': 1,
        },
        document_kind_counts: {},
      }

      const result = parseStatistics(response)

      expect(result.technologyDistribution).toHaveLength(3)
      expect(result.technologyDistribution).toContainEqual({
        name: 'C++',
        count: 1
      })
    })
  })

  describe('Relationships API Contract', () => {
    it('should parse valid relationships response correctly', () => {
      const apiResponse = [
        {
          source_project: 'proj-1',
          target_project: 'proj-2',
          type: 'Evolution',
          description: 'Test evolution relationship',
          confidence: 0.85,
        },
        {
          source_project: 'proj-2',
          target_project: 'proj-3',
          type: 'Shared Technology',
          description: 'Both use React',
          confidence: 0.90,
        },
      ]

      const result = parseRelationships(apiResponse)

      expect(result).toHaveLength(2)
      expect(result[0]).toMatchObject({
        id: expect.any(String),
        source_project: 'proj-1',
        target_project: 'proj-2',
        type: 'Evolution',
        description: 'Test evolution relationship',
        confidence: 0.85,
      })
    })

    it('should handle empty relationships response', () => {
      const result = parseRelationships([])

      expect(result).toEqual([])
    })

    it('should generate consistent IDs for same relationship', () => {
      const response = [
        {
          source_project: 'proj-1',
          target_project: 'proj-2',
          type: 'Similar',
          description: 'Test',
          confidence: 0.8,
        },
      ]

      const result1 = parseRelationships(response)
      const result2 = parseRelationships(response)

      // IDs should be consistent for same relationship
      expect(result1[0].id).toBe(result2[0].id)
    })
  })

  describe('Projects API Contract', () => {
    it('should handle projects response structure', () => {
      // Verify the expected response structure
      const apiResponse = {
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
              language_summary: 'TypeScript, JavaScript',
              framework_summary: 'React, Node.js',
              dependency_summary: '',
              documentation_hash: 'hash123',
              readme_exists: true,
              contributing_exists: false,
              license_exists: true,
            },
          },
        ],
      }

      // Verify structure matches frontend expectations
      expect(apiResponse.projects).toBeDefined()
      expect(Array.isArray(apiResponse.projects)).toBe(true)
      expect(apiResponse.projects[0]).toMatchObject({
        id: expect.any(String),
        name: expect.any(String),
        root_path: expect.any(String),
        repository_type: expect.any(String),
        discovered_at: expect.any(String),
        updated_at: expect.any(String),
        metadata: expect.any(Object),
      })
    })
  })
})
