import { Analysis } from '../types'

interface AnalysisSectionProps {
  analysis: Analysis | null
  loading: boolean
  error: Error | undefined
}

export default function AnalysisSection({ analysis, loading, error }: AnalysisSectionProps) {
  if (loading) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          AI Analysis
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
          AI Analysis
        </h2>
        <p className="text-red-600 dark:text-red-400">Failed to load analysis</p>
      </div>
    )
  }

  if (!analysis) {
    return (
      <div className="bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20 rounded-lg shadow p-6">
        <div className="text-center">
          <div className="text-4xl mb-4">🤖</div>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
            No AI Analysis Available
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            This project hasn't been analyzed by an AI agent yet. Analysis provides deeper insights into architecture, patterns, and relationships.
          </p>
          <div className="text-sm text-gray-500 dark:text-gray-500">
            <p>Analysis requires an AI agent integration with Portfolio.</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
          AI Analysis
        </h2>
        <span className="text-sm text-gray-500 dark:text-gray-400">
          Analyzed {new Date(analysis.analyzed_at).toLocaleDateString()}
        </span>
      </div>

      <div className="space-y-6">
        {analysis.purpose && (
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
              Purpose
            </h3>
            <p className="text-gray-600 dark:text-gray-400">{analysis.purpose}</p>
          </div>
        )}

        {analysis.architecture && (
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
              Architecture
            </h3>
            <p className="text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{analysis.architecture}</p>
          </div>
        )}

        {analysis.summary && (
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
              Summary
            </h3>
            <p className="text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{analysis.summary}</p>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {analysis.strengths && (
            <div>
              <h3 className="text-lg font-medium text-green-700 dark:text-green-300 mb-2">
                Strengths
              </h3>
              <p className="text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{analysis.strengths}</p>
            </div>
          )}

          {analysis.weaknesses && (
            <div>
              <h3 className="text-lg font-medium text-red-700 dark:text-red-300 mb-2">
                Areas for Improvement
              </h3>
              <p className="text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{analysis.weaknesses}</p>
            </div>
          )}
        </div>

        {analysis.reusable_components && (
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
              Reusable Components
            </h3>
            <p className="text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{analysis.reusable_components}</p>
          </div>
        )}

        {analysis.notes && (
          <div>
            <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
              Additional Notes
            </h3>
            <p className="text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{analysis.notes}</p>
          </div>
        )}

        <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Analyzed by: <span className="font-medium">{analysis.analyzer}</span>
          </p>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            Git HEAD: <code className="text-xs bg-gray-100 dark:bg-gray-700 px-1 py-0.5 rounded">{analysis.analyzed_git_head.substring(0, 8)}</code>
          </p>
        </div>
      </div>
    </div>
  )
}