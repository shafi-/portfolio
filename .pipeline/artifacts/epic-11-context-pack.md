# Context Pack: Epic 11 — Dashboard Backend

## Project Identity
Portfolio is a local-first project inventory and knowledge platform. Go engine + SQLite + MCP + HTTP API + read-only dashboard. AI agents are the primary interface; dashboard is read-only exploration.

## Principles (Guideline.md)
- **Engine Knows, Agent Thinks** — Engine does deterministic work only. No AI reasoning in engine.
- **Deterministic by Default** — Repeatable results for same input.
- **Store Facts, Compute Indicators** — Persist git HEAD, hashes, timestamps. Compute needs_analysis, outdated on read.
- **Local First** — All data on user machine. No cloud required.
- **Dashboard is Read-only** — Never invokes AI, never modifies repos, never performs analysis.
- **Single Knowledge Model** — Every interface uses KnowledgeModel.md entities.

## Architecture
```
Dashboard ←→ HTTP API ←→ Portfolio Engine → SQLite → Local Repos
```
Components: Portfolio Engine (Go), HTTP API, Dashboard (read-only), Local Knowledge Store (SQLite).

## Tech Stack
- **Engine:** Go
- **Database:** SQLite
- **Dashboard backend:** Part of Go engine (asset serving + API wrapping)
- **Dashboard frontend:** Separate SPA (Epic 12)

## Knowledge Model (Key Entities)
| Entity | Owner | Key Fields |
|--------|-------|------------|
| Project | Engine | id (UUID), name, root_path, repository_type, discovered_at |
| Metadata | Engine | git_head, default_branch, language_summary, framework_summary, documentation_hash, last_scan_at |
| Documentation | Engine | id, project_id, path, kind (README/ADR/DOC/CHANGELOG), content, content_hash |
| Analysis | Agent | id, project_id, analyzer, summary, purpose, architecture, notes, analyzed_git_head |
| Technology | Engine | id, name, category |
| Relationship | Agent | id, source_project, target_project, type, description, confidence |

## HTTP API (PlatformSpecification.md §3)
| Endpoint | Method | Description |
|----------|--------|-------------|
| /health | GET | Health check |
| /projects | GET | List all projects |
| /projects/{id} | GET | Single project detail |
| /projects/{id}/documents | GET | Project documents |
| /projects/{id}/analysis | GET | Project analysis |
| /search?q= | GET | Unified search |
| /relationships/{projectId} | GET | Project relationships |
| /statistics | GET | Portfolio-wide stats |
| /configuration | GET | Get configuration |
| /configuration | PATCH | Update configuration |

## Dashboard Specification (PlatformSpecification.md §5)
**Read-only.** Pages: Portfolio Overview, Project List (search/filters/sorting), Project Detail (metadata/docs/analysis/relationships), Relationship Explorer, Statistics.
**Never invokes AI. Never modifies knowledge.**

## Epic 11 Stories

| Story | Size | Blocked By | Summary |
|-------|------|------------|---------|
| 11.1 Asset Serving | S | Epic 6 | Serve dashboard SPA assets (HTML/CSS/JS) from embed or disk. MIME types, cache headers, SPA fallback, directory traversal prevention. |
| 11.2 Endpoint Review | S | Epic 6 | Wire Epic 6 handlers for dashboard. Add search filters (technology/framework/date), pagination, highlighted snippets. CORS. Consistent error format. Verify health/config endpoints. |

## Dependencies
- Epic 6 (HTTP API) — provides core handlers and search infrastructure
- Dashboard frontend build output (Epic 12) — static assets to serve

## Key Constraints
- Dashboard consumes HTTP only (agents consume MCP — separate concern)
- Dashboard backend lives in Go engine process
- No authentication (local-only)
- All responses use canonical knowledge model
- Deterministic responses only
- No AI invocation, no mutations
