---
title: Workspaces
description: Group related projects — such as the microservices of one product — and keep track of which analyses are current.
---

# Workspaces

A **workspace** is a named group of projects — for example the microservices that together form one product:

```bash
portfolio workspace create commerce
portfolio workspace add commerce orders-service
portfolio workspace add commerce payments-service
portfolio workspace show commerce
```

Workspaces are **grouping-only** ([ADR-023](https://github.com/shafi-/portfolio/blob/main/docs/ADR.md)): they scope queries and reporting, but never change scan, discovery, or any other engine behavior. Deleting a workspace removes only the grouping — the projects themselves are untouched.

## Analysis freshness

Stored analyses describe a repository at a point in time (the git HEAD they were made against). Portfolio computes **freshness** whenever an analysis is handed out — it compares the analysis's HEAD against the repository's **live HEAD**, so staleness is visible even for commits made after the last scan:

| Status | Meaning |
| --- | --- |
| `fresh` | The analysis HEAD matches the repository's current HEAD |
| `stale` | The repository has moved past the analysis; `commits_behind` says by how much |
| `unknown` | Freshness cannot be determined (no analysis, no HEAD anchor, or the repository is unavailable) — never treated as fresh |

Freshness is **computed, never stored** — a stored freshness flag would go stale itself.

## The staleness gate

The design rule is: **warn before use, refresh only on request**.

- `getAnalysis` (MCP) and `projects get` (CLI) always report freshness alongside the analysis
- `getProjectAnalyzerPrompt` with a project ID embeds the freshness state, the previous analysis as an incremental seed, and an explicit instruction to the agent: *tell the user the analysis is stale and ask whether to refresh; if the user declines, proceed with the existing analysis and note its age*
- `listProjectsNeedingAnalysis` accepts an optional `workspace` scope and uses live-HEAD freshness

Nothing blocks and nothing auto-refreshes — the user stays the decision maker.

## Checking a whole service group

`workspace show` lists every member with its analysis freshness, so you can see at a glance which services' knowledge is current:

```bash
$ portfolio workspace show commerce
Workspace: commerce

PROJECT            ANALYSIS FRESHNESS     ID
orders-service     ⚠ stale (+14 commits)  5f0c…
payments-service   ✓ fresh                9a41…

Total: 2 project(s)
```

Agents get the same view through the [`getWorkspace`](/docs/mcp/tools#workspaces) MCP tool.

## Where to go next

- [CLI: Workspaces](/docs/cli/workspaces) — command reference
- [MCP Tools](/docs/mcp/tools) — agent-facing workspace and analysis tools
