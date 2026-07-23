# Epic 8 — Integration Framework: Requirements

**Milestone:** 2 — Agent Integration
**Status:** Draft
**Last Updated:** 2026-07-22

---

## 1. Feature Overview

Epic 8 implements a first-class integration framework that allows AI coding agents to register, validate, upgrade, and remove themselves as consumers of the Portfolio Engine. Per ADR-013, the engine remains completely agent-agnostic — all agent-specific behavior lives in installable integrations.

The framework provides a common abstraction (`Integration`) with a lifecycle of install → validate → upgrade → remove. Integrations are stored, discovered, and managed via CLI commands. The engine exposes deterministic capabilities via MCP; the integration layer bridges those capabilities to specific AI agents by registering the MCP server and installing agent-specific instructions.

### Actor Map

```
User (CLI) ──> Portfolio Engine ──> Integration Framework
                                      ├── claude-code integration
                                      ├── codex integration
                                      ├── opencode integration
                                      └── (future agents)
                                              │
                                              ▼
                                      MCP Server (Epic 7)
```

---

## 2. Requirements

### 2.1 Functional Requirements

#### FR-1: Integration Interface
- The engine MUST define an `Integration` interface (Go) with methods: `Install`, `Validate`, `Upgrade`, `Remove`.
- Each integration MUST own and manage its own metadata (name, version, agent type, install path, configuration).
- Integrations MUST be discoverable — the engine MUST be able to enumerate all registered integrations.
- The engine MUST NOT import or reference any specific AI agent package.

#### FR-2: Integration Registration
- A CLI command (`portfolio integration install <name>`) MUST register an integration.
- Registration MUST store integration metadata in the database (via `configuration` table or dedicated integration schema).
- Registration MUST validate that the integration's requirements are met (e.g., agent binary exists, MCP server reachable).
- A CLI command (`portfolio integration list`) MUST enumerate all registered integrations with their status (installed, validated, version, agent type).

#### FR-3: Integration Validation
- A CLI command (`portfolio integration validate <name>`) MUST run validation for a specific integration.
- Validation MUST check: MCP server is reachable, required MCP tools are available, agent binary/path exists.
- Validation MUST return a structured result with diagnostics (pass/fail per check, error messages, suggested remediation).
- Validation MUST NOT modify persistent state — it is read-only.
- Self-healing for recoverable issues (restart MCP, recreate config, recreate dir) is applied by the `doctor --fix` command at the Manager level, not by the integration's Validate() method.
- Each check may include a `self_healable` flag indicating whether doctor --fix can repair it.

#### FR-4: Integration Upgrade
- Each integration MUST track its own version (separate from engine version).
- A CLI command (`portfolio integration upgrade <name>`) MUST upgrade a specific integration.
- Upgrade MUST check compatibility between the integration version and the engine version before proceeding.
- On upgrade failure, the system MUST roll back to the previous integration version.
- Rollback MUST restore the previous integration metadata and artifacts.

#### FR-5: Integration Removal
- A CLI command (`portfolio integration remove <name>`) MUST unregister an integration.
- Removal MUST clean up integration artifacts (MCP server registration, agent-specific files).
- Removal MUST NOT affect other integrations or the engine's core data.

### 2.2 Non-Functional Requirements

| ID | Requirement | Rationale |
|----|-------------|-----------|
| NFR-1 | **Agent Agnostic** — Engine must never depend on a specific AI agent. | Core architectural constraint (Guideline.md, ADR-013). |
| NFR-2 | **Deterministic** — All integration operations must be repeatable for the same input. | Consistent with platform design principle. |
| NFR-3 | **CLI Only** — Integration management is CLI-only; never via MCP or Dashboard. | Per Guideline.md: "CLI is Administrative." |
| NFR-4 | **Testable Without Real Agents** — Integration interface behavior must be testable using mocks/fakes. | Guards against coupling to specific agent implementations. |
| NFR-5 | **Isolation** — Integration failures must not crash the engine or affect other integrations. | Each integration runs in its own scope. |
| NFR-6 | **Configurable Install Path** — Integration install location must be user-configurable (defaulting to a standard location). | Supports different environments and CI setups. |
| NFR-7 | **Idempotent Operations** — `install`, `validate`, `upgrade`, `remove` must be safe to run multiple times. | User may retry after failures. |

---

## 3. Edge Cases & Error Handling

