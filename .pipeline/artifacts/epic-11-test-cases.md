# Epic 11 — Dashboard Backend: Test Cases

## Story 11.1 — Static Asset Serving

### TC-11.1-001: Serve index.html from configured asset path
**Title:** Serve dashboard entry point
**Description:** Verify that `GET /` or `GET /index.html` returns the SPA entry point with correct MIME type.
**Preconditions:** Asset path configured and `index.html` exists on disk.
**Steps:**
1. Start Portfolio engine with asset path configured
2. Send `GET /` to the dashboard HTTP server
3. Send `GET /index.html` to the dashboard HTTP server
**Expected Result:** Both requests return HTTP 200 with `Content-Type: text/html` and the contents of `index.html`.
**Story:** 11.1

---

### TC-11.1-002: Serve CSS and JavaScript assets with correct MIME types
**Title:** Serve CSS and JS assets
**Description:** Verify that `.css` and `.js` files are served with correct `Content-Type` headers.
**Preconditions:** Asset path contains `style.css` and `app.js`.
**Steps:**
1. Request `GET /style.css`
2. Request `GET /app.js`
**Expected Result:** CSS returns `Content-Type: text/css; charset=utf-8`. JS returns `Content-Type: application/javascript` or `text/javascript`.
**Story:** 11.1

---

### TC-11.1-003: Serve image assets with correct MIME types
**Title:** Serve image assets
**Description:** Verify that image files (PNG, SVG) are served with correct MIME types.
**Preconditions:** Asset path contains `logo.png` and `icon.svg`.
**Steps:**
1. Request `GET /logo.png`
2. Request `GET /icon.svg`
**Expected Result:** PNG returns `Content-Type: image/png`. SVG returns `Content-Type: image/svg+xml`.
**Story:** 11.1

---

### TC-11.1-004: Cache-Control headers on static assets
**Title:** Cache headers present on static assets
**Description:** Verify that Cache-Control headers are set on static asset responses.
**Preconditions:** Asset path configured with static assets.
**Steps:**
1. Request `GET /style.css`
2. Inspect response headers
**Expected Result:** Response includes `Cache-Control` header. Fingerprinted assets include `immutable` directive. Non-fingerprinted assets include `ETag` or `Last-Modified`.
**Story:** 11.1

---

### TC-11.1-005: Embedded assets served from compiled binary
**Title:** Embedded asset serving
**Description:** Verify that assets embedded in the Go binary are served correctly when external path is not configured.
**Preconditions:** Engine compiled with embedded dashboard assets. No external asset path configured.
**Steps:**
1. Start engine without setting external asset path
2. Request `GET /index.html`
3. Request `GET /app.js`
**Expected Result:** Both requests return HTTP 200 with correct content from embedded assets.
**Story:** 11.1

---

### TC-11.1-006: External assets override embedded assets
**Title:** External asset path takes precedence
**Description:** Verify that when both embedded and external assets are present, external assets are served.
**Preconditions:** Asset path configured pointing to a directory with a modified version of `index.html`.
**Steps:**
1. Configure asset path to a directory with a modified `index.html`
2. Request `GET /index.html`
**Expected Result:** Returns the modified version from the external path, not the embedded version.
**Story:** 11.1

---

### TC-11.1-007: SPA fallback routing serves index.html
**Title:** SPA fallback for client-side routes
**Description:** Verify that non-file, non-API paths serve `index.html` for client-side routing.
**Preconditions:** Asset path configured with valid dashboard build.
**Steps:**
1. Request `GET /projects`
2. Request `GET /settings/profile`
3. Request `GET /search?q=react`
**Expected Result:** All requests return HTTP 200 with `Content-Type: text/html` and contents of `index.html`.
**Story:** 11.1

---

