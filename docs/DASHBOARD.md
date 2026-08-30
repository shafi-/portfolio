# Portfolio Dashboard

Complete web-based dashboard for exploring your software portfolio.

## Features

The dashboard provides a modern, responsive interface for:

- **Portfolio Overview**: Statistics, recent activity, and top technologies
- **Project List**: Searchable, filterable list of all projects
- **Project Details**: Comprehensive view of individual projects with metadata, documentation, and AI analysis
- **Relationship Explorer**: Interactive graph visualization of project relationships
- **Statistics**: Detailed analytics and charts about your portfolio

## Quick Start

### Development Mode

For development with hot reload:

```bash
# Start both Vite dev server and Go backend
./scripts/dev-dashboard.sh
```

This starts:
- Vite dev server on `http://localhost:5173` (with hot reload)
- Portfolio dashboard server on `http://localhost:3000`

The dashboard automatically proxies API requests to the backend.

### Production Build

For production deployment:

```bash
# Build dashboard and create single binary
./scripts/build-dashboard.sh

# Run dashboard with embedded assets
./portfolio dashboard
```

## Development Workflow

### Frontend Development

The dashboard is built with React + TypeScript + Vite:

```bash
cd dashboard

# Install dependencies
npm install

# Start dev server (with hot reload)
npm run dev

# Type checking
npm run typecheck

# Production build
npm run build

# Preview production build
npm run preview
```

### Backend Development

The Go backend serves both API endpoints and dashboard assets:

```bash
# Development mode (serves from filesystem)
./portfolio dashboard --dev --dist ./dashboard/dist

# Production mode (serves embedded assets)
./portfolio dashboard
```

## Architecture

### Frontend Stack

- **Framework**: React 19 with TypeScript
- **Build**: Vite 8 with hot module replacement
- **Routing**: Hash-based routing (React Router)
- **Styling**: Tailwind CSS v4 + shadcn/ui components
- **Charts**: Chart.js for data visualization
- **Graphs**: Cytoscape.js for relationship visualization
- **Icons**: Lucide React

### Backend Integration

- **Asset Serving**: Dual-mode server (embedded/external)
- **API Proxy**: CORS-enabled API routing
- **Single Binary**: Go embed directive for production deployment

### Key Design Decisions

1. **Hash Routing**: Enables deployment without server-side routing support
2. **Progressive Enhancement**: Core features work without AI analysis
3. **System Preference Theme**: Automatically detects dark/light mode preference
4. **Accessibility First**: WCAG 2.1 AA compliance with semantic HTML and ARIA labels
5. **Desktop-First**: Optimized for desktop use with responsive mobile support

## Pages and Components

### Overview (`src/pages/Overview.tsx`)

Portfolio-level statistics and activity:

- **StatCard**: Key metrics (total projects, analyzed projects, technologies)
- **TechBarChart**: Top 10 technologies by project count
- **ActivityTimeline**: Recent activity across projects

### Project List (`src/pages/ProjectList.tsx`)

Searchable and filterable project listing:

- **SearchBar**: Real-time search with debouncing
- **FilterPanel**: Technology and framework filtering
- **ProjectTable**: Paginated project list with sorting
- **Pagination**: Efficient navigation of large datasets

### Project Detail (`src/pages/ProjectDetail.tsx`)

Comprehensive single-project view:

1. **Basic Information**: Name, path, description, repository type
2. **Technology Stack**: Languages, frameworks, dependencies
3. **Documentation**: Index of available documentation
4. **Analysis**: AI-generated insights (progressive enhancement)
5. **Relationships**: Related projects and dependencies

### Relationship Explorer (`src/pages/RelationshipExplorer.tsx`)

Interactive relationship visualization:

- **GraphView**: Cytoscape.js force-directed graph
- **ListView**: Alternative list-based view
- **TypeFilter**: Filter by relationship type
- **DetailView**: Selected relationship details

### Statistics (`src/pages/Statistics.tsx`)

Portfolio analytics and charts:

