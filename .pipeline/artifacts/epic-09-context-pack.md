# Context Pack: Epic 9 — Claude Code Integration

## Project Essentials

**Vision:** Local-first project inventory & knowledge platform. Developers + AI agents understand an entire software portfolio.

**Lifecycle:** Install → Initialize → Forget (minimal ongoing maintenance)

**Key Principle:** *Engine Knows, Agent Thinks* — Engine is deterministic. AI agents do semantic reasoning. Never move semantic reasoning into the engine.

**User Journey:** AI coding agent is primary interface. Dashboard is read-only exploration. CLI is for admin only.

**AI is Optional** — Value after deterministic discovery; AI enriches but isn't required.

---

## Architecture (relevant to Epic 9)

```
AI Coding Agent
       |
Agent Integration  ← Epic 9
       |
MCP / HTTP APIs
       |
Portfolio Engine (Go)
       |
SQLite Knowledge Store
```

**Agent Integration** connects AI agents to the Engine:
- MCP registration
- Agent instructions/skills
- Installation, validation, upgrades

**Engine exposes deterministic capabilities only** — never knows about a specific AI agent.

**Agent Agnostic** — Engine never depends on a specific AI assistant. Agent-specific behavior belongs in installable integrations (see ADR-013).

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Engine | Go |
| Knowledge Store | SQLite |
| Agent Protocol | MCP (Model Context Protocol) |
| Dashboard API | HTTP/REST |
| Git | Discovery & metadata |

---

## Canonical Knowledge Model

### Engine (deterministic) stores:
- filesystem facts, git facts, documentation, metadata, technologies, timestamps

### Agent (semantic) stores:
- summaries, purpose, features, architecture, recommendations, relationships

### Store facts, compute indicators:
- Store: git_head, last_scan, last_analysis, documentation_hash
- Compute: analysis_available, needs_analysis, analysis_outdated, documentation_changed

---

## MCP Tools (what the integration exposes)

**Discovery:** `health()`, `discoverProjects()`, `listProjects()`, `getProject(id)`
**Search:** `searchProjects(query)`, `searchDocumentation(query)`
**Analysis:** `getAnalysis(projectId)`, `storeAnalysis(projectId, analysis)`, `listProjectsNeedingAnalysis()`
**Config:** `getConfiguration()`, `updateConfiguration()`
**Relationships:** `listRelationships(projectId)`

**Rules:** Small composable tools, stateless where possible, deterministic outputs.

---

## Agent Responsibilities (per PlatformSpec)

1. Discover projects
2. Detect changes
3. Decide when analysis is needed
4. Investigate repositories
5. Produce semantic knowledge
6. Persist analyses

**Workflow:** `health()` → `discoverProjects()` → search metadata → if analysis missing/outdated: investigate + produce + `storeAnalysis()` → answer user.

**Constraints:** Never edit repositories. Prefer existing knowledge before re-analysis.

---

## Integration Architecture (ADR-013)

Agent-specific behavior = **installable integrations**, not embedded in engine.

An integration is responsible for:
- Registering MCP server
- Installing agent-specific skills/instructions
- Validating installation
- Upgrading integration
- Removing integration

---

## Epic 9 Stories

| Story | Size | Blocked By | Description |
|-------|------|-----------|-------------|
| 9.1 Install MCP | M | 8.2 | `portfolio install claude` — registers MCP server, configures connection, verifies tools available |
| 9.2 Install Portfolio Skill | M | 9.1 | Installs Portfolio skill/prompt for Claude Code — describes MCP tools + query examples, upgradable |
| 9.3 Verify Integration | S | 9.2 | `portfolio doctor` checks Claude integration — verifies MCP connection, tool availability, reports issues with remediation |
| 9.4 Update Integration | M | 8.4 | `portfolio upgrade claude` — updates MCP config + skill, preserves config, reports changes |
| 9.5 Uninstall Integration | S | 9.4 | `portfolio uninstall claude` — removes MCP registration, removes skill file, preserves project data, idempotent |

**Total:** 3M + 2S (~12 days)
**Can start:** Story 9.1 (after Epic 8.2 complete)

---

## Engineering Guidelines (for implementation)

- **Deterministic by default** — repeatable results for same input, avoid LLM heuristics
- **Capabilities over workflows** — small composable tools, AI agents compose workflows
- **Store facts, compute indicators** — never store derived state unless necessary
- **Single Knowledge Model** — every interface (DB, MCP, HTTP, dashboard, agents) uses the same canonical model
- **Minimize dependencies** — Go engine, minimal deps
- **Composition over inheritance**, cohesive packages, no global state, interface-first design
- **Tests for deterministic logic only** — semantic understanding by AI agents is not tested as deterministic

**When model changes:** 1. Update KnowledgeModel.md → 2. Update PlatformSpecification.md → 3. Update implementation