### TC-11.1-008: Non-existent static asset returns 404
**Title:** Missing asset returns 404
**Description:** Verify that requesting a non-existent static asset returns 404, not a crash or directory listing.
**Preconditions:** Asset path configured. File `nonexistent.js` does not exist.
**Steps:**
1. Request `GET /nonexistent.js`
**Expected Result:** HTTP 404 with JSON body `{"error": "not_found", "message": "asset not found: nonexistent.js"}`.
**Story:** 11.1

---

### TC-11.1-009: Directory traversal attack rejected
**Title:** Directory traversal protection
**Description:** Verify that paths containing `..` are rejected.
**Preconditions:** Asset path configured.
**Steps:**
1. Request `GET /../etc/passwd`
2. Request `GET /assets/../../config.json`
**Expected Result:** HTTP 400 with JSON body `{"error": "bad_request", "message": "invalid path"}`.
**Story:** 11.1

---

### TC-11.1-010: Asset path not configured at startup logs warning
**Title:** Asset path misconfiguration warning
**Description:** Verify that a warning is logged at startup when asset path is configured but missing on disk.
**Preconditions:** Asset path configured to a non-existent directory.
**Steps:**
1. Start engine with asset path pointing to `/nonexistent/dashboard`
2. Check engine logs
**Expected Result:** Warning log emitted: `dashboard asset path not found: /nonexistent/dashboard`. Engine continues running.
**Story:** 11.1

---

### TC-11.1-011: favicon.ico served correctly
**Title:** Favicon serving
**Description:** Verify that `favicon.ico` is served with correct MIME type.
**Preconditions:** Asset path contains `favicon.ico`.
**Steps:**
1. Request `GET /favicon.ico`
**Expected Result:** HTTP 200 with `Content-Type: image/x-icon`.
**Story:** 11.1

---

## Story 11.2 — Dashboard API Integration

### TC-11.2-001: GET /health returns OK
**Title:** Health endpoint returns 200
**Description:** Verify `GET /health` returns status ok with uptime and DB status.
**Preconditions:** Engine running with SQLite knowledge store connected.
**Steps:**
1. Send `GET /health`
**Expected Result:** HTTP 200 with JSON body `{"status": "ok", "uptime": "<duration>", "database": "connected"}`.
**Story:** 11.2

---

### TC-11.2-002: GET /projects returns project list
**Title:** List all projects
**Description:** Verify `GET /projects` returns all projects from the knowledge store.
**Preconditions:** Knowledge store contains 3 projects with deterministic metadata.
**Steps:**
1. Send `GET /projects`
**Expected Result:** HTTP 200 with JSON array of projects. Each project contains `id`, `name`, `root_path`, `repository_type`, `discovered_at`, and metadata fields. Response conforms to canonical Project entity from KnowledgeModel.md.
**Story:** 11.2

---

### TC-11.2-003: GET /projects/{id} returns single project
**Title:** Get project by ID
**Description:** Verify `GET /projects/{id}` returns a single project by its UUID.
**Preconditions:** Knowledge store contains a project with known UUID.
**Steps:**
1. Send `GET /projects/{existing-uuid}`
**Expected Result:** HTTP 200 with JSON body containing the full project entity.
**Story:** 11.2

---

### TC-11.2-004: GET /projects/{id} returns 404 for unknown ID
**Title:** Project not found
**Description:** Verify that requesting a non-existent project UUID returns 404.
**Preconditions:** Knowledge store exists but does not contain the requested UUID.
**Steps:**
1. Send `GET /projects/00000000-0000-0000-0000-000000000000`
**Expected Result:** HTTP 404 with JSON body `{"error": "not_found", "message": "project not found"}`.
**Story:** 11.2

---

### TC-11.2-005: GET /projects/{id}/documents returns documents
**Title:** List project documents
**Description:** Verify `GET /projects/{id}/documents` returns documentation for a project.
**Preconditions:** Project with UUID exists and has 2 indexed documents.
**Steps:**
1. Send `GET /projects/{existing-uuid}/documents`
**Expected Result:** HTTP 200 with JSON array of Document entities containing `path`, `content`, `hash`, `discovered_at`.
**Story:** 11.2

