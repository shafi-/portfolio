interface StatCardProps {
  title: string
  value: number
  icon: string
  trend?: 'up' | 'down' | 'neutral'
}

export default function StatCard({ title, value, icon, trend = 'neutral' }: StatCardProps) {
  const trendColors = {
    up: 'text-green-600 dark:text-green-400',
    down: 'text-red-600 dark:text-red-400',
    neutral: 'text-slate-600 dark:text-slate-400'
  }

  const getTrendIcon = () => {
    switch (trend) {
      case 'up': return '↗'
      case 'down': return '↘'
      default: return '→'
    }
  }

  return (
    <div className="card group animate-fade-in">
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-slate-600 dark:text-slate-400 uppercase tracking-wide">
            {title}
          </p>
          <p className="text-4xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent mt-2 group-hover:scale-105 transition-transform duration-200">
            {value.toLocaleString()}
          </p>
        </div>
        <div className="relative">
          <div className="w-16 h-16 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-slate-700 dark:to-slate-600 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-200">
            <span className="text-3xl" role="img" aria-label={title}>
              {icon}
            </span>
          </div>
          <div className="absolute -top-1 -right-1 w-6 h-6 bg-white dark:bg-slate-700 rounded-full flex items-center justify-center shadow-sm">
            <span className={`text-xs ${trendColors[trend]}`}>{getTrendIcon()}</span>
          </div>
        </div>
      </div>
    </div>
  )
}