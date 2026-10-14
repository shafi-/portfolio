import { Project } from '../types'
import { parseLanguageSummary, parseFrameworkSummary } from '../utils/parsers'

interface MetadataSectionProps {
  project: Project | null
  loading: boolean
  error: Error | undefined
}

export default function MetadataSection({ project, loading, error }: MetadataSectionProps) {
  if (loading) {
    return (
      <div className="card animate-fade-in">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">📊</span>
          </div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
            Metadata
          </h2>
        </div>
        <div className="text-center py-12">
          <div className="inline-block relative">
            <div className="w-12 h-12 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
            <div className="absolute inset-0 w-12 h-12 border-4 border-purple-200 border-r-purple-600 rounded-full animate-spin" style={{ animationDuration: '1.5s' }}></div>
          </div>
          <p className="mt-4 text-slate-600 dark:text-slate-400 font-medium">Loading metadata...</p>
        </div>
      </div>
    )
  }

  if (error || !project) {
    return (
      <div className="card animate-fade-in">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-10 h-10 bg-gradient-to-br from-red-500 to-orange-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">⚠️</span>
          </div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
            Metadata
          </h2>
        </div>
        <div className="text-center py-8">
          <p className="text-red-600 dark:text-red-400 font-medium">Failed to load metadata</p>
        </div>
      </div>
    )
  }

  const metadata = project.metadata
  const languages = parseLanguageSummary(metadata.language_summary)
  const frameworks = parseFrameworkSummary(metadata.framework_summary)

  return (
    <div className="card animate-fade-in">
      <div className="flex items-center space-x-3 mb-6">
        <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
          <span className="text-white text-lg">📊</span>
        </div>
        <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
          Metadata
        </h2>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-4">
          <div className="p-4 bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-800 dark:to-slate-700 rounded-xl">
            <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 flex items-center">
              <span className="mr-2">📁</span>
              Repository Information
            </h3>
            <dl className="space-y-2">
              <div className="flex justify-between items-center py-2 border-b border-slate-200 dark:border-slate-600">
                <dt className="text-sm text-slate-600 dark:text-slate-400">Default Branch</dt>
                <dd className="text-sm font-semibold text-slate-900 dark:text-white bg-slate-200 dark:bg-slate-700 px-2 py-1 rounded">
                  {metadata.default_branch}
                </dd>
              </div>
              <div className="flex justify-between items-center py-2 border-b border-slate-200 dark:border-slate-600">
                <dt className="text-sm text-slate-600 dark:text-slate-400">Git HEAD</dt>
                <dd className="text-sm font-mono font-semibold text-slate-900 dark:text-white bg-slate-200 dark:bg-slate-700 px-2 py-1 rounded">
                  {metadata.git_head.substring(0, 8)}
                </dd>
              </div>
              <div className="flex justify-between items-center py-2">
                <dt className="text-sm text-slate-600 dark:text-slate-400">Last Commit</dt>
                <dd className="text-sm text-slate-900 dark:text-white">
                  {new Date(metadata.last_commit_at).toLocaleString()}
                </dd>
              </div>
            </dl>
          </div>
        </div>

        <div className="space-y-4">
          <div className="p-4 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20 rounded-xl">
            <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 flex items-center">
              <span className="mr-2">🔧</span>
              Technologies & Languages
            </h3>
            <div className="space-y-3">
              <div>
                <dt className="text-xs font-medium text-slate-600 dark:text-slate-400 mb-2">Languages</dt>
                <dd className="flex flex-wrap gap-2">
                  {languages.map(lang => (
                    <span
                      key={lang}
                      className="px-3 py-1.5 text-xs font-medium bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-lg shadow-sm shadow-blue-500/20 hover:shadow-md hover:shadow-blue-500/30 transition-shadow"
                    >
                      {lang}
                    </span>
                  ))}
                </dd>
              </div>
              <div>
                <dt className="text-xs font-medium text-slate-600 dark:text-slate-400 mb-2">Frameworks</dt>
                <dd className="flex flex-wrap gap-2">
                  {frameworks.map(fw => (
                    <span
                      key={fw}
                      className="px-3 py-1.5 text-xs font-medium bg-gradient-to-r from-green-500 to-green-600 text-white rounded-lg shadow-sm shadow-green-500/20 hover:shadow-md hover:shadow-green-500/30 transition-shadow"
                    >
                      {fw}
                    </span>
                  ))}
                </dd>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-4">
          <div className="p-4 bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-900/20 dark:to-emerald-900/20 rounded-xl">
            <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 flex items-center">
              <span className="mr-2">📚</span>
              Documentation
            </h3>
            <div className="space-y-2">
              <div className="flex items-center justify-between p-2 bg-white dark:bg-slate-800 rounded-lg">
                <div className="flex items-center space-x-2">
                  <span className={`w-5 h-5 flex items-center justify-center rounded-full ${metadata.readme_exists ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                    {metadata.readme_exists ? '✓' : '✗'}
                  </span>
                  <span className="text-sm font-medium text-slate-900 dark:text-white">README.md</span>
                </div>
                <span className={`text-xs px-2 py-1 rounded-full ${metadata.readme_exists ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'}`}>
                  {metadata.readme_exists ? 'Present' : 'Missing'}
                </span>
              </div>
              <div className="flex items-center justify-between p-2 bg-white dark:bg-slate-800 rounded-lg">
                <div className="flex items-center space-x-2">
                  <span className={`w-5 h-5 flex items-center justify-center rounded-full ${metadata.contributing_exists ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                    {metadata.contributing_exists ? '✓' : '✗'}
                  </span>
                  <span className="text-sm font-medium text-slate-900 dark:text-white">CONTRIBUTING.md</span>
                </div>
                <span className={`text-xs px-2 py-1 rounded-full ${metadata.contributing_exists ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'}`}>
                  {metadata.contributing_exists ? 'Present' : 'Missing'}
                </span>
              </div>
              <div className="flex items-center justify-between p-2 bg-white dark:bg-slate-800 rounded-lg">
                <div className="flex items-center space-x-2">
                  <span className={`w-5 h-5 flex items-center justify-center rounded-full ${metadata.license_exists ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'}`}>
                    {metadata.license_exists ? '✓' : '✗'}
                  </span>
                  <span className="text-sm font-medium text-slate-900 dark:text-white">LICENSE</span>
                </div>
                <span className={`text-xs px-2 py-1 rounded-full ${metadata.license_exists ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'}`}>
                  {metadata.license_exists ? 'Present' : 'Missing'}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-4">
          <div className="p-4 bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl">
            <h3 className="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 flex items-center">
              <span className="mr-2">📈</span>
              Code Metrics
            </h3>
            <div className="space-y-3">
              {metadata.line_count && (
                <div className="p-3 bg-white dark:bg-slate-800 rounded-lg">
                  <div className="flex justify-between items-center">
                    <div>
                      <dt className="text-xs font-medium text-slate-600 dark:text-slate-400">Total Lines</dt>
                      <dd className="text-2xl font-bold bg-gradient-to-r from-purple-600 to-pink-600 bg-clip-text text-transparent">
                        {metadata.line_count.toLocaleString()}
                      </dd>
                    </div>
                    <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center">
                      <span className="text-white text-lg">📝</span>
                    </div>
                  </div>
                </div>
              )}
              {metadata.source_line_count && (
                <div className="p-3 bg-white dark:bg-slate-800 rounded-lg">
                  <div className="flex justify-between items-center">
                    <div>
                      <dt className="text-xs font-medium text-slate-600 dark:text-slate-400">Source Lines</dt>
                      <dd className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-cyan-600 bg-clip-text text-transparent">
                        {metadata.source_line_count.toLocaleString()}
                      </dd>
                    </div>
                    <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-cyan-500 rounded-lg flex items-center justify-center">
                      <span className="text-white text-lg">⚙️</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}