---

### TC-11.2-006: GET /projects/{id}/documents returns empty array for project with no docs
**Title:** Project documents empty
**Description:** Verify that a project with no indexed documents returns an empty array.
**Preconditions:** Project with UUID exists but has no documents indexed.
**Steps:**
1. Send `GET /projects/{existing-uuid}/documents`
**Expected Result:** HTTP 200 with JSON `[]`.
**Story:** 11.2

---

### TC-11.2-007: GET /projects/{id}/analysis returns analysis
**Title:** Get project analysis
**Description:** Verify `GET /projects/{id}/analysis` returns AI-generated analysis if available.
**Preconditions:** Project has been analyzed by an AI agent. Analysis exists in knowledge store.
**Steps:**
1. Send `GET /projects/{existing-uuid}/analysis`
**Expected Result:** HTTP 200 with JSON Analysis entity containing `summary`, `purpose`, `analyzed_at`, `analyzer`, etc.
**Story:** 11.2

---

### TC-11.2-008: GET /projects/{id}/analysis returns 404 when no analysis exists
**Title:** Analysis not found
**Description:** Verify that requesting analysis for a project that has not been analyzed returns 404.
**Preconditions:** Project exists but has never been analyzed (null analysis).
**Steps:**
1. Send `GET /projects/{existing-uuid}/analysis`
**Expected Result:** HTTP 404 with JSON body `{"error": "not_found", "message": "analysis not found"}`.
**Story:** 11.2

---

### TC-11.2-009: GET /relationships/{projectId} returns relationships
**Title:** Get project relationships
**Description:** Verify `GET /relationships/{projectId}` returns relationships for a project.
**Preconditions:** Project exists and has 2 relationships defined.
**Steps:**
1. Send `GET /relationships/{existing-uuid}`
**Expected Result:** HTTP 200 with JSON array of Relationship entities containing `type`, `targetProjectId`, `description`.
**Story:** 11.2

---

### TC-11.2-010: GET /statistics returns portfolio statistics
**Title:** Portfolio statistics
**Description:** Verify `GET /statistics` returns aggregated portfolio statistics.
**Preconditions:** Knowledge store contains multiple projects with varied metadata.
**Steps:**
1. Send `GET /statistics`
**Expected Result:** HTTP 200 with JSON containing total project count, language breakdown, framework breakdown, documentation counts, etc.
**Story:** 11.2

---

### TC-11.2-011: GET /configuration returns current configuration
**Title:** Get configuration
**Description:** Verify `GET /configuration` returns the current Portfolio engine configuration.
**Preconditions:** Engine configured with non-default values.
**Steps:**
1. Send `GET /configuration`
**Expected Result:** HTTP 200 with JSON body containing all configuration fields (asset path, port, DB path, etc.).
**Story:** 11.2

---

### TC-11.2-012: PATCH /configuration accepts partial updates
**Title:** Update configuration
**Description:** Verify `PATCH /configuration` accepts a partial JSON payload and returns updated configuration.
**Preconditions:** Engine running with default configuration.
**Steps:**
1. Send `PATCH /configuration` with body `{"asset_path": "/new/path"}`
**Expected Result:** HTTP 200 with full JSON configuration showing `asset_path` updated. Other fields unchanged.
**Story:** 11.2

---

### TC-11.2-013: PATCH /configuration with unknown key returns 400
**Title:** Reject unknown config keys
**Description:** Verify that patching with an unknown configuration key returns 400.
**Preconditions:** Engine running.
**Steps:**
1. Send `PATCH /configuration` with body `{"unknown_key": "value"}`
**Expected Result:** HTTP 400 with JSON body `{"error": "bad_request", "message": "unknown configuration key: unknown_key"}`.
**Story:** 11.2

---

