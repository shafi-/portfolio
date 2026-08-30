interface EmptyStateProps {
  message: string
  onClearFilters: () => void
}

export default function EmptyState({ message, onClearFilters }: EmptyStateProps) {
  return (
    <div className="card text-center py-16">
      <div className="w-20 h-20 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-slate-700 dark:to-slate-600 rounded-2xl flex items-center justify-center mx-auto mb-6">
        <span className="text-4xl">🔍</span>
      </div>
      <h3 className="text-xl font-semibold text-slate-900 dark:text-white mb-2">
        No results found
      </h3>
      <p className="text-slate-600 dark:text-slate-400 mb-6 max-w-md mx-auto">
        {message}
      </p>
      <button
        onClick={onClearFilters}
        className="btn btn-primary inline-flex items-center space-x-2"
      >
        <span>✨</span>
        <span>Clear Filters</span>
      </button>
    </div>
  )
}