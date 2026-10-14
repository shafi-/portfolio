export default function Footer() {
  return (
    <footer className="bg-white/50 dark:bg-slate-900/50 backdrop-blur-sm border-t border-slate-200 dark:border-slate-700 mt-auto">
      <div className="container mx-auto px-4 py-6">
        <div className="flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-xs">P</span>
            </div>
            <div>
              <p className="text-sm font-medium text-slate-900 dark:text-slate-100">
                Portfolio Dashboard
              </p>
              <p className="text-xs text-slate-600 dark:text-slate-400">
                Local-first Project Inventory
              </p>
            </div>
          </div>
          <div className="text-center md:text-right">
            <p className="text-xs text-slate-600 dark:text-slate-400">
              Read-only visualization • {new Date().getFullYear()}
            </p>
          </div>
        </div>
      </div>
    </footer>
  )
}