import { Document } from '../types'

interface DocumentationSectionProps {
  documents: Document[]
  loading: boolean
  error: Error | undefined
}

export default function DocumentationSection({ documents, loading, error }: DocumentationSectionProps) {
  if (loading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Documentation
        </h2>
        <div className="text-center py-8">
          <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Documentation
        </h2>
        <p className="text-red-600 dark:text-red-400">Failed to load documentation</p>
      </div>
    )
  }

  if (documents.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Documentation
        </h2>
        <p className="text-gray-600 dark:text-gray-400">No documentation found</p>
      </div>
    )
  }

  // Group documents by kind
  const groupedDocs = documents.reduce((acc, doc) => {
    if (!acc[doc.kind]) {
      acc[doc.kind] = []
    }
    acc[doc.kind].push(doc)
    return acc
  }, {} as Record<string, Document[]>)

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
        Documentation ({documents.length})
      </h2>

      <div className="space-y-6">
        {Object.entries(groupedDocs).map(([kind, docs]) => (
          <div key={kind}>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-3 capitalize">
              {kind}
            </h3>
            <div className="space-y-2">
              {docs.map(doc => (
                <div
                  key={doc.id}
                  className="p-3 border border-gray-200 dark:border-gray-700 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <p className="text-sm font-medium text-gray-900 dark:text-white">
                        {doc.path}
                      </p>
                      <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                        Indexed {new Date(doc.indexed_at).toLocaleDateString()}
                      </p>
                    </div>
                    <span className="text-xs text-gray-500 dark:text-gray-400 ml-4">
                      {doc.content.length} bytes
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}