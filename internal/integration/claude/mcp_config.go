package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers,omitempty"`
}

type MCPServerConfig struct {
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
	Transport string   `json:"transport,omitempty"`
}

func (c *ClaudeCodeIntegration) ensureMCPConfig() error {
	config, err := c.readMCPConfig()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read MCP config: %w", err)
	}

	if config == nil {
		config = &MCPConfig{}
	}

	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServerConfig)
	}

	config.MCPServers["portfolio"] = MCPServerConfig{
		Command:   c.config.BinaryPath,
		Args:      []string{"mcp"},
		Transport: "stdio",
	}

	if err := c.writeMCPConfig(config); err != nil {
		return fmt.Errorf("write MCP config: %w", err)
	}

	return nil
}

func (c *ClaudeCodeIntegration) removeMCPConfig() error {
	config, err := c.readMCPConfig()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read MCP config: %w", err)
	}

	if config.MCPServers != nil {
		delete(config.MCPServers, "portfolio")
	}

	if err := c.writeMCPConfig(config); err != nil {
		return fmt.Errorf("write MCP config: %w", err)
	}

	return nil
}

func (c *ClaudeCodeIntegration) readMCPConfig() (*MCPConfig, error) {
	data, err := os.ReadFile(c.config.ConfigPath)
	if err != nil {
		return nil, err
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse MCP config: %w", err)
	}

	return &config, nil
}

func (c *ClaudeCodeIntegration) writeMCPConfig(config *MCPConfig) error {
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

func (c *ClaudeCodeIntegration) isMCPRegistered() bool {
	config, err := c.readMCPConfig()
	if err != nil {
		return false
	}

	if config.MCPServers == nil {
		return false
	}

	server, exists := config.MCPServers["portfolio"]
	if !exists {
		return false
	}

	if server.Command != c.config.BinaryPath {
		return false
	}

	if server.Transport != "stdio" {
		return false
	}

	return true
}
