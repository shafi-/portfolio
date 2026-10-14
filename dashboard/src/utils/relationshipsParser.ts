/**
 * Parse relationships API response to frontend format
 */

export interface ApiRelationshipResponse {
  source_project: string
  source_project_name: string
  target_project: string
  target_project_name: string
  type: string
  description: string
  confidence: number
}

export interface Relationship {
  id: string
  source_project: string
  source_project_name: string
  target_project: string
  target_project_name: string
  type: string
  description: string
  confidence: number
  created_at: string
}

/**
 * Convert API relationship response to frontend format
 * Generates a synthetic ID from source+target+type since API doesn't return one
 */
export function parseRelationship(apiRel: ApiRelationshipResponse): Relationship {
  // Generate a synthetic ID from the relationship properties
  const syntheticId = `${apiRel.source_project}-${apiRel.target_project}-${apiRel.type}`.replace(/[^a-zA-Z0-9-]/g, '_')

  return {
    id: syntheticId,
    source_project: apiRel.source_project,
    source_project_name: apiRel.source_project_name || apiRel.source_project,
    target_project: apiRel.target_project,
    target_project_name: apiRel.target_project_name || apiRel.target_project,
    type: apiRel.type,
    description: apiRel.description,
    confidence: apiRel.confidence,
    created_at: new Date().toISOString() // API doesn't return created_at, use current time
  }
}

/**
 * Parse array of API relationship responses
 */
export function parseRelationships(apiRels: ApiRelationshipResponse[]): Relationship[] {
  if (!Array.isArray(apiRels)) {
    return []
  }
  return apiRels.map(parseRelationship)
}
