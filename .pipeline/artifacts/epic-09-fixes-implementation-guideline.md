# Epic 9 — Claude Code Integration: Implementation Guidelines

**Version:** 1.0
**Milestone:** 2 — Agent Integration
**Status:** Draft
**Purpose:** Detailed implementation guidelines for Epic 9 fixes — technical standards, package organization, implementation order, code patterns, testing strategy, build steps, and quality gates

---

## 1. Technical Standards

### 1.1 Go Version and Toolchain

| Specification | Requirement |
|---------------|-------------|
| **Go Version** | 1.21+ (minimum), 1.22+ recommended |
| **Module Mode** | `go mod` with Go workspaces if needed |
| **Build Tags** | Use `+build e2e` for end-to-end tests |
| **Test Framework** | Standard `testing.T` package |
| **Test Timeout** | 30s for integration tests, 60s for E2E tests |
| **Linting** | `golangci-lint` with Go 1.21+ ruleset |
| **Static Analysis** | `go vet`, `staticcheck`, `gosec` |

### 1.2 Package Standards

#### Package Naming Convention
```go
// Use lower-case, single-word package names
package claude        // internal/integration/claude/
package integration    // internal/integration/
package cli           // internal/cli/

// No underscores or mixedCase in package names
```

#### Import Organization
```go
// Order: stdlib → third-party → internal
import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"

    "github.com/nerddevsltd/portfolio/internal/integration"
    "github.com/nerddevsltd/portfolio/internal/store"
)
```

#### Code Conventions
```go
// Use explicit error wrapping with %w
return fmt.Errorf("installation failed: %w", err)

// Use context.Context for all async operations
func (i *ClaudeCodeIntegration) Install(ctx context.Context, opts InstallOptions) error

// Use interface-first design
type Integration interface {
    Install(ctx context.Context, opts InstallOptions) (*InstallResult, error)
    Upgrade(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error)
    Remove(ctx context.Context) error
    Validate(ctx context.Context) (*ValidationResult, error)
}

// Use structs for options (not positional args)
type InstallOptions struct {
    SkipValidation bool
    Force          bool
}
```

### 1.3 Error Handling Standards

#### Exit Code Convention
```go
const (
    ExitSuccess           = 0
    ExitUserError         = 1  // Missing Claude Code, invalid args
    ExitSystemError       = 2  // Permission denied, disk full
    ExitIntegrationError  = 3  // Integration not found
    ExitVersionError      = 4  // Engine version incompatible
)
```

#### Error Message Pattern
```go
// All errors must include: what + why + how to fix
return fmt.Errorf("installation failed: %w\n\n"+
    "Diagnostics: %s\n"+
    "To fix: Check Claude Code is installed: https://claude.ai/download",
    err, result.Diagnostics)
```

#### Panic Recovery
```go
// Wrap command handlers with panic recovery
func runInstall(cmd *cobra.Command, args []string) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("unexpected error: %v\n\n"+
                "To fix: File a bug report at https://github.com/nerddevsltd/portfolio/issues",
                r)
        }
    }()

    // ... command logic
}
```

### 1.4 Testing Standards

#### Test File Naming
```go
// Use _test.go suffix
internal/integration/claude/lifecycle_test.go
internal/integration/claude/mcp_integration_test.go
internal/cli/claude_test.go

// Use descriptive test names
func TestLifecycle_FullInstallUpgradeUninstall(t *testing.T)
func TestMCPServer_SpawnAndConnect(t *testing.T)
func TestInstallCommand_ErrorHandling(t *testing.T)
```

#### Test Structure
```go
func TestLifecycle_IdempotentInstall(t *testing.T) {
    // Setup
    sandbox := newSandbox(t)

    // Execute
    result1, err := sandbox.integration.Install(context.Background(), InstallOptions{})
    if err != nil {
        t.Fatalf("First install failed: %v", err)
    }

    // Verify
    if !strings.Contains(result1.Message, "installed successfully") {
        t.Error("Expected success message")
    }

    // Execute again
    result2, err := sandbox.integration.Install(context.Background(), InstallOptions{})
    if err != nil {
        t.Fatalf("Second install failed: %v", err)
    }

    // Verify idempotency
    if !strings.Contains(result2.Message, "already installed") {
        t.Error("Expected 'already installed' message")
    }
}
```

