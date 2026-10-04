---
title: Tools Reference
description: All 42 MCP tools exposed by Portfolio, grouped by capability.
---

# Tools Reference

Portfolio's MCP server registers **44 tools**. They fall into seven families, mirroring the server's source layout. All tools read from (and, where noted, write to) the shared [knowledge store](/docs/concepts/knowledge-store).

::: tip For agents
Tool call conventions below are indicative — your MCP client receives exact JSON Schemas for each tool. Humans can browse the registration code in `internal/mcp/`.
:::

## Discovery & projects

The core loop: find projects, inspect them, check engine health.

| Tool | Description |
| --- | --- |
| `health` | Engine health: database connectivity, project count |
| `discoverProjects` | Run discovery over configured roots |
| `listProjects` | List all discovered projects |
| `getProject` | Full detail for one project |
| `searchProjects` | Search projects by name |

## Documentation search

| Tool | Description |
| --- | --- |
| `searchDocumentation` | Full-text search across indexed docs (READMEs, ADRs, architecture docs, OpenAPI specs, changelogs) with snippets |

## Configuration

| Tool | Description |
| --- | --- |
| `getConfiguration` | Read the active engine configuration |
| `updateConfiguration` | Update configuration values |

## Analysis

Store and reuse per-project analysis so agents (and teammates) don't re-derive it.

Every analysis-bearing response carries a `freshness` block — `fresh`, `stale` (with `commits_behind`), or `unknown` — computed against the repository's live HEAD, so staleness is visible **before** an analysis is relied on. See [Workspaces](/docs/concepts/workspaces) for the gate flow: warn before use, refresh only if the user asks.

| Tool | Description |
| --- | --- |
| `getAnalysis` | Fetch stored analyses for a project, with a freshness report |
| `storeAnalysis` | Persist an analysis result for a project (anchored to the current HEAD) |
| `listProjectsNeedingAnalysis` | Find projects with no or stale analysis; optional `workspace` scope |
| `getProjectAnalyzerPrompt` | Prepared analyzer prompt; with `project_id` it includes freshness, the previous analysis as an incremental seed, and the ask-the-user-before-refreshing instruction |

## Relationships

Typed, confidence-scored edges between projects — the raw material for architecture maps.

| Tool | Description |
| --- | --- |
| `listRelationships` | List relationships (portfolio-wide or per project) |
| `storeRelationship` | Store an edge (`Similar`, `Shared Technology`, `Reuses Component`, …) with a confidence score |

See the [HTTP relationships endpoint](/docs/api/endpoints#relationships) for the wire format.

## Project content

Read code across the whole portfolio without leaving the agent.

| Tool | Description |
| --- | --- |
| `listProjectFiles` | List files in a project |
| `getProjectStructure` | Directory tree of a project |
| `getFileContent` | Read a file's content |
| `searchFiles` | Search file contents within a project |
| `getDependencies` | List a project's dependencies |

## Features

| Tool | Description |
| --- | --- |
| `storeFeature` | Record a discovered feature for a project |
| `listFeatures` | List features of a project |
| `searchFeatures` | Search features across the portfolio |

## Technologies

| Tool | Description |
| --- | --- |
| `storeTechnology` | Register a technology in the portfolio inventory |
| `tagProjectWithTechnology` | Tag a project as using a technology |
| `listTechnologies` | List all known technologies |
| `listProjectTechnologies` | Technologies used by one project |
| `searchByTechnology` | Find projects using a technology — *"which projects use Postgres?"* |

## Cross-project intelligence

Portfolio-level reasoning — the tools that make the whole more than the sum of repos.

| Tool | Description |
| --- | --- |
| `searchAcrossProjects` | Search content across every indexed project |
| `getPortfolioOverview` | High-level portfolio snapshot (counts, health) |
| `compareProjects` | Compare two or more projects side by side |
| `findProjectDependencies` | Map dependencies between projects |
| `analyzeTechnologySpread` | Analyze how technologies are distributed across the portfolio |

## History & code quality

Git-derived and static signals per project.

| Tool | Description |
| --- | --- |
| `getGitHistory` | Commit history for a project |
| `getCommitDiff` | Diff for a specific commit |
| `getFileEvolution` | Change history of a single file |
| `getProjectTimeline` | Project activity over time |
| `analyzeCommitPatterns` | Patterns in commit activity (velocity, cadence) |
| `getCodeComplexity` | Complexity signals for a project |
| `getCodeMetrics` | General code metrics |
| `getTestCoverage` | Test coverage signals |
| `getTechnicalDebt` | Technical-debt indicators |
| `analyzeCodeSmells` | Detect code smells |

## Workspaces

Read access to [workspaces](/docs/concepts/workspaces) — named groups of projects. Management is CLI-side (`portfolio workspace …`); agents get enough to check which services' analyses are current.

| Tool | Description |
| --- | --- |
| `listWorkspaces` | List workspaces with member counts |
| `getWorkspace` | One workspace's members, each with analysis freshness (fresh/stale/none) |

## Idiomatic usage

The order that works well for agents:

1. `health` → confirm the engine is up
2. `listProjects` / `searchProjects` → frame the question
3. `getProject` + `searchDocumentation` → ground the answer in facts
4. `searchAcrossProjects` / `compareProjects` / `analyzeTechnologySpread` → portfolio-level insight
5. `storeAnalysis` / `storeFeature` / `tagProjectWithTechnology` → persist new findings for next time

The [`portfolio manual`](/docs/cli/servers#portfolio-manual) captures these patterns for inclusion in an agent's instructions file.
