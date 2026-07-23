# Epic 8 — Integration Framework: Architecture

**Milestone:** 2 — Agent Integration
**Status:** Draft (Corrected)

---

## 1. Overview

Epic 8 implements an integration framework that allows AI coding agents to register, validate, upgrade, and remove themselves as consumers of the Portfolio Engine. Per ADR-013, the engine remains completely agent-agnostic — all agent-specific behavior lives in installable integrations.

The framework provides a common `Integration` interface. Integrations are managed via CLI commands. The engine exposes deterministic capabilities via MCP (Epic 7); the integration layer bridges those capabilities to specific AI agents.

---

## 2. Package Structure

```
internal/
  integration/
    manager.go              — IntegrationManager: orchestrates lifecycle operations
    manager_test.go         — Unit tests for manager
    interface.go            — Integration interface definition
    types.go                — Shared types: IntegrationMeta, ValidationResult, etc.
    errors.go               — Integration-specific error types
    doctor.go               — Doctor/verify pattern for integration health checks
    doctor_test.go          — Doctor tests

    claude-code/            — Claude Code integration (agent-specific)
    codex/                  — Codex CLI integration (agent-specific)
    opencode/               — OpenCode integration (agent-specific)

    testutil/
      fake_integration.go   — Fake integration for testing
      fake_mcp.go           — Fake MCP server for testing

cmd/
  portfolio/
    commands/
      integration.go    — CLI command: portfolio integration <subcommand>
```

---

## 3. Integration Interface

Defined in `internal/integration/interface.go`. The engine imports only this interface — never any agent-specific package.

```go
package integration

type Integration interface {
    Name() string
    AgentType() string
    Install(ctx context.Context, opts InstallOptions) (*InstallResult, error)
    Validate(ctx context.Context) (*ValidationResult, error)
    Upgrade(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error)
    Remove(ctx context.Context) error
}
```

### Supporting Types

```go
type IntegrationMeta struct {
    Name        string    `json:"name"`
    AgentType   string    `json:"agent_type"`
    Version     string    `json:"version"`
    InstallPath string    `json:"install_path"`
    Status      Status    `json:"status"`
    InstalledAt time.Time `json:"installed_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type InstallOptions struct {
    InstallPath string
    Force       bool
}

type InstallResult struct {
    Meta     IntegrationMeta
    Warnings []string
}

type UpgradeOptions struct {
    TargetVersion string
}

type UpgradeResult struct {
    PreviousVersion string
    NewVersion      string
    RolledBack      bool
}

type ValidationResult struct {
    Passed     bool              `json:"passed"`
    Checks     []ValidationCheck `json:"checks"`
}

type ValidationCheck struct {
    Name        string `json:"name"`
    Passed      bool   `json:"passed"`
    Message     string `json:"message,omitempty"`
    Remediation string `json:"remediation,omitempty"`
    SelfHealable bool  `json:"self_healable,omitempty"`
}
```

`Status` is computed at query time from presence of metadata. Values: `"installed" | "not_installed"`. Not persisted separately — derived from whether a row exists in the store.

---

## 4. State Machine

3 states. No intermediate states — operations resolve to success/failure for the caller.

```
not_installed ──install──> installed
not_installed ──install(fail)──> not_installed
installed ──remove──> not_installed
installed ──upgrade──> installed (version updated)
installed ──upgrade(fail,rollback)──> installed (version restored)
installed ──validate──> installed (read-only, no state change)
```

7 transitions. Validate does not change state. Upgrade keeps same state with new version. Rollback restores version.

No `state.go` file. `Status` is embedded in `IntegrationMeta.Status` — set to `"installed"` on successful install, `"not_installed"` on remove.

---

## 5. IntegrationManager

`IntegrationManager` is the central orchestrator. It holds no agent-specific knowledge — it dispatches to `Integration` implementations.

```go
type Manager struct {
    store     Store
    instances map[string]Integration
    mcp       MCPClient
}

func NewManager(store Store, mcpClient MCPClient) *Manager

