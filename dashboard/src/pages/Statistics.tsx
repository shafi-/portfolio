import { useState, useEffect } from 'react'
import { api } from '../services/api'
import { Statistics } from '../types'
import { Bar, Doughnut } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js/auto'
import ErrorState from '../components/ErrorState'

ChartJS.register(CategoryScale, LinearScale, BarElement, ArcElement, Title, Tooltip, Legend)

export default function StatisticsPage() {
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
          Portfolio Statistics
        </h1>
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">Loading statistics...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Portfolio Statistics
        </h1>
        <ErrorState error={error} onRetry={() => window.location.reload()} />
      </div>
    )
  }

  const techDistributionData = {
    labels: stats?.technologyDistribution.map(t => t.name) || [],
    datasets: [{
      label: 'Projects',
      data: stats?.technologyDistribution.map(t => t.count) || [],
      backgroundColor: 'rgba(59, 130, 246, 0.6)',
      borderColor: 'rgba(59, 130, 246, 1)',
      borderWidth: 1,
    }],
  }

  const maturityDistributionData = {
    labels: ['Low', 'Medium', 'High'],
    datasets: [{
      data: [
        stats?.maturityDistribution.low || 0,
        stats?.maturityDistribution.medium || 0,
        stats?.maturityDistribution.high || 0,
      ],
      backgroundColor: [
        'rgba(239, 68, 68, 0.6)',
        'rgba(234, 179, 8, 0.6)',
        'rgba(34, 197, 94, 0.6)',
      ],
      borderColor: [
        'rgba(239, 68, 68, 1)',
        'rgba(234, 179, 8, 1)',
        'rgba(34, 197, 94, 1)',
      ],
      borderWidth: 1,
    }],
  }

  const analysisCoverageData = {
    labels: ['Analyzed', 'Unanalyzed'],
    datasets: [{
      data: [
        stats?.analyzedProjectCount || 0,
        (stats?.totalProjects || 0) - (stats?.analyzedProjectCount || 0),
      ],
      backgroundColor: [
        'rgba(34, 197, 94, 0.6)',
        'rgba(156, 163, 175, 0.6)',
      ],
      borderColor: [
        'rgba(34, 197, 94, 1)',
        'rgba(156, 163, 175, 1)',
      ],
      borderWidth: 1,
    }],
  }

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        labels: {
          color: 'rgba(156, 163, 175, 1)',
        },
      },
      tooltip: {
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        titleColor: 'white',
        bodyColor: 'white',
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        ticks: {
          color: 'rgba(156, 163, 175, 1)',
        },
        grid: {
          color: 'rgba(156, 163, 175, 0.2)',
        },
      },
      x: {
        ticks: {
          color: 'rgba(156, 163, 175, 1)',
        },
        grid: {
          display: false,
        },
      },
    },
  }

  const doughnutOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom' as const,
        labels: {
          color: 'rgba(156, 163, 175, 1)',
        },
      },
      tooltip: {
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        titleColor: 'white',
        bodyColor: 'white',
      },
    },
  }

  return (
    <div className="space-y-8">
      <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
        Portfolio Statistics
      </h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Technology Distribution */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
            Technology Distribution
          </h2>
          <div style={{ height: '300px' }}>
            <Bar data={techDistributionData} options={chartOptions} />
          </div>
        </div>

        {/* Maturity Distribution */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
            Project Maturity
          </h2>
          <div style={{ height: '300px' }}>
            <Doughnut data={maturityDistributionData} options={doughnutOptions} />
          </div>
        </div>

        {/* Analysis Coverage */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
            Analysis Coverage
          </h2>
          <div style={{ height: '300px' }}>
            <Doughnut data={analysisCoverageData} options={doughnutOptions} />
          </div>
        </div>

        {/* Top Technologies */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4 text-gray-900 dark:text-white">
            Top Technologies
          </h2>
          <ul className="space-y-2">
            {stats?.topTechnologies.slice(0, 10).map((tech, index) => (
              <li key={index} className="flex justify-between items-center">
                <span className="text-gray-700 dark:text-gray-300">{tech.name}</span>
                <span className="font-semibold text-gray-900 dark:text-white">
                  {tech.count} projects
                </span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  )
}