#### Test Timeout
```go
// Set appropriate timeouts
func TestLifecycle_FullInstallUpgradeUninstall(t *testing.T) {
    if deadline, ok := t.Deadline(); ok {
        if time.Until(deadline) < 30*time.Second {
            t.Skip("Skipping test, insufficient time")
        }
    }
    // ... test logic
}
```

---

## 2. Package Organization Updates

### 2.1 New Files Structure

```
internal/
├── integration/
│   └── claude/
│       ├── test_sandbox.go              # NEW: Test sandbox infrastructure
│       ├── lifecycle_test.go            # NEW: Lifecycle integration tests
│       ├── mcp_integration_test.go      # NEW: MCP server integration tests
│       ├── error_scenarios_test.go      # NEW: Error scenario tests
│       ├── edge_cases_test.go           # NEW: Edge case tests
│       ├── idempotency_test.go          # NEW: Idempotency tests
│       └── e2e_test.go                  # NEW: End-to-end tests
└── cli/
    └── claude_test.go                   # NEW: CLI command tests

scripts/
└── verify-epic-09.sh                    # NEW: Quality gate verification script
```

### 2.2 Existing Files to Update

```
internal/
├── cli/
│   ├── root.go                          # UPDATE: Register all subcommands
│   ├── install.go                       # UPDATE: Add claude subcommand
│   ├── upgrade.go                       # UPDATE: Add claude subcommand
│   ├── uninstall.go                     # UPDATE: Add claude subcommand
│   └── doctor.go                        # UPDATE: Add claude subcommand
```

### 2.3 File Dependencies

| File | Depends On |
|------|------------|
| `internal/integration/claude/test_sandbox.go` | `internal/integration/`, `internal/store/`, `testing` |
| `internal/integration/claude/lifecycle_test.go` | `test_sandbox.go`, `ClaudeCodeIntegration` |
| `internal/integration/claude/mcp_integration_test.go` | `cmd/portfolio/mcp` (binary), `exec`, `json` |
| `internal/cli/claude_test.go` | `cobra`, `internal/integration/` |
| `scripts/verify-epic-09.sh` | `go` CLI, `portfolio` binary |

---

## 3. Implementation Order for Each Fix Category

### 3.1 Phase 1 — CLI Integration Foundation (Priority 1)

**Stories:** 9.1, 9.3, 9.4, 9.5

| Order | Task | File | Description |
|-------|------|------|-------------|
| 1.1 | Create command definitions | `internal/cli/claude.go` | Define install, upgrade, uninstall, doctor claude commands with Cobra structs |
| 1.2 | Wire install subcommand | `internal/cli/install.go` | Register install claude subcommand in install command tree |
| 1.3 | Wire upgrade subcommand | `internal/cli/upgrade.go` | Register upgrade claude subcommand in upgrade command tree |
| 1.4 | Wire uninstall subcommand | `internal/cli/uninstall.go` | Register uninstall claude subcommand in uninstall command tree |
| 1.5 | Wire doctor subcommand | `internal/cli/doctor.go` | Register doctor claude subcommand in doctor command tree |
| 1.6 | Implement command handlers | `internal/cli/claude.go` | Write runInstallClaude, runUpgradeClaude, runUninstallClaude, runDoctorClaude functions |
| 1.7 | Add help text | `internal/cli/claude.go` | Add Short, Long, Example, and flags to all commands |
| 1.8 | Add error handling | `internal/cli/claude.go` | Implement panic recovery and error message formatting |
| 1.9 | Add signal handling | `internal/cli/claude.go` | Add SIGINT/SIGTERM handlers with context cancellation |
| 1.10 | Verify CLI wiring | Manual | Run `portfolio --help` and verify all commands appear |

**Acceptance Criteria:** All four commands (install/upgrade/uninstall/doctor claude) appear in `portfolio --help`, subcommands appear in parent command help, commands are executable.

---

### 3.2 Phase 2 — CLI Command Tests (Priority 2)

**Stories:** 9.1, 9.3, 9.4, 9.5

