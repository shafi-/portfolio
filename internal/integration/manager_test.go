package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"project-dash/internal/integration"
	"project-dash/internal/integration/testutil"
)

func TestManager_Install_Success(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	result, err := manager.Install(ctx, "test-integration", integration.InstallOptions{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-integration", result.Name)
	assert.Equal(t, "test-agent", result.AgentType)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Equal(t, integration.StatusInstalled, result.Status)
	assert.True(t, fake.InstallCalled)
	assert.NotEmpty(t, result.InstallPath)
	assert.NotEmpty(t, result.InstalledAt)

	storedMeta, err := store.GetIntegration(ctx, "test-integration")
	require.NoError(t, err)
	assert.Equal(t, result.Name, storedMeta.Name)
	assert.Equal(t, result.Version, storedMeta.Version)
}

func TestManager_Install_Duplicate(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	_, err := manager.Install(ctx, "test-integration", integration.InstallOptions{})

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeAlreadyInstalled, err.(*integration.Error).Code)
	assert.False(t, fake.InstallCalled)
}

func TestManager_Install_Idempotent(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	_, err := manager.Install(ctx, "test-integration", integration.InstallOptions{})
	require.NoError(t, err)

	installCalledCount := fake.InstallCalled
	_, err = manager.Install(ctx, "test-integration", integration.InstallOptions{})

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeAlreadyInstalled, err.(*integration.Error).Code)
	assert.Equal(t, installCalledCount, fake.InstallCalled)
}

func TestManager_Install_WithForce(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	result, err := manager.Install(ctx, "test-integration", integration.InstallOptions{Force: true})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, fake.InstallCalled)
}

func TestManager_Install_NotFound(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")

	_, err := manager.Install(ctx, "non-existent", integration.InstallOptions{})

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeNotFound, err.(*integration.Error).Code)
}

func TestManager_Install_StoreError(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().WithStoreError(assert.AnError)
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	_, err := manager.Install(ctx, "test-integration", integration.InstallOptions{})

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeStoreUnavailable, err.(*integration.Error).Code)
}

func TestManager_List(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test1", "agent1", "1.0.0").
		WithIntegration("test2", "agent2", "2.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")

	result, err := manager.List(ctx)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "test1", result[0].Name)
	assert.Equal(t, "test2", result[1].Name)
}

func TestManager_List_Empty(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")

	result, err := manager.List(ctx)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestManager_Get(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")

	result, err := manager.Get(ctx, "test-integration")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-integration", result.Name)
	assert.Equal(t, "1.0.0", result.Version)
}

func TestManager_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")

	_, err := manager.Get(ctx, "non-existent")

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeNotFound, err.(*integration.Error).Code)
}

func TestMetaToJSONAndBack(t *testing.T) {
	meta := integration.IntegrationMeta{
		Name:             "test",
		AgentType:        "agent",
		Version:          "1.0.0",
		InstallPath:      "/path/to/integration",
		Status:           integration.StatusInstalled,
		MinEngineVersion: "v0.5.0",
		InstalledAt:      time.Now().Format(time.RFC3339),
		UpdatedAt:        time.Now().Format(time.RFC3339),
	}

	data, err := integration.MetaToJSON(meta)
	require.NoError(t, err)
	assert.NotNil(t, data)

	restored, err := integration.MetaFromJSON(data)
	require.NoError(t, err)
	assert.Equal(t, meta.Name, restored.Name)
	assert.Equal(t, meta.AgentType, restored.AgentType)
	assert.Equal(t, meta.Version, restored.Version)
	assert.Equal(t, meta.InstallPath, restored.InstallPath)
	assert.Equal(t, meta.Status, restored.Status)
	assert.Equal(t, meta.MinEngineVersion, restored.MinEngineVersion)
}

func TestGetMetaKey(t *testing.T) {
	key := integration.GetMetaKey("test-integration")
	assert.Equal(t, "integration:test-integration:meta", key)
}
