interface SearchBarProps {
  value: string
  onChange: (value: string) => void
  placeholder: string
}

export default function SearchBar({ value, onChange, placeholder }: SearchBarProps) {
  return (
    <div className="relative flex-1 max-w-md">
      <div className="relative group">
        {/* Search icon */}
        <div className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
          <span className="text-sm">🔍</span>
        </div>

        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className="w-full pl-10 pr-10 py-2.5 border border-slate-300 dark:border-slate-600 rounded-lg
                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent
                   bg-white dark:bg-slate-700 text-slate-900 dark:text-white
                   placeholder:text-slate-400 dark:placeholder:text-slate-500
                   transition-all duration-200 shadow-sm group-focus-within:shadow-md"
          aria-label="Search"
        />

        {/* Clear button */}
        {value && (
          <button
            onClick={() => onChange('')}
            className="absolute right-3 top-1/2 transform -translate-y-1/2
                     text-slate-400 hover:text-slate-600 dark:hover:text-slate-300
                     transition-colors p-1 rounded-full hover:bg-slate-100 dark:hover:bg-slate-600"
            aria-label="Clear search"
          >
            <span className="text-sm">✕</span>
          </button>
        )}
      </div>
    </div>
  )
}