| Order | Task | File | Description |
|-------|------|------|-------------|
| 2.1 | Create test file | `internal/cli/claude_test.go` | Create CLI test file with test setup |
| 2.2 | Test command wiring | `internal/cli/claude_test.go` | TC-FIX-CLI-001 to TC-FIX-CLI-006: verify all commands are discoverable |
| 2.3 | Test help text | `internal/cli/claude_test.go` | TC-FIX-CLI-007 to TC-FIX-CLI-010: verify help text completeness |
| 2.4 | Test error handling | `internal/cli/claude_test.go` | TC-FIX-CLI-011 to TC-FIX-CLI-013: verify error messages and exit codes |
| 2.5 | Test panic recovery | `internal/cli/claude_test.go` | TC-FIX-CLI-014: verify panics are caught |
| 2.6 | Test signal handling | `internal/cli/claude_test.go` | TC-FIX-CLI-015: verify SIGINT/SIGTERM handling |
| 2.7 | Test verbose mode | `internal/cli/claude_test.go` | TC-FIX-CLI-016: verify stack traces in verbose mode |

**Acceptance Criteria:** All CLI command tests pass, coverage for CLI package reaches 75%+.

---

### 3.3 Phase 3 — Integration Test Infrastructure (Priority 1)

**Stories:** 9.1, 9.2, 9.3, 9.4, 9.5

| Order | Task | File | Description |
|-------|------|------|-------------|
| 3.1 | Create sandbox type | `internal/integration/claude/test_sandbox.go` | Define sandbox struct with temp config, skills dir, db |
| 3.2 | Implement newSandbox() | `internal/integration/claude/test_sandbox.go` | Create isolated test environment with temp directories |
| 3.3 | Add mock MCP client | `internal/integration/claude/test_sandbox.go` | Create MockMCPClient for testing without real server |
| 3.4 | Add helper functions | `internal/integration/claude/test_sandbox.go` | Add hashFile(), verifyConfig(), verifySkill() helpers |
| 3.5 | Test sandbox cleanup | `internal/integration/claude/test_sandbox.go` | Verify temp directories are cleaned up after tests |

**Acceptance Criteria:** Sandbox creates isolated test environment, mock MCP client works, cleanup succeeds.

---

### 3.4 Phase 4 — Lifecycle Integration Tests (Priority 1)

**Stories:** 9.1, 9.2, 9.4, 9.5

| Order | Task | File | Description |
|-------|------|------|-------------|
| 4.1 | Test full lifecycle | `internal/integration/claude/lifecycle_test.go` | TC-FIX-IT-001: install → verify → upgrade → verify → uninstall → verify |
| 4.2 | Test idempotent install | `internal/integration/claude/lifecycle_test.go` | TC-FIX-IT-002: second install reports "already installed" |
| 4.3 | Test idempotent uninstall | `internal/integration/claude/lifecycle_test.go` | TC-FIX-IT-003: second uninstall reports "not installed" |
| 4.4 | Test idempotent upgrade | `internal/integration/claude/lifecycle_test.go` | TC-FIX-IT-004: second upgrade reports "already latest" |
| 4.5 | Test config preservation | `internal/integration/claude/lifecycle_test.go` | TC-FIX-IT-005: user settings preserved across upgrade |
| 4.6 | Test rollback on failure | `internal/integration/claude/lifecycle_test.go` | TC-FIX-IT-006, TC-FIX-IT-007: upgrade failure restores state |
| 4.7 | Test clean uninstall | `internal/integration/claude/lifecycle_test.go` | TC-FIX-IT-008: uninstall removes all artifacts |

**Acceptance Criteria:** All lifecycle tests pass, idempotency verified, rollback works, clean uninstall confirmed.

---

### 3.5 Phase 5 — MCP Server Integration Tests (Priority 2)

**Stories:** 9.1, 9.3

| Order | Task | File | Description |
|-------|------|------|-------------|
| 5.1 | Build MCP binary | Test setup | Build `portfolio mcp` binary before tests |
| 5.2 | Test MCP spawn | `internal/integration/claude/mcp_integration_test.go` | TC-FIX-IT-009: spawn and connect via stdio |
| 5.3 | Test health() tool | `internal/integration/claude/mcp_integration_test.go` | TC-FIX-IT-010: health() returns OK |
| 5.4 | Test listProjects() tool | `internal/integration/claude/mcp_integration_test.go` | TC-FIX-IT-011: listProjects() returns valid data |
| 5.5 | Test timeout handling | `internal/integration/claude/mcp_integration_test.go` | TC-FIX-IT-012: handle server timeout |
| 5.6 | Test invalid response | `internal/integration/claude/mcp_integration_test.go` | TC-FIX-IT-013: handle malformed JSON-RPC |

