import { Project } from '../types'
import { parseLanguageSummary } from '../utils/parsers'

interface ProjectTableProps {
  projects: Project[]
  onProjectClick: (project: Project) => void
}

export default function ProjectTable({ projects, onProjectClick }: ProjectTableProps) {
  if (projects.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="w-16 h-16 bg-slate-100 dark:bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-4">
          <span className="text-3xl">📁</span>
        </div>
        <p className="text-slate-600 dark:text-slate-400 font-medium">No projects found</p>
      </div>
    )
  }

  return (
    <div className="card overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-200 dark:divide-slate-700">
          <thead className="bg-slate-50 dark:bg-slate-800/50">
            <tr>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                Name
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                Technologies
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                Last Updated
              </th>
              <th scope="col" className="px-6 py-4 text-left text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                Type
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-slate-900 divide-y divide-slate-200 dark:divide-slate-700">
            {projects.map((project, index) => {
              const languages = parseLanguageSummary(project.metadata.language_summary)
              return (
                <tr
                  key={project.id}
                  onClick={() => onProjectClick(project)}
                  className="hover:bg-blue-50 dark:hover:bg-blue-900/10 cursor-pointer transition-colors duration-150 animate-fade-in"
                  style={{ animationDelay: `${index * 50}ms` }}
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center space-x-3">
                      <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center flex-shrink-0">
                        <span className="text-white font-bold text-sm">
                          {project.name.charAt(0).toUpperCase()}
                        </span>
                      </div>
                      <div>
                        <div className="text-sm font-semibold text-blue-600 dark:text-blue-400">
                          {project.name}
                        </div>
                        <div className="text-xs text-slate-600 dark:text-slate-400 max-w-xs truncate">
                          {project.root_path}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex flex-wrap gap-2">
                      {languages.slice(0, 3).map((lang) => (
                        <span
                          key={lang}
                          className="px-2 py-1 text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-md border border-blue-200 dark:border-blue-800"
                        >
                          {lang}
                        </span>
                      ))}
                      {languages.length > 3 && (
                        <span className="text-xs text-slate-500 dark:text-slate-400 font-medium">
                          +{languages.length - 3} more
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center space-x-2">
                      <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                      <span className="text-sm text-slate-600 dark:text-slate-400">
                        {new Date(project.metadata.last_commit_at).toLocaleDateString()}
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="px-2 py-1 text-xs font-medium bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 rounded-md border border-slate-200 dark:border-slate-600">
                      {project.repository_type}
                    </span>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}