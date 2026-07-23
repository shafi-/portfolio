package testutil

import (
	"context"
	"errors"

	"project-dash/internal/integration"
)

type FakeIntegration struct {
	NameValue      string
	AgentTypeValue string
	VersionValue   string
	Meta           integration.IntegrationMeta
	InstallFn      func(ctx context.Context, opts integration.InstallOptions) (*integration.InstallResult, error)
	ValidateFn     func(ctx context.Context) (*integration.ValidationResult, error)
	UpgradeFn      func(ctx context.Context, opts integration.UpgradeOptions) (*integration.UpgradeResult, error)
	RemoveFn       func(ctx context.Context) error
	InstallCalled  bool
	ValidateCalled bool
	UpgradeCalled  bool
	RemoveCalled   bool
}

func NewFakeIntegration(name, agentType, version string) *FakeIntegration {
	return &FakeIntegration{
		NameValue:      name,
		AgentTypeValue: agentType,
		VersionValue:   version,
		Meta: integration.IntegrationMeta{
			Name:      name,
			AgentType: agentType,
			Version:   version,
			Status:    integration.StatusNotInstalled,
		},
	}
}

func (f *FakeIntegration) Name() string {
	return f.NameValue
}

func (f *FakeIntegration) AgentType() string {
	return f.AgentTypeValue
}

func (f *FakeIntegration) Install(ctx context.Context, opts integration.InstallOptions) (*integration.InstallResult, error) {
	f.InstallCalled = true
	if f.InstallFn != nil {
		return f.InstallFn(ctx, opts)
	}
	f.Meta.Status = integration.StatusInstalled
	f.Meta.InstallPath = opts.InstallPath
	return &integration.InstallResult{
		Meta: f.Meta,
	}, nil
}

func (f *FakeIntegration) Validate(ctx context.Context) (*integration.ValidationResult, error) {
	f.ValidateCalled = true
	if f.ValidateFn != nil {
		return f.ValidateFn(ctx)
	}
	return &integration.ValidationResult{
		Passed: true,
		Checks: []integration.ValidationCheck{},
	}, nil
}

func (f *FakeIntegration) Upgrade(ctx context.Context, opts integration.UpgradeOptions) (*integration.UpgradeResult, error) {
	f.UpgradeCalled = true
	if f.UpgradeFn != nil {
		return f.UpgradeFn(ctx, opts)
	}
	prevVersion := f.VersionValue
	f.VersionValue = opts.TargetVersion
	f.Meta.Version = opts.TargetVersion
	return &integration.UpgradeResult{
		PreviousVersion: prevVersion,
		NewVersion:      opts.TargetVersion,
		RolledBack:      false,
	}, nil
}

func (f *FakeIntegration) Remove(ctx context.Context) error {
	f.RemoveCalled = true
	if f.RemoveFn != nil {
		return f.RemoveFn(ctx)
	}
	f.Meta.Status = integration.StatusNotInstalled
	return nil
}

func NewFakeIntegrationWithError(name, agentType, version string, installErr error) *FakeIntegration {
	return &FakeIntegration{
		NameValue:      name,
		AgentTypeValue: agentType,
		VersionValue:   version,
		Meta: integration.IntegrationMeta{
			Name:      name,
			AgentType: agentType,
			Version:   version,
			Status:    integration.StatusNotInstalled,
		},
		InstallFn: func(ctx context.Context, opts integration.InstallOptions) (*integration.InstallResult, error) {
			return nil, installErr
		},
	}
}

func NewFakeIntegrationWithValidate(name, agentType, version string, validateResult *integration.ValidationResult, validateErr error) *FakeIntegration {
	if validateResult == nil && validateErr == nil {
		validateResult = &integration.ValidationResult{
			Passed: true,
			Checks: []integration.ValidationCheck{},
		}
	}
	return &FakeIntegration{
		NameValue:      name,
		AgentTypeValue: agentType,
		VersionValue:   version,
		Meta: integration.IntegrationMeta{
			Name:      name,
			AgentType: agentType,
			Version:   version,
			Status:    integration.StatusInstalled,
		},
		ValidateFn: func(ctx context.Context) (*integration.ValidationResult, error) {
			if validateErr != nil {
				return nil, validateErr
			}
			if validateResult != nil {
				return validateResult, nil
			}
			return nil, errors.New("no validation result")
		},
	}
}

func NewFakeIntegrationWithUpgrade(name, agentType, version string, upgradeResult *integration.UpgradeResult, upgradeErr error) *FakeIntegration {
	return &FakeIntegration{
		NameValue:      name,
		AgentTypeValue: agentType,
		VersionValue:   version,
		Meta: integration.IntegrationMeta{
			Name:      name,
			AgentType: agentType,
			Version:   version,
			Status:    integration.StatusInstalled,
		},
		UpgradeFn: func(ctx context.Context, opts integration.UpgradeOptions) (*integration.UpgradeResult, error) {
			if upgradeErr != nil {
				return nil, upgradeErr
			}
			return upgradeResult, nil
		},
	}
}

func NewFakeIntegrationWithRemove(name, agentType, version string, removeErr error) *FakeIntegration {
	return &FakeIntegration{
		NameValue:      name,
		AgentTypeValue: agentType,
		VersionValue:   version,
		Meta: integration.IntegrationMeta{
			Name:      name,
			AgentType: agentType,
			Version:   version,
			Status:    integration.StatusInstalled,
		},
		RemoveFn: func(ctx context.Context) error {
			if removeErr != nil {
				return removeErr
			}
			return nil
		},
	}
}

func NewFakeIntegrationAlreadyInstalled() *FakeIntegration {
	return &FakeIntegration{
		NameValue:      "already-installed",
		AgentTypeValue: "test-agent",
		VersionValue:   "1.0.0",
		Meta: integration.IntegrationMeta{
			Name:      "already-installed",
			AgentType: "test-agent",
			Version:   "1.0.0",
			Status:    integration.StatusInstalled,
		},
		InstallFn: func(ctx context.Context, opts integration.InstallOptions) (*integration.InstallResult, error) {
			return nil, integration.ErrAlreadyInstalled
		},
	}
}

func NewFakeIntegrationWithValidateFailure() *FakeIntegration {
	return &FakeIntegration{
		NameValue:      "validation-failure",
		AgentTypeValue: "test-agent",
		VersionValue:   "1.0.0",
		Meta: integration.IntegrationMeta{
			Name:      "validation-failure",
			AgentType: "test-agent",
			Version:   "1.0.0",
			Status:    integration.StatusInstalled,
		},
		ValidateFn: func(ctx context.Context) (*integration.ValidationResult, error) {
			return &integration.ValidationResult{
				Passed: false,
				Checks: []integration.ValidationCheck{
					{
						Name:         "mcp_server_reachable",
						Passed:       false,
						Message:      "MCP server not responding",
						Remediation:  "Start with `portfolio mcp start`",
						SelfHealable: true,
					},
				},
			}, nil
		},
	}
}

func NewFakeIntegrationWithUpgradeFailure() *FakeIntegration {
	return &FakeIntegration{
		NameValue:      "upgrade-failure",
		AgentTypeValue: "test-agent",
		VersionValue:   "1.0.0",
		Meta: integration.IntegrationMeta{
			Name:      "upgrade-failure",
			AgentType: "test-agent",
			Version:   "1.0.0",
			Status:    integration.StatusInstalled,
		},
		UpgradeFn: func(ctx context.Context, opts integration.UpgradeOptions) (*integration.UpgradeResult, error) {
			return nil, errors.New("upgrade failed: network timeout")
		},
	}
}