### TC-11.2-014: PATCH /configuration with wrong value type returns 400
**Title:** Reject wrong value type
**Description:** Verify that patching with an incorrect value type returns 400.
**Preconditions:** Engine running. `page_size` is an integer field.
**Steps:**
1. Send `PATCH /configuration` with body `{"page_size": "not-a-number"}`
**Expected Result:** HTTP 400 with JSON body `{"error": "bad_request", "message": "invalid type for field 'page_size': expected integer"}`.
**Story:** 11.2

---

### TC-11.2-015: CORS headers present for configured origins
**Title:** CORS headers on API responses
**Description:** Verify that CORS headers are set allowing local development origins.
**Preconditions:** Engine configured with CORS origins `http://localhost:3000`.
**Steps:**
1. Send `OPTIONS /projects` with `Origin: http://localhost:3000`
2. Send `GET /projects` with `Origin: http://localhost:3000`
**Expected Result:** Preflight returns 204 with `Access-Control-Allow-Origin: http://localhost:3000`. GET response includes `Access-Control-Allow-Origin` header.
**Story:** 11.2

---

### TC-11.2-016: No authentication required
**Title:** Authentication not required
**Description:** Verify that API endpoints are accessible without any authentication.
**Preconditions:** Engine running.
**Steps:**
1. Send `GET /projects` with no auth headers
2. Send `GET /health` with no auth headers
**Expected Result:** Both requests return 200 (no 401/403).
**Story:** 11.2

---

### TC-11.2-017: Method not allowed returns 405
**Title:** Reject unsupported methods
**Description:** Verify that sending an unsupported HTTP method returns 405 with Allow header.
**Preconditions:** Engine running.
**Steps:**
1. Send `POST /projects`
2. Send `DELETE /projects/{id}`
3. Send `PUT /configuration`
**Expected Result:** Each returns HTTP 405 with `Allow` header listing permitted methods (e.g., `Allow: GET, OPTIONS`).
**Story:** 11.2

---

### TC-11.2-018: Unknown API path returns 404
**Title:** Unknown path returns 404
**Description:** Verify that requesting an undefined API path returns 404.
**Preconditions:** Engine running.
**Steps:**
1. Send `GET /api/undefined-endpoint`
**Expected Result:** HTTP 404 with JSON body `{"error": "not_found", "message": "endpoint not found"}`.
**Story:** 11.2

---

### TC-11.2-019: Error responses use consistent JSON shape
**Title:** Consistent error response format
**Description:** Verify all error responses share the same JSON structure.
**Preconditions:** Engine running.
**Steps:**
1. Send `GET /projects/invalid-uuid`
2. Send `POST /projects`
3. Send `GET /projects/00000000-0000-0000-0000-000000000000/documents`
**Expected Result:** All error responses contain `{"error": "<error_code>", "message": "<human-readable message>"}`. No HTML error pages.
**Story:** 11.2

---

### TC-11.2-020: Database unavailable returns 500
**Title:** Database connection error
**Description:** Verify that when the SQLite database is unavailable, endpoints return 500.
**Preconditions:** Knowledge store connection is broken or database file is missing.
**Steps:**
1. Simulate database connection failure
2. Send `GET /projects`
**Expected Result:** HTTP 500 with JSON body `{"error": "internal_error", "message": "database unavailable"}`.
**Story:** 11.2

---

### TC-11.2-021: All responses use canonical knowledge model
**Title:** Canonical model conformance
**Description:** Verify that all API responses use entities from KnowledgeModel.md with no custom DTOs.
**Preconditions:** Engine running with sample data.
**Steps:**
1. Fetch `GET /projects`
2. Fetch `GET /projects/{id}`
3. Fetch `GET /projects/{id}/documents`
4. Validate each entity shape against KnowledgeModel.md
**Expected Result:** Every JSON field maps to a field defined in KnowledgeModel.md. No custom or synthetic fields are present.
**Story:** 11.2

---

### TC-11.2-022: GET /health includes uptime and DB status
**Title:** Health response detail
**Description:** Verify health endpoint includes uptime duration and database connection status.
**Preconditions:** Engine has been running for at least 2 seconds.
**Steps:**
1. Send `GET /health`
**Expected Result:** JSON body includes `uptime` as a duration string (e.g., `"2.5s"`) and `database` as `"connected"`.
**Story:** 11.2