**Acceptance Criteria:** All MCP server tests pass, tools callable, errors handled gracefully.

---

### 3.6 Phase 6 — Error Scenario Tests (Priority 2)

**Stories:** 9.1, 9.4

| Order | Task | File | Description |
|-------|------|------|-------------|
| 6.1 | Test Claude Code missing | `internal/integration/claude/error_scenarios_test.go` | TC-FIX-IT-014: install fails with instructions |
| 6.2 | Test permission denied | `internal/integration/claude/error_scenarios_test.go` | TC-FIX-IT-015: handle config write permission errors |
| 6.3 | Test corrupt config | `internal/integration/claude/error_scenarios_test.go` | TC-FIX-IT-016: handle invalid JSON |
| 6.4 | Test concurrent install | `internal/integration/claude/error_scenarios_test.go` | TC-FIX-IT-017: handle parallel operations |
| 6.5 | Test disk full | `internal/integration/claude/error_scenarios_test.go` | TC-FIX-IT-018: handle write failures |
| 6.6 | Test version incompatibility | `internal/integration/claude/error_scenarios_test.go` | TC-FIX-IT-019: reject incompatible engine versions |

**Acceptance Criteria:** All error scenario tests pass, errors include remediation steps, no data corruption.

---

### 3.7 Phase 7 — Edge Case Tests (Priority 3)

**Stories:** 9.1, 9.4

| Order | Task | File | Description |
|-------|------|------|-------------|
| 7.1 | Create edge case test file | `internal/integration/claude/edge_cases_test.go` | Create test file |
| 7.2 | Test missing mcpServers key | `internal/integration/claude/edge_cases_test.go` | AC-TC-002.3: create key if missing |
| 7.3 | Test other integrations preserved | `internal/integration/claude/edge_cases_test.go` | AC-TC-002.4: don't remove other MCP entries |
| 7.4 | Test skill dir create failure | `internal/integration/claude/edge_cases_test.go` | AC-TC-002.5: report error gracefully |
| 7.5 | Test MCP server crash | `internal/integration/claude/edge_cases_test.go` | AC-TC-002.6: cleanup on crash |
| 7.6 | Test network timeout | `internal/integration/claude/edge_cases_test.go` | AC-TC-002.8: timeout on slow responses |

**Acceptance Criteria:** All edge case tests pass, no data loss, graceful degradation.

---

### 3.8 Phase 8 — Idempotency Tests (Priority 3)

**Stories:** 9.1, 9.4, 9.5

| Order | Task | File | Description |
|-------|------|------|-------------|
| 8.1 | Create idempotency test file | `internal/integration/claude/idempotency_test.go` | Create test file |
| 8.2 | Test reinstall cycle | `internal/integration/claude/idempotency_test.go` | AC-TC-003.5: install → uninstall → install |
| 8.3 | Test version cycling | `internal/integration/claude/idempotency_test.go` | AC-TC-003.6: upgrade → downgrade → upgrade |
| 8.4 | Test no duplicate entries | `internal/integration/claude/idempotency_test.go` | AC-TC-003.7, AC-TC-003.8: verify no duplicates |

**Acceptance Criteria:** All idempotency tests pass, no duplicate entries, state consistent.

---

### 3.9 Phase 9 — E2E Tests (Priority 4)

**Stories:** 9.1, 9.2, 9.3, 9.4, 9.5

| Order | Task | File | Description |
|-------|------|------|-------------|
| 9.1 | Create E2E test file | `internal/integration/claude/e2e_test.go` | Create test file with `+build e2e` |
| 9.2 | Test fresh install workflow | `internal/integration/claude/e2e_test.go` | AC-QG-003.2: real install, verify, upgrade, uninstall |
| 9.3 | Test in-place upgrade | `internal/integration/claude/e2e_test.go` | AC-QG-003.3: upgrade with settings preservation |
| 9.4 | Test broken install recovery | `internal/integration/claude/e2e_test.go` | AC-QG-003.4: recover from partial failure |
| 9.5 | Test complete uninstall | `internal/integration/claude/e2e_test.go` | AC-QG-003.5, AC-QG-003.7: verify clean state |

