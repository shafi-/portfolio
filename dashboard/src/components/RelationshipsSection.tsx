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
      <div className="card animate-fade-in">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">🔗</span>
          </div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
            Relationships
          </h2>
        </div>
        <div className="text-center py-12">
          <div className="inline-block relative">
            <div className="w-12 h-12 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
            <div className="absolute inset-0 w-12 h-12 border-4 border-purple-200 border-r-purple-600 rounded-full animate-spin" style={{ animationDuration: '1.5s' }}></div>
          </div>
          <p className="mt-4 text-slate-600 dark:text-slate-400 font-medium">Loading relationships...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="card animate-fade-in">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-10 h-10 bg-gradient-to-br from-red-500 to-orange-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">⚠️</span>
          </div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
            Relationships
          </h2>
        </div>
        <div className="text-center py-8">
          <p className="text-red-600 dark:text-red-400 font-medium">Failed to load relationships</p>
        </div>
      </div>
    )
  }

  if (relationships.length === 0) {
    return (
      <div className="card animate-fade-in">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-10 h-10 bg-gradient-to-br from-slate-400 to-slate-500 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">🔗</span>
          </div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
            Relationships
          </h2>
        </div>
        <div className="text-center py-12">
          <div className="w-16 h-16 bg-gradient-to-br from-slate-100 to-slate-200 dark:from-slate-700 dark:to-slate-600 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-3xl">🔗</span>
          </div>
          <p className="text-slate-600 dark:text-slate-400 font-medium">No relationships found</p>
        </div>
      </div>
    )
  }

  return (
    <div className="card animate-fade-in">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">🔗</span>
          </div>
          <div>
            <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
              Relationships
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400">{relationships.length} connections</p>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        {relationships.map(rel => (
          <div
            key={rel.id}
            className="p-5 bg-gradient-to-r from-slate-50 to-slate-100 dark:from-slate-800 dark:to-slate-700 rounded-xl border border-slate-200 dark:border-slate-600 hover:shadow-md hover:border-blue-300 dark:hover:border-blue-700 transition-all duration-200 group"
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="flex items-center space-x-3 mb-3">
                  <button
                    onClick={() => onProjectClick(rel.source_project)}
                    className="text-sm font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline transition-colors"
                  >
                    {rel.source_project_name}
                  </button>
                  <div className="flex items-center text-slate-400 dark:text-slate-500">
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                    </svg>
                  </div>
                  <button
                    onClick={() => onProjectClick(rel.target_project)}
                    className="text-sm font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 hover:underline transition-colors"
                  >
                    {rel.target_project_name}
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
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}