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
        className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Previous page"
      >
        Previous
      </button>

      {pages.map((p, index) => (
        p === 'ellipsis' ? (
          <span key={`ellipsis-${index}`} className="px-2">...</span>
        ) : (
          <button
            key={p}
            onClick={() => onPageChange(p as number)}
            className={`px-3 py-2 border rounded-md ${
              p === page
                ? 'border-blue-600 bg-blue-600 text-white'
                : 'border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-50 dark:hover:bg-gray-600'
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
        className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
        aria-label="Next page"
      >
        Next
      </button>
    </div>
  )
}