**Acceptance Criteria:** All E2E tests pass with `-tags=e2e`, runs in < 60 seconds, clean state verified.

---

### 3.10 Phase 10 — Quality Gate Verification Script (Priority 1)

**Story:** Epic 9 Fix Verification

| Order | Task | File | Description |
|-------|------|------|-------------|
| 10.1 | Create verification script | `scripts/verify-epic-09.sh` | Create bash script with shebang and error handling |
| 10.2 | Add CLI command checks | `scripts/verify-epic-09.sh` | AC-QG-001.2: verify all commands wired |
| 10.3 | Add test coverage checks | `scripts/verify-epic-09.sh` | AC-QG-001.3: verify ≥ 80% coverage |
| 10.4 | Add integration test checks | `scripts/verify-epic-09.sh` | AC-QG-001.4: run integration tests |
| 10.5 | Add MCP tools checks | `scripts/verify-epic-09.sh` | AC-QG-001.5: verify MCP tools available |
| 10.6 | Add lifecycle checks | `scripts/verify-epic-09.sh` | AC-QG-001.6: verify install → verify → uninstall flow |
| 10.7 | Add pass/fail reporting | `scripts/verify-epic-09.sh` | AC-QG-001.6: output ✓/✗ with details |
| 10.8 | Add non-zero exit on failure | `scripts/verify-epic-09.sh` | AC-QG-001.7: exit 1 if any fail |
| 10.9 | Make script executable | `scripts/verify-epic-09.sh` | `chmod +x scripts/verify-epic-09.sh` |
| 10.10 | Test verification script | Manual | Run script, verify all checks pass |

**Acceptance Criteria:** Script executable, all checks implemented, passes when fixes complete, fails when criteria not met.

---

### 3.11 Phase 11 — Test Coverage Gap Analysis and Filling (Priority 2)

**Stories:** All Epic 9 stories

| Order | Task | File | Description |
|-------|------|------|-------------|
| 11.1 | Generate coverage report | Build | Run `go test -coverprofile=coverage.out ./...` |
| 11.2 | Generate HTML report | Build | Run `go tool cover -html=coverage.out -o coverage.html` |
| 11.3 | Analyze coverage gaps | Manual | Identify uncovered lines in internal/cli/ and internal/integration/claude/ |
| 11.4 | Add CLI coverage tests | `internal/cli/claude_test.go` | Add tests for uncovered CLI code |
| 11.5 | Add integration coverage tests | Various | Add tests for uncovered integration code |
| 11.6 | Verify 80% target | Build | Run verification script, confirm ≥ 80% overall |
| 11.7 | Verify 85% target for integration package | Build | Check internal/integration/claude/ coverage ≥ 85% |

**Acceptance Criteria:** Overall coverage ≥ 80%, integration package coverage ≥ 85%, CLI coverage ≥ 75%.

---

## 4. Code Patterns to Follow

### 4.1 Command Handler Pattern

```go
// Standard command handler structure
func runInstallClaude(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    verbose, _ := cmd.Flags().GetBool("verbose")

    // Get integration from manager
    manager, err := integration.GetManager()
    if err != nil {
        return fmt.Errorf("failed to get integration manager: %w", err)
    }

    integration, err := manager.Get("claude")
    if err != nil {
        return fmt.Errorf("Claude Code integration not found: %w\n\n"+
            "To fix: Run 'portfolio doctor' to check integration status", err)
    }

    // Call install with context
    result, err := integration.Install(ctx, integration.InstallOptions{})
    if err != nil {
        return fmt.Errorf("installation failed: %w\n\n"+
            "Diagnostics: %s\n"+
            "To fix: Check Claude Code is installed: https://claude.ai/download",
            err, result.Diagnostics)
    }

    // Success output
    fmt.Printf("Claude Code integration installed successfully.\n")
    fmt.Printf("Version: %s\n", result.Version)
    if verbose {
        fmt.Fprintf(os.Stderr, "Config path: %s\n", result.ConfigPath)
    }

    return nil
}
```

### 4.2 Test Sandbox Pattern

