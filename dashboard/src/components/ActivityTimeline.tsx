import { Activity } from '../types'
import { useNavigate } from 'react-router-dom'

interface ActivityTimelineProps {
  activities: Activity[]
}

export default function ActivityTimeline({ activities }: ActivityTimelineProps) {
  const navigate = useNavigate()

  if (activities.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">
        No recent activity to display
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {activities.map((activity, index) => (
        <div key={index} className="flex items-start space-x-4">
          <div className="flex flex-col items-center">
            <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
            {index < activities.length - 1 && (
              <div className="w-0.5 h-full bg-gray-300 dark:bg-gray-600 mt-1"></div>
            )}
          </div>
          <div className="flex-1 pb-4">
            <button
              onClick={() => navigate(`/projects/${activity.project_id}`)}
              className="text-sm font-medium text-blue-600 dark:text-blue-400 hover:underline"
            >
              {activity.project_name}
            </button>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
              {activity.activity_type}
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
              {new Date(activity.timestamp).toLocaleDateString()}
            </p>
          </div>
        </div>
      ))}
    </div>
  )
}