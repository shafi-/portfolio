package claude

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"project-dash/internal/integration"
)

func TestIntegrationInstallFlow(t *testing.T) {
	tempDir := t.TempDir()

	configPath := filepath.Join(tempDir, "settings.json")
	skillsDir := filepath.Join(tempDir, "skills")
	binaryPath := "/usr/local/bin/portfolio"

	mockStore := &mockStoreWithSave{
		mockStore: &mockStore{
			getIntegrationFunc: func(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
				return nil, integration.ErrNotFound
			},
		},
		saveIntegrationFunc: func(ctx context.Context, meta integration.IntegrationMeta) error {
			return nil
		},
	}

	mockMCP := &mockMCPClient{
		healthFunc: func(ctx context.Context) error {
			return nil
		},
		listToolsFunc: func(ctx context.Context) ([]string, error) {
			return []string{
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
			}, nil
		},
		registerToolsFunc: func(ctx context.Context, tools []integration.ToolDef) error {
			return nil
		},
	}

	c := &ClaudeCodeIntegration{
		store:  mockStore,
		mcp:    mockMCP,
		logger: zap.NewNop(),
		config: ClaudeConfig{
			ConfigPath: configPath,
			SkillsDir:  skillsDir,
			BinaryPath: binaryPath,
		},
	}

	t.Run("fresh install creates all files", func(t *testing.T) {
		if err := c.ensureMCPConfig(); err != nil {
			t.Fatalf("ensureMCPConfig failed: %v", err)
		}

		if err := c.installSkill(); err != nil {
			t.Fatalf("installSkill failed: %v", err)
		}

		if _, err := os.Stat(configPath); err != nil {
			t.Errorf("config file not created: %v", err)
		}

		skillPath := filepath.Join(skillsDir, "portfolio.md")
		if _, err := os.Stat(skillPath); err != nil {
			t.Errorf("skill file not created: %v", err)
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read config failed: %v", err)
		}

		var config MCPConfig
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("parse config failed: %v", err)
		}

		if _, exists := config.MCPServers["portfolio"]; !exists {
			t.Error("portfolio MCP server not registered")
		}
	})

	t.Run("force reinstall overwrites existing files", func(t *testing.T) {
		oldSkillContent := []byte("old skill content")
		skillPath := filepath.Join(skillsDir, "portfolio.md")
		if err := os.WriteFile(skillPath, oldSkillContent, 0644); err != nil {
			t.Fatalf("write old skill failed: %v", err)
		}

		if err := c.installSkill(); err != nil {
			t.Fatalf("installSkill failed: %v", err)
		}

		newSkillContent, err := os.ReadFile(skillPath)
		if err != nil {
			t.Fatalf("read new skill failed: %v", err)
		}

		if string(newSkillContent) == string(oldSkillContent) {
			t.Error("skill file was not overwritten during reinstall")
		}
	})

	t.Run("validation after install", func(t *testing.T) {
		c.logger = zap.NewNop()
		result, err := c.Validate(context.Background())
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if result.Passed {
			t.Error("validation should still fail for some checks like binary exists")
		}

		hasPassingChecks := false
		for _, check := range result.Checks {
			if check.Passed && check.Name == "config_file_exists" {
				hasPassingChecks = true
				break
			}
		}

		if !hasPassingChecks {
			t.Error("should have at least one passing check")
		}
	})
}

func TestIntegrationUpgradeFlow(t *testing.T) {
	tempDir := t.TempDir()

	configPath := filepath.Join(tempDir, "settings.json")
	skillsDir := filepath.Join(tempDir, "skills")
	binaryPath := "/usr/local/bin/portfolio"

	mockStore := &mockStoreWithVersion{
		mockStore: &mockStore{
			getIntegrationFunc: func(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
				if name == "claude" {
					return &integration.IntegrationMeta{
						Name:    "claude",
						Version: "1.0.0",
						Status:  integration.StatusInstalled,
					}, nil
				}
				return nil, integration.ErrNotFound
			},
		},
		saveIntegrationFunc: func(ctx context.Context, meta integration.IntegrationMeta) error {
			return nil
		},
		currentVersion: "1.0.0",
	}

	mockMCP := &mockMCPClient{
		healthFunc: func(ctx context.Context) error {
			return nil
		},
		listToolsFunc: func(ctx context.Context) ([]string, error) {
			return []string{
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
			}, nil
		},
		registerToolsFunc: func(ctx context.Context, tools []integration.ToolDef) error {
			return nil
		},
	}

	c := &ClaudeCodeIntegration{
		store:  mockStore,
		mcp:    mockMCP,
		logger: zap.NewNop(),
		config: ClaudeConfig{
			ConfigPath: configPath,
			SkillsDir:  skillsDir,
			BinaryPath: binaryPath,
		},
	}

	t.Run("upgrade updates all files", func(t *testing.T) {
		oldConfig := MCPConfig{
			MCPServers: map[string]MCPServerConfig{
				"portfolio": {
					Command:   "/old/path/portfolio",
					Args:      []string{"mcp"},
					Transport: "stdio",
				},
			},
		}
		data, _ := json.MarshalIndent(oldConfig, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write old config failed: %v", err)
		}

		if err := c.ensureMCPConfig(); err != nil {
			t.Fatalf("ensureMCPConfig failed: %v", err)
		}

		newData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read new config failed: %v", err)
		}

		var newConfig MCPConfig
		if err := json.Unmarshal(newData, &newConfig); err != nil {
			t.Fatalf("parse new config failed: %v", err)
		}

		server := newConfig.MCPServers["portfolio"]
		if server.Command != binaryPath {
			t.Errorf("expected command '%s', got '%s'", binaryPath, server.Command)
		}
	})

	t.Run("upgrade preserves other MCP servers", func(t *testing.T) {
		configWithOthers := MCPConfig{
			MCPServers: map[string]MCPServerConfig{
				"portfolio": {
					Command:   "/old/path/portfolio",
					Args:      []string{"mcp"},
					Transport: "stdio",
				},
				"other": {
					Command:   "/path/to/other",
					Transport: "stdio",
				},
			},
		}
		data, _ := json.MarshalIndent(configWithOthers, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write config failed: %v", err)
		}

		if err := c.ensureMCPConfig(); err != nil {
			t.Fatalf("ensureMCPConfig failed: %v", err)
		}

		newData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read new config failed: %v", err)
		}

		var newConfig MCPConfig
		if err := json.Unmarshal(newData, &newConfig); err != nil {
			t.Fatalf("parse new config failed: %v", err)
		}

		if _, exists := newConfig.MCPServers["other"]; !exists {
			t.Error("other MCP server was removed during upgrade")
		}
	})
}

