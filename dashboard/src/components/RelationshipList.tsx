import { Relationship } from '../types'
import { useNavigate } from 'react-router-dom'

interface RelationshipListProps {
  relationships: Relationship[]
}

export default function RelationshipList({ relationships }: RelationshipListProps) {
  const navigate = useNavigate()

  if (relationships.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">
        No relationships found matching your filters
      </div>
    )
  }

  return (
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
                  onClick={() => navigate(`/projects/${rel.source_project}`)}
                  className="text-sm font-medium text-blue-600 dark:text-blue-400 hover:underline"
                >
                  {rel.source_project}
                </button>
                <span className="text-gray-400">→</span>
                <button
                  onClick={() => navigate(`/projects/${rel.target_project}`)}
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
              <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                Created {new Date(rel.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}