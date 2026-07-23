package integration

import (
	"context"
)

type Store interface {
	SaveIntegration(ctx context.Context, meta IntegrationMeta) error
	GetIntegration(ctx context.Context, name string) (*IntegrationMeta, error)
	ListIntegrations(ctx context.Context) ([]IntegrationMeta, error)
	DeleteIntegration(ctx context.Context, name string) error
}
