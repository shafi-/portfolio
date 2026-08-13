import { Outlet } from 'react-router-dom'
import NavBar from './NavBar'
import Footer from './Footer'
import SkipNav from './SkipNav'

export default function Layout() {
  return (
    <div className="min-h-screen flex flex-col">
      <SkipNav />
      <NavBar />
      <main id="main-content" className="flex-1 container mx-auto px-4 py-12 max-w-7xl" tabIndex={-1}>
        <div className="animate-fade-in">
          <Outlet />
        </div>
      </main>
      <Footer />
    </div>
  )
}