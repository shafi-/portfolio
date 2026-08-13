import { useState } from 'react'

interface FilterState {
  technologies: string[]
  frameworks: string[]
  repositoryType: string
}

interface FilterDropdownProps {
  filters: FilterState
  onChange: (filters: FilterState) => void
}

export default function FilterDropdown({ filters, onChange }: FilterDropdownProps) {
  const [isOpen, setIsOpen] = useState(false)

  const commonTechs = ['JavaScript', 'TypeScript', 'Python', 'Go', 'Rust', 'Java']
  const commonFrameworks = ['React', 'Vue', 'Angular', 'Node.js', 'Django', 'Spring']

  const toggleTech = (tech: string) => {
    const newTechs = filters.technologies.includes(tech)
      ? filters.technologies.filter(t => t !== tech)
      : [...filters.technologies, tech]
    onChange({ ...filters, technologies: newTechs })
  }

  const toggleFramework = (framework: string) => {
    const newFrameworks = filters.frameworks.includes(framework)
      ? filters.frameworks.filter(f => f !== framework)
      : [...filters.frameworks, framework]
    onChange({ ...filters, frameworks: newFrameworks })
  }

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white hover:bg-gray-50 dark:hover:bg-gray-600"
        aria-expanded={isOpen}
        aria-haspopup="true"
      >
        Filters {filters.technologies.length + filters.frameworks.length > 0 && `(${filters.technologies.length + filters.frameworks.length})`}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-lg z-10 p-4">
          <div className="mb-4">
            <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Technologies</h3>
            <div className="flex flex-wrap gap-2">
              {commonTechs.map(tech => (
                <button
                  key={tech}
                  onClick={() => toggleTech(tech)}
                  className={`px-3 py-1 rounded-full text-sm ${
                    filters.technologies.includes(tech)
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
                  }`}
                >
                  {tech}
                </button>
              ))}
            </div>
          </div>

          <div>
            <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Frameworks</h3>
            <div className="flex flex-wrap gap-2">
              {commonFrameworks.map(framework => (
                <button
                  key={framework}
                  onClick={() => toggleFramework(framework)}
                  className={`px-3 py-1 rounded-full text-sm ${
                    filters.frameworks.includes(framework)
                      ? 'bg-blue-600 text-white'
                      : 'bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white'
                  }`}
                >
                  {framework}
                </button>
              ))}
            </div>
          </div>

          {(filters.technologies.length > 0 || filters.frameworks.length > 0) && (
            <button
              onClick={() => onChange({ technologies: [], frameworks: [], repositoryType: 'all' })}
              className="mt-4 w-full px-3 py-2 text-sm text-red-600 hover:text-red-800 dark:hover:text-red-400"
            >
              Clear all filters
            </button>
          )}
        </div>
      )}
    </div>
  )
}