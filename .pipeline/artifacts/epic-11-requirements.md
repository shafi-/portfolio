# Epic 11 — Dashboard Backend: Requirements

## Feature Overview

The Dashboard Backend is a read-only HTTP serving layer within the Portfolio Go engine that delivers static frontend assets for the dashboard UI and reviews/enhances the Epic 6 HTTP API for dashboard consumption.

This epic establishes the server-side delivery mechanism — the dashboard frontend is a separate concern. The backend lives in the Go engine process, binds to a local port, and serves deterministic data from the SQLite knowledge store. It never invokes AI agents, never modifies repositories, and never performs analysis.

---

## Requirements

### Functional Requirements

#### FR1 — Static Asset Serving (Story 11.1)

| ID | Requirement |
|----|-------------|
| FR1.1 | Serve static frontend assets (HTML, CSS, JS, images, favicon) from a configurable filesystem path. |
| FR1.2 | Support both embedded assets (compiled into Go binary via `//go:embed`) and external assets (loaded from disk). |
| FR1.3 | Return correct MIME types for all served file extensions. |
| FR1.4 | Set Cache-Control headers: immutable cache for fingerprinted assets, ETag for non-fingerprinted, no-cache for index.html. |
| FR1.5 | Serve single-page application entry point (index.html) for all non-API, non-asset paths (SPA routing). |
| FR1.6 | Prevent directory traversal — reject paths containing `..`. Never serve directory listings. |

#### FR2 — Endpoint Review (Story 11.2)

| ID | Requirement |
|----|-------------|
| FR2.1 | Wire all Epic 6 HTTP API endpoints (`/health`, `/projects`, `/projects/{id}`, `/projects/{id}/documents`, `/projects/{id}/analysis`, `/search`, `/relationships/{projectId}`, `/statistics`, `/configuration`) for dashboard consumption. |
| FR2.2 | Add enhanced search filters to `GET /search`: `technology` (filter by technology name), `framework` (filter by framework name), `from`/`to` (ISO 8601 date range on discovered_at / last_scan_at). |
| FR2.3 | Add pagination to search results via `page` and `page_size` query parameters (defaults: page=1, page_size=20; max page_size=100). Response includes `total_results`, `page`, `page_size`, `total_pages`, and `results` array. |
| FR2.4 | Add highlighted snippets in search results showing matching context with query term wrapped in `<mark>` tags. |
| FR2.5 | Verify `GET /health` returns 200 with `status`, `uptime_seconds`, and `database` connectivity (verified via `SELECT 1`). |
| FR2.6 | Verify `GET /configuration` returns current Portfolio configuration as JSON. Verify `PATCH /configuration` accepts partial updates with validation. |
| FR2.7 | Apply CORS headers for local development origins, allowing methods GET, PATCH, OPTIONS. |
| FR2.8 | No authentication required (local-only deployment). |
| FR2.9 | Consistent error responses using standard HTTP status codes (200, 400, 404, 405, 413, 500) and uniform JSON shape `{"error": "code", "message": "..."}`. |
| FR2.10 | All responses use canonical knowledge model entities — no interface-specific representations. |

#### FR3 — Request Limits

| ID | Requirement |
|----|-------------|
| FR3.1 | Maximum request body size for PATCH /configuration is 1MB. Larger payloads return 413. |

### Non-Functional Requirements

| ID | Requirement |
|----|-------------|
| NFR1 | **Local-First:** All data served from local SQLite store. Zero external network calls. |
| NFR2 | **Read-Only:** Dashboard endpoints MUST NOT mutate state. All mutations handled by CLI or MCP. |
| NFR3 | **Performance:** Static asset responses <10ms (cached). API responses <100ms for portfolios up to 500 projects. |
| NFR4 | **Concurrency:** Safe concurrent SQLite access via WAL mode (single writer / multiple readers). |
| NFR5 | **Port Binding:** Configurable host and port (defaults: localhost:8090). |
| NFR6 | **Graceful Shutdown:** HTTP server must drain in-flight requests before exit. |
| NFR7 | **No AI:** Dashboard backend must never import, invoke, or depend on any AI/LLM module. |
| NFR8 | **Single Knowledge Model:** Every response JSON shape must derive from KnowledgeModel.md entities. |
| NFR9 | **Deterministic:** All responses must be deterministic — same query + same data = same result. |

---

## Edge Cases & Error Handling