func (m *Manager) Install(ctx context.Context, name string, opts InstallOptions) (*IntegrationMeta, error)
func (m *Manager) Validate(ctx context.Context, name string) (*ValidationResult, error)
func (m *Manager) Upgrade(ctx context.Context, name string, opts UpgradeOptions) (*IntegrationMeta, error)
func (m *Manager) Remove(ctx context.Context, name string) error
func (m *Manager) List(ctx context.Context) ([]IntegrationMeta, error)
func (m *Manager) Get(ctx context.Context, name string) (*IntegrationMeta, error)
```

Manager instantiates integrations via direct import — no factory registry:

```go
func NewManager(store Store, mcpClient MCPClient) *Manager {
    return &Manager{
        store: store,
        mcp:   mcpClient,
        integrations: map[string]Integration{
            "claude-code": claude.New(store, mcpClient),
            "codex":       codex.New(store, mcpClient),
            "opencode":    opencode.New(store, mcpClient),
        },
    }
}
```

Adding a new agent requires a new package + import line. Config-driven factory with `init()` registration adds indirection for no practical benefit — you can't add an agent by editing YAML alone; Go code change is always required.

### Store Interface

```go
type Store interface {
    SaveIntegration(ctx context.Context, meta IntegrationMeta) error
    GetIntegration(ctx context.Context, name string) (*IntegrationMeta, error)
    ListIntegrations(ctx context.Context) ([]IntegrationMeta, error)
    DeleteIntegration(ctx context.Context, name string) error
}
```

No backup methods. Backup during upgrade is file-level on disk, not in the configuration table.

### MCPClient Interface

```go
type MCPClient interface {
    Health(ctx context.Context) error
    ListTools(ctx context.Context) ([]string, error)
    RegisterTools(ctx context.Context, tools []ToolDef) error
}

type ToolDef struct {
    Name        string
    Description string
    InputSchema map[string]any
}
```

The Manager calls `mcp.RegisterTools()` after `Integration.Install()` succeeds — this keeps MCP internals out of agent-specific packages.

---

## 6. Subcommand Registration

CLI commands are registered under `portfolio integration`.

```
portfolio integration install <name> [--path <install-path>] [--force]
portfolio integration validate <name>
portfolio integration upgrade <name> [--version <target>]
portfolio integration remove <name>
portfolio integration list
portfolio integration doctor [<name>] [--fix]
```

### Command → Manager Mapping

| CLI Command | Manager Method | Description |
|-------------|---------------|-------------|
| `install <name>` | `Manager.Install(name, opts)` | Register and install integration |
| `validate <name>` | `Manager.Validate(name)` | Run validation checks |
| `upgrade <name>` | `Manager.Upgrade(name, opts)` | Upgrade to new version |
| `remove <name>` | `Manager.Remove(name)` | Unregister and clean up |
| `list` | `Manager.List()` | Enumerate all integrations |
| `doctor [<name>]` | `Manager.Validate(name)` for all or one | Full health check |

### CLI Command Structure

```go
// cmd/portfolio/commands/integration.go

func newIntegrationCmd(manager *integration.Manager) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "integration",
        Short: "Manage AI agent integrations",
    }

    cmd.AddCommand(newInstallCmd(manager))
    cmd.AddCommand(newValidateCmd(manager))
    cmd.AddCommand(newUpgradeCmd(manager))
    cmd.AddCommand(newRemoveCmd(manager))
    cmd.AddCommand(newListCmd(manager))
    cmd.AddCommand(newDoctorCmd(manager))

    return cmd
}
```

---

## 7. Config File Management

### Database Storage

Integration metadata uses the `configuration` table with key prefix:

| Key | Value |
|-----|-------|
| `integration:<name>:meta` | JSON-serialized `IntegrationMeta` |

### Filesystem Layout

```
~/.portfolio/
  integrations/
    claude-code/
      metadata.json       — IntegrationMeta (cached copy)
      skills/             — Agent-specific skill files
      config.json         — Agent-specific configuration
      backup/             — Pre-upgrade snapshot (created during upgrade, deleted on success)
    codex/
    opencode/
