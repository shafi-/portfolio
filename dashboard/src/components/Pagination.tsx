interface PaginationState {
  page: number
  pageSize: number
  totalCount: number
  totalPages?: number
}

interface PaginationProps {
  pagination: PaginationState
  onPageChange: (page: number) => void
}

export default function Pagination({ pagination, onPageChange }: PaginationProps) {
  const { page, totalPages = 1 } = pagination

  if (totalPages <= 1) {
    return null
  }

  const pages = []
  const showEllipsis = totalPages > 7

  if (showEllipsis) {
    if (page <= 3) {
      for (let i = 1; i <= 5; i++) pages.push(i)
      pages.push('ellipsis')
      pages.push(totalPages)
    } else if (page >= totalPages - 2) {
      pages.push(1)
      pages.push('ellipsis')
      for (let i = totalPages - 4; i <= totalPages; i++) pages.push(i)
    } else {
      pages.push(1)
      pages.push('ellipsis')
      for (let i = page - 1; i <= page + 1; i++) pages.push(i)
      pages.push('ellipsis')
      pages.push(totalPages)
    }
  } else {
    for (let i = 1; i <= totalPages; i++) pages.push(i)
  }

  return (
    <div className="flex items-center justify-center space-x-2">
      <button
        onClick={() => onPageChange(Math.max(1, page - 1))}
        disabled={page === 1}
        className="px-4 py-2.5 border-2 border-slate-200 dark:border-slate-700 rounded-xl bg-white dark:bg-slate-800 text-slate-900 dark:text-white hover:bg-slate-50 dark:hover:bg-slate-700 hover:border-blue-300 dark:hover:border-blue-700 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:border-slate-200 dark:disabled:hover:border-slate-700 disabled:hover:bg-white dark:disabled:hover:bg-slate-800 transition-all duration-200 font-medium shadow-sm hover:shadow"
        aria-label="Previous page"
      >
        ← Previous
      </button>

      {pages.map((p, index) => (
        p === 'ellipsis' ? (
          <span key={`ellipsis-${index}`} className="px-3 text-slate-500 dark:text-slate-400 font-medium">...</span>
        ) : (
          <button
            key={p}
            onClick={() => onPageChange(p as number)}
            className={`px-4 py-2.5 border-2 rounded-xl font-medium transition-all duration-200 shadow-sm ${
              p === page
                ? 'bg-gradient-to-r from-blue-500 to-purple-600 text-white border-transparent shadow-lg shadow-blue-500/20'
                : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-900 dark:text-white hover:bg-slate-50 dark:hover:bg-slate-700 hover:border-blue-300 dark:hover:border-blue-700 hover:shadow'
            }`}
            aria-label={`Page ${p}`}
            aria-current={p === page ? 'page' : undefined}
          >
            {p}
          </button>
        )
      ))}

      <button
        onClick={() => onPageChange(Math.min(totalPages, page + 1))}
        disabled={page === totalPages}
        className="px-4 py-2.5 border-2 border-slate-200 dark:border-slate-700 rounded-xl bg-white dark:bg-slate-800 text-slate-900 dark:text-white hover:bg-slate-50 dark:hover:bg-slate-700 hover:border-blue-300 dark:hover:border-blue-700 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:border-slate-200 dark:disabled:hover:border-slate-700 disabled:hover:bg-white dark:disabled:hover:bg-slate-800 transition-all duration-200 font-medium shadow-sm hover:shadow"
        aria-label="Next page"
      >
        Next →
      </button>
    </div>
  )
}