- **Technology Distribution**: Bar chart of tech usage
- **Maturity Distribution**: Doughnut chart of project maturity
- **Analysis Coverage**: Doughnut chart of AI analysis status
- **Top Technologies**: Ranked list of most-used technologies

## API Integration

The frontend communicates with the Portfolio API via the `ApiClient` service (`src/services/api.ts`):

- **Caching**: Built-in response caching to reduce API calls
- **Error Handling**: Graceful error states with retry options
- **Type Safety**: Full TypeScript interfaces for all API responses
- **Base URL**: Automatic detection of API base URL

### API Endpoints Used

- `GET /projects` - List all projects
- `GET /projects/{id}` - Get project details
- `GET /projects/{id}/analysis` - Get AI analysis
- `GET /search` - Search projects
- `GET /relationships/{id}` - Get project relationships
- `GET /statistics` - Get portfolio statistics
- `GET /configuration` - Get system configuration

## Accessibility

The dashboard follows WCAG 2.1 AA guidelines:

- **Semantic HTML**: Proper heading hierarchy and landmark regions
- **ARIA Labels**: Descriptive labels for interactive elements
- **Keyboard Navigation**: Full keyboard accessibility
- **Color Contrast**: Minimum 4.5:1 contrast ratio
- **Focus Indicators**: Clear focus states for interactive elements
- **Error Messages**: Accessible error notifications

## Performance Optimization

### Frontend

- **Code Splitting**: Separate chunks for Chart.js and Cytoscape.js
- **Tree Shaking**: Unused code elimination
- **Asset Hashing**: Cache busting with content hashes
- **Lazy Loading**: Chart libraries loaded on demand

### Backend

- **Cache Headers**: Appropriate caching for different asset types
- **Embedded Assets**: Zero filesystem overhead in production
- **Efficient Routing**: Fast API route detection
- **Compression**: Static asset compression for faster transfers

## Deployment

### Production Deployment

```bash
# Build complete dashboard
./scripts/build-dashboard.sh

# The resulting binary includes:
# - Complete Go backend
# - Embedded dashboard assets
# - Single-file deployment

./portfolio dashboard --port 3000
```

### Development Deployment

```bash
# Start development environment
./scripts/dev-dashboard.sh

# Or start components separately:
cd dashboard && npm run dev  # Terminal 1
./portfolio dashboard --dev  # Terminal 2
```

## Configuration

The dashboard respects Portfolio configuration:

- **Database Path**: SQLite database location
- **API Endpoints**: Backend service URLs
- **Logging**: Verbosity and output format
- **CORS**: Cross-origin settings for development

## Troubleshooting

### Build Issues

**Issue**: `go:embed dist/*: no matching files found`

**Solution**: Run `./scripts/build-dashboard.sh` instead of building separately

**Issue**: Vite build fails with module resolution errors

**Solution**: Clear cache and reinstall:
```bash
cd dashboard
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Runtime Issues

**Issue**: Dashboard shows "API not responding"

**Solution**: Ensure database is initialized and backend is running:
```bash
./portfolio doctor
./portfolio dashboard --dev
```

**Issue**: Charts don't render

**Solution**: Check browser console for errors and ensure Chart.js is loaded

### Development Issues

**Issue**: Hot reload not working

**Solution**: Ensure Vite dev server is running on port 5173:
```bash
cd dashboard
npm run dev
```

## Future Enhancements

Potential improvements for future releases:

- **Real-time Updates**: WebSocket integration for live updates
- **Advanced Filtering**: More sophisticated search and filtering options
- **Export Features**: Export project data and statistics
- **Custom Views**: User-configurable dashboard layouts
- **Collaboration**: Share dashboard views with team members
- **Mobile Apps**: Native mobile applications

## Contributing

When contributing to the dashboard:

1. Follow existing code patterns and conventions
2. Ensure accessibility compliance (WCAG 2.1 AA)
3. Test in both light and dark modes
4. Verify responsive design on multiple screen sizes
5. Run type checking: `npm run typecheck`
6. Test production build: `npm run build && npm run preview`

## License

Part of the Portfolio project. See main project license for details.