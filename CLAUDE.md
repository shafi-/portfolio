# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## Project Status

This is a **specification repository** for "Portfolio" — a local-first project inventory and knowledge platform. No source code exists yet. The repository contains planning documents that define the product vision, architecture, and implementation roadmap.

---

## Authoritative Documents

The following documents in `docs/` define the project specifications:

| Document | Purpose |
|----------|---------|
| **KnowledgeModel.md** | Canonical domain model — defines core entities (Project, Documentation, Analysis, Features, Technologies, Relationships) and the separation between deterministic (Engine) and semantic (AI Agent) knowledge. |
| **PlatformSpecification.md** | Implementation contracts — database schema, MCP tools, HTTP API, agent workflows, dashboard specification, and implementation order. |
| **PRD.md** | Product Requirements Document — vision, goals, non-goals, and success criteria. |
| **Guideline.md** | Engineering principles — "Engine Knows, Agent Thinks," deterministic by default, local-first, capabilities over workflows. |
| **tasks/index.md** | Implementation roadmap index — milestones, epics, progress tracking. Detailed stories are in `tasks/epic-*.md` files. |
| **ADR.md** | Architecture Decision Records — incremental decisions like agent integrations being first-class components. |

---

## Architecture Summary

Portfolio is designed as **local-first infrastructure** that enables AI coding agents to understand a developer's entire software portfolio.

**Core Components:**
- **Portfolio Engine** (Go) — deterministic operations: discovery, metadata extraction, documentation indexing, storage
- **MCP Interface** — composable tools for AI agents
- **HTTP API** — for the dashboard
- **Local Knowledge Store** (SQLite)
- **Dashboard** — read-only visualization
- **Agent Integrations** — first-class components (Claude Code, Codex CLI, etc.)

**Key Principle:** The Engine performs deterministic work only; AI coding agents perform semantic reasoning (summaries, architecture understanding, feature extraction).

---

## Engineering Principles

From `Guideline.md`:

1. **Engine Knows, Agent Thinks** — Never move semantic reasoning into the engine.
2. **Deterministic by Default** — Engine operations must be repeatable for the same input.
3. **Store Facts, Compute Indicators** — Persist immutable facts (git HEAD, documentation hash); compute indicators (needs analysis, outdated) when needed.
4. **Local First** — Repositories and knowledge remain on the user's machine.
5. **Capabilities over Workflows** — Expose small, composable capabilities (discoverProjects, listProjects, searchProjects); AI agents compose workflows.
6. **AI is Optional** — Portfolio provides value after deterministic discovery; AI enriches but isn't required.
7. **Dashboard is Read-only** — Visualizes knowledge only; never invokes AI or modifies repositories.
8. **Agent Agnostic** — Engine never depends on a specific AI assistant.
9. **Single Knowledge Model** — Every interface operates on the same canonical model.

---

## User Journey

```text
Install → Initialize → Forget
```

After initialization:
- **AI coding agent** is the primary interface
- **Dashboard** is the primary exploration interface
- **CLI** is reserved for administration (initialization, diagnostics, upgrades, integration management)

---

## Implementation Roadmap

1. **Milestone 1 — Core Engine**: Discovery, metadata extraction, documentation indexing, knowledge store, HTTP API, MCP server
2. **Milestone 2 — Agent Integration**: Integration framework, Claude Code integration, AI analysis
3. **Milestone 3 — Dashboard**: Backend and frontend (read-only)
4. **Milestone 4 — Portfolio Intelligence**: Relationships, insights

---

## When Model Changes

When the domain model changes:
1. Update `KnowledgeModel.md`
2. Update `PlatformSpecification.md`
3. Update implementation

Every significant architectural decision should result in an ADR entry.

---

## Technology Stack

The specifications indicate Go for the Portfolio Engine, SQLite for the local knowledge store, and MCP for AI agent integration. The dashboard frontend is not yet specified.

---

## Development Workflow

This repository supports the devflow development pipeline for systematic feature implementation.

### Available Devflow Agents

**devflow:documentation-readiness**
- Checks repository documentation completeness and quality
- Validates PRD, Architecture, Tech Stack, UI/UX Guidelines, and Backlogs
- Use before beginning feature work to ensure documentation is adequate
- Current status: **Ready for Feature Development** (all required docs present)

**devflow:devflow**
- Use for implementing any feature, epic, or story end-to-end
- Sequences full requirements → merge pipeline across specialized subagents
- Trigger with: `/devflow ...` or natural language like "Implement feature X"
- Use `--resume` to continue an interrupted pipeline

### Development Protocol

When implementing features:
1. Run `devflow:documentation-readiness` if documentation may have changed
2. Use `devflow:devflow` to execute the full development pipeline
3. The pipeline handles: requirements → architecture → implementation → review → commit

---

## Notes for Future Development

- This is planned as a greenfield project — no legacy code to consider
- The project follows a "specification-first" approach
- All implementation should derive from the authoritative documents above
- Tests are required for deterministic logic; semantic understanding by AI agents is not tested as deterministic
