import { Relationship } from '../types'
import { useNavigate } from 'react-router-dom'

interface RelationshipListProps {
  relationships: Relationship[]
}

export default function RelationshipList({ relationships }: RelationshipListProps) {
  const navigate = useNavigate()

  if (relationships.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="w-16 h-16 bg-gradient-to-br from-slate-100 to-slate-200 dark:from-slate-700 dark:to-slate-600 rounded-full flex items-center justify-center mx-auto mb-4">
          <span className="text-3xl">🔍</span>
        </div>
        <p className="text-slate-600 dark:text-slate-400 font-medium">No relationships found matching your filters</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {relationships.map(rel => (
        <div
          key={rel.id}
          className="card p-5 hover:shadow-lg hover:border-blue-300 dark:hover:border-blue-700 transition-all duration-200 group"
        >
          <div className="flex items-center justify-between">
            <div className="flex-1">
              <div className="flex items-center space-x-3 mb-3">
                <button
                  onClick={() => navigate(`/projects/${rel.source_project}`)}
                  className="text-sm font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline transition-colors"
                >
                  {rel.source_project}
                </button>
                <div className="flex items-center text-slate-400 dark:text-slate-500">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                  </svg>
                </div>
                <button
                  onClick={() => navigate(`/projects/${rel.target_project}`)}
                  className="text-sm font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline transition-colors"
                >
                  {rel.target_project}
                </button>
              </div>
              <div className="flex flex-wrap items-center gap-3 mb-3">
                <span className="px-3 py-1.5 text-xs font-semibold bg-gradient-to-r from-purple-500 to-purple-600 text-white rounded-full capitalize shadow-sm shadow-purple-500/20">
                  {rel.type}
                </span>
                {rel.confidence && (
                  <span className="px-3 py-1.5 text-xs font-medium bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-full flex items-center">
                    <span className="mr-1">📊</span>
                    {Math.round(rel.confidence * 100)}% confidence
                  </span>
                )}
              </div>
              {rel.description && (
                <p className="text-sm text-slate-700 dark:text-slate-300 mb-2">
                  {rel.description}
                </p>
              )}
              <p className="text-xs text-slate-500 dark:text-slate-500 flex items-center">
                <span className="mr-1">🕐</span>
                Created {new Date(rel.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}