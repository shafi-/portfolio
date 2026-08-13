import { Outlet } from 'react-router-dom'
import NavBar from './NavBar'
import Footer from './Footer'
import SkipNav from './SkipNav'

export default function Layout() {
  return (
    <div className="min-h-screen flex flex-col">
      <SkipNav />
      <NavBar />
      <main id="main-content" className="flex-1 container mx-auto px-4 py-8" tabIndex={-1}>
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}