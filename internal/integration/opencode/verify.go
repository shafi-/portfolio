package opencode

import (
	"context"
	"fmt"
	"os"

	"project-dash/internal/integration"
)

func (o *OpenCodeIntegration) checkToolsAvailable(ctx context.Context) integration.ValidationCheck {
	tools, err := o.mcp.ListTools(ctx)
	if err != nil {
		return integration.ValidationCheck{
			Name:         "mcp_tools_available",
			Passed:       false,
			Message:      fmt.Sprintf("Failed to list MCP tools: %v", err),
			Remediation:  "Check MCP server logs or restart",
			SelfHealable: true,
		}
	}

	// Same required-tool set as the Claude Code integration for parity.
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
			Remediation: "Reinstall integration: `portfolio install opencode --force`",
		}
	}

	return integration.ValidationCheck{
		Name:    "mcp_tools_available",
		Passed:  true,
		Message: fmt.Sprintf("All %d required MCP tools available", len(requiredTools)),
	}
}

func (o *OpenCodeIntegration) checkOpenCodeInstalled() integration.ValidationCheck {
	if !isOpenCodeInstalled() {
		return integration.ValidationCheck{
			Name:        "opencode_installed",
			Passed:      false,
			Message:     "OpenCode not found on PATH",
			Remediation: "Install OpenCode from https://opencode.ai/docs",
		}
	}

	return integration.ValidationCheck{
		Name:    "opencode_installed",
		Passed:  true,
		Message: "OpenCode is installed",
	}
}

func (o *OpenCodeIntegration) checkBinaryExists() integration.ValidationCheck {
	if _, err := os.Stat(o.config.BinaryPath); err != nil {
		if os.IsNotExist(err) {
			return integration.ValidationCheck{
				Name:        "portfolio_binary_exists",
				Passed:      false,
				Message:     fmt.Sprintf("Portfolio binary not found: %s", o.config.BinaryPath),
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
		Message: fmt.Sprintf("Portfolio binary found at %s", o.config.BinaryPath),
	}
}
