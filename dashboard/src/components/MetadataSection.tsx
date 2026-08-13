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
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Metadata
        </h2>
        <div className="text-center py-8">
          <div className="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
        </div>
      </div>
    )
  }

  if (error || !project) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Metadata
        </h2>
        <p className="text-red-600 dark:text-red-400">Failed to load metadata</p>
      </div>
    )
  }

  const metadata = project.metadata
  const languages = parseLanguageSummary(metadata.language_summary)
  const frameworks = parseFrameworkSummary(metadata.framework_summary)

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
        Metadata
      </h2>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
            Repository Information
          </h3>
          <dl className="space-y-2">
            <div>
              <dt className="text-sm text-gray-600 dark:text-gray-400">Default Branch</dt>
              <dd className="text-sm font-medium text-gray-900 dark:text-white">
                {metadata.default_branch}
              </dd>
            </div>
            <div>
              <dt className="text-sm text-gray-600 dark:text-gray-400">Git HEAD</dt>
              <dd className="text-sm font-mono text-gray-900 dark:text-white">
                {metadata.git_head.substring(0, 8)}
              </dd>
            </div>
            <div>
              <dt className="text-sm text-gray-600 dark:text-gray-400">Last Commit</dt>
              <dd className="text-sm text-gray-900 dark:text-white">
                {new Date(metadata.last_commit_at).toLocaleString()}
              </dd>
            </div>
          </dl>
        </div>

        <div>
          <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
            Technologies & Languages
          </h3>
          <div className="space-y-2">
            <div>
              <dt className="text-sm text-gray-600 dark:text-gray-400">Languages</dt>
              <dd className="flex flex-wrap gap-1 mt-1">
                {languages.map(lang => (
                  <span
                    key={lang}
                    className="px-2 py-1 text-xs bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded"
                  >
                    {lang}
                  </span>
                ))}
              </dd>
            </div>
            <div>
              <dt className="text-sm text-gray-600 dark:text-gray-400">Frameworks</dt>
              <dd className="flex flex-wrap gap-1 mt-1">
                {frameworks.map(fw => (
                  <span
                    key={fw}
                    className="px-2 py-1 text-xs bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded"
                  >
                    {fw}
                  </span>
                ))}
              </dd>
            </div>
          </div>
        </div>

        <div>
          <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
            Documentation
          </h3>
          <div className="space-y-1">
            <div className="flex items-center">
              <span className={`mr-2 ${metadata.readme_exists ? '✅' : '❌'}`} role="img" aria-label={metadata.readme_exists ? 'exists' : 'missing'}>
                {metadata.readme_exists ? '✓' : '✗'}
              </span>
              <span className="text-sm text-gray-900 dark:text-white">README.md</span>
            </div>
            <div className="flex items-center">
              <span className={`mr-2 ${metadata.contributing_exists ? '✅' : '❌'}`} role="img" aria-label={metadata.contributing_exists ? 'exists' : 'missing'}>
                {metadata.contributing_exists ? '✓' : '✗'}
              </span>
              <span className="text-sm text-gray-900 dark:text-white">CONTRIBUTING.md</span>
            </div>
            <div className="flex items-center">
              <span className={`mr-2 ${metadata.license_exists ? '✅' : '❌'}`} role="img" aria-label={metadata.license_exists ? 'exists' : 'missing'}>
                {metadata.license_exists ? '✓' : '✗'}
              </span>
              <span className="text-sm text-gray-900 dark:text-white">LICENSE</span>
            </div>
          </div>
        </div>

        <div>
          <h3 className="text-sm font-medium text-gray-500 dark:text-gray-400 mb-2">
            Code Metrics
          </h3>
          <div className="space-y-1">
            {metadata.line_count && (
              <div>
                <dt className="text-sm text-gray-600 dark:text-gray-400">Total Lines</dt>
                <dd className="text-sm font-medium text-gray-900 dark:text-white">
                  {metadata.line_count.toLocaleString()}
                </dd>
              </div>
            )}
            {metadata.source_line_count && (
              <div>
                <dt className="text-sm text-gray-600 dark:text-gray-400">Source Lines</dt>
                <dd className="text-sm font-medium text-gray-900 dark:text-white">
                  {metadata.source_line_count.toLocaleString()}
                </dd>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}