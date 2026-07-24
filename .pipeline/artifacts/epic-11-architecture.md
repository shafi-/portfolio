# Epic 11 — Dashboard Backend: Architecture

## Overview

The Dashboard Backend is a read-only HTTP serving layer within the Portfolio Go engine. It binds to a configurable local port, serves static frontend assets, and wraps Epic 6 API handlers for dashboard consumption. It lives in the same process as the engine, shares the same SQLite store, and never invokes AI or mutates state.

---

## 1. Package Structure

```
internal/
  dashboard/
    server.go              — HTTP server setup, graceful shutdown, route registration
    server_test.go
    handler.go             — shared JSON response helpers, error contract
    handler_test.go
    assets/
      handler.go           — static file serving + SPA fallback
      handler_test.go
      embed.go             — embedded asset filesystem (go:embed)
    api/
      handler.go           — thin wrapper that delegates to Epic 6 API handlers
      search.go            — enhanced search with filters, pagination, snippets
      health.go            — health endpoint with DB ping
      configuration.go     — config GET/PATCH
    middleware/
      cors.go              — CORS middleware
      limits.go            — request body size limiter
      logging.go           — request logging middleware
    errors.go              — error contract types and helpers
```

### Package Responsibilities

| Package | Responsibility |
|---------|---------------|
| `dashboard` | Server lifecycle, route registration, shared handler utilities |
| `dashboard/assets` | Static file serving, embedded assets, SPA fallback |
| `dashboard/api` | Thin wrappers over Epic 6 handlers + enhanced search + health/config |
| `dashboard/middleware` | CORS, body limits, request logging |
| `dashboard/errors.go` | Error contract types, JSON error response builder |

### Relationship to Epic 6

- `internal/dashboard/api/` wraps `internal/api/` handlers — no business logic duplication
- `internal/dashboard` depends on `pkg/models` and `internal/config`
- `internal/dashboard` MUST NOT import any AI/LLM package
- `internal/dashboard` is imported by `cmd/portfolio/main.go` to start the server

---

## 2. Handler Wrapping Pattern

Dashboard handlers are thin wrappers over Epic 6 API handlers. They call the Epic 6 handler, inspect the error, and write the consistent JSON error contract:

```go
func (h *Handler) GetProjects(w http.ResponseWriter, r *http.Request) {
    data, err := h.epic6.GetProjects(r)
    if err != nil {
        respondError(w, wrapError(err))
        return
    }
    respondJSON(w, 200, data)
}
```

| Concern | Epic 6 (`internal/api`) | Dashboard (`internal/dashboard/api`) |
|---------|------------------------|--------------------------------------|
| Business logic | SQL queries, data transformation | None — delegates entirely |
| Route registration | Registers on shared mux | Registers on same mux (same routes) |
| Error responses | Returns `error` to caller | Wraps with consistent JSON error contract |
| CORS | Not applied | Applied via middleware |
| Request body limits | Not applied | Applied for PATCH |
| Static assets | Not served | Served via `assets/` handler |
| SPA fallback | Not handled | Handled by `assets/` handler |

Search, health, and configuration handlers are defined in the dashboard package because they add dashboard-specific enhancements (filters, pagination, snippets, health format, config partial update) beyond Epic 6's base implementation.

---

## 3. Route Table

| Method | Path | Handler | Middleware | Source |
|--------|------|---------|-----------|--------|
| GET | `/health` | `api/health.go` | CORS | Dashboard-specific |
| GET | `/projects` | `api/handler.go` (wraps Epic 6) | CORS | Wraps Epic 6 |
| GET | `/projects/{id}` | `api/handler.go` (wraps Epic 6) | CORS | Wraps Epic 6 |
| GET | `/projects/{id}/documents` | `api/handler.go` (wraps Epic 6) | CORS | Wraps Epic 6 |
| GET | `/projects/{id}/analysis` | `api/handler.go` (wraps Epic 6) | CORS | Wraps Epic 6 |
| GET | `/search` | `api/search.go` | CORS | Dashboard-enhanced |
| GET | `/relationships/{projectId}` | `api/handler.go` (wraps Epic 6) | CORS | Wraps Epic 6 |
| GET | `/statistics` | `api/handler.go` (wraps Epic 6) | CORS | Wraps Epic 6 |
| GET | `/configuration` | `api/configuration.go` | CORS | Dashboard-specific |
| PATCH | `/configuration` | `api/configuration.go` | CORS + body limit | Dashboard-specific |
| GET | `/assets/*` | `assets/handler.go` | none | Dashboard only |
| GET | `/*` (SPA fallback) | `assets/handler.go` | none | Dashboard only |

### Route Registration Order

1. API routes — registered first to take priority over SPA catch-all
2. Asset routes (`/assets/*`) — registered second
3. SPA fallback (`/*`) — registered last, catches unmatched paths

---

## 4. Static File Serving (Story 11.1)

### Two Modes

1. **Embedded assets** — compiled into Go binary via `//go:embed` for production
2. **External assets** — served from configurable directory on disk (dev)

### Resolution Order

1. If embedded assets compiled in, serve from `embed.FS`
2. If external path configured and exists, serve from disk
3. If neither, return 404

### SPA Fallback

```
Request path → /api/*? → route to API handler
             → /assets/*? → serve file or 404
             → known file? → serve file
             → otherwise → serve index.html (SPA fallback)
```

### Cache Headers

- Fingerprinted assets (`main.a1b2c3.js`): `Cache-Control: public, max-age=31536000, immutable`
- Non-fingerprinted: `Cache-Control: no-cache, ETag: <file-hash>`
- `index.html`: `Cache-Control: no-cache`

### Security

