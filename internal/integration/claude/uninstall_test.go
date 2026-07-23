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

func TestUninstallIntegration(t *testing.T) {
	ctx := context.Background()

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

	integration := &ClaudeCodeIntegration{
		store:  mockStore,
		logger: zap.NewNop(),
		config: ClaudeConfig{
			ConfigPath: configPath,
			SkillsDir:  skillsDir,
		},
	}

	t.Run("uninstall removes MCP config", func(t *testing.T) {
		initialConfig := MCPConfig{
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

		data, _ := json.MarshalIndent(initialConfig, "", "  ")
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			t.Fatalf("write initial config failed: %v", err)
		}

		err := integration.Remove(ctx)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
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

	t.Run("uninstall removes skill file", func(t *testing.T) {
		skillPath := filepath.Join(skillsDir, "portfolio.md")
		if err := os.MkdirAll(skillsDir, 0755); err != nil {
			t.Fatalf("create skills dir failed: %v", err)
		}

		if err := os.WriteFile(skillPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("write skill file failed: %v", err)
		}

		err := integration.Remove(ctx)
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		if _, err := os.Stat(skillPath); err == nil {
			t.Error("skill file still exists after uninstall")
		} else if !os.IsNotExist(err) {
			t.Fatalf("unexpected error checking skill file: %v", err)
		}
	})

	t.Run("uninstall is idempotent", func(t *testing.T) {
		firstErr := integration.Remove(ctx)
		if firstErr != nil {
			t.Fatalf("first Remove failed: %v", firstErr)
		}

		secondErr := integration.Remove(ctx)
		if secondErr != nil {
			t.Fatalf("second Remove failed: %v", secondErr)
		}
	})
}

func TestUninstallNotInstalled(t *testing.T) {
	ctx := context.Background()

	mockStore := &mockStore{
		getIntegrationFunc: func(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
			return nil, integration.ErrNotFound
		},
	}

	tempDir := t.TempDir()
	integration := &ClaudeCodeIntegration{
		store:  mockStore,
		logger: zap.NewNop(),
		config: ClaudeConfig{
			ConfigPath: tempDir + "/settings.json",
			SkillsDir:  tempDir + "/skills",
		},
	}

	err := integration.Remove(ctx)
	if err != nil {
		t.Fatalf("Remove failed for not-installed integration: %v", err)
	}
}

type mockStoreWithDelete struct {
	mockStore             *mockStore
	deleteIntegrationFunc func(context.Context, string) error
}

func (m *mockStoreWithDelete) GetIntegration(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
	return m.mockStore.GetIntegration(ctx, name)
}

func (m *mockStoreWithDelete) SaveIntegration(ctx context.Context, meta integration.IntegrationMeta) error {
	return m.mockStore.SaveIntegration(ctx, meta)
}

func (m *mockStoreWithDelete) DeleteIntegration(ctx context.Context, name string) error {
	if m.deleteIntegrationFunc != nil {
		return m.deleteIntegrationFunc(ctx, name)
	}
	return nil
}

func (m *mockStoreWithDelete) ListIntegrations(ctx context.Context) ([]integration.IntegrationMeta, error) {
	return m.mockStore.ListIntegrations(ctx)
}