```

The install path defaults to `~/.portfolio/integrations/<name>/` and is configurable via `--path` flag on `install`.

No file lock. Single-user CLI tool. Concurrent `portfolio integration install` from two terminals simultaneously is not a real scenario. If it happens, SQLite serializes the metadata write.

---

## 8. Doctor / Verify Pattern

The `doctor` command provides a comprehensive health check for one or all integrations. It is the primary diagnostic tool.

```
portfolio integration doctor [<name>]
portfolio integration doctor [<name>] --fix
```

- Without `--fix`: runs checks, reports pass/fail with remediation steps
- With `--fix`: runs checks, applies self-healable fixes automatically, re-checks

### Doctor Output

```json
{
  "timestamp": "2026-07-22T10:00:00Z",
  "integrations": [
    {
      "name": "claude-code",
      "version": "1.2.0",
      "status": "installed",
      "checks": [
        {
          "name": "mcp_server_reachable",
          "passed": true,
          "message": "MCP server responding at localhost:9090"
        },
        {
          "name": "mcp_tools_available",
          "passed": true,
          "message": "All 8 required MCP tools available"
        },
        {
          "name": "agent_binary_exists",
          "passed": true,
          "message": "claude binary found at /usr/local/bin/claude"
        },
        {
          "name": "config_file_exists",
          "passed": true,
          "message": "Integration config at ~/.portfolio/integrations/claude-code/config.json"
        }
      ],
      "overall": "healthy"
    }
  ]
}
```

### Doctor Algorithm (report-only)

```
doctor(name):
  1. Load integration metadata from store
  2. If not found → return "not installed" diagnostic
  3. Run Validate() on the integration
  4. For each check:
     a. If passed → record as pass
     b. If failed → record as fail + remediation + self_healable flag
  5. Return structured result
```

### Doctor Algorithm (with --fix)

```
doctor(name, --fix):
  1. Follow report-only algorithm
  2. For each failed check with self_healable == true:
     a. Apply fix (restart MCP, recreate config, recreate dir)
     b. Re-run check
  3. Report final results (pass/self-healed/still-failing)
  4. Non-healable failures are always reported with remediation only
```

### Self-Healable Issues (applied with --fix)

| Issue | Self-Heal Action |
|-------|-----------------|
| MCP server not running | Restart MCP server process |
| Config file missing | Recreate from stored metadata |
| Integration directory missing | Recreate directory structure |

### Non-Healable Issues (Report Only)

| Issue | Remediation |
|-------|-------------|
| Agent binary not found | "Install <agent> first." |
| MCP tools missing | "Reinstall integration: `portfolio integration install <name> --force`" |
| Database unavailable | "Ensure Portfolio is initialized: `portfolio init`" |

---

## 9. Upgrade & Rollback

### Upgrade Flow

```
upgrade(name, targetVersion):
  1. Load current metadata from store
  2. Check engine compatibility:
     integration.min_engine_version <= engine_version
     → fail with message if incompatible
  3. Snapshot current state:
     a. Copy integration/<name>/ to integration/<name>/backup/
  4. Execute integration.Upgrade(targetVersion)
  5. Run Validate() on upgraded integration
  6. On success:
     a. Update metadata (version, updated_at)
     b. Delete backup/
     c. Return success
  7. On failure:
     a. Restore integration/<name>/ from backup/
     b. Restore previous version in metadata
     c. Return error with diagnostics
```

### Rollback Guarantees

- **Atomicity:** Snapshot is taken before any mutation. If upgrade fails at any point, the snapshot is restored.
- **Idempotency:** Running `upgrade` again after a successful upgrade reports "already up to date."
- **Version tracking:** Each integration tracks its own version independently of the engine version.

### Engine Version Compatibility

The engine version (`engine_version`) is a build-time constant injected via `-ldflags`:

```go
// cmd/portfolio/main.go
var Version = "dev" // set by: -ldflags="-X main.Version=v0.5.0"
```

Each integration declares its minimum engine requirement as a field in `IntegrationMeta`:

```go
type IntegrationMeta struct {
    // ...
    MinEngineVersion string `json:"min_engine_version"` // e.g. "v0.4.0"
}
```

The Manager's upgrade flow compares `meta.MinEngineVersion <= engineVersion` using semver comparison.

---

## 10. Error Types

```go
package integration

type Error struct {
    Code    string
    Message string
    Cause   error
}

