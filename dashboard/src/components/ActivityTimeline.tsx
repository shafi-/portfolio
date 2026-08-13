import { Activity } from '../types'
import { useNavigate } from 'react-router-dom'

interface ActivityTimelineProps {
  activities: Activity[]
}

export default function ActivityTimeline({ activities }: ActivityTimelineProps) {
  const navigate = useNavigate()

  if (activities.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="w-16 h-16 bg-gradient-to-br from-blue-50 to-purple-50 dark:from-slate-700 dark:to-slate-600 rounded-full flex items-center justify-center mx-auto mb-4">
          <span className="text-3xl">🕐</span>
        </div>
        <p className="text-slate-600 dark:text-slate-400 font-medium">No recent activity to display</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {activities.map((activity, index) => (
        <div key={index} className="flex items-start space-x-4 group">
          <div className="flex flex-col items-center pt-1">
            <div className="w-3 h-3 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full ring-4 ring-white dark:ring-slate-900 group-hover:scale-125 transition-transform duration-200"></div>
            {index < activities.length - 1 && (
              <div className="w-0.5 h-16 bg-gradient-to-b from-blue-200 to-purple-200 dark:from-blue-800 dark:to-purple-800 mt-2"></div>
            )}
          </div>
          <div className="flex-1 pb-6">
            <button
              onClick={() => navigate(`/projects/${activity.project_id}`)}
              className="text-sm font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 transition-colors"
            >
              {activity.project_name}
            </button>
            <p className="text-sm text-slate-700 dark:text-slate-300 mt-1.5">
              {activity.activity_type}
            </p>
            <p className="text-xs text-slate-500 dark:text-slate-500 mt-1 flex items-center space-x-1">
              <span>📅</span>
              <span>{new Date(activity.timestamp).toLocaleDateString()}</span>
            </p>
          </div>
        </div>
      ))}
    </div>
  )
}