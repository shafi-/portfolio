---
title: MCP Server
description: Portfolio's MCP stdio server — what it is, how it runs, and why agents love it.
---

# MCP Server

Portfolio exposes its knowledge base to AI coding agents through the [Model Context Protocol](https://modelcontextprotocol.io) (MCP). The binary itself is the server: run `portfolio mcp` and it speaks JSON-RPC over stdio, offering **44 tools** for discovering, searching, and analyzing your software portfolio.

This is the primary way to use Portfolio day to day — the CLI is for administration; your agent is the interface.

## How it runs

```bash
portfolio mcp
```

You rarely launch the server by hand. [Connecting an agent](/docs/mcp/agents) registers the binary as an MCP server in the agent's own configuration (as `portfolio mcp`), and the agent launches and manages the process when it starts.

Design details worth knowing:

- **stdio transport** — all logging is forced to stderr in this mode so the JSON-RPC channel stays clean.
- **Same binary, same store** — the MCP server reads the same [knowledge store](/docs/concepts/knowledge-store) as the CLI, API, and dashboard. Scan once, and every surface sees it.
- **Local-only** — the server is a child process of your agent on your machine. No network listeners are opened.

## What agents can do with it

The tool catalog is organized in seven areas:

| Area | Tools | Purpose |
| --- | --- | --- |
| Discovery & search | `discoverProjects`, `listProjects`, `getProject`, `searchProjects`, `searchDocumentation`, `health` | Find projects and their docs |
| Project content | `listProjectFiles`, `getProjectStructure`, `getFileContent`, `searchFiles`, `getDependencies` | Read code and structure across the portfolio |
| Analysis | `getAnalysis`, `storeAnalysis`, `listProjectsNeedingAnalysis`, `getProjectAnalyzerPrompt` | Store and reuse per-project analysis |
| Features | `storeFeature`, `listFeatures`, `searchFeatures` | Build a feature inventory |
| Technologies | `storeTechnology`, `tagProjectWithTechnology`, `listTechnologies`, `listProjectTechnologies`, `searchByTechnology` | Track what's used where |
| Cross-project | `searchAcrossProjects`, `getPortfolioOverview`, `compareProjects`, `findProjectDependencies`, `analyzeTechnologySpread` | Portfolio-level reasoning |
| Workspaces | `listWorkspaces`, `getWorkspace` | Group services into named products and check analysis freshness per service |
| History & quality | `getGitHistory`, `getCommitDiff`, `getFileEvolution`, `getProjectTimeline`, `analyzeCommitPatterns`, `getCodeComplexity`, `getCodeMetrics`, `getTestCoverage`, `getTechnicalDebt`, `analyzeCodeSmells` | Git and code-quality signals |

The full reference with parameters lives in [Tools Reference](/docs/mcp/tools).

## Connecting an agent

One command per agent:

```bash
portfolio install claude     # Claude Code
portfolio install opencode   # OpenCode
```

Then ask your agent something portfolio-wide, like *"which of my projects use Redis?"* — it now has the tools to answer from ground truth instead of guessing.

Continue to [Connecting AI Agents](/docs/mcp/agents).
