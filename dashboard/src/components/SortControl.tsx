type SortOption = 'name' | 'date' | 'size'

interface SortControlProps {
  value: SortOption
  onChange: (value: SortOption) => void
}

export default function SortControl({ value, onChange }: SortControlProps) {
  const options = [
    { value: 'name' as SortOption, label: '📝 Name', icon: '📝' },
    { value: 'date' as SortOption, label: '🕐 Recent', icon: '🕐' },
    { value: 'size' as SortOption, label: '📊 Size', icon: '📊' },
  ]

  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value as SortOption)}
        className="appearance-none w-full px-4 py-2.5 bg-white dark:bg-slate-800 text-slate-900 dark:text-white border-2 border-slate-200 dark:border-slate-700 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 cursor-pointer hover:border-blue-300 dark:hover:border-blue-700 font-medium"
        aria-label="Sort projects"
      >
        {options.map(option => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <div className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none">
        <svg className="w-5 h-5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 4h13M3 8h9m-9 4h6m4 0l4-4m0 0l4 4m-4-4v12" />
        </svg>
      </div>
    </div>
  )
}