---

## Story 11.3 — Search Endpoints

### TC-11.3-001: Basic search returns matching results
**Title:** Unified search by keyword
**Description:** Verify `GET /search?q=react` returns matching projects and documents.
**Preconditions:** Knowledge store contains a project with "React" in its name/description and a document mentioning "React".
**Steps:**
1. Send `GET /search?q=react`
**Expected Result:** HTTP 200 with JSON containing `results` array. Results include matching projects and documents. Each result is relevant to the query term.
**Story:** 11.3

---

### TC-11.3-002: Search across projects and documents
**Title:** Cross-entity search
**Description:** Verify search results span both Project and Document entities.
**Preconditions:** Store contains project named "react-app" and unrelated project with a document mentioning "react hooks".
**Steps:**
1. Send `GET /search?q=react`
**Expected Result:** Results array includes both the "react-app" Project and the Document containing "react hooks". Result items include a `type` field indicating the entity kind.
**Story:** 11.3

---

### TC-11.3-003: Missing q parameter returns 400
**Title:** Missing query parameter
**Description:** Verify that search without `q` parameter returns 400.
**Preconditions:** Engine running.
**Steps:**
1. Send `GET /search`
2. Send `GET /search?q=`
**Expected Result:** HTTP 400 with JSON body `{"error": "bad_request", "message": "query parameter 'q' is required"}`.
**Story:** 11.3

---

### TC-11.3-004: No matches return empty results
**Title:** Empty search results
**Description:** Verify that a query with no matches returns an empty results array (not an error).
**Preconditions:** Knowledge store contains data, but no records match "zzzznonexistentzzzz".
**Steps:**
1. Send `GET /search?q=zzzznonexistentzzzz`
**Expected Result:** HTTP 200 with JSON `{"total_results": 0, "results": []}`.
**Story:** 11.3

---

### TC-11.3-005: Technology filter narrows results
**Title:** Filter by technology
**Description:** Verify `?technology=react` filters results to projects using React.
**Preconditions:** Store contains a React project and a Vue project. Searching "component" returns both without filter.
**Steps:**
1. Search `GET /search?q=component&technology=react`
**Expected Result:** Results only include projects/documents associated with React technology. Vue-related results are excluded.
**Story:** 11.3

---

### TC-11.3-006: Framework filter narrows results
**Title:** Filter by framework
**Description:** Verify `?framework=gin` filters results to projects using Gin framework.
**Preconditions:** Store contains a Gin project and an Echo project.
**Steps:**
1. Search `GET /search?q=api&framework=gin`
**Expected Result:** Results only include projects/documents associated with Gin framework.
**Story:** 11.3

---

### TC-11.3-007: Date range filter narrows results
**Title:** Filter by date range
**Description:** Verify `?from=...&to=...` filters results by discovered_at / last_scan_at.
**Preconditions:** Store contains project discovered on 2024-01-15 and another on 2024-06-20.
**Steps:**
1. Search `GET /search?q=project&from=2024-01-01&to=2024-03-01`
**Expected Result:** Results only include the project discovered on 2024-01-15.
**Story:** 11.3

---

### TC-11.3-008: Invalid date format returns 400
**Title:** Invalid date format rejected
**Description:** Verify that non-ISO 8601 date values in `from`/`to` return 400.
**Preconditions:** Engine running.
**Steps:**
1. Send `GET /search?q=test&from=01-15-2024`
2. Send `GET /search?q=test&from=not-a-date`
**Expected Result:** HTTP 400 with JSON body `{"error": "bad_request", "message": "invalid date format — use ISO 8601"}`.
**Story:** 11.3

---

