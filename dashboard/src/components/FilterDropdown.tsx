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

  const activeFilterCount = filters.technologies.length + filters.frameworks.length

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="btn btn-secondary inline-flex items-center space-x-2"
        aria-expanded={isOpen}
        aria-haspopup="true"
      >
        <span>🔧</span>
        <span>Filters</span>
        {activeFilterCount > 0 && (
          <span className="bg-blue-500 text-white text-xs px-2 py-0.5 rounded-full">
            {activeFilterCount}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-80 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl shadow-xl z-10 animate-fade-in">
          <div className="p-4">
            <div className="mb-4">
              <h3 className="font-semibold text-slate-900 dark:text-white mb-3 flex items-center space-x-2">
                <span>⚙️</span>
                <span>Technologies</span>
              </h3>
              <div className="flex flex-wrap gap-2">
                {commonTechs.map(tech => (
                  <button
                    key={tech}
                    onClick={() => toggleTech(tech)}
                    className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-all ${
                      filters.technologies.includes(tech)
                        ? 'bg-blue-500 text-white shadow-md shadow-blue-500/20'
                        : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600'
                    }`}
                  >
                    {tech}
                  </button>
                ))}
              </div>
            </div>

            <div className="mb-4">
              <h3 className="font-semibold text-slate-900 dark:text-white mb-3 flex items-center space-x-2">
                <span>🏗️</span>
                <span>Frameworks</span>
              </h3>
              <div className="flex flex-wrap gap-2">
                {commonFrameworks.map(framework => (
                  <button
                    key={framework}
                    onClick={() => toggleFramework(framework)}
                    className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-all ${
                      filters.frameworks.includes(framework)
                        ? 'bg-blue-500 text-white shadow-md shadow-blue-500/20'
                        : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-600'
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
                className="w-full px-3 py-2 text-sm text-red-600 hover:text-red-800 dark:hover:text-red-400 border border-red-200 dark:border-red-800 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
              >
                ✕ Clear all filters
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}