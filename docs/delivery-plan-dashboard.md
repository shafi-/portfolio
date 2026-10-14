# Dashboard Delivery Plan - Portfolio Milestone 3

## Executive Summary

Implement a complete read-only dashboard for the Portfolio project that provides visualization of the local knowledge store. The dashboard will be served as a single Go binary with embedded static assets, using port 7575 with intelligent fallback for zero-configuration deployment.

**Target Delivery**: Complete dashboard with 5 pages, full API integration, WCAG 2.1 AA accessibility, and <250KB bundle size.

---

## 1. Technical Specifications

### 1.1 User Preferences Applied
- **Visual Design**: System preference auto-detection (dark/light theme) with Inter modern sans-serif fonts
- **Component Library**: shadcn/ui + Tailwind CSS for rapid development with pre-built accessible components
- **Development Workflow**: Hot reload enabled (Vite dev server) + production single-binary deployment
- **Device Priority**: Desktop-first with responsive mobile support
- **Accessibility**: WCAG 2.1 AA compliance with axe-core testing
- **Charts**: Chart.js for functional data visualizations
- **Routing**: Hash-based routing (#/projects, #/relationships, etc.)

### 1.2 Technology Stack
**Backend (Go)**:
- Go 1.22+ with embed.FS for static assets
- Existing Epic 6 HTTP API integration
- Dynamic port selection (7575 default, 7576-7599 fallback)

**Frontend (React)**:
- React 18 + TypeScript 5 (strict mode)
- Vite 5 for build tooling and hot reload
- Tailwind CSS v3 for styling
- shadcn/ui component library
- Chart.js 4 + react-chartjs-2 for visualizations
- React Router v6 with hash-based routing
- @axe-core/react for accessibility testing

---

## 2. Architecture Design

### 2.1 Directory Structure
```
project-dash/
├── cmd/portfolio/
│   └── main.go                 # Entry point with graceful shutdown
├── internal/
│   ├── dashboard/
│   │   ├── server.go           # Main dashboard HTTP server
│   │   ├── port.go             # Dynamic port selection logic
│   │   ├── assets/
│   │   │   ├── handler.go      # Static file serving (dev/prod)
│   │   │   └── embed.go        # Embedded filesystem setup
│   │   ├── api/
│   │   │   ├── health.go       # Health check endpoint
│   │   │   └── config.go       # Configuration endpoint
│   │   └── middleware/
│   │       ├── cors.go         # CORS middleware
│   │       ├── logging.go      # Request logging
│   │       └── limits.go       # Rate limiting
│   └── api/                    # Existing Epic 6 API (projects, search, etc.)
└── dashboard/
    ├── dist/                   # Built dashboard files (generated)
    │   ├── index.html
    │   └── assets/
    │       └── *.js, *.css
    └── src/
        ├── pages/
        │   ├── Overview.tsx
        │   ├── ProjectList.tsx
        │   ├── ProjectDetail.tsx
        │   ├── Relationships.tsx
        │   └── Statistics.tsx
        ├── components/
        │   ├── ui/              # shadcn/ui components
        │   ├── charts/          # Chart components
        │   └── layout/          # Layout components
        ├── config/
        │   └── api.ts           # API client configuration
        ├── App.tsx
        └── main.tsx
```

### 2.2 Server Architecture

**Request Flow**:
```
User Request → Go Server → Route Resolution
├── /api/*          → Epic 6 API handlers → SQLite database
├── /api/health     → Dashboard health handler
├── /api/config     → Dashboard config handler
└── /*              → Static asset handler (embedded or filesystem)
```

**Port Selection Logic**:
1. Try port 7575 (default)
2. If unavailable, try 7576-7599 sequentially
3. If all unavailable, return error
4. Support PORT environment variable override

---

## 3. Bundle Size Optimization Strategy

### 3.1 Target Sizes
**Total dist directory: ~200-250KB**
- Runtime & vendor chunks: ~100KB
- Page chunks: ~25KB each (lazy loaded)
- CSS (purged): ~15KB
- HTML: ~2KB

### 3.2 Optimization Techniques

**Vite Configuration**:
```javascript
export default defineConfig({
  build: {
    target: 'es2020',        // Modern browsers, no polyfills
    minify: 'terser',       // Aggressive minification
    cssCodeSplit: true,     // Separate CSS chunks
    chunkSizeWarningLimit: 50,
    
    terserOptions: {
      compress: {
        drop_console: true,    // Remove console.logs
        drop_debugger: true,
        pure_funcs: ['console.log', 'console.info'],
      },
    },
    
    rollupOptions: {
      output: {
        manualChunks: {
          'react': ['react', 'react-dom', 'react-router-dom'],
          'charts': ['chart.js', 'react-chartjs-2'],
        },
      },
    },
  },
})
```

**Tree Shaking**:
- Install only required shadcn/ui components (button, card, dialog, dropdown-menu, input, table)
- Use specific Chart.js imports, not 'chart.js/auto'
- Lazy load all pages with React.lazy()

**Asset Optimization**:
- PurgeCSS via Tailwind (remove unused styles)
- Image optimization (WebP format, responsive sizing)
- Remove unused dependencies

---

## 4. Implementation Phases

### Phase 1: Frontend Foundation (Epic 12)
**Duration**: 2-3 days

**Tasks**:
1. Create React + TypeScript + Vite project
2. Configure Tailwind CSS + shadcn/ui
3. Set up hash-based routing with React Router
4. Create API client with type safety
5. Implement layout components (Layout, NavBar, Footer)
6. Configure build optimization
7. Set up accessibility testing (axe-core)

**Deliverables**:
- Working React project with hot reload
- Basic layout structure
- API client integrated with Epic 6 endpoints
- Build configuration for optimization

### Phase 2: Page Implementation
**Duration**: 4-5 days

**Tasks**:
1. **Overview Page**: High-level statistics, technology distribution, recent activity
2. **Project List Page**: Searchable, filterable, sortable project grid
3. **Project Detail Page**: Comprehensive project view with progressive enhancement
4. **Relationships Page**: Interactive graph visualization with list fallback
5. **Statistics Page**: Portfolio-wide analytics and visualizations

**Deliverables**:
- 5 fully functional pages with loading/error states
- Chart.js integrations for data visualization
- Responsive design for desktop and mobile
- Progressive enhancement (works without AI analysis)

### Phase 3: Accessibility & Polish
**Duration**: 2 days

**Tasks**:
1. Implement WCAG 2.1 AA requirements
2. Add ARIA labels and semantic HTML
3. Keyboard navigation support
4. Screen reader compatibility testing
5. Color contrast validation
6. Focus management on route changes

**Deliverables**:
- axe-core scans with zero violations
- Keyboard-accessible interface
- Screen reader tested (NVDA, JAWS, VoiceOver)

### Phase 4: Backend Integration
**Duration**: 1-2 days

**Tasks**:
1. Implement Go dashboard server
2. Set up dynamic port selection (7575)
3. Create asset serving logic (dev/prod modes)
4. Add dashboard-specific API endpoints (health, config)
5. Implement graceful shutdown
6. Remove dashboard configuration from TOML

**Deliverables**:
- Go server that serves both API and static assets
- Automatic port selection with fallback
- Zero-configuration startup
- Embedded dashboard in production binary

### Phase 5: Build & Deployment
**Duration**: 1 day

**Tasks**:
1. Configure production build process
2. Set up Go embed for dashboard assets
3. Create build scripts (npm run build + go build)
4. Test single-binary deployment
5. Verify bundle sizes
6. Test port selection fallback

**Deliverables**:
- Production build process
- Single Go binary (~8.25MB with dashboard)
- Build size validation
- Deployment documentation

---

## 5. API Integration

### 5.1 API Endpoints Used

**Epic 6 Endpoints**:
```typescript
// Projects
GET  /api/projects           // List all projects
GET  /api/projects/:id       // Get specific project
GET  /api/projects/:id/documents // Get project documents
GET  /api/projects/:id/analysis   // Get project analysis

// Search & Relationships
GET  /api/search?q=query     // Search across projects
GET  /api/relationships      // Get all relationships
GET  /api/relationships/:id  // Get relationships for project

// Statistics
GET  /api/statistics         // Get portfolio statistics
GET  /api/documents         // Get documents list
```

**Dashboard Endpoints**:
```typescript
GET  /api/health             // Server health check
GET  /api/configuration      // Dashboard configuration
```

### 5.2 API Client Configuration

```typescript
// dashboard/src/config/api.ts
const API_BASE_URL = '/api'  // No proxy needed, served from same origin

export const api = {
  getProjects: () => 
    fetch(`${API_BASE_URL}/projects`).then(r => r.json()),
  
  getProject: (id: string) => 
    fetch(`${API_BASE_URL}/projects/${id}`).then(r => r.json()),
  
  search: (query: string) => 
    fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(query)}`).then(r => r.json()),
  
  // ... other endpoints
}
```

---

## 6. Testing Strategy

### 6.1 Frontend Testing
- Component testing with React Testing Library
- Accessibility testing with axe-core
- Build validation (bundle size checks)
- Manual testing of all 5 pages

### 6.2 Backend Testing
- Unit tests for port selection logic
- Integration tests for asset serving
- API endpoint testing
- Graceful shutdown testing

### 6.3 Integration Testing
- Development mode (external assets)
- Production mode (embedded assets)
- Port fallback scenarios
- API integration validation

---

## 7. Configuration Changes

### 7.1 Removed Configuration
**Before**:
```toml
[dashboard]
host = "localhost"
port = 3000
asset_path = ""
allowed_origins = ["http://localhost:3000"]
```

**After**:
```toml
# Dashboard section completely removed
# Port automatically selected (7575 default, 7576-7599 fallback)
# Assets embedded in binary