### TC-11.3-009: Pagination returns correct slice
**Title:** Basic pagination
**Description:** Verify that `page` and `page_size` parameters return the correct slice of results.
**Preconditions:** Knowledge store has 50 searchable items matching "project".
**Steps:**
1. Search `GET /search?q=project&page=1&page_size=10`
2. Search `GET /search?q=project&page=2&page_size=10`
**Expected Result:** Page 1 returns results 1-10. Page 2 returns results 11-20. No overlap between sets.
**Story:** 11.3

---

### TC-11.3-010: Pagination metadata in response
**Title:** Pagination metadata
**Description:** Verify paginated response includes all required metadata.
**Preconditions:** Knowledge store has 45 matching results for "project".
**Steps:**
1. Send `GET /search?q=project&page=2&page_size=10`
**Expected Result:** Response JSON includes `total_results: 45`, `page: 2`, `page_size: 10`, `total_pages: 5`, and `results` array with 10 items.
**Story:** 11.3

---

### TC-11.3-011: Default pagination values
**Title:** Default pagination
**Description:** Verify that requests without `page`/`page_size` use defaults (page=1, page_size=20).
**Preconditions:** Store has 50 matching results.
**Steps:**
1. Send `GET /search?q=project`
**Expected Result:** Response shows `page: 1`, `page_size: 20`, `total_pages: 3`, results array has 20 items.
**Story:** 11.3

---

### TC-11.3-012: page_size respects maximum cap
**Title:** Page size cap enforced
**Description:** Verify that `page_size` values exceeding 100 are clamped to 100.
**Preconditions:** Store has 500 matching results.
**Steps:**
1. Send `GET /search?q=project&page_size=1000`
**Expected Result:** Response shows `page_size: 100`. Results array has at most 100 items.
**Story:** 11.3

---

### TC-11.3-013: Negative pagination values return 400
**Title:** Invalid pagination values
**Description:** Verify negative page/page_size values are rejected.
**Preconditions:** Engine running.
**Steps:**
1. Send `GET /search?q=test&page=-1`
2. Send `GET /search?q=test&page_size=-5`
**Expected Result:** HTTP 400 with JSON body `{"error": "bad_request", "message": "page and page_size must be positive integers"}`.
**Story:** 11.3

---

### TC-11.3-014: Non-numeric pagination values return 400
**Title:** Non-numeric pagination
**Description:** Verify non-numeric page/page_size values return 400.
**Preconditions:** Engine running.
**Steps:**
1. Send `GET /search?q=test&page=abc`
2. Send `GET /search?q=test&page_size=two`
**Expected Result:** HTTP 400 with JSON body `{"error": "bad_request", "message": "page and page_size must be positive integers"}`.
**Story:** 11.3

---

### TC-11.3-015: Highlighted snippets present in results
**Title:** Snippet highlighting
**Description:** Verify that document search results include highlighted snippets showing matching context.
**Preconditions:** A document contains the sentence "This project uses React for the frontend."
**Steps:**
1. Send `GET /search?q=react`
**Expected Result:** Each matching document result includes a `snippet` field with the query term emphasized. The match context is shown with surrounding text. Highlighting uses HTML `<span>` tags or plain text markers.
**Story:** 11.3

---

### TC-11.3-016: Combined technology + date filter
**Title:** Multi-filter search
**Description:** Verify combining technology and date range filters narrows results correctly.
**Preconditions:** React project discovered 2024-01-15. React project discovered 2024-06-20. Vue project discovered 2024-03-01.
**Steps:**
1. Send `GET /search?q=project&technology=react&from=2024-01-01&to=2024-04-01`
**Expected Result:** Only the React project discovered 2024-01-15 is returned.
**Story:** 11.3

---

### TC-11.3-017: Search is deterministic
**Title:** Deterministic search results
**Description:** Verify that the same query on the same data produces identical results every time.
**Preconditions:** Knowledge store is static (no concurrent modifications).
**Steps:**
1. Send `GET /search?q=react&page=1&page_size=10`
2. Wait 100ms
3. Send `GET /search?q=react&page=1&page_size=10`
**Expected Result:** Both responses have identical `total_results`, `total_pages`, and `results` arrays (same order, same IDs).
**Story:** 11.3

