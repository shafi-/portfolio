---
title: Dashboard
description: Portfolio's embedded, read-only web dashboard — pages, features, and how to run it.
---

# Dashboard

Portfolio includes a read-only web dashboard compiled directly into the binary. Start it with:

```bash
portfolio dashboard
# open http://localhost:3000
```

| Flag | Description |
| --- | --- |
| `-p`, `--port <port>` | Port (default `3000`) |
| `--dev` | Serve assets from the filesystem (dashboard development) |
| `--dist <path>` | Dist folder for dev mode (default `./dashboard/dist`) |

Because the UI is embedded, `go build` produces a fully self-contained binary — no static files to deploy alongside it.

## Pages

The dashboard is deliberately read-only: it browses the knowledge store and never mutates it.

### Project List

All discovered projects in a filterable list — name, path, languages, frameworks, and health signals at a glance.

### Project Detail

Everything known about one project: git intelligence (branch, commits, contributors, remotes), language and framework summaries, capability signals, indexed documents, and analyses.

### Overview

The portfolio landing view — high-level counts and health indicators across all projects.

### Statistics

Portfolio-wide aggregates rendered with charts: language, framework, technology, and document-kind distributions from the [statistics endpoint](/docs/api/endpoints#get-statistics).

### Relationship Explorer

An interactive graph of [project relationships](/docs/api/endpoints#relationships) — shared technology, reused components, evolution — built on Cytoscape.

## Under the hood

- **React 19 + TypeScript + Vite + Tailwind CSS 4**, with Chart.js and Cytoscape for visualizations
- Talks to the [HTTP API](/docs/api/) on the same origin (the dashboard server proxies the API)
- Production build is embedded via `go:embed`; `--dev` mode with Vite hot reload is available for working on the UI itself:

```bash
# terminal 1 — build/watch the dashboard frontend
npm --prefix dashboard run dev

# terminal 2 — serve it from the filesystem
portfolio dashboard --dev
```

See [Building the dashboard from source](/docs/getting-started/installation#building-the-dashboard-from-source) for the full toolchain.
