---
title: Quick Start
description: Go from zero to a fully indexed portfolio in five minutes.
---

# Quick Start

This guide takes you from a fresh install to a fully indexed, AI-ready portfolio in about five minutes.

## 1. Initialize

```bash
portfolio init
```

The interactive setup walks you through choosing which directories to scan (your project roots) and writes your [configuration file](/docs/getting-started/configuration). You can add more roots any time with `portfolio config set-root`.

## 2. Discover projects

```bash
portfolio discover
```

Portfolio walks your configured project roots, finds every Git repository (skipping `node_modules`, `.git`, `vendor`, and other common noise), and records each one in the knowledge base.

## 3. Scan for metadata and docs

```bash
portfolio scan
```

This is the workhorse: for every discovered project it extracts git intelligence (branch, commit counts, contributors, remotes), detects languages, frameworks, and dependencies, and indexes documentation — READMEs, ADRs, architecture docs, OpenAPI specs, and changelogs — into the full-text search store.

To re-scan a single project after big changes:

```bash
portfolio scan --project <project-id>
```

## 4. Check status

```bash
portfolio status
```

Shows the health of the engine: database connectivity, discovered project count, and configuration state. [`portfolio doctor`](/docs/cli/ai-integrations#doctor) runs deeper diagnostics, including per-integration checks.

## 5. Explore

### From the terminal

```bash
portfolio projects list
portfolio projects search <query>
portfolio projects get <project-id>
```

### In the browser

```bash
portfolio dashboard
# Dashboard listening — open http://localhost:3000
```

The [dashboard](/docs/dashboard) gives you a project list, per-project detail pages, portfolio-wide statistics, and a relationship explorer.

### Over HTTP

```bash
portfolio serve
# API server listening on http://127.0.0.1:8080
curl http://127.0.0.1:8080/health
```

See the [HTTP API reference](/docs/api/) for all endpoints.

## 6. Connect your AI agent

```bash
portfolio install claude   # or: portfolio install opencode
```

This registers Portfolio's [MCP server](/docs/mcp/) with your agent so it can search your portfolio, read project docs, and analyze cross-project relationships as tools.

::: tip That's the whole loop
Day-to-day you only need `portfolio scan` occasionally to refresh the knowledge base — everything else reads from it.
:::

## Where your data lives

All state is kept in an obscured SQLite database under your platform's application data directory (for example `~/Library/Application Support/com.portfolio.cli/` on macOS). Nothing leaves your machine. See [Database & Security](/docs/digging-deeper/security).