| # | Edge Case | Handling |
|---|-----------|----------|
| EC-1 | Integration name collision on install | Reject with clear error: "Integration '<name>' is already installed. Use `upgrade` or `remove` first." |
| EC-2 | MCP server not running during validation | Report as diagnostic failure with guidance: "MCP server not reachable at <address>. Start with `portfolio mcp start`." |
| EC-3 | Agent binary missing during validation | Report diagnostic failure: "Required agent binary '<path>' not found. Install <agent> first." |
| EC-4 | Upgrade compatibility mismatch | Block upgrade with message: "Integration vX.Y requires engine vA.B.C (current: vD.E.F). Upgrade engine first." |
| EC-5 | Upgrade failure mid-operation (partial state) | Roll back to previous version. Restore integration metadata from backup taken at upgrade start. |
| EC-6 | Remove integration that is in use (analysis in flight) | Warn user: "Integration '<name>' is active. Analysis in progress may fail. Continue? [y/N]" |
| EC-7 | Remove integration that was never installed | Report: "Integration '<name>' is not installed." — no-op. |
| EC-8 | Validate integration that was never installed | Report: "Integration '<name>' is not installed. Run `portfolio integration install <name>` first." |
| EC-9 | Upgrade integration that was never installed | Report: "Integration '<name>' is not installed. Cannot upgrade." |
| EC-10 | Database unavailable during registration/upgrade/removal | Fail with: "Knowledge store unavailable. Ensure Portfolio is initialized (`portfolio init`)." |
| EC-11 | Concurrent integration operations | No file lock needed — SQLite serializes metadata writes. Single-user CLI tool; concurrent install/upgrade/remove from two terminals simultaneously is not a real scenario. |
| EC-12 | Integration path contains spaces or special characters | Properly quote and escape paths in all shell/file operations. |
| EC-13 | Integration scripts or binaries require elevated permissions | Fail with clear message: "Permission denied at '<path>'. Run with appropriate permissions or configure install path." |
| EC-14 | Network timeout during validation checks (e.g., version lookup) | Fail gracefully: "Could not check latest version. Network unreachable. Proceeding with local validation only." |

---

## 4. Acceptance Criteria

### AC-8.1 — Integration Abstraction
- [ ] `Integration` interface exists in Go code with methods: `Install()`, `Validate()`, `Upgrade()`, `Remove()`.
- [ ] Each integration stores its own metadata (name, version, agent type, install path).
- [ ] Engine code contains zero references to specific AI agents (Claude Code, Codex, OpenCode, Cursor, etc.).
- [ ] Integrations are discoverable via `list` command — engine can enumerate all registered integrations.
- [ ] Interface is composable — integration manager accepts any type satisfying the interface.

### AC-8.2 — Installation Framework
- [ ] `portfolio integration install <name>` command exists and registers an integration.
- [ ] Integration metadata is persisted in the knowledge store (database).
- [ ] Installation validates requirements before writing state (agent binary exists, MCP server reachable).
- [ ] `portfolio integration list` shows all integrations with: name, version, status, installed_at.
- [ ] Duplicate install attempts are rejected with a clear error message.
- [ ] Install is idempotent when integration already exists at the correct version.

### AC-8.3 — Validation
- [ ] `portfolio integration validate <name>` command runs per-integration validation.
- [ ] Validation checks: MCP server is reachable, required MCP tools are available, agent binary exists.
- [ ] Validation returns structured output (structured JSON or formatted table) with pass/fail per check.
- [ ] Validation includes diagnostic messages and suggested remediation for failures.
- [ ] Validation does not modify any persistent state (read-only).
- [ ] Self-healing for recoverable issues is handled by `portfolio integration doctor --fix`, not by Validate() itself.

### AC-8.4 — Upgrade Mechanism
- [ ] Each integration tracks its own version independently of the engine version.
- [ ] `portfolio integration upgrade <name>` command upgrades a specific integration.
- [ ] Upgrade checks compatibility with engine version before proceeding.
- [ ] On failure, upgrade rolls back completely — previous integration version is restored.
- [ ] Rollback restores all prior integration metadata and artifacts.
- [ ] Upgrade is idempotent — running again after success reports "already up to date."

---

## 5. Data Flow

### 5.1 Integration Install Flow

```
User
  │
  ▼
CLI: portfolio integration install claude-code
  │
  ▼
IntegrationManager.Install("claude-code")
  │
  ├── Load integration package (binary/script at known path)
  │
  ├── Validate requirements
  │     ├── Agent binary exists? ──► fail if missing
  │     ├── MCP server reachable? ──► fail if unreachable
  │     └── No name collision?    ──► fail if duplicate
  │
  ├── Execute integration's Install()
  │     ├── Register MCP server tools
  │     ├── Install agent-specific skills/instructions
  │     └── Store agent config files
  │
  ├── Persist integration metadata
  │     ├── name, version, agent_type
  │     ├── install_path, installed_at
  │     └── status: "installed"
  │     │
  │     ▼
  │   Database (configuration table or integration schema)
  │
  └── Return success / failure to CLI
```

### 5.2 Integration Validation Flow

