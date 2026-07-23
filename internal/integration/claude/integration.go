package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
	"project-dash/internal/integration"
)

type ClaudeCodeIntegration struct {
	store  integration.Store
	mcp    integration.MCPClient
	config ClaudeConfig
	logger *zap.Logger
}

type ClaudeConfig struct {
	InstallPath string
	ConfigPath  string
	SkillsDir   string
	BinaryPath  string
}

const (
	Name       = "claude"
	AgentType  = "claude-code"
	Version    = "1.0.0"
	MinEngine  = "1.0.0"
	AnalyzerID = "claude-code"
	Timeout    = 5 * time.Second
)

func New(store integration.Store, mcpClient integration.MCPClient, logger *zap.Logger) (*ClaudeCodeIntegration, error) {
	cfg, err := detectPaths()
	if err != nil {
		return nil, fmt.Errorf("detect Claude Code paths: %w", err)
	}

	return &ClaudeCodeIntegration{
		store:  store,
		mcp:    mcpClient,
		config: cfg,
		logger: logger,
	}, nil
}

func (c *ClaudeCodeIntegration) Name() string {
	return Name
}

func (c *ClaudeCodeIntegration) AgentType() string {
	return AgentType
}

func (c *ClaudeCodeIntegration) Install(ctx context.Context, opts integration.InstallOptions) (*integration.InstallResult, error) {
	c.logger.Info("Installing Claude Code integration")

	if !isClaudeInstalled() {
		return nil, fmt.Errorf("Claude Code CLI not found. Install from https://docs.anthropic.com/claude-code")
	}

	if c.isMCPRegistered() && !opts.Force {
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

	if err := c.ensureMCPConfig(); err != nil {
		return nil, fmt.Errorf("register MCP server: %w", err)
	}

	if err := c.installSkill(); err != nil {
		return nil, fmt.Errorf("install skill: %w", err)
	}

	if err := c.verifyMCPServer(ctx); err != nil {
		c.logger.Warn("MCP server verification failed", zap.Error(err))
		return nil, fmt.Errorf("verify MCP server: %w", err)
	}

	c.logger.Info("Claude Code integration installed successfully")

	return &integration.InstallResult{
		Meta: integration.IntegrationMeta{
			Name:             Name,
			AgentType:        AgentType,
			Version:          Version,
			MinEngineVersion: MinEngine,
		},
	}, nil
}

func (c *ClaudeCodeIntegration) Validate(ctx context.Context) (*integration.ValidationResult, error) {
	c.logger.Info("Validating Claude Code integration")

	checks := []integration.ValidationCheck{}

	checks = append(checks, c.checkClaudeInstalled())
	checks = append(checks, c.checkBinaryExists())
	checks = append(checks, c.checkIntegrationInstalled(ctx))
	checks = append(checks, c.checkConfigFile())
	checks = append(checks, c.checkMCPEntry())
	checks = append(checks, c.checkMCPHealth(ctx))
	checks = append(checks, c.checkToolsAvailable(ctx))
	checks = append(checks, c.checkSkillFile())

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

func (c *ClaudeCodeIntegration) Upgrade(ctx context.Context, opts integration.UpgradeOptions) (*integration.UpgradeResult, error) {
	c.logger.Info("Upgrading Claude Code integration", zap.String("target", opts.TargetVersion))

	if _, err := exec.LookPath("claude"); err != nil {
		return nil, fmt.Errorf("claude binary not found: %w", err)
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

	if err := c.writeMCPConfig(&MCPConfig{
		MCPServers: map[string]MCPServerConfig{
			"portfolio": {
				Command:   c.config.BinaryPath,
				Args:      []string{"mcp"},
				Transport: "stdio",
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("update MCP config: %w", err)
	}

	if err := c.installSkill(); err != nil {
		return nil, fmt.Errorf("update skill: %w", err)
	}

	c.logger.Info("Claude Code integration upgraded successfully")

	return &integration.UpgradeResult{
		NewVersion: opts.TargetVersion,
	}, nil
}

func (c *ClaudeCodeIntegration) Remove(ctx context.Context) error {
	c.logger.Info("Removing Claude Code integration")

	if err := c.removeMCPConfig(); err != nil {
		return fmt.Errorf("remove MCP config: %w", err)
	}

	if err := c.removeSkill(); err != nil {
		return fmt.Errorf("remove skill: %w", err)
	}

	c.logger.Info("Claude Code integration removed successfully")

	return nil
}

func (c *ClaudeCodeIntegration) verifyMCPServer(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.config.BinaryPath, "mcp")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start MCP server: %w", err)
	}
	defer cmd.Process.Kill()

	healthRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "health",
			"arguments": map[string]interface{}{},
		},
	}

	requestData, err := json.Marshal(healthRequest)
	if err != nil {
		return fmt.Errorf("marshal health request: %w", err)
	}

	if _, err := stdin.Write(requestData); err != nil {
		return fmt.Errorf("send health request: %w", err)
	}

	response := make([]byte, 1024)
	n, err := stdout.Read(response)
	if err != nil {
		return fmt.Errorf("read health response: %w", err)
	}

	var healthResponse map[string]interface{}
	if err := json.Unmarshal(response[:n], &healthResponse); err != nil {
		return fmt.Errorf("parse health response: %w", err)
	}

	if result, ok := healthResponse["result"].(map[string]interface{}); ok {
		if status, ok := result["status"].(string); ok && status == "ok" {
			return nil
		}
	}

	return fmt.Errorf("health check failed")
}

func (c *ClaudeCodeIntegration) checkIntegrationInstalled(ctx context.Context) integration.ValidationCheck {
	meta, err := c.store.GetIntegration(ctx, Name)
	if err != nil {
		return integration.ValidationCheck{
			Name:        "integration_installed",
			Passed:      false,
			Message:     "Integration not installed",
			Remediation: "Run `portfolio install claude`",
		}
	}

	if meta.Status != integration.StatusInstalled {
		return integration.ValidationCheck{
			Name:        "integration_installed",
			Passed:      false,
			Message:     "Integration not installed",
			Remediation: "Run `portfolio install claude`",
		}
	}

	return integration.ValidationCheck{
		Name:    "integration_installed",
		Passed:  true,
		Message: fmt.Sprintf("Integration installed (v%s)", meta.Version),
	}
}

func (c *ClaudeCodeIntegration) checkConfigFile() integration.ValidationCheck {
	if _, err := os.Stat(c.config.ConfigPath); err != nil {
		if os.IsNotExist(err) {
			return integration.ValidationCheck{
				Name:        "config_file_exists",
				Passed:      false,
				Message:     fmt.Sprintf("Config file missing: %s", c.config.ConfigPath),
				Remediation: "Run `portfolio install claude`",
			}
		}
		return integration.ValidationCheck{
			Name:        "config_file_exists",
			Passed:      false,
			Message:     fmt.Sprintf("Cannot access config file: %v", err),
			Remediation: "Check permissions on config directory",
		}
	}

	return integration.ValidationCheck{
		Name:    "config_file_exists",
		Passed:  true,
		Message: fmt.Sprintf("Config file found at %s", c.config.ConfigPath),
	}
}

func (c *ClaudeCodeIntegration) checkMCPEntry() integration.ValidationCheck {
	if !c.isMCPRegistered() {
		return integration.ValidationCheck{
			Name:        "mcp_entry_registered",
			Passed:      false,
			Message:     "MCP server entry not found",
			Remediation: "Run `portfolio install claude`",
		}
	}

	return integration.ValidationCheck{
		Name:    "mcp_entry_registered",
		Passed:  true,
		Message: "MCP server 'portfolio' registered",
	}
}

func (c *ClaudeCodeIntegration) checkMCPHealth(ctx context.Context) integration.ValidationCheck {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := c.mcp.Health(ctx); err != nil {
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

func (c *ClaudeCodeIntegration) checkSkillFile() integration.ValidationCheck {
	skillPath := c.skillPath()
	if _, err := os.Stat(skillPath); err != nil {
		if os.IsNotExist(err) {
			return integration.ValidationCheck{
				Name:        "skill_file_exists",
				Passed:      false,
				Message:     fmt.Sprintf("Skill file missing: %s", skillPath),
				Remediation: "Run `portfolio install claude`",
			}
		}
		return integration.ValidationCheck{
			Name:        "skill_file_exists",
			Passed:      false,
			Message:     fmt.Sprintf("Cannot access skill file: %v", err),
			Remediation: "Check permissions on skills directory",
		}
	}

	return integration.ValidationCheck{
		Name:    "skill_file_exists",
		Passed:  true,
		Message: fmt.Sprintf("Skill file found at %s", skillPath),
	}
}

func isVersionCompatible(engineVersion, minVersion string) bool {
	if engineVersion == "" || minVersion == "" {
		return false
	}

	engineSemver, err := semver.NewVersion(engineVersion)
	if err != nil {
		return false
	}

	minSemver, err := semver.NewVersion(minVersion)
	if err != nil {
		return false
	}

	return engineSemver.Compare(minSemver) >= 0
}
