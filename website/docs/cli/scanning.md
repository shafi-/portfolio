---
title: Scanning & Status
description: init, discover, scan, and status — the commands that build and inspect the knowledge base.
---

# Scanning & Status

These four commands drive the write path of the engine: setup, discovery, extraction, and inspection.

## `portfolio init`

Initialize the Portfolio Engine with interactive setup.

```bash
portfolio init
```

The wizard:

1. Asks which directories to use as [project roots](/docs/cli/configuration#portfolio-config-set-root)
2. Creates the [config file](/docs/getting-started/configuration) at the platform location
3. Connects to and initializes the [knowledge store](/docs/concepts/knowledge-store), applying all embedded migrations

Re-running `init` is safe — existing configuration is preserved.

## `portfolio discover`

Run project discovery on all configured roots.

```bash
portfolio discover
```

Finds every Git repository under your roots and registers it in the knowledge store. Discovery is incremental: known projects are updated, new ones are added. See [Discovery](/docs/concepts/discovery).

## `portfolio scan`

Extract metadata and index documentation for projects — the heavy lifter.

```bash
portfolio scan                 # all discovered projects
portfolio scan --project <id>  # a single project
```

| Flag | Description |
| --- | --- |
| `--project <id>` | Scan only the project with this ID |

For each project, `scan`:

- Extracts [git intelligence](/docs/concepts/metadata#git-intelligence) (HEAD, branch, commits, contributors, remotes)
- Detects [languages, frameworks, and dependencies](/docs/concepts/metadata#languages-frameworks-and-dependencies)
- Records [capability and health signals](/docs/concepts/metadata#capability-signals)
- [Indexes documentation](/docs/concepts/knowledge-store#full-text-search) into full-text search with hash-based change tracking

::: tip When to re-scan
Run `portfolio scan` after cloning new repos, after meaningful development sessions, or on a schedule (cron, CI) if you use Portfolio across many machines. Re-scans are cheap thanks to incremental indexing.
:::

## `portfolio status`

Show Portfolio Engine system status.

```bash
portfolio status
```

Reports database connectivity, discovered project count, and configuration state — a quick smoke test that the engine is healthy. For integration-specific diagnostics (Claude Code, OpenCode), use [`portfolio doctor`](/docs/cli/ai-integrations#doctor).