```
User
  │
  ▼
CLI: portfolio integration validate claude-code
  │
  ▼
IntegrationManager.Validate("claude-code")
  │
  ├── Load integration metadata from DB
  │
  ├── Run checks:
  │     ├── Check 1: MCP server health() call ──► pass/fail + diagnostic
  │     ├── Check 2: Required MCP tools exist ──► pass/fail + diagnostic
  │     └── Check 3: Agent binary exists ──► pass/fail + diagnostic
  │
  ├── Attempt self-heal for recoverable issues
  │     └── e.g., restart MCP server, recreate config
  │
  ├── Do NOT modify persistent state
  │
  └── Return ValidationResult{checks: [], passed: bool, diagnostics: []}
```

### 5.3 Integration Upgrade Flow

```
User
  │
  ▼
CLI: portfolio integration upgrade claude-code
  │
  ▼
IntegrationManager.Upgrade("claude-code")
  │
  ├── Load current integration metadata from DB
  │
  ├── Check engine compatibility
  │     └── integration version vs engine version ──► fail if incompatible
  │
  ├── Snapshot current state (backup metadata + artifacts)
  │
  ├── Execute integration's Upgrade()
  │     ├── Download/apply new integration version
  │     └── Update MCP registrations if needed
  │
  ├── Validate new state
  │     └── Run Validate() on upgraded integration
  │
  ├── On success:
  │     ├── Update metadata in DB (version, upgraded_at)
  │     └── Clean up backup
  │
  └── On failure:
        ├── Restore from backup
        └── Return error with diagnostics
```

### 5.4 Integration Remove Flow

```
User
  │
  ▼
CLI: portfolio integration remove claude-code
  │
  ▼
IntegrationManager.Remove("claude-code")
  │
  ├── Load integration metadata from DB
  │
  ├── Warn if active (analyses in flight?)
  │     └── Prompt for confirmation
  │
  ├── Execute integration's Remove()
  │     ├── Unregister MCP tools
  │     ├── Remove agent-specific configs/files
  │     └── Clean up installed artifacts
  │
  ├── Delete integration metadata from DB
  │
  └── Return success
```

---

## 6. Dependencies

### 6.1 Blocking Dependencies

| Dependency | Epic/Component | Rationale |
|------------|---------------|-----------|
| Epic 7 — MCP Server | MCP Tool Definitions | Integration validates MCP reachability and tool availability. MCP server must exist first. |
| Database Schema | Epic 1 — Database | Integration metadata requires a persistence layer (`configuration` table or new `integrations` table). |
| CLI Framework | Epic 0 — CLI Scaffolding | Integration management commands require CLI infrastructure (cobra/urfave-cli). |

### 6.2 Internal Dependencies

| Dependency | Story | Rationale |
|------------|-------|-----------|
| Story 8.1 — Interface | Must exist before 8.2 | Installation framework depends on the `Integration` interface. |
| Story 8.2 — Install | Must exist before 8.3 | Validation operates on installed integrations. |
| Story 8.3 — Validate | Must exist before 8.4 | Upgrade uses Validate() post-install to confirm success. |

### 6.3 External Dependencies

| Dependency | Purpose |
|------------|---------|
| Go standard library + interfaces | No external dependency injection framework — use Go interfaces |
| SQLite (via database layer) | Integration metadata storage |
| MCP SDK (Go) | MCP server reachability checks during validation |
| File system | Integration artifact storage, binary path verification |
| `os/exec` | Check agent binary existence and version |

### 6.4 Non-Dependencies (Explicitly Out of Scope)

| What | Why |
|------|-----|
| AI agent packages in engine | ADR-013 — engine is agent-agnostic |
| Dashboard integration management | CLI is the administrative interface (Guideline.md) |
| MCP-based integration management | CLI only — prevents circular dependency |
| Automatic discovery of new agents | User must explicitly install — no scan for agents |
| Integration marketplace or registry | Out of scope for M2; future consideration |
| Cloud-based integration downloads | Local-first — integrations ship with engine or are installed from local packages |

---

## 7. Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Storage for integration metadata | `configuration` table (key/value) or dedicated `integrations` table | Keep it simple; a dedicated table gives better query capabilities for list/status. |
| Integration packaging | Loose files at a known path (not Go plugins) | Avoids Go plugin complexity and portability issues. Each integration is a directory with metadata + scripts. |
| Version numbering | Semantic Versioning (SemVer) | Industry standard; enables compatibility checks. |
| Rollback mechanism | File-level backup of integration directory before upgrade | Simple, local-first, no external state needed. |
| Self-heal scope | Only recoverable issues (restart server, recreate config) | Avoids over-engineering; complex failures reported to user. |
| Concurrent operation safety | File lock on integration metadata file/directory | Simple, reliable, works across processes. |