| # | Scenario | Expected Behavior |
|---|----------|-------------------|
| EC-1 | Asset path not configured or missing on disk | Return 404. Log warning at startup. |
| EC-2 | Non-existent static asset requested | Return 404 `{"error": "not_found", "message": "..."}` |
| EC-3 | Directory traversal attempt in asset path | Reject paths containing `..`. Return 400. |
| EC-4 | Missing `q` parameter on search | Return 400 `{"error": "bad_request", "message": "query parameter 'q' is required"}` |
| EC-5 | Invalid pagination values (negative, non-numeric, exceeds max) | Clamp to valid range or return 400. |
| EC-6 | Invalid date range format | Return 400 `{"error": "bad_request", "message": "invalid date format — use ISO 8601"}` |
| EC-7 | No results match search | Return 200 with `{"total_results": 0, "results": []}` |
| EC-8 | PATCH configuration with unknown key | Reject unknown keys. Return 400. |
| EC-9 | PATCH configuration with wrong value type | Return 400. |
| EC-10 | Database unavailable | Return 500 `{"error": "internal_error", "message": "database unavailable"}` |
| EC-11 | Concurrent server start on same port | Second instance fails fast with clear error. |
| EC-12 | Request path not found | Return 404. |
| EC-13 | Method not allowed on known endpoint | Return 405 with Allow header. |
| EC-14 | SPA route fallback | Non-file, non-API paths serve index.html (200, not 404). |
| EC-15 | Large search result set | Respect `page_size` cap (100). Paginate with `total_pages` indicator. |
| EC-16 | Health endpoint with DB unavailable | Return 200 with `status: "degraded"`, `database: "unavailable"`. |

---

## Acceptance Criteria

### Story 11.1 — Asset Serving
- AC1.1: Dashboard frontend loads via browser at configured address
- AC1.2: All asset types serve with correct MIME types
- AC1.3: Cache headers present on static assets (immutable for fingerprinted, ETag for non-fingerprinted)
- AC1.4: Works with both embedded and external asset configurations
- AC1.5: SPA fallback routing works — non-file, non-API paths return index.html
- AC1.6: Directory traversal attempts return 400
- AC1.7: Missing assets return 404 (not crash or directory listing)

### Story 11.2 — Endpoint Review
- AC2.1: All Epic 6 HTTP API endpoints respond successfully through dashboard routes
- AC2.2: Search filters (technology, framework, date range) narrow results correctly
- AC2.3: Pagination returns correct slices and `total_pages`
- AC2.4: Highlighted snippets present in search results with `<mark>` tags
- AC2.5: `GET /health` returns status, uptime, and DB connectivity
- AC2.6: `GET /configuration` returns current config as JSON
- AC2.7: `PATCH /configuration` accepts partial updates, returns updated config
- AC2.8: Invalid PATCH payload returns 400 with validation details
- AC2.9: CORS headers present for configured origins
- AC2.10: Requests without authentication succeed
- AC2.11: Error responses use consistent JSON shape with proper HTTP status codes

---

## Data Flow

```
User Browser
    |
    v
Dashboard (SPA) ─── GET /assets/* ──────> Go HTTP Server
    │                                          │
    │          [FileServer: embed or disk]      │
    │  <── HTML/CSS/JS/images ─────────────────+
    │
    │── GET /api/projects ───────────────> Go HTTP Server
    │                                          │
    │          [Wraps Epic 6 handler + cors]   │
    │  <── JSON response ──────────────────────+
    │
    │── GET /search?q=X&technology=Y&page=1 ──> Go HTTP Server
    │                                          │
    │          [Search handler → SQLite FTS]   │
    │          [Apply filters + pagination]    │
    │          [Compute highlighted snippets]  │
    │  <── JSON with pagination + snippets ────+
```

---

## Dependencies

| Dependency | Type | Notes |
|-----------|------|-------|
| Epic 6 (HTTP API) | Hard block | Provides core HTTP handlers and search infrastructure that 11.2 reviews and enhances |
| Dashboard frontend build output | Runtime | Static assets to serve (built separately by Epic 12) |
| Go standard library `net/http` | Runtime | Asset serving and API routing |
| SQLite (via mattn/go-sqlite3 or modernc.org/sqlite) | Runtime | Data queries |

**Blocking Chain:**
```
Epic 6 (HTTP API Core) ──blocks──> 11.1 Asset Serving
Epic 6 (HTTP API Core) ──blocks──> 11.2 Endpoint Review
```

**Implementation Order:**
1. 11.1 Asset Serving — no dependency on 11.2
2. 11.2 Endpoint Review — wraps Epic 6 handlers, adds search enhancements, CORS, error contract