```go
// Sandbox setup for isolated test environment
func newSandbox(t *testing.T) *sandbox {
    t.Helper()
    dir := t.TempDir()

    configPath := filepath.Join(dir, "settings.json")
    skillsDir := filepath.Join(dir, "skills")
    dbPath := filepath.Join(dir, "portfolio.db")
    os.MkdirAll(skillsDir, 0755)

    // Create in-memory database
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    store := store.New(db)

    // Mock MCP client
    mcpClient := &MockMCPClient{
        healthOK: true,
        tools:    []string{"health", "listProjects", "getProject"},
    }

    integration := &ClaudeCodeIntegration{
        store: store,
        mcp:   mcpClient,
        config: ClaudeConfig{
            ConfigPath: configPath,
            SkillsDir:  skillsDir,
            BinaryPath: "/usr/local/bin/portfolio",
        },
    }

    return &sandbox{
        dir:         dir,
        configPath:  configPath,
        skillsDir:   skillsDir,
        dbPath:      dbPath,
        integration: integration,
        mcpClient:   mcpClient,
    }
}
```

### 4.3 Mock MCP Client Pattern

```go
// Mock MCP client for testing without real server
type MockMCPClient struct {
    healthOK       bool
    tools         []string
    simulateError error
}

func (m *MockMCPClient) Initialize() error {
    return m.simulateError
}

func (m *MockMCPClient) CallTool(name string, args map[string]interface{}) (interface{}, error) {
    if m.simulateError != nil {
        return nil, m.simulateError
    }

    switch name {
    case "health":
        if !m.healthOK {
            return nil, fmt.Errorf("health check failed")
        }
        return map[string]interface{}{
            "status": "OK",
            "version": "1.0.0",
        }, nil
    case "listProjects":
        return []interface{}{
            map[string]interface{}{
                "id":   "project-1",
                "name": "Project 1",
                "path": "/path/to/project1",
            },
        }, nil
    default:
        return nil, fmt.Errorf("unknown tool: %s", name)
    }
}
```

### 4.4 Idempotency Check Pattern

```go
// Idempotency check before operation
func (i *ClaudeCodeIntegration) Install(ctx context.Context, opts InstallOptions) (*InstallResult, error) {
    // Check if already installed
    config, err := i.readConfig()
    if err == nil {
        if _, exists := config.MCPServers["portfolio"]; exists {
            return &InstallResult{
                Message: "Claude Code integration already installed",
                Version: config.MCPServers["portfolio"].Version,
            }, nil
        }
    }

    // ... proceed with install
}
```

### 4.5 Rollback Pattern

```go
// Rollback on upgrade failure
func (i *ClaudeCodeIntegration) Upgrade(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error) {
    // Save current state
    configBackup, err := i.readConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to backup config: %w", err)
    }

    skillBackup, err := os.ReadFile(i.config.SkillPath)
    if err != nil {
        return nil, fmt.Errorf("failed to backup skill: %w", err)
    }

    // Attempt upgrade
    err = i.performUpgrade(ctx, opts)
    if err != nil {
        // Rollback
        i.restoreConfig(configBackup)
        os.WriteFile(i.config.SkillPath, skillBackup, 0644)

        return nil, fmt.Errorf("upgrade failed, rolled back: %w", err)
    }

    return &UpgradeResult{
        Message:   "Upgrade successful",
        NewVersion: opts.TargetVersion,
    }, nil
}
```

### 4.6 Error Message Pattern

```go
// Consistent error message format
func formatError(what string, why error, howToFix string) error {
    return fmt.Errorf("%s: %w\n\nTo fix: %s", what, why, howToFix)
}

// Usage:
return formatError(
    "Cannot write to Claude Code config file",
    osErr,
    "Ensure you have write permission to ~/.claude/\n"+
        "Or run with: sudo portfolio install claude\n"+
        "See: https://docs.portfolio.ai/troubleshooting/claude-integration",
)
```

---

## 5. Testing Strategy

### 5.1 Test Hierarchy

```
┌─────────────────────────────────────┐
│  E2E Tests (real CLI, real binary)  │  ← Slowest (60s), runs with -tags=e2e
├─────────────────────────────────────┤
│  Integration Tests (sandbox env)    │  ← Medium speed (30s), runs in CI
├─────────────────────────────────────┤
│  Unit Tests (mocked dependencies)   │  ← Fastest (5s), runs in CI and locally
└─────────────────────────────────────┘
```

### 5.2 Test Execution

