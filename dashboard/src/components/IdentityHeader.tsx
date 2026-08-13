import { Project } from '../types'

interface IdentityHeaderProps {
  project: Project | null
}

export default function IdentityHeader({ project }: IdentityHeaderProps) {
  if (!project) return null

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
            {project.name}
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            {project.root_path}
          </p>
          <div className="flex flex-wrap gap-2">
            <span className="px-3 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded-full text-sm">
              {project.repository_type}
            </span>
            <span className="px-3 py-1 bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 rounded-full text-sm">
              Discovered {new Date(project.discovered_at).toLocaleDateString()}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}