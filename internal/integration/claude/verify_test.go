package claude

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"project-dash/internal/integration"
)

func TestClaudeIntegrationBasics(t *testing.T) {
	mockStore := &mockStore{
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
	}

	mockMCP := &mockMCPClient{
		healthFunc: func(ctx context.Context) error {
			return nil
		},
		listToolsFunc: func(ctx context.Context) ([]string, error) {
			return []string{"health", "listProjects", "searchProjects"}, nil
		},
		registerToolsFunc: func(ctx context.Context, tools []integration.ToolDef) error {
			return nil
		},
	}

	tempDir := t.TempDir()

	integration := &ClaudeCodeIntegration{
		store: mockStore,
		mcp:   mockMCP,
		config: ClaudeConfig{
			ConfigPath: tempDir + "/settings.json",
			SkillsDir:  tempDir + "/skills",
		},
	}

	t.Run("integration name", func(t *testing.T) {
		if integration.Name() != "claude" {
			t.Errorf("expected name 'claude', got '%s'", integration.Name())
		}
	})

	t.Run("agent type", func(t *testing.T) {
		if integration.AgentType() != "claude-code" {
			t.Errorf("expected agent type 'claude-code', got '%s'", integration.AgentType())
		}
	})
}

func TestValidateIntegration(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	mockStore := &mockStore{
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
			ConfigPath: tempDir + "/settings.json",
			SkillsDir:  tempDir + "/skills",
		},
	}

	t.Run("full validation with missing checks", func(t *testing.T) {
		result, err := c.Validate(ctx)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if result.Passed {
			t.Error("validation should fail with missing config and skill files")
		}

		checkNames := map[string]bool{}
		for _, check := range result.Checks {
			checkNames[check.Name] = true
		}

		expectedChecks := []string{
			"claude_installed",
			"portfolio_binary_exists",
			"integration_installed",
			"config_file_exists",
			"mcp_entry_registered",
			"mcp_server_reachable",
			"mcp_tools_available",
			"skill_file_exists",
		}

		for _, expected := range expectedChecks {
			if !checkNames[expected] {
				t.Errorf("missing expected check: %s", expected)
			}
		}
	})

	t.Run("tools available check passes", func(t *testing.T) {
		allTools := []string{
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

		mockMCP.listToolsFunc = func(ctx context.Context) ([]string, error) {
			return allTools, nil
		}

		result, err := c.Validate(ctx)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		toolsCheck := integration.ValidationCheck{}
		for _, check := range result.Checks {
			if check.Name == "mcp_tools_available" {
				toolsCheck = check
				break
			}
		}

		if !toolsCheck.Passed {
			t.Errorf("tools check should pass with all tools available: %s", toolsCheck.Message)
		}
	})

	t.Run("tools available check fails with missing tools", func(t *testing.T) {
		partialTools := []string{
			"health",
			"listProjects",
		}

		mockMCP.listToolsFunc = func(ctx context.Context) ([]string, error) {
			return partialTools, nil
		}

		result, err := c.Validate(ctx)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		toolsCheck := integration.ValidationCheck{}
		for _, check := range result.Checks {
			if check.Name == "mcp_tools_available" {
				toolsCheck = check
				break
			}
		}

		if toolsCheck.Passed {
			t.Error("tools check should fail with missing tools")
		}

		if toolsCheck.Remediation == "" {
			t.Error("tools check should have remediation message")
		}
	})

	t.Run("tools available check fails with error", func(t *testing.T) {
		mockMCP.listToolsFunc = func(ctx context.Context) ([]string, error) {
			return nil, integration.ErrStoreUnavailable
		}

		result, err := c.Validate(ctx)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		toolsCheck := integration.ValidationCheck{}
		for _, check := range result.Checks {
			if check.Name == "mcp_tools_available" {
				toolsCheck = check
				break
			}
		}

		if toolsCheck.Passed {
			t.Error("tools check should fail with server error")
		}

		if !toolsCheck.SelfHealable {
			t.Error("tools check should be self-healable when server fails")
		}
	})
}

func TestValidateIntegrationChecks(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	mockStore := &mockStore{
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
	}

	mockMCP := &mockMCPClient{
		healthFunc: func(ctx context.Context) error {
			return nil
		},
		listToolsFunc: func(ctx context.Context) ([]string, error) {
			return []string{"health"}, nil
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
			ConfigPath: tempDir + "/settings.json",
			SkillsDir:  tempDir + "/skills",
		},
	}

	t.Run("claude installed check", func(t *testing.T) {
		check := c.checkClaudeInstalled()
		if check.Name != "claude_installed" {
			t.Errorf("expected check name 'claude_installed', got '%s'", check.Name)
		}
		if !check.Passed && check.Remediation == "" {
			t.Error("remediation should not be empty when check fails")
		}
	})

	t.Run("binary exists check with missing binary", func(t *testing.T) {
		c.config.BinaryPath = "/nonexistent/binary"
		check := c.checkBinaryExists()
		if check.Passed {
			t.Error("check should fail with missing binary")
		}
		if check.Remediation == "" {
			t.Error("remediation should not be empty")
		}
	})

	t.Run("integration installed check with not found", func(t *testing.T) {
		mockStore.getIntegrationFunc = func(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
			return nil, integration.ErrNotFound
		}
		check := c.checkIntegrationInstalled(ctx)
		if check.Passed {
			t.Error("check should fail when integration not found")
		}
		if check.Remediation == "" {
			t.Error("remediation should not be empty")
		}
	})

	t.Run("integration installed check passes", func(t *testing.T) {
		mockStore.getIntegrationFunc = func(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
			if name == "claude" {
				return &integration.IntegrationMeta{
					Name:    "claude",
					Version: "1.0.0",
					Status:  integration.StatusInstalled,
				}, nil
			}
			return nil, integration.ErrNotFound
		}
		check := c.checkIntegrationInstalled(ctx)
		if !check.Passed {
			t.Errorf("check should pass when integration found: %s", check.Message)
		}
	})
}

type mockStore struct {
	getIntegrationFunc func(context.Context, string) (*integration.IntegrationMeta, error)
}

func (m *mockStore) GetIntegration(ctx context.Context, name string) (*integration.IntegrationMeta, error) {
	if m.getIntegrationFunc != nil {
		return m.getIntegrationFunc(ctx, name)
	}
	return nil, integration.ErrNotFound
}

func (m *mockStore) SaveIntegration(ctx context.Context, meta integration.IntegrationMeta) error {
	return nil
}

func (m *mockStore) DeleteIntegration(ctx context.Context, name string) error {
	return nil
}

func (m *mockStore) ListIntegrations(ctx context.Context) ([]integration.IntegrationMeta, error) {
	return []integration.IntegrationMeta{}, nil
}

type mockMCPClient struct {
	healthFunc        func(context.Context) error
	listToolsFunc     func(context.Context) ([]string, error)
	registerToolsFunc func(context.Context, []integration.ToolDef) error
}

func (m *mockMCPClient) Health(ctx context.Context) error {
	if m.healthFunc != nil {
		return m.healthFunc(ctx)
	}
	return nil
}

func (m *mockMCPClient) ListTools(ctx context.Context) ([]string, error) {
	if m.listToolsFunc != nil {
		return m.listToolsFunc(ctx)
	}
	return []string{}, nil
}

func (m *mockMCPClient) RegisterTools(ctx context.Context, tools []integration.ToolDef) error {
	if m.registerToolsFunc != nil {
		return m.registerToolsFunc(ctx, tools)
	}
	return nil
}
