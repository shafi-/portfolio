import { Relationship } from '../types'

interface RelationshipsSectionProps {
  relationships: Relationship[]
  loading: boolean
  error: Error | undefined
  onProjectClick: (projectId: string) => void
}

export default function RelationshipsSection({ relationships, loading, error, onProjectClick }: RelationshipsSectionProps) {
  if (loading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Relationships
        </h2>
        <div className="text-center py-8">
          <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Relationships
        </h2>
        <p className="text-red-600 dark:text-red-400">Failed to load relationships</p>
      </div>
    )
  }

  if (relationships.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Relationships
        </h2>
        <p className="text-gray-600 dark:text-gray-400">No relationships found</p>
      </div>
    )
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
        Relationships ({relationships.length})
      </h2>

      <div className="space-y-4">
        {relationships.map(rel => (
          <div
            key={rel.id}
            className="p-4 border border-gray-200 dark:border-gray-700 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => onProjectClick(rel.source_project)}
                    className="text-sm font-medium text-blue-600 dark:text-blue-400 hover:underline"
                  >
                    {rel.source_project}
                  </button>
                  <span className="text-gray-400">→</span>
                  <button
                    onClick={() => onProjectClick(rel.target_project)}
                    className="text-sm font-medium text-blue-600 dark:text-blue-400 hover:underline"
                  >
                    {rel.target_project}
                  </button>
                </div>
                <div className="mt-2">
                  <span className="inline-block px-2 py-1 text-xs bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200 rounded capitalize">
                    {rel.type}
                  </span>
                  {rel.confidence && (
                    <span className="ml-2 text-xs text-gray-500 dark:text-gray-400">
                      Confidence: {Math.round(rel.confidence * 100)}%
                    </span>
                  )}
                </div>
                {rel.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
                    {rel.description}
                  </p>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}