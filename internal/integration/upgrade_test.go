package integration_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"project-dash/internal/integration"
	"project-dash/internal/integration/testutil"
)

func TestManager_Upgrade_Success(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "upgrade-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	installPath := filepath.Join(tempDir, "integration")
	require.NoError(t, os.MkdirAll(installPath, 0755))

	meta := integration.IntegrationMeta{
		Name:        "test-integration",
		AgentType:   "test-agent",
		Version:     "1.0.0",
		InstallPath: installPath,
		Status:      integration.StatusInstalled,
	}

	store := testutil.NewFakeStore()
	require.NoError(t, store.SaveIntegration(ctx, meta))

	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	result, err := manager.Upgrade(ctx, "test-integration", integration.UpgradeOptions{TargetVersion: "2.0.0"})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "2.0.0", result.Version)
	assert.True(t, fake.UpgradeCalled)
	assert.Equal(t, "2.0.0", fake.VersionValue)
}

func TestManager_Upgrade_NotInstalled(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	_, err := manager.Upgrade(ctx, "test-integration", integration.UpgradeOptions{TargetVersion: "2.0.0"})

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeNotInstalled, err.(*integration.Error).Code)
}

func TestManager_Upgrade_Idempotent(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	result, err := manager.Upgrade(ctx, "test-integration", integration.UpgradeOptions{TargetVersion: "1.0.0"})

	require.NoError(t, err)
	assert.Equal(t, "1.0.0", result.Version)
	assert.False(t, fake.UpgradeCalled)
}

func TestManager_Upgrade_UpgradeFailed(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegrationWithUpgrade("test-integration", "test-agent", "1.0.0", nil, errors.New("upgrade failed"))
	manager.RegisterIntegration(fake)

	_, err := manager.Upgrade(ctx, "test-integration", integration.UpgradeOptions{TargetVersion: "2.0.0"})

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeUpgradeFailed, err.(*integration.Error).Code)
}

func TestManager_Upgrade_ValidationFailed(t *testing.T) {
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
				Name:   "test_check",
				Passed: false,
			},
		},
	}

	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	fake.ValidateFn = func(ctx context.Context) (*integration.ValidationResult, error) {
		return validateResult, nil
	}
	manager.RegisterIntegration(fake)

	_, err := manager.Upgrade(ctx, "test-integration", integration.UpgradeOptions{TargetVersion: "2.0.0"})

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeUpgradeFailed, err.(*integration.Error).Code)
}

func TestManager_Remove_Success(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	err := manager.Remove(ctx, "test-integration")

	require.NoError(t, err)
	assert.True(t, fake.RemoveCalled)

	_, err = store.GetIntegration(ctx, "test-integration")
	assert.Error(t, err)
}

func TestManager_Remove_NotInstalled(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")

	err := manager.Remove(ctx, "test-integration")

	assert.NoError(t, err)
}

func TestManager_Remove_StoreError(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	store.DeleteError = errors.New("delete failed")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	manager.RegisterIntegration(fake)

	err := manager.Remove(ctx, "test-integration")

	assert.Error(t, err)
	assert.Equal(t, integration.ErrCodeStoreUnavailable, err.(*integration.Error).Code)
}

func TestBackupRestoreIntegration(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("", "backup-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")

	require.NoError(t, os.MkdirAll(sourceDir, 0755))

	testFile := filepath.Join(sourceDir, "test.txt")
	testContent := []byte("test content")
	require.NoError(t, os.WriteFile(testFile, testContent, 0644))

	nestedDir := filepath.Join(sourceDir, "nested")
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	nestedFile := filepath.Join(nestedDir, "nested.txt")
	nestedContent := []byte("nested content")
	require.NoError(t, os.WriteFile(nestedFile, nestedContent, 0644))

	store := testutil.NewFakeStore().
		WithIntegration("test-integration", "test-agent", "1.0.0")
	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	fake := testutil.NewFakeIntegrationWithUpgrade("test-integration", "test-agent", "1.0.0", nil, errors.New("upgrade failed"))
	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	manager.RegisterIntegration(fake)

	_, err = manager.Upgrade(ctx, "test-integration", integration.UpgradeOptions{TargetVersion: "2.0.0"})
	assert.Error(t, err)

	backupPath := filepath.Join(sourceDir, "backup")
	if info, err := os.Stat(backupPath); err == nil {
		if info.IsDir() {
			backupContent, err := os.ReadFile(filepath.Join(backupPath, "test.txt"))
			require.NoError(t, err)
			assert.Equal(t, testContent, backupContent)
		}
	}
}

func TestCleanupBackupIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cleanup-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	installPath := filepath.Join(tempDir, "integration")
	backupPath := filepath.Join(installPath, "backup")

	require.NoError(t, os.MkdirAll(backupPath, 0755))

	testFile := filepath.Join(backupPath, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	ctx := context.Background()
	meta := integration.IntegrationMeta{
		Name:        "test-integration",
		AgentType:   "test-agent",
		Version:     "1.0.0",
		InstallPath: installPath,
		Status:      integration.StatusInstalled,
	}

	store := testutil.NewFakeStore()
	require.NoError(t, store.SaveIntegration(ctx, meta))

	mcpClient := testutil.NewFakeMCPClient()
	logger := zap.NewNop()

	fake := testutil.NewFakeIntegration("test-integration", "test-agent", "1.0.0")
	fake.UpgradeFn = func(ctx context.Context, opts integration.UpgradeOptions) (*integration.UpgradeResult, error) {
		return &integration.UpgradeResult{
			PreviousVersion: "1.0.0",
			NewVersion:      "2.0.0",
			RolledBack:      false,
		}, nil
	}
	fake.ValidateFn = func(ctx context.Context) (*integration.ValidationResult, error) {
		return &integration.ValidationResult{
			Passed: true,
			Checks: []integration.ValidationCheck{},
		}, nil
	}

	manager := integration.NewManager(store, mcpClient, logger, "v0.5.0")
	manager.RegisterIntegration(fake)

	_, err = manager.Upgrade(ctx, "test-integration", integration.UpgradeOptions{TargetVersion: "2.0.0"})
	require.NoError(t, err)

	_, err = os.Stat(backupPath)
	assert.True(t, os.IsNotExist(err), "Backup directory should have been cleaned up")
}
