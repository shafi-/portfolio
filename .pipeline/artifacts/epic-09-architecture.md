# Epic 9 — Claude Code Integration: Architecture

**Version:** 1.0 (Corrected)
**Milestone:** 2 — Agent Integration
**Status:** Draft

---

## 1. Overview

Epic 9 implements the first concrete agent integration (per ADR-013), connecting Claude Code to the Portfolio Engine via the Integration Framework (Epic 8). All Claude-specific behavior lives in the integration package — the Engine remains agent-agnostic.

### Architectural Layering

```
┌─────────────────────────────────────────┐
│           Claude Code (AI Agent)        │
└────────────────┬────────────────────────┘
                 │ MCP stdio transport
┌────────────────▼────────────────────────┐
│        Epic 9 — claude integration      │
│  (install/verify/uninstall)             │
│  Upgrades handled by Epic 8 Manager     │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│     Epic 8 — Integration Framework      │
│  (abstraction, registration, upgrade)   │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│     Epic 7 — MCP Server + Engine        │
└─────────────────────────────────────────┘
```

---

## 2. Package Structure

```
internal/
  integration/
    claude/
      integration.go         — ClaudeCodeIntegration struct (implements Integration)
      mcp_config.go          — MCP config file read/write/merge/remove
      skill.go               — Skill file installation/removal
      verify.go              — Verification checks
      uninstall.go           — Uninstall flow
      paths.go               — Claude Code path detection
      skill.md               — Static Portfolio skill markdown (go:embed)
```

The `claude` package implements `integration.Integration` from Epic 8. It does NOT contain its own upgrade/rollback logic — the Manager handles that.

---

## 3. Integration Interface Implementation

### Epic 8 Interface (`internal/integration/interface.go`)

```go
type Integration interface {
    Name() string
    AgentType() string
    Install(ctx context.Context, opts InstallOptions) (*InstallResult, error)
    Validate(ctx context.Context) (*ValidationResult, error)
    Upgrade(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error)
    Remove(ctx context.Context) error
}
```

### Claude Code Integration (`internal/integration/claude/integration.go`)

```go
type ClaudeCodeIntegration struct {
    store  Store
    mcp    MCPClient
    config ClaudeConfig
}

type ClaudeConfig struct {
    InstallPath string   // Where Claude Code is installed
    ConfigPath  string   // Path to claude_desktop_config.json or ~/.claude/settings.json
    SkillsDir   string   // Claude Code skills directory
    BinaryPath  string   // Path to portfolio MCP binary
}
```

```go
func New(store Store, mcp MCPClient) *ClaudeCodeIntegration {
    cfg := detectPaths()
    return &ClaudeCodeIntegration{
        store:  store,
        mcp:    mcp,
        config: cfg,
    }
}
```

No separate factory or registry — Manager constructs directly.

### Upgrade Method

The integration's `Upgrade()` does only agent-specific work. Backup, rollback, and re-verification are handled by the Manager:

```go
func (c *ClaudeCodeIntegration) Upgrade(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error) {
    if _, err := exec.LookPath("claude"); err != nil {
        return nil, fmt.Errorf("claude binary not found: %w", err)
    }
    if err := c.writeMCPConfig(); err != nil {
        return nil, fmt.Errorf("write MCP config: %w", err)
    }
    if err := c.installSkill(); err != nil {
        return nil, fmt.Errorf("install skill: %w", err)
    }
    return &UpgradeResult{NewVersion: opts.TargetVersion}, nil
}
```

Manager wraps:
```
Manager.Upgrade(name):
  1. Snapshot: copy integration/<name>/ to backup/
  2. Call integration.Upgrade(targetVersion)
  3. Run Validate() on result
  4. Success → delete backup, update version
  5. Failure → restore from backup, return error
```

This is already in Epic 8's design. Epic 9 does NOT reimplement it.

---

## 4. MCP Config File Management

### Config File Location Detection

Claude Code supports multiple config locations. Detection in priority order:

| Priority | Path | Scope |
|----------|------|-------|
| 1 | `~/.claude/settings.json` | Project-agnostic, preferred |
| 2 | `~/Library/Application Support/Claude/claude_desktop_config.json` | macOS desktop app |
| 3 | `$XDG_CONFIG_HOME/claude/settings.json` | Linux |
| 4 | `%APPDATA%/Claude/claude_desktop_config.json` | Windows |

If none exists, the integration creates `~/.claude/settings.json`.

### Config File Structure

