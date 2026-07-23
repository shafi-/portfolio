# Epic 9 — Claude Code Integration: Requirements

**Milestone:** 2 — Agent Integration
**Status:** Draft
**Version:** 1.0

---

## 1. Feature Overview

Epic 9 implements the Claude Code integration for Portfolio. It is the first concrete agent integration built on top of the Integration Framework (Epic 8). The integration enables Claude Code to discover, explore, and reason about a developer's software portfolio through MCP tools and a Portfolio-specific skill.

The integration is an installable component (per ADR-013) that:
- Registers the Portfolio MCP server with Claude Code
- Installs a Portfolio skill (prompt/instructions) into Claude Code
- Provides verification and upgrade capabilities

The Engine itself remains agent-agnostic — all Claude-specific behavior lives in this integration.

---

## 2. Requirements

### 2.1 Functional Requirements

#### FR1 — MCP Server Registration (Story 9.1)
- `portfolio install claude` must register the Portfolio MCP server with Claude Code's configuration.
- Registration must configure the server connection (transport, path, or socket).
- The command must verify MCP tools are available after registration by spawning the MCP server, calling `health()`, and terminating.
- Registration must be idempotent — re-running must not duplicate entries.

#### FR2 — Skill Installation (Story 9.2)
- `portfolio install claude` must install a Portfolio skill/prompt for Claude Code.
- The skill must describe all available MCP tools and their usage.
- The skill must include example queries for common Portfolio interactions.
- The skill must be stored in a location Claude Code can discover.
- The skill must be upgradable alongside the integration.

#### FR3 — Integration Verification (Story 9.3)
- `portfolio doctor` must check the Claude Code integration status.
- Verification must confirm MCP server connection is active.
- Verification must test that MCP tools respond correctly.
- Verification must report issues with actionable remediation steps.
- Verification must report success with version information.

#### FR4 — Integration Upgrade (Story 9.4)
- `portfolio upgrade claude` must update the MCP server and skill to the latest version.
- Upgrade must preserve user configuration.
- Upgrade must report what changed (version diff, changelog).
- Upgrade must be idempotent — re-running after completion reports "already up to date."
- Upgrade must support rollback on failure.

#### FR5 — Uninstall (Story 9.5)
- `portfolio uninstall claude` must remove the MCP server registration from Claude Code configuration.
- Uninstall must remove the Portfolio skill file from Claude Code's skills directory.
- Uninstall must preserve all project data (uninstall is agent-integration only).
- Uninstall must be idempotent — re-running after removal reports "not installed."

### 2.2 Non-Functional Requirements

| ID | Requirement |
|----|-------------|
| NFR1 | The integration must not modify the Portfolio Engine — all Claude-specific code lives in the integration package. |
| NFR2 | Installation must complete in under 5 seconds on modern hardware. |
| NFR3 | Verification must complete in under 3 seconds. |
| NFR4 | The integration must not require network access — all operations are local. |
| NFR5 | The skill/prompt must be plain text or markdown — no binary formats. |
| NFR6 | All integration operations must be idempotent. |
| NFR7 | The integration must preserve user configuration across upgrades. |
| NFR8 | The integration must report clear, actionable error messages. |

---

## 3. Edge Cases & Error Handling

| Scenario | Handling |
|----------|----------|
| Claude Code not installed | Detect absence, report clear error with installation instructions, exit non-zero. |
| MCP server already registered | Idempotent — skip registration, report "already registered." |
| MCP server process fails to start | Capture stderr, report failure with diagnostics, suggest checking port/socket conflicts. |
| MCP tools not responding | Timeout after configurable duration, report which tools failed, suggest restart. |
| Skill file already exists | Idempotent — overwrite with latest version, report "updated." |
| Claude Code config file missing | Create it if possible; if not, report path and manual steps. |
| Permission denied on config write | Report path and suggest `sudo` or manual edit. |
| Integration not installed (doctor) | Report "not installed" with install instructions. |
| Upgrade with no previous version | Report "no previous version found" and exit. |
| Upgrade fails mid-way | Roll back to previous state, report failure with diagnostics. |
| Engine version incompatible | Check compatibility before upgrade, report version mismatch with required version. |
| MCP server binary missing/corrupt | Report path, suggest reinstall. |
| Multiple Claude Code config locations | Detect all possible locations, prefer standard, fall back gracefully. |

---

## 4. Data Flow

### 4.1 Installation Flow (`portfolio install claude`)

