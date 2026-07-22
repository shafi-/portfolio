
# ADR.md

# Architecture Decision Records

This document contains incremental architecture decisions for Portfolio.

---

## ADR-013: Agent Integrations are First-Class Components

**Status:** Accepted

### Context

The Portfolio Engine is designed to be AI-agent agnostic. While Claude Code is the
first supported agent, the architecture should support additional coding agents
without changing the engine.

Examples include:

- Claude Code
- Codex CLI
- OpenCode
- Cursor
- Future MCP-compatible agents

### Decision

Agent-specific behavior shall be implemented as installable integrations rather
than embedded into the engine.

An integration is responsible for:

- Registering the MCP server
- Installing agent-specific skills or instructions
- Validating the installation
- Upgrading the integration
- Removing the integration

The engine exposes deterministic capabilities only and has no knowledge of a
specific AI agent.

### Consequences

Positive:

- Engine remains agent-agnostic.
- New AI agents can be supported independently.
- Agent-specific prompting and workflows evolve separately from the engine.

Negative:

- Each supported agent requires its own integration package.

---

## ADR-014: Install → Initialize → Forget

**Status:** Accepted

### Context

Developer tools often require users to remember commands and perform ongoing
maintenance.

Portfolio aims to be invisible infrastructure rather than another tool that
demands attention.

### Decision

Portfolio follows a simple lifecycle:

```text
Install
    ↓
Initialize
    ↓
Forget
```

The CLI exists primarily for:

- Initialization
- Diagnostics
- Upgrades
- Integration management

After initialization, the primary interaction is through an AI coding agent.

### Consequences

Positive:

- Minimal user friction.
- Natural AI-first workflow.
- Reduced operational complexity for users.

Negative:

- Greater emphasis on robust agent integrations.

---

## ADR-015: KnowledgeModel is the Canonical Source of Truth

**Status:** Accepted

### Context

As the platform grows, multiple interfaces (database, MCP, HTTP API, dashboard,
and AI agents) require a consistent understanding of Portfolio concepts.

### Decision

`KnowledgeModel.md` is the canonical definition of the domain model.

`PlatformSpecification.md` defines how that model is implemented.

All implementations must derive from these documents rather than redefining
entities independently.

### Consequences

Positive:

- Consistent domain language.
- Easier evolution of the platform.
- Reduced duplication across documentation and code.

Negative:

- Changes to the domain model require synchronized updates to downstream
specifications.
