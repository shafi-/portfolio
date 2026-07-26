package opencode

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
	"project-dash/internal/integration"
)

// OpenCodeIntegration installs Portfolio into OpenCode. Unlike Claude Code
// (which ships a local-stdio MCP CLI), OpenCode's official, schema-documented
// method for registering a local MCP server is its config file
// (~/.config/opencode/opencode.json, $schema https://opencode.ai/config.json).
// Per ADR-018, writing that officially-documented file counts as an official
// method, not fragile config editing.
type OpenCodeIntegration struct {
	store  integration.Store
	mcp    integration.MCPClient
	config OpenCodeConfig
	logger *zap.Logger
}

type OpenCodeConfig struct {
	InstallPath string
	ConfigPath  string // opencode.json
	SkillsDir   string // ~/.config/opencode/skills
	BinaryPath  string
}

const (
	Name       = "opencode"
	AgentType  = "opencode"
	Version    = "1.0.0"
	MinEngine  = "0.1.0"
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

func (o *OpenCodeIntegration) Name() string      { return Name }
func (o *OpenCodeIntegration) AgentType() string { return AgentType }

func (o *OpenCodeIntegration) Install(ctx context.Context, opts integration.InstallOptions) (*integration.InstallResult, error) {
	o.logger.Info("Installing OpenCode integration")

	if !isOpenCodeInstalled() {
		return nil, fmt.Errorf("OpenCode not found. Install from https://opencode.ai/docs")
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

	checks = append(checks, o.checkOpenCodeInstalled())
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

	if !isOpenCodeInstalled() {
		return nil, fmt.Errorf("OpenCode not found: %w", exec.ErrNotFound)
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

	// Re-write the config (idempotent merge) and refresh the skill.
	if err := o.ensureMCPConfig(); err != nil {
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

func (o *OpenCodeIntegration) checkIntegrationInstalled(ctx context.Context) integration.ValidationCheck {
	meta, err := o.store.GetIntegration(ctx, Name)
	if err != nil {
		return integration.ValidationCheck{
			Name:        "integration_installed",
			Passed:      false,
			Message:     "Integration not installed",
			Remediation: "Run `portfolio install opencode`",
		}
	}

	if meta.Status != integration.StatusInstalled {
		return integration.ValidationCheck{
			Name:        "integration_installed",
			Passed:      false,
			Message:     "Integration not installed",
			Remediation: "Run `portfolio install opencode`",
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
				Name:        "config_file_exists",
				Passed:      false,
				Message:     fmt.Sprintf("Config file missing: %s", o.config.ConfigPath),
				Remediation: "Run `portfolio install opencode`",
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
		Message: fmt.Sprintf("Config file found at %s", o.config.ConfigPath),
	}
}

func (o *OpenCodeIntegration) checkMCPEntry() integration.ValidationCheck {
	if !o.isMCPRegistered() {
		return integration.ValidationCheck{
			Name:        "mcp_entry_registered",
			Passed:      false,
			Message:     "MCP server entry not found in opencode.json",
			Remediation: "Run `portfolio install opencode`",
		}
	}

	return integration.ValidationCheck{
		Name:    "mcp_entry_registered",
		Passed:  true,
		Message: "MCP server 'portfolio' registered in opencode.json",
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

func (o *OpenCodeIntegration) checkSkillFile() integration.ValidationCheck {
	skillPath := o.skillPath()
	if _, err := os.Stat(skillPath); err != nil {
		if os.IsNotExist(err) {
			return integration.ValidationCheck{
				Name:        "skill_file_exists",
				Passed:      false,
				Message:     fmt.Sprintf("Skill file missing: %s", skillPath),
				Remediation: "Run `portfolio install opencode`",
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
