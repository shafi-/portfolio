import { createHashRouter, RouterProvider } from 'react-router-dom'
import { StrictMode } from 'react'
import './index.css'

// Page components
import Overview from './pages/Overview'
import ProjectList from './pages/ProjectList'
import ProjectDetail from './pages/ProjectDetail'
import RelationshipExplorer from './pages/RelationshipExplorer'
import StatisticsPage from './pages/Statistics'
import Layout from './components/Layout'

const router = createHashRouter([
  {
    path: "/",
    element: <Layout />,
    children: [
      { index: true, element: <Overview /> },
      { path: "projects", element: <ProjectList /> },
      { path: "projects/:id", element: <ProjectDetail /> },
      { path: "relationships", element: <RelationshipExplorer /> },
      { path: "statistics", element: <StatisticsPage /> },
    ],
  },
])

export default function App() {
  return (
    <StrictMode>
      <RouterProvider router={router} />
    </StrictMode>
  )
}