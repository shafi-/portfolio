# Context Pack: Epic 8 — Integration Framework

## Project Identity
Portfolio: local-first project inventory and knowledge platform. Enables developers and AI coding agents to understand an entire software portfolio. Not a project management tool — a portfolio awareness system.

## Architecture (high-level)

```
AI Agent (primary interface)     Dashboard (read-only)
         \                            /
          \                          /
         Agent Integration (MCP/HTTP)
                    |
            Portfolio Engine
              /     |     \
        Discovery  Store   Search
              \     |     /
           Local Repositories
```

### Core Philosophy
**The Engine knows. The AI Agent thinks.**
- Engine: deterministic operations only (discovery, metadata, indexing, storage, search)
- AI Agent: semantic reasoning (summaries, features, architecture understanding, relationships)
- Dashboard: read-only visualization — never invokes AI, never modifies data
- CLI: administration only (init, config, upgrade, integration)

## ADR-013: Agent Integrations are First-Class Components
**Status:** Accepted — governs Epic 8 entirely.

**Decision:** Agent-specific behavior = installable integrations, not embedded in engine.
Integration responsibilities:
- Register MCP server
- Install agent-specific skills/instructions
- Validate installation
- Upgrade integration
- Remove integration

Engine is agent-agnostic — has no knowledge of specific AI agents.

**Consequence:** Each supported agent needs its own integration package.

## ADR-014: Install → Initialize → Forget
User lifecycle. After init, primary interaction through AI agent. CLI for bootstrap only.

## ADR-015: KnowledgeModel is Canonical Source of Truth
KnowledgeModel.md → concepts. PlatformSpecification.md → implementation contracts.

## Engineering Principles (Guideline.md) — key for Epic 8

1. **Engine Knows, Agent Thinks** — Never move semantic reasoning into engine.
2. **Deterministic by Default** — Engine operations repeatable for same input.
3. **Store Facts, Compute Indicators** — Persist immutable facts; compute derived state on demand.
4. **Local First** — Repos and knowledge stay on user machine.
5. **Capabilities over Workflows** — Expose small composable tools; AI agents compose workflows.
6. **AI is Optional** — Value after deterministic discovery; AI enriches.
7. **Agent Agnostic** — Engine never depends on specific AI assistant. Agent-specific behavior in installable integrations.
8. **Single Knowledge Model** — Every interface (DB, MCP, HTTP, Dashboard, Agents) uses same canonical model.
9. **CLI is Administrative** — init, upgrades, diagnostics, integration management.

## Tech Stack
- **Portfolio Engine:** Go (deterministic ops — filesystem, git, DB)
- **Local Knowledge Store:** SQLite
- **Agent Protocol:** MCP (Model Context Protocol)
- **HTTP API:** REST (for dashboard)
- **Dashboard:** asset serving + REST

## Knowledge Model (entities relevant to Epic 8)

**Project** — discovered software project. Identity (UUID, name, root_path, repository_type, discovered_at) + Metadata (git, languages, frameworks, deps, docs, hashes) + Analysis (agent-produced, optional).

**Analysis** — agent-only. Fields: summary, purpose, architecture, maturity, strengths, weaknesses, reusable_components, notes, analyzed_at, analyzed_git_head, analyzer.

**Feature** — capability in a project (belongs to analysis).

**Technology** — normalized tech reference. Used for filtering/relationships.

**Relationship** — connection between two projects (Similar, Evolution, Shared Feature, etc.). Agent-generated.

**Documentation** — engine-extracted (README, docs/*, ADRs, CHANGELOG). Stored searchable, no interpretation.

### Derived Indicators
Store: git_head, last_scan, last_analysis, documentation_hash
Compute: analysis_available, needs_analysis, analysis_outdated, documentation_changed

## PlatformSpecification — relevant contracts

### Database Schema (relevant tables for integration)
```
configuration: key/value store for integration metadata
```

### MCP Tools
**Discovery:** health(), discoverProjects(), listProjects(), getProject(id)
**Search:** searchProjects(query), searchDocumentation(query)
**Analysis:** getAnalysis(projectId), storeAnalysis(projectId, analysis), listProjectsNeedingAnalysis()
**Configuration:** getConfiguration(), updateConfiguration()
**Relationships:** listRelationships(projectId)

Rules: Small composable tools, stateless where possible, deterministic outputs.

### AI Agent Workflow
```
User → Portfolio question → health() → discoverProjects() → Search
  → If semantic knowledge missing/outdated: Investigate repo → Produce analysis → storeAnalysis()
  → Answer user
```
Never edit repositories. Prefer existing knowledge before re-analysis.

### Implementation Order
1. Database → 2. Discovery → 3. Metadata → 4. Doc indexing → 5. Search → 6. HTTP API → 7. MCP → **8. Agent integration** → 9. Dashboard → 10. Portfolio intelligence

## Epic 8: Integration Framework

**Milestone:** 2 — Agent Integration
**Blocked by:** Epic 7 (MCP Server)
**Can start:** After Epic 7 complete
**Total size:** ~12 days (3M + 2S)

### Story 8.1 — Integration Abstraction (M, blocked by Epic 7)
Interface: install, validate, upgrade, remove. Each integration stores its own metadata. Engine stays agent-agnostic. Integrations discoverable.

### Story 8.2 — Installation Framework (M, blocked by 8.1)
Registration command. Store integration metadata in DB. Validate requirements. List installed integrations.

### Story 8.3 — Validation (S, blocked by 8.2)
Validate command per integration. Check: MCP server reachable, tools available. Return validation result with diagnostics. Self-heal or report.

### Story 8.4 — Upgrade Mechanism (M, blocked by 8.3)
Integration version tracking. Upgrade command. Compatibility check with engine version. Rollback on failure.

### Story 8.5 — Removal/Uninstall (S, blocked by 8.4)
CLI remove command. Cleans integration metadata and artifacts. Unregisters MCP config. Idempotent.

## GUARDRAILS for implementation

1. **Engine must not reference any specific AI agent** (no Claude Code, Codex, etc. in engine code)
2. **All agent-specific behavior** goes in the integration layer
3. **CLI commands are the interface** for integration management (not Dashboard, not MCP)
4. **Integration metadata** uses the `configuration` table or a dedicated schema — keep it simple
5. **Follow Go conventions** — interfaces around capabilities, composition, minimal deps
6. **Test deterministic logic** — integration interface behavior must be testable without real agents
7. **Store facts, compute indicators** — store integration metadata; compute "needs upgrade" etc.
