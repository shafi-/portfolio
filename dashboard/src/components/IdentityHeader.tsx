import { Project } from '../types'

interface IdentityHeaderProps {
  project: Project | null
}

export default function IdentityHeader({ project }: IdentityHeaderProps) {
  if (!project) return null

  return (
    <div className="card animate-fade-in">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center space-x-4 mb-4">
            <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg shadow-blue-500/20">
              <span className="text-white font-bold text-2xl">
                {project.name.charAt(0).toUpperCase()}
              </span>
            </div>
            <div>
              <h1 className="text-4xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
                {project.name}
              </h1>
              <p className="text-slate-600 dark:text-slate-400 text-sm mt-1">
                {project.root_path}
              </p>
            </div>
          </div>
          <div className="flex flex-wrap gap-3">
            <span className="px-4 py-2 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-full text-sm font-medium border border-blue-200 dark:border-blue-800">
              📁 {project.repository_type}
            </span>
            <span className="px-4 py-2 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 rounded-full text-sm font-medium border border-green-200 dark:border-green-800">
              🕐 Discovered {new Date(project.discovered_at).toLocaleDateString()}
            </span>
            <span className="px-4 py-2 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded-full text-sm font-medium border border-purple-200 dark:border-purple-800">
              🔄 Updated {new Date(project.updated_at).toLocaleDateString()}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}