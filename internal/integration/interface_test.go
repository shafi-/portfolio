package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"project-dash/internal/integration"
	"project-dash/internal/integration/testutil"
)

func TestFakeIntegration_Name(t *testing.T) {
	fake := testutil.NewFakeIntegration("test", "agent", "1.0.0")
	assert.Equal(t, "test", fake.Name())
}

func TestFakeIntegration_AgentType(t *testing.T) {
	fake := testutil.NewFakeIntegration("test", "agent", "1.0.0")
	assert.Equal(t, "agent", fake.AgentType())
}

func TestFakeIntegration_Install(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeIntegration("test", "agent", "1.0.0")

	result, err := fake.Install(ctx, integration.InstallOptions{InstallPath: "/path"})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test", result.Meta.Name)
	assert.True(t, fake.InstallCalled)
	assert.Equal(t, integration.StatusInstalled, result.Meta.Status)
}

func TestFakeIntegration_InstallWithCallback(t *testing.T) {
	ctx := context.Background()
	customInstall := func(ctx context.Context, opts integration.InstallOptions) (*integration.InstallResult, error) {
		return &integration.InstallResult{
			Meta: integration.IntegrationMeta{
				Name:    "custom",
				Status:  integration.StatusInstalled,
				Version: "2.0.0",
			},
			Warnings: []string{"custom warning"},
		}, nil
	}

	fake := &testutil.FakeIntegration{
		NameValue:      "test",
		AgentTypeValue: "agent",
		VersionValue:   "1.0.0",
		InstallFn:      customInstall,
	}

	result, err := fake.Install(ctx, integration.InstallOptions{})

	require.NoError(t, err)
	assert.Equal(t, "custom", result.Meta.Name)
	assert.Equal(t, []string{"custom warning"}, result.Warnings)
}

func TestFakeIntegration_InstallWithError(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeIntegrationWithError("test", "agent", "1.0.0", integration.ErrAlreadyInstalled)

	_, err := fake.Install(ctx, integration.InstallOptions{})

	assert.Error(t, err)
	assert.Equal(t, integration.ErrAlreadyInstalled, err)
}

func TestFakeIntegration_Validate(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeIntegration("test", "agent", "1.0.0")

	result, err := fake.Validate(ctx)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.True(t, fake.ValidateCalled)
}

func TestFakeIntegration_ValidateWithResult(t *testing.T) {
	ctx := context.Background()
	validateResult := &integration.ValidationResult{
		Passed: false,
		Checks: []integration.ValidationCheck{
			{
				Name:    "test_check",
				Passed:  false,
				Message: "failed",
			},
		},
	}

	fake := testutil.NewFakeIntegrationWithValidate("test", "agent", "1.0.0", validateResult, nil)

	result, err := fake.Validate(ctx)

	require.NoError(t, err)
	assert.Equal(t, validateResult, result)
}

func TestFakeIntegration_Upgrade(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeIntegration("test", "agent", "1.0.0")

	result, err := fake.Upgrade(ctx, integration.UpgradeOptions{TargetVersion: "2.0.0"})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "1.0.0", result.PreviousVersion)
	assert.Equal(t, "2.0.0", result.NewVersion)
	assert.False(t, result.RolledBack)
	assert.Equal(t, "2.0.0", fake.VersionValue)
	assert.True(t, fake.UpgradeCalled)
}

func TestFakeIntegration_UpgradeWithError(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeIntegrationWithUpgrade("test", "agent", "1.0.0", nil, assert.AnError)

	_, err := fake.Upgrade(ctx, integration.UpgradeOptions{TargetVersion: "2.0.0"})

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestFakeIntegration_Remove(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeIntegration("test", "agent", "1.0.0")

	err := fake.Remove(ctx)

	require.NoError(t, err)
	assert.True(t, fake.RemoveCalled)
	assert.Equal(t, integration.StatusNotInstalled, fake.Meta.Status)
}

func TestFakeIntegration_RemoveWithError(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeIntegrationWithRemove("test", "agent", "1.0.0", assert.AnError)

	err := fake.Remove(ctx)

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestFakeMCPClient_Health(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeMCPClient()

	err := fake.Health(ctx)

	require.NoError(t, err)
}

func TestFakeMCPClient_HealthUnhealthy(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeMCPClient().WithUnhealthy()

	err := fake.Health(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not responding")
}

func TestFakeMCPClient_ListTools(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeMCPClient()

	tools, err := fake.ListTools(ctx)

	require.NoError(t, err)
	assert.NotNil(t, tools)
	assert.Greater(t, len(tools), 0)
}

func TestFakeMCPClient_RegisterTools(t *testing.T) {
	ctx := context.Background()
	fake := testutil.NewFakeMCPClient()

	tools := []integration.ToolDef{
		{Name: "test_tool"},
	}

	err := fake.RegisterTools(ctx, tools)

	require.NoError(t, err)
	assert.Contains(t, fake.Tools, "test_tool")
}

func TestFakeStore_SaveAndLoad(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()
	meta := integration.IntegrationMeta{
		Name:      "test",
		AgentType: "agent",
		Version:   "1.0.0",
		Status:    integration.StatusInstalled,
	}

	err := store.SaveIntegration(ctx, meta)
	require.NoError(t, err)

	loaded, err := store.GetIntegration(ctx, "test")
	require.NoError(t, err)
	assert.Equal(t, meta.Name, loaded.Name)
	assert.Equal(t, meta.AgentType, loaded.AgentType)
	assert.Equal(t, meta.Version, loaded.Version)
}

func TestFakeStore_GetNotFound(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()

	_, err := store.GetIntegration(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, integration.ErrNotFound, err)
}

func TestFakeStore_List(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test1", "agent1", "1.0.0").
		WithIntegration("test2", "agent2", "2.0.0")

	list, err := store.ListIntegrations(ctx)

	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestFakeStore_Delete(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore().
		WithIntegration("test", "agent", "1.0.0")

	err := store.DeleteIntegration(ctx, "test")
	require.NoError(t, err)

	_, err = store.GetIntegration(ctx, "test")
	assert.Error(t, err)
	assert.Equal(t, integration.ErrNotFound, err)
}

func TestFakeStore_DeleteNotFound(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewFakeStore()

	err := store.DeleteIntegration(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, integration.ErrNotFound, err)
}

func TestErrorCreation(t *testing.T) {
	err := integration.NewError(integration.ErrCodeNotFound, "test message", assert.AnError)

	assert.Equal(t, integration.ErrCodeNotFound, err.Code)
	assert.Equal(t, "test message", err.Message)
	assert.Equal(t, assert.AnError, err.Cause)
}

func TestErrorUnwrap(t *testing.T) {
	cause := assert.AnError
	err := integration.NewError(integration.ErrCodeNotFound, "test message", cause)

	assert.Equal(t, cause, err.Unwrap())
}

func TestErrorString(t *testing.T) {
	err := integration.NewError(integration.ErrCodeNotFound, "test message", nil)

	str := err.Error()
	assert.Contains(t, str, integration.ErrCodeNotFound)
	assert.Contains(t, str, "test message")
}

func TestErrorStringWithCause(t *testing.T) {
	err := integration.NewError(integration.ErrCodeNotFound, "test message", assert.AnError)

	str := err.Error()
	assert.Contains(t, str, integration.ErrCodeNotFound)
	assert.Contains(t, str, "test message")
	assert.Contains(t, str, assert.AnError.Error())
}
