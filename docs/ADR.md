
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

---

## ADR-016: Official Methods Only for Agent Integrations

**Status:** Accepted

### Context

Portfolio integrations need to register MCP servers with various AI coding agents.
Different agents have different approaches to MCP server registration:

- **Claude Code**: Provides official CLI commands (`claude mcp add/remove/get`)
- **OpenCode**: Partial support (remote servers only via CLI, local requires config editing)
- **Cline**: No CLI support (requires manual config editing)

Initial implementation attempted direct config file manipulation, which created several problems:

1. **Fragility**: Config formats change between tool versions
2. **Maintenance burden**: Breaking changes require immediate fixes
3. **User trust**: Direct file editing feels unsafe
4. **No official support**: Tools don't guarantee config format stability

### Decision

All Portfolio integrations MUST use official tool methods for MCP server registration.

**Requirements:**

1. **Use official CLI**: When available, integrations must use official CLI commands
2. **No direct config editing**: Production code must never directly edit agent config files
3. **Transparent fallback**: For tools without official methods, provide unsafe scripts with warnings
4. **Documentation first**: Manual setup documentation takes precedence over automation

**Implementation Approach:**

| Tool Status | Integration Approach |
|-------------|---------------------|
| Official CLI exists | ✅ Create automated integration using official commands |
| Partial official support | ⚠️ Document limitations, provide unsafe scripts with warnings |
| No official support | ❌ Provide manual setup docs and unsafe scripts only |

**For Tools Without Official Methods:**

- Document manual setup in `docs/integration-guideline.md`
- Create `scripts/unsafe-<tool>-integration.sh` with:
  - Clear warnings that script is unofficial and unsafe
  - User consent requirements
  - Automatic backups before changes
  - Fully visible and reviewable code
- Never embed config manipulation in production code

### Consequences

**Positive:**

- Integrations remain stable across tool updates
- Users trust the integration process
- Reduced maintenance burden
- Clear user expectations

**Negative:**

- Some tools require manual setup instead of full automation
- Users must understand risks for unsafe scripts
- Cannot provide "perfect" automation for every tool

**Implementation Notes:**

- Claude Code integration uses `claude mcp add/remove/get` commands
- OpenCode has `opencode mcp add` but only for remote servers
- Cline requires manual `~/.cline/mcp.json` editing
- All unsafe scripts live in `scripts/` with clear README documentation