const (
    ErrCodeNotFound          = "INTEGRATION_NOT_FOUND"
    ErrCodeAlreadyInstalled  = "INTEGRATION_ALREADY_INSTALLED"
    ErrCodeNotInstalled      = "INTEGRATION_NOT_INSTALLED"
    ErrCodeIncompatible      = "INTEGRATION_INCOMPATIBLE"
    ErrCodeUpgradeFailed     = "INTEGRATION_UPGRADE_FAILED"
    ErrCodeRollbackFailed    = "INTEGRATION_ROLLBACK_FAILED"
    ErrCodeStoreUnavailable  = "STORE_UNAVAILABLE"
    ErrCodePermissionDenied  = "PERMISSION_DENIED"
)
```

---

## 11. Test Strategy

### Test Fixtures

All unit tests use `testutil.FakeIntegration` and `testutil.FakeMCPClient` — no real agent binaries or MCP servers.

```go
// testutil/fake_integration.go
type FakeIntegration struct {
    NameValue       string
    AgentTypeValue  string
    InstallFn       func(ctx context.Context, opts InstallOptions) (*InstallResult, error)
    ValidateFn      func(ctx context.Context) (*ValidationResult, error)
    UpgradeFn       func(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error)
    RemoveFn        func(ctx context.Context) error
}
```

### Unit Tests

| Test | What it covers |
|------|---------------|
| `TestManager_Install_Success` | Happy path: install registers metadata, calls Integration.Install |
| `TestManager_Install_Duplicate` | EC-1: name collision rejected |
| `TestManager_Install_Idempotent` | NFR-7: reinstall at same version is no-op |
| `TestManager_Validate_Success` | All checks pass |
| `TestManager_Validate_NotInstalled` | EC-8: validate non-installed integration |
| `TestManager_Upgrade_Success` | Happy path: version updated, backup cleaned |
| `TestManager_Upgrade_Rollback` | EC-5: failure triggers full rollback |
| `TestManager_Upgrade_Idempotent` | Already at latest = no-op |
| `TestManager_Upgrade_Incompatible` | EC-4: engine version mismatch blocked |
| `TestManager_Remove_Success` | Happy path: metadata deleted, artifacts cleaned |
| `TestManager_Remove_NotInstalled` | EC-7: no-op |
| `TestManager_Remove_ActiveWarning` | EC-6: warns if analyses in flight |
| `TestManager_List` | Enumerates all registered integrations |
| `TestDoctor_AllHealthy` | All checks pass |
| `TestDoctor_WithFailures` | Failures reported with diagnostics |
| `TestDoctor_NotInstalled` | Doctor on non-installed integration |
| `TestDoctor_Fix` | --fix applies self-healable repairs |
| `TestStore_SaveAndLoad` | Integration metadata persistence round-trip |
| `TestStore_List` | Store enumerates all integrations |
| `TestStore_Delete` | Store removes integration metadata |

### Integration Tests (require fake MCP server)

| Test | What it covers |
|------|---------------|
| `TestClaudeCodeIntegration_Install` | Full install flow with fake MCP |
| `TestClaudeCodeIntegration_Validate` | Validation checks with fake MCP |
| `TestClaudeCodeIntegration_Upgrade` | Upgrade with rollback verification |
| `TestClaudeCodeIntegration_Remove` | Clean removal |

---

## 12. Implementation Order

### Story 8.1 — Integration Interface (2 days)

1. Define `Integration` interface in `internal/integration/interface.go`
2. Define supporting types in `internal/integration/types.go`
3. Define error types in `internal/integration/errors.go`
4. Write `Store` interface
5. Write `MCPClient` interface
6. Write `testutil/fake_integration.go` and `testutil/fake_mcp.go`

### Story 8.2 — Installation Framework (3 days)

1. Implement `Manager` struct and constructor (direct import of agent packages)
2. Implement `Manager.Install()` — validate preconditions, dispatch, persist metadata
3. Implement `Manager.List()` and `Manager.Get()`
4. Implement `Store` using `configuration` table (key prefix `integration:<name>:meta`)
5. Register CLI commands in `cmd/portfolio/commands/integration.go`
6. Unit tests for install, list, duplicate detection, idempotency

### Story 8.3 — Validation (2 days)

1. Implement `Manager.Validate()` — load metadata, dispatch, collect results
2. Implement `Manager.doctor()` — run validate for all or one integration
3. Implement `--fix` flag logic — apply self-healable fixes, re-check
4. Register `validate` and `doctor` CLI commands
5. Unit tests for validation, doctor report, doctor --fix

### Story 8.4 — Upgrade Mechanism (3 days)

1. Implement `Manager.Upgrade()` — snapshot, dispatch, validate, rollback
2. Implement backup/restore for integration directory on disk
3. Implement engine version compatibility check
4. Register `upgrade` CLI command
5. Unit tests for upgrade, rollback, idempotency, compatibility check

### Agent-Specific Integrations (parallel, ~1 day each)

1. Implement `claude-code` integration package
2. Implement `codex` integration package
3. Implement `opencode` integration package

Each agent-specific package implements the `Integration` interface, knows its agent's binary name, config paths, and skill format.

---

## 13. Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Storage | `configuration` table with key prefix `integration:<name>:meta` | Reuses existing schema; no migration needed |
| Integration packaging | Loose files at known path (not Go plugins) | Avoids Go plugin complexity and portability issues |
| Version numbering | SemVer | Industry standard; enables compatibility checks |
| Rollback mechanism | File-level backup of integration directory before upgrade | Simple, local-first, no external state needed |
| Self-heal scope | Only recoverable issues (restart server, recreate config) | Avoids over-engineering; complex failures reported to user |
| Concurrent safety | None — SQLite serializes metadata writes | Single-user CLI tool; concurrent integration operations not a real scenario |
| Process isolation | `recover()` guards around all integration dispatch calls | Panicking integration is caught and error returned. Process-level isolation deferred post-M2 |
| Integration discovery | Direct import in Manager constructor | 3 agent types don't justify config-driven factory with `init()` registry |
| Doctor approach | Report-only by default; `--fix` applies self-healable fixes | Separates diagnosis from remediation |
| Upgrade backup | File-level on disk, not in configuration table | Simpler, no schema coupling |

---

## 14. Edge Cases & Error Handling

| # | Edge Case | Handling |
|---|-----------|----------|
| EC-1 | Integration name collision on install | Reject with clear error: "Integration '<name>' is already installed. Use `upgrade` or `remove` first." |
| EC-2 | MCP server not running during validation | Report as diagnostic failure with guidance: "MCP server not reachable at <address>. Start with `portfolio mcp start`." |
| EC-3 | Agent binary missing during validation | Report diagnostic failure: "Required agent binary '<path>' not found. Install <agent> first." |
| EC-4 | Upgrade compatibility mismatch | Block upgrade with message: "Integration vX.Y requires engine vA.B.C (current: vD.E.F). Upgrade engine first." |
| EC-5 | Upgrade failure mid-operation | Roll back to previous version. Restore integration files from backup on disk. |
| EC-6 | Remove non-installed integration | Report: "Integration '<name>' is not installed." — no-op |
| EC-7 | Validate non-installed integration | Report: "Integration '<name>' is not installed. Run `portfolio integration install <name>` first." |
| EC-8 | Upgrade non-installed integration | Report: "Integration '<name>' is not installed. Cannot upgrade." |
| EC-9 | Database unavailable | Fail with: "Knowledge store unavailable. Ensure Portfolio is initialized (`portfolio init`)." |
| EC-10 | Path contains spaces/special chars | Properly quote and escape paths in all shell/file operations |
| EC-11 | Permission denied | Fail with clear message: "Permission denied at '<path>'. Run with appropriate permissions or configure install path." |

---

## 15. Dependencies

### Internal

| Dependency | Epic/Component | Rationale |
|------------|---------------|-----------|
| Epic 7 — MCP Server | MCP Tool Definitions | Integration validates MCP reachability and tool availability |
| Database Schema | Epic 1 — Database | Integration metadata requires `configuration` table |
| CLI Framework | Epic 0 — CLI Scaffolding | Integration management commands require CLI infrastructure |

### External

| Dependency | Purpose |
|------------|---------|
| Go standard library | Interfaces, file I/O, os/exec, sync |
| SQLite (via database layer) | Integration metadata storage |
| MCP SDK (Go) | MCP server reachability checks during validation |
| File system | Integration artifact storage, binary path verification |
| `os/exec` | Check agent binary existence and version |

### Non-Dependencies (Explicitly Out of Scope)

| What | Why |
|------|-----|
| AI agent packages in engine | ADR-013 — engine is agent-agnostic |
| Dashboard integration management | CLI is the administrative interface (Guideline.md) |
| MCP-based integration management | CLI only — prevents circular dependency |
| Automatic discovery of new agents | User must explicitly install — no scan for agents |
| Integration marketplace or registry | Out of scope for M2; future consideration |
| Cloud-based integration downloads | Local-first — integrations ship with engine or are installed from local packages |