```
User
  │
  ▼
portfolio install claude
  │
  ├── 1. Detect Claude Code installation
  │     ├── Found → continue
  │     └── Not found → error with install instructions, exit 1
  │
  ├── 2. Register MCP server
  │     ├── Write MCP config to Claude Code config file
  │     │     (e.g., ~/.claude/settings.json or claude_desktop_config.json)
  │     └── Configure transport (stdio) and server path
  │
  ├── 3. Install Portfolio skill
  │     ├── Copy skill file to Claude Code skills directory
  │     └── Skill describes MCP tools + query examples
  │
  ├── 4. Verify installation (spawn MCP server same way Claude Code does)
  │     ├── Exec `portfolio mcp`, connect via stdio
  │     ├── Send initialize + health() JSON-RPC → expect OK
  │     ├── Terminate server after verification
  │     └── Report success or failure with diagnostics
  │
  └── 5. Report result
        ├── Success: "Claude Code integration installed. MCP tools available."
        └── Failure: error details + remediation steps

### 4.2 Verification Flow (`portfolio doctor`)

```
portfolio doctor
  │
  ├── 1. Check integration metadata in database
  │     ├── Found → continue
  │     └── Not found → report "not installed" with install instructions
  │
  ├── 2. Check MCP server process
  │     ├── Running → continue
  │     └── Not running → report with start instructions
  │
  ├── 3. Call health() tool
  │     ├── Returns OK → continue
  │     └── Error → report with diagnostics
  │
  ├── 4. Call listProjects() tool
  │     ├── Returns results → continue
  │     └── Error → report with diagnostics
  │
  ├── 5. Check skill file exists
  │     ├── Found → report version
  │     └── Missing → report with reinstall instructions
  │
  └── 6. Report summary
        ├── All checks pass → "Claude Code integration: healthy"
        └── Any check fails → detailed report with remediation per check

### 4.3 Upgrade Flow (`portfolio upgrade claude`)

Manager (Epic 8) orchestrates; integration handles agent-specific mutations only.

```
portfolio upgrade claude
  │
  ├── [Manager] 1. Check current integration version
  │     ├── Read version from integration metadata
  │     └── If not installed → error with install instructions
  │
  ├── [Manager] 2. Check engine compatibility
  │     ├── Compatible → continue
  │     └── Incompatible → error with required engine version
  │
  ├── [Manager] 3. Check for latest version
  │     ├── New version available → continue
  │     └── Already latest → report "already up to date", exit 0
  │
  ├── [Manager] 4. Snapshot current state (file-level backup)
  │     ├── Copy integration/<name>/ to backup/
  │     └── Covers: MCP config, skill file, integration metadata
  │
  ├── [Integration] 5. Apply agent-specific upgrade
  │     ├── Write updated MCP config (mcpServers.portfolio entry)
  │     ├── Write updated skill file
  │     └── Update integration metadata version
  │
  ├── [Manager] 6. Verify upgrade
  │     ├── Run Validate() on upgraded integration
  │     ├── Pass → continue
  │     └── Fail → restore from backup, report failure
  │
  └── [Manager] 7. Report result
        ├── Success: delete backup, version before → after, summary of changes
        └── Failure: rollback confirmation, error details
```


---

## 5. Acceptance Criteria

### Story 9.5 — Uninstall Integration
| ID | Criterion |
|----|-----------|
| AC-9.5.1 | `portfolio uninstall claude` exists and is discoverable via `--help`. |
| AC-9.5.2 | Command removes MCP server registration from Claude Code config. |
| AC-9.5.3 | Command removes Portfolio skill file. |
| AC-9.5.4 | All project data is preserved after uninstall. |
| AC-9.5.5 | Re-running uninstall after removal reports "not installed." |

### Story 9.1 — Install MCP

| ID | Criterion |
|----|-----------|
| AC-9.1.1 | `portfolio install claude` exists and is discoverable via `--help`. |
| AC-9.1.2 | Command registers the Portfolio MCP server in Claude Code's MCP configuration. |
| AC-9.1.3 | Server connection is configured (transport type, path/port). |
| AC-9.1.4 | Command verifies MCP tools are available after registration. |
| AC-9.1.5 | Re-running the command is idempotent — reports "already installed." |
| AC-9.1.6 | If Claude Code is not detected, command exits with clear instructions. |

### Story 9.2 — Install Portfolio Skill

| ID | Criterion |
|----|-----------|
| AC-9.2.1 | A Portfolio skill file is installed into Claude Code's skill directory. |
| AC-9.2.2 | The skill describes all MCP tools (health, discoverProjects, listProjects, getProject, searchProjects, searchDocumentation, getAnalysis, storeAnalysis, listProjectsNeedingAnalysis, getConfiguration, updateConfiguration, listRelationships). |
| AC-9.2.3 | The skill includes at least 3 example queries (e.g., "What projects do I have?", "Show me analysis for project X", "Find projects using React"). |
| AC-9.2.4 | The skill is stored as a plain text or markdown file. |
| AC-9.2.5 | The skill is upgradable — `portfolio upgrade claude` replaces the skill file. |

