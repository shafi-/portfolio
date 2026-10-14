---
title: Architecture
description: How Portfolio is built — the engine, its packages, and the surfaces on top.
---

# Architecture

Portfolio is a single Go binary with a layered architecture: deterministic engine at the bottom, surfaces (CLI, API, dashboard, MCP) on top. Everything reads from one shared knowledge store.

## The big picture

```
┌────────────────────────────────────────────────────────────┐
│                        Surfaces                            │
│                                                            │
│   CLI (Cobra)    HTTP API    Dashboard (React,     MCP     │
│   portfolio …    (net/http)  embedded in binary)   (stdio) │
└──────────┬─────────────┬──────────────┬─────────────┬──────┘
           │             │              │             │
┌──────────▼─────────────▼──────────────▼─────────────▼──────┐
│                       Engine (internal/)                   │
│                                                            │
│   discovery   →  metadata   →  indexer      →  store      │
│   find repos     extract       index docs      persist     │
└──────────┬─────────────────────────────────────────────────┘
           │
┌──────────▼─────────────────────────────────────────────────┐
│          Knowledge Store (embedded SQLite + FTS5)          │
└────────────────────────────────────────────────────────────┘
```

## Repository layout

| Path | Contents |
| --- | --- |
| `cmd/portfolio/` | Binary entry point |
| `internal/cli/` | Cobra command tree |
| `internal/api/` | HTTP API server and handlers |
| `internal/mcp/` | MCP server and tool registrations |
| `internal/dashboard/` | Embedded dashboard asset server + router |
| `internal/discovery/` | Project discovery engine |
| `internal/metadata/` | Metadata extraction |
| `internal/indexer/` | Documentation indexing |
| `internal/store/` | Typed stores per aggregate |
| `internal/database/` | SQLite connection + embedded migrations |
| `internal/integration/` | Claude Code / OpenCode integration installers |
| `internal/config/` | TOML configuration provider |
| `internal/logging/` | Structured logging (zap), MCP-safe |
| `pkg/models/` | Shared domain types (`Project`, `Config`, …) |
| `dashboard/` | React 19 + Vite + Tailwind dashboard source |
| `migrations/` | SQL migration sources (embedded at build) |

## Key decisions

### Engine knows, agent thinks

All deterministic work lives in the engine. AI agents never discover or parse anything themselves — they query the engine through MCP tools and reason over the results.

### Everything reads one store

The CLI, API, dashboard, and MCP server all sit on the same SQLite knowledge store. There is no cache to invalidate and no sync step: scan once, and every surface sees the new data.

### The dashboard is embedded

The React dashboard is built at development time and embedded into the binary with `go:embed`. `portfolio dashboard` serves the compiled assets — the binary remains fully self-contained.

### Localhost-only by design

The API binds to `127.0.0.1` and its CORS layer only ever allows localhost origins. Portfolio is a personal, single-machine tool; nothing is designed to be exposed to a network.

## Tech stack

| Layer | Technology |
| --- | --- |
| Language | Go 1.21+ |
| CLI | Cobra |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| Full-text search | SQLite FTS5 |
| MCP | `mark3labs/mcp-go` |
| Logging | zap + lumberjack |
| Config | TOML (`BurntSushi/toml`) |
| Dashboard | React 19, TypeScript, Vite, Tailwind CSS 4, Chart.js, Cytoscape |
| Release | GoReleaser, CGO_ENABLED=0, darwin/linux · amd64/arm64 |
