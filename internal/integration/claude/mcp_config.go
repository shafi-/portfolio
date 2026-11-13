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
	// Always remove first to handle stale entries with wrong paths.
	// The CLI refuses to act when the server exists in multiple scopes, so each
	// scope must be removed explicitly; a missing entry in either scope is fine.
	if err := c.removeMCPConfig(); err != nil {
		return fmt.Errorf("clear existing MCP registrations: %w", err)
	}

	// Register at user scope so the integration is available in every project
	// instead of only the directory the installer ran from.
	args := []string{"mcp", "add", "-s", "user", "portfolio", c.config.BinaryPath, "mcp"}

	cmd := exec.Command("claude", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("claude mcp add failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// removeMCPConfig unregisters the Portfolio MCP server from every scope the CLI
// supports. Entries may exist in local and/or user scope; removing per scope is
// required because a bare `claude mcp remove` fails when the entry is ambiguous.
func (c *ClaudeCodeIntegration) removeMCPConfig() error {
	var failures []string
	for _, scope := range []string{"local", "user"} {
		if err := c.removeMCPScope(scope); err != nil {
			failures = append(failures, fmt.Sprintf("scope %s: %v", scope, err))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return nil
}

// removeMCPScope unregisters the Portfolio MCP server from a single scope.
// A missing entry is treated as success so idempotent removal works.
func (c *ClaudeCodeIntegration) removeMCPScope(scope string) error {
	// Use official Claude Code CLI command: claude mcp remove portfolio -s <scope>
	args := []string{"mcp", "remove", "portfolio", "-s", scope}

	cmd := exec.Command("claude", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Don't fail if MCP server is already removed from this scope.
		// The CLI reports a missing entry as 'No MCP server named "x" in scope',
		// so match that wording case-insensitively alongside generic phrases.
		out := string(output)
		lower := strings.ToLower(out)
		if strings.Contains(lower, "not found") ||
			strings.Contains(lower, "does not exist") ||
			strings.Contains(lower, "no mcp server named") {
			return nil
		}
		return fmt.Errorf("claude mcp remove failed: %w\nOutput: %s", err, out)
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