- Reject paths containing `..` — return 400
- No directory listing — return 404 for directory paths
- Path sanitization before filesystem access

---

## 5. Enhanced Search (Story 11.2)

```
GET /search?q=<term>&technology=<tech>&framework=<fw>&from=<iso>&to=<iso>&page=1&page_size=20
```

### Response Shape

```json
{
  "total_results": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3,
  "results": [
    {
      "type": "project",
      "id": "uuid",
      "name": "my-project",
      "snippet": "...uses <mark>React</mark> for UI..."
    },
    {
      "type": "document",
      "id": "uuid",
      "project_id": "uuid",
      "project_name": "my-project",
      "path": "README.md",
      "kind": "README",
      "snippet": "...built with <mark>React</mark>..."
    }
  ]
}
```

### Implementation

1. Parse and validate query parameters
2. Query SQLite FTS index on `projects.name`, `documents.content`
3. Apply optional filters as WHERE clauses (technology, framework, date range)
4. Apply LIMIT/OFFSET from pagination params
5. Compute snippets: locate query term in result text, wrap in `<mark>` tags
6. Return paginated JSON response

---

## 6. Server Lifecycle

```go
type Server struct {
    httpServer *http.Server
    epic6API   *api.Handler     // Epic 6 API handlers
    store      *sqlite.Store
    config     *models.Config
    logger     *zap.Logger
}

func NewServer(cfg *models.Config, store *sqlite.Store, epic6API *api.Handler, logger *zap.Logger) *Server

func (s *Server) Start() error {
    mux := http.NewServeMux()
    
    dashAPI := dashboardapi.NewHandler(s.epic6API, s.store, s.logger)
    dashAPI.RegisterRoutes(mux)
    
    assetHandler := assets.NewHandler(s.config.Dashboard.AssetsPath, s.logger)
    assetHandler.RegisterRoutes(mux)
    
    handler := middleware.CORS(s.config.Dashboard.AllowedOrigins)(mux)
    handler = middleware.BodyLimit(1 << 20)(handler)
    
    s.httpServer = &http.Server{
        Addr:    fmt.Sprintf("%s:%d", s.config.Dashboard.Host, s.config.Dashboard.Port),
        Handler: handler,
    }
    return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.httpServer.Shutdown(ctx)
}
```

Graceful shutdown on SIGINT/SIGTERM with 30-second timeout.

---

## 7. CORS

```go
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler
```

- Default origins: `["http://localhost:5173", "http://localhost:3000"]`
- Methods: GET, PATCH, OPTIONS
- Headers: Content-Type, Authorization
- Preflight returns 204 No Content
- Configurable via `dashboard.allowed_origins`

---

## 8. Error Contract

```json
{ "error": "error_code", "message": "Human-readable description" }
```

| Code | HTTP Status | When |
|------|-------------|------|
| `bad_request` | 400 | Missing `q`, invalid date, invalid pagination, unknown config keys, directory traversal |
| `not_found` | 404 | Missing project/document/asset, route not found |
| `method_not_allowed` | 405 | Wrong method on known route (includes Allow header) |
| `request_too_large` | 413 | PATCH body exceeds 1MB |
| `internal_error` | 500 | Database unavailable, unexpected panic |

---

## 9. Test Strategy

| Layer | Scope | Approach |
|-------|-------|----------|
| Unit | Asset handler | Table-driven tests with httptest — file found, not found, directory traversal, SPA fallback |
| Unit | Search | Query parsing, filter application, pagination, snippet highlighting |
| Unit | Middleware | CORS preflight/allowed/disallowed, body limit enforcement |
| Unit | Error contract | All error codes and response shapes |
| Integration | Full request flow | Start test server, send HTTP, assert responses |
| Integration | Search with SQLite | In-memory SQLite with seeded data, test filters + pagination |

---

## 10. Implementation Order

### Phase 1: Asset Serving (Story 11.1)

1. Create `internal/dashboard/` package structure
2. Implement `assets/handler.go` — file server, MIME types, cache headers, directory traversal prevention
3. Implement `assets/embed.go` — embedded asset filesystem
4. Implement SPA fallback routing
5. Implement `server.go` — HTTP server setup, route registration, graceful shutdown
6. Unit tests

### Phase 2: Endpoint Review (Story 11.2)

1. Create `internal/dashboard/middleware/cors.go` — CORS middleware
2. Create `internal/dashboard/middleware/limits.go` — body size limiter
3. Create `internal/dashboard/middleware/logging.go` — request logging
4. Create `internal/dashboard/errors.go` — error contract types
5. Create `internal/dashboard/handler.go` — shared JSON helpers, error wrapping
6. Create `internal/dashboard/api/handler.go` — thin wrappers over Epic 6 handlers
7. Implement `api/search.go` — enhanced search with filters, pagination, snippets
8. Implement `api/health.go` — health endpoint with DB ping
9. Implement `api/configuration.go` — config GET and PATCH with validation
10. Wire routes in `server.go`
11. Unit + integration tests

---

## 11. Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Handler reuse | Wrap Epic 6, don't duplicate | Single source of truth for business logic |
| Route prefix | Bare routes (no `/api/`) | Matches PlatformSpecification.md exactly |
| CORS | Hand-rolled middleware | Simple allow-list, no external dep needed |
| Body limit | Middleware on mux level | Simpler than per-route; only PATCH matters |
| SPA fallback | Path inspection in asset handler | No separate router needed |
| Search snippets | Go-side `<mark>` wrapping | Deterministic, no JS required |
| Config PATCH | Partial merge in Go | Full config round-trip, unknown keys rejected |
| Health DB check | `SELECT 1` ping | Lightweight, confirms actual connectivity |