```json
{
  "mcpServers": {
    "portfolio": {
      "command": "/path/to/portfolio",
      "args": ["mcp"],
      "transport": "stdio"
    }
  }
}
```

### Read/Write Strategy (`mcp_config.go`)

```go
func (c *ClaudeCodeIntegration) ensureMCPConfig() error
func (c *ClaudeCodeIntegration) removeMCPConfig() error
func (c *ClaudeCodeIntegration) readMCPConfig() (*MCPConfig, error)
```

Rules:
1. **Read first** — never blindly overwrite. Parse existing JSON, merge `mcpServers.portfolio`.
2. **Idempotent write** — if `mcpServers.portfolio` already matches current config, skip write.
3. **Write directly** — use `os.WriteFile`. JSON encoder writes atomically for small payloads. Temp+rename optional as internal implementation detail.

### Idempotency

```go
func (c *ClaudeCodeIntegration) isMCPRegistered() bool
```

Checks if `mcpServers.portfolio` entry exists with matching command/args. If yes, `Install` skips MCP registration.

### Install-Time Verification

After writing config, `Install()` spawns the MCP server process the same way Claude Code would — exec the `portfolio mcp` command, connect via stdio, send `health()` call, verify response, then terminate. This confirms the server binary works and tools are accessible before reporting success.

```go
func (c *ClaudeCodeIntegration) verifyMCPServer(ctx context.Context) error {
    cmd := exec.Command(c.config.BinaryPath, "mcp")
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()
    cmd.Stderr = os.Stderr
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("start MCP server: %w", err)
    }
    defer cmd.Process.Kill()

    // Send initialize + health() via JSON-RPC, expect OK
    // Timeout: 5s total
}
```

If verification fails, `Install()` reports the failure with diagnostics (binary not found, timeout, unexpected response). The config file is left in place (written state is valid even if runtime verify fails — Claude Code may succeed later).

---

## 5. Skill File Installation

### Skill Markdown (`skill.md` — embedded via `//go:embed`)

Static markdown file. No version injection via template. Version is tracked in the integration metadata database, not in the skill file.

```markdown
# Portfolio — Claude Code Skill

Portfolio helps you understand a developer's entire software portfolio.
Use these MCP tools through Claude Code:

## Tools

### health
Check if Portfolio is running.
Usage: `Call health()`

### discoverProjects
Scan configured directories for new projects.
Usage: `Call discoverProjects()`

### listProjects
List all known projects.
Usage: `Call listProjects()`

### getProject
Get details for a specific project.
Usage: `Call getProject(id: "<project-id>")`

### searchProjects
Search projects by name, language, framework.
Usage: `Call searchProjects(query: "react")`

### searchDocumentation
Search within project documentation.
Usage: `Call searchDocumentation(query: "architecture")`

### getAnalysis
Get semantic analysis for a project.
Usage: `Call getAnalysis(projectId: "<project-id>")`

### storeAnalysis
Store semantic analysis for a project.
Usage: `Call storeAnalysis(projectId: "<project-id>", summary: "...", purpose: "...")`

### listProjectsNeedingAnalysis
Find projects missing or with outdated analysis.
Usage: `Call listProjectsNeedingAnalysis()`

### getConfiguration
View Portfolio configuration.
Usage: `Call getConfiguration()`

### updateConfiguration
Update Portfolio configuration.
Usage: `Call updateConfiguration(key: "...", value: "...")`

### listRelationships
List relationships for a project.
Usage: `Call listRelationships(projectId: "<project-id>")`

## Example Workflows

1. "What projects do I have?"
   → `listProjects()`

2. "Show me analysis for project X"
   → `getAnalysis(projectId: "<id>")`

3. "Find projects using React"
   → `searchProjects(query: "react")`

4. "What changed recently?"
   → `discoverProjects()` → `listProjectsNeedingAnalysis()`

5. "How are my projects related?"
   → `listProjects()` → for each: `listRelationships(projectId: "<id>")`
```

### Installation (`skill.go`)

```go
func (c *ClaudeCodeIntegration) installSkill() error   // Write to skills dir
func (c *ClaudeCodeIntegration) removeSkill() error    // Remove from skills dir
func (c *ClaudeCodeIntegration) skillPath() string     // e.g., ~/.claude/skills/portfolio.md
```

Skill file is written to the Claude Code skills directory. Claude Code auto-discovers `.md` files — no explicit registration needed.

Version is tracked in metadata DB, not in the skill file. If the skill content changes between versions, `Upgrade()` rewrites it. Verification checks if the skill file exists, not its content version.

---

## 6. Verification Flow