```bash
# Run all unit tests (excludes E2E)
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run E2E tests only
go test -tags=e2e ./internal/integration/claude/...

# Run specific test suites
go test -run TestLifecycle ./internal/integration/claude/...
go test -run TestMCPServer ./internal/integration/claude/...
go test -run TestError ./internal/integration/claude/...

# Run with verbose output
go test -v ./internal/cli/...

# Run with race detection
go test -race ./...
```

### 5.3 CI Integration

```yaml
# .github/workflows/epic-09-verify.yml
name: Epic 9 Verification

on:
  pull_request:
    paths:
      - 'internal/integration/claude/**'
      - 'internal/cli/**'
      - 'scripts/verify-epic-09.sh'

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run verification script
        run: ./scripts/verify-epic-09.sh
      - name: Upload coverage report
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: coverage-report
          path: coverage.html
```

### 5.4 Test Data Management

```go
// Use temp directories for test data
func TestLifecycle_FullInstallUpgradeUninstall(t *testing.T) {
    dir := t.TempDir()
    configPath := filepath.Join(dir, "settings.json")
    skillsDir := filepath.Join(dir, "skills")

    // t.TempDir() is automatically cleaned up
}

// Use in-memory databases for tests
func TestInstall_Success(t *testing.T) {
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()

    store := store.New(db)
    // ... test logic
}
```

---

## 6. Build and Verification Steps

### 6.1 Local Build

```bash
# Build portfolio binary
go build -o portfolio ./cmd/portfolio

# Build MCP server binary
go build -o portfolio-mcp ./cmd/mcp

# Verify binary works
./portfolio --help
./portfolio-mcp --help
```

### 6.2 Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...

# Check coverage
go tool cover -func=coverage.out | grep total

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### 6.3 Running Verification Script

```bash
# Make script executable (if not already)
chmod +x scripts/verify-epic-09.sh

# Run verification
./scripts/verify-epic-09.sh

# Run with verbose output
VERBOSE=1 ./scripts/verify-epic-09.sh

# Check exit code
echo $?
# 0 = all passed, 1 = some checks failed
```

### 6.4 Manual Verification

```bash
# Verify CLI commands wired
portfolio --help | grep -E "(install|upgrade|uninstall|doctor)"

# Verify subcommands
portfolio install --help | grep claude
portfolio upgrade --help | grep claude
portfolio uninstall --help | grep claude
portfolio doctor --help | grep claude

# Verify help text
portfolio install claude --help
portfolio upgrade claude --help
portfolio uninstall claude --help
portfolio doctor claude --help

# Verify install lifecycle (in sandbox)
sandbox=$(mktemp -d)
CLAUDE_CONFIG_PATH="$sandbox/settings.json" \
CLAUDE_SKILLS_DIR="$sandbox/skills" \
portfolio install claude

CLAUDE_CONFIG_PATH="$sandbox/settings.json" \
portfolio doctor claude

CLAUDE_CONFIG_PATH="$sandbox/settings.json" \
portfolio uninstall claude

# Verify clean state
cat "$sandbox/settings.json" | grep portfolio || echo "Clean (no portfolio entry)"
rm -rf "$sandbox"
```

### 6.5 CI Verification

```bash
# In CI pipeline:
1. Checkout code
2. Setup Go 1.22
3. Run verification script: ./scripts/verify-epic-09.sh
4. Upload coverage report as artifact
5. Fail PR if verification script exits non-zero
```

---

## 7. Quality Gates

### 7.1 Pre-Merge Requirements

All items must pass before merging to main:

| Category | Requirement | Verification |
|----------|-------------|--------------|
| **CLI Integration** | All commands wired and discoverable | `./scripts/verify-epic-09.sh` check_cli_commands |
| **CLI Integration** | All commands have comprehensive help text | `./scripts/verify-epic-09.sh` check_cli_help |
| **Test Coverage** | Overall coverage ≥ 80% | `./scripts/verify-epic-09.sh` check_test_coverage |
| **Test Coverage** | Integration package coverage ≥ 85% | `go tool cover -func` for internal/integration/claude/ |
| **Test Coverage** | CLI package coverage ≥ 75% | `go tool cover -func` for internal/cli/ |
| **Integration Tests** | All lifecycle tests pass | `go test -run TestLifecycle ./internal/integration/claude/...` |
| **Integration Tests** | All MCP server tests pass | `go test -run TestMCPServer ./internal/integration/claude/...` |
| **Integration Tests** | All error scenario tests pass | `go test -run TestError ./internal/integration/claude/...` |
| **E2E Tests** | All E2E tests pass | `go test -tags=e2e ./internal/integration/claude/...` |
| **Quality Gate** | Verification script passes | `./scripts/verify-epic-09.sh` (exit code 0) |
| **Code Quality** | No lint errors | `golangci-lint run` |
| **Code Quality** | No security vulnerabilities | `gosec ./...` |

