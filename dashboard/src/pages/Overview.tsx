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
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Portfolio Overview
        </h1>
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">Loading portfolio statistics...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Portfolio Overview
        </h1>
        <div className="text-center py-12">
          <h2 className="text-xl font-semibold text-red-600 mb-4">Error loading statistics</h2>
          <p className="text-gray-600 mb-4">{error.message}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
        Portfolio Overview
      </h1>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <StatCard
          title="Total Projects"
          value={stats?.totalProjects || 0}
          icon="📁"
        />
        <StatCard
          title="Active Projects"
          value={stats?.activeProjects || 0}
          icon="🚀"
        />
        <StatCard
          title="Technologies"
          value={stats?.uniqueTechnologies || 0}
          icon="⚙️"
        />
      </div>

      {/* Technology Distribution Chart */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Technology Distribution
        </h2>
        <TechBarChart data={stats?.technologyDistribution || []} />
      </div>

      {/* Recent Activity Timeline */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
          Recent Activity
        </h2>
        <ActivityTimeline activities={stats?.recentActivity || []} />
      </div>
    </div>
  )
}