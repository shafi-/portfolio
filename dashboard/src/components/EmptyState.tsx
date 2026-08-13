interface EmptyStateProps {
  message: string
  onClearFilters: () => void
}

export default function EmptyState({ message, onClearFilters }: EmptyStateProps) {
  return (
    <div className="text-center py-12">
      <div className="text-6xl mb-4">🔍</div>
      <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
        No results found
      </h3>
      <p className="text-gray-600 dark:text-gray-400 mb-4">{message}</p>
      <button
        onClick={onClearFilters}
        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
      >
        Clear Filters
      </button>
    </div>
  )
}