[database]
path = "./portfolio.db"

[logging]
level = "info"
```

### 7.2 Environment Variables (Optional)
```bash
PORT=7575              # Override default port (not recommended)
DASHBOARD_ASSET_PATH="./dashboard/dist"  # Development mode
```

---

## 8. Success Criteria

### 8.1 Functional Requirements
✅ Dashboard serves as single Go binary with embedded assets  
✅ All 5 pages render with deterministic knowledge  
✅ Read-only constraint enforced (no mutations)  
✅ Progressive enhancement (useful without AI analysis)  
✅ All HTTP API endpoints accessible from frontend  

### 8.2 Non-Functional Requirements
✅ Bundle size <250KB total  
✅ WCAG 2.1 AA accessibility compliance  
✅ Fast page loads (<3 seconds initial, <1 second navigation)  
✅ Zero-configuration deployment  
✅ Automatic port selection with fallback  

### 8.3 Engineering Principles
✅ Engine Knows, Agent Thinks, Dashboard Visualizes  
✅ Dashboard is read-only  
✅ Local-first architecture respected  
✅ Single knowledge model across all views  

---

## 9. Development Workflow

### 9.1 Development Mode
```bash
# Terminal 1: Frontend dev server
cd dashboard
npm run dev  # Vite dev server on :5173 with hot reload