---

### TC-11.3-018: Search term with special characters
**Title:** Special characters in query
**Description:** Verify search handles special characters and SQL injection attempts safely.
**Preconditions:** Engine running.
**Steps:**
1. Send `GET /search?q=%27%20OR%20%271%27%3D%271` (SQL injection attempt: `' OR '1'='1`)
2. Send `GET /search?q=<script>alert(1)</script>`
3. Send `GET /search?q=foo%22bar`
**Expected Result:** All requests return HTTP 200 with properly escaped/sanitized queries. No SQL errors, no XSS injection. Special characters are either stripped or treated as literal search terms.
**Story:** 11.3

---

### TC-11.3-019: Case-insensitive search
**Title:** Case-insensitive matching
**Description:** Verify search is case-insensitive.
**Preconditions:** Store contains project named "ReactDashboard".
**Steps:**
1. Send `GET /search?q=reactdashboard`
2. Send `GET /search?q=REACTDASHBOARD`
3. Send `GET /search?q=ReactDashboard`
**Expected Result:** All three queries return identical result sets including the matching project.
**Story:** 11.3

---

### TC-11.3-020: Partial word matching
**Title:** Partial word matching
**Description:** Verify search matches partial words via FTS.
**Preconditions:** Store contains a document mentioning "microservices".
**Steps:**
1. Send `GET /search?q=micro`
**Expected Result:** Returns the document containing "microservices" due to FTS prefix matching.
**Story:** 11.3

---

### TC-11.3-021: Combined framework + date filter
**Title:** Framework and date range filter
**Description:** Verify combining framework and date range filters works correctly.
**Preconditions:** Gin project discovered 2024-01-15. Gin project discovered 2024-06-20. Echo project discovered 2024-03-01.
**Steps:**
1. Send `GET /search?q=api&framework=gin&from=2024-01-01&to=2024-04-01`
**Expected Result:** Only the Gin project discovered 2024-01-15 is returned.
**Story:** 11.3

---

## Integration & Cross-Cutting

### TC-11-INT-001: Full stack health → search → project detail flow
**Title:** End-to-end dashboard data flow
**Description:** Verify a complete user flow from health check through search to project detail.
**Preconditions:** Engine running with a portfolio of 5+ projects.
**Steps:**
1. `GET /health` — verify engine is healthy
2. `GET /search?q=react` — find React projects
3. Pick a project ID from results
4. `GET /projects/{id}` — get project detail
5. `GET /projects/{id}/documents` — get project documents
**Expected Result:** All requests succeed (200). Data flows coherently: the project ID from search step corresponds to data in detail and documents steps.
**Story:** 11.1, 11.2, 11.3

---

### TC-11-INT-002: SPA route does not interfere with API routes
**Title:** SPA fallback isolation
**Description:** Verify that SPA fallback routing does not intercept API paths.
**Preconditions:** Engine running with dashboard assets.
**Steps:**
1. `GET /projects` — should return JSON
2. `GET /api/projects` — does not exist, should return 404 (not index.html)
3. `GET /search?q=test` — should return JSON
**Expected Result:** API paths return JSON responses (not HTML). Only non-file, non-API paths trigger SPA fallback.
**Story:** 11.1, 11.2

---

### TC-11-INT-003: Graceful shutdown drains in-flight requests
**Title:** Graceful shutdown
**Description:** Verify that the HTTP server drains in-flight requests before shutting down.
**Preconditions:** Engine running. A long-running search query is configured to take 500ms.
**Steps:**
1. Send a search request that takes 500ms
2. Immediately send SIGTERM or call shutdown
3. Wait for shutdown to complete
**Expected Result:** The in-flight request completes successfully. New requests after shutdown signal receive connection refusal. Server exits cleanly within configured timeout.
**Story:** 11.1, 11.2, 11.3

