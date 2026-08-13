import { NavLink } from 'react-router-dom'

export default function NavBar() {
  const navItems = [
    { path: '/', label: 'Overview', icon: '🏠' },
    { path: '/projects', label: 'Projects', icon: '📁' },
    { path: '/relationships', label: 'Relationships', icon: '🔗' },
    { path: '/statistics', label: 'Statistics', icon: '📊' },
  ]

  return (
    <nav className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700" role="navigation" aria-label="Main navigation">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center">
            <h1 className="text-xl font-bold text-gray-900 dark:text-white">
              Portfolio
            </h1>
          </div>
          <div className="flex space-x-4">
            {navItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) => `
                  px-3 py-2 rounded-md text-sm font-medium transition-colors
                  ${isActive
                    ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-200'
                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                  }
                `}
                aria-current={({ isActive }) => isActive ? 'page' : undefined}
              >
                <span className="mr-1" role="img" aria-hidden="true">{item.icon}</span>
                {item.label}
              </NavLink>
            ))}
          </div>
        </div>
      </div>
    </nav>
  )
}