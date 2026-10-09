---
title: API Overview
description: Portfolio's localhost HTTP API — conventions, CORS, error format, and content types.
---

# API Overview

Portfolio ships a small, read-mostly HTTP API over the [knowledge store](/docs/concepts/knowledge-store). It backs the dashboard and is handy for scripts and integrations.

## Starting the server

```bash
portfolio serve            # http://127.0.0.1:8080
portfolio serve -p 9000
```

The server binds to **`127.0.0.1` only** — it is not reachable from other machines.

## Conventions

- **Base URL**: `http://127.0.0.1:8080`
- **Content type**: all responses are `application/json`
- **No authentication**: the server is localhost-only by design; it trusts local users because nothing remote can reach it. Do not expose it behind a proxy.
- **Method set**: `GET` for reads, plus `PATCH /configuration` and `POST /relationships/{id}` for writes

## CORS

The CORS layer implements Portfolio's local-first security model:

- Requests **without** an `Origin` header (curl, scripts, server-side clients) are always allowed through.
- Browser requests with an `Origin` are only answered when the origin is a **localhost** origin (`localhost`, `127.0.0.1`, or `[::1]`, any port, http or https).
- Allowed methods: `GET`, `PATCH`, `OPTIONS`. Allowed headers: `Content-Type`, `Authorization`.
- The `[dashboard] allowed_origins` config lists the origins the embedded dashboard expects.

## Error format

Errors return a consistent JSON envelope:

```json
{
  "error": "project not found",
  "code": "NOT_FOUND",
  "type": "resource_error"
}
```

| HTTP status | `code` | `type` |
| --- | --- | --- |
| 400 | `INVALID_INPUT` | `validation_error` |
| 404 | `NOT_FOUND` | `resource_error` |
| 500 | *(omitted)* | `api_error` |
| 503 (unhealthy DB) | *(omitted)* | `api_error` |

Every request is logged (method, path, duration) to the configured log file.

## Endpoint index

| Endpoint | Description |
| --- | --- |
| `GET /health` | Service and database health |
| `GET /projects` | List projects (filter, sort, paginate) |
| `GET /projects/{id}` | Project detail |
| `GET /projects/{id}/analysis` | Project analysis |
| `GET /search` | Portfolio-wide search |
| `GET /configuration` / `PATCH /configuration` | Read / update runtime configuration |
| `GET /statistics` | Portfolio-wide statistics |
| `GET /technologies` | Technology inventory |
| `GET /relationships` | All project relationships |
| `GET /relationships/{id}` | Relationships for one project |
| `POST /relationships/{id}` | Store a relationship |

Full request/response details: [Endpoints](/docs/api/endpoints).