### Check Sequence (`verify.go`)

| # | Check | Method | Pass Condition | Failure Remediation |
|---|-------|--------|----------------|---------------------|
| 1 | Integration installed | Metadata exists in store | Metadata found | "Run `portfolio install claude`" |
| 2 | Config file exists | `os.Stat(configPath)` | File exists | "Config file missing. Run `portfolio install claude` to recreate." |
| 3 | MCP entry present | `isMCPRegistered()` | `mcpServers.portfolio` exists | "MCP entry missing. Run `portfolio install claude`." |
| 4 | MCP health responds | MCP client `health()` call | Returns OK within timeout | "MCP server not responding. Run `portfolio doctor claude --fix`." |
| 5 | Tools available | Call `listProjects()` via MCP | Returns valid response | "MCP tools not available. Check server logs." |
| 6 | Skill file present | `os.Stat(skillPath())` | File exists | "Skill file missing. Run `portfolio upgrade claude`." |

Each MCP call has a 2-second timeout. Total verification completes within 3 seconds.

### Output Format

```json
{
  "integration": "claude",
  "version": "1.0.0",
  "status": "pass",
  "checks": [
    { "name": "integration_metadata", "status": "pass", "message": "Integration registered" },
    { "name": "config_file", "status": "pass", "message": "Config file found at ~/.claude/settings.json" },
    { "name": "mcp_entry", "status": "pass", "message": "MCP server 'portfolio' registered" },
    { "name": "mcp_health", "status": "pass", "message": "health() returned OK" },
    { "name": "mcp_tools", "status": "pass", "message": "listProjects() responded" },
    { "name": "skill_file", "status": "fail", "message": "Skill file missing",
      "remediation": "Run `portfolio install claude` to reinstall." }
  ]
}
```

---

## 7. Upgrade

Upgrade is handled by the Epic 8 Manager. The integration's `Upgrade()` method performs only agent-specific work (rewrite MCP config, rewrite skill file). See Section 3 above for the exact contract.

The integration does NOT implement:
- Backup/restore (Manager handles this)
- Version comparison (Manager handles this)
- Verify-after-upgrade (Manager calls Validate after upgrade)
- Rollback (Manager restores from backup on failure)

---

## 8. Uninstall Flow

### Uninstall Flow (`uninstall.go`)

```
portfolio uninstall claude
    │
    ├── 1. Check installed state
    │     ├── integration metadata found? → no → "not installed", exit 0
    │     └── found → continue
    │
    ├── 2. Remove MCP config entry
    │     ├── Read config file
    │     ├── Remove mcpServers.portfolio entry
    │     └── Write config file
    │
    ├── 3. Remove skill file
    │     ├── os.Remove(skillPath())
    │     └── If file doesn't exist → skip (idempotent)
    │
    ├── 4. Manager removes integration metadata from DB
    │
    └── 5. Report result
          ├── Success: "Claude Code integration uninstalled. Project data preserved."
          └── Already removed: "Claude Code integration is not installed."
```

### Preservation Guarantees

- Uninstall touches **only** Claude Code artifact files (config + skill).
- Does **not** delete the Portfolio database, project data, or engine binary.
- The MCP binary (`portfolio mcp`) remains available — only the Claude Code registration is removed.
- Re-running uninstall after removal is idempotent.

---

## 9. Path Detection (`paths.go`)

### Config File Discovery Priority

```go
func detectConfigPath() (string, error) {
    candidates := []string{
        filepath.Join(homeDir, ".claude", "settings.json"),
        filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
    }
    if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
        candidates = append([]string{filepath.Join(xdg, "claude", "settings.json")}, candidates...)
    }
    for _, p := range candidates {
        if _, err := os.Stat(p); err == nil {
            return p, nil
        }
    }
    return filepath.Join(homeDir, ".claude", "settings.json"), nil
}
```

### Skills Directory Discovery

```go
func detectSkillsDir() string {
    return filepath.Join(homeDir, ".claude", "skills")
}
```

### Agent Binary Detection

```go
func isClaudeInstalled() bool {
    _, err := exec.LookPath("claude")
    return err == nil
}
```

If not found, `install` reports a clear error with installation instructions.

---

## 10. Test Strategy

### Deterministic Tests (Go `testing` package)

All tests use a temp directory sandbox — never touch real Claude Code config.

