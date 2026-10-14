---
title: The Knowledge Store
description: Portfolio's SQLite knowledge base — schema, full-text search, and change tracking.
---

# The Knowledge Store

Everything Portfolio learns is persisted in a single SQLite database — the *knowledge store*. It is the source of truth for the CLI, the HTTP API, the dashboard, and the MCP server.

## Pure-Go SQLite

Portfolio embeds SQLite via a pure-Go driver (`modernc.org/sqlite`), so the binary is fully static: **no CGO, no system SQLite dependency**. The database ships inside the same executable that creates and queries it.

## Location

The database lives in your platform data directory under an [obscured filename](/docs/digging-deeper/security#obscured-database-filename) (`.portfoliodata` by default), alongside the key file and logs:

```
<platform-data-dir>/com.portfolio.cli/
├── .portfoliodata    # the knowledge store
├── db_key            # database key file
├── portfolio.log     # application log
├── error.log         # error log
└── integrations/     # per-agent integration state
```

Legacy installs that used `~/.portfolio/portfolio.db` are migrated automatically.

## Schema overview

The store is organized around a few core tables:

| Table | Purpose |
| --- | --- |
| `projects` | One row per discovered repository, keyed by `root_path` |
| `metadata` | Git intelligence, language/framework summaries, capability signals |
| `documents` + `documents_fts` | Indexed documentation with an FTS5 full-text index |
| `analyses` | Stored analysis results |
| `features` | Extracted feature inventory per project |
| `technologies`, `project_technologies` | Technology registry and per-project tagging |
| `relationships` | Typed edges between projects |
| `dependencies` | Detected dependencies |
| `configuration` | Runtime configuration state |
| `schema_migrations` | Embedded, versioned migrations |

Migrations are embedded in the binary and applied on connect — `portfolio init` or the first `serve`/`dashboard` start brings the schema fully up to date. There is no separate migration step.

## Full-text search

Documents are indexed into an SQLite **FTS5** virtual table. Searches return matched documents with snippets, across READMEs, ADRs, architecture docs, OpenAPI specs, and changelogs of every indexed project at once.

Query it from anywhere:

```bash
# CLI
portfolio projects search "rate limiting"

# HTTP API
curl "http://127.0.0.1:8080/search?q=rate+limiting"

# MCP (from your AI agent)
searchDocumentation({ query: "rate limiting" })
```

## Change tracking

Every indexed document carries a content hash. On subsequent scans, unchanged documents are skipped and only modified files are re-indexed — which keeps re-scans fast even across large portfolios.

## Backup and portability

The knowledge store is a single file. To back it up or move it to another machine, stop any running server and copy the `.portfoliodata` file (and `db_key` if you use key-file encryption). The receiving machine needs Portfolio installed and the config pointing at the copied path.
