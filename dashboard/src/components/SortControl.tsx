type SortOption = 'name' | 'date' | 'size'

interface SortControlProps {
  value: SortOption
  onChange: (value: SortOption) => void
}

export default function SortControl({ value, onChange }: SortControlProps) {
  const options = [
    { value: 'name' as SortOption, label: 'Name' },
    { value: 'date' as SortOption, label: 'Recent' },
    { value: 'size' as SortOption, label: 'Size' },
  ]

  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value as SortOption)}
      className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
      aria-label="Sort projects"
    >
      {options.map(option => (
        <option key={option.value} value={option.value}>
          Sort by {option.label}
        </option>
      ))}
    </select>
  )
}