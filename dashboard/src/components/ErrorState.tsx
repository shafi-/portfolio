import { ApiError } from '../services/api'

interface ErrorStateProps {
  error: Error
  onRetry: () => void
}

export default function ErrorState({ error, onRetry }: ErrorStateProps) {
  const isApiError = error instanceof ApiError
  const errorMessage = isApiError ? error.message : 'An unexpected error occurred'

  return (
    <div className="card text-center py-16">
      <div className="w-20 h-20 bg-gradient-to-br from-red-50 to-orange-50 dark:from-red-900/20 dark:to-orange-900/20 rounded-2xl flex items-center justify-center mx-auto mb-6">
        <span className="text-4xl">⚠️</span>
      </div>
      <h3 className="text-xl font-semibold text-slate-900 dark:text-white mb-2">
        Error Loading Data
      </h3>
      <p className="text-slate-600 dark:text-slate-400 mb-6 max-w-md mx-auto">{errorMessage}</p>

      {isApiError && error.status >= 500 && (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-xl p-4 mb-6 text-left max-w-lg mx-auto">
          <p className="text-sm text-yellow-800 dark:text-yellow-200">
            <strong>🔧 Server Error:</strong> The portfolio engine is experiencing issues.
            Try running <code className="bg-yellow-100 dark:bg-yellow-900 px-2 py-1 rounded text-xs">portfolio doctor</code> to diagnose the problem.
          </p>
        </div>
      )}

      {isApiError && error.status === 404 && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-xl p-4 mb-6 text-left max-w-lg mx-auto">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            <strong>🔍 Data Not Found:</strong> The requested data doesn't exist.
            This might happen if you haven't discovered projects yet. Try running <code className="bg-blue-100 dark:bg-blue-900 px-2 py-1 rounded text-xs">portfolio discover</code>
          </p>
        </div>
      )}

      <div className="flex gap-4 justify-center">
        <button
          onClick={onRetry}
          className="btn btn-primary inline-flex items-center space-x-2"
        >
          <span>🔄</span>
          <span>Retry</span>
        </button>
        <button
          onClick={() => window.location.reload()}
          className="btn btn-secondary inline-flex items-center space-x-2"
        >
          <span>🔃</span>
          <span>Reload Page</span>
        </button>
      </div>

      {process.env.NODE_ENV === 'development' && (
        <details className="mt-8 text-left">
          <summary className="cursor-pointer text-sm text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300 transition-colors">
            Technical Details (Dev Mode)
          </summary>
          <div className="mt-4 p-4 bg-slate-100 dark:bg-slate-800 rounded-lg text-xs overflow-auto max-h-40">
            <pre className="text-slate-700 dark:text-slate-300">
              {JSON.stringify({
                message: error.message,
                status: isApiError ? error.status : 'unknown',
                code: isApiError ? error.code : 'unknown',
                stack: error.stack,
                details: isApiError ? error.details : undefined
              }, null, 2)}
            </pre>
          </div>
        </details>
      )}
    </div>
  )
}