---
title: Endpoints
description: Request and response reference for every Portfolio HTTP API endpoint.
---

# Endpoints

All endpoints are served from `http://127.0.0.1:8080` (see [`portfolio serve`](/docs/cli/servers#portfolio-serve) for flags). Every response is `application/json`; errors use the [standard error envelope](/docs/api/).

## Health

### `GET /health`

Liveness plus database connectivity.

```bash
curl http://127.0.0.1:8080/health
```

```json
{
  "status": "healthy",
  "database_connected": true,
  "project_count": 42
}
```

Returns `503` with `"status": "unhealthy"` when the database is unreachable.

## Projects

### `GET /projects`

List projects with optional filtering, sorting, and pagination.

| Parameter | Description |
| --- | --- |
| `q` | Free-text filter on project name |
| `sort` | Sort order for the result set |
| `limit` | Maximum number of projects to return |
| `offset` | Number of projects to skip |

```bash
curl "http://127.0.0.1:8080/projects?q=billing&limit=10&offset=0"
```

```json
{
  "projects": [
    {
      "id": "5f0c…",
      "name": "billing-api",
      "root_path": "/Users/you/Projects/billing-api",
      "repository_type": "git",
      "discovered_at": "2026-09-30T10:12:00Z",
      "updated_at": "2026-10-03T18:44:00Z",
      "metadata": { "…": "git, languages, frameworks, capabilities" }
    }
  ],
  "total": 42
}
```

### `GET /projects/{id}`

One project with full metadata, and (where present) its indexed documents and analyses.

```bash
curl http://127.0.0.1:8080/projects/5f0c…
```

The single-project response is a `projectResponse` including `metadata`, `documents`, and `analyses` arrays when available. Unknown IDs return `404 NOT_FOUND`.

### `GET /projects/{id}/analysis`

Stored analyses for a project.

```json
{
  "project_id": "5f0c…",
  "analyses": [ { "…": "analysis payloads" } ],
  "count": 2
}
```

## Search

### `GET /search`

Portfolio-wide search across projects **and** indexed documentation.

| Parameter | Description |
| --- | --- |
| `q` | Query text (max 500 characters) |
| `technology` | Repeatable filter, e.g. `technology[]=go&technology[]=postgres` |
| `framework` | Repeatable framework filter |
| `page` | 1-based page number |
| `page_size` | Results per page |

```bash
curl "http://127.0.0.1:8080/search?q=rate+limiting&page_size=5"
```

```json
{
  "results": [
    {
      "type": "document",
      "id": "…",
      "kind": "adr",
      "path": "/Users/you/Projects/gateway/docs/ADR.md",
      "content": "…matched snippet…",
      "rank": 0.87,
      "project": { "id": "5f0c…", "name": "gateway" }
    }
  ],
  "totalCount": 12,
  "page": 1,
  "pageSize": 5,
  "totalPages": 3
}
```

Result `type` distinguishes projects from documents; `rank` orders by relevance.

## Configuration

### `GET /configuration`

Current runtime configuration state.

### `PATCH /configuration`

Update supported configuration values. Send a JSON body with the keys to change; the response echoes the updated configuration.

## Statistics

### `GET /statistics`

Portfolio-wide aggregates used by the dashboard's statistics page.

```json
{
  "total_projects": 42,
  "projects_with_metadata": 40,
  "projects_with_analysis": 18,
  "projects_with_documents": 37,
  "total_documents": 1240,
  "total_dependencies": 3911,
  "total_relationships": 64,
  "language_counts": { "go": 18, "typescript": 12, "php": 6 },
  "framework_counts": { "react": 9, "express": 5 },
  "technology_counts": { "postgres": 14, "redis": 8 },
  "document_kind_counts": { "readme": 40, "adr": 22, "openapi": 6 }
}
```

## Technologies

### `GET /technologies`

The technology inventory across your portfolio — the registry behind [technology filtering in search](#get-search) and [project tagging](/docs/mcp/tools#technologies).

## Relationships

Relationships are typed, confidence-scored edges between projects (`source_project` → `target_project`). Typical types include `Similar`, `Evolution`, `Shared Feature`, `Shared Technology`, and `Reuses Component`.

### `GET /relationships`

All relationships in the portfolio:

```json
[
  {
    "source_project": "5f0c…",
    "source_project_name": "billing-api",
    "target_project": "9a41…",
    "target_project_name": "payments-sdk",
    "type": "Reuses Component",
    "description": "Shares the payments client library",
    "confidence": 0.9
  }
]
```

### `GET /relationships/{id}`

Relationships for a single project (the `{id}` is the source project ID).

### `POST /relationships/{id}`

Store a relationship originating from a project. Writes typically come from AI agents via MCP (`storeRelationship`) — see [Tools Reference](/docs/mcp/tools#relationships).

```bash
curl -X POST http://127.0.0.1:8080/relationships/5f0c… \
  -H "Content-Type: application/json" \
  -d '{
    "target_project": "9a41…",
    "type": "Shared Technology",
    "description": "Both use the internal auth middleware",
    "confidence": 0.85
  }'
```

Returns `201 Created` with the stored relationship.
