import { useEffect, useState } from 'react'
import { api } from '../services/api'
import { Statistics } from '../types'
import StatCard from '../components/StatCard'
import TechBarChart from '../components/TechBarChart'
import ActivityTimeline from '../components/ActivityTimeline'

export default function Overview() {
  const [stats, setStats] = useState<Statistics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    async function loadStatistics() {
      try {
        setLoading(true)
        const statisticsData = await api.getStatistics()
        setStats(statisticsData)
      } catch (err) {
        setError(err as Error)
      } finally {
        setLoading(false)
      }
    }

    loadStatistics()
  }, [])

  if (loading) {
    return (
      <div className="space-y-8">
        <div className="flex items-center space-x-4">
          <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center">
            <span className="text-white font-bold text-lg">P</span>
          </div>
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
              Portfolio Overview
            </h1>
            <p className="text-slate-600 dark:text-slate-400 text-sm">Your development portfolio at a glance</p>
          </div>
        </div>
        <div className="text-center py-16">
          <div className="inline-block relative">
            <div className="w-16 h-16 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
            <div className="absolute inset-0 w-16 h-16 border-4 border-purple-200 border-r-purple-600 rounded-full animate-spin" style={{ animationDuration: '1.5s' }}></div>
          </div>
          <p className="mt-6 text-slate-600 dark:text-slate-400 font-medium">Loading portfolio statistics...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-8">
        <div className="flex items-center space-x-4">
          <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center">
            <span className="text-white font-bold text-lg">P</span>
          </div>
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
              Portfolio Overview
            </h1>
            <p className="text-slate-600 dark:text-slate-400 text-sm">Your development portfolio at a glance</p>
          </div>
        </div>
        <div className="text-center py-16">
          <div className="w-16 h-16 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center mx-auto mb-4">
            <span className="text-3xl">⚠️</span>
          </div>
          <h2 className="text-xl font-semibold text-red-600 dark:text-red-400 mb-2">Unable to load statistics</h2>
          <p className="text-slate-600 dark:text-slate-400 mb-6 max-w-md mx-auto">{error.message}</p>
          <button
            onClick={() => window.location.reload()}
            className="btn btn-primary"
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg shadow-blue-500/20">
            <span className="text-white font-bold text-lg">P</span>
          </div>
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
              Portfolio Overview
            </h1>
            <p className="text-slate-600 dark:text-slate-400 text-sm">Your development portfolio at a glance</p>
          </div>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <StatCard
          title="Total Projects"
          value={stats?.totalProjects || 0}
          icon="📁"
          trend="up"
        />
        <StatCard
          title="Active Projects"
          value={stats?.activeProjects || 0}
          icon="🚀"
          trend="up"
        />
        <StatCard
          title="Technologies"
          value={stats?.uniqueTechnologies || 0}
          icon="⚙️"
          trend="neutral"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Technology Distribution Chart */}
        <div className="card">
          <div className="card-header">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center space-x-2">
              <span>📊</span>
              <span>Technology Distribution</span>
            </h2>
          </div>
          <div className="card-body">
            <TechBarChart data={stats?.technologyDistribution || []} />
          </div>
        </div>

        {/* Recent Activity Timeline */}
        <div className="card">
          <div className="card-header">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center space-x-2">
              <span>🕐</span>
              <span>Recent Activity</span>
            </h2>
          </div>
          <div className="card-body">
            <ActivityTimeline activities={stats?.recentActivity || []} />
          </div>
        </div>
      </div>
    </div>
  )
}