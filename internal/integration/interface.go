package integration

import (
	"context"
)

type Integration interface {
	Name() string
	AgentType() string
	Install(ctx context.Context, opts InstallOptions) (*InstallResult, error)
	Validate(ctx context.Context) (*ValidationResult, error)
	Upgrade(ctx context.Context, opts UpgradeOptions) (*UpgradeResult, error)
	Remove(ctx context.Context) error
}
