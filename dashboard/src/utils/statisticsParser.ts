/**
 * Parse statistics API response to frontend format
 */

export interface ApiStatisticsResponse {
  total_projects: number
  projects_with_metadata: number
  projects_with_analysis: number
  projects_with_documents: number
  total_documents: number
  total_dependencies: number
  total_relationships: number
  language_counts: Record<string, number>
  framework_counts: Record<string, number>
  technology_counts: Record<string, number>
  document_kind_counts: Record<string, number>
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

/**
 * Convert API statistics response to frontend format
 */
export function parseStatistics(apiStats: ApiStatisticsResponse): Statistics {
  // Convert technology counts to array format for frontend consumption
  const technologyDistribution: TechnologyStats[] = Object.entries(apiStats.technology_counts)
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)

  const topTechnologies = technologyDistribution.slice(0, 10)

  // Calculate maturity distribution (placeholder logic)
  const totalProjects = apiStats.total_projects
  const analyzedProjects = apiStats.projects_with_analysis

  const maturityDistribution: MaturityDistribution = {
    low: Math.max(0, totalProjects - analyzedProjects),
    medium: Math.floor(analyzedProjects * 0.6),
    high: Math.ceil(analyzedProjects * 0.4)
  }

  return {
    totalProjects: apiStats.total_projects,
    activeProjects: apiStats.projects_with_metadata || 0,
    uniqueTechnologies: technologyDistribution.length,
    technologyDistribution,
    maturityDistribution,
    analyzedProjectCount: apiStats.projects_with_analysis || 0,
    recentActivity: [], // Not implemented in API yet
    topTechnologies
  }
}

/**
 * Convert counts record to TechnologyStats array
 */
function countsToStats(counts: Record<string, number>): TechnologyStats[] {
  return Object.entries(counts).map(([name, count]) => ({ name, count }))
}
