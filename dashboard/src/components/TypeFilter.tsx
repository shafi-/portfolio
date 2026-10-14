interface TypeFilterProps {
  value: string
  onChange: (value: string) => void
}

export default function TypeFilter({ value, onChange }: TypeFilterProps) {
  const relationshipTypes = [
    'all',
    'Similar',
    'Evolution',
    'Shared Feature',
    'Shared Technology',
    'Reuses Component',
  ]

  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="appearance-none w-full px-4 py-2.5 bg-white dark:bg-slate-800 text-slate-900 dark:text-white border-2 border-slate-200 dark:border-slate-700 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all duration-200 cursor-pointer hover:border-blue-300 dark:hover:border-blue-700 font-medium"
        aria-label="Filter by relationship type"
      >
        {relationshipTypes.map(type => (
          <option key={type} value={type}>
            {type === 'all' ? '🔍 All Types' : type}
          </option>
        ))}
      </select>
      <div className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none">
        <svg className="w-5 h-5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </div>
    </div>
  )
}