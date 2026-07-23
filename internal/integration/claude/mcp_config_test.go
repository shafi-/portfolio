package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMCPConfigOperations(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	binaryPath := "/usr/local/bin/portfolio"

	integration := &ClaudeCodeIntegration{
		config: ClaudeConfig{
			ConfigPath: configPath,
			BinaryPath: binaryPath,
		},
	}

	t.Run("write to empty config", func(t *testing.T) {
		err := integration.ensureMCPConfig()
		if err != nil {
			t.Fatalf("ensureMCPConfig failed: %v", err)
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read config failed: %v", err)
		}

		var config MCPConfig
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("parse config failed: %v", err)
		}

		if config.MCPServers == nil {
			t.Fatal("mcpServers not created")
		}

		server, exists := config.MCPServers["portfolio"]
		if !exists {
			t.Fatal("portfolio server not registered")
		}

		if server.Command != binaryPath {
			t.Errorf("expected command %s, got %s", binaryPath, server.Command)
		}

		if server.Transport != "stdio" {
			t.Errorf("expected transport stdio, got %s", server.Transport)
		}
	})

	t.Run("merge with existing entries", func(t *testing.T) {
		existingConfig := MCPConfig{
			MCPServers: map[string]MCPServerConfig{
				"other": {
					Command:   "/path/to/other",
					Transport: "stdio",
				},
			},
		}

		data, _ := json.MarshalIndent(existingConfig, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write existing config failed: %v", err)
		}

		err := integration.ensureMCPConfig()
		if err != nil {
			t.Fatalf("ensureMCPConfig failed: %v", err)
		}

		data, err = os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read config failed: %v", err)
		}

		var config MCPConfig
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("parse config failed: %v", err)
		}

		if len(config.MCPServers) != 2 {
			t.Errorf("expected 2 servers, got %d", len(config.MCPServers))
		}

		if _, exists := config.MCPServers["other"]; !exists {
			t.Fatal("other server was removed")
		}

		if _, exists := config.MCPServers["portfolio"]; !exists {
			t.Fatal("portfolio server not registered")
		}
	})

	t.Run("remove MCP config", func(t *testing.T) {
		err := integration.removeMCPConfig()
		if err != nil {
			t.Fatalf("removeMCPConfig failed: %v", err)
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read config failed: %v", err)
		}

		var config MCPConfig
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("parse config failed: %v", err)
		}

		if _, exists := config.MCPServers["portfolio"]; exists {
			t.Fatal("portfolio server still registered after removal")
		}

		if _, exists := config.MCPServers["other"]; !exists {
			t.Fatal("other server was removed during portfolio removal")
		}
	})

	t.Run("idempotent re-write", func(t *testing.T) {
		err := integration.ensureMCPConfig()
		if err != nil {
			t.Fatalf("first ensureMCPConfig failed: %v", err)
		}

		firstData, _ := os.ReadFile(configPath)

		err = integration.ensureMCPConfig()
		if err != nil {
			t.Fatalf("second ensureMCPConfig failed: %v", err)
		}

		secondData, _ := os.ReadFile(configPath)

		if string(firstData) != string(secondData) {
			t.Error("config changed on idempotent re-write")
		}
	})
}

func TestIsMCPRegistered(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	binaryPath := "/usr/local/bin/portfolio"

	integration := &ClaudeCodeIntegration{
		config: ClaudeConfig{
			ConfigPath: configPath,
			BinaryPath: binaryPath,
		},
	}

	t.Run("not registered when config missing", func(t *testing.T) {
		if integration.isMCPRegistered() {
			t.Error("should not be registered when config missing")
		}
	})

	t.Run("not registered when entry missing", func(t *testing.T) {
		config := MCPConfig{MCPServers: make(map[string]MCPServerConfig)}
		data, _ := json.MarshalIndent(config, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write config failed: %v", err)
		}

		if integration.isMCPRegistered() {
			t.Error("should not be registered when entry missing")
		}
	})

	t.Run("registered when entry matches", func(t *testing.T) {
		config := MCPConfig{
			MCPServers: map[string]MCPServerConfig{
				"portfolio": {
					Command:   binaryPath,
					Args:      []string{"mcp"},
					Transport: "stdio",
				},
			},
		}
		data, _ := json.MarshalIndent(config, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write config failed: %v", err)
		}

		if !integration.isMCPRegistered() {
			t.Error("should be registered when entry matches")
		}
	})

	t.Run("not registered when binary path differs", func(t *testing.T) {
		config := MCPConfig{
			MCPServers: map[string]MCPServerConfig{
				"portfolio": {
					Command:   "/different/path/portfolio",
					Transport: "stdio",
				},
			},
		}
		data, _ := json.MarshalIndent(config, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write config failed: %v", err)
		}

		if integration.isMCPRegistered() {
			t.Error("should not be registered when binary path differs")
		}
	})
}
