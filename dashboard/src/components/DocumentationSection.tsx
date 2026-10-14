import { Document } from '../types'

interface DocumentationSectionProps {
  documents: Document[]
  loading: boolean
  error: Error | undefined
}

export default function DocumentationSection({ documents, loading, error }: DocumentationSectionProps) {
  if (loading) {
    return (
      <div className="card animate-fade-in">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">📚</span>
          </div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
            Documentation
          </h2>
        </div>
        <div className="text-center py-12">
          <div className="inline-block relative">
            <div className="w-12 h-12 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
            <div className="absolute inset-0 w-12 h-12 border-4 border-purple-200 border-r-purple-600 rounded-full animate-spin" style={{ animationDuration: '1.5s' }}></div>
          </div>
          <p className="mt-4 text-slate-600 dark:text-slate-400 font-medium">Loading documentation...</p>
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
            Documentation
          </h2>
        </div>
        <div className="text-center py-8">
          <p className="text-red-600 dark:text-red-400 font-medium">Failed to load documentation</p>
        </div>
      </div>
    )
  }

  if (documents.length === 0) {
    return (
      <div className="card animate-fade-in">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-10 h-10 bg-gradient-to-br from-slate-400 to-slate-500 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">📚</span>
          </div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
            Documentation
          </h2>
        </div>
        <div className="text-center py-12">
          <div className="w-16 h-16 bg-gradient-to-br from-slate-100 to-slate-200 dark:from-slate-700 dark:to-slate-600 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-3xl">📭</span>
          </div>
          <p className="text-slate-600 dark:text-slate-400 font-medium">No documentation found</p>
          <p className="text-sm text-slate-500 dark:text-slate-500 mt-2">Documentation will appear here when available</p>
        </div>
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
    <div className="card animate-fade-in">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">📚</span>
          </div>
          <div>
            <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
              Documentation
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400">{documents.length} documents indexed</p>
          </div>
        </div>
      </div>

      <div className="space-y-6">
        {Object.entries(groupedDocs).map(([kind, docs]) => (
          <div key={kind} className="space-y-3">
            <div className="flex items-center space-x-2">
              <div className="w-6 h-6 bg-gradient-to-br from-blue-500 to-purple-600 rounded-md flex items-center justify-center">
                <span className="text-white text-xs">📄</span>
              </div>
              <h3 className="text-lg font-semibold text-slate-900 dark:text-white capitalize">
                {kind} ({docs.length})
              </h3>
            </div>
            <div className="space-y-2 pl-8">
              {docs.map(doc => (
                <div
                  key={doc.id}
                  className="p-4 bg-gradient-to-r from-slate-50 to-slate-100 dark:from-slate-800 dark:to-slate-700 rounded-xl border border-slate-200 dark:border-slate-600 hover:shadow-md hover:border-blue-300 dark:hover:border-blue-700 transition-all duration-200"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <p className="text-sm font-semibold text-slate-900 dark:text-white flex items-center">
                        <span className="mr-2">📄</span>
                        {doc.path}
                      </p>
                      <div className="flex items-center mt-2 space-x-4">
                        <p className="text-xs text-slate-600 dark:text-slate-400 flex items-center">
                          <span className="mr-1">🕐</span>
                          Indexed {new Date(doc.indexed_at).toLocaleDateString()}
                        </p>
                        <span className="text-xs px-2 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-full font-medium">
                          {(doc.content.length / 1024).toFixed(1)} KB
                        </span>
                      </div>
                    </div>
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