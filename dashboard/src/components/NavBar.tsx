import { NavLink } from 'react-router-dom'

export default function NavBar() {
  const navItems = [
    { path: '/', label: 'Overview', icon: '🏠' },
    { path: '/projects', label: 'Projects', icon: '📁' },
    { path: '/relationships', label: 'Relationships', icon: '🔗' },
    { path: '/statistics', label: 'Statistics', icon: '📊' },
  ]

  return (
    <nav className="bg-white/80 dark:bg-slate-900/80 backdrop-blur-md border-b border-slate-200 dark:border-slate-700 sticky top-0 z-50" role="navigation" aria-label="Main navigation">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">P</span>
            </div>
            <h1 className="text-xl font-bold bg-gradient-to-r from-slate-900 to-slate-700 dark:from-slate-100 dark:to-slate-300 bg-clip-text text-transparent">
              Portfolio
            </h1>
          </div>
          <div className="flex space-x-2">
            {navItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) => `
                  px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200
                  flex items-center space-x-2
                  ${isActive
                    ? 'bg-blue-500 text-white shadow-md shadow-blue-500/20'
                    : 'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'
                  }
                `}
                aria-current={({ isActive }) => isActive ? 'page' : undefined}
              >
                <span className="text-base" role="img" aria-hidden="true">{item.icon}</span>
                <span>{item.label}</span>
              </NavLink>
            ))}
          </div>
        </div>
      </div>
    </nav>
  )
}