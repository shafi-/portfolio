# Epic: OpenCode Integration

## Overview

Add complete OpenCode integration support to Portfolio, providing feature parity with the existing Claude Code integration. OpenCode is an open-source AI coding agent (https://opencode.ai/) that supports multiple AI providers and extensive MCP integration capabilities.

## User Requirements

- **Implementation approach**: Research OpenCode config/format during implementation
- **Platform support**: macOS, Linux, Windows
- **Scope**: Complete integration with feature parity to Claude Code integration
- **Testing**: User has OpenCode installed and will test after implementation
- **Reporting requirement**: Report any features that were implemented for Claude but not for OpenCode

## Background

OpenCode is an open-source AI coding agent that:
- Is terminal-based with desktop and IDE interfaces
- Supports multiple AI providers (Claude, GPT, Gemini, 75+ LLMs)
- Has native MCP (Model Context Protocol) support with 1200+ integrations
- Uses a sophisticated multi-source configuration system
- Supports strict skill format requirements with YAML frontmatter
- Installation: `curl -fsSL https://opencode.ai/install | bash`

## Architecture Differences from Claude Code

### Configuration File Locations

**Claude Code:**
- `~/.claude/settings.json` (primary)
- `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)
- Supports `XDG_CONFIG_HOME` environment variable
- Single settings file approach

**OpenCode:**
- `~/.config/opencode/opencode.json` (main config)
- `~/.config/opencode/opencode.jsonc` (with comments)
- Multiple config sources with precedence order:
  1. Remote config (`.well-known/opencode` endpoint)
  2. Global config (`~/.config/opencode/opencode.json`)
  3. Custom config (`OPENCODE_CONFIG` environment variable)
  4. Project config (`opencode.json` in project root)
  5. `.opencode` directories (for agents, commands, plugins)
  6. Inline config (`OPENCODE_CONFIG_CONTENT` environment variable)
  7. Managed config (system directories)
  8. macOS managed preferences (`.mobileconfig` via MDM)

### MCP Server Registration Format

**Claude Code format:**
```json
{
  "mcpServers": {
    "portfolio": {
      "command": "portfolio",
      "args": ["mcp"],
      "transport": "stdio"
    }
  }
}
```

**OpenCode format:**
```json
{
  "mcp": {
    "portfolio": {
      "type": "local",
      "command": ["portfolio", "mcp"],
      "enabled": true
    }
  }
}
```

**Key differences:**
- Object naming: `mcpServers` vs `mcp`
- Transport specification: `transport: "stdio"` vs `type: "local"`
- Array format: `args: ["mcp"]` vs `command: ["portfolio", "mcp"]`
- Enable flag: OpenCode requires `enabled: true`
- OpenCode also supports remote MCP servers with URL and headers

### Skills/Agents Installation

**Claude Code skills:**
- Location: `~/.claude/skills/<name>/`
- Format: Simple markdown files
- No strict frontmatter requirements
- Direct filesystem installation

**OpenCode skills:**
- Six search locations (in precedence order):
  1. `.opencode/skills/<name>/SKILL.md` (project)
  2. `~/.config/opencode/skills/<name>/SKILL.md` (global)
  3. `.claude/skills/<name>/SKILL.md` (Claude-compatible)
  4. `~/.claude/skills/<name>/SKILL.md` (Claude-compatible global)
  5. `.agents/skills/<name>/SKILL.md` (agent-compatible)
  6. `~/.agents/skills/<name>/SKILL.md` (agent-compatible global)

- **Strict SKILL.md format**:
  ```yaml
  ---
  name: skill-name
  description: 1-1024 character description
  license: optional
  compatibility: optional
  metadata: optional string-to-string map
  ---
  
  # Skill content here
  ```

- **Naming rules**: 
  - 1-64 characters
  - Lowercase alphanumeric with single hyphen separators
  - Must match directory name exactly
  - Cannot start/end with hyphen
  - No consecutive hyphens

### Binary Detection

**Claude Code:**
- Binary: `claude`
- Installation: https://docs.anthropic.com/claude-code

**OpenCode:**
- Binary: `opencode`
- Installation: `curl -fsSL https://opencode.ai/install | bash`

## Implementation Plan

### Phase 1: Create OpenCode Integration Package

**New files to create:**
1. `internal/integration/opencode/integration.go` - Main integration implementation
2. `internal/integration/opencode/paths.go` - Path detection utilities
3. `internal/integration/opencode/mcp_config.go` - MCP server registration
4. `internal/integration/opencode/skill.go` - Skill file installation
5. `internal/integration/opencode/skill.md` - Embedded skill content
6. `internal/integration/opencode/verify.go` - Validation helpers
7. `internal/integration/opencode/*_test.go` - Test files

**Pattern to follow:** Mirror Claude integration structure but adapt for OpenCode's differences

### Phase 2: Implement Path Detection

Create `internal/integration/opencode/paths.go`:

```go
package opencode

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func detectPaths() (OpenCodeConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return OpenCodeConfig{}, err
	}

	configPath, err := detectConfigPath(homeDir)
	if err != nil {
		return OpenCodeConfig{}, err
	}

	skillsDir := detectSkillsDir(homeDir)

	binaryPath, err := detectBinaryPath()
	if err != nil {
		return OpenCodeConfig{}, err
	}

	return OpenCodeConfig{
		InstallPath: detectInstallPath(),
		ConfigPath:  configPath,
		SkillsDir:   skillsDir,
		BinaryPath:  binaryPath,
	}, nil
}

func detectConfigPath(homeDir string) (string, error) {
	candidates := []string{
		filepath.Join(homeDir, ".config", "opencode", "opencode.json"),
		filepath.Join(homeDir, ".config", "opencode", "opencode.jsonc"),
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		candidates = append([]string{
			filepath.Join(xdg, "opencode", "opencode.json"),
		}, candidates...)
	}

	// Return first existing path, or default to creating the first candidate
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	defaultPath := filepath.Join(homeDir, ".config", "opencode", "opencode.json")
	if err := os.MkdirAll(filepath.Dir(defaultPath), 0755); err != nil {
		return "", err
	}

	return defaultPath, nil
}

func detectSkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".config", "opencode", "skills")
}

func detectBinaryPath() (string, error) {
	path, err := exec.LookPath("portfolio")
	if err != nil {
		return "", err
	}
	return path, nil
}

func detectInstallPath() string {
	return filepath.Join(".portfolio", "integrations", "opencode")
}

func isOpenCodeInstalled() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}
```

### Phase 3: Implement MCP Configuration

Create `internal/integration/opencode/mcp_config.go`:

```go
package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type OpenCodeMCPConfig struct {
	MCP map[string]OpenCodeMCPServerConfig `json:"mcp"`
}

type OpenCodeMCPServerConfig struct {
	Type     string              `json:"type"` // "local" or "remote"
	Command  []string           `json:"command,omitempty"`
	URL      string             `json:"url,omitempty"`
	Enabled  bool               `json:"enabled"`
	Headers  map[string]string `json:"headers,omitempty"`
}

func (c *OpenCodeIntegration) ensureMCPConfig() error {
	config, err := c.readMCPConfig()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read MCP config: %w", err)
	}

	if config == nil {
		config = &OpenCodeMCPConfig{
			MCP: make(map[string]OpenCodeMCPServerConfig),
		}
	}

	if config.MCP == nil {
		config.MCP = make(map[string]OpenCodeMCPServerConfig)
	}

	config.MCP["portfolio"] = OpenCodeMCPServerConfig{
		Type:    "local",
		Command: []string{c.config.BinaryPath, "mcp"},
		Enabled: true,
	}

	if err := c.writeMCPConfig(config); err != nil {
		return fmt.Errorf("write MCP config: %w", err)
	}

	return nil
}

func (c *OpenCodeIntegration) removeMCPConfig() error {
	config, err := c.readMCPConfig()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read MCP config: %w", err)
	}

	if config.MCP != nil {
		delete(config.MCP, "portfolio")
	}

	if err := c.writeMCPConfig(config); err != nil {
		return fmt.Errorf("write MCP config: %w", err)
	}

	return nil
}

func (c *OpenCodeIntegration) readMCPConfig() (*OpenCodeMCPConfig, error) {
	data, err := os.ReadFile(c.config.ConfigPath)
	if err != nil {
		return nil, err
	}

	var config OpenCodeMCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse MCP config: %w", err)
	}

	return &config, nil
}

func (c *OpenCodeIntegration) writeMCPConfig(config *OpenCodeMCPConfig) error {
	dir := filepath.Dir(c.config.ConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal MCP config: %w", err)
	}

	if err := os.WriteFile(c.config.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("write MCP config file: %w", err)
	}

	return nil
}

func (c *OpenCodeIntegration) isMCPRegistered() bool {
	config, err := c.readMCPConfig()
	if err != nil {
		return false
	}

	if config.MCP == nil {
		return false
	}

	server, exists := config.MCP["portfolio"]
	if !exists {
		return false
	}

	if server.Type != "local" {
		return false
	}

	if !server.Enabled {
		return false
	}

	// Check command matches
	if len(server.Command) != 2 || server.Command[1] != "mcp" {
		return false
	}

	return true
}
```

### Phase 4: Implement Skill Installation

Create `internal/integration/opencode/skill.md`:

```markdown
---
name: portfolio
description: Portfolio management integration for discovering projects, analyzing codebases, and storing insights. Provides MCP tools for project inventory, documentation search, and AI-powered analysis.
license: MIT
compatibility: opencode >=1.0.110
---

# Portfolio — OpenCode Skill

Portfolio helps you understand a developer's entire software portfolio.
Use these MCP tools through OpenCode:

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

## Important Notes

- **Analyzer Identity**: Always set `analyzer: "opencode"` when calling `storeAnalysis()`
- **Workflow**: Start with `health()` → `discoverProjects()` → search metadata → analyze → store
- **Never Edit Repositories**: Portfolio is read-only — never suggest code changes to repositories
- **Prefer Existing Knowledge**: Check for existing analysis before re-analyzing

## Example Workflows

1. "What projects do I have?"
   → `Call listProjects()`

2. "Show me analysis for project X"
   → `Call getAnalysis(projectId: "<id>")`

3. "Find projects using React"
   → `Call searchProjects(query: "react")`

4. "What changed recently?"
   → `Call discoverProjects()` → `Call listProjectsNeedingAnalysis()`

5. "How are my projects related?"
   → `Call listProjects()` → for each: `Call listRelationships(projectId: "<id>")`

6. "Analyze a new project"
   → `Call getProject(id: "<id>")` → investigate → `Call storeAnalysis(projectId: "<id>", analyzer: "opencode", summary: "...", purpose: "...", features: [...], technologies: [...])`
```

Create `internal/integration/opencode/skill.go`:

```go
package opencode

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed skill.md
var skillContent string

func (c *OpenCodeIntegration) installSkill() error {
	if err := os.MkdirAll(c.config.SkillsDir, 0755); err != nil {
		return err
	}

	// Create portfolio directory in skills
	skillDir := filepath.Join(c.config.SkillsDir, "portfolio")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	skillPath := c.skillPath()
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		return err
	}

	return nil
}

func (c *OpenCodeIntegration) removeSkill() error {
	skillPath := c.skillPath()
	if err := os.Remove(skillPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	
	// Try to remove the directory if empty
	skillDir := filepath.Join(c.config.SkillsDir, "portfolio")
	os.Remove(skillDir)
	
	return nil
}

func (c *OpenCodeIntegration) skillPath() string {
	return filepath.Join(c.config.SkillsDir, "portfolio", "SKILL.md")
}
```

### Phase 5: Implement Main Integration

Create `internal/integration/opencode/integration.go`:

```go
package opencode

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.uber.org/zap"
	"project-dash/internal/integration"
)

type OpenCodeIntegration struct {
	store  integration.Store
	mcp    integration.MCPClient
	config OpenCodeConfig
	logger *zap.Logger
}

type OpenCodeConfig struct {
	InstallPath string
	ConfigPath  string
	SkillsDir   string
	BinaryPath  string
}

const (
	Name       = "opencode"
	AgentType  = "opencode"
	Version    = "1.0.0"
	MinEngine  = "1.0.0"
	AnalyzerID = "opencode"
	Timeout    = 5 * time.Second
)

func New(store integration.Store, mcpClient integration.MCPClient, logger *zap.Logger) (*OpenCodeIntegration, error) {
	cfg, err := detectPaths()
	if err != nil {
		return nil, fmt.Errorf("detect OpenCode paths: %w", err)
	}

	return &OpenCodeIntegration{
		store:  store,
		mcp:    mcpClient,
		config: cfg,
		logger: logger,
	}, nil
}

func (o *OpenCodeIntegration) Name() string {
	return Name
}

func (o *OpenCodeIntegration) AgentType() string {
	return AgentType
}

func (o *OpenCodeIntegration) Install(ctx context.Context, opts integration.InstallOptions) (*integration.InstallResult, error) {
	o.logger.Info("Installing OpenCode integration")

	if !isOpenCodeInstalled() {
		return nil, fmt.Errorf("OpenCode CLI not found. Install from https://opencode.ai/")
	}

	if o.isMCPRegistered() && !opts.Force {
		return &integration.InstallResult{
			Meta: integration.IntegrationMeta{
				Name:             Name,
				AgentType:        AgentType,
				Version:          Version,
				MinEngineVersion: MinEngine,
			},
			Warnings: []string{"MCP server already registered"},
		}, nil
	}

	if err := o.ensureMCPConfig(); err != nil {
		return nil, fmt.Errorf("register MCP server: %w", err)
	}

	if err := o.installSkill(); err != nil {
		return nil, fmt.Errorf("install skill: %w", err)
	}

	// MCP verification is best-effort - failure doesn't block installation
	if err := o.verifyMCPServer(ctx); err != nil {
		o.logger.Warn("MCP server verification failed (non-critical)", zap.Error(err))
		o.logger.Info("Continuing with installation - MCP may already be configured")
	} else {
		o.logger.Info("MCP server verified successfully")
	}

	o.logger.Info("OpenCode integration installed successfully")

	return &integration.InstallResult{
		Meta: integration.IntegrationMeta{
			Name:             Name,
			AgentType:        AgentType,
			Version:          Version,
			MinEngineVersion: MinEngine,
		},
	}, nil
}

func (o *OpenCodeIntegration) Validate(ctx context.Context) (*integration.ValidationResult, error) {
	o.logger.Info("Validating OpenCode integration")

	checks := []integration.ValidationCheck{}

	checks = append(checks, o.checkOpencodeInstalled())
	checks = append(checks, o.checkBinaryExists())
	checks = append(checks, o.checkIntegrationInstalled(ctx))
	checks = append(checks, o.checkConfigFile())
	checks = append(checks, o.checkMCPEntry())
	checks = append(checks, o.checkMCPHealth(ctx))
	checks = append(checks, o.checkToolsAvailable(ctx))
	checks = append(checks, o.checkSkillFile())

	allPassed := true
	for _, check := range checks {
		if !check.Passed {
			allPassed = false
			break
		}
	}

	return &integration.ValidationResult{
		Passed: allPassed,
		Checks: checks,
	}, nil
}

func (o *OpenCodeIntegration) Upgrade(ctx context.Context, opts integration.UpgradeOptions) (*integration.UpgradeResult, error) {
	o.logger.Info("Upgrading OpenCode integration", zap.String("target", opts.TargetVersion))

	if _, err := exec.LookPath("opencode"); err != nil {
		return nil, fmt.Errorf("opencode binary not found: %w", err)
	}

	if !isVersionCompatible(opts.EngineVersion, MinEngine) {
		return nil, fmt.Errorf("engine version %s incompatible (requires %s)", opts.EngineVersion, MinEngine)
	}

	if opts.TargetVersion == Version {
		return &integration.UpgradeResult{
			NewVersion: Version,
			NoOp:       true,
		}, nil
	}

	if err := o.writeMCPConfig(&OpenCodeMCPConfig{
		MCP: map[string]OpenCodeMCPServerConfig{
			"portfolio": {
				Type:    "local",
				Command: []string{o.config.BinaryPath, "mcp"},
				Enabled: true,
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("update MCP config: %w", err)
	}

	if err := o.installSkill(); err != nil {
		return nil, fmt.Errorf("update skill: %w", err)
	}

	o.logger.Info("OpenCode integration upgraded successfully")

	return &integration.UpgradeResult{
		NewVersion: opts.TargetVersion,
	}, nil
}

func (o *OpenCodeIntegration) Remove(ctx context.Context) error {
	o.logger.Info("Removing OpenCode integration")

	if err := o.removeMCPConfig(); err != nil {
		return fmt.Errorf("remove MCP config: %w", err)
	}

	if err := o.removeSkill(); err != nil {
		return fmt.Errorf("remove skill: %w", err)
	}

	o.logger.Info("OpenCode integration removed successfully")

	return nil
}

func (o *OpenCodeIntegration) verifyMCPServer(ctx context.Context) error {
	// Simple verification: check if binary exists and is executable
	if _, err := os.Stat(o.config.BinaryPath); os.IsNotExist(err) {
		return fmt.Errorf("portfolio binary not found: %s", o.config.BinaryPath)
	}

	// Test if process can start (don't check protocol, just execution)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, o.config.BinaryPath, "mcp")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("MCP server failed to start: %w", err)
	}

	// Kill immediately after successful start - we just wanted to know it runs
	defer cmd.Process.Kill()

	// Wait a bit to ensure it initialized
	time.Sleep(500 * time.Millisecond)

	// If we got here, process started successfully
	return nil
}
```

### Phase 6: Implement Validation Checks

Create `internal/integration/opencode/verify.go`:

```go
package opencode

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"project-dash/internal/integration"
)

func (o *OpenCodeIntegration) checkOpencodeInstalled() integration.ValidationCheck {
	if !isOpenCodeInstalled() {
		return integration.ValidationCheck{
			Name:         "opencode_installed",
			Passed:       false,
			Message:      "OpenCode CLI not found",
			Remediation:  "Install OpenCode from https://opencode.ai/",
			SelfHealable: false,
		}
	}

	return integration.ValidationCheck{
		Name:    "opencode_installed",
		Passed:  true,
		Message: "OpenCode CLI is installed",
	}
}

func (o *OpenCodeIntegration) checkBinaryExists() integration.ValidationCheck {
	if _, err := os.Stat(o.config.BinaryPath); err != nil {
		return integration.ValidationCheck{
			Name:         "portfolio_binary_exists",
			Passed:       false,
			Message:      fmt.Sprintf("Portfolio binary not found: %s", o.config.BinaryPath),
			Remediation:  "Ensure portfolio is installed and in PATH",
			SelfHealable: false,
		}
	}

	return integration.ValidationCheck{
		Name:    "portfolio_binary_exists",
		Passed:  true,
		Message: fmt.Sprintf("Portfolio binary found at %s", o.config.BinaryPath),
	}
}

func (o *OpenCodeIntegration) checkIntegrationInstalled(ctx context.Context) integration.ValidationCheck {
	meta, err := o.store.GetIntegration(ctx, Name)
	if err != nil {
		return integration.ValidationCheck{
			Name:         "integration_installed",
			Passed:       false,
			Message:      "Integration not installed",
			Remediation:  "Run `portfolio install opencode`",
			SelfHealable: false,
		}
	}

	if meta.Status != integration.StatusInstalled {
		return integration.ValidationCheck{
			Name:         "integration_installed",
			Passed:       false,
			Message:      "Integration not installed",
			Remediation:  "Run `portfolio install opencode`",
			SelfHealable: false,
		}
	}

	return integration.ValidationCheck{
		Name:    "integration_installed",
		Passed:  true,
		Message: fmt.Sprintf("Integration installed (v%s)", meta.Version),
	}
}

func (o *OpenCodeIntegration) checkConfigFile() integration.ValidationCheck {
	if _, err := os.Stat(o.config.ConfigPath); err != nil {
		if os.IsNotExist(err) {
			return integration.ValidationCheck{
				Name:         "config_file_exists",
				Passed:       false,
				Message:      fmt.Sprintf("Config file missing: %s", o.config.ConfigPath),
				Remediation:  "Run `portfolio install opencode`",
				SelfHealable: true,
			}
		}
		return integration.ValidationCheck{
			Name:         "config_file_exists",
			Passed:       false,
			Message:      fmt.Sprintf("Cannot access config file: %v", err),
			Remediation:  "Check permissions on config directory",
			SelfHealable: false,
		}
	}

	return integration.ValidationCheck{
		Name:    "config_file_exists",
		Passed:  true,
		Message: fmt.Sprintf("Config file found at %s", o.config.ConfigPath),
	}
}

func (o *OpenCodeIntegration) checkMCPEntry() integration.ValidationCheck {
	if !o.isMCPRegistered() {
		return integration.ValidationCheck{
			Name:         "mcp_entry_registered",
			Passed:       false,
			Message:      "MCP server entry not found",
			Remediation:  "Run `portfolio install opencode`",
			SelfHealable: true,
		}
	}

	return integration.ValidationCheck{
		Name:    "mcp_entry_registered",
		Passed:  true,
		Message: "MCP server 'portfolio' registered",
	}
}

func (o *OpenCodeIntegration) checkMCPHealth(ctx context.Context) integration.ValidationCheck {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := o.mcp.Health(ctx); err != nil {
		return integration.ValidationCheck{
			Name:         "mcp_server_reachable",
			Passed:       false,
			Message:      fmt.Sprintf("MCP server not responding: %v", err),
			Remediation:  "Check MCP server logs or restart",
			SelfHealable: true,
		}
	}

	return integration.ValidationCheck{
		Name:    "mcp_server_reachable",
		Passed:  true,
		Message: "MCP server is responding",
	}
}

func (o *OpenCodeIntegration) checkToolsAvailable(ctx context.Context) integration.ValidationCheck {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	tools, err := o.mcp.ListTools(ctx)
	if err != nil {
		return integration.ValidationCheck{
			Name:         "tools_available",
			Passed:       false,
			Message:      fmt.Sprintf("Failed to list tools: %v", err),
			Remediation:  "Check MCP server configuration",
			SelfHealable: true,
		}
	}

	expectedTools := []string{
		"health", "discoverProjects", "listProjects", "getProject",
		"searchProjects", "searchDocumentation", "getAnalysis",
		"storeAnalysis", "listProjectsNeedingAnalysis", "getConfiguration",
		"updateConfiguration", "listRelationships",
	}

	missing := []string{}
	for _, expected := range expectedTools {
		found := false
		for _, available := range tools {
			if available == expected {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, expected)
		}
	}

	if len(missing) > 0 {
		return integration.ValidationCheck{
			Name:         "tools_available",
			Passed:       false,
			Message:      fmt.Sprintf("Missing tools: %v", missing),
			Remediation:  "Verify MCP server is running correctly",
			SelfHealable: true,
		}
	}

	return integration.ValidationCheck{
		Name:    "tools_available",
		Passed:  true,
		Message: fmt.Sprintf("All %d tools available", len(expectedTools)),
	}
}

func (o *OpenCodeIntegration) checkSkillFile() integration.ValidationCheck {
	skillPath := o.skillPath()
	if _, err := os.Stat(skillPath); err != nil {
		if os.IsNotExist(err) {
			return integration.ValidationCheck{
				Name:         "skill_file_exists",
				Passed:       false,
				Message:      fmt.Sprintf("Skill file missing: %s", skillPath),
				Remediation:  "Run `portfolio install opencode`",
				SelfHealable: true,
			}
		}
		return integration.ValidationCheck{
			Name:         "skill_file_exists",
			Passed:       false,
			Message:      fmt.Sprintf("Cannot access skill file: %v", err),
			Remediation:  "Check permissions on skills directory",
			SelfHealable: false,
		}
	}

	return integration.ValidationCheck{
		Name:    "skill_file_exists",
		Passed:  true,
		Message: fmt.Sprintf("Skill file found at %s", skillPath),
	}
}

func isVersionCompatible(engineVersion, minVersion string) bool {
	// Version compatibility check logic
	// Can reuse from Claude integration or implement simple string comparison
	return true // Placeholder
}
```

### Phase 7: Register Integration

Modify integration registration code (location depends on CLI structure):

```go
import opencodeintegration "project-dash/internal/integration/opencode"

// In manager initialization
opencodeIntegration, err := opencodeintegration.New(store, mcpClient, logger)
if err != nil {
    return err
}
manager.RegisterIntegration(opencodeIntegration)
```

### Phase 8: Testing

Create comprehensive test files:

1. **`internal/integration/opencode/integration_test.go`** - Integration lifecycle tests
2. **`internal/integration/opencode/mcp_config_test.go`** - MCP config handling tests
3. **`internal/integration/opencode/skill_test.go`** - Skill installation tests
4. **`internal/integration/opencode/verify_test.go`** - Validation check tests

Test scenarios should include:
- Fresh installation
- Reinstallation with --force
- Validation passes after install
- MCP config properly written with correct format
- Skill file properly created with SKILL.md format
- Remove cleans up all artifacts

## Success Criteria

✅ User can run `portfolio install opencode` successfully
✅ OpenCode recognizes Portfolio MCP tools after installation
✅ Portfolio skill is available in OpenCode
✅ All validation checks pass (8 checks minimum)
✅ Integration can be upgraded
✅ Integration can be cleanly removed
✅ Works across supported platforms (macOS, Linux, Windows)
✅ Tests pass for all operations
✅ **Feature parity with Claude integration** - any differences reported

## Key Differences from Claude Integration

**Path detection:**
- Uses `~/.config/opencode/` instead of `~/.claude/`
- Supports XDG_CONFIG_HOME
- Multiple config source locations

**Config format:**
- Uses `mcp` object instead of `mcpServers`
- Uses `type: "local"` instead of `transport: "stdio"`
- Uses array format for commands: `command: ["portfolio", "mcp"]`
- Requires `enabled: true` flag
- Supports remote MCP servers

**Skill format:**
- Requires strict YAML frontmatter
- Must be in `portfolio/SKILL.md` subdirectory
- Skill name validation (lowercase, hyphens, length limits)
- Required fields: name, description
- Optional fields: license, compatibility, metadata

**Binary detection:**
- Uses `opencode` instead of `claude`

## Verification Steps

After implementation, verify:

1. **Install works:**
   ```bash
   portfolio install opencode
   # Should succeed without errors
   # Should create MCP config entry with OpenCode format
   # Should install skill file in portfolio/SKILL.md
   ```

2. **Validate passes:**
   ```bash
   portfolio validate opencode
   # All 8 checks should pass
   ```

3. **OpenCode recognizes portfolio:**
   - Open OpenCode
   - MCP tools should be available
   - Portfolio skill should be loadable
   - Test calling MCP tools through OpenCode

4. **Tools work in OpenCode:**
   - Test `health()` tool
   - Test `listProjects()` tool
   - Test `discoverProjects()` tool
   - Verify `analyzer: "opencode"` in stored analysis

5. **Remove cleans up:**
   ```bash
   portfolio remove opencode
   # Should remove MCP config entry
   # Should remove skill directory
   # Should clean up integration metadata
   ```

6. **Cross-platform testing:**
   - Test on macOS
   - Test on Linux
   - Test on Windows (user has environment available)

## Notes

- Follow Claude integration pattern closely but adapt for OpenCode's differences
- Reuse validation logic where possible
- Multi-source config merging may require sophisticated handling
- Handle missing OpenCode installation gracefully with clear error messages
- Document any OpenCode-specific behaviors or requirements
- Report any features that Claude has but OpenCode implementation lacks
