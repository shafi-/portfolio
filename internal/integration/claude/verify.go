package claude

import (
	"context"
	"fmt"
	"os"

	"project-dash/internal/integration"
)

func (c *ClaudeCodeIntegration) checkToolsAvailable(ctx context.Context) integration.ValidationCheck {
	tools, err := c.mcp.ListTools(ctx)
	if err != nil {
		return integration.ValidationCheck{
			Name:         "mcp_tools_available",
			Passed:       false,
			Message:      fmt.Sprintf("Failed to list MCP tools: %v", err),
			Remediation:  "Check MCP server logs or restart",
			SelfHealable: true,
		}
	}

	requiredTools := []string{
		"health",
		"discoverProjects",
		"listProjects",
		"getProject",
		"searchProjects",
		"searchDocumentation",
		"getAnalysis",
		"storeAnalysis",
		"listProjectsNeedingAnalysis",
		"getConfiguration",
		"updateConfiguration",
		"listRelationships",
	}

	missingTools := []string{}
	for _, required := range requiredTools {
		found := false
		for _, available := range tools {
			if available == required {
				found = true
				break
			}
		}
		if !found {
			missingTools = append(missingTools, required)
		}
	}

	if len(missingTools) > 0 {
		return integration.ValidationCheck{
			Name:        "mcp_tools_available",
			Passed:      false,
			Message:     fmt.Sprintf("Missing MCP tools: %v", missingTools),
			Remediation: "Reinstall integration: `portfolio install claude --force`",
		}
	}

	return integration.ValidationCheck{
		Name:    "mcp_tools_available",
		Passed:  true,
		Message: fmt.Sprintf("All %d required MCP tools available", len(requiredTools)),
	}
}

func (c *ClaudeCodeIntegration) checkClaudeInstalled() integration.ValidationCheck {
	if !isClaudeInstalled() {
		return integration.ValidationCheck{
			Name:        "claude_installed",
			Passed:      false,
			Message:     "Claude Code CLI not found",
			Remediation: "Install Claude Code from https://docs.anthropic.com/claude-code",
		}
	}

	return integration.ValidationCheck{
		Name:    "claude_installed",
		Passed:  true,
		Message: "Claude Code CLI is installed",
	}
}

func (c *ClaudeCodeIntegration) checkBinaryExists() integration.ValidationCheck {
	if _, err := os.Stat(c.config.BinaryPath); err != nil {
		if os.IsNotExist(err) {
			return integration.ValidationCheck{
				Name:        "portfolio_binary_exists",
				Passed:      false,
				Message:     fmt.Sprintf("Portfolio binary not found: %s", c.config.BinaryPath),
				Remediation: "Reinstall Portfolio Engine",
			}
		}
		return integration.ValidationCheck{
			Name:        "portfolio_binary_exists",
			Passed:      false,
			Message:     fmt.Sprintf("Cannot access Portfolio binary: %v", err),
			Remediation: "Check permissions on Portfolio binary",
		}
	}

	return integration.ValidationCheck{
		Name:    "portfolio_binary_exists",
		Passed:  true,
		Message: fmt.Sprintf("Portfolio binary found at %s", c.config.BinaryPath),
	}
}
