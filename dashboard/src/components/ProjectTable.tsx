import { Project } from '../types'
import { parseLanguageSummary } from '../utils/parsers'

interface ProjectTableProps {
  projects: Project[]
  onProjectClick: (project: Project) => void
}

export default function ProjectTable({ projects, onProjectClick }: ProjectTableProps) {
  if (projects.length === 0) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-500 dark:text-gray-400">No projects found</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead className="bg-gray-50 dark:bg-gray-800">
          <tr>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Name
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Technologies
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Last Updated
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
              Type
            </th>
          </tr>
        </thead>
        <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
          {projects.map((project) => {
            const languages = parseLanguageSummary(project.metadata.language_summary)
            return (
              <tr
                key={project.id}
                onClick={() => onProjectClick(project)}
                className="hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
              >
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm font-medium text-blue-600 dark:text-blue-400">
                    {project.name}
                  </div>
                  <div className="text-sm text-gray-500 dark:text-gray-400">
                    {project.root_path}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex flex-wrap gap-1">
                    {languages.slice(0, 3).map((lang) => (
                      <span
                        key={lang}
                        className="px-2 py-1 text-xs bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 rounded"
                      >
                        {lang}
                      </span>
                    ))}
                    {languages.length > 3 && (
                      <span className="text-xs text-gray-500 dark:text-gray-400">
                        +{languages.length - 3}
                      </span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {new Date(project.metadata.last_commit_at).toLocaleDateString()}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {project.repository_type}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}