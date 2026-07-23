package integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"project-dash/internal/integration"
	"project-dash/internal/integration/testutil"
)

func TestManager_Validate_Success(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	result, err := manager.Validate(ctx, "test-integration")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.True(t, fake.ValidateCalled)
}

func TestManager_Validate_NotInstalled(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	result, err := manager.Validate(ctx, "test-integration")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Len(t, result.Checks, 1)
	assert.Equal(t, "integration_installed", result.Checks[0].Name)
	assert.Contains(t, result.Checks[0].Message, "not installed")
	assert.False(t, fake.ValidateCalled)
}

func TestManager_Doctor_SingleIntegration(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	result, err := manager.Doctor(ctx, "test-integration", false)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.True(t, fake.ValidateCalled)
}

func TestManager_Doctor_AllIntegrations(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test1", "agent1", "1.0.0").
		WithIntegration("test2", "agent2", "2.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake1 := testutil.NewFakeIntegration("test1", "agent1", "1.0.0")
	fake2 := testutil.NewFakeIntegration("test2", "agent2", "2.0.0")
	manager.RegisterIntegration(fake1)
	manager.RegisterIntegration(fake2)

	result, err := manager.Doctor(ctx, "", false)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.True(t, fake1.ValidateCalled)
	assert.True(t, fake2.ValidateCalled)
}

func TestManager_Doctor_Empty(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")

	result, err := manager.Doctor(ctx, "", false)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.Len(t, result.Checks, 1)
	assert.Equal(t, "no_integrations", result.Checks[0].Name)
}

func TestManager_Doctor_WithValidationFailures(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	validateResult := &integration.ValidationResult{
		Passed: false,
		Checks: []integration.ValidationCheck{
			{
				Name:         "mcp_server_reachable",
				Passed:       false,
				Message:      "MCP server not responding",
				SelfHealable: true,
			},
			{
				Name:   "agent_binary_exists",
				Passed: true,
			},
		},
	}
	fake := testutil.NewFakeIntegrationWithValidate("test-integration", "test-agent", "1.0.0", validateResult, nil)
	manager.RegisterIntegration(fake)

	result, err := manager.Doctor(ctx, "test-integration", false)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Len(t, result.Checks, 2)
}

func TestCheckMCPTools_Success(t *testing.T) {
	ctx := context.Background()
	mcpClient := testutil.NewFakeMCPClient()
	requiredTools := []string{"health", "listProjects"}

	result := integration.CheckMCPTools(ctx, mcpClient, requiredTools)

	assert.True(t, result.Passed)
	assert.Contains(t, result.Message, "required MCP tools available")
}

func TestCheckMCPTools_Missing(t *testing.T) {
	ctx := context.Background()
	mcpClient := testutil.NewFakeMCPClient()
	requiredTools := []string{"health", "nonexistent_tool"}

	result := integration.CheckMCPTools(ctx, mcpClient, requiredTools)

	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "Missing MCP tools")
	assert.Contains(t, result.Message, "nonexistent_tool")
}

func TestCheckMCPTools_ListError(t *testing.T) {
	ctx := context.Background()
	mcpClient := testutil.NewFakeMCPClient().WithToolListError(errors.New("list error"))
	requiredTools := []string{"health"}

	result := integration.CheckMCPTools(ctx, mcpClient, requiredTools)

	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "Failed to list MCP tools")
}

func TestCheckAgentBinary_Exists(t *testing.T) {
	ctx := context.Background()

	result := integration.CheckAgentBinary(ctx, "ls")

	assert.True(t, result.Passed)
	assert.Contains(t, result.Message, "found at")
}

func TestCheckAgentBinary_NotExists(t *testing.T) {
	ctx := context.Background()

	result := integration.CheckAgentBinary(ctx, "nonexistent_binary_xyz123")

	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "not found at")
}

func TestCheckAgentBinary_EmptyPath(t *testing.T) {
	ctx := context.Background()

	result := integration.CheckAgentBinary(ctx, "")

	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "not configured")
}

func TestCheckDirectoryExists_Exists(t *testing.T) {
	result := integration.CheckDirectoryExists("/tmp")

	assert.True(t, result.Passed)
	assert.Contains(t, result.Message, "exists at")
}

func TestCheckDirectoryExists_NotExists(t *testing.T) {
	result := integration.CheckDirectoryExists("/tmp/nonexistent_xyz123")

	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "does not exist")
	assert.True(t, result.SelfHealable)
}

func TestCheckDirectoryExists_EmptyPath(t *testing.T) {
	result := integration.CheckDirectoryExists("")

	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "not configured")
}

func TestCheckConfigFileExists_Exists(t *testing.T) {
	result := integration.CheckConfigFileExists("/etc/hosts")

	assert.True(t, result.Passed)
	assert.Contains(t, result.Message, "exists at")
}

func TestCheckConfigFileExists_NotExists(t *testing.T) {
	result := integration.CheckConfigFileExists("/tmp/nonexistent_xyz123.json")

	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "does not exist")
	assert.True(t, result.SelfHealable)
}

func TestCheckConfigFileExists_EmptyPath(t *testing.T) {
	result := integration.CheckConfigFileExists("")

	assert.False(t, result.Passed)
	assert.Contains(t, result.Message, "not configured")
}