---

### TC-11-INT-004: Concurrent server start fails fast
**Title:** Prevent duplicate server instances
**Description:** Verify that starting a second instance on the same port fails with a clear error.
**Preconditions:** First engine instance is running on port 8090.
**Steps:**
1. Attempt to start a second engine instance on port 8090
**Expected Result:** Second instance exits or logs a clear error: `address already in use` or `port 8090 is already in use`.
**Story:** 11.1, 11.2, 11.3

---

### TC-11-INT-005: Configurable host and port
**Title:** Custom host and port binding
**Description:** Verify that the HTTP server binds to a configured host and port.
**Preconditions:** Engine configured with `host: 127.0.0.1`, `port: 9090`.
**Steps:**
1. Start engine with custom host/port
2. Send `GET /health` to `http://127.0.0.1:9090`
3. Verify the default `localhost:8090` is not bound
**Expected Result:** Health endpoint responds on `127.0.0.1:9090`. No response on `localhost:8090`.
**Story:** 11.1, 11.2, 11.3

---

### TC-11-INT-006: No mutation through any dashboard endpoint
**Title:** Read-only enforcement
**Description:** Verify that no dashboard endpoint mutates the knowledge store.
**Preconditions:** Engine running. Knowledge store contains known state.
**Steps:**
1. Record current project count and latest modification timestamp
2. Call `GET /projects`
3. Call `GET /search?q=test`
4. Call `GET /health`
5. Re-read project count and modification timestamp
**Expected Result:** State is identical before and after. No `POST`, `PUT`, `DELETE`, or `PATCH` operations exist on dashboard-facing endpoints (except `PATCH /configuration`).
**Story:** 11.1, 11.2, 11.3

---

### TC-11-INT-007: Large portfolio performance (500 projects)
**Title:** Performance at scale
**Description:** Verify API responses meet performance targets for a portfolio of 500 projects.
**Preconditions:** Knowledge store seeded with 500 projects and 2000+ documents.
**Steps:**
1. Measure response time for `GET /projects`
2. Measure response time for `GET /search?q=a` (broad match)
3. Measure response time for `GET /statistics`
**Expected Result:** Static asset responses <10ms. API/search responses <100ms for typical queries.
**Story:** 11.1, 11.2, 11.3

---

### TC-11-INT-008: No AI module imports or dependencies
**Title:** No AI dependency
**Description:** Verify that the dashboard backend package never imports or depends on any AI/LLM module.
**Preconditions:** Source code available for static analysis.
**Steps:**
1. Run `go list -json ./...` on the dashboard backend package
2. Check all transitive dependencies
3. Grep for imports of any AI/LLM-related Go modules
**Expected Result:** No AI/LLM modules in dependency graph. No imports of agent, LLM, OpenAI, Anthropic, or similar packages.
**Story:** 11.1, 11.2, 11.3

---

### TC-11-INT-009: All endpoints return Content-Type: application/json
**Title:** JSON content type on API responses
**Description:** Verify all API endpoints return `Content-Type: application/json`.
**Preconditions:** Engine running.
**Steps:**
1. Check `Content-Type` header on `GET /health`
2. Check `Content-Type` header on `GET /projects`
3. Check `Content-Type` header on `GET /search?q=test`
4. Check `Content-Type` header on 400/404/500 error responses
**Expected Result:** All return `Content-Type: application/json`.
**Story:** 11.2, 11.3

---

### TC-11-INT-010: Safe concurrent SQLite access (WAL mode)
**Title:** Concurrent read safety
**Description:** Verify that concurrent read operations do not block or error.
**Preconditions:** Knowledge store uses WAL mode. Engine running.
**Steps:**
1. Send 10 simultaneous `GET /projects` requests
2. Send 10 simultaneous `GET /search?q=a` requests
**Expected Result:** All 20 requests complete successfully with HTTP 200. No `database is locked` errors. No request exceeds 200ms.
**Story:** 11.2, 11.3
