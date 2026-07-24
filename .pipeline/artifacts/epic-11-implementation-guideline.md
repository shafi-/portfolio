# Epic 11 — Dashboard Backend: Implementation Guideline

**Reference:** `.architecture/epic-11-architecture.md`, `.requirements/epic-11-requirements.md`, `docs/tasks/epic-11-dashboard-backend.md`

---

## 1. Technical Standards

### Language & Runtime
- Go 1.22.6, module `github.com/nerddevsltd/portfolio`
- Extends HTTP API from Epic 6 with static file serving
- Stdlib `net/http` `FileServer` for static assets
- Dashboard frontend built separately (Epic 12), served from `dashboard/dist/`

### Package Organization
- Extends `internal/api/` from Epic 6
- Adds static file serving in `server.go`:
```go
mux.Handle("/", http.FileServer(http.Dir("./dashboard/dist")))
```

### Key Design Decisions
- Dashboard SPA served as static files from Go binary (embedded or on-disk)
- Hash-based routing means no catch-all route needed (browser never sends `#/path` to server)
- API and dashboard on same origin (no separate dev server needed for production)
- CORS optional and disabled by default (same-origin only)

## 2. Static File Serving

```go
// Embed dashboard dist in binary (Go 1.22 embed)
//go:embed dashboard/dist/*
var dashboardFS embed.FS

func dashboardFileServer() http.Handler {
    sub, _ := fs.Sub(dashboardFS, "dashboard/dist")
    return http.FileServer(http.FS(sub))
}
```

Or serve from on-disk path for development:
```go
func dashboardFileServer(cfg APIConfig) http.Handler {
    return http.FileServer(http.Dir(cfg.DashboardPath))
}
```

## 3. API Endpoints (from Epic 6)

All HTTP API endpoints from Epic 6 are served on the same port. The dashboard SPA consumes them via relative fetch (no CORS needed for production).

## 4. Implementation Order

### Story 11.1 — Asset Serving
- Configure static file serving for dashboard frontend
- Support embedded (production) and external (development) modes

### Story 11.2 — Dashboard API Integration
- Wire all Epic 6 endpoints for dashboard consumption
- CORS configuration for local development
- Error responses with proper status codes

### Story 11.3 — Search Endpoints
- Unified search with filters (technology, framework, date range)
- Pagination support
- Highlighted snippets

### Story 11.4 — Service Endpoints
- GET /configuration, PATCH /configuration with validation
- GET /health with uptime + DB connectivity

## 5. Testing Strategy

### Unit Tests
- Static file serving (correct MIME types, cache headers)
- Embedded vs external mode switching
- Search endpoint filters and pagination

### Integration Tests
- Full request: serve index.html → JS boots → fetch /projects
- CORS headers present when enabled
- 404 for missing static assets

## 6. Build & Verification

```bash
go build ./cmd/portfolio
go test ./internal/api/... -v -cover
```

## 7. Quality Gates

- [ ] Dashboard static files served with correct MIME types
- [ ] Embed works in production build
- [ ] External path works in development
- [ ] API and dashboard on same origin (no CORS issues in production)
- [ ] All acceptance criteria from `.requirements/epic-11-requirements.md` pass