### Story 9.3 — Verify Integration

| ID | Criterion |
|----|-----------|
| AC-9.3.1 | `portfolio doctor` includes a Claude Code integration check. |
| AC-9.3.2 | Check verifies MCP server connection is active. |
| AC-9.3.3 | Check tests that MCP tools respond correctly (e.g., `health()` returns OK). |
| AC-9.3.4 | Check reports the installed integration version. |
| AC-9.3.5 | If issues are found, check reports them with specific remediation steps. |
| AC-9.3.6 | If integration is not installed, check reports "not installed" with install instructions. |

### Story 9.4 — Update Integration

| ID | Criterion |
|----|-----------|
| AC-9.4.1 | `portfolio upgrade claude` exists and is discoverable via `--help`. |
| AC-9.4.2 | Command writes updated MCP config file (`mcpServers.portfolio` entry) and skill file for the new integration version. |
| AC-9.4.3 | Command updates the skill file to the latest version. |
| AC-9.4.4 | User configuration is preserved across upgrades. |
| AC-9.4.5 | Command reports what changed (version before → after, changelog summary). |
| AC-9.4.6 | If already at latest version, command reports "already up to date" and exits cleanly. |
| AC-9.4.7 | On failure, command rolls back to the previous state. |
| AC-9.4.8 | Command checks engine compatibility before upgrading. |

---

## 6. Dependencies

### 5.1 Internal Dependencies

| Dependency | Story | Notes |
|------------|-------|-------|
| Epic 7 — MCP Server | 7.1–7.5 | MCP server must exist before integration can register it. |
| Epic 8 — Integration Framework | 8.1–8.4 | Integration abstraction, registration, validation, and upgrade mechanism must exist. |
| Story 8.2 — Installation Framework | 8.2 | Integration registration command and metadata storage. |
| Story 8.3 — Validation | 8.3 | Validation interface used by `portfolio doctor`. |
| Story 8.4 — Upgrade Mechanism | 8.4 | Upgrade interface used by `portfolio upgrade claude`. |

### 5.2 External Dependencies

| Dependency | Purpose |
|------------|---------|
| Claude Code CLI | Target for MCP registration and skill installation. |
| Claude Code config file | MCP server configuration (e.g., `claude_desktop_config.json` or `~/.claude/settings.json`). |
| Claude Code skills directory | Location for Portfolio skill file. |
| MCP stdio transport | Communication protocol between Claude Code and Portfolio Engine. |

### 5.3 Dependency Graph

```
Epic 7 (MCP Server)
    │
    ▼
Epic 8 (Integration Framework)
    │
    ├── 8.1 Integration Abstraction
    ├── 8.2 Installation Framework
    ├── 8.3 Validation
    └── 8.4 Upgrade Mechanism
    │
    ▼
Epic 9 (Claude Code Integration)
    │
    ├── 9.1 Install MCP (blocked by 8.2)
    ├── 9.2 Install Portfolio Skill (blocked by 9.1)
    ├── 9.3 Verify Integration (blocked by 9.2)
    ├── 9.4 Update Integration (blocked by 8.4)
    └── 9.5 Uninstall Integration (blocked by 9.4)
```

---

## 7. Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| MCP transport | stdio | Simplest, no port conflicts, follows MCP conventions. |
| Skill format | Plain text / Markdown | Claude Code accepts markdown skills; no binary format needed. |
| Config preservation | Backup before upgrade | Enables rollback on failure. |
| Idempotency | Check-before-write | All install/upgrade operations check state before mutating. |
| Error reporting | Structured diagnostics | Each check in `doctor` produces a pass/fail with message. |

---

## 8. Open Questions

| Question | Context |
|----------|---------|
| Where exactly does Claude Code store MCP config? | Need to verify: `claude_desktop_config.json` vs `~/.claude/settings.json` vs both. |
| Where does Claude Code store skills? | Need to verify the skills directory path and file format. |
| How does Claude Code discover skills? | Need to verify if skills are auto-discovered or require explicit registration. |
| What is the MCP server binary format? | Go binary — need to decide on naming convention and install path. |
| Should the MCP server run as a persistent daemon or be spawned on demand? | stdio transport typically spawns per-request; confirm this is correct for Claude Code. |

---

## 9. Glossary

| Term | Definition |
|------|------------|
| MCP | Model Context Protocol — the protocol AI agents use to communicate with tools. |
| Integration | An installable package that connects Portfolio to a specific AI agent. |
| Skill | A prompt/instruction file that teaches an AI agent how to use Portfolio. |
| Engine | The deterministic Go binary that performs discovery, metadata extraction, and storage. |
| Agent | An AI coding assistant (e.g., Claude Code) that consumes Portfolio data. |
| stdio | Standard input/output transport for MCP — the server process communicates via stdin/stdout. |
