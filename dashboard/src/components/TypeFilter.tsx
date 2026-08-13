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
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
      aria-label="Filter by relationship type"
    >
      {relationshipTypes.map(type => (
        <option key={type} value={type}>
          {type === 'all' ? 'All Types' : type}
        </option>
      ))}
    </select>
  )
}