func TestIntegrationUninstallFlow(t *testing.T) {
	tempDir := t.TempDir()

	configPath := filepath.Join(tempDir, "settings.json")
	skillsDir := filepath.Join(tempDir, "skills")

	mockStore := &mockStoreWithDelete{
		mockStore: &mockStore{
			getIntegrationFunc: func(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
				if name == "claude" {
					return &integration.IntegrationMeta{
						Name:    "claude",
						Version: "1.0.0",
						Status:  integration.StatusInstalled,
					}, nil
				}
				return nil, integration.ErrNotFound
			},
		},
		deleteIntegrationFunc: func(ctx context.Context, name string) error {
			return nil
		},
	}

	c := &ClaudeCodeIntegration{
		store:  mockStore,
		logger: zap.NewNop(),
		config: ClaudeConfig{
			ConfigPath: configPath,
			SkillsDir:  skillsDir,
		},
	}

	t.Run("uninstall removes only portfolio entries", func(t *testing.T) {
		ctx := context.Background()
		configWithOthers := MCPConfig{
			MCPServers: map[string]MCPServerConfig{
				"portfolio": {
					Command:   "/usr/local/bin/portfolio",
					Args:      []string{"mcp"},
					Transport: "stdio",
				},
				"other": {
					Command:   "/path/to/other",
					Transport: "stdio",
				},
			},
		}
		data, _ := json.MarshalIndent(configWithOthers, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write config failed: %v", err)
		}

		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			t.Fatalf("create skills dir failed: %v", err)
		}

		skillPath := filepath.Join(skillsDir, "portfolio.md")
		if err := os.WriteFile(skillPath, []byte("test skill"), 0644); err != nil {
			t.Fatalf("write skill failed: %v", err)
		}

		if err := c.Remove(ctx); err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		if _, err := os.Stat(skillPath); err == nil {
			t.Error("skill file still exists after uninstall")
		} else if !os.IsNotExist(err) {
			t.Fatalf("unexpected error checking skill file: %v", err)
		}

		remainingData, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("read remaining config failed: %v", err)
		}

		var remainingConfig MCPConfig
		if err := json.Unmarshal(remainingData, &remainingConfig); err != nil {
			t.Fatalf("parse remaining config failed: %v", err)
		}

		if _, exists := remainingConfig.MCPServers["portfolio"]; exists {
			t.Error("portfolio MCP server still exists after uninstall")
		}

		if _, exists := remainingConfig.MCPServers["other"]; !exists {
			t.Error("other MCP server was incorrectly removed")
		}
	})

	t.Run("uninstall is idempotent", func(t *testing.T) {
		ctx := context.Background()
		firstErr := c.Remove(ctx)
		if firstErr != nil {
			t.Fatalf("first Remove failed: %v", firstErr)
		}

		secondErr := c.Remove(ctx)
		if secondErr != nil {
			t.Fatalf("second Remove failed: %v", secondErr)
		}
	})
}

type mockStoreWithSave struct {
	mockStore           *mockStore
	saveIntegrationFunc func(context.Context, integration.IntegrationMeta) error
}

func (m *mockStoreWithSave) GetIntegration(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
	return m.mockStore.GetIntegration(ctx, name)
}

func (m *mockStoreWithSave) SaveIntegration(ctx context.Context, meta integration.IntegrationMeta) error {
	if m.saveIntegrationFunc != nil {
		return m.saveIntegrationFunc(ctx, meta)
	}
	return nil
}

func (m *mockStoreWithSave) DeleteIntegration(ctx context.Context, name string) error {
	return m.mockStore.DeleteIntegration(ctx, name)
}

func (m *mockStoreWithSave) ListIntegrations(ctx context.Context) ([]integration.IntegrationMeta, error) {
	return m.mockStore.ListIntegrations(ctx)
}

type mockStoreWithVersion struct {
	mockStore           *mockStore
	saveIntegrationFunc func(context.Context, integration.IntegrationMeta) error
	currentVersion      string
}

func (m *mockStoreWithVersion) GetIntegration(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
	return m.mockStore.GetIntegration(ctx, name)
}

func (m *mockStoreWithVersion) SaveIntegration(ctx context.Context, meta integration.IntegrationMeta) error {
	if m.saveIntegrationFunc != nil {
		return m.saveIntegrationFunc(ctx, meta)
	}
	m.currentVersion = meta.Version
	return nil
}

func (m *mockStoreWithVersion) DeleteIntegration(ctx context.Context, name string) error {
	return m.mockStore.DeleteIntegration(ctx, name)
}

func (m *mockStoreWithVersion) ListIntegrations(ctx context.Context) ([]integration.IntegrationMeta, error) {
	return m.mockStore.ListIntegrations(ctx)
}
