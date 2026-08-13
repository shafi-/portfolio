// Core entity types matching KnowledgeModel.md and PlatformSpecification.md

export interface Project {
  id: string
  name: string
  root_path: string
  repository_type: string
  discovered_at: string
  updated_at: string
  metadata?: Metadata
  documents?: Document[]
  analyses?: Analysis[]
}

export interface Metadata {
  git_head: string
  default_branch: string
  last_commit_at: string
  language_summary: string  // API returns comma-separated string
  framework_summary: string  // API returns comma-separated string
  dependency_summary: string  // API returns comma-separated string
  readme_exists: boolean
  contributing_exists: boolean
  license_exists: boolean
  line_count?: number
  source_line_count?: number
}

export interface Document {
  id: string
  project_id: string
  path: string
  kind: string
  content: string
  content_hash: string
  indexed_at: string
}

export interface Analysis {
  id: string
  project_id: string
  analyzer: string
  summary: string
  purpose: string
  architecture: string
  maturity: string
  strengths: string
  weaknesses: string
  reusable_components: string
  notes: string
  analyzed_git_head: string
  analyzed_at: string
}

export interface Relationship {
  id: string
  source_project: string
  target_project: string
  type: string
  description: string
  confidence: number
  created_at: string
}

export interface SearchResult {
  type: 'project' | 'document'
  id: string
  project_id?: string
  project_name?: string
  name?: string
  path?: string
  kind?: string
  snippet: string
}

export interface SearchRequest {
  query: string
  technologies?: string[]
  frameworks?: string[]
  page: number
  pageSize: number
}

export interface SearchResponse {
  results: SearchResult[]
  totalCount: number
  page: number
  pageSize: number
  totalPages: number
  highlightedSnippets?: Record<string, string>
}

export interface Statistics {
  totalProjects: number
  activeProjects: number
  uniqueTechnologies: number
  technologyDistribution: TechnologyStats[]
  maturityDistribution: MaturityDistribution
  analyzedProjectCount: number
  recentActivity: Activity[]
  topTechnologies: TechnologyStats[]
}

export interface TechnologyStats {
  name: string
  count: number
}

export interface MaturityDistribution {
  low: number
  medium: number
  high: number
}

export interface Activity {
  project_id: string
  project_name: string
  activity_type: string
  timestamp: string
}

export interface Configuration {
  dashboard?: DashboardConfig
  database?: DatabaseConfig
  discovery?: DiscoveryConfig
}

export interface DashboardConfig {
  defaultPageSize: number
  theme: string
}

export interface DatabaseConfig {
  path: string
}

export interface DiscoveryConfig {
  paths: string[]
  excludePatterns: string[]
}

export interface HealthResponse {
  status: string
  uptimeSeconds: number
  databaseConnected: boolean
  dashboardServing: boolean
  version: string
}