interface ErrorStateProps {
  error: Error
  onRetry: () => void
}

export default function ErrorState({ error, onRetry }: ErrorStateProps) {
  return (
    <div className="text-center py-12">
      <div className="text-6xl mb-4">⚠️</div>
      <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
        Error Loading Data
      </h3>
      <p className="text-gray-600 dark:text-gray-400 mb-4">{error.message}</p>
      <button
        onClick={onRetry}
        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
      >
        Retry
      </button>
    </div>
  )
}