| Story | Test Suite | What It Tests | Key Scenarios |
|-------|-----------|---------------|---------------|
| 9.1 | `mcp_config_test.go` | Config file read/write/merge | Write entry to empty config, merge with existing entries, idempotent re-write |
| 9.2 | `skill_test.go` | Skill file install/remove | Write to non-existent dir (creates it), overwrite with newer version, remove missing file |
| 9.3 | `verify_test.go` | Verification checks | Config missing, skill missing, all pass scenario |
| 9.4 | `upgrade_test.go` | Upgrade (agent-specific work only) | Upgrade writes new config + skill, version comparison |
| 9.5 | `uninstall_test.go` | Uninstall flow | Remove MCP entry, remove skill, idempotent re-run, config with other integrations preserved |

### Test Sandbox Pattern

```go
type sandbox struct {
    dir         string
    configPath  string
    skillsDir   string
    integration *ClaudeCodeIntegration
}

func newSandbox(t *testing.T) *sandbox {
    t.Helper()
    dir := t.TempDir()
    configPath := filepath.Join(dir, "settings.json")
    skillsDir := filepath.Join(dir, "skills")
    os.MkdirAll(skillsDir, 0755)
    return &sandbox{...}
}
```

### What Is NOT Tested (per Guideline)

- Actual MCP server process communication — requires running server; covered by E2E tests.
- Claude Code behavior — we test that config files are correctly written, not that Claude Code reads them.
- Rollback/corruption recovery — tested at Epic 8 Manager level, not in agent-specific integration.

---

## 11. Implementation Order

### Phase 1 — Foundation (Story 9.1)

| Step | File | Description |
|------|------|-------------|
| 1.1 | `internal/integration/claude/paths.go` | Claude Code path detection |
| 1.2 | `internal/integration/claude/mcp_config.go` | MCP config read/write/merge/remove |
| 1.3 | `internal/integration/claude/mcp_config_test.go` | Tests for config operations |
| 1.4 | `internal/integration/claude/integration.go` | ClaudeCodeIntegration struct, Install() method with MCP server spawn verification |
| 1.5 | CLI command wiring | `portfolio install claude` → calls integration.Install(), reports diagnostics on verify failure |

### Phase 2 — Skill (Story 9.2)

| Step | File | Description |
|------|------|-------------|
| 2.1 | `internal/integration/claude/skill.md` | Static skill markdown (embedded) |
| 2.2 | `internal/integration/claude/skill.go` | Skill install/remove |
| 2.3 | `internal/integration/claude/skill_test.go` | Tests for skill operations |
| 2.4 | Update `integration.go` | `Install()` also calls `installSkill()` |

### Phase 3 — Verification (Story 9.3)

| Step | File | Description |
|------|------|-------------|
| 3.1 | `internal/integration/claude/verify.go` | Verification checks |
| 3.2 | `internal/integration/claude/verify_test.go` | Tests for each check |
| 3.3 | CLI command wiring | `portfolio doctor` includes claude check |

### Phase 4 — Upgrade (Story 9.4)

| Step | File | Description |
|------|------|-------------|
| 4.1 | Update `integration.go` | `Upgrade()` writes updated config + skill |
| 4.2 | `internal/integration/claude/upgrade_test.go` | Version comparison, agent-specific upgrade tests |
| 4.3 | CLI command wiring | `portfolio upgrade claude` (dispatched by Manager) |

### Phase 5 — Uninstall (Story 9.5)

| Step | File | Description |
|------|------|-------------|
| 5.1 | `internal/integration/claude/uninstall.go` | Uninstall flow |
| 5.2 | `internal/integration/claude/uninstall_test.go` | Idempotency test, data preservation test |
| 5.3 | CLI command wiring | `portfolio uninstall claude` |

---

## 12. Configuration & Metadata Storage

### Integration Metadata in DB

Stored by Epic 8 Manager via the `configuration` table:

| Key | Value |
|-----|-------|
| `integration.claude.meta` | JSON-serialized `IntegrationMeta` |

---

## 13. Error Handling Patterns

All integration commands follow the same error contract:

- **Success**: exit code 0, human-readable success message to stdout
- **User error** (missing Claude Code, wrong args): exit code 1, actionable message to stderr
- **System error** (permission denied, disk full): exit code 2, diagnostic details to stderr

Every error message includes:
1. What went wrong
2. Why it happened (if known)
3. What the user should do next

Example:

```
Error: Cannot write to Claude Code config file

  Tried to write to: /Users/nerddevsltd/.claude/settings.json
  Permission denied.

  To fix:
  1. Ensure you have write permission to ~/.claude/
  2. Or run with: sudo portfolio install claude
  3. Or manually edit the config file (see docs)
```
