interface StatCardProps {
  title: string
  value: number
  icon: string
}

export default function StatCard({ title, value, icon }: StatCardProps) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
            {title}
          </p>
          <p className="text-3xl font-bold text-gray-900 dark:text-white mt-2">
            {value.toLocaleString()}
          </p>
        </div>
        <div className="text-4xl" role="img" aria-label={title}>
          {icon}
        </div>
      </div>
    </div>
  )
}