# Terminal 2: Backend server
cd project-dash
DASHBOARD_ASSET_PATH="./dashboard/dist" go run cmd/portfolio/main.go
# Dashboard on http://localhost:7575
```

### 9.2 Production Build
```bash
# Build frontend
cd dashboard
npm run build  # Creates dist/ directory

# Build Go binary with embedded dashboard
cd ..
go build -o portfolio cmd/portfolio/main.go

# Run single binary
./portfolio
# Dashboard on http://localhost:7575 (or 7576, 7577, etc.)
```

### 9.3 Verification Steps
1. Check bundle sizes: `du -sh dashboard/dist/` (<500KB expected)
2. Test automatic port selection: Run multiple instances
3. Verify API integration: Check browser console for errors
4. Test accessibility: Run axe-core scans
5. Verify progressive enhancement: Test with projects lacking AI analysis

---

## 10. Deliverables Checklist

### Frontend Deliverables
- [ ] React + TypeScript + Vite project structure
- [ ] Tailwind CSS + shadcn/ui configuration
- [ ] Hash-based routing setup
- [ ] Type-safe API client
- [ ] 5 fully functional pages
- [ ] Chart.js visualizations
- [ ] Responsive design (desktop-first)
- [ ] WCAG 2.1 AA compliance
- [ ] Progressive enhancement implementation
- [ ] Build optimization configuration
- [ ] Production build (<250KB)

### Backend Deliverables
- [ ] Dashboard HTTP server implementation
- [ ] Dynamic port selection (7575 default)
- [ ] Asset serving logic (dev/prod modes)
- [ ] Dashboard API endpoints (health, config)
- [ ] Go embed configuration
- [ ] Graceful shutdown implementation
- [ ] Configuration simplification (removed dashboard section)
- [ ] Single-binary deployment

### Documentation Deliverables
- [ ] Build process documentation
- [ ] Deployment instructions
- [ ] API integration guide
- [ ] Accessibility compliance report
- [ ] Troubleshooting guide

---

## 11. Timeline Estimate

**Total Duration**: 10-13 working days

- **Phase 1 (Frontend Foundation)**: 2-3 days
- **Phase 2 (Page Implementation)**: 4-5 days  
- **Phase 3 (Accessibility & Polish)**: 2 days
- **Phase 4 (Backend Integration)**: 1-2 days
- **Phase 5 (Build & Deployment)**: 1 day

---

## 12. Risk Mitigation

### 12.1 Technical Risks
**Risk**: Bundle size exceeds target  
**Mitigation**: Aggressive tree-shaking, Chart.js specific imports, lazy loading

**Risk**: Port conflicts on common development machines  
**Mitigation**: Intelligent fallback to 7576-7599, clear error messaging

**Risk**: shadcn/ui component library limitations  
**Mitigation**: Hybrid approach - use library for common components, build custom for specialized views

### 12.2 Timeline Risks
**Risk**: Accessibility compliance takes longer than expected  
**Mitigation**: Start accessibility testing early, use automated tools (axe-core)

**Risk**: API integration issues with Epic 6  
**Mitigation**: Early integration testing, fallback to deterministic data if semantic data unavailable

---

## 13. Post-Delivery Enhancements (Out of Scope)

These are explicitly deferred for future milestones:
- Real-time data updates (no polling/WebSockets)
- Advanced analytics and visualizations
- Mobile applications
- Team collaboration features
- Portfolio historical snapshots
- Custom dashboard configurations
- Authentication/authorization

---

## 14. Handoff Criteria

The dashboard feature will be considered complete when:
1. All 5 pages are functional with proper loading/error states
2. Bundle size is <250KB (verified)
3. WCAG 2.1 AA compliance achieved (axe-core zero violations)
4. Single Go binary serves everything on port 7575 (with fallback)
5. Zero-configuration deployment works
6. All API endpoints integrate correctly
7. Progressive enhancement works (useful without AI analysis)
8. Build process is documented and repeatable
9. All tests pass (frontend + backend)
10. Documentation is complete

---

**Document Version**: 1.0  
**Last Updated**: 2025-08-09  
**Status**: Ready for Implementation  
**Owner**: Portfolio Team  
**Approvals**: Pending team review

---

## Appendix A: Port Selection Rationale

**Why Port 7575?**
- Rarely used by major frameworks or tools
- Not in common development server ranges (3000-3001, 4000, 5000, 8000, 8080)
- Memorable and easily documented
- Allows range fallback (7575-7599) for conflict resolution

**Fallback Strategy**:
1. Primary: 7575 (documented default)
2. Fallback range: 7576-7599 (24 alternatives)
3. Environment override: PORT variable (advanced users)
4. Error: Clear message if all ports in range unavailable

This provides zero-configuration experience for 99% of use cases while maintaining flexibility for advanced scenarios.