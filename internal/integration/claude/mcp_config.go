package claude

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Legacy config structures kept for reference only - not used in implementation
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers,omitempty"`
}

type MCPServerConfig struct {
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
	Transport string   `json:"transport,omitempty"`
}

// ensureMCPConfig registers the Portfolio MCP server using the official Claude Code CLI
func (c *ClaudeCodeIntegration) ensureMCPConfig() error {
	// Always remove first to handle stale entries with wrong paths
	_ = c.removeMCPConfig()

	// Use official Claude Code CLI command: claude mcp add portfolio /path/to/portfolio mcp
	args := []string{"mcp", "add", "portfolio", c.config.BinaryPath, "mcp"}

	cmd := exec.Command("claude", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("claude mcp add failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// removeMCPConfig unregisters the Portfolio MCP server using the official Claude Code CLI
func (c *ClaudeCodeIntegration) removeMCPConfig() error {
	// Use official Claude Code CLI command: claude mcp remove portfolio
	args := []string{"mcp", "remove", "portfolio"}

	cmd := exec.Command("claude", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Don't fail if MCP server is already removed
		if strings.Contains(string(output), "not found") || strings.Contains(string(output), "does not exist") {
			return nil
		}
		return fmt.Errorf("claude mcp remove failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// isMCPRegistered checks if the Portfolio MCP server is registered using the official Claude Code CLI
func (c *ClaudeCodeIntegration) isMCPRegistered() bool {
	// Use official Claude Code CLI command: claude mcp get portfolio
	args := []string{"mcp", "get", "portfolio"}

	cmd := exec.Command("claude", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// MCP server not found or other error
		return false
	}

	// Check if output contains portfolio binary path
	output := stdout.String()
	if strings.Contains(output, c.config.BinaryPath) {
		return true
	}

	// Try to parse JSON output to verify registration
	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err == nil {
		// Successfully parsed - server is registered
		return true
	}

	return false
}
