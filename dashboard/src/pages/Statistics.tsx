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
      <div className="text-center py-16">
        <div className="inline-block relative">
          <div className="w-16 h-16 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin"></div>
          <div className="absolute inset-0 w-16 h-16 border-4 border-purple-200 border-r-purple-600 rounded-full animate-spin" style={{ animationDuration: '1.5s' }}></div>
        </div>
        <p className="mt-6 text-slate-600 dark:text-slate-400 font-medium">Loading statistics...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div>
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
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {/* Technology Distribution */}
        <div className="card">
          <div className="card-header">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center space-x-2">
              <span>⚙️</span>
              <span>Technology Distribution</span>
            </h2>
          </div>
          <div className="card-body">
            <div style={{ height: '300px' }}>
              <Bar data={techDistributionData} options={chartOptions} />
            </div>
          </div>
        </div>

        {/* Maturity Distribution */}
        <div className="card">
          <div className="card-header">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center space-x-2">
              <span>📈</span>
              <span>Project Maturity</span>
            </h2>
          </div>
          <div className="card-body">
            <div style={{ height: '300px' }}>
              <Doughnut data={maturityDistributionData} options={doughnutOptions} />
            </div>
          </div>
        </div>

        {/* Analysis Coverage */}
        <div className="card">
          <div className="card-header">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center space-x-2">
              <span>🔍</span>
              <span>Analysis Coverage</span>
            </h2>
          </div>
          <div className="card-body">
            <div style={{ height: '300px' }}>
              <Doughnut data={analysisCoverageData} options={doughnutOptions} />
            </div>
          </div>
        </div>

        {/* Top Technologies */}
        <div className="card">
          <div className="card-header">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center space-x-2">
              <span>🏆</span>
              <span>Top Technologies</span>
            </h2>
          </div>
          <div className="card-body">
            <ul className="space-y-3">
              {stats?.topTechnologies.slice(0, 10).map((tech, index) => (
                <li key={index} className="flex justify-between items-center group">
                  <div className="flex items-center space-x-3">
                    <div className="w-8 h-8 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-slate-700 dark:to-slate-600 rounded-lg flex items-center justify-center">
                      <span className="text-sm font-bold text-blue-600 dark:text-blue-400">
                        {index + 1}
                      </span>
                    </div>
                    <span className="text-slate-700 dark:text-slate-300 font-medium">{tech.name}</span>
                  </div>
                  <span className="font-semibold text-slate-900 dark:text-white bg-slate-100 dark:bg-slate-700 px-3 py-1 rounded-full text-sm">
                    {tech.count} projects
                  </span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}