### 7.2 Release Criteria

All pre-merge requirements plus:

| Category | Requirement | Verification |
|----------|-------------|--------------|
| **Documentation** | README.md updated with CLI commands | Manual review |
| **Documentation** | CLAUDE.md updated with integration info | Manual review |
| **Documentation** | CHANGELOG.md updated with fix details | Manual review |
| **Backward Compatibility** | Existing installs can upgrade | Manual test on v1.0 install |
| **Rollback** | Upgrade failure restores state | TC-FIX-IT-006, TC-FIX-IT-007 |
| **Idempotency** | All operations idempotent | TC-FIX-IT-002 to TC-FIX-IT-004 |
| **Clean Uninstall** | Uninstall removes all artifacts | TC-FIX-IT-008 |

### 7.3 Coverage Targets

| Package | Target | Current | Gap |
|---------|--------|---------|-----|
| `internal/integration/claude/` | 85% | 48.5% | 36.5% |
| `internal/cli/` | 75% | 30% | 45% |
| Overall | 80% | 48.5% | 31.5% |

### 7.4 Quality Gate Enforcement

```yaml
# CI Pipeline
on: pull_request
jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - checkout
      - setup-go@v5
      - run: ./scripts/verify-epic-09.sh
      - if: failure()
        run: echo "Quality gate failed. See verification output above." && exit 1
```

---

## 8. Dependencies

### 8.1 Internal Dependencies

| Dependency | For | Status |
|------------|-----|--------|
| Epic 8 — Integration Framework | Integration abstraction, Manager, upgrade mechanism | Complete |
| Epic 7 — MCP Server | MCP server binary for integration tests | Complete |
| Original Epic 9 implementation | Integration package (claude/) | Complete |
| Go testing package | Unit and integration tests | Available |
| Cobra CLI framework | Command wiring and help text | Available |

### 8.2 External Dependencies

| Dependency | Purpose | Version |
|------------|---------|---------|
| `github.com/spf13/cobra` | CLI framework | Latest |
| `bc` command | Coverage percentage comparison | 1.0+ |
| `grep` command | Help text verification | Any |
| `bash` | Verification script | 4.0+ |

### 8.3 Toolchain Dependencies

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Build and test |
| bash | 4.0+ | Verification script |
| bc | 1.0+ | Coverage calculations |
| golangci-lint | Latest | Linting |

---

## 9. Success Metrics

| Metric | Target | Current | Gap |
|--------|--------|---------|-----|
| Overall test coverage | ≥ 80% | 48.5% | 31.5% |
| Integration test coverage | 100% | 0% | 100% |
| CLI command coverage | 100% | 0% | 100% |
| E2E tests passing | 100% | N/A | N/A |
| Quality gate verification | Pass | Fail | - |
| All CLI commands wired | 4/4 | 0/4 | 4 |

---

## 10. Open Questions and Decisions

| Question | Context | Decision |
|----------|---------|----------|
| Should E2E tests run in CI by default? | E2E tests are slower | Run with `-tags=e2e` flag, not in default CI |
| Should coverage report be published as artifact? | Track coverage trends | Yes, upload coverage.html as CI artifact |
| Should verification script be run automatically in CI? | PR gate | Yes, add to required checks for PRs affecting Epic 9 |

---

## 11. References

- Original Epic 9 architecture: `.architecture/epic-09-architecture.md`
- Fix requirements: `.requirements/epic-09-fixes-requirements.md`
- Fix test cases: `.test-cases/epic-09-fixes-test-cases.md`
- Context pack: `.context-pack/epic-09-context-pack.md`
- ADR-013: Agent Integration Architecture
- Platform Specification: MCP tools and interfaces
- Engineering Guidelines: Principles for implementation