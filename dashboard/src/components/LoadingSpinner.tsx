export default function LoadingSpinner() {
  return (
    <div className="text-center py-12">
      <div className="inline-block relative">
        {/* Outer ring */}
        <div className="w-16 h-16 border-4 border-blue-200 dark:border-blue-900 rounded-full animate-spin"></div>
        {/* Inner ring with different timing */}
        <div className="absolute inset-0 w-16 h-16 border-4 border-purple-200 dark:border-purple-900 rounded-full animate-spin"
             style={{ animationDuration: '1.5s', animationDirection: 'reverse' }}>
        </div>
        {/* Center dot */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="w-2 h-2 bg-blue-600 dark:bg-blue-400 rounded-full animate-pulse"></div>
        </div>
      </div>
      <p className="mt-6 text-slate-600 dark:text-slate-400 font-medium">Loading...</p>
    </div>
  )
}