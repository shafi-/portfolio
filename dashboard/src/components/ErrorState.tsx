import { ApiError } from '../services/api'

interface ErrorStateProps {
  error: Error
  onRetry: () => void
}

export default function ErrorState({ error, onRetry }: ErrorStateProps) {
  const isApiError = error instanceof ApiError
  const errorMessage = isApiError ? error.message : 'An unexpected error occurred'

  return (
    <div className="text-center py-12">
      <div className="text-6xl mb-4">⚠️</div>
      <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
        Error Loading Data
      </h3>
      <p className="text-gray-600 dark:text-gray-400 mb-4">{errorMessage}</p>

      {isApiError && error.status >= 500 && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4 mb-4 text-left">
          <p className="text-sm text-yellow-800 dark:text-yellow-200">
            <strong>Server Error:</strong> The portfolio engine is experiencing issues.
            Try running <code className="bg-yellow-100 dark:bg-yellow-900 px-2 py-1 rounded">portfolio doctor</code> to diagnose the problem.
          </p>
        </div>
      )}

      {isApiError && error.status === 404 && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-4 text-left">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            <strong>Data Not Found:</strong> The requested data doesn't exist.
            This might happen if you haven't discovered projects yet. Try running <code className="bg-blue-100 dark:bg-blue-900 px-2 py-1 rounded">portfolio discover</code>
          </p>
        </div>
      )}

      <div className="flex gap-4 justify-center">
        <button
          onClick={onRetry}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Retry
        </button>
        <button
          onClick={() => window.location.reload()}
          className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700"
        >
          Reload Page
        </button>
      </div>

      {process.env.NODE_ENV === 'development' && (
        <details className="mt-6 text-left">
          <summary className="cursor-pointer text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
            Technical Details (Dev Mode)
          </summary>
          <pre className="mt-2 p-4 bg-gray-100 dark:bg-gray-800 rounded text-xs overflow-auto max-h-40">
            {JSON.stringify({
              message: error.message,
              status: isApiError ? error.status : 'unknown',
              code: isApiError ? error.code : 'unknown',
              stack: error.stack,
              details: isApiError ? error.details : undefined
            }, null, 2)}
          </pre>
        </details>
      )}
    </div>
  )
}