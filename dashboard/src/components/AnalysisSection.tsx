import { Analysis } from '../types'

interface AnalysisSectionProps {
  analysis: Analysis | null
  loading: boolean
  error: Error | undefined
}

export default function AnalysisSection({ analysis, loading, error }: AnalysisSectionProps) {
  if (loading) {
    return (
      <div className="card animate-fade-in">
        <div className="flex items-center space-x-3 mb-6">
          <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-pink-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">🤖</span>
          </div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
            AI Analysis
          </h2>
        </div>
        <div className="text-center py-12">
          <div className="inline-block relative">
            <div className="w-12 h-12 border-4 border-purple-200 border-t-purple-600 rounded-full animate-spin"></div>
            <div className="absolute inset-0 w-12 h-12 border-4 border-pink-200 border-r-pink-600 rounded-full animate-spin" style={{ animationDuration: '1.5s' }}></div>
          </div>
          <p className="mt-4 text-slate-600 dark:text-slate-400 font-medium">Loading AI analysis...</p>
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
            AI Analysis
          </h2>
        </div>
        <div className="text-center py-8">
          <p className="text-red-600 dark:text-red-400 font-medium">Failed to load analysis</p>
        </div>
      </div>
    )
  }

  if (!analysis) {
    return (
      <div className="card animate-fade-in bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20">
        <div className="text-center py-12">
          <div className="w-20 h-20 bg-gradient-to-br from-purple-100 to-pink-100 dark:from-purple-900/40 dark:to-pink-900/40 rounded-2xl flex items-center justify-center mx-auto mb-6 shadow-lg shadow-purple-500/10">
            <span className="text-5xl">🤖</span>
          </div>
          <h2 className="text-2xl font-bold text-slate-900 dark:text-white mb-3">
            No AI Analysis Available
          </h2>
          <p className="text-slate-600 dark:text-slate-400 mb-6 max-w-lg mx-auto">
            This project hasn't been analyzed by an AI agent yet. Analysis provides deeper insights into architecture, patterns, and relationships.
          </p>
          <div className="inline-flex items-center space-x-2 px-4 py-2 bg-purple-100 dark:bg-purple-900/30 rounded-full">
            <span className="text-purple-700 dark:text-purple-300 text-sm font-medium">
              💡 Analysis requires an AI agent integration with Portfolio
            </span>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="card animate-fade-in">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-pink-600 rounded-lg flex items-center justify-center">
            <span className="text-white text-lg">🤖</span>
          </div>
          <div>
            <h2 className="text-2xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
              AI Analysis
            </h2>
            <p className="text-sm text-slate-600 dark:text-slate-400 flex items-center">
              <span className="mr-1">🕐</span>
              Analyzed {new Date(analysis.analyzed_at).toLocaleDateString()}
            </p>
          </div>
        </div>
      </div>

      <div className="space-y-6">
        {analysis.purpose && (
          <div className="p-5 bg-gradient-to-br from-blue-50 to-cyan-50 dark:from-blue-900/20 dark:to-cyan-900/20 rounded-xl">
            <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-3 flex items-center">
              <span className="w-6 h-6 bg-gradient-to-br from-blue-500 to-cyan-500 rounded-md flex items-center justify-center mr-2">
                <span className="text-white text-xs">🎯</span>
              </span>
              Purpose
            </h3>
            <p className="text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">{analysis.purpose}</p>
          </div>
        )}

        {analysis.architecture && (
          <div className="p-5 bg-gradient-to-br from-purple-50 to-pink-50 dark:from-purple-900/20 dark:to-pink-900/20 rounded-xl">
            <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-3 flex items-center">
              <span className="w-6 h-6 bg-gradient-to-br from-purple-500 to-pink-500 rounded-md flex items-center justify-center mr-2">
                <span className="text-white text-xs">🏗️</span>
              </span>
              Architecture
            </h3>
            <p className="text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">{analysis.architecture}</p>
          </div>
        )}

        {analysis.summary && (
          <div className="p-5 bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-800 dark:to-slate-700 rounded-xl">
            <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-3 flex items-center">
              <span className="w-6 h-6 bg-gradient-to-br from-slate-500 to-slate-600 rounded-md flex items-center justify-center mr-2">
                <span className="text-white text-xs">📋</span>
              </span>
              Summary
            </h3>
            <p className="text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">{analysis.summary}</p>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {analysis.strengths && (
            <div className="p-5 bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-900/20 dark:to-emerald-900/20 rounded-xl">
              <h3 className="text-lg font-semibold text-green-700 dark:text-green-300 mb-3 flex items-center">
                <span className="w-6 h-6 bg-gradient-to-br from-green-500 to-emerald-500 rounded-md flex items-center justify-center mr-2">
                  <span className="text-white text-xs">💪</span>
                </span>
                Strengths
              </h3>
              <p className="text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">{analysis.strengths}</p>
            </div>
          )}

          {analysis.weaknesses && (
            <div className="p-5 bg-gradient-to-br from-orange-50 to-red-50 dark:from-orange-900/20 dark:to-red-900/20 rounded-xl">
              <h3 className="text-lg font-semibold text-orange-700 dark:text-orange-300 mb-3 flex items-center">
                <span className="w-6 h-6 bg-gradient-to-br from-orange-500 to-red-500 rounded-md flex items-center justify-center mr-2">
                  <span className="text-white text-xs">🎯</span>
                </span>
                Areas for Improvement
              </h3>
              <p className="text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">{analysis.weaknesses}</p>
            </div>
          )}
        </div>

        {analysis.reusable_components && (
          <div className="p-5 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl">
            <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-3 flex items-center">
              <span className="w-6 h-6 bg-gradient-to-br from-blue-500 to-indigo-500 rounded-md flex items-center justify-center mr-2">
                <span className="text-white text-xs">🔄</span>
              </span>
              Reusable Components
            </h3>
            <p className="text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">{analysis.reusable_components}</p>
          </div>
        )}

        {analysis.notes && (
          <div className="p-5 bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-800 dark:to-slate-700 rounded-xl">
            <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-3 flex items-center">
              <span className="w-6 h-6 bg-gradient-to-br from-slate-500 to-slate-600 rounded-md flex items-center justify-center mr-2">
                <span className="text-white text-xs">📝</span>
              </span>
              Additional Notes
            </h3>
            <p className="text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">{analysis.notes}</p>
          </div>
        )}

        <div className="flex items-center justify-between pt-4 border-t border-slate-200 dark:border-slate-700">
          <div className="flex items-center space-x-4">
            <p className="text-sm text-slate-600 dark:text-slate-400 flex items-center">
              <span className="mr-2">🔬</span>
              Analyzed by: <span className="font-semibold text-slate-900 dark:text-white ml-1">{analysis.analyzer}</span>
            </p>
          </div>
          <p className="text-sm text-slate-600 dark:text-slate-400 flex items-center">
            <span className="mr-2">📌</span>
            Git HEAD: <code className="ml-1 text-xs bg-slate-200 dark:bg-slate-700 px-2 py-1 rounded font-mono">{analysis.analyzed_git_head.substring(0, 8)}</code>
          </p>
        </div>
      </div>
    </div>
  )
}